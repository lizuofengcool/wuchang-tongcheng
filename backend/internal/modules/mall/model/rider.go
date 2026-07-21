// Package model 同城商城 - 骑手认证表
// 依据需求文档：骑手端后端扩展
// 对标美团/达达/顺丰同城骑手
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 骑手状态常量 ===
const (
	RiderStatusPending  = 0 // 待审核
	RiderStatusApproved = 1 // 通过
	RiderStatusRejected = 2 // 拒绝
	RiderStatusFrozen   = 3 // 冻结
)

// === 骑手在线状态常量 ===
const (
	RiderOffline    = 0 // 下线
	RiderOnline     = 1 // 在线
	RiderDelivering = 2 // 配送中
)

// === 骑手车辆类型常量 ===
const (
	VehicleTypeElectric = "electric" // 电动车
	VehicleTypeMotor    = "motor"    // 摩托车
	VehicleTypeBicycle  = "bicycle"  // 自行车
	VehicleTypeCar      = "car"      // 汽车
)

// Rider 骑手认证表
type Rider struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	UserID uint  `gorm:"index;not null;uniqueIndex" json:"user_id"` // 关联 users
	ShopID *uint `gorm:"index" json:"shop_id"`                      // 可选，绑定店铺

	// === 基本信息 ===
	RealName string `gorm:"size:50;not null;default:''" json:"real_name"` // 真实姓名
	Phone    string `gorm:"size:20;not null;default:''" json:"phone"`     // 联系电话
	IDCard   string `gorm:"size:18;not null;default:''" json:"id_card"`   // 身份证号
	Avatar   string `gorm:"size:255;not null;default:''" json:"avatar"`   // 头像

	// === 车辆信息 ===
	VehicleType  string `gorm:"size:20;not null;default:''" json:"vehicle_type"`  // 电动车/摩托车/自行车/汽车
	VehiclePlate string `gorm:"size:20;not null;default:''" json:"vehicle_plate"` // 车牌号
	LicenseURL   string `gorm:"size:255;not null;default:''" json:"license_url"`  // 驾驶证

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`                   // 0待审核 1通过 2拒绝 3冻结
	CreditScore int    `gorm:"not null;default:100" json:"credit_score"`        // 信用分
	Level       int    `gorm:"not null;default:1" json:"level"`                 // 等级
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核理由

	// === 统计 ===
	TotalOrders   int     `gorm:"not null;default:0" json:"total_orders"`              // 累计订单数
	TotalEarnings float64 `gorm:"type:decimal(12,2);default:0" json:"total_earnings"`  // 累计收入

	// === 在线状态 ===
	OnlineStatus int `gorm:"default:0;index" json:"online_status"` // 0下线 1在线 2配送中
}

// TableName 表名
func (Rider) TableName() string { return "mall_riders" }
