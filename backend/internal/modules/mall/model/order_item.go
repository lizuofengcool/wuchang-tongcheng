// Package model 同城商城 - 订单明细
// 依据需求文档 1.10：4 维数据隔离（region_id + order_id 订单隔离 + shop_id 商家隔离）
// 对标淘宝/京东订单明细（商品快照，避免商品修改影响历史订单）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// OrderItem 订单明细表
type OrderItem struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	OrderID   uint `gorm:"index;not null" json:"order_id"`     // 订单 ID
	OrderNo   string `gorm:"size:32;not null;default:'';index" json:"order_no"` // 订单号（冗余）
	UserID    uint `gorm:"index;not null" json:"user_id"`      // 买家 ID
	ShopID    uint `gorm:"index;not null" json:"shop_id"`      // 店铺 ID
	ProductID uint `gorm:"index;not null" json:"product_id"`   // 商品 SPU ID
	SkuID     uint `gorm:"index;not null;default:0" json:"sku_id"` // SKU ID

	// === 商品快照（避免商品修改影响历史订单） ===
	ProductName string `gorm:"size:200;not null;default:''" json:"product_name"` // 商品名
	MainImage   string `gorm:"size:255;not null;default:''" json:"main_image"`   // 商品主图
	SkuName     string `gorm:"size:200;not null;default:''" json:"sku_name"`     // SKU 名（如：红色-XL）
	SkuSpecs    string `gorm:"size:500;not null;default:''" json:"sku_specs"`    // SKU 规格文本
	SkuCode     string `gorm:"size:64;not null;default:''" json:"sku_code"`      // SKU 编码

	// === 金额（decimal(12,2)） ===
	Price         float64 `gorm:"type:decimal(12,2);default:0" json:"price"`         // 单价
	Quantity      int     `gorm:"not null;default:1" json:"quantity"`                // 数量
	TotalAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"total_amount"`  // 小计金额（单价×数量）
	DiscountAmount float64 `gorm:"type:decimal(12,2);default:0" json:"discount_amount"` // 优惠分摊
	ShippingFee   float64 `gorm:"type:decimal(12,2);default:0" json:"shipping_fee"` // 运费分摊
	PayAmount     float64 `gorm:"type:decimal(12,2);default:0" json:"pay_amount"`    // 实付金额

	// === 评价 ===
	HasReview   bool `gorm:"default:false" json:"has_review"`     // 是否已评价
	ReviewID    uint `gorm:"not null;default:0;index" json:"review_id"` // 评价 ID

	// === 退款 ===
	RefundStatus int   `gorm:"default:0;index" json:"refund_status"` // 0未退款 1待审核 2已同意 3已拒绝 4已退款 5已关闭
	RefundID     uint  `gorm:"not null;default:0;index" json:"refund_id"` // 退款单 ID

	// === 状态 ===
	Status int `gorm:"default:0;index" json:"status"` // 0待付款 1已付款 2已发货 3已收货 4已完成 5已取消 6已退款 7已关闭
}

// TableName 表名
func (OrderItem) TableName() string { return "mall_order_items" }
