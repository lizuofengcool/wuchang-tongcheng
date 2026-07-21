// Package dto 同城商城 - 评价 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ReviewInfo 评价详情响应
type ReviewInfo struct {
	ID            uint       `json:"id"`
	ProductID     uint       `json:"product_id"`
	SkuID         uint       `json:"sku_id"`
	OrderID       uint       `json:"order_id"`
	OrderNo       string     `json:"order_no"`
	OrderItemID   uint       `json:"order_item_id"`
	UserID        uint       `json:"user_id"`
	ShopID        uint       `json:"shop_id"`
	UserName      string     `json:"user_name"`
	UserAvatar    string     `json:"user_avatar"`
	IsAnonymous   bool       `json:"is_anonymous"`
	Rating        int        `json:"rating"`
	Content       string     `json:"content"`
	Images        interface{} `json:"images"`
	Video         string     `json:"video"`
	SkuName       string     `json:"sku_name"`
	SkuSpecs      string     `json:"sku_specs"`
	Reply         string     `json:"reply"`
	ReplyAt       *time.Time `json:"reply_at"`
	ReplyUserID   uint       `json:"reply_user_id"`
	Tags          interface{} `json:"tags"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	AuditReason   string     `json:"audit_reason"`
	HasSellerReply bool      `json:"has_seller_reply"`
	LikeCount     int        `json:"like_count"`
	DislikeCount  int        `json:"dislike_count"`
	ReplyCount    int        `json:"reply_count"`
	AppendContent string     `json:"append_content"`
	AppendImages  interface{} `json:"append_images"`
	AppendAt      *time.Time `json:"append_at"`
	Type          int        `json:"type"`
	ContentHash   string     `json:"content_hash"`
	RiskScore     int        `json:"risk_score"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateReviewRequest 创建评价请求
type CreateReviewRequest struct {
	OrderID     uint   `json:"order_id" binding:"required"`
	OrderItemID uint   `json:"order_item_id" binding:"required"`
	ProductID   uint   `json:"product_id" binding:"required"`
	SkuID       uint   `json:"sku_id"`
	Rating      int    `json:"rating" binding:"required,min=1,max=5"`
	Content     string `json:"content" binding:"max=1000"`
	Images      interface{} `json:"images"`
	Video       string `json:"video" binding:"max=255"`
	Tags        interface{} `json:"tags"`
	IsAnonymous bool   `json:"is_anonymous"`
	Type        int    `json:"type"` // 1首评 2追评
}

// ReplyReviewRequest 商家回复评价请求
type ReplyReviewRequest struct {
	Reply string `json:"reply" binding:"required,max=500"`
}

// AppendReviewRequest 追加评价请求
type AppendReviewRequest struct {
	AppendContent string `json:"append_content" binding:"required,max=1000"`
	AppendImages  interface{} `json:"append_images"`
}

// ReviewListRequest 评价列表请求
type ReviewListRequest struct {
	utils.Pagination
	ProductID uint   `form:"product_id" json:"product_id"`
	SkuID     uint   `form:"sku_id" json:"sku_id"`
	ShopID    uint   `form:"shop_id" json:"shop_id"`
	UserID    uint   `form:"user_id" json:"user_id"`
	OrderID   uint   `form:"order_id" json:"order_id"`
	Rating    *int   `form:"rating" json:"rating"`
	Status    *int   `form:"status" json:"status"`
	Type      *int   `form:"type" json:"type"`
	HasReply  *bool  `form:"has_reply" json:"has_reply"`
	HasImages *bool  `form:"has_images" json:"has_images"`
	Sort      string `form:"sort" json:"sort"` // newest/oldest/highest/lowest/useful
	Keyword   string `form:"keyword" json:"keyword"`
}

// AdminReviewListRequest 管理后台评价列表请求
type AdminReviewListRequest struct {
	utils.Pagination
	ProductID  uint   `form:"product_id" json:"product_id"`
	ShopID     uint   `form:"shop_id" json:"shop_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	Status     *int   `form:"status" json:"status"`
	Rating     *int   `form:"rating" json:"rating"`
	Keyword    string `form:"keyword" json:"keyword"`
	StartDate  string `form:"start_date" json:"start_date"`
	EndDate    string `form:"end_date" json:"end_date"`
	RegionID   uint   `form:"region_id" json:"region_id"`
}

// ReviewStatsRequest 评价统计请求
type ReviewStatsRequest struct {
	ProductID uint `form:"product_id" json:"product_id"`
	ShopID    uint `form:"shop_id" json:"shop_id"`
}

// ReviewStatsInfo 评价统计响应
type ReviewStatsInfo struct {
	TotalCount   int64   `json:"total_count"`
	AvgRating    float64 `json:"avg_rating"`
	FiveStarCount int64  `json:"five_star_count"`
	FourStarCount int64  `json:"four_star_count"`
	ThreeStarCount int64 `json:"three_star_count"`
	TwoStarCount int64   `json:"two_star_count"`
	OneStarCount int64   `json:"one_star_count"`
	HasImagesCount int64 `json:"has_images_count"`
	HasVideoCount int64  `json:"has_video_count"`
	GoodRate     float64 `json:"good_rate"`
}
