// Package dto 同城车辆买卖数据传输对象 - 车源发布单
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ListingInfo 发布单详情响应
type ListingInfo struct {
	ID                uint       `json:"id"`
	ListingNo         string     `json:"listing_no"`
	CarID             uint       `json:"car_id"`
	ModelID           uint       `json:"model_id"`
	PublisherID       uint       `json:"publisher_id"`
	PublisherName     string     `json:"publisher_name"`
	PublisherAvatar   string     `json:"publisher_avatar"`
	PublisherType     string     `json:"publisher_type"`
	DealerID          uint       `json:"dealer_id"`
	DealerName        string     `json:"dealer_name"`
	ListingType       string     `json:"listing_type"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Price             float64    `json:"price"`
	OriginalPrice     float64    `json:"original_price"`
	PriceNegotiable   bool       `json:"price_negotiable"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	AuditStatus       int        `json:"audit_status"`
	AuditStatusText   string     `json:"audit_status_text"`
	AuditReason       string     `json:"audit_reason"`
	PublishedAt       *time.Time `json:"published_at"`
	OfflineAt         *time.Time `json:"offline_at"`
	ExpiredAt         *time.Time `json:"expired_at"`
	SoldAt            *time.Time `json:"sold_at"`
	ViewCount         int        `json:"view_count"`
	FavCount          int        `json:"fav_count"`
	ContactCount      int        `json:"contact_count"`
	TestDriveCount    int        `json:"test_drive_count"`
	InspectionStatus  int        `json:"inspection_status"`
	InspectionStatusText string  `json:"inspection_status_text"`
	InspectionID      uint       `json:"inspection_id"`
	RealCarVerified   bool       `json:"real_car_verified"`
	Featured          bool       `json:"featured"`
	PromotionLevel    int        `json:"promotion_level"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateListingRequest 创建发布单请求
type CreateListingRequest struct {
	CarID        uint   `json:"car_id" binding:"required"`
	ModelID      uint   `json:"model_id"`
	PublisherType string `json:"publisher_type" binding:"omitempty,oneof=personal dealer agent"`
	DealerID     uint   `json:"dealer_id"`
	DealerName   string `json:"dealer_name" binding:"max=128"`
	ListingType  string `json:"listing_type" binding:"omitempty,oneof=new used replace rental"`
	Title        string `json:"title" binding:"required,max=200"`
	Description  string `json:"description"`
	Price        float64 `json:"price" binding:"min=0"`
	OriginalPrice float64 `json:"original_price"`
	PriceNegotiable bool `json:"price_negotiable"`
}

// UpdateListingRequest 更新发布单请求
type UpdateListingRequest struct {
	Title           *string  `json:"title" binding:"omitempty,max=200"`
	Description     *string  `json:"description"`
	Price           *float64 `json:"price" binding:"omitempty,min=0"`
	OriginalPrice   *float64 `json:"original_price"`
	PriceNegotiable *bool    `json:"price_negotiable"`
	Status          *int     `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// ListingListRequest 发布单列表请求
type ListingListRequest struct {
	CarID            uint   `form:"car_id" json:"car_id"`
	PublisherID      uint   `form:"publisher_id" json:"publisher_id"`
	DealerID         uint   `form:"dealer_id" json:"dealer_id"`
	ListingType      string `form:"listing_type" json:"listing_type"`
	PublisherType    string `form:"publisher_type" json:"publisher_type"`
	Status           *int   `form:"status" json:"status"`
	AuditStatus      *int   `form:"audit_status" json:"audit_status"`
	InspectionStatus *int   `form:"inspection_status" json:"inspection_status"`
	Featured         *bool  `form:"featured" json:"featured"`
	RealCarVerified  *bool  `form:"real_car_verified" json:"real_car_verified"`
	Keyword          string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ListingAdminListRequest 管理后台发布单列表请求
type ListingAdminListRequest struct {
	RegionID      uint   `form:"region_id" json:"region_id"`
	PublisherID   uint   `form:"publisher_id" json:"publisher_id"`
	DealerID      uint   `form:"dealer_id" json:"dealer_id"`
	ListingType   string `form:"listing_type" json:"listing_type"`
	Status        *int   `form:"status" json:"status"`
	AuditStatus   *int   `form:"audit_status" json:"audit_status"`
	InspectionStatus *int `form:"inspection_status" json:"inspection_status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ListingAuditRequest 发布单审核请求
type ListingAuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// InspectionStatusUpdateRequest 检测状态更新请求
type InspectionStatusUpdateRequest struct {
	InspectionStatus int    `json:"inspection_status" binding:"oneof=0 1 2 3 4"`
	InspectionID     uint   `json:"inspection_id"`
	Reason           string `json:"reason" binding:"max=500"`
}
