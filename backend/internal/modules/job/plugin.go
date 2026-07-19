// Package job 同城招聘求职模块插件
// 提供职位发布/搜索/附近/收藏/投递/面试/消息/公司/简历/评价/举报/统计及管理后台审核等业务
// 依据需求文档 2.2.A.10：招聘求职（对标 BOSS直聘/拉勾/58招聘）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案第四章
package job

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/job/handler"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 招聘求职模块插件
type Plugin struct {
	name    string
	version string

	// 8 个 Handler
	jobHandler         *handler.JobHandler
	companyHandler     *handler.CompanyHandler
	resumeHandler      *handler.ResumeHandler
	applicationHandler *handler.ApplicationHandler
	interviewHandler   *handler.InterviewHandler
	messageHandler     *handler.MessageHandler
	reportHandler      *handler.ReportHandler
	statisticsHandler  *handler.StatisticsHandler
}

// NewPlugin 创建招聘求职模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "job", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "job",
		DisplayName:  "同城招聘",
		Category:     "business",
		Description:  "同城招聘求职完整功能：职位/公司/认证/简历/投递/面试/消息/评价/举报/统计/审核",
		Version:      p.version,
		Dependencies: []string{"category"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：19 张表由 backend/migrations/006_job_full.sql 创建，
// 不使用 GORM AutoMigrate 以避免约束名不一致导致的 DROP CONSTRAINT 错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 jobs 主表（子表由 backend/migrations/006_job_full.sql 创建）
	// 注：之前未调用 AutoMigrate 导致 jobs 主表缺失，前端列表报 "relation jobs does not exist"
	if err := db.AutoMigrate(&model.Job{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层 =====
	jobRepo := repository.NewJobRepository(db)
	interactionRepo := repository.NewInteractionRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	certRepo := repository.NewCertificationRepository(db)
	resumeRepo := repository.NewResumeRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	interviewRepo := repository.NewInterviewRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// ===== 依赖注入：Service 层 =====
	jobSvc := service.NewJobService(jobRepo, interactionRepo)
	companySvc := service.NewCompanyService(companyRepo, certRepo, interactionRepo)
	resumeSvc := service.NewResumeService(resumeRepo)
	applicationSvc := service.NewApplicationService(applicationRepo, jobRepo, resumeRepo)
	interviewSvc := service.NewInterviewService(interviewRepo, applicationRepo)
	messageSvc := service.NewMessageService(messageRepo)
	reportSvc := service.NewReportService(reportRepo, reviewRepo)
	statsSvc := service.NewStatisticsService(db)

	// ===== 依赖注入：Handler 层 =====
	p.jobHandler = handler.NewJobHandler(jobSvc)
	p.companyHandler = handler.NewCompanyHandler(companySvc)
	p.resumeHandler = handler.NewResumeHandler(resumeSvc)
	p.applicationHandler = handler.NewApplicationHandler(applicationSvc)
	p.interviewHandler = handler.NewInterviewHandler(interviewSvc)
	p.messageHandler = handler.NewMessageHandler(messageSvc)
	p.reportHandler = handler.NewReportHandler(reportSvc)
	p.statisticsHandler = handler.NewStatisticsHandler(statsSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/job
//
// 路由分组（共 90+ API）：
//   - 公开路由（C 端浏览，无需登录）：职位列表/搜索/附近/详情/相似/公司/评价/统计
//   - 需登录路由（C 端发布/收藏/投递/面试/消息/举报/评价）：职位 CRUD/收藏/推广/公司/简历/投递/面试/消息/举报/评价
//   - 管理后台路由（需 job:audit 权限）：审核/列表/统计/举报处理/评价审核/认证审核
//
// 注意：固定路径（/search /nearby /mine /favorites /companies /resumes /applications /interviews /messages 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "job_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "job_write"))
	favLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "job_fav"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "job_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "job_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("job:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 职位基础（固定路径优先注册）
	router.GET("/search", searchLimiter, p.jobHandler.Search)
	router.GET("/nearby", nearbyLimiter, p.jobHandler.ListNearby)
	router.GET("/advanced-search", searchLimiter, p.jobHandler.AdvancedSearch)
	router.GET("/mine", auth, p.jobHandler.ListMine)
	router.GET("/favorites", auth, readLimiter, p.jobHandler.ListFavs)

	// 职位基础（含 :id 参数）
	router.GET("", readLimiter, p.jobHandler.List)
	router.GET("/:id", readLimiter, p.jobHandler.GetByID)
	router.GET("/:id/similar", readLimiter, p.jobHandler.ListSimilar)
	router.GET("/:id/fav", p.jobHandler.FavStatus)
	router.GET("/:id/applications", auth, readLimiter, p.applicationHandler.ListByJobID)

	// 公司
	router.GET("/companies", readLimiter, p.companyHandler.List)
	router.GET("/companies/:id", readLimiter, p.companyHandler.GetByID)
	router.GET("/companies/:id/certifications", readLimiter, p.companyHandler.ListCertsByCompany)
	router.GET("/companies/:id/reviews", readLimiter, p.reportHandler.ListReviewsByCompany)
	router.GET("/companies/:id/reviews/stats", readLimiter, p.reportHandler.ReviewStats)

	// 认证（公开查询）
	router.GET("/certifications", readLimiter, p.companyHandler.ListCerts)
	router.GET("/certifications/:id", readLimiter, p.companyHandler.GetCert)

	// 评价（公开查询）
	router.GET("/reviews", readLimiter, p.reportHandler.ListReviews)
	router.GET("/reviews/:id", readLimiter, p.reportHandler.GetReview)
	router.GET("/users/:user_id/reviews", readLimiter, p.reportHandler.ListReviewsByUser)

	// 数据统计（公开部分：热门职位/薪资趋势/分类/地区/职位趋势）
	router.GET("/statistics/hot-jobs", readLimiter, p.statisticsHandler.ExportReport)
	router.GET("/statistics/salary-trend", readLimiter, p.statisticsHandler.GetTradeStats)
	router.GET("/statistics/category", readLimiter, p.statisticsHandler.GetCompanyStats)
	router.GET("/statistics/region", readLimiter, p.statisticsHandler.GetTrendStats)
	router.GET("/statistics/job-trend", readLimiter, p.statisticsHandler.GetJobStats)

	// ==================== 需登录路由（C 端发布/收藏/投递/面试/消息） ====================

	// 职位基础 CRUD
	router.POST("", auth, writeLimiter, p.jobHandler.Create)
	router.PUT("/:id", auth, p.jobHandler.Update)
	router.DELETE("/:id", auth, p.jobHandler.Delete)
	router.PUT("/:id/status", auth, p.jobHandler.UpdateStatus)

	// 职位收藏
	router.POST("/:id/fav", auth, favLimiter, p.jobHandler.Fav)
	router.DELETE("/:id/fav", auth, p.jobHandler.Unfav)

	// 职位推广
	router.POST("/:id/promotion", auth, writeLimiter, p.jobHandler.Promotion)

	// 公司管理
	router.POST("/companies", auth, writeLimiter, p.companyHandler.Create)
	router.GET("/companies/mine", auth, p.companyHandler.GetMyCompany)
	router.PUT("/companies/:id", auth, p.companyHandler.Update)
	router.POST("/companies/:id/follow", auth, favLimiter, p.companyHandler.Follow)
	router.DELETE("/companies/:id/follow", auth, p.companyHandler.Unfollow)
	router.GET("/companies/following", auth, readLimiter, p.companyHandler.ListFollowing)

	// 企业认证
	router.POST("/certifications", auth, writeLimiter, p.companyHandler.CreateCert)

	// 简历管理
	router.POST("/resumes", auth, writeLimiter, p.resumeHandler.Create)
	router.PUT("/resumes/:id", auth, p.resumeHandler.Update)
	router.DELETE("/resumes/:id", auth, p.resumeHandler.Delete)
	router.GET("/resumes", auth, readLimiter, p.resumeHandler.List)
	router.GET("/resumes/mine", auth, p.resumeHandler.ListMine)
	router.GET("/resumes/default", auth, p.resumeHandler.GetDefault)
	router.GET("/resumes/:id", auth, readLimiter, p.resumeHandler.GetByID)
	router.PUT("/resumes/:id/default", auth, p.resumeHandler.SetDefault)
	router.PUT("/resumes/:id/status", auth, p.resumeHandler.UpdateStatus)

	// 投递记录
	router.POST("/applications", auth, writeLimiter, p.applicationHandler.Create)
	router.GET("/applications", auth, readLimiter, p.applicationHandler.List)
	router.GET("/applications/stats", auth, p.applicationHandler.Stats)
	router.GET("/applications/:id", auth, p.applicationHandler.GetByID)
	router.PUT("/applications/:id/status", auth, p.applicationHandler.StatusUpdate)
	router.POST("/applications/batch", auth, p.applicationHandler.BatchAction)

	// 面试邀约
	router.POST("/interviews", auth, writeLimiter, p.interviewHandler.Create)
	router.GET("/interviews", auth, readLimiter, p.interviewHandler.List)
	router.GET("/interviews/stats", auth, p.interviewHandler.Stats)
	router.GET("/interviews/:id", auth, p.interviewHandler.GetByID)
	router.PUT("/interviews/:id", auth, p.interviewHandler.Update)
	router.PUT("/interviews/:id/action", auth, p.interviewHandler.Action)
	router.PUT("/interviews/:id/feedback", auth, p.interviewHandler.Feedback)
	router.GET("/applications/:id/interviews", auth, p.interviewHandler.ListByApplication)

	// 沟通消息
	router.POST("/messages", auth, writeLimiter, p.messageHandler.Create)
	router.GET("/messages", auth, readLimiter, p.messageHandler.List)
	router.GET("/messages/unread/count", auth, p.messageHandler.CountUnread)
	router.GET("/messages/:id", auth, p.messageHandler.GetByID)
	router.DELETE("/messages/:id", auth, p.messageHandler.Delete)
	router.POST("/messages/batch-delete", auth, p.messageHandler.BatchDelete)
	router.PUT("/messages/:id/recall", auth, p.messageHandler.Recall)

	// 会话
	router.GET("/conversations", auth, p.messageHandler.ListConversations)
	router.GET("/conversations/:conversation_id/messages", auth, readLimiter, p.messageHandler.ListByConversation)
	router.PUT("/conversations/:conversation_id/read", auth, p.messageHandler.MarkRead)

	// 举报
	router.POST("/reports", auth, writeLimiter, p.reportHandler.CreateReport)
	router.GET("/reports/mine", auth, readLimiter, p.reportHandler.ListMyReports)
	router.POST("/reports/:id/appeal", auth, p.reportHandler.AppealReport)

	// 评价
	router.POST("/reviews", auth, writeLimiter, p.reportHandler.CreateReview)
	router.GET("/reviews/mine", auth, readLimiter, p.reportHandler.ListMyReviews)
	router.PUT("/reviews/:id", auth, p.reportHandler.UpdateReview)
	router.DELETE("/reviews/:id", auth, p.reportHandler.DeleteReview)
	router.POST("/reviews/:id/reply", auth, p.reportHandler.ReplyReview)
	router.POST("/reviews/:id/append", auth, p.reportHandler.AppendReview)
	router.POST("/reviews/:id/like", auth, favLimiter, p.reportHandler.LikeReview)

	// C 端统计
	router.GET("/statistics/recruiter", auth, p.statisticsHandler.GetUserStats)
	router.GET("/statistics/applicant", auth, p.statisticsHandler.GetApplicantStats)
	router.GET("/statistics/dashboard", auth, p.statisticsHandler.GetDashboard)

	// ==================== 管理后台路由（需 job:audit 权限） ====================

	admin := router.Group("/admin")

	// 职位管理
	admin.GET("/list", auditPerm, p.jobHandler.AdminList)
	admin.GET("/:id", auditPerm, p.jobHandler.AdminGetByID)
	admin.PUT("/:id/audit", auditPerm, p.jobHandler.Audit)
	admin.PUT("/:id/status", auditPerm, p.jobHandler.AdminUpdateStatus)

	// 公司审核
	admin.PUT("/companies/:id/audit", auditPerm, p.companyHandler.Audit)

	// 企业认证审核
	admin.PUT("/certifications/:id/process", auditPerm, p.companyHandler.ProcessCert)

	// 举报管理
	admin.GET("/reports", auditPerm, p.reportHandler.ListReports)
	admin.GET("/reports/:id", auditPerm, p.reportHandler.GetReport)
	admin.PUT("/reports/:id/process", auditPerm, p.reportHandler.HandleReport)
	admin.PUT("/reports/:id/appeal", auditPerm, p.reportHandler.ProcessAppeal)

	// 评价审核
	admin.PUT("/reviews/:id/audit", auditPerm, p.reportHandler.AuditReview)

	// 平台总览统计
	admin.GET("/statistics/overview", auditPerm, p.statisticsHandler.GetOverview)
	admin.GET("/statistics/conversion", auditPerm, p.statisticsHandler.GetFunnelStats)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
