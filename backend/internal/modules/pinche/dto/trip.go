// Package dto 同城拼车出行数据传输对象 - 完成行程/行程分享
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// TripInfo 完成行程详情响应
type TripInfo struct {
	ID                  uint       `json:"id"`
	RegionID            uint       `json:"region_id"`
	PincheID            uint       `json:"pinche_id"`
	BookingID           *uint      `json:"booking_id"`
	TripNo              string     `json:"trip_no"`
	DriverID            uint       `json:"driver_id"`
	DriverName          string     `json:"driver_name"`
	DriverPhone         string     `json:"driver_phone"`
	PassengerID         uint       `json:"passenger_id"`
	PassengerName       string     `json:"passenger_name"`
	PassengerPhone      string     `json:"passenger_phone"`
	VehicleID           *uint      `json:"vehicle_id"`
	PlateNo             string     `json:"plate_no"`
	OriginAddress       string     `json:"origin_address"`
	OriginLat           float64    `json:"origin_lat"`
	OriginLng           float64    `json:"origin_lng"`
	DestinationAddress  string     `json:"destination_address"`
	DestinationLat      float64    `json:"destination_lat"`
	DestinationLng      float64    `json:"destination_lng"`
	ActualPickupTime    *time.Time `json:"actual_pickup_time"`
	ActualDropoffTime   *time.Time `json:"actual_dropoff_time"`
	ActualDistanceKM    float64    `json:"actual_distance_km"`
	ActualDurationMin   int        `json:"actual_duration_min"`
	PassengersCount     int        `json:"passengers_count"`
	FareAmount          float64    `json:"fare_amount"`
	TollFee             float64    `json:"toll_fee"`
	TotalAmount         float64    `json:"total_amount"`
	ShareToken          string     `json:"share_token"`
	ShareExpiresAt      *time.Time `json:"share_expires_at"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	DriverConfirmedAt   *time.Time `json:"driver_confirmed_at"`
	PassengerConfirmedAt *time.Time `json:"passenger_confirmed_at"`
	CompletedAt         *time.Time `json:"completed_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

// StartTripRequest 启动行程请求
type StartTripRequest struct {
	PincheID         uint    `json:"pinche_id" binding:"required"`
	BookingID        *uint   `json:"booking_id"`
	ActualPickupLat  float64 `json:"actual_pickup_lat"`
	ActualPickupLng  float64 `json:"actual_pickup_lng"`
	ActualPickupAddr string  `json:"actual_pickup_addr" binding:"max=255"`
}

// CompleteTripRequest 完成行程请求
type CompleteTripRequest struct {
	ActualDropoffLat  float64 `json:"actual_dropoff_lat"`
	ActualDropoffLng  float64 `json:"actual_dropoff_lng"`
	ActualDropoffAddr string  `json:"actual_dropoff_addr" binding:"max=255"`
	ActualDistanceKM  float64 `json:"actual_distance_km" binding:"omitempty,min=0"`
	ActualDurationMin int     `json:"actual_duration_min" binding:"omitempty,min=0"`
	TollFee           float64 `json:"toll_fee" binding:"omitempty,min=0"`
}

// ConfirmTripRequest 确认行程请求（车主/乘客确认完成）
type ConfirmTripRequest struct {
	ConfirmRole string `json:"confirm_role" binding:"required,oneof=driver passenger"`
}

// TripListRequest 行程列表查询请求
type TripListRequest struct {
	UserID      uint   `form:"user_id" json:"user_id"`
	DriverID    uint   `form:"driver_id" json:"driver_id"`
	PassengerID uint   `form:"passenger_id" json:"passenger_id"`
	PincheID    uint   `form:"pinche_id" json:"pinche_id"`
	Status      *int   `form:"status" json:"status"`
	TripNo      string `form:"trip_no" json:"trip_no"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
	utils.Pagination
}

// TripShareResponse 行程分享响应
type TripShareResponse struct {
	ShareToken     string     `json:"share_token"`
	ShareURL       string     `json:"share_url"`
	ShareExpiresAt *time.Time `json:"share_expires_at"`
	TripNo         string     `json:"trip_no"`
	DriverName     string     `json:"driver_name"`
	PlateNo        string     `json:"plate_no"`
	OriginAddress  string     `json:"origin_address"`
	DestinationAddress string  `json:"destination_address"`
	Status         int        `json:"status"`
}

// TripLocationUpdateRequest 实时位置更新请求（行程进行中）
type TripLocationUpdateRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Speed     float64 `json:"speed" binding:"omitempty,min=0"`
	Heading   float64 `json:"heading" binding:"omitempty,min=0,max=360"`
}
