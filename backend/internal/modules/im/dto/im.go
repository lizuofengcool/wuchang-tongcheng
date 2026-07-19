// Package dto IM 消息中台精简版数据传输对象
package dto

import "time"

// SessionInfo 会话信息
type SessionInfo struct {
	ID            uint       `json:"id"`
	SessionID     string     `json:"session_id"`
	SessionType   string     `json:"session_type"`
	Participants  string     `json:"participants"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	UnreadCount   string     `json:"unread_count"`
	Status        int        `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	SessionType  string `json:"session_type" binding:"required,oneof=private group"`
	Participants []uint `json:"participants" binding:"required,min=2"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	MsgType   string `json:"msg_type" binding:"required,oneof=text image voice video card"`
	Content   string `json:"content" binding:"required"`
	Extra     string `json:"extra"` // JSON 字符串
}

// MessageInfo 消息信息
type MessageInfo struct {
	ID         uint      `json:"id"`
	SessionID  string    `json:"session_id"`
	SenderID   uint      `json:"sender_id"`
	MsgType    string    `json:"msg_type"`
	Content    string    `json:"content"`
	Extra      string    `json:"extra"`
	ReadStatus string    `json:"read_status"`
	IsRecalled bool      `json:"is_recalled"`
	CreatedAt  time.Time `json:"created_at"`
}

// NotificationInfo 系统通知信息
type NotificationInfo struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	NotifyType string     `json:"notify_type"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	JumpURL    string     `json:"jump_url"`
	Extra      string     `json:"extra"`
	IsRead     bool       `json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// PushNotificationRequest 推送系统通知请求（其他模块调用）
type PushNotificationRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`
	NotifyType string `json:"notify_type" binding:"required"`
	Title      string `json:"title" binding:"required,max=128"`
	Content    string `json:"content" binding:"required"`
	JumpURL    string `json:"jump_url" binding:"max=256"`
	Extra      string `json:"extra"`
}

// MarkReadRequest 标记已读请求
type MarkReadRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// BindPrivacyNumberRequest 绑定隐私号码请求
type BindPrivacyNumberRequest struct {
	UserIDA   uint   `json:"user_id_a" binding:"required"`
	UserIDB   uint   `json:"user_id_b" binding:"required"`
	RealNoA   string `json:"real_no_a" binding:"required,max=32"`
	RealNoB   string `json:"real_no_b" binding:"required,max=32"`
	BizModule string `json:"biz_module" binding:"max=32"`
	BizID     string `json:"biz_id" binding:"max=128"`
}

// UnbindPrivacyNumberRequest 解绑隐私号码请求
type UnbindPrivacyNumberRequest struct {
	PrivacyNo string `json:"privacy_no" binding:"required"`
}

// PrivacyNumberInfo 隐私号码信息
type PrivacyNumberInfo struct {
	ID          uint       `json:"id"`
	PrivacyNo   string     `json:"privacy_no"`
	RealNoA     string     `json:"real_no_a"`
	RealNoB     string     `json:"real_no_b"`
	UserIDA     uint       `json:"user_id_a"`
	UserIDB     uint       `json:"user_id_b"`
	BizModule   string     `json:"biz_module"`
	BizID       string     `json:"biz_id"`
	CallRecords string     `json:"call_records"`
	BoundAt     time.Time  `json:"bound_at"`
	UnboundAt   *time.Time `json:"unbound_at"`
	Status      int        `json:"status"`
}
