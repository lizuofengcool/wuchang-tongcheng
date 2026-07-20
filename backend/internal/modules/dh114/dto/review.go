// Package dto 同城114数据传输对象 - 评价
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ReviewInfo 评价详情响应
type ReviewInfo struct {
	ID                uint       `json:"id"`
	ReviewNo          string     `json:"review_no"`
	Dh114ID           uint       `json:"dh114_id"`
	BusinessID        uint       `json:"business_id"`
	ReviewerID        uint       `json:"reviewer_id"`
	ReviewerName      string     `json:"reviewer_name"`
	ReviewerAvatar    string     `json:"reviewer_avatar"`
	Rating            int        `json:"rating"`
	TasteRating      int        `json:"taste_rating"`
	ServiceRating    int        `json:"service_rating"`
	EnvironmentRating int       `json:"environment_rating"`
	Content           string     `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          string     `json:"video_url"`
	VideoCover        string     `json:"video_cover"`
	Tags              interface{} `json:"tags"`
	Reply             string     `json:"reply"`
	RepliedAt         *time.Time `json:"replied_at"`
	HasReply          bool       `json:"has_reply"`
	LikeCount         int        `json:"like_count"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	AuditStatus       int        `json:"audit_status"`
	AuditReason       string     `json:"audit_reason"`
	OrderID           uint       `json:"order_id"`
	ConsumedAt        *time.Time `json:"consumed_at"`
	ReviewType        string     `json:"review_type"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateReviewRequest 创建评价请求
type CreateReviewRequest struct {
	Dh114ID           uint   `json:"dh114_id" binding:"required"`
	Rating            int    `json:"rating" binding:"required,min=1,max=5"`
	TasteRating      int    `json:"taste_rating" binding:"omitempty,min=1,max=5"`
	ServiceRating    int    `json:"service_rating" binding:"omitempty,min=1,max=5"`
	EnvironmentRating int   `json:"environment_rating" binding:"omitempty,min=1,max=5"`
	Content           string `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          string `json:"video_url" binding:"max=255"`
	VideoCover        string `json:"video_cover" binding:"max=255"`
	Tags              interface{} `json:"tags"`
	OrderID           uint   `json:"order_id"`
	ConsumedAt        *time.Time `json:"consumed_at"`
	ReviewType        string `json:"review_type" binding:"omitempty,oneof=general order visit"`
}

// UpdateReviewRequest 更新评价请求
type UpdateReviewRequest struct {
	Rating            *int   `json:"rating" binding:"omitempty,min=1,max=5"`
	TasteRating      *int   `json:"taste_rating" binding:"omitempty,min=1,max=5"`
	ServiceRating    *int   `json:"service_rating" binding:"omitempty,min=1,max=5"`
	EnvironmentRating *int  `json:"environment_rating" binding:"omitempty,min=1,max=5"`
	Content           *string `json:"content"`
	Images            interface{} `json:"images"`
	VideoURL          *string `json:"video_url" binding:"max=255"`
	VideoCover        *string `json:"video_cover" binding:"max=255"`
	Tags              interface{} `json:"tags"`
}

// ReviewListRequest 评价列表请求
type ReviewListRequest struct {
	Dh114ID    uint   `form:"dh114_id" json:"dh114_id"`
	ReviewerID uint   `form:"reviewer_id" json:"reviewer_id"`
	Rating     *int   `form:"rating" json:"rating"`
	Status     *int   `form:"status" json:"status"`
	HasReply   *bool  `form:"has_reply" json:"has_reply"`
	Keyword    string `form:"keyword" json:"keyword"`
	Sort       string `form:"sort" json:"sort"`
	utils.Pagination
}

// ReviewReplyRequest 商家回复评价请求
type ReviewReplyRequest struct {
	Content string `json:"content" binding:"required,max=500"`
	Images  interface{} `json:"images"`
}

// ReviewAppendRequest 追加评价请求
type ReviewAppendRequest struct {
	Content string `json:"content" binding:"required,max=1000"`
	Images  interface{} `json:"images"`
}

// ReviewStatsRequest 评价统计请求
type ReviewStatsRequest struct {
	Dh114ID uint `form:"dh114_id" json:"dh114_id"`
}
