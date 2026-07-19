// Package dto 审核规则相关 DTO
// 依据 v3.2.1 架构方案：敏感词/薪资异常/频率限制等审核规则
package dto

import "wuchang-tongcheng/internal/pkg/utils"

// AuditRuleResponse 审核规则响应
type AuditRuleResponse struct {
	ID          uint                   `json:"id"`
	RuleName    string                 `json:"rule_name"`
	RuleType    string                 `json:"rule_type"`
	RuleKey     string                 `json:"rule_key"`
	Pattern     string                 `json:"pattern"`
	Threshold   map[string]interface{} `json:"threshold"`
	Action      string                 `json:"action"`
	PenaltyType string                 `json:"penalty_type"`
	Severity    int                    `json:"severity"`
	Status      int                    `json:"status"`
	StatusText  string                 `json:"status_text"`
	Description string                 `json:"description"`
	Sort        int                    `json:"sort"`
	CreatedAt   string                 `json:"created_at"`
}

// AuditRuleCreateRequest 创建审核规则请求
type AuditRuleCreateRequest struct {
	RuleName    string                 `json:"rule_name" binding:"required,max=128"`
	RuleType    string                 `json:"rule_type" binding:"required,oneof=sensitive_word salary_check frequency fake_recruit prohibited content company resume"`
	RuleKey     string                 `json:"rule_key" binding:"max=64"`
	Pattern     string                 `json:"pattern"`
	Threshold   map[string]interface{} `json:"threshold"`
	Action      string                 `json:"action" binding:"omitempty,oneof=reject approval warning manual"`
	PenaltyType string                 `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d ban30d ban_forever close_job freeze_company"`
	Severity    int                    `json:"severity" binding:"gte=1,lte=5"`
	Status      int                    `json:"status" binding:"omitempty,oneof=0 1"`
	Description string                 `json:"description" binding:"max=500"`
	Sort        int                    `json:"sort"`
}

// AuditRuleUpdateRequest 更新审核规则请求
type AuditRuleUpdateRequest struct {
	RuleName    *string                 `json:"rule_name" binding:"omitempty,max=128"`
	RuleType    *string                 `json:"rule_type" binding:"omitempty,oneof=sensitive_word salary_check frequency fake_recruit prohibited content company resume"`
	RuleKey     *string                 `json:"rule_key" binding:"omitempty,max=64"`
	Pattern     *string                 `json:"pattern"`
	Threshold   map[string]interface{}  `json:"threshold"`
	Action      *string                 `json:"action" binding:"omitempty,oneof=reject approval warning manual"`
	PenaltyType *string                 `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d ban30d ban_forever close_job freeze_company"`
	Severity    *int                    `json:"severity" binding:"omitempty,gte=1,lte=5"`
	Status      *int                    `json:"status" binding:"omitempty,oneof=0 1"`
	Description *string                 `json:"description" binding:"omitempty,max=500"`
	Sort        *int                    `json:"sort"`
}

// AuditRuleListQuery 审核规则列表查询
type AuditRuleListQuery struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	Status   *int   `form:"status" json:"status"`
	utils.Pagination
}
