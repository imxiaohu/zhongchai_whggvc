package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/xiaohu/pingjiao/config"
	"github.com/xiaohu/pingjiao/controllers"
	"github.com/xiaohu/pingjiao/middleware"
	"github.com/xiaohu/pingjiao/models"
	"github.com/xiaohu/pingjiao/services"
	"github.com/xiaohu/pingjiao/utils"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("警告: 未找到.env文件，将使用默认环境变量")
	}

	// 检查并自动启动 OCR 服务（不阻塞）
	utils.EnsureOCRService()

	// 初始化数据库
	models.InitDB()

	// 初始化七牛云服务
	services.InitQiniuService()

	// 初始化RabbitMQ服务
	services.InitRabbitMQService()

	// 初始化多渠道通知服务
	services.InitEmailService()
	services.InitDingTalkService()
	services.InitSMSService()
	services.InitWechatPayService()
	services.InitScoreCheckService()
	services.InitOfflineCacheService()
	services.InitOfflineCacheScheduler()

	// 初始化评论相关服务
	services.InitCommentRateLimitService()
	services.InitCommentNotificationService()

	// 初始化多渠道通知服务
	services.InitMultiChannelNotificationService()

	// 启动成绩检查调度器
	go func() {
		scoreCheckService := services.GetScoreCheckService()
		if scoreCheckService != nil {
			scoreCheckService.StartScoreCheckScheduler()
		}
	}()

	// 初始化同步服务
	controllers.InitSyncService()

	// 初始化应用版本表
	controllers.InitAppVersionTable()

	// 设置运行模式
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     config.GetCORSAllowOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Access-Token", "X-Sign", "X-TIMESTAMP", "cache-control", "pragma", "x-client-id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 注册中间件
	r.Use(middleware.Logger())

	// 公共路由
	public := r.Group("/")
	{
		// 初始化接口
		public.GET("/scloud/init", controllers.ScloudInit)

		// 验证码接口
		public.GET("/scloud/validateCode", controllers.ScloudValidateCode)

		// 登录接口
		public.POST("/scloud/login", controllers.ScloudLogin)
		public.POST("/api/m/sys/mLogin", controllers.ApiMSysMLogin)
		public.POST("/api/user/wx/login", controllers.WechatLogin)

		// 测试接口
		public.GET("/test/school-connection", controllers.TestSchoolServerConnection)

		// 静态文件服务
		public.GET("/uploads/images/:filename", controllers.ServeUploadedFile)

		// 字体文件静态服务
		r.Static("/static", "./static")

		// 微信支付回调（不需要认证）
		public.POST("/api/payment/wechat/callback", controllers.WechatPayCallback)

		// 健康检查接口（公开访问）
		public.GET("/api/health/school-server", controllers.GetSchoolServerStatus)
		public.POST("/api/health/school-server/check", controllers.ForceCheckSchoolServer)
		public.GET("/api/health/maintenance-info", controllers.GetMaintenanceInfo)

		// 应用更新接口（公开访问）
		public.GET("/api/app/info", controllers.GetCurrentAppInfo)
	}

	// 需要认证的路由
	auth := r.Group("/")
	auth.Use(middleware.JWTAuth())
	{
		// 学校信息接口
		auth.GET("/scloudoa/sys/user/getSchool", controllers.GetSchoolProxy)

		// 课程表接口
		auth.POST("/scloud/courseTimetableDetail/getTimetableDay",
			controllers.GetTimetableDay)
		auth.POST("/scloud/courseTimetableDetail/getTimetableList",
			controllers.GetTimetableList)
		auth.GET("/scloud/courseTimetable/getTermWeekNum",
			controllers.GetTermWeekNum)
		auth.GET("/scloud/courseTimetableDetail/getCourseLessonTime",
			controllers.GetCourseLessonTime)
		auth.GET("/scloudoa/scs/course/tCourseTimetableDetail/getCurrentTime",
			controllers.GetCurrentTimeProxy)
		// 课程表API - 移除智能缓存，使用原有的用户个人同步缓存机制
		auth.GET("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByDay",
			controllers.GetCourseTimeTableByDayProxy)
		auth.GET("/scloudoa/scs/course/tCourseTimetableDetail/getCourseTimeTableByWeek",
			controllers.GetCourseTimeTableByWeekProxy)
		auth.GET("/scloudoa/scs/course/tCourseTimetableDetail/getCourseLessonTime",
			controllers.GetCourseLessonTimeProxy)
		// 用户课程表配置 - 移除智能缓存，使用原有缓存机制
		auth.GET("/scloudoa/userQuery/tSysUser/getCourseSchoolTimetable",
			controllers.GetCourseSchoolTimetableProxy)

		// 新闻接口
		auth.GET("/scloudoa/scs/news/eoaNewsType/getEoaNewsTypeList",
			controllers.GetEoaNewsTypeListProxy)
		auth.GET("/scloudoa/scs/news/eoaNews/getEoaNewsListByTypeId",
			controllers.GetEoaNewsListByTypeIdProxy)

		// 附件下载接口
		auth.GET("/api/proxy/download-attachment", controllers.DownloadAttachmentProxy)
		auth.GET("/api/proxy/file/*path", controllers.FileProxy)

		// 学习数据接口 - 移除智能缓存，使用原有缓存机制
		auth.GET("/scloudoa/scs/course/tCourseScore/getLearningData",
			controllers.GetLearningDataProxy)
		auth.GET("/scloudoa/scs/course/tCourseScore/getSemester",
			controllers.GetSemesterProxy)
		auth.GET("/api/m/scs/course/tCourseScore/getSemester",
			controllers.GetSemesterProxy)
		auth.GET("/scloudoa/scs/course/tCourseScore/getCurrentTime",
			controllers.GetScoreCurrentTimeProxy)
		// 课程计划和成绩接口 - 移除智能缓存，使用原有缓存机制
		auth.GET("/scloudoa/scs/course/tCourseScore/getCoursePlan",
			controllers.GetCoursePlanProxy)
		auth.GET("/scloudoa/scs/course/tCourseScore/getCourseScore",
			controllers.GetCourseScoreProxy)

		// 成绩查询接口 - 移除智能缓存，使用原有缓存机制
		auth.GET("/api/m/scs/course/tCourseScore/getScoreList",
			controllers.GetScoreListProxy)
		auth.GET("/api/m/scs/course/tCourseScore/getSemesterScore",
			controllers.GetSemesterScoreProxy)
		auth.GET("/scloudoa/scs/course/tCourseScore/getScoreList",
			controllers.GetScoreListProxy)
		auth.GET("/scloudoa/scs/course/tCourseScore/getSemesterScore",
			controllers.GetSemesterScoreProxy)

		// 发现页面代理路由
		auth.GET("/scloudoa/classroomAttendance/list", controllers.ClassroomAttendanceListProxy)
		auth.GET("/scloudoa/course/tCourseOptionalStudent/getOptionalTeachingClass", controllers.GetOptionalTeachingClassProxy)
		auth.GET("/scloudoa/course/tCourseOptionalStudent/getSemesterOptionalTeachingClass", controllers.GetSemesterOptionalTeachingClassProxy)
		auth.GET("/scloudoa/scs/leave/tStudentLeave/getTStudentLeave", controllers.GetTStudentLeaveProxy)
		auth.GET("/scloudoa/scs/survey/tOaSurvey/getTOaSurvey", controllers.GetTOaSurveyProxy)
		auth.GET("/scloudoa/scs/survey/tOaSurveyQuestion/getTOaSurveyQuestion", controllers.GetTOaSurveyQuestionProxy)
		auth.GET("/scloudoa/scs/survey/tOaSurveyQuestionAnswer/getTOaSurveyQuestionAnswer", controllers.GetTOaSurveyQuestionAnswerProxy)
		auth.GET("/scloudoa/sys/user/queryById", controllers.QueryUserByIdProxy)

		// 档案编辑辅助数据代理路由
		auth.GET("/scloud/student/base/relation", controllers.GetRelationTypesProxy)
		auth.GET("/scloud/student/base/getProvince", controllers.GetProvinceProxy)

		// PC端专属接口（基于JSESSIONID会话）
		auth.GET("/api/pc/login/init", controllers.PCLoginInit)
		auth.POST("/api/pc/login/submit", controllers.PCLoginSubmit)
		auth.GET("/api/pc/captcha-image", controllers.PCGetCaptchaImage)
		auth.GET("/api/pc/teachers", controllers.PCGetTeachers)
		auth.GET("/api/pc/student-info", controllers.PCGetStudentInfo)
		auth.GET("/api/pc/session-status", controllers.PCGetSessionStatus)
		auth.POST("/api/pc/logout", controllers.PCLogout)
		auth.POST("/api/pc/archive-edit", controllers.PCSubmitArchiveEdit)
		auth.POST("/api/pc/auto-login", controllers.PCAutoLogin)
		auth.GET("/api/pc/session-credentials", controllers.PCGetSessionCredentials)
		auth.POST("/api/pc/session-credentials", controllers.PCSetSessionCredentials)

		// PC端专属业务接口（缺课统计/补考查询/违纪查询）
		auth.GET("/api/pc/miss-class", controllers.PCMissClassGetList)
		auth.GET("/api/pc/makeup-exam", controllers.PCMakeupExamQuery)
		auth.GET("/api/pc/disciplinary", controllers.PCDisciplinaryQuery)
		auth.GET("/api/pc/disciplinary/types", controllers.PCDisciplinaryTypes)
		auth.GET("/api/pc/disciplinary/levels", controllers.PCDisciplinaryLevels)

		// 实习相关代理路由
		auth.GET("/scloudoa/jobInternship/internshipRequirementsInquiry/list", controllers.InternshipRequirementsListProxy)
		auth.GET("/scloudoa/jobInternship/internshipPositionInquiry/getJobClassification", controllers.InternshipJobClassificationProxy)
		auth.GET("/scloudoa/jobInternship/internshipPositionInquiry/getIndustryClassification", controllers.InternshipIndustryClassificationProxy)
		auth.GET("/scloudoa/jobInternship/internshipPositionInquiry/list", controllers.InternshipPositionListProxy)
		auth.GET("/scloudoa/jobInternship/internshipPositionApplication/list", controllers.InternshipApplicationListProxy)
		auth.GET("/scloudoa/jobInternship/internshipSignIn/getPlan", controllers.InternshipSignInPlanProxy)
		auth.GET("/scloudoa/jobInternship/myInternshipPlan/list", controllers.MyInternshipPlanListProxy)
		auth.GET("/scloudoa/jobInternship/myInternshipSummary/list", controllers.MyInternshipSummaryListProxy)
		auth.GET("/scloudoa/jobInternship/internshipAppraisalForm/list", controllers.InternshipAppraisalFormListProxy)
		auth.GET("/scloudoa/jobInternship/internshipScoreInquiry/list", controllers.InternshipScoreInquiryListProxy)

		// 报修管理代理路由
		auth.GET("/scloudoa/repairReport/tLogisticsMaintenanceOrder/list", controllers.RepairListProxy)
		auth.GET("/scloudoa/repairReport/tLogisticsMaintenanceOrder/getStudent", controllers.RepairDetailProxy)
		auth.GET("/scloudoa/repairReport/tLogisticsMaintenanceOrder/getLogisticsMaintenanceOrderType", controllers.RepairTypesProxy)
		auth.GET("/scloudoa/sys/common/getFileServerUrl", controllers.FileServerUrlProxy)

		// 银行卡信息代理路由
		auth.GET("/scloudoa/studentBank/getStudentBank", controllers.GetStudentBankProxy)
		auth.PUT("/scloudoa/studentBank/edit", controllers.EditStudentBankProxy)

		// 评教接口
		auth.GET("/scloud/educational/evaluation/getEvaluationList",
			controllers.GetEvaluationList)
		auth.GET("/scloud/educational/evaluation/getEvaluationNorm/:id",
			controllers.GetEvaluationNorm)
		auth.GET("/scloudoa/evaluation/tCourseTeachingEvaluationFirstLevel/getEvaluationStudentConfigList",
			controllers.GetEvaluationStudentConfigListProxy)
		auth.POST("/scloud/educational/evaluation/submit", controllers.SubmitEvaluation)

		// 用户信息接口
		auth.GET("/api/user/info", controllers.GetUserInfo)
		auth.PUT("/api/user/info", controllers.UpdateUserInfo)
		auth.GET("/api/user/archive", controllers.GetUserArchive)
		auth.GET("/api/user/statistics", controllers.GetUserStatistics)

		// 学校账号绑定接口
		auth.POST("/api/user/school/bind", controllers.SchoolBind)

		// 用户设置接口
		auth.GET("/api/user/settings", controllers.GetUserSettings)
		auth.POST("/api/user/settings", controllers.UpdateUserSettings)
		auth.POST("/api/user/settings/reset", controllers.ResetUserSettings)

		// 同步设置接口
		auth.GET("/api/sync/settings", controllers.GetSyncSettings)
		auth.POST("/api/sync/settings", controllers.UpdateSyncSettings)
		auth.GET("/api/sync/status", controllers.GetSyncStatus)
		auth.POST("/api/sync/manual", controllers.ManualSync)
		auth.GET("/api/sync/logs", controllers.GetSyncLogs)

		// 管理员接口（可选）
		auth.GET("/api/admin/user-settings", controllers.GetAllUserSettings)

		// 社团管理接口
		auth.POST("/api/clubs", controllers.CreateClub)
		auth.GET("/api/clubs", controllers.GetClubsList)
		auth.GET("/api/clubs/my", controllers.GetMyClubs) // 必须在 :id 路由之前
		auth.GET("/api/clubs/:id", controllers.GetClubDetail)
		auth.PUT("/api/clubs/:id", controllers.UpdateClub)
		auth.DELETE("/api/clubs/:id", controllers.DeleteClub)
		auth.POST("/api/clubs/:id/join", controllers.JoinClub)
		auth.POST("/api/clubs/:id/leave", controllers.LeaveClub)
		auth.GET("/api/clubs/:id/members", controllers.GetClubMembers)
		auth.PUT("/api/clubs/:id/members/:memberId/role", controllers.UpdateMemberRole)

		// 用户主页接口
		auth.GET("/api/users/:id/profile", controllers.GetUserProfile)
		auth.GET("/api/users/:id/posts", controllers.GetUserPosts)
		auth.GET("/api/users/:id/followers", controllers.GetFollowers)
		auth.GET("/api/users/:id/following", controllers.GetFollowing)
		auth.POST("/api/users/:id/follow", controllers.FollowUserHandler)
		auth.POST("/api/users/:id/unfollow", controllers.UnfollowUserHandler)

		// 帖子管理接口
		auth.POST("/api/posts", controllers.CreatePost)
		auth.GET("/api/posts", controllers.GetPostsList)
		auth.GET("/api/posts/:id", controllers.GetPostDetail)
		auth.PUT("/api/posts/:id", controllers.UpdatePost)
		auth.DELETE("/api/posts/:id", controllers.DeletePost)
		auth.POST("/api/posts/:id/like", controllers.LikePost)
		auth.POST("/api/posts/:id/unlike", controllers.UnlikePost)

		// 评论管理接口
		auth.POST("/api/comments", controllers.CreateComment)
		auth.GET("/api/comments", controllers.GetCommentsList)
		auth.PUT("/api/comments/:id", controllers.UpdateComment)
		auth.DELETE("/api/comments/:id", controllers.DeleteComment)

		// 评论回复接口
		auth.GET("/api/comments/:id/replies", controllers.GetCommentReplies)
		auth.POST("/api/comments/:id/reply", controllers.ReplyToComment)

		// 评论点赞接口
		auth.POST("/api/comments/:id/like", controllers.LikeComment)
		auth.POST("/api/comments/:id/unlike", controllers.UnlikeComment)

		// 热门评论接口
		auth.GET("/api/comments/hot", controllers.GetHotComments)

		// 推荐相关API
		auth.GET("/api/community/recommend", controllers.GetRecommendData)

		// 文件上传接口
		auth.POST("/api/upload/image", controllers.UploadImage)
		auth.POST("/api/upload/images", controllers.UploadMultipleImages)
		auth.GET("/api/upload/token", controllers.GetUploadToken)
		auth.GET("/api/upload/stats", controllers.GetUploadStats)
		auth.DELETE("/api/upload/file", controllers.DeleteUploadedFile)

		// 书签/收藏接口
		auth.POST("/api/bookmarks", controllers.CreateBookmark)
		auth.DELETE("/api/bookmarks/:postId", controllers.DeleteBookmark)
		auth.GET("/api/bookmarks", controllers.GetBookmarks)
		auth.GET("/api/bookmarks/status/:postId", controllers.CheckBookmarkStatus)

		// 通知接口
		auth.GET("/api/notifications", controllers.GetNotifications)
		auth.PUT("/api/notifications/:id/read", controllers.MarkNotificationAsRead)
		auth.PUT("/api/notifications/read-all", controllers.MarkAllNotificationsAsRead)
		auth.GET("/api/notifications/unread-count", controllers.GetUnreadNotificationCount)

		// 举报接口
		auth.POST("/api/reports", controllers.CreateReport)
		auth.GET("/api/reports", controllers.GetReports)              // 管理员
		auth.PUT("/api/reports/:id/review", controllers.ReviewReport) // 管理员

		// 用户屏蔽接口
		auth.POST("/api/users/:id/block", controllers.BlockUser)
		auth.DELETE("/api/users/:id/unblock", controllers.UnblockUser)
		auth.GET("/api/users/blocked", controllers.GetBlockedUsers)
		auth.GET("/api/users/:id/block-status", controllers.CheckBlockStatus)

		// 内容审核接口（管理员）
		auth.GET("/api/moderation/settings", controllers.GetModerationSettings)
		auth.PUT("/api/moderation/settings", controllers.UpdateModerationSetting)
		auth.GET("/api/moderation/pending", controllers.GetPendingContent)
		auth.PUT("/api/moderation/reports/:reportId/approve", controllers.ApproveContent)
		auth.PUT("/api/moderation/reports/:reportId/reject", controllers.RejectContent)

		// 社区须知接口
		communityTermsCtrl := &controllers.CommunityTermsController{}
		auth.GET("/api/community/terms/status", communityTermsCtrl.GetTermsStatus)
		auth.POST("/api/community/terms/agree", communityTermsCtrl.AgreeTerms)

		// 多渠道通知接口
		auth.GET("/api/notification-channels", controllers.GetNotificationChannels)
		auth.PUT("/api/notification-channels", controllers.UpdateNotificationChannels)
		auth.POST("/api/notification-channels/test/:type", controllers.TestNotificationChannel)
		auth.POST("/api/notification-channels/trigger-score-check", controllers.TriggerScoreCheck)
		auth.GET("/api/notification-channels/debug-score-check", controllers.DebugScoreCheck)
		auth.DELETE("/api/notification-channels/clear-error-cache", controllers.ClearErrorCache)

		// 通知设置接口（新增）
		notificationSettingsController := &controllers.NotificationSettingsController{}
		auth.GET("/api/notification-settings", notificationSettingsController.GetNotificationSettings)
		auth.PUT("/api/notification-settings", notificationSettingsController.UpdateNotificationSettings)
		auth.GET("/api/notification-settings/semesters", notificationSettingsController.GetAvailableSemesters)
		auth.GET("/api/notification-settings/current-semester", notificationSettingsController.GetCurrentSemester)
		auth.POST("/api/notification-settings/test-score-check", notificationSettingsController.TestScoreCheck)
		auth.GET("/api/notification-settings/logs", notificationSettingsController.GetNotificationLogs)
		auth.GET("/api/notification-settings/score-check-logs", notificationSettingsController.GetScoreCheckLogs)

		// 短信余额和交易接口
		auth.GET("/api/sms/balance", controllers.GetSMSBalance)
		auth.GET("/api/sms/transactions", controllers.GetSMSTransactions)
		auth.GET("/api/sms/recharge-packages", controllers.GetRechargePackages)
		auth.POST("/api/sms/recharge", controllers.CreateRechargeOrder)

		// 短信模板管理接口
		auth.GET("/api/sms/templates", controllers.GetSMSTemplates)
		auth.GET("/api/sms/templates/enabled", controllers.GetEnabledSMSTemplates)
		auth.GET("/api/sms/templates/:type", controllers.GetSMSTemplate)
		auth.POST("/api/sms/templates/preview", controllers.PreviewSMSTemplate)
		auth.POST("/api/sms/templates/validate", controllers.ValidateSMSTemplateParams)
		auth.POST("/api/sms/templates/test", controllers.SendTestSMS)
		auth.POST("/api/sms/templates/batch", controllers.BatchSendSMS)
		auth.GET("/api/sms/cost", controllers.GetSMSCost)

		// 短信发送接口
		auth.POST("/api/sms/verification-code", controllers.SendVerificationCode)
		auth.POST("/api/sms/course-reminder", controllers.SendCourseReminder)
		auth.POST("/api/sms/exam-notice", controllers.SendExamNotice)
		auth.POST("/api/sms/evaluation-reminder", controllers.SendEvaluationReminder)

		// 邮件模板管理接口
		auth.GET("/api/email/templates", controllers.GetEmailTemplates)
		auth.GET("/api/email/templates/enabled", controllers.GetEnabledEmailTemplates)
		auth.GET("/api/email/templates/:type", controllers.GetEmailTemplate)
		auth.GET("/api/email/templates/category", controllers.GetEmailTemplatesByCategory)
		auth.POST("/api/email/templates/preview", controllers.PreviewEmailTemplate)
		auth.POST("/api/email/templates/validate", controllers.ValidateEmailTemplateParams)
		auth.POST("/api/email/templates/test", controllers.SendTestEmail)
		auth.POST("/api/email/templates/batch", controllers.BatchSendEmail)

		// 邮件发送接口
		auth.POST("/api/email/course-reminder", controllers.SendCourseReminderEmail)
		auth.POST("/api/email/exam-notice", controllers.SendExamNoticeEmail)
		auth.POST("/api/email/evaluation-reminder", controllers.SendEvaluationReminderEmail)

		// 钉钉模板管理接口
		auth.GET("/api/dingtalk/templates", controllers.GetDingTalkTemplates)
		auth.GET("/api/dingtalk/templates/enabled", controllers.GetEnabledDingTalkTemplates)
		auth.GET("/api/dingtalk/templates/:type", controllers.GetDingTalkTemplate)
		auth.GET("/api/dingtalk/templates/category", controllers.GetDingTalkTemplatesByCategory)
		auth.POST("/api/dingtalk/templates/preview", controllers.PreviewDingTalkTemplate)
		auth.POST("/api/dingtalk/templates/validate", controllers.ValidateDingTalkTemplateParams)
		auth.POST("/api/dingtalk/templates/test", controllers.SendTestDingTalk)
		auth.POST("/api/dingtalk/templates/batch", controllers.BatchSendDingTalk)

		// 钉钉发送接口
		auth.POST("/api/dingtalk/course-reminder", controllers.SendCourseReminderDingTalk)
		auth.POST("/api/dingtalk/exam-notice", controllers.SendExamNoticeDingTalk)
		auth.POST("/api/dingtalk/evaluation-reminder", controllers.SendEvaluationReminderDingTalk)

		// 通知日志接口
		auth.GET("/api/notification-logs", controllers.GetNotificationLogs)

		// 支付状态查询接口
		auth.GET("/api/payment/status/:orderId", controllers.QueryPaymentStatus)

		// 管理员支付管理接口
		auth.POST("/api/payment/manual-process", controllers.ManualProcessPayment)
		auth.GET("/api/payment/certificates", controllers.GetPaymentCertificates)

		// 应用更新接口
		auth.POST("/api/app/check-update", controllers.CheckAppUpdate)
		auth.GET("/api/app/versions", controllers.GetAppVersions)
		auth.POST("/api/app/versions", controllers.CreateAppVersion)
		auth.PUT("/api/app/versions/:id", controllers.UpdateAppVersion)
		auth.DELETE("/api/app/versions/:id", controllers.DeleteAppVersion)
	}

	// 启动学校服务器健康检查服务
	healthCheckService := services.GetSchoolHealthCheckService()
	healthCheckService.Start()

	// 获取端口配置
	port := config.GetPort()

	// 设置优雅关闭
	go func() {
		// 启动服务器
		log.Printf("服务器启动在 http://localhost:%s", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 停止同步服务
	controllers.StopSyncService()
	log.Println("同步服务已停止")

	if scheduler := services.GetOfflineCacheScheduler(); scheduler != nil {
		scheduler.Stop()
		log.Println("离线缓存定时任务已停止")
	}

	// 停止健康检查服务
	healthCheckService.Stop()
	log.Println("健康检查服务已停止")

	log.Println("服务器已关闭")
}
