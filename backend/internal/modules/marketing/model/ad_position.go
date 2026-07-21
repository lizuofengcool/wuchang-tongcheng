// Package model 营销活动中台 - 广告位模型（ad 子域）
// 依据架构设计 4.6：首页Banner/列表置顶/详情广告
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 广告位状态常量 ===
const (
	AdStatusDisabled  = 0 // 禁用
	AdStatusEnabled   = 1 // 启用
	AdStatusScheduled = 2 // 待生效
	AdStatusExpired   = 3 // 已过期
)

// === 广告位位置编码常量 ===
const (
	AdPositionHomeBanner    = "home_banner"    // 首页 Banner
	AdPositionListTop       = "list_top"        // 列表置顶
	AdPositionDetailBanner  = "detail_banner"   // 详情页广告
	AdPositionCategoryTop   = "category_top"    // 分类页置顶
	AdPositionSearchTop     = "search_top"      // 搜索页置顶
	AdPositionPopup         = "popup"           // 弹窗广告
)

// AdPosition 广告位模型（ad_positions 表）
type AdPosition struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	PositionCode string     `gorm:"size:50;not null;default:'';index:idx_ad_positions_code" json:"position_code"` // 位置编码
	Title        string     `gorm:"size:100;not null;default:''" json:"title"`                                  // 广告标题
	ImageURL     string     `gorm:"size:500;not null;default:''" json:"image_url"`                              // 广告图片 URL
	LinkURL      string     `gorm:"size:500;not null;default:''" json:"link_url"`                                // 跳转链接
	Sort         int        `gorm:"not null;default:0;index" json:"sort"`                                       // 排序（升序）
	StartAt      *time.Time `gorm:"index" json:"start_at"`                                                       // 生效开始时间
	EndAt        *time.Time `gorm:"index" json:"end_at"`                                                         // 生效结束时间
	Status       int        `gorm:"not null;default:1;index" json:"status"`                                      // 0禁用 1启用 2待生效 3已过期
}

// TableName 表名
func (AdPosition) TableName() string { return "ad_positions" }
