package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
	"gorm.io/gorm"
)

// GetModerationSettings 获取审核设置
func GetModerationSettings(c *gin.Context) {
	// 检查管理员权限
	userType, exists := c.Get("userType")
	if !exists || userType != "admin" {
		c.JSON(http.StatusForbidden, utils.Response{
			Success: false,
			Message: "权限不足",
			Code:    config.CodeForbidden,
		})
		return
	}

	// 获取所有审核设置
	var settings []models.ModerationSetting
	if err := models.DB.Preload("UpdatedByUser").Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取审核设置失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 转换为map格式便于前端使用
	settingsMap := make(map[string]interface{})
	for _, setting := range settings {
		settingsMap[setting.Key] = map[string]interface{}{
			"value":       setting.Value,
			"description": setting.Description,
			"updatedAt":   setting.UpdatedAt,
			"updatedBy":   setting.UpdatedByUser.ToSafeUser(),
		}
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取审核设置成功",
		Code:    config.CodeSuccess,
		Result:  settingsMap,
	})
}

// UpdateModerationSetting 更新审核设置
func UpdateModerationSetting(c *gin.Context) {
	// 检查管理员权限
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

	// 解析请求参数
	var req struct {
		Key         string `json:"key" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 验证设置键
	validKeys := []string{
		models.ModerationKeyMode,
		models.ModerationKeyAutoRules,
		models.ModerationKeyBannedWords,
		models.ModerationKeyRequireApproval,
	}
	isValidKey := false
	for _, key := range validKeys {
		if req.Key == key {
			isValidKey = true
			break
		}
	}
	if !isValidKey {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的设置键",
			Code:    400,
		})
		return
	}

	// 验证审核模式值
	if req.Key == models.ModerationKeyMode {
		if req.Value != models.ModerationModeAuto && req.Value != models.ModerationModeManual {
			c.JSON(http.StatusBadRequest, utils.Response{
				Success: false,
				Message: "无效的审核模式",
				Code:    400,
			})
			return
		}
	}

	// 更新或创建设置
	if err := models.SetModerationSetting(req.Key, req.Value, req.Description, userID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新审核设置失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "更新审核设置成功",
		Code:    config.CodeSuccess,
	})
}

// GetPendingContent 获取待审核内容
func GetPendingContent(c *gin.Context) {
	// 检查管理员权限
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
	page := 1
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	pageSize := 20
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil {
		pageSize = ps
	}
	contentType := c.DefaultQuery("type", "") // post, comment

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var result map[string]interface{}

	if contentType == "post" || contentType == "" {
		// 获取待审核的帖子（通过举报状态判断）
		var reports []models.Report
		var total int64

		query := models.DB.Model(&models.Report{}).Where("target_type = ? AND status = ?", "post", models.ReportStatusPending)
		query.Count(&total)

		offset := (page - 1) * pageSize
		if err := query.Preload("Reporter").
			Order("created_at DESC").
			Offset(offset).Limit(pageSize).
			Find(&reports).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "获取待审核帖子失败",
				Code:    config.CodeServerError,
			})
			return
		}

		// 获取帖子详情
		var posts []models.Post
		var postIDs []uint
		for _, report := range reports {
			postIDs = append(postIDs, report.TargetID)
		}

		if len(postIDs) > 0 {
			models.DB.Preload("Author").Preload("Club").Where("id IN ?", postIDs).Find(&posts)
		}

		// 组合数据
		var pendingPosts []map[string]interface{}
		for _, report := range reports {
			for _, post := range posts {
				if post.ID == report.TargetID {
					pendingPost := map[string]interface{}{
						"report": report,
						"post":   post.ToSafePost(),
					}
					pendingPosts = append(pendingPosts, pendingPost)
					break
				}
			}
		}

		totalPages := (int(total) + pageSize - 1) / pageSize
		result = map[string]interface{}{
			"type":         "post",
			"pendingPosts": pendingPosts,
			"pagination": map[string]interface{}{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": totalPages,
			},
		}
	} else if contentType == "comment" {
		// 获取待审核的评论
		var reports []models.Report
		var total int64

		query := models.DB.Model(&models.Report{}).Where("target_type = ? AND status = ?", "comment", models.ReportStatusPending)
		query.Count(&total)

		offset := (page - 1) * pageSize
		if err := query.Preload("Reporter").
			Order("created_at DESC").
			Offset(offset).Limit(pageSize).
			Find(&reports).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "获取待审核评论失败",
				Code:    config.CodeServerError,
			})
			return
		}

		// 获取评论详情
		var comments []models.Comment
		var commentIDs []uint
		for _, report := range reports {
			commentIDs = append(commentIDs, report.TargetID)
		}

		if len(commentIDs) > 0 {
			models.DB.Preload("User").Preload("Post").Where("id IN ?", commentIDs).Find(&comments)
		}

		// 组合数据
		var pendingComments []map[string]interface{}
		for _, report := range reports {
			for _, comment := range comments {
				if comment.ID == report.TargetID {
					pendingComment := map[string]interface{}{
						"report":  report,
						"comment": comment.ToSafeComment(),
					}
					pendingComments = append(pendingComments, pendingComment)
					break
				}
			}
		}

		totalPages := (int(total) + pageSize - 1) / pageSize
		result = map[string]interface{}{
			"type":            "comment",
			"pendingComments": pendingComments,
			"pagination": map[string]interface{}{
				"page":       page,
				"pageSize":   pageSize,
				"total":      total,
				"totalPages": totalPages,
			},
		}
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取待审核内容成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// ApproveContent 批准内容
func ApproveContent(c *gin.Context) {
	// 检查管理员权限
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
	reportIDStr := c.Param("reportId")
	reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的举报ID",
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

	// 更新举报状态为已拒绝（内容无问题）
	now := models.DB.NowFunc()
	reviewerID := userID
	if err := models.DB.Model(&report).Updates(map[string]interface{}{
		"status":      models.ReportStatusRejected,
		"reviewer_id": reviewerID,
		"reviewed_at": &now,
		"review_note": "内容审核通过，举报无效",
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新举报状态失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "内容审核通过",
		Code:    config.CodeSuccess,
	})
}

// RejectContent 拒绝内容
func RejectContent(c *gin.Context) {
	// 检查管理员权限
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
	reportIDStr := c.Param("reportId")
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
		Action     string `json:"action" binding:"required"` // delete, hide, warn
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

	// 开始事务
	tx := models.DB.Begin()

	// 根据操作类型处理内容
	switch req.Action {
	case "delete":
		// 删除内容
		switch report.TargetType {
		case "post":
			if err := tx.Delete(&models.Post{}, report.TargetID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "删除帖子失败",
					Code:    config.CodeServerError,
				})
				return
			}
		case "comment":
			if err := tx.Delete(&models.Comment{}, report.TargetID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "删除评论失败",
					Code:    config.CodeServerError,
				})
				return
			}
		}
	case "hide":
		// 隐藏内容（设置状态为0）
		switch report.TargetType {
		case "post":
			if err := tx.Model(&models.Post{}).Where("id = ?", report.TargetID).Update("status", 0).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "隐藏帖子失败",
					Code:    config.CodeServerError,
				})
				return
			}
		case "comment":
			if err := tx.Model(&models.Comment{}).Where("id = ?", report.TargetID).Update("status", 0).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, utils.Response{
					Success: false,
					Message: "隐藏评论失败",
					Code:    config.CodeServerError,
				})
				return
			}
		}
	}

	// 更新举报状态为已批准
	now := models.DB.NowFunc()
	reviewerID := userID
	if err := tx.Model(&report).Updates(map[string]interface{}{
		"status":      models.ReportStatusApproved,
		"reviewer_id": reviewerID,
		"reviewed_at": &now,
		"review_note": req.ReviewNote,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新举报状态失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "提交事务失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "内容处理完成",
		Code:    config.CodeSuccess,
	})
}
