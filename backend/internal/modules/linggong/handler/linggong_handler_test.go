// Package handler_test 同城零工兼职主表 Linggong HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 LinggongHandler 全部分支：
//   - 公开只读接口无需登录（GetByID/List/Search/Nearby/ListByEmployer/IncrView/IncrContact/IncrShare）
//   - 需登录接口未登录拦截（Create/Update/Delete/ListMine → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID" / ListByEmployer → "无效的雇主ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/min/max → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeLinggongPublishError=5203 / CodeLinggongError=5201 /
//     CodeLinggongNotFound=5202 / CodeLinggongAuditError=5204 / CodeLinggongStatusInvalid=5206）
//   - 地区ID/用户信息上下文注入（regionID/userID/userName/phone/avatar 透传给 service）
//   - Search 空 keyword 拦截（400 "keyword 必填"）
//   - Nearby 经纬度越界拦截（400 "经纬度参数无效"）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//   - M 端审核/状态/批量操作装配层逻辑
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 car/mall/dh114/house/job/love/marketing/shop/groupbuy handler 测试同风格。
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/linggong/dto"
	linggongHandler "wuchang-tongcheng/internal/modules/linggong/handler"
	"wuchang-tongcheng/internal/modules/linggong/service"
	"wuchang-tongcheng/internal/pkg/utils"

	"errors"
)

// apiResponse 解析统一响应体 {code, message, data}
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// pageData 解析 PageResult {list, total, page, pageSize}
type pageData struct {
	List     json.RawMessage `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// mockLinggongService 内存 mock，实现 service.LinggongService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockLinggongService struct {
	// ===== 调用记录（C 端） =====
	lastCreateRegionID    uint
	lastCreateUserID      uint
	lastCreateUserName    string
	lastCreateUserAvatar  string
	lastCreateUserPhone   string
	lastCreateReq         *dto.CreateLinggongRequest
	lastCreateReqTitle    string

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateLinggongRequest

	lastDeleteID         uint
	lastDeleteOperatorID uint

	lastGetByIDID uint

	lastListRegionID uint
	lastListReq      *dto.LinggongListRequest

	lastListMineUserID   uint
	lastListMinePage      int
	lastListMinePageSize  int

	lastListByEmployerID      uint
	lastListByEmployerPage     int
	lastListByEmployerPageSize int

	lastSearchRegionID  uint
	lastSearchKeyword   string
	lastSearchPage      int
	lastSearchPageSize  int

	lastNearbyRegionID  uint
	lastNearbyLat       float64
	lastNearbyLng       float64
	lastNearbyRadius    float64
	lastNearbyPage      int
	lastNearbyPageSize  int

	lastIncrViewID     uint
	lastIncrContactID  uint
	lastIncrShareID     uint

	// ===== 调用记录（M 端） =====
	lastAdminListReq *dto.LinggongAdminListRequest

	lastAdminGetByIDID uint

	lastAuditID  uint
	lastAuditReq *dto.LinggongAuditRequest

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	lastBatchUpdateStatusIDs    []uint
	lastBatchUpdateStatusStatus int

	// ===== 预设返回值 =====
	createResult *dto.LinggongInfo

	getByIDResult *dto.LinggongInfo

	listResult []dto.LinggongInfo
	listTotal  int64

	listMineResult []dto.LinggongInfo
	listMineTotal  int64

	listByEmployerResult []dto.LinggongInfo
	listByEmployerTotal  int64

	searchResult []dto.LinggongInfo
	searchTotal  int64

	nearbyResult []dto.LinggongInfo
	nearbyTotal  int64

	adminListResult []dto.LinggongInfo
	adminListTotal  int64

	adminGetByIDResult *dto.LinggongInfo

	// ===== 错误注入 =====
	createErr           error
	updateErr           error
	deleteErr           error
	getByIDErr          error
	listErr             error
	listMineErr         error
	listByEmployerErr   error
	searchErr           error
	nearbyErr           error
	incrViewErr         error
	incrContactErr      error
	incrShareErr        error
	adminListErr        error
	adminGetByIDErr     error
	auditErr            error
	updateStatusErr     error
	batchUpdateStatusErr error
}

func (m *mockLinggongService) Create(regionID uint, userID uint, userName string, userAvatar string, userPhone string, req *dto.CreateLinggongRequest) (*dto.LinggongInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateUserName = userName
	m.lastCreateUserAvatar = userAvatar
	m.lastCreateUserPhone = userPhone
	m.lastCreateReq = req
	if req != nil {
		m.lastCreateReqTitle = req.Title
	}
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockLinggongService) Update(id uint, operatorID uint, req *dto.UpdateLinggongRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockLinggongService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockLinggongService) GetByID(id uint) (*dto.LinggongInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockLinggongService) List(regionID uint, req *dto.LinggongListRequest) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pagination := utils.NewPagination(1, 10)
	pagination.Total = m.listTotal
	return pagination, m.listResult, nil
}

func (m *mockLinggongService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastListMineUserID = userID
	m.lastListMinePage = page
	m.lastListMinePageSize = pageSize
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = m.listMineTotal
	return pagination, m.listMineResult, nil
}

func (m *mockLinggongService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastListByEmployerID = employerID
	m.lastListByEmployerPage = page
	m.lastListByEmployerPageSize = pageSize
	if m.listByEmployerErr != nil {
		return nil, nil, m.listByEmployerErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = m.listByEmployerTotal
	return pagination, m.listByEmployerResult, nil
}

func (m *mockLinggongService) Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchKeyword = keyword
	m.lastSearchPage = page
	m.lastSearchPageSize = pageSize
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = m.searchTotal
	return pagination, m.searchResult, nil
}

func (m *mockLinggongService) Nearby(regionID uint, lat, lng, radiusKm float64, page, pageSize int) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyLat = lat
	m.lastNearbyLng = lng
	m.lastNearbyRadius = radiusKm
	m.lastNearbyPage = page
	m.lastNearbyPageSize = pageSize
	if m.nearbyErr != nil {
		return nil, nil, m.nearbyErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = m.nearbyTotal
	return pagination, m.nearbyResult, nil
}

func (m *mockLinggongService) IncrViewCount(id uint) error {
	m.lastIncrViewID = id
	return m.incrViewErr
}

func (m *mockLinggongService) IncrContactCount(id uint) error {
	m.lastIncrContactID = id
	return m.incrContactErr
}

func (m *mockLinggongService) IncrShareCount(id uint) error {
	m.lastIncrShareID = id
	return m.incrShareErr
}

func (m *mockLinggongService) AdminList(req *dto.LinggongAdminListRequest) (*utils.Pagination, []dto.LinggongInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	pagination := utils.NewPagination(1, 10)
	pagination.Total = m.adminListTotal
	return pagination, m.adminListResult, nil
}

func (m *mockLinggongService) AdminGetByID(id uint) (*dto.LinggongInfo, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockLinggongService) Audit(id uint, req *dto.LinggongAuditRequest) error {
	m.lastAuditID = id
	m.lastAuditReq = req
	return m.auditErr
}

func (m *mockLinggongService) UpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.updateStatusErr
}

func (m *mockLinggongService) BatchUpdateStatus(ids []uint, status int) error {
	m.lastBatchUpdateStatusIDs = ids
	m.lastBatchUpdateStatusStatus = status
	return m.batchUpdateStatusErr
}

// 编译期接口实现校验
var _ service.LinggongService = (*mockLinggongService)(nil)

// handlerEnv 测试环境：gin 引擎 + mock service。
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockLinggongService
}

// newHandlerEnv 构造 gin 引擎并注册 linggong 主表路由（路径与 linggong/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉限流/权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockLinggongService{
		createResult: &dto.LinggongInfo{
			ID: 1, UserID: ctxUserID, Title: "周末活动协助", Status: 0, RegionID: regionID,
		},
		getByIDResult: &dto.LinggongInfo{
			ID: 1, UserID: 1, Title: "周末活动协助", Status: 1,
		},
		listResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助", Status: 1},
			{ID: 2, Title: "餐厅服务员", Status: 1},
		},
		listTotal: 2,
		listMineResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助", Status: 0},
		},
		listMineTotal: 1,
		listByEmployerResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助", EmployerID: 7},
		},
		listByEmployerTotal: 1,
		searchResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助"},
		},
		searchTotal: 1,
		nearbyResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助"},
		},
		nearbyTotal: 1,
		adminGetByIDResult: &dto.LinggongInfo{
			ID: 1, UserID: 1, Title: "周末活动协助", Status: 1, AuditStatus: 1,
		},
		adminListResult: []dto.LinggongInfo{
			{ID: 1, Title: "周末活动协助", Status: 1, AuditStatus: 1},
			{ID: 2, Title: "餐厅服务员", Status: 0, AuditStatus: 0},
		},
		adminListTotal: 2,
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和登录用户信息
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Set(middleware.ContextUsername, "张三")
		c.Set(middleware.ContextUserPhone, "13800000000")
		c.Set(middleware.ContextUserAvatar, "https://cdn.example.com/a.png")
		c.Next()
	})

	h := linggongHandler.NewLinggongHandler(mock)
	root := r.Group("/api/v1/linggong")
	// 公开只读接口（固定路径必须注册在 /:id 之前，避免被参数路由吞掉）
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/mine", h.ListMine)
	root.GET("/:id", h.GetByID)
	root.POST("/:id/contact", h.IncrContact)
	root.POST("/:id/share", h.IncrShare)
	root.POST("/:id/view", h.IncrView)
	root.GET("/employers/:id/linggongs", h.ListByEmployer)
	// 需登录接口（C 端）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/linggongs", h.AdminList)
	admin.GET("/linggongs/:id", h.AdminGetByID)
	admin.PUT("/linggongs/:id/audit", h.Audit)
	admin.PUT("/linggongs/:id/status", h.AdminUpdateStatus)
	admin.POST("/linggongs/batch-status", h.BatchUpdateStatus)

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

// doQuery 发起无 body 的请求（GET/DELETE，用于 query 透传断言）。
func (e *handlerEnv) doQuery(t *testing.T, method, path string) *apiResponse {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
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

// parsePage 解析响应 data 为 pageData
func parsePage(t *testing.T, resp *apiResponse) *pageData {
	t.Helper()
	var p pageData
	require.NoError(t, json.Unmarshal(resp.Data, &p))
	return &p
}

// validCreateBody 构造一份可通过 Bind 校验的完整 CreateLinggongRequest 请求体。
func validCreateBody() map[string]interface{} {
	return map[string]interface{}{
		"title":          "周末活动协助",
		"content":        "周末两天活动协助",
		"linggong_type":  "short_term",
		"publisher_type": "personal",
		"billing_type":   "by_day",
		"salary_min":     200,
		"salary_max":     300,
		"salary_unit":    "元/天",
		"settlement":     "T+1",
		"work_intensity": "medium",
		"recruit_count":  5,
		"need_gender":    "any",
		"work_location_type": "onsite",
		"city":           "武昌",
		"address":        "武昌区某路1号",
	}
}

// ==================== 公开只读接口（无需登录） ====================

// ---------- GetByID ----------

func TestLinggongHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	var info dto.LinggongInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "周末活动协助", info.Title)
}

func TestLinggongHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestLinggongHandler_GetByID_NotFound(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDErr = errors.New("岗位不存在")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/99")

	assert.Equal(t, utils.CodeLinggongNotFound, resp.Code)
	assert.Equal(t, "岗位不存在", resp.Message)
}

// ---------- List ----------

func TestLinggongHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong?linggong_type=short_term&page=2&page_size=5")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "short_term", env.mock.lastListReq.LinggongType)

	page := parsePage(t, resp)
	assert.Equal(t, int64(2), page.Total)
	var items []dto.LinggongInfo
	require.NoError(t, json.Unmarshal(page.List, &items))
	assert.Equal(t, 2, len(items))
}

func TestLinggongHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listErr = errors.New("db error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- Search ----------

func TestLinggongHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/search?keyword=周末&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	assert.Equal(t, "周末", env.mock.lastSearchKeyword)
	assert.Equal(t, 1, env.mock.lastSearchPage)
	assert.Equal(t, 10, env.mock.lastSearchPageSize)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestLinggongHandler_Search_EmptyKeyword(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/search?keyword=")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "keyword 必填", resp.Message)
	// 空 keyword 被拦截，service 未被调用
	assert.Equal(t, "", env.mock.lastSearchKeyword)
}

func TestLinggongHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchErr = errors.New("search error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/search?keyword=周末")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "search error", resp.Message)
}

// ---------- Nearby ----------

func TestLinggongHandler_Nearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=30.5&longitude=114.3&radius_km=10&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastNearbyRegionID)
	assert.Equal(t, 30.5, env.mock.lastNearbyLat)
	assert.Equal(t, 114.3, env.mock.lastNearbyLng)
	assert.Equal(t, 10.0, env.mock.lastNearbyRadius)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestLinggongHandler_Nearby_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude 非数字触发 form Bind 失败
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=abc&longitude=114.3")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Nearby_LatOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=200&longitude=114.3")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestLinggongHandler_Nearby_LngOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=30.5&longitude=300")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestLinggongHandler_Nearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.nearbyErr = errors.New("nearby error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=30.5&longitude=114.3")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "nearby error", resp.Message)
}

// ---------- ListByEmployer ----------

func TestLinggongHandler_ListByEmployer_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/employers/7/linggongs?page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListByEmployerID)
	assert.Equal(t, 1, env.mock.lastListByEmployerPage)
	assert.Equal(t, 10, env.mock.lastListByEmployerPageSize)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestLinggongHandler_ListByEmployer_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/employers/abc/linggongs")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的雇主ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListByEmployerID)
}

func TestLinggongHandler_ListByEmployer_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listByEmployerErr = errors.New("employer error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/employers/7/linggongs")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "employer error", resp.Message)
}

// ---------- IncrView / IncrContact / IncrShare（公开 POST） ----------

func TestLinggongHandler_IncrView_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/view")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录浏览", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrViewID)
}

func TestLinggongHandler_IncrView_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/abc/view")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastIncrViewID)
}

func TestLinggongHandler_IncrView_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrViewErr = errors.New("incr view error")
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/view")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "incr view error", resp.Message)
}

func TestLinggongHandler_IncrContact_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/contact")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录联系", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrContactID)
}

func TestLinggongHandler_IncrContact_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/abc/contact")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_IncrContact_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrContactErr = errors.New("incr contact error")
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/contact")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "incr contact error", resp.Message)
}

func TestLinggongHandler_IncrShare_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/share")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录分享", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrShareID)
}

func TestLinggongHandler_IncrShare_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/abc/share")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_IncrShare_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrShareErr = errors.New("incr share error")
	resp := env.doQuery(t, http.MethodPost, "/api/v1/linggong/3/share")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "incr share error", resp.Message)
}

// ==================== 需登录接口（C 端） ====================

// ---------- Create ----------

func TestLinggongHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", validCreateBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	// 用户信息 + regionID 透传
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "张三", env.mock.lastCreateUserName)
	assert.Equal(t, "13800000000", env.mock.lastCreateUserPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateUserAvatar)
	// 请求体透传
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "周末活动协助", env.mock.lastCreateReq.Title)
	assert.Equal(t, "short_term", env.mock.lastCreateReq.LinggongType)
	assert.Equal(t, 5, env.mock.lastCreateReq.RecruitCount)
}

func TestLinggongHandler_Create_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", validCreateBody())

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestLinggongHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/linggong", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestLinggongHandler_Create_BindFail_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	delete(body, "title")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestLinggongHandler_Create_BindFail_InvalidLinggongType(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["linggong_type"] = "invalid_type"
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Create_BindFail_RecruitCountZero(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["recruit_count"] = 0 // min=1
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.createErr = errors.New("发布失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong", validCreateBody())

	// Create 错误走 CodeLinggongPublishError（区别于其他接口的 CodeLinggongError）
	assert.Equal(t, utils.CodeLinggongPublishError, resp.Code)
	assert.Equal(t, "发布失败", resp.Message)
}

// ---------- Update ----------

func TestLinggongHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	title := "更新后标题"
	body := map[string]interface{}{"title": title, "status": 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	require.NotNil(t, env.mock.lastUpdateReq.Title)
	assert.Equal(t, title, *env.mock.lastUpdateReq.Title)
}

func TestLinggongHandler_Update_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLinggongHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestLinggongHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/linggong/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateErr = errors.New("无权操作此岗位")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "无权操作此岗位", resp.Message)
}

// ---------- Delete ----------

func TestLinggongHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/linggong/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestLinggongHandler_Delete_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/linggong/3")

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLinggongHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/linggong/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.deleteErr = errors.New("岗位不存在")
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/linggong/3")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "岗位不存在", resp.Message)
}

// ---------- ListMine ----------

func TestLinggongHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/mine?page=2&page_size=5")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 5, env.mock.lastListMinePageSize)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestLinggongHandler_ListMine_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/mine")

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListMineUserID)
}

func TestLinggongHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.listMineErr = errors.New("list mine error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/mine")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "list mine error", resp.Message)
}

func TestLinggongHandler_ListMine_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 不传分页参数，handler 兜底 page=1/page_size=10
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/mine")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, 1, env.mock.lastListMinePage)
	assert.Equal(t, 10, env.mock.lastListMinePageSize)
}

// ==================== M 端管理 ====================

// ---------- AdminList ----------

func TestLinggongHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/admin/linggongs?status=1&keyword=周末&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	keyword := "周末"
	assert.Equal(t, keyword, env.mock.lastAdminListReq.Keyword)

	page := parsePage(t, resp)
	assert.Equal(t, int64(2), page.Total)
	var items []dto.LinggongInfo
	require.NoError(t, json.Unmarshal(page.List, &items))
	assert.Equal(t, 2, len(items))
}

func TestLinggongHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminListErr = errors.New("admin list error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/admin/linggongs")

	assert.Equal(t, utils.CodeLinggongError, resp.Code)
	assert.Equal(t, "admin list error", resp.Message)
}

// ---------- AdminGetByID ----------

func TestLinggongHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/admin/linggongs/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.LinggongInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestLinggongHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/admin/linggongs/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_AdminGetByID_NotFound(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminGetByIDErr = errors.New("岗位不存在")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/admin/linggongs/99")

	assert.Equal(t, utils.CodeLinggongNotFound, resp.Code)
	assert.Equal(t, "岗位不存在", resp.Message)
}

// ---------- Audit ----------

func TestLinggongHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"audit_status": 1, "audit_reason": "通过"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	require.NotNil(t, env.mock.lastAuditReq)
	assert.Equal(t, 1, env.mock.lastAuditReq.AuditStatus)
	assert.Equal(t, "通过", env.mock.lastAuditReq.AuditReason)
}

func TestLinggongHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/abc/audit", map[string]interface{}{"audit_status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_Audit_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Audit_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status oneof=0 1 2，传 9 触发校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/audit", map[string]interface{}{"audit_status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.auditErr = errors.New("已审核的岗位不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/audit", map[string]interface{}{"audit_status": 1})

	assert.Equal(t, utils.CodeLinggongAuditError, resp.Code)
	assert.Equal(t, "已审核的岗位不能重复审核", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestLinggongHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"status": 2}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestLinggongHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/abc/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLinggongHandler_AdminUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_AdminUpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// status oneof=1 2 3 4 5 6 7，传 0 触发校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/status", map[string]interface{}{"status": 0})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updateStatusErr = errors.New("岗位状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/linggong/admin/linggongs/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, utils.CodeLinggongStatusInvalid, resp.Code)
	assert.Equal(t, "岗位状态不允许此操作", resp.Message)
}

// ---------- BatchUpdateStatus ----------

func TestLinggongHandler_BatchUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"ids": []uint{1, 2, 3}, "status": 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong/admin/linggongs/batch-status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量状态更新完成", resp.Message)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchUpdateStatusIDs)
	assert.Equal(t, 2, env.mock.lastBatchUpdateStatusStatus)
}

func TestLinggongHandler_BatchUpdateStatus_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/linggong/admin/linggongs/batch-status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_BatchUpdateStatus_BindFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// ids binding:"required,min=1"，空数组触发校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong/admin/linggongs/batch-status", map[string]interface{}{"ids": []uint{}, "status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_BatchUpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong/admin/linggongs/batch-status", map[string]interface{}{"ids": []uint{1}, "status": 0})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestLinggongHandler_BatchUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchUpdateStatusErr = errors.New("批量更新失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/linggong/admin/linggongs/batch-status", map[string]interface{}{"ids": []uint{1, 2}, "status": 2})

	assert.Equal(t, utils.CodeLinggongStatusInvalid, resp.Code)
	assert.Equal(t, "批量更新失败", resp.Message)
}

// ==================== 公开读取无需登录聚合 ====================

// TestHandler_PublicRead_NoAuthRequired 验证四个只读接口在 userID=0（未登录）时不被 401 拦截。
func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)

	resp := env.doQuery(t, http.MethodGet, "/api/v1/linggong/3")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/linggong")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/linggong/search?keyword=周末")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=30.5&longitude=114.3")
	assert.Equal(t, 0, resp.Code)
}

// ==================== regionID 注入聚合 ====================

// TestHandler_RegionIDInjection 验证所有带 regionID 入参的接口均透传中间件注入的 regionID。
func TestHandler_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)

	env.doQuery(t, http.MethodGet, "/api/v1/linggong")
	assert.Equal(t, uint(9), env.mock.lastListRegionID)

	env.doQuery(t, http.MethodGet, "/api/v1/linggong/search?keyword=周末")
	assert.Equal(t, uint(9), env.mock.lastSearchRegionID)

	env.doQuery(t, http.MethodGet, "/api/v1/linggong/nearby?latitude=30.5&longitude=114.3")
	assert.Equal(t, uint(9), env.mock.lastNearbyRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/linggong", validCreateBody())
	assert.Equal(t, uint(9), env.mock.lastCreateRegionID)
}
