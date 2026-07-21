// Package dto 营销活动中台数据传输对象 - 优惠券（coupon 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CouponInfo 优惠券详情响应
type CouponInfo struct {
	ID            uint       `json:"id"`
	RegionID      uint       `json:"region_id"`
	Title         string     `json:"title"`
	Type          string     `json:"type"`
	TypeText      string     `json:"type_text"`
	Amount        float64    `json:"amount"`
	Threshold     float64    `json:"threshold"`
	TotalCount    int        `json:"total_count"`
	ReceivedCount int        `json:"received_count"`
	StartAt       *time.Time `json:"start_at"`
	EndAt         *time.Time `json:"end_at"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Title      string     `json:"title" binding:"required,max=100"`
	Type       string     `json:"type" binding:"required,oneof=discount reduce exchange"`
	Amount     float64    `json:"amount" binding:"min=0"`
	Threshold  float64    `json:"threshold" binding:"min=0"`
	TotalCount int        `json:"total_count" binding:"min=0"`
	StartAt    *time.Time `json:"start_at"`
	EndAt      *time.Time `json:"end_at"`
	Status     int        `json:"status"`
}

// UpdateCouponRequest 更新优惠券请求
type UpdateCouponRequest struct {
	Title      *string     `json:"title" binding:"omitempty,max=100"`
	Type       *string     `json:"type" binding:"omitempty,oneof=discount reduce exchange"`
	Amount     *float64    `json:"amount" binding:"omitempty,min=0"`
	Threshold  *float64    `json:"threshold" binding:"omitempty,min=0"`
	TotalCount *int        `json:"total_count" binding:"omitempty,min=0"`
	StartAt    *time.Time  `json:"start_at"`
	EndAt      *time.Time  `json:"end_at"`
	Status     *int        `json:"status"`
}

// CouponListRequest 优惠券列表请求
type CouponListRequest struct {
	Type    string `form:"type" json:"type"`
	Status  *int   `form:"status" json:"status"`
	Keyword string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CouponReceiveRequest 领取优惠券请求
type CouponReceiveRequest struct {
	Source string `json:"source"`
}

// CouponUseRequest 使用优惠券请求
type CouponUseRequest struct {
	OrderID uint `json:"order_id"`
}

// UserCouponInfo 用户优惠券详情响应
type UserCouponInfo struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	CouponID      uint       `json:"coupon_id"`
	Status        string     `json:"status"`
	StatusText    string     `json:"status_text"`
	Source        string     `json:"source"`
	SourceText    string     `json:"source_text"`
	UsedAt        *time.Time `json:"used_at"`
	OrderID       uint       `json:"order_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// 冗余优惠券信息（联表查询时填充）
	CouponTitle  string  `json:"coupon_title,omitempty"`
	CouponType   string  `json:"coupon_type,omitempty"`
	CouponAmount float64 `json:"coupon_amount,omitempty"`
}

// UserCouponListRequest 用户优惠券列表请求
type UserCouponListRequest struct {
	Status string `form:"status" json:"status"`
	utils.Pagination
}

// CouponStatistics 优惠券领取统计
type CouponStatistics struct {
	TotalCoupons    int64 `json:"total_coupons"`
	ActiveCoupons   int64 `json:"active_coupons"`
	TotalReceived   int64 `json:"total_received"`
	TotalUsed       int64 `json:"total_used"`
	ReceiveRate     float64 `json:"receive_rate"` // 领取率
	UsageRate       float64 `json:"usage_rate"`   // 使用率
}
