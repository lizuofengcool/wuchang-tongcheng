// Package dto 分销合伙人中台 - 渠道请求/响应 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ChannelInfo 渠道详情响应
type ChannelInfo struct {
	ID                uint      `json:"id"`
	PartnerID         uint      `json:"partner_id"`
	Code              string    `json:"code"`
	Name              string    `json:"name"`
	ClickCount        int       `json:"click_count"`
	RegisterCount     int       `json:"register_count"`
	OrderCount        int       `json:"order_count"`
	CommissionAmount  float64   `json:"commission_amount"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ChannelCreateRequest 创建渠道请求
type ChannelCreateRequest struct {
	Name string `json:"name" binding:"required,max=100"`
	Code string `json:"code" binding:"omitempty,max=50"` // 不填则自动生成
}

// ChannelUpdateRequest 更新渠道请求
type ChannelUpdateRequest struct {
	Name *string `json:"name" binding:"omitempty,max=100"`
}

// ChannelListRequest 渠道列表请求
type ChannelListRequest struct {
	PartnerID uint   `form:"partner_id" json:"partner_id"`
	Code      string `form:"code" json:"code"`
	Keyword   string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ChannelStatsRequest 渠道统计请求
type ChannelStatsRequest struct {
	PartnerID uint `form:"partner_id" json:"partner_id"`
}

// ChannelStatsResponse 渠道统计响应
type ChannelStatsResponse struct {
	TotalChannels     int     `json:"total_channels"`
	TotalClicks       int     `json:"total_clicks"`
	TotalRegisters    int     `json:"total_registers"`
	TotalOrders       int     `json:"total_orders"`
	TotalCommission   float64 `json:"total_commission"`
}

// ChannelTrackRequest 渠道追踪请求（公开，记录点击/注册）
type ChannelTrackRequest struct {
	Code   string `json:"code" binding:"required"`         // 渠道码
	Action string `json:"action" binding:"required,oneof=click register order"` // 动作类型
}
