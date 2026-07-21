// Package model 分销合伙人中台数据模型 - 推广渠道
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Channel 推广渠道模型（distribution_channels 表）
type Channel struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 关联 ===
	PartnerID uint   `gorm:"index;not null" json:"partner_id"`           // 所属合伙人 ID
	Code      string `gorm:"size:50;uniqueIndex;not null" json:"code"`   // 推广码（唯一）
	Name      string `gorm:"size:100" json:"name"`                        // 渠道名称

	// === 统计 ===
	ClickCount       int     `gorm:"default:0" json:"click_count"`                       // 点击数
	RegisterCount    int     `gorm:"default:0" json:"register_count"`                   // 注册数
	OrderCount       int     `gorm:"default:0" json:"order_count"`                      // 订单数
	CommissionAmount float64 `gorm:"type:decimal(12,2);default:0" json:"commission_amount"` // 累计佣金
}

// TableName 表名
func (Channel) TableName() string { return "distribution_channels" }
