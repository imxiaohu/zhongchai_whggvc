package config

import (
	"fmt"
	"strings"
	"time"
)

// DingTalkTemplateType 钉钉模板类型
type DingTalkTemplateType string

const (
	// 成绩相关模板
	DingTalkTemplateScoreUpdate    DingTalkTemplateType = "score_update"    // 成绩更新通知
	DingTalkTemplateScoreReminder  DingTalkTemplateType = "score_reminder"  // 成绩查询提醒
	DingTalkTemplateGradePublished DingTalkTemplateType = "grade_published" // 成绩发布通知

	// 系统相关模板
	DingTalkTemplateSystemNotice      DingTalkTemplateType = "system_notice"      // 系统通知
	DingTalkTemplateMaintenanceNotice DingTalkTemplateType = "maintenance_notice" // 维护通知
	DingTalkTemplateSecurityAlert     DingTalkTemplateType = "security_alert"     // 安全警告

	// 账户相关模板
	DingTalkTemplateAccountBinding DingTalkTemplateType = "account_binding" // 账户绑定
	DingTalkTemplatePasswordReset  DingTalkTemplateType = "password_reset"  // 密码重置
	DingTalkTemplateLoginAlert     DingTalkTemplateType = "login_alert"     // 登录提醒

	// 课程相关模板
	DingTalkTemplateCourseReminder DingTalkTemplateType = "course_reminder" // 课程提醒
	DingTalkTemplateExamNotice     DingTalkTemplateType = "exam_notice"     // 考试通知
	DingTalkTemplateAssignmentDue  DingTalkTemplateType = "assignment_due"  // 作业截止提醒

	// 评教相关模板
	DingTalkTemplateEvaluationReminder DingTalkTemplateType = "evaluation_reminder" // 评教提醒
	DingTalkTemplateEvaluationDeadline DingTalkTemplateType = "evaluation_deadline" // 评教截止提醒

	// 社区相关模板
	DingTalkTemplateCommunityReply   DingTalkTemplateType = "community_reply"   // 社区回复通知
	DingTalkTemplateCommunityMention DingTalkTemplateType = "community_mention" // 社区@提醒

	// 充值相关模板
	DingTalkTemplateRechargeSuccess DingTalkTemplateType = "recharge_success" // 充值成功
	DingTalkTemplateBalanceLow      DingTalkTemplateType = "balance_low"      // 余额不足提醒

	// 报告模板
	DingTalkTemplateWeeklyReport  DingTalkTemplateType = "weekly_report"  // 周报
	DingTalkTemplateMonthlyReport DingTalkTemplateType = "monthly_report" // 月报
)

// DingTalkMessageType 钉钉消息类型
type DingTalkMessageType string

const (
	DingTalkMessageTypeText     DingTalkMessageType = "text"     // 文本消息
	DingTalkMessageTypeMarkdown DingTalkMessageType = "markdown" // Markdown消息
)

// DingTalkTemplateConfig 钉钉模板配置
type DingTalkTemplateConfig struct {
	Title       string              `json:"title"`       // 消息标题
	Content     string              `json:"content"`     // 消息内容
	MessageType DingTalkMessageType `json:"messageType"` // 消息类型
	Description string              `json:"description"` // 模板描述
	Params      []string            `json:"params"`      // 模板参数列表
	Category    string              `json:"category"`    // 模板分类
	Enabled     bool                `json:"enabled"`     // 是否启用
	AtAll       bool                `json:"atAll"`       // 是否@所有人
	AtMobiles   []string            `json:"atMobiles"`   // @指定手机号
}

// DingTalkTemplateManager 钉钉模板管理器
type DingTalkTemplateManager struct {
	templates map[DingTalkTemplateType]*DingTalkTemplateConfig
}

// NewDingTalkTemplateManager 创建钉钉模板管理器
func NewDingTalkTemplateManager() *DingTalkTemplateManager {
	manager := &DingTalkTemplateManager{
		templates: make(map[DingTalkTemplateType]*DingTalkTemplateConfig),
	}
	manager.initTemplates()
	return manager
}

// initTemplates 初始化所有钉钉模板
func (m *DingTalkTemplateManager) initTemplates() {
	// 成绩更新通知模板
	m.templates[DingTalkTemplateScoreUpdate] = &DingTalkTemplateConfig{
		Title:       "📊 成绩更新通知",
		Content:     getScoreUpdateDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "成绩更新通知",
		Params:      []string{"UserName", "ScoreUpdates", "UpdateTime", "SystemName"},
		Category:    "成绩相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 成绩查询提醒模板
	m.templates[DingTalkTemplateScoreReminder] = &DingTalkTemplateConfig{
		Title:       "📋 成绩查询提醒",
		Content:     getScoreReminderDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "成绩查询提醒",
		Params:      []string{"UserName", "SystemName", "SendTime"},
		Category:    "成绩相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 成绩发布通知模板
	m.templates[DingTalkTemplateGradePublished] = &DingTalkTemplateConfig{
		Title:       "🎓 成绩发布通知",
		Content:     getGradePublishedDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "成绩发布通知",
		Params:      []string{"UserName", "Semester", "CourseName", "SystemName", "SendTime"},
		Category:    "成绩相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 系统通知模板
	m.templates[DingTalkTemplateSystemNotice] = &DingTalkTemplateConfig{
		Title:       "📢 系统通知",
		Content:     getSystemNoticeDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "系统通知",
		Params:      []string{"UserName", "Subject", "Content", "SystemName", "SendTime"},
		Category:    "系统相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 维护通知模板
	m.templates[DingTalkTemplateMaintenanceNotice] = &DingTalkTemplateConfig{
		Title:       "🔧 系统维护通知",
		Content:     getMaintenanceNoticeDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "系统维护通知",
		Params:      []string{"UserName", "MaintenanceTime", "Duration", "SystemName", "SendTime"},
		Category:    "系统相关",
		Enabled:     true,
		AtAll:       true, // 维护通知@所有人
	}

	// 安全警告模板
	m.templates[DingTalkTemplateSecurityAlert] = &DingTalkTemplateConfig{
		Title:       "🔒 安全警告",
		Content:     getSecurityAlertDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "安全警告",
		Params:      []string{"UserName", "LoginTime", "Location", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 账户绑定模板
	m.templates[DingTalkTemplateAccountBinding] = &DingTalkTemplateConfig{
		Title:       "✅ 账户绑定成功",
		Content:     getAccountBindingDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "账户绑定成功",
		Params:      []string{"UserName", "Username", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 密码重置模板
	m.templates[DingTalkTemplatePasswordReset] = &DingTalkTemplateConfig{
		Title:       "🔑 密码重置通知",
		Content:     getPasswordResetDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "密码重置通知",
		Params:      []string{"UserName", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 登录提醒模板
	m.templates[DingTalkTemplateLoginAlert] = &DingTalkTemplateConfig{
		Title:       "🔐 登录提醒",
		Content:     getLoginAlertDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "登录提醒",
		Params:      []string{"UserName", "LoginTime", "Device", "SystemName", "SendTime"},
		Category:    "账户相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 课程提醒模板
	m.templates[DingTalkTemplateCourseReminder] = &DingTalkTemplateConfig{
		Title:       "📚 课程提醒",
		Content:     getCourseReminderDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "课程提醒",
		Params:      []string{"UserName", "CourseTime", "CourseName", "Location", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 考试通知模板
	m.templates[DingTalkTemplateExamNotice] = &DingTalkTemplateConfig{
		Title:       "📝 考试通知",
		Content:     getExamNoticeDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "考试通知",
		Params:      []string{"UserName", "CourseName", "ExamTime", "Location", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 作业截止提醒模板
	m.templates[DingTalkTemplateAssignmentDue] = &DingTalkTemplateConfig{
		Title:       "📋 作业截止提醒",
		Content:     getAssignmentDueDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "作业截止提醒",
		Params:      []string{"UserName", "CourseName", "AssignmentName", "Deadline", "SystemName", "SendTime"},
		Category:    "课程相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 评教提醒模板
	m.templates[DingTalkTemplateEvaluationReminder] = &DingTalkTemplateConfig{
		Title:       "⭐ 评教提醒",
		Content:     getEvaluationReminderDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "评教提醒",
		Params:      []string{"UserName", "Semester", "CourseCount", "SystemName", "SendTime"},
		Category:    "评教相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 评教截止提醒模板
	m.templates[DingTalkTemplateEvaluationDeadline] = &DingTalkTemplateConfig{
		Title:       "⏰ 评教截止提醒",
		Content:     getEvaluationDeadlineDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "评教截止提醒",
		Params:      []string{"UserName", "Deadline", "RemainingCount", "SystemName", "SendTime"},
		Category:    "评教相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 社区回复通知模板
	m.templates[DingTalkTemplateCommunityReply] = &DingTalkTemplateConfig{
		Title:       "💬 社区回复通知",
		Content:     getCommunityReplyDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "社区回复通知",
		Params:      []string{"UserName", "ReplyUser", "PostTitle", "ReplyContent", "SystemName", "SendTime"},
		Category:    "社区相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 社区@提醒模板
	m.templates[DingTalkTemplateCommunityMention] = &DingTalkTemplateConfig{
		Title:       "📢 社区@提醒",
		Content:     getCommunityMentionDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "社区@提醒",
		Params:      []string{"UserName", "MentionUser", "PostTitle", "Content", "SystemName", "SendTime"},
		Category:    "社区相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 充值成功模板
	m.templates[DingTalkTemplateRechargeSuccess] = &DingTalkTemplateConfig{
		Title:       "💰 充值成功通知",
		Content:     getRechargeSuccessDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "充值成功通知",
		Params:      []string{"UserName", "Amount", "Balance", "SystemName", "SendTime"},
		Category:    "充值相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 余额不足提醒模板
	m.templates[DingTalkTemplateBalanceLow] = &DingTalkTemplateConfig{
		Title:       "⚠️ 余额不足提醒",
		Content:     getBalanceLowDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "余额不足提醒",
		Params:      []string{"UserName", "Balance", "SystemName", "SendTime"},
		Category:    "充值相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 周报模板
	m.templates[DingTalkTemplateWeeklyReport] = &DingTalkTemplateConfig{
		Title:       "📊 周报",
		Content:     getWeeklyReportDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "周报",
		Params:      []string{"UserName", "WeekRange", "ReportData", "SystemName", "SendTime"},
		Category:    "报告相关",
		Enabled:     true,
		AtAll:       false,
	}

	// 月报模板
	m.templates[DingTalkTemplateMonthlyReport] = &DingTalkTemplateConfig{
		Title:       "📈 月报",
		Content:     getMonthlyReportDingTalkTemplate(),
		MessageType: DingTalkMessageTypeMarkdown,
		Description: "月报",
		Params:      []string{"UserName", "Month", "ReportData", "SystemName", "SendTime"},
		Category:    "报告相关",
		Enabled:     true,
		AtAll:       false,
	}
}

// GetTemplate 获取指定类型的钉钉模板
func (m *DingTalkTemplateManager) GetTemplate(templateType DingTalkTemplateType) (*DingTalkTemplateConfig, error) {
	template, exists := m.templates[templateType]
	if !exists {
		return nil, fmt.Errorf("钉钉模板不存在: %s", templateType)
	}

	if !template.Enabled {
		return nil, fmt.Errorf("钉钉模板已禁用: %s", templateType)
	}

	return template, nil
}

// ValidateParams 验证模板参数
func (m *DingTalkTemplateManager) ValidateParams(templateType DingTalkTemplateType, params map[string]interface{}) error {
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

// ReplaceParams 替换模板参数
func (m *DingTalkTemplateManager) ReplaceParams(content string, params map[string]interface{}) string {
	result := content
	for key, value := range params {
		placeholder := "{{." + key + "}}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// GetAllTemplates 获取所有模板
func (m *DingTalkTemplateManager) GetAllTemplates() map[DingTalkTemplateType]*DingTalkTemplateConfig {
	return m.templates
}

// GetEnabledTemplates 获取所有启用的模板
func (m *DingTalkTemplateManager) GetEnabledTemplates() map[DingTalkTemplateType]*DingTalkTemplateConfig {
	enabled := make(map[DingTalkTemplateType]*DingTalkTemplateConfig)
	for templateType, template := range m.templates {
		if template.Enabled {
			enabled[templateType] = template
		}
	}
	return enabled
}

// GetTemplatesByCategory 根据分类获取模板
func (m *DingTalkTemplateManager) GetTemplatesByCategory(category string) map[DingTalkTemplateType]*DingTalkTemplateConfig {
	result := make(map[DingTalkTemplateType]*DingTalkTemplateConfig)
	for templateType, template := range m.templates {
		if template.Category == category && template.Enabled {
			result[templateType] = template
		}
	}
	return result
}

// UpdateTemplate 更新模板配置
func (m *DingTalkTemplateManager) UpdateTemplate(templateType DingTalkTemplateType, config *DingTalkTemplateConfig) error {
	if _, exists := m.templates[templateType]; !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	m.templates[templateType] = config
	return nil
}

// EnableTemplate 启用模板
func (m *DingTalkTemplateManager) EnableTemplate(templateType DingTalkTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = true
	return nil
}

// DisableTemplate 禁用模板
func (m *DingTalkTemplateManager) DisableTemplate(templateType DingTalkTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = false
	return nil
}

// FormatDingTalkTime 格式化钉钉时间
func FormatDingTalkTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDingTalkDate 格式化钉钉日期
func FormatDingTalkDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// BuildDingTalkCourseList 构建课程列表字符串（钉钉格式）
func BuildDingTalkCourseList(courses []string) string {
	if len(courses) == 0 {
		return ""
	}

	if len(courses) == 1 {
		return courses[0]
	}

	var result strings.Builder
	for i, course := range courses {
		if i > 0 {
			result.WriteString("\n- ")
		} else {
			result.WriteString("- ")
		}
		result.WriteString(course)
	}

	return result.String()
}

// 全局钉钉模板管理器实例
var globalDingTalkTemplateManager *DingTalkTemplateManager

// GetDingTalkTemplateManager 获取全局钉钉模板管理器
func GetDingTalkTemplateManager() *DingTalkTemplateManager {
	if globalDingTalkTemplateManager == nil {
		globalDingTalkTemplateManager = NewDingTalkTemplateManager()
	}
	return globalDingTalkTemplateManager
}

// 便捷函数：获取特定模板
func GetScoreUpdateDingTalkTemplate() (*DingTalkTemplateConfig, error) {
	return GetDingTalkTemplateManager().GetTemplate(DingTalkTemplateScoreUpdate)
}

func GetSystemNoticeDingTalkTemplate() (*DingTalkTemplateConfig, error) {
	return GetDingTalkTemplateManager().GetTemplate(DingTalkTemplateSystemNotice)
}

func GetCourseReminderDingTalkTemplate() (*DingTalkTemplateConfig, error) {
	return GetDingTalkTemplateManager().GetTemplate(DingTalkTemplateCourseReminder)
}

func GetEvaluationReminderDingTalkTemplate() (*DingTalkTemplateConfig, error) {
	return GetDingTalkTemplateManager().GetTemplate(DingTalkTemplateEvaluationReminder)
}
