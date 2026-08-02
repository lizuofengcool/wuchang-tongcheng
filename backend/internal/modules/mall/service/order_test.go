// Package service 同城商城订单主表 service 层单元测试。
// 使用内存 mock 仓储覆盖订单核心生命周期逻辑，不依赖 DB：
//   - Create：空商品/地址校验/地址归属/商品不存在/商品下架/跨店铺/SKU 不存在/SKU 归属/
//     SKU 库存不足/无规格商品库存不足/店铺不存在/购物车结算清空/仓储错误
//   - Cancel：未找到/非本人/状态非法（已发货后不可取消）/成功
//   - AdminClose：未找到/状态非法（已完成及以后不可关闭）/成功
//   - Ship：未找到/状态非法（仅已付款可发货）/成功
//   - Confirm：未找到/非本人/状态非法（仅已发货可确认）/成功
//   - Complete：未找到/状态非法（仅已收货可完成）/成功
//   - Delete：未找到/成功
//   - GetByID / GetByOrderNo：未找到/成功（含明细）/明细加载失败容忍
//   - ListByUser / ListByShop / AdminList：分页与过滤透传/仓储错误
//   - BatchUpdateStatus / CountByStatus / Summary：透传与日期解析（合法/非法）
//   - AutoClose / AutoConfirm / AutoReview：空列表/批量推进/仓储错误
package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

// ===== 测试辅助函数 =====

func timePtrO(t time.Time) *time.Time { return &t }
func intPtrO(i int) *int               { return &i }

// newOrder 构造带 ID/RegionID/用户的订单
func newOrder(id, regionID, userID, shopID uint, o model.Order) *model.Order {
	o.ID = id
	o.RegionID = regionID
	o.UserID = userID
	o.ShopID = shopID
	if o.OrderNo == "" {
		o.OrderNo = "MAL000000000" // 测试固定订单号
	}
	return &o
}

// newProduct 构造带 ID 的商品
func newProduct(id, shopID uint, p model.Product) *model.Product {
	p.ID = id
	p.ShopID = shopID
	return &p
}

// newSku 构造带 ID 的 SKU
func newSku(id, productID uint, s model.Sku) *model.Sku {
	s.ID = id
	s.ProductID = productID
	return &s
}

// newAddr 构造带 ID 的收货地址
func newAddr(id, userID uint, a model.Address) *model.Address {
	a.ID = id
	a.UserID = userID
	return &a
}

// newShopMall 构造带 ID 的店铺
func newShopMall(id uint, s model.Shop) *model.Shop {
	s.ID = id
	return &s
}

// newCartItem 构造购物车项
func newCartItem(id, userID, skuID uint) *model.Cart {
	c := &model.Cart{UserID: userID, SkuID: skuID}
	c.ID = id
	return c
}

// =====================================================================
// ===== mock: OrderRepository =====
// =====================================================================

type mockOrderRepo struct {
	byID       map[uint]*model.Order
	byOrderNo  map[string]*model.Order
	nextID     uint
	createErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listErr    error
	batchErr   error
	countErr   error
	summaryErr error

	summaryResult *repository.OrderSummaryResult

	// 自动任务列表
	autoCloseList   []model.Order
	autoConfirmList []model.Order
	autoReviewList  []model.Order
	autoListErr     error

	// 列表返回值
	listByUserRet []model.Order
	listByShopRet []model.Order
	adminListRet  []model.Order
	listTotal     int64

	// 调用记录
	createdOrder   *model.Order
	createdItems   []model.OrderItem
	updatedFields  map[uint]map[string]interface{}
	deletedIDs     map[uint]bool
	batchCallIDs   []uint
	batchCallStat  int
	countCallStat  int
	updateCallCnt  int
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		byID:          make(map[uint]*model.Order),
		byOrderNo:     make(map[string]*model.Order),
		nextID:        1,
		updatedFields: make(map[uint]map[string]interface{}),
		deletedIDs:    make(map[uint]bool),
	}
}

func (m *mockOrderRepo) Create(o *model.Order, items []model.OrderItem) error {
	if m.createErr != nil {
		return m.createErr
	}
	if o.ID == 0 {
		o.ID = m.nextID
		m.nextID++
	}
	m.byID[o.ID] = o
	m.byOrderNo[o.OrderNo] = o
	m.createdOrder = o
	m.createdItems = items
	return nil
}

func (m *mockOrderRepo) FindByID(id uint) (*model.Order, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	o, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return o, nil
}

func (m *mockOrderRepo) FindByOrderNo(orderNo string) (*model.Order, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	o, ok := m.byOrderNo[orderNo]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return o, nil
}

func (m *mockOrderRepo) Update(o *model.Order) error          { return m.updateErr }
func (m *mockOrderRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.updateCallCnt++
	if m.updatedFields[id] == nil {
		m.updatedFields[id] = make(map[string]interface{})
	}
	for k, v := range fields {
		m.updatedFields[id][k] = v
	}
	return nil
}
func (m *mockOrderRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs[id] = true
	return nil
}

func (m *mockOrderRepo) ListByUser(userID uint, _ *utils.Pagination, _ repository.OrderListOptions) ([]model.Order, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listByUserRet, m.listTotal, nil
}
func (m *mockOrderRepo) ListByShop(shopID uint, _ *utils.Pagination, _ repository.OrderListOptions) ([]model.Order, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listByShopRet, m.listTotal, nil
}
func (m *mockOrderRepo) AdminList(_ *utils.Pagination, _ repository.AdminOrderListOptions) ([]model.Order, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.adminListRet, m.listTotal, nil
}

func (m *mockOrderRepo) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	return m.UpdateFields(id, fields)
}
func (m *mockOrderRepo) BatchUpdateStatus(ids []uint, status int) error {
	if m.batchErr != nil {
		return m.batchErr
	}
	m.batchCallIDs = ids
	m.batchCallStat = status
	return nil
}

func (m *mockOrderRepo) ListAutoClose(_ *utils.Pagination) ([]model.Order, error) {
	if m.autoListErr != nil {
		return nil, m.autoListErr
	}
	return m.autoCloseList, nil
}
func (m *mockOrderRepo) ListAutoConfirm(_ *utils.Pagination) ([]model.Order, error) {
	if m.autoListErr != nil {
		return nil, m.autoListErr
	}
	return m.autoConfirmList, nil
}
func (m *mockOrderRepo) ListAutoReview(_ *utils.Pagination) ([]model.Order, error) {
	if m.autoListErr != nil {
		return nil, m.autoListErr
	}
	return m.autoReviewList, nil
}

func (m *mockOrderRepo) CountByStatus(userID uint, status int) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	m.countCallStat = status
	return m.listTotal, nil
}

func (m *mockOrderRepo) Summary(opts repository.OrderSummaryOptions) (*repository.OrderSummaryResult, error) {
	if m.summaryErr != nil {
		return nil, m.summaryErr
	}
	if m.summaryResult == nil {
		return &repository.OrderSummaryResult{}, nil
	}
	return m.summaryResult, nil
}

// =====================================================================
// ===== mock: OrderItemRepository =====
// =====================================================================

type mockOrderItemRepo struct {
	byOrder      map[uint][]model.OrderItem
	listByOrdErr error
}

func newMockOrderItemRepo() *mockOrderItemRepo {
	return &mockOrderItemRepo{byOrder: make(map[uint][]model.OrderItem)}
}

func (m *mockOrderItemRepo) Create(item *model.OrderItem) error { return nil }
func (m *mockOrderItemRepo) BatchCreate(items []model.OrderItem) error {
	return nil
}
func (m *mockOrderItemRepo) FindByID(id uint) (*model.OrderItem, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockOrderItemRepo) Update(item *model.OrderItem) error                  { return nil }
func (m *mockOrderItemRepo) UpdateFields(id uint, fields map[string]interface{}) error { return nil }
func (m *mockOrderItemRepo) Delete(id uint) error                                { return nil }
func (m *mockOrderItemRepo) ListByOrder(orderID uint) ([]model.OrderItem, error) {
	if m.listByOrdErr != nil {
		return nil, m.listByOrdErr
	}
	return m.byOrder[orderID], nil
}
func (m *mockOrderItemRepo) ListByOrders(orderIDs []uint) ([]model.OrderItem, error) {
	return nil, nil
}
func (m *mockOrderItemRepo) List(opts repository.OrderItemListOptions, pagination *utils.Pagination) ([]model.OrderItem, int64, error) {
	return nil, 0, nil
}
func (m *mockOrderItemRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.OrderItem, int64, error) {
	return nil, 0, nil
}
func (m *mockOrderItemRepo) UpdateReviewStatus(id uint, hasReview bool, reviewID uint) error {
	return nil
}
func (m *mockOrderItemRepo) UpdateRefundStatus(id uint, refundStatus int, refundID uint) error {
	return nil
}

// =====================================================================
// ===== mock: AddressRepository =====
// =====================================================================

type mockAddressRepo struct {
	byID    map[uint]*model.Address
	findErr error
}

func newMockAddressRepo() *mockAddressRepo {
	return &mockAddressRepo{byID: make(map[uint]*model.Address)}
}

func (m *mockAddressRepo) Create(a *model.Address) error                              { return nil }
func (m *mockAddressRepo) FindByID(id uint) (*model.Address, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	a, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return a, nil
}
func (m *mockAddressRepo) Update(a *model.Address) error                              { return nil }
func (m *mockAddressRepo) UpdateFields(id uint, fields map[string]interface{}) error { return nil }
func (m *mockAddressRepo) Delete(id uint) error                                       { return nil }
func (m *mockAddressRepo) ListByUser(userID uint) ([]model.Address, error)            { return nil, nil }
func (m *mockAddressRepo) List(opts repository.AddressListOptions, pagination *utils.Pagination) ([]model.Address, int64, error) {
	return nil, 0, nil
}
func (m *mockAddressRepo) FindDefault(userID uint) (*model.Address, error) { return nil, gorm.ErrRecordNotFound }
func (m *mockAddressRepo) ClearDefault(userID uint) error                  { return nil }
func (m *mockAddressRepo) SetDefault(userID, id uint) error                { return nil }
func (m *mockAddressRepo) CountByUser(userID uint) (int64, error)          { return 0, nil }

// =====================================================================
// ===== mock: ProductRepository =====
// =====================================================================

type mockProductRepo struct {
	byID    map[uint]*model.Product
	findErr error
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{byID: make(map[uint]*model.Product)}
}

func (m *mockProductRepo) Create(p *model.Product) error                              { return nil }
func (m *mockProductRepo) FindByID(id uint) (*model.Product, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	p, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return p, nil
}
func (m *mockProductRepo) Update(p *model.Product) error                              { return nil }
func (m *mockProductRepo) UpdateFields(id uint, fields map[string]interface{}) error { return nil }
func (m *mockProductRepo) Delete(id uint) error                                       { return nil }
func (m *mockProductRepo) List(regionID uint, pagination *utils.Pagination, opts repository.ProductListOptions) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) AdminList(pagination *utils.Pagination, opts repository.ProductAdminListOptions) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) ListFeatured(regionID uint, limit int) ([]model.Product, error) { return nil, nil }
func (m *mockProductRepo) ListHot(regionID uint, limit int) ([]model.Product, error)      { return nil, nil }
func (m *mockProductRepo) ListNew(regionID uint, limit int) ([]model.Product, error)      { return nil, nil }
func (m *mockProductRepo) IncrViewCount(id uint) error                                    { return nil }
func (m *mockProductRepo) IncrFavoriteCount(id uint) error                                { return nil }
func (m *mockProductRepo) DecrFavoriteCount(id uint) error                                { return nil }
func (m *mockProductRepo) IncrSales(id uint, quantity int) error                          { return nil }
func (m *mockProductRepo) IncrReviewCount(id uint) error                                  { return nil }
func (m *mockProductRepo) UpdateRating(id uint, rating float64, goodRate float64, reviewCount int) error {
	return nil
}
func (m *mockProductRepo) UpdateStock(id uint, stock int) error                  { return nil }
func (m *mockProductRepo) UpdatePriceRange(id uint, minPrice, maxPrice float64) error { return nil }

// =====================================================================
// ===== mock: SkuRepository =====
// =====================================================================

type mockSkuRepo struct {
	byID    map[uint]*model.Sku
	findErr error
}

func newMockSkuRepo() *mockSkuRepo {
	return &mockSkuRepo{byID: make(map[uint]*model.Sku)}
}

func (m *mockSkuRepo) Create(s *model.Sku) error                              { return nil }
func (m *mockSkuRepo) FindByID(id uint) (*model.Sku, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return s, nil
}
func (m *mockSkuRepo) FindBySkuCode(shopID uint, code string) (*model.Sku, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockSkuRepo) Update(s *model.Sku) error                              { return nil }
func (m *mockSkuRepo) UpdateFields(id uint, fields map[string]interface{}) error { return nil }
func (m *mockSkuRepo) Delete(id uint) error                                   { return nil }
func (m *mockSkuRepo) DeleteByProduct(productID uint) error                   { return nil }
func (m *mockSkuRepo) ListByProduct(productID uint) ([]model.Sku, error)      { return nil, nil }
func (m *mockSkuRepo) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Sku, int64, error) {
	return nil, 0, nil
}
func (m *mockSkuRepo) BatchCreate(skus []model.Sku) error          { return nil }
func (m *mockSkuRepo) ReplaceByProduct(productID uint, skus []model.Sku) error {
	return nil
}
func (m *mockSkuRepo) UpdateStock(id uint, stock int) error        { return nil }
func (m *mockSkuRepo) IncrSales(id uint, quantity int) error       { return nil }
func (m *mockSkuRepo) DecrStock(id uint, quantity int) error       { return nil }
func (m *mockSkuRepo) IncStock(id uint, quantity int) error        { return nil }
func (m *mockSkuRepo) BatchUpdateStock(items []repository.SkuStockUpdateItem) error {
	return nil
}

// =====================================================================
// ===== mock: ShopRepository =====
// =====================================================================

type mockShopRepo struct {
	byID    map[uint]*model.Shop
	findErr error
}

func newMockShopRepo() *mockShopRepo {
	return &mockShopRepo{byID: make(map[uint]*model.Shop)}
}

func (m *mockShopRepo) Create(s *model.Shop) error                              { return nil }
func (m *mockShopRepo) FindByID(id uint) (*model.Shop, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	s, ok := m.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return s, nil
}
func (m *mockShopRepo) FindByUserID(userID uint) (*model.Shop, error) { return nil, gorm.ErrRecordNotFound }
func (m *mockShopRepo) Update(s *model.Shop) error                    { return nil }
func (m *mockShopRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	return nil
}
func (m *mockShopRepo) Delete(id uint) error { return nil }
func (m *mockShopRepo) List(regionID uint, pagination *utils.Pagination, opts repository.ShopListOptions) ([]model.Shop, int64, error) {
	return nil, 0, nil
}
func (m *mockShopRepo) AdminList(pagination *utils.Pagination, opts repository.ShopAdminListOptions) ([]model.Shop, int64, error) {
	return nil, 0, nil
}
func (m *mockShopRepo) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Shop, int64, error) {
	return nil, 0, nil
}
func (m *mockShopRepo) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Shop, int64, error) {
	return nil, 0, nil
}
func (m *mockShopRepo) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Shop, int64, error) {
	return nil, 0, nil
}
func (m *mockShopRepo) IncrViewCount(id uint) error              { return nil }
func (m *mockShopRepo) IncrProductCount(id uint) error           { return nil }
func (m *mockShopRepo) DecrProductCount(id uint) error           { return nil }
func (m *mockShopRepo) IncrOrderCount(id uint) error             { return nil }
func (m *mockShopRepo) IncrSaleAmount(id uint, amount float64) error { return nil }
func (m *mockShopRepo) UpdateRating(id uint, rating float64, reviewCount int) error {
	return nil
}

// =====================================================================
// ===== mock: CartRepository =====
// =====================================================================

type mockCartRepo struct {
	byUserSku map[uint]*model.Cart // key = skuID（测试中 userID 固定）
	findErr   error
	deleteErr error
	deletedIDs map[uint]bool
}

func newMockCartRepo() *mockCartRepo {
	return &mockCartRepo{byUserSku: make(map[uint]*model.Cart), deletedIDs: make(map[uint]bool)}
}

func (m *mockCartRepo) Create(c *model.Cart) error                              { return nil }
func (m *mockCartRepo) FindByID(id uint) (*model.Cart, error)                   { return nil, gorm.ErrRecordNotFound }
func (m *mockCartRepo) Update(c *model.Cart) error                              { return nil }
func (m *mockCartRepo) UpdateFields(id uint, fields map[string]interface{}) error { return nil }
func (m *mockCartRepo) Delete(id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedIDs[id] = true
	return nil
}
func (m *mockCartRepo) ListByUser(userID uint) ([]model.Cart, error)            { return nil, nil }
func (m *mockCartRepo) ListByUserAndShop(userID, shopID uint) ([]model.Cart, error) {
	return nil, nil
}
func (m *mockCartRepo) ListSelected(userID uint) ([]model.Cart, error) { return nil, nil }
func (m *mockCartRepo) FindByUserAndSku(userID, skuID uint) (*model.Cart, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	c, ok := m.byUserSku[skuID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return c, nil
}
func (m *mockCartRepo) BatchDelete(ids []uint) error                  { return nil }
func (m *mockCartRepo) DeleteByUser(userID uint) error                { return nil }
func (m *mockCartRepo) DeleteByUserAndShop(userID, shopID uint) error { return nil }
func (m *mockCartRepo) SelectAll(userID uint, selected int) error     { return nil }
func (m *mockCartRepo) SelectItems(userID uint, ids []uint, selected int) error {
	return nil
}
func (m *mockCartRepo) CountByUser(userID uint) (int64, error)            { return 0, nil }
func (m *mockCartRepo) CountSelectedByUser(userID uint) (int64, error)    { return 0, nil }

// =====================================================================
// ===== 装配辅助 =====
// =====================================================================

// orderMocks 汇总所有 mock，便于测试用例快速装配
type orderMocks struct {
	order   *mockOrderRepo
	item    *mockOrderItemRepo
	address *mockAddressRepo
	product *mockProductRepo
	sku     *mockSkuRepo
	shop    *mockShopRepo
	cart    *mockCartRepo
}

func newOrderMocks() *orderMocks {
	return &orderMocks{
		order:   newMockOrderRepo(),
		item:    newMockOrderItemRepo(),
		address: newMockAddressRepo(),
		product: newMockProductRepo(),
		sku:     newMockSkuRepo(),
		shop:    newMockShopRepo(),
		cart:    newMockCartRepo(),
	}
}

func (mk *orderMocks) svc() OrderService {
	return NewOrderService(mk.order, mk.item, mk.address, mk.product, mk.sku, mk.shop, mk.cart)
}

// seedCreateEnv 装配下单所需的基础数据：地址、商品、店铺
func (mk *orderMocks) seedCreateEnv() {
	const userID, shopID, productID, addrID uint = 100, 200, 300, 400
	mk.address.byID[addrID] = newAddr(addrID, userID, model.Address{
		Name:     "张三",
		Phone:    "13800000000",
		Province: "湖北省",
		City:     "武汉市",
		District: "武昌区",
		Detail:   "中南路1号",
		ZipCode:  "430071",
	})
	mk.product.byID[productID] = newProduct(productID, shopID, model.Product{
		Name:      "测试商品",
		MainImage: "https://example.com/p.jpg",
		Price:     99.5,
		Stock:     50,
		Status:    model.ProductStatusOnSale,
	})
	mk.shop.byID[shopID] = newShopMall(shopID, model.Shop{ShopName: "测试店铺"})
}

// =====================================================================
// ===== Create 测试 =====
// =====================================================================

func TestOrder_Create_Success_NoSku(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	svc := mk.svc()

	req := &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{
			{ProductID: 300, Quantity: 2},
		},
		AddressID: 400,
		Remark:    "尽快发货",
		Source:    "app",
	}

	info, err := svc.Create(1, 100, "李四", "13900000000", "https://example.com/a.jpg", "127.0.0.1", "ua", req)
	require.NoError(t, err)
	require.NotNil(t, info)

	// 订单基础字段
	assert.Equal(t, uint(1), info.RegionID)
	assert.Equal(t, uint(100), info.UserID)
	assert.Equal(t, uint(200), info.ShopID)
	assert.Equal(t, "测试店铺", info.ShopName)
	assert.Equal(t, "李四", info.BuyerName)
	assert.Equal(t, "13900000000", info.BuyerPhone)
	assert.Equal(t, model.OrderStatusPending, info.Status)
	assert.Equal(t, "待付款", info.StatusText)
	assert.Equal(t, "尽快发货", info.Remark)
	assert.Equal(t, "app", info.Source)
	assert.Equal(t, "127.0.0.1", info.ClientIP)
	assert.Equal(t, "ua", info.UserAgent)

	// 地址快照
	assert.Equal(t, "张三", info.ReceiverName)
	assert.Equal(t, "13800000000", info.ReceiverPhone)
	assert.Equal(t, "湖北省", info.Province)
	assert.Equal(t, "武汉市", info.City)
	assert.Equal(t, "武昌区", info.District)
	assert.Equal(t, "中南路1号", info.Address)

	// 金额 = 单价 * 数量
	assert.Equal(t, 199.0, info.TotalAmount)
	assert.Equal(t, 199.0, info.PayAmount)

	// 自动关闭时间已设置（15 分钟）
	require.NotNil(t, info.AutoCloseAt)
	assert.True(t, info.AutoCloseAt.After(time.Now()))

	// 订单号前缀
	assert.Contains(t, info.OrderNo, "MAL")

	// 明细
	require.Len(t, info.Items, 1)
	it := info.Items[0]
	assert.Equal(t, uint(300), it.ProductID)
	assert.Equal(t, "测试商品", it.ProductName)
	assert.Equal(t, "https://example.com/p.jpg", it.MainImage)
	assert.Equal(t, 99.5, it.Price)
	assert.Equal(t, 2, it.Quantity)
	assert.Equal(t, 199.0, it.TotalAmount)
	assert.Equal(t, model.OrderStatusPending, it.Status)
	assert.Equal(t, "待付款", it.StatusText)

	// 仓储被调用
	require.NotNil(t, mk.order.createdOrder)
	assert.Equal(t, info.OrderNo, mk.order.createdOrder.OrderNo)
	// 明细订单号同步
	require.Len(t, mk.order.createdItems, 1)
	assert.Equal(t, info.OrderNo, mk.order.createdItems[0].OrderNo)
}

func TestOrder_Create_Success_WithSku(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 增补 SKU
	mk.sku.byID[500] = newSku(500, 300, model.Sku{
		Name:    "红色-XL",
		SkuCode: "SKU-RED-XL",
		Specs:   model.JSONB([]byte(`[{"name":"颜色","value":"红"}]`)),
		Price:   80.0,
		Stock:   10,
	})
	svc := mk.svc()

	req := &dto.CreateOrderRequest{
		Items:     []dto.OrderItemRequest{{ProductID: 300, SkuID: 500, Quantity: 3}},
		AddressID: 400,
	}
	info, err := svc.Create(1, 100, "李四", "13900000000", "", "", "", req)
	require.NoError(t, err)
	require.Len(t, info.Items, 1)
	it := info.Items[0]
	assert.Equal(t, uint(500), it.SkuID)
	assert.Equal(t, "红色-XL", it.SkuName)
	assert.Equal(t, "SKU-RED-XL", it.SkuCode)
	assert.Equal(t, `[{"name":"颜色","value":"红"}]`, it.SkuSpecs)
	assert.Equal(t, 80.0, it.Price)
	assert.Equal(t, 3, it.Quantity)
	assert.Equal(t, 240.0, info.TotalAmount)
}

func TestOrder_Create_EmptyItems(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品不能为空")
}

func TestOrder_Create_AddressNotFound(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 9999,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "收货地址不存在")
}

func TestOrder_Create_AddressNotOwner(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 地址属于 user 100，用 user 200 下单
	_, err := mk.svc().Create(1, 200, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权使用他人收货地址")
}

func TestOrder_Create_AddressRepoError(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.address.findErr = errors.New("db down")
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Equal(t, "db down", err.Error())
}

func TestOrder_Create_ProductNotFound(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 8888, Quantity: 1}}, AddressID: 400,
	})
	assert.ErrorIs(t, err, ErrProductNotFound)
}

func TestOrder_Create_ProductOffShelf(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.product.byID[300].Status = model.ProductStatusOffShelf
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "商品已下架")
}

func TestOrder_Create_MultiShopForbidden(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 第二个商品属于不同店铺
	mk.product.byID[301] = newProduct(301, 999, model.Product{
		Name: "另一店商品", Price: 10, Stock: 5, Status: model.ProductStatusOnSale,
	})
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{
			{ProductID: 300, Quantity: 1},
			{ProductID: 301, Quantity: 1},
		},
		AddressID: 400,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "同一店铺")
}

func TestOrder_Create_SkuNotFound(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, SkuID: 7777, Quantity: 1}}, AddressID: 400,
	})
	assert.ErrorIs(t, err, ErrSkuNotFound)
}

func TestOrder_Create_SkuWrongProduct(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// SKU 属于 product 300，但请求 product 301（同店铺）
	mk.product.byID[301] = newProduct(301, 200, model.Product{
		Name: "同店商品2", Price: 10, Stock: 5, Status: model.ProductStatusOnSale,
	})
	mk.sku.byID[500] = newSku(500, 300, model.Sku{Price: 80, Stock: 10})
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 301, SkuID: 500, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SKU 不属于该商品")
}

func TestOrder_Create_SkuStockShort(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.sku.byID[500] = newSku(500, 300, model.Sku{Price: 80, Stock: 2})
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, SkuID: 500, Quantity: 5}}, AddressID: 400,
	})
	assert.ErrorIs(t, err, ErrOrderStockShort)
}

func TestOrder_Create_ProductStockShort(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.product.byID[300].Stock = 1
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 5}}, AddressID: 400,
	})
	assert.ErrorIs(t, err, ErrOrderStockShort)
}

func TestOrder_Create_ShopNotFound(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	delete(mk.shop.byID, 200)
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.ErrorIs(t, err, ErrShopNotFound)
}

func TestOrder_Create_FromCartClearsItems(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 购物车存在该 SKU 项
	mk.cart.byUserSku[0] = newCartItem(555, 100, 0)
	svc := mk.svc()

	req := &dto.CreateOrderRequest{
		Items:     []dto.OrderItemRequest{{ProductID: 300, SkuID: 0, Quantity: 1}},
		AddressID: 400,
		FromCart:  true,
	}
	_, err := svc.Create(1, 100, "李四", "139", "", "", "", req)
	require.NoError(t, err)
	// 购物车项 555 应被删除
	assert.True(t, mk.cart.deletedIDs[555])
}

func TestOrder_Create_FromCartNoItem(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 购物车无对应项：FindBYUserAndSku 返回 ErrRecordNotFound，不应报错
	_, err := mk.svc().Create(1, 100, "李四", "139", "", "", "", &dto.CreateOrderRequest{
		Items:     []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}},
		AddressID: 400,
		FromCart:  true,
	})
	require.NoError(t, err)
	assert.Empty(t, mk.cart.deletedIDs)
}

func TestOrder_Create_RepoCreateError(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.order.createErr = errors.New("insert failed")
	_, err := mk.svc().Create(1, 100, "李四", "139", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Equal(t, "insert failed", err.Error())
}

func TestOrder_Create_ProductRepoError(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.product.findErr = errors.New("product db error")
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Equal(t, "product db error", err.Error())
}

func TestOrder_Create_SkuRepoError(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	mk.sku.findErr = errors.New("sku db error")
	_, err := mk.svc().Create(1, 100, "", "", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{{ProductID: 300, SkuID: 500, Quantity: 1}}, AddressID: 400,
	})
	assert.Error(t, err)
	assert.Equal(t, "sku db error", err.Error())
}

func TestOrder_Create_TotalAmountMultipleItems(t *testing.T) {
	mk := newOrderMocks()
	mk.seedCreateEnv()
	// 同店铺第二个商品
	mk.product.byID[301] = newProduct(301, 200, model.Product{
		Name: "商品2", Price: 10.5, Stock: 5, Status: model.ProductStatusOnSale,
	})
	info, err := mk.svc().Create(1, 100, "李四", "139", "", "", "", &dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{
			{ProductID: 300, Quantity: 2}, // 99.5 * 2 = 199
			{ProductID: 301, Quantity: 4}, // 10.5 * 4 = 42
		},
		AddressID: 400,
	})
	require.NoError(t, err)
	assert.Equal(t, 241.0, info.TotalAmount)
	assert.Len(t, info.Items, 2)
}

// =====================================================================
// ===== Cancel 测试 =====
// =====================================================================

func TestOrder_Cancel_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPending})
	err := mk.svc().Cancel(1, 100, &dto.CancelOrderRequest{Reason: "不想买了"})
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	require.NotNil(t, fields)
	assert.Equal(t, model.OrderStatusCancelled, fields["status"])
	assert.Equal(t, "不想买了", fields["admin_remark"])
	assert.NotNil(t, fields["cancelled_at"])
}

func TestOrder_Cancel_PaidAllowed(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	err := mk.svc().Cancel(1, 100, &dto.CancelOrderRequest{Reason: ""})
	require.NoError(t, err)
	// 无 reason 时不应写 admin_remark
	fields := mk.order.updatedFields[1]
	_, hasRemark := fields["admin_remark"]
	assert.False(t, hasRemark)
}

func TestOrder_Cancel_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().Cancel(999, 100, &dto.CancelOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_Cancel_NotOwner(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPending})
	err := mk.svc().Cancel(1, 200, &dto.CancelOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderNotOwner)
}

func TestOrder_Cancel_StatusInvalid(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusShipped})
	err := mk.svc().Cancel(1, 100, &dto.CancelOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderStatusInvalid)
}

func TestOrder_Cancel_RepoFindError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.findErr = errors.New("find err")
	err := mk.svc().Cancel(1, 100, &dto.CancelOrderRequest{})
	assert.Error(t, err)
}

// =====================================================================
// ===== AdminClose 测试 =====
// =====================================================================

func TestOrder_AdminClose_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPending})
	err := mk.svc().AdminClose(1, &dto.AdminCloseOrderRequest{AdminRemark: "违规"})
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	require.NotNil(t, fields)
	assert.Equal(t, model.OrderStatusClosed, fields["status"])
	assert.Equal(t, "违规", fields["admin_remark"])
	assert.NotNil(t, fields["cancelled_at"])
}

func TestOrder_AdminClose_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().AdminClose(999, &dto.AdminCloseOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_AdminClose_StatusInvalidCompleted(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusCompleted})
	err := mk.svc().AdminClose(1, &dto.AdminCloseOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderStatusInvalid)
}

func TestOrder_AdminClose_StatusInvalidClosed(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusClosed})
	err := mk.svc().AdminClose(1, &dto.AdminCloseOrderRequest{})
	assert.ErrorIs(t, err, ErrOrderStatusInvalid)
}

func TestOrder_AdminClose_ReceivedAllowed(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusReceived})
	err := mk.svc().AdminClose(1, &dto.AdminCloseOrderRequest{AdminRemark: "强制关闭"})
	require.NoError(t, err)
}

// =====================================================================
// ===== Ship 测试 =====
// =====================================================================

func TestOrder_Ship_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	err := mk.svc().Ship(1, 999, &dto.ShipOrderRequest{
		LogisticsCompany: "顺丰",
		LogisticsNo:      "SF123",
		SellerRemark:     "已发出",
	})
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	require.NotNil(t, fields)
	assert.Equal(t, model.OrderStatusShipped, fields["status"])
	assert.Equal(t, "顺丰", fields["logistics_company"])
	assert.Equal(t, "SF123", fields["logistics_no"])
	assert.Equal(t, "已发出", fields["seller_remark"])
	assert.NotNil(t, fields["shipped_at"])
	assert.NotNil(t, fields["auto_confirm_at"])
}

func TestOrder_Ship_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().Ship(999, 1, &dto.ShipOrderRequest{LogisticsCompany: "x", LogisticsNo: "y"})
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_Ship_StatusInvalid(t *testing.T) {
	mk := newOrderMocks()
	for _, st := range []int{
		model.OrderStatusPending,
		model.OrderStatusShipped,
		model.OrderStatusReceived,
		model.OrderStatusCompleted,
	} {
		mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: st})
		err := mk.svc().Ship(1, 1, &dto.ShipOrderRequest{LogisticsCompany: "x", LogisticsNo: "y"})
		assert.ErrorIs(t, err, ErrOrderStatusInvalid, "status=%d 应拒绝发货", st)
	}
}

func TestOrder_Ship_NoSellerRemark(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	err := mk.svc().Ship(1, 1, &dto.ShipOrderRequest{LogisticsCompany: "中通", LogisticsNo: "ZT1"})
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	_, hasRemark := fields["seller_remark"]
	assert.False(t, hasRemark)
}

// =====================================================================
// ===== Confirm 测试 =====
// =====================================================================

func TestOrder_Confirm_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusShipped})
	err := mk.svc().Confirm(1, 100)
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	require.NotNil(t, fields)
	assert.Equal(t, model.OrderStatusReceived, fields["status"])
	assert.NotNil(t, fields["received_at"])
	assert.NotNil(t, fields["auto_review_at"])
}

func TestOrder_Confirm_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().Confirm(999, 100)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_Confirm_NotOwner(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusShipped})
	err := mk.svc().Confirm(1, 200)
	assert.ErrorIs(t, err, ErrOrderNotOwner)
}

func TestOrder_Confirm_StatusInvalid(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	err := mk.svc().Confirm(1, 100)
	assert.ErrorIs(t, err, ErrOrderStatusInvalid)
}

// =====================================================================
// ===== Complete 测试 =====
// =====================================================================

func TestOrder_Complete_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusReceived})
	err := mk.svc().Complete(1)
	require.NoError(t, err)
	fields := mk.order.updatedFields[1]
	require.NotNil(t, fields)
	assert.Equal(t, model.OrderStatusCompleted, fields["status"])
	assert.NotNil(t, fields["completed_at"])
}

func TestOrder_Complete_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().Complete(999)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_Complete_StatusInvalid(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusShipped})
	err := mk.svc().Complete(1)
	assert.ErrorIs(t, err, ErrOrderStatusInvalid)
}

// =====================================================================
// ===== Delete 测试 =====
// =====================================================================

func TestOrder_Delete_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusClosed})
	err := mk.svc().Delete(1)
	require.NoError(t, err)
	assert.True(t, mk.order.deletedIDs[1])
}

func TestOrder_Delete_NotFound(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().Delete(999)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_Delete_RepoDeleteError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{})
	mk.order.deleteErr = errors.New("delete failed")
	err := mk.svc().Delete(1)
	assert.Error(t, err)
}

// =====================================================================
// ===== GetByID / GetByOrderNo 测试 =====
// =====================================================================

func TestOrder_GetByID_SuccessWithItems(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	mk.item.byOrder[1] = []model.OrderItem{
		{OrderID: 1, ProductName: "商品A", Price: 10, Quantity: 2},
	}
	info, err := mk.svc().GetByID(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), info.ID)
	assert.Equal(t, "已付款", info.StatusText)
	require.Len(t, info.Items, 1)
	assert.Equal(t, "商品A", info.Items[0].ProductName)
}

func TestOrder_GetByID_NotFound(t *testing.T) {
	mk := newOrderMocks()
	_, err := mk.svc().GetByID(999)
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

func TestOrder_GetByID_ItemListErrorTolerated(t *testing.T) {
	mk := newOrderMocks()
	mk.order.byID[1] = newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPaid})
	mk.item.listByOrdErr = errors.New("item err")
	info, err := mk.svc().GetByID(1)
	// 明细加载失败不应导致整体失败
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Empty(t, info.Items)
}

func TestOrder_GetByOrderNo_Success(t *testing.T) {
	mk := newOrderMocks()
	o := newOrder(1, 1, 100, 200, model.Order{OrderNo: "MAL20260101", Status: model.OrderStatusShipped})
	mk.order.byID[1] = o
	mk.order.byOrderNo["MAL20260101"] = o
	mk.item.byOrder[1] = []model.OrderItem{{OrderID: 1}}
	info, err := mk.svc().GetByOrderNo("MAL20260101")
	require.NoError(t, err)
	assert.Equal(t, "MAL20260101", info.OrderNo)
	assert.Equal(t, "已发货", info.StatusText)
	require.Len(t, info.Items, 1)
}

func TestOrder_GetByOrderNo_NotFound(t *testing.T) {
	mk := newOrderMocks()
	_, err := mk.svc().GetByOrderNo("NOPE")
	assert.ErrorIs(t, err, ErrOrderNotFound)
}

// =====================================================================
// ===== ListByUser / ListByShop / AdminList 测试 =====
// =====================================================================

func TestOrder_ListByUser_Success(t *testing.T) {
	mk := newOrderMocks()
	st := model.OrderStatusPaid
	mk.order.listByUserRet = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{OrderNo: "O1"}),
		*newOrder(2, 1, 100, 200, model.Order{OrderNo: "O2"}),
	}
	mk.order.listTotal = 2
	p, list, err := mk.svc().ListByUser(100, &dto.OrderListRequest{Status: &st})
	require.NoError(t, err)
	assert.Equal(t, int64(2), p.Total)
	require.Len(t, list, 2)
	assert.Equal(t, "O1", list[0].OrderNo)
}

func TestOrder_ListByUser_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.listErr = errors.New("list err")
	_, _, err := mk.svc().ListByUser(100, &dto.OrderListRequest{})
	assert.Error(t, err)
}

func TestOrder_ListByShop_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.listByShopRet = []model.Order{*newOrder(1, 1, 100, 200, model.Order{OrderNo: "S1"})}
	mk.order.listTotal = 1
	p, list, err := mk.svc().ListByShop(200, &dto.OrderListRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.Total)
	require.Len(t, list, 1)
}

func TestOrder_ListByShop_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.listErr = errors.New("list err")
	_, _, err := mk.svc().ListByShop(200, &dto.OrderListRequest{})
	assert.Error(t, err)
}

func TestOrder_AdminList_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.adminListRet = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{OrderNo: "A1"}),
		*newOrder(2, 1, 101, 201, model.Order{OrderNo: "A2"}),
		*newOrder(3, 1, 102, 202, model.Order{OrderNo: "A3"}),
	}
	mk.order.listTotal = 3
	p, list, err := mk.svc().AdminList(&dto.AdminOrderListRequest{RegionID: 1})
	require.NoError(t, err)
	assert.Equal(t, int64(3), p.Total)
	require.Len(t, list, 3)
}

func TestOrder_AdminList_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.listErr = errors.New("admin list err")
	_, _, err := mk.svc().AdminList(&dto.AdminOrderListRequest{})
	assert.Error(t, err)
}

// =====================================================================
// ===== BatchUpdateStatus / CountByStatus 测试 =====
// =====================================================================

func TestOrder_BatchUpdateStatus_Success(t *testing.T) {
	mk := newOrderMocks()
	err := mk.svc().BatchUpdateStatus([]uint{1, 2, 3}, model.OrderStatusClosed)
	require.NoError(t, err)
	assert.Equal(t, []uint{1, 2, 3}, mk.order.batchCallIDs)
	assert.Equal(t, model.OrderStatusClosed, mk.order.batchCallStat)
}

func TestOrder_BatchUpdateStatus_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.batchErr = errors.New("batch err")
	err := mk.svc().BatchUpdateStatus([]uint{1}, model.OrderStatusClosed)
	assert.Error(t, err)
}

func TestOrder_CountByStatus_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.listTotal = 42
	cnt, err := mk.svc().CountByStatus(100, model.OrderStatusPending)
	require.NoError(t, err)
	assert.Equal(t, int64(42), cnt)
	assert.Equal(t, model.OrderStatusPending, mk.order.countCallStat)
}

func TestOrder_CountByStatus_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.countErr = errors.New("count err")
	_, err := mk.svc().CountByStatus(100, model.OrderStatusPending)
	assert.Error(t, err)
}

// =====================================================================
// ===== Summary 测试 =====
// =====================================================================

func TestOrder_Summary_SuccessNoDates(t *testing.T) {
	mk := newOrderMocks()
	mk.order.summaryResult = &repository.OrderSummaryResult{
		TotalCount:     10,
		TotalAmount:    1000.5,
		PaidCount:      8,
		PaidAmount:     800.0,
		PendingCount:   1,
		ShippedCount:   1,
		CompletedCount: 5,
		CancelledCount: 1,
		RefundedCount:  1,
		RefundedAmount: 50.0,
	}
	sum, err := mk.svc().Summary(1, 100, 200, "", "")
	require.NoError(t, err)
	assert.Equal(t, int64(10), sum.TotalCount)
	assert.Equal(t, 1000.5, sum.TotalAmount)
	assert.Equal(t, int64(8), sum.PaidCount)
	assert.Equal(t, 800.0, sum.PaidAmount)
	assert.Equal(t, int64(1), sum.PendingCount)
	assert.Equal(t, int64(5), sum.CompletedCount)
	assert.Equal(t, 50.0, sum.RefundedAmount)
}

func TestOrder_Summary_WithValidDates(t *testing.T) {
	mk := newOrderMocks()
	_, err := mk.svc().Summary(1, 100, 200, "2026-01-01", "2026-01-31")
	require.NoError(t, err)
	// 验证仓储收到了带日期的 opts（通过 summaryResult 非空判断调用成功）
	// 这里主要验证日期解析不报错
}

func TestOrder_Summary_WithBadDates(t *testing.T) {
	mk := newOrderMocks()
	// 非法日期格式应被忽略，不应报错
	_, err := mk.svc().Summary(1, 100, 200, "not-a-date", "2026/01/01")
	require.NoError(t, err)
}

func TestOrder_Summary_Error(t *testing.T) {
	mk := newOrderMocks()
	mk.order.summaryErr = errors.New("summary err")
	_, err := mk.svc().Summary(1, 100, 200, "", "")
	assert.Error(t, err)
}

// =====================================================================
// ===== AutoClose / AutoConfirm / AutoReview 测试 =====
// =====================================================================

func TestOrder_AutoClose_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoCloseList = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusPending}),
		*newOrder(2, 1, 101, 201, model.Order{Status: model.OrderStatusPending}),
	}
	cnt, err := mk.svc().AutoClose()
	require.NoError(t, err)
	assert.Equal(t, 2, cnt)
	// 两个订单都被更新为 Closed
	assert.Equal(t, model.OrderStatusClosed, mk.order.updatedFields[1]["status"])
	assert.Equal(t, model.OrderStatusClosed, mk.order.updatedFields[2]["status"])
	assert.NotNil(t, mk.order.updatedFields[1]["cancelled_at"])
}

func TestOrder_AutoClose_Empty(t *testing.T) {
	mk := newOrderMocks()
	cnt, err := mk.svc().AutoClose()
	require.NoError(t, err)
	assert.Equal(t, 0, cnt)
}

func TestOrder_AutoClose_ListError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoListErr = errors.New("auto close list err")
	_, err := mk.svc().AutoClose()
	assert.Error(t, err)
}

func TestOrder_AutoClose_PartialUpdateError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoCloseList = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{}),
		*newOrder(2, 1, 101, 201, model.Order{}),
	}
	// 让第一次 UpdateFields 成功、第二次失败难以精确控制；
	// 这里通过 updateErr 全局失败验证：失败的不计入 cnt，但整体不报错
	mk.order.updateErr = errors.New("update err")
	cnt, err := mk.svc().AutoClose()
	require.NoError(t, err)
	assert.Equal(t, 0, cnt)
}

func TestOrder_AutoConfirm_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoConfirmList = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusShipped}),
	}
	cnt, err := mk.svc().AutoConfirm()
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)
	assert.Equal(t, model.OrderStatusReceived, mk.order.updatedFields[1]["status"])
	assert.NotNil(t, mk.order.updatedFields[1]["received_at"])
	assert.NotNil(t, mk.order.updatedFields[1]["auto_review_at"])
}

func TestOrder_AutoConfirm_Empty(t *testing.T) {
	mk := newOrderMocks()
	cnt, err := mk.svc().AutoConfirm()
	require.NoError(t, err)
	assert.Equal(t, 0, cnt)
}

func TestOrder_AutoConfirm_ListError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoListErr = errors.New("auto confirm list err")
	_, err := mk.svc().AutoConfirm()
	assert.Error(t, err)
}

func TestOrder_AutoReview_Success(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoReviewList = []model.Order{
		*newOrder(1, 1, 100, 200, model.Order{Status: model.OrderStatusReceived}),
		*newOrder(2, 1, 101, 201, model.Order{Status: model.OrderStatusReceived}),
		*newOrder(3, 1, 102, 202, model.Order{Status: model.OrderStatusReceived}),
	}
	cnt, err := mk.svc().AutoReview()
	require.NoError(t, err)
	assert.Equal(t, 3, cnt)
	for i := uint(1); i <= 3; i++ {
		assert.Equal(t, model.OrderStatusCompleted, mk.order.updatedFields[i]["status"])
		assert.Equal(t, true, mk.order.updatedFields[i]["has_review"])
		assert.NotNil(t, mk.order.updatedFields[i]["completed_at"])
	}
}

func TestOrder_AutoReview_Empty(t *testing.T) {
	mk := newOrderMocks()
	cnt, err := mk.svc().AutoReview()
	require.NoError(t, err)
	assert.Equal(t, 0, cnt)
}

func TestOrder_AutoReview_ListError(t *testing.T) {
	mk := newOrderMocks()
	mk.order.autoListErr = errors.New("auto review list err")
	_, err := mk.svc().AutoReview()
	assert.Error(t, err)
}

// =====================================================================
// ===== 边界：状态文本映射 =====
// =====================================================================

func TestOrder_StatusText_AllMappings(t *testing.T) {
	cases := []struct {
		status int
		text   string
	}{
		{model.OrderStatusPending, "待付款"},
		{model.OrderStatusPaid, "已付款"},
		{model.OrderStatusShipped, "已发货"},
		{model.OrderStatusReceived, "已收货"},
		{model.OrderStatusCompleted, "已完成"},
		{model.OrderStatusCancelled, "已取消"},
		{model.OrderStatusRefunded, "已退款"},
		{model.OrderStatusClosed, "已关闭"},
		{99, ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.text, orderStatusText(c.status), "status=%d", c.status)
	}
}

func TestOrder_GenerateOrderNo_Format(t *testing.T) {
	no := generateOrderNo()
	assert.Contains(t, no, "MAL")
	// MAL + 14位时间 + 6位随机 = 3 + 14 + 6 = 23
	assert.Equal(t, 23, len(no))
}
