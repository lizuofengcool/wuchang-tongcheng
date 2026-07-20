// Package dto love 相亲交友数据传输对象 - 推荐
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveRecommendationInfo 推荐响应
type LoveRecommendationInfo struct {
	ID                uint       `json:"id"`
	UserID            uint       `json:"user_id"`
	LoveID            uint       `json:"love_id"`
	TargetUserID      uint       `json:"target_user_id"`
	TargetLoveID      uint       `json:"target_love_id"`
	TargetNickname    string     `json:"target_nickname"`
	TargetAvatar      string     `json:"target_avatar"`
	TargetGender      int        `json:"target_gender"`
	TargetAge         int        `json:"target_age"`
	TargetDistance    float64    `json:"target_distance"`
	TargetVerified    bool       `json:"target_verified"`
	TargetMemberLevel int        `json:"target_member_level"`
	RecType           string     `json:"rec_type"`
	Source            string     `json:"source"`
	Score             float64    `json:"score"`
	InterestMatch     float64    `json:"interest_match"`
	PersonalityMatch  float64    `json:"personality_match"`
	ValueMatch        float64    `json:"value_match"`
	LocationMatch     float64    `json:"location_match"`
	AgeMatch          float64    `json:"age_match"`
	Reason            string     `json:"reason"`
	IsViewed          bool       `json:"is_viewed"`
	IsLiked           bool       `json:"is_liked"`
	IsDisliked        bool       `json:"is_disliked"`
	IsSuperLiked      bool       `json:"is_super_liked"`
	IsSkipped         bool       `json:"is_skipped"`
	IsDismissed       bool       `json:"is_dismissed"`
	Status            int        `json:"status"`
	ExpiredAt         *time.Time `json:"expired_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// LoveRecommendationListRequest 推荐列表请求
type LoveRecommendationListRequest struct {
	RecType string `form:"rec_type" json:"rec_type"`
	Source  string `form:"source" json:"source"`
	Status  *int   `form:"status" json:"status"`
	utils.Pagination
}

// GenerateLoveRecommendationsRequest 生成推荐请求
type GenerateLoveRecommendationsRequest struct {
	RecType string `json:"rec_type" binding:"omitempty,oneof=daily nearby same_city same_hometown interest soulmate new_user"`
	Count   int    `json:"count" binding:"omitempty,min=1,max=100"`
}

// LoveRecommendationActionRequest 推荐操作请求
type LoveRecommendationActionRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Action string `json:"action" binding:"required,oneof=view like dislike super_like skip dismiss"`
}

// LoveRecommendationStatsResponse 推荐统计响应
type LoveRecommendationStatsResponse struct {
	TotalRecommendations int64 `json:"total_recommendations"`
	TodayRecommendations int64 `json:"today_recommendations"`
	ViewedCount          int64 `json:"viewed_count"`
	LikedCount           int64 `json:"liked_count"`
	DislikedCount        int64 `json:"disliked_count"`
	SuperLikedCount      int64 `json:"super_liked_count"`
	SkippedCount         int64 `json:"skipped_count"`
	DismissedCount       int64 `json:"dismissed_count"`
}
