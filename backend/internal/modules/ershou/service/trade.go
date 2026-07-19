// Package service 推广/物流/担保/退款业务逻辑层
// 依据 v3.2.1 架构方案：对标闲鱼/转转/瓜子
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/ershou/dto"
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/modules/ershou/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPromotionNotFound = errors.New("推广记录不存在")
	ErrLogisticsNotFound = errors.New("物流记录不存在")
	ErrEscrowNotFound    = errors.New("担保记录不存在")
	ErrEscrowStatus      = errors.New("担保状态不允许此操作")
	ErrRefundNotFound    = errors.New("退款记录不存在")
	ErrRefundStatus      = errors.New("退款状态不允许此操作")
	ErrRefundExists      = errors.New("该订单已有退款记录")
)

// ===== PromotionService =====

// PromotionService 推广业务接口
type PromotionService interface {
	Create(ershouID, userID uint, req *dto.PromotionCreateRequest) (*dto.PromotionResponse, error)
	ListByErshouID(ershouID uint) ([]dto.PromotionResponse, error)
	Stats(ershouID uint) (*dto.PromotionStatsResponse, error)
}

type promotionService struct {
	repo      repository.PromotionRepository
	ershouRepo repository.ErshouRepository
}

// NewPromotionService 创建推广 service 实例
func NewPromotionService(repo repository.PromotionRepository, ershouRepo repository.ErshouRepository) PromotionService {
	return &promotionService{repo: repo, ershouRepo: ershouRepo}
}

func promotionStatusText(status int) string {
	switch status {
	case model.PromotionStatusPending:
		return "待支付"
	case model.PromotionStatusActive:
		return "推广中"
	case model.PromotionStatusEnded:
		return "已结束"
	case model.PromotionStatusCanceled:
		return "已取消"
	}
	return "未知"
}

func (s *promotionService) Create(ershouID, userID uint, req *dto.PromotionCreateRequest) (*dto.PromotionResponse, error) {
	e, err := s.ershouRepo.FindByID(ershouID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrErshouNotFound
		}
		return nil, err
	}
	if e.UserID != userID {
		return nil, ErrErshouNoPermission
	}

	now := time.Now()
	endTime := now.AddDate(0, 0, req.DurationDays)
	promo := &model.ErshouPromotion{
		ErshouID:      ershouID,
		UserID:        userID,
		PromotionType: req.PromotionType,
		Status:        model.PromotionStatusActive,
		StartTime:     &now,
		EndTime:       &endTime,
		DurationDays:  req.DurationDays,
		Amount:        req.Amount,
		PayMethod:     req.PayMethod,
		PaidAt:        &now,
	}
	promo.RegionID = e.RegionID

	if err := s.repo.Create(promo); err != nil {
		return nil, err
	}
	// 主表推广等级 +1
	_ = s.ershouRepo.UpdateFields(ershouID, map[string]interface{}{
		"promotion_level": gorm.Expr("promotion_level + 1"),
	})
	return s.toPromotionResponse(promo), nil
}

func (s *promotionService) ListByErshouID(ershouID uint) ([]dto.PromotionResponse, error) {
	list, err := s.repo.ListByErshouID(ershouID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.PromotionResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toPromotionResponse(&list[i]))
	}
	return result, nil
}

func (s *promotionService) Stats(ershouID uint) (*dto.PromotionStatsResponse, error) {
	list, err := s.repo.ListByErshouID(ershouID)
	if err != nil {
		return nil, err
	}
	resp := &dto.PromotionStatsResponse{}
	for _, p := range list {
		resp.TotalPromotions++
		if p.Status == model.PromotionStatusActive {
			resp.ActivePromotions++
		}
		resp.TotalImpressions += int64(p.ImpressionCount)
		resp.TotalClicks += int64(p.ClickCount)
		resp.TotalOrders += int64(p.OrderCount)
		resp.TotalAmount += p.Amount
		if p.ClickCount > 0 {
			resp.AvgCTR += float64(p.ImpressionCount) / float64(p.ClickCount)
		}
		if p.Amount > 0 {
			resp.AvgROI += float64(p.OrderCount) / p.Amount
		}
	}
	if resp.TotalPromotions > 0 {
		resp.AvgCTR /= float64(resp.TotalPromotions)
		resp.AvgROI /= float64(resp.TotalPromotions)
	}
	return resp, nil
}

func (s *promotionService) toPromotionResponse(p *model.ErshouPromotion) *dto.PromotionResponse {
	return &dto.PromotionResponse{
		ID:               p.ID,
		ErshouID:         p.ErshouID,
		UserID:           p.UserID,
		PromotionType:    p.PromotionType,
		Status:           p.Status,
		StatusText:       promotionStatusText(p.Status),
		StartTime:        p.StartTime,
		EndTime:          p.EndTime,
		DurationDays:     p.DurationDays,
		Amount:           p.Amount,
		PayMethod:        p.PayMethod,
		PayTradeNo:       p.PayTradeNo,
		PaidAt:           p.PaidAt,
		ImpressionCount:  p.ImpressionCount,
		ClickCount:       p.ClickCount,
		FavCount:         p.FavCount,
		ConsultCount:     p.ConsultCount,
		OrderCount:       p.OrderCount,
		ROI:              p.ROI,
		CreatedAt:        p.CreatedAt,
	}
}

// ===== LogisticsService =====

// LogisticsService 物流业务接口
type LogisticsService interface {
	Create(orderID, userID uint, req *dto.LogisticsCreateRequest) (*dto.LogisticsResponse, error)
	GetByOrderID(orderID, userID uint) (*dto.LogisticsResponse, error)
	Update(orderID, userID uint, req *dto.LogisticsUpdateRequest) (*dto.LogisticsResponse, error)
}

type logisticsService struct {
	repo      repository.LogisticsRepository
	orderRepo repository.OrderRepository
}

// NewLogisticsService 创建物流 service 实例
func NewLogisticsService(repo repository.LogisticsRepository, orderRepo repository.OrderRepository) LogisticsService {
	return &logisticsService{repo: repo, orderRepo: orderRepo}
}

func logisticsStatusText(status int) string {
	switch status {
	case model.LogisticsStatusPending:
		return "待发货"
	case model.LogisticsStatusShipped:
		return "已发货"
	case model.LogisticsStatusInTransit:
		return "运输中"
	case model.LogisticsStatusDelivered:
		return "已签收"
	case model.LogisticsStatusException:
		return "异常"
	}
	return "未知"
}

func (s *logisticsService) Create(orderID, userID uint, req *dto.LogisticsCreateRequest) (*dto.LogisticsResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	// 已存在物流记录
	if existing, _ := s.repo.FindByOrderID(orderID); existing != nil {
		return nil, errors.New("该订单已有物流记录")
	}
	items, _ := s.orderRepo.ListItems(orderID)
	ershouID := uint(0)
	skuID := uint(0)
	if len(items) > 0 {
		ershouID = items[0].ErshouID
		skuID = items[0].SKUID
	}

	now := time.Now()
	l := &model.ErshouLogistics{
		OrderID:        orderID,
		ErshouID:       ershouID,
		SKUID:          skuID,
		ExpressCompany: req.ExpressCompany,
		ExpressCode:    req.ExpressCode,
		TrackingNo:     req.TrackingNo,
		Status:         model.LogisticsStatusShipped,
		ShipperName:    req.ShipperName,
		ShipperPhone:   req.ShipperPhone,
		ShipperAddress: req.ShipperAddress,
		ReceiverName:   order.ContactName,
		ReceiverPhone:  order.ContactPhone,
		ReceiverAddress: order.ContactAddress,
		Weight:         req.Weight,
		Freight:        req.Freight,
		ShippedAt:      &now,
		Remark:         req.Remark,
	}
	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return s.toLogisticsResponse(l), nil
}

func (s *logisticsService) GetByOrderID(orderID, userID uint) (*dto.LogisticsResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	l, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLogisticsNotFound
		}
		return nil, err
	}
	return s.toLogisticsResponse(l), nil
}

func (s *logisticsService) Update(orderID, userID uint, req *dto.LogisticsUpdateRequest) (*dto.LogisticsResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	l, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, ErrLogisticsNotFound
	}
	fields := map[string]interface{}{}
	if req.Status != nil {
		fields["status"] = *req.Status
		if *req.Status == model.LogisticsStatusDelivered {
			now := time.Now()
			fields["delivered_at"] = &now
		}
	}
	if req.TrackingInfo != nil {
		ti, _ := model.FromJSON(req.TrackingInfo)
		fields["tracking_info"] = ti
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	if err := s.repo.Update(l.ID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(l.ID)
	return s.toLogisticsResponse(updated), nil
}

func (s *logisticsService) toLogisticsResponse(l *model.ErshouLogistics) *dto.LogisticsResponse {
	resp := &dto.LogisticsResponse{
		ID:              l.ID,
		OrderID:         l.OrderID,
		ErshouID:        l.ErshouID,
		SKUID:           l.SKUID,
		ExpressCompany:  l.ExpressCompany,
		ExpressCode:     l.ExpressCode,
		TrackingNo:      l.TrackingNo,
		Status:          l.Status,
		StatusText:      logisticsStatusText(l.Status),
		ShipperName:     l.ShipperName,
		ShipperPhone:    l.ShipperPhone,
		ShipperAddress:  l.ShipperAddress,
		ReceiverName:    l.ReceiverName,
		ReceiverPhone:   l.ReceiverPhone,
		ReceiverAddress: l.ReceiverAddress,
		Weight:          l.Weight,
		Freight:         l.Freight,
		ShippedAt:       l.ShippedAt,
		DeliveredAt:     l.DeliveredAt,
		Remark:          l.Remark,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
		TrackingInfo:    []map[string]interface{}{},
	}
	if l.TrackingInfo != nil {
		var m []map[string]interface{}
		_ = l.TrackingInfo.Parse(&m)
		if m != nil {
			resp.TrackingInfo = m
		}
	}
	return resp
}

// ===== EscrowService =====

// EscrowService 担保业务接口
type EscrowService interface {
	Create(orderID, userID uint, req *dto.EscrowCreateRequest) (*dto.EscrowResponse, error)
	GetByOrderID(orderID, userID uint) (*dto.EscrowResponse, error)
	Release(orderID, userID uint, req *dto.EscrowReleaseRequest) (*dto.EscrowResponse, error)
}

type escrowService struct {
	repo      repository.EscrowRepository
	orderRepo repository.OrderRepository
}

// NewEscrowService 创建担保 service 实例
func NewEscrowService(repo repository.EscrowRepository, orderRepo repository.OrderRepository) EscrowService {
	return &escrowService{repo: repo, orderRepo: orderRepo}
}

func escrowStatusText(status int) string {
	switch status {
	case model.EscrowStatusNone:
		return "未启用"
	case model.EscrowStatusFrozen:
		return "资金冻结中"
	case model.EscrowStatusReleased:
		return "已解冻待放款"
	case model.EscrowStatusPaid:
		return "已放款"
	case model.EscrowStatusRefunded:
		return "已退款"
	case model.EscrowStatusDispute:
		return "纠纷处理中"
	}
	return "未知"
}

func (s *escrowService) Create(orderID, userID uint, req *dto.EscrowCreateRequest) (*dto.EscrowResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	if existing, _ := s.repo.FindByOrderID(orderID); existing != nil {
		return nil, errors.New("该订单已存在担保记录")
	}
	now := time.Now()
	e := &model.ErshouEscrow{
		OrderID:      orderID,
		ErshouID:     0,
		BuyerID:      order.BuyerID,
		SellerID:     order.SellerID,
		EscrowAmount: req.EscrowAmount,
		PlatformFee:  req.PlatformFee,
		SellerAmount: req.EscrowAmount - req.PlatformFee,
		Status:       model.EscrowStatusFrozen,
		FrozenAt:     &now,
	}
	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return s.toEscrowResponse(e), nil
}

func (s *escrowService) GetByOrderID(orderID, userID uint) (*dto.EscrowResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	e, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, ErrEscrowNotFound
	}
	return s.toEscrowResponse(e), nil
}

// Release 放款（卖家发起，T+1 自动放款在调度任务中实现）
func (s *escrowService) Release(orderID, userID uint, req *dto.EscrowReleaseRequest) (*dto.EscrowResponse, error) {
	e, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, ErrEscrowNotFound
	}
	if e.Status != model.EscrowStatusReleased && e.Status != model.EscrowStatusFrozen {
		return nil, ErrEscrowStatus
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":  model.EscrowStatusPaid,
		"paid_at": &now,
	}
	if req.DisputeReason != "" {
		fields["dispute_reason"] = req.DisputeReason
	}
	if err := s.repo.Update(e.ID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(e.ID)
	return s.toEscrowResponse(updated), nil
}

func (s *escrowService) toEscrowResponse(e *model.ErshouEscrow) *dto.EscrowResponse {
	return &dto.EscrowResponse{
		ID:                e.ID,
		OrderID:           e.OrderID,
		ErshouID:          e.ErshouID,
		BuyerID:           e.BuyerID,
		SellerID:          e.SellerID,
		EscrowAmount:      e.EscrowAmount,
		PlatformFee:       e.PlatformFee,
		SellerAmount:      e.SellerAmount,
		Status:            e.Status,
		StatusText:        escrowStatusText(e.Status),
		FrozenAt:          e.FrozenAt,
		ReleaseAt:         e.ReleaseAt,
		PaidAt:            e.PaidAt,
		RefundedAt:        e.RefundedAt,
		AutoReleaseAt:     e.AutoReleaseAt,
		DisputeReason:     e.DisputeReason,
		ArbitrationResult: e.ArbitrationResult,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}

// ===== RefundService =====

// RefundService 退款业务接口
type RefundService interface {
	Create(orderID, buyerID uint, req *dto.RefundCreateRequest) (*dto.RefundResponse, error)
	GetByOrderID(orderID, userID uint) (*dto.RefundResponse, error)
	Process(refundID, handlerID uint, req *dto.RefundProcessRequest) (*dto.RefundResponse, error)
	List(userID uint, query dto.OrderQuery) (*utils.Pagination, []dto.RefundResponse, error)
}

type refundService struct {
	repo      repository.RefundRepository
	orderRepo repository.OrderRepository
	escrowRepo repository.EscrowRepository
}

// NewRefundService 创建退款 service 实例
func NewRefundService(
	repo repository.RefundRepository,
	orderRepo repository.OrderRepository,
	escrowRepo repository.EscrowRepository,
) RefundService {
	return &refundService{repo: repo, orderRepo: orderRepo, escrowRepo: escrowRepo}
}

func refundStatusText(status int) string {
	switch status {
	case model.RefundStatusPending:
		return "待卖家处理"
	case model.RefundStatusSellerRejected:
		return "卖家拒绝"
	case model.RefundStatusSellerAgreed:
		return "卖家同意"
	case model.RefundStatusShipping:
		return "退货运输中"
	case model.RefundStatusReceived:
		return "卖家已收到退货"
	case model.RefundStatusRefunding:
		return "退款中"
	case model.RefundStatusCompleted:
		return "退款完成"
	case model.RefundStatusCanceled:
		return "买家取消"
	case model.RefundStatusDispute:
		return "平台介入"
	case model.RefundStatusArbitrated:
		return "仲裁完成"
	}
	return "未知"
}

// genRefundNo 生成退款单号：RF + yyyyMMddHHmmss + 6位随机数
func genRefundNo() string {
	return fmt.Sprintf("RF%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 买家申请退款
// 仅在订单已支付后允许申请退款
func (s *refundService) Create(orderID, buyerID uint, req *dto.RefundCreateRequest) (*dto.RefundResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != buyerID {
		return nil, ErrOrderNoPermission
	}
	// 订单状态校验：仅已支付/已发货/待收货/已完成 可申请退款（不允许待支付/已取消/已关闭）
	switch order.Status {
	case model.OrderStatusPaid, model.OrderStatusShipped,
		model.OrderStatusDelivered, model.OrderStatusCompleted:
		// 允许
	default:
		return nil, ErrOrderStatusInvalid
	}
	// 同一订单只能有一条有效退款
	if existing, _ := s.repo.FindByOrderID(orderID); existing != nil {
		if existing.Status != model.RefundStatusCanceled &&
			existing.Status != model.RefundStatusCompleted {
			return nil, ErrRefundExists
		}
	}
	// 退款金额校验
	if req.RefundAmount > order.TotalAmount {
		return nil, errors.New("退款金额不能超过订单总金额")
	}

	now := time.Now()
	slaDeadline := now.Add(72 * time.Hour) // SLA 72h
	ershouID := uint(0)
	items, _ := s.orderRepo.ListItems(orderID)
	if len(items) > 0 {
		ershouID = items[0].ErshouID
	}

	refund := &model.ErshouRefund{
		RefundNo:        genRefundNo(),
		OrderID:         orderID,
		ErshouID:        ershouID,
		BuyerID:         order.BuyerID,
		SellerID:        order.SellerID,
		RefundType:      req.RefundType,
		Reason:          req.Reason,
		Description:     req.Description,
		RefundAmount:    req.RefundAmount,
		Status:          model.RefundStatusPending,
		SLADeadline:     &slaDeadline,
	}
	if len(req.EvidenceImages) > 0 {
		if jb, err := model.FromJSON(req.EvidenceImages); err == nil {
			refund.EvidenceImages = jb
		}
	}
	if err := s.repo.Create(refund); err != nil {
		return nil, err
	}
	return s.toRefundResponse(refund), nil
}

// GetByOrderID 查询订单退款详情（买卖双方均可查询）
func (s *refundService) GetByOrderID(orderID, userID uint) (*dto.RefundResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if order.BuyerID != userID && order.SellerID != userID {
		return nil, ErrOrderNoPermission
	}
	rf, err := s.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, ErrRefundNotFound
	}
	return s.toRefundResponse(rf), nil
}

// Process 处理退款（卖家 approve/reject，平台 arbitrate）
func (s *refundService) Process(refundID, handlerID uint, req *dto.RefundProcessRequest) (*dto.RefundResponse, error) {
	rf, err := s.repo.FindByID(refundID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}

	now := time.Now()
	fields := map[string]interface{}{}

	switch req.Action {
	case "approve":
		// 仅卖家可同意
		if rf.SellerID != handlerID {
			return nil, ErrOrderNoPermission
		}
		if rf.Status != model.RefundStatusPending {
			return nil, ErrRefundStatus
		}
		// 仅退款（refund）→ 直接进入退款中 → 完成
		// 退货退款（return）/ 换货（exchange）/ 维修（repair）→ 卖家同意后进入退货运输
		if rf.RefundType == model.RefundTypeRefund {
			fields["status"] = model.RefundStatusRefunding
		} else {
			fields["status"] = model.RefundStatusSellerAgreed
		}

	case "reject":
		// 仅卖家可拒绝
		if rf.SellerID != handlerID {
			return nil, ErrOrderNoPermission
		}
		if rf.Status != model.RefundStatusPending {
			return nil, ErrRefundStatus
		}
		fields["status"] = model.RefundStatusSellerRejected
		fields["seller_reason"] = req.SellerReason

	case "arbitrate":
		// 平台介入仲裁（任意状态均可介入）
		fields["status"] = model.RefundStatusArbitrated
		fields["arbitrator_id"] = handlerID
		fields["arbitrated_at"] = &now
		if req.ArbitrationResult != "" {
			fields["arbitration_result"] = req.ArbitrationResult
		}
		// 仲裁完成 = 退款完成（统一标记 completed_at）
		completedAt := now
		fields["completed_at"] = &completedAt
		// 触发担保退款（如启用担保）
		if esc, err := s.escrowRepo.FindByOrderID(rf.OrderID); err == nil && esc != nil {
			_ = s.escrowRepo.Update(esc.ID, map[string]interface{}{
				"status":       model.EscrowStatusRefunded,
				"refunded_at":  &now,
				"dispute_reason": req.ArbitrationResult,
			})
		}

	default:
		return nil, ErrRefundStatus
	}

	if err := s.repo.Update(refundID, fields); err != nil {
		return nil, err
	}
	updated, _ := s.repo.FindByID(refundID)
	return s.toRefundResponse(updated), nil
}

// List 退款列表（按角色查询）
func (s *refundService) List(userID uint, query dto.OrderQuery) (*utils.Pagination, []dto.RefundResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	if query.Role == "" {
		query.Role = "all"
	}
	list, total, err := s.repo.List(repository.RefundListQuery{
		UserID:   userID,
		Role:     query.Role,
		Status:   query.Status,
		RefundNo: query.OrderNo,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RefundResponse, 0, len(list))
	for i := range list {
		result = append(result, *s.toRefundResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *refundService) toRefundResponse(rf *model.ErshouRefund) *dto.RefundResponse {
	resp := &dto.RefundResponse{
		ID:                rf.ID,
		RefundNo:          rf.RefundNo,
		OrderID:           rf.OrderID,
		ErshouID:          rf.ErshouID,
		BuyerID:           rf.BuyerID,
		SellerID:          rf.SellerID,
		RefundType:        rf.RefundType,
		Reason:            rf.Reason,
		Description:       rf.Description,
		EvidenceImages:    []string{},
		RefundAmount:      rf.RefundAmount,
		Status:            rf.Status,
		StatusText:        refundStatusText(rf.Status),
		SellerReason:      rf.SellerReason,
		ArbitrationResult: rf.ArbitrationResult,
		ArbitratorID:      rf.ArbitratorID,
		ArbitratedAt:      rf.ArbitratedAt,
		SLADeadline:       rf.SLADeadline,
		CompletedAt:       rf.CompletedAt,
		CreatedAt:         rf.CreatedAt,
		UpdatedAt:         rf.UpdatedAt,
	}
	if rf.EvidenceImages != nil {
		var imgs []string
		_ = rf.EvidenceImages.Parse(&imgs)
		if imgs != nil {
			resp.EvidenceImages = imgs
		}
	}
	return resp
}
