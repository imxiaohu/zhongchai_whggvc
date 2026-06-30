package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Username          string         `gorm:"size:50;uniqueIndex" json:"username"` // 移除not null约束，允许微信用户暂时不设置用户名
	Password          string         `gorm:"size:100" json:"-"`                   // 存储哈希后的密码
	SchoolPasswordEnc string         `gorm:"size:512" json:"-"`
	Realname          string         `gorm:"size:50" json:"realname"`
	Nickname          string         `gorm:"size:50" json:"nickname"`
	Avatar            string         `gorm:"size:255" json:"avatar"`
	Email             string         `gorm:"size:100" json:"email"`
	Phone             string         `gorm:"size:20" json:"phone"`
	Birthday          string         `gorm:"size:20" json:"birthday"`
	Sex               int            `gorm:"default:0" json:"sex"` // 0: 未知, 1: 男, 2: 女
	IdentityCard      string         `gorm:"size:20" json:"identityCard"`
	UserType          string         `gorm:"size:20;not null;default:'student'" json:"userType"` // student, teacher, admin
	SchoolID          uint           `json:"schoolId"`
	School            School         `gorm:"foreignKey:SchoolID" json:"school,omitempty"`
	ClassName         string         `gorm:"size:50" json:"className"`
	ProfessionID      uint           `json:"professionId"`
	FacultyID         uint           `json:"facultyId"`
	GradeID           uint           `json:"gradeId"`
	CurrentSemester   string         `gorm:"size:50" json:"currentSemester"`

	// 教务系统相关字段
	SchoolToken   string    `gorm:"size:500" json:"-"` // 学校系统token
	TokenExpireAt time.Time `json:"tokenExpireAt"`

	// 微信小程序相关字段
	WechatOpenID  string `gorm:"size:100" json:"wechatOpenId"`
	WechatUnionID string `gorm:"size:100" json:"wechatUnionId"`

	// PC端会话凭证（与移动端Token独立）
	// 显式声明 column，避免未来 NamingStrategy 改动后与 savePCSessionToDB 写入列名不一致
	PCJSESSIONID string    `gorm:"size:256;column:pc_jsession_id" json:"-"`
	PCLoginTime  time.Time `gorm:"column:pc_login_time" json:"pcLoginTime"`
	PCExpireTime time.Time `gorm:"column:pc_expire_time" json:"pcExpireTime"`
	PCUserAgent  string    `gorm:"size:512;column:pc_user_agent" json:"-"`

	LastLoginAt time.Time `json:"lastLoginAt"`
	// 活跃度追踪：用户最后活跃时间，用于个人基础信息缓存的暂停/恢复策略
	LastActiveAt time.Time `json:"lastActiveAt"`
	// 个人基础信息缓存状态（仅作便捷读取用，实际以 SyncSetting.PersonalInfoCacheStatus 为准）
	PersonalInfoCacheActive bool `gorm:"-" json:"personalInfoCacheActive"`
	Status                  int  `gorm:"default:1" json:"status"` // 0: 禁用, 1: 正常

	// 社区须知同意状态
	CommunityTermsAgreed   bool      `gorm:"default:false" json:"communityTermsAgreed"`
	CommunityTermsAgreedAt time.Time `json:"communityTermsAgreedAt"`
}

// SetPassword 设置并加密密码
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// FindUserByUsername 根据用户名查找用户
func FindUserByUsername(username string) (*User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByUsernameIncludeDeleted 根据用户名查找用户（包括软删除的用户）
func FindUserByUsernameIncludeDeleted(username string) (*User, error) {
	var user User
	result := DB.Unscoped().Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateUser 创建新用户
func CreateUser(user *User) error {
	result := DB.Create(user)
	return result.Error
}

// UpdateUser 更新用户信息
func UpdateUser(user *User) error {
	result := DB.Save(user)
	return result.Error
}

// UpdateUserFields selectively updates specific fields on a user by ID.
// Use this instead of UpdateUser when you only need to update a subset of fields
// to avoid accidentally overwriting zero-valued time fields with "0000-00-00"
// which MySQL strict mode rejects.
func UpdateUserFields(userID uint, fields map[string]interface{}) error {
	result := DB.Model(&User{}).Where("id = ?", userID).Updates(fields)
	return result.Error
}

// DeleteUser 删除用户
func DeleteUser(id uint) error {
	result := DB.Delete(&User{}, id)
	return result.Error
}

// FindUserByID 根据ID查找用户
func FindUserByID(id uint) (*User, error) {
	var user User
	result := DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateLastActiveAt 更新用户最后活跃时间
// 每次用户调用 API 时调用，用于个人基础信息缓存的活跃度追踪
func UpdateLastActiveAt(userID uint) error {
	return DB.Model(&User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_active_at": time.Now(),
		}).Error
}

// IsUserActiveWithinDays 检查用户最近 N 天是否有活跃记录
func IsUserActiveWithinDays(userID uint, days int) bool {
	var user User
	if err := DB.Select("last_active_at").Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}
	if user.LastActiveAt.IsZero() {
		// 没有活跃记录时，检查 last_login_at 作为兜底
		if err := DB.Select("last_login_at").Where("id = ?", userID).First(&user).Error; err != nil {
			return false
		}
		return !user.LastLoginAt.IsZero() && time.Since(user.LastLoginAt) < time.Duration(days)*24*time.Hour
	}
	return time.Since(user.LastActiveAt) < time.Duration(days)*24*time.Hour
}

// FindUserByWechatOpenID 根据微信OpenID查找用户
func FindUserByWechatOpenID(openID string) (*User, error) {
	var user User
	result := DB.Where("wechat_open_id = ?", openID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// SafeUser 安全的用户信息结构体，用于公开API返回，不包含敏感信息
type SafeUser struct {
	ID       uint   `json:"id"`
	Realname string `json:"realname"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	UserType string `json:"userType"`
}

// ToSafeUser 将User转换为SafeUser
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:       u.ID,
		Realname: u.Realname,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		UserType: u.UserType,
	}
}
