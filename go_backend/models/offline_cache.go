package models

import "time"

type ClassScheduleCache struct {
	ClassID   string    `gorm:"primaryKey;size:64" json:"classId"`
	Term      string    `gorm:"primaryKey;size:64" json:"term"`
	Data      string    `gorm:"type:longtext" json:"data"`
	FetchedAt time.Time `gorm:"index" json:"fetchedAt"`
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (ClassScheduleCache) TableName() string { return "class_schedule_cache" }

type NewsCache struct {
	CacheKey  string    `gorm:"primaryKey;size:32" json:"cacheKey"`
	Data      string    `gorm:"type:longtext" json:"data"`
	FetchedAt time.Time `gorm:"index" json:"fetchedAt"`
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (NewsCache) TableName() string { return "news_cache" }

type PersonalScoreCache struct {
	UserID    uint      `gorm:"primaryKey" json:"userId"`
	Term      string    `gorm:"primaryKey;size:64" json:"term"`
	Data      string    `gorm:"type:longtext" json:"data"`
	FetchedAt time.Time `gorm:"index" json:"fetchedAt"`
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (PersonalScoreCache) TableName() string { return "personal_score_cache" }

type CurrentTimeCache struct {
	CacheKey  string    `gorm:"primaryKey;size:32" json:"cacheKey"`
	Data      string    `gorm:"type:longtext" json:"data"`
	FetchedAt time.Time `gorm:"index" json:"fetchedAt"`
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (CurrentTimeCache) TableName() string { return "current_time_cache" }

type CourseTimetableWeekCache struct {
	UserID    uint      `gorm:"primaryKey" json:"userId"`
	Semester  string    `gorm:"primaryKey;size:64" json:"semester"`
	Week      int       `gorm:"primaryKey" json:"week"`
	Data      string    `gorm:"type:longtext" json:"data"`
	FetchedAt time.Time `gorm:"index" json:"fetchedAt"`
	ExpiresAt time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (CourseTimetableWeekCache) TableName() string { return "course_timetable_week_cache" }
