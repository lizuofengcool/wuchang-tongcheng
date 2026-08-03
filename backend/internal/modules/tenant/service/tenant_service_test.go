// Package service 多租户分站中台业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖：分站 CRUD/启停/配置复制、配置 Upsert/批量获取、
// 域名绑定/主域名切换/SSL 状态/主域名删除提升、员工 CRUD/角色/重复校验等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/modules/tenant/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ============================================================
// mockStationRepo
// ============================================================

type mockStationRepo struct {
	byID           map[uint]*model.Station
	byRegion       map[uint]*model.Station
	byDomain       map[string]*model.Station
	nextID         uint
	createErr      error
	findErr        error
	findRegionErr  error
	findDomainErr  error
	updateErr      error
	updateFieldsErr error
	deleteErr      error
	listErr        error
	updatedFields  map[uint]map[string]interface{}
	deletedIDs     map[uint]bool
}

func newMockStationRepo() *mockStationRepo {
	return &mockStationRepo{
		byID:          make(map[uint]*model.Station),
		byRegion:      make(map[uint]*model.Station),
		byDomain:      make(map[string]*model.Station),
		nextID:        1,
		updatedFields: make(map[uint]map[string]interface{}),
		deletedIDs:    make(map[uint]bool),
	}
}

func (m *mockStationRepo) Create(s *model.Station) error {
	if m.createErr != nil {
		return m.createErr
	}
	s.ID = m.nextID
	m.nextID++
	cp := *s
	m.byID[s.ID] = &cp
	m.byRegion[s.RegionID] = &cp
	if s.Domain != "" {
		m.byDomain[s.Domain] = &cp
	}
	return nil
}

func (m *mockStationRepo) FindByID(id uint) (*model.Station, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockStationRepo) FindByRegionID(regionID uint) (*model.Station, error) {
	if m.findRegionErr != nil {
		return nil, m.findRegionErr
	}
	s, ok := m.byRegion[regionID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockStationRepo) FindByDomain(domain string) (*model.Station, error) {
	if m.findDomainErr != nil {
		return nil, m.findDomainErr
	}
	s, ok := m.byDomain[domain]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockStationRepo) Update(s *model.Station) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *mockStationRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updatedFields[id] = fields
	s, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["name"]; ok {
		s.Name = v.(string)
	}
	if v, ok := fields["domain"]; ok {
		old := s.Domain
		s.Domain = v.(string)
		if old != "" {
			delete(m.byDomain, old)
		}
		if s.Domain != "" {
			m.byDomain[s.Domain] = s
		}
	}
	if v, ok := fields["status"]; ok {
		s.Status = v.(int)
	}
	return nil
}

func (m *mockStationRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	s, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	delete(m.byRegion, s.RegionID)
	delete(m.byDomain, s.Domain)
	m.deletedIDs[id] = true
	return nil
}

func (m *mockStationRepo) List(pagination *utils.Pagination, opts repository.StationListOptions) ([]model.Station, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Station
	for _, s := range m.byID {
		if opts.RegionID > 0 && s.RegionID != opts.RegionID {
			continue
		}
		if opts.Status != nil && s.Status != *opts.Status {
			continue
		}
		list = append(list, *s)
	}
	return list, int64(len(list)), nil
}

// ============================================================
// mockConfigRepo
// ============================================================

type mockConfigRepo struct {
	byID            map[uint]*model.Config
	byKey           map[string]*model.Config // key = stationID:bizModule:configKey
	nextID          uint
	createErr       error
	findErr         error
	findByKeyErr    error
	upsertErr       error
	updateFieldsErr error
	deleteErr       error
	deleteByStatErr error
	listErr         error
	listByModErr    error
	batchGetErr     error
	deletedByStat   int
}

func newMockConfigRepo() *mockConfigRepo {
	return &mockConfigRepo{
		byID:   make(map[uint]*model.Config),
		byKey:  make(map[string]*model.Config),
		nextID: 1,
	}
}

func cfgKey(stationID uint, bizModule, configKey string) string {
	return fmt.Sprintf("%d/%s/%s", stationID, bizModule, configKey)
}

func (m *mockConfigRepo) Create(c *model.Config) error {
	if m.createErr != nil {
		return m.createErr
	}
	c.ID = m.nextID
	m.nextID++
	cp := *c
	m.byID[c.ID] = &cp
	m.byKey[cfgKey(c.StationID, c.BizModule, c.ConfigKey)] = &cp
	return nil
}

func (m *mockConfigRepo) FindByID(id uint) (*model.Config, error) {
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

func (m *mockConfigRepo) FindByStationAndKey(stationID uint, bizModule, configKey string) (*model.Config, error) {
	if m.findByKeyErr != nil {
		return nil, m.findByKeyErr
	}
	c, ok := m.byKey[cfgKey(stationID, bizModule, configKey)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockConfigRepo) Upsert(c *model.Config) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	if existing, ok := m.byKey[cfgKey(c.StationID, c.BizModule, c.ConfigKey)]; ok {
		existing.ConfigValue = c.ConfigValue
		return nil
	}
	return m.Create(c)
}

func (m *mockConfigRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	c, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["config_value"]; ok {
		c.ConfigValue = v.(string)
	}
	return nil
}

func (m *mockConfigRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	c, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	delete(m.byKey, cfgKey(c.StationID, c.BizModule, c.ConfigKey))
	return nil
}

func (m *mockConfigRepo) DeleteByStation(stationID uint, bizModule string) error {
	if m.deleteByStatErr != nil {
		return m.deleteByStatErr
	}
	for id, c := range m.byID {
		if c.StationID == stationID && (bizModule == "" || c.BizModule == bizModule) {
			delete(m.byID, id)
			delete(m.byKey, cfgKey(c.StationID, c.BizModule, c.ConfigKey))
			m.deletedByStat++
		}
	}
	return nil
}

func (m *mockConfigRepo) List(pagination *utils.Pagination, opts repository.ConfigListOptions) ([]model.Config, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Config
	for _, c := range m.byID {
		if opts.StationID > 0 && c.StationID != opts.StationID {
			continue
		}
		if opts.BizModule != "" && c.BizModule != opts.BizModule {
			continue
		}
		list = append(list, *c)
	}
	return list, int64(len(list)), nil
}

func (m *mockConfigRepo) ListByStationAndModule(stationID uint, bizModule string) ([]model.Config, error) {
	if m.listByModErr != nil {
		return nil, m.listByModErr
	}
	var list []model.Config
	for _, c := range m.byID {
		if c.StationID == stationID && c.BizModule == bizModule {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *mockConfigRepo) BatchGet(stationID uint, bizModule string, keys []string) ([]model.Config, error) {
	if m.batchGetErr != nil {
		return nil, m.batchGetErr
	}
	want := make(map[string]bool)
	for _, k := range keys {
		want[k] = true
	}
	var list []model.Config
	for _, c := range m.byID {
		if c.StationID == stationID && c.BizModule == bizModule && (len(keys) == 0 || want[c.ConfigKey]) {
			list = append(list, *c)
		}
	}
	return list, nil
}

// ============================================================
// mockDomainRepo
// ============================================================

type mockDomainRepo struct {
	byID            map[uint]*model.Domain
	byDomain        map[string]*model.Domain
	nextID          uint
	createErr       error
	findErr         error
	findDomainErr   error
	updateErr       error
	updateFieldsErr error
	deleteErr       error
	listErr         error
	listByStatErr   error
	clearPrimErr    error
	clearedPrimary  map[uint]int
	updatedFields   map[uint]map[string]interface{}
}

func newMockDomainRepo() *mockDomainRepo {
	return &mockDomainRepo{
		byID:          make(map[uint]*model.Domain),
		byDomain:      make(map[string]*model.Domain),
		nextID:        1,
		clearedPrimary: make(map[uint]int),
		updatedFields: make(map[uint]map[string]interface{}),
	}
}

func (m *mockDomainRepo) Create(d *model.Domain) error {
	if m.createErr != nil {
		return m.createErr
	}
	d.ID = m.nextID
	m.nextID++
	cp := *d
	m.byID[d.ID] = &cp
	m.byDomain[d.Domain] = &cp
	return nil
}

func (m *mockDomainRepo) FindByID(id uint) (*model.Domain, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	d, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *d
	return &cp, nil
}

func (m *mockDomainRepo) FindByDomain(domain string) (*model.Domain, error) {
	if m.findDomainErr != nil {
		return nil, m.findDomainErr
	}
	d, ok := m.byDomain[domain]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *d
	return &cp, nil
}

func (m *mockDomainRepo) Update(d *model.Domain) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *d
	m.byID[d.ID] = &cp
	return nil
}

func (m *mockDomainRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updatedFields[id] = fields
	d, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["ssl_status"]; ok {
		d.SSLStatus = v.(string)
	}
	if v, ok := fields["is_primary"]; ok {
		d.IsPrimary = v.(bool)
	}
	return nil
}

func (m *mockDomainRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	d, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	delete(m.byDomain, d.Domain)
	return nil
}

func (m *mockDomainRepo) List(pagination *utils.Pagination, opts repository.DomainListOptions) ([]model.Domain, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Domain
	for _, d := range m.byID {
		if opts.StationID > 0 && d.StationID != opts.StationID {
			continue
		}
		if opts.SSLStatus != "" && d.SSLStatus != opts.SSLStatus {
			continue
		}
		list = append(list, *d)
	}
	return list, int64(len(list)), nil
}

func (m *mockDomainRepo) ListByStation(stationID uint) ([]model.Domain, error) {
	if m.listByStatErr != nil {
		return nil, m.listByStatErr
	}
	var list []model.Domain
	for _, d := range m.byID {
		if d.StationID == stationID {
			list = append(list, *d)
		}
	}
	return list, nil
}

func (m *mockDomainRepo) ClearPrimary(stationID uint) error {
	if m.clearPrimErr != nil {
		return m.clearPrimErr
	}
	m.clearedPrimary[stationID]++
	for _, d := range m.byID {
		if d.StationID == stationID {
			d.IsPrimary = false
		}
	}
	return nil
}

// ============================================================
// mockStaffRepo
// ============================================================

type mockStaffRepo struct {
	byID            map[uint]*model.Staff
	byStatUser      map[uint]*model.Staff // key = stationID*1000000 + userID
	nextID          uint
	createErr       error
	findErr         error
	findByStatUsrErr error
	updateErr       error
	updateFieldsErr  error
	deleteErr       error
	listErr         error
	listByStatErr   error
	listByUserErr   error
}

func newMockStaffRepo() *mockStaffRepo {
	return &mockStaffRepo{
		byID:       make(map[uint]*model.Staff),
		byStatUser: make(map[uint]*model.Staff),
		nextID:     1,
	}
}

func statUserKey(stationID, userID uint) uint { return stationID*1000000 + userID }

func (m *mockStaffRepo) Create(s *model.Staff) error {
	if m.createErr != nil {
		return m.createErr
	}
	s.ID = m.nextID
	m.nextID++
	cp := *s
	m.byID[s.ID] = &cp
	m.byStatUser[statUserKey(s.StationID, s.UserID)] = &cp
	return nil
}

func (m *mockStaffRepo) FindByID(id uint) (*model.Staff, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockStaffRepo) FindByStationAndUser(stationID, userID uint) (*model.Staff, error) {
	if m.findByStatUsrErr != nil {
		return nil, m.findByStatUsrErr
	}
	s, ok := m.byStatUser[statUserKey(stationID, userID)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *mockStaffRepo) Update(s *model.Staff) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *mockStaffRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	s, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["role"]; ok {
		s.Role = v.(string)
	}
	if v, ok := fields["status"]; ok {
		s.Status = v.(int)
	}
	return nil
}

func (m *mockStaffRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	s, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	delete(m.byStatUser, statUserKey(s.StationID, s.UserID))
	return nil
}

func (m *mockStaffRepo) List(pagination *utils.Pagination, opts repository.StaffListOptions) ([]model.Staff, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Staff
	for _, s := range m.byID {
		if opts.StationID > 0 && s.StationID != opts.StationID {
			continue
		}
		if opts.UserID > 0 && s.UserID != opts.UserID {
			continue
		}
		if opts.Role != "" && s.Role != opts.Role {
			continue
		}
		if opts.Status != nil && s.Status != *opts.Status {
			continue
		}
		list = append(list, *s)
	}
	return list, int64(len(list)), nil
}

func (m *mockStaffRepo) ListByStation(stationID uint) ([]model.Staff, error) {
	if m.listByStatErr != nil {
		return nil, m.listByStatErr
	}
	var list []model.Staff
	for _, s := range m.byID {
		if s.StationID == stationID {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockStaffRepo) ListByUser(userID uint) ([]model.Staff, error) {
	if m.listByUserErr != nil {
		return nil, m.listByUserErr
	}
	var list []model.Staff
	for _, s := range m.byID {
		if s.UserID == userID {
			list = append(list, *s)
		}
	}
	return list, nil
}

// ============================================================
// StationService 测试
// ============================================================

func TestStationService_Create(t *testing.T) {
	t.Run("success with default status", func(t *testing.T) {
		repo := newMockStationRepo()
		cfg := newMockConfigRepo()
		svc := NewStationService(repo, cfg)
		req := &dto.CreateStationRequest{RegionID: 10, Name: "武昌站", Domain: "wuchang.example.com"}
		info, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, uint(10), info.RegionID)
		assert.Equal(t, "武昌站", info.Name)
		assert.Equal(t, model.StationStatusEnabled, info.Status)
		assert.Equal(t, "已启用", info.StatusText)
	})

	t.Run("status zero defaults to enabled", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		// Status=0（停用）在 service 中被视为未设置，默认为启用
		req := &dto.CreateStationRequest{RegionID: 11, Name: "默认站", Status: model.StationStatusDisabled}
		info, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, model.StationStatusEnabled, info.Status)
		assert.Equal(t, "已启用", info.StatusText)
	})

	t.Run("region already exists", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 20, Name: "站A"})
		_, err := svc.Create(&dto.CreateStationRequest{RegionID: 20, Name: "站B"})
		assert.ErrorIs(t, err, ErrStationRegionExists)
	})

	t.Run("domain already exists", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 21, Name: "站A", Domain: "a.example.com"})
		_, err := svc.Create(&dto.CreateStationRequest{RegionID: 22, Name: "站B", Domain: "a.example.com"})
		assert.ErrorIs(t, err, ErrStationDomainExists)
	})

	t.Run("create repo error", func(t *testing.T) {
		repo := newMockStationRepo()
		repo.createErr = errors.New("db error")
		svc := NewStationService(repo, newMockConfigRepo())
		_, err := svc.Create(&dto.CreateStationRequest{RegionID: 30, Name: "站"})
		assert.Error(t, err)
	})

	t.Run("config parsed", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		req := &dto.CreateStationRequest{RegionID: 40, Name: "配置站", Config: map[string]any{"k": "v"}}
		info, err := svc.Create(req)
		require.NoError(t, err)
		assert.NotNil(t, info.Config)
	})
}

func TestStationService_Update(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewStationService(newMockStationRepo(), newMockConfigRepo())
		err := svc.Update(999, &dto.UpdateStationRequest{})
		assert.ErrorIs(t, err, ErrStationNotFound)
	})

	t.Run("empty fields no-op", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站"})
		err := svc.Update(1, &dto.UpdateStationRequest{})
		require.NoError(t, err)
		assert.NotContains(t, repo.updatedFields, uint(1))
	})

	t.Run("update fields success", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站"})
		newName := "新名称"
		newStatus := model.StationStatusDisabled
		err := svc.Update(1, &dto.UpdateStationRequest{Name: &newName, Status: &newStatus})
		require.NoError(t, err)
		assert.Equal(t, "新名称", repo.updatedFields[1]["name"])
		assert.Equal(t, model.StationStatusDisabled, repo.updatedFields[1]["status"])
	})

	t.Run("domain conflict with another station", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站A", Domain: "a.example.com"})
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "站B", Domain: "b.example.com"})
		newDomain := "a.example.com"
		err := svc.Update(2, &dto.UpdateStationRequest{Domain: &newDomain})
		assert.ErrorIs(t, err, ErrStationDomainExists)
	})

	t.Run("domain unchanged allowed", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站A", Domain: "a.example.com"})
		sameDomain := "a.example.com"
		err := svc.Update(1, &dto.UpdateStationRequest{Domain: &sameDomain})
		require.NoError(t, err)
	})

	t.Run("find error", func(t *testing.T) {
		repo := newMockStationRepo()
		repo.findErr = errors.New("db down")
		svc := NewStationService(repo, newMockConfigRepo())
		err := svc.Update(1, &dto.UpdateStationRequest{})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrStationNotFound)
	})
}

func TestStationService_Delete(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewStationService(newMockStationRepo(), newMockConfigRepo())
		err := svc.Delete(999)
		assert.ErrorIs(t, err, ErrStationNotFound)
	})

	t.Run("success", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站"})
		err := svc.Delete(1)
		require.NoError(t, err)
		assert.True(t, repo.deletedIDs[1])
	})

	t.Run("delete repo error", func(t *testing.T) {
		repo := newMockStationRepo()
		repo.deleteErr = errors.New("db error")
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站"})
		err := svc.Delete(1)
		assert.Error(t, err)
	})
}

func TestStationService_Get(t *testing.T) {
	repo := newMockStationRepo()
	svc := NewStationService(repo, newMockConfigRepo())
	_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 5, Name: "获取站", Domain: "get.example.com"})

	t.Run("get by id", func(t *testing.T) {
		info, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "获取站", info.Name)
	})

	t.Run("get by id not found", func(t *testing.T) {
		_, err := svc.GetByID(999)
		assert.ErrorIs(t, err, ErrStationNotFound)
	})

	t.Run("get by region", func(t *testing.T) {
		info, err := svc.GetByRegionID(5)
		require.NoError(t, err)
		assert.Equal(t, uint(5), info.RegionID)
	})

	t.Run("get by region not found", func(t *testing.T) {
		_, err := svc.GetByRegionID(999)
		assert.ErrorIs(t, err, ErrStationNotFound)
	})

	t.Run("get by domain", func(t *testing.T) {
		info, err := svc.GetByDomain("get.example.com")
		require.NoError(t, err)
		assert.Equal(t, "get.example.com", info.Domain)
	})

	t.Run("get by domain not found", func(t *testing.T) {
		_, err := svc.GetByDomain("none.example.com")
		assert.ErrorIs(t, err, ErrStationNotFoundByDom)
	})

	t.Run("get by empty domain", func(t *testing.T) {
		_, err := svc.GetByDomain("")
		assert.ErrorIs(t, err, ErrStationNotFoundByDom)
	})
}

func TestStationService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站1"})
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "站2"})
		p, list, err := svc.List(&dto.StationListRequest{})
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int64(2), p.Total)
	})

	t.Run("filter by region", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站1"})
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "站2"})
		_, list, err := svc.List(&dto.StationListRequest{RegionID: 1})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("list error", func(t *testing.T) {
		repo := newMockStationRepo()
		repo.listErr = errors.New("db error")
		svc := NewStationService(repo, newMockConfigRepo())
		_, _, err := svc.List(&dto.StationListRequest{})
		assert.Error(t, err)
	})
}

func TestStationService_UpdateStatus(t *testing.T) {
	t.Run("invalid status", func(t *testing.T) {
		svc := NewStationService(newMockStationRepo(), newMockConfigRepo())
		err := svc.UpdateStatus(1, 99)
		assert.ErrorIs(t, err, ErrStationStatusInvalid)
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewStationService(newMockStationRepo(), newMockConfigRepo())
		err := svc.UpdateStatus(999, model.StationStatusDisabled)
		assert.ErrorIs(t, err, ErrStationNotFound)
	})

	t.Run("success", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "站"})
		err := svc.UpdateStatus(1, model.StationStatusDisabled)
		require.NoError(t, err)
		assert.Equal(t, model.StationStatusDisabled, repo.updatedFields[1]["status"])
	})
}

func TestStationService_CopyConfig(t *testing.T) {
	t.Run("same station", func(t *testing.T) {
		svc := NewStationService(newMockStationRepo(), newMockConfigRepo())
		_, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 1, TargetStationID: 1})
		assert.ErrorIs(t, err, ErrCopySameStation)
	})

	t.Run("source not found", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "目标"}) // id=1
		_, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 99, TargetStationID: 1})
		assert.ErrorIs(t, err, ErrCopySourceNotFound)
	})

	t.Run("target not found", func(t *testing.T) {
		repo := newMockStationRepo()
		svc := NewStationService(repo, newMockConfigRepo())
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "源"}) // id=1
		_, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 1, TargetStationID: 99})
		assert.ErrorIs(t, err, ErrCopyTargetNotFound)
	})

	t.Run("copy by module", func(t *testing.T) {
		repo := newMockStationRepo()
		cfg := newMockConfigRepo()
		svc := NewStationService(repo, cfg)
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "源"}) // id=1
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "目标"}) // id=2
		// 源分站配置
		_ = cfg.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		_ = cfg.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k2", ConfigValue: "v2"})
		_ = cfg.Create(&model.Config{StationID: 1, BizModule: "news", ConfigKey: "k3", ConfigValue: "v3"})

		res, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 1, TargetStationID: 2, BizModule: "mall"})
		require.NoError(t, err)
		assert.Equal(t, 2, res.CopiedCount)
		assert.Equal(t, "mall", res.BizModule)
		// 目标分站应有 2 条 mall 配置
		targetCfgs, _ := cfg.ListByStationAndModule(2, "mall")
		assert.Len(t, targetCfgs, 2)
	})

	t.Run("copy all modules", func(t *testing.T) {
		repo := newMockStationRepo()
		cfg := newMockConfigRepo()
		svc := NewStationService(repo, cfg)
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "源"}) // id=1
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "目标"}) // id=2
		_ = cfg.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		_ = cfg.Create(&model.Config{StationID: 1, BizModule: "news", ConfigKey: "k2", ConfigValue: "v2"})

		res, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 1, TargetStationID: 2})
		require.NoError(t, err)
		assert.Equal(t, 2, res.CopiedCount)
	})

	t.Run("delete by station error", func(t *testing.T) {
		repo := newMockStationRepo()
		cfg := newMockConfigRepo()
		cfg.deleteByStatErr = errors.New("db error")
		svc := NewStationService(repo, cfg)
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 1, Name: "源"}) // id=1
		_, _ = svc.Create(&dto.CreateStationRequest{RegionID: 2, Name: "目标"}) // id=2
		_, err := svc.CopyConfig(&dto.CopyConfigRequest{SourceStationID: 1, TargetStationID: 2, BizModule: "mall"})
		assert.Error(t, err)
	})
}

// ============================================================
// ConfigService 测试
// ============================================================

func TestConfigService_Upsert(t *testing.T) {
	t.Run("station not found", func(t *testing.T) {
		svc := NewConfigService(newMockConfigRepo(), newMockStationRepo())
		_, err := svc.Upsert(&dto.UpsertConfigRequest{StationID: 99, BizModule: "mall", ConfigKey: "k"})
		assert.ErrorIs(t, err, ErrConfigStationNotFnd)
	})

	t.Run("success new", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		cfgRepo := newMockConfigRepo()
		svc := NewConfigService(cfgRepo, stationRepo)
		info, err := svc.Upsert(&dto.UpsertConfigRequest{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		require.NoError(t, err)
		assert.Equal(t, "k1", info.ConfigKey)
		assert.Equal(t, "v1", info.ConfigValue)
	})

	t.Run("success update existing", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		cfgRepo := newMockConfigRepo()
		svc := NewConfigService(cfgRepo, stationRepo)
		_, _ = svc.Upsert(&dto.UpsertConfigRequest{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		info, err := svc.Upsert(&dto.UpsertConfigRequest{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v2"})
		require.NoError(t, err)
		assert.Equal(t, "v2", info.ConfigValue)
	})

	t.Run("upsert repo error", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		cfgRepo := newMockConfigRepo()
		cfgRepo.upsertErr = errors.New("db error")
		svc := NewConfigService(cfgRepo, stationRepo)
		_, err := svc.Upsert(&dto.UpsertConfigRequest{StationID: 1, BizModule: "mall", ConfigKey: "k1"})
		assert.Error(t, err)
	})
}

func TestConfigService_Update(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewConfigService(newMockConfigRepo(), newMockStationRepo())
		err := svc.Update(999, &dto.UpdateConfigRequest{})
		assert.ErrorIs(t, err, ErrConfigNotFound)
	})

	t.Run("empty no-op", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k", ConfigValue: "v"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		err := svc.Update(1, &dto.UpdateConfigRequest{})
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k", ConfigValue: "v"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		newVal := "new"
		err := svc.Update(1, &dto.UpdateConfigRequest{ConfigValue: &newVal})
		require.NoError(t, err)
		c, _ := cfgRepo.FindByID(1)
		assert.Equal(t, "new", c.ConfigValue)
	})
}

func TestConfigService_Delete(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewConfigService(newMockConfigRepo(), newMockStationRepo())
		err := svc.Delete(999)
		assert.ErrorIs(t, err, ErrConfigNotFound)
	})

	t.Run("success", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k", ConfigValue: "v"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		err := svc.Delete(1)
		require.NoError(t, err)
		_, err = cfgRepo.FindByID(1)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestConfigService_GetByID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewConfigService(newMockConfigRepo(), newMockStationRepo())
		_, err := svc.GetByID(999)
		assert.ErrorIs(t, err, ErrConfigNotFound)
	})

	t.Run("success", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k", ConfigValue: "v"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		info, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "k", info.ConfigKey)
	})
}

func TestConfigService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "news", ConfigKey: "k2", ConfigValue: "v2"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		p, list, err := svc.List(&dto.ConfigListRequest{StationID: 1})
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int64(2), p.Total)
	})

	t.Run("filter by module", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "news", ConfigKey: "k2", ConfigValue: "v2"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		_, list, err := svc.List(&dto.ConfigListRequest{StationID: 1, BizModule: "mall"})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("list error", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		cfgRepo.listErr = errors.New("db error")
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		_, _, err := svc.List(&dto.ConfigListRequest{})
		assert.Error(t, err)
	})
}

func TestConfigService_ListByStationAndModule(t *testing.T) {
	cfgRepo := newMockConfigRepo()
	_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
	_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "news", ConfigKey: "k2", ConfigValue: "v2"})
	svc := NewConfigService(cfgRepo, newMockStationRepo())
	list, err := svc.ListByStationAndModule(1, "mall")
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestConfigService_BatchGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k1", ConfigValue: "v1"})
		_ = cfgRepo.Create(&model.Config{StationID: 1, BizModule: "mall", ConfigKey: "k2", ConfigValue: "v2"})
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		res, err := svc.BatchGet(&dto.BatchGetConfigRequest{StationID: 1, BizModule: "mall", ConfigKeys: []string{"k1"}})
		require.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "v1", res[0].ConfigValue)
	})

	t.Run("error", func(t *testing.T) {
		cfgRepo := newMockConfigRepo()
		cfgRepo.batchGetErr = errors.New("db error")
		svc := NewConfigService(cfgRepo, newMockStationRepo())
		_, err := svc.BatchGet(&dto.BatchGetConfigRequest{StationID: 1, BizModule: "mall"})
		assert.Error(t, err)
	})
}

// ============================================================
// DomainService 测试
// ============================================================

func TestDomainService_Create(t *testing.T) {
	t.Run("station not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		_, err := svc.Create(&dto.CreateDomainRequest{StationID: 99, Domain: "a.example.com"})
		assert.ErrorIs(t, err, ErrDomainStationNotFnd)
	})

	t.Run("domain exists", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		svc := NewDomainService(domainRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com"})
		_, err := svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com"})
		assert.ErrorIs(t, err, ErrDomainExists)
	})

	t.Run("success default ssl none", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		svc := NewDomainService(newMockDomainRepo(), stationRepo)
		info, err := svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com"})
		require.NoError(t, err)
		assert.Equal(t, model.DomainSSLNone, info.SSLStatus)
		assert.Equal(t, "未配置", info.SSLText)
		assert.False(t, info.IsPrimary)
	})

	t.Run("primary clears others and syncs station", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		svc := NewDomainService(domainRepo, stationRepo)
		_, err := svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com", IsPrimary: true})
		require.NoError(t, err)
		assert.Equal(t, 1, domainRepo.clearedPrimary[1])
		assert.Equal(t, "a.example.com", stationRepo.updatedFields[1]["domain"])
	})

	t.Run("clear primary error", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		domainRepo.clearPrimErr = errors.New("db error")
		svc := NewDomainService(domainRepo, stationRepo)
		_, err := svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com", IsPrimary: true})
		assert.Error(t, err)
	})

	t.Run("create repo error", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		domainRepo.createErr = errors.New("db error")
		svc := NewDomainService(domainRepo, stationRepo)
		_, err := svc.Create(&dto.CreateDomainRequest{StationID: 1, Domain: "a.example.com"})
		assert.Error(t, err)
	})
}

func TestDomainService_Update(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		err := svc.Update(999, &dto.UpdateDomainRequest{})
		assert.ErrorIs(t, err, ErrDomainNotFound)
	})

	t.Run("empty no-op", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		err := svc.Update(1, &dto.UpdateDomainRequest{})
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		newSSL := model.DomainSSLActive
		err := svc.Update(1, &dto.UpdateDomainRequest{SSLStatus: &newSSL})
		require.NoError(t, err)
		assert.Equal(t, model.DomainSSLActive, domainRepo.updatedFields[1]["ssl_status"])
	})
}

func TestDomainService_Delete(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		err := svc.Delete(999)
		assert.ErrorIs(t, err, ErrDomainNotFound)
	})

	t.Run("delete non-primary success", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		err := svc.Delete(1)
		require.NoError(t, err)
	})

	t.Run("delete primary promotes earliest", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		// id=1 主域名，id=2 普通域名
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com", IsPrimary: true})
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "b.example.com", IsPrimary: false})
		svc := NewDomainService(domainRepo, stationRepo)
		err := svc.Delete(1)
		require.NoError(t, err)
		// 提升 id=2 为主域名
		assert.Equal(t, true, domainRepo.updatedFields[2]["is_primary"])
		assert.Equal(t, "b.example.com", stationRepo.updatedFields[1]["domain"])
	})

	t.Run("delete last primary clears station domain", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com", IsPrimary: true})
		svc := NewDomainService(domainRepo, stationRepo)
		err := svc.Delete(1)
		require.NoError(t, err)
		assert.Equal(t, "", stationRepo.updatedFields[1]["domain"])
	})
}

func TestDomainService_GetByID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		_, err := svc.GetByID(999)
		assert.ErrorIs(t, err, ErrDomainNotFound)
	})

	t.Run("success", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com", SSLStatus: model.DomainSSLActive})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		info, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "a.example.com", info.Domain)
		assert.Equal(t, "已生效", info.SSLText)
	})
}

func TestDomainService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		_ = domainRepo.Create(&model.Domain{StationID: 2, Domain: "b.example.com"})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		p, list, err := svc.List(&dto.DomainListRequest{StationID: 1})
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, int64(1), p.Total)
	})

	t.Run("list error", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		domainRepo.listErr = errors.New("db error")
		svc := NewDomainService(domainRepo, newMockStationRepo())
		_, _, err := svc.List(&dto.DomainListRequest{})
		assert.Error(t, err)
	})
}

func TestDomainService_ListByStation(t *testing.T) {
	domainRepo := newMockDomainRepo()
	_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
	_ = domainRepo.Create(&model.Domain{StationID: 2, Domain: "b.example.com"})
	svc := NewDomainService(domainRepo, newMockStationRepo())
	list, err := svc.ListByStation(1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestDomainService_SetPrimary(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		err := svc.SetPrimary(999)
		assert.ErrorIs(t, err, ErrDomainNotFound)
	})

	t.Run("success", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		svc := NewDomainService(domainRepo, stationRepo)
		err := svc.SetPrimary(1)
		require.NoError(t, err)
		assert.Equal(t, 1, domainRepo.clearedPrimary[1])
		assert.Equal(t, true, domainRepo.updatedFields[1]["is_primary"])
		assert.Equal(t, "a.example.com", stationRepo.updatedFields[1]["domain"])
	})

	t.Run("clear primary error", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		domainRepo.clearPrimErr = errors.New("db error")
		svc := NewDomainService(domainRepo, newMockStationRepo())
		err := svc.SetPrimary(1)
		assert.Error(t, err)
	})
}

func TestDomainService_UpdateSSLStatus(t *testing.T) {
	t.Run("invalid status", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		err := svc.UpdateSSLStatus(1, "invalid")
		assert.ErrorIs(t, err, ErrDomainSSLInvalid)
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewDomainService(newMockDomainRepo(), newMockStationRepo())
		err := svc.UpdateSSLStatus(999, model.DomainSSLActive)
		assert.ErrorIs(t, err, ErrDomainNotFound)
	})

	t.Run("success", func(t *testing.T) {
		domainRepo := newMockDomainRepo()
		_ = domainRepo.Create(&model.Domain{StationID: 1, Domain: "a.example.com"})
		svc := NewDomainService(domainRepo, newMockStationRepo())
		err := svc.UpdateSSLStatus(1, model.DomainSSLActive)
		require.NoError(t, err)
		assert.Equal(t, model.DomainSSLActive, domainRepo.updatedFields[1]["ssl_status"])
	})

	t.Run("ssl text mapping", func(t *testing.T) {
		// 覆盖 domainSSLText 各分支
		assert.Equal(t, "未配置", domainSSLText(model.DomainSSLNone))
		assert.Equal(t, "申请中", domainSSLText(model.DomainSSLPending))
		assert.Equal(t, "已生效", domainSSLText(model.DomainSSLActive))
		assert.Equal(t, "失败", domainSSLText(model.DomainSSLFailed))
		assert.Equal(t, "", domainSSLText("unknown"))
	})
}

// ============================================================
// StaffService 测试
// ============================================================

func TestStaffService_Create(t *testing.T) {
	t.Run("station not found", func(t *testing.T) {
		svc := NewStaffService(newMockStaffRepo(), newMockStationRepo())
		_, err := svc.Create(&dto.CreateStaffRequest{StationID: 99, UserID: 1, Role: model.StaffRoleOperator})
		assert.ErrorIs(t, err, ErrStaffStationNotFnd)
	})

	t.Run("staff exists", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		svc := NewStaffService(newMockStaffRepo(), stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		_, err := svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleManager})
		assert.ErrorIs(t, err, ErrStaffExists)
	})

	t.Run("success default status enabled", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		svc := NewStaffService(newMockStaffRepo(), stationRepo)
		info, err := svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleManager})
		require.NoError(t, err)
		assert.Equal(t, model.StaffStatusEnabled, info.Status)
		assert.Equal(t, "已启用", info.StatusText)
		assert.Equal(t, "管理员", info.RoleText)
	})

	t.Run("success operator role text", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		svc := NewStaffService(newMockStaffRepo(), stationRepo)
		// Status=0（停用）在 service 中被视为未设置，默认为启用
		info, err := svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 2, Role: model.StaffRoleOperator, Status: model.StaffStatusDisabled})
		require.NoError(t, err)
		assert.Equal(t, model.StaffStatusEnabled, info.Status)
		assert.Equal(t, "已启用", info.StatusText)
		assert.Equal(t, "运营员", info.RoleText)
	})

	t.Run("create repo error", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		staffRepo.createErr = errors.New("db error")
		svc := NewStaffService(staffRepo, stationRepo)
		_, err := svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		assert.Error(t, err)
	})
}

func TestStaffService_Update(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewStaffService(newMockStaffRepo(), newMockStationRepo())
		err := svc.Update(999, &dto.UpdateStaffRequest{})
		assert.ErrorIs(t, err, ErrStaffNotFound)
	})

	t.Run("empty no-op", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		err := svc.Update(1, &dto.UpdateStaffRequest{})
		require.NoError(t, err)
	})

	t.Run("success update role and status", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		newRole := model.StaffRoleManager
		newStatus := model.StaffStatusDisabled
		err := svc.Update(1, &dto.UpdateStaffRequest{Role: &newRole, Status: &newStatus, Permissions: map[string]any{"p": 1}})
		require.NoError(t, err)
		s, _ := staffRepo.FindByID(1)
		assert.Equal(t, model.StaffRoleManager, s.Role)
		assert.Equal(t, model.StaffStatusDisabled, s.Status)
	})
}

func TestStaffService_Delete(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewStaffService(newMockStaffRepo(), newMockStationRepo())
		err := svc.Delete(999)
		assert.ErrorIs(t, err, ErrStaffNotFound)
	})

	t.Run("success", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		err := svc.Delete(1)
		require.NoError(t, err)
		_, err = staffRepo.FindByID(1)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestStaffService_GetByID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc := NewStaffService(newMockStaffRepo(), newMockStationRepo())
		_, err := svc.GetByID(999)
		assert.ErrorIs(t, err, ErrStaffNotFound)
	})

	t.Run("success", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleManager})
		info, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.UserID)
		assert.Equal(t, "管理员", info.RoleText)
	})
}

func TestStaffService_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 2, Role: model.StaffRoleManager})
		p, list, err := svc.List(&dto.StaffListRequest{StationID: 1})
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int64(2), p.Total)
	})

	t.Run("filter by role", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		svc := NewStaffService(staffRepo, stationRepo)
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
		_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 2, Role: model.StaffRoleManager})
		_, list, err := svc.List(&dto.StaffListRequest{StationID: 1, Role: model.StaffRoleManager})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})

	t.Run("list error", func(t *testing.T) {
		stationRepo := newMockStationRepo()
		_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站"})
		staffRepo := newMockStaffRepo()
		staffRepo.listErr = errors.New("db error")
		svc := NewStaffService(staffRepo, stationRepo)
		_, _, err := svc.List(&dto.StaffListRequest{})
		assert.Error(t, err)
	})
}

func TestStaffService_ListByStation(t *testing.T) {
	stationRepo := newMockStationRepo()
	_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站A"})
	_ = stationRepo.Create(&model.Station{RegionID: 2, Name: "站B"})
	staffRepo := newMockStaffRepo()
	svc := NewStaffService(staffRepo, stationRepo)
	_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 1, Role: model.StaffRoleOperator})
	_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 2, UserID: 2, Role: model.StaffRoleManager})
	list, err := svc.ListByStation(1)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestStaffService_ListByUser(t *testing.T) {
	stationRepo := newMockStationRepo()
	_ = stationRepo.Create(&model.Station{RegionID: 1, Name: "站A"})
	_ = stationRepo.Create(&model.Station{RegionID: 2, Name: "站B"})
	staffRepo := newMockStaffRepo()
	svc := NewStaffService(staffRepo, stationRepo)
	// 同一用户属于两个分站
	_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 1, UserID: 5, Role: model.StaffRoleOperator})
	_, _ = svc.Create(&dto.CreateStaffRequest{StationID: 2, UserID: 5, Role: model.StaffRoleManager})
	list, err := svc.ListByUser(5)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
