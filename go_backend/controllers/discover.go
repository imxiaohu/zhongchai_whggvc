// Package controllers handles discover feature proxy endpoints.
package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// GetUserAndProxy performs a standard discover proxy request.
// It checks auth, validates school token, and proxies the request to the school server.
func GetUserAndProxy(c *gin.Context, schoolPath string) ([]byte, error) {
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return nil, fmt.Errorf("unauthorized")
	}
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return nil, fmt.Errorf("invalid user id")
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return nil, fmt.Errorf("invalid user id")
	}
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return nil, errors.New("user not found")
	}

	if user.SchoolToken == "" {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("学校账号未登录或令牌缺失，请重新登录学校账号", 502))
		return nil, fmt.Errorf("school token missing")
	}

	proxyClient := utils.NewProxyClient()
	clientID := c.GetHeader("X-Client-Id")

	params := c.Request.URL.Query()
	if clientID != "" {
		params.Set("clientId", clientID)
	}

	body, err := proxyRequestWithClientID(proxyClient, user, "GET", schoolPath, clientID, "", "", params)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ClassroomAttendanceListProxy 考勤记录列表代理
func ClassroomAttendanceListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/classroomAttendance/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取考勤记录失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetOptionalTeachingClassProxy 选课记录代理
func GetOptionalTeachingClassProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/course/tCourseOptionalStudent/getOptionalTeachingClass")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取选课记录失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetSemesterOptionalTeachingClassProxy 学期可选教学班代理
func GetSemesterOptionalTeachingClassProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/course/tCourseOptionalStudent/getSemesterOptionalTeachingClass")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取可选教学班失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetTStudentLeaveProxy 请假记录代理
func GetTStudentLeaveProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/scs/leave/tStudentLeave/getTStudentLeave")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取请假记录失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetTOaSurveyProxy 问卷列表代理
func GetTOaSurveyProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/scs/survey/tOaSurvey/getTOaSurvey")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取问卷列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetTOaSurveyQuestionProxy 问卷问题列表代理
func GetTOaSurveyQuestionProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/scs/survey/tOaSurveyQuestion/getTOaSurveyQuestion")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取问卷问题失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetTOaSurveyQuestionAnswerProxy 问卷答案代理
func GetTOaSurveyQuestionAnswerProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/scs/survey/tOaSurveyQuestionAnswer/getTOaSurveyQuestionAnswer")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取问卷答案失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// QueryUserByIdProxy 用户信息代理
func QueryUserByIdProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/sys/user/queryById")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取用户信息失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// RepairListProxy 报修记录列表代理
func RepairListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/repairReport/tLogisticsMaintenanceOrder/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取报修记录失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// RepairDetailProxy 报修单详情代理
func RepairDetailProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/repairReport/tLogisticsMaintenanceOrder/getStudent")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取报修单详情失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// RepairTypesProxy 报修类型（楼栋）列表代理
func RepairTypesProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/repairReport/tLogisticsMaintenanceOrder/getLogisticsMaintenanceOrderType")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取报修类型失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// FileServerUrlProxy 文件服务器地址代理
func FileServerUrlProxy(c *gin.Context) {
	_, err := GetUserAndProxy(c, "/scloudoa/sys/common/getFileServerUrl")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取文件服务器地址失败: "+err.Error(), 502))
		return
	}
	// 返回后端文件代理地址，前端通过此地址加载文件（带用户认证）
	proxyBaseURL := config.GetAPIBaseURL()
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": proxyBaseURL + "/api/proxy/file",
		"code":    200,
		"result":  proxyBaseURL + "/api/proxy/file",
	})
}

// GetStudentBankProxy 我的银行卡代理（同时返回学生档案信息用于补充个人资料）
func GetStudentBankProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloudoa/studentBank/getStudentBank")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取银行卡信息失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// EditStudentBankProxy 编辑银行卡信息代理
func EditStudentBankProxy(c *gin.Context) {
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return
	}
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	if user.SchoolToken == "" {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("学校账号未登录或令牌缺失，请重新登录学校账号", 502))
		return
	}

	proxyClient := utils.NewProxyClient()
	clientID := c.GetHeader("X-Client-Id")
	overrideCookie := c.GetHeader("Cookie")

	// 自动登录确保会话有效
	if err := proxyClient.AutoLogin(user, false); err != nil {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("学校账号自动登录失败: "+err.Error(), 502))
		return
	}

	// 解析请求体
	var reqData map[string]interface{}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("请求数据格式错误", 400))
		return
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("序列化请求数据失败: "+err.Error(), 400))
		return
	}

	// 代理 PUT 请求到学校服务器
	req, err := http.NewRequest("PUT", proxyClient.BaseURL+"/scloudoa/studentBank/edit", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败: "+err.Error(), 500))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if clientID != "" {
		req.Header.Set("clientId", clientID)
	}
	proxyClient.SetCommonHeaders(req)
	if overrideCookie != "" {
		req.Header.Set("Cookie", overrideCookie)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("编辑银行卡信息失败: "+err.Error(), 502))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("读取学校服务器响应失败: "+err.Error(), 502))
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", respBody)
}

// InternshipRequirementsListProxy 实习要求列表代理
func InternshipRequirementsListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipRequirementsInquiry/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习要求列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipJobClassificationProxy 岗位职业分类代理
func InternshipJobClassificationProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipPositionInquiry/getJobClassification")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取岗位职业分类失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipIndustryClassificationProxy 行业分类代理
func InternshipIndustryClassificationProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipPositionInquiry/getIndustryClassification")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取行业分类失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipPositionListProxy 实习岗位列表代理
func InternshipPositionListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipPositionInquiry/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习岗位列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipApplicationListProxy 实习申请列表代理
func InternshipApplicationListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipPositionApplication/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习申请列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipSignInPlanProxy 实习签到计划代理
func InternshipSignInPlanProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipSignIn/getPlan")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习签到计划失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// MyInternshipPlanListProxy 我的实习计划列表代理
func MyInternshipPlanListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/myInternshipPlan/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取我的实习计划列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// MyInternshipSummaryListProxy 我的实习总结列表代理
func MyInternshipSummaryListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/myInternshipSummary/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取我的实习总结列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipAppraisalFormListProxy 实习鉴定表列表代理
func InternshipAppraisalFormListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipAppraisalForm/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习鉴定表列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetRelationTypesProxy 亲属关系类型列表代理（档案编辑用）
func GetRelationTypesProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloud/student/base/relation")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取亲属关系类型失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// GetProvinceProxy 省份列表代理（档案编辑用）
func GetProvinceProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/scloud/student/base/getProvince")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取省份列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// InternshipScoreInquiryListProxy 实习成绩查询列表代理
func InternshipScoreInquiryListProxy(c *gin.Context) {
	body, err := GetUserAndProxy(c, "/jobInternship/internshipScoreInquiry/list")
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取实习成绩列表失败: "+err.Error(), 502))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}
