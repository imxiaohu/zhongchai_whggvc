// Package controllers handles HTTP request handlers for proxy-related endpoints.
// This file contains the proxy route handler functions.
package controllers

import (
	"github.com/gin-gonic/gin"
)

// 全局代理控制器实例
var ProxyCtrl = NewProxyController()

// GetSchoolProxy 获取学校信息代理
func GetSchoolProxy(c *gin.Context) {
	ProxyCtrl.GetSchool(c)
}

// GetCurrentTimeProxy 获取当前时间代理
func GetCurrentTimeProxy(c *gin.Context) {
	ProxyCtrl.GetCurrentTime(c)
}

// GetCourseTimeTableByDayProxy 获取当日课程表代理
func GetCourseTimeTableByDayProxy(c *gin.Context) {
	ProxyCtrl.GetCourseTimeTableByDay(c)
}

// GetCourseTimeTableByWeekProxy 获取本周课程表代理
func GetCourseTimeTableByWeekProxy(c *gin.Context) {
	ProxyCtrl.GetCourseTimeTableByWeek(c)
}

// GetEoaNewsTypeListProxy 获取新闻类型列表代理
func GetEoaNewsTypeListProxy(c *gin.Context) {
	ProxyCtrl.GetEoaNewsTypeList(c)
}

// GetEoaNewsListByTypeIdProxy 获取新闻列表代理
func GetEoaNewsListByTypeIdProxy(c *gin.Context) {
	ProxyCtrl.GetEoaNewsListByTypeId(c)
}

// GetLearningDataProxy 获取学习数据代理
func GetLearningDataProxy(c *gin.Context) {
	ProxyCtrl.GetLearningData(c)
}

// GetSemesterProxy 获取学期信息代理
func GetSemesterProxy(c *gin.Context) {
	ProxyCtrl.GetSemester(c)
}

// GetScoreCurrentTimeProxy 获取成绩当前时间代理
func GetScoreCurrentTimeProxy(c *gin.Context) {
	ProxyCtrl.GetScoreCurrentTime(c)
}

// GetCoursePlanProxy 获取课程计划代理
func GetCoursePlanProxy(c *gin.Context) {
	ProxyCtrl.GetCoursePlan(c)
}

// GetCourseScoreProxy 获取课程成绩代理
func GetCourseScoreProxy(c *gin.Context) {
	ProxyCtrl.GetCourseScore(c)
}

// GetCourseLessonTimeProxy 获取课程时间段代理
func GetCourseLessonTimeProxy(c *gin.Context) {
	ProxyCtrl.GetCourseLessonTime(c)
}

// GetCourseSchoolTimetableProxy 获取学校课表代理
func GetCourseSchoolTimetableProxy(c *gin.Context) {
	ProxyCtrl.GetCourseSchoolTimetable(c)
}

// GetEvaluationStudentConfigListProxy 获取评教配置代理
func GetEvaluationStudentConfigListProxy(c *gin.Context) {
	ProxyCtrl.GetEvaluationStudentConfigList(c)
}

// GetScoreListProxy 获取成绩列表代理
func GetScoreListProxy(c *gin.Context) {
	ProxyCtrl.GetScoreList(c)
}

// GetSemesterScoreProxy 获取学期统计成绩代理
func GetSemesterScoreProxy(c *gin.Context) {
	ProxyCtrl.GetSemesterScore(c)
}

// DownloadAttachmentProxy 下载附件代理
func DownloadAttachmentProxy(c *gin.Context) {
	ProxyCtrl.DownloadAttachment(c)
}

// FileProxy 文件代理
func FileProxy(c *gin.Context) {
	ProxyCtrl.FileProxy(c)
}
