// Package dto LBS地图中台 - 地理围栏数据传输对象
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// GeofenceInfo 围栏详情响应
type GeofenceInfo struct {
	ID          uint       `json:"id"`
	RegionID    uint       `json:"region_id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Sort        int        `json:"sort"`
	Description string     `json:"description"`
	CenterLat   float64    `json:"center_lat"`
	CenterLng   float64    `json:"center_lng"`
	Radius      float64    `json:"radius"`
	Points      interface{} `json:"points"`
	OwnerID     uint       `json:"owner_id"`
	OwnerType   string     `json:"owner_type"`
	Extra       interface{} `json:"extra"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateGeofenceRequest 创建围栏请求
type CreateGeofenceRequest struct {
	RegionID    uint       `json:"region_id"`
	Name        string     `json:"name" binding:"required,max=100"`
	Type        string     `json:"type" binding:"required,oneof=circle polygon"`
	Status      int        `json:"status" binding:"omitempty,oneof=0 1"`
	Sort        int        `json:"sort"`
	Description string     `json:"description" binding:"max=500"`
	CenterLat   float64    `json:"center_lat"`
	CenterLng   float64    `json:"center_lng"`
	Radius      float64    `json:"radius"`
	Points      interface{} `json:"points"`
	OwnerID     uint       `json:"owner_id"`
	OwnerType   string     `json:"owner_type" binding:"max=32"`
	Extra       interface{} `json:"extra"`
}

// UpdateGeofenceRequest 更新围栏请求
type UpdateGeofenceRequest struct {
	RegionID    *uint       `json:"region_id"`
	Name        *string     `json:"name" binding:"omitempty,max=100"`
	Type        *string     `json:"type" binding:"omitempty,oneof=circle polygon"`
	Status      *int        `json:"status" binding:"omitempty,oneof=0 1"`
	Sort        *int        `json:"sort"`
	Description *string     `json:"description" binding:"max=500"`
	CenterLat   *float64    `json:"center_lat"`
	CenterLng   *float64    `json:"center_lng"`
	Radius      *float64    `json:"radius"`
	Points      interface{} `json:"points"`
	OwnerID     *uint       `json:"owner_id"`
	OwnerType   *string     `json:"owner_type" binding:"max=32"`
	Extra       interface{} `json:"extra"`
}

// GeofenceListRequest 围栏列表请求
type GeofenceListRequest struct {
	RegionID  uint   `form:"region_id" json:"region_id"`
	OwnerID   uint   `form:"owner_id" json:"owner_id"`
	OwnerType string `form:"owner_type" json:"owner_type"`
	Type      string `form:"type" json:"type"`
	Status    *int   `form:"status" json:"status"`
	Keyword   string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CheckPointRequest 检查点是否在围栏内请求
type CheckPointRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// CheckPointResponse 检查点是否在围栏内响应
type CheckPointResponse struct {
	Inside   bool   `json:"inside"`
	GeofenceID uint `json:"geofence_id,omitempty"`
	Name     string `json:"name,omitempty"`
}
