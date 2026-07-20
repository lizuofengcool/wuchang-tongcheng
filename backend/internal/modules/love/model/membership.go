// Package model love 相亲交友数据模型 - 会员订阅表 LoveMembership
// 对标陌陌/Soul：订阅记录（按月/季/年）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveMembership 会员订阅表
// 用户开通会员后生成订阅记录，含支付/续费/取消/退款全生命周期
type LoveMembership struct {
	database.RegionBaseModel

	SubNo string `gorm:"size:64;not null;uniqueIndex" json:"sub_no"` // 订阅单号
	UserID uint `gorm:"index;not null" json:"user_id"`
	LoveID uint `gorm:"index;not null" json:"love_id"`

	// 等级
	LevelCode string `gorm:"size:32;not null;index" json:"level_code"`
	LevelName string `gorm:"size:64;not null" json:"level_name"`
	Level     int    `gorm:"not null;default:0" json:"level"`

	// 套餐
	Plan   string `gorm:"size:32;not null;default:'monthly';index" json:"plan"` // monthly/quarterly/yearly
	Period int    `gorm:"not null;default:1" json:"period"`                     // 周期数

	// 时间
	StartAt time.Time `gorm:"not null;index" json:"start_at"`
	EndAt   time.Time `gorm:"not null;index" json:"end_at"`

	// 价格
	Price         float64 `gorm:"type:decimal(12,2);not null;default:0" json:"price"`
	OriginalPrice float64 `gorm:"type:decimal(12,2);not null;default:0" json:"original_price"`
	Discount      float64 `gorm:"type:decimal(5,2);not null;default:0" json:"discount"`
	PayAmount     float64 `gorm:"type:decimal(12,2);not null;default:0" json:"pay_amount"`

	// 支付
	PayMethod  string `gorm:"size:32;not null;default:''" json:"pay_method"` // wechat/alipay/credits
	PayOrderNo string `gorm:"size:64;not null;default:''" json:"pay_order_no"`
	PayAt      *time.Time `json:"pay_at"`

	// 自动续费
	AutoRenew   bool `gorm:"not null;default:false;index" json:"auto_renew"`
	RenewCount  int  `gorm:"not null;default:0" json:"renew_count"`

	// 权益快照（订阅时的权益 JSONB）
	PerksSnapshot JSONB `gorm:"type:jsonb" json:"perks_snapshot"`

	// 状态/取消/退款
	Status        int        `gorm:"not null;default:0;index" json:"status"` // 0未支付 1有效 2已取消 3已退款 4已过期
	CancelAt      *time.Time `json:"cancel_at"`
	CancelReason  string     `gorm:"size:255;not null;default:''" json:"cancel_reason"`
	RefundAmount  float64    `gorm:"type:decimal(12,2);not null;default:0" json:"refund_amount"`
	RefundAt      *time.Time `json:"refund_at"`
	RefundReason  string     `gorm:"size:255;not null;default:''" json:"refund_reason"`

	// 来源
	Source string `gorm:"size:32;not null;default:'self'" json:"source"` // self/gift/admin/promo
	Remark string `gorm:"size:255;not null;default:''" json:"remark"`
}

// TableName 表名
func (LoveMembership) TableName() string { return "love_memberships" }
