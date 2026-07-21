// Package dto 营销活动中台数据传输对象 - 签到（sign 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// SignRecordInfo 签到记录详情响应
type SignRecordInfo struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	SignDate       time.Time `json:"sign_date"`
	ContinuousDays int       `json:"continuous_days"`
	Points         int       `json:"points"`
	CreatedAt      time.Time `json:"created_at"`
}

// SignCheckInResponse 签到响应
type SignCheckInResponse struct {
	Record         *SignRecordInfo `json:"record"`
	ContinuousDays int             `json:"continuous_days"` // 连续签到天数
	Points         int             `json:"points"`          // 本次获得积分
	TotalPoints    int             `json:"total_points"`    // 累计获得积分
	ExtraReward    interface{}     `json:"extra_reward,omitempty"` // 触发的额外奖励
}

// SignCalendarResponse 签到日历响应
type SignCalendarResponse struct {
	Records       []SignRecordInfo `json:"records"`         // 当月签到记录
	ContinuousDays int              `json:"continuous_days"` // 当前连续签到天数
	MonthDays      int              `json:"month_days"`      // 当月天数
	SignedDays     int              `json:"signed_days"`     // 当月已签天数
	TotalPoints    int              `json:"total_points"`    // 当月累计积分
}

// SignCalendarRequest 签到日历请求
type SignCalendarRequest struct {
	Month string `form:"month" json:"month"` // YYYY-MM，空则取当月
}

// SignRuleInfo 签到规则详情响应
type SignRuleInfo struct {
	ID          uint       `json:"id"`
	Day         int        `json:"day"`
	Points      int        `json:"points"`
	ExtraReward interface{} `json:"extra_reward"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateSignRuleRequest 创建签到规则请求
type CreateSignRuleRequest struct {
	Day         int         `json:"day" binding:"required,min=1"`
	Points      int         `json:"points" binding:"min=0"`
	ExtraReward interface{} `json:"extra_reward"`
	Status      int         `json:"status"`
}

// UpdateSignRuleRequest 更新签到规则请求
type UpdateSignRuleRequest struct {
	Points      *int         `json:"points" binding:"omitempty,min=0"`
	ExtraReward interface{}  `json:"extra_reward"`
	Status      *int         `json:"status"`
}

// SignRuleListRequest 签到规则列表请求
type SignRuleListRequest struct {
	Status *int `form:"status" json:"status"`
	utils.Pagination
}
