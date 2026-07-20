// Package dh114 同城114模块插件
// 提供商户/商户详情/营业时间/菜单/分类/优惠券/团购/评价/收藏/电话本/推荐/统计/审核规则/认证等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
package dh114

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/dh114/handler"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/modules/dh114/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 同城114模块插件
type Plugin struct {
	name    string
	version string

	// 12 个 Handler
	dh114Handler          *handler.Handler
	businessHandler       *handler.BusinessHandler
	categoryHandler      *handler.CategoryHandler
	couponHandler        *handler.CouponHandler
	groupbuyHandler      *handler.GroupbuyHandler
	reviewHandler        *handler.ReviewHandler
	favoriteHandler      *handler.FavoriteHandler
	phoneCallHandler     *handler.PhoneCallHandler
	auditRuleHandler     *handler.AuditRuleHandler
	recommendationHandler *handler.RecommendationHandler
	statisticHandler     *handler.StatisticHandler
	verificationHandler  *handler.VerificationHandler
}

// NewPlugin 创建同城114模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "dh114", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "dh114",
		DisplayName: "同城114",
		Category:    "business",
		Description: "同城114完整功能：商户/商户详情/营业时间/菜单/分类/优惠券/团购/评价/收藏/电话本/推荐/统计/审核规则/认证",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：18 张表由 backend/migrations/023_dh114_full.sql 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
// 此处仅 AutoMigrate 主表 Dh114（与 car/pinche 模块保持一致），子表由 SQL 脚本创建。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 dh114s 主表（子表由 backend/migrations/023_dh114_full.sql 创建）
	if err := db.AutoMigrate(&model.Dh114{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（18 个） =====
	dh114Repo := repository.NewDh114Repository(db)
	imageRepo := repository.NewImageRepository(db)
	visitRepo := repository.NewVisitRepository(db)
	tagRepo := repository.NewTagRepository(db)
	businessRepo := repository.NewBusinessRepository(db)
	businessHourRepo := repository.NewBusinessHourRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	groupbuyRepo := repository.NewGroupbuyRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	reviewReplyRepo := repository.NewReviewReplyRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	phoneCallRepo := repository.NewPhoneCallRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	verificationRepo := repository.NewVerificationRepository(db)

	// ===== 依赖注入：Service 层（12 个） =====
	dh114Svc := service.NewDh114Service(dh114Repo, imageRepo, visitRepo, favoriteRepo, phoneCallRepo)
	businessSvc := service.NewBusinessService(businessRepo, businessHourRepo, menuRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	couponSvc := service.NewCouponService(couponRepo)
	groupbuySvc := service.NewGroupbuyService(groupbuyRepo)
	reviewSvc := service.NewReviewService(reviewRepo, reviewReplyRepo, dh114Repo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo)
	phoneCallSvc := service.NewPhoneCallService(phoneCallRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	recommendationSvc := service.NewRecommendationService(recommendationRepo)
	statisticSvc := service.NewStatisticService(statisticRepo)
	verificationSvc := service.NewVerificationService(verificationRepo)

	// ===== 依赖注入：Handler 层（12 个） =====
	p.dh114Handler = handler.NewHandler(dh114Svc)
	p.businessHandler = handler.NewBusinessHandler(businessSvc)
	p.categoryHandler = handler.NewCategoryHandler(categorySvc)
	p.couponHandler = handler.NewCouponHandler(couponSvc)
	p.groupbuyHandler = handler.NewGroupbuyHandler(groupbuySvc)
	p.reviewHandler = handler.NewReviewHandler(reviewSvc)
	p.favoriteHandler = handler.NewFavoriteHandler(favoriteSvc)
	p.phoneCallHandler = handler.NewPhoneCallHandler(phoneCallSvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)
	p.recommendationHandler = handler.NewRecommendationHandler(recommendationSvc)
	p.statisticHandler = handler.NewStatisticHandler(statisticSvc)
	p.verificationHandler = handler.NewVerificationHandler(verificationSvc)

	// 避免未使用变量告警（tagRepo 暂未在 service 层使用，保留以备后续扩展）
	_ = tagRepo

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/dh114
//
// 路由分组（共 100+ API）：
//   - 公开路由（C 端浏览，无需登录）：商户列表/详情/搜索/附近/分类/优惠券/团购/评价/电话本/推荐/统计只读
//   - 需登录路由（C 端发布/收藏/交易）：商户 CRUD/商户详情/营业时间/菜单/优惠券/团购/评价/收藏/电话本/认证
//   - 管理后台路由（需 dh114:audit 权限）：审核/批量/状态/统计/审核规则/推荐管理
//
// 注意：固定路径（/search /nearby /mine /hot /favs 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "dh114_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "dh114_write"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "dh114_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "dh114_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("dh114:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 商户主表（dh114s）
	router.GET("", readLimiter, p.dh114Handler.List)
	router.GET("/search", searchLimiter, p.dh114Handler.Search)
	router.GET("/advanced-search", searchLimiter, p.dh114Handler.AdvancedSearch)
	router.GET("/nearby", nearbyLimiter, p.dh114Handler.ListNearby)
	router.GET("/:id", readLimiter, p.dh114Handler.GetByID)
	router.POST("/:id/share", p.dh114Handler.IncrShare)
	router.POST("/:id/contact", p.dh114Handler.IncrContact)
	router.POST("/views", p.dh114Handler.RecordView)
	router.POST("/:id/calls", p.dh114Handler.RecordCall)

	// 商户详情（businesses）
	router.GET("/businesses/:id", readLimiter, p.businessHandler.GetByID)
	router.GET("/:id/business", readLimiter, p.businessHandler.GetByDh114ID)
	router.GET("/businesses", readLimiter, p.businessHandler.List)

	// 营业时间（business-hours）
	router.GET("/:id/business-hours", readLimiter, p.businessHandler.ListBusinessHours)

	// 菜单（menus）
	router.GET("/menus", readLimiter, p.businessHandler.ListMenus)
	router.GET("/menus/:id", readLimiter, p.businessHandler.GetMenuByID)
	router.GET("/:id/menus", readLimiter, p.businessHandler.ListMenusByDh114)
	router.GET("/:id/menus/signature", readLimiter, p.businessHandler.ListSignatureMenus)

	// 分类（categories）
	router.GET("/categories", readLimiter, p.categoryHandler.List)
	router.GET("/categories/by-parent", readLimiter, p.categoryHandler.ListByParent)
	router.GET("/categories/by-level", readLimiter, p.categoryHandler.ListByLevel)
	router.GET("/categories/by-business-type", readLimiter, p.categoryHandler.ListByBusinessType)
	router.GET("/categories/:id", readLimiter, p.categoryHandler.GetByID)

	// 优惠券（coupons）
	router.GET("/coupons", readLimiter, p.couponHandler.List)
	router.GET("/coupons/hot", readLimiter, p.couponHandler.ListHot)
	router.GET("/coupons/:id", readLimiter, p.couponHandler.GetByID)
	router.GET("/:id/coupons", readLimiter, p.couponHandler.ListByDh114)

	// 团购（groupbuys）
	router.GET("/groupbuys", readLimiter, p.groupbuyHandler.List)
	router.GET("/groupbuys/hot", readLimiter, p.groupbuyHandler.ListHot)
	router.GET("/groupbuys/:id", readLimiter, p.groupbuyHandler.GetByID)
	router.GET("/:id/groupbuys", readLimiter, p.groupbuyHandler.ListByDh114)

	// 评价（reviews）
	router.GET("/reviews", readLimiter, p.reviewHandler.List)
	router.GET("/reviews/stats", readLimiter, p.reviewHandler.Stats)
	router.GET("/reviews/:id", readLimiter, p.reviewHandler.GetByID)
	router.GET("/reviews/:id/replies", readLimiter, p.reviewHandler.ListReplies)
	router.GET("/:id/reviews", readLimiter, p.reviewHandler.ListByDh114)

	// 电话本（phone-calls）
	router.GET("/phone-calls/:id", readLimiter, p.phoneCallHandler.GetByID)
	router.GET("/phone-calls", readLimiter, p.phoneCallHandler.List)
	router.GET("/:id/phone-calls", readLimiter, p.phoneCallHandler.ListByDh114)
	router.GET("/:id/phone-calls/count", readLimiter, p.phoneCallHandler.CountByDh114)
	router.GET("/:id/phone-calls/today-count", readLimiter, p.phoneCallHandler.CountTodayByDh114)

	// 审核规则（audit-rules）- 公开只读
	router.GET("/audit-rules", readLimiter, p.auditRuleHandler.List)
	router.GET("/audit-rules/enabled", readLimiter, p.auditRuleHandler.ListEnabled)
	router.GET("/audit-rules/type/:rule_type", readLimiter, p.auditRuleHandler.ListByType)
	router.GET("/audit-rules/:id", readLimiter, p.auditRuleHandler.GetByID)

	// 推荐（recommendations）- 公开只读
	router.GET("/recommendations", readLimiter, p.recommendationHandler.ListByType)
	router.GET("/recommendations/:id", readLimiter, p.recommendationHandler.GetByID)
	router.GET("/:id/recommendations", readLimiter, p.recommendationHandler.ListByDh114)

	// 数据统计（公开部分：日期区间/商户/分类/汇总/热门/分类热门）
	router.GET("/statistics/date-range", readLimiter, p.statisticHandler.ListByDateRange)
	router.GET("/statistics/by-category", readLimiter, p.statisticHandler.ListByCategory)
	router.GET("/statistics/hot-business", readLimiter, p.statisticHandler.HotBusiness)
	router.GET("/statistics/hot-categories", readLimiter, p.statisticHandler.HotCategories)
	router.GET("/:id/statistics", readLimiter, p.statisticHandler.ListByDh114)
	router.GET("/:id/statistics/summary", readLimiter, p.statisticHandler.SumByDh114)

	// ==================== 需登录路由（C 端发布/收藏/交易） ====================

	// 商户主表 CRUD
	router.POST("", auth, writeLimiter, p.dh114Handler.Create)
	router.PUT("/:id", auth, writeLimiter, p.dh114Handler.Update)
	router.DELETE("/:id", auth, writeLimiter, p.dh114Handler.Delete)
	router.GET("/mine", auth, readLimiter, p.dh114Handler.ListMine)

	// 商户主表收藏快捷操作
	router.POST("/:id/fav", auth, p.dh114Handler.Fav)
	router.DELETE("/:id/fav", auth, p.dh114Handler.Unfav)
	router.GET("/:id/fav-status", auth, readLimiter, p.dh114Handler.FavStatus)
	router.GET("/favs", auth, readLimiter, p.dh114Handler.ListFavs)

	// 商户详情 CRUD
	router.POST("/businesses", auth, writeLimiter, p.businessHandler.Create)
	router.PUT("/businesses/:id", auth, writeLimiter, p.businessHandler.Update)
	router.DELETE("/businesses/:id", auth, writeLimiter, p.businessHandler.Delete)
	router.PUT("/:id/business/verification-status", auth, p.businessHandler.UpdateVerificationStatus)

	// 营业时间 CRUD
	router.PUT("/:id/business-hours", auth, writeLimiter, p.businessHandler.ReplaceBusinessHours)

	// 菜单 CRUD
	router.POST("/menus", auth, writeLimiter, p.businessHandler.CreateMenu)
	router.PUT("/menus/:id", auth, writeLimiter, p.businessHandler.UpdateMenu)
	router.DELETE("/menus/:id", auth, writeLimiter, p.businessHandler.DeleteMenu)
	router.PUT("/:id/menus", auth, writeLimiter, p.businessHandler.ReplaceMenus)
	router.POST("/menus/:id/order-count", auth, p.businessHandler.IncrMenuOrderCount)

	// 分类 CRUD
	router.POST("/categories", auth, writeLimiter, p.categoryHandler.Create)
	router.PUT("/categories/:id", auth, writeLimiter, p.categoryHandler.Update)
	router.DELETE("/categories/:id", auth, writeLimiter, p.categoryHandler.Delete)
	router.PUT("/categories/:id/status", auth, p.categoryHandler.UpdateStatus)

	// 优惠券 CRUD
	router.POST("/coupons", auth, writeLimiter, p.couponHandler.Create)
	router.PUT("/coupons/:id", auth, writeLimiter, p.couponHandler.Update)
	router.DELETE("/coupons/:id", auth, writeLimiter, p.couponHandler.Delete)
	router.POST("/coupons/:id/receive", auth, writeLimiter, p.couponHandler.Receive)
	router.POST("/coupons/:id/use", auth, writeLimiter, p.couponHandler.Use)

	// 团购 CRUD
	router.POST("/groupbuys", auth, writeLimiter, p.groupbuyHandler.Create)
	router.PUT("/groupbuys/:id", auth, writeLimiter, p.groupbuyHandler.Update)
	router.DELETE("/groupbuys/:id", auth, writeLimiter, p.groupbuyHandler.Delete)
	router.POST("/groupbuys/:id/sold", auth, p.groupbuyHandler.IncrSold)

	// 评价 CRUD + 回复 + 点赞
	router.POST("/reviews", auth, writeLimiter, p.reviewHandler.Create)
	router.PUT("/reviews/:id", auth, writeLimiter, p.reviewHandler.Update)
	router.DELETE("/reviews/:id", auth, writeLimiter, p.reviewHandler.Delete)
	router.GET("/reviews/mine", auth, readLimiter, p.reviewHandler.ListMine)
	router.POST("/reviews/:id/replies", auth, writeLimiter, p.reviewHandler.Reply)
	router.DELETE("/reviews/:id/replies/:reply_id", auth, p.reviewHandler.DeleteReply)
	router.POST("/reviews/:id/like", auth, p.reviewHandler.Like)

	// 收藏 CRUD（独立收藏表）
	router.POST("/favorites", auth, writeLimiter, p.favoriteHandler.Create)
	router.DELETE("/favorites/:id", auth, p.favoriteHandler.Delete)
	router.DELETE("/favorites/by-target", auth, p.favoriteHandler.DeleteByTarget)
	router.PUT("/favorites/:id", auth, p.favoriteHandler.Update)
	router.GET("/favorites/:id", auth, readLimiter, p.favoriteHandler.GetByID)
	router.GET("/favorites", auth, readLimiter, p.favoriteHandler.List)
	router.GET("/favorites/by-type", auth, readLimiter, p.favoriteHandler.ListByType)
	router.GET("/favorites/by-group", auth, readLimiter, p.favoriteHandler.ListByGroup)
	router.GET("/favorites/has-faved", auth, readLimiter, p.favoriteHandler.HasFaved)

	// 电话本 CRUD
	router.POST("/phone-calls", auth, writeLimiter, p.phoneCallHandler.Create)
	router.GET("/phone-calls/mine", auth, readLimiter, p.phoneCallHandler.ListByCaller)

	// 认证 CRUD
	router.POST("/verifications", auth, writeLimiter, p.verificationHandler.Create)
	router.PUT("/verifications/:id", auth, writeLimiter, p.verificationHandler.Update)
	router.DELETE("/verifications/:id", auth, writeLimiter, p.verificationHandler.Delete)
	router.GET("/verifications/:id", auth, readLimiter, p.verificationHandler.GetByID)
	router.GET("/verifications", auth, readLimiter, p.verificationHandler.List)
	router.GET("/verifications/mine", auth, readLimiter, p.verificationHandler.ListMine)
	router.GET("/verifications/latest", auth, readLimiter, p.verificationHandler.FindLatest)
	router.GET("/:id/verifications", auth, readLimiter, p.verificationHandler.ListByDh114)

	// 推荐交互（需登录）
	router.GET("/recommendations/mine", auth, readLimiter, p.recommendationHandler.ListByUser)
	router.POST("/recommendations/:id/clicked", auth, p.recommendationHandler.MarkClicked)
	router.POST("/recommendations/:id/contacted", auth, p.recommendationHandler.MarkContacted)
	router.POST("/recommendations/:id/dismissed", auth, p.recommendationHandler.MarkDismissed)

	// ==================== 管理后台路由（需 dh114:audit 权限） ====================

	admin := router.Group("/admin")

	// 商户主表管理
	admin.GET("/dh114s", auditPerm, p.dh114Handler.AdminList)
	admin.GET("/dh114s/:id", auditPerm, p.dh114Handler.AdminGetByID)
	admin.PUT("/dh114s/:id/audit", auditPerm, p.dh114Handler.Audit)
	admin.POST("/dh114s/batch-audit", auditPerm, p.dh114Handler.BatchAudit)
	admin.PUT("/dh114s/:id/status", auditPerm, p.dh114Handler.AdminUpdateStatus)
	admin.PUT("/dh114s/:id/promotion", auditPerm, p.dh114Handler.UpdatePromotion)

	// 优惠券管理
	admin.GET("/coupons", auditPerm, p.couponHandler.AdminList)
	admin.PUT("/coupons/:id/audit", auditPerm, p.couponHandler.Audit)
	admin.POST("/coupons/batch-audit", auditPerm, p.couponHandler.BatchAudit)
	admin.PUT("/coupons/:id/status", auditPerm, p.couponHandler.AdminUpdateStatus)

	// 团购管理
	admin.GET("/groupbuys", auditPerm, p.groupbuyHandler.AdminList)
	admin.PUT("/groupbuys/:id/audit", auditPerm, p.groupbuyHandler.Audit)
	admin.POST("/groupbuys/batch-audit", auditPerm, p.groupbuyHandler.BatchAudit)
	admin.PUT("/groupbuys/:id/status", auditPerm, p.groupbuyHandler.AdminUpdateStatus)

	// 评价管理
	admin.PUT("/reviews/:id/audit", auditPerm, p.reviewHandler.Audit)
	admin.POST("/reviews/batch-audit", auditPerm, p.reviewHandler.BatchAudit)

	// 电话本管理
	admin.GET("/phone-calls", auditPerm, p.phoneCallHandler.AdminList)

	// 认证管理
	admin.GET("/verifications", auditPerm, p.verificationHandler.AdminList)
	admin.PUT("/verifications/:id/audit", auditPerm, p.verificationHandler.Audit)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.List)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.UpdateStatus)

	// 推荐管理
	admin.GET("/recommendations", auditPerm, p.recommendationHandler.AdminList)
	admin.POST("/recommendations", auditPerm, p.recommendationHandler.AdminCreate)
	admin.DELETE("/recommendations/:id", auditPerm, p.recommendationHandler.AdminDelete)

	// 数据统计管理
	admin.GET("/statistics/overview", auditPerm, p.statisticHandler.Overview)
	admin.POST("/statistics", auditPerm, p.statisticHandler.AdminUpsert)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
