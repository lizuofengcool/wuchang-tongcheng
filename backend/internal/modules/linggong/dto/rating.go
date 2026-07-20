// Package dto 同城零工兼职数据传输对象 - 双向评价
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RatingInfo 评价详情响应
type RatingInfo struct {
	ID                uint       `json:"id"`
	RatingNo          string     `json:"rating_no"`
	LinggongID        uint       `json:"linggong_id"`
	TaskID            uint       `json:"task_id"`
	ApplicationID     uint       `json:"application_id"`
	ContractID        uint       `json:"contract_id"`
	PaymentID         uint       `json:"payment_id"`
	RaterType         string     `json:"rater_type"`
	RaterTypeText     string     `json:"rater_type_text"`
	RaterID           uint       `json:"rater_id"`
	RaterName         string     `json:"rater_name"`
	RaterAvatar       string     `json:"rater_avatar"`
	TargetType        string     `json:"target_type"`
	TargetTypeText    string     `json:"target_type_text"`
	TargetID          uint       `json:"target_id"`
	TargetName        string     `json:"target_name"`
	Rating            int        `json:"rating"`
	Content           string     `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          string     `json:"video_url"`
	IsAnonymous       bool       `json:"is_anonymous"`
	IsRecommended     string     `json:"is_recommended"`
	IsRecommendedText string     `json:"is_recommended_text"`
	Tags              interface{} `json:"tags"`
	DealAmount        float64    `json:"deal_amount"`
	WorkQuality       int        `json:"work_quality"`
	Punctuality       int        `json:"punctuality"`
	Communication     int        `json:"communication"`
	Attitude          int        `json:"attitude"`
	Professionalism   int        `json:"professionalism"`
	PaymentTimeliness int        `json:"payment_timeliness"`
	WorkEnvironment   int        `json:"work_environment"`
	SalaryMatch       int        `json:"salary_match"`
	Reply             string     `json:"reply"`
	ReplyAt           *time.Time `json:"reply_at"`
	AppendContent     string     `json:"append_content"`
	AppendImages      interface{} `json:"append_images"`
	AppendAt          *time.Time `json:"append_at"`
	LikeCount         int        `json:"like_count"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	RejectedReason    string     `json:"rejected_reason"`
	EvaluatedAt       *time.Time `json:"evaluated_at"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateRatingRequest 创建评价请求
type CreateRatingRequest struct {
	LinggongID        uint    `json:"linggong_id" binding:"required"`
	TaskID            uint    `json:"task_id"`
	ApplicationID     uint    `json:"application_id"`
	ContractID        uint    `json:"contract_id"`
	PaymentID         uint    `json:"payment_id"`
	RaterType         string  `json:"rater_type" binding:"omitempty,oneof=employer worker"`
	TargetType        string  `json:"target_type" binding:"omitempty,oneof=employer worker linggong task"`
	TargetID          uint    `json:"target_id" binding:"required"`
	TargetName        string  `json:"target_name" binding:"max=128"`
	Rating            int     `json:"rating" binding:"required,min=1,max=5"`
	Content           string  `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          string  `json:"video_url" binding:"max=255"`
	IsAnonymous       bool    `json:"is_anonymous"`
	IsRecommended     string  `json:"is_recommended" binding:"omitempty,oneof=yes no maybe"`
	Tags              interface{} `json:"tags"`
	DealAmount        float64 `json:"deal_amount"`
	WorkQuality       int     `json:"work_quality" binding:"min=1,max=5"`
	Punctuality       int     `json:"punctuality" binding:"min=1,max=5"`
	Communication     int     `json:"communication" binding:"min=1,max=5"`
	Attitude          int     `json:"attitude" binding:"min=1,max=5"`
	Professionalism   int     `json:"professionalism" binding:"min=1,max=5"`
	PaymentTimeliness int     `json:"payment_timeliness" binding:"min=1,max=5"`
	WorkEnvironment   int     `json:"work_environment" binding:"min=1,max=5"`
	SalaryMatch       int     `json:"salary_match" binding:"min=1,max=5"`
}

// UpdateRatingRequest 更新评价请求
type UpdateRatingRequest struct {
	Rating            *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Content           *string `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          *string `json:"video_url" binding:"omitempty,max=255"`
	IsAnonymous       *bool   `json:"is_anonymous"`
	IsRecommended     *string `json:"is_recommended" binding:"omitempty,oneof=yes no maybe"`
	Tags              interface{} `json:"tags"`
	DealAmount        *float64 `json:"deal_amount"`
	WorkQuality       *int    `json:"work_quality" binding:"omitempty,min=1,max=5"`
	Punctuality       *int    `json:"punctuality" binding:"omitempty,min=1,max=5"`
	Communication     *int    `json:"communication" binding:"omitempty,min=1,max=5"`
	Attitude          *int    `json:"attitude" binding:"omitempty,min=1,max=5"`
	Professionalism   *int    `json:"professionalism" binding:"omitempty,min=1,max=5"`
	PaymentTimeliness *int    `json:"payment_timeliness" binding:"omitempty,min=1,max=5"`
	WorkEnvironment   *int    `json:"work_environment" binding:"omitempty,min=1,max=5"`
	SalaryMatch       *int    `json:"salary_match" binding:"omitempty,min=1,max=5"`
}

// RatingReplyRequest 评价回复请求
type RatingReplyRequest struct {
	Reply string `json:"reply"`
}

// RatingAppendRequest 评价追评请求
type RatingAppendRequest struct {
	AppendContent string `json:"append_content"`
	AppendImages  interface{} `json:"append_images"`
}

// RatingListRequest 评价列表请求
type RatingListRequest struct {
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	TaskID        uint   `form:"task_id" json:"task_id"`
	ApplicationID uint   `form:"application_id" json:"application_id"`
	RaterType     string `form:"rater_type" json:"rater_type"`
	TargetType    string `form:"target_type" json:"target_type"`
	TargetID      uint   `form:"target_id" json:"target_id"`
	Rating        *int   `form:"rating" json:"rating"`
	IsRecommended string `form:"is_recommended" json:"is_recommended"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// RatingAdminListRequest 管理后台评价列表请求
type RatingAdminListRequest struct {
	RegionID      uint   `form:"region_id" json:"region_id"`
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	RaterID       uint   `form:"rater_id" json:"rater_id"`
	TargetType    string `form:"target_type" json:"target_type"`
	TargetID      uint   `form:"target_id" json:"target_id"`
	Rating        *int   `form:"rating" json:"rating"`
	Status        *int   `form:"status" json:"status"`
	Keyword       string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// RatingAuditRequest 评价审核请求
type RatingAuditRequest struct {
	Status         int    `json:"status" binding:"oneof=0 1 2 3"`
	RejectedReason string `json:"rejected_reason" binding:"max=500"`
}
