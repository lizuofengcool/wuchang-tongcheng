// Package model 同城商城 - 商品 SKU 规格
// 依据需求文档 1.10：4 维数据隔离（region_id + shop_id 商家隔离 + product_id 商品）
// 对标淘宝/京东/拼多多 SKU 规格明细
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Sku 商品 SKU 规格表
type Sku struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 关联 ===
	ProductID uint `gorm:"index;not null" json:"product_id"` // 商品 SPU ID
	ShopID    uint `gorm:"index;not null;default:0" json:"shop_id"` // 店铺 ID（冗余，便于商家隔离查询）
	UserID    uint `gorm:"index;not null;default:0" json:"user_id"` // 店主用户 ID（冗余）

	// === 基础信息 ===
	Name      string `gorm:"size:200;not null;default:''" json:"name"`     // SKU 名称（如：红色-XL）
	SkuCode   string `gorm:"size:64;not null;default:'';index" json:"sku_code"` // SKU 编码（商家自定义）
	Barcode   string `gorm:"size:64;not null;default:'';index" json:"barcode"`   // 商品条形码（EAN-13）

	// === 规格 ===
	Specs JSONB `gorm:"type:jsonb" json:"specs"` // 规格键值（如 [{name:颜色,value:红},{name:尺寸,value:XL}]）

	// === 价格 ===
	Price         float64 `gorm:"type:decimal(12,2);default:0;index" json:"price"`         // 现价
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0" json:"original_price"`      // 原价
	CostPrice     float64 `gorm:"type:decimal(12,2);default:0" json:"cost_price"`          // 成本价

	// === 库存与销量 ===
	Stock    int   `gorm:"not null;default:0;index" json:"stock"`     // 库存
	Sales    int64 `gorm:"not null;default:0;index" json:"sales"`     // 销量
	WarnStock int  `gorm:"not null;default:0" json:"warn_stock"`      // 库存预警阈值

	// === 显示 ===
	Image    string `gorm:"size:255;not null;default:''" json:"image"` // SKU 专属图片
	Sort     int    `gorm:"not null;default:0;index" json:"sort"`      // 排序
	Status   int    `gorm:"default:1;index" json:"status"`             // 0禁用 1启用
}

// TableName 表名
func (Sku) TableName() string { return "mall_skus" }
