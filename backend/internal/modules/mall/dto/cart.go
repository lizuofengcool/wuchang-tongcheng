// Package dto 同城商城 - 购物车 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CartInfo 购物车项响应
type CartInfo struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	ShopID      uint      `json:"shop_id"`
	ProductID   uint      `json:"product_id"`
	SkuID       uint      `json:"sku_id"`
	ProductName string    `json:"product_name"`
	MainImage   string    `json:"main_image"`
	SkuName     string    `json:"sku_name"`
	SkuSpecs    string    `json:"sku_specs"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Selected    int       `json:"selected"`
	SelectedText string   `json:"selected_text"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	Subtotal    float64   `json:"subtotal"` // 小计（单价×数量）
	RegionID    uint      `json:"region_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CartGroupByShop 按店铺分组的购物车
type CartGroupByShop struct {
	ShopID    uint       `json:"shop_id"`
	ShopName  string     `json:"shop_name"`
	ShopLogo  string     `json:"shop_logo"`
	Items     []CartInfo `json:"items"`
	TotalAmount float64  `json:"total_amount"`
	TotalCount int       `json:"total_count"`
}

// AddCartRequest 加入购物车请求
type AddCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	SkuID     uint `json:"sku_id"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// UpdateCartRequest 更新购物车请求
type UpdateCartRequest struct {
	Quantity *int `json:"quantity" binding:"omitempty,min=1"`
	Selected *int `json:"selected" binding:"omitempty,oneof=0 1"`
}

// BatchUpdateCartRequest 批量更新购物车请求
type BatchUpdateCartRequest struct {
	Items []CartItemUpdate `json:"items" binding:"required,min=1"`
}

// CartItemUpdate 购物车项更新
type CartItemUpdate struct {
	ID       uint `json:"id" binding:"required"`
	Quantity *int `json:"quantity" binding:"omitempty,min=1"`
	Selected *int `json:"selected" binding:"omitempty,oneof=0 1"`
}

// SelectAllCartRequest 全选/取消全选请求
type SelectAllCartRequest struct {
	Selected int  `json:"selected" binding:"oneof=0 1"`
	ShopID   uint `json:"shop_id"` // 0=全部
}

// CartListRequest 购物车列表请求
type CartListRequest struct {
	utils.Pagination
	ShopID   *uint `form:"shop_id" json:"shop_id"`
	Selected *int  `form:"selected" json:"selected"`
	Status   *int  `form:"status" json:"status"`
}

// CartSummaryRequest 购物车汇总请求
type CartSummaryRequest struct {
	SelectedOnly bool `form:"selected_only" json:"selected_only"`
}

// CartSummary 购物车汇总响应
type CartSummary struct {
	TotalCount   int     `json:"total_count"`
	TotalAmount  float64 `json:"total_amount"`
	SelectedCount int    `json:"selected_count"`
	SelectedAmount float64 `json:"selected_amount"`
	ShopCount    int     `json:"shop_count"`
}
