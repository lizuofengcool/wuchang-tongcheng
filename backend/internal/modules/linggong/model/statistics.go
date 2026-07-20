// Package model 数据统计表
// 岗位/雇主/求职者/技能/地区/平台多维度统计
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 统计类型常量 ===
const (
	StatTypeLinggong   = "linggong"   // 岗位统计
	StatTypeEmployer   = "employer"   // 雇主统计
	StatTypeWorker     = "worker"     // 求职者统计
	StatTypeSkill      = "skill"      // 技能统计
	StatTypeRegion     = "region"     // 地区统计
	StatTypePlatform   = "platform"   // 平台统计
	StatTypeTask       = "task"       // 任务统计
	StatTypeCategory   = "category"   // 分类统计
)

// LinggongStatistic 数据统计表
type LinggongStatistic struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	StatDate         time.Time `gorm:"type:date;not null;index;uniqueIndex:uniq_linggong_stats_date_type_target,priority:1" json:"stat_date"` // 统计日期
	StatType         string    `gorm:"size:32;not null;index;uniqueIndex:uniq_linggong_stats_date_type_target,priority:2" json:"stat_type"`     // linggong/employer/worker/skill/region/platform/task/category
	TargetID         uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_linggong_stats_date_type_target,priority:3" json:"target_id"`     // 目标 ID
	TargetName       string    `gorm:"size:128;not null;default:''" json:"target_name"`                                                       // 目标名称
	ImpressionCount  int       `gorm:"not null;default:0" json:"impression_count"`                                                            // 曝光数
	ClickCount       int       `gorm:"not null;default:0" json:"click_count"`                                                                 // 点击数
	FavCount         int       `gorm:"not null;default:0" json:"fav_count"`                                                                   // 收藏数
	ContactCount     int       `gorm:"not null;default:0" json:"contact_count"`                                                               // 联系数
	ApplicationCount int       `gorm:"not null;default:0" json:"application_count"`                                                          // 报名数
	HiredCount       int       `gorm:"not null;default:0" json:"hired_count"`                                                                 // 录用数
	CompletedCount   int       `gorm:"not null;default:0" json:"completed_count"`                                                            // 完成数
	DealCount        int       `gorm:"not null;default:0" json:"deal_count"`                                                                  // 成交数
	ConversionRate   float64   `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"`                                                    // 转化率
	TotalSalary      float64   `gorm:"type:decimal(12,2);default:0" json:"total_salary"`                                                       // 薪资总额
	AvgSalary        float64   `gorm:"type:decimal(12,2);default:0" json:"avg_salary"`                                                        // 平均薪资
	AvgDealDays      int       `gorm:"not null;default:0" json:"avg_deal_days"`                                                               // 平均完成周期
}

// TableName 表名（linggong_ 前缀）
func (LinggongStatistic) TableName() string { return "linggong_statistics" }
