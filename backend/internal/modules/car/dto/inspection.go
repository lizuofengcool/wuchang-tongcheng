// Package dto 同城车辆买卖数据传输对象 - 车况检测
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// InspectionInfo 检测详情响应
type InspectionInfo struct {
	ID              uint       `json:"id"`
	InspectionNo    string     `json:"inspection_no"`
	CarID           uint       `json:"car_id"`
	ListingID       uint       `json:"listing_id"`
	InspectorID     uint       `json:"inspector_id"`
	InspectorName   string     `json:"inspector_name"`
	InspectorLevel  string     `json:"inspector_level"`
	InspectionType  string     `json:"inspection_type"`
	TotalItems      int        `json:"total_items"`
	PassedItems     int        `json:"passed_items"`
	FailedItems     int        `json:"failed_items"`
	WarningItems    int        `json:"warning_items"`
	OverallScore    float64    `json:"overall_score"`
	ConditionLevel  string     `json:"condition_level"`
	ExteriorScore   float64    `json:"exterior_score"`
	InteriorScore   float64    `json:"interior_score"`
	EngineScore     float64    `json:"engine_score"`
	ChassisScore    float64    `json:"chassis_score"`
	ElectronicsScore float64   `json:"electronics_score"`
	SafetyScore     float64    `json:"safety_score"`
	Items           interface{} `json:"items"`
	AccidentHistory interface{} `json:"accident_history"`
	HasAccident     bool       `json:"has_accident"`
	HasFlood        bool       `json:"has_flood"`
	HasFire         bool       `json:"has_fire"`
	HasOverhaul     bool       `json:"has_overhaul"`
	ReportURL       string     `json:"report_url"`
	ReportImages    interface{} `json:"report_images"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	ReviewedBy      uint       `json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	Status          int        `json:"status"`
	StatusText      string     `json:"status_text"`
	Remark          string     `json:"remark"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateInspectionRequest 创建检测请求
type CreateInspectionRequest struct {
	CarID          uint   `json:"car_id" binding:"required"`
	ListingID      uint   `json:"listing_id"`
	InspectorID    uint   `json:"inspector_id"`
	InspectorName  string `json:"inspector_name" binding:"max=50"`
	InspectorLevel string `json:"inspector_level" binding:"omitempty,oneof=junior senior master expert"`
	InspectionType string `json:"inspection_type" binding:"omitempty,oneof=standard simple deep pre_sale post_sale"`
	TotalItems     int    `json:"total_items"`
	PassedItems    int    `json:"passed_items"`
	FailedItems    int    `json:"failed_items"`
	WarningItems   int    `json:"warning_items"`
	OverallScore   float64 `json:"overall_score"`
	ConditionLevel string `json:"condition_level" binding:"omitempty,oneof=A B C D"`
	ExteriorScore  float64 `json:"exterior_score"`
	InteriorScore  float64 `json:"interior_score"`
	EngineScore    float64 `json:"engine_score"`
	ChassisScore   float64 `json:"chassis_score"`
	ElectronicsScore float64 `json:"electronics_score"`
	SafetyScore    float64 `json:"safety_score"`
	Items          interface{} `json:"items"`
	AccidentHistory interface{} `json:"accident_history"`
	HasAccident    bool   `json:"has_accident"`
	HasFlood       bool   `json:"has_flood"`
	HasFire        bool   `json:"has_fire"`
	HasOverhaul    bool   `json:"has_overhaul"`
	ReportURL      string `json:"report_url" binding:"max=255"`
	ReportImages   interface{} `json:"report_images"`
	Remark         string `json:"remark"`
}

// UpdateInspectionRequest 更新检测请求
type UpdateInspectionRequest struct {
	InspectorID    *uint   `json:"inspector_id"`
	InspectorName  *string `json:"inspector_name" binding:"omitempty,max=50"`
	InspectorLevel *string `json:"inspector_level" binding:"omitempty,oneof=junior senior master expert"`
	InspectionType *string `json:"inspection_type" binding:"omitempty,oneof=standard simple deep pre_sale post_sale"`
	TotalItems     *int    `json:"total_items"`
	PassedItems    *int    `json:"passed_items"`
	FailedItems    *int    `json:"failed_items"`
	WarningItems   *int    `json:"warning_items"`
	OverallScore   *float64 `json:"overall_score"`
	ConditionLevel *string `json:"condition_level" binding:"omitempty,oneof=A B C D"`
	ExteriorScore  *float64 `json:"exterior_score"`
	InteriorScore  *float64 `json:"interior_score"`
	EngineScore    *float64 `json:"engine_score"`
	ChassisScore   *float64 `json:"chassis_score"`
	ElectronicsScore *float64 `json:"electronics_score"`
	SafetyScore    *float64 `json:"safety_score"`
	Items          interface{} `json:"items"`
	AccidentHistory interface{} `json:"accident_history"`
	HasAccident    *bool   `json:"has_accident"`
	HasFlood       *bool   `json:"has_flood"`
	HasFire        *bool   `json:"has_fire"`
	HasOverhaul    *bool   `json:"has_overhaul"`
	ReportURL      *string `json:"report_url" binding:"omitempty,max=255"`
	ReportImages   interface{} `json:"report_images"`
	Remark         *string `json:"remark"`
	Status         *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// InspectionListRequest 检测列表请求
type InspectionListRequest struct {
	CarID           uint   `form:"car_id" json:"car_id"`
	ListingID       uint   `form:"listing_id" json:"listing_id"`
	InspectorID     uint   `form:"inspector_id" json:"inspector_id"`
	InspectionType  string `form:"inspection_type" json:"inspection_type"`
	ConditionLevel string  `form:"condition_level" json:"condition_level"`
	HasAccident    *bool  `form:"has_accident" json:"has_accident"`
	Status         *int   `form:"status" json:"status"`
	Keyword        string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// InspectionReviewRequest 检测复核请求
type InspectionReviewRequest struct {
	ReviewedBy uint   `json:"reviewed_by"`
	Remark     string `json:"remark"`
	Status     int    `json:"status" binding:"oneof=3 4"` // 3 复核通过，4 取消
}
