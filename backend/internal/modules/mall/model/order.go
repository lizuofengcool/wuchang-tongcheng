// Package model 同城商城 - 订单主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 买家 + shop_id 商家）
// 对标淘宝/京东/拼多多订单
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Order 订单主表
type Order struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 订单基础 ===
	OrderNo string `gorm:"size:32;not null;default:'';index" json:"order_no"` // 订单号（业务唯一）
	UserID  uint   `gorm:"index;not null" json:"user_id"`                      // 买家用户 ID
	ShopID  uint   `gorm:"index;not null" json:"shop_id"`                      // 店铺 ID
	ShopName string `gorm:"size:200;not null;default:''" json:"shop_name"`     // 店铺名（冗余）

	// === 买家信息（冗余） ===
	BuyerName   string `gorm:"size:50;not null;default:''" json:"buyer_name"`
	BuyerPhone  string `gorm:"size:32;not null;default:''" json:"buyer_phone"`
	BuyerAvatar string `gorm:"size:255;not null;default:''" json:"buyer_avatar"`

	// === 收货地址快照 ===
	AddressID    uint   `gorm:"not null;default:0;index" json:"address_id"`     // 收货地址 ID
	ReceiverName string `gorm:"size:50;not null;default:''" json:"receiver_name"`   // 收货人
	ReceiverPhone string `gorm:"size:32;not null;default:''" json:"receiver_phone"` // 收货电话
	Province     string `gorm:"size:64;not null;default:''" json:"province"`     // 省
	City         string `gorm:"size:64;not null;default:''" json:"city"`         // 市
	District     string `gorm:"size:64;not null;default:''" json:"district"`     // 区
	Address      string `gorm:"size:500;not null;default:''" json:"address"`     // 详细地址
	ZipCode      string `gorm:"size:20;not null;default:''" json:"zip_code"`     // 邮编

	// === 金额（decimal(12,2)） ===
	TotalAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`    // 商品总金额
	ShippingFee   float64 `gorm:"type:decimal(12,2);default:0" json:"shipping_fee"`    // 运费
	DiscountAmount float64 `gorm:"type:decimal(12,2);default:0" json:"discount_amount"` // 优惠金额
	TaxAmount     float64 `gorm:"type:decimal(12,2);default:0" json:"tax_amount"`      // 税费
	PayAmount     float64 `gorm:"type:decimal(12,2);default:0;index" json:"pay_amount"` // 实付金额

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`         // 0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭
	StatusText  string `gorm:"size:32;not null;default:''" json:"status_text"` // 状态文本（冗余）
	Remark      string `gorm:"size:500;not null;default:''" json:"remark"`     // 买家留言
	SellerRemark string `gorm:"size:500;not null;default:''" json:"seller_remark"` // 卖家备注
	AdminRemark string `gorm:"size:500;not null;default:''" json:"admin_remark"`   // 平台备注

	// === 时间节点 ===
	PaidAt       *time.Time `gorm:"index" json:"paid_at"`        // 支付时间
	ShippedAt    *time.Time `gorm:"index" json:"shipped_at"`     // 发货时间
	ReceivedAt   *time.Time `gorm:"index" json:"received_at"`    // 收货时间
	CompletedAt  *time.Time `gorm:"index" json:"completed_at"`   // 完成时间
	CancelledAt  *time.Time `gorm:"index" json:"cancelled_at"`   // 取消时间
	AutoCloseAt  *time.Time `gorm:"index" json:"auto_close_at"`  // 自动关闭时间（待付款 15 分钟）
	AutoConfirmAt *time.Time `gorm:"index" json:"auto_confirm_at"` // 自动确认收货时间（发货后 7 天）
	AutoReviewAt  *time.Time `gorm:"index" json:"auto_review_at"`  // 自动评价时间（收货后 7 天）

	// === 支付信息 ===
	PaymentMethod string `gorm:"size:32;not null;default:''" json:"payment_method"` // 支付方式 wechat/alipay/balance/cod/bankcard
	PaymentNo     string `gorm:"size:64;not null;default:'';index" json:"payment_no"` // 第三方支付流水号

	// === 来源 ===
	Source     string `gorm:"size:32;not null;default:'app';index" json:"source"`     // 来源 app/web/miniapp/admin
	ClientIP   string `gorm:"size:64;not null;default:''" json:"client_ip"`           // 下单 IP
	UserAgent  string `gorm:"size:255;not null;default:''" json:"user_agent"`         // 下单 UA

	// === 优惠券 ===
	CouponID    uint   `gorm:"not null;default:0;index" json:"coupon_id"`    // 使用的优惠券 ID
	CouponName  string `gorm:"size:128;not null;default:''" json:"coupon_name"` // 优惠券名（冗余）

	// === 评价标记 ===
	HasReview   bool `gorm:"default:false;index" json:"has_review"`   // 是否已评价
	HasSellerReview bool `gorm:"default:false" json:"has_seller_review"` // 是否已评价卖家

	// === 物流 ===
	LogisticsCompany string `gorm:"size:64;not null;default:''" json:"logistics_company"` // 物流公司
	LogisticsNo      string `gorm:"size:64;not null;default:'';index" json:"logistics_no"` // 物流单号

	// === 风控 ===
	RiskScore int `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100
}

// TableName 表名
func (Order) TableName() string { return "mall_orders" }
