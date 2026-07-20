// Package model 拼车取消记录数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheCancel 取消记录
type PincheCancel struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`

	CancelledBy   uint   `gorm:"index;not null" json:"cancelled_by"`
	CancelledRole string `gorm:"size:16;not null;default:'passenger'" json:"cancelled_role"` // driver/passenger
	CancelReason  string `gorm:"size:500" json:"cancel_reason"`
	CancelType    string `gorm:"size:32;not null;default:'user'" json:"cancel_type"`         // user/system/timeout
	CancelTime    *time.Time `gorm:"index" json:"cancel_time"`

	PenaltyAmount float64 `gorm:"type:decimal(12,2);default:0" json:"penalty_amount"`
	PenaltyPaid   bool    `gorm:"default:false" json:"penalty_paid"`
	RefundAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	RefundStatus  int     `gorm:"not null;default:0" json:"refund_status"` // 0待退 1已退 2失败
}

// TableName 表名
func (PincheCancel) TableName() string { return "pinche_cancels" }
