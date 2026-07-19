// Package repository 同城车辆买卖数据访问层 - 车险配置
// 车险为全局配置数据（无 region_id），管理后台维护
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InsuranceRepository 车险配置仓储接口
type InsuranceRepository interface {
	Create(i *model.CarInsurance) error
	FindByID(id uint) (*model.CarInsurance, error)
	FindByCode(code string) (*model.CarInsurance, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query InsuranceListQuery, pagination *utils.Pagination) ([]model.CarInsurance, int64, error)
	ListByIDs(ids []uint) ([]model.CarInsurance, error)
	ListPublished(pagination *utils.Pagination) ([]model.CarInsurance, int64, error)
	ListHot(limit int) ([]model.CarInsurance, error)
	IncrUseCount(id uint, delta int) error
}

// InsuranceListQuery 车险列表查询
type InsuranceListQuery struct {
	InsuranceType string
	Provider      string
	Status        *int
	IsHot         *bool
	Keyword       string
}

type insuranceRepository struct {
	db *gorm.DB
}

// NewInsuranceRepository 创建车险仓储实例
func NewInsuranceRepository(db *gorm.DB) InsuranceRepository {
	return &insuranceRepository{db: db}
}

func (r *insuranceRepository) Create(i *model.CarInsurance) error {
	return r.db.Create(i).Error
}

func (r *insuranceRepository) FindByID(id uint) (*model.CarInsurance, error) {
	var i model.CarInsurance
	if err := r.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *insuranceRepository) FindByCode(code string) (*model.CarInsurance, error) {
	var i model.CarInsurance
	if err := r.db.Where("code = ?", code).First(&i).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *insuranceRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarInsurance{}).Where("id = ?", id).Updates(fields).Error
}

func (r *insuranceRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarInsurance{}, id).Error
}

func (r *insuranceRepository) List(query InsuranceListQuery, pagination *utils.Pagination) ([]model.CarInsurance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarInsurance
	var total int64

	q := r.db.Model(&model.CarInsurance{})
	if query.InsuranceType != "" {
		q = q.Where("insurance_type = ?", query.InsuranceType)
	}
	if query.Provider != "" {
		q = q.Where("provider = ?", query.Provider)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.IsHot != nil {
		q = q.Where("is_hot = ?", *query.IsHot)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ? OR provider ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, use_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListByIDs 按一组 ID 查询（用于车险报价：用户选择一组险种后查询明细）
func (r *insuranceRepository) ListByIDs(ids []uint) ([]model.CarInsurance, error) {
	var list []model.CarInsurance
	if len(ids) == 0 {
		return list, nil
	}
	if err := r.db.Where("id IN ? AND status = ?", ids, model.InsurancePlanStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *insuranceRepository) ListPublished(pagination *utils.Pagination) ([]model.CarInsurance, int64, error) {
	status := model.InsurancePlanStatusPublished
	return r.List(InsuranceListQuery{Status: &status}, pagination)
}

func (r *insuranceRepository) ListHot(limit int) ([]model.CarInsurance, error) {
	var list []model.CarInsurance
	if limit <= 0 {
		limit = 10
	}
	if err := r.db.Where("status = ? AND is_hot = ?", model.InsurancePlanStatusPublished, true).
		Order("use_count DESC, sort ASC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *insuranceRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.CarInsurance{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}
