// Package dto 同城招聘求职数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/拉勾/58招聘
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
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
	Status int `json:"status" binding:"oneof=1 2 3 4"` // 1发布 2关闭 3下架 4过期
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

// ===== 搜索推荐 =====

// AdvancedSearchRequest 高级搜索请求
type AdvancedSearchRequest struct {
	Keyword         string  `form:"keyword" json:"keyword"`
	CategoryID      uint    `form:"category_id" json:"category_id"`
	RecruitmentType string  `form:"recruitment_type" json:"recruitment_type"`
	EmploymentType  string  `form:"employment_type" json:"employment_type"`
	Education       string  `form:"education" json:"education"`
	WorkYearMin     int     `form:"work_year_min" json:"work_year_min"`
	WorkYearMax     int     `form:"work_year_max" json:"work_year_max"`
	SalaryMin       float64 `form:"salary_min" json:"salary_min"`
	SalaryMax       float64 `form:"salary_max" json:"salary_max"`
	WorkCity        string  `form:"work_city" json:"work_city"`
	CompanyID       uint    `form:"company_id" json:"company_id"`
	SkillIDs        []uint  `form:"skill_ids" json:"skill_ids"`
	BenefitIDs      []uint  `form:"benefit_ids" json:"benefit_ids"`
	AllowRemote     *bool   `form:"allow_remote" json:"allow_remote"`
	IsUrgent        *bool   `form:"is_urgent" json:"is_urgent"`
	Featured        *bool   `form:"featured" json:"featured"`
	Verified        *bool   `form:"verified" json:"verified"`
	Sort            string  `form:"sort" json:"sort"` // latest/salary_asc/salary_desc/popular/distance
	Latitude        float64 `form:"latitude" json:"latitude"`
	Longitude       float64 `form:"longitude" json:"longitude"`
	RadiusKm        float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}

// HotSearchResponse 热搜词响应
type HotSearchResponse struct {
	Keyword     string `json:"keyword"`
	SearchCount int64  `json:"search_count"`
	Rank        int    `json:"rank"`
}

// SimilarJobResponse 相似职位推荐响应
type SimilarJobResponse struct {
	JobID      uint    `json:"job_id"`
	Title      string  `json:"title"`
	CompanyName string  `json:"company_name"`
	SalaryMin  float64 `json:"salary_min"`
	SalaryMax  float64 `json:"salary_max"`
	SalaryUnit string  `json:"salary_unit"`
	WorkCity   string  `json:"work_city"`
	CoverImage string  `json:"cover_image"`
	Similarity float64 `json:"similarity"`
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
