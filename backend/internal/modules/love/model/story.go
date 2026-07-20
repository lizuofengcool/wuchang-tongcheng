// Package model love 相亲交友数据模型 - 动态广场表 LoveStory
// 对标陌陌/探探：发布动态（图文/视频/语音）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveStory 动态广场表
type LoveStory struct {
	database.RegionBaseModel

	StoryNo string `gorm:"size:64;not null;uniqueIndex" json:"story_no"`
	LoveID  uint   `gorm:"index;not null" json:"love_id"`
	UserID  uint   `gorm:"index;not null" json:"user_id"`

	// 冗余用户信息
	UserNickname string `gorm:"size:64;not null;default:''" json:"user_nickname"`
	UserAvatar   string `gorm:"size:255;not null;default:''" json:"user_avatar"`
	UserGender   int    `gorm:"not null;default:0" json:"user_gender"`

	// 内容
	Title   string `gorm:"size:200;not null;default:''" json:"title"`
	Content string `gorm:"type:text" json:"content"`

	// 媒体
	MediaType string `gorm:"size:16;not null;default:'image';index" json:"media_type"` // image/video/voice
	ImageUrls JSONB  `gorm:"type:jsonb" json:"image_urls"`
	VideoURL  string `gorm:"size:255;not null;default:''" json:"video_url"`
	VideoCover string `gorm:"size:255;not null;default:''" json:"video_cover"`
	VideoDuration int `gorm:"not null;default:0" json:"video_duration"`
	VoiceURL  string `gorm:"size:255;not null;default:''" json:"voice_url"`
	VoiceDuration int `gorm:"not null;default:0" json:"voice_duration"`

	// 位置
	Location  string  `gorm:"size:128;not null;default:''" json:"location"`
	Latitude  float64 `gorm:"type:decimal(10,7);not null;default:0" json:"latitude"`
	Longitude float64 `gorm:"type:decimal(10,7);not null;default:0" json:"longitude"`

	// 标签/话题
	Tags  JSONB `gorm:"type:jsonb" json:"tags"`
	Topic string `gorm:"size:64;not null;default:'';index" json:"topic"`

	// 统计
	ViewCount    int `gorm:"not null;default:0" json:"view_count"`
	LikeCount    int `gorm:"not null;default:0" json:"like_count"`
	CommentCount int `gorm:"not null;default:0" json:"comment_count"`
	ShareCount   int `gorm:"not null;default:0" json:"share_count"`
	ReportCount  int `gorm:"not null;default:0" json:"report_count"`

	// 运营
	Featured bool `gorm:"not null;default:false;index" json:"featured"`

	// 状态/审核
	Status      int    `gorm:"not null;default:1;index" json:"status"`        // 0下架 1正常 2冻结
	AuditStatus int    `gorm:"not null;default:1;index" json:"audit_status"` // 0待审 1通过 2拒绝
	AuditReason string `gorm:"size:500;not null;default:''" json:"audit_reason"`
	PublishedAt *time.Time `gorm:"index" json:"published_at"`

	// 热度/风控
	HotScore    float64 `gorm:"type:decimal(8,2);not null;default:0;index" json:"hot_score"`
	ContentHash string  `gorm:"size:64;not null;default:''" json:"content_hash"`
	RiskScore   int     `gorm:"not null;default:0" json:"risk_score"`
}

// TableName 表名
func (LoveStory) TableName() string { return "love_stories" }
