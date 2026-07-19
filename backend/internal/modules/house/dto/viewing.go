// Package dto 看房预约 DTO
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
// 7 状态机：待确认/已确认/已完成/已取消/已过期/已改期/已关闭
// 注：model 定义了 6 个状态常量（0待确认 1已确认 2进行中 3完成 4取消 5爽约）
// 服务层在此基础上结合改期/过期/关闭语义实现 7 状态机
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// ViewingResponse 看房预约响应
type ViewingResponse struct {
	ID             uint       `json:"id"`
	ViewingNo      string     `json:"viewing_no"`
	HouseID        uint       `json:"house_id"`
	ListingID      uint       `json:"listing_id"`
	CommunityID    uint       `json:"community_id"`
	UserID         uint       `json:"user_id"`
	UserName       string     `json:"user_name"`
	UserPhone      string     `json:"user_phone"`
	UserAvatar     string     `json:"user_avatar"`
	AgentID        uint       `json:"agent_id"`
	AgentName      string     `json:"agent_name"`
	AgentPhone     string     `json:"agent_phone"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes"`
	ViewingType    string     `json:"viewing_type"`
	ViewingTypeText string    `json:"viewing_type_text"`
	OnlineURL      string     `json:"online_url"`
	OnlinePassword string     `json:"online_password"`
	MeetLocation   string     `json:"meet_location"`
	Remark         string     `json:"remark"`
	Status         int        `json:"status"`
	StatusText     string     `json:"status_text"`
	Result         string     `json:"result"`
	ResultText     string     `json:"result_text"`
	Feedback       string     `json:"feedback"`
	Rating         int        `json:"rating"`
	AttendedAt     *time.Time `json:"attended_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CanceledAt     *time.Time `json:"canceled_at"`
	CanceledReason string     `json:"canceled_reason"`
	CanceledBy     uint       `json:"canceled_by"`
	ReminderSent   bool       `json:"reminder_sent"`
	ReminderSentAt *time.Time `json:"reminder_sent_at"`
	RegionID       uint       `json:"region_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ViewingCreateRequest 创建看房预约请求
type ViewingCreateRequest struct {
	HouseID        uint   `json:"house_id" binding:"required"`
	ListingID      uint   `json:"listing_id"`
	AgentID        uint   `json:"agent_id"`
	ScheduledAt    *time.Time `json:"scheduled_at" binding:"required"`
	DurationMinutes int    `json:"duration_minutes" binding:"gte=10,lte=480"`
	ViewingType    string `json:"viewing_type" binding:"omitempty,oneof=offline online vr"`
	OnlineURL      string `json:"online_url" binding:"max=500"`
	OnlinePassword string `json:"online_password" binding:"max=64"`
	MeetLocation   string `json:"meet_location" binding:"max=255"`
	Remark         string `json:"remark" binding:"max=500"`
}

// ViewingUpdateRequest 更新看房预约请求（仅预约人和经纪人在确认前可改）
type ViewingUpdateRequest struct {
	ScheduledAt    *time.Time `json:"scheduled_at"`
	DurationMinutes int    `json:"duration_minutes" binding:"gte=10,lte=480"`
	ViewingType    string `json:"viewing_type" binding:"omitempty,oneof=offline online vr"`
	OnlineURL      string `json:"online_url"`
	OnlinePassword string `json:"online_password"`
	MeetLocation   string `json:"meet_location"`
	Remark         string `json:"remark"`
}

// ViewingConfirmRequest 确认看房预约请求
type ViewingConfirmRequest struct {
	MeetLocation string `json:"meet_location" binding:"max=255"`
	Remark       string `json:"remark" binding:"max=500"`
}

// ViewingCancelRequest 取消看房预约请求
type ViewingCancelRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// ViewingRescheduleRequest 改期请求
type ViewingRescheduleRequest struct {
	ScheduledAt    *time.Time `json:"scheduled_at" binding:"required"`
	DurationMinutes int    `json:"duration_minutes" binding:"gte=10,lte=480"`
	Remark         string `json:"remark" binding:"max=500"`
}

// ViewingCompleteRequest 完成看房请求
type ViewingCompleteRequest struct {
	Result   string `json:"result" binding:"omitempty,oneof=pending satisfied neutral dissatisfied no_show"`
	Feedback string `json:"feedback"`
	Rating   int    `json:"rating" binding:"gte=0,lte=5"`
}

// ViewingListQuery 看房预约列表查询
type ViewingListQuery struct {
	HouseID     uint   `form:"house_id" json:"house_id"`
	AgentID     uint   `form:"agent_id" json:"agent_id"`
	UserID      uint   `form:"user_id" json:"user_id"`
	ListingID   uint   `form:"listing_id" json:"listing_id"`
	ViewingType string `form:"viewing_type" json:"viewing_type"`
	Status      *int   `form:"status" json:"status"`
	Result      string `form:"result" json:"result"`
	StartDate   string `form:"start_date" json:"start_date"`
	EndDate     string `form:"end_date" json:"end_date"`
	utils.Pagination
}

// ViewingAdminListQuery 管理后台看房预约列表查询
type ViewingAdminListQuery struct {
	RegionID uint   `form:"region_id" json:"region_id"`
	HouseID  uint   `form:"house_id" json:"house_id"`
	UserID   uint   `form:"user_id" json:"user_id"`
	AgentID  uint   `form:"agent_id" json:"agent_id"`
	Status   *int   `form:"status" json:"status"`
	utils.Pagination
}
