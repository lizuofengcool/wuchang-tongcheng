// Package model 审核规则表（全局，BaseModel 无 region_id）
// 提供 M 端规则管理 + 内部审核检查能力（Check 方法供其他 service 调用）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 审核规则类型常量 ===
const (
	LinggongAuditRuleTypeSensitiveWord = "sensitive_word" // 敏感词
	LinggongAuditRuleTypeSalaryCheck    = "salary_check"  // 薪资异常
	LinggongAuditRuleTypeFrequency      = "frequency"     // 频率限制
	LinggongAuditRuleTypeFakeJob        = "fake_job"       // 虚假岗位
	LinggongAuditRuleTypeContact        = "contact"        // 联系方式校验
	LinggongAuditRuleTypeProhibited     = "prohibited"    // 违禁内容
	LinggongAuditRuleTypeDuplicate      = "duplicate"      // 重复发布
	LinggongAuditRuleTypeBlacklist      = "blacklist"      // 黑名单
)

// === 审核动作常量 ===
const (
	LinggongAuditActionReject   = "reject"   // 拒绝
	LinggongAuditActionApproval = "approval" // 转人工
	LinggongAuditActionFilter   = "filter"   // 过滤敏感词
	LinggongAuditActionLimit   = "limit"     // 限流
)

// LinggongAuditRule 审核规则表
type LinggongAuditRule struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RuleName    string `gorm:"size:128;not null" json:"rule_name"`                              // 规则名
	RuleType    string `gorm:"size:32;not null;index" json:"rule_type"`                         // sensitive_word/salary_check/frequency/fake_job/contact/prohibited/duplicate/blacklist
	RuleKey     string `gorm:"size:64;not null;default:'';index" json:"rule_key"`               // 规则键
	Pattern     string `gorm:"type:text" json:"pattern"`                                        // 正则/关键词模式
	Threshold   JSONB  `gorm:"type:jsonb" json:"threshold"`                                     // 阈值 JSON
	Action      string `gorm:"size:32;not null;default:'reject'" json:"action"`                 // reject/approval/filter/limit
	PenaltyType string `gorm:"size:32;not null;default:''" json:"penalty_type"`                // 处罚类型
	Severity    int    `gorm:"not null;default:1;index" json:"severity"`                        // 严重等级 1-5
	Status      int    `gorm:"default:1;index" json:"status"`                                   // 0禁用 1启用
	Description string `gorm:"size:500;not null;default:''" json:"description"`                 // 描述
	Sort        int    `gorm:"not null;default:0" json:"sort"`                                  // 排序
}

// TableName 表名（linggong_ 前缀）
func (LinggongAuditRule) TableName() string { return "linggong_audit_rules" }
