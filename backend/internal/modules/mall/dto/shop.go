// Package dto 同城商城 - 店铺 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ShopInfo 店铺详情响应
type ShopInfo struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"user_id"`
	ShopName     string     `json:"shop_name"`
	Logo         string     `json:"logo"`
	Description  string     `json:"description"`
	ShopType     string     `json:"shop_type"`
	ShopTypeText string     `json:"shop_type_text"`
	Status       int        `json:"status"`
	StatusText   string     `json:"status_text"`
	AuditStatus  int        `json:"audit_status"`
	AuditStatusText string  `json:"audit_status_text"`
	AuditReason  string     `json:"audit_reason"`
	OpenedAt     *time.Time `json:"opened_at"`

	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	Wechat       string `json:"wechat"`
	QQ           string `json:"qq"`

	Province  string  `json:"province"`
	City      string  `json:"city"`
	District  string  `json:"district"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	LicenseNo     string `json:"license_no"`
	LicenseImage  string `json:"license_image"`
	LegalPerson   string `json:"legal_person"`
	LegalPersonID string `json:"legal_person_id"`

	ProductCount  int     `json:"product_count"`
	OrderCount    int64   `json:"order_count"`
	SaleAmount    float64 `json:"sale_amount"`
	Rating        float64 `json:"rating"`
	ReviewCount   int     `json:"review_count"`
	FavoriteCount int     `json:"favorite_count"`
	ViewCount     int     `json:"view_count"`

	Featured       bool       `json:"featured"`
	Verified       bool       `json:"verified"`
	PromotionLevel int        `json:"promotion_level"`
	TrafficWeight  float64    `json:"traffic_weight"`
	VerifiedAt     *time.Time `json:"verified_at"`

	Banners       interface{} `json:"banners"`
	Tags          interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Facilities    interface{} `json:"facilities"`

	HasFaved bool `json:"has_faved"`

	RegionID  uint      `json:"region_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateShopRequest 创建店铺请求
type CreateShopRequest struct {
	ShopName     string `json:"shop_name" binding:"required,max=200"`
	Logo         string `json:"logo" binding:"max=255"`
	Description  string `json:"description"`
	ShopType     string `json:"shop_type" binding:"omitempty,oneof=personal enterprise flagship"`
	ContactName  string `json:"contact_name" binding:"max=50"`
	ContactPhone string `json:"contact_phone" binding:"required,max=32"`
	ContactEmail string `json:"contact_email" binding:"max=128"`
	Wechat       string `json:"wechat" binding:"max=64"`
	QQ           string `json:"qq" binding:"max=32"`
	Province     string `json:"province" binding:"max=64"`
	City         string `json:"city" binding:"max=64"`
	District     string `json:"district" binding:"max=64"`
	Address      string `json:"address" binding:"max=500"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	LicenseNo    string `json:"license_no" binding:"max=64"`
	LicenseImage string `json:"license_image" binding:"max=255"`
	LegalPerson  string `json:"legal_person" binding:"max=64"`
	LegalPersonID string `json:"legal_person_id" binding:"max=32"`
	Banners      interface{} `json:"banners"`
	Tags         interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Facilities   interface{} `json:"facilities"`
}

// UpdateShopRequest 更新店铺请求
type UpdateShopRequest struct {
	ShopName     *string `json:"shop_name" binding:"max=200"`
	Logo         *string `json:"logo" binding:"max=255"`
	Description  *string `json:"description"`
	ShopType     *string `json:"shop_type" binding:"omitempty,oneof=personal enterprise flagship"`
	ContactName  *string `json:"contact_name" binding:"max=50"`
	ContactPhone *string `json:"contact_phone" binding:"max=32"`
	ContactEmail *string `json:"contact_email" binding:"max=128"`
	Wechat       *string `json:"wechat" binding:"max=64"`
	QQ           *string `json:"qq" binding:"max=32"`
	Province     *string `json:"province" binding:"max=64"`
	City         *string `json:"city" binding:"max=64"`
	District     *string `json:"district" binding:"max=64"`
	Address      *string `json:"address" binding:"max=500"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	LicenseNo    *string `json:"license_no" binding:"max=64"`
	LicenseImage *string `json:"license_image" binding:"max=255"`
	LegalPerson  *string `json:"legal_person" binding:"max=64"`
	LegalPersonID *string `json:"legal_person_id" binding:"max=32"`
	Banners      interface{} `json:"banners"`
	Tags         interface{} `json:"tags"`
	BusinessHours interface{} `json:"business_hours"`
	Facilities   interface{} `json:"facilities"`
	Status       *int `json:"status"`
}

// ShopListRequest 店铺列表请求
type ShopListRequest struct {
	utils.Pagination
	Keyword     string `form:"keyword" json:"keyword"`
	ShopType    string `form:"shop_type" json:"shop_type"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	City        string `form:"city" json:"city"`
	District    string `form:"district" json:"district"`
	Verified    *bool  `form:"verified" json:"verified"`
	Featured    *bool  `form:"featured" json:"featured"`
	Sort        string `form:"sort" json:"sort"` // new/rating/sales/featured
	UserID     uint   `form:"user_id" json:"user_id"`
}

// ShopAdminListRequest 管理后台店铺列表请求
type ShopAdminListRequest struct {
	utils.Pagination
	Keyword     string `form:"keyword" json:"keyword"`
	ShopType    string `form:"shop_type" json:"shop_type"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Featured    *bool  `form:"featured" json:"featured"`
	Verified    *bool  `form:"verified" json:"verified"`
	UserID      uint   `form:"user_id" json:"user_id"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// ShopPromotionRequest 店铺推广配置请求（独立于商品推广）
type ShopPromotionRequest struct {
	Featured       *bool    `json:"featured"`
	Verified       *bool    `json:"verified"`
	PromotionLevel *int     `json:"promotion_level"`
	TrafficWeight  *float64 `json:"traffic_weight"`
}
