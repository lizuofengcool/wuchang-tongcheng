// Package handler_test 招聘求职模块 HTTP 处理层单元测试 - 职位主表。
//
// 使用 gin + httptest + mock Service，覆盖 JobHandler 全部分支：
//   - C 端：Create/Update/Delete/GetByID/List/ListNearby/Search/AdvancedSearch/ListMine/ListSimilar/UpdateStatus
//   - 收藏：Fav/Unfav/FavStatus/ListFavs
//   - 推广：Promotion
//   - M 端管理：AdminList/AdminGetByID/Audit/AdminUpdateStatus
//
// 覆盖维度：
//   - 公开接口无需登录（List/GetByID/ListSimilar/FavStatus/ListNearby/Search/AdvancedSearch/AdminList 等）
//   - 用户接口未登录拦截（Create/Update/Delete/ListMine/UpdateStatus/Fav/Unfav/ListFavs/Promotion → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/gte/lte → 400 "参数错误"）
//   - ListNearby 经纬度越界（→ 400 "经纬度参数无效"）
//   - service 成功/错误透传（业务码 2801-2804 区间 + message + data 透传）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 marketing/shop/house/groupbuy/category/region/news/file/setting/permission handler 测试同风格。
package handler_test

import (
	"bytes"
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
	"wuchang-tongcheng/internal/modules/job/dto"
	jobHandler "wuchang-tongcheng/internal/modules/job/handler"
	"wuchang-tongcheng/internal/modules/job/service"
	"wuchang-tongcheng/internal/pkg/utils"
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

// ==================== JobService mock ====================

// mockJobService 内存 mock，实现 service.JobService 接口。
type mockJobService struct {
	// Create
	lastCreateRegionID uint
	lastCreateUserID   uint
	lastCreateName     string
	lastCreatePhone    string
	lastCreateAvatar   string
	lastCreateReq      *dto.CreateJobRequest
	createResult       *dto.JobInfo
	createErr          error

	// Update
	lastUpdateID       uint
	lastUpdateOpID     uint
	lastUpdateReq      *dto.UpdateJobRequest
	updateErr          error

	// Delete
	lastDeleteID   uint
	lastDeleteOpID uint
	deleteErr      error

	// GetByID
	lastGetByIDID   uint
	lastGetByIDUser uint
	getByIDResult   *dto.JobDetailResponse
	getByIDErr      error

	// List
	lastListRegionID uint
	lastListReq      *dto.JobListRequest
	listResult       []dto.JobInfo
	listErr          error

	// ListNearby
	lastNearbyRegionID uint
	lastNearbyReq      *dto.JobNearbyRequest
	nearbyResult       []dto.JobInfo
	nearbyErr          error

	// Search
	lastSearchRegionID uint
	lastSearchReq      *dto.JobSearchRequest
	searchResult       []dto.JobInfo
	searchErr          error

	// AdvancedSearch
	lastAdvancedRegionID uint
	lastAdvancedReq      *dto.AdvancedSearchRequest
	advancedResult       []dto.JobInfo
	advancedErr          error

	// ListMine
	lastListMineUserID uint
	lastListMinePage   int
	lastListMineSize   int
	listMineResult     []dto.JobInfo
	listMineErr        error

	// ListSimilar
	lastSimilarID    uint
	lastSimilarLimit int
	similarResult    []dto.SimilarJobResponse
	similarErr       error

	// UpdateStatus
	lastUpdateStatusID   uint
	lastUpdateStatusOp   uint
	lastUpdateStatusVal  int
	updateStatusErr      error

	// Fav
	lastFavUserID uint
	lastFavJobID  uint
	favResult     *dto.FavResponse
	favErr        error

	// Unfav
	lastUnfavUserID uint
	lastUnfavJobID  uint
	unfavResult     *dto.FavResponse
	unfavErr        error

	// FavStatus
	lastFavStatusUserID uint
	lastFavStatusJobID  uint
	favStatusResult     *dto.FavResponse
	favStatusErr        error

	// ListFavs
	lastListFavsUserID uint
	lastListFavsPage   int
	lastListFavsSize   int
	listFavsResult     []dto.JobInfo
	listFavsErr        error

	// Promotion
	lastPromotionID   uint
	lastPromotionOp   uint
	lastPromotionReq  *dto.JobPromotionRequest
	promotionErr      error

	// AdminList
	lastAdminListReq *dto.JobAdminListRequest
	adminListResult  []dto.JobInfo
	adminListErr     error

	// AdminGetByID
	lastAdminGetID  uint
	adminGetResult  *dto.JobDetailResponse
	adminGetErr     error

	// Audit
	lastAuditID     uint
	lastAuditStatus int
	lastAuditReason string
	auditErr        error

	// AdminUpdateStatus
	lastAdminUpdateID   uint
	lastAdminUpdateVal  int
	adminUpdateErr      error
}

func (m *mockJobService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateJobRequest) (*dto.JobInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateName = userName
	m.lastCreatePhone = userPhone
	m.lastCreateAvatar = userAvatar
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockJobService) Update(id uint, operatorID uint, req *dto.UpdateJobRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOpID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockJobService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOpID = operatorID
	return m.deleteErr
}

func (m *mockJobService) GetByID(id uint, userID uint) (*dto.JobDetailResponse, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUser = userID
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockJobService) List(regionID uint, req *dto.JobListRequest) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockJobService) ListNearby(regionID uint, req *dto.JobNearbyRequest) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastNearbyRegionID = regionID
	m.lastNearbyReq = req
	if m.nearbyErr != nil {
		return nil, nil, m.nearbyErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.nearbyResult))
	return p, m.nearbyResult, nil
}

func (m *mockJobService) Search(regionID uint, req *dto.JobSearchRequest) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.searchResult))
	return p, m.searchResult, nil
}

func (m *mockJobService) AdvancedSearch(regionID uint, req *dto.AdvancedSearchRequest) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastAdvancedRegionID = regionID
	m.lastAdvancedReq = req
	if m.advancedErr != nil {
		return nil, nil, m.advancedErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.advancedResult))
	return p, m.advancedResult, nil
}

func (m *mockJobService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastListMineUserID = userID
	m.lastListMinePage = page
	m.lastListMineSize = pageSize
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listMineResult))
	return p, m.listMineResult, nil
}

func (m *mockJobService) ListSimilar(jobID uint, limit int) ([]dto.SimilarJobResponse, error) {
	m.lastSimilarID = jobID
	m.lastSimilarLimit = limit
	if m.similarErr != nil {
		return nil, m.similarErr
	}
	return m.similarResult, nil
}

func (m *mockJobService) UpdateStatus(id uint, operatorID uint, status int) error {
	m.lastUpdateStatusID = id
	m.lastUpdateStatusOp = operatorID
	m.lastUpdateStatusVal = status
	return m.updateStatusErr
}

func (m *mockJobService) Fav(userID, jobID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavJobID = jobID
	if m.favErr != nil {
		return nil, m.favErr
	}
	return m.favResult, nil
}

func (m *mockJobService) Unfav(userID, jobID uint) (*dto.FavResponse, error) {
	m.lastUnfavUserID = userID
	m.lastUnfavJobID = jobID
	if m.unfavErr != nil {
		return nil, m.unfavErr
	}
	return m.unfavResult, nil
}

func (m *mockJobService) FavStatus(userID, jobID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusJobID = jobID
	if m.favStatusErr != nil {
		return nil, m.favStatusErr
	}
	return m.favStatusResult, nil
}

func (m *mockJobService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastListFavsUserID = userID
	m.lastListFavsPage = page
	m.lastListFavsSize = pageSize
	if m.listFavsErr != nil {
		return nil, nil, m.listFavsErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listFavsResult))
	return p, m.listFavsResult, nil
}

func (m *mockJobService) Promotion(id uint, operatorID uint, req *dto.JobPromotionRequest) error {
	m.lastPromotionID = id
	m.lastPromotionOp = operatorID
	m.lastPromotionReq = req
	return m.promotionErr
}

func (m *mockJobService) AdminList(req *dto.JobAdminListRequest) (*utils.Pagination, []dto.JobInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.adminListResult))
	return p, m.adminListResult, nil
}

func (m *mockJobService) AdminGetByID(id uint) (*dto.JobDetailResponse, error) {
	m.lastAdminGetID = id
	if m.adminGetErr != nil {
		return nil, m.adminGetErr
	}
	return m.adminGetResult, nil
}

func (m *mockJobService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditStatus = auditStatus
	m.lastAuditReason = auditReason
	return m.auditErr
}

func (m *mockJobService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateID = id
	m.lastAdminUpdateVal = status
	return m.adminUpdateErr
}

var _ service.JobService = (*mockJobService)(nil)

// ==================== 测试环境 ====================

// handlerEnv handler 测试环境（聚合 mock service + JobHandler）
type handlerEnv struct {
	engine *gin.Engine
	svc    *mockJobService
}

// newHandlerEnv 构造 gin 引擎并注册 JobHandler 路由（路径与 job/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉权限/限流中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svc := &mockJobService{
		createResult:    &dto.JobInfo{ID: 1, Title: "Java高级工程师", UserID: ctxUserID, RegionID: regionID, Status: 1},
		getByIDResult:   &dto.JobDetailResponse{JobInfo: dto.JobInfo{ID: 1, Title: "Java高级工程师", RegionID: regionID}},
		listResult:      []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}, {ID: 2, Title: "前端工程师"}},
		nearbyResult:    []dto.JobInfo{{ID: 1, Title: "Java高级工程师", Distance: 1.2}},
		searchResult:    []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}},
		advancedResult:  []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}},
		listMineResult:  []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}},
		similarResult:   []dto.SimilarJobResponse{{JobID: 2, Title: "高级Java", SalaryMin: 15000, SalaryUnit: "month", WorkCity: "武汉", Similarity: 0.8}},
		favResult:       &dto.FavResponse{HasFaved: true, FavCount: 5},
		unfavResult:     &dto.FavResponse{HasFaved: false, FavCount: 4},
		favStatusResult: &dto.FavResponse{HasFaved: true, FavCount: 5},
		listFavsResult:  []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}},
		adminListResult: []dto.JobInfo{{ID: 1, Title: "Java高级工程师"}, {ID: 2, Title: "前端工程师"}},
		adminGetResult:  &dto.JobDetailResponse{JobInfo: dto.JobInfo{ID: 1, Title: "Java高级工程师", AuditStatus: 1}},
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

	h := jobHandler.NewJobHandler(svc)

	// 注册路由，路径与 job/plugin.go RegisterRoutes 保持一致（去掉权限/限流中间件，纯测 handler）
	root := r.Group("/api/v1/job")

	// ==================== 公开路由（固定路径优先注册） ====================
	root.GET("/search", h.Search)
	root.GET("/nearby", h.ListNearby)
	root.GET("/advanced-search", h.AdvancedSearch)
	root.GET("/mine", h.ListMine)
	root.GET("/favorites", h.ListFavs)
	root.GET("", h.List)

	// 职位基础（含 :id 参数）
	root.GET("/:id", h.GetByID)
	root.GET("/:id/similar", h.ListSimilar)
	root.GET("/:id/fav", h.FavStatus)

	// ==================== 需登录路由（C 端发布/收藏/推广） ====================
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.PUT("/:id/status", h.UpdateStatus)
	root.POST("/:id/fav", h.Fav)
	root.DELETE("/:id/fav", h.Unfav)
	root.POST("/:id/promotion", h.Promotion)

	// ==================== 管理后台路由（/admin 组） ====================
	admin := root.Group("/admin")
	admin.GET("/list", h.AdminList)
	admin.GET("/:id", h.AdminGetByID)
	admin.PUT("/:id/audit", h.Audit)
	admin.PUT("/:id/status", h.AdminUpdateStatus)

	return &handlerEnv{engine: r.Engine(), svc: svc}
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

// =====================================================================
// ==================== JobHandler.Create 测试 ====================
// =====================================================================

func TestJobHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := map[string]interface{}{
		"title":        "Java高级工程师",
		"category_id":  10,
		"salary_min":   15000,
		"salary_unit":  "month",
		"hiring_count": 2,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	assert.Equal(t, uint(5), env.svc.lastCreateRegionID)
	assert.Equal(t, uint(7), env.svc.lastCreateUserID)
	assert.Equal(t, "张三", env.svc.lastCreateName)
	assert.Equal(t, "13800000000", env.svc.lastCreatePhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.svc.lastCreateAvatar)
	require.NotNil(t, env.svc.lastCreateReq)
	assert.Equal(t, "Java高级工程师", env.svc.lastCreateReq.Title)
	assert.Equal(t, uint(10), env.svc.lastCreateReq.CategoryID)
	var info dto.JobInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestJobHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job", map[string]interface{}{"title": "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastCreateUserID)
}

func TestJobHandler_Create_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/job", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Create_ValidationFail_MissingTitle(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 缺少 required 字段 title
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job", map[string]interface{}{"content": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Create_ValidationFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// status binding:"oneof=0 1"，status=2 触发失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job", map[string]interface{}{"title": "x", "status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.createResult = nil
	env.svc.createErr = errors.New("db insert fail")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job", map[string]interface{}{"title": "Java工程师", "hiring_count": 1})

	// CodeJobPublishError=2803
	assert.Equal(t, 2803, resp.Code)
	assert.Equal(t, "db insert fail", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Update 测试 ====================
// =====================================================================

func TestJobHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	title := "更新后的标题"
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3", map[string]interface{}{"title": title})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastUpdateID)
	assert.Equal(t, uint(7), env.svc.lastUpdateOpID)
	require.NotNil(t, env.svc.lastUpdateReq)
	assert.Equal(t, title, env.svc.lastUpdateReq.Title)
}

func TestJobHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestJobHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastUpdateID)
}

func TestJobHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/job/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Update_ValidationFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// status binding:"omitempty,oneof=0 1 2 3"，status=5 触发失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3", map[string]interface{}{"status": 5})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.updateErr = errors.New("无权操作此职位")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3", map[string]interface{}{"title": "x"})

	// CodeJobError=2801
	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "无权操作此职位", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Delete 测试 ====================
// =====================================================================

func TestJobHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastDeleteID)
	assert.Equal(t, uint(7), env.svc.lastDeleteOpID)
}

func TestJobHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestJobHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.deleteErr = errors.New("职位不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "职位不存在", resp.Message)
}

// =====================================================================
// ==================== JobHandler.GetByID 测试 ====================
// =====================================================================

func TestJobHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.svc.lastGetByIDID)
	assert.Equal(t, uint(7), env.svc.lastGetByIDUser)
	var info dto.JobDetailResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "Java高级工程师", info.Title)
}

func TestJobHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastGetByIDID)
}

func TestJobHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.getByIDResult = nil
	env.svc.getByIDErr = errors.New("职位不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/999", nil)

	// CodeJobNotFound=2802
	assert.Equal(t, 2802, resp.Code)
	assert.Equal(t, "职位不存在", resp.Message)
}

// =====================================================================
// ==================== JobHandler.List 测试 ====================
// =====================================================================

func TestJobHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job?page=1&page_size=10&keyword=Java", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.svc.lastListRegionID)
	require.NotNil(t, env.svc.lastListReq)
	assert.Equal(t, "Java", env.svc.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.JobInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestJobHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 1, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/job", nil)

	assert.Equal(t, uint(8), env.svc.lastListRegionID)
}

func TestJobHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.listResult = nil
	env.svc.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.ListNearby 测试 ====================
// =====================================================================

func TestJobHandler_ListNearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/nearby?latitude=30.5&longitude=114.3&radius_km=5&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.svc.lastNearbyRegionID)
	require.NotNil(t, env.svc.lastNearbyReq)
	assert.Equal(t, 30.5, env.svc.lastNearbyReq.Latitude)
	assert.Equal(t, 114.3, env.svc.lastNearbyReq.Longitude)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestJobHandler_ListNearby_BindFail_MissingLatLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// latitude/longitude binding:"required"，缺省触发 bind 失败
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/nearby", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_ListNearby_InvalidLatLng(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// 纬度越界（>90）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/nearby?latitude=200&longitude=114.3", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestJobHandler_ListNearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.nearbyResult = nil
	env.svc.nearbyErr = errors.New("postgis boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/nearby?latitude=30.5&longitude=114.3", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "postgis boom", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Search 测试 ====================
// =====================================================================

func TestJobHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/search?keyword=Java&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.svc.lastSearchRegionID)
	require.NotNil(t, env.svc.lastSearchReq)
	assert.Equal(t, "Java", env.svc.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestJobHandler_Search_BindFail_EmptyKeyword(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// keyword binding:"required"，空触发 bind 失败
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/search", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.searchResult = nil
	env.svc.searchErr = errors.New("es down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/search?keyword=Java", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "es down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.AdvancedSearch 测试 ====================
// =====================================================================

func TestJobHandler_AdvancedSearch_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/advanced-search?keyword=Java&salary_min=10000&allow_remote=true", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.svc.lastAdvancedRegionID)
	require.NotNil(t, env.svc.lastAdvancedReq)
	assert.Equal(t, "Java", env.svc.lastAdvancedReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestJobHandler_AdvancedSearch_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.advancedResult = nil
	env.svc.advancedErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/advanced-search?keyword=Java", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.ListMine 测试 ====================
// =====================================================================

func TestJobHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/mine?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.svc.lastListMineUserID)
	assert.Equal(t, 1, env.svc.lastListMinePage)
	assert.Equal(t, 10, env.svc.lastListMineSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestJobHandler_ListMine_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastListMineUserID)
}

func TestJobHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.listMineResult = nil
	env.svc.listMineErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/mine", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.ListSimilar 测试 ====================
// =====================================================================

func TestJobHandler_ListSimilar_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/3/similar?limit=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.svc.lastSimilarID)
	assert.Equal(t, 5, env.svc.lastSimilarLimit)
	var list []dto.SimilarJobResponse
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, uint(2), list[0].JobID)
	assert.Equal(t, "高级Java", list[0].Title)
}

func TestJobHandler_ListSimilar_DefaultLimit(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.doJSON(t, http.MethodGet, "/api/v1/job/3/similar", nil)

	// 未传 limit 走 DefaultQuery("limit","10")
	assert.Equal(t, 10, env.svc.lastSimilarLimit)
}

func TestJobHandler_ListSimilar_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/abc/similar", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_ListSimilar_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.svc.similarResult = nil
	env.svc.similarErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/3/similar", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.UpdateStatus 测试 ====================
// =====================================================================

func TestJobHandler_UpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastUpdateStatusID)
	assert.Equal(t, uint(7), env.svc.lastUpdateStatusOp)
	assert.Equal(t, 2, env.svc.lastUpdateStatusVal)
}

func TestJobHandler_UpdateStatus_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestJobHandler_UpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/abc/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_UpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/job/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_UpdateStatus_ValidationFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// status binding:"oneof=1 2 3"，status=9 触发失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3/status", map[string]interface{}{"status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_UpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.updateStatusErr = errors.New("无权操作此职位")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "无权操作此职位", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Fav 测试（收藏） ====================
// =====================================================================

func TestJobHandler_Fav_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(7), env.svc.lastFavUserID)
	assert.Equal(t, uint(3), env.svc.lastFavJobID)
	var res dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &res))
	assert.True(t, res.HasFaved)
	assert.Equal(t, 5, res.FavCount)
}

func TestJobHandler_Fav_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastFavUserID)
}

func TestJobHandler_Fav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_Fav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.favResult = nil
	env.svc.favErr = errors.New("已收藏过")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "已收藏过", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Unfav 测试（取消收藏） ====================
// =====================================================================

func TestJobHandler_Unfav_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
	assert.Equal(t, uint(7), env.svc.lastUnfavUserID)
	assert.Equal(t, uint(3), env.svc.lastUnfavJobID)
	var res dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &res))
	assert.False(t, res.HasFaved)
}

func TestJobHandler_Unfav_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestJobHandler_Unfav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_Unfav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.unfavResult = nil
	env.svc.unfavErr = errors.New("db fail")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// =====================================================================
// ==================== JobHandler.FavStatus 测试 ====================
// =====================================================================

func TestJobHandler_FavStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/3/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.svc.lastFavStatusUserID)
	assert.Equal(t, uint(3), env.svc.lastFavStatusJobID)
	var res dto.FavResponse
	require.NoError(t, json.Unmarshal(resp.Data, &res))
	assert.True(t, res.HasFaved)
}

func TestJobHandler_FavStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_FavStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.favStatusResult = nil
	env.svc.favStatusErr = errors.New("职位不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/999/fav", nil)

	// FavStatus 失败走 CodeJobNotFound=2802
	assert.Equal(t, 2802, resp.Code)
	assert.Equal(t, "职位不存在", resp.Message)
}

// =====================================================================
// ==================== JobHandler.ListFavs 测试 ====================
// =====================================================================

func TestJobHandler_ListFavs_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/favorites?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.svc.lastListFavsUserID)
	assert.Equal(t, 1, env.svc.lastListFavsPage)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestJobHandler_ListFavs_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/favorites", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.svc.lastListFavsUserID)
}

func TestJobHandler_ListFavs_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.listFavsResult = nil
	env.svc.listFavsErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/favorites", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Promotion 测试（推广） ====================
// =====================================================================

func TestJobHandler_Promotion_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	body := map[string]interface{}{
		"promotion_level": 3,
		"traffic_weight":  5.5,
		"is_top":          true,
		"top_days":        7,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/promotion", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "推广设置成功", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastPromotionID)
	assert.Equal(t, uint(7), env.svc.lastPromotionOp)
	require.NotNil(t, env.svc.lastPromotionReq)
	assert.Equal(t, 3, env.svc.lastPromotionReq.PromotionLevel)
	assert.True(t, env.svc.lastPromotionReq.IsTop)
	assert.Equal(t, 7, env.svc.lastPromotionReq.TopDays)
}

func TestJobHandler_Promotion_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/promotion", map[string]interface{}{"promotion_level": 1})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestJobHandler_Promotion_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/abc/promotion", map[string]interface{}{"promotion_level": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_Promotion_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/job/3/promotion", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Promotion_ValidationFail_LevelOverflow(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// promotion_level binding:"gte=0,lte=10"，level=11 触发失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/promotion", map[string]interface{}{"promotion_level": 11})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Promotion_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.svc.promotionErr = errors.New("无权操作此职位")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/job/3/promotion", map[string]interface{}{"promotion_level": 1})

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "无权操作此职位", resp.Message)
}

// =====================================================================
// ==================== JobHandler.AdminList 测试 ====================
// =====================================================================

func TestJobHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/admin/list?page=1&page_size=10&status=1&keyword=Java", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.svc.lastAdminListReq)
	assert.Equal(t, "Java", env.svc.lastAdminListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.JobInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestJobHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.svc.adminListResult = nil
	env.svc.adminListErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/admin/list", nil)

	assert.Equal(t, 2801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== JobHandler.AdminGetByID 测试 ====================
// =====================================================================

func TestJobHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/admin/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.svc.lastAdminGetID)
	var info dto.JobDetailResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "Java高级工程师", info.Title)
}

func TestJobHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/admin/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.svc.adminGetResult = nil
	env.svc.adminGetErr = errors.New("职位不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/job/admin/999", nil)

	// CodeJobNotFound=2802
	assert.Equal(t, 2802, resp.Code)
	assert.Equal(t, "职位不存在", resp.Message)
}

// =====================================================================
// ==================== JobHandler.Audit 测试（审核） ====================
// =====================================================================

func TestJobHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/audit", map[string]interface{}{
		"audit_status": 1,
		"audit_reason": "内容合规",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastAuditID)
	assert.Equal(t, 1, env.svc.lastAuditStatus)
	assert.Equal(t, "内容合规", env.svc.lastAuditReason)
}

func TestJobHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/abc/audit", map[string]interface{}{"audit_status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_Audit_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/job/admin/3/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Audit_ValidationFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// audit_status binding:"oneof=0 1 2"，status=9 触发失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/audit", map[string]interface{}{"audit_status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.svc.auditErr = errors.New("已审核的职位不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/audit", map[string]interface{}{"audit_status": 1})

	// CodeJobAuditError=2804
	assert.Equal(t, 2804, resp.Code)
	assert.Equal(t, "已审核的职位不能重复审核", resp.Message)
}

// =====================================================================
// ==================== JobHandler.AdminUpdateStatus 测试 ====================
// =====================================================================

func TestJobHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/status", map[string]interface{}{"status": 3})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(3), env.svc.lastAdminUpdateID)
	assert.Equal(t, 3, env.svc.lastAdminUpdateVal)
}

func TestJobHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/abc/status", map[string]interface{}{"status": 3})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestJobHandler_AdminUpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/job/admin/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_AdminUpdateStatus_ValidationFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// status binding:"oneof=1 2 3 4"，status=9 触发失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/status", map[string]interface{}{"status": 9})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestJobHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.svc.adminUpdateErr = errors.New("db fail")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/job/admin/3/status", map[string]interface{}{"status": 3})

	// CodeJobAuditError=2804
	assert.Equal(t, 2804, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}
