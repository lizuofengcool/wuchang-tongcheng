// Package lbs LBS地图中台模块插件
// 提供高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离等能力
// 依据 v3.2.1 架构方案第 4.8 节：LBS 地图中台
// 依据需求文档 1.10：4 维数据隔离（region_id）
// 依据 docs/架构设计/PostGIS地理字段规范.md：未启用 PostGIS 时使用 DECIMAL(10,7) 降级方案
package lbs

import (
	"context"
	"os"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/lbs/handler"
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/modules/lbs/repository"
	"wuchang-tongcheng/internal/modules/lbs/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin LBS地图中台模块插件
type Plugin struct {
	name    string
	version string

	// 4 个 Handler
	poiHandler      *handler.POIHandler
	regionHandler   *handler.RegionHandler
	geofenceHandler *handler.GeofenceHandler
	lbsHandler      *handler.LBSHandler
}

// NewPlugin 创建 LBS地图中台模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "lbs", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Meta 返回插件元信息
func (p *Plugin) Meta() plugin.PluginMeta {
	return plugin.PluginMeta{
		Name:        "lbs",
		DisplayName: "LBS地图中台",
		Category:    "middleware",
		Description: "LBS地图中台：高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离",
		Version:     p.version,
		Author:      "wuchang",
	}
}

// Init 初始化插件
// 注入依赖链 repository → service → handler
//
// 注意：3 张表由 backend/migrations/026_lbs_full.sql 创建，
// 此处仅 AutoMigrate 主表 POI（与 dh114/mall 模块保持一致），
// 子表由 SQL 脚本创建（避免 GORM AutoMigrate 约束名不一致错误）。
// 高德 API Key 通过环境变量 LBS_AMAP_KEY 传入；为空时路线规划降级为本地估算。
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移主表 lbs_pois（lbs_regions/lbs_geofences 由 026_lbs_full.sql 创建）
	if err := db.AutoMigrate(&model.POI{}); err != nil {
		return err
	}

	// ===== 依赖注入：Repository 层（3 个） =====
	poiRepo := repository.NewPOIRepository(db)
	regionRepo := repository.NewRegionRepository(db)
	geofenceRepo := repository.NewGeofenceRepository(db)

	// ===== 依赖注入：Service 层（4 个） =====
	poiSvc := service.NewPOIService(poiRepo)
	regionSvc := service.NewRegionService(regionRepo)
	geofenceSvc := service.NewGeofenceService(geofenceRepo)
	amapKey := os.Getenv("LBS_AMAP_KEY")
	routeSvc := service.NewRouteService(amapKey)

	// ===== 依赖注入：Handler 层（4 个） =====
	p.poiHandler = handler.NewPOIHandler(poiSvc)
	p.regionHandler = handler.NewRegionHandler(regionSvc)
	p.geofenceHandler = handler.NewGeofenceHandler(geofenceSvc)
	p.lbsHandler = handler.NewLBSHandler(routeSvc)

	return nil
}

// RegisterRoutes 注册插件路由
// 路由前缀由插件管理器统一添加为 /api/v1/lbs
//
// 路由分组（共 30+ API）：
//   - 公开路由（C 端浏览，无需登录）：POI 列表/详情/附近；区域列表/详情/按城市码/按经纬度；围栏列表/详情/检查点；距离计算/路线规划
//   - 需登录路由（C 端发布）：POI/区域/围栏 CRUD
//   - 管理后台路由（需 lbs:manage 权限）：POI/区域/围栏 管理与状态
//
// 注意：固定路径（/nearby /mine /by-parent /by-city-code /by-location 等）
// 必须注册在 /:id 之前，否则会被 :id 参数路由吞掉。
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())
	readLimiter := coreRouter.WrapGin(middleware.RateLimit(60, 60, "lbs_read"))
	writeLimiter := coreRouter.WrapGin(middleware.RateLimit(10, 60, "lbs_write"))
	nearbyLimiter := coreRouter.WrapGin(middleware.RateLimit(30, 60, "lbs_nearby"))
	managePerm := coreRouter.WrapGin(middleware.RequirePermission("lbs:manage"))

	// ==================== 公开路由（C 端浏览，无需登录） ====================

	// POI（pois）
	router.GET("/pois", readLimiter, p.poiHandler.List)
	router.GET("/pois/nearby", nearbyLimiter, p.poiHandler.ListNearby)
	router.GET("/pois/:id", readLimiter, p.poiHandler.GetByID)

	// 区域分站（regions）
	router.GET("/regions", readLimiter, p.regionHandler.List)
	router.GET("/regions/by-parent/:parent_id", readLimiter, p.regionHandler.ListByParent)
	router.GET("/regions/by-city-code", readLimiter, p.regionHandler.FindByCityCode)
	router.GET("/regions/by-location", readLimiter, p.regionHandler.FindByLocation)
	router.GET("/regions/:id", readLimiter, p.regionHandler.GetByID)

	// 围栏（geofences）
	router.GET("/geofences", readLimiter, p.geofenceHandler.List)
	router.GET("/geofences/by-region/:region_id", readLimiter, p.geofenceHandler.ListByRegion)
	router.GET("/geofences/by-owner/:owner_id", readLimiter, p.geofenceHandler.ListByOwner)
	router.GET("/geofences/:id", readLimiter, p.geofenceHandler.GetByID)
	router.POST("/geofences/:id/check-point", readLimiter, p.geofenceHandler.CheckPoint)
	router.POST("/geofences/check-point-in-region", readLimiter, p.geofenceHandler.CheckPointInRegion)

	// 距离与路线规划（公共）
	router.GET("/distance", readLimiter, p.lbsHandler.CalculateDistance)
	router.GET("/route", readLimiter, p.lbsHandler.PlanRoute)

	// ==================== 需登录路由（C 端发布） ====================

	// POI CRUD
	router.POST("/pois", auth, writeLimiter, p.poiHandler.Create)
	router.PUT("/pois/:id", auth, writeLimiter, p.poiHandler.Update)
	router.DELETE("/pois/:id", auth, writeLimiter, p.poiHandler.Delete)
	router.GET("/pois/mine", auth, readLimiter, p.poiHandler.ListMine)

	// 区域 CRUD
	router.POST("/regions", auth, writeLimiter, p.regionHandler.Create)
	router.PUT("/regions/:id", auth, writeLimiter, p.regionHandler.Update)
	router.DELETE("/regions/:id", auth, writeLimiter, p.regionHandler.Delete)

	// 围栏 CRUD
	router.POST("/geofences", auth, writeLimiter, p.geofenceHandler.Create)
	router.PUT("/geofences/:id", auth, writeLimiter, p.geofenceHandler.Update)
	router.DELETE("/geofences/:id", auth, writeLimiter, p.geofenceHandler.Delete)

	// ==================== 管理后台路由（需 lbs:manage 权限） ====================

	admin := router.Group("/admin")

	// POI 管理
	admin.GET("/pois", managePerm, p.poiHandler.AdminList)
	admin.GET("/pois/categories", managePerm, p.poiHandler.AdminListCategories)
	admin.GET("/pois/:id", managePerm, p.poiHandler.AdminGetByID)
	admin.PUT("/pois/:id/status", managePerm, p.poiHandler.AdminUpdateStatus)
	admin.DELETE("/pois/:id", managePerm, p.poiHandler.AdminDelete)

	// 区域管理
	admin.GET("/regions", managePerm, p.regionHandler.AdminList)
	admin.PUT("/regions/:id/status", managePerm, p.regionHandler.AdminUpdateStatus)

	// 围栏管理
	admin.GET("/geofences", managePerm, p.geofenceHandler.AdminList)
	admin.PUT("/geofences/:id/status", managePerm, p.geofenceHandler.AdminUpdateStatus)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保 Plugin 实现了 plugin.Plugin 接口
var _ plugin.Plugin = (*Plugin)(nil)

// init 自动注册插件（幂等，导入包即注册）
func init() {
	plugin.AutoRegister(NewPlugin())
}
