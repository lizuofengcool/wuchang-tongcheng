// Package service 团购优惠券业务逻辑层单元测试。
// 使用内存 mock 仓储覆盖：团购商品 CRUD 与默认值、审核/上下架、
// 优惠券创建默认值与字段校验、领取校验（禁用/未开始/过期/总量/限领/有效期计算）、
// 订单创建校验（未上架/已结束/库存/限购/优惠券三类优惠计算与适用范围）、
// 订单查询权限、取消/核销状态校验等核心逻辑，不依赖 DB。
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/groupbuy/dto"
	"wuchang-tongcheng/internal/modules/groupbuy/model"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== mockGroupBuyRepo =====

type mockGroupBuyRepo struct {
	byID         map[uint]*model.GroupBuy
	nextID       uint
	createErr    error
	findErr      error
	updateErr    error
	updateFields map[uint]map[string]interface{}
	deleteErr    error
	listErr      error
	deletedIDs   map[uint]bool
}

func newMockGroupBuyRepo() *mockGroupBuyRepo {
	return &mockGroupBuyRepo{
		byID:         make(map[uint]*model.GroupBuy),
		nextID:       1,
		updateFields: make(map[uint]map[string]interface{}),
		deletedIDs:   make(map[uint]bool),
	}
}

func (m *mockGroupBuyRepo) Create(gb *model.GroupBuy) error {
	if m.createErr != nil {
		return m.createErr
	}
	gb.ID = m.nextID
	m.nextID++
	cp := *gb
	m.byID[gb.ID] = &cp
	return nil
}

func (m *mockGroupBuyRepo) FindByID(id uint) (*model.GroupBuy, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	gb, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *gb
	return &cp, nil
}

func (m *mockGroupBuyRepo) Update(gb *model.GroupBuy) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *gb
	m.byID[gb.ID] = &cp
	return nil
}

func (m *mockGroupBuyRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updateFields[id] = fields
	gb, ok := m.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if v, ok := fields["status"]; ok {
		gb.Status = v.(int)
	}
	if v, ok := fields["audit_status"]; ok {
		gb.AuditStatus = v.(int)
	}
	if v, ok := fields["title"]; ok {
		gb.Title = v.(string)
	}
	return nil
}

func (m *mockGroupBuyRepo) Delete(id uint) error {
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

func (m *mockGroupBuyRepo) List(regionID uint, pagination *utils.Pagination, keyword string, status, isRecommend int, shopID uint) ([]model.GroupBuy, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.GroupBuy
	for _, gb := range m.byID {
		if regionID > 0 && gb.RegionID != regionID {
			continue
		}
		if status >= 0 && status <= 2 && gb.Status != status {
			continue
		}
		if (isRecommend == 0 || isRecommend == 1) && gb.IsRecommend != isRecommend {
			continue
		}
		if shopID > 0 && gb.ShopID != shopID {
			continue
		}
		list = append(list, *gb)
	}
	return list, int64(len(list)), nil
}

// ===== mockCouponRepo =====

type mockCouponRepo struct {
	byID         map[uint]*model.Coupon
	nextID       uint
	createErr    error
	findErr      error
	updateErr    error
	deleteErr    error
	listErr      error
	incrReceived map[uint]int
	availableErr error
}

func newMockCouponRepo() *mockCouponRepo {
	return &mockCouponRepo{
		byID:         make(map[uint]*model.Coupon),
		nextID:       1,
		incrReceived: make(map[uint]int),
	}
}

func (m *mockCouponRepo) Create(c *model.Coupon) error {
	if m.createErr != nil {
		return m.createErr
	}
	c.ID = m.nextID
	m.nextID++
	cp := *c
	m.byID[c.ID] = &cp
	return nil
}

func (m *mockCouponRepo) FindByID(id uint) (*model.Coupon, error) {
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

func (m *mockCouponRepo) Update(c *model.Coupon) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	cp := *c
	m.byID[c.ID] = &cp
	return nil
}

func (m *mockCouponRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return nil
}

func (m *mockCouponRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.byID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *mockCouponRepo) List(regionID uint, pagination *utils.Pagination, status, ctype int) ([]model.Coupon, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.Coupon
	for _, c := range m.byID {
		if regionID > 0 && c.RegionID != regionID {
			continue
		}
		if (status == 0 || status == 1) && c.Status != status {
			continue
		}
		if ctype >= 1 && ctype <= 3 && c.Type != ctype {
			continue
		}
		list = append(list, *c)
	}
	return list, int64(len(list)), nil
}

func (m *mockCouponRepo) IncrReceivedCount(id uint) error {
	m.incrReceived[id]++
	if c, ok := m.byID[id]; ok {
		c.ReceivedCount++
	}
	return nil
}

func (m *mockCouponRepo) AvailableList(regionID uint, pagination *utils.Pagination) ([]model.Coupon, int64, error) {
	if m.availableErr != nil {
		return nil, 0, m.availableErr
	}
	now := time.Now()
	var list []model.Coupon
	for _, c := range m.byID {
		if c.Status != 1 {
			continue
		}
		if regionID > 0 && c.RegionID != regionID {
			continue
		}
		if c.StartTime != nil && now.Before(*c.StartTime) {
			continue
		}
		if c.EndTime != nil && now.After(*c.EndTime) {
			continue
		}
		if c.TotalCount > 0 && c.ReceivedCount >= c.TotalCount {
			continue
		}
		list = append(list, *c)
	}
	return list, int64(len(list)), nil
}

// ===== mockUserCouponRepo =====

type mockUserCouponRepo struct {
	byID      map[uint]*model.UserCoupon
	nextID    uint
	createErr error
	findErr   error
	listErr   error
	countErr  error
	updateErr error
}

func newMockUserCouponRepo() *mockUserCouponRepo {
	return &mockUserCouponRepo{
		byID:   make(map[uint]*model.UserCoupon),
		nextID: 1,
	}
}

func (m *mockUserCouponRepo) Create(uc *model.UserCoupon) error {
	if m.createErr != nil {
		return m.createErr
	}
	uc.ID = m.nextID
	m.nextID++
	cp := *uc
	m.byID[uc.ID] = &cp
	return nil
}

func (m *mockUserCouponRepo) FindByID(id uint) (*model.UserCoupon, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	uc, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *uc
	return &cp, nil
}

func (m *mockUserCouponRepo) FindActiveByUserAndCoupon(userID, couponID uint) (*model.UserCoupon, error) {
	for _, uc := range m.byID {
		if uc.UserID == userID && uc.CouponID == couponID && uc.Status == 0 {
			cp := *uc
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserCouponRepo) ListByUser(userID uint, pagination *utils.Pagination, status int) ([]model.UserCoupon, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.UserCoupon
	for _, uc := range m.byID {
		if uc.UserID != userID {
			continue
		}
		if status >= 0 && status <= 2 && uc.Status != status {
			continue
		}
		list = append(list, *uc)
	}
	return list, int64(len(list)), nil
}

func (m *mockUserCouponRepo) CountByUserAndCoupon(userID, couponID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	var count int64
	for _, uc := range m.byID {
		if uc.UserID == userID && uc.CouponID == couponID {
			count++
		}
	}
	return count, nil
}

func (m *mockUserCouponRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return nil
}

// ===== mockOrderRepo =====

type mockOrderRepo struct {
	byID        map[uint]*model.GroupBuyOrder
	nextID      uint
	createErr   error
	findErr     error
	listErr     error
	countErr    error
	sumErr      error
	createTxErr error
	cancelTxErr error
	verifyErr   error
	verifyCalls []struct {
		id   uint
		code string
	}
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		byID:   make(map[uint]*model.GroupBuyOrder),
		nextID: 1,
	}
}

func (m *mockOrderRepo) Create(order *model.GroupBuyOrder) error {
	if m.createErr != nil {
		return m.createErr
	}
	order.ID = m.nextID
	m.nextID++
	cp := *order
	m.byID[order.ID] = &cp
	return nil
}

func (m *mockOrderRepo) FindByID(id uint) (*model.GroupBuyOrder, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	o, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockOrderRepo) FindByOrderNo(orderNo string) (*model.GroupBuyOrder, error) {
	for _, o := range m.byID {
		if o.OrderNo == orderNo {
			cp := *o
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOrderRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockOrderRepo) List(regionID uint, pagination *utils.Pagination, userID, groupbuyID uint, status, payStatus int) ([]model.GroupBuyOrder, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var list []model.GroupBuyOrder
	for _, o := range m.byID {
		if regionID > 0 && o.RegionID != regionID {
			continue
		}
		if userID > 0 && o.UserID != userID {
			continue
		}
		if groupbuyID > 0 && o.GroupBuyID != groupbuyID {
			continue
		}
		if status >= 0 && status <= 4 && o.Status != status {
			continue
		}
		if payStatus >= 0 && payStatus <= 2 && o.PayStatus != payStatus {
			continue
		}
		list = append(list, *o)
	}
	return list, int64(len(list)), nil
}

func (m *mockOrderRepo) CountByUserAndGroupBuy(userID, groupbuyID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	var count int64
	for _, o := range m.byID {
		if o.UserID == userID && o.GroupBuyID == groupbuyID && o.Status != 3 {
			count++
		}
	}
	return count, nil
}

func (m *mockOrderRepo) SumQuantityByUserAndGroupBuy(userID, groupbuyID uint) (int64, error) {
	if m.sumErr != nil {
		return 0, m.sumErr
	}
	var total int64
	for _, o := range m.byID {
		if o.UserID == userID && o.GroupBuyID == groupbuyID && o.Status != 3 {
			total += int64(o.Quantity)
		}
	}
	return total, nil
}

func (m *mockOrderRepo) CreateOrderInTransaction(order *model.GroupBuyOrder, groupbuyID, userCouponID uint, quantity int) error {
	if m.createTxErr != nil {
		return m.createTxErr
	}
	order.ID = m.nextID
	m.nextID++
	cp := *order
	m.byID[order.ID] = &cp
	return nil
}

func (m *mockOrderRepo) CancelOrderInTransaction(order *model.GroupBuyOrder) error {
	if m.cancelTxErr != nil {
		return m.cancelTxErr
	}
	if o, ok := m.byID[order.ID]; ok {
		o.Status = 3
		o.PayStatus = 2
	}
	return nil
}

func (m *mockOrderRepo) VerifyOrder(orderID uint, verifyCode string) error {
	m.verifyCalls = append(m.verifyCalls, struct {
		id   uint
		code string
	}{orderID, verifyCode})
	if m.verifyErr != nil {
		return m.verifyErr
	}
	return nil
}

// ===== 测试辅助 =====

func newTestService() (*service, *mockGroupBuyRepo, *mockCouponRepo, *mockUserCouponRepo, *mockOrderRepo) {
	gbRepo := newMockGroupBuyRepo()
	couponRepo := newMockCouponRepo()
	ucRepo := newMockUserCouponRepo()
	orderRepo := newMockOrderRepo()
	svc := NewService(gbRepo, couponRepo, ucRepo, orderRepo).(*service)
	return svc, gbRepo, couponRepo, ucRepo, orderRepo
}

func newGB(id uint, gb model.GroupBuy) *model.GroupBuy             { gb.ID = id; return &gb }
func newCoupon(id uint, c model.Coupon) *model.Coupon              { c.ID = id; return &c }
func newUC(id uint, uc model.UserCoupon) *model.UserCoupon         { uc.ID = id; return &uc }
func newOrder(id uint, o model.GroupBuyOrder) *model.GroupBuyOrder { o.ID = id; return &o }

// ===== 团购商品测试 =====

func TestCreateGroupBuy(t *testing.T) {
	t.Run("正常创建含默认值", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		req := &dto.CreateGroupBuyRequest{
			Title:         "双人套餐",
			GroupBuyPrice: 99.9,
			Stock:         100,
			PerLimit:      0, // 应默认为 1
			ShopID:        5,
		}
		info, err := svc.CreateGroupBuy(2, 10, req)
		require.NoError(t, err)
		assert.Equal(t, uint(1), info.ID)
		assert.Equal(t, "双人套餐", info.Title)
		assert.Equal(t, 1, info.PerLimit) // 默认值
		assert.Equal(t, 0, info.Status)   // 默认下架
		assert.Equal(t, 0, info.AuditStatus)
		assert.Equal(t, uint(10), info.UserID)
		// 验证 RegionID 写入仓储
		gb := gbRepo.byID[1]
		assert.Equal(t, uint(2), gb.RegionID)
	})

	t.Run("仓储创建失败", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.createErr = errors.New("db error")
		_, err := svc.CreateGroupBuy(1, 1, &dto.CreateGroupBuyRequest{Title: "x", GroupBuyPrice: 1})
		assert.Error(t, err)
	})
}

func TestUpdateGroupBuy(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.UpdateGroupBuy(999, &dto.UpdateGroupBuyRequest{Title: "x"})
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("编辑已审核通过重置审核状态", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{AuditStatus: 1, Title: "旧"})
		err := svc.UpdateGroupBuy(1, &dto.UpdateGroupBuyRequest{Title: "新"})
		require.NoError(t, err)
		assert.Equal(t, 0, gbRepo.byID[1].AuditStatus)
		assert.Equal(t, "新", gbRepo.byID[1].Title)
		// fields 应记录 audit_status=0
		fields := gbRepo.updateFields[1]
		assert.Equal(t, 0, fields["audit_status"])
	})

	t.Run("无字段更新直接返回", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{AuditStatus: 0})
		err := svc.UpdateGroupBuy(1, &dto.UpdateGroupBuyRequest{})
		require.NoError(t, err)
		// is_recommend 和 sort 总是写入，所以 fields 非空
		// 但当只有零值且不触发审核重置时仍会写入 is_recommend/sort
	})
}

func TestDeleteGroupBuy(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.DeleteGroupBuy(999)
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("正常删除", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{})
		err := svc.DeleteGroupBuy(1)
		require.NoError(t, err)
		assert.True(t, gbRepo.deletedIDs[1])
	})
}

func TestGetGroupBuy(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		_, err := svc.GetGroupBuy(999)
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("正常获取并转换DTO", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Title: "套餐", GroupBuyPrice: 50, Stock: 10})
		info, err := svc.GetGroupBuy(1)
		require.NoError(t, err)
		assert.Equal(t, "套餐", info.Title)
		assert.Equal(t, float64(50), info.GroupBuyPrice)
		assert.Equal(t, 10, info.Stock)
	})
}

func TestListGroupBuy(t *testing.T) {
	svc, gbRepo, _, _, _ := newTestService()
	gbRepo.byID[1] = newGB(1, model.GroupBuy{Title: "A", Status: 1})
	gbRepo.byID[1].RegionID = 1
	gbRepo.byID[2] = newGB(2, model.GroupBuy{Title: "B", Status: 0})
	gbRepo.byID[2].RegionID = 1
	gbRepo.byID[3] = newGB(3, model.GroupBuy{Title: "C", Status: 1})
	gbRepo.byID[3].RegionID = 2

	pagination, list, err := svc.ListGroupBuy(1, &dto.GroupBuyListRequest{Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, "A", list[0].Title)
}

func TestUpdateGroupBuyStatus(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.UpdateGroupBuyStatus(999, 1)
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("正常更新", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 0})
		err := svc.UpdateGroupBuyStatus(1, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, gbRepo.byID[1].Status)
	})
}

func TestAuditGroupBuy(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.AuditGroupBuy(999, 1)
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("正常审核", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{AuditStatus: 0})
		err := svc.AuditGroupBuy(1, 1)
		require.NoError(t, err)
		assert.Equal(t, 1, gbRepo.byID[1].AuditStatus)
	})
}

// ===== 优惠券测试 =====

func TestCreateCoupon(t *testing.T) {
	t.Run("默认值填充", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		req := &dto.CreateCouponRequest{
			Name:     "满100减20",
			Type:     1,
			Value:    20,
			PerLimit: 0, // 应默认为 1
			Status:   0, // 应默认为 1
		}
		info, err := svc.CreateCoupon(3, req)
		require.NoError(t, err)
		assert.Equal(t, 1, info.PerLimit)
		assert.Equal(t, 1, info.Status)
		assert.Equal(t, uint(3), couponRepo.byID[info.ID].RegionID)
	})

	t.Run("显式值保留", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		req := &dto.CreateCouponRequest{
			Name:     "折扣券",
			Type:     2,
			Value:    0.8,
			PerLimit: 5,
			Status:   1,
		}
		info, err := svc.CreateCoupon(1, req)
		require.NoError(t, err)
		assert.Equal(t, 5, info.PerLimit)
		assert.Equal(t, 1, info.Status)
	})
}

func TestUpdateCoupon(t *testing.T) {
	t.Run("不存在返回错误", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.UpdateCoupon(999, &dto.UpdateCouponRequest{Name: "x"})
		assert.ErrorIs(t, err, ErrCouponNotFound)
	})

	t.Run("无字段更新直接返回", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		couponRepo.byID[1] = newCoupon(1, model.Coupon{})
		err := svc.UpdateCoupon(1, &dto.UpdateCouponRequest{})
		require.NoError(t, err)
	})
}

func TestDeleteCoupon(t *testing.T) {
	svc, _, couponRepo, _, _ := newTestService()
	couponRepo.byID[1] = newCoupon(1, model.Coupon{})
	err := svc.DeleteCoupon(1)
	require.NoError(t, err)
	_, ok := couponRepo.byID[1]
	assert.False(t, ok)

	err = svc.DeleteCoupon(999)
	assert.ErrorIs(t, err, ErrCouponNotFound)
}

func TestReceiveCoupon(t *testing.T) {
	t.Run("优惠券不存在", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		_, err := svc.ReceiveCoupon(1, 1, 999)
		assert.ErrorIs(t, err, ErrCouponNotFound)
	})

	t.Run("优惠券已禁用", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 0})
		_, err := svc.ReceiveCoupon(1, 1, 1)
		assert.ErrorIs(t, err, ErrCouponDisabled)
	})

	t.Run("优惠券尚未开始", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		future := time.Now().Add(time.Hour)
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, StartTime: &future})
		_, err := svc.ReceiveCoupon(1, 1, 1)
		assert.ErrorIs(t, err, ErrCouponNotActive)
	})

	t.Run("优惠券已过期", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		past := time.Now().Add(-time.Hour)
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, EndTime: &past})
		_, err := svc.ReceiveCoupon(1, 1, 1)
		assert.ErrorIs(t, err, ErrCouponExpired)
	})

	t.Run("达到发放总量", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, TotalCount: 100, ReceivedCount: 100})
		_, err := svc.ReceiveCoupon(1, 1, 1)
		assert.ErrorIs(t, err, ErrCouponLimit)
	})

	t.Run("超过每人限领", func(t *testing.T) {
		svc, _, couponRepo, ucRepo, _ := newTestService()
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, PerLimit: 1})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1})
		_, err := svc.ReceiveCoupon(1, 10, 1)
		assert.ErrorIs(t, err, ErrCouponLimit)
	})

	t.Run("正常领取-固定有效期", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		end := time.Now().Add(24 * time.Hour)
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, ValidityType: 0, EndTime: &end})
		info, err := svc.ReceiveCoupon(1, 10, 1)
		require.NoError(t, err)
		assert.Equal(t, uint(10), info.UserID)
		assert.Equal(t, uint(1), info.CouponID)
		require.NotNil(t, info.ExpireAt)
		assert.Equal(t, end, *info.ExpireAt)
		assert.Equal(t, 1, couponRepo.incrReceived[1])
	})

	t.Run("正常领取-领取后N天有效期", func(t *testing.T) {
		svc, _, couponRepo, _, _ := newTestService()
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, ValidityType: 1, ValidDays: 7})
		info, err := svc.ReceiveCoupon(1, 10, 1)
		require.NoError(t, err)
		require.NotNil(t, info.ExpireAt)
		// 过期时间应在 7 天左右
		diff := info.ExpireAt.Sub(time.Now())
		assert.InDelta(t, 7*24*time.Hour, float64(diff), float64(time.Hour))
	})
}

func TestListCoupon(t *testing.T) {
	svc, _, couponRepo, _, _ := newTestService()
	couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1})
	couponRepo.byID[1].RegionID = 1
	couponRepo.byID[2] = newCoupon(2, model.Coupon{Status: 0, Type: 1})
	couponRepo.byID[2].RegionID = 1
	pagination, list, err := svc.ListCoupon(1, &dto.CouponListRequest{Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, uint(1), list[0].ID)
}

func TestAvailableCoupons(t *testing.T) {
	svc, _, couponRepo, _, _ := newTestService()
	couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1})
	couponRepo.byID[1].RegionID = 1
	couponRepo.byID[2] = newCoupon(2, model.Coupon{Status: 0})
	couponRepo.byID[2].RegionID = 1
	pagination, list, err := svc.AvailableCoupons(1, &dto.CouponListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, uint(1), list[0].ID)
}

func TestMyCoupons(t *testing.T) {
	svc, _, couponRepo, ucRepo, _ := newTestService()
	couponRepo.byID[1] = newCoupon(1, model.Coupon{Name: "券A"})
	ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
	ucRepo.byID[2] = newUC(2, model.UserCoupon{UserID: 20, CouponID: 1, Status: 0})
	pagination, list, err := svc.MyCoupons(10, &dto.CouponListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, uint(10), list[0].UserID)
	require.NotNil(t, list[0].Coupon)
	assert.Equal(t, "券A", list[0].Coupon.Name)
}

// ===== 订单测试 =====

func TestCreateOrder(t *testing.T) {
	t.Run("团购不存在", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 999, Quantity: 1})
		assert.ErrorIs(t, err, ErrGroupBuyNotFound)
	})

	t.Run("团购未上架或未审核", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 0, AuditStatus: 1, Stock: 10})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
		assert.ErrorIs(t, err, ErrGroupBuyNotActive)

		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 0, Stock: 10})
		_, err = svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
		assert.ErrorIs(t, err, ErrGroupBuyNotActive)
	})

	t.Run("团购尚未开始", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		future := time.Now().Add(time.Hour)
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, StartTime: &future, Stock: 10})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
		assert.ErrorIs(t, err, ErrGroupBuyNotActive)
	})

	t.Run("团购已结束", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		past := time.Now().Add(-time.Hour)
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, EndTime: &past, Stock: 10})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
		assert.ErrorIs(t, err, ErrGroupBuyEnded)
	})

	t.Run("库存不足", func(t *testing.T) {
		svc, gbRepo, _, _, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 5, GroupBuyPrice: 10})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 10})
		assert.ErrorIs(t, err, ErrStockInsufficient)
	})

	t.Run("超过每人限购", func(t *testing.T) {
		svc, gbRepo, _, _, orderRepo := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 2, GroupBuyPrice: 10})
		// 已买 1 件，再买 2 件超过限购 2
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, GroupBuyID: 1, Quantity: 1, Status: 1})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 2})
		assert.ErrorIs(t, err, ErrPerLimitExceeded)
	})

	t.Run("正常下单无优惠券", func(t *testing.T) {
		svc, gbRepo, _, _, orderRepo := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 50})
		info, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 3})
		require.NoError(t, err)
		assert.Equal(t, 3, info.Quantity)
		assert.Equal(t, float64(50), info.UnitPrice)
		assert.Equal(t, float64(150), info.TotalPrice)
		assert.Equal(t, float64(0), info.DiscountAmount)
		assert.Equal(t, float64(150), info.PayAmount)
		assert.Equal(t, 1, info.PayStatus)
		assert.Equal(t, 1, info.Status)
		assert.NotEmpty(t, info.OrderNo)
		assert.NotEmpty(t, info.VerifyCode)
		// 验证订单写入仓储
		assert.Equal(t, uint(10), orderRepo.byID[info.ID].UserID)
	})

	t.Run("满减优惠券", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 50})
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1, Value: 20, MinAmount: 100, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
		info, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 3, UserCouponID: 1})
		require.NoError(t, err)
		assert.Equal(t, float64(150), info.TotalPrice)
		assert.Equal(t, float64(20), info.DiscountAmount)
		assert.Equal(t, float64(130), info.PayAmount)
		assert.Equal(t, uint(1), info.CouponID)
	})

	t.Run("折扣优惠券", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 100})
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 2, Value: 0.8, MinAmount: 0, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
		info, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 2, UserCouponID: 1})
		require.NoError(t, err)
		// totalPrice=200, discount=200*(1-0.8)=40, pay=160
		assert.Equal(t, float64(200), info.TotalPrice)
		assert.InDelta(t, float64(40), info.DiscountAmount, 0.01)
		assert.InDelta(t, float64(160), info.PayAmount, 0.01)
	})

	t.Run("代金券优惠券", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 30})
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 3, Value: 10, MinAmount: 0, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
		info, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		require.NoError(t, err)
		assert.Equal(t, float64(30), info.TotalPrice)
		assert.Equal(t, float64(10), info.DiscountAmount)
		assert.Equal(t, float64(20), info.PayAmount)
	})

	t.Run("优惠券未领取或已使用", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 50})
		// 优惠券不存在
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 999})
		assert.ErrorIs(t, err, ErrCouponNotReceived)

		// 优惠券属于其他用户
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1, Value: 10, MinAmount: 0, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 99, CouponID: 1, Status: 0})
		_, err = svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		assert.ErrorIs(t, err, ErrCouponNotReceived)

		// 优惠券已使用
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 1})
		_, err = svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		assert.ErrorIs(t, err, ErrCouponNotReceived)
	})

	t.Run("优惠券已过期", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 50})
		past := time.Now().Add(-time.Hour)
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1, Value: 10, MinAmount: 0, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0, ExpireAt: &past})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		assert.ErrorIs(t, err, ErrCouponExpired)
	})

	t.Run("优惠券不满足最低消费", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 30})
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1, Value: 20, MinAmount: 100, Scope: 0})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		assert.ErrorIs(t, err, ErrCouponNotMatch)
	})

	t.Run("优惠券适用范围不匹配", func(t *testing.T) {
		svc, gbRepo, couponRepo, ucRepo, _ := newTestService()
		// 团购 ShopID=5，优惠券限定 ShopID=99
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 100, ShopID: 5})
		couponRepo.byID[1] = newCoupon(1, model.Coupon{Status: 1, Type: 1, Value: 20, MinAmount: 50, Scope: 1, ScopeID: 99})
		ucRepo.byID[1] = newUC(1, model.UserCoupon{UserID: 10, CouponID: 1, Status: 0})
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1, UserCouponID: 1})
		assert.ErrorIs(t, err, ErrCouponNotMatch)
	})

	t.Run("事务创建失败", func(t *testing.T) {
		svc, gbRepo, _, _, orderRepo := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Status: 1, AuditStatus: 1, Stock: 100, PerLimit: 5, GroupBuyPrice: 50})
		orderRepo.createTxErr = errors.New("tx error")
		_, err := svc.CreateOrder(1, 10, &dto.CreateOrderRequest{GroupBuyID: 1, Quantity: 1})
		assert.Error(t, err)
	})
}

func TestGetOrder(t *testing.T) {
	t.Run("订单不存在", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		_, err := svc.GetOrder(999, 10)
		assert.ErrorIs(t, err, ErrOrderNotFound)
	})

	t.Run("无权查看他人订单", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 99, GroupBuyID: 1})
		_, err := svc.GetOrder(1, 10)
		assert.ErrorIs(t, err, ErrOrderNoPermission)
	})

	t.Run("管理员可查看任意订单", func(t *testing.T) {
		svc, gbRepo, _, _, orderRepo := newTestService()
		gbRepo.byID[1] = newGB(1, model.GroupBuy{Title: "套餐"})
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 99, GroupBuyID: 1})
		// userID=0 表示管理员
		info, err := svc.GetOrder(1, 0)
		require.NoError(t, err)
		assert.Equal(t, uint(99), info.UserID)
		require.NotNil(t, info.GroupBuy)
		assert.Equal(t, "套餐", info.GroupBuy.Title)
	})

	t.Run("正常查看自己的订单", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, GroupBuyID: 1})
		info, err := svc.GetOrder(1, 10)
		require.NoError(t, err)
		assert.Equal(t, uint(10), info.UserID)
	})
}

func TestCancelOrder(t *testing.T) {
	t.Run("订单不存在", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.CancelOrder(999, 10)
		assert.ErrorIs(t, err, ErrOrderNotFound)
	})

	t.Run("无权取消他人订单", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 99, Status: 1})
		err := svc.CancelOrder(1, 10)
		assert.ErrorIs(t, err, ErrOrderNoPermission)
	})

	t.Run("订单状态不允许取消", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, Status: 2}) // 已核销
		err := svc.CancelOrder(1, 10)
		assert.ErrorIs(t, err, ErrOrderStatus)
	})

	t.Run("正常取消", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, Status: 1})
		err := svc.CancelOrder(1, 10)
		require.NoError(t, err)
		assert.Equal(t, 3, orderRepo.byID[1].Status)
	})
}

func TestVerifyOrder(t *testing.T) {
	t.Run("订单不存在", func(t *testing.T) {
		svc, _, _, _, _ := newTestService()
		err := svc.VerifyOrder(999, "12345678")
		assert.ErrorIs(t, err, ErrOrderNotFound)
	})

	t.Run("订单状态不允许核销", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{Status: 0}) // 待支付
		err := svc.VerifyOrder(1, "12345678")
		assert.ErrorIs(t, err, ErrOrderStatus)
	})

	t.Run("正常核销", func(t *testing.T) {
		svc, _, _, _, orderRepo := newTestService()
		orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{Status: 1})
		err := svc.VerifyOrder(1, "87654321")
		require.NoError(t, err)
		require.Len(t, orderRepo.verifyCalls, 1)
		assert.Equal(t, uint(1), orderRepo.verifyCalls[0].id)
		assert.Equal(t, "87654321", orderRepo.verifyCalls[0].code)
	})
}

func TestMyOrders(t *testing.T) {
	svc, gbRepo, _, _, orderRepo := newTestService()
	gbRepo.byID[1] = newGB(1, model.GroupBuy{Title: "套餐A"})
	orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, GroupBuyID: 1, Status: 1})
	orderRepo.byID[2] = newOrder(2, model.GroupBuyOrder{UserID: 20, GroupBuyID: 1, Status: 1})
	pagination, list, err := svc.MyOrders(10, &dto.OrderListRequest{Status: -1, PayStatus: -1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, uint(10), list[0].UserID)
	require.NotNil(t, list[0].GroupBuy)
	assert.Equal(t, "套餐A", list[0].GroupBuy.Title)
}

func TestAdminOrderList(t *testing.T) {
	svc, gbRepo, _, _, orderRepo := newTestService()
	gbRepo.byID[1] = newGB(1, model.GroupBuy{Title: "套餐A"})
	orderRepo.byID[1] = newOrder(1, model.GroupBuyOrder{UserID: 10, GroupBuyID: 1, Status: 1})
	orderRepo.byID[1].RegionID = 1
	orderRepo.byID[2] = newOrder(2, model.GroupBuyOrder{UserID: 20, GroupBuyID: 1, Status: 1})
	orderRepo.byID[2].RegionID = 2
	pagination, list, err := svc.AdminOrderList(1, &dto.OrderListRequest{Status: -1, PayStatus: -1})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pagination.Total)
	require.Len(t, list, 1)
	assert.Equal(t, uint(10), list[0].UserID)
}
