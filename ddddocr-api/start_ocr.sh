#!/bin/bash
#
# start_ocr.sh - 启动 ddddocr OCR 服务（裸机部署）
#
# 用法: ./start_ocr.sh
#
# 环境变量:
#   OCR_HOST     监听地址 (默认 0.0.0.0)
#   OCR_PORT     监听端口 (默认 8899)
#   OCR_WORKERS  worker 数量 (默认 1)
#   OCR_LOG      日志文件路径 (默认 ./ocr.log)
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OCR_HOST="${OCR_HOST:-0.0.0.0}"
OCR_PORT="${OCR_PORT:-8899}"
OCR_WORKERS="${OCR_WORKERS:-1}"
OCR_LOG="${OCR_LOG:-ocr.log}"

echo "[OCR] 启动 ddddocr 服务..."
echo "[OCR] 监听地址: $OCR_HOST:$OCR_PORT"
echo "[OCR] Workers: $OCR_WORKERS"
echo "[OCR] 日志文件: $OCR_LOG"

# 检查 Python3
if ! command -v python3 &>/dev/null; then
    echo "[OCR] 错误: 未找到 python3，请先运行 install.sh"
    exit 1
fi

# 检查依赖
if ! python3 -c "import ddddocr" 2>/dev/null; then
    echo "[OCR] 警告: ddddocr 未安装，尝试安装..."
    ./install.sh
fi

# 检查端口是否已被 ddddocr 服务占用（通过健康检查确认）
if command -v curl &>/dev/null; then
    if curl -sf "http://${OCR_HOST}:${OCR_PORT}/health" -m 3 &>/dev/null; then
        echo "[OCR] ddddocr 服务已在运行 (端口 ${OCR_PORT})，跳过启动"
        exit 0
    fi
fi

# 清理残留进程
PID_FILE="$SCRIPT_DIR/ocr.pid"
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if [ -n "$OLD_PID" ] && kill -0 "$OLD_PID" 2>/dev/null; then
        echo "[OCR] 停止残留进程 (pid=$OLD_PID)..."
        kill "$OLD_PID" 2>/dev/null || true
        sleep 1
    fi
    rm -f "$PID_FILE"
fi

# 也尝试按进程名清理
pkill -f "python3.*ocr_server.py" 2>/dev/null || true
pkill -f "uvicorn.*ddddocr-api" 2>/dev/null || true

# 启动服务（nohup 让进程在父 shell 退出后继续运行，macOS/Linux 通用）
# redirect stdin from /dev/null 确保进程不持有终端
# stderr → stdout → 日志文件，确保输出不丢失
nohup python3 -m uvicorn \
    --host "$OCR_HOST" \
    --port "$OCR_PORT" \
    --workers "$OCR_WORKERS" \
    --log-level info \
    --no-access-log \
    --no-use-colors \
    ocr_server:app \
    </dev/null \
    >> "$OCR_LOG" 2>&1 &

PID=$!
echo "$PID" > "$PID_FILE"
echo "[OCR] 进程已启动 (pid=$PID)"
echo "[OCR] 等待服务就绪..."

# 等待健康检查通过（最多 60 秒）
for i in $(seq 1 60); do
    if curl -sf "http://${OCR_HOST}:${OCR_PORT}/health" -m 3 &>/dev/null; then
        echo "[OCR] 服务就绪! (耗时 ${i}s)"
        echo "[OCR] 健康检查: http://${OCR_HOST}:${OCR_PORT}/health"
        exit 0
    fi
    sleep 1
done

echo "[OCR] 警告: 健康检查超时 (60s)，服务可能未完全就绪"
echo "[OCR] 请检查日志: tail -f $OCR_LOG"
exit 1
