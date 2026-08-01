// Package service 同城二手物品业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖：发布默认值与过期时间、用户隔离更新/删除、
// 详情浏览量+图片拼装+收藏态+留言已读、列表/附近/搜索/我的发布分页、
// 收藏 toggle 语义与计数、留言创建+计数、管理后台列表/详情、
// 审核自动发布/拒绝下架、强制下架/恢复等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockErshouRepo =====

type favKey struct {
	userID, ershouID uint
}

type mockErshouRepo struct {
	byID     map[uint]*model.Ershou
	nextID   uint
	favs     map[favKey]bool
	images   map[uint][]model.ErshouImage
	messages []model.ErshouMessage

	// 可注入错误
	createErr        error
	findErr          error
	updateFieldsErr  error
	deleteErr        error
	listErr          error
	adminListErr     error
	nearbyErr        error
	searchErr        error
	incrViewErr      error
	listImagesErr    error
	replaceImagesErr error
	favExistsErr     error
	createFavErr     error
	deleteFavErr     error
	incrFavErr       error
	decrFavErr       error
	listFavsErr      error
	createMsgErr     error
	listMsgErr       error
	incrMsgErr       error
	markMsgReadErr   error
	listByUserErr    error

	// 调用计数
	incrViewCountCalls   []uint
	incrFavCountCalls    []uint
	decrFavCountCalls    []uint
	incrMsgCountCalls    []uint
	markMsgReadCalls     []struct {
		ershouID uint
		userID   uint
	}
	replaceImagesCalls []struct {
		ershouID uint
		urls     []string
	}
	updateFieldsCalls []struct {
		id     uint
		fields map[string]interface{}
	}
}

func newMockErshouRepo() *mockErshouRepo {
	return &mockErshouRepo{
		byID:     make(map[uint]*model.Ershou),
		nextID:   1,
		favs:     make(map[favKey]bool),
		images:   make(map[uint][]model.ErshouImage),
		messages: []model.ErshouMessage{},
	}
}

func (m *mockErshouRepo) Create(e *model.Ershou) error {
	if m.createErr != nil {
		return m.createErr
	}
	e.ID = m.nextID
	m.nextID++
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}

func (m *mockErshouRepo) FindByID(id uint) (*model.Ershou, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	e, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *e
	return &cp, nil
}

func (m *mockErshouRepo) Update(e *model.Ershou) error {
	cp := *e
	m.byID[e.ID] = &cp
	return nil
}

func (m *mockErshouRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateFieldsErr != nil {
		return m.updateFieldsErr
	}
	m.updateFieldsCalls = append(m.updateFieldsCalls, struct {
		id     uint
		fields map[string]interface{}
	}{id, fields})
	e, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["status"]; ok {
		e.Status = v.(int)
	}
	if v, ok := fields["audit_status"]; ok {
		e.AuditStatus = v.(int)
	}
	if v, ok := fields["audit_reason"]; ok {
		e.AuditReason = v.(string)
	}
	if v, ok := fields["title"]; ok {
		e.Title = v.(string)
	}
	if v, ok := fields["published_at"]; ok {
		if t, ok := v.(*time.Time); ok {
			e.PublishedAt = t
		}
	}
	return nil
}

func (m *mockErshouRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *mockErshouRepo) List(regionID uint, pagination *utils.Pagination, opts repository.ListOptions) ([]model.Ershou, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Ershou
	for _, e := range m.byID {
		if regionID > 0 && e.RegionID != regionID {
			continue
		}
		if opts.Status == -1 {
			// 全部
		} else if opts.Status > 0 {
			if e.Status != opts.Status {
				continue
			}
		} else {
			if e.Status != model.StatusPublished {
				continue
			}
		}
		list = append(list, *e)
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) AdminList(pagination *utils.Pagination, opts repository.AdminListOptions) ([]model.Ershou, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	var list []model.Ershou
	for _, e := range m.byID {
		if opts.RegionID > 0 && e.RegionID != opts.RegionID {
			continue
		}
		if opts.UserID > 0 && e.UserID != opts.UserID {
			continue
		}
		if opts.Status != nil && e.Status != *opts.Status {
			continue
		}
		if opts.AuditStatus != nil && e.AuditStatus != *opts.AuditStatus {
			continue
		}
		list = append(list, *e)
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts repository.ListOptions) ([]model.Ershou, int64, error) {
	if m.nearbyErr != nil {
		return nil, 0, m.nearbyErr
	}
	var list []model.Ershou
	for _, e := range m.byID {
		if e.Status != model.StatusPublished {
			continue
		}
		if regionID > 0 && e.RegionID != regionID {
			continue
		}
		list = append(list, *e)
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Ershou, int64, error) {
	if m.searchErr != nil {
		return nil, 0, m.searchErr
	}
	var list []model.Ershou
	for _, e := range m.byID {
		if e.Status != model.StatusPublished {
			continue
		}
		if regionID > 0 && e.RegionID != regionID {
			continue
		}
		list = append(list, *e)
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) IncrViewCount(id uint) error {
	m.incrViewCountCalls = append(m.incrViewCountCalls, id)
	if m.incrViewErr != nil {
		return m.incrViewErr
	}
	return nil
}

func (m *mockErshouRepo) ListImages(ershouID uint) ([]model.ErshouImage, error) {
	if m.listImagesErr != nil {
		return nil, m.listImagesErr
	}
	return m.images[ershouID], nil
}

func (m *mockErshouRepo) ReplaceImages(ershouID uint, urls []string) error {
	m.replaceImagesCalls = append(m.replaceImagesCalls, struct {
		ershouID uint
		urls     []string
	}{ershouID, urls})
	if m.replaceImagesErr != nil {
		return m.replaceImagesErr
	}
	imgs := make([]model.ErshouImage, 0, len(urls))
	for i, u := range urls {
		imgs = append(imgs, model.ErshouImage{ErshouID: ershouID, URL: u, Sort: i})
	}
	m.images[ershouID] = imgs
	return nil
}

func (m *mockErshouRepo) DeleteImages(ershouID uint) error {
	delete(m.images, ershouID)
	return nil
}

func (m *mockErshouRepo) FavExists(userID, ershouID uint) (bool, error) {
	if m.favExistsErr != nil {
		return false, m.favExistsErr
	}
	return m.favs[favKey{userID, ershouID}], nil
}

func (m *mockErshouRepo) CreateFav(fav *model.ErshouFavorite) error {
	if m.createFavErr != nil {
		return m.createFavErr
	}
	m.favs[favKey{fav.UserID, fav.ErshouID}] = true
	return nil
}

func (m *mockErshouRepo) DeleteFav(userID, ershouID uint) error {
	if m.deleteFavErr != nil {
		return m.deleteFavErr
	}
	delete(m.favs, favKey{userID, ershouID})
	return nil
}

func (m *mockErshouRepo) IncrFavCount(id uint) error {
	m.incrFavCountCalls = append(m.incrFavCountCalls, id)
	if m.incrFavErr != nil {
		return m.incrFavErr
	}
	if e, ok := m.byID[id]; ok {
		e.FavCount++
	}
	return nil
}

func (m *mockErshouRepo) DecrFavCount(id uint) error {
	m.decrFavCountCalls = append(m.decrFavCountCalls, id)
	if m.decrFavErr != nil {
		return m.decrFavErr
	}
	if e, ok := m.byID[id]; ok && e.FavCount > 0 {
		e.FavCount--
	}
	return nil
}

func (m *mockErshouRepo) ListFavs(userID uint, page, pageSize int) ([]model.ErshouFavorite, int64, error) {
	if m.listFavsErr != nil {
		return nil, 0, m.listFavsErr
	}
	var list []model.ErshouFavorite
	for k := range m.favs {
		if k.userID == userID {
			list = append(list, model.ErshouFavorite{UserID: k.userID, ErshouID: k.ershouID})
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	for _, id := range ids {
		result[id] = m.favs[favKey{userID, id}]
	}
	return result, nil
}

func (m *mockErshouRepo) CreateMessage(msg *model.ErshouMessage) error {
	if m.createMsgErr != nil {
		return m.createMsgErr
	}
	msg.ID = uint(len(m.messages) + 1)
	m.messages = append(m.messages, *msg)
	return nil
}

func (m *mockErshouRepo) ListMessages(ershouID uint, page, pageSize int) ([]model.ErshouMessage, int64, error) {
	if m.listMsgErr != nil {
		return nil, 0, m.listMsgErr
	}
	var list []model.ErshouMessage
	for _, msg := range m.messages {
		if msg.ErshouID == ershouID && msg.Status == 1 {
			list = append(list, msg)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockErshouRepo) IncrMessageCount(id uint) error {
	m.incrMsgCountCalls = append(m.incrMsgCountCalls, id)
	if m.incrMsgErr != nil {
		return m.incrMsgErr
	}
	if e, ok := m.byID[id]; ok {
		e.MessageCount++
	}
	return nil
}

func (m *mockErshouRepo) MarkMessagesRead(ershouID uint, userID uint) error {
	m.markMsgReadCalls = append(m.markMsgReadCalls, struct {
		ershouID uint
		userID   uint
	}{ershouID, userID})
	if m.markMsgReadErr != nil {
		return m.markMsgReadErr
	}
	return nil
}

func (m *mockErshouRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Ershou, int64, error) {
	if m.listByUserErr != nil {
		return nil, 0, m.listByUserErr
	}
	var list []model.Ershou
	for _, e := range m.byID {
		if e.UserID == userID {
			list = append(list, *e)
		}
	}
	return list, int64(len(list)), nil
}

// ===== 测试辅助 =====

func newTestErshouService() (*ershouService, *mockErshouRepo) {
	repo := newMockErshouRepo()
	svc := NewErshouService(repo).(*ershouService)
	return svc, repo
}

func newE(id uint, e model.Ershou) *model.Ershou { e.ID = id; return &e }

// ===== toErshouInfo 默认值测试 =====

func TestToErshouInfo(t *testing.T) {
	t.Run("默认值兜底", func(t *testing.T) {
		e := newE(1, model.Ershou{Title: "手机"})
		info := toErshouInfo(e, nil)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, "元", info.PriceUnit)
		assert.Equal(t, model.ConditionUsed, info.Condition)
		assert.Equal(t, model.DeliveryFace, info.DeliveryMethod)
		assert.Equal(t, []string{}, info.Images)
	})
	t.Run("显式值保留", func(t *testing.T) {
		e := newE(1, model.Ershou{PriceUnit: "面议", Condition: model.ConditionNew, DeliveryMethod: model.DeliveryExpress})
		info := toErshouInfo(e, []string{"a.jpg", "b.jpg"})
		assert.Equal(t, "面议", info.PriceUnit)
		assert.Equal(t, model.ConditionNew, info.Condition)
		assert.Equal(t, model.DeliveryExpress, info.DeliveryMethod)
		assert.Equal(t, []string{"a.jpg", "b.jpg"}, info.Images)
	})
}

// ===== Create 发布测试 =====

func TestErshouCreate(t *testing.T) {
	t.Run("正常发布含默认值", func(t *testing.T) {
		svc, repo := newTestErshouService()
		req := &dto.CreateErshouRequest{
			Title:    "iPhone 13",
			Price:    2999,
			Status:   model.StatusPublished,
			Images:   []string{"1.jpg", "2.jpg"},
		}
		info, err := svc.Create(2, 10, "张三", "13800000000", "avatar.jpg", req)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, "iPhone 13", info.Title)
		assert.Equal(t, "元", info.PriceUnit)            // 默认
		assert.Equal(t, model.ConditionUsed, info.Condition)        // 默认
		assert.Equal(t, model.DeliveryFace, info.DeliveryMethod)    // 默认
		assert.Equal(t, model.AuditApproved, info.AuditStatus)      // MVP 发布即通过
		assert.Equal(t, model.StatusPublished, info.Status)
		assert.Equal(t, uint(10), info.UserID)
		assert.Equal(t, "张三", info.UserName)
		assert.Equal(t, uint(2), info.RegionID)
		assert.Equal(t, []string{"1.jpg", "2.jpg"}, info.Images)
		require.NotNil(t, info.ExpiryTime) // 默认 30 天
		assert.NotNil(t, info.PublishedAt) // 发布状态设发布时间
		// 仓储写入
		e := repo.byID[1]
		assert.Equal(t, uint(2), e.RegionID)
		require.NotNil(t, e.ExpiryTime)
		// 图片子表保存
		require.Len(t, repo.replaceImagesCalls, 1)
		assert.Equal(t, uint(1), repo.replaceImagesCalls[0].ershouID)
	})

	t.Run("草稿不设发布时间", func(t *testing.T) {
		svc, repo := newTestErshouService()
		req := &dto.CreateErshouRequest{Title: "草稿", Price: 100, Status: model.StatusDraft}
		info, err := svc.Create(1, 1, "u", "p", "a", req)
		require.NoError(t, err)
		assert.Equal(t, model.StatusDraft, info.Status)
		assert.Nil(t, info.PublishedAt)
		assert.Nil(t, repo.byID[info.ID].PublishedAt)
	})

	t.Run("显式 expireDays 生效", func(t *testing.T) {
		svc, _ := newTestErshouService()
		req := &dto.CreateErshouRequest{Title: "x", Price: 1, ExpireDays: 7}
		info, err := svc.Create(1, 1, "u", "p", "a", req)
		require.NoError(t, err)
		require.NotNil(t, info.ExpiryTime)
		diff := info.ExpiryTime.Sub(time.Now())
		assert.InDelta(t, 7*24*time.Hour, float64(diff), float64(time.Hour))
	})

	t.Run("无图片不调用 ReplaceImages", func(t *testing.T) {
		svc, repo := newTestErshouService()
		_, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateErshouRequest{Title: "x", Price: 1})
		require.NoError(t, err)
		assert.Empty(t, repo.replaceImagesCalls)
	})

	t.Run("仓储创建失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.createErr = errors.New("db error")
		_, err := svc.Create(1, 1, "u", "p", "a", &dto.CreateErshouRequest{Title: "x", Price: 1})
		assert.Error(t, err)
	})
}

// ===== Update 更新测试 =====

func TestErshouUpdate(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		err := svc.Update(999, 10, &dto.UpdateErshouRequest{Title: "x"})
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("无权操作他人物品", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 99})
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Title: "x"})
		assert.ErrorIs(t, err, ErrErshouNoPermission)
	})

	t.Run("FindByID 返回其他错误透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.findErr = errors.New("conn lost")
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Title: "x"})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("普通字段更新", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10, Title: "旧"})
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Title: "新", Price: 200, Brand: "Apple"})
		require.NoError(t, err)
		require.Len(t, repo.updateFieldsCalls, 1)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, "新", fields["title"])
		assert.Equal(t, float64(200), fields["price"])
		assert.Equal(t, "Apple", fields["brand"])
		// 仓储对象同步更新
		assert.Equal(t, "新", repo.byID[1].Title)
	})

	t.Run("状态由草稿改为发布-设发布时间并重置审核", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10, Status: model.StatusDraft, AuditStatus: model.AuditRejected})
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Status: model.StatusPublished})
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.StatusPublished, fields["status"])
		assert.Equal(t, model.AuditApproved, fields["audit_status"])
		_, hasPub := fields["published_at"]
		assert.True(t, hasPub)
		assert.Equal(t, model.StatusPublished, repo.byID[1].Status)
	})

	t.Run("改为下架状态", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10, Status: model.StatusPublished})
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Status: model.StatusOffline})
		require.NoError(t, err)
		assert.Equal(t, model.StatusOffline, repo.updateFieldsCalls[0].fields["status"])
		assert.Equal(t, model.StatusOffline, repo.byID[1].Status)
	})

	t.Run("更新图片子表", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Images: []string{"new.jpg"}})
		require.NoError(t, err)
		require.Len(t, repo.replaceImagesCalls, 1)
		assert.Equal(t, []string{"new.jpg"}, repo.replaceImagesCalls[0].urls)
	})

	t.Run("UpdateFields 失败透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		repo.updateFieldsErr = errors.New("db error")
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Title: "x"})
		assert.Error(t, err)
	})

	t.Run("ReplaceImages 失败透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		repo.replaceImagesErr = errors.New("img error")
		err := svc.Update(1, 10, &dto.UpdateErshouRequest{Images: []string{"x.jpg"}})
		assert.Error(t, err)
	})
}

// ===== Delete 删除测试 =====

func TestErshouDelete(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		err := svc.Delete(999, 10)
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("无权删除他人物品", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 99})
		err := svc.Delete(1, 10)
		assert.ErrorIs(t, err, ErrErshouNoPermission)
	})

	t.Run("正常删除", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		err := svc.Delete(1, 10)
		require.NoError(t, err)
		_, ok := repo.byID[1]
		assert.False(t, ok)
	})

	t.Run("仓储删除失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		repo.deleteErr = errors.New("db error")
		err := svc.Delete(1, 10)
		assert.Error(t, err)
	})
}

// ===== GetByID 详情测试 =====

func TestErshouGetByID(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.GetByID(999, 10)
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("正常获取-增加浏览量+图片拼装+收藏态", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10, Title: "手机", ViewCount: 5})
		repo.images[1] = []model.ErshouImage{{URL: "a.jpg"}, {URL: "b.jpg"}}
		repo.favs[favKey{10, 1}] = true
		info, err := svc.GetByID(1, 10)
		require.NoError(t, err)
		assert.Equal(t, "手机", info.Title)
		assert.Equal(t, 6, info.ViewCount) // service 本地 +1
		assert.Equal(t, []string{"a.jpg", "b.jpg"}, info.Images)
		assert.True(t, info.HasFaved)
		require.Len(t, repo.incrViewCountCalls, 1)
		assert.Equal(t, uint(1), repo.incrViewCountCalls[0])
	})

	t.Run("发布者查看-标记留言已读", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		_, err := svc.GetByID(1, 10)
		require.NoError(t, err)
		require.Len(t, repo.markMsgReadCalls, 1)
		assert.Equal(t, uint(1), repo.markMsgReadCalls[0].ershouID)
		assert.Equal(t, uint(10), repo.markMsgReadCalls[0].userID)
	})

	t.Run("非发布者查看-不标记留言已读", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		_, err := svc.GetByID(1, 20)
		require.NoError(t, err)
		assert.Empty(t, repo.markMsgReadCalls)
	})

	t.Run("未登录用户-无收藏态不标记已读", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 10})
		info, err := svc.GetByID(1, 0)
		require.NoError(t, err)
		assert.False(t, info.HasFaved)
		assert.Empty(t, repo.markMsgReadCalls)
	})

	t.Run("FindByID 其他错误透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.findErr = errors.New("conn lost")
		_, err := svc.GetByID(1, 10)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrErshouNotFound)
	})
}

// ===== List 列表测试 =====

func TestErshouList(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.byID[1] = newE(1, model.Ershou{Title: "A", Status: model.StatusPublished})
	repo.byID[1].RegionID = 1
	repo.byID[2] = newE(2, model.Ershou{Title: "B", Status: model.StatusDraft})
	repo.byID[2].RegionID = 1
	repo.byID[3] = newE(3, model.Ershou{Title: "C", Status: model.StatusPublished})
	repo.byID[3].RegionID = 2

	pagination, list, err := svc.List(1, &dto.ErshouListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "A", list[0].Title)
}

func TestErshouListError(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.listErr = errors.New("db error")
	_, _, err := svc.List(1, &dto.ErshouListRequest{})
	assert.Error(t, err)
}

// ===== ListNearby 附近测试 =====

func TestErshouListNearby(t *testing.T) {
	t.Run("默认半径5km", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Title: "A", Status: model.StatusPublished})
		repo.byID[1].RegionID = 1
		pagination, list, err := svc.ListNearby(1, &dto.ErshouNearbyRequest{Latitude: 30.0, Longitude: 114.0})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.nearbyErr = errors.New("postgis error")
		_, _, err := svc.ListNearby(1, &dto.ErshouNearbyRequest{Latitude: 30.0, Longitude: 114.0})
		assert.Error(t, err)
	})
}

// ===== Search 搜索测试 =====

func TestErshouSearch(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.byID[1] = newE(1, model.Ershou{Title: "iPhone", Status: model.StatusPublished})
	repo.byID[1].RegionID = 1
	repo.byID[2] = newE(2, model.Ershou{Title: "华为", Status: model.StatusPublished})
	repo.byID[2].RegionID = 1
	pagination, list, err := svc.Search(1, &dto.ErshouSearchRequest{Keyword: "手机"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), pagination.Total)
	assert.Len(t, list, 2)
}

func TestErshouSearchError(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.searchErr = errors.New("es error")
	_, _, err := svc.Search(1, &dto.ErshouSearchRequest{Keyword: "x"})
	assert.Error(t, err)
}

// ===== ListMine 我的发布测试 =====

func TestErshouListMine(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.byID[1] = newE(1, model.Ershou{UserID: 10, Title: "A"})
	repo.byID[2] = newE(2, model.Ershou{UserID: 20, Title: "B"})
	pagination, list, err := svc.ListMine(10, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "A", list[0].Title)
}

func TestErshouListMineError(t *testing.T) {
	svc, repo := newTestErshouService()
	repo.listByUserErr = errors.New("db error")
	_, _, err := svc.ListMine(10, 1, 10)
	assert.Error(t, err)
}

// ===== 收藏测试 =====

func TestErshouFav(t *testing.T) {
	t.Run("未登录用户禁止收藏", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.Fav(0, 1)
		assert.ErrorIs(t, err, ErrErshouNoPermission)
	})

	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.Fav(10, 999)
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("未收藏-创建收藏", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20, FavCount: 5})
		resp, err := svc.Fav(10, 1)
		require.NoError(t, err)
		assert.True(t, resp.HasFaved)
		assert.Equal(t, 6, resp.FavCount)
		assert.True(t, repo.favs[favKey{10, 1}])
		require.Len(t, repo.incrFavCountCalls, 1)
		assert.Equal(t, uint(1), repo.incrFavCountCalls[0])
	})

	t.Run("已收藏-取消收藏", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20, FavCount: 5})
		repo.favs[favKey{10, 1}] = true
		resp, err := svc.Fav(10, 1)
		require.NoError(t, err)
		assert.False(t, resp.HasFaved)
		assert.Equal(t, 4, resp.FavCount)
		assert.False(t, repo.favs[favKey{10, 1}])
		require.Len(t, repo.decrFavCountCalls, 1)
	})

	t.Run("FavExists 失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20})
		repo.favExistsErr = errors.New("db error")
		_, err := svc.Fav(10, 1)
		assert.Error(t, err)
	})

	t.Run("CreateFav 失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20})
		repo.createFavErr = errors.New("db error")
		_, err := svc.Fav(10, 1)
		assert.Error(t, err)
	})

	t.Run("DeleteFav 失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20})
		repo.favs[favKey{10, 1}] = true
		repo.deleteFavErr = errors.New("db error")
		_, err := svc.Fav(10, 1)
		assert.Error(t, err)
	})
}

func TestErshouFavStatus(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.FavStatus(10, 999)
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("未登录用户-返回未收藏", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{FavCount: 3})
		resp, err := svc.FavStatus(0, 1)
		require.NoError(t, err)
		assert.False(t, resp.HasFaved)
		assert.Equal(t, 3, resp.FavCount)
	})

	t.Run("登录用户-已收藏", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{FavCount: 3})
		repo.favs[favKey{10, 1}] = true
		resp, err := svc.FavStatus(10, 1)
		require.NoError(t, err)
		assert.True(t, resp.HasFaved)
		assert.Equal(t, 3, resp.FavCount)
	})
}

func TestErshouListFavs(t *testing.T) {
	t.Run("无收藏返回空列表", func(t *testing.T) {
		svc, _ := newTestErshouService()
		pagination, list, err := svc.ListFavs(10, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pagination.Total)
		assert.Len(t, list, 0)
	})

	t.Run("正常返回收藏列表", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Title: "A"})
		repo.byID[2] = newE(2, model.Ershou{Title: "B"})
		repo.favs[favKey{10, 1}] = true
		repo.favs[favKey{10, 2}] = true
		repo.favs[favKey{20, 1}] = true
		pagination, list, err := svc.ListFavs(10, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), pagination.Total)
		assert.Len(t, list, 2)
	})

	t.Run("ListFavs 仓储失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.listFavsErr = errors.New("db error")
		_, _, err := svc.ListFavs(10, 1, 10)
		assert.Error(t, err)
	})
}

// ===== 留言测试 =====

func TestErshouCreateMessage(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.CreateMessage(999, 10, "张三", "avatar", &dto.CreateMessageRequest{Content: "你好"})
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("正常留言-计数+1", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20, MessageCount: 2})
		info, err := svc.CreateMessage(1, 10, "张三", "avatar", &dto.CreateMessageRequest{Content: "想要"})
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, uint(1), info.ErshouID)
		assert.Equal(t, uint(10), info.FromUserID)
		assert.Equal(t, "张三", info.FromName)
		assert.Equal(t, "想要", info.Content)
		// model 层 Status=1（正常），DTO 不暴露 Status，校验仓储写入
		require.Len(t, repo.messages, 1)
		assert.Equal(t, 1, repo.messages[0].Status)
		require.Len(t, repo.incrMsgCountCalls, 1)
		assert.Equal(t, uint(1), repo.incrMsgCountCalls[0])
	})

	t.Run("CreateMessage 仓储失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{UserID: 20})
		repo.createMsgErr = errors.New("db error")
		_, err := svc.CreateMessage(1, 10, "张三", "avatar", &dto.CreateMessageRequest{Content: "x"})
		assert.Error(t, err)
	})
}

func TestErshouListMessages(t *testing.T) {
	t.Run("正常列表", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.messages = []model.ErshouMessage{
			{ID: 1, ErshouID: 1, Content: "A", Status: 1},
			{ID: 2, ErshouID: 1, Content: "B", Status: 1},
			{ID: 3, ErshouID: 2, Content: "C", Status: 1},
			{ID: 4, ErshouID: 1, Content: "D", Status: 0}, // 已删除
		}
		list, total, err := svc.ListMessages(1, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, list, 2)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.listMsgErr = errors.New("db error")
		_, _, err := svc.ListMessages(1, 1, 10)
		assert.Error(t, err)
	})
}

// ===== 管理后台测试 =====

func TestErshouAdminList(t *testing.T) {
	t.Run("正常列表-按状态过滤", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Title: "A", Status: model.StatusDraft, AuditStatus: model.AuditPending})
		repo.byID[2] = newE(2, model.Ershou{Title: "B", Status: model.StatusPublished, AuditStatus: model.AuditApproved})
		draftStatus := model.StatusDraft
		pagination, list, err := svc.AdminList(&dto.ErshouAdminListRequest{Status: &draftStatus})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
		assert.Equal(t, "A", list[0].Title)
	})

	t.Run("按审核状态过滤", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Title: "A", AuditStatus: model.AuditPending})
		repo.byID[2] = newE(2, model.Ershou{Title: "B", AuditStatus: model.AuditApproved})
		approved := model.AuditApproved
		_, list, err := svc.AdminList(&dto.ErshouAdminListRequest{AuditStatus: &approved})
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, "B", list[0].Title)
	})

	t.Run("仓储失败", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.adminListErr = errors.New("db error")
		_, _, err := svc.AdminList(&dto.ErshouAdminListRequest{})
		assert.Error(t, err)
	})
}

func TestErshouAdminGetByID(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		_, err := svc.AdminGetByID(999)
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("正常获取含图片", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Title: "手机"})
		repo.images[1] = []model.ErshouImage{{URL: "a.jpg"}, {URL: "b.jpg"}}
		info, err := svc.AdminGetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "手机", info.Title)
		assert.Equal(t, []string{"a.jpg", "b.jpg"}, info.Images)
	})
}

func TestErshouAudit(t *testing.T) {
	t.Run("物品不存在", func(t *testing.T) {
		svc, _ := newTestErshouService()
		err := svc.Audit(999, model.AuditApproved, "")
		assert.ErrorIs(t, err, ErrErshouNotFound)
	})

	t.Run("审核通过-草稿自动发布", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusDraft, AuditStatus: model.AuditPending})
		err := svc.Audit(1, model.AuditApproved, "")
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.AuditApproved, fields["audit_status"])
		assert.Equal(t, model.StatusPublished, fields["status"])
		_, hasPub := fields["published_at"]
		assert.True(t, hasPub)
	})

	t.Run("审核通过-已发布不重复设发布时间", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusPublished, AuditStatus: model.AuditPending})
		err := svc.Audit(1, model.AuditApproved, "")
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.AuditApproved, fields["audit_status"])
		_, hasStatus := fields["status"]
		assert.False(t, hasStatus) // 已发布不再改 status
		_, hasPub := fields["published_at"]
		assert.False(t, hasPub)
	})

	t.Run("审核拒绝-已发布强制下架", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusPublished, AuditStatus: model.AuditPending})
		err := svc.Audit(1, model.AuditRejected, "违规")
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.AuditRejected, fields["audit_status"])
		assert.Equal(t, "违规", fields["audit_reason"])
		assert.Equal(t, model.StatusOffline, fields["status"])
	})

	t.Run("审核拒绝-草稿不下架", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusDraft, AuditStatus: model.AuditPending})
		err := svc.Audit(1, model.AuditRejected, "不合规")
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.AuditRejected, fields["audit_status"])
		_, hasStatus := fields["status"]
		assert.False(t, hasStatus) // 草稿不触发下架
	})

	t.Run("UpdateFields 失败透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusPublished})
		repo.updateFieldsErr = errors.New("db error")
		err := svc.Audit(1, model.AuditApproved, "")
		assert.Error(t, err)
	})
}

func TestErshouAdminUpdateStatus(t *testing.T) {
	t.Run("恢复发布-设发布时间并重置审核", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusOffline})
		err := svc.AdminUpdateStatus(1, model.StatusPublished)
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.StatusPublished, fields["status"])
		assert.Equal(t, model.AuditApproved, fields["audit_status"])
		_, hasPub := fields["published_at"]
		assert.True(t, hasPub)
	})

	t.Run("下架-不设发布时间", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{Status: model.StatusPublished})
		err := svc.AdminUpdateStatus(1, model.StatusOffline)
		require.NoError(t, err)
		fields := repo.updateFieldsCalls[0].fields
		assert.Equal(t, model.StatusOffline, fields["status"])
		_, hasPub := fields["published_at"]
		assert.False(t, hasPub)
		_, hasAudit := fields["audit_status"]
		assert.False(t, hasAudit)
	})

	t.Run("UpdateFields 失败透传", func(t *testing.T) {
		svc, repo := newTestErshouService()
		repo.byID[1] = newE(1, model.Ershou{})
		repo.updateFieldsErr = errors.New("db error")
		err := svc.AdminUpdateStatus(1, model.StatusOffline)
		assert.Error(t, err)
	})
}
