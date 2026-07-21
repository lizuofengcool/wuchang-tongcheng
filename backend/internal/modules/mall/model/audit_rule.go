// Package model 同城商城 - 审核规则表
// 敏感词/违禁内容/联系方式/价格校验/频率（全局表，BaseModel 无 region_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// AuditRule 审核规则表（全局，BaseModel 无 region_id）
type AuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                          // 规则名
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                     // sensitive_word/prohibited/contact/price_check/frequency
	RuleKey     string `gorm:"size:64;not null;default:'';index" json:"rule_key"`           // 规则键
	Pattern     string `gorm:"type:text" json:"pattern"`                                    // 匹配模式（关键词/正则）
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                                 // 阈值 JSON
	Action      string `gorm:"size:32;not null;default:'reject';index" json:"action"`       // reject/approval/filter/limit
	PenaltyType string `gorm:"size:32;not null;default:''" json:"penalty_type"`            // warning/ban24h/ban7d/ban30d/ban_forever/delete/limit
	Severity    int    `gorm:"not null;default:1;index" json:"severity"`                    // 严重程度 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                              // 0禁用 1启用
	Description string `gorm:"size:500;not null;default:''" json:"description"`            // 描述
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                       // 排序
}

// TableName 表名
func (AuditRule) TableName() string { return "mall_audit_rules" }
