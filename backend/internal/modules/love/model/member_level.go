// Package model love 相亲交友数据模型 - 会员等级定义表 LoveMemberLevel
// 对标陌陌/Soul：基础/高级/VIP/Premium 四级
// 全局表（无 region_id 隔离），由 BaseModel 直接继承
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// LoveMemberLevel 会员等级定义表
// 用于配置不同等级的会员权益、价格、每日限额等
type LoveMemberLevel struct {
	database.BaseModel

	LevelCode string `gorm:"size:32;not null;uniqueIndex" json:"level_code"` // basic/advanced/vip/premium
	LevelName string `gorm:"size:64;not null" json:"level_name"`             // 等级名
	Level     int    `gorm:"not null;uniqueIndex" json:"level"`               // 等级值 1-4
	Description string `gorm:"type:text" json:"description"`
	Icon       string `gorm:"size:255;not null;default:''" json:"icon"`
	Color      string `gorm:"size:32;not null;default:''" json:"color"`

	// 价格
	MonthlyPrice   float64 `gorm:"type:decimal(12,2);not null;default:0" json:"monthly_price"`
	QuarterlyPrice float64 `gorm:"type:decimal(12,2);not null;default:0" json:"quarterly_price"`
	YearlyPrice    float64 `gorm:"type:decimal(12,2);not null;default:0" json:"yearly_price"`

	// 每日限额
	DailySuperLikes      int `gorm:"not null;default:0" json:"daily_super_likes"`
	DailyLikes           int `gorm:"not null;default:0" json:"daily_likes"`
	DailyVisits          int `gorm:"not null;default:0" json:"daily_visits"`
	DailyRecommendations int `gorm:"not null;default:0" json:"daily_recommendations"`

	// 权益
	CanSeeVisitors    bool `gorm:"not null;default:false" json:"can_see_visitors"`
	CanSeeLikes       bool `gorm:"not null;default:false" json:"can_see_likes"`
	CanHideOnline     bool `gorm:"not null;default:false" json:"can_hide_online"`
	CanHideLocation   bool `gorm:"not null;default:false" json:"can_hide_location"`
	CanFilterVerified bool `gorm:"not null;default:false" json:"can_filter_verified"`
	CanAdvancedFilter bool `gorm:"not null;default:false" json:"can_advanced_filter"`
	CanSuperLike      bool `gorm:"not null;default:false" json:"can_super_like"`
	CanUndoSwipe      bool `gorm:"not null;default:false" json:"can_undo_swipe"`
	CanBoostProfile   bool `gorm:"not null;default:false" json:"can_boost_profile"`
	CanSeeMatchScore  bool `gorm:"not null;default:false" json:"can_see_match_score"`

	// 扩展权益 JSONB
	Perks JSONB `gorm:"type:jsonb" json:"perks"`

	// 排序/状态
	Sort   int `gorm:"not null;default:0;index" json:"sort"`
	Status int `gorm:"not null;default:1;index" json:"status"` // 0禁用 1启用
}

// TableName 表名
func (LoveMemberLevel) TableName() string { return "love_member_levels" }
