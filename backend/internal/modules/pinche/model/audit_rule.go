// Package model 拼车审核规则数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheAuditRule 审核规则
type PincheAuditRule struct {
	database.RegionBaseModel

	RuleName    string `gorm:"size:128" json:"rule_name"`
	RuleType    string `gorm:"size:32;index" json:"rule_type"`
	RuleCode    string `gorm:"size:64;index" json:"rule_code"`
	Description string `gorm:"size:500" json:"description"`

	Threshold JSONB `gorm:"type:jsonb" json:"threshold"`
	Action    string `gorm:"size:32;not null;default:'manual_review'" json:"action"` // pass/reject/manual_review
	Priority  int    `gorm:"not null;default:0;index" json:"priority"`

	Status    int        `gorm:"not null;default:1;index" json:"status"` // 0禁用 1启用
	HitCount  int        `gorm:"not null;default:0" json:"hit_count"`
	LastHitAt *time.Time `gorm:"index" json:"last_hit_at"`
}

// TableName 表名
func (PincheAuditRule) TableName() string { return "pinche_audit_rules" }
