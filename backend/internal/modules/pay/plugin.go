// Package pay 支付财务中台精简版插件
// 依据 ershou 模块依赖：担保交易 + 订单 + 退款 + 提现 + 结算 + 资金账户
// 路由前缀 /api/v1/pay
package pay

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/pay/handler"
	"wuchang-tongcheng/internal/modules/pay/repository"
	"wuchang-tongcheng/internal/modules/pay/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 支付中台插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建支付中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "pay", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "pay",
		DisplayName:  "支付财务中台",
		Category:     "middleware",
		Description:  "支付订单、担保交易、退款、提现、结算、资金账户",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注意：不调用 AutoMigrate，表结构由 migrations/005_p1_middlewares.sql 创建
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()
	repo := repository.NewPayRepository(db)
	svc := service.NewPayService(repo)
	p.handler = handler.NewHandler(svc)
	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/pay
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "pay_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "pay_write"))
	financeManage := coreRouter.WrapGin(middleware.RequirePermission("finance:reconcile"))

	// 订单（需登录）
	router.POST("/orders", auth, writeLimiter, p.handler.CreatePayment)
	router.POST("/orders/confirm", writeLimiter, p.handler.ConfirmPayment) // 第三方回调，不强制登录
	router.GET("/orders", auth, readLimiter, p.handler.ListPayments)
	router.GET("/orders/:order_no", auth, readLimiter, p.handler.GetPayment)

	// 担保交易
	router.POST("/escrow/confirm", auth, writeLimiter, p.handler.ConfirmEscrow)
	router.GET("/escrow/:order_no", auth, readLimiter, p.handler.GetEscrow)

	// 退款
	router.POST("/refunds", writeLimiter, p.handler.Refund)
	router.GET("/refunds/:refund_no", auth, readLimiter, p.handler.GetRefund)

	// 提现
	router.POST("/withdrawals", auth, writeLimiter, p.handler.Withdraw)
	router.GET("/withdrawals", auth, readLimiter, p.handler.ListMyWithdrawals)

	// 资金账户
	router.GET("/account", auth, readLimiter, p.handler.GetAccount)

	// 结算查询
	router.GET("/settlements", auth, readLimiter, p.handler.ListSettlements)

	// M 端管理后台（需 finance:reconcile 权限）
	admin := router.Group("/admin")
	admin.GET("/withdrawals/pending", financeManage, p.handler.ListPendingWithdrawals)
	admin.POST("/withdrawals/handle", financeManage, p.handler.HandleWithdrawal)
	admin.POST("/settlements", financeManage, p.handler.Settle)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
