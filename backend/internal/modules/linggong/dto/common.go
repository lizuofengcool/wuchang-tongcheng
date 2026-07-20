// Package dto 同城零工兼职数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案第六章：对标斗米/青团兼职/兼职猫/猪八戒
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
	Status int `json:"status" binding:"oneof=1 2 3 4 5 6 7"` // 1发布 2下架 3过期 4删除 5满员 6关闭 7完成
}

// UpdateStatusRequest C 端用户更新状态
type UpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3 5 6 7"` // 0草稿 1发布 2下架 3过期 5满员 6关闭 7完成
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
	Status int    `json:"status" binding:"oneof=1 2 3 4 5 6 7"`
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
	Keyword        string  `form:"keyword" json:"keyword"`
	LinggongType   string  `form:"linggong_type" json:"linggong_type"`
	PublisherType  string  `form:"publisher_type" json:"publisher_type"`
	BillingType    string  `form:"billing_type" json:"billing_type"`
	Settlement     string  `form:"settlement" json:"settlement"`
	MinSalary      float64 `form:"min_salary" json:"min_salary"`
	MaxSalary      float64 `form:"max_salary" json:"max_salary"`
	SalaryNegotiable *bool `form:"salary_negotiable" json:"salary_negotiable"`
	WorkStartDate  string  `form:"work_start_date" json:"work_start_date"`
	WorkEndDate    string  `form:"work_end_date" json:"work_end_date"`
	MinAge         int     `form:"min_age" json:"min_age"`
	MaxAge         int     `form:"max_age" json:"max_age"`
	NeedGender     string  `form:"need_gender" json:"need_gender"`
	Education      string  `form:"education" json:"education"`
	Province      string  `form:"province" json:"province"`
	City          string  `form:"city" json:"city"`
	District      string  `form:"district" json:"district"`
	WorkLocationType string `form:"work_location_type" json:"work_location_type"`
	SkillID       uint    `form:"skill_id" json:"skill_id"`
	EmployerID    uint    `form:"employer_id" json:"employer_id"`
	Featured      *bool   `form:"featured" json:"featured"`
	Picked        *bool   `form:"picked" json:"picked"`
	Verified      *bool   `form:"verified" json:"verified"`
	EmployerVerified *bool `form:"employer_verified" json:"employer_verified"`
	Sort          string  `form:"sort" json:"sort"` // latest/salary_asc/salary_desc/popular/distance
	Latitude      float64 `form:"latitude" json:"latitude"`
	Longitude     float64 `form:"longitude" json:"longitude"`
	RadiusKm      float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}

// SimilarLinggongResponse 相似岗位推荐响应
type SimilarLinggongResponse struct {
	LinggongID uint    `json:"linggong_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	SalaryMin  float64 `json:"salary_min"`
	SalaryMax  float64 `json:"salary_max"`
	CompanyName string  `json:"company_name"`
	City       string  `json:"city"`
	BillingType string `json:"billing_type"`
	Settlement string  `json:"settlement"`
	Similarity float64 `json:"similarity"`
}

// HotSearchResponse 热搜词响应
type HotSearchResponse struct {
	Keyword     string `json:"keyword"`
	SearchCount int64  `json:"search_count"`
	Rank        int    `json:"rank"`
}

// ===== 通用统计 =====

// RatingStats 评价统计
type RatingStats struct {
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
