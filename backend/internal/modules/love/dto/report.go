// Package dto love 相亲交友数据传输对象 - 举报
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveReportInfo 举报记录响应
type LoveReportInfo struct {
	ID                uint       `json:"id"`
	ReportNo          string     `json:"report_no"`
	ReporterUserID    uint       `json:"reporter_user_id"`
	ReporterLoveID    uint       `json:"reporter_love_id"`
	ReporterNickname  string     `json:"reporter_nickname"`
	ReporterAvatar    string     `json:"reporter_avatar"`
	TargetType        string     `json:"target_type"`
	TargetUserID      uint       `json:"target_user_id"`
	TargetLoveID      uint       `json:"target_love_id"`
	TargetNickname    string     `json:"target_nickname"`
	TargetAvatar      string     `json:"target_avatar"`
	TargetID          uint       `json:"target_id"`
	ReasonType        string     `json:"reason_type"`
	ReasonDetail      string     `json:"reason_detail"`
	EvidenceImages    interface{} `json:"evidence_images"`
	EvidenceVideos    interface{} `json:"evidence_videos"`
	EvidenceText      string     `json:"evidence_text"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	HandledBy         uint       `json:"handled_by"`
	HandledAt         *time.Time `json:"handled_at"`
	HandleResult      string     `json:"handle_result"`
	HandleRemark      string     `json:"handle_remark"`
	PenaltyType       string     `json:"penalty_type"`
	PenaltyDuration   int        `json:"penalty_duration"`
	PenaltyExpiredAt  *time.Time `json:"penalty_expired_at"`
	AppealStatus      int        `json:"appeal_status"`
	AppealReason      string     `json:"appeal_reason"`
	AppealedAt        *time.Time `json:"appealed_at"`
	AppealResult      string     `json:"appeal_result"`
	RiskScore         int        `json:"risk_score"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateLoveReportRequest 举报请求
type CreateLoveReportRequest struct {
	TargetType     string      `json:"target_type" binding:"required,oneof=user story message impression gift"`
	TargetUserID   uint        `json:"target_user_id" binding:"required"`
	TargetLoveID   uint        `json:"target_love_id" binding:"required"`
	TargetID       uint        `json:"target_id"`
	ReasonType     string      `json:"reason_type" binding:"required,oneof=fake_info fraud harassment porn political insult spam minor other"`
	ReasonDetail   string      `json:"reason_detail" binding:"max=2000"`
	EvidenceImages interface{} `json:"evidence_images"`
	EvidenceVideos interface{} `json:"evidence_videos"`
	EvidenceText   string      `json:"evidence_text"`
}

// LoveReportListRequest 举报列表请求
type LoveReportListRequest struct {
	ReporterUserID uint   `form:"reporter_user_id" json:"reporter_user_id"`
	TargetUserID   uint   `form:"target_user_id" json:"target_user_id"`
	TargetType     string `form:"target_type" json:"target_type"`
	ReasonType     string `form:"reason_type" json:"reason_type"`
	Status         *int   `form:"status" json:"status"`
	utils.Pagination
}

// LoveReportHandleRequest 处理举报请求
type LoveReportHandleRequest struct {
	ID            uint   `json:"id" binding:"required"`
	HandleResult  string `json:"handle_result" binding:"required,oneof=valid invalid processing closed"`
	HandleRemark  string `json:"handle_remark" binding:"max=500"`
	PenaltyType   string `json:"penalty_type" binding:"omitempty,oneof=warning freeze ban delete_content deduction none"`
	PenaltyDuration int  `json:"penalty_duration"`
}

// LoveReportAppealRequest 举报申诉请求
type LoveReportAppealRequest struct {
	ID           uint   `json:"id" binding:"required"`
	AppealReason string `json:"appeal_reason" binding:"required,max=2000"`
}

// LoveReportAppealHandleRequest 申诉处理请求
type LoveReportAppealHandleRequest struct {
	ID            uint   `json:"id" binding:"required"`
	AppealResult  string `json:"appeal_result" binding:"required,oneof=approved rejected"`
	AppealRemark  string `json:"appeal_remark" binding:"max=500"`
}

// LoveReportStatsResponse 举报统计响应
type LoveReportStatsResponse struct {
	TotalReports      int64 `json:"total_reports"`
	TodayReports      int64 `json:"today_reports"`
	PendingReports    int64 `json:"pending_reports"`
	HandledReports    int64 `json:"handled_reports"`
	ValidReports      int64 `json:"valid_reports"`
	InvalidReports    int64 `json:"invalid_reports"`
	AppealReports     int64 `json:"appeal_reports"`
}
