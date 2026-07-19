// Package dto 房源主表相关 DTO（发布/更新/查询/列表/详情）
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// HouseInfo 房源详情
type HouseInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Images      []HouseImageInfo `json:"images"` // 图片列表（由 service 拼装）

	// 发布者
	UserID     uint   `json:"user_id"`
	UserName   string `json:"user_name"`
	UserPhone  string `json:"user_phone"`
	UserAvatar string `json:"user_avatar"`

	// 状态
	Status      int        `json:"status"`
	AuditStatus int        `json:"audit_status"`
	AuditReason string     `json:"audit_reason"`
	PublishedAt *time.Time `json:"published_at"`
	RegionID    uint       `json:"region_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 交易类型
	ListingType  string `json:"listing_type"`
	PropertyType string `json:"property_type"`
	SourceType   string `json:"source_type"`

	// 租金相关
	RentPrice      float64 `json:"rent_price"`
	RentUnit       string  `json:"rent_unit"`
	RentType       string  `json:"rent_type"`
	DepositType    string  `json:"deposit_type"`
	PaymentMethod  string  `json:"payment_method"`
	RentNegotiable bool    `json:"rent_negotiable"`
	RentMinMonths  int     `json:"rent_min_months"`
	RentMaxMonths  int     `json:"rent_max_months"`

	// 售价相关
	SalePrice      float64 `json:"sale_price"`
	SaleNegotiable bool    `json:"sale_negotiable"`
	AveragePrice   float64 `json:"average_price"`
	OriginalPrice  float64 `json:"original_price"`

	// 户型
	Rooms     int    `json:"rooms"`
	Halls     int    `json:"halls"`
	Bathrooms int    `json:"bathrooms"`
	Kitchens  int    `json:"kitchens"`
	Balconies int    `json:"balconies"`
	Layout    string `json:"layout"`

	// 面积
	BuildingArea float64 `json:"building_area"`
	InnerArea    float64 `json:"inner_area"`
	PoolRatio    float64 `json:"pool_ratio"`
	UsableArea   float64 `json:"usable_area"`

	// 楼层/朝向
	Floor       int    `json:"floor"`
	TotalFloor  int    `json:"total_floor"`
	FloorType   string `json:"floor_type"`
	Orientation string `json:"orientation"`
	HasElevator bool   `json:"has_elevator"`

	// 装修/产权/年限
	Decoration        string `json:"decoration"`
	PropertyOwnership string `json:"property_ownership"`
	PropertyYears     int    `json:"property_years"`
	BuildingYear      int    `json:"building_year"`
	BuildingAge       int    `json:"building_age"`

	// 关联 ID
	CommunityID *uint `json:"community_id"`
	AgentID     *uint `json:"agent_id"`
	CategoryID  *uint `json:"category_id"`

	// 地理位置冗余
	City             string  `json:"city"`
	District         string  `json:"district"`
	BusinessDistrict string  `json:"business_district"`
	Address          string  `json:"address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`

	// 互动统计
	ViewCount     int        `json:"view_count"`
	FavCount      int        `json:"fav_count"`
	ContactCount  int        `json:"contact_count"`
	ShareCount    int        `json:"share_count"`
	ViewingCount  int        `json:"viewing_count"`
	LastViewingAt *time.Time `json:"last_viewing_at"`

	// 视频/VR/全景
	VideoURL    string `json:"video_url"`
	VideoCover  string `json:"video_cover"`
	VRURL       string `json:"vr_url"`
	PanoramaURL string `json:"panorama_url"`

	// 配套设施/标签/附近 POI（JSONB 解析后）
	Facilities []model.HouseFacilityItem `json:"facilities"`
	Tags       []model.HouseTagItem      `json:"tags"`
	NearbyPOIs []model.HouseNearbyPOI    `json:"nearby_pois"`

	// 运营字段
	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	PromotionLevel int     `json:"promotion_level"`
	TrafficWeight  float64 `json:"traffic_weight"`

	// 真房源认证
	RealHouseVerified   bool       `json:"real_house_verified"`
	RealHouseVerifiedAt *time.Time `json:"real_house_verified_at"`

	// 风控
	ContentHash string `json:"content_hash"`
	RiskScore   int    `json:"risk_score"`
	SameHouseID string `json:"same_house_id"`

	// 关联冗余（列表/详情返回时填充）
	CommunityName string `json:"community_name,omitempty"`
	AgentName     string `json:"agent_name,omitempty"`
	AgentAvatar   string `json:"agent_avatar,omitempty"`
	AgentPhone    string `json:"agent_phone,omitempty"`
	CategoryName  string `json:"category_name,omitempty"`

	// 当前用户是否已收藏（仅登录态列表/详情返回）
	HasFaved bool `json:"has_faved,omitempty"`

	// 仅附近查询返回（公里）
	Distance float64 `json:"distance,omitempty"`
}

// HouseImageInfo 房源图片信息
type HouseImageInfo struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	ImageType string `json:"image_type"`
	Title     string `json:"title"`
	Sort      int    `json:"sort"`
	IsCover   bool   `json:"is_cover"`
}

// CreateHouseRequest C端发布房源请求
type CreateHouseRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image" binding:"max=255"`
	Images      []HouseImageInput `json:"images"`

	// 交易类型
	ListingType  string `json:"listing_type" binding:"omitempty,oneof=rent sale transfer"`
	PropertyType string `json:"property_type" binding:"omitempty,oneof=residential apartment villa loft office shop"`
	SourceType   string `json:"source_type" binding:"omitempty,oneof=personal agent developer"`

	// 租金相关
	RentPrice      float64 `json:"rent_price" binding:"gte=0"`
	RentUnit       string  `json:"rent_unit" binding:"omitempty,oneof=month year day"`
	RentType       string  `json:"rent_type" binding:"omitempty,oneof=entire shared"`
	DepositType    string  `json:"deposit_type" binding:"omitempty,oneof=none one_month two_month three_month pay_one_deposit_three"`
	PaymentMethod  string  `json:"payment_method" binding:"omitempty,oneof=monthly quarterly half_year yearly one_time"`
	RentNegotiable bool    `json:"rent_negotiable"`
	RentMinMonths  int     `json:"rent_min_months" binding:"gte=0"`
	RentMaxMonths  int     `json:"rent_max_months" binding:"gte=0"`

	// 售价相关
	SalePrice      float64 `json:"sale_price" binding:"gte=0"`
	SaleNegotiable bool    `json:"sale_negotiable"`
	OriginalPrice  float64 `json:"original_price" binding:"gte=0"`

	// 户型
	Rooms     int `json:"rooms" binding:"gte=0"`
	Halls     int `json:"halls" binding:"gte=0"`
	Bathrooms int `json:"bathrooms" binding:"gte=0"`
	Kitchens  int `json:"kitchens" binding:"gte=0"`
	Balconies int `json:"balconies" binding:"gte=0"`

	// 面积
	BuildingArea float64 `json:"building_area" binding:"gte=0"`
	InnerArea    float64 `json:"inner_area" binding:"gte=0"`
	PoolRatio    float64 `json:"pool_ratio" binding:"gte=0,lte=1"`
	UsableArea   float64 `json:"usable_area" binding:"gte=0"`

	// 楼层/朝向
	Floor       int    `json:"floor" binding:"gte=0"`
	TotalFloor  int    `json:"total_floor" binding:"gte=0"`
	FloorType   string `json:"floor_type" binding:"omitempty,oneof=low mid high"`
	Orientation string `json:"orientation" binding:"omitempty,oneof=east south west north south_north east_west north_east north_west south_east south_west"`
	HasElevator bool   `json:"has_elevator"`

	// 装修/产权/年限
	Decoration        string `json:"decoration" binding:"omitempty,oneof=rough simple fine luxury"`
	PropertyOwnership string `json:"property_ownership" binding:"omitempty,oneof=commercial reformed affordable small_property"`
	PropertyYears     int    `json:"property_years" binding:"oneof=0 40 50 70"`
	BuildingYear      int    `json:"building_year" binding:"gte=0"`

	// 关联 ID
	CommunityID *uint `json:"community_id"`
	AgentID     *uint `json:"agent_id"`
	CategoryID  *uint `json:"category_id"`

	// 地理位置冗余
	City             string  `json:"city" binding:"max=64"`
	District         string  `json:"district" binding:"max=64"`
	BusinessDistrict string  `json:"business_district" binding:"max=128"`
	Address          string  `json:"address" binding:"max=500"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`

	// 视频/VR/全景
	VideoURL    string `json:"video_url" binding:"max=255"`
	VideoCover  string `json:"video_cover" binding:"max=255"`
	VRURL       string `json:"vr_url" binding:"max=255"`
	PanoramaURL string `json:"panorama_url" binding:"max=255"`

	// 配套设施/标签/附近 POI
	Facilities []model.HouseFacilityItem `json:"facilities"`
	Tags       []model.HouseTagItem      `json:"tags"`
	NearbyPOIs []model.HouseNearbyPOI    `json:"nearby_pois"`

	// 期限/状态
	ExpireDays int `json:"expire_days"` // 过期天数（默认 90 天）
	Status     int `json:"status" binding:"oneof=0 1"` // 0草稿 1直接发布
}

// HouseImageInput 房源图片输入
type HouseImageInput struct {
	URL       string `json:"url" binding:"required"`
	Thumbnail string `json:"thumbnail"`
	ImageType string `json:"image_type" binding:"omitempty,oneof=floor_plan real living_room bedroom kitchen bathroom balcony community certificate"`
	Title     string `json:"title"`
	Sort      int    `json:"sort"`
	IsCover   bool   `json:"is_cover"`
}

// UpdateHouseRequest 更新房源请求
type UpdateHouseRequest struct {
	Title       string `json:"title" binding:"max=200"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image" binding:"max=255"`
	Images      []HouseImageInput `json:"images"`

	ListingType  string `json:"listing_type" binding:"omitempty,oneof=rent sale transfer"`
	PropertyType string `json:"property_type" binding:"omitempty,oneof=residential apartment villa loft office shop"`
	SourceType   string `json:"source_type" binding:"omitempty,oneof=personal agent developer"`

	RentPrice      float64 `json:"rent_price" binding:"gte=0"`
	RentUnit       string  `json:"rent_unit" binding:"omitempty,oneof=month year day"`
	RentType       string  `json:"rent_type" binding:"omitempty,oneof=entire shared"`
	DepositType    string  `json:"deposit_type" binding:"omitempty,oneof=none one_month two_month three_month pay_one_deposit_three"`
	PaymentMethod  string  `json:"payment_method" binding:"omitempty,oneof=monthly quarterly half_year yearly one_time"`
	RentNegotiable *bool   `json:"rent_negotiable"`
	RentMinMonths  int     `json:"rent_min_months"`
	RentMaxMonths  int     `json:"rent_max_months"`

	SalePrice      float64 `json:"sale_price" binding:"gte=0"`
	SaleNegotiable *bool   `json:"sale_negotiable"`
	OriginalPrice  float64 `json:"original_price" binding:"gte=0"`

	Rooms     int `json:"rooms"`
	Halls     int `json:"halls"`
	Bathrooms int `json:"bathrooms"`
	Kitchens  int `json:"kitchens"`
	Balconies int `json:"balconies"`

	BuildingArea float64 `json:"building_area" binding:"gte=0"`
	InnerArea    float64 `json:"inner_area" binding:"gte=0"`
	PoolRatio    float64 `json:"pool_ratio" binding:"gte=0,lte=1"`
	UsableArea   float64 `json:"usable_area" binding:"gte=0"`

	Floor       int    `json:"floor"`
	TotalFloor  int    `json:"total_floor"`
	FloorType   string `json:"floor_type" binding:"omitempty,oneof=low mid high"`
	Orientation string `json:"orientation"`
	HasElevator *bool  `json:"has_elevator"`

	Decoration        string `json:"decoration" binding:"omitempty,oneof=rough simple fine luxury"`
	PropertyOwnership string `json:"property_ownership" binding:"omitempty,oneof=commercial reformed affordable small_property"`
	PropertyYears     int    `json:"property_years"`
	BuildingYear      int    `json:"building_year"`

	CommunityID *uint `json:"community_id"`
	AgentID     *uint `json:"agent_id"`
	CategoryID  *uint `json:"category_id"`

	City             string  `json:"city"`
	District         string  `json:"district"`
	BusinessDistrict string  `json:"business_district"`
	Address          string  `json:"address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`

	VideoURL    string `json:"video_url"`
	VideoCover  string `json:"video_cover"`
	VRURL       string `json:"vr_url"`
	PanoramaURL string `json:"panorama_url"`

	Facilities []model.HouseFacilityItem `json:"facilities"`
	Tags       []model.HouseTagItem      `json:"tags"`
	NearbyPOIs []model.HouseNearbyPOI    `json:"nearby_pois"`

	ExpireDays int `json:"expire_days"`
	Status     int `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// HouseListRequest C端房源列表查询
type HouseListRequest struct {
	CategoryID      uint    `form:"category_id" json:"category_id"`
	CommunityID     uint    `form:"community_id" json:"community_id"`
	AgentID         uint    `form:"agent_id" json:"agent_id"`
	Keyword         string  `form:"keyword" json:"keyword"`
	ListingType     string  `form:"listing_type" json:"listing_type"`
	PropertyType    string  `form:"property_type" json:"property_type"`
	SourceType      string  `form:"source_type" json:"source_type"`
	RentType        string  `form:"rent_type" json:"rent_type"`
	MinRentPrice    float64 `form:"min_rent_price" json:"min_rent_price"`
	MaxRentPrice    float64 `form:"max_rent_price" json:"max_rent_price"`
	MinSalePrice    float64 `form:"min_sale_price" json:"min_sale_price"`
	MaxSalePrice    float64 `form:"max_sale_price" json:"max_sale_price"`
	MinBuildingArea float64 `form:"min_building_area" json:"min_building_area"`
	MaxBuildingArea float64 `form:"max_building_area" json:"max_building_area"`
	Rooms           int     `form:"rooms" json:"rooms"`
	FloorType       string  `form:"floor_type" json:"floor_type"`
	Orientation     string  `form:"orientation" json:"orientation"`
	Decoration      string  `form:"decoration" json:"decoration"`
	HasElevator     *bool   `form:"has_elevator" json:"has_elevator"`
	Featured        *bool   `form:"featured" json:"featured"`
	Verified        *bool   `form:"verified" json:"verified"`
	RealHouseVerified *bool `form:"real_house_verified" json:"real_house_verified"`
	Sort            string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/popular/area_asc/area_desc
	utils.Pagination
}

// HouseNearbyRequest 附近房源查询
type HouseNearbyRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km"`
	ListingType string `form:"listing_type" json:"listing_type"`
	utils.Pagination
}

// HouseSearchRequest 房源搜索（关键词）
type HouseSearchRequest struct {
	Keyword string `form:"keyword" binding:"required,max=100"`
	utils.Pagination
}

// HouseAdvancedSearchRequest 高级搜索请求
type HouseAdvancedSearchRequest struct {
	Keyword         string  `form:"keyword" json:"keyword"`
	CategoryID      uint    `form:"category_id" json:"category_id"`
	CommunityID     uint    `form:"community_id" json:"community_id"`
	ListingType     string  `form:"listing_type" json:"listing_type"`
	PropertyType    string  `form:"property_type" json:"property_type"`
	RentType        string  `form:"rent_type" json:"rent_type"`
	MinRentPrice    float64 `form:"min_rent_price" json:"min_rent_price"`
	MaxRentPrice    float64 `form:"max_rent_price" json:"max_rent_price"`
	MinSalePrice    float64 `form:"min_sale_price" json:"min_sale_price"`
	MaxSalePrice    float64 `form:"max_sale_price" json:"max_sale_price"`
	MinBuildingArea float64 `form:"min_building_area" json:"min_building_area"`
	MaxBuildingArea float64 `form:"max_building_area" json:"max_building_area"`
	Rooms           int     `form:"rooms" json:"rooms"`
	FloorType       string  `form:"floor_type" json:"floor_type"`
	Orientation     string  `form:"orientation" json:"orientation"`
	Decoration      string  `form:"decoration" json:"decoration"`
	HasElevator     *bool   `form:"has_elevator" json:"has_elevator"`
	Featured        *bool   `form:"featured" json:"featured"`
	Verified        *bool   `form:"verified" json:"verified"`
	Latitude        float64 `form:"latitude" json:"latitude"`
	Longitude       float64 `form:"longitude" json:"longitude"`
	RadiusKm        float64 `form:"radius_km" json:"radius_km"`
	Sort            string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/distance/popular
	utils.Pagination
}

// HouseAdminListRequest 管理后台列表查询
type HouseAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	CommunityID uint   `form:"community_id" json:"community_id"`
	ListingType string `form:"listing_type" json:"listing_type"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// HouseDetailResponse 聚合详情响应（主信息 + 图片 + 经纪人 + 小区）
type HouseDetailResponse struct {
	HouseInfo
	Agent     *AgentResponse     `json:"agent,omitempty"`
	Community *CommunityResponse `json:"community,omitempty"`
}
