package services

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"gopkg.in/gomail.v2"
)

// EmailService 邮件服务
type EmailService struct {
	dialer          *gomail.Dialer
	enabled         bool
	templateManager *config.EmailTemplateManager
}

// EmailTemplate 邮件模板数据
type EmailTemplate struct {
	Subject string
	Body    string
	Data    map[string]interface{}
}

var emailService *EmailService

// InitEmailService 初始化邮件服务
func InitEmailService() {
	emailService = NewEmailService()
}

// GetEmailService 获取邮件服务实例
func GetEmailService() *EmailService {
	return emailService
}

// NewEmailService 创建新的邮件服务
func NewEmailService() *EmailService {
	service := &EmailService{
		enabled:         config.GetEnvBool("EMAIL_ENABLED", true),
		templateManager: config.GetEmailTemplateManager(),
	}

	if service.enabled {
		log.Printf("邮件服务启用，开始初始化SMTP连接...")
		service.initSMTP()
	} else {
		log.Printf("邮件服务未启用")
	}

	return service
}

// initSMTP 初始化SMTP连接
func (e *EmailService) initSMTP() {
	host := config.GetEnv("SMTP_HOST", "smtp.qq.com")
	port := config.GetEnvInt("SMTP_PORT", 465)
	username := config.GetEnv("SMTP_USERNAME", "")
	password := config.GetEnv("SMTP_PASSWORD", "")

	e.dialer = gomail.NewDialer(host, port, username, password)
	e.dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	log.Printf("SMTP配置完成: %s:%d", host, port)
}

// IsEnabled 检查邮件服务是否启用
func (e *EmailService) IsEnabled() bool {
	return e.enabled && e.dialer != nil
}

// SendEmail 发送邮件
func (e *EmailService) SendEmail(to, subject, body string, isHTML bool) error {
	if !e.IsEnabled() {
		return fmt.Errorf("邮件服务未启用")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEnv("SMTP_USERNAME", ""))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	return e.dialer.DialAndSend(m)
}

// SendTemplateEmail 发送模板邮件
func (e *EmailService) SendTemplateEmail(to string, templateData EmailTemplate) error {
	if !e.IsEnabled() {
		return fmt.Errorf("邮件服务未启用")
	}

	// 解析模板
	tmpl, err := template.New("email").Parse(templateData.Body)
	if err != nil {
		return fmt.Errorf("解析邮件模板失败: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData.Data); err != nil {
		return fmt.Errorf("渲染邮件模板失败: %w", err)
	}

	return e.SendEmail(to, templateData.Subject, buf.String(), true)
}

// SendScoreUpdateEmail 发送成绩更新邮件
func (e *EmailService) SendScoreUpdateEmail(userID uint, scoreUpdates []ScoreUpdate) error {
	// 获取用户信息
	user, err := models.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取通知配置失败: %w", err)
	}

	if !channel.EmailEnabled || !channel.ScoreUpdateEmail || channel.EmailAddress == "" {
		return fmt.Errorf("用户未启用邮件通知或未配置邮箱地址")
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: "成绩更新通知",
		Body:    getScoreUpdateEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":     user.Realname,
			"ScoreUpdates": scoreUpdates,
			"UpdateTime":   time.Now().Format("2006-01-02 15:04:05"),
			"SystemName":   "评教系统",
		},
	}

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeScoreUpdate,
		Title:     templateData.Subject,
		Content:   fmt.Sprintf("发送成绩更新邮件，包含%d条更新", len(scoreUpdates)),
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建邮件通知日志失败: %v", err)
	}

	// 发送邮件
	err = e.SendTemplateEmail(channel.EmailAddress, templateData)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新邮件通知日志失败: %v", updateErr)
	}

	return err
}

// SendSystemEmail 发送系统邮件
func (e *EmailService) SendSystemEmail(userID uint, subject, content string) error {
	// 获取用户信息
	user, err := models.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取通知配置失败: %w", err)
	}

	if !channel.EmailEnabled || channel.EmailAddress == "" {
		return fmt.Errorf("用户未启用邮件通知或未配置邮箱地址")
	}

	// 准备邮件模板数据
	templateData := EmailTemplate{
		Subject: subject,
		Body:    getSystemEmailTemplate(),
		Data: map[string]interface{}{
			"UserName":   user.Realname,
			"Subject":    subject,
			"Content":    content,
			"SendTime":   time.Now().Format("2006-01-02 15:04:05"),
			"SystemName": "评教系统",
		},
	}

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelEmail,
		Type:      models.NotificationTypeSystem,
		Title:     subject,
		Content:   content,
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建邮件通知日志失败: %v", err)
	}

	// 发送邮件
	err = e.SendTemplateEmail(channel.EmailAddress, templateData)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新邮件通知日志失败: %v", updateErr)
	}

	return err
}

// getScoreUpdateEmailTemplate 获取成绩更新邮件模板
func getScoreUpdateEmailTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>成绩更新通知</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 700px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .score-item { background-color: white; margin: 15px 0; padding: 20px; border-left: 4px solid #4CAF50; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .course-info { background-color: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .score-details { margin: 15px 0; }
        .score-table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        .score-table td { padding: 8px 12px; border: 1px solid #ddd; }
        .score-table tr:nth-child(even) { background-color: #f8f9fa; }
        .final-score { color: #e74c3c; font-weight: bold; }
        .supplementary-score { color: #f39c12; font-weight: bold; }
        .change-info { background-color: #e8f5e8; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 3px solid #4CAF50; }
        .change-new { color: #27ae60; font-weight: bold; }
        .change-update { color: #3498db; font-weight: bold; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .highlight { color: #4CAF50; font-weight: bold; }
        h4 { margin: 10px 0 5px 0; color: #2c3e50; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>成绩更新通知</h1>
        </div>
        <div class="content">
            <p>亲爱的 <span class="highlight">{{.UserName}}</span>，</p>
            <p>您有新的成绩更新，详情如下：</p>
            
            {{range .ScoreUpdates}}
            <div class="score-item">
                <h3>{{.CourseName}}</h3>
                <div class="course-info">
                    <p><strong>课程代码：</strong>{{.CourseCode}}</p>
                    <p><strong>学期：</strong>{{.Semester}}</p>
                    {{if .TeacherNames}}<p><strong>任课教师：</strong>{{.TeacherNames}}</p>{{end}}
                    {{if .CourseProperty}}<p><strong>课程性质：</strong>{{.CourseProperty}}</p>{{end}}
                    <p><strong>学分：</strong>{{.Credit}} | <strong>绩点：</strong>{{.GPA}}</p>
                    {{if .TestNote}}<p><strong>考试状态：</strong>{{.TestNote}}</p>{{end}}
                </div>

                <div class="score-details">
                    <h4>详细成绩信息：</h4>
                    <table class="score-table">
                        <tr>
                            <td><strong>平时分：</strong></td>
                            <td>{{if ne .DailyScore "0"}}{{.DailyScore}}{{else}}-{{end}}</td>
                            <td><strong>考试分：</strong></td>
                            <td>{{if ne .ExamScore "0"}}{{.ExamScore}}{{else}}-{{end}}</td>
                        </tr>
                        <tr>
                            <td><strong>最终分：</strong></td>
                            <td class="final-score">{{if ne .FinalScore "0"}}{{.FinalScore}}{{else}}-{{end}}</td>
                            <td><strong>实践分：</strong></td>
                            <td>{{if ne .PracticalScore "0"}}{{.PracticalScore}}{{else}}-{{end}}</td>
                        </tr>
                        {{if ne .SupplementaryScore "0"}}
                        <tr>
                            <td><strong>补考分：</strong></td>
                            <td colspan="3" class="supplementary-score">{{.SupplementaryScore}}</td>
                        </tr>
                        {{end}}
                    </table>
                </div>

                <div class="change-info">
                    <p><strong>变更类型：</strong>
                        {{if eq .ChangeType "new"}}
                            <span class="change-new">新增成绩</span>
                        {{else}}
                            <span class="change-update">成绩更新</span>
                        {{end}}
                    </p>
                    <p><strong>变更详情：</strong>{{.ScoreType}}
                        {{if .OldScore}}
                            {{.OldScore}} → <span class="highlight">{{.NewScore}}</span>
                        {{else}}
                            <span class="highlight">{{.NewScore}}</span>
                        {{end}}
                    </p>
                    <p><strong>更新时间：</strong>{{.UpdateTime}}</p>
                </div>
            </div>
            {{end}}
            
            <p>请及时登录系统查看详细信息。</p>
        </div>
        <div class="footer">
            <p>此邮件由{{.SystemName}}自动发送，请勿回复。</p>
            <p>发送时间：{{.UpdateTime}}</p>
        </div>
    </div>
</body>
</html>
`
}

// getSystemEmailTemplate 获取系统邮件模板
func getSystemEmailTemplate() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
        .highlight { color: #2196F3; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Subject}}</h1>
        </div>
        <div class="content">
            <p>亲爱的 <span class="highlight">{{.UserName}}</span>，</p>
            <div style="background-color: white; padding: 20px; margin: 20px 0; border-radius: 5px;">
                {{.Content}}
            </div>
        </div>
        <div class="footer">
            <p>此邮件由{{.SystemName}}自动发送，请勿回复。</p>
            <p>发送时间：{{.SendTime}}</p>
        </div>
    </div>
</body>
</html>
`
}

// SendEmailWithTemplate 使用模板发送邮件
func (e *EmailService) SendEmailWithTemplate(userID uint, templateType config.EmailTemplateType, params map[string]interface{}) error {
	if !e.IsEnabled() {
		return fmt.Errorf("邮件服务未启用")
	}

	// 获取模板配置
	emailTemplate, err := e.templateManager.GetTemplate(templateType)
	if err != nil {
		return fmt.Errorf("获取邮件模板失败: %w", err)
	}

	// 验证参数
	if err := e.templateManager.ValidateParams(templateType, params); err != nil {
		return fmt.Errorf("模板参数验证失败: %w", err)
	}

	// 获取用户信息（用于日志记录）
	_, err = models.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取通知配置失败: %w", err)
	}

	if !channel.EmailEnabled || channel.EmailAddress == "" {
		return fmt.Errorf("用户未启用邮件通知或未配置邮箱地址")
	}

	// 解析HTML模板
	htmlTmpl, err := template.New("email_html").Parse(emailTemplate.HTMLBody)
	if err != nil {
		return fmt.Errorf("解析HTML邮件模板失败: %w", err)
	}

	// 解析文本模板
	textTmpl, err := template.New("email_text").Parse(emailTemplate.TextBody)
	if err != nil {
		return fmt.Errorf("解析文本邮件模板失败: %w", err)
	}

	// 渲染HTML模板
	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, params); err != nil {
		return fmt.Errorf("渲染HTML邮件模板失败: %w", err)
	}

	// 渲染文本模板
	var textBuf bytes.Buffer
	if err := textTmpl.Execute(&textBuf, params); err != nil {
		return fmt.Errorf("渲染文本邮件模板失败: %w", err)
	}

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelEmail,
		Type:      string(templateType),
		Title:     emailTemplate.Subject,
		Content:   textBuf.String(),
		Recipient: channel.EmailAddress,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建邮件通知日志失败: %v", err)
	}

	// 发送邮件
	err = e.SendMultipartEmail(channel.EmailAddress, emailTemplate.Subject, htmlBuf.String(), textBuf.String())

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
		log.Printf("更新邮件通知日志失败: %v", updateErr)
	}

	return err
}

// SendMultipartEmail 发送多部分邮件（HTML + 文本）
func (e *EmailService) SendMultipartEmail(to, subject, htmlBody, textBody string) error {
	if !e.IsEnabled() {
		return fmt.Errorf("邮件服务未启用")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.GetEnv("SMTP_USERNAME", ""))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	// 设置多部分内容
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	return e.dialer.DialAndSend(m)
}

// GetTemplateManager 获取模板管理器
func (e *EmailService) GetTemplateManager() *config.EmailTemplateManager {
	return e.templateManager
}

// GetAllTemplates 获取所有可用的邮件模板
func (e *EmailService) GetAllTemplates() map[config.EmailTemplateType]*config.EmailTemplateConfig {
	return e.templateManager.GetAllTemplates()
}

// GetEnabledTemplates 获取所有启用的邮件模板
func (e *EmailService) GetEnabledTemplates() map[config.EmailTemplateType]*config.EmailTemplateConfig {
	return e.templateManager.GetEnabledTemplates()
}

// ValidateTemplateParams 验证模板参数
func (e *EmailService) ValidateTemplateParams(templateType config.EmailTemplateType, params map[string]interface{}) error {
	return e.templateManager.ValidateParams(templateType, params)
}
