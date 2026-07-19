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
	extH    *handler.ExtendHandler
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
	extRepo := repository.NewPayExtendRepository(db)
	svc := service.NewPayService(repo)
	extSvc := service.NewPayExtendService(repo, extRepo)
	p.handler = handler.NewHandler(svc)
	p.extH = handler.NewExtendHandler(extSvc)
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

	// 交易流水
	router.GET("/transactions", auth, readLimiter, p.extH.ListTransactions)
	router.GET("/transactions/:txn_no", auth, readLimiter, p.extH.GetTransaction)

	// 渠道（公开查询 + M 端管理）
	router.GET("/channels", auth, readLimiter, p.extH.ListChannels)
	router.GET("/channels/:id", auth, readLimiter, p.extH.GetChannel)

	// 商户
	router.POST("/merchants", auth, writeLimiter, p.extH.CreateMerchant)
	router.GET("/merchants", auth, readLimiter, p.extH.ListMerchants)
	router.GET("/merchants/:id", auth, readLimiter, p.extH.GetMerchant)

	// 回调（第三方调用，不强制登录）
	router.POST("/callbacks", writeLimiter, p.extH.RecordCallback)
	router.GET("/callbacks/:id", auth, readLimiter, p.extH.GetCallback)

	// 担保争议
	router.POST("/escrow/disputes", auth, writeLimiter, p.extH.CreateDispute)

	// M 端管理后台（需 finance:reconcile 权限）
	admin := router.Group("/admin")
	admin.GET("/withdrawals/pending", financeManage, p.handler.ListPendingWithdrawals)
	admin.POST("/withdrawals/handle", financeManage, p.handler.HandleWithdrawal)
	admin.POST("/settlements", financeManage, p.handler.Settle)
	// 扩展 M 端
	admin.POST("/channels", financeManage, p.extH.CreateChannel)
	admin.POST("/channels/:id", financeManage, p.extH.UpdateChannel)
	admin.DELETE("/channels/:id", financeManage, p.extH.DeleteChannel)
	admin.POST("/merchants/audit", financeManage, p.extH.AuditMerchant)
	admin.GET("/callbacks", financeManage, p.extH.ListCallbacks)
	admin.GET("/disputes", financeManage, p.extH.ListDisputes)
	admin.POST("/disputes/arbitrate", financeManage, p.extH.ArbitrateDispute)
	admin.GET("/statistics", financeManage, p.extH.Statistics)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
