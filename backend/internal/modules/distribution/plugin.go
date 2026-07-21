// Package distribution 分销合伙人中台模块插件
// 依据架构设计 4.5：分销合伙人中台（distribution）
// 职责：二级分销/城市分站分成/推广渠道统计/佣金自动结算/付费合伙人等级
//
// 错误码范围：5901-5930（与 merchant 5701-5730、mall 5401-5430、dh114 5301-5340 等并列）
// 表清单（5 张，均由 backend/migrations/030_distribution_full.sql 创建）：
//   distribution_partners / distribution_channels / distribution_commissions /
//   distribution_levels / distribution_withdrawals
package distribution

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/distribution/handler"
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/modules/distribution/repository"
	"wuchang-tongcheng/internal/modules/distribution/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 分销合伙人中台模块插件
type Plugin struct {
	name    string
	version string

	// 5 个 Handler
	partnerHandler     *handler.PartnerHandler
	channelHandler     *handler.ChannelHandler
	commissionHandler  *handler.CommissionHandler
	levelHandler       *handler.LevelHandler
	withdrawalHandler  *handler.WithdrawalHandler
}

// NewPlugin 创建分销合伙人中台模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "distribution", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "distribution",
		DisplayName: "分销合伙人中台",
		Category:    "middleware",
		Description: "分销合伙人中台：合伙人管理/推广渠道/佣金结算/等级/提现",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：5 张表由 backend/migrations/030_distribution_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
// 此处仅 AutoMigrate 主表 Partner（distribution_partners，与 merchant/mall/dh114 模块保持一致）。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 distribution_partners 主表（其余表由 030_distribution_full.sql 创建）
	if err := db.AutoMigrate(&model.Partner{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（5 个） =====
	partnerRepo := repository.NewPartnerRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	commissionRepo := repository.NewCommissionRepository(db)
	levelRepo := repository.NewLevelRepository(db)
	withdrawalRepo := repository.NewWithdrawalRepository(db)

	// ===== 依赖注入：Service 层（5 个） =====
	// 注意：levelSvc 先创建（partnerSvc 依赖它），使用前向声明技巧
	levelSvc := service.NewLevelService(levelRepo, partnerRepo)
	partnerSvc := service.NewPartnerService(partnerRepo, levelSvc)
	channelSvc := service.NewChannelService(channelRepo)
	commissionSvc := service.NewCommissionService(commissionRepo, partnerRepo, channelRepo)
	withdrawalSvc := service.NewWithdrawalService(withdrawalRepo, partnerRepo)

	// ===== 依赖注入：Handler 层（5 个） =====
	p.partnerHandler = handler.NewPartnerHandler(partnerSvc)
	p.channelHandler = handler.NewChannelHandler(channelSvc)
	p.commissionHandler = handler.NewCommissionHandler(commissionSvc)
	p.levelHandler = handler.NewLevelHandler(levelSvc)
	p.withdrawalHandler = handler.NewWithdrawalHandler(withdrawalSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/distribution
//
// 路由分组：
//   - 公开路由（C 端浏览，无需登录）：合伙人列表/详情 + 渠道列表/详情/追踪 + 佣金列表/详情 + 等级列表/详情 + 提现列表/详情
//   - 需登录路由（C 端申请/操作）：合伙人申请/我的/树 + 渠道 CRUD/我的/统计 + 佣金我的/汇总 + 提现申请/我的
//   - 管理后台路由（需 distribution:manage 权限）：合伙人管理/渠道管理/佣金结算/等级管理/提现审核
//
// 注意：固定路径（/apply /mine /tree /all /track /stats /summary /pending 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "distribution_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "distribution_write"))
	managePerm := coreRouter.WrapGin(middleware.RequirePermission("distribution:manage"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 合伙人（公开只读）
	router.GET("/partners", readLimiter, p.partnerHandler.List)
	router.GET("/partners/:id", readLimiter, p.partnerHandler.GetByID)

	// 渠道（公开只读 + 追踪）
	router.GET("/channels", readLimiter, p.channelHandler.List)
	router.GET("/channels/:id", readLimiter, p.channelHandler.GetByID)
	router.POST("/channels/track", readLimiter, p.channelHandler.Track)

	// 佣金（公开只读）
	router.GET("/commissions", readLimiter, p.commissionHandler.List)
	router.GET("/commissions/:id", readLimiter, p.commissionHandler.GetByID)
	router.GET("/commissions/summary", readLimiter, p.commissionHandler.Summary)

	// 等级（公开只读）
	router.GET("/levels", readLimiter, p.levelHandler.List)
	router.GET("/levels/all", readLimiter, p.levelHandler.ListAll)
	router.GET("/levels/:id", readLimiter, p.levelHandler.GetByID)

	// 提现（公开只读）
	router.GET("/withdrawals", readLimiter, p.withdrawalHandler.List)
	router.GET("/withdrawals/:id", readLimiter, p.withdrawalHandler.GetByID)

	// ==================== 需登录路由（C 端申请/操作） ====================

	// 合伙人（需登录）- 固定路径必须放在 /:id 之前
	router.POST("/partners/apply", auth, writeLimiter, p.partnerHandler.Apply)
	router.GET("/partners/mine", auth, readLimiter, p.partnerHandler.GetMine)
	router.GET("/partners/tree", auth, readLimiter, p.partnerHandler.Tree)

	// 渠道（需登录）- 固定路径必须放在 /:id 之前
	router.POST("/channels", auth, writeLimiter, p.channelHandler.Create)
	router.GET("/channels/mine", auth, readLimiter, p.channelHandler.ListMine)
	router.GET("/channels/stats", auth, readLimiter, p.channelHandler.Stats)
	router.PUT("/channels/:id", auth, writeLimiter, p.channelHandler.Update)
	router.DELETE("/channels/:id", auth, writeLimiter, p.channelHandler.Delete)

	// 佣金（需登录）- 固定路径必须放在 /:id 之前
	router.GET("/commissions/mine", auth, readLimiter, p.commissionHandler.ListMine)

	// 提现（需登录）- 固定路径必须放在 /:id 之前
	router.POST("/withdrawals/apply", auth, writeLimiter, p.withdrawalHandler.Apply)
	router.GET("/withdrawals/mine", auth, readLimiter, p.withdrawalHandler.ListMine)

	// ==================== 管理后台路由（需 distribution:manage 权限） ====================

	admin := router.Group("/admin")

	// 合伙人管理
	admin.GET("/partners", managePerm, p.partnerHandler.AdminList)
	admin.GET("/partners/:id", managePerm, p.partnerHandler.AdminGetByID)
	admin.PUT("/partners/:id", managePerm, p.partnerHandler.AdminUpdate)
	admin.PUT("/partners/:id/status", managePerm, p.partnerHandler.AdminUpdateStatus)
	admin.PUT("/partners/:id/upgrade", managePerm, p.partnerHandler.AdminUpgrade)
	admin.PUT("/partners/:id/commission-rate", managePerm, p.partnerHandler.AdminAdjustRate)

	// 渠道管理
	admin.GET("/channels", managePerm, p.channelHandler.AdminList)

	// 佣金结算管理
	admin.GET("/commissions", managePerm, p.commissionHandler.AdminList)
	admin.POST("/commissions", managePerm, p.commissionHandler.AdminCreate)
	admin.PUT("/commissions/:id/settle", managePerm, p.commissionHandler.AdminSettle)
	admin.POST("/commissions/batch-settle", managePerm, p.commissionHandler.AdminBatchSettle)
	admin.PUT("/commissions/:id/cancel", managePerm, p.commissionHandler.AdminCancel)
	admin.GET("/commissions/summary", managePerm, p.commissionHandler.AdminSummary)

	// 等级管理
	admin.GET("/levels", managePerm, p.levelHandler.AdminList)
	admin.POST("/levels", managePerm, p.levelHandler.AdminCreate)
	admin.PUT("/levels/:id", managePerm, p.levelHandler.AdminUpdate)
	admin.DELETE("/levels/:id", managePerm, p.levelHandler.AdminDelete)
	admin.POST("/levels/check-upgrade", managePerm, p.levelHandler.AdminCheckUpgrade)

	// 提现审核
	admin.GET("/withdrawals", managePerm, p.withdrawalHandler.AdminList)
	admin.GET("/withdrawals/pending", managePerm, p.withdrawalHandler.AdminPending)
	admin.GET("/withdrawals/:id", managePerm, p.withdrawalHandler.GetByID)
	admin.PUT("/withdrawals/:id/audit", managePerm, p.withdrawalHandler.AdminAudit)
	admin.PUT("/withdrawals/:id/pay", managePerm, p.withdrawalHandler.AdminPay)
	admin.PUT("/withdrawals/:id/reject", managePerm, p.withdrawalHandler.AdminReject)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
