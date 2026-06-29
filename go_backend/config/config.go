package config

import (
	"os"
	"strconv"
	"strings"
)

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整数类型的环境变量
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool 获取布尔类型的环境变量
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func GetEnvFirst(keys []string, defaultValue string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return defaultValue
}

func GetEnvBoolFirst(keys []string, defaultValue bool) bool {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			if boolValue, err := strconv.ParseBool(value); err == nil {
				return boolValue
			}
		}
	}
	return defaultValue
}

// GetEnvSlice 获取切片类型的环境变量（逗号分隔）
func GetEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// 配置常量
const (
	// 默认端口
	DefaultPort = "2333"

	// 默认JWT密钥
	DefaultJWTSecret = "please_set_jwt_secret_in_env"

	// 默认JWT过期时间（30天）
	DefaultJWTExpiration = 30 * 24 * 60 * 60

	// 默认数据库文件路径
	DefaultDBPath = "./data/pingjiao.db"

	// MySQL数据库配置
	DefaultMySQLHost     = "localhost"
	DefaultMySQLPort     = "3306"
	DefaultMySQLUser     = "root"
	DefaultMySQLPassword = ""
	DefaultMySQLDatabase = "pingjiao"
	DefaultMySQLCharset  = "utf8mb4"

	// 默认学校服务器地址
	DefaultSchoolServerURL = "https://xs.whggvc.net"

	// 默认请求头
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"

	// 微信小程序配置
	DefaultWechatAppID     = ""
	DefaultWechatAppSecret = ""
)

// 动态配置获取函数
func GetPort() string {
	return GetEnv("PORT", DefaultPort)
}

func GetJWTSecret() string {
	return GetEnv("JWT_SECRET", DefaultJWTSecret)
}

func GetJWTExpiration() int {
	return GetEnvInt("JWT_EXPIRATION", DefaultJWTExpiration)
}

func GetDBPath() string {
	return GetEnv("DB_PATH", DefaultDBPath)
}

// MySQL配置获取函数
func GetDBType() string {
	return GetEnv("DB_TYPE", "sqlite") // 默认使用sqlite，可设置为mysql
}

func GetMySQLHost() string {
	return GetEnv("MYSQL_HOST", DefaultMySQLHost)
}

func GetMySQLPort() string {
	return GetEnv("MYSQL_PORT", DefaultMySQLPort)
}

func GetMySQLUser() string {
	return GetEnv("MYSQL_USER", DefaultMySQLUser)
}

func GetMySQLPassword() string {
	return GetEnv("MYSQL_PASSWORD", DefaultMySQLPassword)
}

func GetMySQLDatabase() string {
	return GetEnv("MYSQL_DATABASE", DefaultMySQLDatabase)
}

func GetMySQLCharset() string {
	return GetEnv("MYSQL_CHARSET", DefaultMySQLCharset)
}

// GetMySQLDSN 获取MySQL连接字符串
func GetMySQLDSN() string {
	return GetMySQLUser() + ":" + GetMySQLPassword() + "@tcp(" + GetMySQLHost() + ":" + GetMySQLPort() + ")/" + GetMySQLDatabase() + "?charset=" + GetMySQLCharset() + "&parseTime=True&loc=Local"
}

func GetSchoolServerURL() string {
	return GetEnv("SCHOOL_SERVER_URL", DefaultSchoolServerURL)
}

func GetAPIBaseURL() string {
	return GetEnv("API_BASE_URL", "http://localhost:2333")
}

func GetCORSAllowOrigins() []string {
	return GetEnvSlice("CORS_ALLOW_ORIGINS", []string{
		"http://localhost:5173",
		"http://localhost:8080",
		"http://localhost:3001",
		"http://127.0.0.1:8080",
		"http://127.0.0.1:3001",
		"http://whggvc.imxiaohu.cn",
		"https://whggvc.imxiaohu.cn",
		"http://api.whggvc.imxiaohu.cn",
		"https://api.whggvc.imxiaohu.cn",
		"https://go.server.zhongchai.imxiaohu.cn",
		"http://go.server.zhongchai.imxiaohu.cn",
		"null",
	})
}

func GetWechatAppID() string {
	return GetEnv("WECHAT_APP_ID", DefaultWechatAppID)
}

func GetWechatAppSecret() string {
	return GetEnv("WECHAT_APP_SECRET", DefaultWechatAppSecret)
}

func GetOfflineCacheSchedulerEnabled() bool {
	return GetEnvBool("OFFLINE_CACHE_SCHEDULER_ENABLED", true)
}

func GetOfflineCacheSchedulerTZ() string {
	return GetEnv("OFFLINE_CACHE_SCHEDULER_TZ", "Local")
}

func GetOfflineCacheClassScheduleMaxWeek() int {
	return GetEnvInt("OFFLINE_CACHE_CLASS_SCHEDULE_MAX_WEEK", 20)
}

func GetSchoolPasswordEncKey() string {
	return GetEnv("SCHOOL_PASSWORD_ENC_KEY", "")
}

func GetOCRServURL() string {
	return GetEnv("OCR_SERVICE_URL", "http://127.0.0.1:8899")
}

// 七牛云配置获取函数
func GetQiniuAccessKey() string {
	return GetEnv("QINIU_ACCESS_KEY", "")
}

func GetQiniuSecretKey() string {
	return GetEnv("QINIU_SECRET_KEY", "")
}

func GetQiniuBucket() string {
	return GetEnv("QINIU_BUCKET", "")
}

func GetQiniuDomain() string {
	return GetEnv("QINIU_DOMAIN", "")
}

func GetQiniuZone() string {
	return GetEnv("QINIU_ZONE", "huadong") // 默认华东区域
}

func GetQiniuUseHTTPS() bool {
	return GetEnvBool("QINIU_USE_HTTPS", true)
}

func GetQiniuUseCDN() bool {
	return GetEnvBool("QINIU_USE_CDN", false)
}

// 阿里云短信配置获取函数
func GetAliyunAccessKeyID() string {
	return GetEnvFirst([]string{"ALIYUN_SMS_ACCESS_KEY_ID", "ALIYUN_ACCESS_KEY_ID"}, "")
}

func GetAliyunAccessKeySecret() string {
	return GetEnvFirst([]string{"ALIYUN_SMS_ACCESS_KEY_SECRET", "ALIYUN_ACCESS_KEY_SECRET"}, "")
}

func GetAliyunSMSRegion() string {
	return GetEnvFirst([]string{"ALIYUN_SMS_REGION", "ALIYUN_REGION"}, "cn-hangzhou")
}

func GetSMSSignName() string {
	return GetEnvFirst([]string{"ALIYUN_SMS_SIGN_NAME", "SMS_SIGN_NAME"}, "学生评教系统")
}

// 缓存配置获取函数
func GetCacheEnabled() bool {
	return GetEnvBool("CACHE_ENABLED", true)
}

func GetCacheType() string {
	return GetEnv("CACHE_TYPE", "redis") // 可选: memory, redis
}

func GetCacheTTL() int {
	return GetEnvInt("CACHE_TTL", 3600) // 默认1小时
}

// Redis配置获取函数
func GetRedisEnabled() bool {
	return GetEnvBool("REDIS_ENABLED", true)
}

func GetRedisHost() string {
	return GetEnv("REDIS_HOST", "localhost")
}

func GetRedisPort() string {
	return GetEnv("REDIS_PORT", "6379")
}

func GetRedisPassword() string {
	return GetEnv("REDIS_PASSWORD", "")
}

func GetRedisDB() int {
	return GetEnvInt("REDIS_DB", 0)
}

func GetRedisPoolSize() int {
	return GetEnvInt("REDIS_POOL_SIZE", 10)
}

func GetRedisTimeout() int {
	return GetEnvInt("REDIS_TIMEOUT", 5) // 默认5秒
}

func GetSMSEnabled() bool {
	return GetEnvBoolFirst([]string{"ALIYUN_SMS_ENABLED", "SMS_ENABLED"}, false)
}

// 响应状态码
const (
	CodeSuccess      = 0
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// 响应消息
const (
	MsgSuccess      = "操作成功"
	MsgUnauthorized = "未授权或登录已过期"
	MsgForbidden    = "禁止访问"
	MsgNotFound     = "资源不存在"
	MsgServerError  = "服务器内部错误"
)
