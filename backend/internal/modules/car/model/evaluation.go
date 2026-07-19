// Package model 车辆评估表（对标瓜子/人人车）
// 市场价/收购价/零售价/批发价/折旧/相似成交
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 评估状态常量 ===
const (
	EvaluationStatusPending   = 0 // 待评估
	EvaluationStatusCompleted = 1 // 已完成
	EvaluationStatusExpired   = 2 // 已过期
	EvaluationStatusCanceled  = 3 // 已取消
)

// === 评估类型常量 ===
const (
	EvaluationTypeOnline  = "online"  // 在线评估
	EvaluationTypeOffline = "offline" // 线下评估
	EvaluationTypeAI      = "ai"      // AI 评估
	EvaluationTypeManual  = "manual"  // 人工评估
)

// === 评估师等级常量 ===
const (
	EvaluatorLevelJunior = "junior" // 初级
	EvaluatorLevelSenior = "senior" // 高级
	EvaluatorLevelMaster = "master" // 资深
)

// CarEvaluation 车辆评估表
type CarEvaluation struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	EvaluationNo       string     `gorm:"size:64;not null;uniqueIndex" json:"evaluation_no"`             // 评估单号
	CarID              uint       `gorm:"not null;index" json:"car_id"`                                  // 车源 ID
	ModelID            uint       `gorm:"not null;default:0;index" json:"model_id"`                      // 车型 ID
	EvaluatorID        uint       `gorm:"not null;index" json:"evaluator_id"`                            // 评估师 ID
	EvaluatorName      string     `gorm:"size:50;not null;default:''" json:"evaluator_name"`             // 评估师姓名
	EvaluatorLevel     string     `gorm:"size:16;not null;default:'junior'" json:"evaluator_level"`      // junior/senior/master
	EvaluationType     string     `gorm:"size:32;not null;default:'online';index" json:"evaluation_type"` // online/offline/ai/manual
	MarketPrice        float64    `gorm:"type:decimal(14,2);default:0" json:"market_price"`              // 市场价
	TradeInPrice       float64    `gorm:"type:decimal(14,2);default:0" json:"trade_in_price"`            // 收购价（置换）
	RetailPrice        float64    `gorm:"type:decimal(14,2);default:0" json:"retail_price"`              // 零售价
	WholesalePrice     float64    `gorm:"type:decimal(14,2);default:0" json:"wholesale_price"`           // 批发价
	DepreciationAmount float64    `gorm:"type:decimal(14,2);default:0" json:"depreciation_amount"`       // 折旧金额
	DepreciationRate   float64    `gorm:"type:decimal(5,2);default:0" json:"depreciation_rate"`          // 折旧率
	FinalPrice         float64    `gorm:"type:decimal(14,2);default:0;index" json:"final_price"`         // 最终估值
	PriceRangeMin      float64    `gorm:"type:decimal(14,2);default:0" json:"price_range_min"`           // 价格区间下限
	PriceRangeMax      float64    `gorm:"type:decimal(14,2);default:0" json:"price_range_max"`           // 价格区间上限
	Factors            JSONB      `gorm:"type:jsonb" json:"factors"`                                     // 评估因子 JSON
	SimilarDeals       JSONB      `gorm:"type:jsonb" json:"similar_deals"`                               // 相似成交 JSON
	Description        string     `gorm:"type:text" json:"description"`                                  // 评估说明
	ReportURL          string     `gorm:"size:255;not null;default:''" json:"report_url"`                // 评估报告 URL
	ValidUntil         *time.Time `gorm:"type:date;index" json:"valid_until"`                            // 估值有效期
	Status             int        `gorm:"default:0;index" json:"status"`                                 // 0待评估 1完成 2过期 3取消
}

// TableName 表名（car_ 前缀）
func (CarEvaluation) TableName() string { return "car_evaluations" }
