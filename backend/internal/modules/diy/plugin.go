// Package diy DIY 前端页面中台模块插件
// 提供可视化页面装修：页面/组件/模板/统计 4 子域
// 依据架构设计 4.12：拖拽生成首页/专题页/店铺页/活动页
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package diy

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/diy/handler"
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/modules/diy/repository"
	"wuchang-tongcheng/internal/modules/diy/service"
	"wuchang-tongcheng/internal/pkg/database"
	"wuchang-tongcheng/internal/pkg/utils"
)

// Plugin DIY 前端页面中台模块插件
type Plugin struct {
	name    string
	version string

	// 4 个 Handler（按子域组织）
	pageHandler      *handler.PageHandler
	componentHandler *handler.ComponentHandler
	templateHandler  *handler.TemplateHandler
	statHandler      *handler.StatHandler
}

// NewPlugin 创建 DIY 前端页面中台模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "diy", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "diy",
		DisplayName: "DIY 中台",
		Category:    "middleware",
		Description: "DIY 前端页面中台完整功能：可视化页面装修/组件库/模板管理/页面统计",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：4 张表由 backend/migrations/031_diy_full.sql 创建，
// 此处不调用 AutoMigrate，避免与 SQL 脚本约束名不一致。
// 仅 AutoMigrate 主表 Page/PageStat/Component/Template 以保证 GORM 软删除字段可用。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// AutoMigrate 4 张表（migration 脚本已建表，此处幂等保证字段完整）
	if err := db.AutoMigrate(
		&model.Page{},
		&model.Component{},
		&model.Template{},
		&model.PageStat{},
	); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（4 个） =====
	pageRepo := repository.NewPageRepository(db)
	componentRepo := repository.NewComponentRepository(db)
	templateRepo := repository.NewTemplateRepository(db)
	statRepo := repository.NewStatRepository(db)

	// ===== 依赖注入：Service 层（4 个） =====
	pageSvc := service.NewPageService(pageRepo)
	componentSvc := service.NewComponentService(componentRepo)
	templateSvc := service.NewTemplateService(templateRepo, pageRepo)
	statSvc := service.NewStatService(statRepo)

	// ===== 依赖注入：Handler 层（4 个） =====
	p.pageHandler = handler.NewPageHandler(pageSvc)
	p.componentHandler = handler.NewComponentHandler(componentSvc)
	p.templateHandler = handler.NewTemplateHandler(templateSvc, pageSvc)
	p.statHandler = handler.NewStatHandler(statSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/diy
//
// 路由分组（共 30+ API）：
//   - 公开路由（C 端浏览，无需登录）：按 slug 获取已发布页面/页面列表/统计记录
//   - 需登录路由（C 端用户操作）：页面 CRUD/发布/下线/复制/我的页面/应用模板
//   - 管理后台路由（需 diy:manage 权限）：组件/模板/统计/全部页面管理
//
// 注意：固定路径（/by-slug/:slug /mine /page/:id 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "diy_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "diy_write"))
	managePerm := coreRouter.WrapGin(middleware.RequirePermission("diy:manage"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 页面 - 公开按 slug 获取已发布页面
	router.GET("/pages/by-slug/:slug", readLimiter, p.pageHandler.GetBySlug)

	// 组件 - 公开按分类获取（C 端编辑器使用）
	router.GET("/components/by-category/:category", readLimiter, p.componentHandler.ListByCategory)

	// 统计 - 公开记录（C 端埋点）
	router.POST("/stats/view", p.statHandler.RecordView)
	router.POST("/stats/click", p.statHandler.RecordClick)
	router.POST("/stats/conversion", p.statHandler.RecordConversion)

	// ==================== 需登录路由（C 端用户操作） ====================

	// 页面 CRUD（C 端用户）
	router.GET("/pages/mine", auth, readLimiter, p.pageHandler.ListMine)
	router.POST("/pages", auth, writeLimiter, p.pageHandler.Create)
	router.PUT("/pages/:id", auth, writeLimiter, p.pageHandler.Update)
	router.DELETE("/pages/:id", auth, writeLimiter, p.pageHandler.Delete)
	router.POST("/pages/:id/publish", auth, writeLimiter, p.pageHandler.Publish)
	router.POST("/pages/:id/offline", auth, writeLimiter, p.pageHandler.Offline)
	router.POST("/pages/:id/copy", auth, writeLimiter, p.pageHandler.Copy)

	// 模板 - 应用模板创建新页面（需登录）
	router.POST("/templates/:id/apply", auth, writeLimiter, p.templateHandler.Apply)

	// ==================== 管理后台路由（需 diy:manage 权限） ====================

	admin := router.Group("/admin")

	// 页面管理（全部页面）
	admin.GET("/pages", managePerm, p.pageHandler.AdminList)
	admin.GET("/pages/:id", managePerm, p.pageHandler.AdminGetByID)
	admin.PUT("/pages/:id/status", managePerm, writeLimiter, p.pageHandler.AdminUpdateStatus)

	// 组件管理
	admin.GET("/components", managePerm, p.componentHandler.List)
	admin.GET("/components/:id", managePerm, p.componentHandler.GetByID)
	admin.POST("/components", managePerm, writeLimiter, p.componentHandler.Create)
	admin.PUT("/components/:id", managePerm, writeLimiter, p.componentHandler.Update)
	admin.DELETE("/components/:id", managePerm, writeLimiter, p.componentHandler.Delete)

	// 模板管理
	admin.GET("/templates", managePerm, p.templateHandler.List)
	admin.GET("/templates/:id", managePerm, p.templateHandler.GetByID)
	admin.POST("/templates", managePerm, writeLimiter, p.templateHandler.Create)
	admin.PUT("/templates/:id", managePerm, writeLimiter, p.templateHandler.Update)
	admin.DELETE("/templates/:id", managePerm, writeLimiter, p.templateHandler.Delete)

	// 统计管理
	admin.GET("/stats/page/:id", managePerm, p.statHandler.ListByPageID)
	admin.GET("/stats/date-range", managePerm, p.statHandler.ListByDateRange)
	admin.GET("/stats/summary/page/:id", managePerm, p.statHandler.SumByPageID)
	admin.GET("/stats/summary", managePerm, p.statHandler.SumByDateRange)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
// 同时注册本模块错误码到 utils（6001-6030 区间，未在 error_code.go 中声明）
func init() {
	// 注册 DIY 中台错误码消息（避开 error_code.go 修改）
	registerDiyCodes()
	plugin.AutoRegister(NewPlugin())
}

// registerDiyCodes 注册 DIY 中台错误码到 utils.RegisterCode
// 错误码区间 6001-6030，与 handler/codes.go 中的常量保持一致
func registerDiyCodes() {
	// page 子域
	utils.RegisterCode(handler.CodeDiyPageError, "页面错误")
	utils.RegisterCode(handler.CodeDiyPageNotFound, "页面不存在")
	utils.RegisterCode(handler.CodeDiyPageNoPermission, "无权操作页面")
	utils.RegisterCode(handler.CodeDiyPageStatusInvalid, "页面状态不允许此操作")
	utils.RegisterCode(handler.CodeDiyPageSlugConflict, "页面 slug 已被占用")
	utils.RegisterCode(handler.CodeDiyPageSlugEmpty, "已发布页面必须设置 slug")

	// component 子域
	utils.RegisterCode(handler.CodeDiyComponentError, "组件错误")
	utils.RegisterCode(handler.CodeDiyComponentNotFound, "组件不存在")
	utils.RegisterCode(handler.CodeDiyComponentCodeConflict, "组件编码已存在")

	// template 子域
	utils.RegisterCode(handler.CodeDiyTemplateError, "模板错误")
	utils.RegisterCode(handler.CodeDiyTemplateNotFound, "模板不存在")

	// stat 子域
	utils.RegisterCode(handler.CodeDiyStatError, "统计错误")
	utils.RegisterCode(handler.CodeDiyStatNotFound, "统计记录不存在")
}
