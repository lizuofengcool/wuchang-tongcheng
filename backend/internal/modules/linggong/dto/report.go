// Package dto 同城零工兼职数据传输对象 - 举报管理（复用 disputes 表）
// 依据需求：M 端举报管理端点，前端字段对齐 reports.vue
// 复用 linggong_disputes 表，前端"举报"语义对应 disputes 表中"申请人向平台举报被申请人"的工单
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 举报状态常量（前端 statusMap 对齐） =====
// 0 待处理 / 1 已核实警告 / 2 已下架 / 3 已封号 / 4 已驳回 / 5 已转交
// 注意：复用 disputes 表 status 字段，含义以举报语义为准（前端 reports.vue statusMap）

// ReportListRequest 举报列表请求（M 端）
type ReportListRequest struct {
	ReportType string `form:"report_type" json:"report_type"` // 举报类型（对应 dispute_type）
	Status     *int   `form:"status" json:"status"`          // 0待处理 1-5 已处理
	TargetType string `form:"target_type" json:"target_type"` // 目标类型（linggong/task/...，仅过滤参考）
	Keyword    string `form:"keyword" json:"keyword"`        // 举报单号/原因 模糊
	utils.Pagination
}

// ReportProcessRequest 处理举报请求
// 字段对齐前端 reports.vue onProcess 提交：
//
//	{status, penalty_type, credit_change, handle_result}
//
// 同时兼容任务描述字段：handle_note / result(valid|invalid)
type ReportProcessRequest struct {
	Status       int    `json:"status" binding:"oneof=1 2 3 4 5"`                          // 1核实警告 2下架 3封号 4驳回 5转交
	HandleResult string `json:"handle_result" binding:"max=500"`                           // 处理说明（前端字段）
	HandleNote   string `json:"handle_note" binding:"max=500"`                             // 处理备注（任务描述字段，兼容）
	Result       string `json:"result" binding:"omitempty,oneof=valid invalid"`           // 任务描述：valid 有效举报 / invalid 无效举报
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning limit credit ban1d ban7d banForever"`
	CreditChange int    `json:"credit_change"`                                            // 信用分变化（penalty_type=credit 时有效）
}

// ReportStats 举报统计
type ReportStats struct {
	Total     int64 `json:"total"`     // 举报总数
	Pending   int64 `json:"pending"`   // 待处理数（status=0）
	Processed int64 `json:"processed"` // 已处理数（status>0）
}

// ReportInfo 举报信息（前端字段名对齐 reports.vue）
// 复用 disputes 表，字段映射：
//
//	DisputeNo        -> ReportNo
//	ApplicantID/Name -> ReporterID/Name
//	RespondentID/Name -> ReportedUserID/Name
//	DisputeType      -> ReportType
//	Title            -> Reason
//	MediationResult  -> HandleResult
//	LinggongID       -> TargetID（默认 target_type=linggong）
type ReportInfo struct {
	ID               uint       `json:"id"`
	ReportNo         string     `json:"report_no"` // = dispute_no
	LinggongID       uint       `json:"linggong_id"`
	TaskID           uint       `json:"task_id"`
	ApplicationID    uint       `json:"application_id"`
	ContractID       uint       `json:"contract_id"`
	PaymentID        uint       `json:"payment_id"`
	TargetType       string     `json:"target_type"` // 默认 'linggong'
	TargetID         uint       `json:"target_id"`   // = linggong_id
	ReportType       string     `json:"report_type"` // = dispute_type
	ReportTypeText   string     `json:"report_type_text"`
	ReporterID       uint       `json:"reporter_id"`       // = applicant_id
	ReporterName     string     `json:"reporter_name"`     // = applicant_name
	ReporterPhone    string     `json:"reporter_phone"`    // disputes 表无对应字段，留空
	ReportedUserID   uint       `json:"reported_user_id"`  // = respondent_id
	ReportedUserName string     `json:"reported_user_name"` // = respondent_name
	ReportedPhone    string     `json:"reported_phone"`    // disputes 表无对应字段，留空
	Reason           string     `json:"reason"`            // = title
	Description      string     `json:"description"`
	EvidenceImages   interface{} `json:"evidence_images"`
	EvidenceVideos   interface{} `json:"evidence_videos"`
	EvidenceDocs     interface{} `json:"evidence_docs"`
	ClaimAmount      float64    `json:"claim_amount"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	HandlerID        uint       `json:"handler_id"`
	HandlerName      string     `json:"handler_name"`
	HandleResult     string     `json:"handle_result"` // = mediation_result
	PenaltyType      string     `json:"penalty_type"`   // 处罚类型（处理时记录到 mediation_result 文本中）
	SLADeadline      *time.Time `json:"sla_deadline"`
	HandledAt        *time.Time `json:"handled_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	ClosedAt         *time.Time `json:"closed_at"`
	RegionID         uint       `json:"region_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
