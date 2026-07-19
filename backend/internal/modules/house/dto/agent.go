// Package dto 经纪人 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// AgentResponse 经纪人详情响应
type AgentResponse struct {
	ID            uint      `json:"id"`
	UserID        uint      `json:"user_id"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	Avatar        string    `json:"avatar"`
	Gender        string    `json:"gender"`
	StoreID       uint      `json:"store_id"`
	StoreName     string    `json:"store_name"`
	Company       string    `json:"company"`
	Title         string    `json:"title"`
	Level         int       `json:"level"`
	LevelText     string    `json:"level_text"`
	LicenseNo     string    `json:"license_no"`
	LicenseImage  string    `json:"license_image"`
	IDCardFront   string    `json:"id_card_front"`
	IDCardBack    string    `json:"id_card_back"`
	BusinessCard  string    `json:"business_card"`
	Description   string    `json:"description"`
	GoodAt        []model.AgentGoodAt      `json:"good_at"`
	ServiceArea   []model.AgentServiceArea `json:"service_area"`
	Rating        float64   `json:"rating"`
	RatingCount   int       `json:"rating_count"`
	ListingCount  int       `json:"listing_count"`
	DealCount     int       `json:"deal_count"`
	TotalAmount   float64   `json:"total_amount"`
	ResponseTime  int       `json:"response_time"`
	ResponseRate  float64   `json:"response_rate"`
	OnlineStatus  int       `json:"online_status"`
	OnlineText    string    `json:"online_text"`
	LastActiveAt  *time.Time `json:"last_active_at"`
	VerifiedAt    *time.Time `json:"verified_at"`
	ApprovedAt    *time.Time `json:"approved_at"`
	RejectedReason string   `json:"rejected_reason"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	FollowerCount int       `json:"follower_count"`
	Tags          []model.HouseTagItem `json:"tags"`
	RegionID      uint      `json:"region_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	HasFollowed   bool      `json:"has_faved,omitempty"` // 当前用户是否已关注
}

// AgentCreateRequest 创建/更新经纪人请求
type AgentCreateRequest struct {
	Name          string                   `json:"name" binding:"required,max=50"`
	Phone         string                   `json:"phone" binding:"required,max=20"`
	Avatar        string                   `json:"avatar" binding:"max=255"`
	Gender        string                   `json:"gender" binding:"omitempty,oneof=unlimited male female"`
	StoreID       uint                     `json:"store_id"`
	StoreName     string                   `json:"store_name" binding:"max=128"`
	Company       string                   `json:"company" binding:"max=128"`
	Title         string                   `json:"title" binding:"max=64"`
	Level         int                      `json:"level" binding:"gte=0,lte=4"`
	LicenseNo     string                   `json:"license_no" binding:"max=64"`
	LicenseImage  string                   `json:"license_image" binding:"max=255"`
	IDCardFront   string                   `json:"id_card_front" binding:"max=255"`
	IDCardBack    string                   `json:"id_card_back" binding:"max=255"`
	BusinessCard  string                   `json:"business_card" binding:"max=255"`
	Description   string                   `json:"description"`
	GoodAt        []model.AgentGoodAt      `json:"good_at"`
	ServiceArea   []model.AgentServiceArea `json:"service_area"`
	Tags          []model.HouseTagItem     `json:"tags"`
}

// AgentListQuery 经纪人列表查询
type AgentListQuery struct {
	City        string  `form:"city" json:"city"`
	Company     string  `form:"company" json:"company"`
	StoreID     uint    `form:"store_id" json:"store_id"`
	Level       *int    `form:"level" json:"level"`
	Status      *int    `form:"status" json:"status"`
	OnlineStatus *int   `form:"online_status" json:"online_status"`
	Keyword     string  `form:"keyword" json:"keyword"`
	Latitude    float64 `form:"latitude" json:"latitude"`
	Longitude   float64 `form:"longitude" json:"longitude"`
	RadiusKm    float64 `form:"radius_km" json:"radius_km"`
	Sort        string  `form:"sort" json:"sort"` // latest/rating/deal_count/listing_count
	utils.Pagination
}

// AgentAdminListQuery 管理后台经纪人列表查询
type AgentAdminListQuery struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	UserID   uint   `form:"user_id" json:"user_id"`
	Status   *int   `form:"status" json:"status"`
	Level    *int   `form:"level" json:"level"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// AgentAuditRequest 经纪人审核请求
type AgentAuditRequest struct {
	Status  int    `json:"status" binding:"oneof=1 2 3 4"` // 1通过 2拒绝 3冻结 4撤销
	Reason  string `json:"reason" binding:"max=500"`
}

// AgentFollowRequest 关注经纪人请求
type AgentFollowRequest struct {
	Notify bool `json:"notify"`
}
