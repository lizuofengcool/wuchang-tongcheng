// Package model 同城商城 - 商品分类
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离）
// 树形分类（parent_id 自关联）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Category 商品分类表
type Category struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 树形结构 ===
	ParentID uint   `gorm:"index;not null;default:0" json:"parent_id"` // 父分类 ID（0=根）
	Name     string `gorm:"size:64;not null;default:'';index" json:"name"` // 分类名
	Icon     string `gorm:"size:255;not null;default:''" json:"icon"`     // 分类图标
	Cover    string `gorm:"size:255;not null;default:''" json:"cover"`    // 分类封面图
	Level    int    `gorm:"not null;default:1;index" json:"level"`        // 层级（1/2/3）
	Path     string `gorm:"size:500;not null;default:'';index" json:"path"` // 路径（如 1,5,12）

	// === 显示与排序 ===
	Sort       int    `gorm:"not null;default:0;index" json:"sort"`       // 排序
	Status     int    `gorm:"default:1;index" json:"status"`              // 0禁用 1启用
	IsShow     bool   `gorm:"default:true;index" json:"is_show"`          // 是否前台显示

	// === 统计 ===
	ProductCount int `gorm:"not null;default:0" json:"product_count"` // 商品数

	// === SEO ===
	Keywords    string `gorm:"size:255;not null;default:''" json:"keywords"`     // SEO 关键词
	Description string `gorm:"size:500;not null;default:''" json:"description"` // SEO 描述
}

// TableName 表名
func (Category) TableName() string { return "mall_categories" }
