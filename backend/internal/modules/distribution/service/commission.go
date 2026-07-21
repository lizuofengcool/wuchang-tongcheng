// Package service 分销合伙人中台业务逻辑层 - 佣金记录
// 职责：计算佣金 / 二级分销 / 结算 / 取消
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/distribution/dto"
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/modules/distribution/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCommissionNotFound      = errors.New("佣金记录不存在")
	ErrCommissionStatusInvalid = errors.New("佣金记录状态不允许此操作")
	ErrCommissionAmountInvalid = errors.New("佣金金额无效")
)

// CommissionService 佣金业务接口
type CommissionService interface {
	// 创建佣金记录（一级 + 自动处理二级分销）
	Create(req *dto.CommissionCreateRequest) (*dto.CommissionInfo, error)
	GetByID(id uint) (*dto.CommissionInfo, error)
	List(req *dto.CommissionListRequest) (*utils.Pagination, []dto.CommissionInfo, error)
	ListByPartner(partnerID uint, page, pageSize int) (*utils.Pagination, []dto.CommissionInfo, error)
	ListPending(page, pageSize int) (*utils.Pagination, []dto.CommissionInfo, error)

	// 汇总
	Summary(partnerID uint) (*dto.CommissionSummaryResponse, error)

	// 结算
	Settle(id uint) error
	BatchSettle(ids []uint) (*dto.CommissionSettleResult, error)
	// 取消
	Cancel(id uint) error
}

type commissionService struct {
	repo        repository.CommissionRepository
	partnerRepo repository.PartnerRepository
	channelRepo repository.ChannelRepository
}

// NewCommissionService 创建佣金 service 实例
func NewCommissionService(
	repo repository.CommissionRepository,
	partnerRepo repository.PartnerRepository,
	channelRepo repository.ChannelRepository,
) CommissionService {
	return &commissionService{
		repo:        repo,
		partnerRepo: partnerRepo,
		channelRepo: channelRepo,
	}
}

// commissionLevelText 级别文本
func commissionLevelText(level int) string {
	switch level {
	case model.CommissionLevelFirst:
		return "一级分销"
	case model.CommissionLevelSecond:
		return "二级分销"
	}
	return ""
}

// commissionStatusText 状态文本
func commissionStatusText(status int) string {
	switch status {
	case model.CommissionStatusPending:
		return "待结算"
	case model.CommissionStatusSettled:
		return "已结算"
	case model.CommissionStatusCanceled:
		return "已取消"
	}
	return ""
}

// toCommissionInfo model -> dto
func toCommissionInfo(c *model.Commission) *dto.CommissionInfo {
	return &dto.CommissionInfo{
		ID:                c.ID,
		PartnerID:         c.PartnerID,
		OrderID:           c.OrderID,
		ChannelID:         c.ChannelID,
		OrderAmount:       c.OrderAmount,
		CommissionAmount:  c.CommissionAmount,
		CommissionRate:    c.CommissionRate,
		Level:             c.Level,
		LevelText:         commissionLevelText(c.Level),
		Status:            c.Status,
		StatusText:        commissionStatusText(c.Status),
		SettledAt:         c.SettledAt,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// getPartnerRate 获取合伙人佣金率
func getPartnerRate(partnerRepo repository.PartnerRepository, partnerID uint) (float64, error) {
	p, err := partnerRepo.FindByID(partnerID)
	if err != nil {
		return 0, err
	}
	if p.CommissionRate > 0 {
		return p.CommissionRate, nil
	}
	return 0, nil
}

// Create 创建佣金记录（含二级分销自动触发）
func (s *commissionService) Create(req *dto.CommissionCreateRequest) (*dto.CommissionInfo, error) {
	if req.OrderAmount < 0 {
		return nil, ErrCommissionAmountInvalid
	}
	if req.PartnerID == 0 || req.OrderID == 0 {
		return nil, errors.New("合伙人和订单 ID 不能为空")
	}

	level := req.Level
	if level == 0 {
		level = model.CommissionLevelFirst
	}

	// 获取合伙人佣金率
	rate, err := getPartnerRate(s.partnerRepo, req.PartnerID)
	if err != nil {
		return nil, err
	}
	amount := req.OrderAmount * rate

	c := &model.Commission{
		PartnerID:        req.PartnerID,
		OrderID:          req.OrderID,
		ChannelID:        req.ChannelID,
		OrderAmount:      req.OrderAmount,
		CommissionAmount: amount,
		CommissionRate:   rate,
		Level:            level,
		Status:           model.CommissionStatusPending,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}

	// 一级分销时，自动处理二级分销（给上级合伙人按二级比例计算）
	if level == model.CommissionLevelFirst {
		if err := s.createSecondLevel(req); err != nil {
			// 二级分销失败不影响一级，仅忽略
			_ = err
		}
	}

	// 累加合伙人待结算佣金
	_ = s.partnerRepo.UpdateFields(req.PartnerID, map[string]interface{}{
		"pending_commission": gorm.Expr("pending_commission + ?", amount),
		"total_commission":   gorm.Expr("total_commission + ?", amount),
	})

	// 累加渠道佣金统计
	if req.ChannelID > 0 {
		_ = s.channelRepo.AddCommission(req.ChannelID, amount)
	}

	return toCommissionInfo(c), nil
}

// createSecondLevel 创建二级分销佣金（给上级）
// 二级分销佣金率 = 一级佣金率 * 0.5（简化策略，可按需调整）
func (s *commissionService) createSecondLevel(req *dto.CommissionCreateRequest) error {
	p, err := s.partnerRepo.FindByID(req.PartnerID)
	if err != nil {
		return err
	}
	if p.ParentID == 0 {
		return nil
	}
	parent, err := s.partnerRepo.FindByID(p.ParentID)
	if err != nil {
		return err
	}
	if parent.Status != model.PartnerStatusActive {
		return nil
	}
	parentRate, err := getPartnerRate(s.partnerRepo, parent.ID)
	if err != nil {
		return err
	}
	if parentRate <= 0 {
		return nil
	}
	// 二级佣金率 = 上级佣金率 * 0.5
	secondRate := parentRate * 0.5
	amount := req.OrderAmount * secondRate

	c := &model.Commission{
		PartnerID:        parent.ID,
		OrderID:          req.OrderID,
		ChannelID:        0,
		OrderAmount:      req.OrderAmount,
		CommissionAmount: amount,
		CommissionRate:   secondRate,
		Level:            model.CommissionLevelSecond,
		Status:           model.CommissionStatusPending,
	}
	if err := s.repo.Create(c); err != nil {
		return err
	}
	_ = s.partnerRepo.UpdateFields(parent.ID, map[string]interface{}{
		"pending_commission": gorm.Expr("pending_commission + ?", amount),
		"total_commission":   gorm.Expr("total_commission + ?", amount),
	})
	return nil
}

// GetByID 详情
func (s *commissionService) GetByID(id uint) (*dto.CommissionInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommissionNotFound
		}
		return nil, err
	}
	return toCommissionInfo(c), nil
}

// List 列表
func (s *commissionService) List(req *dto.CommissionListRequest) (*utils.Pagination, []dto.CommissionInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CommissionListOptions{
		PartnerID: req.PartnerID,
		OrderID:   req.OrderID,
		Level:     req.Level,
		Status:    req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CommissionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCommissionInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByPartner 按合伙人查询
func (s *commissionService) ListByPartner(partnerID uint, page, pageSize int) (*utils.Pagination, []dto.CommissionInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPartner(partnerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CommissionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCommissionInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListPending 待结算列表
func (s *commissionService) ListPending(page, pageSize int) (*utils.Pagination, []dto.CommissionInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListPending(pagination, 0)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.CommissionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toCommissionInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Summary 汇总
func (s *commissionService) Summary(partnerID uint) (*dto.CommissionSummaryResponse, error) {
	total, settled, pending, canceled, count, err := s.repo.SummaryByPartner(partnerID)
	if err != nil {
		return nil, err
	}
	return &dto.CommissionSummaryResponse{
		PartnerID:         partnerID,
		TotalCommission:   total,
		SettledCommission: settled,
		PendingCommission: pending,
		CanceledCommission: canceled,
		TotalCount:        count,
	}, nil
}

// Settle 结算单条
func (s *commissionService) Settle(id uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommissionNotFound
		}
		return err
	}
	if c.Status != model.CommissionStatusPending {
		return ErrCommissionStatusInvalid
	}
	now := time.Now()
	if err := s.repo.UpdateFields(id, map[string]interface{}{
		"status":     model.CommissionStatusSettled,
		"settled_at": &now,
	}); err != nil {
		return err
	}
	// 同步合伙人佣金：待结算减少，已结算增加
	_ = s.partnerRepo.UpdateFields(c.PartnerID, map[string]interface{}{
		"pending_commission":  gorm.Expr("pending_commission - ?", c.CommissionAmount),
		"settled_commission":  gorm.Expr("settled_commission + ?", c.CommissionAmount),
	})
	return nil
}

// BatchSettle 批量结算
func (s *commissionService) BatchSettle(ids []uint) (*dto.CommissionSettleResult, error) {
	result := &dto.CommissionSettleResult{Total: len(ids)}
	failedIDs := make([]uint, 0)
	for _, id := range ids {
		if err := s.Settle(id); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			result.Success++
		}
	}
	result.Failed = len(failedIDs)
	result.FailedIDs = failedIDs
	return result, nil
}

// Cancel 取消佣金
func (s *commissionService) Cancel(id uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommissionNotFound
		}
		return err
	}
	if c.Status == model.CommissionStatusCanceled || c.Status == model.CommissionStatusSettled {
		return ErrCommissionStatusInvalid
	}
	if err := s.repo.UpdateFields(id, map[string]interface{}{
		"status": model.CommissionStatusCanceled,
	}); err != nil {
		return err
	}
	// 同步合伙人待结算佣金扣减
	_ = s.partnerRepo.UpdateFields(c.PartnerID, map[string]interface{}{
		"pending_commission": gorm.Expr("pending_commission - ?", c.CommissionAmount),
		"total_commission":   gorm.Expr("total_commission - ?", c.CommissionAmount),
	})
	return nil
}
