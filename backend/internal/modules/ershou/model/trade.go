// Package model 付费推广 + 物流 + 担保 + 退款（对标闲鱼/转转/瓜子）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 推广类型常量 ===
const (
	PromotionTypeHomeBanner   = "home_banner"   // 首页轮播
	PromotionTypeChannelTop   = "channel_top"   // 频道置顶
	PromotionTypeSearchTop    = "search_top"    // 搜索置顶
	PromotionTypeFeatured     = "featured"      // 精品推荐
	PromotionTypeUrgent       = "urgent"        // 加急置顶
	PromotionTypeRefresh      = "refresh"       // 刷新曝光
)

// === 推广状态常量 ===
const (
	PromotionStatusPending  = 0 // 待支付
	PromotionStatusActive   = 1 // 推广中
	PromotionStatusEnded    = 2 // 已结束
	PromotionStatusCanceled = 3 // 已取消
)

// ErshouPromotion 付费推广记录表
type ErshouPromotion struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ErshouID       uint       `gorm:"not null;index:idx_promo_ershou_status" json:"ershou_id"`              // 关联二手物品ID
	UserID         uint       `gorm:"not null;index" json:"user_id"`                                          // 推广用户ID
	PromotionType  string     `gorm:"size:32;not null;index" json:"promotion_type"`                          // 推广类型 home_banner/channel_top/search_top/featured/urgent/refresh
	Status         int        `gorm:"default:0;index:idx_promo_ershou_status" json:"status"`                  // 0待支付 1推广中 2已结束 3已取消
	StartTime      *time.Time `gorm:"index" json:"start_time"`                                                // 推广开始时间
	EndTime        *time.Time `gorm:"index" json:"end_time"`                                                  // 推广结束时间
	DurationDays   int        `gorm:"default:1" json:"duration_days"`                                         // 推广天数
	Amount         float64    `gorm:"type:decimal(12,2);default:0" json:"amount"`                             // 推广费用
	PayMethod      string     `gorm:"size:32" json:"pay_method"`                                              // 支付方式
	PayTradeNo     string     `gorm:"size:128;index" json:"pay_trade_no"`                                     // 支付流水号
	PaidAt         *time.Time `gorm:"index" json:"paid_at"`                                                    // 支付时间
	ImpressionCount int      `gorm:"default:0" json:"impression_count"`                                       // 曝光数
	ClickCount      int       `gorm:"default:0" json:"click_count"`                                            // 点击数
	FavCount        int       `gorm:"default:0" json:"fav_count"`                                              // 收藏数
	ConsultCount    int       `gorm:"default:0" json:"consult_count"`                                          // 咨询数
	OrderCount      int       `gorm:"default:0" json:"order_count"`                                            // 下单数
	ROI             float64   `gorm:"type:decimal(5,2);default:0" json:"roi"`                                  // 投入产出比
}

// TableName 表名（ers_ 前缀）
func (ErshouPromotion) TableName() string { return "ers_promotions" }

// === 物流状态常量 ===
const (
	LogisticsStatusPending   = 0 // 待发货
	LogisticsStatusShipped   = 1 // 已发货
	LogisticsStatusInTransit = 2 // 运输中
	LogisticsStatusDelivered = 3 // 已签收
	LogisticsStatusException = 4 // 异常
)

// ErshouLogistics 物流记录表
type ErshouLogistics struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	OrderID        uint       `gorm:"not null;uniqueIndex" json:"order_id"`                       // 关联订单ID（一对一）
	ErshouID       uint       `gorm:"not null;index" json:"ershou_id"`                            // 关联二手物品ID
	SKUID          uint       `gorm:"index" json:"sku_id"`                                         // 关联 SKU ID
	ExpressCompany string     `gorm:"size:50" json:"express_company"`                            // 快递公司（顺丰/中通/圆通/韵达/申通/京东）
	ExpressCode    string     `gorm:"size:32" json:"express_code"`                                // 快递公司编码
	TrackingNo     string     `gorm:"size:64;not null;index" json:"tracking_no"`                  // 运单号
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0待发货 1已发货 2运输中 3已签收 4异常
	ShipperName    string     `gorm:"size:50" json:"shipper_name"`                                // 发件人姓名
	ShipperPhone   string     `gorm:"size:20" json:"shipper_phone"`                              // 发件人手机
	ShipperAddress string     `gorm:"size:500" json:"shipper_address"`                            // 发件地址
	ReceiverName   string     `gorm:"size:50" json:"receiver_name"`                              // 收件人姓名
	ReceiverPhone  string     `gorm:"size:20" json:"receiver_phone"`                            // 收件人手机
	ReceiverAddress string    `gorm:"size:500" json:"receiver_address"`                          // 收件地址
	Weight         float64    `gorm:"type:decimal(8,2);default:0" json:"weight"`                  // 重量(kg)
	Freight        float64    `gorm:"type:decimal(12,2);default:0" json:"freight"`               // 运费
	TrackingInfo   JSONB       `gorm:"type:jsonb" json:"tracking_info"`                            // 物流跟踪 JSON [{time,location,action}]
	ShippedAt      *time.Time `gorm:"index" json:"shipped_at"`                                    // 发货时间
	DeliveredAt    *time.Time `gorm:"index" json:"delivered_at"`                                  // 签收时间
	Remark         string     `gorm:"size:500" json:"remark"`                                    // 备注
}

// TableName 表名（ers_ 前缀）
func (ErshouLogistics) TableName() string { return "ers_logistics" }

// === 担保状态常量 ===
const (
	EscrowStatusNone      = 0 // 未启用
	EscrowStatusFrozen    = 1 // 资金冻结中
	EscrowStatusReleased  = 2 // 资金已解冻（待放款）
	EscrowStatusPaid       = 3 // 已放款给卖家
	EscrowStatusRefunded   = 4 // 已退款给买家
	EscrowStatusDispute    = 5 // 纠纷处理中
)

// ErshouEscrow 担保交易表（对标闲鱼/瓜子）
// 实现资金托管：买家付款→平台冻结→收货确认→解冻放款给卖家
type ErshouEscrow struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	OrderID        uint       `gorm:"not null;uniqueIndex" json:"order_id"`                       // 关联订单ID（一对一）
	ErshouID       uint       `gorm:"not null;index" json:"ershou_id"`                            // 关联二手物品ID
	BuyerID        uint       `gorm:"not null;index" json:"buyer_id"`                              // 买家ID
	SellerID       uint       `gorm:"not null;index" json:"seller_id"`                            // 卖家ID
	EscrowAmount   float64    `gorm:"type:decimal(12,2);not null" json:"escrow_amount"`            // 托管金额
	PlatformFee    float64    `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`            // 平台手续费
	SellerAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"seller_amount"`           // 卖家到账金额
	Status         int        `gorm:"default:1;index" json:"status"`                               // 0未启用 1冻结中 2已解冻 3已放款 4已退款 5纠纷中
	FrozenAt       *time.Time `gorm:"index" json:"frozen_at"`                                     // 冻结时间
	ReleaseAt      *time.Time `gorm:"index" json:"release_at"`                                    // 解冻时间（确认收货）
	PaidAt         *time.Time `gorm:"index" json:"paid_at"`                                       // 放款给卖家时间
	RefundedAt     *time.Time `gorm:"index" json:"refunded_at"`                                   // 退款给买家时间
	AutoReleaseAt  *time.Time `gorm:"index" json:"auto_release_at"`                              // 自动放款时间（确认收货 N 天后）
	DisputeReason  string     `gorm:"size:500" json:"dispute_reason"`                              // 纠纷原因
	ArbitrationResult string  `gorm:"type:text" json:"arbitration_result"`                        // 仲裁结果
}

// TableName 表名（ers_ 前缀）
func (ErshouEscrow) TableName() string { return "ers_escrows" }

// === 退款类型常量 ===
const (
	RefundTypeReturn   = "return"   // 退货退款
	RefundTypeExchange = "exchange" // 换货
	RefundTypeRepair   = "repair"   // 维修
	RefundTypeRefund   = "refund"   // 仅退款
)

// === 退款状态常量 ===
const (
	RefundStatusPending        = 0 // 待卖家处理
	RefundStatusSellerRejected = 1 // 卖家拒绝
	RefundStatusSellerAgreed   = 2 // 卖家同意
	RefundStatusShipping       = 3 // 退货运输中
	RefundStatusReceived       = 4 // 卖家已收到退货
	RefundStatusRefunding      = 5 // 退款中
	RefundStatusCompleted      = 6 // 退款完成
	RefundStatusCanceled       = 7 // 买家取消
	RefundStatusDispute        = 8 // 平台介入
	RefundStatusArbitrated     = 9 // 仲裁完成
)

// ErshouRefund 退款工单表
type ErshouRefund struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RefundNo       string     `gorm:"size:64;not null;uniqueIndex" json:"refund_no"`              // 退款单号（业务唯一）
	OrderID        uint       `gorm:"not null;index" json:"order_id"`                              // 关联订单ID
	ErshouID       uint       `gorm:"not null;index" json:"ershou_id"`                             // 关联二手物品ID
	BuyerID        uint       `gorm:"not null;index" json:"buyer_id"`                              // 买家ID
	SellerID       uint       `gorm:"not null;index" json:"seller_id"`                            // 卖家ID
	RefundType     string     `gorm:"size:32;not null;index" json:"refund_type"`                   // 退货/换货/维修/仅退款
	Reason         string     `gorm:"size:500;not null" json:"reason"`                             // 退款原因
	Description    string     `gorm:"type:text" json:"description"`                                 // 详细描述
	EvidenceImages JSONB       `gorm:"type:jsonb" json:"evidence_images"`                            // 证据图片 URL 数组
	RefundAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`           // 退款金额
	Status         int        `gorm:"default:0;index" json:"status"`                              // 9状态机
	SellerReason   string     `gorm:"size:500" json:"seller_reason"`                              // 卖家拒绝原因
	ArbitrationResult string  `gorm:"type:text" json:"arbitration_result"`                        // 仲裁结果
	ArbitratorID   uint       `gorm:"index" json:"arbitrator_id"`                                  // 仲裁员ID
	ArbitratedAt   *time.Time `gorm:"index" json:"arbitrated_at"`                                 // 仲裁时间
	SLADeadline    *time.Time `gorm:"index" json:"sla_deadline"`                                  // SLA 处理截止时间（72h）
	CompletedAt    *time.Time `gorm:"index" json:"completed_at"`                                    // 完成时间
}

// TableName 表名（ers_ 前缀）
func (ErshouRefund) TableName() string { return "ers_refunds" }
