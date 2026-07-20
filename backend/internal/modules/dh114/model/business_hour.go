// Package model 营业时间表（按周设置）
// 每天可设置不同营业时间，支持 24 小时营业
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114BusinessHour 营业时间表
type Dh114BusinessHour struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Dh114ID    uint   `gorm:"not null;index;uniqueIndex:uniq_dh114_business_hours_dh114_weekday" json:"dh114_id"` // 商户 ID
	BusinessID uint   `gorm:"not null;default:0;index" json:"business_id"`                                    // 商户详情 ID
	Weekday    int    `gorm:"not null;default:1;uniqueIndex:uniq_dh114_business_hours_dh114_weekday" json:"weekday"` // 星期几 1-7（周一到周日）
	OpenTime   string `gorm:"size:8;not null;default:'09:00'" json:"open_time"`                              // 开门时间（HH:MM）
	CloseTime  string `gorm:"size:8;not null;default:'22:00'" json:"close_time"`                             // 关门时间（HH:MM）
	IsOpen     bool   `gorm:"not null;default:true;index" json:"is_open"`                                    // 是否营业
	Is24H      bool   `gorm:"not null;default:false" json:"is_24h"`                                          // 是否 24 小时营业
}

// TableName 表名
func (Dh114BusinessHour) TableName() string { return "dh114_business_hours" }
