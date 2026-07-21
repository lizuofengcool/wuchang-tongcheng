// Package model LBS地图中台数据模型 - 地理围栏
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Geofence 地理围栏模型（lbs_geofences 表）
// 用于配送范围、电子围栏、考勤打卡区域等
type Geofence struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	RegionID  uint   `gorm:"not null;index" json:"region_id"`        // 关联区域 ID
	Name      string `gorm:"size:100;not null" json:"name"`           // 围栏名称
	Type      string `gorm:"size:16;not null;default:'circle';index" json:"type"` // 类型：circle/polygon
	Status    int    `gorm:"default:1;index" json:"status"`           // 0禁用 1启用
	Sort      int    `gorm:"not null;default:0" json:"sort"`           // 排序值
	Description string `gorm:"size:500;not null;default:''" json:"description"` // 描述

	// === 圆形围栏参数 ===
	CenterLat float64 `gorm:"type:decimal(10,7);default:0;index" json:"center_lat"` // 圆心纬度
	CenterLng float64 `gorm:"type:decimal(10,7);default:0;index" json:"center_lng"` // 圆心经度
	Radius    float64 `gorm:"type:decimal(10,2);default:0" json:"radius"`           // 半径（米）

	// === 多边形围栏参数 ===
	Points    JSONB  `gorm:"type:jsonb" json:"points"` // 多边形顶点 [{lat,lng},...]

	// === 扩展信息 ===
	OwnerID   uint   `gorm:"index" json:"owner_id"`             // 所有者 ID
	OwnerType string `gorm:"size:32;not null;default:'';index" json:"owner_type"` // 所有者类型：shop/agent/daojia
	Extra     JSONB  `gorm:"type:jsonb" json:"extra"`           // 扩展字段
}

// TableName 表名
func (Geofence) TableName() string { return "lbs_geofences" }
