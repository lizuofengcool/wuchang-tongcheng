// Package dto 同城二手物品数据传输对象
// 依据需求文档 2.2.A.10：商品发布/分类/搜索/留言/交易
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ErshouInfo 二手物品详情
type ErshouInfo struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	CoverImage  string     `json:"cover_image"`
	Images      []string   `json:"images"`           // 图片URL列表（由 service 拼装）
	Summary     string     `json:"summary"`

	// 发布者
	UserID     uint   `json:"user_id"`
	UserName   string `json:"user_name"`
	UserPhone  string `json:"user_phone"`
	UserAvatar string `json:"user_avatar"`

	// 分类
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`

	// 二手核心字段
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	PriceUnit     string  `json:"price_unit"`
	Condition     string  `json:"condition"`
	Brand         string  `json:"brand"`

	// 联系方式
	ContactPhone  string `json:"contact_phone"`
	ContactWechat string `json:"contact_wechat"`

	// 位置
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance,omitempty"` // 仅附近查询返回（公里）

	// 交易方式
	DeliveryMethod string `json:"delivery_method"`

	// 展示控制
	IsUrgent     bool       `json:"is_urgent"`
	ExpiryTime   *time.Time `json:"expiry_time"`
	ViewCount    int        `json:"view_count"`
	FavCount     int        `json:"fav_count"`
	MessageCount int        `json:"message_count"`

	// 状态
	Status      int        `json:"status"`
	AuditStatus int        `json:"audit_status"`
	AuditReason string     `json:"audit_reason"`
	PublishedAt *time.Time `json:"published_at"`
	RegionID    uint       `json:"region_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 当前用户是否已收藏（仅登录态列表/详情返回）
	HasFaved bool `json:"has_faved,omitempty"`
}

// CreateErshouRequest C端发布二手物品请求
type CreateErshouRequest struct {
	Title       string   `json:"title" binding:"required,max=200"`
	Content     string   `json:"content"`
	CoverImage  string   `json:"cover_image" binding:"max=255"`
	Images      []string `json:"images"`              // 图片URL列表
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	Price       float64  `json:"price" binding:"gte=0"`
	OriginalPrice float64 `json:"original_price" binding:"gte=0"`
	PriceUnit   string   `json:"price_unit" binding:"omitempty,oneof=元 万元 面议 免费"`
	Condition   string   `json:"condition" binding:"omitempty,oneof=new almost_new used broken"`
	Brand       string   `json:"brand" binding:"max=100"`
	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	Address     string   `json:"address" binding:"max=255"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	DeliveryMethod string `json:"delivery_method" binding:"omitempty,oneof=face self express"`
	IsUrgent    bool     `json:"is_urgent"`
	ExpireDays  int      `json:"expire_days"`  // 过期天数（默认30天）
	Status      int      `json:"status" binding:"oneof=0 1"`  // 0草稿 1直接发布
}

// UpdateErshouRequest 更新二手物品请求
type UpdateErshouRequest struct {
	Title       string   `json:"title" binding:"max=200"`
	Content     string   `json:"content"`
	CoverImage  string   `json:"cover_image" binding:"max=255"`
	Images      []string `json:"images"`
	Summary     string   `json:"summary" binding:"max=500"`
	CategoryID  uint     `json:"category_id"`
	Price       float64  `json:"price" binding:"gte=0"`
	OriginalPrice float64 `json:"original_price" binding:"gte=0"`
	PriceUnit   string   `json:"price_unit" binding:"omitempty,oneof=元 万元 面议 免费"`
	Condition   string   `json:"condition" binding:"omitempty,oneof=new almost_new used broken"`
	Brand       string   `json:"brand" binding:"max=100"`
	ContactPhone  string `json:"contact_phone" binding:"max=20"`
	ContactWechat string `json:"contact_wechat" binding:"max=50"`
	Address     string   `json:"address" binding:"max=255"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	DeliveryMethod string `json:"delivery_method" binding:"omitempty,oneof=face self express"`
	IsUrgent    *bool    `json:"is_urgent"`
	ExpireDays  int      `json:"expire_days"`
	Status      int      `json:"status" binding:"omitempty,oneof=0 1 3"` // 0草稿 1发布 3下架
}

// ErshouListRequest 列表查询（C端）
type ErshouListRequest struct {
	CategoryID uint    `form:"category_id" json:"category_id"`
	Keyword    string  `form:"keyword" json:"keyword"`
	MinPrice   float64 `form:"min_price" json:"min_price"`
	MaxPrice   float64 `form:"max_price" json:"max_price"`
	Condition  string  `form:"condition" json:"condition"`
	Brand      string  `form:"brand" json:"brand"`
	IsUrgent   *bool   `form:"is_urgent" json:"is_urgent"`
	Sort       string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/popular
	utils.Pagination
}

// ErshouNearbyRequest 附近查询
type ErshouNearbyRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	RadiusKm  float64 `form:"radius_km"` // 默认 5 公里
	utils.Pagination
}

// ErshouSearchRequest 搜索（走 Elasticsearch）
type ErshouSearchRequest struct {
	Keyword string `form:"keyword" binding:"required,max=100"`
	utils.Pagination
}

// ErshouAdminListRequest 管理后台列表查询（M端）
// Status/AuditStatus 使用 *int 指针，nil 表示不传该过滤条件（返回全部），
// 区分"未传"（nil，不过滤）和"传0"（筛选草稿/待审）。
type ErshouAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	CategoryID  uint   `form:"category_id" json:"category_id"`
	Status      *int   `form:"status" json:"status"`         // nil全部 0草稿 1已发布 2已售 3下架 4过期
	AuditStatus *int   `form:"audit_status" json:"audit_status"` // nil全部 0待审 1通过 2拒绝
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// AuditRequest 审核操作请求（M端）
type AuditRequest struct {
	AuditStatus int    `json:"audit_status" binding:"oneof=0 1 2"` // 0待审 1通过 2拒绝
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// AdminUpdateStatusRequest 管理后台强制下架/恢复
type AdminUpdateStatusRequest struct {
	Status int `json:"status" binding:"oneof=1 3 4"` // 1发布 3下架 4过期
}

// CreateMessageRequest C端用户留言请求
type CreateMessageRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}

// MessageInfo 留言信息
type MessageInfo struct {
	ID         uint      `json:"id"`
	ErshouID   uint      `json:"ershou_id"`
	FromUserID uint      `json:"from_user_id"`
	FromName   string    `json:"from_name"`
	FromAvatar string    `json:"from_avatar"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// FavResponse 收藏操作响应
type FavResponse struct {
	HasFaved bool `json:"has_faved"`
	FavCount int  `json:"fav_count"`
}

// ====================================================================
// v3.2.1 扩展 DTO（SKU/订单/拍卖/推广/物流/担保/退款/举报/评价/店铺/标签/品牌/型号/分类属性/审核规则/信用/统计/批量/导出/搜索推荐）
// 依据需求文档 2.2.A.10 + v3.2.1 架构方案第二章
// ====================================================================

// ===== SKU =====

// SKURequest 创建/更新 SKU 请求
type SKURequest struct {
	SKUCode    string  `json:"sku_code" binding:"required,max=64"`
	Name       string  `json:"name" binding:"max=200"`
	Color      string  `json:"color" binding:"max=50"`
	Size       string  `json:"size" binding:"max=50"`
	Version    string  `json:"version" binding:"max=100"`
	Price      float64 `json:"price" binding:"gte=0"`
	Stock      int     `json:"stock" binding:"gte=0"`
	Image      string  `json:"image" binding:"max=255"`
	Weight     float64 `json:"weight" binding:"gte=0"`
	Barcode    string  `json:"barcode" binding:"max=64"`
	Status     int     `json:"status" binding:"oneof=0 1 2"`
	Attributes map[string]interface{} `json:"attributes"`
	Sort       int     `json:"sort"`
}

// SKUResponse SKU 详情响应
type SKUResponse struct {
	ID         uint                  `json:"id"`
	ErshouID   uint                  `json:"ershou_id"`
	SKUCode    string                `json:"sku_code"`
	Name       string                `json:"name"`
	Color      string                `json:"color"`
	Size       string                `json:"size"`
	Version    string                `json:"version"`
	Price      float64               `json:"price"`
	Stock      int                   `json:"stock"`
	SoldCount  int                   `json:"sold_count"`
	Image      string                `json:"image"`
	Weight     float64               `json:"weight"`
	Barcode    string                `json:"barcode"`
	Status     int                   `json:"status"`
	Attributes map[string]interface{} `json:"attributes"`
	Sort       int                   `json:"sort"`
	CreatedAt  time.Time             `json:"created_at"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

// SKUListResponse SKU 列表响应
type SKUListResponse struct {
	List     []SKUResponse `json:"list"`
	Total    int64         `json:"total"`
}

// ===== 订单 =====

// OrderItemRequest 订单子项请求
type OrderItemRequest struct {
	ErshouID uint    `json:"ershou_id" binding:"required"`
	SKUID    uint    `json:"sku_id"`
	Quantity int     `json:"quantity" binding:"required,gte=1"`
	Remark   string  `json:"remark" binding:"max=500"`
}

// OrderCreateRequest 创建订单请求
type OrderCreateRequest struct {
	Items            []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	ShopID           uint               `json:"shop_id"`
	PayMethod        string             `json:"pay_method" binding:"omitempty,oneof=wechat alipay balance installment"`
	DeliveryMethod   string             `json:"delivery_method" binding:"omitempty,oneof=face self express"`
	Remark           string             `json:"remark" binding:"max=500"`
	ContactName      string             `json:"contact_name" binding:"max=50"`
	ContactPhone     string             `json:"contact_phone" binding:"max=20"`
	ContactAddress   string             `json:"contact_address" binding:"max=500"`
	EscrowEnabled    bool               `json:"escrow_enabled"`
	InstallmentEnabled bool             `json:"installment_enabled"`
	InstallmentPeriods int              `json:"installment_periods"`
}

// OrderResponse 订单详情响应
type OrderResponse struct {
	ID                  uint       `json:"id"`
	OrderNo             string     `json:"order_no"`
	BuyerID             uint       `json:"buyer_id"`
	SellerID            uint       `json:"seller_id"`
	ShopID              uint       `json:"shop_id"`
	TotalAmount         float64    `json:"total_amount"`
	ItemAmount          float64    `json:"item_amount"`
	DeliveryFee         float64    `json:"delivery_fee"`
	DiscountAmount      float64    `json:"discount_amount"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	PayMethod           string     `json:"pay_method"`
	PayTradeNo          string     `json:"pay_trade_no"`
	DeliveryMethod      string     `json:"delivery_method"`
	Remark              string     `json:"remark"`
	ContactName         string     `json:"contact_name"`
	ContactPhone        string     `json:"contact_phone"`
	ContactAddress      string     `json:"contact_address"`
	EscrowEnabled       bool       `json:"escrow_enabled"`
	InstallmentEnabled  bool       `json:"installment_enabled"`
	InstallmentPeriods  int        `json:"installment_periods"`
	PaidAt              *time.Time `json:"paid_at"`
	ShippedAt           *time.Time `json:"shipped_at"`
	ReceivedAt          *time.Time `json:"received_at"`
	SettledAt           *time.Time `json:"settled_at"`
	ClosedAt            *time.Time `json:"closed_at"`
	AutoCloseAt         *time.Time `json:"auto_close_at"`
	AutoReceiveAt       *time.Time `json:"auto_receive_at"`
	RegionID            uint       `json:"region_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Items               []OrderItemResponse `json:"items"`
}

// OrderItemResponse 订单子项响应
type OrderItemResponse struct {
	ID         uint    `json:"id"`
	OrderID    uint    `json:"order_id"`
	ErshouID   uint    `json:"ershou_id"`
	SKUID      uint    `json:"sku_id"`
	SKUCode    string  `json:"sku_code"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	Subtotal   float64 `json:"subtotal"`
	Remark     string  `json:"remark"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List  []OrderResponse `json:"list"`
	Total int64           `json:"total"`
}

// OrderQuery 订单列表查询参数
type OrderQuery struct {
	Role     string `form:"role" json:"role"`         // buyer/seller/all
	Status   *int   `form:"status" json:"status"`
	OrderNo  string `form:"order_no" json:"order_no"`
	Keyword  string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// OrderStatusUpdateRequest 订单状态变更请求（状态机变更）
type OrderStatusUpdateRequest struct {
	Action  string `json:"action" binding:"required,oneof=pay ship receive cancel complete"`
	Remark  string `json:"remark" binding:"max=500"`
}

// ===== 拍卖 =====

// AuctionCreateRequest 创建拍卖请求
type AuctionCreateRequest struct {
	StartPrice        float64 `json:"start_price" binding:"required,gte=0"`
	StepPrice         float64 `json:"step_price" binding:"gte=0"`
	ReservePrice      float64 `json:"reserve_price" binding:"gte=0"`
	BondAmount        float64 `json:"bond_amount" binding:"gte=0"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time" binding:"required"`
	AutoExtendEnabled bool    `json:"auto_extend_enabled"`
}

// AuctionBidRequest 出价请求
type AuctionBidRequest struct {
	BidPrice float64 `json:"bid_price" binding:"required,gte=0"`
}

// AuctionResponse 拍卖详情响应
type AuctionResponse struct {
	ID                uint       `json:"id"`
	ErshouID          uint       `json:"ershou_id"`
	StartPrice        float64    `json:"start_price"`
	StepPrice         float64    `json:"step_price"`
	ReservePrice      float64    `json:"reserve_price"`
	BondAmount        float64    `json:"bond_amount"`
	CurrentBidPrice   float64    `json:"current_bid_price"`
	CurrentBidUserID  uint       `json:"current_bid_user_id"`
	BidCount          int        `json:"bid_count"`
	WatcherCount      int        `json:"watcher_count"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	AutoExtendEnabled bool       `json:"auto_extend_enabled"`
	WinnerID          uint       `json:"winner_id"`
	WinnerPrice       float64    `json:"winner_price"`
	ClosedAt          *time.Time `json:"closed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// AuctionListResponse 拍卖列表响应
type AuctionListResponse struct {
	List  []AuctionResponse `json:"list"`
	Total int64             `json:"total"`
}

// ===== 推广 =====

// PromotionCreateRequest 创建推广请求
type PromotionCreateRequest struct {
	PromotionType string  `json:"promotion_type" binding:"required,oneof=home_banner channel_top search_top featured urgent refresh"`
	DurationDays  int     `json:"duration_days" binding:"required,gte=1"`
	Amount        float64 `json:"amount" binding:"gte=0"`
	PayMethod     string  `json:"pay_method" binding:"omitempty,oneof=wechat alipay balance"`
}

// PromotionResponse 推广记录响应
type PromotionResponse struct {
	ID               uint       `json:"id"`
	ErshouID         uint       `json:"ershou_id"`
	UserID           uint       `json:"user_id"`
	PromotionType    string     `json:"promotion_type"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	StartTime        *time.Time `json:"start_time"`
	EndTime          *time.Time `json:"end_time"`
	DurationDays     int        `json:"duration_days"`
	Amount           float64    `json:"amount"`
	PayMethod        string     `json:"pay_method"`
	PayTradeNo       string     `json:"pay_trade_no"`
	PaidAt           *time.Time `json:"paid_at"`
	ImpressionCount  int        `json:"impression_count"`
	ClickCount       int        `json:"click_count"`
	FavCount         int        `json:"fav_count"`
	ConsultCount     int        `json:"consult_count"`
	OrderCount       int        `json:"order_count"`
	ROI              float64    `json:"roi"`
	CreatedAt        time.Time  `json:"created_at"`
}

// PromotionStatsResponse 推广效果统计
type PromotionStatsResponse struct {
	TotalPromotions  int     `json:"total_promotions"`
	ActivePromotions int     `json:"active_promotions"`
	TotalImpressions int64   `json:"total_impressions"`
	TotalClicks      int64   `json:"total_clicks"`
	TotalOrders      int64   `json:"total_orders"`
	TotalAmount      float64 `json:"total_amount"`
	AvgCTR           float64 `json:"avg_ctr"`     // 点击率
	AvgROI           float64 `json:"avg_roi"`     // 投入产出比
}

// ===== 物流 =====

// LogisticsCreateRequest 创建物流请求
type LogisticsCreateRequest struct {
	ExpressCompany string `json:"express_company" binding:"required,max=50"`
	ExpressCode    string `json:"express_code" binding:"max=32"`
	TrackingNo     string `json:"tracking_no" binding:"required,max=64"`
	ShipperName    string `json:"shipper_name" binding:"max=50"`
	ShipperPhone   string `json:"shipper_phone" binding:"max=20"`
	ShipperAddress string `json:"shipper_address" binding:"max=500"`
	Weight         float64 `json:"weight" binding:"gte=0"`
	Freight        float64 `json:"freight" binding:"gte=0"`
	Remark         string `json:"remark" binding:"max=500"`
}

// LogisticsUpdateRequest 更新物流请求
type LogisticsUpdateRequest struct {
	Status         *int   `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
	TrackingInfo   []map[string]interface{} `json:"tracking_info"`
	Remark         string `json:"remark" binding:"max=500"`
}

// LogisticsResponse 物流详情响应
type LogisticsResponse struct {
	ID              uint       `json:"id"`
	OrderID         uint       `json:"order_id"`
	ErshouID        uint       `json:"ershou_id"`
	SKUID           uint       `json:"sku_id"`
	ExpressCompany  string     `json:"express_company"`
	ExpressCode     string     `json:"express_code"`
	TrackingNo      string     `json:"tracking_no"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	ShipperName     string     `json:"shipper_name"`
	ShipperPhone    string     `json:"shipper_phone"`
	ShipperAddress  string     `json:"shipper_address"`
	ReceiverName    string     `json:"receiver_name"`
	ReceiverPhone   string     `json:"receiver_phone"`
	ReceiverAddress string     `json:"receiver_address"`
	Weight          float64    `json:"weight"`
	Freight         float64    `json:"freight"`
	TrackingInfo    []map[string]interface{} `json:"tracking_info"`
	ShippedAt       *time.Time `json:"shipped_at"`
	DeliveredAt     *time.Time `json:"delivered_at"`
	Remark          string     `json:"remark"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ===== 担保 =====

// EscrowCreateRequest 创建担保请求
type EscrowCreateRequest struct {
	EscrowAmount float64 `json:"escrow_amount" binding:"required,gte=0"`
	PlatformFee  float64 `json:"platform_fee" binding:"gte=0"`
}

// EscrowReleaseRequest 放款请求
type EscrowReleaseRequest struct {
	DisputeReason string `json:"dispute_reason" binding:"max=500"`
}

// EscrowResponse 担保详情响应
type EscrowResponse struct {
	ID                uint       `json:"id"`
	OrderID           uint       `json:"order_id"`
	ErshouID          uint       `json:"ershou_id"`
	BuyerID           uint       `json:"buyer_id"`
	SellerID          uint       `json:"seller_id"`
	EscrowAmount      float64    `json:"escrow_amount"`
	PlatformFee       float64    `json:"platform_fee"`
	SellerAmount      float64    `json:"seller_amount"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	FrozenAt          *time.Time `json:"frozen_at"`
	ReleaseAt         *time.Time `json:"release_at"`
	PaidAt            *time.Time `json:"paid_at"`
	RefundedAt        *time.Time `json:"refunded_at"`
	AutoReleaseAt     *time.Time `json:"auto_release_at"`
	DisputeReason     string     `json:"dispute_reason"`
	ArbitrationResult string     `json:"arbitration_result"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ===== 退款 =====

// RefundCreateRequest 申请退款请求
type RefundCreateRequest struct {
	RefundType     string                 `json:"refund_type" binding:"required,oneof=return exchange repair refund"`
	Reason         string                 `json:"reason" binding:"required,max=500"`
	Description    string                 `json:"description"`
	EvidenceImages []string               `json:"evidence_images"`
	RefundAmount   float64                `json:"refund_amount" binding:"required,gte=0"`
}

// RefundProcessRequest 处理退款请求
type RefundProcessRequest struct {
	Action      string `json:"action" binding:"required,oneof=approve reject arbitrate"`
	SellerReason string `json:"seller_reason" binding:"max=500"`
	ArbitrationResult string `json:"arbitration_result"`
}

// RefundResponse 退款详情响应
type RefundResponse struct {
	ID                uint       `json:"id"`
	RefundNo          string     `json:"refund_no"`
	OrderID           uint       `json:"order_id"`
	ErshouID          uint       `json:"ershou_id"`
	BuyerID           uint       `json:"buyer_id"`
	SellerID          uint       `json:"seller_id"`
	RefundType        string     `json:"refund_type"`
	Reason            string     `json:"reason"`
	Description       string     `json:"description"`
	EvidenceImages    []string   `json:"evidence_images"`
	RefundAmount      float64    `json:"refund_amount"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	SellerReason      string     `json:"seller_reason"`
	ArbitrationResult string     `json:"arbitration_result"`
	ArbitratorID      uint       `json:"arbitrator_id"`
	ArbitratedAt      *time.Time `json:"arbitrated_at"`
	SLADeadline       *time.Time `json:"sla_deadline"`
	CompletedAt       *time.Time `json:"completed_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ===== 举报 =====

// ReportCreateRequest 创建举报请求
type ReportCreateRequest struct {
	ErshouID       uint     `json:"ershou_id" binding:"required"`
	ReportType     string   `json:"report_type" binding:"required,oneof=porn scam fake prohibited infringement spam other"`
	Reason         string   `json:"reason" binding:"required,max=500"`
	Description    string   `json:"description"`
	EvidenceImages []string `json:"evidence_images"`
}

// ReportProcessRequest 处理举报请求
type ReportProcessRequest struct {
	Status       int    `json:"status" binding:"oneof=1 2 3 4 5"`
	HandleResult string `json:"handle_result"`
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
}

// ReportResponse 举报详情响应
type ReportResponse struct {
	ID                uint       `json:"id"`
	ReportNo          string     `json:"report_no"`
	ErshouID          uint       `json:"ershou_id"`
	ReporterID        uint       `json:"reporter_id"`
	ReporterName      string     `json:"reporter_name"`
	ReportedUserID    uint       `json:"reported_user_id"`
	ReportedUserName  string     `json:"reported_user_name"`
	ReportType        string     `json:"report_type"`
	Reason            string     `json:"reason"`
	Description       string     `json:"description"`
	EvidenceImages    []string   `json:"evidence_images"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	HandlerID         uint       `json:"handler_id"`
	HandlerName       string     `json:"handler_name"`
	HandleResult      string     `json:"handle_result"`
	PenaltyType       string     `json:"penalty_type"`
	PenaltyUserID     uint       `json:"penalty_user_id"`
	SLADeadline       *time.Time `json:"sla_deadline"`
	HandledAt         *time.Time `json:"handled_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	Status      *int   `form:"status" json:"status"`
	ReportType  string `form:"report_type" json:"report_type"`
	ErshouID    uint   `form:"ershou_id" json:"ershou_id"`
	ReporterID  uint   `form:"reporter_id" json:"reporter_id"`
	utils.Pagination
}

// ===== 评价 =====

// ReviewCreateRequest 创建评价请求
type ReviewCreateRequest struct {
	Rating        int      `json:"rating" binding:"required,min=1,max=5"`
	Content       string   `json:"content" binding:"max=1000"`
	Images        []string `json:"images"`
	VideoURL      string   `json:"video_url" binding:"max=255"`
	IsAnonymous   bool     `json:"is_anonymous"`
	IsRecommended bool     `json:"is_recommended"`
	Tags          []string `json:"tags"`
}

// ReviewReplyRequest 评价回复请求
type ReviewReplyRequest struct {
	Reply string `json:"reply" binding:"required,max=500"`
}

// ReviewResponse 评价详情响应
type ReviewResponse struct {
	ID            uint       `json:"id"`
	OrderID       uint       `json:"order_id"`
	ErshouID      uint       `json:"ershou_id"`
	ReviewerID    uint       `json:"reviewer_id"`
	ReviewerName  string     `json:"reviewer_name"`
	ReviewerAvatar string    `json:"reviewer_avatar"`
	RevieweeID    uint       `json:"reviewee_id"`
	ReviewType    string     `json:"review_type"`
	Rating        int        `json:"rating"`
	Content       string     `json:"content"`
	Images        []string   `json:"images"`
	VideoURL      string     `json:"video_url"`
	IsAnonymous   bool       `json:"is_anonymous"`
	IsRecommended bool       `json:"is_recommended"`
	Tags          []string   `json:"tags"`
	Reply         string     `json:"reply"`
	ReplyAt       *time.Time `json:"reply_at"`
	AppendContent string     `json:"append_content"`
	AppendImages  []string   `json:"append_images"`
	AppendAt      *time.Time `json:"append_at"`
	LikeCount     int        `json:"like_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	ErshouID   uint   `form:"ershou_id" json:"ershou_id"`
	ReviewerID uint   `form:"reviewer_id" json:"reviewer_id"`
	Rating     *int   `form:"rating" json:"rating"`
	Sort       string `form:"sort" json:"sort"` // latest/highest/lowest
	utils.Pagination
}

// ===== 店铺 =====

// ShopCreateRequest 创建店铺请求
type ShopCreateRequest struct {
	ShopName        string   `json:"shop_name" binding:"required,max=128"`
	Logo            string   `json:"logo" binding:"max=255"`
	Banner          string   `json:"banner" binding:"max=255"`
	Description     string   `json:"description"`
	ContactName     string   `json:"contact_name" binding:"max=50"`
	ContactPhone    string   `json:"contact_phone" binding:"max=20"`
	ContactWechat   string   `json:"contact_wechat" binding:"max=50"`
	Address         string   `json:"address" binding:"max=500"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	BusinessLicense string   `json:"business_license" binding:"max=255"`
	LicenseNo       string   `json:"license_no" binding:"max=64"`
	IDCardFront     string   `json:"id_card_front" binding:"max=255"`
	IDCardBack      string   `json:"id_card_back" binding:"max=255"`
	Tags            []string `json:"tags"`
	Deposit         float64  `json:"deposit" binding:"gte=0"`
}

// ShopUpdateRequest 更新店铺请求
type ShopUpdateRequest struct {
	ShopName    string   `json:"shop_name" binding:"max=128"`
	Logo        string   `json:"logo" binding:"max=255"`
	Banner      string   `json:"banner" binding:"max=255"`
	Description string   `json:"description"`
	Address     string   `json:"address" binding:"max=500"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Tags        []string `json:"tags"`
}

// ShopResponse 店铺详情响应
type ShopResponse struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	ShopName        string     `json:"shop_name"`
	Logo            string     `json:"logo"`
	Banner          string     `json:"banner"`
	Description     string     `json:"description"`
	Level           int        `json:"level"`
	LevelText       string     `json:"level_text"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	ContactName     string     `json:"contact_name"`
	ContactPhone    string     `json:"contact_phone"`
	ContactWechat   string     `json:"contact_wechat"`
	Address         string     `json:"address"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	BusinessLicense string     `json:"business_license"`
	LicenseNo       string     `json:"license_no"`
	VerifiedAt      *time.Time `json:"verified_at"`
	FollowerCount   int        `json:"follower_count"`
	ItemCount       int        `json:"item_count"`
	SoldCount       int        `json:"sold_count"`
	TotalAmount     float64    `json:"total_amount"`
	GoodRate        float64    `json:"good_rate"`
	Deposit         float64    `json:"deposit"`
	Tags            []string   `json:"tags"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RejectedReason  string     `json:"rejected_reason"`
	IsFollowing     bool       `json:"is_following,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ShopListQuery 店铺列表查询
type ShopListQuery struct {
	UserID uint   `form:"user_id" json:"user_id"`
	Status *int   `form:"status" json:"status"`
	Level  *int   `form:"level" json:"level"`
	Keyword string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ShopFollowRequest 关注店铺请求
type ShopFollowRequest struct {
	Notify bool `json:"notify"`
}

// ShopFollowerListQuery 粉丝列表查询
type ShopFollowerListQuery struct {
	utils.Pagination
}

// ===== 标签 =====

// TagCreateRequest 创建/更新标签请求
type TagCreateRequest struct {
	Name       string `json:"name" binding:"required,max=64"`
	Type       string `json:"type" binding:"omitempty,oneof=smart operation custom"`
	Color      string `json:"color" binding:"max=16"`
	Icon       string `json:"icon" binding:"max=64"`
	Background string `json:"background" binding:"max=32"`
	Status     int    `json:"status" binding:"oneof=0 1"`
	Sort       int    `json:"sort"`
	IsHot      bool   `json:"is_hot"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Color      string    `json:"color"`
	Icon       string    `json:"icon"`
	Background string    `json:"background"`
	Status     int       `json:"status"`
	Sort       int       `json:"sort"`
	UseCount   int       `json:"use_count"`
	IsHot      bool      `json:"is_hot"`
	CreatorID  uint      `json:"creator_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ===== 品牌型号 =====

// BrandCreateRequest 创建/更新品牌请求
type BrandCreateRequest struct {
	Name             string   `json:"name" binding:"required,max=128"`
	Logo             string   `json:"logo" binding:"max=255"`
	EnglishName      string   `json:"english_name" binding:"max=128"`
	Description      string   `json:"description"`
	Country          string   `json:"country" binding:"max=32"`
	OfficialVerified bool     `json:"official_verified"`
	OfficialURL      string   `json:"official_url" binding:"max=255"`
	CategoryIDs      []uint   `json:"category_ids"`
	Status           int      `json:"status" binding:"oneof=0 1"`
	Sort             int      `json:"sort"`
	Tags             []string `json:"tags,omitempty"`
}

// BrandResponse 品牌响应
type BrandResponse struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	Logo              string    `json:"logo"`
	EnglishName       string    `json:"english_name"`
	Description       string    `json:"description"`
	Country           string    `json:"country"`
	OfficialVerified  bool      `json:"official_verified"`
	OfficialURL       string    `json:"official_url"`
	CategoryIDs       []uint    `json:"category_ids"`
	Status            int       `json:"status"`
	Sort              int       `json:"sort"`
	UseCount          int       `json:"use_count"`
	CreatedAt         time.Time `json:"created_at"`
}

// ModelCreateRequest 创建/更新型号请求
type ModelCreateRequest struct {
	Name            string                 `json:"name" binding:"required,max=128"`
	FullName        string                 `json:"full_name" binding:"max=255"`
	Specifications  map[string]interface{} `json:"specifications"`
	Image           string                 `json:"image" binding:"max=255"`
	ReleaseDate     string                 `json:"release_date" binding:"max=16"`
	Status          int                    `json:"status" binding:"oneof=0 1"`
	Sort            int                    `json:"sort"`
	ReferencePrice  float64                `json:"reference_price" binding:"gte=0"`
}

// ModelResponse 型号响应
type ModelResponse struct {
	ID              uint      `json:"id"`
	BrandID         uint      `json:"brand_id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Specifications  map[string]interface{} `json:"specifications"`
	Image           string    `json:"image"`
	ReleaseDate     string    `json:"release_date"`
	Status          int       `json:"status"`
	Sort            int       `json:"sort"`
	UseCount        int       `json:"use_count"`
	ReferencePrice  float64   `json:"reference_price"`
	CreatedAt       time.Time `json:"created_at"`
}

// ===== 分类属性 =====

// CategoryAttrCreateRequest 创建/更新分类属性请求
type CategoryAttrCreateRequest struct {
	CategoryID    uint     `json:"category_id" binding:"required"`
	AttrName      string   `json:"attr_name" binding:"required,max=64"`
	AttrKey       string   `json:"attr_key" binding:"max=64"`
	AttrType      string   `json:"attr_type" binding:"omitempty,oneof=string number select multi_select date boolean"`
	Options       []string `json:"options"`
	Unit          string   `json:"unit" binding:"max=32"`
	IsRequired    bool     `json:"is_required"`
	IsFilterable  bool     `json:"is_filterable"`
	IsSearchable  bool     `json:"is_searchable"`
	DefaultValue  string   `json:"default_value" binding:"max=255"`
	Placeholder   string   `json:"placeholder" binding:"max=255"`
	Description   string   `json:"description" binding:"max=500"`
	Status        int      `json:"status" binding:"oneof=0 1"`
	Sort          int      `json:"sort"`
}

// CategoryAttrResponse 分类属性响应
type CategoryAttrResponse struct {
	ID            uint       `json:"id"`
	CategoryID    uint       `json:"category_id"`
	AttrName      string     `json:"attr_name"`
	AttrKey       string     `json:"attr_key"`
	AttrType      string     `json:"attr_type"`
	Options       []string   `json:"options"`
	Unit          string     `json:"unit"`
	IsRequired    bool       `json:"is_required"`
	IsFilterable  bool       `json:"is_filterable"`
	IsSearchable  bool       `json:"is_searchable"`
	DefaultValue  string     `json:"default_value"`
	Placeholder   string     `json:"placeholder"`
	Description   string     `json:"description"`
	Status        int        `json:"status"`
	Sort          int        `json:"sort"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ===== 审核规则 =====

// AuditRuleCreateRequest 创建/更新审核规则请求
type AuditRuleCreateRequest struct {
	RuleName    string                 `json:"rule_name" binding:"required,max=128"`
	RuleType    string                 `json:"rule_type" binding:"required,oneof=sensitive_word price_check frequency prohibited content"`
	RuleKey     string                 `json:"rule_key" binding:"max=64"`
	Pattern     string                 `json:"pattern"`
	Threshold   map[string]interface{} `json:"threshold"`
	Action      string                 `json:"action" binding:"omitempty,oneof=reject approval warning manual"`
	PenaltyType string                 `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
	Severity    int                    `json:"severity" binding:"min=1,max=5"`
	Status      int                    `json:"status" binding:"oneof=0 1"`
	Description string                 `json:"description" binding:"max=500"`
	Sort        int                    `json:"sort"`
}

// AuditRuleResponse 审核规则响应
type AuditRuleResponse struct {
	ID          uint       `json:"id"`
	RuleName    string     `json:"rule_name"`
	RuleType    string     `json:"rule_type"`
	RuleKey     string     `json:"rule_key"`
	Pattern     string     `json:"pattern"`
	Threshold   map[string]interface{} `json:"threshold"`
	Action      string     `json:"action"`
	PenaltyType string     `json:"penalty_type"`
	Severity    int        `json:"severity"`
	Status      int        `json:"status"`
	Description string     `json:"description"`
	Sort        int        `json:"sort"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ===== 用户信用 =====

// UserCreditResponse 用户信用响应
type UserCreditResponse struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	CreditScore       int        `json:"credit_score"`
	CreditLevel       int        `json:"credit_level"`
	CreditLevelText   string     `json:"credit_level_text"`
	TotalTransactions int        `json:"total_transactions"`
	SuccessTransactions int      `json:"success_transactions"`
	CancelTransactions int       `json:"cancel_transactions"`
	GoodReviews       int        `json:"good_reviews"`
	MediumReviews     int        `json:"medium_reviews"`
	BadReviews        int        `json:"bad_reviews"`
	GoodRate          float64    `json:"good_rate"`
	Disputes          int        `json:"disputes"`
	Reports           int        `json:"reports"`
	Penalties         int        `json:"penalties"`
	LastTransactionAt *time.Time `json:"last_transaction_at"`
	FrozenReason      string     `json:"frozen_reason"`
	FrozenUntil       *time.Time `json:"frozen_until"`
	IsFrozen          bool       `json:"is_frozen"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UserCreditUpdateRequest 更新用户信用请求
type UserCreditUpdateRequest struct {
	CreditScore    int    `json:"credit_score"`
	CreditLevel    int    `json:"credit_level" binding:"min=0,max=5"`
	FrozenReason   string `json:"frozen_reason" binding:"max=500"`
	FrozenUntil    *time.Time `json:"frozen_until"`
}

// ===== 扩展列表查询 =====

// ErshouListQuery 扩展列表查询（支持品牌/型号/标签/SKU/拍卖/担保等筛选）
type ErshouListQuery struct {
	CategoryID    uint    `form:"category_id" json:"category_id"`
	Keyword       string  `form:"keyword" json:"keyword"`
	MinPrice      float64 `form:"min_price" json:"min_price"`
	MaxPrice      float64 `form:"max_price" json:"max_price"`
	Condition     string  `form:"condition" json:"condition"`
	Brand         string  `form:"brand" json:"brand"`
	BrandID       uint    `form:"brand_id" json:"brand_id"`
	ModelID       uint    `form:"model_id" json:"model_id"`
	ShopID        uint    `form:"shop_id" json:"shop_id"`
	IsAuction     *bool   `form:"is_auction" json:"is_auction"`
	EscrowEnabled *bool   `form:"escrow_enabled" json:"escrow_enabled"`
	FreeShipping  *bool   `form:"free_shipping" json:"free_shipping"`
	IsUrgent      *bool   `form:"is_urgent" json:"is_urgent"`
	Featured      *bool   `form:"featured" json:"featured"`
	Verified      *bool   `form:"verified" json:"verified"`
	TagIDs        []uint  `form:"tag_ids" json:"tag_ids"`
	Sort          string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/popular/hot
	Latitude      float64 `form:"latitude" json:"latitude"`
	Longitude     float64 `form:"longitude" json:"longitude"`
	RadiusKm      float64 `form:"radius_km" json:"radius_km"`
	utils.Pagination
}

// ErshouDetailResponse 聚合详情响应（主信息+SKU+图片+评价+店铺+推广状态+物流+担保）
type ErshouDetailResponse struct {
	ErshouInfo             // 嵌入基础信息
	SKUs          []SKUResponse     `json:"skus"`
	Reviews       []ReviewResponse  `json:"reviews"`
	ReviewStats   ReviewStats       `json:"review_stats"`
	Shop          *ShopResponse     `json:"shop,omitempty"`
	Promotion     *PromotionResponse `json:"promotion,omitempty"`
	Logistics     *LogisticsResponse `json:"logistics,omitempty"`
	Escrow        *EscrowResponse    `json:"escrow,omitempty"`
	Auction       *AuctionResponse   `json:"auction,omitempty"`
}

// ReviewStats 评价统计
type ReviewStats struct {
	TotalReviews int     `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
}

// ===== 统计 =====

// StatisticsResponse 数据统计响应
type StatisticsResponse struct {
	TotalItems       int64   `json:"total_items"`
	TodayNewItems    int64   `json:"today_new_items"`
	ActiveSellers    int64   `json:"active_sellers"`
	TotalOrders      int64   `json:"total_orders"`
	TodayOrders      int64   `json:"today_orders"`
	TotalAmount      float64 `json:"total_amount"`
	TodayAmount      float64 `json:"today_amount"`
	CompletedRate    float64 `json:"completed_rate"`
	RefundRate       float64 `json:"refund_rate"`
	AvgPrice         float64 `json:"avg_price"`
}

// SellerStatisticsResponse 卖家数据统计
type SellerStatisticsResponse struct {
	UserID           uint    `json:"user_id"`
	TotalItems       int64   `json:"total_items"`
	SoldItems        int64   `json:"sold_items"`
	TotalOrders      int64   `json:"total_orders"`
	CompletedOrders  int64   `json:"completed_orders"`
	TotalAmount      float64 `json:"total_amount"`
	AvgRating        float64 `json:"avg_rating"`
	Followers        int64   `json:"followers"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// HotItemResponse 热门商品
type HotItemResponse struct {
	ErshouID    uint    `json:"ershou_id"`
	Title       string  `json:"title"`
	CoverImage  string  `json:"cover_image"`
	Price       float64 `json:"price"`
	ViewCount   int     `json:"view_count"`
	FavCount    int     `json:"fav_count"`
	OrderCount  int     `json:"order_count"`
}

// PriceTrendResponse 价格趋势
type PriceTrendResponse struct {
	Dates  []string  `json:"dates"`
	Prices []float64 `json:"prices"`
	Brand  string    `json:"brand,omitempty"`
}

// ===== 批量操作 =====

// BatchAuditRequest 批量审核请求
type BatchAuditRequest struct {
	IDs         []uint `json:"ids" binding:"required,min=1"`
	AuditStatus int    `json:"audit_status" binding:"oneof=1 2"`
	AuditReason string `json:"audit_reason" binding:"max=500"`
}

// BatchStatusUpdateRequest 批量状态变更请求
type BatchStatusUpdateRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Status int    `json:"status" binding:"oneof=1 3 4"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// BatchResultResponse 批量操作结果
type BatchResultResponse struct {
	Total     int      `json:"total"`
	Success   int      `json:"success"`
	Failed    int      `json:"failed"`
	FailedIDs []uint   `json:"failed_ids,omitempty"`
}

// ===== 导出 =====

// ExportRequest 导出 Excel/CSV 请求
type ExportRequest struct {
	Format   string `json:"format" binding:"required,oneof=excel csv"`
	Status   *int   `json:"status"`
	CategoryID uint `json:"category_id"`
	UserID   uint   `json:"user_id"`
	Keyword  string `json:"keyword"`
}

// ===== 搜索推荐 =====

// AdvancedSearchRequest 高级搜索请求
type AdvancedSearchRequest struct {
	Keyword       string  `form:"keyword" json:"keyword"`
	CategoryID    uint    `form:"category_id" json:"category_id"`
	BrandID       uint    `form:"brand_id" json:"brand_id"`
	ModelID       uint    `form:"model_id" json:"model_id"`
	TagIDs        []uint  `form:"tag_ids" json:"tag_ids"`
	Condition     string  `form:"condition" json:"condition"`
	MinPrice      float64 `form:"min_price" json:"min_price"`
	MaxPrice      float64 `form:"max_price" json:"max_price"`
	RegionID      uint    `form:"region_id" json:"region_id"`
	Latitude      float64 `form:"latitude" json:"latitude"`
	Longitude     float64 `form:"longitude" json:"longitude"`
	RadiusKm      float64 `form:"radius_km" json:"radius_km"`
	Sort          string  `form:"sort" json:"sort"` // latest/price_asc/price_desc/distance/popular
	IsAuction     *bool   `form:"is_auction" json:"is_auction"`
	FreeShipping  *bool   `form:"free_shipping" json:"free_shipping"`
	utils.Pagination
}

// HotSearchResponse 热搜词响应
type HotSearchResponse struct {
	Keyword    string `json:"keyword"`
	SearchCount int64 `json:"search_count"`
	Rank       int    `json:"rank"`
}

// SearchHistoryResponse 搜索历史响应
type SearchHistoryResponse struct {
	Keyword    string    `json:"keyword"`
	SearchedAt time.Time `json:"searched_at"`
}

// SimilarItemResponse 相似推荐响应
type SimilarItemResponse struct {
	ErshouID   uint    `json:"ershou_id"`
	Title      string  `json:"title"`
	CoverImage string  `json:"cover_image"`
	Price      float64 `json:"price"`
	Brand      string  `json:"brand"`
	Similarity float64 `json:"similarity"` // 相似度 0-1
}
