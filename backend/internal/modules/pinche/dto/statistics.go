// Package dto 同城拼车出行数据传输对象 - 统计
// 依据 v3.2.1 架构方案：日统计/周统计/月统计/总统计
// 对标哈啰出行/嘀嗒出行 数据分析
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// StatisticInfo 统计详情响应
type StatisticInfo struct {
	ID                uint      `json:"id"`
	RegionID          uint      `json:"region_id"`
	StatDate          time.Time `json:"stat_date"`
	StatType          string    `json:"stat_type"`
	StatTypeText      string    `json:"stat_type_text"`
	UserID            *uint     `json:"user_id"`
	TotalTrips        int       `json:"total_trips"`
	CompletedTrips    int       `json:"completed_trips"`
	CancelledTrips    int       `json:"cancelled_trips"`
	TotalBookings     int       `json:"total_bookings"`
	CompletedBookings int       `json:"completed_bookings"`
	TotalRevenue      float64   `json:"total_revenue"`
	TotalRefund       float64   `json:"total_refund"`
	AvgRating         float64   `json:"avg_rating"`
	AvgPrice          float64   `json:"avg_price"`
	TotalDistance     float64   `json:"total_distance"`
	TotalDuration     int       `json:"total_duration"`
	TotalPassengers   int       `json:"total_passengers"`
	TotalDrivers      int       `json:"total_drivers"`
	CreatedAt         time.Time `json:"created_at"`
}

// StatisticListRequest 统计列表查询请求
type StatisticListRequest struct {
	StatType  string `form:"stat_type" json:"stat_type"`
	UserID    uint   `form:"user_id" json:"user_id"`
	StartDate string `form:"start_date" json:"start_date"`
	EndDate   string `form:"end_date" json:"end_date"`
	utils.Pagination
}

// StatisticOverviewResponse 平台总览统计响应
type StatisticOverviewResponse struct {
	TotalTrips        int64   `json:"total_trips"`
	CompletedTrips    int64   `json:"completed_trips"`
	CancelledTrips    int64   `json:"cancelled_trips"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalRefund       float64 `json:"total_refund"`
	CompletionRate    float64 `json:"completion_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
}

// StatisticUpsertRequest 创建/更新统计请求（M 端维护）
type StatisticUpsertRequest struct {
	StatDate          time.Time `json:"stat_date" binding:"required"`
	StatType          string    `json:"stat_type" binding:"omitempty,oneof=daily weekly monthly total"`
	UserID            *uint     `json:"user_id"`
	TotalTrips        int       `json:"total_trips"`
	CompletedTrips    int       `json:"completed_trips"`
	CancelledTrips    int       `json:"cancelled_trips"`
	TotalBookings     int       `json:"total_bookings"`
	CompletedBookings int       `json:"completed_bookings"`
	TotalRevenue      float64   `json:"total_revenue"`
	TotalRefund       float64   `json:"total_refund"`
	AvgRating         float64   `json:"avg_rating"`
	AvgPrice          float64   `json:"avg_price"`
	TotalDistance     float64   `json:"total_distance"`
	TotalDuration     int       `json:"total_duration"`
	TotalPassengers   int       `json:"total_passengers"`
	TotalDrivers      int       `json:"total_drivers"`
}
