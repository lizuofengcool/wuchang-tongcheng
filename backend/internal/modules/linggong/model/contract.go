// Package model 电子合同表（对标法大大/e签宝）
// 兼职合同电子化 + 在线签署 + 模板
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 合同类型常量 ===
const (
	ContractTypePartTime    = "part_time"     // 兼职合同
	ContractTypeTemp        = "temp"          // 临时合同
	ContractTypeTask        = "task"          // 任务合同
	ContractTypeService     = "service"       // 服务合同
	ContractTypeInternship  = "internship"    // 实习合同
	ContractTypeProject     = "project"       // 项目合同
	ContractTypeOutsourcing = "outsourcing"  // 外包合同
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
	ContractStatusExpired    = 8 // 已过期
)

// === 签署方式常量 ===
const (
	SignMethodHandwritten   = "handwritten"   // 手写签名
	SignMethodSMS           = "sms"           // 短信验证
	SignMethodFace         = "face"           // 人脸识别
	SignMethodCA            = "ca"            // CA 证书
)

// LinggongContract 电子合同表
type LinggongContract struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ContractNo     string     `gorm:"size:64;not null;uniqueIndex" json:"contract_no"`           // 合同编号
	ContractType   string     `gorm:"size:32;not null;default:'part_time';index" json:"contract_type"` // part_time/temp/task/service/internship/project/outsourcing
	LinggongID     uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID         uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	ApplicationID  uint       `gorm:"not null;default:0;index" json:"application_id"`             // 报名记录 ID
	EmployerID     uint       `gorm:"not null;index" json:"employer_id"`                          // 雇主 ID
	EmployerName   string     `gorm:"size:128;not null;default:''" json:"employer_name"`          // 雇主名称
	EmployerPhone  string     `gorm:"size:20;not null;default:''" json:"employer_phone"`           // 雇主手机
	EmployerIDCard string     `gorm:"size:32;not null;default:''" json:"employer_id_card"`         // 雇主身份证
	EmployerSignURL string    `gorm:"size:255;not null;default:''" json:"employer_sign_url"`       // 雇主签名图片
	WorkerID       uint       `gorm:"not null;index" json:"worker_id"`                            // 求职者 ID
	WorkerName     string     `gorm:"size:50;not null;default:''" json:"worker_name"`              // 求职者姓名
	WorkerPhone    string     `gorm:"size:20;not null;default:''" json:"worker_phone"`            // 求职者手机
	WorkerIDCard   string     `gorm:"size:32;not null;default:''" json:"worker_id_card"`           // 求职者身份证
	WorkerSignURL  string     `gorm:"size:255;not null;default:''" json:"worker_sign_url"`         // 求职者签名图片
	WorkStartDate  *time.Time `gorm:"type:date" json:"work_start_date"`                            // 工作开始日期
	WorkEndDate    *time.Time `gorm:"type:date" json:"work_end_date"`                              // 工作结束日期
	WorkContent    string     `gorm:"type:text" json:"work_content"`                              // 工作内容
	WorkPlace      string     `gorm:"size:500;not null;default:''" json:"work_place"`             // 工作地点
	BillingType    string     `gorm:"size:32;not null;default:'by_day'" json:"billing_type"`      // by_piece/by_hour/by_day/fixed
	SalaryAmount   float64    `gorm:"type:decimal(12,2);default:0" json:"salary_amount"`            // 薪资金额
	SalaryUnit     string     `gorm:"size:32;not null;default:''" json:"salary_unit"`              // 薪资单位
	Settlement     string     `gorm:"size:16;not null;default:'T+1'" json:"settlement"`           // T+0/T+1/T+7/M+1/project
	TotalAmount    float64    `gorm:"type:decimal(12,2);default:0" json:"total_amount"`            // 合同总额
	PaidAmount     float64    `gorm:"type:decimal(12,2);default:0" json:"paid_amount"`              // 已支付金额
	Deposit        float64    `gorm:"type:decimal(12,2);default:0" json:"deposit"`                // 押金
	PenaltyBreach  float64    `gorm:"type:decimal(12,2);default:0" json:"penalty_breach"`          // 违约金
	Confidential   bool       `gorm:"not null;default:false" json:"confidential"`                  // 保密协议
	NonCompete     bool       `gorm:"not null;default:false" json:"non_compete"`                  // 竞业协议
	SignMethod     string     `gorm:"size:32;not null;default:'handwritten'" json:"sign_method"`    // handwritten/sms/face/ca
	ContractURL    string     `gorm:"size:255;not null;default:''" json:"contract_url"`            // 合同 PDF URL
	Attachments    JSONB      `gorm:"type:jsonb" json:"attachments"`                              // 附件 JSON
	EmployerSignedAt *time.Time `gorm:"index" json:"employer_signed_at"`                          // 雇主签署时间
	WorkerSignedAt *time.Time `gorm:"index" json:"worker_signed_at"`                              // 求职者签署时间
	SignedAt       *time.Time `gorm:"index" json:"signed_at"`                                    // 完成签署时间
	EffectiveAt    *time.Time `gorm:"index" json:"effective_at"`                                 // 生效时间
	ExpiredAt      *time.Time `gorm:"index" json:"expired_at"`                                   // 到期时间
	TerminatedAt   *time.Time `gorm:"index" json:"terminated_at"`                                // 终止时间
	CompletedAt   *time.Time `gorm:"index" json:"completed_at"`                                  // 完成时间
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0草稿 1待签 2已签 3生效 4履行 5完成 6终止 7取消 8过期
	TemplateID     uint       `gorm:"not null;default:0" json:"template_id"`                       // 合同模板 ID
	Remark         string     `gorm:"type:text" json:"remark"`                                   // 备注
}

// TableName 表名（linggong_ 前缀）
func (LinggongContract) TableName() string { return "linggong_contracts" }
