// Package category 分类信息模块插件
// 提供本地生活分类管理（如：二手房、招聘、二手物品、车辆等）
package category

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/category/handler"
	"wuchang-tongcheng/internal/modules/category/model"
	"wuchang-tongcheng/internal/modules/category/repository"
	"wuchang-tongcheng/internal/modules/category/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 分类信息模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建分类信息模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "category", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移分类表
	if err := db.AutoMigrate(&model.Category{}); err != nil {
		return err
	}

	// 初始化依赖链
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	p.handler = handler.NewHandler(categoryService)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 公开路由（无需鉴权，PC/小程序门户使用）
	router.GET("", p.handler.GetAll)

	router.POST("", coreRouter.WrapGin(middleware.RequirePermission("category:create")), p.handler.Create)
	router.PUT("/:id", coreRouter.WrapGin(middleware.RequirePermission("category:update")), p.handler.Update)
	router.DELETE("/:id", coreRouter.WrapGin(middleware.RequirePermission("category:delete")), p.handler.Delete)
	router.GET("/:id", coreRouter.WrapGin(middleware.RequirePermission("category:read")), p.handler.GetByID)
	router.GET("/children", coreRouter.WrapGin(middleware.RequirePermission("category:read")), p.handler.GetByParentID)
	router.GET("/tree", coreRouter.WrapGin(middleware.RequirePermission("category:read")), p.handler.GetTree)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
