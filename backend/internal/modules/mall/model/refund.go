// Package model 同城商城 - 退款记录
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 买家 + order_id 订单 + shop_id 商家）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Refund 退款表
type Refund struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	OrderID      uint `gorm:"index;not null" json:"order_id"`      // 订单 ID
	OrderNo      string `gorm:"size:32;not null;default:'';index" json:"order_no"` // 订单号（冗余）
	OrderItemID  uint `gorm:"index;not null;default:0" json:"order_item_id"` // 订单明细 ID（0=整单退款）
	UserID       uint `gorm:"index;not null" json:"user_id"`       // 买家 ID
	ShopID       uint `gorm:"index;not null" json:"shop_id"`       // 店铺 ID
	PaymentID    uint `gorm:"index;not null;default:0" json:"payment_id"` // 支付单 ID

	// === 退款单信息 ===
	RefundNo  string `gorm:"size:64;not null;default:'';index" json:"refund_no"` // 退款单号（业务唯一）
	TradeNo   string `gorm:"size:64;not null;default:'';index" json:"trade_no"`  // 第三方退款流水号

	// === 金额（decimal(12,2)） ===
	Amount       float64 `gorm:"type:decimal(12,2);default:0" json:"amount"`       // 退款金额
	RefundAmount float64 `gorm:"type:decimal(12,2);default:0" json:"refund_amount"` // 实际退款金额（可能扣除手续费）

	// === 退款类型与原因 ===
	RefundType string `gorm:"size:32;not null;default:'only_refund';index" json:"refund_type"` // only_refund 仅退款 / return 退货退款
	Reason     string `gorm:"size:500;not null;default:''" json:"reason"`           // 退款原因
	ReasonCode string `gorm:"size:32;not null;default:''" json:"reason_code"`       // 退款原因码（如 quality_issue）
	Description string `gorm:"type:text" json:"description"`                        // 详细描述
	EvidenceImages JSONB `gorm:"type:jsonb" json:"evidence_images"`                  // 证据图片
	ExpressCompany string `gorm:"size:64;not null;default:''" json:"express_company"` // 退货物流公司
	ExpressNo     string `gorm:"size:64;not null;default:''" json:"express_no"`     // 退货物流单号

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`        // 0待审核 1已同意 2已拒绝 3已退款 4已关闭
	SellerRemark string `gorm:"size:500;not null;default:''" json:"seller_remark"` // 卖家处理备注
	AdminRemark  string `gorm:"size:500;not null;default:''" json:"admin_remark"`   // 平台处理备注

	// === 处理人 ===
	HandlerID   uint   `gorm:"index" json:"handler_id"`               // 处理人 ID
	HandlerName string `gorm:"size:50;not null;default:''" json:"handler_name"` // 处理人姓名

	// === 时间 ===
	ApprovedAt  *time.Time `gorm:"index" json:"approved_at"`   // 同意时间
	RejectedAt  *time.Time `gorm:"index" json:"rejected_at"`   // 拒绝时间
	RefundedAt  *time.Time `gorm:"index" json:"refunded_at"`   // 退款完成时间
	ClosedAt    *time.Time `gorm:"index" json:"closed_at"`     // 关闭时间
	ShippedAt   *time.Time `gorm:"index" json:"shipped_at"`    // 退货发货时间
	ReceivedAt  *time.Time `gorm:"index" json:"received_at"`   // 退货签收时间

	// === 第三方原始响应 ===
	RawResponse JSONB `gorm:"type:jsonb" json:"raw_response"` // 第三方退款响应原始数据
}

// TableName 表名
func (Refund) TableName() string { return "mall_refunds" }
