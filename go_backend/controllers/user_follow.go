package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// FollowUserHandler 关注用户
func FollowUserHandler(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("请先登录", 401))
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	if uint(userID) == uint(targetID) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("不能关注自己", 400))
		return
	}

	// 检查目标用户是否存在
	targetUser, err := models.FindUserByID(uint(targetID))
	if err != nil || targetUser.ID == 0 {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("用户不存在", 404))
		return
	}

	if err := models.FollowUser(uint(userID), uint(targetID)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("关注失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "关注成功",
	}))
}

// UnfollowUserHandler 取消关注
func UnfollowUserHandler(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("请先登录", 401))
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	if err := models.UnfollowUser(uint(userID), uint(targetID)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("取消关注失败", 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"message": "已取消关注",
	}))
}

// GetFollowers 获取用户粉丝列表
func GetFollowers(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	followers, total, err := models.GetFollowers(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取粉丝列表失败", 500))
		return
	}

	records := make([]models.SafeUser, 0)
	for _, f := range followers {
		records = append(records, f.Follower.ToSafeUser())
	}

	currentUID, _ := utils.GetUserIDFromContext(c)
	followedMap := make(map[uint]bool)
	if currentUID > 0 {
		for _, f := range followers {
			followedMap[f.Follower.ID] = models.IsFollowing(uint(currentUID), f.Follower.ID)
		}
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"size":      pageSize,
		"followed":  followedMap,
	}))
}

// GetFollowing 获取用户关注列表
func GetFollowing(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	following, total, err := models.GetFollowing(uint(userID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取关注列表失败", 500))
		return
	}

	records := make([]models.SafeUser, 0)
	for _, f := range following {
		records = append(records, f.Following.ToSafeUser())
	}

	currentUID, _ := utils.GetUserIDFromContext(c)
	followedMap := make(map[uint]bool)
	if currentUID > 0 {
		for _, f := range following {
			followedMap[f.Following.ID] = models.IsFollowing(uint(currentUID), f.Following.ID)
		}
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"size":      pageSize,
		"followed":  followedMap,
	}))
}
