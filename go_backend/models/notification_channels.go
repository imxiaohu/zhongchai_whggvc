package models

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// NotificationChannel 通知渠道配置模型
type NotificationChannel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID    uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// 邮件通知配置
	EmailEnabled bool   `gorm:"default:false" json:"emailEnabled"` // 是否启用邮件通知
	EmailAddress string `gorm:"size:100" json:"emailAddress"`      // 邮件地址

	// 钉钉通知配置
	DingTalkEnabled    bool   `gorm:"default:false" json:"dingTalkEnabled"` // 是否启用钉钉通知
	DingTalkWebhookURL string `gorm:"size:500" json:"dingTalkWebhookUrl"`   // 钉钉Webhook地址
	DingTalkSecret     string `gorm:"size:100" json:"dingTalkSecret"`       // 钉钉机器人密钥（可选）

	// 短信通知配置
	SMSEnabled  bool   `gorm:"default:false" json:"smsEnabled"` // 是否启用短信通知
	PhoneNumber string `gorm:"size:20" json:"phoneNumber"`      // 手机号码

	// 通知类型偏好 - 成绩相关
	ScoreUpdateEmail    bool `gorm:"default:true" json:"scoreUpdateEmail"`     // 成绩更新邮件通知
	ScoreUpdateDingTalk bool `gorm:"default:false" json:"scoreUpdateDingTalk"` // 成绩更新钉钉通知
	ScoreUpdateSMS      bool `gorm:"default:false" json:"scoreUpdateSms"`      // 成绩更新短信通知

	// 通知类型偏好 - 社区互动相关
	CommunityLikeEmail           bool `gorm:"default:false" json:"communityLikeEmail"`           // 帖子点赞邮件通知
	CommunityLikeDingTalk        bool `gorm:"default:false" json:"communityLikeDingTalk"`        // 帖子点赞钉钉通知
	CommunityBookmarkEmail       bool `gorm:"default:false" json:"communityBookmarkEmail"`       // 收藏邮件通知
	CommunityBookmarkDingTalk    bool `gorm:"default:false" json:"communityBookmarkDingTalk"`    // 收藏钉钉通知
	CommunityCommentEmail        bool `gorm:"default:false" json:"communityCommentEmail"`        // 评论邮件通知
	CommunityCommentDingTalk     bool `gorm:"default:false" json:"communityCommentDingTalk"`     // 评论钉钉通知
	CommunityCommentLikeEmail    bool `gorm:"default:false" json:"communityCommentLikeEmail"`    // 评论点赞邮件通知
	CommunityCommentLikeDingTalk bool `gorm:"default:false" json:"communityCommentLikeDingTalk"` // 评论点赞钉钉通知

	// 成绩检查配置
	ScoreCheckEnabled   bool       `gorm:"default:false" json:"scoreCheckEnabled"`              // 是否启用成绩检查
	ScoreCheckFrequency string     `gorm:"size:20;default:'daily'" json:"scoreCheckFrequency"`  // 检查频率: hourly, daily, weekly
	ScoreCheckTime      string     `gorm:"size:10;default:'09:00'" json:"scoreCheckTime"`       // 检查时间 HH:MM
	ScoreCheckSemester  string     `gorm:"size:50;default:'current'" json:"scoreCheckSemester"` // 检测学期: current(当前学期), all(全部学期), 或具体学期名称
	LastScoreCheck      *time.Time `json:"lastScoreCheck"`                                      // 最后检查时间

	// 上课提醒配置
	ClassReminderEnabled       bool   `gorm:"default:false" json:"classReminderEnabled"`         // 是否启用上课提醒
	ClassReminderMinutesBefore int    `gorm:"default:15" json:"classReminderMinutesBefore"`      // 上课前多少分钟提醒
	ClassReminderChannelEmail  bool   `gorm:"default:false" json:"classReminderChannelEmail"`    // 上课提醒邮件渠道
	ClassReminderChannelSMS    bool   `gorm:"default:false" json:"classReminderChannelSms"`      // 上课提醒短信渠道
	ClassReminderChannelDingTalk bool `gorm:"default:false" json:"classReminderChannelDingTalk"` // 上课提醒钉钉渠道
}

// SMSBalance 短信余额模型
type SMSBalance struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID     uint           `gorm:"not null;uniqueIndex" json:"userId"` // 用户ID，唯一索引
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Balance    int            `gorm:"default:0" json:"balance"`    // 余额（分）
	TotalSpent int            `gorm:"default:0" json:"totalSpent"` // 总消费（分）
}

// SMSTransaction 短信交易记录模型
type SMSTransaction struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID        uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User          User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type          string         `gorm:"size:20;not null" json:"type"`            // 交易类型: recharge, consume
	Amount        int            `gorm:"not null" json:"amount"`                  // 金额（分）
	Description   string         `gorm:"size:200" json:"description"`             // 交易描述
	OrderID       string         `gorm:"size:50;index" json:"orderId"`            // 订单ID
	PaymentMethod string         `gorm:"size:20" json:"paymentMethod"`            // 支付方式: wechat
	Status        string         `gorm:"size:20;default:'pending'" json:"status"` // 状态: pending, success, failed
	ExtraData     string         `gorm:"type:text" json:"extraData"`              // 额外数据（JSON格式）
}

// NotificationLog 通知发送日志模型
type NotificationLog struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID     uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Channel    string         `gorm:"size:20;not null" json:"channel"`         // 通知渠道: email, dingtalk, sms
	Type       string         `gorm:"size:50;not null" json:"type"`            // 通知类型: score_update, system, etc.
	Title      string         `gorm:"size:200" json:"title"`                   // 通知标题
	Content    string         `gorm:"type:text" json:"content"`                // 通知内容
	Recipient  string         `gorm:"size:200" json:"recipient"`               // 接收者（邮箱、手机号等）
	Status     string         `gorm:"size:20;default:'pending'" json:"status"` // 发送状态: pending, success, failed
	ErrorMsg   string         `gorm:"size:500" json:"errorMsg"`                // 错误信息
	RetryCount int            `gorm:"default:0" json:"retryCount"`             // 重试次数
	SentAt     *time.Time     `json:"sentAt"`                                  // 发送时间
	ExtraData  string         `gorm:"type:text" json:"extraData"`              // 额外数据（JSON格式）
}

// ScoreSnapshot 成绩快照模型（用于检测成绩变化）
type ScoreSnapshot struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	UserID     uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Semester   string         `gorm:"size:50;not null" json:"semester"`                  // 学期
	CourseCode string         `gorm:"size:50;not null" json:"courseCode"`                // 课程代码
	CourseName string         `gorm:"size:100;not null" json:"courseName"`               // 课程名称
	ScoreType  string         `gorm:"size:20;not null" json:"scoreType"`                 // 成绩类型: daily, exam, final
	Score      string         `gorm:"size:20" json:"score"`                              // 成绩
	Credit     float64        `gorm:"default:0" json:"credit"`                           // 学分
	GPA        float64        `gorm:"default:0" json:"gpa"`                              // 绩点
	Rank       string         `gorm:"size:20" json:"rank"`                               // 排名
	CheckSum   string         `gorm:"size:64;index" json:"checkSum"`                     // 数据校验和（用于检测变化）
	Version    string         `gorm:"size:20;not null;default:'current'" json:"version"` // 版本标识: current, previous
	IsActive   bool           `gorm:"default:true" json:"isActive"`                      // 是否为活跃版本
}

// 通知渠道常量
const (
	ChannelEmail    = "email"
	ChannelDingTalk = "dingtalk"
	ChannelSMS      = "sms"
	ChannelSystem   = "system"
)

// 通知类型常量
const (
	NotificationTypeScoreUpdate          = "score_update"
	NotificationTypeReminder             = "reminder"
	NotificationTypeCommunityLike        = "community_like"
	NotificationTypeCommunityBookmark    = "community_bookmark"
	NotificationTypeCommunityComment     = "community_comment"
	NotificationTypeCommunityCommentLike = "community_comment_like"
	NotificationTypeBatchCache         = "batch_cache"
	NotificationTypeScoreCheck         = "score_check"
)

// 交易类型常量
const (
	TransactionTypeRecharge = "recharge" // 充值
	TransactionTypeConsume  = "consume"  // 消费
)

// 交易状态常量
const (
	TransactionStatusPending = "pending" // 待处理
	TransactionStatusSuccess = "success" // 成功
	TransactionStatusFailed  = "failed"  // 失败
)

// 发送状态常量
const (
	SendStatusPending = "pending" // 待发送
	SendStatusSuccess = "success" // 发送成功
	SendStatusFailed  = "failed"  // 发送失败
)

// 检查频率常量
const (
	CheckFrequencyHourly = "hourly" // 每小时
	CheckFrequencyDaily  = "daily"  // 每天
	CheckFrequencyWeekly = "weekly" // 每周
)

// 成绩快照版本常量
const (
	ScoreVersionCurrent  = "current"  // 当前版本
	ScoreVersionPrevious = "previous" // 上一版本
)

// GetNotificationChannelByUserID 根据用户ID获取通知渠道配置
func GetNotificationChannelByUserID(userID uint) (*NotificationChannel, error) {
	var channel NotificationChannel
	result := DB.Where("user_id = ?", userID).First(&channel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有配置，创建默认配置
			defaultChannel := &NotificationChannel{
				UserID:              userID,
				EmailEnabled:        false,
				DingTalkEnabled:     false,
				SMSEnabled:          false,
				ScoreUpdateEmail:    true,
				ScoreUpdateDingTalk: false,
				ScoreUpdateSMS:      false,
				// 社区互动通知默认关闭，用户可自行开启
				CommunityLikeEmail:           false,
				CommunityLikeDingTalk:        false,
				CommunityBookmarkEmail:       false,
				CommunityBookmarkDingTalk:    false,
				CommunityCommentEmail:        false,
				CommunityCommentDingTalk:     false,
				CommunityCommentLikeEmail:    false,
				CommunityCommentLikeDingTalk: false,
				ScoreCheckEnabled:            false,
				ScoreCheckFrequency:          CheckFrequencyDaily,
				ScoreCheckTime:               "09:00",
				ScoreCheckSemester:           "current",
			}
			if err := DB.Create(defaultChannel).Error; err != nil {
				return nil, err
			}
			return defaultChannel, nil
		}
		return nil, result.Error
	}
	return &channel, nil
}

// GetSMSBalanceByUserID 根据用户ID获取短信余额
func GetSMSBalanceByUserID(userID uint) (*SMSBalance, error) {
	var balance SMSBalance
	result := DB.Where("user_id = ?", userID).First(&balance)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有余额记录，创建默认记录
			defaultBalance := &SMSBalance{
				UserID:     userID,
				Balance:    0,
				TotalSpent: 0,
			}
			if err := DB.Create(defaultBalance).Error; err != nil {
				return nil, err
			}
			return defaultBalance, nil
		}
		return nil, result.Error
	}
	return &balance, nil
}

// CreateNotificationLog 创建通知日志
func CreateNotificationLog(log *NotificationLog) error {
	return DB.Create(log).Error
}

// UpdateNotificationLog 更新通知日志
func UpdateNotificationLog(log *NotificationLog) error {
	return DB.Save(log).Error
}

// GetPendingNotificationLogs 获取待发送的通知日志
func GetPendingNotificationLogs(limit int) ([]NotificationLog, error) {
	var logs []NotificationLog
	err := DB.Where("status = ? AND retry_count < ?", SendStatusPending, 3).
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CreateSMSTransaction 创建短信交易记录
func CreateSMSTransaction(transaction *SMSTransaction) error {
	return DB.Create(transaction).Error
}

// UpdateSMSTransaction 更新短信交易记录
func UpdateSMSTransaction(transaction *SMSTransaction) error {
	return DB.Save(transaction).Error
}

// GetSMSTransactionByOrderID 根据订单ID获取交易记录
func GetSMSTransactionByOrderID(orderID string) (*SMSTransaction, error) {
	var transaction SMSTransaction
	err := DB.Where("order_id = ?", orderID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// UpdateSMSBalance 更新短信余额（事务操作）
func UpdateSMSBalance(userID uint, amount int, transactionType string, description string) error {
	log.Printf("开始更新SMS余额: 用户ID=%d, 金额=%d分, 类型=%s, 描述=%s",
		userID, amount, transactionType, description)

	return DB.Transaction(func(tx *gorm.DB) error {
		// 获取当前余额
		var balance SMSBalance
		if err := tx.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("用户%d没有余额记录，创建新记录", userID)
				// 创建新的余额记录
				balance = SMSBalance{
					UserID:     userID,
					Balance:    0,
					TotalSpent: 0,
				}
				if err := tx.Create(&balance).Error; err != nil {
					log.Printf("创建余额记录失败: 用户ID=%d, 错误: %v", userID, err)
					return err
				}
				log.Printf("创建余额记录成功: 用户ID=%d", userID)
			} else {
				log.Printf("查询余额记录失败: 用户ID=%d, 错误: %v", userID, err)
				return err
			}
		} else {
			log.Printf("找到现有余额记录: 用户ID=%d, 当前余额=%d分, 总消费=%d分",
				userID, balance.Balance, balance.TotalSpent)
		}

		// 记录更新前的余额
		oldBalance := balance.Balance
		oldTotalSpent := balance.TotalSpent

		// 更新余额
		switch transactionType {
		case TransactionTypeRecharge:
			balance.Balance += amount
			log.Printf("充值操作: 用户ID=%d, 充值金额=%d分, 余额从%d分增加到%d分",
				userID, amount, oldBalance, balance.Balance)
		case TransactionTypeConsume:
			if balance.Balance < amount {
				log.Printf("余额不足: 用户ID=%d, 当前余额=%d分, 需要消费=%d分",
					userID, balance.Balance, amount)
				return gorm.ErrInvalidData // 余额不足
			}
			balance.Balance -= amount
			balance.TotalSpent += amount
			log.Printf("消费操作: 用户ID=%d, 消费金额=%d分, 余额从%d分减少到%d分, 总消费从%d分增加到%d分",
				userID, amount, oldBalance, balance.Balance, oldTotalSpent, balance.TotalSpent)
		}

		// 保存余额
		return tx.Save(&balance).Error
	})
}

// GetScoreSnapshotsByUserID 根据用户ID获取成绩快照
func GetScoreSnapshotsByUserID(userID uint) ([]ScoreSnapshot, error) {
	var snapshots []ScoreSnapshot
	err := DB.Where("user_id = ?", userID).Find(&snapshots).Error
	return snapshots, err
}

// GetScoreSnapshotsByUserIDAndSemester 根据用户ID和学期获取成绩快照
func GetScoreSnapshotsByUserIDAndSemester(userID uint, semester string) ([]ScoreSnapshot, error) {
	var snapshots []ScoreSnapshot
	err := DB.Where("user_id = ? AND semester = ? AND version = ? AND is_active = ?",
		userID, semester, ScoreVersionCurrent, true).Find(&snapshots).Error
	return snapshots, err
}

// GetScoreSnapshotsByUserIDAndSemesterWithVersion 根据用户ID、学期和版本获取成绩快照
func GetScoreSnapshotsByUserIDAndSemesterWithVersion(userID uint, semester string, version string) ([]ScoreSnapshot, error) {
	var snapshots []ScoreSnapshot
	err := DB.Where("user_id = ? AND semester = ? AND version = ? AND is_active = ?",
		userID, semester, version, true).Find(&snapshots).Error
	return snapshots, err
}

// CreateOrUpdateScoreSnapshotWithVersion 创建或更新带版本控制的成绩快照
func CreateOrUpdateScoreSnapshotWithVersion(snapshot *ScoreSnapshot) error {
	// 确保设置默认版本
	if snapshot.Version == "" {
		snapshot.Version = ScoreVersionCurrent
	}

	var existing ScoreSnapshot
	result := DB.Where("user_id = ? AND semester = ? AND course_code = ? AND score_type = ? AND version = ?",
		snapshot.UserID, snapshot.Semester, snapshot.CourseCode, snapshot.ScoreType, snapshot.Version).First(&existing)

	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		snapshot.IsActive = true
		return DB.Create(snapshot).Error
	case result.Error != nil:
		return result.Error
	default:
		updates := map[string]interface{}{
			"course_name": snapshot.CourseName,
			"score":       snapshot.Score,
			"credit":      snapshot.Credit,
			"gpa":         snapshot.GPA,
			"rank":        snapshot.Rank,
			"check_sum":   snapshot.CheckSum,
			"is_active":   true,
			"updated_at":  time.Now(),
		}
		return DB.Model(&existing).Updates(updates).Error
	}
}

// RotateScoreSnapshots 轮换成绩快照版本（current -> previous）
func RotateScoreSnapshots(userID uint, semester string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. 将所有 previous 版本标记为非活跃并删除
		if err := tx.Model(&ScoreSnapshot{}).
			Where("user_id = ? AND semester = ? AND version = ?", userID, semester, ScoreVersionPrevious).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// 2. 将 current 版本改为 previous 版本
		if err := tx.Model(&ScoreSnapshot{}).
			Where("user_id = ? AND semester = ? AND version = ?", userID, semester, ScoreVersionCurrent).
			Update("version", ScoreVersionPrevious).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateOrUpdateScoreSnapshot 创建或更新成绩快照（保持向后兼容）
func CreateOrUpdateScoreSnapshot(snapshot *ScoreSnapshot) error {
	return CreateOrUpdateScoreSnapshotWithVersion(snapshot)
}

// BatchCreateOrUpdateScoreSnapshots 批量创建或更新成绩快照（使用 INSERT ON DUPLICATE KEY UPDATE 减少数据库往返）
// 返回失败的数量和第一个错误
func BatchCreateOrUpdateScoreSnapshots(snapshots []ScoreSnapshot) (int, error) {
	if len(snapshots) == 0 {
		return 0, nil
	}

	if len(snapshots) == 1 {
		if err := CreateOrUpdateScoreSnapshotWithVersion(&snapshots[0]); err != nil {
			return 1, err
		}
		return 0, nil
	}

	// 确保所有快照都设置了默认版本
	for i := range snapshots {
		if snapshots[i].Version == "" {
			snapshots[i].Version = ScoreVersionCurrent
		}
		snapshots[i].IsActive = true
	}

	// 批量 INSERT ON DUPLICATE KEY UPDATE
	// 只更新会变化的字段，避免更新 created_at 触发非预期行为
	valueStrings := make([]string, 0, len(snapshots))
	valueArgs := make([]interface{}, 0, len(snapshots)*10)

	for _, s := range snapshots {
		// 检查和由调用方预先计算
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			s.UserID,
			s.Semester,
			s.CourseCode,
			s.CourseName,
			s.ScoreType,
			s.Score,
			s.Credit,
			s.GPA,
			s.Rank,
			s.CheckSum,
			s.Version,
			s.IsActive,
			time.Now(),
		)
	}

	sql := fmt.Sprintf(`INSERT INTO score_snapshots
		(user_id, semester, course_code, course_name, score_type, score, credit, gpa, ` + "`rank`" + `, check_sum, version, is_active, created_at)
		VALUES %s
		ON DUPLICATE KEY UPDATE
		course_name = VALUES(course_name),
		score = VALUES(score),
		credit = VALUES(credit),
		gpa = VALUES(gpa),
		` + "`rank`" + ` = VALUES(` + "`rank`" + `),
		check_sum = VALUES(check_sum),
		is_active = VALUES(is_active),
		updated_at = VALUES(created_at)`,
		strings.Join(valueStrings, ","))

	err := DB.Exec(sql, valueArgs...).Error
	if err != nil {
		log.Printf("[BatchSnapshots] 批量写入失败: %v，fallback 到逐条写入", err)
		failCount := 0
		var lastErr error
		for i := range snapshots {
			if err2 := CreateOrUpdateScoreSnapshotWithVersion(&snapshots[i]); err2 != nil {
				failCount++
				lastErr = err2
				log.Printf("[BatchSnapshots] 单条写入失败 %s/%s: %v",
					snapshots[i].CourseName, snapshots[i].ScoreType, err2)
			}
		}
		return failCount, lastErr
	}

	return 0, nil
}

// GetPreviousScoreSnapshotsByUserIDAndSemester 获取用户指定学期的快照
// 如果存在 current 快照，说明已初始化过，下一次变化应该正常通知
// 如果不存在 current 快照，说明是首次，将变化标记为首次同步（不发送通知）
func GetPreviousScoreSnapshotsByUserIDAndSemester(userID uint, semester string) (bool, error) {
	var count int64
	err := DB.Model(&ScoreSnapshot{}).
		Where("user_id = ? AND semester = ? AND version = ? AND is_active = ?",
			userID, semester, ScoreVersionCurrent, true).
		Count(&count).Error
	return count > 0, err
}
