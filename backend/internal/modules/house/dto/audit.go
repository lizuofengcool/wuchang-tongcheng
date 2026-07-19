// Package dto 审核规则 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AuditRuleResponse 审核规则响应
type AuditRuleResponse struct {
	ID          uint       `json:"id"`
	RuleName    string     `json:"rule_name"`
	RuleType    string     `json:"rule_type"`
	RuleTypeText string    `json:"rule_type_text"`
	RuleKey     string     `json:"rule_key"`
	Pattern     string     `json:"pattern"`
	Threshold   []model.AuditRuleThreshold `json:"threshold"`
	Action      string     `json:"action"`
	ActionText  string     `json:"action_text"`
	PenaltyType string     `json:"penalty_type"`
	Severity    int        `json:"severity"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Description string     `json:"description"`
	Sort        int        `json:"sort"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AuditRuleCreateRequest 创建/更新审核规则请求
type AuditRuleCreateRequest struct {
	RuleName    string                       `json:"rule_name" binding:"required,max=128"`
	RuleType    string                       `json:"rule_type" binding:"required,oneof=sensitive_word price_check frequency fake_house prohibited contact"`
	RuleKey     string                       `json:"rule_key" binding:"max=64"`
	Pattern     string                       `json:"pattern"`
	Threshold   []model.AuditRuleThreshold   `json:"threshold"`
	Action      string                       `json:"action" binding:"omitempty,oneof=reject approval limit warning"`
	PenaltyType string                       `json:"penalty_type" binding:"omitempty,oneof=warning remove ban7d ban30d ban_forever"`
	Severity    int                          `json:"severity" binding:"gte=1,lte=5"`
	Status      int                          `json:"status" binding:"omitempty,oneof=0 1"`
	Description string                       `json:"description" binding:"max=500"`
	Sort        int                          `json:"sort"`
}

// AuditRuleListQuery 审核规则列表查询
type AuditRuleListQuery struct {
	RuleType string `form:"rule_type" json:"rule_type"`
	RuleKey  string `form:"rule_key" json:"rule_key"`
	Action   string `form:"action" json:"action"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ===== 房源分类 + 配套设施 =====

// CategoryResponse 房源分类响应
type CategoryResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	ParentID      uint      `json:"parent_id"`
	Level         int       `json:"level"`
	ListingType   string    `json:"listing_type"`
	PropertyType  string    `json:"property_type"`
	Icon          string    `json:"icon"`
	Color         string    `json:"color"`
	Description   string    `json:"description"`
	Sort          int       `json:"sort"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	HouseCount    int       `json:"house_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Children      []CategoryResponse `json:"children,omitempty"` // 子分类（树形结构）
}

// CategoryCreateRequest 创建/更新分类请求
type CategoryCreateRequest struct {
	Name         string `json:"name" binding:"required,max=64"`
	Code         string `json:"code" binding:"required,max=64"`
	ParentID     uint   `json:"parent_id"`
	Level        int    `json:"level" binding:"gte=1,lte=3"`
	ListingType  string `json:"listing_type" binding:"omitempty,oneof=rent sale transfer"`
	PropertyType string `json:"property_type" binding:"omitempty,oneof=residential apartment villa loft office shop"`
	Icon         string `json:"icon" binding:"max=64"`
	Color        string `json:"color" binding:"max=16"`
	Description  string `json:"description" binding:"max=500"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// CategoryListQuery 分类列表查询
type CategoryListQuery struct {
	ParentID     uint   `form:"parent_id" json:"parent_id"`
	Level        int    `form:"level" json:"level"`
	ListingType  string `form:"listing_type" json:"listing_type"`
	PropertyType string `form:"property_type" json:"property_type"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// FacilityResponse 配套设施响应
type FacilityResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Category    string    `json:"category"`
	CategoryText string   `json:"category_text"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	Background  string    `json:"background"`
	Description string     `json:"description"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	Sort        int       `json:"sort"`
	UseCount    int       `json:"use_count"`
	IsHot       bool      `json:"is_hot"`
	CreatorID   uint      `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FacilityCreateRequest 创建/更新配套设施请求
type FacilityCreateRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	Code        string `json:"code" binding:"required,max=64"`
	Category    string `json:"category" binding:"omitempty,oneof=indoor outdoor appliance furniture network security parking"`
	Icon        string `json:"icon" binding:"max=64"`
	Color       string `json:"color" binding:"max=16"`
	Background  string `json:"background" binding:"max=32"`
	Description string `json:"description" binding:"max=500"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
	Sort        int    `json:"sort"`
	IsHot       bool   `json:"is_hot"`
}

// FacilityListQuery 配套设施列表查询
type FacilityListQuery struct {
	Category string `form:"category" json:"category"`
	Status   *int   `form:"status" json:"status"`
	IsHot    *bool  `form:"is_hot" json:"is_hot"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}
