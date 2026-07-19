// Package service 订单业务逻辑层（11 状态机）
// 依据 v3.2.1 架构方案：状态机 unpaid → paid → shipped → received → completed / cancelled
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound       = errors.New("订单不存在")
	ErrOrderNoPermission   = errors.New("无权操作此订单")
	ErrOrderStatusInvalid  = errors.New("订单状态不允许此操作")
	ErrOrderItemNotFound   = errors.New("订单商品不存在")
	ErrOrderItemOffShelf   = errors.New("商品已下架")
	ErrOrderItemSelfBuy    = errors.New("不能购买自己的商品")
)

// OrderService 订单业务接口
type OrderService interface {
	Create(regionID, buyerID uint, req *dto.OrderCreateRequest) (*dto.OrderResponse, error)
	GetByID(orderID, userID uint) (*dto.OrderResponse, error)
	List(userID uint, query dto.OrderQuery) (*utils.Pagination, []dto.OrderResponse, error)
	Pay(orderID, userID uint) (*dto.OrderResponse, error)
	Ship(orderID, userID uint) (*dto.OrderResponse, error)
	Receive(orderID, userID uint) (*dto.OrderResponse, error)
	Cancel(orderID, userID uint, remark string) (*dto.OrderResponse, error)
	Complete(orderID, userID uint) (*dto.OrderResponse, error)
	UpdateStatus(orderID, userID uint, req *dto.OrderStatusUpdateRequest) (*dto.OrderResponse, error)
}

type orderService struct {
	repo      repository.OrderRepository
	ershouRepo repository.ErshouRepository
	skuRepo   repository.SKURepository
	escrowRepo repository.EscrowRepository
}

// NewOrderService 创建订单 service 实例
func NewOrderService(
	repo repository.OrderRepository,
	ershouRepo repository.ErshouRepository,
	skuRepo repository.SKURepository,
	escrowRepo repository.EscrowRepository,
) OrderService {
	return &orderService{
		repo:       repo,
		ershouRepo: ershouRepo,
		skuRepo:    skuRepo,
		escrowRepo: escrowRepo,
	}
}

// 订单状态文案
func orderStatusText(status int) string {
	switch status {
	case model.OrderStatusPending:
		return "待支付"
	case model.OrderStatusPaid:
		return "已支付待发货"
	case model.OrderStatusShipped:
		return "已发货"
	case model.OrderStatusDelivered:
		return "待收货"
	case model.OrderStatusCompleted:
		return "已完成"
	case model.OrderStatusCanceled:
		return "已取消"
	case model.OrderStatusRefunding:
		return "退款中"
	case model.OrderStatusRefunded:
		return "退款完成"
	case model.OrderStatusDispute:
		return "申诉中"
	case model.OrderStatusDisputeClosed:
		return "申诉完成"
	case model.OrderStatusClosed:
		return "已关闭"
	}
	return "未知"
}

// genOrderNo 生成订单号：ES + yyyyMMddHHmmss + 6位随机数
func genOrderNo() string {
	return fmt.Sprintf("ES%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

func (s *orderService) Create(regionID, buyerID uint, req *dto.OrderCreateRequest) (*dto.OrderResponse, error) {
	if buyerID == 0 {
		return nil, ErrOrderNoPermission
	}

	items := make([]model.ErshouOrderItem, 0, len(req.Items))
	var itemAmount, deliveryFee float64
	sellerID := uint(0)
	shopID := req.ShopID
	ershouIDs := make([]uint, 0, len(req.Items))

	for _, it := range req.Items {
		e, err := s.ershouRepo.FindByID(it.ErshouID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrOrderItemNotFound
			}
			return nil, err
		}
		if e.Status != model.StatusPublished {
			return nil, ErrOrderItemOffShelf
		}
		if e.UserID == buyerID {
			return nil, ErrOrderItemSelfBuy
		}
		if sellerID == 0 {
			sellerID = e.UserID
		} else if sellerID != e.UserID {
			return nil, errors.New("同一订单只能购买同一卖家的商品")
		}
		if shopID == 0 {
			shopID = e.ShopID
		}

		unitPrice := e.Price
		skuCode := ""
		if it.SKUID > 0 {
			sku, err := s.skuRepo.FindByID(it.SKUID)
			if err != nil {
				return nil, ErrSKUNotFound
			}
			if sku.ErshouID != it.ErshouID {
				return nil, ErrSKUNotFound
			}
			if sku.Stock < it.Quantity {
				return nil, ErrSKUStockInsufficient
			}
			unitPrice = sku.Price
			skuCode = sku.SKUCode
		}
		subtotal := unitPrice * float64(it.Quantity)
		itemAmount += subtotal

		items = append(items, model.ErshouOrderItem{
			ErshouID:   it.ErshouID,
			SKUID:      it.SKUID,
			SKUCode:    skuCode,
			Title:      e.Title,
			CoverImage: e.CoverImage,
			Quantity:   it.Quantity,
			UnitPrice:  unitPrice,
			Subtotal:   subtotal,
			Remark:     it.Remark,
		})
		ershouIDs = append(ershouIDs, it.ErshouID)
	}

	payMethod := req.PayMethod
	if payMethod == "" {
		payMethod = model.PayMethodWechat
	}
	deliveryMethod := req.DeliveryMethod
	if deliveryMethod == "" {
		deliveryMethod = model.OrderDeliveryFace
	}

	totalAmount := itemAmount + deliveryFee
	autoCloseAt := time.Now().Add(24 * time.Hour) // 24h 未付款自动关闭

	order := &model.ErshouOrder{
		OrderNo:           genOrderNo(),
		BuyerID:           buyerID,
		SellerID:          sellerID,
		ShopID:            shopID,
		TotalAmount:       totalAmount,
		ItemAmount:        itemAmount,
		DeliveryFee:       deliveryFee,
		Status:            model.OrderStatusPending,
		PayMethod:         payMethod,
		DeliveryMethod:    deliveryMethod,
		Remark:            req.Remark,
		ContactName:       req.ContactName,
		ContactPhone:      req.ContactPhone,
		ContactAddress:    req.ContactAddress,
		EscrowEnabled:     req.EscrowEnabled,
		InstallmentEnabled: req.InstallmentEnabled,
		InstallmentPeriods: req.InstallmentPeriods,
		AutoCloseAt:       &autoCloseAt,
	}
	order.RegionID = regionID

	if err := s.repo.Create(order, items); err != nil {
		return nil, err
	}

	// 占用 SKU 库存（预扣）
	for _, it := range req.Items {
		if it.SKUID > 0 {
			_ = s.skuRepo.DecrStock(it.SKUID, it.Quantity)
		}
	}

	return s.toOrderResponse(order, items), nil
}

func (s *orderService) GetByID(orderID, userID uint) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if userID > 0 && order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	items, _ := s.repo.ListItems(orderID)
	return s.toOrderResponse(order, items), nil
}

func (s *orderService) List(userID uint, query dto.OrderQuery) (*utils.Pagination, []dto.OrderResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	if query.Role == "" {
		query.Role = "all"
	}
	list, total, err := s.repo.List(repository.OrderListQuery{
		UserID:  userID,
		Role:    query.Role,
		Status:  query.Status,
		OrderNo: query.OrderNo,
		Keyword: query.Keyword,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.OrderResponse, 0, len(list))
	for i := range list {
		items, _ := s.repo.ListItems(list[i].ID)
		result = append(result, *s.toOrderResponse(&list[i], items))
	}
	return pagination, result, nil
}

// Pay 支付订单
func (s *orderService) Pay(orderID, userID uint) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID {
		return nil, ErrOrderNoPermission
	}
	if order.Status != model.OrderStatusPending {
		return nil, ErrOrderStatusInvalid
	}

	now := time.Now()
	autoReceiveAt := now.Add(7 * 24 * time.Hour) // 7 天自动收货
	fields := map[string]interface{}{
		"status":         model.OrderStatusPaid,
		"paid_at":        &now,
		"auto_close_at":  nil,
		"auto_receive_at": &autoReceiveAt,
		"pay_trade_no":   fmt.Sprintf("PAY%s%d", now.Format("20060102150405"), orderID),
	}
	if err := s.repo.Update(orderID, fields); err != nil {
		return nil, err
	}

	// 启用担保交易：创建担保记录（冻结资金）
	if order.EscrowEnabled {
		escrow := &model.ErshouEscrow{
			OrderID:      orderID,
			ErshouID:     0,
			BuyerID:      order.BuyerID,
			SellerID:     order.SellerID,
			EscrowAmount: order.TotalAmount,
			Status:       model.EscrowStatusFrozen,
			FrozenAt:     &now,
			// 自动放款：T+1 (确认收货 1 天后)
			AutoReleaseAt: nil,
		}
		_ = s.escrowRepo.Create(escrow)
	}

	// 商品下架
	items, _ := s.repo.ListItems(orderID)
	for _, it := range items {
		_ = s.ershouRepo.UpdateFields(it.ErshouID, map[string]interface{}{
			"status": model.StatusSold,
		})
		if it.SKUID > 0 {
			_ = s.skuRepo.IncrSoldCount(it.SKUID, it.Quantity)
		}
	}

	order.Status = model.OrderStatusPaid
	order.PaidAt = &now
	return s.toOrderResponse(order, items), nil
}

// Ship 发货
func (s *orderService) Ship(orderID, userID uint) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	if order.Status != model.OrderStatusPaid {
		return nil, ErrOrderStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":      model.OrderStatusShipped,
		"shipped_at":  &now,
	}
	if err := s.repo.Update(orderID, fields); err != nil {
		return nil, err
	}
	order.Status = model.OrderStatusShipped
	order.ShippedAt = &now
	items, _ := s.repo.ListItems(orderID)
	return s.toOrderResponse(order, items), nil
}

// Receive 确认收货
func (s *orderService) Receive(orderID, userID uint) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID {
		return nil, ErrOrderNoPermission
	}
	if order.Status != model.OrderStatusShipped && order.Status != model.OrderStatusDelivered {
		return nil, ErrOrderStatusInvalid
	}
	now := time.Now()
	// 确认收货 = 完成订单（合并状态）
	fields := map[string]interface{}{
		"status":       model.OrderStatusCompleted,
		"received_at":  &now,
		"settled_at":   &now,
	}
	if err := s.repo.Update(orderID, fields); err != nil {
		return nil, err
	}

	// 担保放款
	if order.EscrowEnabled {
		if esc, err := s.escrowRepo.FindByOrderID(orderID); err == nil && esc != nil {
			now := now
			_ = s.escrowRepo.Update(esc.ID, map[string]interface{}{
				"status":      model.EscrowStatusReleased,
				"release_at":  &now,
				// 自动放款时间 = T+1
				"auto_release_at": now.Add(24 * time.Hour),
			})
		}
	}

	order.Status = model.OrderStatusCompleted
	order.ReceivedAt = &now
	items, _ := s.repo.ListItems(orderID)
	return s.toOrderResponse(order, items), nil
}

// Cancel 取消订单（仅买家在待支付状态可取消）
func (s *orderService) Cancel(orderID, userID uint, remark string) (*dto.OrderResponse, error) {
	order, err := s.repo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusPaid {
		return nil, ErrOrderStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":    model.OrderStatusCanceled,
		"closed_at": &now,
	}
	if remark != "" {
		fields["remark"] = remark
	}
	if err := s.repo.Update(orderID, fields); err != nil {
		return nil, err
	}

	// 恢复商品状态为已发布
	items, _ := s.repo.ListItems(orderID)
	for _, it := range items {
		_ = s.ershouRepo.UpdateFields(it.ErshouID, map[string]interface{}{
			"status": model.StatusPublished,
		})
	}

	order.Status = model.OrderStatusCanceled
	return s.toOrderResponse(order, items), nil
}

// Complete 完成订单
func (s *orderService) Complete(orderID, userID uint) (*dto.OrderResponse, error) {
	return s.Receive(orderID, userID)
}

// UpdateStatus 通用状态机变更入口
func (s *orderService) UpdateStatus(orderID, userID uint, req *dto.OrderStatusUpdateRequest) (*dto.OrderResponse, error) {
	switch req.Action {
	case "pay":
		return s.Pay(orderID, userID)
	case "ship":
		return s.Ship(orderID, userID)
	case "receive":
		return s.Receive(orderID, userID)
	case "cancel":
		return s.Cancel(orderID, userID, req.Remark)
	case "complete":
		return s.Complete(orderID, userID)
	}
	return nil, ErrOrderStatusInvalid
}

func (s *orderService) toOrderResponse(order *model.ErshouOrder, items []model.ErshouOrderItem) *dto.OrderResponse {
	resp := &dto.OrderResponse{
		ID:                 order.ID,
		OrderNo:            order.OrderNo,
		BuyerID:            order.BuyerID,
		SellerID:           order.SellerID,
		ShopID:             order.ShopID,
		TotalAmount:        order.TotalAmount,
		ItemAmount:         order.ItemAmount,
		DeliveryFee:        order.DeliveryFee,
		DiscountAmount:     order.DiscountAmount,
		Status:             order.Status,
		StatusText:         orderStatusText(order.Status),
		PayMethod:          order.PayMethod,
		PayTradeNo:         order.PayTradeNo,
		DeliveryMethod:     order.DeliveryMethod,
		Remark:             order.Remark,
		ContactName:        order.ContactName,
		ContactPhone:       order.ContactPhone,
		ContactAddress:     order.ContactAddress,
		EscrowEnabled:      order.EscrowEnabled,
		InstallmentEnabled: order.InstallmentEnabled,
		InstallmentPeriods: order.InstallmentPeriods,
		PaidAt:             order.PaidAt,
		ShippedAt:          order.ShippedAt,
		ReceivedAt:         order.ReceivedAt,
		SettledAt:          order.SettledAt,
		ClosedAt:           order.ClosedAt,
		AutoCloseAt:        order.AutoCloseAt,
		AutoReceiveAt:      order.AutoReceiveAt,
		RegionID:           order.RegionID,
		CreatedAt:          order.CreatedAt,
		UpdatedAt:          order.UpdatedAt,
		Items:              []dto.OrderItemResponse{},
	}
	for _, it := range items {
		resp.Items = append(resp.Items, dto.OrderItemResponse{
			ID:         it.ID,
			OrderID:    it.OrderID,
			ErshouID:   it.ErshouID,
			SKUID:      it.SKUID,
			SKUCode:    it.SKUCode,
			Title:      it.Title,
			CoverImage: it.CoverImage,
			Quantity:   it.Quantity,
			UnitPrice:  it.UnitPrice,
			Subtotal:   it.Subtotal,
			Remark:     it.Remark,
		})
	}
	return resp
}

// 辅助：将字符串 ID 转 uint
func parseUint(s string) uint {
	id, _ := strconv.ParseUint(s, 10, 32)
	return uint(id)
}
