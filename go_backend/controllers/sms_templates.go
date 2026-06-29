package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetSMSTemplates 获取所有短信模板
func GetSMSTemplates(c *gin.Context) {
	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	templates := smsService.GetAllTemplates()
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetEnabledSMSTemplates 获取启用的短信模板
func GetEnabledSMSTemplates(c *gin.Context) {
	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	templates := smsService.GetEnabledTemplates()
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetSMSTemplate 获取指定的短信模板
func GetSMSTemplate(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("模板类型不能为空", http.StatusBadRequest))
		return
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	template, err := smsService.GetTemplateManager().GetTemplate(config.SMSTemplateType(templateType))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(err.Error(), http.StatusNotFound))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"template": template,
	}))
}

// PreviewSMSTemplate 预览短信模板内容
func PreviewSMSTemplate(c *gin.Context) {
	var req struct {
		TemplateType string            `json:"templateType" binding:"required"`
		Params       map[string]string `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	content, err := smsService.PreviewSMSContent(config.SMSTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"content": content,
		"length":  len(content),
	}))
}

// ValidateSMSTemplateParams 验证短信模板参数
func ValidateSMSTemplateParams(c *gin.Context) {
	var req struct {
		TemplateType string            `json:"templateType" binding:"required"`
		Params       map[string]string `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.ValidateTemplateParams(config.SMSTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"valid":   true,
		"message": "参数验证通过",
	}))
}

// SendTestSMS 发送测试短信
func SendTestSMS(c *gin.Context) {
	var req struct {
		UserID       uint              `json:"userId" binding:"required"`
		TemplateType string            `json:"templateType" binding:"required"`
		Params       map[string]string `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.SendSMSWithTemplate(req.UserID, config.SMSTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送短信失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "短信发送成功",
	}))
}

// SendVerificationCode 发送验证码短信
func SendVerificationCode(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phoneNumber" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Minutes     int    `json:"minutes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	// 默认5分钟有效期
	if req.Minutes <= 0 {
		req.Minutes = 5
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.SendVerificationCodeSMS(req.PhoneNumber, req.Code, req.Minutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送验证码失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "验证码发送成功",
	}))
}

// SendCourseReminder 发送课程提醒短信
func SendCourseReminder(c *gin.Context) {
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

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.SendCourseReminderSMS(req.UserID, req.CourseName, req.CourseTime, req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送课程提醒失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "课程提醒发送成功",
	}))
}

// SendExamNotice 发送考试通知短信
func SendExamNotice(c *gin.Context) {
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

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.SendExamNoticeSMS(req.UserID, req.CourseName, req.ExamTime, req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送考试通知失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "考试通知发送成功",
	}))
}

// SendEvaluationReminder 发送评教提醒短信
func SendEvaluationReminder(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"userId" binding:"required"`
		Semester    string `json:"semester" binding:"required"`
		CourseCount int    `json:"courseCount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	err := smsService.SendEvaluationReminderSMS(req.UserID, req.Semester, req.CourseCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送评教提醒失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "评教提醒发送成功",
	}))
}

// GetSMSCost 获取短信费用信息
func GetSMSCost(c *gin.Context) {
	cost := services.GetSMSCostPerMessage()
	
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"costPerMessage": cost,
		"costInYuan":     float64(cost) / 100,
		"currency":       "CNY",
	}))
}

// BatchSendSMS 批量发送短信
func BatchSendSMS(c *gin.Context) {
	var req struct {
		UserIDs      []uint            `json:"userIds" binding:"required"`
		TemplateType string            `json:"templateType" binding:"required"`
		Params       map[string]string `json:"params" binding:"required"`
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

	smsService := services.GetSMSService()
	if smsService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("短信服务未初始化", http.StatusInternalServerError))
		return
	}

	var successCount, failCount int
	var errors []string

	for _, userID := range req.UserIDs {
		err := smsService.SendSMSWithTemplate(userID, config.SMSTemplateType(req.TemplateType), req.Params)
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
