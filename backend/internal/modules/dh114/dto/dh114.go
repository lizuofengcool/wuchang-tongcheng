// Package dto 同城114数据传输对象 - 商户黄页主表
// 依据 v3.2.1 架构方案：对标大众点评/美团/58同城
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// Dh114Info 商户详情响应
type Dh114Info struct {
	ID              uint       `json:"id"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	CoverImage      string     `json:"cover_image"`
	UserID          uint       `json:"user_id"`
	UserName        string     `json:"user_name"`
	UserPhone       string     `json:"user_phone"`
	UserAvatar      string     `json:"user_avatar"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AuditStatus     int        `json:"audit_status"`
	AuditStatusText string     `json:"audit_status_text"`
	AuditReason     string     `json:"audit_reason"`
	PublishedAt     *time.Time `json:"published_at"`

	CategoryID   *uint  `json:"category_id"`
	CategoryName string `json:"category_name"`
	BusinessType string `json:"business_type"`
	SourceType   string `json:"source_type"`

	Phone    string `json:"phone"`
	AltPhone string `json:"alt_phone"`
	Website  string `json:"website"`
	Wechat   string `json:"wechat"`

	City             string  `json:"city"`
	District         string  `json:"district"`
	BusinessDistrict string  `json:"business_district"`
	Address          string  `json:"address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Distance         float64 `json:"distance,omitempty"`

	Rating      float64 `json:"rating"`
	ReviewCount int     `json:"review_count"`
	PriceAvg    float64 `json:"price_avg"`

	ViewCount    int        `json:"view_count"`
	FavCount     int        `json:"fav_count"`
	ContactCount int        `json:"contact_count"`
	ShareCount   int        `json:"share_count"`
	CallCount    int        `json:"call_count"`
	LastCallAt   *time.Time `json:"last_call_at"`

	ContentHash string `json:"content_hash"`
	RiskScore   int    `json:"risk_score"`

	VideoURL   string `json:"video_url"`
	VideoCover string `json:"video_cover"`
	VRURL      string `json:"vr_url"`

	Images        interface{} `json:"images"`
	Tags          interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Features      interface{} `json:"features"`

	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	PromotionLevel int     `json:"promotion_level"`
	TrafficWeight  float64 `json:"traffic_weight"`
	VerifiedAt     *time.Time `json:"verified_at"`

	HasFaved bool `json:"has_faved"`

	RegionID  uint      `json:"region_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateDh114Request 发布商户请求
type CreateDh114Request struct {
	Title       string `json:"title" binding:"required,max=200"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image"`
	CategoryID  *uint  `json:"category_id"`
	BusinessType string `json:"business_type" binding:"omitempty,oneof=restaurant retail service entertain hotel medical education life other"`
	SourceType  string `json:"source_type" binding:"omitempty,oneof=personal merchant chain"`
	Phone       string `json:"phone" binding:"required,max=32"`
	AltPhone    string `json:"alt_phone" binding:"max=32"`
	Website     string `json:"website" binding:"max=255"`
	Wechat      string `json:"wechat" binding:"max=64"`
	City        string `json:"city" binding:"max=64"`
	District    string `json:"district" binding:"max=64"`
	BusinessDistrict string `json:"business_district" binding:"max=128"`
	Address     string `json:"address" binding:"max=500"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	VideoURL    string `json:"video_url" binding:"max=255"`
	VideoCover  string `json:"video_cover" binding:"max=255"`
	VRURL       string `json:"vr_url" binding:"max=255"`
	Images      interface{} `json:"images"`
	Tags        interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Features    interface{} `json:"features"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateDh114Request 更新商户请求
type UpdateDh114Request struct {
	Title       *string `json:"title" binding:"omitempty,max=200"`
	Content     *string `json:"content"`
	CoverImage  *string `json:"cover_image" binding:"max=255"`
	CategoryID  *uint  `json:"category_id"`
	CategoryName *string `json:"category_name" binding:"max=64"`
	BusinessType *string `json:"business_type" binding:"omitempty,oneof=restaurant retail service entertain hotel medical education life other"`
	SourceType  *string `json:"source_type" binding:"omitempty,oneof=personal merchant chain"`
	Phone       *string `json:"phone" binding:"max=32"`
	AltPhone    *string `json:"alt_phone" binding:"max=32"`
	Website     *string `json:"website" binding:"max=255"`
	Wechat      *string `json:"wechat" binding:"max=64"`
	City        *string `json:"city" binding:"max=64"`
	District    *string `json:"district" binding:"max=64"`
	BusinessDistrict *string `json:"business_district" binding:"max=128"`
	Address     *string `json:"address" binding:"max=500"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	VideoURL    *string `json:"video_url" binding:"max=255"`
	VideoCover  *string `json:"video_cover" binding:"max=255"`
	VRURL       *string `json:"vr_url" binding:"max=255"`
	Images      interface{} `json:"images"`
	Tags        interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Features    interface{} `json:"features"`
	Status      *int `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// Dh114ListRequest 商户列表请求
type Dh114ListRequest struct {
	Keyword      string  `form:"keyword" json:"keyword"`
	CategoryID   uint    `form:"category_id" json:"category_id"`
	BusinessType string  `form:"business_type" json:"business_type"`
	SourceType   string  `form:"source_type" json:"source_type"`
	City         string  `form:"city" json:"city"`
	District     string  `form:"district" json:"district"`
	MinPrice     float64 `form:"min_price" json:"min_price"`
	MaxPrice     float64 `form:"max_price" json:"max_price"`
	MinRating    float64 `form:"min_rating" json:"min_rating"`
	Featured     *bool   `form:"featured" json:"featured"`
	Picked       *bool   `form:"picked" json:"picked"`
	Verified     *bool   `form:"verified" json:"verified"`
	Sort         string  `form:"sort" json:"sort"`
	utils.Pagination
}

// Dh114NearbyRequest 附近商户请求
type Dh114NearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	CategoryID uint   `form:"category_id" json:"category_id"`
	BusinessType string `form:"business_type" json:"business_type"`
	Keyword   string  `form:"keyword" json:"keyword"`
	Sort      string  `form:"sort" json:"sort"`
	utils.Pagination
}

// Dh114SearchRequest 搜索请求
type Dh114SearchRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// Dh114AdminListRequest 管理后台列表请求
type Dh114AdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	BusinessType string `form:"business_type" json:"business_type"`
	SourceType   string `form:"source_type" json:"source_type"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// PromotionRequest 推广配置请求
type PromotionRequest struct {
	Featured       *bool    `json:"featured"`
	Picked         *bool    `json:"picked"`
	Verified       *bool    `json:"verified"`
	PromotionLevel *int     `json:"promotion_level" binding:"omitempty,min=0,max=10"`
	TrafficWeight  *float64 `json:"traffic_weight" binding:"omitempty,min=0,max=9.99"`
}

// Dh114ViewRequest 浏览记录请求
type Dh114ViewRequest struct {
	Dh114ID   uint    `json:"dh114_id" binding:"required"`
	VisitType string  `json:"visit_type" binding:"omitempty,oneof=business groupbuy coupon"`
	Device    string  `json:"device"`
	Source    string  `json:"source"`
	Duration  int     `json:"duration"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// AdvancedSearchRequest 高级搜索请求
type AdvancedSearchRequest struct {
	Keyword      string  `form:"keyword" json:"keyword"`
	CategoryID   uint    `form:"category_id" json:"category_id"`
	BusinessType string  `form:"business_type" json:"business_type"`
	SourceType   string  `form:"source_type" json:"source_type"`
	City         string  `form:"city" json:"city"`
	District     string  `form:"district" json:"district"`
	MinPrice     float64 `form:"min_price" json:"min_price"`
	MaxPrice     float64 `form:"max_price" json:"max_price"`
	MinRating    float64 `form:"min_rating" json:"min_rating"`
	Featured     *bool   `form:"featured" json:"featured"`
	Picked       *bool   `form:"picked" json:"picked"`
	Verified     *bool   `form:"verified" json:"verified"`
	Sort         string  `form:"sort" json:"sort"`
	Latitude     float64 `form:"latitude" json:"latitude"`
	Longitude    float64 `form:"longitude" json:"longitude"`
	RadiusKm     float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}
