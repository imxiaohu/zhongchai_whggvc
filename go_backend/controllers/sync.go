package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

var syncService *services.SyncService

// InitSyncService 初始化同步服务
func InitSyncService() {
	syncService = services.NewSyncService()
	syncService.Start()
}

// StopSyncService 停止同步服务
func StopSyncService() {
	if syncService != nil {
		syncService.Stop()
	}
}

// GetSyncSettings 获取用户的同步设置
func GetSyncSettings(c *gin.Context) {
	// 从JWT中获取用户ID
	userIdStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStrVal, ok := userIdStr.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userId, err := strconv.ParseUint(userIdStrVal, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取同步设置
	setting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取同步设置失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(setting))
}

// UpdateSyncSettings 更新用户的同步设置
func UpdateSyncSettings(c *gin.Context) {
	// 从JWT中获取用户ID
	userIdStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStrVal, ok := userIdStr.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userId, err := strconv.ParseUint(userIdStrVal, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 解析请求参数
	var req struct {
		Enabled          bool   `json:"enabled"`
		Frequency        string `json:"frequency"`
		TimeRange        string `json:"timeRange"`
		AutoRetryEnabled bool   `json:"autoRetryEnabled"`
		MaxRetryCount    int    `json:"maxRetryCount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数格式错误", 400))
		return
	}

	// 验证频率参数
	validFrequencies := map[string]bool{
		"daily":      true,
		"weekly":     true,
		"every2days": true,
		"every3days": true,
	}

	if !validFrequencies[req.Frequency] {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("无效的同步频率", 400))
		return
	}

	// 获取现有设置
	setting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取同步设置失败", 500))
		return
	}

	// 更新设置
	setting.Enabled = req.Enabled
	setting.Frequency = req.Frequency
	setting.TimeRange = req.TimeRange
	setting.AutoRetryEnabled = req.AutoRetryEnabled
	setting.MaxRetryCount = req.MaxRetryCount

	// 如果启用了同步，计算下次同步时间
	if req.Enabled {
		nextSync := setting.CalculateNextSyncTime()
		setting.NextSyncAt = &nextSync
		setting.SyncStatus = "idle"
	}

	// 保存设置
	if err := models.UpdateSyncSetting(setting); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("更新同步设置失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(setting))
}

// GetSyncStatus 获取同步状态
func GetSyncStatus(c *gin.Context) {
	// 从JWT中获取用户ID
	userIdStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStrVal, ok := userIdStr.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userId, err := strconv.ParseUint(userIdStrVal, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取同步设置
	setting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取同步状态失败", 500))
		return
	}

	// 构建响应数据
	response := map[string]interface{}{
		"enabled":         setting.Enabled,
		"syncStatus":      setting.SyncStatus,
		"lastSyncAt":      setting.LastSyncAt,
		"nextSyncAt":      setting.NextSyncAt,
		"lastSyncMessage": setting.LastSyncMessage,
		"coursesCount":    setting.CoursesCount,
		"retryCount":      setting.RetryCount,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
}

// ManualSync 手动触发同步
func ManualSync(c *gin.Context) {
	// 从JWT中获取用户ID
	userIdStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStrVal, ok := userIdStr.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userId, err := strconv.ParseUint(userIdStrVal, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 检查同步服务是否可用
	if syncService == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewErrorResponse("同步服务不可用", 503))
		return
	}

	// 获取同步设置，检查当前状态
	setting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取同步设置失败", 500))
		return
	}

	// 检查是否正在同步中
	if setting.SyncStatus == "syncing" {
		c.JSON(http.StatusConflict, utils.NewErrorResponse("正在同步中，请稍后再试", 409))
		return
	}

	// 异步执行手动同步
	go func() {
		_ = syncService.ManualSync(uint(userId))
	}()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(map[string]interface{}{
		"message": "手动同步已开始",
	}))
}

// GetSyncLogs 获取同步日志
func GetSyncLogs(c *gin.Context) {
	// 从JWT中获取用户ID
	userIdStr, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStrVal, ok := userIdStr.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userId, err := strconv.ParseUint(userIdStrVal, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取分页参数
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 获取同步日志
	logs, err := models.GetSyncLogsByUserID(uint(userId), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取同步日志失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(logs))
}
