package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// SyncService 同步服务
type SyncService struct {
	ticker      *time.Ticker
	stopChan    chan bool
	running     bool
	userSyncMu  sync.Mutex    // B4 fix: 保护并发同步 map
	userSyncing map[uint]bool // 记录正在同步的用户，防止同一用户并发
}

// NewSyncService 创建新的同步服务
func NewSyncService() *SyncService {
	return &SyncService{
		stopChan:    make(chan bool),
		running:     false,
		userSyncing: make(map[uint]bool),
	}
}

// Start 启动同步服务
func (s *SyncService) Start() {
	if s.running {
		log.Println("同步服务已经在运行中")
		return
	}

	s.running = true
	s.ticker = time.NewTicker(5 * time.Minute) // 每5分钟检查一次

	log.Println("同步服务已启动，每5分钟检查一次同步任务")

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkAndExecuteSync()
			case <-s.stopChan:
				log.Println("同步服务已停止")
				return
			}
		}
	}()
}

// Stop 停止同步服务
func (s *SyncService) Stop() {
	if !s.running {
		return
	}

	s.running = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.stopChan <- true
}

// checkAndExecuteSync 检查并执行同步任务
func (s *SyncService) checkAndExecuteSync() {
	// 1. 课程表同步
	s.checkAndExecuteCourseSync()

	// 2. 个人基础信息缓存（遵守2天不活跃暂停规则）
	s.checkAndExecutePersonalInfoSync()
}

// checkAndExecuteCourseSync 检查并执行课程表同步
func (s *SyncService) checkAndExecuteCourseSync() {
	settings, err := models.GetSyncSettingsForSchedule()
	if err != nil {
		log.Printf("获取同步设置失败: %v", err)
		return
	}

	for _, setting := range settings {
		// 检查是否在允许的时间范围内
		if !s.isInTimeRange(setting.TimeRange) {
			continue
		}

		// B4 fix: 用唯一的 goroutine map 协调，同一用户不会并发同步
		s.userSyncMu.Lock()
		if s.userSyncing[setting.UserID] {
			s.userSyncMu.Unlock()
			log.Printf("用户 %d 已在同步中，跳过", setting.UserID)
			continue
		}
		s.userSyncing[setting.UserID] = true
		s.userSyncMu.Unlock()

		// 执行同步
		go func(userID uint) {
			s.executeSyncForUser(userID)
			s.userSyncMu.Lock()
			delete(s.userSyncing, userID)
			s.userSyncMu.Unlock()
		}(setting.UserID)
	}
}

// checkAndExecutePersonalInfoSync 检查并执行个人基础信息缓存任务
// 规则：用户超过2天不活跃则暂停缓存，下次用户登录时自动恢复
func (s *SyncService) checkAndExecutePersonalInfoSync() {
	settings, err := GetSettingsWithPersonalInfoSyncEnabled()
	if err != nil {
		log.Printf("获取个人信息缓存设置失败: %v", err)
		return
	}

	for _, setting := range settings {
		// 检查时间范围
		if !s.isInTimeRange(setting.TimeRange) {
			continue
		}

		// 活跃度检查：超过2天未活跃则暂停缓存
		isActive := models.IsUserActiveWithinDays(setting.UserID, 2)
		wasPaused := setting.PersonalInfoCacheStatus == "paused"

		if !isActive {
			// 用户不活跃，暂停缓存
			if setting.PersonalInfoCacheStatus != "paused" {
				setting.PersonalInfoCacheStatus = "paused"
				//nolint:errcheck
				models.UpdateSyncSetting(&setting)
				log.Printf("用户 %d 超过2天未活跃，暂停个人基础信息缓存", setting.UserID)
			}
			continue
		}

		// 用户重新活跃：自动恢复缓存
		if wasPaused {
			setting.PersonalInfoCacheStatus = "resumed"
			//nolint:errcheck
			models.UpdateSyncSetting(&setting)
			log.Printf("用户 %d 恢复活跃，自动重新启用个人基础信息缓存", setting.UserID)
		} else if setting.PersonalInfoCacheStatus == "" {
			setting.PersonalInfoCacheStatus = "active"
		}

		// 检查缓存是否已过期（6小时缓存间隔）
		needsRefresh := setting.PersonalInfoLastCachedAt == nil ||
			time.Since(*setting.PersonalInfoLastCachedAt) >= 6*time.Hour

		if !needsRefresh {
			continue
		}

		// 获取用户信息
		user, err := models.FindUserByID(setting.UserID)
		if err != nil || user.Username == "" || user.Password == "" {
			continue
		}

		// 执行个人基础信息缓存（异步）
		go s.executePersonalInfoCache(setting.UserID)
	}
}

// ResumePersonalInfoCacheIfNeeded 检查并恢复用户个人基础信息缓存
// 当用户登录时调用，如果用户之前因不活跃被暂停缓存，则自动恢复
func (s *SyncService) ResumePersonalInfoCacheIfNeeded(userID uint) {
	setting, err := models.GetSyncSettingByUserID(userID)
	if err != nil || !setting.PersonalInfoSyncEnabled {
		return
	}

	// 只有处于暂停状态才需要恢复
	if setting.PersonalInfoCacheStatus != "paused" {
		return
	}

	// 恢复缓存
	setting.PersonalInfoCacheStatus = "resumed"
	if err := models.UpdateSyncSetting(setting); err != nil {
		log.Printf("恢复用户 %d 个人基础信息缓存失败: %v", userID, err)
		return
	}

	log.Printf("用户 %d 登录，恢复个人基础信息缓存", userID)

	// 立即触发一次缓存（异步）
	go s.executePersonalInfoCache(userID)
}

// executePersonalInfoCache 执行个人基础信息缓存
func (s *SyncService) executePersonalInfoCache(userID uint) {
	startTime := time.Now()

	setting, err := models.GetSyncSettingByUserID(userID)
	if err != nil {
		return
	}

	// 双重检查：确保用户活跃
	if !models.IsUserActiveWithinDays(userID, 2) {
		if setting.PersonalInfoCacheStatus != "paused" {
			setting.PersonalInfoCacheStatus = "paused"
			//nolint:errcheck
			models.UpdateSyncSetting(setting)
		}
		return
	}

	// 获取并缓存个人信息
	archive, err := fetchAndCachePersonalInfo(userID)
	duration := time.Since(startTime)

	if err != nil {
		s.logPersonalInfoSyncResult(userID, "personal_info", "failed",
			fmt.Sprintf("个人基础信息缓存失败: %v", err), 0, duration)
		log.Printf("用户 %d 个人基础信息缓存失败: %v", userID, err)
		return
	}

	// 更新缓存时间
	now := time.Now()
	setting.PersonalInfoLastCachedAt = &now
	setting.PersonalInfoCacheStatus = "active"
	//nolint:errcheck
	models.UpdateSyncSetting(setting)

	s.logPersonalInfoSyncResult(userID, "personal_info", "success",
		fmt.Sprintf("个人基础信息缓存成功 (%s)", archive.Realname), 0, duration)
	log.Printf("用户 %d 个人基础信息缓存成功: %s, 耗时 %v", userID, archive.Realname, duration)
}

// notifySchoolPasswordMissing 当检测到用户学校密码未设置时，发送微信/站内通知
// B5 fix: 永久错误（学校密码未设置）应通知用户重新登录设置密码，而非静默失败
func (s *SyncService) notifySchoolPasswordMissing(user *models.User) {
	log.Printf("[SyncService] 通知用户 %d 学校密码未设置，请重新登录设置", user.ID)

	// 1. 创建系统通知存入数据库
	notification := &models.Notification{
		UserID:  user.ID,
		Type:    "system",
		Title:   "课程同步已暂停",
		Content: "您的学校账号密码未设置，自动课程同步已暂停。请退出登录后重新登录，输入学校密码后同步将自动恢复。",
		IsRead:  false,
	}
	if err := models.DB.Create(notification).Error; err != nil {
		log.Printf("[SyncService] 创建系统通知失败: %v", err)
	}

	// 2. 通过多渠道通知服务发送（邮件/钉钉）
	mcns := GetMultiChannelNotificationService()
	if mcns != nil {
		channel, err := models.GetNotificationChannelByUserID(user.ID)
		if err != nil {
			log.Printf("[SyncService] 获取用户 %d 通知配置失败: %v", user.ID, err)
			return
		}

		// 发送邮件通知
		if channel.EmailEnabled && channel.EmailAddress != "" {
			if err := mcns.sendSchoolPasswordMissingEmail(channel, user); err != nil {
				log.Printf("[SyncService] 发送学校密码缺失邮件通知失败: %v", err)
			}
		}

		// 发送钉钉通知
		if channel.DingTalkEnabled && channel.DingTalkWebhookURL != "" {
			if err := mcns.sendSchoolPasswordMissingDingTalk(channel, user); err != nil {
				log.Printf("[SyncService] 发送学校密码缺失钉钉通知失败: %v", err)
			}
		}
	}
}

// fetchAndCachePersonalInfo 获取并缓存个人基础信息
func fetchAndCachePersonalInfo(userID uint) (*PersonalInfoCache, error) {
	user, err := models.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	if user.Username == "" || user.Password == "" {
		return nil, fmt.Errorf("用户未绑定学校账号")
	}

	// 从学校服务器直接获取最新数据
	archive, err := fetchPersonalInfoFromSchoolServer(user)
	if err != nil {
		return nil, err
	}

	// 更新本地用户表的基础字段（姓名、班级等变更时同步）
	updates := make(map[string]interface{})
	if archive.Realname != "" && archive.Realname != user.Realname {
		updates["realname"] = archive.Realname
	}
	if archive.ClassName != "" && archive.ClassName != user.ClassName {
		updates["class_name"] = archive.ClassName
	}
	if len(updates) > 0 {
		if err := models.UpdateUserFields(user.ID, updates); err != nil {
			log.Printf("更新用户 %d 基础信息失败: %v", user.ID, err)
		}
	}

	return archive, nil
}

// fetchPersonalInfoFromSchoolServer 从学校服务器直接获取个人信息
func fetchPersonalInfoFromSchoolServer(user *models.User) (*PersonalInfoCache, error) {
	proxyClient := utils.NewProxyClient()

	body, err := proxyClient.ProxyRequestWithAutoLogin(user, "GET", "/scloud/student/base/getStudentInfo", nil)
	if err != nil {
		return nil, fmt.Errorf("调用学校学生信息API失败: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("解析学生信息响应失败: %w", err)
	}

	if success, ok := data["success"].(bool); !ok || !success {
		msg, _ := data["message"].(string)
		return nil, fmt.Errorf("学校服务器错误: %s", msg)
	}

	result, ok := data["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("学生信息响应格式错误")
	}

	info := &PersonalInfoCache{
		StudentNo:     getStringField(result, "studyNumber"),
		Realname:      getStringField(result, "name"),
		ClassName:     getStringField(result, "className"),
		AdminClass:    getStringField(result, "adminClass"),
		FacultyName:   getStringField(result, "facultyStation"),
		MajorName:     getStringField(result, "professionName"),
		Grade:         getStringField(result, "grade"),
		EntranceDate:  getStringField(result, "entranceDate"),
		StudyForm:     getStringField(result, "studyForm"),
		StudentStatus: getStringField(result, "enrollmentStatus"),
	}

	return info, nil
}

// PersonalInfoCache 个人基础信息缓存结构
type PersonalInfoCache struct {
	StudentNo     string `json:"studentNo"`
	Realname      string `json:"realname"`
	ClassName     string `json:"className"`
	AdminClass    string `json:"adminClass"`
	FacultyName   string `json:"facultyName"`
	MajorName     string `json:"majorName"`
	Grade         string `json:"grade"`
	EntranceDate  string `json:"entranceDate"`
	StudyForm     string `json:"studyForm"`
	StudentStatus string `json:"studentStatus"`
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// logPersonalInfoSyncResult 记录个人信息缓存结果到 sync_log
func (s *SyncService) logPersonalInfoSyncResult(userID uint, syncCategory, status, message string, coursesSync int, duration time.Duration) {
	now := time.Now()
	syncLog := &models.SyncLog{
		UserID:      userID,
		SyncType:    syncCategory,
		Status:      status,
		Message:     message,
		CoursesSync: coursesSync,
		Duration:    int(duration.Milliseconds()),
		SyncedAt:    &now,
	}
	if err := models.CreateSyncLog(syncLog); err != nil {
		log.Printf("创建个人信息缓存日志失败: %v", err)
	}
}

// GetSettingsWithPersonalInfoSyncEnabled 获取所有开启了个人信息缓存的用户设置
func GetSettingsWithPersonalInfoSyncEnabled() ([]models.SyncSetting, error) {
	var settings []models.SyncSetting
	err := models.DB.Where("personal_info_sync_enabled = ?", true).Find(&settings).Error
	return settings, err
}

// isInTimeRange 检查当前时间是否在允许的时间范围内
func (s *SyncService) isInTimeRange(timeRange string) bool {
	if timeRange == "" {
		timeRange = "08:30-22:20"
	}

	now := time.Now()
	currentMinutes := now.Hour()*60 + now.Minute()

	serverStartMinutes := 8*60 + 30 // 08:30
	serverEndMinutes := 22*60 + 20  // 22:20

	if currentMinutes < serverStartMinutes || currentMinutes > serverEndMinutes {
		log.Printf("当前时间 %02d:%02d 不在学校服务器开放时间内 (08:30-22:20)", now.Hour(), now.Minute())
		return false
	}

	parts := strings.Split(timeRange, "-")
	if len(parts) != 2 {
		log.Printf("时间范围格式错误: %s", timeRange)
		return false
	}

	startTime := strings.TrimSpace(parts[0])
	endTime := strings.TrimSpace(parts[1])

	startMinutes, err := s.parseTimeToMinutes(startTime)
	if err != nil {
		log.Printf("解析开始时间失败: %s, %v", startTime, err)
		return false
	}

	endMinutes, err := s.parseTimeToMinutes(endTime)
	if err != nil {
		log.Printf("解析结束时间失败: %s, %v", endTime, err)
		return false
	}

	if startMinutes < serverStartMinutes || startMinutes > serverEndMinutes ||
		endMinutes < serverStartMinutes || endMinutes > serverEndMinutes {
		log.Printf("用户设置的时间范围 %s 超出学校服务器开放时间 (08:30-22:20)", timeRange)
		return false
	}

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

// parseTimeToMinutes 将时间字符串（如 "08:30"）转换为分钟数
func (s *SyncService) parseTimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("时间格式错误: %s", timeStr)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("小时格式错误: %s", parts[0])
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("分钟格式错误: %s", parts[1])
	}

	if hour < 0 || hour > 23 {
		return 0, fmt.Errorf("小时超出范围: %d", hour)
	}

	if minute < 0 || minute > 59 {
		return 0, fmt.Errorf("分钟超出范围: %d", minute)
	}

	return hour*60 + minute, nil
}

// executeSyncForUser 为用户执行同步
// B5/B6 fix: 识别永久错误(学校密码未设置)并通知用户；max_retry_count 达到上限时禁用同步
func (s *SyncService) executeSyncForUser(userID uint) {
	startTime := time.Now()

	settingPtr, err := models.GetSyncSettingByUserID(userID)
	if err != nil {
		s.logSyncResult(userID, "auto", "failed",
			fmt.Sprintf("获取同步设置失败: %v", err), 0, time.Since(startTime))
		return
	}

	// B6 fix: 同步前检查是否已达到重试上限
	if settingPtr.AutoRetryEnabled && settingPtr.RetryCount > 0 && settingPtr.RetryCount >= settingPtr.MaxRetryCount {
		log.Printf("[SyncService] 用户 %d 重试次数已达上限(%d/%d)，跳过本次同步调度",
			userID, settingPtr.RetryCount, settingPtr.MaxRetryCount)
		return
	}

	//nolint:errcheck
	settingPtr.UpdateSyncStatus("syncing", "正在同步课程数据...", false)

	user, err := models.FindUserByID(userID)
	if err != nil {
		s.logSyncResult(userID, "auto", "failed",
			fmt.Sprintf("获取用户信息失败: %v", err), 0, time.Since(startTime))
		//nolint:errcheck
		settingPtr.UpdateSyncStatus("failed", fmt.Sprintf("获取用户信息失败: %v", err), false)
		return
	}

	if user.Username == "" || user.Password == "" {
		s.logSyncResult(userID, "auto", "failed",
			"用户未绑定学校账号", 0, time.Since(startTime))
		//nolint:errcheck
		settingPtr.UpdateSyncStatus("failed", "用户未绑定学校账号", false)
		return
	}

	coursesCount, err := s.syncCoursesForUser(user)
	if err != nil {
		errorMessage := fmt.Sprintf("同步课程失败: %v", err)
		s.logSyncResultWithDetail(userID, "auto", "failed",
			errorMessage, coursesCount, time.Since(startTime), err.Error())
		// B5 fix: 检测永久错误并通知用户
		permanentErr := errors.Is(err, utils.ErrSchoolPasswordNotSet)
		if permanentErr {
			s.notifySchoolPasswordMissing(user)
		}
		//nolint:errcheck
		settingPtr.UpdateSyncStatus("failed", errorMessage, permanentErr)
		return
	}

	//nolint:errcheck
	settingPtr.CoursesCount = coursesCount
	s.logSyncResult(userID, "auto", "success",
		fmt.Sprintf("成功同步 %d 门课程", coursesCount), coursesCount, time.Since(startTime))
	//nolint:errcheck
	settingPtr.UpdateSyncStatus("success", fmt.Sprintf("成功同步 %d 门课程", coursesCount), false)

	log.Printf("用户 %d 自动同步完成，同步了 %d 门课程", userID, coursesCount)
}

// syncCoursesForUser 为用户同步课程数据
func (s *SyncService) syncCoursesForUser(user *models.User) (int, error) {
	proxyClient := utils.NewProxyClient()

	log.Printf("开始为用户 %d 同步整个学期的课程数据", user.ID)

	semesterInfo, err := s.fetchCurrentSemesterInfo(proxyClient, user)
	if err != nil {
		log.Printf("获取当前学期信息失败: %v", err)
		return 0, fmt.Errorf("获取当前学期信息失败: %w", err)
	}

	log.Printf("获取到学期信息: 学期=%s, 总周数=%d, 当前周=%d",
		semesterInfo.CurrentSemester, semesterInfo.WeekCount, semesterInfo.CurrentWeek)

	var allCourses []models.Course
	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 5)

	for week := 1; week <= semesterInfo.WeekCount; week++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("正在获取第 %d 周的课程数据...", w)
			weekCourses, err := s.fetchSemesterTimetableFromSchoolWithWeek(
				proxyClient, user, semesterInfo.CurrentSemester, w)

			if err != nil {
				log.Printf("获取第 %d 周课程数据失败: %v", w, err)
				return
			}

			if len(weekCourses) > 0 {
				mu.Lock()
				allCourses = append(allCourses, weekCourses...)
				mu.Unlock()
			}
		}(week)
	}
	wg.Wait()

	log.Printf("课程数据获取完成，总共获取到 %d 门课程", len(allCourses))

	savedCourses, err := s.saveTimetableToDatabase(allCourses, user.ID)
	if err != nil {
		log.Printf("保存整学期课程数据失败: %v", err)
		return 0, fmt.Errorf("保存整学期课程数据失败: %w", err)
	}

	log.Printf("成功同步整学期 %d 门课程到数据库", len(savedCourses))
	return len(savedCourses), nil
}

// ManualSync 手动同步
func (s *SyncService) ManualSync(userID uint) error {
	startTime := time.Now()

	user, err := models.FindUserByID(userID)
	if err != nil {
		s.logSyncResult(userID, "manual", "failed",
			fmt.Sprintf("获取用户信息失败: %v", err), 0, time.Since(startTime))
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	if user.Username == "" || user.Password == "" {
		s.logSyncResult(userID, "manual", "failed",
			"用户未绑定学校账号", 0, time.Since(startTime))
		return fmt.Errorf("用户未绑定学校账号")
	}

	setting, err := models.GetSyncSettingByUserID(userID)
	if err != nil {
		s.logSyncResult(userID, "manual", "failed",
			fmt.Sprintf("获取同步设置失败: %v", err), 0, time.Since(startTime))
		return fmt.Errorf("获取同步设置失败: %w", err)
	}

	//nolint:errcheck
	setting.UpdateSyncStatus("syncing", "正在手动同步课程数据...", false)

	coursesCount, err := s.syncCoursesForUser(user)
	if err != nil {
		//nolint:errcheck
		errorMessage := fmt.Sprintf("同步课程失败: %v", err)
		s.logSyncResultWithDetail(userID, "manual", "failed",
			errorMessage, coursesCount, time.Since(startTime), err.Error())
		//nolint:errcheck
		permanentErr := errors.Is(err, utils.ErrSchoolPasswordNotSet)
		setting.UpdateSyncStatus("failed", errorMessage, permanentErr)
		return fmt.Errorf("同步课程失败: %w", err)
		//nolint:errcheck
	}

	setting.CoursesCount = coursesCount
	s.logSyncResult(userID, "manual", "success",
		fmt.Sprintf("成功同步 %d 门课程", coursesCount), coursesCount, time.Since(startTime))
	//nolint:errcheck
	setting.UpdateSyncStatus("success", fmt.Sprintf("成功同步 %d 门课程", coursesCount), false)

	return nil
}

// logSyncResult 记录同步结果
func (s *SyncService) logSyncResult(userID uint, syncType, status, message string, coursesCount int, duration time.Duration) {
	s.logSyncResultWithDetail(userID, syncType, status, message, coursesCount, duration, "")
}

// logSyncResultWithDetail 记录同步结果（包含详细错误信息）
func (s *SyncService) logSyncResultWithDetail(userID uint, syncType, status, message string, coursesCount int, duration time.Duration, errorDetail string) {
	now := time.Now()
	log := &models.SyncLog{
		UserID:      userID,
		SyncType:    syncType,
		Status:      status,
		Message:     message,
		CoursesSync: coursesCount,
		Duration:    int(duration.Milliseconds()),
		ErrorDetail: errorDetail,
		SyncedAt:    &now,
	}

	if err := models.CreateSyncLog(log); err != nil {
		fmt.Printf("创建同步日志失败: %v\n", err)
	}
}
