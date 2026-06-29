package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/services"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL           time.Duration             // 缓存过期时间
	KeyGenerator  func(*gin.Context) string // 缓存键生成器
	ShouldCache   func(*gin.Context) bool   // 是否应该缓存
	IgnoreHeaders []string                  // 忽略的请求头
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() CacheConfig {
	config := CacheConfig{
		TTL: 5 * time.Minute, // 默认5分钟
		KeyGenerator: func(c *gin.Context) string {
			if c == nil {
				return "default_key"
			}
			return GenerateDefaultCacheKey(c)
		},
		ShouldCache: func(c *gin.Context) bool {
			if c == nil || c.Request == nil {
				return false
			}
			// 只缓存GET请求
			return c.Request.Method == "GET"
		},
		IgnoreHeaders: []string{
			"Authorization",
			"X-Access-Token",
			"Cookie",
			"User-Agent",
			"Accept-Encoding",
			"Accept-Language",
		},
	}

	// 验证配置的完整性
	if config.KeyGenerator == nil || config.ShouldCache == nil {
		log.Printf("警告：缓存配置不完整")
	}

	return config
}

// CacheMiddleware 缓存中间件
func CacheMiddleware(config ...CacheConfig) gin.HandlerFunc {
	cfg := DefaultCacheConfig()
	if len(config) > 0 {
		// 合并配置，保留默认值
		userConfig := config[0]
		if userConfig.TTL > 0 {
			cfg.TTL = userConfig.TTL
		}
		if userConfig.KeyGenerator != nil {
			cfg.KeyGenerator = userConfig.KeyGenerator
		}
		if userConfig.ShouldCache != nil {
			cfg.ShouldCache = userConfig.ShouldCache
		}
		if userConfig.IgnoreHeaders != nil {
			cfg.IgnoreHeaders = userConfig.IgnoreHeaders
		}
	}

	return func(c *gin.Context) {
		// 防护性检查：确保配置不为空
		if cfg.ShouldCache == nil {
			log.Printf("缓存中间件配置错误：ShouldCache函数为空")
			c.Next()
			return
		}

		// 检查是否应该缓存
		if !cfg.ShouldCache(c) {
			c.Next()
			return
		}

		// 获取缓存服务
		cacheService := services.GetCacheService()
		if cacheService == nil {
			log.Printf("缓存服务未初始化，跳过缓存")
			c.Next()
			return
		}

		// 生成缓存键
		if cfg.KeyGenerator == nil {
			log.Printf("缓存中间件配置错误：KeyGenerator函数为空")
			c.Next()
			return
		}
		cacheKey := cfg.KeyGenerator(c)

		// 尝试从缓存获取数据
		if cachedData, exists := cacheService.Get(cacheKey); exists {
			log.Printf("缓存命中: %s", cacheKey)

			// 设置响应头表明这是缓存的响应
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey)
			c.Header("X-Cache-Type", cacheService.GetCacheType())

			// 返回缓存的数据
			c.Data(http.StatusOK, "application/json", cachedData)
			c.Abort()
			return
		}

		// 缓存未命中，继续处理请求
		log.Printf("缓存未命中: %s", cacheKey)

		// 创建响应写入器来捕获响应数据
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 检查响应状态码，只缓存成功的响应
		if writer.Status() == http.StatusOK && writer.body.Len() > 0 {
			// 将响应数据存入缓存
			responseData := writer.body.Bytes()
			cacheService.Set(cacheKey, responseData, cfg.TTL)

			log.Printf("响应已缓存: %s, 大小: %d bytes", cacheKey, len(responseData))

			// 设置响应头表明这是新的响应
			c.Header("X-Cache", "MISS")
			c.Header("X-Cache-Key", cacheKey)
			c.Header("X-Cache-Type", cacheService.GetCacheType())
		}
	}
}

// responseWriter 自定义响应写入器，用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	// 同时写入原始响应和缓冲区
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	// 同时写入原始响应和缓冲区
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// GenerateDefaultCacheKey 生成默认的缓存键
func GenerateDefaultCacheKey(c *gin.Context) string {
	// 防护性检查
	if c == nil || c.Request == nil || c.Request.URL == nil {
		log.Printf("警告：GenerateDefaultCacheKey收到空的上下文或请求")
		return "invalid_context"
	}

	// 获取用户ID
	userID := ""
	if uid, exists := c.Get("userId"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	// 构建基础键
	path := c.Request.URL.Path
	method := c.Request.Method

	// 获取查询参数并排序
	queryParams := make([]string, 0)
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sort.Strings(queryParams)

	// 获取POST参数（如果是POST请求）
	postParams := make([]string, 0)
	if method == "POST" {
		// 读取请求体
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 重新设置请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 如果是JSON格式，直接使用内容的哈希
				if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
					hash := md5.Sum(bodyBytes)
					postParams = append(postParams, fmt.Sprintf("body_hash=%x", hash))
				}
			}
		}

		// 也包含表单参数
		if err := c.Request.ParseForm(); err == nil {
			for key, values := range c.Request.PostForm {
				for _, value := range values {
					postParams = append(postParams, fmt.Sprintf("%s=%s", key, value))
				}
			}
		}
	}
	sort.Strings(postParams)

	// 构建完整的缓存键
	keyParts := []string{
		fmt.Sprintf("method:%s", method),
		fmt.Sprintf("path:%s", path),
	}

	if userID != "" {
		keyParts = append(keyParts, fmt.Sprintf("user:%s", userID))
	}

	if len(queryParams) > 0 {
		keyParts = append(keyParts, fmt.Sprintf("query:%s", strings.Join(queryParams, "&")))
	}

	if len(postParams) > 0 {
		keyParts = append(keyParts, fmt.Sprintf("post:%s", strings.Join(postParams, "&")))
	}

	// 生成最终的缓存键
	cacheKey := strings.Join(keyParts, "|")

	// 如果键太长，使用MD5哈希
	if len(cacheKey) > 200 {
		hash := md5.Sum([]byte(cacheKey))
		cacheKey = fmt.Sprintf("hash:%x", hash)
	}

	return cacheKey
}

// CacheKeyForCurrentTime 为getCurrentTime接口生成缓存键
func CacheKeyForCurrentTime(c *gin.Context) string {
	userID := ""
	if uid, exists := c.Get("userId"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	// getCurrentTime的缓存键相对简单，主要基于用户和时间
	// 由于时间信息变化不频繁，可以使用较长的缓存时间
	return fmt.Sprintf("current_time:user:%s", userID)
}

// CacheKeyForTimetableDay 为getTimetableDay接口生成缓存键
func CacheKeyForTimetableDay(c *gin.Context) string {
	userID := ""
	if uid, exists := c.Get("userId"); exists {
		if uidStr, ok := uid.(string); ok {
			userID = uidStr
		}
	}

	// 获取日期参数
	date := c.Query("date")
	if date == "" {
		date = c.PostForm("date")
	}

	// 获取周次参数
	week := c.Query("week")
	if week == "" {
		week = c.PostForm("week")
	}

	// 构建缓存键
	keyParts := []string{
		"timetable_day",
		fmt.Sprintf("user:%s", userID),
	}

	if date != "" {
		keyParts = append(keyParts, fmt.Sprintf("date:%s", date))
	}

	if week != "" {
		keyParts = append(keyParts, fmt.Sprintf("week:%s", week))
	}

	return strings.Join(keyParts, ":")
}
