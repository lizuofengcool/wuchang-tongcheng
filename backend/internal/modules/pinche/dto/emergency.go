// Package dto 同城拼车出行数据传输对象 - 紧急联系人/一键报警
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// EmergencyInfo 紧急联系人/报警详情响应
type EmergencyInfo struct {
	ID              uint       `json:"id"`
	RegionID        uint       `json:"region_id"`
	UserID          uint       `json:"user_id"`
	ContactName     string     `json:"contact_name"`
	ContactPhone    string     `json:"contact_phone"`
	ContactRelation string     `json:"contact_relation"`
	IsPrimary       bool       `json:"is_primary"`
	PincheID        *uint      `json:"pinche_id"`
	TripID          *uint      `json:"trip_id"`
	AlertType       string     `json:"alert_type"`
	AlertTypeText   string     `json:"alert_type_text"`
	AlertStatus     int        `json:"alert_status"`
	AlertStatusText string     `json:"alert_status_text"`
	AlertTime       *time.Time `json:"alert_time"`
	AlertLocationLat float64   `json:"alert_location_lat"`
	AlertLocationLng float64   `json:"alert_location_lng"`
	AlertAddress    string     `json:"alert_address"`
	AlertDescription string    `json:"alert_description"`
	AlertEvidence   interface{} `json:"alert_evidence"`
	HandledAt       *time.Time `json:"handled_at"`
	HandlerID       *uint      `json:"handler_id"`
	HandleResult    string     `json:"handle_result"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateEmergencyContactRequest 创建紧急联系人请求
type CreateEmergencyContactRequest struct {
	ContactName     string `json:"contact_name" binding:"required,min=2,max=50"`
	ContactPhone    string `json:"contact_phone" binding:"required,len=11"`
	ContactRelation string `json:"contact_relation" binding:"max=32"`
	IsPrimary       bool   `json:"is_primary"`
}

// UpdateEmergencyContactRequest 更新紧急联系人请求
type UpdateEmergencyContactRequest struct {
	ContactName     *string `json:"contact_name"`
	ContactPhone    *string `json:"contact_phone"`
	ContactRelation *string `json:"contact_relation"`
	IsPrimary       *bool   `json:"is_primary"`
}

// SOSAlertRequest 一键报警请求
type SOSAlertRequest struct {
	PincheID        uint        `json:"pinche_id"`
	TripID          *uint       `json:"trip_id"`
	AlertType       string      `json:"alert_type" binding:"omitempty,oneof=sos share periodic"`
	AlertLocationLat float64    `json:"alert_location_lat"`
	AlertLocationLng float64    `json:"alert_location_lng"`
	AlertAddress    string      `json:"alert_address"`
	AlertDescription string     `json:"alert_description" binding:"max=500"`
	AlertEvidence   interface{} `json:"alert_evidence"`
}

// HandleAlertRequest 处理报警请求
type HandleAlertRequest struct {
	HandleResult string `json:"handle_result" binding:"max=500"`
}

// EmergencyListRequest 紧急联系人/报警列表查询请求
type EmergencyListRequest struct {
	UserID      uint   `form:"user_id" json:"user_id"`
	PincheID    uint   `form:"pinche_id" json:"pinche_id"`
	TripID      uint   `form:"trip_id" json:"trip_id"`
	AlertType   string `form:"alert_type" json:"alert_type"`
	AlertStatus *int   `form:"alert_status" json:"alert_status"`
	utils.Pagination
}

// ShareTripRequest 行程分享请求
type ShareTripRequest struct {
	PincheID uint `json:"pinche_id" binding:"required"`
	Hours    int  `json:"hours" binding:"omitempty,min=1,max=168"` // 1-168小时
}
