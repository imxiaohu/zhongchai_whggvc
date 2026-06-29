package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

type OfflineCacheScheduler struct {
	cron *cron.Cron
}

var globalOfflineCacheScheduler *OfflineCacheScheduler

func InitOfflineCacheScheduler() {
	if !config.GetOfflineCacheSchedulerEnabled() {
		log.Printf("OfflineCacheScheduler disabled")
		return
	}

	locName := config.GetOfflineCacheSchedulerTZ()
	loc, err := time.LoadLocation(locName)
	if err != nil {
		loc = time.Local
	}

	c := cron.New(cron.WithLocation(loc))
	_, _ = c.AddFunc("0 6 * * *", func() { refreshPersonalScoreCache("06:00") })
	_, _ = c.AddFunc("0 12 * * *", func() { refreshPersonalScoreCache("12:00") })
	_, _ = c.AddFunc("0 18 * * *", func() { refreshPersonalScoreCache("18:00") })
	_, _ = c.AddFunc("30 6 * * *", refreshNewsCache)
	_, _ = c.AddFunc("0 6 * * 0", refreshClassScheduleCacheWeekly)

	c.Start()
	globalOfflineCacheScheduler = &OfflineCacheScheduler{cron: c}
	log.Printf("OfflineCacheScheduler started (tz=%s)", loc.String())
}

func GetOfflineCacheScheduler() *OfflineCacheScheduler {
	if globalOfflineCacheScheduler == nil {
		InitOfflineCacheScheduler()
	}
	return globalOfflineCacheScheduler
}

func (s *OfflineCacheScheduler) Stop() {
	if s == nil || s.cron == nil {
		return
	}
	s.cron.Stop()
}

func refreshPersonalScoreCache(tag string) {
	start := startOfDay(time.Now())
	end := start.Add(24 * time.Hour)

	var users []models.User
	err := models.DB.Where("last_login_at >= ? AND last_login_at < ? AND status = 1", start, end).Find(&users).Error
	if err != nil {
		log.Printf("WARN refreshPersonalScoreCache failed to query users: %v", err)
		return
	}
	if len(users) == 0 {
		return
	}

	sort.Slice(users, func(i, j int) bool { return users[i].LastLoginAt.After(users[j].LastLoginAt) })

	proxyClient := utils.NewProxyClient()
	offlineCache := GetOfflineCacheService()

	maxUsers := 500
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}

	for i := range users {
		u := users[i]
		term := u.CurrentSemester
		if term == "" {
			term = "__default__"
		}

		params := url.Values{}
		params.Set("current", "1")
		params.Set("size", "50")
		params.Set("currentSemester", term)
		if body, err := proxyClient.ProxyRequestWithAutoLogin(&u, "GET", "/scloudoa/scs/course/tCourseScore/getCourseScore", params); err == nil {
			_, _ = offlineCache.UpsertPersonalScorePart(u.ID, term, "scoreList", body, 12*time.Hour)
		} else {
			log.Printf("WARN personal score refresh(%s) scoreList failed user=%d: %v", tag, u.ID, err)
		}

		params2 := url.Values{}
		params2.Set("semesterName", term)
		params2.Set("currentSemester", term)
		if body, err := proxyClient.ProxyRequestWithAutoLogin(&u, "GET", "/scloudoa/scs/course/tCourseScore/getSemesterScore", params2); err == nil {
			_, _ = offlineCache.UpsertPersonalScorePart(u.ID, term, "semesterScore", body, 12*time.Hour)
		} else {
			log.Printf("WARN personal score refresh(%s) semesterScore failed user=%d: %v", tag, u.ID, err)
		}

		if body, err := proxyClient.ProxyRequestWithAutoLogin(&u, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", nil); err == nil {
			_, _ = offlineCache.UpsertPersonalScorePart(u.ID, term, "evaluationConfigList", body, 12*time.Hour)
		} else {
			log.Printf("WARN personal score refresh(%s) evaluationConfigList failed user=%d: %v", tag, u.ID, err)
		}
	}
}

func refreshNewsCache() {
	user, ok := pickAnyActiveUser()
	if !ok {
		return
	}
	proxyClient := utils.NewProxyClient()
	offlineCache := GetOfflineCacheService()

	params := url.Values{}
	params.Set("pageNo", "1")
	params.Set("pageSize", "-1")
	body, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList", params)
	if err != nil {
		log.Printf("WARN refreshNewsCache type list failed: %v", err)
		return
	}
	_ = offlineCache.PutNewsCache("eoa_news_type_list", body, 30*time.Minute)

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return
	}
	result, _ := parsed["result"].([]interface{})
	for _, it := range result {
		m, _ := it.(map[string]interface{})
		idFloat, _ := m["id"].(float64)
		id := int(idFloat)
		if id == 0 {
			continue
		}
		p := url.Values{}
		p.Set("pageNo", "1")
		p.Set("pageSize", "6")
		p.Set("ids", fmt.Sprintf("%d", id))
		listBody, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId", p)
		if err != nil {
			continue
		}
		key := fmt.Sprintf("eoa_news_list:%s:%s:%s", p.Get("ids"), p.Get("pageNo"), p.Get("pageSize"))
		_ = offlineCache.PutNewsCache(key, listBody, 30*time.Minute)
	}
}

func refreshClassScheduleCacheWeekly() {
	maxWeek := config.GetOfflineCacheClassScheduleMaxWeek()
	if maxWeek <= 0 {
		return
	}

	var classNames []string
	if err := models.DB.Model(&models.User{}).Distinct("class_name").Where("class_name <> '' AND status = 1").Pluck("class_name", &classNames).Error; err != nil {
		log.Printf("WARN refreshClassScheduleCacheWeekly list classes failed: %v", err)
		return
	}
	if len(classNames) == 0 {
		return
	}

	proxyClient := utils.NewProxyClient()
	offlineCache := GetOfflineCacheService()

	for _, classID := range classNames {
		var u models.User
		err := models.DB.Where("class_name = ? AND status = 1", classID).Order("last_login_at desc").First(&u).Error
		if err != nil {
			continue
		}
		term := u.CurrentSemester
		if term == "" {
			term = "__default__"
		}
		for week := 1; week <= maxWeek; week++ {
			requestURL := fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek?current=1&size=84&currentSemester=%s&nowWeek=%d",
				url.QueryEscape(term), week)
			resp, err := proxyClient.ProxyRequestWithAutoLogin(&u, "GET", requestURL, nil)
			if err != nil {
				continue
			}
			courses, err := parseCoursesFromWeekTimetableResponse(resp, &u, week)
			if err != nil {
				continue
			}
			_, _ = offlineCache.PutClassScheduleWeek(classID, term, week, courses)
		}
	}
}

func pickAnyActiveUser() (*models.User, bool) {
	var u models.User
	err := models.DB.Where("status = 1").Order("last_login_at desc").First(&u).Error
	if err != nil {
		return nil, false
	}
	return &u, true
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func parseCoursesFromWeekTimetableResponse(raw []byte, user *models.User, week int) ([]models.Course, error) {
	var timetableData map[string]interface{}
	if err := json.Unmarshal(raw, &timetableData); err != nil {
		return nil, err
	}
	success, _ := timetableData["success"].(bool)
	if !success {
		return nil, fmt.Errorf("school response not success")
	}
	result, ok := timetableData["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing result")
	}
	coursesData, ok := result["records"].([]interface{})
	if !ok {
		coursesData, ok = result["courses"].([]interface{})
		if !ok {
			return []models.Course{}, nil
		}
	}

	semester, err := models.GetCurrentSemester()
	if err != nil {
		return nil, err
	}

	var courses []models.Course
	for _, courseItem := range coursesData {
		courseMap, ok := courseItem.(map[string]interface{})
		if !ok {
			continue
		}
		courseName, _ := courseMap["courseName"].(string)
		if courseName == "" {
			continue
		}
		teacherNames, _ := courseMap["teacherNames"].(string)
		classroom, _ := courseMap["classroomName"].(string)
		startLessonScope, _ := courseMap["startLessonScope"].(float64)
		endLessonScope, _ := courseMap["endLessonScope"].(float64)
		startTime, endTime := calculateTimeFromLessonScopeLocal(int(startLessonScope), int(endLessonScope))
		weekStr, _ := courseMap["week"].(string)
		weekday := convertWeekStringToNumberLocal(weekStr)
		nowWeekFloat, _ := courseMap["nowweek"].(float64)
		nowWeek := int(nowWeekFloat)
		if nowWeek == 0 {
			nowWeekFloat2, _ := courseMap["nowWeek"].(float64)
			nowWeek = int(nowWeekFloat2)
		}
		if nowWeek == 0 {
			nowWeek = week
		}

		courses = append(courses, models.Course{
			Name:             courseName,
			Code:             fmt.Sprintf("COURSE_%d", time.Now().UnixNano()),
			Credit:           0,
			TeacherName:      teacherNames,
			Classroom:        classroom,
			Weekday:          weekday,
			StartTime:        startTime,
			EndTime:          endTime,
			StartLessonScope: int(startLessonScope),
			EndLessonScope:   int(endLessonScope),
			Week:             nowWeek,
			StartWeek:        1,
			EndWeek:          16,
			SemesterID:       semester.ID,
			UserID:           user.ID,
		})
	}
	return courses, nil
}

func calculateTimeFromLessonScopeLocal(startLesson, endLesson int) (string, string) {
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

func convertWeekStringToNumberLocal(weekStr string) int {
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
	return 1
}

