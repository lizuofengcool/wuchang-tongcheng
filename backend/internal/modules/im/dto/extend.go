// Package dto IM 中台扩展数据传输对象
package dto

import "time"

// GroupInfo 群组信息
type GroupInfo struct {
	ID           uint      `json:"id"`
	GroupID      string    `json:"group_id"`
	GroupName    string    `json:"group_name"`
	Avatar       string    `json:"avatar"`
	Announcement string    `json:"announcement"`
	OwnerID      uint      `json:"owner_id"`
	MemberCount  int       `json:"member_count"`
	MaxMembers   int       `json:"max_members"`
	JoinType     int       `json:"join_type"`
	MuteAll      int       `json:"mute_all"`
	Status       int       `json:"status"`
	Extra        string    `json:"extra"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	GroupName  string `json:"group_name" binding:"required,max=128"`
	Avatar     string `json:"avatar" binding:"max=256"`
	Announcement string `json:"announcement"`
	MemberIDs  []uint `json:"member_ids" binding:"required,min=1"`
	MaxMembers int    `json:"max_members"`
	JoinType   int    `json:"join_type" binding:"oneof=0 1 2"`
}

// UpdateGroupRequest 更新群组请求
type UpdateGroupRequest struct {
	GroupName    string `json:"group_name" binding:"max=128"`
	Avatar       string `json:"avatar" binding:"max=256"`
	Announcement string `json:"announcement"`
	MaxMembers   int    `json:"max_members"`
	JoinType     int    `json:"join_type" binding:"omitempty,oneof=0 1 2"`
	MuteAll      int    `json:"mute_all" binding:"omitempty,oneof=0 1"`
}

// GroupMemberInfo 群成员信息
type GroupMemberInfo struct {
	ID        uint       `json:"id"`
	GroupID   string     `json:"group_id"`
	UserID    uint       `json:"user_id"`
	Role      string     `json:"role"`
	Nickname  string     `json:"nickname"`
	InviterID uint       `json:"inviter_id"`
	JoinedAt  time.Time  `json:"joined_at"`
	MuteUntil *time.Time `json:"mute_until"`
	Status    int        `json:"status"`
}

// AddGroupMembersRequest 添加群成员请求
type AddGroupMembersRequest struct {
	GroupID   string `json:"group_id" binding:"required"`
	UserIDs   []uint `json:"user_ids" binding:"required,min=1"`
}

// RemoveGroupMemberRequest 移除群成员请求
type RemoveGroupMemberRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	UserID  uint   `json:"user_id" binding:"required"`
}

// UserSettingInfo 用户IM设置信息
type UserSettingInfo struct {
	UserID                uint       `json:"user_id"`
	OnlineStatus          string     `json:"online_status"`
	AutoReply             string     `json:"auto_reply"`
	AutoReplyEnabled      int        `json:"auto_reply_enabled"`
	DoNotDisturb          int        `json:"do_not_disturb"`
	NotificationSound     int        `json:"notification_sound"`
	NotificationVibrate   int        `json:"notification_vibrate"`
	SaveToAlbum           int        `json:"save_to_album"`
	LastActiveAt          *time.Time `json:"last_active_at"`
	Extra                 string     `json:"extra"`
}

// UpdateUserSettingRequest 更新用户设置请求
type UpdateUserSettingRequest struct {
	OnlineStatus         string `json:"online_status" binding:"omitempty,oneof=online away busy offline"`
	AutoReply            string `json:"auto_reply" binding:"max=512"`
	AutoReplyEnabled     int    `json:"auto_reply_enabled" binding:"omitempty,oneof=0 1"`
	DoNotDisturb         int    `json:"do_not_disturb" binding:"omitempty,oneof=0 1"`
	NotificationSound    int    `json:"notification_sound" binding:"omitempty,oneof=0 1"`
	NotificationVibrate   int    `json:"notification_vibrate" binding:"omitempty,oneof=0 1"`
	SaveToAlbum          int    `json:"save_to_album" binding:"omitempty,oneof=0 1"`
	Extra                string `json:"extra"`
}

// SessionUserInfo 会话用户关联信息
type SessionUserInfo struct {
	ID         uint       `json:"id"`
	SessionID  string     `json:"session_id"`
	UserID     uint       `json:"user_id"`
	Role       string     `json:"role"`
	Nickname   string     `json:"nickname"`
	JoinedAt   time.Time  `json:"joined_at"`
	LastReadAt *time.Time `json:"last_read_at"`
	MuteUntil  *time.Time `json:"mute_until"`
	IsPinned   int        `json:"is_pinned"`
	IsMuted    int        `json:"is_muted"`
	Status     int        `json:"status"`
}

// RecallMessageRequest 撤回消息请求
type RecallMessageRequest struct {
	MessageID uint `json:"message_id" binding:"required"`
}

// IMStatisticsResponse IM 统计响应
type IMStatisticsResponse struct {
	TotalSessions      int64 `json:"total_sessions"`
	TotalMessages      int64 `json:"total_messages"`
	TotalNotifications int64 `json:"total_notifications"`
	UnreadCount        int64 `json:"unread_count"`
	TotalGroups        int64 `json:"total_groups"`
	TotalGroupMembers  int64 `json:"total_group_members"`
}
