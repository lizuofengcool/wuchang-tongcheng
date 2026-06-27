// Package news 同城头条模块插件
// 提供本地资讯/同城头条的发布、浏览、管理
package news

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/news/handler"
	"wuchang-tongcheng/internal/modules/news/indexer"
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/modules/news/repository"
	"wuchang-tongcheng/internal/modules/news/service"
	"wuchang-tongcheng/internal/pkg/database"
	"wuchang-tongcheng/internal/pkg/logger"

	"go.uber.org/zap"
)

// Plugin 同城头条模块插件
type Plugin struct {
	name     string
	version  string
	handler  *handler.Handler
	newsRepo repository.NewsRepository
}

// NewPlugin 创建同城头条模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "news", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移头条表 + 点赞记录表
	if err := db.AutoMigrate(&model.News{}, &model.NewsLike{}); err != nil {
		return err
	}

	// 初始化依赖链：repo -> indexer -> service -> handler
	newsRepo := repository.NewNewsRepository(db)
	idx := indexer.New()
	newsService := service.NewNewsService(newsRepo, idx)
	p.handler = handler.NewHandler(newsService)
	p.newsRepo = newsRepo

	// MQ 可用时启动 news.es.index 消费者，异步同步 ES
	if err := indexer.StartConsumer(ctx, newsRepo); err != nil {
		logger.Warn("news 索引消费者启动失败，业务不受影响", zap.Error(err))
	}

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 需要登录的接口
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	// 访问限流：单 IP 每分钟最多 60 次，防止恶意刷浏览量
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "news"))
	// 点赞限流：单 IP 每分钟最多 30 次
	likeLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "news_like"))
	// 检索限流：单 IP 每分钟最多 30 次
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "news_search"))

	router.POST("", coreRouter.WrapGin(middleware.RequirePermission("news:create")), p.handler.Create)
	router.PUT("/:id", coreRouter.WrapGin(middleware.RequirePermission("news:update")), p.handler.Update)
	router.DELETE("/:id", coreRouter.WrapGin(middleware.RequirePermission("news:delete")), p.handler.Delete)
	router.GET("/:id", readLimiter, coreRouter.WrapGin(middleware.RequirePermission("news:read")), p.handler.GetByID)
	router.GET("", readLimiter, coreRouter.WrapGin(middleware.RequirePermission("news:read")), p.handler.List)

	// 全文检索：仅需登录（浏览用户也能搜）
	// ES 优先，ES 不可用时 service 内部降级到 DB LIKE
	router.GET("/search", searchLimiter, coreRouter.WrapGin(middleware.RequirePermission("news:read")), p.handler.Search)

	// 点赞：仅需登录（浏览用户也能点赞）
	router.POST("/:id/like", auth, likeLimiter, p.handler.Like)
	router.GET("/:id/like", auth, p.handler.LikeStatus)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
