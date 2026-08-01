// Package service 同城房屋租售主表业务逻辑层单元测试 - 房源包。
// 使用内存 mock 仓储覆盖：发布与默认值兜底、户型/均价计算、图片子表替换与封面兜底、
// 更新/删除权限校验与字段构建、状态机迁移（发布即审核通过）、详情浏览量自增与收藏/浏览记录、
// 附近查询半径兜底、收藏切换（创建/删除 + 计数增减）、相似推荐（相似度计算与自身剔除）、
// 管理端审核去重与状态/推广更新、批量操作结果汇总等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockHouseRepo =====

type updateFieldsCall struct {
	id     uint
	fields map[string]interface{}
}

type mockHouseRepo struct {
	byID   map[uint]*model.House
	nextID uint

	images   map[uint][]model.HouseImage
	favs     map[uint]map[uint]bool // userID -> houseID -> exists
	views    []model.HouseView
	incrView []uint

	createErr       error
	findErr         error
	updateFieldsErr error
	deleteErr       error
	replaceImgErr   error
	listErr         error
	favExistsErr    error
	createFavErr    error
	deleteFavErr    error
	batchErr        error

	updateFieldsCalls []updateFieldsCall
	deleteCalls        []uint
	replaceImagesCalls []struct {
		houseID uint
		imgs    []model.HouseImage
	}
	deleteImagesCalls []uint
	createFavCalls    []model.HouseFavorite
	deleteFavCalls     []struct {
		userID  uint
		houseID uint
	}
	incrFavCountCalls    []uint
	decrFavCountCalls    []uint
	createViewCalls      []model.HouseView
	batchUpdateStatusCalls []struct {
		ids    []uint
		status int
	}
	batchAuditCalls []struct {
		ids         []uint
		auditStatus int
		auditReason string
	}
	batchDeleteCalls []struct {
		ids []uint
	}

	listReturn      []model.House
	listTotal       int64
	adminListReturn []model.House
	adminListTotal  int64
	nearbyReturn   []model.House
	nearbyTotal     int64
	searchReturn    []model.House
	searchTotal     int64
	advancedReturn  []model.House
	advancedTotal   int64
	byUserReturn    []model.House
	byUserTotal     int64
	listFavsReturn  []model.HouseFavorite
	listFavsTotal   int64
}

func newMockHouseRepo() *mockHouseRepo {
	return &mockHouseRepo{
		byID:   make(map[uint]*model.House),
		images: make(map[uint][]model.HouseImage),
		favs:   make(map[uint]map[uint]bool),
		nextID: 1,
	}
}

// seed 预置一条房源并返回其副本指针。
func (m *mockHouseRepo) seed(h *model.House) *model.House {
	if h.ID == 0 {
		h.ID = m.nextID
		m.nextID++
	}
	cp := *h
	m.byID[h.ID] = &cp
	return &cp
}

func (m *mockHouseRepo) Create(h *model.House) error {
	if m.createErr != nil {
		return m.createErr
	}
	if h.ID == 0 {
		h.ID = m.nextID
		m.nextID++
	}
	cp := *h
	m.byID[h.ID] = &cp
	return nil
}

func (m *mockHouseRepo) FindByID(id uint) (*model.House, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	h, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *h
	return &cp, nil
}

func (m *mockHouseRepo) Update(h *model.House) error {
	cp := *h
	m.byID[h.ID] = &cp
	return nil
}

func (m *mockHouseRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, updateFieldsCall{id: id, fields: fields})
	h, ok := m.byID[id]
	if !ok {
		return nil
	}
	if v, ok := fields["status"]; ok {
		if s, ok2 := v.(int); ok2 {
			h.Status = s
		}
	}
	if v, ok := fields["audit_status"]; ok {
		if s, ok2 := v.(int); ok2 {
			h.AuditStatus = s
		}
	}
	if v, ok := fields["audit_reason"]; ok {
		if s, ok2 := v.(string); ok2 {
			h.AuditReason = s
		}
	}
	if v, ok := fields["published_at"]; ok {
		if t, ok2 := v.(*time.Time); ok2 && t != nil {
			h.PublishedAt = t
		}
	}
	if v, ok := fields["title"]; ok {
		if s, ok2 := v.(string); ok2 {
			h.Title = s
		}
	}
	if v, ok := fields["layout"]; ok {
		if s, ok2 := v.(string); ok2 {
			h.Layout = s
		}
	}
	if v, ok := fields["average_price"]; ok {
		if f, ok2 := v.(float64); ok2 {
			h.AveragePrice = f
		}
	}
	if v, ok := fields["featured"]; ok {
		if b, ok2 := v.(bool); ok2 {
			h.Featured = b
		}
	}
	if v, ok := fields["picked"]; ok {
		if b, ok2 := v.(bool); ok2 {
			h.Picked = b
		}
	}
	if v, ok := fields["verified"]; ok {
		if b, ok2 := v.(bool); ok2 {
			h.Verified = b
		}
	}
	if v, ok := fields["real_house_verified"]; ok {
		if b, ok2 := v.(bool); ok2 {
			h.RealHouseVerified = b
		}
	}
	if v, ok := fields["promotion_level"]; ok {
		if i, ok2 := v.(int); ok2 {
			h.PromotionLevel = i
		}
	}
	return nil
}

func (m *mockHouseRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.byID, id)
	return nil
}

func (m *mockHouseRepo) List(regionID uint, req *utils.Pagination, opts repository.HouseListOptions) ([]model.House, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockHouseRepo) AdminList(req *utils.Pagination, opts repository.HouseAdminListOptions) ([]model.House, int64, error) {
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockHouseRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.HouseListOptions) ([]model.House, int64, error) {
	return m.nearbyReturn, m.nearbyTotal, nil
}

func (m *mockHouseRepo) Search(regionID uint, req *utils.Pagination, keyword string) ([]model.House, int64, error) {
	return m.searchReturn, m.searchTotal, nil
}

func (m *mockHouseRepo) AdvancedSearch(regionID uint, req *utils.Pagination, opts repository.HouseAdvancedSearchOptions) ([]model.House, int64, error) {
	return m.advancedReturn, m.advancedTotal, nil
}

func (m *mockHouseRepo) ListByUser(userID uint, req *utils.Pagination) ([]model.House, int64, error) {
	return m.byUserReturn, m.byUserTotal, nil
}

func (m *mockHouseRepo) IncrViewCount(id uint) error {
	m.incrView = append(m.incrView, id)
	if h, ok := m.byID[id]; ok {
		h.ViewCount++
	}
	return nil
}

func (m *mockHouseRepo) IncrContactCount(id uint) error {
	if h, ok := m.byID[id]; ok {
		h.ContactCount++
	}
	return nil
}

func (m *mockHouseRepo) IncrShareCount(id uint) error {
	if h, ok := m.byID[id]; ok {
		h.ShareCount++
	}
	return nil
}

func (m *mockHouseRepo) IncrViewingCount(id uint) error {
	if h, ok := m.byID[id]; ok {
		h.ViewingCount++
	}
	return nil
}

func (m *mockHouseRepo) ListImages(houseID uint) ([]model.HouseImage, error) {
	return m.images[houseID], nil
}

func (m *mockHouseRepo) ReplaceImages(houseID uint, imgs []model.HouseImage) error {
	if m.replaceImgErr != nil {
		return m.replaceImgErr
	}
	m.replaceImagesCalls = append(m.replaceImagesCalls, struct {
		houseID uint
		imgs    []model.HouseImage
	}{houseID: houseID, imgs: imgs})
	m.images[houseID] = imgs
	return nil
}

func (m *mockHouseRepo) DeleteImages(houseID uint) error {
	m.deleteImagesCalls = append(m.deleteImagesCalls, houseID)
	delete(m.images, houseID)
	return nil
}

func (m *mockHouseRepo) FavExists(userID, houseID uint) (bool, error) {
	if m.favExistsErr != nil {
		return false, m.favExistsErr
	}
	return m.favs[userID][houseID], nil
}

func (m *mockHouseRepo) CreateFav(fav *model.HouseFavorite) error {
	if m.createFavErr != nil {
		return m.createFavErr
	}
	m.createFavCalls = append(m.createFavCalls, *fav)
	if m.favs[fav.UserID] == nil {
		m.favs[fav.UserID] = make(map[uint]bool)
	}
	m.favs[fav.UserID][fav.HouseID] = true
	return nil
}

func (m *mockHouseRepo) DeleteFav(userID, houseID uint) error {
	if m.deleteFavErr != nil {
		return m.deleteFavErr
	}
	m.deleteFavCalls = append(m.deleteFavCalls, struct {
		userID  uint
		houseID uint
	}{userID: userID, houseID: houseID})
	if m.favs[userID] != nil {
		delete(m.favs[userID], houseID)
	}
	return nil
}

func (m *mockHouseRepo) IncrFavCount(id uint) error {
	m.incrFavCountCalls = append(m.incrFavCountCalls, id)
	if h, ok := m.byID[id]; ok {
		h.FavCount++
	}
	return nil
}

func (m *mockHouseRepo) DecrFavCount(id uint) error {
	m.decrFavCountCalls = append(m.decrFavCountCalls, id)
	if h, ok := m.byID[id]; ok && h.FavCount > 0 {
		h.FavCount--
	}
	return nil
}

func (m *mockHouseRepo) ListFavs(userID uint, page, pageSize int) ([]model.HouseFavorite, int64, error) {
	return m.listFavsReturn, m.listFavsTotal, nil
}

func (m *mockHouseRepo) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	for _, id := range ids {
		result[id] = m.favs[userID][id]
	}
	return result, nil
}

func (m *mockHouseRepo) CreateView(v *model.HouseView) error {
	m.createViewCalls = append(m.createViewCalls, *v)
	m.views = append(m.views, *v)
	return nil
}

func (m *mockHouseRepo) ListViews(userID uint, page, pageSize int) ([]model.HouseView, int64, error) {
	return m.views, int64(len(m.views)), nil
}

func (m *mockHouseRepo) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	if m.batchErr != nil {
		return 0, m.batchErr
	}
	m.batchUpdateStatusCalls = append(m.batchUpdateStatusCalls, struct {
		ids    []uint
		status int
	}{ids: ids, status: status})
	return int64(len(ids)), nil
}

func (m *mockHouseRepo) BatchAudit(ids []uint, auditStatus int, auditReason string) (int64, error) {
	if m.batchErr != nil {
		return 0, m.batchErr
	}
	m.batchAuditCalls = append(m.batchAuditCalls, struct {
		ids         []uint
		auditStatus int
		auditReason string
	}{ids: ids, auditStatus: auditStatus, auditReason: auditReason})
	return int64(len(ids)), nil
}

func (m *mockHouseRepo) BatchDelete(ids []uint) (int64, error) {
	if m.batchErr != nil {
		return 0, m.batchErr
	}
	m.batchDeleteCalls = append(m.batchDeleteCalls, struct{ ids []uint }{ids: ids})
	return int64(len(ids)), nil
}

func (m *mockHouseRepo) CountByStatus(regionID uint) (map[int]int64, error) {
	return map[int]int64{}, nil
}

func (m *mockHouseRepo) CountPendingAudit(regionID uint) (int64, error) { return 0, nil }

func (m *mockHouseRepo) CountTodayNew(regionID uint) (int64, error) { return 0, nil }

// ===== mockAgentRepo / mockCommunityRepo =====
// 仅 houseService 调用 FindByID，故内嵌接口（nil 默认）仅覆盖该方法。

type mockAgentRepo struct {
	repository.AgentRepository
	agent *model.HouseAgent
	err   error
}

func (m *mockAgentRepo) FindByID(id uint) (*model.HouseAgent, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.agent == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.agent, nil
}

type mockCommunityRepo struct {
	repository.CommunityRepository
	community *model.HouseCommunity
	err       error
}

func (m *mockCommunityRepo) FindByID(id uint) (*model.HouseCommunity, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.community == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.community, nil
}

// ===== 测试装配 =====

func newSvc() (HouseService, *mockHouseRepo, *mockAgentRepo, *mockCommunityRepo) {
	repo := newMockHouseRepo()
	agent := &mockAgentRepo{}
	community := &mockCommunityRepo{}
	return NewHouseService(repo, agent, community), repo, agent, community
}

// mkHouse 构造带指定 ID 的 House。因 ID 为内嵌 BaseModel 的提升字段，
// 不能在复合字面量中直接赋值，故通过取地址后的字段赋值完成。
func mkHouse(id uint, base model.House) model.House {
	base.ID = id
	return base
}

// ===== buildLayout / itoa =====

func TestBuildLayout(t *testing.T) {
	cases := []struct {
		rooms, halls, baths int
		want                string
	}{
		{3, 2, 2, "3室2厅2卫"},
		{2, 1, 1, "2室1厅1卫"},
		{0, 0, 0, ""},
		{4, 0, 0, "4室"},
		{0, 2, 0, "2厅"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, buildLayout(c.rooms, c.halls, c.baths))
	}
}

func TestItoa(t *testing.T) {
	assert.Equal(t, "0", itoa(0))
	assert.Equal(t, "12", itoa(12))
	assert.Equal(t, "-7", itoa(-7))
	assert.Equal(t, "2024", itoa(2024))
}

// ===== Create =====

func TestCreate_DefaultsAndAuditApproved(t *testing.T) {
	svc, repo, _, _ := newSvc()
	req := &dto.CreateHouseRequest{Title: "整租三居室", Rooms: 3, Halls: 2, Bathrooms: 2}

	info, err := svc.Create(5, 100, "张三", "13800000000", "avatar.png", req)
	require.NoError(t, err)
	require.NotNil(t, info)

	// 默认值兜底
	assert.Equal(t, model.ListingTypeRent, info.ListingType)
	assert.Equal(t, model.PropertyTypeResidential, info.PropertyType)
	assert.Equal(t, model.SourceTypePersonal, info.SourceType)
	assert.Equal(t, model.RentUnitMonth, info.RentUnit)
	assert.Equal(t, model.RentTypeEntire, info.RentType)
	assert.Equal(t, model.DepositTypeOneMonth, info.DepositType)
	assert.Equal(t, model.PaymentMethodMonthly, info.PaymentMethod)
	assert.Equal(t, model.FloorTypeMid, info.FloorType)
	assert.Equal(t, model.DecorationRough, info.Decoration)
	assert.Equal(t, model.PropertyOwnershipCommercial, info.PropertyOwnership)
	// MVP：发布即审核通过
	assert.Equal(t, model.AuditApproved, info.AuditStatus)
	// 户型自动拼装
	assert.Equal(t, "3室2厅2卫", info.Layout)
	// 地区与发布者
	assert.Equal(t, uint(5), info.RegionID)
	assert.Equal(t, uint(100), info.UserID)
	assert.Equal(t, "张三", info.UserName)
	// 草稿态不设置发布时间
	assert.Nil(t, info.PublishedAt)
	// Images 字段空切片而非 nil
	assert.Equal(t, []dto.HouseImageInfo{}, info.Images)
	// 落库
	require.Len(t, repo.byID, 1)
	assert.Equal(t, model.AuditApproved, repo.byID[info.ID].AuditStatus)
}

func TestCreate_PublishedSetsPublishedAt(t *testing.T) {
	svc, repo, _, _ := newSvc()
	req := &dto.CreateHouseRequest{Title: "出售房", Status: model.StatusPublished, SalePrice: 5000000, BuildingArea: 100}

	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPublished, info.Status)
	require.NotNil(t, info.PublishedAt)
	require.NotNil(t, repo.byID[info.ID].PublishedAt)
	// 均价 = 售价 / 建筑面积
	assert.InDelta(t, 50000.0, info.AveragePrice, 0.01)
}

func TestCreate_AveragePriceZeroWhenAreaMissing(t *testing.T) {
	svc, _, _, _ := newSvc()
	req := &dto.CreateHouseRequest{Title: "无面积", SalePrice: 100, BuildingArea: 0}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, 0.0, info.AveragePrice)
}

func TestCreate_ImagesReplaceAndCoverFallback(t *testing.T) {
	svc, repo, _, _ := newSvc()
	req := &dto.CreateHouseRequest{
		Title: "带图房源",
		Images: []dto.HouseImageInput{
			{URL: "img1.jpg", ImageType: ""},
			{URL: "img2.jpg", ImageType: model.ImageTypeFloorPlan},
		},
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)

	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, info.ID, repo.replaceImagesCalls[0].houseID)
	imgs := repo.replaceImagesCalls[0].imgs
	require.Len(t, imgs, 2)
	// 未指定类型默认实景图
	assert.Equal(t, model.ImageTypeReal, imgs[0].ImageType)
	// 首张图作为封面（CoverImage 为空时）
	assert.True(t, imgs[0].IsCover)
	assert.False(t, imgs[1].IsCover)
	// 启用状态
	assert.Equal(t, model.ImageStatusEnabled, imgs[0].Status)
}

func TestCreate_RepoError(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.createErr = errors.New("db down")
	_, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateHouseRequest{Title: "x"})
	require.Error(t, err)
}

// ===== Update =====

func TestUpdate_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	err := svc.Update(999, 1, &dto.UpdateHouseRequest{})
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestUpdate_NoPermission(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	err := svc.Update(1, 200, &dto.UpdateHouseRequest{Title: "改标题"})
	assert.ErrorIs(t, err, ErrHouseNoPermission)
}

func TestUpdate_FieldsApplied(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	err := svc.Update(1, 100, &dto.UpdateHouseRequest{
		Title:       "新标题",
		Content:     "新内容",
		CoverImage:  "cover.jpg",
		ListingType: model.ListingTypeSale,
		Rooms:       3, Halls: 1, Bathrooms: 1,
		BuildingArea: 90, SalePrice: 3600000,
	})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, "新标题", fields["title"])
	assert.Equal(t, "新内容", fields["content"])
	assert.Equal(t, "cover.jpg", fields["cover_image"])
	assert.Equal(t, model.ListingTypeSale, fields["listing_type"])
	// 户型重新拼装
	assert.Equal(t, "3室1厅1卫", fields["layout"])
	// 均价重新计算
	assert.InDelta(t, 40000.0, fields["average_price"], 0.01)
}

func TestUpdate_StatusPublishMigration(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, Status: model.StatusDraft})
	err := svc.Update(1, 1, &dto.UpdateHouseRequest{Status: model.StatusPublished})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	require.Contains(t, fields, "published_at")
}

func TestUpdate_StatusOfflineKeepsNoPublishAt(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, Status: model.StatusPublished})
	err := svc.Update(1, 1, &dto.UpdateHouseRequest{Status: model.StatusOffline})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	assert.NotContains(t, fields, "published_at")
}

func TestUpdate_ImagesReplaced(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	err := svc.Update(1, 1, &dto.UpdateHouseRequest{
		Images: []dto.HouseImageInput{{URL: "new.jpg"}},
	})
	require.NoError(t, err)
	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, uint(1), repo.replaceImagesCalls[0].houseID)
	// 更新路径首张图无条件设为封面
	assert.True(t, repo.replaceImagesCalls[0].imgs[0].IsCover)
}

func TestUpdate_UpdateFieldsError(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.Update(1, 1, &dto.UpdateHouseRequest{Title: "x"})
	require.Error(t, err)
}

// ===== Delete =====

func TestDelete_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	err := svc.Delete(999, 1)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestDelete_NoPermission(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	err := svc.Delete(1, 200)
	assert.ErrorIs(t, err, ErrHouseNoPermission)
}

func TestDelete_OwnerDeletesAndImages(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	err := svc.Delete(1, 100)
	require.NoError(t, err)
	require.Contains(t, repo.deleteCalls, uint(1))
	require.Contains(t, repo.deleteImagesCalls, uint(1))
}

func TestDelete_DeleteError(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	repo.deleteErr = errors.New("cannot delete")
	err := svc.Delete(1, 1)
	require.Error(t, err)
}

// ===== GetByID =====

func TestGetByID_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	_, err := svc.GetByID(999, 1)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestGetByID_IncrementsViewAndReturnsImages(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100, Status: model.StatusPublished, AuditStatus: model.AuditApproved})
	repo.images[1] = []model.HouseImage{{URL: "img.jpg", IsCover: true}}

	resp, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	require.NotNil(t, resp)
	// 浏览量自增
	require.Contains(t, repo.incrView, uint(1))
	assert.Equal(t, 1, resp.ViewCount)
	// 图片拼装
	require.Len(t, resp.Images, 1)
	assert.Equal(t, "img.jpg", resp.Images[0].URL)
	assert.True(t, resp.Images[0].IsCover)
	// 未登录不写浏览记录
	assert.Empty(t, repo.createViewCalls)
}

func TestGetByID_LoggedInRecordsViewAndFav(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	// 预置该用户已收藏
	if repo.favs[7] == nil {
		repo.favs[7] = make(map[uint]bool)
	}
	repo.favs[7][1] = true

	resp, err := svc.GetByID(1, 7)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	require.Len(t, repo.createViewCalls, 1)
	assert.Equal(t, uint(7), repo.createViewCalls[0].UserID)
	assert.Equal(t, model.FavoriteTypeHouse, repo.createViewCalls[0].ViewType)
	assert.Equal(t, model.ViewSourceList, repo.createViewCalls[0].Source)
}

func TestGetByID_PopulatesAgentAndCommunity(t *testing.T) {
	svc, repo, agent, community := newSvc()
	agentID := uint(8)
	commID := uint(9)
	repo.seed(&model.House{UserID: 1, AgentID: &agentID, CommunityID: &commID})
	agent.agent = &model.HouseAgent{Name: "经纪人A"}
	community.community = &model.HouseCommunity{Name: "阳光小区"}

	resp, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	require.NotNil(t, resp.Agent)
	assert.Equal(t, "经纪人A", resp.Agent.Name)
	require.NotNil(t, resp.Community)
	assert.Equal(t, "阳光小区", resp.Community.Name)
}

func TestGetByID_AgentNotFoundNoPanic(t *testing.T) {
	svc, repo, _, _ := newSvc()
	aid := uint(99)
	repo.seed(&model.House{UserID: 1, AgentID: &aid})
	resp, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	assert.Nil(t, resp.Agent)
}

// ===== List / Nearby / Search / AdvancedSearch / ListMine =====

func TestList_PaginationAndMapping(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.listReturn = []model.House{{Title: "A"}, {Title: "B"}}
	repo.listTotal = 2

	pg, list, err := svc.List(1, &dto.HouseListRequest{})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), pg.Total)
	assert.Equal(t, "A", list[0].Title)
}

func TestList_RepoError(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.listErr = errors.New("list err")
	_, _, err := svc.List(1, &dto.HouseListRequest{})
	require.Error(t, err)
}

func TestListNearby_RadiusFallback(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.nearbyReturn = []model.House{{}}
	repo.nearbyTotal = 1

	// RadiusKm <= 0 应兜底为 5（由 service 计算，repo 仅校验被调用）
	pg, list, err := svc.ListNearby(1, &dto.HouseNearbyRequest{Latitude: 30, Longitude: 120, RadiusKm: 0})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pg.Total)
	require.Len(t, list, 1)
}

func TestSearch_Mapping(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.searchReturn = []model.House{{Title: "命中"}}
	repo.searchTotal = 1
	_, list, err := svc.Search(1, &dto.HouseSearchRequest{Keyword: "命中"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "命中", list[0].Title)
}

func TestAdvancedSearch_Mapping(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.advancedReturn = []model.House{{}}
	repo.advancedTotal = 1
	pg, list, err := svc.AdvancedSearch(1, &dto.HouseAdvancedSearchRequest{RadiusKm: 3})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pg.Total)
	require.Len(t, list, 1)
}

func TestListMine_Mapping(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.byUserReturn = []model.House{{UserID: 100}}
	repo.byUserTotal = 1
	pg, list, err := svc.ListMine(100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), pg.Total)
	require.Len(t, list, 1)
}

// ===== Fav =====

func TestFav_AnonymousRejected(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	_, err := svc.Fav(0, 1)
	assert.ErrorIs(t, err, ErrHouseNoPermission)
}

func TestFav_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	_, err := svc.Fav(1, 999)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestFav_CreateWhenAbsent(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, FavCount: 2})

	resp, err := svc.Fav(7, 1)
	require.NoError(t, err)
	require.True(t, resp.HasFaved)
	assert.Equal(t, 3, resp.FavCount)
	require.Len(t, repo.createFavCalls, 1)
	assert.Equal(t, model.FavoriteTypeHouse, repo.createFavCalls[0].FavoriteType)
	assert.True(t, repo.createFavCalls[0].Notify)
	require.Contains(t, repo.incrFavCountCalls, uint(1))
}

func TestFav_RemoveWhenExists(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, FavCount: 3})
	if repo.favs[7] == nil {
		repo.favs[7] = make(map[uint]bool)
	}
	repo.favs[7][1] = true

	resp, err := svc.Fav(7, 1)
	require.NoError(t, err)
	require.False(t, resp.HasFaved)
	assert.Equal(t, 2, resp.FavCount)
	require.Len(t, repo.deleteFavCalls, 1)
	require.Contains(t, repo.decrFavCountCalls, uint(1))
}

func TestFav_FavExistsError(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	repo.favExistsErr = errors.New("err")
	_, err := svc.Fav(7, 1)
	require.Error(t, err)
}

// ===== FavStatus =====

func TestFavStatus_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	_, err := svc.FavStatus(1, 999)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestFavStatus_AnonymousReturnsFalse(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, FavCount: 5})
	resp, err := svc.FavStatus(0, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 5, resp.FavCount)
}

func TestFavStatus_LoggedInExists(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, FavCount: 5})
	if repo.favs[7] == nil {
		repo.favs[7] = make(map[uint]bool)
	}
	repo.favs[7][1] = true
	resp, err := svc.FavStatus(7, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
}

// ===== ListFavs =====

func TestListFavs_MapsHousesAndMarksFaved(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100, Title: "A"})
	repo.seed(&model.House{UserID: 100, Title: "B"})
	repo.listFavsReturn = []model.HouseFavorite{{HouseID: 1}, {HouseID: 2}}
	repo.listFavsTotal = 2

	pg, list, err := svc.ListFavs(7, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), pg.Total)
	require.Len(t, list, 2)
	for _, h := range list {
		assert.True(t, h.HasFaved)
	}
}

func TestListFavs_SkipsMissingHouse(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 100})
	// 一条收藏指向已删除的房源
	repo.listFavsReturn = []model.HouseFavorite{{HouseID: 1}, {HouseID: 999}}
	_, list, err := svc.ListFavs(7, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, uint(1), list[0].ID)
}

// ===== 互动 =====

func TestIncrContactCount_Delegates(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, ContactCount: 0})
	require.NoError(t, svc.IncrContactCount(1))
	assert.Equal(t, 1, repo.byID[1].ContactCount)
}

func TestIncrShareCount_Delegates(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, ShareCount: 0})
	require.NoError(t, svc.IncrShareCount(1))
	assert.Equal(t, 1, repo.byID[1].ShareCount)
}

// ===== ListSimilar =====

func TestListSimilar_LimitFallback(t *testing.T) {
	svc, _, _, _ := newSvc()
	// 不预置房源：FindByID 返回 ErrRecordNotFound -> ErrHouseNotFound
	_, err := svc.ListSimilar(999, 0)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestListSimilar_ExcludesSelfAndComputesSimilarity(t *testing.T) {
	svc, repo, _, _ := newSvc()
	commID := uint(5)
	catID := uint(6)
	repo.seed(&model.House{UserID: 1, ListingType: model.ListingTypeRent, CommunityID: &commID})
	// 列表返回 3 条：含自身 + 2 条相似（一条同小区，一条仅同分类）
	repo.listReturn = []model.House{
		mkHouse(1, model.House{CommunityID: &commID, RentPrice: 2000}),                       // 自身，应剔除
		mkHouse(2, model.House{CommunityID: &commID, RentPrice: 2100, Title: "同小区", CoverImage: "c2"}),
		mkHouse(3, model.House{CategoryID: &catID, RentPrice: 3000, Title: "同分类", CoverImage: "c3"}),
	}
	repo.listTotal = 3

	res, err := svc.ListSimilar(1, 5)
	require.NoError(t, err)
	require.Len(t, res, 2)
	// 自身被剔除
	assert.NotContains(t, []uint{res[0].HouseID, res[1].HouseID}, uint(1))
	// 租房取 rent_price
	assert.Equal(t, 2100.0, res[0].Price)
	// 同小区相似度 1.0
	assert.InDelta(t, 1.0, res[0].Similarity, 0.001)
	// 仅同分类相似度 0.5（无 community 命中，无 category 命中时回落 0.5）
	assert.InDelta(t, 0.5, res[1].Similarity, 0.001)
}

func TestListSimilar_SalePriceWhenSale(t *testing.T) {
	svc, repo, _, _ := newSvc()
	commID := uint(5)
	repo.seed(&model.House{UserID: 1, ListingType: model.ListingTypeSale, CommunityID: &commID})
	repo.listReturn = []model.House{mkHouse(2, model.House{CommunityID: &commID, SalePrice: 4000000, Title: "出售相似"})}
	res, err := svc.ListSimilar(1, 5)
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, 4000000.0, res[0].Price)
}

func TestListSimilar_UsesCategoryWhenNoCommunity(t *testing.T) {
	svc, repo, _, _ := newSvc()
	catID := uint(7)
	repo.seed(&model.House{UserID: 1, ListingType: model.ListingTypeRent, CategoryID: &catID})
	repo.listReturn = []model.House{mkHouse(2, model.House{CategoryID: &catID, RentPrice: 1500, Title: "同分类"})}
	res, err := svc.ListSimilar(1, 5)
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.InDelta(t, 0.8, res[0].Similarity, 0.001)
}

// ===== AdminList / AdminGetByID =====

func TestAdminList_Mapping(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.adminListReturn = []model.House{{Title: "M1"}}
	repo.adminListTotal = 1
	pg, list, err := svc.AdminList(&dto.HouseAdminListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pg.Total)
	require.Len(t, list, 1)
}

func TestAdminGetByID_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	_, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestAdminGetByID_WithImages(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	repo.images[1] = []model.HouseImage{{URL: "a.jpg"}}
	resp, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	require.Len(t, resp.Images, 1)
}

// ===== Audit =====

func TestAudit_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	err := svc.Audit(999, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestAudit_DuplicateApprovedRejected(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, AuditStatus: model.AuditApproved})
	err := svc.Audit(1, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrHouseAudited)
}

func TestAudit_RejectAllowsReaudit(t *testing.T) {
	svc, repo, _, _ := newSvc()
	// 已拒绝（非 pending）再次审核到通过：当前非 pending 且与目标相同才报错，不同则放行
	repo.seed(&model.House{UserID: 1, AuditStatus: model.AuditRejected})
	err := svc.Audit(1, model.AuditApproved, "通过")
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	assert.Equal(t, model.AuditApproved, repo.updateFieldsCalls[0].fields["audit_status"])
	assert.Equal(t, "通过", repo.updateFieldsCalls[0].fields["audit_reason"])
}

func TestAudit_PendingToApproved(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, AuditStatus: model.AuditPending})
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
}

// ===== AdminUpdateStatus =====

func TestAdminUpdateStatus_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	err := svc.AdminUpdateStatus(999, model.StatusPublished)
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestAdminUpdateStatus_PublishSetsPublishedAt(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1, Status: model.StatusDraft})
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	require.Contains(t, fields, "published_at")
}

func TestAdminUpdateStatus_OfflineNoPublishedAt(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	err := svc.AdminUpdateStatus(1, model.StatusOffline)
	require.NoError(t, err)
	assert.NotContains(t, repo.updateFieldsCalls[0].fields, "published_at")
}

// ===== UpdatePromotion =====

func TestUpdatePromotion_NotFound(t *testing.T) {
	svc, _, _, _ := newSvc()
	err := svc.UpdatePromotion(999, &dto.HousePromotionRequest{})
	assert.ErrorIs(t, err, ErrHouseNotFound)
}

func TestUpdatePromotion_SetsFields(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	err := svc.UpdatePromotion(1, &dto.HousePromotionRequest{
		PromotionLevel: 8, TrafficWeight: 3.5, Featured: true, Picked: true, Verified: true, RealHouseVerified: true,
	})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, 8, fields["promotion_level"])
	assert.Equal(t, 3.5, fields["traffic_weight"])
	assert.Equal(t, true, fields["featured"])
	assert.Equal(t, true, fields["real_house_verified"])
	require.Contains(t, fields, "real_house_verified_at")
}

func TestUpdatePromotion_NoVerifiedAtWhenFalse(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.seed(&model.House{UserID: 1})
	err := svc.UpdatePromotion(1, &dto.HousePromotionRequest{RealHouseVerified: false})
	require.NoError(t, err)
	assert.NotContains(t, repo.updateFieldsCalls[0].fields, "real_house_verified_at")
}

// ===== 批量操作 =====

func TestBatchAudit_Success(t *testing.T) {
	svc, repo, _, _ := newSvc()
	resp, err := svc.BatchAudit([]uint{1, 2, 3}, model.AuditApproved, "ok")
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 3, resp.Success)
	assert.Equal(t, 0, resp.Failed)
	require.Len(t, repo.batchAuditCalls, 1)
}

func TestBatchAudit_ErrorReturnsFailed(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.batchErr = errors.New("batch err")
	resp, err := svc.BatchAudit([]uint{1, 2}, model.AuditApproved, "")
	require.Error(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 0, resp.Success)
	assert.Equal(t, 2, resp.Failed)
	assert.Equal(t, []uint{1, 2}, resp.FailedIDs)
}

func TestBatchUpdateStatus_Success(t *testing.T) {
	svc, repo, _, _ := newSvc()
	resp, err := svc.BatchUpdateStatus([]uint{1, 2}, model.StatusOffline)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Success)
	require.Len(t, repo.batchUpdateStatusCalls, 1)
	assert.Equal(t, model.StatusOffline, repo.batchUpdateStatusCalls[0].status)
}

func TestBatchDelete_DeletesImages(t *testing.T) {
	svc, repo, _, _ := newSvc()
	resp, err := svc.BatchDelete([]uint{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Success)
	// 每条 id 都触发图片删除
	require.Len(t, repo.deleteImagesCalls, 3)
	require.Len(t, repo.batchDeleteCalls, 1)
}

func TestBatchDelete_ErrorReturnsFailed(t *testing.T) {
	svc, repo, _, _ := newSvc()
	repo.batchErr = errors.New("err")
	resp, err := svc.BatchDelete([]uint{1})
	require.Error(t, err)
	assert.Equal(t, 1, resp.Failed)
	assert.Equal(t, []uint{1}, resp.FailedIDs)
}
