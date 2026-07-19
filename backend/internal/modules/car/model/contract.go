// Package model 合同电子化表（对标瓜子）
// 买卖/置换/租车 + 电子签 + 贷款/服务费
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 合同类型常量 ===
const (
	ContractTypeSale    = "sale"    // 买卖合同
	ContractTypeReplace = "replace" // 置换协议
	ContractTypeRental  = "rental"  // 租车协议
	ContractTypeFinance = "finance" // 贷款合同
)

// === 合同状态常量 ===
const (
	ContractStatusDraft      = 0 // 草稿
	ContractStatusPending    = 1 // 待签署
	ContractStatusSigned     = 2 // 已签署
	ContractStatusEffective  = 3 // 已生效
	ContractStatusFulfilling = 4 // 履行中
	ContractStatusCompleted  = 5 // 已完成
	ContractStatusTerminated = 6 // 已终止
	ContractStatusCanceled   = 7 // 已取消
)

// === 付款方式常量 ===
const (
	PaymentMethodFull        = "full"        // 全款
	PaymentMethodLoan        = "loan"        // 贷款
	PaymentMethodInstallment = "installment" // 分期
	PaymentMethodDeposit     = "deposit"     // 定金 + 尾款
)

// CarContract 合同电子化表
type CarContract struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ContractNo      string     `gorm:"size:64;not null;uniqueIndex" json:"contract_no"`                // 合同编号
	ContractType    string     `gorm:"size:32;not null;default:'sale';index" json:"contract_type"`     // sale/replace/rental/finance
	CarID           uint       `gorm:"not null;index" json:"car_id"`                                   // 车源 ID
	ListingID       uint       `gorm:"not null;default:0;index" json:"listing_id"`                     // 发布 ID
	SellerID        uint       `gorm:"not null;index" json:"seller_id"`                                // 卖方 ID
	SellerName      string     `gorm:"size:50;not null;default:''" json:"seller_name"`                 // 卖方姓名
	SellerPhone     string     `gorm:"size:20;not null;default:''" json:"seller_phone"`                // 卖方手机
	SellerIDCard    string     `gorm:"size:32;not null;default:''" json:"seller_id_card"`              // 卖方身份证
	BuyerID         uint       `gorm:"not null;index" json:"buyer_id"`                                 // 买方 ID
	BuyerName       string     `gorm:"size:50;not null;default:''" json:"buyer_name"`                  // 买方姓名
	BuyerPhone      string     `gorm:"size:20;not null;default:''" json:"buyer_phone"`                 // 买方手机
	BuyerIDCard     string     `gorm:"size:32;not null;default:''" json:"buyer_id_card"`               // 买方身份证
	AgentID         uint       `gorm:"not null;default:0;index" json:"agent_id"`                       // 经办人 ID
	AgentName       string     `gorm:"size:50;not null;default:''" json:"agent_name"`                  // 经办人姓名
	DealPrice       float64    `gorm:"type:decimal(14,2);default:0;index" json:"deal_price"`           // 成交价
	Deposit         float64    `gorm:"type:decimal(14,2);default:0" json:"deposit"`                    // 定金
	PaymentMethod   string     `gorm:"size:32;not null;default:'full'" json:"payment_method"`          // full/loan/installment/deposit
	LoanAmount      float64    `gorm:"type:decimal(14,2);default:0" json:"loan_amount"`                // 贷款金额
	LoanPeriods     int        `gorm:"not null;default:0" json:"loan_periods"`                         // 贷款期数
	TransferFee     float64    `gorm:"type:decimal(14,2);default:0" json:"transfer_fee"`               // 过户费
	ServiceFee      float64    `gorm:"type:decimal(14,2);default:0" json:"service_fee"`                // 服务费
	ContractURL     string     `gorm:"size:255;not null;default:''" json:"contract_url"`               // 合同 PDF URL
	Attachments     JSONB      `gorm:"type:jsonb" json:"attachments"`                                  // 附件 JSON
	SignedAt        *time.Time `gorm:"index" json:"signed_at"`                                         // 签署时间
	EffectiveAt     *time.Time `gorm:"index" json:"effective_at"`                                      // 生效时间
	ExpiredAt       *time.Time `gorm:"index" json:"expired_at"`                                        // 到期时间
	TerminatedAt    *time.Time `gorm:"index" json:"terminated_at"`                                     // 终止时间
	Status          int        `gorm:"default:0;index" json:"status"`                                 // 0草稿 1待签 2已签 3生效 4履行 5完成 6终止 7取消
	Remark          string     `gorm:"type:text" json:"remark"`                                        // 备注
}

// TableName 表名（car_ 前缀）
func (CarContract) TableName() string { return "car_contracts" }
