// Package model 商品 SKU 规格子表（对标闲鱼/转转）
// 支持颜色×尺寸×版本组合 + 独立库存 + 独立价格
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// ErshouSKU 商品 SKU 规格子表
// 关联 erhous.id，每条记录代表一种规格组合（如 颜色=黑 / 尺寸=XL / 版本=Pro）
type ErshouSKU struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	ErshouID   uint    `gorm:"not null;index;uniqueIndex:uniq_sku_ershou_code" json:"ershou_id"` // 关联二手物品ID
	SKUCode    string  `gorm:"size:64;not null;uniqueIndex:uniq_sku_ershou_code" json:"sku_code"` // SKU 编码（业务唯一，如 BLACK-XL-PRO）
	Name       string  `gorm:"size:200" json:"name"`                                              // SKU 名称（如 黑色 XL Pro）
	Color      string  `gorm:"size:50" json:"color"`                                             // 颜色（黑/白/红）
	Size       string  `gorm:"size:50" json:"size"`                                              // 尺寸（S/M/L/XL）
	Version    string  `gorm:"size:100" json:"version"`                                         // 版本（标准版/Pro/Max）
	Price      float64 `gorm:"type:decimal(12,2);default:0" json:"price"`                        // SKU 独立售价
	Stock      int     `gorm:"default:0" json:"stock"`                                           // 库存数量
	SoldCount  int     `gorm:"default:0" json:"sold_count"`                                       // 已售数量
	Image      string  `gorm:"size:255" json:"image"`                                            // SKU 图片
	Weight     float64 `gorm:"type:decimal(8,2);default:0" json:"weight"`                          // 重量(kg)
	Barcode    string  `gorm:"size:64;index" json:"barcode"`                                       // 条形码
	Status     int     `gorm:"default:1;index" json:"status"`                                     // 0下架 1上架 2售罄
	Attributes JSONB   `gorm:"type:jsonb" json:"attributes"`                                       // 其他属性 {key:value}
	Sort       int     `gorm:"default:0" json:"sort"`                                             // 排序
}

// TableName 表名（ers_ 前缀，依据数据库分表前缀规范）
func (ErshouSKU) TableName() string { return "ers_skus" }
