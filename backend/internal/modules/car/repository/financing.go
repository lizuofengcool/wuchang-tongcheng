// Package repository 同城车辆买卖数据访问层 - 分期方案
// 分期方案为全局配置数据（无 region_id），管理后台维护
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// FinancingRepository 分期方案仓储接口
type FinancingRepository interface {
	Create(f *model.CarFinancing) error
	FindByID(id uint) (*model.CarFinancing, error)
	FindByCode(code string) (*model.CarFinancing, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query FinancingListQuery, pagination *utils.Pagination) ([]model.CarFinancing, int64, error)
	ListPublished(pagination *utils.Pagination) ([]model.CarFinancing, int64, error)
	ListHot(limit int) ([]model.CarFinancing, error)
	IncrUseCount(id uint, delta int) error
}

// FinancingListQuery 分期方案列表查询
type FinancingListQuery struct {
	FinancingType string
	Provider      string
	Status        *int
	IsHot         *bool
	Keyword       string
}

type financingRepository struct {
	db *gorm.DB
}

// NewFinancingRepository 创建分期方案仓储实例
func NewFinancingRepository(db *gorm.DB) FinancingRepository {
	return &financingRepository{db: db}
}

func (r *financingRepository) Create(f *model.CarFinancing) error {
	return r.db.Create(f).Error
}

func (r *financingRepository) FindByID(id uint) (*model.CarFinancing, error) {
	var f model.CarFinancing
	if err := r.db.First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *financingRepository) FindByCode(code string) (*model.CarFinancing, error) {
	var f model.CarFinancing
	if err := r.db.Where("code = ?", code).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *financingRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarFinancing{}).Where("id = ?", id).Updates(fields).Error
}

func (r *financingRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarFinancing{}, id).Error
}

func (r *financingRepository) List(query FinancingListQuery, pagination *utils.Pagination) ([]model.CarFinancing, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarFinancing
	var total int64

	q := r.db.Model(&model.CarFinancing{})
	if query.FinancingType != "" {
		q = q.Where("financing_type = ?", query.FinancingType)
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

func (r *financingRepository) ListPublished(pagination *utils.Pagination) ([]model.CarFinancing, int64, error) {
	status := model.FinancingStatusPublished
	return r.List(FinancingListQuery{Status: &status}, pagination)
}

func (r *financingRepository) ListHot(limit int) ([]model.CarFinancing, error) {
	var list []model.CarFinancing
	if limit <= 0 {
		limit = 10
	}
	if err := r.db.Where("status = ? AND is_hot = ?", model.FinancingStatusPublished, true).
		Order("use_count DESC, sort ASC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *financingRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.CarFinancing{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}
