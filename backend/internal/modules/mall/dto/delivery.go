// Package dto 同城商城 - 配送单 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// DeliveryInfo 配送单详情响应
type DeliveryInfo struct {
	ID             uint       `json:"id"`
	OrderID        uint       `json:"order_id"`
	RiderID        *uint      `json:"rider_id"`
	ShopID         uint       `json:"shop_id"`
	UserID         uint       `json:"user_id"`
	DeliveryNo     string     `json:"delivery_no"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	PickupAddress  string     `json:"pickup_address"`
	PickupLat      float64    `json:"pickup_lat"`
	PickupLng      float64    `json:"pickup_lng"`
	DeliveryAddress string    `json:"delivery_address"`
	DeliveryLat    float64    `json:"delivery_lat"`
	DeliveryLng    float64    `json:"delivery_lng"`
	Distance       float64    `json:"distance"`
	DeliveryFee    float64    `json:"delivery_fee"`
	Tip            float64    `json:"tip"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	PickedAt       *time.Time `json:"picked_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	CancelReason   string     `json:"cancel_reason"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DeliveryCreateRequest 配送单创建请求（系统内部调用）
type DeliveryCreateRequest struct {
	OrderID         uint    `json:"order_id" binding:"required"`
	ShopID          uint    `json:"shop_id" binding:"required"`
	UserID          uint    `json:"user_id" binding:"required"`
	DeliveryNo      string  `json:"delivery_no" binding:"required,max=50"`
	PickupAddress   string  `json:"pickup_address" binding:"max=500"`
	PickupLat       float64 `json:"pickup_lat"`
	PickupLng       float64 `json:"pickup_lng"`
	DeliveryAddress string  `json:"delivery_address" binding:"max=500"`
	DeliveryLat     float64 `json:"delivery_lat"`
	DeliveryLng     float64 `json:"delivery_lng"`
	Distance        float64 `json:"distance"`
	DeliveryFee     float64 `json:"delivery_fee"`
	Tip             float64 `json:"tip"`
}

// DeliveryListRequest 配送单列表请求（抢单大厅/我的配送单）
type DeliveryListRequest struct {
	utils.Pagination
	Keyword    string `form:"keyword" json:"keyword"`
	Status     *int   `form:"status" json:"status"`
	RiderID    uint   `form:"rider_id" json:"rider_id"`
	ShopID     uint   `form:"shop_id" json:"shop_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	RegionID   uint   `form:"region_id" json:"region_id"`
	DeliveryNo string `form:"delivery_no" json:"delivery_no"`
}

// DeliveryCancelRequest 配送单取消请求
type DeliveryCancelRequest struct {
	CancelReason string `json:"cancel_reason" binding:"max=500"`
}

// DeliveryStatsResponse 配送统计响应
type DeliveryStatsResponse struct {
	TotalOrders    int64   `json:"total_orders"`
	CompletedCount int64   `json:"completed_count"`
	PendingCount   int64   `json:"pending_count"`
	CancelledCount int64   `json:"cancelled_count"`
	TotalEarnings  float64 `json:"total_earnings"`
	TodayOrders    int64   `json:"today_orders"`
	TodayEarnings  float64 `json:"today_earnings"`
}
