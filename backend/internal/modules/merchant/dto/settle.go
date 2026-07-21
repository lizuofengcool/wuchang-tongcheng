// Package dto 商户中台数据传输对象 - 结算
// 对标美团/有赞商户结算系统
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// SettleInfo 结算单详情响应
type SettleInfo struct {
	ID           uint       `json:"id"`
	ShopID       uint       `json:"shop_id"`
	Period       string     `json:"period"`
	TotalAmount  float64    `json:"total_amount"`
	PlatformFee  float64    `json:"platform_fee"`
	ShopAmount   float64    `json:"shop_amount"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	SettledAt    *time.Time `json:"settled_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SettleListRequest 结算列表请求
type SettleListRequest struct {
	ShopID uint   `form:"shop_id" json:"shop_id"`
	Period string `form:"period" json:"period"`
	Status *int   `form:"status" json:"status"`
	utils.Pagination
}

// SettleGenerateRequest 生成结算单请求
type SettleGenerateRequest struct {
	ShopID       uint    `json:"shop_id" binding:"required"`
	Period       string `json:"period" binding:"required"` // YYYY-MM
	TotalAmount  float64 `json:"total_amount" binding:"required"`
	PlatformRate float64 `json:"platform_rate"` // 平台佣金比例 0-1
}

// SettleWithdrawRequest 提现申请请求
type SettleWithdrawRequest struct {
	ID uint `json:"id" binding:"required"`
}

// SettleAuditRequest 提现审核请求
type SettleAuditRequest struct {
	Status int    `json:"status" binding:"oneof=1 2 3"` // 1通过 2拒绝 3撤销
	Reason string `json:"reason" binding:"max=500"`
}

// SettleSummaryRequest 汇总查询请求
type SettleSummaryRequest struct {
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	Period    string `form:"period" json:"period"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
}

// SettleSummary 汇总结果
type SettleSummary struct {
	ShopID       uint    `json:"shop_id"`
	Period       string  `json:"period"`
	TotalAmount  float64 `json:"total_amount"`
	PlatformFee  float64 `json:"platform_fee"`
	ShopAmount   float64 `json:"shop_amount"`
	Count        int64   `json:"count"`
}
