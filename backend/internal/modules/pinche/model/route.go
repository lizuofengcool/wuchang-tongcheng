// Package model 拼车路线数据模型
// 起点终点+途经点+距离+时长
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// PincheRoute 拼车路线表
type PincheRoute struct {
	database.RegionBaseModel

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	RouteName string `gorm:"size:128" json:"route_name"`

	// 起终点
	OriginAddress      string  `gorm:"size:255" json:"origin_address"`
	OriginLat          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lat"`
	OriginLng          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lng"`
	DestinationAddress string  `gorm:"size:255" json:"destination_address"`
	DestinationLat     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lat"`
	DestinationLng     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lng"`

	// 途经点（JSONB 数组：[{address,lat,lng}]）
	Waypoints JSONB `gorm:"type:jsonb" json:"waypoints"`

	// 行程信息
	DistanceKM     float64 `gorm:"type:decimal(10,2);default:0" json:"distance_km"`
	DurationMin    int     `gorm:"not null;default:0" json:"duration_min"`
	EstimatedPrice float64 `gorm:"type:decimal(12,2);default:0" json:"estimated_price"`
	TollFee        float64 `gorm:"type:decimal(12,2);default:0" json:"toll_fee"`

	IsCommon bool `gorm:"default:false;index" json:"is_common"`
	UseCount int  `gorm:"not null;default:0" json:"use_count"`
	Status   int  `gorm:"not null;default:1" json:"status"`
}

// TableName 表名
func (PincheRoute) TableName() string { return "pinche_routes" }
