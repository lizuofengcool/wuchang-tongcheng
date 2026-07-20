// Package model love 相亲交友数据模型 - 拉黑名单表 LoveBlock
// 对标陌陌/探探：拉黑后无法匹配/聊天/查看
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// LoveBlock 拉黑名单表
// 唯一约束 (user_id, blocked_user_id) 保证不重复拉黑
type LoveBlock struct {
	database.RegionBaseModel

	UserID         uint `gorm:"index;not null" json:"user_id"`          // 拉黑者 UserID
	LoveID         uint `gorm:"index;not null" json:"love_id"`          // 拉黑者 LoveID
	BlockedUserID  uint `gorm:"index;not null" json:"blocked_user_id"`  // 被拉黑者 UserID
	BlockedLoveID  uint `gorm:"index;not null" json:"blocked_love_id"`  // 被拉黑者 LoveID

	BlockedNickname string `gorm:"size:64;not null;default:''" json:"blocked_nickname"`
	BlockedAvatar   string `gorm:"size:255;not null;default:''" json:"blocked_avatar"`

	Reason   string `gorm:"size:255;not null;default:''" json:"reason"`
	ReportID uint   `gorm:"not null;default:0" json:"report_id"` // 关联举报 ID
	Status   int    `gorm:"not null;default:1;index" json:"status"`
}

// TableName 表名
func (LoveBlock) TableName() string { return "love_blocks" }
