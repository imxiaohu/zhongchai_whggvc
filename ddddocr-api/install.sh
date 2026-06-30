#!/bin/bash
#
# install.sh - 安装 ddddocr-api 依赖（裸机部署）
#
# 用法: ./install.sh
#
# 支持: Linux (Ubuntu/Debian/CentOS), macOS
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "========================================"
echo "  ddddocr-api 依赖安装"
echo "========================================"

# 检查 Python3
if ! command -v python3 &>/dev/null; then
    echo "[ERROR] 未找到 python3，请先安装 Python 3.9+"
    exit 1
fi

PYTHON_CMD="python3"
PYTHON_VERSION=$($PYTHON_CMD -c 'import sys; print(f"{sys.version_info.major}.{sys.version_info.minor}")')
echo "[INFO] Python 版本: $PYTHON_VERSION"

# 检测操作系统
OS_TYPE=$(uname -s)
echo "[INFO] 操作系统: $OS_TYPE"

# 安装系统依赖
install_system_deps() {
    case "$OS_TYPE" in
        Linux)
            if command -v apt-get &>/dev/null; then
                echo "[INFO] 安装系统依赖 (apt-get)..."
                sudo apt-get update
                sudo apt-get install -y \
                    libgl1-mesa-glx \
                    libglib2.0-0 \
                    libgomp1 \
                    libxcb1 \
                    libx11-6 \
                    fonts-noto-cjk \
                    2>/dev/null || \
                sudo apt-get install -y \
                    libgl1 \
                    libglib2.0-0 \
                    libgomp1 \
                    fonts-noto-cjk \
                    2>/dev/null || true
            elif command -v yum &>/dev/null; then
                echo "[INFO] 安装系统依赖 (yum)..."
                sudo yum install -y \
                    mesa-libGL \
                    glib2 \
                    libgomp \
                    fonts-noto-cjk \
                    2>/dev/null || true
            elif command -v dnf &>/dev/null; then
                echo "[INFO] 安装系统依赖 (dnf)..."
                sudo dnf install -y \
                    mesa-libGL \
                    glib2 \
                    libgomp \
                    2>/dev/null || true
            fi
            ;;
        Darwin)
            echo "[INFO] macOS 通常已包含所需库，跳过系统依赖安装"
            ;;
        *)
            echo "[WARN] 未知操作系统: $OS_TYPE，跳过系统依赖"
            ;;
    esac
}

# 安装 Python 依赖
install_python_deps() {
    echo "[INFO] 安装 Python 依赖..."

    PIP_BASE=("$PYTHON_CMD" -m pip install)

    # macOS 需要 --break-system-packages
    if [ "$OS_TYPE" = "Darwin" ]; then
        PIP_BASE+=(--break-system-packages)
    fi

    PIP_BASE+=(--upgrade pip setuptools wheel)

    echo "[INFO] 升级 pip/setuptools/wheel..."
    "${PIP_BASE[@]}" 2>/dev/null || "$PYTHON_CMD" -m pip install --upgrade pip setuptools wheel

    if [ -f "$SCRIPT_DIR/requirements.txt" ]; then
        echo "[INFO] 从 requirements.txt 安装..."
        "$PYTHON_CMD" -m pip install "${PIP_BASE[@]:1}" -r "$SCRIPT_DIR/requirements.txt"
    else
        echo "[INFO] 直接安装 ddddocr..."
        "$PYTHON_CMD" -m pip install "${PIP_BASE[@]:1}" \
            ddddocr \
            fastapi \
            uvicorn[standard] \
            python-multipart \
            pydantic
    fi
}

# 验证安装
verify_install() {
    echo "[INFO] 验证安装..."
    if "$PYTHON_CMD" -c "import ddddocr; print(f'ddddocr {ddddocr.__version__} OK')" 2>/dev/null; then
        echo "[INFO] ddddocr 安装成功"
    else
        echo "[WARN] ddddocr 验证失败，请检查日志"
        return 1
    fi

    if "$PYTHON_CMD" -c "import fastapi; print(f'fastapi {fastapi.__version__} OK')" 2>/dev/null; then
        echo "[INFO] fastapi 安装成功"
    fi

    if "$PYTHON_CMD" -c "import uvicorn; print(f'uvicorn {uvicorn.__version__} OK')" 2>/dev/null; then
        echo "[INFO] uvicorn 安装成功"
    fi
}

echo ""
echo "[STEP 1/3] 安装系统依赖..."
install_system_deps

echo ""
echo "[STEP 2/3] 安装 Python 依赖（首次可能需要几分钟）..."
install_python_deps

echo ""
echo "[STEP 3/3] 验证安装..."
verify_install

echo ""
echo "========================================"
echo "  安装完成!"
echo ""
echo "  启动服务: ./start_ocr.sh"
echo "  测试健康: curl http://localhost:8899/health"
echo "========================================"
