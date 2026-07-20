// Package dto 同城拼车出行数据传输对象 - 通用 DTO
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
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

// ShareResponse 行程分享响应
type ShareResponse struct {
	ShareToken string `json:"share_token"`
	ShareURL   string `json:"share_url"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// BoardingResponse 上车码响应
type BoardingResponse struct {
	BoardingCode string `json:"boarding_code"`
	BookingID    uint   `json:"booking_id"`
}

// ===== 审核/状态 =====

// AuditRequest 审核操作请求（M 端）
type AuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// AdminUpdateStatusRequest 管理后台强制下架/恢复
type AdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3 4"`
}

// UpdateStatusRequest C 端用户更新状态
type UpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1 2 3 4"`
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
	Status int    `json:"status" binding:"oneof=0 1 2 3 4"`
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

// ===== 时间区间 =====

// DateRangeRequest 日期区间请求
type DateRangeRequest struct {
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
	utils.Pagination
}

// SimpleResponse 简单响应
type SimpleResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
