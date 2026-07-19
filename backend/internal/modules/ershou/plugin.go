// Package ershou 同城二手物品模块插件
// 提供二手物品的发布、浏览、搜索、附近、留言、收藏及管理后台审核等业务
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package ershou

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/ershou/handler"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/modules/ershou/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 二手物品模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建二手物品模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "ershou", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "ershou",
		DisplayName:  "同城二手",
		Category:     "business",
		Description:  "提供二手物品的发布、浏览、搜索、附近、留言、收藏及管理后台审核等业务",
		Version:      p.version,
		Dependencies: []string{"category"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注册表自动迁移 + 注入依赖链 repository → service → handler
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移：主表 + 图片子表 + 收藏关联表 + 留言表
	if err := db.AutoMigrate(
		&model.Ershou{},
		&model.ErshouImage{},
		&model.ErshouFavorite{},
		&model.ErshouMessage{},
	); err != nil {
		return err
	}

	// 依赖注入
	ershouRepo := repository.NewErshouRepository(db)
	ershouService := service.NewErshouService(ershouRepo)
	p.handler = handler.NewHandler(ershouService)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/ershou
//
// 路由分组：
//   - 公开路由（无需登录，C 端浏览）：List/Search/Nearby/GetByID/ListMessages/FavStatus
//   - 需登录路由（C 端发布/收藏/留言）：Create/Update/Delete/ListMine/Fav/ListFavs/CreateMessage
//   - 管理后台路由（需 content:audit 权限）：AdminList/AdminGetByID/Audit/AdminUpdateStatus
//
// 注意：固定路径（/search /nearby /favorites /mine）必须注册在 /:id 之前，
// 否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "ershou_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "ershou_write"))
	favLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_fav"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "ershou_nearby"))

	// === 公开路由（C 端浏览，无需登录） ===
	router.GET("", readLimiter, p.handler.List)
	router.GET("/search", searchLimiter, p.handler.Search)
	// /nearby、/favorites、/mine 需注册在 /:id 之前，避免被 :id 参数路由吞掉
	router.GET("/nearby", nearbyLimiter, p.handler.Nearby)
	router.GET("/:id", readLimiter, p.handler.GetByID)
	router.GET("/:id/messages", readLimiter, p.handler.ListMessages)
	router.GET("/:id/fav", p.handler.FavStatus)

	// === 需登录路由（C 端发布/收藏/留言） ===
	router.POST("", auth, writeLimiter, p.handler.Create)
	router.PUT("/:id", auth, p.handler.Update)
	router.DELETE("/:id", auth, p.handler.Delete)

	router.GET("/mine", auth, p.handler.ListMine)
	router.GET("/favorites", auth, readLimiter, p.handler.ListFavs)

	router.POST("/:id/fav", auth, favLimiter, p.handler.Fav)
	router.POST("/:id/messages", auth, writeLimiter, p.handler.CreateMessage)

	// === 管理后台路由（需 content:audit 权限，auditor + super_admin 可访问） ===
	admin := router.Group("/admin")
	admin.GET("/list", coreRouter.WrapGin(middleware.RequirePermission("content:audit")), p.handler.AdminList)
	admin.GET("/:id", coreRouter.WrapGin(middleware.RequirePermission("content:audit")), p.handler.AdminGetByID)
	admin.PUT("/:id/audit", coreRouter.WrapGin(middleware.RequirePermission("content:audit")), p.handler.Audit)
	admin.PUT("/:id/status", coreRouter.WrapGin(middleware.RequirePermission("content:audit")), p.handler.AdminUpdateStatus)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
