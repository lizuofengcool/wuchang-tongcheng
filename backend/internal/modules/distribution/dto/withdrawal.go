// Package dto 分销合伙人中台 - 提现请求/响应 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// WithdrawalInfo 提现记录详情响应
type WithdrawalInfo struct {
	ID          uint       `json:"id"`
	PartnerID   uint       `json:"partner_id"`
	Amount      float64    `json:"amount"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	BankInfo    interface{} `json:"bank_info"`
	AuditReason string     `json:"audit_reason"`
	AuditedAt   *time.Time `json:"audited_at"`
	PaidAt      *time.Time `json:"paid_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// WithdrawalApplyRequest 提现申请
type WithdrawalApplyRequest struct {
	Amount   float64     `json:"amount" binding:"required,min=0.01"`
	BankInfo interface{} `json:"bank_info" binding:"required"`
}

// WithdrawalAuditRequest 提现审核（通过/拒绝）
type WithdrawalAuditRequest struct {
	Status int    `json:"status" binding:"required,oneof=1 3"` // 1审核通过 3拒绝
	Reason string `json:"reason" binding:"max=500"`
}

// WithdrawalPayRequest 打款确认
type WithdrawalPayRequest struct {
	Reason string `json:"reason" binding:"max=500"` // 打款备注（可选）
}

// WithdrawalListRequest 提现列表请求
type WithdrawalListRequest struct {
	PartnerID uint   `form:"partner_id" json:"partner_id"`
	Status    *int   `form:"status" json:"status"`
	utils.Pagination
}
