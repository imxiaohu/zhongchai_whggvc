#!/bin/bash

# 应用版本管理表初始化脚本
# 用于设置应用更新功能所需的数据库表和初始数据

echo "=== 应用版本管理表初始化 ==="

# 检查数据库文件是否存在
DB_FILE="pingjiao.db"
if [ ! -f "$DB_FILE" ]; then
    echo "错误: 数据库文件 $DB_FILE 不存在"
    echo "请先运行应用程序以创建数据库"
    exit 1
fi

# 执行SQL脚本
echo "正在执行SQL脚本..."
sqlite3 "$DB_FILE" < scripts/init_app_versions.sql

if [ $? -eq 0 ]; then
    echo "✅ 应用版本表初始化成功"
    
    # 显示创建的表信息
    echo ""
    echo "=== 表结构信息 ==="
    sqlite3 "$DB_FILE" ".schema app_versions"
    
    echo ""
    echo "=== 初始数据 ==="
    sqlite3 "$DB_FILE" "SELECT app_id, version, version_code, platform, size, is_active FROM app_versions ORDER BY platform, version_code;"
    
    echo ""
    echo "=== 使用说明 ==="
    echo "1. 启动Go后端服务: go run main.go"
    echo "2. 打开测试页面: http://localhost:2333/test_app_update.html"
    echo "3. 在uni-app项目中测试更新功能"
    echo ""
    echo "=== API端点 ==="
    echo "- GET  /api/app/info           - 获取应用信息"
    echo "- POST /api/app/check-update   - 检查更新"
    echo "- GET  /api/app/versions       - 获取版本列表"
    echo "- POST /api/app/versions       - 创建新版本"
    echo "- PUT  /api/app/versions/:id   - 更新版本"
    echo "- DELETE /api/app/versions/:id - 删除版本"
    
else
    echo "❌ 应用版本表初始化失败"
    exit 1
fi
