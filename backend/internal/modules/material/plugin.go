// Package material 素材存储中台精简版插件
// 依据 ershou 模块依赖：图片/视频 + 以图搜图 + 缩略图 + 水印
// 路由前缀 /api/v1/material
package material

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/material/handler"
	"wuchang-tongcheng/internal/modules/material/repository"
	"wuchang-tongcheng/internal/modules/material/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 素材中台插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建素材中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "material", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "material",
		DisplayName:  "素材存储中台",
		Category:     "middleware",
		Description:  "图片/视频上传、缩略图、水印、以图搜图",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注意：不调用 AutoMigrate，表结构由 migrations/005_p1_middlewares.sql 创建
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()
	repo := repository.NewMaterialRepository(db)
	svc := service.NewMaterialService(repo)
	p.handler = handler.NewHandler(svc)
	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/material
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "material_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(20, 60, "material_write"))

	// 文件上传/查询/删除
	router.POST("/files", auth, writeLimiter, p.handler.Upload)
	router.GET("/files", auth, readLimiter, p.handler.ListFiles)
	router.GET("/files/:file_id", auth, readLimiter, p.handler.GetFile)
	router.DELETE("/files/:file_id", auth, writeLimiter, p.handler.DeleteFile)

	// 以图搜图
	router.POST("/search-by-image", auth, readLimiter, p.handler.SearchByImage)

	// 图片处理
	router.POST("/watermark", auth, writeLimiter, p.handler.AddWatermark)
	router.POST("/thumbnails", auth, writeLimiter, p.handler.GenerateThumbnail)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
