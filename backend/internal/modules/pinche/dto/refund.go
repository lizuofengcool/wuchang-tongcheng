// Package dto 同城拼车出行数据传输对象 - 退款（基于 pinche_refunds 表）
package dto

import (
	"time"
)

// RefundInfo 退款详情
type RefundInfo struct {
	ID            uint       `json:"id"`
	RegionID      uint       `json:"region_id"`
	PaymentID     uint       `json:"payment_id"`
	BookingID     uint       `json:"booking_id"`
	PincheID      uint       `json:"pinche_id"`
	OrderID       uint       `json:"order_id"`
	RefundNo      string     `json:"refund_no"`
	RefundAmount  float64    `json:"refund_amount"`
	RefundReason  string     `json:"reason"`
	RefundType    string     `json:"refund_type"`
	RefundStatus  int        `json:"status"`
	StatusText    string     `json:"status_text"`
	RefundMethod  string     `json:"refund_method"`
	RefundedAt    *time.Time `json:"refunded_at"`
	ThirdPartyNo  string     `json:"third_party_no"`
	OperatorID    *uint      `json:"operator_id"`
	HandlerID     *uint      `json:"handler_id"`
	HandledAt     *time.Time `json:"handled_at"`
	PaidAt        *time.Time `json:"paid_at"`
	HandleNote    string     `json:"handle_note"`
	RejectReason  string     `json:"reject_reason"`
	ApplicantID   uint       `json:"applicant_id"`
	ApplicantName string     `json:"applicant_name"`
	ApplicantPhone string    `json:"applicant_phone"`
	PayeeAccount  string     `json:"payee_account"`
	OriginalAmount float64   `json:"original_amount"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ProcessRefundRequest 处理退款请求
type ProcessRefundRequest struct {
	Action        string  `json:"action" binding:"omitempty,oneof=approve reject mark_paid"`
	RefundMethod  string  `json:"refund_method"`
	ActualAmount  float64 `json:"actual_amount"`
	HandleNote    string  `json:"handle_note" binding:"max=500"`
	RejectReason  string  `json:"reject_reason"`
	Status        int     `json:"status"`
	PaidAt        *time.Time `json:"paid_at"`
}

// RefundListRequest 退款列表查询请求
type RefundListRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	RefundNo   string `form:"refund_no" json:"refund_no"`
	OrderID    string `form:"order_id" json:"order_id"`
	BookingID  string `form:"booking_id" json:"booking_id"`
	RefundType string `form:"refund_type" json:"refund_type"`
	Status     *int   `form:"status" json:"status"`
	StartDate  string `form:"start_date" json:"start_date"`
	EndDate    string `form:"end_date" json:"end_date"`
}

// RefundStatsResponse 退款统计响应
type RefundStatsResponse struct {
	Total       int64   `json:"total"`
	TotalAmount float64 `json:"total_amount"`
	Pending     int64   `json:"pending"`
	Processing  int64   `json:"processing"`
	Success     int64   `json:"success"`
	Rejected    int64   `json:"rejected"`
}

// RefundListResult 退款列表响应（带统计）
type RefundListResult struct {
	List     []RefundInfo        `json:"list"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Stats    RefundStatsResponse `json:"stats"`
}
