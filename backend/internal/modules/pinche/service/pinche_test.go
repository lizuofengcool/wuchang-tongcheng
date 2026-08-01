// Package service 同城拼车出行业务逻辑层单元测试 - 拼车主表。
// 使用内存 mock 仓储覆盖：发布与默认值兜底（trip_type/role/total_seats/payment_method）、
// 发布即审核通过、发布时间补齐、AvailableSeats=TotalSeats、JSONB 字段解析与非法值兜底、
// 更新字段构建与无操作短路、更新权限校验、删除权限校验、详情浏览量自增、
// 列表/附近（默认半径兜底）/搜索（出发日期区间构建）/我的发布的分页与错误传递、
// 智能匹配（默认半径兜底、匹配度排序、零分过滤、ListMatch 错误传递）、calcMatchScore 各分支
// （起点过远/终点过远/无出发时间加分/低于阈值过滤）、haversineDistance 已知值、
// 互动计数委托、浏览记录忽略错误、管理端列表/详情、审核状态校验（已审核拒绝重复）、
// 管理端状态变更委托、批量审核/状态变更/删除的成功与部分失败统计。不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockPincheRepo =====

type updateCall struct {
	id     uint
	fields map[string]interface{}
}

type auditCall struct {
	id          uint
	auditStatus int
	reason      string
}

type statusCall struct {
	id     uint
	status int
}

type mockPincheRepo struct {
	byID   map[uint]*model.Pinche
	nextID uint

	// 列表返回值
	listReturn        []model.Pinche
	listTotal         int64
	adminListReturn   []model.Pinche
	adminListTotal    int64
	listMineReturn    []model.Pinche
	listMineTotal     int64
	nearbyReturn      []model.Pinche
	nearbyTotal       int64
	listMatchReturn   []model.Pinche
	listMatchTotal    int64
	countByStatusRet  int64

	// 附近半径捕获（断言默认值兜底）
	lastNearbyRadius float64
	// List 选项捕获（断言 Search 出发日期区间构建）
	lastListOpts repository.PincheListOptions

	// 调用记录
	updateCalls   []updateCall
	deleteCalls   []uint
	auditCalls    []auditCall
	statusCalls   []statusCall
	incrView      []uint
	incrContact   []uint
	incrShare     []uint
	incrFavCalls  []struct {
		id    uint
		delta int
	}

	// 错误注入
	createErr     error
	findErr       error
	updateErr     error
	deleteErr     error
	listErr       error
	adminListErr  error
	listMineErr   error
	nearbyErr     error
	listMatchErr  error
	updateAuditErr error
	updateStatusErr error
	incrViewErr   error
	incrContactErr error
	incrShareErr  error
	incrFavErr    error
	countByStatusErr error

	// 按 ID 注入失败（用于批量操作的部分失败场景）
	failIDs map[uint]bool
}

func newMockPincheRepo() *mockPincheRepo {
	return &mockPincheRepo{
		byID:    make(map[uint]*model.Pinche),
		nextID:  1,
		failIDs: make(map[uint]bool),
	}
}

// seed 预置一条拼车行程并返回其副本指针（自动分配 ID）。
func (m *mockPincheRepo) seed(p *model.Pinche) *model.Pinche {
	if p.ID == 0 {
		p.ID = m.nextID
		m.nextID++
	}
	cp := *p
	m.byID[p.ID] = &cp
	return &cp
}

func (m *mockPincheRepo) Create(p *model.Pinche) error {
	if m.createErr != nil {
		return m.createErr
	}
	if p.ID == 0 {
		p.ID = m.nextID
		m.nextID++
	}
	cp := *p
	m.byID[p.ID] = &cp
	return nil
}

func (m *mockPincheRepo) FindByID(id uint) (*model.Pinche, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	p, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockPincheRepo) Update(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updateCalls = append(m.updateCalls, updateCall{id: id, fields: fields})
	return nil
}

func (m *mockPincheRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if m.failIDs[id] {
		return errors.New("conflict")
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.byID, id)
	return nil
}

func (m *mockPincheRepo) List(regionID uint, pagination *utils.Pagination, opts repository.PincheListOptions) ([]model.Pinche, int64, error) {
	m.lastListOpts = opts
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listReturn, m.listTotal, nil
}

func (m *mockPincheRepo) AdminList(pagination *utils.Pagination, opts repository.PincheAdminListOptions) ([]model.Pinche, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	return m.adminListReturn, m.adminListTotal, nil
}

func (m *mockPincheRepo) ListMine(userID uint, pagination *utils.Pagination) ([]model.Pinche, int64, error) {
	if m.listMineErr != nil {
		return nil, 0, m.listMineErr
	}
	return m.listMineReturn, m.listMineTotal, nil
}

func (m *mockPincheRepo) ListNearby(regionID uint, pagination *utils.Pagination, opts repository.PincheNearbyOptions) ([]model.Pinche, int64, error) {
	m.lastNearbyRadius = opts.RadiusKm
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	return m.nearbyReturn, m.nearbyTotal, nil
}

func (m *mockPincheRepo) ListMatch(regionID uint, pagination *utils.Pagination, opts repository.PincheListOptions) ([]model.Pinche, int64, error) {
	if m.listMatchErr != nil {
		return nil, 0, m.listMatchErr
	}
	return m.listMatchReturn, m.listMatchTotal, nil
}

func (m *mockPincheRepo) UpdateStatus(id uint, status int) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if m.failIDs[id] {
		return errors.New("conflict")
	}
	m.statusCalls = append(m.statusCalls, statusCall{id: id, status: status})
	return nil
}

func (m *mockPincheRepo) UpdateAudit(id uint, auditStatus int, reason string) error {
	if m.updateAuditErr != nil {
		return m.updateAuditErr
	}
	if m.failIDs[id] {
		return errors.New("conflict")
	}
	m.auditCalls = append(m.auditCalls, auditCall{id: id, auditStatus: auditStatus, reason: reason})
	return nil
}

func (m *mockPincheRepo) IncrViewCount(id uint) error {
	if m.incrViewErr != nil {
		return m.incrViewErr
	}
	m.incrView = append(m.incrView, id)
	if p, ok := m.byID[id]; ok {
		p.ViewCount++
	}
	return nil
}

func (m *mockPincheRepo) IncrContactCount(id uint) error {
	if m.incrContactErr != nil {
		return m.incrContactErr
	}
	m.incrContact = append(m.incrContact, id)
	return nil
}

func (m *mockPincheRepo) IncrShareCount(id uint) error {
	if m.incrShareErr != nil {
		return m.incrShareErr
	}
	m.incrShare = append(m.incrShare, id)
	return nil
}

func (m *mockPincheRepo) IncrFavCount(id uint, delta int) error {
	if m.incrFavErr != nil {
		return m.incrFavErr
	}
	m.incrFavCalls = append(m.incrFavCalls, struct {
		id    uint
		delta int
	}{id: id, delta: delta})
	return nil
}

func (m *mockPincheRepo) CountByStatus(regionID uint, status int) (int64, error) {
	if m.countByStatusErr != nil {
		return 0, m.countByStatusErr
	}
	return m.countByStatusRet, nil
}

// ===== 测试装配 =====

func newSvc() (PincheService, *mockPincheRepo) {
	repo := newMockPincheRepo()
	return NewPincheService(repo), repo
}

// 指针辅助（intPtr 已在 pinche.go 中定义，复用之）。
func strPtr(s string) *string   { return &s }
func f64Ptr(f float64) *float64 { return &f }
func uintPtr(u uint) *uint      { return &u }

// validCreateReq 构造一份通过默认 binding 的最小创建请求。
func validCreateReq() *dto.CreatePincheRequest {
	dep := time.Now().Add(2 * time.Hour)
	return &dto.CreatePincheRequest{
		Title:           "武昌→光谷 顺风车",
		DepartureTime:   &dep,
		PickupLocation:  "武昌火车站",
		DropoffLocation: "光谷广场",
		PricePerSeat:    30,
	}
}

// ===== 文本辅助 =====

func TestPincheStatusText(t *testing.T) {
	assert.Equal(t, "草稿", pincheStatusText(model.PincheStatusDraft))
	assert.Equal(t, "已发布", pincheStatusText(model.PincheStatusPublished))
	assert.Equal(t, "已结束", pincheStatusText(model.PincheStatusFinished))
	assert.Equal(t, "已取消", pincheStatusText(model.PincheStatusCancelled))
	assert.Equal(t, "进行中", pincheStatusText(model.PincheStatusOngoing))
	assert.Equal(t, "", pincheStatusText(999))
}

func TestPincheAuditStatusText(t *testing.T) {
	assert.Equal(t, "待审", pincheAuditStatusText(model.PincheAuditPending))
	assert.Equal(t, "通过", pincheAuditStatusText(model.PincheAuditApproved))
	assert.Equal(t, "拒绝", pincheAuditStatusText(model.PincheAuditRejected))
	assert.Equal(t, "", pincheAuditStatusText(999))
}

func TestToPincheInfo_BasicAndStatusText(t *testing.T) {
	p := &model.Pinche{
		Status:      model.PincheStatusPublished,
		AuditStatus: model.PincheAuditApproved,
		Title:       "测试行程",
		TotalSeats:  4,
	}
	info := toPincheInfo(p)
	assert.Equal(t, "已发布", info.StatusText)
	assert.Equal(t, "通过", info.AuditStatusText)
	assert.Equal(t, "测试行程", info.Title)
	assert.Equal(t, 4, info.TotalSeats)
}

func TestToPincheInfo_JSONBPassthrough(t *testing.T) {
	jb, err := model.FromJSON(map[string]interface{}{"pet": true})
	require.NoError(t, err)
	p := &model.Pinche{Features: jb, Tags: jb}
	info := toPincheInfo(p)
	assert.NotNil(t, info.Features)
	assert.NotNil(t, info.Tags)
}

func TestToPincheInfo_JSONBNilOmitted(t *testing.T) {
	p := &model.Pinche{}
	info := toPincheInfo(p)
	// Features/Tags 为 nil 时 info 字段保持零值（interface{} nil）
	assert.Nil(t, info.Features)
	assert.Nil(t, info.Tags)
}

// ===== Create =====

func TestCreate_DefaultsAndAuditApproved(t *testing.T) {
	svc, repo := newSvc()
	info, err := svc.Create(5, 100, "张三", "13800000000", "avatar.png", validCreateReq())
	require.NoError(t, err)
	require.NotNil(t, info)

	// 默认值兜底
	assert.Equal(t, model.TripTypeShunfeng, info.TripType)
	assert.Equal(t, model.RoleDriver, info.Role)
	assert.Equal(t, 4, info.TotalSeats)
	assert.Equal(t, model.PaymentMethodCash, info.PaymentMethod)
	// MVP：发布即审核通过 + 已发布
	assert.Equal(t, model.PincheAuditApproved, info.AuditStatus)
	assert.Equal(t, model.PincheStatusPublished, info.Status)
	require.NotNil(t, info.PublishedAt)
	// 地区与发布者
	assert.Equal(t, uint(5), info.RegionID)
	assert.Equal(t, uint(100), info.UserID)
	assert.Equal(t, "张三", info.UserName)
	assert.Equal(t, "13800000000", info.UserPhone)
	assert.Equal(t, "avatar.png", info.UserAvatar)
	// 落库
	require.Len(t, repo.byID, 1)
	assert.Equal(t, model.PincheAuditApproved, repo.byID[info.ID].AuditStatus)
	assert.Equal(t, model.PincheStatusPublished, repo.byID[info.ID].Status)
}

func TestCreate_AvailableSeatsEqualsTotal(t *testing.T) {
	svc, repo := newSvc()
	req := validCreateReq()
	req.TotalSeats = 6
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, 6, info.TotalSeats)
	assert.Equal(t, 6, info.AvailableSeats)
	assert.Equal(t, 0, info.BookedSeats)
	require.Equal(t, 6, repo.byID[info.ID].AvailableSeats)
}

func TestCreate_DefaultsRespectedWhenProvided(t *testing.T) {
	svc, _ := newSvc()
	req := validCreateReq()
	req.TripType = model.TripTypeBaoche
	req.Role = model.RolePassenger
	req.TotalSeats = 2
	req.PaymentMethod = model.PaymentMethodWechat
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Equal(t, model.TripTypeBaoche, info.TripType)
	assert.Equal(t, model.RolePassenger, info.Role)
	assert.Equal(t, 2, info.TotalSeats)
	assert.Equal(t, model.PaymentMethodWechat, info.PaymentMethod)
}

func TestCreate_JSONBFieldsParsed(t *testing.T) {
	svc, repo := newSvc()
	req := validCreateReq()
	req.Features = map[string]interface{}{"ac": true}
	req.Tags = []string{"hot", "weekend"}
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.NotNil(t, info.Features)
	assert.NotNil(t, info.Tags)
	require.Contains(t, repo.byID, info.ID)
	assert.NotNil(t, repo.byID[info.ID].Features)
	assert.NotNil(t, repo.byID[info.ID].Tags)
}

func TestCreate_JSONBInvalidDoesNotPanic(t *testing.T) {
	svc, _ := newSvc()
	// channel 无法被 json.Marshal，FromJSON 返回 error，service 静默忽略
	req := validCreateReq()
	req.Features = make(chan int)
	info, err := svc.Create(1, 1, "u", "p", "a", req)
	require.NoError(t, err)
	assert.Nil(t, info.Features)
}

func TestCreate_CreateErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.createErr = errors.New("db down")
	info, err := svc.Create(1, 1, "u", "p", "a", validCreateReq())
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, "db down", err.Error())
}

// ===== Update =====

func TestUpdate_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Update(999, 1, &dto.UpdatePincheRequest{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrPincheNotFound)
}

func TestUpdate_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Update(1, 1, &dto.UpdatePincheRequest{Title: strPtr("x")})
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestUpdate_NoPermission(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	err := svc.Update(1, 200, &dto.UpdatePincheRequest{Title: strPtr("x")})
	assert.ErrorIs(t, err, ErrPincheNoPermission)
}

func TestUpdate_FieldBuilding(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})

	dep := time.Now().Add(3 * time.Hour)
	req := &dto.UpdatePincheRequest{
		Title:           strPtr("新标题"),
		Content:         strPtr("新内容"),
		CoverImage:      strPtr("cover.png"),
		DepartureTime:   &dep,
		PickupLocation:  strPtr("新起点"),
		PickupLat:       f64Ptr(30.1),
		PickupLng:       f64Ptr(114.1),
		DropoffLocation: strPtr("新终点"),
		DropoffLat:      f64Ptr(30.2),
		DropoffLng:      f64Ptr(114.2),
		DistanceKm:      f64Ptr(15.5),
		DurationMin:     intPtr(45),
		TotalSeats:      intPtr(3),
		PricePerSeat:    f64Ptr(40),
		TollFee:         f64Ptr(10),
		VehicleID:       uintPtr(7),
		RouteID:         uintPtr(8),
		PaymentMethod:   strPtr(model.PaymentMethodAlipay),
		Features:        map[string]interface{}{"ac": true},
		Tags:            []string{"vip"},
	}

	err := svc.Update(1, 100, req)
	require.NoError(t, err)
	require.Len(t, repo.updateCalls, 1)
	fields := repo.updateCalls[0].fields
	assert.Equal(t, uint(1), repo.updateCalls[0].id)
	assert.Equal(t, "新标题", fields["title"])
	assert.Equal(t, "新内容", fields["content"])
	assert.Equal(t, "cover.png", fields["cover_image"])
	assert.Equal(t, dep, fields["departure_time"])
	assert.Equal(t, "新起点", fields["pickup_location"])
	assert.Equal(t, 30.1, fields["pickup_lat"])
	assert.Equal(t, 114.1, fields["pickup_lng"])
	assert.Equal(t, "新终点", fields["dropoff_location"])
	assert.Equal(t, 30.2, fields["dropoff_lat"])
	assert.Equal(t, 114.2, fields["dropoff_lng"])
	assert.Equal(t, 15.5, fields["distance_km"])
	assert.Equal(t, 45, fields["duration_min"])
	assert.Equal(t, 3, fields["total_seats"])
	assert.Equal(t, 40.0, fields["price_per_seat"])
	assert.Equal(t, 10.0, fields["toll_fee"])
	assert.Equal(t, uint(7), fields["vehicle_id"])
	assert.Equal(t, uint(8), fields["route_id"])
	assert.Equal(t, model.PaymentMethodAlipay, fields["payment_method"])
	assert.NotNil(t, fields["features"])
	assert.NotNil(t, fields["tags"])
}

func TestUpdate_NoOpWhenNoFields(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	err := svc.Update(1, 100, &dto.UpdatePincheRequest{})
	require.NoError(t, err)
	assert.Empty(t, repo.updateCalls)
}

func TestUpdate_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	repo.updateErr = errors.New("update failed")
	err := svc.Update(1, 100, &dto.UpdatePincheRequest{Title: strPtr("x")})
	require.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
}

// ===== Delete =====

func TestDelete_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Delete(999, 1)
	assert.ErrorIs(t, err, ErrPincheNotFound)
}

func TestDelete_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Delete(1, 1)
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestDelete_NoPermission(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	err := svc.Delete(1, 200)
	assert.ErrorIs(t, err, ErrPincheNoPermission)
}

func TestDelete_Success(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	err := svc.Delete(1, 100)
	require.NoError(t, err)
	require.Len(t, repo.deleteCalls, 1)
	assert.Equal(t, uint(1), repo.deleteCalls[0])
	assert.NotContains(t, repo.byID, uint(1))
}

func TestDelete_RepoDeleteErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100})
	repo.deleteErr = errors.New("delete failed")
	err := svc.Delete(1, 100)
	require.Error(t, err)
	assert.Equal(t, "delete failed", err.Error())
}

// ===== GetByID =====

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newSvc()
	info, err := svc.GetByID(999, 1)
	assert.ErrorIs(t, err, ErrPincheNotFound)
	assert.Nil(t, info)
}

func TestGetByID_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	info, err := svc.GetByID(1, 1)
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, "conn lost", err.Error())
}

func TestGetByID_IncrView(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100, ViewCount: 5})
	info, err := svc.GetByID(1, 200)
	require.NoError(t, err)
	require.NotNil(t, info)
	// 返回 info 浏览量自增
	assert.Equal(t, 6, info.ViewCount)
	require.Len(t, repo.incrView, 1)
	assert.Equal(t, uint(1), repo.incrView[0])
}

func TestGetByID_IncrViewErrorIgnored(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100, ViewCount: 5})
	repo.incrViewErr = errors.New("incr failed")
	// IncrViewCount 错误被忽略，仍返回详情
	info, err := svc.GetByID(1, 200)
	require.NoError(t, err)
	require.NotNil(t, info)
	// 即便仓储自增失败，service 层仍对返回 info 的 ViewCount +1
	assert.Equal(t, 6, info.ViewCount)
}

// ===== List / Nearby / Search / Mine =====

func TestList_PaginationAndResult(t *testing.T) {
	svc, repo := newSvc()
	repo.listReturn = []model.Pinche{{Title: "A"}, {Title: "B"}}
	repo.listTotal = 2
	pg, list, err := svc.List(5, &dto.PincheListRequest{Pagination: utils.Pagination{Page: 1, PageSize: 10}})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "A", list[0].Title)
	assert.Equal(t, int64(2), pg.Total)
}

func TestList_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listErr = errors.New("list failed")
	pg, list, err := svc.List(5, &dto.PincheListRequest{})
	require.Error(t, err)
	assert.Nil(t, pg)
	assert.Nil(t, list)
}

func TestListNearby_DefaultRadiusWhenZero(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyReturn = []model.Pinche{{Title: "N1"}}
	repo.nearbyTotal = 1
	pg, list, err := svc.ListNearby(5, &dto.PincheNearbyRequest{Latitude: 30.0, Longitude: 114.0})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pg.Total)
	// RadiusKm=0 时兜底为 5
	assert.Equal(t, float64(5), repo.lastNearbyRadius)
}

func TestListNearby_KeepsProvidedRadius(t *testing.T) {
	svc, repo := newSvc()
	_, _, err := svc.ListNearby(5, &dto.PincheNearbyRequest{Latitude: 30.0, Longitude: 114.0, RadiusKm: 20})
	require.NoError(t, err)
	assert.Equal(t, float64(20), repo.lastNearbyRadius)
}

func TestListNearby_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.nearbyErr = errors.New("nearby failed")
	pg, list, err := svc.ListNearby(5, &dto.PincheNearbyRequest{Latitude: 30.0, Longitude: 114.0})
	require.Error(t, err)
	assert.Nil(t, pg)
	assert.Nil(t, list)
}

func TestSearch_WithDepartureDate(t *testing.T) {
	svc, repo := newSvc()
	repo.listReturn = []model.Pinche{{Title: "S1"}}
	repo.listTotal = 1
	pg, list, err := svc.Search(5, &dto.PincheSearchRequest{DepartureDate: "2026-08-02", Keyword: "光谷"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pg.Total)
	// 出发日期被展开为当天起止区间
	assert.Equal(t, "2026-08-02 00:00:00", repo.lastListOpts.DepartureFrom)
	assert.Equal(t, "2026-08-02 23:59:59", repo.lastListOpts.DepartureTo)
	assert.Equal(t, "光谷", repo.lastListOpts.Keyword)
	assert.Equal(t, "latest", repo.lastListOpts.Sort)
}

func TestSearch_WithoutDepartureDate(t *testing.T) {
	svc, repo := newSvc()
	repo.listReturn = []model.Pinche{{Title: "S2"}}
	repo.listTotal = 1
	_, _, err := svc.Search(5, &dto.PincheSearchRequest{Keyword: "x"})
	require.NoError(t, err)
	// 未传出发日期时不构建区间
	assert.Empty(t, repo.lastListOpts.DepartureFrom)
	assert.Empty(t, repo.lastListOpts.DepartureTo)
}

func TestSearch_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listErr = errors.New("search failed")
	pg, list, err := svc.Search(5, &dto.PincheSearchRequest{})
	require.Error(t, err)
	assert.Nil(t, pg)
	assert.Nil(t, list)
}

func TestListMine_Result(t *testing.T) {
	svc, repo := newSvc()
	repo.listMineReturn = []model.Pinche{{Title: "M1"}}
	repo.listMineTotal = 1
	pg, list, err := svc.ListMine(100, 1, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(1), pg.Total)
}

func TestListMine_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listMineErr = errors.New("mine failed")
	pg, list, err := svc.ListMine(100, 1, 10)
	require.Error(t, err)
	assert.Nil(t, pg)
	assert.Nil(t, list)
}

// ===== Match =====

// matchPinche 构造一条车主行程（已发布、有座位）。
func matchPinche(id uint, pickupLat, pickupLng, dropLat, dropLng float64, price float64, dep *time.Time) model.Pinche {
	return model.Pinche{
		UserID:          id * 10,
		Role:            model.RoleDriver,
		Status:          model.PincheStatusPublished,
		AuditStatus:     model.PincheAuditApproved,
		AvailableSeats:  4,
		PricePerSeat:    price,
		PickupLat:       pickupLat,
		PickupLng:       pickupLng,
		DropoffLat:      dropLat,
		DropoffLng:      dropLng,
		DepartureTime:   dep,
	}
}

func TestMatch_EmptyList(t *testing.T) {
	svc, repo := newSvc()
	repo.listMatchReturn = nil
	repo.listMatchTotal = 0
	resp, err := svc.Match(5, &dto.PincheMatchRequest{
		PickupLocation:  "武昌",
		PickupLat:       30.0,
		PickupLng:       114.0,
		DropoffLocation: "光谷",
		DropoffLat:      30.1,
		DropoffLng:      114.1,
		Seats:           1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.List)
}

func TestMatch_DefaultMaxRadiusIncludes(t *testing.T) {
	svc, repo := newSvc()
	// MaxRadiusKm=0 兜底为 10；起点距离约 5.5km（<10），终点同点（0km），应被纳入
	repo.listMatchReturn = []model.Pinche{matchPinche(1, 30.05, 114.0, 30.1, 114.1, 30, nil)}
	repo.listMatchTotal = 1
	resp, err := svc.Match(5, &dto.PincheMatchRequest{
		PickupLocation:  "武昌",
		PickupLat:       30.0,
		PickupLng:       114.0,
		DropoffLocation: "光谷",
		DropoffLat:      30.1,
		DropoffLng:      114.1,
		Seats:           1,
		MaxPrice:        100,
	})
	require.NoError(t, err)
	require.Len(t, resp.List, 1)
	assert.Greater(t, resp.List[0].MatchScore, float64(0))
}

func TestMatch_SmallRadiusExcludes(t *testing.T) {
	svc, repo := newSvc()
	// MaxRadiusKm=1；起点距离约 5.5km（>1）→ 匹配度 0 → 过滤
	repo.listMatchReturn = []model.Pinche{matchPinche(1, 30.05, 114.0, 30.1, 114.1, 30, nil)}
	repo.listMatchTotal = 1
	resp, err := svc.Match(5, &dto.PincheMatchRequest{
		PickupLocation:  "武昌",
		PickupLat:       30.0,
		PickupLng:       114.0,
		DropoffLocation: "光谷",
		DropoffLat:      30.1,
		DropoffLng:      114.1,
		Seats:           1,
		MaxRadiusKm:     1,
	})
	require.NoError(t, err)
	assert.Empty(t, resp.List)
	// Total 取自 ListMatch 的 total，而非过滤后数量
	assert.Equal(t, 1, resp.Total)
}

func TestMatch_ScoresAndSorting(t *testing.T) {
	svc, repo := newSvc()
	// A：起终点与请求完全重合 → 高分；B：起点约 5.5km → 较低分
	// 仓储返回顺序为 [B, A]，验证输出按分数降序排列为 [A, B]
	A := matchPinche(1, 30.0, 114.0, 30.1, 114.1, 30, nil)
	B := matchPinche(2, 30.05, 114.0, 30.1, 114.1, 30, nil)
	repo.listMatchReturn = []model.Pinche{B, A}
	repo.listMatchTotal = 2
	resp, err := svc.Match(5, &dto.PincheMatchRequest{
		PickupLocation:  "武昌",
		PickupLat:       30.0,
		PickupLng:       114.0,
		DropoffLocation: "光谷",
		DropoffLat:      30.1,
		DropoffLng:      114.1,
		Seats:           1,
		MaxPrice:        100,
	})
	require.NoError(t, err)
	require.Len(t, resp.List, 2)
	// A 排前
	assert.Equal(t, uint(10), resp.List[0].PincheInfo.UserID)
	assert.Equal(t, uint(20), resp.List[1].PincheInfo.UserID)
	assert.Greater(t, resp.List[0].MatchScore, resp.List[1].MatchScore)
	// A 满分：起点 30 + 终点 30 + 无出发时间 10 + 座位 10 + 价格 10 = 90
	assert.InDelta(t, 90.0, resp.List[0].MatchScore, 0.01)
}

func TestMatch_ListMatchErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.listMatchErr = errors.New("match failed")
	resp, err := svc.Match(5, &dto.PincheMatchRequest{
		PickupLocation:  "武昌",
		PickupLat:       30.0,
		PickupLng:       114.0,
		DropoffLocation: "光谷",
		DropoffLat:      30.1,
		DropoffLng:      114.1,
	})
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "match failed", err.Error())
}

// ===== calcMatchScore / haversineDistance =====

func TestCalcMatchScore_PickupTooFar_ReturnsZero(t *testing.T) {
	svc := &pincheService{repo: newMockPincheRepo()}
	req := &dto.PincheMatchRequest{
		PickupLat: 30.0, PickupLng: 114.0,
		DropoffLat: 30.1, DropoffLng: 114.1,
		Seats: 1,
	}
	// 起点相距约 111km（>10）→ 0
	p := &model.Pinche{PickupLat: 31.0, PickupLng: 114.0, DropoffLat: 30.1, DropoffLng: 114.1, AvailableSeats: 4, PricePerSeat: 30}
	score, reasons := svc.calcMatchScore(p, req, 10)
	assert.Equal(t, float64(0), score)
	assert.Nil(t, reasons)
}

func TestCalcMatchScore_DropoffTooFar_ReturnsZero(t *testing.T) {
	svc := &pincheService{repo: newMockPincheRepo()}
	req := &dto.PincheMatchRequest{
		PickupLat: 30.0, PickupLng: 114.0,
		DropoffLat: 30.1, DropoffLng: 114.1,
		Seats: 1,
	}
	// 起点同点（0km <10），但终点相距约 111km（>20）→ 0
	p := &model.Pinche{PickupLat: 30.0, PickupLng: 114.0, DropoffLat: 31.1, DropoffLng: 114.1, AvailableSeats: 4, PricePerSeat: 30}
	score, reasons := svc.calcMatchScore(p, req, 10)
	assert.Equal(t, float64(0), score)
	assert.Nil(t, reasons)
}

func TestCalcMatchScore_NoDepartureTime_AddsTen(t *testing.T) {
	svc := &pincheService{repo: newMockPincheRepo()}
	req := &dto.PincheMatchRequest{
		PickupLat: 30.0, PickupLng: 114.0,
		DropoffLat: 30.1, DropoffLng: 114.1,
		Seats: 1, MaxPrice: 100,
	}
	// 请求与行程均无出发时间 → 走 else 分支 +10，但不追加"出发时间相近"reason
	p := &model.Pinche{PickupLat: 30.0, PickupLng: 114.0, DropoffLat: 30.1, DropoffLng: 114.1, AvailableSeats: 4, PricePerSeat: 30}
	score, reasons := svc.calcMatchScore(p, req, 10)
	// 起点 30 + 终点 30 + 无出发时间 10 + 座位 10 + 价格 10 = 90
	assert.InDelta(t, 90.0, score, 0.01)
	require.Contains(t, reasons, "起点相近")
	require.Contains(t, reasons, "终点相近")
	require.NotContains(t, reasons, "出发时间相近") // 无出发时间分支不追加该 reason
}

func TestCalcMatchScore_DepartureTimeClose_AddsReason(t *testing.T) {
	svc := &pincheService{repo: newMockPincheRepo()}
	now := time.Now()
	req := &dto.PincheMatchRequest{
		PickupLat: 30.0, PickupLng: 114.0,
		DropoffLat: 30.1, DropoffLng: 114.1,
		DepartureTime: &now,
		Seats: 1, MaxPrice: 100,
	}
	// 行程出发时间与请求接近（1 小时内）→ 走 if 分支并追加 reason
	dep := now.Add(1 * time.Hour)
	p := &model.Pinche{PickupLat: 30.0, PickupLng: 114.0, DropoffLat: 30.1, DropoffLng: 114.1, DepartureTime: &dep, AvailableSeats: 4, PricePerSeat: 30}
	score, reasons := svc.calcMatchScore(p, req, 10)
	assert.Greater(t, score, float64(0))
	require.Contains(t, reasons, "出发时间相近")
}

func TestCalcMatchScore_ScoreBelowThreshold_ReturnsZero(t *testing.T) {
	svc := &pincheService{repo: newMockPincheRepo()}
	// 构造分数 < 30 的场景：起点贴 10km 上限（低分）、终点贴 20km 上限（低分）、
	// 无出发时间 else +10、座位不足不加分、价格 MaxPrice=0 +5
	// 起点 ~9km → 30*(1-9/10)=3；终点 ~19km → 30*(1-19/20)=1.5；
	// 合计 ≈ 3+1.5+10+5 = 19.5 < 30 → 过滤
	req := &dto.PincheMatchRequest{
		PickupLat: 30.0, PickupLng: 114.0,
		DropoffLat: 30.1, DropoffLng: 114.1,
		Seats: 2,
	}
	p := &model.Pinche{
		PickupLat: 30.081, PickupLng: 114.0,    // 约 9km（< 10 上限）
		DropoffLat: 30.271, DropoffLng: 114.1,  // 约 19km（< 2*10=20 上限）
		AvailableSeats: 0, PricePerSeat: 999,
	}
	score, reasons := svc.calcMatchScore(p, req, 10)
	assert.Equal(t, float64(0), score)
	assert.Nil(t, reasons)
}

func TestHaversineDistance_SamePointZero(t *testing.T) {
	assert.Equal(t, float64(0), haversineDistance(30.0, 114.0, 30.0, 114.0))
}

func TestHaversineDistance_KnownDistance(t *testing.T) {
	// 纬度相差 1 度 ≈ 111km
	d := haversineDistance(30.0, 114.0, 31.0, 114.0)
	assert.InDelta(t, 111.0, d, 1.0)
}

// ===== 互动 =====

func TestIncrContact_Delegates(t *testing.T) {
	svc, repo := newSvc()
	err := svc.IncrContact(7)
	require.NoError(t, err)
	require.Len(t, repo.incrContact, 1)
	assert.Equal(t, uint(7), repo.incrContact[0])
}

func TestIncrContact_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.incrContactErr = errors.New("incr failed")
	err := svc.IncrContact(7)
	require.Error(t, err)
}

func TestIncrShare_Delegates(t *testing.T) {
	svc, repo := newSvc()
	err := svc.IncrShare(8)
	require.NoError(t, err)
	require.Len(t, repo.incrShare, 1)
	assert.Equal(t, uint(8), repo.incrShare[0])
}

func TestIncrShare_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.incrShareErr = errors.New("incr failed")
	err := svc.IncrShare(8)
	require.Error(t, err)
}

func TestRecordView_IncrViewAndNilError(t *testing.T) {
	svc, repo := newSvc()
	// RecordView 恒返回 nil，且调用 IncrViewCount
	err := svc.RecordView(100, &dto.PincheViewRequest{PincheID: 9, Source: "detail"})
	require.NoError(t, err)
	require.Len(t, repo.incrView, 1)
	assert.Equal(t, uint(9), repo.incrView[0])
}

func TestRecordView_IncrViewErrorIgnored(t *testing.T) {
	svc, repo := newSvc()
	repo.incrViewErr = errors.New("incr failed")
	// IncrViewCount 错误被忽略，RecordView 仍返回 nil
	err := svc.RecordView(100, &dto.PincheViewRequest{PincheID: 9})
	require.NoError(t, err)
}

// ===== Admin =====

func TestAdminList_Result(t *testing.T) {
	svc, repo := newSvc()
	repo.adminListReturn = []model.Pinche{{Title: "A1"}, {Title: "A2"}}
	repo.adminListTotal = 2
	pg, list, err := svc.AdminList(&dto.PincheAdminListRequest{Pagination: utils.Pagination{Page: 1, PageSize: 10}})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, int64(2), pg.Total)
}

func TestAdminList_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.adminListErr = errors.New("admin list failed")
	pg, list, err := svc.AdminList(&dto.PincheAdminListRequest{})
	require.Error(t, err)
	assert.Nil(t, pg)
	assert.Nil(t, list)
}

func TestAdminGetByID_NotFound(t *testing.T) {
	svc, _ := newSvc()
	info, err := svc.AdminGetByID(999)
	assert.ErrorIs(t, err, ErrPincheNotFound)
	assert.Nil(t, info)
}

func TestAdminGetByID_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	info, err := svc.AdminGetByID(1)
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, "conn lost", err.Error())
}

func TestAdminGetByID_Success(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{UserID: 100, Title: "详情"})
	info, err := svc.AdminGetByID(1)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "详情", info.Title)
	// AdminGetByID 不自增浏览量
	assert.Empty(t, repo.incrView)
}

func TestAudit_NotFound(t *testing.T) {
	svc, _ := newSvc()
	err := svc.Audit(999, model.PincheAuditApproved, "")
	assert.ErrorIs(t, err, ErrPincheNotFound)
}

func TestAudit_FindErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.findErr = errors.New("conn lost")
	err := svc.Audit(1, model.PincheAuditApproved, "")
	require.Error(t, err)
	assert.Equal(t, "conn lost", err.Error())
}

func TestAudit_AlreadyAudited(t *testing.T) {
	svc, repo := newSvc()
	// 非待审状态（已通过）→ 拒绝重复审核
	repo.seed(&model.Pinche{AuditStatus: model.PincheAuditApproved})
	err := svc.Audit(1, model.PincheAuditRejected, "违规")
	assert.ErrorIs(t, err, ErrPincheAudited)
	assert.Empty(t, repo.auditCalls)
}

func TestAudit_Success(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{AuditStatus: model.PincheAuditPending})
	err := svc.Audit(1, model.PincheAuditApproved, "")
	require.NoError(t, err)
	require.Len(t, repo.auditCalls, 1)
	assert.Equal(t, uint(1), repo.auditCalls[0].id)
	assert.Equal(t, model.PincheAuditApproved, repo.auditCalls[0].auditStatus)
}

func TestAudit_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.seed(&model.Pinche{AuditStatus: model.PincheAuditPending})
	repo.updateAuditErr = errors.New("audit failed")
	err := svc.Audit(1, model.PincheAuditApproved, "")
	require.Error(t, err)
	assert.Equal(t, "audit failed", err.Error())
}

func TestAdminUpdateStatus_Delegates(t *testing.T) {
	svc, repo := newSvc()
	err := svc.AdminUpdateStatus(3, model.PincheStatusCancelled)
	require.NoError(t, err)
	require.Len(t, repo.statusCalls, 1)
	assert.Equal(t, uint(3), repo.statusCalls[0].id)
	assert.Equal(t, model.PincheStatusCancelled, repo.statusCalls[0].status)
}

func TestAdminUpdateStatus_ErrorPropagated(t *testing.T) {
	svc, repo := newSvc()
	repo.updateStatusErr = errors.New("status failed")
	err := svc.AdminUpdateStatus(3, model.PincheStatusCancelled)
	require.Error(t, err)
	assert.Equal(t, "status failed", err.Error())
}

// ===== 批量操作 =====

func TestBatchAudit_AllSuccess(t *testing.T) {
	svc, _ := newSvc()
	resp, err := svc.BatchAudit(&dto.BatchAuditRequest{IDs: []uint{1, 2, 3}, AuditStatus: model.PincheAuditApproved})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 3, resp.Success)
	assert.Equal(t, 0, resp.Failed)
	assert.Empty(t, resp.FailedIDs)
}

func TestBatchAudit_PartialFailure(t *testing.T) {
	svc, repo := newSvc()
	// 仅 id=2 失败
	repo.failIDs[2] = true
	resp, err := svc.BatchAudit(&dto.BatchAuditRequest{IDs: []uint{1, 2, 3}, AuditStatus: model.PincheAuditRejected, AuditReason: "违规"})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 2, resp.Success)
	assert.Equal(t, 1, resp.Failed)
	require.Equal(t, []uint{2}, resp.FailedIDs)
}

func TestBatchUpdateStatus_AllSuccess(t *testing.T) {
	svc, _ := newSvc()
	resp, err := svc.BatchUpdateStatus(&dto.BatchStatusUpdateRequest{IDs: []uint{1, 2}, Status: model.PincheStatusFinished})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 2, resp.Success)
	assert.Equal(t, 0, resp.Failed)
}

func TestBatchUpdateStatus_PartialFailure(t *testing.T) {
	svc, repo := newSvc()
	// 仅 id=1 失败
	repo.failIDs[1] = true
	resp, err := svc.BatchUpdateStatus(&dto.BatchStatusUpdateRequest{IDs: []uint{1, 2}, Status: model.PincheStatusCancelled})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 1, resp.Success)
	assert.Equal(t, 1, resp.Failed)
	require.Equal(t, []uint{1}, resp.FailedIDs)
}

func TestBatchDelete_AllSuccess(t *testing.T) {
	svc, repo := newSvc()
	resp, err := svc.BatchDelete(&dto.BatchDeleteRequest{IDs: []uint{1, 2, 3}})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 3, resp.Success)
	assert.Equal(t, 0, resp.Failed)
	require.Len(t, repo.deleteCalls, 3)
}

func TestBatchDelete_PartialFailure(t *testing.T) {
	svc, repo := newSvc()
	// 仅 id=3 失败
	repo.failIDs[3] = true
	resp, err := svc.BatchDelete(&dto.BatchDeleteRequest{IDs: []uint{1, 2, 3}})
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 2, resp.Success)
	assert.Equal(t, 1, resp.Failed)
	require.Equal(t, []uint{3}, resp.FailedIDs)
}
