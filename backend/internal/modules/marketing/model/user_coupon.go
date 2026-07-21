// Package model 营销活动中台 - 用户优惠券模型（coupon 子域）
// 依据架构设计 4.6：用户领取/使用/过期
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 用户优惠券状态常量 ===
const (
	UserCouponStatusUnused  = "unused"  // 未使用
	UserCouponStatusUsed    = "used"    // 已使用
	UserCouponStatusExpired = "expired" // 已过期
)

// === 用户优惠券来源常量 ===
const (
	UserCouponSourceReceive  = "receive"  // 主动领取
	UserCouponSourceGift     = "gift"     // 系统赠送
	UserCouponSourceActivity = "activity" // 活动奖励
	UserCouponSourceNewUser  = "new_user" // 新人礼包
)

// UserCoupon 用户优惠券模型（user_coupons 表）
type UserCoupon struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at（用户隔离，无 region_id）

	UserID  uint       `gorm:"not null;index:idx_user_coupons_user" json:"user_id"`        // 用户 ID
	CouponID uint      `gorm:"not null;index:idx_user_coupons_coupon" json:"coupon_id"`    // 优惠券 ID
	Status  string     `gorm:"size:20;not null;default:'unused';index" json:"status"`      // unused/used/expired
	Source  string     `gorm:"size:32;not null;default:'receive';index" json:"source"`     // 领取来源
	UsedAt  *time.Time `json:"used_at"`                                                    // 使用时间
	OrderID uint       `gorm:"not null;default:0;index" json:"order_id"`                   // 使用的订单 ID
}

// TableName 表名
func (UserCoupon) TableName() string { return "user_coupons" }
