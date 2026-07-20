// Package model love 相亲交友数据模型 - 推荐池表 LoveRecommendation
// 对标陌陌/探探：每日推荐/附近推荐/同城推荐
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveRecommendation 推荐池表
// 唯一约束 (user_id, target_user_id, rec_type) 保证同一类型不重复推荐
type LoveRecommendation struct {
	database.RegionBaseModel

	UserID         uint `gorm:"index;not null" json:"user_id"`         // 推荐接收者 UserID
	LoveID         uint `gorm:"index;not null" json:"love_id"`         // 推荐接收者 LoveID
	TargetUserID   uint `gorm:"index;not null" json:"target_user_id"`  // 被推荐者 UserID
	TargetLoveID   uint `gorm:"index;not null" json:"target_love_id"`  // 被推荐者 LoveID
	TargetNickname string `gorm:"size:64;not null;default:''" json:"target_nickname"`
	TargetAvatar   string `gorm:"size:255;not null;default:''" json:"target_avatar"`
	TargetGender   int    `gorm:"not null;default:0" json:"target_gender"`
	TargetAge      int    `gorm:"not null;default:0" json:"target_age"`
	TargetDistance float64 `gorm:"type:decimal(10,2);not null;default:0" json:"target_distance"` // 距离（公里）

	// 推荐类型与来源
	RecType string `gorm:"size:32;not null;default:'daily';index" json:"rec_type"` // daily/nearby/same_city/same_hometown/interest/soulmate/new_user
	Source  string `gorm:"size:32;not null;default:'algorithm';index" json:"source"` // algorithm/manual/ai/boost

	// 评分（5 维度对标灵魂匹配）
	Score            float64 `gorm:"type:decimal(5,2);not null;default:0;index" json:"score"`
	InterestMatch    float64 `gorm:"type:decimal(5,2);not null;default:0" json:"interest_match"`
	PersonalityMatch float64 `gorm:"type:decimal(5,2);not null;default:0" json:"personality_match"`
	ValueMatch       float64 `gorm:"type:decimal(5,2);not null;default:0" json:"value_match"`
	LocationMatch    float64 `gorm:"type:decimal(5,2);not null;default:0" json:"location_match"`
	AgeMatch         float64 `gorm:"type:decimal(5,2);not null;default:0" json:"age_match"`

	Reason string `gorm:"size:500;not null;default:''" json:"reason"` // 推荐理由（标签词）

	// 互动状态
	IsViewed     bool       `gorm:"not null;default:false" json:"is_viewed"`
	IsLiked      bool       `gorm:"not null;default:false" json:"is_liked"`
	IsDisliked   bool       `gorm:"not null;default:false" json:"is_disliked"`
	IsSuperLiked bool       `gorm:"not null;default:false" json:"is_super_liked"`
	IsSkipped    bool       `gorm:"not null;default:false" json:"is_skipped"`
	IsDismissed  bool       `gorm:"not null;default:false" json:"is_dismissed"`

	ViewedAt     *time.Time `json:"viewed_at"`
	LikedAt      *time.Time `json:"liked_at"`
	DislikedAt   *time.Time `json:"disliked_at"`
	SuperLikedAt *time.Time `json:"super_liked_at"`
	SkippedAt    *time.Time `json:"skipped_at"`
	DismissedAt  *time.Time `json:"dismissed_at"`
	ExpiredAt    *time.Time `gorm:"index" json:"expired_at"`

	Status int `gorm:"not null;default:0;index" json:"status"` // 0待处理 1已查看 2已喜欢 3已不喜欢 4已跳过 5已忽略 6已过期
}

// TableName 表名
func (LoveRecommendation) TableName() string { return "love_recommendations" }
