// Package dto 支付财务中台精简版数据传输对象
package dto

import "time"

// PaymentOrderInfo 支付订单信息
type PaymentOrderInfo struct {
	ID            uint       `json:"id"`
	OrderNo       string     `json:"order_no"`
	UserID        uint       `json:"user_id"`
	BizModule     string     `json:"biz_module"`
	BizID         string     `json:"biz_id"`
	Title         string     `json:"title"`
	Amount        float64    `json:"amount"`
	PayMethod     string     `json:"pay_method"`
	PayStatus     int        `json:"pay_status"`
	ThirdPartyNo  string     `json:"third_party_no"`
	PaidAt        *time.Time `json:"paid_at"`
	ExpireAt      *time.Time `json:"expire_at"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreatePaymentRequest 创建支付订单请求
type CreatePaymentRequest struct {
	BizModule string  `json:"biz_module" binding:"required,max=32"`
	BizID     string  `json:"biz_id" binding:"required,max=128"`
	Title     string  `json:"title" binding:"required,max=256"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	PayMethod string  `json:"pay_method" binding:"required,oneof=wechat alipay balance point giftcard"`
	ExpireSec int     `json:"expire_sec"` // 订单过期秒数（默认 1800）
	Extra     string  `json:"extra"`      // JSON 字符串
}

// ConfirmPaymentRequest 确认支付请求（第三方回调或主动确认）
type ConfirmPaymentRequest struct {
	OrderNo       string `json:"order_no" binding:"required"`
	ThirdPartyNo  string `json:"third_party_no" binding:"required"`
	PayMethod     string `json:"pay_method" binding:"required"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderNo string  `json:"order_no" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
	Reason  string  `json:"reason" binding:"required,max=256"`
}

// ConfirmEscrowRequest 确认收货（放款担保账户）
type ConfirmEscrowRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
}

// WithdrawRequest 提现申请请求
type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	BankCardNo  string  `json:"bank_card_no" binding:"required,max=32"`
	BankName    string  `json:"bank_name" binding:"required,max=64"`
	HolderName  string  `json:"holder_name" binding:"required,max=32"`
	Phone       string  `json:"phone" binding:"max=20"`
}

// WithdrawActionRequest 提现审批请求（M 端）
type WithdrawActionRequest struct {
	WithdrawalNo string `json:"withdrawal_no" binding:"required"`
	Action       string `json:"action" binding:"required,oneof=approve reject paid failed"`
	Reason       string `json:"reason" binding:"max=256"`
}

// SettleRequest 手动触发结算请求（M 端）
type SettleRequest struct {
	MerchantID  uint   `json:"merchant_id" binding:"required"`
	PeriodType  string `json:"period_type" binding:"omitempty,oneof=T1 T7 monthly"`
}

// AccountInfo 资金账户信息
type AccountInfo struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	Balance       float64 `json:"balance"`
	FrozenAmount  float64 `json:"frozen_amount"`
	TotalIncome   float64 `json:"total_income"`
	TotalExpense  float64 `json:"total_expense"`
	BankCards     string  `json:"bank_cards"`
}

// EscrowInfo 担保账户信息
type EscrowInfo struct {
	ID            uint       `json:"id"`
	OrderID       uint       `json:"order_id"`
	OrderNo       string     `json:"order_no"`
	UserID        uint       `json:"user_id"`
	MerchantID    uint       `json:"merchant_id"`
	Amount        float64    `json:"amount"`
	Status        int        `json:"status"`
	FrozenAt      time.Time  `json:"frozen_at"`
	ReleaseAt     *time.Time `json:"release_at"`
	AutoReleaseAt *time.Time `json:"auto_release_at"`
}

// RefundInfo 退款单信息
type RefundInfo struct {
	ID                  uint       `json:"id"`
	RefundNo            string     `json:"refund_no"`
	OrderID             uint       `json:"order_id"`
	OrderNo             string     `json:"order_no"`
	UserID              uint       `json:"user_id"`
	Amount              float64    `json:"amount"`
	Reason              string     `json:"reason"`
	Status              int        `json:"status"`
	ThirdPartyRefundNo  string     `json:"third_party_refund_no"`
	RefundMethod        string     `json:"refund_method"`
	ProcessedAt         *time.Time `json:"processed_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

// WithdrawInfo 提现单信息
type WithdrawInfo struct {
	ID            uint       `json:"id"`
	WithdrawalNo  string     `json:"withdrawal_no"`
	UserID        uint       `json:"user_id"`
	Amount        float64    `json:"amount"`
	Fee           float64    `json:"fee"`
	ActualAmount  float64    `json:"actual_amount"`
	BankCard      string     `json:"bank_card"`
	Status        int        `json:"status"`
	RejectReason  string     `json:"reject_reason"`
	ProcessedAt   *time.Time `json:"processed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// SettlementInfo 结算单信息
type SettlementInfo struct {
	ID                uint       `json:"id"`
	SettlementNo      string     `json:"settlement_no"`
	MerchantID        uint       `json:"merchant_id"`
	PeriodType        string     `json:"period_type"`
	PeriodStart       time.Time  `json:"period_start"`
	PeriodEnd         time.Time  `json:"period_end"`
	OrderCount        int        `json:"order_count"`
	TotalAmount       float64    `json:"total_amount"`
	Commission        float64    `json:"commission"`
	SettlementAmount  float64    `json:"settlement_amount"`
	Status            int        `json:"status"`
	SettledAt         *time.Time `json:"settled_at"`
	CreatedAt         time.Time  `json:"created_at"`
}
