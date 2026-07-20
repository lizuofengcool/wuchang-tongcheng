// Package model 收藏表 + 收藏分组
// 支持商户/团购/优惠券收藏，分组管理
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Favorite 收藏表
type Dh114Favorite struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID       uint   `gorm:"not null;index;uniqueIndex:uniq_dh114_favorites_user_target_type" json:"user_id"`         // 用户 ID
	Dh114ID      uint   `gorm:"not null;index;uniqueIndex:uniq_dh114_favorites_user_target_type" json:"dh114_id"`      // 商户 ID
	BusinessID   uint   `gorm:"not null;default:0;index" json:"business_id"`                                          // 商户详情 ID
	FavoriteType string `gorm:"size:32;not null;default:'business';index;uniqueIndex:uniq_dh114_favorites_user_target_type" json:"favorite_type"` // business/groupbuy/coupon
	GroupID      uint   `gorm:"not null;default:0;index" json:"group_id"`                                            // 收藏分组 ID
	Remark       string `gorm:"size:200;not null;default:''" json:"remark"`                                          // 备注
}

// TableName 表名
func (Dh114Favorite) TableName() string { return "dh114_favorites" }
