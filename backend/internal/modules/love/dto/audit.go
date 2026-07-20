// Package dto love 相亲交友数据传输对象 - 审核规则
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveAuditRuleInfo 审核规则详情响应
type LoveAuditRuleInfo struct {
	ID          uint      `json:"id"`
	RuleName    string    `json:"rule_name"`
	RuleType    string    `json:"rule_type"`
	RuleKey     string    `json:"rule_key"`
	Pattern     string    `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string    `json:"action"`
	ActionText  string    `json:"action_text"`
	PenaltyType string    `json:"penalty_type"`
	Severity    int       `json:"severity"`
	SeverityText string   `json:"severity_text"`
	Scope       string    `json:"scope"`
	TargetType  string    `json:"target_type"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// CreateLoveAuditRuleRequest 创建审核规则请求
type CreateLoveAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"required,oneof=sensitive_word prohibited contact fake_info frequency porn political age_check photo_check"`
	RuleKey     string      `json:"rule_key" binding:"max=64"`
	Pattern     string      `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"omitempty,oneof=reject review block shadow warning"`
	PenaltyType string      `json:"penalty_type" binding:"omitempty,oneof=warning freeze ban delete_content deduction none"`
	Severity    int         `json:"severity" binding:"omitempty,min=1,max=4"`
	Scope       string      `json:"scope" binding:"omitempty,oneof=all profile story message impression bio nickname"`
	TargetType  string      `json:"target_type" binding:"omitempty,oneof=all profile story message impression"`
	Description string      `json:"description" binding:"max=500"`
	Sort        int         `json:"sort"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateLoveAuditRuleRequest 更新审核规则请求
type UpdateLoveAuditRuleRequest struct {
	RuleName    *string     `json:"rule_name" binding:"omitempty,max=128"`
	RuleType    *string     `json:"rule_type" binding:"omitempty,oneof=sensitive_word prohibited contact fake_info frequency porn political age_check photo_check"`
	RuleKey     *string     `json:"rule_key" binding:"omitempty,max=64"`
	Pattern     *string     `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      *string     `json:"action" binding:"omitempty,oneof=reject review block shadow warning"`
	PenaltyType *string     `json:"penalty_type" binding:"omitempty,oneof=warning freeze ban delete_content deduction none"`
	Severity    *int        `json:"severity" binding:"omitempty,min=1,max=4"`
	Scope       *string     `json:"scope" binding:"omitempty,oneof=all profile story message impression bio nickname"`
	TargetType  *string     `json:"target_type" binding:"omitempty,oneof=all profile story message impression"`
	Description *string     `json:"description" binding:"omitempty,max=500"`
	Sort        *int        `json:"sort"`
	Status      *int        `json:"status" binding:"omitempty,oneof=0 1"`
}

// LoveAuditRuleListRequest 审核规则列表请求
type LoveAuditRuleListRequest struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	RuleKey  string `form:"rule_key" json:"rule_key"`
	Action   string `form:"action" json:"action"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// LoveAuditCheckRequest 审核检查请求
type LoveAuditCheckRequest struct {
	Type    string      `json:"type" binding:"required"`
	Content string      `json:"content"`
	Data    interface{} `json:"data"`
}

// LoveAuditCheckResponse 审核检查响应
type LoveAuditCheckResponse struct {
	Passed  bool                    `json:"passed"`
	Action  string                  `json:"action"`
	Matched []LoveAuditMatchedItem  `json:"matched,omitempty"`
	Reason  string                  `json:"reason"`
}

// LoveAuditMatchedItem 命中规则项
type LoveAuditMatchedItem struct {
	RuleID   uint   `json:"rule_id"`
	RuleName string `json:"rule_name"`
	RuleType string `json:"rule_type"`
	Pattern  string `json:"pattern"`
	Severity int    `json:"severity"`
}
