// Package model 订单主表 + 订单子项（对标闲鱼/转转）
// 担保交易 11 状态机：0待支付/1已支付待发货/2已发货/3待收货/4已完成/
//                    5已取消/6退款中/7退款完成/8申诉中/9申诉完成/10已关闭
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 订单状态常量（11 状态机，对标闲鱼担保交易） ===
const (
	OrderStatusPending         = 0  // 待支付
	OrderStatusPaid             = 1  // 已支付待发货
	OrderStatusShipped          = 2  // 已发货
	OrderStatusDelivered        = 3  // 待收货
	OrderStatusCompleted        = 4  // 已完成
	OrderStatusCanceled         = 5  // 已取消
	OrderStatusRefunding        = 6  // 退款中
	OrderStatusRefunded         = 7  // 退款完成
	OrderStatusDispute          = 8  // 申诉中
	OrderStatusDisputeClosed    = 9  // 申诉完成
	OrderStatusClosed           = 10 // 已关闭
)

// === 支付方式常量 ===
const (
	PayMethodWechat  = "wechat"  // 微信支付
	PayMethodAlipay  = "alipay"  // 支付宝
	PayMethodBalance = "balance" // 余额支付
	PayMethodInstallment = "installment" // 分期付款
)

// === 物流方式常量 ===
const (
	OrderDeliveryFace    = "face"    // 当面交易
	OrderDeliverySelf    = "self"    // 自提
	OrderDeliveryExpress = "express" // 快递
)

// ErshouOrder 订单主表（对标闲鱼/转转）
type ErshouOrder struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	OrderNo       string     `gorm:"size:64;not null;uniqueIndex" json:"order_no"`              // 订单号（业务唯一）
	BuyerID       uint       `gorm:"not null;index:idx_order_buyer" json:"buyer_id"`              // 买家ID
	SellerID      uint       `gorm:"not null;index:idx_order_seller" json:"seller_id"`             // 卖家ID
	ShopID        uint       `gorm:"index" json:"shop_id"`                                         // 店铺ID（个人闲置为 0）
	TotalAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"total_amount"`            // 订单总金额（商品+运费）
	ItemAmount    float64    `gorm:"type:decimal(12,2);default:0" json:"item_amount"`             // 商品总金额
	DeliveryFee   float64    `gorm:"type:decimal(12,2);default:0" json:"delivery_fee"`            // 运费
	DiscountAmount float64   `gorm:"type:decimal(12,2);default:0" json:"discount_amount"`          // 优惠金额
	Status        int        `gorm:"default:0;index:idx_order_buyer;index:idx_order_seller" json:"status"` // 11状态机
	PayMethod     string     `gorm:"size:32;default:'wechat'" json:"pay_method"`                  // 支付方式 wechat/alipay/balance/installment
	PayTradeNo    string     `gorm:"size:128;index" json:"pay_trade_no"`                          // 第三方支付流水号
	DeliveryMethod string    `gorm:"size:32;default:'face'" json:"delivery_method"`               // 物流方式 face/self/express
	Remark        string     `gorm:"size:500" json:"remark"`                                       // 买家备注
	ContactName   string     `gorm:"size:50" json:"contact_name"`                                 // 收件人姓名
	ContactPhone  string     `gorm:"size:20" json:"contact_phone"`                                // 收件人手机
	ContactAddress string   `gorm:"size:500" json:"contact_address"`                              // 收件地址
	EscrowEnabled bool       `gorm:"default:false;index" json:"escrow_enabled"`                    // 是否启用担保交易
	InstallmentEnabled bool  `gorm:"default:false" json:"installment_enabled"`                     // 是否分期付款
	InstallmentPeriods int   `gorm:"default:0" json:"installment_periods"`                          // 分期期数
	PaidAt        *time.Time `gorm:"index" json:"paid_at"`                                         // 支付时间
	ShippedAt     *time.Time `gorm:"index" json:"shipped_at"`                                      // 发货时间
	ReceivedAt    *time.Time `gorm:"index" json:"received_at"`                                     // 收货时间
	SettledAt     *time.Time `gorm:"index" json:"settled_at"`                                      // 结算时间
	ClosedAt      *time.Time `gorm:"index" json:"closed_at"`                                       // 关闭时间
	AutoCloseAt   *time.Time `gorm:"index" json:"auto_close_at"`                                   // 自动关闭时间（24h未付款）
	AutoReceiveAt *time.Time `gorm:"index" json:"auto_receive_at"`                                 // 自动收货时间（7d不确认）
}

// TableName 表名（ers_ 前缀）
func (ErshouOrder) TableName() string { return "ers_orders" }

// ErshouOrderItem 订单子项（对标转转）
// 一个订单可包含多个商品子项（多件合并下单）
type ErshouOrderItem struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	OrderID   uint    `gorm:"not null;index:idx_order_item_order;uniqueIndex:uniq_order_item" json:"order_id"`   // 关联订单ID
	ErshouID  uint    `gorm:"not null;index" json:"ershou_id"`                                                                  // 关联二手物品ID
	SKUID     uint    `gorm:"index" json:"sku_id"`                                                                              // 关联 SKU ID（无 SKU 时为 0）
	SKUCode   string  `gorm:"size:64" json:"sku_code"`                                                                          // SKU 编码冗余
	Title     string  `gorm:"size:200" json:"title"`                                                                              // 商品标题快照
	CoverImage string  `gorm:"size:255" json:"cover_image"`                                                                       // 商品图片快照
	Quantity  int     `gorm:"not null;default:1;uniqueIndex:uniq_order_item" json:"quantity"`                                     // 购买数量
	UnitPrice float64 `gorm:"type:decimal(12,2);not null" json:"unit_price"`                                                      // 单价（下单时快照）
	Subtotal  float64 `gorm:"type:decimal(12,2);not null" json:"subtotal"`                                                       // 小计 = 单价 × 数量
	Remark    string  `gorm:"size:500" json:"remark"`                                                                             // 子项备注
}

// TableName 表名（ers_ 前缀）
func (ErshouOrderItem) TableName() string { return "ers_order_items" }
