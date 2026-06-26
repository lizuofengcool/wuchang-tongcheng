// Package model 分类信息数据模型
package model

import "wuchang-tongcheng/internal/pkg/database"

// Category 分类信息模型
// 用于本地生活分类（如：二手房、招聘、二手物品、车辆等）
type Category struct {
	database.RegionBaseModel
	Name     string `gorm:"size:50;not null" json:"name"`          // 分类名称
	Icon     string `gorm:"size:255" json:"icon"`                  // 分类图标
	ParentID uint   `gorm:"index;default:0" json:"parent_id"`      // 父级ID，0为顶级
	Level    int    `gorm:"default:1" json:"level"`                // 层级
	Sort     int    `gorm:"default:0" json:"sort"`                // 排序（越大越靠前）
	Status   int    `gorm:"default:1" json:"status"`              // 状态 1正常 0禁用
}

// TableName 表名
func (Category) TableName() string {
	return "categories"
}
