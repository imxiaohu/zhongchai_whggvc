package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
)

// DingTalkService 钉钉通知服务
type DingTalkService struct {
	enabled bool
	client  *http.Client
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType  string                   `json:"msgtype"`
	Text     *DingTalkTextMessage     `json:"text,omitempty"`
	Markdown *DingTalkMarkdownMessage `json:"markdown,omitempty"`
	At       *DingTalkAtMessage       `json:"at,omitempty"`
}

// DingTalkTextMessage 钉钉文本消息
type DingTalkTextMessage struct {
	Content string `json:"content"`
}

// DingTalkMarkdownMessage 钉钉Markdown消息
type DingTalkMarkdownMessage struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingTalkAtMessage 钉钉@消息
type DingTalkAtMessage struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

// DingTalkResponse 钉钉响应
type DingTalkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

var dingTalkService *DingTalkService

// InitDingTalkService 初始化钉钉服务
func InitDingTalkService() {
	dingTalkService = NewDingTalkService()
}

// GetDingTalkService 获取钉钉服务实例
func GetDingTalkService() *DingTalkService {
	return dingTalkService
}

// NewDingTalkService 创建新的钉钉服务
func NewDingTalkService() *DingTalkService {
	service := &DingTalkService{
		enabled: true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	log.Printf("钉钉通知服务初始化完成")
	return service
}

// IsEnabled 检查钉钉服务是否启用
func (d *DingTalkService) IsEnabled() bool {
	return d.enabled
}

// ValidateWebhookURL 验证钉钉Webhook地址的有效性
func (d *DingTalkService) ValidateWebhookURL(webhookURL string) error {
	if !d.IsEnabled() {
		return fmt.Errorf("钉钉服务未启用")
	}

	// 发送测试消息
	testMessage := DingTalkMessage{
		MsgType: "text",
		Text: &DingTalkTextMessage{
			Content: "钉钉机器人配置测试消息",
		},
	}

	return d.sendMessage(webhookURL, "", testMessage)
}

// SendTextMessage 发送文本消息
func (d *DingTalkService) SendTextMessage(webhookURL, secret, content string) error {
	if !d.IsEnabled() {
		return fmt.Errorf("钉钉服务未启用")
	}

	message := DingTalkMessage{
		MsgType: "text",
		Text: &DingTalkTextMessage{
			Content: content,
		},
	}

	return d.sendMessage(webhookURL, secret, message)
}

// SendMarkdownMessage 发送Markdown消息
func (d *DingTalkService) SendMarkdownMessage(webhookURL, secret, title, content string) error {
	if !d.IsEnabled() {
		return fmt.Errorf("钉钉服务未启用")
	}

	message := DingTalkMessage{
		MsgType: "markdown",
		Markdown: &DingTalkMarkdownMessage{
			Title: title,
			Text:  content,
		},
	}

	return d.sendMessage(webhookURL, secret, message)
}

// SendScoreUpdateMessage 发送成绩更新消息
func (d *DingTalkService) SendScoreUpdateMessage(userID uint, scoreUpdates []ScoreUpdate) error {
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

	if !channel.DingTalkEnabled || !channel.ScoreUpdateDingTalk || channel.DingTalkWebhookURL == "" {
		return fmt.Errorf("用户未启用钉钉通知或未配置Webhook地址")
	}

	// 构建详细的Markdown消息内容
	title := "📊 成绩更新通知"
	content := fmt.Sprintf("## %s\n\n", title)
	content += fmt.Sprintf("**学生：** %s\n\n", user.Realname)
	content += fmt.Sprintf("**更新时间：** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("**更新数量：** %d 条\n\n", len(scoreUpdates))
	content += "---\n\n"

	for i, update := range scoreUpdates {
		content += fmt.Sprintf("### %d. %s\n\n", i+1, update.CourseName)

		// 基本信息
		content += "**📋 课程信息**\n\n"
		if update.CourseCode != "" {
			content += fmt.Sprintf("- **课程代码：** %s\n", update.CourseCode)
		}
		if update.Semester != "" {
			content += fmt.Sprintf("- **学期：** %s\n", update.Semester)
		}
		if update.TeacherNames != "" {
			content += fmt.Sprintf("- **任课教师：** %s\n", update.TeacherNames)
		}
		if update.CourseProperty != "" {
			content += fmt.Sprintf("- **课程性质：** %s\n", update.CourseProperty)
		}
		content += fmt.Sprintf("- **学分：** %.1f | **绩点：** %.1f\n", update.Credit, update.GPA)
		if update.TestNote != "" {
			content += fmt.Sprintf("- **考试状态：** %s\n", update.TestNote)
		}
		content += "\n"

		// 详细成绩信息
		content += "**📊 详细成绩**\n\n"
		content += "| 成绩类型 | 分数 |\n"
		content += "|---------|------|\n"

		if update.DailyScore != "0" && update.DailyScore != "" {
			content += fmt.Sprintf("| 平时分 | %s |\n", update.DailyScore)
		}
		if update.ExamScore != "0" && update.ExamScore != "" {
			content += fmt.Sprintf("| 考试分 | %s |\n", update.ExamScore)
		}
		if update.FinalScore != "0" && update.FinalScore != "" {
			content += fmt.Sprintf("| **最终分** | **%s** |\n", update.FinalScore)
		}
		if update.PracticalScore != "0" && update.PracticalScore != "" {
			content += fmt.Sprintf("| 实践分 | %s |\n", update.PracticalScore)
		}
		if update.SupplementaryScore != "0" && update.SupplementaryScore != "" {
			content += fmt.Sprintf("| 补考分 | %s |\n", update.SupplementaryScore)
		}
		content += "\n"

		// 变更信息
		content += "**🔄 变更详情**\n\n"
		if update.ChangeType == "new" {
			content += "- **变更类型：** 🆕 新增成绩\n"
		} else {
			content += "- **变更类型：** 📝 成绩更新\n"
		}
		content += fmt.Sprintf("- **变更项目：** %s\n", update.ScoreType)
		if update.OldScore != "" {
			content += fmt.Sprintf("- **成绩变化：** %s → **%s**\n", update.OldScore, update.NewScore)
		} else {
			content += fmt.Sprintf("- **成绩：** **%s**\n", update.NewScore)
		}
		content += fmt.Sprintf("- **更新时间：** %s\n\n", update.UpdateTime)

		if i < len(scoreUpdates)-1 {
			content += "---\n\n"
		}
	}

	content += "\n---\n\n"
	content += "💡 **温馨提示：** 请及时登录系统查看完整的成绩信息\n\n"
	content += "*此消息由评教系统自动发送，请勿回复*"

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeScoreUpdate,
		Title:     title,
		Content:   fmt.Sprintf("发送成绩更新钉钉消息，包含%d条更新", len(scoreUpdates)),
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建钉钉通知日志失败: %v", err)
	}

	// 发送消息
	err = d.SendMarkdownMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, title, content)

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
		log.Printf("更新钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// SendSystemMessage 发送系统消息
func (d *DingTalkService) SendSystemMessage(userID uint, title, content string) error {
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

	if !channel.DingTalkEnabled || channel.DingTalkWebhookURL == "" {
		return fmt.Errorf("用户未启用钉钉通知或未配置Webhook地址")
	}

	// 构建Markdown消息内容
	markdownContent := fmt.Sprintf("## 📢 %s\n\n", title)
	markdownContent += fmt.Sprintf("**接收人：** %s\n\n", user.Realname)
	markdownContent += fmt.Sprintf("**发送时间：** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	markdownContent += "### 消息内容\n\n"
	markdownContent += content + "\n\n"
	markdownContent += "---\n\n"
	markdownContent += "*此消息由评教系统自动发送*"

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelDingTalk,
		Type:      models.NotificationTypeSystem,
		Title:     title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建钉钉通知日志失败: %v", err)
	}

	// 发送消息
	err = d.SendMarkdownMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, title, markdownContent)

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
		log.Printf("更新钉钉通知日志失败: %v", updateErr)
	}

	return err
}

// sendMessage 发送消息到钉钉
func (d *DingTalkService) sendMessage(webhookURL, secret string, message DingTalkMessage) error {
	// 如果有密钥，需要计算签名
	finalURL := webhookURL
	if secret != "" {
		timestamp := time.Now().UnixNano() / 1e6
		sign := d.calculateSign(timestamp, secret)
		finalURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
	}

	// 序列化消息
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送HTTP请求
	resp, err := d.client.Post(finalURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var dingResp DingTalkResponse
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应状态
	if dingResp.ErrCode != 0 {
		return fmt.Errorf("钉钉API返回错误: %d - %s", dingResp.ErrCode, dingResp.ErrMsg)
	}

	return nil
}

// calculateSign 计算钉钉机器人签名
func (d *DingTalkService) calculateSign(timestamp int64, secret string) string {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// SendDingTalkWithTemplate 使用模板发送钉钉消息
func (d *DingTalkService) SendDingTalkWithTemplate(userID uint, templateType config.DingTalkTemplateType, params map[string]interface{}) error {
	if !d.IsEnabled() {
		return fmt.Errorf("钉钉服务未启用")
	}

	// 获取模板管理器
	templateManager := config.GetDingTalkTemplateManager()

	// 获取模板配置
	dingTalkTemplate, err := templateManager.GetTemplate(templateType)
	if err != nil {
		return fmt.Errorf("获取钉钉模板失败: %w", err)
	}

	// 验证参数
	if err := templateManager.ValidateParams(templateType, params); err != nil {
		return fmt.Errorf("模板参数验证失败: %w", err)
	}

	// 获取用户信息
	_, err = models.FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 获取通知渠道配置
	channel, err := models.GetNotificationChannelByUserID(userID)
	if err != nil {
		return fmt.Errorf("获取通知配置失败: %w", err)
	}

	if !channel.DingTalkEnabled || channel.DingTalkWebhookURL == "" {
		return fmt.Errorf("用户未启用钉钉通知或未配置Webhook地址")
	}

	// 替换模板参数
	content := templateManager.ReplaceParams(dingTalkTemplate.Content, params)

	// 创建通知日志
	notificationLog := &models.NotificationLog{
		UserID:    userID,
		Channel:   models.ChannelDingTalk,
		Type:      string(templateType),
		Title:     dingTalkTemplate.Title,
		Content:   content,
		Recipient: channel.DingTalkWebhookURL,
		Status:    models.SendStatusPending,
	}

	if err := models.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("创建钉钉通知日志失败: %v", err)
	}

	// 发送消息
	var sendErr error
	if dingTalkTemplate.MessageType == config.DingTalkMessageTypeMarkdown {
		sendErr = d.SendMarkdownMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, dingTalkTemplate.Title, content)
	} else {
		sendErr = d.SendTextMessage(channel.DingTalkWebhookURL, channel.DingTalkSecret, content)
	}

	// 更新日志状态
	if sendErr != nil {
		notificationLog.Status = models.SendStatusFailed
		notificationLog.ErrorMsg = sendErr.Error()
	} else {
		notificationLog.Status = models.SendStatusSuccess
		now := time.Now()
		notificationLog.SentAt = &now
	}

	// 更新通知日志
	if updateErr := models.UpdateNotificationLog(notificationLog); updateErr != nil {
		log.Printf("更新钉钉通知日志失败: %v", updateErr)
	}

	return sendErr
}
