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

// CreateBookmark 添加书签/收藏
func CreateBookmark(c *gin.Context) {
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
		PostID interface{} `json:"postId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "参数错误: " + err.Error(),
			Code:    400,
		})
		return
	}

	// 转换 PostID（兼容字符串和数字）
	var postID uint
	switch v := req.PostID.(type) {
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Response{
				Success: false,
				Message: "帖子ID格式错误",
				Code:    400,
			})
			return
		}
		postID = uint(id)
	case float64:
		postID = uint(v)
	default:
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "帖子ID格式错误",
			Code:    400,
		})
		return
	}

	// 检查帖子是否存在
	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
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

	// 检查是否已经收藏
	if models.IsPostBookmarked(userID, postID) {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "已经收藏过该帖子",
			Code:    400,
		})
		return
	}

	// 创建收藏记录
	bookmark := models.Bookmark{
		UserID: userID,
		PostID: postID,
	}

	// 开始事务
	tx := models.DB.Begin()

	// 创建收藏记录
	if err := tx.Create(&bookmark).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "添加收藏失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 更新帖子收藏数
	if err := tx.Model(&models.Post{}).Where("id = ?", postID).
		Update("bookmarks_count", gorm.Expr("bookmarks_count + 1")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新收藏数失败",
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

	// 发送收藏通知
	SendBookmarkNotification(postID, userID)

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "收藏成功",
		Code:    config.CodeSuccess,
		Result:  bookmark,
	})
}

// DeleteBookmark 删除书签/收藏
func DeleteBookmark(c *gin.Context) {
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

	// 获取帖子ID
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的帖子ID",
			Code:    400,
		})
		return
	}

	// 查找收藏记录
	var bookmark models.Bookmark
	if err := models.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&bookmark).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.Response{
				Success: false,
				Message: "收藏记录不存在",
				Code:    config.CodeNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "查询收藏记录失败",
				Code:    config.CodeServerError,
			})
		}
		return
	}

	// 开始事务
	tx := models.DB.Begin()

	// 删除收藏记录
	if err := tx.Delete(&bookmark).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "删除收藏失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 更新帖子收藏数
	if err := tx.Model(&models.Post{}).Where("id = ?", postID).
		Update("bookmarks_count", gorm.Expr("bookmarks_count - 1")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "更新收藏数失败",
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
		Message: "取消收藏成功",
		Code:    config.CodeSuccess,
	})
}

// GetBookmarks 获取用户收藏列表
func GetBookmarks(c *gin.Context) {
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
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取收藏列表
	bookmarks, total, err := models.GetBookmarksByUserID(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取收藏列表失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 转换为安全的帖子信息
	var safePosts []models.SafePost
	for _, bookmark := range bookmarks {
		if bookmark.Post.ID > 0 {
			safePosts = append(safePosts, bookmark.Post.ToSafePost())
		}
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"posts": safePosts,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取收藏列表成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// CheckBookmarkStatus 检查帖子收藏状态
func CheckBookmarkStatus(c *gin.Context) {
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

	// 获取帖子ID
	postIDStr := c.Param("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的帖子ID",
			Code:    400,
		})
		return
	}

	// 检查收藏状态
	isBookmarked := models.IsPostBookmarked(userID, uint(postID))

	result := map[string]interface{}{
		"isBookmarked": isBookmarked,
		"postId":       postID,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取收藏状态成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}
