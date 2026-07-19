// Package ai AI 智能中台精简版插件
// 依据 ershou 模块依赖：标题优化 + 描述生成 + 价格建议 + 摘要生成 + 任务管理
// 路由前缀 /api/v1/ai
package ai

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/ai/handler"
	"wuchang-tongcheng/internal/modules/ai/repository"
	"wuchang-tongcheng/internal/modules/ai/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin AI 中台插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建 AI 中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "ai", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "ai",
		DisplayName:  "AI智能中台",
		Category:     "middleware",
		Description:  "AI标题优化、描述生成、价格建议、摘要、任务管理",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注意：不调用 AutoMigrate，表结构由 migrations/005_p1_middlewares.sql 创建
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()
	repo := repository.NewAIRepository(db)
	svc := service.NewAIService(repo)
	p.handler = handler.NewHandler(svc)
	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/ai
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "ai_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "ai_write"))
	aiManage := coreRouter.WrapGin(middleware.RequirePermission("ai:manage"))

	// 任务
	router.POST("/tasks", auth, writeLimiter, p.handler.CreateTask)
	router.POST("/tasks/:task_id/run", auth, writeLimiter, p.handler.RunTask)
	router.GET("/tasks/:task_id", auth, readLimiter, p.handler.GetTask)
	router.GET("/tasks", auth, readLimiter, p.handler.ListTasks)

	// 模型管理（M 端）
	router.POST("/models", auth, aiManage, writeLimiter, p.handler.AddModel)
	router.GET("/models", auth, readLimiter, p.handler.ListModels)
	router.POST("/models/:id/status", auth, aiManage, writeLimiter, p.handler.UpdateModelStatus)

	// 提示词模板
	router.POST("/prompts", auth, aiManage, writeLimiter, p.handler.AddPrompt)
	router.GET("/prompts", auth, readLimiter, p.handler.ListPrompts)
	router.POST("/prompts/render", auth, readLimiter, p.handler.RenderPrompt)

	// 高级 AI 接口（业务侧）
	router.POST("/optimize-title", auth, writeLimiter, p.handler.OptimizeTitle)
	router.POST("/generate-description", auth, writeLimiter, p.handler.GenerateDescription)
	router.POST("/suggest-price", auth, writeLimiter, p.handler.SuggestPrice)

	// 生成记录
	router.GET("/generations", auth, readLimiter, p.handler.ListMyGenerations)
	router.POST("/generations/rate", auth, writeLimiter, p.handler.RateGeneration)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
