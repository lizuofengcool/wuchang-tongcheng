// Package model 同城商城 - 数据统计表
// 日统计/店铺统计/商品统计/分类统计
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Statistic 统计表
type Statistic struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	StatDate    time.Time `gorm:"type:date;not null;index;uniqueIndex:uniq_mall_statistics_date_type_target" json:"stat_date"` // 统计日期
	StatType    string    `gorm:"size:32;not null;default:'daily';index;uniqueIndex:uniq_mall_statistics_date_type_target" json:"stat_type"` // daily/shop/product/category
	ShopID      uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_mall_statistics_date_type_target" json:"shop_id"` // 店铺 ID（shop 类型时使用）
	ProductID   uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_mall_statistics_date_type_target" json:"product_id"` // 商品 ID（product 类型时使用）
	CategoryID  *uint     `gorm:"index;uniqueIndex:uniq_mall_statistics_date_type_target" json:"category_id"` // 分类 ID（category 类型时使用）

	// === 订单统计 ===
	OrderCount       int64   `gorm:"not null;default:0" json:"order_count"`        // 订单数
	OrderAmount      float64 `gorm:"type:decimal(14,2);default:0" json:"order_amount"` // 订单金额
	PaidOrderCount   int64   `gorm:"not null;default:0" json:"paid_order_count"`    // 已付款订单数
	PaidOrderAmount  float64 `gorm:"type:decimal(14,2);default:0" json:"paid_order_amount"` // 已付款订单金额
	CancelledOrderCount int64 `gorm:"not null;default:0" json:"cancelled_order_count"` // 取消订单数
	RefundCount      int64   `gorm:"not null;default:0" json:"refund_count"`         // 退款订单数
	RefundAmount     float64 `gorm:"type:decimal(14,2);default:0" json:"refund_amount"` // 退款金额

	// === 商品统计 ===
	ViewCount        int64 `gorm:"not null;default:0" json:"view_count"`     // 浏览数
	FavoriteCount    int64 `gorm:"not null;default:0" json:"favorite_count"` // 收藏数
	CartCount        int64 `gorm:"not null;default:0" json:"cart_count"`     // 加购数
	SalesCount       int64 `gorm:"not null;default:0" json:"sales_count"`    // 销量

	// === 评价统计 ===
	ReviewCount      int64   `gorm:"not null;default:0" json:"review_count"`        // 评价数
	NewReviewCount   int64   `gorm:"not null;default:0" json:"new_review_count"`    // 新增评价数
	AvgRating        float64 `gorm:"type:decimal(3,2);default:0" json:"avg_rating"` // 平均评分
	GoodRate         float64 `gorm:"type:decimal(5,2);default:0" json:"good_rate"`  // 好评率

	// === 用户统计 ===
	NewBuyerCount    int64 `gorm:"not null;default:0" json:"new_buyer_count"`  // 新买家数
	ActiveBuyerCount int64 `gorm:"not null;default:0" json:"active_buyer_count"` // 活跃买家数
	RepurchaseCount  int64 `gorm:"not null;default:0" json:"repurchase_count"`  // 复购数

	// === 转化率 ===
	ConversionRate   float64 `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"` // 转化率（订单/浏览 ×100）
}

// TableName 表名
func (Statistic) TableName() string { return "mall_statistics" }
