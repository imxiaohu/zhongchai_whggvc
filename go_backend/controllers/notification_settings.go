package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// NotificationSettingsController 通知设置控制器
type NotificationSettingsController struct{}

// GetNotificationSettings 获取用户通知设置
func (nsc *NotificationSettingsController) GetNotificationSettings(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取通知配置失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewStandardResponse("获取通知配置成功", channel))
}

// UpdateNotificationSettings 更新用户通知设置
func (nsc *NotificationSettingsController) UpdateNotificationSettings(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	var updateData models.NotificationChannel
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 获取现有配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取通知配置失败: "+err.Error(), 500))
		return
	}

	// 更新配置
	channel.EmailEnabled = updateData.EmailEnabled
	channel.EmailAddress = updateData.EmailAddress
	channel.DingTalkEnabled = updateData.DingTalkEnabled
	channel.DingTalkWebhookURL = updateData.DingTalkWebhookURL
	channel.DingTalkSecret = updateData.DingTalkSecret
	channel.SMSEnabled = updateData.SMSEnabled
	channel.PhoneNumber = updateData.PhoneNumber

	// 更新通知类型偏好
	channel.ScoreUpdateEmail = updateData.ScoreUpdateEmail
	channel.ScoreUpdateDingTalk = updateData.ScoreUpdateDingTalk
	channel.ScoreUpdateSMS = updateData.ScoreUpdateSMS
	channel.CommunityLikeEmail = updateData.CommunityLikeEmail
	channel.CommunityLikeDingTalk = updateData.CommunityLikeDingTalk
	channel.CommunityBookmarkEmail = updateData.CommunityBookmarkEmail
	channel.CommunityBookmarkDingTalk = updateData.CommunityBookmarkDingTalk
	channel.CommunityCommentEmail = updateData.CommunityCommentEmail
	channel.CommunityCommentDingTalk = updateData.CommunityCommentDingTalk
	channel.CommunityCommentLikeEmail = updateData.CommunityCommentLikeEmail
	channel.CommunityCommentLikeDingTalk = updateData.CommunityCommentLikeDingTalk

	// 更新成绩检查配置
	channel.ScoreCheckEnabled = updateData.ScoreCheckEnabled
	channel.ScoreCheckFrequency = updateData.ScoreCheckFrequency
	channel.ScoreCheckTime = updateData.ScoreCheckTime
	channel.ScoreCheckSemester = updateData.ScoreCheckSemester

	// 保存到数据库
	if err := models.DB.Save(channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("保存通知配置失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewStandardResponse("更新通知配置成功", channel))
}

// GetAvailableSemesters 获取可用的学期列表
func (nsc *NotificationSettingsController) GetAvailableSemesters(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取用户信息
	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败: "+err.Error(), 500))
		return
	}

	// 检查用户是否有学校账号绑定
	if user.Username == "" || user.Password == "" {
		// 返回默认学期列表
		defaultSemesters := []services.SemesterInfo{
			{Name: "current", Code: "current", IsCurrent: true},
			{Name: "all", Code: "all", IsCurrent: false},
			{Name: "2024-2025学年第二学期", Code: "2024-2025-2", IsCurrent: true},
			{Name: "2024-2025学年第一学期", Code: "2024-2025-1", IsCurrent: false},
			{Name: "2023-2024学年第二学期", Code: "2023-2024-2", IsCurrent: false},
		}
		c.JSON(http.StatusOK, utils.NewStandardResponse("获取学期列表成功", defaultSemesters))
		return
	}

	// 使用带缓存的学期列表获取服务
	semesters, err := nsc.getCachedAvailableSemesters(user)
	if err != nil {
		// 如果获取失败，返回默认学期列表
		defaultSemesters := []services.SemesterInfo{
			{Name: "current", Code: "current", IsCurrent: true},
			{Name: "all", Code: "all", IsCurrent: false},
			{Name: "2024-2025学年第二学期", Code: "2024-2025-2", IsCurrent: true},
			{Name: "2024-2025学年第一学期", Code: "2024-2025-1", IsCurrent: false},
			{Name: "2023-2024学年第二学期", Code: "2023-2024-2", IsCurrent: false},
		}
		c.JSON(http.StatusOK, utils.NewStandardResponse("获取学期列表成功（使用默认数据）", defaultSemesters))
		return
	}

	// 在学期列表前添加特殊选项
	specialOptions := []services.SemesterInfo{
		{Name: "current", Code: "current", IsCurrent: true},
		{Name: "all", Code: "all", IsCurrent: false},
	}

	allSemesters := specialOptions
	allSemesters = append(allSemesters, semesters...)
	c.JSON(http.StatusOK, utils.NewStandardResponse("获取学期列表成功", allSemesters))
}

// TestScoreCheck 测试成绩检查功能
func (nsc *NotificationSettingsController) TestScoreCheck(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取用户通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取通知配置失败: "+err.Error(), 500))
		return
	}

	var testResults []map[string]interface{}

	// 测试邮件通知
	if channel.EmailEnabled && channel.ScoreUpdateEmail && channel.EmailAddress != "" {
		emailService := services.GetEmailService()
		if emailService != nil && emailService.IsEnabled() {
			// 发送测试邮件
			testScoreUpdates := []services.ScoreUpdate{
				{
					CourseName: "[测试] 高等数学",
					CourseCode: "MATH001",
					Semester:   "2024-2025学年第二学期",
					ScoreType:  "考试分",
					OldScore:   "85",
					NewScore:   "90",
					UpdateTime: time.Now().Format("2006-01-02 15:04:05"),

					// 详细成绩信息
					DailyScore:         "88",
					ExamScore:          "90",
					FinalScore:         "89",
					PracticalScore:     "0",
					SupplementaryScore: "0",

					// 课程信息
					Credit:         4.0,
					GPA:            4.5,
					CourseProperty: "必修课",
					TeacherNames:   "#张教授(MATH001)#",
					TestNote:       "正常",
					ChangeType:     "update",
				},
			}

			if err := emailService.SendScoreUpdateEmail(userID, testScoreUpdates); err != nil {
				testResults = append(testResults, map[string]interface{}{
					"channel": "email",
					"status":  "error",
					"message": "邮件发送失败: " + err.Error(),
				})
			} else {
				testResults = append(testResults, map[string]interface{}{
					"channel": "email",
					"status":  "success",
					"message": "测试邮件发送成功",
				})
			}
		} else {
			testResults = append(testResults, map[string]interface{}{
				"channel": "email",
				"status":  "error",
				"message": "邮件服务不可用",
			})
		}
	}

	// 测试钉钉通知
	if channel.DingTalkEnabled && channel.ScoreUpdateDingTalk && channel.DingTalkWebhookURL != "" {
		dingTalkService := services.GetDingTalkService()
		if dingTalkService != nil && dingTalkService.IsEnabled() {
			// 发送测试钉钉消息
			testScoreUpdates := []services.ScoreUpdate{
				{
					CourseName: "[测试] 高等数学",
					CourseCode: "MATH001",
					Semester:   "2024-2025学年第二学期",
					ScoreType:  "考试分",
					OldScore:   "85",
					NewScore:   "90",
					UpdateTime: time.Now().Format("2006-01-02 15:04:05"),

					// 详细成绩信息
					DailyScore:         "88",
					ExamScore:          "90",
					FinalScore:         "89",
					PracticalScore:     "0",
					SupplementaryScore: "0",

					// 课程信息
					Credit:         4.0,
					GPA:            4.5,
					CourseProperty: "必修课",
					TeacherNames:   "#张教授(MATH001)#",
					TestNote:       "正常",
					ChangeType:     "update",
				},
			}

			if err := dingTalkService.SendScoreUpdateMessage(userID, testScoreUpdates); err != nil {
				testResults = append(testResults, map[string]interface{}{
					"channel": "dingtalk",
					"status":  "error",
					"message": "钉钉消息发送失败: " + err.Error(),
				})
			} else {
				testResults = append(testResults, map[string]interface{}{
					"channel": "dingtalk",
					"status":  "success",
					"message": "测试钉钉消息发送成功",
				})
			}
		} else {
			testResults = append(testResults, map[string]interface{}{
				"channel": "dingtalk",
				"status":  "error",
				"message": "钉钉服务不可用",
			})
		}
	}

	// 测试短信通知
	if channel.SMSEnabled && channel.ScoreUpdateSMS && channel.PhoneNumber != "" {
		smsService := services.GetSMSService()
		if smsService != nil && smsService.IsEnabled() {
			// 发送测试短信
			if err := smsService.SendScoreTestNotificationSMS(userID); err != nil {
				testResults = append(testResults, map[string]interface{}{
					"channel": "sms",
					"status":  "error",
					"message": "短信发送失败: " + err.Error(),
				})
			} else {
				testResults = append(testResults, map[string]interface{}{
					"channel": "sms",
					"status":  "success",
					"message": "测试短信发送成功",
				})
			}
		} else {
			testResults = append(testResults, map[string]interface{}{
				"channel": "sms",
				"status":  "error",
				"message": "短信服务不可用",
			})
		}
	}

	if len(testResults) == 0 {
		c.JSON(http.StatusOK, utils.NewStandardResponse("测试完成", map[string]interface{}{
			"message": "未配置任何通知渠道",
			"results": testResults,
		}))
		return
	}

	// 统计测试结果
	successCount := 0
	for _, result := range testResults {
		if result["status"] == "success" {
			successCount++
		}
	}

	message := fmt.Sprintf("测试完成，成功 %d/%d 个通知渠道", successCount, len(testResults))

	c.JSON(http.StatusOK, utils.NewStandardResponse("测试完成", map[string]interface{}{
		"message":      message,
		"results":      testResults,
		"successCount": successCount,
		"totalCount":   len(testResults),
	}))
}

// GetNotificationLogs 获取通知日志
func (nsc *NotificationSettingsController) GetNotificationLogs(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	notificationType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询条件
	query := models.DB.Where("user_id = ?", userID)
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// 获取总数
	var total int64
	query.Model(&models.NotificationLog{}).Count(&total)

	// 获取日志列表
	var logs []models.NotificationLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取通知日志失败: "+err.Error(), 500))
		return
	}

	response := map[string]interface{}{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"pages":    (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewStandardResponse("获取通知日志成功", response))
}

// GetScoreCheckLogs 获取成绩检查日志
func (nsc *NotificationSettingsController) GetScoreCheckLogs(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询成绩检查相关的日志
	var logs []models.NotificationLog
	var total int64

	// 查询条件：成绩更新通知、批量缓存日志
	query := models.DB.Model(&models.NotificationLog{}).Where("user_id = ? AND (type = ? OR type = ? OR type = ?)",
		userID, "score_update", "batch_cache", "score_check")

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取成绩检查日志失败: "+err.Error(), 500))
		return
	}

	// 获取最后检查时间和配置信息
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户配置失败: "+err.Error(), 500))
		return
	}

	response := map[string]interface{}{
		"logs":           logs,
		"total":          total,
		"page":           page,
		"pageSize":       pageSize,
		"pages":          (total + int64(pageSize) - 1) / int64(pageSize),
		"lastCheckTime":  channel.LastScoreCheck,
		"checkFrequency": channel.ScoreCheckFrequency,
		"checkSemester":  channel.ScoreCheckSemester,
		"checkEnabled":   channel.ScoreCheckEnabled,
	}

	c.JSON(http.StatusOK, utils.NewStandardResponse("获取成绩检查日志成功", response))
}

// GetCurrentSemester 获取当前学期信息
func (nsc *NotificationSettingsController) GetCurrentSemester(c *gin.Context) {
	// 使用工具函数获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未登录", 401))
		return
	}

	// 获取用户信息
	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败: "+err.Error(), 500))
		return
	}

	// 优先从用户表中获取当前学期
	if user.CurrentSemester != "" {
		currentSemester := &services.SemesterInfo{
			Name:      user.CurrentSemester,
			Code:      user.CurrentSemester,
			IsCurrent: true,
		}
		c.JSON(http.StatusOK, utils.NewStandardResponse("获取当前学期成功", currentSemester))
		return
	}

	// 如果用户表中没有设置当前学期，尝试从学校API获取
	schoolAPIService := services.NewSchoolAPIService()
	currentSemester, err := schoolAPIService.GetCurrentSemester(user)
	if err != nil {
		// 返回默认学期
		defaultSemester := &services.SemesterInfo{
			Name:      "2024-2025学年第二学期",
			Code:      "2024-2025-2",
			IsCurrent: true,
		}
		c.JSON(http.StatusOK, utils.NewStandardResponse("获取当前学期成功（使用默认数据）", defaultSemester))
		return
	}

	// 将获取到的学期信息更新到用户表中
	//nolint:errcheck
	models.UpdateUserFields(user.ID, map[string]interface{}{
		"current_semester": currentSemester.Name,
	})

	c.JSON(http.StatusOK, utils.NewStandardResponse("获取当前学期成功", currentSemester))
}

// getCachedAvailableSemesters 获取可用学期列表（已移除缓存功能）
func (nsc *NotificationSettingsController) getCachedAvailableSemesters(user *models.User) ([]services.SemesterInfo, error) {
	schoolAPIService := services.NewSchoolAPIService()
	return schoolAPIService.GetAvailableSemesters(user)
}
