// Package model love 相亲交友数据模型 - 印象标签表 LoveImpression
// 对标陌陌/探探：他人评价（如"温柔"、"有趣"）
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// LoveImpression 印象标签表
// 已匹配过的用户可对对方留下印象标签
type LoveImpression struct {
	database.RegionBaseModel

	LoveID uint `gorm:"index;not null" json:"love_id"`     // 被评价人 LoveID
	UserID uint `gorm:"index;not null" json:"user_id"`     // 被评价人 UserID

	FromUserID       uint   `gorm:"index;not null" json:"from_user_id"`
	FromUserNickname string `gorm:"size:64;not null;default:''" json:"from_user_nickname"`
	FromUserAvatar   string `gorm:"size:255;not null;default:''" json:"from_user_avatar"`

	Tag     string `gorm:"size:32;not null;index" json:"tag"`     // 标签文本
	Content string `gorm:"type:text" json:"content"`              // 评价内容

	Anonymous    bool `gorm:"not null;default:false" json:"anonymous"`        // 匿名
	IsAnonymous  bool `gorm:"not null;default:false" json:"is_anonymous"`     // 是否匿名（冗余字段）
	MatchID      uint `gorm:"not null;default:0" json:"match_id"`             // 关联匹配记录
	Status       int  `gorm:"not null;default:1;index" json:"status"`         // 0隐藏 1正常
}

// TableName 表名
func (LoveImpression) TableName() string { return "love_impressions" }
