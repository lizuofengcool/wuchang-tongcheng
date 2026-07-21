// Package service 同城商城业务逻辑层 - 支付
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
	ErrPaymentNotFound = errors.New("支付记录不存在")
)

// 支付单过期时长（2 小时）
const paymentExpiredHours = 2

// PaymentService 支付业务接口
type PaymentService interface {
	Create(regionID, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error)
	HandleCallback(req *dto.PaymentCallbackRequest) error
	Close(id uint, req *dto.ClosePaymentRequest) error

	GetByID(id uint) (*dto.PaymentInfo, error)
	GetByPaymentNo(paymentNo string) (*dto.PaymentInfo, error)
	GetByOrderID(orderID uint) (*dto.PaymentInfo, error)
	List(req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)

	Stats(regionID, shopID uint, startDate, endDate string) (*dto.PaymentStats, error)
}

type paymentService struct {
	repo     repository.PaymentRepository
	orderRepo repository.OrderRepository
}

// NewPaymentService 创建支付 service 实例
func NewPaymentService(repo repository.PaymentRepository, orderRepo repository.OrderRepository) PaymentService {
	return &paymentService{repo: repo, orderRepo: orderRepo}
}

// paymentStatusText 支付状态文本
func paymentStatusText(s int) string {
	switch s {
	case model.PaymentStatusPending:
		return "待支付"
	case model.PaymentStatusSuccess:
		return "成功"
	case model.PaymentStatusFailed:
		return "失败"
	case model.PaymentStatusRefunded:
		return "已退款"
	case model.PaymentStatusClosed:
		return "已关闭"
	}
	return ""
}

// toPaymentInfo model -> dto
func toPaymentInfo(p *model.Payment) *dto.PaymentInfo {
	info := &dto.PaymentInfo{
		ID:           p.ID,
		OrderID:      p.OrderID,
		OrderNo:      p.OrderNo,
		UserID:       p.UserID,
		ShopID:       p.ShopID,
		PaymentNo:    p.PaymentNo,
		TradeNo:      p.TradeNo,
		OutTradeNo:   p.OutTradeNo,
		Amount:       p.Amount,
		RefundAmount: p.RefundAmount,
		Method:       p.Method,
		Channel:      p.Channel,
		Status:       p.Status,
		StatusText:   paymentStatusText(p.Status),
		ErrorCode:    p.ErrorCode,
		ErrorMsg:     p.ErrorMsg,
		PaidAt:       p.PaidAt,
		ExpiredAt:    p.ExpiredAt,
		ClosedAt:     p.ClosedAt,
		RefundedAt:   p.RefundedAt,
		ClientIP:     p.ClientIP,
		UserAgent:    p.UserAgent,
		RegionID:     p.RegionID,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
	if p.RawResponse != nil {
		info.RawResponse = p.RawResponse
	}
	return info
}

// generatePaymentNo 生成支付单号
func generatePaymentNo() string {
	return fmt.Sprintf("MALP%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建支付单
func (s *paymentService) Create(regionID, userID uint, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error) {
	// 校验订单
	o, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if o.UserID != userID {
		return nil, ErrOrderNotOwner
	}
	if o.Status != model.OrderStatusPending {
		return nil, ErrOrderStatusInvalid
	}

	now := time.Now()
	expiredAt := now.Add(time.Hour * time.Duration(paymentExpiredHours))

	p := &model.Payment{
		OrderID:    req.OrderID,
		OrderNo:    req.OrderNo,
		UserID:     userID,
		ShopID:     o.ShopID,
		PaymentNo:  generatePaymentNo(),
		OutTradeNo: o.OrderNo,
		Amount:     o.PayAmount,
		Method:     req.PaymentMethod,
		Channel:    req.Channel,
		Status:     model.PaymentStatusPending,
		ExpiredAt:  &expiredAt,
	}
	p.RegionID = regionID

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}

	// 同步订单 payment_method 和 payment_no
	_ = s.orderRepo.UpdateFields(o.ID, map[string]interface{}{
		"payment_method": req.PaymentMethod,
		"payment_no":     p.PaymentNo,
	})

	return toPaymentInfo(p), nil
}

// HandleCallback 处理第三方支付回调
func (s *paymentService) HandleCallback(req *dto.PaymentCallbackRequest) error {
	p, err := s.repo.FindByPaymentNo(req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"trade_no":   req.TradeNo,
		"error_code": req.ErrorCode,
		"error_msg":  req.ErrorMsg,
	}
	if req.RawResponse != nil {
		if b, err := model.FromJSON(req.RawResponse); err == nil {
			fields["raw_response"] = b
		}
	}

	// 状态映射：第三方回调 status 1=成功 其他=失败
	if req.Status == 1 {
		now := time.Now()
		fields["status"] = model.PaymentStatusSuccess
		fields["paid_at"] = &now
		// 同步更新订单为已付款
		_ = s.orderRepo.UpdateFields(p.OrderID, map[string]interface{}{
			"status":  model.OrderStatusPaid,
			"paid_at": &now,
		})
	} else {
		fields["status"] = model.PaymentStatusFailed
	}

	return s.repo.UpdateFields(p.ID, fields)
}

// Close 关闭支付单
func (s *paymentService) Close(id uint, req *dto.ClosePaymentRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}
	if p.Status != model.PaymentStatusPending {
		return errors.New("仅待支付状态的支付单可关闭")
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":    model.PaymentStatusClosed,
		"closed_at": &now,
	}
	_ = req
	return s.repo.UpdateFields(id, fields)
}

// GetByID 获取支付详情
func (s *paymentService) GetByID(id uint) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// GetByPaymentNo 按支付单号查询
func (s *paymentService) GetByPaymentNo(paymentNo string) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByPaymentNo(paymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// GetByOrderID 按订单 ID 查询支付单
func (s *paymentService) GetByOrderID(orderID uint) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// List 支付列表（管理后台）
func (s *paymentService) List(req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PaymentListOptions{
		OrderID:   req.OrderID,
		OrderNo:   req.OrderNo,
		UserID:    req.UserID,
		ShopID:    req.ShopID,
		PaymentNo: req.PaymentNo,
		TradeNo:   req.TradeNo,
		Method:    req.Method,
		Status:    req.Status,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		RegionID:  req.RegionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPaymentInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出
func (s *paymentService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPaymentInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByShop 按店铺列出
func (s *paymentService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPaymentInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Stats 支付统计
func (s *paymentService) Stats(regionID, shopID uint, startDate, endDate string) (*dto.PaymentStats, error) {
	opts := repository.PaymentStatsOptions{
		RegionID:  regionID,
		ShopID:    shopID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	result, err := s.repo.Stats(opts)
	if err != nil {
		return nil, err
	}
	return &dto.PaymentStats{
		TotalCount:    result.TotalCount,
		TotalAmount:   result.TotalAmount,
		SuccessCount:  result.SuccessCount,
		SuccessAmount: result.SuccessAmount,
		PendingCount:  result.PendingCount,
		FailedCount:   result.FailedCount,
		RefundedCount: result.RefundedCount,
		RefundAmount:  result.RefundAmount,
	}, nil
}
