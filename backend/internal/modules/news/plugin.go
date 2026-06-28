// Package news 同城分类信息模块插件
// 提供本地分类信息的发布、浏览、管理、收藏、评论、消息通知
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

type Plugin struct {
	name     string
	version  string
	handler  *handler.Handler
	newsRepo repository.NewsRepository
}

func NewPlugin() *Plugin {
	return &Plugin{name: "news", version: "2.0.0"}
}

func (p *Plugin) Name() string    { return p.name }
func (p *Plugin) Version() string { return p.version }

func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	if err := db.AutoMigrate(
		&model.News{},
		&model.NewsLike{},
		&model.NewsFavorite{},
		&model.NewsComment{},
		&model.Message{},
	); err != nil {
		return err
	}

	newsRepo := repository.NewNewsRepository(db)
	idx := indexer.New()
	newsService := service.NewNewsService(newsRepo, idx)
	p.handler = handler.NewHandler(newsService)
	p.newsRepo = newsRepo

	if err := indexer.StartConsumer(ctx, newsRepo); err != nil {
		logger.Warn("news 索引消费者启动失败，业务不受影响", zap.Error(err))
	}

	return nil
}

func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "news"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "news_create"))
	likeLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "news_like"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "news_search"))

	// === 公开路由（无需登录，PC/小程序门户浏览） ===
	router.GET("", readLimiter, p.handler.List)
	router.GET("/search", searchLimiter, p.handler.Search)
	router.GET("/:id", readLimiter, p.handler.GetByID)
	router.GET("/:id/comments", p.handler.ListComments)

	// 点赞/收藏状态查询
	router.GET("/:id/like", p.handler.LikeStatus)
	router.GET("/:id/fav", p.handler.FavStatus)

	// === 需登录的路由 ===
	// 发布/编辑/删除分类信息
	router.POST("", auth, writeLimiter, p.handler.Create)
	router.PUT("/:id", auth, p.handler.Update)
	router.DELETE("/:id", auth, p.handler.Delete)

	// 点赞/收藏/评论
	router.POST("/:id/like", auth, likeLimiter, p.handler.Like)
	router.POST("/:id/fav", auth, likeLimiter, p.handler.Fav)
	router.POST("/:id/comments", auth, p.handler.CreateComment)
	router.DELETE("/comments/:id", auth, p.handler.DeleteComment)

	// === 消息通知 ===
	router.GET("/messages", auth, p.handler.ListMessages)
	router.GET("/messages/unread", auth, p.handler.UnreadCount)
	router.PUT("/messages/read", auth, p.handler.MarkRead)
}

func (p *Plugin) Close() error { return nil }

var _ plugin.Plugin = (*Plugin)(nil)