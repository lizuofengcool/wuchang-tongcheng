// Package permission 权限模块插件
// 实现RBAC权限管理：用户-角色-权限
package permission

import (
	"context"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	coreRouter "wuchang-tongcheng/internal/core/router"
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
	// 注入角色编码查询器，用于超级管理员（admin 角色）直通权限校验
	middleware.SetRoleCodeFetcher(svc.GetRoleCodesByUserID)

	return nil
}

// RegisterRoutes 注册插件路由
func (p *Plugin) RegisterRoutes(router plugin.RouterGroup) {
	auth := coreRouter.WrapGin(middleware.AuthRequired())

	// 角色管理
	router.POST("/roles", coreRouter.WrapGin(middleware.RequirePermission("role:create")), p.handler.CreateRole)
	router.PUT("/roles/:id", coreRouter.WrapGin(middleware.RequirePermission("role:update")), p.handler.UpdateRole)
	router.DELETE("/roles/:id", coreRouter.WrapGin(middleware.RequirePermission("role:delete")), p.handler.DeleteRole)
	router.GET("/roles/:id", coreRouter.WrapGin(middleware.RequirePermission("role:read")), p.handler.GetRoleByID)
	router.GET("/roles", coreRouter.WrapGin(middleware.RequirePermission("role:read")), p.handler.ListRoles)
	router.GET("/users/:id/roles", coreRouter.WrapGin(middleware.RequirePermission("role:read")), p.handler.UserRoles)
	router.GET("/roles/:id/permissions", coreRouter.WrapGin(middleware.RequirePermission("role:read")), p.handler.RolePermissions)

	// 权限管理
	router.POST("/permissions", coreRouter.WrapGin(middleware.RequirePermission("permission:create")), p.handler.CreatePermission)
	router.PUT("/permissions/:id", coreRouter.WrapGin(middleware.RequirePermission("permission:update")), p.handler.UpdatePermission)
	router.DELETE("/permissions/:id", coreRouter.WrapGin(middleware.RequirePermission("permission:delete")), p.handler.DeletePermission)
	router.GET("/permissions", coreRouter.WrapGin(middleware.RequirePermission("permission:read")), p.handler.ListPermissions)
	router.GET("/permissions/:id", coreRouter.WrapGin(middleware.RequirePermission("permission:read")), p.handler.GetPermissionByID)

	// 分配
	router.POST("/assign-roles", coreRouter.WrapGin(middleware.RequirePermission("permission:assign")), p.handler.AssignRoles)            // 给用户分配角色
	router.POST("/assign-permissions", coreRouter.WrapGin(middleware.RequirePermission("permission:assign")), p.handler.AssignPermissions) // 给角色分配权限

	// 当前用户权限（仅需登录）
	router.GET("/my-permissions", auth, p.handler.MyPermissions)
	router.GET("/my-auth", auth, p.handler.MyAuth)
}

// Close 关闭插件
func (p *Plugin) Close() error { return nil }

// 确保Plugin实现了plugin.Plugin接口
var _ plugin.Plugin = (*Plugin)(nil)
