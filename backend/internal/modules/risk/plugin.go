// Package risk 风控审核中台精简版插件
// 依据 ershou 模块依赖：举报 + 敏感词 DFA + 内容审核 + 黑名单 + 风险分 + 违规处罚
// 路由前缀 /api/v1/risk
package risk

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/risk/handler"
	"wuchang-tongcheng/internal/modules/risk/repository"
	"wuchang-tongcheng/internal/modules/risk/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 风控中台插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建风控中台插件
func NewPlugin() *Plugin {
	return &Plugin{name: "risk", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:         "risk",
		DisplayName:  "风控审核中台",
		Category:     "middleware",
		Description:  "举报、敏感词DFA、内容审核、黑名单、风险分、违规处罚",
		Version:      p.version,
		Dependencies: []string{"user"},
		Author:       "wuchang",
	}
}

// Init 初始化插件
// 注意：不调用 AutoMigrate，表结构由 migrations/005_p1_middlewares.sql 创建
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()
	repo := repository.NewRiskRepository(db)
	svc := service.NewRiskService(repo)
	p.handler = handler.NewHandler(svc)
	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/risk
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "risk_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(20, 60, "risk_write"))
	riskManage := coreRouter.WrapGin(middleware.RequirePermission("risk:manage"))

	// 举报（用户）
	router.POST("/reports", auth, writeLimiter, p.handler.CreateReport)
	router.GET("/reports/:id", auth, readLimiter, p.handler.GetReport)
	// 举报管理（M 端）
	router.GET("/reports", auth, riskManage, readLimiter, p.handler.ListReports)
	router.POST("/reports/handle", auth, riskManage, writeLimiter, p.handler.HandleReport)

	// 敏感词管理（M 端）
	router.POST("/sensitive-words", auth, riskManage, writeLimiter, p.handler.AddSensitiveWord)
	router.DELETE("/sensitive-words/:id", auth, riskManage, writeLimiter, p.handler.DeleteSensitiveWord)
	router.GET("/sensitive-words", auth, riskManage, readLimiter, p.handler.ListSensitiveWords)

	// 文本/内容审核（公开给已登录模块调用）
	router.POST("/check-text", auth, readLimiter, p.handler.CheckText)
	router.POST("/audit", auth, readLimiter, p.handler.AuditContent)

	// 黑名单管理（M 端）
	router.POST("/blacklist", auth, riskManage, writeLimiter, p.handler.AddToBlacklist)
	router.POST("/blacklist/check", auth, readLimiter, p.handler.CheckBlacklist)
	router.GET("/blacklist", auth, riskManage, readLimiter, p.handler.ListBlacklist)

	// 用户风险分查询
	router.GET("/scores/:user_id", auth, readLimiter, p.handler.GetUserScore)

	// 违规处罚
	router.GET("/violations/:user_id", auth, readLimiter, p.handler.ListUserViolations)
	router.POST("/violations/appeal", auth, writeLimiter, p.handler.AppealViolation)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
