// Package model love 相亲交友数据模型 - 虚拟礼物表 LoveGift
// 对标陌陌/Soul：礼物定义（玫瑰/动画礼物/连击礼物）
// 全局表（无 region_id 隔离）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// LoveGift 虚拟礼物定义表
// 由 M 端配置，C 端用户使用金币购买后赠送
type LoveGift struct {
	database.BaseModel

	GiftCode string `gorm:"size:32;not null;uniqueIndex" json:"gift_code"` // rose/bouquet/ring/car
	GiftName string `gorm:"size:64;not null" json:"gift_name"`
	Category string `gorm:"size:32;not null;default:'common';index" json:"category"` // common/luxury/animated/festival/limited
	Description string `gorm:"type:text" json:"description"`
	Icon       string `gorm:"size:255;not null;default:''" json:"icon"`

	// 动画
	AnimationURL      string `gorm:"size:255;not null;default:''" json:"animation_url"`
	AnimationType     string `gorm:"size:32;not null;default:''" json:"animation_type"`
	AnimationDuration int    `gorm:"not null;default:0" json:"animation_duration"`

	// 价格（金币）
	Price         float64 `gorm:"type:decimal(12,2);not null;default:0;index" json:"price"`
	OriginalPrice float64 `gorm:"type:decimal(12,2);not null;default:0" json:"original_price"`
	DiscountPrice float64 `gorm:"type:decimal(12,2);not null;default:0" json:"discount_price"`
	MemberLevel   int     `gorm:"not null;default:0;index" json:"member_level"` // 享受会员价所需等级

	// 魅力值
	CharmValue int `gorm:"not null;default:0" json:"charm_value"` // 礼物增加的魅力值

	// 限制
	IsLimited bool `gorm:"not null;default:false" json:"is_limited"`
	IsAnimated bool `gorm:"not null;default:false" json:"is_animated"`
	IsCombo   bool `gorm:"not null;default:false" json:"is_combo"`     // 是否支持连击
	ComboMin  int  `gorm:"not null;default:0" json:"combo_min"`
	ComboMax  int  `gorm:"not null;default:0" json:"combo_max"`
	DailyLimit int `gorm:"not null;default:0" json:"daily_limit"`      // 每日赠送上限

	// 排序/状态
	Sort   int        `gorm:"not null;default:0;index" json:"sort"`
	Status int        `gorm:"not null;default:1;index" json:"status"` // 0下架 1上架
	StartAt *time.Time `json:"start_at"`                              // 限时礼物开始
	EndAt   *time.Time `json:"end_at"`                                // 限时礼物结束
}

// TableName 表名
func (LoveGift) TableName() string { return "love_gifts" }
