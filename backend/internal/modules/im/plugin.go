// Package im IM 消息中台精简版插件
// 依据 ershou 模块依赖：私聊 + 系统通知 + 隐私号码
// 路由前缀 /api/v1/im
package im

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/im/handler"
	"wuchang-tongcheng/internal/modules/im/repository"
	"wuchang-tongcheng/internal/modules/im/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin IM 中台插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建 IM 中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "im", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "im",
		DisplayName:  "IM消息中台",
		Category:     "middleware",
		Description:  "私聊会话、消息、系统通知、隐私号码",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()
	repo := repository.NewIMRepository(db)
	svc := service.NewIMService(repo)
	p.handler = handler.NewHandler(svc)
	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/im
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "im_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(20, 60, "im_write"))

	// 会话
	router.POST("/sessions", auth, writeLimiter, p.handler.CreateSession)
	router.GET("/sessions", auth, readLimiter, p.handler.ListSessions)
	router.GET("/sessions/:session_id", auth, readLimiter, p.handler.GetSession)

	// 消息
	router.POST("/messages", auth, writeLimiter, p.handler.SendMessage)
	router.GET("/sessions/:session_id/messages", auth, readLimiter, p.handler.GetHistory)
	router.POST("/sessions/:session_id/read", auth, writeLimiter, p.handler.MarkRead)

	// 系统通知
	router.GET("/notifications", auth, readLimiter, p.handler.ListNotifications)
	router.GET("/notifications/unread", auth, readLimiter, p.handler.ListUnreadNotifications)
	router.POST("/notifications/read-all", auth, writeLimiter, p.handler.MarkAllNotificationsRead)

	// 隐私号码
	router.POST("/privacy-numbers", auth, writeLimiter, p.handler.BindPrivacyNumber)
	router.POST("/privacy-numbers/unbind", auth, writeLimiter, p.handler.UnbindPrivacyNumber)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件
func init() {
	plugin.AutoRegister(NewPlugin())
}
