// Package love 同城相亲交友模块插件
// 提供资料/匹配/喜欢/拉黑/访客/印象/故事/认证/会员/审核规则/礼物/推荐/举报/通知/隐私/会话等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标 Soul / 陌陌 / 探探 / 百合网
package love

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/love/handler"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/modules/love/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 相亲交友模块插件
type Plugin struct {
	name    string
	version string

	// 18 个 Handler
	loveHandler            *handler.LoveHandler
	profileHandler        *handler.ProfileHandler
	matchHandler          *handler.MatchHandler
	likeHandler           *handler.LikeHandler
	blockHandler          *handler.BlockHandler
	visitHandler          *handler.VisitHandler
	impressionHandler     *handler.ImpressionHandler
	storyHandler          *handler.StoryHandler
	verificationHandler   *handler.VerificationHandler
	memberLevelHandler    *handler.MemberLevelHandler
	membershipHandler     *handler.MembershipHandler
	auditRuleHandler      *handler.AuditRuleHandler
	giftHandler           *handler.GiftHandler
	recommendationHandler *handler.RecommendationHandler
	reportHandler         *handler.ReportHandler
	notificationHandler   *handler.NotificationHandler
	privacyHandler        *handler.PrivacyHandler
	chatSessionHandler    *handler.ChatSessionHandler
}

// NewPlugin 创建相亲交友模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "love", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "love",
		DisplayName: "同城相亲交友",
		Category:    "business",
		Description: "同城相亲交友完整功能：资料/匹配/喜欢/拉黑/访客/印象/故事/认证/会员等级/会员订阅/审核规则/礼物/推荐/举报/通知/隐私/会话",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：22 张表（含 8 张复用表）由 backend/migrations 创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 loves 主表（子表由 migrations 创建）
	// 注：未调用 AutoMigrate 会导致 loves 主表缺失，前端列表报 "relation loves does not exist"
	if err := db.AutoMigrate(&model.Love{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（19 个） =====
	loveRepo := repository.NewLoveRepository(db)
	profileRepo := repository.NewLoveProfileRepository(db)
	matchRepo := repository.NewLoveMatchRepository(db)
	likeRepo := repository.NewLoveLikeRepository(db)
	blockRepo := repository.NewLoveBlockRepository(db)
	visitRepo := repository.NewLoveVisitRepository(db)
	impressionRepo := repository.NewLoveImpressionRepository(db)
	storyRepo := repository.NewLoveStoryRepository(db)
	verificationRepo := repository.NewLoveVerificationRepository(db)
	memberLevelRepo := repository.NewLoveMemberLevelRepository(db)
	membershipRepo := repository.NewLoveMembershipRepository(db)
	auditRuleRepo := repository.NewLoveAuditRuleRepository(db)
	giftRepo := repository.NewLoveGiftRepository(db)
	giftRecordRepo := repository.NewLoveGiftRecordRepository(db)
	recommendationRepo := repository.NewLoveRecommendationRepository(db)
	reportRepo := repository.NewLoveReportRepository(db)
	notificationRepo := repository.NewLoveNotificationRepository(db)
	privacyRepo := repository.NewLovePrivacySettingRepository(db)
	chatSessionRepo := repository.NewLoveChatSessionRepository(db)

	// ===== 依赖注入：Service 层（18 个） =====
	loveSvc := service.NewLoveService(loveRepo)
	profileSvc := service.NewLoveProfileService(profileRepo)
	matchSvc := service.NewLoveMatchService(matchRepo)
	likeSvc := service.NewLoveLikeService(likeRepo)
	blockSvc := service.NewLoveBlockService(blockRepo)
	visitSvc := service.NewLoveVisitService(visitRepo)
	impressionSvc := service.NewLoveImpressionService(impressionRepo)
	storySvc := service.NewLoveStoryService(storyRepo)
	verificationSvc := service.NewLoveVerificationService(verificationRepo)
	memberLevelSvc := service.NewLoveMemberLevelService(memberLevelRepo)
	membershipSvc := service.NewLoveMembershipService(membershipRepo, memberLevelRepo)
	auditRuleSvc := service.NewLoveAuditRuleService(auditRuleRepo)
	giftSvc := service.NewLoveGiftService(giftRepo, giftRecordRepo)
	recommendationSvc := service.NewLoveRecommendationService(recommendationRepo)
	reportSvc := service.NewLoveReportService(reportRepo)
	notificationSvc := service.NewLoveNotificationService(notificationRepo)
	privacySvc := service.NewLovePrivacyService(privacyRepo)
	chatSessionSvc := service.NewLoveChatSessionService(chatSessionRepo)

	// ===== 依赖注入：Handler 层（18 个） =====
	p.loveHandler = handler.NewLoveHandler(loveSvc)
	p.profileHandler = handler.NewProfileHandler(profileSvc)
	p.matchHandler = handler.NewMatchHandler(matchSvc)
	p.likeHandler = handler.NewLikeHandler(likeSvc)
	p.blockHandler = handler.NewBlockHandler(blockSvc)
	p.visitHandler = handler.NewVisitHandler(visitSvc)
	p.impressionHandler = handler.NewImpressionHandler(impressionSvc)
	p.storyHandler = handler.NewStoryHandler(storySvc)
	p.verificationHandler = handler.NewVerificationHandler(verificationSvc)
	p.memberLevelHandler = handler.NewMemberLevelHandler(memberLevelSvc)
	p.membershipHandler = handler.NewMembershipHandler(membershipSvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)
	p.giftHandler = handler.NewGiftHandler(giftSvc)
	p.recommendationHandler = handler.NewRecommendationHandler(recommendationSvc)
	p.reportHandler = handler.NewReportHandler(reportSvc)
	p.notificationHandler = handler.NewNotificationHandler(notificationSvc)
	p.privacyHandler = handler.NewPrivacyHandler(privacySvc)
	p.chatSessionHandler = handler.NewChatSessionHandler(chatSessionSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/love
//
// 路由分组（共 130+ API）：
//   - 公开路由（C 端浏览，无需登录）：资料/故事/印象/审核规则检查/礼物列表
//   - 需登录路由（C 端发布/交易/互动）：资料 CRUD/喜欢/拉黑/访客/印象/故事/认证/会员/礼物/推荐/举报/通知/隐私/匹配/会话
//   - 管理后台路由（需 love:audit 权限）：审核/批量/统计/审核规则/举报处理/会员等级/会员订阅/礼物管理
//
// 注意：固定路径（/search /nearby /me /matched /unread 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "love_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "love_write"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "love_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "love_nearby"))
	likeLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "love_like"))
	visitLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "love_visit"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("love:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 主表 Love（loves）
	router.GET("", readLimiter, p.loveHandler.List)
	router.GET("/search", searchLimiter, p.loveHandler.Search)
	router.GET("/nearby", nearbyLimiter, p.loveHandler.Nearby)
	router.GET("/advanced-search", searchLimiter, p.loveHandler.AdvancedSearch)
	router.GET("/:id", readLimiter, p.loveHandler.GetByID)

	// 资料（profiles）
	router.GET("/profiles", readLimiter, p.profileHandler.List)
	router.GET("/profiles/:id", readLimiter, p.profileHandler.GetByID)
	router.GET("/profiles/by-love/:id", readLimiter, p.profileHandler.GetByLoveID)
	router.GET("/profiles/by-user/:id", readLimiter, p.profileHandler.GetByUserID)

	// 故事（stories）
	router.GET("/stories", readLimiter, p.storyHandler.List)
	router.GET("/stories/featured", readLimiter, p.storyHandler.ListFeatured)
	router.GET("/stories/topic/:topic", readLimiter, p.storyHandler.ListByTopic)
	router.GET("/stories/by-love/:id", readLimiter, p.storyHandler.ListByLoveID)
	router.GET("/stories/by-user/:id", readLimiter, p.storyHandler.ListByUser)
	router.GET("/stories/:id", readLimiter, p.storyHandler.GetByID)

	// 印象标签（impressions）
	router.GET("/impressions", readLimiter, p.impressionHandler.List)
	router.GET("/impressions/by-love/:id", readLimiter, p.impressionHandler.ListByLoveID)
	router.GET("/impressions/:id", readLimiter, p.impressionHandler.GetByID)
	router.GET("/impressions/stats", readLimiter, p.impressionHandler.Stats)

	// 审核规则公开检查接口
	router.POST("/audit-rules/check", readLimiter, p.auditRuleHandler.Check)

	// 礼物列表（公开浏览）
	router.GET("/gifts", readLimiter, p.giftHandler.ListAvailable)
	router.GET("/gifts/available", readLimiter, p.giftHandler.ListAvailable)

	// 会员等级（公开浏览，便于 C 端展示会员套餐）
	router.GET("/member-levels", readLimiter, p.memberLevelHandler.List)

	// 匹配公开查看（对方点开匹配详情，需登录态由 C 端组承担）
	// ==================== 需登录路由（C 端发布/交易/互动） ====================

	// 主表 Love CRUD
	router.POST("", auth, writeLimiter, p.loveHandler.Create)
	router.PUT("/:id", auth, writeLimiter, p.loveHandler.Update)
	router.GET("/me", auth, readLimiter, p.loveHandler.GetByUserID)
	router.PUT("/:id/location", auth, writeLimiter, p.loveHandler.UpdateLocation)
	router.PUT("/:id/voice-intro", auth, writeLimiter, p.loveHandler.UpdateVoiceIntro)
	router.GET("/match-score", auth, readLimiter, p.loveHandler.MatchScore)

	// 资料 CRUD
	router.POST("/profiles", auth, writeLimiter, p.profileHandler.Create)
	router.PUT("/profiles/:id", auth, writeLimiter, p.profileHandler.Update)
	router.PUT("/profiles/by-love/:id", auth, writeLimiter, p.profileHandler.UpdateByLoveID)
	router.PUT("/profiles/:id/step", auth, writeLimiter, p.profileHandler.UpdateStep)

	// 喜欢/不喜欢/心动信号（likes）
	router.POST("/likes", auth, likeLimiter, p.likeHandler.Act)
	router.DELETE("/likes/:id", auth, writeLimiter, p.likeHandler.Undo)
	router.GET("/likes/:id", auth, readLimiter, p.likeHandler.GetByID)
	router.GET("/likes", auth, readLimiter, p.likeHandler.List)
	router.GET("/likes/matched", auth, readLimiter, p.likeHandler.ListMatched)
	router.GET("/likes/has-liked", auth, readLimiter, p.likeHandler.HasLiked)
	router.GET("/likes/today-stats", auth, readLimiter, p.likeHandler.TodayStats)

	// 拉黑（blocks）
	router.POST("/blocks", auth, writeLimiter, p.blockHandler.Block)
	router.DELETE("/blocks/:id", auth, writeLimiter, p.blockHandler.Unblock)
	router.GET("/blocks/:id", auth, readLimiter, p.blockHandler.GetByID)
	router.GET("/blocks", auth, readLimiter, p.blockHandler.List)
	router.GET("/blocks/has-blocked", auth, readLimiter, p.blockHandler.HasBlocked)

	// 访客（visits）
	router.POST("/visits", auth, visitLimiter, p.visitHandler.Visit)
	router.GET("/visits/:id", auth, readLimiter, p.visitHandler.GetByID)
	router.GET("/visits", auth, readLimiter, p.visitHandler.List)
	router.GET("/visits/by-visitor", auth, readLimiter, p.visitHandler.ListByVisitor)
	router.GET("/visits/unread", auth, readLimiter, p.visitHandler.ListUnread)
	router.POST("/visits/:id/read", auth, writeLimiter, p.visitHandler.MarkRead)
	router.POST("/visits/read-all", auth, writeLimiter, p.visitHandler.MarkAllRead)
	router.GET("/visits/stats", auth, readLimiter, p.visitHandler.Stats)

	// 印象标签 CRUD
	router.POST("/impressions", auth, writeLimiter, p.impressionHandler.Create)
	router.DELETE("/impressions/:id", auth, writeLimiter, p.impressionHandler.Delete)

	// 故事 CRUD + 互动
	router.POST("/stories", auth, writeLimiter, p.storyHandler.Create)
	router.PUT("/stories/:id", auth, writeLimiter, p.storyHandler.Update)
	router.DELETE("/stories/:id", auth, writeLimiter, p.storyHandler.Delete)
	router.POST("/stories/:id/view", auth, readLimiter, p.storyHandler.IncrView)
	router.POST("/stories/:id/like", auth, readLimiter, p.storyHandler.IncrLike)
	router.POST("/stories/:id/unlike", auth, readLimiter, p.storyHandler.DecrLike)
	router.POST("/stories/:id/comment", auth, readLimiter, p.storyHandler.IncrComment)
	router.POST("/stories/:id/share", auth, readLimiter, p.storyHandler.IncrShare)

	// 认证（verifications）
	router.POST("/verifications", auth, writeLimiter, p.verificationHandler.Submit)
	router.GET("/verifications/:id", auth, readLimiter, p.verificationHandler.GetByID)
	router.GET("/verifications/by-love/:id", auth, readLimiter, p.verificationHandler.GetByLoveID)
	router.GET("/verifications/by-user/:id", auth, readLimiter, p.verificationHandler.GetByUserID)
	router.GET("/verifications", auth, readLimiter, p.verificationHandler.List)

	// 会员订阅（memberships）
	router.POST("/memberships", auth, writeLimiter, p.membershipHandler.Open)
	router.GET("/memberships/:id", auth, readLimiter, p.membershipHandler.GetByID)
	router.GET("/memberships/me", auth, readLimiter, p.membershipHandler.GetMyActive)
	router.GET("/memberships", auth, readLimiter, p.membershipHandler.List)
	router.GET("/memberships/by-user/:id", auth, readLimiter, p.membershipHandler.ListByUser)
	router.POST("/memberships/:id/cancel", auth, writeLimiter, p.membershipHandler.Cancel)
	router.POST("/memberships/:id/refund", auth, writeLimiter, p.membershipHandler.Refund)
	router.POST("/memberships/:id/mark-paid", auth, writeLimiter, p.membershipHandler.MarkPaid)
	router.PUT("/memberships/:id/auto-renew", auth, writeLimiter, p.membershipHandler.UpdateAutoRenew)

	// 匹配（matches）
	router.GET("/matches/:id", auth, readLimiter, p.matchHandler.GetByID)
	router.GET("/matches", auth, readLimiter, p.matchHandler.List)
	router.POST("/matches/:id/dissolve", auth, writeLimiter, p.matchHandler.Dissolve)
	router.GET("/matches/count", auth, readLimiter, p.matchHandler.CountByUser)
	router.GET("/matches/today", auth, readLimiter, p.matchHandler.CountToday)

	// 会话（chat-sessions）
	router.GET("/chat-sessions/:id", auth, readLimiter, p.chatSessionHandler.GetByID)
	router.GET("/chat-sessions/by-match/:id", auth, readLimiter, p.chatSessionHandler.GetByMatchID)
	router.GET("/chat-sessions", auth, readLimiter, p.chatSessionHandler.List)
	router.POST("/chat-sessions/:id/action", auth, writeLimiter, p.chatSessionHandler.Action)
	router.POST("/chat-sessions/:id/read", auth, writeLimiter, p.chatSessionHandler.MarkRead)
	router.GET("/chat-sessions/active-count", auth, readLimiter, p.chatSessionHandler.CountActive)

	// 礼物（gifts）C 端送出
	router.POST("/gifts/send", auth, writeLimiter, p.giftHandler.Send)

	// 推荐（recommendations）
	router.POST("/recommendations/generate", auth, writeLimiter, p.recommendationHandler.Generate)
	router.GET("/recommendations/:id", auth, readLimiter, p.recommendationHandler.GetByID)
	router.GET("/recommendations", auth, readLimiter, p.recommendationHandler.List)
	router.GET("/recommendations/by-type/:type", auth, readLimiter, p.recommendationHandler.ListByType)
	router.POST("/recommendations/:id/action", auth, writeLimiter, p.recommendationHandler.Action)
	router.GET("/recommendations/stats", auth, readLimiter, p.recommendationHandler.Stats)

	// 举报（reports）
	router.POST("/reports", auth, writeLimiter, p.reportHandler.Create)
	router.GET("/reports/:id", auth, readLimiter, p.reportHandler.GetByID)
	router.GET("/reports/mine", auth, readLimiter, p.reportHandler.ListByReporter)
	router.POST("/reports/:id/appeal", auth, writeLimiter, p.reportHandler.Appeal)

	// 通知（notifications）
	router.GET("/notifications", auth, readLimiter, p.notificationHandler.List)
	router.GET("/notifications/unread", auth, readLimiter, p.notificationHandler.ListUnread)
	router.GET("/notifications/:id", auth, readLimiter, p.notificationHandler.GetByID)
	router.POST("/notifications/:id/read", auth, writeLimiter, p.notificationHandler.MarkRead)
	router.POST("/notifications/read-all", auth, writeLimiter, p.notificationHandler.MarkAllRead)
	router.POST("/notifications/batch-read", auth, writeLimiter, p.notificationHandler.BatchMarkRead)
	router.DELETE("/notifications/:id", auth, writeLimiter, p.notificationHandler.Delete)
	router.GET("/notifications/unread-count", auth, readLimiter, p.notificationHandler.CountUnread)
	router.GET("/notifications/stats", auth, readLimiter, p.notificationHandler.Stats)

	// 隐私（privacy）
	router.GET("/privacy", auth, readLimiter, p.privacyHandler.Get)
	router.GET("/privacy/by-love/:id", auth, readLimiter, p.privacyHandler.GetByLoveID)
	router.PUT("/privacy", auth, writeLimiter, p.privacyHandler.Update)
	router.POST("/privacy/reset", auth, writeLimiter, p.privacyHandler.Reset)
	router.GET("/privacy/is-visible", auth, readLimiter, p.privacyHandler.IsVisible)

	// ==================== 管理后台路由（需 love:audit 权限） ====================

	admin := router.Group("/admin")

	// 主表 Love 管理
	admin.GET("/loves", auditPerm, p.loveHandler.AdminList)
	admin.GET("/loves/:id", auditPerm, p.loveHandler.AdminGetByID)
	admin.PUT("/loves/:id/audit", auditPerm, p.loveHandler.Audit)
	admin.PUT("/loves/:id/status", auditPerm, p.loveHandler.AdminUpdateStatus)
	admin.PUT("/loves/:id/featured", auditPerm, p.loveHandler.SetFeatured)
	admin.PUT("/loves/:id/picked", auditPerm, p.loveHandler.SetPicked)
	admin.PUT("/loves/batch-audit", auditPerm, p.loveHandler.BatchAudit)
	admin.PUT("/loves/batch-status", auditPerm, p.loveHandler.BatchUpdateStatus)

	// 故事管理
	admin.GET("/stories", auditPerm, p.storyHandler.AdminList)
	admin.PUT("/stories/:id/audit", auditPerm, p.storyHandler.Audit)
	admin.PUT("/stories/:id/status", auditPerm, p.storyHandler.UpdateStatus)
	admin.PUT("/stories/:id/featured", auditPerm, p.storyHandler.SetFeatured)
	admin.PUT("/stories/batch-audit", auditPerm, p.storyHandler.BatchAudit)

	// 认证管理
	admin.GET("/verifications", auditPerm, p.verificationHandler.List)
	admin.PUT("/verifications/:id/audit", auditPerm, p.verificationHandler.Audit)
	admin.GET("/verifications/pending-count", auditPerm, p.verificationHandler.CountPending)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.List)
	admin.GET("/audit-rules/all", auditPerm, p.auditRuleHandler.ListAll)
	admin.GET("/audit-rules/:id", auditPerm, p.auditRuleHandler.GetByID)
	admin.GET("/audit-rules/by-key/:key", auditPerm, p.auditRuleHandler.GetByRuleKey)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.UpdateStatus)
	admin.PUT("/audit-rules/batch-status", auditPerm, p.auditRuleHandler.BatchUpdateStatus)

	// 礼物管理
	admin.GET("/gifts", auditPerm, p.giftHandler.List)
	admin.GET("/gifts/:id", auditPerm, p.giftHandler.GetByID)
	admin.POST("/gifts", auditPerm, p.giftHandler.Create)
	admin.PUT("/gifts/:id", auditPerm, p.giftHandler.Update)
	admin.DELETE("/gifts/:id", auditPerm, p.giftHandler.Delete)
	admin.PUT("/gifts/:id/status", auditPerm, p.giftHandler.UpdateStatus)
	admin.PUT("/gifts/batch-status", auditPerm, p.giftHandler.BatchUpdateStatus)

	// 会员等级管理
	admin.GET("/member-levels", auditPerm, p.memberLevelHandler.List)
	admin.GET("/member-levels/all", auditPerm, p.memberLevelHandler.ListAll)
	admin.GET("/member-levels/:id", auditPerm, p.memberLevelHandler.GetByID)
	admin.GET("/member-levels/by-code/:code", auditPerm, p.memberLevelHandler.GetByLevelCode)
	admin.POST("/member-levels", auditPerm, p.memberLevelHandler.Create)
	admin.PUT("/member-levels/:id", auditPerm, p.memberLevelHandler.Update)
	admin.DELETE("/member-levels/:id", auditPerm, p.memberLevelHandler.Delete)

	// 会员订阅管理
	admin.GET("/memberships", auditPerm, p.membershipHandler.List)
	admin.GET("/memberships/:id", auditPerm, p.membershipHandler.GetByID)
	admin.GET("/memberships/by-user/:id", auditPerm, p.membershipHandler.ListByUser)
	admin.POST("/memberships/:id/refund", auditPerm, p.membershipHandler.Refund)
	admin.PUT("/memberships/:id/auto-renew", auditPerm, p.membershipHandler.UpdateAutoRenew)

	// 举报管理
	admin.GET("/reports", auditPerm, p.reportHandler.List)
	admin.GET("/reports/pending", auditPerm, p.reportHandler.ListPending)
	admin.GET("/reports/by-target", auditPerm, p.reportHandler.ListByTarget)
	admin.GET("/reports/:id", auditPerm, p.reportHandler.GetByID)
	admin.PUT("/reports/:id/handle", auditPerm, p.reportHandler.Handle)
	admin.PUT("/reports/:id/appeal", auditPerm, p.reportHandler.HandleAppeal)
	admin.PUT("/reports/:id/risk-score", auditPerm, p.reportHandler.UpdateRiskScore)
	admin.DELETE("/reports/:id", auditPerm, p.reportHandler.Delete)
	admin.GET("/reports/stats", auditPerm, p.reportHandler.Stats)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
