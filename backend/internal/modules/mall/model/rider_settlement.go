// Package model 同城商城 - 骑手结算表
// 依据需求文档：骑手端后端扩展
// 对标美团/达达骑手月结
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 结算状态常量 ===
const (
	RiderSettlementStatusPending   = 0 // 待结算
	RiderSettlementStatusSettled   = 1 // 已结算
	RiderSettlementStatusWithdrawn = 2 // 已提现
)

// RiderSettlement 骑手结算表
type RiderSettlement struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	RiderID uint `gorm:"index;not null" json:"rider_id"` // 骑手 ID

	// === 结算周期 ===
	Period string `gorm:"size:20;not null;default:'';index" json:"period"` // 结算周期 2026-07

	// === 统计 ===
	TotalOrders int     `gorm:"not null;default:0" json:"total_orders"`             // 订单总数
	TotalAmount float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`   // 订单总额
	TotalFee    float64 `gorm:"type:decimal(12,2);default:0" json:"total_fee"`      // 配送费总额
	TotalTip    float64 `gorm:"type:decimal(12,2);default:0" json:"total_tip"`      // 小费总额
	PlatformFee float64 `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`   // 平台抽成
	NetAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"net_amount"`     // 实发金额

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`                   // 0待结算 1已结算 2已提现
	SettledAt   *time.Time `gorm:"index" json:"settled_at"`                         // 结算时间
	WithdrawnAt *time.Time `gorm:"index" json:"withdrawn_at"`                       // 提现时间

	// === 备注 ===
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核备注
}

// TableName 表名
func (RiderSettlement) TableName() string { return "mall_rider_settlements" }
