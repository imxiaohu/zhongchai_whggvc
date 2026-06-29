#!/bin/bash
# changelog.sh - 自动生成 CHANGELOG.md
# 用法: bash scripts/changelog.sh [版本号] [--full]
#   版本号  - 可选，指定本次发布的版本标签 (如 1.0.0)
#   --full  - 可选，生成完整 changelog (包含所有版本)，默认只生成自上次标签后的变更

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CHGLOG_DIR="$PROJECT_ROOT/.chglog"
OUTPUT_FILE="$PROJECT_ROOT/CHANGELOG.md"
TEMPLATE_FILE="$CHGLOG_DIR/CHANGELOG.tpl.md"

cd "$PROJECT_ROOT"

# 确保 git-chglog 在 PATH 中
export PATH="$PATH:$(go env GOPATH 2>/dev/null)/bin"

# 检查 git-chglog 是否安装
check_git_chglog() {
    if ! command -v git-chglog &> /dev/null; then
        echo "⚠️  git-chglog 未安装，正在安装..."
        if command -v go &> /dev/null; then
            go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
            export PATH="$PATH:$(go env GOPATH)/bin"
            echo "✅ git-chglog 安装成功"
        else
            echo "❌ Go 未安装，请先安装 Go: https://go.dev/dl/"
            exit 1
        fi
    fi
}

# 检查环境
init_env() {
    if [ ! -d "$CHGLOG_DIR" ]; then
        echo "❌ .git-chglog 目录不存在"
        exit 1
    fi
    if [ ! -f "$TEMPLATE_FILE" ]; then
        echo "❌ 模板文件 $TEMPLATE_FILE 不存在"
        exit 1
    fi
}

# 生成增量 changelog (自上次标签以来的变更)
generate_incremental() {
    local version="${1:-}"
    echo "📝 生成增量 CHANGELOG..."

    local temp_file
    temp_file=$(mktemp)

    if [ -n "$version" ]; then
        echo "   版本: $version"
        # 生成指定版本到最新标签的 changelog
        git-chglog --config "$CHGLOG_DIR/config.yml" \
            --template "$TEMPLATE_FILE" \
            "$version..HEAD" > "$temp_file"
    else
        # 生成自上次标签以来的变更
        git-chglog --config "$CHGLOG_DIR/config.yml" \
            --template "$TEMPLATE_FILE" \
            "$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)..HEAD" > "$temp_file" || true
    fi

    # 将自动生成的内容插入到手动维护内容之前
    if [ -f "$OUTPUT_FILE" ]; then
        # 找到 ---END--- 标记的位置
        local marker_line
        marker_line=$(grep -n "^---END---" "$OUTPUT_FILE" 2>/dev/null | head -1 | cut -d: -f1 || echo "0")

        if [ "$marker_line" != "0" ]; then
            # 保留头部说明部分
            head -n $((marker_line - 1)) "$OUTPUT_FILE" > "$temp_file.tmp"
            echo "" >> "$temp_file.tmp"
            cat "$temp_file" >> "$temp_file.tmp"
            mv "$temp_file.tmp" "$temp_file"
        fi
    fi

    cp "$temp_file" "$OUTPUT_FILE"
    rm -f "$temp_file"

    echo "✅ CHANGELOG.md 已更新"
}

# 生成完整 changelog (所有版本)
generate_full() {
    echo "📝 生成完整 CHANGELOG..."
    # 处理无 tag 的情况，自动从 package.json 读取版本创建初始 tag
    if ! git describe --tags --abbrev=0 &>/dev/null; then
        local init_tag="${1:-}"
        if [ -z "$init_tag" ] && [ -f "$PROJECT_ROOT/package.json" ]; then
            init_tag=$(node -e "const p=require('$PROJECT_ROOT/package.json'); console.log(p.version)" 2>/dev/null || echo "1.0.0")
        fi
        init_tag="${init_tag:-1.0.0}"
        echo "⚠️  未检测到 git tag，自动创建初始标签: v$init_tag"
        git tag -a "v$init_tag" -m "Initial release v$init_tag"
    fi
    git-chglog --config "$CHGLOG_DIR/config.yml" \
        --template "$TEMPLATE_FILE" \
        --output "$OUTPUT_FILE" \
        2>&1 | head -20 || echo "   (部分版本可能无符合规范的提交)"
    echo "✅ 完整 CHANGELOG.md 已生成"
}

# 主流程
main() {
    echo "🚀 Changelog 生成器"
    echo "=================="

    check_git_chglog
    init_env

    local version=""
    local full=false

    for arg in "$@"; do
        case "$arg" in
            --full) full=true ;;
            -h|--help)
                echo "用法: $0 [版本号] [--full] [--help]"
                echo "  版本号  - 指定本次发布的版本标签 (如 1.0.0)"
                echo "  --full  - 生成完整 changelog"
                echo "  --help  - 显示帮助"
                exit 0
                ;;
            -*) ;;
            *) version="$arg" ;;
        esac
    done

    if [ "$full" = true ]; then
        generate_full "$version"
    else
        generate_incremental "$version"
    fi

    echo ""
    echo "📄 输出: $OUTPUT_FILE"
    echo ""
    echo "💡 下一步:"
    echo "   1. 提交代码: git add . && git commit -m 'chore: update CHANGELOG'"
    echo "   2. 打标签:   git tag -a v${version:-x.y.z} -m 'Release v${version:-x.y.z}'"
    echo "   3. 推送:     git push && git push --tags"
}

main "$@"
