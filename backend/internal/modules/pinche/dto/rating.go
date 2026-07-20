// Package dto 同城拼车出行数据传输对象 - 评价
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// RatingInfo 评价详情响应
type RatingInfo struct {
	ID            uint       `json:"id"`
	RegionID      uint       `json:"region_id"`
	PincheID      uint       `json:"pinche_id"`
	BookingID     *uint      `json:"booking_id"`
	TripID        *uint      `json:"trip_id"`
	RaterID       uint       `json:"rater_id"`
	RaterName     string     `json:"rater_name"`
	RaterAvatar   string     `json:"rater_avatar"`
	RateeID       uint       `json:"ratee_id"`
	RateeName     string     `json:"ratee_name"`
	RateeAvatar   string     `json:"ratee_avatar"`
	RatingType    string     `json:"rating_type"`
	Rating        int        `json:"rating"`
	Content       string     `json:"content"`
	Images        interface{} `json:"images"`
	Tags          interface{} `json:"tags"`
	IsAnonymous   bool       `json:"is_anonymous"`
	Reply         string     `json:"reply"`
	ReplyAt       *time.Time `json:"reply_at"`
	LikeCount     int        `json:"like_count"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateRatingRequest 创建评价请求
type CreateRatingRequest struct {
	PincheID    uint        `json:"pinche_id" binding:"required"`
	BookingID   *uint       `json:"booking_id"`
	TripID      *uint       `json:"trip_id"`
	RatingType  string      `json:"rating_type" binding:"omitempty,oneof=passenger_to_driver driver_to_passenger"`
	Rating      int         `json:"rating" binding:"required,min=1,max=5"`
	Content     string      `json:"content" binding:"max=1000"`
	Images      interface{} `json:"images"`
	Tags        interface{} `json:"tags"`
	IsAnonymous bool        `json:"is_anonymous"`
}

// UpdateRatingRequest 更新评价请求
type UpdateRatingRequest struct {
	Content     *string     `json:"content"`
	Images      interface{} `json:"images"`
	Tags        interface{} `json:"tags"`
	IsAnonymous *bool       `json:"is_anonymous"`
	Status      *int        `json:"status"`
}

// RatingReplyRequest 评价回复请求
type RatingReplyRequest struct {
	Reply string `json:"reply" binding:"max=500"`
}

// RatingListRequest 评价列表查询请求
type RatingListRequest struct {
	PincheID   uint   `form:"pinche_id" json:"pinche_id"`
	BookingID  uint   `form:"booking_id" json:"booking_id"`
	RaterID    uint   `form:"rater_id" json:"rater_id"`
	RateeID    uint   `form:"ratee_id" json:"ratee_id"`
	RatingType string `form:"rating_type" json:"rating_type"`
	Rating     *int   `form:"rating" json:"rating"`
	Status     *int   `form:"status" json:"status"`
	Keyword    string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// RatingStatsResponse 评价统计响应
type RatingStatsResponse struct {
	TotalReviews int     `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
	HasReplyRate float64 `json:"has_reply_rate"`
}
