// Package model 分销合伙人中台数据模型 - 合伙人等级
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 等级状态常量 ===
const (
	LevelStatusDisabled = 0 // 禁用
	LevelStatusEnabled  = 1 // 启用
)

// Level 合伙人等级模型（distribution_levels 表）
type Level struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 等级定义 ===
	Level          int     `gorm:"uniqueIndex;not null" json:"level"`                 // 等级值 1/2/3
	Name           string  `gorm:"size:64;not null" json:"name"`                      // 等级名称
	RequiredAmount float64 `gorm:"type:decimal(12,2);default:0" json:"required_amount"` // 升级所需累计佣金
	CommissionRate float64 `gorm:"type:decimal(5,4);default:0" json:"commission_rate"`  // 默认佣金比例

	// === 扩展 ===
	ExtraBenefits JSONB `gorm:"type:jsonb" json:"extra_benefits"` // 额外权益 JSONB
	Status        int   `gorm:"default:1;index" json:"status"`    // 0禁用 1启用
}

// TableName 表名
func (Level) TableName() string { return "distribution_levels" }
