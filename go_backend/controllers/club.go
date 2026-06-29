package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// CreateClub 创建社团
func CreateClub(c *gin.Context) {
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

	// 获取用户信息
	user, err := models.FindUserByID(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("用户不存在", 404))
		return
	}

	// 解析请求参数
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		LogoURL     string `json:"logoUrl"`
		Tags        string `json:"tags"`
		ContactInfo string `json:"contactInfo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 创建社团
	club := &models.Club{
		Name:          req.Name,
		Description:   req.Description,
		LogoURL:       req.LogoURL,
		CreatorID:     uint(userIdUint),
		SchoolID:      user.SchoolID,
		Tags:          req.Tags,
		ContactInfo:   req.ContactInfo,
		EstablishedAt: time.Now(),
		Status:        1,
	}

	if err := models.CreateClub(club); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建社团失败: "+err.Error(), 500))
		return
	}

	// 创建者自动成为管理员
	if err := models.JoinClub(club.ID, uint(userIdUint)); err != nil {
		// 如果加入失败，记录日志但不影响创建结果
		// log.Printf("创建者加入社团失败: %v", err)
	} else {
		// 更新创建者为管理员
		member, _ := models.CheckClubMembership(club.ID, uint(userIdUint))
		if member != nil {
			member.Role = "admin"
			models.DB.Save(member)
		}
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(club))
}

// GetClubsList 获取社团列表
func GetClubsList(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	schoolIdStr := c.Query("schoolId")

	var schoolId uint
	if schoolIdStr != "" {
		schoolIdUint64, err := strconv.ParseUint(schoolIdStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse("学校ID格式错误", 400))
			return
		}
		schoolId = uint(schoolIdUint64)
	}

	// 从数据库获取社团列表
	clubs, total, err := models.GetClubsList(page, pageSize, schoolId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取社团列表失败: "+err.Error(), 500))
		return
	}

	// 获取当前用户ID，检查每个社团的加入状态
	uid, _ := utils.GetUserIDFromContext(c)
	memberMap := make(map[uint]bool)
	if uid > 0 && len(clubs) > 0 {
		clubIDs := make([]uint, len(clubs))
		for i, club := range clubs {
			clubIDs[i] = club.ID
		}
		memberMap = models.GetUserMembershipMap(uid, clubIDs)
	}

	// 转换为安全的社团数据，不包含创建者敏感信息
	safeClubs := make([]models.SafeClub, len(clubs))
	for i, club := range clubs {
		safeClubs[i] = club.ToSafeClub(memberMap[club.ID])
	}

	// 构造响应数据
	result := map[string]interface{}{
		"records": safeClubs, // 使用安全的社团数据
		"total":   total,
		"size":    pageSize,
		"current": page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// GetMyClubs 获取我的社团列表
func GetMyClubs(c *gin.Context) {
	// 获取用户ID
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

	// 获取用户加入的社团列表
	clubs, err := models.GetUserClubs(uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取我的社团失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(clubs))
}

// GetClubDetail 获取社团详情
func GetClubDetail(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
		return
	}

	club, err := models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	// 检查当前用户是否为社团成员
	var isMember bool
	var isAdmin bool
	if userIdInterface, exists := c.Get("userId"); exists {
		if userId, ok := userIdInterface.(string); ok {
			if userIdUint, err := strconv.ParseUint(userId, 10, 64); err == nil {
				_, err := models.CheckClubMembership(uint(clubId), uint(userIdUint))
				isMember = err == nil
				isAdmin = models.IsClubAdmin(uint(clubId), uint(userIdUint))
			}
		}
	}

	// 转换为安全的社团数据，不包含创建者敏感信息
	safeClub := club.ToSafeClub(isMember)

	// 构造响应数据
	result := map[string]interface{}{
		"club":     safeClub, // 使用安全的社团数据
		"isMember": isMember,
		"isAdmin":  isAdmin,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// UpdateClub 更新社团信息
func UpdateClub(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
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

	// 检查权限
	if !models.IsClubAdmin(uint(clubId), uint(userIdUint)) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限操作", 403))
		return
	}

	// 获取社团信息
	club, err := models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	// 解析请求参数
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LogoURL     string `json:"logoUrl"`
		Tags        string `json:"tags"`
		ContactInfo string `json:"contactInfo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 更新社团信息
	if req.Name != "" {
		club.Name = req.Name
	}
	if req.Description != "" {
		club.Description = req.Description
	}
	if req.LogoURL != "" {
		club.LogoURL = req.LogoURL
	}
	if req.Tags != "" {
		club.Tags = req.Tags
	}
	if req.ContactInfo != "" {
		club.ContactInfo = req.ContactInfo
	}

	if err := models.UpdateClub(club); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("更新社团失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(club))
}

// DeleteClub 删除社团
func DeleteClub(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
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

	// 检查权限（只有创建者可以删除社团）
	club, err := models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	if club.CreatorID != uint(userIdUint) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("只有创建者可以删除社团", 403))
		return
	}

	if err := models.DeleteClub(uint(clubId)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("删除社团失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// JoinClub 加入社团
func JoinClub(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
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

	// 检查社团是否存在
	_, err = models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	// 检查是否已经是成员
	_, err = models.CheckClubMembership(uint(clubId), uint(userIdUint))
	if err == nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("您已经是该社团成员", 400))
		return
	}

	// 加入社团
	if err := models.JoinClub(uint(clubId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("加入社团失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// LeaveClub 退出社团
func LeaveClub(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
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

	// 检查是否为社团成员
	_, err = models.CheckClubMembership(uint(clubId), uint(userIdUint))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("您不是该社团成员", 400))
		return
	}

	// 检查是否为创建者（创建者不能退出）
	club, err := models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	if club.CreatorID == uint(userIdUint) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("创建者不能退出社团", 400))
		return
	}

	// 退出社团
	if err := models.LeaveClub(uint(clubId), uint(userIdUint)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("退出社团失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}

// GetClubMembers 获取社团成员列表
func GetClubMembers(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 检查社团是否存在
	_, err = models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	// 获取成员列表
	members, total, err := models.GetClubMembers(uint(clubId), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("获取成员列表失败: "+err.Error(), 500))
		return
	}

	// 转换为安全的成员数据，不包含用户敏感信息
	safeMembers := make([]models.SafeClubMember, len(members))
	for i, member := range members {
		safeMembers[i] = member.ToSafeClubMember()
	}

	// 构造响应数据
	result := map[string]interface{}{
		"records": safeMembers, // 使用安全的成员数据
		"total":   total,
		"size":    pageSize,
		"current": page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// UpdateMemberRole 更新成员角色
func UpdateMemberRole(c *gin.Context) {
	clubIdStr := c.Param("id")
	clubId, err := strconv.ParseUint(clubIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("社团ID格式错误", 400))
		return
	}

	memberIdStr := c.Param("memberId")
	memberId, err := strconv.ParseUint(memberIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("成员ID格式错误", 400))
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

	// 检查权限（只有管理员可以更新成员角色）
	if !models.IsClubAdmin(uint(clubId), uint(userIdUint)) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("无权限操作", 403))
		return
	}

	// 解析请求参数
	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("参数错误: "+err.Error(), 400))
		return
	}

	// 验证角色值
	if req.Role != "admin" && req.Role != "member" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("角色值无效", 400))
		return
	}

	// 检查目标成员是否存在
	member, err := models.CheckClubMembership(uint(clubId), uint(memberId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("成员不存在", 404))
		return
	}

	// 检查是否为创建者（创建者角色不能被修改）
	club, err := models.FindClubByID(uint(clubId))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("社团不存在", 404))
		return
	}

	if club.CreatorID == uint(memberId) {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("不能修改创建者的角色", 400))
		return
	}

	// 更新成员角色
	member.Role = req.Role
	if err := models.DB.Save(member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("更新成员角色失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(member))
}
