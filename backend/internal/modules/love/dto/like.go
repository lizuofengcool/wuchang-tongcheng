// Package dto love 相亲交友数据传输对象 - 喜欢/不喜欢/心动信号
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveLikeInfo 喜欢记录响应
type LoveLikeInfo struct {
	ID             uint       `json:"id"`
	UserID         uint       `json:"user_id"`
	LoveID         uint       `json:"love_id"`
	TargetUserID   uint       `json:"target_user_id"`
	TargetLoveID   uint       `json:"target_love_id"`
	TargetNickname string     `json:"target_nickname"`
	TargetAvatar   string     `json:"target_avatar"`
	TargetGender   int        `json:"target_gender"`
	Action         string     `json:"action"`
	SuperLike      bool       `json:"super_like"`
	MatchScore     float64    `json:"match_score"`
	Source         string     `json:"source"`
	IsMatched      bool       `json:"is_matched"`
	MatchID        uint       `json:"match_id"`
	MatchedAt      *time.Time `json:"matched_at"`
	Status         int        `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateLoveLikeRequest 喜欢/不喜欢/心动信号请求
type CreateLoveLikeRequest struct {
	TargetUserID uint   `json:"target_user_id" binding:"required"`
	TargetLoveID uint   `json:"target_love_id" binding:"required"`
	Action       string `json:"action" binding:"required,oneof=like dislike skip super"`
	Source       string `json:"source" binding:"omitempty,oneof=recommend nearby search story match"`
}

// LoveLikeListRequest 喜欢列表请求
type LoveLikeListRequest struct {
	UserID       uint   `form:"user_id" json:"user_id"`
	TargetUserID uint   `form:"target_user_id" json:"target_user_id"`
	Action       string `form:"action" json:"action"`
	SuperLike    *bool  `form:"super_like" json:"super_like"`
	IsMatched    *bool  `form:"is_matched" json:"is_matched"`
	utils.Pagination
}

// UndoLoveLikeRequest 撤销喜欢请求
type UndoLoveLikeRequest struct {
	LikeID uint   `json:"like_id" binding:"required"`
	Reason string `json:"reason" binding:"max=255"`
}

// LoveLikeTodayStatsResponse 今日喜欢统计
type LoveLikeTodayStatsResponse struct {
	TotalLikes      int `json:"total_likes"`
	TotalDislikes   int `json:"total_dislikes"`
	TotalSuperLikes int `json:"total_super_likes"`
	TotalSkips      int `json:"total_skips"`
	TotalMatches    int `json:"total_matches"`
	LimitLikes      int `json:"limit_likes"`
	LimitSuperLikes int `json:"limit_super_likes"`
	RemainingLikes  int `json:"remaining_likes"`
	RemainingSuperLikes int `json:"remaining_super_likes"`
}
