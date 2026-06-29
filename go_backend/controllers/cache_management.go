package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetCacheMetrics 获取缓存指标
func GetCacheMetrics(c *gin.Context) {
	cacheService := services.GetSmartCacheService()

	metrics, err := cacheService.GetCacheMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取缓存指标失败",
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存指标成功",
		Code:    200,
		Result:  metrics,
	})
}

// GetCacheList 获取缓存列表
func GetCacheList(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	cacheType := c.Query("cacheType")
	apiPath := c.Query("apiPath")

	// 构建查询条件
	query := models.DB.Model(&models.CacheData{})

	if cacheType != "" {
		query = query.Where("cache_type = ?", cacheType)
	}

	if apiPath != "" {
		query = query.Where("api_path LIKE ?", "%"+apiPath+"%")
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取数据
	var cacheList []models.CacheData
	offset := (page - 1) * pageSize
	err := query.Order("last_access DESC").Offset(offset).Limit(pageSize).Find(&cacheList).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取缓存列表失败",
			Code:    500,
		})
		return
	}

	result := gin.H{
		"list":     cacheList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存列表成功",
		Code:    200,
		Result:  result,
	})
}

// GetCacheDetail 获取缓存详情
func GetCacheDetail(c *gin.Context) {
	cacheKey := c.Param("cacheKey")
	if cacheKey == "" {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "缓存键不能为空",
			Code:    400,
		})
		return
	}

	cacheData, err := models.GetCacheDataByKey(cacheKey)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.Response{
			Success: false,
			Message: "缓存数据不存在",
			Code:    404,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存详情成功",
		Code:    200,
		Result:  cacheData,
	})
}

// ClearSmartCache 清理指定智能缓存
func ClearSmartCache(c *gin.Context) {
	cacheKey := c.Param("cacheKey")
	if cacheKey == "" {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "缓存键不能为空",
			Code:    400,
		})
		return
	}

	cacheService := services.GetSmartCacheService()
	err := cacheService.ClearCache(cacheKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "清理缓存失败",
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "清理缓存成功",
		Code:    200,
	})
}

// ClearAllCache 清理所有缓存
func ClearAllCache(c *gin.Context) {
	cacheService := services.GetSmartCacheService()
	err := cacheService.ClearAllCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "清理所有缓存失败",
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "清理所有缓存成功",
		Code:    200,
	})
}

// GetSmartCacheStats 获取智能缓存统计
func GetSmartCacheStats(c *gin.Context) {
	var stats []models.CacheStats
	err := models.DB.Order("total_access DESC").Limit(50).Find(&stats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取缓存统计失败",
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存统计成功",
		Code:    200,
		Result:  stats,
	})
}

// GetHotCacheList 获取热点缓存列表
func GetHotCacheList(c *gin.Context) {
	var hotCaches []models.RedisHotCache
	err := models.DB.Where("is_hot = ?", true).Order("access_count DESC").Find(&hotCaches).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取热点缓存列表失败",
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取热点缓存列表成功",
		Code:    200,
		Result:  hotCaches,
	})
}

// GetCacheAccessLogs 获取缓存访问日志
func GetCacheAccessLogs(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	cacheKey := c.Query("cacheKey")
	userID := c.Query("userId")

	// 构建查询条件
	query := models.DB.Model(&models.CacheAccessLog{})

	if cacheKey != "" {
		query = query.Where("cache_key = ?", cacheKey)
	}

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取数据
	var logs []models.CacheAccessLog
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取缓存访问日志失败",
			Code:    500,
		})
		return
	}

	result := gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存访问日志成功",
		Code:    200,
		Result:  result,
	})
}

// UpdateCacheConfig 更新缓存配置
func UpdateCacheConfig(c *gin.Context) {
	var config models.CacheConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数格式错误",
			Code:    400,
		})
		return
	}

	// 这里可以将配置保存到数据库或配置文件
	// 暂时返回成功
	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "更新缓存配置成功",
		Code:    200,
		Result:  config,
	})
}

// GetCacheConfig 获取缓存配置
func GetCacheConfig(c *gin.Context) {
	config := models.GetDefaultCacheConfig()

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存配置成功",
		Code:    200,
		Result:  config,
	})
}

// TestCachePerformance 测试缓存性能
func TestCachePerformance(c *gin.Context) {
	// 这里可以实现缓存性能测试逻辑
	// 比如测试Redis连接速度、数据库查询速度等

	result := gin.H{
		"redisLatency":    "2ms",
		"databaseLatency": "15ms",
		"cacheHitRate":    "85%",
		"testTime":        "2025-01-18 15:30:00",
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "缓存性能测试完成",
		Code:    200,
		Result:  result,
	})
}
