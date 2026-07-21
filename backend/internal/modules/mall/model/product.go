// Package model 同城商城 - 商品 SPU 主表
// 依据需求文档 1.10：4 维数据隔离（region_id + shop_id 商家隔离）
// 对标淘宝/京东/拼多多商品 SPU
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Product 商品 SPU 主表
type Product struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at

	// === 基础信息 ===
	ShopID     uint   `gorm:"index;not null" json:"shop_id"`                       // 店铺 ID
	UserID     uint   `gorm:"index;not null" json:"user_id"`                       // 发布者（店主） ID
	CategoryID uint   `gorm:"index;not null;default:0" json:"category_id"`         // 商品分类 ID
	BrandID    uint   `gorm:"index;not null;default:0" json:"brand_id"`            // 品牌 ID（预留）
	Name       string `gorm:"size:200;not null;default:'';index" json:"name"`      // 商品名
	Subtitle   string `gorm:"size:500;not null;default:''" json:"subtitle"`        // 副标题
	MainImage  string `gorm:"size:255;not null;default:''" json:"main_image"`      // 主图 URL
	Detail     string `gorm:"type:text" json:"detail"`                              // 商品详情（富文本 HTML）
	ProductType string `gorm:"size:32;not null;default:'physical';index" json:"product_type"` // physical/virtual/service

	// === 价格 ===
	Price         float64 `gorm:"type:decimal(12,2);default:0;index" json:"price"`         // 现价
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0" json:"original_price"`       // 原价（划线价）
	CostPrice     float64 `gorm:"type:decimal(12,2);default:0" json:"cost_price"`           // 成本价（仅店主可见）
	MinPrice      float64 `gorm:"type:decimal(12,2);default:0" json:"min_price"`            // SKU 最低价（冗余）
	MaxPrice      float64 `gorm:"type:decimal(12,2);default:0" json:"max_price"`            // SKU 最高价（冗余）

	// === 库存与销量 ===
	Stock       int   `gorm:"not null;default:0" json:"stock"`         // 总库存（冗余 = sum(sku.stock)）
	Sales       int64 `gorm:"not null;default:0;index" json:"sales"`   // 销量
	VirtualSales int64 `gorm:"not null;default:0" json:"virtual_sales"` // 虚拟销量
	StockWarn    int   `gorm:"not null;default:0" json:"stock_warn"`    // 库存预警阈值

	// === 状态 ===
	Status      int    `gorm:"default:0;index" json:"status"`         // 0草稿 1在售 2下架 3售罄 4回收站
	AuditStatus int    `gorm:"default:0;index" json:"audit_status"`   // 0待审 1通过 2拒绝
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"` // 审核拒绝原因
	PublishedAt *time.Time `gorm:"index" json:"published_at"`         // 上架时间

	// === 统计 ===
	ViewCount    int     `gorm:"not null;default:0" json:"view_count"`     // 浏览数
	FavoriteCount int    `gorm:"not null;default:0" json:"favorite_count"` // 收藏数
	ReviewCount  int     `gorm:"not null;default:0" json:"review_count"`   // 评价数
	Rating       float64 `gorm:"type:decimal(3,2);default:0;index" json:"rating"` // 综合评分 0-5
	GoodRate     float64 `gorm:"type:decimal(5,2);default:0" json:"good_rate"`    // 好评率

	// === 运营字段 ===
	Featured       bool    `gorm:"default:false;index" json:"featured"`                   // 精选推荐
	Recommended    bool    `gorm:"default:false;index" json:"recommended"`                // 店长推荐
	NewArrival     bool    `gorm:"default:false;index" json:"new_arrival"`                // 新品
	HotSale        bool    `gorm:"default:false;index" json:"hot_sale"`                   // 热销
	PromotionLevel int     `gorm:"default:0" json:"promotion_level"`                     // 推广等级 0-10
	TrafficWeight  float64 `gorm:"type:decimal(3,2);default:1.00" json:"traffic_weight"` // 流量权重
	Sort           int     `gorm:"not null;default:0;index" json:"sort"`                 // 排序

	// === 物流参数 ===
	FreeShipping    bool    `gorm:"default:false" json:"free_shipping"`           // 包邮
	ShippingFee     float64 `gorm:"type:decimal(10,2);default:0" json:"shipping_fee"` // 固定运费
	ShippingTemplateID uint `gorm:"not null;default:0" json:"shipping_template_id"`    // 运费模板 ID（预留）
	Weight          float64 `gorm:"type:decimal(10,3);default:0" json:"weight"`       // 重量（kg）
	Volume          float64 `gorm:"type:decimal(10,3);default:0" json:"volume"`       // 体积（m³）

	// === JSONB 字段 ===
	Images     JSONB `gorm:"type:jsonb" json:"images"`     // 商品图片列表
	Specs      JSONB `gorm:"type:jsonb" json:"specs"`      // 规格定义（颜色/尺寸等）
	Attributes JSONB `gorm:"type:jsonb" json:"attributes"` // 商品属性（品牌/产地等）
	Tags       JSONB `gorm:"type:jsonb" json:"tags"`       // 商品标签
	SkuSpecs   JSONB `gorm:"type:jsonb" json:"sku_specs"`  // SKU 规格组合预览

	// === 风控 ===
	ContentHash string `gorm:"size:64;index" json:"content_hash"` // 图文指纹
	RiskScore   int    `gorm:"default:0;index" json:"risk_score"` // 风险评分 0-100
}

// TableName 表名
func (Product) TableName() string { return "mall_products" }
