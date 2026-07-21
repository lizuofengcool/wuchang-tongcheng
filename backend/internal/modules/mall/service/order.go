// Package service 同城商城业务逻辑层 - 订单
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound      = errors.New("订单不存在")
	ErrOrderNotOwner      = errors.New("无权操作他人订单")
	ErrOrderStatusInvalid = errors.New("订单状态不允许此操作")
	ErrOrderStockShort    = errors.New("商品库存不足")
)

// 订单自动关闭/确认/评价时长
const (
	orderAutoCloseMinutes  = 15  // 待付款 15 分钟自动关闭
	orderAutoConfirmHours  = 72  // 发货后 72 小时自动确认收货
	orderAutoReviewHours   = 168 // 收货后 168 小时（7天）自动评价
)

// OrderService 订单业务接口
type OrderService interface {
	Create(regionID, userID uint, buyerName, buyerPhone, buyerAvatar, clientIP, userAgent string, req *dto.CreateOrderRequest) (*dto.OrderInfo, error)
	Cancel(id, userID uint, req *dto.CancelOrderRequest) error
	AdminClose(id uint, req *dto.AdminCloseOrderRequest) error
	Ship(id, shipperID uint, req *dto.ShipOrderRequest) error
	Confirm(id, userID uint) error
	Complete(id uint) error
	Delete(id uint) error

	GetByID(id uint) (*dto.OrderInfo, error)
	GetByOrderNo(orderNo string) (*dto.OrderInfo, error)
	ListByUser(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error)
	ListByShop(shopID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error)
	AdminList(req *dto.AdminOrderListRequest) (*utils.Pagination, []dto.OrderInfo, error)

	BatchUpdateStatus(ids []uint, status int) error
	CountByStatus(userID uint, status int) (int64, error)
	Summary(regionID, userID, shopID uint, startDate, endDate string) (*dto.OrderSummary, error)

	// 定时任务
	AutoClose() (int, error)
	AutoConfirm() (int, error)
	AutoReview() (int, error)
}

type orderService struct {
	repo         repository.OrderRepository
	itemRepo     repository.OrderItemRepository
	addressRepo  repository.AddressRepository
	productRepo  repository.ProductRepository
	skuRepo      repository.SkuRepository
	shopRepo     repository.ShopRepository
	cartRepo     repository.CartRepository
}

// NewOrderService 创建订单 service 实例
func NewOrderService(
	repo repository.OrderRepository,
	itemRepo repository.OrderItemRepository,
	addressRepo repository.AddressRepository,
	productRepo repository.ProductRepository,
	skuRepo repository.SkuRepository,
	shopRepo repository.ShopRepository,
	cartRepo repository.CartRepository,
) OrderService {
	return &orderService{
		repo:        repo,
		itemRepo:    itemRepo,
		addressRepo: addressRepo,
		productRepo: productRepo,
		skuRepo:     skuRepo,
		shopRepo:    shopRepo,
		cartRepo:    cartRepo,
	}
}

// orderStatusText 订单状态文本
func orderStatusText(s int) string {
	switch s {
	case model.OrderStatusPending:
		return "待付款"
	case model.OrderStatusPaid:
		return "已付款"
	case model.OrderStatusShipped:
		return "已发货"
	case model.OrderStatusReceived:
		return "已收货"
	case model.OrderStatusCompleted:
		return "已完成"
	case model.OrderStatusCancelled:
		return "已取消"
	case model.OrderStatusRefunded:
		return "已退款"
	case model.OrderStatusClosed:
		return "已关闭"
	}
	return ""
}

// toOrderInfo model -> dto（不含 items）
func toOrderInfo(o *model.Order) *dto.OrderInfo {
	return &dto.OrderInfo{
		ID:              o.ID,
		OrderNo:         o.OrderNo,
		UserID:          o.UserID,
		ShopID:          o.ShopID,
		ShopName:        o.ShopName,
		BuyerName:       o.BuyerName,
		BuyerPhone:      o.BuyerPhone,
		BuyerAvatar:     o.BuyerAvatar,
		AddressID:       o.AddressID,
		ReceiverName:    o.ReceiverName,
		ReceiverPhone:   o.ReceiverPhone,
		Province:        o.Province,
		City:            o.City,
		District:        o.District,
		Address:         o.Address,
		ZipCode:         o.ZipCode,
		TotalAmount:     o.TotalAmount,
		ShippingFee:     o.ShippingFee,
		DiscountAmount:  o.DiscountAmount,
		TaxAmount:       o.TaxAmount,
		PayAmount:       o.PayAmount,
		Status:          o.Status,
		StatusText:      orderStatusText(o.Status),
		Remark:          o.Remark,
		SellerRemark:    o.SellerRemark,
		AdminRemark:     o.AdminRemark,
		PaidAt:          o.PaidAt,
		ShippedAt:       o.ShippedAt,
		ReceivedAt:      o.ReceivedAt,
		CompletedAt:     o.CompletedAt,
		CancelledAt:     o.CancelledAt,
		AutoCloseAt:     o.AutoCloseAt,
		AutoConfirmAt:   o.AutoConfirmAt,
		AutoReviewAt:    o.AutoReviewAt,
		PaymentMethod:   o.PaymentMethod,
		PaymentNo:       o.PaymentNo,
		Source:          o.Source,
		ClientIP:        o.ClientIP,
		UserAgent:       o.UserAgent,
		CouponID:        o.CouponID,
		CouponName:      o.CouponName,
		HasReview:       o.HasReview,
		HasSellerReview: o.HasSellerReview,
		LogisticsCompany: o.LogisticsCompany,
		LogisticsNo:     o.LogisticsNo,
		RiskScore:       o.RiskScore,
		RegionID:        o.RegionID,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

// toOrderItemInfo model.OrderItem -> dto.OrderItemInfo
func toOrderItemInfo(it *model.OrderItem) dto.OrderItemInfo {
	return dto.OrderItemInfo{
		ID:              it.ID,
		OrderID:         it.OrderID,
		ProductID:       it.ProductID,
		SkuID:           it.SkuID,
		ProductName:     it.ProductName,
		MainImage:       it.MainImage,
		SkuName:         it.SkuName,
		SkuSpecs:        it.SkuSpecs,
		SkuCode:         it.SkuCode,
		Price:           it.Price,
		Quantity:        it.Quantity,
		TotalAmount:     it.TotalAmount,
		DiscountAmount:  it.DiscountAmount,
		ShippingFee:     it.ShippingFee,
		PayAmount:       it.PayAmount,
		HasReview:       it.HasReview,
		ReviewID:        it.ReviewID,
		RefundStatus:    it.RefundStatus,
		RefundID:        it.RefundID,
		Status:          it.Status,
		StatusText:      orderItemStatusText(it.Status),
	}
}

// generateOrderNo 生成订单号
func generateOrderNo() string {
	return fmt.Sprintf("MAL%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建订单
func (s *orderService) Create(regionID, userID uint, buyerName, buyerPhone, buyerAvatar, clientIP, userAgent string, req *dto.CreateOrderRequest) (*dto.OrderInfo, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("订单商品不能为空")
	}

	// 校验收货地址
	addr, err := s.addressRepo.FindByID(req.AddressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("收货地址不存在")
		}
		return nil, err
	}
	if addr.UserID != userID {
		return nil, errors.New("无权使用他人收货地址")
	}

	// 校验商品并构造订单明细
	items := make([]model.OrderItem, 0, len(req.Items))
	var totalAmount float64
	var shopID uint
	for i, reqItem := range req.Items {
		// 查商品
		p, err := s.productRepo.FindByID(reqItem.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, err
		}
		if p.Status != model.ProductStatusOnSale {
			return nil, errors.New("商品已下架")
		}
		if i == 0 {
			shopID = p.ShopID
		} else if p.ShopID != shopID {
			return nil, errors.New("订单商品必须属于同一店铺")
		}

		// 取 SKU
		var sku *model.Sku
		skuName := ""
		skuSpecs := ""
		skuCode := ""
		price := p.Price
		if reqItem.SkuID > 0 {
			sku, err = s.skuRepo.FindByID(reqItem.SkuID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrSkuNotFound
				}
				return nil, err
			}
			if sku.ProductID != p.ID {
				return nil, errors.New("SKU 不属于该商品")
			}
			if sku.Stock < reqItem.Quantity {
				return nil, ErrOrderStockShort
			}
			skuName = sku.Name
			skuSpecs = string(sku.Specs)
			skuCode = sku.SkuCode
			price = sku.Price
		} else {
			// 不带 SKU：直接使用商品库存校验（暂以 sales 作为粗校验）
			if p.Stock < reqItem.Quantity {
				return nil, ErrOrderStockShort
			}
		}

		subtotal := price * float64(reqItem.Quantity)
		totalAmount += subtotal

		items = append(items, model.OrderItem{
			ProductID:    reqItem.ProductID,
			SkuID:        reqItem.SkuID,
			UserID:       userID,
			ShopID:       shopID,
			ProductName:  p.Name,
			MainImage:    p.MainImage,
			SkuName:      skuName,
			SkuSpecs:     skuSpecs,
			SkuCode:      skuCode,
			Price:        price,
			Quantity:     reqItem.Quantity,
			TotalAmount:  subtotal,
			PayAmount:    subtotal,
			Status:       model.OrderStatusPending,
		})
	}

	// 查店铺
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	now := time.Now()
	autoCloseAt := now.Add(time.Minute * time.Duration(orderAutoCloseMinutes))

	o := &model.Order{
		OrderNo:       generateOrderNo(),
		UserID:        userID,
		ShopID:        shopID,
		ShopName:      shop.ShopName,
		BuyerName:     buyerName,
		BuyerPhone:    buyerPhone,
		BuyerAvatar:   buyerAvatar,
		AddressID:     addr.ID,
		ReceiverName:  addr.Name,
		ReceiverPhone: addr.Phone,
		Province:      addr.Province,
		City:          addr.City,
		District:      addr.District,
		Address:       addr.Detail,
		ZipCode:       addr.ZipCode,
		TotalAmount:   totalAmount,
		PayAmount:     totalAmount,
		Status:        model.OrderStatusPending,
		StatusText:    orderStatusText(model.OrderStatusPending),
		Remark:        req.Remark,
		PaymentMethod: req.PaymentMethod,
		Source:        req.Source,
		ClientIP:      clientIP,
		UserAgent:     userAgent,
		CouponID:      req.CouponID,
		AutoCloseAt:   &autoCloseAt,
	}
	o.RegionID = regionID

	// 同步 order_no 到 items
	for i := range items {
		items[i].OrderNo = o.OrderNo
	}

	if err := s.repo.Create(o, items); err != nil {
		return nil, err
	}

	// 从购物车结算的，清空对应购物车项
	if req.FromCart {
		for _, it := range req.Items {
			if cartItem, err := s.cartRepo.FindByUserAndSku(userID, it.SkuID); err == nil && cartItem != nil {
				_ = s.cartRepo.Delete(cartItem.ID)
			}
		}
	}

	info := toOrderInfo(o)
	itemInfos := make([]dto.OrderItemInfo, 0, len(items))
	for i := range items {
		itemInfos = append(itemInfos, toOrderItemInfo(&items[i]))
	}
	info.Items = itemInfos
	return info, nil
}

// Cancel 买家取消订单
func (s *orderService) Cancel(id, userID uint, req *dto.CancelOrderRequest) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.UserID != userID {
		return ErrOrderNotOwner
	}
	if o.Status > model.OrderStatusPaid {
		return ErrOrderStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":       model.OrderStatusCancelled,
		"cancelled_at": &now,
	}
	if req.Reason != "" {
		fields["admin_remark"] = req.Reason
	}
	return s.repo.UpdateFields(id, fields)
}

// AdminClose 管理后台关闭订单
func (s *orderService) AdminClose(id uint, req *dto.AdminCloseOrderRequest) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.Status >= model.OrderStatusCompleted {
		return ErrOrderStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status": model.OrderStatusClosed,
		"cancelled_at": &now,
	}
	if req.AdminRemark != "" {
		fields["admin_remark"] = req.AdminRemark
	}
	return s.repo.UpdateFields(id, fields)
}

// Ship 卖家发货
func (s *orderService) Ship(id, shipperID uint, req *dto.ShipOrderRequest) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.Status != model.OrderStatusPaid {
		return ErrOrderStatusInvalid
	}

	now := time.Now()
	autoConfirmAt := now.Add(time.Hour * time.Duration(orderAutoConfirmHours))
	fields := map[string]interface{}{
		"status":           model.OrderStatusShipped,
		"shipped_at":       &now,
		"auto_confirm_at":  &autoConfirmAt,
		"logistics_company": req.LogisticsCompany,
		"logistics_no":      req.LogisticsNo,
	}
	if req.SellerRemark != "" {
		fields["seller_remark"] = req.SellerRemark
	}
	_ = shipperID
	return s.repo.UpdateFields(id, fields)
}

// Confirm 买家确认收货
func (s *orderService) Confirm(id, userID uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.UserID != userID {
		return ErrOrderNotOwner
	}
	if o.Status != model.OrderStatusShipped {
		return ErrOrderStatusInvalid
	}

	now := time.Now()
	autoReviewAt := now.Add(time.Hour * time.Duration(orderAutoReviewHours))
	fields := map[string]interface{}{
		"status":          model.OrderStatusReceived,
		"received_at":     &now,
		"auto_review_at":  &autoReviewAt,
	}
	return s.repo.UpdateFields(id, fields)
}

// Complete 完成订单（评价后或自动完成）
func (s *orderService) Complete(id uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	if o.Status != model.OrderStatusReceived {
		return ErrOrderStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.OrderStatusCompleted,
		"completed_at":  &now,
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除订单
func (s *orderService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取订单详情（含 items）
func (s *orderService) GetByID(id uint) (*dto.OrderInfo, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	info := toOrderInfo(o)
	// 加载明细
	items, err := s.itemRepo.ListByOrder(o.ID)
	if err == nil {
		itemInfos := make([]dto.OrderItemInfo, 0, len(items))
		for i := range items {
			itemInfos = append(itemInfos, toOrderItemInfo(&items[i]))
		}
		info.Items = itemInfos
	}
	return info, nil
}

// GetByOrderNo 按订单号查询
func (s *orderService) GetByOrderNo(orderNo string) (*dto.OrderInfo, error) {
	o, err := s.repo.FindByOrderNo(orderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	info := toOrderInfo(o)
	items, err := s.itemRepo.ListByOrder(o.ID)
	if err == nil {
		itemInfos := make([]dto.OrderItemInfo, 0, len(items))
		for i := range items {
			itemInfos = append(itemInfos, toOrderItemInfo(&items[i]))
		}
		info.Items = itemInfos
	}
	return info, nil
}

// ListByUser 按买家列出订单
func (s *orderService) ListByUser(userID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.OrderListOptions{
		Status:    req.Status,
		ShopID:    req.ShopID,
		OrderNo:   req.OrderNo,
		Keyword:   req.Keyword,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Buyer:     req.Buyer,
	}
	list, total, err := s.repo.ListByUser(userID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.OrderInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByShop 按店铺列出订单
func (s *orderService) ListByShop(shopID uint, req *dto.OrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.OrderListOptions{
		Status:    req.Status,
		OrderNo:   req.OrderNo,
		Keyword:   req.Keyword,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Buyer:     req.Buyer,
	}
	list, total, err := s.repo.ListByShop(shopID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.OrderInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminList 管理后台订单列表
func (s *orderService) AdminList(req *dto.AdminOrderListRequest) (*utils.Pagination, []dto.OrderInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AdminOrderListOptions{
		RegionID:       req.RegionID,
		UserID:         req.UserID,
		ShopID:         req.ShopID,
		Status:         req.Status,
		OrderNo:        req.OrderNo,
		PaymentMethod:  req.PaymentMethod,
		Keyword:        req.Keyword,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.OrderInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// BatchUpdateStatus 批量更新订单状态
func (s *orderService) BatchUpdateStatus(ids []uint, status int) error {
	return s.repo.BatchUpdateStatus(ids, status)
}

// CountByStatus 按状态统计订单数
func (s *orderService) CountByStatus(userID uint, status int) (int64, error) {
	return s.repo.CountByStatus(userID, status)
}

// Summary 订单汇总
func (s *orderService) Summary(regionID, userID, shopID uint, startDate, endDate string) (*dto.OrderSummary, error) {
	opts := repository.OrderSummaryOptions{
		RegionID: regionID,
		UserID:   userID,
		ShopID:   shopID,
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			opts.StartDate = t
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			opts.EndDate = t.Add(24 * time.Hour)
		}
	}
	result, err := s.repo.Summary(opts)
	if err != nil {
		return nil, err
	}
	return &dto.OrderSummary{
		TotalCount:     result.TotalCount,
		TotalAmount:    result.TotalAmount,
		PaidCount:      result.PaidCount,
		PaidAmount:     result.PaidAmount,
		PendingCount:   result.PendingCount,
		ShippedCount:   result.ShippedCount,
		CompletedCount: result.CompletedCount,
		CancelledCount: result.CancelledCount,
		RefundedCount:  result.RefundedCount,
		RefundedAmount: result.RefundedAmount,
	}, nil
}

// AutoClose 自动关闭超时未付款订单
func (s *orderService) AutoClose() (int, error) {
	list, err := s.repo.ListAutoClose(utils.NewPagination(1, 100))
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range list {
		now := time.Now()
		fields := map[string]interface{}{
			"status":       model.OrderStatusClosed,
			"cancelled_at": &now,
		}
		if err := s.repo.UpdateFields(list[i].ID, fields); err == nil {
			count++
		}
	}
	return count, nil
}

// AutoConfirm 自动确认收货
func (s *orderService) AutoConfirm() (int, error) {
	list, err := s.repo.ListAutoConfirm(utils.NewPagination(1, 100))
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range list {
		now := time.Now()
		autoReviewAt := now.Add(time.Hour * time.Duration(orderAutoReviewHours))
		fields := map[string]interface{}{
			"status":         model.OrderStatusReceived,
			"received_at":    &now,
			"auto_review_at": &autoReviewAt,
		}
		if err := s.repo.UpdateFields(list[i].ID, fields); err == nil {
			count++
		}
	}
	return count, nil
}

// AutoReview 自动评价（默认好评）
func (s *orderService) AutoReview() (int, error) {
	list, err := s.repo.ListAutoReview(utils.NewPagination(1, 100))
	if err != nil {
		return 0, err
	}
	count := 0
	for i := range list {
		now := time.Now()
		fields := map[string]interface{}{
			"status":        model.OrderStatusCompleted,
			"completed_at":  &now,
			"has_review":    true,
		}
		if err := s.repo.UpdateFields(list[i].ID, fields); err == nil {
			count++
		}
	}
	return count, nil
}
