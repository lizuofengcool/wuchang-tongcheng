// Package model 车况检测表（对标瓜子）
// 254项检测/检测师/三级评分/事故记录/复核
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// === 检测状态常量 ===
const (
	InspectionStatusPending    = 0 // 待检测
	InspectionStatusInProgress = 1 // 检测中
	InspectionStatusCompleted  = 2 // 已完成
	InspectionStatusReviewed   = 3 // 已复核
	InspectionStatusCanceled   = 4 // 已取消
)

// === 检测类型常量 ===
const (
	InspectionTypeStandard  = "standard"  // 标准检测（254项）
	InspectionTypeSimple    = "simple"    // 简检
	InspectionTypeDeep      = "deep"      // 深度检测
	InspectionTypePreSale   = "pre_sale"  // 售前检测
	InspectionTypePostSale  = "post_sale" // 售后检测
)

// === 检测师等级常量 ===
const (
	InspectorLevelJunior  = "junior"  // 初级
	InspectorLevelSenior  = "senior"  // 高级
	InspectorLevelMaster  = "master"  // 资深
	InspectorLevelExpert  = "expert"  // 专家
)

// === 车况等级常量（与主表 condition_level 同步） ===
const (
	InspectionConditionA = "A" // 极佳
	InspectionConditionB = "B" // 良好
	InspectionConditionC = "C" // 一般
	InspectionConditionD = "D" // 较差
)

// CarInspection 车况检测表
type CarInspection struct {
	database.RegionBaseModel // id/region_id/created_at/updated_at/deleted_at
	InspectionNo      string     `gorm:"size:64;not null;uniqueIndex" json:"inspection_no"`              // 检测单号
	CarID             uint       `gorm:"not null;index" json:"car_id"`                                    // 车源 ID
	ListingID         uint       `gorm:"not null;default:0;index" json:"listing_id"`                      // 发布 ID
	InspectorID       uint       `gorm:"not null;index" json:"inspector_id"`                              // 检测师 ID
	InspectorName     string     `gorm:"size:50;not null;default:''" json:"inspector_name"`               // 检测师姓名
	InspectorLevel    string     `gorm:"size:16;not null;default:'junior'" json:"inspector_level"`        // junior/senior/master/expert
	InspectionType    string     `gorm:"size:32;not null;default:'standard';index" json:"inspection_type"` // standard/simple/deep/pre_sale/post_sale
	TotalItems        int        `gorm:"not null;default:254" json:"total_items"`                         // 检测项总数
	PassedItems       int        `gorm:"not null;default:0" json:"passed_items"`                          // 通过项数
	FailedItems       int        `gorm:"not null;default:0" json:"failed_items"`                          // 不通过项数
	WarningItems      int        `gorm:"not null;default:0" json:"warning_items"`                         // 警告项数
	OverallScore      float64    `gorm:"type:decimal(5,2);default:0" json:"overall_score"`                // 综合评分
	ConditionLevel    string     `gorm:"size:16;not null;default:'A';index" json:"condition_level"`       // A/B/C/D
	ExteriorScore     float64    `gorm:"type:decimal(5,2);default:0" json:"exterior_score"`               // 外观评分
	InteriorScore     float64    `gorm:"type:decimal(5,2);default:0" json:"interior_score"`               // 内饰评分
	EngineScore       float64    `gorm:"type:decimal(5,2);default:0" json:"engine_score"`                 // 发动机评分
	ChassisScore      float64    `gorm:"type:decimal(5,2);default:0" json:"chassis_score"`                // 底盘评分
	ElectronicsScore  float64    `gorm:"type:decimal(5,2);default:0" json:"electronics_score"`            // 电子设备评分
	SafetyScore       float64    `gorm:"type:decimal(5,2);default:0" json:"safety_score"`                 // 安全评分
	Items             JSONB      `gorm:"type:jsonb" json:"items"`                                         // 254项检测明细 JSON
	AccidentHistory   JSONB      `gorm:"type:jsonb" json:"accident_history"`                              // 事故历史 JSON
	HasAccident       bool       `gorm:"not null;default:false;index" json:"has_accident"`                // 是否有事故
	HasFlood          bool       `gorm:"not null;default:false" json:"has_flood"`                         // 是否泡水
	HasFire           bool       `gorm:"not null;default:false" json:"has_fire"`                          // 是否火烧
	HasOverhaul        bool       `gorm:"not null;default:false" json:"has_overhaul"`                       // 是否大修
	ReportURL         string     `gorm:"size:255;not null;default:''" json:"report_url"`                  // 检测报告 PDF URL
	ReportImages      JSONB      `gorm:"type:jsonb" json:"report_images"`                                 // 报告图片
	StartedAt         *time.Time `gorm:"index" json:"started_at"`                                         // 开始时间
	CompletedAt       *time.Time `gorm:"index" json:"completed_at"`                                       // 完成时间
	ReviewedBy        uint       `gorm:"not null;default:0" json:"reviewed_by"`                           // 复核人 ID
	ReviewedAt        *time.Time `gorm:"index" json:"reviewed_at"`                                        // 复核时间
	Status            int        `gorm:"default:0;index" json:"status"`                                   // 0待检 1检测中 2完成 3复核 4取消
	Remark            string     `gorm:"type:text" json:"remark"`                                         // 备注
}

// TableName 表名（car_ 前缀）
func (CarInspection) TableName() string { return "car_inspections" }
