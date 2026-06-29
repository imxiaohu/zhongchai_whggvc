# ============================================================
# ddddocr-api Dockerfile
# 用于构建 ddddocr 验证码识别微服务的 Docker 镜像
# ============================================================
# 构建阶段：安装依赖
FROM python:3.10-slim AS builder

WORKDIR /app

# 安装系统依赖（ddddocr 需要 libgl1-mesa-glx/libglib2.0-0）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        libgl1-mesa-glx \
        libglib2.0-0 \
        libgomp1 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 安装 Python 依赖
# 先安装 CPU 优化的 openblas/lapack 以加速 numpy
RUN pip install --no-cache-dir --upgrade pip setuptools wheel

# 安装 ddddocr（包含所有模型）
RUN pip install --no-cache-dir ddddocr==1.4.11

# ============================================================
# 运行阶段：最小化镜像
# ============================================================
FROM python:3.10-slim

WORKDIR /app

# 从 builder 复制已安装的系统依赖
COPY --from=builder /usr/lib /usr/lib
COPY --from=builder /usr/local/lib/python3.10/dist-packages /usr/local/lib/python3.10/dist-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# 运行时必要依赖（SSL、字符编码）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libgl1-mesa-glx \
        libglib2.0-0 \
        libgomp1 \
        ca-certificates \
        fonts-noto-cjk \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Python 运行时依赖（ddddocr 基于 onnxruntime/cv2 需要这些）
RUN pip install --no-cache-dir \
    ddddocr==1.4.11 \
    onnxruntime==1.18.1 \
    opencv-python-headless==4.10.0.84 \
    Pillow==10.4.0 \
    numpy==1.26.4 \
    click==8.1.7 \
    uvicorn==0.32.0 \
    fastapi==0.115.0 \
    pydantic==2.9.0

# 非 root 用户运行（安全加固）
RUN useradd --create-home --shell /bin/bash --gid 0 --uid 1000 ocruser || true
USER ocruser

# 健康检查：每 30s 检查一次 /health 接口
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8899/health')" 2>/dev/null && echo "OK"

# 默认端口（可被环境变量覆盖）
EXPOSE 8899

# 启动命令：默认启动 ddddocr API 服务
# 也可直接运行: python -m ddddocr api --host=0.0.0.0 --port=8899 --workers=2
CMD ["python", "-m", "ddddocr", "api", "--host=0.0.0.0", "--port=8899", "--workers=2"]
