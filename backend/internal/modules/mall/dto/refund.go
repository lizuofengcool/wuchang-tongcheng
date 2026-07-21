// Package dto 同城商城 - 退款 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RefundInfo 退款详情响应
type RefundInfo struct {
	ID             uint       `json:"id"`
	OrderID        uint       `json:"order_id"`
	OrderNo        string     `json:"order_no"`
	OrderItemID    uint       `json:"order_item_id"`
	UserID         uint       `json:"user_id"`
	ShopID         uint       `json:"shop_id"`
	PaymentID      uint       `json:"payment_id"`
	RefundNo       string     `json:"refund_no"`
	TradeNo        string     `json:"trade_no"`
	Amount         float64    `json:"amount"`
	RefundAmount   float64    `json:"refund_amount"`
	RefundType     string     `json:"refund_type"`
	Reason         string     `json:"reason"`
	ReasonCode     string     `json:"reason_code"`
	Description    string     `json:"description"`
	EvidenceImages interface{} `json:"evidence_images"`
	ExpressCompany string     `json:"express_company"`
	ExpressNo      string     `json:"express_no"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	SellerRemark   string     `json:"seller_remark"`
	AdminRemark    string     `json:"admin_remark"`
	HandlerID      uint       `json:"handler_id"`
	HandlerName    string     `json:"handler_name"`
	ApprovedAt     *time.Time `json:"approved_at"`
	RejectedAt     *time.Time `json:"rejected_at"`
	RefundedAt     *time.Time `json:"refunded_at"`
	ClosedAt       *time.Time `json:"closed_at"`
	ShippedAt      *time.Time `json:"shipped_at"`
	ReceivedAt     *time.Time `json:"received_at"`
	RawResponse    interface{} `json:"raw_response"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateRefundRequest 创建退款请求
type CreateRefundRequest struct {
	OrderID        uint   `json:"order_id" binding:"required"`
	OrderItemID    uint   `json:"order_item_id"`
	RefundType     string `json:"refund_type" binding:"required,oneof=refund refund_return"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Reason         string `json:"reason" binding:"required,max=200"`
	ReasonCode     string `json:"reason_code" binding:"max=64"`
	Description    string `json:"description" binding:"max=500"`
	EvidenceImages interface{} `json:"evidence_images"`
}

// SellerProcessRefundRequest 卖家处理退款请求
type SellerProcessRefundRequest struct {
	Status       int    `json:"status" binding:"required,oneof=1 2 3"` // 1同意 2拒绝 3协商中
	SellerRemark string `json:"seller_remark" binding:"max=500"`
}

// AdminProcessRefundRequest 管理后台处理退款请求
type AdminProcessRefundRequest struct {
	Status      int    `json:"status" binding:"required,oneof=1 2 3 4"` // 1同意退款 2拒绝 3协商 4强制退款
	AdminRemark string `json:"admin_remark" binding:"max=500"`
}

// RefundShipRequest 退款退货物流请求（买家填退货单号）
type RefundShipRequest struct {
	ExpressCompany string `json:"express_company" binding:"required,max=64"`
	ExpressNo      string `json:"express_no" binding:"required,max=64"`
}

// RefundListRequest 退款列表请求
type RefundListRequest struct {
	utils.Pagination
	OrderID   uint   `form:"order_id" json:"order_id"`
	OrderNo   string `form:"order_no" json:"order_no"`
	UserID    uint   `form:"user_id" json:"user_id"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	RefundNo  string `form:"refund_no" json:"refund_no"`
	Status    *int   `form:"status" json:"status"`
	RefundType string `form:"refund_type" json:"refund_type"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	Keyword   string `form:"keyword" json:"keyword"`
	RegionID  uint   `form:"region_id" json:"region_id"`
}

// RefundStats 退款统计
type RefundStats struct {
	TotalCount       int64   `json:"total_count"`
	TotalAmount      float64 `json:"total_amount"`
	PendingCount     int64   `json:"pending_count"`
	ApprovedCount    int64   `json:"approved_count"`
	RejectedCount    int64   `json:"rejected_count"`
	RefundedCount    int64   `json:"refunded_count"`
	RefundedAmount   float64 `json:"refunded_amount"`
	ReturningCount   int64   `json:"returning_count"`
}
