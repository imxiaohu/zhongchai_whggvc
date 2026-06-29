package services

import (
	"fmt"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// CommentRateLimitService 评论频率限制服务
type CommentRateLimitService struct {
	// Redis服务已移除，限流功能暂时禁用
}

// NewCommentRateLimitService 创建评论频率限制服务
func NewCommentRateLimitService() *CommentRateLimitService {
	return &CommentRateLimitService{}
}

// CheckRateLimit 检查用户评论频率限制（已移除Redis，暂时禁用限流）
func (crls *CommentRateLimitService) CheckRateLimit(userID uint) error {
	// Redis服务已移除，暂时跳过频率限制
	return nil
}

// CheckSpamContent 检查垃圾内容
func (crls *CommentRateLimitService) CheckSpamContent(content string, userID uint) error {
	// 检查内容长度
	if len(content) > 2000 {
		return fmt.Errorf("评论内容过长，最多2000字符")
	}

	// 检查是否包含敏感词（简单实现）
	spamWords := []string{"垃圾", "广告", "刷屏", "spam"}
	for _, word := range spamWords {
		if contains(content, word) {
			return fmt.Errorf("评论内容包含敏感词汇")
		}
	}

	// Redis服务已移除，暂时跳过重复内容检查

	return nil
}

// RecordCommentActivity 记录用户评论活动（已移除Redis，暂时禁用）
func (crls *CommentRateLimitService) RecordCommentActivity(userID uint, postID uint, commentID uint) {
	// Redis服务已移除，暂时跳过活动记录
}

// GetUserCommentStats 获取用户评论统计
func (crls *CommentRateLimitService) GetUserCommentStats(userID uint) map[string]interface{} {
	stats := map[string]interface{}{
		"todayCount":  0,
		"weekCount":   0,
		"monthCount":  0,
		"totalCount":  0,
		"lastComment": nil,
	}

	// 从数据库获取统计数据
	var todayCount, weekCount, monthCount, totalCount int64

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	// 今日评论数
	models.DB.Model(&models.Comment{}).
		Where("user_id = ? AND created_at >= ? AND status = 1", userID, today).
		Count(&todayCount)

	// 本周评论数
	models.DB.Model(&models.Comment{}).
		Where("user_id = ? AND created_at >= ? AND status = 1", userID, weekAgo).
		Count(&weekCount)

	// 本月评论数
	models.DB.Model(&models.Comment{}).
		Where("user_id = ? AND created_at >= ? AND status = 1", userID, monthAgo).
		Count(&monthCount)

	// 总评论数
	models.DB.Model(&models.Comment{}).
		Where("user_id = ? AND status = 1", userID).
		Count(&totalCount)

	// 最近一条评论
	var lastComment models.Comment
	result := models.DB.Where("user_id = ? AND status = 1", userID).
		Order("created_at DESC").First(&lastComment)

	stats["todayCount"] = todayCount
	stats["weekCount"] = weekCount
	stats["monthCount"] = monthCount
	stats["totalCount"] = totalCount

	if result.Error == nil {
		stats["lastComment"] = map[string]interface{}{
			"id":        lastComment.ID,
			"content":   lastComment.Content,
			"createdAt": lastComment.CreatedAt,
		}
	}

	return stats
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr)))
}

// containsInMiddle 检查字符串中间是否包含子字符串
func containsInMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// 全局评论频率限制服务实例
var globalCommentRateLimitService *CommentRateLimitService

// InitCommentRateLimitService 初始化全局评论频率限制服务
func InitCommentRateLimitService() {
	globalCommentRateLimitService = NewCommentRateLimitService()
}

// GetCommentRateLimitService 获取全局评论频率限制服务实例
func GetCommentRateLimitService() *CommentRateLimitService {
	if globalCommentRateLimitService == nil {
		InitCommentRateLimitService()
	}
	return globalCommentRateLimitService
}
