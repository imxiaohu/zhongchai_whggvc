// Package controllers handles HTTP request handlers for course-related endpoints.
// This file contains database and fetch helper functions for timetable operations.
package controllers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// saveTimetableToDatabase 将课表数据保存到本地数据库
func saveTimetableToDatabase(courses []models.Course, userId uint) ([]models.Course, error) {
	var savedCourses []models.Course

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	for _, course := range courses {
		// 检查课程是否已存在（基于课程名、教师、时间、用户ID和学期ID）
		var existingCourse models.Course
		err := models.DB.Where(
			"name = ? AND teacher_name = ? AND start_time = ? AND end_time = ? AND weekday = ? AND user_id = ? AND semester_id = ?",
			course.Name, course.TeacherName, course.StartTime, course.EndTime, course.Weekday, userId, semester.ID,
		).First(&existingCourse).Error

		if err != nil {
			// 课程不存在，创建新课程
			newCourse := models.Course{
				Name:        course.Name,
				Code:        course.Code,
				Credit:      course.Credit,
				TeacherName: course.TeacherName,
				Classroom:   course.Classroom,
				Weekday:     course.Weekday,
				StartTime:   course.StartTime,
				EndTime:     course.EndTime,
				StartWeek:   course.StartWeek,
				EndWeek:     course.EndWeek,
				SemesterID:  semester.ID,
				UserID:      userId,
			}

			if err := models.DB.Create(&newCourse).Error; err != nil {
				fmt.Printf("创建课程失败: %v\n", err)
				continue
			}
			savedCourses = append(savedCourses, newCourse)
		} else {
			// 课程已存在，更新信息
			existingCourse.Classroom = course.Classroom
			existingCourse.StartWeek = course.StartWeek
			existingCourse.EndWeek = course.EndWeek
			if err := models.DB.Save(&existingCourse).Error; err != nil {
				fmt.Printf("更新课程失败: %v\n", err)
				continue
			}
			savedCourses = append(savedCourses, existingCourse)
		}
	}

	return savedCourses, nil
}

// fetchTimetableFromSchool 从学校服务器获取课程表数据（按日期）
func fetchTimetableFromSchool(proxyClient *utils.ProxyClient, user *models.User, date string) ([]models.Course, error) {
	// 从学校服务器获取课程表数据
	timetableResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByDay?date=%s", date), nil)
	if err != nil {
		return nil, fmt.Errorf("学校服务器连接失败，请稍后再试")
	}

	var timetableData map[string]interface{}
	if err := json.Unmarshal(timetableResp, &timetableData); err != nil {
		return nil, fmt.Errorf("学校服务器响应格式错误")
	}

	// 检查响应是否成功
	success, ok := timetableData["success"].(bool)
	if !ok || !success {
		message, _ := timetableData["message"].(string)
		if message == "" {
			message = "学校服务器暂时不可用"
		}
		return nil, fmt.Errorf("学校服务器错误: %s", message)
	}

	// 提取课程数据
	result, ok := timetableData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("学校服务器数据格式错误")
	}

	// 尝试从 records 字段获取课程数据（学校服务器实际返回的字段）
	coursesData, ok := result["records"].([]interface{})
	if !ok {
		// 如果 records 不存在，尝试 courses 字段（备用）
		coursesData, ok = result["courses"].([]interface{})
		if !ok {
			// 如果没有课程数据，返回空数组
			return []models.Course{}, nil
		}
	}

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	// 转换课程数据并保存到数据库
	var courses []models.Course
	for _, courseData := range coursesData {
		courseMap, ok := courseData.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析课程信息 - 使用学校服务器实际返回的字段名
		courseName, _ := courseMap["courseName"].(string)
		if courseName == "" {
			continue
		}

		// 从学校服务器响应中提取字段
		teacherNames, _ := courseMap["teacherNames"].(string)
		classroom, _ := courseMap["classroomName"].(string)

		// 解析课程时间段
		startLessonScope, _ := courseMap["startLessonScope"].(float64)
		endLessonScope, _ := courseMap["endLessonScope"].(float64)

		// 根据课程时间段计算开始和结束时间
		startTime, endTime := calculateTimeFromLessonScope(int(startLessonScope), int(endLessonScope))

		// 解析星期几 - 从 week 字段转换
		weekStr, _ := courseMap["week"].(string)
		weekday := convertWeekStringToNumber(weekStr)

		// 从学校服务器响应中提取周次信息
		nowWeekFloat, _ := courseMap["nowweek"].(float64)
		nowWeek := int(nowWeekFloat)
		if nowWeek == 0 {
			// 如果没有nowweek字段，尝试nowWeek字段
			nowWeekFloat2, _ := courseMap["nowWeek"].(float64)
			nowWeek = int(nowWeekFloat2)
		}
		if nowWeek == 0 {
			nowWeek = 1 // 默认第1周
		}

		// 默认学分为0，因为学校服务器响应中没有学分信息
		var credit float64 = 0

		// 创建课程记录
		course := models.Course{
			Name:             courseName,
			Code:             fmt.Sprintf("COURSE_%d", time.Now().UnixNano()), // 生成简单的课程代码
			Credit:           credit,
			TeacherName:      teacherNames,
			Classroom:        classroom,
			Weekday:          weekday,
			StartTime:        startTime,
			EndTime:          endTime,
			StartLessonScope: int(startLessonScope), // 保存节次信息
			EndLessonScope:   int(endLessonScope),   // 保存节次信息
			Week:             nowWeek,               // 保存周次信息
			StartWeek:        1,
			EndWeek:          16,
			SemesterID:       semester.ID,
			UserID:           user.ID,
		}

		// 检查课程是否已存在
		var existingCourse models.Course
		err := models.DB.Where("name = ? AND user_id = ? AND semester_id = ? AND weekday = ?",
			courseName, user.ID, semester.ID, weekday).First(&existingCourse).Error

		if err != nil {
			// 课程不存在，创建新课程
			if err := models.DB.Create(&course).Error; err == nil {
				courses = append(courses, course)
			}
		} else {
			// 课程已存在，更新信息
			existingCourse.TeacherName = teacherNames
			existingCourse.Credit = credit
			existingCourse.Classroom = classroom
			existingCourse.StartTime = startTime
			existingCourse.EndTime = endTime
			if err := models.DB.Save(&existingCourse).Error; err == nil {
				courses = append(courses, existingCourse)
			}
		}
	}

	return courses, nil
}

// fetchSemesterTimetableFromSchool 从学校服务器获取整个学期的课程表数据
func fetchSemesterTimetableFromSchool(proxyClient *utils.ProxyClient, user *models.User, currentSemester string) ([]models.Course, error) {
	// 构建请求URL，获取整个学期的课表数据
	requestURL := fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek?current=1&size=84&currentSemester=%s",
		url.QueryEscape(currentSemester))

	fmt.Printf("请求整学期课表数据，URL: %s\n", requestURL)

	// 从学校服务器获取课程表数据
	timetableResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("学校服务器连接失败，请稍后再试")
	}

	var timetableData map[string]interface{}
	if err := json.Unmarshal(timetableResp, &timetableData); err != nil {
		return nil, fmt.Errorf("学校服务器响应格式错误")
	}

	// 检查响应是否成功
	success, ok := timetableData["success"].(bool)
	if !ok || !success {
		message, _ := timetableData["message"].(string)
		if message == "" {
			message = "学校服务器暂时不可用"
		}
		return nil, fmt.Errorf("学校服务器错误: %s", message)
	}

	// 提取课程数据
	result, ok := timetableData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("学校服务器数据格式错误")
	}

	// 尝试从 records 字段获取课程数据（学校服务器实际返回的字段）
	coursesData, ok := result["records"].([]interface{})
	if !ok {
		// 如果 records 不存在，尝试 courses 字段（备用）
		coursesData, ok = result["courses"].([]interface{})
		if !ok {
			// 如果没有课程数据，返回空数组
			fmt.Printf("学校服务器返回的数据中没有找到课程信息\n")
			return []models.Course{}, nil
		}
	}

	fmt.Printf("从学校服务器获取到 %d 条课程数据\n", len(coursesData))

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	var courses []models.Course
	for _, courseItem := range coursesData {
		courseData, ok := courseItem.(map[string]interface{})
		if !ok {
			continue
		}

		// 提取课程信息
		courseName, _ := courseData["courseName"].(string)
		if courseName == "" {
			continue
		}

		// 提取教师姓名
		teacherNames, _ := courseData["teacherNames"].(string)

		// 提取教室信息
		classroom, _ := courseData["classroomName"].(string)

		// 提取时间信息
		startTime, _ := courseData["startTime"].(string)
		endTime, _ := courseData["endTime"].(string)

		// 提取星期几信息
		weekdayFloat, ok := courseData["weekday"].(float64)
		var weekday int
		if ok {
			weekday = int(weekdayFloat)
		} else {
			// 如果没有weekday字段，尝试从其他字段推断
			weekday = 1 // 默认周一
		}

		// 提取周次信息
		startWeekFloat, ok := courseData["startWeek"].(float64)
		startWeek := 1
		if ok {
			startWeek = int(startWeekFloat)
		}

		endWeekFloat, ok := courseData["endWeek"].(float64)
		var endWeek = 16
		if ok {
			endWeek = int(endWeekFloat)
		}

		// 创建课程记录
		course := models.Course{
			Name:        courseName,
			Code:        fmt.Sprintf("COURSE_%d", time.Now().UnixNano()),
			Credit:      0,
			TeacherName: teacherNames,
			Classroom:   classroom,
			Weekday:     weekday,
			StartTime:   startTime,
			EndTime:     endTime,
			StartWeek:   startWeek,
			EndWeek:     endWeek,
			SemesterID:  semester.ID,
			UserID:      user.ID,
		}

		courses = append(courses, course)
	}

	fmt.Printf("成功解析 %d 条课程数据\n", len(courses))
	return courses, nil
}

// fetchCoursePlanFromSchoolForCoursePlan 从学校服务器获取课程计划数据
func fetchCoursePlanFromSchoolForCoursePlan(proxyClient *utils.ProxyClient, user *models.User, currentSemester string) ([]models.Course, error) {
	// 构建请求URL
	reqURL := "/scloudoa/scs/course/tCourseScore/getCoursePlan?current=1&size=50"
	if currentSemester != "" {
		reqURL += fmt.Sprintf("&currentSemester=%s", url.QueryEscape(currentSemester))
	}

	// 从学校服务器获取课程计划数据
	coursePlanResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("获取课程计划失败: %w", err)
	}

	var coursePlanData map[string]interface{}
	if err := json.Unmarshal(coursePlanResp, &coursePlanData); err != nil {
		return nil, fmt.Errorf("解析课程计划数据失败: %w", err)
	}

	// 检查响应是否成功
	success, ok := coursePlanData["success"].(bool)
	if !ok || !success {
		return nil, fmt.Errorf("学校服务器返回错误")
	}

	// 提取课程数据
	result, ok := coursePlanData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("课程计划数据格式错误")
	}

	records, ok := result["records"].([]interface{})
	if !ok {
		// 如果没有课程数据，返回空数组
		return []models.Course{}, nil
	}

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	// 转换课程数据并保存到数据库
	var courses []models.Course
	for _, courseData := range records {
		courseMap, ok := courseData.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析课程信息
		courseName, _ := courseMap["courseName"].(string)
		if courseName == "" {
			continue
		}

		// 提取其他字段
		creditStr, _ := courseMap["credit"].(string)
		teacherNames, _ := courseMap["teacherNames"].(string)

		// 转换学分
		var credit float64 = 0
		if creditStr != "" {
			if parsedCredit, err := strconv.ParseFloat(creditStr, 64); err == nil {
				credit = parsedCredit
			}
		}

		// 清理教师姓名（去掉#号）
		if teacherNames != "" {
			teacherNames = strings.ReplaceAll(teacherNames, "#", "")
			if idx := strings.Index(teacherNames, "("); idx != -1 {
				teacherNames = teacherNames[:idx]
			}
		}

		// 创建课程记录
		course := models.Course{
			Name:        courseName,
			Code:        fmt.Sprintf("COURSE_%d", time.Now().UnixNano()), // 生成简单的课程代码
			Credit:      credit,
			TeacherName: teacherNames,
			Classroom:   "待定", // 课程计划中没有教室信息
			Weekday:     1,    // 默认周一
			StartTime:   "08:00",
			EndTime:     "09:40",
			StartWeek:   1,
			EndWeek:     16,
			SemesterID:  semester.ID,
			UserID:      user.ID,
		}

		// 检查课程是否已存在
		var existingCourse models.Course
		err := models.DB.Where("name = ? AND user_id = ? AND semester_id = ?",
			courseName, user.ID, semester.ID).First(&existingCourse).Error

		if err != nil {
			// 课程不存在，创建新课程
			if err := models.DB.Create(&course).Error; err == nil {
				courses = append(courses, course)
			}
		} else {
			// 课程已存在，更新信息
			existingCourse.TeacherName = teacherNames
			existingCourse.Credit = credit
			if err := models.DB.Save(&existingCourse).Error; err == nil {
				courses = append(courses, existingCourse)
			}
		}
	}

	return courses, nil
}

// fetchTimetableFromSchoolByWeek 从学校服务器获取课程表数据（按周获取）
func fetchTimetableFromSchoolByWeek(proxyClient *utils.ProxyClient, user *models.User, currentSemester string, week int) ([]models.Course, error) {
	classID := user.ClassName
	offlineCache := services.GetOfflineCacheService()
	if classID != "" {
		bundle, meta, found, err := offlineCache.GetClassScheduleBundle(classID, currentSemester)
		if err == nil && found && bundle.Weeks != nil {
			if cachedWeek, ok := bundle.Weeks[week]; ok && len(cachedWeek) > 0 {
				avoidPenetration := !services.IsWithinDailyRefreshWindow(time.Now(), 6, 0, 23, 30)
				fresh := !meta.ExpiresAt.IsZero() && time.Now().Before(meta.ExpiresAt)
				if avoidPenetration || fresh {
					return cachedWeek, nil
				}
			}
		}
	}

	// 构建请求URL，获取指定周次的课表数据
	requestURL := fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek?current=1&size=84&currentSemester=%s&nowWeek=%d",
		url.QueryEscape(currentSemester), week)

	fmt.Printf("请求学校服务器课表数据，URL: %s\n", requestURL)

	// 从学校服务器获取课程表数据
	timetableResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", requestURL, nil)
	if err != nil {
		if classID != "" {
			bundle, meta, found, cacheErr := offlineCache.GetClassScheduleBundle(classID, currentSemester)
			if cacheErr == nil && found && bundle.Weeks != nil {
				if cachedWeek, ok := bundle.Weeks[week]; ok && len(cachedWeek) > 0 {
					fmt.Printf("学校服务器不可用，返回班级课表缓存 class=%s term=%s week=%d cachedAt=%s\n", classID, currentSemester, week, meta.FetchedAt.Format("2006-01-02 15:04:05"))
					return cachedWeek, nil
				}
			}
		}
		return nil, fmt.Errorf("学校服务器连接失败，请稍后再试")
	}

	var timetableData map[string]interface{}
	if err := json.Unmarshal(timetableResp, &timetableData); err != nil {
		return nil, fmt.Errorf("学校服务器响应格式错误")
	}

	// 检查响应是否成功
	success, ok := timetableData["success"].(bool)
	if !ok || !success {
		message, _ := timetableData["message"].(string)
		if message == "" {
			message = "学校服务器暂时不可用"
		}
		return nil, fmt.Errorf("学校服务器错误: %s", message)
	}

	// 提取课程数据
	result, ok := timetableData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("学校服务器数据格式错误")
	}

	// 尝试从 records 字段获取课程数据（学校服务器实际返回的字段）
	coursesData, ok := result["records"].([]interface{})
	if !ok {
		// 如果 records 不存在，尝试 courses 字段（备用）
		coursesData, ok = result["courses"].([]interface{})
		if !ok {
			// 如果没有课程数据，返回空数组
			return []models.Course{}, nil
		}
	}

	// 获取当前学期
	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	// 转换课程数据
	var courses []models.Course
	for _, courseData := range coursesData {
		courseMap, ok := courseData.(map[string]interface{})
		if !ok {
			continue
		}

		// 解析课程信息 - 使用学校服务器实际返回的字段名
		courseName, _ := courseMap["courseName"].(string)
		if courseName == "" {
			continue
		}

		// 从学校服务器响应中提取字段
		teacherNames, _ := courseMap["teacherNames"].(string)
		classroom, _ := courseMap["classroomName"].(string)

		// 解析课程时间段
		startLessonScope, _ := courseMap["startLessonScope"].(float64)
		endLessonScope, _ := courseMap["endLessonScope"].(float64)

		// 根据课程时间段计算开始和结束时间
		startTime, endTime := calculateTimeFromLessonScope(int(startLessonScope), int(endLessonScope))

		// 解析星期几 - 从 week 字段转换
		weekStr, _ := courseMap["week"].(string)
		weekday := convertWeekStringToNumber(weekStr)

		// 从学校服务器响应中提取周次信息
		nowWeekFloat, _ := courseMap["nowweek"].(float64)
		nowWeek := int(nowWeekFloat)
		if nowWeek == 0 {
			// 如果没有nowweek字段，尝试nowWeek字段
			nowWeekFloat2, _ := courseMap["nowWeek"].(float64)
			nowWeek = int(nowWeekFloat2)
		}
		if nowWeek == 0 {
			nowWeek = week // 使用请求的周次
		}

		// 默认学分为0，因为学校服务器响应中没有学分信息
		var credit float64 = 0

		// 创建课程记录
		course := models.Course{
			Name:             courseName,
			Code:             fmt.Sprintf("COURSE_%d", time.Now().UnixNano()), // 生成简单的课程代码
			Credit:           credit,
			TeacherName:      teacherNames,
			Classroom:        classroom,
			Weekday:          weekday,
			StartTime:        startTime,
			EndTime:          endTime,
			StartLessonScope: int(startLessonScope), // 保存节次信息
			EndLessonScope:   int(endLessonScope),   // 保存节次信息
			Week:             nowWeek,               // 保存周次信息
			StartWeek:        1,
			EndWeek:          16,
			SemesterID:       semester.ID,
			UserID:           user.ID,
		}

		courses = append(courses, course)
	}

	if classID != "" {
		_, _ = offlineCache.PutClassScheduleWeek(classID, currentSemester, week, courses)
	}
	return courses, nil
}

// isNetworkError 检查是否是网络连接错误
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// 检查常见的网络错误
	networkErrors := []string{
		"EOF",
		"connection refused",
		"connection reset",
		"timeout",
		"network is unreachable",
		"no such host",
		"自动登录失败",
		"发送请求失败",
	}

	for _, netErr := range networkErrors {
		if strings.Contains(errStr, netErr) {
			return true
		}
	}

	return false
}

// calculateTimeFromLessonScope 根据课程时间段计算开始和结束时间
func calculateTimeFromLessonScope(startLesson, endLesson int) (string, string) {
	// 课程时间段对应表
	timeSlots := map[int][]string{
		1:  {"08:00", "08:45"},
		2:  {"08:55", "09:40"},
		3:  {"10:00", "10:45"},
		4:  {"10:55", "11:40"},
		5:  {"14:00", "14:45"},
		6:  {"14:55", "15:40"},
		7:  {"16:00", "16:45"},
		8:  {"16:55", "17:40"},
		9:  {"19:00", "19:45"},
		10: {"19:55", "20:40"},
		11: {"20:50", "21:35"},
	}

	startTime := "08:00"
	endTime := "09:40"

	if startSlot, ok := timeSlots[startLesson]; ok {
		startTime = startSlot[0]
	}
	if endSlot, ok := timeSlots[endLesson]; ok {
		endTime = endSlot[1]
	}

	return startTime, endTime
}

// convertWeekStringToNumber 将星期字符串转换为数字
func convertWeekStringToNumber(weekStr string) int {
	weekMap := map[string]int{
		"一": 1, "周一": 1, "星期一": 1,
		"二": 2, "周二": 2, "星期二": 2,
		"三": 3, "周三": 3, "星期三": 3,
		"四": 4, "周四": 4, "星期四": 4,
		"五": 5, "周五": 5, "星期五": 5,
		"六": 6, "周六": 6, "星期六": 6,
		"日": 7, "周日": 7, "星期日": 7,
	}

	if weekday, ok := weekMap[weekStr]; ok {
		return weekday
	}
	return 1 // 默认周一
}

// convertWeekNumberToString 将星期数字转换为中文字符串
func convertWeekNumberToString(weekday int) string {
	weekMap := map[int]string{
		1: "一",
		2: "二",
		3: "三",
		4: "四",
		5: "五",
		6: "六",
		7: "日",
	}

	if weekStr, ok := weekMap[weekday]; ok {
		return weekStr
	}
	return "一" // 默认周一
}

// calculateLessonScopeFromTime 根据时间计算节次范围
func calculateLessonScopeFromTime(startTime, endTime string) (int, int) {
	// 课程时间段对应表
	timeSlots := map[string]int{
		"08:00": 1, "08:45": 1,
		"08:55": 2, "09:40": 2,
		"10:00": 3, "10:45": 3,
		"10:55": 4, "11:40": 4,
		"14:00": 5, "14:45": 5,
		"14:55": 6, "15:40": 6,
		"16:00": 7, "16:45": 7,
		"16:55": 8, "17:40": 8,
		"19:00": 9, "19:45": 9,
		"19:55": 10, "20:40": 10,
		"20:50": 11, "21:35": 11,
	}

	startLesson := 1
	endLesson := 1

	if lesson, ok := timeSlots[startTime]; ok {
		startLesson = lesson
	}
	if lesson, ok := timeSlots[endTime]; ok {
		endLesson = lesson
	}

	return startLesson, endLesson
}
