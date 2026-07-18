// Package handler_test 分类模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock CategoryService，覆盖 handler 全部分支：
//   - 未登录拦截（userID=0 → 401）
//   - URL :id 参数解析失败（非数字 → 400）
//   - 请求体 Bind 失败（非法 JSON → 400）
//   - service 成功/错误透传（业务码 + message + data 透传）
//   - 地区ID 上下文注入（regionID 透传给 service）
//   - query 参数兜底（parent_id 缺失默认 0）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
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
	"wuchang-tongcheng/internal/modules/category/dto"
	catHandler "wuchang-tongcheng/internal/modules/category/handler"
	"wuchang-tongcheng/internal/modules/category/service"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// mockCategoryService 内存 mock，实现 service.CategoryService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockCategoryService struct {
	// 调用记录
	lastCreateRegionID uint
	lastCreateReq      *dto.CreateCategoryRequest
	lastUpdateID       uint
	lastUpdateReq      *dto.UpdateCategoryRequest
	lastDeleteID       uint
	lastGetByID        uint
	lastGetByParent    struct {
		ParentID uint
		RegionID uint
	}
	lastGetAllRegionID uint
	lastGetTreeRegionID uint

	// 返回值预设
	createResult *dto.CategoryInfo
	createErr    error
	updateErr    error
	deleteErr    error
	getByIDResult *dto.CategoryInfo
	getByIDErr    error
	getByParentResult []dto.CategoryInfo
	getByParentErr    error
	getAllResult []dto.CategoryInfo
	getAllErr    error
	getTreeResult []dto.CategoryTree
	getTreeErr    error
}

func (m *mockCategoryService) Create(regionID uint, req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateReq = req
	return m.createResult, m.createErr
}
func (m *mockCategoryService) Update(id uint, req *dto.UpdateCategoryRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockCategoryService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockCategoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	m.lastGetByID = id
	return m.getByIDResult, m.getByIDErr
}
func (m *mockCategoryService) GetByParentID(parentID uint, regionID uint) ([]dto.CategoryInfo, error) {
	m.lastGetByParent.ParentID = parentID
	m.lastGetByParent.RegionID = regionID
	return m.getByParentResult, m.getByParentErr
}
func (m *mockCategoryService) GetAll(regionID uint) ([]dto.CategoryInfo, error) {
	m.lastGetAllRegionID = regionID
	return m.getAllResult, m.getAllErr
}
func (m *mockCategoryService) GetTree(regionID uint) ([]dto.CategoryTree, error) {
	m.lastGetTreeRegionID = regionID
	return m.getTreeResult, m.getTreeErr
}

// 确保 mockCategoryService 实现 service.CategoryService 接口
var _ service.CategoryService = (*mockCategoryService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockCategoryService
}

// newHandlerEnv 构造 gin 引擎并注册 category 路由（与 category/plugin.go RegisterRoutes 路径一致）。
// ctxUserID 用于模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 用于模拟 Region 中间件注入的 region_id。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockCategoryService{
		createResult:      &dto.CategoryInfo{ID: 1, Name: "二手物品", ParentID: 0, Level: 1, Sort: 10, Status: 1},
		getByIDResult:     &dto.CategoryInfo{ID: 1, Name: "二手物品", ParentID: 0, Level: 1, Sort: 10, Status: 1},
		getByParentResult: []dto.CategoryInfo{{ID: 2, Name: "手机", ParentID: 1, Level: 2, Sort: 5, Status: 1}},
		getAllResult:       []dto.CategoryInfo{{ID: 1, Name: "二手物品", ParentID: 0, Level: 1, Sort: 10, Status: 1}},
		getTreeResult: []dto.CategoryTree{{
			CategoryInfo: dto.CategoryInfo{ID: 1, Name: "二手物品", ParentID: 0, Level: 1, Sort: 10, Status: 1},
			Children: []dto.CategoryTree{{
				CategoryInfo: dto.CategoryInfo{ID: 2, Name: "手机", ParentID: 1, Level: 2, Sort: 5, Status: 1},
			}},
		}},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := catHandler.NewHandler(mock)
	// 注册路由，路径与 category/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/category")
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
	env := newHandlerEnv(t, 1, 2)
	// Level 必须满足 dto.CreateCategoryRequest 的 binding:"oneof=1 2 3"（handler 不使用 req.Level，
	// service 按 ParentID 重新计算，但绑定校验先于此执行，故需传入合法值）
	body := dto.CreateCategoryRequest{Name: "二手物品", Icon: "phone", ParentID: 0, Level: 1, Sort: 10, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/category", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	// 透传 regionID
	assert.Equal(t, uint(2), env.mock.lastCreateRegionID)
	// 透传请求体
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "二手物品", env.mock.lastCreateReq.Name)
	assert.Equal(t, uint(0), env.mock.lastCreateReq.ParentID)
	// data 透传 service 返回值
	var info dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "二手物品", info.Name)
}

func TestHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2) // 未登录
	body := dto.CreateCategoryRequest{Name: "二手物品"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/category", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	// 未登录不应调用 service
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 非法 JSON 触发 ShouldBind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/category", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.createErr = errors.New("父分类不存在")
	env.mock.createResult = nil
	body := dto.CreateCategoryRequest{Name: "子分类", ParentID: 999, Level: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/category", body)

	// 业务码 CodeCategoryError=2301 + err.Error() 透传
	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "父分类不存在", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateRegionID)
}

// ---------- Update ----------

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateCategoryRequest{Name: "二手手机", Sort: 20, Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/category/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "二手手机", env.mock.lastUpdateReq.Name)
}

func TestHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/category/5", dto.UpdateCategoryRequest{Name: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/category/abc", dto.UpdateCategoryRequest{Name: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的分类ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/category/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateErr = errors.New("分类不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/category/5", dto.UpdateCategoryRequest{Name: "x"})

	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "分类不存在", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/category/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteID)
}

func TestHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/category/7", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/category/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的分类ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.deleteErr = errors.New("该分类存在子分类，无法删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/category/7", nil)

	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "该分类存在子分类，无法删除", resp.Message)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetByID)
	var info dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/notnum", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的分类ID", resp.Message)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getByIDErr = errors.New("分类不存在")
	env.mock.getByIDResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/999", nil)

	// CodeCategoryNotFound=2302
	assert.Equal(t, 2302, resp.Code)
	assert.Equal(t, "分类不存在", resp.Message)
}

// ---------- GetByParentID ----------

func TestHandler_GetByParentID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/children?parent_id=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastGetByParent.ParentID)
	assert.Equal(t, uint(2), env.mock.lastGetByParent.RegionID)
	var list []dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "手机", list[0].Name)
}

func TestHandler_GetByParentID_DefaultParentZero(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 不传 parent_id → 默认 0
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/children", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetByParent.ParentID)
	assert.Equal(t, uint(2), env.mock.lastGetByParent.RegionID)
}

func TestHandler_GetByParentID_InvalidParentID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 非数字 parent_id → ParseUint 失败兜底为 0（handler 不报错）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/children?parent_id=abc", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetByParent.ParentID)
}

func TestHandler_GetByParentID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getByParentErr = errors.New("db down")
	env.mock.getByParentResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/children?parent_id=1", nil)

	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetTree ----------

func TestHandler_GetTree_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/tree", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastGetTreeRegionID)
	var tree []dto.CategoryTree
	require.NoError(t, json.Unmarshal(resp.Data, &tree))
	require.Len(t, tree, 1)
	assert.Equal(t, "二手物品", tree[0].Name)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "手机", tree[0].Children[0].Name)
}

func TestHandler_GetTree_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getTreeErr = errors.New("redis boom")
	env.mock.getTreeResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/tree", nil)

	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// ---------- GetAll ----------

func TestHandler_GetAll_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastGetAllRegionID)
	var list []dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "二手物品", list[0].Name)
}

func TestHandler_GetAll_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getAllErr = errors.New("boom")
	env.mock.getAllResult = nil
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category", nil)

	assert.Equal(t, 2301, resp.Code)
	assert.Equal(t, "boom", resp.Message)
}

// ---------- regionID 注入：默认地区兜底 ----------

func TestHandler_RegionID_ZeroFallback(t *testing.T) {
	// regionID=0 模拟 Region 中间件未注入，handler.getRegionID 兜底返回 DefaultRegionID=2
	env := newHandlerEnv(t, 1, 0)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/category/tree", nil)

	// getRegionID: ctx.Get(RegionIDKey) 取到 0 但类型断言 uint(0) 成功 → 返回 0
	// 这里验证 handler 透传 context 中的 regionID（0），不兜底（兜底只在 ctx.Get 返回 !ok 时）
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetTreeRegionID)
}
