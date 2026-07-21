// Package dto 同城商城 - 订单明细 DTO
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// OrderItemDetailInfo 订单明细详情响应（管理后台使用，含更多字段）
type OrderItemDetailInfo struct {
	ID             uint    `json:"id"`
	OrderID        uint    `json:"order_id"`
	OrderNo        string  `json:"order_no"`
	ProductID      uint    `json:"product_id"`
	SkuID          uint    `json:"sku_id"`
	ShopID         uint    `json:"shop_id"`
	ProductName    string  `json:"product_name"`
	MainImage      string  `json:"main_image"`
	SkuName        string  `json:"sku_name"`
	SkuSpecs       string  `json:"sku_specs"`
	SkuCode        string  `json:"sku_code"`
	Price          float64 `json:"price"`
	Quantity       int     `json:"quantity"`
	TotalAmount    float64 `json:"total_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	ShippingFee    float64 `json:"shipping_fee"`
	PayAmount      float64 `json:"pay_amount"`
	HasReview      bool    `json:"has_review"`
	ReviewID       uint    `json:"review_id"`
	RefundStatus   int     `json:"refund_status"`
	RefundID       uint    `json:"refund_id"`
	Status         int     `json:"status"`
	StatusText     string  `json:"status_text"`
	RegionID       uint    `json:"region_id"`
}

// OrderItemListRequest 订单明细列表请求（管理后台）
type OrderItemListRequest struct {
	utils.Pagination
	OrderID   uint   `form:"order_id" json:"order_id"`
	OrderNo   string `form:"order_no" json:"order_no"`
	ProductID uint   `form:"product_id" json:"product_id"`
	SkuID     uint   `form:"sku_id" json:"sku_id"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	Status    *int   `form:"status" json:"status"`
	Keyword   string `form:"keyword" json:"keyword"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	RegionID  uint   `form:"region_id" json:"region_id"`
}

// OrderItemReviewRequest 订单明细评价请求
type OrderItemReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"max=1000"`
	Images  interface{} `json:"images"`
	Video   string `json:"video"`
	Tags    interface{} `json:"tags"`
	IsAnonymous bool `json:"is_anonymous"`
}

// OrderItemRefundRequest 订单明细退款请求
type OrderItemRefundRequest struct {
	RefundType string `json:"refund_type" binding:"required,oneof=refund refund_return"`
	Reason     string `json:"reason" binding:"required,max=200"`
	ReasonCode string `json:"reason_code"`
	Description string `json:"description" binding:"max=500"`
	EvidenceImages interface{} `json:"evidence_images"`
	Amount     float64 `json:"amount"`
}
