// Package model 营销活动中台 - 签到规则模型（sign 子域）
// 依据架构设计 4.6：连续签到奖励配置
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 签到规则状态常量 ===
const (
	SignRuleStatusDisabled = 0 // 禁用
	SignRuleStatusEnabled  = 1 // 启用
)

// SignRule 签到规则模型（sign_rules 表）
// 按连续签到天数配置奖励积分与额外奖励
type SignRule struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at

	Day         int    `gorm:"not null;uniqueIndex:uniq_sign_rules_day" json:"day"` // 连续签到第 N 天
	Points      int    `gorm:"not null;default:0" json:"points"`                    // 奖励积分
	ExtraReward JSONB  `gorm:"type:jsonb" json:"extra_reward"`                      // 额外奖励（JSONB，如优惠券/积分倍数）
	Status      int    `gorm:"not null;default:1;index" json:"status"`              // 0禁用 1启用
}

// TableName 表名
func (SignRule) TableName() string { return "sign_rules" }
