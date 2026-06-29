// Package controllers handles HTTP request handlers for proxy-related endpoints.
// This file contains the ProxyController struct and its HTTP handler methods.
package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// ProxyController 代理控制器
type ProxyController struct {
	proxyClient *utils.ProxyClient
}

// NewProxyController 创建新的代理控制器
func NewProxyController() *ProxyController {
	return &ProxyController{
		proxyClient: utils.NewProxyClient(),
	}
}

// GetSchool 获取学校信息
func (pc *ProxyController) GetSchool(c *gin.Context) {
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
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}
	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/sys/user/getSchool", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学校信息失败: "+err.Error(), 500))
		return
	}
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetCurrentTime 获取当前时间信息（缓存优先逻辑）
func (pc *ProxyController) GetCurrentTime(c *gin.Context) {
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
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	offlineCache := services.GetOfflineCacheService()
	cached, found, cacheErr := offlineCache.GetCurrentTimeCache()
	if cacheErr == nil && found && !cached.Meta.ExpiresAt.IsZero() && time.Now().Before(cached.Meta.ExpiresAt) {
		var resp map[string]interface{}
		if json.Unmarshal(cached.Data, &resp) == nil {
			resp["fromCache"] = true
			resp["cacheUpdatedAt"] = cached.Meta.FetchedAt.Format("2006-01-02 15:04:05")
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime", nil)
	if err != nil {
		if found {
			var staleResp map[string]interface{}
			if json.Unmarshal(cached.Data, &staleResp) == nil {
				staleResp["fromCache"] = true
				staleResp["cacheUpdatedAt"] = cached.Meta.FetchedAt.Format("2006-01-02 15:04:05")
				staleResp["stale"] = true
				c.JSON(http.StatusOK, staleResp)
				return
			}
		}
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取当前时间失败: "+err.Error(), 500))
		return
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
		return
	}

	_ = offlineCache.PutCurrentTimeCache(body, 30*time.Minute)

	c.JSON(http.StatusOK, resp)
}

// GetCourseTimeTableByDay 获取当日课程表
func (pc *ProxyController) GetCourseTimeTableByDay(c *gin.Context) {
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
	syncSetting, err := models.GetSyncSettingByUserID(uint(userIdUint))
	if err != nil {
		syncSetting = &models.SyncSetting{Enabled: false}
	}
	if syncSetting.Enabled {
		dateStr := time.Now().Format("2006-01-02")
		if queryDate := c.Query("date"); queryDate != "" {
			dateStr = queryDate
		}
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			weekday := int(targetDate.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			semester, err := models.GetCurrentSemester()
			if err == nil {
				daysSinceStart := int(targetDate.Sub(semester.StartDate).Hours() / 24)
				week := daysSinceStart/7 + 1
				courses, err := models.GetCoursesByUserAndDay(uint(userIdUint), semester.ID, weekday, week)
				if err == nil && len(courses) > 0 {
					records := make([]map[string]interface{}, 0)
					for _, course := range courses {
						records = append(records, map[string]interface{}{
							"id": course.ID, "courseName": course.Name, "courseCode": course.Code,
							"credit": course.Credit, "teacherName": course.TeacherName, "classroom": course.Classroom,
							"weekDay": course.Weekday, "startTime": course.StartTime, "endTime": course.EndTime,
							"startWeek": course.StartWeek, "endWeek": course.EndWeek, "date": dateStr, "week": week,
						})
					}
					c.JSON(http.StatusOK, map[string]interface{}{
						"code": 0, "message": "", "success": true,
						"timestamp": time.Now().UnixNano() / 1000000,
						"result": map[string]interface{}{
							"pages": 1, "records": records, "searchCount": true,
							"size": 84, "total": len(records),
						},
					})
					return
				}
			}
		}
	}
	params := url.Values{}
	params.Set("current", c.DefaultQuery("current", "1"))
	params.Set("size", c.DefaultQuery("size", "-1"))
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByDay", params)
}

// GetCourseTimeTableByWeek 获取本周课程表（缓存优先逻辑）
func (pc *ProxyController) GetCourseTimeTableByWeek(c *gin.Context) {
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
	currentSemester := c.Query("currentSemester")
	nowWeekStr := c.Query("nowWeek")
	current := c.DefaultQuery("current", "1")
	size := c.DefaultQuery("size", "84")
	nowWeek := 1
	if nowWeekStr != "" {
		if parsed, err := strconv.Atoi(nowWeekStr); err == nil {
			nowWeek = parsed
		}
	}
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败", 500))
		return
	}
	semester, err := models.GetCurrentSemester()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取学期信息失败", 500))
		return
	}
	syncSetting, err := models.GetSyncSettingByUserID(uint(userIdUint))
	if err != nil {
		syncSetting = &models.SyncSetting{Enabled: false}
	}
	offlineCache := services.GetOfflineCacheService()

	// 1. 如果开启了用户同步，优先使用数据库
	var dataSourceType DataSourceType
	var cacheUpdatedAt string
	var courses []models.Course
	useUserSyncData := false
	if syncSetting.Enabled {
		courses, err = models.GetCoursesByUserAndWeek(uint(userIdUint), semester.ID, nowWeek)
		if err == nil && len(courses) > 0 {
			dataSourceType = DataSourceDatabase
			useUserSyncData = true
			if syncSetting.LastSyncAt != nil {
				cacheUpdatedAt = syncSetting.LastSyncAt.Format("2006-01-02 15:04:05")
			} else {
				cacheUpdatedAt = time.Now().Format("2006-01-02 15:04:05")
			}
		}
	}

	// 2. 缓存命中（缓存优先于学校服务器调用）
	if len(courses) == 0 {
		cached, found, cacheErr := offlineCache.GetCourseTimetableWeekCache(uint(userIdUint), currentSemester, nowWeek)
		if cacheErr == nil && found {
			cacheFresh := !cached.Meta.ExpiresAt.IsZero() && time.Now().Before(cached.Meta.ExpiresAt)

			var cachedResp map[string]interface{}
			if json.Unmarshal(cached.Data, &cachedResp) == nil {
				if cacheFresh {
					cachedResp["fromCache"] = true
					cachedResp["cacheUpdatedAt"] = cached.Meta.FetchedAt.Format("2006-01-02 15:04:05")
					c.JSON(http.StatusOK, cachedResp)
					return
				}
				// 缓存过期，返回 stale 数据
				cachedResp["fromCache"] = true
				cachedResp["stale"] = true
				cachedResp["cacheUpdatedAt"] = cached.Meta.FetchedAt.Format("2006-01-02 15:04:05")
				c.JSON(http.StatusOK, cachedResp)
				return
			}
		}

		// 3. 缓存未命中，从学校服务器获取
		dataSourceType = DataSourceSchool
		cacheUpdatedAt = time.Now().Format("2006-01-02 15:04:05")
		requestURL := fmt.Sprintf("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek?current=%s&size=%s&currentSemester=%s&nowWeek=%d",
			current, size, url.QueryEscape(currentSemester), nowWeek)
		proxyClient := utils.NewProxyClient()
		timetableResp, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", requestURL, nil)
		if err != nil {
			// 学校服务器失败，尝试数据库兜底
			courses, _ = models.GetCoursesByUserAndWeek(uint(userIdUint), semester.ID, nowWeek)
			if len(courses) > 0 {
				dataSourceType = DataSourceDatabase
				useUserSyncData = true
				cacheUpdatedAt = "服务器维护中，使用缓存数据"
			}
		} else {
			courses, err = parseSchoolTimetableResponse(timetableResp, uint(userIdUint), semester.ID, nowWeek)
			if err != nil {
				courses, _ = models.GetCoursesByUserAndWeek(uint(userIdUint), semester.ID, nowWeek)
				if len(courses) > 0 {
					dataSourceType = DataSourceDatabase
					useUserSyncData = true
					cacheUpdatedAt = "解析失败，使用缓存数据"
				}
			} else {
				// 成功，缓存学校接口响应（15分钟）
				ttl := 15 * time.Minute
				_ = offlineCache.PutCourseTimetableWeekCache(uint(userIdUint), currentSemester, nowWeek, timetableResp, ttl)
			}
		}
	}

	courseRecords := make([]gin.H, 0, len(courses))
	for _, course := range courses {
		var startLessonScope, endLessonScope int
		if course.StartLessonScope > 0 && course.EndLessonScope > 0 {
			startLessonScope = course.StartLessonScope
			endLessonScope = course.EndLessonScope
		} else {
			startLessonScope = config.GetLessonFromTime(course.StartTime)
			endLessonScope = config.GetLessonFromTime(course.EndTime)
		}
		lessonScopeLenght := endLessonScope - startLessonScope + 1
		var lessonScope string
		if lessonScopeLenght == 1 {
			lessonScope = fmt.Sprintf("#%d#", startLessonScope)
		} else {
			var scopeParts []string
			for i := startLessonScope; i <= endLessonScope; i++ {
				scopeParts = append(scopeParts, fmt.Sprintf("#%d#", i))
			}
			lessonScope = strings.Join(scopeParts, ",")
		}
		weekMap := map[int]string{1: "一", 2: "二", 3: "三", 4: "四", 5: "五", 6: "六", 7: "日"}
		weekStr := weekMap[course.Weekday]
		if weekStr == "" {
			weekStr = "一"
		}
		courseRecords = append(courseRecords, gin.H{
			"id": course.ID, "courseName": course.Name, "teacherNames": course.TeacherName,
			"classroomName": course.Classroom, "startTime": course.StartTime, "endTime": course.EndTime,
			"startLessonScope": startLessonScope, "endLessonScope": endLessonScope,
			"lessonScope": lessonScope, "lessonScopeLenght": lessonScopeLenght,
			"week": weekStr, "weekday": course.Weekday, "nowweek": nowWeek, "nowWeek": nowWeek,
			"shortDate": time.Now().Format("2006-01-02"),
			"semester":  semester.Name, "currentSemester": currentSemester,
		})
	}
	response := gin.H{
		"success": true, "code": 200, "message": "获取成功",
		"result":         gin.H{"current": 1, "pages": 1, "records": courseRecords, "size": len(courseRecords), "total": len(courseRecords)},
		"dataSourceType": dataSourceType, "cacheUpdatedAt": cacheUpdatedAt,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	if useUserSyncData {
		response["dataSourceNote"] = "用户同步数据"
	} else {
		switch dataSourceType {
		case DataSourceSchool:
			response["dataSourceNote"] = "学校服务器实时数据"
		default:
			response["dataSourceNote"] = "智能缓存数据"
		}
	}

	// 如果没有可用数据，将数据库结果也缓存起来
	if len(courses) > 0 && dataSourceType == DataSourceDatabase && !useUserSyncData {
		cacheResp, _ := json.Marshal(response)
		_ = offlineCache.PutCourseTimetableWeekCache(uint(userIdUint), currentSemester, nowWeek, cacheResp, 1*time.Hour)
	}

	c.JSON(http.StatusOK, response)
}

// GetEoaNewsTypeList 获取新闻类型列表
func (pc *ProxyController) GetEoaNewsTypeList(c *gin.Context) {
	params := url.Values{}
	params.Set("pageNo", c.DefaultQuery("pageNo", "1"))
	params.Set("pageSize", c.DefaultQuery("pageSize", "-1"))
	cacheKey := "eoa_news_type_list"
	offlineCache := services.GetOfflineCacheService()
	fallback := func() interface{} {
		return map[string]interface{}{
			"code": 0, "message": "", "success": true,
			"result": []map[string]interface{}{
				{"id": 18, "name": "通知公告"},
				{"id": 19, "name": "学校新闻"},
			},
		}
	}
	res, err := offlineCache.GetOrRefreshNews(cacheKey, 30*time.Minute, func() ([]byte, error) {
		userIdInterface, exists := c.Get("userId")
		if !exists {
			return nil, fmt.Errorf("unauthorized")
		}
		userIdStr, ok := userIdInterface.(string)
		if !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		user, err := models.FindUserByID(uint(userIdUint))
		if err != nil {
			return nil, err
		}
		return pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList", params)
	})
	if err != nil {
		c.JSON(http.StatusOK, fallback())
		return
	}
	var response map[string]interface{}
	if err := json.Unmarshal(res.Data, &response); err != nil {
		c.JSON(http.StatusOK, fallback())
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetEoaNewsListByTypeId 获取新闻列表
func (pc *ProxyController) GetEoaNewsListByTypeId(c *gin.Context) {
	params := url.Values{}
	params.Set("pageNo", c.DefaultQuery("pageNo", "1"))
	params.Set("pageSize", c.DefaultQuery("pageSize", "6"))
	params.Set("ids", c.Query("ids"))
	cacheKey := fmt.Sprintf("eoa_news_list:%s:%s:%s", params.Get("ids"), params.Get("pageNo"), params.Get("pageSize"))
	offlineCache := services.GetOfflineCacheService()
	fallback := func() interface{} {
		return map[string]interface{}{
			"code": 0, "message": "", "success": true,
			"result": map[string]interface{}{
				"records": []map[string]interface{}{
					{"id": 1, "title": "学校通知公告", "content": "这是一条示例通知公告内容。", "summary": "示例通知公告摘要", "publishTime": "2024-12-20 10:00:00", "author": "学校办公室", "viewCount": 100},
					{"id": 2, "title": "重要通知", "content": "这是另一条重要通知内容。", "summary": "重要通知摘要", "publishTime": "2024-12-19 15:30:00", "author": "教务处", "viewCount": 85},
				},
				"total": 2, "current": 1, "size": 6,
			},
		}
	}
	res, err := offlineCache.GetOrRefreshNews(cacheKey, 30*time.Minute, func() ([]byte, error) {
		userIdInterface, exists := c.Get("userId")
		if !exists {
			return nil, fmt.Errorf("unauthorized")
		}
		userIdStr, ok := userIdInterface.(string)
		if !ok {
			return nil, fmt.Errorf("unauthorized")
		}
		userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		user, err := models.FindUserByID(uint(userIdUint))
		if err != nil {
			return nil, err
		}
		return pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId", params)
	})
	if err != nil {
		c.JSON(http.StatusOK, fallback())
		return
	}
	var response map[string]interface{}
	if err := json.Unmarshal(res.Data, &response); err != nil {
		c.JSON(http.StatusOK, fallback())
		return
	}
	c.JSON(http.StatusOK, response)
}

// GetLearningData 获取学习数据
func (pc *ProxyController) GetLearningData(c *gin.Context) {
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseScore/getLearningData", nil)
}

// GetSemester 获取学期信息
func (pc *ProxyController) GetSemester(c *gin.Context) {
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseScore/getSemester", nil)
}

// GetScoreCurrentTime 获取成绩当前时间信息
func (pc *ProxyController) GetScoreCurrentTime(c *gin.Context) {
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseScore/getCurrentTime", nil)
}

// GetCoursePlan 获取课程计划
func (pc *ProxyController) GetCoursePlan(c *gin.Context) {
	params := url.Values{}
	params.Set("current", c.DefaultQuery("current", "1"))
	params.Set("size", c.DefaultQuery("size", "15"))
	params.Set("currentSemester", c.Query("currentSemester"))
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseScore/getCoursePlan", params)
}

// GetCourseScore 获取课程成绩
func (pc *ProxyController) GetCourseScore(c *gin.Context) {
	params := url.Values{}
	params.Set("current", c.DefaultQuery("current", "1"))
	params.Set("size", c.DefaultQuery("size", "15"))
	params.Set("currentSemester", c.Query("currentSemester"))
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseScore/getCourseScore", params)
}

// GetScoreList 获取成绩列表（缓存优先逻辑）
func (pc *ProxyController) GetScoreList(c *gin.Context) {
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
	current := c.DefaultQuery("current", "1")
	size := c.DefaultQuery("size", "50")
	semesterName := c.Query("semesterName")
	if semesterName == "" {
		semesterName = c.Query("currentSemester")
	}
	if semesterName == "" {
		semesterName = "全部"
	}
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败", 500))
		return
	}
	offlineCache := services.GetOfflineCacheService()
	bundle, meta, found, cacheErr := offlineCache.GetPersonalScoreBundle(user.ID, semesterName)
	if cacheErr == nil && found && len(bundle.ScoreList) > 0 && !meta.ExpiresAt.IsZero() && time.Now().Before(meta.ExpiresAt) {
		cachedResp, err := standardizeAPIResponse(bundle.ScoreList, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if err == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	params := url.Values{}
	params.Set("current", current)
	params.Set("size", size)
	params.Set("currentSemester", semesterName)
	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseScore/getCourseScore", params)
	if err == nil {
		resp, serr := standardizeAPIResponse(body, DataSourceSchool, time.Now().Format("2006-01-02 15:04:05"))
		if serr != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
			return
		}
		if resp.Code >= 500 {
			//nolint:errcheck
			//nolint:errcheck
			err = fmt.Errorf("upstream code %d", resp.Code)
		} else {
			_, _ = offlineCache.UpsertPersonalScorePart(user.ID, semesterName, "scoreList", body, 12*time.Hour)
			c.JSON(http.StatusOK, resp)
			return
		}
		//nolint:ineffassign
	}
	if cacheErr == nil && found && len(bundle.ScoreList) > 0 {
		cachedResp, serr := standardizeAPIResponse(bundle.ScoreList, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if serr == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器不可用且无可用缓存", 503))
}

// GetSemesterScore 获取学期统计成绩（缓存优先逻辑）
func (pc *ProxyController) GetSemesterScore(c *gin.Context) {
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
	semesterName := c.Query("semesterName")
	if semesterName == "" {
		semesterName = c.Query("currentSemester")
	}
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败", 500))
		return
	}
	if semesterName == "" {
		semesterName = user.CurrentSemester
	}
	offlineCache := services.GetOfflineCacheService()
	bundle, meta, found, cacheErr := offlineCache.GetPersonalScoreBundle(user.ID, semesterName)
	if cacheErr == nil && found && len(bundle.SemesterScore) > 0 && !meta.ExpiresAt.IsZero() && time.Now().Before(meta.ExpiresAt) {
		cachedResp, err := standardizeAPIResponse(bundle.SemesterScore, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if err == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	params := url.Values{}
	if semesterName != "" {
		params.Set("semesterName", semesterName)
		params.Set("currentSemester", semesterName)
	}
	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/scs/course/tCourseScore/getSemesterScore", params)
	if err == nil {
		resp, serr := standardizeAPIResponse(body, DataSourceSchool, time.Now().Format("2006-01-02 15:04:05"))
		if serr != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
			return
		}
		//nolint:errcheck
		//nolint:errcheck
		if resp.Code >= 500 {
			err = fmt.Errorf("upstream code %d", resp.Code)
		} else {
			_, _ = offlineCache.UpsertPersonalScorePart(user.ID, semesterName, "semesterScore", body, 12*time.Hour)
			c.JSON(http.StatusOK, resp)
			return
		}
		//nolint:ineffassign
	}
	if cacheErr == nil && found && len(bundle.SemesterScore) > 0 {
		cachedResp, serr := standardizeAPIResponse(bundle.SemesterScore, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if serr == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	c.JSON(http.StatusOK, &StandardResponse{
		Success:        true,
		Code:           200,
		Message:        "学校服务器不可用，暂无统计数据",
		Result:         map[string]interface{}{"gpa": "0.00", "averageScore": "0.00", "creditTotal": "0", "passRate": "0"},
		DataSourceType: DataSourceDatabase,
		CacheUpdatedAt: "",
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	})
}

// GetCourseLessonTime 获取课表时间配置
func (pc *ProxyController) GetCourseLessonTime(c *gin.Context) {
	proxyRequest(pc, c, "GET", "/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime", nil)
}

// GetCourseSchoolTimetable 获取学校课表信息
func (pc *ProxyController) GetCourseSchoolTimetable(c *gin.Context) {
	proxyRequest(pc, c, "GET", "/scloudoa/userQuery/tSysUser/getCourseSchoolTimetable", nil)
}

// GetEvaluationStudentConfigList 获取评教配置
func (pc *ProxyController) GetEvaluationStudentConfigList(c *gin.Context) {
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("获取用户信息失败", 500))
		return
	}
	term := user.CurrentSemester
	if term == "" {
		term = "__default__"
	}
	offlineCache := services.GetOfflineCacheService()
	bundle, meta, found, cacheErr := offlineCache.GetPersonalScoreBundle(user.ID, term)
	if cacheErr == nil && found && len(bundle.EvaluationConfigList) > 0 && !meta.ExpiresAt.IsZero() && time.Now().Before(meta.ExpiresAt) {
		cachedResp, err := standardizeAPIResponse(bundle.EvaluationConfigList, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if err == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	body, err := pc.proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", nil)
	if err == nil {
		resp, serr := standardizeAPIResponse(body, DataSourceSchool, time.Now().Format("2006-01-02 15:04:05"))
		if serr != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析响应失败", 500))
			return
			//nolint:errcheck
			//nolint:errcheck
		}
		if resp.Code >= 500 {
			err = fmt.Errorf("upstream code %d", resp.Code)
		} else {
			_, _ = offlineCache.UpsertPersonalScorePart(user.ID, term, "evaluationConfigList", body, 12*time.Hour)
			c.JSON(http.StatusOK, resp)
			return
		}
		//nolint:ineffassign
	}
	if cacheErr == nil && found && len(bundle.EvaluationConfigList) > 0 {
		cachedResp, serr := standardizeAPIResponse(bundle.EvaluationConfigList, DataSourceDatabase, meta.FetchedAt.Format("2006-01-02 15:04:05"))
		if serr == nil {
			c.JSON(http.StatusOK, cachedResp)
			return
		}
	}
	c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器不可用且无可用缓存", 503))
}

// DownloadAttachment 下载附件
func (pc *ProxyController) DownloadAttachment(c *gin.Context) {
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
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}
	attachmentUrl := c.Query("attachmentUrl")
	fileName := c.Query("fileName")
	if fileName == "" {
		fileName = "attachment"
	}
	fileData, contentType, err := pc.proxyClient.DownloadFileWithAuth(user, attachmentUrl, fileName)
	if err != nil {
		// #region agent log
		log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=BACKEND-ERR location=proxy_controller.go:DownloadAttachment:fail message=Download failed, will return JSON error error=%s", err.Error())
		// #endregion
		if strings.Contains(err.Error(), "文件服务器认证失败") || strings.Contains(err.Error(), "未登陆") {
			fileURL, getURLErr := pc.proxyClient.GetFileURLViaGetFileList(user, attachmentUrl, fileName)
			if getURLErr == nil && fileURL != "" {
				c.JSON(http.StatusOK, map[string]interface{}{
					"success": false, "message": "文件服务器需要特殊认证，请点击链接手动下载", "code": 200,
					"result": map[string]interface{}{"downloadUrl": fileURL, "fileName": fileName, "needManualDownload": true},
				})
				return
			}
			constructedURL := fmt.Sprintf("http://scs.whggvc.net/scscloud/%s", attachmentUrl)
			c.JSON(http.StatusOK, map[string]interface{}{
				"success": false, "message": "文件服务器需要特殊认证，请点击链接手动下载", "code": 200,
				"result": map[string]interface{}{"downloadUrl": constructedURL, "fileName": fileName, "needManualDownload": true},
			})
			return
		}
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("附件下载失败: "+err.Error(), 404))
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.Itoa(len(fileData)))
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Access-Token")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Type, Content-Length")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, contentType, fileData)
}

// FileProxy 文件代理（用于 img src 等直接加载场景）
func (pc *ProxyController) FileProxy(c *gin.Context) {
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
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// path 格式: /group1/M00/00/4D/wKjJiWlBF9GAWPDDAAmaysfVODg529.jpg
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("文件路径不能为空", 400))
		return
	}

	// 自动补全斜杠（gin 的 * 通配符不会捕获前导斜杠）
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	fileData, contentType, err := pc.proxyClient.DownloadFileWithAuth(user, filePath, filepath.Base(filePath))
	if err != nil {
		if strings.Contains(err.Error(), "文件服务器认证失败") || strings.Contains(err.Error(), "未登陆") {
			c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("文件服务器认证失败，请重新登录", 401))
			return
		}
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("文件加载失败: "+err.Error(), 404))
		return
	}

	// 设置缓存控制头，让浏览器缓存图片
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.Itoa(len(fileData)))
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Access-Token")
	c.Header("Access-Control-Expose-Headers", "Content-Type, Content-Length")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, contentType, fileData)
}
