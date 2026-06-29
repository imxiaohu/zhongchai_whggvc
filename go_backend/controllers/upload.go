package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

// UploadImage 上传图片
func UploadImage(c *gin.Context) {
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

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("获取文件失败: "+err.Error(), 400))
		return
	}
	//nolint:errcheck
	defer file.Close()

	// 检查文件大小（限制为10MB）
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("文件大小超过限制（最大10MB）", 400))
		return
	}

	// 检查文件类型
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	isAllowed := false
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("不支持的文件类型，仅支持: jpg, jpeg, png, gif, webp, svg", 400))
		return
	}

	// 尝试使用七牛云上传
	qiniuService := services.GetQiniuService()
	if qiniuService != nil && qiniuService.IsEnabled() {
		// 使用七牛云上传
		result, err := qiniuService.UploadFile(file, header, uint(userIdUint))
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("上传到七牛云失败: "+err.Error(), 500))
			return
		}

		// 返回七牛云上传结果
		response := map[string]interface{}{
			"url":      result.URL,
			"filename": result.Key,
			"size":     result.Size,
			"type":     ext,
			"hash":     result.Hash,
			"storage":  "qiniu",
		}

		c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
		return
	}

	// 七牛云未启用，使用本地存储
	// 创建上传目录
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建上传目录失败: "+err.Error(), 500))
		return
	}

	// 生成唯一文件名
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d%s", userIdUint, timestamp, ext)
	filePath := filepath.Join(uploadDir, filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建文件失败: "+err.Error(), 500))
		return
	}
	//nolint:errcheck
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("保存文件失败: "+err.Error(), 500))
		return
	}

	// 构造文件URL
	fileURL := fmt.Sprintf("/uploads/images/%s", filename)

	// 返回结果
	result := map[string]interface{}{
		"url":      fileURL,
		"filename": filename,
		"size":     header.Size,
		"type":     ext,
		"storage":  "local",
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(result))
}

// UploadMultipleImages 批量上传图片
func UploadMultipleImages(c *gin.Context) {
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

	// 解析多文件表单
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("解析表单失败: "+err.Error(), 400))
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("没有选择文件", 400))
		return
	}

	// 限制批量上传数量
	maxFiles := 9
	if len(files) > maxFiles {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(fmt.Sprintf("最多只能上传%d个文件", maxFiles), 400))
		return
	}

	// 预先验证所有文件
	for _, header := range files {
		// 检查文件大小
		maxSize := int64(10 * 1024 * 1024) // 10MB
		if header.Size > maxSize {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse(fmt.Sprintf("文件 %s 大小超过限制（最大10MB）", header.Filename), 400))
			return
		}

		// 检查文件类型
		allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		ext := strings.ToLower(filepath.Ext(header.Filename))
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if ext == allowedType {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusBadRequest, utils.NewErrorResponse(fmt.Sprintf("文件 %s 类型不支持，仅支持: jpg, jpeg, png, gif, webp, svg", header.Filename), 400))
			return
		}
	}

	// 尝试使用七牛云批量上传
	qiniuService := services.GetQiniuService()
	if qiniuService != nil && qiniuService.IsEnabled() {
		// 使用七牛云批量上传
		results, errors := qiniuService.UploadMultipleFiles(files, uint(userIdUint))

		// 转换结果格式
		var successResults []map[string]interface{}
		for _, result := range results {
			ext := strings.ToLower(filepath.Ext(result.Key))
			successResults = append(successResults, map[string]interface{}{
				"url":      result.URL,
				"filename": result.Key,
				"size":     result.Size,
				"type":     ext,
				"hash":     result.Hash,
				"storage":  "qiniu",
			})
		}

		// 构造响应
		response := map[string]interface{}{
			"success":  successResults,
			"errors":   errors,
			"total":    len(files),
			"uploaded": len(successResults),
			"failed":   len(errors),
		}

		c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
		return
	}

	// 七牛云未启用，使用本地存储
	// 创建上传目录
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("创建上传目录失败: "+err.Error(), 500))
		return
	}

	var results []map[string]interface{}
	var errors []string

	// 处理每个文件
	for i, header := range files {

		// 打开文件
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("打开文件 %s 失败: %v", header.Filename, err))
			continue
		}

		// 生成唯一文件名
		timestamp := time.Now().Unix()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		filename := fmt.Sprintf("%d_%d_%d%s", userIdUint, timestamp, i, ext)
		filePath := filepath.Join(uploadDir, filename)

		// 创建目标文件
		dst, err := os.Create(filePath)
		if err != nil {
			//nolint:errcheck
			file.Close()
			errors = append(errors, fmt.Sprintf("创建文件 %s 失败: %v", header.Filename, err))
			continue
		}

		// 复制文件内容
		if _, err := io.Copy(dst, file); err != nil {
			//nolint:errcheck
			file.Close()
			//nolint:errcheck
			dst.Close()
			errors = append(errors, fmt.Sprintf("保存文件 %s 失败: %v", header.Filename, err))
			continue
		}

		//nolint:errcheck
		file.Close()
		//nolint:errcheck
		dst.Close()

		// 构造文件URL
		fileURL := fmt.Sprintf("/uploads/images/%s", filename)

		// 添加到结果列表
		results = append(results, map[string]interface{}{
			"url":          fileURL,
			"filename":     filename,
			"originalName": header.Filename,
			"size":         header.Size,
			"type":         ext,
		})
	}

	// 构造响应
	response := map[string]interface{}{
		"success":  results,
		"errors":   errors,
		"total":    len(files),
		"uploaded": len(results),
		"failed":   len(errors),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
}

// ServeUploadedFile 提供上传文件的静态服务
func ServeUploadedFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("文件名不能为空", 400))
		return
	}

	// 构造文件路径
	filePath := filepath.Join("./uploads/images", filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("文件不存在", 404))
		return
	}

	// 设置适当的Content-Type
	ext := strings.ToLower(filepath.Ext(filename))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".svg":
		contentType = "image/svg+xml"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // 缓存1年
	c.File(filePath)
}

// GetUploadToken 获取七牛云上传凭证（用于前端直传）
func GetUploadToken(c *gin.Context) {
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

	// 获取七牛云服务
	qiniuService := services.GetQiniuService()
	if qiniuService == nil || !qiniuService.IsEnabled() {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "七牛云服务未启用")
		return
	}

	// 生成上传凭证
	keyPrefix := fmt.Sprintf("avatars/%d_", userIdUint) // 修改为avatars前缀，支持头像上传
	token, err := qiniuService.GenerateUploadToken(keyPrefix)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "生成上传凭证失败")
		return
	}

	// 获取七牛云配置信息
	stats := qiniuService.GetUploadStats()

	// 返回上传凭证和配置信息
	result := map[string]interface{}{
		"token":        token,
		"keyPrefix":    keyPrefix,
		"bucket":       stats["bucket"],
		"domain":       stats["domain"],
		"uploadUrl":    "https://upload.qiniup.com", // 添加上传URL
		"maxSize":      "10MB",
		"allowedTypes": []string{"jpg", "jpeg", "png", "gif", "webp", "svg"},
	}

	utils.SuccessResponse(c, result)
}

// GetUploadStats 获取上传服务状态
func GetUploadStats(c *gin.Context) {
	qiniuService := services.GetQiniuService()

	var stats map[string]interface{}
	if qiniuService != nil {
		stats = qiniuService.GetUploadStats()
	} else {
		stats = map[string]interface{}{
			"enabled":      false,
			"status":       "disabled",
			"reason":       "服务未初始化",
			"storage":      "local",
			"maxSize":      "10MB",
			"allowedTypes": []string{"jpg", "jpeg", "png", "gif", "webp", "svg"},
		}
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(stats))
}

// DeleteUploadedFile 删除已上传的文件
func DeleteUploadedFile(c *gin.Context) {
	// 获取当前用户ID
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("未授权", 401))
		return
	}

	// 获取文件key参数
	fileKey := c.Query("key")
	if fileKey == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("文件key不能为空", 400))
		return
	}

	// 获取七牛云服务
	qiniuService := services.GetQiniuService()
	if qiniuService == nil || !qiniuService.IsEnabled() {
		// 七牛云未启用，尝试删除本地文件
		filePath := filepath.Join("./uploads/images", filepath.Base(fileKey))
		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("删除本地文件失败: "+err.Error(), 500))
			return
		}
		c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
		return
	}

	// 删除七牛云文件
	if err := qiniuService.DeleteFile(fileKey); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("删除文件失败: "+err.Error(), 500))
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(nil))
}
