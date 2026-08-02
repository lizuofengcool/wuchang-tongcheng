// Package handler_test 同城114商户黄页主表 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 dh114 Handler（商户主表）全部分支：
//   - 公开接口无需登录（List/GetByID/Search/AdvancedSearch/ListNearby/IncrShare/IncrContact/RecordCall/RecordView/FavStatus）
//   - 用户接口未登录拦截（Create/Update/Delete/ListMine/Fav/Unfav/ListFavs → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/max → 400 "参数错误"）
//   - ListNearby 缺少经纬度（lat/lng 带 required → Bind 失败 → 400 "参数错误"）；radius_km 缺省补 5
//   - Fav/Unfav 语义（"收藏成功" / "已取消收藏"）
//   - service 成功/错误透传（业务码 CodeDh114Error=5301/CodeDh114NotFound=5302/CodeDh114PublishError=5303/
//     CodeDh114AuditError=5304/CodeDh114StatusInvalid=5306/CodeDh114FavoriteError=5319/CodeDh114PhoneCallError=5323）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - RecordCall 客户端 IP 提取（X-Forwarded-For 首段 / X-Real-IP）与 User-Agent 透传
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 house/job/marketing/shop/groupbuy handler 测试同风格。
//
// 注：RecordView handler 依赖 URL :id 参数，生产 plugin.go 将其注册在 POST /views（无 :id），
// 此处测试桩将其挂在 POST /:id/view 以便注入 :id、完整覆盖 handler 逻辑。
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
	"wuchang-tongcheng/internal/modules/dh114/dto"
	dh114Handler "wuchang-tongcheng/internal/modules/dh114/handler"
	"wuchang-tongcheng/internal/modules/dh114/service"
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
	PageSize int             `json:"page_size"`
}

// mockDh114Service 内存 mock，实现 service.Dh114Service 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockDh114Service struct {
	// ===== 调用记录 =====
	// C 端
	lastCreateRegionID    uint
	lastCreateUserID      uint
	lastCreateUserName    string
	lastCreateUserPhone   string
	lastCreateUserAvatar  string
	lastCreateReq         *dto.CreateDh114Request

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateDh114Request

	lastDeleteID         uint
	lastDeleteOperatorID uint

	lastGetByIDID     uint
	lastGetByIDUserID uint

	lastListRegionID uint
	lastListReq      *dto.Dh114ListRequest

	lastListNearbyRegionID uint
	lastListNearbyReq      *dto.Dh114NearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.Dh114SearchRequest

	lastAdvancedSearchRegionID uint
	lastAdvancedSearchReq      *dto.AdvancedSearchRequest

	lastListMineUserID   uint
	lastListMinePage      int
	lastListMinePageSize int

	// 收藏
	lastFavUserID      uint
	lastFavDh114ID     uint
	lastUnfavUserID    uint
	lastUnfavDh114ID   uint
	lastFavStatusUserID  uint
	lastFavStatusDh114ID uint
	lastListFavsUserID   uint
	lastListFavsReq      *dto.FavoriteListRequest

	// 互动
	lastIncrShareID   uint
	lastIncrContactID uint

	lastRecordCallUserID      uint
	lastRecordCallDh114ID     uint
	lastRecordCallPhone       string
	lastRecordCallReq         *dto.PhoneCallRequest
	lastRecordCallIP          string
	lastRecordCallUserAgent   string

	lastRecordViewUserID uint
	lastRecordViewIP     string
	lastRecordViewReq    *dto.Dh114ViewRequest

	// M 端
	lastAdminListReq    *dto.Dh114AdminListRequest
	lastAdminGetByIDID  uint
	lastAuditID         uint
	lastAuditStatus     int
	lastAuditReason     string
	lastBatchAuditReq   *dto.BatchAuditRequest
	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int
	lastUpdatePromotionID       uint
	lastUpdatePromotionReq      *dto.PromotionRequest

	// ===== 返回值预设 =====
	createResult *dto.Dh114Info
	createErr    error

	updateErr error
	deleteErr error

	getByIDResult *dto.Dh114Info
	getByIDErr    error

	listResult []dto.Dh114Info
	listErr    error

	listNearbyResult []dto.Dh114Info
	listNearbyErr    error

	searchResult []dto.Dh114Info
	searchErr    error

	advancedSearchResult []dto.Dh114Info
	advancedSearchErr    error

	listMineResult []dto.Dh114Info
	listMineErr    error

	favResult     *dto.FavResponse
	favErr        error
	unfavResult   *dto.FavResponse
	unfavErr      error
	favStatusResult *dto.FavResponse
	favStatusErr    error
	listFavsResult  []dto.Dh114Info
	listFavsErr     error

	incrShareErr   error
	incrContactErr error

	recordCallResult *dto.CallResponse
	recordCallErr    error
	recordViewErr    error

	adminListResult []dto.Dh114Info
	adminListErr    error

	adminGetByIDResult *dto.Dh114Info
	adminGetByIDErr    error

	auditErr error

	batchAuditResult *dto.BatchResultResponse
	batchAuditErr    error

	adminUpdateStatusErr error
	updatePromotionErr   error
}

// ===== C 端 =====

func (m *mockDh114Service) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateDh114Request) (*dto.Dh114Info, error) {
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

func (m *mockDh114Service) Update(id uint, operatorID uint, req *dto.UpdateDh114Request) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockDh114Service) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockDh114Service) GetByID(id uint, userID uint) (*dto.Dh114Info, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUserID = userID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockDh114Service) List(regionID uint, req *dto.Dh114ListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockDh114Service) ListNearby(regionID uint, req *dto.Dh114NearbyRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastListNearbyRegionID = regionID
	m.lastListNearbyReq = req
	if m.listNearbyErr != nil {
		return nil, nil, m.listNearbyErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listNearbyResult))
	return p, m.listNearbyResult, nil
}

func (m *mockDh114Service) Search(regionID uint, req *dto.Dh114SearchRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.searchResult))
	return p, m.searchResult, nil
}

func (m *mockDh114Service) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastAdvancedSearchRegionID = regionID
	m.lastAdvancedSearchReq = req
	if m.advancedSearchErr != nil {
		return nil, nil, m.advancedSearchErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.advancedSearchResult))
	return p, m.advancedSearchResult, nil
}

func (m *mockDh114Service) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastListMineUserID = userID
	m.lastListMinePage = page
	m.lastListMinePageSize = pageSize
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listMineResult))
	return p, m.listMineResult, nil
}

// ===== 收藏 =====

func (m *mockDh114Service) Fav(userID, dh114ID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavDh114ID = dh114ID
	if m.favErr != nil {
		return nil, m.favErr
	}
	return m.favResult, nil
}

func (m *mockDh114Service) Unfav(userID, dh114ID uint) (*dto.FavResponse, error) {
	m.lastUnfavUserID = userID
	m.lastUnfavDh114ID = dh114ID
	if m.unfavErr != nil {
		return nil, m.unfavErr
	}
	return m.unfavResult, nil
}

func (m *mockDh114Service) FavStatus(userID, dh114ID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusDh114ID = dh114ID
	if m.favStatusErr != nil {
		return nil, m.favStatusErr
	}
	return m.favStatusResult, nil
}

func (m *mockDh114Service) ListFavs(userID uint, req *dto.FavoriteListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastListFavsUserID = userID
	m.lastListFavsReq = req
	if m.listFavsErr != nil {
		return nil, nil, m.listFavsErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listFavsResult))
	return p, m.listFavsResult, nil
}

// ===== 互动 =====

func (m *mockDh114Service) IncrContact(id uint) error {
	m.lastIncrContactID = id
	return m.incrContactErr
}

func (m *mockDh114Service) IncrShare(id uint) error {
	m.lastIncrShareID = id
	return m.incrShareErr
}

func (m *mockDh114Service) RecordCall(userID uint, dh114ID uint, phone string, req *dto.PhoneCallRequest, ip, userAgent string) (*dto.CallResponse, error) {
	m.lastRecordCallUserID = userID
	m.lastRecordCallDh114ID = dh114ID
	m.lastRecordCallPhone = phone
	m.lastRecordCallReq = req
	m.lastRecordCallIP = ip
	m.lastRecordCallUserAgent = userAgent
	if m.recordCallErr != nil {
		return nil, m.recordCallErr
	}
	return m.recordCallResult, nil
}

func (m *mockDh114Service) RecordView(userID uint, ip string, req *dto.Dh114ViewRequest) error {
	m.lastRecordViewUserID = userID
	m.lastRecordViewIP = ip
	m.lastRecordViewReq = req
	return m.recordViewErr
}

// ===== M 端管理 =====

func (m *mockDh114Service) AdminList(req *dto.Dh114AdminListRequest) (*utils.Pagination, []dto.Dh114Info, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.adminListResult))
	return p, m.adminListResult, nil
}

func (m *mockDh114Service) AdminGetByID(id uint) (*dto.Dh114Info, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockDh114Service) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockDh114Service) BatchAudit(req *dto.BatchAuditRequest) (*dto.BatchResultResponse, error) {
	m.lastBatchAuditReq = req
	if m.batchAuditErr != nil {
		return nil, m.batchAuditErr
	}
	return m.batchAuditResult, nil
}

func (m *mockDh114Service) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.adminUpdateStatusErr
}

func (m *mockDh114Service) UpdatePromotion(id uint, req *dto.PromotionRequest) error {
	m.lastUpdatePromotionID = id
	m.lastUpdatePromotionReq = req
	return m.updatePromotionErr
}

// 确保 mockDh114Service 实现 service.Dh114Service 接口
var _ service.Dh114Service = (*mockDh114Service)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockDh114Service
}

// newHandlerEnv 构造 gin 引擎并注册 dh114 商户主表路由（路径与 dh114/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 同时注入 username/phone/avatar 冗余字段，用于 Create 透传断言。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockDh114Service{
		createResult:         &dto.Dh114Info{ID: 1, Title: "沙县小吃", Status: 1, AuditStatus: 1, UserID: ctxUserID, RegionID: regionID, Phone: "13800000000"},
		getByIDResult:        &dto.Dh114Info{ID: 1, Title: "沙县小吃", Status: 1, AuditStatus: 1},
		listResult:           []dto.Dh114Info{{ID: 1, Title: "沙县小吃", Status: 1}, {ID: 2, Title: "兰州拉面", Status: 1}},
		listNearbyResult:     []dto.Dh114Info{{ID: 1, Title: "沙县小吃", Distance: 1.2}},
		searchResult:         []dto.Dh114Info{{ID: 1, Title: "沙县小吃"}},
		advancedSearchResult: []dto.Dh114Info{{ID: 1, Title: "沙县小吃"}},
		listMineResult:       []dto.Dh114Info{{ID: 1, Title: "沙县小吃", UserID: ctxUserID}},
		favResult:            &dto.FavResponse{HasFaved: true, FavCount: 10},
		unfavResult:          &dto.FavResponse{HasFaved: false, FavCount: 9},
		favStatusResult:      &dto.FavResponse{HasFaved: false, FavCount: 9},
		listFavsResult:       []dto.Dh114Info{{ID: 1, Title: "沙县小吃"}},
		recordCallResult:     &dto.CallResponse{CallNo: "CALL123", Phone: "13800000000", CallCount: 5},
		adminListResult:      []dto.Dh114Info{{ID: 1, Title: "沙县小吃", Status: 1}},
		adminGetByIDResult:   &dto.Dh114Info{ID: 1, Title: "沙县小吃", Status: 1},
		batchAuditResult:     &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id（+ 冗余字段）
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Set(middleware.ContextUsername, "张三")
		c.Set(middleware.ContextUserPhone, "13800000000")
		c.Set(middleware.ContextUserAvatar, "https://cdn.example.com/a.png")
		c.Next()
	})

	h := dh114Handler.NewHandler(mock)
	root := r.Group("/api/v1/dh114")
	// 公开接口（C 端浏览 + 互动）
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/advanced-search", h.AdvancedSearch)
	root.GET("/nearby", h.ListNearby)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/fav-status", h.FavStatus)
	root.POST("/:id/share", h.IncrShare)
	root.POST("/:id/contact", h.IncrContact)
	root.POST("/:id/calls", h.RecordCall)
	root.POST("/:id/view", h.RecordView) // 桩路由：注入 :id 以覆盖 handler 逻辑（生产注册在 /views）
	// 需登录接口（C 端发布/收藏）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/mine", h.ListMine)
	root.GET("/favs", h.ListFavs)
	root.POST("/:id/fav", h.Fav)
	root.DELETE("/:id/fav", h.Unfav)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/dh114s", h.AdminList)
	admin.GET("/dh114s/:id", h.AdminGetByID)
	admin.PUT("/dh114s/:id/audit", h.Audit)
	admin.POST("/dh114s/batch-audit", h.BatchAudit)
	admin.PUT("/dh114s/:id/status", h.AdminUpdateStatus)
	admin.PUT("/dh114s/:id/promotion", h.UpdatePromotion)

	return &handlerEnv{engine: r.Engine(), mock: mock}
}

// doJSON 发起 JSON 请求，返回解析后的响应体。
func (e *handlerEnv) doJSON(t *testing.T, method, path string, body interface{}) *apiResponse {
	t.Helper()
	return e.doJSONWithHeaders(t, method, path, body, nil)
}

// doJSONWithHeaders 发起带自定义请求头的 JSON 请求。
func (e *handlerEnv) doJSONWithHeaders(t *testing.T, method, path string, body interface{}, headers map[string]string) *apiResponse {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
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

// validUpdateBody 构造一份可通过 Bind 校验的完整 UpdateDh114Request 请求体。
// UpdateDh114Request 的若干 *string 字段带 max 校验但未加 omitempty，
// nil 指针会触发 max 失败，故此处对全部 max 约束字段填入合法值。
func validUpdateBody() map[string]interface{} {
	return map[string]interface{}{
		"title":            "新标题",
		"cover_image":      "https://cdn.example.com/cover.png",
		"category_name":    "美食",
		"phone":            "13800000000",
		"alt_phone":        "13900000000",
		"website":          "https://example.com",
		"wechat":           "wx123",
		"city":             "武汉",
		"district":         "武昌区",
		"business_district": "中南",
		"address":          "中南路1号",
		"video_url":        "https://cdn.example.com/v.mp4",
		"video_cover":      "https://cdn.example.com/vc.png",
		"vr_url":           "https://cdn.example.com/vr",
	}
}

// ==================== 公开接口（C 端浏览，无需登录） ====================

// ---------- List ----------

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114?page=1&page_size=10&keyword=沙县", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "沙县", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.Dh114Info
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "沙县小吃", list[0].Title)
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.doJSON(t, http.MethodGet, "/api/v1/dh114", nil)

	require.NotNil(t, env.mock.lastListReq)
	// 未传 page/page_size → parsePagination 兜底 1/10
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	assert.Equal(t, 10, env.mock.lastListReq.PageSize)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114", nil)

	// CodeDh114Error=5301
	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 5, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/dh114", nil)

	assert.Equal(t, uint(8), env.mock.lastListRegionID)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	assert.Equal(t, uint(0), env.mock.lastGetByIDUserID)
	var info dto.Dh114Info
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "沙县小吃", info.Title)
}

func TestHandler_GetByID_LoginUserID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.doJSON(t, http.MethodGet, "/api/v1/dh114/3", nil)

	// 登录后 GetByID 透传 userID 用于收藏状态
	assert.Equal(t, uint(7), env.mock.lastGetByIDUserID)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("商户不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/999", nil)

	// CodeDh114NotFound=5302
	assert.Equal(t, utils.CodeDh114NotFound, resp.Code)
	assert.Equal(t, "商户不存在", resp.Message)
}

// ---------- Search ----------

func TestHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/search?keyword=沙县&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "沙县", env.mock.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("es boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/search?keyword=沙县", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "es boom", resp.Message)
}

// ---------- AdvancedSearch ----------

func TestHandler_AdvancedSearch_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/advanced-search?business_type=restaurant&min_rating=4.0", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastAdvancedSearchRegionID)
	require.NotNil(t, env.mock.lastAdvancedSearchReq)
	assert.Equal(t, "restaurant", env.mock.lastAdvancedSearchReq.BusinessType)
	assert.Equal(t, 4.0, env.mock.lastAdvancedSearchReq.MinRating)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_AdvancedSearch_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.advancedSearchResult = nil
	env.mock.advancedSearchErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/advanced-search", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- ListNearby ----------

func TestHandler_ListNearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/nearby?latitude=31.23&longitude=121.47&radius_km=3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListNearbyRegionID)
	require.NotNil(t, env.mock.lastListNearbyReq)
	assert.Equal(t, 31.23, env.mock.lastListNearbyReq.Latitude)
	assert.Equal(t, 121.47, env.mock.lastListNearbyReq.Longitude)
	assert.Equal(t, 3.0, env.mock.lastListNearbyReq.RadiusKm)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListNearby_RadiusDefault(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// radius_km 缺省 → handler 兜底为 5
	env.doJSON(t, http.MethodGet, "/api/v1/dh114/nearby?latitude=31.23&longitude=121.47", nil)

	require.NotNil(t, env.mock.lastListNearbyReq)
	assert.Equal(t, 5.0, env.mock.lastListNearbyReq.RadiusKm)
}

func TestHandler_ListNearby_MissingLatLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude/longitude 为 required，缺失 → Bind 失败 → 400
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/nearby", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_ListNearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listNearbyResult = nil
	env.mock.listNearbyErr = errors.New("postgis error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/nearby?latitude=31.23&longitude=121.47", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "postgis error", resp.Message)
}

// ==================== 需登录接口（C 端发布/收藏） ====================

// ---------- Create ----------

func TestHandler_Create_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114", map[string]interface{}{
		"title": "沙县小吃", "phone": "13800000000",
	})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCreateUserID)
}

func TestHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/dh114", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// title/phone 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114", map[string]interface{}{
		"content": "无标题无电话",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114", map[string]interface{}{
		"title": "沙县小吃", "phone": "13800000000", "business_type": "restaurant",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	// 上下文透传
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "张三", env.mock.lastCreateUserName)
	assert.Equal(t, "13800000000", env.mock.lastCreateUserPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateUserAvatar)
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "沙县小吃", env.mock.lastCreateReq.Title)
	assert.Equal(t, "restaurant", env.mock.lastCreateReq.BusinessType)
	var info dto.Dh114Info
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("标题违规")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114", map[string]interface{}{
		"title": "沙县小吃", "phone": "13800000000",
	})

	// CodeDh114PublishError=5303
	assert.Equal(t, utils.CodeDh114PublishError, resp.Code)
	assert.Equal(t, "标题违规", resp.Message)
}

// ---------- Update ----------

func TestHandler_Update_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/1", map[string]interface{}{"title": "新标题"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/abc", map[string]interface{}{"title": "新标题"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/dh114/1", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/3", validUpdateBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "新标题", *env.mock.lastUpdateReq.Title)
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updateErr = errors.New("无权操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/3", validUpdateBody())

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "无权操作", resp.Message)
}

// ---------- Delete ----------

func TestHandler_Delete_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/1", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.deleteErr = errors.New("无权删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/3", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "无权删除", resp.Message)
}

// ---------- ListMine ----------

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/mine?page=2&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 20, env.mock.lastListMinePageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listMineResult = nil
	env.mock.listMineErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/mine", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- Fav ----------

func TestHandler_Fav_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/1/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Fav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Fav_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastFavUserID)
	assert.Equal(t, uint(3), env.mock.lastFavDh114ID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.True(t, fr.HasFaved)
	assert.Equal(t, 10, fr.FavCount)
}

func TestHandler_Fav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.favResult = nil
	env.mock.favErr = errors.New("已收藏过")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/fav", nil)

	// CodeDh114FavoriteError=5319
	assert.Equal(t, utils.CodeDh114FavoriteError, resp.Code)
	assert.Equal(t, "已收藏过", resp.Message)
}

// ---------- Unfav ----------

func TestHandler_Unfav_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/1/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Unfav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Unfav_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastUnfavUserID)
	assert.Equal(t, uint(3), env.mock.lastUnfavDh114ID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.HasFaved)
}

func TestHandler_Unfav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.unfavResult = nil
	env.mock.unfavErr = errors.New("未收藏")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/dh114/3/fav", nil)

	assert.Equal(t, utils.CodeDh114FavoriteError, resp.Code)
	assert.Equal(t, "未收藏", resp.Message)
}

// ---------- FavStatus（公开，未登录返回 has_faved=false） ----------

func TestHandler_FavStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/3/fav-status", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastFavStatusUserID)
	assert.Equal(t, uint(3), env.mock.lastFavStatusDh114ID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.HasFaved)
}

func TestHandler_FavStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/abc/fav-status", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_FavStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.favStatusResult = nil
	env.mock.favStatusErr = errors.New("商户不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/999/fav-status", nil)

	assert.Equal(t, utils.CodeDh114NotFound, resp.Code)
	assert.Equal(t, "商户不存在", resp.Message)
}

// ---------- ListFavs ----------

func TestHandler_ListFavs_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/favs", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListFavs_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/favs?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListFavsUserID)
	require.NotNil(t, env.mock.lastListFavsReq)
	assert.Equal(t, 1, env.mock.lastListFavsReq.Page)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListFavs_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listFavsResult = nil
	env.mock.listFavsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/favs", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== 互动（公开） ====================

// ---------- IncrShare ----------

func TestHandler_IncrShare_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/abc/share", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrShare_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/share", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录分享", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrShareID)
}

func TestHandler_IncrShare_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrShareErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/share", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- IncrContact ----------

func TestHandler_IncrContact_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/abc/contact", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrContact_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/contact", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录联系", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrContactID)
}

func TestHandler_IncrContact_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrContactErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/contact", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- RecordCall ----------

func TestHandler_RecordCall_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/abc/calls", map[string]interface{}{
		"dh114_id": 1, "call_type": "click",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_RecordCall_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// call_type 非法 → oneof 校验失败 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/calls", map[string]interface{}{
		"dh114_id": 3, "call_type": "invalid",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RecordCall_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/dh114/3/calls", map[string]interface{}{
		"dh114_id": 3, "call_type": "click",
	}, map[string]string{
		"X-Forwarded-For": "1.2.3.4, 5.6.7.8",
		"User-Agent":      "Mozilla/5.0",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastRecordCallDh114ID)
	assert.Equal(t, uint(7), env.mock.lastRecordCallUserID)
	require.NotNil(t, env.mock.lastRecordCallReq)
	assert.Equal(t, "click", env.mock.lastRecordCallReq.CallType)
	// X-Forwarded-For 首段作为客户端 IP
	assert.Equal(t, "1.2.3.4", env.mock.lastRecordCallIP)
	assert.Equal(t, "Mozilla/5.0", env.mock.lastRecordCallUserAgent)
	var cr dto.CallResponse
	require.NoError(t, json.Unmarshal(resp.Data, &cr))
	assert.Equal(t, "CALL123", cr.CallNo)
	assert.Equal(t, 5, cr.CallCount)
}

func TestHandler_RecordCall_RealIP(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// 无 X-Forwarded-For → 取 X-Real-IP
	env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/dh114/3/calls", map[string]interface{}{
		"dh114_id": 3, "call_type": "call",
	}, map[string]string{
		"X-Real-IP": "9.9.9.9",
	})

	assert.Equal(t, "9.9.9.9", env.mock.lastRecordCallIP)
}

func TestHandler_RecordCall_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.recordCallResult = nil
	env.mock.recordCallErr = errors.New("电话记录失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/calls", map[string]interface{}{
		"dh114_id": 3, "call_type": "click",
	})

	// CodeDh114PhoneCallError=5323
	assert.Equal(t, utils.CodeDh114PhoneCallError, resp.Code)
	assert.Equal(t, "电话记录失败", resp.Message)
}

// ---------- RecordView ----------

func TestHandler_RecordView_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/view", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录浏览", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastRecordViewUserID)
	require.NotNil(t, env.mock.lastRecordViewReq)
	assert.Equal(t, uint(3), env.mock.lastRecordViewReq.Dh114ID)
	assert.Equal(t, "business", env.mock.lastRecordViewReq.VisitType)
}

func TestHandler_RecordView_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.recordViewErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/3/view", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== M 端管理 ====================

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/admin/dh114s?page=1&page_size=10&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, 1, env.mock.lastAdminListReq.Page)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.Dh114Info
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/admin/dh114s", nil)

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- AdminGetByID ----------

func TestHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/admin/dh114s/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.Dh114Info
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "沙县小吃", info.Title)
}

func TestHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/admin/dh114s/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminGetByIDResult = nil
	env.mock.adminGetByIDErr = errors.New("商户不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/dh114/admin/dh114s/999", nil)

	assert.Equal(t, utils.CodeDh114NotFound, resp.Code)
	assert.Equal(t, "商户不存在", resp.Message)
}

// ---------- Audit ----------

func TestHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/abc/audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Audit_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status 非法（oneof=0 1 2）→ Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/audit", map[string]interface{}{
		"audit_status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/audit", map[string]interface{}{
		"audit_status": 1, "audit_reason": "内容合规",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditStatus)
	assert.Equal(t, "内容合规", env.mock.lastAuditReason)
}

func TestHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.auditErr = errors.New("已审核不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/audit", map[string]interface{}{
		"audit_status": 1,
	})

	// CodeDh114AuditError=5304
	assert.Equal(t, utils.CodeDh114AuditError, resp.Code)
	assert.Equal(t, "已审核不能重复审核", resp.Message)
}

// ---------- BatchAudit ----------

func TestHandler_BatchAudit_BindFail_MissingIDs(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// ids 为 required,min=1，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/admin/dh114s/batch-audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchAudit_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status 非法（oneof=1 2）→ Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/admin/dh114s/batch-audit", map[string]interface{}{
		"ids": []uint{1, 2}, "audit_status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchAudit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/admin/dh114s/batch-audit", map[string]interface{}{
		"ids": []uint{1, 2}, "audit_status": 1, "audit_reason": "批量通过",
	})

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastBatchAuditReq)
	assert.Equal(t, []uint{1, 2}, env.mock.lastBatchAuditReq.IDs)
	assert.Equal(t, 1, env.mock.lastBatchAuditReq.AuditStatus)
	assert.Equal(t, "批量通过", env.mock.lastBatchAuditReq.AuditReason)
	var br dto.BatchResultResponse
	require.NoError(t, json.Unmarshal(resp.Data, &br))
	assert.Equal(t, 2, br.Total)
	assert.Equal(t, 2, br.Success)
}

func TestHandler_BatchAudit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.batchAuditResult = nil
	env.mock.batchAuditErr = errors.New("批量审核失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/dh114/admin/dh114s/batch-audit", map[string]interface{}{
		"ids": []uint{1}, "audit_status": 2,
	})

	assert.Equal(t, utils.CodeDh114AuditError, resp.Code)
	assert.Equal(t, "批量审核失败", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/abc/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// status 非法（oneof=1 2 3 4）→ Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/status", map[string]interface{}{
		"status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/status", map[string]interface{}{
		"status": 2, // 强制下架
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态已更新", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.adminUpdateStatusErr = errors.New("状态非法")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/status", map[string]interface{}{
		"status": 2,
	})

	// CodeDh114StatusInvalid=5306
	assert.Equal(t, utils.CodeDh114StatusInvalid, resp.Code)
	assert.Equal(t, "状态非法", resp.Message)
}

// ---------- UpdatePromotion ----------

func TestHandler_UpdatePromotion_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/abc/promotion", map[string]interface{}{
		"promotion_level": 5,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_UpdatePromotion_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// promotion_level 越界（max=10）→ Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/promotion", map[string]interface{}{
		"promotion_level": 11,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	featured := true
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/promotion", map[string]interface{}{
		"featured": featured, "promotion_level": 5, "traffic_weight": 1.5,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "推广配置已更新", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdatePromotionID)
	require.NotNil(t, env.mock.lastUpdatePromotionReq)
	require.NotNil(t, env.mock.lastUpdatePromotionReq.PromotionLevel)
	assert.Equal(t, 5, *env.mock.lastUpdatePromotionReq.PromotionLevel)
}

func TestHandler_UpdatePromotion_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updatePromotionErr = errors.New("推广更新失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/dh114/admin/dh114s/3/promotion", map[string]interface{}{
		"promotion_level": 5,
	})

	assert.Equal(t, utils.CodeDh114Error, resp.Code)
	assert.Equal(t, "推广更新失败", resp.Message)
}
