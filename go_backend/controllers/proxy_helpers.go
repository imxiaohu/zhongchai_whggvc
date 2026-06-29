// Package controllers handles HTTP request handlers for proxy-related endpoints.
// This file contains response standardization utilities, helper functions for proxy requests.
package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// proxyRequest 通用代理请求处理
func proxyRequest(pc *ProxyController, c *gin.Context, method, path string, params url.Values) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 代理请求到学校服务器
	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, method, path, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("代理请求失败: "+err.Error(), 500))
		return
	}

	// 统一化响应格式 - 学校服务器数据
	standardizedResponse, err := standardizeAPIResponse(body, DataSourceSchool, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
		return
	}

	c.JSON(http.StatusOK, standardizedResponse)
}

// parseSchoolTimetableResponse 解析学校服务器的课表响应
func parseSchoolTimetableResponse(responseBody []byte, userID uint, semesterID uint, nowWeek int) ([]models.Course, error) {
	standardizedResponse, err := standardizeAPIResponse(responseBody, DataSourceSchool, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("标准化响应失败: %w", err)
	}

	var records []interface{}
	if result, ok := standardizedResponse.Result.(map[string]interface{}); ok {
		if recordsArray, ok := result["records"].([]interface{}); ok {
			records = recordsArray
		}
	}

	var courses []models.Course
	for _, record := range records {
		if courseData, ok := record.(map[string]interface{}); ok {
			startLessonScope := getIntValue(courseData, "startLessonScope")
			endLessonScope := getIntValue(courseData, "endLessonScope")

			if startLessonScope == 0 || endLessonScope == 0 {
				startTime := getStringValue(courseData, "startTime")
				endTime := getStringValue(courseData, "endTime")
				if startTime != "" && endTime != "" {
					startLessonScope = config.GetLessonFromTime(startTime)
					endLessonScope = config.GetLessonFromTime(endTime)
				} else {
					startLessonScope = 1
					endLessonScope = 1
				}
			}

			course := models.Course{
				UserID:             userID,
				SemesterID:         semesterID,
				Name:               getStringValue(courseData, "courseName"),
				TeacherName:        getStringValue(courseData, "teacherNames"),
				Classroom:          getStringValue(courseData, "classroomName"),
				Weekday:            getWeekdayFromString(getStringValue(courseData, "week")),
				StartTime:          getStringValue(courseData, "startTime"),
				EndTime:            getStringValue(courseData, "endTime"),
				StartLessonScope:   startLessonScope,
				EndLessonScope:     endLessonScope,
				Week:               nowWeek,
			}
			courses = append(courses, course)
		}
	}

	return courses, nil
}
