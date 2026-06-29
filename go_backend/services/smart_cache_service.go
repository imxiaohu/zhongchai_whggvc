package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiaohu/pingjiao/models"
	"gorm.io/gorm"
)

// SmartCacheService 智能缓存服务
type SmartCacheService struct {
	db          *gorm.DB
	redisClient *redis.Client
	config      *models.CacheConfig
	healthCheck *SchoolHealthCheckService
}

// 全局智能缓存服务实例
var globalSmartCacheService *SmartCacheService
var smartCacheOnce sync.Once

// GetSmartCacheService 获取智能缓存服务实例
func GetSmartCacheService() *SmartCacheService {
	smartCacheOnce.Do(func() {
		globalSmartCacheService = NewSmartCacheService()
	})
	return globalSmartCacheService
}

// NewSmartCacheService 创建智能缓存服务
func NewSmartCacheService() *SmartCacheService {
	// 获取Redis客户端
	redisService := GetRedisService()
	var redisClient *redis.Client
	if redisService != nil && redisService.IsEnabled() {
		redisClient = redisService.GetClient()
	}

	service := &SmartCacheService{
		db:          models.DB,
		redisClient: redisClient,
		config:      models.GetDefaultCacheConfig(),
		healthCheck: GetSchoolHealthCheckService(),
	}

	// 启动后台任务
	go service.startBackgroundTasks()

	return service
}

// CacheRequest 缓存请求结构
type CacheRequest struct {
	APIPath   string                 `json:"apiPath"`
	UserID    string                 `json:"userId"`
	Params    map[string]interface{} `json:"params"`
	IPAddress string                 `json:"ipAddress"`
	UserAgent string                 `json:"userAgent"`
}

// GetOrSetCache 获取或设置缓存
func (s *SmartCacheService) GetOrSetCache(req *CacheRequest, dataFetcher func() (interface{}, error)) (*models.CacheResponse, error) {
	// 1. 构建缓存键
	requestHash := models.GenerateRequestHash(req.Params)
	keyBuilder := &models.CacheKeyBuilder{
		APIPath:     req.APIPath,
		UserID:      req.UserID,
		RequestHash: requestHash,
	}

	cacheKey := keyBuilder.BuildKey()
	redisKey := keyBuilder.BuildRedisKey()

	// 2. 检查学校服务器健康状态
	isServerAlive := s.healthCheck.IsServerAlive()

	// 3. 尝试从Redis获取热点缓存
	if redisData, err := s.getFromRedis(redisKey); err == nil {
		log.Printf("从Redis缓存获取数据: %s", cacheKey)
		s.recordCacheAccess(cacheKey, req)

		// 尝试获取原始缓存数据的更新时间
		var updatedAt time.Time
		if dbCache, err := models.GetCacheDataByKey(cacheKey); err == nil {
			updatedAt = dbCache.UpdatedAt
		} else {
			updatedAt = time.Now() // 如果无法获取，使用当前时间
		}

		return &models.CacheResponse{
			Data:       redisData,
			FromCache:  true,
			CacheType:  "redis",
			ServerDown: !isServerAlive,
			Timestamp:  time.Now(),
			UpdatedAt:  updatedAt,
		}, nil
	}

	// 4. 尝试从数据库获取缓存
	if dbCache, err := models.GetCacheDataByKey(cacheKey); err == nil {
		log.Printf("从数据库缓存获取数据: %s", cacheKey)

		// 更新访问统计
		s.recordCacheAccess(cacheKey, req)

		// 检查是否需要提升到Redis
		s.checkAndPromoteToRedis(dbCache, redisKey)

		var data interface{}
		if err := json.Unmarshal([]byte(dbCache.Data), &data); err == nil {
			// 获取缓存更新时间
			updatedAt := dbCache.UpdatedAt

			// 如果是课表API且有用户ID，尝试获取用户最新同步时间
			if req.UserID != "" && (req.APIPath == "/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek" ||
				req.APIPath == "/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByDay") {

				// 解析用户ID
				if userIdUint, err := strconv.ParseUint(req.UserID, 10, 32); err == nil {
					// 获取用户最新同步时间
					if latestSyncTime, err := models.GetLatestSyncTimeByUserID(uint(userIdUint)); err == nil && latestSyncTime != nil {
						log.Printf("使用用户 %s 的最新同步时间: %v", req.UserID, latestSyncTime)
						updatedAt = *latestSyncTime
					} else {
						log.Printf("获取用户 %s 最新同步时间失败: %v", req.UserID, err)
					}
				}
			}

			return &models.CacheResponse{
				Data:       data,
				FromCache:  true,
				CacheType:  "database",
				ServerDown: !isServerAlive,
				Timestamp:  time.Now(),
				UpdatedAt:  updatedAt,
			}, nil
		}
	}

	// 5. 如果服务器离线，返回错误
	if !isServerAlive {
		return &models.CacheResponse{
			Data:       nil,
			FromCache:  false,
			CacheType:  "none",
			ServerDown: true,
			Timestamp:  time.Now(),
			UpdatedAt:  time.Time{}, // 空时间
		}, fmt.Errorf("学校服务器离线且无缓存数据")
	}

	// 6. 从学校服务器获取新数据
	log.Printf("从学校服务器获取新数据: %s", cacheKey)
	freshData, err := dataFetcher()
	if err != nil {
		return nil, fmt.Errorf("获取新数据失败: %w", err)
	}

	// 7. 检查响应是否为错误响应，不缓存错误数据
	if !s.isValidResponse(freshData) {
		log.Printf("检测到错误响应，不进行缓存: %s", cacheKey)
		return nil, fmt.Errorf("学校服务器返回错误响应")
	}

	// 8. 保存到数据库缓存
	s.saveToDatabase(cacheKey, req, freshData)

	// 8. 记录访问
	s.recordCacheAccess(cacheKey, req)

	currentTime := time.Now()
	return &models.CacheResponse{
		Data:       freshData,
		FromCache:  false,
		CacheType:  "fresh",
		ServerDown: false,
		Timestamp:  currentTime,
		UpdatedAt:  currentTime, // 新数据的更新时间就是当前时间
	}, nil
}

// getFromRedis 从Redis获取数据
func (s *SmartCacheService) getFromRedis(redisKey string) (interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("Redis客户端未初始化")
	}

	ctx := context.Background()

	dataStr, err := s.redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal([]byte(dataStr), &data)
	return data, err
}

// saveToDatabase 保存到数据库
func (s *SmartCacheService) saveToDatabase(cacheKey string, req *CacheRequest, data interface{}) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Printf("序列化数据失败: %v", err)
		return
	}

	// 确定缓存类型
	cacheType := "user"
	if s.config.IsPublicAPI(req.APIPath) {
		cacheType = "public"
	}

	// 创建或更新缓存数据
	cacheData := &models.CacheData{
		CacheKey:    cacheKey,
		CacheType:   cacheType,
		UserID:      req.UserID,
		APIPath:     req.APIPath,
		RequestHash: models.GenerateRequestHash(req.Params),
		Data:        string(dataJSON),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	// 尝试更新现有记录
	if existingCache, err := models.GetCacheDataByKey(cacheKey); err == nil {
		existingCache.Data = string(dataJSON)
		existingCache.LastAccess = time.Now()
		//nolint:errcheck
		models.UpdateCacheData(existingCache)
	} else {
		//nolint:errcheck
		models.CreateCacheData(cacheData)
	}
}

// checkAndPromoteToRedis 检查并提升到Redis
func (s *SmartCacheService) checkAndPromoteToRedis(dbCache *models.CacheData, redisKey string) {
	if dbCache.AccessCount >= s.config.HotDataThreshold && s.redisClient != nil {
		log.Printf("提升热点数据到Redis: %s (访问次数: %d)", dbCache.CacheKey, dbCache.AccessCount)

		ctx := context.Background()
		err := s.redisClient.Set(ctx, redisKey, dbCache.Data, s.config.RedisExpireTime).Err()
		if err != nil {
			log.Printf("保存到Redis失败: %v", err)
			return
		}

		// 记录热点缓存
		hotCache := &models.RedisHotCache{
			CacheKey:    dbCache.CacheKey,
			AccessCount: dbCache.AccessCount,
			IsHot:       true,
			HotSince:    time.Now(),
		}
		models.DB.Save(hotCache)
	}
}

// recordCacheAccess 记录缓存访问
func (s *SmartCacheService) recordCacheAccess(cacheKey string, req *CacheRequest) {
	// 增加访问次数
	//nolint:errcheck
	models.IncrementCacheAccess(cacheKey)

	// 记录访问日志
	accessLog := &models.CacheAccessLog{
		CacheKey:  cacheKey,
		UserID:    req.UserID,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	//nolint:errcheck
	models.CreateCacheAccessLog(accessLog)

	// 更新统计
	//nolint:errcheck
	models.UpdateCacheStats(cacheKey, 0) // 响应时间在这里设为0，实际应该传入真实值
}

// startBackgroundTasks 启动后台任务
func (s *SmartCacheService) startBackgroundTasks() {
	// 每小时清理一次过期数据
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupExpiredData()
	}
}

// cleanupExpiredData 清理过期数据
func (s *SmartCacheService) cleanupExpiredData() {
	log.Println("开始清理过期缓存数据...")

	// 清理数据库中的旧数据
	err := models.CleanupOldCacheData(s.config.DatabaseCleanupDays)
	if err != nil {
		log.Printf("清理数据库缓存失败: %v", err)
	}

	log.Println("缓存数据清理完成")
}

// GetCacheMetrics 获取缓存指标
func (s *SmartCacheService) GetCacheMetrics() (*models.CacheMetrics, error) {
	var totalEntries, dbEntries int64

	// 统计数据库缓存条目
	models.DB.Model(&models.CacheData{}).Count(&dbEntries)
	totalEntries = dbEntries

	// 统计Redis缓存条目（这里简化处理）
	var redisEntries int64 = 0
	if s.redisClient != nil {
		ctx := context.Background()
		redisKeys, err := s.redisClient.Keys(ctx, "hot_cache:*").Result()
		if err == nil {
			redisEntries = int64(len(redisKeys))
		}
	}
	totalEntries += redisEntries

	// 计算命中率等指标（这里简化处理）
	metrics := &models.CacheMetrics{
		TotalCacheEntries:    totalEntries,
		DatabaseCacheEntries: dbEntries,
		RedisCacheEntries:    redisEntries,
		HitRate:              0.85, // 示例值
		MissRate:             0.15, // 示例值
		AvgResponseTime:      120,  // 示例值
	}

	return metrics, nil
}

// ClearCache 清理指定缓存
func (s *SmartCacheService) ClearCache(cacheKey string) error {
	// 从数据库删除
	err := models.DB.Where("cache_key = ?", cacheKey).Delete(&models.CacheData{}).Error
	if err != nil {
		return err
	}

	// 从Redis删除
	if s.redisClient != nil {
		ctx := context.Background()
		redisKey := fmt.Sprintf("hot_cache:%s", cacheKey)
		s.redisClient.Del(ctx, redisKey)
	}

	return nil
}

// ClearAllCache 清理所有缓存
func (s *SmartCacheService) ClearAllCache() error {
	// 清理数据库
	err := models.DB.Where("1 = 1").Delete(&models.CacheData{}).Error
	if err != nil {
		return err
	}

	// 清理Redis
	if s.redisClient != nil {
		ctx := context.Background()
		keys, err := s.redisClient.Keys(ctx, "hot_cache:*").Result()
		if err == nil && len(keys) > 0 {
			s.redisClient.Del(ctx, keys...)
		}
	}

	return nil
}

// isValidResponse 检查响应是否为有效响应（不是错误响应）
func (s *SmartCacheService) isValidResponse(data interface{}) bool {
	// 检查是否为map类型
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return true // 非map类型认为是有效的
	}

	// 检查是否包含错误字段
	if status, exists := dataMap["status"]; exists {
		if statusCode, ok := status.(float64); ok && statusCode >= 400 {
			log.Printf("检测到HTTP错误状态码: %.0f", statusCode)
			return false
		}
	}

	// 检查是否包含error字段
	if errorField, exists := dataMap["error"]; exists && errorField != nil {
		log.Printf("检测到错误字段: %v", errorField)
		return false
	}

	// 检查success字段
	if success, exists := dataMap["success"]; exists {
		if successBool, ok := success.(bool); ok && !successBool {
			log.Printf("检测到success=false")
			return false
		}
	}

	// 检查message字段是否包含错误信息
	if message, exists := dataMap["message"]; exists {
		if messageStr, ok := message.(string); ok {
			if messageStr == "No message available" ||
				messageStr == "Not Found" ||
				messageStr == "Unauthorized" ||
				messageStr == "Internal Server Error" {
				log.Printf("检测到错误消息: %s", messageStr)
				return false
			}
		}
	}

	return true
}
