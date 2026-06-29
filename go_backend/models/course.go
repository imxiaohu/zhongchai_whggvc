package models

import (
	"time"

	"gorm.io/gorm"
)

// Semester 学期模型
type Semester struct {
	gorm.Model `json:",inline"`
	ID         uint      `gorm:"primarykey" json:"id"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	Code       string    `gorm:"size:20;uniqueIndex;not null" json:"code"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	IsCurrent  bool      `gorm:"default:false" json:"isCurrent"`
	Status     int       `gorm:"default:1" json:"status"` // 0: 禁用, 1: 正常
}

// Course 课程模型
type Course struct {
	gorm.Model       `json:",inline"`
	ID               uint     `gorm:"primarykey" json:"id"`
	Name             string   `gorm:"size:100;not null" json:"name"`
	Code             string   `gorm:"size:50;not null" json:"code"`
	Credit           float64  `json:"credit"`
	Hours            int      `json:"hours"`
	TeacherName      string   `gorm:"size:50" json:"teacherName"`
	Classroom        string   `gorm:"size:50" json:"classroom"`
	Weekday          int      `json:"weekday"` // 1-7 表示周一到周日
	StartTime        string   `gorm:"size:20" json:"startTime"`
	EndTime          string   `gorm:"size:20" json:"endTime"`
	StartLessonScope int      `json:"startLessonScope"` // 开始节次
	EndLessonScope   int      `json:"endLessonScope"`   // 结束节次
	Week             int      `json:"week"`             // 具体是第几周的课程
	StartWeek        int      `json:"startWeek"`        // 课程开始周次（保留用于兼容）
	EndWeek          int      `json:"endWeek"`          // 课程结束周次（保留用于兼容）
	SemesterID       uint     `json:"semesterId"`
	Semester         Semester `gorm:"foreignKey:SemesterID" json:"semester,omitempty"`
	UserID           uint     `json:"userId"`
	User             User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status           int      `gorm:"default:1" json:"status"` // 0: 禁用, 1: 正常
}

// GlobalCache 全局缓存模型 - 用于存储全校师生共用的数据
type GlobalCache struct {
	gorm.Model `json:",inline"`
	ID         uint      `gorm:"primarykey" json:"id"`
	CacheKey   string    `gorm:"size:100;uniqueIndex;not null" json:"cacheKey"` // 缓存键
	CacheData  string    `gorm:"type:text;not null" json:"cacheData"`           // 缓存数据(JSON格式)
	UpdatedAt  time.Time `json:"updatedAt"`                                     // 更新时间
	ExpiresAt  time.Time `json:"expiresAt"`                                     // 过期时间
	Status     int       `gorm:"default:1" json:"status"`                       // 0: 禁用, 1: 正常
}

// GetCurrentSemester 获取当前学期
func GetCurrentSemester() (*Semester, error) {
	var semester Semester
	result := DB.Where("is_current = ?", true).First(&semester)
	if result.Error != nil {
		return nil, result.Error
	}
	return &semester, nil
}

// GetAllSemesters 获取所有学期
func GetAllSemesters() ([]Semester, error) {
	var semesters []Semester
	result := DB.Order("start_date DESC").Find(&semesters)
	if result.Error != nil {
		return nil, result.Error
	}
	return semesters, nil
}

// GetCoursesByUserAndSemester 获取用户在指定学期的课程
func GetCoursesByUserAndSemester(userID uint, semesterID uint) ([]Course, error) {
	var courses []Course
	result := DB.Where("user_id = ? AND semester_id = ?", userID, semesterID).Find(&courses)
	if result.Error != nil {
		return nil, result.Error
	}
	return courses, nil
}

// GetCoursesByUserAndDay 获取用户在指定学期和星期几的课程
func GetCoursesByUserAndDay(userID uint, semesterID uint, weekday int, week int) ([]Course, error) {
	var courses []Course
	result := DB.Where(
		"user_id = ? AND semester_id = ? AND weekday = ? AND week = ?",
		userID, semesterID, weekday, week,
	).Find(&courses)
	if result.Error != nil {
		return nil, result.Error
	}
	return courses, nil
}

// GetCoursesByUserAndDate 根据用户和日期获取课程
func GetCoursesByUserAndDate(userID uint, date string) ([]Course, error) {
	var courses []Course
	// 简单实现：这里假设 week 和 weekday 可以通过 date 推算
	// 实际应用中可能需要更复杂的逻辑，或者在 Course 模型中增加 date 字段
	// 这里我们通过 weekday 和 week 过滤
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	weekday := int(targetDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	semester, err := GetCurrentSemester()
	if err != nil {
		return nil, err
	}

	daysSinceStart := int(targetDate.Sub(semester.StartDate).Hours() / 24)
	week := daysSinceStart/7 + 1

	result := DB.Where("user_id = ? AND weekday = ? AND week = ?", userID, weekday, week).Find(&courses)
	return courses, result.Error
}

// GetCoursesByUserAndWeek 获取用户在指定学期和周次的所有课程
func GetCoursesByUserAndWeek(userID uint, semesterID uint, week int) ([]Course, error) {
	var courses []Course
	result := DB.Where(
		"user_id = ? AND semester_id = ? AND week = ?",
		userID, semesterID, week,
	).Order("weekday ASC, start_time ASC").Find(&courses)
	if result.Error != nil {
		return nil, result.Error
	}
	return courses, nil
}

// initDefaultSemesters 初始化默认学期数据
func initDefaultSemesters() {
	// 检查是否已存在学期数据
	var count int64
	DB.Model(&Semester{}).Count(&count)
	if count > 0 {
		return
	}

	// 创建默认学期
	semesters := []Semester{
		{
			Name:      "2023-2024学年第一学期",
			Code:      "2023-2024-1",
			StartDate: time.Date(2023, 9, 1, 0, 0, 0, 0, time.Local),
			EndDate:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local),
			IsCurrent: false,
			Status:    1,
		},
		{
			Name:      "2023-2024学年第二学期",
			Code:      "2023-2024-2",
			StartDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local),
			EndDate:   time.Date(2024, 7, 31, 0, 0, 0, 0, time.Local),
			IsCurrent: true,
			Status:    1,
		},
		{
			Name:      "2024-2025学年第一学期",
			Code:      "2024-2025-1",
			StartDate: time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local),
			EndDate:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.Local),
			IsCurrent: false,
			Status:    1,
		},
	}

	for _, s := range semesters {
		DB.Create(&s)
	}
}

// GetGlobalCache 获取全局缓存
func GetGlobalCache(cacheKey string) (*GlobalCache, error) {
	var cache GlobalCache
	result := DB.Where("cache_key = ? AND status = 1", cacheKey).First(&cache)
	if result.Error != nil {
		return nil, result.Error
	}

	// 检查是否过期
	if time.Now().After(cache.ExpiresAt) {
		return nil, gorm.ErrRecordNotFound
	}

	return &cache, nil
}

// SetGlobalCache 设置全局缓存
func SetGlobalCache(cacheKey, cacheData string, ttl time.Duration) error {
	now := time.Now()
	expiresAt := now.Add(ttl)

	// 尝试更新现有缓存
	var cache GlobalCache
	result := DB.Where("cache_key = ?", cacheKey).First(&cache)

	if result.Error != nil {
		// 缓存不存在，创建新的
		cache = GlobalCache{
			CacheKey:  cacheKey,
			CacheData: cacheData,
			UpdatedAt: now,
			ExpiresAt: expiresAt,
			Status:    1,
		}
		return DB.Create(&cache).Error
	} else {
		// 缓存存在，更新数据
		cache.CacheData = cacheData
		cache.UpdatedAt = now
		cache.ExpiresAt = expiresAt
		cache.Status = 1
		return DB.Save(&cache).Error
	}
}

// DeleteExpiredGlobalCache 删除过期的全局缓存
func DeleteExpiredGlobalCache() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&GlobalCache{}).Error
}

// GetAllGlobalCaches 获取所有有效的全局缓存
func GetAllGlobalCaches() ([]GlobalCache, error) {
	var caches []GlobalCache
	result := DB.Where("status = 1 AND expires_at > ?", time.Now()).Find(&caches)
	return caches, result.Error
}
