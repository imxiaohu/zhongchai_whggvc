// Package controllers handles HTTP request handlers for course-related endpoints.
// This file contains evaluation and timetable list functions including GetTermWeekNum,
// GetTimetableList, and GetCourseLessonTime.
package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetTermWeekNum 获取学期周次信息
func GetTermWeekNum(c *gin.Context) {
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

	// 从学校服务器获取当前时间信息
	currentTimeResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime", nil)
	if err != nil {
		// 检查是否是网络连接错误
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取当前时间失败: "+err.Error(), 500))
		}
		return
	}

	var currentTimeData map[string]interface{}
	if err := json.Unmarshal(currentTimeResp, &currentTimeData); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析当前时间数据失败: "+err.Error(), 500))
		return
	}

	// 检查响应是否成功
	success, ok := currentTimeData["success"].(bool)
	if !ok || !success {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("学校服务器返回错误", 500))
		return
	}

	// 提取结果数据
	result, ok := currentTimeData["result"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("当前时间数据格式错误", 500))
		return
	}

	// 提取周次信息
	currentSemester, _ := result["currentSemester"].(string)
	nowWeek, _ := result["nowweek"].(float64)
	weekCount, _ := result["weekCount"].(float64)
	shortDate, _ := result["shortDate"].(string)
	startDate, _ := result["startDate"].(string)
	week, _ := result["week"].(string)

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"currentSemester": currentSemester,
			"nowWeek":         int(nowWeek),
			"weekCount":       int(weekCount),
			"shortDate":       shortDate,
			"startDate":       startDate,
			"week":            week,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetTimetableList 获取周课表列表
func GetTimetableList(c *gin.Context) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	// 转换用户ID
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("用户ID格式错误", 500))
		return
	}

	// 解析请求参数
	var requestData struct {
		Week            string `form:"week" json:"week"`
		CurrentSemester string `form:"currentSemester" json:"currentSemester"`
	}

	// 尝试从表单数据解析
	if err := c.ShouldBind(&requestData); err != nil {
		// 如果表单解析失败，尝试从原始body解析
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("读取请求体失败", 400))
			return
		}

		// 解析URL编码的数据
		values, err := url.ParseQuery(string(bodyBytes))
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("解析请求参数失败", 400))
			return
		}

		requestData.Week = values.Get("week")
		requestData.CurrentSemester = values.Get("currentSemester")
	}

	if requestData.Week == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("周次参数不能为空", 400))
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("用户不存在", 404))
		return
	}

	// 计算周次对应的日期范围
	weekNum, err := strconv.Atoi(requestData.Week)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("周次格式错误", 400))
		return
	}

	// 获取当前学期信息
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取当前学期失败", 500))
		return
	}

	// 检查用户是否开启了服务器同步
	syncSetting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		fmt.Printf("获取同步设置失败: %v\n", err)
		// 如果获取同步设置失败，继续使用学校服务器
		syncSetting = &models.SyncSetting{Enabled: false}
	}

	// 获取一周的课表数据
	weekCourses := make(map[string][]gin.H)
	var allCourses []models.Course
	fromDatabase := false

	// 计算该周的日期范围（假设学期开始日期为2024年9月2日，周一）
	semesterStart := time.Date(2024, 9, 2, 0, 0, 0, 0, time.Local) // 假设的学期开始日期
	weekStart := semesterStart.AddDate(0, 0, (weekNum-1)*7)

	// 如果用户开启了服务器同步，优先从数据库获取
	if syncSetting.Enabled {
		fmt.Printf("用户 %d 开启了服务器同步，从数据库获取课程表数据\n", userId)
		allCourses, err = models.GetCoursesByUserAndWeek(uint(userId), semester.ID, weekNum)
		switch {
		case err != nil:
			fmt.Printf("从数据库获取课程表失败: %v\n", err)
		case len(allCourses) > 0:
			fmt.Printf("从数据库成功获取 %d 条课程数据\n", len(allCourses))
			fromDatabase = true
		default:
			fmt.Printf("数据库中没有找到课程数据，回退到学校服务器\n")
		}
	}

	// 如果没有从数据库获取到数据，则从学校服务器获取
	if !fromDatabase {
		// 创建代理客户端
		proxyClient := utils.NewProxyClient()

		// 从学校服务器获取一周的课表数据（只调用一次）
		// 使用按周获取的API
		allCourses, err = fetchTimetableFromSchoolByWeek(proxyClient, user, semester.Code, weekNum)
		if err != nil {
			// 检查是否是网络连接错误
			if isNetworkError(err) {
				c.JSON(http.StatusServiceUnavailable, utils.NewErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
				return
			}
			// 如果获取失败，使用空数组
			allCourses = []models.Course{}
		}
	}

	// 按星期几分组课程
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)

		// 计算当前日期的星期数字
		currentWeekday := int(currentDate.Weekday())
		if currentWeekday == 0 {
			currentWeekday = 7 // 将周日从0改为7
		}

		// 过滤出当天的课程
		var courseList []gin.H
		for _, course := range allCourses {
			// 只处理当天的课程（根据课程的weekday字段过滤）
			if course.Weekday != currentWeekday {
				continue // 跳过不是当天的课程
			}

			// 优先使用课程模型中存储的节次信息，如果没有则根据时间计算
			var startLessonScope, endLessonScope int
			var lessonScope string
			var lessonScopeLenght int

			// 检查课程模型是否有节次信息（从学校服务器获取的数据会有这些字段）
			if course.StartLessonScope > 0 && course.EndLessonScope > 0 {
				// 直接使用存储的节次信息
				startLessonScope = course.StartLessonScope
				endLessonScope = course.EndLessonScope
				lessonScopeLenght = endLessonScope - startLessonScope + 1

				// 生成lessonScope字符串
				if lessonScopeLenght == 1 {
					lessonScope = fmt.Sprintf("#%d#", startLessonScope)
				} else {
					var scopeParts []string
					for j := startLessonScope; j <= endLessonScope; j++ {
						scopeParts = append(scopeParts, fmt.Sprintf("#%d#", j))
					}
					lessonScope = strings.Join(scopeParts, ",")
				}
			} else {
				// 根据时间计算节次信息（兼容旧数据）
				startLessonScope, endLessonScope = calculateLessonScopeFromTime(course.StartTime, course.EndTime)
				lessonScopeLenght = endLessonScope - startLessonScope + 1

				// 生成lessonScope字符串
				if lessonScopeLenght == 1 {
					lessonScope = fmt.Sprintf("#%d#", startLessonScope)
				} else {
					var scopeParts []string
					for j := startLessonScope; j <= endLessonScope; j++ {
						scopeParts = append(scopeParts, fmt.Sprintf("#%d#", j))
					}
					lessonScope = strings.Join(scopeParts, ",")
				}
			}

			// 转换当前日期的星期数字为中文
			weekStr := convertWeekNumberToString(currentWeekday)

			courseList = append(courseList, gin.H{
				"id":                        course.ID,
				"courseName":                course.Name,
				"teacherNames":              course.TeacherName,
				"classroomName":             course.Classroom,
				"startTime":                 course.StartTime,
				"endTime":                   course.EndTime,
				"startLessonScope":          startLessonScope,
				"endLessonScope":            endLessonScope,
				"lessonScope":               lessonScope,
				"lessonScopeLenght":         lessonScopeLenght,
				"week":                      weekStr,
				"weekday":                   currentWeekday,
				"nowweek":                   weekNum,
				"nowWeek":                   weekNum,
				"shortDate":                 currentDate.Format("2006-01-02"),
				"semester":                  semester.Name,
				"facultyName":               "信息工程学院", // 默认学院名
				"courseTeachingClassTaskID": nil,
				"currentSemester":           nil,
				"startDate":                 nil,
				"studentCount":              nil,
				"teacherID":                 nil,
				"teacherName":               nil,
				"teachingClassName":         nil,
				"weekCount":                 nil,
				"name":                      nil,
			})
		}

		// 按星期几分组
		weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
		weekday := weekdays[currentDate.Weekday()]
		weekCourses[weekday] = courseList
	}

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result":  weekCourses,
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseLessonTime 获取课程时间段信息（带缓存）
func GetCourseLessonTime(c *gin.Context) {
	cacheKey := "global:course_lesson_time"

	// 尝试从数据库缓存获取
	cache, err := models.GetGlobalCache(cacheKey)
	if err == nil {
		// 缓存命中，解析并返回数据
		var cachedData []gin.H
		if json.Unmarshal([]byte(cache.CacheData), &cachedData) == nil {
			response := gin.H{
				"success":        true,
				"message":        "获取成功",
				"code":           0,
				"result":         cachedData,
				"fromCache":      true,
				"cacheUpdatedAt": cache.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
			c.JSON(http.StatusOK, response)
			return
		}
	}

	// 缓存未命中，检查学校服务器状态
	healthService := services.GetSchoolHealthCheckService()
	if !healthService.IsServerAlive() {
		// 学校服务器不可用，返回维护提示
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "学校服务器正在维护，请稍后再试",
			"code":    503,
			"result":  healthService.GetMaintenanceMessage(),
		})
		return
	}

	// 尝试从学校服务器获取
	// 从上下文获取用户ID（如果有的话）
	userIdInterface, exists := c.Get("userId")
	if !exists {
		// 如果没有用户ID，返回本地默认的时间段数据
		getCourseLessonTimeLocal(c)
		return
	}

	// 转换用户ID
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		getCourseLessonTimeLocal(c)
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		getCourseLessonTimeLocal(c)
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		getCourseLessonTimeLocal(c)
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 从学校服务器获取课程时间段信息
	lessonTimeResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime", nil)
	if err != nil {
		// 学校服务器连接失败，返回本地默认数据
		getCourseLessonTimeLocal(c)
		return
	}

	var lessonTimeData map[string]interface{}
	if err := json.Unmarshal(lessonTimeResp, &lessonTimeData); err != nil {
		getCourseLessonTimeLocal(c)
		return
	}

	// 检查响应是否成功
	success, ok := lessonTimeData["success"].(bool)
	if !ok || !success {
		getCourseLessonTimeLocal(c)
		return
	}

	// 提取结果数据
	result, ok := lessonTimeData["result"].([]interface{})
	if !ok {
		getCourseLessonTimeLocal(c)
		return
	}

	// 转换为gin.H格式
	var timePeriods []gin.H
	for _, item := range result {
		if itemMap, ok := item.(map[string]interface{}); ok {
			timePeriods = append(timePeriods, gin.H(itemMap))
		}
	}

	// 将数据缓存到数据库（缓存24小时，因为课程时间段很少变化）
	resultJSON, _ := json.Marshal(timePeriods)
	//nolint:errcheck
	models.SetGlobalCache(cacheKey, string(resultJSON), 24*time.Hour)

	// 构建响应数据
	response := gin.H{
		"success":        true,
		"message":        "获取成功",
		"code":           0,
		"result":         timePeriods,
		"fromCache":      false,
		"cacheUpdatedAt": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, response)
}

// getCourseLessonTimeLocal 获取本地默认的课程时间段信息
func getCourseLessonTimeLocal(c *gin.Context) {
	// 返回标准的课程时间段数据
	timePeriods := []gin.H{
		{"id": 1, "name": "第1节", "startTime": "08:00", "endTime": "08:45"},
		{"id": 2, "name": "第2节", "startTime": "08:55", "endTime": "09:40"},
		{"id": 3, "name": "第3节", "startTime": "10:00", "endTime": "10:45"},
		{"id": 4, "name": "第4节", "startTime": "10:55", "endTime": "11:40"},
		{"id": 5, "name": "第5节", "startTime": "14:00", "endTime": "14:45"},
		{"id": 6, "name": "第6节", "startTime": "14:55", "endTime": "15:40"},
		{"id": 7, "name": "第7节", "startTime": "16:00", "endTime": "16:45"},
		{"id": 8, "name": "第8节", "startTime": "16:55", "endTime": "17:40"},
		{"id": 9, "name": "第9节", "startTime": "19:00", "endTime": "19:45"},
		{"id": 10, "name": "第10节", "startTime": "19:55", "endTime": "20:40"},
		{"id": 11, "name": "第11节", "startTime": "20:50", "endTime": "21:35"},
	}

	// 将默认数据缓存到数据库
	cacheKey := "global:course_lesson_time"
	resultJSON, _ := json.Marshal(timePeriods)
	//nolint:errcheck
	models.SetGlobalCache(cacheKey, string(resultJSON), 24*time.Hour)

	response := gin.H{
		"success":        true,
		"message":        "获取成功",
		"code":           0,
		"result":         timePeriods,
		"fromCache":      false,
		"cacheUpdatedAt": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, response)
}
