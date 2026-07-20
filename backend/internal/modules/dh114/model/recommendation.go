// Package model 推荐商家表
// 首页推荐/分类推荐/附近推荐/个性化推荐
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Dh114Recommendation 推荐商家表
type Dh114Recommendation struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	UserID      uint   `gorm:"not null;default:0;index" json:"user_id"`                          // 推荐给的用户 ID（0 表示全员）
	Dh114ID     uint   `gorm:"not null;index" json:"dh114_id"`                                    // 商户 ID
	BusinessID  uint   `gorm:"not null;default:0;index" json:"business_id"`                      // 商户详情 ID
	RecommendType string `gorm:"size:32;not null;default:'home';index" json:"recommend_type"`   // home/category/nearby/personalized
	Position    int    `gorm:"not null;default:0;index" json:"position"`                        // 推荐位置
	Score       float64 `gorm:"type:decimal(5,2);default:0" json:"score"`                       // 推荐评分
	Reason      string `gorm:"size:500;not null;default:''" json:"reason"`                       // 推荐理由
	CategoryID  *uint  `gorm:"index" json:"category_id"`                                         // 关联分类 ID（用于分类推荐）
	ExpireAt    *time.Time `gorm:"index" json:"expire_at"`                                       // 推荐过期时间

	// === 状态 ===
	Status      int        `gorm:"default:0;index" json:"status"`                                // 0已展示 1已点击 2已联系 3已忽略
	ClickedAt   *time.Time `gorm:"index" json:"clicked_at"`                                      // 点击时间
	ContactedAt *time.Time `gorm:"index" json:"contacted_at"`                                    // 联系时间
	DismissedAt *time.Time `gorm:"index" json:"dismissed_at"`                                    // 忽略时间
}

// TableName 表名
func (Dh114Recommendation) TableName() string { return "dh114_recommendations" }
