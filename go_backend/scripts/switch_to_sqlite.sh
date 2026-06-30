#!/bin/bash

# 切换到SQLite数据库的脚本

echo "正在切换到SQLite数据库..."

# 检查是否存在SQLite配置示例文件
if [ ! -f ".env.sqlite.example" ]; then
    echo "错误：找不到 .env.sqlite.example 文件"
    exit 1
fi

# 备份当前配置（如果存在）
if [ -f ".env" ]; then
    echo "备份当前配置到 .env.backup"
    cp .env .env.backup
fi

# 复制SQLite配置
echo "复制SQLite配置..."
cp .env.sqlite.example .env

echo "SQLite配置已设置，数据库文件将保存在 ./data/pingjiao.db"

echo ""
echo "重启应用程序："
echo "  go run main.go"

echo ""
echo "切换完成！"
