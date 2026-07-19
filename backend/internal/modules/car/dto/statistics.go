// Package dto 同城车辆买卖数据传输对象 - 统计
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// OverviewResponse 平台总览统计响应
type OverviewResponse struct {
	TotalCars       int64   `json:"total_cars"`
	TotalPublished  int64   `json:"total_published"`
	TotalSold       int64   `json:"total_sold"`
	TotalUsers      int64   `json:"total_users"`
	TotalDealers    int64   `json:"total_dealers"`
	TotalDealings   int64   `json:"total_dealings"`
	TotalAmount     float64 `json:"total_amount"`
	TotalInspections int64  `json:"total_inspections"`
	TotalTestDrives int64   `json:"total_test_drives"`
	TotalTransfers  int64   `json:"total_transfers"`
	TodayNew        int64   `json:"today_new"`
	TodayPublished  int64   `json:"today_published"`
	TodaySold       int64   `json:"today_sold"`
	PendingAudit    int64   `json:"pending_audit"`
	PendingReports  int64   `json:"pending_reports"`
}

// SellerOverviewResponse 卖家统计响应
type SellerOverviewResponse struct {
	UserID         uint    `json:"user_id"`
	TotalCars      int64   `json:"total_cars"`
	PublishedCars  int64   `json:"published_cars"`
	SoldCars       int64   `json:"sold_cars"`
	OfflineCars    int64   `json:"offline_cars"`
	TotalViews     int64   `json:"total_views"`
	TotalFavs      int64   `json:"total_favs"`
	TotalContacts  int64   `json:"total_contacts"`
	TotalTestDrives int64  `json:"total_test_drives"`
	TotalDeals     int64   `json:"total_deals"`
	TotalAmount    float64 `json:"total_amount"`
	AvgDealDays    int     `json:"avg_deal_days"`
	ConversionRate float64 `json:"conversion_rate"`
}

// HotItemResponse 热门车源响应
type HotItemResponse struct {
	CarID      uint    `json:"car_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Price      float64 `json:"price"`
	BrandName  string  `json:"brand_name"`
	ModelName  string  `json:"model_name"`
	Year       int     `json:"year"`
	ViewCount  int     `json:"view_count"`
	FavCount   int     `json:"fav_count"`
	ContactCount int   `json:"contact_count"`
}

// PriceTrendResponse 价格趋势响应
type PriceTrendResponse struct {
	Date      string  `json:"date"`
	AvgPrice  float64 `json:"avg_price"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	DealCount int     `json:"deal_count"`
}

// StatisticInfo 统计记录响应
type StatisticInfo struct {
	ID              uint      `json:"id"`
	StatDate        time.Time `json:"stat_date"`
	StatType        string    `json:"stat_type"`
	TargetID        uint      `json:"target_id"`
	TargetName      string    `json:"target_name"`
	ImpressionCount int       `json:"impression_count"`
	ClickCount      int       `json:"click_count"`
	FavCount        int       `json:"fav_count"`
	ContactCount    int       `json:"contact_count"`
	TestDriveCount  int       `json:"test_drive_count"`
	DealCount       int       `json:"deal_count"`
	ConversionRate  float64   `json:"conversion_rate"`
	AvgPrice        float64   `json:"avg_price"`
	AvgDealDays     int       `json:"avg_deal_days"`
	RegionID        uint      `json:"region_id"`
}

// StatisticListRequest 统计列表请求
type StatisticListRequest struct {
	StatType string `form:"stat_type" json:"stat_type"`
	TargetID uint   `form:"target_id" json:"target_id"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	utils.Pagination
}

// CarViewRequest 浏览记录请求
type CarViewRequest struct {
	CarID     uint   `json:"car_id" binding:"required"`
	ListingID uint   `json:"listing_id"`
	Device    string `json:"device" binding:"omitempty,oneof=pc wap app miniapp"`
	Source    string `json:"source" binding:"omitempty,oneof=search category recommend direct share"`
	Duration  int    `json:"duration"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// RecommendationInfo 推荐记录响应
type RecommendationInfo struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	CarID          uint       `json:"car_id"`
	RecType        string     `json:"rec_type"`
	Source         string     `json:"source"`
	Score          float64    `json:"score"`
	Reason         string     `json:"reason"`
	PriceMatch     float64    `json:"price_match"`
	BrandMatch     float64    `json:"brand_match"`
	TypeMatch      float64    `json:"type_match"`
	ConditionMatch float64    `json:"condition_match"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	ClickedAt      *time.Time `json:"clicked_at"`
	ContactedAt    *time.Time `json:"contacted_at"`
	ViewedAt       *time.Time `json:"viewed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RecommendationListRequest 推荐列表请求
type RecommendationListRequest struct {
	UserID  uint   `form:"user_id" json:"user_id"`
	CarID   uint   `form:"car_id" json:"car_id"`
	RecType string `form:"rec_type" json:"rec_type"`
	Source  string `form:"source" json:"source"`
	Status  *int   `form:"status" json:"status"`
	utils.Pagination
}

// EscrowInfo 担保交易响应
type EscrowInfo struct {
	ID                uint       `json:"id"`
	EscrowNo          string     `json:"escrow_no"`
	EscrowType        string     `json:"escrow_type"`
	CarID             uint       `json:"car_id"`
	ListingID         uint       `json:"listing_id"`
	ContractID        uint       `json:"contract_id"`
	PayerID           uint       `json:"payer_id"`
	PayeeID           uint       `json:"payee_id"`
	DealerID          uint       `json:"dealer_id"`
	Amount            float64    `json:"amount"`
	PlatformFee       float64    `json:"platform_fee"`
	DealerFee         float64    `json:"dealer_fee"`
	PayeeAmount       float64    `json:"payee_amount"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	PayMethod         string     `json:"pay_method"`
	PayTradeNo        string     `json:"pay_trade_no"`
	PaidAt            *time.Time `json:"paid_at"`
	FrozenAt          *time.Time `json:"frozen_at"`
	ReleaseAt         *time.Time `json:"release_at"`
	RefundedAt        *time.Time `json:"refunded_at"`
	AutoReleaseAt     *time.Time `json:"auto_release_at"`
	DisputeReason     string     `json:"dispute_reason"`
	ArbitrationResult string     `json:"arbitration_result"`
	CompletedAt       *time.Time `json:"completed_at"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// EscrowListRequest 担保交易列表请求
type EscrowListRequest struct {
	CarID      uint   `form:"car_id" json:"car_id"`
	PayerID    uint   `form:"payer_id" json:"payer_id"`
	PayeeID    uint   `form:"payee_id" json:"payee_id"`
	DealerID   uint   `form:"dealer_id" json:"dealer_id"`
	EscrowType string `form:"escrow_type" json:"escrow_type"`
	Status     *int   `form:"status" json:"status"`
	utils.Pagination
}

// EscrowActionRequest 担保操作请求
type EscrowActionRequest struct {
	Action         string `json:"action" binding:"required,oneof=release refund dispute arbitrate"`
	DisputeReason  string `json:"dispute_reason" binding:"max=500"`
	ArbitrationResult string `json:"arbitration_result"`
}

// ContractInfo 合同响应
type ContractInfo struct {
	ID              uint       `json:"id"`
	ContractNo      string     `json:"contract_no"`
	ContractType    string     `json:"contract_type"`
	CarID           uint       `json:"car_id"`
	ListingID       uint       `json:"listing_id"`
	SellerID        uint       `json:"seller_id"`
	SellerName      string     `json:"seller_name"`
	SellerPhone     string     `json:"seller_phone"`
	SellerIDCard    string     `json:"seller_id_card"`
	BuyerID         uint       `json:"buyer_id"`
	BuyerName       string     `json:"buyer_name"`
	BuyerPhone      string     `json:"buyer_phone"`
	BuyerIDCard     string     `json:"buyer_id_card"`
	AgentID         uint       `json:"agent_id"`
	AgentName       string     `json:"agent_name"`
	DealPrice       float64    `json:"deal_price"`
	Deposit         float64    `json:"deposit"`
	PaymentMethod   string     `json:"payment_method"`
	LoanAmount      float64    `json:"loan_amount"`
	LoanPeriods     int        `json:"loan_periods"`
	TransferFee     float64    `json:"transfer_fee"`
	ServiceFee      float64    `json:"service_fee"`
	ContractURL     string     `json:"contract_url"`
	Attachments     interface{} `json:"attachments"`
	SignedAt        *time.Time `json:"signed_at"`
	EffectiveAt     *time.Time `json:"effective_at"`
	ExpiredAt       *time.Time `json:"expired_at"`
	TerminatedAt    *time.Time `json:"terminated_at"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	Remark          string     `json:"remark"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ContractListRequest 合同列表请求
type ContractListRequest struct {
	CarID        uint   `form:"car_id" json:"car_id"`
	SellerID     uint   `form:"seller_id" json:"seller_id"`
	BuyerID      uint   `form:"buyer_id" json:"buyer_id"`
	ContractType string `form:"contract_type" json:"contract_type"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}
