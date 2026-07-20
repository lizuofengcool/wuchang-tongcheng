// Package dto love 相亲交友数据传输对象 - 访客
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveVisitInfo 访客记录响应
type LoveVisitInfo struct {
	ID              uint       `json:"id"`
	LoveID          uint       `json:"love_id"`
	UserID          uint       `json:"user_id"`
	VisitorUserID   uint       `json:"visitor_user_id"`
	VisitorLoveID   uint       `json:"visitor_love_id"`
	VisitorNickname string     `json:"visitor_nickname"`
	VisitorAvatar   string     `json:"visitor_avatar"`
	VisitorGender   int        `json:"visitor_gender"`
	VisitorVerified bool       `json:"visitor_verified"`
	VisitType       string     `json:"visit_type"`
	Source          string     `json:"source"`
	Duration        int        `json:"duration"`
	IsHidden        bool       `json:"is_hidden"`
	IsRead          bool       `json:"is_read"`
	Status          int        `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateLoveVisitRequest 访客记录请求
type CreateLoveVisitRequest struct {
	TargetUserID uint   `json:"target_user_id" binding:"required"`
	TargetLoveID uint   `json:"target_love_id" binding:"required"`
	VisitType    string `json:"visit_type" binding:"omitempty,oneof=profile story photo"`
	Source       string `json:"source" binding:"omitempty,oneof=recommend nearby search story match profile other"`
	Duration     int    `json:"duration"`
}

// LoveVisitListRequest 访客列表请求
type LoveVisitListRequest struct {
	UserID         uint   `form:"user_id" json:"user_id"`
	VisitorUserID  uint   `form:"visitor_user_id" json:"visitor_user_id"`
	VisitType      string `form:"visit_type" json:"visit_type"`
	IsRead         *bool  `form:"is_read" json:"is_read"`
	utils.Pagination
}

// LoveVisitStatsResponse 访客统计响应
type LoveVisitStatsResponse struct {
	TotalVisitors    int64 `json:"total_visitors"`
	TodayVisitors    int64 `json:"today_visitors"`
	UnreadVisitors   int64 `json:"unread_visitors"`
	WeeklyVisitors   int64 `json:"weekly_visitors"`
	MonthlyVisitors  int64 `json:"monthly_visitors"`
}
