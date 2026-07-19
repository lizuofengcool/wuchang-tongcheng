// Package dto 同城车辆买卖数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案第六章：对标瓜子/人人车/懂车帝/毛豆新车/易鑫车贷
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 通用响应 =====

// IDResponse 仅返回 ID 的响应
type IDResponse struct {
	ID uint `json:"id"`
}

// FavResponse 收藏操作响应
type FavResponse struct {
	HasFaved bool `json:"has_faved"`
	FavCount int  `json:"fav_count"`
}

// ===== 审核/状态 =====

// AuditRequest 审核操作请求（M 端）
type AuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// AdminUpdateStatusRequest 管理后台强制下架/恢复
type AdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=1 2 3 4 5"` // 1发布 2下架 3过期 4删除 5已售
}

// UpdateStatusRequest C 端用户更新状态
type UpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3 5"` // 0草稿 1发布 2下架 3过期 5已售
}

// ===== 批量操作 =====

// BatchAuditRequest 批量审核请求
type BatchAuditRequest struct {
	IDs         []uint `json:"ids" binding:"required,min=1"`
	AuditStatus int    `json:"audit_status" binding:"oneof=1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// BatchStatusUpdateRequest 批量状态变更请求
type BatchStatusUpdateRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status" binding:"oneof=1 2 3 4 5"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// BatchResultResponse 批量操作结果
type BatchResultResponse struct {
	Total     int    `json:"total"`
	Success   int    `json:"success"`
	Failed    int    `json:"failed"`
	FailedIDs []uint `json:"failed_ids,omitempty"`
}

// ===== 导出 =====

// ExportRequest 导出 Excel/CSV 请求
type ExportRequest struct {
	Format     string `json:"format" binding:"required,oneof=excel csv"`
	Status     *int   `json:"status"`
	CategoryID uint   `json:"category_id"`
	UserID     uint   `json:"user_id"`
	Keyword    string `json:"keyword"`
}

// ===== 搜索推荐 =====

// AdvancedSearchRequest 高级搜索请求
type AdvancedSearchRequest struct {
	Keyword         string  `form:"keyword" json:"keyword"`
	CategoryID      uint    `form:"category_id" json:"category_id"`
	BrandID         uint    `form:"brand_id" json:"brand_id"`
	ModelID         uint    `form:"model_id" json:"model_id"`
	CarType         string  `form:"car_type" json:"car_type"`
	ListingType     string  `form:"listing_type" json:"listing_type"`
	SourceType      string  `form:"source_type" json:"source_type"`
	FuelType        string  `form:"fuel_type" json:"fuel_type"`
	Transmission    string  `form:"transmission" json:"transmission"`
	MinPrice        float64 `form:"min_price" json:"min_price"`
	MaxPrice        float64 `form:"max_price" json:"max_price"`
	MinMileage      float64 `form:"min_mileage" json:"min_mileage"`
	MaxMileage      float64 `form:"max_mileage" json:"max_mileage"`
	MinYear         int     `form:"min_year" json:"min_year"`
	MaxYear         int     `form:"max_year" json:"max_year"`
	ConditionLevel  string  `form:"condition_level" json:"condition_level"`
	City            string  `form:"city" json:"city"`
	Featured        *bool   `form:"featured" json:"featured"`
	Picked          *bool   `form:"picked" json:"picked"`
	Verified        *bool   `form:"verified" json:"verified"`
	RealCarVerified *bool   `form:"real_car_verified" json:"real_car_verified"`
	PriceNegotiable *bool   `form:"price_negotiable" json:"price_negotiable"`
	Sort            string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/mileage_asc/year_desc/popular/distance
	Latitude        float64 `form:"latitude" json:"latitude"`
	Longitude       float64 `form:"longitude" json:"longitude"`
	RadiusKm        float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}

// SimilarCarResponse 相似车源推荐响应
type SimilarCarResponse struct {
	CarID      uint    `json:"car_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Price      float64 `json:"price"`
	BrandName  string  `json:"brand_name"`
	ModelName  string  `json:"model_name"`
	Year       int     `json:"year"`
	Mileage    float64 `json:"mileage"`
	Similarity float64 `json:"similarity"`
}

// HotSearchResponse 热搜词响应
type HotSearchResponse struct {
	Keyword     string `json:"keyword"`
	SearchCount int64  `json:"search_count"`
	Rank        int    `json:"rank"`
}

// ===== 通用统计 =====

// ReviewStats 评价统计
type ReviewStats struct {
	TotalReviews int     `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
}

// ===== 时间区间 =====

// DateRangeRequest 日期区间请求
type DateRangeRequest struct {
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
}
