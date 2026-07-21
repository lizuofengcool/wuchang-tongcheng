// Package model DIY 前端页面中台 - DIY 页面模型（diy_pages 表）
// 依据架构设计 4.12：拖拽生成首页/专题页/店铺页/活动页
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 页面类型常量 ===
const (
	PageTypeHome     = "home"     // 首页
	PageTypeTopic    = "topic"    // 专题页
	PageTypeShop     = "shop"     // 店铺页
	PageTypeActivity = "activity" // 活动页
)

// === 页面状态常量 ===
const (
	PageStatusDraft    = 0 // 草稿
	PageStatusPublish  = 1 // 已发布
	PageStatusOffline  = 2 // 已下线
)

// Page DIY 页面模型（diy_pages 表）
type Page struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	// === 基础信息 ===
	Title   string `gorm:"size:100;not null;default:'';index" json:"title"`   // 页面标题
	Type    string `gorm:"size:32;not null;default:'home';index" json:"type"` // home/topic/shop/activity
	Slug    string `gorm:"size:100;not null;default:'';index" json:"slug"`    // URL Slug（按 slug 获取已发布页面）
	Status  int    `gorm:"not null;default:0;index" json:"status"`            // 0草稿 1发布 2下线
	UserID  uint   `gorm:"index;not null;default:0" json:"user_id"`           // 创建者 ID
	BizID   uint   `gorm:"index;not null;default:0" json:"biz_id"`            // 业务 ID（如店铺 ID/活动 ID，0 表示无关联）

	// === 时间 ===
	PublishedAt *time.Time `gorm:"index" json:"published_at"` // 发布时间

	// === JSONB 配置 ===
	Components JSONB `gorm:"type:jsonb" json:"components"` // 组件配置（拖拽组件数组）
	Settings   JSONB `gorm:"type:jsonb" json:"settings"`   // 页面设置（如 SEO/背景/全局样式）
}

// TableName 表名
func (Page) TableName() string { return "diy_pages" }
