// Package model 拼车车主认证数据模型
// 身份证+驾驶证+行驶证+车辆照片
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheDriver 车主认证
type PincheDriver struct {
	database.RegionBaseModel

	UserID     uint   `gorm:"index;not null" json:"user_id"`
	UserName   string `gorm:"size:50" json:"user_name"`
	UserPhone  string `gorm:"size:20" json:"user_phone"`
	UserAvatar string `gorm:"size:255" json:"user_avatar"`

	// 实名信息
	RealName    string `gorm:"size:50" json:"real_name"`
	IDCardNo    string `gorm:"size:32" json:"id_card_no"`
	IDCardFront string `gorm:"size:255" json:"id_card_front"`
	IDCardBack  string `gorm:"size:255" json:"id_card_back"`

	// 驾驶证
	DriverLicenseNo     string     `gorm:"size:32" json:"driver_license_no"`
	DriverLicenseFront  string     `gorm:"size:255" json:"driver_license_front"`
	DriverLicenseBack   string     `gorm:"size:255" json:"driver_license_back"`
	LicenseIssueDate    *time.Time `gorm:"type:date" json:"license_issue_date"`
	LicenseExpiryDate   *time.Time `gorm:"type:date;index" json:"license_expiry_date"`

	// 行驶证与车辆
	VehicleLicenseNo    string `gorm:"size:32" json:"vehicle_license_no"`
	VehicleLicenseFront string `gorm:"size:255" json:"vehicle_license_front"`
	VehicleLicenseBack  string `gorm:"size:255" json:"vehicle_license_back"`
	CarPhoto            string `gorm:"size:255" json:"car_photo"`

	// 状态
	Status     int        `gorm:"default:0;index" json:"status"` // 0待审 1通过 2拒绝 3已过期
	AuditReason string    `gorm:"size:500" json:"audit_reason"`
	AuditedAt  *time.Time `gorm:"index" json:"audited_at"`
	AuditorID  *uint      `gorm:"index" json:"auditor_id"`
	Verified   bool       `gorm:"default:false;index" json:"verified"`
	VerifiedAt *time.Time `gorm:"index" json:"verified_at"`

	// 统计
	RatingAvg   float64 `gorm:"type:decimal(3,2);default:5.00" json:"rating_avg"`
	TripCount   int     `gorm:"not null;default:0" json:"trip_count"`
	TotalIncome float64 `gorm:"type:decimal(12,2);default:0" json:"total_income"`
}

// TableName 表名
func (PincheDriver) TableName() string { return "pinche_drivers" }
