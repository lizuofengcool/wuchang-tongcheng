// Package model 权限数据模型
// 实现标准RBAC: 用户-角色-权限
package model

import "wuchang-tongcheng/internal/pkg/database"

// Role 角色
type Role struct {
	database.BaseModel
	Name        string `gorm:"size:50;uniqueIndex;not null" json:"name"` // 角色名称
	Code        string `gorm:"size:50;uniqueIndex;not null" json:"code"` // 角色编码（如 admin/editor/user）
	Description string `gorm:"size:255" json:"description"`              // 描述
	Sort        int    `gorm:"default:0" json:"sort"`                    // 排序
	Status      int    `gorm:"default:1" json:"status"`                  // 状态 1正常 0禁用
}

// TableName 表名
func (Role) TableName() string { return "roles" }

// Permission 权限
type Permission struct {
	database.BaseModel
	Name       string `gorm:"size:50;not null" json:"name"`            // 权限名称
	Code       string `gorm:"size:100;uniqueIndex;not null" json:"code"` // 权限编码（如 user:create user:read）
	Type       int    `gorm:"default:1" json:"type"`                    // 类型 1菜单 2按钮 3接口
	ParentID   uint   `gorm:"index;default:0" json:"parent_id"`         // 父级ID
	Path       string `gorm:"size:255" json:"path"`                    // 前端路由/接口路径
	Method     string `gorm:"size:20" json:"method"`                   // HTTP方法（接口类型时用）
	Sort       int    `gorm:"default:0" json:"sort"`                   // 排序
	Status     int    `gorm:"default:1" json:"status"`                 // 状态 1正常 0禁用
}

// TableName 表名
func (Permission) TableName() string { return "permissions" }

// UserRole 用户-角色关联
type UserRole struct {
	database.BaseModel
	UserID uint `gorm:"uniqueIndex:idx_user_role;not null" json:"user_id"`
	RoleID uint `gorm:"uniqueIndex:idx_user_role;not null" json:"role_id"`
}

// TableName 表名
func (UserRole) TableName() string { return "user_roles" }

// RolePermission 角色-权限关联
type RolePermission struct {
	database.BaseModel
	RoleID       uint `gorm:"uniqueIndex:idx_role_perm;not null" json:"role_id"`
	PermissionID uint `gorm:"uniqueIndex:idx_role_perm;not null" json:"permission_id"`
}

// TableName 表名
func (RolePermission) TableName() string { return "role_permissions" }
