package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
	"gorm.io/gorm"
)

// CreateReport 提交举报
func CreateReport(c *gin.Context) {
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
		TargetType  string `json:"targetType" binding:"required"` // post, comment, user
		TargetID    uint   `json:"targetId" binding:"required"`   // 目标ID
		Reason      string `json:"reason" binding:"required"`     // 举报原因
		Description string `json:"description"`                   // 详细描述
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 验证举报类型
	if req.TargetType != "post" && req.TargetType != "comment" && req.TargetType != "user" {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的举报类型",
			Code:    400,
		})
		return
	}

	// 验证举报原因
	validReasons := []string{
		models.ReportReasonSpam,
		models.ReportReasonInappropriate,
		models.ReportReasonHarassment,
		models.ReportReasonFakeInfo,
		models.ReportReasonViolence,
		models.ReportReasonInfringement,
		models.ReportReasonOther,
	}
	isValidReason := false
	for _, reason := range validReasons {
		if req.Reason == reason {
			isValidReason = true
			break
		}
	}
	if !isValidReason {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的举报原因",
			Code:    400,
		})
		return
	}

	// 检查目标是否存在
	switch req.TargetType {
	case "post":
		var post models.Post
		if err := models.DB.First(&post, req.TargetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, utils.Response{
					Success: false,
					Message: "帖子不存在",
					Code:    config.CodeNotFound,
				})
			} else {
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "查询帖子失败",
					Code:    config.CodeServerError,
				})
			}
			return
		}
	case "comment":
		var comment models.Comment
		if err := models.DB.First(&comment, req.TargetID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, utils.Response{
					Success: false,
					Message: "评论不存在",
					Code:    config.CodeNotFound,
				})
			} else {
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "查询评论失败",
					Code:    config.CodeServerError,
				})
			}
			return
		}
	case "user":
		var user models.User
		if err := models.DB.First(&user, req.TargetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.Response{
				Success: false,
				Message: "用户不存在",
					Code:    config.CodeNotFound,
				})
			} else {
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "查询用户失败",
					Code:    config.CodeServerError,
				})
			}
			return
		}
	}

	// 检查是否已经举报过
	var existingReport models.Report
	if err := models.DB.Where("reporter_id = ? AND target_type = ? AND target_id = ? AND status = ?",
		userID, req.TargetType, req.TargetID, models.ReportStatusPending).First(&existingReport).Error; err == nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "您已经举报过该内容，请等待审核结果",
			Code:    400,
		})
		return
	}

	// 创建举报记录
	report := models.Report{
		ReporterID:  userID,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		Reason:      req.Reason,
		Description: req.Description,
		Status:      models.ReportStatusPending,
	}

	if err := models.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "提交举报失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 发送到内容审核队列
	rabbitMQ := services.GetRabbitMQService()
	if rabbitMQ != nil && rabbitMQ.IsEnabled() {
		// 获取审核模式
		moderationMode := models.ModerationModeManual // 默认手动审核
		if setting, err := models.GetModerationSetting(models.ModerationKeyMode); err == nil {
			moderationMode = setting.Value
		}

		msg := services.ModerationMessage{
			Type:       moderationMode,
			TargetType: req.TargetType,
			TargetID:   req.TargetID,
			UserID:     userID,
			Data: map[string]interface{}{
				"reportId":    report.ID,
				"reason":      req.Reason,
				"description": req.Description,
			},
		}
		//nolint:errcheck
		rabbitMQ.PublishModerationMessage(msg)
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "举报提交成功，我们会尽快处理",
		Code:    config.CodeSuccess,
		Result:  report,
	})
}

// GetReports 获取举报列表（管理员）
func GetReports(c *gin.Context) {
	// 获取用户信息
	userType, exists := c.Get("userType")
	if !exists || userType != "admin" {
		c.JSON(http.StatusForbidden, utils.Response{
			Success: false,
			Message: "权限不足",
			Code:    config.CodeForbidden,
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.DefaultQuery("status", "")
	targetType := c.DefaultQuery("targetType", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := models.DB.Model(&models.Report{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页查询
	var reports []models.Report
	offset := (page - 1) * pageSize
	if err := query.Preload("Reporter").Preload("Reviewer").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取举报列表失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"reports": reports,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取举报列表成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// ReviewReport 审核举报
func ReviewReport(c *gin.Context) {
	// 获取用户信息
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.Response{
			Success: false,
			Message: config.MsgUnauthorized,
			Code:    config.CodeUnauthorized,
		})
		return
	}

	userType, exists := c.Get("userType")
	if !exists || userType != "admin" {
		c.JSON(http.StatusForbidden, utils.Response{
			Success: false,
			Message: "权限不足",
			Code:    config.CodeForbidden,
		})
		return
	}

	// 获取举报ID
	reportIDStr := c.Param("id")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的举报ID",
			Code:    400,
		})
		return
	}

	// 解析请求参数
	var req struct {
		Status     string `json:"status" binding:"required"` // approved, rejected
		ReviewNote string `json:"reviewNote"`                // 审核备注
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 验证状态
	if req.Status != models.ReportStatusApproved && req.Status != models.ReportStatusRejected {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的审核状态",
			Code:    400,
		})
		return
	}

	// 查找举报记录
	var report models.Report
	if err := models.DB.First(&report, reportID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.Response{
				Success: false,
				Message: "举报记录不存在",
				Code:    config.CodeNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "查询举报记录失败",
				Code:    config.CodeServerError,
			})
		}
		return
	}

	// 检查是否已经审核过
	if report.Status != models.ReportStatusPending {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "该举报已经审核过",
			Code:    400,
		})
		return
	}

	// 更新举报状态
	now := models.DB.NowFunc()
	reviewerID := userID
	if err := models.DB.Model(&report).Updates(map[string]interface{}{
		"status":      req.Status,
		"reviewer_id": reviewerID,
		"reviewed_at": &now,
		"review_note": req.ReviewNote,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新举报状态失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 如果举报被批准，可以在这里添加对被举报内容的处理逻辑
	// 例如：删除帖子、禁用用户等

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "审核完成",
		Code:    config.CodeSuccess,
	})
}
