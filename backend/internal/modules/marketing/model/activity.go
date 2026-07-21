// Package model 营销活动中台 - 营销活动模型（activity 子域）
// 依据架构设计 4.6：拼团/砍价/秒杀/抽奖
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 活动类型常量 ===
const (
	ActivityTypeGroupBuy = "groupbuy" // 拼团
	ActivityTypeBargain  = "bargain"  // 砍价
	ActivityTypeSeckill  = "seckill"  // 秒杀
	ActivityTypeLottery  = "lottery"  // 抽奖
)

// === 活动状态常量 ===
const (
	ActivityStatusDisabled  = 0 // 禁用
	ActivityStatusPending   = 1 // 待开始
	ActivityStatusOngoing   = 2 // 进行中
	ActivityStatusEnded     = 3 // 已结束
	ActivityStatusCancelled = 4 // 已取消
)

// Activity 营销活动模型（activities 表）
type Activity struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at（地区隔离）

	Title       string     `gorm:"size:100;not null;default:'';index" json:"title"`        // 活动标题
	Type        string     `gorm:"size:20;not null;default:'groupbuy';index" json:"type"`  // groupbuy/bargain/seckill/lottery
	Description string     `gorm:"type:text" json:"description"`                           // 活动描述
	CoverImage  string     `gorm:"size:500;not null;default:''" json:"cover_image"`        // 封面图
	StartAt     *time.Time `gorm:"index" json:"start_at"`                                  // 开始时间
	EndAt       *time.Time `gorm:"index" json:"end_at"`                                    // 结束时间
	Status      int        `gorm:"not null;default:1;index" json:"status"`                 // 0禁用 1待开始 2进行中 3已结束 4已取消
	Config      JSONB      `gorm:"type:jsonb" json:"config"`                               // 活动配置（JSONB，如拼团人数/砍价底价/秒杀库存/抽奖概率）
}

// TableName 表名
func (Activity) TableName() string { return "activities" }
