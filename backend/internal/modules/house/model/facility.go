// Package model 配套设施表（对标贝壳）
// 家具/家电/独立卫浴/阳台/车位/暖气等
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 设施分类常量 ===
const (
	FacilityCategoryIndoor     = "indoor"      // 室内设施
	FacilityCategoryOutdoor    = "outdoor"     // 室外设施
	FacilityCategoryAppliance  = "appliance"   // 家电
	FacilityCategoryFurniture  = "furniture"   // 家具
	FacilityCategoryNetwork    = "network"     // 网络
	FacilityCategorySecurity   = "security"    // 安全
	FacilityCategoryParking    = "parking"     // 停车
)

// === 设施状态常量 ===
const (
	FacilityStatusDisabled = 0 // 禁用
	FacilityStatusEnabled  = 1 // 启用
)

// HouseFacility 配套设施表
type HouseFacility struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null" json:"name"`                       // 设施名称
	Code        string `gorm:"size:64;not null;uniqueIndex" json:"code"`           // 设施编码
	Category    string `gorm:"size:32;not null;default:'indoor';index" json:"category"` // indoor/outdoor/appliance/furniture/network/security/parking
	Icon        string `gorm:"size:64;not null;default:''" json:"icon"`            // 图标
	Color       string `gorm:"size:16;not null;default:'#409EFF'" json:"color"`    // 颜色
	Background  string `gorm:"size:32;not null;default:''" json:"background"`      // 背景色
	Description string `gorm:"size:500;not null;default:''" json:"description"`   // 描述
	Status      int    `gorm:"default:1;index" json:"status"`                      // 0禁用 1启用
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`               // 排序
	UseCount    int    `gorm:"not null;default:0" json:"use_count"`                // 使用数
	IsHot       bool   `gorm:"not null;default:false;index" json:"is_hot"`         // 是否热门
	CreatorID   uint   `gorm:"not null;default:0;index" json:"creator_id"`         // 创建人 ID
}

// TableName 表名（house_ 前缀）
func (HouseFacility) TableName() string { return "house_facilities" }
