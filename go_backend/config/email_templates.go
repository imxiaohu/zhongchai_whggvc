package config

import (
	"fmt"
	"strings"
	"time"
)

// EmailTemplateType 邮件模板类型
type EmailTemplateType string

const (
	// 成绩相关模板
	EmailTemplateScoreUpdate    EmailTemplateType = "score_update"    // 成绩更新通知
	EmailTemplateScoreReminder  EmailTemplateType = "score_reminder"  // 成绩查询提醒
	EmailTemplateGradePublished EmailTemplateType = "grade_published" // 成绩发布通知

	// 系统相关模板
	EmailTemplateSystemNotice      EmailTemplateType = "system_notice"      // 系统通知
	EmailTemplateMaintenanceNotice EmailTemplateType = "maintenance_notice" // 维护通知
	EmailTemplateSecurityAlert     EmailTemplateType = "security_alert"     // 安全警告

	// 账户相关模板
	EmailTemplateAccountBinding EmailTemplateType = "account_binding" // 账户绑定
	EmailTemplatePasswordReset  EmailTemplateType = "password_reset"  // 密码重置
	EmailTemplateLoginAlert     EmailTemplateType = "login_alert"     // 登录提醒

	// 课程相关模板
	EmailTemplateCourseReminder EmailTemplateType = "course_reminder" // 课程提醒
	EmailTemplateExamNotice     EmailTemplateType = "exam_notice"     // 考试通知
	EmailTemplateAssignmentDue  EmailTemplateType = "assignment_due"  // 作业截止提醒

	// 评教相关模板
	EmailTemplateEvaluationReminder EmailTemplateType = "evaluation_reminder" // 评教提醒
	EmailTemplateEvaluationDeadline EmailTemplateType = "evaluation_deadline" // 评教截止提醒

	// 社区相关模板
	EmailTemplateCommunityReply   EmailTemplateType = "community_reply"   // 社区回复通知
	EmailTemplateCommunityMention EmailTemplateType = "community_mention" // 社区@提醒

	// 充值相关模板
	EmailTemplateRechargeSuccess EmailTemplateType = "recharge_success" // 充值成功
	EmailTemplateBalanceLow      EmailTemplateType = "balance_low"      // 余额不足提醒

	// 验证码模板
	EmailTemplateVerificationCode EmailTemplateType = "verification_code" // 验证码

	// 报告模板
	EmailTemplateWeeklyReport  EmailTemplateType = "weekly_report"  // 周报
	EmailTemplateMonthlyReport EmailTemplateType = "monthly_report" // 月报
)

// EmailTemplateConfig 邮件模板配置
type EmailTemplateConfig struct {
	Subject     string   `json:"subject"`     // 邮件主题
	HTMLBody    string   `json:"htmlBody"`    // HTML邮件内容
	TextBody    string   `json:"textBody"`    // 纯文本邮件内容
	Description string   `json:"description"` // 模板描述
	Params      []string `json:"params"`      // 模板参数列表
	Category    string   `json:"category"`    // 模板分类
	Enabled     bool     `json:"enabled"`     // 是否启用
}

// EmailTemplateManager 邮件模板管理器
type EmailTemplateManager struct {
	templates map[EmailTemplateType]*EmailTemplateConfig
}

// NewEmailTemplateManager 创建邮件模板管理器
func NewEmailTemplateManager() *EmailTemplateManager {
	manager := &EmailTemplateManager{
		templates: make(map[EmailTemplateType]*EmailTemplateConfig),
	}
	manager.initTemplates()
	return manager
}

// initTemplates 初始化所有邮件模板
func (m *EmailTemplateManager) initTemplates() {
	// 成绩更新通知模板
	m.templates[EmailTemplateScoreUpdate] = &EmailTemplateConfig{
		Subject:     "📊 成绩更新通知",
		HTMLBody:    getScoreUpdateHTMLTemplate(),
		TextBody:    getScoreUpdateTextTemplate(),
		Description: "成绩更新通知邮件",
		Params:      []string{"UserName", "ScoreUpdates", "UpdateTime", "SystemName"},
		Category:    "成绩相关",
		Enabled:     true,
	}

	// 成绩查询提醒模板
	m.templates[EmailTemplateScoreReminder] = &EmailTemplateConfig{
		Subject:     "📋 成绩查询提醒",
		HTMLBody:    getScoreReminderHTMLTemplate(),
		TextBody:    getScoreReminderTextTemplate(),
		Description: "成绩查询提醒邮件",
		Params:      []string{"UserName", "SystemName", "SendTime"},
		Category:    "成绩相关",
		Enabled:     true,
	}

	// 成绩发布通知模板
	m.templates[EmailTemplateGradePublished] = &EmailTemplateConfig{
		Subject:     "🎓 成绩发布通知",
		HTMLBody:    getGradePublishedHTMLTemplate(),
		TextBody:    getGradePublishedTextTemplate(),
		Description: "成绩发布通知邮件",
		Params:      []string{"UserName", "Semester", "CourseName", "SystemName", "SendTime"},
		Category:    "成绩相关",
		Enabled:     true,
	}

	// 系统通知模板
	m.templates[EmailTemplateSystemNotice] = &EmailTemplateConfig{
		Subject:     "📢 系统通知",
		HTMLBody:    getSystemNoticeHTMLTemplate(),
		TextBody:    getSystemNoticeTextTemplate(),
		Description: "系统通知邮件",
		Params:      []string{"UserName", "Subject", "Content", "SystemName", "SendTime"},
		Category:    "系统相关",
		Enabled:     true,
	}

	// 维护通知模板
	m.templates[EmailTemplateMaintenanceNotice] = &EmailTemplateConfig{
		Subject:     "🔧 系统维护通知",
		HTMLBody:    getMaintenanceNoticeHTMLTemplate(),
		TextBody:    getMaintenanceNoticeTextTemplate(),
		Description: "系统维护通知邮件",
		Params:      []string{"UserName", "MaintenanceTime", "Duration", "SystemName", "SendTime"},
		Category:    "系统相关",
		Enabled:     true,
	}

	// 安全警告模板
	m.templates[EmailTemplateSecurityAlert] = &EmailTemplateConfig{
		Subject:     "🔒 安全警告",
		HTMLBody:    getSecurityAlertHTMLTemplate(),
		TextBody:    getSecurityAlertTextTemplate(),
		Description: "安全警告邮件",
		Params:      []string{"UserName", "LoginTime", "Location", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
	}

	// 账户绑定模板
	m.templates[EmailTemplateAccountBinding] = &EmailTemplateConfig{
		Subject:     "✅ 账户绑定成功",
		HTMLBody:    getAccountBindingHTMLTemplate(),
		TextBody:    getAccountBindingTextTemplate(),
		Description: "账户绑定成功邮件",
		Params:      []string{"UserName", "Username", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
	}

	// 密码重置模板
	m.templates[EmailTemplatePasswordReset] = &EmailTemplateConfig{
		Subject:     "🔑 密码重置通知",
		HTMLBody:    getPasswordResetHTMLTemplate(),
		TextBody:    getPasswordResetTextTemplate(),
		Description: "密码重置通知邮件",
		Params:      []string{"UserName", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
	}

	// 登录提醒模板
	m.templates[EmailTemplateLoginAlert] = &EmailTemplateConfig{
		Subject:     "🔐 登录提醒",
		HTMLBody:    getLoginAlertHTMLTemplate(),
		TextBody:    getLoginAlertTextTemplate(),
		Description: "登录提醒邮件",
		Params:      []string{"UserName", "LoginTime", "Device", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
	}

	// 课程提醒模板
	m.templates[EmailTemplateCourseReminder] = &EmailTemplateConfig{
		Subject:     "📚 课程提醒",
		HTMLBody:    getCourseReminderHTMLTemplate(),
		TextBody:    getCourseReminderTextTemplate(),
		Description: "课程提醒邮件",
		Params:      []string{"UserName", "CourseTime", "CourseName", "Location", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
	}

	// 考试通知模板
	m.templates[EmailTemplateExamNotice] = &EmailTemplateConfig{
		Subject:     "📝 考试通知",
		HTMLBody:    getExamNoticeHTMLTemplate(),
		TextBody:    getExamNoticeTextTemplate(),
		Description: "考试通知邮件",
		Params:      []string{"UserName", "CourseName", "ExamTime", "Location", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
	}

	// 作业截止提醒模板
	m.templates[EmailTemplateAssignmentDue] = &EmailTemplateConfig{
		Subject:     "📋 作业截止提醒",
		HTMLBody:    getAssignmentDueHTMLTemplate(),
		TextBody:    getAssignmentDueTextTemplate(),
		Description: "作业截止提醒邮件",
		Params:      []string{"UserName", "CourseName", "AssignmentName", "Deadline", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
	}

	// 评教提醒模板
	m.templates[EmailTemplateEvaluationReminder] = &EmailTemplateConfig{
		Subject:     "⭐ 评教提醒",
		HTMLBody:    getEvaluationReminderHTMLTemplate(),
		TextBody:    getEvaluationReminderTextTemplate(),
		Description: "评教提醒邮件",
		Params:      []string{"UserName", "Semester", "CourseCount", "SystemName", "SendTime"},
		Category:    "评教相关",
		Enabled:     true,
	}

	// 评教截止提醒模板
	m.templates[EmailTemplateEvaluationDeadline] = &EmailTemplateConfig{
		Subject:     "⏰ 评教截止提醒",
		HTMLBody:    getEvaluationDeadlineHTMLTemplate(),
		TextBody:    getEvaluationDeadlineTextTemplate(),
		Description: "评教截止提醒邮件",
		Params:      []string{"UserName", "Deadline", "RemainingCount", "SystemName", "SendTime"},
		Category:    "评教相关",
		Enabled:     true,
	}

	// 社区回复通知模板
	m.templates[EmailTemplateCommunityReply] = &EmailTemplateConfig{
		Subject:     "💬 社区回复通知",
		HTMLBody:    getCommunityReplyHTMLTemplate(),
		TextBody:    getCommunityReplyTextTemplate(),
		Description: "社区回复通知邮件",
		Params:      []string{"UserName", "ReplyUser", "PostTitle", "ReplyContent", "SystemName", "SendTime"},
		Category:    "社区相关",
		Enabled:     true,
	}

	// 社区@提醒模板
	m.templates[EmailTemplateCommunityMention] = &EmailTemplateConfig{
		Subject:     "📢 社区@提醒",
		HTMLBody:    getCommunityMentionHTMLTemplate(),
		TextBody:    getCommunityMentionTextTemplate(),
		Description: "社区@提醒邮件",
		Params:      []string{"UserName", "MentionUser", "PostTitle", "Content", "SystemName", "SendTime"},
		Category:    "社区相关",
		Enabled:     true,
	}

	// 充值成功模板
	m.templates[EmailTemplateRechargeSuccess] = &EmailTemplateConfig{
		Subject:     "💰 充值成功通知",
		HTMLBody:    getRechargeSuccessHTMLTemplate(),
		TextBody:    getRechargeSuccessTextTemplate(),
		Description: "充值成功通知邮件",
		Params:      []string{"UserName", "Amount", "Balance", "SystemName", "SendTime"},
		Category:    "充值相关",
		Enabled:     true,
	}

	// 余额不足提醒模板
	m.templates[EmailTemplateBalanceLow] = &EmailTemplateConfig{
		Subject:     "⚠️ 余额不足提醒",
		HTMLBody:    getBalanceLowHTMLTemplate(),
		TextBody:    getBalanceLowTextTemplate(),
		Description: "余额不足提醒邮件",
		Params:      []string{"UserName", "Balance", "SystemName", "SendTime"},
		Category:    "充值相关",
		Enabled:     true,
	}

	// 验证码模板
	m.templates[EmailTemplateVerificationCode] = &EmailTemplateConfig{
		Subject:     "🔐 验证码",
		HTMLBody:    getVerificationCodeHTMLTemplate(),
		TextBody:    getVerificationCodeTextTemplate(),
		Description: "验证码邮件",
		Params:      []string{"UserName", "Code", "Minutes", "SystemName", "SendTime"},
		Category:    "验证码",
		Enabled:     true,
	}

	// 周报模板
	m.templates[EmailTemplateWeeklyReport] = &EmailTemplateConfig{
		Subject:     "📊 周报",
		HTMLBody:    getWeeklyReportHTMLTemplate(),
		TextBody:    getWeeklyReportTextTemplate(),
		Description: "周报邮件",
		Params:      []string{"UserName", "WeekRange", "ReportData", "SystemName", "SendTime"},
		Category:    "报告相关",
		Enabled:     true,
	}

	// 月报模板
	m.templates[EmailTemplateMonthlyReport] = &EmailTemplateConfig{
		Subject:     "📈 月报",
		HTMLBody:    getMonthlyReportHTMLTemplate(),
		TextBody:    getMonthlyReportTextTemplate(),
		Description: "月报邮件",
		Params:      []string{"UserName", "Month", "ReportData", "SystemName", "SendTime"},
		Category:    "报告相关",
		Enabled:     true,
	}
}

// GetTemplate 获取指定类型的邮件模板
func (m *EmailTemplateManager) GetTemplate(templateType EmailTemplateType) (*EmailTemplateConfig, error) {
	template, exists := m.templates[templateType]
	if !exists {
		return nil, fmt.Errorf("邮件模板不存在: %s", templateType)
	}

	if !template.Enabled {
		return nil, fmt.Errorf("邮件模板已禁用: %s", templateType)
	}

	return template, nil
}

// ValidateParams 验证模板参数
func (m *EmailTemplateManager) ValidateParams(templateType EmailTemplateType, params map[string]interface{}) error {
	template, err := m.GetTemplate(templateType)
	if err != nil {
		return err
	}

	// 检查必需参数
	for _, requiredParam := range template.Params {
		if _, exists := params[requiredParam]; !exists {
			return fmt.Errorf("缺少必需参数: %s", requiredParam)
		}
	}

	return nil
}

// GetAllTemplates 获取所有模板
func (m *EmailTemplateManager) GetAllTemplates() map[EmailTemplateType]*EmailTemplateConfig {
	return m.templates
}

// GetEnabledTemplates 获取所有启用的模板
func (m *EmailTemplateManager) GetEnabledTemplates() map[EmailTemplateType]*EmailTemplateConfig {
	enabled := make(map[EmailTemplateType]*EmailTemplateConfig)
	for templateType, template := range m.templates {
		if template.Enabled {
			enabled[templateType] = template
		}
	}
	return enabled
}

// GetTemplatesByCategory 根据分类获取模板
func (m *EmailTemplateManager) GetTemplatesByCategory(category string) map[EmailTemplateType]*EmailTemplateConfig {
	result := make(map[EmailTemplateType]*EmailTemplateConfig)
	for templateType, template := range m.templates {
		if template.Category == category && template.Enabled {
			result[templateType] = template
		}
	}
	return result
}

// UpdateTemplate 更新模板配置
func (m *EmailTemplateManager) UpdateTemplate(templateType EmailTemplateType, config *EmailTemplateConfig) error {
	if _, exists := m.templates[templateType]; !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	m.templates[templateType] = config
	return nil
}

// EnableTemplate 启用模板
func (m *EmailTemplateManager) EnableTemplate(templateType EmailTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = true
	return nil
}

// DisableTemplate 禁用模板
func (m *EmailTemplateManager) DisableTemplate(templateType EmailTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = false
	return nil
}

// FormatEmailTime 格式化邮件时间
func FormatEmailTime(t time.Time) string {
	return t.Format("2006年01月02日 15:04:05")
}

// FormatEmailDate 格式化邮件日期
func FormatEmailDate(t time.Time) string {
	return t.Format("2006年01月02日")
}

// BuildEmailCourseList 构建课程列表字符串（邮件格式）
func BuildEmailCourseList(courses []string) string {
	if len(courses) == 0 {
		return ""
	}

	if len(courses) == 1 {
		return courses[0]
	}

	return strings.Join(courses, "、")
}

// 全局邮件模板管理器实例
var globalEmailTemplateManager *EmailTemplateManager

// GetEmailTemplateManager 获取全局邮件模板管理器
func GetEmailTemplateManager() *EmailTemplateManager {
	if globalEmailTemplateManager == nil {
		globalEmailTemplateManager = NewEmailTemplateManager()
	}
	return globalEmailTemplateManager
}

// 便捷函数：获取特定模板
func GetScoreUpdateEmailTemplate() (*EmailTemplateConfig, error) {
	return GetEmailTemplateManager().GetTemplate(EmailTemplateScoreUpdate)
}

func GetSystemNoticeEmailTemplate() (*EmailTemplateConfig, error) {
	return GetEmailTemplateManager().GetTemplate(EmailTemplateSystemNotice)
}

func GetCourseReminderEmailTemplate() (*EmailTemplateConfig, error) {
	return GetEmailTemplateManager().GetTemplate(EmailTemplateCourseReminder)
}

func GetEvaluationReminderEmailTemplate() (*EmailTemplateConfig, error) {
	return GetEmailTemplateManager().GetTemplate(EmailTemplateEvaluationReminder)
}
