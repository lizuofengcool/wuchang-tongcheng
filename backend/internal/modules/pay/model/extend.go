// Package model 支付中台扩展数据模型
// 依据 012_pay_full.sql：交易流水/渠道/商户/回调
package model

import (
	"time"
)

// Transaction 交易流水记录
// 对标 Stripe charge / 支付宝 trade_record
type Transaction struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	RegionID       uint       `gorm:"index;not null;default:1" json:"region_id"`
	TransactionNo  string     `gorm:"size:64;not null;uniqueIndex" json:"transaction_no"`
	OrderID        uint       `gorm:"index;not null;default:0" json:"order_id"`
	OrderNo        string     `gorm:"size:64" json:"order_no"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	Channel        string     `gorm:"size:32" json:"channel"`
	ThirdPartyNo   string     `gorm:"size:128" json:"third_party_no"`
	Amount         float64    `gorm:"type:decimal(12,2);default:0" json:"amount"`
	Fee            float64    `gorm:"type:decimal(12,2);default:0" json:"fee"`
	Status         int        `gorm:"default:0;index" json:"status"`
	ChannelResp    string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"channel_resp"`
	PaidAt         *time.Time `gorm:"index" json:"paid_at"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Transaction) TableName() string { return "pay_transactions" }

// Channel 支付渠道配置
type Channel struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RegionID     uint      `gorm:"index;not null;default:1" json:"region_id"`
	ChannelCode  string    `gorm:"size:32" json:"channel_code"`
	ChannelName  string    `gorm:"size:64" json:"channel_name"`
	MerchantNo   string    `gorm:"size:64" json:"merchant_no"`
	AppID        string    `gorm:"size:128" json:"app_id"`
	SecretKey    string    `gorm:"size:512" json:"secret_key"`
	PublicKey    string    `gorm:"type:text" json:"public_key"`
	PrivateKey   string    `gorm:"type:text" json:"private_key"`
	CallbackURL  string    `gorm:"size:256" json:"callback_url"`
	NotifyURL    string    `gorm:"size:256" json:"notify_url"`
	FeeRate      float64   `gorm:"type:decimal(6,4);default:0" json:"fee_rate"`
	FeeFixed     float64   `gorm:"type:decimal(12,2);default:0" json:"fee_fixed"`
	Config       string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"config"`
	Sort         int       `gorm:"default:0" json:"sort"`
	Status       int       `gorm:"default:1;index" json:"status"`
	CreatedAt    time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Channel) TableName() string { return "pay_channels" }

// Merchant 商户配置
type Merchant struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	RegionID         uint      `gorm:"index;not null;default:1" json:"region_id"`
	MerchantNo       string    `gorm:"size:64;not null;uniqueIndex" json:"merchant_no"`
	MerchantName     string    `gorm:"size:128" json:"merchant_name"`
	UserID           uint      `gorm:"index" json:"user_id"`
	ContactName      string    `gorm:"size:64" json:"contact_name"`
	ContactPhone     string    `gorm:"size:32" json:"contact_phone"`
	FeeRate          float64   `gorm:"type:decimal(6,4);default:0" json:"fee_rate"`
	SettlementCycle  string    `gorm:"size:16;default:'T1'" json:"settlement_cycle"`
	BusinessLicense  string    `gorm:"size:64" json:"business_license"`
	BusinessScope    string    `gorm:"type:text" json:"business_scope"`
	BankAccount      string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"bank_account"`
	Status           int       `gorm:"default:0;index" json:"status"`
	AuditRemark      string    `gorm:"size:256" json:"audit_remark"`
	CreatedAt        time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (Merchant) TableName() string { return "pay_merchants" }

// Callback 三方回调通知记录
type Callback struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	RegionID      uint       `gorm:"index;not null;default:1" json:"region_id"`
	OrderNo       string     `gorm:"size:64" json:"order_no"`
	Channel       string     `gorm:"size:32" json:"channel"`
	ThirdPartyNo  string     `gorm:"size:128" json:"third_party_no"`
	NotifyType    string     `gorm:"size:32" json:"notify_type"`
	RawData       string     `gorm:"type:text" json:"raw_data"`
	ParsedData    string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"parsed_data"`
	Signature     string     `gorm:"size:512" json:"signature"`
	Status        int        `gorm:"default:0;index" json:"status"`
	ProcessCount  int        `gorm:"default:0" json:"process_count"`
	ErrorMsg      string     `gorm:"type:text" json:"error_msg"`
	ProcessedAt   *time.Time `gorm:"index" json:"processed_at"`
	CreatedAt     time.Time  `gorm:"not null;default:now()" json:"created_at"`
}

// TableName 表名
func (Callback) TableName() string { return "pay_callbacks" }
