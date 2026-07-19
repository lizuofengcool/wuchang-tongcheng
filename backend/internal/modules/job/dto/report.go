// Package dto 举报/评价相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘/看准
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 举报 =====

// ReportResponse 举报详情响应
type ReportResponse struct {
	ID               uint       `json:"id"`
	ReportNo         string     `json:"report_no"`
	TargetType       string     `json:"target_type"`
	TargetID         uint       `json:"target_id"`
	TargetUserID     uint       `json:"target_user_id"`
	ReporterID       uint       `json:"reporter_id"`
	ReporterName     string     `json:"reporter_name"`
	ReportedUserID   uint       `json:"reported_user_id"`
	ReportedUserName string     `json:"reported_user_name"`
	ReportType       string     `json:"report_type"`
	Reason           string     `json:"reason"`
	Description      string     `json:"description"`
	EvidenceImages   []string   `json:"evidence_images"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	HandlerID        uint       `json:"handler_id"`
	HandlerName      string     `json:"handler_name"`
	HandleResult     string     `json:"handle_result"`
	PenaltyType      string     `json:"penalty_type"`
	PenaltyUserID    uint       `json:"penalty_user_id"`
	SLADeadline      *time.Time `json:"sla_deadline"`
	HandledAt        *time.Time `json:"handled_at"`
	AppealReason     string     `json:"appeal_reason"`
	AppealedAt       *time.Time `json:"appealed_at"`
	AppealResult     string     `json:"appeal_result"`
	AppealHandlerID  uint       `json:"appeal_handler_id"`
	AppealHandledAt  *time.Time `json:"appeal_handled_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

// ReportCreateRequest 创建举报请求
type ReportCreateRequest struct {
	TargetType     string   `json:"target_type" binding:"required,oneof=job company resume recruiter review"`
	TargetID       uint     `json:"target_id" binding:"required"`
	TargetUserID   uint     `json:"target_user_id"`
	ReportType     string   `json:"report_type" binding:"required,oneof=fake scam porn prohibited infringement spam harassment privacy salary_anomaly other"`
	Reason         string   `json:"reason" binding:"required,max=500"`
	Description    string   `json:"description"`
	EvidenceImages []string `json:"evidence_images"`
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	Status     *int   `form:"status" json:"status"`
	ReportType string `form:"report_type" json:"report_type"`
	TargetType string `form:"target_type" json:"target_type"`
	TargetID   uint   `form:"target_id" json:"target_id"`
	ReporterID uint   `form:"reporter_id" json:"reporter_id"`
	utils.Pagination
}

// ReportProcessRequest 处理举报请求
type ReportProcessRequest struct {
	Status        int    `json:"status" binding:"oneof=1 2 3"` // 1处理中 2成立 3不成立
	HandleResult  string `json:"handle_result" binding:"max=2000"`
	PenaltyType   string `json:"penalty_type" binding:"omitempty,oneof=warning limit ban1d ban7d ban30d ban_forever close_job freeze_company"`
}

// ReportAppealRequest 举报申诉请求
type ReportAppealRequest struct {
	AppealReason string `json:"appeal_reason" binding:"required,max=2000"`
}

// ReportAppealProcessRequest 申诉处理请求
type ReportAppealProcessRequest struct {
	AppealResult string `json:"appeal_result" binding:"required,max=2000"`
}

// ===== 评价 =====

// ReviewResponse 公司评价响应
type ReviewResponse struct {
	ID            uint       `json:"id"`
	CompanyID     uint       `json:"company_id"`
	ReviewerID    uint       `json:"reviewer_id"`
	ReviewerName  string     `json:"reviewer_name"`
	ReviewerAvatar string    `json:"reviewer_avatar"`
	ReviewType    string     `json:"review_type"`
	Rating        int        `json:"rating"`
	Content       string     `json:"content"`
	Images        []string   `json:"images"`
	VideoURL      string     `json:"video_url"`
	IsAnonymous   bool       `json:"is_anonymous"`
	IsRecommended bool       `json:"is_recommended"`
	Tags          []string   `json:"tags"`
	Position      string     `json:"position"`
	Department    string     `json:"department"`
	WorkDuration  string     `json:"work_duration"`
	SalaryRange   string     `json:"salary_range"`
	Pros          string     `json:"pros"`
	Cons          string     `json:"cons"`
	Advice        string     `json:"advice"`
	Reply         string     `json:"reply"`
	ReplyAt       *time.Time `json:"reply_at"`
	AppendContent string     `json:"append_content"`
	AppendImages  []string   `json:"append_images"`
	AppendAt      *time.Time `json:"append_at"`
	LikeCount     int        `json:"like_count"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	RegionID      uint       `json:"region_id"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ReviewCreateRequest 创建评价请求
type ReviewCreateRequest struct {
	CompanyID     uint     `json:"company_id" binding:"required"`
	ReviewType    string   `json:"review_type" binding:"omitempty,oneof=employee former_employee interviewee candidate"`
	Rating        int      `json:"rating" binding:"required,gte=1,lte=5"`
	Content       string   `json:"content" binding:"required,max=2000"`
	Images        []string `json:"images"`
	VideoURL      string   `json:"video_url"`
	IsAnonymous   bool     `json:"is_anonymous"`
	IsRecommended bool     `json:"is_recommended"`
	Tags          []string `json:"tags"`
	Position      string   `json:"position"`
	Department    string   `json:"department"`
	WorkDuration  string   `json:"work_duration"`
	SalaryRange   string   `json:"salary_range"`
	Pros          string   `json:"pros"`
	Cons          string   `json:"cons"`
	Advice        string   `json:"advice"`
}

// ReviewUpdateRequest 更新评价请求
type ReviewUpdateRequest struct {
	Rating        *int     `json:"rating" binding:"omitempty,gte=1,lte=5"`
	Content       string   `json:"content" binding:"max=2000"`
	Images        []string `json:"images"`
	VideoURL      string   `json:"video_url"`
	IsAnonymous   *bool    `json:"is_anonymous"`
	IsRecommended *bool    `json:"is_recommended"`
	Tags          []string `json:"tags"`
	Pros          string   `json:"pros"`
	Cons          string   `json:"cons"`
	Advice        string   `json:"advice"`
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	CompanyID  uint   `form:"company_id" json:"company_id"`
	ReviewerID uint   `form:"reviewer_id" json:"reviewer_id"`
	ReviewType string `form:"review_type" json:"review_type"`
	Rating     *int   `form:"rating" json:"rating"`
	Status     *int   `form:"status" json:"status"`
	utils.Pagination
}

// ReviewReplyRequest 公司回复评价请求
type ReviewReplyRequest struct {
	Reply string `json:"reply" binding:"required,max=2000"`
}

// ReviewAppendRequest 追评请求
type ReviewAppendRequest struct {
	AppendContent string   `json:"append_content" binding:"required,max=2000"`
	AppendImages  []string `json:"append_images"`
}

// ReviewStatsResponse 评价统计响应
type ReviewStatsResponse struct {
	CompanyID        uint    `json:"company_id"`
	TotalReviews     int64   `json:"total_reviews"`
	AvgRating        float64 `json:"avg_rating"`
	GoodRate         float64 `json:"good_rate"`
	MediumRate       float64 `json:"medium_rate"`
	BadRate          float64 `json:"bad_rate"`
	RecommendCount   int64   `json:"recommend_count"`
}
