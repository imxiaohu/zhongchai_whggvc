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

// GetDingTalkTemplates 获取所有钉钉模板
func GetDingTalkTemplates(c *gin.Context) {
	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	templates := dingTalkTemplateManager.GetAllTemplates()
	
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetEnabledDingTalkTemplates 获取启用的钉钉模板
func GetEnabledDingTalkTemplates(c *gin.Context) {
	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	templates := dingTalkTemplateManager.GetEnabledTemplates()
	
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"total":     len(templates),
	}))
}

// GetDingTalkTemplate 获取指定的钉钉模板
func GetDingTalkTemplate(c *gin.Context) {
	templateType := c.Param("type")
	if templateType == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("模板类型不能为空", http.StatusBadRequest))
		return
	}

	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	template, err := dingTalkTemplateManager.GetTemplate(config.DingTalkTemplateType(templateType))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse(err.Error(), http.StatusNotFound))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"template": template,
	}))
}

// GetDingTalkTemplatesByCategory 根据分类获取钉钉模板
func GetDingTalkTemplatesByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("分类不能为空", http.StatusBadRequest))
		return
	}

	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	templates := dingTalkTemplateManager.GetTemplatesByCategory(category)
	
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"templates": templates,
		"category":  category,
		"total":     len(templates),
	}))
}

// PreviewDingTalkTemplate 预览钉钉模板内容
func PreviewDingTalkTemplate(c *gin.Context) {
	var req struct {
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	template, err := dingTalkTemplateManager.GetTemplate(config.DingTalkTemplateType(req.TemplateType))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	// 验证参数
	if err := dingTalkTemplateManager.ValidateParams(config.DingTalkTemplateType(req.TemplateType), req.Params); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	// 替换参数
	content := dingTalkTemplateManager.ReplaceParams(template.Content, req.Params)

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"title":       template.Title,
		"content":     content,
		"messageType": template.MessageType,
		"params":      req.Params,
	}))
}

// ValidateDingTalkTemplateParams 验证钉钉模板参数
func ValidateDingTalkTemplateParams(c *gin.Context) {
	var req struct {
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	dingTalkTemplateManager := config.GetDingTalkTemplateManager()
	err := dingTalkTemplateManager.ValidateParams(config.DingTalkTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"valid":   true,
		"message": "参数验证通过",
	}))
}

// SendTestDingTalk 发送测试钉钉消息
func SendTestDingTalk(c *gin.Context) {
	var req struct {
		UserID       uint                   `json:"userId" binding:"required"`
		TemplateType string                 `json:"templateType" binding:"required"`
		Params       map[string]interface{} `json:"params" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	dingTalkService := services.GetDingTalkService()
	if dingTalkService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("钉钉服务未初始化", http.StatusInternalServerError))
		return
	}

	err := dingTalkService.SendDingTalkWithTemplate(req.UserID, config.DingTalkTemplateType(req.TemplateType), req.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送钉钉消息失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "钉钉消息发送成功",
	}))
}

// SendCourseReminderDingTalk 发送课程提醒钉钉消息
func SendCourseReminderDingTalk(c *gin.Context) {
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

	dingTalkService := services.GetDingTalkService()
	if dingTalkService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("钉钉服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":   "用户", // 这里应该从数据库获取真实用户名
		"CourseTime": req.CourseTime,
		"CourseName": req.CourseName,
		"Location":   req.Location,
		"SystemName": "评教系统",
		"SendTime":   config.FormatDingTalkTime(time.Now()),
	}

	err := dingTalkService.SendDingTalkWithTemplate(req.UserID, config.DingTalkTemplateCourseReminder, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送课程提醒钉钉消息失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "课程提醒钉钉消息发送成功",
	}))
}

// SendExamNoticeDingTalk 发送考试通知钉钉消息
func SendExamNoticeDingTalk(c *gin.Context) {
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

	dingTalkService := services.GetDingTalkService()
	if dingTalkService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("钉钉服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":   "用户", // 这里应该从数据库获取真实用户名
		"CourseName": req.CourseName,
		"ExamTime":   req.ExamTime,
		"Location":   req.Location,
		"SystemName": "评教系统",
		"SendTime":   config.FormatDingTalkTime(time.Now()),
	}

	err := dingTalkService.SendDingTalkWithTemplate(req.UserID, config.DingTalkTemplateExamNotice, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送考试通知钉钉消息失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "考试通知钉钉消息发送成功",
	}))
}

// SendEvaluationReminderDingTalk 发送评教提醒钉钉消息
func SendEvaluationReminderDingTalk(c *gin.Context) {
	var req struct {
		UserID      uint   `json:"userId" binding:"required"`
		Semester    string `json:"semester" binding:"required"`
		CourseCount int    `json:"courseCount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求参数错误: "+err.Error(), http.StatusBadRequest))
		return
	}

	dingTalkService := services.GetDingTalkService()
	if dingTalkService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("钉钉服务未初始化", http.StatusInternalServerError))
		return
	}

	params := map[string]interface{}{
		"UserName":    "用户", // 这里应该从数据库获取真实用户名
		"Semester":    req.Semester,
		"CourseCount": strconv.Itoa(req.CourseCount),
		"SystemName":  "评教系统",
		"SendTime":    config.FormatDingTalkTime(time.Now()),
	}

	err := dingTalkService.SendDingTalkWithTemplate(req.UserID, config.DingTalkTemplateEvaluationReminder, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("发送评教提醒钉钉消息失败: "+err.Error(), http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "评教提醒钉钉消息发送成功",
	}))
}

// BatchSendDingTalk 批量发送钉钉消息
func BatchSendDingTalk(c *gin.Context) {
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

	dingTalkService := services.GetDingTalkService()
	if dingTalkService == nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("钉钉服务未初始化", http.StatusInternalServerError))
		return
	}

	var successCount, failCount int
	var errors []string

	for _, userID := range req.UserIDs {
		err := dingTalkService.SendDingTalkWithTemplate(userID, config.DingTalkTemplateType(req.TemplateType), req.Params)
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
