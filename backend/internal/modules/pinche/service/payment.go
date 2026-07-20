// Package service 同城拼车出行业务逻辑层 - 支付（含 ETC）
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound      = errors.New("支付记录不存在")
	ErrPaymentStatusInvalid = errors.New("支付状态不允许此操作")
	ErrPaymentNoInvalid     = errors.New("支付单号无效")
)

// PaymentService 支付业务接口
type PaymentService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error)
	Callback(req *dto.PaymentCallbackRequest) (*dto.PaymentInfo, error)
	GetByID(id uint) (*dto.PaymentInfo, error)
	GetByPaymentNo(paymentNo string) (*dto.PaymentInfo, error)
	ListByPayer(payerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByPayee(payeeID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)
	ListByBooking(bookingID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error)

	// ETC 结算
	ETCSettlement(req *dto.ETCSettlementRequest) (*dto.PaymentInfo, error)

	// M 端
	AdminList(req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error)
	UpdateStatus(id uint, status int) error
}

type paymentService struct {
	repo repository.PaymentRepository
}

// NewPaymentService 创建支付 service 实例
func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{repo: repo}
}

// paymentStatusText 状态文本
func paymentStatusText(status int) string {
	switch status {
	case model.PaymentStatusPending:
		return "待支付"
	case model.PaymentStatusPaid:
		return "已支付"
	case model.PaymentStatusRefund:
		return "已退款"
	case model.PaymentStatusFailed:
		return "已失败"
	}
	return ""
}

// toPaymentInfo model -> dto
func toPaymentInfo(p *model.PinchePayment) *dto.PaymentInfo {
	return &dto.PaymentInfo{
		ID:                p.ID,
		RegionID:          p.RegionID,
		PincheID:          p.PincheID,
		BookingID:         p.BookingID,
		PaymentNo:         p.PaymentNo,
		PayerID:           p.PayerID,
		PayerName:         p.PayerName,
		PayeeID:           p.PayeeID,
		PayeeName:         p.PayeeName,
		Amount:            p.Amount,
		InsuranceFee:      p.InsuranceFee,
		ServiceFee:        p.ServiceFee,
		TotalAmount:       p.TotalAmount,
		PaymentMethod:     p.PaymentMethod,
		PaymentStatus:     p.PaymentStatus,
		PaymentStatusText: paymentStatusText(p.PaymentStatus),
		PaidAt:            p.PaidAt,
		RefundedAt:        p.RefundedAt,
		ThirdPartyNo:      p.ThirdPartyNo,
		RefundAmount:      p.RefundAmount,
		ETCLaneID:         p.ETCLaneID,
		ETCEntryTime:      p.ETCEntryTime,
		ETCExitTime:       p.ETCExitTime,
		CreatedAt:         p.CreatedAt,
	}
}

// genPaymentNo 生成支付单号 PAY + yyyyMMddHHmmss + 8位hex
func genPaymentNo() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("PAY%s%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}

// Create 创建支付记录（待支付状态）
func (s *paymentService) Create(regionID uint, userID uint, userName string, req *dto.CreatePaymentRequest) (*dto.PaymentInfo, error) {
	paymentNo := genPaymentNo()
	// 这里仅创建支付记录，金额、收款方等需要上游（booking）填充。
	// 简化处理：amount/total_amount=0，待回调或上游补齐
	p := &model.PinchePayment{
		BookingID:     req.BookingID,
		PaymentNo:     paymentNo,
		PayerID:       userID,
		PayerName:     userName,
		PaymentMethod: req.PaymentMethod,
		PaymentStatus: model.PaymentStatusPending,
		ETCLaneID:     req.ETCLaneID,
	}
	p.RegionID = regionID

	// ETC 支付：默认等回调设置入/出口时间
	if req.PaymentMethod == model.PaymentMethodETC {
		now := time.Now()
		p.ETCEntryTime = &now
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPaymentInfo(p), nil
}

// Callback 支付回调
func (s *paymentService) Callback(req *dto.PaymentCallbackRequest) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByPaymentNo(req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNoInvalid
		}
		return nil, err
	}
	if p.PaymentStatus != model.PaymentStatusPending {
		return nil, ErrPaymentStatusInvalid
	}
	fields := map[string]interface{}{
		"payment_status":  req.PaymentStatus,
		"third_party_no":  req.ThirdPartyNo,
	}
	if req.PaymentStatus == model.PaymentStatusPaid {
		now := time.Now()
		if req.PaidAt != nil {
			fields["paid_at"] = *req.PaidAt
		} else {
			fields["paid_at"] = now
		}
	}
	if err := s.repo.Update(p.ID, fields); err != nil {
		return nil, err
	}
	p.PaymentStatus = req.PaymentStatus
	p.ThirdPartyNo = req.ThirdPartyNo
	if req.PaidAt != nil {
		p.PaidAt = req.PaidAt
	}
	return toPaymentInfo(p), nil
}

// GetByID 获取详情
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

// ListByPayer 按付款方查询
func (s *paymentService) ListByPayer(payerID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPayer(payerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByPayee 按收款方查询
func (s *paymentService) ListByPayee(payeeID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPayee(payeeID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByBooking 按预订查询
func (s *paymentService) ListByBooking(bookingID uint, page, pageSize int) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByBooking(bookingID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// ETCSettlement ETC 结算
func (s *paymentService) ETCSettlement(req *dto.ETCSettlementRequest) (*dto.PaymentInfo, error) {
	p, err := s.repo.FindByID(req.PaymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if p.PaymentMethod != model.PaymentMethodETC {
		return nil, ErrPaymentStatusInvalid
	}
	var entryTime, exitTime interface{}
	if req.EntryTime != nil {
		entryTime = *req.EntryTime
	}
	if req.ExitTime != nil {
		exitTime = *req.ExitTime
	}
	if err := s.repo.UpdateETC(p.ID, req.ETCLaneID, entryTime, exitTime); err != nil {
		return nil, err
	}
	// 更新金额
	if req.Amount > 0 {
		_ = s.repo.Update(p.ID, map[string]interface{}{
			"amount":       req.Amount,
			"total_amount": req.Amount,
		})
		p.Amount = req.Amount
		p.TotalAmount = req.Amount
	}
	p.ETCLaneID = req.ETCLaneID
	if req.EntryTime != nil {
		p.ETCEntryTime = req.EntryTime
	}
	if req.ExitTime != nil {
		p.ETCExitTime = req.ExitTime
	}
	return toPaymentInfo(p), nil
}

// AdminList 管理后台支付列表
func (s *paymentService) AdminList(req *dto.PaymentListRequest) (*utils.Pagination, []dto.PaymentInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.PaymentListOptions{
		PincheID:      req.PincheID,
		BookingID:     req.BookingID,
		PayerID:       req.PayerID,
		PayeeID:       req.PayeeID,
		PaymentMethod: req.PaymentMethod,
		Status:        req.Status,
		PaymentNo:     req.PaymentNo,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.PaymentInfo, 0, len(list))
	for i := range list {
		result = append(result, *toPaymentInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 管理后台更新状态
func (s *paymentService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
