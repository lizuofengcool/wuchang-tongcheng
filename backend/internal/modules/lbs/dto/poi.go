// Package dto LBS地图中台 - POI 数据传输对象
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// POIInfo POI 详情响应
type POIInfo struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	Category     string     `json:"category"`
	Phone        string     `json:"phone"`
	Icon         string     `json:"icon"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	Latitude     float64    `json:"latitude"`
	Longitude    float64    `json:"longitude"`
	Distance     float64    `json:"distance,omitempty"`
	UserID       uint       `json:"user_id"`
	Source       string     `json:"source"`
	ExternalID   string     `json:"external_id"`
	Tags         interface{} `json:"tags"`
	Extra        interface{} `json:"extra"`
	PublishedAt  *time.Time `json:"published_at"`
	RegionID     uint       `json:"region_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreatePOIRequest 创建 POI 请求
type CreatePOIRequest struct {
	Name       string      `json:"name" binding:"required,max=200"`
	Address    string      `json:"address" binding:"max=500"`
	Category   string      `json:"category" binding:"max=64"`
	Phone      string      `json:"phone" binding:"max=32"`
	Icon       string      `json:"icon" binding:"max=255"`
	Latitude   float64     `json:"latitude" binding:"required"`
	Longitude  float64     `json:"longitude" binding:"required"`
	Source     string      `json:"source" binding:"omitempty,oneof=manual amap import"`
	ExternalID string      `json:"external_id" binding:"max=64"`
	Tags       interface{} `json:"tags"`
	Extra      interface{} `json:"extra"`
	Status     int         `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// UpdatePOIRequest 更新 POI 请求
type UpdatePOIRequest struct {
	Name       *string     `json:"name" binding:"omitempty,max=200"`
	Address    *string     `json:"address" binding:"max=500"`
	Category   *string     `json:"category" binding:"max=64"`
	Phone      *string     `json:"phone" binding:"max=32"`
	Icon       *string     `json:"icon" binding:"max=255"`
	Latitude   *float64    `json:"latitude"`
	Longitude  *float64    `json:"longitude"`
	Source     *string     `json:"source" binding:"omitempty,oneof=manual amap import"`
	ExternalID *string     `json:"external_id" binding:"max=64"`
	Tags       interface{} `json:"tags"`
	Extra      interface{} `json:"extra"`
	Status     *int        `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// POIListRequest POI 列表请求
type POIListRequest struct {
	Keyword    string  `form:"keyword" json:"keyword"`
	Category   string  `form:"category" json:"category"`
	Status     *int    `form:"status" json:"status"`
	Source     string  `form:"source" json:"source"`
	UserID     uint    `form:"user_id" json:"user_id"`
	Sort       string `form:"sort" json:"sort"`
	utils.Pagination
}

// POIAdminListRequest 管理后台 POI 列表请求
type POIAdminListRequest struct {
	RegionID   uint   `form:"region_id" json:"region_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	Category   string `form:"category" json:"category"`
	Status     *int   `form:"status" json:"status"`
	Source     string `form:"source" json:"source"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}
