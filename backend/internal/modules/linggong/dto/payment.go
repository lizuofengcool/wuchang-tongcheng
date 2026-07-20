// Package dto 同城零工兼职数据传输对象 - 薪资支付 + 纠纷 + 提现
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PaymentInfo 薪资支付详情响应
type PaymentInfo struct {
	ID              uint       `json:"id"`
	PaymentNo       string     `json:"payment_no"`
	LinggongID      uint       `json:"linggong_id"`
	TaskID          uint       `json:"task_id"`
	ApplicationID   uint       `json:"application_id"`
	ContractID      uint       `json:"contract_id"`
	EmployerID      uint       `json:"employer_id"`
	EmployerName    string     `json:"employer_name"`
	WorkerID        uint       `json:"worker_id"`
	WorkerName      string     `json:"worker_name"`
	WorkerPhone     string     `json:"worker_phone"`
	WorkerBankAccount string   `json:"worker_bank_account"`
	WorkerAlipay    string     `json:"worker_alipay"`
	WorkerWechat    string     `json:"worker_wechat"`
	PaymentType     string     `json:"payment_type"`
	PaymentTypeText string     `json:"payment_type_text"`
	Amount          float64    `json:"amount"`
	WorkHours       float64    `json:"work_hours"`
	WorkDays        int        `json:"work_days"`
	TaskCount       int        `json:"task_count"`
	UnitPrice       float64    `json:"unit_price"`
	Quantity        float64    `json:"quantity"`
	Settlement      string     `json:"settlement"`
	SettlementText  string     `json:"settlement_text"`
	SettlementStatus int       `json:"settlement_status"`
	SettlementStatusText string `json:"settlement_status_text"`
	SettlementAt    *time.Time `json:"settlement_at"`
	DueAt           *time.Time `json:"due_at"`
	PayMethod       string     `json:"pay_method"`
	PayMethodText   string     `json:"pay_method_text"`
	PayTradeNo      string     `json:"pay_trade_no"`
	PayChannel      string     `json:"pay_channel"`
	PayeeName       string     `json:"payee_name"`
	PlatformFee    float64    `json:"platform_fee"`
	TaxAmount       float64    `json:"tax_amount"`
	ActualAmount    float64    `json:"actual_amount"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	FailedReason    string     `json:"failed_reason"`
	PaidAt          *time.Time `json:"paid_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at"`
	CanceledAt      *time.Time `json:"canceled_at"`
	WorkStartDate   *time.Time `json:"work_start_date"`
	WorkEndDate     *time.Time `json:"work_end_date"`
	EvidenceImages   interface{} `json:"evidence_images"`
	Remark          string     `json:"remark"`
	InvoiceURL      string     `json:"invoice_url"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	LinggongID        uint    `json:"linggong_id" binding:"required"`
	TaskID            uint    `json:"task_id"`
	ApplicationID     uint    `json:"application_id"`
	ContractID        uint    `json:"contract_id"`
	EmployerID        uint    `json:"employer_id" binding:"required"`
	EmployerName      string  `json:"employer_name" binding:"max=128"`
	WorkerID          uint    `json:"worker_id" binding:"required"`
	WorkerName        string  `json:"worker_name" binding:"max=50"`
	WorkerPhone       string  `json:"worker_phone" binding:"max=20"`
	WorkerBankAccount string  `json:"worker_bank_account" binding:"max=64"`
	WorkerAlipay      string  `json:"worker_alipay" binding:"max=128"`
	WorkerWechat      string  `json:"worker_wechat" binding:"max=128"`
	PaymentType       string  `json:"payment_type" binding:"omitempty,oneof=salary bonus overtime allowance reimburse penalty refund deposit"`
	Amount            float64 `json:"amount" binding:"min=0"`
	WorkHours         float64 `json:"work_hours"`
	WorkDays          int     `json:"work_days"`
	TaskCount         int     `json:"task_count"`
	UnitPrice         float64 `json:"unit_price"`
	Quantity          float64 `json:"quantity"`
	Settlement        string  `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	PayMethod         string  `json:"pay_method" binding:"omitempty,oneof=wechat alipay bank cash balance escrow"`
	PayChannel        string  `json:"pay_channel" binding:"max=32"`
	PayeeName         string  `json:"payee_name" binding:"max=50"`
	WorkStartDate     *time.Time `json:"work_start_date"`
	WorkEndDate       *time.Time `json:"work_end_date"`
	EvidenceImages    interface{} `json:"evidence_images"`
	Remark            string  `json:"remark"`
	InvoiceURL        string  `json:"invoice_url" binding:"max=255"`
}

// UpdatePaymentRequest 更新支付请求
type UpdatePaymentRequest struct {
	Amount        *float64 `json:"amount" binding:"omitempty,min=0"`
	WorkHours     *float64 `json:"work_hours"`
	WorkDays      *int     `json:"work_days"`
	TaskCount     *int     `json:"task_count"`
	UnitPrice     *float64 `json:"unit_price"`
	Quantity      *float64 `json:"quantity"`
	Settlement    *string  `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	PayMethod     *string  `json:"pay_method" binding:"omitempty,oneof=wechat alipay bank cash balance escrow"`
	PayeeName     *string  `json:"payee_name" binding:"omitempty,max=50"`
	EvidenceImages interface{} `json:"evidence_images"`
	Remark        *string  `json:"remark"`
	InvoiceURL    *string  `json:"invoice_url" binding:"omitempty,max=255"`
}

// PaymentListRequest 支付列表请求
type PaymentListRequest struct {
	LinggongID      uint   `form:"linggong_id" json:"linggong_id"`
	TaskID          uint   `form:"task_id" json:"task_id"`
	ApplicationID   uint   `form:"application_id" json:"application_id"`
	EmployerID      uint   `form:"employer_id" json:"employer_id"`
	WorkerID        uint   `form:"worker_id" json:"worker_id"`
	PaymentType     string `form:"payment_type" json:"payment_type"`
	Settlement      string `form:"settlement" json:"settlement"`
	Status          *int   `form:"status" json:"status"`
	SettlementStatus *int  `form:"settlement_status" json:"settlement_status"`
	Keyword         string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// PaymentAdminListRequest 管理后台支付列表请求
type PaymentAdminListRequest struct {
	RegionID      uint   `form:"region_id" json:"region_id"`
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	EmployerID    uint   `form:"employer_id" json:"employer_id"`
	WorkerID      uint   `form:"worker_id" json:"worker_id"`
	PaymentType   string `form:"payment_type" json:"payment_type"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// PaymentStatusUpdateRequest 支付状态更新请求
type PaymentStatusUpdateRequest struct {
	Status        int    `json:"status" binding:"oneof=0 1 2 3 4 5 6 7"`
	FailedReason  string `json:"failed_reason" binding:"max=500"`
	PayTradeNo    string `json:"pay_trade_no" binding:"max=128"`
}

// PaymentSettleRequest 薪资结算请求
type PaymentSettleRequest struct {
	SettlementStatus int `json:"settlement_status" binding:"oneof=0 1 2 3"`
}

// DisputeInfo 纠纷信息
type DisputeInfo struct {
	ID                 uint       `json:"id"`
	DisputeNo          string     `json:"dispute_no"`
	LinggongID         uint       `json:"linggong_id"`
	TaskID             uint       `json:"task_id"`
	ApplicationID      uint       `json:"application_id"`
	ContractID         uint       `json:"contract_id"`
	PaymentID          uint       `json:"payment_id"`
	DisputeType        string     `json:"dispute_type"`
	DisputeTypeText    string     `json:"dispute_type_text"`
	ApplicantType      string     `json:"applicant_type"`
	ApplicantTypeText  string     `json:"applicant_type_text"`
	ApplicantID        uint       `json:"applicant_id"`
	ApplicantName      string     `json:"applicant_name"`
	RespondentID       uint       `json:"respondent_id"`
	RespondentName     string     `json:"respondent_name"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	EvidenceImages     interface{} `json:"evidence_images"`
	EvidenceVideos     interface{} `json:"evidence_videos"`
	EvidenceDocs       interface{} `json:"evidence_docs"`
	ClaimAmount        float64    `json:"claim_amount"`
	Status             int        `json:"status"`
	StatusText         string     `json:"status_text"`
	HandlerID          uint       `json:"handler_id"`
	HandlerName        string     `json:"handler_name"`
	MediationResult    string     `json:"mediation_result"`
	ArbitrationResult  string     `json:"arbitration_result"`
	FinalResult        string     `json:"final_result"`
	FinalResultText    string     `json:"final_result_text"`
	CompensationAmount float64    `json:"compensation_amount"`
	SLADeadline        *time.Time `json:"sla_deadline"`
	HandledAt          *time.Time `json:"handled_at"`
	ResolvedAt         *time.Time `json:"resolved_at"`
	ClosedAt           *time.Time `json:"closed_at"`
	AppealReason       string     `json:"appeal_reason"`
	AppealedAt         *time.Time `json:"appealed_at"`
	AppealResult       string     `json:"appeal_result"`
	AppealHandlerID    uint       `json:"appeal_handler_id"`
	AppealHandledAt    *time.Time `json:"appeal_handled_at"`
	RegionID           uint       `json:"region_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateDisputeRequest 创建纠纷请求
type CreateDisputeRequest struct {
	LinggongID      uint    `json:"linggong_id" binding:"required"`
	TaskID          uint    `json:"task_id"`
	ApplicationID   uint    `json:"application_id"`
	ContractID      uint    `json:"contract_id"`
	PaymentID       uint    `json:"payment_id"`
	DisputeType     string  `json:"dispute_type" binding:"omitempty,oneof=salary quality attendance breach discrimination harassment fraud safety other"`
	ApplicantType   string  `json:"applicant_type" binding:"omitempty,oneof=worker employer platform"`
	ApplicantID     uint    `json:"applicant_id" binding:"required"`
	ApplicantName   string  `json:"applicant_name" binding:"max=50"`
	RespondentID    uint    `json:"respondent_id" binding:"required"`
	RespondentName  string  `json:"respondent_name" binding:"max=50"`
	Title           string  `json:"title" binding:"required,max=200"`
	Description     string  `json:"description"`
	EvidenceImages  interface{} `json:"evidence_images"`
	EvidenceVideos  interface{} `json:"evidence_videos"`
	EvidenceDocs    interface{} `json:"evidence_docs"`
	ClaimAmount     float64 `json:"claim_amount" binding:"min=0"`
}

// DisputeHandleRequest 纠纷处理请求
type DisputeHandleRequest struct {
	Status             int     `json:"status" binding:"oneof=1 2 3 4 5 6 7"`
	MediationResult    string  `json:"mediation_result"`
	ArbitrationResult  string  `json:"arbitration_result"`
	FinalResult        string  `json:"final_result" binding:"omitempty,oneof=worker_win employer_win compromise platform_decision reject"`
	CompensationAmount float64 `json:"compensation_amount" binding:"min=0"`
}

// DisputeListRequest 纠纷列表请求
type DisputeListRequest struct {
	LinggongID   uint   `form:"linggong_id" json:"linggong_id"`
	DisputeType  string `form:"dispute_type" json:"dispute_type"`
	ApplicantID  uint   `form:"applicant_id" json:"applicant_id"`
	RespondentID uint   `form:"respondent_id" json:"respondent_id"`
	Status       *int   `form:"status" json:"status"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// WithdrawalInfo 提现信息
type WithdrawalInfo struct {
	ID                uint       `json:"id"`
	WithdrawalNo      string     `json:"withdrawal_no"`
	UserID            uint       `json:"user_id"`
	UserType          string     `json:"user_type"`
	UserTypeText      string     `json:"user_type_text"`
	UserName          string     `json:"user_name"`
	UserPhone         string     `json:"user_phone"`
	Amount            float64    `json:"amount"`
	Fee               float64    `json:"fee"`
	Tax               float64    `json:"tax"`
	ActualAmount      float64    `json:"actual_amount"`
	BalanceBefore     float64    `json:"balance_before"`
	BalanceAfter      float64    `json:"balance_after"`
	Method            string     `json:"method"`
	MethodText        string     `json:"method_text"`
	PayeeName         string     `json:"payee_name"`
	PayeeAccount      string     `json:"payee_account"`
	PayeeBank         string     `json:"payee_bank"`
	PayeeBankBranch   string     `json:"payee_bank_branch"`
	BankCardNo        string     `json:"bank_card_no"`
	AlipayAccount     string     `json:"alipay_account"`
	WechatAccount     string     `json:"wechat_account"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	FailedReason      string     `json:"failed_reason"`
	ReviewedBy        uint       `json:"reviewed_by"`
	ReviewedByName    string     `json:"reviewed_by_name"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewedRemark    string     `json:"reviewed_remark"`
	PayTradeNo        string     `json:"pay_trade_no"`
	PayChannel        string     `json:"pay_channel"`
	ProcessedAt       *time.Time `json:"processed_at"`
	SucceededAt       *time.Time `json:"succeeded_at"`
	CanceledAt        *time.Time `json:"canceled_at"`
	EstimatedArrival  *time.Time `json:"estimated_arrival"`
	Remark            string     `json:"remark"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateWithdrawalRequest 创建提现请求
type CreateWithdrawalRequest struct {
	UserType       string  `json:"user_type" binding:"omitempty,oneof=worker employer"`
	Amount         float64 `json:"amount" binding:"required,min=0.01"`
	Method         string  `json:"method" binding:"omitempty,oneof=wechat alipay bank balance"`
	PayeeName      string  `json:"payee_name" binding:"max=50"`
	PayeeAccount   string  `json:"payee_account" binding:"max=128"`
	PayeeBank       string  `json:"payee_bank" binding:"max=64"`
	PayeeBankBranch string `json:"payee_bank_branch" binding:"max=128"`
	BankCardNo     string  `json:"bank_card_no" binding:"max=64"`
	AlipayAccount  string  `json:"alipay_account" binding:"max=128"`
	WechatAccount  string  `json:"wechat_account" binding:"max=128"`
	Remark         string  `json:"remark"`
}

// WithdrawalListRequest 提现列表请求
type WithdrawalListRequest struct {
	UserID    uint   `form:"user_id" json:"user_id"`
	UserType  string `form:"user_type" json:"user_type"`
	Status    *int   `form:"status" json:"status"`
	Method    string `form:"method" json:"method"`
	Keyword   string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// WithdrawalAuditRequest 提现审核请求
type WithdrawalAuditRequest struct {
	Status     int    `json:"status" binding:"oneof=1 5"`
	FailedReason string `json:"failed_reason" binding:"max=500"`
	ReviewedRemark string `json:"reviewed_remark" binding:"max=500"`
}
