// Package dto 合同电子化 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ContractResponse 合同详情响应
type ContractResponse struct {
	ID                uint       `json:"id"`
	ContractNo        string     `json:"contract_no"`
	ContractType      string     `json:"contract_type"`
	HouseID           uint       `json:"house_id"`
	ListingID         uint       `json:"listing_id"`
	CommunityID       uint       `json:"community_id"`
	PartyAID          uint       `json:"party_a_id"`
	PartyAName        string     `json:"party_a_name"`
	PartyAPhone       string     `json:"party_a_phone"`
	PartyAIDCard      string     `json:"party_a_id_card"`
	PartyBID          uint       `json:"party_b_id"`
	PartyBName        string     `json:"party_b_name"`
	PartyBPhone       string     `json:"party_b_phone"`
	PartyBIDCard      string     `json:"party_b_id_card"`
	AgentID           uint       `json:"agent_id"`
	AgentName         string     `json:"agent_name"`
	Title             string     `json:"title"`
	Content           string     `json:"content"`
	Attachments       []model.ContractAttachment `json:"attachments"`
	Amount            float64    `json:"amount"`
	Deposit           float64    `json:"deposit"`
	Commission        float64    `json:"commission"`
	CommissionPayer   string     `json:"commission_payer"`
	StartDate         *time.Time `json:"start_date"`
	EndDate           *time.Time `json:"end_date"`
	PaymentMethod     string     `json:"payment_method"`
	SignMethod        string     `json:"sign_method"`
	PartyASignedAt    *time.Time `json:"party_a_signed_at"`
	PartyBSignedAt    *time.Time `json:"party_b_signed_at"`
	AgentSignedAt     *time.Time `json:"agent_signed_at"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	EffectiveAt       *time.Time `json:"effective_at"`
	TerminatedAt      *time.Time `json:"terminated_at"`
	TerminatedReason  string     `json:"terminated_reason"`
	ArchivedAt        *time.Time `json:"archived_at"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ContractCreateRequest 创建合同请求
type ContractCreateRequest struct {
	ContractType    string                   `json:"contract_type" binding:"required,oneof=rent sale deposit agency transfer"`
	HouseID         uint                     `json:"house_id" binding:"required"`
	ListingID       uint                     `json:"listing_id"`
	CommunityID     uint                     `json:"community_id"`
	PartyAID        uint                     `json:"party_a_id" binding:"required"`
	PartyAName      string                   `json:"party_a_name" binding:"required,max=50"`
	PartyAPhone     string                   `json:"party_a_phone" binding:"required,max=20"`
	PartyAIDCard    string                   `json:"party_a_id_card" binding:"max=32"`
	PartyBID        uint                     `json:"party_b_id" binding:"required"`
	PartyBName      string                   `json:"party_b_name" binding:"required,max=50"`
	PartyBPhone     string                   `json:"party_b_phone" binding:"required,max=20"`
	PartyBIDCard    string                   `json:"party_b_id_card" binding:"max=32"`
	AgentID         uint                     `json:"agent_id"`
	AgentName       string                   `json:"agent_name" binding:"max=50"`
	Title           string                   `json:"title" binding:"required,max=200"`
	Content         string                   `json:"content"`
	Attachments     []model.ContractAttachment `json:"attachments"`
	Amount          float64                  `json:"amount" binding:"gte=0"`
	Deposit         float64                  `json:"deposit" binding:"gte=0"`
	Commission      float64                  `json:"commission" binding:"gte=0"`
	CommissionPayer string                   `json:"commission_payer" binding:"omitempty,oneof=both buyer seller agent"`
	StartDate       *time.Time               `json:"start_date"`
	EndDate         *time.Time               `json:"end_date"`
	PaymentMethod   string                   `json:"payment_method" binding:"omitempty,oneof=monthly quarterly half_year yearly one_time"`
	SignMethod      string                   `json:"sign_method" binding:"omitempty,oneof=online offline mixed"`
	Status          int                      `json:"status" binding:"oneof=0 1"` // 0草稿 1待签
}

// ContractUpdateRequest 更新合同请求
type ContractUpdateRequest struct {
	Title           string                   `json:"title" binding:"max=200"`
	Content         string                   `json:"content"`
	Attachments     []model.ContractAttachment `json:"attachments"`
	Amount          float64                  `json:"amount" binding:"gte=0"`
	Deposit         float64                  `json:"deposit" binding:"gte=0"`
	Commission      float64                  `json:"commission" binding:"gte=0"`
	CommissionPayer string                   `json:"commission_payer"`
	StartDate       *time.Time               `json:"start_date"`
	EndDate         *time.Time               `json:"end_date"`
	PaymentMethod   string                   `json:"payment_method"`
	SignMethod      string                   `json:"sign_method"`
}

// ContractSignRequest 合同签署请求
type ContractSignRequest struct {
	Party string `json:"party" binding:"required,oneof=a b agent"` // 签署方
}

// ContractTerminateRequest 合同终止请求
type ContractTerminateRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// ContractListQuery 合同列表查询
type ContractListQuery struct {
	HouseID      uint   `form:"house_id" json:"house_id"`
	ListingID    uint   `form:"listing_id" json:"listing_id"`
	PartyAID     uint   `form:"party_a_id" json:"party_a_id"`
	PartyBID     uint   `form:"party_b_id" json:"party_b_id"`
	AgentID      uint   `form:"agent_id" json:"agent_id"`
	ContractType string `form:"contract_type" json:"contract_type"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ContractAdminListQuery 管理后台合同列表查询
type ContractAdminListQuery struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	HouseID  uint   `form:"house_id" json:"house_id"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}
