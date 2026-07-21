// Package model DIY 前端页面中台 - DIY 模板模型（diy_templates 表）
// 页面模板：可应用于新页面，也可将现有页面保存为模板
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 模板分类常量 ===
const (
	TemplateCategoryHome     = "home"     // 首页模板
	TemplateCategoryTopic    = "topic"    // 专题页模板
	TemplateCategoryShop     = "shop"     // 店铺页模板
	TemplateCategoryActivity = "activity" // 活动页模板
)

// === 模板状态常量 ===
const (
	TemplateStatusDisabled = 0 // 禁用
	TemplateStatusEnabled  = 1 // 启用
)

// Template DIY 模板模型（diy_templates 表）
type Template struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at

	// === 基础信息 ===
	Name        string `gorm:"size:100;not null;default:'';index" json:"name"`        // 模板名称
	Thumbnail   string `gorm:"size:500;not null;default:''" json:"thumbnail"`         // 缩略图 URL
	Description string `gorm:"type:text" json:"description"`                          // 模板描述
	Category    string `gorm:"size:32;not null;default:'home';index" json:"category"` // home/topic/shop/activity
	Status      int    `gorm:"not null;default:1;index" json:"status"`                // 0禁用 1启用

	// === JSONB 配置 ===
	Pages JSONB `gorm:"type:jsonb" json:"pages"` // 模板页面配置（包含 components/settings）
}

// TableName 表名
func (Template) TableName() string { return "diy_templates" }
