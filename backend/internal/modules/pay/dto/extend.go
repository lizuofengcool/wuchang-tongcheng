// Package dto 支付中台扩展数据传输对象
// 依据 012_pay_full.sql 扩展：交易流水/渠道/商户/回调/担保争议
package dto

import "time"

// TransactionInfo 交易流水信息
type TransactionInfo struct {
	ID             uint       `json:"id"`
	TransactionNo  string     `json:"transaction_no"`
	OrderID        uint       `json:"order_id"`
	OrderNo        string     `json:"order_no"`
	UserID         uint       `json:"user_id"`
	Channel        string     `json:"channel"`
	ThirdPartyNo   string     `json:"third_party_no"`
	Amount         float64    `json:"amount"`
	Fee            float64    `json:"fee"`
	Status         int        `json:"status"`
	ChannelResp    string     `json:"channel_resp"`
	PaidAt         *time.Time `json:"paid_at"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

// TransactionListRequest 交易流水列表查询
type TransactionListRequest struct {
	UserID   uint   `form:"user_id" json:"user_id"`
	OrderNo  string `form:"order_no" json:"order_no"`
	Channel  string `form:"channel" json:"channel"`
	Status   int    `form:"status" json:"status"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

// ChannelInfo 支付渠道信息
type ChannelInfo struct {
	ID          uint    `json:"id"`
	ChannelCode string  `json:"channel_code"`
	ChannelName string  `json:"channel_name"`
	MerchantNo  string  `json:"merchant_no"`
	AppID       string  `json:"app_id"`
	CallbackURL string  `json:"callback_url"`
	NotifyURL   string  `json:"notify_url"`
	FeeRate     float64 `json:"fee_rate"`
	FeeFixed    float64 `json:"fee_fixed"`
	Status      int     `json:"status"`
	Sort        int     `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateChannelRequest 创建支付渠道请求
type CreateChannelRequest struct {
	ChannelCode string  `json:"channel_code" binding:"required,oneof=wechat alipay unionpay stripe"`
	ChannelName string  `json:"channel_name" binding:"required,max=64"`
	MerchantNo  string  `json:"merchant_no" binding:"max=64"`
	AppID       string  `json:"app_id" binding:"max=128"`
	SecretKey   string  `json:"secret_key" binding:"max=512"`
	PublicKey   string  `json:"public_key"`
	PrivateKey  string  `json:"private_key"`
	CallbackURL string  `json:"callback_url" binding:"max=256"`
	NotifyURL   string  `json:"notify_url" binding:"max=256"`
	FeeRate     float64 `json:"fee_rate"`
	FeeFixed    float64 `json:"fee_fixed"`
	Config      string  `json:"config"`
	Sort        int     `json:"sort"`
	Status      int     `json:"status"`
}

// UpdateChannelRequest 更新支付渠道请求
type UpdateChannelRequest struct {
	ChannelName string  `json:"channel_name" binding:"max=64"`
	MerchantNo  string  `json:"merchant_no" binding:"max=64"`
	AppID       string  `json:"app_id" binding:"max=128"`
	SecretKey   string  `json:"secret_key" binding:"max=512"`
	PublicKey   string  `json:"public_key"`
	PrivateKey  string  `json:"private_key"`
	CallbackURL string  `json:"callback_url" binding:"max=256"`
	NotifyURL   string  `json:"notify_url" binding:"max=256"`
	FeeRate     float64 `json:"fee_rate"`
	FeeFixed    float64 `json:"fee_fixed"`
	Config      string  `json:"config"`
	Sort        int     `json:"sort"`
	Status      int     `json:"status"`
}

// MerchantInfo 商户信息
type MerchantInfo struct {
	ID              uint      `json:"id"`
	MerchantNo      string    `json:"merchant_no"`
	MerchantName    string    `json:"merchant_name"`
	UserID          uint      `json:"user_id"`
	ContactName     string    `json:"contact_name"`
	ContactPhone    string    `json:"contact_phone"`
	FeeRate         float64   `json:"fee_rate"`
	SettlementCycle string    `json:"settlement_cycle"`
	BusinessLicense string    `json:"business_license"`
	BusinessScope   string    `json:"business_scope"`
	BankAccount     string    `json:"bank_account"`
	Status          int       `json:"status"`
	AuditRemark     string    `json:"audit_remark"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateMerchantRequest 创建商户请求
type CreateMerchantRequest struct {
	MerchantNo      string  `json:"merchant_no" binding:"required,max=64"`
	MerchantName    string  `json:"merchant_name" binding:"required,max=128"`
	UserID          uint    `json:"user_id"`
	ContactName     string  `json:"contact_name" binding:"max=64"`
	ContactPhone    string  `json:"contact_phone" binding:"max=32"`
	FeeRate         float64 `json:"fee_rate"`
	SettlementCycle string  `json:"settlement_cycle" binding:"omitempty,oneof=T1 T7 monthly"`
	BusinessLicense string  `json:"business_license" binding:"max=64"`
	BusinessScope   string  `json:"business_scope"`
	BankAccount     string  `json:"bank_account"`
}

// AuditMerchantRequest 商户审核请求
type AuditMerchantRequest struct {
	ID          uint   `json:"id" binding:"required"`
	Status      int    `json:"status" binding:"required,oneof=0 1 2 3"`
	AuditRemark string `json:"audit_remark" binding:"max=256"`
}

// CallbackInfo 回调通知信息
type CallbackInfo struct {
	ID            uint       `json:"id"`
	OrderNo       string     `json:"order_no"`
	Channel       string     `json:"channel"`
	ThirdPartyNo  string     `json:"third_party_no"`
	NotifyType    string     `json:"notify_type"`
	RawData       string     `json:"raw_data"`
	ParsedData    string     `json:"parsed_data"`
	Signature     string     `json:"signature"`
	Status        int        `json:"status"`
	ProcessCount  int        `json:"process_count"`
	ErrorMsg      string     `json:"error_msg"`
	ProcessedAt   *time.Time `json:"processed_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// RecordCallbackRequest 记录三方回调请求
type RecordCallbackRequest struct {
	OrderNo      string `json:"order_no" binding:"required"`
	Channel      string `json:"channel" binding:"required"`
	ThirdPartyNo string `json:"third_party_no"`
	NotifyType   string `json:"notify_type" binding:"required,oneof=pay refund escrow"`
	RawData      string `json:"raw_data"`
	ParsedData   string `json:"parsed_data"`
	Signature    string `json:"signature"`
}

// EscrowDisputeRequest 担保争议请求
type EscrowDisputeRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
	Reason  string `json:"reason" binding:"required,max=512"`
}

// EscrowArbitrateRequest 担保仲裁请求（M 端）
type EscrowArbitrateRequest struct {
	OrderNo   string `json:"order_no" binding:"required"`
	Result    string `json:"result" binding:"required,oneof=release refund split"`
	Remark    string `json:"remark" binding:"max=512"`
	BuyerRatio float64 `json:"buyer_ratio"` // 买家分账比例（split时）
}

// PayStatisticsResponse 支付统计响应
type PayStatisticsResponse struct {
	TotalOrders       int64   `json:"total_orders"`
	TotalAmount       float64 `json:"total_amount"`
	RefundAmount      float64 `json:"refund_amount"`
	EscrowAmount      float64 `json:"escrow_amount"`
	SuccessRate       float64 `json:"success_rate"`        // 成功率
	RefundRate        float64 `json:"refund_rate"`
	TodayCount        int64   `json:"today_count"`
	TodayAmount       float64 `json:"today_amount"`
	ChannelStats      []ChannelStatItem `json:"channel_stats"`
}

// ChannelStatItem 渠道统计项
type ChannelStatItem struct {
	Channel  string  `json:"channel"`
	Count    int64   `json:"count"`
	Amount   float64 `json:"amount"`
}
