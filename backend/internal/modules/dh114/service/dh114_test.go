// Package service 同城114商户黄页主表业务逻辑层单元测试 - 商户主表。
// 使用内存 mock 仓储覆盖：发布与默认值兜底（业务类型/来源类型）、发布即审核通过、
// 发布状态记录发布时间、JSONB 字段解析、更新字段构建与发布时间补齐、删除权限校验、
// 详情浏览量自增与收藏状态、列表/附近/搜索/高级搜索/我的发布的分页与错误传递、
// 收藏切换（创建/删除 + 计数增减）、收藏状态查询、收藏列表跳过缺失商户、
// 联系/分享计数、电话拨打记录（默认拨打类型/主叫信息/计数与最近拨打时间）、
// 浏览记录（默认访问类型/地区兜底）、管理端列表/详情、审核去重与状态联动（通过同步发布）、
// 批量审核结果统计、管理端状态更新（发布补齐发布时间）、推广配置（认证时间/精选/甄选/权重）等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockDh114Repo =====

type mockDh114Repo struct {
	byID   map[uint]*model.Dh114
	nextID uint

	createErr       error
	findErr         error
	updateFieldsErr error
	deleteErr       error
	listErr         error
	nearbyErr       error
	searchErr       error
	byUserErr       error
	incrViewErr     error
	incrFavErr      error
	decrFavErr      error

	updateFieldsCalls []struct {
		id     uint
		fields map[string]interface{}
	}
	deleteCalls   []uint
	incrViewCalls []uint
	incrFavCalls  []uint
	decrFavCalls  []uint

	listReturn      []model.Dh114
	listTotal       int64
	adminListReturn []model.Dh114
	adminListTotal  int64
	nearbyReturn    []model.Dh114
	nearbyTotal     int64
	searchReturn    []model.Dh114
	searchTotal     int64
	byUserReturn    []model.Dh114
	byUserTotal     int64
}

func newMockDh114Repo() *mockDh114Repo {
	return &mockDh114Repo{
		byID:   make(map[uint]*model.Dh114),
		nextID: 100,
	}
}

func (m *mockDh114Repo) Create(d *model.Dh114) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.nextID++
	d.ID = m.nextID
	m.byID[d.ID] = d
	return nil
}

func (m *mockDh114Repo) FindByID(id uint) (*model.Dh114, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	d, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *d
	cp.ID = id // 从 map key 回填 ID（结构体字面量无法直接设置嵌入字段 ID）
	return &cp, nil
}

func (m *mockDh114Repo) Update(d *model.Dh114) error {
	m.byID[d.ID] = d
	return nil
}

func (m *mockDh114Repo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, struct {
		id     uint
		fields map[string]interface{}
	}{id: id, fields: fields})
	d, ok := m.byID[id]
	if ok {
		if v, exists := fields["status"]; exists {
			if sv, ok := v.(int); ok {
				d.Status = sv
			}
		}
		if v, exists := fields["audit_status"]; exists {
			if sv, ok := v.(int); ok {
				d.AuditStatus = sv
			}
		}
		if v, exists := fields["published_at"]; exists {
			if tv, ok := v.(*time.Time); ok {
				d.PublishedAt = tv
			}
		}
		if v, exists := fields["featured"]; exists {
			if bv, ok := v.(bool); ok {
				d.Featured = bv
			}
		}
		if v, exists := fields["picked"]; exists {
			if bv, ok := v.(bool); ok {
				d.Picked = bv
			}
		}
		if v, exists := fields["verified"]; exists {
			if bv, ok := v.(bool); ok {
				d.Verified = bv
			}
		}
		if v, exists := fields["verified_at"]; exists {
			if tv, ok := v.(*time.Time); ok {
				d.VerifiedAt = tv
			}
		}
		if v, exists := fields["promotion_level"]; exists {
			if iv, ok := v.(int); ok {
				d.PromotionLevel = iv
			}
		}
		if v, exists := fields["traffic_weight"]; exists {
			if fv, ok := v.(float64); ok {
				d.TrafficWeight = fv
			}
		}
	}
	return nil
}

func (m *mockDh114Repo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.byID, id)
	return nil
}

func (m *mockDh114Repo) List(regionID uint, pagination *utils.Pagination, opts repository.Dh114ListOptions) ([]model.Dh114, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockDh114Repo) AdminList(pagination *utils.Pagination, opts repository.Dh114AdminListOptions) ([]model.Dh114, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockDh114Repo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.Dh114ListOptions) ([]model.Dh114, int64, error) {
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	return m.nearbyReturn, m.nearbyTotal, nil
}

func (m *mockDh114Repo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Dh114, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.searchReturn, m.searchTotal, nil
}

func (m *mockDh114Repo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error) {
	if m.byUserErr != nil {
		return nil, 0, m.byUserErr
	}
	return m.byUserReturn, m.byUserTotal, nil
}

func (m *mockDh114Repo) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error) {
	return nil, 0, nil
}

func (m *mockDh114Repo) IncrViewCount(id uint) error {
	if m.incrViewErr != nil {
		return m.incrViewErr
	}
	m.incrViewCalls = append(m.incrViewCalls, id)
	return nil
}

func (m *mockDh114Repo) IncrFavCount(id uint) error {
	if m.incrFavErr != nil {
		return m.incrFavErr
	}
	m.incrFavCalls = append(m.incrFavCalls, id)
	if d, ok := m.byID[id]; ok {
		d.FavCount++
	}
	return nil
}

func (m *mockDh114Repo) DecrFavCount(id uint) error {
	if m.decrFavErr != nil {
		return m.decrFavErr
	}
	m.decrFavCalls = append(m.decrFavCalls, id)
	if d, ok := m.byID[id]; ok && d.FavCount > 0 {
		d.FavCount--
	}
	return nil
}

func (m *mockDh114Repo) IncrContactCount(id uint) error { return nil }
func (m *mockDh114Repo) IncrShareCount(id uint) error   { return nil }
func (m *mockDh114Repo) IncrCallCount(id uint) error    { return nil }

func (m *mockDh114Repo) FavExists(userID, dh114ID uint) (bool, error) {
	return false, nil
}
func (m *mockDh114Repo) CreateFav(fav *model.Dh114Favorite) error    { return nil }
func (m *mockDh114Repo) DeleteFav(userID, dh114ID uint) error        { return nil }
func (m *mockDh114Repo) ListFavs(userID uint, page, pageSize int) ([]model.Dh114Favorite, int64, error) {
	return nil, 0, nil
}
func (m *mockDh114Repo) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	return make(map[uint]bool), nil
}

// ===== mockImageRepo =====

type mockImageRepo struct{}

func (m *mockImageRepo) Create(img *model.Dh114Image) error                  { return nil }
func (m *mockImageRepo) ListByDh114(dh114ID uint) ([]model.Dh114Image, error) { return nil, nil }
func (m *mockImageRepo) ReplaceImages(dh114ID uint, images []model.Dh114Image) error {
	return nil
}
func (m *mockImageRepo) Delete(id uint) error           { return nil }
func (m *mockImageRepo) DeleteByDh114(dh114ID uint) error { return nil }

// ===== mockVisitRepo =====

type mockVisitRepo struct {
	createErr error
	createCnt int
	lastVisit *model.Dh114Visit
	listErr   error
}

func (m *mockVisitRepo) Create(v *model.Dh114Visit) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createCnt++
	m.lastVisit = v
	return nil
}
func (m *mockVisitRepo) ListByDh114(dh114ID uint, page, pageSize int) ([]model.Dh114Visit, int64, error) {
	return nil, 0, m.listErr
}
func (m *mockVisitRepo) ListByUser(userID uint, page, pageSize int) ([]model.Dh114Visit, int64, error) {
	return nil, 0, m.listErr
}
func (m *mockVisitRepo) CountByDh114(dh114ID uint) (int64, error) { return 0, nil }

// ===== mockFavoriteRepo =====

type mockFavoriteRepo struct {
	existsReturn bool
	existsErr    error
	createErr    error
	deleteErr    error
	listReturn   []model.Dh114Favorite
	listTotal    int64
	listErr      error

	createCalls []model.Dh114Favorite
	deleteCalls []struct {
		userID       uint
		dh114ID      uint
		favoriteType string
	}
}

func (m *mockFavoriteRepo) Create(fav *model.Dh114Favorite) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createCalls = append(m.createCalls, *fav)
	return nil
}
func (m *mockFavoriteRepo) FindByID(id uint) (*model.Dh114Favorite, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockFavoriteRepo) Update(id uint, fields map[string]interface{}) error { return nil }
func (m *mockFavoriteRepo) Delete(id uint) error                                { return nil }
func (m *mockFavoriteRepo) DeleteByUserAndTarget(userID uint, dh114ID uint, favoriteType string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCalls = append(m.deleteCalls, struct {
		userID       uint
		dh114ID      uint
		favoriteType string
	}{userID: userID, dh114ID: dh114ID, favoriteType: favoriteType})
	return nil
}
func (m *mockFavoriteRepo) Exists(userID uint, dh114ID uint, favoriteType string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	return m.existsReturn, nil
}
func (m *mockFavoriteRepo) List(query repository.FavoriteListQuery, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}
func (m *mockFavoriteRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return m.List(repository.FavoriteListQuery{UserID: userID}, pagination)
}
func (m *mockFavoriteRepo) ListByType(userID uint, favoriteType string, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return m.List(repository.FavoriteListQuery{UserID: userID, FavoriteType: favoriteType}, pagination)
}
func (m *mockFavoriteRepo) ListByGroup(userID uint, groupID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return m.List(repository.FavoriteListQuery{UserID: userID, GroupID: groupID}, pagination)
}
func (m *mockFavoriteRepo) HasFavedBatch(userID uint, ids []uint, favoriteType string) (map[uint]bool, error) {
	return make(map[uint]bool), nil
}

// ===== mockPhoneCallRepo =====

type mockPhoneCallRepo struct {
	createErr error
	createCnt int
	lastCall  *model.Dh114PhoneCall
}

func (m *mockPhoneCallRepo) Create(c *model.Dh114PhoneCall) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createCnt++
	m.lastCall = c
	return nil
}
func (m *mockPhoneCallRepo) FindByID(id uint) (*model.Dh114PhoneCall, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockPhoneCallRepo) List(query repository.PhoneCallListQuery, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	return nil, 0, nil
}
func (m *mockPhoneCallRepo) ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	return nil, 0, nil
}
func (m *mockPhoneCallRepo) ListByCaller(callerID uint, pagination *utils.Pagination) ([]model.Dh114PhoneCall, int64, error) {
	return nil, 0, nil
}
func (m *mockPhoneCallRepo) CountByDh114(dh114ID uint) (int64, error)          { return 0, nil }
func (m *mockPhoneCallRepo) CountTodayByDh114(dh114ID uint) (int64, error)     { return 0, nil }

// ===== 测试辅助 =====

func newDh114ServiceWithMocks() (Dh114Service, *mockDh114Repo, *mockVisitRepo, *mockFavoriteRepo, *mockPhoneCallRepo) {
	repo := newMockDh114Repo()
	visitRepo := &mockVisitRepo{}
	favRepo := &mockFavoriteRepo{}
	callRepo := &mockPhoneCallRepo{}
	svc := NewDh114Service(repo, &mockImageRepo{}, visitRepo, favRepo, callRepo)
	return svc, repo, visitRepo, favRepo, callRepo
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
func floatPtr(f float64) *float64 { return &f }
func uintPtr(u uint) *uint    { return &u }

// favWithID 设置收藏记录的嵌入 ID 字段（结构体字面量无法直接设置嵌入字段 ID）
func favWithID(f *model.Dh114Favorite, id uint) *model.Dh114Favorite { f.ID = id; return f }

// dh114WithID 构造带 ID 的 Dh114（ID 来自嵌入 RegionBaseModel.BaseModel，
// 无法在复合字面量中直接设置，需通过字段赋值）。
func dh114WithID(id uint, d model.Dh114) model.Dh114 { d.ID = id; return d }

// ==================== Create ====================

func TestDh114Create_DefaultValuesFallback(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	req := &dto.CreateDh114Request{
		Title: "测试商户",
		Phone: "13800000000",
		Status: 1,
	}
	info, err := svc.Create(1, 10, "张三", "13800000000", "avatar.png", req)
	require.NoError(t, err)
	require.NotNil(t, info)

	d := repo.byID[info.ID]
	require.NotNil(t, d)
	assert.Equal(t, model.BusinessTypeOther, d.BusinessType, "空业务类型应兜底为 other")
	assert.Equal(t, model.SourceTypePersonal, d.SourceType, "空来源类型应兜底为 personal")
	assert.Equal(t, model.AuditApproved, d.AuditStatus, "MVP 发布即审核通过")
	assert.NotNil(t, d.PublishedAt, "status=1 应记录发布时间")
	assert.Equal(t, uint(1), d.RegionID)
	assert.Equal(t, uint(10), d.UserID)
	assert.Equal(t, "张三", d.UserName)
	assert.Equal(t, "13800000000", d.UserPhone)
	assert.Equal(t, "avatar.png", d.UserAvatar)
}

func TestDh114Create_DraftStatusNoPublishedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	req := &dto.CreateDh114Request{
		Title:  "草稿商户",
		Phone:  "13800000000",
		Status: 0,
	}
	info, err := svc.Create(1, 10, "张三", "13800000000", "", req)
	require.NoError(t, err)
	d := repo.byID[info.ID]
	require.NotNil(t, d)
	assert.Nil(t, d.PublishedAt, "草稿状态不应记录发布时间")
}

func TestDh114Create_PreservedBusinessTypeAndSource(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	req := &dto.CreateDh114Request{
		Title:        "餐饮商户",
		Phone:        "13800000000",
		BusinessType: model.BusinessTypeRestaurant,
		SourceType:   model.SourceTypeChain,
		Status:       0,
	}
	_, err := svc.Create(1, 10, "李四", "13900000000", "", req)
	require.NoError(t, err)
	// 注意：SourceType 传入 chain（非 SourceType 常量，仅作字符串透传）
	d := repo.byID[101]
	require.NotNil(t, d)
	assert.Equal(t, model.BusinessTypeRestaurant, d.BusinessType)
	assert.Equal(t, "chain", d.SourceType)
}

func TestDh114Create_JSONBFieldsParsed(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	req := &dto.CreateDh114Request{
		Title:         "带图商户",
		Phone:         "13800000000",
		Status:        0,
		Images:        []model.BusinessImageItem{{URL: "a.png", Sort: 0}},
		Tags:          []model.BusinessTagItem{{Text: "热推", Color: "#f00"}},
		BusinessHours: []model.BusinessHourItem{{Weekday: 1, OpenTime: "09:00", CloseTime: "18:00"}},
		Features:      []model.FacilityItem{{Code: "wifi", Name: "WiFi", Has: true}},
	}
	_, err := svc.Create(1, 10, "王五", "13800000000", "", req)
	require.NoError(t, err)
	d := repo.byID[101]
	require.NotNil(t, d)
	assert.NotEmpty(t, d.Images)
	assert.NotEmpty(t, d.Tags)
	assert.NotEmpty(t, d.BusinessHours)
	assert.NotEmpty(t, d.Features)

	// 验证 JSONB 可解析回结构
	var imgs []model.BusinessImageItem
	require.NoError(t, d.Images.Parse(&imgs))
	require.Len(t, imgs, 1)
	assert.Equal(t, "a.png", imgs[0].URL)
}

func TestDh114Create_JSONBParseFailureIgnored(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	// 传入无法 Marshal 的对象：channel
	req := &dto.CreateDh114Request{
		Title:  "异常JSONB",
		Phone:  "13800000000",
		Status: 0,
		Images: make(chan int),
	}
	_, err := svc.Create(1, 10, "赵六", "13800000000", "", req)
	require.NoError(t, err, "JSONB 解析失败应被忽略，不阻断创建")
	d := repo.byID[101]
	require.NotNil(t, d)
	assert.Nil(t, d.Images, "解析失败的 JSONB 字段应保持 nil")
}

func TestDh114Create_CreateErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.createErr = errors.New("db down")
	req := &dto.CreateDh114Request{Title: "x", Phone: "13800000000", Status: 0}
	_, err := svc.Create(1, 10, "x", "x", "", req)
	require.Error(t, err)
	assert.Equal(t, "db down", err.Error())
}

func TestDh114Create_InfoMapping(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	req := &dto.CreateDh114Request{
		Title:        "映射测试",
		Content:      "简介内容",
		CoverImage:   "cover.png",
		Phone:        "13800000000",
		AltPhone:     "021-1234567",
		Website:      "https://example.com",
		Wechat:       "wxid_abc",
		City:         "五常",
		District:     "五常市",
		BusinessDistrict: "中心商圈",
		Address:      "某街某号",
		Latitude:     44.9,
		Longitude:    127.1,
		VideoURL:     "video.mp4",
		VideoCover:   "vc.png",
		VRURL:        "vr.html",
		Status:       1,
		BusinessType: model.BusinessTypeRetail,
		SourceType:   model.SourceTypeMerchant,
	}
	info, err := svc.Create(1, 10, "孙七", "13800000000", "av.png", req)
	require.NoError(t, err)
	assert.Equal(t, "映射测试", info.Title)
	assert.Equal(t, "简介内容", info.Content)
	assert.Equal(t, "cover.png", info.CoverImage)
	assert.Equal(t, uint(10), info.UserID)
	assert.Equal(t, "孙七", info.UserName)
	assert.Equal(t, model.StatusPublished, info.Status)
	assert.Equal(t, "已发布", info.StatusText)
	assert.Equal(t, model.AuditApproved, info.AuditStatus)
	assert.Equal(t, "通过", info.AuditStatusText)
	assert.Equal(t, model.BusinessTypeRetail, info.BusinessType)
	assert.Equal(t, model.SourceTypeMerchant, info.SourceType)
	assert.Equal(t, "五常", info.City)
	assert.Equal(t, "中心商圈", info.BusinessDistrict)
	assert.Equal(t, uint(1), info.RegionID)
	assert.NotNil(t, info.PublishedAt)
}

// ==================== toDh114Info ====================

func TestToDh114Info_StatusTextMapping(t *testing.T) {
	cases := []struct {
		status int
		text   string
	}{
		{model.StatusDraft, "草稿"},
		{model.StatusPublished, "已发布"},
		{model.StatusOffline, "已下架"},
		{model.StatusExpired, "已过期"},
		{model.StatusDeleted, "已删除"},
		{99, ""},
	}
	for _, c := range cases {
		d := &model.Dh114{Status: c.status}
		info := toDh114Info(d)
		assert.Equal(t, c.text, info.StatusText, "status=%d", c.status)
	}
}

func TestToDh114Info_AuditStatusTextMapping(t *testing.T) {
	cases := []struct {
		status int
		text   string
	}{
		{model.AuditPending, "待审"},
		{model.AuditApproved, "通过"},
		{model.AuditRejected, "拒绝"},
		{99, ""},
	}
	for _, c := range cases {
		d := &model.Dh114{AuditStatus: c.status}
		info := toDh114Info(d)
		assert.Equal(t, c.text, info.AuditStatusText, "audit_status=%d", c.status)
	}
}

func TestToDh114Info_NilJSONBFields(t *testing.T) {
	d := &model.Dh114{Title: "x"}
	info := toDh114Info(d)
	assert.Nil(t, info.Images)
	assert.Nil(t, info.Tags)
	assert.Nil(t, info.BusinessHours)
	assert.Nil(t, info.Features)
}

func TestToDh114InfoWithFav(t *testing.T) {
	d := &model.Dh114{Title: "x"}
	info := toDh114InfoWithFav(d, true)
	assert.True(t, info.HasFaved)
	info2 := toDh114InfoWithFav(d, false)
	assert.False(t, info2.HasFaved)
}

// ==================== Update ====================

func TestDh114Update_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	err := svc.Update(999, 10, &dto.UpdateDh114Request{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114Update_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	err := svc.Update(1, 10, &dto.UpdateDh114Request{Title: strPtr("x")})
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114Update_NoPermission(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	err := svc.Update(1, 99, &dto.UpdateDh114Request{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrDh114NoPermission)
}

func TestDh114Update_EmptyFieldsNoOp(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	err := svc.Update(1, 10, &dto.UpdateDh114Request{})
	require.NoError(t, err)
	assert.Empty(t, repo.updateFieldsCalls, "无字段更新不应调用 UpdateFields")
}

func TestDh114Update_FieldsBuilding(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	req := &dto.UpdateDh114Request{
		Title:           strPtr("新标题"),
		Content:         strPtr("新简介"),
		CoverImage:      strPtr("new.png"),
		CategoryID:      uintPtr(5),
		CategoryName:    strPtr("美食"),
		BusinessType:    strPtr(model.BusinessTypeRestaurant),
		SourceType:      strPtr(model.SourceTypeMerchant),
		Phone:           strPtr("13900000000"),
		AltPhone:        strPtr("021-999"),
		Website:         strPtr("https://new.com"),
		Wechat:          strPtr("newwx"),
		City:            strPtr("五常"),
		District:        strPtr("五常市"),
		BusinessDistrict: strPtr("新区"),
		Address:         strPtr("新址"),
		Latitude:        floatPtr(45.0),
		Longitude:       floatPtr(127.5),
		VideoURL:        strPtr("v2.mp4"),
		VideoCover:      strPtr("vc2.png"),
		VRURL:           strPtr("vr2.html"),
	}
	err := svc.Update(1, 10, req)
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, "新标题", fields["title"])
	assert.Equal(t, "新简介", fields["content"])
	assert.Equal(t, "new.png", fields["cover_image"])
	assert.Equal(t, uint(5), fields["category_id"])
	assert.Equal(t, "美食", fields["category_name"])
	assert.Equal(t, model.BusinessTypeRestaurant, fields["business_type"])
	assert.Equal(t, model.SourceTypeMerchant, fields["source_type"])
	assert.Equal(t, "13900000000", fields["phone"])
	assert.Equal(t, "021-999", fields["alt_phone"])
	assert.Equal(t, "https://new.com", fields["website"])
	assert.Equal(t, "newwx", fields["wechat"])
	assert.Equal(t, "五常", fields["city"])
	assert.Equal(t, "五常市", fields["district"])
	assert.Equal(t, "新区", fields["business_district"])
	assert.Equal(t, "新址", fields["address"])
	assert.Equal(t, 45.0, fields["latitude"])
	assert.Equal(t, 127.5, fields["longitude"])
	assert.Equal(t, "v2.mp4", fields["video_url"])
	assert.Equal(t, "vc2.png", fields["video_cover"])
	assert.Equal(t, "vr2.html", fields["vr_url"])
}

func TestDh114Update_PublishedAtSetOnPublish(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10, Status: model.StatusDraft}
	status := model.StatusPublished
	err := svc.Update(1, 10, &dto.UpdateDh114Request{Status: &status})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	_, hasPub := fields["published_at"]
	assert.True(t, hasPub, "首次发布应补齐 published_at")
}

func TestDh114Update_PublishedAtNotResetIfAlreadySet(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	old := time.Now().Add(-time.Hour)
	repo.byID[1] = &model.Dh114{UserID: 10, Status: model.StatusOffline, PublishedAt: &old}
	status := model.StatusPublished
	err := svc.Update(1, 10, &dto.UpdateDh114Request{Status: &status})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "已有 published_at 不应重置")
}

func TestDh114Update_OfflineNoPublishedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10, Status: model.StatusPublished}
	status := model.StatusOffline
	err := svc.Update(1, 10, &dto.UpdateDh114Request{Status: &status})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "下架不应设置 published_at")
}

func TestDh114Update_JSONBFields(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	req := &dto.UpdateDh114Request{
		Images:        []model.BusinessImageItem{{URL: "u1.png"}},
		Tags:          []model.BusinessTagItem{{Text: "t1"}},
		BusinessHours: []model.BusinessHourItem{{Weekday: 2}},
		Features:      []model.FacilityItem{{Code: "park", Name: "停车"}},
	}
	err := svc.Update(1, 10, req)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.NotNil(t, fields["images"])
	assert.NotNil(t, fields["tags"])
	assert.NotNil(t, fields["business_hours"])
	assert.NotNil(t, fields["features"])
}

func TestDh114Update_JSONBParseFailureSkipped(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	req := &dto.UpdateDh114Request{
		Title:  strPtr("ok"),
		Images: make(chan int),
	}
	err := svc.Update(1, 10, req)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasImg := fields["images"]
	assert.False(t, hasImg, "JSONB 解析失败应跳过该字段")
	_, hasTitle := fields["title"]
	assert.True(t, hasTitle, "其他字段应正常写入")
}

func TestDh114Update_UpdateFieldsError(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.Update(1, 10, &dto.UpdateDh114Request{Title: strPtr("x")})
	require.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
}

// ==================== Delete ====================

func TestDh114Delete_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	err := svc.Delete(999, 10)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114Delete_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	err := svc.Delete(1, 10)
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114Delete_NoPermission(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	err := svc.Delete(1, 99)
	assert.ErrorIs(t, err, ErrDh114NoPermission)
}

func TestDh114Delete_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	err := svc.Delete(1, 10)
	require.NoError(t, err)
	assert.Contains(t, repo.deleteCalls, uint(1))
}

func TestDh114Delete_DeleteError(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	repo.deleteErr = errors.New("del failed")
	err := svc.Delete(1, 10)
	require.Error(t, err)
	assert.Equal(t, "del failed", err.Error())
}

// ==================== GetByID ====================

func TestDh114GetByID_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.GetByID(999, 10)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114GetByID_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	_, err := svc.GetByID(1, 10)
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114GetByID_IncrViewAndNoFavForGuest(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10, Title: "商户A"}
	info, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	assert.Equal(t, "商户A", info.Title)
	assert.False(t, info.HasFaved, "游客无收藏状态")
	assert.Contains(t, repo.incrViewCalls, uint(1), "应自增浏览量")
	_ = favRepo // 游客不查收藏
}

func TestDh114GetByID_HasFavedWhenUserLoggedIn(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	favRepo.existsReturn = true
	info, err := svc.GetByID(1, 20)
	require.NoError(t, err)
	assert.True(t, info.HasFaved)
}

func TestDh114GetByID_FavExistsErrorIgnored(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	favRepo.existsErr = errors.New("redis down")
	info, err := svc.GetByID(1, 20)
	require.NoError(t, err, "收藏查询错误应被忽略不阻断详情")
	assert.False(t, info.HasFaved)
}

func TestDh114GetByID_IncrViewErrorIgnored(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{UserID: 10}
	repo.incrViewErr = errors.New("redis down")
	info, err := svc.GetByID(1, 0)
	require.NoError(t, err, "浏览量自增失败应被忽略")
	assert.NotNil(t, info)
}

// ==================== List ====================

func TestDh114List_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "A"}), dh114WithID(2, model.Dh114{Title: "B"})}
	repo.listTotal = 2
	pag, list, err := svc.List(1, &dto.Dh114ListRequest{})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), pag.Total)
	assert.Equal(t, "A", list[0].Title)
}

func TestDh114List_ErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listErr = errors.New("list err")
	_, _, err := svc.List(1, &dto.Dh114ListRequest{})
	require.Error(t, err)
	assert.Equal(t, "list err", err.Error())
}

func TestDh114List_EmptyResult(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listReturn = nil
	repo.listTotal = 0
	pag, list, err := svc.List(1, &dto.Dh114ListRequest{Sort: "rating_desc"})
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, int64(0), pag.Total)
}

func TestDh114List_PaginationFromRequest(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listReturn = []model.Dh114{dh114WithID(1, model.Dh114{})}
	repo.listTotal = 100
	pag, _, err := svc.List(1, &dto.Dh114ListRequest{Pagination: utils.Pagination{Page: 3, PageSize: 20}})
	require.NoError(t, err)
	assert.Equal(t, 3, pag.Page)
	assert.Equal(t, 20, pag.PageSize)
	assert.Equal(t, int64(100), pag.Total)
}

// ==================== ListNearby ====================

func TestDh114ListNearby_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.nearbyReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "附近A", Distance: 1.2})}
	repo.nearbyTotal = 1
	pag, list, err := svc.ListNearby(1, &dto.Dh114NearbyRequest{Latitude: 44.9, Longitude: 127.1, RadiusKm: 5})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pag.Total)
}

func TestDh114ListNearby_ErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.nearbyErr = errors.New("nearby err")
	_, _, err := svc.ListNearby(1, &dto.Dh114NearbyRequest{Latitude: 44.9, Longitude: 127.1})
	require.Error(t, err)
	assert.Equal(t, "nearby err", err.Error())
}

// ==================== Search ====================

func TestDh114Search_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.searchReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "结果"})}
	repo.searchTotal = 1
	pag, list, err := svc.Search(1, &dto.Dh114SearchRequest{Keyword: "测试"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pag.Total)
}

func TestDh114Search_ErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.searchErr = errors.New("search err")
	_, _, err := svc.Search(1, &dto.Dh114SearchRequest{Keyword: "x"})
	require.Error(t, err)
	assert.Equal(t, "search err", err.Error())
}

// ==================== ListMine ====================

func TestDh114ListMine_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byUserReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "我的"})}
	repo.byUserTotal = 1
	pag, list, err := svc.ListMine(10, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pag.Total)
}

func TestDh114ListMine_ErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byUserErr = errors.New("mine err")
	_, _, err := svc.ListMine(10, 1, 10)
	require.Error(t, err)
	assert.Equal(t, "mine err", err.Error())
}

// ==================== AdvancedSearch ====================

func TestDh114AdvancedSearch_NearbyPath(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.nearbyReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "附近"})}
	repo.nearbyTotal = 1
	req := &dto.AdvancedSearchRequest{
		Latitude:  44.9,
		Longitude: 127.1,
		RadiusKm:  3,
		Keyword:   "餐",
	}
	pag, list, err := svc.AdvancedSearch(1, req)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pag.Total)
}

func TestDh114AdvancedSearch_ListPath(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listReturn = []model.Dh114{dh114WithID(2, model.Dh114{Title: "列表"})}
	repo.listTotal = 1
	req := &dto.AdvancedSearchRequest{
		Keyword:      "餐",
		CategoryID:   5,
		BusinessType: model.BusinessTypeRestaurant,
		City:         "五常",
	}
	pag, list, err := svc.AdvancedSearch(1, req)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pag.Total)
}

func TestDh114AdvancedSearch_NearbyErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.nearbyErr = errors.New("nearby err")
	req := &dto.AdvancedSearchRequest{Latitude: 44.9, Longitude: 127.1, RadiusKm: 3}
	_, _, err := svc.AdvancedSearch(1, req)
	require.Error(t, err)
	assert.Equal(t, "nearby err", err.Error())
}

func TestDh114AdvancedSearch_ListErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listErr = errors.New("list err")
	req := &dto.AdvancedSearchRequest{Keyword: "x"}
	_, _, err := svc.AdvancedSearch(1, req)
	require.Error(t, err)
	assert.Equal(t, "list err", err.Error())
}

// ==================== Fav ====================

func TestDh114Fav_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.Fav(10, 999)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114Fav_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	_, err := svc.Fav(10, 1)
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114Fav_CreateNew(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	resp, err := svc.Fav(10, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 6, resp.FavCount, "收藏数应 +1")
	require.Len(t, favRepo.createCalls, 1)
	assert.Equal(t, uint(10), favRepo.createCalls[0].UserID)
	assert.Equal(t, uint(1), favRepo.createCalls[0].Dh114ID)
	assert.Equal(t, model.FavoriteTypeBusiness, favRepo.createCalls[0].FavoriteType)
	assert.Contains(t, repo.incrFavCalls, uint(1))
}

func TestDh114Fav_AlreadyExistsNoCreate(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	favRepo.existsReturn = true
	resp, err := svc.Fav(10, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 5, resp.FavCount, "已收藏不应重复计数")
	assert.Empty(t, favRepo.createCalls, "已收藏不应再创建")
	assert.Empty(t, repo.incrFavCalls, "已收藏不应自增计数")
}

func TestDh114Fav_ExistsError(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	favRepo.existsErr = errors.New("redis down")
	_, err := svc.Fav(10, 1)
	require.Error(t, err)
	assert.Equal(t, "redis down", err.Error())
}

func TestDh114Fav_CreateFavError(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	favRepo.createErr = errors.New("create fav failed")
	_, err := svc.Fav(10, 1)
	require.Error(t, err)
	assert.Equal(t, "create fav failed", err.Error())
}

func TestDh114Fav_IncrFavErrorIgnored(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	repo.incrFavErr = errors.New("incr err")
	resp, err := svc.Fav(10, 1)
	require.NoError(t, err, "计数自增失败应被忽略")
	assert.Equal(t, 6, resp.FavCount, "响应仍按 +1 计算")
}

// ==================== Unfav ====================

func TestDh114Unfav_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.Unfav(10, 999)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114Unfav_DeleteExisting(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	resp, err := svc.Unfav(10, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 4, resp.FavCount, "收藏数应 -1")
	require.Len(t, favRepo.deleteCalls, 1)
	assert.Equal(t, model.FavoriteTypeBusiness, favRepo.deleteCalls[0].favoriteType)
	assert.Contains(t, repo.decrFavCalls, uint(1))
}

func TestDh114Unfav_ZeroFavCountNotNegative(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 0}
	resp, err := svc.Unfav(10, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 0, resp.FavCount, "收藏数不应为负")
}

func TestDh114Unfav_DeleteError(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	favRepo.deleteErr = errors.New("del fav err")
	_, err := svc.Unfav(10, 1)
	require.Error(t, err)
	assert.Equal(t, "del fav err", err.Error())
}

func TestDh114Unfav_DecrFavErrorIgnored(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 5}
	repo.decrFavErr = errors.New("decr err")
	resp, err := svc.Unfav(10, 1)
	require.NoError(t, err, "计数自减失败应被忽略")
	assert.Equal(t, 4, resp.FavCount)
}

// ==================== FavStatus ====================

func TestDh114FavStatus_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.FavStatus(10, 999)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114FavStatus_GuestFalse(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 3}
	resp, err := svc.FavStatus(0, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 3, resp.FavCount)
}

func TestDh114FavStatus_UserTrue(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 3}
	favRepo.existsReturn = true
	resp, err := svc.FavStatus(10, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 3, resp.FavCount)
}

func TestDh114FavStatus_ExistsErrorIgnored(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{FavCount: 3}
	favRepo.existsErr = errors.New("redis down")
	resp, err := svc.FavStatus(10, 1)
	require.NoError(t, err, "收藏查询错误应被忽略")
	assert.False(t, resp.HasFaved)
}

// ==================== ListFavs ====================

func TestDh114ListFavs_Success(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Title: "A"}
	repo.byID[2] = &model.Dh114{Title: "B"}
	favRepo.listReturn = []model.Dh114Favorite{
		*favWithID(&model.Dh114Favorite{UserID: 10, Dh114ID: 1, FavoriteType: model.FavoriteTypeBusiness}, 100),
		*favWithID(&model.Dh114Favorite{UserID: 10, Dh114ID: 2, FavoriteType: model.FavoriteTypeBusiness}, 101),
	}
	favRepo.listTotal = 2
	pag, list, err := svc.ListFavs(10, &dto.FavoriteListRequest{})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), pag.Total)
	assert.True(t, list[0].HasFaved)
	assert.Equal(t, uint(100), list[0].ID, "应返回收藏记录 ID")
}

func TestDh114ListFavs_DefaultFavoriteType(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{}
	favRepo.listReturn = []model.Dh114Favorite{*favWithID(&model.Dh114Favorite{Dh114ID: 1}, 100)}
	favRepo.listTotal = 1
	_, list, err := svc.ListFavs(10, &dto.FavoriteListRequest{FavoriteType: ""})
	require.NoError(t, err)
	require.Len(t, list, 1)
}

func TestDh114ListFavs_SkipMissingJobs(t *testing.T) {
	svc, repo, _, favRepo, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Title: "A"}
	// dh114_id=2 不存在于 repo，应被跳过
	favRepo.listReturn = []model.Dh114Favorite{
		*favWithID(&model.Dh114Favorite{Dh114ID: 1}, 100),
		*favWithID(&model.Dh114Favorite{Dh114ID: 2}, 101),
	}
	favRepo.listTotal = 2
	_, list, err := svc.ListFavs(10, &dto.FavoriteListRequest{})
	require.NoError(t, err)
	require.Len(t, list, 1, "缺失的商户应被跳过")
	assert.Equal(t, "A", list[0].Title)
}

func TestDh114ListFavs_ErrorPropagation(t *testing.T) {
	svc, _, _, favRepo, _ := newDh114ServiceWithMocks()
	favRepo.listErr = errors.New("list fav err")
	_, _, err := svc.ListFavs(10, &dto.FavoriteListRequest{})
	require.Error(t, err)
	assert.Equal(t, "list fav err", err.Error())
}

func TestDh114ListFavs_EmptyResult(t *testing.T) {
	svc, _, _, favRepo, _ := newDh114ServiceWithMocks()
	favRepo.listReturn = nil
	favRepo.listTotal = 0
	pag, list, err := svc.ListFavs(10, &dto.FavoriteListRequest{})
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Equal(t, int64(0), pag.Total)
}

// ==================== IncrContact / IncrShare ====================

func TestDh114IncrContact_DelegatesToRepo(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	// IncrContact 直接调用 repo.IncrContactCount（mock 返回 nil）
	err := svc.IncrContact(1)
	require.NoError(t, err)
}

func TestDh114IncrShare_DelegatesToRepo(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	err := svc.IncrShare(1)
	require.NoError(t, err)
}

// ==================== RecordCall ====================

func TestDh114RecordCall_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.RecordCall(10, 999, "13800000000", &dto.PhoneCallRequest{Dh114ID: 999}, "127.0.0.1", "UA")
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114RecordCall_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	_, err := svc.RecordCall(10, 1, "13800000000", &dto.PhoneCallRequest{Dh114ID: 1}, "127.0.0.1", "UA")
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114RecordCall_SuccessLoggedIn(t *testing.T) {
	svc, repo, _, _, callRepo := newDh114ServiceWithMocks()
	d := &model.Dh114{UserID: 10, UserName: "张三", UserPhone: "13800000000", CallCount: 3}
	d.RegionID = 2
	repo.byID[1] = d
	resp, err := svc.RecordCall(10, 1, "13800000000", &dto.PhoneCallRequest{
		Dh114ID:  1,
		CallType: model.CallTypeCall,
		Device:   "app",
	}, "127.0.0.1", "Mozilla")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.CallNo)
	assert.Equal(t, "13800000000", resp.Phone)
	assert.Equal(t, 4, resp.CallCount, "拨打计数应 +1")
	require.Equal(t, 1, callRepo.createCnt)
	c := callRepo.lastCall
	require.NotNil(t, c)
	assert.Equal(t, model.CallTypeCall, c.CallType)
	assert.Equal(t, "app", c.Device)
	assert.Equal(t, "127.0.0.1", c.IP)
	assert.Equal(t, "Mozilla", c.UserAgent)
	assert.Equal(t, uint(10), c.CallerID)
	assert.Equal(t, "张三", c.CallerName, "登录用户应回填主叫昵称")
	assert.Equal(t, "13800000000", c.CallerPhone, "登录用户应回填主叫号码")
	assert.Equal(t, uint(2), c.RegionID, "应继承商户地区 ID")
	assert.Equal(t, model.CallStatusSuccess, c.Status)
	// 应同时更新 last_call_at
	require.Len(t, repo.updateFieldsCalls, 1)
	_, hasLastCall := repo.updateFieldsCalls[0].fields["last_call_at"]
	assert.True(t, hasLastCall, "应更新最近拨打时间")
}

func TestDh114RecordCall_DefaultCallType(t *testing.T) {
	svc, repo, _, _, callRepo := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{CallCount: 0}
	_, err := svc.RecordCall(0, 1, "13800000000", &dto.PhoneCallRequest{Dh114ID: 1}, "127.0.0.1", "UA")
	require.NoError(t, err)
	c := callRepo.lastCall
	require.NotNil(t, c)
	assert.Equal(t, model.CallTypeClick, c.CallType, "空拨打类型应兜底为 click")
	assert.Equal(t, uint(0), c.CallerID, "游客主叫 ID 为 0")
	assert.Empty(t, c.CallerName, "游客无主叫昵称")
	assert.Empty(t, c.CallerPhone, "游客无主叫号码")
}

func TestDh114RecordCall_CreateError(t *testing.T) {
	svc, repo, _, _, callRepo := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{CallCount: 0}
	callRepo.createErr = errors.New("create call err")
	_, err := svc.RecordCall(0, 1, "13800000000", &dto.PhoneCallRequest{Dh114ID: 1}, "127.0.0.1", "UA")
	require.Error(t, err)
	assert.Equal(t, "create call err", err.Error())
}

// ==================== RecordView ====================

func TestDh114RecordView_Success(t *testing.T) {
	svc, _, visitRepo, _, _ := newDh114ServiceWithMocks()
	err := svc.RecordView(10, "127.0.0.1", &dto.Dh114ViewRequest{
		Dh114ID:   1,
		VisitType: "business",
		Device:    "pc",
		Source:    "search",
		Duration:  30,
	})
	require.NoError(t, err)
	require.Equal(t, 1, visitRepo.createCnt)
	v := visitRepo.lastVisit
	require.NotNil(t, v)
	assert.Equal(t, uint(10), v.UserID)
	assert.Equal(t, uint(1), v.Dh114ID)
	assert.Equal(t, "business", v.VisitType)
	assert.Equal(t, "pc", v.Device)
	assert.Equal(t, "search", v.Source)
	assert.Equal(t, 30, v.Duration)
	assert.Equal(t, "127.0.0.1", v.IP)
}

func TestDh114RecordView_DefaultVisitType(t *testing.T) {
	svc, _, visitRepo, _, _ := newDh114ServiceWithMocks()
	err := svc.RecordView(0, "127.0.0.1", &dto.Dh114ViewRequest{Dh114ID: 1})
	require.NoError(t, err)
	v := visitRepo.lastVisit
	require.NotNil(t, v)
	assert.Equal(t, "business", v.VisitType, "空访问类型应兜底为 business")
	assert.Equal(t, uint(1), v.RegionID, "默认地区 ID 应为 1")
}

func TestDh114RecordView_CreateError(t *testing.T) {
	svc, _, visitRepo, _, _ := newDh114ServiceWithMocks()
	visitRepo.createErr = errors.New("create visit err")
	err := svc.RecordView(10, "127.0.0.1", &dto.Dh114ViewRequest{Dh114ID: 1})
	require.Error(t, err)
	assert.Equal(t, "create visit err", err.Error())
}

// ==================== AdminList ====================

func TestDh114AdminList_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.adminListReturn = []model.Dh114{dh114WithID(1, model.Dh114{Title: "A"}), dh114WithID(2, model.Dh114{Title: "B"})}
	repo.adminListTotal = 2
	pag, list, err := svc.AdminList(&dto.Dh114AdminListRequest{RegionID: 1, UserID: 10})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), pag.Total)
}

func TestDh114AdminList_ErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.listErr = errors.New("admin list err")
	_, _, err := svc.AdminList(&dto.Dh114AdminListRequest{})
	require.Error(t, err)
	assert.Equal(t, "admin list err", err.Error())
}

// ==================== AdminGetByID ====================

func TestDh114AdminGetByID_Success(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Title: "商户A"}
	info, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, "商户A", info.Title)
	assert.False(t, info.HasFaved, "管理端详情无收藏状态")
}

func TestDh114AdminGetByID_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	_, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114AdminGetByID_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	_, err := svc.AdminGetByID(1)
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

// ==================== Audit ====================

func TestDh114Audit_NotFound(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	err := svc.Audit(999, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrDh114NotFound)
}

func TestDh114Audit_FindErrorPropagation(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.findErr = errors.New("conn lost")
	err := svc.Audit(1, model.AuditApproved, "")
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDh114Audit_ApproveDraftSyncsPublish(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Status: model.StatusDraft, AuditStatus: model.AuditPending}
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	assert.Equal(t, model.StatusPublished, fields["status"], "审核通过且为草稿应同步发布")
	_, hasPub := fields["published_at"]
	assert.True(t, hasPub, "应补齐发布时间")
}

func TestDh114Audit_ApproveAlreadyPublishedNoStatusChange(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	old := time.Now().Add(-time.Hour)
	repo.byID[1] = &model.Dh114{Status: model.StatusPublished, AuditStatus: model.AuditPending, PublishedAt: &old}
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	_, hasStatus := fields["status"]
	assert.False(t, hasStatus, "已发布不应再改状态")
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "已有发布时间不应重置")
}

func TestDh114Audit_ApprovePublishedWithoutPublishedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	// 已发布但无 published_at（异常数据），审核通过应补齐
	repo.byID[1] = &model.Dh114{Status: model.StatusPublished, AuditStatus: model.AuditPending, PublishedAt: nil}
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasPub := fields["published_at"]
	assert.True(t, hasPub, "已发布但缺 published_at 应补齐")
	_, hasStatus := fields["status"]
	assert.False(t, hasStatus, "非草稿不应改状态")
}

func TestDh114Audit_Reject(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Status: model.StatusPublished, AuditStatus: model.AuditPending}
	err := svc.Audit(1, model.AuditRejected, "资质不符")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditRejected, fields["audit_status"])
	assert.Equal(t, "资质不符", fields["audit_reason"])
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "拒绝不应设置发布时间")
}

func TestDh114Audit_UpdateFieldsError(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{AuditStatus: model.AuditPending}
	repo.updateFieldsErr = errors.New("update err")
	err := svc.Audit(1, model.AuditApproved, "")
	require.Error(t, err)
	assert.Equal(t, "update err", err.Error())
}

// ==================== BatchAudit ====================

func TestDh114BatchAudit_AllSuccess(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{AuditStatus: model.AuditPending}
	repo.byID[2] = &model.Dh114{AuditStatus: model.AuditPending}
	result, err := svc.BatchAudit(&dto.BatchAuditRequest{
		IDs:         []uint{1, 2},
		AuditStatus: model.AuditApproved,
		AuditReason: "通过",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Success)
	assert.Equal(t, 0, result.Failed)
	assert.Empty(t, result.FailedIDs)
}

func TestDh114BatchAudit_PartialFailure(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{AuditStatus: model.AuditPending}
	// id=2 不存在
	result, err := svc.BatchAudit(&dto.BatchAuditRequest{
		IDs:         []uint{1, 2},
		AuditStatus: model.AuditApproved,
	})
	require.NoError(t, err, "批量审核部分失败不应返回错误")
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 1, result.Success)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, []uint{2}, result.FailedIDs)
}

func TestDh114BatchAudit_EmptyIDs(t *testing.T) {
	svc, _, _, _, _ := newDh114ServiceWithMocks()
	result, err := svc.BatchAudit(&dto.BatchAuditRequest{IDs: []uint{}, AuditStatus: model.AuditApproved})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Equal(t, 0, result.Success)
	assert.Equal(t, 0, result.Failed)
}

// ==================== AdminUpdateStatus ====================

func TestAdminUpdateStatus_PublishSetsPublishedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Status: model.StatusOffline, PublishedAt: nil}
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	_, hasPub := fields["published_at"]
	assert.True(t, hasPub, "发布应补齐 published_at")
}

func TestAdminUpdateStatus_PublishNoResetIfAlreadySet(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	old := time.Now().Add(-time.Hour)
	repo.byID[1] = &model.Dh114{Status: model.StatusOffline, PublishedAt: &old}
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "已有 published_at 不应重置")
}

func TestAdminUpdateStatus_OfflineNoPublishedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{Status: model.StatusPublished}
	err := svc.AdminUpdateStatus(1, model.StatusOffline)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	_, hasPub := fields["published_at"]
	assert.False(t, hasPub, "下架不应设置 published_at")
}

func TestAdminUpdateStatus_UpdateFieldsError(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.byID[1] = &model.Dh114{}
	repo.updateFieldsErr = errors.New("update err")
	err := svc.AdminUpdateStatus(1, model.StatusOffline)
	require.Error(t, err)
	assert.Equal(t, "update err", err.Error())
}

// ==================== UpdatePromotion ====================

func TestUpdatePromotion_EmptyFieldsNoOp(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{})
	require.NoError(t, err)
	assert.Empty(t, repo.updateFieldsCalls, "无字段不应调用 UpdateFields")
}

func TestUpdatePromotion_FieldsBuilding(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	level := 8
	weight := 9.5
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{
		Featured:       boolPtr(true),
		Picked:         boolPtr(true),
		PromotionLevel: &level,
		TrafficWeight:  &weight,
	})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, true, fields["featured"])
	assert.Equal(t, true, fields["picked"])
	assert.Equal(t, 8, fields["promotion_level"])
	assert.Equal(t, 9.5, fields["traffic_weight"])
	_, hasVerified := fields["verified"]
	assert.False(t, hasVerified, "未传 verified 不应写入")
	_, hasVerifiedAt := fields["verified_at"]
	assert.False(t, hasVerifiedAt)
}

func TestUpdatePromotion_VerifiedTrueSetsVerifiedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{Verified: boolPtr(true)})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, true, fields["verified"])
	_, hasVerifiedAt := fields["verified_at"]
	assert.True(t, hasVerifiedAt, "认证为 true 应设置认证时间")
}

func TestUpdatePromotion_VerifiedFalseNoVerifiedAt(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{Verified: boolPtr(false)})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, false, fields["verified"])
	_, hasVerifiedAt := fields["verified_at"]
	assert.False(t, hasVerifiedAt, "取消认证不应设置认证时间")
}

func TestUpdatePromotion_UpdateFieldsError(t *testing.T) {
	svc, repo, _, _, _ := newDh114ServiceWithMocks()
	repo.updateFieldsErr = errors.New("update err")
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{Featured: boolPtr(true)})
	require.Error(t, err)
	assert.Equal(t, "update err", err.Error())
}

// ==================== NewDh114Service ====================

func TestNewDh114Service_NotNil(t *testing.T) {
	repo := newMockDh114Repo()
	svc := NewDh114Service(repo, &mockImageRepo{}, &mockVisitRepo{}, &mockFavoriteRepo{}, &mockPhoneCallRepo{})
	require.NotNil(t, svc)
}

func TestGenerateDh114No_Format(t *testing.T) {
	no := generateDh114No("DH114CALL")
	assert.Contains(t, no, "DH114CALL")
	assert.Len(t, no, len("DH114CALL")+14+6, "前缀14位时间戳6位随机数")
}
