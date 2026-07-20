// Package dto 同城拼车出行数据传输对象 - 车辆
package dto

import (
	"wuchang-tongcheng/internal/pkg/utils"
)

// VehicleInfo 车辆详情响应
type VehicleInfo struct {
	ID                  uint        `json:"id"`
	RegionID            uint        `json:"region_id"`
	DriverID            uint        `json:"driver_id"`
	UserID              uint        `json:"user_id"`
	PlateNo             string      `json:"plate_no"`
	Brand               string      `json:"brand"`
	Model               string      `json:"model"`
	Year                int         `json:"year"`
	Color               string      `json:"color"`
	SeatCount           int         `json:"seat_count"`
	VehicleType         string      `json:"vehicle_type"`
	FuelType            string      `json:"fuel_type"`
	VehiclePhotos       interface{} `json:"vehicle_photos"`
	VehicleLicensePhoto string      `json:"vehicle_license_photo"`
	InsurancePhoto      string      `json:"insurance_photo"`
	Status              int         `json:"status"`
	StatusText          string      `json:"status_text"`
	AuditStatus         int         `json:"audit_status"`
	AuditReason         string      `json:"audit_reason"`
	IsDefault           bool        `json:"is_default"`
}

// CreateVehicleRequest 创建车辆请求
type CreateVehicleRequest struct {
	PlateNo             string      `json:"plate_no" binding:"required,max=16"`
	Brand               string      `json:"brand" binding:"required,max=64"`
	Model               string      `json:"model" binding:"max=128"`
	Year                int         `json:"year" binding:"min=1980,max=2030"`
	Color               string      `json:"color" binding:"max=32"`
	SeatCount           int         `json:"seat_count" binding:"required,min=2,max=10"`
	VehicleType         string      `json:"vehicle_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck"`
	FuelType            string      `json:"fuel_type" binding:"omitempty,oneof=gasoline electric hybrid"`
	VehiclePhotos       interface{} `json:"vehicle_photos"`
	VehicleLicensePhoto string      `json:"vehicle_license_photo" binding:"required"`
	InsurancePhoto      string      `json:"insurance_photo" binding:"required"`
	IsDefault           bool        `json:"is_default"`
}

// UpdateVehicleRequest 更新车辆请求
type UpdateVehicleRequest struct {
	PlateNo             *string     `json:"plate_no"`
	Brand               *string     `json:"brand"`
	Model               *string     `json:"model"`
	Year                *int        `json:"year"`
	Color               *string     `json:"color"`
	SeatCount           *int        `json:"seat_count"`
	VehicleType         *string     `json:"vehicle_type"`
	FuelType            *string     `json:"fuel_type"`
	VehiclePhotos       interface{} `json:"vehicle_photos"`
	VehicleLicensePhoto *string     `json:"vehicle_license_photo"`
	InsurancePhoto      *string     `json:"insurance_photo"`
	IsDefault           *bool       `json:"is_default"`
}

// VehicleListRequest 车辆列表查询请求
type VehicleListRequest struct {
	DriverID uint   `form:"driver_id" json:"driver_id"`
	UserID   uint   `form:"user_id" json:"user_id"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// VehicleReviewRequest 车辆审核请求
type VehicleReviewRequest struct {
	Status int    `json:"status" binding:"oneof=1 2"`
	Reason string `json:"reason" binding:"max=500"`
}
