// Package model 优惠券表（对标大众点评/美团）
// 满减/折扣/代金券/礼品券
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Coupon 优惠券表
type Dh114Coupon struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	CouponNo   string `gorm:"size:64;not null;uniqueIndex:uniq_dh114_coupons_no" json:"coupon_no"` // 优惠券单号
	Dh114ID    uint   `gorm:"not null;index" json:"dh114_id"`                                       // 商户 ID
	BusinessID uint   `gorm:"not null;default:0;index" json:"business_id"`                          // 商户详情 ID

	// === 基本信息 ===
	Title       string `gorm:"size:200;not null" json:"title"`                  // 优惠券标题
	Description string `gorm:"type:text" json:"description"`                    // 描述
	CoverImage  string `gorm:"size:255;not null;default:''" json:"cover_image"` // 封面图

	// === 类型与面值 ===
	CouponType string  `gorm:"size:32;not null;default:'discount';index" json:"coupon_type"` // discount/full_reduction/cash/gift
	FaceValue  float64 `gorm:"type:decimal(12,2);default:0" json:"face_value"`               // 面值
	Threshold  float64 `gorm:"type:decimal(12,2);default:0" json:"threshold"`                 // 使用门槛（满减券的门槛）
	Discount   float64 `gorm:"type:decimal(3,2);default:0" json:"discount"`                   // 折扣率（如 0.85 表示 85 折）
	MaxDiscount float64 `gorm:"type:decimal(12,2);default:0" json:"max_discount"`             // 最大优惠金额

	// === 库存 ===
	TotalCount   int `gorm:"not null;default:0" json:"total_count"`     // 总数量
	IssuedCount  int `gorm:"not null;default:0;index" json:"issued_count"` // 已领取数
	UsedCount    int `gorm:"not null;default:0;index" json:"used_count"` // 已使用数
	PerUserLimit int `gorm:"not null;default:1" json:"per_user_limit"`   // 每人限领

	// === 时间 ===
	StartTime  *time.Time `gorm:"index" json:"start_time"`   // 领取开始时间
	EndTime    *time.Time `gorm:"index" json:"end_time"`     // 领取结束时间
	ValidStart *time.Time `gorm:"type:date;index" json:"valid_start"` // 使用开始日期
	ValidEnd   *time.Time `gorm:"type:date" json:"valid_end"`         // 使用结束日期
	ValidDays  int        `gorm:"not null;default:0" json:"valid_days"` // 领取后有效天数（0 表示按日期）

	// === 使用规则 ===
	UseInstructions JSONB `gorm:"type:jsonb" json:"use_instructions"` // 使用规则 JSON
	UseThreshold    float64 `gorm:"type:decimal(12,2);default:0" json:"use_threshold"` // 使用门槛

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`                       // 0草稿 1已发布 2已抢完 3已下架 4已过期
	AuditStatus int    `gorm:"default:0;index" json:"audit_status"`                 // 审核状态
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"`   // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`                       // 发布时间

	// === 运营 ===
	Featured bool `gorm:"not null;default:false;index" json:"featured"` // 精选推荐
}

// TableName 表名
func (Dh114Coupon) TableName() string { return "dh114_coupons" }
