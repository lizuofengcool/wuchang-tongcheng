// Package model 车型分类表（对标懂车帝）
// 轿车/SUV/MPV/新能源/跑车/皮卡
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 分类状态常量 ===
const (
	CategoryStatusDraft     = 0 // 草稿
	CategoryStatusPublished = 1 // 已发布
	CategoryStatusOffline   = 2 // 已下架
)

// CarCategory 车型分类表
type CarCategory struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string `gorm:"size:64;not null" json:"name"`                                  // 分类名
	Code        string `gorm:"size:64;not null;uniqueIndex" json:"code"`                      // 分类编码
	ParentID    uint   `gorm:"not null;default:0;index" json:"parent_id"`                     // 父分类 ID（0 为一级）
	Level       int    `gorm:"not null;default:1;index" json:"level"`                         // 层级 1/2/3
	CarType     string `gorm:"size:32;not null;default:'';index" json:"car_type"`             // 关联车型：sedan/suv/mpv/new_energy/sports/truck
	Icon        string `gorm:"size:64;not null;default:''" json:"icon"`                       // 图标
	Color       string `gorm:"size:32;not null;default:''" json:"color"`                      // 颜色
	Description string `gorm:"size:500;not null;default:''" json:"description"`               // 描述
	Sort        int    `gorm:"not null;default:0;index" json:"sort"`                          // 排序
	Status      int    `gorm:"default:1;index" json:"status"`                                 // 0草稿 1已发布 2下架
	CarCount    int    `gorm:"not null;default:0" json:"car_count"`                           // 车源数
}

// TableName 表名（car_ 前缀）
func (CarCategory) TableName() string { return "car_categories" }
