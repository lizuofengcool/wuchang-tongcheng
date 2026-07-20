// Package model love 相亲交友数据模型 - 聊天会话表 LoveChatSession
// 对标陌陌/Soul：匹配后开聊，未读/最后消息
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 会话状态 ===
const (
	ChatSessionStatusActive    = 1 // 活跃
	ChatSessionStatusMuted     = 2 // 免打扰
	ChatSessionStatusDissolved = 3 // 已解散
	ChatSessionStatusBlocked   = 4 // 已拉黑
)

// === 消息类型 ===
const (
	ChatMessageTypeText  = "text"  // 文本
	ChatMessageTypeImage = "image" // 图片
	ChatMessageTypeVoice = "voice" // 语音
	ChatMessageTypeVideo = "video" // 视频
	ChatMessageTypeGift  = "gift"  // 礼物
	ChatMessageTypeCard  = "card"  // 名片
	ChatMessageTypeSystem = "system" // 系统
)

// LoveChatSession 匹配后聊天会话表
// 唯一约束 session_no + match_id 保证一会话一记录
type LoveChatSession struct {
	database.RegionBaseModel

	SessionNo string `gorm:"size:64;not null;uniqueIndex" json:"session_no"` // 会话编号
	MatchID   uint   `gorm:"not null;uniqueIndex;index" json:"match_id"`     // 关联匹配 ID

	UserIDA uint `gorm:"index;not null" json:"user_id_a"` // 用户 A UserID
	UserIDB uint `gorm:"index;not null" json:"user_id_b"` // 用户 B UserID
	LoveIDA uint `gorm:"index;not null" json:"love_id_a"` // 用户 A LoveID
	LoveIDB uint `gorm:"index;not null" json:"love_id_b"` // 用户 B LoveID

	NicknameA string `gorm:"size:64;not null;default:''" json:"nickname_a"`
	NicknameB string `gorm:"size:64;not null;default:''" json:"nickname_b"`
	AvatarA   string `gorm:"size:255;not null;default:''" json:"avatar_a"`
	AvatarB   string `gorm:"size:255;not null;default:''" json:"avatar_b"`

	// 最后一条消息（冗余便于会话列表展示）
	LastMessageID      uint       `gorm:"not null;default:0" json:"last_message_id"`
	LastMessageContent string     `gorm:"type:text;not null;default:''" json:"last_message_content"`
	LastMessageType    string     `gorm:"size:16;not null;default:''" json:"last_message_type"`
	LastMessageAt      *time.Time `gorm:"index" json:"last_message_at"`
	LastSenderID       uint       `gorm:"not null;default:0" json:"last_sender_id"`

	// 未读数（双方各自计数）
	UnreadCountA int `gorm:"not null;default:0" json:"unread_count_a"`
	UnreadCountB int `gorm:"not null;default:0" json:"unread_count_b"`

	// 用户级开关（双方各自）
	MutedA   bool `gorm:"not null;default:false" json:"muted_a"`     // A 是否免打扰
	MutedB   bool `gorm:"not null;default:false" json:"muted_b"`     // B 是否免打扰
	PinnedA  bool `gorm:"not null;default:false" json:"pinned_a"`    // A 是否置顶
	PinnedB  bool `gorm:"not null;default:false" json:"pinned_b"`    // B 是否置顶
	DeletedA bool `gorm:"not null;default:false" json:"deleted_a"`   // A 是否删除
	DeletedB bool `gorm:"not null;default:false" json:"deleted_b"`   // B 是否删除

	// 状态
	Status        int        `gorm:"not null;default:1;index" json:"status"` // 1活跃 3已解散 4已拉黑
	DissolvedAt   *time.Time `json:"dissolved_at"`
	DissolveReason string    `gorm:"size:255;not null;default:''" json:"dissolve_reason"`
	DissolveBy    uint       `gorm:"not null;default:0" json:"dissolve_by"`

	// 统计
	MessageCount int `gorm:"not null;default:0" json:"message_count"`
	GiftCount    int `gorm:"not null;default:0" json:"gift_count"`
}

// TableName 表名
func (LoveChatSession) TableName() string { return "love_chat_sessions" }
