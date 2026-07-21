// Package model 同城商城 - 举报表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 举报人）
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// Report 举报表
type Report struct {
	database.RegionBaseModel // 含 id/region_id/created_at/updated_at/deleted_at
	ReportNo    string `gorm:"size:32;not null;default:'';index" json:"report_no"` // 举报单号
	ReporterID  uint   `gorm:"not null;default:0;index" json:"reporter_id"`         // 举报人 ID
	ReporterName string `gorm:"size:50;not null;default:''" json:"reporter_name"`  // 举报人昵称

	// === 被举报对象 ===
	TargetType string `gorm:"size:32;not null;default:'';index" json:"target_type"` // shop/product/review/order/user
	TargetID   uint   `gorm:"not null;default:0;index" json:"target_id"`            // 被举报对象 ID
	TargetName string `gorm:"size:200;not null;default:''" json:"target_name"`      // 被举报对象名称

	// === 举报内容 ===
	ReportType     string `gorm:"size:32;not null;default:'';index" json:"report_type"` // 举报类型
	ReportReason   string `gorm:"size:500;not null;default:''" json:"report_reason"`    // 举报原因
	Description    string `gorm:"type:text;not null;default:''" json:"description"`    // 详细描述
	EvidenceImages JSONB  `gorm:"type:jsonb" json:"evidence_images"`                   // 证据图片 JSON
	ContactInfo    string `gorm:"size:100;not null;default:''" json:"contact_info"`    // 联系方式

	// === 处理 ===
	Status       int        `gorm:"not null;default:0;index" json:"status"`        // 0待处理 1已核实警告 2已下架 3已封号 4已驳回 5已转交
	HandlerID    uint       `gorm:"index" json:"handler_id"`                       // 处理人 ID
	HandlerName  string     `gorm:"size:50;not null;default:''" json:"handler_name"` // 处理人姓名
	HandleResult string     `gorm:"size:500;not null;default:''" json:"handle_result"` // 处理结果
	HandledAt    *time.Time `gorm:"index" json:"handled_at"`                        // 处理时间

	// === 处罚 ===
	PenaltyType     string `gorm:"size:32;not null;default:''" json:"penalty_type"` // 处罚类型 warning/limit/ban1d/ban7d/banForever
	PenaltyTargetID uint   `gorm:"index" json:"penalty_target_id"`                  // 处罚对象 ID
}

// TableName 表名
func (Report) TableName() string { return "mall_reports" }
