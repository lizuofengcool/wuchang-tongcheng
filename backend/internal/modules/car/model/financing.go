// Package model 分期付款方案表（对标易鑫车贷/毛豆新车）
// 首付/月供/利率/期数/资方
package model

import (
	"wuchang-tongcheng/internal/pkg/database"
)

// === 分期类型常量 ===
const (
	FinancingTypeLoan           = "loan"            // 贷款购车
	FinancingTypeLease          = "lease"           // 融资租赁
	FinancingTypeBalloon        = "balloon"         // 气球贷（尾款）
	FinancingTypeZeroDown       = "zero_down"       // 零首付
	FinancingTypeReplace        = "replace_loan"    // 置换贷款
)

// === 状态常量 ===
const (
	FinancingStatusDraft     = 0 // 草稿
	FinancingStatusPublished = 1 // 已发布
	FinancingStatusOffline   = 2 // 已下架
)

// CarFinancing 分期付款方案表
type CarFinancing struct {
	database.BaseModel // id/created_at/updated_at/deleted_at
	Name              string  `gorm:"size:64;not null" json:"name"`                                  // 方案名
	Code              string  `gorm:"size:64;not null;uniqueIndex" json:"code"`                      // 方案编码
	FinancingType     string  `gorm:"size:32;not null;default:'loan';index" json:"financing_type"`    // loan/lease/balloon/zero_down/replace_loan
	MinDownPayment    float64 `gorm:"type:decimal(5,2);default:20.00" json:"min_down_payment"`       // 最低首付比例（%）
	MaxDownPayment    float64 `gorm:"type:decimal(5,2);default:80.00" json:"max_down_payment"`       // 最高首付比例（%）
	InterestRate      float64 `gorm:"type:decimal(6,4);default:0" json:"interest_rate"`              // 月利率
	AnnualRate        float64 `gorm:"type:decimal(6,4);default:0" json:"annual_rate"`                // 年化利率
	MinPeriods        int     `gorm:"not null;default:12" json:"min_periods"`                        // 最短期数
	MaxPeriods        int     `gorm:"not null;default:60" json:"max_periods"`                        // 最长期数
	MaxAmount         float64 `gorm:"type:decimal(14,2);default:0" json:"max_amount"`                // 最高贷款额
	Provider          string  `gorm:"size:128;not null;default:'';index" json:"provider"`            // 资方
	Description       string  `gorm:"size:500;not null;default:''" json:"description"`               // 方案说明
	Sort              int     `gorm:"not null;default:0;index" json:"sort"`                          // 排序
	Status            int     `gorm:"default:1;index" json:"status"`                                 // 0草稿 1已发布 2下架
	IsHot             bool    `gorm:"not null;default:false;index" json:"is_hot"`                    // 热门
	UseCount          int     `gorm:"not null;default:0" json:"use_count"`                           // 使用次数
}

// TableName 表名（car_ 前缀）
func (CarFinancing) TableName() string { return "car_financing" }
