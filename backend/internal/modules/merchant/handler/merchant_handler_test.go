// Package handler_test 商户中台 5 大 Handler HTTP 处理层装配单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 handler 装配层全部分支：
//   - ShopHandler：店铺列表/详情/搜索/入驻/更新/我的/认领 + M 端列表/详情/状态/信用分/等级
//   - StaffHandler：员工 CRUD + 详情/列表 + 权限分配 + 角色切换
//   - SettleHandler：结算列表/详情/按店铺 + 生成/提现/审核 + 按店铺/按周期汇总
//   - CategoryHandler：类目树/列表/详情 + CRUD + 状态更新
//   - VerificationHandler：认证提交/更新/删除 + 详情/列表/按店铺 + M 端列表/审核
//
// 鉴权由 AuthRequired / RequirePermission 中间件负责（测试中去掉，纯测 handler 装配层）。
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 ai/shop/tenant/material/dh114 handler 测试同风格。
package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wuchang-tongcheng/internal/core/middleware"
	coreRouter "wuchang-tongcheng/internal/core/router"
	"wuchang-tongcheng/internal/modules/merchant/dto"
	merchantHandler "wuchang-tongcheng/internal/modules/merchant/handler"
	"wuchang-tongcheng/internal/modules/merchant/service"
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

// ==================== mock ShopService ====================

type mockShopService struct {
	// List / Search
	lastListRegionID uint
	lastListReq      *dto.ShopListRequest
	listResult       []dto.ShopInfo
	listErr          error

	// Search
	lastSearchRegionID uint
	lastSearchReq      *dto.ShopListRequest

	// GetByID / AdminGetByID
	lastGetID    uint
	getByIDRes   *dto.ShopInfo
	getByIDErr   error

	// Apply
	lastApplyRegionID uint
	lastApplyUserID   uint
	lastApplyReq      *dto.CreateShopRequest
	applyResult       *dto.ShopInfo
	applyErr          error

	// Update
	lastUpdateID     uint
	lastUpdateUserID  uint
	lastUpdateReq     *dto.UpdateShopRequest
	updateErr         error

	// ListMine
	lastListMineUserID uint
	listMineResult     []dto.ShopInfo
	listMineErr        error

	// Claim
	lastClaimShopID uint
	lastClaimUserID uint
	claimResult     *dto.ShopInfo
	claimErr        error

	// AdminList
	lastAdminListReq *dto.ShopAdminListRequest
	adminListResult  []dto.ShopInfo
	adminListErr     error

	// UpdateStatus
	lastUpdateStatusID  uint
	lastUpdateStatusVal int
	updateStatusErr     error

	// UpdateCreditScore
	lastCreditID     uint
	lastCreditDelta  int
	lastCreditReason string
	creditErr        error

	// UpdateLevel
	lastLevelID  uint
	lastLevelVal int
	levelErr     error
}

func (m *mockShopService) Apply(regionID uint, ownerID uint, req *dto.CreateShopRequest) (*dto.ShopInfo, error) {
	m.lastApplyRegionID = regionID
	m.lastApplyUserID = ownerID
	m.lastApplyReq = req
	return m.applyResult, m.applyErr
}

func (m *mockShopService) Update(id uint, operatorID uint, req *dto.UpdateShopRequest) error {
	m.lastUpdateID = id
	m.lastUpdateUserID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockShopService) GetByID(id uint) (*dto.ShopInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockShopService) List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockShopService) Search(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastSearchRegionID = regionID
	m.lastSearchReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockShopService) ListMine(ownerID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastListMineUserID = ownerID
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	pg := utils.NewPagination(page, pageSize)
	pg.Total = int64(len(m.listMineResult))
	return pg, m.listMineResult, nil
}

func (m *mockShopService) Claim(shopID, userID uint) (*dto.ShopInfo, error) {
	m.lastClaimShopID = shopID
	m.lastClaimUserID = userID
	return m.claimResult, m.claimErr
}

func (m *mockShopService) UpdateCreditScore(id uint, delta int, reason string) error {
	m.lastCreditID = id
	m.lastCreditDelta = delta
	m.lastCreditReason = reason
	return m.creditErr
}

func (m *mockShopService) UpdateLevel(id uint, level int) error {
	m.lastLevelID = id
	m.lastLevelVal = level
	return m.levelErr
}

func (m *mockShopService) AdminList(req *dto.ShopAdminListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	m.lastAdminListReq = req
	if m.adminListErr != nil {
		return nil, nil, m.adminListErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.adminListResult))
	return pg, m.adminListResult, nil
}

func (m *mockShopService) AdminGetByID(id uint) (*dto.ShopInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockShopService) UpdateStatus(id uint, status int) error {
	m.lastUpdateStatusID = id
	m.lastUpdateStatusVal = status
	return m.updateStatusErr
}

var _ service.ShopService = (*mockShopService)(nil)

// ==================== mock StaffService ====================

type mockStaffService struct {
	// Create
	lastCreateReq *dto.CreateStaffRequest
	createResult  *dto.StaffInfo
	createErr     error

	// Update
	lastUpdateID     uint
	lastUpdateUserID uint
	lastUpdateReq    *dto.UpdateStaffRequest
	updateErr        error

	// Delete
	lastDeleteID     uint
	lastDeleteUserID uint
	deleteErr        error

	// GetByID
	lastGetID  uint
	getByIDRes *dto.StaffInfo
	getByIDErr error

	// List
	lastListReq *dto.StaffListRequest
	listResult  []dto.StaffInfo
	listErr     error

	// AssignPermissions
	lastAssignID     uint
	lastAssignUserID uint
	lastAssignPerms  interface{}
	assignErr        error

	// SwitchRole
	lastSwitchID     uint
	lastSwitchUserID uint
	lastSwitchRole   string
	switchErr        error
}

func (m *mockStaffService) Create(req *dto.CreateStaffRequest) (*dto.StaffInfo, error) {
	m.lastCreateReq = req
	return m.createResult, m.createErr
}

func (m *mockStaffService) Update(id uint, operatorID uint, req *dto.UpdateStaffRequest) error {
	m.lastUpdateID = id
	m.lastUpdateUserID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockStaffService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteUserID = operatorID
	return m.deleteErr
}

func (m *mockStaffService) GetByID(id uint) (*dto.StaffInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockStaffService) List(req *dto.StaffListRequest) (*utils.Pagination, []dto.StaffInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockStaffService) ListByShop(shopID uint) ([]dto.StaffInfo, error) {
	return m.listResult, m.listErr
}

func (m *mockStaffService) ListByUser(userID uint) ([]dto.StaffInfo, error) {
	return m.listResult, m.listErr
}

func (m *mockStaffService) AssignPermissions(id uint, operatorID uint, permissions interface{}) error {
	m.lastAssignID = id
	m.lastAssignUserID = operatorID
	m.lastAssignPerms = permissions
	return m.assignErr
}

func (m *mockStaffService) SwitchRole(id uint, operatorID uint, role string) error {
	m.lastSwitchID = id
	m.lastSwitchUserID = operatorID
	m.lastSwitchRole = role
	return m.switchErr
}

var _ service.StaffService = (*mockStaffService)(nil)

// ==================== mock SettleService ====================

type mockSettleService struct {
	// List
	lastListReq *dto.SettleListRequest
	listResult  []dto.SettleInfo
	listErr     error

	// GetByID
	lastGetID  uint
	getByIDRes *dto.SettleInfo
	getByIDErr error

	// ListByShop
	lastListByShopID uint
	listByShopResult []dto.SettleInfo
	listByShopErr    error

	// Generate
	lastGenerateReq *dto.SettleGenerateRequest
	generateResult   *dto.SettleInfo
	generateErr      error

	// Withdraw
	lastWithdrawID uint
	withdrawErr    error

	// AuditWithdraw
	lastAuditID  uint
	lastAuditReq *dto.SettleAuditRequest
	auditErr     error

	// SummaryByShop
	lastSummaryShopID uint
	summaryByShopRes  *dto.SettleSummary
	summaryByShopErr  error

	// SummaryByPeriod
	lastSummaryPeriod string
	summaryByPerRes   *dto.SettleSummary
	summaryByPerErr   error
}

func (m *mockSettleService) Generate(req *dto.SettleGenerateRequest) (*dto.SettleInfo, error) {
	m.lastGenerateReq = req
	return m.generateResult, m.generateErr
}

func (m *mockSettleService) GetByID(id uint) (*dto.SettleInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockSettleService) List(req *dto.SettleListRequest) (*utils.Pagination, []dto.SettleInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockSettleService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.SettleInfo, error) {
	m.lastListByShopID = shopID
	if m.listByShopErr != nil {
		return nil, nil, m.listByShopErr
	}
	pg := utils.NewPagination(page, pageSize)
	pg.Total = int64(len(m.listByShopResult))
	return pg, m.listByShopResult, nil
}

func (m *mockSettleService) Withdraw(id uint) error {
	m.lastWithdrawID = id
	return m.withdrawErr
}

func (m *mockSettleService) AuditWithdraw(id uint, req *dto.SettleAuditRequest) error {
	m.lastAuditID = id
	m.lastAuditReq = req
	return m.auditErr
}

func (m *mockSettleService) SummaryByShop(shopID uint) (*dto.SettleSummary, error) {
	m.lastSummaryShopID = shopID
	return m.summaryByShopRes, m.summaryByShopErr
}

func (m *mockSettleService) SummaryByPeriod(period string) (*dto.SettleSummary, error) {
	m.lastSummaryPeriod = period
	return m.summaryByPerRes, m.summaryByPerErr
}

var _ service.SettleService = (*mockSettleService)(nil)

// ==================== mock CategoryService ====================

type mockCategoryService struct {
	// Create
	lastCreateReq *dto.CreateCategoryRequest
	createResult  *dto.CategoryInfo
	createErr     error

	// Update
	lastUpdateID  uint
	lastUpdateReq *dto.UpdateCategoryRequest
	updateErr     error

	// Delete
	lastDeleteID uint
	deleteErr    error

	// GetByID
	lastGetID  uint
	getByIDRes *dto.CategoryInfo
	getByIDErr error

	// List
	lastListReq *dto.CategoryListRequest
	listResult  []dto.CategoryInfo
	listErr     error

	// Tree
	treeResult []dto.CategoryInfo
	treeErr    error

	// UpdateStatus
	lastStatusID  uint
	lastStatusVal int
	statusErr     error
}

func (m *mockCategoryService) Create(req *dto.CreateCategoryRequest) (*dto.CategoryInfo, error) {
	m.lastCreateReq = req
	return m.createResult, m.createErr
}

func (m *mockCategoryService) Update(id uint, req *dto.UpdateCategoryRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockCategoryService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockCategoryService) GetByID(id uint) (*dto.CategoryInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockCategoryService) List(req *dto.CategoryListRequest) (*utils.Pagination, []dto.CategoryInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockCategoryService) Tree() ([]dto.CategoryInfo, error) {
	return m.treeResult, m.treeErr
}

func (m *mockCategoryService) UpdateStatus(id uint, status int) error {
	m.lastStatusID = id
	m.lastStatusVal = status
	return m.statusErr
}

var _ service.CategoryService = (*mockCategoryService)(nil)

// ==================== mock VerificationService ====================

type mockVerificationService struct {
	// Create
	lastCreateRegionID uint
	lastCreateUserID   uint
	lastCreateReq      *dto.CreateVerificationRequest
	createResult       *dto.VerificationInfo
	createErr          error

	// Update
	lastUpdateID     uint
	lastUpdateUserID uint
	lastUpdateReq    *dto.UpdateVerificationRequest
	updateErr        error

	// Delete
	lastDeleteID     uint
	lastDeleteUserID uint
	deleteErr        error

	// GetByID
	lastGetID  uint
	getByIDRes *dto.VerificationInfo
	getByIDErr error

	// List / AdminList
	lastListReq *dto.VerificationListRequest
	listResult  []dto.VerificationInfo
	listErr     error

	// ListByShop
	lastListByShopID uint
	listByShopResult []dto.VerificationInfo
	listByShopErr    error

	// Audit
	lastAuditID  uint
	lastAuditReq *dto.VerificationAuditRequest
	auditErr     error
}

func (m *mockVerificationService) Create(regionID uint, userID uint, req *dto.CreateVerificationRequest) (*dto.VerificationInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateUserID = userID
	m.lastCreateReq = req
	return m.createResult, m.createErr
}

func (m *mockVerificationService) Update(id uint, operatorID uint, req *dto.UpdateVerificationRequest) error {
	m.lastUpdateID = id
	m.lastUpdateUserID = operatorID
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockVerificationService) Delete(id uint, operatorID uint) error {
	m.lastDeleteID = id
	m.lastDeleteUserID = operatorID
	return m.deleteErr
}

func (m *mockVerificationService) GetByID(id uint) (*dto.VerificationInfo, error) {
	m.lastGetID = id
	return m.getByIDRes, m.getByIDErr
}

func (m *mockVerificationService) List(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockVerificationService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error) {
	m.lastListByShopID = shopID
	if m.listByShopErr != nil {
		return nil, nil, m.listByShopErr
	}
	pg := utils.NewPagination(page, pageSize)
	pg.Total = int64(len(m.listByShopResult))
	return pg, m.listByShopResult, nil
}

func (m *mockVerificationService) AdminList(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	pg := utils.NewPagination(req.Page, req.PageSize)
	pg.Total = int64(len(m.listResult))
	return pg, m.listResult, nil
}

func (m *mockVerificationService) Audit(id uint, req *dto.VerificationAuditRequest) error {
	m.lastAuditID = id
	m.lastAuditReq = req
	return m.auditErr
}

var _ service.VerificationService = (*mockVerificationService)(nil)

// ==================== 测试环境 ====================

// handlerEnv merchant handler 测试环境
type handlerEnv struct {
	engine *gin.Engine
	shop   *mockShopService
	staff  *mockStaffService
	settle *mockSettleService
	cate   *mockCategoryService
	verify *mockVerificationService
}

// newHandlerEnv 构造 gin 引擎并注册 merchant 全部路由（路径与 merchant/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉 AuthRequired / RequirePermission 中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	shopMock := &mockShopService{
		listResult: []dto.ShopInfo{
			{ID: 1, RegionID: regionID, OwnerID: ctxUserID, Name: "武昌餐饮", Status: 1, CreditScore: 100, Level: 1},
			{ID: 2, RegionID: regionID, OwnerID: ctxUserID, Name: "洪山超市", Status: 1, CreditScore: 95, Level: 2},
		},
		getByIDRes: &dto.ShopInfo{
			ID: 1, RegionID: regionID, OwnerID: ctxUserID, Name: "武昌餐饮", Status: 1, CreditScore: 100, Level: 1,
		},
		applyResult: &dto.ShopInfo{
			ID: 3, RegionID: regionID, OwnerID: ctxUserID, Name: "新入驻店铺", Status: 0, CreditScore: 100, Level: 1,
		},
		listMineResult: []dto.ShopInfo{
			{ID: 1, RegionID: regionID, OwnerID: ctxUserID, Name: "武昌餐饮", Status: 1},
		},
		claimResult: &dto.ShopInfo{
			ID: 4, RegionID: regionID, OwnerID: ctxUserID, Name: "认领店铺", Status: 1,
		},
		adminListResult: []dto.ShopInfo{
			{ID: 1, RegionID: regionID, Name: "武昌餐饮", Status: 1},
		},
	}

	staffMock := &mockStaffService{
		createResult: &dto.StaffInfo{ID: 10, ShopID: 1, UserID: 100, Role: "clerk", RoleText: "店员", Status: 1},
		getByIDRes:   &dto.StaffInfo{ID: 10, ShopID: 1, UserID: 100, Role: "manager", RoleText: "管理员", Status: 1},
		listResult: []dto.StaffInfo{
			{ID: 10, ShopID: 1, UserID: 100, Role: "manager", Status: 1},
		},
	}

	settleMock := &mockSettleService{
		listResult: []dto.SettleInfo{
			{ID: 1, ShopID: 1, Period: "2026-07", TotalAmount: 10000, PlatformFee: 500, ShopAmount: 9500, Status: 1},
		},
		getByIDRes: &dto.SettleInfo{
			ID: 1, ShopID: 1, Period: "2026-07", TotalAmount: 10000, PlatformFee: 500, ShopAmount: 9500, Status: 1,
		},
		listByShopResult: []dto.SettleInfo{
			{ID: 1, ShopID: 1, Period: "2026-07", Status: 1},
		},
		generateResult: &dto.SettleInfo{
			ID: 2, ShopID: 1, Period: "2026-08", TotalAmount: 8000, PlatformFee: 400, ShopAmount: 7600, Status: 0,
		},
		summaryByShopRes:  &dto.SettleSummary{ShopID: 1, TotalAmount: 10000, PlatformFee: 500, ShopAmount: 9500, Count: 1},
		summaryByPerRes:   &dto.SettleSummary{Period: "2026-07", TotalAmount: 10000, PlatformFee: 500, ShopAmount: 9500, Count: 1},
	}

	cateMock := &mockCategoryService{
		treeResult: []dto.CategoryInfo{
			{ID: 1, ParentID: 0, Name: "餐饮美食", Status: 1, Children: []dto.CategoryInfo{
				{ID: 2, ParentID: 1, Name: "中餐", Status: 1},
			}},
		},
		listResult: []dto.CategoryInfo{
			{ID: 1, ParentID: 0, Name: "餐饮美食", Status: 1},
			{ID: 2, ParentID: 1, Name: "中餐", Status: 1},
		},
		getByIDRes:  &dto.CategoryInfo{ID: 1, ParentID: 0, Name: "餐饮美食", Status: 1},
		createResult: &dto.CategoryInfo{ID: 3, ParentID: 0, Name: "生活服务", Status: 1},
	}

	verifyMock := &mockVerificationService{
		createResult: &dto.VerificationInfo{ID: 1, ShopID: 1, RegionID: regionID, Type: "business", TypeText: "企业认证", Status: 0, StatusText: "待审"},
		getByIDRes:   &dto.VerificationInfo{ID: 1, ShopID: 1, Type: "business", TypeText: "企业认证", Status: 0, StatusText: "待审"},
		listResult: []dto.VerificationInfo{
			{ID: 1, ShopID: 1, Type: "business", Status: 0},
		},
		listByShopResult: []dto.VerificationInfo{
			{ID: 1, ShopID: 1, Type: "business", Status: 0},
		},
	}

	r := coreRouter.NewRouter()
	// 模拟 Region + Auth 中间件：注入 region_id 和 user_id
	r.Use(func(c *gin.Context) {
		c.Set(middleware.RegionIDKey, regionID)
		c.Set(middleware.ContextUserID, ctxUserID)
		c.Next()
	})

	sh := merchantHandler.NewShopHandler(shopMock)
	sth := merchantHandler.NewStaffHandler(staffMock)
	seh := merchantHandler.NewSettleHandler(settleMock)
	ch := merchantHandler.NewCategoryHandler(cateMock)
	vh := merchantHandler.NewVerificationHandler(verifyMock)

	root := r.Group("/api/v1/merchant")

	// 公开路由（与 plugin.go 一致）
	root.GET("/shops", sh.List)
	root.GET("/shops/search", sh.Search)
	root.GET("/shops/:id", sh.GetByID)
	root.GET("/categories/tree", ch.Tree)
	root.GET("/categories", ch.List)
	root.GET("/categories/:id", ch.GetByID)
	root.GET("/staff/:id", sth.GetByID)
	root.GET("/staff", sth.List)
	root.GET("/settles", seh.List)
	root.GET("/settles/:id", seh.GetByID)
	root.GET("/shops/:id/settles", seh.ListByShop)
	root.GET("/verifications", vh.List)
	root.GET("/verifications/:id", vh.GetByID)
	root.GET("/shops/:id/verifications", vh.ListByShop)

	// 需登录路由
	root.POST("/shops/apply", sh.Apply)
	root.POST("/shops/claim", sh.Claim)
	root.GET("/shops/mine", sh.ListMine)
	root.PUT("/shops/:id", sh.Update)
	root.POST("/staff", sth.Create)
	root.PUT("/staff/:id", sth.Update)
	root.DELETE("/staff/:id", sth.Delete)
	root.PUT("/staff/:id/permissions", sth.AssignPermissions)
	root.PUT("/staff/:id/role", sth.SwitchRole)
	root.POST("/verifications", vh.Create)
	root.PUT("/verifications/:id", vh.Update)
	root.DELETE("/verifications/:id", vh.Delete)

	// 管理后台路由
	admin := root.Group("/admin")
	admin.GET("/shops", sh.AdminList)
	admin.GET("/shops/:id", sh.AdminGetByID)
	admin.PUT("/shops/:id/status", sh.UpdateStatus)
	admin.PUT("/shops/:id/credit", sh.UpdateCreditScore)
	admin.PUT("/shops/:id/level", sh.UpdateLevel)
	admin.GET("/settles", seh.List)
	admin.GET("/settles/:id", seh.GetByID)
	admin.POST("/settles", seh.Generate)
	admin.PUT("/settles/:id/withdraw", seh.Withdraw)
	admin.PUT("/settles/:id/audit", seh.AuditWithdraw)
	admin.GET("/settles/summary-by-shop", seh.SummaryByShop)
	admin.GET("/settles/summary-by-period", seh.SummaryByPeriod)
	admin.GET("/categories", ch.List)
	admin.GET("/categories/:id", ch.GetByID)
	admin.POST("/categories", ch.Create)
	admin.PUT("/categories/:id", ch.Update)
	admin.DELETE("/categories/:id", ch.Delete)
	admin.PUT("/categories/:id/status", ch.UpdateStatus)
	admin.GET("/verifications", vh.AdminList)
	admin.PUT("/verifications/:id/audit", vh.Audit)

	return &handlerEnv{
		engine: r.Engine(),
		shop:   shopMock,
		staff:  staffMock,
		settle: settleMock,
		cate:   cateMock,
		verify: verifyMock,
	}
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

// assertParamError 断言 Bind 失败响应（code=5722，消息以 "参数错误" 开头）
func assertParamError(t *testing.T, resp *apiResponse) {
	t.Helper()
	assert.Equal(t, 5722, resp.Code)
	assert.True(t, strings.HasPrefix(resp.Message, "参数错误"), "expected message start with 参数错误, got: %s", resp.Message)
}

// assertUnauthorized 断言未登录响应
func assertUnauthorized(t *testing.T, resp *apiResponse) {
	t.Helper()
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

// ==================== ShopHandler ====================

func TestShopHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops?page=2&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "success", resp.Message)
	require.NotNil(t, env.shop.lastListReq)
	assert.Equal(t, uint(5), env.shop.lastListRegionID)
	assert.Equal(t, 2, env.shop.lastListReq.Page)
	assert.Equal(t, 5, env.shop.lastListReq.PageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	assert.Equal(t, 2, p.Page)
	assert.Equal(t, 5, p.PageSize)
}

func TestShopHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.shop.lastListReq)
	// 未传 page 时回退默认值 1/10
	assert.Equal(t, 1, env.shop.lastListReq.Page)
	assert.Equal(t, 10, env.shop.lastListReq.PageSize)
}

func TestShopHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.listErr = errors.New("数据库不可用")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops", nil)

	assert.Equal(t, 5701, resp.Code)
	assert.Equal(t, "数据库不可用", resp.Message)
}

func TestShopHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.shop.lastGetID)
	var info dto.ShopInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "武昌餐饮", info.Name)
	assert.Equal(t, 1, info.Status)
}

func TestShopHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.getByIDErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/99", nil)

	assert.Equal(t, 5702, resp.Code)
	assert.Equal(t, "店铺不存在", resp.Message)
}

func TestShopHandler_Search_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/search?keyword=餐饮&page=1&page_size=8", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.shop.lastSearchRegionID)
	require.NotNil(t, env.shop.lastSearchReq)
	assert.Equal(t, "餐饮", env.shop.lastSearchReq.Keyword)
	assert.Equal(t, 1, env.shop.lastSearchReq.Page)
	assert.Equal(t, 8, env.shop.lastSearchReq.PageSize)
}

func TestShopHandler_Search_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/search", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.shop.lastSearchReq)
	assert.Equal(t, 1, env.shop.lastSearchReq.Page)
	assert.Equal(t, 10, env.shop.lastSearchReq.PageSize)
}

func TestShopHandler_Apply_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/apply", map[string]interface{}{
		"name": "新店铺",
	})

	assertUnauthorized(t, resp)
	assert.Nil(t, env.shop.lastApplyReq)
}

func TestShopHandler_Apply_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/merchant/shops/apply", "{bad json", "application/json")

	assertParamError(t, resp)
}

func TestShopHandler_Apply_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/apply", map[string]interface{}{})

	assertParamError(t, resp)
}

func TestShopHandler_Apply_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/apply", map[string]interface{}{
		"name": "新入驻店铺",
		"logo": "/uploads/logo.png",
		"intro": "主营餐饮",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "入驻申请已提交", resp.Message)
	assert.Equal(t, uint(5), env.shop.lastApplyRegionID)
	assert.Equal(t, uint(100), env.shop.lastApplyUserID)
	require.NotNil(t, env.shop.lastApplyReq)
	assert.Equal(t, "新入驻店铺", env.shop.lastApplyReq.Name)
	var info dto.ShopInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "新入驻店铺", info.Name)
}

func TestShopHandler_Apply_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.applyErr = errors.New("入驻失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/apply", map[string]interface{}{
		"name": "新店铺",
	})

	assert.Equal(t, 5719, resp.Code)
	assert.Equal(t, "入驻失败", resp.Message)
}

func TestShopHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/shops/1", map[string]interface{}{
		"name": "改名",
	})

	assertUnauthorized(t, resp)
}

func TestShopHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/shops/abc", map[string]interface{}{
		"name": "改名",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_Update_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/merchant/shops/1", "{bad", "application/json")

	assertParamError(t, resp)
}

func TestShopHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/shops/1", map[string]interface{}{
		"intro": "更新后的简介",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.shop.lastUpdateID)
	assert.Equal(t, uint(100), env.shop.lastUpdateUserID)
}

func TestShopHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.updateErr = errors.New("无权操作此店铺")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/shops/1", map[string]interface{}{
		"name": "改名",
	})

	assert.Equal(t, 5704, resp.Code)
	assert.Equal(t, "无权操作此店铺", resp.Message)
}

func TestShopHandler_ListMine_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/mine", nil)

	assertUnauthorized(t, resp)
	assert.Equal(t, uint(0), env.shop.lastListMineUserID)
}

func TestShopHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/mine?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(100), env.shop.lastListMineUserID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestShopHandler_Claim_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/claim", map[string]interface{}{
		"shop_id": 1,
	})

	assertUnauthorized(t, resp)
}

func TestShopHandler_Claim_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/claim", map[string]interface{}{})

	assertParamError(t, resp)
}

func TestShopHandler_Claim_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/claim", map[string]interface{}{
		"shop_id": 4,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "认领成功", resp.Message)
	assert.Equal(t, uint(4), env.shop.lastClaimShopID)
	assert.Equal(t, uint(100), env.shop.lastClaimUserID)
}

func TestShopHandler_Claim_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.claimErr = errors.New("店铺已被认领")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/shops/claim", map[string]interface{}{
		"shop_id": 4,
	})

	assert.Equal(t, 5720, resp.Code)
	assert.Equal(t, "店铺已被认领", resp.Message)
}

func TestShopHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/shops?status=1&keyword=武昌", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.shop.lastAdminListReq)
	assert.Equal(t, "武昌", env.shop.lastAdminListReq.Keyword)
}

func TestShopHandler_AdminList_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/shops", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.shop.lastAdminListReq)
	assert.Equal(t, 1, env.shop.lastAdminListReq.Page)
	assert.Equal(t, 10, env.shop.lastAdminListReq.PageSize)
}

func TestShopHandler_AdminGetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/shops/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.shop.lastGetID)
}

func TestShopHandler_AdminGetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/shops/xyz", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_AdminGetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.getByIDErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/shops/99", nil)

	assert.Equal(t, 5702, resp.Code)
}

func TestShopHandler_UpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/abc/status", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_UpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/status", map[string]interface{}{
		"status": 9, // oneof=0 1 2
	})

	assertParamError(t, resp)
}

func TestShopHandler_UpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/status", map[string]interface{}{
		"status": 2,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(1), env.shop.lastUpdateStatusID)
	assert.Equal(t, 2, env.shop.lastUpdateStatusVal)
}

func TestShopHandler_UpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.updateStatusErr = errors.New("店铺状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/status", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 5705, resp.Code)
}

func TestShopHandler_UpdateCreditScore_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/abc/credit", map[string]interface{}{
		"delta":  -10,
		"reason": "违规",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_UpdateCreditScore_BindFail_MissingDelta(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/credit", map[string]interface{}{
		"reason": "违规",
	})

	assertParamError(t, resp)
}

func TestShopHandler_UpdateCreditScore_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/credit", map[string]interface{}{
		"delta":  -10,
		"reason": "违规扣分",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "信用分调整成功", resp.Message)
	assert.Equal(t, uint(1), env.shop.lastCreditID)
	assert.Equal(t, -10, env.shop.lastCreditDelta)
	assert.Equal(t, "违规扣分", env.shop.lastCreditReason)
}

func TestShopHandler_UpdateCreditScore_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.creditErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/99/credit", map[string]interface{}{
		"delta": -10,
	})

	assert.Equal(t, 5702, resp.Code)
}

func TestShopHandler_UpdateLevel_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/abc/level", map[string]interface{}{
		"level": 3,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestShopHandler_UpdateLevel_BindFail_OutOfRange(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/level", map[string]interface{}{
		"level": 99, // min=1,max=10
	})

	assertParamError(t, resp)
}

func TestShopHandler_UpdateLevel_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/1/level", map[string]interface{}{
		"level": 3,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "等级调整成功", resp.Message)
	assert.Equal(t, uint(1), env.shop.lastLevelID)
	assert.Equal(t, 3, env.shop.lastLevelVal)
}

func TestShopHandler_UpdateLevel_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.shop.levelErr = errors.New("店铺不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/shops/99/level", map[string]interface{}{
		"level": 3,
	})

	assert.Equal(t, 5702, resp.Code)
}

// ==================== StaffHandler ====================

func TestStaffHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/staff", map[string]interface{}{
		"shop_id": 1, "user_id": 100,
	})

	assertUnauthorized(t, resp)
	assert.Nil(t, env.staff.lastCreateReq)
}

func TestStaffHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/staff", map[string]interface{}{
		"shop_id": 1, // 缺 user_id
	})

	assertParamError(t, resp)
}

func TestStaffHandler_Create_BindFail_InvalidRole(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/staff", map[string]interface{}{
		"shop_id": 1, "user_id": 100, "role": "boss", // oneof=owner manager clerk
	})

	assertParamError(t, resp)
}

func TestStaffHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/staff", map[string]interface{}{
		"shop_id": 1, "user_id": 100, "role": "clerk",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "员工添加成功", resp.Message)
	require.NotNil(t, env.staff.lastCreateReq)
	assert.Equal(t, uint(1), env.staff.lastCreateReq.ShopID)
	assert.Equal(t, uint(100), env.staff.lastCreateReq.UserID)
	assert.Equal(t, "clerk", env.staff.lastCreateReq.Role)
	var info dto.StaffInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "clerk", info.Role)
}

func TestStaffHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.createErr = errors.New("员工已存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/staff", map[string]interface{}{
		"shop_id": 1, "user_id": 100,
	})

	assert.Equal(t, 5707, resp.Code)
	assert.Equal(t, "员工已存在", resp.Message)
}

func TestStaffHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1", map[string]interface{}{
		"role": "manager",
	})

	assertUnauthorized(t, resp)
}

func TestStaffHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/abc", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestStaffHandler_Update_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1", map[string]interface{}{
		"status": 9, // oneof=1 2
	})

	assertParamError(t, resp)
}

func TestStaffHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.staff.lastUpdateID)
	assert.Equal(t, uint(100), env.staff.lastUpdateUserID)
}

func TestStaffHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.updateErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/99", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 5706, resp.Code)
}

func TestStaffHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/staff/1", nil)

	assertUnauthorized(t, resp)
}

func TestStaffHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/staff/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestStaffHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/staff/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.staff.lastDeleteID)
	assert.Equal(t, uint(100), env.staff.lastDeleteUserID)
}

func TestStaffHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.deleteErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/staff/99", nil)

	assert.Equal(t, 5706, resp.Code)
}

func TestStaffHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff/10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(10), env.staff.lastGetID)
	var info dto.StaffInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "manager", info.Role)
}

func TestStaffHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestStaffHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.getByIDErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff/99", nil)

	assert.Equal(t, 5706, resp.Code)
}

func TestStaffHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff?shop_id=1&role=manager&page=1&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.staff.lastListReq)
	assert.Equal(t, uint(1), env.staff.lastListReq.ShopID)
	assert.Equal(t, "manager", env.staff.lastListReq.Role)
	assert.Equal(t, 1, env.staff.lastListReq.Page)
	assert.Equal(t, 5, env.staff.lastListReq.PageSize)
}

func TestStaffHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.staff.lastListReq)
	assert.Equal(t, 1, env.staff.lastListReq.Page)
	assert.Equal(t, 10, env.staff.lastListReq.PageSize)
}

func TestStaffHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/staff", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestStaffHandler_AssignPermissions_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1/permissions", map[string]interface{}{
		"permissions": []string{"order:view"},
	})

	assertUnauthorized(t, resp)
}

func TestStaffHandler_AssignPermissions_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/abc/permissions", map[string]interface{}{
		"permissions": []string{"order:view"},
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestStaffHandler_AssignPermissions_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1/permissions", map[string]interface{}{
		"permissions": []string{"order:view", "order:edit"},
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "权限分配成功", resp.Message)
	assert.Equal(t, uint(1), env.staff.lastAssignID)
	assert.Equal(t, uint(100), env.staff.lastAssignUserID)
}

func TestStaffHandler_AssignPermissions_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.assignErr = errors.New("员工不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/99/permissions", map[string]interface{}{
		"permissions": []string{"order:view"},
	})

	assert.Equal(t, 5706, resp.Code)
}

func TestStaffHandler_SwitchRole_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1/role", map[string]interface{}{
		"role": "manager",
	})

	assertUnauthorized(t, resp)
}

func TestStaffHandler_SwitchRole_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/abc/role", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestStaffHandler_SwitchRole_BindFail_InvalidRole(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1/role", map[string]interface{}{
		"role": "boss", // oneof=owner manager clerk
	})

	assertParamError(t, resp)
}

func TestStaffHandler_SwitchRole_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/1/role", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "角色切换成功", resp.Message)
	assert.Equal(t, uint(1), env.staff.lastSwitchID)
	assert.Equal(t, uint(100), env.staff.lastSwitchUserID)
	assert.Equal(t, "manager", env.staff.lastSwitchRole)
}

func TestStaffHandler_SwitchRole_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.staff.switchErr = errors.New("员工角色无效")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/staff/99/role", map[string]interface{}{
		"role": "manager",
	})

	assert.Equal(t, 5708, resp.Code)
}

// ==================== SettleHandler ====================

func TestSettleHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles?shop_id=1&period=2026-07&page=1&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.settle.lastListReq)
	assert.Equal(t, uint(1), env.settle.lastListReq.ShopID)
	assert.Equal(t, "2026-07", env.settle.lastListReq.Period)
	assert.Equal(t, 1, env.settle.lastListReq.Page)
	assert.Equal(t, 5, env.settle.lastListReq.PageSize)
}

func TestSettleHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.settle.lastListReq)
	assert.Equal(t, 1, env.settle.lastListReq.Page)
	assert.Equal(t, 10, env.settle.lastListReq.PageSize)
}

func TestSettleHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestSettleHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.settle.lastGetID)
	var info dto.SettleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "2026-07", info.Period)
	assert.Equal(t, float64(10000), info.TotalAmount)
}

func TestSettleHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSettleHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.getByIDErr = errors.New("结算单不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/settles/99", nil)

	assert.Equal(t, 5709, resp.Code)
}

func TestSettleHandler_ListByShop_InvalidShopID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/abc/settles", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的店铺ID", resp.Message)
}

func TestSettleHandler_ListByShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/1/settles?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.settle.lastListByShopID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestSettleHandler_ListByShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.listByShopErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/1/settles", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestSettleHandler_Generate_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/settles", map[string]interface{}{
		"shop_id": 1, // 缺 period / total_amount
	})

	assertParamError(t, resp)
}

func TestSettleHandler_Generate_BindFail_InvalidJSON(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/merchant/admin/settles", "{bad", "application/json")

	assertParamError(t, resp)
}

func TestSettleHandler_Generate_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/settles", map[string]interface{}{
		"shop_id": 1, "period": "2026-08", "total_amount": 8000, "platform_rate": 0.05,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "结算单生成成功", resp.Message)
	require.NotNil(t, env.settle.lastGenerateReq)
	assert.Equal(t, uint(1), env.settle.lastGenerateReq.ShopID)
	assert.Equal(t, "2026-08", env.settle.lastGenerateReq.Period)
	assert.Equal(t, float64(8000), env.settle.lastGenerateReq.TotalAmount)
	assert.Equal(t, 0.05, env.settle.lastGenerateReq.PlatformRate)
}

func TestSettleHandler_Generate_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.generateErr = errors.New("该周期结算单已存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/settles", map[string]interface{}{
		"shop_id": 1, "period": "2026-08", "total_amount": 8000,
	})

	assert.Equal(t, 5710, resp.Code)
	assert.Equal(t, "该周期结算单已存在", resp.Message)
}

func TestSettleHandler_Withdraw_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/abc/withdraw", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSettleHandler_Withdraw_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/1/withdraw", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "提现申请已提交", resp.Message)
	assert.Equal(t, uint(1), env.settle.lastWithdrawID)
}

func TestSettleHandler_Withdraw_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.withdrawErr = errors.New("结算单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/1/withdraw", nil)

	assert.Equal(t, 5711, resp.Code)
}

func TestSettleHandler_AuditWithdraw_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/abc/audit", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSettleHandler_AuditWithdraw_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/1/audit", map[string]interface{}{
		"status": 9, // oneof=1 2 3
	})

	assertParamError(t, resp)
}

func TestSettleHandler_AuditWithdraw_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/1/audit", map[string]interface{}{
		"status": 1, "reason": "审核通过",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(1), env.settle.lastAuditID)
	require.NotNil(t, env.settle.lastAuditReq)
	assert.Equal(t, 1, env.settle.lastAuditReq.Status)
	assert.Equal(t, "审核通过", env.settle.lastAuditReq.Reason)
}

func TestSettleHandler_AuditWithdraw_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.auditErr = errors.New("结算单状态不允许此操作")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/settles/1/audit", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 5711, resp.Code)
}

func TestSettleHandler_SummaryByShop_MissingShopID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-shop", nil)

	assert.Equal(t, 5722, resp.Code)
	assert.Equal(t, "shop_id 必填", resp.Message)
}

func TestSettleHandler_SummaryByShop_InvalidShopID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-shop?shop_id=abc", nil)

	assert.Equal(t, 5722, resp.Code)
	assert.Equal(t, "无效的 shop_id", resp.Message)
}

func TestSettleHandler_SummaryByShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-shop?shop_id=1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.settle.lastSummaryShopID)
	var sum dto.SettleSummary
	require.NoError(t, json.Unmarshal(resp.Data, &sum))
	assert.Equal(t, uint(1), sum.ShopID)
	assert.Equal(t, float64(10000), sum.TotalAmount)
	assert.Equal(t, int64(1), sum.Count)
}

func TestSettleHandler_SummaryByShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.summaryByShopErr = errors.New("汇总失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-shop?shop_id=1", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestSettleHandler_SummaryByPeriod_MissingPeriod(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-period", nil)

	assert.Equal(t, 5722, resp.Code)
	assert.Equal(t, "period 必填", resp.Message)
}

func TestSettleHandler_SummaryByPeriod_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-period?period=2026-07", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "2026-07", env.settle.lastSummaryPeriod)
	var sum dto.SettleSummary
	require.NoError(t, json.Unmarshal(resp.Data, &sum))
	assert.Equal(t, "2026-07", sum.Period)
	assert.Equal(t, float64(10000), sum.TotalAmount)
}

func TestSettleHandler_SummaryByPeriod_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.settle.summaryByPerErr = errors.New("汇总失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/settles/summary-by-period?period=2026-07", nil)

	assert.Equal(t, 5701, resp.Code)
}

// ==================== CategoryHandler ====================

func TestCategoryHandler_Tree_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories/tree", nil)

	assert.Equal(t, 0, resp.Code)
	var tree []dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &tree))
	require.Len(t, tree, 1)
	assert.Equal(t, "餐饮美食", tree[0].Name)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, "中餐", tree[0].Children[0].Name)
}

func TestCategoryHandler_Tree_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.treeErr = errors.New("加载失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories/tree", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestCategoryHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories?parent_id=0&page=1&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.cate.lastListReq)
	assert.Equal(t, 1, env.cate.lastListReq.Page)
	assert.Equal(t, 5, env.cate.lastListReq.PageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
}

func TestCategoryHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.cate.lastListReq)
	assert.Equal(t, 1, env.cate.lastListReq.Page)
	assert.Equal(t, 10, env.cate.lastListReq.PageSize)
}

func TestCategoryHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestCategoryHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.cate.lastGetID)
	var info dto.CategoryInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "餐饮美食", info.Name)
}

func TestCategoryHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCategoryHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.getByIDErr = errors.New("类目不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/categories/99", nil)

	assert.Equal(t, 5713, resp.Code)
}

func TestCategoryHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/categories", map[string]interface{}{})

	assertParamError(t, resp)
}

func TestCategoryHandler_Create_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/categories", map[string]interface{}{
		"name": "新类目", "status": 9, // oneof=0 1
	})

	assertParamError(t, resp)
}

func TestCategoryHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/categories", map[string]interface{}{
		"parent_id": 0, "name": "生活服务", "sort": 5, "status": 1,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "类目创建成功", resp.Message)
	require.NotNil(t, env.cate.lastCreateReq)
	assert.Equal(t, "生活服务", env.cate.lastCreateReq.Name)
	assert.Equal(t, 5, env.cate.lastCreateReq.Sort)
}

func TestCategoryHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.createErr = errors.New("创建失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/admin/categories", map[string]interface{}{
		"name": "新类目",
	})

	assert.Equal(t, 5701, resp.Code)
}

func TestCategoryHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/abc", map[string]interface{}{
		"name": "改名",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCategoryHandler_Update_BindFail_NameTooLong(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/1", map[string]interface{}{
		"name": strings.Repeat("x", 100), // max=64
	})

	assertParamError(t, resp)
}

func TestCategoryHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/1", map[string]interface{}{
		"name": "改名",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.cate.lastUpdateID)
}

func TestCategoryHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.updateErr = errors.New("类目不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/99", map[string]interface{}{
		"name": "改名",
	})

	assert.Equal(t, 5713, resp.Code)
}

func TestCategoryHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/admin/categories/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCategoryHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/admin/categories/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.cate.lastDeleteID)
}

func TestCategoryHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.deleteErr = errors.New("该类目下有子类目，无法删除")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/admin/categories/1", nil)

	assert.Equal(t, 5714, resp.Code)
	assert.Equal(t, "该类目下有子类目，无法删除", resp.Message)
}

func TestCategoryHandler_UpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/abc/status", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCategoryHandler_UpdateStatus_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/1/status", map[string]interface{}{
		"status": 9, // oneof=0 1
	})

	assertParamError(t, resp)
}

func TestCategoryHandler_UpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/1/status", map[string]interface{}{
		"status": 0,
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态更新成功", resp.Message)
	assert.Equal(t, uint(1), env.cate.lastStatusID)
	assert.Equal(t, 0, env.cate.lastStatusVal)
}

func TestCategoryHandler_UpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.cate.statusErr = errors.New("类目状态无效")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/categories/1/status", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 5715, resp.Code)
}

// ==================== VerificationHandler ====================

func TestVerificationHandler_Create_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/verifications", map[string]interface{}{
		"shop_id": 1, "type": "business",
	})

	assertUnauthorized(t, resp)
	assert.Nil(t, env.verify.lastCreateReq)
}

func TestVerificationHandler_Create_BindFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/verifications", map[string]interface{}{
		"type": "business", // 缺 shop_id
	})

	assertParamError(t, resp)
}

func TestVerificationHandler_Create_BindFail_InvalidType(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/verifications", map[string]interface{}{
		"shop_id": 1, "type": "unknown", // oneof=business personal
	})

	assertParamError(t, resp)
}

func TestVerificationHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/verifications", map[string]interface{}{
		"shop_id": 1, "type": "business", "license_no": "91420100MA001", "legal_person": "张三",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "认证已提交", resp.Message)
	assert.Equal(t, uint(5), env.verify.lastCreateRegionID)
	assert.Equal(t, uint(100), env.verify.lastCreateUserID)
	require.NotNil(t, env.verify.lastCreateReq)
	assert.Equal(t, uint(1), env.verify.lastCreateReq.ShopID)
	assert.Equal(t, "business", env.verify.lastCreateReq.Type)
	assert.Equal(t, "张三", env.verify.lastCreateReq.LegalPerson)
	var info dto.VerificationInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "business", info.Type)
}

func TestVerificationHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.createErr = errors.New("认证提交失败")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/merchant/verifications", map[string]interface{}{
		"shop_id": 1, "type": "business",
	})

	assert.Equal(t, 5701, resp.Code)
	assert.Equal(t, "认证提交失败", resp.Message)
}

func TestVerificationHandler_Update_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/verifications/1", map[string]interface{}{
		"legal_person": "李四",
	})

	assertUnauthorized(t, resp)
}

func TestVerificationHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/verifications/abc", map[string]interface{}{
		"legal_person": "李四",
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestVerificationHandler_Update_BindFail_InvalidType(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/verifications/1", map[string]interface{}{
		"type": "unknown", // oneof=business personal
	})

	assertParamError(t, resp)
}

func TestVerificationHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/verifications/1", map[string]interface{}{
		"legal_person": "李四",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(1), env.verify.lastUpdateID)
	assert.Equal(t, uint(100), env.verify.lastUpdateUserID)
}

func TestVerificationHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.updateErr = errors.New("认证记录不存在")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/verifications/99", map[string]interface{}{
		"legal_person": "李四",
	})

	assert.Equal(t, 5716, resp.Code)
}

func TestVerificationHandler_Delete_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/verifications/1", nil)

	assertUnauthorized(t, resp)
}

func TestVerificationHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/verifications/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestVerificationHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/verifications/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(1), env.verify.lastDeleteID)
	assert.Equal(t, uint(100), env.verify.lastDeleteUserID)
}

func TestVerificationHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.deleteErr = errors.New("认证记录不存在")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/merchant/verifications/99", nil)

	assert.Equal(t, 5716, resp.Code)
}

func TestVerificationHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications/1", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.verify.lastGetID)
	var info dto.VerificationInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "business", info.Type)
	assert.Equal(t, "企业认证", info.TypeText)
}

func TestVerificationHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestVerificationHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.getByIDErr = errors.New("认证记录不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications/99", nil)

	assert.Equal(t, 5716, resp.Code)
}

func TestVerificationHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications?shop_id=1&type=business&page=1&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.verify.lastListReq)
	assert.Equal(t, uint(1), env.verify.lastListReq.ShopID)
	assert.Equal(t, "business", env.verify.lastListReq.Type)
	assert.Equal(t, 1, env.verify.lastListReq.Page)
	assert.Equal(t, 5, env.verify.lastListReq.PageSize)
}

func TestVerificationHandler_List_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.verify.lastListReq)
	assert.Equal(t, 1, env.verify.lastListReq.Page)
	assert.Equal(t, 10, env.verify.lastListReq.PageSize)
}

func TestVerificationHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.listErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/verifications", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestVerificationHandler_ListByShop_InvalidShopID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/abc/verifications", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的商户ID", resp.Message)
}

func TestVerificationHandler_ListByShop_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/1/verifications?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(1), env.verify.lastListByShopID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestVerificationHandler_ListByShop_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.listByShopErr = errors.New("查询失败")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/shops/1/verifications", nil)

	assert.Equal(t, 5701, resp.Code)
}

func TestVerificationHandler_AdminList_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/verifications?status=0&page=1&page_size=5", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.verify.lastListReq)
	assert.Equal(t, 5, env.verify.lastListReq.PageSize)
}

func TestVerificationHandler_AdminList_DefaultPagination(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/merchant/admin/verifications", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.verify.lastListReq)
	assert.Equal(t, 1, env.verify.lastListReq.Page)
	assert.Equal(t, 10, env.verify.lastListReq.PageSize)
}

func TestVerificationHandler_Audit_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/verifications/abc/audit", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestVerificationHandler_Audit_BindFail_InvalidStatus(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/verifications/1/audit", map[string]interface{}{
		"status": 9, // oneof=1 2
	})

	assertParamError(t, resp)
}

func TestVerificationHandler_Audit_Success(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/verifications/1/audit", map[string]interface{}{
		"status": 1, "audit_reason": "材料齐全",
	})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "审核完成", resp.Message)
	assert.Equal(t, uint(1), env.verify.lastAuditID)
	require.NotNil(t, env.verify.lastAuditReq)
	assert.Equal(t, 1, env.verify.lastAuditReq.Status)
	assert.Equal(t, "材料齐全", env.verify.lastAuditReq.AuditReason)
}

func TestVerificationHandler_Audit_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 100, 5)
	env.verify.auditErr = errors.New("认证已审核，不能重复审核")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/merchant/admin/verifications/1/audit", map[string]interface{}{
		"status": 1,
	})

	assert.Equal(t, 5717, resp.Code)
	assert.Equal(t, "认证已审核，不能重复审核", resp.Message)
}
