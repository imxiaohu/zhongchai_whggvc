package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
	"gorm.io/gorm"
)

// GetNotifications 获取用户通知列表
func GetNotifications(c *gin.Context) {
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
	unreadOnly := c.DefaultQuery("unreadOnly", "false") == "true"

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取通知列表
	notifications, total, err := models.GetNotificationsByUserID(userID, page, pageSize, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "获取通知列表失败",
			Code:    config.CodeServerError,
		})
		return
	}

	// 计算分页信息
	totalPages := (int(total) + pageSize - 1) / pageSize

	result := map[string]interface{}{
		"notifications": notifications,
		"pagination": map[string]interface{}{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取通知列表成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// MarkNotificationAsRead 标记通知为已读
func MarkNotificationAsRead(c *gin.Context) {
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

	// 获取通知ID
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Response{
			Success: false,
			Message: "无效的通知ID",
			Code:    400,
		})
		return
	}

	// 查找通知
	var notification models.Notification
	if err := models.DB.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.Response{
				Success: false,
				Message: "通知不存在",
				Code:    config.CodeNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, utils.Response{
				Success: false,
				Message: "查询通知失败",
				Code:    config.CodeServerError,
			})
		}
		return
	}

	// 更新为已读
	if err := models.DB.Model(&notification).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "标记通知已读失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "标记通知已读成功",
		Code:    config.CodeSuccess,
	})
}

// MarkAllNotificationsAsRead 标记所有通知为已读
func MarkAllNotificationsAsRead(c *gin.Context) {
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

	// 更新所有未读通知为已读
	if err := models.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.Response{
			Success: false,
			Message: "标记所有通知已读失败",
			Code:    config.CodeServerError,
		})
		return
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "标记所有通知已读成功",
		Code:    config.CodeSuccess,
	})
}

// GetUnreadNotificationCount 获取未读通知数量
func GetUnreadNotificationCount(c *gin.Context) {
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

	// 获取未读通知数量
	count := models.GetUnreadNotificationCount(userID)

	result := map[string]interface{}{
		"unreadCount": count,
	}

	c.JSON(http.StatusOK, utils.Response{
		Success: true,
		Message: "获取未读通知数量成功",
		Code:    config.CodeSuccess,
		Result:  result,
	})
}

// SendLikeNotification 发送点赞通知
func SendLikeNotification(postID, fromUserID uint) {
	// 获取帖子信息
	var post models.Post
	if err := models.DB.Preload("Author").First(&post, postID).Error; err != nil {
		return
	}

	// 临时允许给自己发通知（用于测试）
	// if post.AuthorID == fromUserID {
	// 	return
	// }

	// 获取点赞用户信息
	var fromUser models.User
	if err := models.DB.First(&fromUser, fromUserID).Error; err != nil {
		return
	}

	// 创建通知记录
	notification := models.Notification{
		UserID:      post.AuthorID,
		Type:        models.NotificationTypeLike,
		Title:       "收到新的点赞",
		Content:     fromUser.Realname + " 点赞了你的帖子《" + post.Title + "》",
		RelatedID:   &postID,
		RelatedType: models.RelatedTypePost,
		FromUserID:  &fromUserID,
		IsRead:      false,
	}

	// 保存到数据库
	if err := models.DB.Create(&notification).Error; err != nil {
		return
	}

	// 发送到消息队列
	rabbitMQ := services.GetRabbitMQService()
	if rabbitMQ != nil && rabbitMQ.IsEnabled() {
		msg := services.NotificationMessage{
			Type:        models.NotificationTypeLike,
			UserID:      post.AuthorID,
			FromUserID:  fromUserID,
			Title:       notification.Title,
			Content:     notification.Content,
			RelatedID:   postID,
			RelatedType: models.RelatedTypePost,
			Data: map[string]interface{}{
				"postTitle": post.Title,
				"fromUser":  fromUser.Nickname,
			},
		}
		//nolint:errcheck
		rabbitMQ.PublishNotification(msg)
	}

	// 发送多渠道通知
	multiChannelService := services.GetMultiChannelNotificationService()
	log.Printf("多渠道通知服务状态: %v", multiChannelService != nil)
	if multiChannelService != nil {
		communityData := services.CommunityNotificationData{
			Type:         "like",
			UserID:       post.AuthorID,
			FromUserID:   fromUserID,
			PostID:       postID,
			PostTitle:    post.Title,
			FromUserName: fromUser.Realname,
		}
		log.Printf("准备发送点赞多渠道通知: 用户%d -> 用户%d", fromUserID, post.AuthorID)
		go func() {
			if err := multiChannelService.SendCommunityNotification(communityData); err != nil {
				log.Printf("发送点赞多渠道通知失败: %v", err)
			} else {
				log.Printf("点赞多渠道通知发送成功")
			}
		}()
	} else {
		log.Printf("多渠道通知服务未初始化")
	}
}

// SendCommentNotification 发送评论通知
func SendCommentNotification(postID, commentID, fromUserID uint) {
	// 获取帖子信息
	var post models.Post
	if err := models.DB.Preload("Author").First(&post, postID).Error; err != nil {
		return
	}

	// 临时允许给自己发通知（用于测试）
	// if post.AuthorID == fromUserID {
	// 	return
	// }

	// 获取评论用户信息
	var fromUser models.User
	if err := models.DB.First(&fromUser, fromUserID).Error; err != nil {
		return
	}

	// 创建通知记录
	notification := models.Notification{
		UserID:      post.AuthorID,
		Type:        models.NotificationTypeComment,
		Title:       "收到新的评论",
		Content:     fromUser.Realname + " 评论了你的帖子《" + post.Title + "》",
		RelatedID:   &commentID,
		RelatedType: models.RelatedTypeComment,
		FromUserID:  &fromUserID,
		IsRead:      false,
	}

	// 保存到数据库
	if err := models.DB.Create(&notification).Error; err != nil {
		return
	}

	// 发送到消息队列
	rabbitMQ := services.GetRabbitMQService()
	if rabbitMQ != nil && rabbitMQ.IsEnabled() {
		msg := services.NotificationMessage{
			Type:        models.NotificationTypeComment,
			UserID:      post.AuthorID,
			FromUserID:  fromUserID,
			Title:       notification.Title,
			Content:     notification.Content,
			RelatedID:   commentID,
			RelatedType: models.RelatedTypeComment,
			Data: map[string]interface{}{
				"postTitle": post.Title,
				"fromUser":  fromUser.Nickname,
			},
		}
		//nolint:errcheck
		rabbitMQ.PublishNotification(msg)
	}

	// 发送多渠道通知
	multiChannelService := services.GetMultiChannelNotificationService()
	if multiChannelService != nil {
		// 获取评论内容
		var comment models.Comment
		var commentContent *string
		if err := models.DB.First(&comment, commentID).Error; err == nil {
			commentContent = &comment.Content
		}

		communityData := services.CommunityNotificationData{
			Type:           "comment",
			UserID:         post.AuthorID,
			FromUserID:     fromUserID,
			PostID:         postID,
			PostTitle:      post.Title,
			FromUserName:   fromUser.Realname,
			CommentID:      &commentID,
			CommentContent: commentContent,
		}
		go func() {
			if err := multiChannelService.SendCommunityNotification(communityData); err != nil {
				log.Printf("发送评论多渠道通知失败: %v", err)
			}
		}()
	}
}

// SendBookmarkNotification 发送收藏通知
func SendBookmarkNotification(postID, fromUserID uint) {
	// 获取帖子信息
	var post models.Post
	if err := models.DB.Preload("Author").First(&post, postID).Error; err != nil {
		return
	}

	// 临时允许给自己发通知（用于测试）
	// if post.AuthorID == fromUserID {
	// 	return
	// }

	// 获取收藏用户信息
	var fromUser models.User
	if err := models.DB.First(&fromUser, fromUserID).Error; err != nil {
		return
	}

	// 创建通知记录
	notification := models.Notification{
		UserID:      post.AuthorID,
		Type:        models.NotificationTypeBookmark,
		Title:       "收到新的收藏",
		Content:     fromUser.Realname + " 收藏了你的帖子《" + post.Title + "》",
		RelatedID:   &postID,
		RelatedType: models.RelatedTypePost,
		FromUserID:  &fromUserID,
		IsRead:      false,
	}

	// 保存到数据库
	if err := models.DB.Create(&notification).Error; err != nil {
		return
	}

	// 发送到消息队列
	rabbitMQ := services.GetRabbitMQService()
	if rabbitMQ != nil && rabbitMQ.IsEnabled() {
		msg := services.NotificationMessage{
			Type:        models.NotificationTypeBookmark,
			UserID:      post.AuthorID,
			FromUserID:  fromUserID,
			Title:       notification.Title,
			Content:     notification.Content,
			RelatedID:   postID,
			RelatedType: models.RelatedTypePost,
			Data: map[string]interface{}{
				"postTitle": post.Title,
				"fromUser":  fromUser.Nickname,
			},
		}
		//nolint:errcheck
		rabbitMQ.PublishNotification(msg)
	}

	// 发送多渠道通知
	multiChannelService := services.GetMultiChannelNotificationService()
	if multiChannelService != nil {
		communityData := services.CommunityNotificationData{
			Type:         "bookmark",
			UserID:       post.AuthorID,
			FromUserID:   fromUserID,
			PostID:       postID,
			PostTitle:    post.Title,
			FromUserName: fromUser.Realname,
		}
		go func() {
			if err := multiChannelService.SendCommunityNotification(communityData); err != nil {
				log.Printf("发送收藏多渠道通知失败: %v", err)
			}
		}()
	}
}
