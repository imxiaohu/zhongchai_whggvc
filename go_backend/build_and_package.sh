#!/bin/bash

# build_and_package.sh - 构建和打包部署包
# 生成包含 Go 服务 + ddddocr OCR 服务的完整部署包

set -e

echo "========================================"
echo "  Pingjiao 部署包构建"
echo "========================================"

# 1. 编译 Go 程序 (Linux 64位)
echo ""
echo "[1/7] 编译 Go 程序 (Linux amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pingjiao_server main.go

if [ ! -f "pingjiao_server" ]; then
    echo "错误: 编译失败，未找到 pingjiao_server"
    exit 1
fi
echo "    编译完成"

# 2. 创建部署目录
DEPLOY_DIR="pingjiao_deploy_$(date +%Y%m%d_%H%M%S)"
echo ""
echo "[2/7] 创建部署目录: ${DEPLOY_DIR}"
mkdir -p ${DEPLOY_DIR}

# 3. 复制 Go 服务文件
echo ""
echo "[3/7] 复制 Go 服务文件..."
cp pingjiao_server ${DEPLOY_DIR}/

# .env: 从 .env.example 复制并生成（如果 .env 不存在）
if [ -f ".env" ]; then
    cp .env ${DEPLOY_DIR}/.env
    echo "    已复制 .env"
elif [ -f ".env.example" ]; then
    cp .env.example ${DEPLOY_DIR}/.env
    echo "    已从 .env.example 生成 .env"
fi

cp -r certs ${DEPLOY_DIR}/
cp -r static ${DEPLOY_DIR}/

# 4. 打包 ddddocr-api 目录（从上级 ../ddddocr-api 复制）
echo ""
echo "[4/7] 打包 ddddocr-api..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DDDDOCR_SRC="${SCRIPT_DIR}/../ddddocr-api"

if [ -d "$DDDDOCR_SRC" ]; then
    cp -r "$DDDDOCR_SRC" "${DEPLOY_DIR}/ddddocr-api"
    echo "    已复制 ddddocr-api/ 目录"
else
    echo "    警告: 未找到 ../ddddocr-api 目录，跳过"
fi

# 5. 生成部署脚本
echo ""
echo "[5/7] 生成部署脚本..."

# ---------- 一键部署脚本 deploy.sh ----------
cat > ${DEPLOY_DIR}/deploy.sh << 'DEPLOY_EOF'
#!/bin/bash
#
# deploy.sh - 一键部署脚本（安装依赖 + 启动服务）
#
# 用法: ./deploy.sh
#
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "  Pingjiao 一键部署"
echo "========================================"

# 安装 OCR 依赖（如已安装则跳过）
if ! python3 -c "import ddddocr" 2>/dev/null; then
    echo ""
    echo "[STEP 1/4] 安装 ddddocr 依赖..."
    if [ -x "./ddddocr-api/install.sh" ]; then
        ./ddddocr-api/install.sh
    else
        echo "错误: ddddocr-api/install.sh 不存在或无执行权限"
        exit 1
    fi
else
    echo "[STEP 1/4] ddddocr 已安装，跳过"
fi

# 启动 OCR 服务
echo ""
echo "[STEP 2/4] 启动 ddddocr OCR 服务..."
if [ -x "./ddddocr-api/start_ocr.sh" ]; then
    ./ddddocr-api/start_ocr.sh
else
    echo "错误: ddddocr-api/start_ocr.sh 不存在或无执行权限"
    exit 1
fi

# 启动 Go 服务
echo ""
echo "[STEP 3/4] 启动 Go 服务..."
./start.sh

echo ""
echo "[STEP 4/4] 检查服务状态..."
sleep 2
./status.sh

echo ""
echo "========================================"
echo "  部署完成!"
echo "  OCR 服务:  http://localhost:8899/health"
echo "  Go 服务:   http://localhost:2333/api/health"
echo "========================================"
DEPLOY_EOF

# ---------- 启动脚本 start.sh ----------
cat > ${DEPLOY_DIR}/start.sh << 'START_EOF'
#!/bin/bash
#
# start.sh - 启动 Pingjiao 服务（OCR + Go）
#
# 启动顺序：先启动 OCR，等健康检查通过后再启动 Go
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[START] 启动 Pingjiao 服务..."

# --- OCR 服务 ---
if [ -x "./ddddocr-api/start_ocr.sh" ]; then
    echo "[START] 启动 ddddocr OCR 服务..."
    ./ddddocr-api/start_ocr.sh
else
    echo "[START] 警告: ddddocr-api/start_ocr.sh 不存在，跳过 OCR"
fi

# --- Go 服务 ---
echo "[START] 启动 Go 服务..."

# 检查是否已有 Go 进程在运行
if pgrep -f "pingjiao_server" > /dev/null; then
    echo "[START] 已有 Go 进程在运行，先停止..."
    ./stop.sh
    sleep 2
fi

# 用 setsid 启动 Go 服务（独立会话，防止 shell 退出影响进程）
setsid ./pingjiao_server > pingjiao.log 2>&1 &
PID=$!

echo "[START] Go 服务已启动 (PID: $PID)"
echo "[START] 日志文件: pingjiao.log"

# 等待 Go 服务响应（最多 30s）
echo "[START] 等待 Go 服务就绪..."
for i in $(seq 1 30); do
    if curl -sf "http://localhost:2333/api/health" -m 2 &>/dev/null; then
        echo "[START] Go 服务就绪 (耗时 ${i}s)"
        break
    fi
    sleep 1
done

# 检查 Go 进程是否存活
if kill -0 $PID 2>/dev/null; then
    echo "[START] 服务启动成功!"
else
    echo "[START] 错误: Go 服务启动失败，请检查日志"
    tail -30 pingjiao.log
    exit 1
fi

echo ""
echo "[START] 服务状态:"
./status.sh
START_EOF

# ---------- 停止脚本 stop.sh ----------
cat > ${DEPLOY_DIR}/stop.sh << 'STOP_EOF'
#!/bin/bash
#
# stop.sh - 停止 Pingjiao 服务（Go + OCR）
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "[STOP] 停止 Pingjiao 服务..."

# 先停 Go（业务优先）
if pgrep -f "pingjiao_server" > /dev/null 2>&1; then
    echo "[STOP] 停止 Go 服务..."
    pkill -f "pingjiao_server" 2>/dev/null || true
    sleep 2
    # 如果还在跑，强制杀
    if pgrep -f "pingjiao_server" > /dev/null 2>&1; then
        echo "[STOP] 强制停止 Go 服务..."
        pkill -9 -f "pingjiao_server" 2>/dev/null || true
    fi
    echo "[STOP] Go 服务已停止"
fi

# 再停 OCR
if [ -x "./ddddocr-api/stop_ocr.sh" ]; then
    echo "[STOP] 停止 ddddocr OCR 服务..."
    ./ddddocr-api/stop_ocr.sh
fi

echo "[STOP] 所有服务已停止"
STOP_EOF

# ---------- 状态检查脚本 status.sh ----------
cat > ${DEPLOY_DIR}/status.sh << 'STATUS_EOF'
#!/bin/bash
#
# status.sh - 查看服务状态（OCR + Go）
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "  Pingjiao 服务状态"
echo "========================================"

# --- Go 服务 ---
echo ""
echo "【Go 服务】"
if pgrep -f "pingjiao_server" > /dev/null 2>&1; then
    PID=$(pgrep -f "pingjiao_server")
    echo "  状态:   运行中 (PID: $PID)"
    if curl -sf "http://localhost:2333/api/health" -m 3 &>/dev/null; then
        echo "  HTTP:   正常"
    else
        echo "  HTTP:   异常（连接失败）"
    fi
else
    echo "  状态:   未运行"
fi

# --- OCR 服务 ---
echo ""
echo "【ddddocr OCR 服务】"
OCR_HOST="${OCR_HOST:-127.0.0.1}"
OCR_PORT="${OCR_PORT:-8899}"

if curl -sf "http://${OCR_HOST}:${OCR_PORT}/health" -m 3 &>/dev/null; then
    echo "  状态:   运行中"
    echo "  健康:   正常"
    echo "  地址:   http://${OCR_HOST}:${OCR_PORT}/health"
else
    echo "  状态:   未运行或不可达"
    echo "  地址:   http://${OCR_HOST}:${OCR_PORT}/health"
fi

# OCR PID
if [ -f "./ddddocr-api/ocr.pid" ]; then
    OCR_PID=$(cat ./ddddocr-api/ocr.pid 2>/dev/null || echo "未知")
    echo "  PID:    $OCR_PID"
fi

echo ""
echo "【日志】"
for log in pingjiao.log ./ddddocr-api/ocr.log; do
    if [ -f "$log" ]; then
        echo "  $log (最后5行):"
        tail -5 "$log" | sed 's/^/    /'
    fi
done

echo ""
echo "========================================"
STATUS_EOF

# ---------- systemd 服务文件（可选）----------
echo ""
echo "[6/7] 生成 systemd 服务文件..."

# OCR 服务
cat > ${DEPLOY_DIR}/pingjiao-ocr.service << 'OCRSVC_EOF'
[Unit]
Description=Pingjiao ddddocr OCR Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/pingjiao/ddddocr-api
ExecStart=/opt/pingjiao/ddddocr-api/start_ocr.sh
ExecStop=/opt/pingjiao/ddddocr-api/stop_ocr.sh
Restart=on-failure
RestartSec=10
StandardOutput=append:/var/log/pingjiao-ocr-stdout.log
StandardError=append:/var/log/pingjiao-ocr-stderr.log

[Install]
WantedBy=multi-user.target
OCRSVC_EOF

# Go 服务（依赖 OCR）
cat > ${DEPLOY_DIR}/pingjiao.service << 'GOSVC_EOF'
[Unit]
Description=Pingjiao Go Backend Service
After=network.target pingjiao-ocr.service
Requires=pingjiao-ocr.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/pingjiao
ExecStart=/opt/pingjiao/pingjiao_server
ExecStop=/bin/bash -c 'kill $(pgrep -f pingjiao_server) && sleep 2 && pkill -9 -f pingjiao_server || true'
Restart=on-failure
RestartSec=10
StandardOutput=append:/var/log/pingjiao-stdout.log
StandardError=append:/var/log/pingjiao-stderr.log
Environment="OCR_HOST=127.0.0.1"
Environment="OCR_PORT=8899"

[Install]
WantedBy=multi-user.target
GOSVC_EOF

# 6. 设置执行权限
echo ""
echo "[7/7] 设置执行权限..."
chmod +x ${DEPLOY_DIR}/*.sh
chmod +x ${DEPLOY_DIR}/ddddocr-api/*.sh 2>/dev/null || true

# 7. 打包
echo ""
echo "========================================"
echo "  打包..."
tar -czf ${DEPLOY_DIR}.tar.gz ${DEPLOY_DIR}

# 清理编译文件
rm -f pingjiao_server

echo ""
echo "========================================"
echo "  构建完成!"
echo ""
echo "  部署包:   ${DEPLOY_DIR}.tar.gz"
echo "  部署目录: ${DEPLOY_DIR}"
echo ""
echo "  部署步骤:"
echo "  1. 上传到服务器: scp ${DEPLOY_DIR}.tar.gz root@your-server:/www/wwwroot/"
echo "  2. 解压: tar -xzf ${DEPLOY_DIR}.tar.gz"
echo "  3. 一键部署: cd ${DEPLOY_DIR} && ./deploy.sh"
echo ""
echo "  或手动分步:"
echo "    cd ${DEPLOY_DIR}"
echo "    ./ddddocr-api/install.sh   # 首次安装 Python 依赖"
echo "    ./start.sh                 # 启动所有服务"
echo ""
echo "  可选 systemd 安装:"
echo "    cp pingjiao-ocr.service /etc/systemd/system/"
echo "    cp pingjiao.service /etc/systemd/system/"
echo "    systemctl daemon-reload"
echo "    systemctl enable pingjiao-ocr pingjiao"
echo "    systemctl start pingjiao-ocr pingjiao"
echo "========================================"
