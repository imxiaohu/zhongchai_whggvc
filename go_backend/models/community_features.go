package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Bookmark 书签/收藏模型
type Bookmark struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID    uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PostID    uint           `gorm:"not null;index" json:"postId"` // 帖子ID
	Post      Post           `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

// Notification 通知模型
type Notification struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID      uint           `gorm:"not null;index" json:"userId"` // 接收通知的用户ID
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type        string         `gorm:"size:20;not null" json:"type"`   // 通知类型: like, comment, system
	Title       string         `gorm:"size:100;not null" json:"title"` // 通知标题
	Content     string         `gorm:"size:500" json:"content"`        // 通知内容
	RelatedID   *uint          `gorm:"index" json:"relatedId"`         // 相关对象ID（帖子ID、评论ID等）
	RelatedType string         `gorm:"size:20" json:"relatedType"`     // 相关对象类型: post, comment
	FromUserID  *uint          `gorm:"index" json:"fromUserId"`        // 触发通知的用户ID
	FromUser    *User          `gorm:"foreignKey:FromUserID" json:"fromUser,omitempty"`
	IsRead      bool           `gorm:"default:false" json:"isRead"` // 是否已读
	Data        string         `gorm:"type:text" json:"data"`       // 额外数据，JSON格式
}

// Report 举报模型
type Report struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	ReporterID  uint           `gorm:"not null;index" json:"reporterId"` // 举报人ID
	Reporter    User           `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	TargetType  string         `gorm:"size:20;not null" json:"targetType"`      // 举报目标类型: post, comment, user
	TargetID    uint           `gorm:"not null;index" json:"targetId"`          // 举报目标ID
	Reason      string         `gorm:"size:50;not null" json:"reason"`          // 举报原因分类
	Description string         `gorm:"size:500" json:"description"`             // 详细描述
	Status      string         `gorm:"size:20;default:'pending'" json:"status"` // 状态: pending, approved, rejected
	ReviewerID  *uint          `gorm:"index" json:"reviewerId"`                 // 审核人ID
	Reviewer    *User          `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ReviewedAt  *time.Time     `json:"reviewedAt"`                 // 审核时间
	ReviewNote  string         `gorm:"size:500" json:"reviewNote"` // 审核备注
}

// UserBlock 用户屏蔽模型
type UserBlock struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID        uint           `gorm:"not null;index" json:"userId"` // 屏蔽者ID
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BlockedUserID uint           `gorm:"not null;index" json:"blockedUserId"` // 被屏蔽者ID
	BlockedUser   User           `gorm:"foreignKey:BlockedUserID" json:"blockedUser,omitempty"`
	Reason        string         `gorm:"size:200" json:"reason"` // 屏蔽原因
}

// UserFollow 用户关注模型
type UserFollow struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	FollowerID  uint           `gorm:"not null;uniqueIndex:idx_follower_following" json:"followerId"`  // 关注者ID
	Follower    User           `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	FollowingID uint           `gorm:"not null;uniqueIndex:idx_follower_following" json:"followingId"` // 被关注者ID
	Following   User           `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// FollowUser 关注用户
func FollowUser(followerID, followingID uint) error {
	if followerID == followingID {
		return fmt.Errorf("不能关注自己")
	}
	follow := UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return DB.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		FirstOrCreate(&follow).Error
}

// UnfollowUser 取消关注
func UnfollowUser(followerID, followingID uint) error {
	return DB.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&UserFollow{}).Error
}

// IsFollowing 检查是否关注
func IsFollowing(followerID, followingID uint) bool {
	var count int64
	DB.Model(&UserFollow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count)
	return count > 0
}

// GetFollowersCount 获取粉丝数
func GetFollowersCount(userID uint) int64 {
	var count int64
	DB.Model(&UserFollow{}).Where("following_id = ?", userID).Count(&count)
	return count
}

// GetFollowingCount 获取关注数
func GetFollowingCount(userID uint) int64 {
	var count int64
	DB.Model(&UserFollow{}).Where("follower_id = ?", userID).Count(&count)
	return count
}

// GetFollowers 获取粉丝列表
func GetFollowers(userID uint, page, pageSize int) ([]UserFollow, int64, error) {
	var followers []UserFollow
	var total int64

	if err := DB.Model(&UserFollow{}).Where("following_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Preload("Follower").
		Where("following_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&followers).Error; err != nil {
		return nil, 0, err
	}

	return followers, total, nil
}

// GetFollowing 获取关注列表
func GetFollowing(userID uint, page, pageSize int) ([]UserFollow, int64, error) {
	var following []UserFollow
	var total int64

	if err := DB.Model(&UserFollow{}).Where("follower_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := DB.Preload("Following").
		Where("follower_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&following).Error; err != nil {
		return nil, 0, err
	}

	return following, total, nil
}

// ModerationSetting 内容审核设置模型
type ModerationSetting struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Key             string         `gorm:"size:50;uniqueIndex;not null" json:"key"` // 设置键
	Value           string         `gorm:"size:500;not null" json:"value"`          // 设置值
	Description     string         `gorm:"size:200" json:"description"`             // 设置描述
	UpdatedByUserID uint           `gorm:"not null" json:"updatedByUserId"`         // 更新者ID
	UpdatedByUser   User           `gorm:"foreignKey:UpdatedByUserID" json:"updatedByUser,omitempty"`
}

// 举报原因常量
const (
	ReportReasonSpam          = "spam"          // 垃圾信息
	ReportReasonInappropriate = "inappropriate" // 不当内容
	ReportReasonHarassment    = "harassment"    // 骚扰
	ReportReasonFakeInfo      = "fake_info"     // 虚假信息
	ReportReasonViolence      = "violence"      // 暴力内容
	ReportReasonInfringement  = "infringement"  // 侵权
	ReportReasonOther         = "other"         // 其他
)

// 通知类型常量
const (
	NotificationTypeLike     = "like"     // 点赞通知
	NotificationTypeComment  = "comment"  // 评论通知
	NotificationTypeBookmark = "bookmark" // 收藏通知
	NotificationTypeSystem   = "system"   // 系统通知
)

// 审核设置键常量
const (
	ModerationKeyMode            = "moderation_mode"       // 审核模式: auto, manual
	ModerationKeyAutoRules       = "auto_moderation_rules" // 自动审核规则
	ModerationKeyBannedWords     = "banned_words"          // 禁用词列表
	ModerationKeyRequireApproval = "require_approval"      // 是否需要审核
)

// 审核模式常量
const (
	ModerationModeAuto   = "auto"   // 自动审核
	ModerationModeManual = "manual" // 手动审核
)

// 举报状态常量
const (
	ReportStatusPending  = "pending"  // 待处理
	ReportStatusApproved = "approved" // 已通过（举报有效）
	ReportStatusRejected = "rejected" // 已拒绝（举报无效）
)

// 相关对象类型常量
const (
	RelatedTypePost    = "post"    // 帖子
	RelatedTypeComment = "comment" // 评论
	RelatedTypeUser    = "user"    // 用户
)

// GetBookmarksByUserID 获取用户的收藏列表
func GetBookmarksByUserID(userID uint, page, pageSize int) ([]Bookmark, int64, error) {
	var bookmarks []Bookmark
	var total int64

	// 计算总数
	if err := DB.Model(&Bookmark{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := DB.Preload("Post").Preload("Post.Author").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&bookmarks).Error; err != nil {
		return nil, 0, err
	}

	return bookmarks, total, nil
}

// IsPostBookmarked 检查帖子是否被用户收藏
func IsPostBookmarked(userID, postID uint) bool {
	var count int64
	DB.Model(&Bookmark{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count)
	return count > 0
}

// GetNotificationsByUserID 获取用户的通知列表
func GetNotificationsByUserID(userID uint, page, pageSize int, unreadOnly bool) ([]Notification, int64, error) {
	var notifications []Notification
	var total int64

	query := DB.Model(&Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("FromUser").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// GetUnreadNotificationCount 获取用户未读通知数量
func GetUnreadNotificationCount(userID uint) int64 {
	var count int64
	DB.Model(&Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count)
	return count
}

// IsUserBlocked 检查用户是否被屏蔽
func IsUserBlocked(userID, blockedUserID uint) bool {
	var count int64
	DB.Model(&UserBlock{}).Where("user_id = ? AND blocked_user_id = ?", userID, blockedUserID).Count(&count)
	return count > 0
}

// GetBlockedUserIDs 获取用户屏蔽的所有用户ID列表
func GetBlockedUserIDs(userID uint) ([]uint, error) {
	var blocks []UserBlock
	if err := DB.Where("user_id = ?", userID).Find(&blocks).Error; err != nil {
		return nil, err
	}

	var blockedIDs []uint
	for _, block := range blocks {
		blockedIDs = append(blockedIDs, block.BlockedUserID)
	}

	return blockedIDs, nil
}

// GetModerationSetting 获取审核设置
func GetModerationSetting(key string) (*ModerationSetting, error) {
	var setting ModerationSetting
	if err := DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// SetModerationSetting 设置审核配置
func SetModerationSetting(key, value, description string, userID uint) error {
	setting := ModerationSetting{
		Key:             key,
		Value:           value,
		Description:     description,
		UpdatedByUserID: userID,
	}

	return DB.Save(&setting).Error
}
