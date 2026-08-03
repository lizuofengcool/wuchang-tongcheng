// Package handler_test 相亲交友主表 Love HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 LoveHandler 全部分支：
//   - 公开只读接口无需登录（GetByID/List/Nearby/Search/AdvancedSearch）
//   - 需登录接口未登录拦截（Create/Update/GetByUserID/UpdateLocation/UpdateVoiceIntro/MatchScore → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/max/min → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeLoveError=5001 / CodeLoveNotFound=5002）
//   - 地区ID/用户信息上下文注入（regionID/userID 透传给 service）
//   - Create 用户画像透传（regionID/userID）
//   - GetByID 浏览者 userID 透传（未登录 userID=0）
//   - MatchScore target_user_id 参数校验（空 → "target_user_id 不能为空" / 非数字 → "target_user_id 无效"）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//   - M 端审核/状态/精选/推荐/批量操作装配层逻辑
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 car/mall/dh114/house/job/marketing/shop/groupbuy handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/love/dto"
	loveHandler "wuchang-tongcheng/internal/modules/love/handler"
	"wuchang-tongcheng/internal/modules/love/service"
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

// mockLoveService 内存 mock，实现 service.LoveService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockLoveService struct {
	// ===== 调用记录 =====
	// C 端
	lastCreateRegionID uint
	lastCreateUserID   uint
	lastCreateReq      *dto.CreateLoveRequest

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateLoveRequest

	lastGetByIDID          uint
	lastGetByIDViewerUserID uint

	lastGetByUserID uint

	lastListRegionID uint
	lastListReq      *dto.LoveListRequest

	lastNearbyRegionID uint
	lastNearbyReq      *dto.LoveNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.LoveAdvancedSearchRequest

	lastAdvancedRegionID uint
	lastAdvancedReq      *dto.LoveAdvancedSearchRequest

	lastUpdateLocationID      uint
	lastUpdateLocationUserID  uint
	lastUpdateLocationReq     *dto.UpdateLocationRequest

	lastUpdateVoiceIntroID     uint
	lastUpdateVoiceIntroUserID uint
	lastUpdateVoiceIntroReq    *dto.UpdateVoiceIntroRequest

	lastMatchScoreUserA uint
	lastMatchScoreUserB uint

	// M 端
	lastAdminListReq *dto.LoveListRequest

	lastAdminGetByIDID uint

	lastAuditID         uint
	lastAuditStatus     int
	lastAuditReason     string

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	lastSetFeaturedID      uint
	lastSetFeaturedValue   bool

	lastSetPickedID    uint
	lastSetPickedValue bool

	lastBatchAuditIDs         []uint
	lastBatchAuditStatus      int
	lastBatchAuditReason      string

	lastBatchUpdateStatusIDs    []uint
	lastBatchUpdateStatusStatus int

	// ===== 返回值预设 =====
	createResult *dto.LoveInfo
	createErr    error

	updateErr error

	getByIDResult *dto.LoveInfo
	getByIDErr    error

	getByUserIDResult *dto.LoveInfo
	getByUserIDErr    error

	listResult []dto.LoveInfo
	listErr    error
	listTotal  int64

	nearbyResult []dto.LoveInfo
	nearbyErr    error
	nearbyTotal  int64

	searchResult []dto.LoveInfo
	searchErr    error
	searchTotal  int64

	advancedResult []dto.LoveInfo
	advancedErr    error
	advancedTotal  int64

	updateLocationErr error

	updateVoiceIntroErr error

	matchScoreResult *dto.LoveMatchScoreResponse
	matchScoreErr    error

	adminListResult []dto.LoveInfo
	adminListErr    error
	adminListTotal  int64

	adminGetByIDResult *dto.LoveInfo
	adminGetByIDErr    error

	auditErr error

	adminUpdateStatusErr error

	setFeaturedErr error

	setPickedErr error

	batchAuditErr error

	batchUpdateStatusErr error
}

// ===== C 端 =====

func (m *mockLoveService) Create(regionID uint, userID uint, req *dto.CreateLoveRequest) (*dto.LoveInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockLoveService) Update(id uint, operatorID uint, req *dto.UpdateLoveRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockLoveService) GetByID(id uint, viewerUserID uint) (*dto.LoveInfo, error) {
	m.lastGetByIDID = id
	m.lastGetByIDViewerUserID = viewerUserID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockLoveService) GetByUserID(userID uint) (*dto.LoveInfo, error) {
	m.lastGetByUserID = userID
	if m.getByUserIDErr != nil {
		return nil, m.getByUserIDErr
	}
	return m.getByUserIDResult, nil
}

func (m *mockLoveService) List(regionID uint, req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}

func (m *mockLoveService) ListNearby(regionID uint, req *dto.LoveNearbyRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyReq = req
	if m.nearbyErr != nil {
		return nil, nil, m.nearbyErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.nearbyTotal}, m.nearbyResult, nil
}

func (m *mockLoveService) Search(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.searchTotal}, m.searchResult, nil
}

func (m *mockLoveService) AdvancedSearch(regionID uint, req *dto.LoveAdvancedSearchRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	m.lastAdvancedRegionID = regionID
	m.lastAdvancedReq = req
	if m.advancedErr != nil {
		return nil, nil, m.advancedErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.advancedTotal}, m.advancedResult, nil
}

func (m *mockLoveService) UpdateLocation(id uint, userID uint, req *dto.UpdateLocationRequest) error {
	m.lastUpdateLocationID = id
	m.lastUpdateLocationUserID = userID
	m.lastUpdateLocationReq = req
	return m.updateLocationErr
}

func (m *mockLoveService) UpdateVoiceIntro(id uint, userID uint, req *dto.UpdateVoiceIntroRequest) error {
	m.lastUpdateVoiceIntroID = id
	m.lastUpdateVoiceIntroUserID = userID
	m.lastUpdateVoiceIntroReq = req
	return m.updateVoiceIntroErr
}

func (m *mockLoveService) MatchScore(userA, userB uint) (*dto.LoveMatchScoreResponse, error) {
	m.lastMatchScoreUserA = userA
	m.lastMatchScoreUserB = userB
	if m.matchScoreErr != nil {
		return nil, m.matchScoreErr
	}
	return m.matchScoreResult, nil
}

// ===== M 端 =====

func (m *mockLoveService) AdminList(req *dto.LoveListRequest) (*utils.Pagination, []dto.LoveInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.adminListTotal}, m.adminListResult, nil
}

func (m *mockLoveService) AdminGetByID(id uint) (*dto.LoveInfo, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockLoveService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockLoveService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.adminUpdateStatusErr
}

func (m *mockLoveService) SetFeatured(id uint, featured bool) error {
	m.lastSetFeaturedID = id
	m.lastSetFeaturedValue = featured
	return m.setFeaturedErr
}

func (m *mockLoveService) SetPicked(id uint, picked bool) error {
	m.lastSetPickedID = id
	m.lastSetPickedValue = picked
	return m.setPickedErr
}

func (m *mockLoveService) BatchAudit(ids []uint, auditStatus int, auditReason string) error {
	m.lastBatchAuditIDs = ids
	m.lastBatchAuditStatus = auditStatus
	m.lastBatchAuditReason = auditReason
	return m.batchAuditErr
}

func (m *mockLoveService) BatchUpdateStatus(ids []uint, status int) error {
	m.lastBatchUpdateStatusIDs = ids
	m.lastBatchUpdateStatusStatus = status
	return m.batchUpdateStatusErr
}

// 编译期接口实现校验
var _ service.LoveService = (*mockLoveService)(nil)

// handlerEnv 测试环境：gin 引擎 + mock service。
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockLoveService
}

// newHandlerEnv 构造 gin 引擎并注册 love 主表路由（路径与 love/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉限流/权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockLoveService{
		createResult: &dto.LoveInfo{
			ID: 1, UserID: ctxUserID, Nickname: "小红", Gender: 2, Age: 26, Status: 1, RegionID: regionID,
		},
		getByIDResult: &dto.LoveInfo{
			ID: 1, UserID: 1, Nickname: "小红", Gender: 2, Age: 26, Status: 1,
		},
		getByUserIDResult: &dto.LoveInfo{
			ID: 1, UserID: ctxUserID, Nickname: "小红", Gender: 2, Status: 1,
		},
		listResult: []dto.LoveInfo{
			{ID: 1, Nickname: "小红", Gender: 2, Age: 26, Status: 1},
			{ID: 2, Nickname: "小明", Gender: 1, Age: 28, Status: 1},
		},
		listTotal: 2,
		nearbyResult: []dto.LoveInfo{
			{ID: 1, Nickname: "小红", Gender: 2, Age: 26, Distance: 1.2},
		},
		nearbyTotal: 1,
		searchResult: []dto.LoveInfo{
			{ID: 1, Nickname: "小红", Gender: 2, Age: 26},
		},
		searchTotal: 1,
		advancedResult: []dto.LoveInfo{
			{ID: 1, Nickname: "小红", Gender: 2, Age: 26},
			{ID: 2, Nickname: "小明", Gender: 1, Age: 28},
			{ID: 3, Nickname: "小蓝", Gender: 2, Age: 24},
		},
		advancedTotal: 3,
		matchScoreResult: &dto.LoveMatchScoreResponse{
			TotalScore: 88.5, InterestMatch: 90, PersonalityMatch: 85, Reason: "灵魂伴侣般匹配",
		},
		adminGetByIDResult: &dto.LoveInfo{
			ID: 1, UserID: 1, Nickname: "小红", Status: 1, AuditStatus: 1,
		},
		adminListResult: []dto.LoveInfo{
			{ID: 1, Nickname: "小红", Status: 1, AuditStatus: 1},
			{ID: 2, Nickname: "小明", Status: 1, AuditStatus: 0},
		},
		adminListTotal: 2,
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := loveHandler.NewLoveHandler(mock)
	root := r.Group("/api/v1/love")
	// 公开只读接口
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/advanced-search", h.AdvancedSearch)
	root.GET("/:id", h.GetByID)
	// 需登录接口（C 端）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.GET("/me", h.GetByUserID)
	root.PUT("/:id/location", h.UpdateLocation)
	root.PUT("/:id/voice-intro", h.UpdateVoiceIntro)
	root.GET("/match-score", h.MatchScore)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/loves", h.AdminList)
	admin.GET("/loves/:id", h.AdminGetByID)
	admin.PUT("/loves/:id/audit", h.Audit)
	admin.PUT("/loves/:id/status", h.AdminUpdateStatus)
	admin.PUT("/loves/:id/featured", h.SetFeatured)
	admin.PUT("/loves/:id/picked", h.SetPicked)
	admin.PUT("/loves/batch-audit", h.BatchAudit)
	admin.PUT("/loves/batch-status", h.BatchUpdateStatus)

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

// parsePage 解析响应 data 为 pageData
func parsePage(t *testing.T, resp *apiResponse) *pageData {
	t.Helper()
	var p pageData
	require.NoError(t, json.Unmarshal(resp.Data, &p))
	return &p
}

// validCreateBody 构造一份可通过 Bind 校验的完整 CreateLoveRequest 请求体。
func validCreateBody() map[string]interface{} {
	return map[string]interface{}{
		"nickname":  "小红",
		"avatar":    "https://cdn.example.com/a.png",
		"gender":    2,
		"height":    165,
		"weight":    50,
		"hometown":  "湖北武汉",
		"residence": "武昌",
		"education": "本科",
		"bio":       "热爱生活，期待遇见",
	}
}

// ==================== 公开只读接口（无需登录） ====================

// ---------- GetByID ----------

func TestLoveHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	// 未登录 viewerUserID=0 透传给 service
	assert.Equal(t, uint(0), env.mock.lastGetByIDViewerUserID)
	var info dto.LoveInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "小红", info.Nickname)
}

func TestLoveHandler_GetByID_Success_WithLogin(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	assert.Equal(t, uint(7), env.mock.lastGetByIDViewerUserID)
}

func TestLoveHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestLoveHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("用户资料不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/999", nil)

	// CodeLoveNotFound=5002
	assert.Equal(t, utils.CodeLoveNotFound, resp.Code)
	assert.Equal(t, "用户资料不存在", resp.Message)
}

// ---------- List ----------

func TestLoveHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love?page=1&page_size=10&keyword=小红", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "小红", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.PageSize)
}

func TestLoveHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love", nil)

	// CodeLoveError=5001
	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "查询失败", resp.Message)
}

// ---------- Nearby ----------

func TestLoveHandler_Nearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/nearby?latitude=30.55&longitude=114.30&radius_km=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastNearbyRegionID)
	require.NotNil(t, env.mock.lastNearbyReq)
	assert.Equal(t, 30.55, env.mock.lastNearbyReq.Latitude)
	assert.Equal(t, 114.30, env.mock.lastNearbyReq.Longitude)
	assert.Equal(t, 5.0, env.mock.lastNearbyReq.RadiusKm)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestLoveHandler_Nearby_BindError_MissingLat(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/nearby?longitude=114.30", nil)

	// latitude 为 required
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Nearby_BindError_MissingLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/nearby?latitude=30.55", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Nearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.nearbyResult = nil
	env.mock.nearbyErr = errors.New("空间查询不可用")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/nearby?latitude=30.55&longitude=114.30", nil)

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "空间查询不可用", resp.Message)
}

// ---------- Search ----------

func TestLoveHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/search?keyword=小红", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "小红", env.mock.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestLoveHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("搜索引擎故障")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/search?keyword=小红", nil)

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "搜索引擎故障", resp.Message)
}

// ---------- AdvancedSearch ----------

func TestLoveHandler_AdvancedSearch_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/advanced-search?keyword=小红&min_age=20&max_age=30", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastAdvancedRegionID)
	require.NotNil(t, env.mock.lastAdvancedReq)
	assert.Equal(t, "小红", env.mock.lastAdvancedReq.Keyword)
	assert.Equal(t, 20, env.mock.lastAdvancedReq.MinAge)
	assert.Equal(t, 30, env.mock.lastAdvancedReq.MaxAge)
	p := parsePage(t, resp)
	assert.Equal(t, int64(3), p.Total)
}

func TestLoveHandler_AdvancedSearch_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.advancedResult = nil
	env.mock.advancedErr = errors.New("高级查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/advanced-search", nil)

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "高级查询失败", resp.Message)
}

// ==================== 需登录接口（C 端） ====================

// ---------- Create ----------

func TestLoveHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", validCreateBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	// 上下文透传断言
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	// 请求体透传断言
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "小红", env.mock.lastCreateReq.Nickname)
	assert.Equal(t, 2, env.mock.lastCreateReq.Gender)
	assert.Equal(t, 165, env.mock.lastCreateReq.Height)
	var info dto.LoveInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestLoveHandler_Create_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", validCreateBody())

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestLoveHandler_Create_BindError_MissingNickname(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	delete(body, "nickname") // nickname 为 required
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Create_BindError_InvalidGender(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["gender"] = 9 // oneof=0 1 2 校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Create_BindError_HeightOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["height"] = 999 // max=300 校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Create_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/love", "{invalid json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("已存在资料")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/love", validCreateBody())

	// CodeLoveError=5001
	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "已存在资料", resp.Message)
}

// ---------- Update ----------

func TestLoveHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	newNick := "小红改了昵称"
	body := map[string]interface{}{"nickname": newNick, "bio": "更新了简介"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	require.NotNil(t, env.mock.lastUpdateReq.Nickname)
	assert.Equal(t, newNick, *env.mock.lastUpdateReq.Nickname)
}

func TestLoveHandler_Update_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3", map[string]interface{}{"nickname": "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestLoveHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/abc", map[string]interface{}{"nickname": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_Update_BindError_NicknameTooLong(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := map[string]interface{}{"nickname": strings.Repeat("a", 65)} // max=64
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Update_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/3", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateErr = errors.New("无权操作此资料")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3", map[string]interface{}{"nickname": "x"})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "无权操作此资料", resp.Message)
}

// ---------- GetByUserID（/me） ----------

func TestLoveHandler_GetByUserID_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/me", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastGetByUserID)
	var info dto.LoveInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestLoveHandler_GetByUserID_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/me", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLoveHandler_GetByUserID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.getByUserIDResult = nil
	env.mock.getByUserIDErr = errors.New("用户资料不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/me", nil)

	assert.Equal(t, utils.CodeLoveNotFound, resp.Code)
	assert.Equal(t, "用户资料不存在", resp.Message)
}

// ---------- UpdateLocation ----------

func TestLoveHandler_UpdateLocation_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/location", map[string]interface{}{
		"latitude":  30.55,
		"longitude": 114.30,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateLocationID)
	assert.Equal(t, uint(7), env.mock.lastUpdateLocationUserID)
	require.NotNil(t, env.mock.lastUpdateLocationReq)
	assert.Equal(t, 30.55, env.mock.lastUpdateLocationReq.Latitude)
	assert.Equal(t, 114.30, env.mock.lastUpdateLocationReq.Longitude)
}

func TestLoveHandler_UpdateLocation_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/location", map[string]interface{}{
		"latitude": 30.55, "longitude": 114.30,
	})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLoveHandler_UpdateLocation_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/abc/location", map[string]interface{}{
		"latitude": 30.55, "longitude": 114.30,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_UpdateLocation_BindError_MissingLat(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/location", map[string]interface{}{
		"longitude": 114.30,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_UpdateLocation_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/3/location", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_UpdateLocation_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateLocationErr = errors.New("无权操作此资料")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/location", map[string]interface{}{
		"latitude": 30.55, "longitude": 114.30,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "无权操作此资料", resp.Message)
}

// ---------- UpdateVoiceIntro ----------

func TestLoveHandler_UpdateVoiceIntro_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/voice-intro", map[string]interface{}{
		"voice_intro_url": "https://cdn.example.com/intro.mp3",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateVoiceIntroID)
	assert.Equal(t, uint(7), env.mock.lastUpdateVoiceIntroUserID)
	require.NotNil(t, env.mock.lastUpdateVoiceIntroReq)
	assert.Equal(t, "https://cdn.example.com/intro.mp3", env.mock.lastUpdateVoiceIntroReq.VoiceIntroURL)
}

func TestLoveHandler_UpdateVoiceIntro_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/voice-intro", map[string]interface{}{
		"voice_intro_url": "https://cdn.example.com/intro.mp3",
	})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLoveHandler_UpdateVoiceIntro_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/abc/voice-intro", map[string]interface{}{
		"voice_intro_url": "https://cdn.example.com/intro.mp3",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_UpdateVoiceIntro_BindError_MissingURL(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/voice-intro", map[string]interface{}{})

	// voice_intro_url 为 required
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_UpdateVoiceIntro_BindError_URLTooLong(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/voice-intro", map[string]interface{}{
		"voice_intro_url": strings.Repeat("a", 256), // max=255
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_UpdateVoiceIntro_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/3/voice-intro", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_UpdateVoiceIntro_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateVoiceIntroErr = errors.New("无权操作此资料")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/3/voice-intro", map[string]interface{}{
		"voice_intro_url": "https://cdn.example.com/intro.mp3",
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "无权操作此资料", resp.Message)
}

// ---------- MatchScore ----------

func TestLoveHandler_MatchScore_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/match-score?target_user_id=12", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastMatchScoreUserA)
	assert.Equal(t, uint(12), env.mock.lastMatchScoreUserB)
	var ms dto.LoveMatchScoreResponse
	require.NoError(t, json.Unmarshal(resp.Data, &ms))
	assert.Equal(t, 88.5, ms.TotalScore)
	assert.Equal(t, "灵魂伴侣般匹配", ms.Reason)
}

func TestLoveHandler_MatchScore_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/match-score?target_user_id=12", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestLoveHandler_MatchScore_EmptyTarget(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/match-score", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "target_user_id 不能为空", resp.Message)
}

func TestLoveHandler_MatchScore_InvalidTarget(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/match-score?target_user_id=abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "target_user_id 无效", resp.Message)
}

func TestLoveHandler_MatchScore_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.matchScoreResult = nil
	env.mock.matchScoreErr = errors.New("用户资料不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/match-score?target_user_id=12", nil)

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "用户资料不存在", resp.Message)
}

// ==================== M 端管理后台接口 ====================

// ---------- AdminList ----------

func TestLoveHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/admin/loves?page=1&page_size=10&keyword=小红&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, "小红", env.mock.lastAdminListReq.Keyword)
	require.NotNil(t, env.mock.lastAdminListReq.Status)
	assert.Equal(t, 1, *env.mock.lastAdminListReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
}

func TestLoveHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("管理列表查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/admin/loves", nil)

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "管理列表查询失败", resp.Message)
}

// ---------- AdminGetByID ----------

func TestLoveHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/admin/loves/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.LoveInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestLoveHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/admin/loves/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminGetByIDResult = nil
	env.mock.adminGetByIDErr = errors.New("用户资料不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/love/admin/loves/999", nil)

	assert.Equal(t, utils.CodeLoveNotFound, resp.Code)
	assert.Equal(t, "用户资料不存在", resp.Message)
}

// ---------- Audit ----------

func TestLoveHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/audit", map[string]interface{}{
		"audit_status": 1,
		"audit_reason": "资料合规",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditStatus)
	assert.Equal(t, "资料合规", env.mock.lastAuditReason)
}

func TestLoveHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/abc/audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_Audit_BindError_MissingStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/audit", map[string]interface{}{})

	// audit_status 缺省为 0，required 校验失败
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Audit_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/audit", map[string]interface{}{
		"audit_status": 9, // oneof=1 2 3
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Audit_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.auditErr = errors.New("已审核的资料不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "已审核的资料不能重复审核", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestLoveHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/status", map[string]interface{}{
		"status": 2, // 冻结
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestLoveHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/abc/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_AdminUpdateStatus_BindError_MissingStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/status", map[string]interface{}{})

	// status 缺省为 0，required 校验失败
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_AdminUpdateStatus_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/3/status", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminUpdateStatusErr = errors.New("状态迁移非法")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "状态迁移非法", resp.Message)
}

// ---------- SetFeatured ----------

func TestLoveHandler_SetFeatured_Success_True(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/featured", map[string]interface{}{
		"featured": true,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "设置成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastSetFeaturedID)
	assert.True(t, env.mock.lastSetFeaturedValue)
}

func TestLoveHandler_SetFeatured_Success_False(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/featured", map[string]interface{}{
		"featured": false,
	})

	assert.Equal(t, 0, resp.Code)
	assert.False(t, env.mock.lastSetFeaturedValue)
}

func TestLoveHandler_SetFeatured_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/abc/featured", map[string]interface{}{
		"featured": true,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_SetFeatured_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/3/featured", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_SetFeatured_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.setFeaturedErr = errors.New("精选设置失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/featured", map[string]interface{}{
		"featured": true,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "精选设置失败", resp.Message)
}

// ---------- SetPicked ----------

func TestLoveHandler_SetPicked_Success_True(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/picked", map[string]interface{}{
		"picked": true,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "设置成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastSetPickedID)
	assert.True(t, env.mock.lastSetPickedValue)
}

func TestLoveHandler_SetPicked_Success_False(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/picked", map[string]interface{}{
		"picked": false,
	})

	assert.Equal(t, 0, resp.Code)
	assert.False(t, env.mock.lastSetPickedValue)
}

func TestLoveHandler_SetPicked_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/abc/picked", map[string]interface{}{
		"picked": true,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestLoveHandler_SetPicked_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/3/picked", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_SetPicked_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.setPickedErr = errors.New("推荐设置失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/3/picked", map[string]interface{}{
		"picked": true,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "推荐设置失败", resp.Message)
}

// ---------- BatchAudit ----------

func TestLoveHandler_BatchAudit_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", map[string]interface{}{
		"ids":          []uint{1, 2, 3},
		"audit_status": 1,
		"audit_reason": "批量通过",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量审核成功", resp.Message)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchAuditIDs)
	assert.Equal(t, 1, env.mock.lastBatchAuditStatus)
	assert.Equal(t, "批量通过", env.mock.lastBatchAuditReason)
}

func TestLoveHandler_BatchAudit_BindError_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", map[string]interface{}{
		"ids":          []uint{},
		"audit_status": 1,
	})

	// ids 为 required,min=1
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchAudit_BindError_MissingIDs(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchAudit_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", map[string]interface{}{
		"ids":          []uint{1},
		"audit_status": 9, // oneof=1 2 3
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchAudit_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchAudit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.batchAuditErr = errors.New("批量审核失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-audit", map[string]interface{}{
		"ids":          []uint{1, 2},
		"audit_status": 1,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "批量审核失败", resp.Message)
}

// ---------- BatchUpdateStatus ----------

func TestLoveHandler_BatchUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", map[string]interface{}{
		"ids":    []uint{1, 2, 3},
		"status": 2,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量更新成功", resp.Message)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchUpdateStatusIDs)
	assert.Equal(t, 2, env.mock.lastBatchUpdateStatusStatus)
}

func TestLoveHandler_BatchUpdateStatus_BindError_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", map[string]interface{}{
		"ids":    []uint{},
		"status": 2,
	})

	// ids 为 required,min=1
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchUpdateStatus_BindError_MissingIDs(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchUpdateStatus_BindError_MissingStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", map[string]interface{}{
		"ids": []uint{1},
	})

	// status 缺省为 0，required 校验失败
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchUpdateStatus_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestLoveHandler_BatchUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.batchUpdateStatusErr = errors.New("批量更新失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/love/admin/loves/batch-status", map[string]interface{}{
		"ids":    []uint{1, 2},
		"status": 2,
	})

	assert.Equal(t, utils.CodeLoveError, resp.Code)
	assert.Equal(t, "批量更新失败", resp.Message)
}
