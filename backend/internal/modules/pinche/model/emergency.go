// Package model 拼车紧急联系人/一键报警数据模型
package model

import (
	"time"

	"wuchang-tongcheng/internal/pkg/database"
)

// PincheEmergency 紧急联系人/一键报警
type PincheEmergency struct {
	database.RegionBaseModel

	// 紧急联系人
	UserID         uint   `gorm:"index;not null" json:"user_id"`
	ContactName    string `gorm:"size:50" json:"contact_name"`
	ContactPhone   string `gorm:"size:20" json:"contact_phone"`
	ContactRelation string `gorm:"size:32" json:"contact_relation"`
	IsPrimary      bool   `gorm:"default:false" json:"is_primary"`

	// 关联行程
	PincheID *uint `gorm:"index" json:"pinche_id"`
	TripID   *uint `gorm:"index" json:"trip_id"`

	// 报警信息
	AlertType   string `gorm:"size:16;not null;default:'sos';index" json:"alert_type"` // sos/share/periodic
	AlertStatus int    `gorm:"default:0;index" json:"alert_status"`                     // 0未处理 1处理中 2已处理
	AlertTime   *time.Time `gorm:"index" json:"alert_time"`
	AlertLocationLat float64 `gorm:"type:decimal(10,7);default:0" json:"alert_location_lat"`
	AlertLocationLng float64 `gorm:"type:decimal(10,7);default:0" json:"alert_location_lng"`
	AlertAddress    string `gorm:"size:255" json:"alert_address"`
	AlertDescription string `gorm:"size:500" json:"alert_description"`
	AlertEvidence   JSONB  `gorm:"type:jsonb" json:"alert_evidence"`

	HandledAt    *time.Time `gorm:"index" json:"handled_at"`
	HandlerID    *uint      `gorm:"index" json:"handler_id"`
	HandleResult string     `gorm:"size:500" json:"handle_result"`
}

// TableName 表名
func (PincheEmergency) TableName() string { return "pinche_emergencies" }
