// Package dto 同城商城 - 骑手 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RiderInfo 骑手详情响应
type RiderInfo struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	ShopID        *uint      `json:"shop_id"`
	RealName      string     `json:"real_name"`
	Phone         string     `json:"phone"`
	IDCard        string     `json:"id_card"`
	Avatar        string     `json:"avatar"`
	VehicleType   string     `json:"vehicle_type"`
	VehicleTypeText string   `json:"vehicle_type_text"`
	VehiclePlate  string     `json:"vehicle_plate"`
	LicenseURL    string     `json:"license_url"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	CreditScore   int        `json:"credit_score"`
	Level         int        `json:"level"`
	TotalOrders   int        `json:"total_orders"`
	TotalEarnings float64    `json:"total_earnings"`
	OnlineStatus  int        `json:"online_status"`
	OnlineStatusText string  `json:"online_status_text"`
	AuditReason   string     `json:"audit_reason"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// RiderApplyRequest 骑手申请请求
type RiderApplyRequest struct {
	RealName     string `json:"real_name" binding:"required,max=50"`
	Phone        string `json:"phone" binding:"required,max=20"`
	IDCard       string `json:"id_card" binding:"max=18"`
	Avatar       string `json:"avatar" binding:"max=255"`
	VehicleType  string `json:"vehicle_type" binding:"omitempty,oneof=electric motor bicycle car"`
	VehiclePlate string `json:"vehicle_plate" binding:"max=20"`
	LicenseURL   string `json:"license_url" binding:"max=255"`
	ShopID       *uint  `json:"shop_id"`
}

// RiderUpdateRequest 更新骑手资料请求
type RiderUpdateRequest struct {
	RealName     *string `json:"real_name" binding:"max=50"`
	Phone        *string `json:"phone" binding:"max=20"`
	IDCard       *string `json:"id_card" binding:"max=18"`
	Avatar       *string `json:"avatar" binding:"max=255"`
	VehicleType  *string `json:"vehicle_type" binding:"omitempty,oneof=electric motor bicycle car"`
	VehiclePlate *string `json:"vehicle_plate" binding:"max=20"`
	LicenseURL   *string `json:"license_url" binding:"max=255"`
}

// RiderListRequest 骑手列表请求（C 端 / 管理端共用）
type RiderListRequest struct {
	utils.Pagination
	Keyword      string `form:"keyword" json:"keyword"`
	Status       *int   `form:"status" json:"status"`
	OnlineStatus *int   `form:"online_status" json:"online_status"`
	UserID       uint   `form:"user_id" json:"user_id"`
	ShopID       uint   `form:"shop_id" json:"shop_id"`
	RegionID     uint   `form:"region_id" json:"region_id"`
}

// RiderAuditRequest 骑手审核请求（M 端）
type RiderAuditRequest struct {
	Status      int    `json:"status" binding:"oneof=1 2 3"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// RiderStatusUpdateRequest 骑手状态更新请求（M 端冻结/解冻）
type RiderStatusUpdateRequest struct {
	Status int `json:"status" binding:"oneof=1 3"` // 1解冻（通过）3冻结
}

// RiderEarningsResponse 骑手收益统计响应
type RiderEarningsResponse struct {
	TotalOrders    int64   `json:"total_orders"`
	TotalEarnings  float64 `json:"total_earnings"`
	MonthOrders    int64   `json:"month_orders"`
	MonthEarnings  float64 `json:"month_earnings"`
	TodayOrders    int64   `json:"today_orders"`
	TodayEarnings  float64 `json:"today_earnings"`
	PendingEarnings float64 `json:"pending_earnings"` // 待结算
}
