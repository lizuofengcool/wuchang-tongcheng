// Package model DIY 前端页面中台 - DIY 组件模型（diy_components 表）
// 组件库：基础组件/布局组件/业务组件
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 组件分类常量 ===
const (
	ComponentCategoryBasic    = "basic"    // 基础组件（文本/图片/按钮等）
	ComponentCategoryLayout   = "layout"   // 布局组件（栅格/容器/Tab 等）
	ComponentCategoryBusiness = "business" // 业务组件（商品列表/店铺信息/活动卡片等）
)

// === 组件状态常量 ===
const (
	ComponentStatusDisabled = 0 // 禁用
	ComponentStatusEnabled  = 1 // 启用
)

// Component DIY 组件模型（diy_components 表）
type Component struct {
	database.BaseModel // 含 id/created_at/updated_at/deleted_at

	// === 基础信息 ===
	Name        string `gorm:"size:64;not null;default:'';index" json:"name"`        // 组件名称
	Code        string `gorm:"size:64;not null;default:'';uniqueIndex" json:"code"` // 组件编码（唯一）
	Category    string `gorm:"size:32;not null;default:'basic';index" json:"category"` // basic/layout/business
	Description string `gorm:"type:text" json:"description"`                          // 组件描述
	Thumbnail   string `gorm:"size:500;not null;default:''" json:"thumbnail"`         // 缩略图 URL
	Status      int    `gorm:"not null;default:1;index" json:"status"`                // 0禁用 1启用

	// === JSONB 配置 ===
	Config JSONB `gorm:"type:jsonb" json:"config"` // 组件配置模板（默认属性 schema）
}

// TableName 表名
func (Component) TableName() string { return "diy_components" }
