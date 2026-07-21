// Package model LBS地图中台数据模型 - 区域分站
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Region 区域分站模型（lbs_regions 表）
// 用于分站隔离、配送范围、行政区域判断
type Region struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at（无 region_id，自身即区域）

	// === 基础信息 ===
	Name        string `gorm:"size:100;not null" json:"name"`           // 区域名称
	CityCode    string `gorm:"size:20;not null;default:'';index" json:"city_code"` // 城市编码
	ParentID    uint   `gorm:"not null;default:0;index" json:"parent_id"` // 父区域 ID（0=根）
	Level       int    `gorm:"not null;default:1;index" json:"level"`    // 层级：1省 2市 3区 4乡镇
	Path        string `gorm:"size:500;not null;default:''" json:"path"` // 路径如 1,5,12
	Sort        int    `gorm:"not null;default:0" json:"sort"`           // 排序值
	Status      int    `gorm:"default:1;index" json:"status"`            // 0禁用 1启用

	// === 中心点（DECIMAL 降级方案） ===
	CenterLat   float64 `gorm:"type:decimal(10,7);default:0;index" json:"center_lat"` // 中心纬度
	CenterLng   float64 `gorm:"type:decimal(10,7);default:0;index" json:"center_lng"` // 中心经度

	// === 边界（多边形顶点列表 JSONB） ===
	Boundary    JSONB  `gorm:"type:jsonb" json:"boundary"` // 边界多边形 [{lat,lng},...]

	// === 扩展信息 ===
	AdCode      string `gorm:"size:20;not null;default:'';index" json:"ad_code"` // 行政区划代码
	ZipCode     string `gorm:"size:20;not null;default:''" json:"zip_code"`      // 邮编
	Description string `gorm:"size:500;not null;default:''" json:"description"`  // 描述
}

// TableName 表名
func (Region) TableName() string { return "lbs_regions" }
