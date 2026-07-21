// Package dto 同城商城 - 统计 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticInfo 统计详情响应
type StatisticInfo struct {
	ID               uint      `json:"id"`
	StatDate         string    `json:"stat_date"`
	StatType         string    `json:"stat_type"`
	ShopID           uint      `json:"shop_id"`
	ProductID        uint      `json:"product_id"`
	CategoryID       uint      `json:"category_id"`
	OrderCount       int64     `json:"order_count"`
	OrderAmount      float64   `json:"order_amount"`
	PaidOrderCount   int64     `json:"paid_order_count"`
	PaidOrderAmount  float64   `json:"paid_order_amount"`
	CancelledOrderCount int64  `json:"cancelled_order_count"`
	RefundCount      int64     `json:"refund_count"`
	RefundAmount     float64   `json:"refund_amount"`
	ViewCount        int64     `json:"view_count"`
	FavoriteCount    int64     `json:"favorite_count"`
	CartCount        int64     `json:"cart_count"`
	SalesCount       int64     `json:"sales_count"`
	ReviewCount      int64     `json:"review_count"`
	NewReviewCount   int64     `json:"new_review_count"`
	AvgRating        float64   `json:"avg_rating"`
	GoodRate         float64   `json:"good_rate"`
	NewBuyerCount    int64     `json:"new_buyer_count"`
	ActiveBuyerCount int64     `json:"active_buyer_count"`
	RepurchaseCount  int64     `json:"repurchase_count"`
	ConversionRate   float64   `json:"conversion_rate"`
	RegionID         uint      `json:"region_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// StatisticListRequest 统计列表请求
type StatisticListRequest struct {
	utils.Pagination
	StartDate   string `form:"start_date" json:"start_date"`
	EndDate     string `form:"end_date" json:"end_date"`
	StatType    string `form:"stat_type" json:"stat_type"`
	ShopID      uint   `form:"shop_id" json:"shop_id"`
	ProductID   uint   `form:"product_id" json:"product_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// StatisticSummaryRequest 统计汇总请求
type StatisticSummaryRequest struct {
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	GroupBy   string `form:"group_by" json:"group_by"` // day/week/month/shop/product/category
}

// StatisticSummary 统计汇总响应
type StatisticSummary struct {
	TotalOrderCount      int64   `json:"total_order_count"`
	TotalOrderAmount     float64 `json:"total_order_amount"`
	TotalPaidAmount      float64 `json:"total_paid_amount"`
	TotalRefundAmount    float64 `json:"total_refund_amount"`
	TotalViewCount       int64   `json:"total_view_count"`
	TotalFavoriteCount   int64   `json:"total_favorite_count"`
	TotalSalesCount      int64   `json:"total_sales_count"`
	TotalReviewCount     int64   `json:"total_review_count"`
	AvgRating            float64 `json:"avg_rating"`
	TotalNewBuyerCount   int64   `json:"total_new_buyer_count"`
	TotalActiveBuyerCount int64  `json:"total_active_buyer_count"`
	TotalRepurchaseCount int64   `json:"total_repurchase_count"`
	AvgConversionRate    float64 `json:"avg_conversion_rate"`
	Groups               []StatisticGroupItem `json:"groups,omitempty"`
}

// StatisticGroupItem 分组统计条目
type StatisticGroupItem struct {
	GroupKey     string  `json:"group_key"`
	OrderCount   int64   `json:"order_count"`
	OrderAmount  float64 `json:"order_amount"`
	PaidAmount   float64 `json:"paid_amount"`
	SalesCount   int64   `json:"sales_count"`
	ViewCount    int64   `json:"view_count"`
}

// UpsertStatisticRequest 写入/更新统计请求（M 端定时任务）
type UpsertStatisticRequest struct {
	StatDate         string `json:"stat_date" binding:"required"`
	StatType         string `json:"stat_type" binding:"required,oneof=daily shop product category region"`
	ShopID           uint   `json:"shop_id"`
	ProductID        uint   `json:"product_id"`
	CategoryID       uint   `json:"category_id"`
	OrderCount       int64  `json:"order_count"`
	OrderAmount      float64 `json:"order_amount"`
	PaidOrderCount   int64  `json:"paid_order_count"`
	PaidOrderAmount  float64 `json:"paid_order_amount"`
	CancelledOrderCount int64 `json:"cancelled_order_count"`
	RefundCount      int64  `json:"refund_count"`
	RefundAmount     float64 `json:"refund_amount"`
	ViewCount        int64  `json:"view_count"`
	FavoriteCount    int64  `json:"favorite_count"`
	CartCount        int64  `json:"cart_count"`
	SalesCount       int64  `json:"sales_count"`
	ReviewCount      int64  `json:"review_count"`
	NewReviewCount   int64  `json:"new_review_count"`
	AvgRating        float64 `json:"avg_rating"`
	GoodRate         float64 `json:"good_rate"`
	NewBuyerCount    int64  `json:"new_buyer_count"`
	ActiveBuyerCount int64  `json:"active_buyer_count"`
	RepurchaseCount  int64  `json:"repurchase_count"`
	ConversionRate   float64 `json:"conversion_rate"`
}
