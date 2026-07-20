// Package model 拼车完成行程数据模型（含行程分享 share_token）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheTrip 完成行程
type PincheTrip struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`
	TripNo    string `gorm:"size:32;index" json:"trip_no"`

	DriverID    uint   `gorm:"index;not null" json:"driver_id"`
	DriverName  string `gorm:"size:50" json:"driver_name"`
	DriverPhone string `gorm:"size:20" json:"driver_phone"`
	PassengerID uint   `gorm:"index;not null" json:"passenger_id"`
	PassengerName string `gorm:"size:50" json:"passenger_name"`
	PassengerPhone string `gorm:"size:20" json:"passenger_phone"`
	VehicleID  *uint  `gorm:"index" json:"vehicle_id"`
	PlateNo    string `gorm:"size:16" json:"plate_no"`

	// 行程信息
	OriginAddress      string  `gorm:"size:255" json:"origin_address"`
	OriginLat          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lat"`
	OriginLng          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lng"`
	DestinationAddress string  `gorm:"size:255" json:"destination_address"`
	DestinationLat     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lat"`
	DestinationLng     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lng"`

	// 实际行程
	ActualPickupTime  *time.Time `gorm:"index" json:"actual_pickup_time"`
	ActualDropoffTime *time.Time `gorm:"index" json:"actual_dropoff_time"`
	ActualDistanceKM  float64    `gorm:"type:decimal(10,2);default:0" json:"actual_distance_km"`
	ActualDurationMin int        `gorm:"not null;default:0" json:"actual_duration_min"`
	PassengersCount   int        `gorm:"not null;default:1" json:"passengers_count"`

	// 金额
	FareAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"fare_amount"`
	TollFee      float64 `gorm:"type:decimal(12,2);default:0" json:"toll_fee"`
	TotalAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`

	// 行程分享
	ShareToken     string     `gorm:"size:64;index" json:"share_token"`
	ShareExpiresAt *time.Time `gorm:"index" json:"share_expires_at"`

	// 状态
	Status              int        `gorm:"default:0;index" json:"status"` // 0进行中 1已完成 2异常结束
	DriverConfirmedAt   *time.Time `gorm:"index" json:"driver_confirmed_at"`
	PassengerConfirmedAt *time.Time `gorm:"index" json:"passenger_confirmed_at"`
	CompletedAt         *time.Time `gorm:"index" json:"completed_at"`
}

// TableName 表名
func (PincheTrip) TableName() string { return "pinche_trips" }
