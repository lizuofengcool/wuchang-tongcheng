// Package dto 同城拼车出行数据传输对象 - 拼车主表
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PincheInfo 拼车行程详情响应
type PincheInfo struct {
	ID              uint       `json:"id"`
	RegionID        uint       `json:"region_id"`
	UserID          uint       `json:"user_id"`
	UserName        string     `json:"user_name"`
	UserPhone       string     `json:"user_phone"`
	UserAvatar      string     `json:"user_avatar"`
	TripType        string     `json:"trip_type"`
	Role            string     `json:"role"`
	Title           string     `json:"title"`
	Content         string     `json:"content"`
	CoverImage      string     `json:"cover_image"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	AuditStatus     int        `json:"audit_status"`
	AuditStatusText string     `json:"audit_status_text"`
	AuditReason     string     `json:"audit_reason"`
	PublishedAt     *time.Time `json:"published_at"`

	DepartureTime   *time.Time `json:"departure_time"`
	PickupLocation  string     `json:"pickup_location"`
	PickupLat       float64    `json:"pickup_lat"`
	PickupLng       float64    `json:"pickup_lng"`
	DropoffLocation string     `json:"dropoff_location"`
	DropoffLat      float64    `json:"dropoff_lat"`
	DropoffLng      float64    `json:"dropoff_lng"`
	DistanceKm      float64    `json:"distance_km"`
	DurationMin     int        `json:"duration_min"`

	TotalSeats     int `json:"total_seats"`
	AvailableSeats int `json:"available_seats"`
	BookedSeats    int `json:"booked_seats"`

	PricePerSeat float64 `json:"price_per_seat"`
	TotalAmount  float64 `json:"total_amount"`
	TollFee      float64 `json:"toll_fee"`

	VehicleID          *uint `json:"vehicle_id"`
	DriverID           *uint `json:"driver_id"`
	RouteID            *uint `json:"route_id"`
	InsuranceID        *uint `json:"insurance_id"`
	TripID             *uint `json:"trip_id"`
	EmergencyContactID *uint `json:"emergency_contact_id"`

	ShareToken    string `json:"share_token"`
	PaymentMethod string `json:"payment_method"`

	Features      interface{} `json:"features"`
	Tags          interface{} `json:"tags"`

	ViewCount    int `json:"view_count"`
	FavCount     int `json:"fav_count"`
	ContactCount int `json:"contact_count"`
	ShareCount   int `json:"share_count"`

	Featured       bool `json:"featured"`
	Picked         bool `json:"picked"`
	Verified       bool `json:"verified"`
	PromotionLevel int  `json:"promotion_level"`

	ContentHash string `json:"content_hash"`
	RiskScore   int    `json:"risk_score"`

	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CancelledAt *time.Time `json:"cancelled_at"`

	Distance float64 `json:"distance,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// CreatePincheRequest 创建拼车行程请求
type CreatePincheRequest struct {
	TripType        string     `json:"trip_type" binding:"omitempty,oneof=shunfeng pinche baoche"`
	Role            string     `json:"role" binding:"omitempty,oneof=driver passenger"`
	Title           string     `json:"title" binding:"max=200"`
	Content         string     `json:"content" binding:"max=5000"`
	CoverImage      string     `json:"cover_image"`
	DepartureTime   *time.Time `json:"departure_time" binding:"required"`
	PickupLocation  string     `json:"pickup_location" binding:"required,min=2,max=255"`
	PickupLat       float64    `json:"pickup_lat"`
	PickupLng       float64    `json:"pickup_lng"`
	DropoffLocation string     `json:"dropoff_location" binding:"required,min=2,max=255"`
	DropoffLat      float64    `json:"dropoff_lat"`
	DropoffLng      float64    `json:"dropoff_lng"`
	DistanceKm      float64    `json:"distance_km"`
	DurationMin     int        `json:"duration_min"`
	TotalSeats      int        `json:"total_seats" binding:"omitempty,min=1,max=10"`
	PricePerSeat    float64    `json:"price_per_seat" binding:"min=0"`
	TollFee         float64    `json:"toll_fee"`
	VehicleID       *uint      `json:"vehicle_id"`
	RouteID         *uint      `json:"route_id"`
	PaymentMethod   string     `json:"payment_method" binding:"omitempty,oneof=cash wechat alipay balance etc"`
	Features        interface{} `json:"features"`
	Tags            interface{} `json:"tags"`
}

// UpdatePincheRequest 更新拼车行程请求
type UpdatePincheRequest struct {
	Title           *string     `json:"title"`
	Content         *string     `json:"content"`
	CoverImage      *string     `json:"cover_image"`
	DepartureTime   *time.Time  `json:"departure_time"`
	PickupLocation  *string     `json:"pickup_location"`
	PickupLat       *float64    `json:"pickup_lat"`
	PickupLng       *float64    `json:"pickup_lng"`
	DropoffLocation *string     `json:"dropoff_location"`
	DropoffLat      *float64    `json:"dropoff_lat"`
	DropoffLng      *float64    `json:"dropoff_lng"`
	DistanceKm      *float64    `json:"distance_km"`
	DurationMin     *int        `json:"duration_min"`
	TotalSeats      *int        `json:"total_seats"`
	PricePerSeat    *float64    `json:"price_per_seat"`
	TollFee         *float64    `json:"toll_fee"`
	VehicleID       *uint       `json:"vehicle_id"`
	RouteID         *uint       `json:"route_id"`
	PaymentMethod   *string     `json:"payment_method"`
	Features        interface{} `json:"features"`
	Tags            interface{} `json:"tags"`
}

// PincheListRequest 列表查询请求
type PincheListRequest struct {
	TripType      string  `form:"trip_type" json:"trip_type"`
	Role          string  `form:"role" json:"role"`
	Status        *int    `form:"status" json:"status"`
	MinPrice      float64 `form:"min_price" json:"min_price"`
	MaxPrice      float64 `form:"max_price" json:"max_price"`
	MinSeats      int     `form:"min_seats" json:"min_seats"`
	DepartureFrom string  `form:"departure_from" json:"departure_from"`
	DepartureTo   string  `form:"departure_to" json:"departure_to"`
	PickupCity    string  `form:"pickup_city" json:"pickup_city"`
	DropoffCity   string  `form:"dropoff_city" json:"dropoff_city"`
	Keyword       string  `form:"keyword" json:"keyword"`
	Sort          string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/departure_asc/distance/popular
	utils.Pagination
}

// PincheNearbyRequest 附近查询请求
type PincheNearbyRequest struct {
	Latitude  float64 `form:"latitude" json:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" json:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km" json:"radius_km"`
	TripType  string  `form:"trip_type" json:"trip_type"`
	Role      string  `form:"role" json:"role"`
	utils.Pagination
}

// PincheSearchRequest 搜索请求
type PincheSearchRequest struct {
	Keyword        string  `form:"keyword" json:"keyword"`
	PickupLocation string  `form:"pickup_location" json:"pickup_location"`
	DropoffLocation string `form:"dropoff_location" json:"dropoff_location"`
	DepartureDate  string  `form:"departure_date" json:"departure_date"`
	TripType       string  `form:"trip_type" json:"trip_type"`
	MinPrice       float64 `form:"min_price" json:"min_price"`
	MaxPrice       float64 `form:"max_price" json:"max_price"`
	MinSeats       int     `form:"min_seats" json:"min_seats"`
	utils.Pagination
}

// PincheMatchRequest 智能匹配请求
type PincheMatchRequest struct {
	PickupLocation  string  `json:"pickup_location" binding:"required"`
	PickupLat       float64 `json:"pickup_lat"`
	PickupLng       float64 `json:"pickup_lng"`
	DropoffLocation string  `json:"dropoff_location" binding:"required"`
	DropoffLat      float64 `json:"dropoff_lat"`
	DropoffLng      float64 `json:"dropoff_lng"`
	DepartureTime   *time.Time `json:"departure_time"`
	Seats           int     `json:"seats" binding:"omitempty,min=1,max=10"`
	MaxPrice        float64 `json:"max_price"`
	MaxRadiusKm     float64 `json:"max_radius_km"`
	utils.Pagination
}

// PincheMatchItem 匹配项
type PincheMatchItem struct {
	PincheInfo
	MatchScore float64 `json:"match_score"` // 匹配度 0-100
	MatchReasons []string `json:"match_reasons"`
}

// PincheMatchResponse 匹配响应
type PincheMatchResponse struct {
	Total int               `json:"total"`
	List  []PincheMatchItem `json:"list"`
}

// PincheViewRequest 浏览记录请求
type PincheViewRequest struct {
	PincheID uint `json:"pinche_id" binding:"required"`
	Source   string `json:"source"` // list/detail/search/recommend
}

// PincheAdminListRequest 管理后台列表请求
type PincheAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	TripType    string `form:"trip_type" json:"trip_type"`
	Role        string `form:"role" json:"role"`
	Status      *int   `form:"status" json:"status"`
	AuditStatus *int   `form:"audit_status" json:"audit_status"`
	Keyword     string `form:"keyword" json:"keyword"`
	StartTime   string `form:"start_time" json:"start_time"`
	EndTime     string `form:"end_time" json:"end_time"`
	utils.Pagination
}
