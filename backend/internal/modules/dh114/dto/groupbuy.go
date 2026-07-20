// Package dto 同城114数据传输对象 - 团购
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// GroupbuyInfo 团购详情响应
type GroupbuyInfo struct {
	ID              uint       `json:"id"`
	GroupbuyNo      string     `json:"groupbuy_no"`
	Dh114ID         uint       `json:"dh114_id"`
	BusinessID      uint       `json:"business_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	CoverImage      string     `json:"cover_image"`
	Images          interface{} `json:"images"`
	OriginalPrice   float64    `json:"original_price"`
	GroupbuyPrice   float64    `json:"groupbuy_price"`
	Discount        float64    `json:"discount"`
	TotalCount      int        `json:"total_count"`
	SoldCount       int        `json:"sold_count"`
	PerUserLimit    int        `json:"per_user_limit"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidWeekdays   interface{} `json:"valid_weekdays"`
	UseInstructions interface{} `json:"use_instructions"`
	UseTimeRanges   interface{} `json:"use_time_ranges"`
	NeedReservation bool       `json:"need_reservation"`
	ViewCount       int        `json:"view_count"`
	FavCount        int        `json:"fav_count"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AuditStatus     int        `json:"audit_status"`
	AuditReason     string     `json:"audit_reason"`
	PublishedAt     *time.Time `json:"published_at"`
	Featured        bool       `json:"featured"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateGroupbuyRequest 创建团购请求
type CreateGroupbuyRequest struct {
	Dh114ID         uint       `json:"dh114_id" binding:"required"`
	Title           string     `json:"title" binding:"required,max=200"`
	Description     string     `json:"description"`
	CoverImage      string     `json:"cover_image" binding:"max=255"`
	Images          interface{} `json:"images"`
	OriginalPrice   float64    `json:"original_price" binding:"min=0"`
	GroupbuyPrice   float64    `json:"groupbuy_price" binding:"required,min=0"`
	TotalCount      int        `json:"total_count" binding:"min=0"`
	PerUserLimit    int        `json:"per_user_limit" binding:"min=0"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidWeekdays   interface{} `json:"valid_weekdays"`
	UseInstructions interface{} `json:"use_instructions"`
	UseTimeRanges   interface{} `json:"use_time_ranges"`
	NeedReservation bool       `json:"need_reservation"`
	Status          int        `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateGroupbuyRequest 更新团购请求
type UpdateGroupbuyRequest struct {
	Title           *string `json:"title" binding:"max=200"`
	Description     *string `json:"description"`
	CoverImage      *string `json:"cover_image" binding:"max=255"`
	Images          interface{} `json:"images"`
	OriginalPrice   *float64 `json:"original_price" binding:"min=0"`
	GroupbuyPrice   *float64 `json:"groupbuy_price" binding:"min=0"`
	TotalCount      *int    `json:"total_count" binding:"min=0"`
	PerUserLimit    *int    `json:"per_user_limit" binding:"min=0"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ValidStart      *time.Time `json:"valid_start"`
	ValidEnd        *time.Time `json:"valid_end"`
	ValidWeekdays   interface{} `json:"valid_weekdays"`
	UseInstructions interface{} `json:"use_instructions"`
	UseTimeRanges   interface{} `json:"use_time_ranges"`
	NeedReservation *bool   `json:"need_reservation"`
	Status          *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// GroupbuyListRequest 团购列表请求
type GroupbuyListRequest struct {
	Dh114ID    uint   `form:"dh114_id" json:"dh114_id"`
	Status     *int   `form:"status" json:"status"`
	Featured   *bool  `form:"featured" json:"featured"`
	MinPrice   float64 `form:"min_price" json:"min_price"`
	MaxPrice   float64 `form:"max_price" json:"max_price"`
	Keyword    string `form:"keyword" json:"keyword"`
	Sort       string `form:"sort" json:"sort"`
	utils.Pagination
}

// GroupbuyAdminListRequest 管理后台团购列表请求
type GroupbuyAdminListRequest struct {
	Dh114ID     uint   `form:"dh114_id" json:"dh114_id"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}
