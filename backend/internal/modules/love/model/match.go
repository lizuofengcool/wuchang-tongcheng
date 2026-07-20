// Package model love 相亲交友数据模型 - 匹配记录表 LoveMatch
// 对标探探/Soul：双向喜欢则匹配成功，开始聊天
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveMatch 匹配记录表
// 用户 A 与 B 双向喜欢后生成匹配记录，含灵魂匹配维度评分
type LoveMatch struct {
	database.RegionBaseModel

	MatchNo string `gorm:"size:64;not null;uniqueIndex" json:"match_no"` // 匹配单号
	UserIDA uint   `gorm:"index;not null" json:"user_id_a"`              // 用户 A
	UserIDB uint   `gorm:"index;not null" json:"user_id_b"`              // 用户 B
	LoveIDA uint   `gorm:"index;not null" json:"love_id_a"`              // 资料 A
	LoveIDB uint   `gorm:"index;not null" json:"love_id_b"`              // 资料 B

	// 匹配评分（灵魂匹配算法核心字段）
	MatchScore      float64 `gorm:"type:decimal(5,2);not null;default:0" json:"match_score"`       // 总匹配分（0-100）
	MatchType       string  `gorm:"size:32;not null;default:'both_like'" json:"match_type"`        // 匹配类型
	InterestMatch   float64 `gorm:"type:decimal(5,2);not null;default:0" json:"interest_match"`    // 兴趣匹配
	PersonalityMatch float64 `gorm:"type:decimal(5,2);not null;default:0" json:"personality_match"` // 性格匹配
	ValueMatch      float64 `gorm:"type:decimal(5,2);not null;default:0" json:"value_match"`       // 价值观匹配
	LocationMatch   float64 `gorm:"type:decimal(5,2);not null;default:0" json:"location_match"`    // 地理匹配
	AgeMatch        float64 `gorm:"type:decimal(5,2);not null;default:0" json:"age_match"`         // 年龄匹配

	// 匹配时间
	MatchedAt time.Time `gorm:"not null;default:now();index" json:"matched_at"`

	// 关联聊天会话
	ChatSessionID uint `gorm:"not null;default:0" json:"chat_session_id"`

	// 状态
	Status int `gorm:"not null;default:1;index" json:"status"` // 1活跃 2解除 3已解除 4已拉黑

	// 最近消息冗余（避免连表）
	LastMessageAt      *time.Time `gorm:"index" json:"last_message_at"`
	LastMessageContent string     `gorm:"type:text" json:"last_message_content"`
	LastMessageType    string     `gorm:"size:16;not null;default:''" json:"last_message_type"`
	UnreadCountA       int        `gorm:"not null;default:0" json:"unread_count_a"`
	UnreadCountB       int        `gorm:"not null;default:0" json:"unread_count_b"`
	UnmutedByA         bool       `gorm:"not null;default:false" json:"unmuted_by_a"`
	UnmutedByB         bool       `gorm:"not null;default:false" json:"unmuted_by_b"`

	// 解除匹配
	DissolvedAt   *time.Time `json:"dissolved_at"`
	DissolveReason string    `gorm:"size:255;not null;default:''" json:"dissolve_reason"`
	DissolveBy    uint       `gorm:"not null;default:0" json:"dissolve_by"`
}

// TableName 表名
func (LoveMatch) TableName() string { return "love_matches" }
