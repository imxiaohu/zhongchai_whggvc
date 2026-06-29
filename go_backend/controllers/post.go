package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
	"gorm.io/gorm"
)

// CreatePost 创建帖子
func CreatePost(c *gin.Context) {
	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 解析请求参数
	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content" binding:"required"`
		Summary    string `json:"summary"`
		Images     string `json:"images"`
		Type       string `json:"type"`
		ClubID     uint   `json:"clubId"`
		IsOfficial bool   `json:"isOfficial"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 如果是社团帖子，检查用户是否为社团成员
	if req.ClubID > 0 {
		_, err := models.CheckClubMembership(req.ClubID, uint(userIdUint))
		if err != nil {
			c.JSON(http.StatusForbidden, utils.NewErrorResponse("您不是该社团成员，无法发帖", 403))
			return
		}
	}

	// 如果是官方帖子，检查用户权限
	if req.IsOfficial {
		user, err := models.FindUserByID(uint(userIdUint))
		if err != nil || user.UserType != "admin" {
			c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限发布官方帖子", 403))
			return
		}
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "article"
	}

	// 创建帖子
	post := &models.Post{
		Title:       req.Title,
		Content:     req.Content,
		Summary:     req.Summary,
		Images:      req.Images,
		Type:        req.Type,
		AuthorID:    uint(userIdUint),
		IsOfficial:  req.IsOfficial,
		Status:      1,
		PublishedAt: time.Now(),
	}

	// 设置社团ID，如果为0则设置为nil（官方帖子）
	if req.ClubID > 0 {
		post.ClubID = &req.ClubID
	}

	if err := models.CreatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建帖子失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(post))
}

// GetPostsList 获取帖子列表
func GetPostsList(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 解析筛选参数
	clubIdStr := c.Query("clubId")
	postType := c.Query("type")
	isOfficialStr := c.Query("isOfficial")

	var clubId uint
	if clubIdStr != "" {
		clubIdUint64, err := strconv.ParseUint(clubIdStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
			return
		}
		clubId = uint(clubIdUint64)
	}

	var isOfficial bool
	if isOfficialStr == "true" {
		isOfficial = true
	}

	// 获取当前用户ID，用于过滤被屏蔽用户的内容
	var currentUserID uint
	if userIdInterface, exists := c.Get("userId"); exists {
		if userId, ok := userIdInterface.(string); ok {
			if userIdUint, err := strconv.ParseUint(userId, 10, 64); err == nil {
				currentUserID = uint(userIdUint)
			}
		}
	}

	// 获取被屏蔽用户列表
	var blockedUserIDs []uint
	if currentUserID > 0 {
		blockedUserIDs, _ = models.GetBlockedUserIDs(currentUserID)
	}

	// 获取帖子列表
	posts, total, err := models.GetPostsList(page, pageSize, clubId, postType, isOfficial, blockedUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取帖子列表失败: "+err.Error(), 500))
		return
	}

	// 转换为安全的帖子数据
	var safePosts []models.SafePost
	for _, post := range posts {
		safePosts = append(safePosts, post.ToSafePost())
	}

	// 构造响应数据
	result := map[string]interface{}{
		"records": safePosts, // 使用安全的帖子数据
		"total":   total,
		"size":    pageSize,
		"current": page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	postIdStr := c.Param("id")
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	post, err := models.FindPostByID(uint(postId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 检查当前用户是否点赞了该帖子和是否收藏了该帖子
	var isLiked bool
	var isBookmarked bool
	if userIdInterface, exists := c.Get("userId"); exists {
		if userId, ok := userIdInterface.(string); ok {
			if userIdUint, err := strconv.ParseUint(userId, 10, 32); err == nil {
				// 检查点赞状态
				var interaction models.PostInteraction
				result := models.DB.Where("post_id = ? AND user_id = ? AND type = 'like' AND status = 1",
					postId, userIdUint).First(&interaction)
				isLiked = result.Error == nil

				// 检查收藏状态
				isBookmarked = models.IsPostBookmarked(uint(userIdUint), uint(postId))
			}
		}
	}

	// 转换为安全的帖子数据，不包含作者敏感信息
	safePost := post.ToSafePost()

	// 构造响应数据
	result := map[string]interface{}{
		"post":         safePost, // 使用安全的帖子数据
		"isLiked":      isLiked,
		"isBookmarked": isBookmarked,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	postIdStr := c.Param("id")
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取帖子信息
	post, err := models.FindPostByID(uint(postId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 检查权限（只有作者或社团管理员可以编辑）
	canEdit := post.AuthorID == uint(userIdUint)
	if !canEdit && post.ClubID != nil && *post.ClubID > 0 {
		canEdit = models.IsClubAdmin(*post.ClubID, uint(userIdUint))
	}
	if !canEdit {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限编辑该帖子", 403))
		return
	}

	// 解析请求参数
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Summary string `json:"summary"`
		Images  string `json:"images"`
		Type    string `json:"type"`
		IsTop   bool   `json:"isTop"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 更新帖子信息
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Summary != "" {
		post.Summary = req.Summary
	}
	if req.Images != "" {
		post.Images = req.Images
	}
	if req.Type != "" {
		post.Type = req.Type
	}

	// 只有管理员可以设置置顶
	if req.IsTop && post.ClubID != nil && *post.ClubID > 0 && models.IsClubAdmin(*post.ClubID, uint(userIdUint)) {
		post.IsTop = req.IsTop
	}

	if err := models.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("更新帖子失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(post))
}

// DeletePost 删除帖子
func DeletePost(c *gin.Context) {
	postIdStr := c.Param("id")
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取帖子信息
	post, err := models.FindPostByID(uint(postId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 检查权限（只有作者或社团管理员可以删除）
	canDelete := post.AuthorID == uint(userIdUint)
	if !canDelete && post.ClubID != nil && *post.ClubID > 0 {
		canDelete = models.IsClubAdmin(*post.ClubID, uint(userIdUint))
	}
	if !canDelete {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限删除该帖子", 403))
		return
	}

	if err := models.DeletePost(uint(postId)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("删除帖子失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// LikePost 点赞帖子
func LikePost(c *gin.Context) {
	postIdStr := c.Param("id")
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 检查帖子是否存在（不更新浏览量）
	var post models.Post
	if err := models.DB.Where("id = ? AND status = 1", postId).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("查询帖子失败: "+err.Error(), 500))
		}
		return
	}

	if err := models.LikePost(uint(postId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("点赞失败: "+err.Error(), 500))
		return
	}

	// 发送点赞通知
	SendLikeNotification(uint(postId), uint(userIdUint))

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// UnlikePost 取消点赞帖子
func UnlikePost(c *gin.Context) {
	postIdStr := c.Param("id")
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 检查帖子是否存在（不更新浏览量）
	var post models.Post
	if err := models.DB.Where("id = ? AND status = 1", postId).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("查询帖子失败: "+err.Error(), 500))
		}
		return
	}

	if err := models.UnlikePost(uint(postId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("取消点赞失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}
