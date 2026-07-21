// Package dto 同城商城 - 订单 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// OrderInfo 订单详情响应
type OrderInfo struct {
	ID            uint       `json:"id"`
	OrderNo       string     `json:"order_no"`
	UserID        uint       `json:"user_id"`
	ShopID        uint       `json:"shop_id"`
	ShopName      string     `json:"shop_name"`

	BuyerName     string     `json:"buyer_name"`
	BuyerPhone    string     `json:"buyer_phone"`
	BuyerAvatar   string     `json:"buyer_avatar"`

	AddressID     uint       `json:"address_id"`
	ReceiverName  string     `json:"receiver_name"`
	ReceiverPhone string     `json:"receiver_phone"`
	Province      string     `json:"province"`
	City          string     `json:"city"`
	District      string     `json:"district"`
	Address       string     `json:"address"`
	ZipCode       string     `json:"zip_code"`

	TotalAmount    float64 `json:"total_amount"`
	ShippingFee    float64 `json:"shipping_fee"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount      float64 `json:"tax_amount"`
	PayAmount      float64 `json:"pay_amount"`

	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	Remark        string     `json:"remark"`
	SellerRemark  string     `json:"seller_remark"`
	AdminRemark   string     `json:"admin_remark"`

	PaidAt       *time.Time `json:"paid_at"`
	ShippedAt    *time.Time `json:"shipped_at"`
	ReceivedAt   *time.Time `json:"received_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CancelledAt  *time.Time `json:"cancelled_at"`
	AutoCloseAt  *time.Time `json:"auto_close_at"`
	AutoConfirmAt *time.Time `json:"auto_confirm_at"`
	AutoReviewAt  *time.Time `json:"auto_review_at"`

	PaymentMethod string `json:"payment_method"`
	PaymentNo     string `json:"payment_no"`

	Source    string `json:"source"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`

	CouponID    uint   `json:"coupon_id"`
	CouponName  string `json:"coupon_name"`

	HasReview       bool `json:"has_review"`
	HasSellerReview bool `json:"has_seller_review"`

	LogisticsCompany string `json:"logistics_company"`
	LogisticsNo      string `json:"logistics_no"`

	RiskScore int `json:"risk_score"`

	Items []OrderItemInfo `json:"items,omitempty"` // 订单明细（详情时返回）

	RegionID  uint      `json:"region_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItemInfo 订单明细响应
type OrderItemInfo struct {
	ID            uint    `json:"id"`
	OrderID       uint    `json:"order_id"`
	ProductID     uint    `json:"product_id"`
	SkuID         uint    `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	MainImage     string  `json:"main_image"`
	SkuName       string  `json:"sku_name"`
	SkuSpecs      string  `json:"sku_specs"`
	SkuCode       string  `json:"sku_code"`
	Price         float64 `json:"price"`
	Quantity      int     `json:"quantity"`
	TotalAmount   float64 `json:"total_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	ShippingFee   float64 `json:"shipping_fee"`
	PayAmount     float64 `json:"pay_amount"`
	HasReview     bool    `json:"has_review"`
	ReviewID      uint    `json:"review_id"`
	RefundStatus  int     `json:"refund_status"`
	RefundID      uint    `json:"refund_id"`
	Status        int     `json:"status"`
	StatusText    string  `json:"status_text"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Items         []OrderItemRequest `json:"items" binding:"required,min=1"`
	AddressID     uint               `json:"address_id" binding:"required"`
	Remark        string             `json:"remark" binding:"max=500"`
	CouponID      uint               `json:"coupon_id"`
	PaymentMethod string             `json:"payment_method" binding:"omitempty,oneof=wechat alipay balance cod bankcard"`
	Source        string             `json:"source"`
	FromCart      bool               `json:"from_cart"` // 是否从购物车结算（结算后清空对应购物车项）
}

// OrderItemRequest 订单商品项请求
type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	SkuID     uint `json:"sku_id"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// OrderListRequest 订单列表请求
type OrderListRequest struct {
	utils.Pagination
	Status    *int   `form:"status" json:"status"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	OrderNo   string `form:"order_no" json:"order_no"`
	Keyword   string `form:"keyword" json:"keyword"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	Buyer     string `form:"buyer" json:"buyer"`
}

// AdminOrderListRequest 管理后台订单列表请求
type AdminOrderListRequest struct {
	utils.Pagination
	Keyword     string `form:"keyword" json:"keyword"`
	Status      *int   `form:"status" json:"status"`
	ShopID      uint   `form:"shop_id" json:"shop_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	OrderNo     string `form:"order_no" json:"order_no"`
	PaymentMethod string `form:"payment_method" json:"payment_method"`
	StartDate   string `form:"start_date" json:"start_date"`
	EndDate     string `form:"end_date" json:"end_date"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	LogisticsCompany string `json:"logistics_company" binding:"required,max=64"`
	LogisticsNo      string `json:"logistics_no" binding:"required,max=64"`
	SellerRemark     string `json:"seller_remark" binding:"max=500"`
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

// AdminCloseOrderRequest 管理后台关闭订单请求
type AdminCloseOrderRequest struct {
	AdminRemark string `json:"admin_remark" binding:"max=500"`
}

// OrderPayRequest 订单支付请求
type OrderPayRequest struct {
	PaymentMethod string `json:"payment_method" binding:"oneof=wechat alipay balance cod bankcard"`
	Channel       string `json:"channel"` // mp/h5/pc/miniapp
	ReturnURL     string `json:"return_url"`
}

// OrderPayResponse 订单支付响应
type OrderPayResponse struct {
	OrderID    uint   `json:"order_id"`
	OrderNo    string `json:"order_no"`
	PaymentNo  string `json:"payment_no"`
	PayURL     string `json:"pay_url,omitempty"`     // H5/PC 支付链接
	PayParams  interface{} `json:"pay_params,omitempty"` // 小程序/App 调起参数
	Status     int    `json:"status"`
}

// OrderSummaryRequest 订单统计请求
type OrderSummaryRequest struct {
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
}

// OrderSummary 订单统计响应
type OrderSummary struct {
	TotalCount      int64   `json:"total_count"`
	TotalAmount     float64 `json:"total_amount"`
	PaidCount       int64   `json:"paid_count"`
	PaidAmount      float64 `json:"paid_amount"`
	PendingCount    int64   `json:"pending_count"`
	ShippedCount    int64   `json:"shipped_count"`
	CompletedCount  int64   `json:"completed_count"`
	CancelledCount  int64   `json:"cancelled_count"`
	RefundedCount   int64   `json:"refunded_count"`
	RefundedAmount  float64 `json:"refunded_amount"`
}
