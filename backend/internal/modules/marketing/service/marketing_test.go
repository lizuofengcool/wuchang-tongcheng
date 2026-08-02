// Package service 营销活动中台业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖四个子域核心逻辑，不依赖 DB：
//   - ad 子域：广告位 CRUD、默认状态、字段映射、空字段、未找到错误、列表与按编码查询
//   - activity 子域：活动 CRUD、默认状态、配置解析、状态更新、自动推进、统计聚合
//   - coupon 子域：优惠券 CRUD、领取校验（禁用/未开始/过期/抢完/重复领取）、
//     使用/退还状态校验、我的优惠券冗余填充、统计与过期检查
//   - sign 子域：每日签到（重复签到/连续天数计算/规则命中/规则禁用）、
//     签到日历（月份格式校验/默认当月/连续天数）、签到规则 CRUD 与唯一性校验
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 测试辅助函数 =====

func timePtr(t time.Time) *time.Time { return &t }

func intPtr(i int) *int { return &i }

func floatPtr(f float64) *float64 { return &f }

func strPtr(s string) *string { return &s }

// 构造带 ID/RegionID 的广告位（嵌入式字段通过赋值设置）
func newAd(id, regionID uint, a model.AdPosition) *model.AdPosition {
	a.ID = id
	a.RegionID = regionID
	return &a
}

func newActivity(id, regionID uint, a model.Activity) *model.Activity {
	a.ID = id
	a.RegionID = regionID
	return &a
}

func newCoupon(id, regionID uint, c model.Coupon) *model.Coupon {
	c.ID = id
	c.RegionID = regionID
	return &c
}

func newUC(id uint, uc model.UserCoupon) *model.UserCoupon {
	uc.ID = id
	return &uc
}

func newRecord(id uint, r model.SignRecord) model.SignRecord {
	r.ID = id
	return r
}

func newRule(id uint, r model.SignRule) *model.SignRule {
	r.ID = id
	return &r
}

// paginateAd 对内存列表按分页切片
func paginateAd(list []model.AdPosition, p *utils.Pagination) []model.AdPosition {
	if p == nil || p.PageSize <= 0 {
		return list
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * p.PageSize
	if start >= len(list) {
		return nil
	}
	end := start + p.PageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// =====================================================================
// ===== ad 子域 mock =====
// =====================================================================

type mockAdRepo struct {
	byID        map[uint]*model.AdPosition
	nextID      uint
	createErr   error
	findErr     error
	updateErr   error
	deleteErr   error
	listErr     error
	byCodeErr   error
	updated     map[uint]map[string]interface{}
	deletedIDs  map[uint]bool
	byCodeCalls int
}

func newMockAdRepo() *mockAdRepo {
	return &mockAdRepo{
		byID:       make(map[uint]*model.AdPosition),
		nextID:     1,
		updated:    make(map[uint]map[string]interface{}),
		deletedIDs: make(map[uint]bool),
	}
}

func (m *mockAdRepo) Create(a *model.AdPosition) error {
	if m.createErr != nil {
		return m.createErr
	}
	a.ID = m.nextID
	m.nextID++
	cp := *a
	m.byID[a.ID] = &cp
	return nil
}

func (m *mockAdRepo) FindByID(id uint) (*model.AdPosition, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	a, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *mockAdRepo) Update(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated[id] = fields
	return nil
}

func (m *mockAdRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.byID, id)
	m.deletedIDs[id] = true
	return nil
}

func (m *mockAdRepo) List(regionID uint, query repository.AdPositionListQuery, p *utils.Pagination) ([]model.AdPosition, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var filtered []model.AdPosition
	for _, a := range m.byID {
		if regionID > 0 && a.RegionID != regionID {
			continue
		}
		if query.PositionCode != "" && a.PositionCode != query.PositionCode {
			continue
		}
		if query.Status != nil && a.Status != *query.Status {
			continue
		}
		filtered = append(filtered, *a)
	}
	return paginateAd(filtered, p), int64(len(filtered)), nil
}

func (m *mockAdRepo) FindByPositionCode(regionID uint, positionCode string, p *utils.Pagination) ([]model.AdPosition, int64, error) {
	if m.byCodeErr != nil {
		return nil, 0, m.byCodeErr
	}
	m.byCodeCalls++
	return m.List(regionID, repository.AdPositionListQuery{PositionCode: positionCode}, p)
}

// ===== ad 子域测试 =====

func TestAdService_Create_DefaultStatus(t *testing.T) {
	repo := newMockAdRepo()
	svc := NewAdService(repo)

	info, err := svc.Create(5, &dto.CreateAdPositionRequest{
		PositionCode: model.AdPositionHomeBanner,
		Title:        "首页Banner",
		ImageURL:     "https://cdn/a.png",
		LinkURL:      "https://example.com",
		Sort:         10,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, uint(5), info.RegionID)
	assert.Equal(t, model.AdPositionHomeBanner, info.PositionCode)
	assert.Equal(t, "首页Banner", info.Title)
	assert.Equal(t, 10, info.Sort)
	// status 未传 → 默认启用
	assert.Equal(t, model.AdStatusEnabled, info.Status)
	assert.Equal(t, "启用", info.StatusText)
}

func TestAdService_Create_ExplicitStatus(t *testing.T) {
	repo := newMockAdRepo()
	svc := NewAdService(repo)

	info, err := svc.Create(1, &dto.CreateAdPositionRequest{
		PositionCode: "popup",
		Title:        "弹窗",
		ImageURL:     "u",
		Status:       model.AdStatusScheduled,
	})
	require.NoError(t, err)
	assert.Equal(t, model.AdStatusScheduled, info.Status)
	assert.Equal(t, "待生效", info.StatusText)
}

func TestAdService_Create_RepoError(t *testing.T) {
	repo := newMockAdRepo()
	repo.createErr = errors.New("db down")
	svc := NewAdService(repo)

	_, err := svc.Create(1, &dto.CreateAdPositionRequest{PositionCode: "c", Title: "t", ImageURL: "i"})
	assert.Equal(t, "db down", err.Error())
}

func TestAdService_GetByID(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[7] = newAd(7, 2, model.AdPosition{PositionCode: "list_top", Title: "T", Status: model.AdStatusExpired})
	svc := NewAdService(repo)

	info, err := svc.GetByID(7)
	require.NoError(t, err)
	assert.Equal(t, "list_top", info.PositionCode)
	assert.Equal(t, "已过期", info.StatusText)
}

func TestAdService_GetByID_NotFound(t *testing.T) {
	repo := newMockAdRepo()
	svc := NewAdService(repo)

	_, err := svc.GetByID(99)
	assert.ErrorIs(t, err, ErrAdNotFound)
}

func TestAdService_GetByID_RepoError(t *testing.T) {
	repo := newMockAdRepo()
	repo.findErr = errors.New("boom")
	svc := NewAdService(repo)

	_, err := svc.GetByID(1)
	assert.Equal(t, "boom", err.Error())
}

func TestAdService_Update(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[3] = newAd(3, 1, model.AdPosition{PositionCode: "c", Title: "old", Status: model.AdStatusEnabled})
	svc := NewAdService(repo)

	err := svc.Update(3, &dto.UpdateAdPositionRequest{
		Title:  strPtr("new"),
		Sort:   intPtr(5),
		Status: intPtr(model.AdStatusDisabled),
	})
	require.NoError(t, err)
	assert.Equal(t, "new", repo.updated[3]["title"])
	assert.Equal(t, 5, repo.updated[3]["sort"])
	assert.Equal(t, model.AdStatusDisabled, repo.updated[3]["status"])
}

func TestAdService_Update_NotFound(t *testing.T) {
	repo := newMockAdRepo()
	svc := NewAdService(repo)

	err := svc.Update(404, &dto.UpdateAdPositionRequest{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrAdNotFound)
}

func TestAdService_Update_EmptyFields(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[1] = newAd(1, 1, model.AdPosition{})
	svc := NewAdService(repo)

	err := svc.Update(1, &dto.UpdateAdPositionRequest{})
	require.NoError(t, err)
	_, ok := repo.updated[1]
	assert.False(t, ok, "空字段不应调用 Update")
}

func TestAdService_Delete(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[2] = newAd(2, 1, model.AdPosition{})
	svc := NewAdService(repo)

	require.NoError(t, svc.Delete(2))
	assert.True(t, repo.deletedIDs[2])
}

func TestAdService_Delete_NotFound(t *testing.T) {
	repo := newMockAdRepo()
	svc := NewAdService(repo)

	assert.ErrorIs(t, svc.Delete(9), ErrAdNotFound)
}

func TestAdService_List(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[1] = newAd(1, 5, model.AdPosition{PositionCode: "home_banner", Title: "A", Status: model.AdStatusEnabled})
	repo.byID[2] = newAd(2, 5, model.AdPosition{PositionCode: "popup", Title: "B", Status: model.AdStatusDisabled})
	repo.byID[3] = newAd(3, 5, model.AdPosition{PositionCode: "home_banner", Title: "C", Status: model.AdStatusEnabled})
	svc := NewAdService(repo)

	st := model.AdStatusEnabled
	p, list, err := svc.List(5, &dto.AdPositionListRequest{Status: &st})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	assert.Len(t, list, 2)
}

func TestAdService_List_RepoError(t *testing.T) {
	repo := newMockAdRepo()
	repo.listErr = errors.New("list fail")
	svc := NewAdService(repo)

	_, _, err := svc.List(1, &dto.AdPositionListRequest{})
	assert.Equal(t, "list fail", err.Error())
}

func TestAdService_ListByPositionCode(t *testing.T) {
	repo := newMockAdRepo()
	repo.byID[1] = newAd(1, 1, model.AdPosition{PositionCode: "home_banner", Title: "A", Status: model.AdStatusEnabled})
	svc := NewAdService(repo)

	p, list, err := svc.ListByPositionCode(1, "home_banner", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	assert.Len(t, list, 1)
	assert.Equal(t, 1, repo.byCodeCalls)
}

// =====================================================================
// ===== activity 子域 mock =====
// =====================================================================

type mockActivityRepo struct {
	byID         map[uint]*model.Activity
	nextID       uint
	createErr    error
	findErr      error
	updateErr    error
	deleteErr    error
	listErr      error
	ongoingErr   error
	upcomingErr  error
	endedErr     error
	updated      map[uint]map[string]interface{}
	deletedIDs   map[uint]bool
	autoAffected int64
	autoErr      error
	ongoingList  []model.Activity
	upcomingList []model.Activity
	endedList    []model.Activity
}

func newMockActivityRepo() *mockActivityRepo {
	return &mockActivityRepo{
		byID:       make(map[uint]*model.Activity),
		nextID:     1,
		updated:    make(map[uint]map[string]interface{}),
		deletedIDs: make(map[uint]bool),
	}
}

func (m *mockActivityRepo) Create(a *model.Activity) error {
	if m.createErr != nil {
		return m.createErr
	}
	a.ID = m.nextID
	m.nextID++
	cp := *a
	m.byID[a.ID] = &cp
	return nil
}

func (m *mockActivityRepo) FindByID(id uint) (*model.Activity, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	a, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *mockActivityRepo) Update(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updated[id] = fields
	return nil
}

func (m *mockActivityRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.byID, id)
	m.deletedIDs[id] = true
	return nil
}

func (m *mockActivityRepo) List(regionID uint, query repository.ActivityListQuery, p *utils.Pagination) ([]model.Activity, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var filtered []model.Activity
	for _, a := range m.byID {
		if regionID > 0 && a.RegionID != regionID {
			continue
		}
		if query.Type != "" && a.Type != query.Type {
			continue
		}
		if query.Status != nil && a.Status != *query.Status {
			continue
		}
		filtered = append(filtered, *a)
	}
	if p == nil || p.PageSize <= 0 {
		return filtered, int64(len(filtered)), nil
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * p.PageSize
	if start >= len(filtered) {
		return nil, int64(len(filtered)), nil
	}
	end := start + p.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], int64(len(filtered)), nil
}

func (m *mockActivityRepo) ListOngoing(regionID uint, p *utils.Pagination) ([]model.Activity, int64, error) {
	if m.ongoingErr != nil {
		return nil, 0, m.ongoingErr
	}
	return m.ongoingList, int64(len(m.ongoingList)), nil
}

func (m *mockActivityRepo) ListUpcoming(regionID uint, p *utils.Pagination) ([]model.Activity, int64, error) {
	if m.upcomingErr != nil {
		return nil, 0, m.upcomingErr
	}
	return m.upcomingList, int64(len(m.upcomingList)), nil
}

func (m *mockActivityRepo) ListEnded(regionID uint, p *utils.Pagination) ([]model.Activity, int64, error) {
	if m.endedErr != nil {
		return nil, 0, m.endedErr
	}
	return m.endedList, int64(len(m.endedList)), nil
}

func (m *mockActivityRepo) UpdateStatusByTime(now time.Time) (int64, error) {
	return m.autoAffected, m.autoErr
}

// ===== activity 子域测试 =====

func TestActivityService_Create_DefaultStatus(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	start := time.Now().Add(time.Hour)
	end := start.Add(24 * time.Hour)
	info, err := svc.Create(3, &dto.CreateActivityRequest{
		Title:   "国庆拼团",
		Type:    model.ActivityTypeGroupBuy,
		StartAt: &start,
		EndAt:   &end,
		Config:  map[string]interface{}{"people": 3},
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, uint(3), info.RegionID)
	assert.Equal(t, model.ActivityTypeGroupBuy, info.Type)
	assert.Equal(t, "拼团", info.TypeText)
	// status 未传 → 默认待开始
	assert.Equal(t, model.ActivityStatusPending, info.Status)
	assert.Equal(t, "待开始", info.StatusText)
	// config 经 FromJSON 序列化后再解析回来
	cfg, ok := info.Config.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), cfg["people"])
}

func TestActivityService_Create_ExplicitStatus(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	info, err := svc.Create(1, &dto.CreateActivityRequest{
		Title:  "秒杀",
		Type:   model.ActivityTypeSeckill,
		Status: model.ActivityStatusOngoing,
	})
	require.NoError(t, err)
	assert.Equal(t, model.ActivityStatusOngoing, info.Status)
	assert.Equal(t, "秒杀", info.TypeText)
}

func TestActivityService_Create_RepoError(t *testing.T) {
	repo := newMockActivityRepo()
	repo.createErr = errors.New("insert fail")
	svc := NewActivityService(repo)

	_, err := svc.Create(1, &dto.CreateActivityRequest{Title: "t", Type: model.ActivityTypeLottery})
	assert.Equal(t, "insert fail", err.Error())
}

func TestActivityService_GetByID_NotFound(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	_, err := svc.GetByID(1)
	assert.ErrorIs(t, err, ErrActivityNotFound)
}

func TestActivityService_Update(t *testing.T) {
	repo := newMockActivityRepo()
	repo.byID[1] = newActivity(1, 1, model.Activity{Title: "old", Type: model.ActivityTypeGroupBuy, Status: model.ActivityStatusPending})
	svc := NewActivityService(repo)

	err := svc.Update(1, &dto.UpdateActivityRequest{
		Title:  strPtr("new"),
		Status: intPtr(model.ActivityStatusOngoing),
	})
	require.NoError(t, err)
	assert.Equal(t, "new", repo.updated[1]["title"])
	assert.Equal(t, model.ActivityStatusOngoing, repo.updated[1]["status"])
}

func TestActivityService_Update_NotFound(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	err := svc.Update(404, &dto.UpdateActivityRequest{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrActivityNotFound)
}

func TestActivityService_Update_EmptyFields(t *testing.T) {
	repo := newMockActivityRepo()
	repo.byID[1] = newActivity(1, 1, model.Activity{})
	svc := NewActivityService(repo)

	require.NoError(t, svc.Update(1, &dto.UpdateActivityRequest{}))
	_, ok := repo.updated[1]
	assert.False(t, ok)
}

func TestActivityService_Delete_NotFound(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	assert.ErrorIs(t, svc.Delete(1), ErrActivityNotFound)
}

func TestActivityService_Delete_Success(t *testing.T) {
	repo := newMockActivityRepo()
	repo.byID[2] = newActivity(2, 1, model.Activity{})
	svc := NewActivityService(repo)

	require.NoError(t, svc.Delete(2))
	assert.True(t, repo.deletedIDs[2])
}

func TestActivityService_List(t *testing.T) {
	repo := newMockActivityRepo()
	repo.byID[1] = newActivity(1, 5, model.Activity{Title: "A", Type: model.ActivityTypeBargain, Status: model.ActivityStatusOngoing})
	repo.byID[2] = newActivity(2, 5, model.Activity{Title: "B", Type: model.ActivityTypeSeckill, Status: model.ActivityStatusPending})
	repo.byID[3] = newActivity(3, 5, model.Activity{Title: "C", Type: model.ActivityTypeBargain, Status: model.ActivityStatusOngoing})
	svc := NewActivityService(repo)

	p, list, err := svc.List(5, &dto.ActivityListRequest{Type: model.ActivityTypeBargain})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	assert.Len(t, list, 2)
	for _, a := range list {
		assert.Equal(t, model.ActivityTypeBargain, a.Type)
		assert.Equal(t, "砍价", a.TypeText)
	}
}

func TestActivityService_ListOngoing(t *testing.T) {
	repo := newMockActivityRepo()
	repo.ongoingList = []model.Activity{*newActivity(1, 1, model.Activity{Title: "O1"})}
	svc := NewActivityService(repo)

	p, list, err := svc.ListOngoing(1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	assert.Len(t, list, 1)
}

func TestActivityService_ListUpcoming_RepoError(t *testing.T) {
	repo := newMockActivityRepo()
	repo.upcomingErr = errors.New("err")
	svc := NewActivityService(repo)

	_, _, err := svc.ListUpcoming(1, 1, 10)
	assert.Equal(t, "err", err.Error())
}

func TestActivityService_ListEnded(t *testing.T) {
	repo := newMockActivityRepo()
	repo.endedList = []model.Activity{*newActivity(1, 1, model.Activity{Title: "E1"})}
	svc := NewActivityService(repo)

	_, list, err := svc.ListEnded(1, 1, 10)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestActivityService_UpdateStatus_NotFound(t *testing.T) {
	repo := newMockActivityRepo()
	svc := NewActivityService(repo)

	assert.ErrorIs(t, svc.UpdateStatus(1, model.ActivityStatusOngoing), ErrActivityNotFound)
}

func TestActivityService_UpdateStatus_Success(t *testing.T) {
	repo := newMockActivityRepo()
	repo.byID[1] = newActivity(1, 1, model.Activity{})
	svc := NewActivityService(repo)

	require.NoError(t, svc.UpdateStatus(1, model.ActivityStatusEnded))
	assert.Equal(t, model.ActivityStatusEnded, repo.updated[1]["status"])
}

func TestActivityService_AutoUpdateStatus(t *testing.T) {
	repo := newMockActivityRepo()
	repo.autoAffected = 5
	repo.autoErr = errors.New("auto err")
	svc := NewActivityService(repo)

	n, err := svc.AutoUpdateStatus()
	assert.Equal(t, int64(5), n)
	assert.Equal(t, "auto err", err.Error())
}

func TestActivityService_Statistics(t *testing.T) {
	repo := newMockActivityRepo()
	regionID := uint(5)
	repo.byID[1] = newActivity(1, regionID, model.Activity{Status: model.ActivityStatusOngoing})
	repo.byID[2] = newActivity(2, regionID, model.Activity{Status: model.ActivityStatusOngoing})
	repo.byID[3] = newActivity(3, regionID, model.Activity{Status: model.ActivityStatusPending})
	repo.byID[4] = newActivity(4, regionID, model.Activity{Status: model.ActivityStatusEnded})
	repo.byID[5] = newActivity(5, 99, model.Activity{Status: model.ActivityStatusOngoing})
	svc := NewActivityService(repo)

	stats, err := svc.Statistics(regionID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), stats.TotalActivities)
	assert.Equal(t, int64(2), stats.OngoingActivities)
	assert.Equal(t, int64(1), stats.PendingActivities)
	assert.Equal(t, int64(1), stats.EndedActivities)
}

func TestActivityService_Statistics_ListError(t *testing.T) {
	repo := newMockActivityRepo()
	repo.listErr = errors.New("list err")
	svc := NewActivityService(repo)

	_, err := svc.Statistics(1)
	assert.Equal(t, "list err", err.Error())
}

// =====================================================================
// ===== coupon 子域 mock =====
// =====================================================================

type mockCouponRepo struct {
	coupons           map[uint]*model.Coupon
	nextCouponID      uint
	userCoupons       map[uint]*model.UserCoupon
	nextUserCouponID  uint
	createErr         error
	findErr           error
	updateErr         error
	deleteErr         error
	listErr           error
	listAvailErr      error
	incrErr           error
	createUCErr       error
	findUCErr         error
	listUCErr         error
	updateUCErr       error
	expireUCErr       error
	expireUCAffected  int64
	updatedCoupons    map[uint]map[string]interface{}
	updatedUserCoupons map[uint]map[string]interface{}
	incrCalls         int
}

func newMockCouponRepo() *mockCouponRepo {
	return &mockCouponRepo{
		coupons:            make(map[uint]*model.Coupon),
		nextCouponID:       1,
		userCoupons:        make(map[uint]*model.UserCoupon),
		nextUserCouponID:   1,
		updatedCoupons:     make(map[uint]map[string]interface{}),
		updatedUserCoupons: make(map[uint]map[string]interface{}),
	}
}

func (m *mockCouponRepo) Create(c *model.Coupon) error {
	if m.createErr != nil {
		return m.createErr
	}
	c.ID = m.nextCouponID
	m.nextCouponID++
	cp := *c
	m.coupons[c.ID] = &cp
	return nil
}

func (m *mockCouponRepo) FindByID(id uint) (*model.Coupon, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	c, ok := m.coupons[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockCouponRepo) Update(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedCoupons[id] = fields
	return nil
}

func (m *mockCouponRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.coupons, id)
	return nil
}

func (m *mockCouponRepo) List(regionID uint, query repository.CouponListQuery, p *utils.Pagination) ([]model.Coupon, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var filtered []model.Coupon
	for _, c := range m.coupons {
		if regionID > 0 && c.RegionID != regionID {
			continue
		}
		if query.Type != "" && c.Type != query.Type {
			continue
		}
		if query.Status != nil && c.Status != *query.Status {
			continue
		}
		filtered = append(filtered, *c)
	}
	if p == nil || p.PageSize <= 0 {
		return filtered, int64(len(filtered)), nil
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * p.PageSize
	if start >= len(filtered) {
		return nil, int64(len(filtered)), nil
	}
	end := start + p.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], int64(len(filtered)), nil
}

func (m *mockCouponRepo) ListAvailable(regionID uint, p *utils.Pagination) ([]model.Coupon, int64, error) {
	if m.listAvailErr != nil {
		return nil, 0, m.listAvailErr
	}
	return m.List(regionID, repository.CouponListQuery{}, p)
}

func (m *mockCouponRepo) IncrReceivedCount(id uint) error {
	if m.incrErr != nil {
		return m.incrErr
	}
	m.incrCalls++
	if c, ok := m.coupons[id]; ok {
		c.ReceivedCount++
	}
	return nil
}

func (m *mockCouponRepo) CreateUserCoupon(uc *model.UserCoupon) error {
	if m.createUCErr != nil {
		return m.createUCErr
	}
	uc.ID = m.nextUserCouponID
	m.nextUserCouponID++
	cp := *uc
	m.userCoupons[uc.ID] = &cp
	return nil
}

func (m *mockCouponRepo) FindUserCoupon(userID uint, couponID uint) (*model.UserCoupon, error) {
	for _, uc := range m.userCoupons {
		if uc.UserID == userID && uc.CouponID == couponID {
			cp := *uc
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockCouponRepo) FindUserCouponByID(id uint) (*model.UserCoupon, error) {
	if m.findUCErr != nil {
		return nil, m.findUCErr
	}
	uc, ok := m.userCoupons[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *uc
	return &cp, nil
}

func (m *mockCouponRepo) ListUserCoupons(query repository.UserCouponListQuery, p *utils.Pagination) ([]model.UserCoupon, int64, error) {
	if m.listUCErr != nil {
		return nil, 0, m.listUCErr
	}
	var filtered []model.UserCoupon
	for _, uc := range m.userCoupons {
		if query.UserID > 0 && uc.UserID != query.UserID {
			continue
		}
		if query.CouponID > 0 && uc.CouponID != query.CouponID {
			continue
		}
		if query.Status != "" && uc.Status != query.Status {
			continue
		}
		filtered = append(filtered, *uc)
	}
	return filtered, int64(len(filtered)), nil
}

func (m *mockCouponRepo) UpdateUserCoupon(id uint, fields map[string]interface{}) error {
	if m.updateUCErr != nil {
		return m.updateUCErr
	}
	m.updatedUserCoupons[id] = fields
	return nil
}

func (m *mockCouponRepo) CountUserCoupon(userID uint, couponID uint) (int64, error) {
	var n int64
	for _, uc := range m.userCoupons {
		if uc.UserID == userID && uc.CouponID == couponID {
			n++
		}
	}
	return n, nil
}

func (m *mockCouponRepo) ExpireUserCoupons(now time.Time) (int64, error) {
	if m.expireUCErr != nil {
		return 0, m.expireUCErr
	}
	return m.expireUCAffected, nil
}

// ===== coupon 子域测试 =====

func TestCouponService_Create_DefaultStatus(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	info, err := svc.Create(2, &dto.CreateCouponRequest{
		Title:      "满100减20",
		Type:       model.CouponTypeReduce,
		Amount:     20,
		Threshold:  100,
		TotalCount: 1000,
	})
	require.NoError(t, err)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, uint(2), info.RegionID)
	assert.Equal(t, model.CouponTypeReduce, info.Type)
	assert.Equal(t, "满减券", info.TypeText)
	// status 未传 → 默认进行中
	assert.Equal(t, model.CouponStatusActive, info.Status)
	assert.Equal(t, "进行中", info.StatusText)
	assert.Equal(t, 1000, info.TotalCount)
}

func TestCouponService_Create_ExplicitStatus(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	info, err := svc.Create(1, &dto.CreateCouponRequest{
		Title:  "折扣券",
		Type:   model.CouponTypeDiscount,
		Amount: 0.85,
		Status: model.CouponStatusDraft,
	})
	require.NoError(t, err)
	assert.Equal(t, model.CouponStatusDraft, info.Status)
	assert.Equal(t, "草稿", info.StatusText)
	assert.Equal(t, "折扣券", info.TypeText)
}

func TestCouponService_GetByID_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	_, err := svc.GetByID(1)
	assert.ErrorIs(t, err, ErrCouponNotFound)
}

func TestCouponService_Update(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Title: "old", Status: model.CouponStatusActive})
	svc := NewCouponService(repo)

	err := svc.Update(1, &dto.UpdateCouponRequest{
		Title:      strPtr("new"),
		Amount:     floatPtr(30),
		TotalCount: intPtr(500),
		Status:     intPtr(model.CouponStatusOffline),
	})
	require.NoError(t, err)
	assert.Equal(t, "new", repo.updatedCoupons[1]["title"])
	assert.Equal(t, float64(30), repo.updatedCoupons[1]["amount"])
	assert.Equal(t, 500, repo.updatedCoupons[1]["total_count"])
	assert.Equal(t, model.CouponStatusOffline, repo.updatedCoupons[1]["status"])
}

func TestCouponService_Update_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Update(9, &dto.UpdateCouponRequest{Title: strPtr("x")}), ErrCouponNotFound)
}

func TestCouponService_Update_EmptyFields(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{})
	svc := NewCouponService(repo)

	require.NoError(t, svc.Update(1, &dto.UpdateCouponRequest{}))
	_, ok := repo.updatedCoupons[1]
	assert.False(t, ok)
}

func TestCouponService_Delete_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Delete(1), ErrCouponNotFound)
}

func TestCouponService_List(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 5, model.Coupon{Title: "A", Type: model.CouponTypeReduce, Status: model.CouponStatusActive})
	repo.coupons[2] = newCoupon(2, 5, model.Coupon{Title: "B", Type: model.CouponTypeDiscount, Status: model.CouponStatusActive})
	repo.coupons[3] = newCoupon(3, 5, model.Coupon{Title: "C", Type: model.CouponTypeReduce, Status: model.CouponStatusActive})
	svc := NewCouponService(repo)

	p, list, err := svc.List(5, &dto.CouponListRequest{Type: model.CouponTypeReduce})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	assert.Len(t, list, 2)
}

func TestCouponService_ListAvailable(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive})
	svc := NewCouponService(repo)

	p, list, err := svc.ListAvailable(1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	assert.Len(t, list, 1)
}

// --- Receive ---

func TestCouponService_Receive_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 99, ""), ErrCouponNotFound)
}

func TestCouponService_Receive_StatusInvalid(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusDisabled})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 1, ""), ErrCouponStatusInvalid)
}

func TestCouponService_Receive_NotStarted(t *testing.T) {
	repo := newMockCouponRepo()
	future := time.Now().Add(time.Hour)
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, StartAt: &future})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 1, ""), ErrCouponNotStarted)
}

func TestCouponService_Receive_Expired(t *testing.T) {
	repo := newMockCouponRepo()
	past := time.Now().Add(-time.Hour)
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, EndAt: &past})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 1, ""), ErrCouponExpired)
}

func TestCouponService_Receive_SoldOut(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, TotalCount: 10, ReceivedCount: 10})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 1, ""), ErrCouponSoldOut)
}

func TestCouponService_Receive_AlreadyReceived(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, TotalCount: 100})
	repo.userCoupons[1] = newUC(1, model.UserCoupon{UserID: 1, CouponID: 1, Status: model.UserCouponStatusUnused})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Receive(1, 1, ""), ErrCouponAlreadyRecv)
}

func TestCouponService_Receive_Success_DefaultSource(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, TotalCount: 100})
	svc := NewCouponService(repo)

	require.NoError(t, svc.Receive(1, 1, ""))
	// 默认来源应为 receive
	uc := repo.userCoupons[1]
	require.NotNil(t, uc)
	assert.Equal(t, uint(1), uc.UserID)
	assert.Equal(t, uint(1), uc.CouponID)
	assert.Equal(t, model.UserCouponSourceReceive, uc.Source)
	assert.Equal(t, model.UserCouponStatusUnused, uc.Status)
	assert.Equal(t, 1, repo.incrCalls)
}

func TestCouponService_Receive_Success_GivenSource(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, TotalCount: 0})
	svc := NewCouponService(repo)

	require.NoError(t, svc.Receive(2, 1, model.UserCouponSourceGift))
	uc := repo.userCoupons[1]
	require.NotNil(t, uc)
	assert.Equal(t, model.UserCouponSourceGift, uc.Source)
}

func TestCouponService_Receive_CreateUserCouponError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive})
	repo.createUCErr = errors.New("uc insert fail")
	svc := NewCouponService(repo)

	err := svc.Receive(1, 1, "")
	assert.Equal(t, "uc insert fail", err.Error())
}

func TestCouponService_Receive_IncrError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive})
	repo.incrErr = errors.New("incr fail")
	svc := NewCouponService(repo)

	err := svc.Receive(1, 1, "")
	assert.Equal(t, "incr fail", err.Error())
}

// --- Use ---

func TestCouponService_Use_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Use(99, 100), ErrUserCouponNotFound)
}

func TestCouponService_Use_AlreadyUsed(t *testing.T) {
	repo := newMockCouponRepo()
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusUsed})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Use(1, 100), ErrUserCouponUsed)
}

func TestCouponService_Use_Expired(t *testing.T) {
	repo := newMockCouponRepo()
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusExpired})
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Use(1, 100), ErrUserCouponExpired)
}

func TestCouponService_Use_Success(t *testing.T) {
	repo := newMockCouponRepo()
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusUnused})
	svc := NewCouponService(repo)

	require.NoError(t, svc.Use(1, 200))
	fields := repo.updatedUserCoupons[1]
	assert.Equal(t, model.UserCouponStatusUsed, fields["status"])
	assert.Equal(t, uint(200), fields["order_id"])
	_, ok := fields["used_at"]
	assert.True(t, ok, "应记录 used_at")
}

// --- Refund ---

func TestCouponService_Refund_NotFound(t *testing.T) {
	repo := newMockCouponRepo()
	svc := NewCouponService(repo)

	assert.ErrorIs(t, svc.Refund(99), ErrUserCouponNotFound)
}

func TestCouponService_Refund_NotUsed(t *testing.T) {
	repo := newMockCouponRepo()
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusUnused})
	svc := NewCouponService(repo)

	// 未使用状态不允许退还（沿用 ErrUserCouponUsed 语义）
	assert.ErrorIs(t, svc.Refund(1), ErrUserCouponUsed)
}

func TestCouponService_Refund_Success(t *testing.T) {
	repo := newMockCouponRepo()
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusUsed})
	svc := NewCouponService(repo)

	require.NoError(t, svc.Refund(1))
	fields := repo.updatedUserCoupons[1]
	assert.Equal(t, model.UserCouponStatusUnused, fields["status"])
	assert.Equal(t, nil, fields["used_at"])
}

// --- ListMine ---

func TestCouponService_ListMine_WithEnrichment(t *testing.T) {
	repo := newMockCouponRepo()
	repo.coupons[10] = newCoupon(10, 1, model.Coupon{Title: "满50减10", Type: model.CouponTypeReduce, Amount: 10})
	repo.userCoupons[1] = newUC(1, model.UserCoupon{UserID: 7, CouponID: 10, Status: model.UserCouponStatusUnused, Source: model.UserCouponSourceReceive})
	svc := NewCouponService(repo)

	p, list, err := svc.ListMine(7, &dto.UserCouponListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "满50减10", list[0].CouponTitle)
	assert.Equal(t, model.CouponTypeReduce, list[0].CouponType)
	assert.Equal(t, float64(10), list[0].CouponAmount)
	assert.Equal(t, "未使用", list[0].StatusText)
	assert.Equal(t, "主动领取", list[0].SourceText)
}

func TestCouponService_ListMine_RepoError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.listUCErr = errors.New("list uc fail")
	svc := NewCouponService(repo)

	_, _, err := svc.ListMine(1, &dto.UserCouponListRequest{})
	assert.Equal(t, "list uc fail", err.Error())
}

// --- Statistics ---

func TestCouponService_Statistics(t *testing.T) {
	repo := newMockCouponRepo()
	regionID := uint(3)
	repo.coupons[1] = newCoupon(1, regionID, model.Coupon{Status: model.CouponStatusActive, ReceivedCount: 50})
	repo.coupons[2] = newCoupon(2, regionID, model.Coupon{Status: model.CouponStatusActive, ReceivedCount: 30})
	repo.coupons[3] = newCoupon(3, regionID, model.Coupon{Status: model.CouponStatusOffline, ReceivedCount: 10})
	// 已使用的用户券 2 张
	repo.userCoupons[1] = newUC(1, model.UserCoupon{Status: model.UserCouponStatusUsed})
	repo.userCoupons[2] = newUC(2, model.UserCoupon{Status: model.UserCouponStatusUsed})
	svc := NewCouponService(repo)

	stats, err := svc.Statistics(regionID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalCoupons)
	assert.Equal(t, int64(2), stats.ActiveCoupons)
	assert.Equal(t, int64(90), stats.TotalReceived) // 50+30+10
	assert.Equal(t, int64(2), stats.TotalUsed)
	assert.Equal(t, float64(90)/float64(3), stats.ReceiveRate)
	assert.Equal(t, float64(2)/float64(90), stats.UsageRate)
}

func TestCouponService_Statistics_ListError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.listErr = errors.New("list err")
	svc := NewCouponService(repo)

	_, err := svc.Statistics(1)
	assert.Equal(t, "list err", err.Error())
}

// --- ExpireCoupons ---

func TestCouponService_ExpireCoupons(t *testing.T) {
	repo := newMockCouponRepo()
	past := time.Now().Add(-time.Hour)
	// 进行中但已过期 end_at → 应被更新为已过期
	repo.coupons[1] = newCoupon(1, 1, model.Coupon{Status: model.CouponStatusActive, EndAt: &past})
	// 进行中但未过期 → 不更新
	repo.coupons[2] = newCoupon(2, 1, model.Coupon{Status: model.CouponStatusActive})
	repo.expireUCAffected = 4
	svc := NewCouponService(repo)

	affected, err := svc.ExpireCoupons()
	require.NoError(t, err)
	// 1 张优惠券过期 + 4 张用户券过期
	assert.Equal(t, int64(5), affected)
	assert.Equal(t, model.CouponStatusExpired, repo.updatedCoupons[1]["status"])
	_, ok := repo.updatedCoupons[2]
	assert.False(t, ok, "未过期优惠券不应被更新")
}

func TestCouponService_ExpireCoupons_ListError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.listErr = errors.New("list err")
	svc := NewCouponService(repo)

	_, err := svc.ExpireCoupons()
	assert.Equal(t, "list err", err.Error())
}

func TestCouponService_ExpireCoupons_ExpireUCError(t *testing.T) {
	repo := newMockCouponRepo()
	repo.expireUCErr = errors.New("expire uc fail")
	svc := NewCouponService(repo)

	affected, err := svc.ExpireCoupons()
	assert.Equal(t, int64(0), affected)
	assert.Equal(t, "expire uc fail", err.Error())
}

// =====================================================================
// ===== sign 子域 mock =====
// =====================================================================

type mockSignRepo struct {
	rules           map[uint]*model.SignRule
	nextRuleID      uint
	records         []*model.SignRecord
	nextRecordID    uint
	createRecordErr error
	findToday       *model.SignRecord
	findTodayErr    error
	findLatest      *model.SignRecord
	findLatestErr   error
	listByMonth     []model.SignRecord
	listByMonthErr  error
	sumPoints       int
	sumErr          error
	createRuleErr   error
	findRuleByIDErr error
	findRuleByDay   *model.SignRule
	findRuleByDayErr error
	updateRuleErr   error
	deleteRuleErr   error
	listRulesErr    error
	listEnabledErr  error
	enabledRules    []model.SignRule
	createdRules    []*model.SignRule
	updatedRules    map[uint]map[string]interface{}
	deletedRules    map[uint]bool
}

func newMockSignRepo() *mockSignRepo {
	return &mockSignRepo{
		rules:        make(map[uint]*model.SignRule),
		nextRuleID:   1,
		nextRecordID: 1,
		updatedRules: make(map[uint]map[string]interface{}),
		deletedRules: make(map[uint]bool),
	}
}

func (m *mockSignRepo) CreateRecord(r *model.SignRecord) error {
	if m.createRecordErr != nil {
		return m.createRecordErr
	}
	r.ID = m.nextRecordID
	m.nextRecordID++
	cp := *r
	m.records = append(m.records, &cp)
	return nil
}

func (m *mockSignRepo) FindTodayRecord(userID uint, date time.Time) (*model.SignRecord, error) {
	if m.findTodayErr != nil {
		return nil, m.findTodayErr
	}
	if m.findToday != nil {
		return m.findToday, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockSignRepo) FindLatestRecord(userID uint) (*model.SignRecord, error) {
	if m.findLatestErr != nil {
		return nil, m.findLatestErr
	}
	if m.findLatest != nil {
		return m.findLatest, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockSignRepo) ListRecordsByMonth(userID uint, start, end time.Time) ([]model.SignRecord, error) {
	if m.listByMonthErr != nil {
		return nil, m.listByMonthErr
	}
	return m.listByMonth, nil
}

func (m *mockSignRepo) SumPoints(userID uint) (int, error) {
	if m.sumErr != nil {
		return 0, m.sumErr
	}
	return m.sumPoints, nil
}

func (m *mockSignRepo) CreateRule(rule *model.SignRule) error {
	if m.createRuleErr != nil {
		return m.createRuleErr
	}
	rule.ID = m.nextRuleID
	m.nextRuleID++
	cp := *rule
	m.rules[rule.ID] = &cp
	m.createdRules = append(m.createdRules, &cp)
	return nil
}

func (m *mockSignRepo) FindRuleByID(id uint) (*model.SignRule, error) {
	if m.findRuleByIDErr != nil {
		return nil, m.findRuleByIDErr
	}
	r, ok := m.rules[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockSignRepo) FindRuleByDay(day int) (*model.SignRule, error) {
	if m.findRuleByDayErr != nil {
		return nil, m.findRuleByDayErr
	}
	if m.findRuleByDay != nil {
		return m.findRuleByDay, nil
	}
	for _, r := range m.rules {
		if r.Day == day {
			cp := *r
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockSignRepo) UpdateRule(id uint, fields map[string]interface{}) error {
	if m.updateRuleErr != nil {
		return m.updateRuleErr
	}
	m.updatedRules[id] = fields
	return nil
}

func (m *mockSignRepo) DeleteRule(id uint) error {
	if m.deleteRuleErr != nil {
		return m.deleteRuleErr
	}
	delete(m.rules, id)
	m.deletedRules[id] = true
	return nil
}

func (m *mockSignRepo) ListRules(query repository.SignRuleListQuery, p *utils.Pagination) ([]model.SignRule, int64, error) {
	if m.listRulesErr != nil {
		return nil, 0, m.listRulesErr
	}
	var list []model.SignRule
	for _, r := range m.rules {
		if query.Status != nil && r.Status != *query.Status {
			continue
		}
		list = append(list, *r)
	}
	return list, int64(len(list)), nil
}

func (m *mockSignRepo) ListEnabledRules() ([]model.SignRule, error) {
	if m.listEnabledErr != nil {
		return nil, m.listEnabledErr
	}
	if m.enabledRules != nil {
		return m.enabledRules, nil
	}
	var list []model.SignRule
	for _, r := range m.rules {
		if r.Status == model.SignRuleStatusEnabled {
			list = append(list, *r)
		}
	}
	return list, nil
}

// ===== sign 子域测试 =====

func TestSignService_CheckIn_AlreadyToday(t *testing.T) {
	repo := newMockSignRepo()
	repo.findToday = &model.SignRecord{UserID: 1}
	svc := NewSignService(repo)

	_, err := svc.CheckIn(1)
	assert.ErrorIs(t, err, ErrSignAlreadyToday)
}

func TestSignService_CheckIn_FindTodayOtherError(t *testing.T) {
	repo := newMockSignRepo()
	repo.findTodayErr = errors.New("today boom")
	svc := NewSignService(repo)

	_, err := svc.CheckIn(1)
	assert.Equal(t, "today boom", err.Error())
}

func TestSignService_CheckIn_FirstDay(t *testing.T) {
	repo := newMockSignRepo()
	// 今日未签到（findToday 为 nil → 返回 ErrRecordNotFound），无历史记录
	repo.sumPoints = 42
	svc := NewSignService(repo)

	resp, err := svc.CheckIn(1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.ContinuousDays)
	assert.Equal(t, 1, resp.Points) // 无规则 → 基础 1 积分
	assert.Equal(t, 42, resp.TotalPoints)
	require.NotNil(t, resp.Record)
	assert.Equal(t, uint(1), resp.Record.UserID)
	assert.Equal(t, 1, resp.Record.ContinuousDays)
}

func TestSignService_CheckIn_ContinuousFromYesterday(t *testing.T) {
	repo := newMockSignRepo()
	yesterday := todayDate().AddDate(0, 0, -1)
	repo.findLatest = &model.SignRecord{UserID: 1, SignDate: yesterday, ContinuousDays: 2}
	repo.sumPoints = 100
	svc := NewSignService(repo)

	resp, err := svc.CheckIn(1)
	require.NoError(t, err)
	assert.Equal(t, 3, resp.ContinuousDays) // 2 + 1
}

func TestSignService_CheckIn_BrokenStreak(t *testing.T) {
	repo := newMockSignRepo()
	twoDaysAgo := todayDate().AddDate(0, 0, -2)
	repo.findLatest = &model.SignRecord{UserID: 1, SignDate: twoDaysAgo, ContinuousDays: 5}
	svc := NewSignService(repo)

	resp, err := svc.CheckIn(1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.ContinuousDays) // 中断 → 重置为 1
}

func TestSignService_CheckIn_RuleHitPoints(t *testing.T) {
	repo := newMockSignRepo()
	repo.findLatest = &model.SignRecord{UserID: 1, SignDate: todayDate().AddDate(0, 0, -1), ContinuousDays: 6}
	// 第 7 天命中规则：10 积分 + 额外奖励
	extra, _ := model.FromJSON(map[string]interface{}{"coupon": "C7"})
	repo.findRuleByDay = &model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled, ExtraReward: extra}
	svc := NewSignService(repo)

	resp, err := svc.CheckIn(1)
	require.NoError(t, err)
	assert.Equal(t, 7, resp.ContinuousDays)
	assert.Equal(t, 10, resp.Points)
	em, ok := resp.ExtraReward.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "C7", em["coupon"])
}

func TestSignService_CheckIn_DisabledRuleFallsBackToBase(t *testing.T) {
	repo := newMockSignRepo()
	repo.findLatest = &model.SignRecord{UserID: 1, SignDate: todayDate().AddDate(0, 0, -1), ContinuousDays: 6}
	repo.findRuleByDay = &model.SignRule{Day: 7, Points: 99, Status: model.SignRuleStatusDisabled}
	svc := NewSignService(repo)

	resp, err := svc.CheckIn(1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Points) // 规则禁用 → 基础 1 积分
}

func TestSignService_CheckIn_CreateRecordError(t *testing.T) {
	repo := newMockSignRepo()
	repo.createRecordErr = errors.New("create record fail")
	svc := NewSignService(repo)

	_, err := svc.CheckIn(1)
	assert.Equal(t, "create record fail", err.Error())
}

func TestSignService_GetCalendar_InvalidMonth(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	_, err := svc.GetCalendar(1, "2026/07")
	assert.EqualError(t, err, "月份格式应为 YYYY-MM")
}

func TestSignService_GetCalendar_DefaultMonth(t *testing.T) {
	repo := newMockSignRepo()
	repo.listByMonth = []model.SignRecord{
		newRecord(1, model.SignRecord{UserID: 1, Points: 2, ContinuousDays: 1}),
		newRecord(2, model.SignRecord{UserID: 1, Points: 3, ContinuousDays: 2}),
	}
	// 最新签到为今日 → 连续天数有效
	repo.findLatest = &model.SignRecord{SignDate: todayDate(), ContinuousDays: 2}
	svc := NewSignService(repo)

	resp, err := svc.GetCalendar(1, "")
	require.NoError(t, err)
	assert.Len(t, resp.Records, 2)
	assert.Equal(t, 2, resp.SignedDays)
	assert.Equal(t, 5, resp.TotalPoints)
	assert.Equal(t, 2, resp.ContinuousDays)
	assert.GreaterOrEqual(t, resp.MonthDays, 28)
}

func TestSignService_GetCalendar_SpecificMonth(t *testing.T) {
	repo := newMockSignRepo()
	repo.listByMonth = []model.SignRecord{newRecord(1, model.SignRecord{UserID: 1, Points: 5})}
	svc := NewSignService(repo)

	resp, err := svc.GetCalendar(1, "2026-02")
	require.NoError(t, err)
	assert.Equal(t, 1, resp.SignedDays)
	assert.Equal(t, 5, resp.TotalPoints)
	assert.Equal(t, 28, resp.MonthDays) // 2026 年 2 月 28 天
}

func TestSignService_GetCalendar_ListError(t *testing.T) {
	repo := newMockSignRepo()
	repo.listByMonthErr = errors.New("list month fail")
	svc := NewSignService(repo)

	_, err := svc.GetCalendar(1, "2026-02")
	assert.Equal(t, "list month fail", err.Error())
}

// --- sign rule CRUD ---

func TestSignService_CreateRule_DefaultStatus(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	info, err := svc.CreateRule(&dto.CreateSignRuleRequest{Day: 7, Points: 10})
	require.NoError(t, err)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, 7, info.Day)
	assert.Equal(t, 10, info.Points)
	assert.Equal(t, model.SignRuleStatusEnabled, info.Status) // 默认启用
	assert.Equal(t, "启用", info.StatusText)
}

func TestSignService_CreateRule_WithExtraReward(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	info, err := svc.CreateRule(&dto.CreateSignRuleRequest{
		Day:         30,
		Points:      50,
		ExtraReward: map[string]interface{}{"coupon": "C30"},
	})
	require.NoError(t, err)
	em, ok := info.ExtraReward.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "C30", em["coupon"])
}

func TestSignService_CreateRule_Exists(t *testing.T) {
	repo := newMockSignRepo()
	repo.findRuleByDay = &model.SignRule{Day: 7}
	svc := NewSignService(repo)

	_, err := svc.CreateRule(&dto.CreateSignRuleRequest{Day: 7, Points: 10})
	assert.ErrorIs(t, err, ErrSignRuleExists)
}

func TestSignService_CreateRule_FindByDayOtherError(t *testing.T) {
	repo := newMockSignRepo()
	repo.findRuleByDayErr = errors.New("find day boom")
	svc := NewSignService(repo)

	_, err := svc.CreateRule(&dto.CreateSignRuleRequest{Day: 7, Points: 10})
	assert.Equal(t, "find day boom", err.Error())
}

func TestSignService_CreateRule_CreateError(t *testing.T) {
	repo := newMockSignRepo()
	repo.createRuleErr = errors.New("create rule fail")
	svc := NewSignService(repo)

	_, err := svc.CreateRule(&dto.CreateSignRuleRequest{Day: 7, Points: 10})
	assert.Equal(t, "create rule fail", err.Error())
}

func TestSignService_UpdateRule(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled})
	svc := NewSignService(repo)

	err := svc.UpdateRule(1, &dto.UpdateSignRuleRequest{
		Points: intPtr(20),
		Status: intPtr(model.SignRuleStatusDisabled),
	})
	require.NoError(t, err)
	assert.Equal(t, 20, repo.updatedRules[1]["points"])
	assert.Equal(t, model.SignRuleStatusDisabled, repo.updatedRules[1]["status"])
}

func TestSignService_UpdateRule_NotFound(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	assert.ErrorIs(t, svc.UpdateRule(99, &dto.UpdateSignRuleRequest{Points: intPtr(1)}), ErrSignRuleNotFound)
}

func TestSignService_UpdateRule_EmptyFields(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10})
	svc := NewSignService(repo)

	require.NoError(t, svc.UpdateRule(1, &dto.UpdateSignRuleRequest{}))
	_, ok := repo.updatedRules[1]
	assert.False(t, ok)
}

func TestSignService_DeleteRule_NotFound(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	assert.ErrorIs(t, svc.DeleteRule(1), ErrSignRuleNotFound)
}

func TestSignService_DeleteRule_Success(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10})
	svc := NewSignService(repo)

	require.NoError(t, svc.DeleteRule(1))
	assert.True(t, repo.deletedRules[1])
}

func TestSignService_GetRuleByID_NotFound(t *testing.T) {
	repo := newMockSignRepo()
	svc := NewSignService(repo)

	_, err := svc.GetRuleByID(1)
	assert.ErrorIs(t, err, ErrSignRuleNotFound)
}

func TestSignService_GetRuleByID_Success(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled})
	svc := NewSignService(repo)

	info, err := svc.GetRuleByID(1)
	require.NoError(t, err)
	assert.Equal(t, 7, info.Day)
	assert.Equal(t, "启用", info.StatusText)
}

func TestSignService_ListRules(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled})
	repo.rules[2] = newRule(2, model.SignRule{Day: 14, Points: 20, Status: model.SignRuleStatusDisabled})
	svc := NewSignService(repo)

	p, list, err := svc.ListRules(&dto.SignRuleListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	assert.Len(t, list, 2)
}

func TestSignService_ListRules_FilterByStatus(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled})
	repo.rules[2] = newRule(2, model.SignRule{Day: 14, Points: 20, Status: model.SignRuleStatusDisabled})
	svc := NewSignService(repo)

	p, list, err := svc.ListRules(&dto.SignRuleListRequest{Status: intPtr(model.SignRuleStatusEnabled)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	assert.Len(t, list, 1)
	assert.Equal(t, 7, list[0].Day)
}

func TestSignService_ListEnabledRules(t *testing.T) {
	repo := newMockSignRepo()
	repo.rules[1] = newRule(1, model.SignRule{Day: 7, Points: 10, Status: model.SignRuleStatusEnabled})
	repo.rules[2] = newRule(2, model.SignRule{Day: 14, Points: 20, Status: model.SignRuleStatusDisabled})
	svc := NewSignService(repo)

	list, err := svc.ListEnabledRules()
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, 7, list[0].Day)
}

func TestSignService_ListEnabledRules_Error(t *testing.T) {
	repo := newMockSignRepo()
	repo.listEnabledErr = errors.New("enabled err")
	svc := NewSignService(repo)

	_, err := svc.ListEnabledRules()
	assert.Equal(t, "enabled err", err.Error())
}
