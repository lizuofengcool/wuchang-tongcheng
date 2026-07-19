// Package model 担保交易 + 成交记录 + 推荐记录表
// HouseEscrow 担保交易；HouseDeal 成交；HouseRecommendation 推荐
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 担保类型常量 ===
const (
	EscrowTypeDeposit       = "deposit"        // 定金托管
	EscrowTypeCommission    = "commission"     // 佣金托管
	EscrowTypeFullPayment   = "full_payment"   // 全款托管
	EscrowTypeRent          = "rent"           // 租金托管
	EscrowTypeDownPayment   = "down_payment"   // 首付托管
)

// === 担保状态常量 ===
const (
	EscrowStatusPending      = 0 // 待支付
	EscrowStatusPaid         = 1 // 已支付（冻结中）
	EscrowStatusReleased     = 2 // 已解冻（放款）
	EscrowStatusRefunded     = 3 // 已退款
	EscrowStatusDisputed     = 4 // 争议中
	EscrowStatusArbitrated   = 5 // 已仲裁
	EscrowStatusCanceled     = 6 // 已取消
)

// === 支付方式常量 ===
const (
	EscrowPayWechat  = "wechat"  // 微信支付
	EscrowPayAlipay  = "alipay"  // 支付宝
	EscrowPayBank    = "bank"    // 银行卡
	EscrowPayBalance = "balance" // 余额支付
)

// === 成交类型常量 ===
const (
	DealTypeSale  = "sale"  // 买卖
	DealTypeRent  = "rent"  // 租赁
	DealTypeTransfer = "transfer" // 转让
)

// === 成交状态常量 ===
const (
	DealStatusPending    = 0 // 待确认
	DealStatusConfirmed  = 1 // 已确认
	DealStatusCompleted  = 2 // 已完成
	DealStatusCanceled   = 3 // 已取消
	DealStatusDisputed   = 4 // 争议中
)

// === 推荐类型常量 ===
const (
	RecTypeHouseToUser   = "house_to_user"   // 房荐人
	RecTypeUserToHouse   = "user_to_house"   // 人荐房
	RecTypeSimilar       = "similar"         // 相似房源
	RecTypeNearby        = "nearby"          // 附近房源
	RecTypeRecentlyViewed = "recently_viewed" // 看过又推荐
)

// === 推荐来源常量 ===
const (
	RecSourceAI      = "ai"       // AI 推荐
	RecSourceManual  = "manual"   // 人工推荐
	RecSourceHot     = "hot"      // 热门推荐
	RecSourceNew     = "new"      // 新房推荐
)

// === 推荐状态常量 ===
const (
	RecStatusPending   = 0 // 待展示
	RecStatusShown     = 1 // 已展示
	RecStatusClicked   = 2 // 已点击
	RecStatusContacted = 3 // 已联系
	RecStatusDismissed = 4 // 已忽略
	RecStatusExpired   = 5 // 已过期
)

// HouseEscrow 担保交易表
type HouseEscrow struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	EscrowNo        string     `gorm:"size:64;not null;uniqueIndex" json:"escrow_no"`                       // 担保单号
	EscrowType      string     `gorm:"size:32;not null;default:'deposit';index" json:"escrow_type"`          // deposit/commission/full_payment/rent/down_payment
	HouseID         uint       `gorm:"not null;default:0;index" json:"house_id"`                            // 房源 ID
	ListingID       uint       `gorm:"not null;default:0;index" json:"listing_id"`                          // 发布 ID
	ContractID      uint       `gorm:"not null;default:0;index" json:"contract_id"`                         // 合同 ID
	CommunityID     uint       `gorm:"not null;default:0;index" json:"community_id"`                        // 小区 ID
	PayerID         uint       `gorm:"not null;index" json:"payer_id"`                                       // 付款方 ID
	PayeeID         uint       `gorm:"not null;index" json:"payee_id"`                                       // 收款方 ID
	AgentID         uint       `gorm:"not null;default:0;index" json:"agent_id"`                             // 经纪人 ID
	Amount          float64    `gorm:"type:decimal(14,2);default:0" json:"amount"`                          // 担保金额
	PlatformFee     float64    `gorm:"type:decimal(14,2);default:0" json:"platform_fee"`                    // 平台手续费
	AgentFee        float64    `gorm:"type:decimal(14,2);default:0" json:"agent_fee"`                       // 经纪人佣金
	PayeeAmount     float64    `gorm:"type:decimal(14,2);default:0" json:"payee_amount"`                    // 收款方实收
	Status          int        `gorm:"default:1;index" json:"status"`                                        // 0待支付 1已支付 2已放款 3已退款 4争议 5仲裁 6取消
	PayMethod       string     `gorm:"size:32;not null;default:'wechat'" json:"pay_method"`                 // wechat/alipay/bank/balance
	PayTradeNo      string     `gorm:"size:128;not null;default:'';index" json:"pay_trade_no"`              // 第三方支付单号
	PaidAt          *time.Time `gorm:"index" json:"paid_at"`                                                 // 支付时间
	FrozenAt        *time.Time `gorm:"index" json:"frozen_at"`                                               // 冻结时间
	ReleaseAt       *time.Time `gorm:"index" json:"release_at"`                                              // 放款时间
	RefundedAt      *time.Time `gorm:"index" json:"refunded_at"`                                             // 退款时间
	AutoReleaseAt   *time.Time `gorm:"index" json:"auto_release_at"`                                         // 自动放款时间
	DisputeReason   string     `gorm:"size:500;not null;default:''" json:"dispute_reason"`                  // 争议原因
	ArbitrationResult string   `gorm:"type:text" json:"arbitration_result"`                                  // 仲裁结果
	CompletedAt     *time.Time `gorm:"index" json:"completed_at"`                                            // 完成时间
}

// TableName 表名（house_ 前缀）
func (HouseEscrow) TableName() string { return "house_escrows" }

// HouseDeal 成交记录表
type HouseDeal struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	DealNo        string     `gorm:"size:64;not null;uniqueIndex" json:"deal_no"`                          // 成交单号
	HouseID       uint       `gorm:"not null;index" json:"house_id"`                                       // 房源 ID
	ListingID     uint       `gorm:"not null;default:0;index" json:"listing_id"`                           // 发布 ID
	CommunityID   uint       `gorm:"not null;default:0;index" json:"community_id"`                         // 小区 ID
	ContractID    uint       `gorm:"not null;default:0;index" json:"contract_id"`                          // 合同 ID
	EscrowID      uint       `gorm:"not null;default:0;index" json:"escrow_id"`                            // 担保 ID
	DealType      string     `gorm:"size:16;not null;default:'sale';index" json:"deal_type"`               // sale/rent/transfer
	SellerID      uint       `gorm:"not null;default:0;index" json:"seller_id"`                            // 卖方/出租方 ID
	SellerName    string     `gorm:"size:50;not null;default:''" json:"seller_name"`                       // 卖方姓名
	BuyerID       uint       `gorm:"not null;default:0;index" json:"buyer_id"`                             // 买方/承租方 ID
	BuyerName     string     `gorm:"size:50;not null;default:''" json:"buyer_name"`                        // 买方姓名
	AgentID       uint       `gorm:"not null;default:0;index" json:"agent_id"`                              // 经纪人 ID
	AgentName     string     `gorm:"size:50;not null;default:''" json:"agent_name"`                        // 经纪人姓名
	DealPrice     float64    `gorm:"type:decimal(14,2);default:0;index" json:"deal_price"`                 // 成交价
	AveragePrice  float64    `gorm:"type:decimal(10,2);default:0" json:"average_price"`                    // 均价
	Commission    float64    `gorm:"type:decimal(14,2);default:0" json:"commission"`                       // 佣金
	DealDate      *time.Time `gorm:"type:date;index" json:"deal_date"`                                     // 成交日期
	ListedAt      *time.Time `gorm:"index" json:"listed_at"`                                               // 挂牌时间
	DealDays      int        `gorm:"not null;default:0" json:"deal_days"`                                  // 成交周期（天）
	PaymentMethod string     `gorm:"size:32;not null;default:''" json:"payment_method"`                   // 付款方式
	LoanAmount    float64    `gorm:"type:decimal(14,2);default:0" json:"loan_amount"`                      // 贷款金额
	LoanPeriods   int        `gorm:"not null;default:0" json:"loan_periods"`                               // 贷款期数
	Status        int        `gorm:"default:1;index" json:"status"`                                        // 0待确认 1已确认 2完成 3取消 4争议
	CompletedAt   *time.Time `gorm:"index" json:"completed_at"`                                            // 完成时间
	CanceledAt    *time.Time `gorm:"index" json:"canceled_at"`                                              // 取消时间
	CanceledReason string    `gorm:"size:500;not null;default:''" json:"canceled_reason"`                  // 取消原因
	Remark        string     `gorm:"size:500;not null;default:''" json:"remark"`                           // 备注
}

// TableName 表名（house_ 前缀）
func (HouseDeal) TableName() string { return "house_deals" }

// HouseRecommendation 推荐记录表
type HouseRecommendation struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID           uint       `gorm:"not null;index;uniqueIndex:uniq_house_recs_user_house_type" json:"user_id"`         // 用户 ID
	HouseID          uint       `gorm:"not null;index;uniqueIndex:uniq_house_recs_user_house_type" json:"house_id"`        // 房源 ID
	RecType          string     `gorm:"size:32;not null;default:'house_to_user';index;uniqueIndex:uniq_house_recs_user_house_type" json:"rec_type"` // house_to_user/user_to_house/similar/nearby/recently_viewed
	Source           string     `gorm:"size:32;not null;default:'ai';index" json:"source"`                                  // ai/manual/hot/new
	Score            float64    `gorm:"type:decimal(5,2);default:0;index" json:"score"`                                     // 推荐评分
	Reason           string     `gorm:"size:500;not null;default:''" json:"reason"`                                         // 推荐理由
	PriceMatch       float64    `gorm:"type:decimal(5,2);default:0" json:"price_match"`                                     // 价格匹配度
	LocationMatch    float64    `gorm:"type:decimal(5,2);default:0" json:"location_match"`                                  // 位置匹配度
	LayoutMatch      float64    `gorm:"type:decimal(5,2);default:0" json:"layout_match"`                                    // 户型匹配度
	FacilityMatch    float64    `gorm:"type:decimal(5,2);default:0" json:"facility_match"`                                  // 配套匹配度
	Status           int        `gorm:"default:0;index" json:"status"`                                                       // 0待展示 1已展示 2点击 3联系 4忽略 5过期
	ClickedAt        *time.Time `gorm:"index" json:"clicked_at"`                                                             // 点击时间
	ContactedAt      *time.Time `gorm:"index" json:"contacted_at"`                                                           // 联系时间
	ViewedAt         *time.Time `gorm:"index" json:"viewed_at"`                                                              // 查看时间
	DismissedAt      *time.Time `gorm:"index" json:"dismissed_at"`                                                           // 忽略时间
	ExpiredAt        *time.Time `gorm:"index" json:"expired_at"`                                                             // 过期时间
}

// TableName 表名（house_ 前缀）
func (HouseRecommendation) TableName() string { return "house_recommendations" }
