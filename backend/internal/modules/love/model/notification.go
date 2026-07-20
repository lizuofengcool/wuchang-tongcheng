// Package model love 相亲交友数据模型 - 通知表 LoveNotification
// 对标陌陌/Soul：喜欢/匹配/访客/礼物通知
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 推送状态 ===
const (
	NotifyPushStatusPending  = "pending"  // 待推送
	NotifyPushStatusSuccess  = "success"  // 推送成功
	NotifyPushStatusFailed   = "failed"   // 推送失败
	NotifyPushStatusSkipped  = "skipped"  // 跳过推送
	NotifyPushStatusDisabled = "disabled" // 用户关闭推送
)

// === 通知状态 ===
const (
	NotifyStatusDeleted = 0 // 已删除
	NotifyStatusNormal  = 1 // 正常
)

// LoveNotification 通知表
// 接收方 user_id 唯一标识收件人，from_user_id 为触发方
type LoveNotification struct {
	database.RegionBaseModel

	UserID uint `gorm:"index;not null" json:"user_id"` // 接收者 UserID
	LoveID uint `gorm:"index;not null" json:"love_id"` // 接收者 LoveID

	Type    string `gorm:"size:32;not null;default:'system';index" json:"type"`    // 通知类型（参考 NotifyType*）
	Title   string `gorm:"size:200;not null;default:''" json:"title"`              // 标题
	Content string `gorm:"type:text" json:"content"`                                // 内容

	FromUserID       uint   `gorm:"not null;default:0" json:"from_user_id"`        // 来源用户 UserID
	FromUserNickname string `gorm:"size:64;not null;default:''" json:"from_user_nickname"`
	FromUserAvatar   string `gorm:"size:255;not null;default:''" json:"from_user_avatar"`

	// 目标资源（如匹配 ID/动态 ID/礼物 ID 等）
	TargetType string `gorm:"size:32;not null;default:''" json:"target_type"` // match/story/gift/visit/verification
	TargetID   uint   `gorm:"not null;default:0" json:"target_id"`
	ActionURL  string `gorm:"size:255;not null;default:''" json:"action_url"` // 跳转 URL

	// 扩展字段
	Extra JSONB `gorm:"type:jsonb" json:"extra"`

	// 已读
	IsRead bool       `gorm:"not null;default:false;index" json:"is_read"`
	ReadAt *time.Time `json:"read_at"`

	// 推送
	IsPushed   bool       `gorm:"not null;default:false" json:"is_pushed"`
	PushedAt   *time.Time `json:"pushed_at"`
	PushStatus string     `gorm:"size:16;not null;default:''" json:"push_status"`
	PushError  string     `gorm:"size:500;not null;default:''" json:"push_error"`

	Status    int        `gorm:"not null;default:1;index" json:"status"` // 0删除 1正常
	ExpiredAt *time.Time `json:"expired_at"`                              // 过期时间（用于自动清理）
}

// TableName 表名
func (LoveNotification) TableName() string { return "love_notifications" }
