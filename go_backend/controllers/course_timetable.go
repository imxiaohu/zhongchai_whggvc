// Package controllers handles HTTP request handlers for course-related endpoints.
// This file contains timetable retrieval functions including GetTimetableDay, GetCurrentTime,
// GetCourseTimeTableByDay, and GetCourseTimeTableByWeek.
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetTimetableDay 获取课程时间表（按天）
func GetTimetableDay(c *gin.Context) {
	var req TimetableDayRequest

	// 尝试从JSON请求体获取参数
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果JSON解析失败，尝试从Query参数获取
		req.Date = c.Query("date")
		req.Week = c.Query("week")
	}

	// 如果还是没有参数，尝试从Form获取
	if req.Date == "" && req.Week == "" {
		req.Date = c.PostForm("date")
		req.Week = c.PostForm("week")
	}

	// 参数验证
	if req.Date == "" && req.Week == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("日期或周次参数不能为空", 400))
		return
	}

	// 解析日期
	var targetDate time.Time
	var err error
	if req.Date != "" {
		targetDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("日期格式错误", 400))
			return
		}
	} else {
		// 使用当前日期
		targetDate = time.Now()
	}

	// 获取星期几（1-7表示周一到周日）
	weekday := int(targetDate.Weekday())
	if weekday == 0 {
		//nolint:ineffassign
		weekday = 7 // 将周日从0改为7
	}

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取学期信息失败", 500))
		return
	}

	// 计算当前是第几周
	//nolint:ineffassign
	week := 1
	if req.Week != "" {
		week, err = strconv.Atoi(req.Week)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("周次格式错误", 400))
			return
		}
	} else {
		// 计算当前日期是学期的第几周
		daysSinceStart := int(targetDate.Sub(semester.StartDate).Hours() / 24)
		week = daysSinceStart/7 + 1
	}

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

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("用户不存在", 404))
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 从学校服务器获取课程表数据
	courses, err := fetchTimetableFromSchool(proxyClient, user, targetDate.Format("2006-01-02"))
	if err != nil {
		fmt.Printf("从学校服务器获取课表失败: %v，尝试从本地数据库获取\n", err)
		// 尝试从本地数据库获取缓存数据
		dbCourses, dbErr := models.GetCoursesByUserAndDate(uint(userId), targetDate.Format("2006-01-02"))
		if dbErr == nil && len(dbCourses) > 0 {
			courses = dbCourses
			fmt.Printf("成功从本地数据库获取了 %d 条缓存课表数据\n", len(courses))
		} else {
			// 如果数据库也没有，则返回错误
			if isNetworkError(err) {
				c.JSON(http.StatusServiceUnavailable, utils.NewErrorResponse("学校服务器暂时关闭，且无本地缓存，请稍后再试", 503))
			} else {
				c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取课程表失败: "+err.Error(), 500))
			}
			return
		}
	}

	// 将课程数据保存到本地数据库
	savedCourses, err := saveTimetableToDatabase(courses, uint(userId))
	if err != nil {
		fmt.Printf("保存课表数据到数据库失败: %v\n", err)
		// 不影响返回，继续返回从学校服务器获取的数据
	} else {
		fmt.Printf("成功保存 %d 条课表数据到数据库\n", len(savedCourses))
	}

	// 构建响应数据 - 转换为前端期望的格式
	var courseList []gin.H
	for _, course := range courses {
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
				for i := startLessonScope; i <= endLessonScope; i++ {
					scopeParts = append(scopeParts, fmt.Sprintf("#%d#", i))
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
				for i := startLessonScope; i <= endLessonScope; i++ {
					scopeParts = append(scopeParts, fmt.Sprintf("#%d#", i))
				}
				lessonScope = strings.Join(scopeParts, ",")
			}
		}

		// 转换星期数字为中文
		weekStr := convertWeekNumberToString(course.Weekday)

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
			"weekday":                   course.Weekday,
			"nowweek":                   week,
			"nowWeek":                   week, // 前端期望的字段名
			"shortDate":                 targetDate.Format("2006-01-02"),
			"semester":                  semester.Name,
			"facultyName":               user.ClassName, // 暂时使用班级名作为学院名占位，或根据实际业务调整
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

	// 如果没有课程，返回空数组而不是 null
	if courseList == nil {
		courseList = []gin.H{}
	}

	// 检查请求来源，判断是移动端还是Web端
	userAgent := c.GetHeader("User-Agent")
	isMobileRequest := strings.Contains(userAgent, "MicroMessenger") ||
		strings.Contains(userAgent, "Mobile") ||
		c.GetHeader("X-Client-Id") != ""

	if isMobileRequest {
		// 移动端格式：返回 result.records 结构
		response := gin.H{
			"success": true,
			"message": "获取成功",
			"code":    0,
			"result": gin.H{
				"records": courseList,
			},
		}
		c.JSON(http.StatusOK, response)
	} else {
		// Web端格式：直接返回课程数组
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(courseList))
	}
}

// GetCurrentTime 获取当前时间信息（带缓存）
func GetCurrentTime(c *gin.Context) {
	cacheKey := "global:current_time"

	// 尝试从数据库缓存获取
	cache, err := models.GetGlobalCache(cacheKey)
	if err == nil {
		// 缓存命中，解析并返回数据
		var cachedData map[string]interface{}
		if json.Unmarshal([]byte(cache.CacheData), &cachedData) == nil {
			// 添加缓存标识
			cachedData["fromCache"] = true
			cachedData["cacheUpdatedAt"] = cache.UpdatedAt.Format("2006-01-02 15:04:05")

			response := gin.H{
				"success": true,
				"message": "获取成功",
				"code":    0,
				"result":  cachedData,
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

	// 从上下文获取用户ID（如果有的话）
	userIdInterface, exists := c.Get("userId")
	if !exists {
		// 如果没有用户ID，返回本地计算的时间信息
		getCurrentTimeLocal(c)
		return
	}

	// 转换用户ID
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		getCurrentTimeLocal(c)
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		getCurrentTimeLocal(c)
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		getCurrentTimeLocal(c)
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 从学校服务器获取当前时间信息
	currentTimeResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime", nil)
	if err != nil {
		// 学校服务器连接失败，返回本地时间
		getCurrentTimeLocal(c)
		return
	}

	var currentTimeData map[string]interface{}
	if err := json.Unmarshal(currentTimeResp, &currentTimeData); err != nil {
		getCurrentTimeLocal(c)
		return
	}

	// 检查响应是否成功
	success, ok := currentTimeData["success"].(bool)
	if !ok || !success {
		getCurrentTimeLocal(c)
		return
	}

	// 提取结果数据
	result, ok := currentTimeData["result"].(map[string]interface{})
	if !ok {
		getCurrentTimeLocal(c)
		return
	}

	// 添加缓存标识
	result["fromCache"] = false
	result["cacheUpdatedAt"] = time.Now().Format("2006-01-02 15:04:05")

	// 将数据缓存到数据库（缓存30分钟）
	resultJSON, _ := json.Marshal(result)
	//nolint:errcheck
	models.SetGlobalCache(cacheKey, string(resultJSON), 30*time.Minute)

	// 构建响应数据
	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result":  result,
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseTimeTableByDay 获取当日课程表
func GetCourseTimeTableByDay(c *gin.Context) {
	// 获取请求参数
	date := c.Query("date")

	// 参数验证
	if date == "" {
		// 使用当前日期
		date = time.Now().Format("2006-01-02")
	}

	// 解析日期
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("日期格式错误", 400))
		return
	}

	// 获取星期几（1-7表示周一到周日）
	weekday := int(targetDate.Weekday())
	if weekday == 0 {
		weekday = 7 // 将周日从0改为7
	}

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}

	// 计算当前是第几周
	daysSinceStart := int(targetDate.Sub(semester.StartDate).Hours() / 24)
	week := daysSinceStart/7 + 1

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

	// 检查用户是否开启了服务器同步
	syncSetting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		fmt.Printf("获取同步设置失败: %v\n", err)
		// 如果获取同步设置失败，继续使用学校服务器
		syncSetting = &models.SyncSetting{Enabled: false}
	}

	var courses []models.Course
	fromDatabase := false

	// 如果用户开启了服务器同步，优先从数据库获取
	if syncSetting.Enabled {
		fmt.Printf("用户 %d 开启了服务器同步，从数据库获取课程表数据\n", userId)
		courses, err = models.GetCoursesByUserAndDay(uint(userId), semester.ID, weekday, week)
		if err != nil {
			fmt.Printf("从数据库获取课程表失败: %v\n", err)
			// 如果数据库获取失败，回退到学校服务器
		} else if len(courses) > 0 {
			fmt.Printf("从数据库成功获取 %d 条课程数据\n", len(courses))
			// 直接使用数据库数据，跳过学校服务器请求
			fromDatabase = true
		} else {
			fmt.Printf("数据库中没有找到课程数据，回退到学校服务器\n")
		}
	}

	// 如果没有从数据库获取到数据，则从学校服务器获取
	if !fromDatabase {
		// 创建代理客户端
		proxyClient := utils.NewProxyClient()

		// 从学校服务器获取课程表数据
		courses, err = fetchTimetableFromSchool(proxyClient, user, date)
		if err != nil {
			// 检查是否是网络连接错误
			if isNetworkError(err) {
				c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			} else {
				c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取课程表失败: "+err.Error(), 500))
			}
			return
		}

		// 将课程数据保存到本地数据库
		savedCourses, err := saveTimetableToDatabase(courses, uint(userId))
		if err != nil {
			fmt.Printf("保存课表数据到数据库失败: %v\n", err)
			// 不影响返回，继续返回从学校服务器获取的数据
		} else {
			fmt.Printf("成功保存 %d 条课表数据到数据库\n", len(savedCourses))
		}
	}

	// 构建响应数据
	var courseList []gin.H
	for _, course := range courses {
		courseList = append(courseList, gin.H{
			"id":          course.ID,
			"name":        course.Name,
			"code":        course.Code,
			"credit":      course.Credit,
			"teacherName": course.TeacherName,
			"classroom":   course.Classroom,
			"weekday":     course.Weekday,
			"startTime":   course.StartTime,
			"endTime":     course.EndTime,
			"startWeek":   course.StartWeek,
			"endWeek":     course.EndWeek,
		})
	}

	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"date":     targetDate.Format("2006-01-02"),
			"weekday":  weekday,
			"week":     week,
			"semester": semester.Name,
			"courses":  courseList,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCourseTimeTableByWeek 获取本周课程表
func GetCourseTimeTableByWeek(c *gin.Context) {
	// 获取请求参数
	weekStr := c.Query("week")

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}

	// 计算当前是第几周
	//nolint:errcheck
	//nolint:errcheck
	week := 1
	if weekStr != "" {
		week, err = strconv.Atoi(weekStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("周次格式错误", 400))
			return
		//nolint:ineffassign
		}
	} else {
		// 计算当前日期是学期的第几周
		now := time.Now()
		daysSinceStart := int(now.Sub(semester.StartDate).Hours() / 24)
		week = daysSinceStart/7 + 1
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

	// 检查用户是否开启了服务器同步
	syncSetting, err := models.GetSyncSettingByUserID(uint(userId))
	if err != nil {
		fmt.Printf("获取同步设置失败: %v\n", err)
		// 如果获取同步设置失败，继续使用学校服务器
		syncSetting = &models.SyncSetting{Enabled: false}
	}

	// 获取当前学期
	semester, err = models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}

	// 构建一周的课程表
	weekCourses := make(map[int][]gin.H)

	// 如果用户开启了服务器同步，优先从数据库获取整周数据
	if syncSetting.Enabled {
		fmt.Printf("用户 %d 开启了服务器同步，从数据库获取整周课程表数据\n", userId)
		allCourses, err := models.GetCoursesByUserAndWeek(uint(userId), semester.ID, week)
		if err != nil {
			fmt.Printf("从数据库获取整周课程表失败: %v\n", err)
		} else if len(allCourses) > 0 {
			fmt.Printf("从数据库成功获取 %d 条整周课程数据\n", len(allCourses))

			// 按星期几分组课程
			for weekday := 1; weekday <= 7; weekday++ {
				var courseList []gin.H
				for _, course := range allCourses {
					if course.Weekday == weekday {
						courseList = append(courseList, gin.H{
							"id":          course.ID,
							"name":        course.Name,
							"code":        course.Code,
							"credit":      course.Credit,
							"teacherName": course.TeacherName,
							"classroom":   course.Classroom,
							"weekday":     course.Weekday,
							"startTime":   course.StartTime,
							"endTime":     course.EndTime,
							"startWeek":   course.StartWeek,
							"endWeek":     course.EndWeek,
						})
					}
				}
				weekCourses[weekday] = courseList
			}

			// 直接返回数据库数据
			response := gin.H{
				"success": true,
				"message": "获取成功",
				"code":    0,
				"result": gin.H{
					"week":     week,
					"semester": semester.Name,
					"courses":  weekCourses,
				},
			}
			c.JSON(http.StatusOK, response)
			return
		} else {
			fmt.Printf("数据库中没有找到整周课程数据，回退到学校服务器\n")
		}
	}

	// 如果没有从数据库获取到数据，检查学校服务器状态
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

	// 从学校服务器获取
	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 从学校服务器获取课程表数据（按周获取）
	fmt.Printf("开始从学校服务器获取课程表数据，周次: %d\n", week)
	allCourses, err := fetchTimetableFromSchoolByWeek(proxyClient, user, semester.Code, week)
	fmt.Printf("从学校服务器获取到 %d 条课程数据\n", len(allCourses))

	// 调试：检查获取到的课程数据
	for i, course := range allCourses {
		fmt.Printf("课程 %d: %s, StartLessonScope=%d, EndLessonScope=%d\n", i, course.Name, course.StartLessonScope, course.EndLessonScope)
	}
	if err != nil {
		// 检查是否是网络连接错误
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		// 如果获取失败，使用空数组
		allCourses = []models.Course{}
	} else {
		// 将课程数据保存到本地数据库
		savedCourses, err := saveTimetableToDatabase(allCourses, uint(userId))
		if err != nil {
			fmt.Printf("保存课表数据到数据库失败: %v\n", err)
		} else {
			fmt.Printf("成功保存 %d 条课表数据到数据库 (周次: %d)\n", len(savedCourses), week)
		}
	}

	// 按星期几分组课程
	for weekday := 1; weekday <= 7; weekday++ {
		var courses []models.Course
		for _, course := range allCourses {
			if course.Weekday == weekday {
				courses = append(courses, course)
			}
		}

		// 构建响应数据
		var courseList []gin.H
		for _, course := range courses {
			// 优先使用课程模型中存储的节次信息，如果没有则根据时间计算
			var startLessonScope, endLessonScope int
			var lessonScope string
			var lessonScopeLenght int

			// 添加调试信息
			fmt.Printf("课程 %s: StartLessonScope=%d, EndLessonScope=%d\n", course.Name, course.StartLessonScope, course.EndLessonScope)

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
					for i := startLessonScope; i <= endLessonScope; i++ {
						scopeParts = append(scopeParts, fmt.Sprintf("#%d#", i))
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
					for i := startLessonScope; i <= endLessonScope; i++ {
						scopeParts = append(scopeParts, fmt.Sprintf("#%d#", i))
					}
					lessonScope = strings.Join(scopeParts, ",")
				}
			}

			// 转换星期数字为中文
			weekStr := convertWeekNumberToString(course.Weekday)

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
				"weekday":                   course.Weekday,
				"nowweek":                   week,
				"nowWeek":                   week, // 前端期望的字段名
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

		weekCourses[weekday] = courseList
	}

	response := gin.H{
		"success": true,
		"message": "获取成功",
		"code":    0,
		"result": gin.H{
			"week":     week,
			"semester": semester.Name,
			"courses":  weekCourses,
		},
	}

	c.JSON(http.StatusOK, response)
}
