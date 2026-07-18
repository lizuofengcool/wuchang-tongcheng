// Package dto 团购优惠券数据传输对象
package dto

import "time"

// ===== 团购商品 =====

// GroupBuyInfo 团购商品信息
type GroupBuyInfo struct {
	ID            uint       `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Cover         string     `json:"cover"`
	OriginalPrice float64    `json:"original_price"`
	GroupBuyPrice float64    `json:"groupbuy_price"`
	Stock         int        `json:"stock"`
	SoldCount     int        `json:"sold_count"`
	PerLimit      int        `json:"per_limit"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	Status        int        `json:"status"`
	AuditStatus   int        `json:"audit_status"`
	IsRecommend   int        `json:"is_recommend"`
	Sort          int        `json:"sort"`
	ShopID        uint       `json:"shop_id"`
	UserID        uint       `json:"user_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateGroupBuyRequest 创建团购请求
type CreateGroupBuyRequest struct {
	Title         string     `json:"title" binding:"required,max=200"`
	Description   string     `json:"description"`
	Cover         string     `json:"cover" binding:"max=255"`
	OriginalPrice float64    `json:"original_price"`
	GroupBuyPrice float64    `json:"groupbuy_price" binding:"required"`
	Stock         int        `json:"stock"`
	PerLimit      int        `json:"per_limit"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	IsRecommend   int        `json:"is_recommend"`
	Sort          int        `json:"sort"`
	ShopID        uint       `json:"shop_id"`
}

// UpdateGroupBuyRequest 更新团购请求
type UpdateGroupBuyRequest struct {
	Title         string     `json:"title" binding:"max=200"`
	Description   string     `json:"description"`
	Cover         string     `json:"cover" binding:"max=255"`
	OriginalPrice float64    `json:"original_price"`
	GroupBuyPrice float64    `json:"groupbuy_price"`
	Stock         int        `json:"stock"`
	PerLimit      int        `json:"per_limit"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	IsRecommend   int        `json:"is_recommend"`
	Sort          int        `json:"sort"`
}

// GroupBuyListRequest 团购列表查询
type GroupBuyListRequest struct {
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
	Keyword     string `form:"keyword"`
	Status      int    `form:"status"`       // -1全部 0下架 1上架 2已结束
	IsRecommend int    `form:"is_recommend"` // -1全部 0否 1是
	ShopID      uint   `form:"shop_id"`
}

// UpdateGroupBuyStatusRequest 上下架请求
type UpdateGroupBuyStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// AuditGroupBuyRequest 审核请求
type AuditGroupBuyRequest struct {
	AuditStatus int `json:"audit_status" binding:"oneof=1 2"` // 1通过 2拒绝
}

// ===== 优惠券 =====

// CouponInfo 优惠券信息
type CouponInfo struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Type          int        `json:"type"`
	Value         float64    `json:"value"`
	MinAmount     float64    `json:"min_amount"`
	Scope         int        `json:"scope"`
	ScopeID       uint       `json:"scope_id"`
	TotalCount    int        `json:"total_count"`
	ReceivedCount int        `json:"received_count"`
	UsedCount     int        `json:"used_count"`
	PerLimit      int        `json:"per_limit"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	ValidityType  int        `json:"validity_type"`
	ValidDays     int        `json:"valid_days"`
	Status        int        `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateCouponRequest 创建优惠券请求
type CreateCouponRequest struct {
	Name         string     `json:"name" binding:"required,max=100"`
	Type         int        `json:"type" binding:"oneof=1 2 3"`
	Value        float64    `json:"value"`
	MinAmount    float64    `json:"min_amount"`
	Scope        int        `json:"scope" binding:"oneof=0 1 2"`
	ScopeID      uint       `json:"scope_id"`
	TotalCount   int        `json:"total_count"`
	PerLimit     int        `json:"per_limit"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ValidityType int        `json:"validity_type" binding:"oneof=0 1"`
	ValidDays    int        `json:"valid_days"`
	Status       int        `json:"status" binding:"oneof=0 1"`
}

// UpdateCouponRequest 更新优惠券请求
type UpdateCouponRequest struct {
	Name         string     `json:"name" binding:"max=100"`
	Type         int        `json:"type" binding:"omitempty,oneof=1 2 3"`
	Value        float64    `json:"value"`
	MinAmount    float64    `json:"min_amount"`
	Scope        int        `json:"scope" binding:"omitempty,oneof=0 1 2"`
	ScopeID      uint       `json:"scope_id"`
	TotalCount   int        `json:"total_count"`
	PerLimit     int        `json:"per_limit"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ValidityType int        `json:"validity_type" binding:"omitempty,oneof=0 1"`
	ValidDays    int        `json:"valid_days"`
	Status       int        `json:"status" binding:"omitempty,oneof=0 1"`
}

// CouponListRequest 优惠券列表查询
type CouponListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
	Status   int `form:"status"` // -1全部 0禁用 1启用
	Type     int `form:"type"`   // 0全部 1满减 2折扣 3代金券
}

// UserCouponInfo 用户优惠券信息
type UserCouponInfo struct {
	ID         uint        `json:"id"`
	UserID     uint        `json:"user_id"`
	CouponID   uint        `json:"coupon_id"`
	Status     int         `json:"status"`
	ReceivedAt time.Time   `json:"received_at"`
	UsedAt     *time.Time  `json:"used_at"`
	ExpireAt   *time.Time  `json:"expire_at"`
	OrderID    uint        `json:"order_id"`
	Coupon     *CouponInfo `json:"coupon,omitempty"`
}

// ===== 团购订单 =====

// OrderInfo 订单信息
type OrderInfo struct {
	ID             uint          `json:"id"`
	OrderNo        string        `json:"order_no"`
	UserID         uint          `json:"user_id"`
	GroupBuyID     uint          `json:"groupbuy_id"`
	Quantity       int           `json:"quantity"`
	UnitPrice      float64       `json:"unit_price"`
	TotalPrice     float64       `json:"total_price"`
	CouponID       uint          `json:"coupon_id"`
	DiscountAmount float64       `json:"discount_amount"`
	PayAmount      float64       `json:"pay_amount"`
	PayStatus      int           `json:"pay_status"`
	PayAt          *time.Time    `json:"pay_at"`
	VerifyStatus   int           `json:"verify_status"`
	VerifyCode     string        `json:"verify_code"`
	VerifyAt       *time.Time    `json:"verify_at"`
	Status         int           `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	GroupBuy       *GroupBuyInfo `json:"groupbuy,omitempty"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	GroupBuyID   uint `json:"groupbuy_id" binding:"required"`
	Quantity     int  `json:"quantity" binding:"required,min=1"`
	UserCouponID uint `json:"user_coupon_id"`
}

// OrderListRequest 订单列表查询
type OrderListRequest struct {
	Page       int  `form:"page"`
	PageSize   int  `form:"page_size"`
	Status     int  `form:"status"`      // -1全部
	PayStatus  int  `form:"pay_status"`  // -1全部
	UserID     uint `form:"user_id"`
	GroupBuyID uint `form:"groupbuy_id"`
}

// VerifyOrderRequest 核销订单请求
type VerifyOrderRequest struct {
	VerifyCode string `json:"verify_code" binding:"required"`
}
