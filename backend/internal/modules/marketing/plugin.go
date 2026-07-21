// Package marketing 营销活动中台模块插件
// 提供广告位/优惠券/签到/营销活动等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：4 子域 ad/coupon/sign/activity
package marketing

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/marketing/handler"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/modules/marketing/service"
	"wuchang-tongcheng/internal/pkg/database"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Plugin 营销活动中台模块插件
type Plugin struct {
	name    string
	version string

	// 4 个 Handler（按子域组织）
	adHandler       *handler.AdHandler
	couponHandler   *handler.CouponHandler
	signHandler     *handler.SignHandler
	activityHandler *handler.ActivityHandler
}

// NewPlugin 创建营销活动中台模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "marketing", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "marketing",
		DisplayName: "营销中台",
		Category:    "middleware",
		Description: "营销活动中台完整功能：广告位/优惠券/签到/营销活动（拼团/砍价/秒杀/抽奖）",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：6 张表由 backend/migrations/029_marketing_full.sql 创建，
// 此处不调用 AutoMigrate，避免与 SQL 脚本约束名不一致。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// ===== 依赖注入：Repository 层（4 个） =====
	adRepo := repository.NewAdRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	signRepo := repository.NewSignRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// ===== 依赖注入：Service 层（4 个） =====
	adSvc := service.NewAdService(adRepo)
	couponSvc := service.NewCouponService(couponRepo)
	signSvc := service.NewSignService(signRepo)
	activitySvc := service.NewActivityService(activityRepo)

	// ===== 依赖注入：Handler 层（4 个） =====
	p.adHandler = handler.NewAdHandler(adSvc)
	p.couponHandler = handler.NewCouponHandler(couponSvc)
	p.signHandler = handler.NewSignHandler(signSvc)
	p.activityHandler = handler.NewActivityHandler(activitySvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/marketing
//
// 路由分组（共 40+ API）：
//   - 公开路由（C 端浏览，无需登录）：广告位按位置/优惠券列表/活动列表/签到规则
//   - 需登录路由（C 端用户操作）：签到/领取优惠券/我的优惠券/使用优惠券
//   - 管理后台路由（需 marketing:manage 权限）：广告位/优惠券/签到规则/活动 CRUD
//
// 注意：固定路径（/available /ongoing /upcoming /ended /calendar /check-in /my-coupons 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "marketing_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "marketing_write"))
	managePerm := coreRouter.WrapGin(middleware.RequirePermission("marketing:manage"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 广告位 - 公开按位置编码查询
	router.GET("/positions/:code/ads", readLimiter, p.adHandler.ListByPositionCode)

	// 优惠券 - 公开列表与详情
	router.GET("/coupons/available", readLimiter, p.couponHandler.ListAvailable)
	router.GET("/coupons/:id", readLimiter, p.couponHandler.GetByID)

	// 营销活动 - 公开列表与详情
	router.GET("/activities/ongoing", readLimiter, p.activityHandler.ListOngoing)
	router.GET("/activities/upcoming", readLimiter, p.activityHandler.ListUpcoming)
	router.GET("/activities/ended", readLimiter, p.activityHandler.ListEnded)
	router.GET("/activities/:id", readLimiter, p.activityHandler.GetByID)

	// 签到规则 - 公开只读
	router.GET("/sign/rules/enabled", readLimiter, p.signHandler.ListEnabledRules)

	// ==================== 需登录路由（C 端用户操作） ====================

	// 签到
	router.POST("/sign/check-in", auth, writeLimiter, p.signHandler.CheckIn)
	router.GET("/sign/calendar", auth, readLimiter, p.signHandler.GetCalendar)

	// 优惠券 - 领取/使用/退还/我的
	router.POST("/coupons/:id/receive", auth, writeLimiter, p.couponHandler.Receive)
	router.POST("/user-coupons/:id/use", auth, writeLimiter, p.couponHandler.Use)
	router.POST("/user-coupons/:id/refund", auth, writeLimiter, p.couponHandler.Refund)
	router.GET("/my-coupons", auth, readLimiter, p.couponHandler.ListMine)

	// ==================== 管理后台路由（需 marketing:manage 权限） ====================

	admin := router.Group("/admin")

	// 广告位管理
	admin.GET("/ads", managePerm, p.adHandler.List)
	admin.POST("/ads", managePerm, writeLimiter, p.adHandler.Create)
	admin.GET("/ads/:id", managePerm, p.adHandler.GetByID)
	admin.PUT("/ads/:id", managePerm, writeLimiter, p.adHandler.Update)
	admin.DELETE("/ads/:id", managePerm, writeLimiter, p.adHandler.Delete)

	// 优惠券管理
	admin.GET("/coupons", managePerm, p.couponHandler.List)
	admin.POST("/coupons", managePerm, writeLimiter, p.couponHandler.Create)
	admin.GET("/coupons/:id", managePerm, p.couponHandler.GetByID)
	admin.PUT("/coupons/:id", managePerm, writeLimiter, p.couponHandler.Update)
	admin.DELETE("/coupons/:id", managePerm, writeLimiter, p.couponHandler.Delete)
	admin.GET("/coupons/statistics", managePerm, p.couponHandler.Statistics)

	// 签到规则管理
	admin.GET("/sign-rules", managePerm, p.signHandler.ListRules)
	admin.POST("/sign-rules", managePerm, writeLimiter, p.signHandler.CreateRule)
	admin.GET("/sign-rules/:id", managePerm, p.signHandler.GetRuleByID)
	admin.PUT("/sign-rules/:id", managePerm, writeLimiter, p.signHandler.UpdateRule)
	admin.DELETE("/sign-rules/:id", managePerm, writeLimiter, p.signHandler.DeleteRule)

	// 营销活动管理
	admin.GET("/activities", managePerm, p.activityHandler.List)
	admin.POST("/activities", managePerm, writeLimiter, p.activityHandler.Create)
	admin.GET("/activities/:id", managePerm, p.activityHandler.GetByID)
	admin.PUT("/activities/:id", managePerm, writeLimiter, p.activityHandler.Update)
	admin.DELETE("/activities/:id", managePerm, writeLimiter, p.activityHandler.Delete)
	admin.PUT("/activities/:id/status", managePerm, writeLimiter, p.activityHandler.UpdateStatus)
	admin.GET("/activities/statistics", managePerm, p.activityHandler.Statistics)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
// 同时注册本模块错误码到 utils（5801-5830 区间，未在 error_code.go 中声明）
func init() {
	// 注册营销中台错误码消息（避开 error_code.go 修改）
	registerMarketingCodes()
	plugin.AutoRegister(NewPlugin())
}

// registerMarketingCodes 注册营销中台错误码到 utils.RegisterCode
// 错误码区间 5801-5830，与 handler/codes.go 中的常量保持一致
func registerMarketingCodes() {
	utils.RegisterCode(handler.CodeMarketingAdError, "广告位错误")
	utils.RegisterCode(handler.CodeMarketingAdNotFound, "广告位不存在")
	utils.RegisterCode(handler.CodeMarketingAdStatusInvalid, "广告位状态不允许此操作")

	utils.RegisterCode(handler.CodeMarketingCouponError, "优惠券错误")
	utils.RegisterCode(handler.CodeMarketingCouponNotFound, "优惠券不存在")
	utils.RegisterCode(handler.CodeMarketingCouponStatusInvalid, "优惠券状态不允许此操作")
	utils.RegisterCode(handler.CodeMarketingCouponSoldOut, "优惠券已抢完")
	utils.RegisterCode(handler.CodeMarketingCouponExpired, "优惠券已过期")
	utils.RegisterCode(handler.CodeMarketingCouponNotStarted, "优惠券尚未开始领取")
	utils.RegisterCode(handler.CodeMarketingCouponAlreadyRecv, "已领取过该优惠券")
	utils.RegisterCode(handler.CodeMarketingUserCouponNotFound, "用户优惠券不存在")
	utils.RegisterCode(handler.CodeMarketingUserCouponUsed, "用户优惠券已使用")
	utils.RegisterCode(handler.CodeMarketingUserCouponExpired, "用户优惠券已过期")

	utils.RegisterCode(handler.CodeMarketingSignError, "签到错误")
	utils.RegisterCode(handler.CodeMarketingSignRuleError, "签到规则错误")
	utils.RegisterCode(handler.CodeMarketingSignRuleNotFound, "签到规则不存在")

	utils.RegisterCode(handler.CodeMarketingActivityError, "营销活动错误")
	utils.RegisterCode(handler.CodeMarketingActivityNotFound, "营销活动不存在")
	utils.RegisterCode(handler.CodeMarketingActivityStatusInvalid, "营销活动状态不允许此操作")
	utils.RegisterCode(handler.CodeMarketingActivityNotOngoing, "活动未在进行中")
}
