package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
)

// 分级超时配置（按 "分接口分级" 策略）
// 默认 TCP 3s / TLS 3s / 等响应头 5s / 总 8s。
// 生产到学校服务器跨省跨 ISP 时，实测单次握手就可能挂 30s+，
// 把各阶段压到秒级，让上游不通时也能秒失败、走缓存降级，避免连接池被占满。
//
// 注意：env 在第一次 NewProxyClient() 时才读取（懒初始化）。
// main.go 用 godotenv.Load() 在 main() 里加载 .env，如果放在 init() 里
// 读 env 会过早（godotenv.Load 还没执行），导致永远走默认值。
var (
	globalHTTPTransport     *http.Transport
	globalHTTPTransportOnce sync.Once
)

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// buildHTTPTransport 按当前进程 env 构造 HTTP Transport。
func buildHTTPTransport() *http.Transport {
	dialTimeout := envInt("HTTP_DIAL_TIMEOUT_SEC", 3)
	tlsTimeout := envInt("HTTP_TLS_TIMEOUT_SEC", 3)
	respHeaderTimeout := envInt("HTTP_RESP_HEADER_TIMEOUT_SEC", 5)
	idleTimeout := envInt("HTTP_IDLE_TIMEOUT_SEC", 90)
	maxIdle := envInt("HTTP_MAX_IDLE_CONNS", 100)
	maxIdlePerHost := envInt("HTTP_MAX_IDLE_CONNS_PER_HOST", 20)

	return &http.Transport{
		MaxIdleConns:        maxIdle,
		IdleConnTimeout:     time.Duration(idleTimeout) * time.Second,
		MaxIdleConnsPerHost: maxIdlePerHost,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(dialTimeout) * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   time.Duration(tlsTimeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(respHeaderTimeout) * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// ensureHTTPTransport 懒初始化全局 Transport，进程内只构造一次。
// 第一次调用发生在 NewProxyClient() 里——此时 main() 已执行 godotenv.Load，env 就绪。
func ensureHTTPTransport() *http.Transport {
	globalHTTPTransportOnce.Do(func() {
		globalHTTPTransport = buildHTTPTransport()
	})
	return globalHTTPTransport
}

// ProxyTimeoutSeconds 返回当前 ProxyClient 的整体请求超时（秒）
func ProxyTimeoutSeconds() int {
	return envInt("HTTP_TOTAL_TIMEOUT_SEC", 8)
}

// ProxyClient 代理客户端
type ProxyClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewProxyClient 创建新的代理客户端，每个客户端拥有独立的 CookieJar 以实现用户隔离
func NewProxyClient() *ProxyClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		jar = nil
	}
	client := &http.Client{
		Timeout:   time.Duration(ProxyTimeoutSeconds()) * time.Second,
		Jar:       jar,
		Transport: ensureHTTPTransport(),
	}
	return &ProxyClient{
		BaseURL:    config.GetSchoolServerURL(),
		HTTPClient: client,
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Result  struct {
		UserInfo struct {
			ID              string `json:"id"`
			Username        string `json:"username"`
			Realname        string `json:"realname"`
			Avatar          string `json:"avatar"`
			Birthday        string `json:"birthday"`
			Sex             int    `json:"sex"`
			Email           string `json:"email"`
			Phone           string `json:"phone"`
			ClassName       string `json:"className"`
			SchoolID        int    `json:"schoolId"`
			ProfessionID    int    `json:"professionId"`
			FacultyID       int    `json:"facultyId"`
			GradeID         int    `json:"gradeId"`
			CurrentSemester string `json:"currentSemester"`
			IdentityCard    string `json:"identityCard"`
		} `json:"userInfo"`
		Token string `json:"token"`
	} `json:"result"`
	Timestamp int64 `json:"timestamp"`
}

// ProxyLogin 代理登录到学校服务器
func (p *ProxyClient) ProxyLogin(username, password string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("序列化登录请求失败: %w", err)
	}

	// 创建请求
	loginURL := p.BaseURL + "/scloudoa/sys/mLogin"
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	p.setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &loginResp, nil
}

// ProxyRequest 代理请求到学校服务器
func (p *ProxyClient) ProxyRequest(method, path string, params url.Values, token string) ([]byte, error) {
	// 构建完整URL
	fullURL := p.BaseURL + path
	if method == "GET" && params != nil {
		fullURL += "?" + params.Encode()
	}

	var req *http.Request
	var err error

	if method == "POST" && params != nil {
		req, err = http.NewRequest(method, fullURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
	}

	// 设置请求头
	p.setCommonHeaders(req)
	if token != "" {
		req.Header.Set("X-Access-Token", token)
	}

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, nil
}

// ProxyRequestWithClientHeaders 支持携带用户信息（自动登录）、token、clientId 的通用代理
func (p *ProxyClient) ProxyRequestWithClientHeaders(method, path string, params url.Values, user *models.User, clientID, overrideToken, overrideCookie string) ([]byte, error) {
	// 确保用户已登录（如果没有外部 token）
	if overrideToken == "" {
		if err := p.AutoLogin(user, false); err != nil {
			return nil, err
		}
	} else {
		// 如果有外部 token，更新用户 token 以便后续请求复用
		user.SchoolToken = overrideToken
	}

	fullURL := p.BaseURL + path
	if params == nil {
		params = url.Values{}
	}
	if method == "GET" && len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	var req *http.Request
	var err error
	if method == "POST" && params != nil {
		req, err = http.NewRequest(method, fullURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest(method, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
	}

	// 通用头
	p.setCommonHeaders(req)
	// 学校 token
	if user.SchoolToken != "" {
		req.Header.Set("X-Access-Token", user.SchoolToken)
		req.Header.Set("Authorization", "Bearer "+user.SchoolToken)
	}
	// 前端 clientId 透传
	if clientID != "" {
		req.Header.Set("X-Client-Id", clientID)
	}
	// 附加 cookie 以兼容部分接口校验
	if user.SchoolToken != "" {
		cookies := []*http.Cookie{
			{Name: "token", Value: user.SchoolToken, Path: "/"},
		}
		if clientID != "" {
			cookies = append(cookies, &http.Cookie{Name: "clientId", Value: clientID, Path: "/"})
		}
		p.HTTPClient.Jar.SetCookies(req.URL, cookies)
	}
	// 透传前端带来的 Cookie（如 JSESSIONID 等）
	if overrideCookie != "" {
		req.Header.Set("Cookie", overrideCookie)
	}

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	return body, nil
}

// setCommonHeaders 设置通用请求头
func (p *ProxyClient) setCommonHeaders(req *http.Request) {
	timestamp := time.Now().Format("20060102150405")
	sign := p.generateSign(timestamp)

	req.Header.Set("User-Agent", config.DefaultUserAgent)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-Sign", sign)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
}

// SetCommonHeaders 导出通用请求头设置
func (p *ProxyClient) SetCommonHeaders(req *http.Request) {
	p.setCommonHeaders(req)
}

// generateSign 生成签名
func (p *ProxyClient) generateSign(timestamp string) string {
	// 这里使用简单的MD5签名，实际项目中可能需要更复杂的签名算法
	data := timestamp + "pingjiao_sign_key"
	hash := md5.Sum([]byte(data))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

// debugAutoLoginEnabled 控制 AutoLogin 是否打印诊断日志。
// 生产默认 false：token 有效复用时的"成功"日志没价值，但日志量很大；
// 排查问题时设环境变量 DEBUG_AUTO_LOGIN=true 打开。
func debugAutoLoginEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("DEBUG_AUTO_LOGIN")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// AutoLogin 自动登录并更新用户token。force=true 时跳过本地缓存检查，强制重新登录学校系统
func (p *ProxyClient) AutoLogin(user *models.User, force bool) error {
	encKey := config.GetSchoolPasswordEncKey()
	debug := debugAutoLoginEnabled()

	if force {
		if debug {
			log.Printf("[DEBUG-AUTO-LOGIN] userID=%d force=true, skipping token check, will re-login school system", user.ID)
		}
	} else {
		tokenValid := time.Now().Before(user.TokenExpireAt) && user.SchoolToken != ""
		if tokenValid {
			// 复用成功路径：默认静默，需要诊断再开 DEBUG_AUTO_LOGIN
			if debug {
				log.Printf("[DEBUG-AUTO-LOGIN] userID=%d tokenValid=true tokenExpireAt=%v schoolTokenLen=%d schoolPasswordEncLen=%d encKeyLen=%d username=%s userType=%s",
					user.ID, user.TokenExpireAt.Format("2006-01-02T15:04:05Z07:00"), len(user.SchoolToken), len(user.SchoolPasswordEnc), len(encKey), user.Username, user.UserType)
			}
			return nil
		}
		// token 即将过期/缺失，要走重新登录的路径——这里打一条 warning，保留可观测性
		if debug {
			log.Printf("[DEBUG-AUTO-LOGIN] userID=%d tokenValid=false tokenExpireAt=%v schoolTokenLen=%d (will re-login)",
				user.ID, user.TokenExpireAt.Format("2006-01-02T15:04:05Z07:00"), len(user.SchoolToken))
		}
	}

	// 检查学校密码是否已设置（用于自动登录）
	if user.SchoolPasswordEnc == "" {
		if debugAutoLoginEnabled() {
			log.Printf("[DEBUG-AUTO-LOGIN-FAIL] userID=%d username=%s userType=%s wechatOpenID=%s wechatUnionID=%s - no schoolPasswordEnc, cannot auto-login",
				user.ID, user.Username, user.UserType, user.WechatOpenID, user.WechatUnionID)
		}
		return ErrSchoolPasswordNotSet
	}

	return withAutoLoginLock(user.ID, func() error {
		if !force {
			latest, err := models.FindUserByID(user.ID)
			if err == nil {
				if time.Now().Before(latest.TokenExpireAt) && latest.SchoolToken != "" {
					*user = *latest
					return nil
				}
			}
		}

		password, err := user.GetSchoolPassword()
		if err != nil {
			return fmt.Errorf("自动登录失败: %w", err)
		}

		loginResp, err := p.ProxyLogin(user.Username, password)
		if debugAutoLoginEnabled() {
			log.Printf("[DEBUG-AUTO-LOGIN] ProxyLogin result: success=%v message=%s tokenLen=%d",
				loginResp.Success, loginResp.Message, len(loginResp.Result.Token))
		}
		if err != nil {
			if isSchoolUnreachableErr(err) {
				recordSchoolUnreachable(user.ID)
			}
			return fmt.Errorf("自动登录失败: %w", err)
		}
		if !loginResp.Success {
			return fmt.Errorf("登录失败: %s", loginResp.Message)
		}
		if loginResp.Result.Token == "" {
			return fmt.Errorf("登录失败: 学校服务器未返回有效token")
		}

		updates := map[string]interface{}{
			"school_token":     loginResp.Result.Token,
			"token_expire_at":  time.Now().Add(24 * time.Hour),
			"realname":         loginResp.Result.UserInfo.Realname,
			"avatar":           loginResp.Result.UserInfo.Avatar,
			"birthday":         loginResp.Result.UserInfo.Birthday,
			"sex":              loginResp.Result.UserInfo.Sex,
			"email":            loginResp.Result.UserInfo.Email,
			"phone":            loginResp.Result.UserInfo.Phone,
			"class_name":       loginResp.Result.UserInfo.ClassName,
			"profession_id":    loginResp.Result.UserInfo.ProfessionID,
			"faculty_id":       loginResp.Result.UserInfo.FacultyID,
			"grade_id":         loginResp.Result.UserInfo.GradeID,
			"current_semester": loginResp.Result.UserInfo.CurrentSemester,
			"identity_card":    loginResp.Result.UserInfo.IdentityCard,
		}
		if err := models.UpdateUserFields(user.ID, updates); err != nil {
			return fmt.Errorf("更新用户信息失败: %w", err)
		}
		return nil
	})
}

// ProxyRequestWithAutoLogin 带自动登录的代理请求
func (p *ProxyClient) ProxyRequestWithAutoLogin(user *models.User, method, path string, params url.Values) ([]byte, error) {
	debug := debugAutoLoginEnabled()
	// 确保用户已登录
	if err := p.AutoLogin(user, false); err != nil {
		if debug {
			log.Printf("[DEBUG-PROXY] userID=%d path=%s AutoLogin error: %v", user.ID, path, err)
		}
		return nil, err
	}

	// 发送代理请求
	body, err := p.ProxyRequest(method, path, params, user.SchoolToken)
	if err != nil {
		return nil, err
	}

	// 检查响应是否表示token过期
	if p.isTokenExpired(body) {
		if debug {
			log.Printf("[DEBUG-PROXY-RETRY] userID=%d path=%s - isTokenExpired=true, will force re-login school system", user.ID, path)
		}
		if err := p.AutoLogin(user, true); err != nil {
			if debug {
				log.Printf("[DEBUG-PROXY-RETRY] userID=%d path=%s - AutoLogin(force) failed: %v", user.ID, path, err)
			}
			return nil, err
		}
		// 重新发送请求
		body, err = p.ProxyRequest(method, path, params, user.SchoolToken)
		if debug {
			log.Printf("[DEBUG-PROXY-RETRY] userID=%d path=%s - retried with new token, err=%v", user.ID, path, err)
		}
		return body, err
	}

	return body, nil
}

// isTokenExpired 检查响应是否表示token过期
func (p *ProxyClient) isTokenExpired(body []byte) bool {
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return false
	}

	// 检查常见的token过期标识
	if code, ok := resp["code"].(float64); ok {
		if code == 401 || code == 403 {
			return true
		}
	}

	if message, ok := resp["message"].(string); ok {
		if strings.Contains(strings.ToLower(message), "token") ||
			strings.Contains(strings.ToLower(message), "unauthorized") ||
			strings.Contains(strings.ToLower(message), "登录") {
			return true
		}
	}

	return false
}

// DownloadFileWithAuth 带认证的文件下载
func (p *ProxyClient) DownloadFileWithAuth(user *models.User, attachmentUrl, fileName string) ([]byte, string, error) {
	fmt.Printf("下载附件请求: URL=%s, 文件名=%s\n", attachmentUrl, fileName)

	// 确保用户已登录
	if err := p.AutoLogin(user, false); err != nil {
		return nil, "", fmt.Errorf("自动登录失败: %w", err)
	}

	// 方法1: 通过学校的官方下载API
	downloadAPIURL := fmt.Sprintf("https://xs.whggvc.net/scloudoa/sys/common/download/%s", attachmentUrl)
	fileData, contentType, err := p.downloadFromURL(user, downloadAPIURL, fileName)
	if err == nil && len(fileData) > 100 && !strings.Contains(contentType, "application/json") {
		fmt.Printf("通过下载API成功，文件大小: %d bytes, Content-Type: %s\n", len(fileData), contentType)
		return fileData, contentType, nil
	}
	fmt.Printf("通过下载API失败: %v\n", err)

	// 方法2: 通过getFileList API获取文件URL并下载
	fileData, contentType, err = p.downloadViaGetFileList(user, attachmentUrl, fileName)
	if err == nil && len(fileData) > 100 && !strings.Contains(contentType, "application/json") {
		fmt.Printf("通过getFileList成功，文件大小: %d bytes, Content-Type: %s\n", len(fileData), contentType)
		return fileData, contentType, nil
	}
	fmt.Printf("getFileList方式失败: %v\n", err)

	// 方法3: 回退到直接下载方式
	return p.downloadDirectly(user, attachmentUrl, fileName)
}

// downloadViaGetFileList 通过getFileList API下载文件
func (p *ProxyClient) downloadViaGetFileList(user *models.User, attachmentUrl, fileName string) ([]byte, string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"fileList": []map[string]interface{}{
			{
				"name":          fileName,
				"url":           fmt.Sprintf("http://scs.whggvc.net/scscloud/%s", attachmentUrl),
				"size":          "98.912KB",
				"attachmentUrl": attachmentUrl,
				"isDownload":    1,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建请求
	getFileListURL := "https://xs.whggvc.net/scloudoa/sys/common/getFileList"
	req, err := http.NewRequest("POST", getFileListURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置头部
	p.setCommonHeaders(req)
	req.Header.Set("X-Access-Token", user.SchoolToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// 读取响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("getFileList API响应: %s\n", string(respData))

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应是否成功
	if success, ok := result["success"].(bool); !ok || !success {
		return nil, "", fmt.Errorf("getFileList API返回失败")
	}

	// 获取文件URL
	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("响应格式错误")
	}

	fileList, ok := resultData["fileList"].([]interface{})
	if !ok || len(fileList) == 0 {
		return nil, "", fmt.Errorf("文件列表为空")
	}

	fileInfo, ok := fileList[0].(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("文件信息格式错误")
	}

	fileURL, ok := fileInfo["url"].(string)
	if !ok || fileURL == "" {
		return nil, "", fmt.Errorf("文件URL为空")
	}

	fmt.Printf("通过getFileList获取到文件URL: %s\n", fileURL)

	// 下载文件
	return p.downloadFromURL(user, fileURL, fileName)
}

// downloadDirectly 直接下载文件（原有方式）
func (p *ProxyClient) downloadDirectly(user *models.User, attachmentUrl, fileName string) ([]byte, string, error) {
	// 尝试多个可能的基础URL（包括scscloud路径，这是getFileList返回的URL所使用的）
	baseUrls := []string{
		"https://scs.whggvc.net/scsoa/sys/common/static/",
		"https://scs.whggvc.net/scloudoa/sys/common/static/",
		"https://scs.whggvc.net/static/",
		"http://scs.whggvc.net/scscloud/",
		"https://scs.whggvc.net/scscloud/",
	}

	// #region agent log
	log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H1 location=proxy.go:downloadDirectly message=Testing downloadDirectly with baseUrls count=%d attachmentUrl=%s", len(baseUrls), attachmentUrl)
	// #endregion

	var lastErr error
	for _, baseURL := range baseUrls {
		fullURL := baseURL + attachmentUrl
		fmt.Printf("尝试下载链接: %s\n", fullURL)

		// #region agent log
		log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H1 location=proxy.go:downloadDirectly:try message=Attempting URL baseURL=%s fullURL=%s", baseURL, fullURL)
		// #endregion

		fileData, contentType, err := p.downloadFromURL(user, fullURL, fileName)
		if err == nil {
			// #region agent log
			log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H1 location=proxy.go:downloadDirectly:success message=Success baseURL=%s fileSize=%d contentType=%s", baseURL, len(fileData), contentType)
			// #endregion
			return fileData, contentType, nil
		}

		fmt.Printf("下载失败: %v\n", err)
		lastErr = err
	}

	return nil, "", fmt.Errorf("所有下载方式都失败了，最后错误: %w", lastErr)
}

// downloadFromURL 从指定URL下载文件
func (p *ProxyClient) downloadFromURL(user *models.User, fileURL, fileName string) ([]byte, string, error) {
	// #region agent log
	log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H2H3 location=proxy.go:downloadFromURL:enter message=Entering downloadFromURL fileURL=%s", fileURL)
	// #endregion

	// 创建请求
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置通用头部
	p.setCommonHeaders(req)
	req.Header.Set("X-Access-Token", user.SchoolToken)
	req.Header.Set("Accept", "application/octet-stream, */*")

	// 添加Referer头，模拟从学校网站访问
	req.Header.Set("Referer", "https://xs.whggvc.net/scloudapp/")

	// 如果是文件服务器，设置正确的Host头
	if strings.Contains(fileURL, "scs.whggvc.net") {
		req.Header.Set("Host", "scs.whggvc.net")
		// 尝试添加Cookie来模拟浏览器会话
		req.Header.Set("Cookie", fmt.Sprintf("X-Access-Token=%s", user.SchoolToken))
	}

	// #region agent log
	log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H-AUTH location=proxy.go:downloadFromURL:request message=Request prepared tokenLen=%d url=%s host=%s cookie=%s referer=%s", len(user.SchoolToken), fileURL, req.Host, req.Header.Get("Cookie"), req.Header.Get("Referer"))
	// #endregion

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// #region agent log
	setCookies := resp.Header.Values("Set-Cookie")
	log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H-AUTH location=proxy.go:downloadFromURL:response message=Got response statusCode=%d dataSize=%d contentType=%s setCookies=%v url=%s", resp.StatusCode, 0, resp.Header.Get("Content-Type"), setCookies, fileURL)
	// #endregion

	// 读取响应体
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应失败: %w", err)
	}

	// #region agent log
	log.Printf("[DEBUG-DOWNLOAD] sessionId=28a407 runId=pre-fix hypothesisId=H2H3 location=proxy.go:downloadFromURL:response message=Got response statusCode=%d dataSize=%d contentType=%s url=%s", resp.StatusCode, len(fileData), resp.Header.Get("Content-Type"), fileURL)
	// #endregion

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// 根据文件扩展名推断Content-Type
		switch {
		case strings.HasSuffix(strings.ToLower(fileName), ".pdf"):
			contentType = "application/pdf"
		case strings.HasSuffix(strings.ToLower(fileName), ".doc"):
			contentType = "application/msword"
		case strings.HasSuffix(strings.ToLower(fileName), ".docx"):
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case strings.HasSuffix(strings.ToLower(fileName), ".xls"):
			contentType = "application/vnd.ms-excel"
		case strings.HasSuffix(strings.ToLower(fileName), ".xlsx"):
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		default:
			contentType = "application/octet-stream"
		}
	}

	fmt.Printf("DownloadFromURL: 文件大小=%d, Content-Type=%s\n", len(fileData), contentType)

	// 检查是否返回了JSON错误响应
	if strings.Contains(contentType, "application/json") {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(fileData, &errorResp); err == nil {
			if status, ok := errorResp["status"].(float64); ok && status == 4 {
				if message, ok := errorResp["message"].(string); ok {
					fmt.Printf("DownloadFromURL: 服务器返回认证错误: %s\n", message)
					return nil, "", fmt.Errorf("文件服务器认证失败: %s", message)
				}
			}
		}
	}

	// 如果返回的是HTML，说明可能是错误页面或登录页面
	if strings.Contains(contentType, "text/html") {
		preview := string(fileData)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("DownloadFromURL: 检测到HTML内容，可能是错误页面: %s\n", preview)

		// 检查是否包含登录相关的关键词
		htmlContent := strings.ToLower(string(fileData))
		if strings.Contains(htmlContent, "login") || strings.Contains(htmlContent, "登录") ||
			strings.Contains(htmlContent, "unauthorized") || strings.Contains(htmlContent, "403") ||
			strings.Contains(htmlContent, "404") || strings.Contains(htmlContent, "error") {
			return nil, "", fmt.Errorf("服务器返回错误页面，可能需要重新登录或文件不存在")
		}

		// 如果是其他HTML内容，也返回错误
		return nil, "", fmt.Errorf("服务器返回HTML页面而不是文件内容，请检查文件链接")
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	return fileData, contentType, nil
}

// GetFileURLViaGetFileList 通过getFileList API获取文件URL（不下载文件）
func (p *ProxyClient) GetFileURLViaGetFileList(user *models.User, attachmentUrl, fileName string) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"fileList": []map[string]interface{}{
			{
				"name":          fileName,
				"url":           fmt.Sprintf("http://scs.whggvc.net/scscloud/%s", attachmentUrl),
				"size":          "98.912KB",
				"attachmentUrl": attachmentUrl,
				"isDownload":    1,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构建请求体失败: %w", err)
	}

	// 创建请求
	getFileListURL := "https://xs.whggvc.net/scloudoa/sys/common/getFileList"
	req, err := http.NewRequest("POST", getFileListURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置头部
	p.setCommonHeaders(req)
	req.Header.Set("X-Access-Token", user.SchoolToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	// 发送请求
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	// 读取响应
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respData, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查响应是否成功
	if success, ok := result["success"].(bool); !ok || !success {
		return "", fmt.Errorf("getFileList API返回失败")
	}

	// 获取文件URL
	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("响应格式错误")
	}

	fileList, ok := resultData["fileList"].([]interface{})
	if !ok || len(fileList) == 0 {
		return "", fmt.Errorf("文件列表为空")
	}

	fileInfo, ok := fileList[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("文件信息格式错误")
	}

	fileURL, ok := fileInfo["url"].(string)
	if !ok || fileURL == "" {
		return "", fmt.Errorf("文件URL为空")
	}

	return fileURL, nil
}
