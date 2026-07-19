// Package model 支付财务中台精简版数据模型
// 依据 ershou 模块依赖：担保交易 + 订单 + 退款 + 提现 + 结算 + 资金账户
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 7.1：通用字段 id/region_id/created_at/updated_at/deleted_at
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 支付方式常量 ===
const (
	PayMethodWechat   = "wechat"   // 微信支付
	PayMethodAlipay   = "alipay"   // 支付宝
	PayMethodBalance  = "balance"  // 余额
	PayMethodPoint    = "point"    // 积分
	PayMethodGiftCard = "giftcard" // 礼品卡
)

// === 订单支付状态常量 ===
const (
	PayStatusPending   = 0 // 待支付
	PayStatusPaid      = 1 // 已支付
	PayStatusClosed    = 2 // 已关闭
	PayStatusRefunded  = 3 // 已退款
	PayStatusPartRefund = 4 // 部分退款
)

// === 担保状态常量 ===
const (
	EscrowStatusFrozen     = 0 // 冻结中
	EscrowStatusReleased   = 1 // 已放款
	EscrowStatusRefunded   = 2 // 已退款
	EscrowStatusPartRefund = 3 // 已部分退款
)

// === 退款状态常量 ===
const (
	RefundStatusPending  = 0 // 待处理
	RefundStatusRefunded = 1 // 已退款
	RefundStatusRejected = 2 // 已拒绝
	RefundStatusProcessing = 3 // 处理中
)

// === 提现状态常量 ===
const (
	WithdrawStatusPending  = 0 // 待审核
	WithdrawStatusApproved = 1 // 已通过
	WithdrawStatusRejected = 2 // 已拒绝
	WithdrawStatusPaid     = 3 // 已打款
	WithdrawStatusFailed   = 4 // 打款失败
)

// === 结算状态常量 ===
const (
	SettleStatusPending = 0 // 待结算
	SettleStatusDone    = 1 // 已结算
	SettleStatusFailed  = 2 // 已失败
)

// === 结算周期常量 ===
const (
	PeriodT1      = "T1"
	PeriodT7      = "T7"
	PeriodMonthly = "monthly"
)

// PaymentOrder 支付订单
type PaymentOrder struct {
	database.RegionBaseModel

	OrderNo      string    `gorm:"size:64;not null;uniqueIndex" json:"order_no"`        // 订单号
	UserID       uint      `gorm:"index;not null" json:"user_id"`                       // 用户ID
	BizModule    string    `gorm:"size:32" json:"biz_module"`                           // 业务模块名
	BizID        string    `gorm:"size:128" json:"biz_id"`                              // 业务订单ID
	Title        string    `gorm:"size:256" json:"title"`                               // 订单标题
	Amount       float64   `gorm:"type:decimal(12,2);default:0" json:"amount"`          // 金额
	PayMethod    string    `gorm:"size:32" json:"pay_method"`                           // 支付方式
	PayStatus    int       `gorm:"default:0;index" json:"pay_status"`                   // 支付状态
	ThirdPartyNo string    `gorm:"size:128" json:"third_party_no"`                      // 第三方流水号
	PaidAt       *time.Time `gorm:"index" json:"paid_at"`                               // 支付时间
	ExpireAt     *time.Time `gorm:"index" json:"expire_at"`                             // 过期时间
	Extra        string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`         // 扩展字段 JSON
}

// TableName 表名
func (PaymentOrder) TableName() string { return "pay_orders" }

// EscrowAccount 担保交易（资金托管→确认收货→放款）
type EscrowAccount struct {
	database.RegionBaseModel

	OrderID        uint       `gorm:"index;not null" json:"order_id"`               // 订单ID
	OrderNo        string     `gorm:"size:64" json:"order_no"`                      // 订单号
	UserID         uint       `gorm:"index;not null" json:"user_id"`                // 买家ID
	MerchantID     uint       `gorm:"index" json:"merchant_id"`                     // 卖家/商家ID
	Amount         float64    `gorm:"type:decimal(12,2);default:0" json:"amount"`   // 托管金额
	Status         int        `gorm:"default:0;index" json:"status"`                // 状态
	FrozenAt       time.Time  `gorm:"not null;default:now()" json:"frozen_at"`      // 冻结时间
	ReleaseAt      *time.Time `gorm:"index" json:"release_at"`                      // 放款时间
	AutoReleaseAt  *time.Time `gorm:"index" json:"auto_release_at"`                 // 自动放款时间
}

// TableName 表名
func (EscrowAccount) TableName() string { return "pay_escrows" }

// RefundOrder 退款单
type RefundOrder struct {
	database.RegionBaseModel

	RefundNo           string     `gorm:"size:64;not null;uniqueIndex" json:"refund_no"`        // 退款单号
	OrderID            uint       `gorm:"index;not null" json:"order_id"`                       // 订单ID
	OrderNo            string     `gorm:"size:64" json:"order_no"`                              // 订单号
	UserID             uint       `gorm:"index;not null" json:"user_id"`                        // 用户ID
	Amount             float64    `gorm:"type:decimal(12,2);default:0" json:"amount"`           // 退款金额
	Reason             string     `gorm:"size:256" json:"reason"`                               // 退款原因
	Status             int        `gorm:"default:0;index" json:"status"`                        // 状态
	ThirdPartyRefundNo string     `gorm:"size:128" json:"third_party_refund_no"`                // 第三方退款号
	RefundMethod       string     `gorm:"size:32" json:"refund_method"`                         // 退款方式
	ProcessedAt        *time.Time `gorm:"index" json:"processed_at"`                            // 处理时间
}

// TableName 表名
func (RefundOrder) TableName() string { return "pay_refunds" }

// Withdrawal 提现单
type Withdrawal struct {
	database.RegionBaseModel

	WithdrawalNo  string     `gorm:"size:64;not null;uniqueIndex" json:"withdrawal_no"` // 提现单号
	UserID        uint       `gorm:"index;not null" json:"user_id"`                     // 用户ID
	Amount        float64    `gorm:"type:decimal(12,2);default:0" json:"amount"`        // 提现金额
	Fee           float64    `gorm:"type:decimal(12,2);default:0" json:"fee"`           // 手续费
	ActualAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"actual_amount"` // 实际到账
	BankCard      string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"bank_card"`    // 银行卡信息 JSON
	Status        int        `gorm:"default:0;index" json:"status"`                     // 状态
	RejectReason  string     `gorm:"size:256" json:"reject_reason"`                     // 拒绝原因
	ProcessedAt   *time.Time `gorm:"index" json:"processed_at"`                         // 处理时间
}

// TableName 表名
func (Withdrawal) TableName() string { return "pay_withdrawals" }

// Settlement 结算单
type Settlement struct {
	database.RegionBaseModel

	SettlementNo      string     `gorm:"size:64;not null;uniqueIndex" json:"settlement_no"` // 结算单号
	MerchantID        uint       `gorm:"index;not null" json:"merchant_id"`                 // 商家ID
	PeriodType        string     `gorm:"size:16;default:'T1'" json:"period_type"`           // 结算周期
	PeriodStart       time.Time  `gorm:"not null" json:"period_start"`                      // 周期开始
	PeriodEnd         time.Time  `gorm:"not null" json:"period_end"`                        // 周期结束
	OrderCount        int        `gorm:"default:0" json:"order_count"`                      // 订单数
	TotalAmount       float64    `gorm:"type:decimal(12,2);default:0" json:"total_amount"`  // 总金额
	Commission        float64    `gorm:"type:decimal(12,2);default:0" json:"commission"`    // 平台抽成
	SettlementAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"settlement_amount"` // 到账金额
	Status            int        `gorm:"default:0;index" json:"status"`                     // 状态
	SettledAt         *time.Time `gorm:"index" json:"settled_at"`                           // 结算时间
}

// TableName 表名
func (Settlement) TableName() string { return "pay_settlements" }

// Account 用户资金账户
type Account struct {
	database.RegionBaseModel

	UserID         uint    `gorm:"index;not null;uniqueIndex" json:"user_id"`             // 用户ID（唯一）
	Balance        float64 `gorm:"type:decimal(12,2);default:0" json:"balance"`           // 可用余额
	FrozenAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"frozen_amount"`     // 冻结金额
	TotalIncome    float64 `gorm:"type:decimal(12,2);default:0" json:"total_income"`      // 累计收入
	TotalExpense   float64 `gorm:"type:decimal(12,2);default:0" json:"total_expense"`     // 累计支出
	BankCards      string  `gorm:"type:jsonb;default:'[]'::jsonb" json:"bank_cards"`      // 银行卡列表 JSON
}

// TableName 表名
func (Account) TableName() string { return "pay_accounts" }
