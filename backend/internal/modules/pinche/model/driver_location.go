// Package model 拼车车主实时位置数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheDriverLocation 车主实时位置
type PincheDriverLocation struct {
	database.BaseModel // 仅 id/created_at/updated_at/deleted_at，无 region_id

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	TripID    *uint `gorm:"index" json:"trip_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`
	DriverID  uint  `gorm:"index;not null" json:"driver_id"`
	UserID    uint  `gorm:"index;not null" json:"user_id"`

	Lat          float64   `gorm:"type:decimal(10,7);default:0" json:"lat"`
	Lng          float64   `gorm:"type:decimal(10,7);default:0" json:"lng"`
	Speed        float64   `gorm:"type:decimal(6,2);default:0" json:"speed"`
	Heading      float64   `gorm:"type:decimal(5,2);default:0" json:"heading"`
	Altitude     float64   `gorm:"type:decimal(8,2);default:0" json:"altitude"`
	LocationTime time.Time `gorm:"index" json:"location_time"`

	IsShared bool `gorm:"default:false" json:"is_shared"`
}

// TableName 表名
func (PincheDriverLocation) TableName() string { return "pinche_driver_locations" }
