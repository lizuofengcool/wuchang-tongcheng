// Package tenant 多租户分站中台模块插件
// 依据架构设计 ershou-大模块架构方案.md 第 4.10 节：多租户分站中台
// 职责：无限城市分站/独立配置/独立域名/独立运营权限/配置一键复制/数据隔离
//
// 错误码段：5601-5630（在 handler/common.go 中定义，不修改 error_code.go）
// 权限码：tenant:manage（admin 路由）
package tenant

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/tenant/handler"
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/modules/tenant/repository"
	"wuchang-tongcheng/internal/modules/tenant/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 多租户分站中台插件
type Plugin struct {
	name    string
	version string

	// 4 个 Handler
	stationHandler *handler.StationHandler
	staffHandler   *handler.StaffHandler
	configHandler  *handler.ConfigHandler
	domainHandler  *handler.DomainHandler
}

// NewPlugin 创建多租户分站中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "tenant", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "tenant",
		DisplayName: "分站中台",
		Category:    "middleware",
		Description: "多租户分站中台：分站/员工/配置/域名，支持独立配置与域名绑定、配置一键复制",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
// 4 张表通过 GORM AutoMigrate 创建，同时由 backend/migrations/027_tenant_full.sql 保证幂等（索引/触发器/注释）
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 4 张表
	if err := db.AutoMigrate(
		&model.Station{},
		&model.Staff{},
		&model.Config{},
		&model.Domain{},
	); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（4 个） =====
	stationRepo := repository.NewStationRepository(db)
	staffRepo := repository.NewStaffRepository(db)
	configRepo := repository.NewConfigRepository(db)
	domainRepo := repository.NewDomainRepository(db)

	// ===== 依赖注入：Service 层（4 个） =====
	stationSvc := service.NewStationService(stationRepo, configRepo)
	staffSvc := service.NewStaffService(staffRepo, stationRepo)
	configSvc := service.NewConfigService(configRepo, stationRepo)
	domainSvc := service.NewDomainService(domainRepo, stationRepo)

	// ===== 依赖注入：Handler 层（4 个） =====
	p.stationHandler = handler.NewStationHandler(stationSvc)
	p.staffHandler = handler.NewStaffHandler(staffSvc)
	p.configHandler = handler.NewConfigHandler(configSvc)
	p.domainHandler = handler.NewDomainHandler(domainSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/tenant
//
// 路由分组：
//   - 公开路由（C 端根据域名识别当前分站，无需登录）
//   - 需登录路由（分站列表）
//   - 管理后台路由（需 tenant:manage 权限）：分站/员工/配置/域名 CRUD
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "tenant_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "tenant_write"))
	managePerm := coreRouter.WrapGin(middleware.RequirePermission("tenant:manage"))

	// ==================== 公开路由（C 端，无需登录） ====================
	// 根据域名获取当前分站
	router.GET("/stations/current", readLimiter, p.stationHandler.GetCurrent)

	// ==================== 需登录路由 ====================
	// 分站列表
	router.GET("/stations", auth, readLimiter, p.stationHandler.List)

	// ==================== 管理后台路由（需 tenant:manage 权限） ====================
	admin := router.Group("/admin")

	// 分站管理
	admin.GET("/stations", managePerm, p.stationHandler.List)
	admin.GET("/stations/:id", managePerm, p.stationHandler.GetByID)
	admin.POST("/stations", managePerm, writeLimiter, p.stationHandler.Create)
	admin.PUT("/stations/:id", managePerm, writeLimiter, p.stationHandler.Update)
	admin.DELETE("/stations/:id", managePerm, writeLimiter, p.stationHandler.Delete)
	admin.PUT("/stations/:id/status", managePerm, writeLimiter, p.stationHandler.UpdateStatus)
	admin.POST("/stations/copy-config", managePerm, writeLimiter, p.stationHandler.CopyConfig)

	// 员工管理
	admin.GET("/staff", managePerm, p.staffHandler.List)
	admin.GET("/staff/:id", managePerm, p.staffHandler.GetByID)
	admin.POST("/staff", managePerm, writeLimiter, p.staffHandler.Create)
	admin.PUT("/staff/:id", managePerm, writeLimiter, p.staffHandler.Update)
	admin.DELETE("/staff/:id", managePerm, writeLimiter, p.staffHandler.Delete)
	admin.GET("/staff/by-station/:station_id", managePerm, p.staffHandler.ListByStation)

	// 配置管理
	admin.GET("/configs", managePerm, p.configHandler.List)
	admin.GET("/configs/:id", managePerm, p.configHandler.GetByID)
	admin.POST("/configs", managePerm, writeLimiter, p.configHandler.Upsert)
	admin.PUT("/configs/:id", managePerm, writeLimiter, p.configHandler.Update)
	admin.DELETE("/configs/:id", managePerm, writeLimiter, p.configHandler.Delete)
	admin.GET("/configs/by-station-module", managePerm, p.configHandler.ListByStationAndModule)
	admin.POST("/configs/batch-get", managePerm, p.configHandler.BatchGet)

	// 域名管理
	admin.GET("/domains", managePerm, p.domainHandler.List)
	admin.GET("/domains/:id", managePerm, p.domainHandler.GetByID)
	admin.POST("/domains", managePerm, writeLimiter, p.domainHandler.Create)
	admin.PUT("/domains/:id", managePerm, writeLimiter, p.domainHandler.Update)
	admin.DELETE("/domains/:id", managePerm, writeLimiter, p.domainHandler.Delete)
	admin.PUT("/domains/:id/primary", managePerm, p.domainHandler.SetPrimary)
	admin.PUT("/domains/:id/ssl", managePerm, p.domainHandler.UpdateSSL)
	admin.GET("/domains/by-station/:station_id", managePerm, p.domainHandler.ListByStation)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
