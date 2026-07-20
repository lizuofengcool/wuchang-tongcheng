// Package model 商家分类表（对标大众点评/美团）
// 餐饮/零售/服务/娱乐/酒店/医疗/教育/生活服务 全行业分类
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 分类状态常量 ===
const (
	CategoryStatusDraft     = 0 // 草稿
	CategoryStatusPublished = 1 // 已发布
	CategoryStatusOffline   = 2 // 已下架
)

// Dh114Category 商家分类表
type Dh114Category struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null" json:"name"`                                  // 分类名
	Code        string `gorm:"size:64;not null;uniqueIndex" json:"code"`                      // 分类编码
	ParentID    uint   `gorm:"not null;default:0;index" json:"parent_id"`                     // 父分类 ID（0 为一级）
	Level       int    `gorm:"not null;default:1;index" json:"level"`                         // 层级 1/2/3
	BusinessType string `gorm:"size:32;not null;default:'other';index" json:"business_type"` // 关联商户类型
	Icon        string `gorm:"size:64;not null;default:''" json:"icon"`                       // 图标
	Color       string `gorm:"size:32;not null;default:''" json:"color"`                      // 颜色
	Description string `gorm:"size:500;not null;default:''" json:"description"`              // 描述
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                         // 排序
	Status      int    `gorm:"default:1;index" json:"status"`                                 // 0草稿 1已发布 2下架
	BusinessCount int  `gorm:"not null;default:0" json:"business_count"`                     // 商户数
}

// TableName 表名
func (Dh114Category) TableName() string { return "dh114_categories" }
