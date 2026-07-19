// Package dto 同城车辆买卖数据传输对象 - 审核规则
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleInfo 审核规则详情响应
type AuditRuleInfo struct {
	ID          uint      `json:"id"`
	RuleName    string    `json:"rule_name"`
	RuleType    string    `json:"rule_type"`
	RuleKey     string    `json:"rule_key"`
	Pattern     string    `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string    `json:"action"`
	PenaltyType string    `json:"penalty_type"`
	Severity    int       `json:"severity"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	Description string    `json:"description"`
	Sort        int       `json:"sort"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// CreateAuditRuleRequest 创建审核规则请求
type CreateAuditRuleRequest struct {
	RuleName    string      `json:"rule_name" binding:"required,max=128"`
	RuleType    string      `json:"rule_type" binding:"required,oneof=sensitive_word price_check frequency fake_car mileage_check vin_check prohibited contact"`
	RuleKey     string      `json:"rule_key" binding:"max=64"`
	Pattern     string      `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      string      `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType string      `json:"penalty_type" binding:"omitempty,oneof=warning ban24h ban7d ban30d ban_forever delete_car limit"`
	Severity    int         `json:"severity" binding:"omitempty,min=1,max=5"`
	Status      int         `json:"status" binding:"omitempty,oneof=0 1"`
	Description string      `json:"description" binding:"max=500"`
	Sort        int         `json:"sort"`
}

// UpdateAuditRuleRequest 更新审核规则请求
type UpdateAuditRuleRequest struct {
	RuleName    *string     `json:"rule_name" binding:"omitempty,max=128"`
	RuleType    *string     `json:"rule_type" binding:"omitempty,oneof=sensitive_word price_check frequency fake_car mileage_check vin_check prohibited contact"`
	RuleKey     *string     `json:"rule_key" binding:"omitempty,max=64"`
	Pattern     *string     `json:"pattern"`
	Threshold   interface{} `json:"threshold"`
	Action      *string     `json:"action" binding:"omitempty,oneof=reject approval filter limit"`
	PenaltyType *string     `json:"penalty_type" binding:"omitempty,oneof=warning ban24h ban7d ban30d ban_forever delete_car limit"`
	Severity    *int        `json:"severity" binding:"omitempty,min=1,max=5"`
	Status      *int        `json:"status" binding:"omitempty,oneof=0 1"`
	Description *string     `json:"description" binding:"omitempty,max=500"`
	Sort        *int        `json:"sort"`
}

// AuditRuleListRequest 审核规则列表请求
type AuditRuleListRequest struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	RuleKey  string `form:"rule_key" json:"rule_key"`
	Action   string `form:"action" json:"action"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// AuditCheckRequest 审核检查请求（内部调用）
type AuditCheckRequest struct {
	Type    string      `json:"type" binding:"required"`
	Content string      `json:"content"`
	Data    interface{} `json:"data"`
}

// AuditCheckResponse 审核检查响应
type AuditCheckResponse struct {
	Passed  bool              `json:"passed"`
	Action  string            `json:"action"`
	Matched []AuditMatchedItem `json:"matched,omitempty"`
	Reason  string            `json:"reason"`
}

// AuditMatchedItem 命中规则项
type AuditMatchedItem struct {
	RuleID   uint   `json:"rule_id"`
	RuleName string `json:"rule_name"`
	RuleType string `json:"rule_type"`
	Pattern  string `json:"pattern"`
	Severity int    `json:"severity"`
}
