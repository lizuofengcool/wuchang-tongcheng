// Package dto love 相亲交友数据传输对象 - 礼物
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveGiftInfo 礼物响应
type LoveGiftInfo struct {
	ID               uint       `json:"id"`
	GiftCode         string     `json:"gift_code"`
	GiftName         string     `json:"gift_name"`
	Category         string     `json:"category"`
	Description      string     `json:"description"`
	Icon             string     `json:"icon"`
	AnimationURL     string     `json:"animation_url"`
	AnimationType    string     `json:"animation_type"`
	AnimationDuration int       `json:"animation_duration"`
	Price            float64    `json:"price"`
	OriginalPrice    float64    `json:"original_price"`
	DiscountPrice    float64    `json:"discount_price"`
	MemberLevel      int        `json:"member_level"`
	CharmValue       int        `json:"charm_value"`
	IsLimited        bool       `json:"is_limited"`
	IsAnimated       bool       `json:"is_animated"`
	IsCombo          bool       `json:"is_combo"`
	ComboMin         int        `json:"combo_min"`
	ComboMax         int        `json:"combo_max"`
	DailyLimit       int        `json:"daily_limit"`
	Sort             int        `json:"sort"`
	Status           int        `json:"status"`
	StatusText       string     `json:"status_text"`
	StartAt          *time.Time `json:"start_at"`
	EndAt            *time.Time `json:"end_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateLoveGiftRequest 创建礼物请求
type CreateLoveGiftRequest struct {
	GiftCode         string `json:"gift_code" binding:"required,max=32"`
	GiftName         string `json:"gift_name" binding:"required,max=64"`
	Category         string `json:"category" binding:"omitempty,oneof=common luxury animated festival limited"`
	Description      string `json:"description"`
	Icon             string `json:"icon" binding:"max=255"`
	AnimationURL     string `json:"animation_url" binding:"max=255"`
	AnimationType    string `json:"animation_type" binding:"max=32"`
	AnimationDuration int    `json:"animation_duration"`
	Price            float64 `json:"price" binding:"min=0"`
	OriginalPrice    float64 `json:"original_price"`
	DiscountPrice    float64 `json:"discount_price"`
	MemberLevel      int     `json:"member_level"`
	CharmValue       int     `json:"charm_value"`
	IsLimited        bool    `json:"is_limited"`
	IsAnimated       bool    `json:"is_animated"`
	IsCombo          bool    `json:"is_combo"`
	ComboMin         int     `json:"combo_min"`
	ComboMax         int     `json:"combo_max"`
	DailyLimit       int     `json:"daily_limit"`
	Sort             int     `json:"sort"`
	Status           int     `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateLoveGiftRequest 更新礼物请求
type UpdateLoveGiftRequest struct {
	GiftName         *string `json:"gift_name" binding:"omitempty,max=64"`
	Category         *string `json:"category" binding:"omitempty,oneof=common luxury animated festival limited"`
	Description      *string `json:"description"`
	Icon             *string `json:"icon" binding:"omitempty,max=255"`
	AnimationURL     *string `json:"animation_url" binding:"omitempty,max=255"`
	AnimationType    *string `json:"animation_type" binding:"omitempty,max=32"`
	AnimationDuration *int   `json:"animation_duration"`
	Price            *float64 `json:"price"`
	OriginalPrice    *float64 `json:"original_price"`
	DiscountPrice    *float64 `json:"discount_price"`
	MemberLevel      *int    `json:"member_level"`
	CharmValue       *int    `json:"charm_value"`
	IsLimited        *bool   `json:"is_limited"`
	IsAnimated       *bool   `json:"is_animated"`
	IsCombo          *bool   `json:"is_combo"`
	ComboMin         *int    `json:"combo_min"`
	ComboMax         *int    `json:"combo_max"`
	DailyLimit       *int    `json:"daily_limit"`
	Sort             *int    `json:"sort"`
	Status           *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// LoveGiftListRequest 礼物列表请求
type LoveGiftListRequest struct {
	Category    string `form:"category" json:"category"`
	MemberLevel *int   `form:"member_level" json:"member_level"`
	Status      *int   `form:"status" json:"status"`
	utils.Pagination
}

// SendLoveGiftRequest 送礼请求
type SendLoveGiftRequest struct {
	GiftID    uint `json:"gift_id" binding:"required"`
	ToUserID  uint `json:"to_user_id" binding:"required"`
	ToLoveID  uint `json:"to_love_id" binding:"required"`
	Count     int    `json:"count" binding:"min=1,max=999"`
	Message   string `json:"message" binding:"max=200"`
	IsCombo   bool   `json:"is_combo"`
	MatchID   uint   `json:"match_id"`
	SessionID uint   `json:"session_id"`
}

// LoveGiftRecordInfo 送礼记录响应
type LoveGiftRecordInfo struct {
	ID         uint      `json:"id"`
	GiftID     uint      `json:"gift_id"`
	GiftName   string    `json:"gift_name"`
	GiftIcon   string    `json:"gift_icon"`
	GiftPrice  float64   `json:"gift_price"`
	Count      int       `json:"count"`
	TotalPrice float64   `json:"total_price"`
	CharmValue int       `json:"charm_value"`
	FromUserID uint      `json:"from_user_id"`
	FromLoveID uint      `json:"from_love_id"`
	FromNickname string  `json:"from_nickname"`
	FromAvatar string    `json:"from_avatar"`
	ToUserID   uint      `json:"to_user_id"`
	ToLoveID   uint      `json:"to_love_id"`
	ToNickname string    `json:"to_nickname"`
	ToAvatar   string    `json:"to_avatar"`
	Message    string    `json:"message"`
	IsCombo    bool      `json:"is_combo"`
	MatchID    uint      `json:"match_id"`
	SessionID  uint      `json:"session_id"`
	CreatedAt  time.Time `json:"created_at"`
}
