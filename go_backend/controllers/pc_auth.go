// Package controllers handles PC-only (scloud) endpoints.
// These endpoints are distinct from mobile API endpoints and require JSESSIONID-based session management.
package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// PC端基础域名
const pcBaseURL = "https://xs.whggvc.net:443/scloud"

// PCLoginResult PC登录结果，支持自动登录成功或降级为手动验证
type PCLoginResult struct {
	SessionID   string // 有效会话ID（自动登录成功时）
	User        *models.User
	NeedManual  bool   // 是否需要手动验证（OCR不可用时）
	CaptchaData []byte // 验证码图片原始数据（NeedManual=true时）
	CaptchaSID  string // 验证码对应的JSESSIONID（NeedManual=true时）
}

// PCClient 独立的PC端HTTP客户端，每个用户一个独立的CookieJar
type PCClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Jar        http.CookieJar
}

func newPCClient() *PCClient {
	jar, _ := cookiejar.New(nil)
	return &PCClient{
		BaseURL: pcBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Jar: jar,
	}
}

// PCUserSession PC用户会话结构，用于内存缓存
type PCUserSession struct {
	JSESSIONID string
	LoginTime  time.Time
	ExpireTime time.Time
}

var (
	pcSessionCache = make(map[uint]*PCUserSession)
	sessionLock    sync.RWMutex
)

// PCSessionTTL PC会话有效期（默认30分钟）
const PCSessionTTL = 30 * time.Minute

func getPCSession(userID uint) *PCUserSession {
	sessionLock.RLock()
	defer sessionLock.RUnlock()
	if sess, ok := pcSessionCache[userID]; ok {
		if time.Now().Before(sess.ExpireTime) {
			return sess
		}
	}
	return nil
}

func setPCSession(userID uint, jsessionID string) {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	pcSessionCache[userID] = &PCUserSession{
		JSESSIONID: jsessionID,
		LoginTime:  time.Now(),
		ExpireTime: time.Now().Add(PCSessionTTL),
	}
}

func clearPCSession(userID uint) {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	delete(pcSessionCache, userID)
}

// PCLoginInit 获取PC端登录页面并初始化会话，返回验证码图片URL
func PCLoginInit(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	// 使用用户独立的PCClient
	client := getPCClientForUser(userID)

	// GET /scloud/login - 获取登录页并初始化会话
	req, err := http.NewRequest("GET", client.BaseURL+"/login", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCRequestHeaders(req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败，请稍后再试", 503))
		return
	}
	defer resp.Body.Close()

	// 从响应头提取 JSESSIONID
	jsessionID := extractJSESSIONID(resp.Cookies())
	if jsessionID == "" {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("无法获取会话，请稍后再试", 500))
		return
	}

	// 同步注入 jar，避免后续 captcha/submit 用错 JSESSIONID
	baseU, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(baseU, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID, Path: "/"}})

	// 保存会话到缓存
	setPCSession(userID, jsessionID)

	// 持久化到数据库
	savePCSessionToDB(userID, jsessionID, req.Header.Get("User-Agent"))

	// 返回验证码图片URL（前端直接加载）
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"captchaUrl": client.BaseURL + "/validateCode",
		"sessionId":  jsessionID,
		"message":    "会话初始化成功，请在60秒内完成验证",
	}))
}

// PCLoginSubmit 提交PC端登录表单
func PCLoginSubmit(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	var req struct {
		Captcha   string `json:"captcha" binding:"required"`
		StudentNo string `json:"studentNo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("验证码不能为空", 400))
		return
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 使用用户独立的PCClient
	client := getPCClientForUser(userID)

	// 获取 PC 登录密码（与移动端 mLogin 不同，PC 端 /login 走网页表单，
	// 不依赖 captcha 走自动登录路径，仅在 needManual 时被前端调用）
	pcPassword, err := user.GetSchoolPassword()
	if err != nil || pcPassword == "" {
		// 兼容历史账号没有 PC 密码的场景：仍退回"用学号"（与旧实现一致）
		pcPassword = user.Username
	}

	// 获取已有JSESSIONID（优先从缓存，其次从数据库）
	session := getPCSession(userID)
	var jsessionID string
	if session != nil {
		jsessionID = session.JSESSIONID
	}
	if jsessionID == "" {
		jsessionID = user.PCJSESSIONID
	}

	if jsessionID == "" {
		// 没有会话，先初始化
		initReq, _ := http.NewRequest("GET", client.BaseURL+"/login", nil)
		setPCRequestHeaders(initReq)
		initResp, err := client.HTTPClient.Do(initReq)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
			return
		}
		jsessionID = extractJSESSIONID(initResp.Cookies())
		initResp.Body.Close()
		if jsessionID == "" {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("无法建立会话，请重试", 500))
			return
		}
		// 同步写入 jar，避免后续请求拿到不同 JSESSIONID
		u, _ := url.Parse(client.BaseURL)
		client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID, Path: "/"}})
	}

	// POST /scloud/login - 提交登录表单（手动降级路径，用户已输入 captcha）
	formData := url.Values{}
	formData.Set("username", user.Username)
	formData.Set("password", pcPassword) // 使用保存的 PC 密码；缺省降级为学号
	formData.Set("randomcode", req.Captcha)

	loginReq, err := http.NewRequest("POST", client.BaseURL+"/login", strings.NewReader(formData.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCLoginSubmitHeaders(loginReq, jsessionID)

	loginResp, err := client.HTTPClient.Do(loginReq)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	defer loginResp.Body.Close()

	// 判断登录结果：302重定向到/index表示成功
	location := loginResp.Header.Get("Location")
	if strings.Contains(location, "/index") {
		// 登录成功，保存会话
		newJsessionID := extractJSESSIONID(loginResp.Cookies())
		if newJsessionID == "" {
			newJsessionID = jsessionID
		}
		setPCSession(userID, newJsessionID)
		savePCSessionToDB(userID, newJsessionID, loginReq.Header.Get("User-Agent"))

		log.Printf("[PC-LOGIN] userID=%d login success, jsessionID=%s", userID, newJsessionID[:8]+"...")
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"success":   true,
			"message":   "PC端登录成功",
			"sessionId": newJsessionID,
		}))
		return
	}

	// 登录失败
	clearPCSession(userID)
	c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("验证码错误或登录失败，请重试", 401))
}

// PCGetCaptchaImage 获取验证码图片（代理）
// 修复：用户没有 PC 会话时，先自动 init 一份；不再允许"裸拿验证码"
func PCGetCaptchaImage(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	client := getPCClientForUser(userID)

	session := getPCSession(userID)
	var jsessionID string
	if session != nil {
		jsessionID = session.JSESSIONID
	}
	if jsessionID == "" {
		// 没有会话：自动 init，避免前端拿到一个属于孤儿会话的验证码导致提交必失败
		initReq, _ := http.NewRequest("GET", client.BaseURL+"/login", nil)
		setPCRequestHeaders(initReq)
		initResp, err := client.HTTPClient.Do(initReq)
		if err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		jsessionID = extractJSESSIONID(initResp.Cookies())
		initResp.Body.Close()
		if jsessionID != "" {
			baseU, _ := url.Parse(client.BaseURL)
			client.Jar.SetCookies(baseU, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID, Path: "/"}})
			setPCSession(userID, jsessionID)
			savePCSessionToDB(userID, jsessionID, "")
		}
	}

	req, err := http.NewRequest("GET", client.BaseURL+"/validateCode", nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	setPCRequestHeaders(req)
	if jsessionID != "" {
		req.Header.Set("Cookie", "JSESSIONID="+jsessionID)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 透传响应头（过滤掉会污染客户端 jar 的 Set-Cookie，避免 jar 中混入不一致的会话）
	for name, values := range resp.Header {
		if strings.EqualFold(name, "Set-Cookie") {
			continue
		}
		for _, value := range values {
			c.Header(name, value)
		}
	}
	io.Copy(c.Writer, resp.Body)
}

// PCGetTeachers 获取我的老师列表（PC端接口）
func PCGetTeachers(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	// 获取分页参数
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "15"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	gradeName := c.DefaultQuery("gradeName", "")

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 15
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	// 获取PC会话（不存在时后端自动尝试登录），复用返回的user避免重复查询
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

	// 复用用户专属Client（带独立CookieJar）
	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := fmt.Sprintf("%s/courseOptionTeacher/getTeacherByPage?pageSize=%d&pageNum=%d&gradeName=%s&currentsemester=&_=%d",
		client.BaseURL, pageSize, pageNum, gradeName, time.Now().UnixMilli())

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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	log.Printf("[DEBUG-PC] userID=%d teachers: first response status=%d body_len=%d isHTML=%v", userID, resp.StatusCode, len(body), isHTMLResponse(body))
	if isHTMLResponse(body) {
		log.Printf("[DEBUG-PC] userID=%d teachers: got HTML response (session expired), retrying", userID)
		clearPCSession(userID)
		// 重试一次登录
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
				log.Printf("[DEBUG-PC] userID=%d teachers: retry success, body_len=%d", userID, len(body2))
				var schoolData2 interface{}
				if err := json.Unmarshal(body2, &schoolData2); err == nil {
					c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData2))
				} else {
					c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body2)))
				}
				return
			}
		}
		log.Printf("[DEBUG-PC] userID=%d teachers: retry failed, returning 503", userID)
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	log.Printf("[DEBUG-PC] userID=%d teachers: success body_len=%d", userID, len(body))
	// 包装为标准响应格式 {success: true, result: <school_data>}
	var schoolData interface{}
	if err := json.Unmarshal(body, &schoolData); err == nil {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(schoolData))
	} else {
		c.JSON(http.StatusOK, utils.NewSuccessResponse(string(body)))
	}
}

// PCGetStudentInfo 获取学生档案信息（从PC端HTML页面解析）
func PCGetStudentInfo(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}
	_ = user

	// 获取PC会话（不存在时后端自动尝试登录）
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
	if loginRes.SessionID == "" {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用，请稍后再试", 503))
		return
	}
	jsessionID := loginRes.SessionID

	// 复用用户专属Client（带独立CookieJar）
	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	// 调用PC端档案页面
	apiURL := client.BaseURL + "/student/base"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCRequestHeaders(req)
	req.Header.Set("Referer", client.BaseURL+"/scloud/index")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取响应失败", 500))
		return
	}

	// 检查是否需要重试（会话过期时返回登录页HTML）
	if isHTMLResponse(body) && strings.Contains(string(body), "login") {
		log.Printf("[DEBUG-PC] userID=%d student-info: got login page, session expired", userID)
		clearPCSession(userID)
		// 重试一次登录
		client2 := getPCClientForUser(userID)
		user2, _ := models.FindUserByID(userID)
		loginResult := tryPCAutoLogin(client2, user2)
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
		setPCRequestHeaders(req2)
		req2.Header.Set("Referer", client2.BaseURL+"/scloud/index")
		resp2, err2 := client2.HTTPClient.Do(req2)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			if !isHTMLResponse(body2) || !strings.Contains(string(body2), "login") {
				data := parseArchiveHTML(string(body2))
				c.JSON(http.StatusOK, utils.NewSuccessResponse(data))
				return
			}
			resp2.Body.Close()
		}
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用", 503))
		return
	}

	// 解析HTML并返回结构化数据
	data := parseArchiveHTML(string(body))
	c.JSON(http.StatusOK, utils.NewSuccessResponse(data))
}

// parseArchiveHTML 解析PC端档案HTML页面，提取表单数据
func parseArchiveHTML(html string) map[string]interface{} {
	result := make(map[string]interface{})

	// 基础资料 - 学号（在第二个 <div class="row"> 中查找）
	// 学号在第2个row的div中
	studentNoRe := regexp.MustCompile(`(?s)<div class="row">.*?<label>学号：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := studentNoRe.FindStringSubmatch(html); len(matches) > 1 {
		result["studentNo"] = strings.TrimSpace(matches[1])
	}

	// 考生号
	candidateNoRe := regexp.MustCompile(`(?s)<label>考生号：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := candidateNoRe.FindStringSubmatch(html); len(matches) > 1 {
		result["candidateNo"] = strings.TrimSpace(matches[1])
	}

	// 姓名
	nameRe := regexp.MustCompile(`(?s)<label>姓名：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := nameRe.FindStringSubmatch(html); len(matches) > 1 {
		result["name"] = strings.TrimSpace(matches[1])
	}

	// 曾用名
	nameUsedBeforeRe := regexp.MustCompile(`name="nameUsedBefore"[^>]*value="([^"]*)"`)
	if matches := nameUsedBeforeRe.FindStringSubmatch(html); len(matches) > 1 {
		result["nameUsedBefore"] = strings.TrimSpace(matches[1])
	}

	// 性别
	genderRe := regexp.MustCompile(`(?s)<label>性别：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := genderRe.FindStringSubmatch(html); len(matches) > 1 {
		result["gender"] = strings.TrimSpace(matches[1])
	}

	// 入学总分
	entranceScoreRe := regexp.MustCompile(`(?s)<label>入学总分：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := entranceScoreRe.FindStringSubmatch(html); len(matches) > 1 {
		result["entranceScore"] = strings.TrimSpace(matches[1])
	}

	// 出生日期
	birthDateRe := regexp.MustCompile(`(?s)<label>出生日期：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := birthDateRe.FindStringSubmatch(html); len(matches) > 1 {
		result["birthDate"] = strings.TrimSpace(matches[1])
	}

	// 籍贯
	nativePlaceRe := regexp.MustCompile(`(?s)<label>籍贯：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]*)</label>`)
	if matches := nativePlaceRe.FindStringSubmatch(html); len(matches) > 1 {
		result["nativePlace"] = strings.TrimSpace(matches[1])
	}

	// 毕业学校
	graduatedSchoolRe := regexp.MustCompile(`(?s)<label>毕业学校：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]*)</label>`)
	if matches := graduatedSchoolRe.FindStringSubmatch(html); len(matches) > 1 {
		result["graduatedSchool"] = strings.TrimSpace(matches[1])
	}

	// 证件号
	idCardRe := regexp.MustCompile(`(?s)<label>证件号：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := idCardRe.FindStringSubmatch(html); len(matches) > 1 {
		result["idCard"] = strings.TrimSpace(matches[1])
	}

	// 婚否
	isMarriedRe := regexp.MustCompile(`name="isMarried"[^>]*>.*?<option[^>]*selected[^>]*value="(\d)"`)
	if matches := isMarriedRe.FindStringSubmatch(html); len(matches) > 1 {
		result["isMarried"] = matches[1]
	}

	// 生源省市
	sourceProvinceRe := regexp.MustCompile(`(?s)<label>生源省市：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label[^>]*>([^<]*)</label>`)
	if matches := sourceProvinceRe.FindStringSubmatch(html); len(matches) > 1 {
		result["sourceProvince"] = strings.TrimSpace(matches[1])
	}

	// 民族
	nationRe := regexp.MustCompile(`id="myNation"[^>]*value="([^"]*)"`)
	if matches := nationRe.FindStringSubmatch(html); len(matches) > 1 {
		result["nation"] = strings.TrimSpace(matches[1])
	}

	// 是否贫困地区
	isPoorAreaRe := regexp.MustCompile(`(?s)<label>是否贫困地区：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label[^>]*>([^<]*)</label>`)
	if matches := isPoorAreaRe.FindStringSubmatch(html); len(matches) > 1 {
		result["isPoorArea"] = strings.TrimSpace(matches[1])
	}

	// 政治面貌
	politicalStatusRe := regexp.MustCompile(`(?s)<label>政治面貌：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label[^>]*>([^<]+)</label>`)
	if matches := politicalStatusRe.FindStringSubmatch(html); len(matches) > 1 {
		result["politicalStatus"] = strings.TrimSpace(matches[1])
	}

	// 健康状况
	healthConditionRe := regexp.MustCompile(`id="myHealthCondition"[^>]*value="([^"]*)"`)
	if matches := healthConditionRe.FindStringSubmatch(html); len(matches) > 1 {
		result["healthCondition"] = strings.TrimSpace(matches[1])
	}

	// ========== 学业信息 ==========
	// 校区
	campusRe := regexp.MustCompile(`(?s)<label>校区：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := campusRe.FindStringSubmatch(html); len(matches) > 1 {
		result["campus"] = strings.TrimSpace(matches[1])
	}

	// 学籍状态
	studentStatusRe := regexp.MustCompile(`(?s)<label>学籍状态：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := studentStatusRe.FindStringSubmatch(html); len(matches) > 1 {
		result["studentStatus"] = strings.TrimSpace(matches[1])
	}

	// 院系名称
	departmentRe := regexp.MustCompile(`(?s)<label>院系名称：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := departmentRe.FindStringSubmatch(html); len(matches) > 1 {
		result["department"] = strings.TrimSpace(matches[1])
	}

	// 入学日期
	entranceDateRe := regexp.MustCompile(`(?s)<label>入学日期：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := entranceDateRe.FindStringSubmatch(html); len(matches) > 1 {
		result["entranceDate"] = strings.TrimSpace(matches[1])
	}

	// 专业名称
	majorRe := regexp.MustCompile(`(?s)<label>专业名称：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := majorRe.FindStringSubmatch(html); len(matches) > 1 {
		result["major"] = strings.TrimSpace(matches[1])
	}

	// 预计毕业日期
	expectedGraduationDateRe := regexp.MustCompile(`(?s)<label>预计毕业日期：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label[^>]*>([^<]*)</label>`)
	if matches := expectedGraduationDateRe.FindStringSubmatch(html); len(matches) > 1 {
		result["expectedGraduationDate"] = strings.TrimSpace(matches[1])
	}

	// 行政班
	classNameRe := regexp.MustCompile(`(?s)<label>行政班：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := classNameRe.FindStringSubmatch(html); len(matches) > 1 {
		result["className"] = strings.TrimSpace(matches[1])
	}

	// 学习形式
	learningFormRe := regexp.MustCompile(`(?s)<label>学习形式：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := learningFormRe.FindStringSubmatch(html); len(matches) > 1 {
		result["learningForm"] = strings.TrimSpace(matches[1])
	}

	// 年级
	gradeRe := regexp.MustCompile(`(?s)<label>年级：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := gradeRe.FindStringSubmatch(html); len(matches) > 1 {
		result["grade"] = strings.TrimSpace(matches[1])
	}

	// 学制
	academicSystemRe := regexp.MustCompile(`(?s)<label>学制：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := academicSystemRe.FindStringSubmatch(html); len(matches) > 1 {
		result["academicSystem"] = strings.TrimSpace(matches[1])
	}

	// 录取类别
	admissionCategoryRe := regexp.MustCompile(`(?s)<label>录取类别：</label>\s*</div>\s*<div class="col-md-3[^"]*">\s*<label>([^<]+)</label>`)
	if matches := admissionCategoryRe.FindStringSubmatch(html); len(matches) > 1 {
		result["admissionCategory"] = strings.TrimSpace(matches[1])
	}

	// ========== 联系方式 ==========
	// 手机号码
	phoneRe := regexp.MustCompile(`name="phoneNumber"[^>]*value="([^"]*)"`)
	if matches := phoneRe.FindStringSubmatch(html); len(matches) > 1 {
		result["phoneNumber"] = strings.TrimSpace(matches[1])
	}

	// 联系地址
	personalAddressRe := regexp.MustCompile(`name="personalAddress"[^>]*value="([^"]*)"`)
	if matches := personalAddressRe.FindStringSubmatch(html); len(matches) > 1 {
		result["personalAddress"] = strings.TrimSpace(matches[1])
	}

	// EMAIL
	emailRe := regexp.MustCompile(`name="email"[^>]*value="([^"]*)"`)
	if matches := emailRe.FindStringSubmatch(html); len(matches) > 1 {
		result["email"] = strings.TrimSpace(matches[1])
	}

	// QQ
	qqRe := regexp.MustCompile(`name="qq"[^>]*value="([^"]*)"`)
	if matches := qqRe.FindStringSubmatch(html); len(matches) > 1 {
		result["qq"] = strings.TrimSpace(matches[1])
	}

	// ========== 户口情况 ==========
	// 户口类型
	householdTypeRe := regexp.MustCompile(`id="myHouseholdRegistrationType"[^>]*value="([^"]*)"`)
	if matches := householdTypeRe.FindStringSubmatch(html); len(matches) > 1 {
		result["householdRegistrationType"] = strings.TrimSpace(matches[1])
	}

	// 户口登记机关
	householdOfficeRe := regexp.MustCompile(`name="householdRegistrationOffice"[^>]*value="([^"]*)"`)
	if matches := householdOfficeRe.FindStringSubmatch(html); len(matches) > 1 {
		result["householdRegistrationOffice"] = strings.TrimSpace(matches[1])
	}

	// 户籍地-省
	householdProvinceRe := regexp.MustCompile(`id="myHouseholdRegistrationProvinceID"[^>]*value="([^"]*)"`)
	if matches := householdProvinceRe.FindStringSubmatch(html); len(matches) > 1 {
		result["householdRegistrationProvinceID"] = strings.TrimSpace(matches[1])
	}

	// 户籍地-市
	householdCityRe := regexp.MustCompile(`id="myHouseholdRegistrationCityID"[^>]*value="([^"]*)"`)
	if matches := householdCityRe.FindStringSubmatch(html); len(matches) > 1 {
		result["householdRegistrationCityID"] = strings.TrimSpace(matches[1])
	}

	// 户籍地址
	householdAddressRe := regexp.MustCompile(`name="householdRegistrationAddress"[^>]*value="([^"]*)"`)
	if matches := householdAddressRe.FindStringSubmatch(html); len(matches) > 1 {
		result["householdRegistrationAddress"] = strings.TrimSpace(matches[1])
	}

	// 火车乘车区间-出发站
	trainStartRe := regexp.MustCompile(`name="trainStartStation"[^>]*value="([^"]*)"`)
	if matches := trainStartRe.FindStringSubmatch(html); len(matches) > 1 {
		result["trainStartStation"] = strings.TrimSpace(matches[1])
	}

	// 火车乘车区间-到达站
	trainStopRe := regexp.MustCompile(`name="trainStopStation"[^>]*value="([^"]*)"`)
	if matches := trainStopRe.FindStringSubmatch(html); len(matches) > 1 {
		result["trainStopStation"] = strings.TrimSpace(matches[1])
	}

	// 是否随迁子女
	isTrailingRe := regexp.MustCompile(`id="myIsTrailingChildren"[^>]*value="([^"]*)"`)
	if matches := isTrailingRe.FindStringSubmatch(html); len(matches) > 1 {
		result["isTrailingChildren"] = strings.TrimSpace(matches[1])
	}

	// ========== 家庭资料 ==========
	// 家庭地址
	familyAddressRe := regexp.MustCompile(`name="familyAddress"[^>]*value="([^"]*)"`)
	if matches := familyAddressRe.FindStringSubmatch(html); len(matches) > 1 {
		result["familyAddress"] = strings.TrimSpace(matches[1])
	}

	// 家庭电话-区号
	familyAreaCodeRe := regexp.MustCompile(`name="familyPhoneAreaCode"[^>]*value="([^"]*)"`)
	if matches := familyAreaCodeRe.FindStringSubmatch(html); len(matches) > 1 {
		result["familyPhoneAreaCode"] = strings.TrimSpace(matches[1])
	}

	// 家庭电话
	familyPhoneRe := regexp.MustCompile(`name="familyPhone"[^>]*value="([^"]*)"`)
	if matches := familyPhoneRe.FindStringSubmatch(html); len(matches) > 1 {
		result["familyPhone"] = strings.TrimSpace(matches[1])
	}

	// 邮政编码
	familyPostRe := regexp.MustCompile(`name="familyPost"[^>]*value="([^"]*)"`)
	if matches := familyPostRe.FindStringSubmatch(html); len(matches) > 1 {
		result["familyPost"] = strings.TrimSpace(matches[1])
	}

	// ========== 学业调整信息（异动）==========
	// 解析异动表格
	errors := parseAdjustmentTable(html)
	if len(errors) > 0 {
		result["adjustments"] = errors
	}

	return result
}

// parseAdjustmentTable 解析学业调整信息表格
func parseAdjustmentTable(html string) []map[string]string {
	var adjustments []map[string]string

	// 匹配异动表格行
	// <tr>...<td>异动信息</td><td>异动详情</td><td>操作人</td><td>异动日期</td><td>操作日期</td></tr>
	rowRe := regexp.MustCompile(`(?s)<tr>\s*<td[^>]*>\s*([^<]*)\s*</td>\s*<td[^>]*>\s*([^<]*)\s*</td>\s*<td[^>]*>\s*([^<]*)\s*</td>\s*<td[^>]*>\s*([^<]*)\s*</td>\s*<td[^>]*>\s*([^<]*)\s*</td>\s*</tr>`)

	matches := rowRe.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			adjustment := map[string]string{
				"type":       strings.TrimSpace(match[1]),
				"detail":     strings.TrimSpace(match[2]),
				"operator":   strings.TrimSpace(match[3]),
				"changeDate": strings.TrimSpace(match[4]),
				"opDate":     strings.TrimSpace(match[5]),
			}
			// 跳过表头
			if adjustment["type"] != "异动信息" && adjustment["type"] != "" {
				adjustments = append(adjustments, adjustment)
			}
		}
	}

	return adjustments
}

// PCGetSessionStatus 查询PC会话状态
func PCGetSessionStatus(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	session := getPCSession(userID)
	if session != nil {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"loggedIn":   true,
			"loginAt":    session.LoginTime.Format("2006-01-02 15:04:05"),
			"expiresAt":  session.ExpireTime.Format("2006-01-02 15:04:05"),
			"ttlSeconds": int(time.Until(session.ExpireTime).Seconds()),
		}))
		return
	}

	user, _ := models.FindUserByID(userID)
	if user != nil && user.PCJSESSIONID != "" && time.Now().Before(user.PCExpireTime) {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"loggedIn":   true,
			"loginAt":    user.PCLoginTime.Format("2006-01-02 15:04:05"),
			"expiresAt":  user.PCExpireTime.Format("2006-01-02 15:04:05"),
			"ttlSeconds": int(time.Until(user.PCExpireTime).Seconds()),
			"fromDB":     true,
		}))
		return
	}

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"loggedIn":   false,
		"captchaUrl": pcBaseURL + "/validateCode",
	}))
}

// PCLogout 登出PC端会话
func PCLogout(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	clearPCSession(userID)
	clearPCSessionFromDB(userID)

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"message": "PC会话已清除",
	}))
}

// PCAutoLogin 全自动PC端登录（OCR 自动识别 + 重试，仅在所有重试都失败时降级手动）
// 成功时返回凭证信息，前端需存储到 Storage
func PCAutoLogin(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	user, err := models.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewStandardErrorResponse("用户不存在", 404))
		return
	}

	// 获取 PC 密码
	if _, err := user.GetSchoolPassword(); err != nil {
		log.Printf("[PC-AUTO] userID=%d GetSchoolPassword failed: %v", userID, err)
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("PC登录密码未设置，请在设置中绑定学校账号密码", 400))
		return
	}

	// 走统一的自动登录（包含 5 次 OCR 重试，仅在全部失败时降级手动）
	client := getPCClientForUser(userID)
	loginResult := tryPCAutoLogin(client, user)

	if loginResult.NeedManual {
		// OCR 重试全部失败（极少见）：降级让用户手动输入
		// 保存 latest 会话
		if loginResult.CaptchaSID != "" {
			setPCSession(userID, loginResult.CaptchaSID)
		}
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"needManual": true,
			"sessionId":  loginResult.CaptchaSID,
			"captcha":    "data:image/png;base64," + base64.StdEncoding.EncodeToString(loginResult.CaptchaData),
			"message":    "自动识别多次失败，请在弹窗中输入验证码",
		}))
		return
	}

	if loginResult.SessionID != "" {
		expireTime := time.Now().Add(PCSessionTTL)
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"success":    true,
			"message":    "PC端登录成功（自动识别）",
			"sessionId":  loginResult.SessionID,
			"expireTime": expireTime.Format("2006-01-02T15:04:05Z"),
			"ttlSeconds": int(PCSessionTTL.Seconds()),
		}))
		return
	}

	c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("自动登录失败，请稍后重试", 500))
}

// PCGetSessionCredentials 获取会话凭证（供前端存储到 Storage）
func PCGetSessionCredentials(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	session := getPCSession(userID)
	if session != nil {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"sessionId":  session.JSESSIONID,
			"loginAt":    session.LoginTime.Format(time.RFC3339),
			"expireAt":   session.ExpireTime.Format(time.RFC3339),
			"ttlSeconds": int(time.Until(session.ExpireTime).Seconds()),
			"stored":     true,
		}))
		return
	}

	user, _ := models.FindUserByID(userID)
	if user != nil && user.PCJSESSIONID != "" && time.Now().Before(user.PCExpireTime) {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"sessionId":  user.PCJSESSIONID,
			"loginAt":    user.PCLoginTime.Format(time.RFC3339),
			"expireAt":   user.PCExpireTime.Format(time.RFC3339),
			"ttlSeconds": int(time.Until(user.PCExpireTime).Seconds()),
			"stored":     true,
		}))
		return
	}

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"stored":     false,
		"ttlSeconds": 0,
	}))
}

// PCSetSessionCredentials 前端存储的凭证同步到后端
// 前端从 Storage 恢复会话时调用，后端验证并恢复内存缓存
func PCSetSessionCredentials(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
		ExpireAt  string `json:"expireAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("sessionId不能为空", 400))
		return
	}

	expireTime, err := time.Parse(time.RFC3339, req.ExpireAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("expireAt格式错误", 400))
		return
	}

	if time.Now().After(expireTime) {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"valid":   false,
			"message": "会话已过期，请重新授权",
		}))
		return
	}

	// 验证 JSESSIONID 是否仍然有效
	client := newPCClient()
	u, _ := url.Parse(client.BaseURL + "/index")
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: req.SessionID}})
	client.HTTPClient.Jar = jar

	// 发送一个 XHR 请求验证会话
	checkReq, _ := http.NewRequest("GET", client.BaseURL+"/student/base/getStudentInfo", nil)
	setPCAPIHeaders(checkReq)
	checkReq.Header.Set("Cookie", "JSESSIONID="+req.SessionID)
	checkReq.Header.Set("Referer", client.BaseURL+"/index")
	checkReq.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.HTTPClient.Do(checkReq)
	if err != nil || resp == nil {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"valid":   false,
			"message": "会话验证失败，请重新授权",
		}))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if isHTMLResponse(body) {
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"valid":   false,
			"message": "会话已失效，请重新授权",
		}))
		return
	}

	// 会话有效，恢复到缓存和数据库
	setPCSession(userID, req.SessionID)
	savePCSessionToDB(userID, req.SessionID, "")

	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"valid":      true,
		"message":    "会话恢复成功",
		"sessionId":  req.SessionID,
		"expireAt":   expireTime.Format(time.RFC3339),
		"ttlSeconds": int(time.Until(expireTime).Seconds()),
	}))
}

// PCSubmitArchiveEdit 提交档案编辑（联系方式/家庭成员/学校经历）
func PCSubmitArchiveEdit(c *gin.Context) {
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
	jsessionID := loginRes.SessionID
	if jsessionID == "" {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务暂不可用，请稍后再试", 503))
		return
	}

	var reqData map[string]interface{}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("请求数据格式错误", 400))
		return
	}

	client := getPCClientForUser(userID)
	u, _ := url.Parse(client.BaseURL)
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: jsessionID}})

	apiURL := client.BaseURL + "/student/base/edit"

	jsonData, _ := json.Marshal(reqData)
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建请求失败", 500))
		return
	}
	setPCAPIHeaders(httpReq)
	httpReq.Header.Set("Referer", client.BaseURL+"/student/base")
	httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, utils.NewStandardErrorResponse("学校服务器连接失败", 503))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if isHTMLResponse(body) {
		clearPCSession(userID)
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("PC会话已过期，请重新登录", 401))
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// ===== 私有辅助函数 =====

func getUserIDFromContext(c *gin.Context) (uint, error) {
	uidInterface, exists := c.Get("userId")
	if !exists {
		return 0, fmt.Errorf("unauthorized")
	}
	uidStr, ok := uidInterface.(string)
	if !ok {
		return 0, fmt.Errorf("invalid user id")
	}
	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user id")
	}
	return uint(uid), nil
}

func setPCRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

func setPCLoginSubmitHeaders(req *http.Request, jsessionID string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Origin", "https://xs.whggvc.net")
	req.Header.Set("Referer", "https://xs.whggvc.net/scloud/login")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Cookie", "JSESSIONID="+jsessionID)
}

func setPCAPIHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="149", "Chromium";v="149", "Not)A;Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "macOS")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "no-cache")
}

func extractJSESSIONID(cookies []*http.Cookie) string {
	for _, c := range cookies {
		if c.Name == "JSESSIONID" {
			return c.Value
		}
	}
	return ""
}

// getOrRefreshPCSession 获取有效PC会话，不存在时自动尝试登录
// OCR不可用时会返回 NeedManual=true，前端需展示验证码让用户手动输入
func getOrRefreshPCSession(userID uint) *PCLoginResult {
	result := &PCLoginResult{}

	// 1. 内存缓存命中
	if session := getPCSession(userID); session != nil {
		user, _ := models.FindUserByID(userID)
		result.SessionID = session.JSESSIONID
		result.User = user
		return result
	}

	// 2. 数据库会话有效
	user, _ := models.FindUserByID(userID)
	result.User = user
	if user != nil && user.PCJSESSIONID != "" && time.Now().Before(user.PCExpireTime) {
		setPCSession(userID, user.PCJSESSIONID)
		result.SessionID = user.PCJSESSIONID
		return result
	}

	// 3. 数据库无会话或已过期，自动尝试登录
	if user == nil {
		return result
	}

	client := getPCClientForUser(userID)
	loginResult := tryPCAutoLogin(client, user)
	if loginResult.NeedManual {
		return loginResult
	}
	if loginResult.SessionID != "" {
		result.SessionID = loginResult.SessionID
		return result
	}
	return result
}

// tryPCAutoLogin 尝试自动登录，OCR不可用时返回 NeedManual 供前端手动验证
func tryPCAutoLogin(client *PCClient, user *models.User) *PCLoginResult {
	result := &PCLoginResult{User: user}

	log.Printf("[DEBUG-AUTO-LOGIN] userID=%d jar cookies before init: %v", user.ID, client.Jar.Cookies(reqURL(client.BaseURL)))

	pcPassword, err := user.GetSchoolPassword()
	if err != nil {
		log.Printf("[PC-AUTO] userID=%d GetSchoolPassword failed: %v", user.ID, err)
		return result
	}

	// Step 1: GET /scloud/login 初始化会话（不再提前 return：若服务器未下发 Set-Cookie，
	// 下方 OCR 循环每次都会从 jar 兜底读 JSESSIONID，避免浪费 5 次 OCR 重试机会）
	initReq, _ := http.NewRequest("GET", client.BaseURL+"/login", nil)
	setPCRequestHeaders(initReq)
	initResp, err := client.HTTPClient.Do(initReq)
	if err != nil {
		log.Printf("[PC-AUTO] userID=%d init request failed: %v (will continue to OCR loop with jar fallback)", user.ID, err)
	} else {
		jsessionID := extractJSESSIONID(initResp.Cookies())
		log.Printf("[DEBUG-AUTO-LOGIN] userID=%d init: got JSESSIONID=%s resp.Cookies=%v", user.ID, jsessionID, initResp.Cookies())
		initResp.Body.Close()
		if jsessionID == "" {
			log.Printf("[PC-AUTO] userID=%d no JSESSIONID from init (OCR loop will use jar fallback)", user.ID)
		}
	}

	// 优先尝试 N 次全自动 OCR 登录：每次拉新验证码，重新 OCR 后提交。
	// 失败原因分两类：
	//   a) OCR 服务本身不可用 → 直接降级手动
	//   b) OCR 识别出错的验证码 → 重试（最多 maxRetries 次）
	maxRetries := 5
	ocrService := utils.GetOCRService()
	for attempt := 1; attempt <= maxRetries; attempt++ {
		initReq, _ := http.NewRequest("GET", client.BaseURL+"/login", nil)
		setPCRequestHeaders(initReq)
		initResp, err := client.HTTPClient.Do(initReq)
		if err != nil {
			log.Printf("[PC-AUTO] userID=%d attempt %d: init request failed: %v", user.ID, attempt, err)
			continue
		}
		io.ReadAll(initResp.Body)
		initResp.Body.Close()

		var curJsessionID string
		for _, c := range client.Jar.Cookies(reqURL(client.BaseURL)) {
			if c.Name == "JSESSIONID" {
				curJsessionID = c.Value
				break
			}
		}
		log.Printf("[DEBUG-AUTO-LOGIN] userID=%d attempt %d: jar cookies=%v extracted=%q", user.ID, attempt, client.Jar.Cookies(reqURL(client.BaseURL)), curJsessionID)
		if curJsessionID == "" {
			log.Printf("[PC-AUTO] userID=%d attempt %d: no JSESSIONID from init", user.ID, attempt)
			continue
		}

		captchaReq, _ := http.NewRequest("GET", client.BaseURL+"/validateCode", nil)
		setPCRequestHeaders(captchaReq)
		captchaReq.Header.Set("Cookie", "JSESSIONID="+curJsessionID)
		captchaResp, err := client.HTTPClient.Do(captchaReq)
		if err != nil {
			log.Printf("[PC-AUTO] userID=%d attempt %d: captcha request failed: %v", user.ID, attempt, err)
			continue
		}
		captchaData, err := io.ReadAll(captchaResp.Body)
		captchaResp.Body.Close()
		if err != nil || len(captchaData) < 100 {
			log.Printf("[PC-AUTO] userID=%d attempt %d: captcha read failed", user.ID, attempt)
			continue
		}

		// OCR 识别
		ocrResult, ocrErr := ocrService.RecognizeFromBytes(captchaData)
		if ocrErr != nil {
			// OCR 服务自身不可用（如网络/配置问题）：立即降级手动
			log.Printf("[PC-AUTO] userID=%d attempt %d: OCR service unavailable: %v, falling back to manual", user.ID, attempt, ocrErr)
			result.NeedManual = true
			result.CaptchaData = captchaData
			result.CaptchaSID = curJsessionID
			return result
		}

		// POST 提交登录
		formData := url.Values{}
		formData.Set("username", user.Username)
		formData.Set("password", pcPassword)
		formData.Set("randomcode", ocrResult)

		loginReq, _ := http.NewRequest("POST", client.BaseURL+"/login", strings.NewReader(formData.Encode()))
		setPCLoginSubmitHeaders(loginReq, curJsessionID)

		loginResp, err := client.HTTPClient.Do(loginReq)
		if err != nil {
			log.Printf("[PC-AUTO] userID=%d attempt %d: login request failed: %v", user.ID, attempt, err)
			continue
		}
		defer loginResp.Body.Close()

		location := loginResp.Header.Get("Location")
		log.Printf("[DEBUG-AUTO-LOGIN] userID=%d attempt %d: status=%d Location=%q JSESSIONID=%s", user.ID, attempt, loginResp.StatusCode, location, curJsessionID[:8])

		if loginResp.StatusCode == 302 && strings.Contains(location, "/index") {
			newJsessionID := extractJSESSIONID(loginResp.Cookies())
			if newJsessionID == "" {
				newJsessionID = curJsessionID
			}
			setPCSession(user.ID, newJsessionID)
			savePCSessionToDB(user.ID, newJsessionID, loginReq.Header.Get("User-Agent"))
			log.Printf("[PC-AUTO] userID=%d login success (OCR=%s, attempt=%d)", user.ID, ocrResult, attempt)
			result.SessionID = newJsessionID
			return result
		}

		// OCR 识别错误或密码错误都会回到登录页 → 继续重试换新验证码
		body, _ := io.ReadAll(io.LimitReader(loginResp.Body, 300))
		log.Printf("[DEBUG-AUTO-LOGIN] userID=%d attempt %d: login failed body=%q (will retry with new captcha)", user.ID, attempt, body)

		if attempt < maxRetries {
			time.Sleep(300 * time.Millisecond)
		}
	}

	// 所有自动尝试都失败（包括 OCR 一直识别错），降级为手动验证
	// 重新获取一次会话和验证码给用户输入
	initReq2, _ := http.NewRequest("GET", client.BaseURL+"/login", nil)
	setPCRequestHeaders(initReq2)
	initResp2, _ := client.HTTPClient.Do(initReq2)
	newSID := extractJSESSIONID(initResp2.Cookies())
	if initResp2 != nil {
		io.ReadAll(initResp2.Body)
		initResp2.Body.Close()
	}
	// jar fallback：服务器若未下发 Set-Cookie（jar 中已存在的会话仍合法），用 jar 里的 JSESSIONID，
	// 否则验证码请求会与表单提交用错会话，导致用户在弹窗里输入再多次也无效
	if newSID == "" {
		for _, c := range client.Jar.Cookies(reqURL(client.BaseURL)) {
			if c.Name == "JSESSIONID" {
				newSID = c.Value
				break
			}
		}
		if newSID != "" {
			log.Printf("[PC-AUTO] userID=%d manual fallback: no Set-Cookie from init, using jar JSESSIONID", user.ID)
		}
	}

	captchaReq2, _ := http.NewRequest("GET", client.BaseURL+"/validateCode", nil)
	setPCRequestHeaders(captchaReq2)
	if newSID != "" {
		captchaReq2.Header.Set("Cookie", "JSESSIONID="+newSID)
	}
	captchaResp2, _ := client.HTTPClient.Do(captchaReq2)
	captchaData2, _ := io.ReadAll(captchaResp2.Body)
	captchaResp2.Body.Close()
	if newSID != "" {
		setPCSession(user.ID, newSID)
	}

	log.Printf("[PC-AUTO] userID=%d all %d attempts failed, falling back to manual", user.ID, maxRetries)
	result.NeedManual = true
	result.CaptchaData = captchaData2
	result.CaptchaSID = newSID
	return result
}

func savePCSessionToDB(userID uint, jsessionID, userAgent string) {
	updates := map[string]interface{}{
		"pc_jsession_id": jsessionID,
		"pc_login_time":  time.Now(),
		"pc_expire_time": time.Now().Add(PCSessionTTL),
	}
	if userAgent != "" {
		updates["pc_user_agent"] = userAgent
	}
	models.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
}

func clearPCSessionFromDB(userID uint) {
	models.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"pc_jsession_id": "",
		"pc_expire_time": nil,
	})
}

func isHTMLResponse(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return false
	}
	return trimmed[0] == '<'
}

func reqURL(rawURL string) *url.URL {
	u, _ := url.Parse(rawURL)
	return u
}

// 全局PC Client映射（每个用户一个独立Client，带读写锁）
var (
	pcClientMap = make(map[uint]*PCClient)
	clientLock  sync.RWMutex
)

func getPCClientForUser(userID uint) *PCClient {
	// 快速路径：读锁检查缓存
	clientLock.RLock()
	if client, ok := pcClientMap[userID]; ok {
		clientLock.RUnlock()
		return client
	}
	clientLock.RUnlock()

	// 慢路径：写锁创建新Client（双重检查）
	clientLock.Lock()
	defer clientLock.Unlock()
	if client, ok := pcClientMap[userID]; ok {
		return client
	}

	var client *PCClient
	jar, _ := cookiejar.New(nil)
	existingJsessionID := ""
	user, err := models.FindUserByID(userID)
	if err == nil && user.PCJSESSIONID != "" {
		existingJsessionID = user.PCJSESSIONID
	}
	client = &PCClient{
		BaseURL: pcBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Jar: jar,
	}
	if existingJsessionID != "" {
		log.Printf("[DEBUG-AUTO-LOGIN] userID=%d pre-loading JSESSIONID from DB: %s", userID, existingJsessionID)
		u, _ := url.Parse(client.BaseURL)
		jar.SetCookies(u, []*http.Cookie{{Name: "JSESSIONID", Value: existingJsessionID}})
	}
	pcClientMap[userID] = client
	return client
}
