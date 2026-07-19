// Package dto 同城车辆买卖数据传输对象 - 试驾预约
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// TestDriveInfo 试驾预约详情响应
type TestDriveInfo struct {
	ID              uint       `json:"id"`
	DriveNo         string     `json:"drive_no"`
	CarID           uint       `json:"car_id"`
	ListingID       uint       `json:"listing_id"`
	UserID          uint       `json:"user_id"`
	UserName        string     `json:"user_name"`
	UserPhone       string     `json:"user_phone"`
	UserAvatar      string     `json:"user_avatar"`
	DealerID        uint       `json:"dealer_id"`
	DealerName      string     `json:"dealer_name"`
	SalesID         uint       `json:"sales_id"`
	SalesName       string     `json:"sales_name"`
	AppointmentDate *time.Time `json:"appointment_date"`
	AppointmentTime string     `json:"appointment_time"`
	Address         string     `json:"address"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	DriveType       string     `json:"drive_type"`
	LicenseStatus   string     `json:"license_status"`
	LicenseImages   interface{} `json:"license_images"`
	Remark          string     `json:"remark"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	CancelReason    string     `json:"cancel_reason"`
	Result          string     `json:"result"`
	ResultRemark    string     `json:"result_remark"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CanceledAt      *time.Time `json:"canceled_at"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateTestDriveRequest 创建试驾预约请求
type CreateTestDriveRequest struct {
	CarID           uint       `json:"car_id" binding:"required"`
	ListingID       uint       `json:"listing_id"`
	DealerID        uint       `json:"dealer_id"`
	DealerName      string     `json:"dealer_name" binding:"max=128"`
	SalesID         uint       `json:"sales_id"`
	SalesName       string     `json:"sales_name" binding:"max=50"`
	AppointmentDate *time.Time `json:"appointment_date" binding:"required"`
	AppointmentTime string     `json:"appointment_time" binding:"required,max=32"`
	Address         string     `json:"address" binding:"required,max=500"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	DriveType       string     `json:"drive_type" binding:"omitempty,oneof=test_drive viewing delivery"`
	Remark          string     `json:"remark" binding:"max=500"`
}

// UpdateTestDriveRequest 更新试驾预约请求
type UpdateTestDriveRequest struct {
	SalesID         *uint      `json:"sales_id"`
	SalesName       *string    `json:"sales_name" binding:"omitempty,max=50"`
	AppointmentDate *time.Time `json:"appointment_date"`
	AppointmentTime *string    `json:"appointment_time" binding:"omitempty,max=32"`
	Address         *string    `json:"address" binding:"omitempty,max=500"`
	Latitude        *float64   `json:"latitude"`
	Longitude       *float64   `json:"longitude"`
	DriveType       *string    `json:"drive_type" binding:"omitempty,oneof=test_drive viewing delivery"`
	LicenseStatus   *string    `json:"license_status" binding:"omitempty,oneof=unsubmitted pending verified rejected"`
	LicenseImages   interface{} `json:"license_images"`
	Remark          *string    `json:"remark" binding:"omitempty,max=500"`
	Status          *int       `json:"status" binding:"omitempty,oneof=0 1 2 3 4 5"`
}

// TestDriveListRequest 试驾预约列表请求
type TestDriveListRequest struct {
	CarID      uint   `form:"car_id" json:"car_id"`
	ListingID  uint   `form:"listing_id" json:"listing_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	DealerID   uint   `form:"dealer_id" json:"dealer_id"`
	SalesID    uint   `form:"sales_id" json:"sales_id"`
	DriveType  string `form:"drive_type" json:"drive_type"`
	Status     *int   `form:"status" json:"status"`
	StartDate  string `form:"start_date" json:"start_date"`
	EndDate    string `form:"end_date" json:"end_date"`
	utils.Pagination
}

// TestDriveStatusUpdateRequest 状态更新请求
type TestDriveStatusUpdateRequest struct {
	Status       int    `json:"status" binding:"oneof=0 1 2 3 4 5"`
	CancelReason string `json:"cancel_reason" binding:"max=500"`
	Result       string `json:"result" binding:"omitempty,oneof=satisfied dissatisfied undecided purchased not_purchased"`
	ResultRemark string `json:"result_remark"`
}

// TestDriveLicenseUploadRequest 驾照上传请求
type TestDriveLicenseUploadRequest struct {
	LicenseImages interface{} `json:"license_images"`
}
