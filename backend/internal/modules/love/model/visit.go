// Package model love 相亲交友数据模型 - 访客记录表 LoveVisit
// 对标陌陌/探探：谁看过我
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// LoveVisit 访客记录表
// 唯一约束 (user_id, visitor_user_id) 保证同一访客每日聚合
type LoveVisit struct {
	database.RegionBaseModel

	LoveID         uint `gorm:"index;not null" json:"love_id"`          // 被访者 LoveID
	UserID         uint `gorm:"index;not null" json:"user_id"`          // 被访者 UserID
	VisitorUserID  uint `gorm:"index;not null" json:"visitor_user_id"`  // 访客 UserID
	VisitorLoveID  uint `gorm:"index;not null" json:"visitor_love_id"`  // 访客 LoveID
	VisitorNickname string `gorm:"size:64;not null;default:''" json:"visitor_nickname"`
	VisitorAvatar   string `gorm:"size:255;not null;default:''" json:"visitor_avatar"`
	VisitorGender   int    `gorm:"not null;default:0" json:"visitor_gender"`

	VisitType string `gorm:"size:32;not null;default:'profile';index" json:"visit_type"` // profile/story/photo
	Source    string `gorm:"size:32;not null;default:'recommend'" json:"source"`         // 来源

	IP        string `gorm:"size:64;not null;default:''" json:"ip"`
	UserAgent string `gorm:"size:255;not null;default:''" json:"user_agent"`
	Duration  int    `gorm:"not null;default:0" json:"duration"` // 停留时长（秒）

	IsHidden bool `gorm:"not null;default:false" json:"is_hidden"` // 是否隐藏（高级会员可隐身访问）
	IsRead   bool `gorm:"not null;default:false" json:"is_read"`   // 是否已读
	Status   int  `gorm:"not null;default:1;index" json:"status"`
}

// TableName 表名
func (LoveVisit) TableName() string { return "love_visits" }
