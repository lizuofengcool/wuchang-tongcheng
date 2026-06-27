package middleware

import (
	"net/http"

	"wuchang-tongcheng/internal/core/response"

	"github.com/gin-gonic/gin"
)

// PermissionChecker 权限校验函数类型
// 返回 true 表示用户拥有该权限
type PermissionChecker func(userID uint, permCode string) (bool, error)

// RoleCodeFetcher 用户角色编码查询函数类型
type RoleCodeFetcher func(userID uint) ([]string, error)

// 全局权限校验器（由 permission 插件在 Init 时注入）
var permissionChecker PermissionChecker

// 全局角色编码查询器（由 permission 插件在 Init 时注入，用于超级管理员直通）
var roleCodeFetcher RoleCodeFetcher

// SuperAdminRoleCode 超级管理员角色编码，拥有该角色的用户直通所有权限校验
const SuperAdminRoleCode = "admin"

// SetPermissionChecker 注入权限校验器
// 在 permission 插件 Init 时调用：middleware.SetPermissionChecker(svc.HasPermission)
func SetPermissionChecker(checker PermissionChecker) {
	permissionChecker = checker
}

// SetRoleCodeFetcher 注入角色编码查询器，用于超级管理员直通
func SetRoleCodeFetcher(fetcher RoleCodeFetcher) {
	roleCodeFetcher = fetcher
}

// IsSuperAdmin 判断用户是否为超级管理员（拥有 admin 角色编码）
func IsSuperAdmin(userID uint) bool {
	if roleCodeFetcher == nil || userID == 0 {
		return false
	}
	codes, err := roleCodeFetcher(userID)
	if err != nil {
		return false
	}
	for _, code := range codes {
		if code == SuperAdminRoleCode {
			return true
		}
	}
	return false
}

// RequirePermission 需要指定权限的中间件
// 用法：router.POST("/admin/users", middleware.RequirePermission("user:create"), handler)
// 超级管理员（admin 角色）直通所有权限校验
func RequirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusOK, response.Unauthorized("请先登录"))
			c.Abort()
			return
		}

		// 超级管理员直通
		if IsSuperAdmin(userID) {
			c.Next()
			return
		}

		// 如果未注入权限校验器，默认放行（开发阶段）
		if permissionChecker == nil {
			c.Next()
			return
		}

		ok, err := permissionChecker(userID, permCode)
		if err != nil {
			c.JSON(http.StatusOK, response.ServerError("权限校验失败"))
			c.Abort()
			return
		}
		if !ok {
			c.JSON(http.StatusOK, response.Forbidden("权限不足，需要权限: "+permCode))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 需要指定角色的中间件（基于角色code列表，任一匹配即通过）
// 用法：router.POST("/admin/users", middleware.RequireRole("admin"), handler)
func RequireRole(roleChecker func(userID uint) ([]string, error), roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusOK, response.Unauthorized("请先登录"))
			c.Abort()
			return
		}

		if roleChecker == nil {
			c.Next()
			return
		}

		userRoles, err := roleChecker(userID)
		if err != nil {
			c.JSON(http.StatusOK, response.ServerError("角色校验失败"))
			c.Abort()
			return
		}

		roleMap := make(map[string]struct{}, len(userRoles))
		for _, r := range userRoles {
			roleMap[r] = struct{}{}
		}
		for _, need := range roles {
			if _, ok := roleMap[need]; ok {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusOK, response.Forbidden("需要角色权限"))
		c.Abort()
	}
}
