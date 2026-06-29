package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/utils"
)

// PCMissClassGetList 获取缺课统计列表
func PCMissClassGetList(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	gradeName := c.DefaultQuery("gradeName", "")
	studySystem := c.DefaultQuery("studySystem", "")
	semester := c.DefaultQuery("semester", "")

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 15
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	loginRes := getOrRefreshPCSession(userID)
	if loginRes.NeedManual {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginRes.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginRes.CaptchaData),
			"message":    "请手动输入验证码",
		}))
		return
	}
	if loginRes.SessionID == "" || loginRes.User == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用，请稍后再试", 503))
		return
	}
	jsessionID := loginRes.SessionID
	user := loginRes.User

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := fmt.Sprintf("%s/course/missClass/getList?pageSize=%d&pageNum=%d&gradeName=%s&studySystem=%s&semester=%s",
		client.BaseURL, pageSize, pageNum, gradeName, studySystem, url.QueryEscape(semester))

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	if isHTMLResponse(body) {
		clearPCSession(userID)
		client2 := getPCClientForUser(userID)
		loginResult := tryPCAutoLogin(client2, user)
		if loginResult.NeedManual {
			c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
				"needManual": true,
				"sessionId":  loginResult.CaptchaSID,
				"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginResult.CaptchaData),
				"message":    "请手动输入验证码",
			}))
			return
		}
		if loginResult.SessionID == "" {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
			return
		}
		u2, _ := url.Parse(client2.BaseURL)
		client2.Jar.SetCookies(u2, []*http.Cookie{{Name: "JSESSIONID", Value: loginResult.SessionID}})
		req2, _ := http.NewRequest("POST", apiURL, nil)
		setPCAPIHeaders(req2)
		resp2, err2 := client2.HTTPClient.Do(req2)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			if !isHTMLResponse(body2) {
				var schoolData interface{}
				if err := json.Unmarshal(body2, &schoolData); err == nil {
					c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
					return
				}
			}
		}
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}

// PCMakeupExamQuery 补考查询
func PCMakeupExamQuery(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	semester := c.DefaultQuery("semester", "")

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 15
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	loginRes := getOrRefreshPCSession(userID)
	if loginRes.NeedManual {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginRes.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginRes.CaptchaData),
			"message":    "请手动输入验证码",
		}))
		return
	}
	if loginRes.SessionID == "" || loginRes.User == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用，请稍后再试", 503))
		return
	}
	jsessionID := loginRes.SessionID
	user := loginRes.User

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := fmt.Sprintf("%s/score/getCourseScoreArrangeByPage?pageSize=%d&pageNum=%d&semester=%s&_=%d",
		client.BaseURL, pageSize, pageNum, url.QueryEscape(semester), time.Now().UnixMilli())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	if isHTMLResponse(body) {
		clearPCSession(userID)
		client2 := getPCClientForUser(userID)
		loginResult := tryPCAutoLogin(client2, user)
		if loginResult.NeedManual {
			c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
				"needManual": true,
				"sessionId":  loginResult.CaptchaSID,
				"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginResult.CaptchaData),
				"message":    "请手动输入验证码",
			}))
			return
		}
		if loginResult.SessionID == "" {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
			return
		}
		u2, _ := url.Parse(client2.BaseURL)
		client2.Jar.SetCookies(u2, []*http.Cookie{{Name: "JSESSIONID", Value: loginResult.SessionID}})
		req2, _ := http.NewRequest("GET", apiURL, nil)
		setPCAPIHeaders(req2)
		resp2, err2 := client2.HTTPClient.Do(req2)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			if !isHTMLResponse(body2) {
				var schoolData interface{}
				if err := json.Unmarshal(body2, &schoolData); err == nil {
					c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
					return
				}
			}
		}
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}

// PCDisciplinaryQuery 违纪查询
func PCDisciplinaryQuery(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	disciplinaryTypeID := c.DefaultQuery("learnerdisciplinarytypeid", "")
	disciplinaryLevelID := c.DefaultQuery("learnerdisciplinarylevelid", "")
	createTimeRange := c.DefaultQuery("createTimeRange", "")

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 15
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	loginRes := getOrRefreshPCSession(userID)
	if loginRes.NeedManual {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginRes.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginRes.CaptchaData),
			"message":    "请手动输入验证码",
		}))
		return
	}
	if loginRes.SessionID == "" || loginRes.User == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用，请稍后再试", 503))
		return
	}
	jsessionID := loginRes.SessionID
	user := loginRes.User

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := fmt.Sprintf("%s/learner/disciplinary/getDisciplinaryByPage?pageSize=%d&pageNum=%d&learnerdisciplinarytypeid=%s&learnerdisciplinarylevelid=%s&createTimeRange=%s&_=%d",
		client.BaseURL, pageSize, pageNum, disciplinaryTypeID, disciplinaryLevelID, url.QueryEscape(createTimeRange), time.Now().UnixMilli())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	if isHTMLResponse(body) {
		clearPCSession(userID)
		client2 := getPCClientForUser(userID)
		loginResult := tryPCAutoLogin(client2, user)
		if loginResult.NeedManual {
			c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
				"needManual": true,
				"sessionId":  loginResult.CaptchaSID,
				"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginResult.CaptchaData),
				"message":    "请手动输入验证码",
			}))
			return
		}
		if loginResult.SessionID == "" {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
			return
		}
		u2, _ := url.Parse(client2.BaseURL)
		client2.Jar.SetCookies(u2, []*http.Cookie{{Name: "JSESSIONID", Value: loginResult.SessionID}})
		req2, _ := http.NewRequest("GET", apiURL, nil)
		setPCAPIHeaders(req2)
		resp2, err2 := client2.HTTPClient.Do(req2)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			if !isHTMLResponse(body2) {
				var schoolData interface{}
				if err := json.Unmarshal(body2, &schoolData); err == nil {
					c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
					return
				}
			}
		}
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}

// PCDisciplinaryTypes 获取违纪类型列表
func PCDisciplinaryTypes(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	loginRes := getOrRefreshPCSession(userID)
	if loginRes.NeedManual {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginRes.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginRes.CaptchaData),
			"message":    "请手动输入验证码",
		}))
		return
	}
	if loginRes.SessionID == "" || loginRes.User == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}
	jsessionID := loginRes.SessionID

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := client.BaseURL + "/learner/disciplinary/getDisciplinaryTypeList"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	if isHTMLResponse(body) {
		clearPCSession(userID)
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}

// PCDisciplinaryLevels 获取违纪级别列表
func PCDisciplinaryLevels(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	loginRes := getOrRefreshPCSession(userID)
	if loginRes.NeedManual {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginRes.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginRes.CaptchaData),
			"message":    "请手动输入验证码",
		}))
		return
	}
	if loginRes.SessionID == "" || loginRes.User == nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}
	jsessionID := loginRes.SessionID

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := client.BaseURL + "/learner/disciplinary/getDisciplinaryLevelList"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	if isHTMLResponse(body) {
		clearPCSession(userID)
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}
