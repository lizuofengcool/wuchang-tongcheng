// Package model 地区数据模型
package model

import "wuchang-tongcheng/internal/pkg/database"

// Region 地区模型
type Region struct {
	database.BaseModel
	Name     string `gorm:"size:50;not null" json:"name"`           // 地区名称
	Code     string `gorm:"size:20;uniqueIndex;not null" json:"code"` // 地区编码
	ParentID uint   `gorm:"index;default:0" json:"parent_id"`       // 父级ID，0为顶级
	Level    int    `gorm:"default:1" json:"level"`                 // 层级 1省 2市 3区县
	Sort     int    `gorm:"default:0" json:"sort"`                 // 排序
	Status   int    `gorm:"default:1" json:"status"`               // 状态 1正常 0禁用
}

// TableName 表名
func (Region) TableName() string {
	return "regions"
}
