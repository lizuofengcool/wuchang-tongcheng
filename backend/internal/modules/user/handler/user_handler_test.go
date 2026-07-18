// Package handler_test 用户模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock UserService，覆盖 handler 全部分支：
//   - 公开接口（Register/Login/SendSMSCode/SMSLogin/OAuthLogin）：Bind 失败 + service 错误码路由
//   - 需登录接口（GetUserInfo/UpdateProfile/ChangePassword）：userID=0 → 401 拦截 + Bind 失败 + service 错误码路由
//   - 管理后台接口（ListUsers/AdminGetUser/AdminCreateUser/AdminUpdateUser/UpdateUserStatus/ResetPassword/DeleteUser）：
//     userID=0 → 401 拦截 + URL :id 解析失败 + Bind 失败 + service 错误码路由
//   - regionID 注入（Register/SMSLogin/OAuthLogin/AdminCreateUser/ListUsers 从上下文读取 regionID）
//   - provider 路径参数透传（OAuthLogin）
//   - sentinel 错误码路由：
//     ErrUserAlreadyExists → 2003 / ErrPasswordInvalid → 2004 / ErrUserNotFound → 2002 /
//     ErrSMSNotConfigured|ErrSMSCodeInvalid → 4004 / ErrOAuthNotConfigured → 4006 / 其他 → 2001
//
// 不依赖 DB/Redis/Docker/真实短信或 OAuth 服务，纯内存 mock service 验证 handler 装配层逻辑。
// 与 file/setting/region/category/news/permission handler 测试同风格。
// user 模块原有 user_e2e_test.go 是真实 service + seed 的端到端测试（依赖 DB），
// 本测试是隔离的装配层单测，二者互补：e2e 验证全链路集成，本测试聚焦 handler 参数解析/错误码路由/中间件协作。
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/user/dto"
	userHandler "wuchang-tongcheng/internal/modules/user/handler"
	"wuchang-tongcheng/internal/modules/user/service"
	"wuchang-tongcheng/internal/pkg/utils"
)

// userAPIResponse 解析统一响应体 {code, message, data}
// 注意：与 user_e2e_test.go 的 apiResponse 同构但独立命名，避免同包类型重声明冲突。
type userAPIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockUserService 内存 mock，实现 service.UserService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockUserService struct {
	// 调用记录
	lastRegisterRegionID uint
	lastRegisterReq      *dto.RegisterRequest
	lastLoginReq         *dto.LoginRequest
	lastSMSPhone         string
	lastSMSLoginRegionID uint
	lastSMSLoginPhone    string
	lastSMSLoginCode     string
	lastOAuthCtxRegionID uint
	lastOAuthProvider    string
	lastOAuthCode        string
	lastGetUserID        uint
	lastUpdateUserID     uint
	lastUpdateReq        *dto.UpdateProfileRequest
	lastChangePwdID      uint
	lastChangePwdReq     *dto.ChangePasswordRequest
	lastListRegionID     uint
	lastListReq          *dto.ListUsersRequest
	lastAdminGetID       uint
	lastAdminCreateRegID uint
	lastAdminCreateReq   *dto.AdminCreateUserRequest
	lastAdminUpdateID    uint
	lastAdminUpdateReq   *dto.AdminUpdateUserRequest
	lastStatusID         uint
	lastStatusStatus     int
	lastResetID          uint
	lastResetReq         *dto.ResetPasswordRequest
	lastDeleteID         uint

	// 返回值预设
	registerResult   *dto.UserInfo
	registerErr      error
	loginResult      *dto.LoginResponse
	loginErr         error
	smsResult        *dto.SendSMSCodeResponse
	smsErr           error
	smsLoginResult   *dto.LoginResponse
	smsLoginErr      error
	oauthResult      *dto.LoginResponse
	oauthErr         error
	getUserResult    *dto.UserInfo
	getUserErr       error
	updateErr        error
	changePwdErr     error
	listPagination   *utils.Pagination
	listResult       []dto.UserInfo
	listErr          error
	adminCreateResult *dto.UserInfo
	adminCreateErr   error
	adminUpdateErr   error
	statusErr        error
	resetErr         error
	deleteErr        error
}

func (m *mockUserService) Register(regionID uint, req *dto.RegisterRequest) (*dto.UserInfo, error) {
	m.lastRegisterRegionID = regionID
	m.lastRegisterReq = req
	return m.registerResult, m.registerErr
}

func (m *mockUserService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	m.lastLoginReq = req
	return m.loginResult, m.loginErr
}

func (m *mockUserService) SendSMSCode(ctx context.Context, phone string) (*dto.SendSMSCodeResponse, error) {
	_ = ctx // handler 内部用 context.WithTimeout 创建新 ctx，此处仅占位不校验
	m.lastSMSPhone = phone
	return m.smsResult, m.smsErr
}

func (m *mockUserService) LoginBySMS(ctx context.Context, regionID uint, phone, code string) (*dto.LoginResponse, error) {
	_ = ctx
	m.lastSMSLoginRegionID = regionID
	m.lastSMSLoginPhone = phone
	m.lastSMSLoginCode = code
	return m.smsLoginResult, m.smsLoginErr
}

func (m *mockUserService) OAuthLogin(ctx context.Context, regionID uint, provider, code string) (*dto.LoginResponse, error) {
	_ = ctx
	m.lastOAuthCtxRegionID = regionID
	m.lastOAuthProvider = provider
	m.lastOAuthCode = code
	return m.oauthResult, m.oauthErr
}

func (m *mockUserService) GetUserInfo(userID uint) (*dto.UserInfo, error) {
	m.lastGetUserID = userID
	return m.getUserResult, m.getUserErr
}

func (m *mockUserService) UpdateProfile(userID uint, req *dto.UpdateProfileRequest) error {
	m.lastUpdateUserID = userID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockUserService) ChangePassword(userID uint, req *dto.ChangePasswordRequest) error {
	m.lastChangePwdID = userID
	m.lastChangePwdReq = req
	return m.changePwdErr
}

func (m *mockUserService) ListUsers(regionID uint, req *dto.ListUsersRequest) (*utils.Pagination, []dto.UserInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	return m.listPagination, m.listResult, m.listErr
}

func (m *mockUserService) AdminCreateUser(regionID uint, req *dto.AdminCreateUserRequest) (*dto.UserInfo, error) {
	m.lastAdminCreateRegID = regionID
	m.lastAdminCreateReq = req
	return m.adminCreateResult, m.adminCreateErr
}

func (m *mockUserService) AdminUpdateUser(id uint, req *dto.AdminUpdateUserRequest) error {
	m.lastAdminUpdateID = id
	m.lastAdminUpdateReq = req
	return m.adminUpdateErr
}

func (m *mockUserService) UpdateUserStatus(id uint, status int) error {
	m.lastStatusID = id
	m.lastStatusStatus = status
	return m.statusErr
}

func (m *mockUserService) ResetPassword(id uint, req *dto.ResetPasswordRequest) error {
	m.lastResetID = id
	m.lastResetReq = req
	return m.resetErr
}

func (m *mockUserService) DeleteUser(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

// 确保 mockUserService 实现 service.UserService 接口
var _ service.UserService = (*mockUserService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockUserService
}

// newHandlerEnv 构造 gin 引擎并注册 user 路由（与 user/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// ctxRegionID 用于模拟 Region 中间件注入的 region_id（0 表示未注入，handler 兜底 DefaultRegionID）。
// 注意：handler 层不挂权限中间件，权限校验由 plugin.go 的 RequirePermission 负责，
// 此处仅测 handler 装配层逻辑（登录拦截/参数解析/Bind/错误码路由/regionID 注入）。
func newHandlerEnv(t *testing.T, ctxUserID uint, ctxRegionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// 预设成功返回值
	userInfo := &dto.UserInfo{
		ID:       1,
		Username: "testuser",
		Nickname: "测试用户",
		Avatar:   "https://cdn.example.com/avatar/1.png",
		Phone:    "13800138000",
		Email:    "test@example.com",
		Gender:   1,
		Status:   1,
	}
	loginResp := &dto.LoginResponse{
		Token:    "jwt.token.example",
		Expires:  86400,
		UserInfo: *userInfo,
	}

	mock := &mockUserService{
		registerResult:    userInfo,
		loginResult:       loginResp,
		smsResult:         &dto.SendSMSCodeResponse{DevCode: "123456"},
		smsLoginResult:    loginResp,
		oauthResult:       loginResp,
		getUserResult:     userInfo,
		adminCreateResult: userInfo,
		listPagination:    &utils.Pagination{Page: 1, PageSize: 10, Total: 1},
		listResult:        []dto.UserInfo{*userInfo},
	}

	r := coreRouter.NewRouter()
	// 模拟 Auth + Region 中间件：注入 user_id 与 region_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, ctxUserID)
		if ctxRegionID > 0 {
			c.Set(middleware.RegionIDKey, ctxRegionID)
		}
		c.Next()
	})

	h := userHandler.NewHandler(mock)
	// 注册路由，路径与 user/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/user")
	root.POST("/register", h.Register)
	root.POST("/login", h.Login)
	root.POST("/sms/code", h.SendSMSCode)
	root.POST("/login/sms", h.SMSLogin)
	root.POST("/login/oauth/:provider", h.OAuthLogin)
	root.GET("/info", h.GetUserInfo)
	root.PUT("/profile", h.UpdateProfile)
	root.PUT("/password", h.ChangePassword)
	admin := root.Group("/admin")
	admin.GET("/users", h.ListUsers)
	admin.POST("/users", h.AdminCreateUser)
	admin.GET("/users/:id", h.AdminGetUser)
	admin.PUT("/users/:id", h.AdminUpdateUser)
	admin.PUT("/users/:id/status", h.UpdateUserStatus)
	admin.PUT("/users/:id/password", h.ResetPassword)
	admin.DELETE("/users/:id", h.DeleteUser)

	return &handlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。body 为 nil 时发送空 body。
func (e *handlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *userAPIResponse {
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

	var resp userAPIResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doRaw 发起原始请求（用于测试 Bind 失败：非法 JSON body）。
func (e *handlerEnv) doRaw(t *testing.T, method, path string, rawBody string, contentType string) *userAPIResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp userAPIResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// doQuery 发起 GET 请求带 query 参数（用于 ListUsers）。
func (e *handlerEnv) doQuery(t *testing.T, path string) *userAPIResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var resp userAPIResponse
	if w.Body.Len() > 0 {
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	}
	return &resp
}

// ---------- Register ----------

func TestHandler_Register_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口，userID 无影响
	body := dto.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Nickname: "新用户",
		Phone:    "13900139000",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/register", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "注册成功", resp.Message)
	// 透传 regionID
	assert.Equal(t, uint(5), env.mock.lastRegisterRegionID)
	// 透传请求体
	require.NotNil(t, env.mock.lastRegisterReq)
	assert.Equal(t, "newuser", env.mock.lastRegisterReq.Username)
	assert.Equal(t, "新用户", env.mock.lastRegisterReq.Nickname)
	// data 透传 service 返回值
	var info dto.UserInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "testuser", info.Username)
}

func TestHandler_Register_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/register", "{not json", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	// service 不应被调用
	assert.Nil(t, env.mock.lastRegisterReq)
}

func TestHandler_Register_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.registerErr = service.ErrUserAlreadyExists
	body := dto.RegisterRequest{Username: "dup", Password: "password123"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/register", body)

	// ErrUserAlreadyExists → CodeUserAlreadyExists 2003
	assert.Equal(t, utils.CodeUserAlreadyExists, resp.Code)
	assert.Equal(t, service.ErrUserAlreadyExists.Error(), resp.Message)
}

// ---------- Login ----------

func TestHandler_Login_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	body := dto.LoginRequest{Username: "testuser", Password: "password123"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "登录成功", resp.Message)
	require.NotNil(t, env.mock.lastLoginReq)
	assert.Equal(t, "testuser", env.mock.lastLoginReq.Username)
	// data 透传 LoginResponse
	var loginResp dto.LoginResponse
	require.NoError(t, json.Unmarshal(resp.Data, &loginResp))
	assert.Equal(t, "jwt.token.example", loginResp.Token)
	assert.Equal(t, 86400, loginResp.Expires)
	assert.Equal(t, "testuser", loginResp.UserInfo.Username)
}

func TestHandler_Login_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/login", "{bad json", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastLoginReq)
}

func TestHandler_Login_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	env.mock.loginErr = service.ErrPasswordInvalid
	body := dto.LoginRequest{Username: "testuser", Password: "wrong"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login", body)

	// ErrPasswordInvalid → CodeUserPasswordError 2004
	assert.Equal(t, utils.CodeUserPasswordError, resp.Code)
	assert.Equal(t, service.ErrPasswordInvalid.Error(), resp.Message)
}

// ---------- SendSMSCode ----------

func TestHandler_SendSMSCode_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	body := dto.SendSMSCodeRequest{Phone: "13800138000"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/sms/code", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "验证码已发送", resp.Message)
	assert.Equal(t, "13800138000", env.mock.lastSMSPhone)
	// data 透传 DevCode
	var data dto.SendSMSCodeResponse
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, "123456", data.DevCode)
}

func TestHandler_SendSMSCode_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/sms/code", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Empty(t, env.mock.lastSMSPhone)
}

func TestHandler_SendSMSCode_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	env.mock.smsErr = service.ErrSMSNotConfigured
	body := dto.SendSMSCodeRequest{Phone: "13800138000"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/sms/code", body)

	// ErrSMSNotConfigured → CodeSMSError 4004
	assert.Equal(t, utils.CodeSMSError, resp.Code)
	assert.Equal(t, service.ErrSMSNotConfigured.Error(), resp.Message)
}

// ---------- SMSLogin ----------

func TestHandler_SMSLogin_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 7) // regionID=7
	body := dto.SMSLoginRequest{Phone: "13800138000", Code: "123456"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login/sms", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "登录成功", resp.Message)
	// regionID 注入
	assert.Equal(t, uint(7), env.mock.lastSMSLoginRegionID)
	assert.Equal(t, "13800138000", env.mock.lastSMSLoginPhone)
	assert.Equal(t, "123456", env.mock.lastSMSLoginCode)
}

func TestHandler_SMSLogin_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 7)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/login/sms", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Empty(t, env.mock.lastSMSLoginPhone)
}

func TestHandler_SMSLogin_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 7)
	env.mock.smsLoginErr = service.ErrSMSCodeInvalid
	body := dto.SMSLoginRequest{Phone: "13800138000", Code: "wrong"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login/sms", body)

	// ErrSMSCodeInvalid → CodeSMSError 4004
	assert.Equal(t, utils.CodeSMSError, resp.Code)
	assert.Equal(t, service.ErrSMSCodeInvalid.Error(), resp.Message)
}

// ---------- OAuthLogin ----------

func TestHandler_OAuthLogin_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 9) // regionID=9
	body := dto.OAuthLoginRequest{Code: "wechat_auth_code_xxx"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login/oauth/wechat", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "登录成功", resp.Message)
	// regionID + provider + code 透传
	assert.Equal(t, uint(9), env.mock.lastOAuthCtxRegionID)
	assert.Equal(t, "wechat", env.mock.lastOAuthProvider)
	assert.Equal(t, "wechat_auth_code_xxx", env.mock.lastOAuthCode)
}

func TestHandler_OAuthLogin_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 9)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/login/oauth/wechat", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Empty(t, env.mock.lastOAuthCode)
}

func TestHandler_OAuthLogin_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 9)
	env.mock.oauthErr = service.ErrOAuthNotConfigured
	body := dto.OAuthLoginRequest{Code: "xxx"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/login/oauth/wechat", body)

	// ErrOAuthNotConfigured → CodeOAuthError 4006
	assert.Equal(t, utils.CodeOAuthError, resp.Code)
	assert.Equal(t, service.ErrOAuthNotConfigured.Error(), resp.Message)
}

// ---------- GetUserInfo ----------

func TestHandler_GetUserInfo_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0) // 登录用户 userID=1
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/info", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastGetUserID)
	var info dto.UserInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "testuser", info.Username)
}

func TestHandler_GetUserInfo_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0) // 未登录
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/info", nil)

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetUserID)
}

func TestHandler_GetUserInfo_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.getUserErr = service.ErrUserNotFound
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/info", nil)

	// ErrUserNotFound → CodeUserNotFound 2002
	assert.Equal(t, utils.CodeUserNotFound, resp.Code)
	assert.Equal(t, service.ErrUserNotFound.Error(), resp.Message)
}

// ---------- UpdateProfile ----------

func TestHandler_UpdateProfile_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	body := dto.UpdateProfileRequest{Nickname: "新昵称", Gender: 2}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/profile", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastUpdateUserID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "新昵称", env.mock.lastUpdateReq.Nickname)
	assert.Equal(t, 2, env.mock.lastUpdateReq.Gender)
}

func TestHandler_UpdateProfile_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/profile", dto.UpdateProfileRequest{Nickname: "x"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateUserID)
}

func TestHandler_UpdateProfile_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/user/profile", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateUserID) // Bind 失败前未记录
}

func TestHandler_UpdateProfile_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.updateErr = errors.New("昵称冲突")
	body := dto.UpdateProfileRequest{Nickname: "x"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/profile", body)

	// 其他错误 → CodeUserError 2001
	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "昵称冲突", resp.Message)
}

// ---------- ChangePassword ----------

func TestHandler_ChangePassword_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	body := dto.ChangePasswordRequest{OldPassword: "old123", NewPassword: "new123456"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/password", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "密码修改成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastChangePwdID)
	require.NotNil(t, env.mock.lastChangePwdReq)
	assert.Equal(t, "old123", env.mock.lastChangePwdReq.OldPassword)
	assert.Equal(t, "new123456", env.mock.lastChangePwdReq.NewPassword)
}

func TestHandler_ChangePassword_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/password", dto.ChangePasswordRequest{OldPassword: "x", NewPassword: "y"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastChangePwdID)
}

func TestHandler_ChangePassword_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/user/password", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastChangePwdID)
}

func TestHandler_ChangePassword_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.changePwdErr = service.ErrOldPasswordWrong
	body := dto.ChangePasswordRequest{OldPassword: "wrong", NewPassword: "new123456"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/password", body)

	// ErrOldPasswordWrong → CodeUserPasswordError 2004
	assert.Equal(t, utils.CodeUserPasswordError, resp.Code)
	assert.Equal(t, service.ErrOldPasswordWrong.Error(), resp.Message)
}

// ---------- ListUsers ----------

func TestHandler_ListUsers_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5) // 登录 + regionID=5
	resp := env.doQuery(t, "/api/v1/user/admin/users?page=1&page_size=10&keyword=test&status=1")

	assert.Equal(t, 0, resp.Code)
	// regionID 注入
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	assert.Equal(t, 10, env.mock.lastListReq.PageSize)
	assert.Equal(t, "test", env.mock.lastListReq.Keyword)
	assert.Equal(t, 1, env.mock.lastListReq.Status)
	// data 透传 PageResult（扁平结构：list/total/page/pageSize，与 response.PageResult 一致）
	var data struct {
		List     []dto.UserInfo `json:"list"`
		Total    int64          `json:"total"`
		Page     int            `json:"page"`
		PageSize int            `json:"pageSize"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Len(t, data.List, 1)
	assert.Equal(t, int64(1), data.Total)
}

func TestHandler_ListUsers_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, "/api/v1/user/admin/users")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastListReq)
}

func TestHandler_ListUsers_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listErr = errors.New("DB 查询失败")
	resp := env.doQuery(t, "/api/v1/user/admin/users")

	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "DB 查询失败", resp.Message)
}

// ---------- AdminGetUser ----------

func TestHandler_AdminGetUser_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/admin/users/42", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(42), env.mock.lastGetUserID) // 复用 GetUserInfo service 方法
	var info dto.UserInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "testuser", info.Username)
}

func TestHandler_AdminGetUser_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/admin/users/42", nil)

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetUserID)
}

func TestHandler_AdminGetUser_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/admin/users/abc", nil)

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetUserID)
}

func TestHandler_AdminGetUser_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.getUserErr = service.ErrUserNotFound
	resp := env.doJSON(t, http.MethodGet, "/api/v1/user/admin/users/999", nil)

	assert.Equal(t, utils.CodeUserNotFound, resp.Code)
	assert.Equal(t, service.ErrUserNotFound.Error(), resp.Message)
}

// ---------- AdminCreateUser ----------

func TestHandler_AdminCreateUser_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5) // regionID=5
	body := dto.AdminCreateUserRequest{
		Username: "newadmin",
		Password: "password123",
		Nickname: "管理员新建",
		Phone:    "13700137000",
		Email:    "new@example.com",
		Gender:   1,
		Status:   1,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/admin/users", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	// regionID 注入
	assert.Equal(t, uint(5), env.mock.lastAdminCreateRegID)
	require.NotNil(t, env.mock.lastAdminCreateReq)
	assert.Equal(t, "newadmin", env.mock.lastAdminCreateReq.Username)
	assert.Equal(t, "管理员新建", env.mock.lastAdminCreateReq.Nickname)
}

func TestHandler_AdminCreateUser_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/admin/users", dto.AdminCreateUserRequest{
		Username: "x", Password: "password123",
	})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastAdminCreateReq)
}

func TestHandler_AdminCreateUser_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/user/admin/users", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastAdminCreateReq)
}

func TestHandler_AdminCreateUser_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminCreateErr = service.ErrUserAlreadyExists
	body := dto.AdminCreateUserRequest{Username: "dup", Password: "password123"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/user/admin/users", body)

	// ErrUserAlreadyExists → CodeUserAlreadyExists 2003
	assert.Equal(t, utils.CodeUserAlreadyExists, resp.Code)
	assert.Equal(t, service.ErrUserAlreadyExists.Error(), resp.Message)
}

// ---------- AdminUpdateUser ----------

func TestHandler_AdminUpdateUser_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	body := dto.AdminUpdateUserRequest{Nickname: "改后昵称"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/7", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastAdminUpdateID)
	require.NotNil(t, env.mock.lastAdminUpdateReq)
	assert.Equal(t, "改后昵称", env.mock.lastAdminUpdateReq.Nickname)
}

func TestHandler_AdminUpdateUser_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/7", dto.AdminUpdateUserRequest{Nickname: "x"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAdminUpdateID)
}

func TestHandler_AdminUpdateUser_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/xyz", dto.AdminUpdateUserRequest{Nickname: "x"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAdminUpdateID)
}

func TestHandler_AdminUpdateUser_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/user/admin/users/7", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAdminUpdateID)
}

func TestHandler_AdminUpdateUser_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.adminUpdateErr = errors.New("更新失败")
	body := dto.AdminUpdateUserRequest{Nickname: "x"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/7", body)

	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "更新失败", resp.Message)
}

// ---------- UpdateUserStatus ----------

func TestHandler_UpdateUserStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	body := dto.UpdateUserStatusRequest{Status: 0}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/3/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastStatusID)
	assert.Equal(t, 0, env.mock.lastStatusStatus)
}

func TestHandler_UpdateUserStatus_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/3/status", dto.UpdateUserStatusRequest{Status: 0})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastStatusID)
}

func TestHandler_UpdateUserStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/abc/status", dto.UpdateUserStatusRequest{Status: 0})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastStatusID)
}

func TestHandler_UpdateUserStatus_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/user/admin/users/3/status", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastStatusID)
}

func TestHandler_UpdateUserStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.statusErr = errors.New("状态切换失败")
	body := dto.UpdateUserStatusRequest{Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/3/status", body)

	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "状态切换失败", resp.Message)
}

// ---------- ResetPassword ----------

func TestHandler_ResetPassword_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	body := dto.ResetPasswordRequest{NewPassword: "newpassword123"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/8/password", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "密码重置成功", resp.Message)
	assert.Equal(t, uint(8), env.mock.lastResetID)
	require.NotNil(t, env.mock.lastResetReq)
	assert.Equal(t, "newpassword123", env.mock.lastResetReq.NewPassword)
}

func TestHandler_ResetPassword_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/8/password", dto.ResetPasswordRequest{NewPassword: "newpassword123"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastResetID)
}

func TestHandler_ResetPassword_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/xyz/password", dto.ResetPasswordRequest{NewPassword: "newpassword123"})

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastResetID)
}

func TestHandler_ResetPassword_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/user/admin/users/8/password", "{bad", "application/json")

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastResetID)
}

func TestHandler_ResetPassword_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.resetErr = errors.New("重置失败")
	body := dto.ResetPasswordRequest{NewPassword: "newpassword123"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/user/admin/users/8/password", body)

	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "重置失败", resp.Message)
}

// ---------- DeleteUser ----------

func TestHandler_DeleteUser_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/user/admin/users/9", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastDeleteID)
}

func TestHandler_DeleteUser_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/user/admin/users/9", nil)

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_DeleteUser_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/user/admin/users/abc", nil)

	assert.NotEqual(t, 0, resp.Code)
	assert.Equal(t, "无效的用户ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_DeleteUser_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 0)
	env.mock.deleteErr = errors.New("删除失败")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/user/admin/users/9", nil)

	assert.Equal(t, utils.CodeUserError, resp.Code)
	assert.Equal(t, "删除失败", resp.Message)
}
