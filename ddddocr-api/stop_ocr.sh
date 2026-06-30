#!/bin/bash
#
# stop_ocr.sh - 停止 ddddocr OCR 服务
#
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[OCR] 停止 ddddocr 服务..."

# 1. 读 PID 文件
PID_FILE="$SCRIPT_DIR/ocr.pid"
STOPPED=0

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        echo "[OCR] 停止进程 (pid=$PID)..."
        kill "$PID" 2>/dev/null || true
        # 等最多 5 秒
        for i in $(seq 1 10); do
            if ! kill -0 "$PID" 2>/dev/null; then
                echo "[OCR] 进程已停止"
                STOPPED=1
                break
            fi
            sleep 0.5
        done
        if [ "$STOPPED" -eq 0 ]; then
            echo "[OCR] 强制杀死进程 (pid=$PID)..."
            kill -9 "$PID" 2>/dev/null || true
        fi
    fi
    rm -f "$PID_FILE"
fi

# 2. 按进程名兜底清理
if pgrep -f "python3.*ocr_server.py" > /dev/null 2>&1; then
    echo "[OCR] 按进程名停止残留进程..."
    pkill -f "python3.*ocr_server.py" 2>/dev/null || true
    pkill -f "uvicorn.*ocr_server:app" 2>/dev/null || true
fi

echo "[OCR] 停止完成"
