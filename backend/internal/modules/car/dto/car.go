// Package dto 同城车辆买卖数据传输对象 - 车源主表
// 依据 v3.2.1 架构方案第六章：对标瓜子/人人车/懂车帝/毛豆新车/易鑫车贷
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// CarInfo 车源详情响应
type CarInfo struct {
	ID                  uint       `json:"id"`
	Title               string     `json:"title"`
	Content             string     `json:"content"`
	CoverImage          string     `json:"cover_image"`
	Images              []CarImageInfo `json:"images"`
	UserID              uint       `json:"user_id"`
	UserName            string     `json:"user_name"`
	UserPhone           string     `json:"user_phone"`
	UserAvatar          string     `json:"user_avatar"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	AuditStatus         int        `json:"audit_status"`
	AuditStatusText     string     `json:"audit_status_text"`
	AuditReason         string     `json:"audit_reason"`
	PublishedAt         *time.Time `json:"published_at"`

	ListingType string `json:"listing_type"`
	SourceType  string `json:"source_type"`
	CarType     string `json:"car_type"`

	BrandID    *uint  `json:"brand_id"`
	BrandName  string `json:"brand_name"`
	ModelID    *uint  `json:"model_id"`
	ModelName  string `json:"model_name"`
	Series     string `json:"series"`
	CategoryID *uint  `json:"category_id"`

	Price           float64 `json:"price"`
	OriginalPrice   float64 `json:"original_price"`
	AveragePrice    float64 `json:"average_price"`
	PriceNegotiable bool    `json:"price_negotiable"`
	DealerPrice     float64 `json:"dealer_price"`

	RegistrationYear      int        `json:"registration_year"`
	RegistrationMonth     int        `json:"registration_month"`
	FirstRegistrationDate *time.Time `json:"first_registration_date"`
	Mileage               float64    `json:"mileage"`
	MileageUnit           string     `json:"mileage_unit"`

	Displacement     float64 `json:"displacement"`
	Transmission     string  `json:"transmission"`
	FuelType         string  `json:"fuel_type"`
	EmissionStandard string  `json:"emission_standard"`
	EngineType       string  `json:"engine_type"`
	Horsepower       int     `json:"horsepower"`

	ExteriorColor string `json:"exterior_color"`
	InteriorColor string `json:"interior_color"`
	SeatCount     int    `json:"seat_count"`
	DoorCount     int    `json:"door_count"`

	ConditionLevel  string `json:"condition_level"`
	ConditionScore  int    `json:"condition_score"`
	AccidentCount   int    `json:"accident_count"`

	TransferCount          int        `json:"transfer_count"`
	LastTransferDate       *time.Time `json:"last_transfer_date"`
	AnnualInspectionDue    *time.Time `json:"annual_inspection_due"`
	AnnualInspectionStatus string     `json:"annual_inspection_status"`
	InsuranceDue           *time.Time `json:"insurance_due"`
	InsuranceStatus        string     `json:"insurance_status"`
	CommercialInsuranceDue *time.Time `json:"commercial_insurance_due"`

	VIN             string `json:"vin"`
	LicensePlate    string `json:"license_plate"`
	LicenseLocation string `json:"license_location"`
	EngineNo        string `json:"engine_no"`

	UseType                string  `json:"use_type"`
	MaintenanceCount       int     `json:"maintenance_count"`
	LastMaintenanceMileage float64 `json:"last_maintenance_mileage"`

	City             string  `json:"city"`
	District         string  `json:"district"`
	BusinessDistrict string  `json:"business_district"`
	Address          string  `json:"address"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Distance         float64 `json:"distance,omitempty"`

	ViewCount       int        `json:"view_count"`
	FavCount        int        `json:"fav_count"`
	ContactCount    int        `json:"contact_count"`
	ShareCount      int        `json:"share_count"`
	TestDriveCount  int        `json:"test_drive_count"`
	LastTestDriveAt *time.Time `json:"last_test_drive_at"`

	ContentHash string `json:"content_hash"`
	RiskScore   int    `json:"risk_score"`
	SameCarID   string `json:"same_car_id"`

	VideoURL       string `json:"video_url"`
	VideoCover     string `json:"video_cover"`
	Panorama360URL string `json:"panorama_360_url"`

	Features        interface{} `json:"features"`
	Tags            interface{} `json:"tags"`
	InspectionItems interface{} `json:"inspection_items"`
	AccidentHistory interface{} `json:"accident_history"`

	Featured       bool    `json:"featured"`
	Picked         bool    `json:"picked"`
	Verified       bool    `json:"verified"`
	PromotionLevel int     `json:"promotion_level"`
	TrafficWeight  float64 `json:"traffic_weight"`

	RealCarVerified   bool       `json:"real_car_verified"`
	RealCarVerifiedAt *time.Time `json:"real_car_verified_at"`

	HasFaved bool `json:"has_faved"`

	RegionID   uint      `json:"region_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CarImageInfo 车源图片信息
type CarImageInfo struct {
	ID          uint   `json:"id"`
	ImageType   string `json:"image_type"`
	URL         string `json:"url"`
	Thumbnail   string `json:"thumbnail"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	IsCover     bool   `json:"is_cover"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int    `json:"size"`
	Tag         string `json:"tag"`
}

// CreateCarRequest 发布车源请求
type CreateCarRequest struct {
	Title       string `json:"title" binding:"required,max=200"`
	Content     string `json:"content"`
	CoverImage  string `json:"cover_image"`
	ListingType string `json:"listing_type" binding:"omitempty,oneof=new used replace rental"`
	SourceType  string `json:"source_type" binding:"omitempty,oneof=personal dealer manufacturer"`
	CarType     string `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`

	BrandID    *uint  `json:"brand_id"`
	BrandName  string `json:"brand_name" binding:"max=64"`
	ModelID    *uint  `json:"model_id"`
	ModelName  string `json:"model_name" binding:"max=128"`
	Series     string `json:"series" binding:"max=64"`
	CategoryID *uint  `json:"category_id"`

	Price           float64 `json:"price" binding:"min=0"`
	OriginalPrice   float64 `json:"original_price"`
	PriceNegotiable bool    `json:"price_negotiable"`
	DealerPrice     float64 `json:"dealer_price"`

	RegistrationYear  int        `json:"registration_year"`
	RegistrationMonth int        `json:"registration_month"`
	FirstRegistrationDate *time.Time `json:"first_registration_date"`
	Mileage           float64    `json:"mileage" binding:"min=0"`
	MileageUnit       string     `json:"mileage_unit" binding:"omitempty,oneof=km mile"`

	Displacement     float64 `json:"displacement"`
	Transmission     string  `json:"transmission" binding:"omitempty,oneof=manual auto cvt dct amt"`
	FuelType         string  `json:"fuel_type" binding:"omitempty,oneof=gasoline diesel hybrid pure_electric range_extender hydrogen"`
	EmissionStandard string  `json:"emission_standard" binding:"omitempty,oneof=china_1 china_2 china_3 china_4 china_5 china_6"`
	EngineType       string  `json:"engine_type" binding:"max=64"`
	Horsepower       int     `json:"horsepower"`

	ExteriorColor string `json:"exterior_color" binding:"max=32"`
	InteriorColor string `json:"interior_color" binding:"max=32"`
	SeatCount     int    `json:"seat_count"`
	DoorCount     int    `json:"door_count"`

	ConditionLevel string `json:"condition_level" binding:"omitempty,oneof=A B C D"`

	TransferCount          int        `json:"transfer_count"`
	LastTransferDate       *time.Time `json:"last_transfer_date"`
	AnnualInspectionDue    *time.Time `json:"annual_inspection_due"`
	AnnualInspectionStatus string     `json:"annual_inspection_status" binding:"omitempty,oneof=valid expiring expired none"`
	InsuranceDue           *time.Time `json:"insurance_due"`
	InsuranceStatus        string     `json:"insurance_status" binding:"omitempty,oneof=valid expiring expired none"`
	CommercialInsuranceDue *time.Time `json:"commercial_insurance_due"`

	VIN             string `json:"vin" binding:"max=32"`
	LicensePlate    string `json:"license_plate" binding:"max=32"`
	LicenseLocation string `json:"license_location" binding:"max=64"`
	EngineNo        string `json:"engine_no" binding:"max=64"`

	UseType                string  `json:"use_type" binding:"omitempty,oneof=non_operational operational"`
	MaintenanceCount       int     `json:"maintenance_count"`
	LastMaintenanceMileage float64 `json:"last_maintenance_mileage"`

	City             string  `json:"city" binding:"max=64"`
	District         string  `json:"district" binding:"max=64"`
	BusinessDistrict string  `json:"business_district" binding:"max=128"`
	Address          string  `json:"address" binding:"max=500"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`

	VideoURL       string `json:"video_url" binding:"max=255"`
	VideoCover     string `json:"video_cover" binding:"max=255"`
	Panorama360URL string `json:"panorama_360_url" binding:"max=255"`

	Features        interface{} `json:"features"`
	Tags            interface{} `json:"tags"`
	InspectionItems interface{} `json:"inspection_items"`
	AccidentHistory interface{} `json:"accident_history"`

	Images []CarImageInput `json:"images"`

	Status int `json:"status" binding:"omitempty,oneof=0 1"`
}

// CarImageInput 图片输入
type CarImageInput struct {
	ImageType   string `json:"image_type" binding:"omitempty,oneof=exterior interior engine chassis accident document dashboard wheel trunk other"`
	URL         string `json:"url" binding:"required"`
	Thumbnail   string `json:"thumbnail"`
	Title       string `json:"title" binding:"max=128"`
	Description string `json:"description" binding:"max=500"`
	Sort        int    `json:"sort"`
	IsCover     bool   `json:"is_cover"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int    `json:"size"`
	Tag         string `json:"tag" binding:"max=64"`
}

// UpdateCarRequest 更新车源请求
type UpdateCarRequest struct {
	Title       *string `json:"title" binding:"omitempty,max=200"`
	Content     *string `json:"content"`
	CoverImage  *string `json:"cover_image" binding:"omitempty,max=255"`
	ListingType *string `json:"listing_type" binding:"omitempty,oneof=new used replace rental"`
	SourceType  *string `json:"source_type" binding:"omitempty,oneof=personal dealer manufacturer"`
	CarType     *string `json:"car_type" binding:"omitempty,oneof=sedan suv mpv new_energy sports truck van bus"`

	BrandID    *uint  `json:"brand_id"`
	BrandName  *string `json:"brand_name" binding:"omitempty,max=64"`
	ModelID    *uint  `json:"model_id"`
	ModelName  *string `json:"model_name" binding:"omitempty,max=128"`
	Series     *string `json:"series" binding:"omitempty,max=64"`
	CategoryID *uint  `json:"category_id"`

	Price           *float64 `json:"price" binding:"omitempty,min=0"`
	OriginalPrice   *float64 `json:"original_price"`
	PriceNegotiable *bool    `json:"price_negotiable"`
	DealerPrice     *float64 `json:"dealer_price"`

	RegistrationYear  *int        `json:"registration_year"`
	RegistrationMonth *int        `json:"registration_month"`
	FirstRegistrationDate *time.Time `json:"first_registration_date"`
	Mileage           *float64    `json:"mileage" binding:"omitempty,min=0"`
	MileageUnit       *string     `json:"mileage_unit" binding:"omitempty,oneof=km mile"`

	Displacement     *float64 `json:"displacement"`
	Transmission     *string  `json:"transmission" binding:"omitempty,oneof=manual auto cvt dct amt"`
	FuelType         *string  `json:"fuel_type" binding:"omitempty,oneof=gasoline diesel hybrid pure_electric range_extender hydrogen"`
	EmissionStandard *string  `json:"emission_standard" binding:"omitempty,oneof=china_1 china_2 china_3 china_4 china_5 china_6"`
	EngineType       *string  `json:"engine_type" binding:"omitempty,max=64"`
	Horsepower       *int     `json:"horsepower"`

	ExteriorColor *string `json:"exterior_color" binding:"omitempty,max=32"`
	InteriorColor *string `json:"interior_color" binding:"omitempty,max=32"`
	SeatCount     *int    `json:"seat_count"`
	DoorCount     *int    `json:"door_count"`

	ConditionLevel *string `json:"condition_level" binding:"omitempty,oneof=A B C D"`
	ConditionScore *int    `json:"condition_score"`
	AccidentCount  *int    `json:"accident_count"`

	TransferCount          *int        `json:"transfer_count"`
	LastTransferDate       *time.Time  `json:"last_transfer_date"`
	AnnualInspectionDue    *time.Time  `json:"annual_inspection_due"`
	AnnualInspectionStatus *string     `json:"annual_inspection_status" binding:"omitempty,oneof=valid expiring expired none"`
	InsuranceDue           *time.Time  `json:"insurance_due"`
	InsuranceStatus        *string     `json:"insurance_status" binding:"omitempty,oneof=valid expiring expired none"`
	CommercialInsuranceDue *time.Time  `json:"commercial_insurance_due"`

	VIN             *string `json:"vin" binding:"omitempty,max=32"`
	LicensePlate    *string `json:"license_plate" binding:"omitempty,max=32"`
	LicenseLocation *string `json:"license_location" binding:"omitempty,max=64"`
	EngineNo        *string `json:"engine_no" binding:"omitempty,max=64"`

	UseType                *string  `json:"use_type" binding:"omitempty,oneof=non_operational operational"`
	MaintenanceCount       *int     `json:"maintenance_count"`
	LastMaintenanceMileage *float64 `json:"last_maintenance_mileage"`

	City             *string `json:"city" binding:"omitempty,max=64"`
	District         *string `json:"district" binding:"omitempty,max=64"`
	BusinessDistrict *string `json:"business_district" binding:"omitempty,max=128"`
	Address          *string `json:"address" binding:"omitempty,max=500"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`

	VideoURL       *string `json:"video_url" binding:"omitempty,max=255"`
	VideoCover     *string `json:"video_cover" binding:"omitempty,max=255"`
	Panorama360URL *string `json:"panorama_360_url" binding:"omitempty,max=255"`

	Features        interface{} `json:"features"`
	Tags            interface{} `json:"tags"`
	InspectionItems interface{} `json:"inspection_items"`
	AccidentHistory interface{} `json:"accident_history"`

	Images *[]CarImageInput `json:"images"`

	Status *int `json:"status" binding:"omitempty,oneof=0 1 2 3 5"`
}

// CarListRequest C 端车源列表请求
type CarListRequest struct {
	Keyword        string  `form:"keyword" json:"keyword"`
	CategoryID     uint    `form:"category_id" json:"category_id"`
	BrandID        uint    `form:"brand_id" json:"brand_id"`
	ModelID        uint    `form:"model_id" json:"model_id"`
	CarType        string  `form:"car_type" json:"car_type"`
	ListingType    string  `form:"listing_type" json:"listing_type"`
	SourceType     string  `form:"source_type" json:"source_type"`
	FuelType       string  `form:"fuel_type" json:"fuel_type"`
	Transmission   string  `form:"transmission" json:"transmission"`
	MinPrice       float64 `form:"min_price" json:"min_price"`
	MaxPrice       float64 `form:"max_price" json:"max_price"`
	MinMileage     float64 `form:"min_mileage" json:"min_mileage"`
	MaxMileage     float64 `form:"max_mileage" json:"max_mileage"`
	MinYear        int     `form:"min_year" json:"min_year"`
	MaxYear        int     `form:"max_year" json:"max_year"`
	ConditionLevel string  `form:"condition_level" json:"condition_level"`
	City           string  `form:"city" json:"city"`
	Featured       *bool   `form:"featured" json:"featured"`
	Picked         *bool   `form:"picked" json:"picked"`
	Verified       *bool   `form:"verified" json:"verified"`
	RealCarVerified *bool  `form:"real_car_verified" json:"real_car_verified"`
	PriceNegotiable *bool  `form:"price_negotiable" json:"price_negotiable"`
	Sort           string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/mileage_asc/year_desc/popular
	utils.Pagination
}

// CarNearbyRequest 附近车源请求
type CarNearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required,min=-180,max=180"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	CategoryID uint   `form:"category_id" json:"category_id"`
	BrandID   uint    `form:"brand_id" json:"brand_id"`
	CarType   string  `form:"car_type" json:"car_type"`
	MinPrice  float64 `form:"min_price" json:"min_price"`
	MaxPrice  float64 `form:"max_price" json:"max_price"`
	Sort      string  `form:"sort" json:"sort"`
	utils.Pagination
}

// CarSearchRequest 搜索请求
type CarSearchRequest struct {
	Keyword string `form:"keyword" json:"keyword" binding:"required"`
	utils.Pagination
}

// CarAdminListRequest 管理后台列表请求
type CarAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	BrandID     uint   `form:"brand_id" json:"brand_id"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	ListingType string `form:"listing_type" json:"listing_type"`
	SourceType  string `form:"source_type" json:"source_type"`
	CarType     string `form:"car_type" json:"car_type"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// RealCarVerifyRequest 真车认证请求（M 端）
type RealCarVerifyRequest struct {
	Verified bool   `json:"verified"`
	Reason   string `json:"reason" binding:"max=500"`
}

// PromotionRequest 推广配置请求（M 端）
type PromotionRequest struct {
	Featured       *bool    `json:"featured"`
	Picked         *bool    `json:"picked"`
	Verified       *bool    `json:"verified"`
	PromotionLevel *int     `json:"promotion_level" binding:"omitempty,min=0,max=10"`
	TrafficWeight  *float64 `json:"traffic_weight" binding:"omitempty,min=0,max=9.99"`
}
