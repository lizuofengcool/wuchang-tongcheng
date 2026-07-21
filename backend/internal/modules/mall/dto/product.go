// Package dto 同城商城 - 商品 SPU DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ProductInfo 商品详情响应
type ProductInfo struct {
	ID            uint       `json:"id"`
	ShopID        uint       `json:"shop_id"`
	UserID        uint       `json:"user_id"`
	CategoryID    uint       `json:"category_id"`
	BrandID       uint       `json:"brand_id"`
	Name          string     `json:"name"`
	Subtitle      string     `json:"subtitle"`
	MainImage     string     `json:"main_image"`
	Detail        string     `json:"detail"`
	ProductType   string     `json:"product_type"`
	ProductTypeText string   `json:"product_type_text"`

	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	MinPrice      float64 `json:"min_price"`
	MaxPrice      float64 `json:"max_price"`

	Stock         int     `json:"stock"`
	Sales         int64   `json:"sales"`
	VirtualSales  int64   `json:"virtual_sales"`
	StockWarn     int     `json:"stock_warn"`

	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	AuditStatus   int        `json:"audit_status"`
	AuditStatusText string   `json:"audit_status_text"`
	AuditReason   string     `json:"audit_reason"`
	PublishedAt   *time.Time `json:"published_at"`

	ViewCount    int     `json:"view_count"`
	FavoriteCount int    `json:"favorite_count"`
	ReviewCount  int     `json:"review_count"`
	Rating       float64 `json:"rating"`
	GoodRate     float64 `json:"good_rate"`

	Featured       bool    `json:"featured"`
	Recommended    bool    `json:"recommended"`
	NewArrival     bool    `json:"new_arrival"`
	HotSale        bool    `json:"hot_sale"`
	PromotionLevel int     `json:"promotion_level"`
	TrafficWeight  float64 `json:"traffic_weight"`
	Sort           int     `json:"sort"`

	FreeShipping      bool    `json:"free_shipping"`
	ShippingFee       float64 `json:"shipping_fee"`
	ShippingTemplateID uint   `json:"shipping_template_id"`
	Weight            float64 `json:"weight"`
	Volume            float64 `json:"volume"`

	Images     interface{} `json:"images"`
	Specs      interface{} `json:"specs"`
	Attributes interface{} `json:"attributes"`
	Tags       interface{} `json:"tags"`
	SkuSpecs   interface{} `json:"sku_specs"`

	Skus []SkuInfo `json:"skus,omitempty"` // SKU 列表（详情时返回）

	ShopName string `json:"shop_name,omitempty"` // 店铺名（冗余）
	ShopLogo string `json:"shop_logo,omitempty"` // 店铺 Logo（冗余）

	HasFaved bool `json:"has_faved"`

	RegionID  uint      `json:"region_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	ShopID        uint   `json:"shop_id" binding:"required"`
	CategoryID    uint   `json:"category_id" binding:"required"`
	Name          string `json:"name" binding:"required,max=200"`
	Subtitle      string `json:"subtitle" binding:"max=500"`
	MainImage     string `json:"main_image" binding:"max=255"`
	Detail        string `json:"detail"`
	ProductType   string `json:"product_type" binding:"omitempty,oneof=physical virtual service"`
	Price         float64 `json:"price" binding:"required"`
	OriginalPrice float64 `json:"original_price"`
	CostPrice     float64 `json:"cost_price"`
	Stock         int    `json:"stock"`
	StockWarn     int    `json:"stock_warn"`
	Status        int    `json:"status"`
	FreeShipping  bool   `json:"free_shipping"`
	ShippingFee   float64 `json:"shipping_fee"`
	Weight        float64 `json:"weight"`
	Volume        float64 `json:"volume"`
	Images        interface{} `json:"images"`
	Specs         interface{} `json:"specs"`
	Attributes    interface{} `json:"attributes"`
	Tags          interface{} `json:"tags"`
	Skus          []CreateSkuRequest `json:"skus"` // 一并创建 SKU
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Name          *string `json:"name" binding:"max=200"`
	Subtitle      *string `json:"subtitle" binding:"max=500"`
	MainImage     *string `json:"main_image" binding:"max=255"`
	Detail        *string `json:"detail"`
	CategoryID    *uint   `json:"category_id"`
	BrandID       *uint   `json:"brand_id"`
	ProductType   *string `json:"product_type" binding:"omitempty,oneof=physical virtual service"`
	Price         *float64 `json:"price"`
	OriginalPrice *float64 `json:"original_price"`
	CostPrice     *float64 `json:"cost_price"`
	Stock         *int    `json:"stock"`
	StockWarn     *int    `json:"stock_warn"`
	Status        *int    `json:"status"`
	FreeShipping  *bool   `json:"free_shipping"`
	ShippingFee   *float64 `json:"shipping_fee"`
	Weight        *float64 `json:"weight"`
	Volume        *float64 `json:"volume"`
	Featured      *bool   `json:"featured"`
	Recommended   *bool   `json:"recommended"`
	NewArrival    *bool   `json:"new_arrival"`
	HotSale       *bool   `json:"hot_sale"`
	Sort          *int    `json:"sort"`
	Images        interface{} `json:"images"`
	Specs         interface{} `json:"specs"`
	Attributes    interface{} `json:"attributes"`
	Tags          interface{} `json:"tags"`
}

// ProductListRequest 商品列表请求
type ProductListRequest struct {
	utils.Pagination
	Keyword      string  `form:"keyword" json:"keyword"`
	ShopID       uint    `form:"shop_id" json:"shop_id"`
	CategoryID   uint    `form:"category_id" json:"category_id"`
	BrandID      uint    `form:"brand_id" json:"brand_id"`
	ProductType  string  `form:"product_type" json:"product_type"`
	Status       *int    `form:"status" json:"status"`
	AuditStatus  *int    `form:"audit_status" json:"audit_status"`
	MinPrice     float64 `form:"min_price" json:"min_price"`
	MaxPrice     float64 `form:"max_price" json:"max_price"`
	Featured     *bool   `form:"featured" json:"featured"`
	Recommended  *bool   `form:"recommended" json:"recommended"`
	NewArrival   *bool   `form:"new_arrival" json:"new_arrival"`
	HotSale      *bool   `form:"hot_sale" json:"hot_sale"`
	FreeShipping *bool   `form:"free_shipping" json:"free_shipping"`
	Sort         string  `form:"sort" json:"sort"` // new/price_asc/price_desc/sales
	UserID       uint    `form:"user_id" json:"user_id"`
}

// ProductAdminListRequest 管理后台商品列表请求
type ProductAdminListRequest struct {
	utils.Pagination
	Keyword     string `form:"keyword" json:"keyword"`
	ShopID      uint   `form:"shop_id" json:"shop_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	BrandID     uint   `form:"brand_id" json:"brand_id"`
	ProductType string `form:"product_type" json:"product_type"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Featured    *bool  `form:"featured" json:"featured"`
	Recommended *bool  `form:"recommended" json:"recommended"`
	UserID      uint   `form:"user_id" json:"user_id"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// PromotionRequest 推广配置请求
type PromotionRequest struct {
	Featured       *bool    `json:"featured"`
	Recommended    *bool    `json:"recommended"`
	NewArrival     *bool    `json:"new_arrival"`
	HotSale        *bool    `json:"hot_sale"`
	PromotionLevel *int     `json:"promotion_level"`
	TrafficWeight  *float64 `json:"traffic_weight"`
	Sort           *int     `json:"sort"`
}
