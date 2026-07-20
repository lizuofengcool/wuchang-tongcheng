// Package dto 同城114数据传输对象 - 电话拨打记录
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// PhoneCallInfo 电话拨打记录响应
type PhoneCallInfo struct {
	ID           uint       `json:"id"`
	CallNo       string     `json:"call_no"`
	Dh114ID      uint       `json:"dh114_id"`
	BusinessID   uint       `json:"business_id"`
	Phone        string     `json:"phone"`
	CallerID     uint       `json:"caller_id"`
	CallerPhone  string     `json:"caller_phone"`
	CallerName   string     `json:"caller_name"`
	CallType     string     `json:"call_type"`
	Device       string     `json:"device"`
	IP           string     `json:"ip"`
	Status       string     `json:"status"`
	Duration     int        `json:"duration"`
	CalledAt     time.Time  `json:"called_at"`
	RegionID     uint       `json:"region_id"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PhoneCallRequest 电话拨打请求
type PhoneCallRequest struct {
	Dh114ID  uint   `json:"dh114_id" binding:"required"`
	CallType string `json:"call_type" binding:"omitempty,oneof=click call"`
	Device   string `json:"device"`
}

// PhoneCallListRequest 电话拨打列表请求
type PhoneCallListRequest struct {
	Dh114ID   uint   `form:"dh114_id" json:"dh114_id"`
	CallerID  uint   `form:"caller_id" json:"caller_id"`
	CallType  string `form:"call_type" json:"call_type"`
	Status    string `form:"status" json:"status"`
	utils.Pagination
}

// PhoneCallAdminListRequest 管理后台电话拨打列表请求
type PhoneCallAdminListRequest struct {
	Dh114ID    uint   `form:"dh114_id" json:"dh114_id"`
	CallerID   uint   `form:"caller_id" json:"caller_id"`
	CallType   string `form:"call_type" json:"call_type"`
	Status     string `form:"status" json:"status"`
	utils.Pagination
}
