package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/utils"
)

// AppVersion 应用版本信息结构体
type AppVersion struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AppID        string    `json:"appId" gorm:"not null;index"`
	Version      string    `json:"version" gorm:"not null"`
	VersionCode  int       `json:"versionCode" gorm:"not null"`
	Platform     string    `json:"platform" gorm:"not null;index"` // iOS, Android, H5, WeChat
	DownloadURL  string    `json:"downloadUrl"`
	Size         string    `json:"size"`
	ReleaseNotes string    `json:"releaseNotes" gorm:"type:text"`
	IsForced     bool      `json:"isForced" gorm:"default:false"`
	IsActive     bool      `json:"isActive" gorm:"default:true"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CheckUpdateRequest 检查更新请求结构体
type CheckUpdateRequest struct {
	AppID       string `json:"appId" binding:"required"`
	Version     string `json:"version" binding:"required"`
	VersionCode string `json:"versionCode" binding:"required"`
	Platform    string `json:"platform" binding:"required"`
}

// CheckUpdateResponse 检查更新响应结构体
type CheckUpdateResponse struct {
	HasUpdate  bool                  `json:"hasUpdate"`
	UpdateInfo *AppVersionUpdateInfo `json:"updateInfo,omitempty"`
}

// AppVersionUpdateInfo 更新信息结构体
type AppVersionUpdateInfo struct {
	Version      string `json:"version"`
	VersionCode  int    `json:"versionCode"`
	DownloadURL  string `json:"downloadUrl"`
	Size         string `json:"size"`
	ReleaseNotes string `json:"releaseNotes"`
	IsForced     bool   `json:"isForced"`
}

// CheckAppUpdate 检查应用更新
func CheckAppUpdate(c *gin.Context) {
	var req CheckUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	// 转换版本代码为整数
	currentVersionCode, err := strconv.Atoi(req.VersionCode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid version code")
		return
	}

	// 查询最新版本
	var latestVersion AppVersion
	result := models.DB.Where("app_id = ? AND platform = ? AND is_active = ?",
		req.AppID, req.Platform, true).
		Order("version_code DESC").
		First(&latestVersion)

	if result.Error != nil {
		// 如果没有找到版本信息，返回无需更新
		utils.SuccessResponse(c, CheckUpdateResponse{
			HasUpdate: false,
		})
		return
	}

	// 比较版本
	hasUpdate := latestVersion.VersionCode > currentVersionCode

	response := CheckUpdateResponse{
		HasUpdate: hasUpdate,
	}

	if hasUpdate {
		response.UpdateInfo = &AppVersionUpdateInfo{
			Version:      latestVersion.Version,
			VersionCode:  latestVersion.VersionCode,
			DownloadURL:  latestVersion.DownloadURL,
			Size:         latestVersion.Size,
			ReleaseNotes: latestVersion.ReleaseNotes,
			IsForced:     latestVersion.IsForced,
		}
	}

	utils.SuccessResponse(c, response)
}

// GetAppVersions 获取应用版本列表（管理员接口）
func GetAppVersions(c *gin.Context) {
	appID := c.Query("appId")
	platform := c.Query("platform")
	page := 1
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	pageSize := 10
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "10")); err == nil {
		pageSize = ps
	}

	query := models.DB.Model(&AppVersion{})

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	var total int64
	query.Count(&total)

	var versions []AppVersion
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&versions)

	utils.SuccessResponse(c, gin.H{
		"versions": versions,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// CreateAppVersion 创建应用版本（管理员接口）
func CreateAppVersion(c *gin.Context) {
	var version AppVersion
	if err := c.ShouldBindJSON(&version); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	// 检查版本是否已存在
	var existingVersion AppVersion
	result := models.DB.Where("app_id = ? AND platform = ? AND version_code = ?",
		version.AppID, version.Platform, version.VersionCode).First(&existingVersion)

	if result.Error == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Version already exists")
		return
	}

	// 创建版本记录
	if err := models.DB.Create(&version).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create version")
		return
	}

	utils.SuccessResponse(c, version)
}

// UpdateAppVersion 更新应用版本（管理员接口）
func UpdateAppVersion(c *gin.Context) {
	id := c.Param("id")

	var version AppVersion
	if err := models.DB.First(&version, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Version not found")
		return
	}

	var updateData AppVersion
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	// 更新版本信息
	if err := models.DB.Model(&version).Updates(updateData).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update version")
		return
	}

	utils.SuccessResponse(c, version)
}

// DeleteAppVersion 删除应用版本（管理员接口）
func DeleteAppVersion(c *gin.Context) {
	id := c.Param("id")

	var version AppVersion
	if err := models.DB.First(&version, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Version not found")
		return
	}

	if err := models.DB.Delete(&version).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete version")
		return
	}

	utils.SuccessResponse(c, gin.H{"message": "Version deleted successfully"})
}

// GetCurrentAppInfo 获取当前应用信息
func GetCurrentAppInfo(c *gin.Context) {
	// 从manifest.json或配置中获取应用信息
	appInfo := gin.H{
		"appId":       "__UNI__84DD641",
		"name":        "pingjiao",
		"version":     "1.0.0",
		"versionCode": 100,
		"platforms":   []string{"iOS", "Android", "H5", "WeChat"},
	}

	utils.SuccessResponse(c, appInfo)
}

// CompareVersions 比较版本号
func CompareVersions(version1, version2 string) int {
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Part, v2Part int

		if i < len(v1Parts) {
			if p, err := strconv.Atoi(v1Parts[i]); err == nil {
				v1Part = p
			}
		}
		if i < len(v2Parts) {
			if p, err := strconv.Atoi(v2Parts[i]); err == nil {
				v2Part = p
			}
		}

		if v1Part < v2Part {
			return -1
		} else if v1Part > v2Part {
			return 1
		}
	}

	return 0
}

// InitAppVersionTable 初始化应用版本表
func InitAppVersionTable() {
	if err := models.DB.AutoMigrate(&AppVersion{}); err != nil {
		fmt.Printf("Failed to migrate AppVersion table: %v\n", err)
	} else {
		fmt.Println("AppVersion table migrated successfully")
	}
}
