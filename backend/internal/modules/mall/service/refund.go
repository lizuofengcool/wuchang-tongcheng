// Package service 同城商城业务逻辑层 - 退款
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
	ErrRefundNotFound      = errors.New("退款记录不存在")
	ErrRefundNotOwner      = errors.New("无权操作他人退款")
	ErrRefundStatusInvalid = errors.New("退款状态不允许此操作")
)

// RefundService 退款业务接口
type RefundService interface {
	Create(regionID, userID uint, req *dto.CreateRefundRequest) (*dto.RefundInfo, error)
	SellerProcess(id, sellerID uint, req *dto.SellerProcessRefundRequest) error
	AdminProcess(id, adminID uint, adminName string, req *dto.AdminProcessRefundRequest) error
	Ship(id, userID uint, req *dto.RefundShipRequest) error

	GetByID(id uint) (*dto.RefundInfo, error)
	GetByRefundNo(refundNo string) (*dto.RefundInfo, error)
	List(req *dto.RefundListRequest) (*utils.Pagination, []dto.RefundInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RefundInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.RefundInfo, error)
	ListByOrder(orderID uint) ([]dto.RefundInfo, error)

	Stats(regionID, shopID uint, startDate, endDate string) (*dto.RefundStats, error)
}

type refundService struct {
	repo         repository.RefundRepository
	orderRepo    repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	paymentRepo  repository.PaymentRepository
}

// NewRefundService 创建退款 service 实例
func NewRefundService(
	repo repository.RefundRepository,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	paymentRepo repository.PaymentRepository,
) RefundService {
	return &refundService{
		repo:          repo,
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		paymentRepo:   paymentRepo,
	}
}

// refundStatusText 退款状态文本
func refundStatusText(s int) string {
	switch s {
	case model.RefundStatusPending:
		return "待审核"
	case model.RefundStatusApproved:
		return "已同意"
	case model.RefundStatusRejected:
		return "已拒绝"
	case model.RefundStatusRefunded:
		return "已退款"
	case model.RefundStatusClosed:
		return "已关闭"
	}
	return ""
}

// toRefundInfo model -> dto
func toRefundInfo(r *model.Refund) *dto.RefundInfo {
	info := &dto.RefundInfo{
		ID:             r.ID,
		OrderID:        r.OrderID,
		OrderNo:        r.OrderNo,
		OrderItemID:    r.OrderItemID,
		UserID:         r.UserID,
		ShopID:         r.ShopID,
		PaymentID:      r.PaymentID,
		RefundNo:       r.RefundNo,
		TradeNo:        r.TradeNo,
		Amount:         r.Amount,
		RefundAmount:   r.RefundAmount,
		RefundType:     r.RefundType,
		Reason:         r.Reason,
		ReasonCode:     r.ReasonCode,
		Description:    r.Description,
		ExpressCompany: r.ExpressCompany,
		ExpressNo:      r.ExpressNo,
		Status:         r.Status,
		StatusText:     refundStatusText(r.Status),
		SellerRemark:   r.SellerRemark,
		AdminRemark:    r.AdminRemark,
		HandlerID:      r.HandlerID,
		HandlerName:    r.HandlerName,
		ApprovedAt:     r.ApprovedAt,
		RejectedAt:     r.RejectedAt,
		RefundedAt:     r.RefundedAt,
		ClosedAt:       r.ClosedAt,
		ShippedAt:      r.ShippedAt,
		ReceivedAt:     r.ReceivedAt,
		RegionID:       r.RegionID,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.EvidenceImages != nil {
		info.EvidenceImages = r.EvidenceImages
	}
	if r.RawResponse != nil {
		info.RawResponse = r.RawResponse
	}
	return info
}

// generateRefundNo 生成退款单号
func generateRefundNo() string {
	return fmt.Sprintf("MALR%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建退款单
func (s *refundService) Create(regionID, userID uint, req *dto.CreateRefundRequest) (*dto.RefundInfo, error) {
	// 校验订单
	o, err := s.orderRepo.FindByID(req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if o.UserID != userID {
		return nil, ErrRefundNotOwner
	}
	// 仅已付款及以上可退款（但不能是已关闭）
	if o.Status < model.OrderStatusPaid || o.Status >= model.OrderStatusClosed {
		return nil, ErrOrderStatusInvalid
	}

	// 默认退款金额为订单实付金额
	amount := req.Amount
	if amount == 0 {
		amount = o.PayAmount
	}

	r := &model.Refund{
		OrderID:     req.OrderID,
		OrderNo:     o.OrderNo,
		OrderItemID: req.OrderItemID,
		UserID:      userID,
		ShopID:      o.ShopID,
		RefundNo:    generateRefundNo(),
		Amount:      amount,
		RefundType:  req.RefundType,
		Reason:      req.Reason,
		ReasonCode:  req.ReasonCode,
		Description: req.Description,
		Status:      model.RefundStatusPending,
	}
	r.RegionID = regionID

	// 关联支付单
	if p, err := s.paymentRepo.FindByOrderID(o.ID); err == nil {
		r.PaymentID = p.ID
		r.TradeNo = p.TradeNo
	}

	if req.EvidenceImages != nil {
		if b, err := model.FromJSON(req.EvidenceImages); err == nil {
			r.EvidenceImages = b
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}

	// 同步订单明细退款状态
	if req.OrderItemID > 0 {
		_ = s.orderItemRepo.UpdateRefundStatus(req.OrderItemID, 1, r.ID)
	}

	return toRefundInfo(r), nil
}

// SellerProcess 卖家处理退款
func (s *refundService) SellerProcess(id, sellerID uint, req *dto.SellerProcessRefundRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}
	if r.Status != model.RefundStatusPending && r.Status != model.RefundStatusApproved {
		return ErrRefundStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":        req.Status,
		"seller_remark": req.SellerRemark,
		"handler_id":    sellerID,
	}
	switch req.Status {
	case model.RefundStatusApproved:
		fields["approved_at"] = &now
	case model.RefundStatusRejected:
		fields["rejected_at"] = &now
	case model.RefundStatusRefunded:
		fields["refunded_at"] = &now
		fields["refund_amount"] = r.Amount
		// 同步订单为已退款
		_ = s.orderRepo.UpdateFields(r.OrderID, map[string]interface{}{
			"status": model.OrderStatusRefunded,
		})
		// 同步订单明细退款状态
		if r.OrderItemID > 0 {
			_ = s.orderItemRepo.UpdateRefundStatus(r.OrderItemID, 4, r.ID)
		}
	}
	return s.repo.UpdateFields(id, fields)
}

// AdminProcess 管理后台处理退款
func (s *refundService) AdminProcess(id, adminID uint, adminName string, req *dto.AdminProcessRefundRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":       req.Status,
		"admin_remark": req.AdminRemark,
		"handler_id":   adminID,
		"handler_name": adminName,
	}
	switch req.Status {
	case model.RefundStatusApproved:
		fields["approved_at"] = &now
	case model.RefundStatusRejected:
		fields["rejected_at"] = &now
	case model.RefundStatusRefunded:
		fields["refunded_at"] = &now
		fields["refund_amount"] = r.Amount
		// 同步订单为已退款
		_ = s.orderRepo.UpdateFields(r.OrderID, map[string]interface{}{
			"status": model.OrderStatusRefunded,
		})
		// 同步订单明细退款状态
		if r.OrderItemID > 0 {
			_ = s.orderItemRepo.UpdateRefundStatus(r.OrderItemID, 4, r.ID)
		}
	}
	return s.repo.UpdateFields(id, fields)
}

// Ship 买家退货物流（退款退货场景）
func (s *refundService) Ship(id, userID uint, req *dto.RefundShipRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefundNotFound
		}
		return err
	}
	if r.UserID != userID {
		return ErrRefundNotOwner
	}
	if r.RefundType != model.RefundTypeReturn {
		return errors.New("仅退货退款类型可填写退货物流")
	}

	now := time.Now()
	fields := map[string]interface{}{
		"express_company": req.ExpressCompany,
		"express_no":      req.ExpressNo,
		"shipped_at":      &now,
	}
	return s.repo.UpdateFields(id, fields)
}

// GetByID 获取退款详情
func (s *refundService) GetByID(id uint) (*dto.RefundInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return toRefundInfo(r), nil
}

// GetByRefundNo 按退款单号查询
func (s *refundService) GetByRefundNo(refundNo string) (*dto.RefundInfo, error) {
	r, err := s.repo.FindByRefundNo(refundNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return toRefundInfo(r), nil
}

// List 退款列表（管理后台）
func (s *refundService) List(req *dto.RefundListRequest) (*utils.Pagination, []dto.RefundInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RefundListOptions{
		OrderID:    req.OrderID,
		OrderNo:    req.OrderNo,
		UserID:     req.UserID,
		ShopID:     req.ShopID,
		RefundNo:   req.RefundNo,
		Status:     req.Status,
		RefundType: req.RefundType,
		Keyword:    req.Keyword,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		RegionID:   req.RegionID,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RefundInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRefundInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出
func (s *refundService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RefundInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RefundInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRefundInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByShop 按店铺列出
func (s *refundService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.RefundInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByShop(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RefundInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRefundInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByOrder 按订单列出
func (s *refundService) ListByOrder(orderID uint) ([]dto.RefundInfo, error) {
	list, err := s.repo.ListByOrder(orderID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.RefundInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRefundInfo(&list[i]))
	}
	return infos, nil
}

// Stats 退款统计
func (s *refundService) Stats(regionID, shopID uint, startDate, endDate string) (*dto.RefundStats, error) {
	opts := repository.RefundStatsOptions{
		RegionID:  regionID,
		ShopID:    shopID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	result, err := s.repo.Stats(opts)
	if err != nil {
		return nil, err
	}
	return &dto.RefundStats{
		TotalCount:     result.TotalCount,
		TotalAmount:    result.TotalAmount,
		PendingCount:   result.PendingCount,
		ApprovedCount:  result.ApprovedCount,
		RejectedCount:  result.RejectedCount,
		RefundedCount:  result.RefundedCount,
		RefundedAmount: result.RefundedAmount,
		ReturningCount: result.ReturningCount,
	}, nil
}
