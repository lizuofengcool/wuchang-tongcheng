// Package dto 同城商城 - 举报 DTO
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportInfo 举报详情响应
type ReportInfo struct {
	ID             uint       `json:"id"`
	ReportNo       string     `json:"report_no"`
	ReporterID     uint       `json:"reporter_id"`
	ReporterName   string     `json:"reporter_name"`
	TargetType     string     `json:"target_type"`
	TargetID       uint       `json:"target_id"`
	TargetName     string     `json:"target_name"`
	ReportType     string     `json:"report_type"`
	ReportReason   string     `json:"report_reason"`
	Description    string     `json:"description"`
	EvidenceImages interface{} `json:"evidence_images"`
	ContactInfo    string     `json:"contact_info"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	HandlerID      uint       `json:"handler_id"`
	HandlerName    string     `json:"handler_name"`
	HandleResult   string     `json:"handle_result"`
	HandledAt      *time.Time `json:"handled_at"`
	PenaltyType    string     `json:"penalty_type"`
	PenaltyTargetID uint      `json:"penalty_target_id"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateReportRequest 创建举报请求
type CreateReportRequest struct {
	TargetType     string      `json:"target_type" binding:"required,oneof=product shop review user order"`
	TargetID       uint        `json:"target_id" binding:"required"`
	TargetName     string      `json:"target_name" binding:"max=255"`
	ReportType     string      `json:"report_type" binding:"required,oneof=porn scam fake prohibited infringement spam counterfeit other"`
	ReportReason   string      `json:"report_reason" binding:"required,max=500"`
	Description    string      `json:"description" binding:"max=1000"`
	EvidenceImages interface{} `json:"evidence_images"`
	ContactInfo    string      `json:"contact_info" binding:"max=100"`
}

// ProcessReportRequest 处理举报请求
type ProcessReportRequest struct {
	Status       int    `json:"status" binding:"required,oneof=1 2 3 4 5"` // 1有效 2无效 3处理中 4已处罚 5已驳回
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d banForever"`
	HandleResult string `json:"handle_result" binding:"max=500"`
	HandleNote   string `json:"handle_note" binding:"max=500"`
}

// ReportListRequest 举报列表请求
type ReportListRequest struct {
	utils.Pagination
	Keyword     string `form:"keyword" json:"keyword"`
	Status      *int   `form:"status" json:"status"`
	ReportType  string `form:"report_type" json:"report_type"`
	TargetType  string `form:"target_type" json:"target_type"`
	TargetID    uint   `form:"target_id" json:"target_id"`
	ReporterID  uint   `form:"reporter_id" json:"reporter_id"`
	HandlerID   uint   `form:"handler_id" json:"handler_id"`
	StartDate   string `form:"start_date" json:"start_date"`
	EndDate     string `form:"end_date" json:"end_date"`
	RegionID    uint   `form:"region_id" json:"region_id"`
}

// ReportStats 举报统计
type ReportStats struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Processed int64 `json:"processed"`
	Valid     int64 `json:"valid"`
	Invalid   int64 `json:"invalid"`
}
