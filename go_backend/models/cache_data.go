package models

import (
	"crypto/md5"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CacheData 缓存数据模型
type CacheData struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CacheKey    string         `gorm:"uniqueIndex;size:255;not null" json:"cacheKey"`
	CacheType   string         `gorm:"index;size:50;not null" json:"cacheType"` // 缓存类型：public, user
	UserID      string         `gorm:"index;size:100" json:"userId"`            // 用户ID，公共数据为空
	APIPath     string         `gorm:"index;size:500;not null" json:"apiPath"`  // API路径
	RequestHash string         `gorm:"index;size:64" json:"requestHash"`        // 请求参数哈希
	Data        string         `gorm:"type:longtext;not null" json:"data"`      // 缓存的JSON数据
	AccessCount int64          `gorm:"default:0" json:"accessCount"`            // 访问次数
	LastAccess  time.Time      `gorm:"index" json:"lastAccess"`                 // 最后访问时间
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// CacheAccessLog 缓存访问日志
type CacheAccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CacheKey  string    `gorm:"index;size:255;not null" json:"cacheKey"`
	UserID    string    `gorm:"index;size:100" json:"userId"`
	IPAddress string    `gorm:"size:45" json:"ipAddress"`
	UserAgent string    `gorm:"size:500" json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
}

// CacheStats 缓存统计
type CacheStats struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CacheKey        string    `gorm:"uniqueIndex;size:255;not null" json:"cacheKey"`
	TotalAccess     int64     `gorm:"default:0" json:"totalAccess"`     // 总访问次数
	DailyAccess     int64     `gorm:"default:0" json:"dailyAccess"`     // 今日访问次数
	WeeklyAccess    int64     `gorm:"default:0" json:"weeklyAccess"`    // 本周访问次数
	MonthlyAccess   int64     `gorm:"default:0" json:"monthlyAccess"`   // 本月访问次数
	UniqueUsers     int64     `gorm:"default:0" json:"uniqueUsers"`     // 独立用户数
	AvgResponseTime float64   `gorm:"default:0" json:"avgResponseTime"` // 平均响应时间(ms)
	LastResetDate   time.Time `json:"lastResetDate"`                    // 最后重置日期
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// RedisHotCache Redis热点缓存记录
type RedisHotCache struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CacheKey    string    `gorm:"uniqueIndex;size:255;not null" json:"cacheKey"`
	AccessCount int64     `gorm:"default:0" json:"accessCount"`
	IsHot       bool      `gorm:"default:false" json:"isHot"`
	HotSince    time.Time `json:"hotSince"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName 设置表名
func (CacheData) TableName() string {
	return "cache_data"
}

func (CacheAccessLog) TableName() string {
	return "cache_access_logs"
}

func (CacheStats) TableName() string {
	return "cache_stats"
}

func (RedisHotCache) TableName() string {
	return "redis_hot_cache"
}

// CacheResponse 缓存响应结构
type CacheResponse struct {
	Data       interface{} `json:"data"`
	FromCache  bool        `json:"fromCache"`
	CacheType  string      `json:"cacheType"`  // "database", "redis", "fresh"
	ServerDown bool        `json:"serverDown"` // 服务器是否离线
	Timestamp  time.Time   `json:"timestamp"`  // 响应时间戳
	UpdatedAt  time.Time   `json:"updatedAt"`  // 缓存更新时间
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 热点数据阈值配置
	HotDataThreshold    int64         `json:"hotDataThreshold"`    // 访问次数阈值，超过此值进入Redis
	HotDataTimeWindow   time.Duration `json:"hotDataTimeWindow"`   // 时间窗口
	RedisExpireTime     time.Duration `json:"redisExpireTime"`     // Redis过期时间
	DatabaseCleanupDays int           `json:"databaseCleanupDays"` // 数据库清理天数

	// 公共数据API列表
	PublicAPIs []string `json:"publicAPIs"`
}

// GetDefaultCacheConfig 获取默认缓存配置
func GetDefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		HotDataThreshold:    100,                // 100次访问后进入Redis
		HotDataTimeWindow:   24 * time.Hour,     // 24小时时间窗口
		RedisExpireTime:     7 * 24 * time.Hour, // Redis缓存7天
		DatabaseCleanupDays: 30,                 // 30天后清理数据库缓存

		// 公共数据API列表（仅包含真正的公共数据，不包含用户个人数据）
		PublicAPIs: []string{
			"/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime",      // 当前周数据
			"/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime", // 课程表时间段配置
			"/scloudoa/scs/course/tCourseScore/getSemester",                   // 学期列表
			"/api/m/scs/course/tCourseScore/getSemester",                      // 学期列表（移动端）
			"/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList",               // 新闻类型列表
			"/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId",               // 新闻列表
			"/scloudoa/sys/announcement/list",                                 // 公告列表
			"/scloudoa/sys/news/list",                                         // 系统新闻
			"/scloudoa/sys/notice/list",                                       // 通知列表
		},
	}
}

// IsPublicAPI 检查是否为公共API
func (c *CacheConfig) IsPublicAPI(apiPath string) bool {
	for _, publicAPI := range c.PublicAPIs {
		if apiPath == publicAPI {
			return true
		}
	}
	return false
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	TotalCacheEntries    int64   `json:"totalCacheEntries"`
	DatabaseCacheEntries int64   `json:"databaseCacheEntries"`
	RedisCacheEntries    int64   `json:"redisCacheEntries"`
	HitRate              float64 `json:"hitRate"`
	MissRate             float64 `json:"missRate"`
	AvgResponseTime      float64 `json:"avgResponseTime"`
	TotalRequests        int64   `json:"totalRequests"`
	CacheHits            int64   `json:"cacheHits"`
	CacheMisses          int64   `json:"cacheMisses"`
	ServerDownCount      int64   `json:"serverDownCount"`
}

// CacheKeyBuilder 缓存键构建器
type CacheKeyBuilder struct {
	APIPath     string
	UserID      string
	RequestHash string
}

// BuildKey 构建缓存键
func (b *CacheKeyBuilder) BuildKey() string {
	if b.UserID == "" {
		// 公共数据
		return fmt.Sprintf("public:%s:%s", b.APIPath, b.RequestHash)
	}
	// 用户数据
	return fmt.Sprintf("user:%s:%s:%s", b.UserID, b.APIPath, b.RequestHash)
}

// BuildRedisKey 构建Redis键
func (b *CacheKeyBuilder) BuildRedisKey() string {
	return fmt.Sprintf("hot_cache:%s", b.BuildKey())
}

// GenerateRequestHash 生成请求参数哈希
func GenerateRequestHash(params map[string]interface{}) string {
	hash := md5.New()
	for key, value := range params {
		//nolint:errcheck
		fmt.Fprintf(hash, "%s:%v", key, value)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// 数据库操作方法

// CreateCacheData 创建缓存数据
func CreateCacheData(cacheData *CacheData) error {
	return DB.Create(cacheData).Error
}

// GetCacheDataByKey 根据键获取缓存数据
func GetCacheDataByKey(cacheKey string) (*CacheData, error) {
	var cacheData CacheData
	err := DB.Where("cache_key = ?", cacheKey).First(&cacheData).Error
	if err != nil {
		return nil, err
	}
	return &cacheData, nil
}

// UpdateCacheData 更新缓存数据
func UpdateCacheData(cacheData *CacheData) error {
	return DB.Save(cacheData).Error
}

// IncrementCacheAccess 增加缓存访问次数
func IncrementCacheAccess(cacheKey string) error {
	return DB.Model(&CacheData{}).Where("cache_key = ?", cacheKey).
		Updates(map[string]interface{}{
			"access_count": gorm.Expr("access_count + 1"),
			"last_access":  time.Now(),
		}).Error
}

// GetHotCacheKeys 获取热点缓存键
func GetHotCacheKeys(threshold int64) ([]string, error) {
	var cacheKeys []string
	err := DB.Model(&CacheData{}).
		Where("access_count >= ?", threshold).
		Pluck("cache_key", &cacheKeys).Error
	return cacheKeys, err
}

// CleanupOldCacheData 清理旧的缓存数据
func CleanupOldCacheData(days int) error {
	cutoffTime := time.Now().AddDate(0, 0, -days)
	return DB.Where("last_access < ?", cutoffTime).Delete(&CacheData{}).Error
}

// CreateCacheAccessLog 创建缓存访问日志
func CreateCacheAccessLog(log *CacheAccessLog) error {
	return DB.Create(log).Error
}

// UpdateCacheStats 更新缓存统计
func UpdateCacheStats(cacheKey string, responseTime float64) error {
	var stats CacheStats
	err := DB.Where("cache_key = ?", cacheKey).First(&stats).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新的统计记录
		stats = CacheStats{
			CacheKey:        cacheKey,
			TotalAccess:     1,
			DailyAccess:     1,
			WeeklyAccess:    1,
			MonthlyAccess:   1,
			UniqueUsers:     1,
			AvgResponseTime: responseTime,
			LastResetDate:   time.Now(),
		}
		return DB.Create(&stats).Error
	} else if err != nil {
		return err
	}

	// 更新统计
	stats.TotalAccess++
	stats.DailyAccess++
	stats.WeeklyAccess++
	stats.MonthlyAccess++
	stats.AvgResponseTime = (stats.AvgResponseTime + responseTime) / 2

	return DB.Save(&stats).Error
}
