package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetCacheStats 获取缓存统计信息
func GetCacheStats(c *gin.Context) {
	cacheService := services.GetCacheService()
	stats := cacheService.GetStats()

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取缓存统计信息成功",
		Code:    0,
		Result:  stats,
	})
}

// ClearCache 清空缓存
func ClearCache(c *gin.Context) {
	cacheService := services.GetCacheService()
	cacheService.Clear()

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "缓存已清空（包括内存缓存和Redis缓存）",
		Code:    0,
		Result:  nil,
	})
}

// GetRedisStats 获取Redis统计信息
func GetRedisStats(c *gin.Context) {
	redisService := services.GetRedisService()
	stats := redisService.GetStats()

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取Redis统计信息成功",
		Code:    0,
		Result:  stats,
	})
}
