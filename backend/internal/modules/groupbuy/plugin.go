// Package groupbuy 团购优惠券模块插件
// 提供团购商品、优惠券、用户优惠券、团购订单的完整管理
package groupbuy

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/groupbuy/handler"
	"wuchang-tongcheng/internal/modules/groupbuy/model"
	"wuchang-tongcheng/internal/modules/groupbuy/repository"
	"wuchang-tongcheng/internal/modules/groupbuy/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 团购优惠券模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建团购优惠券模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "groupbuy", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "groupbuy",
		DisplayName:  "团购优惠",
		Category:     "business",
		Description:  "提供团购商品、优惠券、用户优惠券、团购订单的完整管理",
		Version:      p.version,
		Dependencies: []string{"shop"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移团购相关表
	if err := db.AutoMigrate(
		&model.GroupBuy{},
		&model.Coupon{},
		&model.UserCoupon{},
		&model.GroupBuyOrder{},
	); err != nil {
		return err
	}

	// 初始化依赖链
	gbRepo := repository.NewGroupBuyRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	ucRepo := repository.NewUserCouponRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	svc := service.NewService(gbRepo, couponRepo, ucRepo, orderRepo)
	p.handler = handler.NewHandler(svc)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// ===== 公开接口（无需登录） =====
	router.GET("/list", p.handler.List)
	router.GET("/coupons", p.handler.ListCoupons)
	router.GET("/:id", p.handler.GetByID)

	// ===== 用户接口（需登录） =====
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	router.POST("/orders", auth, p.handler.CreateOrder)
	router.GET("/orders/my", auth, p.handler.MyOrders)
	router.GET("/orders/:id", auth, p.handler.GetOrder)
	router.POST("/orders/:id/cancel", auth, p.handler.CancelOrder)
	router.GET("/coupons/my", auth, p.handler.MyCoupons)
	router.POST("/coupons/:id/receive", auth, p.handler.ReceiveCoupon)

	// ===== 核销接口（需 groupbuy:verify 权限） =====
	router.POST("/orders/:id/verify",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:verify")),
		p.handler.VerifyOrder)

	// ===== 管理接口（需对应权限） =====
	router.POST("/admin",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:create")),
		p.handler.AdminCreate)
	router.GET("/admin",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:read")),
		p.handler.AdminList)
	router.PUT("/admin/:id",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:update")),
		p.handler.AdminUpdate)
	router.DELETE("/admin/:id",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:delete")),
		p.handler.AdminDelete)
	router.PUT("/admin/:id/status",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:update")),
		p.handler.AdminUpdateStatus)
	router.PUT("/admin/:id/audit",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:audit")),
		p.handler.AdminAudit)
	router.GET("/admin/orders",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:read")),
		p.handler.AdminOrderList)

	// 优惠券管理
	router.POST("/admin/coupons",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:create")),
		p.handler.AdminCreateCoupon)
	router.GET("/admin/coupons",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:read")),
		p.handler.AdminListCoupons)
	router.PUT("/admin/coupons/:id",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:update")),
		p.handler.AdminUpdateCoupon)
	router.DELETE("/admin/coupons/:id",
		coreRouter.WrapGin(middleware.RequirePermission("groupbuy:delete")),
		p.handler.AdminDeleteCoupon)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
