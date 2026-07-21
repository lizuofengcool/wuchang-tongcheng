// Package service 分销合伙人中台业务逻辑层 - 提现
// 职责：申请提现 / 审核 / 打款 / 拒绝 + 余额校验
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
	ErrWithdrawalNotFound      = errors.New("提现记录不存在")
	ErrWithdrawalStatusInvalid = errors.New("提现状态不允许此操作")
	ErrWithdrawalAmountInvalid = errors.New("提现金额无效")
	ErrWithdrawalInsufficient  = errors.New("可提现余额不足")
)

// WithdrawalService 提现业务接口
type WithdrawalService interface {
	Apply(partnerID uint, req *dto.WithdrawalApplyRequest) (*dto.WithdrawalInfo, error)
	GetByID(id uint) (*dto.WithdrawalInfo, error)
	List(req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error)
	ListByPartner(partnerID uint, page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error)
	ListPending(page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error)

	// 管理端
	Audit(id uint, req *dto.WithdrawalAuditRequest) error
	Pay(id uint, req *dto.WithdrawalPayRequest) error
	Reject(id uint, reason string) error
}

type withdrawalService struct {
	repo        repository.WithdrawalRepository
	partnerRepo repository.PartnerRepository
}

// NewWithdrawalService 创建提现 service 实例
func NewWithdrawalService(repo repository.WithdrawalRepository, partnerRepo repository.PartnerRepository) WithdrawalService {
	return &withdrawalService{repo: repo, partnerRepo: partnerRepo}
}

// withdrawalStatusText 状态文本
func withdrawalStatusText(status int) string {
	switch status {
	case model.WithdrawalStatusPending:
		return "申请中"
	case model.WithdrawalStatusAudited:
		return "已审核"
	case model.WithdrawalStatusPaid:
		return "已打款"
	case model.WithdrawalStatusRejected:
		return "已拒绝"
	}
	return ""
}

// toWithdrawalInfo model -> dto
func toWithdrawalInfo(w *model.Withdrawal) *dto.WithdrawalInfo {
	info := &dto.WithdrawalInfo{
		ID:          w.ID,
		PartnerID:   w.PartnerID,
		Amount:      w.Amount,
		Status:      w.Status,
		StatusText:  withdrawalStatusText(w.Status),
		AuditReason: w.AuditReason,
		AuditedAt:   w.AuditedAt,
		PaidAt:      w.PaidAt,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
	if w.BankInfo != nil {
		info.BankInfo = w.BankInfo
	}
	return info
}

// Apply 申请提现
// 校验：合伙人存在且正常 / 金额有效 / 可提现余额充足
// 可提现余额 = 累计佣金 - 已结算佣金 - 申请中+已审核的提现合计
func (s *withdrawalService) Apply(partnerID uint, req *dto.WithdrawalApplyRequest) (*dto.WithdrawalInfo, error) {
	if partnerID == 0 {
		return nil, ErrPartnerNotFound
	}
	if req.Amount <= 0 {
		return nil, ErrWithdrawalAmountInvalid
	}

	p, err := s.partnerRepo.FindByID(partnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPartnerNotFound
		}
		return nil, err
	}
	if p.Status != model.PartnerStatusActive {
		return nil, ErrPartnerStatusInvalid
	}

	// 计算可提现余额 = 累计佣金 - 已结算佣金 - 在途提现（申请中+已审核）
	frozen, err := s.repo.SumPendingByPartner(partnerID)
	if err != nil {
		return nil, err
	}
	available := p.TotalCommission - p.SettledCommission - frozen
	if req.Amount > available {
		return nil, ErrWithdrawalInsufficient
	}

	// 落库银行信息
	var bankInfo model.JSONB
	if req.BankInfo != nil {
		if b, err := model.FromJSON(req.BankInfo); err == nil {
			bankInfo = b
		}
	}

	w := &model.Withdrawal{
		PartnerID: partnerID,
		Amount:    req.Amount,
		Status:    model.WithdrawalStatusPending,
		BankInfo:  bankInfo,
	}
	if err := s.repo.Create(w); err != nil {
		return nil, err
	}
	return toWithdrawalInfo(w), nil
}

// GetByID 详情
func (s *withdrawalService) GetByID(id uint) (*dto.WithdrawalInfo, error) {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWithdrawalNotFound
		}
		return nil, err
	}
	return toWithdrawalInfo(w), nil
}

// List 列表（管理端可按合伙人/状态过滤）
func (s *withdrawalService) List(req *dto.WithdrawalListRequest) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.WithdrawalListOptions{
		PartnerID: req.PartnerID,
		Status:    req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toWithdrawalInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByPartner 我的提现
func (s *withdrawalService) ListByPartner(partnerID uint, page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPartner(partnerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toWithdrawalInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListPending 待审核列表（申请中+已审核）
func (s *withdrawalService) ListPending(page, pageSize int) (*utils.Pagination, []dto.WithdrawalInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListPending(pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.WithdrawalInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toWithdrawalInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Audit 审核（通过/拒绝）
// 通过：状态 0 -> 1；拒绝：状态 0 -> 3
func (s *withdrawalService) Audit(id uint, req *dto.WithdrawalAuditRequest) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawalNotFound
		}
		return err
	}
	if w.Status != model.WithdrawalStatusPending {
		return ErrWithdrawalStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":       req.Status,
		"audit_reason": req.Reason,
		"audited_at":   &now,
	}
	return s.repo.UpdateFields(id, fields)
}

// Pay 打款确认（状态 1 -> 2）
// 同步更新合伙人已结算佣金
func (s *withdrawalService) Pay(id uint, req *dto.WithdrawalPayRequest) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawalNotFound
		}
		return err
	}
	if w.Status != model.WithdrawalStatusAudited {
		return ErrWithdrawalStatusInvalid
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":      model.WithdrawalStatusPaid,
		"paid_at":     &now,
		"audit_reason": req.Reason,
	}
	if err := s.repo.UpdateFields(id, fields); err != nil {
		return err
	}

	// 同步更新合伙人已结算佣金（累加）
	p, err := s.partnerRepo.FindByID(w.PartnerID)
	if err == nil && p != nil {
		newSettled := p.SettledCommission + w.Amount
		_ = s.partnerRepo.UpdateFields(p.ID, map[string]interface{}{
			"settled_commission": newSettled,
		})
	}
	return nil
}

// Reject 拒绝提现（状态 0 或 1 -> 3）
func (s *withdrawalService) Reject(id uint, reason string) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWithdrawalNotFound
		}
		return err
	}
	if w.Status == model.WithdrawalStatusPaid || w.Status == model.WithdrawalStatusRejected {
		return ErrWithdrawalStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":       model.WithdrawalStatusRejected,
		"audit_reason": reason,
		"audited_at":   &now,
	}
	return s.repo.UpdateFields(id, fields)
}
