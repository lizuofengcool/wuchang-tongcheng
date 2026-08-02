// Package handler_test 商家模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 handler 全部分支：
//   - 公开接口无需登录（List/GetByID/GetImages/GetReviews 在 userID=0 时不被 401 拦截）
//   - 用户接口未登录拦截（Apply/GetMyShop/UpdateMyShop/AddImage/DeleteImage/CreateReview → 401 "请先登录"）
//   - 管理接口鉴权由 RequirePermission 中间件负责（测试中去掉，纯测 handler 装配层）
//   - URL :id 参数解析失败（非数字 → 400，区分"无效的店铺ID"/"无效的图片ID"/"无效的评价ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/max → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeShopError=2501/CodeShopNotFound=2502/CodeShopAuditError=2504/CodeShopReviewError=2506/CodeShopApplyError=2507 + message + data 透传）
//   - 地区ID 上下文注入（regionID 透传给 service）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 category/region/news/file/setting/permission/house/groupbuy handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/shop/dto"
	shopHandler "wuchang-tongcheng/internal/modules/shop/handler"
	"wuchang-tongcheng/internal/modules/shop/service"
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

// mockShopService 内存 mock，实现 service.ShopService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockShopService struct {
	// GetByID
	lastGetByIDID uint
	getByIDResult *dto.ShopInfo
	getByIDErr    error

	// List
	lastListRegionID uint
	lastListReq      *dto.ShopListRequest
	listResult       []dto.ShopInfo
	listErr          error

	// GetImages
	lastGetImagesShopID uint
	getImagesResult     []dto.ShopImageInfo
	getImagesErr        error

	// GetReviews
	lastGetReviewsShopID uint
	lastGetReviewsReq    *dto.ReviewListRequest
	getReviewsResult     []dto.ShopReviewInfo
	getReviewsErr        error

	// Apply
	lastApplyRegionID uint
	lastApplyUserID   uint
	lastApplyReq      *dto.ApplyShopRequest
	applyResult       *dto.ShopInfo
	applyErr          error

	// GetMyShop
	lastGetMyShopUserID   uint
	lastGetMyShopRegionID uint
	getMyShopResult       *dto.ShopInfo
	getMyShopErr          error

	// UpdateMyShop
	lastUpdateMyShopUserID   uint
	lastUpdateMyShopRegionID uint
	lastUpdateMyShopReq      *dto.UpdateShopRequest
	updateMyShopErr          error

	// AddImage
	lastAddImageUserID   uint
	lastAddImageRegionID uint
	lastAddImageReq      *dto.AddShopImageRequest
	addImageResult       *dto.ShopImageInfo
	addImageErr          error

	// DeleteImage
	lastDeleteImageUserID uint
	lastDeleteImageID     uint
	deleteImageErr        error

	// CreateReview
	lastCreateReviewRegionID uint
	lastCreateReviewUserID   uint
	lastCreateReviewShopID   uint
	lastCreateReviewReq      *dto.CreateReviewRequest
	createReviewResult       *dto.ShopReviewInfo
	createReviewErr          error

	// AdminList
	lastAdminListRegionID uint
	lastAdminListReq      *dto.AdminShopListRequest
	adminListResult       []dto.ShopInfo
	adminListErr          error

	// AuditShop
	lastAuditShopID  uint
	lastAuditShopReq *dto.AuditShopRequest
	auditShopErr     error

	// UpdateShopStatus
	lastUpdateShopStatusID  uint
	lastUpdateShopStatusReq *dto.UpdateShopStatusRequest
	updateShopStatusErr     error

	// SetRecommend
	lastSetRecommendID  uint
	lastSetRecommendReq *dto.SetRecommendRequest
	setRecommendErr     error

	// DeleteShop
	lastDeleteShopID uint
	deleteShopErr    error

	// AdminReviewList
	lastAdminReviewListReq *dto.AdminReviewListRequest
	adminReviewListResult  []dto.ShopReviewInfo
	adminReviewListErr     error

	// AuditReview
	lastAuditReviewID  uint
	lastAuditReviewReq *dto.AuditReviewRequest
	auditReviewErr     error
}

// ===== 公开接口 =====

func (m *mockShopService) GetByID(id uint) (*dto.ShopInfo, error) {
	m.lastGetByIDID = id
	return m.getByIDResult, m.getByIDErr
}

func (m *mockShopService) List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.listResult))
	return pagination, m.listResult, nil
}

func (m *mockShopService) GetImages(shopID uint) ([]dto.ShopImageInfo, error) {
	m.lastGetImagesShopID = shopID
	return m.getImagesResult, m.getImagesErr
}

func (m *mockShopService) GetReviews(shopID uint, req *dto.ReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error) {
	m.lastGetReviewsShopID = shopID
	m.lastGetReviewsReq = req
	if m.getReviewsErr != nil {
		return nil, nil, m.getReviewsErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.getReviewsResult))
	return pagination, m.getReviewsResult, nil
}

// ===== 用户接口 =====

func (m *mockShopService) Apply(regionID uint, userID uint, req *dto.ApplyShopRequest) (*dto.ShopInfo, error) {
	m.lastApplyRegionID = regionID
	m.lastApplyUserID = userID
	m.lastApplyReq = req
	return m.applyResult, m.applyErr
}

func (m *mockShopService) GetMyShop(userID uint, regionID uint) (*dto.ShopInfo, error) {
	m.lastGetMyShopUserID = userID
	m.lastGetMyShopRegionID = regionID
	return m.getMyShopResult, m.getMyShopErr
}

func (m *mockShopService) UpdateMyShop(userID uint, regionID uint, req *dto.UpdateShopRequest) error {
	m.lastUpdateMyShopUserID = userID
	m.lastUpdateMyShopRegionID = regionID
	m.lastUpdateMyShopReq = req
	return m.updateMyShopErr
}

func (m *mockShopService) AddImage(userID uint, regionID uint, req *dto.AddShopImageRequest) (*dto.ShopImageInfo, error) {
	m.lastAddImageUserID = userID
	m.lastAddImageRegionID = regionID
	m.lastAddImageReq = req
	return m.addImageResult, m.addImageErr
}

func (m *mockShopService) DeleteImage(userID uint, imageID uint) error {
	m.lastDeleteImageUserID = userID
	m.lastDeleteImageID = imageID
	return m.deleteImageErr
}

func (m *mockShopService) CreateReview(regionID uint, userID uint, shopID uint, req *dto.CreateReviewRequest) (*dto.ShopReviewInfo, error) {
	m.lastCreateReviewRegionID = regionID
	m.lastCreateReviewUserID = userID
	m.lastCreateReviewShopID = shopID
	m.lastCreateReviewReq = req
	return m.createReviewResult, m.createReviewErr
}

// ===== 管理接口 =====

func (m *mockShopService) AdminList(regionID uint, req *dto.AdminShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastAdminListRegionID = regionID
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.adminListResult))
	return pagination, m.adminListResult, nil
}

func (m *mockShopService) AuditShop(id uint, req *dto.AuditShopRequest) error {
	m.lastAuditShopID = id
	m.lastAuditShopReq = req
	return m.auditShopErr
}

func (m *mockShopService) UpdateShopStatus(id uint, req *dto.UpdateShopStatusRequest) error {
	m.lastUpdateShopStatusID = id
	m.lastUpdateShopStatusReq = req
	return m.updateShopStatusErr
}

func (m *mockShopService) SetRecommend(id uint, req *dto.SetRecommendRequest) error {
	m.lastSetRecommendID = id
	m.lastSetRecommendReq = req
	return m.setRecommendErr
}

func (m *mockShopService) DeleteShop(id uint) error {
	m.lastDeleteShopID = id
	return m.deleteShopErr
}

func (m *mockShopService) AdminReviewList(req *dto.AdminReviewListRequest) (*utils.Pagination, []dto.ShopReviewInfo, error) {
	m.lastAdminReviewListReq = req
	if m.adminReviewListErr != nil {
		return nil, nil, m.adminReviewListErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.adminReviewListResult))
	return pagination, m.adminReviewListResult, nil
}

func (m *mockShopService) AuditReview(id uint, req *dto.AuditReviewRequest) error {
	m.lastAuditReviewID = id
	m.lastAuditReviewReq = req
	return m.auditReviewErr
}

// 确保 mockShopService 实现 service.ShopService 接口
var _ service.ShopService = (*mockShopService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockShopService
}

// newHandlerEnv 构造 gin 引擎并注册 shop 路由（路径与 shop/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockShopService{
		getByIDResult:        &dto.ShopInfo{ID: 1, Name: "张三便利店", Phone: "13800000000", Status: 1, AuditStatus: 1, RegionID: regionID},
		listResult:           []dto.ShopInfo{{ID: 1, Name: "张三便利店", Status: 1, AuditStatus: 1}, {ID: 2, Name: "李四水果店", Status: 1, AuditStatus: 1}},
		getImagesResult:      []dto.ShopImageInfo{{ID: 1, ShopID: 1, ImageURL: "https://cdn.example.com/1.png", Sort: 1}},
		getReviewsResult:     []dto.ShopReviewInfo{{ID: 1, ShopID: 1, UserID: 1, Rating: 5, Content: "很好", Status: 1}},
		applyResult:          &dto.ShopInfo{ID: 1, Name: "新店", UserID: ctxUserID, Status: 1, AuditStatus: 0},
		getMyShopResult:      &dto.ShopInfo{ID: 1, Name: "我的店铺", UserID: ctxUserID, Status: 1, AuditStatus: 1},
		addImageResult:       &dto.ShopImageInfo{ID: 1, ShopID: 1, ImageURL: "https://cdn.example.com/new.png", Sort: 1},
		createReviewResult:   &dto.ShopReviewInfo{ID: 1, ShopID: 1, UserID: ctxUserID, Rating: 5, Content: "好评", Status: 0},
		adminListResult:      []dto.ShopInfo{{ID: 1, Name: "张三便利店", Status: 1, AuditStatus: 1}},
		adminReviewListResult: []dto.ShopReviewInfo{{ID: 1, ShopID: 1, UserID: 1, Rating: 5, Content: "很好", Status: 1}},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := shopHandler.NewHandler(mock)
	// 注册路由，路径与 shop/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/shop")
	// 公开接口
	root.GET("/list", h.List)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/images", h.GetImages)
	root.GET("/:id/reviews", h.GetReviews)
	// 用户接口（去掉 AuthRequired 中间件，由 handler 内 userID==0 拦截）
	root.POST("/apply", h.Apply)
	root.GET("/my", h.GetMyShop)
	root.PUT("/my", h.UpdateMyShop)
	root.POST("/my/images", h.AddImage)
	root.DELETE("/my/images/:id", h.DeleteImage)
	root.POST("/:id/reviews", h.CreateReview)
	// 管理接口（去掉 RequirePermission 中间件）
	root.GET("/admin/list", h.AdminList)
	root.PUT("/admin/:id/audit", h.AuditShop)
	root.PUT("/admin/:id/status", h.UpdateShopStatus)
	root.PUT("/admin/:id/recommend", h.SetRecommend)
	root.DELETE("/admin/:id", h.DeleteShop)
	root.GET("/admin/reviews", h.AdminReviewList)
	root.PUT("/admin/reviews/:id/audit", h.AuditReview)

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

// ==================== 公开接口 ====================

// ---------- List ----------

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/list?page=1&page_size=10&keyword=便利店", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, 1, env.mock.lastListReq.Page)
	assert.Equal(t, 10, env.mock.lastListReq.PageSize)
	assert.Equal(t, "便利店", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.ShopInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "张三便利店", list[0].Name)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/list", nil)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 5, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/shop/list", nil)

	assert.Equal(t, uint(8), env.mock.lastListRegionID)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	var info dto.ShopInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "张三便利店", info.Name)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/999", nil)

	// CodeShopNotFound=2502
	assert.Equal(t, 2502, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- GetImages ----------

func TestHandler_GetImages_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/3/images", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetImagesShopID)
	var list []dto.ShopImageInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "https://cdn.example.com/1.png", list[0].ImageURL)
}

func TestHandler_GetImages_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/abc/images", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetImagesShopID)
}

func TestHandler_GetImages_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.getImagesResult = nil
	env.mock.getImagesErr = errors.New("redis boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/1/images", nil)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// ---------- GetReviews ----------

func TestHandler_GetReviews_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/3/reviews?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetReviewsShopID)
	require.NotNil(t, env.mock.lastGetReviewsReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.ShopReviewInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, 5, list[0].Rating)
}

func TestHandler_GetReviews_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/abc/reviews", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetReviewsShopID)
}

func TestHandler_GetReviews_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.getReviewsResult = nil
	env.mock.getReviewsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/1/reviews", nil)

	// CodeShopReviewError=2506
	assert.Equal(t, 2506, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- 公开读取无需登录聚合 ----------

func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	// userID=0 时四个公开读路径均不被 401 拦截
	env := newHandlerEnv(t, 0, 2)

	r1 := env.doJSON(t, http.MethodGet, "/api/v1/shop/list", nil)
	assert.Equal(t, 0, r1.Code)

	r2 := env.doJSON(t, http.MethodGet, "/api/v1/shop/1", nil)
	assert.Equal(t, 0, r2.Code)

	r3 := env.doJSON(t, http.MethodGet, "/api/v1/shop/1/images", nil)
	assert.Equal(t, 0, r3.Code)

	r4 := env.doJSON(t, http.MethodGet, "/api/v1/shop/1/reviews", nil)
	assert.Equal(t, 0, r4.Code)
}

// ==================== 用户接口 ====================

// ---------- Apply ----------

func TestHandler_Apply_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.ApplyShopRequest{Name: "新店", Phone: "13800000000", Address: "五常市", CategoryID: 5}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/apply", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "申请成功，请等待审核", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastApplyRegionID)
	assert.Equal(t, uint(1), env.mock.lastApplyUserID)
	require.NotNil(t, env.mock.lastApplyReq)
	assert.Equal(t, "新店", env.mock.lastApplyReq.Name)
	assert.Equal(t, "13800000000", env.mock.lastApplyReq.Phone)
	assert.Equal(t, uint(5), env.mock.lastApplyReq.CategoryID)
	var info dto.ShopInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_Apply_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	body := dto.ApplyShopRequest{Name: "新店", Phone: "13800000000"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/apply", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastApplyReq)
}

func TestHandler_Apply_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/apply", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastApplyReq)
}

func TestHandler_Apply_BindError_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 name（required）+ phone（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/apply", `{"description":"x"}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastApplyReq)
}

func TestHandler_Apply_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.applyResult = nil
	env.mock.applyErr = errors.New("您已申请过店铺")
	body := dto.ApplyShopRequest{Name: "新店", Phone: "13800000000"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/apply", body)

	// CodeShopApplyError=2507
	assert.Equal(t, 2507, resp.Code)
	assert.Equal(t, "您已申请过店铺", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastApplyRegionID)
}

// ---------- GetMyShop ----------

func TestHandler_GetMyShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/my", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastGetMyShopUserID)
	assert.Equal(t, uint(2), env.mock.lastGetMyShopRegionID)
	var info dto.ShopInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "我的店铺", info.Name)
}

func TestHandler_GetMyShop_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/my", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetMyShopUserID)
}

func TestHandler_GetMyShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getMyShopResult = nil
	env.mock.getMyShopErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/my", nil)

	// CodeShopNotFound=2502
	assert.Equal(t, 2502, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- UpdateMyShop ----------

func TestHandler_UpdateMyShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateShopRequest{Name: "新名字", Phone: "13900000000", Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/my", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastUpdateMyShopUserID)
	assert.Equal(t, uint(2), env.mock.lastUpdateMyShopRegionID)
	require.NotNil(t, env.mock.lastUpdateMyShopReq)
	assert.Equal(t, "新名字", env.mock.lastUpdateMyShopReq.Name)
	assert.Equal(t, 1, env.mock.lastUpdateMyShopReq.Status)
}

func TestHandler_UpdateMyShop_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/my", dto.UpdateShopRequest{Name: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastUpdateMyShopReq)
}

func TestHandler_UpdateMyShop_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/my", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastUpdateMyShopReq)
}

func TestHandler_UpdateMyShop_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=9 不满足 oneof=0 1 2
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/my", `{"name":"x","status":9}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastUpdateMyShopReq)
}

func TestHandler_UpdateMyShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateMyShopErr = errors.New("无权操作此店铺")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/my", dto.UpdateShopRequest{Name: "x"})

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "无权操作此店铺", resp.Message)
}

// ---------- AddImage ----------

func TestHandler_AddImage_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.AddShopImageRequest{ImageURL: "https://cdn.example.com/new.png", Sort: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/my/images", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "上传成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastAddImageUserID)
	assert.Equal(t, uint(2), env.mock.lastAddImageRegionID)
	require.NotNil(t, env.mock.lastAddImageReq)
	assert.Equal(t, "https://cdn.example.com/new.png", env.mock.lastAddImageReq.ImageURL)
	var info dto.ShopImageInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_AddImage_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/my/images", dto.AddShopImageRequest{ImageURL: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastAddImageReq)
}

func TestHandler_AddImage_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/my/images", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastAddImageReq)
}

func TestHandler_AddImage_BindError_MissingImageURL(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 image_url（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/my/images", `{"sort":1}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastAddImageReq)
}

func TestHandler_AddImage_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.addImageResult = nil
	env.mock.addImageErr = errors.New("店铺未审核通过")
	body := dto.AddShopImageRequest{ImageURL: "https://cdn.example.com/new.png"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/my/images", body)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "店铺未审核通过", resp.Message)
}

// ---------- DeleteImage ----------

func TestHandler_DeleteImage_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/my/images/5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.mock.lastDeleteImageUserID)
	assert.Equal(t, uint(5), env.mock.lastDeleteImageID)
}

func TestHandler_DeleteImage_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/my/images/5", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteImageUserID)
}

func TestHandler_DeleteImage_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/my/images/abc", nil)

	// :id 为图片ID，消息为"无效的图片ID"
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的图片ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteImageID)
}

func TestHandler_DeleteImage_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.deleteImageErr = errors.New("店铺图片不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/my/images/5", nil)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "店铺图片不存在", resp.Message)
}

// ---------- CreateReview ----------

func TestHandler_CreateReview_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.CreateReviewRequest{Rating: 5, Content: "好评"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/3/reviews", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "评价成功，请等待审核", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateReviewRegionID)
	assert.Equal(t, uint(1), env.mock.lastCreateReviewUserID)
	assert.Equal(t, uint(3), env.mock.lastCreateReviewShopID)
	require.NotNil(t, env.mock.lastCreateReviewReq)
	assert.Equal(t, 5, env.mock.lastCreateReviewReq.Rating)
	var info dto.ShopReviewInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_CreateReview_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/3/reviews", dto.CreateReviewRequest{Rating: 5})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReviewReq)
}

func TestHandler_CreateReview_InvalidShopID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/abc/reviews", dto.CreateReviewRequest{Rating: 5})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCreateReviewShopID)
}

func TestHandler_CreateReview_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/3/reviews", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReviewReq)
}

func TestHandler_CreateReview_BindError_InvalidRating(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// rating=0 不满足 required,min=1
	resp := env.doRaw(t, http.MethodPost, "/api/v1/shop/3/reviews", `{"rating":0,"content":"x"}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateReviewReq)
}

func TestHandler_CreateReview_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.createReviewResult = nil
	env.mock.createReviewErr = errors.New("店铺未审核通过")
	body := dto.CreateReviewRequest{Rating: 5, Content: "好评"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/shop/3/reviews", body)

	// CodeShopReviewError=2506
	assert.Equal(t, 2506, resp.Code)
	assert.Equal(t, "店铺未审核通过", resp.Message)
}

// ==================== 管理接口 ====================

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/admin/list?page=1&page_size=10&audit_status=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastAdminListRegionID)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, 1, env.mock.lastAdminListReq.AuditStatus)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.ShopInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/admin/list", nil)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- AuditShop ----------

func TestHandler_AuditShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.AuditShopRequest{AuditStatus: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastAuditShopID)
	require.NotNil(t, env.mock.lastAuditShopReq)
	assert.Equal(t, 1, env.mock.lastAuditShopReq.AuditStatus)
}

func TestHandler_AuditShop_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/abc/audit", dto.AuditShopRequest{AuditStatus: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAuditShopID)
}

func TestHandler_AuditShop_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/5/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAuditShopID)
}

func TestHandler_AuditShop_BindError_InvalidAuditStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// audit_status=0 不满足 oneof=1 2
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/5/audit", `{"audit_status":0}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_AuditShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.auditShopErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/audit", dto.AuditShopRequest{AuditStatus: 1})

	// CodeShopAuditError=2504
	assert.Equal(t, 2504, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- UpdateShopStatus ----------

func TestHandler_UpdateShopStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateShopStatusRequest{Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateShopStatusID)
	require.NotNil(t, env.mock.lastUpdateShopStatusReq)
	assert.Equal(t, 1, env.mock.lastUpdateShopStatusReq.Status)
}

func TestHandler_UpdateShopStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/abc/status", dto.UpdateShopStatusRequest{Status: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateShopStatusID)
}

func TestHandler_UpdateShopStatus_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=9 不满足 oneof=0 1 2
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/5/status", `{"status":9}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateShopStatusID)
}

func TestHandler_UpdateShopStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateShopStatusErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/status", dto.UpdateShopStatusRequest{Status: 1})

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- SetRecommend ----------

func TestHandler_SetRecommend_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.SetRecommendRequest{IsRecommend: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/recommend", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "设置成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastSetRecommendID)
	require.NotNil(t, env.mock.lastSetRecommendReq)
	assert.Equal(t, 1, env.mock.lastSetRecommendReq.IsRecommend)
}

func TestHandler_SetRecommend_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/abc/recommend", dto.SetRecommendRequest{IsRecommend: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastSetRecommendID)
}

func TestHandler_SetRecommend_BindError_InvalidIsRecommend(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// is_recommend=9 不满足 oneof=0 1
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/5/recommend", `{"is_recommend":9}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastSetRecommendID)
}

func TestHandler_SetRecommend_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.setRecommendErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/5/recommend", dto.SetRecommendRequest{IsRecommend: 1})

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- DeleteShop ----------

func TestHandler_DeleteShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/admin/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteShopID)
}

func TestHandler_DeleteShop_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/admin/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteShopID)
}

func TestHandler_DeleteShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.deleteShopErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/shop/admin/7", nil)

	// CodeShopError=2501
	assert.Equal(t, 2501, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- AdminReviewList ----------

func TestHandler_AdminReviewList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/admin/reviews?page=1&page_size=10&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminReviewListReq)
	assert.Equal(t, 1, env.mock.lastAdminReviewListReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.ShopReviewInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_AdminReviewList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.adminReviewListResult = nil
	env.mock.adminReviewListErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/shop/admin/reviews", nil)

	// CodeShopReviewError=2506
	assert.Equal(t, 2506, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- AuditReview ----------

func TestHandler_AuditReview_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.AuditReviewRequest{Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/reviews/5/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastAuditReviewID)
	require.NotNil(t, env.mock.lastAuditReviewReq)
	assert.Equal(t, 1, env.mock.lastAuditReviewReq.Status)
}

func TestHandler_AuditReview_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/reviews/abc/audit", dto.AuditReviewRequest{Status: 1})

	// :id 为评价ID，消息为"无效的评价ID"
	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的评价ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAuditReviewID)
}

func TestHandler_AuditReview_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/reviews/5/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastAuditReviewID)
}

func TestHandler_AuditReview_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=9 不满足 oneof=1 2
	resp := env.doRaw(t, http.MethodPut, "/api/v1/shop/admin/reviews/5/audit", `{"status":9}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_AuditReview_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.auditReviewErr = errors.New("评价不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/shop/admin/reviews/5/audit", dto.AuditReviewRequest{Status: 1})

	// CodeShopReviewError=2506
	assert.Equal(t, 2506, resp.Code)
	assert.Equal(t, "评价不存在", resp.Message)
}

// ==================== regionID 注入聚合 ====================

func TestHandler_RegionIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 regionID 的接口均透传 context 中的 regionID
	env := newHandlerEnv(t, 1, 9)

	// 公开 List
	env.doJSON(t, http.MethodGet, "/api/v1/shop/list", nil)
	assert.Equal(t, uint(9), env.mock.lastListRegionID)

	// 用户接口
	env.doJSON(t, http.MethodPost, "/api/v1/shop/apply", dto.ApplyShopRequest{Name: "新店", Phone: "13800000000"})
	assert.Equal(t, uint(9), env.mock.lastApplyRegionID)

	env.doJSON(t, http.MethodGet, "/api/v1/shop/my", nil)
	assert.Equal(t, uint(9), env.mock.lastGetMyShopRegionID)

	env.doJSON(t, http.MethodPut, "/api/v1/shop/my", dto.UpdateShopRequest{Name: "x"})
	assert.Equal(t, uint(9), env.mock.lastUpdateMyShopRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/shop/my/images", dto.AddShopImageRequest{ImageURL: "https://cdn.example.com/x.png"})
	assert.Equal(t, uint(9), env.mock.lastAddImageRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/shop/1/reviews", dto.CreateReviewRequest{Rating: 5})
	assert.Equal(t, uint(9), env.mock.lastCreateReviewRegionID)

	// 管理接口
	env.doJSON(t, http.MethodGet, "/api/v1/shop/admin/list", nil)
	assert.Equal(t, uint(9), env.mock.lastAdminListRegionID)
}
