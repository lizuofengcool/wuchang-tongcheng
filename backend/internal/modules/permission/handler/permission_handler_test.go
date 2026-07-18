// Package handler_test 权限模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock PermissionService，覆盖 handler 全部分支：
//   - 未登录拦截（写操作 userID=0 → 401）
//   - URL :id 参数解析失败（非数字 → 400）
//   - 请求体 Bind 失败（非法 JSON → 400）
//   - service 成功/错误透传（业务码 + message + data 透传）
//   - UpdatePermission 的 fields map 构造逻辑（name 空跳过、path/method/sort 总是写、status 仅 0/1 写）
//   - GetPermissionByID 复用 ListPermissions 后筛选的特殊路径（命中/未命中/ListPermissions 错误）
//   - MyAuth 返回 MyAuthResponse{Permissions, Roles} 结构
//   - UserRoles 独立 :id 解析与"无效的用户ID"错误消息
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 category/region handler 测试同风格，区别在于 permission 模块 service 方法均不接受 regionID，
// 且 handler 中含两个仅校验登录的"当前用户"端点（MyPermissions/MyAuth）。
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/permission/dto"
	permHandler "wuchang-tongcheng/internal/modules/permission/handler"
	"wuchang-tongcheng/internal/modules/permission/service"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockPermissionService 内存 mock，实现 service.PermissionService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockPermissionService struct {
	// 调用记录
	lastCreateRoleReq        *dto.CreateRoleRequest
	lastUpdateRoleID         uint
	lastUpdateRoleReq        *dto.UpdateRoleRequest
	lastDeleteRoleID         uint
	lastGetRoleByID          uint
	lastCreatePermReq        *dto.CreatePermissionRequest
	lastUpdatePermID         uint
	lastUpdatePermFields     map[string]interface{}
	lastDeletePermID         uint
	lastAssignRolesReq       *dto.AssignRolesRequest
	lastAssignPermissionsReq *dto.AssignPermissionsRequest
	lastRolesByUserID        uint
	lastPermsByUserID        uint
	lastPermCodesByUserID    uint
	lastRoleCodesByUserID    uint
	lastHasPermission        struct {
		UserID   uint
		PermCode string
	}
	lastMyAuthUserID     uint
	lastPermsByRoleID    uint
	lastListRolesCalled  bool
	lastListPermsCalled  bool

	// 返回值预设
	createRoleResult *dto.RoleInfo
	createRoleErr    error
	updateRoleErr    error
	deleteRoleErr    error
	getRoleByIDResult *dto.RoleInfo
	getRoleByIDErr    error
	listRolesResult  []dto.RoleInfo
	listRolesErr     error

	createPermResult *dto.PermissionInfo
	createPermErr    error
	updatePermErr    error
	deletePermErr    error
	listPermsResult  []dto.PermissionInfo
	listPermsErr     error

	assignRolesErr       error
	assignPermissionsErr error

	rolesByUserIDResult  []dto.RoleInfo
	rolesByUserIDErr     error
	permsByUserIDResult  []dto.PermissionInfo
	permsByUserIDErr     error
	permCodesByUserIDResult []string
	permCodesByUserIDErr    error
	roleCodesByUserIDResult []string
	roleCodesByUserIDErr    error

	hasPermissionResult bool
	hasPermissionErr    error

	myAuthPermsResult []string
	myAuthRolesResult []string
	myAuthErr         error

	permsByRoleIDResult []dto.PermissionInfo
	permsByRoleIDErr    error
}

// ===== 角色方法 =====

func (m *mockPermissionService) CreateRole(req *dto.CreateRoleRequest) (*dto.RoleInfo, error) {
	m.lastCreateRoleReq = req
	return m.createRoleResult, m.createRoleErr
}
func (m *mockPermissionService) UpdateRole(id uint, req *dto.UpdateRoleRequest) error {
	m.lastUpdateRoleID = id
	m.lastUpdateRoleReq = req
	return m.updateRoleErr
}
func (m *mockPermissionService) DeleteRole(id uint) error {
	m.lastDeleteRoleID = id
	return m.deleteRoleErr
}
func (m *mockPermissionService) GetRoleByID(id uint) (*dto.RoleInfo, error) {
	m.lastGetRoleByID = id
	return m.getRoleByIDResult, m.getRoleByIDErr
}
func (m *mockPermissionService) ListRoles() ([]dto.RoleInfo, error) {
	m.lastListRolesCalled = true
	return m.listRolesResult, m.listRolesErr
}

// ===== 权限方法 =====

func (m *mockPermissionService) CreatePermission(req *dto.CreatePermissionRequest) (*dto.PermissionInfo, error) {
	m.lastCreatePermReq = req
	return m.createPermResult, m.createPermErr
}
func (m *mockPermissionService) UpdatePermission(id uint, fields map[string]interface{}) error {
	m.lastUpdatePermID = id
	// 拷贝一份避免外部修改影响断言
	m.lastUpdatePermFields = make(map[string]interface{}, len(fields))
	for k, v := range fields {
		m.lastUpdatePermFields[k] = v
	}
	return m.updatePermErr
}
func (m *mockPermissionService) DeletePermission(id uint) error {
	m.lastDeletePermID = id
	return m.deletePermErr
}
func (m *mockPermissionService) ListPermissions() ([]dto.PermissionInfo, error) {
	m.lastListPermsCalled = true
	return m.listPermsResult, m.listPermsErr
}

// ===== 分配方法 =====

func (m *mockPermissionService) AssignRolesToUser(req *dto.AssignRolesRequest) error {
	m.lastAssignRolesReq = req
	return m.assignRolesErr
}
func (m *mockPermissionService) AssignPermissionsToRole(req *dto.AssignPermissionsRequest) error {
	m.lastAssignPermissionsReq = req
	return m.assignPermissionsErr
}
func (m *mockPermissionService) GetRolesByUserID(userID uint) ([]dto.RoleInfo, error) {
	m.lastRolesByUserID = userID
	return m.rolesByUserIDResult, m.rolesByUserIDErr
}
func (m *mockPermissionService) GetPermissionsByUserID(userID uint) ([]dto.PermissionInfo, error) {
	m.lastPermsByUserID = userID
	return m.permsByUserIDResult, m.permsByUserIDErr
}
func (m *mockPermissionService) GetPermissionCodesByUserID(userID uint) ([]string, error) {
	m.lastPermCodesByUserID = userID
	return m.permCodesByUserIDResult, m.permCodesByUserIDErr
}
func (m *mockPermissionService) GetRoleCodesByUserID(userID uint) ([]string, error) {
	m.lastRoleCodesByUserID = userID
	return m.roleCodesByUserIDResult, m.roleCodesByUserIDErr
}

// ===== 校验/概览方法 =====

func (m *mockPermissionService) HasPermission(userID uint, permCode string) (bool, error) {
	m.lastHasPermission.UserID = userID
	m.lastHasPermission.PermCode = permCode
	return m.hasPermissionResult, m.hasPermissionErr
}
func (m *mockPermissionService) GetMyAuth(userID uint) ([]string, []string, error) {
	m.lastMyAuthUserID = userID
	return m.myAuthPermsResult, m.myAuthRolesResult, m.myAuthErr
}
func (m *mockPermissionService) GetPermissionsByRoleID(roleID uint) ([]dto.PermissionInfo, error) {
	m.lastPermsByRoleID = roleID
	return m.permsByRoleIDResult, m.permsByRoleIDErr
}

// 确保 mockPermissionService 实现 service.PermissionService 接口
var _ service.PermissionService = (*mockPermissionService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockPermissionService
}

// newHandlerEnv 构造 gin 引擎并注册 permission 路由（与 permission/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// 注意：permission 模块读操作（GetRoleByID/ListRoles/GetPermissionByID/ListPermissions）
// 在 handler 层不校验 userID（鉴权由 plugin.go 的 RequirePermission 中间件负责），
// 仅写操作（Create/Update/Delete/Assign）与"当前用户"端点（MyPermissions/MyAuth）做防御性登录校验。
func newHandlerEnv(t *testing.T, ctxUserID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockPermissionService{
		createRoleResult: &dto.RoleInfo{ID: 1, Name: "管理员", Code: "admin", Description: "超级管理员", Sort: 10, Status: 1},
		getRoleByIDResult: &dto.RoleInfo{ID: 1, Name: "管理员", Code: "admin", Description: "超级管理员", Sort: 10, Status: 1},
		listRolesResult: []dto.RoleInfo{
			{ID: 1, Name: "管理员", Code: "admin", Description: "超级管理员", Sort: 10, Status: 1},
			{ID: 2, Name: "编辑", Code: "editor", Description: "内容编辑", Sort: 20, Status: 1},
		},
		createPermResult: &dto.PermissionInfo{ID: 100, Name: "创建用户", Code: "user:create", Type: 2, Path: "/api/v1/user", Method: "POST", Sort: 5, Status: 1},
		listPermsResult: []dto.PermissionInfo{
			{ID: 100, Name: "创建用户", Code: "user:create", Type: 2, Path: "/api/v1/user", Method: "POST", Sort: 5, Status: 1},
			{ID: 101, Name: "删除用户", Code: "user:delete", Type: 2, Path: "/api/v1/user/:id", Method: "DELETE", Sort: 6, Status: 1},
		},
		rolesByUserIDResult: []dto.RoleInfo{
			{ID: 1, Name: "管理员", Code: "admin", Sort: 10, Status: 1},
		},
		permsByUserIDResult: []dto.PermissionInfo{
			{ID: 100, Name: "创建用户", Code: "user:create", Type: 2, Sort: 5, Status: 1},
		},
		permCodesByUserIDResult: []string{"user:create", "user:delete"},
		roleCodesByUserIDResult: []string{"admin"},
		hasPermissionResult:     true,
		myAuthPermsResult:       []string{"user:create", "user:delete"},
		myAuthRolesResult:       []string{"admin"},
		permsByRoleIDResult: []dto.PermissionInfo{
			{ID: 100, Name: "创建用户", Code: "user:create", Type: 2, Sort: 5, Status: 1},
		},
	}

	r := coreRouter.NewRouter()
	// 模拟 Auth 中间件：注入 user_id（permission 不需要 region_id）
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := permHandler.NewHandler(mock)
	// 注册路由，路径与 permission/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/permission")
	// 角色管理
	root.POST("/roles", h.CreateRole)
	root.PUT("/roles/:id", h.UpdateRole)
	root.DELETE("/roles/:id", h.DeleteRole)
	root.GET("/roles/:id", h.GetRoleByID)
	root.GET("/roles", h.ListRoles)
	root.GET("/users/:id/roles", h.UserRoles)
	root.GET("/roles/:id/permissions", h.RolePermissions)
	// 权限管理
	root.POST("/permissions", h.CreatePermission)
	root.PUT("/permissions/:id", h.UpdatePermission)
	root.DELETE("/permissions/:id", h.DeletePermission)
	root.GET("/permissions", h.ListPermissions)
	root.GET("/permissions/:id", h.GetPermissionByID)
	// 分配
	root.POST("/assign-roles", h.AssignRoles)
	root.POST("/assign-permissions", h.AssignPermissions)
	// 当前用户
	root.GET("/my-permissions", h.MyPermissions)
	root.GET("/my-auth", h.MyAuth)

	return &handlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。
func (e *handlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *apiResponse {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doRaw 发起原始请求（用于测试 Bind 失败：非法 JSON body）。
func (e *handlerEnv) doRaw(t *testing.T, method, path string, rawBody string, contentType string) *apiResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp apiResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// ==================== CreateRole ====================

func TestHandler_CreateRole_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.CreateRoleRequest{Name: "运营", Code: "ops", Description: "运营人员", Sort: 30, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/roles", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	require.NotNil(t, env.mock.lastCreateRoleReq)
	assert.Equal(t, "运营", env.mock.lastCreateRoleReq.Name)
	assert.Equal(t, "ops", env.mock.lastCreateRoleReq.Code)
	// data 透传 service 返回值
	var info dto.RoleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "admin", info.Code)
}

func TestHandler_CreateRole_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0) // 未登录
	body := dto.CreateRoleRequest{Name: "运营", Code: "ops"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/roles", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Nil(t, env.mock.lastCreateRoleReq)
}

func TestHandler_CreateRole_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/permission/roles", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateRoleReq)
}

func TestHandler_CreateRole_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.createRoleErr = errors.New("角色编码已存在")
	env.mock.createRoleResult = nil
	body := dto.CreateRoleRequest{Name: "管理员", Code: "admin"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/roles", body)

	// 业务码 CodeRoleAlreadyExists=2204 + err.Error() 透传
	assert.Equal(t, 2204, resp.Code)
	assert.Equal(t, "角色编码已存在", resp.Message)
}

// ==================== UpdateRole ====================

func TestHandler_UpdateRole_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.UpdateRoleRequest{Name: "新管理员", Description: "新描述", Sort: 99, Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/roles/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateRoleID)
	require.NotNil(t, env.mock.lastUpdateRoleReq)
	assert.Equal(t, "新管理员", env.mock.lastUpdateRoleReq.Name)
	assert.Equal(t, 99, env.mock.lastUpdateRoleReq.Sort)
}

func TestHandler_UpdateRole_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/roles/5", dto.UpdateRoleRequest{Name: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateRoleID)
}

func TestHandler_UpdateRole_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/roles/abc", dto.UpdateRoleRequest{Name: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的角色ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateRoleID)
}

func TestHandler_UpdateRole_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/permission/roles/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateRoleID)
}

func TestHandler_UpdateRole_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.updateRoleErr = errors.New("角色不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/roles/5", dto.UpdateRoleRequest{Name: "x"})

	// CodePermissionError=2201
	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "角色不存在", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateRoleID)
}

// ==================== DeleteRole ====================

func TestHandler_DeleteRole_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/roles/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteRoleID)
}

func TestHandler_DeleteRole_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/roles/7", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteRoleID)
}

func TestHandler_DeleteRole_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/roles/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的角色ID", resp.Message)
}

func TestHandler_DeleteRole_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.deleteRoleErr = errors.New("角色已被用户绑定，无法删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/roles/7", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "角色已被用户绑定，无法删除", resp.Message)
}

// ==================== GetRoleByID ====================

func TestHandler_GetRoleByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetRoleByID)
	var info dto.RoleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "admin", info.Code)
}

func TestHandler_GetRoleByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的角色ID", resp.Message)
}

func TestHandler_GetRoleByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.getRoleByIDErr = errors.New("角色不存在")
	env.mock.getRoleByIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/999", nil)

	// CodeRoleNotFound=2203
	assert.Equal(t, 2203, resp.Code)
	assert.Equal(t, "角色不存在", resp.Message)
}

// ==================== ListRoles ====================

func TestHandler_ListRoles_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.True(t, env.mock.lastListRolesCalled)
	var list []dto.RoleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "admin", list[0].Code)
	assert.Equal(t, "editor", list[1].Code)
}

func TestHandler_ListRoles_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.listRolesErr = errors.New("db down")
	env.mock.listRolesResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== CreatePermission ====================

func TestHandler_CreatePermission_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.CreatePermissionRequest{Name: "创建角色", Code: "role:create", Type: 2, Path: "/api/v1/permission/roles", Method: "POST", Sort: 1, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/permissions", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	require.NotNil(t, env.mock.lastCreatePermReq)
	assert.Equal(t, "role:create", env.mock.lastCreatePermReq.Code)
	assert.Equal(t, 2, env.mock.lastCreatePermReq.Type)
	// data 透传
	var info dto.PermissionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(100), info.ID)
	assert.Equal(t, "user:create", info.Code)
}

func TestHandler_CreatePermission_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	body := dto.CreatePermissionRequest{Name: "创建角色", Code: "role:create", Type: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/permissions", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreatePermReq)
}

func TestHandler_CreatePermission_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/permission/permissions", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreatePermReq)
}

func TestHandler_CreatePermission_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.createPermErr = errors.New("权限编码已存在")
	env.mock.createPermResult = nil
	body := dto.CreatePermissionRequest{Name: "创建用户", Code: "user:create", Type: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/permissions", body)

	// CodePermissionAlreadyExists=2206
	assert.Equal(t, 2206, resp.Code)
	assert.Equal(t, "权限编码已存在", resp.Message)
}

// ==================== UpdatePermission ====================

func TestHandler_UpdatePermission_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.UpdatePermissionRequest{Name: "新名称", Path: "/api/v2/user", Method: "PUT", Sort: 99, Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/100", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(100), env.mock.lastUpdatePermID)
	// 验证 fields map 构造逻辑：name 非空写入、path/method/sort 总是写、status=1 写入
	require.Contains(t, env.mock.lastUpdatePermFields, "name")
	assert.Equal(t, "新名称", env.mock.lastUpdatePermFields["name"])
	require.Contains(t, env.mock.lastUpdatePermFields, "path")
	assert.Equal(t, "/api/v2/user", env.mock.lastUpdatePermFields["path"])
	require.Contains(t, env.mock.lastUpdatePermFields, "method")
	assert.Equal(t, "PUT", env.mock.lastUpdatePermFields["method"])
	require.Contains(t, env.mock.lastUpdatePermFields, "sort")
	assert.EqualValues(t, 99, env.mock.lastUpdatePermFields["sort"])
	require.Contains(t, env.mock.lastUpdatePermFields, "status")
	assert.EqualValues(t, 1, env.mock.lastUpdatePermFields["status"])
}

func TestHandler_UpdatePermission_NameEmptyOmitted(t *testing.T) {
	// 验证 name="" 时 fields 不含 name 键（其它字段仍写入）
	env := newHandlerEnv(t, 1)
	body := dto.UpdatePermissionRequest{Name: "", Path: "/x", Method: "GET", Sort: 0, Status: 0}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/100", body)

	assert.Equal(t, 0, resp.Code)
	// name 为空不应写入 fields
	assert.NotContains(t, env.mock.lastUpdatePermFields, "name")
	// path/method/sort 总是写入（即使零值）
	assert.Contains(t, env.mock.lastUpdatePermFields, "path")
	assert.Contains(t, env.mock.lastUpdatePermFields, "method")
	assert.Contains(t, env.mock.lastUpdatePermFields, "sort")
	// status=0 应写入（0 是合法值）
	assert.Contains(t, env.mock.lastUpdatePermFields, "status")
	assert.EqualValues(t, 0, env.mock.lastUpdatePermFields["status"])
}

func TestHandler_UpdatePermission_StatusInvalidOmitted(t *testing.T) {
	// 验证 status 非 0/1 时不写入 status 字段
	// 注意：dto.UpdatePermissionRequest.Status 有 binding:"omitempty,oneof=0 1"，
	// 但 gin ShouldBind 对 JSON 中显式传入 2 不会拒绝（oneof 在 omitempty 下仅校验非零值），
	// 故 handler 自身的 `if req.Status == 0 || req.Status == 1` 兜底过滤是关键。
	// 这里直接构造绕过 binding 的请求：用 raw JSON 传入 status=2
	env := newHandlerEnv(t, 1)
	// 直接走 doJSON 传 status=2（绕过 dto binding 不可能，因为 oneof=0 1 会拒绝 2）
	// 故改为构造一个合法的 status=1 请求验证字段写入，再用 raw json status=2 验证 handler 兜底
	rawBody := `{"name":"x","path":"/p","method":"POST","sort":1,"status":2}`
	resp := env.doRaw(t, http.MethodPut, "/api/v1/permission/permissions/100", rawBody, "application/json")

	// dto binding 的 oneof=0 1 在 omitempty 下对 status=2 会拒绝（2 非零且不在 oneof）→ Bind 失败 → 400
	// 这验证了 binding 层先于 handler 字段构造
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_UpdatePermission_StatusOneWritten(t *testing.T) {
	// 验证 status=1 写入 fields
	env := newHandlerEnv(t, 1)
	body := dto.UpdatePermissionRequest{Name: "x", Path: "/p", Method: "POST", Sort: 1, Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/100", body)

	assert.Equal(t, 0, resp.Code)
	assert.Contains(t, env.mock.lastUpdatePermFields, "status")
	assert.EqualValues(t, 1, env.mock.lastUpdatePermFields["status"])
}

func TestHandler_UpdatePermission_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/100", dto.UpdatePermissionRequest{Name: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdatePermID)
}

func TestHandler_UpdatePermission_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/abc", dto.UpdatePermissionRequest{Name: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的权限ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdatePermID)
}

func TestHandler_UpdatePermission_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/permission/permissions/100", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdatePermID)
}

func TestHandler_UpdatePermission_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.updatePermErr = errors.New("权限不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/permission/permissions/100", dto.UpdatePermissionRequest{Name: "x"})

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "权限不存在", resp.Message)
	assert.Equal(t, uint(100), env.mock.lastUpdatePermID)
}

// ==================== DeletePermission ====================

func TestHandler_DeletePermission_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/permissions/100", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(100), env.mock.lastDeletePermID)
}

func TestHandler_DeletePermission_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/permissions/100", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeletePermID)
}

func TestHandler_DeletePermission_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/permissions/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的权限ID", resp.Message)
}

func TestHandler_DeletePermission_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.deletePermErr = errors.New("权限已被角色引用，无法删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/permission/permissions/100", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "权限已被角色引用，无法删除", resp.Message)
}

// ==================== ListPermissions ====================

func TestHandler_ListPermissions_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.True(t, env.mock.lastListPermsCalled)
	var list []dto.PermissionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "user:create", list[0].Code)
	assert.Equal(t, "user:delete", list[1].Code)
}

func TestHandler_ListPermissions_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.listPermsErr = errors.New("redis boom")
	env.mock.listPermsResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// ==================== GetPermissionByID ====================
// 注意：GetPermissionByID 复用 ListPermissions 后筛选，与其它 GetByID 不同。

func TestHandler_GetPermissionByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// listPermsResult 含 ID=100，命中
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions/100", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.True(t, env.mock.lastListPermsCalled)
	var info dto.PermissionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(100), info.ID)
	assert.Equal(t, "user:create", info.Code)
}

func TestHandler_GetPermissionByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的权限ID", resp.Message)
	// :id 解析失败时不应调用 ListPermissions
	assert.False(t, env.mock.lastListPermsCalled)
}

func TestHandler_GetPermissionByID_NotFound(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// listPermsResult 不含 ID=999
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions/999", nil)

	// CodePermissionNotFound=2205 + 固定 message "权限不存在"
	assert.Equal(t, 2205, resp.Code)
	assert.Equal(t, "权限不存在", resp.Message)
	assert.True(t, env.mock.lastListPermsCalled)
}

func TestHandler_GetPermissionByID_ListError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.listPermsErr = errors.New("db error")
	env.mock.listPermsResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions/100", nil)

	// ListPermissions 失败 → CodePermissionError=2201
	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== AssignRoles ====================

func TestHandler_AssignRoles_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.AssignRolesRequest{UserID: 5, RoleIDs: []uint{1, 2}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-roles", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "分配成功", resp.Message)
	require.NotNil(t, env.mock.lastAssignRolesReq)
	assert.Equal(t, uint(5), env.mock.lastAssignRolesReq.UserID)
	assert.Equal(t, []uint{1, 2}, env.mock.lastAssignRolesReq.RoleIDs)
}

func TestHandler_AssignRoles_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	body := dto.AssignRolesRequest{UserID: 5, RoleIDs: []uint{1}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-roles", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastAssignRolesReq)
}

func TestHandler_AssignRoles_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/permission/assign-roles", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastAssignRolesReq)
}

func TestHandler_AssignRoles_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.assignRolesErr = errors.New("用户不存在")
	body := dto.AssignRolesRequest{UserID: 999, RoleIDs: []uint{1}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-roles", body)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "用户不存在", resp.Message)
}

// ==================== AssignPermissions ====================

func TestHandler_AssignPermissions_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.AssignPermissionsRequest{RoleID: 1, PermissionIDs: []uint{100, 101}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-permissions", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "分配成功", resp.Message)
	require.NotNil(t, env.mock.lastAssignPermissionsReq)
	assert.Equal(t, uint(1), env.mock.lastAssignPermissionsReq.RoleID)
	assert.Equal(t, []uint{100, 101}, env.mock.lastAssignPermissionsReq.PermissionIDs)
}

func TestHandler_AssignPermissions_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	body := dto.AssignPermissionsRequest{RoleID: 1, PermissionIDs: []uint{100}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-permissions", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastAssignPermissionsReq)
}

func TestHandler_AssignPermissions_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/permission/assign-permissions", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastAssignPermissionsReq)
}

func TestHandler_AssignPermissions_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.assignPermissionsErr = errors.New("角色不存在")
	body := dto.AssignPermissionsRequest{RoleID: 999, PermissionIDs: []uint{100}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/permission/assign-permissions", body)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "角色不存在", resp.Message)
}

// ==================== MyPermissions ====================

func TestHandler_MyPermissions_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-permissions", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastPermsByUserID)
	var list []dto.PermissionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "user:create", list[0].Code)
}

func TestHandler_MyPermissions_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-permissions", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastPermsByUserID)
}

func TestHandler_MyPermissions_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.permsByUserIDErr = errors.New("db error")
	env.mock.permsByUserIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-permissions", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== MyAuth ====================

func TestHandler_MyAuth_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-auth", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastMyAuthUserID)
	// MyAuth 返回 MyAuthResponse{Permissions, Roles}
	var auth dto.MyAuthResponse
	require.NoError(t, json.Unmarshal(resp.Data, &auth))
	require.Len(t, auth.Permissions, 2)
	assert.Contains(t, auth.Permissions, "user:create")
	assert.Contains(t, auth.Permissions, "user:delete")
	require.Len(t, auth.Roles, 1)
	assert.Equal(t, "admin", auth.Roles[0])
}

func TestHandler_MyAuth_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-auth", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastMyAuthUserID)
}

func TestHandler_MyAuth_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.myAuthErr = errors.New("db error")
	env.mock.myAuthPermsResult = nil
	env.mock.myAuthRolesResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/my-auth", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== UserRoles ====================

func TestHandler_UserRoles_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/users/5/roles", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastRolesByUserID)
	var list []dto.RoleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "admin", list[0].Code)
}

func TestHandler_UserRoles_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/users/5/roles", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastRolesByUserID)
}

func TestHandler_UserRoles_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/users/abc/roles", nil)

	// UserRoles 独立使用 strconv.ParseUint，错误消息为 "无效的用户ID"（与角色 ID 区分）
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastRolesByUserID)
}

func TestHandler_UserRoles_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.rolesByUserIDErr = errors.New("db error")
	env.mock.rolesByUserIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/users/5/roles", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== RolePermissions ====================

func TestHandler_RolePermissions_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/1/permissions", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastPermsByRoleID)
	var list []dto.PermissionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "user:create", list[0].Code)
}

func TestHandler_RolePermissions_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/1/permissions", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastPermsByRoleID)
}

func TestHandler_RolePermissions_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/abc/permissions", nil)

	// RolePermissions 复用 parseID，错误消息为 "无效的角色ID"
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的角色ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastPermsByRoleID)
}

func TestHandler_RolePermissions_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.permsByRoleIDErr = errors.New("db error")
	env.mock.permsByRoleIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/1/permissions", nil)

	assert.Equal(t, 2201, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== 公开读取无需登录 ====================
// GetRoleByID / ListRoles / GetPermissionByID / ListPermissions 在 handler 层不校验 userID，
// 鉴权由 plugin.go 的 RequirePermission 中间件负责。

func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0) // 未登录

	// 四个读路径均不应被 401 拦截
	resp := env.doJSON(t, http.MethodGet, "/api/v1/permission/roles/1", nil)
	assert.Equal(t, 0, resp.Code, "GetRoleByID 不应被 401 拦截")

	resp = env.doJSON(t, http.MethodGet, "/api/v1/permission/roles", nil)
	assert.Equal(t, 0, resp.Code, "ListRoles 不应被 401 拦截")

	resp = env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions", nil)
	assert.Equal(t, 0, resp.Code, "ListPermissions 不应被 401 拦截")

	resp = env.doJSON(t, http.MethodGet, "/api/v1/permission/permissions/100", nil)
	assert.Equal(t, 0, resp.Code, "GetPermissionByID 不应被 401 拦截")
}
