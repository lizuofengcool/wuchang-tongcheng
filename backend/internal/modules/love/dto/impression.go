// Package dto love 相亲交友数据传输对象 - 印象标签
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveImpressionInfo 印象标签响应
type LoveImpressionInfo struct {
	ID                uint      `json:"id"`
	LoveID            uint      `json:"love_id"`
	UserID            uint      `json:"user_id"`
	FromUserID        uint      `json:"from_user_id"`
	FromUserNickname  string    `json:"from_user_nickname"`
	FromUserAvatar    string    `json:"from_user_avatar"`
	Tag               string    `json:"tag"`
	Content           string    `json:"content"`
	Anonymous         bool      `json:"anonymous"`
	IsAnonymous       bool      `json:"is_anonymous"`
	MatchID           uint      `json:"match_id"`
	Status            int       `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

// CreateLoveImpressionRequest 创建印象标签请求
type CreateLoveImpressionRequest struct {
	TargetUserID uint   `json:"target_user_id" binding:"required"`
	TargetLoveID uint   `json:"target_love_id" binding:"required"`
	Tag          string `json:"tag" binding:"required,max=32"`
	Content      string `json:"content" binding:"max=500"`
	Anonymous    bool   `json:"anonymous"`
	MatchID      uint   `json:"match_id"`
}

// LoveImpressionListRequest 印象标签列表请求
type LoveImpressionListRequest struct {
	LoveID     uint   `form:"love_id" json:"love_id"`
	UserID     uint   `form:"user_id" json:"user_id"`
	FromUserID uint   `form:"from_user_id" json:"from_user_id"`
	Tag        string `form:"tag" json:"tag"`
	utils.Pagination
}

// LoveImpressionStatsResponse 印象统计响应
type LoveImpressionStatsResponse struct {
	TotalImpressions int64            `json:"total_impressions"`
	TopTags          []LoveImpressionTagStat `json:"top_tags"`
}

// LoveImpressionTagStat 印象标签统计
type LoveImpressionTagStat struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}
