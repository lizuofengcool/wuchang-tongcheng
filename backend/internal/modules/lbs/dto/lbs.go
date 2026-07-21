// Package dto LBS地图中台 - 公共查询数据传输对象
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// NearbyRequest 附近检索请求
// GET /api/v1/lbs/pois/nearby?lat=&lng=&radius_km=
type NearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	Category  string  `form:"category" json:"category"`
	Keyword   string  `form:"keyword" json:"keyword"`
	Sort      string  `form:"sort" json:"sort"` // distance_asc / created_desc
	utils.Pagination
}

// DistanceRequest 距离计算请求
// GET /api/v1/lbs/distance?from_lat=&from_lng=&to_lat=&to_lng=
type DistanceRequest struct {
	FromLat float64 `form:"from_lat" json:"from_lat" binding:"required"`
	FromLng float64 `form:"from_lng" json:"from_lng" binding:"required"`
	ToLat   float64 `form:"to_lat" json:"to_lat" binding:"required"`
	ToLng   float64 `form:"to_lng" json:"to_lng" binding:"required"`
}

// DistanceResponse 距离计算响应
type DistanceResponse struct {
	// 直线距离（Haversine 公式，公里）
	StraightKm float64 `json:"straight_km"`
	// 直线距离（米）
	StraightMeter float64 `json:"straight_meter"`
}

// RouteRequest 路线规划请求
// GET /api/v1/lbs/route?from_lat=&from_lng=&to_lat=&to_lng=&mode=
type RouteRequest struct {
	FromLat float64 `form:"from_lat" json:"from_lat" binding:"required"`
	FromLng float64 `form:"from_lng" json:"from_lng" binding:"required"`
	ToLat   float64 `form:"to_lat" json:"to_lat" binding:"required"`
	ToLng   float64 `form:"to_lng" json:"to_lng" binding:"required"`
	Mode   string  `form:"mode" json:"mode"` // driving/walking/transit/riding
}

// RouteResponse 路线规划响应
type RouteResponse struct {
	Mode         string  `json:"mode"`
	DistanceKm   float64 `json:"distance_km"`    // 路线距离（公里）
	DurationMin  int     `json:"duration_min"`   // 预计耗时（分钟）
	StraightKm   float64 `json:"straight_km"`    // 直线距离（公里）
	Source       string  `json:"source"`         // 数据来源：local/amap
	Steps        []RouteStep `json:"steps,omitempty"`
}

// RouteStep 路线步骤
type RouteStep struct {
	Instruction string  `json:"instruction"`
	DistanceKm  float64 `json:"distance_km"`
	DurationMin int     `json:"duration_min"`
}

// RegionQueryRequest 根据经纬度判断分站请求
// GET /api/v1/lbs/regions/by-location?lat=&lng=
type RegionQueryRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
}

// RegionQueryResponse 分站判断响应
type RegionQueryResponse struct {
	RegionID  uint   `json:"region_id"`
	Name      string `json:"name"`
	CityCode  string `json:"city_code"`
	AdCode    string `json:"ad_code"`
	Inside    bool   `json:"inside"`
	Source    string `json:"source"`
}
