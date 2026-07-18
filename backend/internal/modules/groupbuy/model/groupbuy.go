// Package model 团购优惠券数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// GroupBuy 团购商品模型
type GroupBuy struct {
	database.RegionBaseModel
	Title         string     `gorm:"size:200;not null" json:"title"`            // 标题
	Description   string     `gorm:"type:text" json:"description"`             // 描述
	Cover         string     `gorm:"size:255" json:"cover"`                    // 封面图
	OriginalPrice float64    `gorm:"type:decimal(10,2);default:0" json:"original_price"` // 原价
	GroupBuyPrice float64    `gorm:"type:decimal(10,2);default:0" json:"groupbuy_price"` // 团购价
	Stock         int        `gorm:"default:0" json:"stock"`                   // 库存
	SoldCount     int        `gorm:"default:0" json:"sold_count"`              // 已售数量
	PerLimit      int        `gorm:"default:1" json:"per_limit"`               // 每人限购
	StartTime     *time.Time `gorm:"index" json:"start_time"`                  // 开始时间
	EndTime       *time.Time `gorm:"index" json:"end_time"`                    // 结束时间
	Status        int        `gorm:"default:0;index" json:"status"`            // 状态 0下架 1上架 2已结束
	AuditStatus   int        `gorm:"default:0;index" json:"audit_status"`      // 审核状态 0待审核 1通过 2拒绝
	IsRecommend   int        `gorm:"default:0" json:"is_recommend"`            // 是否推荐
	Sort          int        `gorm:"default:0" json:"sort"`                    // 排序
	ShopID        uint       `gorm:"index;default:0" json:"shop_id"`           // 商家ID，0表示平台团购
	UserID        uint       `gorm:"index" json:"user_id"`                     // 发布者ID
}

// TableName 表名
func (GroupBuy) TableName() string {
	return "groupbuy"
}

// Coupon 优惠券模型
type Coupon struct {
	database.RegionBaseModel
	Name          string     `gorm:"size:100;not null" json:"name"`            // 名称
	Type          int        `gorm:"default:1" json:"type"`                    // 类型 1满减 2折扣 3代金券
	Value         float64    `gorm:"type:decimal(10,2);default:0" json:"value"` // 面值
	MinAmount     float64    `gorm:"type:decimal(10,2);default:0" json:"min_amount"` // 最低消费
	Scope         int        `gorm:"default:0" json:"scope"`                   // 适用范围 0全部 1指定商家 2指定分类
	ScopeID       uint       `gorm:"default:0" json:"scope_id"`                // 关联ID
	TotalCount    int        `gorm:"default:0" json:"total_count"`             // 发放总量
	ReceivedCount int        `gorm:"default:0" json:"received_count"`          // 已领取数
	UsedCount     int        `gorm:"default:0" json:"used_count"`              // 已使用数
	PerLimit      int        `gorm:"default:1" json:"per_limit"`               // 每人限领
	StartTime     *time.Time `gorm:"index" json:"start_time"`                  // 开始时间
	EndTime       *time.Time `gorm:"index" json:"end_time"`                    // 结束时间
	ValidityType  int        `gorm:"default:0" json:"validity_type"`           // 有效期类型 0固定时间 1领取后N天
	ValidDays     int        `gorm:"default:0" json:"valid_days"`              // 有效天数
	Status        int        `gorm:"default:1;index" json:"status"`            // 状态 0禁用 1启用
}

// TableName 表名
func (Coupon) TableName() string {
	return "coupon"
}

// UserCoupon 用户优惠券模型
type UserCoupon struct {
	database.RegionBaseModel
	UserID     uint       `gorm:"index;not null" json:"user_id"`             // 用户ID
	CouponID   uint       `gorm:"index;not null" json:"coupon_id"`           // 优惠券ID
	Status     int        `gorm:"default:0;index" json:"status"`             // 状态 0未使用 1已使用 2已过期
	ReceivedAt time.Time  `json:"received_at"`                               // 领取时间
	UsedAt     *time.Time `json:"used_at"`                                   // 使用时间
	ExpireAt   *time.Time `gorm:"index" json:"expire_at"`                    // 过期时间
	OrderID    uint       `gorm:"default:0" json:"order_id"`                 // 关联订单ID
}

// TableName 表名
func (UserCoupon) TableName() string {
	return "user_coupon"
}

// GroupBuyOrder 团购订单模型
type GroupBuyOrder struct {
	database.RegionBaseModel
	OrderNo        string     `gorm:"size:64;uniqueIndex;not null" json:"order_no"` // 订单号
	UserID         uint       `gorm:"index;not null" json:"user_id"`                // 用户ID
	GroupBuyID     uint       `gorm:"index;not null" json:"groupbuy_id"`            // 团购ID
	Quantity       int        `gorm:"default:1" json:"quantity"`                    // 购买数量
	UnitPrice      float64    `gorm:"type:decimal(10,2)" json:"unit_price"`         // 单价
	TotalPrice     float64    `gorm:"type:decimal(10,2)" json:"total_price"`        // 总价
	CouponID       uint       `gorm:"default:0" json:"coupon_id"`                   // 优惠券ID
	UserCouponID   uint       `gorm:"default:0" json:"user_coupon_id"`              // 用户优惠券ID
	DiscountAmount float64    `gorm:"type:decimal(10,2);default:0" json:"discount_amount"` // 优惠金额
	PayAmount      float64    `gorm:"type:decimal(10,2)" json:"pay_amount"`         // 实付金额
	PayStatus      int        `gorm:"default:0;index" json:"pay_status"`            // 支付状态 0待支付 1已支付 2已退款
	PayAt          *time.Time `json:"pay_at"`                                       // 支付时间
	VerifyStatus   int        `gorm:"default:0;index" json:"verify_status"`         // 核销状态 0未核销 1已核销
	VerifyCode     string     `gorm:"size:32;index" json:"verify_code"`             // 核销码
	VerifyAt       *time.Time `json:"verify_at"`                                    // 核销时间
	Status         int        `gorm:"default:0;index" json:"status"`                // 订单状态 0待支付 1已支付 2已核销 3已取消 4已退款
}

// TableName 表名
func (GroupBuyOrder) TableName() string {
	return "groupbuy_order"
}
