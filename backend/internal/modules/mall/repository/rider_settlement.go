// Package repository 同城商城数据访问层 - 骑手结算
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RiderSettlementRepository 骑手结算仓储接口
type RiderSettlementRepository interface {
	Create(s *model.RiderSettlement) error
	FindByID(id uint) (*model.RiderSettlement, error)
	FindByRiderAndPeriod(riderID uint, period string) (*model.RiderSettlement, error)
	Update(s *model.RiderSettlement) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts RiderSettlementListOptions) ([]model.RiderSettlement, int64, error)
	ListByRider(riderID uint, pagination *utils.Pagination) ([]model.RiderSettlement, int64, error)

	// 统计
	SumByStatus(riderID uint, status int) (float64, int64, error)
}

// RiderSettlementListOptions 结算单列表过滤条件
type RiderSettlementListOptions struct {
	RiderID  uint
	Period   string
	Status   *int
	RegionID uint
}

type riderSettlementRepository struct {
	db *gorm.DB
}

// NewRiderSettlementRepository 创建骑手结算仓储实例
func NewRiderSettlementRepository(db *gorm.DB) RiderSettlementRepository {
	return &riderSettlementRepository{db: db}
}

func (r *riderSettlementRepository) Create(s *model.RiderSettlement) error {
	return r.db.Create(s).Error
}

func (r *riderSettlementRepository) FindByID(id uint) (*model.RiderSettlement, error) {
	var s model.RiderSettlement
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *riderSettlementRepository) FindByRiderAndPeriod(riderID uint, period string) (*model.RiderSettlement, error) {
	var s model.RiderSettlement
	if err := r.db.Where("rider_id = ? AND period = ?", riderID, period).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *riderSettlementRepository) Update(s *model.RiderSettlement) error {
	return r.db.Save(s).Error
}

func (r *riderSettlementRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.RiderSettlement{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riderSettlementRepository) Delete(id uint) error {
	return r.db.Delete(&model.RiderSettlement{}, id).Error
}

func (r *riderSettlementRepository) List(pagination *utils.Pagination, opts RiderSettlementListOptions) ([]model.RiderSettlement, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.RiderSettlement
	var total int64

	query := r.db.Model(&model.RiderSettlement{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.RiderID > 0 {
		query = query.Where("rider_id = ?", opts.RiderID)
	}
	if opts.Period != "" {
		query = query.Where("period = ?", opts.Period)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riderSettlementRepository) ListByRider(riderID uint, pagination *utils.Pagination) ([]model.RiderSettlement, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.RiderSettlement
	var total int64

	query := r.db.Model(&model.RiderSettlement{}).Where("rider_id = ?", riderID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("period DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// SumByStatus 统计骑手某状态的结算总净额和数量
func (r *riderSettlementRepository) SumByStatus(riderID uint, status int) (float64, int64, error) {
	var result struct {
		Total float64 `gorm:"column:total"`
		Count int64   `gorm:"column:cnt"`
	}
	query := r.db.Model(&model.RiderSettlement{}).Where("status = ?", status)
	if riderID > 0 {
		query = query.Where("rider_id = ?", riderID)
	}
	err := query.Select("COALESCE(SUM(net_amount), 0) AS total, COUNT(*) AS cnt").Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Total, result.Count, nil
}
