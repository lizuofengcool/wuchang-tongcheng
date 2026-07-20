// Package linggong 同城零工兼职模块插件
// 提供岗位发布/任务包/报名/雇主认证/求职者档案/合同/支付/评价/技能/证书/信用/审核规则/统计等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标斗米/青团兼职/兼职猫/猪八戒
package linggong

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/linggong/handler"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 零工兼职模块插件
type Plugin struct {
	name    string
	version string

	// 12 个 Handler
	linggongHandler       *handler.LinggongHandler
	taskHandler           *handler.TaskHandler
	applicationHandler    *handler.ApplicationHandler
	employerHandler       *handler.EmployerHandler
	workerHandler         *handler.WorkerHandler
	contractHandler       *handler.ContractHandler
	paymentHandler        *handler.PaymentHandler
	ratingHandler         *handler.RatingHandler
	skillHandler          *handler.SkillHandler
	certificationHandler  *handler.CertificationHandler
	creditHandler         *handler.CreditHandler
	auditRuleHandler      *handler.AuditRuleHandler
}

// NewPlugin 创建零工兼职模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "linggong", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "linggong",
		DisplayName: "同城零工兼职",
		Category:    "business",
		Description: "同城零工兼职完整功能：岗位发布/任务包/报名/雇主认证/求职者档案/合同/支付/评价/技能/证书/信用/审核规则/统计",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：18 张表由 backend/migrations 下 SQL 迁移脚本创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
// 此处仅 AutoMigrate 主表 Linggong（与 pinche 模块保持一致），子表由 SQL 脚本创建。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 linggong 主表（子表由 backend/migrations 下 SQL 脚本创建）
	if err := db.AutoMigrate(&model.Linggong{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（18 个） =====
	linggongRepo := repository.NewLinggongRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	employerRepo := repository.NewEmployerRepository(db)
	workerRepo := repository.NewWorkerRepository(db)
	contractRepo := repository.NewContractRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	ratingRepo := repository.NewRatingRepository(db)
	skillRepo := repository.NewSkillRepository(db)
	certificationRepo := repository.NewCertificationRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	disputeRepo := repository.NewDisputeRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)

	// ===== 依赖注入：Service 层（18 个） =====
	linggongSvc := service.NewLinggongService(linggongRepo)
	taskSvc := service.NewTaskService(taskRepo)
	applicationSvc := service.NewApplicationService(applicationRepo)
	attendanceSvc := service.NewAttendanceService(attendanceRepo)
	employerSvc := service.NewEmployerService(employerRepo)
	workerSvc := service.NewWorkerService(workerRepo)
	contractSvc := service.NewContractService(contractRepo)
	paymentSvc := service.NewPaymentService(paymentRepo)
	ratingSvc := service.NewRatingService(ratingRepo)
	skillSvc := service.NewSkillService(skillRepo)
	certificationSvc := service.NewCertificationService(certificationRepo)
	creditSvc := service.NewCreditService(creditRepo)
	disputeSvc := service.NewDisputeService(disputeRepo)
	withdrawalSvc := service.NewWithdrawalService(withdrawalRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo)
	recommendationSvc := service.NewRecommendationService(recommendationRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	statisticSvc := service.NewStatisticService(statisticRepo)

	// ===== 依赖注入：Handler 层（12 个） =====
	p.linggongHandler = handler.NewLinggongHandler(linggongSvc)
	p.taskHandler = handler.NewTaskHandler(taskSvc)
	p.applicationHandler = handler.NewApplicationHandler(applicationSvc)
	p.employerHandler = handler.NewEmployerHandler(employerSvc)
	p.workerHandler = handler.NewWorkerHandler(workerSvc)
	p.contractHandler = handler.NewContractHandler(contractSvc)
	p.paymentHandler = handler.NewPaymentHandler(paymentSvc)
	p.ratingHandler = handler.NewRatingHandler(ratingSvc)
	p.skillHandler = handler.NewSkillHandler(skillSvc)
	p.certificationHandler = handler.NewCertificationHandler(certificationSvc)
	p.creditHandler = handler.NewCreditHandler(creditSvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)

	// 避免未使用变量告警（attendance/dispute/withdrawal/favorite/recommendation/statistic 暂未在 handler 层直接使用，保留以备后续扩展）
	_ = attendanceSvc
	_ = disputeSvc
	_ = withdrawalSvc
	_ = favoriteSvc
	_ = recommendationSvc
	_ = statisticSvc

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/linggong
//
// 路由分组：
//   - 公开路由（C 端浏览，无需登录）：岗位列表/详情/搜索/附近/技能/分类/审核规则只读
//   - 需登录路由（C 端发布/交易）：岗位 CRUD/任务/报名/签约/考勤/评价/举报/提现
//   - 管理后台路由（需 linggong:audit 权限）：审核/批量/统计/审核规则
//
// 注意：固定路径（/search /nearby /mine /hot 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "linggong_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "linggong_write"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "linggong_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "linggong_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("linggong:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 岗位主表（linggongs）
	router.GET("", readLimiter, p.linggongHandler.List)
	router.GET("/search", searchLimiter, p.linggongHandler.Search)
	router.GET("/nearby", nearbyLimiter, p.linggongHandler.Nearby)
	router.GET("/:id", readLimiter, p.linggongHandler.GetByID)
	router.POST("/:id/contact", p.linggongHandler.IncrContact)
	router.POST("/:id/share", p.linggongHandler.IncrShare)
	router.POST("/:id/view", p.linggongHandler.IncrView)
	router.GET("/employers/:id/linggongs", readLimiter, p.linggongHandler.ListByEmployer)

	// 任务包（tasks） - 公开查询
	router.GET("/tasks", readLimiter, p.taskHandler.List)
	router.GET("/tasks/:id", readLimiter, p.taskHandler.GetByID)
	router.GET("/tasks/employer/:id", readLimiter, p.taskHandler.ListByEmployer)
	router.GET("/:id/tasks", readLimiter, p.taskHandler.ListByLinggong)

	// 报名申请（applications） - 公开查询
	router.GET("/applications", readLimiter, p.applicationHandler.List)
	router.GET("/applications/:id", readLimiter, p.applicationHandler.GetByID)
	router.GET("/:id/applications", readLimiter, p.applicationHandler.ListByLinggong)

	// 雇主认证（employers）- 公开查询
	router.GET("/employers", readLimiter, p.employerHandler.List)
	router.GET("/employers/:id", readLimiter, p.employerHandler.GetByID)

	// 求职者档案（workers）- 公开查询
	router.GET("/workers", readLimiter, p.workerHandler.List)
	router.GET("/workers/:id", readLimiter, p.workerHandler.GetByID)
	router.GET("/workers/:id/certifications", readLimiter, p.certificationHandler.ListByWorker)

	// 合同（contracts）- 公开查询
	router.GET("/contracts", readLimiter, p.contractHandler.List)
	router.GET("/contracts/:id", readLimiter, p.contractHandler.GetByID)
	router.GET("/contracts/no/:contract_no", readLimiter, p.contractHandler.GetByContractNo)
	router.GET("/:id/contracts", readLimiter, p.contractHandler.ListByLinggong)

	// 支付（payments）- 公开查询
	router.GET("/payments/:id", readLimiter, p.paymentHandler.GetByID)
	router.GET("/payments/no/:payment_no", readLimiter, p.paymentHandler.GetByPaymentNo)
	router.GET("/:id/payments", readLimiter, p.paymentHandler.ListByLinggong)

	// 评价（ratings）- 公开查询
	router.GET("/ratings", readLimiter, p.ratingHandler.List)
	router.GET("/ratings/stats", readLimiter, p.ratingHandler.Stats)
	router.GET("/ratings/:id", readLimiter, p.ratingHandler.GetByID)
	router.GET("/ratings/no/:rating_no", readLimiter, p.ratingHandler.GetByRatingNo)
	router.GET("/ratings/target/:target_type/:id", readLimiter, p.ratingHandler.ListByTarget)
	router.GET("/:id/ratings", readLimiter, p.ratingHandler.ListByLinggong)

	// 技能标签（skills）- 公开只读
	router.GET("/skills", readLimiter, p.skillHandler.List)
	router.GET("/skills/hot", readLimiter, p.skillHandler.ListHot)
	router.GET("/skills/category/:category", readLimiter, p.skillHandler.ListByCategory)
	router.GET("/skills/parent/:id", readLimiter, p.skillHandler.ListByParent)
	router.GET("/skills/:id", readLimiter, p.skillHandler.GetByID)

	// 资质证书（certifications）- 公开查询
	router.GET("/certifications", readLimiter, p.certificationHandler.List)
	router.GET("/certifications/:id", readLimiter, p.certificationHandler.GetByID)

	// 信用分（credits）- 公开查询
	router.GET("/credits/:id", readLimiter, p.creditHandler.GetByID)
	router.GET("/credits", readLimiter, p.creditHandler.List)
	router.GET("/credits/user/:id", readLimiter, p.creditHandler.ListByUser)

	// 审核规则（audit-rules）- 公开只读
	router.GET("/audit-rules", readLimiter, p.auditRuleHandler.List)
	router.GET("/audit-rules/enabled", readLimiter, p.auditRuleHandler.ListEnabled)
	router.GET("/audit-rules/type/:rule_type", readLimiter, p.auditRuleHandler.ListByType)
	router.GET("/audit-rules/:id", readLimiter, p.auditRuleHandler.GetByID)

	// ==================== 需登录路由（C 端发布/收藏/交易） ====================

	// 岗位主表 CRUD
	router.POST("", auth, writeLimiter, p.linggongHandler.Create)
	router.PUT("/:id", auth, p.linggongHandler.Update)
	router.DELETE("/:id", auth, p.linggongHandler.Delete)
	router.GET("/mine", auth, readLimiter, p.linggongHandler.ListMine)

	// 任务包 CRUD + 领取/交付/验收
	router.POST("/tasks", auth, writeLimiter, p.taskHandler.Create)
	router.PUT("/tasks/:id", auth, p.taskHandler.Update)
	router.DELETE("/tasks/:id", auth, p.taskHandler.Delete)
	router.POST("/tasks/:id/claim", auth, writeLimiter, p.taskHandler.Claim)
	router.POST("/tasks/:id/submit", auth, writeLimiter, p.taskHandler.Submit)
	router.POST("/tasks/:id/verify", auth, writeLimiter, p.taskHandler.Verify)

	// 报名 CRUD + 审核/取消
	router.POST("/applications", auth, writeLimiter, p.applicationHandler.Create)
	router.PUT("/applications/:id", auth, p.applicationHandler.Update)
	router.DELETE("/applications/:id", auth, p.applicationHandler.Delete)
	router.PUT("/applications/:id/audit", auth, p.applicationHandler.Audit)
	router.POST("/applications/:id/cancel", auth, writeLimiter, p.applicationHandler.Cancel)
	router.GET("/applications/mine", auth, readLimiter, p.applicationHandler.ListByWorker)
	router.GET("/applications/employer", auth, readLimiter, p.applicationHandler.ListByEmployer)

	// 雇主认证 CRUD
	router.POST("/employers", auth, writeLimiter, p.employerHandler.Create)
	router.PUT("/employers/:id", auth, p.employerHandler.Update)
	router.DELETE("/employers/:id", auth, p.employerHandler.Delete)
	router.GET("/employers/me", auth, readLimiter, p.employerHandler.GetByUserID)

	// 求职者档案 CRUD
	router.POST("/workers", auth, writeLimiter, p.workerHandler.Create)
	router.PUT("/workers/:id", auth, p.workerHandler.Update)
	router.DELETE("/workers/:id", auth, p.workerHandler.Delete)
	router.GET("/workers/me", auth, readLimiter, p.workerHandler.GetByUserID)

	// 合同 CRUD + 签署
	router.POST("/contracts", auth, writeLimiter, p.contractHandler.Create)
	router.PUT("/contracts/:id", auth, p.contractHandler.Update)
	router.DELETE("/contracts/:id", auth, p.contractHandler.Delete)
	router.POST("/contracts/:id/sign", auth, writeLimiter, p.contractHandler.Sign)
	router.PUT("/contracts/:id/status", auth, p.contractHandler.UpdateStatus)
	router.GET("/contracts/employer/:id", auth, readLimiter, p.contractHandler.ListByEmployer)
	router.GET("/contracts/worker/:id", auth, readLimiter, p.contractHandler.ListByWorker)

	// 支付 CRUD + 结算
	router.POST("/payments", auth, writeLimiter, p.paymentHandler.Create)
	router.PUT("/payments/:id", auth, p.paymentHandler.Update)
	router.PUT("/payments/:id/status", auth, p.paymentHandler.UpdateStatus)
	router.POST("/payments/:id/settle", auth, writeLimiter, p.paymentHandler.Settle)
	router.GET("/payments/employer/:id", auth, readLimiter, p.paymentHandler.ListByEmployer)
	router.GET("/payments/worker/:id", auth, readLimiter, p.paymentHandler.ListByWorker)

	// 评价 CRUD + 回复/追评/点赞
	router.POST("/ratings", auth, writeLimiter, p.ratingHandler.Create)
	router.PUT("/ratings/:id", auth, p.ratingHandler.Update)
	router.DELETE("/ratings/:id", auth, p.ratingHandler.Delete)
	router.POST("/ratings/:id/reply", auth, writeLimiter, p.ratingHandler.Reply)
	router.POST("/ratings/:id/append", auth, writeLimiter, p.ratingHandler.Append)
	router.POST("/ratings/:id/like", auth, p.ratingHandler.Like)
	router.GET("/ratings/rater", auth, readLimiter, p.ratingHandler.ListByRater)

	// 资质证书 CRUD
	router.POST("/certifications", auth, writeLimiter, p.certificationHandler.Create)
	router.PUT("/certifications/:id", auth, p.certificationHandler.Update)
	router.DELETE("/certifications/:id", auth, p.certificationHandler.Delete)
	router.GET("/certifications/mine", auth, readLimiter, p.certificationHandler.ListMine)

	// 信用分查询
	router.GET("/credits/score", auth, readLimiter, p.creditHandler.GetScore)

	// ==================== 管理后台路由（需 linggong:audit 权限） ====================

	admin := router.Group("/admin")

	// 岗位管理
	admin.GET("/linggongs", auditPerm, p.linggongHandler.AdminList)
	admin.GET("/linggongs/:id", auditPerm, p.linggongHandler.AdminGetByID)
	admin.PUT("/linggongs/:id/audit", auditPerm, p.linggongHandler.Audit)
	admin.PUT("/linggongs/:id/status", auditPerm, p.linggongHandler.AdminUpdateStatus)
	admin.POST("/linggongs/batch-status", auditPerm, p.linggongHandler.BatchUpdateStatus)

	// 任务包管理
	admin.GET("/tasks", auditPerm, p.taskHandler.AdminList)
	admin.GET("/tasks/:id", auditPerm, p.taskHandler.AdminGetByID)
	admin.PUT("/tasks/:id/status", auditPerm, p.taskHandler.AdminUpdateStatus)

	// 报名管理
	admin.GET("/applications", auditPerm, p.applicationHandler.AdminList)

	// 雇主认证管理
	admin.GET("/employers", auditPerm, p.employerHandler.AdminList)
	admin.PUT("/employers/:id/audit", auditPerm, p.employerHandler.Audit)
	admin.PUT("/employers/:id/status", auditPerm, p.employerHandler.AdminUpdateStatus)

	// 求职者档案管理
	admin.GET("/workers", auditPerm, p.workerHandler.AdminList)
	admin.PUT("/workers/:id/audit", auditPerm, p.workerHandler.Audit)
	admin.PUT("/workers/:id/status", auditPerm, p.workerHandler.AdminUpdateStatus)

	// 合同管理
	admin.GET("/contracts", auditPerm, p.contractHandler.AdminList)

	// 支付管理
	admin.GET("/payments", auditPerm, p.paymentHandler.AdminList)

	// 评价管理
	admin.GET("/ratings", auditPerm, p.ratingHandler.AdminList)
	admin.PUT("/ratings/:id/audit", auditPerm, p.ratingHandler.Audit)

	// 技能管理
	admin.GET("/skills", auditPerm, p.skillHandler.AdminList)
	admin.POST("/skills", auditPerm, p.skillHandler.Create)
	admin.PUT("/skills/:id", auditPerm, p.skillHandler.Update)
	admin.DELETE("/skills/:id", auditPerm, p.skillHandler.Delete)
	admin.PUT("/skills/:id/status", auditPerm, p.skillHandler.AdminUpdateStatus)

	// 资质证书管理
	admin.PUT("/certifications/:id/verify", auditPerm, p.certificationHandler.Verify)

	// 信用分管理
	admin.POST("/credits/adjust", auditPerm, p.creditHandler.Adjust)
	admin.DELETE("/credits/:id", auditPerm, p.creditHandler.Delete)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.AdminList)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.AdminUpdateStatus)
	admin.POST("/audit-rules/batch-delete", auditPerm, p.auditRuleHandler.BatchDelete)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
