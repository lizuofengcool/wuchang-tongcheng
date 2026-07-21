// Package dto 同城商城 - 支付 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentInfo 支付详情响应
type PaymentInfo struct {
	ID           uint       `json:"id"`
	OrderID      uint       `json:"order_id"`
	OrderNo      string     `json:"order_no"`
	UserID       uint       `json:"user_id"`
	ShopID       uint       `json:"shop_id"`
	PaymentNo    string     `json:"payment_no"`
	TradeNo      string     `json:"trade_no"`
	OutTradeNo   string     `json:"out_trade_no"`
	Amount       float64    `json:"amount"`
	RefundAmount float64    `json:"refund_amount"`
	Method       string     `json:"method"`
	Channel      string     `json:"channel"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	ErrorCode    string     `json:"error_code"`
	ErrorMsg     string     `json:"error_msg"`
	PaidAt       *time.Time `json:"paid_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	RefundedAt   *time.Time `json:"refunded_at"`
	RawResponse  interface{} `json:"raw_response"`
	ClientIP     string     `json:"client_ip"`
	UserAgent    string     `json:"user_agent"`
	RegionID     uint       `json:"region_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID      uint   `json:"order_id" binding:"required"`
	OrderNo      string `json:"order_no" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=wechat alipay balance cod bankcard"`
	Channel      string `json:"channel" binding:"omitempty,oneof=mp h5 pc miniapp"`
	ReturnURL    string `json:"return_url"`
}

// PaymentCallbackRequest 支付回调请求（第三方回调）
type PaymentCallbackRequest struct {
	PaymentNo   string `json:"payment_no"`
	TradeNo     string `json:"trade_no"`
	OutTradeNo  string `json:"out_trade_no"`
	Status      int    `json:"status"`
	Amount      float64 `json:"amount"`
	ErrorCode   string `json:"error_code"`
	ErrorMsg    string `json:"error_msg"`
	RawResponse interface{} `json:"raw_response"`
	Sign        string `json:"sign"`
}

// PaymentListRequest 支付列表请求（管理后台）
type PaymentListRequest struct {
	utils.Pagination
	OrderID     uint   `form:"order_id" json:"order_id"`
	OrderNo     string `form:"order_no" json:"order_no"`
	UserID      uint   `form:"user_id" json:"user_id"`
	ShopID      uint   `form:"shop_id" json:"shop_id"`
	PaymentNo   string `form:"payment_no" json:"payment_no"`
	TradeNo     string `form:"trade_no" json:"trade_no"`
	Method      string `form:"method" json:"method"`
	Status      *int   `form:"status" json:"status"`
	StartDate   string `form:"start_date" json:"start_date"`
	EndDate     string `form:"end_date" json:"end_date"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// ClosePaymentRequest 关闭支付单请求
type ClosePaymentRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

// PaymentStats 支付统计
type PaymentStats struct {
	TotalCount    int64   `json:"total_count"`
	TotalAmount   float64 `json:"total_amount"`
	SuccessCount  int64   `json:"success_count"`
	SuccessAmount float64 `json:"success_amount"`
	PendingCount  int64   `json:"pending_count"`
	FailedCount   int64   `json:"failed_count"`
	RefundedCount int64   `json:"refunded_count"`
	RefundAmount  float64 `json:"refund_amount"`
}
