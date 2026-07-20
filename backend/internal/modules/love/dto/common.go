// Package dto love 相亲交友数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案第六章：对标 Soul / 陌陌 / 探探 / 百合网
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 通用响应 =====

// IDResponse 仅返回 ID 的响应
type LoveIDResponse struct {
	ID uint `json:"id"`
}

// ===== 审核/状态 =====

// LoveAuditRequest 审核操作请求（M 端）
type LoveAuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// LoveAdminUpdateStatusRequest 管理后台强制下架/恢复
type LoveAdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3"` // 0禁用 1正常 2冻结 3注销
}

// LoveUpdateStatusRequest C 端用户更新状态
type LoveUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 3"` // 0禁用 1正常 3注销
}

// ===== 批量操作 =====

// LoveBatchAuditRequest 批量审核请求
type LoveBatchAuditRequest struct {
	IDs         []uint `json:"ids" binding:"required,min=1"`
	AuditStatus int    `json:"audit_status" binding:"oneof=1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// LoveBatchStatusUpdateRequest 批量状态变更请求
type LoveBatchStatusUpdateRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status" binding:"oneof=0 1 2 3"`
}

// LoveBatchDeleteRequest 批量删除请求
type LoveBatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// LoveBatchResultResponse 批量操作结果
type LoveBatchResultResponse struct {
	Total     int    `json:"total"`
	Success   int    `json:"success"`
	Failed    int    `json:"failed"`
	FailedIDs []uint `json:"failed_ids,omitempty"`
}

// ===== 导出 =====

// LoveExportRequest 导出 Excel/CSV 请求
type LoveExportRequest struct {
	Format     string `json:"format" binding:"required,oneof=excel csv"`
	Status     *int   `json:"status"`
	AuditStatus *int  `json:"audit_status"`
	Gender     *int   `json:"gender"`
	UserID     uint   `json:"user_id"`
	Keyword    string `json:"keyword"`
}

// ===== 搜索推荐 =====

// LoveAdvancedSearchRequest 高级搜索请求
type LoveAdvancedSearchRequest struct {
	Keyword      string  `form:"keyword" json:"keyword"`
	Gender       *int    `form:"gender" json:"gender"` // 1男 2女
	MinAge       int     `form:"min_age" json:"min_age"`
	MaxAge       int     `form:"max_age" json:"max_age"`
	MinHeight    int     `form:"min_height" json:"min_height"`
	MaxHeight    int     `form:"max_height" json:"max_height"`
	Education    string  `form:"education" json:"education"`
	Occupation   string  `form:"occupation" json:"occupation"`
	Marriage     string  `form:"marriage" json:"marriage"`
	House        string  `form:"house" json:"house"`
	Car          string  `form:"car" json:"car"`
	Hometown     string  `form:"hometown" json:"hometown"`
	Residence    string  `form:"residence" json:"residence"`
	Verified     *bool   `form:"verified" json:"verified"`
	RealNameVerified *bool `form:"real_name_verified" json:"real_name_verified"`
	PhotoVerified *bool  `form:"photo_verified" json:"photo_verified"`
	VideoVerified *bool  `form:"video_verified" json:"video_verified"`
	MemberLevel  *int    `form:"member_level" json:"member_level"`
	Featured     *bool   `form:"featured" json:"featured"`
	Picked       *bool   `form:"picked" json:"picked"`
	Online       *bool   `form:"online" json:"online"`
	HasVoice     *bool   `form:"has_voice" json:"has_voice"`
	Sort         string  `form:"sort" json:"sort"` // latest/popular/distance/age/active
	Latitude     float64 `form:"latitude" json:"latitude"`
	Longitude    float64 `form:"longitude" json:"longitude"`
	RadiusKm     float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}

// LoveSimilarResponse 相似推荐响应
type LoveSimilarResponse struct {
	LoveID    uint    `json:"love_id"`
	Nickname  string  `json:"nickname"`
	Avatar    string  `json:"avatar"`
	Age       int     `json:"age"`
	Gender    int     `json:"gender"`
	Residence string  `json:"residence"`
	Score     float64 `json:"score"`
}

// LoveHotSearchResponse 热搜词响应
type LoveHotSearchResponse struct {
	Keyword     string `json:"keyword"`
	SearchCount int64  `json:"search_count"`
	Rank        int    `json:"rank"`
}

// ===== 通用统计 =====

// LoveReviewStats 评价统计
type LoveReviewStats struct {
	TotalReviews int     `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
}

// ===== 时间区间 =====

// LoveDateRangeRequest 日期区间请求
type LoveDateRangeRequest struct {
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
}

// ===== 通用操作响应 =====

// LoveActionResponse 通用操作响应（喜欢/拉黑/举报等）
type LoveActionResponse struct {
	Success   bool   `json:"success"`
	Action    string `json:"action"`
	Matched   bool   `json:"matched,omitempty"`
	MatchID   uint   `json:"match_id,omitempty"`
	SessionID uint   `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}
