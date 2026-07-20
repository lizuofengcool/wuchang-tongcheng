// Package model 拼车投诉数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheComplaint 投诉
type PincheComplaint struct {
	database.RegionBaseModel

	PincheID  uint  `gorm:"index;not null" json:"pinche_id"`
	BookingID *uint `gorm:"index" json:"booking_id"`
	TripID    *uint `gorm:"index" json:"trip_id"`

	ComplainantID   uint   `gorm:"index;not null" json:"complainant_id"`
	ComplainantName string `gorm:"size:50" json:"complainant_name"`
	RespondentID    uint   `gorm:"index;not null" json:"respondent_id"`
	RespondentName  string `gorm:"size:50" json:"respondent_name"`

	ComplaintType   string `gorm:"size:32" json:"complaint_type"`
	ComplaintReason string `gorm:"size:500" json:"complaint_reason"`
	Description     string `gorm:"type:text" json:"description"`
	EvidenceImages  JSONB  `gorm:"type:jsonb" json:"evidence_images"`

	Status       int        `gorm:"default:0;index" json:"status"` // 0待处理 1处理中 2已处理 3已驳回
	HandlerID    *uint      `gorm:"index" json:"handler_id"`
	HandlerName  string     `gorm:"size:50" json:"handler_name"`
	HandleResult string     `gorm:"size:500" json:"handle_result"`
	HandledAt    *time.Time `gorm:"index" json:"handled_at"`

	PenaltyType   string `gorm:"size:32" json:"penalty_type"`
	PenaltyUserID *uint  `gorm:"index" json:"penalty_user_id"`
	SLADeadline   *time.Time `gorm:"index" json:"sla_deadline"`

	AppealedAt       *time.Time `gorm:"index" json:"appealed_at"`
	AppealResult     string     `gorm:"size:500" json:"appeal_result"`
	AppealHandlerID  *uint      `gorm:"index" json:"appeal_handler_id"`
}

// TableName 表名
func (PincheComplaint) TableName() string { return "pinche_complaints" }
