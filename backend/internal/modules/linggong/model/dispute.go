// Package model 纠纷表（对标闲鱼/瓜子）
// 工单 + 证据 + 调解 + 仲裁
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 纠纷状态常量 ===
const (
	DisputeStatusPending    = 0 // 待处理
	DisputeStatusProcessing = 1 // 处理中
	DisputeStatusMediated   = 2 // 已调解
	DisputeStatusArbitrated = 3 // 已仲裁
	DisputeStatusResolved   = 4 // 已解决
	DisputeStatusRejected   = 5 // 已驳回
	DisputeStatusCanceled   = 6 // 已取消
	DisputeStatusAppealed   = 7 // 申诉中
)

// === 纠纷类型常量 ===
const (
	DisputeTypeSalary        = "salary"         // 薪资纠纷
	DisputeTypeQuality       = "quality"        // 工作质量纠纷
	DisputeTypeAttendance    = "attendance"    // 考勤纠纷
	DisputeTypeBreach        = "breach"         // 违约
	DisputeTypeDiscrimination = "discrimination" // 歧视
	DisputeTypeHarassment    = "harassment"    // 骚扰
	DisputeTypeFraud         = "fraud"           // 欺诈
	DisputeTypeSafety        = "safety"          // 安全问题
	DisputeTypeOther         = "other"          // 其他
)

// === 申请人类型常量 ===
const (
	DisputeApplicantWorker   = "worker"   // 求职者申请
	DisputeApplicantEmployer = "employer" // 雇主申请
	DisputeApplicantPlatform = "platform" // 平台介入
)

// === 处理结果常量 ===
const (
	DisputeResultWorkerWin    = "worker_win"     // 求职者胜
	DisputeResultEmployerWin  = "employer_win"   // 雇主胜
	DisputeResultCompromise   = "compromise"     // 协商解决
	DisputeResultPlatformDecision = "platform_decision" // 平台裁决
	DisputeResultReject       = "reject"          // 驳回
)

// LinggongDispute 纠纷表
type LinggongDispute struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	DisputeNo       string     `gorm:"size:64;not null;uniqueIndex" json:"dispute_no"`              // 纠纷单号
	LinggongID     uint       `gorm:"not null;index" json:"linggong_id"`                          // 岗位 ID
	TaskID         uint       `gorm:"not null;default:0;index" json:"task_id"`                   // 任务包 ID
	ApplicationID  uint       `gorm:"not null;default:0;index" json:"application_id"`             // 报名记录 ID
	ContractID     uint       `gorm:"not null;default:0;index" json:"contract_id"`               // 合同 ID
	PaymentID      uint       `gorm:"not null;default:0;index" json:"payment_id"`                 // 支付单 ID
	DisputeType    string     `gorm:"size:32;not null;default:'other';index" json:"dispute_type"` // salary/quality/attendance/breach/discrimination/harassment/fraud/safety/other
	ApplicantType  string     `gorm:"size:16;not null;default:'worker';index" json:"applicant_type"` // worker/employer/platform
	ApplicantID    uint       `gorm:"not null;index" json:"applicant_id"`                          // 申请人 ID
	ApplicantName  string     `gorm:"size:50;not null;default:''" json:"applicant_name"`           // 申请人姓名
	RespondentID   uint       `gorm:"not null;index" json:"respondent_id"`                         // 被申请人 ID
	RespondentName string     `gorm:"size:50;not null;default:''" json:"respondent_name"`         // 被申请人姓名
	Title          string     `gorm:"size:200;not null" json:"title"`                              // 纠纷标题
	Description   string     `gorm:"type:text" json:"description"`                                // 纠纷描述
	EvidenceImages JSONB      `gorm:"type:jsonb" json:"evidence_images"`                          // 证据图片
	EvidenceVideos JSONB      `gorm:"type:jsonb" json:"evidence_videos"`                          // 证据视频
	EvidenceDocs   JSONB      `gorm:"type:jsonb" json:"evidence_docs"`                            // 证据文档
	ClaimAmount    float64    `gorm:"type:decimal(12,2);default:0" json:"claim_amount"`            // 诉求金额
	Status         int        `gorm:"default:0;index" json:"status"`                              // 0待处理 1处理中 2调解 3仲裁 4解决 5驳回 6取消 7申诉
	HandlerID     uint       `gorm:"not null;default:0;index" json:"handler_id"`                  // 处理人 ID
	HandlerName   string     `gorm:"size:50;not null;default:''" json:"handler_name"`             // 处理人姓名
	MediationResult string   `gorm:"type:text" json:"mediation_result"`                          // 调解结果
	ArbitrationResult string `gorm:"type:text" json:"arbitration_result"`                          // 仲裁结果
	FinalResult    string     `gorm:"size:32;not null;default:''" json:"final_result"`            // 最终结果 worker_win/employer_win/compromise/platform_decision/reject
	CompensationAmount float64 `gorm:"type:decimal(12,2);default:0" json:"compensation_amount"`  // 赔偿金额
	SLADeadline   *time.Time `gorm:"index" json:"sla_deadline"`                                   // SLA 截止时间
	HandledAt     *time.Time `gorm:"index" json:"handled_at"`                                    // 处理时间
	ResolvedAt    *time.Time `gorm:"index" json:"resolved_at"`                                  // 解决时间
	ClosedAt      *time.Time `gorm:"index" json:"closed_at"`                                     // 关闭时间
	AppealReason  string     `gorm:"type:text" json:"appeal_reason"`                              // 申诉原因
	AppealedAt    *time.Time `gorm:"index" json:"appealed_at"`                                   // 申诉时间
	AppealResult  string     `gorm:"type:text" json:"appeal_result"`                              // 申诉结果
	AppealHandlerID uint     `gorm:"not null;default:0" json:"appeal_handler_id"`                 // 申诉处理人 ID
	AppealHandledAt *time.Time `gorm:"index" json:"appeal_handled_at"`                            // 申诉处理时间
}

// TableName 表名（linggong_ 前缀）
func (LinggongDispute) TableName() string { return "linggong_disputes" }
