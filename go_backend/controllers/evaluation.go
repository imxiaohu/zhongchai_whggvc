package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// EvaluationItem 评教项目结构
type EvaluationItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	Options     []struct {
		Value int    `json:"value"`
		Label string `json:"label"`
	} `json:"options"`
}

// GetEvaluationList 获取评教列表
func GetEvaluationList(c *gin.Context) {
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
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("用户ID格式错误", 500))
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()
	clientID := c.GetHeader("X-Client-Id")
	// 评教接口需要学校系统的 token，强制使用自动登录获取的 SchoolToken，忽略前端传入的业务 JWT
	overrideToken := ""
	overrideCookie := c.GetHeader("Cookie")

	// 若没有任何学校 token，避免直接请求导致 400，提示前端重新登录学校账号
	if user.SchoolToken == "" {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("学校账号未登录或令牌缺失，请重新登录学校账号", 502))
		return
	}

	// 直接代理学校评教列表接口（使用学校端实际可用的路径）
	params := url.Values{}
	if clientID != "" {
		params.Set("clientId", clientID)
	}

	body, err := proxyRequestWithClientID(proxyClient, user, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", clientID, overrideToken, overrideCookie, params)
	if err != nil {
		fmt.Printf("[EvaluationList] proxy error: %v\n", err)
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取评教列表失败: "+err.Error(), 502))
		return
	}

	// 若学校返回 HTML（常见登录页/错误页），再重打一次 init 后重试
	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		fmt.Printf("[EvaluationList] upstream returned HTML: %.200s\n", bodyStr)
		if err := triggerSchoolInit(proxyClient, user, clientID, overrideToken, overrideCookie); err == nil {
			body, err = proxyRequestWithClientID(proxyClient, user, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", clientID, overrideToken, overrideCookie, params)
			if err == nil {
				bodyStr = strings.TrimSpace(string(body))
			}
		}
	}

	// 仍然是 HTML 则报错
	if strings.HasPrefix(bodyStr, "<") {
		fmt.Printf("[EvaluationList] upstream returned HTML after retry: %.200s\n", bodyStr)
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取评教列表失败：学校接口返回非JSON（可能未登录或服务器异常）", 502))
		return
	}

	// 学校接口有两种可能：直接数组，或包装对象
	var arr []interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"result":  arr,
		})
		return
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("[EvaluationList] unmarshal object failed: %v, raw: %.200s\n", err, bodyStr)
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("解析评教列表失败: "+err.Error(), 502))
		return
	}

	if success, ok := resp["success"].(bool); ok && !success {
		msg, _ := resp["message"].(string)
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": msg,
			"result":  []interface{}{},
		})
		return
	}

	if result, ok := resp["result"]; ok && result != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"result":  result,
		})
		return
	}

	// 如果没有 result 字段，直接透传完整响应，避免误判为空
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"result":  resp,
	})
}

// GetEvaluationNorm 获取评教指标
func GetEvaluationNorm(c *gin.Context) {
	// 从上下文获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
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

	evalID := c.Param("id")
	if evalID == "" {
		evalID = c.Query("id")
	}
	if evalID == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("缺少评教ID", 400))
		return
	}

	proxyClient := utils.NewProxyClient()
	clientID := c.GetHeader("X-Client-Id")
	overrideToken := "" // 强制使用学校 token
	overrideCookie := c.GetHeader("Cookie")

	// 优先从新的列表接口获取题目，避免旧接口 502/HTML
	if detail := fetchEvaluationDetailFromList(proxyClient, user, clientID, overrideToken, overrideCookie, evalID); detail != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"result":  detail,
		})
		return
	}

	// 兜底：尝试旧接口
	path := fmt.Sprintf("/scloud/educational/evaluation/getEvaluationNorm/%s", evalID)
	params := url.Values{}
	if clientID != "" {
		params.Set("clientId", clientID)
	}
	body, err := proxyRequestWithClientID(proxyClient, user, "GET", path, clientID, overrideToken, overrideCookie, params)
	if err != nil {
		fmt.Printf("[EvaluationNorm] proxy error: %v\n", err)
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取评教指标失败: "+err.Error(), 502))
		return
	}

	// 若学校返回 HTML（常见登录页/错误页），直接提示
	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		fmt.Printf("[EvaluationNorm] upstream returned HTML: %.200s\n", bodyStr)
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("获取评教指标失败：学校接口返回非JSON（可能未登录或服务器异常）", 502))
		return
	}

	// 可能直接返回数组
	var arr []interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"result":  arr,
		})
		return
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Printf("[EvaluationNorm] unmarshal object failed: %v, raw: %.200s\n", err, bodyStr)
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("解析评教指标失败: "+err.Error(), 502))
		return
	}
	result, ok := resp["result"]
	if !ok || result == nil {
		result = []interface{}{}
	}
	// 若 result 是空数组，说明学校没有返回评教指标
	if arr, isArr := result.([]interface{}); isArr && len(arr) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"result":  []interface{}{},
			"_empty":  true,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ok",
		"result":  result,
	})
}

// 从新列表接口中查找具体评教项的题目
func fetchEvaluationDetailFromList(proxyClient *utils.ProxyClient, user *models.User, clientID, overrideToken, overrideCookie, evalID string) interface{} {
	// 复用列表请求
	params := url.Values{}
	if clientID != "" {
		params.Set("clientId", clientID)
	}
	body, err := proxyRequestWithClientID(proxyClient, user, "GET", "/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList", clientID, overrideToken, overrideCookie, params)
	if err != nil {
		return nil
	}

	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		return nil
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err != nil {
		// 尝试包装对象
		var obj map[string]interface{}
		if err := json.Unmarshal(body, &obj); err != nil {
			return nil
		}
		if res, ok := obj["result"].([]interface{}); ok {
			for _, v := range res {
				if m, ok := v.(map[string]interface{}); ok {
					arr = append(arr, m)
				}
			}
		} else {
			return nil
		}
	}

	for _, item := range arr {
		// 支持多种可能的 ID 字段名称
		cfgID := getStringFromAny(item["courseTeachingEvaluationStudentConfigID"])
		if cfgID == "" {
			cfgID = getStringFromAny(item["courseTeachingEvaluationStudentConfigId"])
		}
		itemID := getStringFromAny(item["id"])
		taskID := getStringFromAny(item["courseTeachingClassTaskID"])
		if taskID == "" {
			taskID = getStringFromAny(item["courseTeachingClassTaskId"])
		}

		idMatch := cfgID == evalID || itemID == evalID || taskID == evalID
		if !idMatch {
			// 尝试数值比较（处理学校接口返回数字类型的情况）
			cfgNum := getIntFromAny(item["courseTeachingEvaluationStudentConfigID"])
			if cfgNum == 0 {
				cfgNum = getIntFromAny(item["courseTeachingEvaluationStudentConfigId"])
			}
			itemNum := getIntFromAny(item["id"])
			taskNum := getIntFromAny(item["courseTeachingClassTaskID"])
			if taskNum == 0 {
				taskNum = getIntFromAny(item["courseTeachingClassTaskId"])
			}
			evalNum, _ := strconv.ParseInt(evalID, 10, 64)
			idMatch = cfgNum == evalNum || itemNum == evalNum || taskNum == evalNum
		}
		if !idMatch {
			continue
		}
		if qs, ok := item["courseTeachingEvaluationQuestionsVOS"]; ok {
			return qs
		}
		if qs, ok := item["courseTeachingEvaluationQuestionsDTOS"]; ok {
			return qs
		}
	}
	return nil
}

// 带前端 clientId 透传的代理请求（GET）
func proxyRequestWithClientID(proxyClient *utils.ProxyClient, user *models.User, method, path, clientID, overrideToken, overrideCookie string, params url.Values) ([]byte, error) {
	// 使用统一的 ProxyRequest 方法处理 header/签名/token/重试逻辑
	return proxyClient.ProxyRequestWithClientHeaders(method, path, params, user, clientID, overrideToken, overrideCookie)
}

// triggerSchoolInit 调用学校 init 接口以建立会话
func triggerSchoolInit(proxyClient *utils.ProxyClient, user *models.User, clientID, overrideToken, overrideCookie string) error {
	_, err := proxyClient.ProxyRequestWithClientHeaders("GET", "/scloud/init", nil, user, clientID, overrideToken, overrideCookie)
	return err
}

func getIntFromAny(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case json.Number:
		i, _ := val.Int64()
		return i
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

func getFloatFromAny(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

func getStringFromAny(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case fmt.Stringer:
		return val.String()
	}
	return ""
}

// SubmitEvaluation 一键满分提交，按学校新接口格式透传
func SubmitEvaluation(c *gin.Context) {
	// 请求体为数组
	var payload []map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("请求数据格式错误", 400))
		return
	}

	// 用户校验
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}
	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
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

	// 代理客户端
	proxyClient := utils.NewProxyClient()
	clientID := c.GetHeader("X-Client-Id")
	overrideCookie := c.GetHeader("Cookie")

	// 确保学校会话
	if err := proxyClient.AutoLogin(user, false); err != nil {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("学校账号自动登录失败: "+err.Error(), 502))
		return
	}
	_ = triggerSchoolInit(proxyClient, user, clientID, "", overrideCookie)

	// 按学校接口格式转换
	type Question struct {
		CourseTeachingEvaluationStudentConfigID int         `json:"courseTeachingEvaluationStudentConfigID"`
		CourseTeachingClassTaskID               int         `json:"courseTeachingClassTaskID"`
		// Question 评教题目项（传给学校接口的内层结构）
		OtherContect                            interface{} `json:"otherContect"`
		TeacherID                               int         `json:"teacherID"`
		TeacherName                             string      `json:"teacherName"`
		Score                                   interface{} `json:"score"`
		GetScore                                interface{} `json:"getScore"`
		Content                                 string      `json:"content"`
		CourseTeachingEvaluationNormID          int         `json:"courseTeachingEvaluationNormID"`
		EvaluationNumber                        string      `json:"evaluationNumber"`
	}

	type SubmitItem struct {
		CourseTeachingEvaluationStudentConfigID int         `json:"courseTeachingEvaluationStudentConfigID"`
		CourseTeachingClassTaskID               int         `json:"courseTeachingClassTaskID"`
		OtherContect                            interface{} `json:"otherContect"`
		TeacherID                               int         `json:"teacherID"`
		TeacherName                             string      `json:"teacherName"`
		CourseTeachingEvaluationQuestionsDTOS   []Question  `json:"courseTeachingEvaluationQuestionsDTOS"`
	}

	var submitItems []SubmitItem
	for _, item := range payload {
		// 取 configID
		configID := int(getIntFromAny(item["courseTeachingEvaluationStudentConfigID"]))
		if configID == 0 {
			configID = int(getIntFromAny(item["id"]))
		}
		// class task
		classTaskID := int(getIntFromAny(item["courseTeachingClassTaskID"]))
		teacherID := int(getIntFromAny(item["teacherID"]))
		teacherName, _ := item["teacherName"].(string)
		if teacherName == "" {
			teacherName, _ = item["name"].(string)
		}

		// detail/list of questions
		var questionsSrc []interface{}
		if v, ok := item["detail"].([]interface{}); ok {
			questionsSrc = v
		} else if v, ok := item["courseTeachingEvaluationQuestionsVOS"].([]interface{}); ok {
			questionsSrc = v
		} else if v, ok := item["courseTeachingEvaluationQuestionsDTOS"].([]interface{}); ok {
			questionsSrc = v
		}

		var questions []Question
		for _, q := range questionsSrc {
			qm, ok := q.(map[string]interface{})
			if !ok {
				continue
			}
			score := getFloatFromAny(qm["score"])
			if score == 0 {
				score = 5
			}
			getScore := getFloatFromAny(qm["getScore"])
			if getScore == 0 {
				getScore = score
			}
			questions = append(questions, Question{
				CourseTeachingEvaluationStudentConfigID: configID,
				CourseTeachingClassTaskID:               classTaskID,
				OtherContect:                            qm["otherContect"],
				TeacherID:                               int(getIntFromAny(qm["teacherID"])),
				TeacherName:                             teacherName,
				Score:                                   score,
				GetScore:                                getScore,
				Content:                                 getStringFromAny(qm["content"]),
				CourseTeachingEvaluationNormID:          int(getIntFromAny(qm["courseTeachingEvaluationNormID"])),
				EvaluationNumber:                        getStringFromAny(qm["evaluationNumber"]),
			})
		}

		submitItems = append(submitItems, SubmitItem{
			CourseTeachingEvaluationStudentConfigID: configID,
			CourseTeachingClassTaskID:               classTaskID,
			OtherContect:                            item["otherContect"],
			TeacherID:                               teacherID,
			TeacherName:                             teacherName,
			CourseTeachingEvaluationQuestionsDTOS:   questions,
		})
	}

	// 序列化并代理 POST
	jsonData, err := json.Marshal(submitItems)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("序列化提交数据失败: "+err.Error(), 400))
		return
	}

	req, err := http.NewRequest("POST", proxyClient.BaseURL+"/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/insertResult", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("创建请求失败: "+err.Error(), 502))
		return
	}
	proxyClient.SetCommonHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	if user.SchoolToken != "" {
		req.Header.Set("X-Access-Token", user.SchoolToken)
	}
	if clientID != "" {
		req.Header.Set("X-Client-Id", clientID)
	}
	if overrideCookie != "" {
		req.Header.Set("Cookie", overrideCookie)
	}

	resp, err := proxyClient.HTTPClient.Do(req)
	if err != nil {
		if isNetworkError(err) {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器暂时关闭，请稍后再试", 503))
			return
		}
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("提交评教失败: "+err.Error(), 502))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(body))
	if strings.HasPrefix(bodyStr, "<") {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("提交评教失败：学校接口返回非JSON（可能未登录或服务器异常）", 502))
		return
	}

	var schoolResp map[string]interface{}
	if err := json.Unmarshal(body, &schoolResp); err != nil {
		c.JSON(http.StatusBadGateway, utils.NewStandardErrorResponse("解析学校响应失败: "+err.Error(), 502))
		return
	}

	c.JSON(http.StatusOK, schoolResp)
}


