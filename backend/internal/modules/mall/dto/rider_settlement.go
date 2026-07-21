// Package dto 同城商城 - 骑手结算 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RiderSettlementInfo 结算单详情响应
type RiderSettlementInfo struct {
	ID           uint       `json:"id"`
	RiderID      uint       `json:"rider_id"`
	Period       string     `json:"period"`
	TotalOrders  int        `json:"total_orders"`
	TotalAmount  float64    `json:"total_amount"`
	TotalFee     float64    `json:"total_fee"`
	TotalTip     float64    `json:"total_tip"`
	PlatformFee  float64    `json:"platform_fee"`
	NetAmount    float64    `json:"net_amount"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	SettledAt    *time.Time `json:"settled_at"`
	WithdrawnAt  *time.Time `json:"withdrawn_at"`
	AuditReason  string     `json:"audit_reason"`
	RegionID     uint       `json:"region_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RiderSettlementListRequest 结算单列表请求
type RiderSettlementListRequest struct {
	utils.Pagination
	RiderID  uint   `form:"rider_id" json:"rider_id"`
	Period   string `form:"period" json:"period"`
	Status   *int   `form:"status" json:"status"`
	RegionID uint   `form:"region_id" json:"region_id"`
}

// RiderSettlementGenerateRequest 生成结算单请求（M 端触发）
type RiderSettlementGenerateRequest struct {
	RiderID    uint   `json:"rider_id" binding:"required"`
	Period     string `json:"period" binding:"required,max=20"` // YYYY-MM
	PlatformFee float64 `json:"platform_fee"` // 平台抽成（金额）
}

// RiderSettlementAuditRequest 结算单审核请求（M 端）
type RiderSettlementAuditRequest struct {
	Status      int    `json:"status" binding:"oneof=1 2"` // 1已结算 2已提现
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// RiderSettlementWithdrawRequest 提现申请请求（骑手 C 端）
type RiderSettlementWithdrawRequest struct {
	ID uint `json:"id" binding:"required"` // 结算单 ID
}

// RiderSettlementStatsResponse 结算统计响应
type RiderSettlementStatsResponse struct {
	TotalSettlements int64   `json:"total_settlements"`
	TotalNetAmount   float64 `json:"total_net_amount"`
	PendingCount     int64   `json:"pending_count"`
	PendingAmount    float64 `json:"pending_amount"`
	SettledCount     int64   `json:"settled_count"`
	SettledAmount    float64 `json:"settled_amount"`
	WithdrawnCount   int64   `json:"withdrawn_count"`
	WithdrawnAmount  float64 `json:"withdrawn_amount"`
}
