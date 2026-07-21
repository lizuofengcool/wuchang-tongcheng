// Package model 同城商城 - 购物车
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 用户隔离）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Cart 购物车表
type Cart struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	UserID    uint `gorm:"index;not null" json:"user_id"`        // 用户 ID
	ShopID    uint `gorm:"index;not null" json:"shop_id"`        // 店铺 ID
	ProductID uint `gorm:"index;not null" json:"product_id"`     // 商品 SPU ID
	SkuID     uint `gorm:"index;not null;default:0" json:"sku_id"` // SKU ID（0=无规格商品）

	// === 冗余信息（避免查询时多次 join，下单时锁定快照） ===
	ProductName string `gorm:"size:200;not null;default:''" json:"product_name"` // 商品名
	MainImage   string `gorm:"size:255;not null;default:''" json:"main_image"`   // 商品主图
	SkuName     string `gorm:"size:200;not null;default:''" json:"sku_name"`     // SKU 名（如：红色-XL）
	SkuSpecs    string `gorm:"size:500;not null;default:''" json:"sku_specs"`    // SKU 规格文本

	// === 价格与数量 ===
	Price    float64 `gorm:"type:decimal(12,2);default:0" json:"price"`    // 加入时单价
	Quantity int     `gorm:"not null;default:1" json:"quantity"`            // 数量
	Selected int     `gorm:"not null;default:1;index" json:"selected"`     // 0未选中 1已选中

	// === 状态 ===
	Status int `gorm:"default:1;index" json:"status"` // 0失效（商品下架/删除） 1有效
}

// TableName 表名
func (Cart) TableName() string { return "mall_cart" }
