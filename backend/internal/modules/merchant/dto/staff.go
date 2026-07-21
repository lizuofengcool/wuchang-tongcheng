// Package dto 商户中台数据传输对象 - 员工
// 对标美团/大众点评商家员工管理
package dto

import (
	"time"

	"wuchang-tongcheng/internal/pkg/utils"
)

// StaffInfo 员工详情响应
type StaffInfo struct {
	ID          uint      `json:"id"`
	ShopID      uint      `json:"shop_id"`
	UserID      uint      `json:"user_id"`
	Role        string    `json:"role"`
	RoleText    string    `json:"role_text"`
	Permissions interface{} `json:"permissions"`
	Status      int       `json:"status"`
	StatusText  string    `json:"status_text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateStaffRequest 添加员工请求
type CreateStaffRequest struct {
	ShopID      uint        `json:"shop_id" binding:"required"`
	UserID      uint        `json:"user_id" binding:"required"`
	Role        string      `json:"role" binding:"omitempty,oneof=owner manager clerk"`
	Permissions interface{} `json:"permissions"`
}

// UpdateStaffRequest 更新员工请求
type UpdateStaffRequest struct {
	Role        *string     `json:"role" binding:"omitempty,oneof=owner manager clerk"`
	Permissions interface{} `json:"permissions"`
	Status      *int        `json:"status" binding:"omitempty,oneof=1 2"`
}

// StaffListRequest 员工列表请求
type StaffListRequest struct {
	ShopID uint   `form:"shop_id" json:"shop_id"`
	UserID uint   `form:"user_id" json:"user_id"`
	Role   string `form:"role" json:"role"`
	Status *int   `form:"status" json:"status"`
	utils.Pagination
}

// StaffPermissionUpdateRequest 权限分配请求
type StaffPermissionUpdateRequest struct {
	Permissions interface{} `json:"permissions"`
}

// StaffRoleSwitchRequest 角色切换请求
type StaffRoleSwitchRequest struct {
	Role string `json:"role" binding:"oneof=owner manager clerk"`
}
