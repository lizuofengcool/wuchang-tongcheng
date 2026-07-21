// Package mall 同城商城模块插件
// 提供店铺/商品/SKU/购物车/订单/支付/退款/物流/评价/收货地址/分类/统计/审核规则/举报等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id + shop_id）
// 依据需求文档 7.2：主表 mall_shops（店铺列表）
// 依据 v3.2.1 架构方案：对标淘宝/京东/拼多多同城商城
package mall

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/mall/handler"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/modules/mall/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 同城商城模块插件
type Plugin struct {
	name    string
	version string

	// 15 个 Handler
	addressHandler    *handler.AddressHandler
	auditRuleHandler  *handler.AuditRuleHandler
	cartHandler       *handler.CartHandler
	categoryHandler   *handler.CategoryHandler
	logisticsHandler  *handler.LogisticsHandler
	orderHandler      *handler.OrderHandler
	orderItemHandler  *handler.OrderItemHandler
	paymentHandler    *handler.PaymentHandler
	productHandler    *handler.ProductHandler
	refundHandler     *handler.RefundHandler
	reportHandler     *handler.ReportHandler
	reviewHandler     *handler.ReviewHandler
	shopHandler       *handler.ShopHandler
	skuHandler        *handler.SkuHandler
	statisticHandler  *handler.StatisticHandler
}

// NewPlugin 创建同城商城模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "mall", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "mall",
		DisplayName: "同城商城",
		Category:    "business",
		Description: "同城商城完整功能：店铺/商品/SKU/购物车/订单/支付/退款/物流/评价/收货地址/分类/统计/审核规则/举报",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：仅 AutoMigrate主表 Shop（mall_shops），子表由 backend/migrations SQL 脚本创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移主表 mall_shops
	if err := db.AutoMigrate(&model.Shop{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（15 个） =====
	addressRepo := repository.NewAddressRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	cartRepo := repository.NewCartRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	logisticsRepo := repository.NewLogisticsRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	productRepo := repository.NewProductRepository(db)
	refundRepo := repository.NewRefundRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	shopRepo := repository.NewShopRepository(db)
	skuRepo := repository.NewSkuRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)

	// ===== 依赖注入：Service 层（15 个） =====
	addressSvc := service.NewAddressService(addressRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	cartSvc := service.NewCartService(cartRepo, productRepo, skuRepo, shopRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	logisticsSvc := service.NewLogisticsService(logisticsRepo)
	orderSvc := service.NewOrderService(orderRepo, orderItemRepo, addressRepo, productRepo, skuRepo, shopRepo, cartRepo)
	orderItemSvc := service.NewOrderItemService(orderItemRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, orderRepo)
	productSvc := service.NewProductService(productRepo, shopRepo)
	refundSvc := service.NewRefundService(refundRepo, orderRepo, orderItemRepo, paymentRepo)
	reportSvc := service.NewReportService(reportRepo)
	reviewSvc := service.NewReviewService(reviewRepo, orderItemRepo)
	shopSvc := service.NewShopService(shopRepo)
	skuSvc := service.NewSkuService(skuRepo, productRepo)
	statisticSvc := service.NewStatisticService(statisticRepo)

	// ===== 依赖注入：Handler 层（15 个） =====
	p.addressHandler = handler.NewAddressHandler(addressSvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)
	p.cartHandler = handler.NewCartHandler(cartSvc)
	p.categoryHandler = handler.NewCategoryHandler(categorySvc)
	p.logisticsHandler = handler.NewLogisticsHandler(logisticsSvc)
	p.orderHandler = handler.NewOrderHandler(orderSvc)
	p.orderItemHandler = handler.NewOrderItemHandler(orderItemSvc)
	p.paymentHandler = handler.NewPaymentHandler(paymentSvc)
	p.productHandler = handler.NewProductHandler(productSvc)
	p.refundHandler = handler.NewRefundHandler(refundSvc)
	p.reportHandler = handler.NewReportHandler(reportSvc)
	p.reviewHandler = handler.NewReviewHandler(reviewSvc)
	p.shopHandler = handler.NewShopHandler(shopSvc)
	p.skuHandler = handler.NewSkuHandler(skuSvc)
	p.statisticHandler = handler.NewStatisticHandler(statisticSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/mall
//
// 路由分组（共 180+ API）：
//   - 公开路由（C 端浏览，无需登录）：店铺/商品/分类/评价/统计/SKU/物流/订单/支付/退款/举报只读
//   - 需登录路由（C 端发布/交易）：CRUD/购物车/订单/支付/退款/评价/收货地址
//   - 管理后台路由（需 mall:audit 权限）：审核/批量/状态/统计/审核规则管理
//
// 注意：固定路径（/search /mine /featured /hot /new /by-shop /by-category 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "mall_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "mall_write"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "mall_search"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("mall:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 店铺（shops）
	router.GET("/shops", readLimiter, p.shopHandler.List)
	router.GET("/shops/search", searchLimiter, p.shopHandler.Search)
	router.GET("/shops/by-user/:user_id", readLimiter, p.shopHandler.ListByUser)
	router.GET("/shops/by-category/:category_id", readLimiter, p.shopHandler.ListByCategory)
	router.GET("/shops/:id", readLimiter, p.shopHandler.GetByID)
	router.POST("/shops/:id/view", readLimiter, p.shopHandler.IncrViewCount)

	// 商品（products）
	router.GET("/products", readLimiter, p.productHandler.List)
	router.GET("/products/search", searchLimiter, p.productHandler.Search)
	router.GET("/products/featured", readLimiter, p.productHandler.ListFeatured)
	router.GET("/products/hot", readLimiter, p.productHandler.ListHot)
	router.GET("/products/new", readLimiter, p.productHandler.ListNew)
	router.GET("/products/by-shop/:shop_id", readLimiter, p.productHandler.ListByShop)
	router.GET("/products/by-category/:category_id", readLimiter, p.productHandler.ListByCategory)
	router.GET("/products/:id", readLimiter, p.productHandler.GetByID)
	router.POST("/products/:id/view", readLimiter, p.productHandler.IncrViewCount)

	// 商城分类（categories）
	router.GET("/categories", readLimiter, p.categoryHandler.List)
	router.GET("/categories/tree", readLimiter, p.categoryHandler.ListTree)
	router.GET("/categories/by-parent", readLimiter, p.categoryHandler.ListByParent)
	router.GET("/categories/enabled", readLimiter, p.categoryHandler.ListEnabled)
	router.GET("/categories/:id", readLimiter, p.categoryHandler.GetByID)

	// SKU 规格表（skus）
	router.GET("/skus/by-product/:product_id", readLimiter, p.skuHandler.ListByProduct)
	router.GET("/skus/by-shop/:shop_id", readLimiter, p.skuHandler.ListByShop)
	router.GET("/skus/:id", readLimiter, p.skuHandler.GetByID)

	// 审核规则（audit-rules）- 公开只读
	router.GET("/audit-rules", readLimiter, p.auditRuleHandler.List)
	router.GET("/audit-rules/enabled", readLimiter, p.auditRuleHandler.ListEnabled)
	router.GET("/audit-rules/type/:rule_type", readLimiter, p.auditRuleHandler.ListByType)
	router.GET("/audit-rules/:id", readLimiter, p.auditRuleHandler.GetByID)

	// 评价（reviews）- 公开只读
	router.GET("/reviews", readLimiter, p.reviewHandler.List)
	router.GET("/reviews/stats", readLimiter, p.reviewHandler.Stats)
	router.GET("/reviews/by-product/:product_id", readLimiter, p.reviewHandler.ListByProduct)
	router.GET("/reviews/by-shop/:shop_id", readLimiter, p.reviewHandler.ListByShop)
	router.GET("/reviews/by-order/:order_id", readLimiter, p.reviewHandler.ListByOrder)
	router.GET("/reviews/:id", readLimiter, p.reviewHandler.GetByID)

	// 数据统计（statistics）- 公开只读部分
	router.GET("/statistics/summary", readLimiter, p.statisticHandler.Summary)
	router.GET("/statistics/hot-products", readLimiter, p.statisticHandler.HotProducts)
	router.GET("/statistics/hot-shops", readLimiter, p.statisticHandler.HotShops)
	router.GET("/statistics/hot-categories", readLimiter, p.statisticHandler.HotCategories)

	// 物流（logistics）- 公开只读 + 回调
	router.GET("/logistics", readLimiter, p.logisticsHandler.List)
	router.GET("/logistics/by-order/:order_id", readLimiter, p.logisticsHandler.GetByOrderID)
	router.GET("/logistics/by-tracking-no", readLimiter, p.logisticsHandler.GetByTrackingNo)
	router.GET("/logistics/by-shop/:shop_id", readLimiter, p.logisticsHandler.ListByShop)
	router.GET("/logistics/:id", readLimiter, p.logisticsHandler.GetByID)
	router.POST("/logistics/callback", p.logisticsHandler.Callback)

	// 订单（orders）- 公开只读
	router.GET("/orders/by-shop/:shop_id", readLimiter, p.orderHandler.ListByShop)
	router.GET("/orders/by-no/:order_no", readLimiter, p.orderHandler.GetByOrderNo)
	router.GET("/orders/:id", readLimiter, p.orderHandler.GetByID)

	// 订单明细（order-items）- 公开只读
	router.GET("/order-items/by-order/:order_id", readLimiter, p.orderItemHandler.ListByOrder)
	router.GET("/order-items/:id", readLimiter, p.orderItemHandler.GetByID)

	// 支付（payments）- 公开只读 + 回调
	router.GET("/payments/by-no/:payment_no", readLimiter, p.paymentHandler.GetByPaymentNo)
	router.GET("/payments/by-order/:order_id", readLimiter, p.paymentHandler.GetByOrderID)
	router.GET("/payments/by-shop/:shop_id", readLimiter, p.paymentHandler.ListByShop)
	router.GET("/payments/:id", readLimiter, p.paymentHandler.GetByID)
	router.POST("/payments/callback", p.paymentHandler.Callback)

	// 退款（refunds）- 公开只读
	router.GET("/refunds", readLimiter, p.refundHandler.List)
	router.GET("/refunds/by-refund-no/:refund_no", readLimiter, p.refundHandler.GetByRefundNo)
	router.GET("/refunds/by-order/:order_id", readLimiter, p.refundHandler.ListByOrder)
	router.GET("/refunds/by-shop/:shop_id", readLimiter, p.refundHandler.ListByShop)
	router.GET("/refunds/:id", readLimiter, p.refundHandler.GetByID)
	router.GET("/refunds/stats", readLimiter, p.refundHandler.Stats)

	// 举报（reports）- 公开只读
	router.GET("/reports", readLimiter, p.reportHandler.List)
	router.GET("/reports/by-user/:user_id", readLimiter, p.reportHandler.ListByUser)
	router.GET("/reports/by-target", readLimiter, p.reportHandler.ListByTarget)
	router.GET("/reports/stats", readLimiter, p.reportHandler.Stats)
	router.GET("/reports/:id", readLimiter, p.reportHandler.GetByID)

	// ==================== 需登录路由（C 端发布/交易） ====================

	// 店铺 CRUD
	router.POST("/shops", auth, writeLimiter, p.shopHandler.Create)
	router.PUT("/shops/:id", auth, writeLimiter, p.shopHandler.Update)
	router.DELETE("/shops/:id", auth, writeLimiter, p.shopHandler.Delete)
	router.GET("/shops/mine", auth, readLimiter, p.shopHandler.GetByUserID)
	router.PUT("/shops/:id/status", auth, writeLimiter, p.shopHandler.UpdateStatus)

	// 商品 CRUD
	router.POST("/products", auth, writeLimiter, p.productHandler.Create)
	router.PUT("/products/:id", auth, writeLimiter, p.productHandler.Update)
	router.DELETE("/products/:id", auth, writeLimiter, p.productHandler.Delete)
	router.GET("/products/mine", auth, readLimiter, p.productHandler.ListByUser)
	router.PUT("/products/:id/status", auth, writeLimiter, p.productHandler.UpdateStatus)

	// SKU CRUD
	router.POST("/skus", auth, writeLimiter, p.skuHandler.Create)
	router.PUT("/skus/:id", auth, writeLimiter, p.skuHandler.Update)
	router.DELETE("/skus/:id", auth, writeLimiter, p.skuHandler.Delete)
	router.PUT("/skus/:id/stock", auth, writeLimiter, p.skuHandler.UpdateStock)
	router.PUT("/skus/batch-stock", auth, writeLimiter, p.skuHandler.BatchUpdateStock)

	// 收货地址 CRUD
	router.POST("/addresses", auth, writeLimiter, p.addressHandler.Create)
	router.GET("/addresses", auth, readLimiter, p.addressHandler.ListByUser)
	router.GET("/addresses/default", auth, readLimiter, p.addressHandler.GetDefault)
	router.GET("/addresses/:id", auth, readLimiter, p.addressHandler.GetByID)
	router.PUT("/addresses/:id", auth, writeLimiter, p.addressHandler.Update)
	router.PUT("/addresses/:id/default", auth, writeLimiter, p.addressHandler.SetDefault)
	router.DELETE("/addresses/:id", auth, writeLimiter, p.addressHandler.Delete)

	// 购物车 CRUD
	router.POST("/cart", auth, writeLimiter, p.cartHandler.Add)
	router.GET("/cart", auth, readLimiter, p.cartHandler.ListByUser)
	router.GET("/cart/summary", auth, readLimiter, p.cartHandler.Summary)
	router.GET("/cart/count", auth, readLimiter, p.cartHandler.CountByUser)
	router.GET("/cart/count-selected", auth, readLimiter, p.cartHandler.CountSelectedByUser)
	router.GET("/cart/selected", auth, readLimiter, p.cartHandler.ListSelected)
	router.GET("/cart/group-by-shop", auth, readLimiter, p.cartHandler.ListGroupByShop)
	router.GET("/cart/by-shop/:shop_id", auth, readLimiter, p.cartHandler.ListByUserAndShop)
	router.GET("/cart/:id", auth, readLimiter, p.cartHandler.GetByID)
	router.PUT("/cart/:id", auth, writeLimiter, p.cartHandler.Update)
	router.PUT("/cart/batch", auth, writeLimiter, p.cartHandler.BatchUpdate)
	router.PUT("/cart/select-all", auth, writeLimiter, p.cartHandler.SelectAll)
	router.DELETE("/cart/:id", auth, writeLimiter, p.cartHandler.Delete)
	router.DELETE("/cart/batch", auth, writeLimiter, p.cartHandler.BatchDelete)
	router.DELETE("/cart/clear", auth, writeLimiter, p.cartHandler.ClearByUser)
	router.DELETE("/cart/clear-by-shop/:shop_id", auth, writeLimiter, p.cartHandler.ClearByUserAndShop)

	// 订单 CRUD + 状态流转
	router.POST("/orders", auth, writeLimiter, p.orderHandler.Create)
	router.GET("/orders/mine", auth, readLimiter, p.orderHandler.ListByUser)
	router.GET("/orders/count-by-status", auth, readLimiter, p.orderHandler.CountByStatus)
	router.GET("/orders/summary", auth, readLimiter, p.orderHandler.Summary)
	router.PUT("/orders/:id/cancel", auth, writeLimiter, p.orderHandler.Cancel)
	router.PUT("/orders/:id/ship", auth, writeLimiter, p.orderHandler.Ship)
	router.PUT("/orders/:id/confirm", auth, writeLimiter, p.orderHandler.Confirm)
	router.PUT("/orders/:id/complete", auth, writeLimiter, p.orderHandler.Complete)
	router.DELETE("/orders/:id", auth, writeLimiter, p.orderHandler.Delete)

	// 订单明细 - 用户视角
	router.GET("/order-items/mine", auth, readLimiter, p.orderItemHandler.ListByUser)

	// 支付 CRUD
	router.POST("/payments", auth, writeLimiter, p.paymentHandler.Create)
	router.PUT("/payments/:id/close", auth, writeLimiter, p.paymentHandler.Close)
	router.GET("/payments/mine", auth, readLimiter, p.paymentHandler.ListByUser)

	// 退款 CRUD + 状态流转
	router.POST("/refunds", auth, writeLimiter, p.refundHandler.Create)
	router.GET("/refunds/mine", auth, readLimiter, p.refundHandler.ListByUser)
	router.PUT("/refunds/:id/seller-process", auth, writeLimiter, p.refundHandler.SellerProcess)
	router.PUT("/refunds/:id/ship", auth, writeLimiter, p.refundHandler.Ship)

	// 评价 CRUD + 点赞
	router.POST("/reviews", auth, writeLimiter, p.reviewHandler.Create)
	router.GET("/reviews/mine", auth, readLimiter, p.reviewHandler.ListByUser)
	router.PUT("/reviews/:id/reply", auth, writeLimiter, p.reviewHandler.Reply)
	router.PUT("/reviews/:id/append", auth, writeLimiter, p.reviewHandler.Append)
	router.DELETE("/reviews/:id", auth, writeLimiter, p.reviewHandler.Delete)
	router.POST("/reviews/:id/like", auth, p.reviewHandler.Like)
	router.POST("/reviews/:id/dislike", auth, p.reviewHandler.Dislike)

	// 举报 CRUD（Create 需 auth 登录；Process/Delete 需 auth + mall:audit 权限）
	router.POST("/reports", auth, writeLimiter, p.reportHandler.Create)

	// 物流 CRUD
	router.POST("/logistics", auth, writeLimiter, p.logisticsHandler.Create)
	router.PUT("/logistics/:id", auth, writeLimiter, p.logisticsHandler.Update)
	router.DELETE("/logistics/:id", auth, writeLimiter, p.logisticsHandler.Delete)
	router.GET("/logistics/mine", auth, readLimiter, p.logisticsHandler.ListByUser)
	router.PUT("/logistics/:id/status", auth, writeLimiter, p.logisticsHandler.UpdateStatus)

	// ==================== 管理后台路由（需 mall:audit 权限） ====================

	admin := router.Group("/admin")

	// 店铺管理
	admin.GET("/shops", auditPerm, p.shopHandler.AdminList)
	admin.PUT("/shops/:id/audit", auditPerm, p.shopHandler.Audit)
	admin.PUT("/shops/:id/promotion", auditPerm, p.shopHandler.UpdatePromotion)

	// 商品管理
	admin.GET("/products", auditPerm, p.productHandler.AdminList)
	admin.PUT("/products/:id/audit", auditPerm, p.productHandler.Audit)
	admin.PUT("/products/:id/promotion", auditPerm, p.productHandler.UpdatePromotion)

	// 分类管理
	admin.GET("/categories", auditPerm, p.categoryHandler.List)
	admin.POST("/categories", auditPerm, p.categoryHandler.Create)
	admin.PUT("/categories/:id", auditPerm, p.categoryHandler.Update)
	admin.DELETE("/categories/:id", auditPerm, p.categoryHandler.Delete)
	admin.PUT("/categories/:id/status", auditPerm, p.categoryHandler.UpdateStatus)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.List)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.UpdateStatus)
	admin.POST("/audit-rules/check", auditPerm, p.auditRuleHandler.Check)

	// 订单管理
	admin.GET("/orders", auditPerm, p.orderHandler.AdminList)
	admin.PUT("/orders/:id/close", auditPerm, p.orderHandler.AdminClose)
	admin.PUT("/orders/batch-status", auditPerm, p.orderHandler.BatchUpdateStatus)
	admin.POST("/orders/auto-close", auditPerm, p.orderHandler.AutoClose)
	admin.POST("/orders/auto-confirm", auditPerm, p.orderHandler.AutoConfirm)
	admin.POST("/orders/auto-review", auditPerm, p.orderHandler.AutoReview)

	// 订单明细管理
	admin.GET("/order-items", auditPerm, p.orderItemHandler.List)
	admin.PUT("/order-items/:id/review-status", auditPerm, p.orderItemHandler.UpdateReviewStatus)
	admin.PUT("/order-items/:id/refund-status", auditPerm, p.orderItemHandler.UpdateRefundStatus)

	// 支付管理
	admin.GET("/payments", auditPerm, p.paymentHandler.List)
	admin.GET("/payments/stats", auditPerm, p.paymentHandler.Stats)

	// 退款管理
	admin.PUT("/refunds/:id/admin-process", auditPerm, p.refundHandler.AdminProcess)

	// 评价管理
	admin.GET("/reviews", auditPerm, p.reviewHandler.AdminList)
	admin.PUT("/reviews/:id/status", auditPerm, p.reviewHandler.UpdateStatus)

	// 举报管理
	admin.GET("/reports", auditPerm, p.reportHandler.List)
	admin.PUT("/reports/:id/process", auditPerm, p.reportHandler.Process)
	admin.DELETE("/reports/:id", auditPerm, p.reportHandler.Delete)

	// 收货地址管理（仅列表）
	admin.GET("/addresses", auditPerm, p.addressHandler.List)

	// 物流管理
	admin.PUT("/logistics/:id/status", auditPerm, p.logisticsHandler.UpdateStatus)

	// 数据统计管理
	admin.GET("/statistics", auditPerm, p.statisticHandler.List)
	admin.POST("/statistics", auditPerm, p.statisticHandler.Upsert)
	admin.GET("/statistics/overview", auditPerm, p.statisticHandler.Overview)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
