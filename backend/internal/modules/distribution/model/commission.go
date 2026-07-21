// Package model 分销合伙人中台数据模型 - 佣金记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 佣金级别常量 ===
const (
	CommissionLevelFirst  = 1 // 一级分销（直接推广）
	CommissionLevelSecond = 2 // 二级分销（下级推广）
)

// === 佣金状态常量 ===
const (
	CommissionStatusPending  = 0 // 待结算
	CommissionStatusSettled  = 1 // 已结算
	CommissionStatusCanceled = 2 // 已取消
)

// Commission 佣金记录模型（distribution_commissions 表）
type Commission struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 关联 ===
	PartnerID uint `gorm:"index;not null" json:"partner_id"` // 合伙人 ID
	OrderID   uint `gorm:"index;not null" json:"order_id"`   // 订单 ID
	ChannelID uint `gorm:"index" json:"channel_id"`          // 渠道 ID（可空）

	// === 金额 ===
	OrderAmount      float64 `gorm:"type:decimal(12,2);default:0" json:"order_amount"`           // 订单金额
	CommissionAmount float64 `gorm:"type:decimal(12,2);default:0" json:"commission_amount"`     // 佣金金额
	CommissionRate   float64 `gorm:"type:decimal(5,4);default:0" json:"commission_rate"`         // 佣金比例（快照）

	// === 级别与状态 ===
	Level     int        `gorm:"default:1;index" json:"level"`       // 1一级 2二级
	Status    int        `gorm:"default:0;index" json:"status"`      // 0待结算 1已结算 2已取消
	SettledAt *time.Time `gorm:"index" json:"settled_at"`            // 结算时间
}

// TableName 表名
func (Commission) TableName() string { return "distribution_commissions" }
