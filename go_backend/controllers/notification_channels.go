package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetNotificationChannels 获取用户通知渠道配置
func GetNotificationChannels(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取通知配置失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取通知配置成功",
		Code:    config.CodeSuccess,
		Result:  channel,
	})
}

// UpdateNotificationChannels 更新用户通知渠道配置
func UpdateNotificationChannels(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 解析请求参数
	var req models.NotificationChannel
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 获取现有配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取通知配置失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 更新配置
	channel.EmailEnabled = req.EmailEnabled
	channel.EmailAddress = req.EmailAddress
	channel.DingTalkEnabled = req.DingTalkEnabled
	channel.DingTalkWebhookURL = req.DingTalkWebhookURL
	channel.DingTalkSecret = req.DingTalkSecret
	channel.SMSEnabled = req.SMSEnabled
	channel.PhoneNumber = req.PhoneNumber
	channel.ScoreUpdateEmail = req.ScoreUpdateEmail
	channel.ScoreUpdateDingTalk = req.ScoreUpdateDingTalk
	channel.ScoreUpdateSMS = req.ScoreUpdateSMS
	// 更新社区互动通知偏好
	channel.CommunityLikeEmail = req.CommunityLikeEmail
	channel.CommunityLikeDingTalk = req.CommunityLikeDingTalk
	channel.CommunityBookmarkEmail = req.CommunityBookmarkEmail
	channel.CommunityBookmarkDingTalk = req.CommunityBookmarkDingTalk
	channel.CommunityCommentEmail = req.CommunityCommentEmail
	channel.CommunityCommentDingTalk = req.CommunityCommentDingTalk
	channel.CommunityCommentLikeEmail = req.CommunityCommentLikeEmail
	channel.CommunityCommentLikeDingTalk = req.CommunityCommentLikeDingTalk
	channel.ScoreCheckEnabled = req.ScoreCheckEnabled
	channel.ScoreCheckFrequency = req.ScoreCheckFrequency
	channel.ScoreCheckTime = req.ScoreCheckTime

	// 验证钉钉Webhook地址
	if channel.DingTalkEnabled && channel.DingTalkWebhookURL != "" {
		dingTalkService := services.GetDingTalkService()
		if dingTalkService != nil && dingTalkService.IsEnabled() {
			if err := dingTalkService.ValidateWebhookURL(channel.DingTalkWebhookURL); err != nil {
				c.JSON(http.StatusBadRequest, utils.Response{
					Success: false,
					Message: "钉钉Webhook地址验证失败: " + err.Error(),
					Code:    400,
				})
				return
			}
		}
	}

	// 保存配置
	if err := models.DB.Save(channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "保存通知配置失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "更新通知配置成功",
		Code:    config.CodeSuccess,
		Result:  channel,
	})
}

// TestNotificationChannel 测试通知渠道
func TestNotificationChannel(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取测试类型
	channelType := c.Param("type")
	if channelType == "" {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "请指定测试的通知渠道类型",
			Code:    400,
		})
		return
	}

	var err error
	switch channelType {
	case "email":
		emailService := services.GetEmailService()
		if emailService != nil && emailService.IsEnabled() {
			err = emailService.SendSystemEmail(userID, "测试邮件", "这是一封测试邮件，用于验证邮件通知功能是否正常。")
		} else {
			err = fmt.Errorf("邮件服务未启用")
		}
	case "dingtalk":
		dingTalkService := services.GetDingTalkService()
		if dingTalkService != nil && dingTalkService.IsEnabled() {
			err = dingTalkService.SendSystemMessage(userID, "测试消息", "这是一条测试消息，用于验证钉钉通知功能是否正常。")
		} else {
			err = fmt.Errorf("钉钉服务未启用")
		}
	case "sms":
		smsService := services.GetSMSService()
		if smsService != nil && smsService.IsEnabled() {
			// 使用专门的测试通知模板
			err = smsService.SendScoreTestNotificationSMS(userID)
		} else {
			err = fmt.Errorf("短信服务未启用")
		}
	default:
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "不支持的通知渠道类型",
			Code:    400,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "发送测试通知失败: " + err.Error(),
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "测试通知发送成功",
		Code:    config.CodeSuccess,
	})
}

// GetSMSBalance 获取短信余额
func GetSMSBalance(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取短信余额
	balance, err := models.GetSMSBalanceByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取短信余额失败",
			Code:    config.CodeServerError,
		})
		return
	}

	result := map[string]interface{}{
		"balance":        balance.Balance,
		"totalSpent":     balance.TotalSpent,
		"balanceYuan":    float64(balance.Balance) / 100,
		"totalSpentYuan": float64(balance.TotalSpent) / 100,
		"smsCost":        services.GetSMSService().GetSMSCost(),
		"smsCostYuan":    float64(services.GetSMSService().GetSMSCost()) / 100,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取短信余额成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// GetSMSTransactions 获取短信交易记录
func GetSMSTransactions(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取分页参数
	page := 1
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	pageSize := 20
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil {
		pageSize = ps
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询交易记录
	var transactions []models.SMSTransaction
	var total int64

	query := models.DB.Model(&models.SMSTransaction{}).Where("user_id = ?", userID)

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取交易记录失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"transactions": transactions,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取交易记录成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// GetRechargePackages 获取充值套餐
func GetRechargePackages(c *gin.Context) {
	wechatPayService := services.GetWechatPayService()
	packages := wechatPayService.GetRechargePackages()

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取充值套餐成功",
		Code:    config.CodeSuccess,
		Result:  packages,
	})
}

// CreateRechargeOrder 创建充值订单
func CreateRechargeOrder(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 解析请求参数
	var req struct {
		PackageID int `json:"packageId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 获取充值套餐信息
	wechatPayService := services.GetWechatPayService()
	packages := wechatPayService.GetRechargePackages()

	var selectedPackage map[string]interface{}
	for _, pkg := range packages {
		if pkg["id"].(int) == req.PackageID {
			selectedPackage = pkg
			break
		}
	}

	if selectedPackage == nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的充值套餐ID",
			Code:    400,
		})
		return
	}

	// 生成订单ID
	orderID := wechatPayService.GenerateOrderID(userID)

	// 创建支付订单
	paymentReq := services.PaymentRequest{
		OrderID:     orderID,
		Amount:      int64(selectedPackage["amount"].(int)),
		Description: selectedPackage["description"].(string),
		UserID:      userID,
	}

	paymentResp, err := wechatPayService.CreateNativePayment(paymentReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "创建支付订单失败: " + err.Error(),
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "创建充值订单成功",
		Code:    config.CodeSuccess,
		Result:  paymentResp,
	})
}

// GetNotificationLogs 获取通知发送日志
func GetNotificationLogs(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取分页参数
	page := 1
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	pageSize := 20
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil {
		pageSize = ps
	}
	channel := c.DefaultQuery("channel", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询通知日志
	var logs []models.NotificationLog
	var total int64

	query := models.DB.Model(&models.NotificationLog{}).Where("user_id = ?", userID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取通知日志失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"logs": logs,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取通知日志成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// TriggerScoreCheck 手动触发成绩检查
func TriggerScoreCheck(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取成绩检查服务
	scoreCheckService := services.GetScoreCheckService()
	if scoreCheckService == nil || !scoreCheckService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, utils.Response{
			Success: false,
			Message: "成绩检查服务未启用",
			Code:    503,
		})
		return
	}

	// 执行成绩检查
	scoreUpdates, err := scoreCheckService.CheckUserScores(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "成绩检查失败: " + err.Error(),
			Code:    config.CodeServerError,
		})
		return
	}

	// 更新最后检查时间
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err == nil {
		now := time.Now()
		channel.LastScoreCheck = &now
		models.DB.Save(channel)
	}

	result := map[string]interface{}{
		"updateCount":  len(scoreUpdates),
		"scoreUpdates": scoreUpdates,
		"checkTime":    time.Now(),
		"hasUpdates":   len(scoreUpdates) > 0,
	}

	message := fmt.Sprintf("成绩检查完成，发现%d条更新", len(scoreUpdates))
	if len(scoreUpdates) == 0 {
		message = "成绩检查完成，暂无更新"
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: message,
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// DebugScoreCheck 调试成绩检查（用于排查问题）
func DebugScoreCheck(c *gin.Context) {
	// 获取用户ID
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	// 获取用户信息
	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.Response{
			Success: false,
			Message: "用户不存在",
			Code:    config.CodeNotFound,
		})
		return
	}

	// 获取成绩检查服务
	scoreCheckService := services.GetScoreCheckService()
	if scoreCheckService == nil || !scoreCheckService.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, utils.Response{
			Success: false,
			Message: "成绩检查服务未启用",
			Code:    503,
		})
		return
	}

	debugInfo := map[string]interface{}{
		"userInfo": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"realname": user.Realname,
		},
		"steps": []map[string]interface{}{},
	}

	// 步骤1: 检查用户是否有成绩快照
	hasSnapshots, err := scoreCheckService.UserHasScoreSnapshots(userID)
	step1 := map[string]interface{}{
		"step":         1,
		"description":  "检查用户成绩快照",
		"hasSnapshots": hasSnapshots,
		"error":        nil,
	}
	if err != nil {
		step1["error"] = err.Error()
	}
	debugInfo["steps"] = append(debugInfo["steps"].([]map[string]interface{}), step1)

	// 步骤2: 获取学期列表
	semesters, err := scoreCheckService.GetSemestersToCheck(user, "current")
	step2 := map[string]interface{}{
		"step":        2,
		"description": "获取要检查的学期",
		"semesters":   semesters,
		"error":       nil,
	}
	if err != nil {
		step2["error"] = err.Error()
	}
	debugInfo["steps"] = append(debugInfo["steps"].([]map[string]interface{}), step2)

	// 步骤3: 尝试获取成绩数据
	if len(semesters) > 0 {
		semester := semesters[0]
		currentScores, err := scoreCheckService.FetchCurrentScoresBySemesterWithCache(user, semester)
		step3 := map[string]interface{}{
			"step":        3,
			"description": fmt.Sprintf("获取学期%s的成绩数据", semester),
			"semester":    semester,
			"scoreCount":  len(currentScores),
			"error":       nil,
		}
		if err != nil {
			step3["error"] = err.Error()
		} else {
			// 只显示前3条成绩作为示例
			sampleScores := []map[string]interface{}{}
			for i, score := range currentScores {
				if i >= 3 {
					break
				}
				sampleScores = append(sampleScores, map[string]interface{}{
					"courseName": score.CourseName,
					"scoreType":  score.ScoreType,
					"score":      score.Score,
					"version":    score.Version,
				})
			}
			step3["sampleScores"] = sampleScores
		}
		debugInfo["steps"] = append(debugInfo["steps"].([]map[string]interface{}), step3)
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "成绩检查调试信息获取成功",
		Code:    config.CodeSuccess,
		Result:  debugInfo,
	})
}

// ClearErrorCache 清理错误缓存数据（已移除缓存功能，此接口保留用于兼容性）
func ClearErrorCache(c *gin.Context) {
	// 缓存功能已移除，此接口保留用于兼容性
	result := map[string]interface{}{
		"totalFound":   0,
		"clearedCount": 0,
		"cacheKeys":    []string{},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "缓存功能已移除",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}
