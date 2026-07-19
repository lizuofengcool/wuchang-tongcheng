// Package dto 面试邀约相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘多轮面试 + Offer
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// InterviewResponse 面试邀约响应
type InterviewResponse struct {
	ID                  uint       `json:"id"`
	InterviewNo         string     `json:"interview_no"`
	ApplicationID       uint       `json:"application_id"`
	JobID               uint       `json:"job_id"`
	ApplicantID         uint       `json:"applicant_id"`
	RecruiterID         uint       `json:"recruiter_id"`
	CompanyID           uint       `json:"company_id"`
	Round               int        `json:"round"`
	InterviewType       string     `json:"interview_type"`
	InterviewTypeText   string     `json:"interview_type_text"`
	ScheduledAt         *time.Time `json:"scheduled_at"`
	DurationMinutes     int        `json:"duration_minutes"`
	Location            string     `json:"location"`
	OnlineURL           string     `json:"online_url"`
	OnlinePassword      string     `json:"online_password,omitempty"`
	InterviewerName     string     `json:"interviewer_name"`
	InterviewerPosition string     `json:"interviewer_position"`
	ContactPhone        string     `json:"contact_phone"`
	Status              int        `json:"status"`
	StatusText          string     `json:"status_text"`
	Result              string     `json:"result"`
	ResultText          string     `json:"result_text"`
	Feedback            string     `json:"feedback"`
	Rating              int        `json:"rating"`
	SalaryOffered       float64    `json:"salary_offered"`
	PositionOffered     string     `json:"position_offered"`
	EntryDate           *time.Time `json:"entry_date"`
	Attachments         []map[string]interface{} `json:"attachments"`
	ConfirmedAt         *time.Time `json:"confirmed_at"`
	AttendedAt          *time.Time `json:"attended_at"`
	CompletedAt         *time.Time `json:"completed_at"`
	CanceledAt          *time.Time `json:"canceled_at"`
	CanceledReason      string     `json:"canceled_reason"`
	RegionID            uint       `json:"region_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// 关联冗余
	JobTitle      string `json:"job_title,omitempty"`
	CompanyName   string `json:"company_name,omitempty"`
	ApplicantName string `json:"applicant_name,omitempty"`
}

// InterviewCreateRequest 创建面试邀约请求
type InterviewCreateRequest struct {
	ApplicationID       uint       `json:"application_id" binding:"required"`
	Round               int        `json:"round" binding:"gte=1"`
	InterviewType       string     `json:"interview_type" binding:"omitempty,oneof=onsite online phone video"`
	ScheduledAt         *time.Time `json:"scheduled_at"`
	DurationMinutes     int        `json:"duration_minutes" binding:"gte=0"`
	Location            string     `json:"location" binding:"max=255"`
	OnlineURL           string     `json:"online_url" binding:"max=500"`
	OnlinePassword      string     `json:"online_password" binding:"max=64"`
	InterviewerName     string     `json:"interviewer_name" binding:"max=50"`
	InterviewerPosition string     `json:"interviewer_position" binding:"max=64"`
	ContactPhone        string     `json:"contact_phone" binding:"max=20"`
}

// InterviewUpdateRequest 更新面试邀约
type InterviewUpdateRequest struct {
	ScheduledAt         *time.Time `json:"scheduled_at"`
	DurationMinutes     int        `json:"duration_minutes"`
	Location            string     `json:"location" binding:"max=255"`
	OnlineURL           string     `json:"online_url" binding:"max=500"`
	OnlinePassword      string     `json:"online_password" binding:"max=64"`
	InterviewerName     string     `json:"interviewer_name" binding:"max=50"`
	InterviewerPosition string     `json:"interviewer_position" binding:"max=64"`
	ContactPhone        string     `json:"contact_phone" binding:"max=20"`
}

// InterviewActionRequest 面试操作请求
type InterviewActionRequest struct {
	Action  string `json:"action" binding:"required,oneof=confirm reschedule attend complete cancel noshow"`
	Reason  string `json:"reason" binding:"max=500"`
	// reschedule
	NewScheduledAt *time.Time `json:"new_scheduled_at"`
}

// InterviewFeedbackRequest 面试反馈请求
type InterviewFeedbackRequest struct {
	Result          string  `json:"result" binding:"required,oneof=pending pass reject next_round offer"`
	Feedback        string  `json:"feedback" binding:"max=2000"`
	Rating          int     `json:"rating" binding:"gte=0,lte=5"`
	SalaryOffered   float64 `json:"salary_offered" binding:"gte=0"`
	PositionOffered string  `json:"position_offered" binding:"max=128"`
	EntryDate       *time.Time `json:"entry_date"`
}

// InterviewListQuery 面试列表查询
type InterviewListQuery struct {
	Role          string `form:"role" json:"role"`                   // applicant/recruiter/all
	Status        *int   `form:"status" json:"status"`
	Result        string `form:"result" json:"result"`
	JobID         uint   `form:"job_id" json:"job_id"`
	ApplicationID uint   `form:"application_id" json:"application_id"`
	CompanyID     uint   `form:"company_id" json:"company_id"`
	InterviewNo   string `form:"interview_no" json:"interview_no"`
	StartTime     *time.Time `form:"start_time" json:"start_time"`
	EndTime       *time.Time `form:"end_time" json:"end_time"`
	utils.Pagination
}

// InterviewStatsResponse 面试统计响应
type InterviewStatsResponse struct {
	TotalInterviews int64   `json:"total_interviews"`
	PendingCount    int64   `json:"pending_count"`
	ConfirmedCount  int64   `json:"confirmed_count"`
	CompletedCount  int64   `json:"completed_count"`
	CanceledCount   int64   `json:"canceled_count"`
	NoShowCount     int64   `json:"no_show_count"`
	PassRate        float64 `json:"pass_rate"`
	OfferCount      int64   `json:"offer_count"`
}
