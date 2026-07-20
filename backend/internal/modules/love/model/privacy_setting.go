// Package model love 相亲交友数据模型 - 隐私设置表 LovePrivacySetting
// 对标陌陌/探探：隐藏在线/位置/年龄/距离/通讯录
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 可见性等级 ===
const (
	VisibilityEveryone = 0 // 所有人可见
	VisibilityMember   = 1 // 仅会员可见
	VisibilityMatched  = 2 // 仅匹配可见
	VisibilityHidden   = 3 // 完全隐藏
)

// LovePrivacySetting 隐私设置表
// 唯一约束 (user_id) 与 (love_id) 保证一人一记录
type LovePrivacySetting struct {
	database.RegionBaseModel

	UserID uint `gorm:"index;not null;uniqueIndex" json:"user_id"` // 用户 UserID
	LoveID uint `gorm:"index;not null;uniqueIndex" json:"love_id"` // 用户 LoveID

	// 隐藏字段（true 表示隐藏）
	HideOnline        bool `gorm:"not null;default:false" json:"hide_online"`         // 隐藏在线状态
	HideLocation      bool `gorm:"not null;default:false" json:"hide_location"`       // 隐藏位置
	HideAge           bool `gorm:"not null;default:false" json:"hide_age"`            // 隐藏年龄
	HideDistance      bool `gorm:"not null;default:false" json:"hide_distance"`       // 隐藏距离
	HideConstellation bool `gorm:"not null;default:false" json:"hide_constellation"`  // 隐藏星座
	HideHometown      bool `gorm:"not null;default:false" json:"hide_hometown"`       // 隐藏家乡
	HideOccupation    bool `gorm:"not null;default:false" json:"hide_occupation"`     // 隐藏职业
	HideIncome        bool `gorm:"not null;default:false" json:"hide_income"`         // 隐藏收入
	HideLastActive    bool `gorm:"not null;default:true" json:"hide_last_active"`     // 隐藏最近活跃（默认隐藏）
	HideVisitors      bool `gorm:"not null;default:false" json:"hide_visitors"`       // 隐藏访客列表

	// 权限控制
	OnlyVerifiedCanSee   bool `gorm:"not null;default:false" json:"only_verified_can_see"`   // 仅认证用户可见
	OnlyVerifiedCanMatch bool `gorm:"not null;default:false" json:"only_verified_can_match"` // 仅认证用户可匹配
	OnlyMemberCanChat    bool `gorm:"not null;default:false" json:"only_member_can_chat"`    // 仅会员可发起聊天
	BlockStrangers      bool `gorm:"not null;default:false" json:"block_strangers"`         // 屏蔽陌生人消息
	BlockSameCity       bool `gorm:"not null;default:false" json:"block_same_city"`         // 屏蔽同城（仅外地）
	AllowPhoneLookup    bool `gorm:"not null;default:false" json:"allow_phone_lookup"`      // 允许手机号搜索
	AllowContactImport  bool `gorm:"not null;default:false" json:"allow_contact_import"`    // 允许通讯录匹配
	AllowRecommendation bool `gorm:"not null;default:true" json:"allow_recommendation"`     // 允许被推荐
	AllowStory          bool `gorm:"not null;default:true" json:"allow_story"`              // 允许发布动态
	AllowImpression     bool `gorm:"not null;default:true" json:"allow_impression"`         // 允许他人评价

	// 可见性等级
	DistanceVisibility int `gorm:"not null;default:0" json:"distance_visibility"` // 0所有人 1仅会员 2仅匹配 3完全隐藏
	AgeVisibility      int `gorm:"not null;default:0" json:"age_visibility"`      // 同上

	Status int `gorm:"not null;default:1;index" json:"status"`
}

// TableName 表名
func (LovePrivacySetting) TableName() string { return "love_privacy_settings" }
