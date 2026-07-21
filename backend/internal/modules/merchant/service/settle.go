// Package service 商户中台业务逻辑层 - 结算
// 对标美团/有赞商户结算系统
// 按月生成结算单：总金额 / 平台佣金 / 商户应得
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrSettleNotFound       = errors.New("结算单不存在")
	ErrSettleExists         = errors.New("该周期结算单已存在")
	ErrSettleStatusInvalid  = errors.New("结算单状态不允许此操作")
	ErrSettleAmountInvalid  = errors.New("结算金额无效")
	ErrSettleRateInvalid    = errors.New("平台佣金比例无效")
)

// SettleService 结算业务接口
type SettleService interface {
	// 生成结算单（M 端）
	Generate(req *dto.SettleGenerateRequest) (*dto.SettleInfo, error)
	GetByID(id uint) (*dto.SettleInfo, error)
	List(req *dto.SettleListRequest) (*utils.Pagination, []dto.SettleInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.SettleInfo, error)

	// 提现
	Withdraw(id uint) error
	AuditWithdraw(id uint, req *dto.SettleAuditRequest) error

	// 汇总
	SummaryByShop(shopID uint) (*dto.SettleSummary, error)
	SummaryByPeriod(period string) (*dto.SettleSummary, error)
}

type settleService struct {
	repo repository.SettleRepository
}

// NewSettleService 创建结算 service 实例
func NewSettleService(repo repository.SettleRepository) SettleService {
	return &settleService{repo: repo}
}

// settleStatusText 结算状态文本
func settleStatusText(status int) string {
	switch status {
	case model.SettleStatusPending:
		return "待结算"
	case model.SettleStatusSettled:
		return "已结算"
	case model.SettleStatusPaid:
		return "已提现"
	case model.SettleStatusCanceled:
		return "已撤销"
	}
	return ""
}

// Generate 生成结算单
func (s *settleService) Generate(req *dto.SettleGenerateRequest) (*dto.SettleInfo, error) {
	if req.TotalAmount < 0 {
		return nil, ErrSettleAmountInvalid
	}
	rate := req.PlatformRate
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	// 检查该周期是否已存在
	if existing, err := s.repo.FindByPeriod(req.ShopID, req.Period); err == nil && existing != nil {
		return nil, ErrSettleExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	platformFee := req.TotalAmount * rate
	shopAmount := req.TotalAmount - platformFee

	settle := &model.Settle{
		ShopID:       req.ShopID,
		Period:       req.Period,
		TotalAmount:  req.TotalAmount,
		PlatformFee:  platformFee,
		ShopAmount:   shopAmount,
		Status:       model.SettleStatusPending,
	}
	if err := s.repo.Create(settle); err != nil {
		return nil, err
	}
	return s.toInfo(settle), nil
}

// GetByID 结算单详情
func (s *settleService) GetByID(id uint) (*dto.SettleInfo, error) {
	settle, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettleNotFound
		}
		return nil, err
	}
	return s.toInfo(settle), nil
}

// List 结算列表
func (s *settleService) List(req *dto.SettleListRequest) (*utils.Pagination, []dto.SettleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.SettleListOptions{
		ShopID: req.ShopID,
		Period: req.Period,
		Status: req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.SettleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByShop 按店铺查询
func (s *settleService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.SettleInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.FindByShopID(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.SettleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// Withdraw 提现申请
func (s *settleService) Withdraw(id uint) error {
	settle, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSettleNotFound
		}
		return err
	}
	if settle.Status != model.SettleStatusPending && settle.Status != model.SettleStatusSettled {
		return ErrSettleStatusInvalid
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":     model.SettleStatusSettled,
		"settled_at": now,
	})
}

// AuditWithdraw 提现审核
func (s *settleService) AuditWithdraw(id uint, req *dto.SettleAuditRequest) error {
	settle, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSettleNotFound
		}
		return err
	}
	if settle.Status != model.SettleStatusSettled && settle.Status != model.SettleStatusPending {
		return ErrSettleStatusInvalid
	}
	fields := map[string]interface{}{
		"status":        req.Status,
		"audit_reason":  req.Reason,
	}
	if req.Status == model.SettleStatusPaid {
		fields["settled_at"] = time.Now()
	}
	return s.repo.UpdateFields(id, fields)
}

// SummaryByShop 按店铺汇总
func (s *settleService) SummaryByShop(shopID uint) (*dto.SettleSummary, error) {
	totalAmount, platformFee, shopAmount, count, err := s.repo.SummaryByShop(shopID)
	if err != nil {
		return nil, err
	}
	return &dto.SettleSummary{
		ShopID:      shopID,
		TotalAmount: totalAmount,
		PlatformFee: platformFee,
		ShopAmount:  shopAmount,
		Count:       count,
	}, nil
}

// SummaryByPeriod 按周期汇总
func (s *settleService) SummaryByPeriod(period string) (*dto.SettleSummary, error) {
	totalAmount, platformFee, shopAmount, count, err := s.repo.SummaryByPeriod(period)
	if err != nil {
		return nil, err
	}
	return &dto.SettleSummary{
		Period:      period,
		TotalAmount: totalAmount,
		PlatformFee: platformFee,
		ShopAmount:  shopAmount,
		Count:       count,
	}, nil
}

// toInfo 模型转 DTO
func (s *settleService) toInfo(m *model.Settle) *dto.SettleInfo {
	return &dto.SettleInfo{
		ID:          m.ID,
		ShopID:      m.ShopID,
		Period:      m.Period,
		TotalAmount: m.TotalAmount,
		PlatformFee: m.PlatformFee,
		ShopAmount:  m.ShopAmount,
		Status:      m.Status,
		StatusText:  settleStatusText(m.Status),
		SettledAt:   m.SettledAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
