// Package ai AI 智能中台精简版插件
// 依据 ershou 模块依赖：标题优化 + 描述生成 + 价格建议 + 摘要生成 + 任务管理
// 扩展：审核/推荐/对话/模型配置/训练数据
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
	extH    *handler.ExtendHandler
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
		Description:  "AI标题优化、描述生成、价格建议、摘要、任务管理、审核、推荐、对话",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注意：不调用 AutoMigrate，表结构由 migrations/016_ai_full.sql 创建
// 不使用 GORM AutoMigrate 以避免约束名不一致错误（与 car/house/job 设计一致）
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 原有依赖注入
	repo := repository.NewAIRepository(db)
	svc := service.NewAIService(repo)
	p.handler = handler.NewHandler(svc)

	// 扩展依赖注入
	extRepo := repository.NewAIExtendRepository(db)
	extSvc := service.NewAIExtendService(extRepo)
	p.extH = handler.NewExtendHandler(extSvc)

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

	// ===== 扩展路由 =====

	// 审核
	router.POST("/audit/submit", auth, writeLimiter, p.extH.SubmitAudit)
	router.GET("/audit/:task_id", auth, readLimiter, p.extH.GetAuditResult)
	router.GET("/audit", auth, readLimiter, p.extH.ListAuditResults)
	router.PUT("/audit/:task_id/review", auth, aiManage, writeLimiter, p.extH.ReviewAudit)

	// 推荐
	router.GET("/recommendations", auth, readLimiter, p.extH.GetRecommendations)
	router.POST("/recommendations/:id/click", auth, writeLimiter, p.extH.TrackClick)
	router.POST("/recommendations/:id/dwell", auth, writeLimiter, p.extH.TrackDwell)
	router.POST("/recommendations/:id/feedback", auth, writeLimiter, p.extH.FeedbackRecommendation)

	// 模型配置（M 端）— 注意路径与上方 /models/:id/status 不冲突
	router.GET("/model-configs", auth, readLimiter, p.extH.ListModelConfigs)
	router.POST("/model-configs", auth, aiManage, writeLimiter, p.extH.UpsertModelConfig)
	router.DELETE("/model-configs/:id", auth, aiManage, writeLimiter, p.extH.DeleteModelConfig)

	// 对话
	router.POST("/chat/sessions", auth, writeLimiter, p.extH.CreateChatSession)
	router.GET("/chat/sessions", auth, readLimiter, p.extH.ListChatSessions)
	router.POST("/chat/messages", auth, writeLimiter, p.extH.Chat)
	router.GET("/chat/sessions/:session_id/messages", auth, readLimiter, p.extH.ListChatMessages)
	router.POST("/chat/messages/:id/feedback", auth, writeLimiter, p.extH.MessageFeedback)

	// 训练数据（M 端）
	router.GET("/training-data", auth, aiManage, readLimiter, p.extH.ListTrainingData)
	router.POST("/training-data", auth, aiManage, writeLimiter, p.extH.CreateTrainingData)
	router.PUT("/training-data/:id/used", auth, aiManage, writeLimiter, p.extH.MarkTrainingDataUsed)

	// 统计（M 端）
	router.GET("/admin/statistics", auth, aiManage, readLimiter, p.extH.GetStatistics)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
