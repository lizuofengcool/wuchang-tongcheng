// Package dto 同城零工兼职数据传输对象 - 报名记录
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ApplicationInfo 报名详情响应
type ApplicationInfo struct {
	ID                uint       `json:"id"`
	ApplicationNo     string     `json:"application_no"`
	LinggongID        uint       `json:"linggong_id"`
	TaskID            uint       `json:"task_id"`
	EmployerID        uint       `json:"employer_id"`
	EmployerName      string     `json:"employer_name"`
	WorkerID          uint       `json:"worker_id"`
	WorkerName        string     `json:"worker_name"`
	WorkerAvatar      string     `json:"worker_avatar"`
	WorkerPhone       string     `json:"worker_phone"`
	WorkerAge         int        `json:"worker_age"`
	WorkerGender      string     `json:"worker_gender"`
	WorkerCity        string     `json:"worker_city"`
	WorkerCreditScore int        `json:"worker_credit_score"`
	WorkerProfileID   uint       `json:"worker_profile_id"`
	Source            string     `json:"source"`
	SourceText        string     `json:"source_text"`
	Method            string     `json:"method"`
	MethodText        string     `json:"method_text"`
	Status            int        `json:"status"`
	StatusText        string     `json:"status_text"`
	AppliedCount      int        `json:"applied_count"`
	CoverLetter       string     `json:"cover_letter"`
	EmployerRemark    string     `json:"employer_remark"`
	RejectReason      string     `json:"reject_reason"`
	CancelReason      string     `json:"cancel_reason"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ConfirmedAt       *time.Time `json:"confirmed_at"`
	OnboardedAt       *time.Time `json:"onboarded_at"`
	CompletedAt       *time.Time `json:"completed_at"`
	CanceledAt        *time.Time `json:"canceled_at"`
	Evaluated         bool       `json:"evaluated"`
	AttachmentURL     string     `json:"attachment_url"`
	RegionID          uint       `json:"region_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateApplicationRequest 创建报名请求
type CreateApplicationRequest struct {
	LinggongID   uint   `json:"linggong_id" binding:"required"`
	TaskID       uint   `json:"task_id"`
	Source       string `json:"source" binding:"omitempty,oneof=search recommend share direct invite favorite"`
	Method       string `json:"method" binding:"omitempty,oneof=online phone onsite"`
	CoverLetter  string `json:"cover_letter"`
	AttachmentURL string `json:"attachment_url" binding:"max=255"`
}

// UpdateApplicationRequest 更新报名请求
type UpdateApplicationRequest struct {
	EmployerRemark *string `json:"employer_remark"`
	Status         *int    `json:"status" binding:"omitempty,oneof=0 1 2 3 4 5 6 7 8 9 10"`
}

// ApplicationListRequest 报名列表请求
type ApplicationListRequest struct {
	LinggongID  uint   `form:"linggong_id" json:"linggong_id"`
	TaskID      uint   `form:"task_id" json:"task_id"`
	EmployerID  uint   `form:"employer_id" json:"employer_id"`
	WorkerID    uint   `form:"worker_id" json:"worker_id"`
	Status      *int   `form:"status" json:"status"`
	Source      string `form:"source" json:"source"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ApplicationAdminListRequest 管理后台报名列表请求
type ApplicationAdminListRequest struct {
	RegionID    uint   `form:"region_id" json:"region_id"`
	LinggongID  uint   `form:"linggong_id" json:"linggong_id"`
	EmployerID  uint   `form:"employer_id" json:"employer_id"`
	WorkerID    uint   `form:"worker_id" json:"worker_id"`
	Status      *int   `form:"status" json:"status"`
	Keyword     string `form:"keyword" json:"keyword"`
	utils.Pagination
}

// ApplicationAuditRequest 报名审核请求（雇主审核求职者）
type ApplicationAuditRequest struct {
	Status        int    `json:"status" binding:"oneof=1 2 5 6 7 8 9 10"`
	EmployerRemark string `json:"employer_remark"`
	RejectReason  string `json:"reject_reason" binding:"max=500"`
}

// ApplicationCancelRequest 取消报名请求
type ApplicationCancelRequest struct {
	CancelReason string `json:"cancel_reason" binding:"max=500"`
}

// AttendanceInfo 考勤信息
type AttendanceInfo struct {
	ID              uint       `json:"id"`
	AttendanceNo    string     `json:"attendance_no"`
	ApplicationID   uint       `json:"application_id"`
	LinggongID      uint       `json:"linggong_id"`
	WorkerID        uint       `json:"worker_id"`
	WorkerName      string     `json:"worker_name"`
	EmployerID      uint       `json:"employer_id"`
	AttendanceType  string     `json:"attendance_type"`
	AttendanceTypeText string  `json:"attendance_type_text"`
	ClockTime       *time.Time `json:"clock_time"`
	ClockMethod     string     `json:"clock_method"`
	ClockMethodText string     `json:"clock_method_text"`
	Latitude        float64    `json:"latitude"`
	Longitude       float64    `json:"longitude"`
	Address         string     `json:"address"`
	WifiSSID        string     `json:"wifi_ssid"`
	FaceConfidence  float64    `json:"face_confidence"`
	PhotoURL        string     `json:"photo_url"`
	Status          string     `json:"status"`
	StatusText      string     `json:"status_text"`
	WorkHours       float64    `json:"work_hours"`
	Remark          string     `json:"remark"`
	RegionID        uint       `json:"region_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ClockInRequest 打卡请求
type ClockInRequest struct {
	ApplicationID  uint    `json:"application_id" binding:"required"`
	AttendanceType string  `json:"attendance_type" binding:"omitempty,oneof=clock_in clock_out break resume overtime"`
	ClockMethod    string  `json:"clock_method" binding:"omitempty,oneof=gps wifi face manual qr_code"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Address        string `json:"address" binding:"max=500"`
	WifiSSID       string `json:"wifi_ssid" binding:"max=128"`
	FaceConfidence float64 `json:"face_confidence"`
	PhotoURL       string  `json:"photo_url" binding:"max=255"`
	Remark         string `json:"remark"`
}

// AttendanceListRequest 考勤列表请求
type AttendanceListRequest struct {
	ApplicationID uint   `form:"application_id" json:"application_id"`
	LinggongID    uint   `form:"linggong_id" json:"linggong_id"`
	WorkerID      uint   `form:"worker_id" json:"worker_id"`
	EmployerID    uint   `form:"employer_id" json:"employer_id"`
	AttendanceType string `form:"attendance_type" json:"attendance_type"`
	Status        string `form:"status" json:"status"`
	utils.Pagination
}
