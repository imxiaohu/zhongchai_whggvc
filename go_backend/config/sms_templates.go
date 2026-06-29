package config

import (
	"fmt"
	"strings"
	"time"
)

// SMSTemplateType 短信模板类型
type SMSTemplateType string

const (
	// 成绩相关模板
	TemplateScoreUpdate             SMSTemplateType = "score_update"              // 成绩更新通知
	TemplateScoreReminder           SMSTemplateType = "score_reminder"            // 成绩查询提醒
	TemplateGradePublished          SMSTemplateType = "grade_published"           // 成绩发布通知
	TemplateScoreTestNotification   SMSTemplateType = "score_test_notification"   // 成绩检查测试通知
	TemplateScoreChangeNotification SMSTemplateType = "score_change_notification" // 成绩变化通知

	// 系统相关模板
	TemplateSystemNotice      SMSTemplateType = "system_notice"      // 系统通知
	TemplateMaintenanceNotice SMSTemplateType = "maintenance_notice" // 维护通知
	TemplateSecurityAlert     SMSTemplateType = "security_alert"     // 安全警告

	// 账户相关模板
	TemplateAccountBinding SMSTemplateType = "account_binding" // 账户绑定
	TemplatePasswordReset  SMSTemplateType = "password_reset"  // 密码重置
	TemplateLoginAlert     SMSTemplateType = "login_alert"     // 登录提醒

	// 课程相关模板
	TemplateCourseReminder SMSTemplateType = "course_reminder" // 课程提醒
	TemplateExamNotice     SMSTemplateType = "exam_notice"     // 考试通知
	TemplateAssignmentDue  SMSTemplateType = "assignment_due"  // 作业截止提醒

	// 评教相关模板
	TemplateEvaluationReminder SMSTemplateType = "evaluation_reminder" // 评教提醒
	TemplateEvaluationDeadline SMSTemplateType = "evaluation_deadline" // 评教截止提醒

	// 社区相关模板
	TemplateCommunityReply   SMSTemplateType = "community_reply"   // 社区回复通知
	TemplateCommunityMention SMSTemplateType = "community_mention" // 社区@提醒

	// 充值相关模板
	TemplateRechargeSuccess SMSTemplateType = "recharge_success" // 充值成功
	TemplateBalanceLow      SMSTemplateType = "balance_low"      // 余额不足提醒

	// 验证码模板
	TemplateVerificationCode SMSTemplateType = "verification_code" // 验证码
)

// SMSTemplateConfig 短信模板配置
type SMSTemplateConfig struct {
	TemplateCode string   `json:"templateCode"` // 阿里云模板代码
	SignName     string   `json:"signName"`     // 短信签名
	Description  string   `json:"description"`  // 模板描述
	Content      string   `json:"content"`      // 模板内容（用于显示）
	Params       []string `json:"params"`       // 模板参数列表
	MaxLength    int      `json:"maxLength"`    // 最大长度限制
	Enabled      bool     `json:"enabled"`      // 是否启用
}

// SMSTemplateManager 短信模板管理器
type SMSTemplateManager struct {
	templates map[SMSTemplateType]*SMSTemplateConfig
}

// NewSMSTemplateManager 创建短信模板管理器
func NewSMSTemplateManager() *SMSTemplateManager {
	manager := &SMSTemplateManager{
		templates: make(map[SMSTemplateType]*SMSTemplateConfig),
	}
	manager.initTemplates()
	return manager
}

// initTemplates 初始化所有短信模板
func (m *SMSTemplateManager) initTemplates() {
	signName := GetSMSSignName()

	// 成绩更新通知模板
	m.templates[TemplateScoreUpdate] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SCORE_UPDATE", "SMS_489480145"),
		SignName:     signName,
		Description:  "成绩更新通知",
		Content:      "亲爱的${name}同学您好，您设置的成绩检查服务已经获取到您的成绩有更新，请打开微信小程序-成绩查询查看您的期末成绩吧！",
		Params:       []string{"name"},
		MaxLength:    100,
		Enabled:      true,
	}

	// 成绩查询提醒模板
	m.templates[TemplateScoreReminder] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SCORE_REMINDER", "SMS_123456790"),
		SignName:     signName,
		Description:  "成绩查询提醒",
		Content:      "亲爱的${name}同学，您有新的成绩可以查询，请及时登录系统查看。【${signName}】",
		Params:       []string{"name"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 成绩发布通知模板
	m.templates[TemplateGradePublished] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_GRADE_PUBLISHED", "SMS_123456791"),
		SignName:     signName,
		Description:  "成绩发布通知",
		Content:      "亲爱的${name}同学，${semester}学期${course}课程成绩已发布，请及时查看。【${signName}】",
		Params:       []string{"name", "semester", "course"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 成绩检查测试通知模板 - 测试通过通知
	m.templates[TemplateScoreTestNotification] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SCORE_TEST", "SMS_489500152"),
		SignName:     signName,
		Description:  "测试通过通知",
		Content:      "恭喜您${name}，收到这条短信，代表您已经成功配置了 成绩通知！【${signName}】",
		Params:       []string{"name"},
		MaxLength:    150,
		Enabled:      true,
	}

	// 成绩变化通知模板
	m.templates[TemplateScoreChangeNotification] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SCORE_CHANGE", "SMS_489480145"),
		SignName:     signName,
		Description:  "成绩变化通知",
		Content:      "亲爱的的${name}同学您好，您设置的成绩检查服务已经获取到您的成绩有更新，请打卡微信小程序-成绩查询查看您的期末成绩吧！【${signName}】",
		Params:       []string{"name"},
		MaxLength:    150,
		Enabled:      true,
	}

	// 系统通知模板
	m.templates[TemplateSystemNotice] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SYSTEM_NOTICE", "SMS_123456792"),
		SignName:     signName,
		Description:  "系统通知",
		Content:      "亲爱的${name}同学，${content}。【${signName}】",
		Params:       []string{"name", "content"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 维护通知模板
	m.templates[TemplateMaintenanceNotice] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_MAINTENANCE", "SMS_123456793"),
		SignName:     signName,
		Description:  "系统维护通知",
		Content:      "系统将于${time}进行维护，预计持续${duration}，期间服务可能中断，请合理安排时间。【${signName}】",
		Params:       []string{"time", "duration"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 安全警告模板
	m.templates[TemplateSecurityAlert] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_SECURITY_ALERT", "SMS_123456794"),
		SignName:     signName,
		Description:  "安全警告",
		Content:      "检测到您的账户${time}在${location}有异常登录，如非本人操作请及时修改密码。【${signName}】",
		Params:       []string{"time", "location"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 账户绑定模板
	m.templates[TemplateAccountBinding] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_ACCOUNT_BINDING", "SMS_123456795"),
		SignName:     signName,
		Description:  "账户绑定通知",
		Content:      "您的手机号已成功绑定到${username}账户，如非本人操作请联系管理员。【${signName}】",
		Params:       []string{"username"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 密码重置模板
	m.templates[TemplatePasswordReset] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_PASSWORD_RESET", "SMS_123456796"),
		SignName:     signName,
		Description:  "密码重置通知",
		Content:      "您的账户密码已重置，请使用新密码登录并及时修改。如非本人操作请联系管理员。【${signName}】",
		Params:       []string{},
		MaxLength:    70,
		Enabled:      true,
	}

	// 登录提醒模板
	m.templates[TemplateLoginAlert] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_LOGIN_ALERT", "SMS_123456797"),
		SignName:     signName,
		Description:  "登录提醒",
		Content:      "您的账户于${time}在${device}设备登录，如非本人操作请及时修改密码。【${signName}】",
		Params:       []string{"time", "device"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 课程提醒模板
	m.templates[TemplateCourseReminder] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_COURSE_REMINDER", "SMS_123456798"),
		SignName:     signName,
		Description:  "课程提醒",
		Content:      "提醒：您今天${time}有${course}课程，地点：${location}，请准时参加。【${signName}】",
		Params:       []string{"time", "course", "location"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 考试通知模板
	m.templates[TemplateExamNotice] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_EXAM_NOTICE", "SMS_123456799"),
		SignName:     signName,
		Description:  "考试通知",
		Content:      "考试提醒：${course}将于${time}在${location}举行考试，请携带相关证件准时参加。【${signName}】",
		Params:       []string{"course", "time", "location"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 作业截止提醒模板
	m.templates[TemplateAssignmentDue] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_ASSIGNMENT_DUE", "SMS_123456800"),
		SignName:     signName,
		Description:  "作业截止提醒",
		Content:      "作业提醒：${course}的${assignment}将于${deadline}截止，请及时提交。【${signName}】",
		Params:       []string{"course", "assignment", "deadline"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 评教提醒模板
	m.templates[TemplateEvaluationReminder] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_EVALUATION_REMINDER", "SMS_123456801"),
		SignName:     signName,
		Description:  "评教提醒",
		Content:      "评教提醒：${semester}学期课程评教已开始，请及时完成${count}门课程的评教。【${signName}】",
		Params:       []string{"semester", "count"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 评教截止提醒模板
	m.templates[TemplateEvaluationDeadline] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_EVALUATION_DEADLINE", "SMS_123456802"),
		SignName:     signName,
		Description:  "评教截止提醒",
		Content:      "评教截止提醒：课程评教将于${deadline}截止，您还有${count}门课程未完成评教。【${signName}】",
		Params:       []string{"deadline", "count"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 社区回复通知模板
	m.templates[TemplateCommunityReply] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_COMMUNITY_REPLY", "SMS_123456803"),
		SignName:     signName,
		Description:  "社区回复通知",
		Content:      "${user}回复了您的帖子《${title}》，快来查看吧！【${signName}】",
		Params:       []string{"user", "title"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 社区@提醒模板
	m.templates[TemplateCommunityMention] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_COMMUNITY_MENTION", "SMS_123456804"),
		SignName:     signName,
		Description:  "社区@提醒",
		Content:      "${user}在帖子《${title}》中@了您，快来查看吧！【${signName}】",
		Params:       []string{"user", "title"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 充值成功模板
	m.templates[TemplateRechargeSuccess] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_RECHARGE_SUCCESS", "SMS_123456805"),
		SignName:     signName,
		Description:  "充值成功通知",
		Content:      "充值成功！您已充值${amount}元，当前短信余额${balance}元。【${signName}】",
		Params:       []string{"amount", "balance"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 余额不足提醒模板
	m.templates[TemplateBalanceLow] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_BALANCE_LOW", "SMS_123456806"),
		SignName:     signName,
		Description:  "余额不足提醒",
		Content:      "您的短信余额不足${balance}元，为避免影响通知接收，请及时充值。【${signName}】",
		Params:       []string{"balance"},
		MaxLength:    70,
		Enabled:      true,
	}

	// 验证码模板
	m.templates[TemplateVerificationCode] = &SMSTemplateConfig{
		TemplateCode: GetEnv("SMS_TEMPLATE_VERIFICATION_CODE", "SMS_123456807"),
		SignName:     signName,
		Description:  "验证码",
		Content:      "您的验证码是${code}，${minutes}分钟内有效，请勿泄露给他人。【${signName}】",
		Params:       []string{"code", "minutes"},
		MaxLength:    70,
		Enabled:      true,
	}
}

// GetTemplate 获取指定类型的短信模板
func (m *SMSTemplateManager) GetTemplate(templateType SMSTemplateType) (*SMSTemplateConfig, error) {
	template, exists := m.templates[templateType]
	if !exists {
		return nil, fmt.Errorf("短信模板不存在: %s", templateType)
	}

	if !template.Enabled {
		return nil, fmt.Errorf("短信模板已禁用: %s", templateType)
	}

	return template, nil
}

// ValidateParams 验证模板参数
func (m *SMSTemplateManager) ValidateParams(templateType SMSTemplateType, params map[string]string) error {
	template, err := m.GetTemplate(templateType)
	if err != nil {
		return err
	}

	// 检查必需参数
	for _, requiredParam := range template.Params {
		if value, exists := params[requiredParam]; !exists || value == "" {
			return fmt.Errorf("缺少必需参数: %s", requiredParam)
		}
	}

	// 检查内容长度
	content := m.ReplaceParams(template.Content, params)
	if len(content) > template.MaxLength {
		return fmt.Errorf("短信内容超过最大长度限制 %d 字符，当前长度: %d", template.MaxLength, len(content))
	}

	return nil
}

// ReplaceParams 替换模板参数
func (m *SMSTemplateManager) ReplaceParams(content string, params map[string]string) string {
	result := content
	for key, value := range params {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// GetAllTemplates 获取所有模板
func (m *SMSTemplateManager) GetAllTemplates() map[SMSTemplateType]*SMSTemplateConfig {
	return m.templates
}

// GetEnabledTemplates 获取所有启用的模板
func (m *SMSTemplateManager) GetEnabledTemplates() map[SMSTemplateType]*SMSTemplateConfig {
	enabled := make(map[SMSTemplateType]*SMSTemplateConfig)
	for templateType, template := range m.templates {
		if template.Enabled {
			enabled[templateType] = template
		}
	}
	return enabled
}

// UpdateTemplate 更新模板配置
func (m *SMSTemplateManager) UpdateTemplate(templateType SMSTemplateType, config *SMSTemplateConfig) error {
	if _, exists := m.templates[templateType]; !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	m.templates[templateType] = config
	return nil
}

// EnableTemplate 启用模板
func (m *SMSTemplateManager) EnableTemplate(templateType SMSTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = true
	return nil
}

// DisableTemplate 禁用模板
func (m *SMSTemplateManager) DisableTemplate(templateType SMSTemplateType) error {
	template, exists := m.templates[templateType]
	if !exists {
		return fmt.Errorf("模板不存在: %s", templateType)
	}

	template.Enabled = false
	return nil
}

// FormatTime 格式化时间为短信友好格式
func FormatTime(t time.Time) string {
	return t.Format("01月02日 15:04")
}

// FormatDate 格式化日期为短信友好格式
func FormatDate(t time.Time) string {
	return t.Format("01月02日")
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// 尝试在合适的位置截断
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}

	return s[:maxLen]
}

// BuildCourseList 构建课程列表字符串
func BuildCourseList(courses []string, maxCourses int) string {
	if len(courses) == 0 {
		return ""
	}

	if len(courses) == 1 {
		return courses[0]
	}

	if len(courses) <= maxCourses {
		return strings.Join(courses, "、")
	}

	// 显示前几门课程，然后加上"等X门课程"
	displayed := courses[:maxCourses-1]
	return strings.Join(displayed, "、") + fmt.Sprintf("等%d门课程", len(courses))
}

// 全局模板管理器实例
var globalTemplateManager *SMSTemplateManager

// GetSMSTemplateManager 获取全局短信模板管理器
func GetSMSTemplateManager() *SMSTemplateManager {
	if globalTemplateManager == nil {
		globalTemplateManager = NewSMSTemplateManager()
	}
	return globalTemplateManager
}

// 便捷函数：获取特定模板
func GetScoreUpdateTemplate() (*SMSTemplateConfig, error) {
	return GetSMSTemplateManager().GetTemplate(TemplateScoreUpdate)
}

func GetSystemNoticeTemplate() (*SMSTemplateConfig, error) {
	return GetSMSTemplateManager().GetTemplate(TemplateSystemNotice)
}

func GetVerificationCodeTemplate() (*SMSTemplateConfig, error) {
	return GetSMSTemplateManager().GetTemplate(TemplateVerificationCode)
}

func GetCourseReminderTemplate() (*SMSTemplateConfig, error) {
	return GetSMSTemplateManager().GetTemplate(TemplateCourseReminder)
}

func GetEvaluationReminderTemplate() (*SMSTemplateConfig, error) {
	return GetSMSTemplateManager().GetTemplate(TemplateEvaluationReminder)
}
