// Package model 薪资支付表（对标兼职猫日结 + 支付中台）
// 薪资日结：T+0/T+1/T+7 多种结算方式 + 工资单/支付记录
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 支付状态常量 ===
const (
	PaymentStatusPending    = 0 // 待支付
	PaymentStatusProcessing = 1 // 处理中
	PaymentStatusPaid       = 2 // 已支付
	PaymentStatusFailed     = 3 // 支付失败
	PaymentStatusRefunded   = 4 // 已退款
	PaymentStatusDisputed  = 5 // 争议中
	PaymentStatusCanceled   = 6 // 已取消
	PaymentStatusPartial    = 7 // 部分支付
)

// === 支付方式常量 ===
const (
	PaymentMethodWechat    = "wechat"     // 微信支付
	PaymentMethodAlipay    = "alipay"     // 支付宝
	PaymentMethodBank      = "bank"       // 银行卡
	PaymentMethodCash      = "cash"       // 现金
	PaymentMethodBalance   = "balance"    // 余额支付
	PaymentMethodEscrow    = "escrow"     // 担保支付
)

// === 支付类型常量 ===
const (
	PaymentTypeSalary     = "salary"     // 工资
	PaymentTypeBonus      = "bonus"      // 奖金
	PaymentTypeOvertime   = "overtime"   // 加班费
	PaymentTypeAllowance  = "allowance"  // 补贴
	PaymentTypeReimburse  = "reimburse"  // 报销
	PaymentTypePenalty    = "penalty"    // 罚款
	PaymentTypeRefund     = "refund"     // 退款
	PaymentTypeDeposit    = "deposit"    // 押金
)

// === 结算状态常量 ===
const (
	SettlementStatusPending    = 0 // 待结算
	SettlementStatusSettling   = 1 // 结算中
	SettlementStatusSettled    = 2 // 已结算
	SettlementStatusFailed     = 3 // 结算失败
)

// LinggongPayment 薪资支付表
type LinggongPayment struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	PaymentNo       string     `gorm:"size:64;not null;uniqueIndex" json:"payment_no"`             // 支付单号
	LinggongID     uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID         uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	ApplicationID  uint       `gorm:"not null;default:0;index" json:"application_id"`             // 报名记录 ID
	ContractID     uint       `gorm:"not null;default:0;index" json:"contract_id"`               // 合同 ID
	EmployerID     uint       `gorm:"not null;index" json:"employer_id"`                          // 雇主 ID
	EmployerName   string     `gorm:"size:128;not null;default:''" json:"employer_name"`          // 雇主名称
	WorkerID       uint       `gorm:"not null;index" json:"worker_id"`                            // 求职者 ID
	WorkerName     string     `gorm:"size:50;not null;default:''" json:"worker_name"`              // 求职者姓名
	WorkerPhone    string     `gorm:"size:20;not null;default:''" json:"worker_phone"`            // 求职者手机
	WorkerBankAccount string  `gorm:"size:64;not null;default:''" json:"worker_bank_account"`      // 求职者银行账号
	WorkerAlipay   string     `gorm:"size:128;not null;default:''" json:"worker_alipay"`          // 求职者支付宝
	WorkerWechat   string     `gorm:"size:128;not null;default:''" json:"worker_wechat"`          // 求职者微信
	PaymentType    string     `gorm:"size:32;not null;default:'salary';index" json:"payment_type"` // salary/bonus/overtime/allowance/reimburse/penalty/refund/deposit
	Amount         float64    `gorm:"type:decimal(12,2);default:0;index" json:"amount"`            // 支付金额
	WorkHours      float64    `gorm:"type:decimal(8,2);default:0" json:"work_hours"`               // 工作时长
	WorkDays       int        `gorm:"not null;default:0" json:"work_days"`                        // 工作天数
	TaskCount      int        `gorm:"not null;default:0" json:"task_count"`                       // 任务数
	UnitPrice      float64    `gorm:"type:decimal(12,2);default:0" json:"unit_price"`              // 单价
	Quantity       float64    `gorm:"type:decimal(10,2);default:0" json:"quantity"`                // 数量
	Settlement     string     `gorm:"size:16;not null;default:'T+1';index" json:"settlement"`      // T+0/T+1/T+3/T+7/M+1/project
	SettlementStatus int      `gorm:"default:0;index" json:"settlement_status"`                    // 0待结算 1结算中 2已结算 3失败
	SettlementAt   *time.Time `gorm:"index" json:"settlement_at"`                                // 结算时间
	DueAt          *time.Time `gorm:"index" json:"due_at"`                                        // 应结时间
	PayMethod      string     `gorm:"size:32;not null;default:'wechat';index" json:"pay_method"`   // wechat/alipay/bank/cash/balance/escrow
	PayTradeNo     string     `gorm:"size:128;not null;default:'';index" json:"pay_trade_no"`      // 第三方支付单号
	PayChannel     string     `gorm:"size:32;not null;default:''" json:"pay_channel"`              // 支付渠道
	PayeeName      string     `gorm:"size:50;not null;default:''" json:"payee_name"`               // 收款人姓名
	PlatformFee   float64    `gorm:"type:decimal(12,2);default:0" json:"platform_fee"`              // 平台手续费
	TaxAmount     float64    `gorm:"type:decimal(12,2);default:0" json:"tax_amount"`                // 税费
	ActualAmount  float64    `gorm:"type:decimal(12,2);default:0" json:"actual_amount"`          // 实际到账
	Status        int        `gorm:"default:0;index" json:"status"`                              // 0待支付 1处理中 2已支付 3失败 4退款 5争议 6取消 7部分
	FailedReason  string     `gorm:"size:500;not null;default:''" json:"failed_reason"`           // 失败原因
	PaidAt        *time.Time `gorm:"index" json:"paid_at"`                                        // 支付时间
	ConfirmedAt   *time.Time `gorm:"index" json:"confirmed_at"`                                  // 确认时间
	CanceledAt    *time.Time `gorm:"index" json:"canceled_at"`                                   // 取消时间
	WorkStartDate *time.Time `gorm:"type:date" json:"work_start_date"`                            // 工作开始日期
	WorkEndDate   *time.Time `gorm:"type:date" json:"work_end_date"`                              // 工作结束日期
	EvidenceImages JSONB     `gorm:"type:jsonb" json:"evidence_images"`                          // 工作凭证图片
	Remark         string     `gorm:"type:text" json:"remark"`                                    // 备注
	InvoiceURL    string     `gorm:"size:255;not null;default:''" json:"invoice_url"`             // 发票 URL
}

// TableName 表名（linggong_ 前缀）
func (LinggongPayment) TableName() string { return "linggong_payments" }
