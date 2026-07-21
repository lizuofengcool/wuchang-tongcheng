// Package model 同城商城 - 支付记录
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 买家 + order_id 订单）
// 对接 pay 中台 / 微信 / 支付宝
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Payment 支付记录表
type Payment struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	OrderID   uint `gorm:"index;not null" json:"order_id"`     // 订单 ID
	OrderNo   string `gorm:"size:32;not null;default:'';index" json:"order_no"` // 订单号（冗余）
	UserID    uint `gorm:"index;not null" json:"user_id"`      // 买家 ID
	ShopID    uint `gorm:"index;not null" json:"shop_id"`      // 店铺 ID

	// === 支付单信息 ===
	PaymentNo    string `gorm:"size:64;not null;default:'';index" json:"payment_no"`     // 平台支付单号（业务唯一）
	TradeNo      string `gorm:"size:64;not null;default:'';index" json:"trade_no"`       // 第三方交易号（微信/支付宝）
	OutTradeNo   string `gorm:"size:64;not null;default:''" json:"out_trade_no"`         // 商户订单号（传给第三方）

	// === 金额（decimal(12,2)） ===
	Amount       float64 `gorm:"type:decimal(12,2);default:0" json:"amount"`       // 支付金额
	RefundAmount float64 `gorm:"type:decimal(12,2);default:0" json:"refund_amount"` // 已退款金额

	// === 支付方式 ===
	Method    string `gorm:"size:32;not null;default:'wechat';index" json:"method"` // wechat/alipay/balance/cod/bankcard
	Channel   string `gorm:"size:32;not null;default:''" json:"channel"`            // 渠道（mp/h5/pc/miniapp）

	// === 状态 ===
	Status    int    `gorm:"default:0;index" json:"status"`           // 0待支付 1成功 2失败 3已退款 4已关闭
	ErrorCode string `gorm:"size:32;not null;default:''" json:"error_code"` // 失败错误码
	ErrorMsg  string `gorm:"size:500;not null;default:''" json:"error_msg"`  // 失败错误消息

	// === 时间 ===
	PaidAt     *time.Time `gorm:"index" json:"paid_at"`      // 支付成功时间
	ExpiredAt  *time.Time `gorm:"index" json:"expired_at"`   // 支付单过期时间
	ClosedAt   *time.Time `gorm:"index" json:"closed_at"`    // 关闭时间
	RefundedAt *time.Time `gorm:"index" json:"refunded_at"`  // 退款完成时间

	// === 第三方原始响应 ===
	RawResponse JSONB `gorm:"type:jsonb" json:"raw_response"` // 第三方回调原始数据

	// === 客户端 ===
	ClientIP  string `gorm:"size:64;not null;default:''" json:"client_ip"`   // 客户端 IP
	UserAgent string `gorm:"size:255;not null;default:''" json:"user_agent"` // 客户端 UA
}

// TableName 表名
func (Payment) TableName() string { return "mall_payments" }
