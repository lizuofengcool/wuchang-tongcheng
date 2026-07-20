// Package model 提现表（对接支付中台）
// 求职者提现 + 雇主提现 + 提现状态机
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 提现状态常量 ===
const (
	WithdrawalStatusPending   = 0 // 待审核
	WithdrawalStatusApproved  = 1 // 已审核
	WithdrawalStatusProcessing = 2 // 处理中
	WithdrawalStatusSucceeded = 3 // 已到账
	WithdrawalStatusFailed    = 4 // 失败
	WithdrawalStatusRejected  = 5 // 已驳回
	WithdrawalStatusCanceled  = 6 // 已取消
)

// === 提现方式常量 ===
const (
	WithdrawalMethodWechat   = "wechat"   // 微信
	WithdrawalMethodAlipay   = "alipay"   // 支付宝
	WithdrawalMethodBank     = "bank"     // 银行卡
	WithdrawalMethodBalance  = "balance" // 余额
)

// === 用户类型常量 ===
const (
	WithdrawalUserTypeWorker   = "worker"   // 求职者提现
	WithdrawalUserTypeEmployer = "employer" // 雇主提现
)

// LinggongWithdrawal 提现表
type LinggongWithdrawal struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	WithdrawalNo    string     `gorm:"size:64;not null;uniqueIndex" json:"withdrawal_no"`          // 提现单号
	UserID          uint       `gorm:"not null;index" json:"user_id"`                              // 用户 ID
	UserType        string     `gorm:"size:16;not null;default:'worker';index" json:"user_type"`   // worker/employer
	UserName        string     `gorm:"size:50;not null;default:''" json:"user_name"`               // 用户姓名
	UserPhone       string     `gorm:"size:20;not null;default:''" json:"user_phone"`              // 用户手机
	Amount          float64    `gorm:"type:decimal(12,2);default:0;index" json:"amount"`            // 提现金额
	Fee             float64    `gorm:"type:decimal(12,2);default:0" json:"fee"`                      // 手续费
	Tax             float64    `gorm:"type:decimal(12,2);default:0" json:"tax"`                     // 税费
	ActualAmount    float64    `gorm:"type:decimal(12,2);default:0" json:"actual_amount"`           // 实际到账
	BalanceBefore   float64    `gorm:"type:decimal(12,2);default:0" json:"balance_before"`          // 提现前余额
	BalanceAfter    float64    `gorm:"type:decimal(12,2);default:0" json:"balance_after"`            // 提现后余额
	Method          string     `gorm:"size:16;not null;default:'wechat';index" json:"method"`      // wechat/alipay/bank/balance
	PayeeName       string     `gorm:"size:50;not null;default:''" json:"payee_name"`               // 收款人姓名
	PayeeAccount    string     `gorm:"size:128;not null;default:''" json:"payee_account"`           // 收款账号
	PayeeBank       string     `gorm:"size:64;not null;default:''" json:"payee_bank"`               // 收款银行
	PayeeBankBranch string     `gorm:"size:128;not null;default:''" json:"payee_bank_branch"`       // 收款支行
	BankCardNo      string     `gorm:"size:64;not null;default:''" json:"bank_card_no"`             // 银行卡号
	AlipayAccount   string     `gorm:"size:128;not null;default:''" json:"alipay_account"`          // 支付宝账号
	WechatAccount   string     `gorm:"size:128;not null;default:''" json:"wechat_account"`          // 微信账号
	Status          int        `gorm:"default:0;index" json:"status"`                              // 0待审 1已审 2处理中 3已到账 4失败 5驳回 6取消
	FailedReason    string     `gorm:"size:500;not null;default:''" json:"failed_reason"`          // 失败原因
	ReviewedBy      uint       `gorm:"not null;default:0" json:"reviewed_by"`                       // 审核人 ID
	ReviewedByName  string     `gorm:"size:50;not null;default:''" json:"reviewed_by_name"`         // 审核人姓名
	ReviewedAt      *time.Time `gorm:"index" json:"reviewed_at"`                                    // 审核时间
	ReviewedRemark  string     `gorm:"size:500;not null;default:''" json:"reviewed_remark"`         // 审核备注
	PayTradeNo      string     `gorm:"size:128;not null;default:'';index" json:"pay_trade_no"`       // 第三方支付单号
	PayChannel      string     `gorm:"size:32;not null;default:''" json:"pay_channel"`               // 支付渠道
	ProcessedAt     *time.Time `gorm:"index" json:"processed_at"`                                  // 处理时间
	SucceededAt     *time.Time `gorm:"index" json:"succeeded_at"`                                  // 到账时间
	CanceledAt      *time.Time `gorm:"index" json:"canceled_at"`                                   // 取消时间
	EstimatedArrival *time.Time `gorm:"index" json:"estimated_arrival"`                              // 预计到账时间
	Remark          string     `gorm:"type:text" json:"remark"`                                    // 备注
}

// TableName 表名（linggong_ 前缀）
func (LinggongWithdrawal) TableName() string { return "linggong_withdrawals" }
