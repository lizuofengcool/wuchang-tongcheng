// Package dto 同城拼车出行数据传输对象 - 车主认证
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// DriverInfo 车主认证详情响应
type DriverInfo struct {
	ID                  uint       `json:"id"`
	RegionID            uint       `json:"region_id"`
	UserID              uint       `json:"user_id"`
	UserName            string     `json:"user_name"`
	UserPhone           string     `json:"user_phone"`
	UserAvatar          string     `json:"user_avatar"`
	RealName            string     `json:"real_name"`
	IDCardNo            string     `json:"id_card_no"` // 脱敏后
	IDCardFront         string     `json:"id_card_front"`
	IDCardBack          string     `json:"id_card_back"`
	DriverLicenseNo     string     `json:"driver_license_no"`
	DriverLicenseFront  string     `json:"driver_license_front"`
	DriverLicenseBack   string     `json:"driver_license_back"`
	LicenseIssueDate    *time.Time `json:"license_issue_date"`
	LicenseExpiryDate   *time.Time `json:"license_expiry_date"`
	VehicleLicenseNo    string     `json:"vehicle_license_no"`
	VehicleLicenseFront string     `json:"vehicle_license_front"`
	VehicleLicenseBack  string     `json:"vehicle_license_back"`
	CarPhoto            string     `json:"car_photo"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	AuditReason         string     `json:"audit_reason"`
	AuditedAt           *time.Time `json:"audited_at"`
	AuditorID           *uint      `json:"auditor_id"`
	Verified            bool       `json:"verified"`
	VerifiedAt          *time.Time `json:"verified_at"`
	RatingAvg           float64    `json:"rating_avg"`
	TripCount           int        `json:"trip_count"`
	TotalIncome         float64    `json:"total_income"`
	CreatedAt           time.Time  `json:"created_at"`
}

// CreateDriverRequest 车主认证请求
type CreateDriverRequest struct {
	RealName            string `json:"real_name" binding:"required,min=2,max=50"`
	IDCardNo            string `json:"id_card_no" binding:"required,len=18"`
	IDCardFront         string `json:"id_card_front" binding:"required"`
	IDCardBack          string `json:"id_card_back" binding:"required"`
	DriverLicenseNo     string `json:"driver_license_no" binding:"required"`
	DriverLicenseFront  string `json:"driver_license_front" binding:"required"`
	DriverLicenseBack   string `json:"driver_license_back" binding:"required"`
	LicenseIssueDate    *time.Time `json:"license_issue_date"`
	LicenseExpiryDate   *time.Time `json:"license_expiry_date" binding:"required"`
	VehicleLicenseNo    string `json:"vehicle_license_no" binding:"required"`
	VehicleLicenseFront string `json:"vehicle_license_front" binding:"required"`
	VehicleLicenseBack  string `json:"vehicle_license_back" binding:"required"`
	CarPhoto            string `json:"car_photo" binding:"required"`
}

// UpdateDriverRequest 更新车主认证请求
type UpdateDriverRequest struct {
	IDCardFront         *string    `json:"id_card_front"`
	IDCardBack          *string    `json:"id_card_back"`
	DriverLicenseFront  *string    `json:"driver_license_front"`
	DriverLicenseBack   *string    `json:"driver_license_back"`
	LicenseExpiryDate   *time.Time `json:"license_expiry_date"`
	VehicleLicenseFront *string    `json:"vehicle_license_front"`
	VehicleLicenseBack  *string    `json:"vehicle_license_back"`
	CarPhoto            *string    `json:"car_photo"`
}

// DriverListRequest 车主列表查询请求
type DriverListRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	Status   *int   `form:"status" json:"status"`
	Verified *bool  `form:"verified" json:"verified"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// DriverReviewRequest 车主审核请求
type DriverReviewRequest struct {
	Status int    `json:"status" binding:"oneof=1 2"`
	Reason string `json:"reason" binding:"max=500"`
}
