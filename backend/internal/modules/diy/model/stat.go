// Package model DIY 前端页面中台 - DIY 页面统计模型（diy_page_stats 表）
// 按日期汇总页面浏览/点击/转化数据
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PageStat DIY 页面统计模型（diy_page_stats 表）
type PageStat struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at

	// === 关联 ===
	PageID uint `gorm:"index;not null;default:0" json:"page_id"` // 页面 ID

	// === 统计数据 ===
	ViewCount       int       `gorm:"not null;default:0" json:"view_count"`        // 浏览数
	ClickCount      int       `gorm:"not null;default:0" json:"click_count"`       // 点击数
	ConversionCount int       `gorm:"not null;default:0" json:"conversion_count"`  // 转化数
	StatDate        time.Time `gorm:"type:date;index" json:"stat_date"`            // 统计日期
}

// TableName 表名
func (PageStat) TableName() string { return "diy_page_stats" }
