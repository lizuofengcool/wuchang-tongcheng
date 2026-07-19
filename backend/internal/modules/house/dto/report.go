// Package dto 举报 + 评价 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package dto

import (
	"time"

	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 举报 =====

// ReportResponse 举报工单响应
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
	EvidenceImages   []model.EvidenceImage `json:"evidence_images"`
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
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ReportCreateRequest 创建举报请求
type ReportCreateRequest struct {
	TargetType     string                 `json:"target_type" binding:"required,oneof=house listing agent community review"`
	TargetID       uint                   `json:"target_id" binding:"required"`
	TargetUserID   uint                   `json:"target_user_id"`
	ReportType     string                 `json:"report_type" binding:"required,oneof=fake_house price_fraud sold_still_listed contact_invalid scam porn illegal infringement other"`
	Reason         string                 `json:"reason" binding:"required,max=500"`
	Description    string                 `json:"description"`
	EvidenceImages []model.EvidenceImage  `json:"evidence_images"`
}

// ReportProcessRequest 处理举报请求
type ReportProcessRequest struct {
	Status       int    `json:"status" binding:"oneof=1 2 3"` // 1处理中 2已处理 3驳回
	HandleResult string `json:"handle_result" binding:"max=500"`
	PenaltyType  string `json:"penalty_type" binding:"omitempty,oneof=warning remove ban7d ban30d ban_forever"`
}

// ReportAppealRequest 申诉请求
type ReportAppealRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// ReportAppealHandleRequest 申诉处理请求
type ReportAppealHandleRequest struct {
	AppealResult string `json:"appeal_result" binding:"required,max=500"`
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	TargetType   string `form:"target_type" json:"target_type"`
	TargetID     uint   `form:"target_id" json:"target_id"`
	ReporterID   uint   `form:"reporter_id" json:"reporter_id"`
	ReportType   string `form:"report_type" json:"report_type"`
	Status       *int   `form:"status" json:"status"`
	PenaltyType  string `form:"penalty_type" json:"penalty_type"`
	Keyword      string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ReportAdminListQuery 管理后台举报列表查询
type ReportAdminListQuery struct {
	TargetType  string `form:"target_type" json:"target_type"`
	ReportType  string `form:"report_type" json:"report_type"`
	Status      *int   `form:"status" json:"status"`
	HandlerID   uint   `form:"handler_id" json:"handler_id"`
	PenaltyType string `form:"penalty_type" json:"penalty_type"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ===== 评价 =====

// ReviewResponse 评价响应
type ReviewResponse struct {
	ID               uint       `json:"id"`
	TargetType       string     `json:"target_type"`
	TargetID         uint       `json:"target_id"`
	ReviewerID       uint       `json:"reviewer_id"`
	ReviewerName     string     `json:"reviewer_name"`
	ReviewerAvatar   string     `json:"reviewer_avatar"`
	ReviewType       string     `json:"review_type"`
	Rating           int        `json:"rating"`
	Content          string     `json:"content"`
	Images           []model.ReviewImage `json:"images"`
	VideoURL         string     `json:"video_url"`
	IsAnonymous      bool       `json:"is_anonymous"`
	IsRecommended    bool       `json:"is_recommended"`
	Tags             []model.HouseTagItem `json:"tags"`
	DealAmount       float64    `json:"deal_amount"`
	ServiceAttitude  int        `json:"service_attitude"`
	ProfessionalSkill int       `json:"professional_skill"`
	Reply            string     `json:"reply"`
	ReplyAt          *time.Time `json:"reply_at"`
	AppendContent    string     `json:"append_content"`
	AppendImages     []model.ReviewImage `json:"append_images"`
	AppendAt         *time.Time `json:"append_at"`
	LikeCount        int        `json:"like_count"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	RegionID         uint       `json:"region_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	HasLiked         bool       `json:"has_liked,omitempty"` // 当前用户是否已点赞
}

// ReviewCreateRequest 创建评价请求
type ReviewCreateRequest struct {
	TargetType        string                 `json:"target_type" binding:"required,oneof=agent community house"`
	TargetID          uint                   `json:"target_id" binding:"required"`
	ReviewType        string                 `json:"review_type" binding:"omitempty,oneof=tenant buyer seller landlord"`
	Rating            int                    `json:"rating" binding:"required,gte=1,lte=5"`
	Content           string                 `json:"content" binding:"required,max=1000"`
	Images            []model.ReviewImage    `json:"images"`
	VideoURL          string                 `json:"video_url" binding:"max=255"`
	IsAnonymous       bool                   `json:"is_anonymous"`
	IsRecommended     bool                   `json:"is_recommended"`
	Tags              []model.HouseTagItem   `json:"tags"`
	DealAmount        float64                `json:"deal_amount" binding:"gte=0"`
	ServiceAttitude   int                    `json:"service_attitude" binding:"gte=1,lte=5"`
	ProfessionalSkill int                    `json:"professional_skill" binding:"gte=1,lte=5"`
}

// ReviewReplyRequest 评价回复请求
type ReviewReplyRequest struct {
	Reply string `json:"reply" binding:"required,max=500"`
}

// ReviewAppendRequest 追评请求
type ReviewAppendRequest struct {
	AppendContent string                `json:"append_content" binding:"required,max=500"`
	AppendImages  []model.ReviewImage   `json:"append_images"`
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	TargetType    string `form:"target_type" json:"target_type"`
	TargetID      uint   `form:"target_id" json:"target_id"`
	ReviewerID    uint   `form:"reviewer_id" json:"reviewer_id"`
	ReviewType    string `form:"review_type" json:"review_type"`
	Rating        *int   `form:"rating" json:"rating"`
	IsRecommended *bool  `form:"is_recommended" json:"is_recommended"`
	Status        *int   `form:"status" json:"status"`
	Sort          string `form:"sort" json:"sort"` // latest/rating/useful
	utils.Pagination
}

// ReviewAdminListQuery 管理后台评价列表查询
type ReviewAdminListQuery struct {
	RegionID   uint   `form:"region_id" json:"region_id"`
	TargetType string `form:"target_type" json:"target_type"`
	TargetID   uint   `form:"target_id" json:"target_id"`
	ReviewerID uint   `form:"reviewer_id" json:"reviewer_id"`
	Rating     *int   `form:"rating" json:"rating"`
	Status     *int   `form:"status" json:"status"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}
