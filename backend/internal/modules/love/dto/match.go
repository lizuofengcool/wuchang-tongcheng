// Package dto love 相亲交友数据传输对象 - 匹配
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveMatchInfo 匹配记录响应
type LoveMatchInfo struct {
	ID              uint       `json:"id"`
	MatchNo         string     `json:"match_no"`
	UserIDA         uint       `json:"user_id_a"`
	UserIDB         uint       `json:"user_id_b"`
	LoveIDA         uint       `json:"love_id_a"`
	LoveIDB         uint       `json:"love_id_b"`
	MatchScore      float64    `json:"match_score"`
	MatchType       string     `json:"match_type"`
	InterestMatch   float64    `json:"interest_match"`
	PersonalityMatch float64   `json:"personality_match"`
	ValueMatch      float64    `json:"value_match"`
	LocationMatch   float64    `json:"location_match"`
	AgeMatch        float64    `json:"age_match"`
	MatchedAt       time.Time  `json:"matched_at"`
	ChatSessionID   uint       `json:"chat_session_id"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	LastMessageAt   *time.Time `json:"last_message_at"`
	LastMessageContent string  `json:"last_message_content"`
	LastMessageType string     `json:"last_message_type"`
	UnreadCount     int        `json:"unread_count"`
	DissolvedAt     *time.Time `json:"dissolved_at"`
	DissolveReason  string     `json:"dissolve_reason"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 对方信息（C 端展示用）
	PartnerNickname string `json:"partner_nickname,omitempty"`
	PartnerAvatar   string `json:"partner_avatar,omitempty"`
	PartnerGender   int    `json:"partner_gender,omitempty"`
	PartnerAge      int    `json:"partner_age,omitempty"`
}

// LoveMatchListRequest 匹配列表请求
type LoveMatchListRequest struct {
	Status  *int   `form:"status" json:"status"`
	MatchType string `form:"match_type" json:"match_type"`
	utils.Pagination
}

// DissolveMatchRequest 解除匹配请求
type DissolveMatchRequest struct {
	Reason string `json:"reason" binding:"max=255"`
}

// LoveMatchDetailRequest 匹配详情请求
type LoveMatchDetailRequest struct {
	MatchID uint `json:"match_id" binding:"required"`
}
