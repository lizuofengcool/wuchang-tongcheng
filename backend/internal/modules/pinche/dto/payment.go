// Package dto 同城拼车出行数据传输对象 - 支付
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentInfo 支付详情响应
type PaymentInfo struct {
	ID            uint       `json:"id"`
	RegionID      uint       `json:"region_id"`
	PincheID      uint       `json:"pinche_id"`
	BookingID     uint       `json:"booking_id"`
	PaymentNo     string     `json:"payment_no"`
	PayerID       uint       `json:"payer_id"`
	PayerName     string     `json:"payer_name"`
	PayeeID       uint       `json:"payee_id"`
	PayeeName     string     `json:"payee_name"`
	Amount        float64    `json:"amount"`
	InsuranceFee  float64    `json:"insurance_fee"`
	ServiceFee    float64    `json:"service_fee"`
	TotalAmount   float64    `json:"total_amount"`
	PaymentMethod string     `json:"payment_method"`
	PaymentStatus int        `json:"payment_status"`
	PaymentStatusText string  `json:"payment_status_text"`
	PaidAt        *time.Time `json:"paid_at"`
	RefundedAt    *time.Time `json:"refunded_at"`
	ThirdPartyNo  string     `json:"third_party_no"`
	RefundAmount  float64    `json:"refund_amount"`
	ETCLaneID     string     `json:"etc_lane_id"`
	ETCEntryTime  *time.Time `json:"etc_entry_time"`
	ETCExitTime   *time.Time `json:"etc_exit_time"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	BookingID     uint    `json:"booking_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=cash wechat alipay balance etc"`
	ETCLaneID     string  `json:"etc_lane_id"`
}

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PaymentNo     string `json:"payment_no" binding:"required"`
	ThirdPartyNo  string `json:"third_party_no"`
	PaymentStatus int    `json:"payment_status" binding:"oneof=1 3"`
	PaidAt        *time.Time `json:"paid_at"`
}

// PaymentListRequest 支付列表查询请求
type PaymentListRequest struct {
	PincheID      uint   `form:"pinche_id" json:"pinche_id"`
	BookingID     uint   `form:"booking_id" json:"booking_id"`
	PayerID       uint   `form:"payer_id" json:"payer_id"`
	PayeeID       uint   `form:"payee_id" json:"payee_id"`
	PaymentMethod string `form:"payment_method" json:"payment_method"`
	Status        *int   `form:"status" json:"status"`
	PaymentNo     string `form:"payment_no" json:"payment_no"`
	utils.Pagination
}

// ETCSettlementRequest ETC 结算请求
type ETCSettlementRequest struct {
	PaymentID   uint       `json:"payment_id" binding:"required"`
	ETCLaneID   string     `json:"etc_lane_id" binding:"required"`
	EntryTime   *time.Time `json:"entry_time"`
	ExitTime    *time.Time `json:"exit_time"`
	Amount      float64    `json:"amount" binding:"min=0"`
}
