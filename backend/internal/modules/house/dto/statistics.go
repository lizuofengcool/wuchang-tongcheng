// Package dto 数据统计 + 房贷计算 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 统计 =====

// StatResponse 数据统计响应
type StatResponse struct {
	ID              uint      `json:"id"`
	RegionID        uint      `json:"region_id"`
	StatDate        time.Time `json:"stat_date"`
	StatType        string    `json:"stat_type"`
	TargetID        uint      `json:"target_id"`
	TargetName      string    `json:"target_name"`
	ImpressionCount int       `json:"impression_count"`
	ClickCount      int       `json:"click_count"`
	FavCount        int       `json:"fav_count"`
	ContactCount    int       `json:"contact_count"`
	ViewingCount    int       `json:"viewing_count"`
	DealCount       int       `json:"deal_count"`
	ConversionRate  float64   `json:"conversion_rate"`
	AvgSalePrice    float64   `json:"avg_sale_price"`
	AvgRentPrice    float64   `json:"avg_rent_price"`
	AvgDealDays     int       `json:"avg_deal_days"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// StatListQuery 统计列表查询
type StatListQuery struct {
	RegionID  uint      `form:"region_id" json:"region_id"`
	StatType  string    `form:"stat_type" json:"stat_type"`
	TargetID  uint      `form:"target_id" json:"target_id"`
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
	utils.Pagination
}

// OverviewResponse 平台总览统计
type OverviewResponse struct {
	TotalHouses      int64   `json:"total_houses"`
	TotalListings    int64   `json:"total_listings"`
	TotalCommunities int64   `json:"total_communities"`
	TotalAgents      int64   `json:"total_agents"`
	TotalDeals       int64   `json:"total_deals"`
	TotalViewings    int64   `json:"total_viewings"`
	TotalReports     int64   `json:"total_reports"`
	PendingReports   int64   `json:"pending_reports"`
	PendingAudits    int64   `json:"pending_audits"`
	TodayNewHouses   int64   `json:"today_new_houses"`
	TodayNewListings int64   `json:"today_new_listings"`
	TodayDeals       int64   `json:"today_deals"`
	TodayViewings    int64   `json:"today_viewings"`
}

// HotHouseResponse 热门房源
type HotHouseResponse struct {
	HouseID    uint    `json:"house_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Price      float64 `json:"price"`
	Layout     string  `json:"layout"`
	BuildingArea float64 `json:"building_area"`
	CommunityName string `json:"community_name"`
	ViewCount  int     `json:"view_count"`
	FavCount   int     `json:"fav_count"`
	ContactCount int   `json:"contact_count"`
}

// PriceTrendResponse 价格趋势
type PriceTrendResponse struct {
	StatDate     time.Time `json:"stat_date"`
	AvgSalePrice float64   `json:"avg_sale_price"`
	AvgRentPrice float64   `json:"avg_rent_price"`
	DealCount    int       `json:"deal_count"`
}

// SellerOverviewResponse 卖家/房东总览
type SellerOverviewResponse struct {
	UserID         uint    `json:"user_id"`
	TotalListings  int64   `json:"total_listings"`
	ActiveListings int64   `json:"active_listings"`
	TotalViews     int64   `json:"total_views"`
	TotalFavs      int64   `json:"total_favs"`
	TotalContacts  int64   `json:"total_contacts"`
	TotalViewings  int64   `json:"total_viewings"`
	TotalDeals     int64   `json:"total_deals"`
	TotalAmount    float64 `json:"total_amount"`
	AvgDealDays    int     `json:"avg_deal_days"`
	ResponseRate   float64 `json:"response_rate"`
}

// ===== 房贷计算 =====

// MortgageResponse 房贷方案响应
type MortgageResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	LoanType       string    `json:"loan_type"`
	LoanTypeText   string    `json:"loan_type_text"`
	MinDownPayment float64   `json:"min_down_payment"`
	MaxDownPayment float64   `json:"max_down_payment"`
	InterestRate   float64   `json:"interest_rate"`
	MaxPeriods     int       `json:"max_periods"`
	MaxAmount      float64   `json:"max_amount"`
	Description    string    `json:"description"`
	Sort           int       `json:"sort"`
	Status         int       `json:"status"`
	IsHot          bool      `json:"is_hot"`
	UseCount       int       `json:"use_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MortgageCreateRequest 创建/更新房贷方案请求
type MortgageCreateRequest struct {
	Name           string  `json:"name" binding:"required,max=64"`
	Code           string  `json:"code" binding:"required,max=64"`
	LoanType       string  `json:"loan_type" binding:"omitempty,oneof=commercial provident_fund combined full"`
	MinDownPayment float64 `json:"min_down_payment" binding:"gte=0,lte=100"`
	MaxDownPayment float64 `json:"max_down_payment" binding:"gte=0,lte=100"`
	InterestRate   float64 `json:"interest_rate" binding:"gte=0"`
	MaxPeriods     int     `json:"max_periods" binding:"gte=0"`
	MaxAmount      float64 `json:"max_amount" binding:"gte=0"`
	Description    string  `json:"description" binding:"max=500"`
	Sort           int     `json:"sort"`
	Status         int     `json:"status" binding:"omitempty,oneof=0 1"`
	IsHot          bool    `json:"is_hot"`
}

// MortgageCalculateRequest 房贷计算请求
type MortgageCalculateRequest struct {
	LoanType     string  `json:"loan_type" binding:"required,oneof=commercial provident_fund combined full"`
	TotalPrice   float64 `json:"total_price" binding:"required,gte=0"`
	DownPayment  float64 `json:"down_payment" binding:"gte=0"`  // 首付比例（%）
	Periods      int     `json:"periods" binding:"required,gte=1,lte=360"` // 期数（月）
	InterestRate float64 `json:"interest_rate" binding:"gte=0"` // 年利率（小数 0.049=4.9%）
}

// MortgageCalculateResponse 房贷计算结果
type MortgageCalculateResponse struct {
	LoanType       string  `json:"loan_type"`
	TotalPrice     float64 `json:"total_price"`
	DownPayment    float64 `json:"down_payment"`        // 首付金额
	DownPaymentPct float64 `json:"down_payment_pct"`    // 首付比例
	LoanAmount     float64 `json:"loan_amount"`         // 贷款金额
	Periods        int     `json:"periods"`             // 期数
	InterestRate   float64 `json:"interest_rate"`       // 年利率
	MonthlyPayment float64 `json:"monthly_payment"`     // 月供（等额本息）
	TotalPayment   float64 `json:"total_payment"`       // 还款总额
	TotalInterest  float64 `json:"total_interest"`      // 总利息
	MonthlyPrincipal float64 `json:"monthly_principal"` // 月供本金（等额本金首月）
	MonthlyInterest  float64 `json:"monthly_interest"`  // 月供利息（等额本息首月）
}

// MortgageListQuery 房贷方案列表查询
type MortgageListQuery struct {
	LoanType string `form:"loan_type" json:"loan_type"`
	Status   *int   `form:"status" json:"status"`
	IsHot    *bool  `form:"is_hot" json:"is_hot"`
	utils.Pagination
}

// ===== 担保交易 + 成交 + 推荐 =====

// EscrowResponse 担保交易响应
type EscrowResponse struct {
	ID                uint       `json:"id"`
	EscrowNo          string     `json:"escrow_no"`
	EscrowType        string     `json:"escrow_type"`
	HouseID           uint       `json:"house_id"`
	ListingID         uint       `json:"listing_id"`
	ContractID        uint       `json:"contract_id"`
	CommunityID       uint       `json:"community_id"`
	PayerID           uint       `json:"payer_id"`
	PayeeID           uint       `json:"payee_id"`
	AgentID           uint       `json:"agent_id"`
	Amount            float64    `json:"amount"`
	PlatformFee       float64    `json:"platform_fee"`
	AgentFee          float64    `json:"agent_fee"`
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

// EscrowCreateRequest 创建担保交易请求
type EscrowCreateRequest struct {
	EscrowType  string  `json:"escrow_type" binding:"required,oneof=deposit commission full_payment rent down_payment"`
	HouseID     uint    `json:"house_id" binding:"required"`
	ListingID   uint    `json:"listing_id"`
	ContractID  uint    `json:"contract_id"`
	CommunityID uint    `json:"community_id"`
	PayeeID     uint    `json:"payee_id" binding:"required"`
	AgentID     uint    `json:"agent_id"`
	Amount      float64 `json:"amount" binding:"required,gte=0"`
	PayMethod   string  `json:"pay_method" binding:"omitempty,oneof=wechat alipay bank balance"`
}

// EscrowListQuery 担保交易列表查询
type EscrowListQuery struct {
	HouseID    uint   `form:"house_id" json:"house_id"`
	PayerID    uint   `form:"payer_id" json:"payer_id"`
	PayeeID    uint   `form:"payee_id" json:"payee_id"`
	AgentID    uint   `form:"agent_id" json:"agent_id"`
	EscrowType string `form:"escrow_type" json:"escrow_type"`
	Status     *int   `form:"status" json:"status"`
	utils.Pagination
}

// EscrowDisputeRequest 担保争议请求
type EscrowDisputeRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// EscrowArbitrateRequest 担保仲裁请求
type EscrowArbitrateRequest struct {
	Result       string `json:"result" binding:"required,max=500"`
	ToPayer      bool   `json:"to_payer"` // true 退款给付款方，false 放款给收款方
}

// DealResponse 成交记录响应
type DealResponse struct {
	ID             uint       `json:"id"`
	DealNo         string     `json:"deal_no"`
	HouseID        uint       `json:"house_id"`
	ListingID      uint       `json:"listing_id"`
	CommunityID    uint       `json:"community_id"`
	ContractID     uint       `json:"contract_id"`
	EscrowID       uint       `json:"escrow_id"`
	DealType       string     `json:"deal_type"`
	SellerID       uint       `json:"seller_id"`
	SellerName     string     `json:"seller_name"`
	BuyerID        uint       `json:"buyer_id"`
	BuyerName      string     `json:"buyer_name"`
	AgentID        uint       `json:"agent_id"`
	AgentName      string     `json:"agent_name"`
	DealPrice      float64    `json:"deal_price"`
	AveragePrice   float64    `json:"average_price"`
	Commission     float64    `json:"commission"`
	DealDate       *time.Time `json:"deal_date"`
	ListedAt       *time.Time `json:"listed_at"`
	DealDays       int        `json:"deal_days"`
	PaymentMethod  string     `json:"payment_method"`
	LoanAmount     float64    `json:"loan_amount"`
	LoanPeriods    int        `json:"loan_periods"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	CompletedAt    *time.Time `json:"completed_at"`
	CanceledAt     *time.Time `json:"canceled_at"`
	CanceledReason string     `json:"canceled_reason"`
	Remark         string     `json:"remark"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DealCreateRequest 创建成交记录请求
type DealCreateRequest struct {
	HouseID       uint      `json:"house_id" binding:"required"`
	ListingID     uint      `json:"listing_id"`
	CommunityID   uint      `json:"community_id"`
	ContractID    uint      `json:"contract_id"`
	EscrowID      uint      `json:"escrow_id"`
	DealType      string    `json:"deal_type" binding:"required,oneof=sale rent transfer"`
	SellerID      uint      `json:"seller_id"`
	SellerName    string    `json:"seller_name" binding:"max=50"`
	BuyerID       uint      `json:"buyer_id"`
	BuyerName     string    `json:"buyer_name" binding:"max=50"`
	AgentID       uint      `json:"agent_id"`
	AgentName     string    `json:"agent_name" binding:"max=50"`
	DealPrice     float64   `json:"deal_price" binding:"gte=0"`
	DealDate      *time.Time `json:"deal_date"`
	PaymentMethod string    `json:"payment_method" binding:"max=32"`
	LoanAmount    float64   `json:"loan_amount" binding:"gte=0"`
	LoanPeriods   int       `json:"loan_periods" binding:"gte=0"`
	Remark        string    `json:"remark" binding:"max=500"`
}

// DealListQuery 成交记录列表查询
type DealListQuery struct {
	HouseID    uint   `form:"house_id" json:"house_id"`
	CommunityID uint  `form:"community_id" json:"community_id"`
	SellerID   uint   `form:"seller_id" json:"seller_id"`
	BuyerID    uint   `form:"buyer_id" json:"buyer_id"`
	AgentID    uint   `form:"agent_id" json:"agent_id"`
	DealType   string `form:"deal_type" json:"deal_type"`
	Status     *int   `form:"status" json:"status"`
	utils.Pagination
}

// DealConfirmRequest 成交确认请求
type DealConfirmRequest struct {
	Remark string `json:"remark" binding:"max=500"`
}

// DealCancelRequest 成交取消请求
type DealCancelRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	HouseID         uint       `json:"house_id"`
	RecType         string     `json:"rec_type"`
	Source          string     `json:"source"`
	Score           float64    `json:"score"`
	Reason          string     `json:"reason"`
	PriceMatch      float64    `json:"price_match"`
	LocationMatch   float64    `json:"location_match"`
	LayoutMatch     float64    `json:"layout_match"`
	FacilityMatch   float64    `json:"facility_match"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	ClickedAt       *time.Time `json:"clicked_at"`
	ContactedAt     *time.Time `json:"contacted_at"`
	ViewedAt        *time.Time `json:"viewed_at"`
	DismissedAt     *time.Time `json:"dismissed_at"`
	ExpiredAt       *time.Time `json:"expired_at"`
	CreatedAt       time.Time  `json:"created_at"`
	// 关联房源冗余
	HouseTitle      string  `json:"house_title,omitempty"`
	HouseCoverImage string  `json:"house_cover_image,omitempty"`
	HousePrice      float64 `json:"house_price,omitempty"`
	HouseLayout     string  `json:"house_layout,omitempty"`
	CommunityName   string  `json:"community_name,omitempty"`
}

// RecommendationListQuery 推荐列表查询
type RecommendationListQuery struct {
	UserID  uint   `form:"user_id" json:"user_id"`
	RecType string `form:"rec_type" json:"rec_type"`
	Source  string `form:"source" json:"source"`
	Status  *int   `form:"status" json:"status"`
	utils.Pagination
}

// VRTourResponse VR 看房响应
type VRTourResponse struct {
	ID              uint       `json:"id"`
	HouseID         uint       `json:"house_id"`
	ListingID       uint       `json:"listing_id"`
	CommunityID     uint       `json:"community_id"`
	VRNo            string     `json:"vr_no"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	VRType          string     `json:"vr_type"`
	VRTypeText      string     `json:"vr_type_text"`
	VRURL           string     `json:"vr_url"`
	CoverImage      string     `json:"cover_image"`
	Scenes          []model.VRScene `json:"scenes"`
	DurationSeconds int        `json:"duration_seconds"`
	ViewCount       int        `json:"view_count"`
	ShareCount      int        `json:"share_count"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	RecorderID      uint       `json:"recorder_id"`
	RecorderName    string     `json:"recorder_name"`
	RecordedAt      *time.Time `json:"recorded_at"`
	PublishedAt     *time.Time `json:"published_at"`
	OfflineAt       *time.Time `json:"offline_at"`
	Equipment       string     `json:"equipment"`
	Resolution      string     `json:"resolution"`
	FileSize        int64      `json:"file_size"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// VRTourCreateRequest 创建 VR 看房请求
type VRTourCreateRequest struct {
	HouseID         uint              `json:"house_id" binding:"required"`
	ListingID       uint              `json:"listing_id"`
	CommunityID     uint              `json:"community_id"`
	Title           string            `json:"title" binding:"required,max=200"`
	Description     string            `json:"description"`
	VRType          string            `json:"vr_type" binding:"omitempty,oneof=panorama vr video 3d"`
	VRURL           string            `json:"vr_url" binding:"required,max=500"`
	CoverImage      string            `json:"cover_image" binding:"max=500"`
	Scenes          []model.VRScene   `json:"scenes"`
	DurationSeconds int               `json:"duration_seconds" binding:"gte=0"`
	RecorderID      uint              `json:"recorder_id"`
	RecorderName    string            `json:"recorder_name" binding:"max=50"`
	Equipment       string            `json:"equipment" binding:"max=64"`
	Resolution      string            `json:"resolution" binding:"max=32"`
	FileSize        int64             `json:"file_size" binding:"gte=0"`
	Status          int               `json:"status" binding:"omitempty,oneof=0 1"`
}

// VRTourListQuery VR 看房列表查询
type VRTourListQuery struct {
	HouseID     uint   `form:"house_id" json:"house_id"`
	ListingID   uint   `form:"listing_id" json:"listing_id"`
	CommunityID uint   `form:"community_id" json:"community_id"`
	VRType      string `form:"vr_type" json:"vr_type"`
	Status      *int   `form:"status" json:"status"`
	utils.Pagination
}
