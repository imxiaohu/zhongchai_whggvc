package models

import (
	"time"

	"gorm.io/gorm"
)

// Evaluation 评教模型
type Evaluation struct {
	gorm.Model  `json:",inline"`
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Description string    `gorm:"size:500" json:"description"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	SemesterID  uint      `json:"semesterId"`
	Semester    Semester  `gorm:"foreignKey:SemesterID" json:"semester,omitempty"`
	Status      int       `gorm:"default:1" json:"status"` // 0: 未开始, 1: 进行中, 2: 已结束
	IsRequired  bool      `gorm:"default:true" json:"isRequired"`
	Items       string    `gorm:"type:text" json:"items"` // JSON格式存储评教项目
}

// EvaluationResult 评教结果模型
type EvaluationResult struct {
	gorm.Model   `json:",inline"`
	ID           uint       `gorm:"primarykey" json:"id"`
	EvaluationID uint       `json:"evaluationId"`
	Evaluation   Evaluation `gorm:"foreignKey:EvaluationID" json:"evaluation,omitempty"`
	UserID       uint       `json:"userId"`
	User         User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CourseID     uint       `json:"courseId"`
	Course       Course     `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Score        float64    `json:"score"`
	Comment      string     `gorm:"size:500" json:"comment"`
	Details      string     `gorm:"type:text" json:"details"` // JSON格式存储评教详情
	Status       int        `gorm:"default:1" json:"status"`  // 0: 草稿, 1: 已提交
}

// GetActiveEvaluations 获取当前活动的评教
func GetActiveEvaluations() ([]Evaluation, error) {
	var evaluations []Evaluation
	result := DB.Where("status = ? AND start_time <= ? AND end_time >= ?", 1, time.Now(), time.Now()).Find(&evaluations)
	if result.Error != nil {
		return nil, result.Error
	}
	return evaluations, nil
}

// GetEvaluationsBySemester 获取指定学期的评教
func GetEvaluationsBySemester(semesterID uint) ([]Evaluation, error) {
	var evaluations []Evaluation
	result := DB.Where("semester_id = ?", semesterID).Find(&evaluations)
	if result.Error != nil {
		return nil, result.Error
	}
	return evaluations, nil
}

// GetEvaluationResultsByUser 获取用户的评教结果
func GetEvaluationResultsByUser(userID uint, evaluationID uint) ([]EvaluationResult, error) {
	var results []EvaluationResult
	query := DB.Where("user_id = ?", userID)
	if evaluationID > 0 {
		query = query.Where("evaluation_id = ?", evaluationID)
	}
	result := query.Find(&results)
	if result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}
