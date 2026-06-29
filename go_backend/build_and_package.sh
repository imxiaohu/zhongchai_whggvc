#!/bin/bash

# 本地构建和打包脚本
# 生成可以手动上传的部署包

set -e

echo "🚀 开始构建部署包..."

# 1. 编译 Go 程序 (Linux 64位)
echo "📦 编译 Go 程序 (Linux amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pingjiao_server main.go

if [ ! -f "pingjiao_server" ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译完成"

# 2. 创建部署目录
DEPLOY_DIR="pingjiao_deploy_$(date +%Y%m%d_%H%M%S)"
echo "📁 创建部署目录: ${DEPLOY_DIR}"
mkdir -p ${DEPLOY_DIR}

# 3. 复制必要文件
echo "📋 复制文件..."
cp pingjiao_server ${DEPLOY_DIR}/
cp .env ${DEPLOY_DIR}/
cp -r certs ${DEPLOY_DIR}/
cp -r static ${DEPLOY_DIR}/

# 4. 创建服务管理脚本
echo "📝 创建服务管理脚本..."

# 启动脚本
cat > ${DEPLOY_DIR}/start.sh << 'EOF'
#!/bin/bash

echo "🚀 启动 Pingjiao 服务..."

# 检查是否已有进程在运行
if pgrep -f "pingjiao_server" > /dev/null; then
    echo "⚠️  服务已在运行，先停止现有服务"
    pkill -f "pingjiao_server"
    sleep 2
fi

# 启动服务
nohup ./pingjiao_server > pingjiao.log 2>&1 &
PID=$!

echo "✅ 服务已启动 (PID: $PID)"
echo "📝 日志文件: pingjiao.log"

# 等待服务启动
sleep 3

# 检查服务状态
if kill -0 $PID 2>/dev/null; then
    echo "🎉 服务启动成功！"
    echo "🔗 API地址: http://localhost:2333"
    echo "🏥 健康检查: curl http://localhost:2333/api/health"
else
    echo "❌ 服务启动失败，请检查日志"
    tail -20 pingjiao.log
fi
EOF

# 停止脚本
cat > ${DEPLOY_DIR}/stop.sh << 'EOF'
#!/bin/bash

echo "🛑 停止 Pingjiao 服务..."

if pgrep -f "pingjiao_server" > /dev/null; then
    pkill -f "pingjiao_server"
    echo "✅ 服务已停止"
else
    echo "ℹ️  服务未运行"
fi
EOF

# 状态检查脚本
cat > ${DEPLOY_DIR}/status.sh << 'EOF'
#!/bin/bash

echo "📊 Pingjiao 服务状态"
echo "===================="

if pgrep -f "pingjiao_server" > /dev/null; then
    PID=$(pgrep -f "pingjiao_server")
    echo "✅ 服务正在运行 (PID: $PID)"
    echo ""
    echo "📈 进程信息:"
    ps -p $PID -o pid,ppid,cmd,%mem,%cpu
    echo ""
    echo "📝 最近日志 (最后10行):"
    if [ -f "pingjiao.log" ]; then
        tail -10 pingjiao.log
    else
        echo "日志文件不存在"
    fi
else
    echo "❌ 服务未运行"
fi

echo ""
echo "🔗 测试连接:"
if curl -s -f http://localhost:2333/api/health > /dev/null 2>&1; then
    echo "✅ API 健康检查通过"
else
    echo "❌ API 健康检查失败"
fi
EOF

# 重启脚本
cat > ${DEPLOY_DIR}/restart.sh << 'EOF'
#!/bin/bash

echo "🔄 重启 Pingjiao 服务..."
./stop.sh
sleep 2
./start.sh
EOF

# 设置执行权限
chmod +x ${DEPLOY_DIR}/*.sh

# 5. 创建部署说明
cat > ${DEPLOY_DIR}/README.md << 'EOF'
# Pingjiao 服务部署包

## 文件说明
- `pingjiao_server`: 主程序可执行文件
- `.env`: 环境配置文件
- `certs/`: 微信支付证书目录
- `static/`: 静态资源目录
- `start.sh`: 启动服务脚本
- `stop.sh`: 停止服务脚本
- `restart.sh`: 重启服务脚本
- `status.sh`: 查看服务状态脚本

## 部署步骤

1. 将整个目录上传到服务器
2. 进入部署目录
3. 运行启动脚本: `./start.sh`

## 服务管理

- 启动服务: `./start.sh`
- 停止服务: `./stop.sh`
- 重启服务: `./restart.sh`
- 查看状态: `./status.sh`
- 查看日志: `tail -f pingjiao.log`

## 端口配置
- 默认端口: 2333
- 健康检查: http://localhost:2333/api/health

## 注意事项
- 确保端口 2333 未被占用
- 确保有足够的磁盘空间
- 建议使用 systemd 或其他进程管理器管理服务
EOF

# 6. 创建 systemd 服务文件（可选）
cat > ${DEPLOY_DIR}/pingjiao.service << 'EOF'
[Unit]
Description=Pingjiao Go Backend Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/pingjiao
ExecStart=/opt/pingjiao/pingjiao_server
Restart=always
RestartSec=5
StandardOutput=append:/var/log/pingjiao.log
StandardError=append:/var/log/pingjiao.log

[Install]
WantedBy=multi-user.target
EOF

# 7. 打包
echo "🗜️  打包部署文件..."
tar -czf ${DEPLOY_DIR}.tar.gz ${DEPLOY_DIR}

# 8. 显示结果
echo ""
echo "🎉 构建完成！"
echo "📦 部署包: ${DEPLOY_DIR}.tar.gz"
echo "📁 部署目录: ${DEPLOY_DIR}"
echo ""
echo "📋 部署步骤:"
echo "1. 将 ${DEPLOY_DIR}.tar.gz 上传到服务器"
echo "2. 解压: tar -xzf ${DEPLOY_DIR}.tar.gz"
echo "3. 进入目录: cd ${DEPLOY_DIR}"
echo "4. 启动服务: ./start.sh"
echo ""
echo "🔧 可选: 安装为系统服务"
echo "sudo cp pingjiao.service /etc/systemd/system/"
echo "sudo systemctl daemon-reload"
echo "sudo systemctl enable pingjiao"
echo "sudo systemctl start pingjiao"

# 清理编译文件
rm -f pingjiao_server

echo ""
echo "✅ 构建完成！"
