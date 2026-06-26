// Package permission 权限模块插件
// 实现RBAC权限管理：用户-角色-权限
package permission

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/modules/permission/handler"
	"wuchang-tongcheng/internal/modules/permission/model"
	"wuchang-tongcheng/internal/modules/permission/repository"
	"wuchang-tongcheng/internal/modules/permission/service"
	"wuchang-tongcheng/internal/pkg/database"
)

// Plugin 权限模块插件
type Plugin struct {
	name    string
	version string
	handler *handler.Handler
}

// NewPlugin 创建权限模块插件
func NewPlugin() *Plugin {
	return &Plugin{name: "permission", version: "1.0.0"}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return p.name }

// Version 返回插件版本号
func (p *Plugin) Version() string { return p.version }

// Init 初始化插件
func (p *Plugin) Init(ctx context.Context) error {
	db := database.GetDB()

	// 自动迁移权限相关表
	if err := db.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
	); err != nil {
		return err
	}

	// 初始化依赖链
	repo := repository.NewPermissionRepository(db)
	svc := service.NewPermissionService(repo)
	p.handler = handler.NewHandler(svc)

	// 注入权限校验器到中间件层（解耦中间件与业务模块的循环依赖）
	// 这样其他模块的路由可以用 middleware.RequirePermission("xxx") 做权限控制
	middleware.SetPermissionChecker(svc.HasPermission)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	// 角色管理
	router.POST("/roles", p.handler.CreateRole)
	router.PUT("/roles/:id", p.handler.UpdateRole)
	router.DELETE("/roles/:id", p.handler.DeleteRole)
	router.GET("/roles/:id", p.handler.GetRoleByID)
	router.GET("/roles", p.handler.ListRoles)
	router.GET("/users/:id/roles", p.handler.UserRoles)

	// 权限管理
	router.POST("/permissions", p.handler.CreatePermission)
	router.DELETE("/permissions/:id", p.handler.DeletePermission)
	router.GET("/permissions", p.handler.ListPermissions)

	// 分配
	router.POST("/assign-roles", p.handler.AssignRoles)          // 给用户分配角色
	router.POST("/assign-permissions", p.handler.AssignPermissions) // 给角色分配权限

	// 当前用户权限
	router.GET("/my-permissions", p.handler.MyPermissions)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
