package services

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// createScoreUpdate 创建详细的成绩更新记录
func (s *ScoreCheckService) createScoreUpdate(current models.ScoreSnapshot, snapshot *models.ScoreSnapshot, semester string, detailedScores map[string]string, changeType string) ScoreUpdate {
	oldScore := ""
	if snapshot != nil {
		oldScore = snapshot.Score
	}

	// 获取原始成绩数据以提取额外信息
	originalData, err := s.getOriginalScoreData(current.UserID, current.CourseCode, semester)
	if err != nil {
		log.Printf("获取原始成绩数据失败: %v", err)
	}

	update := ScoreUpdate{
		CourseName: fmt.Sprintf("[%s] %s", semester, current.CourseName),
		CourseCode: current.CourseCode,
		Semester:   semester,
		ScoreType:  s.getScoreTypeText(current.ScoreType),
		OldScore:   oldScore,
		NewScore:   current.Score,
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),

		// 详细成绩信息
		DailyScore:         detailedScores["daily"],
		ExamScore:          detailedScores["exam"],
		FinalScore:         detailedScores["final"],
		PracticalScore:     detailedScores["practical"],
		SupplementaryScore: detailedScores["supplementary"],

		// 课程信息
		Credit:     current.Credit,
		GPA:        current.GPA,
		ChangeType: changeType,
	}

	// 如果有原始数据，填充额外信息
	if originalData != nil {
		update.CourseProperty = originalData.CourseProperty
		update.TeacherNames = originalData.TeacherNames
		update.TestNote = originalData.TestNote
	}

	return update
}

// getOriginalScoreData 获取原始成绩数据（从API或缓存）
func (s *ScoreCheckService) getOriginalScoreData(userID uint, courseCode, semester string) (*OriginalScoreData, error) {
	// 这里可以从缓存或重新调用API获取原始数据
	// 为了简化，我们先返回默认值
	return &OriginalScoreData{
		CourseProperty: "必修课",
		TeacherNames:   "",
		TestNote:       "正常",
	}, nil
}

// OriginalScoreData 原始成绩数据结构
type OriginalScoreData struct {
	CourseProperty string
	TeacherNames   string
	TestNote       string
}

// getSemestersToCheck 根据用户配置获取要检测的学期列表
func (s *ScoreCheckService) getSemestersToCheck(user *models.User, semesterConfig string) ([]string, error) {
	switch semesterConfig {
	case "current":
		// 获取当前学期
		currentSemester, err := s.schoolAPIService.GetCurrentSemester(user)
		if err != nil {
			log.Printf("获取当前学期失败，使用默认学期: %v", err)
			return []string{"2024-2025学年第二学期"}, nil
		}
		return []string{currentSemester.Name}, nil

	case "all":
		// 获取所有可用学期
		semesters, err := s.schoolAPIService.GetAvailableSemesters(user)
		if err != nil {
			log.Printf("获取所有学期失败，使用默认学期: %v", err)
			return []string{"2024-2025学年第二学期"}, nil
		}

		var semesterNames []string
		for _, semester := range semesters {
			semesterNames = append(semesterNames, semester.Name)
		}
		return semesterNames, nil

	default:
		// 用户指定的具体学期
		if semesterConfig != "" {
			return []string{semesterConfig}, nil
		}
		// 默认使用当前学期
		return []string{"2024-2025学年第二学期"}, nil
	}
}

// fetchCurrentScoresBySemester 根据学期获取当前成绩数据
func (s *ScoreCheckService) fetchCurrentScoresBySemester(user *models.User, semester string) ([]models.ScoreSnapshot, error) {
	// 调用学校API服务获取成绩数据
	response, err := s.fetchScoreDataFromSchool(user, semester)
	if err != nil {
		return nil, fmt.Errorf("从学校服务器获取成绩失败: %w", err)
	}

	// 解析响应数据为ScoreSnapshot格式
	snapshots, err := s.parseScoreDataToSnapshots(response, user.ID, semester)
	if err != nil {
		return nil, fmt.Errorf("解析成绩数据失败: %w", err)
	}

	return snapshots, nil
}

// updateSnapshotsInPlace 原地更新快照（修复版：避免版本轮换导致的误报）
// 核心逻辑：直接在 current 版本上做 upsert，不做轮换。
// 历史快照通过对比来判断是否有真实变化，首次同步时变化不发送通知。
func (s *ScoreCheckService) updateSnapshotsInPlace(userID uint, semester string, currentScores []models.ScoreSnapshot) error {
	log.Printf("[ScoreCheck] 原地更新用户%d学期%s的成绩快照（%d条）", userID, semester, len(currentScores))

	// 预计算所有快照的校验和
	for i := range currentScores {
		currentScores[i].UserID = userID
		currentScores[i].Semester = semester
		currentScores[i].Version = models.ScoreVersionCurrent
		currentScores[i].IsActive = true
		currentScores[i].CheckSum = s.calculateCheckSum(&currentScores[i])
	}

	// 使用批量 INSERT ON DUPLICATE KEY UPDATE
	failCount, lastErr := models.BatchCreateOrUpdateScoreSnapshots(currentScores)
	successCount := len(currentScores) - failCount

	log.Printf("[ScoreCheck] 用户%d学期%s快照更新完成: 成功%d条, 失败%d条",
		userID, semester, successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("快照更新失败: 成功%d条, 失败%d条: %w", successCount, failCount, lastErr)
	}

	return nil
}

// calculateCheckSum 计算成绩数据校验和
func (s *ScoreCheckService) calculateCheckSum(score *models.ScoreSnapshot) string {
	data := fmt.Sprintf("%s_%s_%s_%s_%.2f_%.2f_%s",
		score.Semester, score.CourseCode, score.CourseName, score.ScoreType,
		score.Credit, score.GPA, score.Score)

	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getScoreTypeText 获取成绩类型文本
func (s *ScoreCheckService) getScoreTypeText(scoreType string) string {
	switch scoreType {
	case "daily":
		return "平时分"
	case "exam":
		return "考试分"
	case "practical":
		return "实践分"
	case "final":
		return "最终分"
	case "supplementary":
		return "补考分"
	default:
		return scoreType
	}
}

// CheckAllUsersScores 检查所有用户的成绩更新（优化版：减少 DB 查询，按频率过滤）
func (s *ScoreCheckService) CheckAllUsersScores() {
	if !s.IsEnabled() {
		log.Printf("[ScoreCheck] 成绩检查服务未启用")
		return
	}

	startTime := time.Now()
	log.Printf("[ScoreCheck] === 开始执行成绩检查任务 === %s", startTime.Format("2006-01-02 15:04:05"))

	// 获取所有启用成绩检查的用户（使用索引优化）
	var channels []models.NotificationChannel
	if err := models.DB.Where("score_check_enabled = ?", true).Find(&channels).Error; err != nil {
		log.Printf("[ScoreCheck] 获取启用成绩检查的用户失败: %v", err)
		return
	}

	log.Printf("[ScoreCheck] 找到%d个启用成绩检查的用户", len(channels))

	checkedCount := 0
	updatedCount := 0
	errorCount := 0
	skippedCount := 0

	for i, channel := range channels {
		// 冷却检查：只有在距上次检查超过配置频率时才真正执行
		if !s.shouldCheckForScheduler(&channel) {
			log.Printf("[ScoreCheck] 用户%d跳过检查 (频率: %s, 上次检查: %v)",
				channel.UserID, channel.ScoreCheckFrequency, channel.LastScoreCheck)
			skippedCount++
			continue
		}

		log.Printf("[ScoreCheck] 用户 %d/%d (ID: %d) 开始检查",
			i+1, len(channels), channel.UserID)
		checkedCount++

		scoreUpdates, err := s.CheckUserScores(channel.UserID)
		if err != nil {
			log.Printf("[ScoreCheck] 检查用户%d成绩失败: %v", channel.UserID, err)
			errorCount++
			continue
		}

		if len(scoreUpdates) > 0 {
			log.Printf("[ScoreCheck] 用户%d检测到%d条真实成绩更新", channel.UserID, len(scoreUpdates))
			updatedCount++

			for j, update := range scoreUpdates {
				log.Printf("[ScoreCheck]   更新%d: %s (%s) %s",
					j+1, update.CourseName, update.ScoreType, update.ChangeType)
			}

			s.sendScoreUpdateNotifications(channel.UserID, scoreUpdates)
		} else {
			log.Printf("[ScoreCheck] 用户%d无成绩更新", channel.UserID)
		}

		// 更新最后检查时间
		now := time.Now()
		channel.LastScoreCheck = &now
		if err := models.DB.Save(&channel).Error; err != nil {
			log.Printf("[ScoreCheck] 更新用户%d最后检查时间失败: %v", channel.UserID, err)
		}
	}

	duration := time.Since(startTime)
	log.Printf("[ScoreCheck] === 成绩检查任务完成 ===")
	log.Printf("[ScoreCheck] 统计: 总用户=%d, 实际检查=%d, 有更新=%d, 失败=%d, 跳过=%d, 耗时=%v",
		len(channels), checkedCount, updatedCount, errorCount, skippedCount, duration)
}

// shouldCheckForScheduler 调度器专用的冷却检查（不查快照，避免额外 DB 开销）
func (s *ScoreCheckService) shouldCheckForScheduler(channel *models.NotificationChannel) bool {
	if channel.LastScoreCheck == nil {
		log.Printf("[ScoreCheck] 用户%d从未检查过成绩，强制检查", channel.UserID)
		return true
	}

	now := time.Now()
	timeSinceLastCheck := now.Sub(*channel.LastScoreCheck)

	log.Printf("[ScoreCheck] 用户%d冷却检查: 频率=%s, 距上次=%.1f小时",
		channel.UserID, channel.ScoreCheckFrequency, timeSinceLastCheck.Hours())

	return s.shouldCheckFrequencyOnly(channel)
}

// shouldCheckFrequencyOnly 仅根据频率判断是否需要检查（不含快照检查）
func (s *ScoreCheckService) shouldCheckFrequencyOnly(channel *models.NotificationChannel) bool {
	if channel.LastScoreCheck == nil {
		return true
	}

	now := time.Now()
	timeSinceLastCheck := now.Sub(*channel.LastScoreCheck)

	switch channel.ScoreCheckFrequency {
	case models.CheckFrequencyHourly:
		return timeSinceLastCheck >= time.Hour
	case models.CheckFrequencyDaily:
		return timeSinceLastCheck >= 24*time.Hour
	case models.CheckFrequencyWeekly:
		return timeSinceLastCheck >= 7*24*time.Hour
	default:
		return timeSinceLastCheck >= 24*time.Hour
	}
}

// sendScoreUpdateNotifications 发送成绩更新通知
func (s *ScoreCheckService) sendScoreUpdateNotifications(userID uint, scoreUpdates []ScoreUpdate) {
	log.Printf("[ScoreCheck] 开始发送成绩更新通知，用户ID: %d, 更新数量: %d", userID, len(scoreUpdates))

	// 使用多渠道通知服务
	multiChannelService := GetMultiChannelNotificationService()
	if multiChannelService != nil {
		notificationData := ScoreUpdateNotificationData{
			UserID:       userID,
			ScoreUpdates: scoreUpdates,
			UpdateCount:  len(scoreUpdates),
		}

		if err := multiChannelService.SendScoreUpdateNotification(notificationData); err != nil {
			log.Printf("[ScoreCheck] 发送成绩更新多渠道通知失败: %v", err)
			s.sendDirectNotifications(userID, scoreUpdates)
		} else {
			log.Printf("[ScoreCheck] 成绩更新多渠道通知发送成功")
		}
	} else {
		log.Printf("多渠道通知服务不可用，使用备用方案")
		s.sendDirectNotifications(userID, scoreUpdates)
	}

	// 发送到RabbitMQ队列进行异步处理（可选）
	rabbitMQ := GetRabbitMQService()
	if rabbitMQ != nil && rabbitMQ.IsEnabled() {
		// 序列化成绩更新数据
		updatesJSON, err := json.Marshal(scoreUpdates)
		if err != nil {
			log.Printf("序列化成绩更新数据失败: %v", err)
			return
		}

		msg := NotificationMessage{
			Type:        models.NotificationTypeScoreUpdate,
			UserID:      userID,
			Title:       "成绩更新通知",
			Content:     fmt.Sprintf("您有%d条成绩更新", len(scoreUpdates)),
			RelatedType: "score_update",
			Data: map[string]interface{}{
				"scoreUpdates": string(updatesJSON),
				"updateCount":  len(scoreUpdates),
			},
		}

		if err := rabbitMQ.PublishNotification(msg); err != nil {
			log.Printf("发送成绩更新通知到队列失败: %v", err)
		}
	}
}

// sendDirectNotifications 直接发送通知
func (s *ScoreCheckService) sendDirectNotifications(userID uint, scoreUpdates []ScoreUpdate) {
	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		log.Printf("获取用户%d通知配置失败: %v", userID, err)
		return
	}

	// 发送邮件通知
	if channel.EmailEnabled && channel.ScoreUpdateEmail {
		emailService := GetEmailService()
		if emailService != nil && emailService.IsEnabled() {
			if err := emailService.SendScoreUpdateEmail(userID, scoreUpdates); err != nil {
				log.Printf("发送成绩更新邮件失败: %v", err)
			}
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.ScoreUpdateDingTalk {
		dingTalkService := GetDingTalkService()
		if dingTalkService != nil && dingTalkService.IsEnabled() {
			if err := dingTalkService.SendScoreUpdateMessage(userID, scoreUpdates); err != nil {
				log.Printf("发送成绩更新钉钉消息失败: %v", err)
			}
		}
	}

	// 发送短信通知
	if channel.SMSEnabled && channel.ScoreUpdateSMS {
		smsService := GetSMSService()
		if smsService != nil && smsService.IsEnabled() {
			if err := smsService.SendScoreUpdateSMS(userID, scoreUpdates); err != nil {
				log.Printf("发送成绩更新短信失败: %v", err)
			}
		}
	}
}

// StartScoreCheckScheduler 启动成绩检查调度器
func (s *ScoreCheckService) StartScoreCheckScheduler() {
	if !s.IsEnabled() {
		log.Printf("成绩检查服务未启用，跳过调度器启动")
		return
	}

	log.Printf("启动成绩检查调度器...")

	// 每6分钟检查一次（服务端检查频率）
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()

		// 启动时立即执行一次检查
		log.Printf("[ScoreCheck] 调度器启动，首次执行检查")
		s.CheckAllUsersScores()

		for range ticker.C {
			log.Printf("[ScoreCheck] 定时器触发，执行成绩检查")
			s.CheckAllUsersScores()
		}
	}()

	log.Printf("[ScoreCheck] 调度器已启动，每5分钟触发一次（实际请求由用户配置的检查频率控制）")
}

// UserHasScoreSnapshots 检查用户是否有成绩快照
func (s *ScoreCheckService) UserHasScoreSnapshots(userID uint) (bool, error) {
	var count int64
	err := models.DB.Model(&models.ScoreSnapshot{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error

	if err != nil {
		log.Printf("[ScoreCheck] 检查用户%d快照数量失败: %v", userID, err)
		return false, err
	}

	return count > 0, nil
}

// GetSemestersToCheck 根据用户配置获取要检测的学期列表（公开方法）
func (s *ScoreCheckService) GetSemestersToCheck(user *models.User, semesterConfig string) ([]string, error) {
	return s.getSemestersToCheck(user, semesterConfig)
}

// FetchCurrentScoresBySemesterWithCache 优先从缓存获取学期成绩数据（公开方法）
func (s *ScoreCheckService) FetchCurrentScoresBySemesterWithCache(user *models.User, semester string) ([]models.ScoreSnapshot, error) {
	return s.fetchCurrentScoresBySemesterWithCache(user, semester)
}

// CheckScoreUpdate 检查单条成绩更新（辅助方法）
func (s *ScoreCheckService) CheckScoreUpdate(current models.ScoreSnapshot, snapshot *models.ScoreSnapshot, semester string) *ScoreUpdate {
	// 检查是否有变化
	if snapshot != nil && current.CheckSum == snapshot.CheckSum {
		return nil // 无变化
	}

	detailedScores := s.buildDetailedScoreInfo([]models.ScoreSnapshot{current}, current)
	changeType := "new"
	if snapshot != nil {
		changeType = "update"
	}

	update := s.createScoreUpdate(current, snapshot, semester, detailedScores, changeType)
	return &update
}

// FormatScoreUpdateForDisplay 格式化成绩更新用于显示
func (s *ScoreCheckService) FormatScoreUpdateForDisplay(update *ScoreUpdate) string {
	var parts []string

	if update.ChangeType == "update" && update.OldScore != "" {
		parts = append(parts, fmt.Sprintf("%s: %s → %s", update.ScoreType, update.OldScore, update.NewScore))
	} else {
		parts = append(parts, fmt.Sprintf("%s: %s", update.ScoreType, update.NewScore))
	}

	return strings.Join(parts, ", ")
}
