// Package controllers handles HTTP request handlers for course-related endpoints.
// This file contains score-related functions including GetScoreCurrentTime, GetSemester,
// GetLearningData, TestSchoolServerConnection, and GetCoursePlan.
package controllers

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// GetScoreCurrentTime 获取成绩系统当前时间信息
func GetScoreCurrentTime(c *gin.Context) {
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
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetSemester 获取学期信息
func GetSemester(c *gin.Context) {
	// 获取所有学期
	semesters, err := models.GetAllSemesters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}

	// 构建响应数据
	var semesterList []gin.H
	for _, semester := range semesters {
		semesterList = append(semesterList, gin.H{
			"id":        semester.ID,
			"name":      semester.Name,
			"code":      semester.Code,
			"startDate": semester.StartDate.Format("2006-01-02"),
			"endDate":   semester.EndDate.Format("2006-01-02"),
			"isCurrent": semester.IsCurrent,
		})
	}

	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result":  semesterList,
	}

	c.JSON(http.StatusOK, response)
}

// GetLearningData 获取学习数据
func GetLearningData(c *gin.Context) {
	// 从上下文获取用户ID
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	// 生成模拟数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"totalCredit":      120.5,
			"earnedCredit":     85.0,
			"gpa":              3.5,
			"failedCourses":    2,
			"excellentCourses": 15,
			"goodCourses":      8,
			"passCourses":      5,
		},
	}

	c.JSON(http.StatusOK, response)
}

// TestSchoolServerConnection 测试学校服务器连接
func TestSchoolServerConnection(c *gin.Context) {
	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 测试基本连接
	fmt.Printf("测试学校服务器连接: %s\n", proxyClient.BaseURL)

	// 尝试访问一个简单的端点
	resp, err := proxyClient.HTTPClient.Get(proxyClient.BaseURL + "/scloud/init")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("连接学校服务器失败: %v", err),
			"server":  proxyClient.BaseURL,
		})
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("读取响应失败: %v", err),
			"server":  proxyClient.BaseURL,
		})
		return
	}

	fmt.Printf("连接成功，响应长度: %d\n", len(body))

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "学校服务器连接正常",
		"server":       proxyClient.BaseURL,
		"statusCode":   resp.StatusCode,
		"responseSize": len(body),
		"response":     string(body)[:int(math.Min(float64(len(body)), 500))],
	})
}

// GetCoursePlan 获取课程计划
func GetCoursePlan(c *gin.Context) {
	// 获取请求参数
	currentStr := c.Query("current")
	sizeStr := c.Query("size")
	currentSemester := c.Query("currentSemester")

	// 参数验证和默认值
	current := 1
	size := 10
	var err error
	if currentStr != "" {
		current, err = strconv.Atoi(currentStr)
		if err != nil || current < 1 {
			current = 1
		}
	}
	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size < 1 {
			size = 10
		}
	}

	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	// 转换用户ID
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 从学校服务器获取课程计划数据
	courses, err := fetchCoursePlanFromSchoolForCoursePlan(proxyClient, user, currentSemester)
	if err != nil {
		// 检查是否是网络连接错误
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取课程计划失败: "+err.Error(), 500))
		}
		return
	}

	// 分页处理
	total := len(courses)
	start := (current - 1) * size
	end := start + size
	if start >= total {
		start = 0
		end = 0
	}
	if end > total {
		end = total
	}

	var pagedCourses []models.Course
	if start < end {
		pagedCourses = courses[start:end]
	} else {
		pagedCourses = []models.Course{}
	}

	// 构建响应数据
	var courseList []gin.H
	for _, course := range pagedCourses {
		courseList = append(courseList, gin.H{
			"id":          course.ID,
			"name":        course.Name,
			"code":        course.Code,
			"credit":      course.Credit,
			"hours":       course.Hours,
			"teacherName": course.TeacherName,
			"score":       generateRandomScore(),
			"status":      "已修读",
		})
	}

	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"records":     courseList,
			"total":       total,
			"size":        size,
			"current":     current,
			"pages":       (total + size - 1) / size,
			"searchCount": true,
		},
	}

	c.JSON(http.StatusOK, response)
}

// generateRandomScore 生成随机成绩
func generateRandomScore() float64 {
	// 生成60-100之间的随机成绩
	score := 60.0 + float64(rand.Intn(41))
	// 保留一位小数
	score = float64(int(score*10)) / 10
	return score
}
