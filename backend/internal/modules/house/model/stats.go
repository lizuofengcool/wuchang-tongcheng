// Package model 数据统计 + 房贷计算配置表
// HouseStatistic 数据统计；HouseMortgage 房贷配置
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 统计类型常量 ===
const (
	StatTypeHouse      = "house"      // 按房源统计
	StatTypeCommunity  = "community"  // 按小区统计
	StatTypeAgent      = "agent"      // 按经纪人统计
	StatTypeRegion     = "region"     // 按地区统计
	StatTypeCategory   = "category"   // 按分类统计
	StatTypeOverall    = "overall"    // 全局统计
)

// === 房贷类型常量 ===
const (
	MortgageTypeCommercial    = "commercial"     // 商业贷款
	MortgageTypeProvidentFund = "provident_fund" // 公积金贷款
	MortgageTypeCombined      = "combined"       // 组合贷款
	MortgageTypeFull          = "full"           // 全款
)

// === 房贷状态常量 ===
const (
	MortgageStatusDisabled = 0 // 禁用
	MortgageStatusEnabled  = 1 // 启用
)

// HouseStatistic 数据统计表
type HouseStatistic struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	RegionID         uint      `gorm:"not null;default:1;index" json:"region_id"`                       // 地区 ID
	StatDate         time.Time `gorm:"type:date;not null;uniqueIndex:uniq_house_stats_date_type_target" json:"stat_date"` // 统计日期
	StatType         string    `gorm:"size:32;not null;index;uniqueIndex:uniq_house_stats_date_type_target" json:"stat_type"` // house/community/agent/region/category/overall
	TargetID         uint      `gorm:"not null;default:0;index;uniqueIndex:uniq_house_stats_date_type_target" json:"target_id"` // 目标 ID
	TargetName       string    `gorm:"size:128;not null;default:''" json:"target_name"`                // 目标名称
	ImpressionCount  int       `gorm:"not null;default:0" json:"impression_count"`                      // 曝光数
	ClickCount       int       `gorm:"not null;default:0" json:"click_count"`                           // 点击数
	FavCount         int       `gorm:"not null;default:0" json:"fav_count"`                             // 收藏数
	ContactCount     int       `gorm:"not null;default:0" json:"contact_count"`                         // 联系数
	ViewingCount     int       `gorm:"not null;default:0" json:"viewing_count"`                         // 看房数
	DealCount        int       `gorm:"not null;default:0" json:"deal_count"`                            // 成交数
	ConversionRate   float64   `gorm:"type:decimal(5,2);default:0" json:"conversion_rate"`              // 转化率
	AvgSalePrice     float64   `gorm:"type:decimal(10,2);default:0" json:"avg_sale_price"`              // 平均售价
	AvgRentPrice     float64   `gorm:"type:decimal(10,2);default:0" json:"avg_rent_price"`              // 平均租金
	AvgDealDays      int       `gorm:"not null;default:0" json:"avg_deal_days"`                         // 平均成交周期（天）
}

// TableName 表名（house_ 前缀）
func (HouseStatistic) TableName() string { return "house_statistics" }

// HouseMortgage 房贷计算配置表
type HouseMortgage struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name            string  `gorm:"size:64;not null" json:"name"`                                  // 方案名
	Code            string  `gorm:"size:64;not null;uniqueIndex" json:"code"`                       // 方案编码
	LoanType        string  `gorm:"size:32;not null;default:'commercial';index" json:"loan_type"`   // commercial/provident_fund/combined/full
	MinDownPayment  float64 `gorm:"type:decimal(5,2);default:30.00" json:"min_down_payment"`        // 最低首付比例（%）
	MaxDownPayment  float64 `gorm:"type:decimal(5,2);default:100.00" json:"max_down_payment"`       // 最高首付比例（%）
	InterestRate    float64 `gorm:"type:decimal(6,4);default:0" json:"interest_rate"`               // 利率（小数 0.0490=4.90%）
	MaxPeriods      int     `gorm:"not null;default:360" json:"max_periods"`                        // 最大期数（月）
	MaxAmount       float64 `gorm:"type:decimal(14,2);default:0" json:"max_amount"`                 // 最大贷款额
	Description     string  `gorm:"size:500;not null;default:''" json:"description"`               // 描述
	Sort            int     `gorm:"not null;default:0;index" json:"sort"`                           // 排序
	Status          int     `gorm:"default:1;index" json:"status"`                                  // 0禁用 1启用
	IsHot           bool    `gorm:"not null;default:false;index" json:"is_hot"`                     // 是否热门
	UseCount        int     `gorm:"not null;default:0" json:"use_count"`                            // 使用次数
}

// TableName 表名（house_ 前缀）
func (HouseMortgage) TableName() string { return "house_mortgages" }
