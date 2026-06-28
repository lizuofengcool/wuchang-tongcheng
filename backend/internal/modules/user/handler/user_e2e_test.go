// Package handler 用户模块 HTTP 端到端集成测试。
//
// 用真实 PostgreSQL 容器 + gin 引擎 + httptest，验证完整 HTTP 链路：
// 全局中间件（Region + Auth）→ 路由 → handler → service → repository → DB。
//
// 覆盖场景：
//   - 注册（成功 / 重复用户名 / 参数校验失败）
//   - 登录（成功 / 密码错误 / 用户不存在）
//   - 鉴权（无 token 401 / 有效 token 200 / 无效 token 视为游客）
//   - 个人资料（更新资料 / 修改密码后旧密码失效）
//   - 管理后台权限（admin 超管直通 / 普通用户 403）
//
// 无 Docker 时自动 skip。
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	userHandler "wuchang-tongcheng/internal/modules/user/handler"
	userRepo "wuchang-tongcheng/internal/modules/user/repository"
	userService "wuchang-tongcheng/internal/modules/user/service"
	permRepo "wuchang-tongcheng/internal/modules/permission/repository"
	permService "wuchang-tongcheng/internal/modules/permission/service"
	"wuchang-tongcheng/internal/pkg/seed"
	"wuchang-tongcheng/internal/testutil/pgtest"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// e2eEnv 端到端测试环境
type e2eEnv struct {
	engine *gin.Engine
}

// setupUserE2E 构造 gin 引擎并注册 user 路由，注入真实 DB。
// 同时执行 seed.Run 创建 admin 超管 + 全部权限码，并注入权限校验器。
func setupUserE2E(t *testing.T) *e2eEnv {
	t.Helper()
	db := pgtest.SetupPostgres(t)
	pgtest.MigrateAll(t, db)

	// seed 创建地区/权限码/admin 角色/admin 用户（admin/admin123）
	require.NoError(t, seed.Run(db))

	// JWT 用默认 secret 即可（jwt 包默认值已就绪，无需 Init）
	// 权限校验器：用真实 permission service 注入（admin 直通 + 普通权限校验）
	pRepo := permRepo.NewPermissionRepository(db)
	pSvc := permService.NewPermissionService(pRepo)
	middleware.SetPermissionChecker(pSvc.HasPermission)
	middleware.SetRoleCodeFetcher(pSvc.GetRoleCodesByUserID)
	t.Cleanup(func() {
		// 重置全局权限校验器，避免影响其他测试
		middleware.SetPermissionChecker(nil)
		middleware.SetRoleCodeFetcher(nil)
	})

	// 构造 gin 引擎 + 全局中间件（顺序：Region 先于 Auth，保证 handler 能拿到 region_id）
	gin.SetMode(gin.TestMode)
	r := coreRouter.NewRouter()
	r.Use(middleware.Region(), middleware.Auth())

	// user 模块依赖链（绕过 database 全局单例，直接注入 DB）
	uRepo := userRepo.NewUserRepository(db)
	uSvc := userService.NewUserService(uRepo)
	h := userHandler.NewHandler(uSvc)

	// 注册 user 路由（与 user/plugin.go RegisterRoutes 保持一致）
	root := r.Group("/api/v1/user")
	root.POST("/register", h.Register)
	root.POST("/login", coreRouter.WrapGin(middleware.RateLimit(5, 60, "login")), h.Login)

	auth := coreRouter.WrapGin(middleware.AuthRequired())
	root.GET("/info", auth, h.GetUserInfo)
	root.PUT("/profile", auth, h.UpdateProfile)
	root.PUT("/password", auth, h.ChangePassword)

	admin := root.Group("/admin")
	admin.GET("/users", coreRouter.WrapGin(middleware.RequirePermission("user:read")), h.ListUsers)

	return &e2eEnv{engine: r.Engine()}
}

// doJSON 发起 JSON 请求，返回响应体。
// token 非空时设置 Authorization: Bearer <token>；regionID>0 时设置 X-Region-ID 头。
func (e *e2eEnv) doJSON(t *testing.T, method, path string, body interface{}, token string, regionID uint) *apiResponse {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if regionID > 0 {
		req.Header.Set("X-Region-ID", itoa(regionID))
	}

	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp), "响应体非 JSON: %s", w.Body.String())
	}
	return &resp
}

// loginAndGetToken 登录并提取 token（失败则 t.Fatal）
func (e *e2eEnv) loginAndGetToken(t *testing.T, username, password string) string {
	t.Helper()
	resp := e.doJSON(t, "POST", "/api/v1/user/login", map[string]string{
		"username": username,
		"password": password,
	}, "", 0)
	require.Equal(t, 0, resp.Code, "登录失败: %s", resp.Message)

	var data struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	require.NotEmpty(t, data.Token, "token 不应为空")
	return data.Token
}

// itoa 简单 uint 转字符串（避免引入 strconv 占行）
func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// ===== 测试用例 =====

// TestE2E_User_RegisterSuccess 注册成功
func TestE2E_User_RegisterSuccess(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "newuser1",
		"password": "pass123456",
		"nickname": "NewUser",
	}, "", 2)

	assert.Equal(t, 0, resp.Code, "注册应成功: %s", resp.Message)

	var data struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, "newuser1", data.Username)
	assert.Equal(t, "NewUser", data.Nickname)
	assert.NotZero(t, data.ID)
}

// TestE2E_User_RegisterDuplicate 重复用户名注册失败
func TestE2E_User_RegisterDuplicate(t *testing.T) {
	e := setupUserE2E(t)

	// 第一次注册
	resp1 := e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "dupuser",
		"password": "pass123456",
	}, "", 2)
	require.Equal(t, 0, resp1.Code)

	// 重复注册
	resp2 := e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "dupuser",
		"password": "pass123456",
	}, "", 2)
	assert.NotEqual(t, 0, resp2.Code, "重复用户名应失败")
	assert.Contains(t, resp2.Message, "已存在")
}

// TestE2E_User_RegisterInvalidParams 参数校验失败（密码太短）
func TestE2E_User_RegisterInvalidParams(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "shortpwd",
		"password": "123", // 少于 6 位
	}, "", 2)
	assert.NotEqual(t, 0, resp.Code, "密码过短应校验失败")
}

// TestE2E_User_LoginSuccess 登录成功并拿到 token
func TestE2E_User_LoginSuccess(t *testing.T) {
	e := setupUserE2E(t)

	// 先注册
	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "loginuser",
		"password": "pass123456",
	}, "", 2)

	token := e.loginAndGetToken(t, "loginuser", "pass123456")
	assert.NotEmpty(t, token)
}

// TestE2E_User_LoginWrongPassword 密码错误登录失败
func TestE2E_User_LoginWrongPassword(t *testing.T) {
	e := setupUserE2E(t)

	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "pwduser",
		"password": "pass123456",
	}, "", 2)

	resp := e.doJSON(t, "POST", "/api/v1/user/login", map[string]string{
		"username": "pwduser",
		"password": "wrongpassword",
	}, "", 0)
	assert.NotEqual(t, 0, resp.Code, "密码错误应登录失败")
	assert.Contains(t, resp.Message, "密码")
}

// TestE2E_User_LoginUserNotExist 用户不存在
func TestE2E_User_LoginUserNotExist(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "POST", "/api/v1/user/login", map[string]string{
		"username": "ghost_user",
		"password": "whatever",
	}, "", 0)
	assert.NotEqual(t, 0, resp.Code, "不存在的用户应登录失败")
}

// TestE2E_User_GetInfoWithoutToken 无 token 访问受保护接口 401
func TestE2E_User_GetInfoWithoutToken(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "GET", "/api/v1/user/info", nil, "", 0)
	assert.Equal(t, 401, resp.Code, "无 token 应返回 401")
	assert.Contains(t, resp.Message, "登录")
}

// TestE2E_User_GetInfoWithToken 有效 token 访问 200
func TestE2E_User_GetInfoWithToken(t *testing.T) {
	e := setupUserE2E(t)

	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "infouser",
		"password": "pass123456",
		"nickname": "Info",
	}, "", 2)
	token := e.loginAndGetToken(t, "infouser", "pass123456")

	resp := e.doJSON(t, "GET", "/api/v1/user/info", nil, token, 0)
	require.Equal(t, 0, resp.Code, resp.Message)

	var data struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, "infouser", data.Username)
	assert.Equal(t, "Info", data.Nickname)
}

// TestE2E_User_GetInfoWithInvalidToken 无效 token 视为游客 → 401
func TestE2E_User_GetInfoWithInvalidToken(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "GET", "/api/v1/user/info", nil, "invalid.token.here", 0)
	assert.Equal(t, 401, resp.Code, "无效 token 应视为游客返回 401")
}

// TestE2E_User_UpdateProfile 更新个人资料后查询生效
func TestE2E_User_UpdateProfile(t *testing.T) {
	e := setupUserE2E(t)

	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "profileuser",
		"password": "pass123456",
	}, "", 2)
	token := e.loginAndGetToken(t, "profileuser", "pass123456")

	// 更新昵称和邮箱
	resp := e.doJSON(t, "PUT", "/api/v1/user/profile", map[string]interface{}{
		"nickname": "ProfileUpdated",
		"email":    "updated@example.com",
		"gender":   2,
	}, token, 0)
	require.Equal(t, 0, resp.Code, resp.Message)

	// 查询验证
	infoResp := e.doJSON(t, "GET", "/api/v1/user/info", nil, token, 0)
	require.Equal(t, 0, infoResp.Code)

	var data struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Gender   int    `json:"gender"`
	}
	require.NoError(t, json.Unmarshal(infoResp.Data, &data))
	assert.Equal(t, "ProfileUpdated", data.Nickname)
	assert.Equal(t, "updated@example.com", data.Email)
	assert.Equal(t, 2, data.Gender)
}

// TestE2E_User_ChangePassword 修改密码后旧密码失效、新密码可用
func TestE2E_User_ChangePassword(t *testing.T) {
	e := setupUserE2E(t)

	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "pwdchange",
		"password": "oldpass123",
	}, "", 2)
	token := e.loginAndGetToken(t, "pwdchange", "oldpass123")

	// 修改密码
	resp := e.doJSON(t, "PUT", "/api/v1/user/password", map[string]string{
		"old_password": "oldpass123",
		"new_password": "newpass456",
	}, token, 0)
	require.Equal(t, 0, resp.Code, resp.Message)

	// 旧密码登录失败
	oldResp := e.doJSON(t, "POST", "/api/v1/user/login", map[string]string{
		"username": "pwdchange",
		"password": "oldpass123",
	}, "", 0)
	assert.NotEqual(t, 0, oldResp.Code, "旧密码应登录失败")

	// 新密码登录成功
	newToken := e.loginAndGetToken(t, "pwdchange", "newpass456")
	assert.NotEmpty(t, newToken)
}

// TestE2E_User_ChangePassword_OldWrong 原密码错误修改失败
func TestE2E_User_ChangePassword_OldWrong(t *testing.T) {
	e := setupUserE2E(t)

	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "pwdwrong",
		"password": "pass123456",
	}, "", 2)
	token := e.loginAndGetToken(t, "pwdwrong", "pass123456")

	resp := e.doJSON(t, "PUT", "/api/v1/user/password", map[string]string{
		"old_password": "wrongold",
		"new_password": "newpass789",
	}, token, 0)
	assert.NotEqual(t, 0, resp.Code, "原密码错误应修改失败")
}

// TestE2E_Admin_SuperAdminDirectPass admin 超管直通权限校验
func TestE2E_Admin_SuperAdminDirectPass(t *testing.T) {
	e := setupUserE2E(t)

	// seed 创建的 admin/admin123 登录
	token := e.loginAndGetToken(t, "admin", "admin123")

	// 访问 admin/users，admin 角色直通
	resp := e.doJSON(t, "GET", "/api/v1/user/admin/users?page=1&page_size=10", nil, token, 0)
	require.Equal(t, 0, resp.Code, "admin 超管应直通: %s", resp.Message)

	var data struct {
		Total int64 `json:"total"`
		List  []struct {
			Username string `json:"username"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.GreaterOrEqual(t, data.Total, int64(1), "至少有 admin 自己")
	// 应包含 admin 用户
	found := false
	for _, u := range data.List {
		if u.Username == "admin" {
			found = true
			break
		}
	}
	assert.True(t, found, "用户列表应包含 admin")
}

// TestE2E_Admin_NormalUserForbidden 普通用户访问 admin 接口 403
func TestE2E_Admin_NormalUserForbidden(t *testing.T) {
	e := setupUserE2E(t)

	// 注册普通用户
	e.doJSON(t, "POST", "/api/v1/user/register", map[string]string{
		"username": "normaluser",
		"password": "pass123456",
	}, "", 2)
	token := e.loginAndGetToken(t, "normaluser", "pass123456")

	// 访问 admin/users，普通用户无 user:read 权限 → 403
	resp := e.doJSON(t, "GET", "/api/v1/user/admin/users?page=1&page_size=10", nil, token, 0)
	assert.Equal(t, 403, resp.Code, "普通用户应 403")
	assert.Contains(t, resp.Message, "权限")
}

// TestE2E_Admin_NormalUserWithoutToken 401
func TestE2E_Admin_NormalUserWithoutToken(t *testing.T) {
	e := setupUserE2E(t)

	resp := e.doJSON(t, "GET", "/api/v1/user/admin/users?page=1&page_size=10", nil, "", 0)
	assert.Equal(t, 401, resp.Code, "无 token 访问 admin 应 401")
}
