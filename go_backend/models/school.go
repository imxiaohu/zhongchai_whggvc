package models

import (
	"gorm.io/gorm"
)

// School 学校模型
type School struct {
	gorm.Model  `json:",inline"`
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Logo        string `gorm:"size:255" json:"logo"`
	Address     string `gorm:"size:255" json:"address"`
	Description string `gorm:"size:500" json:"description"`
	Website     string `gorm:"size:255" json:"website"`
	ApiBaseUrl  string `gorm:"size:255" json:"apiBaseUrl"`
	Status      int    `gorm:"default:1" json:"status"` // 0: 禁用, 1: 正常
}

// FindSchoolByCode 根据学校代码查找学校
func FindSchoolByCode(code string) (*School, error) {
	var school School
	result := DB.Where("code = ?", code).First(&school)
	if result.Error != nil {
		return nil, result.Error
	}
	return &school, nil
}

// GetAllSchools 获取所有学校
func GetAllSchools() ([]School, error) {
	var schools []School
	result := DB.Find(&schools)
	if result.Error != nil {
		return nil, result.Error
	}
	return schools, nil
}

// initDefaultSchool 初始化默认学校数据
func initDefaultSchool() {
	// 检查是否已存在学校数据
	var count int64
	DB.Model(&School{}).Count(&count)
	if count > 0 {
		return
	}

	// 创建默认学校
	defaultSchool := School{
		Name:        "武汉光谷职业学院",
		Code:        "whggvc",
		Logo:        "https://xs.whggvc.net/logo.png",
		Address:     "湖北省武汉市洪山区光谷大道",
		Description: "武汉光谷职业学院是一所位于湖北省武汉市的全日制普通高等职业院校。",
		Website:     "https://www.whggvc.net",
		ApiBaseUrl:  "https://xs.whggvc.net",
		Status:      1,
	}

	DB.Create(&defaultSchool)
}
