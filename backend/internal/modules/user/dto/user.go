// Package dto 用户模块数据传输对象
package dto

import "time"

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	Expires  int    `json:"expires"` // 过期时间（秒）
	UserInfo UserInfo `json:"user_info"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Gender    int       `json:"gender"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateProfileRequest 更新个人资料请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"max=100"`
	Gender   int    `json:"gender" binding:"oneof=0 1 2"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// ===== 管理后台相关 DTO =====

// ListUsersRequest 用户列表查询请求
type ListUsersRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Keyword  string `form:"keyword" json:"keyword"` // 用户名/昵称模糊搜索
	Status   int    `form:"status" json:"status"`   // -1全部 0禁用 1正常
}

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"max=100"`
	Gender   int    `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Avatar   string `json:"avatar" binding:"max=255"`
	Phone    string `json:"phone" binding:"max=20"`
	Email    string `json:"email" binding:"max=100"`
	Gender   int    `json:"gender" binding:"omitempty,oneof=0 1 2"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status int `json:"status" binding:"oneof=0 1"`
}

// ResetPasswordRequest 管理员重置用户密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}
