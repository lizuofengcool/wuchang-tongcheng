// Package model 标签表
// 商户标签/评价标签/美食标签/服务标签
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Tag 标签表
type Dh114Tag struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name    string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_tags_name_type" json:"name"`        // 标签名
	Code    string `gorm:"size:64;not null;default:'';index" json:"code"`                            // 标签编码
	TagType string `gorm:"size:32;not null;default:'business';index;uniqueIndex:uniq_dh114_tags_name_type" json:"tag_type"` // business/review/food/service
	ParentID uint  `gorm:"not null;default:0;index" json:"parent_id"`                                // 父标签 ID
	Level   int    `gorm:"not null;default:1" json:"level"`                                          // 层级
	Icon    string `gorm:"size:64;not null;default:''" json:"icon"`                                  // 图标
	Color   string `gorm:"size:32;not null;default:''" json:"color"`                                // 颜色
	Sort    int    `gorm:"not null;default:0;index" json:"sort"`                                    // 排序
	Status  int    `gorm:"default:1;index" json:"status"`                                            // 0禁用 1启用
	UseCount int   `gorm:"not null;default:0" json:"use_count"`                                     // 使用次数
}

// TableName 表名
func (Dh114Tag) TableName() string { return "dh114_tags" }
