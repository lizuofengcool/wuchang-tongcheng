// Package dto DIY 前端页面中台数据传输对象 - 页面（page 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PageInfo 页面详情响应
type PageInfo struct {
	ID          uint       `json:"id"`
	RegionID    uint       `json:"region_id"`
	Title       string     `json:"title"`
	Type        string     `json:"type"`
	TypeText    string     `json:"type_text"`
	Slug        string     `json:"slug"`
	Status      int        `json:"status"`
	StatusText  string     `json:"status_text"`
	UserID      uint       `json:"user_id"`
	BizID       uint       `json:"biz_id"`
	PublishedAt *time.Time `json:"published_at"`
	Components  interface{} `json:"components"`
	Settings    interface{} `json:"settings"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreatePageRequest 创建页面请求
type CreatePageRequest struct {
	Title      string      `json:"title" binding:"required,max=100"`
	Type       string      `json:"type" binding:"omitempty,oneof=home topic shop activity"`
	Slug       string      `json:"slug" binding:"max=100"`
	BizID      uint        `json:"biz_id"`
	Components interface{} `json:"components"`
	Settings   interface{} `json:"settings"`
	Status     int         `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdatePageRequest 更新页面请求
type UpdatePageRequest struct {
	Title      *string      `json:"title" binding:"omitempty,max=100"`
	Type       *string      `json:"type" binding:"omitempty,oneof=home topic shop activity"`
	Slug       *string      `json:"slug" binding:"omitempty,max=100"`
	BizID      *uint        `json:"biz_id"`
	Components interface{}  `json:"components"`
	Settings   interface{}  `json:"settings"`
	Status     *int         `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// PageListRequest 页面列表请求
type PageListRequest struct {
	Keyword string `form:"keyword" json:"keyword"`
	Type    string `form:"type" json:"type"`
	Status  *int   `form:"status" json:"status"`
	UserID  uint   `form:"user_id" json:"user_id"`
	utils.Pagination
}

// PageListAdminRequest 管理后台页面列表请求（含 region_id）
type PageListAdminRequest struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	Keyword  string `form:"keyword" json:"keyword"`
	Type     string `form:"type" json:"type"`
	Status   *int   `form:"status" json:"status"`
	UserID   uint   `form:"user_id" json:"user_id"`
	utils.Pagination
}
