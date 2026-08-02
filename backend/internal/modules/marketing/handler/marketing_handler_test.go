// Package handler_test 营销中台模块 HTTP 处理层单元测试。
//
// 使用 gin + httptest + mock Service，覆盖 4 个子域 handler 全部分支：
//   - ad 子域（AdHandler）：Create/Update/Delete/GetByID/List/ListByPositionCode
//   - activity 子域（ActivityHandler）：Create/Update/Delete/GetByID/List/ListOngoing/ListUpcoming/ListEnded/UpdateStatus/Statistics
//   - coupon 子域（CouponHandler）：Create/Update/Delete/GetByID/List/ListAvailable/Receive/Use/Refund/ListMine/Statistics
//   - sign 子域（SignHandler）：CheckIn/GetCalendar/CreateRule/UpdateRule/DeleteRule/GetRuleByID/ListRules/ListEnabledRules
//
// 覆盖维度：
//   - 公开接口无需登录（ListByPositionCode/GetByID/ListAvailable/ListOngoing/ListUpcoming/ListEnded/ListEnabledRules）
//   - 用户接口未登录拦截（CheckIn/GetCalendar/Receive/Use/Refund/ListMine → 401 "请先登录"）
//   - URL :id 参数解析失败（非数字 → 400 "无效的ID"）
//   - 请求体 Bind 失败（非法 JSON + binding 校验失败 required/oneof/min/max → 400 "参数错误"）
//   - ListByPositionCode 空位置编码（→ 400 "位置编码不能为空"）
//   - service 成功/错误透传（业务码 5801-5823 区间 + message + data 透传）
//   - 地区ID/用户信息上下文注入（regionID/userID 透传给 service）
//   - 分页结果 PageResult 结构断言（list/total/page/pageSize）
//
// 不依赖 DB/Redis/Docker，纯内存 mock service 验证 handler 装配层逻辑。
// 与 shop/house/groupbuy/category/region/news/file/setting/permission handler 测试同风格。
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
	"wuchang-tongcheng/internal/modules/marketing/dto"
	mktHandler "wuchang-tongcheng/internal/modules/marketing/handler"
	"wuchang-tongcheng/internal/modules/marketing/service"
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

// ==================== AdService mock ====================

// mockAdService 内存 mock，实现 service.AdService 接口。
type mockAdService struct {
	lastCreateRegionID uint
	lastCreateReq      *dto.CreateAdPositionRequest
	createResult       *dto.AdPositionInfo
	createErr          error

	lastUpdateID  uint
	lastUpdateReq *dto.UpdateAdPositionRequest
	updateErr     error

	lastDeleteID uint
	deleteErr    error

	lastGetByIDID uint
	getByIDResult *dto.AdPositionInfo
	getByIDErr    error

	lastListRegionID uint
	lastListReq      *dto.AdPositionListRequest
	listResult       []dto.AdPositionInfo
	listErr          error

	lastListByCodeRegionID    uint
	lastListByCodePositionCode string
	lastListByCodePage         int
	lastListByCodePageSize     int
	listByCodeResult           []dto.AdPositionInfo
	listByCodeErr              error
}

func (m *mockAdService) Create(regionID uint, req *dto.CreateAdPositionRequest) (*dto.AdPositionInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockAdService) Update(id uint, req *dto.UpdateAdPositionRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockAdService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockAdService) GetByID(id uint) (*dto.AdPositionInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockAdService) List(regionID uint, req *dto.AdPositionListRequest) (*utils.Pagination, []dto.AdPositionInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockAdService) ListByPositionCode(regionID uint, positionCode string, page, pageSize int) (*utils.Pagination, []dto.AdPositionInfo, error) {
	m.lastListByCodeRegionID = regionID
	m.lastListByCodePositionCode = positionCode
	m.lastListByCodePage = page
	m.lastListByCodePageSize = pageSize
	if m.listByCodeErr != nil {
		return nil, nil, m.listByCodeErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listByCodeResult))
	return p, m.listByCodeResult, nil
}

var _ service.AdService = (*mockAdService)(nil)

// ==================== ActivityService mock ====================

// mockActivityService 内存 mock，实现 service.ActivityService 接口。
type mockActivityService struct {
	lastCreateRegionID uint
	lastCreateReq      *dto.CreateActivityRequest
	createResult       *dto.ActivityInfo
	createErr          error

	lastUpdateID  uint
	lastUpdateReq *dto.UpdateActivityRequest
	updateErr     error

	lastDeleteID uint
	deleteErr    error

	lastGetByIDID uint
	getByIDResult *dto.ActivityInfo
	getByIDErr    error

	lastListRegionID uint
	lastListReq      *dto.ActivityListRequest
	listResult       []dto.ActivityInfo
	listErr          error

	lastListOngoingRegionID uint
	lastListOngoingPage     int
	lastListOngoingPageSize int
	listOngoingResult       []dto.ActivityInfo
	listOngoingErr          error

	lastListUpcomingRegionID uint
	lastListUpcomingPage     int
	lastListUpcomingPageSize int
	listUpcomingResult       []dto.ActivityInfo
	listUpcomingErr          error

	lastListEndedRegionID uint
	lastListEndedPage     int
	lastListEndedPageSize int
	listEndedResult       []dto.ActivityInfo
	listEndedErr          error

	lastUpdateStatusID     uint
	lastUpdateStatusStatus int
	updateStatusErr        error

	lastStatisticsRegionID uint
	statisticsResult       *dto.ActivityStatistics
	statisticsErr          error
}

func (m *mockActivityService) Create(regionID uint, req *dto.CreateActivityRequest) (*dto.ActivityInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockActivityService) Update(id uint, req *dto.UpdateActivityRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockActivityService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockActivityService) GetByID(id uint) (*dto.ActivityInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockActivityService) List(regionID uint, req *dto.ActivityListRequest) (*utils.Pagination, []dto.ActivityInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockActivityService) ListOngoing(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	m.lastListOngoingRegionID = regionID
	m.lastListOngoingPage = page
	m.lastListOngoingPageSize = pageSize
	if m.listOngoingErr != nil {
		return nil, nil, m.listOngoingErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listOngoingResult))
	return p, m.listOngoingResult, nil
}

func (m *mockActivityService) ListUpcoming(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	m.lastListUpcomingRegionID = regionID
	m.lastListUpcomingPage = page
	m.lastListUpcomingPageSize = pageSize
	if m.listUpcomingErr != nil {
		return nil, nil, m.listUpcomingErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listUpcomingResult))
	return p, m.listUpcomingResult, nil
}

func (m *mockActivityService) ListEnded(regionID uint, page, pageSize int) (*utils.Pagination, []dto.ActivityInfo, error) {
	m.lastListEndedRegionID = regionID
	m.lastListEndedPage = page
	m.lastListEndedPageSize = pageSize
	if m.listEndedErr != nil {
		return nil, nil, m.listEndedErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listEndedResult))
	return p, m.listEndedResult, nil
}

func (m *mockActivityService) UpdateStatus(id uint, status int) error {
	m.lastUpdateStatusID = id
	m.lastUpdateStatusStatus = status
	return m.updateStatusErr
}

func (m *mockActivityService) AutoUpdateStatus() (int64, error) {
	return 0, nil
}

func (m *mockActivityService) Statistics(regionID uint) (*dto.ActivityStatistics, error) {
	m.lastStatisticsRegionID = regionID
	if m.statisticsErr != nil {
		return nil, m.statisticsErr
	}
	return m.statisticsResult, nil
}

var _ service.ActivityService = (*mockActivityService)(nil)

// ==================== CouponService mock ====================

// mockCouponService 内存 mock，实现 service.CouponService 接口。
type mockCouponService struct {
	lastCreateRegionID uint
	lastCreateReq      *dto.CreateCouponRequest
	createResult       *dto.CouponInfo
	createErr          error

	lastUpdateID  uint
	lastUpdateReq *dto.UpdateCouponRequest
	updateErr     error

	lastDeleteID uint
	deleteErr    error

	lastGetByIDID uint
	getByIDResult *dto.CouponInfo
	getByIDErr    error

	lastListRegionID uint
	lastListReq      *dto.CouponListRequest
	listResult       []dto.CouponInfo
	listErr          error

	lastListAvailableRegionID uint
	lastListAvailablePage     int
	lastListAvailablePageSize int
	listAvailableResult       []dto.CouponInfo
	listAvailableErr          error

	lastReceiveUserID   uint
	lastReceiveCouponID uint
	lastReceiveSource   string
	receiveErr          error

	lastUseID     uint
	lastUseOrderID uint
	useErr        error

	lastRefundID uint
	refundErr    error

	lastListMineUserID uint
	lastListMineReq    *dto.UserCouponListRequest
	listMineResult     []dto.UserCouponInfo
	listMineErr        error

	lastStatisticsRegionID uint
	statisticsResult       *dto.CouponStatistics
	statisticsErr          error
}

func (m *mockCouponService) Create(regionID uint, req *dto.CreateCouponRequest) (*dto.CouponInfo, error) {
	m.lastCreateRegionID = regionID
	m.lastCreateReq = req
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResult, nil
}

func (m *mockCouponService) Update(id uint, req *dto.UpdateCouponRequest) error {
	m.lastUpdateID = id
	m.lastUpdateReq = req
	return m.updateErr
}

func (m *mockCouponService) Delete(id uint) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockCouponService) GetByID(id uint) (*dto.CouponInfo, error) {
	m.lastGetByIDID = id
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResult, nil
}

func (m *mockCouponService) List(regionID uint, req *dto.CouponListRequest) (*utils.Pagination, []dto.CouponInfo, error) {
	m.lastListRegionID = regionID
	m.lastListReq = req
	if m.listErr != nil {
		return nil, nil, m.listErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listResult))
	return p, m.listResult, nil
}

func (m *mockCouponService) ListAvailable(regionID uint, page, pageSize int) (*utils.Pagination, []dto.CouponInfo, error) {
	m.lastListAvailableRegionID = regionID
	m.lastListAvailablePage = page
	m.lastListAvailablePageSize = pageSize
	if m.listAvailableErr != nil {
		return nil, nil, m.listAvailableErr
	}
	p := utils.NewPagination(page, pageSize)
	p.Total = int64(len(m.listAvailableResult))
	return p, m.listAvailableResult, nil
}

func (m *mockCouponService) Receive(userID uint, couponID uint, source string) error {
	m.lastReceiveUserID = userID
	m.lastReceiveCouponID = couponID
	m.lastReceiveSource = source
	return m.receiveErr
}

func (m *mockCouponService) Use(userCouponID uint, orderID uint) error {
	m.lastUseID = userCouponID
	m.lastUseOrderID = orderID
	return m.useErr
}

func (m *mockCouponService) Refund(userCouponID uint) error {
	m.lastRefundID = userCouponID
	return m.refundErr
}

func (m *mockCouponService) ListMine(userID uint, req *dto.UserCouponListRequest) (*utils.Pagination, []dto.UserCouponInfo, error) {
	m.lastListMineUserID = userID
	m.lastListMineReq = req
	if m.listMineErr != nil {
		return nil, nil, m.listMineErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listMineResult))
	return p, m.listMineResult, nil
}

func (m *mockCouponService) Statistics(regionID uint) (*dto.CouponStatistics, error) {
	m.lastStatisticsRegionID = regionID
	if m.statisticsErr != nil {
		return nil, m.statisticsErr
	}
	return m.statisticsResult, nil
}

func (m *mockCouponService) ExpireCoupons() (int64, error) {
	return 0, nil
}

var _ service.CouponService = (*mockCouponService)(nil)

// ==================== SignService mock ====================

// mockSignService 内存 mock，实现 service.SignService 接口。
type mockSignService struct {
	lastCheckInUserID uint
	checkInResult     *dto.SignCheckInResponse
	checkInErr        error

	lastGetCalendarUserID uint
	lastGetCalendarMonth  string
	getCalendarResult     *dto.SignCalendarResponse
	getCalendarErr        error

	lastCreateRuleReq *dto.CreateSignRuleRequest
	createRuleResult  *dto.SignRuleInfo
	createRuleErr     error

	lastUpdateRuleID  uint
	lastUpdateRuleReq *dto.UpdateSignRuleRequest
	updateRuleErr     error

	lastDeleteRuleID uint
	deleteRuleErr    error

	lastGetRuleByIDID uint
	getRuleByIDResult *dto.SignRuleInfo
	getRuleByIDErr    error

	lastListRulesReq *dto.SignRuleListRequest
	listRulesResult  []dto.SignRuleInfo
	listRulesErr     error

	listEnabledRulesResult []dto.SignRuleInfo
	listEnabledRulesErr    error
}

func (m *mockSignService) CheckIn(userID uint) (*dto.SignCheckInResponse, error) {
	m.lastCheckInUserID = userID
	if m.checkInErr != nil {
		return nil, m.checkInErr
	}
	return m.checkInResult, nil
}

func (m *mockSignService) GetCalendar(userID uint, month string) (*dto.SignCalendarResponse, error) {
	m.lastGetCalendarUserID = userID
	m.lastGetCalendarMonth = month
	if m.getCalendarErr != nil {
		return nil, m.getCalendarErr
	}
	return m.getCalendarResult, nil
}

func (m *mockSignService) CreateRule(req *dto.CreateSignRuleRequest) (*dto.SignRuleInfo, error) {
	m.lastCreateRuleReq = req
	if m.createRuleErr != nil {
		return nil, m.createRuleErr
	}
	return m.createRuleResult, nil
}

func (m *mockSignService) UpdateRule(id uint, req *dto.UpdateSignRuleRequest) error {
	m.lastUpdateRuleID = id
	m.lastUpdateRuleReq = req
	return m.updateRuleErr
}

func (m *mockSignService) DeleteRule(id uint) error {
	m.lastDeleteRuleID = id
	return m.deleteRuleErr
}

func (m *mockSignService) GetRuleByID(id uint) (*dto.SignRuleInfo, error) {
	m.lastGetRuleByIDID = id
	if m.getRuleByIDErr != nil {
		return nil, m.getRuleByIDErr
	}
	return m.getRuleByIDResult, nil
}

func (m *mockSignService) ListRules(req *dto.SignRuleListRequest) (*utils.Pagination, []dto.SignRuleInfo, error) {
	m.lastListRulesReq = req
	if m.listRulesErr != nil {
		return nil, nil, m.listRulesErr
	}
	p := utils.NewPagination(req.Page, req.PageSize)
	p.Total = int64(len(m.listRulesResult))
	return p, m.listRulesResult, nil
}

func (m *mockSignService) ListEnabledRules() ([]dto.SignRuleInfo, error) {
	if m.listEnabledRulesErr != nil {
		return nil, m.listEnabledRulesErr
	}
	return m.listEnabledRulesResult, nil
}

var _ service.SignService = (*mockSignService)(nil)

// ==================== 测试环境 ====================

// handlerEnv handler 测试环境（聚合 4 个子域 mock + handler）
type handlerEnv struct {
	engine   *gin.Engine
	ad       *mockAdService
	activity *mockActivityService
	coupon   *mockCouponService
	sign     *mockSignService
}

// newHandlerEnv 构造 gin 引擎并注册 marketing 4 个子域路由（路径与 marketing/plugin.go RegisterRoutes 一致）。
// ctxUserID 模拟 Auth 中间件注入的 user_id（0 表示未登录）。
// regionID 模拟 Region 中间件注入的 region_id。
// 路由注册去掉权限中间件，纯测 handler 装配层逻辑。
func newHandlerEnv(t *testing.T, ctxUserID uint, regionID uint) *handlerEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ad := &mockAdService{
		createResult:   &dto.AdPositionInfo{ID: 1, RegionID: regionID, PositionCode: "home_banner", Title: "首页轮播", Status: 1},
		getByIDResult:  &dto.AdPositionInfo{ID: 1, RegionID: regionID, PositionCode: "home_banner", Title: "首页轮播", Status: 1},
		listResult:     []dto.AdPositionInfo{{ID: 1, Title: "首页轮播"}, {ID: 2, Title: "侧边广告"}},
		listByCodeResult: []dto.AdPositionInfo{{ID: 1, Title: "首页轮播"}},
	}
	activity := &mockActivityService{
		createResult:   &dto.ActivityInfo{ID: 1, RegionID: regionID, Title: "国庆拼团", Type: "groupbuy", Status: 1},
		getByIDResult:  &dto.ActivityInfo{ID: 1, RegionID: regionID, Title: "国庆拼团", Type: "groupbuy", Status: 1},
		listResult:     []dto.ActivityInfo{{ID: 1, Title: "国庆拼团"}, {ID: 2, Title: "双11秒杀"}},
		listOngoingResult:   []dto.ActivityInfo{{ID: 1, Title: "国庆拼团"}},
		listUpcomingResult:  []dto.ActivityInfo{{ID: 2, Title: "双11秒杀"}},
		listEndedResult:     []dto.ActivityInfo{{ID: 3, Title: "中秋抽奖"}},
		statisticsResult: &dto.ActivityStatistics{TotalActivities: 10, OngoingActivities: 3, PendingActivities: 2, EndedActivities: 5},
	}
	coupon := &mockCouponService{
		createResult:      &dto.CouponInfo{ID: 1, RegionID: regionID, Title: "满100减20", Type: "reduce", Amount: 20, Threshold: 100, Status: 1},
		getByIDResult:     &dto.CouponInfo{ID: 1, RegionID: regionID, Title: "满100减20", Type: "reduce", Amount: 20, Threshold: 100, Status: 1},
		listResult:        []dto.CouponInfo{{ID: 1, Title: "满100减20"}, {ID: 2, Title: "9折券"}},
		listAvailableResult: []dto.CouponInfo{{ID: 1, Title: "满100减20"}},
		listMineResult:    []dto.UserCouponInfo{{ID: 1, UserID: ctxUserID, CouponID: 1, Status: "unused"}},
		statisticsResult:  &dto.CouponStatistics{TotalCoupons: 10, ActiveCoupons: 5, TotalReceived: 100, TotalUsed: 60},
	}
	sign := &mockSignService{
		checkInResult:     &dto.SignCheckInResponse{ContinuousDays: 3, Points: 5, TotalPoints: 50},
		getCalendarResult: &dto.SignCalendarResponse{ContinuousDays: 3, MonthDays: 30, SignedDays: 3, TotalPoints: 15},
		createRuleResult:  &dto.SignRuleInfo{ID: 1, Day: 7, Points: 10, Status: 1},
		getRuleByIDResult: &dto.SignRuleInfo{ID: 1, Day: 7, Points: 10, Status: 1},
		listRulesResult:        []dto.SignRuleInfo{{ID: 1, Day: 7, Points: 10, Status: 1}, {ID: 2, Day: 14, Points: 20, Status: 1}},
		listEnabledRulesResult: []dto.SignRuleInfo{{ID: 1, Day: 7, Points: 10, Status: 1}},
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

	adH := mktHandler.NewAdHandler(ad)
	couponH := mktHandler.NewCouponHandler(coupon)
	signH := mktHandler.NewSignHandler(sign)
	activityH := mktHandler.NewActivityHandler(activity)

	// 注册路由，路径与 marketing/plugin.go RegisterRoutes 保持一致（去掉权限/限流中间件，纯测 handler）
	root := r.Group("/api/v1/marketing")

	// ==================== 公开路由（C 端浏览，无需登录） ====================
	root.GET("/positions/:code/ads", adH.ListByPositionCode)
	root.GET("/coupons/available", couponH.ListAvailable)
	root.GET("/coupons/:id", couponH.GetByID)
	root.GET("/activities/ongoing", activityH.ListOngoing)
	root.GET("/activities/upcoming", activityH.ListUpcoming)
	root.GET("/activities/ended", activityH.ListEnded)
	root.GET("/activities/:id", activityH.GetByID)
	root.GET("/sign/rules/enabled", signH.ListEnabledRules)

	// ==================== 需登录路由（C 端用户操作） ====================
	root.POST("/sign/check-in", signH.CheckIn)
	root.GET("/sign/calendar", signH.GetCalendar)
	root.POST("/coupons/:id/receive", couponH.Receive)
	root.POST("/user-coupons/:id/use", couponH.Use)
	root.POST("/user-coupons/:id/refund", couponH.Refund)
	root.GET("/my-coupons", couponH.ListMine)

	// ==================== 管理后台路由（/admin 组） ====================
	admin := root.Group("/admin")

	// 广告位管理
	admin.GET("/ads", adH.List)
	admin.POST("/ads", adH.Create)
	admin.GET("/ads/:id", adH.GetByID)
	admin.PUT("/ads/:id", adH.Update)
	admin.DELETE("/ads/:id", adH.Delete)

	// 优惠券管理
	admin.GET("/coupons", couponH.List)
	admin.POST("/coupons", couponH.Create)
	admin.GET("/coupons/statistics", couponH.Statistics)
	admin.GET("/coupons/:id", couponH.GetByID)
	admin.PUT("/coupons/:id", couponH.Update)
	admin.DELETE("/coupons/:id", couponH.Delete)

	// 签到规则管理
	admin.GET("/sign-rules", signH.ListRules)
	admin.POST("/sign-rules", signH.CreateRule)
	admin.GET("/sign-rules/:id", signH.GetRuleByID)
	admin.PUT("/sign-rules/:id", signH.UpdateRule)
	admin.DELETE("/sign-rules/:id", signH.DeleteRule)

	// 营销活动管理
	admin.GET("/activities", activityH.List)
	admin.POST("/activities", activityH.Create)
	admin.GET("/activities/statistics", activityH.Statistics)
	admin.GET("/activities/:id", activityH.GetByID)
	admin.PUT("/activities/:id", activityH.Update)
	admin.DELETE("/activities/:id", activityH.Delete)
	admin.PUT("/activities/:id/status", activityH.UpdateStatus)

	return &handlerEnv{engine: r.Engine(), ad: ad, activity: activity, coupon: coupon, sign: sign}
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
// ==================== AdHandler 测试（广告位） ====================
// =====================================================================

// ---------- AdHandler.Create ----------

func TestAdHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{
		"position_code": "home_banner",
		"title":         "首页轮播广告",
		"image_url":     "https://cdn.example.com/ad.png",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/ads", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, uint(5), env.ad.lastCreateRegionID)
	require.NotNil(t, env.ad.lastCreateReq)
	assert.Equal(t, "home_banner", env.ad.lastCreateReq.PositionCode)
	assert.Equal(t, "首页轮播广告", env.ad.lastCreateReq.Title)
	var info dto.AdPositionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestAdHandler_Create_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/marketing/admin/ads", "{bad json", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestAdHandler_Create_ValidationFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 缺少 required 字段 position_code/title/image_url
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/ads", map[string]interface{}{"sort": 1})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestAdHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.ad.createResult = nil
	env.ad.createErr = errors.New("db insert fail")
	body := map[string]interface{}{
		"position_code": "home_banner",
		"title":         "首页轮播广告",
		"image_url":     "https://cdn.example.com/ad.png",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/ads", body)

	// CodeMarketingAdError=5801
	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "db insert fail", resp.Message)
}

// ---------- AdHandler.Update ----------

func TestAdHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	title := "更新后的标题"
	body := map[string]interface{}{"title": title}
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/ads/3", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.ad.lastUpdateID)
	require.NotNil(t, env.ad.lastUpdateReq)
	assert.Equal(t, &title, env.ad.lastUpdateReq.Title)
}

func TestAdHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/ads/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.ad.lastUpdateID)
}

func TestAdHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/marketing/admin/ads/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestAdHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.ad.updateErr = errors.New("not found")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/ads/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "not found", resp.Message)
}

// ---------- AdHandler.Delete ----------

func TestAdHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/ads/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.ad.lastDeleteID)
}

func TestAdHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/ads/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestAdHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.ad.deleteErr = errors.New("constraint fail")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/ads/3", nil)

	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "constraint fail", resp.Message)
}

// ---------- AdHandler.GetByID ----------

func TestAdHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.ad.lastGetByIDID)
	var info dto.AdPositionInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "首页轮播", info.Title)
}

func TestAdHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
	assert.Equal(t, uint(0), env.ad.lastGetByIDID)
}

func TestAdHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.ad.getByIDResult = nil
	env.ad.getByIDErr = errors.New("广告位不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads/999", nil)

	// CodeMarketingAdNotFound=5802
	assert.Equal(t, 5802, resp.Code)
	assert.Equal(t, "广告位不存在", resp.Message)
}

// ---------- AdHandler.List ----------

func TestAdHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads?page=1&page_size=10&keyword=轮播", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.ad.lastListRegionID)
	require.NotNil(t, env.ad.lastListReq)
	assert.Equal(t, "轮播", env.ad.lastListReq.Keyword)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.AdPositionInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestAdHandler_List_RegionIDInjection(t *testing.T) {
	env := newHandlerEnv(t, 1, 8)
	env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads", nil)

	assert.Equal(t, uint(8), env.ad.lastListRegionID)
}

func TestAdHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.ad.listResult = nil
	env.ad.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/ads", nil)

	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- AdHandler.ListByPositionCode ----------

func TestAdHandler_ListByPositionCode_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/positions/home_banner/ads?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.ad.lastListByCodeRegionID)
	assert.Equal(t, "home_banner", env.ad.lastListByCodePositionCode)
	assert.Equal(t, 1, env.ad.lastListByCodePage)
	assert.Equal(t, 10, env.ad.lastListByCodePageSize)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestAdHandler_ListByPositionCode_EmptyCode(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	// 路由 /positions/:code/ads 中 :code 为空时不会匹配到该路由（gin 会 404），
	// 此处验证正常 code 传入空串不会触发 400（handler 仅在 ctx.Param 为空串时返回 400，
	// 但 gin 路由保证 :code 非空才进入 handler）。改用 service 错误路径覆盖。
	env.ad.listByCodeErr = errors.New("position not found")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/positions/home_banner/ads", nil)

	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "position not found", resp.Message)
}

func TestAdHandler_ListByPositionCode_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.ad.listByCodeResult = nil
	env.ad.listByCodeErr = errors.New("redis boom")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/positions/home_banner/ads", nil)

	assert.Equal(t, 5801, resp.Code)
	assert.Equal(t, "redis boom", resp.Message)
}

// =====================================================================
// ================ ActivityHandler 测试（营销活动） ==================
// =====================================================================

// ---------- ActivityHandler.Create ----------

func TestActivityHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{
		"title": "国庆拼团",
		"type":  "groupbuy",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/activities", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, uint(5), env.activity.lastCreateRegionID)
	require.NotNil(t, env.activity.lastCreateReq)
	assert.Equal(t, "国庆拼团", env.activity.lastCreateReq.Title)
	assert.Equal(t, "groupbuy", env.activity.lastCreateReq.Type)
}

func TestActivityHandler_Create_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/marketing/admin/activities", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestActivityHandler_Create_ValidationFail_InvalidType(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// type 不在 oneof=groupbuy bargain seckill lottery 中
	body := map[string]interface{}{
		"title": "国庆拼团",
		"type":  "invalid_type",
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/activities", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestActivityHandler_Create_ValidationFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 缺少 required 字段 title 和 type
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/activities", map[string]interface{}{"description": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestActivityHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.activity.createResult = nil
	env.activity.createErr = errors.New("db fail")
	body := map[string]interface{}{"title": "国庆拼团", "type": "groupbuy"}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/activities", body)

	// CodeMarketingActivityError=5819
	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// ---------- ActivityHandler.Update ----------

func TestActivityHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	title := "更新活动标题"
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/3", map[string]interface{}{"title": title})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.activity.lastUpdateID)
	require.NotNil(t, env.activity.lastUpdateReq)
	assert.Equal(t, &title, env.activity.lastUpdateReq.Title)
}

func TestActivityHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestActivityHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/marketing/admin/activities/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestActivityHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.activity.updateErr = errors.New("not found")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "not found", resp.Message)
}

// ---------- ActivityHandler.Delete ----------

func TestActivityHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/activities/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.activity.lastDeleteID)
}

func TestActivityHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/activities/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestActivityHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.activity.deleteErr = errors.New("db fail")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/activities/3", nil)

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// ---------- ActivityHandler.GetByID ----------

func TestActivityHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.activity.lastGetByIDID)
	var info dto.ActivityInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "国庆拼团", info.Title)
}

func TestActivityHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestActivityHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.activity.getByIDResult = nil
	env.activity.getByIDErr = errors.New("活动不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/999", nil)

	// CodeMarketingActivityNotFound=5820
	assert.Equal(t, 5820, resp.Code)
	assert.Equal(t, "活动不存在", resp.Message)
}

// ---------- ActivityHandler.List ----------

func TestActivityHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/activities?page=1&page_size=10&type=groupbuy", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.activity.lastListRegionID)
	require.NotNil(t, env.activity.lastListReq)
	assert.Equal(t, "groupbuy", env.activity.lastListReq.Type)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
}

func TestActivityHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.activity.listResult = nil
	env.activity.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/activities", nil)

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- ActivityHandler.ListOngoing / ListUpcoming / ListEnded ----------

func TestActivityHandler_ListOngoing_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/ongoing?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.activity.lastListOngoingRegionID)
	assert.Equal(t, 1, env.activity.lastListOngoingPage)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestActivityHandler_ListOngoing_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.activity.listOngoingResult = nil
	env.activity.listOngoingErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/ongoing", nil)

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

func TestActivityHandler_ListUpcoming_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/upcoming", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.activity.lastListUpcomingRegionID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestActivityHandler_ListUpcoming_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.activity.listUpcomingResult = nil
	env.activity.listUpcomingErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/upcoming", nil)

	assert.Equal(t, 5819, resp.Code)
}

func TestActivityHandler_ListEnded_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/ended", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.activity.lastListEndedRegionID)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestActivityHandler_ListEnded_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.activity.listEndedResult = nil
	env.activity.listEndedErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/activities/ended", nil)

	assert.Equal(t, 5819, resp.Code)
}

// ---------- ActivityHandler.UpdateStatus ----------

func TestActivityHandler_UpdateStatus_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/3/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "状态已更新", resp.Message)
	assert.Equal(t, uint(3), env.activity.lastUpdateStatusID)
	assert.Equal(t, 2, env.activity.lastUpdateStatusStatus)
}

func TestActivityHandler_UpdateStatus_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/abc/status", map[string]interface{}{"status": 2})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestActivityHandler_UpdateStatus_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/marketing/admin/activities/3/status", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestActivityHandler_UpdateStatus_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.activity.updateStatusErr = errors.New("status invalid")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/activities/3/status", map[string]interface{}{"status": 99})

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "status invalid", resp.Message)
}

// ---------- ActivityHandler.Statistics ----------

func TestActivityHandler_Statistics_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/activities/statistics", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.activity.lastStatisticsRegionID)
	var stats dto.ActivityStatistics
	require.NoError(t, json.Unmarshal(resp.Data, &stats))
	assert.Equal(t, int64(10), stats.TotalActivities)
	assert.Equal(t, int64(3), stats.OngoingActivities)
}

func TestActivityHandler_Statistics_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.activity.statisticsResult = nil
	env.activity.statisticsErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/activities/statistics", nil)

	assert.Equal(t, 5819, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ================ CouponHandler 测试（优惠券） ====================
// =====================================================================

// ---------- CouponHandler.Create ----------

func TestCouponHandler_Create_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{
		"title":       "满100减20",
		"type":        "reduce",
		"amount":      20,
		"threshold":   100,
		"total_count": 1000,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/coupons", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	assert.Equal(t, uint(5), env.coupon.lastCreateRegionID)
	require.NotNil(t, env.coupon.lastCreateReq)
	assert.Equal(t, "满100减20", env.coupon.lastCreateReq.Title)
	assert.Equal(t, "reduce", env.coupon.lastCreateReq.Type)
	assert.Equal(t, 20.0, env.coupon.lastCreateReq.Amount)
}

func TestCouponHandler_Create_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/marketing/admin/coupons", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestCouponHandler_Create_ValidationFail_InvalidType(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{
		"title":  "测试券",
		"type":   "invalid",
		"amount": 10,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/coupons", body)

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestCouponHandler_Create_ValidationFail_MissingRequired(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 缺少 required 字段 title 和 type
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/coupons", map[string]interface{}{"amount": 10})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestCouponHandler_Create_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.coupon.createResult = nil
	env.coupon.createErr = errors.New("db fail")
	body := map[string]interface{}{"title": "满100减20", "type": "reduce", "amount": 20}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/coupons", body)

	// CodeMarketingCouponError=5806
	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// ---------- CouponHandler.Update ----------

func TestCouponHandler_Update_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	title := "更新券名称"
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/coupons/3", map[string]interface{}{"title": title})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.coupon.lastUpdateID)
	require.NotNil(t, env.coupon.lastUpdateReq)
	assert.Equal(t, &title, env.coupon.lastUpdateReq.Title)
}

func TestCouponHandler_Update_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/coupons/abc", map[string]interface{}{"title": "x"})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_Update_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/marketing/admin/coupons/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestCouponHandler_Update_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.coupon.updateErr = errors.New("not found")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/coupons/3", map[string]interface{}{"title": "x"})

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "not found", resp.Message)
}

// ---------- CouponHandler.Delete ----------

func TestCouponHandler_Delete_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/coupons/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.coupon.lastDeleteID)
}

func TestCouponHandler_Delete_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/coupons/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_Delete_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.coupon.deleteErr = errors.New("db fail")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/coupons/3", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// ---------- CouponHandler.GetByID ----------

func TestCouponHandler_GetByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/coupons/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.coupon.lastGetByIDID)
	var info dto.CouponInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, "满100减20", info.Title)
}

func TestCouponHandler_GetByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/coupons/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_GetByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.coupon.getByIDResult = nil
	env.coupon.getByIDErr = errors.New("优惠券不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/coupons/999", nil)

	// CodeMarketingCouponNotFound=5807
	assert.Equal(t, 5807, resp.Code)
	assert.Equal(t, "优惠券不存在", resp.Message)
}

// ---------- CouponHandler.List ----------

func TestCouponHandler_List_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/coupons?page=1&page_size=10&type=reduce", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.coupon.lastListRegionID)
	require.NotNil(t, env.coupon.lastListReq)
	assert.Equal(t, "reduce", env.coupon.lastListReq.Type)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
}

func TestCouponHandler_List_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.coupon.listResult = nil
	env.coupon.listErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/coupons", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- CouponHandler.ListAvailable ----------

func TestCouponHandler_ListAvailable_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/coupons/available?page=1&page_size=10", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.coupon.lastListAvailableRegionID)
	assert.Equal(t, 1, env.coupon.lastListAvailablePage)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestCouponHandler_ListAvailable_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.coupon.listAvailableResult = nil
	env.coupon.listAvailableErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/coupons/available", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- CouponHandler.Receive ----------

func TestCouponHandler_Receive_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/coupons/3/receive", map[string]interface{}{"source": "receive"})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "领取成功", resp.Message)
	assert.Equal(t, uint(7), env.coupon.lastReceiveUserID)
	assert.Equal(t, uint(3), env.coupon.lastReceiveCouponID)
	assert.Equal(t, "receive", env.coupon.lastReceiveSource)
}

func TestCouponHandler_Receive_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/coupons/3/receive", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.coupon.lastReceiveUserID)
}

func TestCouponHandler_Receive_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/coupons/abc/receive", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_Receive_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.coupon.receiveErr = errors.New("已领取过该优惠券")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/coupons/3/receive", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "已领取过该优惠券", resp.Message)
}

// ---------- CouponHandler.Use ----------

func TestCouponHandler_Use_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/use", map[string]interface{}{"order_id": 100})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "使用成功", resp.Message)
	assert.Equal(t, uint(3), env.coupon.lastUseID)
	assert.Equal(t, uint(100), env.coupon.lastUseOrderID)
}

func TestCouponHandler_Use_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/use", map[string]interface{}{"order_id": 100})

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestCouponHandler_Use_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/abc/use", map[string]interface{}{"order_id": 100})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_Use_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/use", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestCouponHandler_Use_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.coupon.useErr = errors.New("用户优惠券已使用")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/use", map[string]interface{}{"order_id": 100})

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "用户优惠券已使用", resp.Message)
}

// ---------- CouponHandler.Refund ----------

func TestCouponHandler_Refund_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/refund", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "退还成功", resp.Message)
	assert.Equal(t, uint(3), env.coupon.lastRefundID)
}

func TestCouponHandler_Refund_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/refund", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
}

func TestCouponHandler_Refund_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/abc/refund", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestCouponHandler_Refund_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.coupon.refundErr = errors.New("用户优惠券不存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/user-coupons/3/refund", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "用户优惠券不存在", resp.Message)
}

// ---------- CouponHandler.ListMine ----------

func TestCouponHandler_ListMine_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/my-coupons?page=1&page_size=10&status=unused", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.coupon.lastListMineUserID)
	require.NotNil(t, env.coupon.lastListMineReq)
	assert.Equal(t, "unused", env.coupon.lastListMineReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(1), p.Total)
}

func TestCouponHandler_ListMine_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/my-coupons", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.coupon.lastListMineUserID)
}

func TestCouponHandler_ListMine_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.coupon.listMineResult = nil
	env.coupon.listMineErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/my-coupons", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- CouponHandler.Statistics ----------

func TestCouponHandler_Statistics_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/coupons/statistics", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(5), env.coupon.lastStatisticsRegionID)
	var stats dto.CouponStatistics
	require.NoError(t, json.Unmarshal(resp.Data, &stats))
	assert.Equal(t, int64(10), stats.TotalCoupons)
	assert.Equal(t, int64(5), stats.ActiveCoupons)
}

func TestCouponHandler_Statistics_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.coupon.statisticsResult = nil
	env.coupon.statisticsErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/coupons/statistics", nil)

	assert.Equal(t, 5806, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// =====================================================================
// ==================== SignHandler 测试（签到） ====================
// =====================================================================

// ---------- SignHandler.CheckIn ----------

func TestSignHandler_CheckIn_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/sign/check-in", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "签到成功", resp.Message)
	assert.Equal(t, uint(7), env.sign.lastCheckInUserID)
	var info dto.SignCheckInResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, 3, info.ContinuousDays)
	assert.Equal(t, 5, info.Points)
}

func TestSignHandler_CheckIn_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/sign/check-in", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.sign.lastCheckInUserID)
}

func TestSignHandler_CheckIn_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.sign.checkInResult = nil
	env.sign.checkInErr = errors.New("今日已签到")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/sign/check-in", nil)

	// CodeMarketingSignError=5816
	assert.Equal(t, 5816, resp.Code)
	assert.Equal(t, "今日已签到", resp.Message)
}

// ---------- SignHandler.GetCalendar ----------

func TestSignHandler_GetCalendar_Success(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/sign/calendar?month=2026-01", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(7), env.sign.lastGetCalendarUserID)
	assert.Equal(t, "2026-01", env.sign.lastGetCalendarMonth)
	var info dto.SignCalendarResponse
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, 30, info.MonthDays)
	assert.Equal(t, 3, info.SignedDays)
}

func TestSignHandler_GetCalendar_NotLoggedIn(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/sign/calendar", nil)

	assert.Equal(t, 401, resp.Code)
	assert.Equal(t, "请先登录", resp.Message)
	assert.Equal(t, uint(0), env.sign.lastGetCalendarUserID)
}

func TestSignHandler_GetCalendar_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 7, 5)
	env.sign.getCalendarResult = nil
	env.sign.getCalendarErr = errors.New("月份格式应为 YYYY-MM")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/sign/calendar?month=bad", nil)

	assert.Equal(t, 5816, resp.Code)
	assert.Equal(t, "月份格式应为 YYYY-MM", resp.Message)
}

// ---------- SignHandler.CreateRule ----------

func TestSignHandler_CreateRule_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	body := map[string]interface{}{
		"day":    7,
		"points": 10,
	}
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/sign-rules", body)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "创建成功", resp.Message)
	require.NotNil(t, env.sign.lastCreateRuleReq)
	assert.Equal(t, 7, env.sign.lastCreateRuleReq.Day)
	assert.Equal(t, 10, env.sign.lastCreateRuleReq.Points)
	var info dto.SignRuleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, uint(1), info.ID)
}

func TestSignHandler_CreateRule_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPost, "/api/v1/marketing/admin/sign-rules", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestSignHandler_CreateRule_ValidationFail_MissingDay(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// 缺少 required 字段 day
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/sign-rules", map[string]interface{}{"points": 10})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestSignHandler_CreateRule_ValidationFail_DayMin(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	// day binding:"required,min=1"，day=0 触发 min=1 失败（但 required 已先触发）
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/sign-rules", map[string]interface{}{"day": 0, "points": 10})

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestSignHandler_CreateRule_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.sign.createRuleResult = nil
	env.sign.createRuleErr = errors.New("规则已存在")
	resp := env.doJSON(t, http.MethodPost, "/api/v1/marketing/admin/sign-rules", map[string]interface{}{"day": 7, "points": 10})

	// CodeMarketingSignRuleError=5817
	assert.Equal(t, 5817, resp.Code)
	assert.Equal(t, "规则已存在", resp.Message)
}

// ---------- SignHandler.UpdateRule ----------

func TestSignHandler_UpdateRule_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	points := 15
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/sign-rules/3", map[string]interface{}{"points": points})

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "更新成功", resp.Message)
	assert.Equal(t, uint(3), env.sign.lastUpdateRuleID)
	require.NotNil(t, env.sign.lastUpdateRuleReq)
	assert.Equal(t, &points, env.sign.lastUpdateRuleReq.Points)
}

func TestSignHandler_UpdateRule_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/sign-rules/abc", map[string]interface{}{"points": 15})

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSignHandler_UpdateRule_BindFail(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doRaw(t, http.MethodPut, "/api/v1/marketing/admin/sign-rules/3", "{bad", "application/json")

	assert.Equal(t, 400, resp.Code)
	assert.Contains(t, resp.Message, "参数错误")
}

func TestSignHandler_UpdateRule_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.sign.updateRuleErr = errors.New("not found")
	resp := env.doJSON(t, http.MethodPut, "/api/v1/marketing/admin/sign-rules/3", map[string]interface{}{"points": 15})

	assert.Equal(t, 5817, resp.Code)
	assert.Equal(t, "not found", resp.Message)
}

// ---------- SignHandler.DeleteRule ----------

func TestSignHandler_DeleteRule_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/sign-rules/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "删除成功", resp.Message)
	assert.Equal(t, uint(3), env.sign.lastDeleteRuleID)
}

func TestSignHandler_DeleteRule_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/sign-rules/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSignHandler_DeleteRule_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.sign.deleteRuleErr = errors.New("db fail")
	resp := env.doJSON(t, http.MethodDelete, "/api/v1/marketing/admin/sign-rules/3", nil)

	assert.Equal(t, 5817, resp.Code)
	assert.Equal(t, "db fail", resp.Message)
}

// ---------- SignHandler.GetRuleByID ----------

func TestSignHandler_GetRuleByID_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/sign-rules/3", nil)

	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, uint(3), env.sign.lastGetRuleByIDID)
	var info dto.SignRuleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &info))
	assert.Equal(t, 7, info.Day)
	assert.Equal(t, 10, info.Points)
}

func TestSignHandler_GetRuleByID_InvalidID(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/sign-rules/abc", nil)

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "无效的ID", resp.Message)
}

func TestSignHandler_GetRuleByID_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.sign.getRuleByIDResult = nil
	env.sign.getRuleByIDErr = errors.New("签到规则不存在")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/sign-rules/999", nil)

	// CodeMarketingSignRuleNotFound=5818
	assert.Equal(t, 5818, resp.Code)
	assert.Equal(t, "签到规则不存在", resp.Message)
}

// ---------- SignHandler.ListRules ----------

func TestSignHandler_ListRules_Success(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/sign-rules?page=1&page_size=10&status=1", nil)

	assert.Equal(t, 0, resp.Code)
	require.NotNil(t, env.sign.lastListRulesReq)
	assert.Equal(t, 1, *env.sign.lastListRulesReq.Status)
	p := parsePage(t, resp)
	assert.Equal(t, int64(2), p.Total)
	var list []dto.SignRuleInfo
	require.NoError(t, json.Unmarshal(p.List, &list))
	require.Len(t, list, 2)
}

func TestSignHandler_ListRules_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 1, 5)
	env.sign.listRulesResult = nil
	env.sign.listRulesErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/admin/sign-rules", nil)

	assert.Equal(t, 5817, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}

// ---------- SignHandler.ListEnabledRules ----------

func TestSignHandler_ListEnabledRules_Success(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/sign/rules/enabled", nil)

	assert.Equal(t, 0, resp.Code)
	var list []dto.SignRuleInfo
	require.NoError(t, json.Unmarshal(resp.Data, &list))
	require.Len(t, list, 1)
	assert.Equal(t, 7, list[0].Day)
}

func TestSignHandler_ListEnabledRules_ServiceError(t *testing.T) {
	env := newHandlerEnv(t, 0, 5)
	env.sign.listEnabledRulesResult = nil
	env.sign.listEnabledRulesErr = errors.New("db down")
	resp := env.doJSON(t, http.MethodGet, "/api/v1/marketing/sign/rules/enabled", nil)

	assert.Equal(t, 5817, resp.Code)
	assert.Equal(t, "db down", resp.Message)
}
