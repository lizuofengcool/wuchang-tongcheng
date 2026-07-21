// Package merchant 商户中台模块插件
// 依据架构设计 4.4：商家商户中台
// 职责：入驻/认领/店铺管理/CRM/商家权限/商家结算/商家营销工具
// 说明：所有"经纪人/企业/师傅/商家"角色统一复用此中台
//       （房产经纪人、招聘企业、二手商家、到家师傅都是 merchant）
//
// 错误码范围：5701-5730（与 mall/dh114/pinche 等模块并列）
// 表清单（5 张，均由 backend/migrations/028_merchant_full.sql 创建）：
//   merchant_shops / merchant_staff / merchant_settles / merchant_categories / merchant_verifications
package merchant

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/merchant/handler"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/modules/merchant/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 商户中台模块插件
type Plugin struct {
	name    string
	version string

	// 5 个 Handler
	shopHandler         *handler.ShopHandler
	staffHandler        *handler.StaffHandler
	settleHandler       *handler.SettleHandler
	categoryHandler     *handler.CategoryHandler
	verificationHandler *handler.VerificationHandler
}

// NewPlugin 创建商户中台模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "merchant", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "merchant",
		DisplayName: "商户中台",
		Category:    "middleware",
		Description: "商户中台：入驻/认领/店铺管理/员工权限/结算/类目/认证",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：5 张表由 backend/migrations/028_merchant_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
// 此处仅 AutoMigrate 主表 Shop（merchant_shops，与 mall/dh114 模块保持一致）。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 merchant_shops 主表（其余表由 028_merchant_full.sql 创建）
	if err := db.AutoMigrate(&model.Shop{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（5 个） =====
	shopRepo := repository.NewShopRepository(db)
	staffRepo := repository.NewStaffRepository(db)
	settleRepo := repository.NewSettleRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	verificationRepo := repository.NewVerificationRepository(db)

	// ===== 依赖注入：Service 层（5 个） =====
	shopSvc := service.NewShopService(shopRepo)
	staffSvc := service.NewStaffService(staffRepo)
	settleSvc := service.NewSettleService(settleRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	verificationSvc := service.NewVerificationService(verificationRepo)

	// ===== 依赖注入：Handler 层（5 个） =====
	p.shopHandler = handler.NewShopHandler(shopSvc)
	p.staffHandler = handler.NewStaffHandler(staffSvc)
	p.settleHandler = handler.NewSettleHandler(settleSvc)
	p.categoryHandler = handler.NewCategoryHandler(categorySvc)
	p.verificationHandler = handler.NewVerificationHandler(verificationSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/merchant
//
// 路由分组：
//   - 公开路由（C 端浏览，无需登录）：店铺列表/详情/搜索 + 类目树 + 类目列表/详情 + 员工详情 + 结算列表/详情 + 认证详情/列表
//   - 需登录路由（C 端发布/认领/操作）：店铺入驻/更新/我的/认领 + 员工 CRUD + 认证提交/更新/删除
//   - 管理后台路由（需 merchant:audit 权限）：店铺审核/状态/信用分/等级 + 结算生成/提现/审核 + 类目 CRUD + 认证审核
//
// 注意：固定路径（/search /mine /apply /claim /tree 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "merchant_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "merchant_write"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("merchant:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 店铺（公开）
	router.GET("/shops", readLimiter, p.shopHandler.List)
	router.GET("/shops/search", readLimiter, p.shopHandler.Search)
	router.GET("/shops/:id", readLimiter, p.shopHandler.GetByID)

	// 类目（公开）
	router.GET("/categories/tree", readLimiter, p.categoryHandler.Tree)
	router.GET("/categories", readLimiter, p.categoryHandler.List)
	router.GET("/categories/:id", readLimiter, p.categoryHandler.GetByID)

	// 员工（公开只读）
	router.GET("/staff/:id", readLimiter, p.staffHandler.GetByID)
	router.GET("/staff", readLimiter, p.staffHandler.List)

	// 结算（公开只读）
	router.GET("/settles", readLimiter, p.settleHandler.List)
	router.GET("/settles/:id", readLimiter, p.settleHandler.GetByID)
	router.GET("/shops/:id/settles", readLimiter, p.settleHandler.ListByShop)

	// 认证（公开只读）
	router.GET("/verifications", readLimiter, p.verificationHandler.List)
	router.GET("/verifications/:id", readLimiter, p.verificationHandler.GetByID)
	router.GET("/shops/:id/verifications", readLimiter, p.verificationHandler.ListByShop)

	// ==================== 需登录路由（C 端发布/认领/操作） ====================

	// 店铺（需登录）- 固定路径必须放在 /:id 之前
	router.POST("/shops/apply", auth, writeLimiter, p.shopHandler.Apply)
	router.POST("/shops/claim", auth, writeLimiter, p.shopHandler.Claim)
	router.GET("/shops/mine", auth, readLimiter, p.shopHandler.ListMine)
	router.PUT("/shops/:id", auth, writeLimiter, p.shopHandler.Update)

	// 员工（需登录）
	router.POST("/staff", auth, writeLimiter, p.staffHandler.Create)
	router.PUT("/staff/:id", auth, writeLimiter, p.staffHandler.Update)
	router.DELETE("/staff/:id", auth, writeLimiter, p.staffHandler.Delete)
	router.PUT("/staff/:id/permissions", auth, writeLimiter, p.staffHandler.AssignPermissions)
	router.PUT("/staff/:id/role", auth, writeLimiter, p.staffHandler.SwitchRole)

	// 认证（需登录）
	router.POST("/verifications", auth, writeLimiter, p.verificationHandler.Create)
	router.PUT("/verifications/:id", auth, writeLimiter, p.verificationHandler.Update)
	router.DELETE("/verifications/:id", auth, writeLimiter, p.verificationHandler.Delete)

	// ==================== 管理后台路由（需 merchant:audit 权限） ====================

	admin := router.Group("/admin")

	// 店铺管理
	admin.GET("/shops", auditPerm, p.shopHandler.AdminList)
	admin.GET("/shops/:id", auditPerm, p.shopHandler.AdminGetByID)
	admin.PUT("/shops/:id/status", auditPerm, p.shopHandler.UpdateStatus)
	admin.PUT("/shops/:id/credit", auditPerm, p.shopHandler.UpdateCreditScore)
	admin.PUT("/shops/:id/level", auditPerm, p.shopHandler.UpdateLevel)

	// 结算管理
	admin.GET("/settles", auditPerm, p.settleHandler.List)
	admin.GET("/settles/:id", auditPerm, p.settleHandler.GetByID)
	admin.POST("/settles", auditPerm, p.settleHandler.Generate)
	admin.PUT("/settles/:id/withdraw", auditPerm, p.settleHandler.Withdraw)
	admin.PUT("/settles/:id/audit", auditPerm, p.settleHandler.AuditWithdraw)
	admin.GET("/settles/summary-by-shop", auditPerm, p.settleHandler.SummaryByShop)
	admin.GET("/settles/summary-by-period", auditPerm, p.settleHandler.SummaryByPeriod)

	// 类目管理
	admin.GET("/categories", auditPerm, p.categoryHandler.List)
	admin.GET("/categories/:id", auditPerm, p.categoryHandler.GetByID)
	admin.POST("/categories", auditPerm, p.categoryHandler.Create)
	admin.PUT("/categories/:id", auditPerm, p.categoryHandler.Update)
	admin.DELETE("/categories/:id", auditPerm, p.categoryHandler.Delete)
	admin.PUT("/categories/:id/status", auditPerm, p.categoryHandler.UpdateStatus)

	// 认证管理
	admin.GET("/verifications", auditPerm, p.verificationHandler.AdminList)
	admin.PUT("/verifications/:id/audit", auditPerm, p.verificationHandler.Audit)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
