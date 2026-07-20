// Package dto 同城拼车出行数据传输对象 - 预订
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// BookingInfo 预订详情响应
type BookingInfo struct {
	ID              uint       `json:"id"`
	RegionID        uint       `json:"region_id"`
	PincheID        uint       `json:"pinche_id"`
	BookingNo       string     `json:"booking_no"`
	PassengerID     uint       `json:"passenger_id"`
	PassengerName   string     `json:"passenger_name"`
	PassengerPhone  string     `json:"passenger_phone"`
	PassengerAvatar string     `json:"passenger_avatar"`
	DriverID        uint       `json:"driver_id"`
	DriverName      string     `json:"driver_name"`
	DriverPhone     string     `json:"driver_phone"`
	Seats           int        `json:"seats"`
	PickupLocation  string     `json:"pickup_location"`
	PickupLat       float64    `json:"pickup_lat"`
	PickupLng       float64    `json:"pickup_lng"`
	DropoffLocation string     `json:"dropoff_location"`
	DropoffLat      float64    `json:"dropoff_lat"`
	DropoffLng      float64    `json:"dropoff_lng"`
	UnitPrice       float64    `json:"unit_price"`
	TotalAmount     float64    `json:"total_amount"`
	InsuranceFee    float64    `json:"insurance_fee"`
	ServiceFee      float64    `json:"service_fee"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	PaymentID       *uint      `json:"payment_id"`
	BoardingCode    string     `json:"boarding_code"`
	PaidAt          *time.Time `json:"paid_at"`
	BoardedAt       *time.Time `json:"boarded_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CancelledAt     *time.Time `json:"cancelled_at"`
	CancelReason    string     `json:"cancel_reason"`
	CancelledBy     *uint      `json:"cancelled_by"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateBookingRequest 创建预订请求
type CreateBookingRequest struct {
	PincheID        uint   `json:"pinche_id" binding:"required"`
	Seats           int    `json:"seats" binding:"required,min=1,max=10"`
	PickupLocation  string `json:"pickup_location" binding:"required,max=255"`
	PickupLat       float64 `json:"pickup_lat"`
	PickupLng       float64 `json:"pickup_lng"`
	DropoffLocation string `json:"dropoff_location" binding:"required,max=255"`
	DropoffLat      float64 `json:"dropoff_lat"`
	DropoffLng      float64 `json:"dropoff_lng"`
	BuyInsurance    bool   `json:"buy_insurance"`
}

// UpdateBookingRequest 更新预订请求
type UpdateBookingRequest struct {
	Seats           *int    `json:"seats"`
	PickupLocation  *string `json:"pickup_location"`
	PickupLat       *float64 `json:"pickup_lat"`
	PickupLng       *float64 `json:"pickup_lng"`
	DropoffLocation *string `json:"dropoff_location"`
	DropoffLat      *float64 `json:"dropoff_lat"`
	DropoffLng      *float64 `json:"dropoff_lng"`
}

// CancelBookingRequest 取消预订请求
type CancelBookingRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

// ConfirmBoardingRequest 确认上车请求
type ConfirmBoardingRequest struct {
	BoardingCode string `json:"boarding_code" binding:"required"`
}

// BookingListRequest 预订列表查询请求
type BookingListRequest struct {
	PincheID    uint   `form:"pinche_id" json:"pinche_id"`
	PassengerID uint   `form:"passenger_id" json:"passenger_id"`
	DriverID    uint   `form:"driver_id" json:"driver_id"`
	Status      *int   `form:"status" json:"status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}
