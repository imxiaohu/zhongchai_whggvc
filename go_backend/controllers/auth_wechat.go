// Package controllers handles HTTP request handlers for authentication endpoints.
// This file contains WeChat login and school binding functions.
package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// WechatLogin 处理微信登录请求
func WechatLogin(c *gin.Context) {
	var req WechatLoginRequest

	// 解析请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("缺少参数code", 400))
		return
	}

	// 获取微信配置
	appID := config.GetWechatAppID()
	appSecret := config.GetWechatAppSecret()

	if appID == "" || appSecret == "" {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("微信配置未设置", 500))
		return
	}

	// 调用微信API获取session信息
	wxLoginURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, req.Code)

	resp, err := http.Get(wxLoginURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("微信登录失败", 500))
		return
	}
	//nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("读取微信响应失败", 500))
		return
	}

	var wxResp WechatSessionResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("解析微信响应失败", 500))
		return
	}

	// 检查微信API错误
	if wxResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse(fmt.Sprintf("微信登录失败: %s", wxResp.ErrMsg), 400))
		return
	}

	// 查找或创建用户
	user, err := models.FindUserByWechatOpenID(wxResp.OpenID)
	if err != nil {
		// 创建新用户，使用临时用户名标识，等待用户绑定学校账号时设置真实用户名
		tempUsername := fmt.Sprintf("wx_%s", wxResp.OpenID) // 使用wx_前缀+OpenID作为临时用户名
		user = &models.User{
			Username:      tempUsername, // 使用临时用户名，避免唯一约束冲突
			WechatOpenID:  wxResp.OpenID,
			WechatUnionID: wxResp.UnionID,
			Nickname:      "微信用户",
			UserType:      "student",
			TokenExpireAt: time.Now().Add(24 * time.Hour), // 设置token过期时间
			LastLoginAt:   time.Now(),
			Status:        1,
		}

		// 获取默认学校
		schools, _ := models.GetAllSchools()
		if len(schools) > 0 {
			user.SchoolID = schools[0].ID
		}

		// 保存用户
		if err := models.CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("创建用户失败", 500))
			return
		}
	} else {
		// 更新最后登录时间
		//nolint:errcheck
		models.UpdateUserFields(user.ID, map[string]interface{}{
			"last_login_at": time.Now(),
		})
	}

	// 更新活跃状态并恢复个人信息缓存
	onUserLogin(user.ID)

	// 生成JWT令牌
	token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.WechatOpenID, "wechat")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("生成令牌失败", 500))
		return
	}

	// 检查是否绑定学校账号
	// 绑定条件：用户名不以"wx_"开头（不是临时用户名）且有密码和真实姓名
	hasSchoolAccount := user.Username != "" &&
		!strings.HasPrefix(user.Username, "wx_") &&
		user.Password != "" &&
		user.Realname != ""

	// 构建用户信息响应，使用与账号密码登录一致的字段名
	userInfo := gin.H{
		"id":               fmt.Sprintf("%d", user.ID),
		"username":         user.Username,
		"realname":         user.Realname,
		"avatar":           user.Avatar,
		"email":            user.Email,
		"phone":            user.Phone,
		"birthday":         user.Birthday,
		"sex":              user.Sex,
		"identityCard":     user.IdentityCard,
		"schoolId":         user.SchoolID,
		"className":        user.ClassName,
		"professionId":     user.ProfessionID,
		"facultyId":        user.FacultyID,
		"gradeId":          user.GradeID,
		"currentSemester":  user.CurrentSemester,
		"hasSchoolAccount": hasSchoolAccount, // 添加绑定状态字段
		"wechatOpenId":     user.WechatOpenID,
		"nickname":         user.Nickname,
	}

	if hasSchoolAccount {
		// 已绑定学校账号，使用真实姓名作为显示名称
		userInfo["realname"] = user.Realname
	} else {
		// 未绑定学校账号，使用微信昵称
		userInfo["realname"] = user.Nickname
	}

	// 返回登录成功响应，使用标准化响应格式与账号密码登录保持一致
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"token":    token,
		"userInfo": userInfo,
	}))
}

// SchoolBind 处理学校账号绑定请求
func SchoolBind(c *gin.Context) {
	var req SchoolBindRequest

	// 解析请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("学号、密码和验证码不能为空", 400))
		return
	}

	// 从JWT中获取用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户ID格式错误", 400))
		return
	}

	// 查询用户信息
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		// 保留业务 code 字段以兼容前端 TOKEN_INVALID 判断
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "登录状态已失效，请重新登录",
			"code":    "TOKEN_INVALID",
		})
		return
	}

	// 创建代理客户端
	proxyClient := utils.NewProxyClient()

	// 尝试通过学校服务器验证登录
	loginResp, err := proxyClient.ProxyLogin(req.StudentID, req.Password)
	if err != nil || !loginResp.Success {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("学号或密码错误", 401))
		return
	}

	// 检查是否已存在相同学号的用户
	existingUser, err := models.FindUserByUsername(req.StudentID)
	if err == nil {
		// 如果已存在学号用户，将微信信息绑定到现有用户
		existingUser.WechatOpenID = user.WechatOpenID
		existingUser.WechatUnionID = user.WechatUnionID
		log.Printf("[DEBUG-WECHAT-BIND] userID=%d existingUserID=%d username=%s - existing user, setting schoolPasswordEnc (len=%d)",
			user.ID, existingUser.ID, req.StudentID, len(req.Password))
		if err := existingUser.SetSchoolPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("保存学校密码失败", 500))
			return
		}
		log.Printf("[DEBUG-WECHAT-BIND] userID=%d existingUserID=%d - schoolPasswordEnc set successfully (encLen=%d)",
			user.ID, existingUser.ID, len(existingUser.SchoolPasswordEnc))
		existingUser.Realname = loginResp.Result.UserInfo.Realname
		existingUser.Avatar = loginResp.Result.UserInfo.Avatar
		existingUser.Birthday = loginResp.Result.UserInfo.Birthday
		existingUser.Sex = loginResp.Result.UserInfo.Sex
		existingUser.Email = loginResp.Result.UserInfo.Email
		existingUser.Phone = loginResp.Result.UserInfo.Phone
		existingUser.IdentityCard = loginResp.Result.UserInfo.IdentityCard
		existingUser.ClassName = loginResp.Result.UserInfo.ClassName
		existingUser.ProfessionID = uint(loginResp.Result.UserInfo.ProfessionID)
		existingUser.FacultyID = uint(loginResp.Result.UserInfo.FacultyID)
		existingUser.GradeID = uint(loginResp.Result.UserInfo.GradeID)
		existingUser.CurrentSemester = loginResp.Result.UserInfo.CurrentSemester
		existingUser.SchoolToken = loginResp.Result.Token
		existingUser.TokenExpireAt = time.Now().Add(24 * time.Hour)
		existingUser.LastLoginAt = time.Now()

		// 更新现有用户
		if err := models.UpdateUserFields(existingUser.ID, map[string]interface{}{
			"realname":         loginResp.Result.UserInfo.Realname,
			"avatar":           loginResp.Result.UserInfo.Avatar,
			"birthday":         loginResp.Result.UserInfo.Birthday,
			"sex":              loginResp.Result.UserInfo.Sex,
			"email":            loginResp.Result.UserInfo.Email,
			"phone":            loginResp.Result.UserInfo.Phone,
			"identity_card":    loginResp.Result.UserInfo.IdentityCard,
			"class_name":       loginResp.Result.UserInfo.ClassName,
			"profession_id":    loginResp.Result.UserInfo.ProfessionID,
			"faculty_id":       loginResp.Result.UserInfo.FacultyID,
			"grade_id":         loginResp.Result.UserInfo.GradeID,
			"current_semester": loginResp.Result.UserInfo.CurrentSemester,
			"school_token":     loginResp.Result.Token,
			"token_expire_at":  time.Now().Add(24 * time.Hour),
			"last_login_at":    time.Now(),
		}); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("绑定失败", 500))
			return
		}

		onUserLogin(existingUser.ID)

		// 删除微信临时用户
		models.DB.Delete(user)

		// 生成新的JWT token指向合并后的用户
		newToken, err := utils.GenerateToken(fmt.Sprintf("%d", existingUser.ID), existingUser.Username, "wechat")
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("生成令牌失败", 500))
			return
		}

		// 构建用户信息响应，使用与登录API一致的格式
		userInfo := gin.H{
			"id":              fmt.Sprintf("%d", existingUser.ID),
			"username":        existingUser.Username,
			"realname":        existingUser.Realname,
			"avatar":          existingUser.Avatar,
			"email":           existingUser.Email,
			"phone":           existingUser.Phone,
			"birthday":        existingUser.Birthday,
			"sex":             existingUser.Sex,
			"identityCard":    existingUser.IdentityCard,
			"schoolId":        existingUser.SchoolID,
			"className":       existingUser.ClassName,
			"professionId":    existingUser.ProfessionID,
			"facultyId":       existingUser.FacultyID,
			"gradeId":         existingUser.GradeID,
			"currentSemester": existingUser.CurrentSemester,
		}

		// 返回绑定成功响应，使用标准化响应格式
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"token":    newToken,
			"userInfo": userInfo,
		}))
		return
	} else {
		// 如果不存在学号用户，直接更新当前微信用户
		user.Username = req.StudentID
		log.Printf("[DEBUG-WECHAT-BIND] userID=%d username=%s - new user, setting schoolPasswordEnc (len=%d)",
			user.ID, req.StudentID, len(req.Password))
		if err := user.SetSchoolPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("保存学校密码失败", 500))
			return
		}
		log.Printf("[DEBUG-WECHAT-BIND] userID=%d - schoolPasswordEnc set successfully (encLen=%d)",
			user.ID, len(user.SchoolPasswordEnc))
		user.Realname = loginResp.Result.UserInfo.Realname
		user.Avatar = loginResp.Result.UserInfo.Avatar
		user.Birthday = loginResp.Result.UserInfo.Birthday
		user.Sex = loginResp.Result.UserInfo.Sex
		user.Email = loginResp.Result.UserInfo.Email
		user.Phone = loginResp.Result.UserInfo.Phone
		user.IdentityCard = loginResp.Result.UserInfo.IdentityCard
		user.ClassName = loginResp.Result.UserInfo.ClassName
		user.ProfessionID = uint(loginResp.Result.UserInfo.ProfessionID)
		user.FacultyID = uint(loginResp.Result.UserInfo.FacultyID)
		user.GradeID = uint(loginResp.Result.UserInfo.GradeID)
		user.CurrentSemester = loginResp.Result.UserInfo.CurrentSemester
		user.SchoolToken = loginResp.Result.Token
		user.TokenExpireAt = time.Now().Add(24 * time.Hour)

		// 保存更新的用户信息
		if err := models.UpdateUserFields(user.ID, map[string]interface{}{
			"username":            user.Username,
			"school_password_enc": user.SchoolPasswordEnc,
			"realname":            loginResp.Result.UserInfo.Realname,
			"avatar":              loginResp.Result.UserInfo.Avatar,
			"birthday":            loginResp.Result.UserInfo.Birthday,
			"sex":                 loginResp.Result.UserInfo.Sex,
			"email":               loginResp.Result.UserInfo.Email,
			"phone":               loginResp.Result.UserInfo.Phone,
			"identity_card":       loginResp.Result.UserInfo.IdentityCard,
			"class_name":          loginResp.Result.UserInfo.ClassName,
			"profession_id":       loginResp.Result.UserInfo.ProfessionID,
			"faculty_id":          loginResp.Result.UserInfo.FacultyID,
			"grade_id":            loginResp.Result.UserInfo.GradeID,
			"current_semester":    loginResp.Result.UserInfo.CurrentSemester,
			"school_token":        loginResp.Result.Token,
			"token_expire_at":     time.Now().Add(24 * time.Hour),
		}); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("绑定失败", 500))
			return
		}

		onUserLogin(user.ID)

		// 生成新的JWT token
		newToken, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, "wechat")
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("生成令牌失败", 500))
			return
		}

		// 构建用户信息响应，使用与登录API一致的格式
		userInfo := gin.H{
			"id":              fmt.Sprintf("%d", user.ID),
			"username":        user.Username,
			"realname":        user.Realname,
			"avatar":          user.Avatar,
			"email":           user.Email,
			"phone":           user.Phone,
			"birthday":        user.Birthday,
			"sex":             user.Sex,
			"identityCard":    user.IdentityCard,
			"schoolId":        user.SchoolID,
			"className":       user.ClassName,
			"professionId":    user.ProfessionID,
			"facultyId":       user.FacultyID,
			"gradeId":         user.GradeID,
			"currentSemester": user.CurrentSemester,
		}

		// 返回绑定成功响应，使用标准化响应格式
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"token":    newToken,
			"userInfo": userInfo,
		}))
		return
	}
}
