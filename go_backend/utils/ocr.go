package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/config"
)

// OCRService ddddocr 识别服务客户端，支持重试、健康检查与降级
type OCRService struct {
	BaseURL string
	Client  *http.Client
	healthy bool
	mu      sync.RWMutex
}

// ddddocr API 响应格式
type ddddocrResponse struct {
	Result         interface{} `json:"result"`
	ProcessingTime float64     `json:"processing_time"`
}

var (
	_ocrService *OCRService
	_once       sync.Once
)

// GetOCRService 获取全局单例 OCR 服务实例
// 线程安全，首次调用后返回同一实例
func GetOCRService() *OCRService {
	_once.Do(func() {
		timeout := time.Duration(config.GetEnvInt("OCR_TIMEOUT_MS", 15000)) * time.Millisecond
		_ocrService = &OCRService{
			BaseURL: config.GetOCRServURL(),
			Client: &http.Client{
				Timeout: timeout,
				Transport: &http.Transport{
					MaxIdleConns:        10,
					MaxIdleConnsPerHost: 5,
					IdleConnTimeout:     90 * time.Second,
				},
			},
			healthy: true,
		}
	})
	return _ocrService
}

// IsHealthy 返回 OCR 服务是否可达（用于监控）
func (s *OCRService) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthy
}

// HealthCheck 探测 OCR 服务是否可用（不计入业务重试）
// 调用后更新内部健康状态
func (s *OCRService) HealthCheck() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.BaseURL+"/health", nil)
	if err != nil {
		s.setHealthy(false)
		return false
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		s.setHealthy(false)
		return false
	}
	defer resp.Body.Close()

	s.setHealthy(resp.StatusCode == http.StatusOK)
	return resp.StatusCode == http.StatusOK
}

func (s *OCRService) setHealthy(ok bool) {
	s.mu.Lock()
	s.healthy = ok
	s.mu.Unlock()
}

// RecognizeFromBytes 将图片字节流发送给 ddddocr API 识别
// 内部自动重试最多 maxRetries 次（指数退避），仍然失败则返回 error
func (s *OCRService) RecognizeFromBytes(imageData []byte) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("图片数据为空")
	}

	var lastErr error
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := s.doRecognize(imageData)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if attempt < maxRetries {
			backoff := time.Duration(attempt*attempt*100) * time.Millisecond
			time.Sleep(backoff)
		}
	}

	return "", fmt.Errorf("OCR识别失败（已重试%d次）: %w", maxRetries, lastErr)
}

// doRecognize 单次识别请求
// ddddocr API: POST /ocr 接受 JSON {"image": "base64..."} 返回 {"result": "...", "processing_time": ...}
func (s *OCRService) doRecognize(imageData []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 构建 JSON 请求体
	b64 := base64.StdEncoding.EncodeToString(imageData)
	payload := map[string]interface{}{"image": b64}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.BaseURL+"/ocr", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OCR服务请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", fmt.Errorf("读取OCR响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// ddddocr 错误时返回 {"detail": "错误信息"}
		var errResp map[string]string
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("OCR识别返回失败: %s", errResp["detail"])
		}
		return "", fmt.Errorf("OCR服务返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var result ddddocrResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析OCR响应失败: %w, body=%s", err, string(body))
	}

	if result.Result == nil {
		return "", fmt.Errorf("OCR识别返回空结果")
	}

	resultStr, ok := result.Result.(string)
	if !ok {
		return "", fmt.Errorf("OCR识别结果格式异常")
	}

	if resultStr == "" {
		return "", fmt.Errorf("OCR识别结果为空字符串")
	}

	return resultStr, nil
}
