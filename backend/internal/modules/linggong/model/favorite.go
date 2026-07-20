// Package model 收藏表
// 求职者收藏岗位 + 雇主收藏求职者
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 收藏类型常量 ===
const (
	FavoriteTypeLinggong = "linggong" // 岗位
	FavoriteTypeWorker   = "worker"   // 求职者
	FavoriteTypeEmployer = "employer" // 雇主
	FavoriteTypeTask     = "task"      // 任务
	FavoriteTypeSearch   = "search"    // 搜索条件
)

// LinggongFavorite 收藏表
type LinggongFavorite struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	UserID         uint   `gorm:"not null;index;uniqueIndex:uniq_linggong_favorites_user_target_type" json:"user_id"`                   // 用户 ID
	TargetID       uint   `gorm:"not null;index;uniqueIndex:uniq_linggong_favorites_user_target_type" json:"target_id"`              // 目标 ID
	FavoriteType   string `gorm:"size:32;not null;default:'linggong';index;uniqueIndex:uniq_linggong_favorites_user_target_type" json:"favorite_type"` // linggong/worker/employer/task/search
	Remark         string `gorm:"size:200;not null;default:''" json:"remark"`                       // 备注
	NotifyOnUpdate bool   `gorm:"not null;default:false" json:"notify_on_update"`                  // 更新时通知
}

// TableName 表名（linggong_ 前缀）
func (LinggongFavorite) TableName() string { return "linggong_favorites" }
