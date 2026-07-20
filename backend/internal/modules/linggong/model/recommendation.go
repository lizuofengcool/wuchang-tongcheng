// Package model 推荐岗位表（对标斗米智能推荐 + 猪八戒威客匹配）
// 岗位推荐 + 求职者推荐 + 热门推荐
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 推荐类型常量 ===
const (
	RecTypeLinggongToWorker = "linggong_to_worker" // 岗位推人
	RecTypeWorkerToLinggong = "worker_to_linggong" // 人推岗位
	RecTypeSimilar          = "similar"             // 相似岗位
	RecTypeNearby           = "nearby"              // 附近岗位
	RecTypeRecentlyViewed   = "recently_viewed"     // 看过又推荐
	RecTypeHot              = "hot"                  // 热门推荐
	RecTypeNew              = "new"                  // 最新推荐
	RecTypeBySkill          = "by_skill"            // 技能匹配
	RecTypeByCategory       = "by_category"         // 分类匹配
)

// === 推荐来源常量 ===
const (
	RecSourceAI     = "ai"       // AI 推荐
	RecSourceManual = "manual"   // 人工推荐
	RecSourceHot    = "hot"      // 热门
	RecSourceBehavior = "behavior" // 行为推荐
)

// === 推荐状态常量 ===
const (
	RecStatusPending   = 0 // 待展示
	RecStatusShown     = 1 // 已展示
	RecStatusClicked   = 2 // 已点击
	RecStatusApplied   = 3 // 已报名
	RecStatusDismissed = 4 // 已忽略
	RecStatusExpired   = 5 // 已过期
)

// LinggongRecommendation 推荐岗位表
type LinggongRecommendation struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	UserID          uint       `gorm:"not null;index;uniqueIndex:uniq_linggong_recs_user_target_type" json:"user_id"`                   // 用户 ID
	LinggongID     uint       `gorm:"not null;index;uniqueIndex:uniq_linggong_recs_user_target_type" json:"linggong_id"`             // 岗位 ID
	RecType         string     `gorm:"size:32;not null;default:'linggong_to_worker';index;uniqueIndex:uniq_linggong_recs_user_target_type" json:"rec_type"` // linggong_to_worker/worker_to_linggong/similar/nearby/recently_viewed/hot/new
	Source          string     `gorm:"size:32;not null;default:'ai';index" json:"source"`          // ai/manual/hot/behavior
	Score           float64    `gorm:"type:decimal(5,2);default:0;index" json:"score"`            // 推荐评分
	Reason          string     `gorm:"size:500;not null;default:''" json:"reason"`                 // 推荐理由
	SalaryMatch     float64    `gorm:"type:decimal(5,2);default:0" json:"salary_match"`             // 薪资匹配度
	SkillMatch      float64    `gorm:"type:decimal(5,2);default:0" json:"skill_match"`            // 技能匹配度
	LocationMatch   float64    `gorm:"type:decimal(5,2);default:0" json:"location_match"`         // 地理匹配度
	TimeMatch       float64    `gorm:"type:decimal(5,2);default:0" json:"time_match"`              // 时间匹配度
	CreditMatch     float64    `gorm:"type:decimal(5,2);default:0" json:"credit_match"`           // 信用匹配度
	Status          int        `gorm:"default:0;index" json:"status"`                              // 0待展示 1已展示 2点击 3报名 4忽略 5过期
	ClickedAt       *time.Time `gorm:"index" json:"clicked_at"`                                    // 点击时间
	AppliedAt       *time.Time `gorm:"index" json:"applied_at"`                                    // 报名时间
	ViewedAt        *time.Time `gorm:"index" json:"viewed_at"`                                     // 查看时间
	DismissedAt     *time.Time `gorm:"index" json:"dismissed_at"`                                  // 忽略时间
	ExpiredAt       *time.Time `gorm:"index" json:"expired_at"`                                    // 过期时间
}

// TableName 表名（linggong_ 前缀）
func (LinggongRecommendation) TableName() string { return "linggong_recommendations" }
