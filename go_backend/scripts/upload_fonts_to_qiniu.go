package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// 字体文件配置
var fontFiles = []string{
	"materialdesignicons-webfont.eot",
	"materialdesignicons-webfont.ttf",
	"materialdesignicons-webfont.woff",
	"materialdesignicons-webfont.woff2",
}

func main() {
	// 加载环境变量
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
	}

	// 获取七牛云配置
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	bucket := os.Getenv("QINIU_BUCKET")
	domain := os.Getenv("QINIU_DOMAIN")

	if accessKey == "" || secretKey == "" || bucket == "" {
		log.Fatal("七牛云配置不完整，请检查环境变量 QINIU_ACCESS_KEY, QINIU_SECRET_KEY, QINIU_BUCKET")
	}

	fmt.Printf("七牛云配置:\n")
	fmt.Printf("  Bucket: %s\n", bucket)
	fmt.Printf("  Domain: %s\n", domain)
	fmt.Printf("  AccessKey: %s...\n", accessKey[:10])

	// 创建七牛云客户端
	mac := qbox.NewMac(accessKey, secretKey)
	cfg := storage.Config{
		Zone:          &storage.ZoneHuanan, // 修改为华南区域
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	// 字体文件本地路径
	fontDir := "../static/fonts"

	fmt.Printf("\n开始上传字体文件到七牛云...\n")

	for _, fontFile := range fontFiles {
		localPath := filepath.Join(fontDir, fontFile)

		// 检查文件是否存在
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			log.Printf("警告: 字体文件不存在: %s", localPath)
			continue
		}

		// 七牛云存储路径
		qiniuKey := fmt.Sprintf("fonts/%s", fontFile)

		// 上传文件
		err := uploadFile(mac, &cfg, bucket, qiniuKey, localPath)
		if err != nil {
			log.Printf("上传失败 %s: %v", fontFile, err)
		} else {
			fmt.Printf("✅ 上传成功: %s -> %s\n", fontFile, qiniuKey)

			// 构造访问URL
			var fileURL string
			if domain != "" {
				if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
					fileURL = fmt.Sprintf("%s/%s", domain, qiniuKey)
				} else {
					fileURL = fmt.Sprintf("https://%s/%s", domain, qiniuKey)
				}
			} else {
				fileURL = fmt.Sprintf("https://%s.qiniucdn.com/%s", bucket, qiniuKey)
			}
			fmt.Printf("   访问URL: %s\n", fileURL)
		}
	}

	fmt.Printf("\n字体文件上传完成！\n")

	// 输出前端配置建议
	var baseURL string
	if domain != "" {
		if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
			baseURL = fmt.Sprintf("%s/fonts", domain)
		} else {
			baseURL = fmt.Sprintf("https://%s/fonts", domain)
		}
	} else {
		baseURL = fmt.Sprintf("https://%s.qiniucdn.com/fonts", bucket)
	}

	fmt.Printf("\n📝 前端配置建议:\n")
	fmt.Printf("请将以下路径配置到前端字体配置中:\n")
	fmt.Printf("$mdi-font-path: \"%s\";\n", baseURL)
}

// uploadFile 上传文件到七牛云
func uploadFile(mac *qbox.Mac, cfg *storage.Config, bucket, key, localFile string) error {
	// 生成上传凭证
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)

	// 创建表单上传器
	formUploader := storage.NewFormUploader(cfg)
	ret := storage.PutRet{}

	// 上传文件
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, nil)
	if err != nil {
		return err
	}

	return nil
}
