// Package model 分销合伙人中台数据模型
// 依据架构设计 4.5：分销合伙人中台（distribution）
// 职责：二级分销/城市分站分成/推广渠道统计/佣金自动结算/付费合伙人等级
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 合伙人等级常量 ===
const (
	PartnerLevelNormal   = 1 // 普通合伙人
	PartnerLevelSenior   = 2 // 高级合伙人
	PartnerLevelCity     = 3 // 城市合伙人
)

// === 合伙人状态常量 ===
const (
	PartnerStatusPending  = 0 // 待审核
	PartnerStatusActive   = 1 // 正常
	PartnerStatusFrozen   = 2 // 冻结
	PartnerStatusRejected = 3 // 拒绝
	PartnerStatusQuit     = 4 // 退出
)

// Partner 分销合伙人模型（distribution_partners 表）
type Partner struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	UserID   uint `gorm:"index;not null" json:"user_id"`    // 关联用户 ID
	ParentID uint `gorm:"index;default:0" json:"parent_id"` // 上级合伙人 ID（0=无上级）

	// === 等级与佣金 ===
	Level           int     `gorm:"default:1;index" json:"level"`                              // 1普通 2高级 3城市合伙人
	CommissionRate  float64 `gorm:"type:decimal(5,4);default:0" json:"commission_rate"`       // 佣金比例 0-1
	TotalCommission float64 `gorm:"type:decimal(12,2);default:0" json:"total_commission"`     // 累计佣金
	SettledCommission float64 `gorm:"type:decimal(12,2);default:0" json:"settled_commission"` // 已结算佣金
	PendingCommission  float64 `gorm:"type:decimal(12,2);default:0" json:"pending_commission"` // 待结算佣金

	// === 状态 ===
	Status   int        `gorm:"default:0;index" json:"status"` // 0待审核 1正常 2冻结 3拒绝 4退出
	JoinedAt *time.Time `gorm:"index" json:"joined_at"`         // 加入时间
}

// TableName 表名
func (Partner) TableName() string { return "distribution_partners" }
