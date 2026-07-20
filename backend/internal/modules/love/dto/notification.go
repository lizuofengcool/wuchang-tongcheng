// Package dto love 相亲交友数据传输对象 - 通知
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveNotificationInfo 通知响应
type LoveNotificationInfo struct {
	ID               uint       `json:"id"`
	UserID           uint       `json:"user_id"`
	LoveID           uint       `json:"love_id"`
	Type             string     `json:"type"`
	TypeText         string     `json:"type_text"`
	Title            string     `json:"title"`
	Content          string     `json:"content"`
	FromUserID       uint       `json:"from_user_id"`
	FromUserNickname string     `json:"from_user_nickname"`
	FromUserAvatar   string     `json:"from_user_avatar"`
	TargetType       string     `json:"target_type"`
	TargetID         uint       `json:"target_id"`
	ActionURL        string     `json:"action_url"`
	Extra            interface{} `json:"extra"`
	IsRead           bool       `json:"is_read"`
	ReadAt           *time.Time `json:"read_at"`
	Status           int        `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
}

// LoveNotificationListRequest 通知列表请求
type LoveNotificationListRequest struct {
	Type   string `form:"type" json:"type"`
	IsRead *bool  `form:"is_read" json:"is_read"`
	utils.Pagination
}

// LoveNotificationBatchReadRequest 批量已读请求
type LoveNotificationBatchReadRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// LoveNotificationStatsResponse 通知统计响应
type LoveNotificationStatsResponse struct {
	TotalNotifications int64 `json:"total_notifications"`
	UnreadCount        int64 `json:"unread_count"`
	LikeCount          int64 `json:"like_count"`
	SuperLikeCount     int64 `json:"super_like_count"`
	MatchCount         int64 `json:"match_count"`
	VisitCount         int64 `json:"visit_count"`
	GiftCount          int64 `json:"gift_count"`
	MessageCount       int64 `json:"message_count"`
	SystemCount        int64 `json:"system_count"`
}
