// Package controllers handles HTTP request handlers for course-related endpoints.
// This file contains timetable helper functions used by the timetable handlers.
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// TimetableDayRequest 获取课表请求结构
type TimetableDayRequest struct {
	Date string `json:"date" form:"date"`
	Week string `json:"week" form:"week"`
}

// getCurrentTimeLocal 获取本地计算的当前时间信息
func getCurrentTimeLocal(c *gin.Context) {
	// 获取当前时间
	now := time.Now()

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}

	// 计算当前是第几周
	daysSinceStart := int(now.Sub(semester.StartDate).Hours() / 24)
	week := daysSinceStart/7 + 1

	// 获取星期几（1-7表示周一到周日）
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"currentSemester": semester.Code,
			"currentWeek":     week,
			"currentDay":      weekday,
			"date":            now.Format("2006-01-02"),
			"time":            now.Format("15:04:05"),
			"fromCache":       false,
			"cacheUpdatedAt":  now.Format("2006-01-02 15:04:05"),
		},
	}

	c.JSON(http.StatusOK, response)
}
