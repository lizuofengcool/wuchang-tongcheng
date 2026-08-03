// Package handler_test 多租户分站中台 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 tenant 4 个 Handler 全部分支：
//   - StationHandler：GetCurrent（公开，domain 必填）/ List（分页兜底）/ GetByID / Create / Update / Delete /
//     UpdateStatus（启停）/ CopyConfig（配置复制）
//   - StaffHandler：List / GetByID / Create / Update / Delete / ListByStation
//   - ConfigHandler：List / GetByID / Upsert / Update / Delete / ListByStationAndModule（query 必填+parseUint）/ BatchGet
//   - DomainHandler：List / GetByID / Create / Update / Delete / SetPrimary / UpdateSSL / ListByStation
//
// 覆盖维度：
//   - URL :id / :station_id 参数解析失败（非数字 → 400 "无效的ID" / "无效的分站ID"）
//   - ListByStationAndModule 缺参（→ 400 "station_id 和 biz_module 参数不能为空"）/ 非数字 station_id（→ 400 "无效的 station_id"）
//   - GetCurrent 空 domain（→ 400 "domain 参数不能为空"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/max → 400 "参数错误"）
//   - service 成功/错误透传（业务码 CodeTenantError=5601/CodeTenantNotFound=5602/CodeTenantStatusInvalid=5605/
//     CodeTenantStaffError=5606/CodeTenantStaffNotFound=5607/CodeTenantConfigError=5609/CodeTenantConfigNotFound=5610/
//     CodeTenantDomainError=5611/CodeTenantDomainNotFound=5612/CodeTenantCopyError=5613/CodeTenantSSLInvalid=5614）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）与未传分页兜底 1/10
//   - 成功响应消息断言（"分站创建成功"/"更新成功"/"删除成功"/"状态已更新"/"配置复制完成"/
//     "员工添加成功"/"配置保存成功"/"域名绑定成功"/"已设为主域名"/"SSL 状态已更新"）
//
// 注：tenant handler 不在 handler 内做登录校验/上下文注入（鉴权与地区注入由 AuthRequired/Region 中间件完成，
// 与 mall/dh114 在 handler 内显式校验不同），故本测试聚焦参数解析/Bind/service 透传，纯内存 mock 不依赖 DB/Redis/Docker。
// 与 mall/dh114/house/job handler 测试同风格。
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

	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/tenant/dto"
	tenantHandler "wuchang-tongcheng/internal/modules/tenant/handler"
	"wuchang-tongcheng/internal/modules/tenant/service"
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

// ============================================================
// mockStationService
// ============================================================

type mockStationService struct {
	lastCreateReq     *dto.CreateStationRequest
	lastUpdateID      uint
	lastUpdateReq     *dto.UpdateStationRequest
	lastDeleteID      uint
	lastGetByIDID     uint
	lastGetByDomain   string
	lastListReq       *dto.StationListRequest
	lastUpdateStatID  uint
	lastUpdateStat    int
	lastCopyConfigReq *dto.CopyConfigRequest

	createResult     *dto.StationInfo
	createErr        error
	updateErr        error
	deleteErr        error
	getByIDResult    *dto.StationInfo
	getByIDErr       error
	getByDomainRes   *dto.StationInfo
	getByDomainErr   error
	listResult       []dto.StationInfo
	listErr          error
	listTotal        int64
	updateStatusErr  error
	copyConfigResult *dto.CopyConfigResult
	copyConfigErr    error
}

func (m *mockStationService) Create(req *dto.CreateStationRequest) (*dto.StationInfo, error) {
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}
func (m *mockStationService) Update(id uint, req *dto.UpdateStationRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockStationService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockStationService) GetByID(id uint) (*dto.StationInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}
func (m *mockStationService) GetByRegionID(regionID uint) (*dto.StationInfo, error) {
	return nil, nil // handler 未使用
}
func (m *mockStationService) GetByDomain(domain string) (*dto.StationInfo, error) {
	m.lastGetByDomain = domain
	if m.getByDomainErr != nil {
		return nil, m.getByDomainErr
	}
	return m.getByDomainRes, nil
}
func (m *mockStationService) List(req *dto.StationListRequest) (*utils.Pagination, []dto.StationInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}
func (m *mockStationService) UpdateStatus(id uint, status int) error {
	m.lastUpdateStatID = id
	m.lastUpdateStat = status
	return m.updateStatusErr
}
func (m *mockStationService) CopyConfig(req *dto.CopyConfigRequest) (*dto.CopyConfigResult, error) {
	m.lastCopyConfigReq = req
	if m.copyConfigErr != nil {
		return nil, m.copyConfigErr
	}
	return m.copyConfigResult, nil
}

var _ service.StationService = (*mockStationService)(nil)

// ============================================================
// mockStaffService
// ============================================================

type mockStaffService struct {
	lastCreateReq        *dto.CreateStaffRequest
	lastUpdateID         uint
	lastUpdateReq        *dto.UpdateStaffRequest
	lastDeleteID         uint
	lastGetByIDID        uint
	lastListReq          *dto.StaffListRequest
	lastListByStationID  uint

	createResult    *dto.StaffInfo
	createErr       error
	updateErr       error
	deleteErr       error
	getByIDResult   *dto.StaffInfo
	getByIDErr      error
	listResult      []dto.StaffInfo
	listErr         error
	listTotal       int64
	listByStationR  []dto.StaffInfo
	listByStationErr error
}

func (m *mockStaffService) Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error) {
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}
func (m *mockStaffService) Update(id uint, req *dto.UpdateStaffRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockStaffService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockStaffService) GetByID(id uint) (*dto.StaffInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}
func (m *mockStaffService) List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}
func (m *mockStaffService) ListByStation(stationID uint) ([]dto.StaffInfo, error) {
	m.lastListByStationID = stationID
	if m.listByStationErr != nil {
		return nil, m.listByStationErr
	}
	return m.listByStationR, nil
}
func (m *mockStaffService) ListByUser(userID uint) ([]dto.StaffInfo, error) {
	return nil, nil // handler 未使用
}

var _ service.StaffService = (*mockStaffService)(nil)

// ============================================================
// mockConfigService
// ============================================================

type mockConfigService struct {
	lastUpsertReq              *dto.UpsertConfigRequest
	lastUpdateID               uint
	lastUpdateReq              *dto.UpdateConfigRequest
	lastDeleteID               uint
	lastGetByIDID              uint
	lastListReq                *dto.ConfigListRequest
	lastListBySMStationID      uint
	lastListBySMModule         string
	lastBatchGetReq            *dto.BatchGetConfigRequest

	upsertResult      *dto.ConfigInfo
	upsertErr         error
	updateErr         error
	deleteErr         error
	getByIDResult     *dto.ConfigInfo
	getByIDErr        error
	listResult        []dto.ConfigInfo
	listErr           error
	listTotal         int64
	listBySMResult    []dto.ConfigInfo
	listBySMErr       error
	batchGetResult    []dto.ConfigKeyValue
	batchGetErr       error
}

func (m *mockConfigService) Upsert(req *dto.UpsertConfigRequest) (*dto.ConfigInfo, error) {
	m.lastUpsertReq = req
	if m.upsertErr != nil {
		return nil, m.upsertErr
	}
	return m.upsertResult, nil
}
func (m *mockConfigService) Update(id uint, req *dto.UpdateConfigRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockConfigService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockConfigService) GetByID(id uint) (*dto.ConfigInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}
func (m *mockConfigService) List(req *dto.ConfigListRequest) (*utils.Pagination, []dto.ConfigInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}
func (m *mockConfigService) ListByStationAndModule(stationID uint, bizModule string) ([]dto.ConfigInfo, error) {
	m.lastListBySMStationID = stationID
	m.lastListBySMModule = bizModule
	if m.listBySMErr != nil {
		return nil, m.listBySMErr
	}
	return m.listBySMResult, nil
}
func (m *mockConfigService) BatchGet(req *dto.BatchGetConfigRequest) ([]dto.ConfigKeyValue, error) {
	m.lastBatchGetReq = req
	if m.batchGetErr != nil {
		return nil, m.batchGetErr
	}
	return m.batchGetResult, nil
}

var _ service.ConfigService = (*mockConfigService)(nil)

// ============================================================
// mockDomainService
// ============================================================

type mockDomainService struct {
	lastCreateReq       *dto.CreateDomainRequest
	lastUpdateID        uint
	lastUpdateReq       *dto.UpdateDomainRequest
	lastDeleteID        uint
	lastGetByIDID       uint
	lastListReq         *dto.DomainListRequest
	lastListByStationID uint
	lastSetPrimaryID    uint
	lastUpdateSSLID     uint
	lastUpdateSSLStatus string

	createResult     *dto.DomainInfo
	createErr        error
	updateErr        error
	deleteErr        error
	getByIDResult    *dto.DomainInfo
	getByIDErr       error
	listResult       []dto.DomainInfo
	listErr          error
	listTotal        int64
	listByStationRes []dto.DomainInfo
	listByStationErr error
	setPrimaryErr    error
	updateSSLErr     error
}

func (m *mockDomainService) Create(req *dto.CreateDomainRequest) (*dto.DomainInfo, error) {
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}
func (m *mockDomainService) Update(id uint, req *dto.UpdateDomainRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}
func (m *mockDomainService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}
func (m *mockDomainService) GetByID(id uint) (*dto.DomainInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}
func (m *mockDomainService) List(req *dto.DomainListRequest) (*utils.Pagination, []dto.DomainInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	return &utils.Pagination{Page: req.Page, PageSize: req.PageSize, Total: m.listTotal}, m.listResult, nil
}
func (m *mockDomainService) ListByStation(stationID uint) ([]dto.DomainInfo, error) {
	m.lastListByStationID = stationID
	if m.listByStationErr != nil {
		return nil, m.listByStationErr
	}
	return m.listByStationRes, nil
}
func (m *mockDomainService) SetPrimary(id uint) error {
	m.lastSetPrimaryID = id
	return m.setPrimaryErr
}
func (m *mockDomainService) UpdateSSLStatus(id uint, status string) error {
	m.lastUpdateSSLID = id
	m.lastUpdateSSLStatus = status
	return m.updateSSLErr
}

var _ service.DomainService = (*mockDomainService)(nil)

// ============================================================
// 测试环境
// ============================================================

// handlerEnv 测试环境：gin 引擎 + 4 个 mock service。
type handlerEnv struct {
	engine  *gin.Engine
	station *mockStationService
	staff   *mockStaffService
	config  *mockConfigService
	domain  *mockDomainService
}

// newHandlerEnv 构造 gin 引擎并注册 tenant 全部路由（路径与 tenant/plugin.go RegisterRoutes 一致）。
// 路由注册去掉鉴权/限流/权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	stationMock := &mockStationService{
		createResult: &dto.StationInfo{
			ID: 1, RegionID: 5, Name: "武昌分站", Domain: "wuchang.example.com",
			Status: 1, StatusText: "已启用",
		},
		getByIDResult: &dto.StationInfo{
			ID: 1, RegionID: 5, Name: "武昌分站", Domain: "wuchang.example.com", Status: 1,
		},
		getByDomainRes: &dto.StationInfo{
			ID: 1, RegionID: 5, Name: "武昌分站", Domain: "wuchang.example.com", Status: 1,
		},
		listResult: []dto.StationInfo{
			{ID: 1, RegionID: 5, Name: "武昌分站", Status: 1},
			{ID: 2, RegionID: 6, Name: "洪山分站", Status: 1},
		},
		listTotal: 2,
		copyConfigResult: &dto.CopyConfigResult{
			SourceStationID: 1, TargetStationID: 2, BizModule: "mall", CopiedCount: 3,
		},
	}

	staffMock := &mockStaffService{
		createResult: &dto.StaffInfo{
			ID: 10, StationID: 1, UserID: 100, Role: "operator", RoleText: "运营员", Status: 1,
		},
		getByIDResult: &dto.StaffInfo{
			ID: 10, StationID: 1, UserID: 100, Role: "manager", RoleText: "管理员", Status: 1,
		},
		listResult: []dto.StaffInfo{
			{ID: 10, StationID: 1, UserID: 100, Role: "operator", Status: 1},
		},
		listTotal: 1,
		listByStationR: []dto.StaffInfo{
			{ID: 10, StationID: 1, UserID: 100, Role: "manager", Status: 1},
			{ID: 11, StationID: 1, UserID: 101, Role: "operator", Status: 1},
		},
	}

	configMock := &mockConfigService{
		upsertResult: &dto.ConfigInfo{
			ID: 1, StationID: 1, BizModule: "mall", ConfigKey: "site_name", ConfigValue: "武昌商城",
		},
		getByIDResult: &dto.ConfigInfo{
			ID: 1, StationID: 1, BizModule: "mall", ConfigKey: "site_name", ConfigValue: "武昌商城",
		},
		listResult: []dto.ConfigInfo{
			{ID: 1, StationID: 1, BizModule: "mall", ConfigKey: "site_name", ConfigValue: "武昌商城"},
		},
		listTotal: 1,
		listBySMResult: []dto.ConfigInfo{
			{ID: 1, StationID: 1, BizModule: "mall", ConfigKey: "site_name", ConfigValue: "武昌商城"},
			{ID: 2, StationID: 1, BizModule: "mall", ConfigKey: "logo", ConfigValue: "https://cdn.example.com/logo.png"},
		},
		batchGetResult: []dto.ConfigKeyValue{
			{ConfigKey: "site_name", ConfigValue: "武昌商城"},
		},
	}

	domainMock := &mockDomainService{
		createResult: &dto.DomainInfo{
			ID: 1, StationID: 1, Domain: "wuchang.example.com", IsPrimary: true, SSLStatus: "active", SSLText: "已生效",
		},
		getByIDResult: &dto.DomainInfo{
			ID: 1, StationID: 1, Domain: "wuchang.example.com", IsPrimary: true, SSLStatus: "active",
		},
		listResult: []dto.DomainInfo{
			{ID: 1, StationID: 1, Domain: "wuchang.example.com", IsPrimary: true, SSLStatus: "active"},
		},
		listTotal: 1,
		listByStationRes: []dto.DomainInfo{
			{ID: 1, StationID: 1, Domain: "wuchang.example.com", IsPrimary: true, SSLStatus: "active"},
			{ID: 2, StationID: 1, Domain: "m.wuchang.example.com", IsPrimary: false, SSLStatus: "none"},
		},
	}

	r := coreRouter.NewRouter()
	sh := tenantHandler.NewStationHandler(stationMock)
	sth := tenantHandler.NewStaffHandler(staffMock)
	ch := tenantHandler.NewConfigHandler(configMock)
	dh := tenantHandler.NewDomainHandler(domainMock)

	root := r.Group("/api/v1/tenant")
	// 公开
	root.GET("/stations/current", sh.GetCurrent)
	// 需登录
	root.GET("/stations", sh.List)

	admin := root.Group("/admin")
	// 分站管理
	admin.GET("/stations", sh.List)
	admin.GET("/stations/:id", sh.GetByID)
	admin.POST("/stations", sh.Create)
	admin.PUT("/stations/:id", sh.Update)
	admin.DELETE("/stations/:id", sh.Delete)
	admin.PUT("/stations/:id/status", sh.UpdateStatus)
	admin.POST("/stations/copy-config", sh.CopyConfig)
	// 员工管理
	admin.GET("/staff", sth.List)
	admin.GET("/staff/:id", sth.GetByID)
	admin.POST("/staff", sth.Create)
	admin.PUT("/staff/:id", sth.Update)
	admin.DELETE("/staff/:id", sth.Delete)
	admin.GET("/staff/by-station/:station_id", sth.ListByStation)
	// 配置管理
	admin.GET("/configs", ch.List)
	admin.GET("/configs/:id", ch.GetByID)
	admin.POST("/configs", ch.Upsert)
	admin.PUT("/configs/:id", ch.Update)
	admin.DELETE("/configs/:id", ch.Delete)
	admin.GET("/configs/by-station-module", ch.ListByStationAndModule)
	admin.POST("/configs/batch-get", ch.BatchGet)
	// 域名管理
	admin.GET("/domains", dh.List)
	admin.GET("/domains/:id", dh.GetByID)
	admin.POST("/domains", dh.Create)
	admin.PUT("/domains/:id", dh.Update)
	admin.DELETE("/domains/:id", dh.Delete)
	admin.PUT("/domains/:id/primary", dh.SetPrimary)
	admin.PUT("/domains/:id/ssl", dh.UpdateSSL)
	admin.GET("/domains/by-station/:station_id", dh.ListByStation)

	return &handlerEnv{
		engine:  r.Engine(),
		station: stationMock,
		staff:   staffMock,
		config:  configMock,
		domain:  domainMock,
	}
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

// ==================== 合法请求体构造器 ====================

func validCreateStationBody() map[string]interface{} {
	return map[string]interface{}{
		"region_id": 5, "name": "武昌分站", "domain": "wuchang.example.com", "status": 1,
	}
}
func validUpdateStationBody() map[string]interface{} {
	return map[string]interface{}{"name": "武昌新分站", "status": 1}
}
func validUpdateStationStatusBody() map[string]interface{} {
	return map[string]interface{}{"status": 1}
}
func validCopyConfigBody() map[string]interface{} {
	return map[string]interface{}{
		"source_station_id": 1, "target_station_id": 2, "biz_module": "mall",
	}
}
func validCreateStaffBody() map[string]interface{} {
	return map[string]interface{}{
		"station_id": 1, "user_id": 100, "role": "operator", "status": 1,
	}
}
func validUpdateStaffBody() map[string]interface{} {
	return map[string]interface{}{"role": "manager", "status": 1}
}
func validUpsertConfigBody() map[string]interface{} {
	return map[string]interface{}{
		"station_id": 1, "biz_module": "mall", "config_key": "site_name", "config_value": "武昌商城",
	}
}
func validUpdateConfigBody() map[string]interface{} {
	return map[string]interface{}{"config_value": "新名称"}
}
func validBatchGetConfigBody() map[string]interface{} {
	return map[string]interface{}{
		"station_id": 1, "biz_module": "mall", "config_keys": []string{"site_name", "logo"},
	}
}
func validCreateDomainBody() map[string]interface{} {
	return map[string]interface{}{
		"station_id": 1, "domain": "wuchang.example.com", "is_primary": true, "ssl_status": "active",
	}
}
func validUpdateDomainBody() map[string]interface{} {
	return map[string]interface{}{"ssl_status": "active"}
}
func validUpdateSSLBody() map[string]interface{} {
	return map[string]interface{}{"ssl_status": "active"}
}

// ==================== StationHandler ====================

// ---------- GetCurrent（公开） ----------

func TestStationHandler_GetCurrent_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations/current?domain=wuchang.example.com", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "wuchang.example.com", env.station.lastGetByDomain)
	var info dto.StationInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "武昌分站", info.Name)
}

func TestStationHandler_GetCurrent_EmptyDomain(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations/current", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "domain 参数不能为空", resp.Message)
	assert.Equal(t, "", env.station.lastGetByDomain)
}

func TestStationHandler_GetCurrent_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.getByDomainRes = nil
	env.station.getByDomainErr = errors.New("当前域名未匹配到分站")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations/current?domain=unknown.com", nil)

	assert.Equal(t, tenantHandler.CodeTenantNotFound, resp.Code)
	assert.Equal(t, "当前域名未匹配到分站", resp.Message)
}

// ---------- List ----------

func TestStationHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations?page=2&page_size=5&keyword=武昌", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.station.lastListReq)
	assert.Equal(t, 2, env.station.lastListReq.Page)
	assert.Equal(t, 5, env.station.lastListReq.PageSize)
	assert.Equal(t, "武昌", env.station.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	assert.Equal(t, 2, p.Page)
	assert.Equal(t, 5, p.PageSize)
	var list []dto.StationInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestStationHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t)
	env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations", nil)

	require.NotNil(t, env.station.lastListReq)
	// 未传 page/page_size → parsePagination 兜底 1/10
	assert.Equal(t, 1, env.station.lastListReq.Page)
	assert.Equal(t, 10, env.station.lastListReq.PageSize)
}

func TestStationHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.listResult = nil
	env.station.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/stations", nil)

	assert.Equal(t, tenantHandler.CodeTenantError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetByID（admin） ----------

func TestStationHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/stations/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.station.lastGetByIDID)
	var info dto.StationInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestStationHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/stations/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.station.lastGetByIDID)
}

func TestStationHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.getByIDResult = nil
	env.station.getByIDErr = errors.New("分站不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/stations/999", nil)

	assert.Equal(t, tenantHandler.CodeTenantNotFound, resp.Code)
	assert.Equal(t, "分站不存在", resp.Message)
}

// ---------- Create（admin） ----------

func TestStationHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/stations", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// region_id / name 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations", map[string]interface{}{
		"domain": "x.example.com",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_Create_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// status oneof=0 1，传 9 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations", map[string]interface{}{
		"region_id": 5, "name": "武昌分站", "status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations", validCreateStationBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "分站创建成功", resp.Message)
	require.NotNil(t, env.station.lastCreateReq)
	assert.Equal(t, uint(5), env.station.lastCreateReq.RegionID)
	assert.Equal(t, "武昌分站", env.station.lastCreateReq.Name)
	assert.Equal(t, 1, env.station.lastCreateReq.Status)
	var info dto.StationInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "武昌分站", info.Name)
}

func TestStationHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.createResult = nil
	env.station.createErr = errors.New("该地区已存在分站")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations", validCreateStationBody())

	assert.Equal(t, tenantHandler.CodeTenantError, resp.Code)
	assert.Equal(t, "该地区已存在分站", resp.Message)
}

// ---------- Update（admin） ----------

func TestStationHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/abc", validUpdateStationBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.station.lastUpdateID)
}

func TestStationHandler_Update_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/stations/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_Update_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// status oneof=0 1，传 9 → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3", map[string]interface{}{
		"status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3", validUpdateStationBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.station.lastUpdateID)
	require.NotNil(t, env.station.lastUpdateReq)
	require.NotNil(t, env.station.lastUpdateReq.Name)
	assert.Equal(t, "武昌新分站", *env.station.lastUpdateReq.Name)
}

func TestStationHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.updateErr = errors.New("域名已被其他分站占用")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3", validUpdateStationBody())

	assert.Equal(t, tenantHandler.CodeTenantError, resp.Code)
	assert.Equal(t, "域名已被其他分站占用", resp.Message)
}

// ---------- Delete（admin） ----------

func TestStationHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/stations/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.station.lastDeleteID)
}

func TestStationHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/stations/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.station.lastDeleteID)
}

func TestStationHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.deleteErr = errors.New("分站不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/stations/3", nil)

	assert.Equal(t, tenantHandler.CodeTenantError, resp.Code)
	assert.Equal(t, "分站不存在", resp.Message)
}

// ---------- UpdateStatus（admin） ----------

func TestStationHandler_UpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/abc/status", validUpdateStationStatusBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.station.lastUpdateStatID)
}

func TestStationHandler_UpdateStatus_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/stations/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_UpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// status oneof=0 1，传 9 → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3/status", map[string]interface{}{
		"status": 9,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_UpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3/status", validUpdateStationStatusBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态已更新", resp.Message)
	assert.Equal(t, uint(3), env.station.lastUpdateStatID)
	assert.Equal(t, 1, env.station.lastUpdateStat)
}

func TestStationHandler_UpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.updateStatusErr = errors.New("分站状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/stations/3/status", validUpdateStationStatusBody())

	assert.Equal(t, tenantHandler.CodeTenantStatusInvalid, resp.Code)
	assert.Equal(t, "分站状态不允许此操作", resp.Message)
}

// ---------- CopyConfig（admin） ----------

func TestStationHandler_CopyConfig_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/stations/copy-config", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_CopyConfig_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// source_station_id / target_station_id 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations/copy-config", map[string]interface{}{
		"biz_module": "mall",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStationHandler_CopyConfig_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations/copy-config", validCopyConfigBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "配置复制完成", resp.Message)
	require.NotNil(t, env.station.lastCopyConfigReq)
	assert.Equal(t, uint(1), env.station.lastCopyConfigReq.SourceStationID)
	assert.Equal(t, uint(2), env.station.lastCopyConfigReq.TargetStationID)
	assert.Equal(t, "mall", env.station.lastCopyConfigReq.BizModule)
	var r dto.CopyConfigResult
	require.NoError(t, json.Unmarshal(resp.Data, &r))
	assert.Equal(t, 3, r.CopiedCount)
}

func TestStationHandler_CopyConfig_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.station.copyConfigResult = nil
	env.station.copyConfigErr = errors.New("源分站与目标分站不能相同")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/stations/copy-config", validCopyConfigBody())

	assert.Equal(t, tenantHandler.CodeTenantCopyError, resp.Code)
	assert.Equal(t, "源分站与目标分站不能相同", resp.Message)
}

// ==================== StaffHandler ====================

// ---------- List ----------

func TestStaffHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff?page=1&page_size=10&station_id=1&role=operator", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.staff.lastListReq)
	assert.Equal(t, 1, env.staff.lastListReq.Page)
	assert.Equal(t, 10, env.staff.lastListReq.PageSize)
	assert.Equal(t, uint(1), env.staff.lastListReq.StationID)
	assert.Equal(t, "operator", env.staff.lastListReq.Role)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.StaffInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
}

func TestStaffHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t)
	env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff", nil)

	require.NotNil(t, env.staff.lastListReq)
	assert.Equal(t, 1, env.staff.lastListReq.Page)
	assert.Equal(t, 10, env.staff.lastListReq.PageSize)
}

func TestStaffHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.listResult = nil
	env.staff.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff", nil)

	assert.Equal(t, tenantHandler.CodeTenantStaffError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetByID ----------

func TestStaffHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(10), env.staff.lastGetByIDID)
	var info dto.StaffInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "manager", info.Role)
}

func TestStaffHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.staff.lastGetByIDID)
}

func TestStaffHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.getByIDResult = nil
	env.staff.getByIDErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/999", nil)

	assert.Equal(t, tenantHandler.CodeTenantStaffNotFound, resp.Code)
	assert.Equal(t, "员工不存在", resp.Message)
}

// ---------- Create ----------

func TestStaffHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/staff", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStaffHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// station_id / user_id / role 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/staff", map[string]interface{}{
		"station_id": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStaffHandler_Create_BindFail_InvalidRole(t *testing.T) {
	env := newHandlerEnv(t)
	// role oneof=operator manager，传 admin → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/staff", map[string]interface{}{
		"station_id": 1, "user_id": 100, "role": "admin",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStaffHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/staff", validCreateStaffBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "员工添加成功", resp.Message)
	require.NotNil(t, env.staff.lastCreateReq)
	assert.Equal(t, uint(1), env.staff.lastCreateReq.StationID)
	assert.Equal(t, uint(100), env.staff.lastCreateReq.UserID)
	assert.Equal(t, "operator", env.staff.lastCreateReq.Role)
	var info dto.StaffInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "operator", info.Role)
}

func TestStaffHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.createResult = nil
	env.staff.createErr = errors.New("该用户已是本分站员工")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/staff", validCreateStaffBody())

	assert.Equal(t, tenantHandler.CodeTenantStaffError, resp.Code)
	assert.Equal(t, "该用户已是本分站员工", resp.Message)
}

// ---------- Update ----------

func TestStaffHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/staff/abc", validUpdateStaffBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.staff.lastUpdateID)
}

func TestStaffHandler_Update_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/staff/10", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStaffHandler_Update_BindFail_InvalidRole(t *testing.T) {
	env := newHandlerEnv(t)
	// role oneof=operator manager，传 admin → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/staff/10", map[string]interface{}{
		"role": "admin",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestStaffHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/staff/10", validUpdateStaffBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(10), env.staff.lastUpdateID)
	require.NotNil(t, env.staff.lastUpdateReq)
	require.NotNil(t, env.staff.lastUpdateReq.Role)
	assert.Equal(t, "manager", *env.staff.lastUpdateReq.Role)
}

func TestStaffHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.updateErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/staff/10", validUpdateStaffBody())

	assert.Equal(t, tenantHandler.CodeTenantStaffError, resp.Code)
	assert.Equal(t, "员工不存在", resp.Message)
}

// ---------- Delete ----------

func TestStaffHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/staff/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.staff.lastDeleteID)
}

func TestStaffHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/staff/10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(10), env.staff.lastDeleteID)
}

func TestStaffHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.deleteErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/staff/10", nil)

	assert.Equal(t, tenantHandler.CodeTenantStaffError, resp.Code)
	assert.Equal(t, "员工不存在", resp.Message)
}

// ---------- ListByStation ----------

func TestStaffHandler_ListByStation_InvalidStationID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/by-station/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的分站ID", resp.Message)
	assert.Equal(t, uint(0), env.staff.lastListByStationID)
}

func TestStaffHandler_ListByStation_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/by-station/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.staff.lastListByStationID)
	var list []dto.StaffInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
	assert.Equal(t, "manager", list[0].Role)
}

func TestStaffHandler_ListByStation_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.staff.listByStationR = nil
	env.staff.listByStationErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/staff/by-station/1", nil)

	assert.Equal(t, tenantHandler.CodeTenantStaffError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== ConfigHandler ====================

// ---------- List ----------

func TestConfigHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs?page=1&page_size=10&biz_module=mall", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.config.lastListReq)
	assert.Equal(t, 1, env.config.lastListReq.Page)
	assert.Equal(t, 10, env.config.lastListReq.PageSize)
	assert.Equal(t, "mall", env.config.lastListReq.BizModule)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.ConfigInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "site_name", list[0].ConfigKey)
}

func TestConfigHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t)
	env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs", nil)

	require.NotNil(t, env.config.lastListReq)
	assert.Equal(t, 1, env.config.lastListReq.Page)
	assert.Equal(t, 10, env.config.lastListReq.PageSize)
}

func TestConfigHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.listResult = nil
	env.config.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs", nil)

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetByID ----------

func TestConfigHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.config.lastGetByIDID)
	var info dto.ConfigInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "site_name", info.ConfigKey)
}

func TestConfigHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.config.lastGetByIDID)
}

func TestConfigHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.getByIDResult = nil
	env.config.getByIDErr = errors.New("配置不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/999", nil)

	assert.Equal(t, tenantHandler.CodeTenantConfigNotFound, resp.Code)
	assert.Equal(t, "配置不存在", resp.Message)
}

// ---------- Upsert ----------

func TestConfigHandler_Upsert_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/configs", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestConfigHandler_Upsert_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// station_id / biz_module / config_key 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs", map[string]interface{}{
		"station_id": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestConfigHandler_Upsert_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs", validUpsertConfigBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "配置保存成功", resp.Message)
	require.NotNil(t, env.config.lastUpsertReq)
	assert.Equal(t, uint(1), env.config.lastUpsertReq.StationID)
	assert.Equal(t, "mall", env.config.lastUpsertReq.BizModule)
	assert.Equal(t, "site_name", env.config.lastUpsertReq.ConfigKey)
	assert.Equal(t, "武昌商城", env.config.lastUpsertReq.ConfigValue)
	var info dto.ConfigInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "site_name", info.ConfigKey)
}

func TestConfigHandler_Upsert_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.upsertResult = nil
	env.config.upsertErr = errors.New("所属分站不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs", validUpsertConfigBody())

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "所属分站不存在", resp.Message)
}

// ---------- Update ----------

func TestConfigHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/configs/abc", validUpdateConfigBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.config.lastUpdateID)
}

func TestConfigHandler_Update_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/configs/1", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestConfigHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/configs/1", validUpdateConfigBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.config.lastUpdateID)
	require.NotNil(t, env.config.lastUpdateReq)
	require.NotNil(t, env.config.lastUpdateReq.ConfigValue)
	assert.Equal(t, "新名称", *env.config.lastUpdateReq.ConfigValue)
}

func TestConfigHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.updateErr = errors.New("配置不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/configs/1", validUpdateConfigBody())

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "配置不存在", resp.Message)
}

// ---------- Delete ----------

func TestConfigHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/configs/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.config.lastDeleteID)
}

func TestConfigHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/configs/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.config.lastDeleteID)
}

func TestConfigHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.deleteErr = errors.New("配置不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/configs/1", nil)

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "配置不存在", resp.Message)
}

// ---------- ListByStationAndModule ----------

func TestConfigHandler_ListByStationAndModule_MissingParams(t *testing.T) {
	env := newHandlerEnv(t)
	// 缺 station_id / biz_module → 400
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/by-station-module", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "station_id 和 biz_module 参数不能为空", resp.Message)
}

func TestConfigHandler_ListByStationAndModule_MissingBizModule(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/by-station-module?station_id=1", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "station_id 和 biz_module 参数不能为空", resp.Message)
}

func TestConfigHandler_ListByStationAndModule_InvalidStationID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/by-station-module?station_id=abc&biz_module=mall", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的 station_id", resp.Message)
}

func TestConfigHandler_ListByStationAndModule_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/by-station-module?station_id=1&biz_module=mall", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.config.lastListBySMStationID)
	assert.Equal(t, "mall", env.config.lastListBySMModule)
	var list []dto.ConfigInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
}

func TestConfigHandler_ListByStationAndModule_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.listBySMResult = nil
	env.config.listBySMErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/configs/by-station-module?station_id=1&biz_module=mall", nil)

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ---------- BatchGet ----------

func TestConfigHandler_BatchGet_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/configs/batch-get", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestConfigHandler_BatchGet_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// station_id / biz_module 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs/batch-get", map[string]interface{}{
		"config_keys": []string{"site_name"},
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestConfigHandler_BatchGet_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs/batch-get", validBatchGetConfigBody())

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.config.lastBatchGetReq)
	assert.Equal(t, uint(1), env.config.lastBatchGetReq.StationID)
	assert.Equal(t, "mall", env.config.lastBatchGetReq.BizModule)
	require.Len(t, env.config.lastBatchGetReq.ConfigKeys, 2)
	var list []dto.ConfigKeyValue
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "site_name", list[0].ConfigKey)
}

func TestConfigHandler_BatchGet_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.config.batchGetResult = nil
	env.config.batchGetErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/configs/batch-get", validBatchGetConfigBody())

	assert.Equal(t, tenantHandler.CodeTenantConfigError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}

// ==================== DomainHandler ====================

// ---------- List ----------

func TestDomainHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains?page=1&page_size=10&station_id=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.domain.lastListReq)
	assert.Equal(t, 1, env.domain.lastListReq.Page)
	assert.Equal(t, 10, env.domain.lastListReq.PageSize)
	assert.Equal(t, uint(1), env.domain.lastListReq.StationID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
	var list []dto.DomainInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 1)
	assert.Equal(t, "wuchang.example.com", list[0].Domain)
}

func TestDomainHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t)
	env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains", nil)

	require.NotNil(t, env.domain.lastListReq)
	assert.Equal(t, 1, env.domain.lastListReq.Page)
	assert.Equal(t, 10, env.domain.lastListReq.PageSize)
}

func TestDomainHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.listResult = nil
	env.domain.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains", nil)

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- GetByID ----------

func TestDomainHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.domain.lastGetByIDID)
	var info dto.DomainInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "wuchang.example.com", info.Domain)
	assert.True(t, info.IsPrimary)
}

func TestDomainHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastGetByIDID)
}

func TestDomainHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.getByIDResult = nil
	env.domain.getByIDErr = errors.New("域名不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/999", nil)

	assert.Equal(t, tenantHandler.CodeTenantDomainNotFound, resp.Code)
	assert.Equal(t, "域名不存在", resp.Message)
}

// ---------- Create ----------

func TestDomainHandler_Create_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/tenant/admin/domains", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t)
	// station_id / domain 为 required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/domains", map[string]interface{}{
		"is_primary": true,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_Create_BindFail_InvalidSSLStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// ssl_status oneof=none pending active failed，传 invalid → Bind 失败
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/domains", map[string]interface{}{
		"station_id": 1, "domain": "x.example.com", "ssl_status": "invalid",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/domains", validCreateDomainBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "域名绑定成功", resp.Message)
	require.NotNil(t, env.domain.lastCreateReq)
	assert.Equal(t, uint(1), env.domain.lastCreateReq.StationID)
	assert.Equal(t, "wuchang.example.com", env.domain.lastCreateReq.Domain)
	assert.True(t, env.domain.lastCreateReq.IsPrimary)
	assert.Equal(t, "active", env.domain.lastCreateReq.SSLStatus)
	var info dto.DomainInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "wuchang.example.com", info.Domain)
}

func TestDomainHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.createResult = nil
	env.domain.createErr = errors.New("域名已被绑定")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/tenant/admin/domains", validCreateDomainBody())

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "域名已被绑定", resp.Message)
}

// ---------- Update ----------

func TestDomainHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/abc", validUpdateDomainBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastUpdateID)
}

func TestDomainHandler_Update_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/domains/1", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_Update_BindFail_InvalidSSLStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// ssl_status oneof=none pending active failed，传 invalid → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1", map[string]interface{}{
		"ssl_status": "invalid",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1", validUpdateDomainBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.domain.lastUpdateID)
	require.NotNil(t, env.domain.lastUpdateReq)
	require.NotNil(t, env.domain.lastUpdateReq.SSLStatus)
	assert.Equal(t, "active", *env.domain.lastUpdateReq.SSLStatus)
}

func TestDomainHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.updateErr = errors.New("域名不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1", validUpdateDomainBody())

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "域名不存在", resp.Message)
}

// ---------- Delete ----------

func TestDomainHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/domains/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastDeleteID)
}

func TestDomainHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/domains/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.domain.lastDeleteID)
}

func TestDomainHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.deleteErr = errors.New("域名不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/tenant/admin/domains/1", nil)

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "域名不存在", resp.Message)
}

// ---------- SetPrimary ----------

func TestDomainHandler_SetPrimary_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/abc/primary", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastSetPrimaryID)
}

func TestDomainHandler_SetPrimary_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/primary", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "已设为主域名", resp.Message)
	assert.Equal(t, uint(1), env.domain.lastSetPrimaryID)
}

func TestDomainHandler_SetPrimary_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.setPrimaryErr = errors.New("域名不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/primary", nil)

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "域名不存在", resp.Message)
}

// ---------- UpdateSSL ----------

func TestDomainHandler_UpdateSSL_InvalidID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/abc/ssl", validUpdateSSLBody())

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastUpdateSSLID)
}

func TestDomainHandler_UpdateSSL_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/ssl", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_UpdateSSL_BindFail_MissingSSLStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// ssl_status required，缺失 → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/ssl", map[string]interface{}{})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_UpdateSSL_BindFail_InvalidSSLStatus(t *testing.T) {
	env := newHandlerEnv(t)
	// ssl_status oneof=none pending active failed，传 invalid → Bind 失败
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/ssl", map[string]interface{}{
		"ssl_status": "invalid",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestDomainHandler_UpdateSSL_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/ssl", validUpdateSSLBody())

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "SSL 状态已更新", resp.Message)
	assert.Equal(t, uint(1), env.domain.lastUpdateSSLID)
	assert.Equal(t, "active", env.domain.lastUpdateSSLStatus)
}

func TestDomainHandler_UpdateSSL_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.updateSSLErr = errors.New("SSL 状态无效")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/tenant/admin/domains/1/ssl", validUpdateSSLBody())

	assert.Equal(t, tenantHandler.CodeTenantSSLInvalid, resp.Code)
	assert.Equal(t, "SSL 状态无效", resp.Message)
}

// ---------- ListByStation ----------

func TestDomainHandler_ListByStation_InvalidStationID(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/by-station/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的分站ID", resp.Message)
	assert.Equal(t, uint(0), env.domain.lastListByStationID)
}

func TestDomainHandler_ListByStation_Success(t *testing.T) {
	env := newHandlerEnv(t)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/by-station/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.domain.lastListByStationID)
	var list []dto.DomainInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 2)
	assert.True(t, list[0].IsPrimary)
}

func TestDomainHandler_ListByStation_ServiceError(t *testing.T) {
	env := newHandlerEnv(t)
	env.domain.listByStationRes = nil
	env.domain.listByStationErr = errors.New("db error")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/tenant/admin/domains/by-station/1", nil)

	assert.Equal(t, tenantHandler.CodeTenantDomainError, resp.Code)
	assert.Equal(t, "db error", resp.Message)
}
