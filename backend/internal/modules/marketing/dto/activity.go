// Package dto 营销活动中台数据传输对象 - 营销活动（activity 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ActivityInfo 营销活动详情响应
type ActivityInfo struct {
	ID          uint       `json:"id"`
	RegionID    uint       `json:"region_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	TypeText    string     `json:"type_text"`
	Description string     `json:"description"`
	CoverImage  string     `json:"cover_image"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	Config      interface{} `json:"config"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateActivityRequest 创建活动请求
type CreateActivityRequest struct {
	Title       string      `json:"title" binding:"required,max=100"`
	Type        string      `json:"type" binding:"required,oneof=groupbuy bargain seckill lottery"`
	Description string      `json:"description"`
	CoverImage  string      `json:"cover_image" binding:"max=500"`
	StartAt     *time.Time  `json:"start_at"`
	EndAt       *time.Time  `json:"end_at"`
	Status      int         `json:"status"`
	Config      interface{} `json:"config"`
}

// UpdateActivityRequest 更新活动请求
type UpdateActivityRequest struct {
	Title       *string     `json:"title" binding:"omitempty,max=100"`
	Type        *string     `json:"type" binding:"omitempty,oneof=groupbuy bargain seckill lottery"`
	Description *string     `json:"description"`
	CoverImage  *string     `json:"cover_image" binding:"omitempty,max=500"`
	StartAt     *time.Time  `json:"start_at"`
	EndAt       *time.Time  `json:"end_at"`
	Status      *int        `json:"status"`
	Config      interface{} `json:"config"`
}

// ActivityListRequest 活动列表请求
type ActivityListRequest struct {
	Type    string `form:"type" json:"type"`
	Status  *int   `form:"status" json:"status"`
	Keyword string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ActivityParticipateRequest 参与活动请求
type ActivityParticipateRequest struct {
	OrderID uint        `json:"order_id"`
	Detail  interface{} `json:"detail"` // 参与详情（如拼团 ID/砍价进度/秒杀 SKU/抽奖号码）
}

// ActivityParticipationInfo 活动参与记录响应
type ActivityParticipationInfo struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"user_id"`
	ActivityID uint       `json:"activity_id"`
	OrderID    uint       `json:"order_id"`
	Detail     interface{} `json:"detail"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ActivityParticipationListRequest 活动参与记录列表请求
type ActivityParticipationListRequest struct {
	ActivityID uint `form:"activity_id" json:"activity_id"`
	utils.Pagination
}

// ActivityStatistics 活动统计
type ActivityStatistics struct {
	TotalActivities   int64 `json:"total_activities"`
	OngoingActivities int64 `json:"ongoing_activities"`
	PendingActivities int64 `json:"pending_activities"`
	EndedActivities   int64 `json:"ended_activities"`
}
