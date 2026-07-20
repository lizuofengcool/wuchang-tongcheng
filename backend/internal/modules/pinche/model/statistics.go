// Package model 拼车统计数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheStatistic 统计
type PincheStatistic struct {
	database.RegionBaseModel

	StatDate time.Time `gorm:"type:date;index" json:"stat_date"`
	StatType string    `gorm:"size:16;not null;default:'daily';index" json:"stat_type"` // daily/weekly/monthly/total
	UserID   *uint     `gorm:"index" json:"user_id"`

	TotalTrips        int     `gorm:"not null;default:0" json:"total_trips"`
	CompletedTrips    int     `gorm:"not null;default:0" json:"completed_trips"`
	CancelledTrips    int     `gorm:"not null;default:0" json:"cancelled_trips"`
	TotalBookings     int     `gorm:"not null;default:0" json:"total_bookings"`
	CompletedBookings int     `gorm:"not null;default:0" json:"completed_bookings"`

	TotalRevenue float64 `gorm:"type:decimal(12,2);default:0" json:"total_revenue"`
	TotalRefund  float64 `gorm:"type:decimal(12,2);default:0" json:"total_refund"`
	AvgRating    float64 `gorm:"type:decimal(3,2);default:0" json:"avg_rating"`
	AvgPrice     float64 `gorm:"type:decimal(12,2);default:0" json:"avg_price"`

	TotalDistance float64 `gorm:"type:decimal(10,2);default:0" json:"total_distance"`
	TotalDuration int     `gorm:"not null;default:0" json:"total_duration"`

	TotalPassengers int `gorm:"not null;default:0" json:"total_passengers"`
	TotalDrivers    int `gorm:"not null;default:0" json:"total_drivers"`
}

// TableName 表名
func (PincheStatistic) TableName() string { return "pinche_statistics" }
