// Package dto DIY 前端页面中台数据传输对象 - 统计（stat 子域）
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// StatInfo 统计详情响应
type StatInfo struct {
	ID              uint      `json:"id"`
	PageID          uint      `json:"page_id"`
	ViewCount       int       `json:"view_count"`
	ClickCount      int       `json:"click_count"`
	ConversionCount int       `json:"conversion_count"`
	StatDate        time.Time `json:"stat_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// StatSummaryRequest 统计汇总请求
type StatSummaryRequest struct {
	PageID    uint      `form:"page_id" json:"page_id"`
	StartDate time.Time `form:"start_date" json:"start_date" time_format:"2006-01-02"`
	EndDate   time.Time `form:"end_date" json:"end_date" time_format:"2006-01-02"`
	utils.Pagination
}

// StatSummary 统计汇总响应
type StatSummary struct {
	PageID          uint   `json:"page_id"`
	TotalView       int64  `json:"total_view"`
	TotalClick      int64  `json:"total_click"`
	TotalConversion int64  `json:"total_conversion"`
	DailyList       []StatInfo `json:"daily_list"`
}

// RecordViewRequest 记录浏览请求
type RecordViewRequest struct {
	PageID uint `json:"page_id" binding:"required"`
}

// RecordClickRequest 记录点击请求
type RecordClickRequest struct {
	PageID uint `json:"page_id" binding:"required"`
}

// RecordConversionRequest 记录转化请求
type RecordConversionRequest struct {
	PageID uint `json:"page_id" binding:"required"`
}
