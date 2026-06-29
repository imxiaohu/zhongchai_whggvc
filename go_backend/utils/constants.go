package utils

import (
	"time"
)

// =============================================================================
// 时间常量 (time.Duration)
// 所有硬编码的时间常量必须使用此处的命名常量，禁止裸写数字
// =============================================================================

const (
	// HTTP 请求超时
	DefaultHTTPTimeout = 30 * time.Second

	// CORS 配置
	CORSMaxAge = 12 * time.Hour

	// 健康检查
	HealthCheckInterval = 5 * time.Minute
	HealthCheckTimeout  = 10 * time.Second

	// PC 会话
	PCSessionTTL = 30 * time.Minute
)
