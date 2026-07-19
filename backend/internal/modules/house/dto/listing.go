// Package dto 房源发布 DTO（与 houses 主表 1:1 冗余发布信息）
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ListingInfo 发布信息响应
type ListingInfo struct {
	ID              uint       `json:"id"`
	ListingNo       string     `json:"listing_no"`
	HouseID         uint       `json:"house_id"`
	CommunityID     uint       `json:"community_id"`
	AgentID         uint       `json:"agent_id"`
	PublisherID     uint       `json:"publisher_id"`
	PublisherName   string     `json:"publisher_name"`
	PublisherPhone  string     `json:"publisher_phone"`
	PublisherAvatar string     `json:"publisher_avatar"`
	PublisherType   string     `json:"publisher_type"`
	ListingType     string     `json:"listing_type"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Price           float64    `json:"price"`
	PriceUnit       string     `json:"price_unit"`
	Decoration      string     `json:"decoration"`
	Orientation     string     `json:"orientation"`
	Layout          string     `json:"layout"`
	BuildingArea    float64    `json:"building_area"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AuditStatus     int        `json:"audit_status"`
	AuditReason     string     `json:"audit_reason"`
	PublishedAt     *time.Time `json:"published_at"`
	ExpiredAt       *time.Time `json:"expired_at"`
	RefreshedAt     *time.Time `json:"refreshed_at"`
	OfflineAt       *time.Time `json:"offline_at"`
	RefreshCount    int        `json:"refresh_count"`
	ViewCount       int        `json:"view_count"`
	FavCount        int        `json:"fav_count"`
	ContactCount    int        `json:"contact_count"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateListingRequest 创建发布请求
type CreateListingRequest struct {
	HouseID       uint   `json:"house_id" binding:"required"`
	CommunityID   uint   `json:"community_id"`
	AgentID       uint   `json:"agent_id"`
	PublisherType string `json:"publisher_type" binding:"omitempty,oneof=personal agent developer"`
	ListingType   string `json:"listing_type" binding:"omitempty,oneof=rent sale transfer"`
	Title         string `json:"title" binding:"required,max=200"`
	Description   string `json:"description"`
	Price         float64 `json:"price" binding:"gte=0"`
	PriceUnit     string `json:"price_unit" binding:"omitempty,oneof=month year day"`
	Decoration    string `json:"decoration" binding:"omitempty,oneof=rough simple fine luxury"`
	Orientation   string `json:"orientation"`
	Layout        string `json:"layout" binding:"max=32"`
	BuildingArea  float64 `json:"building_area" binding:"gte=0"`
	ExpireDays    int     `json:"expire_days"` // 过期天数（默认 90）
	Status        int     `json:"status" binding:"oneof=0 1"` // 0草稿 1直接发布
}

// UpdateListingRequest 更新发布请求
type UpdateListingRequest struct {
	Title        string  `json:"title" binding:"max=200"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" binding:"gte=0"`
	PriceUnit    string  `json:"price_unit" binding:"omitempty,oneof=month year day"`
	Decoration   string  `json:"decoration"`
	Orientation  string  `json:"orientation"`
	Layout       string  `json:"layout"`
	BuildingArea float64 `json:"building_area" binding:"gte=0"`
	ExpireDays   int     `json:"expire_days"`
	Status       int     `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// ListingListQuery 发布列表查询（C 端）
type ListingListQuery struct {
	HouseID       uint   `form:"house_id" json:"house_id"`
	CommunityID   uint   `form:"community_id" json:"community_id"`
	AgentID       uint   `form:"agent_id" json:"agent_id"`
	PublisherID   uint   `form:"publisher_id" json:"publisher_id"`
	PublisherType string `form:"publisher_type" json:"publisher_type"`
	ListingType   string `form:"listing_type" json:"listing_type"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	Sort          string `form:"sort" json:"sort"` // latest/price_asc/price_desc/popular
	utils.Pagination
}

// ListingAdminListQuery 管理后台发布列表查询
type ListingAdminListQuery struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	HouseID     uint   `form:"house_id" json:"house_id"`
	PublisherID uint   `form:"publisher_id" json:"publisher_id"`
	ListingType string `form:"listing_type" json:"listing_type"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ListingRefreshRequest 刷新发布请求
type ListingRefreshRequest struct {
	// 无参数，仅触发刷新
}

// ListingAuditRequest 发布审核请求
type ListingAuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}
