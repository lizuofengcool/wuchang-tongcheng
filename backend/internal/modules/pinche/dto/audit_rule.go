// Package dto 同城拼车出行数据传输对象 - 审核规则
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleInfo 审核规则详情响应
type AuditRuleInfo struct {
	ID          uint       `json:"id"`
	RegionID    uint       `json:"region_id"`
	RuleName    string     `json:"rule_name"`
	RuleType    string     `json:"rule_type"`
	RuleCode    string     `json:"rule_code"`
	Description string     `json:"description"`
	Threshold   interface{} `json:"threshold"`
	Action      string     `json:"action"`
	ActionText  string     `json:"action_text"`
	Priority    int        `json:"priority"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	HitCount    int        `json:"hit_count"`
	LastHitAt   *time.Time `json:"last_hit_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AuditRuleListRequest 审核规则列表查询请求
type AuditRuleListRequest struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	RuleCode string `form:"rule_code" json:"rule_code"`
	Action   string `form:"action" json:"action"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CreateAuditRuleRequest 创建审核规则请求
type CreateAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"required,max=32"`
	RuleCode    string      `json:"rule_code" binding:"max=64"`
	Description string      `json:"description" binding:"max=500"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"omitempty,oneof=pass reject manual_review"`
	Priority    int         `json:"priority" binding:"omitempty,min=0"`
}

// UpdateAuditRuleRequest 更新审核规则请求
type UpdateAuditRuleRequest struct {
	RuleName    *string     `json:"rule_name" binding:"omitempty,max=128"`
	RuleType    *string     `json:"rule_type" binding:"omitempty,max=32"`
	Description *string     `json:"description" binding:"omitempty,max=500"`
	Threshold   interface{} `json:"threshold"`
	Action      *string     `json:"action" binding:"omitempty,oneof=pass reject manual_review"`
	Priority    *int        `json:"priority" binding:"omitempty,min=0"`
	Status      *int        `json:"status" binding:"omitempty,oneof=0 1"`
}
