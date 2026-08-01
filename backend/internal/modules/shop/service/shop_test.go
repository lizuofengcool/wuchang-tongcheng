// Package service 商家模块业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖：店铺/相册/评价三仓储 CRUD、DTO 转换、
// 商家入驻申请（重复校验）、编辑店铺（权限+字段过滤）、图片删除（归属校验）、
// 评价发表（审核状态校验）、店铺审核（状态联动）、评价审核（评分重算）等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/shop/dto"
	"wuchang-tongcheng/internal/modules/shop/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockShopRepo =====

type mockShopRepo struct {
	byID          map[uint]*model.Shop
	byUser        map[uint]*model.Shop // key = userID*1000 + regionID
	nextID        uint
	createErr     error
	findErr       error
	findByUserErr error
	updateErr     error
	deleteErr     error
	listErr       error
	adminListErr  error
	incrViewsErr  error
	updateRating  map[uint]float32
	updateRatErr  error
	updatedFields map[uint]map[string]interface{}
	deletedIDs    map[uint]bool
	incrViewsCall map[uint]int
}

func newMockShopRepo() *mockShopRepo {
	return &mockShopRepo{
		byID:          make(map[uint]*model.Shop),
		byUser:        make(map[uint]*model.Shop),
		nextID:        1,
		updateRating:  make(map[uint]float32),
		updatedFields: make(map[uint]map[string]interface{}),
		deletedIDs:    make(map[uint]bool),
		incrViewsCall: make(map[uint]int),
	}
}

func userKey(userID, regionID uint) uint { return userID*1000 + regionID }

func (m *mockShopRepo) Create(shop *model.Shop) error {
	if m.createErr != nil {
		return m.createErr
	}
	shop.ID = m.nextID
	m.nextID++
	cp := *shop
	m.byID[shop.ID] = &cp
	if shop.UserID > 0 {
		m.byUser[userKey(shop.UserID, shop.RegionID)] = &cp
	}
	return nil
}

func (m *mockShopRepo) FindByID(id uint) (*model.Shop, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	shop, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *shop
	return &cp, nil
}

func (m *mockShopRepo) FindByUserID(userID uint, regionID uint) (*model.Shop, error) {
	if m.findByUserErr != nil {
		return nil, m.findByUserErr
	}
	shop, ok := m.byUser[userKey(userID, regionID)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *shop
	return &cp, nil
}

func (m *mockShopRepo) Update(shop *model.Shop) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *shop
	m.byID[shop.ID] = &cp
	return nil
}

func (m *mockShopRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedFields[id] = fields
	shop, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["audit_status"]; ok {
		shop.AuditStatus = v.(int)
	}
	if v, ok := fields["status"]; ok {
		shop.Status = v.(int)
	}
	if v, ok := fields["is_recommend"]; ok {
		shop.IsRecommend = v.(int)
	}
	if v, ok := fields["name"]; ok {
		shop.Name = v.(string)
	}
	return nil
}

func (m *mockShopRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	m.deletedIDs[id] = true
	return nil
}

func (m *mockShopRepo) List(regionID uint, pagination *utils.Pagination, categoryID uint, isRecommend int, keyword string) ([]model.Shop, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Shop
	for _, s := range m.byID {
		if s.AuditStatus != model.AuditStatusApproved {
			continue
		}
		if regionID > 0 && s.RegionID != regionID {
			continue
		}
		if categoryID > 0 && s.CategoryID != categoryID {
			continue
		}
		if (isRecommend == 0 || isRecommend == 1) && s.IsRecommend != isRecommend {
			continue
		}
		if keyword != "" && !contains(s.Name, keyword) {
			continue
		}
		list = append(list, *s)
	}
	return list, int64(len(list)), nil
}

func (m *mockShopRepo) AdminList(regionID uint, pagination *utils.Pagination, categoryID uint, auditStatus int, status int, isRecommend int, keyword string) ([]model.Shop, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	var list []model.Shop
	for _, s := range m.byID {
		if regionID > 0 && s.RegionID != regionID {
			continue
		}
		if categoryID > 0 && s.CategoryID != categoryID {
			continue
		}
		if auditStatus >= 0 && auditStatus <= 2 && s.AuditStatus != auditStatus {
			continue
		}
		if status >= 0 && status <= 2 && s.Status != status {
			continue
		}
		if (isRecommend == 0 || isRecommend == 1) && s.IsRecommend != isRecommend {
			continue
		}
		if keyword != "" && !contains(s.Name, keyword) {
			continue
		}
		list = append(list, *s)
	}
	return list, int64(len(list)), nil
}

func (m *mockShopRepo) IncrViews(id uint) error {
	if m.incrViewsErr != nil {
		return m.incrViewsErr
	}
	m.incrViewsCall[id]++
	return nil
}

func (m *mockShopRepo) UpdateRating(id uint, rating float32) error {
	if m.updateRatErr != nil {
		return m.updateRatErr
	}
	m.updateRating[id] = rating
	if s, ok := m.byID[id]; ok {
		s.Rating = rating
	}
	return nil
}

// ===== mockShopImageRepo =====

type mockShopImageRepo struct {
	byID      map[uint]*model.ShopImage
	nextID    uint
	createErr error
	findErr   error
	listErr   error
	deleteErr error
	deletedIDs map[uint]bool
}

func newMockShopImageRepo() *mockShopImageRepo {
	return &mockShopImageRepo{
		byID:       make(map[uint]*model.ShopImage),
		nextID:     1,
		deletedIDs: make(map[uint]bool),
	}
}

func (m *mockShopImageRepo) Create(img *model.ShopImage) error {
	if m.createErr != nil {
		return m.createErr
	}
	img.ID = m.nextID
	m.nextID++
	cp := *img
	m.byID[img.ID] = &cp
	return nil
}

func (m *mockShopImageRepo) FindByID(id uint) (*model.ShopImage, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	img, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *img
	return &cp, nil
}

func (m *mockShopImageRepo) FindByShopID(shopID uint) ([]model.ShopImage, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var list []model.ShopImage
	for _, img := range m.byID {
		if img.ShopID == shopID {
			list = append(list, *img)
		}
	}
	return list, nil
}

func (m *mockShopImageRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	m.deletedIDs[id] = true
	return nil
}

// ===== mockShopReviewRepo =====

type mockShopReviewRepo struct {
	byID        map[uint]*model.ShopReview
	nextID      uint
	createErr   error
	findErr     error
	approvedErr error
	adminListErr error
	updateErr   error
	avgErr      error
	avgResult   map[uint]struct {
		avg   float32
		count int64
	}
	updatedFields map[uint]map[string]interface{}
}

func newMockShopReviewRepo() *mockShopReviewRepo {
	return &mockShopReviewRepo{
		byID:          make(map[uint]*model.ShopReview),
		nextID:        1,
		avgResult:     make(map[uint]struct {
			avg   float32
			count int64
		}),
		updatedFields: make(map[uint]map[string]interface{}),
	}
}

func (m *mockShopReviewRepo) Create(review *model.ShopReview) error {
	if m.createErr != nil {
		return m.createErr
	}
	review.ID = m.nextID
	m.nextID++
	cp := *review
	m.byID[review.ID] = &cp
	return nil
}

func (m *mockShopReviewRepo) FindByID(id uint) (*model.ShopReview, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	r, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r
	return &cp, nil
}

func (m *mockShopReviewRepo) FindApprovedByShopID(shopID uint, pagination *utils.Pagination) ([]model.ShopReview, int64, error) {
	if m.approvedErr != nil {
		return nil, 0, m.approvedErr
	}
	var list []model.ShopReview
	for _, r := range m.byID {
		if r.ShopID == shopID && r.Status == model.ReviewStatusApproved {
			list = append(list, *r)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockShopReviewRepo) AdminList(pagination *utils.Pagination, shopID uint, status int) ([]model.ShopReview, int64, error) {
	if m.adminListErr != nil {
		return nil, 0, m.adminListErr
	}
	var list []model.ShopReview
	for _, r := range m.byID {
		if shopID > 0 && r.ShopID != shopID {
			continue
		}
		if status >= 0 && status <= 2 && r.Status != status {
			continue
		}
		list = append(list, *r)
	}
	return list, int64(len(list)), nil
}

func (m *mockShopReviewRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updatedFields[id] = fields
	if r, ok := m.byID[id]; ok {
		if v, ok := fields["status"]; ok {
			r.Status = v.(int)
		}
	}
	return nil
}

func (m *mockShopReviewRepo) AvgRatingByShopID(shopID uint) (float32, int64, error) {
	if m.avgErr != nil {
		return 0, 0, m.avgErr
	}
	if res, ok := m.avgResult[shopID]; ok {
		return res.avg, res.count, nil
	}
	// 默认按已通过评价计算
	var sum float32
	var count int64
	for _, r := range m.byID {
		if r.ShopID == shopID && r.Status == model.ReviewStatusApproved {
			sum += float32(r.Rating)
			count++
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return sum / float32(count), count, nil
}

// ===== 测试辅助 =====

func newTestService() (*shopService, *mockShopRepo, *mockShopImageRepo, *mockShopReviewRepo) {
	shopRepo := newMockShopRepo()
	imageRepo := newMockShopImageRepo()
	reviewRepo := newMockShopReviewRepo()
	svc := NewShopService(shopRepo, imageRepo, reviewRepo).(*shopService)
	return svc, shopRepo, imageRepo, reviewRepo
}

func newShop(id uint, s model.Shop) *model.Shop {
	s.ID = id
	return &s
}

func newImage(id uint, img model.ShopImage) *model.ShopImage {
	img.ID = id
	return &img
}

func newReview(id uint, r model.ShopReview) *model.ShopReview {
	r.ID = id
	return &r
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ===== 纯函数测试 =====

func TestToShopInfo(t *testing.T) {
	t.Run("完整字段转换", func(t *testing.T) {
		s := &model.Shop{
			Name:          "五常大米店",
			Logo:          "https://example.com/logo.png",
			Description:   "专营大米",
			Phone:         "13800000000",
			Address:       "五常市",
			Longitude:     127.15,
			Latitude:      44.93,
			CategoryID:    5,
			BusinessHours: "08:00-22:00",
			Status:        model.ShopStatusOpen,
			AuditStatus:   model.AuditStatusApproved,
			Rating:        4.5,
			Views:         100,
			IsRecommend:   1,
			Sort:          10,
			UserID:        8,
		}
		s.ID = 1
		s.RegionID = 2
		info := toShopInfo(s)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, "五常大米店", info.Name)
		assert.Equal(t, "https://example.com/logo.png", info.Logo)
		assert.Equal(t, "专营大米", info.Description)
		assert.Equal(t, "13800000000", info.Phone)
		assert.Equal(t, "五常市", info.Address)
		assert.Equal(t, 127.15, info.Longitude)
		assert.Equal(t, 44.93, info.Latitude)
		assert.Equal(t, uint(5), info.CategoryID)
		assert.Equal(t, "08:00-22:00", info.BusinessHours)
		assert.Equal(t, model.ShopStatusOpen, info.Status)
		assert.Equal(t, model.AuditStatusApproved, info.AuditStatus)
		assert.Equal(t, float32(4.5), info.Rating)
		assert.Equal(t, 100, info.Views)
		assert.Equal(t, 1, info.IsRecommend)
		assert.Equal(t, 10, info.Sort)
		assert.Equal(t, uint(8), info.UserID)
		assert.Equal(t, uint(2), info.RegionID)
	})

	t.Run("NilSafe", func(t *testing.T) {
		info := toShopInfo(&model.Shop{})
		assert.NotNil(t, info)
		assert.Equal(t, uint(0), info.ID)
		assert.Equal(t, "", info.Name)
	})
}

func TestToImageInfo(t *testing.T) {
	t.Run("完整字段转换", func(t *testing.T) {
		img := &model.ShopImage{ShopID: 3, ImageURL: "https://example.com/1.jpg", Sort: 2}
		img.ID = 9
		info := toImageInfo(img)
		assert.Equal(t, uint(9), info.ID)
		assert.Equal(t, uint(3), info.ShopID)
		assert.Equal(t, "https://example.com/1.jpg", info.ImageURL)
		assert.Equal(t, 2, info.Sort)
	})

	t.Run("NilSafe", func(t *testing.T) {
		info := toImageInfo(&model.ShopImage{})
		assert.NotNil(t, info)
		assert.Equal(t, uint(0), info.ShopID)
	})
}

func TestToReviewInfo(t *testing.T) {
	t.Run("完整字段转换含回复时间", func(t *testing.T) {
		rt, err := time.Parse(time.RFC3339, "2026-01-02T15:04:05Z")
		require.NoError(t, err)
		r := &model.ShopReview{
			ShopID:  3,
			UserID:  7,
			Rating:  5,
			Content: "好评",
			Reply:   "谢谢",
			ReplyAt: &rt,
			Status:  model.ReviewStatusApproved,
		}
		r.ID = 11
		info := toReviewInfo(r)
		assert.Equal(t, uint(11), info.ID)
		assert.Equal(t, uint(3), info.ShopID)
		assert.Equal(t, uint(7), info.UserID)
		assert.Equal(t, 5, info.Rating)
		assert.Equal(t, "好评", info.Content)
		assert.Equal(t, "谢谢", info.Reply)
		require.NotNil(t, info.ReplyAt)
		assert.Equal(t, rt, *info.ReplyAt)
		assert.Equal(t, model.ReviewStatusApproved, info.Status)
	})

	t.Run("NilSafe 回复时间为空", func(t *testing.T) {
		info := toReviewInfo(&model.ShopReview{})
		assert.NotNil(t, info)
		assert.Nil(t, info.ReplyAt)
	})
}

// ===== 构造函数测试 =====

func TestNewShopService(t *testing.T) {
	svc := NewShopService(newMockShopRepo(), newMockShopImageRepo(), newMockShopReviewRepo())
	_, ok := svc.(*shopService)
	assert.True(t, ok, "NewShopService 应返回 *shopService 类型")
}

// ===== 公开接口测试 =====

func TestGetByID(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		_, err := svc.GetByID(999)
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findErr = errors.New("db error")
		_, err := svc.GetByID(1)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("正常获取并自增浏览量", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "店铺A", Views: 10})
		info, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, "店铺A", info.Name)
		assert.Equal(t, 11, info.Views) // service 层 shop.Views++ 后返回
		assert.Equal(t, 1, shopRepo.incrViewsCall[1])
	})

	t.Run("IncrViews 失败不阻断返回", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "店铺A", Views: 5})
		shopRepo.incrViewsErr = errors.New("redis down")
		info, err := svc.GetByID(1)
		require.NoError(t, err) // IncrViews 错误被忽略
		assert.Equal(t, 6, info.Views)
	})
}

func TestList(t *testing.T) {
	t.Run("空列表", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		pagination, list, err := svc.List(1, &dto.ShopListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), pagination.Total)
		assert.Empty(t, list)
	})

	t.Run("多结果与地区隔离", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "A", AuditStatus: model.AuditStatusApproved})
		shopRepo.byID[1].RegionID = 1
		shopRepo.byID[2] = newShop(2, model.Shop{Name: "B", AuditStatus: model.AuditStatusApproved})
		shopRepo.byID[2].RegionID = 2
		shopRepo.byID[3] = newShop(3, model.Shop{Name: "C", AuditStatus: model.AuditStatusPending})
		shopRepo.byID[3].RegionID = 1
		pagination, list, err := svc.List(1, &dto.ShopListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
		assert.Equal(t, "A", list[0].Name)
	})

	t.Run("请求体 RegionID 覆盖入参", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "A", AuditStatus: model.AuditStatusApproved})
		shopRepo.byID[1].RegionID = 2
		// 入参 regionID=1，但 req.RegionID=2 应覆盖
		pagination, list, err := svc.List(1, &dto.ShopListRequest{RegionID: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
		assert.Equal(t, "A", list[0].Name)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.listErr = errors.New("db error")
		_, _, err := svc.List(1, &dto.ShopListRequest{})
		assert.Error(t, err)
	})

	t.Run("分页默认值兜底", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		pagination, _, err := svc.List(1, &dto.ShopListRequest{Page: 0, PageSize: 0})
		require.NoError(t, err)
		assert.Equal(t, 1, pagination.Page)
		assert.Equal(t, 10, pagination.PageSize)
	})
}

func TestGetImages(t *testing.T) {
	t.Run("空列表", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		list, err := svc.GetImages(1)
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("多结果", func(t *testing.T) {
		svc, _, imageRepo, _ := newTestService()
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5, ImageURL: "a.jpg", Sort: 1})
		imageRepo.byID[2] = newImage(2, model.ShopImage{ShopID: 5, ImageURL: "b.jpg", Sort: 2})
		imageRepo.byID[3] = newImage(3, model.ShopImage{ShopID: 9, ImageURL: "c.jpg", Sort: 1})
		list, err := svc.GetImages(5)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, _, imageRepo, _ := newTestService()
		imageRepo.listErr = errors.New("db error")
		_, err := svc.GetImages(1)
		assert.Error(t, err)
	})
}

func TestGetReviews(t *testing.T) {
	t.Run("空列表", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		pagination, list, err := svc.GetReviews(1, &dto.ReviewListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), pagination.Total)
		assert.Empty(t, list)
	})

	t.Run("仅返回已通过评价", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5, Status: model.ReviewStatusApproved, Rating: 5})
		reviewRepo.byID[2] = newReview(2, model.ShopReview{ShopID: 5, Status: model.ReviewStatusPending, Rating: 4})
		reviewRepo.byID[3] = newReview(3, model.ShopReview{ShopID: 5, Status: model.ReviewStatusApproved, Rating: 3})
		pagination, list, err := svc.GetReviews(5, &dto.ReviewListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), pagination.Total)
		require.Len(t, list, 2)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.approvedErr = errors.New("db error")
		_, _, err := svc.GetReviews(1, &dto.ReviewListRequest{})
		assert.Error(t, err)
	})
}

// ===== 用户接口测试 =====

func TestApply(t *testing.T) {
	t.Run("正常申请含默认值", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		req := &dto.ApplyShopRequest{
			Name:       "新店铺",
			Phone:      "13800000000",
			CategoryID: 3,
		}
		info, err := svc.Apply(2, 10, req)
		require.NoError(t, err)
		assert.Equal(t, "新店铺", info.Name)
		assert.Equal(t, model.ShopStatusClosed, info.Status)       // 默认歇业
		assert.Equal(t, model.AuditStatusPending, info.AuditStatus) // 默认待审核
		assert.Equal(t, uint(10), info.UserID)
		assert.Equal(t, uint(2), info.RegionID)
		// 验证仓储写入
		shop := shopRepo.byID[info.ID]
		require.NotNil(t, shop)
		assert.Equal(t, uint(2), shop.RegionID)
		assert.Equal(t, uint(10), shop.UserID)
	})

	t.Run("重复申请返回错误", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		_, err := svc.Apply(2, 10, &dto.ApplyShopRequest{Name: "x", Phone: "1"})
		assert.ErrorIs(t, err, ErrShopAlreadyExists)
	})

	t.Run("FindByUserID 非 NotFound 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findByUserErr = errors.New("db error")
		_, err := svc.Apply(2, 10, &dto.ApplyShopRequest{Name: "x", Phone: "1"})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopAlreadyExists)
	})

	t.Run("Create 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.createErr = errors.New("db error")
		_, err := svc.Apply(2, 10, &dto.ApplyShopRequest{Name: "x", Phone: "1"})
		assert.Error(t, err)
	})
}

func TestGetMyShop(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		_, err := svc.GetMyShop(10, 2)
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("正常获取", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "我的店", UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		info, err := svc.GetMyShop(10, 2)
		require.NoError(t, err)
		assert.Equal(t, "我的店", info.Name)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findByUserErr = errors.New("db error")
		_, err := svc.GetMyShop(10, 2)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})
}

func TestUpdateMyShop(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{Name: "x"})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	// 注：UpdateMyShop 内部先 FindByUserID(userID, regionID)，已按 userID 过滤，
	// 故 shop.UserID != userID 的 ErrShopNoPermission 分支为不可达的防御性代码，
	// 不构造对应用例。

	t.Run("全零值无字段更新直接返回", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		// longitude/latitude/category_id 总是写入，故 fields 非空，这里用全零值验证
		// 实际上 longitude/latitude/category_id 总会被写入 map，故不会进入 len==0 分支
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{})
		require.NoError(t, err)
	})

	t.Run("部分字段更新", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10, Name: "旧"})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{
			Name:     "新",
			Phone:    "13900000000",
			Status:   model.ShopStatusOpen,
		})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, "新", fields["name"])
		assert.Equal(t, "13900000000", fields["phone"])
		assert.Equal(t, model.ShopStatusOpen, fields["status"])
		// longitude/latitude/category_id 总是写入
		assert.Contains(t, fields, "longitude")
		assert.Contains(t, fields, "latitude")
		assert.Contains(t, fields, "category_id")
	})

	t.Run("Status 越界不写入", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		// Status=3 越界（仅 0-2 合法），不应写入 status 字段
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{Status: 3})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.NotContains(t, fields, "status")
	})

	t.Run("UpdateFields 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		shopRepo.updateErr = errors.New("db error")
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{Name: "x"})
		assert.Error(t, err)
	})

	t.Run("FindByUserID 非 NotFound 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findByUserErr = errors.New("db error")
		err := svc.UpdateMyShop(10, 2, &dto.UpdateShopRequest{Name: "x"})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})
}

func TestAddImage(t *testing.T) {
	t.Run("店铺不存在", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		_, err := svc.AddImage(10, 2, &dto.AddShopImageRequest{ImageURL: "a.jpg"})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	// 注：AddImage 内部先 FindByUserID(userID, regionID)，已按 userID 过滤，
	// 故 shop.UserID != userID 的 ErrShopNoPermission 分支为不可达的防御性代码，
	// 不构造对应用例。

	t.Run("正常添加含 RegionID 写入", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		info, err := svc.AddImage(10, 2, &dto.AddShopImageRequest{ImageURL: "a.jpg", Sort: 5})
		require.NoError(t, err)
		assert.Equal(t, "a.jpg", info.ImageURL)
		assert.Equal(t, 5, info.Sort)
		assert.Equal(t, uint(1), info.ShopID)
		img := imageRepo.byID[info.ID]
		require.NotNil(t, img)
		assert.Equal(t, uint(2), img.RegionID)
	})

	t.Run("Create 错误透传", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{UserID: 10})
		shopRepo.byID[1].RegionID = 2
		shopRepo.byUser[userKey(10, 2)] = shopRepo.byID[1]
		imageRepo.createErr = errors.New("db error")
		_, err := svc.AddImage(10, 2, &dto.AddShopImageRequest{ImageURL: "a.jpg"})
		assert.Error(t, err)
	})
}

func TestDeleteImage(t *testing.T) {
	t.Run("图片不存在", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.DeleteImage(10, 999)
		assert.ErrorIs(t, err, ErrShopImageNotFound)
	})

	t.Run("图片查找错误透传", func(t *testing.T) {
		svc, _, imageRepo, _ := newTestService()
		imageRepo.findErr = errors.New("db error")
		err := svc.DeleteImage(10, 1)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopImageNotFound)
	})

	t.Run("店铺不存在", func(t *testing.T) {
		svc, _, imageRepo, _ := newTestService()
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5})
		err := svc.DeleteImage(10, 1)
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("无权操作", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{UserID: 99})
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5})
		err := svc.DeleteImage(10, 1)
		assert.ErrorIs(t, err, ErrShopNoPermission)
	})

	t.Run("正常删除", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{UserID: 10})
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5})
		err := svc.DeleteImage(10, 1)
		require.NoError(t, err)
		assert.True(t, imageRepo.deletedIDs[1])
	})

	t.Run("店铺查找错误透传", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5})
		shopRepo.findErr = errors.New("db error")
		err := svc.DeleteImage(10, 1)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("Delete 错误透传", func(t *testing.T) {
		svc, shopRepo, imageRepo, _ := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{UserID: 10})
		imageRepo.byID[1] = newImage(1, model.ShopImage{ShopID: 5})
		imageRepo.deleteErr = errors.New("db error")
		err := svc.DeleteImage(10, 1)
		assert.Error(t, err)
	})
}

func TestCreateReview(t *testing.T) {
	t.Run("店铺不存在", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		_, err := svc.CreateReview(2, 10, 999, &dto.CreateReviewRequest{Rating: 5})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("店铺未审核通过", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{AuditStatus: model.AuditStatusPending})
		_, err := svc.CreateReview(2, 10, 1, &dto.CreateReviewRequest{Rating: 5})
		assert.ErrorIs(t, err, ErrShopNotApproved)

		shopRepo.byID[1].AuditStatus = model.AuditStatusRejected
		_, err = svc.CreateReview(2, 10, 1, &dto.CreateReviewRequest{Rating: 5})
		assert.ErrorIs(t, err, ErrShopNotApproved)
	})

	t.Run("正常发表含默认状态与 RegionID", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{AuditStatus: model.AuditStatusApproved})
		info, err := svc.CreateReview(2, 10, 1, &dto.CreateReviewRequest{Rating: 5, Content: "好评"})
		require.NoError(t, err)
		assert.Equal(t, 5, info.Rating)
		assert.Equal(t, "好评", info.Content)
		assert.Equal(t, model.ReviewStatusPending, info.Status) // 默认待审核
		assert.Equal(t, uint(1), info.ShopID)
		assert.Equal(t, uint(10), info.UserID)
		r := reviewRepo.byID[info.ID]
		require.NotNil(t, r)
		assert.Equal(t, uint(2), r.RegionID)
	})

	t.Run("店铺查找错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findErr = errors.New("db error")
		_, err := svc.CreateReview(2, 10, 1, &dto.CreateReviewRequest{Rating: 5})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("Create 错误透传", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{AuditStatus: model.AuditStatusApproved})
		reviewRepo.createErr = errors.New("db error")
		_, err := svc.CreateReview(2, 10, 1, &dto.CreateReviewRequest{Rating: 5})
		assert.Error(t, err)
	})
}

// ===== 管理接口测试 =====

func TestAdminList(t *testing.T) {
	t.Run("空列表", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		pagination, list, err := svc.AdminList(1, &dto.AdminShopListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), pagination.Total)
		assert.Empty(t, list)
	})

	t.Run("多结果与请求体 RegionID 覆盖", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Name: "A"})
		shopRepo.byID[1].RegionID = 1
		shopRepo.byID[2] = newShop(2, model.Shop{Name: "B"})
		shopRepo.byID[2].RegionID = 2
		// 入参 regionID=1，但 req.RegionID=2 应覆盖
		pagination, list, err := svc.AdminList(1, &dto.AdminShopListRequest{RegionID: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
		assert.Equal(t, "B", list[0].Name)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.adminListErr = errors.New("db error")
		_, _, err := svc.AdminList(1, &dto.AdminShopListRequest{})
		assert.Error(t, err)
	})
}

func TestAuditShop(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.AuditShop(999, &dto.AuditShopRequest{AuditStatus: model.AuditStatusApproved})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("审核通过置为营业中", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{AuditStatus: model.AuditStatusPending, Status: model.ShopStatusClosed})
		err := svc.AuditShop(1, &dto.AuditShopRequest{AuditStatus: model.AuditStatusApproved})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, model.AuditStatusApproved, fields["audit_status"])
		assert.Equal(t, model.ShopStatusOpen, fields["status"])
	})

	t.Run("审核拒绝置为歇业", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{AuditStatus: model.AuditStatusPending, Status: model.ShopStatusOpen})
		err := svc.AuditShop(1, &dto.AuditShopRequest{AuditStatus: model.AuditStatusRejected})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, model.AuditStatusRejected, fields["audit_status"])
		assert.Equal(t, model.ShopStatusClosed, fields["status"])
	})

	t.Run("查找错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findErr = errors.New("db error")
		err := svc.AuditShop(1, &dto.AuditShopRequest{AuditStatus: model.AuditStatusApproved})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("UpdateFields 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{})
		shopRepo.updateErr = errors.New("db error")
		err := svc.AuditShop(1, &dto.AuditShopRequest{AuditStatus: model.AuditStatusApproved})
		assert.Error(t, err)
	})
}

func TestUpdateShopStatus(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.UpdateShopStatus(999, &dto.UpdateShopStatusRequest{Status: model.ShopStatusOpen})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("正常更新", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{Status: model.ShopStatusClosed})
		err := svc.UpdateShopStatus(1, &dto.UpdateShopStatusRequest{Status: model.ShopStatusOpen})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, model.ShopStatusOpen, fields["status"])
	})

	t.Run("UpdateFields 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{})
		shopRepo.updateErr = errors.New("db error")
		err := svc.UpdateShopStatus(1, &dto.UpdateShopStatusRequest{Status: model.ShopStatusOpen})
		assert.Error(t, err)
	})
}

func TestSetRecommend(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.SetRecommend(999, &dto.SetRecommendRequest{IsRecommend: 1})
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("正常设置推荐", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{IsRecommend: 0})
		err := svc.SetRecommend(1, &dto.SetRecommendRequest{IsRecommend: 1})
		require.NoError(t, err)
		fields := shopRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, 1, fields["is_recommend"])
	})

	t.Run("UpdateFields 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{})
		shopRepo.updateErr = errors.New("db error")
		err := svc.SetRecommend(1, &dto.SetRecommendRequest{IsRecommend: 1})
		assert.Error(t, err)
	})
}

func TestDeleteShop(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.DeleteShop(999)
		assert.ErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("正常删除", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{})
		err := svc.DeleteShop(1)
		require.NoError(t, err)
		assert.True(t, shopRepo.deletedIDs[1])
	})

	t.Run("查找错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.findErr = errors.New("db error")
		err := svc.DeleteShop(1)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopNotFound)
	})

	t.Run("Delete 错误透传", func(t *testing.T) {
		svc, shopRepo, _, _ := newTestService()
		shopRepo.byID[1] = newShop(1, model.Shop{})
		shopRepo.deleteErr = errors.New("db error")
		err := svc.DeleteShop(1)
		assert.Error(t, err)
	})
}

func TestAdminReviewList(t *testing.T) {
	t.Run("空列表", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		pagination, list, err := svc.AdminReviewList(&dto.AdminReviewListRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), pagination.Total)
		assert.Empty(t, list)
	})

	t.Run("多结果与店铺过滤", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5, Status: model.ReviewStatusApproved})
		reviewRepo.byID[2] = newReview(2, model.ShopReview{ShopID: 9, Status: model.ReviewStatusApproved})
		// Status=-1 表示不筛选状态（DTO 注释：-1不筛选），零值 0 会过滤出待审核
		pagination, list, err := svc.AdminReviewList(&dto.AdminReviewListRequest{ShopID: 5, Status: -1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), pagination.Total)
		require.Len(t, list, 1)
		assert.Equal(t, uint(5), list[0].ShopID)
	})

	t.Run("仓储错误透传", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.adminListErr = errors.New("db error")
		_, _, err := svc.AdminReviewList(&dto.AdminReviewListRequest{})
		assert.Error(t, err)
	})
}

func TestAuditReview(t *testing.T) {
	t.Run("评价不存在", func(t *testing.T) {
		svc, _, _, _ := newTestService()
		err := svc.AuditReview(999, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		assert.ErrorIs(t, err, ErrShopReviewNotFound)
	})

	t.Run("评价查找错误透传", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.findErr = errors.New("db error")
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrShopReviewNotFound)
	})

	t.Run("UpdateFields 错误透传", func(t *testing.T) {
		svc, _, _, reviewRepo := newTestService()
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5})
		reviewRepo.updateErr = errors.New("db error")
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		assert.Error(t, err)
	})

	t.Run("审核通过重算评分", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{Rating: 0})
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5, Status: model.ReviewStatusPending, Rating: 4})
		reviewRepo.byID[2] = newReview(2, model.ShopReview{ShopID: 5, Status: model.ReviewStatusApproved, Rating: 5})
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		require.NoError(t, err)
		// 验证评价状态更新
		fields := reviewRepo.updatedFields[1]
		require.NotNil(t, fields)
		assert.Equal(t, model.ReviewStatusApproved, fields["status"])
		// 验证评分重算：(4+5)/2 = 4.5
		assert.Equal(t, float32(4.5), shopRepo.updateRating[5])
		assert.Equal(t, float32(4.5), shopRepo.byID[5].Rating)
	})

	t.Run("无已通过评价时评分为0", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{Rating: 4.0})
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5, Status: model.ReviewStatusPending, Rating: 4})
		// 审核拒绝，已通过评价仍为 0 条
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusRejected})
		require.NoError(t, err)
		assert.Equal(t, float32(0), shopRepo.updateRating[5])
	})

	t.Run("AvgRating 错误透传", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{})
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5})
		reviewRepo.avgErr = errors.New("db error")
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		assert.Error(t, err)
	})

	t.Run("UpdateRating 错误透传", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{})
		reviewRepo.byID[1] = newReview(1, model.ShopReview{ShopID: 5, Status: model.ReviewStatusApproved, Rating: 5})
		shopRepo.updateRatErr = errors.New("db error")
		err := svc.AuditReview(1, &dto.AuditReviewRequest{Status: model.ReviewStatusApproved})
		assert.Error(t, err)
	})
}

// ===== 评分重算辅助测试 =====

func TestRecomputeShopRating(t *testing.T) {
	t.Run("预置平均值透传", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{})
		// 预置 AvgRatingByShopID 返回值
		reviewRepo.avgResult[5] = struct {
			avg   float32
			count int64
		}{avg: 4.65, count: 3}
		err := svc.recomputeShopRating(5)
		require.NoError(t, err)
		// 4.65 保留一位小数 = 4.7（math.Round(4.65*10)/10 = 4.7，浮点误差内）
		assert.InDelta(t, float32(4.7), float64(shopRepo.updateRating[5]), 0.01)
	})

	t.Run("评分一位小数舍入", func(t *testing.T) {
		svc, shopRepo, _, reviewRepo := newTestService()
		shopRepo.byID[5] = newShop(5, model.Shop{})
		// 4.444 → 4.4
		reviewRepo.avgResult[5] = struct {
			avg   float32
			count int64
		}{avg: 4.444, count: 2}
		err := svc.recomputeShopRating(5)
		require.NoError(t, err)
		assert.InDelta(t, float32(4.4), float64(shopRepo.updateRating[5]), 0.01)
	})
}
