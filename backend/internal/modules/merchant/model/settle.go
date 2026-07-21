// Package model 商户中台 - 商户结算表（merchant_settles）
// 对标美团/有赞商户结算系统
// 按月生成结算单：总金额 / 平台佣金 / 商户应得
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Settle 商户结算表
type Settle struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 关联 ===
	ShopID uint   `gorm:"index;not null" json:"shop_id"` // 所属商户 ID
	Period string `gorm:"size:20;not null;index" json:"period"` // 结算周期 YYYY-MM

	// === 金额（DECIMAL(12,2)） ===
	TotalAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`   // 总金额
	PlatformFee  float64 `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`   // 平台佣金
	ShopAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"shop_amount"`     // 商户应得

	// === 状态 ===
	Status    int        `gorm:"default:0;index" json:"status"`     // 0待结算 1已结算 2已提现 3已撤销
	SettledAt *time.Time `gorm:"index" json:"settled_at"`           // 结算时间
}

// TableName 表名
func (Settle) TableName() string { return "merchant_settles" }
