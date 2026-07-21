// Package model LBS地图中台数据模型
// 依据 v3.2.1 架构方案第 4.8 节：高德定位/附近检索/距离排序/POI/路线规划/地理围栏/分站区域隔离
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离）
// 依据 docs/架构设计/PostGIS地理字段规范.md：未启用 PostGIS 时使用 DECIMAL(10,7) 降级方案
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 状态常量 ===
const (
	LBSPoiStatusOffline   = 0 // 下架
	LBSPoiStatusOnline    = 1 // 上线
	LBSPoiStatusPending   = 2 // 待审
	LBSPoiStatusRejected  = 3 // 拒绝
	LBSPoiStatusDeleted   = 4 // 删除
)

// === 区域状态常量 ===
const (
	LBSRegionStatusDisabled = 0 // 禁用
	LBSRegionStatusEnabled  = 1 // 启用
)

// === 围栏类型常量 ===
const (
	LBSGeofenceTypeCircle  = "circle"  // 圆形
	LBSGeofenceTypePolygon = "polygon" // 多边形
)

// === 围栏状态常量 ===
const (
	LBSGeofenceStatusDisabled = 0 // 禁用
	LBSGeofenceStatusEnabled  = 1 // 启用
)

// ============================================================
// JSONB 字段类型（兼容 GORM 与 PostgreSQL jsonb）
// ============================================================

// JSONB 包装 []byte 以便与 PostgreSQL jsonb 类型交互
// 实现 driver.Valuer 与 sql.Scanner 接口，支持 GORM 自动映射
type JSONB []byte

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[:0], v...)
		return nil
	case string:
		*j = []byte(v)
		return nil
	}
	return errors.New("lbs.JSONB.Scan: unsupported source type")
}

// MarshalJSON 实现 json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil || len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现 json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("lbs.JSONB.UnmarshalJSON: nil pointer")
	}
	*j = append((*j)[:0], data...)
	return nil
}

// Bytes 返回底层字节切片的只读副本
func (j JSONB) Bytes() []byte {
	if j == nil {
		return nil
	}
	out := make([]byte, len(j))
	copy(out, j)
	return out
}

// String 返回字符串形式
func (j JSONB) String() string {
	if j == nil || len(j) == 0 {
		return ""
	}
	return string(j)
}

// Parse 尝试反序列化为目标对象
func (j JSONB) Parse(v interface{}) error {
	if j == nil || len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, v)
}

// FromJSON 从任意 Go 对象构造 JSONB
func FromJSON(v interface{}) (JSONB, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSONB(b), nil
}

// ============================================================
// 专用结构定义（用于 JSONB 字段反序列化）
// ============================================================

// GeofencePoint 多边形围栏顶点
type GeofencePoint struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// RegionBoundaryItem 区域边界顶点（多边形外环）
type RegionBoundaryItem struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// ============================================================
// POI 兴趣点模型
// ============================================================

// POI POI 兴趣点模型（lbs_pois 表）
// 存储地图标注点：商户/地标/兴趣点等
type POI struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Name      string `gorm:"size:200;not null" json:"name"`        // POI 名称
	Address   string `gorm:"size:500;not null;default:''" json:"address"` // 详细地址
	Category  string `gorm:"size:64;not null;default:'';index" json:"category"` // 分类（如 restaurant/landmark/bus_station）
	Phone     string `gorm:"size:32;not null;default:''" json:"phone"` // 联系电话
	Icon      string `gorm:"size:255;not null;default:''" json:"icon"` // 图标 URL
	Status    int    `gorm:"default:1;index" json:"status"`           // 0下架 1上线 2待审 3拒绝 4删除

	// === 地理位置（DECIMAL 降级方案，精度 1cm） ===
	Latitude  float64 `gorm:"type:decimal(10,7);default:0;index" json:"latitude"`  // 纬度
	Longitude float64 `gorm:"type:decimal(10,7);default:0;index" json:"longitude"` // 经度

	// === 扩展信息 ===
	UserID     uint       `gorm:"index" json:"user_id"`            // 创建者 ID
	Source     string     `gorm:"size:32;not null;default:'manual';index" json:"source"` // 来源：manual/amap/import
	ExternalID string     `gorm:"size:64;not null;default:'';index" json:"external_id"` // 外部 ID（如高德 POI ID）
	Tags       JSONB      `gorm:"type:jsonb" json:"tags"`           // 标签数组
	Extra      JSONB      `gorm:"type:jsonb" json:"extra"`         // 扩展字段
	PublishedAt *time.Time `gorm:"index" json:"published_at"`       // 发布时间

	// Distance 仅在 Nearby 查询时由 SQL 计算并回填，非持久化字段（公里）
	Distance float64 `gorm:"-" json:"-"`
}

// TableName 表名
func (POI) TableName() string { return "lbs_pois" }
