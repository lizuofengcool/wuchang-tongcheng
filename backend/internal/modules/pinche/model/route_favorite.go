// Package model 拼车常用路线收藏数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheRouteFavorite 常用路线收藏
type PincheRouteFavorite struct {
	database.RegionBaseModel

	UserID       uint   `gorm:"index;not null" json:"user_id"`
	RouteID      *uint  `gorm:"index" json:"route_id"`
	FavoriteName string `gorm:"size:128" json:"favorite_name"`

	OriginAddress      string  `gorm:"size:255" json:"origin_address"`
	OriginLat          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lat"`
	OriginLng          float64 `gorm:"type:decimal(10,7);default:0" json:"origin_lng"`
	DestinationAddress string  `gorm:"size:255" json:"destination_address"`
	DestinationLat     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lat"`
	DestinationLng     float64 `gorm:"type:decimal(10,7);default:0" json:"destination_lng"`

	UseCount   int        `gorm:"not null;default:0" json:"use_count"`
	LastUsedAt *time.Time `gorm:"index" json:"last_used_at"`
}

// TableName 表名
func (PincheRouteFavorite) TableName() string { return "pinche_route_favorites" }
