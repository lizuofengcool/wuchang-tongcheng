// Package model 担保交易 + 推荐记录表
// CarEscrow 担保交易；CarRecommendation 推荐
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 担保类型常量 ===
const (
	EscrowTypeDeposit     = "deposit"      // 定金托管
	EscrowTypeFullPayment = "full_payment" // 全款托管
	EscrowTypeDownPayment = "down_payment" // 首付托管
	EscrowTypeCommission  = "commission"   // 佣金托管
	EscrowTypeBalance     = "balance"      // 尾款托管
)

// === 担保状态常量 ===
const (
	EscrowStatusPending     = 0 // 待支付
	EscrowStatusPaid        = 1 // 已支付（冻结中）
	EscrowStatusReleased    = 2 // 已解冻（放款）
	EscrowStatusRefunded    = 3 // 已退款
	EscrowStatusDisputed    = 4 // 争议中
	EscrowStatusArbitrated  = 5 // 已仲裁
	EscrowStatusCanceled    = 6 // 已取消
)

// === 支付方式常量 ===
const (
	EscrowPayWechat  = "wechat"  // 微信支付
	EscrowPayAlipay  = "alipay"  // 支付宝
	EscrowPayBank    = "bank"    // 银行卡
	EscrowPayBalance = "balance" // 余额支付
)

// CarEscrow 担保交易表
type CarEscrow struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	EscrowNo         string     `gorm:"size:64;not null;uniqueIndex" json:"escrow_no"`                 // 担保单号
	EscrowType       string     `gorm:"size:32;not null;default:'deposit';index" json:"escrow_type"`    // deposit/full_payment/down_payment/commission/balance
	CarID            uint       `gorm:"not null;default:0;index" json:"car_id"`                         // 车源 ID
	ListingID        uint       `gorm:"not null;default:0;index" json:"listing_id"`                     // 发布 ID
	ContractID       uint       `gorm:"not null;default:0;index" json:"contract_id"`                    // 合同 ID
	PayerID          uint       `gorm:"not null;index" json:"payer_id"`                                 // 付款方 ID
	PayeeID          uint       `gorm:"not null;index" json:"payee_id"`                                 // 收款方 ID
	DealerID         uint       `gorm:"not null;default:0;index" json:"dealer_id"`                      // 车商 ID
	Amount           float64    `gorm:"type:decimal(14,2);default:0" json:"amount"`                    // 担保金额
	PlatformFee      float64    `gorm:"type:decimal(14,2);default:0" json:"platform_fee"`              // 平台手续费
	DealerFee        float64    `gorm:"type:decimal(14,2);default:0" json:"dealer_fee"`                // 车商佣金
	PayeeAmount      float64    `gorm:"type:decimal(14,2);default:0" json:"payee_amount"`              // 收款方实收
	Status           int        `gorm:"default:1;index" json:"status"`                                  // 0待支付 1已支付 2已放款 3已退款 4争议 5仲裁 6取消
	PayMethod        string     `gorm:"size:32;not null;default:'wechat'" json:"pay_method"`            // wechat/alipay/bank/balance
	PayTradeNo       string     `gorm:"size:128;not null;default:'';index" json:"pay_trade_no"`         // 第三方支付单号
	PaidAt           *time.Time `gorm:"index" json:"paid_at"`                                            // 支付时间
	FrozenAt         *time.Time `gorm:"index" json:"frozen_at"`                                          // 冻结时间
	ReleaseAt        *time.Time `gorm:"index" json:"release_at"`                                         // 放款时间
	RefundedAt       *time.Time `gorm:"index" json:"refunded_at"`                                        // 退款时间
	AutoReleaseAt    *time.Time `gorm:"index" json:"auto_release_at"`                                    // 自动放款时间
	DisputeReason    string     `gorm:"size:500;not null;default:''" json:"dispute_reason"`             // 争议原因
	ArbitrationResult string    `gorm:"type:text" json:"arbitration_result"`                             // 仲裁结果
	CompletedAt      *time.Time `gorm:"index" json:"completed_at"`                                       // 完成时间
}

// TableName 表名（car_ 前缀）
func (CarEscrow) TableName() string { return "car_escrows" }

// === 推荐类型常量 ===
const (
	RecTypeCarToUser      = "car_to_user"      // 车荐人
	RecTypeUserToCar      = "user_to_car"      // 人荐车
	RecTypeSimilar        = "similar"          // 相似车源
	RecTypeNearby         = "nearby"           // 附近车源
	RecTypeRecentlyViewed = "recently_viewed"  // 看过又推荐
	RecTypeHot            = "hot"              // 热门推荐
)

// === 推荐来源常量 ===
const (
	RecSourceAI     = "ai"       // AI 推荐
	RecSourceManual = "manual"   // 人工推荐
	RecSourceHot    = "hot"      // 热门
	RecSourceNew    = "new"      // 新车
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

// CarRecommendation 推荐记录表
type CarRecommendation struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID          uint       `gorm:"not null;index;uniqueIndex:uniq_car_recs_user_car_type" json:"user_id"`                  // 用户 ID
	CarID           uint       `gorm:"not null;index;uniqueIndex:uniq_car_recs_user_car_type" json:"car_id"`                   // 车源 ID
	RecType         string     `gorm:"size:32;not null;default:'car_to_user';index;uniqueIndex:uniq_car_recs_user_car_type" json:"rec_type"` // car_to_user/user_to_car/similar/nearby/recently_viewed/hot
	Source          string     `gorm:"size:32;not null;default:'ai';index" json:"source"`                                     // ai/manual/hot/new
	Score           float64    `gorm:"type:decimal(5,2);default:0;index" json:"score"`                                        // 推荐评分
	Reason          string     `gorm:"size:500;not null;default:''" json:"reason"`                                            // 推荐理由
	PriceMatch      float64    `gorm:"type:decimal(5,2);default:0" json:"price_match"`                                        // 价格匹配度
	BrandMatch      float64    `gorm:"type:decimal(5,2);default:0" json:"brand_match"`                                        // 品牌匹配度
	TypeMatch       float64    `gorm:"type:decimal(5,2);default:0" json:"type_match"`                                         // 车型匹配度
	ConditionMatch  float64    `gorm:"type:decimal(5,2);default:0" json:"condition_match"`                                    // 车况匹配度
	Status          int        `gorm:"default:0;index" json:"status"`                                                          // 0待展示 1已展示 2点击 3联系 4忽略 5过期
	ClickedAt       *time.Time `gorm:"index" json:"clicked_at"`                                                                // 点击时间
	ContactedAt     *time.Time `gorm:"index" json:"contacted_at"`                                                              // 联系时间
	ViewedAt        *time.Time `gorm:"index" json:"viewed_at"`                                                                 // 查看时间
	DismissedAt     *time.Time `gorm:"index" json:"dismissed_at"`                                                              // 忽略时间
	ExpiredAt       *time.Time `gorm:"index" json:"expired_at"`                                                                // 过期时间
}

// TableName 表名（car_ 前缀）
func (CarRecommendation) TableName() string { return "car_recommendations" }
