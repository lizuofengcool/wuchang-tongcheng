// Package dto 同城拼车出行数据传输对象 - 路线
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// RouteInfo 路线详情响应
type RouteInfo struct {
	ID                  uint        `json:"id"`
	RegionID            uint        `json:"region_id"`
	UserID              uint        `json:"user_id"`
	RouteName           string      `json:"route_name"`
	OriginAddress       string      `json:"origin_address"`
	OriginLat           float64     `json:"origin_lat"`
	OriginLng           float64     `json:"origin_lng"`
	DestinationAddress  string      `json:"destination_address"`
	DestinationLat      float64     `json:"destination_lat"`
	DestinationLng      float64     `json:"destination_lng"`
	Waypoints           interface{} `json:"waypoints"`
	DistanceKM          float64     `json:"distance_km"`
	DurationMin         int         `json:"duration_min"`
	EstimatedPrice      float64     `json:"estimated_price"`
	TollFee             float64     `json:"toll_fee"`
	IsCommon            bool        `json:"is_common"`
	UseCount            int         `json:"use_count"`
	Status              int         `json:"status"`
}

// CreateRouteRequest 创建路线请求
type CreateRouteRequest struct {
	RouteName           string      `json:"route_name" binding:"required,max=128"`
	OriginAddress       string      `json:"origin_address" binding:"required,max=255"`
	OriginLat           float64     `json:"origin_lat"`
	OriginLng           float64     `json:"origin_lng"`
	DestinationAddress  string      `json:"destination_address" binding:"required,max=255"`
	DestinationLat      float64     `json:"destination_lat"`
	DestinationLng      float64     `json:"destination_lng"`
	Waypoints           interface{} `json:"waypoints"`
	DistanceKM          float64     `json:"distance_km"`
	DurationMin         int         `json:"duration_min"`
	EstimatedPrice      float64     `json:"estimated_price"`
	TollFee             float64     `json:"toll_fee"`
	IsCommon            bool        `json:"is_common"`
}

// UpdateRouteRequest 更新路线请求
type UpdateRouteRequest struct {
	RouteName           *string      `json:"route_name"`
	OriginAddress       *string      `json:"origin_address"`
	OriginLat           *float64     `json:"origin_lat"`
	OriginLng           *float64     `json:"origin_lng"`
	DestinationAddress  *string      `json:"destination_address"`
	DestinationLat      *float64     `json:"destination_lat"`
	DestinationLng      *float64     `json:"destination_lng"`
	Waypoints           interface{}  `json:"waypoints"`
	DistanceKM          *float64     `json:"distance_km"`
	DurationMin         *int         `json:"duration_min"`
	EstimatedPrice      *float64     `json:"estimated_price"`
	TollFee             *float64     `json:"toll_fee"`
	IsCommon            *bool        `json:"is_common"`
	Status              *int         `json:"status"`
}

// RouteListRequest 路线列表查询请求
type RouteListRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	IsCommon *bool  `form:"is_common" json:"is_common"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}
