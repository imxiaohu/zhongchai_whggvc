package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Club 社团/俱乐部模型
type Club struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Name          string         `gorm:"size:100;not null" json:"name"` // 社团名称
	Description   string         `gorm:"type:text" json:"description"`  // 社团描述
	LogoURL       string         `gorm:"size:255" json:"logoUrl"`       // 社团logo
	CreatorID     uint           `gorm:"not null" json:"creatorId"`     // 创建者ID
	Creator       User           `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	SchoolID      uint           `json:"schoolId"` // 学校ID
	School        School         `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	MemberCount   int            `gorm:"default:1" json:"memberCount"`    // 成员数量
	PostCount     int            `gorm:"default:0" json:"postCount"`      // 帖子数量
	Status        int            `gorm:"default:1" json:"status"`         // 状态: 0-禁用, 1-正常, 2-待审核
	IsOfficial    bool           `gorm:"default:false" json:"isOfficial"` // 是否官方社团
	Tags          string         `gorm:"size:255" json:"tags"`            // 标签，逗号分隔
	ContactInfo   string         `gorm:"size:255" json:"contactInfo"`     // 联系方式
	EstablishedAt time.Time      `json:"establishedAt"`                   // 成立时间
}

// SafeClub 安全的社团结构体，用于API返回，不包含创建者敏感信息
type SafeClub struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	LogoURL       string    `json:"logoUrl"`
	Creator       SafeUser  `json:"creator"` // 使用安全的用户信息
	School        School    `json:"school"`
	MemberCount   int       `json:"memberCount"`
	PostCount     int       `json:"postCount"`
	Status        int       `json:"status"`
	IsOfficial    bool      `json:"isOfficial"`
	Tags          string    `json:"tags"`
	ContactInfo   string    `json:"contactInfo"`
	EstablishedAt time.Time `json:"establishedAt"`
	IsMember      bool      `json:"isMember"`
}

func sanitizeBackticks(s string) string {
	if s == "" {
		return ""
	}
	out := strings.TrimSpace(s)
	out = strings.ReplaceAll(out, "\\`", "")
	out = strings.ReplaceAll(out, "`", "")
	return strings.TrimSpace(out)
}

func sanitizeSchoolFields(s School) School {
	s.Logo = sanitizeBackticks(s.Logo)
	s.Website = sanitizeBackticks(s.Website)
	s.ApiBaseUrl = sanitizeBackticks(s.ApiBaseUrl)
	return s
}

// ToSafeClub 将Club转换为SafeClub，isMember 表示当前用户是否已加入该社团
func (c *Club) ToSafeClub(isMember bool) SafeClub {
	return SafeClub{
		ID:            c.ID,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
		Name:          c.Name,
		Description:   c.Description,
		LogoURL:       sanitizeBackticks(c.LogoURL),
		Creator:       c.Creator.ToSafeUser(), // 转换为安全的用户信息
		School:        sanitizeSchoolFields(c.School),
		MemberCount:   c.MemberCount,
		PostCount:     c.PostCount,
		Status:        c.Status,
		IsOfficial:    c.IsOfficial,
		Tags:          c.Tags,
		ContactInfo:   c.ContactInfo,
		EstablishedAt: c.EstablishedAt,
		IsMember:      isMember,
	}
}

// ClubMember 社团成员模型
type ClubMember struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	ClubID    uint           `gorm:"not null" json:"clubId"` // 社团ID
	Club      Club           `gorm:"foreignKey:ClubID" json:"club,omitempty"`
	UserID    uint           `gorm:"not null" json:"userId"` // 用户ID
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role      string         `gorm:"size:20;default:'member'" json:"role"` // 角色: admin-管理员, member-普通成员
	Status    int            `gorm:"default:1" json:"status"`              // 状态: 0-已退出, 1-正常, 2-待审核
	JoinedAt  time.Time      `gorm:"autoCreateTime" json:"joinedAt"`       // 加入时间
}

// SafeClubMember 安全的社团成员结构体，用于API返回，不包含用户敏感信息
type SafeClubMember struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ClubID    uint      `json:"clubId"`
	User      SafeUser  `json:"user"` // 使用安全的用户信息
	Role      string    `json:"role"`
	Status    int       `json:"status"`
	JoinedAt  time.Time `json:"joinedAt"`
}

// ToSafeClubMember 将ClubMember转换为SafeClubMember
func (cm *ClubMember) ToSafeClubMember() SafeClubMember {
	return SafeClubMember{
		ID:        cm.ID,
		CreatedAt: cm.CreatedAt,
		UpdatedAt: cm.UpdatedAt,
		ClubID:    cm.ClubID,
		User:      cm.User.ToSafeUser(), // 转换为安全的用户信息
		Role:      cm.Role,
		Status:    cm.Status,
		JoinedAt:  cm.JoinedAt,
	}
}

// CreateClub 创建社团
func CreateClub(club *Club) error {
	return DB.Create(club).Error
}

// FindClubByID 根据ID查找社团
func FindClubByID(id uint) (*Club, error) {
	var club Club
	result := DB.Preload("Creator").Preload("School").First(&club, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &club, nil
}

// GetClubsList 获取社团列表
func GetClubsList(page, pageSize int, schoolID uint) ([]Club, int64, error) {
	var clubs []Club
	var total int64

	query := DB.Model(&Club{}).Where("status = 1")
	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("Creator").Preload("School").
		Order("is_official DESC, member_count DESC, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&clubs)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return clubs, total, nil
}

// UpdateClub 更新社团信息
func UpdateClub(club *Club) error {
	return DB.Save(club).Error
}

// DeleteClub 删除社团
func DeleteClub(id uint) error {
	return DB.Delete(&Club{}, id).Error
}

// JoinClub 加入社团
func JoinClub(clubID, userID uint) error {
	member := &ClubMember{
		ClubID: clubID,
		UserID: userID,
		Role:   "member",
		Status: 1,
	}

	// 开始事务
	tx := DB.Begin()

	// 创建成员记录
	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新社团成员数量
	if err := tx.Model(&Club{}).Where("id = ?", clubID).
		Update("member_count", gorm.Expr("member_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// LeaveClub 退出社团
func LeaveClub(clubID, userID uint) error {
	// 开始事务
	tx := DB.Begin()

	// 软删除成员记录
	if err := tx.Where("club_id = ? AND user_id = ?", clubID, userID).
		Delete(&ClubMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新社团成员数量
	if err := tx.Model(&Club{}).Where("id = ?", clubID).
		Update("member_count", gorm.Expr("member_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// GetClubMembers 获取社团成员列表
func GetClubMembers(clubID uint, page, pageSize int) ([]ClubMember, int64, error) {
	var members []ClubMember
	var total int64

	query := DB.Model(&ClubMember{}).Where("club_id = ? AND status = 1", clubID)

	// 计算总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Preload("User").Order("role DESC, joined_at ASC").
		Offset(offset).Limit(pageSize).Find(&members)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return members, total, nil
}

// CheckClubMembership 检查用户是否为社团成员
func CheckClubMembership(clubID, userID uint) (*ClubMember, error) {
	var member ClubMember
	result := DB.Where("club_id = ? AND user_id = ? AND status = 1", clubID, userID).First(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return &member, nil
}

// IsClubAdmin 检查用户是否为社团管理员
func IsClubAdmin(clubID, userID uint) bool {
	var count int64
	DB.Model(&ClubMember{}).Where("club_id = ? AND user_id = ? AND role = 'admin' AND status = 1",
		clubID, userID).Count(&count)
	return count > 0
}

// GetUserClubs 获取用户加入的社团列表
func GetUserClubs(userID uint) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 查询用户加入的社团，包含角色信息
	query := `
		SELECT
			c.id, c.name, c.description, c.logo_url, c.member_count, c.post_count,
			c.tags, c.contact_info, c.created_at, c.updated_at,
			cm.role, cm.created_at as joined_at
		FROM clubs c
		INNER JOIN club_members cm ON c.id = cm.club_id
		WHERE cm.user_id = ? AND cm.status = 1 AND c.status = 1
		ORDER BY cm.role DESC, cm.created_at DESC
	`

	rows, err := DB.Raw(query, userID).Rows()
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer rows.Close()

	for rows.Next() {
		club := make(map[string]interface{})
		var id, memberCount, postCount int
		var name, description, logoUrl, tags, contactInfo, role string
		var createdAt, updatedAt, joinedAt time.Time

		err := rows.Scan(
			&id, &name, &description, &logoUrl, &memberCount, &postCount,
			&tags, &contactInfo, &createdAt, &updatedAt,
			&role, &joinedAt,
		)
		if err != nil {
			return nil, err
		}

		club["id"] = id
		club["name"] = name
		club["description"] = description
		club["logoUrl"] = logoUrl
		club["memberCount"] = memberCount
		club["postCount"] = postCount
		club["tags"] = tags
		club["contactInfo"] = contactInfo
		club["createdAt"] = createdAt
		club["updatedAt"] = updatedAt
		club["role"] = role
		club["joinedAt"] = joinedAt

		results = append(results, club)
	}

	return results, nil
}

// AddMember 添加社团成员（别名，用于兼容性）
func AddMember(clubID, userID uint, role string) error {
	member := &ClubMember{
		ClubID: clubID,
		UserID: userID,
		Role:   role,
		Status: 1,
	}

	tx := DB.Begin()

	if err := tx.Create(member).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&Club{}).Where("id = ?", clubID).
		Update("member_count", gorm.Expr("member_count + 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// RemoveMember 移除社团成员（别名，用于兼容性）
func RemoveMember(clubID, userID uint) error {
	tx := DB.Begin()

	if err := tx.Where("club_id = ? AND user_id = ?", clubID, userID).
		Delete(&ClubMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&Club{}).Where("id = ?", clubID).
		Update("member_count", gorm.Expr("member_count - 1")).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateMemberRole 更新成员角色
func UpdateMemberRole(clubID, userID uint, role string) error {
	return DB.Model(&ClubMember{}).
		Where("club_id = ? AND user_id = ?", clubID, userID).
		Update("role", role).Error
}

// GetMemberRole 获取成员角色
func GetMemberRole(clubID, userID uint) (string, error) {
	var member ClubMember
	result := DB.Where("club_id = ? AND user_id = ? AND status = 1", clubID, userID).First(&member)
	if result.Error != nil {
		return "", result.Error
	}
	return member.Role, nil
}

// GetUserMembershipMap 批量检查用户对多个社团的加入状态，返回 clubId -> isMember 的映射
func GetUserMembershipMap(userID uint, clubIDs []uint) map[uint]bool {
	result := make(map[uint]bool)
	if userID == 0 || len(clubIDs) == 0 {
		return result
	}

	var clubIDList []uint
	DB.Model(&ClubMember{}).
		Where("user_id = ? AND status = 1", userID).
		Where("club_id IN ?", clubIDs).
		Pluck("club_id", &clubIDList)

	for _, id := range clubIDList {
		result[id] = true
	}
	return result
}

// initDefaultCommunityData 初始化默认社区数据
func initDefaultCommunityData() {
	// 检查是否已存在社团数据
	var clubCount int64
	DB.Model(&Club{}).Count(&clubCount)
	if clubCount > 0 {
		return
	}

	// 获取默认学校ID
	var school School
	DB.First(&school)

	// 创建默认管理员用户（如果不存在）
	var adminUser User
	result := DB.Where("user_type = 'admin'").First(&adminUser)
	if result.Error != nil {
		// 创建默认管理员用户
		adminUser = User{
			Username: "admin",
			Realname: "系统管理员",
			Nickname: "管理员",
			UserType: "admin",
			SchoolID: school.ID,
			Status:   1,
		}
		DB.Create(&adminUser)
	}

	// 创建默认社团
	clubs := []Club{
		{
			Name:          "学生会",
			Description:   "学校学生会是学生自己的群众性组织，是学校联系学生的桥梁和纽带。学生会以全心全意为同学服务为宗旨，发挥学校党政联系广大同学的桥梁和纽带作用。",
			LogoURL:       "/uploads/images/default_club_logo.png",
			CreatorID:     adminUser.ID,
			SchoolID:      school.ID,
			MemberCount:   1,
			PostCount:     0,
			Status:        1,
			IsOfficial:    true,
			Tags:          "学生组织,官方,服务",
			ContactInfo:   "学生处办公室",
			EstablishedAt: time.Now().AddDate(-1, 0, 0), // 一年前成立
		},
		{
			Name:          "计算机协会",
			Description:   "计算机协会致力于为广大计算机爱好者提供学习交流平台，组织各类技术讲座、编程竞赛和项目实践活动。",
			LogoURL:       "/uploads/images/default_club_logo.png",
			CreatorID:     adminUser.ID,
			SchoolID:      school.ID,
			MemberCount:   1,
			PostCount:     0,
			Status:        1,
			IsOfficial:    false,
			Tags:          "技术,编程,计算机",
			ContactInfo:   "信息技术楼201",
			EstablishedAt: time.Now().AddDate(0, -6, 0), // 半年前成立
		},
		{
			Name:          "文学社",
			Description:   "文学社是热爱文学创作的同学们的聚集地，我们定期举办诗歌朗诵、文学创作比赛等活动，为文学爱好者提供展示才华的舞台。",
			LogoURL:       "/uploads/images/default_club_logo.png",
			CreatorID:     adminUser.ID,
			SchoolID:      school.ID,
			MemberCount:   1,
			PostCount:     0,
			Status:        1,
			IsOfficial:    false,
			Tags:          "文学,创作,艺术",
			ContactInfo:   "图书馆三楼文学社办公室",
			EstablishedAt: time.Now().AddDate(0, -3, 0), // 三个月前成立
		},
	}

	for _, club := range clubs {
		DB.Create(&club)

		// 创建者自动成为管理员成员
		member := ClubMember{
			ClubID: club.ID,
			UserID: adminUser.ID,
			Role:   "admin",
			Status: 1,
		}
		DB.Create(&member)
	}

	// 创建一些默认帖子
	var studentClub Club
	DB.Where("name = ?", "学生会").First(&studentClub)

	posts := []Post{
		{
			Title:       "欢迎新同学加入我们的大家庭！",
			Content:     "<p>亲爱的同学们，欢迎大家来到我们学校！</p><p>学生会是一个充满活力的组织，我们致力于为同学们提供更好的校园生活体验。</p><p>如果你有任何建议或想法，欢迎随时联系我们！</p>",
			Summary:     "学生会欢迎新同学的加入，介绍学生会的作用和联系方式。",
			Type:        "announcement",
			ClubID:      &studentClub.ID,
			AuthorID:    adminUser.ID,
			IsOfficial:  true,
			IsTop:       true,
			Status:      1,
			PublishedAt: time.Now().AddDate(0, 0, -7), // 一周前发布
		},
		{
			Title:       "关于举办校园文化节的通知",
			Content:     "<p>为了丰富同学们的校园文化生活，学校决定举办第五届校园文化节。</p><p><strong>活动时间：</strong>2024年7月15日-7月20日</p><p><strong>活动地点：</strong>学校大礼堂及各教学楼</p><p><strong>参与方式：</strong>请各社团积极报名参加</p>",
			Summary:     "学校将举办第五届校园文化节，时间为7月15日-20日，欢迎各社团报名参加。",
			Type:        "announcement",
			ClubID:      nil, // 官方帖子，不属于任何社团
			AuthorID:    adminUser.ID,
			IsOfficial:  true,
			IsTop:       false,
			Status:      1,
			PublishedAt: time.Now().AddDate(0, 0, -3), // 三天前发布
		},
	}

	for _, post := range posts {
		DB.Create(&post)

		// 更新社团帖子数量
		if post.ClubID != nil && *post.ClubID > 0 {
			DB.Model(&Club{}).Where("id = ?", *post.ClubID).
				Update("post_count", gorm.Expr("post_count + 1"))
		}
	}
}
