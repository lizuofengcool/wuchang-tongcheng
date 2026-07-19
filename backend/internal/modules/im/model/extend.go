// Package model IM 中台扩展数据模型
// 依据 013_im_full.sql：消息已读/会话用户/用户设置/群组/群成员
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 群成员角色 ===
const (
	GroupRoleOwner  = "owner"
	GroupRoleAdmin  = "admin"
	GroupRoleMember = "member"
)

// === 加入类型 ===
const (
	JoinTypeAny     = 0 // 任意加入
	JoinTypeAudit   = 1 // 需审核
	JoinTypeInvite  = 2 // 仅邀请
)

// === 用户在线状态 ===
const (
	OnlineStatusOnline  = "online"
	OnlineStatusAway    = "away"
	OnlineStatusBusy    = "busy"
	OnlineStatusOffline = "offline"
)

// MessageRead 消息已读记录
type MessageRead struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	RegionID  uint      `gorm:"index;not null;default:1" json:"region_id"`
	MessageID uint      `gorm:"not null;uniqueIndex:uk_im_msg_reads,priority:1" json:"message_id"`
	SessionID string    `gorm:"size:64;not null;default:''" json:"session_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_im_msg_reads,priority:2;index" json:"user_id"`
	ReadAt    time.Time `gorm:"not null;default:now()" json:"read_at"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

// TableName 表名
func (MessageRead) TableName() string { return "im_message_reads" }

// SessionUser 会话用户关联
type SessionUser struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	RegionID    uint       `gorm:"index;not null;default:1" json:"region_id"`
	SessionID   string     `gorm:"size:64;not null;uniqueIndex:uk_im_session_users,priority:1" json:"session_id"`
	UserID      uint       `gorm:"not null;index;uniqueIndex:uk_im_session_users,priority:2" json:"user_id"`
	Role        string     `gorm:"size:16;not null;default:'member'" json:"role"`
	Nickname    string     `gorm:"size:64;not null;default:''" json:"nickname"`
	JoinedAt    time.Time  `gorm:"not null;default:now()" json:"joined_at"`
	LastReadAt  *time.Time `gorm:"index" json:"last_read_at"`
	MuteUntil  *time.Time `gorm:"index" json:"mute_until"`
	IsPinned    int        `gorm:"default:0" json:"is_pinned"`
	IsMuted     int        `gorm:"default:0" json:"is_muted"`
	Status      int        `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (SessionUser) TableName() string { return "im_session_users" }

// UserSetting 用户 IM 设置
type UserSetting struct {
	ID                   uint       `gorm:"primarykey" json:"id"`
	RegionID             uint       `gorm:"index;not null;default:1" json:"region_id"`
	UserID               uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	OnlineStatus         string     `gorm:"size:16;not null;default:'online'" json:"online_status"`
	AutoReply            string     `gorm:"type:text;not null;default:''" json:"auto_reply"`
	AutoReplyEnabled     int        `gorm:"default:0" json:"auto_reply_enabled"`
	DoNotDisturb         int        `gorm:"default:0" json:"do_not_disturb"`
	NotificationSound    int        `gorm:"default:1" json:"notification_sound"`
	NotificationVibrate  int        `gorm:"default:1" json:"notification_vibrate"`
	SaveToAlbum          int        `gorm:"default:0" json:"save_to_album"`
	LastActiveAt         *time.Time `gorm:"index" json:"last_active_at"`
	Extra                string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`
	CreatedAt            time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (UserSetting) TableName() string { return "im_user_settings" }

// Group IM 群组
type Group struct {
	database.RegionBaseModel

	GroupID      string `gorm:"size:64;not null;uniqueIndex" json:"group_id"`
	GroupName    string `gorm:"size:128;not null;default:''" json:"group_name"`
	Avatar       string `gorm:"size:256;not null;default:''" json:"avatar"`
	Announcement string `gorm:"type:text;not null;default:''" json:"announcement"`
	OwnerID      uint   `gorm:"index;not null" json:"owner_id"`
	MemberCount  int    `gorm:"default:0" json:"member_count"`
	MaxMembers   int    `gorm:"default:500" json:"max_members"`
	JoinType     int    `gorm:"default:0" json:"join_type"`
	MuteAll      int    `gorm:"default:0" json:"mute_all"`
	Status       int    `gorm:"default:1;index" json:"status"`
	Extra        string `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`
}

// TableName 表名
func (Group) TableName() string { return "im_groups" }

// GroupMember 群成员
type GroupMember struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	RegionID  uint       `gorm:"index;not null;default:1" json:"region_id"`
	GroupID   string     `gorm:"size:64;not null;uniqueIndex:uk_im_group_members,priority:1" json:"group_id"`
	UserID    uint       `gorm:"not null;index;uniqueIndex:uk_im_group_members,priority:2" json:"user_id"`
	Role      string     `gorm:"size:16;not null;default:'member'" json:"role"`
	Nickname  string     `gorm:"size:64;not null;default:''" json:"nickname"`
	InviterID uint       `gorm:"default:0" json:"inviter_id"`
	JoinedAt  time.Time  `gorm:"not null;default:now()" json:"joined_at"`
	MuteUntil *time.Time `gorm:"index" json:"mute_until"`
	Status    int        `gorm:"default:1;index" json:"status"`
	CreatedAt time.Time  `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName 表名
func (GroupMember) TableName() string { return "im_group_members" }
