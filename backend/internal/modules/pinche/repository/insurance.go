// Package repository 同城拼车出行数据访问层 - 顺风车保险
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InsuranceListOptions 保险列表过滤条件
type InsuranceListOptions struct {
	PincheID  uint
	BookingID uint
	Status    *int
	PolicyNo  string
}

// InsuranceRepository 保险仓储接口
type InsuranceRepository interface {
	Create(i *model.PincheInsurance) error
	FindByID(id uint) (*model.PincheInsurance, error)
	FindByPolicyNo(policyNo string) (*model.PincheInsurance, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts InsuranceListOptions) ([]model.PincheInsurance, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheInsurance, int64, error)
	ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PincheInsurance, int64, error)

	UpdateStatus(id uint, status int) error
	UpdateClaim(id uint, amount float64, reason string) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type insuranceRepository struct {
	db *gorm.DB
}

// NewInsuranceRepository 创建保险仓储实例
func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return &insuranceRepository{db: db}
}

func (r *insuranceRepository) Create(i *model.PincheInsurance) error {
	return r.db.Create(i).Error
}

func (r *insuranceRepository) FindByID(id uint) (*model.PincheInsurance, error) {
	var i model.PincheInsurance
	if err := r.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *insuranceRepository) FindByPolicyNo(policyNo string) (*model.PincheInsurance, error) {
	var i model.PincheInsurance
	if err := r.db.Where("policy_no = ?", policyNo).First(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *insuranceRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheInsurance{}).Where("id = ?", id).Updates(fields).Error
}

func (r *insuranceRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheInsurance{}, id).Error
}

func (r *insuranceRepository) List(regionID uint, pagination *utils.Pagination, opts InsuranceListOptions) ([]model.PincheInsurance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheInsurance
	var total int64

	query := r.db.Model(&model.PincheInsurance{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.PolicyNo != "" {
		query = query.Where("policy_no ILIKE ?", "%"+opts.PolicyNo+"%")
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

func (r *insuranceRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheInsurance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheInsurance
	var total int64

	query := r.db.Model(&model.PincheInsurance{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *insuranceRepository) ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PincheInsurance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheInsurance
	var total int64

	query := r.db.Model(&model.PincheInsurance{}).Where("booking_id = ?", bookingID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *insuranceRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheInsurance{}).Where("id = ?", id).Update("status", status).Error
}

func (r *insuranceRepository) UpdateClaim(id uint, amount float64, reason string) error {
	return r.db.Model(&model.PincheInsurance{}).Where("id = ?", id).
		Updates(map[string]interface{}{"claim_amount": amount, "claim_reason": reason}).Error
}

func (r *insuranceRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheInsurance{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
