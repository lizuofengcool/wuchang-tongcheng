// Package dto 营销活动中台数据传输对象 - 广告位（ad 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// AdPositionInfo 广告位详情响应
type AdPositionInfo struct {
	ID           uint       `json:"id"`
	RegionID     uint       `json:"region_id"`
	PositionCode string     `json:"position_code"`
	Title        string     `json:"title"`
	ImageURL     string     `json:"image_url"`
	LinkURL      string     `json:"link_url"`
	Sort         int        `json:"sort"`
	StartAt      *time.Time `json:"start_at"`
	EndAt        *time.Time `json:"end_at"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateAdPositionRequest 创建广告位请求
type CreateAdPositionRequest struct {
	PositionCode string     `json:"position_code" binding:"required,max=50"`
	Title        string     `json:"title" binding:"required,max=100"`
	ImageURL     string     `json:"image_url" binding:"required,max=500"`
	LinkURL      string     `json:"link_url" binding:"max=500"`
	Sort         int        `json:"sort"`
	StartAt      *time.Time `json:"start_at"`
	EndAt        *time.Time `json:"end_at"`
	Status       int        `json:"status"`
}

// UpdateAdPositionRequest 更新广告位请求
type UpdateAdPositionRequest struct {
	PositionCode *string     `json:"position_code" binding:"omitempty,max=50"`
	Title        *string     `json:"title" binding:"omitempty,max=100"`
	ImageURL     *string     `json:"image_url" binding:"omitempty,max=500"`
	LinkURL      *string     `json:"link_url" binding:"omitempty,max=500"`
	Sort         *int        `json:"sort"`
	StartAt      *time.Time  `json:"start_at"`
	EndAt        *time.Time  `json:"end_at"`
	Status       *int        `json:"status"`
}

// AdPositionListRequest 广告位列表请求
type AdPositionListRequest struct {
	PositionCode string `form:"position_code" json:"position_code"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// AdPositionByCodeRequest 按位置编码获取广告请求
type AdPositionByCodeRequest struct {
	PositionCode string `form:"position_code" json:"position_code" binding:"required"`
	utils.Pagination
}
