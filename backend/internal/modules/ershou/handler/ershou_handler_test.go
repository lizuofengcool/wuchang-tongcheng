// Package handler_test 二手物品模块主表 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 Handler 全部分支：
//   - 公开接口无需登录（List/GetByID/Nearby/Search/ListMessages/FavStatus 在 userID=0 时不被 401 拦截）
//   - 用户接口未登录拦截（Create/Update/Delete/ListMine/Fav/ListFavs/CreateMessage → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/gte → 400 "参数错误"）
//   - Nearby 经纬度越界（lat<-90||lat>90||lng<-180||lng>180 → 400 "经纬度参数无效"）
//   - Search 关键词为空（required 校验失败 → 400 "参数错误"）
//   - Fav toggle 语义（HasFaved=true→"收藏成功"，HasFaved=false→"已取消收藏"）
//   - service 成功/错误透传（业务码 CodeErshouError=2701/CodeErshouNotFound=2702/CodeErshouPublishError=2703/CodeErshouAuditError=2704 + message + data 透传）
//   - 地区ID/用户信息上下文注入（regionID/userID/username/phone/avatar 透传给 service）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 groupbuy/house/category/region/news/file/setting/permission handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/ershou/dto"
	ershouHandler "wuchang-tongcheng/internal/modules/ershou/handler"
	"wuchang-tongcheng/internal/modules/ershou/service"
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

// mockErshouService 内存 mock，实现 service.ErshouService 接口。
// 记录最近一次调用入参，返回预设 result/err，便于断言 handler 装配层逻辑。
type mockErshouService struct {
	// ===== 调用记录 =====
	// C 端
	lastCreateRegionID  uint
	lastCreateUserID    uint
	lastCreateUserName  string
	lastCreateUserPhone string
	lastCreateUserAvatar string
	lastCreateReq       *dto.CreateErshouRequest

	lastUpdateID         uint
	lastUpdateOperatorID uint
	lastUpdateReq        *dto.UpdateErshouRequest

	lastDeleteID         uint
	lastDeleteOperatorID uint

	lastGetByIDID     uint
	lastGetByIDUserID uint

	lastListRegionID uint
	lastListReq      *dto.ErshouListRequest

	lastListNearbyRegionID uint
	lastListNearbyReq      *dto.ErshouNearbyRequest

	lastSearchRegionID uint
	lastSearchReq      *dto.ErshouSearchRequest

	lastListMineUserID   uint
	lastListMinePage     int
	lastListMinePageSize int

	// 收藏
	lastFavUserID   uint
	lastFavErshouID uint

	lastFavStatusUserID   uint
	lastFavStatusErshouID uint

	lastListFavsUserID   uint
	lastListFavsPage     int
	lastListFavsPageSize int

	// 留言
	lastCreateMsgErshouID   uint
	lastCreateMsgFromUserID uint
	lastCreateMsgFromName   string
	lastCreateMsgFromAvatar string
	lastCreateMsgReq        *dto.CreateMessageRequest

	lastListMsgErshouID   uint
	lastListMsgPage       int
	lastListMsgPageSize   int

	// M 端管理
	lastAdminListReq *dto.ErshouAdminListRequest

	lastAdminGetByIDID uint

	lastAuditID          uint
	lastAuditAuditStatus int
	lastAuditAuditReason string

	lastAdminUpdateStatusID     uint
	lastAdminUpdateStatusStatus int

	// ===== 返回值预设 =====
	createResult *dto.ErshouInfo
	createErr    error
	updateErr    error
	deleteErr    error
	getByIDResult *dto.ErshouInfo
	getByIDErr    error
	listResult    []dto.ErshouInfo
	listErr       error
	listNearbyResult []dto.ErshouInfo
	listNearbyErr    error
	searchResult   []dto.ErshouInfo
	searchErr      error
	listMineResult []dto.ErshouInfo
	listMineErr    error

	favResult *dto.FavResponse
	favErr    error
	favStatusResult *dto.FavResponse
	favStatusErr    error
	listFavsResult []dto.ErshouInfo
	listFavsErr    error

	createMsgResult *dto.MessageInfo
	createMsgErr    error
	listMsgResult   []dto.MessageInfo
	listMsgTotal    int64
	listMsgErr      error

	adminListResult []dto.ErshouInfo
	adminListErr    error
	adminGetByIDResult *dto.ErshouInfo
	adminGetByIDErr    error
	auditErr           error
	adminUpdateStatusErr error
}

// ===== C 端 =====

func (m *mockErshouService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateErshouRequest) (*dto.ErshouInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateUserName = userName
	m.lastCreateUserPhone = userPhone
	m.lastCreateUserAvatar = userAvatar
	m.lastCreateReq = req
	return m.createResult, m.createErr
}

func (m *mockErshouService) Update(id uint, operatorID uint, req *dto.UpdateErshouRequest) error {
	m.lastUpdateID = id
	m.lastUpdateOperatorID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockErshouService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteOperatorID = operatorID
	return m.deleteErr
}

func (m *mockErshouService) GetByID(id uint, userID uint) (*dto.ErshouInfo, error) {
	m.lastGetByIDID = id
	m.lastGetByIDUserID = userID
	return m.getByIDResult, m.getByIDErr
}

func (m *mockErshouService) List(regionID uint, req *dto.ErshouListRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.listResult))
	return pagination, m.listResult, nil
}

func (m *mockErshouService) ListNearby(regionID uint, req *dto.ErshouNearbyRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastListNearbyRegionID = regionID
	m.lastListNearbyReq = req
	if m.listNearbyErr != nil {
		return nil, nil, m.listNearbyErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.listNearbyResult))
	return pagination, m.listNearbyResult, nil
}

func (m *mockErshouService) Search(regionID uint, req *dto.ErshouSearchRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.searchErr != nil {
		return nil, nil, m.searchErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.searchResult))
	return pagination, m.searchResult, nil
}

func (m *mockErshouService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastListMineUserID = userID
	m.lastListMinePage = page
	m.lastListMinePageSize = pageSize
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = int64(len(m.listMineResult))
	return pagination, m.listMineResult, nil
}

// ===== 收藏 =====

func (m *mockErshouService) Fav(userID, ershouID uint) (*dto.FavResponse, error) {
	m.lastFavUserID = userID
	m.lastFavErshouID = ershouID
	return m.favResult, m.favErr
}

func (m *mockErshouService) FavStatus(userID, ershouID uint) (*dto.FavResponse, error) {
	m.lastFavStatusUserID = userID
	m.lastFavStatusErshouID = ershouID
	return m.favStatusResult, m.favStatusErr
}

func (m *mockErshouService) ListFavs(userID uint, page, pageSize int) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastListFavsUserID = userID
	m.lastListFavsPage = page
	m.lastListFavsPageSize = pageSize
	if m.listFavsErr != nil {
		return nil, nil, m.listFavsErr
	}
	pagination := utils.NewPagination(page, pageSize)
	pagination.Total = int64(len(m.listFavsResult))
	return pagination, m.listFavsResult, nil
}

// ===== 留言 =====

func (m *mockErshouService) CreateMessage(ershouID uint, fromUserID uint, fromName string, fromAvatar string, req *dto.CreateMessageRequest) (*dto.MessageInfo, error) {
	m.lastCreateMsgErshouID = ershouID
	m.lastCreateMsgFromUserID = fromUserID
	m.lastCreateMsgFromName = fromName
	m.lastCreateMsgFromAvatar = fromAvatar
	m.lastCreateMsgReq = req
	return m.createMsgResult, m.createMsgErr
}

func (m *mockErshouService) ListMessages(ershouID uint, page, pageSize int) ([]dto.MessageInfo, int64, error) {
	m.lastListMsgErshouID = ershouID
	m.lastListMsgPage = page
	m.lastListMsgPageSize = pageSize
	if m.listMsgErr != nil {
		return nil, 0, m.listMsgErr
	}
	return m.listMsgResult, m.listMsgTotal, nil
}

// ===== M 端管理 =====

func (m *mockErshouService) AdminList(req *dto.ErshouAdminListRequest) (*utils.Pagination, []dto.ErshouInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	pagination.Total = int64(len(m.adminListResult))
	return pagination, m.adminListResult, nil
}

func (m *mockErshouService) AdminGetByID(id uint) (*dto.ErshouInfo, error) {
	m.lastAdminGetByIDID = id
	return m.adminGetByIDResult, m.adminGetByIDErr
}

func (m *mockErshouService) Audit(id uint, auditStatus int, auditReason string) error {
	m.lastAuditID = id
	m.lastAuditAuditStatus = auditStatus
	m.lastAuditAuditReason = auditReason
	return m.auditErr
}

func (m *mockErshouService) AdminUpdateStatus(id uint, status int) error {
	m.lastAdminUpdateStatusID = id
	m.lastAdminUpdateStatusStatus = status
	return m.adminUpdateStatusErr
}

// 确保 mockErshouService 实现 service.ErshouService 接口
var _ service.ErshouService = (*mockErshouService)(nil)

// handlerEnv handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	mock   *mockErshouService
}

// newHandlerEnv 构造 gin 引擎并注册 ershou 主表路由（路径与 ershou/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 登录态时注入 username/phone/avatar，用于 Create/CreateMessage 冗余字段透传断言。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mock := &mockErshouService{
		createResult: &dto.ErshouInfo{ID: 1, Title: "iPhone 13 二手", Price: 2999, Condition: "used", PriceUnit: "元", UserID: ctxUserID, RegionID: regionID, Status: 1, AuditStatus: 1},
		getByIDResult: &dto.ErshouInfo{ID: 1, Title: "iPhone 13 二手", Price: 2999, Condition: "used", Status: 1, AuditStatus: 1},
		listResult:    []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1}, {ID: 2, Title: "小米电视", Price: 800, Status: 1}},
		listNearbyResult: []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Distance: 1.2, Status: 1}},
		searchResult:   []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1}},
		listMineResult: []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1}},
		favResult:      &dto.FavResponse{HasFaved: true, FavCount: 5},
		favStatusResult: &dto.FavResponse{HasFaved: false, FavCount: 4},
		listFavsResult: []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1}},
		createMsgResult: &dto.MessageInfo{ID: 1, ErshouID: 1, FromUserID: ctxUserID, FromName: "testuser", Content: "还在吗？"},
		listMsgResult:   []dto.MessageInfo{{ID: 1, ErshouID: 1, FromUserID: ctxUserID, Content: "还在吗？"}},
		listMsgTotal:    1,
		adminListResult: []dto.ErshouInfo{{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1, AuditStatus: 0}},
		adminGetByIDResult: &dto.ErshouInfo{ID: 1, Title: "iPhone 13 二手", Price: 2999, Status: 1, AuditStatus: 0},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id 及 profile
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		if ctxUserID != 0 {
			c.Set(middleware.ContextUsername, "testuser")
			c.Set(middleware.ContextUserPhone, "13800000000")
			c.Set(middleware.ContextUserAvatar, "https://cdn.example.com/a.png")
		}
		c.Next()
	})

	h := ershouHandler.NewHandler(mock)
	// 注册路由，路径与 ershou/plugin.go RegisterRoutes 保持一致（去掉权限中间件，纯测 handler）
	root := r.Group("/api/v1/ershou")
	// 公开接口（固定路径须在 :id 之前注册）
	root.GET("", h.List)
	root.GET("/search", h.Search)
	root.GET("/nearby", h.Nearby)
	root.GET("/:id", h.GetByID)
	root.GET("/:id/messages", h.ListMessages)
	root.GET("/:id/fav", h.FavStatus)
	// 用户接口
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
	root.GET("/mine", h.ListMine)
	root.GET("/favorites", h.ListFavs)
	root.POST("/:id/fav", h.Fav)
	root.POST("/:id/messages", h.CreateMessage)
	// 管理接口
	admin := root.Group("/admin")
	admin.GET("/list", h.AdminList)
	admin.GET("/:id", h.AdminGetByID)
	admin.PUT("/:id/audit", h.Audit)
	admin.PUT("/:id/status", h.AdminUpdateStatus)

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

// ==================== C 端：Create ====================

func TestHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou", dto.CreateErshouRequest{
		Title:      "iPhone 13 二手",
		Price:      2999,
		CategoryID: 10,
		Status:     1,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "发布成功", resp.Message)
	// 上下文注入断言
	assert.Equal(t, uint(3), env.mock.lastCreateRegionID)
	assert.Equal(t, uint(7), env.mock.lastCreateUserID)
	assert.Equal(t, "testuser", env.mock.lastCreateUserName)
	assert.Equal(t, "13800000000", env.mock.lastCreateUserPhone)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateUserAvatar)
	require.NotNil(t, env.mock.lastCreateReq)
	assert.Equal(t, "iPhone 13 二手", env.mock.lastCreateReq.Title)
	assert.Equal(t, 2999.0, env.mock.lastCreateReq.Price)
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3) // 未登录
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou", dto.CreateErshouRequest{Title: "x", Status: 1})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Nil(t, env.mock.lastCreateReq)
}

func TestHandler_Create_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ershou", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_BindError_Validation(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	// Title 缺失（required）
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou", map[string]interface{}{"price": 100, "status": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.createResult = nil
	env.mock.createErr = errors.New("发布失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou", dto.CreateErshouRequest{Title: "x", Status: 1})

	// CodeErshouPublishError=2703
	assert.Equal(t, 2703, resp.Code)
	assert.Equal(t, "发布失败", resp.Message)
}

// ==================== C 端：Update ====================

func TestHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/5", dto.UpdateErshouRequest{Title: "降价了", Price: 2500})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastUpdateID)
	assert.Equal(t, uint(7), env.mock.lastUpdateOperatorID)
	require.NotNil(t, env.mock.lastUpdateReq)
	assert.Equal(t, "降价了", env.mock.lastUpdateReq.Title)
}

func TestHandler_Update_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/5", dto.UpdateErshouRequest{Title: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/abc", dto.UpdateErshouRequest{Title: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Update_BindError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/ershou/5", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.updateErr = errors.New("无权操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/5", dto.UpdateErshouRequest{Title: "x"})

	// CodeErshouError=2701
	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "无权操作", resp.Message)
}

// ==================== C 端：Delete ====================

func TestHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/ershou/5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastDeleteID)
	assert.Equal(t, uint(7), env.mock.lastDeleteOperatorID)
}

func TestHandler_Delete_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/ershou/5", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/ershou/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.deleteErr = errors.New("无权操作此二手物品")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/ershou/5", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "无权操作此二手物品", resp.Message)
}

// ==================== C 端：GetByID ====================

func TestHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 3) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/9", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.Equal(t, uint(9), env.mock.lastGetByIDID)
	assert.Equal(t, uint(0), env.mock.lastGetByIDUserID) // 未登录 userID=0
}

func TestHandler_GetByID_LoginUser(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.doJSON(t, http.MethodGet, "/api/v1/ershou/9", nil)

	assert.Equal(t, uint(7), env.mock.lastGetByIDUserID)
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	env.mock.getByIDResult = nil
	env.mock.getByIDErr = errors.New("二手物品不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/9", nil)

	// CodeErshouNotFound=2702
	assert.Equal(t, 2702, resp.Code)
	assert.Equal(t, "二手物品不存在", resp.Message)
}

// ==================== C 端：List ====================

func TestHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou?page=1&page_size=10&keyword=手机", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastListRegionID)
	require.NotNil(t, env.mock.lastListReq)
	assert.Equal(t, "手机", env.mock.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.ErshouInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "iPhone 13 二手", list[0].Title)
}

func TestHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.listResult = nil
	env.mock.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 5, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/ershou", nil)

	assert.Equal(t, uint(8), env.mock.lastListRegionID)
}

// ==================== C 端：Nearby ====================

func TestHandler_Nearby_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby?latitude=30.5&longitude=114.3&radius_km=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(2), env.mock.lastListNearbyRegionID)
	require.NotNil(t, env.mock.lastListNearbyReq)
	assert.Equal(t, 30.5, env.mock.lastListNearbyReq.Latitude)
	assert.Equal(t, 114.3, env.mock.lastListNearbyReq.Longitude)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Nearby_BindError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	// 缺少 required 的 latitude/longitude → bind 失败
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Nearby_LatOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby?latitude=91&longitude=114", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestHandler_Nearby_LngOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby?latitude=30&longitude=181", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestHandler_Nearby_LatNegativeOutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby?latitude=-91&longitude=114", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "经纬度参数无效", resp.Message)
}

func TestHandler_Nearby_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 2)
	env.mock.listNearbyResult = nil
	env.mock.listNearbyErr = errors.New("postgis down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/nearby?latitude=30&longitude=114", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "postgis down", resp.Message)
}

// ==================== C 端：Search ====================

func TestHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/search?keyword=手机&page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastSearchRegionID)
	require.NotNil(t, env.mock.lastSearchReq)
	assert.Equal(t, "手机", env.mock.lastSearchReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_Search_EmptyKeyword(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// keyword 为空串：ErshouSearchRequest.Keyword 带 binding:"required"，
	// 空串触发 required 校验失败 → Bind 失败 → 400 "参数错误: ..."
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/search?keyword=", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Search_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.mock.searchResult = nil
	env.mock.searchErr = errors.New("es boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/search?keyword=手机", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "es boom", resp.Message)
}

// ==================== C 端：ListMine ====================

func TestHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/mine?page=2&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListMineUserID)
	assert.Equal(t, 2, env.mock.lastListMinePage)
	assert.Equal(t, 5, env.mock.lastListMinePageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListMine_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.doJSON(t, http.MethodGet, "/api/v1/ershou/mine", nil)

	// 不传分页参数 → 默认 page=1, page_size=10
	assert.Equal(t, 1, env.mock.lastListMinePage)
	assert.Equal(t, 10, env.mock.lastListMinePageSize)
}

func TestHandler_ListMine_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/mine", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.listMineResult = nil
	env.mock.listMineErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/mine", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== 收藏：Fav ====================

func TestHandler_Fav_Success_Faved(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.favResult = &dto.FavResponse{HasFaved: true, FavCount: 6}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "收藏成功", resp.Message)
	assert.Equal(t, uint(7), env.mock.lastFavUserID)
	assert.Equal(t, uint(5), env.mock.lastFavErshouID)
}

func TestHandler_Fav_Success_Unfaved(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.favResult = &dto.FavResponse{HasFaved: false, FavCount: 4}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已取消收藏", resp.Message)
}

func TestHandler_Fav_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/fav", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_Fav_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Fav_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.favResult = nil
	env.mock.favErr = errors.New("已删除")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/fav", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "已删除", resp.Message)
}

// ==================== 收藏：FavStatus ====================

func TestHandler_FavStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 3) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/5/fav", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(0), env.mock.lastFavStatusUserID)
	assert.Equal(t, uint(5), env.mock.lastFavStatusErshouID)
}

func TestHandler_FavStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/abc/fav", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_FavStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	env.mock.favStatusResult = nil
	env.mock.favStatusErr = errors.New("二手物品不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/5/fav", nil)

	// CodeErshouNotFound=2702
	assert.Equal(t, 2702, resp.Code)
	assert.Equal(t, "二手物品不存在", resp.Message)
}

// ==================== 收藏：ListFavs ====================

func TestHandler_ListFavs_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/favorites?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.mock.lastListFavsUserID)
	assert.Equal(t, 1, env.mock.lastListFavsPage)
	assert.Equal(t, 10, env.mock.lastListFavsPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_ListFavs_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/favorites", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_ListFavs_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.listFavsResult = nil
	env.mock.listFavsErr = errors.New("redis down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/favorites", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "redis down", resp.Message)
}

// ==================== 留言：CreateMessage ====================

func TestHandler_CreateMessage_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/messages", dto.CreateMessageRequest{Content: "还在吗？"})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "留言成功", resp.Message)
	// 上下文注入断言（ershouID/fromUserID/fromName/fromAvatar）
	assert.Equal(t, uint(5), env.mock.lastCreateMsgErshouID)
	assert.Equal(t, uint(7), env.mock.lastCreateMsgFromUserID)
	assert.Equal(t, "testuser", env.mock.lastCreateMsgFromName)
	assert.Equal(t, "https://cdn.example.com/a.png", env.mock.lastCreateMsgFromAvatar)
	require.NotNil(t, env.mock.lastCreateMsgReq)
	assert.Equal(t, "还在吗？", env.mock.lastCreateMsgReq.Content)
}

func TestHandler_CreateMessage_Unauthorized(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/messages", dto.CreateMessageRequest{Content: "x"})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestHandler_CreateMessage_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/abc/messages", dto.CreateMessageRequest{Content: "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_CreateMessage_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/ershou/5/messages", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_CreateMessage_BindError_Validation(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	// Content 缺失（required）
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/messages", map[string]interface{}{})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_CreateMessage_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.createMsgResult = nil
	env.mock.createMsgErr = errors.New("物品不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/ershou/5/messages", dto.CreateMessageRequest{Content: "x"})

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "物品不存在", resp.Message)
}

// ==================== 留言：ListMessages ====================

func TestHandler_ListMessages_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 3) // 公开接口未登录也可访问
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/5/messages?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.mock.lastListMsgErshouID)
	assert.Equal(t, 1, env.mock.lastListMsgPage)
	assert.Equal(t, 10, env.mock.lastListMsgPageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.MessageInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "还在吗？", list[0].Content)
}

func TestHandler_ListMessages_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/abc/messages", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_ListMessages_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 3)
	env.mock.listMsgResult = nil
	env.mock.listMsgErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/5/messages", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== M 端：AdminList ====================

func TestHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/admin/list?page=1&page_size=20&keyword=手机", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.mock.lastAdminListReq)
	assert.Equal(t, "手机", env.mock.lastAdminListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestHandler_AdminList_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.adminListResult = nil
	env.mock.adminListErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/admin/list", nil)

	assert.Equal(t, 2701, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ==================== M 端：AdminGetByID ====================

func TestHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/admin/9", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(9), env.mock.lastAdminGetByIDID)
}

func TestHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/admin/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.adminGetByIDResult = nil
	env.mock.adminGetByIDErr = errors.New("二手物品不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/ershou/admin/9", nil)

	// CodeErshouNotFound=2702
	assert.Equal(t, 2702, resp.Code)
	assert.Equal(t, "二手物品不存在", resp.Message)
}

// ==================== M 端：Audit ====================

func TestHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/audit", dto.AuditRequest{AuditStatus: 1, AuditReason: "内容合规"})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastAuditID)
	assert.Equal(t, 1, env.mock.lastAuditAuditStatus)
	assert.Equal(t, "内容合规", env.mock.lastAuditAuditReason)
}

func TestHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/abc/audit", dto.AuditRequest{AuditStatus: 1})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_Audit_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/ershou/admin/5/audit", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_BindError_Validation(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	// audit_status=5 不在 oneof=0 1 2 → 校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/audit", map[string]interface{}{"audit_status": 5})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.auditErr = errors.New("已审核的物品不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/audit", dto.AuditRequest{AuditStatus: 1})

	// CodeErshouAuditError=2704
	assert.Equal(t, 2704, resp.Code)
	assert.Equal(t, "已审核的物品不能重复审核", resp.Message)
}

// ==================== M 端：AdminUpdateStatus ====================

func TestHandler_AdminUpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/status", dto.AdminUpdateStatusRequest{Status: 3})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(5), env.mock.lastAdminUpdateStatusID)
	assert.Equal(t, 3, env.mock.lastAdminUpdateStatusStatus)
}

func TestHandler_AdminUpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/abc/status", dto.AdminUpdateStatusRequest{Status: 3})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestHandler_AdminUpdateStatus_BindError_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/ershou/admin/5/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_BindError_Validation(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	// status=99 不在 oneof=1 3 4 → 校验失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/status", map[string]interface{}{"status": 99})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestHandler_AdminUpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 3)
	env.mock.adminUpdateStatusErr = errors.New("状态非法")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/ershou/admin/5/status", dto.AdminUpdateStatusRequest{Status: 3})

	// CodeErshouAuditError=2704
	assert.Equal(t, 2704, resp.Code)
	assert.Equal(t, "状态非法", resp.Message)
}
