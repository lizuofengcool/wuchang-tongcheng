// Package model 拼车支付记录数据模型（含 ETC 支付）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PinchePayment 支付记录
type PinchePayment struct {
	database.RegionBaseModel

	PincheID  uint `gorm:"index;not null" json:"pinche_id"`
	BookingID uint `gorm:"index;not null" json:"booking_id"`

	PaymentNo string `gorm:"size:32;index" json:"payment_no"`
	PayerID   uint   `gorm:"index;not null" json:"payer_id"`
	PayerName string `gorm:"size:50" json:"payer_name"`
	PayeeID   uint   `gorm:"index;not null" json:"payee_id"`
	PayeeName string `gorm:"size:50" json:"payee_name"`

	Amount       float64 `gorm:"type:decimal(12,2);default:0" json:"amount"`
	InsuranceFee float64 `gorm:"type:decimal(12,2);default:0" json:"insurance_fee"`
	ServiceFee   float64 `gorm:"type:decimal(12,2);default:0" json:"service_fee"`
	TotalAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`

	PaymentMethod string `gorm:"size:16;not null;default:'cash';index" json:"payment_method"` // cash/wechat/alipay/balance/etc
	PaymentStatus int    `gorm:"default:0;index" json:"payment_status"`                       // 0待支付 1已支付 2已退款 3已失败
	PaidAt        *time.Time `gorm:"index" json:"paid_at"`
	RefundedAt    *time.Time `gorm:"index" json:"refunded_at"`
	ThirdPartyNo  string     `gorm:"size:64" json:"third_party_no"`
	RefundAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`

	// ETC 专用
	ETCLaneID    string     `gorm:"size:32" json:"etc_lane_id"`
	ETCEntryTime *time.Time `gorm:"index" json:"etc_entry_time"`
	ETCExitTime  *time.Time `gorm:"index" json:"etc_exit_time"`
}

// TableName 表名
func (PinchePayment) TableName() string { return "pinche_payments" }
