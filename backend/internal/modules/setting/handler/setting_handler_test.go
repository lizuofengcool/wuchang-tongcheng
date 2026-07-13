// Package handler_test 系统设置模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock SettingService，覆盖 handler 全部分支：
//   - 未登录拦截（写操作 userID=0 → 401：Create/Update/Delete/BatchUpdate）
//   - URL :id 参数解析失败（非数字 → 400 "无效的配置ID"）
//   - 请求体 Bind 失败（非法 JSON → 400 "参数错误"）
//   - service 成功/错误透传（业务码 + message + data 透传）
//   - regionID 注入（Create/GetByGroup/GetAll/BatchUpdate 从上下文读取 regionID，缺失兜底 DefaultRegionID=2）
//   - 公开读取无需登录（GetByID/GetByGroup/GetAll 在 handler 层不校验 userID，
//     鉴权由 plugin.go 的 RequirePermission 中间件负责，handler 层仅写操作做防御性登录校验）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 region_handler_test.go 同风格，区别在于 setting 接收 regionID 入参（地区数据隔离），
// 且新增 BatchUpdate 批量更新接口、GetByGroup 按分组返回扁平列表（无树结构）。
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
	"wuchang-tongcheng/internal/modules/setting/dto"
	setHandler "wuchang-tongcheng/internal/modules/setting/handler"
	"wuchang-tongcheng/internal/modules/setting/service"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockSettingService 内存 mock，实现 service.SettingService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockSettingService struct {
	// 调用记录
	lastCreateRegionID      uint
	lastCreateReq           *dto.CreateSettingRequest
	lastUpdateID            uint
	lastUpdateReq           *dto.UpdateSettingRequest
	lastDeleteID            uint
	lastGetByID             uint
	lastGetByGroupGroup     string
	lastGetByGroupRegionID  uint
	lastGetAllRegionID      uint
	lastBatchUpdateRegionID uint
	lastBatchUpdateReq      *dto.BatchUpdateRequest

	// 返回值预设
	createResult       *dto.SettingInfo
	createErr          error
	updateErr          error
	deleteErr          error
	getByIDResult      *dto.SettingInfo
	getByIDErr         error
	getByGroupResult   []dto.SettingInfo
	getByGroupErr      error
	getAllResult       map[string][]dto.SettingInfo
	getAllErr          error
	batchUpdateErr     error
}

func (m *mockSettingService) Create(regionID uint, req *dto.CreateSettingRequest) (*dto.SettingInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateReq = req
	return m.createResult, m.createErr
}
func (m *mockSettingService) Update(id uint, req *dto.UpdateSettingRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockSettingService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockSettingService) GetByID(id uint) (*dto.SettingInfo, error) {
	m.lastGetByID = id
	return m.getByIDResult, m.getByIDErr
}
func (m *mockSettingService) GetByGroup(group string, regionID uint) ([]dto.SettingInfo, error) {
	m.lastGetByGroupGroup = group
	m.lastGetByGroupRegionID = regionID
	return m.getByGroupResult, m.getByGroupErr
}
func (m *mockSettingService) GetAll(regionID uint) (map[string][]dto.SettingInfo, error) {
	m.lastGetAllRegionID = regionID
	return m.getAllResult, m.getAllErr
}
func (m *mockSettingService) BatchUpdate(regionID uint, req *dto.BatchUpdateRequest) error {
	m.lastBatchUpdateRegionID = regionID
	m.lastBatchUpdateReq = req
	return m.batchUpdateErr
}
// GetValue 供其他模块读取配置，handler 层不暴露，留空 stub 以满足接口
func (m *mockSettingService) GetValue(group, key string, regionID uint) (string, error) {
	return "", nil
}

// 确保 mockSettingService 实现 service.SettingService 接口
var _ service.SettingService = (*mockSettingService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockSettingService
}

// newHandlerEnv 构造 gin 引擎并注册 setting 路由（与 setting/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// ctxRegionID 用于模拟 Region 中间件注入的 region_id（0 表示未注入，handler 会兜底 DefaultRegionID=2）。
// 注意：setting handler 的读操作（GetByID/GetByGroup/GetAll）不校验 userID，
// 故 ctxUserID 仅影响写操作（Create/Update/Delete/BatchUpdate）的登录拦截分支。
func newHandlerEnv(t *testing.T, ctxUserID uint, ctxRegionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockSettingService{
		createResult: &dto.SettingInfo{
			ID: 1, Group: "site", Key: "site_name", Value: "五常同城",
			ParsedValue: "五常同城", ValueType: "string",
			Description: "站点名称", Sort: 10,
		},
		getByIDResult: &dto.SettingInfo{
			ID: 1, Group: "site", Key: "site_name", Value: "五常同城",
			ParsedValue: "五常同城", ValueType: "string",
			Description: "站点名称", Sort: 10,
		},
		getByGroupResult: []dto.SettingInfo{
			{ID: 1, Group: "site", Key: "site_name", Value: "五常同城", ParsedValue: "五常同城", ValueType: "string", Sort: 10},
			{ID: 2, Group: "site", Key: "site_url", Value: "https://wuchang.example.com", ParsedValue: "https://wuchang.example.com", ValueType: "string", Sort: 20},
		},
		getAllResult: map[string][]dto.SettingInfo{
			"site": {
				{ID: 1, Group: "site", Key: "site_name", Value: "五常同城", ParsedValue: "五常同城", ValueType: "string", Sort: 10},
			},
			"sms": {
				{ID: 3, Group: "sms", Key: "provider", Value: "aliyun", ParsedValue: "aliyun", ValueType: "string", Sort: 0},
			},
		},
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

	h := setHandler.NewHandler(mock)
	// 注册路由，路径与 setting/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/setting")
	root.GET("", h.GetAll)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/:id", h.GetByID)
	root.GET("/group/:group", h.GetByGroup)
	root.PUT("/batch", h.BatchUpdate)

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

// ---------- Create ----------

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5) // 登录用户 + regionID=5
	body := dto.CreateSettingRequest{
		Group: "site", Key: "site_name", Value: "五常同城",
		ValueType: "string", Description: "站点名称", Sort: 10,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/setting", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	// 透传 regionID + 请求体
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "site", env.mock.lastCreateReq.Group)
	assert.Equal(t, "site_name", env.mock.lastCreateReq.Key)
	assert.Equal(t, "五常同城", env.mock.lastCreateReq.Value)
	// data 透传 service 返回值
	var info dto.SettingInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "site_name", info.Key)
}

func TestHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	body := dto.CreateSettingRequest{Group: "site", Key: "site_name"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/setting", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/setting", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.createErr = errors.New("配置键已存在")
	env.mock.createResult = nil
	body := dto.CreateSettingRequest{Group: "site", Key: "site_name"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/setting", body)

	// 业务码 CodeAlreadyExists=1007 + err.Error() 透传
	assert.Equal(t, 1007, resp.Code)
	assert.Equal(t, "配置键已存在", resp.Message)
}

// ---------- Update ----------

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.UpdateSettingRequest{Value: "新名称", Description: "新描述", Sort: 20}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "新名称", env.mock.lastUpdateReq.Value)
	assert.Equal(t, "新描述", env.mock.lastUpdateReq.Description)
	assert.Equal(t, 20, env.mock.lastUpdateReq.Sort)
}

func TestHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/5", dto.UpdateSettingRequest{Value: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/abc", dto.UpdateSettingRequest{Value: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的配置ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/setting/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updateErr = errors.New("配置项不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/5", dto.UpdateSettingRequest{Value: "x"})

	// 业务码 CodeSystemError=1001 + err.Error() 透传
	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "配置项不存在", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/setting/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteID)
}

func TestHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/setting/7", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/setting/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的配置ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.deleteErr = errors.New("配置项不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/setting/7", nil)

	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "配置项不存在", resp.Message)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetByID)
	var info dto.SettingInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "site_name", info.Key)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的配置ID", resp.Message)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.getByIDErr = errors.New("配置项不存在")
	env.mock.getByIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/999", nil)

	// 业务码 CodeNotFound=1006 + err.Error() 透传
	assert.Equal(t, 1006, resp.Code)
	assert.Equal(t, "配置项不存在", resp.Message)
}

// ---------- GetByGroup ----------

func TestHandler_GetByGroup_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5) // regionID=5
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/group/site", nil)

	assert.Equal(t, 0, resp.Code)
	// 透传 group + regionID
	assert.Equal(t, "site", env.mock.lastGetByGroupGroup)
	assert.Equal(t, uint(5), env.mock.lastGetByGroupRegionID)
	var list []dto.SettingInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "site_name", list[0].Key)
	assert.Equal(t, "site_url", list[1].Key)
}

func TestHandler_GetByGroup_DifferentGroup(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/group/sms", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "sms", env.mock.lastGetByGroupGroup)
}

func TestHandler_GetByGroup_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.getByGroupErr = errors.New("db down")
	env.mock.getByGroupResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting/group/site", nil)

	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetAll ----------

func TestHandler_GetAll_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting", nil)

	assert.Equal(t, 0, resp.Code)
	// 透传 regionID
	assert.Equal(t, uint(5), env.mock.lastGetAllRegionID)
	// data 是 map[string][]SettingInfo
	var m map[string][]dto.SettingInfo
	require.NoError(t, json.Unmarshal(resp.Data, &m))
	assert.Contains(t, m, "site")
	assert.Contains(t, m, "sms")
	assert.Len(t, m["site"], 1)
	assert.Len(t, m["sms"], 1)
}

func TestHandler_GetAll_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.getAllErr = errors.New("boom")
	env.mock.getAllResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting", nil)

	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "boom", resp.Message)
}

// ---------- BatchUpdate ----------

func TestHandler_BatchUpdate_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.BatchUpdateRequest{
		Items: []dto.BatchItem{
			{Key: "site_name", Value: "新名称"},
			{Key: "site_url", Value: "https://new.example.com"},
		},
	}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/batch", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量更新成功", resp.Message)
	// 透传 regionID + items
	assert.Equal(t, uint(5), env.mock.lastBatchUpdateRegionID)
	require.NotNil(t, env.mock.lastBatchUpdateReq)
	require.Len(t, env.mock.lastBatchUpdateReq.Items, 2)
	assert.Equal(t, "site_name", env.mock.lastBatchUpdateReq.Items[0].Key)
	assert.Equal(t, "新名称", env.mock.lastBatchUpdateReq.Items[0].Value)
	assert.Equal(t, "site_url", env.mock.lastBatchUpdateReq.Items[1].Key)
}

func TestHandler_BatchUpdate_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	body := dto.BatchUpdateRequest{Items: []dto.BatchItem{{Key: "k", Value: "v"}}}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/batch", body)

	assert.Equal(t, 401, resp.Code)
	assert.Nil(t, env.mock.lastBatchUpdateReq)
}

func TestHandler_BatchUpdate_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/setting/batch", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastBatchUpdateReq)
}

func TestHandler_BatchUpdate_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchUpdateErr = errors.New("配置值与值类型不匹配")
	body := dto.BatchUpdateRequest{Items: []dto.BatchItem{{Key: "k", Value: "v"}}}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/setting/batch", body)

	assert.Equal(t, 1001, resp.Code)
	assert.Equal(t, "配置值与值类型不匹配", resp.Message)
}

// ---------- regionID 注入与默认兜底 ----------

// TestHandler_RegionID_DefaultFallback 验证未注入 regionID 时 handler 兜底为 DefaultRegionID=2。
// 测试所有接收 regionID 的 handler（Create/GetByGroup/GetAll/BatchUpdate）的兜底路径。
func TestHandler_RegionID_DefaultFallback(t *testing.T) {
	env := newHandlerEnv(t, 1, 0) // 登录但不注入 regionID

	// Create
	resp := env.doJSON(t, http.MethodPost, "/api/v1/setting", dto.CreateSettingRequest{Group: "g", Key: "k"})
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(middleware.DefaultRegionID), env.mock.lastCreateRegionID)

	// GetByGroup
	resp = env.doJSON(t, http.MethodGet, "/api/v1/setting/group/site", nil)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(middleware.DefaultRegionID), env.mock.lastGetByGroupRegionID)

	// GetAll
	resp = env.doJSON(t, http.MethodGet, "/api/v1/setting", nil)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(middleware.DefaultRegionID), env.mock.lastGetAllRegionID)

	// BatchUpdate
	resp = env.doJSON(t, http.MethodPut, "/api/v1/setting/batch", dto.BatchUpdateRequest{Items: []dto.BatchItem{{Key: "k", Value: "v"}}})
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(middleware.DefaultRegionID), env.mock.lastBatchUpdateRegionID)
}

// ---------- 公开读取：无需登录 ----------

// TestHandler_PublicRead_NoAuthRequired 验证 setting 的读操作在 handler 层不校验 userID，
// 即使未登录（userID=0）也能正常返回数据。鉴权由 plugin.go 的 RequirePermission 中间件负责，
// handler 层仅写操作做防御性登录校验。
func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	// 三个读路径均不应被 401 拦截
	resp := env.doJSON(t, http.MethodGet, "/api/v1/setting", nil)
	assert.Equal(t, 0, resp.Code)

	resp = env.doJSON(t, http.MethodGet, "/api/v1/setting/1", nil)
	assert.Equal(t, 0, resp.Code)

	resp = env.doJSON(t, http.MethodGet, "/api/v1/setting/group/site", nil)
	assert.Equal(t, 0, resp.Code)
}
