package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"

	"github.com/gin-gonic/gin"
)

// CommunityTermsController 社区须知控制器
type CommunityTermsController struct{}

// getUserID 从 context 中提取用户 ID（string -> uint64）
func getUserID(c *gin.Context) (uint64, error) {
	userIDInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未登录", 401))
		return 0, fmt.Errorf("unauthorized")
	}
	userIDStr, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("用户ID格式错误", 500))
		return 0, fmt.Errorf("invalid user id")
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("用户ID格式错误", 500))
		return 0, fmt.Errorf("invalid user id")
	}
	return userID, nil
}

// GetTermsStatus 获取用户社区须知同意状态
func (ctrl *CommunityTermsController) GetTermsStatus(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		return
	}

	var user models.User
	if err := models.DB.Select("community_terms_agreed, community_terms_agreed_at").
		First(&user, uint(userID)).Error; err != nil {
		// migration 未执行（列不存在）或其他 db 错误，都视为未同意
		c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
			"agreed":         false,
			"agreedAt":       nil,
			"currentVersion": "1.0",
		}))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"agreed":         user.CommunityTermsAgreed,
		"agreedAt":       user.CommunityTermsAgreedAt,
		"currentVersion": "1.0",
	}))
}

// AgreeTerms 同意社区须知
func (ctrl *CommunityTermsController) AgreeTerms(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		return
	}

	now := time.Now()
	if err := models.DB.Model(&models.User{}).
		Where("id = ?", uint(userID)).
		Updates(map[string]interface{}{
			"community_terms_agreed":     true,
			"community_terms_agreed_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("保存失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"agreed":   true,
		"agreedAt": now,
	}))
}
