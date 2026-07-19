// Package car 同城车辆买卖模块插件
// 提供车源发布/检测/评估/分期/车险/试驾/过户/合同/担保/评价/举报/统计等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标瓜子/人人车/懂车帝/毛豆新车/易鑫车贷
package car

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/car/handler"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/modules/car/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 车辆买卖模块插件
type Plugin struct {
	name    string
	version string

	// 14 个 Handler（聚合了 9 个文件）
	carHandler            *handler.CarHandler
	listingHandler        *handler.ListingHandler
	inspectionHandler     *handler.InspectionHandler
	evaluationHandler     *handler.EvaluationHandler
	financingHandler      *handler.FinancingHandler
	insuranceHandler      *handler.InsuranceHandler
	testDriveHandler      *handler.TestDriveHandler
	transferHandler       *handler.TransferHandler
	reportHandler         *handler.ReportHandler
	statisticsHandler     *handler.StatisticsHandler
	tradeHandler          *handler.TradeHandler
	recommendationHandler *handler.RecommendationHandler
	auditRuleHandler      *handler.AuditRuleHandler
	catalogHandler        *handler.CatalogHandler
}

// NewPlugin 创建车辆买卖模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "car", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "car",
		DisplayName: "同城车辆买卖",
		Category:    "business",
		Description: "同城车辆买卖完整功能：车源/发布单/检测/评估/分期/车险/试驾/过户/合同/担保/评价/举报/统计/推荐/审核规则/车型库",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：22 张表（含 8 张复用表）由 backend/migrations/004_car_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 cars 主表（子表由 backend/migrations/008_car_full.sql 创建）
	// 注：之前未调用 AutoMigrate 导致 cars 主表缺失，前端列表报 "relation cars does not exist"
	if err := db.AutoMigrate(&model.Car{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（20 个） =====
	carRepo := repository.NewCarRepository(db)
	imageRepo := repository.NewImageRepository(db)
	interactionRepo := repository.NewInteractionRepository(db)
	listingRepo := repository.NewListingRepository(db)
	inspectionRepo := repository.NewInspectionRepository(db)
	evaluationRepo := repository.NewEvaluationRepository(db)
	financingRepo := repository.NewFinancingRepository(db)
	insuranceRepo := repository.NewInsuranceRepository(db)
	testDriveRepo := repository.NewTestDriveRepository(db)
	transferRepo := repository.NewTransferRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	escrowRepo := repository.NewEscrowRepository(db)
	contractRepo := repository.NewContractRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)
	modelRepo := repository.NewModelRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	brandRepo := repository.NewBrandRepository(db)

	// ===== 依赖注入：Service 层（15 个） =====
	carSvc := service.NewCarService(carRepo, imageRepo)
	listingSvc := service.NewListingService(listingRepo)
	inspectionSvc := service.NewInspectionService(inspectionRepo)
	evaluationSvc := service.NewEvaluationService(evaluationRepo)
	financingSvc := service.NewFinancingService(financingRepo)
	insuranceSvc := service.NewInsuranceService(insuranceRepo)
	testDriveSvc := service.NewTestDriveService(testDriveRepo)
	transferSvc := service.NewTransferService(transferRepo)
	reportSvc := service.NewReportService(reportRepo)
	reviewSvc := service.NewReviewService(reviewRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	escrowSvc := service.NewEscrowService(escrowRepo)
	contractSvc := service.NewContractService(contractRepo)
	recommendationSvc := service.NewRecommendationService(recommendationRepo)
	statisticSvc := service.NewStatisticService(statisticRepo)

	// ===== 依赖注入：Handler 层（14 个） =====
	p.carHandler = handler.NewCarHandler(carSvc)
	p.listingHandler = handler.NewListingHandler(listingSvc)
	p.inspectionHandler = handler.NewInspectionHandler(inspectionSvc)
	p.evaluationHandler = handler.NewEvaluationHandler(evaluationSvc)
	p.financingHandler = handler.NewFinancingHandler(financingSvc)
	p.insuranceHandler = handler.NewInsuranceHandler(insuranceSvc)
	p.testDriveHandler = handler.NewTestDriveHandler(testDriveSvc)
	p.transferHandler = handler.NewTransferHandler(transferSvc)
	p.reportHandler = handler.NewReportHandler(reportSvc, reviewSvc)
	p.statisticsHandler = handler.NewStatisticsHandler(statisticSvc)
	p.tradeHandler = handler.NewTradeHandler(escrowSvc, contractSvc)
	p.recommendationHandler = handler.NewRecommendationHandler(recommendationSvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)
	p.catalogHandler = handler.NewCatalogHandler(modelRepo, categoryRepo, brandRepo)

	// 避免未使用变量告警
	_ = interactionRepo

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/car
//
// 路由分组（共 80+ API）：
//   - 公开路由（C 端浏览，无需登录）：车源/发布单/检测/评估/分期/车险/车型库/统计/评价
//   - 需登录路由（C 端发布/交易/收藏）：车源 CRUD/收藏/试驾/过户/合同/担保/评价/举报
//   - 管理后台路由（需 car:audit 权限）：审核/批量/统计/审核规则/举报处理
//
// 注意：固定路径（/search /nearby /mine /hot /by-target 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "car_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "car_write"))
	favLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "car_fav"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "car_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "car_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("car:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 车源主表（cars）
	router.GET("", readLimiter, p.carHandler.List)
	router.GET("/search", searchLimiter, p.carHandler.Search)
	router.GET("/nearby", nearbyLimiter, p.carHandler.Nearby)
	router.GET("/advanced-search", searchLimiter, p.carHandler.AdvancedSearch)
	router.GET("/:id", readLimiter, p.carHandler.GetByID)
	router.GET("/:id/fav", p.carHandler.FavStatus)

	// 发布单（listings）
	router.GET("/listings", readLimiter, p.listingHandler.List)
	router.GET("/listings/:id", readLimiter, p.listingHandler.GetByID)

	// 检测（inspections）
	router.GET("/inspections", readLimiter, p.inspectionHandler.List)
	router.GET("/inspections/:id", readLimiter, p.inspectionHandler.GetByID)
	router.GET("/cars/:id/inspection", readLimiter, p.inspectionHandler.GetByCarID)

	// 评估（evaluations）
	router.GET("/evaluations", readLimiter, p.evaluationHandler.List)
	router.GET("/evaluations/:id", readLimiter, p.evaluationHandler.GetByID)
	router.GET("/cars/:id/evaluation", readLimiter, p.evaluationHandler.GetByCarID)
	router.GET("/cars/:id/evaluations", readLimiter, p.evaluationHandler.ListByCarID)

	// 分期（financings）
	router.GET("/financings", readLimiter, p.financingHandler.ListPublished)
	router.GET("/financings/hot", readLimiter, p.financingHandler.ListHot)
	router.GET("/financings/:id", readLimiter, p.financingHandler.GetByID)

	// 车险（insurances）
	router.GET("/insurances", readLimiter, p.insuranceHandler.ListPublished)
	router.GET("/insurances/hot", readLimiter, p.insuranceHandler.ListHot)
	router.GET("/insurances/:id", readLimiter, p.insuranceHandler.GetByID)

	// 试驾（test-drives）
	router.GET("/test-drives", readLimiter, p.testDriveHandler.List)
	router.GET("/test-drives/:id", readLimiter, p.testDriveHandler.GetByID)
	router.GET("/dealers/:id/test-drives", readLimiter, p.testDriveHandler.ListByDealer)

	// 过户（transfers）
	router.GET("/transfers/:id", readLimiter, p.transferHandler.GetByID)
	router.GET("/cars/:id/transfer", readLimiter, p.transferHandler.GetByCarID)

	// 担保（escrows）
	router.GET("/cars/:id/escrows", readLimiter, p.tradeHandler.ListEscrowsByCarID)

	// 合同（contracts）
	router.GET("/cars/:id/contracts", readLimiter, p.tradeHandler.ListContractsByCarID)

	// 评价（reviews）
	router.GET("/reviews", readLimiter, p.reportHandler.ListReviews)
	router.GET("/reviews/by-target", readLimiter, p.reportHandler.ListReviewsByTarget)
	router.GET("/reviews/stats", readLimiter, p.reportHandler.ReviewStats)
	router.GET("/reviews/:id", readLimiter, p.reportHandler.GetReview)

	// 举报（reports）
	router.GET("/reports/by-target", readLimiter, p.reportHandler.ListReportsByTarget)

	// 推荐（recommendations）
	router.GET("/cars/:id/recommendations", readLimiter, p.recommendationHandler.ListByCarID)

	// 车型库（models / categories / brands）
	router.GET("/models", readLimiter, p.catalogHandler.ListModels)
	router.GET("/models/:id", readLimiter, p.catalogHandler.GetModel)
	router.GET("/brands", readLimiter, p.catalogHandler.ListBrands)
	router.GET("/brands/all", readLimiter, p.catalogHandler.ListAllBrands)
	router.GET("/brands/:brand/models", readLimiter, p.catalogHandler.ListModelsByBrand)
	router.GET("/categories", readLimiter, p.catalogHandler.ListCategories)
	router.GET("/categories/level/:level", readLimiter, p.catalogHandler.ListCategoriesByLevel)
	router.GET("/categories/:id", readLimiter, p.catalogHandler.GetCategory)
	router.GET("/categories/:id/children", readLimiter, p.catalogHandler.ListCategoriesByParent)

	// 数据统计（公开部分：热门车源/价格趋势）
	router.GET("/statistics/hot-cars", readLimiter, p.statisticsHandler.HotCars)
	router.GET("/statistics/price-trend", readLimiter, p.statisticsHandler.PriceTrend)

	// ==================== 需登录路由（C 端发布/收藏/交易） ====================

	// 车源基础 CRUD
	router.POST("", auth, writeLimiter, p.carHandler.Create)
	router.PUT("/:id", auth, p.carHandler.Update)
	router.DELETE("/:id", auth, p.carHandler.Delete)
	router.GET("/mine", auth, readLimiter, p.carHandler.ListMine)
	router.GET("/favorites", auth, readLimiter, p.carHandler.ListFavs)
	router.POST("/:id/fav", auth, favLimiter, p.carHandler.Fav)
	router.POST("/:id/contact", auth, p.carHandler.IncrContact)
	router.POST("/:id/share", auth, p.carHandler.IncrShare)
	router.POST("/:id/views", auth, p.carHandler.RecordView)

	// 发布单 CRUD
	router.POST("/listings", auth, writeLimiter, p.listingHandler.Create)
	router.PUT("/listings/:id", auth, p.listingHandler.Update)
	router.DELETE("/listings/:id", auth, p.listingHandler.Delete)
	router.GET("/listings/mine", auth, readLimiter, p.listingHandler.ListMine)

	// 检测 CRUD
	router.POST("/inspections", auth, writeLimiter, p.inspectionHandler.Create)
	router.PUT("/inspections/:id", auth, p.inspectionHandler.Update)
	router.DELETE("/inspections/:id", auth, p.inspectionHandler.Delete)
	router.GET("/inspections/mine", auth, readLimiter, p.inspectionHandler.ListByInspector)

	// 评估 CRUD + 在线估值
	router.POST("/evaluations", auth, writeLimiter, p.evaluationHandler.Create)
	router.PUT("/evaluations/:id", auth, p.evaluationHandler.Update)
	router.DELETE("/evaluations/:id", auth, p.evaluationHandler.Delete)
	router.GET("/evaluations/mine", auth, readLimiter, p.evaluationHandler.ListByEvaluator)
	router.POST("/evaluations/online", auth, writeLimiter, p.evaluationHandler.OnlineEvaluate)

	// 分期计算 + 车险报价
	router.POST("/financings/calculate", auth, writeLimiter, p.financingHandler.Calculate)
	router.POST("/insurances/quote", auth, writeLimiter, p.insuranceHandler.Quote)

	// 试驾预约 + 状态机
	router.POST("/test-drives", auth, writeLimiter, p.testDriveHandler.Create)
	router.PUT("/test-drives/:id", auth, p.testDriveHandler.Update)
	router.POST("/test-drives/:id/cancel", auth, p.testDriveHandler.Cancel)
	router.GET("/test-drives/mine", auth, readLimiter, p.testDriveHandler.ListByUser)
	router.GET("/test-drives/sales", auth, readLimiter, p.testDriveHandler.ListBySales)
	router.POST("/test-drives/:id/license", auth, writeLimiter, p.testDriveHandler.UploadLicense)

	// 过户办理 + 状态机
	router.POST("/transfers", auth, writeLimiter, p.transferHandler.Create)
	router.PUT("/transfers/:id", auth, p.transferHandler.Update)
	router.DELETE("/transfers/:id", auth, p.transferHandler.Delete)
	router.GET("/transfers", auth, readLimiter, p.transferHandler.List)
	router.GET("/transfers/sold", auth, readLimiter, p.transferHandler.ListBySeller)
	router.GET("/transfers/bought", auth, readLimiter, p.transferHandler.ListByBuyer)

	// 担保交易
	router.GET("/escrows", auth, readLimiter, p.tradeHandler.ListEscrows)
	router.GET("/escrows/:id", auth, readLimiter, p.tradeHandler.GetEscrow)
	router.GET("/escrows/no/:escrow_no", auth, readLimiter, p.tradeHandler.GetEscrowByNo)
	router.POST("/escrows/:id/action", auth, writeLimiter, p.tradeHandler.EscrowAction)

	// 合同
	router.GET("/contracts", auth, readLimiter, p.tradeHandler.ListContracts)
	router.GET("/contracts/:id", auth, readLimiter, p.tradeHandler.GetContract)
	router.GET("/contracts/no/:contract_no", auth, readLimiter, p.tradeHandler.GetContractByNo)
	router.POST("/contracts/:id/sign", auth, p.tradeHandler.SignContract)
	router.POST("/contracts/:id/terminate", auth, p.tradeHandler.TerminateContract)

	// 评价 CRUD
	router.POST("/reviews", auth, writeLimiter, p.reportHandler.CreateReview)
	router.PUT("/reviews/:id", auth, p.reportHandler.UpdateReview)
	router.DELETE("/reviews/:id", auth, p.reportHandler.DeleteReview)
	router.GET("/reviews/mine", auth, readLimiter, p.reportHandler.ListMyReviews)
	router.POST("/reviews/:id/reply", auth, writeLimiter, p.reportHandler.ReplyReview)
	router.POST("/reviews/:id/append", auth, writeLimiter, p.reportHandler.AppendReview)
	router.POST("/reviews/:id/like", auth, p.reportHandler.LikeReview)

	// 举报 + 申诉
	router.POST("/reports", auth, writeLimiter, p.reportHandler.CreateReport)
	router.GET("/reports/mine", auth, readLimiter, p.reportHandler.ListMyReports)
	router.POST("/reports/:id/appeal", auth, writeLimiter, p.reportHandler.AppealReport)

	// 推荐
	router.GET("/recommendations", auth, readLimiter, p.recommendationHandler.ListByUser)
	router.POST("/recommendations/:id/click", auth, p.recommendationHandler.MarkClicked)
	router.POST("/recommendations/:id/contact", auth, p.recommendationHandler.MarkContacted)
	router.POST("/recommendations/:id/dismiss", auth, p.recommendationHandler.MarkDismissed)

	// 卖家统计
	router.GET("/statistics/seller", auth, p.statisticsHandler.SellerOverview)

	// ==================== 管理后台路由（需 car:audit 权限） ====================

	admin := router.Group("/admin")

	// 车源管理
	admin.GET("/cars", auditPerm, p.carHandler.AdminList)
	admin.GET("/cars/:id", auditPerm, p.carHandler.AdminGetByID)
	admin.PUT("/cars/:id/audit", auditPerm, p.carHandler.Audit)
	admin.PUT("/cars/:id/status", auditPerm, p.carHandler.AdminUpdateStatus)
	admin.PUT("/cars/:id/real-car-verify", auditPerm, p.carHandler.RealCarVerify)
	admin.PUT("/cars/:id/promotion", auditPerm, p.carHandler.UpdatePromotion)

	// 发布单管理
	admin.GET("/listings", auditPerm, p.listingHandler.AdminList)
	admin.GET("/listings/:id", auditPerm, p.listingHandler.AdminGetByID)
	admin.PUT("/listings/:id/audit", auditPerm, p.listingHandler.Audit)
	admin.PUT("/listings/:id/status", auditPerm, p.listingHandler.AdminUpdateStatus)
	admin.PUT("/listings/:id/inspection-status", auditPerm, p.listingHandler.UpdateInspectionStatus)

	// 检测管理
	admin.GET("/inspections", auditPerm, p.inspectionHandler.AdminList)
	admin.GET("/inspections/:id", auditPerm, p.inspectionHandler.AdminGetByID)
	admin.PUT("/inspections/:id/review", auditPerm, p.inspectionHandler.Review)
	admin.PUT("/inspections/:id/status", auditPerm, p.inspectionHandler.AdminUpdateStatus)

	// 评估管理
	admin.GET("/evaluations", auditPerm, p.evaluationHandler.AdminList)
	admin.GET("/evaluations/:id", auditPerm, p.evaluationHandler.AdminGetByID)
	admin.PUT("/evaluations/:id/status", auditPerm, p.evaluationHandler.AdminUpdateStatus)

	// 分期方案管理
	admin.GET("/financings", auditPerm, p.financingHandler.AdminList)
	admin.GET("/financings/:id", auditPerm, p.financingHandler.AdminGetByID)
	admin.POST("/financings", auditPerm, p.financingHandler.Create)
	admin.PUT("/financings/:id", auditPerm, p.financingHandler.Update)
	admin.DELETE("/financings/:id", auditPerm, p.financingHandler.Delete)
	admin.PUT("/financings/:id/status", auditPerm, p.financingHandler.AdminUpdateStatus)

	// 车险管理
	admin.GET("/insurances", auditPerm, p.insuranceHandler.AdminList)
	admin.GET("/insurances/:id", auditPerm, p.insuranceHandler.AdminGetByID)
	admin.POST("/insurances", auditPerm, p.insuranceHandler.Create)
	admin.PUT("/insurances/:id", auditPerm, p.insuranceHandler.Update)
	admin.DELETE("/insurances/:id", auditPerm, p.insuranceHandler.Delete)
	admin.PUT("/insurances/:id/status", auditPerm, p.insuranceHandler.AdminUpdateStatus)

	// 试驾管理
	admin.GET("/test-drives", auditPerm, p.testDriveHandler.AdminList)
	admin.GET("/test-drives/:id", auditPerm, p.testDriveHandler.AdminGetByID)
	admin.PUT("/test-drives/:id/status", auditPerm, p.testDriveHandler.UpdateStatus)

	// 过户管理
	admin.GET("/transfers", auditPerm, p.transferHandler.AdminList)
	admin.GET("/transfers/:id", auditPerm, p.transferHandler.AdminGetByID)
	admin.PUT("/transfers/:id/status", auditPerm, p.transferHandler.UpdateStatus)

	// 担保交易管理
	admin.GET("/escrows", auditPerm, p.tradeHandler.AdminListEscrows)
	admin.GET("/escrows/:id", auditPerm, p.tradeHandler.AdminGetEscrow)
	admin.PUT("/escrows/:id/status", auditPerm, p.tradeHandler.UpdateEscrowStatus)

	// 合同管理
	admin.GET("/contracts", auditPerm, p.tradeHandler.AdminListContracts)
	admin.GET("/contracts/:id", auditPerm, p.tradeHandler.AdminGetContract)
	admin.PUT("/contracts/:id/status", auditPerm, p.tradeHandler.UpdateContractStatus)

	// 评价管理
	admin.GET("/reviews", auditPerm, p.reportHandler.AdminListReviews)
	admin.PUT("/reviews/:id/status", auditPerm, p.reportHandler.UpdateReviewStatus)

	// 举报管理
	admin.GET("/reports", auditPerm, p.reportHandler.ListReports)
	admin.GET("/reports/:id", auditPerm, p.reportHandler.GetReport)
	admin.PUT("/reports/:id/process", auditPerm, p.reportHandler.ProcessReport)
	admin.PUT("/reports/:id/appeal", auditPerm, p.reportHandler.ProcessAppeal)
	admin.PUT("/reports/:id/status", auditPerm, p.reportHandler.UpdateReportStatus)

	// 推荐管理
	admin.GET("/recommendations", auditPerm, p.recommendationHandler.AdminList)
	admin.DELETE("/recommendations/:id", auditPerm, p.recommendationHandler.Delete)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.List)
	admin.GET("/audit-rules/:id", auditPerm, p.auditRuleHandler.GetByID)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.UpdateStatus)
	admin.POST("/audit-rules/check", auditPerm, p.auditRuleHandler.Check)

	// 平台总览统计
	admin.GET("/statistics/overview", auditPerm, p.statisticsHandler.Overview)
	admin.GET("/statistics", auditPerm, p.statisticsHandler.AdminList)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
