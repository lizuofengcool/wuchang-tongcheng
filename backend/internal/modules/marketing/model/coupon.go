// Package model 营销活动中台 - 优惠券模型（coupon 子域）
// 依据架构设计 4.6：满减/折扣/兑换券
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 优惠券类型常量 ===
const (
	CouponTypeDiscount = "discount"  // 折扣券
	CouponTypeReduce   = "reduce"    // 满减券
	CouponTypeExchange = "exchange"  // 兑换券
)

// === 优惠券状态常量 ===
const (
	CouponStatusDisabled = 0 // 禁用
	CouponStatusActive   = 1 // 进行中
	CouponStatusDraft    = 2 // 草稿
	CouponStatusOffline  = 3 // 已下架
	CouponStatusExpired  = 4 // 已过期
	CouponStatusSoldOut  = 5 // 已抢完
)

// Coupon 优惠券模型（coupons 表）
type Coupon struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	Title         string     `gorm:"size:100;not null;default:'';index" json:"title"`           // 优惠券标题
	Type          string     `gorm:"size:20;not null;default:'reduce';index" json:"type"`       // discount/reduce/exchange
	Amount        float64    `gorm:"type:decimal(12,2);not null;default:0" json:"amount"`       // 面值/折扣率（折扣券 0.01-0.99）
	Threshold     float64    `gorm:"type:decimal(12,2);not null;default:0" json:"threshold"`     // 使用门槛（满 N 元可用）
	TotalCount    int        `gorm:"not null;default:0" json:"total_count"`                      // 发放总量（0=不限）
	ReceivedCount int        `gorm:"not null;default:0" json:"received_count"`                   // 已领取数
	StartAt       *time.Time `gorm:"index" json:"start_at"`                                      // 领取开始时间
	EndAt         *time.Time `gorm:"index" json:"end_at"`                                        // 领取结束时间
	Status        int        `gorm:"not null;default:1;index" json:"status"`                     // 0禁用 1进行中 2草稿 3已下架 4已过期 5已抢完
}

// TableName 表名
func (Coupon) TableName() string { return "coupons" }
