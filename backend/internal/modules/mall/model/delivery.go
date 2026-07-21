// Package model 同城商城 - 配送单表
// 依据需求文档：骑手端后端扩展
// 对标美团/达达/顺丰同城配送
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 配送单状态常量 ===
const (
	DeliveryStatusPending    = 0 // 待接单
	DeliveryStatusAccepted   = 1 // 已接单
	DeliveryStatusArrived    = 2 // 到店
	DeliveryStatusPicked     = 3 // 取货
	DeliveryStatusDelivering = 4 // 配送中
	DeliveryStatusDelivered  = 5 // 已送达
	DeliveryStatusCancelled  = 6 // 已取消
)

// Delivery 配送单表
type Delivery struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	OrderID uint  `gorm:"index;not null;uniqueIndex" json:"order_id"` // 关联 mall_orders
	RiderID *uint `gorm:"index" json:"rider_id"`                      // 认领骑手
	ShopID  uint  `gorm:"index;not null" json:"shop_id"`              // 店铺 ID
	UserID  uint  `gorm:"index;not null" json:"user_id"`              // 收货人用户 ID

	// === 配送单号 ===
	DeliveryNo string `gorm:"size:50;not null;default:'';uniqueIndex" json:"delivery_no"` // 配送单号

	// === 状态 ===
	Status int `gorm:"default:0;index" json:"status"` // 0待接单 1已接单 2到店 3取货 4配送中 5已送达 6已取消

	// === 取货地址 ===
	PickupAddress string  `gorm:"size:500;not null;default:''" json:"pickup_address"` // 取货地址
	PickupLat     float64 `gorm:"type:decimal(10,7);default:0" json:"pickup_lat"`     // 取货纬度
	PickupLng     float64 `gorm:"type:decimal(10,7);default:0" json:"pickup_lng"`     // 取货经度

	// === 送达地址 ===
	DeliveryAddress string  `gorm:"size:500;not null;default:''" json:"delivery_address"` // 送达地址
	DeliveryLat     float64 `gorm:"type:decimal(10,7);default:0" json:"delivery_lat"`     // 送达纬度
	DeliveryLng     float64 `gorm:"type:decimal(10,7);default:0" json:"delivery_lng"`     // 送达经度

	// === 距离与金额 ===
	Distance    float64 `gorm:"type:decimal(10,2);default:0" json:"distance"`     // 公里
	DeliveryFee float64 `gorm:"type:decimal(12,2);default:0" json:"delivery_fee"` // 配送费
	Tip         float64 `gorm:"type:decimal(12,2);default:0" json:"tip"`          // 小费

	// === 时间节点 ===
	AcceptedAt  *time.Time `gorm:"index" json:"accepted_at"`  // 接单时间
	PickedAt    *time.Time `gorm:"index" json:"picked_at"`    // 取货时间
	DeliveredAt *time.Time `gorm:"index" json:"delivered_at"` // 送达时间

	// === 备注 ===
	CancelReason string `gorm:"size:500;not null;default:''" json:"cancel_reason"` // 取消原因
}

// TableName 表名
func (Delivery) TableName() string { return "mall_deliveries" }
