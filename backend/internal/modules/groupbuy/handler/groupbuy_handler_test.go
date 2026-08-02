// Package handler_test 团购优惠券模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 handler 全部分支：
//   - 公开接口无需登录（List/GetByID/ListCoupons 在 userID=0 时不被 401 拦截）
//   - 用户接口未登录拦截（CreateOrder/MyOrders/GetOrder/CancelOrder/ReceiveCoupon/MyCoupons → 401）
//   - 管理接口 AdminCreate 防御性登录校验（userID=0 → 401）；其余管理接口鉴权由 RequirePermission 中间件负责
//   - URL :id 参数解析失败（非数字 → 400，区分"无效的团购ID"/"无效的订单ID"/"无效的优惠券ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeGroupBuyError=2601/CodeGroupBuyNotFound=2602/CodeCouponError=2605/CodeOrderError=2609/CodeOrderNotFound=2610 + message + data 透传）
//   - 地区ID 上下文注入（regionID 透传给 service）
//   - 状态默认值兜底（公开 List status 越界→1、AdminList status 越界→-1、AdminListCoupons status 越界→-1）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 category/region/news/file/setting/permission handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/groupbuy/dto"
	gbHandler "wuchang-tongcheng/internal/modules/groupbuy/handler"
	"wuchang-tongcheng/internal/modules/groupbuy/service"
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

// mockGroupBuyService 内存 mock，实现 service.Service 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockGroupBuyService struct {
	// 团购商品调用记录
	lastCreateGBRegionID uint
	lastCreateGBUserID   uint
	lastCreateGBReq      *dto.CreateGroupBuyRequest

	lastUpdateGBID  uint
	lastUpdateGBReq *dto.UpdateGroupBuyRequest

	lastDeleteGBID uint

	lastGetGBID uint

	lastListGBRegionID uint
	lastListGBReq      *dto.GroupBuyListRequest

	lastUpdateGBStatusID     uint
	lastUpdateGBStatusStatus int

	lastAuditGBID          uint
	lastAuditGBAuditStatus int

	// 优惠券调用记录
	lastCreateCouponRegionID uint
	lastCreateCouponReq      *dto.CreateCouponRequest

	lastUpdateCouponID  uint
	lastUpdateCouponReq *dto.UpdateCouponRequest

	lastDeleteCouponID uint

	lastListCouponRegionID uint
	lastListCouponReq      *dto.CouponListRequest

	lastAvailableCouponsRegionID uint
	lastAvailableCouponsReq      *dto.CouponListRequest

	lastReceiveCouponRegionID uint
	lastReceiveCouponUserID   uint
	lastReceiveCouponID       uint

	lastMyCouponsUserID uint
	lastMyCouponsReq    *dto.CouponListRequest

	// 订单调用记录
	lastCreateOrderRegionID uint
	lastCreateOrderUserID   uint
	lastCreateOrderReq      *dto.CreateOrderRequest

	lastGetOrderID     uint
	lastGetOrderUserID uint

	lastMyOrdersUserID uint
	lastMyOrdersReq    *dto.OrderListRequest

	lastCancelOrderID     uint
	lastCancelOrderUserID uint

	lastVerifyOrderID     uint
	lastVerifyOrderVerify string

	lastAdminOrderListRegionID uint
	lastAdminOrderListReq      *dto.OrderListRequest

	// 返回值预设
	createGBResult *dto.GroupBuyInfo
	createGBErr    error
	updateGBErr    error
	deleteGBErr    error
	getGBResult    *dto.GroupBuyInfo
	getGBErr       error
	listGBResult   []dto.GroupBuyInfo
	listGBErr      error
	updateGBStatErr error
	auditGBErr     error

	createCouponResult *dto.CouponInfo
	createCouponErr    error
	updateCouponErr    error
	deleteCouponErr    error
	listCouponResult   []dto.CouponInfo
	listCouponErr      error
	availableResult    []dto.CouponInfo
	availableErr       error
	receiveResult      *dto.UserCouponInfo
	receiveErr         error
	myCouponsResult    []dto.UserCouponInfo
	myCouponsErr       error

	createOrderResult *dto.OrderInfo
	createOrderErr    error
	getOrderResult    *dto.OrderInfo
	getOrderErr       error
	myOrdersResult    []dto.OrderInfo
	myOrdersErr       error
	cancelOrderErr    error
	verifyOrderErr    error
	adminOrderListResult []dto.OrderInfo
	adminOrderListErr   error
}

// ===== 团购商品 =====

func (m *mockGroupBuyService) CreateGroupBuy(regionID, userID uint, req *dto.CreateGroupBuyRequest) (*dto.GroupBuyInfo, error) {
	m.lastCreateGBRegionID = regionID
	m.lastCreateGBUserID = userID
	m.lastCreateGBReq = req
	return m.createGBResult, m.createGBErr
}
func (m *mockGroupBuyService) UpdateGroupBuy(id uint, req *dto.UpdateGroupBuyRequest) error {
	m.lastUpdateGBID = id
	m.lastUpdateGBReq = req
	return m.updateGBErr
}
func (m *mockGroupBuyService) DeleteGroupBuy(id uint) error {
	m.lastDeleteGBID = id
	return m.deleteGBErr
}
func (m *mockGroupBuyService) GetGroupBuy(id uint) (*dto.GroupBuyInfo, error) {
	m.lastGetGBID = id
	return m.getGBResult, m.getGBErr
}
func (m *mockGroupBuyService) ListGroupBuy(regionID uint, req *dto.GroupBuyListRequest) (*utils.Pagination, []dto.GroupBuyInfo, error) {
	m.lastListGBRegionID = regionID
	m.lastListGBReq = req
	if m.listGBErr != nil {
		return nil, nil, m.listGBErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.listGBResult))
	return pagination, m.listGBResult, nil
}
func (m *mockGroupBuyService) UpdateGroupBuyStatus(id uint, status int) error {
	m.lastUpdateGBStatusID = id
	m.lastUpdateGBStatusStatus = status
	return m.updateGBStatErr
}
func (m *mockGroupBuyService) AuditGroupBuy(id uint, auditStatus int) error {
	m.lastAuditGBID = id
	m.lastAuditGBAuditStatus = auditStatus
	return m.auditGBErr
}

// ===== 优惠券 =====

func (m *mockGroupBuyService) CreateCoupon(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error) {
	m.lastCreateCouponRegionID = regionID
	m.lastCreateCouponReq = req
	return m.createCouponResult, m.createCouponErr
}
func (m *mockGroupBuyService) UpdateCoupon(id uint, req *dto.UpdateCouponRequest) error {
	m.lastUpdateCouponID = id
	m.lastUpdateCouponReq = req
	return m.updateCouponErr
}
func (m *mockGroupBuyService) DeleteCoupon(id uint) error {
	m.lastDeleteCouponID = id
	return m.deleteCouponErr
}
func (m *mockGroupBuyService) ListCoupon(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	m.lastListCouponRegionID = regionID
	m.lastListCouponReq = req
	if m.listCouponErr != nil {
		return nil, nil, m.listCouponErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.listCouponResult))
	return pagination, m.listCouponResult, nil
}
func (m *mockGroupBuyService) AvailableCoupons(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	m.lastAvailableCouponsRegionID = regionID
	m.lastAvailableCouponsReq = req
	if m.availableErr != nil {
		return nil, nil, m.availableErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.availableResult))
	return pagination, m.availableResult, nil
}
func (m *mockGroupBuyService) ReceiveCoupon(regionID, userID, couponID uint) (*dto.UserCouponInfo, error) {
	m.lastReceiveCouponRegionID = regionID
	m.lastReceiveCouponUserID = userID
	m.lastReceiveCouponID = couponID
	return m.receiveResult, m.receiveErr
}
func (m *mockGroupBuyService) MyCoupons(userID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error) {
	m.lastMyCouponsUserID = userID
	m.lastMyCouponsReq = req
	if m.myCouponsErr != nil {
		return nil, nil, m.myCouponsErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.myCouponsResult))
	return pagination, m.myCouponsResult, nil
}

// ===== 订单 =====

func (m *mockGroupBuyService) CreateOrder(regionID, userID uint, req *dto.CreateOrderRequest) (*dto.OrderInfo, error) {
	m.lastCreateOrderRegionID = regionID
	m.lastCreateOrderUserID = userID
	m.lastCreateOrderReq = req
	return m.createOrderResult, m.createOrderErr
}
func (m *mockGroupBuyService) GetOrder(id, userID uint) (*dto.OrderInfo, error) {
	m.lastGetOrderID = id
	m.lastGetOrderUserID = userID
	return m.getOrderResult, m.getOrderErr
}
func (m *mockGroupBuyService) MyOrders(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	m.lastMyOrdersUserID = userID
	m.lastMyOrdersReq = req
	if m.myOrdersErr != nil {
		return nil, nil, m.myOrdersErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.myOrdersResult))
	return pagination, m.myOrdersResult, nil
}
func (m *mockGroupBuyService) CancelOrder(id, userID uint) error {
	m.lastCancelOrderID = id
	m.lastCancelOrderUserID = userID
	return m.cancelOrderErr
}
func (m *mockGroupBuyService) VerifyOrder(id uint, verifyCode string) error {
	m.lastVerifyOrderID = id
	m.lastVerifyOrderVerify = verifyCode
	return m.verifyOrderErr
}
func (m *mockGroupBuyService) AdminOrderList(regionID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	m.lastAdminOrderListRegionID = regionID
	m.lastAdminOrderListReq = req
	if m.adminOrderListErr != nil {
		return nil, nil, m.adminOrderListErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.adminOrderListResult))
	return pagination, m.adminOrderListResult, nil
}

// 确保 mockGroupBuyService 实现 service.Service 接口
var _ service.Service = (*mockGroupBuyService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockGroupBuyService
}

// newHandlerEnv 构造 gin 引擎并注册 groupbuy 路由（路径与 groupbuy/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockGroupBuyService{
		createGBResult: &dto.GroupBuyInfo{ID: 1, Title: "双人火锅套餐", GroupBuyPrice: 99.9, Stock: 50, Status: 1, AuditStatus: 1, ShopID: 10, UserID: ctxUserID},
		getGBResult:    &dto.GroupBuyInfo{ID: 1, Title: "双人火锅套餐", GroupBuyPrice: 99.9, Stock: 50, Status: 1, AuditStatus: 1},
		listGBResult:   []dto.GroupBuyInfo{{ID: 1, Title: "双人火锅套餐", GroupBuyPrice: 99.9, Status: 1}, {ID: 2, Title: "KTV欢唱", GroupBuyPrice: 58, Status: 1}},
		createCouponResult: &dto.CouponInfo{ID: 1, Name: "满100减20", Type: 1, Value: 20, MinAmount: 100, Status: 1},
		listCouponResult:   []dto.CouponInfo{{ID: 1, Name: "满100减20", Type: 1, Status: 1}},
		availableResult:    []dto.CouponInfo{{ID: 1, Name: "满100减20", Type: 1, Status: 1}},
		receiveResult:      &dto.UserCouponInfo{ID: 1, UserID: ctxUserID, CouponID: 1, Status: 0},
		myCouponsResult:    []dto.UserCouponInfo{{ID: 1, UserID: ctxUserID, CouponID: 1, Status: 0}},
		createOrderResult:  &dto.OrderInfo{ID: 1, OrderNo: "GB20260101120000", UserID: ctxUserID, GroupBuyID: 1, Quantity: 2, PayAmount: 199.8, PayStatus: 1, Status: 1, VerifyCode: "12345678"},
		getOrderResult:     &dto.OrderInfo{ID: 1, OrderNo: "GB20260101120000", UserID: ctxUserID, GroupBuyID: 1, Quantity: 2, PayStatus: 1, Status: 1},
		myOrdersResult:     []dto.OrderInfo{{ID: 1, OrderNo: "GB20260101120000", UserID: ctxUserID, Quantity: 2, Status: 1}},
		adminOrderListResult: []dto.OrderInfo{{ID: 1, OrderNo: "GB20260101120000", UserID: ctxUserID, Quantity: 2, Status: 1}},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	h := gbHandler.NewHandler(mock)
	// 注册路由，路径与 groupbuy/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/groupbuy")
	// 公开接口
	root.GET("/list", h.List)
	root.GET("/coupons", h.ListCoupons)
	root.GET("/:id", h.GetByID)
	// 用户接口
	root.POST("/orders", h.CreateOrder)
	root.GET("/orders/my", h.MyOrders)
	root.GET("/orders/:id", h.GetOrder)
	root.POST("/orders/:id/cancel", h.CancelOrder)
	root.GET("/coupons/my", h.MyCoupons)
	root.POST("/coupons/:id/receive", h.ReceiveCoupon)
	// 核销接口
	root.POST("/orders/:id/verify", h.VerifyOrder)
	// 管理接口 - 团购商品
	root.POST("/admin", h.AdminCreate)
	root.GET("/admin", h.AdminList)
	root.PUT("/admin/:id", h.AdminUpdate)
	root.DELETE("/admin/:id", h.AdminDelete)
	root.PUT("/admin/:id/status", h.AdminUpdateStatus)
	root.PUT("/admin/:id/audit", h.AdminAudit)
	root.GET("/admin/orders", h.AdminOrderList)
	// 管理接口 - 优惠券
	root.POST("/admin/coupons", h.AdminCreateCoupon)
	root.GET("/admin/coupons", h.AdminListCoupons)
	root.PUT("/admin/coupons/:id", h.AdminUpdateCoupon)
	root.DELETE("/admin/coupons/:id", h.AdminDeleteCoupon)

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
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list?page=1&page_size=10&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastListGBRegionID)
	require.NotNil(t, env.mock.lastListGBReq)
	assert.Equal(t, 1, env.mock.lastListGBReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.GroupBuyInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "双人火锅套餐", list[0].Title)
}

func TestHandler_List_StatusOutOfRange_DefaultOn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	// status=99 越界（>2）→ 公开列表兜底为 1（只查上架）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list?status=99", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListGBReq)
	assert.Equal(t, 1, env.mock.lastListGBReq.Status)
}

func TestHandler_List_StatusNegative_DefaultOn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	// status=-1 越界（<0）→ 公开列表兜底为 1
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list?status=-1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListGBReq)
	assert.Equal(t, 1, env.mock.lastListGBReq.Status)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.listGBResult = nil
	env.mock.listGBErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list", nil)

	// CodeGroupBuyError=2601
	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 5, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list", nil)

	assert.Equal(t, uint(8), env.mock.lastListGBRegionID)
}

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastGetGBID)
	var info dto.GroupBuyInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "双人火锅套餐", info.Title)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的团购ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetGBID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.getGBResult = nil
	env.mock.getGBErr = errors.New("团购商品不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/999", nil)

	// CodeGroupBuyNotFound=2602
	assert.Equal(t, 2602, resp.Code)
	assert.Equal(t, "团购商品不存在", resp.Message)
}

// ---------- ListCoupons ----------

func TestHandler_ListCoupons_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons?page=1&page_size=20", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastAvailableCouponsRegionID)
	require.NotNil(t, env.mock.lastAvailableCouponsReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.CouponInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "满100减20", list[0].Name)
}

func TestHandler_ListCoupons_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.availableResult = nil
	env.mock.availableErr = errors.New("redis boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons", nil)

	// CodeCouponError=2605
	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// ---------- 公开读取无需登录聚合 ----------

func TestHandler_PublicRead_NoAuthRequired(t *testing.T) {
	// userID=0 时三个公开读路径均不被 401 拦截
	env := newHandlerEnv(t, 0, 2)

	r1 := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list", nil)
	assert.Equal(t, 0, r1.Code)

	r2 := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/1", nil)
	assert.Equal(t, 0, r2.Code)

	r3 := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons", nil)
	assert.Equal(t, 0, r3.Code)
}

// ==================== 用户接口 ====================

// ---------- CreateOrder ----------

func TestHandler_CreateOrder_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "下单成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateOrderRegionID)
	assert.Equal(t, uint(1), env.mock.lastCreateOrderUserID)
	require.NotNil(t, env.mock.lastCreateOrderReq)
	assert.Equal(t, uint(1), env.mock.lastCreateOrderReq.GroupBuyID)
	assert.Equal(t, 2, env.mock.lastCreateOrderReq.Quantity)
	var info dto.OrderInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "GB20260101120000", info.OrderNo)
}

func TestHandler_CreateOrder_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	body := dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 2}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateOrderReq)
}

func TestHandler_CreateOrder_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/orders", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateOrderReq)
}

func TestHandler_CreateOrder_BindError_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 groupbuy_id（required）+ quantity（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/orders", `{"quantity":1}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateOrderReq)
}

func TestHandler_CreateOrder_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.createOrderResult = nil
	env.mock.createOrderErr = errors.New("库存不足")
	body := dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 100}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders", body)

	// CodeOrderError=2609
	assert.Equal(t, 2609, resp.Code)
	assert.Equal(t, "库存不足", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateOrderRegionID)
}

// ---------- MyOrders ----------

func TestHandler_MyOrders_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/my?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastMyOrdersUserID)
	require.NotNil(t, env.mock.lastMyOrdersReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.OrderInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_MyOrders_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/my", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastMyOrdersUserID)
}

func TestHandler_MyOrders_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.myOrdersResult = nil
	env.mock.myOrdersErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/my", nil)

	assert.Equal(t, 2609, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- GetOrder ----------

func TestHandler_GetOrder_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastGetOrderID)
	assert.Equal(t, uint(1), env.mock.lastGetOrderUserID)
	var info dto.OrderInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "GB20260101120000", info.OrderNo)
}

func TestHandler_GetOrder_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/5", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastGetOrderID)
}

func TestHandler_GetOrder_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的订单ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetOrderID)
}

func TestHandler_GetOrder_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.getOrderResult = nil
	env.mock.getOrderErr = errors.New("订单不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/orders/999", nil)

	// CodeOrderNotFound=2610
	assert.Equal(t, 2610, resp.Code)
	assert.Equal(t, "订单不存在", resp.Message)
}

// ---------- CancelOrder ----------

func TestHandler_CancelOrder_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/5/cancel", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "取消成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastCancelOrderID)
	assert.Equal(t, uint(1), env.mock.lastCancelOrderUserID)
}

func TestHandler_CancelOrder_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/5/cancel", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastCancelOrderID)
}

func TestHandler_CancelOrder_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/xyz/cancel", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的订单ID", resp.Message)
}

func TestHandler_CancelOrder_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.cancelOrderErr = errors.New("订单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/5/cancel", nil)

	assert.Equal(t, 2609, resp.Code)
	assert.Equal(t, "订单状态不允许此操作", resp.Message)
}

// ---------- ReceiveCoupon ----------

func TestHandler_ReceiveCoupon_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/coupons/3/receive", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "领取成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastReceiveCouponRegionID)
	assert.Equal(t, uint(1), env.mock.lastReceiveCouponUserID)
	assert.Equal(t, uint(3), env.mock.lastReceiveCouponID)
	var info dto.UserCouponInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_ReceiveCoupon_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/coupons/3/receive", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastReceiveCouponID)
}

func TestHandler_ReceiveCoupon_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/coupons/abc/receive", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的优惠券ID", resp.Message)
}

func TestHandler_ReceiveCoupon_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.receiveResult = nil
	env.mock.receiveErr = errors.New("优惠券已过期")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/coupons/3/receive", nil)

	// CodeCouponError=2605
	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "优惠券已过期", resp.Message)
}

// ---------- MyCoupons ----------

func TestHandler_MyCoupons_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons/my?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.mock.lastMyCouponsUserID)
	require.NotNil(t, env.mock.lastMyCouponsReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.UserCouponInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_MyCoupons_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons/my", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastMyCouponsUserID)
}

func TestHandler_MyCoupons_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.myCouponsResult = nil
	env.mock.myCouponsErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons/my", nil)

	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- VerifyOrder ----------

func TestHandler_VerifyOrder_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.VerifyOrderRequest{VerifyCode: "12345678"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/5/verify", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "核销成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastVerifyOrderID)
	assert.Equal(t, "12345678", env.mock.lastVerifyOrderVerify)
}

func TestHandler_VerifyOrder_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.VerifyOrderRequest{VerifyCode: "12345678"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/abc/verify", body)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的订单ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastVerifyOrderID)
}

func TestHandler_VerifyOrder_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/orders/5/verify", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_VerifyOrder_BindError_MissingVerifyCode(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 verify_code（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/orders/5/verify", `{}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, "", env.mock.lastVerifyOrderVerify)
}

func TestHandler_VerifyOrder_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.verifyOrderErr = errors.New("核销码错误")
	body := dto.VerifyOrderRequest{VerifyCode: "00000000"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders/5/verify", body)

	assert.Equal(t, 2609, resp.Code)
	assert.Equal(t, "核销码错误", resp.Message)
}

// ==================== 管理接口 - 团购商品 ====================

// ---------- AdminCreate ----------

func TestHandler_AdminCreate_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.CreateGroupBuyRequest{Title: "双人火锅套餐", GroupBuyPrice: 99.9, Stock: 50, ShopID: 10}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateGBRegionID)
	assert.Equal(t, uint(1), env.mock.lastCreateGBUserID)
	require.NotNil(t, env.mock.lastCreateGBReq)
	assert.Equal(t, "双人火锅套餐", env.mock.lastCreateGBReq.Title)
	assert.Equal(t, 99.9, env.mock.lastCreateGBReq.GroupBuyPrice)
	var info dto.GroupBuyInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_AdminCreate_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	body := dto.CreateGroupBuyRequest{Title: "x", GroupBuyPrice: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin", body)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateGBReq)
}

func TestHandler_AdminCreate_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/admin", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateGBReq)
}

func TestHandler_AdminCreate_BindError_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 groupbuy_price（required）
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/admin", `{"title":"x"}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateGBReq)
}

func TestHandler_AdminCreate_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.createGBResult = nil
	env.mock.createGBErr = errors.New("店铺不存在")
	body := dto.CreateGroupBuyRequest{Title: "x", GroupBuyPrice: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin", body)

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

// ---------- AdminUpdate ----------

func TestHandler_AdminUpdate_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateGroupBuyRequest{Title: "新版套餐", Stock: 100}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateGBID)
	require.NotNil(t, env.mock.lastUpdateGBReq)
	assert.Equal(t, "新版套餐", env.mock.lastUpdateGBReq.Title)
	assert.Equal(t, 100, env.mock.lastUpdateGBReq.Stock)
}

func TestHandler_AdminUpdate_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/abc", dto.UpdateGroupBuyRequest{Title: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的团购ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateGBID)
}

func TestHandler_AdminUpdate_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/groupbuy/admin/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateGBID)
}

func TestHandler_AdminUpdate_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateGBErr = errors.New("团购商品不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5", dto.UpdateGroupBuyRequest{Title: "x"})

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "团购商品不存在", resp.Message)
}

// ---------- AdminDelete ----------

func TestHandler_AdminDelete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteGBID)
}

func TestHandler_AdminDelete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的团购ID", resp.Message)
}

func TestHandler_AdminDelete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.deleteGBErr = errors.New("团购商品不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/7", nil)

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "团购商品不存在", resp.Message)
}

// ---------- AdminUpdateStatus ----------

func TestHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateGroupBuyStatusRequest{Status: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5/status", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "操作成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateGBStatusID)
	assert.Equal(t, 1, env.mock.lastUpdateGBStatusStatus)
}

func TestHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/abc/status", dto.UpdateGroupBuyStatusRequest{Status: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的团购ID", resp.Message)
}

func TestHandler_AdminUpdateStatus_BindError_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=5 不满足 oneof=0 1
	resp := env.doRaw(t, http.MethodPut, "/api/v1/groupbuy/admin/5/status", `{"status":5}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, 0, env.mock.lastUpdateGBStatusStatus)
}

func TestHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateGBStatErr = errors.New("团购商品不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5/status", dto.UpdateGroupBuyStatusRequest{Status: 1})

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "团购商品不存在", resp.Message)
}

// ---------- AdminAudit ----------

func TestHandler_AdminAudit_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.AuditGroupBuyRequest{AuditStatus: 1}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5/audit", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastAuditGBID)
	assert.Equal(t, 1, env.mock.lastAuditGBAuditStatus)
}

func TestHandler_AdminAudit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/abc/audit", dto.AuditGroupBuyRequest{AuditStatus: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的团购ID", resp.Message)
}

func TestHandler_AdminAudit_BindError_InvalidAuditStatus(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// audit_status=0 不满足 oneof=1 2
	resp := env.doRaw(t, http.MethodPut, "/api/v1/groupbuy/admin/5/audit", `{"audit_status":0}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_AdminAudit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.auditGBErr = errors.New("团购商品不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/5/audit", dto.AuditGroupBuyRequest{AuditStatus: 1})

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "团购商品不存在", resp.Message)
}

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastListGBRegionID)
	require.NotNil(t, env.mock.lastListGBReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.GroupBuyInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestHandler_AdminList_StatusOutOfRange_DefaultAll(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=99 越界（>2）→ 管理端兜底为 -1（全部）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin?status=99", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListGBReq)
	assert.Equal(t, -1, env.mock.lastListGBReq.Status)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.listGBResult = nil
	env.mock.listGBErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin", nil)

	assert.Equal(t, 2601, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- AdminOrderList ----------

func TestHandler_AdminOrderList_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/orders?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastAdminOrderListRegionID)
	require.NotNil(t, env.mock.lastAdminOrderListReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.OrderInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_AdminOrderList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.adminOrderListResult = nil
	env.mock.adminOrderListErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/orders", nil)

	// CodeOrderError=2609
	assert.Equal(t, 2609, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== 管理接口 - 优惠券 ====================

// ---------- AdminCreateCoupon ----------

func TestHandler_AdminCreateCoupon_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.CreateCouponRequest{Name: "满100减20", Type: 1, Value: 20, MinAmount: 100, Scope: 0, ValidityType: 0, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, uint(2), env.mock.lastCreateCouponRegionID)
	require.NotNil(t, env.mock.lastCreateCouponReq)
	assert.Equal(t, "满100减20", env.mock.lastCreateCouponReq.Name)
	assert.Equal(t, 1, env.mock.lastCreateCouponReq.Type)
	var info dto.CouponInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestHandler_AdminCreateCoupon_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", "{not json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateCouponReq)
}

func TestHandler_AdminCreateCoupon_BindError_MissingName(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// 缺少 name（required）+ type 不满足 oneof=1 2 3
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", `{"type":1,"scope":0,"validity_type":0,"status":1}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, env.mock.lastCreateCouponReq)
}

func TestHandler_AdminCreateCoupon_BindError_InvalidType(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// type=9 不满足 oneof=1 2 3
	resp := env.doRaw(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", `{"name":"x","type":9,"scope":0,"validity_type":0,"status":1}`, "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
}

func TestHandler_AdminCreateCoupon_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.createCouponResult = nil
	env.mock.createCouponErr = errors.New("优惠券创建失败")
	body := dto.CreateCouponRequest{Name: "x", Type: 1, Scope: 0, ValidityType: 0, Status: 1}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", body)

	// CodeCouponError=2605
	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "优惠券创建失败", resp.Message)
}

// ---------- AdminUpdateCoupon ----------

func TestHandler_AdminUpdateCoupon_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	body := dto.UpdateCouponRequest{Name: "满200减50", Value: 50}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/coupons/5", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateCouponID)
	require.NotNil(t, env.mock.lastUpdateCouponReq)
	assert.Equal(t, "满200减50", env.mock.lastUpdateCouponReq.Name)
}

func TestHandler_AdminUpdateCoupon_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/coupons/abc", dto.UpdateCouponRequest{Name: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的优惠券ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateCouponID)
}

func TestHandler_AdminUpdateCoupon_BindError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/groupbuy/admin/coupons/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastUpdateCouponID)
}

func TestHandler_AdminUpdateCoupon_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.updateCouponErr = errors.New("优惠券不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/groupbuy/admin/coupons/5", dto.UpdateCouponRequest{Name: "x"})

	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "优惠券不存在", resp.Message)
}

// ---------- AdminDeleteCoupon ----------

func TestHandler_AdminDeleteCoupon_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/coupons/7", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastDeleteCouponID)
}

func TestHandler_AdminDeleteCoupon_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/coupons/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的优惠券ID", resp.Message)
}

func TestHandler_AdminDeleteCoupon_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.deleteCouponErr = errors.New("优惠券不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/groupbuy/admin/coupons/7", nil)

	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "优惠券不存在", resp.Message)
}

// ---------- AdminListCoupons ----------

func TestHandler_AdminListCoupons_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/coupons?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastListCouponRegionID)
	require.NotNil(t, env.mock.lastListCouponReq)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.CouponInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestHandler_AdminListCoupons_StatusOutOfRange_DefaultAll(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	// status=99 越界（>1）→ 管理端兜底为 -1（全部）
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/coupons?status=99", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastListCouponReq)
	assert.Equal(t, -1, env.mock.lastListCouponReq.Status)
}

func TestHandler_AdminListCoupons_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 2)
	env.mock.listCouponResult = nil
	env.mock.listCouponErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/coupons", nil)

	assert.Equal(t, 2605, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== regionID 注入聚合 ====================

func TestHandler_RegionIDInjection_Aggregate(t *testing.T) {
	// 验证所有接收 regionID 的接口均透传 context 中的 regionID
	env := newHandlerEnv(t, 1, 9)

	// 团购商品读/写
	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/list", nil)
	assert.Equal(t, uint(9), env.mock.lastListGBRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin", dto.CreateGroupBuyRequest{Title: "x", GroupBuyPrice: 1})
	assert.Equal(t, uint(9), env.mock.lastCreateGBRegionID)

	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin", nil)
	assert.Equal(t, uint(9), env.mock.lastListGBRegionID)

	// 优惠券读/写
	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/coupons", nil)
	assert.Equal(t, uint(9), env.mock.lastAvailableCouponsRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/admin/coupons", dto.CreateCouponRequest{Name: "x", Type: 1, Scope: 0, ValidityType: 0, Status: 1})
	assert.Equal(t, uint(9), env.mock.lastCreateCouponRegionID)

	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/coupons", nil)
	assert.Equal(t, uint(9), env.mock.lastListCouponRegionID)

	// 订单读/写
	env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/orders", dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
	assert.Equal(t, uint(9), env.mock.lastCreateOrderRegionID)

	env.doJSON(t, http.MethodPost, "/api/v1/groupbuy/coupons/1/receive", nil)
	assert.Equal(t, uint(9), env.mock.lastReceiveCouponRegionID)

	env.doJSON(t, http.MethodGet, "/api/v1/groupbuy/admin/orders", nil)
	assert.Equal(t, uint(9), env.mock.lastAdminOrderListRegionID)
}
