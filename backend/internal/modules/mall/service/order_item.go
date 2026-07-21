// Package service 同城商城业务逻辑层 - 订单明细
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrOrderItemNotFound = errors.New("订单明细不存在")
)

// OrderItemService 订单明细业务接口
type OrderItemService interface {
	GetByID(id uint) (*dto.OrderItemDetailInfo, error)
	ListByOrder(orderID uint) ([]dto.OrderItemDetailInfo, error)
	List(req *dto.OrderItemListRequest) (*utils.Pagination, []dto.OrderItemDetailInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.OrderItemDetailInfo, error)

	UpdateReviewStatus(id uint, hasReview bool, reviewID uint) error
	UpdateRefundStatus(id uint, refundStatus int, refundID uint) error
}

type orderItemService struct {
	repo repository.OrderItemRepository
}

// NewOrderItemService 创建订单明细 service 实例
func NewOrderItemService(repo repository.OrderItemRepository) OrderItemService {
	return &orderItemService{repo: repo}
}

// orderItemStatusText 订单明细状态文本
func orderItemStatusText(s int) string {
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

// toOrderItemDetailInfo model -> dto
func toOrderItemDetailInfo(it *model.OrderItem) *dto.OrderItemDetailInfo {
	return &dto.OrderItemDetailInfo{
		ID:             it.ID,
		OrderID:        it.OrderID,
		OrderNo:        it.OrderNo,
		ProductID:      it.ProductID,
		SkuID:          it.SkuID,
		ShopID:         it.ShopID,
		ProductName:    it.ProductName,
		MainImage:      it.MainImage,
		SkuName:        it.SkuName,
		SkuSpecs:       it.SkuSpecs,
		SkuCode:        it.SkuCode,
		Price:          it.Price,
		Quantity:       it.Quantity,
		TotalAmount:    it.TotalAmount,
		DiscountAmount: it.DiscountAmount,
		ShippingFee:    it.ShippingFee,
		PayAmount:      it.PayAmount,
		HasReview:      it.HasReview,
		ReviewID:       it.ReviewID,
		RefundStatus:   it.RefundStatus,
		RefundID:       it.RefundID,
		Status:         it.Status,
		StatusText:     orderItemStatusText(it.Status),
		RegionID:       it.RegionID,
	}
}

// GetByID 获取订单明细详情
func (s *orderItemService) GetByID(id uint) (*dto.OrderItemDetailInfo, error) {
	it, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderItemNotFound
		}
		return nil, err
	}
	return toOrderItemDetailInfo(it), nil
}

// ListByOrder 按订单列出明细
func (s *orderItemService) ListByOrder(orderID uint) ([]dto.OrderItemDetailInfo, error) {
	list, err := s.repo.ListByOrder(orderID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.OrderItemDetailInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderItemDetailInfo(&list[i]))
	}
	return infos, nil
}

// List 订单明细列表（管理后台）
func (s *orderItemService) List(req *dto.OrderItemListRequest) (*utils.Pagination, []dto.OrderItemDetailInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.OrderItemListOptions{
		OrderID:   req.OrderID,
		OrderNo:   req.OrderNo,
		ProductID: req.ProductID,
		SkuID:     req.SkuID,
		ShopID:    req.ShopID,
		Status:    req.Status,
		Keyword:   req.Keyword,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		RegionID:  req.RegionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.OrderItemDetailInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderItemDetailInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出订单明细
func (s *orderItemService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.OrderItemDetailInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.OrderItemDetailInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toOrderItemDetailInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateReviewStatus 更新评价状态
func (s *orderItemService) UpdateReviewStatus(id uint, hasReview bool, reviewID uint) error {
	return s.repo.UpdateReviewStatus(id, hasReview, reviewID)
}

// UpdateRefundStatus 更新退款状态
func (s *orderItemService) UpdateRefundStatus(id uint, refundStatus int, refundID uint) error {
	return s.repo.UpdateRefundStatus(id, refundStatus, refundID)
}
