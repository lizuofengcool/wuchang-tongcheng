// Package model 同城商城 - 收货地址
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 用户隔离）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Address 收货地址表
type Address struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	UserID uint `gorm:"index;not null" json:"user_id"` // 用户 ID

	// === 收货人信息 ===
	Name   string `gorm:"size:50;not null;default:''" json:"name"`     // 收货人姓名
	Phone  string `gorm:"size:32;not null;default:'';index" json:"phone"` // 收货人手机号
	ZipCode string `gorm:"size:20;not null;default:''" json:"zip_code"`   // 邮编

	// === 行政区划 ===
	Province     string `gorm:"size:64;not null;default:'';index" json:"province"` // 省
	City         string `gorm:"size:64;not null;default:'';index" json:"city"`     // 市
	District     string `gorm:"size:64;not null;default:'';index" json:"district"` // 区/县
	ProvinceCode string `gorm:"size:16;not null;default:''" json:"province_code"`  // 省级码
	CityCode     string `gorm:"size:16;not null;default:''" json:"city_code"`      // 市级码
	DistrictCode string `gorm:"size:16;not null;default:''" json:"district_code"`  // 区县级码

	// === 详细地址 ===
	Detail      string  `gorm:"size:500;not null;default:''" json:"detail"`        // 详细地址
	Latitude    float64 `gorm:"type:decimal(10,7);default:0" json:"latitude"`      // 纬度
	Longitude   float64 `gorm:"type:decimal(10,7);default:0" json:"longitude"`     // 经度

	// === 标签与默认 ===
	Tag       string `gorm:"size:32;not null;default:''" json:"tag"`        // 标签（家/公司/学校）
	IsDefault int    `gorm:"not null;default:0;index" json:"is_default"`    // 0非默认 1默认

	// === 状态 ===
	Status int `gorm:"default:1;index" json:"status"` // 0禁用 1启用
}

// TableName 表名
func (Address) TableName() string { return "mall_addresses" }
