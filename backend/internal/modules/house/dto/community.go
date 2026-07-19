// Package dto 小区信息 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// CommunityResponse 小区信息响应
type CommunityResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Alias           string     `json:"alias"`
	City            string     `json:"city"`
	District        string     `json:"district"`
	BusinessDistrict string    `json:"business_district"`
	Address         string     `json:"address"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	BuildingCount   int        `json:"building_count"`
	HouseCount      int        `json:"house_count"`
	BuildingYear    int        `json:"building_year"`
	BuildingType    string     `json:"building_type"`
	Developer       string     `json:"developer"`
	PropertyCompany string     `json:"property_company"`
	PropertyFee     float64    `json:"property_fee"`
	ParkingRatio    string     `json:"parking_ratio"`
	GreeningRate    float64    `json:"greening_rate"`
	PlotRatio       float64    `json:"plot_ratio"`
	AvgSalePrice    float64    `json:"avg_sale_price"`
	AvgRentPrice    float64    `json:"avg_rent_price"`
	Description     string     `json:"description"`
	CoverImage      string     `json:"cover_image"`
	Images          []model.CommunityImage `json:"images"`
	NearbyPOIs      []model.HouseNearbyPOI `json:"nearby_pois"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	FollowerCount   int        `json:"follower_count"`
	OnSaleCount     int        `json:"on_sale_count"`
	OnRentCount     int        `json:"on_rent_count"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	HasFollowed     bool       `json:"has_faved,omitempty"` // 当前用户是否已关注
	Distance        float64    `json:"distance,omitempty"`  // 附近查询返回（公里）
}

// CommunityCreateRequest 创建/更新小区请求
type CommunityCreateRequest struct {
	Name            string                   `json:"name" binding:"required,max=128"`
	Alias           string                   `json:"alias" binding:"max=128"`
	City            string                   `json:"city" binding:"max=64"`
	District        string                   `json:"district" binding:"max=64"`
	BusinessDistrict string                   `json:"business_district" binding:"max=128"`
	Address         string                   `json:"address" binding:"max=500"`
	Latitude        float64                  `json:"latitude"`
	Longitude       float64                  `json:"longitude"`
	BuildingCount   int                      `json:"building_count" binding:"gte=0"`
	HouseCount      int                      `json:"house_count" binding:"gte=0"`
	BuildingYear    int                      `json:"building_year" binding:"gte=0"`
	BuildingType    string                   `json:"building_type" binding:"omitempty,oneof=plate tower plate_tower villa bungalow mixed"`
	Developer       string                   `json:"developer" binding:"max=128"`
	PropertyCompany string                   `json:"property_company" binding:"max=128"`
	PropertyFee     float64                  `json:"property_fee" binding:"gte=0"`
	ParkingRatio    string                   `json:"parking_ratio" binding:"max=32"`
	GreeningRate    float64                  `json:"greening_rate" binding:"gte=0,lte=1"`
	PlotRatio       float64                  `json:"plot_ratio" binding:"gte=0"`
	AvgSalePrice    float64                  `json:"avg_sale_price" binding:"gte=0"`
	AvgRentPrice    float64                  `json:"avg_rent_price" binding:"gte=0"`
	Description     string                   `json:"description"`
	CoverImage      string                   `json:"cover_image" binding:"max=255"`
	Images          []model.CommunityImage   `json:"images"`
	NearbyPOIs      []model.HouseNearbyPOI   `json:"nearby_pois"`
	Status          int                      `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// CommunityListQuery 小区列表查询（C 端）
type CommunityListQuery struct {
	City            string  `form:"city" json:"city"`
	District        string  `form:"district" json:"district"`
	BusinessDistrict string  `form:"business_district" json:"business_district"`
	BuildingType    string  `form:"building_type" json:"building_type"`
	Keyword         string  `form:"keyword" json:"keyword"`
	Latitude        float64 `form:"latitude" json:"latitude"`
	Longitude       float64 `form:"longitude" json:"longitude"`
	RadiusKm        float64 `form:"radius_km" json:"radius_km"`
	Sort            string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/follower/house_count
	utils.Pagination
}

// CommunityAdminListQuery 管理后台小区列表查询
type CommunityAdminListQuery struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	City     string `form:"city" json:"city"`
	District string `form:"district" json:"district"`
	Status   *int   `form:"status" json:"status"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// CommunityFollowRequest 关注小区请求
type CommunityFollowRequest struct {
	Notify bool `json:"notify"`
}
