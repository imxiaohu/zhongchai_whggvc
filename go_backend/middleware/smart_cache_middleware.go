package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
)

// SmartCacheConfig 智能缓存配置
type SmartCacheConfig struct {
	// 是否启用缓存
	Enabled bool
	// 缓存的API路径列表
	CachedAPIs []string
	// 是否为公共API（不需要用户ID）
	IsPublicAPI bool
	// 自定义数据获取器
	DataFetcher func(c *gin.Context) (interface{}, error)
}

// SmartCacheMiddleware 智能缓存中间件
func SmartCacheMiddleware(config SmartCacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("智能缓存中间件被触发 - 路径: %s", c.Request.URL.Path)

		// 检查是否启用缓存
		if !config.Enabled {
			log.Printf("智能缓存未启用，跳过")
			c.Next()
			return
		}

		// 检查是否为需要缓存的API
		requestPath := c.Request.URL.Path
		if !isAPIPathCached(requestPath, config.CachedAPIs) {
			log.Printf("API路径不在缓存列表中: %s", requestPath)
			c.Next()
			return
		}

		log.Printf("API路径在缓存列表中，开始处理缓存逻辑")

		// 获取智能缓存服务
		cacheService := services.GetSmartCacheService()

		// 构建缓存请求
		cacheReq, err := buildCacheRequest(c, config.IsPublicAPI)
		if err != nil {
			log.Printf("构建缓存请求失败: %v", err)
			// 如果构建缓存请求失败，继续正常流程
			c.Next()
			return
		}

		log.Printf("缓存请求构建成功 - API: %s, 用户: %s", cacheReq.APIPath, cacheReq.UserID)

		// 定义数据获取器
		var dataFetcher func() (interface{}, error)
		if config.DataFetcher != nil {
			// 使用自定义数据获取器
			dataFetcher = func() (interface{}, error) {
				return config.DataFetcher(c)
			}
		} else {
			// 使用默认数据获取器（继续执行后续处理器）
			dataFetcher = func() (interface{}, error) {
				return executeHandlersAndGetResponse(c)
			}
		}

		// 获取或设置缓存
		cacheResp, err := cacheService.GetOrSetCache(cacheReq, dataFetcher)
		if err != nil {
			// 缓存失败，返回错误
			log.Printf("智能缓存获取失败 - API: %s, 用户: %s, 错误: %v", cacheReq.APIPath, cacheReq.UserID, err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "服务暂时不可用",
				"code":    503,
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		log.Printf("智能缓存获取成功，开始构建响应")

		// 构建响应
		response := buildCacheResponse(cacheResp)

		// 设置响应头
		setCacheHeaders(c, cacheResp)

		log.Printf("智能缓存中间件返回响应，阻止后续处理器执行")

		// 返回响应
		c.JSON(http.StatusOK, response)
		c.Abort()
	}
}

// isAPIPathCached 检查API路径是否需要缓存
func isAPIPathCached(requestPath string, cachedAPIs []string) bool {
	log.Printf("检查API路径是否需要缓存 - 请求路径: %s", requestPath)
	log.Printf("缓存API列表: %v", cachedAPIs)

	if len(cachedAPIs) == 0 {
		log.Printf("缓存API列表为空，默认缓存所有")
		return true // 如果没有指定，默认缓存所有
	}

	for _, apiPath := range cachedAPIs {
		if strings.Contains(requestPath, apiPath) {
			log.Printf("找到匹配的API路径: %s", apiPath)
			return true
		}
	}
	log.Printf("没有找到匹配的API路径")
	return false
}

// buildCacheRequest 构建缓存请求
func buildCacheRequest(c *gin.Context, isPublicAPI bool) (*services.CacheRequest, error) {
	// 获取API路径
	apiPath := c.Request.URL.Path

	// 获取用户ID
	var userID string
	if !isPublicAPI {
		if userIDInterface, exists := c.Get("userId"); exists {
			userID = fmt.Sprintf("%v", userIDInterface)
		}
	}

	// 获取请求参数
	params := make(map[string]interface{})

	// 获取查询参数
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// 获取POST参数（如果是POST请求）
	if c.Request.Method == "POST" {
		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			// 恢复请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// 解析JSON参数
			var postParams map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &postParams); err == nil {
				for key, value := range postParams {
					params[key] = value
				}
			}
		}
	}

	// 获取客户端信息
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	return &services.CacheRequest{
		APIPath:   apiPath,
		UserID:    userID,
		Params:    params,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// executeHandlersAndGetResponse 执行处理器并获取响应
func executeHandlersAndGetResponse(c *gin.Context) (interface{}, error) {
	// 创建响应记录器
	recorder := &responseRecorder{
		ResponseWriter: c.Writer,
		body:           bytes.NewBuffer(nil),
		statusCode:     200, // 默认状态码
	}
	c.Writer = recorder

	// 执行后续处理器
	c.Next()

	// 检查是否有错误
	if len(c.Errors) > 0 {
		return nil, fmt.Errorf("处理请求时发生错误: %v", c.Errors.Last().Error())
	}

	// 检查响应体是否为空
	if recorder.body.Len() == 0 {
		return nil, fmt.Errorf("响应体为空")
	}

	// 解析响应
	var response interface{}
	if err := json.Unmarshal(recorder.body.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return response, nil
}

// responseRecorder 响应记录器
type responseRecorder struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	// 只记录数据到缓冲区，不写入到原始ResponseWriter
	// 这样可以避免双重输出
	return r.body.Write(data)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	// 不调用原始的WriteHeader，避免重复设置
}

// buildCacheResponse 构建缓存响应
func buildCacheResponse(cacheResp *models.CacheResponse) gin.H {
	// 添加调试日志
	log.Printf("构建缓存响应 - CacheType: %s, FromCache: %t, ServerDown: %t",
		cacheResp.CacheType, cacheResp.FromCache, cacheResp.ServerDown)

	// 检查缓存数据是否已经是标准化格式
	if dataMap, ok := cacheResp.Data.(map[string]interface{}); ok {
		if isStandardizedResponse(dataMap) {
			log.Printf("数据已经是标准化格式，更新数据源信息")

			// 如果已经是标准化格式，更新时间戳和数据源信息
			dataMap["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

			// 根据缓存响应更新数据源类型和缓存时间
			dataSourceType := determineDataSourceType(cacheResp)
			dataMap["dataSourceType"] = dataSourceType
			dataMap["cacheUpdatedAt"] = getCacheUpdatedTime(cacheResp)

			// 添加数据来源说明
			if dataSourceType == "school" {
				dataMap["dataSourceNote"] = "学校服务器实时数据"
			} else {
				dataMap["dataSourceNote"] = "智能缓存数据"
			}

			log.Printf("设置 dataSourceType: %s, dataSourceNote: %s",
				dataSourceType, dataMap["dataSourceNote"])

			// 如果服务器离线，更新消息和数据源类型
			if cacheResp.ServerDown {
				dataMap["message"] = "学校服务器维护中，返回缓存数据"
				dataMap["dataSourceType"] = "database"
				dataMap["dataSourceNote"] = "智能缓存数据"
			}

			return dataMap
		}
	}

	log.Printf("数据不是标准化格式，构建新的标准化响应")

	// 首先尝试标准化缓存数据
	standardizedData := standardizeCacheData(cacheResp.Data)

	// 确定数据源类型
	dataSourceType := determineDataSourceType(cacheResp)
	log.Printf("确定数据源类型: %s", dataSourceType)

	// 构建标准化响应
	response := gin.H{
		"success":        true,
		"code":           200,
		"message":        "获取成功",
		"result":         standardizedData,
		"dataSourceType": dataSourceType,
		"cacheUpdatedAt": getCacheUpdatedTime(cacheResp),
		"timestamp":      time.Now().Format("2006-01-02 15:04:05"),
	}

	// 添加数据来源说明
	if dataSourceType == "school" {
		response["dataSourceNote"] = "学校服务器实时数据"
	} else {
		response["dataSourceNote"] = "智能缓存数据"
	}

	log.Printf("设置 dataSourceNote: %s", response["dataSourceNote"])

	// 如果服务器离线，更新消息和数据源类型
	if cacheResp.ServerDown {
		response["message"] = "学校服务器维护中，返回缓存数据"
		response["dataSourceType"] = "database"
		response["dataSourceNote"] = "智能缓存数据"
	}

	return response
}

// isStandardizedResponse 检查响应是否已经是标准化格式
func isStandardizedResponse(data map[string]interface{}) bool {
	// 检查是否包含标准化响应的必需字段
	_, hasSuccess := data["success"]
	_, hasCode := data["code"]
	_, hasMessage := data["message"]
	_, hasResult := data["result"]
	_, hasDataSourceType := data["dataSourceType"]
	_, hasCacheUpdatedAt := data["cacheUpdatedAt"]

	return hasSuccess && hasCode && hasMessage && hasResult && hasDataSourceType && hasCacheUpdatedAt
}

// standardizeCacheData 标准化缓存数据，移除多层嵌套
func standardizeCacheData(data interface{}) interface{} {
	if dataMap, ok := data.(map[string]interface{}); ok {
		// 检查是否需要标准化
		if needsStandardizationFromMap(dataMap) {
			return standardizeResponseFromMap(dataMap)
		}
	}
	return data
}

// needsStandardizationFromMap 检查map数据是否需要标准化
func needsStandardizationFromMap(data map[string]interface{}) bool {
	// 检查是否有多层嵌套的result字段
	if result, ok := data["result"].(map[string]interface{}); ok {
		if _, hasNestedResult := result["result"]; hasNestedResult {
			return true
		}
	}
	return false
}

// standardizeResponseFromMap 标准化map格式的响应数据
func standardizeResponseFromMap(data map[string]interface{}) interface{} {
	// 递归提取最深层的result数据
	current := data
	var finalResult interface{}

	// 持续深入直到找到最终的数据层
	for {
		if result, ok := current["result"].(map[string]interface{}); ok {
			// 检查是否还有更深层的result
			if nestedResult, hasNested := result["result"]; hasNested {
				// 如果嵌套的result是map，继续深入
				if _, isMap := nestedResult.(map[string]interface{}); isMap {
					current = result
					continue
				} else {
					// 如果嵌套的result不是map，说明这是最终数据
					finalResult = nestedResult
					break
				}
			} else {
				// 没有更深层的result，这就是最终数据
				finalResult = result
				break
			}
		} else {
			// 当前层没有result字段，返回当前层
			finalResult = current
			break
		}
	}

	return finalResult
}

// determineDataSourceType 确定数据源类型
func determineDataSourceType(cacheResp *models.CacheResponse) string {
	if cacheResp.ServerDown || cacheResp.CacheType == "database" {
		return "database"
	}
	return "school"
}

// getCacheUpdatedTime 获取缓存更新时间
func getCacheUpdatedTime(cacheResp *models.CacheResponse) string {
	if !cacheResp.UpdatedAt.IsZero() {
		return cacheResp.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return cacheResp.Timestamp.Format("2006-01-02 15:04:05")
}

// setCacheHeaders 设置缓存相关的HTTP头
func setCacheHeaders(c *gin.Context, cacheResp *models.CacheResponse) {
	// 设置缓存类型头
	c.Header("X-Cache-Type", cacheResp.CacheType)
	c.Header("X-From-Cache", strconv.FormatBool(cacheResp.FromCache))
	c.Header("X-Server-Down", strconv.FormatBool(cacheResp.ServerDown))
	c.Header("X-Cache-Timestamp", cacheResp.Timestamp.Format(time.RFC3339))

	// 设置缓存控制头
	if cacheResp.FromCache {
		// 如果是缓存数据，设置较短的缓存时间
		c.Header("Cache-Control", "public, max-age=300") // 5分钟
	} else {
		// 如果是新数据，设置较长的缓存时间
		c.Header("Cache-Control", "public, max-age=1800") // 30分钟
	}
}

// 预定义的智能缓存配置

// PublicAPISmartCache 公共API智能缓存配置
// 仅缓存公共数据，不包含用户个人数据
func PublicAPISmartCache() SmartCacheConfig {
	return SmartCacheConfig{
		Enabled:     true,
		IsPublicAPI: true,
		CachedAPIs: []string{
			"/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime",      // 当前周数据
			"/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime", // 课程表时间段配置
			"/scloudoa/scs/course/tCourseScore/getSemester",                   // 学期列表
			"/api/m/scs/course/tCourseScore/getSemester",                      // 学期列表（移动端）
			"/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList",               // 新闻类型列表
			"/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId",               // 新闻列表
			"/scloudoa/sys/announcement/list",                                 // 公告列表
			"/scloudoa/sys/news/list",                                         // 系统新闻
			"/scloudoa/sys/notice/list",                                       // 通知列表
		},
	}
}

// UserAPISmartCache 用户API智能缓存配置
// 注意：此配置已废弃，智能缓存仅处理公共数据
// 用户个人数据（如课程表、成绩等）应使用原有的个人同步缓存机制
func UserAPISmartCache() SmartCacheConfig {
	return SmartCacheConfig{
		Enabled:     false, // 禁用用户API智能缓存
		IsPublicAPI: false,
		CachedAPIs:  []string{}, // 清空缓存API列表
	}
}

// CustomSmartCache 自定义智能缓存配置
func CustomSmartCache(apiPaths []string, isPublic bool, dataFetcher func(c *gin.Context) (interface{}, error)) SmartCacheConfig {
	return SmartCacheConfig{
		Enabled:     true,
		IsPublicAPI: isPublic,
		CachedAPIs:  apiPaths,
		DataFetcher: dataFetcher,
	}
}

// 注意：课程表数据获取器已移除
// 课程表数据应使用原有的用户个人同步缓存机制，不使用智能缓存

// HealthCheckSmartCache 带健康检查的智能缓存
func HealthCheckSmartCache(config SmartCacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查学校服务器健康状态
		healthService := services.GetSchoolHealthCheckService()

		// 如果服务器离线，强制使用缓存
		if !healthService.IsServerAlive() {
			config.Enabled = true
			SmartCacheMiddleware(config)(c)
			return
		}

		// 服务器正常，按配置执行
		SmartCacheMiddleware(config)(c)
	}
}
