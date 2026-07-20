// Package dto love 相亲交友数据传输对象 - 聊天会话
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveChatSessionInfo 聊天会话响应
type LoveChatSessionInfo struct {
	ID                uint       `json:"id"`
	SessionNo         string     `json:"session_no"`
	MatchID           uint       `json:"match_id"`
	UserIDA           uint       `json:"user_id_a"`
	UserIDB           uint       `json:"user_id_b"`
	LoveIDA           uint       `json:"love_id_a"`
	LoveIDB           uint       `json:"love_id_b"`
	NicknameA         string     `json:"nickname_a"`
	NicknameB         string     `json:"nickname_b"`
	AvatarA           string     `json:"avatar_a"`
	AvatarB           string     `json:"avatar_b"`
	LastMessageID     uint       `json:"last_message_id"`
	LastMessageContent string    `json:"last_message_content"`
	LastMessageType   string     `json:"last_message_type"`
	LastMessageAt     *time.Time `json:"last_message_at"`
	LastSenderID      uint       `json:"last_sender_id"`
	UnreadCount       int        `json:"unread_count"`
	Muted             bool       `json:"muted"`
	Pinned            bool       `json:"pinned"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	MessageCount      int        `json:"message_count"`
	GiftCount         int        `json:"gift_count"`
	DissolvedAt       *time.Time `json:"dissolved_at"`
	DissolveReason    string     `json:"dissolve_reason"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// 对方信息（C 端展示用）
	PartnerNickname string `json:"partner_nickname,omitempty"`
	PartnerAvatar   string `json:"partner_avatar,omitempty"`
	PartnerUserID   uint   `json:"partner_user_id,omitempty"`
	PartnerLoveID   uint   `json:"partner_love_id,omitempty"`
}

// LoveChatSessionListRequest 会话列表请求
type LoveChatSessionListRequest struct {
	Status *int `form:"status" json:"status"`
	utils.Pagination
}

// LoveChatSessionActionRequest 会话操作请求
type LoveChatSessionActionRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Action string `json:"action" binding:"required,oneof=mute unmute pin unpin delete dissolve"`
	Reason string `json:"reason" binding:"max=255"`
}

// LoveChatReadRequest 会话已读请求
type LoveChatReadRequest struct {
	ID uint `json:"id" binding:"required"`
}

// LoveChatMessageInfo 聊天消息响应（轻量级，仅会话维度展示用）
type LoveChatMessageInfo struct {
	ID         uint      `json:"id"`
	SessionID  uint      `json:"session_id"`
	SenderID   uint      `json:"sender_id"`
	Type       string    `json:"type"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	IsRecalled bool      `json:"is_recalled"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoveChatMessageListRequest 消息列表请求
type LoveChatMessageListRequest struct {
	SessionID uint `form:"session_id" json:"session_id" binding:"required"`
	utils.Pagination
}
