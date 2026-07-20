// Package pinche 同城拼车出行模块插件
// 提供拼车发布/预订/行程/车主认证/车辆/支付/评价/保险/紧急联系/路线/审核规则/统计等业务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：内容审核必须做（MVP 简化为发布即通过，M 端可手动审核/下架）
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package pinche

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/pinche/handler"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/modules/pinche/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 拼车出行模块插件
type Plugin struct {
	name    string
	version string

	// 12 个 Handler
	pincheHandler      *handler.PincheHandler
	routeHandler       *handler.RouteHandler
	tripHandler        *handler.TripHandler
	bookingHandler     *handler.BookingHandler
	driverHandler      *handler.DriverHandler
	vehicleHandler     *handler.VehicleHandler
	paymentHandler     *handler.PaymentHandler
	ratingHandler      *handler.RatingHandler
	insuranceHandler   *handler.InsuranceHandler
	emergencyHandler   *handler.EmergencyHandler
	auditRuleHandler   *handler.AuditRuleHandler
	statisticsHandler  *handler.StatisticsHandler
}

// NewPlugin 创建拼车出行模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "pinche", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "pinche",
		DisplayName: "同城拼车出行",
		Category:    "business",
		Description: "同城拼车出行完整功能：拼车发布/预订/行程/车主认证/车辆/支付/评价/保险/紧急联系/路线/审核规则/统计",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：19 张表由 backend/migrations 下 SQL 迁移脚本创建，
// 约束名遵循 PostgreSQL 默认命名规则，不使用 GORM AutoMigrate 以避免约束名不一致错误。
// 此处仅 AutoMigrate 主表 Pinche（与 car 模块保持一致），子表由 SQL 脚本创建。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移 pinche 主表（子表由 backend/migrations 下 SQL 脚本创建）
	if err := db.AutoMigrate(&model.Pinche{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（18 个，含 1 个 driver_location 和 1 个 route_favorite） =====
	pincheRepo := repository.NewPincheRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeFavRepo := repository.NewRouteFavoriteRepository(db)
	tripRepo := repository.NewTripRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	driverLocationRepo := repository.NewDriverLocationRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	ratingRepo := repository.NewRatingRepository(db)
	insuranceRepo := repository.NewInsuranceRepository(db)
	emergencyRepo := repository.NewEmergencyRepository(db)
	auditRuleRepo := repository.NewAuditRuleRepository(db)
	statisticRepo := repository.NewStatisticRepository(db)
	cancelRepo := repository.NewCancelRepository(db)
	refundRepo := repository.NewRefundRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// ===== 依赖注入：Service 层（12 个） =====
	pincheSvc := service.NewPincheService(pincheRepo)
	routeSvc := service.NewRouteService(routeRepo, routeFavRepo)
	tripSvc := service.NewTripService(tripRepo)
	bookingSvc := service.NewBookingService(bookingRepo, pincheRepo)
	driverSvc := service.NewDriverService(driverRepo)
	vehicleSvc := service.NewVehicleService(vehicleRepo)
	paymentSvc := service.NewPaymentService(paymentRepo)
	ratingSvc := service.NewRatingService(ratingRepo)
	insuranceSvc := service.NewInsuranceService(insuranceRepo)
	emergencySvc := service.NewEmergencyService(emergencyRepo)
	auditRuleSvc := service.NewAuditRuleService(auditRuleRepo)
	statisticSvc := service.NewStatisticService(statisticRepo)

	// ===== 依赖注入：Handler 层（12 个） =====
	p.pincheHandler = handler.NewPincheHandler(pincheSvc)
	p.routeHandler = handler.NewRouteHandler(routeSvc)
	p.tripHandler = handler.NewTripHandler(tripSvc)
	p.bookingHandler = handler.NewBookingHandler(bookingSvc)
	p.driverHandler = handler.NewDriverHandler(driverSvc)
	p.vehicleHandler = handler.NewVehicleHandler(vehicleSvc)
	p.paymentHandler = handler.NewPaymentHandler(paymentSvc)
	p.ratingHandler = handler.NewRatingHandler(ratingSvc)
	p.insuranceHandler = handler.NewInsuranceHandler(insuranceSvc)
	p.emergencyHandler = handler.NewEmergencyHandler(emergencySvc)
	p.auditRuleHandler = handler.NewAuditRuleHandler(auditRuleSvc)
	p.statisticsHandler = handler.NewStatisticsHandler(statisticSvc)

	// 避免未使用变量告警（driver_location/cancel/refund/complaint/message 暂未在 service 层使用，保留以备后续扩展）
	_ = driverLocationRepo
	_ = cancelRepo
	_ = refundRepo
	_ = complaintRepo
	_ = messageRepo

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/pinche
//
// 路由分组（共 80+ API）：
//   - 公开路由（C 端浏览，无需登录）：拼车列表/详情/搜索/附近/路线/审核规则只读/统计只读
//   - 需登录路由（C 端发布/交易/收藏）：拼车 CRUD/预订/行程/车主认证/车辆/支付/评价/保险/紧急联系
//   - 管理后台路由（需 pinche:audit 权限）：审核/批量/统计/审核规则/报警处理
//
// 注意：固定路径（/search /nearby /mine /match /hot 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "pinche_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "pinche_write"))
	searchLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "pinche_search"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "pinche_nearby"))
	auditPerm := coreRouter.WrapGin(middleware.RequirePermission("pinche:audit"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// 拼车主表（pinches）
	router.GET("", readLimiter, p.pincheHandler.List)
	router.GET("/search", searchLimiter, p.pincheHandler.Search)
	router.GET("/nearby", nearbyLimiter, p.pincheHandler.Nearby)
	router.GET("/match", searchLimiter, p.pincheHandler.Match)
	router.GET("/:id", readLimiter, p.pincheHandler.GetByID)
	router.POST("/:id/contact", p.pincheHandler.IncrContact)
	router.POST("/:id/share", p.pincheHandler.IncrShare)
	router.POST("/views", p.pincheHandler.RecordView)

	// 路线（routes）
	router.GET("/routes", readLimiter, p.routeHandler.List)
	router.GET("/routes/common", readLimiter, p.routeHandler.ListCommon)
	router.GET("/routes/:id", readLimiter, p.routeHandler.GetByID)

	// 行程（trips） - 公开查询
	router.GET("/trips", readLimiter, p.tripHandler.AdminList)
	router.GET("/trips/:id", readLimiter, p.tripHandler.GetByID)
	router.GET("/trips/no/:trip_no", readLimiter, p.tripHandler.GetByTripNo)
	router.GET("/trips/share/:token", readLimiter, p.tripHandler.GetByShareToken)
	router.GET("/pinches/:id/trips", readLimiter, p.tripHandler.ListByPinche)

	// 预订（bookings） - 公开查询
	router.GET("/bookings", readLimiter, p.bookingHandler.List)
	router.GET("/bookings/:id", readLimiter, p.bookingHandler.GetByID)
	router.GET("/pinches/:id/bookings", readLimiter, p.bookingHandler.ListByPinche)

	// 车主认证（drivers）- 公开查询
	router.GET("/drivers", readLimiter, p.driverHandler.List)
	router.GET("/drivers/:id", readLimiter, p.driverHandler.GetByID)

	// 车辆（vehicles）- 公开查询
	router.GET("/vehicles", readLimiter, p.vehicleHandler.AdminList)
	router.GET("/vehicles/:id", readLimiter, p.vehicleHandler.GetByID)
	router.GET("/vehicles/driver/:id", readLimiter, p.vehicleHandler.ListByDriver)

	// 支付（payments）- 公开查询
	router.GET("/payments/:id", readLimiter, p.paymentHandler.GetByID)
	router.GET("/payments/no/:payment_no", readLimiter, p.paymentHandler.GetByPaymentNo)
	router.GET("/payments/callback", p.paymentHandler.Callback) // 支付回调（POST 也支持，下方注册 POST）

	// 评价（ratings）- 公开查询
	router.GET("/ratings", readLimiter, p.ratingHandler.ListByPinche) // 默认按 pinche_id 查询
	router.GET("/ratings/:id", readLimiter, p.ratingHandler.GetByID)
	router.GET("/ratings/stats", readLimiter, p.ratingHandler.Stats)
	router.GET("/pinches/:id/ratings", readLimiter, p.ratingHandler.ListByPinche)

	// 保险（insurances）- 公开查询
	router.GET("/insurances/:id", readLimiter, p.insuranceHandler.GetByID)
	router.GET("/insurances/no/:policy_no", readLimiter, p.insuranceHandler.GetByPolicyNo)
	router.GET("/pinches/:id/insurances", readLimiter, p.insuranceHandler.ListByPinche)
	router.GET("/bookings/:id/insurances", readLimiter, p.insuranceHandler.ListByBooking)

	// 紧急联系/报警 - 公开查询报警详情
	router.GET("/emergencies/:id", readLimiter, p.emergencyHandler.GetAlert)
	router.GET("/pinches/:id/emergencies", readLimiter, p.emergencyHandler.ListAlertsByPinche)
	router.GET("/trips/:id/emergencies", readLimiter, p.emergencyHandler.ListAlertsByTrip)

	// 审核规则（audit-rules）- 公开只读
	router.GET("/audit-rules", readLimiter, p.auditRuleHandler.List)
	router.GET("/audit-rules/enabled", readLimiter, p.auditRuleHandler.ListEnabled)
	router.GET("/audit-rules/type/:rule_type", readLimiter, p.auditRuleHandler.ListByType)
	router.GET("/audit-rules/:id", readLimiter, p.auditRuleHandler.GetByID)

	// 数据统计（公开部分：列表/详情/总览/区间查询）
	router.GET("/statistics", readLimiter, p.statisticsHandler.List)
	router.GET("/statistics/overview", readLimiter, p.statisticsHandler.Overview)
	router.GET("/statistics/date-range", readLimiter, p.statisticsHandler.ListByDateRange)
	router.GET("/statistics/:id", readLimiter, p.statisticsHandler.GetByID)

	// ==================== 需登录路由（C 端发布/收藏/交易） ====================

	// 拼车主表 CRUD
	router.POST("", auth, writeLimiter, p.pincheHandler.Create)
	router.PUT("/:id", auth, p.pincheHandler.Update)
	router.DELETE("/:id", auth, p.pincheHandler.Delete)
	router.GET("/mine", auth, readLimiter, p.pincheHandler.ListMine)

	// 路线 CRUD + 收藏
	router.POST("/routes", auth, writeLimiter, p.routeHandler.Create)
	router.PUT("/routes/:id", auth, p.routeHandler.Update)
	router.DELETE("/routes/:id", auth, p.routeHandler.Delete)
	router.GET("/routes/mine", auth, readLimiter, p.routeHandler.ListMine)
	router.POST("/routes/:id/fav", auth, p.routeHandler.Fav)
	router.GET("/routes/favorites", auth, readLimiter, p.routeHandler.ListFavs)
	router.POST("/routes/:id/use", auth, p.routeHandler.IncrUseCount)

	// 行程 CRUD
	router.POST("/trips", auth, writeLimiter, p.tripHandler.Start)
	router.PUT("/trips/:id/complete", auth, writeLimiter, p.tripHandler.Complete)
	router.POST("/trips/:id/confirm", auth, p.tripHandler.Confirm)
	router.GET("/trips/mine", auth, readLimiter, p.tripHandler.ListByUser)
	router.GET("/trips/driver/:id", auth, readLimiter, p.tripHandler.ListByDriver)
	router.GET("/trips/passenger/:id", auth, readLimiter, p.tripHandler.ListByPassenger)
	router.POST("/trips/:id/share", auth, writeLimiter, p.tripHandler.Share)

	// 预订 CRUD + 上车确认
	router.POST("/bookings", auth, writeLimiter, p.bookingHandler.Create)
	router.PUT("/bookings/:id", auth, p.bookingHandler.Update)
	router.POST("/bookings/:id/cancel", auth, writeLimiter, p.bookingHandler.Cancel)
	router.POST("/bookings/:id/boarding", auth, writeLimiter, p.bookingHandler.ConfirmBoarding)
	router.GET("/bookings/mine", auth, readLimiter, p.bookingHandler.ListByPassenger)
	router.GET("/bookings/driver", auth, readLimiter, p.bookingHandler.ListByDriver)

	// 车主认证 CRUD
	router.POST("/drivers", auth, writeLimiter, p.driverHandler.Create)
	router.PUT("/drivers/:id", auth, p.driverHandler.Update)
	router.GET("/drivers/me", auth, readLimiter, p.driverHandler.GetByUserID)

	// 车辆 CRUD
	router.POST("/vehicles", auth, writeLimiter, p.vehicleHandler.Create)
	router.PUT("/vehicles/:id", auth, p.vehicleHandler.Update)
	router.DELETE("/vehicles/:id", auth, p.vehicleHandler.Delete)
	router.GET("/vehicles/mine", auth, readLimiter, p.vehicleHandler.ListByUser)
	router.POST("/vehicles/:id/default", auth, p.vehicleHandler.SetDefault)

	// 支付 CRUD + ETC
	router.POST("/payments", auth, writeLimiter, p.paymentHandler.Create)
	router.POST("/payments/callback", p.paymentHandler.Callback) // 第三方支付回调（POST，无需登录）
	router.GET("/payments/payer", auth, readLimiter, p.paymentHandler.ListByPayer)
	router.GET("/payments/payee", auth, readLimiter, p.paymentHandler.ListByPayee)
	router.GET("/bookings/:id/payments", auth, readLimiter, p.paymentHandler.ListByBooking)
	router.POST("/payments/etc", auth, writeLimiter, p.paymentHandler.ETCSettlement)

	// 评价 CRUD
	router.POST("/ratings", auth, writeLimiter, p.ratingHandler.Create)
	router.PUT("/ratings/:id", auth, p.ratingHandler.Update)
	router.DELETE("/ratings/:id", auth, p.ratingHandler.Delete)
	router.GET("/ratings/rater", auth, readLimiter, p.ratingHandler.ListByRater)
	router.GET("/ratings/ratee", auth, readLimiter, p.ratingHandler.ListByRatee)
	router.POST("/ratings/:id/reply", auth, writeLimiter, p.ratingHandler.Reply)
	router.POST("/ratings/:id/like", auth, p.ratingHandler.Like)

	// 保险 CRUD
	router.POST("/insurances", auth, writeLimiter, p.insuranceHandler.Create)
	router.POST("/insurances/:id/claim", auth, writeLimiter, p.insuranceHandler.Claim)
	router.POST("/insurances/quote", auth, writeLimiter, p.insuranceHandler.Quote)

	// 紧急联系人 CRUD
	router.POST("/emergency-contacts", auth, writeLimiter, p.emergencyHandler.CreateContact)
	router.PUT("/emergency-contacts/:id", auth, p.emergencyHandler.UpdateContact)
	router.DELETE("/emergency-contacts/:id", auth, p.emergencyHandler.DeleteContact)
	router.GET("/emergency-contacts", auth, readLimiter, p.emergencyHandler.ListContacts)

	// 一键报警
	router.POST("/emergencies/sos", auth, writeLimiter, p.emergencyHandler.SOS)
	router.GET("/emergencies/mine", auth, readLimiter, p.emergencyHandler.ListMyAlerts)

	// 我的统计
	router.GET("/statistics/mine", auth, readLimiter, p.statisticsHandler.ListByUser)

	// ==================== 管理后台路由（需 pinche:audit 权限） ====================

	admin := router.Group("/admin")

	// 拼车行程管理
	admin.GET("/pinches", auditPerm, p.pincheHandler.AdminList)
	admin.GET("/pinches/:id", auditPerm, p.pincheHandler.AdminGetByID)
	admin.PUT("/pinches/:id/audit", auditPerm, p.pincheHandler.Audit)
	admin.PUT("/pinches/:id/status", auditPerm, p.pincheHandler.AdminUpdateStatus)
	admin.POST("/pinches/batch-audit", auditPerm, p.pincheHandler.BatchAudit)
	admin.POST("/pinches/batch-status", auditPerm, p.pincheHandler.BatchUpdateStatus)
	admin.POST("/pinches/batch-delete", auditPerm, p.pincheHandler.BatchDelete)

	// 行程管理
	admin.GET("/trips", auditPerm, p.tripHandler.AdminList)
	admin.PUT("/trips/:id/status", auditPerm, p.tripHandler.UpdateStatus)

	// 预订管理
	admin.PUT("/bookings/:id/status", auditPerm, p.bookingHandler.AdminUpdateStatus)

	// 车主认证管理
	admin.GET("/drivers", auditPerm, p.driverHandler.AdminList)
	admin.PUT("/drivers/:id/review", auditPerm, p.driverHandler.Review)
	admin.PUT("/drivers/:id/status", auditPerm, p.driverHandler.UpdateStatus)

	// 车辆管理
	admin.GET("/vehicles", auditPerm, p.vehicleHandler.AdminList)
	admin.PUT("/vehicles/:id/review", auditPerm, p.vehicleHandler.Review)
	admin.PUT("/vehicles/:id/status", auditPerm, p.vehicleHandler.UpdateStatus)

	// 支付管理
	admin.GET("/payments", auditPerm, p.paymentHandler.AdminList)
	admin.PUT("/payments/:id/status", auditPerm, p.paymentHandler.UpdateStatus)

	// 评价管理
	admin.GET("/ratings", auditPerm, p.ratingHandler.AdminList)
	admin.PUT("/ratings/:id/status", auditPerm, p.ratingHandler.UpdateStatus)

	// 保险管理
	admin.GET("/insurances", auditPerm, p.insuranceHandler.AdminList)
	admin.PUT("/insurances/:id/status", auditPerm, p.insuranceHandler.UpdateStatus)

	// 紧急联系/报警管理
	admin.GET("/emergencies", auditPerm, p.emergencyHandler.AdminListAlerts)
	admin.PUT("/emergencies/:id/handle", auditPerm, p.emergencyHandler.HandleAlert)
	admin.PUT("/emergencies/:id/status", auditPerm, p.emergencyHandler.UpdateAlertStatus)

	// 审核规则管理
	admin.GET("/audit-rules", auditPerm, p.auditRuleHandler.AdminList)
	admin.POST("/audit-rules", auditPerm, p.auditRuleHandler.Create)
	admin.PUT("/audit-rules/:id", auditPerm, p.auditRuleHandler.Update)
	admin.DELETE("/audit-rules/:id", auditPerm, p.auditRuleHandler.Delete)
	admin.PUT("/audit-rules/:id/status", auditPerm, p.auditRuleHandler.UpdateStatus)
	admin.POST("/audit-rules/:id/hit", auditPerm, p.auditRuleHandler.IncrHitCount)

	// 数据统计管理
	admin.GET("/statistics", auditPerm, p.statisticsHandler.AdminList)
	admin.POST("/statistics", auditPerm, p.statisticsHandler.Upsert)
	admin.DELETE("/statistics/:id", auditPerm, p.statisticsHandler.Delete)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
