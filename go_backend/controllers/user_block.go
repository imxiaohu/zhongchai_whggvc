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

// BlockUser 屏蔽用户
func BlockUser(c *gin.Context) {
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

	// 获取要屏蔽的用户ID
	blockedUserIDStr := c.Param("id")
	blockedUserID, err := strconv.ParseUint(blockedUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的用户ID",
			Code:    400,
		})
		return
	}

	// 不能屏蔽自己
	if userID == uint(blockedUserID) {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "不能屏蔽自己",
			Code:    400,
		})
		return
	}

	// 检查要屏蔽的用户是否存在
	var blockedUser models.User
	if err := models.DB.First(&blockedUser, blockedUserID).Error; err != nil {
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

	// 检查是否已经屏蔽
	if models.IsUserBlocked(userID, uint(blockedUserID)) {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "已经屏蔽该用户",
			Code:    400,
		})
		return
	}

	// 解析请求参数
	var req struct {
		Reason string `json:"reason"` // 屏蔽原因
	}
	//nolint:errcheck
	c.ShouldBindJSON(&req)

	// 创建屏蔽记录
	userBlock := models.UserBlock{
		UserID:        userID,
		BlockedUserID: uint(blockedUserID),
		Reason:        req.Reason,
	}

	if err := models.DB.Create(&userBlock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "屏蔽用户失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "屏蔽用户成功",
		Code:    config.CodeSuccess,
		Result:  userBlock,
	})
}

// UnblockUser 取消屏蔽用户
func UnblockUser(c *gin.Context) {
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

	// 获取要取消屏蔽的用户ID
	blockedUserIDStr := c.Param("id")
	blockedUserID, err := strconv.ParseUint(blockedUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的用户ID",
			Code:    400,
		})
		return
	}

	// 查找屏蔽记录
	var userBlock models.UserBlock
	if err := models.DB.Where("user_id = ? AND blocked_user_id = ?", userID, blockedUserID).First(&userBlock).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.Response{
				Success: false,
				Message: "屏蔽记录不存在",
				Code:    config.CodeNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "查询屏蔽记录失败",
				Code:    config.CodeServerError,
			})
		}
		return
	}

	// 删除屏蔽记录
	if err := models.DB.Delete(&userBlock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "取消屏蔽失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "取消屏蔽成功",
		Code:    config.CodeSuccess,
	})
}

// GetBlockedUsers 获取屏蔽用户列表
func GetBlockedUsers(c *gin.Context) {
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询屏蔽列表
	var userBlocks []models.UserBlock
	var total int64

	query := models.DB.Model(&models.UserBlock{}).Where("user_id = ?", userID)

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("BlockedUser").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&userBlocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取屏蔽列表失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 转换为安全的用户信息
	var blockedUsers []map[string]interface{}
	for _, userBlock := range userBlocks {
		blockedUser := map[string]interface{}{
			"id":          userBlock.ID,
			"blockedAt":   userBlock.CreatedAt,
			"reason":      userBlock.Reason,
			"blockedUser": userBlock.BlockedUser.ToSafeUser(),
		}
		blockedUsers = append(blockedUsers, blockedUser)
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"blockedUsers": blockedUsers,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取屏蔽列表成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// CheckBlockStatus 检查用户屏蔽状态
func CheckBlockStatus(c *gin.Context) {
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

	// 获取要检查的用户ID
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的用户ID",
			Code:    400,
		})
		return
	}

	// 检查屏蔽状态
	isBlocked := models.IsUserBlocked(userID, uint(targetUserID))

	result := map[string]interface{}{
		"isBlocked":    isBlocked,
		"targetUserId": targetUserID,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取屏蔽状态成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}
