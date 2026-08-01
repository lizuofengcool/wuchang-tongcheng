// Package service 同城车辆买卖主表业务逻辑层单元测试 - 车源包。
// 使用内存 mock 仓储覆盖：发布与默认值兜底（listing/source/car_type/mileage_unit/
// condition_level/use_type）、发布即审核通过、发布时间补齐、图片子表替换、
// 更新字段构建与发布时间补齐与状态机迁移、删除权限校验、详情浏览量自增与收藏状态、
// 列表/附近/搜索/我的发布/高级搜索（带经纬度半径走附近分支）的分页与错误传递、
// 收藏切换（创建/删除 + 计数增减）与未登录态、浏览记录设备/来源默认值兜底、
// 管理端审核状态联动（通过同步发布、拒绝同步下架）、管理端状态更新发布时间补齐、
// 真车认证时间戳构建、推广字段构建等核心逻辑。不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockCarRepo =====

type updateFieldsCall struct {
	id     uint
	fields map[string]interface{}
}

type mockCarRepo struct {
	byID   map[uint]*model.Car
	nextID uint

	images map[uint][]model.CarImage
	favs   map[uint]map[uint]bool // userID -> carID -> exists
	views  []model.CarView

	incrView      []uint
	incrFavCount  []uint
	decrFavCount  []uint
	incrContact   []uint
	incrShare     []uint
	incrTestDrive []uint

	createErr       error
	findErr         error
	updateFieldsErr error
	deleteErr       error
	replaceImgErr   error
	listErr         error
	nearbyErr       error
	searchErr       error
	byUserErr       error
	adminListErr    error
	favExistsErr    error
	createFavErr    error
	deleteFavErr    error
	listFavsErr     error
	createViewErr   error

	updateFieldsCalls  []updateFieldsCall
	deleteCalls        []uint
	replaceImagesCalls []struct {
		carID    uint
		regionID uint
		imgs     []model.CarImage
	}
	deleteImagesCalls []uint
	createFavCalls    []model.CarFavorite
	deleteFavCalls    []struct {
		userID uint
		carID  uint
	}

	listReturn      []model.Car
	listTotal       int64
	adminListReturn []model.Car
	adminListTotal  int64
	nearbyReturn    []model.Car
	nearbyTotal     int64
	searchReturn    []model.Car
	searchTotal     int64
	byUserReturn    []model.Car
	byUserTotal     int64
	listFavsReturn  []model.CarFavorite
	listFavsTotal   int64

	// 附近分支捕获到的半径，用于断言默认值兜底
	lastNearbyRadius float64
}

func newMockCarRepo() *mockCarRepo {
	return &mockCarRepo{
		byID:   make(map[uint]*model.Car),
		images: make(map[uint][]model.CarImage),
		favs:   make(map[uint]map[uint]bool),
		nextID: 1,
	}
}

// seed 预置一条车源并返回其副本指针。
func (m *mockCarRepo) seed(c *model.Car) *model.Car {
	if c.ID == 0 {
		c.ID = m.nextID
		m.nextID++
	}
	cp := *c
	m.byID[c.ID] = &cp
	return &cp
}

func (m *mockCarRepo) Create(c *model.Car) error {
	if m.createErr != nil {
		return m.createErr
	}
	if c.ID == 0 {
		c.ID = m.nextID
		m.nextID++
	}
	cp := *c
	m.byID[c.ID] = &cp
	return nil
}

func (m *mockCarRepo) FindByID(id uint) (*model.Car, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	c, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockCarRepo) Update(c *model.Car) error {
	cp := *c
	m.byID[c.ID] = &cp
	return nil
}

func (m *mockCarRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, updateFieldsCall{id: id, fields: fields})
	c, ok := m.byID[id]
	if !ok {
		return nil
	}
	if v, ok := fields["status"]; ok {
		if s, ok2 := v.(int); ok2 {
			c.Status = s
		}
	}
	if v, ok := fields["audit_status"]; ok {
		if s, ok2 := v.(int); ok2 {
			c.AuditStatus = s
		}
	}
	if v, ok := fields["audit_reason"]; ok {
		if s, ok2 := v.(string); ok2 {
			c.AuditReason = s
		}
	}
	if v, ok := fields["published_at"]; ok {
		if t, ok2 := v.(*time.Time); ok2 && t != nil {
			c.PublishedAt = t
		}
	}
	if v, ok := fields["title"]; ok {
		if s, ok2 := v.(string); ok2 {
			c.Title = s
		}
	}
	if v, ok := fields["price"]; ok {
		if f, ok2 := v.(float64); ok2 {
			c.Price = f
		}
	}
	if v, ok := fields["mileage"]; ok {
		if f, ok2 := v.(float64); ok2 {
			c.Mileage = f
		}
	}
	if v, ok := fields["featured"]; ok {
		if b, ok2 := v.(bool); ok2 {
			c.Featured = b
		}
	}
	if v, ok := fields["picked"]; ok {
		if b, ok2 := v.(bool); ok2 {
			c.Picked = b
		}
	}
	if v, ok := fields["verified"]; ok {
		if b, ok2 := v.(bool); ok2 {
			c.Verified = b
		}
	}
	if v, ok := fields["promotion_level"]; ok {
		if i, ok2 := v.(int); ok2 {
			c.PromotionLevel = i
		}
	}
	if v, ok := fields["traffic_weight"]; ok {
		if f, ok2 := v.(float64); ok2 {
			c.TrafficWeight = f
		}
	}
	if v, ok := fields["real_car_verified"]; ok {
		if b, ok2 := v.(bool); ok2 {
			c.RealCarVerified = b
		}
	}
	if v, ok := fields["real_car_verified_at"]; ok {
		if t, ok2 := v.(*time.Time); ok2 {
			c.RealCarVerifiedAt = t
		} else if t, ok2 := v.(time.Time); ok2 {
			cp := t
			c.RealCarVerifiedAt = &cp
		} else if v == nil {
			c.RealCarVerifiedAt = nil
		}
	}
	return nil
}

func (m *mockCarRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.byID, id)
	return nil
}

func (m *mockCarRepo) List(regionID uint, pagination *utils.Pagination, opts repository.CarListOptions) ([]model.Car, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockCarRepo) AdminList(pagination *utils.Pagination, opts repository.CarAdminListOptions) ([]model.Car, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockCarRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.CarListOptions) ([]model.Car, int64, error) {
	m.lastNearbyRadius = radiusKm
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	return m.nearbyReturn, m.nearbyTotal, nil
}

func (m *mockCarRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Car, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	return m.searchReturn, m.searchTotal, nil
}

func (m *mockCarRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Car, int64, error) {
	if m.byUserErr != nil {
		return nil, 0, m.byUserErr
	}
	return m.byUserReturn, m.byUserTotal, nil
}

func (m *mockCarRepo) IncrViewCount(id uint) error {
	m.incrView = append(m.incrView, id)
	if c, ok := m.byID[id]; ok {
		c.ViewCount++
	}
	return nil
}

func (m *mockCarRepo) IncrFavCount(id uint) error {
	m.incrFavCount = append(m.incrFavCount, id)
	if c, ok := m.byID[id]; ok {
		c.FavCount++
	}
	return nil
}

func (m *mockCarRepo) DecrFavCount(id uint) error {
	m.decrFavCount = append(m.decrFavCount, id)
	if c, ok := m.byID[id]; ok && c.FavCount > 0 {
		c.FavCount--
	}
	return nil
}

func (m *mockCarRepo) IncrContactCount(id uint) error {
	m.incrContact = append(m.incrContact, id)
	if c, ok := m.byID[id]; ok {
		c.ContactCount++
	}
	return nil
}

func (m *mockCarRepo) IncrShareCount(id uint) error {
	m.incrShare = append(m.incrShare, id)
	if c, ok := m.byID[id]; ok {
		c.ShareCount++
	}
	return nil
}

func (m *mockCarRepo) IncrTestDriveCount(id uint) error {
	m.incrTestDrive = append(m.incrTestDrive, id)
	if c, ok := m.byID[id]; ok {
		c.TestDriveCount++
	}
	return nil
}

func (m *mockCarRepo) ListImages(carID uint) ([]model.CarImage, error) {
	return m.images[carID], nil
}

func (m *mockCarRepo) ReplaceImages(carID uint, regionID uint, images []model.CarImage) error {
	if m.replaceImgErr != nil {
		return m.replaceImgErr
	}
	m.replaceImagesCalls = append(m.replaceImagesCalls, struct {
		carID    uint
		regionID uint
		imgs     []model.CarImage
	}{carID: carID, regionID: regionID, imgs: images})
	m.images[carID] = images
	return nil
}

func (m *mockCarRepo) DeleteImages(carID uint) error {
	m.deleteImagesCalls = append(m.deleteImagesCalls, carID)
	delete(m.images, carID)
	return nil
}

func (m *mockCarRepo) FavExists(userID, carID uint) (bool, error) {
	if m.favExistsErr != nil {
		return false, m.favExistsErr
	}
	if m.favs[userID] == nil {
		return false, nil
	}
	return m.favs[userID][carID], nil
}

func (m *mockCarRepo) CreateFav(fav *model.CarFavorite) error {
	if m.createFavErr != nil {
		return m.createFavErr
	}
	m.createFavCalls = append(m.createFavCalls, *fav)
	if m.favs[fav.UserID] == nil {
		m.favs[fav.UserID] = make(map[uint]bool)
	}
	m.favs[fav.UserID][fav.CarID] = true
	return nil
}

func (m *mockCarRepo) DeleteFav(userID, carID uint) error {
	if m.deleteFavErr != nil {
		return m.deleteFavErr
	}
	m.deleteFavCalls = append(m.deleteFavCalls, struct {
		userID uint
		carID  uint
	}{userID: userID, carID: carID})
	if m.favs[userID] != nil {
		delete(m.favs[userID], carID)
	}
	return nil
}

func (m *mockCarRepo) ListFavs(userID uint, page, pageSize int) ([]model.CarFavorite, int64, error) {
	if m.listFavsErr != nil {
		return nil, 0, m.listFavsErr
	}
	return m.listFavsReturn, m.listFavsTotal, nil
}

func (m *mockCarRepo) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if m.favs[userID] != nil {
			result[id] = m.favs[userID][id]
		}
	}
	return result, nil
}

func (m *mockCarRepo) CreateView(v *model.CarView) error {
	if m.createViewErr != nil {
		return m.createViewErr
	}
	m.views = append(m.views, *v)
	return nil
}

// ===== 测试装配 =====

func newSvc() (CarService, *mockCarRepo) {
	repo := newMockCarRepo()
	// carService 持有 imageRepo 字段但在主表 service 流程中未调用，故传 nil 即可。
	return NewCarService(repo, nil), repo
}

// mkCar 不再需要：seed 会自动分配 ID（从 1 起），测试通过 GetByID(1,...) 访问首条记录。

// strPtr 返回字符串指针。
func strPtr(s string) *string   { return &s }
func intPtr(i int) *int         { return &i }
func f64Ptr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool      { return &b }

// ===== 文本辅助 =====

func TestCarStatusText(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{model.StatusDraft, "草稿"},
		{model.StatusPublished, "已发布"},
		{model.StatusOffline, "已下架"},
		{model.StatusExpired, "已过期"},
		{model.StatusDeleted, "已删除"},
		{model.StatusSold, "已售出"},
		{99, ""},
		{0, "草稿"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, carStatusText(c.status))
	}
}

func TestCarAuditStatusText(t *testing.T) {
	assert.Equal(t, "待审", carAuditStatusText(model.AuditPending))
	assert.Equal(t, "通过", carAuditStatusText(model.AuditApproved))
	assert.Equal(t, "拒绝", carAuditStatusText(model.AuditRejected))
	assert.Equal(t, "", carAuditStatusText(99))
}

// ===== toCarInfo =====

func TestToCarInfo_BasicAndStatusText(t *testing.T) {
	c := &model.Car{
		Title:       "奥迪 A4L",
		Status:      model.StatusPublished,
		AuditStatus: model.AuditApproved,
		Price:       200000,
	}
	info := toCarInfo(c, nil)
	assert.Equal(t, c.Title, info.Title)
	assert.Equal(t, c.Price, info.Price)
	assert.Equal(t, "已发布", info.StatusText)
	assert.Equal(t, "通过", info.AuditStatusText)
	// 无图片时 Images 为 nil（不强制空切片）
	assert.Nil(t, info.Images)
}

func TestToCarInfo_JSONBPassthrough(t *testing.T) {
	feat, err := model.FromJSON(map[string]interface{}{"sunroof": true})
	require.NoError(t, err)
	tags, err := model.FromJSON([]string{"hot", "new"})
	require.NoError(t, err)
	c := &model.Car{
		Features: feat,
		Tags:     tags,
	}
	info := toCarInfo(c, nil)
	assert.NotNil(t, info.Features)
	assert.NotNil(t, info.Tags)
}

func TestToCarInfo_JSONBNilOmitted(t *testing.T) {
	c := &model.Car{}
	info := toCarInfo(c, nil)
	assert.Nil(t, info.Features)
	assert.Nil(t, info.Tags)
	assert.Nil(t, info.InspectionItems)
	assert.Nil(t, info.AccidentHistory)
}

func TestToCarInfo_Images(t *testing.T) {
	c := &model.Car{}
	imgs := []model.CarImage{
		{ImageType: "exterior", URL: "u1", IsCover: true, Sort: 0},
		{ImageType: "interior", URL: "u2", Sort: 1},
	}
	info := toCarInfo(c, imgs)
	require.Len(t, info.Images, 2)
	assert.Equal(t, "u1", info.Images[0].URL)
	assert.True(t, info.Images[0].IsCover)
	assert.Equal(t, "interior", info.Images[1].ImageType)
}

// ===== Create =====

func TestCreate_DefaultsAndAuditApproved(t *testing.T) {
	svc, repo := newSvc()
	req := &dto.CreateCarRequest{Title: "二手奥迪 A4L", Price: 200000}

	info, err := svc.Create(5, 100, "张三", "13800000000", "avatar.png", req)
	require.NoError(t, err)
	require.NotNil(t, info)

	// 默认值兜底
	assert.Equal(t, model.ListingTypeUsed, info.ListingType)
	assert.Equal(t, model.SourceTypePersonal, info.SourceType)
	assert.Equal(t, model.CarTypeSedan, info.CarType)
	assert.Equal(t, model.MileageUnitKM, info.MileageUnit)
	assert.Equal(t, model.ConditionLevelA, info.ConditionLevel)
	assert.Equal(t, model.UseTypeNonOperational, info.UseType)
	// MVP：发布即审核通过
	assert.Equal(t, model.AuditApproved, info.AuditStatus)
	// 地区与发布者
	assert.Equal(t, uint(5), info.RegionID)
	assert.Equal(t, uint(100), info.UserID)
	assert.Equal(t, "张三", info.UserName)
	assert.Equal(t, "13800000000", info.UserPhone)
	assert.Equal(t, "avatar.png", info.UserAvatar)
	// 默认 status=0（草稿），不设置发布时间
	assert.Equal(t, model.StatusDraft, info.Status)
	assert.Nil(t, info.PublishedAt)
	// 落库
	require.Len(t, repo.byID, 1)
	assert.Equal(t, model.AuditApproved, repo.byID[info.ID].AuditStatus)
}

func TestCreate_PublishedSetsPublishedAt(t *testing.T) {
	svc, repo := newSvc()
	req := &dto.CreateCarRequest{Title: "现车", Status: model.StatusPublished, Price: 100000}

	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPublished, info.Status)
	require.NotNil(t, info.PublishedAt)
	require.NotNil(t, repo.byID[info.ID].PublishedAt)
}

func TestCreate_DraftNoPublishedAt(t *testing.T) {
	svc, _ := newSvc()
	req := &dto.CreateCarRequest{Title: "草稿", Status: model.StatusDraft}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Nil(t, info.PublishedAt)
}

func TestCreate_DefaultsRespectedWhenProvided(t *testing.T) {
	svc, _ := newSvc()
	req := &dto.CreateCarRequest{
		Title:          "新能源 SUV",
		ListingType:    model.ListingTypeNew,
		SourceType:     model.SourceTypeDealer,
		CarType:        model.CarTypeSUV,
		MileageUnit:    model.MileageUnitMile,
		ConditionLevel: model.ConditionLevelB,
		UseType:        model.UseTypeOperational,
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, model.ListingTypeNew, info.ListingType)
	assert.Equal(t, model.SourceTypeDealer, info.SourceType)
	assert.Equal(t, model.CarTypeSUV, info.CarType)
	assert.Equal(t, model.MileageUnitMile, info.MileageUnit)
	assert.Equal(t, model.ConditionLevelB, info.ConditionLevel)
	assert.Equal(t, model.UseTypeOperational, info.UseType)
}

func TestCreate_JSONBFieldsParsed(t *testing.T) {
	svc, repo := newSvc()
	req := &dto.CreateCarRequest{
		Title:           "带配置",
		Features:        map[string]interface{}{"sunroof": true, "leather": true},
		Tags:            []string{"hot"},
		InspectionItems: []interface{}{"item1"},
		AccidentHistory: []interface{}{"acc1"},
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.NotNil(t, info.Features)
	assert.NotNil(t, info.Tags)
	assert.NotNil(t, info.InspectionItems)
	assert.NotNil(t, info.AccidentHistory)
	require.Contains(t, repo.byID, info.ID)
	assert.NotNil(t, repo.byID[info.ID].Features)
	assert.NotNil(t, repo.byID[info.ID].Tags)
}

func TestCreate_JSONBInvalidDoesNotPanic(t *testing.T) {
	svc, _ := newSvc()
	// channel 无法被 json.Marshal，FromJSON 会返回 error，service 静默忽略
	req := &dto.CreateCarRequest{
		Title:    "无效 JSONB",
		Features: make(chan int),
	}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Nil(t, info.Features)
}

func TestCreate_ImagesReplaced(t *testing.T) {
	svc, repo := newSvc()
	req := &dto.CreateCarRequest{
		Title: "带图",
		Images: []dto.CarImageInput{
			{ImageType: "exterior", URL: "u1", IsCover: true},
			{ImageType: "interior", URL: "u2"},
		},
	}
	info, err := svc.Create(7, 9, "u", "p", "a", req)
	require.NoError(t, err)
	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, info.ID, repo.replaceImagesCalls[0].carID)
	assert.Equal(t, uint(7), repo.replaceImagesCalls[0].regionID)
	require.Len(t, repo.replaceImagesCalls[0].imgs, 2)
	assert.Equal(t, "u1", repo.replaceImagesCalls[0].imgs[0].URL)
	// 返回 info 时未拼装图片（Create 不回查）
	assert.Nil(t, info.Images)
}

func TestCreate_CreateErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.createErr = errors.New("db down")
	info, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateCarRequest{Title: "x"})
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, "db down", err.Error())
}

// ===== Update =====

func TestUpdate_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Update(999, 1, &dto.UpdateCarRequest{Title: strPtr("new")})
	assert.ErrorIs(t, err, ErrCarNotFound)
}

func TestUpdate_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Update(1, 1, &dto.UpdateCarRequest{Title: strPtr("new")})
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrCarNotFound)
	assert.Equal(t, "conn lost", err.Error())
}

func TestUpdate_NoPermission(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.Update(1, 200, &dto.UpdateCarRequest{Title: strPtr("new")})
	assert.ErrorIs(t, err, ErrCarNoPermission)
}

func TestUpdate_FieldBuilding(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	price := 188000.0
	mileage := 50000.0
	err := svc.Update(1, 100, &dto.UpdateCarRequest{
		Title:        strPtr("新标题"),
		Price:        &price,
		Mileage:      &mileage,
		VIN:          strPtr("LSJW32"),
		LicensePlate: strPtr("鄂A12345"),
	})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	call := repo.updateFieldsCalls[0]
	assert.Equal(t, "新标题", call.fields["title"])
	assert.Equal(t, 188000.0, call.fields["price"])
	assert.Equal(t, 50000.0, call.fields["mileage"])
	assert.Equal(t, "LSJW32", call.fields["vin"])
	assert.Equal(t, "鄂A12345", call.fields["license_plate"])
}

func TestUpdate_StatusPublishedFromDraftSetsPublishedAtAndAudit(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusDraft})
	status := model.StatusPublished
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Status: &status})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	require.Contains(t, fields, "published_at")
	_, ok := fields["published_at"].(*time.Time)
	assert.True(t, ok)
}

func TestUpdate_StatusPublishedWhenAlreadyPublishedNoRebuild(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusPublished, AuditStatus: model.AuditApproved})
	status := model.StatusPublished
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Status: &status})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	// 已经是发布态，不再补 published_at / audit_status
	assert.NotContains(t, fields, "published_at")
	assert.NotContains(t, fields, "audit_status")
}

func TestUpdate_StatusOfflineOnlySetsStatus(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusPublished})
	status := model.StatusOffline
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Status: &status})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	assert.NotContains(t, fields, "published_at")
	assert.NotContains(t, fields, "audit_status")
}

func TestUpdate_NoOpWhenNoFieldsAndNoImages(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.Update(1, 100, &dto.UpdateCarRequest{})
	require.NoError(t, err)
	assert.Empty(t, repo.updateFieldsCalls)
	assert.Empty(t, repo.replaceImagesCalls)
}

func TestUpdate_ImagesReplaced(t *testing.T) {
	svc, repo := newSvc()
	c := &model.Car{UserID: 100}
	c.RegionID = 7 // RegionID 为内嵌 RegionBaseModel 的提升字段，需在字面量外赋值
	repo.seed(c)
	imgs := []dto.CarImageInput{{ImageType: "exterior", URL: "u1"}}
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Images: &imgs})
	require.NoError(t, err)
	require.Len(t, repo.replaceImagesCalls, 1)
	assert.Equal(t, uint(1), repo.replaceImagesCalls[0].carID)
	assert.Equal(t, uint(7), repo.replaceImagesCalls[0].regionID)
	require.Len(t, repo.replaceImagesCalls[0].imgs, 1)
	assert.Equal(t, "u1", repo.replaceImagesCalls[0].imgs[0].URL)
}

func TestUpdate_UpdateFieldsErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Title: strPtr("x")})
	assert.Equal(t, "update failed", err.Error())
}

func TestUpdate_ReplaceImagesErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.replaceImgErr = errors.New("replace failed")
	imgs := []dto.CarImageInput{{URL: "u1"}}
	err := svc.Update(1, 100, &dto.UpdateCarRequest{Images: &imgs})
	assert.Equal(t, "replace failed", err.Error())
}

func TestUpdate_JSONBFieldsBuilt(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.Update(1, 100, &dto.UpdateCarRequest{
		Features: map[string]interface{}{"sunroof": true},
		Tags:     []string{"hot"},
	})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	_, hasFeatures := fields["features"]
	assert.True(t, hasFeatures)
	_, hasTags := fields["tags"]
	assert.True(t, hasTags)
}

// ===== Delete =====

func TestDelete_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Delete(999, 1)
	assert.ErrorIs(t, err, ErrCarNotFound)
}

func TestDelete_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Delete(1, 1)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDelete_NoPermission(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.Delete(1, 200)
	assert.ErrorIs(t, err, ErrCarNoPermission)
}

func TestDelete_Success(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.Delete(1, 100)
	require.NoError(t, err)
	require.Len(t, repo.deleteCalls, 1)
	assert.Equal(t, uint(1), repo.deleteCalls[0])
}

func TestDelete_RepoDeleteErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.deleteErr = errors.New("delete failed")
	err := svc.Delete(1, 100)
	assert.Equal(t, "delete failed", err.Error())
}

// ===== GetByID =====

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newSvc()
	info, err := svc.GetByID(999, 1)
	assert.ErrorIs(t, err, ErrCarNotFound)
	assert.Nil(t, info)
}

func TestGetByID_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	info, err := svc.GetByID(1, 1)
	assert.Equal(t, "conn lost", err.Error())
	assert.Nil(t, info)
}

func TestGetByID_IncrViewAndHasFaved(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, ViewCount: 5, FavCount: 3})
	repo.favs[7] = map[uint]bool{1: true}
	info, err := svc.GetByID(1, 7)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Len(t, repo.incrView, 1)
	assert.Equal(t, uint(1), repo.incrView[0])
	assert.Equal(t, 6, info.ViewCount) // service 内 c.ViewCount++
	assert.True(t, info.HasFaved)
	assert.Equal(t, []dto.CarImageInfo{}, info.Images) // images nil → 空切片
}

func TestGetByID_NoFavCheckWhenUserIDZero(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	info, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	assert.False(t, info.HasFaved)
}

func TestGetByID_ImagesAttached(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.images[1] = []model.CarImage{{URL: "u1"}, {URL: "u2"}}
	info, err := svc.GetByID(1, 0)
	require.NoError(t, err)
	require.Len(t, info.Images, 2)
}

// ===== List / ListNearby / Search / ListMine / AdvancedSearch =====

func TestList_PaginationAndResult(t *testing.T) {
	svc, repo := newSvc()
	repo.listReturn = []model.Car{{Title: "a"}, {Title: "b"}}
	repo.listTotal = 2
	p, list, err := svc.List(5, &dto.CarListRequest{
		CarType: "suv",
		Sort:    "price_asc",
	})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "a", list[0].Title)
	assert.Equal(t, int64(2), p.Total)
}

func TestList_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listErr = errors.New("list failed")
	p, list, err := svc.List(1, &dto.CarListRequest{})
	require.Error(t, err)
	assert.Nil(t, p)
	assert.Nil(t, list)
	assert.Equal(t, "list failed", err.Error())
}

func TestListNearby_DefaultRadiusWhenZero(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyReturn = []model.Car{{}}
	repo.nearbyTotal = 1
	p, list, err := svc.ListNearby(5, &dto.CarNearbyRequest{Latitude: 30.0, Longitude: 114.0, RadiusKm: 0})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
	// service 兜底 radiusKm<=0 → 5
	assert.Equal(t, 5.0, repo.lastNearbyRadius)
}

func TestListNearby_KeepsProvidedRadius(t *testing.T) {
	svc, repo := newSvc()
	_, _, err := svc.ListNearby(5, &dto.CarNearbyRequest{Latitude: 30.0, Longitude: 114.0, RadiusKm: 12.5})
	require.NoError(t, err)
	assert.Equal(t, 12.5, repo.lastNearbyRadius)
}

func TestListNearby_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyErr = errors.New("nearby failed")
	p, list, err := svc.ListNearby(5, &dto.CarNearbyRequest{Latitude: 30.0, Longitude: 114.0})
	require.Error(t, err)
	assert.Nil(t, p)
	assert.Nil(t, list)
	assert.Equal(t, "nearby failed", err.Error())
}

func TestSearch_Result(t *testing.T) {
	svc, repo := newSvc()
	repo.searchReturn = []model.Car{{Title: "奥迪"}}
	repo.searchTotal = 1
	p, list, err := svc.Search(5, &dto.CarSearchRequest{Keyword: "奥迪"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
}

func TestSearch_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.searchErr = errors.New("search failed")
	_, _, err := svc.Search(1, &dto.CarSearchRequest{Keyword: "x"})
	assert.Equal(t, "search failed", err.Error())
}

func TestListMine_Result(t *testing.T) {
	svc, repo := newSvc()
	repo.byUserReturn = []model.Car{{Title: "我的车"}}
	repo.byUserTotal = 1
	p, list, err := svc.ListMine(100, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
}

func TestListMine_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.byUserErr = errors.New("by user failed")
	_, _, err := svc.ListMine(100, 1, 10)
	assert.Equal(t, "by user failed", err.Error())
}

func TestAdvancedSearch_WithLocationUsesNearby(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyReturn = []model.Car{{Title: "附近车"}}
	repo.nearbyTotal = 1
	p, list, err := svc.AdvancedSearch(5, &dto.AdvancedSearchRequest{
		Latitude:  30.0,
		Longitude: 114.0,
		RadiusKm:  8,
		Keyword:   "奥迪",
	})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
	assert.Equal(t, 8.0, repo.lastNearbyRadius)
}

func TestAdvancedSearch_WithoutLocationUsesList(t *testing.T) {
	svc, repo := newSvc()
	repo.listReturn = []model.Car{{Title: "列表车"}}
	repo.listTotal = 1
	p, list, err := svc.AdvancedSearch(5, &dto.AdvancedSearchRequest{Keyword: "奥迪"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
}

func TestAdvancedSearch_NearbyErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyErr = errors.New("nearby failed")
	_, _, err := svc.AdvancedSearch(5, &dto.AdvancedSearchRequest{Latitude: 30.0, Longitude: 114.0, RadiusKm: 8})
	assert.Equal(t, "nearby failed", err.Error())
}

func TestAdvancedSearch_ListErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listErr = errors.New("list failed")
	_, _, err := svc.AdvancedSearch(5, &dto.AdvancedSearchRequest{Keyword: "x"})
	assert.Equal(t, "list failed", err.Error())
}

// ===== Fav =====

func TestFav_UserIDZero(t *testing.T) {
	svc, _ := newSvc()
	resp, err := svc.Fav(0, 1)
	assert.ErrorIs(t, err, ErrCarNoPermission)
	assert.Nil(t, resp)
}

func TestFav_NotFound(t *testing.T) {
	svc, _ := newSvc()
	resp, err := svc.Fav(1, 999)
	assert.ErrorIs(t, err, ErrCarNotFound)
	assert.Nil(t, resp)
}

func TestFav_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	_, err := svc.Fav(1, 1)
	assert.Equal(t, "conn lost", err.Error())
}

func TestFav_AddWhenNotFaved(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, FavCount: 2})
	resp, err := svc.Fav(7, 1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 3, resp.FavCount) // c.FavCount + 1
	require.Len(t, repo.createFavCalls, 1)
	assert.Equal(t, uint(7), repo.createFavCalls[0].UserID)
	assert.Equal(t, uint(1), repo.createFavCalls[0].CarID)
	require.Len(t, repo.incrFavCount, 1)
	assert.Equal(t, uint(1), repo.incrFavCount[0])
}

func TestFav_RemoveWhenAlreadyFaved(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, FavCount: 3})
	repo.favs[7] = map[uint]bool{1: true}
	resp, err := svc.Fav(7, 1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 2, resp.FavCount) // c.FavCount - 1
	require.Len(t, repo.deleteFavCalls, 1)
	assert.Equal(t, uint(7), repo.deleteFavCalls[0].userID)
	assert.Equal(t, uint(1), repo.deleteFavCalls[0].carID)
	require.Len(t, repo.decrFavCount, 1)
	assert.Equal(t, uint(1), repo.decrFavCount[0])
	assert.Empty(t, repo.createFavCalls)
}

func TestFav_FavExistsErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.favExistsErr = errors.New("fav exists failed")
	_, err := svc.Fav(7, 1)
	assert.Equal(t, "fav exists failed", err.Error())
}

func TestFav_CreateFavErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.createFavErr = errors.New("create fav failed")
	_, err := svc.Fav(7, 1)
	assert.Equal(t, "create fav failed", err.Error())
}

func TestFav_DeleteFavErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, FavCount: 2})
	repo.favs[7] = map[uint]bool{1: true}
	repo.deleteFavErr = errors.New("delete fav failed")
	_, err := svc.Fav(7, 1)
	assert.Equal(t, "delete fav failed", err.Error())
}

// ===== FavStatus =====

func TestFavStatus_NotFound(t *testing.T) {
	svc, _ := newSvc()
	resp, err := svc.FavStatus(1, 999)
	assert.ErrorIs(t, err, ErrCarNotFound)
	assert.Nil(t, resp)
}

func TestFavStatus_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	_, err := svc.FavStatus(1, 1)
	assert.Equal(t, "conn lost", err.Error())
}

func TestFavStatus_UserIDZero(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, FavCount: 5})
	resp, err := svc.FavStatus(0, 1)
	require.NoError(t, err)
	assert.False(t, resp.HasFaved)
	assert.Equal(t, 5, resp.FavCount)
}

func TestFavStatus_Normal(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, FavCount: 5})
	repo.favs[7] = map[uint]bool{1: true}
	resp, err := svc.FavStatus(7, 1)
	require.NoError(t, err)
	assert.True(t, resp.HasFaved)
	assert.Equal(t, 5, resp.FavCount)
}

func TestFavStatus_FavExistsErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.favExistsErr = errors.New("fav exists failed")
	_, err := svc.FavStatus(7, 1)
	assert.Equal(t, "fav exists failed", err.Error())
}

// ===== ListFavs =====

func TestListFavs_Empty(t *testing.T) {
	svc, repo := newSvc()
	repo.listFavsReturn = nil
	repo.listFavsTotal = 0
	p, list, err := svc.ListFavs(7, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), p.Total)
	assert.Equal(t, []dto.CarInfo{}, list)
}

func TestListFavs_WithResults(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Title: "车1"})
	repo.seed(&model.Car{UserID: 100, Title: "车2"})
	repo.listFavsReturn = []model.CarFavorite{{UserID: 7, CarID: 1}, {UserID: 7, CarID: 2}, {UserID: 7, CarID: 999}}
	repo.listFavsTotal = 3
	p, list, err := svc.ListFavs(7, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), p.Total)
	require.Len(t, list, 2) // CarID=999 不存在，被跳过
}

func TestListFavs_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listFavsErr = errors.New("list favs failed")
	_, _, err := svc.ListFavs(7, 1, 10)
	assert.Equal(t, "list favs failed", err.Error())
}

// ===== IncrContact / IncrShare =====

func TestIncrContact_Delegates(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, ContactCount: 0})
	err := svc.IncrContact(1)
	require.NoError(t, err)
	require.Len(t, repo.incrContact, 1)
	assert.Equal(t, uint(1), repo.incrContact[0])
}

func TestIncrShare_Delegates(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, ShareCount: 0})
	err := svc.IncrShare(1)
	require.NoError(t, err)
	require.Len(t, repo.incrShare, 1)
	assert.Equal(t, uint(1), repo.incrShare[0])
}

// ===== RecordView =====

func TestRecordView_DefaultsDeviceAndSource(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.RecordView(7, "127.0.0.1", &dto.CarViewRequest{CarID: 1})
	require.NoError(t, err)
	require.Len(t, repo.views, 1)
	assert.Equal(t, "pc", repo.views[0].Device)
	assert.Equal(t, "direct", repo.views[0].Source)
	assert.Equal(t, "127.0.0.1", repo.views[0].IP)
	assert.Equal(t, uint(7), repo.views[0].UserID)
	require.Len(t, repo.incrView, 1)
	assert.Equal(t, uint(1), repo.incrView[0])
}

func TestRecordView_RespectsProvidedDeviceAndSource(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	err := svc.RecordView(7, "1.1.1.1", &dto.CarViewRequest{
		CarID:    1,
		Device:   "app",
		Source:   "search",
		Duration: 30,
	})
	require.NoError(t, err)
	require.Len(t, repo.views, 1)
	assert.Equal(t, "app", repo.views[0].Device)
	assert.Equal(t, "search", repo.views[0].Source)
	assert.Equal(t, 30, repo.views[0].Duration)
}

func TestRecordView_CreateViewErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.createViewErr = errors.New("create view failed")
	err := svc.RecordView(7, "1.1.1.1", &dto.CarViewRequest{CarID: 1})
	assert.Equal(t, "create view failed", err.Error())
}

// ===== AdminList / AdminGetByID =====

func TestAdminList_Result(t *testing.T) {
	svc, repo := newSvc()
	repo.adminListReturn = []model.Car{{Title: "车1"}}
	repo.adminListTotal = 1
	status := model.StatusPublished
	p, list, err := svc.AdminList(&dto.CarAdminListRequest{Status: &status, Keyword: "奥迪"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), p.Total)
}

func TestAdminList_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.adminListErr = errors.New("admin list failed")
	_, _, err := svc.AdminList(&dto.CarAdminListRequest{})
	assert.Equal(t, "admin list failed", err.Error())
}

func TestAdminGetByID_NotFound(t *testing.T) {
	svc, _ := newSvc()
	info, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrCarNotFound)
	assert.Nil(t, info)
}

func TestAdminGetByID_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	_, err := svc.AdminGetByID(1)
	assert.Equal(t, "conn lost", err.Error())
}

func TestAdminGetByID_WithImages(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	repo.images[1] = []model.CarImage{{URL: "u1"}}
	info, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	require.Len(t, info.Images, 1)
}

func TestAdminGetByID_NoImagesReturnsEmptySlice(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100})
	info, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	assert.Equal(t, []dto.CarImageInfo{}, info.Images)
}

// ===== Audit =====

func TestAudit_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Audit(999, model.AuditApproved, "")
	assert.ErrorIs(t, err, ErrCarNotFound)
}

func TestAudit_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Audit(1, model.AuditApproved, "")
	assert.Equal(t, "conn lost", err.Error())
}

func TestAudit_ApprovedFromDraftSetsStatusPublished(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusDraft})
	err := svc.Audit(1, model.AuditApproved, "")
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	assert.Equal(t, "", fields["audit_reason"])
	assert.Equal(t, model.StatusPublished, fields["status"])
	require.Contains(t, fields, "published_at")
}

func TestAudit_ApprovedFromPublishedDoesNotChangeStatus(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusPublished})
	err := svc.Audit(1, model.AuditApproved, "ok")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	assert.Equal(t, "ok", fields["audit_reason"])
	assert.NotContains(t, fields, "status")
	assert.NotContains(t, fields, "published_at")
}

func TestAudit_RejectedFromPublishedSetsOffline(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusPublished})
	err := svc.Audit(1, model.AuditRejected, "信息不实")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.AuditRejected, fields["audit_status"])
	assert.Equal(t, "信息不实", fields["audit_reason"])
	assert.Equal(t, model.StatusOffline, fields["status"])
}

func TestAudit_RejectedFromDraftDoesNotChangeStatus(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusDraft})
	err := svc.Audit(1, model.AuditRejected, "reason")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.NotContains(t, fields, "status")
}

func TestAudit_UpdateFieldsErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Car{UserID: 100, Status: model.StatusDraft})
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.Audit(1, model.AuditApproved, "")
	assert.Equal(t, "update failed", err.Error())
}

// ===== AdminUpdateStatus =====

func TestAdminUpdateStatus_PublishedSetsPublishedAtAndAudit(t *testing.T) {
	svc, repo := newSvc()
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusPublished, fields["status"])
	assert.Equal(t, model.AuditApproved, fields["audit_status"])
	require.Contains(t, fields, "published_at")
}

func TestAdminUpdateStatus_OfflineOnlySetsStatus(t *testing.T) {
	svc, repo := newSvc()
	err := svc.AdminUpdateStatus(1, model.StatusOffline)
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, model.StatusOffline, fields["status"])
	assert.NotContains(t, fields, "published_at")
	assert.NotContains(t, fields, "audit_status")
}

func TestAdminUpdateStatus_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.AdminUpdateStatus(1, model.StatusPublished)
	assert.Equal(t, "update failed", err.Error())
}

// ===== RealCarVerify =====

func TestRealCarVerify_VerifiedSetsTimestamp(t *testing.T) {
	svc, repo := newSvc()
	err := svc.RealCarVerify(1, true, "认证通过")
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, true, fields["real_car_verified"])
	ts, ok := fields["real_car_verified_at"].(*time.Time)
	require.True(t, ok, "real_car_verified_at 应为 *time.Time")
	require.NotNil(t, ts)
}

func TestRealCarVerify_UnverifiedClearsTimestamp(t *testing.T) {
	svc, repo := newSvc()
	err := svc.RealCarVerify(1, false, "认证失败")
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, false, fields["real_car_verified"])
	assert.Nil(t, fields["real_car_verified_at"])
}

func TestRealCarVerify_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.updateFieldsErr = errors.New("update failed")
	err := svc.RealCarVerify(1, true, "")
	assert.Equal(t, "update failed", err.Error())
}

// ===== UpdatePromotion =====

func TestUpdatePromotion_NoOpWhenEmpty(t *testing.T) {
	svc, repo := newSvc()
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{})
	require.NoError(t, err)
	assert.Empty(t, repo.updateFieldsCalls)
}

func TestUpdatePromotion_AllFields(t *testing.T) {
	svc, repo := newSvc()
	featured := true
	picked := true
	verified := true
	level := 5
	weight := 2.5
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{
		Featured:       &featured,
		Picked:         &picked,
		Verified:       &verified,
		PromotionLevel: &level,
		TrafficWeight:  &weight,
	})
	require.NoError(t, err)
	require.Len(t, repo.updateFieldsCalls, 1)
	fields := repo.updateFieldsCalls[0].fields
	assert.Equal(t, true, fields["featured"])
	assert.Equal(t, true, fields["picked"])
	assert.Equal(t, true, fields["verified"])
	assert.Equal(t, 5, fields["promotion_level"])
	assert.Equal(t, 2.5, fields["traffic_weight"])
}

func TestUpdatePromotion_PartialFields(t *testing.T) {
	svc, repo := newSvc()
	featured := false
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{Featured: &featured})
	require.NoError(t, err)
	fields := repo.updateFieldsCalls[0].fields
	assert.Len(t, fields, 1)
	assert.Equal(t, false, fields["featured"])
}

func TestUpdatePromotion_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.updateFieldsErr = errors.New("update failed")
	featured := true
	err := svc.UpdatePromotion(1, &dto.PromotionRequest{Featured: &featured})
	assert.Equal(t, "update failed", err.Error())
}
