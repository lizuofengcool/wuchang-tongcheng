// Package middleware 权限中间件单元测试
// 覆盖 IsSuperAdmin、RequirePermission 全路径（未登录/admin 直通/校验器 nil 放行/
// 校验失败/权限不足/通过）、RequireRole。使用 mock 注入，无 DB/Redis 依赖。
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/pkg/jwt"
)

// resetPermissionInjectors 重置全局注入器，避免测试间相互污染
func resetPermissionInjectors(t *testing.T) {
	t.Helper()
	origChecker := permissionChecker
	origFetcher := roleCodeFetcher
	permissionChecker = nil
	roleCodeFetcher = nil
	t.Cleanup(func() {
		permissionChecker = origChecker
		roleCodeFetcher = origFetcher
	})
}

// newPermissionRouter 构造路由：Auth() → RequirePermission("user:create")
func newPermissionRouter() *gin.Engine {
	r := gin.New()
	r.Use(Auth())
	admin := r.Group("/admin", RequirePermission("user:create"))
	admin.POST("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok"})
	})
	return r
}

// ===== IsSuperAdmin =====

func TestIsSuperAdmin_UserIDZero(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = func(uint) ([]string, error) { return []string{"admin"}, nil }
	assert.False(t, IsSuperAdmin(0), "userID=0 应返回 false")
}

func TestIsSuperAdmin_FetcherNil(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = nil
	assert.False(t, IsSuperAdmin(100), "fetcher=nil 应返回 false")
}

func TestIsSuperAdmin_FetcherError(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = func(uint) ([]string, error) { return nil, errors.New("db down") }
	assert.False(t, IsSuperAdmin(100), "fetcher 出错应返回 false")
}

func TestIsSuperAdmin_NoAdminCode(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = func(uint) ([]string, error) { return []string{"editor", "auditor"}, nil }
	assert.False(t, IsSuperAdmin(100), "无 admin 角色码应返回 false")
}

func TestIsSuperAdmin_HasAdminCode(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = func(uint) ([]string, error) { return []string{"editor", "admin"}, nil }
	assert.True(t, IsSuperAdmin(100), "包含 admin 角色码应返回 true")
}

// ===== RequirePermission 全路径 =====

// TestRequirePermission_NoToken 未登录返回 401
func TestRequirePermission_NoToken(t *testing.T) {
	resetPermissionInjectors(t)
	r := newPermissionRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":401`)
	assert.Contains(t, w.Body.String(), "请先登录")
}

// TestRequirePermission_SuperAdminDirectPass 超级管理员直通
func TestRequirePermission_SuperAdminDirectPass(t *testing.T) {
	resetPermissionInjectors(t)
	roleCodeFetcher = func(uint) ([]string, error) { return []string{"admin"}, nil }
	// 即使 permissionChecker 返回 false，admin 也应直通
	permissionChecker = func(uint, string) (bool, error) { return false, nil }

	token, _ := jwt.GenerateToken(1, "root")
	r := newPermissionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestRequirePermission_CheckerNil 开发阶段未注入校验器，默认放行
func TestRequirePermission_CheckerNil(t *testing.T) {
	resetPermissionInjectors(t)
	// roleCodeFetcher 也为 nil，IsSuperAdmin 返回 false，permissionChecker 为 nil → 放行
	token, _ := jwt.GenerateToken(2, "alice")
	r := newPermissionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestRequirePermission_CheckerError 校验失败返回 500
func TestRequirePermission_CheckerError(t *testing.T) {
	resetPermissionInjectors(t)
	permissionChecker = func(uint, string) (bool, error) { return false, errors.New("db down") }

	token, _ := jwt.GenerateToken(3, "bob")
	r := newPermissionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	assert.Contains(t, w.Body.String(), "权限校验失败")
}

// TestRequirePermission_Forbidden 普通用户无权限返回 403
func TestRequirePermission_Forbidden(t *testing.T) {
	resetPermissionInjectors(t)
	permissionChecker = func(uint, string) (bool, error) { return false, nil }

	token, _ := jwt.GenerateToken(4, "carol")
	r := newPermissionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":403`)
	assert.Contains(t, w.Body.String(), "权限不足")
	assert.Contains(t, w.Body.String(), "user:create")
}

// TestRequirePermission_Pass 拥有权限通过
func TestRequirePermission_Pass(t *testing.T) {
	resetPermissionInjectors(t)
	permissionChecker = func(uint, string) (bool, error) { return true, nil }

	token, _ := jwt.GenerateToken(5, "dave")
	r := newPermissionRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// ===== RequireRole =====

func newRoleRouter(checker func(uint) ([]string, error), roles ...string) *gin.Engine {
	r := gin.New()
	r.Use(Auth())
	admin := r.Group("/admin", RequireRole(checker, roles...))
	admin.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})
	return r
}

// TestRequireRole_NoToken 未登录返回 401
func TestRequireRole_NoToken(t *testing.T) {
	r := newRoleRouter(func(uint) ([]string, error) { return []string{"editor"}, nil }, "editor")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":401`)
}

// TestRequireRole_CheckerNil 未注入校验器，默认放行
func TestRequireRole_CheckerNil(t *testing.T) {
	r := newRoleRouter(nil, "editor")
	token, _ := jwt.GenerateToken(10, "alice")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestRequireRole_Hit 拥有任一角色即通过
func TestRequireRole_Hit(t *testing.T) {
	checker := func(uint) ([]string, error) { return []string{"editor", "auditor"}, nil }
	r := newRoleRouter(checker, "admin", "auditor")
	token, _ := jwt.GenerateToken(11, "bob")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

// TestRequireRole_Miss 无任一指定角色返回 403
func TestRequireRole_Miss(t *testing.T) {
	checker := func(uint) ([]string, error) { return []string{"editor"}, nil }
	r := newRoleRouter(checker, "admin", "auditor")
	token, _ := jwt.GenerateToken(12, "carol")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":403`)
}

// TestRequireRole_CheckerError 校验失败返回 500
func TestRequireRole_CheckerError(t *testing.T) {
	checker := func(uint) ([]string, error) { return nil, errors.New("db down") }
	r := newRoleRouter(checker, "admin")
	token, _ := jwt.GenerateToken(13, "dave")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":500`)
	assert.Contains(t, w.Body.String(), "角色校验失败")
}
