package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
)

// SMSService 短信服务
type SMSService struct {
	client          *dysmsapi.Client
	enabled         bool
	templateManager *config.SMSTemplateManager
}

// SMSTemplate 短信模板（保持向后兼容）
type SMSTemplate struct {
	TemplateCode string
	SignName     string
	Params       map[string]string
}

var smsService *SMSService

// 短信费用常量（分）
const (
	SMSCostPerMessage = 10 // 0.1元/条
)

// InitSMSService 初始化短信服务
func InitSMSService() {
	smsService = NewSMSService()
}

// GetSMSService 获取短信服务实例
func GetSMSService() *SMSService {
	return smsService
}

// NewSMSService 创建新的短信服务
func NewSMSService() *SMSService {
	service := &SMSService{
		enabled:         config.GetSMSEnabled(),
		templateManager: config.GetSMSTemplateManager(),
	}

	if service.enabled {
		log.Printf("短信服务启用，开始初始化阿里云SMS客户端...")
		if err := service.initClient(); err != nil {
			log.Printf("初始化阿里云SMS客户端失败: %v", err)
			service.enabled = false
		} else {
			log.Printf("阿里云SMS客户端初始化成功")
		}
	} else {
		log.Printf("短信服务未启用")
	}

	return service
}

// initClient 初始化阿里云SMS客户端
func (s *SMSService) initClient() error {
	accessKeyId := config.GetAliyunAccessKeyID()
	accessKeySecret := config.GetAliyunAccessKeySecret()
	region := config.GetAliyunSMSRegion()

	if accessKeyId == "" || accessKeySecret == "" {
		return fmt.Errorf("阿里云AccessKey配置不完整")
	}

	client, err := dysmsapi.NewClientWithAccessKey(region, accessKeyId, accessKeySecret)
	if err != nil {
		return fmt.Errorf("创建阿里云SMS客户端失败: %w", err)
	}

	s.client = client
	return nil
}

// IsEnabled 检查短信服务是否启用
func (s *SMSService) IsEnabled() bool {
	return s.enabled && s.client != nil
}

// SendSMS 发送短信
func (s *SMSService) SendSMS(phoneNumber string, template SMSTemplate) error {
	if !s.IsEnabled() {
		return fmt.Errorf("短信服务未启用")
	}

	// 创建发送请求
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phoneNumber
	request.SignName = template.SignName
	request.TemplateCode = template.TemplateCode

	// 设置模板参数
	if len(template.Params) > 0 {
		paramsJSON, err := json.Marshal(template.Params)
		if err != nil {
			return fmt.Errorf("序列化短信参数失败: %w", err)
		}
		request.TemplateParam = string(paramsJSON)
	}

	// 发送短信
	response, err := s.client.SendSms(request)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "InvalidAccessKeyId.NotFound") {
			return fmt.Errorf("短信服务配置错误：阿里云AccessKey无效或不存在")
		}
		if strings.Contains(msg, "SignatureDoesNotMatch") {
			return fmt.Errorf("短信服务配置错误：阿里云AccessKeySecret不匹配")
		}
		if strings.Contains(msg, "InvalidSignName") || strings.Contains(msg, "SIGN_NAME_ILLEGAL") {
			return fmt.Errorf("短信服务配置错误：短信签名不正确或未审核通过")
		}
		if strings.Contains(msg, "InvalidTemplateCode") || strings.Contains(msg, "TEMPLATE_INVALID") {
			return fmt.Errorf("短信服务配置错误：短信模板不正确或未审核通过")
		}
		return fmt.Errorf("发送短信失败: %w", err)
	}

	// 检查响应状态
	if response.Code != "OK" {
		switch response.Code {
		case "InvalidAccessKeyId.NotFound":
			return fmt.Errorf("短信服务配置错误：阿里云AccessKey无效或不存在")
		case "SignatureDoesNotMatch":
			return fmt.Errorf("短信服务配置错误：阿里云AccessKeySecret不匹配")
		case "InvalidSignName", "SIGN_NAME_ILLEGAL":
			return fmt.Errorf("短信服务配置错误：短信签名不正确或未审核通过")
		case "InvalidTemplateCode", "TEMPLATE_INVALID":
			return fmt.Errorf("短信服务配置错误：短信模板不正确或未审核通过")
		}
		return fmt.Errorf("短信发送失败: %s - %s", response.Code, response.Message)
	}

	return nil
}

// SendScoreUpdateSMS 发送成绩更新短信
func (s *SMSService) SendScoreUpdateSMS(userID uint, scoreUpdates []ScoreUpdate) error {
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

	if !channel.SMSEnabled || !channel.ScoreUpdateSMS || channel.PhoneNumber == "" {
		return fmt.Errorf("用户未启用短信通知或未配置手机号")
	}

	// 检查余额
	balance, err := models.GetSMSBalanceByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取短信余额失败: %w", err)
	}

	if balance.Balance < SMSCostPerMessage {
		return fmt.Errorf("短信余额不足，当前余额: %.2f元", float64(balance.Balance)/100)
	}

	// 使用新的模板发送短信
	params := map[string]string{
		"name": user.Realname,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateScoreUpdate, params)
}

// SendSystemSMS 发送系统短信
func (s *SMSService) SendSystemSMS(userID uint, content string) error {
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

	if !channel.SMSEnabled || channel.PhoneNumber == "" {
		return fmt.Errorf("用户未启用短信通知或未配置手机号")
	}

	// 检查余额
	balance, err := models.GetSMSBalanceByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取短信余额失败: %w", err)
	}

	if balance.Balance < SMSCostPerMessage {
		return fmt.Errorf("短信余额不足，当前余额: %.2f元", float64(balance.Balance)/100)
	}

	// 使用系统通知模板
	params := map[string]string{
		"name":    user.Realname,
		"content": content,
	}

	// 获取系统通知模板
	templateConfig, err := s.templateManager.GetTemplate(config.TemplateSystemNotice)
	if err != nil {
		return fmt.Errorf("获取系统通知模板失败: %w", err)
	}

	// 验证参数
	if err := s.templateManager.ValidateParams(config.TemplateSystemNotice, params); err != nil {
		return fmt.Errorf("系统通知模板参数验证失败: %w", err)
	}

	// 构建短信模板
	template := SMSTemplate{
		TemplateCode: templateConfig.TemplateCode,
		SignName:     templateConfig.SignName,
		Params:       params,
	}

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelSMS,
		Type:      models.NotificationTypeSystem,
		Title:     "系统短信通知",
		Content:   content,
		Recipient: channel.PhoneNumber,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建短信通知日志失败: %v", err)
	}

	// 发送短信
	err = s.SendSMS(channel.PhoneNumber, template)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now

		// 扣除短信费用
		if balanceErr := models.UpdateSMSBalance(userID, SMSCostPerMessage, models.TransactionTypeConsume,
			"系统短信通知"); balanceErr != nil {
			log.Printf("扣除短信费用失败: %v", balanceErr)
		}

		// 创建消费记录
		transaction := &models.SMSTransaction{
			UserID:      userID,
			Type:        models.TransactionTypeConsume,
			Amount:      SMSCostPerMessage,
			Description: "系统短信通知",
			Status:      models.TransactionStatusSuccess,
		}
		if transErr := models.CreateSMSTransaction(transaction); transErr != nil {
			log.Printf("创建短信消费记录失败: %v", transErr)
		}
	}

	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新短信通知日志失败: %v", updateErr)
	}

	return err
}

// GetSMSCost 获取短信费用
func (s *SMSService) GetSMSCost() int {
	return SMSCostPerMessage
}

// ScoreUpdate 成绩更新数据结构（扩展版本包含详细信息）
type ScoreUpdate struct {
	CourseName string `json:"courseName"` // 课程名称
	CourseCode string `json:"courseCode"` // 课程代码
	Semester   string `json:"semester"`   // 学期
	ScoreType  string `json:"scoreType"`  // 成绩类型（平时分、考试分、最终分等）
	OldScore   string `json:"oldScore"`   // 原成绩
	NewScore   string `json:"newScore"`   // 新成绩
	UpdateTime string `json:"updateTime"` // 更新时间

	// 详细成绩信息
	DailyScore         string `json:"dailyScore"`         // 平时分
	ExamScore          string `json:"examScore"`          // 考试分
	FinalScore         string `json:"finalScore"`         // 最终分
	PracticalScore     string `json:"practicalScore"`     // 实践分
	SupplementaryScore string `json:"supplementaryScore"` // 补考分

	// 课程信息
	Credit         float64 `json:"credit"`         // 学分
	GPA            float64 `json:"gpa"`            // 绩点
	CourseProperty string  `json:"courseProperty"` // 课程性质（必修课、选修课等）
	TeacherNames   string  `json:"teacherNames"`   // 教师姓名
	TestNote       string  `json:"testNote"`       // 考试状态（正常、缓考等）

	// 变更类型
	ChangeType string `json:"changeType"` // 变更类型：new（新增）、update（更新）

	// 是否为首次同步（首次同步的变化不发送通知，避免误报）
	IsFirstSync bool `json:"isFirstSync"`
}

// SendSMSWithTemplate 使用模板发送短信
func (s *SMSService) SendSMSWithTemplate(userID uint, templateType config.SMSTemplateType, params map[string]string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("短信服务未启用")
	}

	// 获取模板配置
	template, err := s.templateManager.GetTemplate(templateType)
	if err != nil {
		return fmt.Errorf("获取短信模板失败: %w", err)
	}

	// 验证参数
	if err := s.templateManager.ValidateParams(templateType, params); err != nil {
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

	if !channel.SMSEnabled || channel.PhoneNumber == "" {
		return fmt.Errorf("用户未启用短信通知或未配置手机号")
	}

	// 检查余额
	balance, err := models.GetSMSBalanceByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取短信余额失败: %w", err)
	}

	if balance.Balance < SMSCostPerMessage {
		return fmt.Errorf("短信余额不足，当前余额: %.2f元", float64(balance.Balance)/100)
	}

	// 构建短信模板
	smsTemplate := SMSTemplate{
		TemplateCode: template.TemplateCode,
		SignName:     template.SignName,
		Params:       params,
	}

	// 创建通知日志
	content := s.templateManager.ReplaceParams(template.Content, params)
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelSMS,
		Type:      string(templateType),
		Title:     template.Description,
		Content:   content,
		Recipient: channel.PhoneNumber,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建短信通知日志失败: %v", err)
	}

	// 发送短信
	err = s.SendSMS(channel.PhoneNumber, smsTemplate)

	// 更新日志状态
	if err != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = err.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now

		// 扣除短信费用
		if balanceErr := models.UpdateSMSBalance(userID, SMSCostPerMessage, models.TransactionTypeConsume,
			fmt.Sprintf("%s - %s", template.Description, content)); balanceErr != nil {
			log.Printf("扣除短信费用失败: %v", balanceErr)
		}

		// 创建消费记录
		transaction := &models.SMSTransaction{
			UserID:      userID,
			Type:        models.TransactionTypeConsume,
			Amount:      SMSCostPerMessage,
			Description: fmt.Sprintf("%s - %s", template.Description, content),
			Status:      models.TransactionStatusSuccess,
		}
		if transErr := models.CreateSMSTransaction(transaction); transErr != nil {
			log.Printf("创建短信消费记录失败: %v", transErr)
		}
	}

	// 更新通知日志
	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新短信通知日志失败: %v", updateErr)
	}

	return err
}

// SendVerificationCodeSMS 发送验证码短信
func (s *SMSService) SendVerificationCodeSMS(phoneNumber, code string, minutes int) error {
	if !s.IsEnabled() {
		return fmt.Errorf("短信服务未启用")
	}

	// 获取验证码模板
	template, err := s.templateManager.GetTemplate(config.TemplateVerificationCode)
	if err != nil {
		return fmt.Errorf("获取验证码模板失败: %w", err)
	}

	// 构建参数
	params := map[string]string{
		"code":    code,
		"minutes": fmt.Sprintf("%d", minutes),
	}

	// 验证参数
	if err := s.templateManager.ValidateParams(config.TemplateVerificationCode, params); err != nil {
		return fmt.Errorf("验证码模板参数验证失败: %w", err)
	}

	// 构建短信模板
	smsTemplate := SMSTemplate{
		TemplateCode: template.TemplateCode,
		SignName:     template.SignName,
		Params:       params,
	}

	// 发送短信
	return s.SendSMS(phoneNumber, smsTemplate)
}

// SendCourseReminderSMS 发送课程提醒短信
func (s *SMSService) SendCourseReminderSMS(userID uint, courseName, courseTime, location string) error {
	params := map[string]string{
		"time":     courseTime,
		"course":   courseName,
		"location": location,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateCourseReminder, params)
}

// SendExamNoticeSMS 发送考试通知短信
func (s *SMSService) SendExamNoticeSMS(userID uint, courseName, examTime, location string) error {
	params := map[string]string{
		"course":   courseName,
		"time":     examTime,
		"location": location,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateExamNotice, params)
}

// SendEvaluationReminderSMS 发送评教提醒短信
func (s *SMSService) SendEvaluationReminderSMS(userID uint, semester string, courseCount int) error {
	params := map[string]string{
		"semester": semester,
		"count":    fmt.Sprintf("%d", courseCount),
	}

	return s.SendSMSWithTemplate(userID, config.TemplateEvaluationReminder, params)
}

// SendMaintenanceNoticeSMS 发送维护通知短信
func (s *SMSService) SendMaintenanceNoticeSMS(userID uint, maintenanceTime, duration string) error {
	params := map[string]string{
		"time":     maintenanceTime,
		"duration": duration,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateMaintenanceNotice, params)
}

// SendRechargeSuccessSMS 发送充值成功短信
func (s *SMSService) SendRechargeSuccessSMS(userID uint, amount, balance float64) error {
	params := map[string]string{
		"amount":  fmt.Sprintf("%.2f", amount),
		"balance": fmt.Sprintf("%.2f", balance),
	}

	return s.SendSMSWithTemplate(userID, config.TemplateRechargeSuccess, params)
}

// SendBalanceLowSMS 发送余额不足提醒短信
func (s *SMSService) SendBalanceLowSMS(userID uint, balance float64) error {
	params := map[string]string{
		"balance": fmt.Sprintf("%.2f", balance),
	}

	return s.SendSMSWithTemplate(userID, config.TemplateBalanceLow, params)
}

// SendSecurityAlertSMS 发送安全警告短信
func (s *SMSService) SendSecurityAlertSMS(userID uint, loginTime, location string) error {
	params := map[string]string{
		"time":     loginTime,
		"location": location,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateSecurityAlert, params)
}

// SendAccountBindingSMS 发送账户绑定短信
func (s *SMSService) SendAccountBindingSMS(userID uint, username string) error {
	params := map[string]string{
		"username": username,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateAccountBinding, params)
}

// SendCommunityReplySMS 发送社区回复通知短信
func (s *SMSService) SendCommunityReplySMS(userID uint, replyUser, postTitle string) error {
	// 截断标题长度
	title := config.TruncateString(postTitle, 20)

	params := map[string]string{
		"user":  replyUser,
		"title": title,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateCommunityReply, params)
}

// SendCommunityMentionSMS 发送社区@提醒短信
func (s *SMSService) SendCommunityMentionSMS(userID uint, mentionUser, postTitle string) error {
	// 截断标题长度
	title := config.TruncateString(postTitle, 20)

	params := map[string]string{
		"user":  mentionUser,
		"title": title,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateCommunityMention, params)
}

// SendAssignmentDueSMS 发送作业截止提醒短信
func (s *SMSService) SendAssignmentDueSMS(userID uint, courseName, assignmentName, deadline string) error {
	params := map[string]string{
		"course":     courseName,
		"assignment": assignmentName,
		"deadline":   deadline,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateAssignmentDue, params)
}

// SendEvaluationDeadlineSMS 发送评教截止提醒短信
func (s *SMSService) SendEvaluationDeadlineSMS(userID uint, deadline string, remainingCount int) error {
	params := map[string]string{
		"deadline": deadline,
		"count":    fmt.Sprintf("%d", remainingCount),
	}

	return s.SendSMSWithTemplate(userID, config.TemplateEvaluationDeadline, params)
}

// SendGradePublishedSMS 发送成绩发布通知短信
func (s *SMSService) SendGradePublishedSMS(userID uint, semester, courseName string) error {
	params := map[string]string{
		"semester": semester,
		"course":   courseName,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateGradePublished, params)
}

// SendLoginAlertSMS 发送登录提醒短信
func (s *SMSService) SendLoginAlertSMS(userID uint, loginTime, device string) error {
	params := map[string]string{
		"time":   loginTime,
		"device": device,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateLoginAlert, params)
}

// SendScoreTestNotificationSMS 发送成绩检查测试通知短信
func (s *SMSService) SendScoreTestNotificationSMS(userID uint) error {
	// 获取用户信息
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 确定用户名称，优先使用真实姓名，否则使用默认名称
	name := user.Realname
	if name == "" {
		name = "同学" // 默认名称，符合个人姓名规范
	}

	// 使用测试通过通知模板
	params := map[string]string{
		"name": name,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateScoreTestNotification, params)
}

// SendScoreChangeNotificationSMS 发送成绩变化通知短信
func (s *SMSService) SendScoreChangeNotificationSMS(userID uint) error {
	// 获取用户信息
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 确定用户名称，优先使用真实姓名，否则使用默认名称
	name := user.Realname
	if name == "" {
		name = "同学" // 默认名称，符合个人姓名规范
	}

	// 使用成绩变化通知模板
	params := map[string]string{
		"name": name,
	}

	return s.SendSMSWithTemplate(userID, config.TemplateScoreChangeNotification, params)
}

// GetTemplateManager 获取模板管理器
func (s *SMSService) GetTemplateManager() *config.SMSTemplateManager {
	return s.templateManager
}

// GetAllTemplates 获取所有可用的短信模板
func (s *SMSService) GetAllTemplates() map[config.SMSTemplateType]*config.SMSTemplateConfig {
	return s.templateManager.GetAllTemplates()
}

// GetEnabledTemplates 获取所有启用的短信模板
func (s *SMSService) GetEnabledTemplates() map[config.SMSTemplateType]*config.SMSTemplateConfig {
	return s.templateManager.GetEnabledTemplates()
}

// ValidateTemplateParams 验证模板参数
func (s *SMSService) ValidateTemplateParams(templateType config.SMSTemplateType, params map[string]string) error {
	return s.templateManager.ValidateParams(templateType, params)
}

// PreviewSMSContent 预览短信内容
func (s *SMSService) PreviewSMSContent(templateType config.SMSTemplateType, params map[string]string) (string, error) {
	template, err := s.templateManager.GetTemplate(templateType)
	if err != nil {
		return "", err
	}

	if err := s.templateManager.ValidateParams(templateType, params); err != nil {
		return "", err
	}

	return s.templateManager.ReplaceParams(template.Content, params), nil
}

// GetSMSCostPerMessage 获取每条短信费用
func GetSMSCostPerMessage() int {
	return SMSCostPerMessage
}
