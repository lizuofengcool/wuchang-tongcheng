// Package handler_test 地区模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock RegionService，覆盖 handler 全部分支：
//   - 未登录拦截（写操作 userID=0 → 401）
//   - URL :id 参数解析失败（非数字 → 400）
//   - 请求体 Bind 失败（非法 JSON → 400）
//   - service 成功/错误透传（业务码 + message + data 透传）
//   - query 参数兜底（parent_id 缺失/非数字默认 0）
//   - 公开读取无需登录（GetAll/GetByID/GetByParentID/GetTree 在 handler 层不校验 userID，
//     鉴权由 plugin.go 的 RequirePermission 中间件负责，handler 层仅写操作做防御性登录校验）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 category_handler_test.go 同风格，区别在于 region 是地区维度本身，service 方法不接受 regionID。
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
	"wuchang-tongcheng/internal/modules/region/dto"
	regHandler "wuchang-tongcheng/internal/modules/region/handler"
	"wuchang-tongcheng/internal/modules/region/service"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockRegionService 内存 mock，实现 service.RegionService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockRegionService struct {
	// 调用记录
	lastCreateReq   *dto.CreateRegionRequest
	lastUpdateID    uint
	lastUpdateReq   *dto.UpdateRegionRequest
	lastDeleteID    uint
	lastGetByID     uint
	lastGetByParent uint

	// 返回值预设
	createResult       *dto.RegionInfo
	createErr          error
	updateErr          error
	deleteErr          error
	getByIDResult      *dto.RegionInfo
	getByIDErr         error
	getByParentResult  []dto.RegionInfo
	getByParentErr     error
	getAllResult       []dto.RegionInfo
	getAllErr          error
	getTreeResult      []dto.RegionTree
	getTreeErr         error
}

func (m *mockRegionService) Create(req *dto.CreateRegionRequest) (*dto.RegionInfo, error) {
	m.lastCreateReq = req
	return m.createResult, m.createErr
}
func (m *mockRegionService) Update(id uint, req *dto.UpdateRegionRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockRegionService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockRegionService) GetByID(id uint) (*dto.RegionInfo, error) {
	m.lastGetByID = id
	return m.getByIDResult, m.getByIDErr
}
func (m *mockRegionService) GetByParentID(parentID uint) ([]dto.RegionInfo, error) {
	m.lastGetByParent = parentID
	return m.getByParentResult, m.getByParentErr
}
func (m *mockRegionService) GetAll() ([]dto.RegionInfo, error) {
	return m.getAllResult, m.getAllErr
}
func (m *mockRegionService) GetTree() ([]dto.RegionTree, error) {
	return m.getTreeResult, m.getTreeErr
}

// 确保 mockRegionService 实现 service.RegionService 接口
var _ service.RegionService = (*mockRegionService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockRegionService
}

// newHandlerEnv 构造 gin 引擎并注册 region 路由（与 region/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// 注意：region handler 的读操作（GetAll/GetByID/GetByParentID/GetTree）不校验 userID，
// 故 ctxUserID 仅影响写操作（Create/Update/Delete）的登录拦截分支。
func newHandlerEnv(t *testing.T, ctxUserID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockRegionService{
		createResult:  &dto.RegionInfo{ID: 1, Name: "黑龙江省", Code: "HLJ", ParentID: 0, Level: 1, Sort: 10, Status: 1},
		getByIDResult: &dto.RegionInfo{ID: 1, Name: "黑龙江省", Code: "HLJ", ParentID: 0, Level: 1, Sort: 10, Status: 1},
		getByParentResult: []dto.RegionInfo{
			{ID: 2, Name: "哈尔滨市", Code: "HRB", ParentID: 1, Level: 2, Sort: 5, Status: 1},
		},
		getAllResult: []dto.RegionInfo{
			{ID: 1, Name: "黑龙江省", Code: "HLJ", ParentID: 0, Level: 1, Sort: 10, Status: 1},
		},
		getTreeResult: []dto.RegionTree{{
			RegionInfo: dto.RegionInfo{ID: 1, Name: "黑龙江省", Code: "HLJ", ParentID: 0, Level: 1, Sort: 10, Status: 1},
			Children: []dto.RegionTree{{
				RegionInfo: dto.RegionInfo{ID: 2, Name: "哈尔滨市", Code: "HRB", ParentID: 1, Level: 2, Sort: 5, Status: 1},
			}},
		}},
	}

	r := coreRouter.NewRouter()
	// 模拟 Auth 中间件：注入 user_id（region 不需要 region_id，故不注入）
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := regHandler.NewHandler(mock)
	// 注册路由，路径与 region/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/region")
	root.GET("", h.GetAll)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/:id", h.GetByID)
	root.GET("/children", h.GetByParentID)
	root.GET("/tree", h.GetTree)

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
	env := newHandlerEnv(t, 1)
	// Level 必须满足 dto.CreateRegionRequest 的 binding:"oneof=1 2 3"（handler 不使用 req.Level，
	// service 按 ParentID 重新计算，但绑定校验先于此执行，故需传入合法值）
	body := dto.CreateRegionRequest{Name: "黑龙江省", Code: "HLJ", ParentID: 0, Level: 1, Sort: 10, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/region", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	// 透传请求体
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "黑龙江省", env.mock.lastCreateReq.Name)
	assert.Equal(t, "HLJ", env.mock.lastCreateReq.Code)
	assert.Equal(t, uint(0), env.mock.lastCreateReq.ParentID)
	// data 透传 service 返回值
	var info dto.RegionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "黑龙江省", info.Name)
}

func TestHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0) // 未登录
	body := dto.CreateRegionRequest{Name: "黑龙江省", Code: "HLJ", Level: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/region", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/region", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.createErr = errors.New("地区编码已存在")
	env.mock.createResult = nil
	body := dto.CreateRegionRequest{Name: "黑龙江省", Code: "HLJ", Level: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/region", body)

	// 业务码 CodeRegionError=2101 + err.Error() 透传
	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "地区编码已存在", resp.Message)
}

// ---------- Update ----------

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	body := dto.UpdateRegionRequest{Name: "黑龙江", Sort: 20, Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/region/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "黑龙江", env.mock.lastUpdateReq.Name)
}

func TestHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/region/5", dto.UpdateRegionRequest{Name: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/region/abc", dto.UpdateRegionRequest{Name: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的地区ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/region/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.updateErr = errors.New("地区不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/region/5", dto.UpdateRegionRequest{Name: "x"})

	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "地区不存在", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/region/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteID)
}

func TestHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/region/7", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/region/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的地区ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.deleteErr = errors.New("该地区存在子地区，无法删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/region/7", nil)

	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "该地区存在子地区，无法删除", resp.Message)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetByID)
	var info dto.RegionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "黑龙江省", info.Name)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的地区ID", resp.Message)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.getByIDErr = errors.New("地区不存在")
	env.mock.getByIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/999", nil)

	// CodeRegionNotFound=2102
	assert.Equal(t, 2102, resp.Code)
	assert.Equal(t, "地区不存在", resp.Message)
}

// ---------- GetByParentID ----------

func TestHandler_GetByParentID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/children?parent_id=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastGetByParent)
	var list []dto.RegionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "哈尔滨市", list[0].Name)
}

func TestHandler_GetByParentID_DefaultParentZero(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// 不传 parent_id → 默认 0（取顶层省级行政区）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/children", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetByParent)
}

func TestHandler_GetByParentID_InvalidParentID(t *testing.T) {
	env := newHandlerEnv(t, 1)
	// 非数字 parent_id → ParseUint 失败兜底为 0（handler 不报错）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/children?parent_id=abc", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetByParent)
}

func TestHandler_GetByParentID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.getByParentErr = errors.New("db down")
	env.mock.getByParentResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/children?parent_id=1", nil)

	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetTree ----------

func TestHandler_GetTree_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/tree", nil)

	assert.Equal(t, 0, resp.Code)
	var tree []dto.RegionTree
	require.NoError(t, json.Unmarshal(resp.Data, &tree))
	require.Len(t, tree, 1)
	assert.Equal(t, "黑龙江省", tree[0].Name)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "哈尔滨市", tree[0].Children[0].Name)
}

func TestHandler_GetTree_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.getTreeErr = errors.New("redis boom")
	env.mock.getTreeResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region/tree", nil)

	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// ---------- GetAll ----------

func TestHandler_GetAll_Success(t *testing.T) {
	env := newHandlerEnv(t, 1)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region", nil)

	assert.Equal(t, 0, resp.Code)
	var list []dto.RegionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "黑龙江省", list[0].Name)
}

func TestHandler_GetAll_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1)
	env.mock.getAllErr = errors.New("boom")
	env.mock.getAllResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region", nil)

	assert.Equal(t, 2101, resp.Code)
	assert.Equal(t, "boom", resp.Message)
}

// ---------- 公开读取：无需登录 ----------

// TestHandler_PublicRead_NoAuthRequired 验证 region 的读操作在 handler 层不校验 userID，
// 即使未登录（userID=0）也能正常返回数据。鉴权由 plugin.go 的 RequirePermission 中间件负责，
// handler 层仅写操作做防御性登录校验。GetAll 路由在 plugin.go 中完全公开（无中间件）。
func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0) // 未登录
	// 四个读路径均不应被 401 拦截
	resp := env.doJSON(t, http.MethodGet, "/api/v1/region", nil)
	assert.Equal(t, 0, resp.Code)

	resp = env.doJSON(t, http.MethodGet, "/api/v1/region/1", nil)
	assert.Equal(t, 0, resp.Code)

	resp = env.doJSON(t, http.MethodGet, "/api/v1/region/children?parent_id=0", nil)
	assert.Equal(t, 0, resp.Code)

	resp = env.doJSON(t, http.MethodGet, "/api/v1/region/tree", nil)
	assert.Equal(t, 0, resp.Code)
}
