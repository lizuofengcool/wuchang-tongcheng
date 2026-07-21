// Package repository 分销合伙人中台数据访问层 - 提现记录
package repository

import (
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// WithdrawalListOptions 提现列表过滤条件
type WithdrawalListOptions struct {
	PartnerID uint
	Status    *int
}

// WithdrawalRepository 提现仓储接口
type WithdrawalRepository interface {
	Create(w *model.Withdrawal) error
	FindByID(id uint) (*model.Withdrawal, error)
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts WithdrawalListOptions) ([]model.Withdrawal, int64, error)
	ListByPartner(partnerID uint, pagination *utils.Pagination) ([]model.Withdrawal, int64, error)
	ListPending(pagination *utils.Pagination) ([]model.Withdrawal, int64, error)

	// 汇总
	SumPendingByPartner(partnerID uint) (float64, error) // 申请中+已审核的金额合计（冻结金额）
	SumPaidByPartner(partnerID uint) (float64, error)    // 已打款金额合计
}

type withdrawalRepository struct {
	db *gorm.DB
}

// NewWithdrawalRepository 创建提现仓储实例
func NewWithdrawalRepository(db *gorm.DB) WithdrawalRepository {
	return &withdrawalRepository{db: db}
}

func (r *withdrawalRepository) Create(w *model.Withdrawal) error {
	return r.db.Create(w).Error
}

func (r *withdrawalRepository) FindByID(id uint) (*model.Withdrawal, error) {
	var w model.Withdrawal
	if err := r.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *withdrawalRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Withdrawal{}).Where("id = ?", id).Updates(fields).Error
}

func (r *withdrawalRepository) Delete(id uint) error {
	return r.db.Delete(&model.Withdrawal{}, id).Error
}

func (r *withdrawalRepository) List(pagination *utils.Pagination, opts WithdrawalListOptions) ([]model.Withdrawal, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Withdrawal
	var total int64

	query := r.db.Model(&model.Withdrawal{})
	if opts.PartnerID > 0 {
		query = query.Where("partner_id = ?", opts.PartnerID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *withdrawalRepository) ListByPartner(partnerID uint, pagination *utils.Pagination) ([]model.Withdrawal, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Withdrawal
	var total int64
	query := r.db.Model(&model.Withdrawal{}).Where("partner_id = ?", partnerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *withdrawalRepository) ListPending(pagination *utils.Pagination) ([]model.Withdrawal, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Withdrawal
	var total int64
	query := r.db.Model(&model.Withdrawal{}).
		Where("status IN ?", []int{model.WithdrawalStatusPending, model.WithdrawalStatusAudited})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *withdrawalRepository) SumPendingByPartner(partnerID uint) (float64, error) {
	var sum float64
	err := r.db.Model(&model.Withdrawal{}).
		Where("partner_id = ? AND status IN ?", partnerID, []int{model.WithdrawalStatusPending, model.WithdrawalStatusAudited}).
		Select("COALESCE(SUM(amount),0)").
		Scan(&sum).Error
	if err != nil {
		return 0, err
	}
	return sum, nil
}

func (r *withdrawalRepository) SumPaidByPartner(partnerID uint) (float64, error) {
	var sum float64
	err := r.db.Model(&model.Withdrawal{}).
		Where("partner_id = ? AND status = ?", partnerID, model.WithdrawalStatusPaid).
		Select("COALESCE(SUM(amount),0)").
		Scan(&sum).Error
	if err != nil {
		return 0, err
	}
	return sum, nil
}
