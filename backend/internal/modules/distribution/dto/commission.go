// Package dto 分销合伙人中台 - 佣金请求/响应 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CommissionInfo 佣金记录详情响应
type CommissionInfo struct {
	ID                uint       `json:"id"`
	PartnerID         uint       `json:"partner_id"`
	OrderID           uint       `json:"order_id"`
	ChannelID         uint       `json:"channel_id"`
	OrderAmount       float64    `json:"order_amount"`
	CommissionAmount  float64    `json:"commission_amount"`
	CommissionRate    float64    `json:"commission_rate"`
	Level             int        `json:"level"`
	LevelText         string     `json:"level_text"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	SettledAt         *time.Time `json:"settled_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CommissionCreateRequest 创建佣金记录（订单结算时由系统/管理端触发）
type CommissionCreateRequest struct {
	PartnerID  uint    `json:"partner_id" binding:"required"`
	OrderID    uint    `json:"order_id" binding:"required"`
	ChannelID  uint    `json:"channel_id"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0"`
	Level      int     `json:"level" binding:"omitempty,oneof=1 2"`
}

// CommissionListRequest 佣金列表请求
type CommissionListRequest struct {
	PartnerID uint   `form:"partner_id" json:"partner_id"`
	OrderID   uint   `form:"order_id" json:"order_id"`
	Level     *int   `form:"level" json:"level"`
	Status    *int   `form:"status" json:"status"`
	utils.Pagination
}

// CommissionSummaryRequest 佣金汇总请求
type CommissionSummaryRequest struct {
	PartnerID uint `form:"partner_id" json:"partner_id"`
}

// CommissionSummaryResponse 佣金汇总响应
type CommissionSummaryResponse struct {
	PartnerID         uint    `json:"partner_id"`
	TotalCommission   float64 `json:"total_commission"`
	SettledCommission float64 `json:"settled_commission"`
	PendingCommission float64 `json:"pending_commission"`
	CanceledCommission float64 `json:"canceled_commission"`
	TotalCount        int64   `json:"total_count"`
}

// CommissionSettleRequest 批量结算请求
type CommissionSettleRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"` // 佣金记录 ID 列表
}

// CommissionSettleResult 批量结算结果
type CommissionSettleResult struct {
	Total   int   `json:"total"`
	Success int   `json:"success"`
	Failed  int   `json:"failed"`
	FailedIDs []uint `json:"failed_ids"`
}
