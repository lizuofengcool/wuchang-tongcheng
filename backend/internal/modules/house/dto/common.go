// Package dto 同城房屋租售数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家/安居客/我爱我家/58房产
package dto

import (
	"time"
)

// ===== 通用响应 =====

// IDResponse 仅返回 ID 的响应（创建/更新成功）
type IDResponse struct {
	ID uint `json:"id"`
}

// FavResponse 收藏操作响应
type FavResponse struct {
	HasFaved bool `json:"has_faved"`
	FavCount int  `json:"fav_count"`
}

// ===== 审核相关 =====

// AuditRequest 审核操作请求（M 端）
type AuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// AdminUpdateStatusRequest 管理后台强制下架/恢复
type AdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=1 2 3 4"` // 1发布 2下架 3过期 4删除
}

// UpdateHouseStatusRequest 房源上下架请求
type UpdateHouseStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3"` // 0草稿 1发布 2下架 3过期
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
	Status int    `json:"status" binding:"oneof=1 2 3 4"`
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

// ===== 评价统计 =====

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

// ===== 推广请求 =====

// HousePromotionRequest 推广请求
type HousePromotionRequest struct {
	PromotionLevel int     `json:"promotion_level" binding:"gte=0,lte=10"`
	TrafficWeight  float64 `json:"traffic_weight" binding:"gte=0,lte=9.99"`
	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	RealHouseVerified bool `json:"real_house_verified"`
}

// ===== 相似推荐响应 =====

// SimilarHouseResponse 相似房源推荐响应
type SimilarHouseResponse struct {
	HouseID    uint    `json:"house_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Price      float64 `json:"price"`
	Layout     string  `json:"layout"`
	BuildingArea float64 `json:"building_area"`
	CommunityName string `json:"community_name"`
	Similarity float64 `json:"similarity"`
}
