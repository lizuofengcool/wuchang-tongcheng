// Package dto 同城114数据传输对象 - 举报
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportInfo 举报详情响应
// 字段命名与前端 reports.vue 对齐：
//   - reason 字段对应 model.ReportReason
//   - dh114_id 字段在 target_type='dh114' 时等于 target_id
//   - reported_user_id/reported_user_name 在 target_type='user' 时分别等于 target_id/target_name
type ReportInfo struct {
	ID                uint       `json:"id"`
	ReportNo          string     `json:"report_no"`
	ReporterID        uint       `json:"reporter_id"`
	ReporterName      string     `json:"reporter_name"`

	// 被举报对象
	TargetType        string     `json:"target_type"`
	TargetID          uint       `json:"target_id"`
	TargetName        string     `json:"target_name"`
	Dh114ID           uint       `json:"dh114_id"`            // 当 target_type='dh114' 时等于 target_id，便于前端使用
	ReportedUserID    uint       `json:"reported_user_id"`    // 当 target_type='user' 时等于 target_id
	ReportedUserName  string     `json:"reported_user_name"`  // 当 target_type='user' 时等于 target_name

	// 举报内容
	ReportType        string     `json:"report_type"`
	Reason            string     `json:"reason"`              // 对应 model.ReportReason
	Description       string     `json:"description"`
	EvidenceImages    interface{} `json:"evidence_images"`
	ContactInfo       string     `json:"contact_info"`

	// 处理
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	HandlerID         uint       `json:"handler_id"`
	HandlerName       string     `json:"handler_name"`
	HandleResult      string     `json:"handle_result"`
	HandledAt         *time.Time `json:"handled_at"`
	PenaltyType       string     `json:"penalty_type"`

	// SLA 截止时间（前端展示用，由 created_at + 48h 计算）
	SlaDeadline       *time.Time `json:"sla_deadline"`

	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateReportRequest 创建举报请求
// 字段命名与前端 reports.vue 对齐：dh114_id / report_type / reason / description / evidence_images
type CreateReportRequest struct {
	Dh114ID        uint        `json:"dh114_id" binding:"required"`
	ReportType     string      `json:"report_type" binding:"required,oneof=porn scam fake prohibited infringement spam other"`
	Reason         string      `json:"reason" binding:"required,max=500"`
	Description    string      `json:"description" binding:"max=1000"`
	EvidenceImages interface{} `json:"evidence_images"`
	ContactInfo    string      `json:"contact_info" binding:"max=100"`

	// 可选扩展字段（兼容 target_type/target_id 用法）
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
	TargetName string `json:"target_name"`
}

// ProcessReportRequest 处理举报请求
// 字段命名与前端 reports.vue 对齐：status / penalty_type / handle_result
// 同时兼容任务规范中的 handle_note / result 字段
type ProcessReportRequest struct {
	Status       int    `json:"status" binding:"required,oneof=1 2 3 4 5"`
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
	HandleResult string `json:"handle_result" binding:"max=500"`
	HandleNote   string `json:"handle_note" binding:"max=500"` // 兼容字段，存在时与 handle_result 合并
	Result       string `json:"result" binding:"omitempty,oneof=valid invalid"` // 兼容字段
}

// ReportListRequest 举报列表请求
// 支持前端字段（page/page_size/report_type/status/dh114_id）和任务规范字段（keyword/target_type）
type ReportListRequest struct {
	utils.Pagination
	Keyword    string `form:"keyword" json:"keyword"`
	Status     *int   `form:"status" json:"status"`
	ReportType string `form:"report_type" json:"report_type"`
	TargetType string `form:"target_type" json:"target_type"`
	Dh114ID    uint   `form:"dh114_id" json:"dh114_id"`
}

// ReportStats 举报统计
type ReportStats struct {
	Total     int64 `json:"total"`     // 总数
	Pending   int64 `json:"pending"`   // 待处理数（status=0）
	Processed int64 `json:"processed"` // 已处理数（status>0）
}

// ReportListResponse 举报列表响应（包含统计）
type ReportListResponse struct {
	List  []ReportInfo `json:"list"`
	Total int64        `json:"total"`
	Stats ReportStats  `json:"stats"`
}
