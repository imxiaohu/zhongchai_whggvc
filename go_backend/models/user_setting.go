package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserSetting 用户设置模型
type UserSetting struct {
	gorm.Model
	UserID                uint   `gorm:"uniqueIndex;not null" json:"userId"`                   // 用户ID，唯一索引
	User                  User   `gorm:"foreignKey:UserID" json:"user,omitempty"`              // 关联用户
	ClientType            string `gorm:"size:20;default:'web'" json:"clientType"`              // 客户端类型: web, mobile, wechat, app
	Language              string `gorm:"size:10;default:'zh-CN'" json:"language"`              // 语言设置: zh-CN, en-US
	Theme                 string `gorm:"size:20;default:'auto'" json:"theme"`                  // 主题设置: light, dark, auto
	Nickname              string `gorm:"size:50" json:"nickname"`                              // 用户昵称
	ErrorNotification     bool   `gorm:"default:true" json:"errorNotification"`                // 错误提示管理：是否显示错误通知
	ErrorNotificationMode string `gorm:"size:20;default:'popup'" json:"errorNotificationMode"` // 错误提示模式: popup, toast, silent

	// 服务器同步配置
	SyncEnabled      bool   `gorm:"default:false" json:"syncEnabled"`                   // 是否启用服务器同步
	SyncFrequency    string `gorm:"size:20;default:'daily'" json:"syncFrequency"`       // 同步频率: daily, weekly, every2days, every3days
	SyncTimeRange    string `gorm:"size:50;default:'08:30-22:20'" json:"syncTimeRange"` // 同步时间范围
	SyncAutoRetry    bool   `gorm:"default:true" json:"syncAutoRetry"`                  // 是否启用自动重试
	SyncNotification bool   `gorm:"default:true" json:"syncNotification"`               // 同步完成通知

	// 界面设置
	ShowWelcomeGuide   bool `gorm:"default:true" json:"showWelcomeGuide"`   // 是否显示欢迎引导
	CompactMode        bool `gorm:"default:false" json:"compactMode"`       // 紧凑模式
	ShowAvatarInHeader bool `gorm:"default:true" json:"showAvatarInHeader"` // 头部显示头像

	// 隐私设置
	DataCollection   bool `gorm:"default:true" json:"dataCollection"`   // 数据收集同意
	AnalyticsEnabled bool `gorm:"default:true" json:"analyticsEnabled"` // 分析统计启用

	// 通知设置
	PushNotification     bool `gorm:"default:true" json:"pushNotification"`     // 推送通知
	EmailNotification    bool `gorm:"default:false" json:"emailNotification"`   // 邮件通知
	NewsNotification     bool `gorm:"default:true" json:"newsNotification"`     // 新闻通知
	ScoreNotification    bool `gorm:"default:true" json:"scoreNotification"`    // 成绩通知
	ScheduleNotification bool `gorm:"default:true" json:"scheduleNotification"` // 课程表通知

	// 扩展字段
	CustomSettings string    `gorm:"type:text" json:"customSettings"` // 自定义设置JSON字符串
	LastModifiedAt time.Time `json:"lastModifiedAt"`                  // 最后修改时间
}

// GetUserSettingByUserID 根据用户ID获取用户设置
func GetUserSettingByUserID(userID uint) (*UserSetting, error) {
	var setting UserSetting
	result := DB.Where("user_id = ?", userID).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有设置，创建默认设置
			defaultSetting := &UserSetting{
				UserID:                userID,
				ClientType:            "web",
				Language:              "zh-CN",
				Theme:                 "auto",
				ErrorNotification:     true,
				ErrorNotificationMode: "popup",
				SyncEnabled:           false,
				SyncFrequency:         "daily",
				SyncTimeRange:         "08:30-22:20",
				SyncAutoRetry:         true,
				SyncNotification:      true,
				ShowWelcomeGuide:      true,
				CompactMode:           false,
				ShowAvatarInHeader:    true,
				DataCollection:        true,
				AnalyticsEnabled:      true,
				PushNotification:      true,
				EmailNotification:     false,
				NewsNotification:      true,
				ScoreNotification:     true,
				ScheduleNotification:  true,
				LastModifiedAt:        time.Now(),
			}
			if err := DB.Create(defaultSetting).Error; err != nil {
				return nil, err
			}
			return defaultSetting, nil
		}
		return nil, result.Error
	}
	return &setting, nil
}

// UpdateUserSetting 更新用户设置
func UpdateUserSetting(setting *UserSetting) error {
	setting.LastModifiedAt = time.Now()
	return DB.Save(setting).Error
}

// CreateUserSetting 创建用户设置
func CreateUserSetting(setting *UserSetting) error {
	setting.LastModifiedAt = time.Now()
	return DB.Create(setting).Error
}

// DeleteUserSetting 删除用户设置
func DeleteUserSetting(userID uint) error {
	return DB.Where("user_id = ?", userID).Delete(&UserSetting{}).Error
}

// GetAllUserSettings 获取所有用户设置（管理员功能）
func GetAllUserSettings(page, pageSize int) ([]UserSetting, int64, error) {
	var settings []UserSetting
	var total int64

	// 计算总数
	DB.Model(&UserSetting{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := DB.Preload("User").Offset(offset).Limit(pageSize).Find(&settings)

	return settings, total, result.Error
}

// BatchUpdateUserSettings 批量更新用户设置（管理员功能）
func BatchUpdateUserSettings(userIDs []uint, updates map[string]interface{}) error {
	updates["last_modified_at"] = time.Now()
	return DB.Model(&UserSetting{}).Where("user_id IN ?", userIDs).Updates(updates).Error
}
