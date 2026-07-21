// Package model 商户中台 - 商户类目表（merchant_categories）
// 树形结构：parent_id 自引用，全行业分类
// 类目为全局数据，使用 BaseModel（无 region_id 隔离）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Category 商户类目表
type Category struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 树形结构 ===
	ParentID uint   `gorm:"not null;default:0;index" json:"parent_id"` // 父类目 ID（0=根）
	Name     string `gorm:"size:64;not null;default:'';index" json:"name"` // 类目名
	Icon     string `gorm:"size:255;not null;default:''" json:"icon"`      // 图标 URL

	// === 显示与排序 ===
	Sort   int `gorm:"not null;default:0;index" json:"sort"`     // 排序
	Status int `gorm:"default:1;index" json:"status"`           // 0禁用 1启用
}

// TableName 表名
func (Category) TableName() string { return "merchant_categories" }
