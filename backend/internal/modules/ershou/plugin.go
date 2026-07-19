// Package ershou 同城二手物品模块插件
// 提供二手物品的发布、浏览、搜索、附近、留言、收藏及管理后台审核等业务
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标闲鱼/转转/瓜子/贝壳/58同城
package ershou

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/ershou/handler"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 二手物品模块插件
type Plugin struct {
	name    string
	version string

	// 9 个 Handler
	handler          *handler.Handler
	skuHandler       *handler.SKUHandler
	orderHandler     *handler.OrderHandler
	auctionHandler   *handler.AuctionHandler
	tradeHandler     *handler.TradeHandler
	riskHandler      *handler.RiskHandler
	shopHandler      *handler.ShopHandler
	catalogHandler   *handler.CatalogHandler
	statsHandler     *handler.StatisticsHandler
	batchHandler     *handler.BatchHandler
}

// NewPlugin 创建二手物品模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "ershou", version: "2.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "ershou",
		DisplayName:  "同城二手",
		Category:     "business",
		Description:  "同城二手交易完整功能：商品/SKU/订单/拍卖/担保/物流/退款/评价/举报/店铺/标签/品牌/型号/分类属性/审核规则/用户信用/统计/批量/导出",
		Version:      p.version,
		Dependencies: []string{"category"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：23 张表（4 基础 + 19 扩展）由 backend/migrations/003_ershou_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则（如 ers_orders_order_no_key），
// 不使用 GORM AutoMigrate 以避免约束名不一致导致的 DROP CONSTRAINT 错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// ===== 依赖注入：Repository 层 =====
	ershouRepo := repository.NewErshouRepository(db)
	skuRepo := repository.NewSKURepository(db)
	orderRepo := repository.NewOrderRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	promotionRepo := repository.NewPromotionRepository(db)
	logisticsRepo := repository.NewLogisticsRepository(db)
	escrowRepo := repository.NewEscrowRepository(db)
	refundRepo := repository.NewRefundRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	userCreditRepo := repository.NewUserCreditRepository(db)
	shopRepo := repository.NewShopRepository(db)
	shopFollowerRepo := repository.NewShopFollowerRepository(db)
	tagRepo := repository.NewTagRepository(db)
	brandRepo := repository.NewBrandRepository(db)
	modelRepo := repository.NewModelRepository(db)
	categoryAttrRepo := repository.NewCategoryAttrRepository(db)

	// ===== 依赖注入：Service 层 =====
	ershouSvc := service.NewErshouService(ershouRepo)
	skuSvc := service.NewSKUService(skuRepo, ershouRepo)
	orderSvc := service.NewOrderService(orderRepo, ershouRepo, skuRepo, escrowRepo)
	auctionSvc := service.NewAuctionService(auctionRepo, ershouRepo)
	promotionSvc := service.NewPromotionService(promotionRepo, ershouRepo)
	logisticsSvc := service.NewLogisticsService(logisticsRepo, orderRepo)
	escrowSvc := service.NewEscrowService(escrowRepo, orderRepo)
	refundSvc := service.NewRefundService(refundRepo, orderRepo, escrowRepo)
	reportSvc := service.NewReportService(reportRepo, ershouRepo, userCreditRepo)
	reviewSvc := service.NewReviewService(reviewRepo, orderRepo, userCreditRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	userCreditSvc := service.NewUserCreditService(userCreditRepo)
	shopSvc := service.NewShopService(shopRepo, shopFollowerRepo)
	tagSvc := service.NewTagService(tagRepo)
	brandSvc := service.NewBrandService(brandRepo)
	modelSvc := service.NewModelService(modelRepo)
	categoryAttrSvc := service.NewCategoryAttrService(categoryAttrRepo)
	statsSvc := service.NewStatisticsService(db)
	batchSvc := service.NewBatchService(db, ershouRepo)

	// ===== 依赖注入：Handler 层 =====
	p.handler = handler.NewHandler(ershouSvc)
	p.skuHandler = handler.NewSKUHandler(skuSvc)
	p.orderHandler = handler.NewOrderHandler(orderSvc)
	p.auctionHandler = handler.NewAuctionHandler(auctionSvc)
	p.tradeHandler = handler.NewTradeHandler(promotionSvc, logisticsSvc, escrowSvc, refundSvc)
	p.riskHandler = handler.NewRiskHandler(reportSvc, reviewSvc, auditRuleSvc, userCreditSvc)
	p.shopHandler = handler.NewShopHandler(shopSvc)
	p.catalogHandler = handler.NewCatalogHandler(tagSvc, brandSvc, modelSvc, categoryAttrSvc)
	p.statsHandler = handler.NewStatisticsHandler(statsSvc)
	p.batchHandler = handler.NewBatchHandler(batchSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/ershou
//
// 路由分组（共 100+ API）：
//   - 公开路由（C 端浏览，无需登录）：商品列表/搜索/附近/详情/留言/SKU/拍卖/推广/评价/店铺/标签/品牌/型号/分类属性/统计
//   - 需登录路由（C 端发布/交易/收藏）：商品 CRUD/收藏/留言/SKU/拍卖/推广/订单/物流/担保/退款/评价/举报/店铺/关注/信用
//   - 管理后台路由（需 content:audit 权限）：商品审核/批量/导出/统计/审核规则/店铺审核/用户信用/举报处理
//
// 注意：固定路径（/search /nearby /favorites /mine /orders /shops /tags /brands /models 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "ershou_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "ershou_write"))
	favLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_fav"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("content:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 商品基础（ershou 主表）
	router.GET("", readLimiter, p.handler.List)
	router.GET("/search", searchLimiter, p.handler.Search)
	router.GET("/nearby", nearbyLimiter, p.handler.Nearby)
	router.GET("/:id", readLimiter, p.handler.GetByID)
	router.GET("/:id/messages", readLimiter, p.handler.ListMessages)
	router.GET("/:id/fav", p.handler.FavStatus)

	// 商品 SKU
	router.GET("/:id/skus", readLimiter, p.skuHandler.List)

	// 商品拍卖
	router.GET("/:id/auction", readLimiter, p.auctionHandler.GetByErshouID)
	router.GET("/auctions", readLimiter, p.auctionHandler.List)

	// 商品推广
	router.GET("/:id/promotions", readLimiter, p.tradeHandler.ListPromotions)
	router.GET("/:id/promotions/stats", readLimiter, p.tradeHandler.PromotionStats)

	// 商品评价
	router.GET("/:id/reviews", readLimiter, p.riskHandler.ListReviewsByErshouID)
	router.GET("/:id/reviews/stats", readLimiter, p.riskHandler.ReviewStats)
	router.GET("/reviews", readLimiter, p.riskHandler.ListReviews)
	router.GET("/reviews/:id", readLimiter, p.riskHandler.GetReview)

	// 商品举报
	router.GET("/:id/reports", readLimiter, p.riskHandler.ListReportsByErshouID)

	// 店铺
	router.GET("/shops", readLimiter, p.shopHandler.List)
	router.GET("/shops/:id", readLimiter, p.shopHandler.GetByID)
	router.GET("/shops/:id/followers", readLimiter, p.shopHandler.ListFollowers)

	// 标签
	router.GET("/tags", readLimiter, p.catalogHandler.ListTags)
	router.GET("/tags/hot", readLimiter, p.catalogHandler.ListHotTags)
	router.GET("/tags/:id", readLimiter, p.catalogHandler.GetTag)

	// 品牌
	router.GET("/brands", readLimiter, p.catalogHandler.ListBrands)
	router.GET("/brands/:id", readLimiter, p.catalogHandler.GetBrand)
	router.GET("/brands/:id/models", readLimiter, p.catalogHandler.ListModelsByBrandID)

	// 型号
	router.GET("/models", readLimiter, p.catalogHandler.ListModels)
	router.GET("/models/:id", readLimiter, p.catalogHandler.GetModel)

	// 分类属性
	router.GET("/category-attrs", readLimiter, p.catalogHandler.ListCategoryAttrs)
	router.GET("/category-attrs/:id", readLimiter, p.catalogHandler.GetCategoryAttr)
	router.GET("/categories/:id/attrs", readLimiter, p.catalogHandler.ListCategoryAttrsByCategoryID)

	// 审核规则（启用列表，C 端发布时使用）
	router.GET("/audit-rules/enabled", readLimiter, p.riskHandler.ListEnabledAuditRules)

	// 数据统计（公开部分：热门商品/价格趋势）
	router.GET("/statistics/hot-items", readLimiter, p.statsHandler.HotItems)
	router.GET("/statistics/price-trend", readLimiter, p.statsHandler.PriceTrend)

	// ==================== 需登录路由（C 端发布/收藏/留言） ====================

	// 商品基础 CRUD
	router.POST("", auth, writeLimiter, p.handler.Create)
	router.PUT("/:id", auth, p.handler.Update)
	router.DELETE("/:id", auth, p.handler.Delete)
	router.GET("/mine", auth, p.handler.ListMine)
	router.GET("/favorites", auth, readLimiter, p.handler.ListFavs)
	router.POST("/:id/fav", auth, favLimiter, p.handler.Fav)
	router.POST("/:id/messages", auth, writeLimiter, p.handler.CreateMessage)

	// SKU 管理（仅发布者本人）
	router.POST("/:id/skus", auth, writeLimiter, p.skuHandler.Create)
	router.PUT("/:id/skus/:sku_id", auth, p.skuHandler.Update)
	router.DELETE("/:id/skus/:sku_id", auth, p.skuHandler.Delete)

	// 拍卖管理
	router.POST("/:id/auction", auth, writeLimiter, p.auctionHandler.Create)
	router.POST("/:id/auction/bid", auth, writeLimiter, p.auctionHandler.Bid)
	router.POST("/:id/auction/end", auth, p.auctionHandler.EndManually)

	// 推广管理
	router.POST("/:id/promotions", auth, writeLimiter, p.tradeHandler.CreatePromotion)

	// 评价
	router.POST("/orders/:id/reviews", auth, writeLimiter, p.riskHandler.CreateReview)
	router.POST("/reviews/:id/reply", auth, writeLimiter, p.riskHandler.ReplyReview)

	// 举报
	router.POST("/reports", auth, writeLimiter, p.riskHandler.CreateReport)

	// 店铺
	router.POST("/shops", auth, writeLimiter, p.shopHandler.Create)
	router.GET("/shops/mine", auth, p.shopHandler.GetMyShop)
	router.PUT("/shops/:id", auth, p.shopHandler.Update)
	router.POST("/shops/:id/follow", auth, favLimiter, p.shopHandler.Follow)
	router.DELETE("/shops/:id/follow", auth, p.shopHandler.Unfollow)
	router.GET("/shops/following", auth, readLimiter, p.shopHandler.ListUserFollowing)

	// 用户信用
	router.GET("/credit", auth, p.riskHandler.GetUserCredit)

	// 卖家统计
	router.GET("/statistics/seller", auth, p.statsHandler.SellerOverview)

	// ==================== 订单路由（需登录） ====================

	router.POST("/orders", auth, writeLimiter, p.orderHandler.Create)
	router.GET("/orders/:id", auth, p.orderHandler.GetByID)
	router.GET("/orders", auth, readLimiter, p.orderHandler.List)
	router.PUT("/orders/:id/status", auth, p.orderHandler.UpdateStatus)
	router.POST("/orders/:id/pay", auth, p.orderHandler.Pay)
	router.POST("/orders/:id/ship", auth, p.orderHandler.Ship)
	router.POST("/orders/:id/receive", auth, p.orderHandler.Receive)
	router.POST("/orders/:id/cancel", auth, p.orderHandler.Cancel)

	// ==================== 交易路由：物流/担保/退款（需登录） ====================

	// 物流
	router.POST("/orders/:id/logistics", auth, writeLimiter, p.tradeHandler.CreateLogistics)
	router.GET("/orders/:id/logistics", auth, p.tradeHandler.GetLogistics)
	router.PUT("/orders/:id/logistics", auth, p.tradeHandler.UpdateLogistics)

	// 担保
	router.POST("/orders/:id/escrow", auth, p.tradeHandler.CreateEscrow)
	router.GET("/orders/:id/escrow", auth, p.tradeHandler.GetEscrow)
	router.POST("/orders/:id/escrow/release", auth, p.tradeHandler.ReleaseEscrow)

	// 退款
	router.POST("/orders/:id/refund", auth, writeLimiter, p.tradeHandler.CreateRefund)
	router.GET("/orders/:id/refund", auth, p.tradeHandler.GetRefund)
	router.PUT("/refunds/:id/process", auth, p.tradeHandler.ProcessRefund)
	router.GET("/refunds", auth, readLimiter, p.tradeHandler.ListRefunds)

	// ==================== 管理后台路由（需 content:audit 权限） ====================

	// 商品管理
	admin := router.Group("/admin")
	admin.GET("/list", auditPerm, p.handler.AdminList)
	admin.GET("/:id", auditPerm, p.handler.AdminGetByID)
	admin.PUT("/:id/audit", auditPerm, p.handler.Audit)
	admin.PUT("/:id/status", auditPerm, p.handler.AdminUpdateStatus)

	// 平台总览统计
	router.GET("/statistics/overview", auth, auditPerm, p.statsHandler.Overview)

	// 批量操作
	router.POST("/batch/audit", auth, auditPerm, p.batchHandler.Audit)
	router.POST("/batch/status", auth, auditPerm, p.batchHandler.UpdateStatus)
	router.POST("/batch/delete", auth, auditPerm, p.batchHandler.Delete)
	router.POST("/batch/export", auth, auditPerm, p.batchHandler.Export)

	// 审核规则管理
	router.POST("/audit-rules", auth, auditPerm, p.riskHandler.CreateAuditRule)
	router.GET("/audit-rules", auth, auditPerm, p.riskHandler.ListAuditRules)
	router.GET("/audit-rules/:id", auth, auditPerm, p.riskHandler.GetAuditRule)
	router.PUT("/audit-rules/:id", auth, auditPerm, p.riskHandler.UpdateAuditRule)
	router.DELETE("/audit-rules/:id", auth, auditPerm, p.riskHandler.DeleteAuditRule)

	// 举报管理
	router.GET("/reports", auth, auditPerm, p.riskHandler.ListReports)
	router.GET("/reports/:id", auth, auditPerm, p.riskHandler.GetReport)
	router.PUT("/reports/:id/process", auth, auditPerm, p.riskHandler.ProcessReport)

	// 店铺审核
	router.PUT("/shops/:id/audit", auth, auditPerm, p.shopHandler.Audit)
	router.PUT("/shops/:id/status", auth, auditPerm, p.shopHandler.UpdateStatus)

	// 用户信用管理
	router.GET("/credit/:user_id", auth, auditPerm, p.riskHandler.GetUserCreditByID)
	router.PUT("/credit/:user_id", auth, auditPerm, p.riskHandler.UpdateUserCredit)

	// 标签管理
	router.POST("/tags", auth, auditPerm, p.catalogHandler.CreateTag)
	router.PUT("/tags/:id", auth, auditPerm, p.catalogHandler.UpdateTag)
	router.DELETE("/tags/:id", auth, auditPerm, p.catalogHandler.DeleteTag)

	// 品牌管理
	router.POST("/brands", auth, auditPerm, p.catalogHandler.CreateBrand)
	router.PUT("/brands/:id", auth, auditPerm, p.catalogHandler.UpdateBrand)
	router.DELETE("/brands/:id", auth, auditPerm, p.catalogHandler.DeleteBrand)

	// 型号管理
	router.POST("/brands/:id/models", auth, auditPerm, p.catalogHandler.CreateModel)
	router.PUT("/models/:id", auth, auditPerm, p.catalogHandler.UpdateModel)
	router.DELETE("/models/:id", auth, auditPerm, p.catalogHandler.DeleteModel)

	// 分类属性管理
	router.POST("/category-attrs", auth, auditPerm, p.catalogHandler.CreateCategoryAttr)
	router.PUT("/category-attrs/:id", auth, auditPerm, p.catalogHandler.UpdateCategoryAttr)
	router.DELETE("/category-attrs/:id", auth, auditPerm, p.catalogHandler.DeleteCategoryAttr)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
