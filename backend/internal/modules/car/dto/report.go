// Package dto 同城车辆买卖数据传输对象 - 举报/评价
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ReportInfo 举报详情响应
type ReportInfo struct {
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
	EvidenceImages   interface{} `json:"evidence_images"`
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

// CreateReportRequest 创建举报请求
type CreateReportRequest struct {
	TargetType     string      `json:"target_type" binding:"required,oneof=car listing dealer user review"`
	TargetID       uint        `json:"target_id" binding:"required"`
	TargetUserID   uint        `json:"target_user_id" binding:"required"`
	ReportType     string      `json:"report_type" binding:"required,oneof=fake_car fraud porn infringement illegal_car false_mileage accident_hide other"`
	Reason         string      `json:"reason" binding:"required,max=500"`
	Description    string      `json:"description"`
	EvidenceImages interface{} `json:"evidence_images"`
}

// ProcessReportRequest 处理举报请求
type ProcessReportRequest struct {
	Status       int    `json:"status" binding:"oneof=1 2 3"`
	HandleResult string `json:"handle_result"`
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning ban24h ban7d ban30d ban_forever delete_car limit"`
	PenaltyUserID uint  `json:"penalty_user_id"`
}

// AppealReportRequest 申诉请求
type AppealReportRequest struct {
	AppealReason string `json:"appeal_reason" binding:"required,max=500"`
}

// ProcessAppealRequest 处理申诉请求
type ProcessAppealRequest struct {
	AppealResult string `json:"appeal_result" binding:"required"`
}

// ReportListRequest 举报列表请求
type ReportListRequest struct {
	TargetType string `form:"target_type" json:"target_type"`
	TargetID   uint   `form:"target_id" json:"target_id"`
	ReporterID uint   `form:"reporter_id" json:"reporter_id"`
	ReportType string `form:"report_type" json:"report_type"`
	Status     *int   `form:"status" json:"status"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ReviewInfo 评价详情响应
type ReviewInfo struct {
	ID               uint       `json:"id"`
	TargetType       string     `json:"target_type"`
	TargetID         uint       `json:"target_id"`
	ReviewerID       uint       `json:"reviewer_id"`
	ReviewerName     string     `json:"reviewer_name"`
	ReviewerAvatar   string     `json:"reviewer_avatar"`
	ReviewType       string     `json:"review_type"`
	Rating           int        `json:"rating"`
	Content          string     `json:"content"`
	Images           interface{} `json:"images"`
	VideoURL         string     `json:"video_url"`
	IsAnonymous      bool       `json:"is_anonymous"`
	IsRecommended    bool       `json:"is_recommended"`
	Tags             interface{} `json:"tags"`
	DealAmount       float64    `json:"deal_amount"`
	ExteriorRating   int        `json:"exterior_rating"`
	InteriorRating   int        `json:"interior_rating"`
	EngineRating     int        `json:"engine_rating"`
	PaperworkRating  int        `json:"paperwork_rating"`
	ServiceAttitude  int        `json:"service_attitude"`
	ProfessionalSkill int       `json:"professional_skill"`
	Reply            string     `json:"reply"`
	ReplyAt          *time.Time `json:"reply_at"`
	AppendContent    string     `json:"append_content"`
	AppendImages     interface{} `json:"append_images"`
	AppendAt         *time.Time `json:"append_at"`
	LikeCount        int        `json:"like_count"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	RegionID         uint       `json:"region_id"`
	CreatedAt        time.Time  `json:"created_at"`
}

// CreateReviewRequest 创建评价请求
type CreateReviewRequest struct {
	TargetType        string      `json:"target_type" binding:"omitempty,oneof=dealer car sales"`
	TargetID          uint        `json:"target_id" binding:"required"`
	ReviewType        string      `json:"review_type" binding:"omitempty,oneof=buyer seller"`
	Rating            int         `json:"rating" binding:"required,min=1,max=5"`
	Content           string      `json:"content" binding:"required"`
	Images            interface{} `json:"images"`
	VideoURL          string      `json:"video_url" binding:"max=255"`
	IsAnonymous       bool        `json:"is_anonymous"`
	IsRecommended     bool        `json:"is_recommended"`
	Tags              interface{} `json:"tags"`
	DealAmount        float64     `json:"deal_amount"`
	ExteriorRating    int         `json:"exterior_rating" binding:"omitempty,min=1,max=5"`
	InteriorRating    int         `json:"interior_rating" binding:"omitempty,min=1,max=5"`
	EngineRating      int         `json:"engine_rating" binding:"omitempty,min=1,max=5"`
	PaperworkRating   int         `json:"paperwork_rating" binding:"omitempty,min=1,max=5"`
	ServiceAttitude   int         `json:"service_attitude" binding:"omitempty,min=1,max=5"`
	ProfessionalSkill int         `json:"professional_skill" binding:"omitempty,min=1,max=5"`
}

// UpdateReviewRequest 更新评价请求
type UpdateReviewRequest struct {
	Content       *string `json:"content"`
	IsAnonymous   *bool   `json:"is_anonymous"`
	IsRecommended *bool   `json:"is_recommended"`
	Tags          interface{} `json:"tags"`
	Status        *int    `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

// ReviewReplyRequest 评价回复请求
type ReviewReplyRequest struct {
	Reply string `json:"reply" binding:"required,max=2000"`
}

// ReviewAppendRequest 评价追评请求
type ReviewAppendRequest struct {
	AppendContent string      `json:"append_content" binding:"required,max=2000"`
	AppendImages  interface{} `json:"append_images"`
}

// ReviewListRequest 评价列表请求
type ReviewListRequest struct {
	TargetType string `form:"target_type" json:"target_type"`
	TargetID   uint   `form:"target_id" json:"target_id"`
	ReviewerID uint   `form:"reviewer_id" json:"reviewer_id"`
	ReviewType string `form:"review_type" json:"review_type"`
	Rating     *int   `form:"rating" json:"rating"`
	Status     *int   `form:"status" json:"status"`
	HasReply   *bool  `form:"has_reply" json:"has_reply"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ReviewStatsResponse 评价统计响应
type ReviewStatsResponse struct {
	TotalReviews int     `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
	HasReplyRate float64 `json:"has_reply_rate"`
}
