// Package model 菜单/服务项目表
// 菜品/服务项目/价格/图片
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Menu 菜单/服务项目表
type Dh114Menu struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	Dh114ID    uint   `gorm:"not null;index" json:"dh114_id"`                                       // 商户 ID
	BusinessID uint   `gorm:"not null;default:0;index" json:"business_id"`                          // 商户详情 ID
	MenuType   string `gorm:"size:32;not null;default:'dish';index" json:"menu_type"`               // dish/service
	Name       string `gorm:"size:128;not null" json:"name"`                                        // 名称
	Description string `gorm:"size:500;not null;default:''" json:"description"`                     // 描述
	Price      float64 `gorm:"type:decimal(12,2);default:0;index" json:"price"`                     // 价格
	OriginalPrice float64 `gorm:"type:decimal(12,2);default:0" json:"original_price"`              // 原价
	Image      string `gorm:"size:255;not null;default:''" json:"image"`                            // 图片 URL
	Unit       string `gorm:"size:32;not null;default:''" json:"unit"`                              // 单位（份/次/小时等）
	Sort       int    `gorm:"not null;default:0;index" json:"sort"`                                // 排序
	Status     int    `gorm:"default:1;index" json:"status"`                                        // 0下架 1上架
	OrderCount int    `gorm:"not null;default:0" json:"order_count"`                              // 销量
	Tags       JSONB  `gorm:"type:jsonb" json:"tags"`                                              // 标签 JSON
	IsSignature bool  `gorm:"not null;default:false;index" json:"is_signature"`                    // 是否招牌
}

// TableName 表名
func (Dh114Menu) TableName() string { return "dh114_menus" }
