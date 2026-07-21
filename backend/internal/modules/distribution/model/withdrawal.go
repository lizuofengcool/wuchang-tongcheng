// Package model 分销合伙人中台数据模型 - 提现记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 提现状态常量 ===
const (
	WithdrawalStatusPending  = 0 // 申请中
	WithdrawalStatusAudited  = 1 // 已审核
	WithdrawalStatusPaid     = 2 // 已打款
	WithdrawalStatusRejected = 3 // 已拒绝
)

// Withdrawal 提现记录模型（distribution_withdrawals 表）
type Withdrawal struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 关联 ===
	PartnerID uint `gorm:"index;not null" json:"partner_id"` // 合伙人 ID

	// === 金额 ===
	Amount float64 `gorm:"type:decimal(12,2);default:0" json:"amount"` // 提现金额

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`        // 0申请中 1已审核 2已打款 3已拒绝
	BankInfo    JSONB      `gorm:"type:jsonb" json:"bank_info"`          // 银行/账户信息 JSONB
	AuditReason string     `gorm:"size:500;default:''" json:"audit_reason"` // 审核备注
	AuditedAt   *time.Time `gorm:"index" json:"audited_at"`              // 审核时间
	PaidAt      *time.Time `gorm:"index" json:"paid_at"`                 // 打款时间
}

// TableName 表名
func (Withdrawal) TableName() string { return "distribution_withdrawals" }
