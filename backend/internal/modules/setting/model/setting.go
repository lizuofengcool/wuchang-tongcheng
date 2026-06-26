// Package model 系统设置数据模型
package model

import "wuchang-tongcheng/internal/pkg/database"

// Setting 系统设置
// 采用KV结构，按group分组，支持多种值类型
type Setting struct {
	database.RegionBaseModel
	Group       string `gorm:"size:50;index;not null" json:"group"`   // 分组（如 site, sms, payment）
	Key         string `gorm:"size:100;uniqueIndex:idx_group_key;not null" json:"key"` // 配置键
	Value       string `gorm:"type:text" json:"value"`                // 配置值
	ValueType   string `gorm:"size:20;default:string" json:"value_type"` // 值类型 string/number/bool/json
	Description string `gorm:"size:255" json:"description"`           // 描述
	Sort        int    `gorm:"default:0" json:"sort"`                  // 排序
}

// TableName 表名
func (Setting) TableName() string {
	return "settings"
}
