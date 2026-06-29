package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// GetRecommendData 获取推荐数据 (包括推荐社团和推荐动态)
func GetRecommendData(c *gin.Context) {
	// 1. 获取当前用户ID（用于过滤屏蔽的用户，如果有的话）
	uid, _ := utils.GetUserIDFromContext(c)

	// 2. 获取推荐社团 (官方社团或成员最多的社团)
	clubs, _, err := models.GetClubsList(1, 10, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取推荐社团失败", 500))
		return
	}

	// 3. 获取推荐动态 (热门动态或置顶动态)
	// 这里简单获取最新动态，实际项目中可以根据算法推荐
	var blockedIDs []uint
	if uid > 0 {
		blockedIDs, _ = models.GetBlockedUserIDs(uid)
	}
	posts, _, err := models.GetPostsList(1, 10, 0, "", false, blockedIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取推荐动态失败", 500))
		return
	}

	// 4. 转换为安全模型返回
	safeClubs := make([]models.SafeClub, 0)
	memberMap := make(map[uint]bool)
	if uid > 0 && len(clubs) > 0 {
		clubIDs := make([]uint, len(clubs))
		for i, club := range clubs {
			clubIDs[i] = club.ID
		}
		memberMap = models.GetUserMembershipMap(uid, clubIDs)
	}
	for _, club := range clubs {
		safeClubs = append(safeClubs, club.ToSafeClub(memberMap[club.ID]))
	}

	safePosts := make([]models.SafePost, 0)
	for _, post := range posts {
		safePosts = append(safePosts, post.ToSafePost())
	}

	// 5. 返回结果
	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"clubs": safeClubs,
		"posts": safePosts,
	}))
}

// UserProfileResponse 用户主页响应结构
type UserProfileResponse struct {
	ID              uint   `json:"id"`
	Realname        string `json:"realname"`
	Nickname        string `json:"nickname"`
	Username        string `json:"username"`
	Avatar          string `json:"avatar"`
	CoverColor      string `json:"coverColor"`
	PostsCount      int    `json:"postsCount"`
	FollowersCount  int    `json:"followersCount"`
	FollowingCount  int    `json:"followingCount"`
	LikesCount      int    `json:"likesCount"`
	IsFollowed      bool   `json:"isFollowed"`
	IsMutual        bool   `json:"isMutual"` // 是否互相关注
	IsOfficial      bool   `json:"isOfficial"`
}

// GetUserProfile 获取用户主页信息
func GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 查询用户
	user, err := models.FindUserByID(uint(userID))
	if err != nil || user.ID == 0 {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("用户不存在", 404))
		return
	}

	// 统计该用户的帖子数
	var postsCount int64
	models.DB.Model(&models.Post{}).Where("author_id = ? AND status = 1", userID).Count(&postsCount)

	// 获取真实关注数据
	followersCount := models.GetFollowersCount(uint(userID))
	followingCount := models.GetFollowingCount(uint(userID))

	// 检查当前登录用户是否关注了此用户，以及是否互相关注
	currentUID, _ := utils.GetUserIDFromContext(c)
	isFollowed := false
	isMutual := false
	if currentUID > 0 && currentUID != uint(userID) {
		isFollowed = models.IsFollowing(currentUID, uint(userID))
		if isFollowed {
			isMutual = models.IsFollowing(uint(userID), currentUID)
		}
	}

	resp := UserProfileResponse{
		ID:             user.ID,
		Realname:       user.Realname,
		Nickname:       user.Nickname,
		Username:       user.Username,
		Avatar:         user.Avatar,
		PostsCount:     int(postsCount),
		FollowersCount: int(followersCount),
		FollowingCount: int(followingCount),
		LikesCount:     0,
		IsFollowed:     isFollowed,
		IsMutual:       isMutual,
		IsOfficial:     user.UserType == "admin" || user.UserType == "teacher",
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(resp))
}

// GetUserPosts 获取用户发布的帖子列表
func GetUserPosts(c *gin.Context) {
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

	offset := (page - 1) * pageSize

	// 直接按作者ID查询，排除被屏蔽用户的帖子
	var currentUserID uint
	if uid, _ := utils.GetUserIDFromContext(c); uid > 0 {
		currentUserID = uid
	}
	var blockedUserIDs []uint
	if currentUserID > 0 {
		blockedUserIDs, _ = models.GetBlockedUserIDs(currentUserID)
	}

	query := models.DB.Model(&models.Post{}).Where("author_id = ? AND status = 1", userID)
	if len(blockedUserIDs) > 0 {
		query = query.Where("author_id NOT IN ?", blockedUserIDs)
	}

	var total int64
	query.Count(&total)

	var posts []models.Post
	if err := query.Preload("Author").Preload("Club").
		Order("published_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取帖子列表失败", 500))
		return
	}

	safePosts := make([]models.SafePost, 0)
	for _, post := range posts {
		safePosts = append(safePosts, post.ToSafePost())
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(gin.H{
		"records": safePosts,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	}))
}
