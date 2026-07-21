// Package model 同城商城 - 物流记录
// 依据需求文档 1.10：4 维数据隔离（region_id + order_id 订单 + shop_id 商家）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Logistics 物流表
type Logistics struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	OrderID   uint   `gorm:"index;not null" json:"order_id"`                // 订单 ID
	OrderNo   string `gorm:"size:32;not null;default:'';index" json:"order_no"` // 订单号（冗余）
	UserID    uint   `gorm:"index;not null" json:"user_id"`                 // 买家 ID
	ShopID    uint   `gorm:"index;not null" json:"shop_id"`                 // 店铺 ID

	// === 物流信息 ===
	Company       string `gorm:"size:64;not null;default:'';index" json:"company"`     // 物流公司
	CompanyCode   string `gorm:"size:32;not null;default:''" json:"company_code"`      // 物流公司编码（SF/YTO/ZTO...）
	TrackingNo    string `gorm:"size:64;not null;default:'';index" json:"tracking_no"` // 物流单号
	CourierName   string `gorm:"size:50;not null;default:''" json:"courier_name"`      // 快递员姓名
	CourierPhone  string `gorm:"size:32;not null;default:''" json:"courier_phone"`     // 快递员电话

	// === 发货/收货地址 ===
	SenderName    string `gorm:"size:50;not null;default:''" json:"sender_name"`       // 发件人
	SenderPhone   string `gorm:"size:32;not null;default:''" json:"sender_phone"`      // 发件电话
	SenderAddress string `gorm:"size:500;not null;default:''" json:"sender_address"`   // 发件地址
	ReceiverName  string `gorm:"size:50;not null;default:''" json:"receiver_name"`     // 收件人
	ReceiverPhone string `gorm:"size:32;not null;default:''" json:"receiver_phone"`    // 收件电话
	ReceiverAddress string `gorm:"size:500;not null;default:''" json:"receiver_address"` // 收件地址

	// === 状态 ===
	Status    int        `gorm:"default:0;index" json:"status"`         // 0待发货 1已发货 2运输中 3已派送 4已签收 5已退回
	ShippedAt *time.Time `gorm:"index" json:"shipped_at"`               // 发货时间
	InTransitAt *time.Time `gorm:"index" json:"in_transit_at"`           // 运输中时间
	DeliveredAt *time.Time `gorm:"index" json:"delivered_at"`             // 派送时间
	ReceivedAt *time.Time `gorm:"index" json:"received_at"`              // 签收时间
	ReturnedAt *time.Time `gorm:"index" json:"returned_at"`              // 退回时间

	// === 物流轨迹（JSONB） ===
	Traces   JSONB `gorm:"type:jsonb" json:"traces"`     // 物流轨迹 [{time,desc,status}]

	// === 重量与体积 ===
	Weight float64 `gorm:"type:decimal(10,3);default:0" json:"weight"` // 重量 kg
	Volume float64 `gorm:"type:decimal(10,3);default:0" json:"volume"` // 体积 m³
	Pieces int     `gorm:"not null;default:1" json:"pieces"`           // 包裹件数

	// === 运费 ===
	Freight       float64 `gorm:"type:decimal(12,2);default:0" json:"freight"`       // 运费
	InsuredFee    float64 `gorm:"type:decimal(12,2);default:0" json:"insured_fee"`    // 保价费
	CodFee        float64 `gorm:"type:decimal(12,2);default:0" json:"cod_fee"`        // 货到付款手续费
}

// TableName 表名
func (Logistics) TableName() string { return "mall_logistics" }
