// Package model 数据统计 + 薪资范围配置（对标 BOSS直聘）
// 曝光/点击/投递/面试/Offer/入职转化 + 月/年/日/时薪资
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 统计类型常量 ===
const (
	StatTypeJob     = "job"     // 职位维度
	StatTypeCompany = "company" // 公司维度
	StatTypeCategory = "category" // 分类维度
	StatTypeRegion  = "region"   // 地区维度
	StatTypeOverall = "overall" // 全局
	StatTypeUser    = "user"     // 用户维度
)

// JobStatistic 数据统计表
type JobStatistic struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	StatDate         time.Time `gorm:"type:date;not null;index;uniqueIndex:uniq_job_stats_date_type_target" json:"stat_date"` // 统计日期
	StatType         string    `gorm:"size:32;not null;index;uniqueIndex:uniq_job_stats_date_type_target" json:"stat_type"` // 统计维度
	TargetID         uint      `gorm:"not null;index;uniqueIndex:uniq_job_stats_date_type_target" json:"target_id"`         // 目标 ID（职位/公司/分类/地区）
	TargetName       string    `gorm:"size:128" json:"target_name"`                                                          // 目标名称
	ImpressionCount  int       `gorm:"default:0" json:"impression_count"`                                                    // 曝光数
	ClickCount       int       `gorm:"default:0" json:"click_count"`                                                          // 点击数
	FavCount         int       `gorm:"default:0" json:"fav_count"`                                                            // 收藏数
	DeliverCount     int       `gorm:"default:0" json:"deliver_count"`                                                        // 投递数
	InterviewCount   int       `gorm:"default:0" json:"interview_count"`                                                      // 面试数
	OfferCount       int       `gorm:"default:0" json:"offer_count"`                                                          // Offer 数
	OnboardingCount  int       `gorm:"default:0" json:"onboarding_count"`                                                      // 入职数
	ConversionRate   float64   `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"`                                  // 转化率（投递→入职）
	AvgSalary        float64   `gorm:"type:decimal(12,2);default:0" json:"avg_salary"`                                       // 平均薪资
	MedianSalary     float64   `gorm:"type:decimal(12,2);default:0" json:"median_salary"`                                    // 中位数薪资
	Retention30D     int       `gorm:"default:0" json:"retention_30d"`                                                       // 30 天留存
	Retention90D     int       `gorm:"default:0" json:"retention_90d"`                                                        // 90 天留存
}

// TableName 表名（job_ 前缀）
func (JobStatistic) TableName() string { return "job_statistics" }

// === 薪资周期常量 ===
const (
	SalaryPeriodMonth = "month" // 月薪
	SalaryPeriodYear  = "year"  // 年薪
	SalaryPeriodDay   = "day"   // 日薪
	SalaryPeriodHour  = "hour"  // 时薪
)

// === 薪资状态常量 ===
const (
	SalaryRangeStatusDisabled = 0 // 禁用
	SalaryRangeStatusEnabled  = 1 // 启用
)

// JobSalaryRange 薪资范围配置表
type JobSalaryRange struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name        string  `gorm:"size:64;not null" json:"name"`                                                // 名称
	Code        string  `gorm:"size:64;not null;uniqueIndex:uniq_job_salary_ranges_code" json:"code"`        // 编码
	MinAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"min_amount"`                              // 最低金额
	MaxAmount   float64 `gorm:"type:decimal(12,2);default:0" json:"max_amount"`                              // 最高金额
	Currency    string  `gorm:"size:8;default:'CNY'" json:"currency"`                                          // 币种
	Period      string  `gorm:"size:16;default:'month'" json:"period"`                                       // 周期 month/year/day/hour
	Description string  `gorm:"size:500" json:"description"`                                                 // 描述
	Sort        int     `gorm:"default:0;index" json:"sort"`                                                  // 排序
	Status      int     `gorm:"default:1;index" json:"status"`                                                // 0禁用 1启用
	IsHot       bool    `gorm:"default:false;index" json:"is_hot"`                                          // 是否热门
	UseCount    int     `gorm:"default:0" json:"use_count"`                                                  // 使用次数
}

// TableName 表名（job_ 前缀）
func (JobSalaryRange) TableName() string { return "job_salary_ranges" }
