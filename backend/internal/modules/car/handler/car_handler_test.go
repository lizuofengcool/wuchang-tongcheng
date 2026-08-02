// Package handler_test 同城车辆买卖主表 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 car CarHandler 全部分支：
//   - 公开只读接口无需登录（GetByID/List/Nearby/Search/AdvancedSearch/FavStatus）
//   - 需登录接口未登录拦截（Create/Update/Delete/ListMine/ListFavs/Fav → 401 "请先登录"）
//   - 无登录校验的公开互动接口（IncrContact/IncrShare/RecordView）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/max/min → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeCarError=3001 / CodeCarNotFound=3002 /
//     CodeCarPublishError=3003 / CodeCarAuditError=3004）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - Create 用户画像透传（userName/userPhone/userAvatar）
//   - RecordView 客户端 IP 提取（X-Forwarded-For 首段 / X-Real-IP）与 userID 透传
//   - Fav toggle 语义（HasFaved=true → "收藏成功" / false → "已取消收藏"）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 mall/dh114/house/job/marketing/shop/groupbuy handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/car/dto"
	carHandler "wuchang-tongcheng/internal/modules/car/handler"
	"wuchang-tongcheng/internal/modules/car/service"
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

// mockCarService 内存 mock，实现 service.CarService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockCarService struct {
	// ===== 调用记录 =====
	// C 端发布/更新/删除
	lastCreateRegionID   uint
	lastCreateUserID     uint
	lastCreateUserName   string
	lastCreateUserPhone  string
	lastCreateUserAvatar string
	lastCreateReq        *dto.CreateCarRequest

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateCarRequest

	lastDeleteID         uint
	lastDeleteOperatorID uint

	// 只读
	lastGetByIDID     uint
	lastGetByIDUserID uint

	lastListRegionID uint
	lastListReq      *dto.CarListRequest

	lastNearbyRegionID uint
	lastNearbyReq      *dto.CarNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.CarSearchRequest

	lastAdvancedRegionID uint
	lastAdvancedReq      *dto.AdvancedSearchRequest

	lastListMineUserID   uint
	lastListMinePage     int
	lastListMinePageSize int

	// 收藏
	lastFavUserID uint
	lastFavCarID  uint

	lastFavStatusUserID uint
	lastFavStatusCarID  uint

	lastListFavsUserID   uint
	lastListFavsPage     int
	lastListFavsPageSize int

	// 互动
	lastIncrContactID uint
	lastIncrShareID   uint

	lastRecordViewUserID uint
	lastRecordViewIP     string
	lastRecordViewReq    *dto.CarViewRequest

	// M 端
	lastAdminListReq *dto.CarAdminListRequest

	lastAdminGetByIDID uint

	lastAuditID         uint
	lastAuditStatus     int
	lastAuditReason     string

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	lastRealCarVerifyID      uint
	lastRealCarVerifyVerified bool
	lastRealCarVerifyReason  string

	lastUpdatePromotionID uint
	lastUpdatePromotionReq *dto.PromotionRequest

	// ===== 返回值预设 =====
	createResult *dto.CarInfo
	createErr    error

	updateErr error
	deleteErr error

	getByIDResult *dto.CarInfo
	getByIDErr    error

	listResult []dto.CarInfo
	listErr    error
	listTotal  int64

	nearbyResult []dto.CarInfo
	nearbyErr    error
	nearbyTotal  int64

	searchResult []dto.CarInfo
	searchErr    error
	searchTotal  int64

	advancedResult []dto.CarInfo
	advancedErr    error
	advancedTotal  int64

	listMineResult []dto.CarInfo
	listMineErr    error
	listMineTotal  int64

	favResult *dto.FavResponse
	favErr    error

	favStatusResult *dto.FavResponse
	favStatusErr    error

	listFavsResult []dto.CarInfo
	listFavsErr    error
	listFavsTotal  int64

	incrContactErr error
	incrShareErr   error
	recordViewErr  error

	adminListResult []dto.CarInfo
	adminListErr    error
	adminListTotal  int64

	adminGetByIDResult *dto.CarInfo
	adminGetByIDErr    error

	auditErr error

	adminUpdateStatusErr error

	realCarVerifyErr error

	updatePromotionErr error
}

// ===== C 端 =====

func (m *mockCarService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateCarRequest) (*dto.CarInfo, error) {
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

func (m *mockCarService) Update(id uint, operatorID uint, req *dto.UpdateCarRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockCarService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockCarService) GetByID(id uint, userID uint) (*dto.CarInfo, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUserID = userID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockCarService) List(regionID uint, req *dto.CarListRequest) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}

func (m *mockCarService) ListNearby(regionID uint, req *dto.CarNearbyRequest) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyReq = req
	if m.nearbyErr != nil {
		return nil, nil, m.nearbyErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.nearbyTotal}, m.nearbyResult, nil
}

func (m *mockCarService) Search(regionID uint, req *dto.CarSearchRequest) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.searchTotal}, m.searchResult, nil
}

func (m *mockCarService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastListMineUserID = userID
	m.lastListMinePage = page
	m.lastListMinePageSize = pageSize
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	return &utils.Pagination{Page: page, PageSize: pageSize, Total: m.listMineTotal}, m.listMineResult, nil
}

func (m *mockCarService) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastAdvancedRegionID = regionID
	m.lastAdvancedReq = req
	if m.advancedErr != nil {
		return nil, nil, m.advancedErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.advancedTotal}, m.advancedResult, nil
}

// ===== 收藏 =====

func (m *mockCarService) Fav(userID, carID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavCarID = carID
	if m.favErr != nil {
		return nil, m.favErr
	}
	return m.favResult, nil
}

func (m *mockCarService) FavStatus(userID, carID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusCarID = carID
	if m.favStatusErr != nil {
		return nil, m.favStatusErr
	}
	return m.favStatusResult, nil
}

func (m *mockCarService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastListFavsUserID = userID
	m.lastListFavsPage = page
	m.lastListFavsPageSize = pageSize
	if m.listFavsErr != nil {
		return nil, nil, m.listFavsErr
	}
	return &utils.Pagination{Page: page, PageSize: pageSize, Total: m.listFavsTotal}, m.listFavsResult, nil
}

// ===== 互动 =====

func (m *mockCarService) IncrContact(id uint) error {
	m.lastIncrContactID = id
	return m.incrContactErr
}

func (m *mockCarService) IncrShare(id uint) error {
	m.lastIncrShareID = id
	return m.incrShareErr
}

func (m *mockCarService) RecordView(userID uint, ip string, req *dto.CarViewRequest) error {
	m.lastRecordViewUserID = userID
	m.lastRecordViewIP = ip
	m.lastRecordViewReq = req
	return m.recordViewErr
}

// ===== M 端 =====

func (m *mockCarService) AdminList(req *dto.CarAdminListRequest) (*utils.Pagination, []dto.CarInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.adminListTotal}, m.adminListResult, nil
}

func (m *mockCarService) AdminGetByID(id uint) (*dto.CarInfo, error) {
	m.lastAdminGetByIDID = id
	if m.adminGetByIDErr != nil {
		return nil, m.adminGetByIDErr
	}
	return m.adminGetByIDResult, nil
}

func (m *mockCarService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockCarService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.adminUpdateStatusErr
}

func (m *mockCarService) RealCarVerify(id uint, verified bool, reason string) error {
	m.lastRealCarVerifyID = id
	m.lastRealCarVerifyVerified = verified
	m.lastRealCarVerifyReason = reason
	return m.realCarVerifyErr
}

func (m *mockCarService) UpdatePromotion(id uint, req *dto.PromotionRequest) error {
	m.lastUpdatePromotionID = id
	m.lastUpdatePromotionReq = req
	return m.updatePromotionErr
}

// 编译期接口实现校验
var _ service.CarService = (*mockCarService)(nil)

// handlerEnv 测试环境：gin 引擎 + mock service。
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockCarService
}

// newHandlerEnv 构造 gin 引擎并注册 car 车源主表路由（路径与 car/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 同时注入 username/phone/avatar 冗余字段，用于 Create 透传断言。
// 路由注册去掉限流/权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockCarService{
		createResult: &dto.CarInfo{
			ID:        1,
			Title:     "2018款 本田思域 1.5T CVT",
			UserID:    ctxUserID,
			UserName:  "张三",
			Price:     128000,
			Status:    1,
			City:      "武昌",
		},
		getByIDResult: &dto.CarInfo{
			ID: 1, Title: "2018款 本田思域", UserID: 1, Price: 128000, Status: 1,
		},
		adminGetByIDResult: &dto.CarInfo{
			ID: 1, Title: "2018款 本田思域", Status: 1, AuditStatus: 1,
		},
		listResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", Price: 128000, Status: 1},
			{ID: 2, Title: "丰田卡罗拉", Price: 98000, Status: 1},
		},
		listTotal: 2,
		nearbyResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", Price: 128000, Distance: 1.2},
		},
		nearbyTotal: 1,
		searchResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", Price: 128000},
		},
		searchTotal: 1,
		advancedResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", Price: 128000},
			{ID: 2, Title: "丰田卡罗拉", Price: 98000},
			{ID: 3, Title: "大众朗逸", Price: 86000},
		},
		advancedTotal: 3,
		listMineResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", UserID: ctxUserID},
		},
		listMineTotal: 1,
		favResult: &dto.FavResponse{HasFaved: true, FavCount: 5},
		favStatusResult: &dto.FavResponse{HasFaved: false, FavCount: 4},
		listFavsResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域"},
		},
		listFavsTotal: 1,
		adminListResult: []dto.CarInfo{
			{ID: 1, Title: "本田思域", Status: 1, AuditStatus: 1},
			{ID: 2, Title: "丰田卡罗拉", Status: 1, AuditStatus: 0},
		},
		adminListTotal: 2,
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

	h := carHandler.NewCarHandler(mock)
	root := r.Group("/api/v1/car")
	// 公开只读接口
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/advanced-search", h.AdvancedSearch)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/fav", h.FavStatus)
	// 需登录接口（C 端）
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/mine", h.ListMine)
	root.GET("/favorites", h.ListFavs)
	root.POST("/:id/fav", h.Fav)
	root.POST("/:id/contact", h.IncrContact)
	root.POST("/:id/share", h.IncrShare)
	root.POST("/:id/views", h.RecordView)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/cars", h.AdminList)
	admin.GET("/cars/:id", h.AdminGetByID)
	admin.PUT("/cars/:id/audit", h.Audit)
	admin.PUT("/cars/:id/status", h.AdminUpdateStatus)
	admin.PUT("/cars/:id/real-car-verify", h.RealCarVerify)
	admin.PUT("/cars/:id/promotion", h.UpdatePromotion)

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

// validCreateBody 构造一份可通过 Bind 校验的完整 CreateCarRequest 请求体。
func validCreateBody() map[string]interface{} {
	return map[string]interface{}{
		"title":        "2018款 本田思域 1.5T CVT 热血版",
		"content":      "车况精品，无大事故，支持检测",
		"cover_image":  "https://cdn.example.com/cover.jpg",
		"listing_type": "used",
		"source_type":  "personal",
		"car_type":     "sedan",
		"brand_id":     10,
		"brand_name":   "本田",
		"price":        128000,
		"mileage":      5.6,
		"mileage_unit": "km",
		"transmission": "cvt",
		"fuel_type":    "gasoline",
		"city":         "武昌",
		"status":       1,
	}
}

// ==================== 公开只读接口（无需登录） ====================

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	// 未登录 userID=0 透传给 service
	assert.Equal(t, uint(0), env.mock.lastGetByIDUserID)
	var info dto.CarInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "2018款 本田思域", info.Title)
}

func TestHandler_GetByID_Success_WithLogin(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	assert.Equal(t, uint(7), env.mock.lastGetByIDUserID)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("车源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/999", nil)

	// CodeCarNotFound=3002
	assert.Equal(t, utils.CodeCarNotFound, resp.Code)
	assert.Equal(t, "车源不存在", resp.Message)
}

// ---------- List ----------

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car?page=1&page_size=10&keyword=本田", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "本田", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, 10, p.PageSize)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car", nil)

	// CodeCarError=3001
	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "查询失败", resp.Message)
}

// ---------- Nearby ----------

func TestHandler_Nearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/nearby?latitude=30.55&longitude=114.30&radius_km=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastNearbyRegionID)
	require.NotNil(t, env.mock.lastNearbyReq)
	assert.Equal(t, 30.55, env.mock.lastNearbyReq.Latitude)
	assert.Equal(t, 114.30, env.mock.lastNearbyReq.Longitude)
	assert.Equal(t, 5.0, env.mock.lastNearbyReq.RadiusKm)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Nearby_BindError_MissingLat(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/nearby?longitude=114.30", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Nearby_BindError_MissingLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/nearby?latitude=30.55", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Nearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.nearbyResult = nil
	env.mock.nearbyErr = errors.New("空间查询不可用")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/nearby?latitude=30.55&longitude=114.30", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "空间查询不可用", resp.Message)
}

// ---------- Search ----------

func TestHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/search?keyword=本田", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "本田", env.mock.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Search_BindError_EmptyKeyword(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/search", nil)

	// binding:"required" 关键字缺失/空 → Bind 失败
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("搜索引擎故障")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/search?keyword=本田", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "搜索引擎故障", resp.Message)
}

// ---------- AdvancedSearch ----------

func TestHandler_AdvancedSearch_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/advanced-search?keyword=本田&min_price=50000&max_price=150000", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastAdvancedRegionID)
	require.NotNil(t, env.mock.lastAdvancedReq)
	assert.Equal(t, "本田", env.mock.lastAdvancedReq.Keyword)
	assert.Equal(t, 50000.0, env.mock.lastAdvancedReq.MinPrice)
	assert.Equal(t, 150000.0, env.mock.lastAdvancedReq.MaxPrice)
	p := parsePage(t, resp)
	assert.Equal(t, int64(3), p.Total)
}

func TestHandler_AdvancedSearch_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.advancedResult = nil
	env.mock.advancedErr = errors.New("高级查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/advanced-search", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "高级查询失败", resp.Message)
}

// ---------- FavStatus（公开，未登录返回 has_faved=false） ----------

func TestHandler_FavStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(8), env.mock.lastFavStatusCarID)
	assert.Equal(t, uint(0), env.mock.lastFavStatusUserID)
	var fr dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &fr))
	assert.False(t, fr.HasFaved)
}

func TestHandler_FavStatus_Success_WithLogin(t *testing.T) {
	env := newHandlerEnv(t, 9, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/8/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(9), env.mock.lastFavStatusUserID)
}

func TestHandler_FavStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_FavStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.favStatusResult = nil
	env.mock.favStatusErr = errors.New("车源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/999/fav", nil)

	assert.Equal(t, utils.CodeCarNotFound, resp.Code)
	assert.Equal(t, "车源不存在", resp.Message)
}

// ==================== 需登录接口（C 端发布/收藏） ====================

// ---------- Create ----------

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car", validCreateBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	// 上下文透传断言
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "张三", env.mock.lastCreateUserName)
	assert.Equal(t, "13800000000", env.mock.lastCreateUserPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateUserAvatar)
	// 请求体透传断言
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "2018款 本田思域 1.5T CVT 热血版", env.mock.lastCreateReq.Title)
	assert.Equal(t, "used", env.mock.lastCreateReq.ListingType)
	assert.Equal(t, 128000.0, env.mock.lastCreateReq.Price)
	var info dto.CarInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_Create_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car", validCreateBody())

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_BindError_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	delete(body, "title") // title 为 required
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindError_InvalidListingType(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := validCreateBody()
	body["listing_type"] = "invalid_type" // oneof 校验失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/car", "{invalid json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("图片处理失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car", validCreateBody())

	// CodeCarPublishError=3003
	assert.Equal(t, utils.CodeCarPublishError, resp.Code)
	assert.Equal(t, "图片处理失败", resp.Message)
}

// ---------- Update ----------

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	newTitle := "降价急售 本田思域"
	body := map[string]interface{}{"title": newTitle, "price": 118000}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	require.NotNil(t, env.mock.lastUpdateReq.Title)
	assert.Equal(t, newTitle, *env.mock.lastUpdateReq.Title)
}

func TestHandler_Update_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateID)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Update_BindError_TitleTooLong(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := map[string]interface{}{"title": strings.Repeat("a", 201)} // max=200
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/3", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Update_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/car/3", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.updateErr = errors.New("无权操作此车源")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "无权操作此车源", resp.Message)
}

// ---------- Delete ----------

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/car/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestHandler_Delete_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/car/3", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/car/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.deleteErr = errors.New("无权操作此车源")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/car/3", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "无权操作此车源", resp.Message)
}

// ---------- ListMine ----------

func TestHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/mine?page=2&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 5, env.mock.lastListMinePageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	assert.Equal(t, 2, p.Page)
}

func TestHandler_ListMine_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/mine", nil)

	assert.Equal(t, 0, resp.Code)
	// parsePagination 默认 page=1 page_size=10
	assert.Equal(t, 1, env.mock.lastListMinePage)
	assert.Equal(t, 10, env.mock.lastListMinePageSize)
}

func TestHandler_ListMine_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.listMineResult = nil
	env.mock.listMineErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/mine", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "查询失败", resp.Message)
}

// ---------- ListFavs ----------

func TestHandler_ListFavs_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/favorites?page=1&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListFavsUserID)
	assert.Equal(t, 1, env.mock.lastListFavsPage)
	assert.Equal(t, 20, env.mock.lastListFavsPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListFavs_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/favorites", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListFavs_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.listFavsResult = nil
	env.mock.listFavsErr = errors.New("收藏查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/favorites", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "收藏查询失败", resp.Message)
}

// ---------- Fav（toggle） ----------

func TestHandler_Fav_Success_Faved(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.favResult = &dto.FavResponse{HasFaved: true, FavCount: 6}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastFavUserID)
	assert.Equal(t, uint(3), env.mock.lastFavCarID)
}

func TestHandler_Fav_Success_Unfaved(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.favResult = &dto.FavResponse{HasFaved: false, FavCount: 5}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
}

func TestHandler_Fav_NotLogin(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Fav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Fav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.favResult = nil
	env.mock.favErr = errors.New("收藏失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/fav", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "收藏失败", resp.Message)
}

// ==================== 公开互动接口（无登录校验） ====================

// ---------- IncrContact ----------

func TestHandler_IncrContact_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开，无需登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/contact", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录联系", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrContactID)
}

func TestHandler_IncrContact_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/abc/contact", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrContact_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrContactErr = errors.New("记录失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/contact", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "记录失败", resp.Message)
}

// ---------- IncrShare ----------

func TestHandler_IncrShare_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/share", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录分享", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastIncrShareID)
}

func TestHandler_IncrShare_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/abc/share", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_IncrShare_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.incrShareErr = errors.New("记录失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/3/share", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "记录失败", resp.Message)
}

// ---------- RecordView ----------

func TestHandler_RecordView_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5) // 登录用户带 userID
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/1/views", map[string]interface{}{
		"car_id": 1,
		"device": "pc",
		"source": "search",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已记录浏览", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastRecordViewUserID)
	require.NotNil(t, env.mock.lastRecordViewReq)
	assert.Equal(t, uint(1), env.mock.lastRecordViewReq.CarID)
	assert.Equal(t, "pc", env.mock.lastRecordViewReq.Device)
	assert.Equal(t, "search", env.mock.lastRecordViewReq.Source)
}

func TestHandler_RecordView_Success_Anonymous(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录也可记录浏览
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/1/views", map[string]interface{}{
		"car_id": 1,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastRecordViewUserID)
}

func TestHandler_RecordView_IP_XForwardedFor_Single(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/car/1/views",
		map[string]interface{}{"car_id": 1}, map[string]string{"X-Forwarded-For": "1.2.3.4"})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "1.2.3.4", env.mock.lastRecordViewIP)
}

func TestHandler_RecordView_IP_XForwardedFor_Multiple(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/car/1/views",
		map[string]interface{}{"car_id": 1}, map[string]string{"X-Forwarded-For": "1.2.3.4, 5.6.7.8"})

	assert.Equal(t, 0, resp.Code)
	// 多 IP 取首段并去空白
	assert.Equal(t, "1.2.3.4", env.mock.lastRecordViewIP)
}

func TestHandler_RecordView_IP_XRealIP(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/car/1/views",
		map[string]interface{}{"car_id": 1}, map[string]string{"X-Real-IP": "9.9.9.9"})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "9.9.9.9", env.mock.lastRecordViewIP)
}

func TestHandler_RecordView_BindError_MissingCarID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/1/views", map[string]interface{}{})

	// car_id 为 required
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RecordView_BindError_InvalidDevice(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/1/views", map[string]interface{}{
		"car_id": 1,
		"device": "tv", // oneof 校验失败
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RecordView_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/car/1/views", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RecordView_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.mock.recordViewErr = errors.New("浏览记录写入失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/car/1/views", map[string]interface{}{"car_id": 1})

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "浏览记录写入失败", resp.Message)
}

// ==================== M 端管理后台接口 ====================

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/admin/cars?page=1&page_size=10&keyword=本田&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, "本田", env.mock.lastAdminListReq.Keyword)
	require.NotNil(t, env.mock.lastAdminListReq.Status)
	assert.Equal(t, 1, *env.mock.lastAdminListReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("管理列表查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/admin/cars", nil)

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "管理列表查询失败", resp.Message)
}

// ---------- AdminGetByID ----------

func TestHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/admin/cars/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastAdminGetByIDID)
	var info dto.CarInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/admin/cars/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminGetByIDResult = nil
	env.mock.adminGetByIDErr = errors.New("车源不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/car/admin/cars/999", nil)

	assert.Equal(t, utils.CodeCarNotFound, resp.Code)
	assert.Equal(t, "车源不存在", resp.Message)
}

// ---------- Audit ----------

func TestHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/audit", map[string]interface{}{
		"audit_status": 1,
		"audit_reason": "内容合规",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditStatus)
	assert.Equal(t, "内容合规", env.mock.lastAuditReason)
}

func TestHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/abc/audit", map[string]interface{}{
		"audit_status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Audit_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/audit", map[string]interface{}{
		"audit_status": 9, // oneof=0 1 2
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/car/admin/cars/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.auditErr = errors.New("已审核的车源不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/audit", map[string]interface{}{
		"audit_status": 1,
	})

	// CodeCarAuditError=3004
	assert.Equal(t, utils.CodeCarAuditError, resp.Code)
	assert.Equal(t, "已审核的车源不能重复审核", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/status", map[string]interface{}{
		"status": 2, // 下架
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 2, env.mock.lastAdminUpdateStatusStatus)
}

func TestHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/abc/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminUpdateStatus_BindError_MissingStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/status", map[string]interface{}{})

	// status 缺省为 0，不在 oneof=1 2 3 4 5 内 → Bind 失败
	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/status", map[string]interface{}{
		"status": 99,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/car/admin/cars/3/status", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.adminUpdateStatusErr = errors.New("状态迁移非法")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "状态迁移非法", resp.Message)
}

// ---------- RealCarVerify ----------

func TestHandler_RealCarVerify_Success_Verified(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/real-car-verify", map[string]interface{}{
		"verified": true,
		"reason":   "线下核验通过",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "认证状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastRealCarVerifyID)
	assert.True(t, env.mock.lastRealCarVerifyVerified)
	assert.Equal(t, "线下核验通过", env.mock.lastRealCarVerifyReason)
}

func TestHandler_RealCarVerify_Success_Unverified(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/real-car-verify", map[string]interface{}{
		"verified": false,
		"reason":   "材料不全",
	})

	assert.Equal(t, 0, resp.Code)
	assert.False(t, env.mock.lastRealCarVerifyVerified)
	assert.Equal(t, "材料不全", env.mock.lastRealCarVerifyReason)
}

func TestHandler_RealCarVerify_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/abc/real-car-verify", map[string]interface{}{
		"verified": true,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_RealCarVerify_BindError_ReasonTooLong(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/real-car-verify", map[string]interface{}{
		"verified": true,
		"reason":   strings.Repeat("x", 501), // max=500
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RealCarVerify_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/car/admin/cars/3/real-car-verify", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_RealCarVerify_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.realCarVerifyErr = errors.New("认证更新失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/real-car-verify", map[string]interface{}{
		"verified": true,
	})

	assert.Equal(t, utils.CodeCarAuditError, resp.Code)
	assert.Equal(t, "认证更新失败", resp.Message)
}

// ---------- UpdatePromotion ----------

func TestHandler_UpdatePromotion_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	featured := true
	picked := false
	level := 3
	weight := 5.5
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/promotion", map[string]interface{}{
		"featured":        featured,
		"picked":          picked,
		"promotion_level": level,
		"traffic_weight":  weight,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "推广配置更新成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastUpdatePromotionID)
	require.NotNil(t, env.mock.lastUpdatePromotionReq)
	require.NotNil(t, env.mock.lastUpdatePromotionReq.Featured)
	assert.True(t, *env.mock.lastUpdatePromotionReq.Featured)
	require.NotNil(t, env.mock.lastUpdatePromotionReq.PromotionLevel)
	assert.Equal(t, 3, *env.mock.lastUpdatePromotionReq.PromotionLevel)
}

func TestHandler_UpdatePromotion_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/abc/promotion", map[string]interface{}{
		"featured": true,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_UpdatePromotion_BindError_LevelOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/promotion", map[string]interface{}{
		"promotion_level": 11, // max=10
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_BindError_WeightOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/promotion", map[string]interface{}{
		"traffic_weight": 10.5, // max=9.99
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/car/admin/cars/3/promotion", "not-json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_UpdatePromotion_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.updatePromotionErr = errors.New("推广配置更新失败")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/car/admin/cars/3/promotion", map[string]interface{}{
		"featured": true,
	})

	assert.Equal(t, utils.CodeCarError, resp.Code)
	assert.Equal(t, "推广配置更新失败", resp.Message)
}
