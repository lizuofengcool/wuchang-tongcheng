// Package dto love 相亲交友数据传输对象 - 拉黑
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// LoveBlockInfo 拉黑记录响应
type LoveBlockInfo struct {
	ID               uint      `json:"id"`
	UserID           uint      `json:"user_id"`
	LoveID           uint      `json:"love_id"`
	BlockedUserID    uint      `json:"blocked_user_id"`
	BlockedLoveID    uint      `json:"blocked_love_id"`
	BlockedNickname  string    `json:"blocked_nickname"`
	BlockedAvatar    string    `json:"blocked_avatar"`
	Reason           string    `json:"reason"`
	ReportID         uint      `json:"report_id"`
	Status           int       `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// CreateLoveBlockRequest 拉黑请求
type CreateLoveBlockRequest struct {
	BlockedUserID uint   `json:"blocked_user_id" binding:"required"`
	BlockedLoveID uint   `json:"blocked_love_id" binding:"required"`
	Reason        string `json:"reason" binding:"max=255"`
	ReportID      uint   `json:"report_id"`
}

// LoveBlockListRequest 拉黑列表请求
type LoveBlockListRequest struct {
	UserID uint `form:"user_id" json:"user_id"`
	utils.Pagination
}
