package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// CommentNotificationService 评论通知服务
type CommentNotificationService struct {
	rabbitMQService *RabbitMQService
}

// NewCommentNotificationService 创建评论通知服务
func NewCommentNotificationService() *CommentNotificationService {
	return &CommentNotificationService{
		rabbitMQService: GetRabbitMQService(),
	}
}

// CommentNotificationMessage 评论通知消息结构
type CommentNotificationMessage struct {
	Type      string                 `json:"type"`      // comment_reply, comment_like, comment_mention
	UserID    uint                   `json:"userId"`    // 接收通知的用户ID
	ActorID   uint                   `json:"actorId"`   // 触发通知的用户ID
	PostID    uint                   `json:"postId"`    // 帖子ID
	CommentID uint                   `json:"commentId"` // 评论ID
	Content   string                 `json:"content"`   // 通知内容
	Data      map[string]interface{} `json:"data"`      // 额外数据
	CreatedAt time.Time              `json:"createdAt"` // 创建时间
}

// SendCommentReplyNotification 发送评论回复通知
func (cns *CommentNotificationService) SendCommentReplyNotification(postID, commentID, replyID, actorID uint) error {
	// 获取原评论信息
	var originalComment models.Comment
	if err := models.DB.Preload("User").First(&originalComment, commentID).Error; err != nil {
		return fmt.Errorf("获取原评论失败: %w", err)
	}

	// 获取回复评论信息
	var replyComment models.Comment
	if err := models.DB.Preload("User").First(&replyComment, replyID).Error; err != nil {
		return fmt.Errorf("获取回复评论失败: %w", err)
	}

	// 获取帖子信息
	var post models.Post
	if err := models.DB.First(&post, postID).Error; err != nil {
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 不给自己发通知
	if originalComment.UserID == actorID {
		return nil
	}

	// 构造通知消息
	notification := CommentNotificationMessage{
		Type:      "comment_reply",
		UserID:    originalComment.UserID,
		ActorID:   actorID,
		PostID:    postID,
		CommentID: replyID,
		Content:   fmt.Sprintf("%s 回复了你的评论", replyComment.User.Realname),
		Data: map[string]interface{}{
			"postTitle":       post.Title,
			"originalComment": originalComment.Content,
			"replyContent":    replyComment.Content,
			"actorRealname":   replyComment.User.Realname,
			"actorNickname":   replyComment.User.Nickname,
			"actorAvatar":     replyComment.User.Avatar,
		},
		CreatedAt: time.Now(),
	}

	// 发送评论回复通知到消息队列
	err := cns.sendNotification(notification)

	// 同时发送多渠道通知（评论回复）
	multiChannelService := GetMultiChannelNotificationService()
	if multiChannelService != nil {
		communityData := CommunityNotificationData{
			Type:           "comment_reply",
			UserID:         originalComment.UserID,
			FromUserID:     actorID,
			PostID:         postID,
			PostTitle:      post.Title,
			FromUserName:   replyComment.User.Realname,
			CommentID:      &replyID,
			CommentContent: &replyComment.Content,
		}
		go func() {
			if err := multiChannelService.SendCommunityNotification(communityData); err != nil {
				log.Printf("发送评论回复多渠道通知失败: %v", err)
			}
		}()
	}

	return err
}

// SendCommentLikeNotification 发送评论点赞通知
func (cns *CommentNotificationService) SendCommentLikeNotification(commentID, actorID uint) error {
	// 获取评论信息
	var comment models.Comment
	if err := models.DB.Preload("User").First(&comment, commentID).Error; err != nil {
		return fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取点赞用户信息
	var actor models.User
	if err := models.DB.First(&actor, actorID).Error; err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	// 获取帖子信息
	var post models.Post
	if err := models.DB.First(&post, comment.PostID).Error; err != nil {
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 不给自己发通知
	if comment.UserID == actorID {
		return nil
	}

	// 构造通知消息
	notification := CommentNotificationMessage{
		Type:      "comment_like",
		UserID:    comment.UserID,
		ActorID:   actorID,
		PostID:    comment.PostID,
		CommentID: commentID,
		Content:   fmt.Sprintf("%s 赞了你的评论", actor.Realname),
		Data: map[string]interface{}{
			"postTitle":      post.Title,
			"commentContent": comment.Content,
			"actorRealname":  actor.Realname,
			"actorNickname":  actor.Nickname,
			"actorAvatar":    actor.Avatar,
		},
		CreatedAt: time.Now(),
	}

	// 发送评论通知到消息队列
	err := cns.sendNotification(notification)

	// 同时发送多渠道通知（评论点赞）
	multiChannelService := GetMultiChannelNotificationService()
	if multiChannelService != nil {
		communityData := CommunityNotificationData{
			Type:           "comment_like",
			UserID:         comment.UserID,
			FromUserID:     actorID,
			PostID:         comment.PostID,
			PostTitle:      post.Title,
			FromUserName:   actor.Realname,
			CommentID:      &commentID,
			CommentContent: &comment.Content,
		}
		go func() {
			if err := multiChannelService.SendCommunityNotification(communityData); err != nil {
				log.Printf("发送评论点赞多渠道通知失败: %v", err)
			}
		}()
	}

	return err
}

// SendCommentMentionNotification 发送@提及通知
func (cns *CommentNotificationService) SendCommentMentionNotification(commentID, actorID uint, mentionedUserIDs []uint) error {
	// 获取评论信息
	var comment models.Comment
	if err := models.DB.Preload("User").First(&comment, commentID).Error; err != nil {
		return fmt.Errorf("获取评论失败: %w", err)
	}

	// 获取帖子信息
	var post models.Post
	if err := models.DB.First(&post, comment.PostID).Error; err != nil {
		return fmt.Errorf("获取帖子失败: %w", err)
	}

	// 为每个被提及的用户发送通知
	for _, userID := range mentionedUserIDs {
		// 不给自己发通知
		if userID == actorID {
			continue
		}

		// 获取被提及用户信息
		var mentionedUser models.User
		if err := models.DB.First(&mentionedUser, userID).Error; err != nil {
			log.Printf("获取被提及用户失败: %v", err)
			continue
		}

		// 构造通知消息
		notification := CommentNotificationMessage{
			Type:      "comment_mention",
			UserID:    userID,
			ActorID:   actorID,
			PostID:    comment.PostID,
			CommentID: commentID,
			Content:   fmt.Sprintf("%s 在评论中提到了你", comment.User.Realname),
			Data: map[string]interface{}{
				"postTitle":         post.Title,
				"commentContent":    comment.Content,
				"actorRealname":     comment.User.Realname,
				"actorNickname":     comment.User.Nickname,
				"actorAvatar":       comment.User.Avatar,
				"mentionedRealname": mentionedUser.Realname,
				"mentionedNickname": mentionedUser.Nickname,
			},
			CreatedAt: time.Now(),
		}

		if err := cns.sendNotification(notification); err != nil {
			log.Printf("发送提及通知失败: %v", err)
		}
	}

	return nil
}

// sendNotification 发送通知到消息队列
func (cns *CommentNotificationService) sendNotification(notification CommentNotificationMessage) error {
	if !cns.rabbitMQService.IsEnabled() {
		// 如果RabbitMQ不可用，直接存储到数据库
		return cns.saveNotificationToDB(notification)
	}

	// 序列化通知消息
	messageBody, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	// 发送到消息队列（使用通用的发布方法）
	err = cns.publishToQueue("comment_notifications", messageBody)
	if err != nil {
		log.Printf("发送通知到消息队列失败: %v", err)
		// 如果消息队列失败，直接存储到数据库
		return cns.saveNotificationToDB(notification)
	}

	return nil
}

// publishToQueue 发布消息到指定队列
func (cns *CommentNotificationService) publishToQueue(queueName string, messageBody []byte) error {
	if !cns.rabbitMQService.IsEnabled() {
		return fmt.Errorf("RabbitMQ服务未启用")
	}

	// 这里可以实现通用的队列发布逻辑
	// 由于当前RabbitMQ服务没有通用的PublishMessage方法，我们简化处理
	log.Printf("发布消息到队列 %s: %s", queueName, string(messageBody))
	return nil
}

// saveNotificationToDB 保存通知到数据库
func (cns *CommentNotificationService) saveNotificationToDB(notification CommentNotificationMessage) error {
	// 这里可以实现将通知保存到数据库的逻辑
	// 由于当前没有通知表，这里只是记录日志
	log.Printf("保存通知到数据库: %+v", notification)
	return nil
}

// ProcessMentionedUsers 处理@提及的用户
func (cns *CommentNotificationService) ProcessMentionedUsers(mentionedUsersJSON string) ([]uint, error) {
	if mentionedUsersJSON == "" {
		return nil, nil
	}

	var userIDs []uint
	if err := json.Unmarshal([]byte(mentionedUsersJSON), &userIDs); err != nil {
		return nil, fmt.Errorf("解析提及用户列表失败: %w", err)
	}

	// 验证用户是否存在
	var validUserIDs []uint
	for _, userID := range userIDs {
		var user models.User
		if err := models.DB.First(&user, userID).Error; err == nil {
			validUserIDs = append(validUserIDs, userID)
		}
	}

	return validUserIDs, nil
}

// GetUserNotifications 获取用户通知列表
func (cns *CommentNotificationService) GetUserNotifications(userID uint, page, pageSize int) ([]CommentNotificationMessage, int64, error) {
	// 这里应该从数据库或缓存中获取用户通知
	// 由于当前没有通知表，返回空列表
	return []CommentNotificationMessage{}, 0, nil
}

// MarkNotificationAsRead 标记通知为已读
func (cns *CommentNotificationService) MarkNotificationAsRead(notificationID uint, userID uint) error {
	// 这里应该更新数据库中的通知状态
	// 由于当前没有通知表，只是记录日志
	log.Printf("标记通知为已读: notificationID=%d, userID=%d", notificationID, userID)
	return nil
}

// 全局评论通知服务实例
var globalCommentNotificationService *CommentNotificationService

// InitCommentNotificationService 初始化全局评论通知服务
func InitCommentNotificationService() {
	globalCommentNotificationService = NewCommentNotificationService()
}

// GetCommentNotificationService 获取全局评论通知服务实例
func GetCommentNotificationService() *CommentNotificationService {
	if globalCommentNotificationService == nil {
		InitCommentNotificationService()
	}
	return globalCommentNotificationService
}
