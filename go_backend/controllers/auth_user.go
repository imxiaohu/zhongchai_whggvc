// Package controllers handles HTTP request handlers for authentication endpoints.
// This file contains user-related functions.
package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// GetSchool 获取学校信息
func GetSchool(c *gin.Context) {
	// 获取所有学校
	schools, err := models.GetAllSchools()
	if err != nil || len(schools) == 0 {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("未找到学校信息", 404))
		return
	}

	// 返回第一个学校信息
	school := schools[0]
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"id":          school.ID,
		"name":        school.Name,
		"code":        school.Code,
		"logo":        school.Logo,
		"address":     school.Address,
		"description": school.Description,
		"website":     school.Website,
	}))
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
	// 从上下文获取用户ID
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

	// 查询用户信息
	var user models.User
	result := models.DB.First(&user, userIdStr)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 检查是否绑定学校账号
	// 绑定条件：
	// 1. 有用户名和密码
	// 2. 用户名不是微信临时用户名（不以"wx_"开头）
	// 3. 用户名不等于微信OpenID
	// 4. 有真实姓名（从学校系统获取）
	hasSchoolAccount := user.Username != "" &&
		user.Password != "" &&
		!strings.HasPrefix(user.Username, "wx_") &&
		user.Username != user.WechatOpenID

	// 构建完整的用户信息响应
	userInfo := gin.H{
		"id":               fmt.Sprintf("%d", user.ID),
		"username":         user.Username,
		"realname":         user.Realname,
		"nickname":         user.Nickname,
		"avatar":           user.Avatar,
		"avatarUrl":        user.Avatar, // 兼容前端期望的字段名
		"email":            user.Email,
		"phone":            user.Phone,
		"birthday":         user.Birthday,
		"sex":              user.Sex,
		"identityCard":     user.IdentityCard,
		"userType":         user.UserType,
		"schoolId":         user.SchoolID,
		"className":        user.ClassName,
		"professionId":     user.ProfessionID,
		"facultyId":        user.FacultyID,
		"gradeId":          user.GradeID,
		"currentSemester":  user.CurrentSemester,
		"hasSchoolAccount": hasSchoolAccount,
		"wechatOpenId":     user.WechatOpenID,
		"lastLoginAt":      user.LastLoginAt,
		"status":           user.Status,
	}

	// 返回用户信息
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(userInfo))
}

// GetUserStatistics 获取用户统计信息
func GetUserStatistics(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	hasSchoolAccount := user.Username != "" &&
		user.Password != "" &&
		!strings.HasPrefix(user.Username, "wx_") &&
		user.Username != user.WechatOpenID

	if !hasSchoolAccount {
		c.JSON(http.StatusOK, utils.NewStandardResponse("获取用户统计成功", gin.H{
			"hasSchoolAccount":        false,
			"semester":                user.CurrentSemester,
			"averageScore":            nil,
			"creditTotal":             nil,
			"pendingEvaluationCount":  0,
			"evaluationDataAvailable": false,
			"scoreDataAvailable":      false,
		}))
		return
	}

	currentSemester := user.CurrentSemester
	if currentSemester == "" {
		schoolAPIService := services.NewSchoolAPIService()
		semesterInfo, err := schoolAPIService.GetCurrentSemester(user)
		if err == nil && semesterInfo != nil && semesterInfo.Name != "" {
			currentSemester = semesterInfo.Name
		}
		if currentSemester != "" && user.CurrentSemester != currentSemester {
			//nolint:errcheck
			models.UpdateUserFields(user.ID, map[string]interface{}{
				"current_semester": currentSemester,
			})
		}
	}

	var averageScore *float64
	var creditTotal *float64
	scoreDataAvailable := false

	if currentSemester != "" {
		statsResp := make(map[string]interface{})
		statsResp["averageScore"] = nil
		statsResp["creditTotal"] = nil
		averageScore, creditTotal = extractSemesterStatistics(statsResp)
		scoreDataAvailable = averageScore != nil || creditTotal != nil
	}

	pendingEvaluationCount := 0
	evaluationDataAvailable := false

	{
		proxyClient := utils.NewProxyClient()
		clientID := c.GetHeader("X-Client-Id")
		overrideCookie := c.GetHeader("Cookie")
		params := url.Values{}
		if clientID != "" {
			params.Set("clientId", clientID)
		}

		body, err := proxyRequestWithClientID(proxyClient, user, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", clientID, "", overrideCookie, params)
		if err == nil && len(body) > 0 {
			if count := extractListCount(body); count >= 0 {
				pendingEvaluationCount = count
				evaluationDataAvailable = true
			}
		}
	}

	c.JSON(http.StatusOK, utils.NewStandardResponse("获取用户统计成功", gin.H{
		"hasSchoolAccount":        true,
		"semester":                currentSemester,
		"averageScore":            averageScore,
		"creditTotal":             creditTotal,
		"pendingEvaluationCount":  pendingEvaluationCount,
		"evaluationDataAvailable": evaluationDataAvailable,
		"scoreDataAvailable":      scoreDataAvailable,
	}))
}

func extractSemesterStatistics(statsResp map[string]interface{}) (*float64, *float64) {
	if statsResp == nil {
		return nil, nil
	}
	resultMap, ok := statsResp["result"].(map[string]interface{})
	if !ok || resultMap == nil {
		return nil, nil
	}

	var avg *float64
	if v, ok := resultMap["averageScore"]; ok {
		avg = toFloatPointer(v, 2)
	}
	if avg == nil {
		if v, ok := resultMap["avgScore"]; ok {
			avg = toFloatPointer(v, 2)
		}
	}

	var credit *float64
	if v, ok := resultMap["creditTotal"]; ok {
		credit = toFloatPointer(v, 2)
	}
	if credit == nil {
		if v, ok := resultMap["totalCredit"]; ok {
			credit = toFloatPointer(v, 2)
		}
	}

	return avg, credit
}

func toFloatPointer(v interface{}, decimals int) *float64 {
	if v == nil {
		return nil
	}

	var f float64
	switch t := v.(type) {
	case float64:
		f = t
	case float32:
		f = float64(t)
	case int:
		f = float64(t)
	case int64:
		f = float64(t)
	case json.Number:
		if fv, err := t.Float64(); err == nil {
			f = fv
		} else {
			return nil
		}
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		fv, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		f = fv
	default:
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == "" {
			return nil
		}
		fv, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		f = fv
	}

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	pow := math.Pow(10, float64(decimals))
	rounded := math.Round(f*pow) / pow
	return &rounded
}

func formatNumericValue(v interface{}, decimals int) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%.*f", decimals, t)
	case float32:
		return fmt.Sprintf("%.*f", decimals, t)
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return fmt.Sprintf("%.*f", decimals, f)
		}
		if i, err := t.Int64(); err == nil {
			return fmt.Sprintf("%d", i)
		}
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func extractListCount(body []byte) int {
	var arr []interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		return len(arr)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return -1
	}

	result := resp["result"]
	switch r := result.(type) {
	case []interface{}:
		return len(r)
	case map[string]interface{}:
		if records, ok := r["records"].([]interface{}); ok {
			return len(records)
		}
	}

	if data := resp["data"]; data != nil {
		if list, ok := data.([]interface{}); ok {
			return len(list)
		}
		if m, ok := data.(map[string]interface{}); ok {
			if records, ok := m["records"].([]interface{}); ok {
				return len(records)
			}
		}
	}

	return -1
}

// GetUserArchive 获取学生完整档案信息
// 策略：内存缓存优先，缓存未命中则从银行卡接口 + JSON API 拼接
func GetUserArchive(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	hasSchoolAccount := user.Username != "" &&
		user.Password != "" &&
		!strings.HasPrefix(user.Username, "wx_") &&
		user.Username != user.WechatOpenID

	baseInfo := gin.H{
		"id":               fmt.Sprintf("%d", user.ID),
		"username":         user.Username,
		"realname":         user.Realname,
		"nickname":         user.Nickname,
		"avatar":           user.Avatar,
		"avatarUrl":        user.Avatar,
		"email":            user.Email,
		"phone":            user.Phone,
		"birthday":         user.Birthday,
		"sex":              user.Sex,
		"identityCard":     user.IdentityCard,
		"userType":         user.UserType,
		"schoolId":         user.SchoolID,
		"className":        user.ClassName,
		"professionId":     user.ProfessionID,
		"facultyId":        user.FacultyID,
		"gradeId":          user.GradeID,
		"currentSemester":  user.CurrentSemester,
		"hasSchoolAccount": hasSchoolAccount,
		"status":           user.Status,
	}

	if !hasSchoolAccount || user.SchoolToken == "" {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(baseInfo))
		return
	}

	clientID := c.GetHeader("X-Client-Id")
	overrideCookie := c.GetHeader("Cookie")

	archive, err := services.FetchAndCacheArchive(user, clientID, overrideCookie)
	if err != nil {
		log.Printf("[GET-ARCHIVE] fetchAndCache error: userID=%d err=%v", userID, err)
	}

	if archive != nil {
		mergeArchiveToBase(baseInfo, archive)
	}

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(baseInfo))
}

// mergeArchiveToBase 将档案数据合并到 baseInfo
func mergeArchiveToBase(baseInfo gin.H, archive *services.StudentArchive) {
	// 基础资料
	if archive.StudentNo != "" {
		baseInfo["studentNo"] = archive.StudentNo
	}
	if archive.ExamNo != "" {
		baseInfo["examNo"] = archive.ExamNo
	}
	if archive.NameUsedBefore != "" {
		baseInfo["nameUsedBefore"] = archive.NameUsedBefore
	}
	if archive.EntranceScore != "" {
		baseInfo["entranceScore"] = archive.EntranceScore
	}
	if archive.GraduateType != "" {
		baseInfo["graduateType"] = archive.GraduateType
	}
	if archive.NativePlace != "" {
		baseInfo["nativePlace"] = archive.NativePlace
	}
	if archive.GraduateSchool != "" {
		baseInfo["graduateSchool"] = archive.GraduateSchool
	}
	if archive.GraduateForm != "" {
		baseInfo["graduateForm"] = archive.GraduateForm
	}
	if archive.IsMarried != "" {
		baseInfo["isMarried"] = archive.IsMarried
	}
	if archive.SourceProvince != "" {
		baseInfo["sourceProvince"] = archive.SourceProvince
	}
	if archive.Nation != "" {
		baseInfo["nation"] = archive.Nation
	}
	if archive.IsPoorArea != "" {
		baseInfo["isPoorArea"] = archive.IsPoorArea
	}
	if archive.HealthCondition != "" {
		baseInfo["healthCondition"] = archive.HealthCondition
	}

	// 学业信息
	if archive.Campus != "" {
		baseInfo["campus"] = archive.Campus
	}
	if archive.StudentStatus != "" {
		baseInfo["studentStatus"] = archive.StudentStatus
	}
	if archive.FacultyName != "" {
		baseInfo["facultyName"] = archive.FacultyName
	}
	if archive.EnrollmentDate != "" {
		baseInfo["enrollmentDate"] = archive.EnrollmentDate
	}
	if archive.MajorName != "" {
		baseInfo["majorName"] = archive.MajorName
	}
	if archive.ExpectedGradDate != "" {
		baseInfo["expectedGradDate"] = archive.ExpectedGradDate
	}
	if archive.AdminClass != "" {
		baseInfo["adminClass"] = archive.AdminClass
	}
	if archive.StudyForm != "" {
		baseInfo["studyForm"] = archive.StudyForm
	}
	if archive.EducationYears != "" {
		baseInfo["educationYears"] = archive.EducationYears
	}
	if archive.EnrollmentType != "" {
		baseInfo["enrollmentType"] = archive.EnrollmentType
	}
	if archive.Counselor != "" {
		baseInfo["counselor"] = archive.Counselor
	}

	// 联系方式
	if archive.Phone != "" {
		baseInfo["phone"] = archive.Phone
	}
	if archive.Email != "" {
		baseInfo["email"] = archive.Email
	}

	// 银行卡信息
	if archive.BankCardNumber != "" {
		baseInfo["bankCardNumber"] = archive.BankCardNumber
	}
	if archive.BankName != "" {
		baseInfo["bankName"] = archive.BankName
	}
	if archive.BankProvinceCity != "" {
		baseInfo["bankProvinceCity"] = archive.BankProvinceCity
	}
	if archive.BankSubBranch != "" {
		baseInfo["bankSubBranch"] = archive.BankSubBranch
	}
	if archive.BankCardType != "" {
		baseInfo["bankCardType"] = archive.BankCardType
	}
	if archive.CardHolder != "" {
		baseInfo["cardHolder"] = archive.CardHolder
	}

	// 家庭资料
	if archive.FamilyAddress != "" {
		baseInfo["familyAddress"] = archive.FamilyAddress
	}
	if archive.FamilyPhone != "" {
		baseInfo["familyPhone"] = archive.FamilyPhone
	}
	if archive.FamilyPost != "" {
		baseInfo["familyPost"] = archive.FamilyPost
	}
	if len(archive.FamilyMembers) > 0 {
		baseInfo["familyMembers"] = archive.FamilyMembers
	}

	// 学校经历
	if len(archive.SchoolExperiences) > 0 {
		baseInfo["schoolExperiences"] = archive.SchoolExperiences
	}

	// 学业调整
	if len(archive.AcademicChanges) > 0 {
		baseInfo["academicChanges"] = archive.AcademicChanges
	}
}

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(c *gin.Context) {
	// 从上下文获取用户ID
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

	// 查询用户信息
	var user models.User
	result := models.DB.First(&user, userIdStr)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 解析请求参数
	var req struct {
		Avatar   string `json:"avatar"`
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 更新用户信息
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// 保存更新
	updates := make(map[string]interface{})
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if len(updates) > 0 {
		if err := models.UpdateUserFields(user.ID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("更新失败: "+err.Error(), 500))
			return
		}
	}

	// 返回更新后的用户信息
	userInfo := gin.H{
		"id":       fmt.Sprintf("%d", user.ID),
		"username": user.Username,
		"realname": user.Realname,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"email":    user.Email,
	}

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(userInfo))
}
