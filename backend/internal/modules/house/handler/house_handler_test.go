// Package handler_test 房屋租售模块房源主表 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 HouseHandler 全部分支：
//   - 公开接口无需登录（List/GetByID/Search/AdvancedSearch/ListNearby/ListSimilar/FavStatus/IncrContactCount/IncrShareCount）
//   - 用户接口未登录拦截（Create/Update/Delete/ListMine/Fav/ListFavs → 401）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/gte → 400 "参数错误"）
//   - ListNearby 经纬度越界（lat<-90||lat>90||lng<-180||lng>180 → 400 "经纬度参数无效"）
//   - Search 关键词为空（→ 400 "关键词不能为空"）
//   - Fav toggle 语义（HasFaved=true→"收藏成功"，HasFaved=false→"已取消收藏"）
//   - service 成功/错误透传（业务码 CodeHouseError=2901/CodeHouseNotFound=2902/CodeHousePublishError=2903/CodeHouseAuditError=2904 + message + data 透传）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 groupbuy/category/region/news/file/setting/permission handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/house/dto"
	houseHandler "wuchang-tongcheng/internal/modules/house/handler"
	"wuchang-tongcheng/internal/modules/house/service"
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

// mockHouseService 内存 mock，实现 service.HouseService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockHouseService struct {
	// ===== 调用记录 =====
	// C 端
	lastCreateRegionID    uint
	lastCreateUserID       uint
	lastCreateUserName     string
	lastCreateUserPhone    string
	lastCreateUserAvatar   string
	lastCreateReq          *dto.CreateHouseRequest

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateHouseRequest

	lastDeleteID         uint
	lastDeleteOperatorID uint

	lastGetByIDID     uint
	lastGetByIDUserID uint

	lastListRegionID uint
	lastListReq      *dto.HouseListRequest

	lastListNearbyRegionID uint
	lastListNearbyReq      *dto.HouseNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.HouseSearchRequest

	lastAdvancedSearchRegionID uint
	lastAdvancedSearchReq      *dto.HouseAdvancedSearchRequest

	lastListMineUserID   uint
	lastListMinePage      int
	lastListMinePageSize int

	// 收藏
	lastFavUserID   uint
	lastFavHouseID  uint
	lastFavStatusUserID  uint
	lastFavStatusHouseID uint
	lastListFavsUserID   uint
	lastListFavsPage      int
	lastListFavsPageSize int

	// 互动
	lastIncrContactID uint
	lastIncrShareID   uint

	// 相似推荐
	lastListSimilarID    uint
	lastListSimilarLimit int

	// M 端
	lastAdminListReq *dto.HouseAdminListRequest

	lastAdminGetByIDID uint

	lastAuditID          uint
	lastAuditStatus      int
	lastAuditReason      string

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	lastUpdatePromotionID  uint
	lastUpdatePromotionReq *dto.HousePromotionRequest

	// 批量
	lastBatchAuditIDs     []uint
	lastBatchAuditStatus  int
	lastBatchAuditReason  string

	lastBatchUpdateStatusIDs     []uint
	lastBatchUpdateStatusStatus int

	lastBatchDeleteIDs []uint

	// ===== 返回值预设 =====
	createResult  *dto.HouseInfo
	createErr     error

	updateErr error
	deleteErr error

	getByIDResult *dto.HouseDetailResponse
	getByIDErr    error

	listResult []dto.HouseInfo
	listErr    error

	listNearbyResult []dto.HouseInfo
	listNearbyErr    error

	searchResult []dto.HouseInfo
	searchErr    error

	advancedSearchResult []dto.HouseInfo
	advancedSearchErr    error

	listMineResult []dto.HouseInfo
	listMineErr    error

	favResult *dto.FavResponse
	favErr    error

	favStatusResult *dto.FavResponse
	favStatusErr    error

	listFavsResult []dto.HouseInfo
	listFavsErr    error

	incrContactErr error
	incrShareErr   error

	listSimilarResult []dto.SimilarHouseResponse
	listSimilarErr    error

	adminListResult []dto.HouseInfo
	adminListErr    error

	adminGetByIDResult *dto.HouseDetailResponse
	adminGetByIDErr    error

	auditErr            error
	adminUpdateStatusErr error
	updatePromotionErr  error

	batchAuditResult *dto.BatchResultResponse
	batchAuditErr    error

	batchUpdateStatusResult *dto.BatchResultResponse
	batchUpdateStatusErr    error

	batchDeleteResult *dto.BatchResultResponse
	batchDeleteErr    error
}

// ===== C 端 =====

func (m *mockHouseService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateHouseRequest) (*dto.HouseInfo, error) {
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

func (m *mockHouseService) Update(id uint, operatorID uint, req *dto.UpdateHouseRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockHouseService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockHouseService) GetByID(id uint, userID uint) (*dto.HouseDetailResponse, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUserID = userID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockHouseService) List(regionID uint, req *dto.HouseListRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockHouseService) ListNearby(regionID uint, req *dto.HouseNearbyRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastListNearbyRegionID = regionID
	m.lastListNearbyReq = req
	if m.listNearbyErr != nil {
		return nil, nil, m.listNearbyErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listNearbyResult))
	return p, m.listNearbyResult, nil
}

func (m *mockHouseService) Search(regionID uint, req *dto.HouseSearchRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.searchResult))
	return p, m.searchResult, nil
}

func (m *mockHouseService) AdvancedSearch(regionID uint, req *dto.HouseAdvancedSearchRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastAdvancedSearchRegionID = regionID
	m.lastAdvancedSearchReq = req
	if m.advancedSearchErr != nil {
		return nil, nil, m.advancedSearchErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.advancedSearchResult))
	return p, m.advancedSearchResult, nil
}

func (m *mockHouseService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error) {
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

func (m *mockHouseService) Fav(userID, houseID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavHouseID = houseID
	if m.favErr != nil {
		return nil, m.favErr
	}
	return m.favResult, nil
}

func (m *mockHouseService) FavStatus(userID, houseID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusHouseID = houseID
	if m.favStatusErr != nil {
		return nil, m.favStatusErr
	}
	return m.favStatusResult, nil
}

func (m *mockHouseService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastListFavsUserID = userID
	m.lastListFavsPage = page
	m.lastListFavsPageSize = pageSize
	if m.listFavsErr != nil {
		return nil, nil, m.listFavsErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listFavsResult))
	return p, m.listFavsResult, nil
}

// ===== 互动 =====

func (m *mockHouseService) IncrContactCount(id uint) error {
	m.lastIncrContactID = id
	return m.incrContactErr
}

func (m *mockHouseService) IncrShareCount(id uint) error {
	m.lastIncrShareID = id
	return m.incrShareErr
}

// ===== 相似推荐 =====

func (m *mockHouseService) ListSimilar(houseID uint, limit int) ([]dto.SimilarHouseResponse, error) {
	m.lastListSimilarID = houseID
	m.lastListSimilarLimit = limit
	if m.listSimilarErr != nil {
		return nil, m.listSimilarErr
	}
	return m.listSimilarResult, nil
}

// ===== M 端管理 =====

func (m *mockHouseService) AdminList(req *dto.HouseAdminListRequest) (*utils.Pagination, []dto.HouseInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.adminListResult))
	return p, m.adminListResult, nil
}

func (m *mockHouseService) AdminGetByID(id uint) (*dto.HouseDetailResponse, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockHouseService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockHouseService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.adminUpdateStatusErr
}

func (m *mockHouseService) UpdatePromotion(id uint, req *dto.HousePromotionRequest) error {
	m.lastUpdatePromotionID = id
	m.lastUpdatePromotionReq = req
	return m.updatePromotionErr
}

// ===== 批量操作 =====

func (m *mockHouseService) BatchAudit(ids []uint, auditStatus int, auditReason string) (*dto.BatchResultResponse, error) {
	m.lastBatchAuditIDs = ids
	m.lastBatchAuditStatus = auditStatus
	m.lastBatchAuditReason = auditReason
	if m.batchAuditErr != nil {
		return nil, m.batchAuditErr
	}
	return m.batchAuditResult, nil
}

func (m *mockHouseService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	m.lastBatchUpdateStatusIDs = ids
	m.lastBatchUpdateStatusStatus = status
	if m.batchUpdateStatusErr != nil {
		return nil, m.batchUpdateStatusErr
	}
	return m.batchUpdateStatusResult, nil
}

func (m *mockHouseService) BatchDelete(ids []uint) (*dto.BatchResultResponse, error) {
	m.lastBatchDeleteIDs = ids
	if m.batchDeleteErr != nil {
		return nil, m.batchDeleteErr
	}
	return m.batchDeleteResult, nil
}

// 确保 mockHouseService 实现 service.HouseService 接口
var _ service.HouseService = (*mockHouseService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockHouseService
}

// newHandlerEnv 构造 gin 引擎并注册 house 房源主表路由（路径与 house/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 同时注入 username/phone/avatar 冗余字段，用于 Create 透传断言。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockHouseService{
		createResult: &dto.HouseInfo{ID: 1, Title: "精装两室一厅", Status: 1, AuditStatus: 1, UserID: ctxUserID, RegionID: regionID, ListingType: "rent"},
		getByIDResult: &dto.HouseDetailResponse{HouseInfo: dto.HouseInfo{ID: 1, Title: "精装两室一厅", Status: 1, AuditStatus: 1}},
		listResult:    []dto.HouseInfo{{ID: 1, Title: "精装两室一厅", Status: 1}, {ID: 2, Title: "南北通透三室", Status: 1}},
		listNearbyResult: []dto.HouseInfo{{ID: 1, Title: "精装两室一厅", Distance: 1.2}},
		searchResult:     []dto.HouseInfo{{ID: 1, Title: "精装两室一厅"}},
		advancedSearchResult: []dto.HouseInfo{{ID: 1, Title: "精装两室一厅"}},
		listMineResult:   []dto.HouseInfo{{ID: 1, Title: "精装两室一厅", UserID: ctxUserID}},
		favResult:        &dto.FavResponse{HasFaved: true, FavCount: 10},
		favStatusResult:  &dto.FavResponse{HasFaved: false, FavCount: 9},
		listFavsResult:   []dto.HouseInfo{{ID: 1, Title: "精装两室一厅"}},
		listSimilarResult: []dto.SimilarHouseResponse{{HouseID: 2, Title: "相似房源", Similarity: 0.8}},
		adminListResult:  []dto.HouseInfo{{ID: 1, Title: "精装两室一厅", Status: 1}},
		adminGetByIDResult: &dto.HouseDetailResponse{HouseInfo: dto.HouseInfo{ID: 1, Title: "精装两室一厅", Status: 1}},
		batchAuditResult:         &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
		batchUpdateStatusResult:  &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
		batchDeleteResult:         &dto.BatchResultResponse{Total: 2, Success: 2, Failed: 0},
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

	h := houseHandler.NewHouseHandler(mock)
	// 注册路由，路径与 house/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/house")
	// 公开接口（C 端浏览）
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/advanced-search", h.AdvancedSearch)
	root.GET("/nearby", h.ListNearby)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/similar", h.ListSimilar)
	root.GET("/:id/fav", h.FavStatus)
	root.POST("/:id/contact", h.IncrContactCount)
	root.POST("/:id/share", h.IncrShareCount)
	// 需登录接口（C 端发布/收藏/互动/推广）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/mine", h.ListMine)
	root.GET("/favorites", h.ListFavs)
	root.POST("/:id/fav", h.Fav)
	root.PUT("/:id/promotion", h.UpdatePromotion)
	// 管理后台接口（/admin 组）
	admin := root.Group("/admin")
	admin.GET("/list", h.AdminList)
	admin.GET("/:id", h.AdminGetByID)
	admin.PUT("/:id/audit", h.Audit)
	admin.PUT("/:id/status", h.AdminUpdateStatus)
	admin.POST("/batch/audit", h.BatchAudit)
	admin.POST("/batch/status", h.BatchUpdateStatus)
	admin.POST("/batch/delete", h.BatchDelete)

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

// ==================== 公开接口（C 端浏览，无需登录） ====================

// ---------- List ----------

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house?page=1&page_size=10&keyword=精装", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "精装", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.HouseInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "精装两室一厅", list[0].Title)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house", nil)

	// CodeHouseError=2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 5, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/house", nil)

	assert.Equal(t, uint(8), env.mock.lastListRegionID)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	assert.Equal(t, uint(0), env.mock.lastGetByIDUserID)
	var info dto.HouseInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "精装两室一厅", info.Title)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("房源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/999", nil)

	// CodeHouseNotFound=2902
	assert.Equal(t, 2902, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- Search ----------

func TestHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/search?keyword=精装&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "精装", env.mock.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Search_EmptyKeyword(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// keyword 为空串：HouseSearchRequest.Keyword 带 binding:"required"，
	// 空串触发 required 校验失败 → Bind 失败 → 400 "参数错误: ..."
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/search?keyword=", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("es boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/search?keyword=精装", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "es boom", resp.Message)
}

// ---------- AdvancedSearch ----------

func TestHandler_AdvancedSearch_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/advanced-search?listing_type=rent&min_rent_price=1000", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastAdvancedSearchRegionID)
	require.NotNil(t, env.mock.lastAdvancedSearchReq)
	assert.Equal(t, "rent", env.mock.lastAdvancedSearchReq.ListingType)
	assert.Equal(t, 1000.0, env.mock.lastAdvancedSearchReq.MinRentPrice)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_AdvancedSearch_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.advancedSearchResult = nil
	env.mock.advancedSearchErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/advanced-search", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- ListNearby ----------

func TestHandler_ListNearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/nearby?latitude=31.23&longitude=121.47&radius_km=3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListNearbyRegionID)
	require.NotNil(t, env.mock.lastListNearbyReq)
	assert.Equal(t, 31.23, env.mock.lastListNearbyReq.Latitude)
	assert.Equal(t, 121.47, env.mock.lastListNearbyReq.Longitude)
	assert.Equal(t, 3.0, env.mock.lastListNearbyReq.RadiusKm)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListNearby_MissingLatLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude/longitude 为 required，缺失 → Bind 失败 → 400
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/nearby", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_ListNearby_InvalidLatLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude 越界（>90）→ handler 校验返回"经纬度参数无效"
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/nearby?latitude=999&longitude=121.47", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestHandler_ListNearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listNearbyResult = nil
	env.mock.listNearbyErr = errors.New("postgis down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/nearby?latitude=31.23&longitude=121.47", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "postgis down", resp.Message)
}

// ---------- ListSimilar ----------

func TestHandler_ListSimilar_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/3/similar?limit=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastListSimilarID)
	assert.Equal(t, 5, env.mock.lastListSimilarLimit)
	var list []dto.SimilarHouseResponse
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, uint(2), list[0].HouseID)
}

func TestHandler_ListSimilar_DefaultLimit(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// 不传 limit，handler 用 DefaultQuery("limit","5") → limit=5
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/3/similar", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, 5, env.mock.lastListSimilarLimit)
}

func TestHandler_ListSimilar_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/abc/similar", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_ListSimilar_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listSimilarResult = nil
	env.mock.listSimilarErr = errors.New("no similar")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/3/similar", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "no similar", resp.Message)
}

// ---------- FavStatus（公开，未登录返回 has_faved=false） ----------

func TestHandler_FavStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastFavStatusUserID)
	assert.Equal(t, uint(3), env.mock.lastFavStatusHouseID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.HasFaved)
}

func TestHandler_FavStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_FavStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.favStatusResult = nil
	env.mock.favStatusErr = errors.New("房源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/9/fav", nil)

	assert.Equal(t, 2902, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- IncrContactCount（公开） ----------

func TestHandler_IncrContactCount_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/contact", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastIncrContactID)
}

func TestHandler_IncrContactCount_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/abc/contact", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrContactCount_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrContactErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/contact", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- IncrShareCount（公开） ----------

func TestHandler_IncrShareCount_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/share", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastIncrShareID)
}

func TestHandler_IncrShareCount_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/abc/share", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrShareCount_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrShareErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/share", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- 公开读取无需登录聚合 ----------

func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	// userID=0 时多个公开读路径均不被 401 拦截
	env := newHandlerEnv(t, 0, 2)

	r1 := env.doJSON(t, http.MethodGet, "/api/v1/house", nil)
	assert.Equal(t, 0, r1.Code)

	r2 := env.doJSON(t, http.MethodGet, "/api/v1/house/1", nil)
	assert.Equal(t, 0, r2.Code)

	r3 := env.doJSON(t, http.MethodGet, "/api/v1/house/search?keyword=x", nil)
	assert.Equal(t, 0, r3.Code)

	r4 := env.doJSON(t, http.MethodGet, "/api/v1/house/1/similar", nil)
	assert.Equal(t, 0, r4.Code)

	r5 := env.doJSON(t, http.MethodGet, "/api/v1/house/1/fav", nil)
	assert.Equal(t, 0, r5.Code)

	r6 := env.doJSON(t, http.MethodPost, "/api/v1/house/1/contact", nil)
	assert.Equal(t, 0, r6.Code)

	r7 := env.doJSON(t, http.MethodPost, "/api/v1/house/1/share", nil)
	assert.Equal(t, 0, r7.Code)
}

// ==================== 需登录接口（C 端发布/收藏/互动/推广） ====================

// ---------- Create ----------

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.CreateHouseRequest{Title: "精装两室一厅", ListingType: "rent", RentPrice: 3000, BuildingArea: 80, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "张三", env.mock.lastCreateUserName)
	assert.Equal(t, "13800000000", env.mock.lastCreateUserPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateUserAvatar)
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "精装两室一厅", env.mock.lastCreateReq.Title)
	assert.Equal(t, "rent", env.mock.lastCreateReq.ListingType)
	var info dto.HouseInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	body := dto.CreateHouseRequest{Title: "精装两室一厅"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCreateUserID)
}

func TestHandler_Create_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 非法 JSON → Bind 失败
	resp := env.doRaw(t, http.MethodPost, "/api/v1/house", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// title 为空（required）→ binding 校验失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house", map[string]interface{}{"title": "", "status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("发布失败")
	body := dto.CreateHouseRequest{Title: "精装两室一厅", Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house", body)

	// CodeHousePublishError=2903
	assert.Equal(t, 2903, resp.Code)
	assert.Equal(t, "发布失败", resp.Message)
}

// ---------- Update ----------

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.UpdateHouseRequest{Title: "更新后的标题", RentPrice: 3500}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "更新后的标题", env.mock.lastUpdateReq.Title)
	assert.Equal(t, 3500.0, env.mock.lastUpdateReq.RentPrice)
}

func TestHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	body := dto.UpdateHouseRequest{Title: "x"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/3", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.UpdateHouseRequest{Title: "x"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/abc", body)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/house/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updateErr = errors.New("无权操作此房源")
	body := dto.UpdateHouseRequest{Title: "x"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/3", body)

	// CodeHouseError=2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "无权操作此房源", resp.Message)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/house/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/house/3", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/house/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.deleteErr = errors.New("房源不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/house/9", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- ListMine ----------

func TestHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/mine?page=2&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 5, env.mock.lastListMinePageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListMine_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListMineUserID)
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listMineResult = nil
	env.mock.listMineErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/mine", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- Fav（toggle） ----------

func TestHandler_Fav_Success_Faved(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// mock 默认 HasFaved=true → "收藏成功"
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastFavUserID)
	assert.Equal(t, uint(3), env.mock.lastFavHouseID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.True(t, fr.HasFaved)
}

func TestHandler_Fav_Success_UnFaved(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.favResult = &dto.FavResponse{HasFaved: false, FavCount: 8}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.HasFaved)
}

func TestHandler_Fav_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/3/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastFavUserID)
}

func TestHandler_Fav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Fav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.favResult = nil
	env.mock.favErr = errors.New("房源不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/9/fav", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- ListFavs ----------

func TestHandler_ListFavs_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/favorites?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListFavsUserID)
	assert.Equal(t, 1, env.mock.lastListFavsPage)
	assert.Equal(t, 10, env.mock.lastListFavsPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListFavs_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/favorites", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListFavsUserID)
}

func TestHandler_ListFavs_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listFavsResult = nil
	env.mock.listFavsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/favorites", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- UpdatePromotion（M 端推广配置） ----------

func TestHandler_UpdatePromotion_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.HousePromotionRequest{PromotionLevel: 5, TrafficWeight: 1.5, Featured: true}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/3/promotion", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "推广配置更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdatePromotionID)
	require.NotNil(t, env.mock.lastUpdatePromotionReq)
	assert.Equal(t, 5, env.mock.lastUpdatePromotionReq.PromotionLevel)
	assert.Equal(t, 1.5, env.mock.lastUpdatePromotionReq.TrafficWeight)
	assert.True(t, env.mock.lastUpdatePromotionReq.Featured)
}

func TestHandler_UpdatePromotion_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := dto.HousePromotionRequest{PromotionLevel: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/abc/promotion", body)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_UpdatePromotion_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/house/3/promotion", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// promotion_level 越界（>10，binding gte=0,lte=10）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/3/promotion", map[string]interface{}{"promotion_level": 99})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.updatePromotionErr = errors.New("房源不存在")
	body := dto.HousePromotionRequest{PromotionLevel: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/9/promotion", body)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ==================== 管理后台接口（/admin 组） ====================

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/admin/list?page=1&page_size=20&status=1&keyword=精装", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, "精装", env.mock.lastAdminListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/admin/list", nil)

	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- AdminGetByID ----------

func TestHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/admin/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.HouseInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/admin/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.adminGetByIDResult = nil
	env.mock.adminGetByIDErr = errors.New("房源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/house/admin/9", nil)

	assert.Equal(t, 2902, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- Audit ----------

func TestHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.AuditRequest{AuditStatus: 1, AuditReason: "审核通过"}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/3/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditStatus)
	assert.Equal(t, "审核通过", env.mock.lastAuditReason)
}

func TestHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.AuditRequest{AuditStatus: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/abc/audit", body)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Audit_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/house/admin/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// audit_status 不在 oneof=0 1 2（传 9）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/3/audit", map[string]interface{}{"audit_status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.auditErr = errors.New("已审核的房源不能重复审核")
	body := dto.AuditRequest{AuditStatus: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/3/audit", body)

	// CodeHouseAuditError=2904
	assert.Equal(t, 2904, resp.Code)
	assert.Equal(t, "已审核的房源不能重复审核", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.AdminUpdateStatusRequest{Status: 2} // 2 下架
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/3/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.AdminUpdateStatusRequest{Status: 2}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/abc/status", body)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/house/admin/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// status 不在 oneof=1 2 3 4（传 9）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/3/status", map[string]interface{}{"status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.adminUpdateStatusErr = errors.New("房源不存在")
	body := dto.AdminUpdateStatusRequest{Status: 2}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/house/admin/9/status", body)

	assert.Equal(t, 2904, resp.Code)
	assert.Equal(t, "房源不存在", resp.Message)
}

// ---------- BatchAudit ----------

func TestHandler_BatchAudit_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.BatchAuditRequest{IDs: []uint{1, 2}, AuditStatus: 1, AuditReason: "批量通过"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量审核完成", resp.Message)
	assert.Equal(t, []uint{1, 2}, env.mock.lastBatchAuditIDs)
	assert.Equal(t, 1, env.mock.lastBatchAuditStatus)
	assert.Equal(t, "批量通过", env.mock.lastBatchAuditReason)
	var br dto.BatchResultResponse
	require.NoError(t, json.Unmarshal(resp.Data, &br))
	assert.Equal(t, 2, br.Total)
	assert.Equal(t, 2, br.Success)
}

func TestHandler_BatchAudit_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/house/admin/batch/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchAudit_ValidationFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// ids 为空（required,min=1）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/audit", map[string]interface{}{"ids": []uint{}, "audit_status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchAudit_ValidationFail_BadStatus(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// audit_status 不在 oneof=1 2（传 9）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/audit", map[string]interface{}{"ids": []uint{1}, "audit_status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchAudit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.batchAuditResult = nil
	env.mock.batchAuditErr = errors.New("batch db error")
	body := dto.BatchAuditRequest{IDs: []uint{1}, AuditStatus: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/audit", body)

	assert.Equal(t, 2904, resp.Code)
	assert.Equal(t, "batch db error", resp.Message)
}

// ---------- BatchUpdateStatus ----------

func TestHandler_BatchUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.BatchStatusUpdateRequest{IDs: []uint{1, 2}, Status: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量状态更新完成", resp.Message)
	assert.Equal(t, []uint{1, 2}, env.mock.lastBatchUpdateStatusIDs)
	assert.Equal(t, 2, env.mock.lastBatchUpdateStatusStatus)
	var br dto.BatchResultResponse
	require.NoError(t, json.Unmarshal(resp.Data, &br))
	assert.Equal(t, 2, br.Success)
}

func TestHandler_BatchUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/house/admin/batch/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchUpdateStatus_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// status 不在 oneof=1 2 3 4（传 9）→ 校验失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/status", map[string]interface{}{"ids": []uint{1}, "status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.batchUpdateStatusResult = nil
	env.mock.batchUpdateStatusErr = errors.New("db error")
	body := dto.BatchStatusUpdateRequest{IDs: []uint{1}, Status: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/status", body)

	assert.Equal(t, 2904, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- BatchDelete ----------

func TestHandler_BatchDelete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := dto.BatchDeleteRequest{IDs: []uint{1, 2}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/delete", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量删除完成", resp.Message)
	assert.Equal(t, []uint{1, 2}, env.mock.lastBatchDeleteIDs)
	var br dto.BatchResultResponse
	require.NoError(t, json.Unmarshal(resp.Data, &br))
	assert.Equal(t, 2, br.Success)
}

func TestHandler_BatchDelete_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/house/admin/batch/delete", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchDelete_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// ids 为空 → 校验失败 → 400
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/delete", map[string]interface{}{"ids": []uint{}})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchDelete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.batchDeleteResult = nil
	env.mock.batchDeleteErr = errors.New("db error")
	body := dto.BatchDeleteRequest{IDs: []uint{1}}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/house/admin/batch/delete", body)

	// BatchDelete 用 CodeHouseError=2901
	assert.Equal(t, 2901, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}
