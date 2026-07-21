// Package dto 同城商城 - 审核规则 DTO
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleInfo 审核规则详情响应
type AuditRuleInfo struct {
	ID          uint        `json:"id"`
	RuleName    string      `json:"rule_name"`
	RuleType    string      `json:"rule_type"`
	RuleKey     string      `json:"rule_key"`
	Pattern     string      `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action"`
	PenaltyType string      `json:"penalty_type"`
	Severity    int         `json:"severity"`
	Status      int         `json:"status"`
	StatusText  string      `json:"status_text"`
	Description string      `json:"description"`
	Sort        int         `json:"sort"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// CreateAuditRuleRequest 创建审核规则请求
type CreateAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"required,oneof=keyword image text behavior price stock"`
	RuleKey     string      `json:"rule_key" binding:"required,max=64"`
	Pattern     string      `json:"pattern" binding:"max=500"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"required,oneof=block review warning pass"`
	PenaltyType string      `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
	Severity    int         `json:"severity" binding:"omitempty,min=0,max=10"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
	Description string      `json:"description" binding:"max=500"`
	Sort        int         `json:"sort"`
}

// UpdateAuditRuleRequest 更新审核规则请求
type UpdateAuditRuleRequest struct {
	RuleName    *string      `json:"rule_name" binding:"max=128"`
	Pattern     *string      `json:"pattern" binding:"max=500"`
	Threshold   interface{}  `json:"threshold"`
	Action      *string      `json:"action" binding:"omitempty,oneof=block review warning pass"`
	PenaltyType *string      `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
	Severity    *int         `json:"severity" binding:"omitempty,min=0,max=10"`
	Status      *int         `json:"status" binding:"omitempty,oneof=0 1"`
	Description *string      `json:"description" binding:"max=500"`
	Sort        *int         `json:"sort"`
}

// AuditRuleListRequest 审核规则列表请求
type AuditRuleListRequest struct {
	utils.Pagination
	Keyword   string `form:"keyword" json:"keyword"`
	RuleType  string `form:"rule_type" json:"rule_type"`
	RuleKey   string `form:"rule_key" json:"rule_key"`
	Action    string `form:"action" json:"action"`
	Status    *int   `form:"status" json:"status"`
	Severity  *int   `form:"severity" json:"severity"`
}

// AuditCheckRequest 内容审核检测请求
type AuditCheckRequest struct {
	Type    string      `json:"type" binding:"required,oneof=product shop review order"`
	Content interface{} `json:"content"`
	UserID  uint        `json:"user_id"`
}

// AuditCheckResponse 内容审核检测结果
type AuditCheckResponse struct {
	Pass     bool              `json:"pass"`
	Action   string            `json:"action"`
	HitRules []AuditRuleInfo   `json:"hit_rules"`
	Score    int               `json:"score"`
}
