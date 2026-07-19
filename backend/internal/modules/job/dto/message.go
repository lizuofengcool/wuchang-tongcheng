// Package dto 沟通消息相关 DTO
// 依据 v3.2.1 架构方案第四章：对标 BOSS直聘在线聊天
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// MessageResponse 消息响应
type MessageResponse struct {
	ID             uint       `json:"id"`
	ConversationID string     `json:"conversation_id"`
	JobID          uint       `json:"job_id"`
	ApplicationID  uint       `json:"application_id"`
	FromUserID     uint       `json:"from_user_id"`
	ToUserID       uint       `json:"to_user_id"`
	FromName       string     `json:"from_name"`
	FromAvatar     string     `json:"from_avatar"`
	ToName         string     `json:"to_name"`
	ToAvatar       string     `json:"to_avatar"`
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"`
	Attachments    []map[string]interface{} `json:"attachments"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at"`
	IsRecruiter    bool       `json:"is_recruiter"`
	IsSystem       bool       `json:"is_system"`
	Status         int        `json:"status"`
	Source         string     `json:"source"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

// MessageCreateRequest 发送消息请求
type MessageCreateRequest struct {
	ToUserID      uint                   `json:"to_user_id" binding:"required"`
	JobID         uint                   `json:"job_id"`
	ApplicationID uint                   `json:"application_id"`
	ConversationID string                 `json:"conversation_id"`
	Content       string                 `json:"content" binding:"max=5000"`
	MessageType   string                 `json:"message_type" binding:"omitempty,oneof=text image voice video file card system recruit resume interview offer greeting"`
	Attachments   []map[string]interface{} `json:"attachments"`
	IsRecruiter   bool                   `json:"is_recruiter"`
}

// MessageListQuery 消息列表查询
type MessageListQuery struct {
	ConversationID string `form:"conversation_id" json:"conversation_id"`
	JobID          uint   `form:"job_id" json:"job_id"`
	ApplicationID  uint   `form:"application_id" json:"application_id"`
	Role           string `form:"role" json:"role"` // from/to/all
	MessageType    string `form:"message_type" json:"message_type"`
	IsRead         *bool  `form:"is_read" json:"is_read"`
	IsSystem       *bool  `form:"is_system" json:"is_system"`
	Keyword        string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ConversationResponse 会话列表项
type ConversationResponse struct {
	ConversationID  string     `json:"conversation_id"`
	JobID           uint       `json:"job_id"`
	ApplicationID   uint       `json:"application_id"`
	FromUserID      uint       `json:"from_user_id"`
	ToUserID        uint       `json:"to_user_id"`
	FromName        string     `json:"from_name"`
	FromAvatar      string     `json:"from_avatar"`
	ToName          string     `json:"to_name"`
	ToAvatar        string     `json:"to_avatar"`
	LastContent     string     `json:"last_content"`
	LastMessageType string     `json:"last_message_type"`
	LastMessageAt   time.Time  `json:"last_message_at"`
	UnreadCount     int64      `json:"unread_count"`
}

// MessageBatchDeleteRequest 批量删除消息请求
type MessageBatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
