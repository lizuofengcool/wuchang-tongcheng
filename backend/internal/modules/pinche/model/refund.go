// Package model 拼车退款记录数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheRefund 退款记录
type PincheRefund struct {
	database.RegionBaseModel

	PaymentID uint `gorm:"index;not null" json:"payment_id"`
	BookingID uint `gorm:"index;not null" json:"booking_id"`
	PincheID  uint `gorm:"index;not null" json:"pinche_id"`

	RefundNo     string  `gorm:"size:32;index" json:"refund_no"`
	RefundAmount float64 `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	RefundReason string  `gorm:"size:500" json:"refund_reason"`
	RefundStatus int     `gorm:"default:0;index" json:"refund_status"` // 0待退 1已退 2失败
	RefundMethod string  `gorm:"size:16;not null;default:'original'" json:"refund_method"`

	RefundedAt   *time.Time `gorm:"index" json:"refunded_at"`
	ThirdPartyNo string     `gorm:"size:64" json:"third_party_no"`

	OperatorID *uint      `gorm:"index" json:"operator_id"`
	HandledAt  *time.Time `gorm:"index" json:"handled_at"`
}

// TableName 表名
func (PincheRefund) TableName() string { return "pinche_refunds" }
