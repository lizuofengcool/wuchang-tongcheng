// Package handler_test 同城拼车出行主表 Pinche HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 PincheHandler 全部分支：
//   - 公开只读接口无需登录（GetByID/List/Search/Nearby/Match/IncrContact/IncrShare/RecordView）
//   - 需登录接口未登录拦截（Create/Update/Delete/ListMine → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/min/max → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodePincheError=5101 / CodePincheNotFound=5102 /
//     CodePincheAuditError=5122）
//   - 地区ID/用户信息上下文注入（regionID/userID/userName/phone/avatar 透传给 service）
//   - Nearby 经纬度越界拦截（400 "经纬度参数无效"）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//   - M 端审核/状态/批量操作装配层逻辑
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 linggong/car/mall/dh114/house/job/love/marketing/shop/groupbuy handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/pinche/dto"
	pincheHandler "wuchang-tongcheng/internal/modules/pinche/handler"
	"wuchang-tongcheng/internal/modules/pinche/service"
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

// mockPincheService 内存 mock，实现 service.PincheService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockPincheService struct {
	// ===== 调用记录（C 端） =====
	lastCreateRegionID   uint
	lastCreateUserID     uint
	lastCreateUserName   string
	lastCreateUserPhone  string
	lastCreateUserAvatar string
	lastCreateReq        *dto.CreatePincheRequest

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdatePincheRequest

	lastDeleteID         uint
	lastDeleteOperatorID uint

	lastGetByIDID     uint
	lastGetByIDUserID uint

	lastListRegionID uint
	lastListReq      *dto.PincheListRequest

	lastNearbyRegionID uint
	lastNearbyReq      *dto.PincheNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.PincheSearchRequest

	lastListMineUserID   uint
	lastListMinePage     int
	lastListMinePageSize int

	lastMatchRegionID uint
	lastMatchReq      *dto.PincheMatchRequest

	lastIncrContactID uint
	lastIncrShareID   uint

	lastRecordViewUserID uint
	lastRecordViewReq    *dto.PincheViewRequest

	// ===== 调用记录（M 端） =====
	lastAdminListReq *dto.PincheAdminListRequest

	lastAdminGetByIDID uint

	lastAuditID       uint
	lastAuditStatus   int
	lastAuditReason   string

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	lastBatchAuditReq          *dto.BatchAuditRequest
	lastBatchUpdateStatusReq   *dto.BatchStatusUpdateRequest
	lastBatchDeleteReq         *dto.BatchDeleteRequest

	// ===== 预设返回值 =====
	createResult *dto.PincheInfo

	getByIDResult *dto.PincheInfo

	listResult []dto.PincheInfo
	listTotal  int64

	nearbyResult []dto.PincheInfo
	nearbyTotal  int64

	searchResult []dto.PincheInfo
	searchTotal  int64

	listMineResult []dto.PincheInfo
	listMineTotal  int64

	matchResult *dto.PincheMatchResponse

	adminListResult []dto.PincheInfo
	adminListTotal  int64

	adminGetByIDResult *dto.PincheInfo

	batchAuditResult         *dto.BatchResultResponse
	batchUpdateStatusResult  *dto.BatchResultResponse
	batchDeleteResult        *dto.BatchResultResponse

	// ===== 错误注入 =====
	createErr            error
	updateErr            error
	deleteErr            error
	getByIDErr           error
	listErr              error
	nearbyErr            error
	searchErr            error
	listMineErr          error
	matchErr             error
	incrContactErr       error
	incrShareErr         error
	recordViewErr        error
	adminListErr         error
	adminGetByIDErr      error
	auditErr             error
	updateStatusErr      error
	batchAuditErr        error
	batchUpdateStatusErr error
	batchDeleteErr       error
}

func (m *mockPincheService) Create(regionID uint, userID uint, userName, userPhone, userAvatar string, req *dto.CreatePincheRequest) (*dto.PincheInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateUserName = userName
	m.lastCreateUserPhone = userPhone
	m.lastCreateUserAvatar = userAvatar
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockPincheService) Update(id uint, operatorID uint, req *dto.UpdatePincheRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockPincheService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockPincheService) GetByID(id uint, userID uint) (*dto.PincheInfo, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUserID = userID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockPincheService) List(regionID uint, req *dto.PincheListRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pagination := utils.NewPagination(1, 10)
	pagination.Total = m.listTotal
	return pagination, m.listResult, nil
}

func (m *mockPincheService) ListNearby(regionID uint, req *dto.PincheNearbyRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyReq = req
	if m.nearbyErr != nil {
		return nil, nil, m.nearbyErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = m.nearbyTotal
	return pagination, m.nearbyResult, nil
}

func (m *mockPincheService) Search(regionID uint, req *dto.PincheSearchRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = m.searchTotal
	return pagination, m.searchResult, nil
}

func (m *mockPincheService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.PincheInfo, error) {
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

func (m *mockPincheService) Match(regionID uint, req *dto.PincheMatchRequest) (*dto.PincheMatchResponse, error) {
	m.lastMatchRegionID = regionID
	m.lastMatchReq = req
	if m.matchErr != nil {
		return nil, m.matchErr
	}
	return m.matchResult, nil
}

func (m *mockPincheService) IncrContact(id uint) error {
	m.lastIncrContactID = id
	return m.incrContactErr
}

func (m *mockPincheService) IncrShare(id uint) error {
	m.lastIncrShareID = id
	return m.incrShareErr
}

func (m *mockPincheService) RecordView(userID uint, req *dto.PincheViewRequest) error {
	m.lastRecordViewUserID = userID
	m.lastRecordViewReq = req
	return m.recordViewErr
}

func (m *mockPincheService) AdminList(req *dto.PincheAdminListRequest) (*utils.Pagination, []dto.PincheInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	pagination := utils.NewPagination(1, 10)
	pagination.Total = m.adminListTotal
	return pagination, m.adminListResult, nil
}

func (m *mockPincheService) AdminGetByID(id uint) (*dto.PincheInfo, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockPincheService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockPincheService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.updateStatusErr
}

func (m *mockPincheService) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	m.lastBatchAuditReq = req
	if m.batchAuditErr != nil {
		return nil, m.batchAuditErr
	}
	return m.batchAuditResult, nil
}

func (m *mockPincheService) BatchUpdateStatus(req *dto.BatchStatusUpdateRequest) (*dto.BatchResultResponse, error) {
	m.lastBatchUpdateStatusReq = req
	if m.batchUpdateStatusErr != nil {
		return nil, m.batchUpdateStatusErr
	}
	return m.batchUpdateStatusResult, nil
}

func (m *mockPincheService) BatchDelete(req *dto.BatchDeleteRequest) (*dto.BatchResultResponse, error) {
	m.lastBatchDeleteReq = req
	if m.batchDeleteErr != nil {
		return nil, m.batchDeleteErr
	}
	return m.batchDeleteResult, nil
}

// 编译期接口实现校验
var _ service.PincheService = (*mockPincheService)(nil)

// handlerEnv 测试环境：gin 引擎 + mock service。
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockPincheService
}

// newHandlerEnv 构造 gin 引擎并注册 pinche 主表路由（路径与 pinche/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉限流/权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockPincheService{
		createResult: &dto.PincheInfo{
			ID: 1, UserID: ctxUserID, Title: "武昌到光谷顺风车", Status: 1, RegionID: regionID,
		},
		getByIDResult: &dto.PincheInfo{
			ID: 1, UserID: 1, Title: "武昌到光谷顺风车", Status: 1,
		},
		listResult: []dto.PincheInfo{
			{ID: 1, Title: "武昌到光谷顺风车", Status: 1},
			{ID: 2, Title: "光谷到武昌拼车", Status: 1},
		},
		listTotal: 2,
		nearbyResult: []dto.PincheInfo{
			{ID: 1, Title: "武昌到光谷顺风车"},
		},
		nearbyTotal: 1,
		searchResult: []dto.PincheInfo{
			{ID: 1, Title: "武昌到光谷顺风车"},
		},
		searchTotal: 1,
		listMineResult: []dto.PincheInfo{
			{ID: 1, Title: "武昌到光谷顺风车", Status: 1},
		},
		listMineTotal: 1,
		matchResult: &dto.PincheMatchResponse{
			Total: 1,
			List: []dto.PincheMatchItem{
				{PincheInfo: dto.PincheInfo{ID: 1, Title: "武昌到光谷顺风车"}, MatchScore: 85.5},
			},
		},
		adminGetByIDResult: &dto.PincheInfo{
			ID: 1, UserID: 1, Title: "武昌到光谷顺风车", Status: 1, AuditStatus: 1,
		},
		adminListResult: []dto.PincheInfo{
			{ID: 1, Title: "武昌到光谷顺风车", Status: 1, AuditStatus: 1},
			{ID: 2, Title: "光谷到武昌拼车", Status: 0, AuditStatus: 0},
		},
		adminListTotal: 2,
		batchAuditResult:         &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
		batchUpdateStatusResult:  &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
		batchDeleteResult:        &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
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

	h := pincheHandler.NewPincheHandler(mock)
	root := r.Group("/api/v1/pinche")
	// 公开只读接口（固定路径必须注册在 /:id 之前，避免被参数路由吞掉）
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/match", h.Match)
	root.GET("/mine", h.ListMine)
	root.POST("/views", h.RecordView)
	root.GET("/:id", h.GetByID)
	root.POST("/:id/contact", h.IncrContact)
	root.POST("/:id/share", h.IncrShare)
	// 需登录接口（C 端）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/pinches", h.AdminList)
	admin.GET("/pinches/:id", h.AdminGetByID)
	admin.PUT("/pinches/:id/audit", h.Audit)
	admin.PUT("/pinches/:id/status", h.AdminUpdateStatus)
	admin.POST("/pinches/batch-audit", h.BatchAudit)
	admin.POST("/pinches/batch-status", h.BatchUpdateStatus)
	admin.POST("/pinches/batch-delete", h.BatchDelete)

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

// validCreateBody 构造一份可通过 Bind 校验的完整 CreatePincheRequest 请求体。
func validCreateBody() map[string]interface{} {
	return map[string]interface{}{
		"trip_type":        "shunfeng",
		"role":             "driver",
		"title":            "武昌到光谷顺风车",
		"content":          "每天早上8点出发，武昌到光谷",
		"departure_time":   "2026-12-31T10:00:00Z",
		"pickup_location":  "武昌区中南路",
		"pickup_lat":       30.5728,
		"pickup_lng":       114.3055,
		"dropoff_location": "光谷广场",
		"dropoff_lat":      30.5419,
		"dropoff_lng":      114.4268,
		"distance_km":      15.5,
		"duration_min":     40,
		"total_seats":      3,
		"price_per_seat":   20,
		"toll_fee":         5,
		"payment_method":   "cash",
	}
}

// ==================== 公开只读接口（无需登录） ====================

// ---------- GetByID ----------

func TestPincheHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	var info dto.PincheInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "武昌到光谷顺风车", info.Title)
}

func TestPincheHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestPincheHandler_GetByID_NotFound(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDErr = errors.New("拼车行程不存在")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/99")

	assert.Equal(t, utils.CodePincheNotFound, resp.Code)
	assert.Equal(t, "拼车行程不存在", resp.Message)
}

func TestPincheHandler_GetByID_UserIDTransparency(t *testing.T) {
	// GetByID 透传 userID（登录用户带 userID，未登录带 0）
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastGetByIDUserID)
}

// ---------- List ----------

func TestPincheHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche?trip_type=shunfeng&page=2&page_size=5")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "shunfeng", env.mock.lastListReq.TripType)

	page := parsePage(t, resp)
	assert.Equal(t, int64(2), page.Total)
	var items []dto.PincheInfo
	require.NoError(t, json.Unmarshal(page.List, &items))
	assert.Equal(t, 2, len(items))
}

func TestPincheHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listErr = errors.New("db error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- Search ----------

func TestPincheHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/search?keyword=武昌&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "武昌", env.mock.lastSearchReq.Keyword)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestPincheHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchErr = errors.New("search error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/search?keyword=武昌")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "search error", resp.Message)
}

// ---------- Nearby ----------

func TestPincheHandler_Nearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=30.5&longitude=114.3&radius_km=10&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastNearbyRegionID)
	require.NotNil(t, env.mock.lastNearbyReq)
	assert.Equal(t, 30.5, env.mock.lastNearbyReq.Latitude)
	assert.Equal(t, 114.3, env.mock.lastNearbyReq.Longitude)
	assert.Equal(t, 10.0, env.mock.lastNearbyReq.RadiusKm)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestPincheHandler_Nearby_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude 非数字触发 form Bind 失败
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=abc&longitude=114.3")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Nearby_LatOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=200&longitude=114.3")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestPincheHandler_Nearby_LngOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=30.5&longitude=300")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestPincheHandler_Nearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.nearbyErr = errors.New("nearby error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=30.5&longitude=114.3")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "nearby error", resp.Message)
}

// ---------- Match（GET，query 参数） ----------

func TestPincheHandler_Match_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// PincheMatchRequest 无 form 标签，gin form binding 按 field name 匹配 query key
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/match?PickupLocation=武昌区&DropoffLocation=光谷&PickupLat=30.5&PickupLng=114.3&DropoffLat=30.6&DropoffLng=114.4&Seats=2&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastMatchRegionID)
	require.NotNil(t, env.mock.lastMatchReq)
	assert.Equal(t, "武昌区", env.mock.lastMatchReq.PickupLocation)
	assert.Equal(t, "光谷", env.mock.lastMatchReq.DropoffLocation)
	assert.Equal(t, 30.5, env.mock.lastMatchReq.PickupLat)
	assert.Equal(t, 2, env.mock.lastMatchReq.Seats)

	var matchResp dto.PincheMatchResponse
	require.NoError(t, json.Unmarshal(resp.Data, &matchResp))
	assert.Equal(t, 1, matchResp.Total)
	assert.Equal(t, 1, len(matchResp.List))
}

func TestPincheHandler_Match_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// 缺少必填的 PickupLocation / DropoffLocation
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/match")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastMatchReq)
}

func TestPincheHandler_Match_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.matchErr = errors.New("match error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/match?PickupLocation=武昌区&DropoffLocation=光谷")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "match error", resp.Message)
}

// ---------- IncrContact / IncrShare（公开 POST） ----------

func TestPincheHandler_IncrContact_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/3/contact")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录联系", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrContactID)
}

func TestPincheHandler_IncrContact_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/abc/contact")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastIncrContactID)
}

func TestPincheHandler_IncrContact_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrContactErr = errors.New("incr contact error")
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/3/contact")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "incr contact error", resp.Message)
}

func TestPincheHandler_IncrShare_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/3/share")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录分享", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrShareID)
}

func TestPincheHandler_IncrShare_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/abc/share")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestPincheHandler_IncrShare_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrShareErr = errors.New("incr share error")
	resp := env.doQuery(t, http.MethodPost, "/api/v1/pinche/3/share")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "incr share error", resp.Message)
}

// ---------- RecordView（公开 POST，JSON body） ----------

func TestPincheHandler_RecordView_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/views", map[string]interface{}{
		"pinche_id": 3,
		"source":    "detail",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录浏览", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastRecordViewReq.PincheID)
}

func TestPincheHandler_RecordView_WithUserID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/views", map[string]interface{}{
		"pinche_id": 3,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastRecordViewUserID)
}

func TestPincheHandler_RecordView_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/pinche/views", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastRecordViewReq)
}

func TestPincheHandler_RecordView_BindFail_MissingPincheID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/views", map[string]interface{}{
		"source": "detail",
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_RecordView_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.recordViewErr = errors.New("record view error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/views", map[string]interface{}{
		"pinche_id": 3,
	})

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "record view error", resp.Message)
}

// ==================== 需登录接口（C 端） ====================

// ---------- Create ----------

func TestPincheHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", validCreateBody())

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
	assert.Equal(t, "武昌到光谷顺风车", env.mock.lastCreateReq.Title)
	assert.Equal(t, "shunfeng", env.mock.lastCreateReq.TripType)
	assert.Equal(t, 3, env.mock.lastCreateReq.TotalSeats)
}

func TestPincheHandler_Create_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", validCreateBody())

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestPincheHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/pinche", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestPincheHandler_Create_BindFail_MissingDepartureTime(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	delete(body, "departure_time")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestPincheHandler_Create_BindFail_MissingPickupLocation(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	delete(body, "pickup_location")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Create_BindFail_InvalidTripType(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["trip_type"] = "invalid_type"
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Create_BindFail_TotalSeatsExceedMax(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["total_seats"] = 20 // max=10
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", body)

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.createErr = errors.New("发布失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche", validCreateBody())

	// Create 错误走 CodePincheError
	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "发布失败", resp.Message)
}

// ---------- Update ----------

func TestPincheHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	title := "更新后标题"
	body := map[string]interface{}{"title": title, "price_per_seat": 25}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	require.NotNil(t, env.mock.lastUpdateReq.Title)
	assert.Equal(t, title, *env.mock.lastUpdateReq.Title)
}

func TestPincheHandler_Update_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestPincheHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestPincheHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/pinche/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateErr = errors.New("无权操作此拼车行程")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "无权操作此拼车行程", resp.Message)
}

// ---------- Delete ----------

func TestPincheHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/pinche/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestPincheHandler_Delete_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/pinche/3")

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestPincheHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/pinche/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestPincheHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.deleteErr = errors.New("拼车行程不存在")
	resp := env.doQuery(t, http.MethodDelete, "/api/v1/pinche/3")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "拼车行程不存在", resp.Message)
}

// ---------- ListMine ----------

func TestPincheHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/mine?page=2&page_size=5")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 5, env.mock.lastListMinePageSize)

	page := parsePage(t, resp)
	assert.Equal(t, int64(1), page.Total)
}

func TestPincheHandler_ListMine_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/mine")

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListMineUserID)
}

func TestPincheHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.listMineErr = errors.New("list mine error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/mine")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "list mine error", resp.Message)
}

func TestPincheHandler_ListMine_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 不传分页参数，handler 兜底 page=1/page_size=10
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/mine")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, 1, env.mock.lastListMinePage)
	assert.Equal(t, 10, env.mock.lastListMinePageSize)
}

// ==================== M 端管理 ====================

// ---------- AdminList ----------

func TestPincheHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/admin/pinches?status=1&keyword=武昌&page=1&page_size=10")

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, "武昌", env.mock.lastAdminListReq.Keyword)

	page := parsePage(t, resp)
	assert.Equal(t, int64(2), page.Total)
	var items []dto.PincheInfo
	require.NoError(t, json.Unmarshal(page.List, &items))
	assert.Equal(t, 2, len(items))
}

func TestPincheHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminListErr = errors.New("admin list error")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/admin/pinches")

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "admin list error", resp.Message)
}

// ---------- AdminGetByID ----------

func TestPincheHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/admin/pinches/3")

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.PincheInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestPincheHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/admin/pinches/abc")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestPincheHandler_AdminGetByID_NotFound(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminGetByIDErr = errors.New("拼车行程不存在")
	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/admin/pinches/99")

	assert.Equal(t, utils.CodePincheNotFound, resp.Code)
	assert.Equal(t, "拼车行程不存在", resp.Message)
}

// ---------- Audit ----------

func TestPincheHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"audit_status": 1, "audit_reason": "通过"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditStatus)
	assert.Equal(t, "通过", env.mock.lastAuditReason)
}

func TestPincheHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/abc/audit", map[string]interface{}{"audit_status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestPincheHandler_Audit_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Audit_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status oneof=0 1 2，传 9 触发校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/audit", map[string]interface{}{"audit_status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.auditErr = errors.New("已审核的拼车行程不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/audit", map[string]interface{}{"audit_status": 1})

	assert.Equal(t, utils.CodePincheAuditError, resp.Code)
	assert.Equal(t, "已审核的拼车行程不能重复审核", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestPincheHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"status": 2}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestPincheHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/abc/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestPincheHandler_AdminUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_AdminUpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// status oneof=0 1 2 3 4，传 9 触发校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/status", map[string]interface{}{"status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updateStatusErr = errors.New("状态更新失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/pinche/admin/pinches/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "状态更新失败", resp.Message)
}

// ---------- BatchAudit ----------

func TestPincheHandler_BatchAudit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"ids": []uint{1, 2, 3}, "audit_status": 1, "audit_reason": "通过"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量审核完成", resp.Message)
	require.NotNil(t, env.mock.lastBatchAuditReq)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchAuditReq.IDs)
	assert.Equal(t, 1, env.mock.lastBatchAuditReq.AuditStatus)
}

func TestPincheHandler_BatchAudit_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchAudit_BindFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-audit", map[string]interface{}{
		"ids":          []uint{},
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchAudit_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status oneof=1 2，传 0 触发校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-audit", map[string]interface{}{
		"ids":          []uint{1},
		"audit_status": 0,
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchAudit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchAuditErr = errors.New("批量审核失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-audit", map[string]interface{}{
		"ids":          []uint{1, 2},
		"audit_status": 1,
	})

	assert.Equal(t, utils.CodePincheAuditError, resp.Code)
	assert.Equal(t, "批量审核失败", resp.Message)
}

// ---------- BatchUpdateStatus ----------

func TestPincheHandler_BatchUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"ids": []uint{1, 2, 3}, "status": 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量状态更新完成", resp.Message)
	require.NotNil(t, env.mock.lastBatchUpdateStatusReq)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchUpdateStatusReq.IDs)
	assert.Equal(t, 2, env.mock.lastBatchUpdateStatusReq.Status)
}

func TestPincheHandler_BatchUpdateStatus_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchUpdateStatus_BindFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-status", map[string]interface{}{
		"ids":    []uint{},
		"status": 2,
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchUpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// status oneof=0 1 2 3 4，传 9 触发校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-status", map[string]interface{}{
		"ids":    []uint{1},
		"status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchUpdateStatusErr = errors.New("批量更新失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-status", map[string]interface{}{
		"ids":    []uint{1, 2},
		"status": 2,
	})

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "批量更新失败", resp.Message)
}

// ---------- BatchDelete ----------

func TestPincheHandler_BatchDelete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{"ids": []uint{1, 2, 3}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-delete", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量删除完成", resp.Message)
	require.NotNil(t, env.mock.lastBatchDeleteReq)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchDeleteReq.IDs)
}

func TestPincheHandler_BatchDelete_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-delete", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchDelete_BindFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-delete", map[string]interface{}{
		"ids": []uint{},
	})

	assert.Equal(t, 400, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"))
}

func TestPincheHandler_BatchDelete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchDeleteErr = errors.New("批量删除失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/pinche/admin/pinches/batch-delete", map[string]interface{}{
		"ids": []uint{1, 2},
	})

	assert.Equal(t, utils.CodePincheError, resp.Code)
	assert.Equal(t, "批量删除失败", resp.Message)
}

// ==================== 公开读取无需登录聚合 ====================

// TestHandler_PublicRead_NoAuthRequired 验证只读接口在 userID=0（未登录）时不被 401 拦截。
func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)

	resp := env.doQuery(t, http.MethodGet, "/api/v1/pinche/3")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/pinche")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/pinche/search?keyword=武昌")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=30.5&longitude=114.3")
	assert.Equal(t, 0, resp.Code)

	resp = env.doQuery(t, http.MethodGet, "/api/v1/pinche/match?PickupLocation=武昌&DropoffLocation=光谷")
	assert.Equal(t, 0, resp.Code)
}

// ==================== regionID 注入聚合 ====================

// TestHandler_RegionIDInjection 验证所有带 regionID 入参的接口均透传中间件注入的 regionID。
func TestHandler_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 7, 9)

	env.doQuery(t, http.MethodGet, "/api/v1/pinche")
	assert.Equal(t, uint(9), env.mock.lastListRegionID)

	env.doQuery(t, http.MethodGet, "/api/v1/pinche/search?keyword=武昌")
	assert.Equal(t, uint(9), env.mock.lastSearchRegionID)

	env.doQuery(t, http.MethodGet, "/api/v1/pinche/nearby?latitude=30.5&longitude=114.3")
	assert.Equal(t, uint(9), env.mock.lastNearbyRegionID)

	env.doQuery(t, http.MethodGet, "/api/v1/pinche/match?PickupLocation=武昌&DropoffLocation=光谷")
	assert.Equal(t, uint(9), env.mock.lastMatchRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/pinche", validCreateBody())
	assert.Equal(t, uint(9), env.mock.lastCreateRegionID)
}
