package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetEmailTemplates 获取所有邮件模板
func GetEmailTemplates(c *gin.Context) {
	emailTemplateManager := config.GetEmailTemplateManager()
	templates := emailTemplateManager.GetAllTemplates()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetEnabledEmailTemplates 获取启用的邮件模板
func GetEnabledEmailTemplates(c *gin.Context) {
	emailTemplateManager := config.GetEmailTemplateManager()
	templates := emailTemplateManager.GetEnabledTemplates()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetEmailTemplate 获取指定的邮件模板
func GetEmailTemplate(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("模板类型不能为空", http.StatusBadRequest))
		return
	}

	emailTemplateManager := config.GetEmailTemplateManager()
	template, err := emailTemplateManager.GetTemplate(config.EmailTemplateType(templateType))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(err.Error(), http.StatusNotFound))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"template": template,
	}))
}

// GetEmailTemplatesByCategory 根据分类获取邮件模板
func GetEmailTemplatesByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("分类不能为空", http.StatusBadRequest))
		return
	}

	emailTemplateManager := config.GetEmailTemplateManager()
	templates := emailTemplateManager.GetTemplatesByCategory(category)

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"category":  category,
		"total":     len(templates),
	}))
}

// PreviewEmailTemplate 预览邮件模板内容
func PreviewEmailTemplate(c *gin.Context) {
	var req struct {
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
		Format       string                 `json:"format"` // html 或 text
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailTemplateManager := config.GetEmailTemplateManager()
	template, err := emailTemplateManager.GetTemplate(config.EmailTemplateType(req.TemplateType))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	// 验证参数
	if err := emailTemplateManager.ValidateParams(config.EmailTemplateType(req.TemplateType), req.Params); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	// 选择格式
	content := template.HTMLBody
	if req.Format == "text" {
		content = template.TextBody
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"subject": template.Subject,
		"content": content,
		"format":  req.Format,
		"params":  req.Params,
	}))
}

// ValidateEmailTemplateParams 验证邮件模板参数
func ValidateEmailTemplateParams(c *gin.Context) {
	var req struct {
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailTemplateManager := config.GetEmailTemplateManager()
	err := emailTemplateManager.ValidateParams(config.EmailTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"valid":   true,
		"message": "参数验证通过",
	}))
}

// SendTestEmail 发送测试邮件
func SendTestEmail(c *gin.Context) {
	var req struct {
		UserID       uint                   `json:"userId" binding:"required"`
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailService := services.GetEmailService()
	if emailService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("邮件服务未初始化", http.StatusInternalServerError))
		return
	}

	err := emailService.SendEmailWithTemplate(req.UserID, config.EmailTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送邮件失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "邮件发送成功",
	}))
}

// SendCourseReminderEmail 发送课程提醒邮件
func SendCourseReminderEmail(c *gin.Context) {
	var req struct {
		UserID     uint   `json:"userId" binding:"required"`
		CourseName string `json:"courseName" binding:"required"`
		CourseTime string `json:"courseTime" binding:"required"`
		Location   string `json:"location" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailService := services.GetEmailService()
	if emailService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("邮件服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":   "用户", // 这里应该从数据库获取真实用户名
		"CourseTime": req.CourseTime,
		"CourseName": req.CourseName,
		"Location":   req.Location,
		"SystemName": "评教系统",
		"SendTime":   config.FormatEmailTime(time.Now()),
	}

	err := emailService.SendEmailWithTemplate(req.UserID, config.EmailTemplateCourseReminder, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送课程提醒邮件失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "课程提醒邮件发送成功",
	}))
}

// SendExamNoticeEmail 发送考试通知邮件
func SendExamNoticeEmail(c *gin.Context) {
	var req struct {
		UserID     uint   `json:"userId" binding:"required"`
		CourseName string `json:"courseName" binding:"required"`
		ExamTime   string `json:"examTime" binding:"required"`
		Location   string `json:"location" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailService := services.GetEmailService()
	if emailService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("邮件服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":   "用户", // 这里应该从数据库获取真实用户名
		"CourseName": req.CourseName,
		"ExamTime":   req.ExamTime,
		"Location":   req.Location,
		"SystemName": "评教系统",
		"SendTime":   config.FormatEmailTime(time.Now()),
	}

	err := emailService.SendEmailWithTemplate(req.UserID, config.EmailTemplateExamNotice, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送考试通知邮件失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "考试通知邮件发送成功",
	}))
}

// SendEvaluationReminderEmail 发送评教提醒邮件
func SendEvaluationReminderEmail(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"userId" binding:"required"`
		Semester    string `json:"semester" binding:"required"`
		CourseCount int    `json:"courseCount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	emailService := services.GetEmailService()
	if emailService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("邮件服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":    "用户", // 这里应该从数据库获取真实用户名
		"Semester":    req.Semester,
		"CourseCount": strconv.Itoa(req.CourseCount),
		"SystemName":  "评教系统",
		"SendTime":    config.FormatEmailTime(time.Now()),
	}

	err := emailService.SendEmailWithTemplate(req.UserID, config.EmailTemplateEvaluationReminder, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送评教提醒邮件失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "评教提醒邮件发送成功",
	}))
}

// BatchSendEmail 批量发送邮件
func BatchSendEmail(c *gin.Context) {
	var req struct {
		UserIDs      []uint                 `json:"userIds" binding:"required"`
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID列表不能为空", http.StatusBadRequest))
		return
	}

	if len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("批量发送用户数量不能超过100个", http.StatusBadRequest))
		return
	}

	emailService := services.GetEmailService()
	if emailService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("邮件服务未初始化", http.StatusInternalServerError))
		return
	}

	var successCount, failCount int
	var errors []string

	for _, userID := range req.UserIDs {
		err := emailService.SendEmailWithTemplate(userID, config.EmailTemplateType(req.TemplateType), req.Params)
		if err != nil {
			failCount++
			errors = append(errors, "用户"+strconv.Itoa(int(userID))+": "+err.Error())
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"total":        len(req.UserIDs),
		"successCount": successCount,
		"failCount":    failCount,
		"errors":       errors,
	}))
}
