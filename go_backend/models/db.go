package models

import (
	"log"
	"os"
	"path/filepath"

	"github.com/xiaohu/pingjiao/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	dbType := config.GetDBType()

	// 配置GORM日志
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	var db *gorm.DB
	var err error

	if dbType == "mysql" {
		// 使用MySQL数据库
		dsn := config.GetMySQLDSN()
		log.Printf("连接MySQL数据库: %s", dsn)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
		})
	} else {
		// 使用SQLite数据库（默认）
		dbPath := config.GetDBPath()

		// 确保数据目录存在
		dbDir := filepath.Dir(dbPath)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			_ = os.MkdirAll(dbDir, 0755)
		}

		log.Printf("连接SQLite数据库: %s", dbPath)
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: gormLogger,
		})
	}

	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	// 设置全局DB变量
	DB = db

	// 自动迁移数据库模型
	migrateModels()

	// 初始化默认数据
	initDefaultData()

	log.Printf("数据库初始化完成，使用 %s 数据库", dbType)
}

// migrateModels 自动迁移数据库模型
func migrateModels() {
	_ = DB.AutoMigrate(
		&User{},
		&School{},
		&News{},
		&Course{},
		&Semester{},
		&Evaluation{},
		&SyncSetting{},
		&SyncLog{},
		&UserSetting{},
		&Club{},
		&ClubMember{},
		&Post{},
		&PostInteraction{},
		&Comment{},
		&CommentInteraction{}, // 新增评论互动表
		&GlobalCache{},        // 新增全局缓存表
		&ClassScheduleCache{},
		&NewsCache{},
		&PersonalScoreCache{},
		&CurrentTimeCache{},
		&CourseTimetableWeekCache{},
		// 新增社区功能模型
		&Bookmark{},
		&Notification{},
		&Report{},
		&UserBlock{},
		&UserFollow{},
		&ModerationSetting{},
		// 新增多渠道通知功能模型
		&NotificationChannel{},
		&SMSBalance{},
		&SMSTransaction{},
		&NotificationLog{},
		&ScoreSnapshot{},
		// 新增智能缓存系统模型
	)
}

// initDefaultData 初始化默认数据
func initDefaultData() {
	// 初始化默认学校
	initDefaultSchool()

	// 初始化默认学期
	initDefaultSemesters()

	// 初始化默认新闻
	initDefaultNews()

	// 初始化默认社区数据
	initDefaultCommunityData()
}
