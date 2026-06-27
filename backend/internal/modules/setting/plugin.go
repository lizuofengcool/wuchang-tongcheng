// Package setting 系统设置模块插件
// 提供KV配置存储，按group分组，支持多地区
package setting

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/setting/handler"
	"wuchang-tongcheng/internal/modules/setting/model"
	"wuchang-tongcheng/internal/modules/setting/repository"
	"wuchang-tongcheng/internal/modules/setting/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 系统设置模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建系统设置模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "setting", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移设置表
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		return err
	}

	// 初始化依赖链
	repo := repository.NewSettingRepository(db)
	svc := service.NewSettingService(repo)
	p.handler = handler.NewHandler(svc)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	router.POST("", coreRouter.WrapGin(middleware.RequirePermission("setting:create")), p.handler.Create)
	router.PUT("/:id", coreRouter.WrapGin(middleware.RequirePermission("setting:update")), p.handler.Update)
	router.DELETE("/:id", coreRouter.WrapGin(middleware.RequirePermission("setting:delete")), p.handler.Delete)
	router.GET("/:id", coreRouter.WrapGin(middleware.RequirePermission("setting:read")), p.handler.GetByID)
	router.GET("/group/:group", coreRouter.WrapGin(middleware.RequirePermission("setting:read")), p.handler.GetByGroup)
	router.GET("", coreRouter.WrapGin(middleware.RequirePermission("setting:read")), p.handler.GetAll)
	router.PUT("/batch", coreRouter.WrapGin(middleware.RequirePermission("setting:update")), p.handler.BatchUpdate)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
