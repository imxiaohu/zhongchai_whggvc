package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// UserSettingUpdateRequest 用户设置更新请求结构
type UserSettingUpdateRequest struct {
	ClientType            *string `json:"clientType,omitempty"`            // 客户端类型
	Language              *string `json:"language,omitempty"`              // 语言设置
	Theme                 *string `json:"theme,omitempty"`                 // 主题设置
	Nickname              *string `json:"nickname,omitempty"`              // 用户昵称
	ErrorNotification     *bool   `json:"errorNotification,omitempty"`     // 错误提示管理
	ErrorNotificationMode *string `json:"errorNotificationMode,omitempty"` // 错误提示模式

	// 服务器同步配置
	SyncEnabled      *bool   `json:"syncEnabled,omitempty"`      // 是否启用服务器同步
	SyncFrequency    *string `json:"syncFrequency,omitempty"`    // 同步频率
	SyncTimeRange    *string `json:"syncTimeRange,omitempty"`    // 同步时间范围
	SyncAutoRetry    *bool   `json:"syncAutoRetry,omitempty"`    // 是否启用自动重试
	SyncNotification *bool   `json:"syncNotification,omitempty"` // 同步完成通知

	// 界面设置
	ShowWelcomeGuide   *bool `json:"showWelcomeGuide,omitempty"`   // 是否显示欢迎引导
	CompactMode        *bool `json:"compactMode,omitempty"`        // 紧凑模式
	ShowAvatarInHeader *bool `json:"showAvatarInHeader,omitempty"` // 头部显示头像

	// 隐私设置
	DataCollection   *bool `json:"dataCollection,omitempty"`   // 数据收集同意
	AnalyticsEnabled *bool `json:"analyticsEnabled,omitempty"` // 分析统计启用

	// 通知设置
	PushNotification     *bool `json:"pushNotification,omitempty"`     // 推送通知
	EmailNotification    *bool `json:"emailNotification,omitempty"`    // 邮件通知
	NewsNotification     *bool `json:"newsNotification,omitempty"`     // 新闻通知
	ScoreNotification    *bool `json:"scoreNotification,omitempty"`    // 成绩通知
	ScheduleNotification *bool `json:"scheduleNotification,omitempty"` // 课程表通知

	// 扩展字段
	CustomSettings *string `json:"customSettings,omitempty"` // 自定义设置JSON字符串
}

// GetUserSettings 获取用户设置
func GetUserSettings(c *gin.Context) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取用户设置
	setting, err := models.GetUserSettingByUserID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户设置失败", 500))
		return
	}

	// 返回用户设置
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(setting))
}

// UpdateUserSettings 更新用户设置
func UpdateUserSettings(c *gin.Context) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}

	// 解析请求数据
	var req UserSettingUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("请求数据格式错误", 400))
		return
	}

	// 获取现有用户设置
	setting, err := models.GetUserSettingByUserID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户设置失败", 500))
		return
	}

	// 更新设置字段（只更新非空字段）
	if req.ClientType != nil {
		setting.ClientType = *req.ClientType
	}
	if req.Language != nil {
		setting.Language = *req.Language
	}
	if req.Theme != nil {
		setting.Theme = *req.Theme
	}
	if req.Nickname != nil {
		setting.Nickname = *req.Nickname
	}
	if req.ErrorNotification != nil {
		setting.ErrorNotification = *req.ErrorNotification
	}
	if req.ErrorNotificationMode != nil {
		setting.ErrorNotificationMode = *req.ErrorNotificationMode
	}

	// 服务器同步配置
	if req.SyncEnabled != nil {
		setting.SyncEnabled = *req.SyncEnabled
	}
	if req.SyncFrequency != nil {
		setting.SyncFrequency = *req.SyncFrequency
	}
	if req.SyncTimeRange != nil {
		setting.SyncTimeRange = *req.SyncTimeRange
	}
	if req.SyncAutoRetry != nil {
		setting.SyncAutoRetry = *req.SyncAutoRetry
	}
	if req.SyncNotification != nil {
		setting.SyncNotification = *req.SyncNotification
	}

	// 界面设置
	if req.ShowWelcomeGuide != nil {
		setting.ShowWelcomeGuide = *req.ShowWelcomeGuide
	}
	if req.CompactMode != nil {
		setting.CompactMode = *req.CompactMode
	}
	if req.ShowAvatarInHeader != nil {
		setting.ShowAvatarInHeader = *req.ShowAvatarInHeader
	}

	// 隐私设置
	if req.DataCollection != nil {
		setting.DataCollection = *req.DataCollection
	}
	if req.AnalyticsEnabled != nil {
		setting.AnalyticsEnabled = *req.AnalyticsEnabled
	}

	// 通知设置
	if req.PushNotification != nil {
		setting.PushNotification = *req.PushNotification
	}
	if req.EmailNotification != nil {
		setting.EmailNotification = *req.EmailNotification
	}
	if req.NewsNotification != nil {
		setting.NewsNotification = *req.NewsNotification
	}
	if req.ScoreNotification != nil {
		setting.ScoreNotification = *req.ScoreNotification
	}
	if req.ScheduleNotification != nil {
		setting.ScheduleNotification = *req.ScheduleNotification
	}

	// 扩展字段
	if req.CustomSettings != nil {
		setting.CustomSettings = *req.CustomSettings
	}

	// 保存更新
	if err := models.UpdateUserSetting(setting); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("更新用户设置失败", 500))
		return
	}

	// 返回更新后的设置
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(setting))
}

// ResetUserSettings 重置用户设置为默认值
func ResetUserSettings(c *gin.Context) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}

	// 删除现有设置
	if err := models.DeleteUserSetting(uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("删除用户设置失败", 500))
		return
	}

	// 重新获取默认设置
	setting, err := models.GetUserSettingByUserID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建默认设置失败", 500))
		return
	}

	// 返回默认设置
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(setting))
}

// GetAllUserSettings 获取所有用户设置（管理员功能）
func GetAllUserSettings(c *gin.Context) {
	// 从查询参数获取分页信息
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取所有用户设置
	settings, total, err := models.GetAllUserSettings(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户设置列表失败", 500))
		return
	}

	// 返回分页数据
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"list":     settings,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}
