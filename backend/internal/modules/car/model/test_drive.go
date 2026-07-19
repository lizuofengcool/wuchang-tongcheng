// Package model 试驾预约表（对标懂车帝/汽车之家）
// 时间/地点/销售/驾照/结果
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 试驾状态常量 ===
const (
	TestDriveStatusPending    = 0 // 待确认
	TestDriveStatusConfirmed  = 1 // 已确认
	TestDriveStatusInProgress = 2 // 试驾中
	TestDriveStatusCompleted  = 3 // 已完成
	TestDriveStatusCanceled   = 4 // 已取消
	TestDriveStatusNoShow     = 5 // 未到店
)

// === 试驾类型常量 ===
const (
	TestDriveTypeTestDrive = "test_drive" // 试驾
	TestDriveTypeViewing   = "viewing"    // 看车
	TestDriveTypeDelivery  = "delivery"   // 上门试驾
)

// === 驾照状态常量 ===
const (
	LicenseStatusUnsubmitted = "unsubmitted" // 未提交
	LicenseStatusPending     = "pending"     // 待审核
	LicenseStatusVerified    = "verified"    // 已验证
	LicenseStatusRejected    = "rejected"    // 不通过
)

// === 试驾结果常量 ===
const (
	TestDriveResultSatisfied      = "satisfied"      // 满意
	TestDriveResultDissatisfied   = "dissatisfied"   // 不满意
	TestDriveResultUndecided      = "undecided"      // 未决定
	TestDriveResultPurchased      = "purchased"      // 已购买
	TestDriveResultNotPurchased   = "not_purchased"  // 未购买
)

// CarTestDrive 试驾预约表
type CarTestDrive struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	DriveNo          string     `gorm:"size:64;not null;uniqueIndex" json:"drive_no"`                  // 预约单号
	CarID            uint       `gorm:"not null;index" json:"car_id"`                                  // 车源 ID
	ListingID        uint       `gorm:"not null;default:0;index" json:"listing_id"`                    // 发布 ID
	UserID           uint       `gorm:"not null;index" json:"user_id"`                                 // 用户 ID
	UserName         string     `gorm:"size:50;not null;default:''" json:"user_name"`                  // 用户姓名
	UserPhone        string     `gorm:"size:20;not null;default:''" json:"user_phone"`                 // 用户手机
	UserAvatar       string     `gorm:"size:255;not null;default:''" json:"user_avatar"`               // 用户头像
	DealerID         uint       `gorm:"not null;default:0;index" json:"dealer_id"`                     // 车商 ID
	DealerName       string     `gorm:"size:128;not null;default:''" json:"dealer_name"`               // 车商名
	SalesID          uint       `gorm:"not null;default:0;index" json:"sales_id"`                      // 销售 ID
	SalesName        string     `gorm:"size:50;not null;default:''" json:"sales_name"`                 // 销售姓名
	AppointmentDate  *time.Time `gorm:"type:date;not null;index" json:"appointment_date"`              // 预约日期
	AppointmentTime  string     `gorm:"size:32;not null;default:''" json:"appointment_time"`            // 预约时段
	Address          string     `gorm:"size:500;not null;default:''" json:"address"`                   // 试驾地址
	Latitude         float64    `gorm:"type:decimal(10,7);default:0" json:"latitude"`                  // 纬度
	Longitude        float64    `gorm:"type:decimal(10,7);default:0" json:"longitude"`                 // 经度
	DriveType        string     `gorm:"size:32;not null;default:'test_drive';index" json:"drive_type"`  // test_drive/viewing/delivery
	LicenseStatus    string     `gorm:"size:16;not null;default:'unsubmitted'" json:"license_status"`   // unsubmitted/pending/verified/rejected
	LicenseImages    JSONB      `gorm:"type:jsonb" json:"license_images"`                              // 驾照图片
	Remark           string     `gorm:"size:500;not null;default:''" json:"remark"`                    // 备注
	Status           int        `gorm:"default:0;index" json:"status"`                                 // 0待确认 1确认 2试驾中 3完成 4取消 5未到店
	CancelReason     string     `gorm:"size:500;not null;default:''" json:"cancel_reason"`             // 取消原因
	Result           string     `gorm:"size:32;not null;default:''" json:"result"`                     // satisfied/dissatisfied/undecided/purchased/not_purchased
	ResultRemark     string     `gorm:"type:text" json:"result_remark"`                                // 结果备注
	StartedAt        *time.Time `gorm:"index" json:"started_at"`                                       // 开始时间
	CompletedAt      *time.Time `gorm:"index" json:"completed_at"`                                     // 完成时间
	CanceledAt       *time.Time `gorm:"index" json:"canceled_at"`                                      // 取消时间
}

// TableName 表名（car_ 前缀）
func (CarTestDrive) TableName() string { return "car_test_drives" }
