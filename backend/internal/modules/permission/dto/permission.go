// Package dto 权限模块数据传输对象
package dto

// RoleInfo 角色信息
type RoleInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

// PermissionInfo 权限信息
type PermissionInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Type     int    `json:"type"`
	ParentID uint   `json:"parent_id"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Code        string `json:"code" binding:"required,max=50"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"max=50"`
	Description string `json:"description" binding:"max=255"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name     string `json:"name" binding:"required,max=50"`
	Code     string `json:"code" binding:"required,max=100"`
	Type     int    `json:"type" binding:"oneof=1 2 3"`
	ParentID uint   `json:"parent_id"`
	Path     string `json:"path" binding:"max=255"`
	Method   string `json:"method" binding:"max=20"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

// AssignRolesRequest 给用户分配角色
type AssignRolesRequest struct {
	UserID  uint   `json:"user_id" binding:"required"`
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

// AssignPermissionsRequest 给角色分配权限
type AssignPermissionsRequest struct {
	RoleID        uint   `json:"role_id" binding:"required"`
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
