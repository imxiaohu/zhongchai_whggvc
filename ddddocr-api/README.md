# ddddocr-api 验证码识别微服务

基于 [ddddocr](https://github.com/sml2h3/ddddocr) 的轻量级验证码识别 API 服务，支持 Docker 一键部署。

## 快速开始

### 1. 启动服务（开发/测试）

```bash
cd ddddocr-api

# 复制环境变量配置
cp .env.example .env

# 构建并启动（首次会自动下载 ~500MB 模型文件）
docker compose --profile dev up -d

# 查看日志
docker compose logs -f ddddocr-api

# 验证服务健康
curl http://localhost:8899/health
```

### 2. 生产环境部署

```bash
# 同样复制环境变量
cp .env.example .env

# 生产环境通常不需要 dev profile
docker compose --profile prod up -d

# 配置开机自启（systemd）
sudo systemctl edit --full --force docker-compose.service
```

systemd 示例配置：

```ini
[Unit]
Description=Docker Compose ddddocr-api
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/path/to/pingjiao/ddddocr-api
ExecStart=/usr/local/bin/docker compose --profile prod up -d
ExecStop=/usr/local/bin/docker compose down
TimeoutStartSec=300

[Install]
WantedBy=multi-user.target
```

## API 接口

ddddocr API 服务提供以下接口：

### 健康检查

```
GET /health
```

返回：
```json
{"success": true, "msg": "ddddocr works!"}
```

### 识别验证码（二进制）

```
POST /ocr/recognize
Content-Type: application/octet-stream

<binary image data>
```

返回：
```json
{"success": true, "result": "AB12", "error": ""}
```

### 识别验证码（Base64）

```
POST /ocr/recognize/b64
Content-Type: application/json

{"image": "<base64 encoded image>"}
```

返回：
```json
{"success": true, "result": "AB12", "error": ""}
```

### 批量识别

```
POST /ocr/recognize/batch
Content-Type: application/json

{"images": ["<b64 img1>", "<b64 img2>"]}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `OCR_PORT` | `8899` | 服务端口 |
| `OCR_WORKERS` | `2` | Worker 数量，每个约 300MB 内存 |
| `OCR_BETA` | `false` | 是否使用 Beta 模型（更高识别率） |
| `OCR_USE_GPU` | `false` | 是否启用 GPU 加速 |
| `OCR_CPU_LIMIT` | `2` | CPU 核心限制 |
| `OCR_MEM_LIMIT` | `2G` | 内存限制 |

## 与 Go 后端集成

Go 后端通过 `OCR_SERVICE_URL` 环境变量连接本服务：

```bash
# go_backend/.env
OCR_SERVICE_URL=http://127.0.0.1:8899
```

### Docker Compose 一体化部署

将 OCR 服务集成到主项目 docker-compose：

```yaml
services:
  ddddocr-api:
    build: ./ddddocr-api
    container_name: ddddocr-api
    ports:
      - "127.0.0.1:8899:8899"   # 仅本地监听，对外不暴露
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
    volumes:
      - ddddocr-cache:/root/.ddddocr
    networks:
      - backend-network

  pingjiao-backend:
    image: your-pingjiao-backend:latest
    container_name: pingjiao-backend
    depends_on:
      ddddocr-api:
        condition: service_healthy
    environment:
      - OCR_SERVICE_URL=http://ddddocr-api:8899
    networks:
      - backend-network

volumes:
  ddddocr-cache:
```

### 资源规划

| 场景 | CPU | 内存 | Workers |
|------|-----|------|---------|
| 开发/测试 | 1核 | 1G | 1 |
| 小规模生产 | 2核 | 2G | 2 |
| 中等规模（~100 QPS） | 4核 | 4G | 3 |
| 大规模（>200 QPS） | 8核+GPU | 8G+ | 4+GPU |

## 内存占用

ddddocr 基于 ONNX Runtime，首次加载模型约占用 200-400MB 内存，运行时会根据并发自动扩缩。

**实测内存占用：**
- 冷启动后：~300MB
- 每次识别：+5-10MB（GC 后释放）
- 2 workers 稳定运行：~600MB

## 故障排查

### 首次启动下载模型失败

```bash
# 手动下载模型
docker compose run --rm ddddocr-api python -c "import ddddocr; ddddocr.DdddOcr()"
```

### 服务启动但健康检查失败

```bash
# 查看详细日志
docker compose logs ddddocr-api

# 手动测试识别
curl -X POST http://localhost:8899/ocr/recognize \
  -H "Content-Type: application/octet-stream" \
  --data-binary @test.png
```

### GPU 不可用

确认已安装 [nvidia-container-toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html)：

```bash
docker run --rm --gpus all nvidia/cuda:11.8-base-ubuntu22.04 nvidia-smi
```

## 安全建议

1. **内网访问**：OCR 服务无需公网访问，仅供后端内网调用，端口绑定 `127.0.0.1`
2. **非 root 用户**：Dockerfile 已配置使用 `ocruser` 运行
3. **定期更新**：定期 `docker compose pull` 获取最新版 ddddocr 以提升识别率
4. **资源限制**：生产环境务必配置 CPU/内存限制，防止 OOM
