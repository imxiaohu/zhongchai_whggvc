package services

import (
	"fmt"
	"log"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// sendLikeEmailNotification 发送点赞邮件通知
func (mcns *MultiChannelNotificationService) sendLikeEmailNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	subject := "收到新的点赞"
	content := fmt.Sprintf("%s 点赞了你的帖子《%s》", data.FromUserName, data.PostTitle)

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeCommunityLike,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建点赞邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getCommunityEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":     user.Realname,
			"Subject":      subject,
			"Content":      content,
			"PostTitle":    data.PostTitle,
			"FromUserName": data.FromUserName,
			"ActionType":   "点赞",
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
		log.Printf("更新点赞邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendLikeDingTalkNotification 发送点赞钉钉通知
func (mcns *MultiChannelNotificationService) sendLikeDingTalkNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	title := "收到新的点赞"
	content := fmt.Sprintf("**%s** 点赞了你的帖子\n\n**帖子标题：** %s\n\n**时间：** %s",
		data.FromUserName, data.PostTitle, time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeCommunityLike,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建点赞钉钉通知日志失败: %v", err)
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
		log.Printf("更新点赞钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendBookmarkEmailNotification 发送收藏邮件通知
func (mcns *MultiChannelNotificationService) sendBookmarkEmailNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	subject := "收到新的收藏"
	content := fmt.Sprintf("%s 收藏了你的帖子《%s》", data.FromUserName, data.PostTitle)

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeCommunityBookmark,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建收藏邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getCommunityEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":     user.Realname,
			"Subject":      subject,
			"Content":      content,
			"PostTitle":    data.PostTitle,
			"FromUserName": data.FromUserName,
			"ActionType":   "收藏",
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
		log.Printf("更新收藏邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendBookmarkDingTalkNotification 发送收藏钉钉通知
func (mcns *MultiChannelNotificationService) sendBookmarkDingTalkNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	title := "收到新的收藏"
	content := fmt.Sprintf("**%s** 收藏了你的帖子\n\n**帖子标题：** %s\n\n**时间：** %s",
		data.FromUserName, data.PostTitle, time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeCommunityBookmark,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建收藏钉钉通知日志失败: %v", err)
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
		log.Printf("更新收藏钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentEmailNotification 发送评论邮件通知
func (mcns *MultiChannelNotificationService) sendCommentEmailNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	subject := "收到新的评论"
	content := fmt.Sprintf("%s 评论了你的帖子《%s》", data.FromUserName, data.PostTitle)
	if data.CommentContent != nil {
		content += fmt.Sprintf("：%s", *data.CommentContent)
	}

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeCommunityComment,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getCommunityEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":       user.Realname,
			"Subject":        subject,
			"Content":        content,
			"PostTitle":      data.PostTitle,
			"FromUserName":   data.FromUserName,
			"ActionType":     "评论",
			"CommentContent": data.CommentContent,
			"SendTime":       time.Now().Format("2006-01-02 15:04:05"),
			"SystemName":     "评教系统",
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
		log.Printf("更新评论邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentDingTalkNotification 发送评论钉钉通知
func (mcns *MultiChannelNotificationService) sendCommentDingTalkNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	title := "收到新的评论"
	content := fmt.Sprintf("**%s** 评论了你的帖子\n\n**帖子标题：** %s", data.FromUserName, data.PostTitle)
	if data.CommentContent != nil {
		content += fmt.Sprintf("\n\n**评论内容：** %s", *data.CommentContent)
	}
	content += fmt.Sprintf("\n\n**时间：** %s", time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeCommunityComment,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论钉钉通知日志失败: %v", err)
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
		log.Printf("更新评论钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentLikeEmailNotification 发送评论点赞邮件通知
func (mcns *MultiChannelNotificationService) sendCommentLikeEmailNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	subject := "收到新的评论点赞"
	content := fmt.Sprintf("%s 点赞了你的评论", data.FromUserName)

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeCommunityCommentLike,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论点赞邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getCommunityEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":       user.Realname,
			"Subject":        subject,
			"Content":        content,
			"PostTitle":      data.PostTitle,
			"FromUserName":   data.FromUserName,
			"ActionType":     "点赞",
			"CommentContent": data.CommentContent,
			"SendTime":       time.Now().Format("2006-01-02 15:04:05"),
			"SystemName":     "评教系统",
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
		log.Printf("更新评论点赞邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentLikeDingTalkNotification 发送评论点赞钉钉通知
func (mcns *MultiChannelNotificationService) sendCommentLikeDingTalkNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	title := "收到新的评论点赞"
	content := fmt.Sprintf("**%s** 点赞了你的评论\n\n**帖子标题：** %s", data.FromUserName, data.PostTitle)
	if data.CommentContent != nil {
		content += fmt.Sprintf("\n\n**评论内容：** %s", *data.CommentContent)
	}
	content += fmt.Sprintf("\n\n**时间：** %s", time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeCommunityCommentLike,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论点赞钉钉通知日志失败: %v", err)
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
		log.Printf("更新评论点赞钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentReplyEmailNotification 发送评论回复邮件通知
func (mcns *MultiChannelNotificationService) sendCommentReplyEmailNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	subject := "收到新的评论回复"
	content := fmt.Sprintf("%s 回复了你的评论", data.FromUserName)

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeCommunityComment,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论回复邮件通知日志失败: %v", err)
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getCommunityEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":       user.Realname,
			"Subject":        subject,
			"Content":        content,
			"PostTitle":      data.PostTitle,
			"FromUserName":   data.FromUserName,
			"ActionType":     "回复",
			"CommentContent": data.CommentContent,
			"SendTime":       time.Now().Format("2006-01-02 15:04:05"),
			"SystemName":     "评教系统",
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
		log.Printf("更新评论回复邮件通知日志失败: %v", updateErr)
	}

	return err
}

// sendCommentReplyDingTalkNotification 发送评论回复钉钉通知
func (mcns *MultiChannelNotificationService) sendCommentReplyDingTalkNotification(channel *models.NotificationChannel, user *models.User, data CommunityNotificationData) error {
	title := "收到新的评论回复"
	content := fmt.Sprintf("**%s** 回复了你的评论\n\n**帖子标题：** %s", data.FromUserName, data.PostTitle)
	if data.CommentContent != nil {
		content += fmt.Sprintf("\n\n**回复内容：** %s", *data.CommentContent)
	}
	content += fmt.Sprintf("\n\n**时间：** %s", time.Now().Format("2006-01-02 15:04:05"))

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    data.UserID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeCommunityComment,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建评论回复钉钉通知日志失败: %v", err)
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
		log.Printf("更新评论回复钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// getCommunityEmailTemplate 获取社区通知邮件模板
func getCommunityEmailTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .content { background: #fff; padding: 20px; border: 1px solid #e9ecef; border-radius: 8px; }
        .footer { text-align: center; margin-top: 20px; color: #6c757d; font-size: 12px; }
        .action-type { color: #007bff; font-weight: bold; }
        .post-title { color: #28a745; font-weight: bold; }
        .from-user { color: #dc3545; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>{{.SystemName}} - 社区通知</h2>
        </div>
        <div class="content">
            <p>亲爱的 <strong>{{.UserName}}</strong>，</p>
            <p>您收到了一个新的社区互动通知：</p>
            <p><span class="from-user">{{.FromUserName}}</span> <span class="action-type">{{.ActionType}}</span> 了您的帖子 <span class="post-title">《{{.PostTitle}}》</span></p>
            {{if .CommentContent}}
            <p><strong>评论内容：</strong>{{.CommentContent}}</p>
            {{end}}
            <p><strong>通知时间：</strong>{{.SendTime}}</p>
        </div>
        <div class="footer">
            <p>此邮件由{{.SystemName}}自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`
}
