// Package service 同城商城业务逻辑层 - 骑手结算
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	// ErrRiderSettlementNotFound 结算单不存在
	ErrRiderSettlementNotFound = errors.New("结算单不存在")
	// ErrRiderSettlementAudited 结算单已审核
	ErrRiderSettlementAudited = errors.New("结算单已审核")
	// ErrRiderSettlementError 结算错误
	ErrRiderSettlementError = errors.New("结算错误")
)

// RiderSettlementService 骑手结算业务接口
type RiderSettlementService interface {
	Generate(regionID uint, req *dto.RiderSettlementGenerateRequest) (*dto.RiderSettlementInfo, error)
	GetByID(id uint) (*dto.RiderSettlementInfo, error)
	List(req *dto.RiderSettlementListRequest) (*utils.Pagination, []dto.RiderSettlementInfo, error)
	ListByUser(userID uint, req *dto.RiderSettlementListRequest) (*utils.Pagination, []dto.RiderSettlementInfo, error)
	Audit(id uint, req *dto.RiderSettlementAuditRequest) error
	Withdraw(userID uint, req *dto.RiderSettlementWithdrawRequest) error
	Stats(userID uint) (*dto.RiderSettlementStatsResponse, error)
}

type riderSettlementService struct {
	repo         repository.RiderSettlementRepository
	riderRepo    repository.RiderRepository
	deliveryRepo repository.DeliveryRepository
}

// NewRiderSettlementService 创建骑手结算 service 实例
func NewRiderSettlementService(
	repo repository.RiderSettlementRepository,
	riderRepo repository.RiderRepository,
	deliveryRepo repository.DeliveryRepository,
) RiderSettlementService {
	return &riderSettlementService{repo: repo, riderRepo: riderRepo, deliveryRepo: deliveryRepo}
}

func settlementStatusText(s int) string {
	switch s {
	case model.RiderSettlementStatusPending:
		return "待结算"
	case model.RiderSettlementStatusSettled:
		return "已结算"
	case model.RiderSettlementStatusWithdrawn:
		return "已提现"
	}
	return ""
}

func toRiderSettlementInfo(s *model.RiderSettlement) *dto.RiderSettlementInfo {
	return &dto.RiderSettlementInfo{
		ID:          s.ID,
		RiderID:     s.RiderID,
		Period:      s.Period,
		TotalOrders: s.TotalOrders,
		TotalAmount: s.TotalAmount,
		TotalFee:    s.TotalFee,
		TotalTip:    s.TotalTip,
		PlatformFee: s.PlatformFee,
		NetAmount:   s.NetAmount,
		Status:      s.Status,
		StatusText:  settlementStatusText(s.Status),
		SettledAt:   s.SettledAt,
		WithdrawnAt: s.WithdrawnAt,
		AuditReason: s.AuditReason,
		RegionID:    s.RegionID,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// Generate 生成结算单
// 业务规则：net_amount = total_fee + total_tip - platform_fee
func (s *riderSettlementService) Generate(regionID uint, req *dto.RiderSettlementGenerateRequest) (*dto.RiderSettlementInfo, error) {
	rider, err := s.riderRepo.FindByID(req.RiderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}

	// 解析周期：YYYY-MM
	period := req.Period
	// 周期起始 / 结束
	periodStart, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, ErrRiderSettlementError
	}
	periodEnd := periodStart.AddDate(0, 1, 0)

	// 统计周期内已送达订单的配送费 + 小费
	totalEarnings, totalOrders, err := s.deliveryRepo.SumEarnings(rider.ID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// 拆分：分别查配送费总额和小费总额
	totalFee, totalTip, _, err := s.deliveryRepo.SumFeeAndTip(rider.ID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// 平台抽成
	platformFee := req.PlatformFee
	netAmount := totalFee + totalTip - platformFee

	// 幂等：如果该周期已有结算单，更新；否则创建
	existing, err := s.repo.FindByRiderAndPeriod(rider.ID, period)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		fields := map[string]interface{}{
			"total_orders":  int(totalOrders),
			"total_fee":     totalFee,
			"total_tip":     totalTip,
			"platform_fee":  platformFee,
			"net_amount":    netAmount,
		}
		if err := s.repo.UpdateFields(existing.ID, fields); err != nil {
			return nil, err
		}
		updated, err := s.repo.FindByID(existing.ID)
		if err != nil {
			return nil, err
		}
		return toRiderSettlementInfo(updated), nil
	}

	settlement := &model.RiderSettlement{
		RiderID:     rider.ID,
		Period:      period,
		TotalOrders: int(totalOrders),
		TotalAmount: totalEarnings,
		TotalFee:    totalFee,
		TotalTip:    totalTip,
		PlatformFee: platformFee,
		NetAmount:   netAmount,
		Status:      model.RiderSettlementStatusPending,
	}
	settlement.RegionID = regionID
	if err := s.repo.Create(settlement); err != nil {
		return nil, err
	}
	return toRiderSettlementInfo(settlement), nil
}

// GetByID 获取结算单详情
func (s *riderSettlementService) GetByID(id uint) (*dto.RiderSettlementInfo, error) {
	settlement, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderSettlementNotFound
		}
		return nil, err
	}
	return toRiderSettlementInfo(settlement), nil
}

// List 结算单列表（管理端）
func (s *riderSettlementService) List(req *dto.RiderSettlementListRequest) (*utils.Pagination, []dto.RiderSettlementInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RiderSettlementListOptions{
		RiderID:  req.RiderID,
		Period:   req.Period,
		Status:   req.Status,
		RegionID: req.RegionID,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RiderSettlementInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRiderSettlementInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按当前用户列出结算单
func (s *riderSettlementService) ListByUser(userID uint, req *dto.RiderSettlementListRequest) (*utils.Pagination, []dto.RiderSettlementInfo, error) {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrRiderNotFound
		}
		return nil, nil, err
	}
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListByRider(rider.ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RiderSettlementInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRiderSettlementInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Audit 结算单审核（M 端）
func (s *riderSettlementService) Audit(id uint, req *dto.RiderSettlementAuditRequest) error {
	settlement, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderSettlementNotFound
		}
		return err
	}
	if settlement.Status != model.RiderSettlementStatusPending {
		return ErrRiderSettlementAudited
	}
	fields := map[string]interface{}{
		"status": req.Status,
	}
	if req.AuditReason != "" {
		fields["audit_reason"] = req.AuditReason
	}
	if req.Status == model.RiderSettlementStatusSettled {
		now := time.Now()
		fields["settled_at"] = &now
	} else if req.Status == model.RiderSettlementStatusWithdrawn {
		now := time.Now()
		fields["withdrawn_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// Withdraw 提现申请（骑手 C 端）
func (s *riderSettlementService) Withdraw(userID uint, req *dto.RiderSettlementWithdrawRequest) error {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderNotFound
		}
		return err
	}
	settlement, err := s.repo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRiderSettlementNotFound
		}
		return err
	}
	if settlement.RiderID != rider.ID {
		return errors.New("无权操作他人结算单")
	}
	if settlement.Status != model.RiderSettlementStatusSettled {
		return ErrDeliveryStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":       model.RiderSettlementStatusWithdrawn,
		"withdrawn_at": &now,
	}
	return s.repo.UpdateFields(req.ID, fields)
}

// Stats 结算统计
func (s *riderSettlementService) Stats(userID uint) (*dto.RiderSettlementStatsResponse, error) {
	rider, err := s.riderRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRiderNotFound
		}
		return nil, err
	}
	resp := &dto.RiderSettlementStatsResponse{}
	// 待结算
	if amt, cnt, err := s.repo.SumByStatus(rider.ID, model.RiderSettlementStatusPending); err == nil {
		resp.PendingCount = cnt
		resp.PendingAmount = amt
	}
	// 已结算
	if amt, cnt, err := s.repo.SumByStatus(rider.ID, model.RiderSettlementStatusSettled); err == nil {
		resp.SettledCount = cnt
		resp.SettledAmount = amt
	}
	// 已提现
	if amt, cnt, err := s.repo.SumByStatus(rider.ID, model.RiderSettlementStatusWithdrawn); err == nil {
		resp.WithdrawnCount = cnt
		resp.WithdrawnAmount = amt
	}
	resp.TotalSettlements = resp.PendingCount + resp.SettledCount + resp.WithdrawnCount
	resp.TotalNetAmount = resp.PendingAmount + resp.SettledAmount + resp.WithdrawnAmount
	return resp, nil
}
