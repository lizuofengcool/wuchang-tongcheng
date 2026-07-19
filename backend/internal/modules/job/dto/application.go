// Package dto 投递相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘 9 状态机投递
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ApplicationResponse 投递记录响应
type ApplicationResponse struct {
	ID              uint       `json:"id"`
	ApplicationNo   string     `json:"application_no"`
	JobID           uint       `json:"job_id"`
	ResumeID        uint       `json:"resume_id"`
	ApplicantID     uint       `json:"applicant_id"`
	RecruiterID     uint       `json:"recruiter_id"`
	CompanyID       uint       `json:"company_id"`
	PositionName    string     `json:"position_name"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	Source          string     `json:"source"`
	CoverLetter     string     `json:"cover_letter"`
	Attachments     []map[string]interface{} `json:"attachments"`
	ReadAt          *time.Time `json:"read_at"`
	RepliedAt       *time.Time `json:"replied_at"`
	InterviewCount  int        `json:"interview_count"`
	OfferAt         *time.Time `json:"offer_at"`
	OfferAmount     float64    `json:"offer_amount"`
	RejectedReason  string     `json:"rejected_reason"`
	RejectedAt      *time.Time `json:"rejected_at"`
	WithdrawnAt     *time.Time `json:"withdrawn_at"`
	WithdrawnReason string     `json:"withdrawn_reason"`
	CompletedAt     *time.Time `json:"completed_at"`
	ExpiredAt       *time.Time `json:"expired_at"`
	SLADeadline     *time.Time `json:"sla_deadline"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 关联冗余（详情/列表展示用）
	JobTitle        string `json:"job_title,omitempty"`
	CompanyName     string `json:"company_name,omitempty"`
	CompanyLogo     string `json:"company_logo,omitempty"`
	ApplicantName   string `json:"applicant_name,omitempty"`
	ApplicantAvatar string `json:"applicant_avatar,omitempty"`
	RecruiterName   string `json:"recruiter_name,omitempty"`
}

// ApplicationCreateRequest 投递简历请求
type ApplicationCreateRequest struct {
	JobID       uint                   `json:"job_id" binding:"required"`
	ResumeID    uint                   `json:"resume_id"`
	CoverLetter string                 `json:"cover_letter"`
	Attachments []map[string]interface{} `json:"attachments"`
	Source      string                 `json:"source" binding:"omitempty,oneof=proactive recruiter recommend search ad"`
}

// ApplicationListQuery 投递记录列表查询
type ApplicationListQuery struct {
	Role        string `form:"role" json:"role"`               // applicant/recruiter/all
	Status      *int   `form:"status" json:"status"`
	JobID       uint   `form:"job_id" json:"job_id"`
	CompanyID   uint   `form:"company_id" json:"company_id"`
	ApplicationNo string `form:"application_no" json:"application_no"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ApplicationStatusUpdateRequest 投递状态变更请求
type ApplicationStatusUpdateRequest struct {
	Action string `json:"action" binding:"required,oneof=read unsuitable interview interviewing offer onboard withdraw reactivate"`
	Reason string `json:"reason" binding:"max=500"`
	OfferAmount float64 `json:"offer_amount"`
}

// ApplicationBatchActionRequest 批量操作投递记录
type ApplicationBatchActionRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Action string `json:"action" binding:"required,oneof=read unsuitable offer onboard"`
	Reason string `json:"reason" binding:"max=500"`
}
