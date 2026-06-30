#!/bin/bash

# 切换到MySQL数据库的脚本

echo "正在切换到MySQL数据库..."

# 检查是否存在MySQL配置示例文件
if [ ! -f ".env.mysql.example" ]; then
    echo "错误：找不到 .env.mysql.example 文件"
    exit 1
fi

# 备份当前配置（如果存在）
if [ -f ".env" ]; then
    echo "备份当前配置到 .env.backup"
    cp .env .env.backup
fi

# 复制MySQL配置
echo "复制MySQL配置..."
cp .env.mysql.example .env

echo "请编辑 .env 文件，设置正确的MySQL连接信息："
echo "  MYSQL_HOST=localhost"
echo "  MYSQL_PORT=3306"
echo "  MYSQL_USER=your_username"
echo "  MYSQL_PASSWORD=your_password"
echo "  MYSQL_DATABASE=pingjiao"

echo ""
echo "然后运行以下命令初始化MySQL数据库："
echo "  mysql -u root -p < scripts/init_mysql.sql"

echo ""
echo "最后重启应用程序："
echo "  go run main.go"

echo ""
echo "切换完成！"
