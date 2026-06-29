package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"
)

// Post 帖子模型
type Post struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Title          string         `gorm:"size:200;not null" json:"title"`        // 标题
	Content        string         `gorm:"type:text" json:"content"`              // 内容
	Summary        string         `gorm:"size:500" json:"summary"`               // 摘要
	Images         string         `gorm:"type:text" json:"images"`               // 图片URLs，JSON数组格式
	Type           string         `gorm:"size:20;default:'article'" json:"type"` // 类型: article-文章, announcement-公告, activity-活动
	ClubID         *uint          `json:"clubId"`                                // 社团ID，为空表示官方帖子
	Club           *Club          `gorm:"foreignKey:ClubID" json:"club,omitempty"`
	AuthorID       uint           `gorm:"not null" json:"authorId"` // 作者ID
	Author         User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	LikesCount     int            `gorm:"default:0" json:"likesCount"`     // 点赞数
	CommentsCount  int            `gorm:"default:0" json:"commentsCount"`  // 评论数
	BookmarksCount int            `gorm:"default:0" json:"bookmarksCount"` // 收藏数
	ViewsCount     int            `gorm:"default:0" json:"viewsCount"`     // 浏览数
	IsTop          bool           `gorm:"default:false" json:"isTop"`      // 是否置顶
	IsOfficial     bool           `gorm:"default:false" json:"isOfficial"` // 是否官方帖子
	Status         int            `gorm:"default:1" json:"status"`         // 状态: 0-草稿, 1-已发布, 2-已删除
	PublishedAt    time.Time      `json:"publishedAt"`                     // 发布时间
}

// SafePost 安全的帖子结构体，用于API返回，不包含作者敏感信息
type SafePost struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Summary        string    `json:"summary"`
	Images         string    `json:"images"`
	Type           string    `json:"type"`
	ClubID         *uint     `json:"clubId"`
	Club           *Club     `json:"club,omitempty"`
	Author         SafeUser  `json:"author"` // 使用安全的用户信息
	IsOfficial     bool      `json:"isOfficial"`
	IsTop          bool      `json:"isTop"`
	ViewsCount     int       `json:"viewsCount"`
	LikesCount     int       `json:"likesCount"`
	CommentsCount  int       `json:"commentsCount"`
	BookmarksCount int       `json:"bookmarksCount"`
	PublishedAt    time.Time `json:"publishedAt"`
	Status         int       `json:"status"`
}

func normalizeImagesJSON(images string) string {
	raw := sanitizeBackticks(images)
	if raw == "" {
		return "[]"
	}

	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		clean := make([]string, 0, len(arr))
		for _, it := range arr {
			v := sanitizeBackticks(it)
			if v != "" {
				clean = append(clean, v)
			}
		}
		b, _ := json.Marshal(clean)
		return string(b)
	}

	re := regexp.MustCompile(`https?://[^\s"'` + "`" + `\\\]]+`)
	matches := re.FindAllString(raw, -1)
	if len(matches) == 0 {
		return "[]"
	}
	clean := make([]string, 0, len(matches))
	for _, it := range matches {
		v := sanitizeBackticks(it)
		if v != "" {
			clean = append(clean, v)
		}
	}
	b, _ := json.Marshal(clean)
	return string(b)
}

// ToSafePost 将Post转换为SafePost
func (p *Post) ToSafePost() SafePost {
	return SafePost{
		ID:             p.ID,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		Title:          p.Title,
		Content:        p.Content,
		Summary:        p.Summary,
		Images:         normalizeImagesJSON(p.Images),
		Type:           p.Type,
		ClubID:         p.ClubID,
		Club:           p.Club,
		Author:         p.Author.ToSafeUser(), // 转换为安全的用户信息
		IsOfficial:     p.IsOfficial,
		IsTop:          p.IsTop,
		ViewsCount:     p.ViewsCount,
		LikesCount:     p.LikesCount,
		CommentsCount:  p.CommentsCount,
		BookmarksCount: p.BookmarksCount,
		PublishedAt:    p.PublishedAt,
		Status:         p.Status,
	}
}

// PostInteraction 帖子互动模型
type PostInteraction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	PostID    uint           `gorm:"not null" json:"postId"` // 帖子ID
	Post      Post           `gorm:"foreignKey:PostID" json:"post,omitempty"`
	UserID    uint           `gorm:"not null" json:"userId"` // 用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type      string         `gorm:"size:20;not null" json:"type"` // 类型: like-点赞, favorite-收藏, report-举报
	Status    int            `gorm:"default:1" json:"status"`      // 状态: 0-取消, 1-有效
}

// CommentInteraction 评论互动模型
type CommentInteraction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	CommentID uint           `gorm:"not null" json:"commentId"` // 评论ID
	Comment   Comment        `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	UserID    uint           `gorm:"not null" json:"userId"` // 用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type      string         `gorm:"size:20;not null" json:"type"` // 类型: like-点赞, report-举报
	Status    int            `gorm:"default:1" json:"status"`      // 状态: 0-取消, 1-有效
}

// Comment 评论模型
type Comment struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	PostID         uint           `gorm:"not null;index" json:"postId"` // 帖子ID
	Post           Post           `gorm:"foreignKey:PostID" json:"post,omitempty"`
	UserID         uint           `gorm:"not null;index" json:"userId"` // 用户ID
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content        string         `gorm:"type:text;not null" json:"content"` // 评论内容
	Images         string         `gorm:"type:text" json:"images"`           // 图片URLs，JSON数组格式
	ParentID       *uint          `gorm:"index" json:"parentId"`             // 父评论ID，NULL表示顶级评论
	Parent         *Comment       `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	RootID         *uint          `gorm:"index" json:"rootId"`                 // 根评论ID，用于快速查找评论树
	Level          int            `gorm:"default:0;index" json:"level"`        // 评论层级，0为顶级评论
	Path           string         `gorm:"type:varchar(500);index" json:"path"` // 评论路径，如"1/2/3"，用于排序和查询
	RepliesCount   int            `gorm:"default:0" json:"repliesCount"`       // 回复数量
	LikesCount     int            `gorm:"default:0" json:"likesCount"`         // 点赞数
	MentionedUsers string         `gorm:"type:text" json:"mentionedUsers"`     // @提及的用户ID列表，JSON格式
	IsHot          bool           `gorm:"default:false;index" json:"isHot"`    // 是否为热门评论
	Status         int            `gorm:"default:1;index" json:"status"`       // 状态: 0-已删除, 1-正常, 2-待审核
}

// SafeComment 安全的评论结构体，用于API返回，不包含用户敏感信息
type SafeComment struct {
	ID             uint          `json:"id"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	PostID         uint          `json:"postId"`
	User           SafeUser      `json:"user"` // 使用安全的用户信息
	Content        string        `json:"content"`
	Images         string        `json:"images"`
	ParentID       *uint         `json:"parentId"`
	RootID         *uint         `json:"rootId"`
	Level          int           `json:"level"`
	Path           string        `json:"path"`
	RepliesCount   int           `json:"repliesCount"`
	LikesCount     int           `json:"likesCount"`
	MentionedUsers string        `json:"mentionedUsers"`
	IsHot          bool          `json:"isHot"`
	Status         int           `json:"status"`
	IsLiked        bool          `json:"isLiked"`               // 当前用户是否已点赞
	Replies        []SafeComment `json:"replies,omitempty"`     // 子评论列表
	ReplyToUser    *SafeUser     `json:"replyToUser,omitempty"` // 回复的用户信息
}

// ToSafeComment 将Comment转换为SafeComment
func (c *Comment) ToSafeComment() SafeComment {
	return c.ToSafeCommentWithUser(0) // 默认不传用户ID
}

// ToSafeCommentWithUser 将Comment转换为SafeComment，包含用户点赞状态
func (c *Comment) ToSafeCommentWithUser(userID uint) SafeComment {
	safeComment := SafeComment{
		ID:             c.ID,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		PostID:         c.PostID,
		User:           c.User.ToSafeUser(), // 转换为安全的用户信息
		Content:        c.Content,
		Images:         c.Images,
		ParentID:       c.ParentID,
		RootID:         c.RootID,
		Level:          c.Level,
		Path:           c.Path,
		RepliesCount:   c.RepliesCount,
		LikesCount:     c.LikesCount,
		MentionedUsers: c.MentionedUsers,
		IsHot:          c.IsHot,
		Status:         c.Status,
		IsLiked:        false, // 默认未点赞
	}

	// 如果传入了用户ID，检查点赞状态
	if userID > 0 {
		safeComment.IsLiked = IsCommentLiked(c.ID, userID)
	}

	// 如果有父评论且已预加载，获取回复的用户信息
	if c.Parent != nil && c.Parent.User.ID > 0 {
		parentUser := c.Parent.User.ToSafeUser()
		safeComment.ReplyToUser = &parentUser
	}

	return safeComment
}

// CreatePost 创建帖子
func CreatePost(post *Post) error {
	// 开始事务
	tx := DB.Begin()

	// 创建帖子
	if err := tx.Create(post).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果是社团帖子，更新社团帖子数量
	if post.ClubID != nil && *post.ClubID > 0 {
		if err := tx.Model(&Club{}).Where("id = ?", *post.ClubID).
			Update("post_count", gorm.Expr("post_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// FindPostByID 根据ID查找帖子
func FindPostByID(id uint) (*Post, error) {
	var post Post
	result := DB.Preload("Author").Preload("Club").First(&post, id)
	if result.Error != nil {
		return nil, result.Error
	}

	// 更新浏览量
	DB.Model(&post).Update("views_count", post.ViewsCount+1)

	return &post, nil
}

// GetPostsList 获取帖子列表
func GetPostsList(page, pageSize int, clubID uint, postType string, isOfficial bool, blockedUserIDs []uint) ([]Post, int64, error) {
	var posts []Post
	var total int64

	query := DB.Model(&Post{}).Where("status = 1")

	// 过滤被屏蔽用户的帖子
	if len(blockedUserIDs) > 0 {
		query = query.Where("author_id NOT IN ?", blockedUserIDs)
	}

	// 筛选条件
	if clubID > 0 {
		query = query.Where("club_id = ?", clubID)
	} else if isOfficial {
		// 官方帖子：club_id 为 NULL
		query = query.Where("club_id IS NULL")
	}
	if postType != "" {
		query = query.Where("type = ?", postType)
	}
	if isOfficial {
		query = query.Where("is_official = true")
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("Author").Preload("Club").
		Order("is_top DESC, published_at DESC").
		Offset(offset).Limit(pageSize).Find(&posts)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return posts, total, nil
}

// UpdatePost 更新帖子
func UpdatePost(post *Post) error {
	return DB.Save(post).Error
}

// DeletePost 删除帖子
func DeletePost(id uint) error {
	// 开始事务
	tx := DB.Begin()

	// 获取帖子信息
	var post Post
	if err := tx.First(&post, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 软删除帖子
	if err := tx.Delete(&Post{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果是社团帖子，更新社团帖子数量
	if post.ClubID != nil && *post.ClubID > 0 {
		if err := tx.Model(&Club{}).Where("id = ?", *post.ClubID).
			Update("post_count", gorm.Expr("post_count - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// LikePost 点赞帖子
func LikePost(postID, userID uint) error {
	// 开始事务
	tx := DB.Begin()

	// 检查是否已经点赞
	var existing PostInteraction
	result := tx.Where("post_id = ? AND user_id = ? AND type = 'like'", postID, userID).First(&existing)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 创建点赞记录
		interaction := &PostInteraction{
			PostID: postID,
			UserID: userID,
			Type:   "like",
			Status: 1,
		}
		if err := tx.Create(interaction).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新帖子点赞数（使用条件更新防止负数）
		if err := tx.Model(&Post{}).Where("id = ? AND likes_count >= 0", postID).
			Update("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if existing.Status == 0 {
		// 重新点赞
		if err := tx.Model(&existing).Update("status", 1).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新帖子点赞数（使用条件更新防止负数）
		if err := tx.Model(&Post{}).Where("id = ? AND likes_count >= 0", postID).
			Update("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// 如果已点赞且状态为1，不做任何操作

	return tx.Commit().Error
}

// UnlikePost 取消点赞帖子
func UnlikePost(postID, userID uint) error {
	// 开始事务
	tx := DB.Begin()

	// 更新点赞状态（只更新有效点赞）
	result := tx.Model(&PostInteraction{}).
		Where("post_id = ? AND user_id = ? AND type = 'like' AND status = 1", postID, userID).
		Update("status", 0)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 只有当有实际更新时才减少点赞数
	if result.RowsAffected > 0 {
		// 更新帖子点赞数（使用条件更新防止负数）
		if err := tx.Model(&Post{}).Where("id = ? AND likes_count > 0", postID).
			Update("likes_count", gorm.Expr("likes_count - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// CreateComment 创建评论
func CreateComment(comment *Comment) error {
	// 开始事务
	tx := DB.Begin()

	// 创建评论
	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新帖子评论数
	if err := tx.Model(&Post{}).Where("id = ?", comment.PostID).
		Update("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetCommentsByPostID 获取帖子评论列表
func GetCommentsByPostID(postID uint, page, pageSize int) ([]Comment, int64, error) {
	var comments []Comment
	var total int64

	query := DB.Model(&Comment{}).Where("post_id = ? AND status = 1", postID)

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("User").Preload("Parent").
		Order("created_at ASC").
		Offset(offset).Limit(pageSize).Find(&comments)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return comments, total, nil
}

// DeleteComment 删除评论
func DeleteComment(id uint) error {
	// 开始事务
	tx := DB.Begin()

	// 获取评论信息
	var comment Comment
	if err := tx.First(&comment, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 软删除评论
	if err := tx.Model(&comment).Update("status", 0).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新帖子评论数
	if err := tx.Model(&Post{}).Where("id = ?", comment.PostID).
		Update("comments_count", gorm.Expr("comments_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CreateCommentWithHierarchy 创建带层级关系的评论
func CreateCommentWithHierarchy(comment *Comment) error {
	// 开始事务
	tx := DB.Begin()

	// 如果是回复评论，设置层级信息
	if comment.ParentID != nil && *comment.ParentID > 0 {
		var parentComment Comment
		if err := tx.First(&parentComment, *comment.ParentID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 检查层级深度限制（最多5层）
		if parentComment.Level >= 4 {
			tx.Rollback()
			return fmt.Errorf("评论层级不能超过5层")
		}

		// 设置层级信息
		comment.Level = parentComment.Level + 1
		comment.RootID = parentComment.RootID
		if comment.RootID == nil {
			// 如果父评论没有根ID，说明父评论就是根评论
			comment.RootID = comment.ParentID
		}

		// 设置路径
		if parentComment.Path == "" {
			comment.Path = fmt.Sprintf("%d/%d", *comment.ParentID, comment.ID)
		} else {
			comment.Path = fmt.Sprintf("%s/%d", parentComment.Path, comment.ID)
		}

		// 更新父评论的回复数
		if err := tx.Model(&Comment{}).Where("id = ?", *comment.ParentID).
			Update("replies_count", gorm.Expr("replies_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 顶级评论
		comment.Level = 0
		comment.RootID = nil
		comment.Path = ""
	}

	// 创建评论
	if err := tx.Create(comment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果是顶级评论，更新路径
	if comment.Level == 0 {
		comment.Path = fmt.Sprintf("%d", comment.ID)
		if err := tx.Model(comment).Update("path", comment.Path).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新帖子评论数
	if err := tx.Model(&Post{}).Where("id = ?", comment.PostID).
		Update("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetCommentReplies 获取评论的回复列表
func GetCommentReplies(commentID uint, page, pageSize int, sortBy string) ([]Comment, int64, error) {
	var comments []Comment
	var total int64

	query := DB.Model(&Comment{}).Where("parent_id = ? AND status = 1", commentID)

	// 计算总数
	query.Count(&total)

	// 排序
	orderBy := "created_at ASC" // 默认按时间升序
	switch sortBy {
	case "time_desc":
		orderBy = "created_at DESC"
	case "likes":
		orderBy = "likes_count DESC, created_at ASC"
	case "hot":
		orderBy = "is_hot DESC, likes_count DESC, created_at ASC"
	}

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("User").Preload("Parent.User").
		Order(orderBy).
		Offset(offset).Limit(pageSize).Find(&comments)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return comments, total, nil
}

// GetCommentTree 获取评论树结构
func GetCommentTree(postID uint, page, pageSize int, sortBy string) ([]Comment, int64, error) {
	var rootComments []Comment
	var total int64

	// 只获取顶级评论
	query := DB.Model(&Comment{}).Where("post_id = ? AND parent_id IS NULL AND status = 1", postID)

	// 计算总数
	query.Count(&total)

	// 排序
	orderBy := "created_at ASC" // 默认按时间升序
	switch sortBy {
	case "time_desc":
		orderBy = "created_at DESC"
	case "likes":
		orderBy = "likes_count DESC, created_at ASC"
	case "hot":
		orderBy = "is_hot DESC, likes_count DESC, created_at ASC"
	}

	// 分页查询顶级评论
	offset := (page - 1) * pageSize
	result := query.Preload("User").
		Order(orderBy).
		Offset(offset).Limit(pageSize).Find(&rootComments)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return rootComments, total, nil
}

// GetCommentsByRootID 获取指定根评论下的所有子评论
func GetCommentsByRootID(rootID uint, sortBy string) ([]Comment, error) {
	var comments []Comment

	query := DB.Model(&Comment{}).Where("root_id = ? AND status = 1", rootID)

	// 排序
	orderBy := "path ASC, created_at ASC" // 默认按路径和时间排序
	switch sortBy {
	case "time_desc":
		orderBy = "path ASC, created_at DESC"
	case "likes":
		orderBy = "path ASC, likes_count DESC, created_at ASC"
	}

	result := query.Preload("User").Preload("Parent.User").
		Order(orderBy).Find(&comments)

	if result.Error != nil {
		return nil, result.Error
	}

	return comments, nil
}

// UpdateCommentHotStatus 更新评论热门状态
func UpdateCommentHotStatus(commentID uint, isHot bool) error {
	return DB.Model(&Comment{}).Where("id = ?", commentID).
		Update("is_hot", isHot).Error
}

// GetHotComments 获取热门评论
func GetHotComments(postID uint, limit int) ([]Comment, error) {
	var comments []Comment

	result := DB.Model(&Comment{}).
		Where("post_id = ? AND is_hot = true AND status = 1", postID).
		Preload("User").
		Order("likes_count DESC, created_at DESC").
		Limit(limit).Find(&comments)

	if result.Error != nil {
		return nil, result.Error
	}

	return comments, nil
}

// LikeComment 点赞评论
func LikeComment(commentID, userID uint) error {
	// 检查是否已经点赞
	var existing CommentInteraction
	result := DB.Where("comment_id = ? AND user_id = ? AND type = 'like'", commentID, userID).First(&existing)

	// 开始事务
	tx := DB.Begin()

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 创建点赞记录
		interaction := &CommentInteraction{
			CommentID: commentID,
			UserID:    userID,
			Type:      "like",
			Status:    1,
		}
		if err := tx.Create(interaction).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新评论点赞数（使用条件更新防止负数）
		if err := tx.Model(&Comment{}).Where("id = ? AND likes_count >= 0", commentID).
			Update("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if existing.Status == 0 {
		// 重新点赞
		if err := tx.Model(&existing).Update("status", 1).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 更新评论点赞数（使用条件更新防止负数）
		if err := tx.Model(&Comment{}).Where("id = ? AND likes_count >= 0", commentID).
			Update("likes_count", gorm.Expr("likes_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// 如果已点赞且状态为1，不做任何操作

	return tx.Commit().Error
}

// UnlikeComment 取消点赞评论
func UnlikeComment(commentID, userID uint) error {
	// 开始事务
	tx := DB.Begin()

	// 更新点赞状态（只更新有效点赞）
	result := tx.Model(&CommentInteraction{}).
		Where("comment_id = ? AND user_id = ? AND type = 'like' AND status = 1", commentID, userID).
		Update("status", 0)

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	// 只有当有实际更新时才减少点赞数
	if result.RowsAffected > 0 {
		// 更新评论点赞数（使用条件更新防止负数）
		if err := tx.Model(&Comment{}).Where("id = ? AND likes_count > 0", commentID).
			Update("likes_count", gorm.Expr("likes_count - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// IsCommentLiked 检查用户是否已点赞评论
func IsCommentLiked(commentID, userID uint) bool {
	var interaction CommentInteraction
	result := DB.Where("comment_id = ? AND user_id = ? AND type = 'like' AND status = 1", commentID, userID).First(&interaction)
	return result.Error == nil
}

// DeleteCommentWithHierarchy 删除评论（保持层级关系）
func DeleteCommentWithHierarchy(id uint) error {
	// 开始事务
	tx := DB.Begin()

	// 获取评论信息
	var comment Comment
	if err := tx.First(&comment, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 软删除评论（设置状态为0，保持回复链完整性）
	if err := tx.Model(&comment).Update("status", 0).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果有父评论，更新父评论的回复数
	if comment.ParentID != nil && *comment.ParentID > 0 {
		if err := tx.Model(&Comment{}).Where("id = ?", *comment.ParentID).
			Update("replies_count", gorm.Expr("replies_count - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新帖子评论数
	if err := tx.Model(&Post{}).Where("id = ?", comment.PostID).
		Update("comments_count", gorm.Expr("comments_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
