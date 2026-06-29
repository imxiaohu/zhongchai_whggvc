package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SyncSetting 同步设置模型
type SyncSetting struct {
	gorm.Model
	UserID           uint       `gorm:"uniqueIndex;not null" json:"userId"`
	User             User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Enabled          bool       `gorm:"default:false;index:idx_sync_enabled_schedule" json:"enabled"`             // 是否启用同步; 复合索引支持调度查询
	Frequency        string     `gorm:"size:20;default:'daily'" json:"frequency"`                                 // 同步频率: daily, weekly, every2days, every3days
	TimeRange        string     `gorm:"size:50;default:'08:30-22:20'" json:"timeRange"`                           // 同步时间范围
	LastSyncAt       *time.Time `json:"lastSyncAt"`                                                               // 最后同步时间
	NextSyncAt       *time.Time `gorm:"index:idx_sync_enabled_schedule" json:"nextSyncAt"`                        // 下次同步时间; 复合索引支持调度查询
	SyncStatus       string     `gorm:"size:20;default:'idle';index:idx_sync_enabled_schedule" json:"syncStatus"` // 同步状态: idle, syncing, success, failed; 复合索引支持调度查询
	LastSyncMessage  string     `gorm:"size:500" json:"lastSyncMessage"`                                          // 最后同步消息
	CoursesCount     int        `gorm:"default:0" json:"coursesCount"`                                            // 同步的课程数量
	AutoRetryEnabled bool       `gorm:"default:true" json:"autoRetryEnabled"`                                     // 是否启用自动重试
	RetryCount       int        `gorm:"default:0" json:"retryCount"`                                              // 重试次数
	MaxRetryCount    int        `gorm:"default:3" json:"maxRetryCount"`                                           // 最大重试次数
	// 个人基础信息缓存
	PersonalInfoSyncEnabled  bool       `gorm:"default:false" json:"personalInfoSyncEnabled"`            // 是否开启个人基础信息缓存（定时从学校服务器获取）
	PersonalInfoCacheStatus  string     `gorm:"size:20;default:'active'" json:"personalInfoCacheStatus"` // 缓存状态: active(活跃), paused(暂停-用户不活跃), resumed(恢复中)
	PersonalInfoLastCachedAt *time.Time `json:"personalInfoLastCachedAt"`                                // 上次个人基础信息缓存时间
}

// SyncLog 同步日志模型
type SyncLog struct {
	gorm.Model
	UserID      uint       `gorm:"not null" json:"userId"`
	User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SyncType    string     `gorm:"size:20;not null" json:"syncType"` // 同步类型: auto, manual
	Status      string     `gorm:"size:20;not null" json:"status"`   // 同步状态: success, failed
	Message     string     `gorm:"size:500" json:"message"`          // 同步消息
	CoursesSync int        `gorm:"default:0" json:"coursesSync"`     // 同步的课程数量
	Duration    int        `gorm:"default:0" json:"duration"`        // 同步耗时（毫秒）
	ErrorDetail string     `gorm:"size:1000" json:"errorDetail"`     // 错误详情
	SyncedAt    *time.Time `json:"syncedAt"`                         // 同步时间
}

// GetSyncSettingByUserID 根据用户ID获取同步设置
func GetSyncSettingByUserID(userID uint) (*SyncSetting, error) {
	var setting SyncSetting
	result := DB.Where("user_id = ?", userID).First(&setting)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有设置，创建默认设置
			defaultSetting := &SyncSetting{
				UserID:                  userID,
				Enabled:                 false,
				Frequency:               "daily",
				TimeRange:               "08:30-22:20",
				SyncStatus:              "idle",
				AutoRetryEnabled:        true,
				MaxRetryCount:           3,
				PersonalInfoSyncEnabled: false,
				PersonalInfoCacheStatus: "active",
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

// UpdateSyncSetting 更新同步设置
func UpdateSyncSetting(setting *SyncSetting) error {
	return DB.Save(setting).Error
}

// GetEnabledSyncSettings 获取所有启用的同步设置
func GetEnabledSyncSettings() ([]SyncSetting, error) {
	var settings []SyncSetting
	result := DB.Where("enabled = ?", true).Find(&settings)
	return settings, result.Error
}

// GetSyncSettingsForSchedule 获取需要调度的同步设置
// B6 fix: 排除 sync_status=disabled 的用户，不再调度已禁用的同步
func GetSyncSettingsForSchedule() ([]SyncSetting, error) {
	var settings []SyncSetting
	now := time.Now()
	result := DB.Where("enabled = ? AND sync_status NOT IN (?, ?) AND next_sync_at IS NOT NULL AND next_sync_at <= ?",
		true, "syncing", "disabled", now).Find(&settings)
	return settings, result.Error
}

// CreateSyncLog 创建同步日志
func CreateSyncLog(log *SyncLog) error {
	return DB.Create(log).Error
}

// GetSyncLogsByUserID 根据用户ID获取同步日志
func GetSyncLogsByUserID(userID uint, limit int) ([]SyncLog, error) {
	var logs []SyncLog
	result := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)
	return logs, result.Error
}

// GetLatestSyncTimeByUserID 根据用户ID获取最新的同步时间
func GetLatestSyncTimeByUserID(userID uint) (*time.Time, error) {
	var log SyncLog
	result := DB.Where("user_id = ? AND status = ?", userID, "success").
		Order("created_at DESC").
		First(&log)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 没有找到同步记录
		}
		return nil, result.Error
	}

	return &log.CreatedAt, nil
}

// CalculateNextSyncTime 计算下次同步时间
// B3 fix: 正确解析并遵守用户的 TimeRange 设置
func (s *SyncSetting) CalculateNextSyncTime() time.Time {
	now := time.Now()

	// 解析用户设置的 TimeRange，默认为 08:30-22:20
	startHour, startMin := 8, 30
	endHour, endMin := 22, 20

	if s.TimeRange != "" {
		parts := strings.Split(s.TimeRange, "-")
		if len(parts) == 2 {
			if h, m, err := parseTimeToMinutes(parts[0]); err == nil {
				startHour, startMin = h, m
			}
			if h, m, err := parseTimeToMinutes(parts[1]); err == nil {
				endHour, endMin = h, m
			}
		}
	}

	// 根据频率计算基础时间
	var nextSync time.Time
	switch s.Frequency {
	case "daily":
		nextSync = now.Add(24 * time.Hour)
	case "weekly":
		nextSync = now.Add(7 * 24 * time.Hour)
	case "every2days":
		nextSync = now.Add(2 * 24 * time.Hour)
	case "every3days":
		nextSync = now.Add(3 * 24 * time.Hour)
	default:
		nextSync = now.Add(24 * time.Hour)
	}

	// 将下次同步时间对齐到用户设置的时间范围起点
	nextSync = time.Date(nextSync.Year(), nextSync.Month(), nextSync.Day(),
		startHour, startMin, 0, 0, nextSync.Location())

	endToday := time.Date(nextSync.Year(), nextSync.Month(), nextSync.Day(),
		endHour, endMin, 0, 0, nextSync.Location())

	if nextSync.After(endToday) || nextSync.Equal(endToday) {
		nextSync = nextSync.Add(24 * time.Hour)
		nextSync = time.Date(nextSync.Year(), nextSync.Month(), nextSync.Day(),
			startHour, startMin, 0, 0, nextSync.Location())
	}

	return nextSync
}

// parseTimeToMinutes 辅助函数：将 "HH:MM" 解析为 (hour, minute, error)
func parseTimeToMinutes(timeStr string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(timeStr), ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("时间格式错误")
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m, nil
}

// UpdateSyncStatus 更新同步状态
// B6 fix: 修复 max_retry_count 不生效、失败后 next_sync_at 写死次日 8:30 的问题
//   - 瞬时错误：达到 max_retry_count 则禁用同步并停止调度
//   - 永久错误（如学校密码未设置）：直接禁用 + 停止调度
//   - 失败后使用指数退避计算下次同步时间，不再写死 8:30
func (s *SyncSetting) UpdateSyncStatus(status, message string, permanentErr bool) error {
	s.SyncStatus = status
	s.LastSyncMessage = message
	if status == "success" || status == "failed" {
		now := time.Now()
		s.LastSyncAt = &now
		if status == "success" {
			s.RetryCount = 0
			nextSync := s.CalculateNextSyncTime()
			s.NextSyncAt = &nextSync
		} else {
			s.RetryCount++
			if s.AutoRetryEnabled {
				if permanentErr {
					// 永久错误：直接禁用同步，不再重试
					s.Enabled = false
					s.SyncStatus = "disabled"
					s.RetryCount = 0
					s.NextSyncAt = nil
				} else if s.RetryCount >= s.MaxRetryCount {
					// 瞬时错误达到重试上限：禁用同步，等待人工介入
					s.Enabled = false
					s.SyncStatus = "disabled"
					s.NextSyncAt = nil
					log.Printf("[SyncService] 用户 %d 同步重试次数已达上限(%d/%d)，已禁用自动同步",
						s.UserID, s.RetryCount, s.MaxRetryCount)
				} else {
					// 瞬时错误未达上限：指数退避计算下次同步时间
					backoffMinutes := calculateBackoffMinutes(s.RetryCount)
					nextSync := now.Add(time.Duration(backoffMinutes) * time.Minute)
					// 截断秒和毫秒
					nextSync = time.Date(nextSync.Year(), nextSync.Month(), nextSync.Day(),
						nextSync.Hour(), nextSync.Minute(), 0, 0, nextSync.Location())
					s.NextSyncAt = &nextSync
					log.Printf("[SyncService] 用户 %d 同步失败(第 %d 次/%d)，%.0f 分钟后重试",
						s.UserID, s.RetryCount, s.MaxRetryCount, float64(backoffMinutes))
				}
			} else {
				s.NextSyncAt = nil
			}
		}
	}
	return UpdateSyncSetting(s)
}

// calculateBackoffMinutes 计算指数退避分钟数
// 重试次数: 1→5min, 2→15min, 3→30min, 4→60min, 5→120min, >=6→240min
func calculateBackoffMinutes(retryCount int) int {
	switch {
	case retryCount == 1:
		return 5
	case retryCount == 2:
		return 15
	case retryCount == 3:
		return 30
	case retryCount == 4:
		return 60
	case retryCount == 5:
		return 120
	default:
		return 240
	}
}
