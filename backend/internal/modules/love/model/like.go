// Package model love 相亲交友数据模型 - 喜欢/不喜欢/心动信号表 LoveLike
// 对标探探/Soul：滑动卡片，双向喜欢则匹配
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveLike 喜欢记录表
// 唯一约束 (user_id, target_user_id) 保证同一对用户只有一条记录
type LoveLike struct {
	database.RegionBaseModel

	UserID         uint `gorm:"index;not null" json:"user_id"`         // 操作者 UserID
	LoveID         uint `gorm:"index;not null" json:"love_id"`         // 操作者 LoveID
	TargetUserID   uint `gorm:"index;not null" json:"target_user_id"`  // 目标 UserID
	TargetLoveID   uint `gorm:"index;not null" json:"target_love_id"`  // 目标 LoveID

	TargetNickname string `gorm:"size:64;not null;default:''" json:"target_nickname"`
	TargetAvatar   string `gorm:"size:255;not null;default:''" json:"target_avatar"`
	TargetGender   int    `gorm:"not null;default:0" json:"target_gender"`

	// 动作
	Action    string `gorm:"size:16;not null;default:'like';index" json:"action"` // like/dislike/skip/super
	SuperLike bool   `gorm:"not null;default:false;index" json:"super_like"`      // 是否心动信号

	// 匹配评分（操作时计算的）
	MatchScore float64 `gorm:"type:decimal(5,2);not null;default:0" json:"match_score"`

	// 来源
	Source string `gorm:"size:32;not null;default:'recommend'" json:"source"` // recommend/nearby/search/story
	IP     string `gorm:"size:64;not null;default:''" json:"ip"`

	// 是否匹配
	IsMatched bool       `gorm:"not null;default:false;index" json:"is_matched"`
	MatchID   uint       `gorm:"not null;default:0" json:"match_id"`
	MatchedAt *time.Time `json:"matched_at"`

	// 撤销
	UndoneAt   *time.Time `json:"undone_at"`
	UndoReason string     `gorm:"size:255;not null;default:''" json:"undo_reason"`
	Status     int        `gorm:"not null;default:1" json:"status"` // 0已撤销 1有效
}

// TableName 表名
func (LoveLike) TableName() string { return "love_likes" }
