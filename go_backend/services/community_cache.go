package services

import (
	"fmt"
	"time"

	"github.com/xiaohu/pingjiao/models"
)

// CommunityCacheService 社区缓存服务
type CommunityCacheService struct {
	cacheService *CacheService
}

// NewCommunityCacheService 创建社区缓存服务
func NewCommunityCacheService() *CommunityCacheService {
	return &CommunityCacheService{
		cacheService: GetCacheService(),
	}
}

// 缓存键前缀常量
const (
	CacheKeyClubList     = "community:clubs:list"
	CacheKeyClubDetail   = "community:club:detail"
	CacheKeyClubMembers  = "community:club:members"
	CacheKeyPostList     = "community:posts:list"
	CacheKeyPostDetail   = "community:post:detail"
	CacheKeyCommentList  = "community:comments:list"
	CacheKeyUserClubs    = "community:user:clubs"
	CacheKeyPopularPosts = "community:posts:popular"
	CacheKeyHotClubs     = "community:clubs:hot"
)

// 缓存时间常量
const (
	CacheTTLClubList     = 10 * time.Minute // 社团列表缓存10分钟
	CacheTTLClubDetail   = 30 * time.Minute // 社团详情缓存30分钟
	CacheTTLClubMembers  = 15 * time.Minute // 社团成员缓存15分钟
	CacheTTLPostList     = 5 * time.Minute  // 帖子列表缓存5分钟
	CacheTTLPostDetail   = 20 * time.Minute // 帖子详情缓存20分钟
	CacheTTLCommentList  = 3 * time.Minute  // 评论列表缓存3分钟
	CacheTTLUserClubs    = 30 * time.Minute // 用户社团缓存30分钟
	CacheTTLPopularPosts = 1 * time.Hour    // 热门帖子缓存1小时
	CacheTTLHotClubs     = 2 * time.Hour    // 热门社团缓存2小时
)

// GetClubsList 获取社团列表缓存
func (ccs *CommunityCacheService) GetClubsList(page, pageSize int, schoolID uint) ([]models.Club, int64, bool) {
	key := fmt.Sprintf("%s:page:%d:size:%d:school:%d", CacheKeyClubList, page, pageSize, schoolID)

	var result struct {
		Clubs []models.Club `json:"clubs"`
		Total int64         `json:"total"`
	}

	found, err := ccs.cacheService.GetJSON(key, &result)
	if err != nil || !found {
		return nil, 0, false
	}

	return result.Clubs, result.Total, true
}

// SetClubsList 设置社团列表缓存
func (ccs *CommunityCacheService) SetClubsList(page, pageSize int, schoolID uint, clubs []models.Club, total int64) {
	key := fmt.Sprintf("%s:page:%d:size:%d:school:%d", CacheKeyClubList, page, pageSize, schoolID)

	data := struct {
		Clubs []models.Club `json:"clubs"`
		Total int64         `json:"total"`
	}{
		Clubs: clubs,
		Total: total,
	}

	//nolint:errcheck
	ccs.cacheService.SetJSON(key, data, CacheTTLClubList)
}

// GetClubDetail 获取社团详情缓存
func (ccs *CommunityCacheService) GetClubDetail(clubID uint) (*models.Club, bool) {
	key := fmt.Sprintf("%s:%d", CacheKeyClubDetail, clubID)

	var club models.Club
	found, err := ccs.cacheService.GetJSON(key, &club)
	if err != nil || !found {
		return nil, false
	}

	return &club, true
}

// SetClubDetail 设置社团详情缓存
func (ccs *CommunityCacheService) SetClubDetail(club *models.Club) {
	key := fmt.Sprintf("%s:%d", CacheKeyClubDetail, club.ID)
	//nolint:errcheck
	ccs.cacheService.SetJSON(key, club, CacheTTLClubDetail)
}

// GetClubMembers 获取社团成员缓存
func (ccs *CommunityCacheService) GetClubMembers(clubID uint, page, pageSize int) ([]models.ClubMember, int64, bool) {
	key := fmt.Sprintf("%s:%d:page:%d:size:%d", CacheKeyClubMembers, clubID, page, pageSize)

	var result struct {
		Members []models.ClubMember `json:"members"`
		Total   int64               `json:"total"`
	}

	found, err := ccs.cacheService.GetJSON(key, &result)
	if err != nil || !found {
		return nil, 0, false
	}

	return result.Members, result.Total, true
}

// SetClubMembers 设置社团成员缓存
func (ccs *CommunityCacheService) SetClubMembers(clubID uint, page, pageSize int, members []models.ClubMember, total int64) {
	key := fmt.Sprintf("%s:%d:page:%d:size:%d", CacheKeyClubMembers, clubID, page, pageSize)

	data := struct {
		Members []models.ClubMember `json:"members"`
		Total   int64               `json:"total"`
	}{
		Members: members,
		Total:   total,
	}

	//nolint:errcheck
	ccs.cacheService.SetJSON(key, data, CacheTTLClubMembers)
}

// GetPostsList 获取帖子列表缓存
func (ccs *CommunityCacheService) GetPostsList(page, pageSize int, clubID uint, postType string, isOfficial bool) ([]models.Post, int64, bool) {
	key := fmt.Sprintf("%s:page:%d:size:%d:club:%d:type:%s:official:%t",
		CacheKeyPostList, page, pageSize, clubID, postType, isOfficial)

	var result struct {
		Posts []models.Post `json:"posts"`
		Total int64         `json:"total"`
	}

	found, err := ccs.cacheService.GetJSON(key, &result)
	if err != nil || !found {
		return nil, 0, false
	}

	return result.Posts, result.Total, true
}

// SetPostsList 设置帖子列表缓存
func (ccs *CommunityCacheService) SetPostsList(page, pageSize int, clubID uint, postType string, isOfficial bool, posts []models.Post, total int64) {
	key := fmt.Sprintf("%s:page:%d:size:%d:club:%d:type:%s:official:%t",
		CacheKeyPostList, page, pageSize, clubID, postType, isOfficial)

	data := struct {
		Posts []models.Post `json:"posts"`
		Total int64         `json:"total"`
	}{
		Posts: posts,
		Total: total,
	}

	//nolint:errcheck
	ccs.cacheService.SetJSON(key, data, CacheTTLPostList)
}

// GetPostDetail 获取帖子详情缓存
func (ccs *CommunityCacheService) GetPostDetail(postID uint) (*models.Post, bool) {
	key := fmt.Sprintf("%s:%d", CacheKeyPostDetail, postID)

	var post models.Post
	found, err := ccs.cacheService.GetJSON(key, &post)
	if err != nil || !found {
		return nil, false
	}

	return &post, true
}

// SetPostDetail 设置帖子详情缓存
func (ccs *CommunityCacheService) SetPostDetail(post *models.Post) {
	//nolint:errcheck
	key := fmt.Sprintf("%s:%d", CacheKeyPostDetail, post.ID)
	//nolint:errcheck
	ccs.cacheService.SetJSON(key, post, CacheTTLPostDetail)
}

// GetCommentsList 获取评论列表缓存
func (ccs *CommunityCacheService) GetCommentsList(postID uint, page, pageSize int) ([]models.Comment, int64, bool) {
	key := fmt.Sprintf("%s:%d:page:%d:size:%d", CacheKeyCommentList, postID, page, pageSize)

	var result struct {
		Comments []models.Comment `json:"comments"`
		Total    int64            `json:"total"`
	}

	found, err := ccs.cacheService.GetJSON(key, &result)
	if err != nil || !found {
		return nil, 0, false
	}

	return result.Comments, result.Total, true
}

// SetCommentsList 设置评论列表缓存
func (ccs *CommunityCacheService) SetCommentsList(postID uint, page, pageSize int, comments []models.Comment, total int64) {
	key := fmt.Sprintf("%s:%d:page:%d:size:%d", CacheKeyCommentList, postID, page, pageSize)

	data := struct {
		Comments []models.Comment `json:"comments"`
		Total    int64            `json:"total"`
	}{
		Comments: comments,
		Total:    total,
	//nolint:errcheck
	}

//nolint:errcheck

//nolint:errcheck

	//nolint:errcheck
	ccs.cacheService.SetJSON(key, data, CacheTTLCommentList)
}

// InvalidateClubCache 清除社团相关缓存
func (ccs *CommunityCacheService) InvalidateClubCache(clubID uint) {
	// 清除社团详情缓存
	detailKey := fmt.Sprintf("%s:%d", CacheKeyClubDetail, clubID)
	ccs.cacheService.Delete(detailKey)

	// 清除社团列表缓存（模糊匹配）
	// 注意：这里简化处理，实际应用中可能需要更精确的缓存失效策略
	ccs.InvalidateClubListCache()

	// 清除社团成员缓存（模糊匹配）
	ccs.InvalidateClubMembersCache(clubID)
}

// InvalidateClubListCache 清除社团列表缓存
func (ccs *CommunityCacheService) InvalidateClubListCache() {
	// 清除所有社团列表缓存
	// 使用模式匹配删除所有相关的缓存键
	pattern := CacheKeyClubList + ":*"
	ccs.cacheService.DeletePattern(pattern)
}

// InvalidateClubMembersCache 清除社团成员缓存
func (ccs *CommunityCacheService) InvalidateClubMembersCache(clubID uint) {
	// 清除指定社团的成员列表缓存（所有分页）
	pattern := fmt.Sprintf("%s:%d:*", CacheKeyClubMembers, clubID)
	ccs.cacheService.DeletePattern(pattern)
}

// InvalidatePostCache 清除帖子相关缓存
func (ccs *CommunityCacheService) InvalidatePostCache(postID uint, clubID uint) {
	// 清除帖子详情缓存
	detailKey := fmt.Sprintf("%s:%d", CacheKeyPostDetail, postID)
	ccs.cacheService.Delete(detailKey)

	// 清除帖子列表缓存
	ccs.InvalidatePostListCache()

	// 清除评论列表缓存
	ccs.InvalidateCommentListCache(postID)
}

// InvalidatePostListCache 清除帖子列表缓存
func (ccs *CommunityCacheService) InvalidatePostListCache() {
	// 清除所有帖子列表缓存（所有分页和筛选条件）
	pattern := CacheKeyPostList + ":*"
	ccs.cacheService.DeletePattern(pattern)
}

// InvalidateCommentListCache 清除评论列表缓存
func (ccs *CommunityCacheService) InvalidateCommentListCache(postID uint) {
	// 清除指定帖子的评论列表缓存（所有分页）
	pattern := fmt.Sprintf("%s:%d:*", CacheKeyCommentList, postID)
	ccs.cacheService.DeletePattern(pattern)
}

// 全局社区缓存服务实例
var globalCommunityCacheService *CommunityCacheService

// InitCommunityCacheService 初始化全局社区缓存服务
func InitCommunityCacheService() {
	globalCommunityCacheService = NewCommunityCacheService()
}

// GetCommunityCacheService 获取全局社区缓存服务实例
func GetCommunityCacheService() *CommunityCacheService {
	if globalCommunityCacheService == nil {
		InitCommunityCacheService()
	}
	return globalCommunityCacheService
}
