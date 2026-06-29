// Package controllers handles HTTP request handlers for authentication endpoints.
// This file contains login-related functions.
package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
	"gorm.io/gorm"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
}

// WechatLoginRequest 微信登录请求结构
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WechatSessionResponse 微信登录会话响应结构
type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// SchoolBindRequest 学校账号绑定请求结构
type SchoolBindRequest struct {
	StudentID  string `json:"studentId" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Captcha    string `json:"captcha" binding:"required"`
	SchoolName string `json:"schoolName"`
	ClientID   string `json:"clientId"`
}

// ScloudLogin 处理Web端登录请求
func ScloudLogin(c *gin.Context) {
	var req LoginRequest

	// 解析请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("请求数据格式错误", 400))
		return
	}

	// 基本参数验证
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户名和密码不能为空", 400))
		return
	}

	// 验证码验证（实际项目中应从Redis或其他存储中获取）
	captchaCode, _ := c.Cookie("captcha_code")
	if captchaCode != "" && req.Captcha != "" && !strings.EqualFold(captchaCode, req.Captcha) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("验证码错误", 400))
		return
	}

	user, err := models.FindUserByUsername(req.Username)
	if err == nil && user.CheckPassword(req.Password) {
		log.Printf("[DEBUG-LOGIN] userID=%d username=%s - local password match, setting schoolPasswordEnc (len=%d)",
			user.ID, req.Username, len(req.Password))
		err2 := user.SetSchoolPassword(req.Password)
		log.Printf("[DEBUG-LOGIN] userID=%d - SetSchoolPassword result: err=%v encLen=%d",
			user.ID, err2, len(user.SchoolPasswordEnc))
		user.LastLoginAt = time.Now()
		//nolint:errcheck
		models.UpdateUserFields(user.ID, map[string]interface{}{
			"school_password_enc": user.SchoolPasswordEnc,
			"last_login_at":       user.LastLoginAt,
		})
		services.CacheArchiveOnLogin(user.ID, user.Username, user.Password, user.SchoolToken)
		token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, user.UserType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成令牌失败", 500))
			return
		}
		c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
			"token": token,
			"userInfo": gin.H{
				"id":              fmt.Sprintf("%d", user.ID),
				"username":        user.Username,
				"realname":        user.Realname,
				"avatar":          user.Avatar,
				"birthday":        user.Birthday,
				"sex":             user.Sex,
				"email":           user.Email,
				"phone":           user.Phone,
				"className":       user.ClassName,
				"schoolId":        user.SchoolID,
				"professionId":    user.ProfessionID,
				"facultyId":       user.FacultyID,
				"gradeId":         user.GradeID,
				"currentSemester": user.CurrentSemester,
				"identityCard":    user.IdentityCard,
			},
		}))
		return
	}

	proxyClient := utils.NewProxyClient()
	loginResp, err := proxyClient.ProxyLogin(req.Username, req.Password)
	if err != nil || !loginResp.Success {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("用户名或密码错误", 401))
		return
	}

	user, err = upsertUserFromSchoolLogin(req.Username, req.Password, loginResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("登录失败", 500))
		return
	}

	// 异步预缓存学生档案
	services.CacheArchiveOnLogin(user.ID, user.Username, user.Password, user.SchoolToken)

	// 生成JWT令牌
	token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, user.UserType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("生成令牌失败", 500))
		return
	}

	// 返回登录成功响应
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"token": token,
		"userInfo": gin.H{
			"id":              fmt.Sprintf("%d", user.ID),
			"username":        user.Username,
			"realname":        user.Realname,
			"avatar":          user.Avatar,
			"birthday":        user.Birthday,
			"sex":             user.Sex,
			"email":           user.Email,
			"phone":           user.Phone,
			"className":       user.ClassName,
			"schoolId":        user.SchoolID,
			"professionId":    user.ProfessionID,
			"facultyId":       user.FacultyID,
			"gradeId":         user.GradeID,
			"currentSemester": user.CurrentSemester,
			"identityCard":    user.IdentityCard,
		},
	}))
}

// ApiMSysMLogin 处理移动端登录请求
func ApiMSysMLogin(c *gin.Context) {
	var req LoginRequest

	// 解析请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("请求数据格式错误", 400))
		return
	}

	// 基本参数验证
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, utils.NewStandardErrorResponse("用户名和密码不能为空", 400))
		return
	}

	user, err := models.FindUserByUsername(req.Username)
	if err == nil && user.CheckPassword(req.Password) {
		//nolint:errcheck
		_ = user.SetSchoolPassword(req.Password)
		//nolint:errcheck
		models.UpdateUserFields(user.ID, map[string]interface{}{
			"school_password_enc": user.SchoolPasswordEnc,
			"last_login_at":      time.Now(),
		})
		services.CacheArchiveOnLogin(user.ID, user.Username, user.Password, user.SchoolToken)
		token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, "mobile")
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("生成令牌失败", 500))
			return
		}
		c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
			"userInfo": gin.H{
				"id":              fmt.Sprintf("%d", user.ID),
				"username":        user.Username,
				"realname":        user.Realname,
				"avatar":          user.Avatar,
				"birthday":        user.Birthday,
				"sex":             user.Sex,
				"email":           user.Email,
				"phone":           user.Phone,
				"className":       user.ClassName,
				"schoolId":        user.SchoolID,
				"professionId":    user.ProfessionID,
				"facultyId":       user.FacultyID,
				"gradeId":         user.GradeID,
				"currentSemester": user.CurrentSemester,
				"identityCard":    user.IdentityCard,
			},
			"token": token,
		}))
		return
	}

	proxyClient := utils.NewProxyClient()
	loginResp, err := proxyClient.ProxyLogin(req.Username, req.Password)
	if err != nil || !loginResp.Success {
		c.JSON(http.StatusUnauthorized, utils.NewStandardErrorResponse("用户名或密码错误", 401))
		return
	}

	user, err = upsertUserFromSchoolLogin(req.Username, req.Password, loginResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("登录失败", 500))
		return
	}

	// 异步预缓存学生档案
	services.CacheArchiveOnLogin(user.ID, user.Username, user.Password, user.SchoolToken)

	// 生成JWT令牌
	token, err := utils.GenerateToken(fmt.Sprintf("%d", user.ID), user.Username, "mobile")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewStandardErrorResponse("生成令牌失败", 500))
		return
	}

	// 返回登录成功响应，格式与学校服务器保持一致
	c.JSON(http.StatusOK, utils.NewStandardSuccessResponse(gin.H{
		"userInfo": gin.H{
			"id":              fmt.Sprintf("%d", user.ID),
			"username":        user.Username,
			"realname":        user.Realname,
			"avatar":          user.Avatar,
			"birthday":        user.Birthday,
			"sex":             user.Sex,
			"email":           user.Email,
			"phone":           user.Phone,
			"className":       user.ClassName,
			"schoolId":        user.SchoolID,
			"professionId":    user.ProfessionID,
			"facultyId":       user.FacultyID,
			"gradeId":         user.GradeID,
			"currentSemester": user.CurrentSemester,
			"identityCard":    user.IdentityCard,
		},
		"token": token,
	}))
}

func upsertUserFromSchoolLogin(username string, password string, loginResp *utils.LoginResponse) (*models.User, error) {
	now := time.Now()

	user, err := models.FindUserByUsername(username)
	if err != nil {
		deletedUser, deletedErr := models.FindUserByUsernameIncludeDeleted(username)
		if deletedErr == nil {
			deletedUser.DeletedAt = gorm.DeletedAt{}
			if err := deletedUser.SetPassword(password); err != nil {
				return nil, err
			}
			_ = deletedUser.SetSchoolPassword(password)
			log.Printf("[DEBUG-UPSERT] userID=%d username=%s (deleted-reactivated) - SetSchoolPassword err=%v encLen=%d",
				deletedUser.ID, username, nil, len(deletedUser.SchoolPasswordEnc))
			deletedUser.Realname = loginResp.Result.UserInfo.Realname
			deletedUser.Nickname = loginResp.Result.UserInfo.Realname
			deletedUser.Avatar = loginResp.Result.UserInfo.Avatar
			deletedUser.Birthday = loginResp.Result.UserInfo.Birthday
			deletedUser.Sex = loginResp.Result.UserInfo.Sex
			deletedUser.Email = loginResp.Result.UserInfo.Email
			deletedUser.Phone = loginResp.Result.UserInfo.Phone
			deletedUser.IdentityCard = loginResp.Result.UserInfo.IdentityCard
			deletedUser.ClassName = loginResp.Result.UserInfo.ClassName
			deletedUser.ProfessionID = uint(loginResp.Result.UserInfo.ProfessionID)
			deletedUser.FacultyID = uint(loginResp.Result.UserInfo.FacultyID)
			deletedUser.GradeID = uint(loginResp.Result.UserInfo.GradeID)
			deletedUser.CurrentSemester = loginResp.Result.UserInfo.CurrentSemester
			deletedUser.SchoolToken = loginResp.Result.Token
			deletedUser.TokenExpireAt = now.Add(24 * time.Hour)
			deletedUser.LastLoginAt = now
			deletedUser.Status = 1
			if err := models.DB.Unscoped().Save(deletedUser).Error; err != nil {
				return nil, err
			}
			return deletedUser, nil
		}

		user = &models.User{
			Username:        username,
			Realname:        loginResp.Result.UserInfo.Realname,
			Nickname:        loginResp.Result.UserInfo.Realname,
			Avatar:          loginResp.Result.UserInfo.Avatar,
			Birthday:        loginResp.Result.UserInfo.Birthday,
			Sex:             loginResp.Result.UserInfo.Sex,
			Email:           loginResp.Result.UserInfo.Email,
			Phone:           loginResp.Result.UserInfo.Phone,
			IdentityCard:    loginResp.Result.UserInfo.IdentityCard,
			UserType:        "student",
			ClassName:       loginResp.Result.UserInfo.ClassName,
			ProfessionID:    uint(loginResp.Result.UserInfo.ProfessionID),
			FacultyID:       uint(loginResp.Result.UserInfo.FacultyID),
			GradeID:         uint(loginResp.Result.UserInfo.GradeID),
			CurrentSemester: loginResp.Result.UserInfo.CurrentSemester,
			SchoolToken:     loginResp.Result.Token,
			TokenExpireAt:   now.Add(24 * time.Hour),
			LastLoginAt:     now,
			Status:          1,
		}
		if err := user.SetPassword(password); err != nil {
			return nil, err
		}
		_ = user.SetSchoolPassword(password)
		schools, _ := models.GetAllSchools()
		if len(schools) > 0 {
			user.SchoolID = schools[0].ID
		}
		if err := models.CreateUser(user); err != nil {
			return nil, err
		}
		return user, nil
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	_ = user.SetSchoolPassword(password)
	user.Realname = loginResp.Result.UserInfo.Realname
	user.Nickname = loginResp.Result.UserInfo.Realname
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
	user.TokenExpireAt = now.Add(24 * time.Hour)
	user.LastLoginAt = now
	updates := map[string]interface{}{
		"password":          user.Password,
		"school_password_enc": user.SchoolPasswordEnc,
		"nickname":          loginResp.Result.UserInfo.Realname,
		"avatar":            loginResp.Result.UserInfo.Avatar,
		"birthday":          loginResp.Result.UserInfo.Birthday,
		"sex":               loginResp.Result.UserInfo.Sex,
		"email":             loginResp.Result.UserInfo.Email,
		"phone":             loginResp.Result.UserInfo.Phone,
		"identity_card":     loginResp.Result.UserInfo.IdentityCard,
		"class_name":        loginResp.Result.UserInfo.ClassName,
		"profession_id":     loginResp.Result.UserInfo.ProfessionID,
		"faculty_id":       loginResp.Result.UserInfo.FacultyID,
		"grade_id":         loginResp.Result.UserInfo.GradeID,
		"current_semester": loginResp.Result.UserInfo.CurrentSemester,
		"school_token":     loginResp.Result.Token,
		"token_expire_at":  now.Add(24 * time.Hour),
		"last_login_at":    now,
	}
	if err := models.UpdateUserFields(user.ID, updates); err != nil {
		return nil, err
	}
	return user, nil
}

// onUserLogin 用户登录后的统一处理钩子
// 1. 更新用户最后活跃时间
// 2. 如果用户开启了个人信息缓存且之前被暂停，则自动恢复缓存
func onUserLogin(userID uint) {
	// 更新最后活跃时间
	if err := models.UpdateLastActiveAt(userID); err != nil {
		log.Printf("更新用户 %d 活跃时间失败: %v", userID, err)
	}

	// 恢复被暂停的个人信息缓存（如有）
	if syncService != nil {
		syncService.ResumePersonalInfoCacheIfNeeded(userID)
	}
}
