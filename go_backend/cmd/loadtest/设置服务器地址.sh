#!/bin/bash

# 压测服务器地址设置脚本
# 用于快速修改压测目标服务器地址

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置文件路径
CONFIG_FILE="压测配置.json"
GO_FILE="压测工具.go"

# 打印彩色输出
print_info() {
    echo -e "${BLUE}[信息]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[成功]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

print_error() {
    echo -e "${RED}[错误]${NC} $1"
}

print_title() {
    echo -e "${CYAN}$1${NC}"
}

# 显示当前配置
show_current_config() {
    print_title "=== 当前压测配置 ==="
    
    if [ -f "$GO_FILE" ]; then
        # 从Go文件中提取当前默认地址
        current_url=$(grep -o 'return ".*"' "$GO_FILE" | tail -1 | sed 's/return "//;s/"//')
        echo "当前默认服务器地址: $current_url"
    fi
    
    if [ -f "$CONFIG_FILE" ]; then
        # 从JSON配置文件中提取地址
        if command -v jq &> /dev/null; then
            config_url=$(jq -r '.压测配置.服务器地址' "$CONFIG_FILE" 2>/dev/null || echo "无法读取")
            echo "配置文件中的地址: $config_url"
            
            echo ""
            echo "备用地址列表:"
            jq -r '.压测配置.备用地址[]' "$CONFIG_FILE" 2>/dev/null | while read -r url; do
                echo "  - $url"
            done
        else
            print_warning "未安装jq工具，无法解析JSON配置文件"
        fi
    fi
    
    # 显示环境变量
    if [ -n "$压测服务器地址" ]; then
        echo "环境变量 压测服务器地址: $压测服务器地址"
    fi
    if [ -n "$LOAD_TEST_URL" ]; then
        echo "环境变量 LOAD_TEST_URL: $LOAD_TEST_URL"
    fi
    
    echo ""
}

# 显示预设地址选项
show_preset_options() {
    print_title "=== 预设服务器地址 ==="
    echo "1. 生产服务器: https://go.server.zhongchai.imxiaohu.cn"
    echo "2. 本地开发: http://localhost:2333"
    echo "3. 测试服务器: https://test.server.com"
    echo "4. 预发布环境: https://staging.server.com"
    echo "5. 自定义地址"
    echo ""
}

# 更新Go文件中的默认地址
update_go_file() {
    local new_url="$1"
    
    if [ -f "$GO_FILE" ]; then
        # 备份原文件
        cp "$GO_FILE" "${GO_FILE}.backup"
        
        # 更新默认地址
        sed -i.tmp "s|return \".*\"|return \"$new_url\"|g" "$GO_FILE"
        rm "${GO_FILE}.tmp" 2>/dev/null || true
        
        print_success "已更新 $GO_FILE 中的默认地址"
    else
        print_error "未找到 $GO_FILE 文件"
        return 1
    fi
}

# 更新JSON配置文件
update_config_file() {
    local new_url="$1"
    
    if [ -f "$CONFIG_FILE" ] && command -v jq &> /dev/null; then
        # 备份原文件
        cp "$CONFIG_FILE" "${CONFIG_FILE}.backup"
        
        # 更新配置文件
        jq --arg url "$new_url" '.压测配置.服务器地址 = $url' "$CONFIG_FILE" > "${CONFIG_FILE}.tmp"
        mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"
        
        print_success "已更新 $CONFIG_FILE 中的服务器地址"
    elif [ ! -f "$CONFIG_FILE" ]; then
        print_warning "未找到 $CONFIG_FILE 文件"
    elif ! command -v jq &> /dev/null; then
        print_warning "未安装jq工具，跳过JSON配置文件更新"
    fi
}

# 设置环境变量
set_environment_variable() {
    local new_url="$1"
    local shell_rc=""
    
    # 检测shell类型
    if [ -n "$ZSH_VERSION" ]; then
        shell_rc="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        shell_rc="$HOME/.bashrc"
    else
        shell_rc="$HOME/.profile"
    fi
    
    echo ""
    print_info "您可以设置环境变量以覆盖默认配置："
    echo ""
    echo "临时设置（当前会话）："
    echo "  export 压测服务器地址=\"$new_url\""
    echo "  export LOAD_TEST_URL=\"$new_url\""
    echo ""
    echo "永久设置（添加到 $shell_rc）："
    echo "  echo 'export 压测服务器地址=\"$new_url\"' >> $shell_rc"
    echo "  echo 'export LOAD_TEST_URL=\"$new_url\"' >> $shell_rc"
    echo ""
    
    read -p "是否要设置临时环境变量？(y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        export 压测服务器地址="$new_url"
        export LOAD_TEST_URL="$new_url"
        print_success "已设置临时环境变量"
    fi
}

# 验证服务器连接
test_server_connection() {
    local url="$1"
    local health_endpoint="${url}/api/health/school-server"
    
    print_info "正在测试服务器连接: $health_endpoint"
    
    if curl -s --connect-timeout 10 --max-time 30 "$health_endpoint" > /dev/null 2>&1; then
        print_success "服务器连接正常"
        return 0
    else
        print_warning "服务器连接失败或健康检查接口无响应"
        return 1
    fi
}

# 主菜单
main_menu() {
    while true; do
        clear
        print_title "========================================"
        print_title "  压测服务器地址配置工具"
        print_title "========================================"
        echo ""
        
        show_current_config
        show_preset_options
        
        read -p "请选择操作 (1-5, q退出): " choice
        
        case $choice in
            1)
                new_url="https://go.server.zhongchai.imxiaohu.cn"
                ;;
            2)
                new_url="http://localhost:2333"
                ;;
            3)
                new_url="https://test.server.com"
                ;;
            4)
                new_url="https://staging.server.com"
                ;;
            5)
                echo ""
                read -p "请输入自定义服务器地址: " new_url
                if [ -z "$new_url" ]; then
                    print_error "地址不能为空"
                    read -p "按回车键继续..."
                    continue
                fi
                ;;
            q|Q)
                print_info "退出配置工具"
                exit 0
                ;;
            *)
                print_error "无效选择，请重新输入"
                read -p "按回车键继续..."
                continue
                ;;
        esac
        
        echo ""
        print_info "新的服务器地址: $new_url"
        echo ""
        
        # 确认更改
        read -p "确认更新配置？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "取消更新"
            read -p "按回车键继续..."
            continue
        fi
        
        # 测试连接
        echo ""
        read -p "是否要测试服务器连接？(Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            test_server_connection "$new_url"
            echo ""
        fi
        
        # 更新配置
        print_info "正在更新配置..."
        update_go_file "$new_url"
        update_config_file "$new_url"
        set_environment_variable "$new_url"
        
        echo ""
        print_success "配置更新完成！"
        echo ""
        print_info "现在您可以运行压测工具："
        echo "  go run 压测工具.go 快速"
        echo "  go run 压测工具.go 中等"
        echo ""
        print_info "或者使用命令行参数指定地址："
        echo "  go run 压测工具.go -u $new_url 快速"
        echo ""
        
        read -p "按回车键继续..."
    done
}

# 显示使用说明
show_usage() {
    cat << EOF
压测服务器地址配置工具

用法: $0 [选项] [地址]

选项:
  -h, --help          显示此帮助信息
  -s, --show          显示当前配置
  -t, --test URL      测试指定地址的连接
  -q, --quick URL     快速设置地址（不进入交互模式）

示例:
  $0                                    # 进入交互模式
  $0 -s                                # 显示当前配置
  $0 -t http://localhost:2333          # 测试连接
  $0 -q http://localhost:2333          # 快速设置地址

EOF
}

# 解析命令行参数
case "${1:-}" in
    -h|--help)
        show_usage
        exit 0
        ;;
    -s|--show)
        show_current_config
        exit 0
        ;;
    -t|--test)
        if [ -z "$2" ]; then
            print_error "请提供要测试的URL"
            exit 1
        fi
        test_server_connection "$2"
        exit $?
        ;;
    -q|--quick)
        if [ -z "$2" ]; then
            print_error "请提供服务器地址"
            exit 1
        fi
        print_info "快速设置服务器地址: $2"
        update_go_file "$2"
        update_config_file "$2"
        print_success "配置更新完成！"
        exit 0
        ;;
    "")
        # 无参数，进入交互模式
        main_menu
        ;;
    *)
        print_error "未知选项: $1"
        show_usage
        exit 1
        ;;
esac
