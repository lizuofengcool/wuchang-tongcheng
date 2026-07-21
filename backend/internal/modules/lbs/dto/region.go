// Package dto LBS地图中台 - 区域分站数据传输对象
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RegionInfo 区域详情响应
type RegionInfo struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	CityCode    string     `json:"city_code"`
	ParentID    uint       `json:"parent_id"`
	Level       int        `json:"level"`
	Path        string     `json:"path"`
	Sort        int        `json:"sort"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	CenterLat   float64    `json:"center_lat"`
	CenterLng   float64    `json:"center_lng"`
	Boundary    interface{} `json:"boundary"`
	AdCode      string     `json:"ad_code"`
	ZipCode     string     `json:"zip_code"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRegionRequest 创建区域请求
type CreateRegionRequest struct {
	Name        string      `json:"name" binding:"required,max=100"`
	CityCode    string      `json:"city_code" binding:"max=20"`
	ParentID    uint        `json:"parent_id"`
	Level       int         `json:"level" binding:"omitempty,oneof=1 2 3 4"`
	Sort        int         `json:"sort"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
	CenterLat   float64     `json:"center_lat"`
	CenterLng   float64     `json:"center_lng"`
	Boundary    interface{} `json:"boundary"`
	AdCode      string      `json:"ad_code" binding:"max=20"`
	ZipCode     string      `json:"zip_code" binding:"max=20"`
	Description string      `json:"description" binding:"max=500"`
}

// UpdateRegionRequest 更新区域请求
type UpdateRegionRequest struct {
	Name        *string     `json:"name" binding:"omitempty,max=100"`
	CityCode    *string     `json:"city_code" binding:"max=20"`
	ParentID    *uint       `json:"parent_id"`
	Level       *int        `json:"level" binding:"omitempty,oneof=1 2 3 4"`
	Sort        *int        `json:"sort"`
	Status      *int        `json:"status" binding:"omitempty,oneof=0 1"`
	CenterLat   *float64    `json:"center_lat"`
	CenterLng   *float64    `json:"center_lng"`
	Boundary    interface{} `json:"boundary"`
	AdCode      *string     `json:"ad_code" binding:"max=20"`
	ZipCode     *string     `json:"zip_code" binding:"max=20"`
	Description *string     `json:"description" binding:"max=500"`
}

// RegionListRequest 区域列表请求
type RegionListRequest struct {
	Keyword   string `form:"keyword" json:"keyword"`
	CityCode  string `form:"city_code" json:"city_code"`
	ParentID  uint   `form:"parent_id" json:"parent_id"`
	Level     int    `form:"level" json:"level"`
	Status    *int   `form:"status" json:"status"`
	utils.Pagination
}
