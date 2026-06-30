#!/usr/bin/env python3
"""
ddddocr-api 独立服务
基于 FastAPI + uvicorn + ddddocr 的验证码识别 API 服务。

用法:
    python3 ocr_server.py                    # 默认 0.0.0.0:8899
    python3 ocr_server.py --host 127.0.0.1 --port 8899 --workers 1

环境变量:
    OCR_HOST      监听地址 (默认 0.0.0.0)
    OCR_PORT      监听端口 (默认 8899)
    OCR_WORKERS   uvicorn worker 数量 (默认 1)
    OCR_LOG_LEVEL uvicorn 日志级别 (默认 info)
"""

import argparse
import base64
import io
import json
import logging
import os
import sys
import time
from threading import Lock
from concurrent.futures import ThreadPoolExecutor

import uvicorn
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse
from pydantic import BaseModel

try:
    import ddddocr
    DDDDOCR_AVAILABLE = True
except ImportError:
    DDDDOCR_AVAILABLE = False
    print("[OCR] 警告: ddddocr 未安装，将返回降级响应", file=sys.stderr)


# ---------- 配置 ----------
DEFAULT_HOST = "0.0.0.0"
DEFAULT_PORT = 8899
DEFAULT_WORKERS = 1
OCR_MODEL_CACHE_DIR = os.path.expanduser("~/.ddddocr")

# ---------- 日志 ----------
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger("ddddocr-api")


# ---------- FastAPI App ----------
app = FastAPI(
    title="ddddocr-api",
    description="ddddocr 验证码识别 API 服务",
    version="1.0.0",
)


# ---------- 全局资源 ----------
ocr_instance = None
ocr_lock = Lock()
_init_error = None


def get_ocr():
    global ocr_instance, _init_error
    if _init_error:
        raise RuntimeError(_init_error)
    if ocr_instance is None:
        with ocr_lock:
            if ocr_instance is None:
                if not DDDDOCR_AVAILABLE:
                    _init_error = "ddddocr 模块未安装，请运行 install.sh 或 pip install ddddocr"
                    raise RuntimeError(_init_error)
                logger.info("正在加载 ddddocr 模型（约需 10-30 秒，首次运行会下载模型）...")
                t0 = time.time()
                ocr_instance = ddddocr.DdddOcr(show_ad=False)
                logger.info(f"ddddocr 模型加载完成，耗时 {time.time()-t0:.1f}s")
    return ocr_instance


# ---------- 路由 ----------


@app.get("/health")
async def health():
    """健康检查"""
    try:
        get_ocr()
    except Exception as e:
        return JSONResponse({"success": False, "msg": str(e)}, status_code=503)
    return {"success": True, "msg": "ddddocr works!"}


class RecognizeB64Request(BaseModel):
    image: str


@app.post("/ocr/recognize/b64")
async def recognize_b64(req: RecognizeB64Request):
    """Base64 图片识别"""
    try:
        img_bytes = base64.b64decode(req.image)
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Base64 解码失败: {e}")

    try:
        ocr = get_ocr()
        t0 = time.time()
        result = ocr.classification(img_bytes)
        processing_time = time.time() - t0
        logger.info(f"识别结果: {result} | 耗时: {processing_time:.3f}s")
        return {"success": True, "result": result, "processing_time": processing_time}
    except Exception as e:
        logger.error(f"识别失败: {e}")
        return JSONResponse(
            {"success": False, "result": None, "error": str(e)},
            status_code=500,
        )


@app.post("/ocr/recognize")
async def recognize_raw(request: Request):
    """原始二进制图片识别（Content-Type: application/octet-stream）"""
    try:
        body = await request.body()
        if not body:
            raise HTTPException(status_code=400, detail="请求体为空")
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"读取请求体失败: {e}")

    try:
        ocr = get_ocr()
        t0 = time.time()
        result = ocr.classification(body)
        processing_time = time.time() - t0
        logger.info(f"识别结果: {result} | 耗时: {processing_time:.3f}s")
        return {"success": True, "result": result, "processing_time": processing_time}
    except Exception as e:
        logger.error(f"识别失败: {e}")
        return JSONResponse(
            {"success": False, "result": None, "error": str(e)},
            status_code=500,
        )


class BatchB64Request(BaseModel):
    images: list[str]


@app.post("/ocr/recognize/batch")
async def recognize_batch(req: BatchB64Request):
    """批量 Base64 图片识别"""
    if not req.images:
        raise HTTPException(status_code=400, detail="images 列表为空")

    results = []
    for i, b64 in enumerate(req.images):
        try:
            img_bytes = base64.b64decode(b64)
            ocr = get_ocr()
            result = ocr.classification(img_bytes)
            results.append({"index": i, "success": True, "result": result})
        except Exception as e:
            results.append({"index": i, "success": False, "error": str(e)})

    return {"success": True, "results": results}


@app.get("/")
async def root():
    return {
        "service": "ddddocr-api",
        "version": "1.0.0",
        "docs": "/docs",
        "health": "/health",
    }


# ---------- 启动入口（独立运行） ----------
def main():
    parser = argparse.ArgumentParser(description="ddddocr-api 独立服务")
    parser.add_argument("--host", default=os.getenv("OCR_HOST", DEFAULT_HOST), help="监听地址")
    parser.add_argument("--port", type=int, default=int(os.getenv("OCR_PORT", DEFAULT_PORT)), help="监听端口")
    parser.add_argument("--workers", type=int, default=int(os.getenv("OCR_WORKERS", DEFAULT_WORKERS)), help="uvicorn worker 数量")
    parser.add_argument("--log-level", default=os.getenv("OCR_LOG_LEVEL", "info"), help="日志级别")
    args = parser.parse_args()

    logger.info(f"启动 ddddocr-api: {args.host}:{args.port} (workers={args.workers})")

    uvicorn.run(
        "ocr_server:app",
        host=args.host,
        port=args.port,
        workers=args.workers,
        log_level=args.log_level,
        access_log=False,
    )


if __name__ == "__main__":
    main()
