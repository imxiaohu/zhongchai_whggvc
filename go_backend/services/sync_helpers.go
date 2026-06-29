package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// fetchSemesterTimetableFromSchoolWithWeek 从学校服务器获取指定周次的课程表数据
func (s *SyncService) fetchSemesterTimetableFromSchoolWithWeek(proxyClient *utils.ProxyClient, user *models.User, currentSemester string, weekNum int) ([]models.Course, error) {
	// 构建请求URL，获取整个学期的课表数据
	// 添加nowWeek参数，指定周次来获取数据
	requestURL := fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek?current=1&size=84&currentSemester=%s&nowWeek=%d",
		url.QueryEscape(currentSemester), weekNum)

	log.Printf("请求整学期课表数据，URL: %s", requestURL)

	// 从学校服务器获取课程表数据
	timetableResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("学校服务器连接失败，请稍后再试")
	}

	// 添加调试信息：打印原始响应
	log.Printf("学校服务器原始响应: %s", string(timetableResp))

	var timetableData map[string]interface{}
	if err := json.Unmarshal(timetableResp, &timetableData); err != nil {
		return nil, fmt.Errorf("学校服务器响应格式错误: %w", err)
	}

	// 添加调试信息：打印解析后的数据结构
	log.Printf("解析后的响应数据: %+v", timetableData)

	// 检查响应是否成功
	success, ok := timetableData["success"].(bool)
	if !ok || !success {
		message, _ := timetableData["message"].(string)
		if message == "" {
			message = "学校服务器暂时不可用"
		}
		log.Printf("学校服务器返回错误: success=%v, message=%s", success, message)
		return nil, fmt.Errorf("学校服务器错误: %s", message)
	}

	// 提取课程数据
	result, ok := timetableData["result"].(map[string]interface{})
	if !ok {
		log.Printf("无法提取result字段，完整响应: %+v", timetableData)
		return nil, fmt.Errorf("学校服务器数据格式错误")
	}

	log.Printf("result字段内容: %+v", result)

	// 尝试从 records 字段获取课程数据（学校服务器实际返回的字段）
	coursesData, ok := result["records"].([]interface{})
	if !ok {
		// 如果 records 不存在，尝试 courses 字段（备用）
		coursesData, ok = result["courses"].([]interface{})
		if !ok {
			// 如果没有课程数据，返回空数组
			log.Printf("学校服务器返回的数据中没有找到课程信息，result字段: %+v", result)
			return []models.Course{}, nil
		}
	}

	log.Printf("从学校服务器获取到 %d 条课程数据", len(coursesData))

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

		// 如果没有直接的时间信息，尝试从课程时间段计算
		if startTime == "" || endTime == "" {
			startLessonScope, _ := courseData["startLessonScope"].(float64)
			endLessonScope, _ := courseData["endLessonScope"].(float64)
			startTime, endTime = s.calculateTimeFromLessonScope(int(startLessonScope), int(endLessonScope))
		}

		// 提取星期几信息
		weekdayFloat, ok := courseData["weekday"].(float64)
		var weekday int
		if ok {
			weekday = int(weekdayFloat)
		} else {
			// 如果没有weekday字段，尝试从week字段推断
			weekStr, _ := courseData["week"].(string)
			weekday = s.convertWeekStringToNumber(weekStr)
		}

		// 提取周次信息 - 对于按周同步，每个课程记录对应具体的周次
		startWeekFloat, ok := courseData["startWeek"].(float64)
		startWeek := weekNum // 使用当前请求的周次
		if ok {
			startWeek = int(startWeekFloat)
		}

		endWeekFloat, ok := courseData["endWeek"].(float64)
		endWeek := weekNum // 使用当前请求的周次
		if ok {
			endWeek = int(endWeekFloat)
		}

		// 当前周次 - 这是关键字段，表示这个课程记录是第几周的
		currentWeek := weekNum

		// 创建课程记录 - 每个记录对应具体的周次
		course := models.Course{
			Name:        courseName,
			Code:        fmt.Sprintf("COURSE_%d_W%d", time.Now().UnixNano(), weekNum),
			Credit:      0,
			TeacherName: teacherNames,
			Classroom:   classroom,
			Weekday:     weekday,
			StartTime:   startTime,
			EndTime:     endTime,
			Week:        currentWeek, // 关键：标识这是第几周的课程
			StartWeek:   startWeek,   // 保留用于兼容
			EndWeek:     endWeek,     // 保留用于兼容
			SemesterID:  semester.ID,
			UserID:      user.ID,
		}

		courses = append(courses, course)
	}

	log.Printf("成功解析 %d 条课程数据", len(courses))
	return courses, nil
}

// SyncSemesterInfo 同步服务的学期信息结构体
type SyncSemesterInfo struct {
	CurrentSemester string
	CurrentWeek     int
	WeekCount       int
}

// fetchCurrentSemesterInfo 从学校服务器获取当前学期信息
func (s *SyncService) fetchCurrentSemesterInfo(proxyClient *utils.ProxyClient, user *models.User) (*SyncSemesterInfo, error) {
	// 从学校服务器获取当前时间信息
	currentTimeResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET",
		"/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime", nil)
	if err != nil {
		return nil, fmt.Errorf("获取当前时间信息失败: %w", err)
	}

	var currentTimeData map[string]interface{}
	if err := json.Unmarshal(currentTimeResp, &currentTimeData); err != nil {
		return nil, fmt.Errorf("解析当前时间信息失败: %w", err)
	}

	log.Printf("学校服务器当前时间信息响应: %+v", currentTimeData)

	// 检查响应是否成功
	success, ok := currentTimeData["success"].(bool)
	if !ok || !success {
		message, _ := currentTimeData["message"].(string)
		if message == "" {
			message = "学校服务器暂时不可用"
		}
		return nil, fmt.Errorf("学校服务器错误: %s", message)
	}

	// 提取结果数据
	result, ok := currentTimeData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("学校服务器数据格式错误")
	}

	// 提取学期信息
	currentSemester, _ := result["currentSemester"].(string)
	if currentSemester == "" {
		// 如果没有获取到学期信息，使用默认值
		currentSemester = "2024-2025学年第二学期"
	}

	// 提取当前周
	currentWeek := 1
	if nowWeek, ok := result["nowweek"].(float64); ok {
		currentWeek = int(nowWeek)
	} else if nowWeek, ok := result["currentWeek"].(float64); ok {
		currentWeek = int(nowWeek)
	}

	// 提取总周数
	weekCount := 20 // 默认20周
	if wc, ok := result["weekCount"].(float64); ok {
		weekCount = int(wc)
	} else if wc, ok := result["totalWeeks"].(float64); ok {
		weekCount = int(wc)
	}

	semesterInfo := &SyncSemesterInfo{
		CurrentSemester: currentSemester,
		CurrentWeek:     currentWeek,
		WeekCount:       weekCount,
	}

	return semesterInfo, nil
}

// saveTimetableToDatabase 将课表数据保存到本地数据库
// P2 fix: 批量预加载现有课程，消除逐条查询的 N+1 问题
func (s *SyncService) saveTimetableToDatabase(courses []models.Course, userId uint) ([]models.Course, error) {
	var savedCourses []models.Course

	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, fmt.Errorf("获取当前学期失败: %w", err)
	}

	if len(courses) == 0 {
		return savedCourses, nil
	}

	// 批量预加载：一次性查出该用户该学期所有相关课程
	// 构建 where name IN (...) 条件
	nameSet := make(map[string]struct{})
	for _, c := range courses {
		nameSet[c.Name] = struct{}{}
	}
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}

	var existingCourses []models.Course
	// P2 fix: 一次性批量查询，避免 N+1
	if err := models.DB.Where(
		"name IN ? AND user_id = ? AND semester_id = ?",
		names, userId, semester.ID,
	).Find(&existingCourses).Error; err != nil {
		return nil, fmt.Errorf("批量查询现有课程失败: %w", err)
	}

	// 精确匹配查找表
	type courseKey struct {
		name      string
		week      int
		weekday   int
		startTime string
		endTime   string
		teacher   string
	}
	existingMap := make(map[courseKey]models.Course)
	for _, ec := range existingCourses {
		k := courseKey{
			name:      ec.Name,
			week:      ec.Week,
			weekday:   ec.Weekday,
			startTime: ec.StartTime,
			endTime:   ec.EndTime,
			teacher:   ec.TeacherName,
		}
		existingMap[k] = ec
	}

	// 批量创建和更新
	var toCreate, toUpdate []models.Course
	for _, course := range courses {
		k := courseKey{
			name:      course.Name,
			week:      course.Week,
			weekday:   course.Weekday,
			startTime: course.StartTime,
			endTime:   course.EndTime,
			teacher:   course.TeacherName,
		}
		if ec, found := existingMap[k]; found {
			// B6 fix: 整体替换 map entry，确保 GORM Save 携带完整 ID
			existingMap[k] = models.Course{
				ID:          ec.ID,
				Name:        ec.Name,
				Code:        ec.Code,
				Credit:      ec.Credit,
				TeacherName: ec.TeacherName,
				Classroom:   course.Classroom,
				Weekday:     ec.Weekday,
				StartTime:   ec.StartTime,
				EndTime:     ec.EndTime,
				Week:        ec.Week,
				StartWeek:   course.StartWeek,
				EndWeek:     course.EndWeek,
				SemesterID:  ec.SemesterID,
				UserID:      ec.UserID,
				Status:      ec.Status,
			}
			toUpdate = append(toUpdate, existingMap[k])
		} else {
			newCourse := models.Course{
				Name:       course.Name,
				Code:       course.Code,
				Credit:     course.Credit,
				TeacherName: course.TeacherName,
				Classroom:  course.Classroom,
				Weekday:    course.Weekday,
				StartTime:  course.StartTime,
				EndTime:    course.EndTime,
				Week:       course.Week,
				StartWeek:  course.StartWeek,
				EndWeek:    course.EndWeek,
				SemesterID: semester.ID,
				UserID:     userId,
			}
			toCreate = append(toCreate, newCourse)
		}
	}

	// 批量创建
	if len(toCreate) > 0 {
		if err := models.DB.CreateInBatches(toCreate, 50).Error; err != nil {
			log.Printf("批量创建课程失败: %v", err)
		} else {
			savedCourses = append(savedCourses, toCreate...)
			log.Printf("批量创建 %d 门课程", len(toCreate))
		}
	}

	// 批量更新
	if len(toUpdate) > 0 {
		for i := range toUpdate {
			if err := models.DB.Save(&toUpdate[i]).Error; err != nil {
				log.Printf("更新课程 %s 失败: %v", toUpdate[i].Name, err)
				continue
			}
			savedCourses = append(savedCourses, toUpdate[i])
		}
		log.Printf("更新 %d 门课程", len(toUpdate))
	}

	return savedCourses, nil
}

// calculateTimeFromLessonScope 根据课程时间段计算开始和结束时间
func (s *SyncService) calculateTimeFromLessonScope(startLesson, endLesson int) (string, string) {
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
func (s *SyncService) convertWeekStringToNumber(weekStr string) int {
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
