// Package model 沟通消息表（对标 BOSS直聘在线聊天）
// 在线聊天 + 卡片消息 + 系统消息
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 消息类型常量 ===
const (
	MessageTypeText       = "text"        // 文本
	MessageTypeImage      = "image"       // 图片
	MessageTypeVoice      = "voice"       // 语音
	MessageTypeVideo      = "video"       // 视频
	MessageTypeFile       = "file"        // 文件
	MessageTypeCard       = "card"        // 卡片（职位/简历/面试卡片）
	MessageTypeSystem     = "system"      // 系统消息
	MessageTypeRecruit    = "recruit"     // 招聘卡片
	MessageTypeResume     = "resume"      // 简历卡片
	MessageTypeInterview  = "interview"   // 面试卡片
	MessageTypeOffer      = "offer"       // Offer 卡片
	MessageTypeGreeting   = "greeting"   // 打招呼
	MessageTypeReadNotice = "read_notice" // 已读回执
)

// === 消息状态常量 ===
const (
	MessageStatusNormal  = 1 // 正常
	MessageStatusRecall  = 2 // 已撤回
	MessageStatusDeleted = 3 // 已删除
	MessageStatusBanned  = 4 // 已屏蔽
)

// === 消息来源常量 ===
const (
	MessageSourceChat   = "chat"   // 聊天
	MessageSourceSystem = "system" // 系统
	MessageSourceBot   = "bot"    // 机器人
	MessageSourcePush  = "push"   // 推送
)

// JobMessage 沟通消息表
type JobMessage struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	ConversationID string    `gorm:"size:64;not null;index" json:"conversation_id"`                 // 会话 ID
	JobID          uint      `gorm:"index" json:"job_id"`                                          // 关联职位 ID
	ApplicationID  uint      `gorm:"index" json:"application_id"`                                  // 关联投递记录 ID
	FromUserID     uint      `gorm:"not null;index:idx_job_msg_from_read" json:"from_user_id"`      // 发送者 ID
	ToUserID       uint      `gorm:"not null;index:idx_job_msg_to_read" json:"to_user_id"`          // 接收者 ID
	FromName       string    `gorm:"size:50" json:"from_name"`                                    // 发送者昵称
	FromAvatar     string    `gorm:"size:255" json:"from_avatar"`                                 // 发送者头像
	ToName         string    `gorm:"size:50" json:"to_name"`                                      // 接收者昵称
	ToAvatar       string    `gorm:"size:255" json:"to_avatar"`                                   // 接收者头像
	Content        string    `gorm:"type:text" json:"content"`                                     // 消息内容
	MessageType    string    `gorm:"size:32;default:'text';index" json:"message_type"`             // 消息类型
	Attachments    JSONB     `gorm:"type:jsonb" json:"attachments"`                                // 附件 [{type,name,url,size}]
	IsRead         bool      `gorm:"default:false;index:idx_job_msg_from_read;index:idx_job_msg_to_read" json:"is_read"` // 是否已读
	ReadAt         *time.Time `gorm:"index" json:"read_at"`                                       // 已读时间
	IsRecruiter    bool      `gorm:"default:false" json:"is_recruiter"`                           // 发送者是否为招聘者
	IsSystem       bool      `gorm:"default:false;index" json:"is_system"`                        // 是否系统消息
	Status         int       `gorm:"default:1;index" json:"status"`                              // 1正常 2撤回 3删除 4屏蔽
	Source         string    `gorm:"size:32;default:'chat';index" json:"source"`                   // 消息来源
}

// TableName 表名（job_ 前缀）
func (JobMessage) TableName() string { return "job_messages" }
