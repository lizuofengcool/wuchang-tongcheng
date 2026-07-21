// Package repository 分销合伙人中台数据访问层 - 佣金记录
package repository

import (
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CommissionListOptions 佣金列表过滤条件
type CommissionListOptions struct {
	PartnerID uint
	OrderID   uint
	Level     *int
	Status    *int
}

// CommissionRepository 佣金仓储接口
type CommissionRepository interface {
	Create(c *model.Commission) error
	FindByID(id uint) (*model.Commission, error)
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts CommissionListOptions) ([]model.Commission, int64, error)
	ListByPartner(partnerID uint, pagination *utils.Pagination) ([]model.Commission, int64, error)
	ListByOrder(orderID uint) ([]model.Commission, error)
	ListPending(pagination *utils.Pagination, partnerID uint) ([]model.Commission, int64, error)

	// 汇总
	SummaryByPartner(partnerID uint) (total, settled, pending, canceled float64, count int64, err error)

	// 批量结算
	BatchSettle(ids []uint) (int64, error)
}

type commissionRepository struct {
	db *gorm.DB
}

// NewCommissionRepository 创建佣金仓储实例
func NewCommissionRepository(db *gorm.DB) CommissionRepository {
	return &commissionRepository{db: db}
}

func (r *commissionRepository) Create(c *model.Commission) error {
	return r.db.Create(c).Error
}

func (r *commissionRepository) FindByID(id uint) (*model.Commission, error) {
	var c model.Commission
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *commissionRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Commission{}).Where("id = ?", id).Updates(fields).Error
}

func (r *commissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Commission{}, id).Error
}

func (r *commissionRepository) List(pagination *utils.Pagination, opts CommissionListOptions) ([]model.Commission, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Commission
	var total int64

	query := r.db.Model(&model.Commission{})
	if opts.PartnerID > 0 {
		query = query.Where("partner_id = ?", opts.PartnerID)
	}
	if opts.OrderID > 0 {
		query = query.Where("order_id = ?", opts.OrderID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
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

func (r *commissionRepository) ListByPartner(partnerID uint, pagination *utils.Pagination) ([]model.Commission, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Commission
	var total int64
	query := r.db.Model(&model.Commission{}).Where("partner_id = ?", partnerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *commissionRepository) ListByOrder(orderID uint) ([]model.Commission, error) {
	var list []model.Commission
	if err := r.db.Where("order_id = ?", orderID).
		Order("level ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *commissionRepository) ListPending(pagination *utils.Pagination, partnerID uint) ([]model.Commission, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Commission
	var total int64
	query := r.db.Model(&model.Commission{}).Where("status = ?", model.CommissionStatusPending)
	if partnerID > 0 {
		query = query.Where("partner_id = ?", partnerID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *commissionRepository) SummaryByPartner(partnerID uint) (total, settled, pending, canceled float64, count int64, err error) {
	type sumResult struct {
		Total    float64 `gorm:"column:total"`
		Settled  float64 `gorm:"column:settled"`
		Pending  float64 `gorm:"column:pending"`
		Canceled float64 `gorm:"column:canceled"`
		Count    int64   `gorm:"column:count"`
	}
	var res sumResult
	query := r.db.Model(&model.Commission{}).Where("partner_id = ?", partnerID)
	if err := query.Select(
		"COALESCE(SUM(commission_amount),0) AS total, " +
			"COALESCE(SUM(CASE WHEN status = 1 THEN commission_amount ELSE 0 END),0) AS settled, " +
			"COALESCE(SUM(CASE WHEN status = 0 THEN commission_amount ELSE 0 END),0) AS pending, " +
			"COALESCE(SUM(CASE WHEN status = 2 THEN commission_amount ELSE 0 END),0) AS canceled, " +
			"COUNT(*) AS count",
	).Scan(&res).Error; err != nil {
		return 0, 0, 0, 0, 0, err
	}
	return res.Total, res.Settled, res.Pending, res.Canceled, res.Count, nil
}

func (r *commissionRepository) BatchSettle(ids []uint) (int64, error) {
	result := r.db.Model(&model.Commission{}).
		Where("id IN ? AND status = ?", ids, model.CommissionStatusPending).
		Update("status", model.CommissionStatusSettled)
	return result.RowsAffected, result.Error
}
