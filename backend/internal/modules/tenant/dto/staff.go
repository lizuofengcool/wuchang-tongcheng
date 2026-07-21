// Package dto 多租户分站数据传输对象 - 员工
package dto

import "wuchang-tongcheng/internal/pkg/utils"

// StaffInfo 员工详情响应
type StaffInfo struct {
	ID          uint   `json:"id"`
	StationID   uint   `json:"station_id"`
	UserID      uint   `json:"user_id"`
	Role        string `json:"role"`
	RoleText    string `json:"role_text"`
	Permissions any    `json:"permissions"`
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateStaffRequest 创建员工请求
type CreateStaffRequest struct {
	StationID   uint   `json:"station_id" binding:"required"`
	UserID      uint   `json:"user_id" binding:"required"`
	Role        string `json:"role" binding:"required,oneof=operator manager"`
	Permissions any    `json:"permissions"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateStaffRequest 更新员工请求
type UpdateStaffRequest struct {
	Role        *string `json:"role" binding:"omitempty,oneof=operator manager"`
	Permissions any     `json:"permissions"`
	Status      *int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// StaffListRequest 员工列表请求
type StaffListRequest struct {
	StationID uint   `form:"station_id" json:"station_id"`
	UserID    uint   `form:"user_id" json:"user_id"`
	Role      string `form:"role" json:"role"`
	Status    *int   `form:"status" json:"status"`
	utils.Pagination
}
