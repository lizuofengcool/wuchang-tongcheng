// Package dto 同城零工兼职数据传输对象 - 电子合同
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ContractInfo 合同详情响应
type ContractInfo struct {
	ID                uint       `json:"id"`
	ContractNo       string     `json:"contract_no"`
	ContractType      string     `json:"contract_type"`
	ContractTypeText  string     `json:"contract_type_text"`
	LinggongID        uint       `json:"linggong_id"`
	TaskID            uint       `json:"task_id"`
	ApplicationID     uint       `json:"application_id"`
	EmployerID        uint       `json:"employer_id"`
	EmployerName      string     `json:"employer_name"`
	EmployerPhone     string     `json:"employer_phone"`
	EmployerIDCard    string     `json:"employer_id_card"`
	EmployerSignURL   string     `json:"employer_sign_url"`
	WorkerID          uint       `json:"worker_id"`
	WorkerName        string     `json:"worker_name"`
	WorkerPhone       string     `json:"worker_phone"`
	WorkerIDCard      string     `json:"worker_id_card"`
	WorkerSignURL     string     `json:"worker_sign_url"`
	WorkStartDate     *time.Time `json:"work_start_date"`
	WorkEndDate       *time.Time `json:"work_end_date"`
	WorkContent       string     `json:"work_content"`
	WorkPlace         string     `json:"work_place"`
	BillingType       string     `json:"billing_type"`
	SalaryAmount      float64    `json:"salary_amount"`
	SalaryUnit        string     `json:"salary_unit"`
	Settlement        string     `json:"settlement"`
	SettlementText    string     `json:"settlement_text"`
	TotalAmount       float64    `json:"total_amount"`
	PaidAmount        float64    `json:"paid_amount"`
	Deposit           float64    `json:"deposit"`
	PenaltyBreach     float64    `json:"penalty_breach"`
	Confidential      bool       `json:"confidential"`
	NonCompete        bool       `json:"non_compete"`
	SignMethod        string     `json:"sign_method"`
	SignMethodText    string     `json:"sign_method_text"`
	ContractURL       string     `json:"contract_url"`
	Attachments       interface{} `json:"attachments"`
	EmployerSignedAt  *time.Time `json:"employer_signed_at"`
	WorkerSignedAt    *time.Time `json:"worker_signed_at"`
	SignedAt          *time.Time `json:"signed_at"`
	EffectiveAt       *time.Time `json:"effective_at"`
	ExpiredAt         *time.Time `json:"expired_at"`
	TerminatedAt      *time.Time `json:"terminated_at"`
	CompletedAt       *time.Time `json:"completed_at"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	TemplateID        uint       `json:"template_id"`
	Remark            string     `json:"remark"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateContractRequest 创建合同请求
type CreateContractRequest struct {
	ContractType    string     `json:"contract_type" binding:"omitempty,oneof=part_time temp task service internship project outsourcing"`
	LinggongID      uint       `json:"linggong_id" binding:"required"`
	TaskID          uint       `json:"task_id"`
	ApplicationID   uint       `json:"application_id"`
	EmployerID      uint       `json:"employer_id" binding:"required"`
	EmployerName    string     `json:"employer_name" binding:"max=128"`
	EmployerPhone   string     `json:"employer_phone" binding:"max=20"`
	EmployerIDCard  string     `json:"employer_id_card" binding:"max=32"`
	WorkerID        uint       `json:"worker_id" binding:"required"`
	WorkerName      string     `json:"worker_name" binding:"max=50"`
	WorkerPhone     string     `json:"worker_phone" binding:"max=20"`
	WorkerIDCard    string     `json:"worker_id_card" binding:"max=32"`
	WorkStartDate   *time.Time `json:"work_start_date"`
	WorkEndDate     *time.Time `json:"work_end_date"`
	WorkContent     string     `json:"work_content"`
	WorkPlace       string     `json:"work_place" binding:"max=500"`
	BillingType     string     `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	SalaryAmount    float64    `json:"salary_amount" binding:"min=0"`
	SalaryUnit      string     `json:"salary_unit" binding:"max=32"`
	Settlement      string     `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	TotalAmount     float64    `json:"total_amount" binding:"min=0"`
	Deposit         float64    `json:"deposit" binding:"min=0"`
	PenaltyBreach   float64    `json:"penalty_breach" binding:"min=0"`
	Confidential    bool       `json:"confidential"`
	NonCompete      bool       `json:"non_compete"`
	SignMethod      string     `json:"sign_method" binding:"omitempty,oneof=handwritten sms face ca"`
	ContractURL     string     `json:"contract_url" binding:"max=255"`
	Attachments     interface{} `json:"attachments"`
	TemplateID      uint       `json:"template_id"`
	Remark          string     `json:"remark"`
}

// UpdateContractRequest 更新合同请求
type UpdateContractRequest struct {
	WorkContent   *string `json:"work_content"`
	WorkPlace     *string `json:"work_place" binding:"omitempty,max=500"`
	WorkStartDate *time.Time `json:"work_start_date"`
	WorkEndDate   *time.Time `json:"work_end_date"`
	BillingType   *string `json:"billing_type" binding:"omitempty,oneof=by_piece by_hour by_day by_week by_month fixed negotiable"`
	SalaryAmount  *float64 `json:"salary_amount" binding:"omitempty,min=0"`
	SalaryUnit    *string `json:"salary_unit" binding:"omitempty,max=32"`
	Settlement    *string `json:"settlement" binding:"omitempty,oneof=T+0 T+1 T+3 T+7 M+1 project"`
	TotalAmount   *float64 `json:"total_amount" binding:"omitempty,min=0"`
	Deposit       *float64 `json:"deposit" binding:"omitempty,min=0"`
	PenaltyBreach  *float64 `json:"penalty_breach" binding:"omitempty,min=0"`
	Confidential  *bool   `json:"confidential"`
	NonCompete    *bool   `json:"non_compete"`
	ContractURL   *string `json:"contract_url" binding:"omitempty,max=255"`
	Attachments   interface{} `json:"attachments"`
	Remark        *string `json:"remark"`
}

// ContractListRequest 合同列表请求
type ContractListRequest struct {
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	TaskID        uint   `form:"task_id" json:"task_id"`
	ApplicationID uint   `form:"application_id" json:"application_id"`
	EmployerID    uint   `form:"employer_id" json:"employer_id"`
	WorkerID      uint   `form:"worker_id" json:"worker_id"`
	ContractType  string `form:"contract_type" json:"contract_type"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ContractAdminListRequest 管理后台合同列表请求
type ContractAdminListRequest struct {
	RegionID      uint   `form:"region_id" json:"region_id"`
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	EmployerID    uint   `form:"employer_id" json:"employer_id"`
	WorkerID      uint   `form:"worker_id" json:"worker_id"`
	ContractType  string `form:"contract_type" json:"contract_type"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ContractSignRequest 合同签署请求
type ContractSignRequest struct {
	SignRole   string `json:"sign_role" binding:"required,oneof=employer worker"`
	SignURL    string `json:"sign_url" binding:"max=255"`
	SignMethod string `json:"sign_method" binding:"omitempty,oneof=handwritten sms face ca"`
	SMSCode    string `json:"sms_code" binding:"max=16"`
}

// ContractStatusUpdateRequest 合同状态更新请求
type ContractStatusUpdateRequest struct {
	Status int    `json:"status" binding:"oneof=0 1 2 3 4 5 6 7 8"`
	Remark string `json:"remark"`
}
