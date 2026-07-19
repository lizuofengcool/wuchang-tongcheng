// Package house 同城房屋租售模块插件
// 提供房源发布/浏览/搜索/收藏/经纪/小区/合同/看房/举报/评价/统计/担保/成交/VR/房贷/分类/设施/审核规则等业务
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家/安居客/我爱我家/58房产
package house

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/house/handler"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/modules/house/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 房屋租售模块插件
type Plugin struct {
	name    string
	version string

	// 8 个 Handler
	houseHandler      *handler.HouseHandler
	listingHandler    *handler.ListingHandler
	communityHandler  *handler.CommunityHandler
	agentHandler      *handler.AgentHandler
	contractHandler   *handler.ContractHandler
	viewingHandler    *handler.ViewingHandler
	reportHandler     *handler.ReportHandler
	statisticsHandler *handler.StatisticsHandler
}

// NewPlugin 创建房屋租售模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "house", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "house",
		DisplayName:  "同城房屋租售",
		Category:     "business",
		Description:  "同城房屋租售完整功能：房源/房源发布/小区/经纪人/合同/看房/举报/评价/统计/担保交易/成交/VR看房/房贷/分类/配套设施/审核规则",
		Version:      p.version,
		Dependencies: []string{},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：22 张表（4 复用 + 18 专属）由 backend/migrations/004_house_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致导致的 DROP CONSTRAINT 错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 houses 主表（子表由 backend/migrations/007_house_full.sql 创建）
	// 注：之前未调用 AutoMigrate 导致 houses 主表缺失，前端列表报 "relation houses does not exist"
	if err := db.AutoMigrate(&model.House{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层 =====
	houseRepo := repository.NewHouseRepository(db)
	listingRepo := repository.NewListingRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	contractRepo := repository.NewContractRepository(db)
	viewingRepo := repository.NewViewingRepository(db)
	riskRepo := repository.NewRiskRepository(db)
	interactionRepo := repository.NewInteractionRepository(db)
	escrowRepo := repository.NewEscrowRepository(db)
	dealRepo := repository.NewDealRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	facilityRepo := repository.NewFacilityRepository(db)
	mortgageRepo := repository.NewMortgageRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)

	// ===== 依赖注入：Service 层 =====
	houseSvc := service.NewHouseService(houseRepo, agentRepo, communityRepo)
	listingSvc := service.NewListingService(listingRepo)
	communitySvc := service.NewCommunityService(communityRepo)
	agentSvc := service.NewAgentService(agentRepo)
	contractSvc := service.NewContractService(contractRepo)
	viewingSvc := service.NewViewingService(viewingRepo)
	reportSvc := service.NewReportService(riskRepo)
	reviewSvc := service.NewReviewService(riskRepo)
	statSvc := service.NewStatisticService(
		statisticRepo, houseRepo, listingRepo, agentRepo, communityRepo,
		dealRepo, viewingRepo, riskRepo,
	)
	mortgSvc := service.NewMortgageService(mortgageRepo)
	escrowSvc := service.NewEscrowService(escrowRepo)
	dealSvc := service.NewDealService(dealRepo)
	recSvc := service.NewRecommendationService(interactionRepo)
	vrSvc := service.NewVRTourService(interactionRepo)
	catSvc := service.NewCategoryService(categoryRepo)
	facSvc := service.NewFacilityService(facilityRepo)
	ruleSvc := service.NewAuditRuleService(riskRepo)

	// ===== 依赖注入：Handler 层 =====
	p.houseHandler = handler.NewHouseHandler(houseSvc)
	p.listingHandler = handler.NewListingHandler(listingSvc)
	p.communityHandler = handler.NewCommunityHandler(communitySvc)
	p.agentHandler = handler.NewAgentHandler(agentSvc)
	p.contractHandler = handler.NewContractHandler(contractSvc)
	p.viewingHandler = handler.NewViewingHandler(viewingSvc)
	p.reportHandler = handler.NewReportHandler(reportSvc, reviewSvc)
	p.statisticsHandler = handler.NewStatisticsHandler(
		statSvc, mortgSvc, escrowSvc, dealSvc, recSvc, vrSvc,
		catSvc, facSvc, ruleSvc,
	)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/house
//
// 路由分组（共 80+ API）：
//   - 公开路由（C 端浏览，无需登录）：房源列表/搜索/附近/详情/相似推荐/小区/经纪人/分类/设施/VR/统计/评价/举报
//   - 需登录路由（C 端发布/交易/收藏）：房源 CRUD/收藏/互动/推广/房源发布/小区/经纪人/合同/看房/担保/成交/推荐/VR/评价/举报
//   - 管理后台路由（需 house:audit 权限）：房源审核/批量操作/统计/审核规则/分类/设施/房贷/担保仲裁/成交完成/评价管理/举报处理
//
// 注意：固定路径（/search /nearby /advanced-search /mine /favorites /listings /communities /agents /contracts /viewings 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "house_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "house_write"))
	favLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "house_fav"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "house_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "house_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("house:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 房源主表
	router.GET("", readLimiter, p.houseHandler.List)
	router.GET("/search", searchLimiter, p.houseHandler.Search)
	router.GET("/advanced-search", searchLimiter, p.houseHandler.AdvancedSearch)
	router.GET("/nearby", nearbyLimiter, p.houseHandler.ListNearby)
	router.GET("/:id", readLimiter, p.houseHandler.GetByID)
	router.GET("/:id/similar", readLimiter, p.houseHandler.ListSimilar)
	router.GET("/:id/fav", p.houseHandler.FavStatus)
	router.POST("/:id/contact", readLimiter, p.houseHandler.IncrContactCount)
	router.POST("/:id/share", readLimiter, p.houseHandler.IncrShareCount)
	// 房源关联资源
	router.GET("/:id/reviews", readLimiter, p.reportHandler.ListReviewsByTarget)
	router.GET("/:id/reviews/stats", readLimiter, p.reportHandler.ReviewStats)
	router.GET("/:id/reports", readLimiter, p.reportHandler.ListReportsByTarget)
	router.GET("/:id/escrows", readLimiter, p.statisticsHandler.ListEscrowsByHouse)
	router.GET("/:id/deals", readLimiter, p.statisticsHandler.ListDealsByHouse)
	// 通用目标关联资源（评价/举报）
	// 注意：使用 /by-target/ 前缀避免与 /:id 路由参数冲突（Gin 不允许同层级不同通配符名）
	router.GET("/by-target/:target_type/:target_id/reviews", readLimiter, p.reportHandler.ListReviewsByTarget)
	router.GET("/by-target/:target_type/:target_id/reviews/stats", readLimiter, p.reportHandler.ReviewStats)
	router.GET("/by-target/:target_type/:target_id/reports", readLimiter, p.reportHandler.ListReportsByTarget)

	// 房源发布（listing）
	router.GET("/listings", readLimiter, p.listingHandler.List)

	// 小区
	router.GET("/communities", readLimiter, p.communityHandler.List)
	router.GET("/communities/nearby", nearbyLimiter, p.communityHandler.ListNearby)
	router.GET("/communities/:id", readLimiter, p.communityHandler.GetByID)

	// 经纪人
	router.GET("/agents", readLimiter, p.agentHandler.List)
	router.GET("/agents/:id", readLimiter, p.agentHandler.GetByID)

	// 评价
	router.GET("/reviews", readLimiter, p.reportHandler.ListReviews)
	router.GET("/reviews/:id", readLimiter, p.reportHandler.GetReview)

	// 担保交易（公开查询房源关联的担保）

	// 成交记录（公开查询）
	router.GET("/deals", readLimiter, p.statisticsHandler.ListDeals)

	// VR 看房（公开浏览）
	router.GET("/vr-tours", readLimiter, p.statisticsHandler.ListVRTours)
	router.GET("/vr-tours/:id", readLimiter, p.statisticsHandler.GetVRTour)
	router.POST("/vr-tours/:id/share", readLimiter, p.statisticsHandler.ShareVRTour)

	// 房贷方案（公开查询 + 计算）
	router.GET("/mortgages", readLimiter, p.statisticsHandler.ListMortgages)
	router.GET("/mortgages/:id", readLimiter, p.statisticsHandler.GetMortgage)
	router.POST("/mortgages/calculate", readLimiter, p.statisticsHandler.CalculateMortgage)

	// 房源分类（公开查询）
	router.GET("/categories", readLimiter, p.statisticsHandler.ListCategories)
	router.GET("/categories/all", readLimiter, p.statisticsHandler.ListAllCategories)
	router.GET("/categories/parent/:parent_id", readLimiter, p.statisticsHandler.ListCategoriesByParent)
	router.GET("/categories/:id", readLimiter, p.statisticsHandler.GetCategory)

	// 配套设施（公开查询）
	router.GET("/facilities", readLimiter, p.statisticsHandler.ListFacilities)
	router.GET("/facilities/all", readLimiter, p.statisticsHandler.ListAllFacilities)
	router.GET("/facilities/hot", readLimiter, p.statisticsHandler.ListHotFacilities)
	router.GET("/facilities/:id", readLimiter, p.statisticsHandler.GetFacility)

	// 审核规则（启用列表，C 端发布时使用）
	router.GET("/audit-rules/enabled", readLimiter, p.statisticsHandler.ListEnabledAuditRules)

	// 数据统计（公开部分：价格趋势）
	router.GET("/statistics/price-trend", readLimiter, p.statisticsHandler.PriceTrend)

	// ==================== 需登录路由（C 端发布/收藏/交易） ====================

	// 房源基础 CRUD
	router.POST("", auth, writeLimiter, p.houseHandler.Create)
	router.PUT("/:id", auth, p.houseHandler.Update)
	router.DELETE("/:id", auth, p.houseHandler.Delete)
	router.GET("/mine", auth, p.houseHandler.ListMine)
	router.GET("/favorites", auth, readLimiter, p.houseHandler.ListFavs)
	router.POST("/:id/fav", auth, favLimiter, p.houseHandler.Fav)
	router.PUT("/:id/promotion", auth, p.houseHandler.UpdatePromotion)

	// 房源发布（listing）管理
	router.POST("/listings", auth, writeLimiter, p.listingHandler.Create)
	router.PUT("/listings/:id", auth, p.listingHandler.Update)
	router.DELETE("/listings/:id", auth, p.listingHandler.Delete)
	router.GET("/listings/:id", auth, p.listingHandler.GetByID)
	router.GET("/listings/mine", auth, p.listingHandler.ListMine)
	router.POST("/listings/:id/refresh", auth, p.listingHandler.Refresh)

	// 小区管理
	router.POST("/communities", auth, writeLimiter, p.communityHandler.Create)
	router.PUT("/communities/:id", auth, p.communityHandler.Update)
	router.POST("/communities/:id/follow", auth, favLimiter, p.communityHandler.Follow)
	router.GET("/communities/:id/follow", auth, p.communityHandler.FollowStatus)

	// 经纪人管理
	router.POST("/agents", auth, writeLimiter, p.agentHandler.Create)
	router.PUT("/agents/:id", auth, p.agentHandler.Update)
	router.GET("/agents/mine", auth, p.agentHandler.GetMine)
	router.POST("/agents/:id/follow", auth, favLimiter, p.agentHandler.Follow)
	router.GET("/agents/:id/follow", auth, p.agentHandler.FollowStatus)

	// 合同管理
	router.POST("/contracts", auth, writeLimiter, p.contractHandler.Create)
	router.PUT("/contracts/:id", auth, p.contractHandler.Update)
	router.GET("/contracts", auth, readLimiter, p.contractHandler.List)
	router.GET("/contracts/mine", auth, p.contractHandler.ListMine)
	router.GET("/contracts/:id", auth, p.contractHandler.GetByID)
	router.POST("/contracts/:id/sign", auth, p.contractHandler.Sign)
	router.POST("/contracts/:id/terminate", auth, p.contractHandler.Terminate)

	// 看房预约管理
	router.POST("/viewings", auth, writeLimiter, p.viewingHandler.Create)
	router.PUT("/viewings/:id", auth, p.viewingHandler.Update)
	router.GET("/viewings", auth, readLimiter, p.viewingHandler.List)
	router.GET("/viewings/mine", auth, p.viewingHandler.ListMine)
	router.GET("/viewings/:id", auth, p.viewingHandler.GetByID)
	router.POST("/viewings/:id/confirm", auth, p.viewingHandler.Confirm)
	router.POST("/viewings/:id/cancel", auth, p.viewingHandler.Cancel)
	router.POST("/viewings/:id/reschedule", auth, p.viewingHandler.Reschedule)
	router.POST("/viewings/:id/complete", auth, p.viewingHandler.Complete)

	// 评价管理（C 端）
	router.POST("/reviews", auth, writeLimiter, p.reportHandler.CreateReview)
	router.POST("/reviews/:id/reply", auth, writeLimiter, p.reportHandler.ReplyReview)
	router.POST("/reviews/:id/append", auth, writeLimiter, p.reportHandler.AppendReview)
	router.POST("/reviews/:id/like", auth, p.reportHandler.LikeReview)
	router.GET("/reviews/mine", auth, p.reportHandler.ListMyReviews)

	// 举报管理（C 端）
	router.POST("/reports", auth, writeLimiter, p.reportHandler.CreateReport)
	router.GET("/reports", auth, readLimiter, p.reportHandler.ListReports)
	router.GET("/reports/mine", auth, p.reportHandler.ListMyReports)
	router.GET("/reports/:id", auth, p.reportHandler.GetReport)
	router.GET("/reports/no/:no", auth, p.reportHandler.GetReportByNo)
	router.POST("/reports/:id/appeal", auth, p.reportHandler.AppealReport)

	// 担保交易（C 端）
	router.POST("/escrows", auth, writeLimiter, p.statisticsHandler.CreateEscrow)
	router.GET("/escrows", auth, readLimiter, p.statisticsHandler.ListEscrows)
	router.GET("/escrows/:id", auth, p.statisticsHandler.GetEscrow)
	router.GET("/escrows/mine-payer", auth, p.statisticsHandler.ListMyPayerEscrows)
	router.GET("/escrows/mine-payee", auth, p.statisticsHandler.ListMyPayeeEscrows)
	router.POST("/escrows/:id/pay", auth, p.statisticsHandler.MarkEscrowPaid)
	router.POST("/escrows/:id/release", auth, p.statisticsHandler.ReleaseEscrow)
	router.POST("/escrows/:id/refund", auth, p.statisticsHandler.RefundEscrow)
	router.POST("/escrows/:id/dispute", auth, p.statisticsHandler.DisputeEscrow)
	router.POST("/escrows/:id/cancel", auth, p.statisticsHandler.CancelEscrow)

	// 成交记录（C 端）
	router.POST("/deals", auth, writeLimiter, p.statisticsHandler.CreateDeal)
	router.GET("/deals/:id", auth, p.statisticsHandler.GetDeal)
	router.POST("/deals/:id/confirm", auth, p.statisticsHandler.ConfirmDeal)
	router.POST("/deals/:id/cancel", auth, p.statisticsHandler.CancelDeal)

	// 推荐（C 端）
	router.GET("/recommendations/mine", auth, p.statisticsHandler.ListMyRecommendations)
	router.GET("/recommendations/:id", auth, p.statisticsHandler.GetRecommendation)
	router.POST("/recommendations/:id/click", auth, p.statisticsHandler.MarkRecClicked)
	router.POST("/recommendations/:id/contact", auth, p.statisticsHandler.MarkRecContacted)
	router.POST("/recommendations/:id/view", auth, p.statisticsHandler.MarkRecViewed)
	router.POST("/recommendations/:id/dismiss", auth, p.statisticsHandler.MarkRecDismissed)

	// VR 看房管理（C 端）
	router.POST("/vr-tours", auth, writeLimiter, p.statisticsHandler.CreateVRTour)
	router.PUT("/vr-tours/:id", auth, p.statisticsHandler.UpdateVRTour)
	router.DELETE("/vr-tours/:id", auth, p.statisticsHandler.DeleteVRTour)
	router.POST("/vr-tours/:id/publish", auth, p.statisticsHandler.PublishVRTour)
	router.POST("/vr-tours/:id/offline", auth, p.statisticsHandler.OfflineVRTour)

	// 数据统计 - 平台总览（登录用户可看汇总）
	router.GET("/statistics/overview", auth, p.statisticsHandler.Overview)

	// ==================== 管理后台路由（需 house:audit 权限） ====================

	// 房源管理
	admin := router.Group("/admin")
	admin.GET("/list", auditPerm, p.houseHandler.AdminList)
	admin.GET("/:id", auditPerm, p.houseHandler.AdminGetByID)
	admin.PUT("/:id/audit", auditPerm, p.houseHandler.Audit)
	admin.PUT("/:id/status", auditPerm, p.houseHandler.AdminUpdateStatus)
	admin.POST("/batch/audit", auditPerm, p.houseHandler.BatchAudit)
	admin.POST("/batch/status", auditPerm, p.houseHandler.BatchUpdateStatus)
	admin.POST("/batch/delete", auditPerm, p.houseHandler.BatchDelete)

	// 房源发布审核
	router.GET("/admin/listings", auth, auditPerm, p.listingHandler.AdminList)
	router.PUT("/admin/listings/:id/audit", auth, auditPerm, p.listingHandler.Audit)
	router.PUT("/admin/listings/:id/status", auth, auditPerm, p.listingHandler.UpdateStatus)

	// 小区审核
	router.GET("/admin/communities", auth, auditPerm, p.communityHandler.AdminList)
	router.PUT("/admin/communities/:id/status", auth, auditPerm, p.communityHandler.UpdateStatus)

	// 经纪人审核
	router.GET("/admin/agents", auth, auditPerm, p.agentHandler.AdminList)
	router.PUT("/admin/agents/:id/audit", auth, auditPerm, p.agentHandler.Audit)
	router.PUT("/admin/agents/:id/online-status", auth, auditPerm, p.agentHandler.UpdateOnlineStatus)

	// 合同管理
	router.GET("/admin/contracts", auth, auditPerm, p.contractHandler.AdminList)

	// 看房预约管理
	router.GET("/admin/viewings", auth, auditPerm, p.viewingHandler.AdminList)

	// 评价管理
	router.GET("/admin/reviews", auth, auditPerm, p.reportHandler.AdminListReviews)
	router.PUT("/admin/reviews/:id/status", auth, auditPerm, p.reportHandler.UpdateReviewStatus)
	router.POST("/admin/reviews/batch", auth, auditPerm, p.reportHandler.BatchUpdateReviewStatus)
	router.DELETE("/admin/reviews/:id", auth, auditPerm, p.reportHandler.DeleteReview)

	// 举报管理
	router.GET("/admin/reports", auth, auditPerm, p.reportHandler.AdminListReports)
	router.GET("/admin/reports/:id", auth, auditPerm, p.reportHandler.GetReport)
	router.PUT("/admin/reports/:id/process", auth, auditPerm, p.reportHandler.ProcessReport)
	router.PUT("/admin/reports/:id/appeal", auth, auditPerm, p.reportHandler.AppealHandleReport)
	router.POST("/admin/reports/batch", auth, auditPerm, p.reportHandler.BatchUpdateReportStatus)
	router.GET("/admin/reports/pending-count", auth, auditPerm, p.reportHandler.CountPendingReports)

	// 担保交易管理
	router.GET("/admin/escrows", auth, auditPerm, p.statisticsHandler.AdminListEscrows)
	router.GET("/admin/escrows/disputed", auth, auditPerm, p.statisticsHandler.ListDisputedEscrows)
	router.PUT("/admin/escrows/:id/arbitrate", auth, auditPerm, p.statisticsHandler.ArbitrateEscrow)

	// 成交记录管理
	router.GET("/admin/deals", auth, auditPerm, p.statisticsHandler.AdminListDeals)
	router.POST("/admin/deals/:id/complete", auth, auditPerm, p.statisticsHandler.CompleteDeal)

	// 数据统计管理
	router.GET("/admin/statistics", auth, auditPerm, p.statisticsHandler.ListStats)
	router.GET("/admin/statistics/by-type", auth, auditPerm, p.statisticsHandler.ListStatsByType)
	router.GET("/admin/statistics/:id", auth, auditPerm, p.statisticsHandler.GetStat)

	// 房贷方案管理
	router.POST("/admin/mortgages", auth, auditPerm, p.statisticsHandler.CreateMortgage)
	router.PUT("/admin/mortgages/:id", auth, auditPerm, p.statisticsHandler.UpdateMortgage)
	router.DELETE("/admin/mortgages/:id", auth, auditPerm, p.statisticsHandler.DeleteMortgage)
	router.PUT("/admin/mortgages/:id/status", auth, auditPerm, p.statisticsHandler.UpdateMortgageStatus)

	// 房源分类管理
	router.POST("/admin/categories", auth, auditPerm, p.statisticsHandler.CreateCategory)
	router.PUT("/admin/categories/:id", auth, auditPerm, p.statisticsHandler.UpdateCategory)
	router.DELETE("/admin/categories/:id", auth, auditPerm, p.statisticsHandler.DeleteCategory)
	router.PUT("/admin/categories/:id/status", auth, auditPerm, p.statisticsHandler.UpdateCategoryStatus)

	// 配套设施管理
	router.POST("/admin/facilities", auth, auditPerm, p.statisticsHandler.CreateFacility)
	router.PUT("/admin/facilities/:id", auth, auditPerm, p.statisticsHandler.UpdateFacility)
	router.DELETE("/admin/facilities/:id", auth, auditPerm, p.statisticsHandler.DeleteFacility)
	router.PUT("/admin/facilities/:id/status", auth, auditPerm, p.statisticsHandler.UpdateFacilityStatus)

	// 审核规则管理
	router.POST("/admin/audit-rules", auth, auditPerm, p.statisticsHandler.CreateAuditRule)
	router.GET("/admin/audit-rules", auth, auditPerm, p.statisticsHandler.ListAuditRules)
	router.GET("/admin/audit-rules/:id", auth, auditPerm, p.statisticsHandler.GetAuditRule)
	router.PUT("/admin/audit-rules/:id", auth, auditPerm, p.statisticsHandler.UpdateAuditRule)
	router.DELETE("/admin/audit-rules/:id", auth, auditPerm, p.statisticsHandler.DeleteAuditRule)
	router.PUT("/admin/audit-rules/:id/status", auth, auditPerm, p.statisticsHandler.UpdateAuditRuleStatus)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
