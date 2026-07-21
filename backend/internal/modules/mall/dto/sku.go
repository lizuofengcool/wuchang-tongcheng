// Package dto 同城商城 - SKU DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// SkuInfo SKU 详情响应
type SkuInfo struct {
	ID            uint      `json:"id"`
	ProductID     uint      `json:"product_id"`
	ShopID        uint      `json:"shop_id"`
	Name          string    `json:"name"`
	SkuCode       string    `json:"sku_code"`
	Barcode       string    `json:"barcode"`
	Specs         interface{} `json:"specs"`
	Price         float64   `json:"price"`
	OriginalPrice float64   `json:"original_price"`
	CostPrice     float64   `json:"cost_price"`
	Stock         int       `json:"stock"`
	Sales         int64     `json:"sales"`
	WarnStock     int       `json:"warn_stock"`
	Image         string    `json:"image"`
	Sort          int       `json:"sort"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateSkuRequest 创建 SKU 请求
type CreateSkuRequest struct {
	Name          string      `json:"name" binding:"max=200"`
	SkuCode       string      `json:"sku_code" binding:"max=64"`
	Barcode       string      `json:"barcode" binding:"max=64"`
	Specs         interface{} `json:"specs"`
	Price         float64     `json:"price" binding:"required"`
	OriginalPrice float64     `json:"original_price"`
	CostPrice     float64     `json:"cost_price"`
	Stock         int         `json:"stock"`
	WarnStock     int         `json:"warn_stock"`
	Image         string      `json:"image" binding:"max=255"`
	Sort          int         `json:"sort"`
	Status        int         `json:"status"`
}

// UpdateSkuRequest 更新 SKU 请求
type UpdateSkuRequest struct {
	Name          *string     `json:"name" binding:"max=200"`
	SkuCode       *string     `json:"sku_code" binding:"max=64"`
	Barcode       *string     `json:"barcode" binding:"max=64"`
	Specs         interface{} `json:"specs"`
	Price         *float64    `json:"price"`
	OriginalPrice *float64    `json:"original_price"`
	CostPrice     *float64    `json:"cost_price"`
	Stock         *int        `json:"stock"`
	WarnStock     *int        `json:"warn_stock"`
	Image         *string     `json:"image" binding:"max=255"`
	Sort          *int        `json:"sort"`
	Status        *int        `json:"status"`
}

// SkuListRequest SKU 列表请求
type SkuListRequest struct {
	utils.Pagination
	ProductID uint   `form:"product_id" json:"product_id"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	SkuCode   string `form:"sku_code" json:"sku_code"`
	Barcode   string `form:"barcode" json:"barcode"`
	Status    *int   `form:"status" json:"status"`
}

// BatchUpdateStockRequest 批量更新库存请求
type BatchUpdateStockRequest struct {
	Items []SkuStockUpdate `json:"items" binding:"required,min=1"`
}

// SkuStockUpdate SKU 库存更新
type SkuStockUpdate struct {
	SkuID    uint `json:"sku_id" binding:"required"`
	Stock    int  `json:"stock"`     // 新库存（绝对值）
	Quantity int  `json:"quantity"`  // 变更数量（增量，可为负）
}
