// Package shop 商家模块插件
// 提供店铺入驻、相册、评价及管理后台审核等业务
package shop

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/shop/handler"
	"wuchang-tongcheng/internal/modules/shop/model"
	"wuchang-tongcheng/internal/modules/shop/repository"
	"wuchang-tongcheng/internal/modules/shop/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 商家模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建商家模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "shop", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "shop",
		DisplayName:  "商家入驻",
		Category:     "business",
		Description:  "提供店铺入驻、相册、评价及管理后台审核等业务",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移店铺、相册、评价表
	if err := db.AutoMigrate(&model.Shop{}, &model.ShopImage{}, &model.ShopReview{}); err != nil {
		return err
	}

	// 初始化依赖链
	shopRepo := repository.NewShopRepository(db)
	imageRepo := repository.NewShopImageRepository(db)
	reviewRepo := repository.NewShopReviewRepository(db)
	shopService := service.NewShopService(shopRepo, imageRepo, reviewRepo)
	p.handler = handler.NewHandler(shopService)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/shop
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 公开接口（无需登录，全局 Auth 中间件以游客身份放行）
	router.GET("/list", p.handler.List)
	router.GET("/:id", p.handler.GetByID)
	router.GET("/:id/images", p.handler.GetImages)
	router.GET("/:id/reviews", p.handler.GetReviews)

	// 用户接口（需登录）
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	router.POST("/apply", auth, p.handler.Apply)
	router.GET("/my", auth, p.handler.GetMyShop)
	router.PUT("/my", auth, p.handler.UpdateMyShop)
	router.POST("/my/images", auth, p.handler.AddImage)
	router.DELETE("/my/images/:id", auth, p.handler.DeleteImage)
	router.POST("/:id/reviews", auth, p.handler.CreateReview)

	// 管理接口（需登录 + 权限）
	admin := router.Group("/admin")
	admin.GET("/list", coreRouter.WrapGin(middleware.RequirePermission("shop:read")), p.handler.AdminList)
	admin.PUT("/:id/audit", coreRouter.WrapGin(middleware.RequirePermission("shop:audit")), p.handler.AuditShop)
	admin.PUT("/:id/status", coreRouter.WrapGin(middleware.RequirePermission("shop:update")), p.handler.UpdateShopStatus)
	admin.PUT("/:id/recommend", coreRouter.WrapGin(middleware.RequirePermission("shop:update")), p.handler.SetRecommend)
	admin.DELETE("/:id", coreRouter.WrapGin(middleware.RequirePermission("shop:delete")), p.handler.DeleteShop)
	admin.GET("/reviews", coreRouter.WrapGin(middleware.RequirePermission("shop:read")), p.handler.AdminReviewList)
	admin.PUT("/reviews/:id/audit", coreRouter.WrapGin(middleware.RequirePermission("shop:audit")), p.handler.AuditReview)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
