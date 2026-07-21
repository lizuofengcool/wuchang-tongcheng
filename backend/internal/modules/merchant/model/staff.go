// Package model 商户中台 - 商户员工表（merchant_staff）
// 对标美团/大众点评商家员工管理
// 一个店铺可拥有多个员工（owner/manager/clerk），权限通过 JSONB 配置
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Staff 商户员工表
type Staff struct {
	database.BaseModel // id/created_at/updated_at/deleted_at

	// === 关联 ===
	ShopID uint   `gorm:"index;not null" json:"shop_id"` // 所属商户 ID
	UserID uint   `gorm:"index;not null" json:"user_id"` // 关联用户 ID
	Role   string `gorm:"size:20;not null;default:'clerk';index" json:"role"` // owner/manager/clerk

	// === 权限（JSONB） ===
	Permissions JSONB `gorm:"type:jsonb" json:"permissions"` // 权限配置 JSON

	// === 状态 ===
	Status int `gorm:"default:1;index" json:"status"` // 1在职 2停用
}

// TableName 表名
func (Staff) TableName() string { return "merchant_staff" }
