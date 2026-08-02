// Package handler_test 同城商城订单主表 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 mall OrderHandler 全部分支：
//   - 公开只读接口无需登录（GetByID/GetByOrderNo/ListByShop）
//   - 需登录接口未登录拦截（Create/Cancel/Ship/Confirm/ListByUser/CountByStatus → 401 "请先登录"）
//   - 无登录校验的管理接口（Complete/Delete/AdminList/AdminClose/BatchUpdateStatus/AutoClose/AutoConfirm/AutoReview）
//   - URL :id / :shop_id 参数解析失败（非数字 → 400 "无效的ID" / "无效的店铺ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/min/oneof/max → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeMallOrderError=5411 / CodeMallOrderNotFound=5412）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - Create 客户端 IP 提取（X-Forwarded-For 首段 / X-Real-IP）与 User-Agent 透传
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 dh114/house/job/marketing/shop/groupbuy handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/mall/dto"
	mallHandler "wuchang-tongcheng/internal/modules/mall/handler"
	"wuchang-tongcheng/internal/modules/mall/service"
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

// mockOrderService 内存 mock，实现 service.OrderService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockOrderService struct {
	// ===== 调用记录 =====
	// C 端创建
	lastCreateRegionID    uint
	lastCreateUserID      uint
	lastCreateBuyerName   string
	lastCreateBuyerPhone  string
	lastCreateBuyerAvatar string
	lastCreateIP          string
	lastCreateUserAgent  string
	lastCreateReq         *dto.CreateOrderRequest

	lastCancelID     uint
	lastCancelUserID uint
	lastCancelReq     *dto.CancelOrderRequest

	lastShipID        uint
	lastShipShipperID uint
	lastShipReq       *dto.ShipOrderRequest

	lastConfirmID     uint
	lastConfirmUserID  uint
	lastCompleteID     uint
	lastDeleteID       uint

	// 只读
	lastGetByIDID       uint
	lastGetByOrderNo    string
	lastListByShopID    uint
	lastListByShopReq   *dto.OrderListRequest
	lastListByUserReq   *dto.OrderListRequest
	lastListByUserID    uint
	lastAdminListReq    *dto.AdminOrderListRequest
	lastCountByStatusID uint
	lastCountByStatus   int
	lastSummaryReq      *dto.OrderSummary
	lastSummaryRegionID uint
	lastSummaryUserID   uint
	lastSummaryShopID   uint
	lastSummaryStart    string
	lastSummaryEnd      string

	// 批量/定时
	lastBatchIDs    []uint
	lastBatchStatus int

	// ===== 返回值预设 =====
	createResult *dto.OrderInfo
	createErr    error

	cancelErr error

	shipErr error

	confirmErr error

	completeErr error
	deleteErr   error

	getByIDResult *dto.OrderInfo
	getByIDErr    error

	getByOrderNoResult *dto.OrderInfo
	getByOrderNoErr    error

	listByUserResult []dto.OrderInfo
	listByUserErr    error
	listByUserTotal  int64

	listByShopResult []dto.OrderInfo
	listByShopErr    error
	listByShopTotal  int64

	adminListResult []dto.OrderInfo
	adminListErr    error
	adminListTotal  int64

	countByStatusResult int64
	countByStatusErr    error

	summaryResult *dto.OrderSummary
	summaryErr    error

	batchErr error

	autoCloseResult   int
	autoCloseErr      error
	autoConfirmResult int
	autoConfirmErr    error
	autoReviewResult  int
	autoReviewErr     error
}

// ===== Create =====

func (m *mockOrderService) Create(regionID, userID uint, buyerName, buyerPhone, buyerAvatar, clientIP, userAgent string, req *dto.CreateOrderRequest) (*dto.OrderInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateBuyerName = buyerName
	m.lastCreateBuyerPhone = buyerPhone
	m.lastCreateBuyerAvatar = buyerAvatar
	m.lastCreateIP = clientIP
	m.lastCreateUserAgent = userAgent
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

// ===== Cancel =====

func (m *mockOrderService) Cancel(id, userID uint, req *dto.CancelOrderRequest) error {
	m.lastCancelID = id
	m.lastCancelUserID = userID
	m.lastCancelReq = req
	return m.cancelErr
}

// ===== AdminClose =====

func (m *mockOrderService) AdminClose(id uint, req *dto.AdminCloseOrderRequest) error {
	m.lastCancelID = id
	m.lastCancelReq = &dto.CancelOrderRequest{Reason: req.AdminRemark}
	return m.cancelErr
}

// ===== Ship =====

func (m *mockOrderService) Ship(id, shipperID uint, req *dto.ShipOrderRequest) error {
	m.lastShipID = id
	m.lastShipShipperID = shipperID
	m.lastShipReq = req
	return m.shipErr
}

// ===== Confirm =====

func (m *mockOrderService) Confirm(id, userID uint) error {
	m.lastConfirmID = id
	m.lastConfirmUserID = userID
	return m.confirmErr
}

// ===== Complete =====

func (m *mockOrderService) Complete(id uint) error {
	m.lastCompleteID = id
	return m.completeErr
}

// ===== Delete =====

func (m *mockOrderService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

// ===== GetByID =====

func (m *mockOrderService) GetByID(id uint) (*dto.OrderInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

// ===== GetByOrderNo =====

func (m *mockOrderService) GetByOrderNo(orderNo string) (*dto.OrderInfo, error) {
	m.lastGetByOrderNo = orderNo
	if m.getByOrderNoErr != nil {
		return nil, m.getByOrderNoErr
	}
	return m.getByOrderNoResult, nil
}

// ===== ListByUser =====

func (m *mockOrderService) ListByUser(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	m.lastListByUserID = userID
	m.lastListByUserReq = req
	if m.listByUserErr != nil {
		return nil, nil, m.listByUserErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listByUserTotal}, m.listByUserResult, nil
}

// ===== ListByShop =====

func (m *mockOrderService) ListByShop(shopID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	m.lastListByShopID = shopID
	m.lastListByShopReq = req
	if m.listByShopErr != nil {
		return nil, nil, m.listByShopErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listByShopTotal}, m.listByShopResult, nil
}

// ===== AdminList =====

func (m *mockOrderService) AdminList(req *dto.AdminOrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.adminListTotal}, m.adminListResult, nil
}

// ===== BatchUpdateStatus =====

func (m *mockOrderService) BatchUpdateStatus(ids []uint, status int) error {
	m.lastBatchIDs = ids
	m.lastBatchStatus = status
	return m.batchErr
}

// ===== CountByStatus =====

func (m *mockOrderService) CountByStatus(userID uint, status int) (int64, error) {
	m.lastCountByStatusID = userID
	m.lastCountByStatus = status
	if m.countByStatusErr != nil {
		return 0, m.countByStatusErr
	}
	return m.countByStatusResult, nil
}

// ===== Summary =====

func (m *mockOrderService) Summary(regionID, userID, shopID uint, startDate, endDate string) (*dto.OrderSummary, error) {
	m.lastSummaryRegionID = regionID
	m.lastSummaryUserID = userID
	m.lastSummaryShopID = shopID
	m.lastSummaryStart = startDate
	m.lastSummaryEnd = endDate
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	return m.summaryResult, nil
}

// ===== 定时任务 =====

func (m *mockOrderService) AutoClose() (int, error) {
	if m.autoCloseErr != nil {
		return 0, m.autoCloseErr
	}
	return m.autoCloseResult, nil
}

func (m *mockOrderService) AutoConfirm() (int, error) {
	if m.autoConfirmErr != nil {
		return 0, m.autoConfirmErr
	}
	return m.autoConfirmResult, nil
}

func (m *mockOrderService) AutoReview() (int, error) {
	if m.autoReviewErr != nil {
		return 0, m.autoReviewErr
	}
	return m.autoReviewResult, nil
}

// 编译期接口实现校验
var _ service.OrderService = (*mockOrderService)(nil)

// handlerEnv 测试环境：gin 引擎 + mock service。
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockOrderService
}

// newHandlerEnv 构造 gin 引擎并注册 mall 订单主表路由（路径与 mall/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 同时注入 username/phone/avatar 冗余字段，用于 Create 透传断言。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockOrderService{
		createResult: &dto.OrderInfo{
			ID:        1,
			OrderNo:   "WC20260801001",
			UserID:    ctxUserID,
			ShopID:    2,
			ShopName:  "武昌特产店",
			PayAmount: 99.50,
			Status:    1,
		},
		getByIDResult: &dto.OrderInfo{
			ID: 1, OrderNo: "WC20260801001", ShopID: 2, ShopName: "武昌特产店", PayAmount: 99.50, Status: 1,
		},
		getByOrderNoResult: &dto.OrderInfo{
			ID: 1, OrderNo: "WC20260801001", Status: 1,
		},
		listByUserResult: []dto.OrderInfo{
			{ID: 1, OrderNo: "WC20260801001", Status: 1},
			{ID: 2, OrderNo: "WC20260801002", Status: 2},
		},
		listByUserTotal: 2,
		listByShopResult: []dto.OrderInfo{
			{ID: 1, OrderNo: "WC20260801001", ShopID: 5, Status: 1},
		},
		listByShopTotal: 1,
		adminListResult: []dto.OrderInfo{
			{ID: 1, OrderNo: "WC20260801001", Status: 1},
			{ID: 2, OrderNo: "WC20260801002", Status: 2},
			{ID: 3, OrderNo: "WC20260801003", Status: 3},
		},
		adminListTotal:      3,
		countByStatusResult: 5,
		summaryResult: &dto.OrderSummary{
			TotalCount: 10, TotalAmount: 999.00, PaidCount: 5, PaidAmount: 499.50,
		},
		autoCloseResult:   3,
		autoConfirmResult: 2,
		autoReviewResult:  1,
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

	h := mallHandler.NewOrderHandler(mock)
	root := r.Group("/api/v1/mall")
	// 公开只读接口
	root.GET("/orders/by-shop/:shop_id", h.ListByShop)
	root.GET("/orders/by-no/:order_no", h.GetByOrderNo)
	root.GET("/orders/:id", h.GetByID)
	// 需登录接口（C 端）
	root.POST("/orders", h.Create)
	root.GET("/orders/mine", h.ListByUser)
	root.GET("/orders/count-by-status", h.CountByStatus)
	root.GET("/orders/summary", h.Summary)
	root.PUT("/orders/:id/cancel", h.Cancel)
	root.PUT("/orders/:id/ship", h.Ship)
	root.PUT("/orders/:id/confirm", h.Confirm)
	root.PUT("/orders/:id/complete", h.Complete)
	root.DELETE("/orders/:id", h.Delete)
	// 管理后台接口（/admin 组，去掉权限中间件）
	admin := root.Group("/admin")
	admin.GET("/orders", h.AdminList)
	admin.PUT("/orders/:id/close", h.AdminClose)
	admin.PUT("/orders/batch-status", h.BatchUpdateStatus)
	admin.POST("/orders/auto-close", h.AutoClose)
	admin.POST("/orders/auto-confirm", h.AutoConfirm)
	admin.POST("/orders/auto-review", h.AutoReview)

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

// validCreateBody 构造一份可通过 Bind 校验的完整 CreateOrderRequest 请求体。
func validCreateBody() map[string]interface{} {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{"product_id": 10, "sku_id": 100, "quantity": 2},
		},
		"address_id":     1,
		"remark":         "尽快发货",
		"payment_method": "wechat",
		"from_cart":      true,
	}
}

// ==================== 公开只读接口（无需登录） ====================

// ---------- GetByID ----------

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.mock.lastGetByIDID)
	var info dto.OrderInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "WC20260801001", info.OrderNo)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastGetByIDID)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("订单不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/999", nil)

	// CodeMallOrderNotFound=5412
	assert.Equal(t, utils.CodeMallOrderNotFound, resp.Code)
	assert.Equal(t, "订单不存在", resp.Message)
}

// ---------- GetByOrderNo ----------

func TestHandler_GetByOrderNo_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-no/WC20260801001", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "WC20260801001", env.mock.lastGetByOrderNo)
	var info dto.OrderInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "WC20260801001", info.OrderNo)
}

func TestHandler_GetByOrderNo_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.getByOrderNoResult = nil
	env.mock.getByOrderNoErr = errors.New("订单号无效")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-no/NOTEXIST", nil)

	assert.Equal(t, utils.CodeMallOrderNotFound, resp.Code)
	assert.Equal(t, "订单号无效", resp.Message)
}

// ---------- ListByShop ----------

func TestHandler_ListByShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-shop/7?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListByShopID)
	require.NotNil(t, env.mock.lastListByShopReq)
	assert.Equal(t, 1, env.mock.lastListByShopReq.Page)
	assert.Equal(t, 10, env.mock.lastListByShopReq.PageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.OrderInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, uint(5), list[0].ShopID)
}

func TestHandler_ListByShop_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-shop/7", nil)

	require.NotNil(t, env.mock.lastListByShopReq)
	// 未传 page/page_size → parsePagination 兜底 1/10
	assert.Equal(t, 1, env.mock.lastListByShopReq.Page)
	assert.Equal(t, 10, env.mock.lastListByShopReq.PageSize)
}

func TestHandler_ListByShop_InvalidShopID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-shop/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListByShopID)
}

func TestHandler_ListByShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.listByShopResult = nil
	env.mock.listByShopErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/by-shop/7", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== 需登录接口（C 端） ====================

// ---------- Create ----------

func TestHandler_Create_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody())

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCreateUserID)
}

func TestHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/mall/orders", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// items / address_id 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/orders", map[string]interface{}{
		"remark": "无商品无地址",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindFail_EmptyItems(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// items 为 required,min=1，空数组 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/orders", map[string]interface{}{
		"items":      []map[string]interface{}{},
		"address_id": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "下单成功", resp.Message)
	// 上下文透传
	assert.Equal(t, uint(5), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "张三", env.mock.lastCreateBuyerName)
	assert.Equal(t, "13800000000", env.mock.lastCreateBuyerPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateBuyerAvatar)
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, uint(1), env.mock.lastCreateReq.AddressID)
	require.Len(t, env.mock.lastCreateReq.Items, 1)
	assert.Equal(t, uint(10), env.mock.lastCreateReq.Items[0].ProductID)
	assert.Equal(t, 2, env.mock.lastCreateReq.Items[0].Quantity)
	assert.Equal(t, "wechat", env.mock.lastCreateReq.PaymentMethod)
	var info dto.OrderInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "WC20260801001", info.OrderNo)
}

func TestHandler_Create_IPForwardedFor(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// X-Forwarded-For 含多 IP，取首段
	env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody(), map[string]string{
		"X-Forwarded-For": "203.0.113.5, 70.41.122.8",
		"User-Agent":      "Mozilla/5.0 (iPhone; iOS)",
	})

	assert.Equal(t, "203.0.113.5", env.mock.lastCreateIP)
	assert.Equal(t, "Mozilla/5.0 (iPhone; iOS)", env.mock.lastCreateUserAgent)
}

func TestHandler_Create_IPRealIP(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 无 X-Forwarded-For，回退 X-Real-IP
	env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody(), map[string]string{
		"X-Real-IP": "198.51.100.7",
	})

	assert.Equal(t, "198.51.100.7", env.mock.lastCreateIP)
}

func TestHandler_Create_IPForwardedForSingle(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// X-Forwarded-For 单 IP（无逗号）
	env.doJSONWithHeaders(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody(), map[string]string{
		"X-Forwarded-For": "203.0.113.99",
	})

	assert.Equal(t, "203.0.113.99", env.mock.lastCreateIP)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("商品库存不足")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/orders", validCreateBody())

	// CodeMallOrderError=5411
	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "商品库存不足", resp.Message)
}

// ---------- Cancel ----------

func TestHandler_Cancel_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/cancel", map[string]interface{}{
		"reason": "不想买了",
	})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCancelUserID)
}

func TestHandler_Cancel_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/abc/cancel", map[string]interface{}{
		"reason": "不想买了",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Cancel_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/mall/orders/3/cancel", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Cancel_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/cancel", map[string]interface{}{
		"reason": "不想买了",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "取消成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastCancelID)
	assert.Equal(t, uint(7), env.mock.lastCancelUserID)
	require.NotNil(t, env.mock.lastCancelReq)
	assert.Equal(t, "不想买了", env.mock.lastCancelReq.Reason)
}

func TestHandler_Cancel_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.cancelErr = errors.New("订单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/cancel", map[string]interface{}{
		"reason": "不想买了",
	})

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "订单状态不允许此操作", resp.Message)
}

// ---------- Ship ----------

func TestHandler_Ship_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/ship", map[string]interface{}{
		"logistics_company": "顺丰", "logistics_no": "SF123",
	})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastShipShipperID)
}

func TestHandler_Ship_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/abc/ship", map[string]interface{}{
		"logistics_company": "顺丰", "logistics_no": "SF123",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Ship_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// logistics_company/logistics_no 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/ship", map[string]interface{}{
		"seller_remark": "缺物流信息",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Ship_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/mall/orders/3/ship", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Ship_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/ship", map[string]interface{}{
		"logistics_company": "顺丰速运", "logistics_no": "SF123456", "seller_remark": "已发出",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发货成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastShipID)
	// shipperID 透传登录 userID
	assert.Equal(t, uint(7), env.mock.lastShipShipperID)
	require.NotNil(t, env.mock.lastShipReq)
	assert.Equal(t, "顺丰速运", env.mock.lastShipReq.LogisticsCompany)
	assert.Equal(t, "SF123456", env.mock.lastShipReq.LogisticsNo)
	assert.Equal(t, "已发出", env.mock.lastShipReq.SellerRemark)
}

func TestHandler_Ship_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.shipErr = errors.New("无权操作他人订单")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/ship", map[string]interface{}{
		"logistics_company": "顺丰", "logistics_no": "SF123",
	})

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "无权操作他人订单", resp.Message)
}

// ---------- Confirm ----------

func TestHandler_Confirm_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/confirm", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastConfirmUserID)
}

func TestHandler_Confirm_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/abc/confirm", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Confirm_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/confirm", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "确认收货成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastConfirmID)
	assert.Equal(t, uint(7), env.mock.lastConfirmUserID)
}

func TestHandler_Confirm_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.confirmErr = errors.New("订单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/confirm", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "订单状态不允许此操作", resp.Message)
}

// ---------- Complete ----------

func TestHandler_Complete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/abc/complete", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCompleteID)
}

func TestHandler_Complete_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/complete", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "订单已完成", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastCompleteID)
}

func TestHandler_Complete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.completeErr = errors.New("订单不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/orders/3/complete", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "订单不存在", resp.Message)
}

// ---------- Delete ----------

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/mall/orders/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastDeleteID)
}

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/mall/orders/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastDeleteID)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.deleteErr = errors.New("订单不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/mall/orders/3", nil)

	// Delete 使用 CodeMallOrderNotFound=5412
	assert.Equal(t, utils.CodeMallOrderNotFound, resp.Code)
	assert.Equal(t, "订单不存在", resp.Message)
}

// ---------- ListByUser ----------

func TestHandler_ListByUser_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastListByUserID)
}

func TestHandler_ListByUser_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/mine?page=2&page_size=5&status=2", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListByUserID)
	require.NotNil(t, env.mock.lastListByUserReq)
	assert.Equal(t, 2, env.mock.lastListByUserReq.Page)
	assert.Equal(t, 5, env.mock.lastListByUserReq.PageSize)
	require.NotNil(t, env.mock.lastListByUserReq.Status)
	assert.Equal(t, 2, *env.mock.lastListByUserReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	assert.Equal(t, 2, p.Page)
	assert.Equal(t, 5, p.PageSize)
	var list []dto.OrderInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestHandler_ListByUser_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/mine", nil)

	require.NotNil(t, env.mock.lastListByUserReq)
	// 未传 page/page_size → parsePagination 兜底 1/10
	assert.Equal(t, 1, env.mock.lastListByUserReq.Page)
	assert.Equal(t, 10, env.mock.lastListByUserReq.PageSize)
}

func TestHandler_ListByUser_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.listByUserResult = nil
	env.mock.listByUserErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/mine", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- CountByStatus ----------

func TestHandler_CountByStatus_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 5) // 未登录
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/count-by-status?status=1", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.mock.lastCountByStatusID)
}

func TestHandler_CountByStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/count-by-status?status=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastCountByStatusID)
	assert.Equal(t, 1, env.mock.lastCountByStatus)
	// data 结构：{count: int64}
	var data struct {
		Count int64 `json:"count"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, int64(5), data.Count)
}

func TestHandler_CountByStatus_DefaultStatus(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 未传 status → parseQueryInt 兜底 -1
	env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/count-by-status", nil)

	assert.Equal(t, -1, env.mock.lastCountByStatus)
}

func TestHandler_CountByStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.countByStatusErr = errors.New("stat error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/count-by-status?status=1", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "stat error", resp.Message)
}

// ---------- Summary ----------

func TestHandler_Summary_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/summary?shop_id=9&start_date=2026-08-01&end_date=2026-08-31", nil)

	assert.Equal(t, 0, resp.Code)
	// regionID 注入
	assert.Equal(t, uint(5), env.mock.lastSummaryRegionID)
	assert.Equal(t, uint(7), env.mock.lastSummaryUserID)
	assert.Equal(t, uint(9), env.mock.lastSummaryShopID)
	assert.Equal(t, "2026-08-01", env.mock.lastSummaryStart)
	assert.Equal(t, "2026-08-31", env.mock.lastSummaryEnd)
	var s dto.OrderSummary
	require.NoError(t, json.Unmarshal(resp.Data, &s))
	assert.Equal(t, int64(10), s.TotalCount)
	assert.Equal(t, 999.00, s.TotalAmount)
}

func TestHandler_Summary_DefaultShopID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	// 未传 shop_id → parseQueryInt 兜底 0
	env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/summary", nil)

	assert.Equal(t, uint(0), env.mock.lastSummaryShopID)
}

func TestHandler_Summary_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.mock.summaryErr = errors.New("summary error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/orders/summary", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "summary error", resp.Message)
}

// ==================== 管理后台接口（/admin 组） ====================

// ---------- AdminList ----------

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 8)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/admin/orders?page=1&page_size=10&keyword=特产", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, 1, env.mock.lastAdminListReq.Page)
	assert.Equal(t, 10, env.mock.lastAdminListReq.PageSize)
	assert.Equal(t, "特产", env.mock.lastAdminListReq.Keyword)
	// 未传 region_id → 兜底注入 regionID
	assert.Equal(t, uint(8), env.mock.lastAdminListReq.RegionID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(3), p.Total)
	var list []dto.OrderInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 3)
}

func TestHandler_AdminList_RegionIDFromQuery(t *testing.T) {
	env := newHandlerEnv(t, 0, 8)
	// 显式传 region_id → 不被覆盖
	env.doJSON(t, http.MethodGet, "/api/v1/mall/admin/orders?region_id=99", nil)

	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, uint(99), env.mock.lastAdminListReq.RegionID)
}

func TestHandler_AdminList_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 0, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/mall/admin/orders", nil)

	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, 1, env.mock.lastAdminListReq.Page)
	assert.Equal(t, 10, env.mock.lastAdminListReq.PageSize)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 8)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/mall/admin/orders", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- AdminClose ----------

func TestHandler_AdminClose_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/abc/close", map[string]interface{}{
		"admin_remark": "违规",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminClose_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/mall/admin/orders/3/close", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminClose_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/3/close", map[string]interface{}{
		"admin_remark": "违规商品",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "订单已关闭", resp.Message)
	assert.Equal(t, uint(3), env.mock.lastCancelID)
}

func TestHandler_AdminClose_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.cancelErr = errors.New("订单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/3/close", map[string]interface{}{
		"admin_remark": "违规商品",
	})

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "订单状态不允许此操作", resp.Message)
}

// ---------- BatchUpdateStatus ----------

func TestHandler_BatchUpdateStatus_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/mall/admin/orders/batch-status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchUpdateStatus_BindFail_EmptyIDs(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// ids 为 required,min=1，空数组 → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/batch-status", map[string]interface{}{
		"ids":    []uint{},
		"status": 5,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_BatchUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/batch-status", map[string]interface{}{
		"ids":    []uint{1, 2, 3},
		"status": 5,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "批量更新成功", resp.Message)
	assert.Equal(t, []uint{1, 2, 3}, env.mock.lastBatchIDs)
	assert.Equal(t, 5, env.mock.lastBatchStatus)
}

func TestHandler_BatchUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.batchErr = errors.New("部分订单状态不允许")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/mall/admin/orders/batch-status", map[string]interface{}{
		"ids": []uint{1, 2}, "status": 5,
	})

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "部分订单状态不允许", resp.Message)
}

// ---------- AutoClose ----------

func TestHandler_AutoClose_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-close", nil)

	assert.Equal(t, 0, resp.Code)
	var data struct {
		ClosedCount int `json:"closed_count"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, 3, data.ClosedCount)
}

func TestHandler_AutoClose_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.autoCloseErr = errors.New("cron error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-close", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "cron error", resp.Message)
}

// ---------- AutoConfirm ----------

func TestHandler_AutoConfirm_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-confirm", nil)

	assert.Equal(t, 0, resp.Code)
	var data struct {
		ConfirmedCount int `json:"confirmed_count"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, 2, data.ConfirmedCount)
}

func TestHandler_AutoConfirm_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.autoConfirmErr = errors.New("cron error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-confirm", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "cron error", resp.Message)
}

// ---------- AutoReview ----------

func TestHandler_AutoReview_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-review", nil)

	assert.Equal(t, 0, resp.Code)
	var data struct {
		ReviewedCount int `json:"reviewed_count"`
	}
	require.NoError(t, json.Unmarshal(resp.Data, &data))
	assert.Equal(t, 1, data.ReviewedCount)
}

func TestHandler_AutoReview_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.autoReviewErr = errors.New("cron error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/mall/admin/orders/auto-review", nil)

	assert.Equal(t, utils.CodeMallOrderError, resp.Code)
	assert.Equal(t, "cron error", resp.Message)
}
