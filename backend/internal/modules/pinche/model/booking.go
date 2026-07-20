// Package model 拼车预订记录数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheBooking 拼车预订记录
type PincheBooking struct {
	database.RegionBaseModel

	PincheID  uint   `gorm:"index;not null" json:"pinche_id"`
	BookingNo string `gorm:"size:32;index" json:"booking_no"`

	// 乘客
	PassengerID     uint   `gorm:"index;not null" json:"passenger_id"`
	PassengerName   string `gorm:"size:50" json:"passenger_name"`
	PassengerPhone  string `gorm:"size:20" json:"passenger_phone"`
	PassengerAvatar string `gorm:"size:255" json:"passenger_avatar"`

	// 车主
	DriverID    uint   `gorm:"index;not null" json:"driver_id"`
	DriverName  string `gorm:"size:50" json:"driver_name"`
	DriverPhone string `gorm:"size:20" json:"driver_phone"`

	// 预订信息
	Seats           int     `gorm:"not null;default:1" json:"seats"`
	PickupLocation  string  `gorm:"size:255" json:"pickup_location"`
	PickupLat       float64 `gorm:"type:decimal(10,7);default:0" json:"pickup_lat"`
	PickupLng       float64 `gorm:"type:decimal(10,7);default:0" json:"pickup_lng"`
	DropoffLocation string  `gorm:"size:255" json:"dropoff_location"`
	DropoffLat      float64 `gorm:"type:decimal(10,7);default:0" json:"dropoff_lat"`
	DropoffLng      float64 `gorm:"type:decimal(10,7);default:0" json:"dropoff_lng"`

	// 金额
	UnitPrice     float64 `gorm:"type:decimal(12,2);default:0" json:"unit_price"`
	TotalAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`
	InsuranceFee  float64 `gorm:"type:decimal(12,2);default:0" json:"insurance_fee"`
	ServiceFee    float64 `gorm:"type:decimal(12,2);default:0" json:"service_fee"`

	// 状态
	Status      int    `gorm:"default:0;index" json:"status"` // 0待支付 1已支付 2已上车 3已完成 4已取消 5已退款
	PaymentID   *uint  `gorm:"index" json:"payment_id"`
	BoardingCode string `gorm:"size:16" json:"boarding_code"`

	// 时间节点
	PaidAt      *time.Time `gorm:"index" json:"paid_at"`
	BoardedAt   *time.Time `gorm:"index" json:"boarded_at"`
	CompletedAt *time.Time `gorm:"index" json:"completed_at"`
	CancelledAt *time.Time `gorm:"index" json:"cancelled_at"`
	CancelReason string    `gorm:"size:500" json:"cancel_reason"`
	CancelledBy *uint      `gorm:"index" json:"cancelled_by"`
}

// TableName 表名
func (PincheBooking) TableName() string { return "pinche_bookings" }
