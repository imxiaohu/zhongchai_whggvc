package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
)

// MultiChannelNotificationService 多渠道通知服务
type MultiChannelNotificationService struct {
	emailService       *EmailService
	dingTalkService    *DingTalkService
	templateManager    *config.EmailTemplateManager
	smsTemplateManager *config.SMSTemplateManager
}

var multiChannelNotificationService *MultiChannelNotificationService

// InitMultiChannelNotificationService 初始化多渠道通知服务
func InitMultiChannelNotificationService() {
	multiChannelNotificationService = NewMultiChannelNotificationService()
}

// GetMultiChannelNotificationService 获取多渠道通知服务实例
func GetMultiChannelNotificationService() *MultiChannelNotificationService {
	return multiChannelNotificationService
}

// NewMultiChannelNotificationService 创建新的多渠道通知服务
func NewMultiChannelNotificationService() *MultiChannelNotificationService {
	return &MultiChannelNotificationService{
		emailService:       GetEmailService(),
		dingTalkService:    GetDingTalkService(),
		templateManager:    config.GetEmailTemplateManager(),
		smsTemplateManager: config.GetSMSTemplateManager(),
	}
}

// CommunityNotificationData 社区通知数据结构
type CommunityNotificationData struct {
	Type           string  // like, bookmark, comment
	UserID         uint    // 接收通知的用户ID
	FromUserID     uint    // 触发通知的用户ID
	PostID         uint    // 帖子ID
	PostTitle      string  // 帖子标题
	FromUserName   string  // 触发用户昵称
	CommentID      *uint   // 评论ID（仅评论通知）
	CommentContent *string // 评论内容（仅评论通知）
}

// SendCommunityNotification 发送社区互动通知
func (mcns *MultiChannelNotificationService) SendCommunityNotification(data CommunityNotificationData) error {
	log.Printf("开始发送社区通知: 类型=%s, 用户=%d, 来源用户=%d", data.Type, data.UserID, data.FromUserID)

	// 获取用户通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(data.UserID)
	if err != nil {
		log.Printf("获取用户%d通知配置失败: %v", data.UserID, err)
		return err
	}

	log.Printf("用户%d通知配置: 邮箱启用=%v, 钉钉启用=%v", data.UserID, channel.EmailEnabled, channel.DingTalkEnabled)

	// 获取用户信息
	user, err := models.FindUserByID(data.UserID)
	if err != nil {
		log.Printf("获取用户%d信息失败: %v", data.UserID, err)
		return err
	}

	// 根据通知类型检查用户偏好并发送通知
	switch data.Type {
	case "like":
		return mcns.sendLikeNotifications(channel, user, data)
	case "bookmark":
		return mcns.sendBookmarkNotifications(channel, user, data)
	case "comment":
		return mcns.sendCommentNotifications(channel, user, data)
	case "comment_like":
		return mcns.sendCommentLikeNotifications(channel, user, data)
	case "comment_reply":
		return mcns.sendCommentReplyNotifications(channel, user, data)
	default:
		return fmt.Errorf("不支持的通知类型: %s", data.Type)
	}
}

// sendLikeNotifications 发送点赞通知
func (mcns *MultiChannelNotificationService) sendLikeNotifications(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.CommunityLikeEmail && channel.EmailAddress != "" {
		if err := mcns.sendLikeEmailNotification(channel, user, data); err != nil {
			log.Printf("发送点赞邮件通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.CommunityLikeDingTalk && channel.DingTalkWebhookURL != "" {
		if err := mcns.sendLikeDingTalkNotification(channel, user, data); err != nil {
			log.Printf("发送点赞钉钉通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// sendBookmarkNotifications 发送收藏通知
func (mcns *MultiChannelNotificationService) sendBookmarkNotifications(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.CommunityBookmarkEmail && channel.EmailAddress != "" {
		if err := mcns.sendBookmarkEmailNotification(channel, user, data); err != nil {
			log.Printf("发送收藏邮件通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.CommunityBookmarkDingTalk && channel.DingTalkWebhookURL != "" {
		if err := mcns.sendBookmarkDingTalkNotification(channel, user, data); err != nil {
			log.Printf("发送收藏钉钉通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// sendCommentNotifications 发送评论通知
func (mcns *MultiChannelNotificationService) sendCommentNotifications(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.CommunityCommentEmail && channel.EmailAddress != "" {
		if err := mcns.sendCommentEmailNotification(channel, user, data); err != nil {
			log.Printf("发送评论邮件通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.CommunityCommentDingTalk && channel.DingTalkWebhookURL != "" {
		if err := mcns.sendCommentDingTalkNotification(channel, user, data); err != nil {
			log.Printf("发送评论钉钉通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// sendCommentLikeNotifications 发送评论点赞通知
func (mcns *MultiChannelNotificationService) sendCommentLikeNotifications(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.CommunityCommentLikeEmail && channel.EmailAddress != "" {
		if err := mcns.sendCommentLikeEmailNotification(channel, user, data); err != nil {
			log.Printf("发送评论点赞邮件通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.CommunityCommentLikeDingTalk && channel.DingTalkWebhookURL != "" {
		if err := mcns.sendCommentLikeDingTalkNotification(channel, user, data); err != nil {
			log.Printf("发送评论点赞钉钉通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// sendCommentReplyNotifications 发送评论回复通知
func (mcns *MultiChannelNotificationService) sendCommentReplyNotifications(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.CommunityCommentEmail && channel.EmailAddress != "" {
		if err := mcns.sendCommentReplyEmailNotification(channel, user, data); err != nil {
			log.Printf("发送评论回复邮件通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.CommunityCommentDingTalk && channel.DingTalkWebhookURL != "" {
		if err := mcns.sendCommentReplyDingTalkNotification(channel, user, data); err != nil {
			log.Printf("发送评论回复钉钉通知失败: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// ScoreUpdateNotificationData 成绩更新通知数据结构
type ScoreUpdateNotificationData struct {
	UserID       uint          // 接收通知的用户ID
	ScoreUpdates []ScoreUpdate // 成绩更新列表
	UpdateCount  int           // 更新数量
}

// SendScoreUpdateNotification 发送成绩更新通知
func (mcns *MultiChannelNotificationService) SendScoreUpdateNotification(data ScoreUpdateNotificationData) error {
	log.Printf("开始发送成绩更新通知: 用户=%d, 更新数量=%d", data.UserID, data.UpdateCount)

	// 获取用户通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(data.UserID)
	if err != nil {
		log.Printf("获取用户%d通知配置失败: %v", data.UserID, err)
		return err
	}

	log.Printf("用户%d成绩通知配置: 邮箱启用=%v, 钉钉启用=%v, 短信启用=%v",
		data.UserID, channel.EmailEnabled && channel.ScoreUpdateEmail,
		channel.DingTalkEnabled && channel.ScoreUpdateDingTalk,
		channel.SMSEnabled && channel.ScoreUpdateSMS)

	// 获取用户信息
	user, err := models.FindUserByID(data.UserID)
	if err != nil {
		log.Printf("获取用户%d信息失败: %v", data.UserID, err)
		return err
	}

	return mcns.sendScoreUpdateNotifications(channel, user, data)
}

// sendScoreUpdateNotifications 发送成绩更新通知
func (mcns *MultiChannelNotificationService) sendScoreUpdateNotifications(channel *models.NotificationChannel, user *models.User, data ScoreUpdateNotificationData) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// 发送邮件通知
	if channel.EmailEnabled && channel.ScoreUpdateEmail && channel.EmailAddress != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := mcns.sendScoreUpdateEmailNotification(channel, user, data); err != nil {
				log.Printf("发送成绩更新邮件通知失败: %v", err)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}()
	}

	// 发送钉钉通知
	if channel.DingTalkEnabled && channel.ScoreUpdateDingTalk && channel.DingTalkWebhookURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := mcns.sendScoreUpdateDingTalkNotification(channel, user, data); err != nil {
				log.Printf("发送成绩更新钉钉通知失败: %v", err)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}()
	}

	// 发送短信通知
	if channel.SMSEnabled && channel.ScoreUpdateSMS && channel.PhoneNumber != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := mcns.sendScoreUpdateSMSNotification(channel, user, data); err != nil {
				log.Printf("发送成绩更新短信通知失败: %v", err)
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("部分通知发送失败: %v", errors)
	}
	return nil
}

// sendScoreUpdateEmailNotification 发送成绩更新邮件通知
func (mcns *MultiChannelNotificationService) sendScoreUpdateEmailNotification(channel *models.NotificationChannel, user *models.User, data ScoreUpdateNotificationData) error {
	subject := fmt.Sprintf("成绩更新通知 - 您有%d条成绩更新", data.UpdateCount)

	// 构建成绩更新内容
	var contentBuilder []string
	contentBuilder = append(contentBuilder, fmt.Sprintf("亲爱的%s，您有%d条成绩更新：", user.Realname, data.UpdateCount))
	contentBuilder = append(contentBuilder, "")

	for i, update := range data.ScoreUpdates {
		if i >= 10 { // 最多显示10条
			contentBuilder = append(contentBuilder, fmt.Sprintf("... 还有%d条更新", data.UpdateCount-10))
			break
		}

		if update.OldScore == "" {
			contentBuilder = append(contentBuilder, fmt.Sprintf("• %s (%s): 新增成绩 %s",
				update.CourseName, update.ScoreType, update.NewScore))
		} else {
			contentBuilder = append(contentBuilder, fmt.Sprintf("• %s (%s): %s → %s",
				update.CourseName, update.ScoreType, update.OldScore, update.NewScore))
		}
	}

	content := fmt.Sprintf("%s\n\n请及时登录系统查看详细信息。",
		fmt.Sprintf("%s", contentBuilder))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeScoreUpdate,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建成绩更新邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getScoreUpdateEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":     user.Realname,
			"Subject":      subject,
			"Content":      content,
			"ScoreUpdates": data.ScoreUpdates,
			"UpdateCount":  data.UpdateCount,
			"SendTime":     time.Now().Format("2006-01-02 15:04:05"),
			"SystemName":   "评教系统",
		},
	}

	// 发送邮件
	err := mcns.emailService.SendTemplateEmail(channel.EmailAddress, templateData)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	// 更新通知日志
	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新成绩更新邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendScoreUpdateDingTalkNotification 发送成绩更新钉钉通知
func (mcns *MultiChannelNotificationService) sendScoreUpdateDingTalkNotification(channel *models.NotificationChannel, user *models.User, data ScoreUpdateNotificationData) error {
	title := fmt.Sprintf("成绩更新通知 - %d条更新", data.UpdateCount)

	// 构建钉钉消息内容
	content := fmt.Sprintf("**%s** 您好！\n\n您有 **%d** 条成绩更新：\n\n", user.Realname, data.UpdateCount)

	for i, update := range data.ScoreUpdates {
		if i >= 5 { // 钉钉消息最多显示5条
			content += fmt.Sprintf("... 还有%d条更新\n\n", data.UpdateCount-5)
			break
		}

		if update.OldScore == "" {
			content += fmt.Sprintf("• **%s** (%s): 新增成绩 **%s**\n",
				update.CourseName, update.ScoreType, update.NewScore)
		} else {
			content += fmt.Sprintf("• **%s** (%s): %s → **%s**\n",
				update.CourseName, update.ScoreType, update.OldScore, update.NewScore)
		}
	}

	content += fmt.Sprintf("\n**时间：** %s\n\n请及时登录系统查看详细信息。",
		time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeScoreUpdate,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建成绩更新钉钉通知日志失败: %v", err)
	}

	// 发送钉钉消息
	err := mcns.dingTalkService.SendMarkdownMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, title, content)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	// 更新通知日志
	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新成绩更新钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendScoreUpdateSMSNotification 发送成绩更新短信通知
func (mcns *MultiChannelNotificationService) sendScoreUpdateSMSNotification(channel *models.NotificationChannel, user *models.User, data ScoreUpdateNotificationData) error {
	// 使用简化的短信内容（符合模板要求）
	shortContent := "发送成绩更新短信通知"

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelSMS,
		Type:      models.NotificationTypeScoreUpdate,
		Title:     "成绩更新通知",
		Content:   shortContent,
		Recipient: channel.PhoneNumber,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建成绩更新短信通知日志失败: %v", err)
	}

	// 发送短信
	smsService := GetSMSService()
	var err error
	if smsService != nil && smsService.IsEnabled() {
		// 使用配置的短信模板和签名
		templateConfig, tmplErr := mcns.smsTemplateManager.GetTemplate(config.TemplateScoreUpdate)
		if tmplErr != nil || templateConfig == nil {
			err = fmt.Errorf("获取短信模板配置失败: %w", tmplErr)
		} else {
			signName := config.GetEnvFirst([]string{"ALIYUN_SMS_SIGN_NAME", "SMS_SIGN_NAME"}, "学生评教系统")
			template := SMSTemplate{
				TemplateCode: templateConfig.TemplateCode,
				SignName:     signName,
				Params: map[string]string{
					"name": user.Realname,
				},
			}
			err = smsService.SendSMS(channel.PhoneNumber, template)
		}
	} else {
		err = fmt.Errorf("短信服务未启用")
	}

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	// 更新通知日志
	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新成绩更新短信通知日志失败: %v", updateErr)
	}

	return err
}

// sendSchoolPasswordMissingEmail 发送学校密码缺失邮件通知
// B5 fix: 当检测到用户学校密码未设置时，通知用户重新登录设置密码
func (mcns *MultiChannelNotificationService) sendSchoolPasswordMissingEmail(channel *models.NotificationChannel, user *models.User) error {
	subject := "重要通知：课程自动同步已暂停，请重新设置学校密码"

	content := fmt.Sprintf("%s 您好！\n\n您的课程自动同步已暂停，原因：系统检测到您的学校账号未设置密码，无法进行自动登录同步。\n\n请按以下步骤恢复同步：\n1. 退出当前账号登录\n2. 重新登录，输入您的学校账号密码\n3. 登录成功后，同步将自动恢复\n\n如有疑问，请联系管理员。\n\n— 学生评教系统\n%s",
		user.Realname, time.Now().Format("2006-01-02 15:04:05"))

	notificationLog := &models.NotificationLog{
		UserID:    user.ID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeReminder,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建学校密码缺失邮件通知日志失败: %v", err)
	}

	templateData := EmailTemplate{
		Subject: subject,
		Body:    content,
		Data: map[string]interface{}{
			"UserName":   user.Realname,
			"Subject":    subject,
			"Content":    content,
			"SendTime":   time.Now().Format("2006-01-02 15:04:05"),
			"SystemName": "学生评教系统",
		},
	}

	err := mcns.emailService.SendTemplateEmail(channel.EmailAddress, templateData)

	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新学校密码缺失邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendSchoolPasswordMissingDingTalk 发送学校密码缺失钉钉通知
// B5 fix: 当检测到用户学校密码未设置时，通过钉钉通知用户重新登录设置密码
func (mcns *MultiChannelNotificationService) sendSchoolPasswordMissingDingTalk(channel *models.NotificationChannel, user *models.User) error {
	title := "重要通知：课程自动同步已暂停"
	content := fmt.Sprintf("**%s** 您好！\n\n您的课程自动同步已暂停。\n\n**原因：** 系统检测到您的学校账号未设置密码，无法进行自动登录同步。\n\n**恢复步骤：**\n1. 退出当前账号登录\n2. 重新登录，输入您的学校账号密码\n3. 登录成功后，同步将自动恢复\n\n如有疑问，请联系管理员。\n\n— 学生评教系统\n**%s**",
		user.Realname, time.Now().Format("2006-01-02 15:04"))

	notificationLog := &models.NotificationLog{
		UserID:    user.ID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeReminder,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建学校密码缺失钉钉通知日志失败: %v", err)
	}

	err := mcns.dingTalkService.SendMarkdownMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, title, content)

	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新学校密码缺失钉钉通知日志失败: %v", updateErr)
	}

	return err
}
