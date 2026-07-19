// Package model 房源分类表（对标贝壳）
// 整租/合租/独栋/公寓/别墅/写字楼/商铺
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 分类状态常量 ===
const (
	CategoryStatusDisabled = 0 // 禁用
	CategoryStatusEnabled  = 1 // 启用
)

// HouseCategory 房源分类表
type HouseCategory struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name         string `gorm:"size:64;not null" json:"name"`                                    // 分类名
	Code         string `gorm:"size:64;not null;uniqueIndex" json:"code"`                        // 分类编码
	ParentID     uint   `gorm:"not null;default:0;index" json:"parent_id"`                       // 父分类 ID（0=一级）
	Level        int    `gorm:"not null;default:1;index" json:"level"`                           // 层级 1/2/3
	ListingType  string `gorm:"size:16;not null;default:'rent';index" json:"listing_type"`       // rent/sale/transfer
	PropertyType string `gorm:"size:32;not null;default:'residential';index" json:"property_type"` // residential/apartment/villa/loft/office/shop
	Icon         string `gorm:"size:64;not null;default:''" json:"icon"`                         // 图标
	Color        string `gorm:"size:16;not null;default:'#409EFF'" json:"color"`                 // 颜色
	Description  string `gorm:"size:500;not null;default:''" json:"description"`                // 描述
	Sort         int    `gorm:"not null;default:0;index" json:"sort"`                            // 排序
	Status       int    `gorm:"default:1;index" json:"status"`                                   // 0禁用 1启用
	HouseCount   int    `gorm:"not null;default:0" json:"house_count"`                           // 房源数
}

// TableName 表名（house_ 前缀）
func (HouseCategory) TableName() string { return "house_categories" }
