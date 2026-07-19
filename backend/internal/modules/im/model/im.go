// Package model IM 消息中台精简版数据模型
// 依据 ershou 模块依赖：私聊 + 系统通知 + 隐私号码
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 会话类型常量 ===
const (
	SessionTypePrivate = "private" // 一对一私聊
	SessionTypeGroup   = "group"   // 群聊
)

// === 消息类型常量 ===
const (
	MsgTypeText  = "text"  // 文本
	MsgTypeImage = "image" // 图片
	MsgTypeVoice = "voice" // 语音
	MsgTypeVideo = "video" // 视频
	MsgTypeCard  = "card"  // 商品卡片
)

// === 通知类型常量 ===
const (
	NotifyTypeOrder    = "order"    // 交易状态变更
	NotifyTypeRefund   = "refund"   // 退款通知
	NotifyTypeActivity = "activity" // 活动通知
	NotifyTypeSystem   = "system"   // 系统通知
)

// === 隐私号码状态常量 ===
const (
	PrivacyStatusBound   = 1 // 绑定中
	PrivacyStatusUnbound = 0 // 已解绑
)

// Session IM 会话
type Session struct {
	database.RegionBaseModel

	SessionID     string    `gorm:"size:64;not null;uniqueIndex" json:"session_id"`        // 会话ID
	SessionType   string    `gorm:"size:16;default:'private'" json:"session_type"`         // 会话类型
	Participants  string    `gorm:"type:jsonb;default:'[]'::jsonb" json:"participants"`     // 参与者 ID JSON
	LastMessage   string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"last_message"`     // 最后一条消息快照
	LastMessageAt *time.Time `gorm:"index" json:"last_message_at"`                          // 最后消息时间
	UnreadCount   string    `gorm:"type:jsonb;default:'{}'::jsonb" json:"unread_count"`     // 各用户未读数
	Status        int       `gorm:"default:1;index" json:"status"`                          // 状态：1正常 0禁用
}

// TableName 表名
func (Session) TableName() string { return "im_sessions" }

// Message IM 消息
type Message struct {
	database.RegionBaseModel

	SessionID    string `gorm:"size:64;not null;index" json:"session_id"`                 // 会话ID
	SenderID     uint   `gorm:"index;not null" json:"sender_id"`                          // 发送者ID
	MsgType      string `gorm:"size:16;default:'text'" json:"msg_type"`                   // 消息类型
	Content      string `gorm:"type:text" json:"content"`                                 // 消息内容
	Extra        string `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`              // 扩展信息
	ReadStatus   string `gorm:"type:jsonb;default:'{}'::jsonb" json:"read_status"`        // 各用户已读状态
	IsRecalled   bool   `gorm:"default:false" json:"is_recalled"`                          // 是否撤回
}

// TableName 表名
func (Message) TableName() string { return "im_messages" }

// SystemNotification 系统通知
type SystemNotification struct {
	database.RegionBaseModel

	UserID     uint       `gorm:"index;not null" json:"user_id"`                          // 接收用户ID
	NotifyType string     `gorm:"size:32;index" json:"notify_type"`                       // 通知类型
	Title      string     `gorm:"size:128" json:"title"`                                  // 标题
	Content    string     `gorm:"type:text" json:"content"`                               // 内容
	JumpURL    string     `gorm:"size:256" json:"jump_url"`                               // 跳转链接
	Extra      string     `gorm:"type:jsonb;default:'{}'::jsonb" json:"extra"`            // 扩展字段
	IsRead     bool       `gorm:"default:false;index" json:"is_read"`                     // 是否已读
	ReadAt     *time.Time `gorm:"index" json:"read_at"`                                   // 阅读时间
}

// TableName 表名
func (SystemNotification) TableName() string { return "im_system_notifications" }

// PrivacyNumber 隐私号码
type PrivacyNumber struct {
	database.RegionBaseModel

	PrivacyNo string     `gorm:"size:32;index" json:"privacy_no"`                         // 虚拟号
	RealNoA   string     `gorm:"size:32;not null" json:"real_no_a"`                       // 真实号A
	RealNoB   string     `gorm:"size:32;not null" json:"real_no_b"`                       // 真实号B
	UserIDA   uint       `gorm:"index;not null" json:"user_id_a"`                         // 用户A
	UserIDB   uint       `gorm:"index;not null" json:"user_id_b"`                         // 用户B
	BizModule string     `gorm:"size:32" json:"biz_module"`                               // 业务模块
	BizID     string     `gorm:"size:128" json:"biz_id"`                                  // 业务ID
	CallRecords string   `gorm:"type:jsonb;default:'[]'::jsonb" json:"call_records"`      // 通话记录 JSON
	BoundAt   time.Time  `gorm:"not null;default:now()" json:"bound_at"`                  // 绑定时间
	UnboundAt *time.Time `gorm:"index" json:"unbound_at"`                                 // 解绑时间
	Status    int        `gorm:"default:1;index" json:"status"`                           // 状态：1绑定 0解绑
}

// TableName 表名
func (PrivacyNumber) TableName() string { return "im_privacy_numbers" }
