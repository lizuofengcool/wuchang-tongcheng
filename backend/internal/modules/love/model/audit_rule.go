// Package model love 相亲交友数据模型 - 审核规则表 LoveAuditRule
// 对标陌陌/探探：敏感词/违规内容/频率限制
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 注：审核规则为全局配置，使用 BaseModel（无 region_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 审核范围 ===
const (
	AuditScopeAll       = "all"       // 全部
	AuditScopeProfile   = "profile"   // 资料
	AuditScopeStory     = "story"     // 动态
	AuditScopeMessage   = "message"   // 消息
	AuditScopeImpression = "impression" // 印象
	AuditScopeBio       = "bio"       // 简介
	AuditScopeNickname  = "nickname"  // 昵称
)

// === 严重程度 ===
const (
	AuditSeverityLow      = 1 // 低
	AuditSeverityMedium   = 2 // 中
	AuditSeverityHigh     = 3 // 高
	AuditSeverityCritical = 4 // 严重
)

// LoveAuditRule 审核规则表
// 全局配置表，使用 BaseModel（无 region_id）
type LoveAuditRule struct {
	database.BaseModel

	RuleName string `gorm:"size:128;not null" json:"rule_name"`     // 规则名称
	RuleType string `gorm:"size:32;not null;index" json:"rule_type"` // 规则类型（参考 AuditRuleType*）
	RuleKey  string `gorm:"size:64;not null;uniqueIndex" json:"rule_key"` // 规则键（唯一标识）

	Pattern   string `gorm:"type:text" json:"pattern"`     // 匹配模式（正则/关键词列表，按行分隔）
	Threshold JSONB  `gorm:"type:jsonb" json:"threshold"`   // 阈值配置（频率/次数等）

	Action      string `gorm:"size:32;not null;default:'reject';index" json:"action"`       // 动作（参考 AuditAction*）
	PenaltyType string `gorm:"size:32;not null;default:''" json:"penalty_type"`             // 处罚类型
	Severity    int    `gorm:"not null;default:1;index" json:"severity"`                    // 严重程度 1-4

	Scope      string `gorm:"size:32;not null;default:'all'" json:"scope"`                  // 适用范围
	TargetType string `gorm:"size:32;not null;default:'all'" json:"target_type"`            // 目标类型

	Description string `gorm:"size:500;not null;default:''" json:"description"`

	Sort   int `gorm:"not null;default:0;index" json:"sort"`
	Status int `gorm:"not null;default:1;index" json:"status"` // 0禁用 1启用
}

// TableName 表名
func (LoveAuditRule) TableName() string { return "love_audit_rules" }
