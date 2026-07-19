// Package module 模块注册表插件
// 模块管理自身的 Plugin 实现，启动时同步所有插件元信息到 modules 表
package module

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/pkg/database"
	"wuchang-tongcheng/internal/pkg/logger"

	"go.uber.org/zap"
)

// Plugin 模块管理插件
type Plugin struct {
	name    string
	version string
	handler *Handler
}

// NewPlugin 创建模块管理插件
func NewPlugin() *Plugin {
	return &Plugin{name: "modules", version: "1.0.0"}
}

// Name 返回插件名称（复数，路由前缀 /api/v1/modules）
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "modules",
		DisplayName:  "模块管理",
		Category:     "system",
		Description:  "模块注册表与开关管理，提供模块列表查询、启停、元信息更新",
		Version:      p.version,
		Dependencies: []string{},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 1. 检查 modules 表是否已通过 SQL 迁移脚本创建（避免与 GORM AutoMigrate 约束名冲突）
// 2. 初始化依赖链 repository → service → handler
// 3. 调用 SyncAllFromManager 将所有已注册插件同步到 modules 表
//
// 注意：modules 表由 backend/migrations/001_p0_baseline.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则（如 modules_name_key），
// 不使用 GORM AutoMigrate 以避免约束名不一致导致的 DROP CONSTRAINT 错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 初始化依赖链
	repo := NewRepository(db)
	svc := NewService(repo)
	p.handler = NewHandler(svc)

	// 同步所有已注册插件到 modules 表
	// 此时所有插件的 init() 已执行完毕（modules 包导入顺序保证），
	// manager 中已包含全部插件，可安全同步
	if err := svc.SyncAllFromManager(); err != nil {
		logger.Warn("modules 表同步失败，模块总控面板可能展示不全", zap.Error(err))
		// 不阻塞启动，运维可手动重新同步
	} else {
		logger.Info("modules 表同步完成")
	}

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/modules
//
// 路由分组：
//   - 查询类（需登录）：List / GetByName
//   - 操作类（需 module:manage 权限）：Enable / Disable / Update
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	manage := coreRouter.WrapGin(middleware.RequirePermission("module:manage"))

	// 查询类（需登录）
	router.GET("", auth, p.handler.List)
	router.GET("/:name", auth, p.handler.GetByName)

	// 操作类（需 module:manage 权限）
	router.POST("/:name/enable", manage, p.handler.Enable)
	router.POST("/:name/disable", manage, p.handler.Disable)
	router.PUT("/:name", manage, p.handler.Update)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)
