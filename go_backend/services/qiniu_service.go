package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/xiaohu/pingjiao/config"
)

// QiniuService 七牛云存储服务
type QiniuService struct {
	accessKey string
	secretKey string
	bucket    string
	domain    string
	mac       *qbox.Mac
	cfg       *storage.Config
}

// QiniuUploadResult 七牛云上传结果
type QiniuUploadResult struct {
	Key      string `json:"key"`      // 文件名
	Hash     string `json:"hash"`     // 文件hash
	URL      string `json:"url"`      // 访问URL
	Size     int64  `json:"size"`     // 文件大小
	MimeType string `json:"mimeType"` // 文件类型
}

// NewQiniuService 创建七牛云服务实例
func NewQiniuService() *QiniuService {
	accessKey := config.GetQiniuAccessKey()
	secretKey := config.GetQiniuSecretKey()
	bucket := config.GetQiniuBucket()
	domain := config.GetQiniuDomain()

	if accessKey == "" || secretKey == "" || bucket == "" {
		log.Println("七牛云配置不完整，将使用本地存储")
		return nil
	}

	mac := qbox.NewMac(accessKey, secretKey)

	// 根据配置选择存储区域
	var zone *storage.Zone
	zoneConfig := config.GetQiniuZone()
	switch zoneConfig {
	case "huadong":
		zone = &storage.ZoneHuadong
	case "huabei":
		zone = &storage.ZoneHuabei
	case "huanan":
		zone = &storage.ZoneHuanan
	case "beimei":
		zone = &storage.ZoneBeimei
	case "dongnan_ya":
		zone = &storage.ZoneXinjiapo
	default:
		// 如果配置不正确，使用自动检测
		zone = nil
	}

	cfg := &storage.Config{
		Zone:          zone,
		UseHTTPS:      config.GetQiniuUseHTTPS(),
		UseCdnDomains: config.GetQiniuUseCDN(),
	}

	return &QiniuService{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		domain:    domain,
		mac:       mac,
		cfg:       cfg,
	}
}

// IsEnabled 检查七牛云服务是否启用
func (qs *QiniuService) IsEnabled() bool {
	return qs != nil && qs.accessKey != "" && qs.secretKey != "" && qs.bucket != ""
}

// UploadFile 上传文件到七牛云
func (qs *QiniuService) UploadFile(file multipart.File, header *multipart.FileHeader, userID uint) (*QiniuUploadResult, error) {
	if !qs.IsEnabled() {
		return nil, fmt.Errorf("七牛云服务未启用")
	}

	// 生成唯一文件名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	timestamp := time.Now().Unix()
	key := fmt.Sprintf("community/images/%d_%d%s", userID, timestamp, ext)

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 重置文件指针
	//nolint:errcheck
	_, _ = file.Seek(0, 0)

	// 生成上传凭证
	putPolicy := storage.PutPolicy{
		Scope: qs.bucket,
	}
	upToken := putPolicy.UploadToken(qs.mac)

	// 创建表单上传器
	formUploader := storage.NewFormUploader(qs.cfg)
	ret := storage.PutRet{}

	// 上传文件
	err = formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(fileBytes), header.Size, nil)
	if err != nil {
		return nil, fmt.Errorf("上传到七牛云失败: %w", err)
	}

	// 构造访问URL
	var fileURL string
	protocol := "http"
	if qs.cfg.UseHTTPS {
		protocol = "https"
	}

	if qs.domain != "" {
		// 使用自定义域名
		if strings.HasPrefix(qs.domain, "http://") || strings.HasPrefix(qs.domain, "https://") {
			fileURL = fmt.Sprintf("%s/%s", qs.domain, ret.Key)
		} else {
			fileURL = fmt.Sprintf("%s://%s/%s", protocol, qs.domain, ret.Key)
		}
	} else {
		// 使用默认域名
		fileURL = fmt.Sprintf("%s://%s.qiniucdn.com/%s", protocol, qs.bucket, ret.Key)
	}

	return &QiniuUploadResult{
		Key:      ret.Key,
		Hash:     ret.Hash,
		URL:      fileURL,
		Size:     header.Size,
		MimeType: header.Header.Get("Content-Type"),
	}, nil
}

// UploadMultipleFiles 批量上传文件到七牛云
func (qs *QiniuService) UploadMultipleFiles(files []*multipart.FileHeader, userID uint) ([]*QiniuUploadResult, []string) {
	var results []*QiniuUploadResult
	var errors []string

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("打开文件 %s 失败: %v", header.Filename, err))
			continue
		}

		//nolint:errcheck
		result, err := qs.UploadFile(file, header, userID)
		//nolint:errcheck
		_ = file.Close()

		if err != nil {
			errors = append(errors, fmt.Sprintf("上传文件 %s 失败: %v", header.Filename, err))
			continue
		}

		results = append(results, result)
	}

	return results, errors
}

// DeleteFile 删除七牛云文件
func (qs *QiniuService) DeleteFile(key string) error {
	if !qs.IsEnabled() {
		return fmt.Errorf("七牛云服务未启用")
	}

	bucketManager := storage.NewBucketManager(qs.mac, qs.cfg)
	err := bucketManager.Delete(qs.bucket, key)
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFileInfo 获取文件信息
func (qs *QiniuService) GetFileInfo(key string) (*storage.FileInfo, error) {
	if !qs.IsEnabled() {
		return nil, fmt.Errorf("七牛云服务未启用")
	}

	bucketManager := storage.NewBucketManager(qs.mac, qs.cfg)
	fileInfo, err := bucketManager.Stat(qs.bucket, key)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	return &fileInfo, nil
}

// GenerateUploadToken 生成上传凭证（用于前端直传）
func (qs *QiniuService) GenerateUploadToken(keyPrefix string) (string, error) {
	if !qs.IsEnabled() {
		return "", fmt.Errorf("七牛云服务未启用")
	}

	putPolicy := storage.PutPolicy{
		Scope:      qs.bucket,        // 修改为只指定bucket，不限制key前缀
		Expires:    3600,             // 1小时有效期
		InsertOnly: 1,                // 只允许新增，不允许覆盖
		FsizeLimit: 10 * 1024 * 1024, // 10MB文件大小限制
		MimeLimit:  "image/*",        // 只允许图片类型
	}

	return putPolicy.UploadToken(qs.mac), nil
}

// ValidateFileType 验证文件类型
func (qs *QiniuService) ValidateFileType(filename string) bool {
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			return true
		}
	}

	return false
}

// ValidateFileSize 验证文件大小
func (qs *QiniuService) ValidateFileSize(size int64) bool {
	maxSize := int64(10 * 1024 * 1024) // 10MB
	return size <= maxSize
}

// GetUploadStats 获取上传统计信息
func (qs *QiniuService) GetUploadStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":      qs.IsEnabled(),
		"bucket":       qs.bucket,
		"domain":       qs.domain,
		"maxSize":      "10MB",
		"allowedTypes": []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"},
	}

	if qs.IsEnabled() {
		stats["status"] = "active"
	} else {
		stats["status"] = "disabled"
		stats["reason"] = "配置不完整"
	}

	return stats
}

// 全局七牛云服务实例
var globalQiniuService *QiniuService

// InitQiniuService 初始化全局七牛云服务
func InitQiniuService() {
	globalQiniuService = NewQiniuService()
	if globalQiniuService != nil && globalQiniuService.IsEnabled() {
		log.Println("七牛云存储服务初始化成功")
	} else {
		log.Println("七牛云存储服务未启用，将使用本地存储")
	}
}

// GetQiniuService 获取全局七牛云服务实例
func GetQiniuService() *QiniuService {
	if globalQiniuService == nil {
		InitQiniuService()
	}
	return globalQiniuService
}
