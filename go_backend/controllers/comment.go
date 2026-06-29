package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// CreateComment 创建评论
func CreateComment(c *gin.Context) {
	// 获取当前用户ID
	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	userIdStr, ok := userIdInterface.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 解析请求参数
	var req struct {
		PostID   interface{} `json:"postId" binding:"required"`
		Content  string      `json:"content"`
		Images   string      `json:"images"` // 图片URLs，JSON数组格式
		ParentID interface{} `json:"parentId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 手动验证内容不为空（去除空格后）
	if strings.TrimSpace(req.Content) == "" && req.Images == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论内容不能为空", 400))
		return
	}

	// 转换 PostID
	var postID uint
	switch v := req.PostID.(type) {
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
			return
		}
		postID = uint(id)
	case float64:
		postID = uint(v)
	default:
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 转换 ParentID
	var parentID uint
	if req.ParentID != nil {
		switch v := req.ParentID.(type) {
		case string:
			if v != "" {
				id, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					c.JSON(http.StatusBadRequest, utils.NewErrorResponse("父评论ID格式错误", 400))
					return
				}
				parentID = uint(id)
			}
		case float64:
			parentID = uint(v)
		}
	}

	// 检查帖子是否存在
	_, err = models.FindPostByID(postID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 如果是回复评论，检查父评论是否存在
	if parentID > 0 {
		var parentComment models.Comment
		result := models.DB.First(&parentComment, parentID)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("父评论不存在", 404))
			return
		}

		// 确保父评论属于同一个帖子
		if parentComment.PostID != postID {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("父评论不属于该帖子", 400))
			return
		}
	}

	// 检查评论频率限制
	rateLimitService := services.GetCommentRateLimitService()
	if err := rateLimitService.CheckRateLimit(uint(userIdUint)); err != nil {
		c.JSON(http.StatusTooManyRequests, utils.NewErrorResponse(err.Error(), 429))
		return
	}

	// 检查垃圾内容
	if err := rateLimitService.CheckSpamContent(req.Content, uint(userIdUint)); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), 400))
		return
	}

	// 创建评论
	comment := &models.Comment{
		PostID:  postID,
		UserID:  uint(userIdUint),
		Content: req.Content,
		Images:  req.Images,
		Status:  1,
	}

	// 只有当 parentID > 0 时才设置父评论ID
	if parentID > 0 {
		comment.ParentID = &parentID
	}

	if err := models.CreateCommentWithHierarchy(comment); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建评论失败: "+err.Error(), 500))
		return
	}

	// 发送评论通知
	SendCommentNotification(comment.PostID, comment.ID, comment.UserID)

	// 预加载用户信息
	models.DB.Preload("User").First(comment, comment.ID)

	// 转换为安全的评论数据，不包含用户敏感信息
	safeComment := comment.ToSafeComment()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(safeComment))
}

// GetCommentsList 获取评论列表
func GetCommentsList(c *gin.Context) {
	// 获取帖子ID
	postIdStr := c.Query("postId")
	if postIdStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID不能为空", 400))
		return
	}

	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	sortBy := c.DefaultQuery("sortBy", "time") // time, time_desc, likes, hot

	// 获取当前用户ID（可选，用于检查点赞状态）
	var currentUserID uint = 0
	if userIdInterface, exists := c.Get("userId"); exists {
		if userId, ok := userIdInterface.(string); ok {
			if userIdUint, err := strconv.ParseUint(userId, 10, 64); err == nil {
				currentUserID = uint(userIdUint)
			}
		}
	}

	// 检查帖子是否存在
	_, err = models.FindPostByID(uint(postId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 获取评论树结构（只获取顶级评论）
	comments, total, err := models.GetCommentTree(uint(postId), page, pageSize, sortBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取评论列表失败: "+err.Error(), 500))
		return
	}

	// 转换为安全的评论数据，包含用户点赞状态
	safeComments := make([]models.SafeComment, len(comments))
	for i, comment := range comments {
		safeComments[i] = comment.ToSafeCommentWithUser(currentUserID)
	}

	// 构造响应数据
	result := map[string]interface{}{
		"records": safeComments, // 使用安全的评论数据
		"total":   total,
		"size":    pageSize,
		"current": page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// UpdateComment 更新评论
func UpdateComment(c *gin.Context) {
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
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
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取评论信息
	var comment models.Comment
	result := models.DB.First(&comment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("评论不存在", 404))
		return
	}

	// 检查权限（只有评论作者可以编辑）
	if comment.UserID != uint(userIdUint) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限编辑该评论", 403))
		return
	}

	// 解析请求参数
	var req struct {
		Content string `json:"content" binding:"required"`
		Images  string `json:"images"` // 图片URLs，JSON数组格式
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 更新评论内容
	comment.Content = req.Content
	comment.Images = req.Images
	if err := models.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("更新评论失败: "+err.Error(), 500))
		return
	}

	// 预加载用户信息
	models.DB.Preload("User").First(&comment, comment.ID)

	// 转换为安全的评论数据，不包含用户敏感信息
	safeComment := comment.ToSafeComment()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(safeComment))
}

// DeleteComment 删除评论
func DeleteComment(c *gin.Context) {
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
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
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 获取评论信息
	var comment models.Comment
	result := models.DB.Preload("Post").First(&comment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("评论不存在", 404))
		return
	}

	// 检查权限（评论作者、帖子作者或社团管理员可以删除）
	canDelete := comment.UserID == uint(userIdUint) // 评论作者
	if !canDelete && comment.Post.ID > 0 {
		canDelete = comment.Post.AuthorID == uint(userIdUint) // 帖子作者
	}
	if !canDelete && comment.Post.ID > 0 && comment.Post.ClubID != nil && *comment.Post.ClubID > 0 {
		canDelete = models.IsClubAdmin(*comment.Post.ClubID, uint(userIdUint)) // 社团管理员
	}

	if !canDelete {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限删除该评论", 403))
		return
	}

	if err := models.DeleteComment(uint(commentId)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("删除评论失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// GetCommentReplies 获取评论的回复列表
func GetCommentReplies(c *gin.Context) {
	// 获取评论ID
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	sortBy := c.DefaultQuery("sortBy", "time") // time, time_desc, likes, hot

	// 获取当前用户ID（可选，用于检查点赞状态）
	var currentUserID uint = 0
	if userIdInterface, exists := c.Get("userId"); exists {
		if userId, ok := userIdInterface.(string); ok {
			if userIdUint, err := strconv.ParseUint(userId, 10, 64); err == nil {
				currentUserID = uint(userIdUint)
			}
		}
	}

	// 检查评论是否存在
	var parentComment models.Comment
	result := models.DB.First(&parentComment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("评论不存在", 404))
		return
	}

	// 获取回复列表
	replies, total, err := models.GetCommentReplies(uint(commentId), page, pageSize, sortBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取回复列表失败: "+err.Error(), 500))
		return
	}

	// 转换为安全的评论数据，包含用户点赞状态
	safeReplies := make([]models.SafeComment, len(replies))
	for i, reply := range replies {
		safeReplies[i] = reply.ToSafeCommentWithUser(currentUserID)
	}

	// 构造响应数据
	result_data := map[string]interface{}{
		"records": safeReplies,
		"total":   total,
		"size":    pageSize,
		"current": page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result_data))
}

// ReplyToComment 回复评论
func ReplyToComment(c *gin.Context) {
	// 获取评论ID
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
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
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 解析请求参数
	var req struct {
		Content        string `json:"content"`
		Images         string `json:"images"`         // 图片URLs，JSON数组格式
		MentionedUsers string `json:"mentionedUsers"` // @提及的用户ID列表，JSON格式
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 手动验证内容不为空
	if strings.TrimSpace(req.Content) == "" && req.Images == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("回复内容不能为空", 400))
		return
	}

	// 检查评论频率限制
	rateLimitService := services.GetCommentRateLimitService()
	if err := rateLimitService.CheckRateLimit(uint(userIdUint)); err != nil {
		c.JSON(http.StatusTooManyRequests, utils.NewErrorResponse(err.Error(), 429))
		return
	}

	// 检查垃圾内容
	if err := rateLimitService.CheckSpamContent(req.Content, uint(userIdUint)); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error(), 400))
		return
	}

	// 检查父评论是否存在
	var parentComment models.Comment
	result := models.DB.First(&parentComment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("父评论不存在", 404))
		return
	}

	// 创建回复评论
	parentID := uint(commentId)
	comment := &models.Comment{
		PostID:         parentComment.PostID,
		UserID:         uint(userIdUint),
		Content:        req.Content,
		Images:         req.Images,
		ParentID:       &parentID,
		MentionedUsers: req.MentionedUsers,
		Status:         1,
	}

	if err := models.CreateCommentWithHierarchy(comment); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建回复失败: "+err.Error(), 500))
		return
	}

	// 发送回复通知
	notificationService := services.GetCommentNotificationService()

	// 发送回复通知给父评论作者
	if comment.ParentID != nil {
		//nolint:errcheck
		_ = notificationService.SendCommentReplyNotification(comment.PostID, *comment.ParentID, comment.ID, comment.UserID)
	}

	// 处理@提及通知
	if req.MentionedUsers != "" {
		mentionedUserIDs, err := notificationService.ProcessMentionedUsers(req.MentionedUsers)
		if err == nil && len(mentionedUserIDs) > 0 {
			//nolint:errcheck
			_ = notificationService.SendCommentMentionNotification(comment.ID, comment.UserID, mentionedUserIDs)
		}
	}

	// 记录评论活动
	rateLimitService.RecordCommentActivity(comment.UserID, comment.PostID, comment.ID)

	// 预加载用户信息
	models.DB.Preload("User").Preload("Parent.User").First(comment, comment.ID)

	// 转换为安全的评论数据
	safeComment := comment.ToSafeComment()

	c.JSON(http.StatusOK, utils.NewSuccessResponse(safeComment))
}

// LikeComment 点赞评论
func LikeComment(c *gin.Context) {
	// 获取评论ID
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
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
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 检查评论是否存在
	var comment models.Comment
	result := models.DB.First(&comment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("评论不存在", 404))
		return
	}

	if err := models.LikeComment(uint(commentId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("点赞失败: "+err.Error(), 500))
		return
	}

	// 发送点赞通知
	notificationService := services.GetCommentNotificationService()
	//nolint:errcheck
	//nolint:errcheck
	_ = notificationService.SendCommentLikeNotification(uint(commentId), uint(userIdUint))

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// UnlikeComment 取消点赞评论
func UnlikeComment(c *gin.Context) {
	// 获取评论ID
	commentIdStr := c.Param("id")
	commentId, err := strconv.ParseUint(commentIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("评论ID格式错误", 400))
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
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}
	userIdUint, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("用户ID格式错误", 400))
		return
	}

	// 检查评论是否存在
	var comment models.Comment
	result := models.DB.First(&comment, commentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("评论不存在", 404))
		return
	}

	if err := models.UnlikeComment(uint(commentId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("取消点赞失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// GetHotComments 获取热门评论
func GetHotComments(c *gin.Context) {
	// 获取帖子ID
	postIdStr := c.Query("postId")
	if postIdStr == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID不能为空", 400))
		return
	}

	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("帖子ID格式错误", 400))
		return
	}

	// 解析限制数量
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// 检查帖子是否存在
	_, err = models.FindPostByID(uint(postId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("帖子不存在", 404))
		return
	}

	// 获取热门评论
	comments, err := models.GetHotComments(uint(postId), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取热门评论失败: "+err.Error(), 500))
		return
	}

	// 转换为安全的评论数据
	safeComments := make([]models.SafeComment, len(comments))
	for i, comment := range comments {
		safeComments[i] = comment.ToSafeComment()
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(safeComments))
}
