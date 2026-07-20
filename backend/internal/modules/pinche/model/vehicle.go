// Package model 拼车车辆信息数据模型
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// PincheVehicle 车辆信息
type PincheVehicle struct {
	database.RegionBaseModel

	DriverID uint `gorm:"index;not null" json:"driver_id"`
	UserID   uint `gorm:"index;not null" json:"user_id"`

	// 车辆信息
	PlateNo     string `gorm:"size:16;index" json:"plate_no"`
	Brand       string `gorm:"size:64" json:"brand"`
	Model       string `gorm:"size:128" json:"model"`
	Year        int    `gorm:"not null;default:0" json:"year"`
	Color       string `gorm:"size:32" json:"color"`
	SeatCount   int    `gorm:"not null;default:5" json:"seat_count"`
	VehicleType string `gorm:"size:32;not null;default:'sedan'" json:"vehicle_type"` // sedan/suv/mpv/new_energy
	FuelType    string `gorm:"size:32;not null;default:'gasoline'" json:"fuel_type"` // gasoline/electric/hybrid

	// 照片
	VehiclePhotos      JSONB `gorm:"type:jsonb" json:"vehicle_photos"`
	VehicleLicensePhoto string `gorm:"size:255" json:"vehicle_license_photo"`
	InsurancePhoto     string `gorm:"size:255" json:"insurance_photo"`

	// 状态
	Status      int    `gorm:"default:0;index" json:"status"` // 0待审 1通过 2拒绝
	AuditStatus int    `gorm:"default:0;index" json:"audit_status"`
	AuditReason string `gorm:"size:500" json:"audit_reason"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"`
}

// TableName 表名
func (PincheVehicle) TableName() string { return "pinche_vehicles" }
