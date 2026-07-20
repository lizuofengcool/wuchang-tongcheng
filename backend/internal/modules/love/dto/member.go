// Package dto love 相亲交友数据传输对象 - 会员
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveMemberLevelInfo 会员等级详情响应
type LoveMemberLevelInfo struct {
	ID                uint      `json:"id"`
	LevelCode         string    `json:"level_code"`
	LevelName         string    `json:"level_name"`
	Level             int       `json:"level"`
	Description       string    `json:"description"`
	Icon              string    `json:"icon"`
	Color             string    `json:"color"`
	MonthlyPrice      float64   `json:"monthly_price"`
	QuarterlyPrice    float64   `json:"quarterly_price"`
	YearlyPrice       float64   `json:"yearly_price"`
	DailySuperLikes   int       `json:"daily_super_likes"`
	DailyLikes        int       `json:"daily_likes"`
	DailyVisits       int       `json:"daily_visits"`
	DailyRecommendations int    `json:"daily_recommendations"`
	CanSeeVisitors    bool      `json:"can_see_visitors"`
	CanSeeLikes       bool      `json:"can_see_likes"`
	CanHideOnline     bool      `json:"can_hide_online"`
	CanHideLocation   bool      `json:"can_hide_location"`
	CanFilterVerified bool      `json:"can_filter_verified"`
	CanAdvancedFilter bool      `json:"can_advanced_filter"`
	CanSuperLike      bool      `json:"can_super_like"`
	CanUndoSwipe      bool      `json:"can_undo_swipe"`
	CanBoostProfile   bool      `json:"can_boost_profile"`
	CanSeeMatchScore  bool      `json:"can_see_match_score"`
	Perks             interface{} `json:"perks"`
	Sort              int       `json:"sort"`
	Status            int       `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreateLoveMemberLevelRequest 创建会员等级请求
type CreateLoveMemberLevelRequest struct {
	LevelCode      string `json:"level_code" binding:"required,max=32"`
	LevelName      string `json:"level_name" binding:"required,max=64"`
	Level          int    `json:"level" binding:"min=1,max=10"`
	Description    string `json:"description"`
	Icon           string `json:"icon" binding:"max=255"`
	Color          string `json:"color" binding:"max=32"`
	MonthlyPrice   float64 `json:"monthly_price"`
	QuarterlyPrice float64 `json:"quarterly_price"`
	YearlyPrice    float64 `json:"yearly_price"`
	DailySuperLikes int    `json:"daily_super_likes"`
	DailyLikes     int    `json:"daily_likes"`
	DailyVisits    int    `json:"daily_visits"`
	DailyRecommendations int `json:"daily_recommendations"`
	CanSeeVisitors bool   `json:"can_see_visitors"`
	CanSeeLikes    bool   `json:"can_see_likes"`
	CanHideOnline  bool   `json:"can_hide_online"`
	CanHideLocation bool  `json:"can_hide_location"`
	CanFilterVerified bool `json:"can_filter_verified"`
	CanAdvancedFilter bool `json:"can_advanced_filter"`
	CanSuperLike   bool   `json:"can_super_like"`
	CanUndoSwipe   bool   `json:"can_undo_swipe"`
	CanBoostProfile bool  `json:"can_boost_profile"`
	CanSeeMatchScore bool `json:"can_see_match_score"`
	Perks          interface{} `json:"perks"`
	Sort           int    `json:"sort"`
	Status         int    `json:"status"`
}

// UpdateLoveMemberLevelRequest 更新会员等级请求
type UpdateLoveMemberLevelRequest struct {
	LevelName      *string `json:"level_name" binding:"omitempty,max=64"`
	Description    *string `json:"description"`
	Icon           *string `json:"icon" binding:"omitempty,max=255"`
	Color          *string `json:"color" binding:"omitempty,max=32"`
	MonthlyPrice   *float64 `json:"monthly_price"`
	QuarterlyPrice *float64 `json:"quarterly_price"`
	YearlyPrice    *float64 `json:"yearly_price"`
	DailySuperLikes *int   `json:"daily_super_likes"`
	DailyLikes     *int    `json:"daily_likes"`
	DailyVisits    *int    `json:"daily_visits"`
	DailyRecommendations *int `json:"daily_recommendations"`
	Sort           *int    `json:"sort"`
	Status         *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// LoveMembershipInfo 会员订阅响应
type LoveMembershipInfo struct {
	ID            uint       `json:"id"`
	SubNo         string     `json:"sub_no"`
	UserID        uint       `json:"user_id"`
	LoveID        uint       `json:"love_id"`
	LevelCode     string     `json:"level_code"`
	LevelName     string     `json:"level_name"`
	Level         int        `json:"level"`
	Plan          string     `json:"plan"`
	Period        int        `json:"period"`
	StartAt       time.Time  `json:"start_at"`
	EndAt         time.Time  `json:"end_at"`
	Price         float64    `json:"price"`
	PayAmount     float64    `json:"pay_amount"`
	PayMethod     string     `json:"pay_method"`
	PayOrderNo    string     `json:"pay_order_no"`
	PayAt         *time.Time `json:"pay_at"`
	AutoRenew     bool       `json:"auto_renew"`
	RenewCount    int        `json:"renew_count"`
	PerksSnapshot interface{} `json:"perks_snapshot"`
	Status        int        `json:"status"`
	StatusText    string     `json:"status_text"`
	CancelAt      *time.Time `json:"cancel_at"`
	CancelReason  string     `json:"cancel_reason"`
	RefundAmount  float64    `json:"refund_amount"`
	RefundAt      *time.Time `json:"refund_at"`
	RefundReason  string     `json:"refund_reason"`
	Source        string     `json:"source"`
	Remark        string     `json:"remark"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CreateLoveMembershipRequest 开通会员请求
type CreateLoveMembershipRequest struct {
	LevelCode string `json:"level_code" binding:"required"`
	Plan      string `json:"plan" binding:"required,oneof=monthly quarterly yearly"`
	Period    int    `json:"period" binding:"min=1"`
	PayMethod string `json:"pay_method" binding:"omitempty,oneof=wechat alipay credits"`
	AutoRenew bool   `json:"auto_renew"`
}

// CancelLoveMembershipRequest 取消订阅请求
type CancelLoveMembershipRequest struct {
	Reason string `json:"reason" binding:"max=255"`
}

// RefundLoveMembershipRequest 退款请求
type RefundLoveMembershipRequest struct {
	Reason string `json:"reason" binding:"required,max=255"`
	Amount float64 `json:"amount" binding:"min=0"`
}

// LoveMembershipListRequest 会员订阅列表请求
type LoveMembershipListRequest struct {
	UserID    uint   `form:"user_id" json:"user_id"`
	LoveID    uint   `form:"love_id" json:"love_id"`
	LevelCode string `form:"level_code" json:"level_code"`
	Plan      string `form:"plan" json:"plan"`
	Status    *int   `form:"status" json:"status"`
	utils.Pagination
}
