// Package dto 同城拼车出行数据传输对象 - 举报（基于 pinche_complaints 表）
package dto

import (
	"time"
)

// ComplaintInfo 举报详情（前端字段命名为 reports 风格）
type ComplaintInfo struct {
	ID              uint       `json:"id"`
	RegionID        uint       `json:"region_id"`
	ReportNo        string     `json:"report_no"`
	PincheID        uint       `json:"pinche_id"`
	BookingID       *uint      `json:"booking_id"`
	TripID          *uint      `json:"trip_id"`
	ReporterID      uint       `json:"reporter_id"`
	ReporterName    string     `json:"reporter_name"`
	RespondentID    uint       `json:"respondent_id"`
	RespondentName  string     `json:"respondent_name"`
	ReportType      string     `json:"report_type"`
	TargetType      string     `json:"target_type"`
	TargetID        uint       `json:"target_id"`
	Reason          string     `json:"reason"`
	Description     string     `json:"description"`
	EvidenceImages  interface{} `json:"evidence_images,omitempty"`
	EvidenceCount   int        `json:"evidence_count"`
	EvidenceURLs    interface{} `json:"evidence_urls,omitempty"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	HandlerID       *uint      `json:"handler_id"`
	HandlerName     string     `json:"handler_name"`
	HandleResult    string     `json:"handle_result"`
	HandledAt       *time.Time `json:"handled_at"`
	PenaltyType     string     `json:"penalty_type"`
	PenaltyUserID   *uint      `json:"penalty_user_id"`
	SLADeadline     *time.Time `json:"sla_deadline"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateComplaintRequest 创建举报请求
type CreateComplaintRequest struct {
	TargetType string `json:"target_type"`
	TargetID   uint   `json:"target_id"`
	ReportType string `json:"report_type"`
	Reason     string `json:"reason" binding:"max=500"`
	PincheID   uint   `json:"pinche_id"`
	BookingID  *uint  `json:"booking_id"`
	TripID     *uint  `json:"trip_id"`
}

// ProcessComplaintRequest 处理举报请求
type ProcessComplaintRequest struct {
	HandleNote  string `json:"handle_note" binding:"max=500"`
	Result      string `json:"result" binding:"omitempty,oneof=valid invalid partial"`
	PenaltyType string `json:"penalty_type"`
	Status      int    `json:"status"`
}

// ComplaintListRequest 举报列表查询请求
type ComplaintListRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Keyword    string `form:"keyword" json:"keyword"`
	ReportType string `form:"report_type" json:"report_type"`
	TargetType string `form:"target_type" json:"target_type"`
	Status     *int   `form:"status" json:"status"`
	StartDate  string `form:"start_date" json:"start_date"`
	EndDate    string `form:"end_date" json:"end_date"`
}

// ComplaintStatsResponse 举报统计响应
type ComplaintStatsResponse struct {
	Total         int64 `json:"total"`
	Pending       int64 `json:"pending"`
	HighPriority  int64 `json:"high_priority"`
	Processed     int64 `json:"processed"`
}

// ComplaintListResult 举报列表响应（带统计）
type ComplaintListResult struct {
	List      []ComplaintInfo           `json:"list"`
	Total     int64                     `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
	Stats     ComplaintStatsResponse    `json:"stats"`
}
