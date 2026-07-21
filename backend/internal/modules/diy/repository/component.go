// Package repository DIY 前端页面中台数据访问层 - 组件（component 子域）
package repository

import (
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ComponentListOptions 组件列表过滤条件
type ComponentListOptions struct {
	Category string
	Status   *int
	Keyword  string
}

// ComponentRepository 组件仓储接口
type ComponentRepository interface {
	// CRUD
	Create(c *model.Component) error
	FindByID(id uint) (*model.Component, error)
	Update(c *model.Component) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 查询
	FindByCode(code string) (*model.Component, error)
	List(opts ComponentListOptions, pagination *utils.Pagination) ([]model.Component, int64, error)
	ListByCategory(category string, pagination *utils.Pagination) ([]model.Component, int64, error)
}

type componentRepository struct {
	db *gorm.DB
}

// NewComponentRepository 创建组件仓储实例
func NewComponentRepository(db *gorm.DB) ComponentRepository {
	return &componentRepository{db: db}
}

func (r *componentRepository) Create(c *model.Component) error {
	return r.db.Create(c).Error
}

func (r *componentRepository) FindByID(id uint) (*model.Component, error) {
	var c model.Component
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *componentRepository) Update(c *model.Component) error {
	return r.db.Save(c).Error
}

func (r *componentRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Component{}).Where("id = ?", id).Updates(fields).Error
}

func (r *componentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Component{}, id).Error
}

func (r *componentRepository) FindByCode(code string) (*model.Component, error) {
	var c model.Component
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *componentRepository) List(opts ComponentListOptions, pagination *utils.Pagination) ([]model.Component, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Component
	var total int64

	query := r.db.Model(&model.Component{})
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("category ASC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *componentRepository) ListByCategory(category string, pagination *utils.Pagination) ([]model.Component, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 100)
	}
	var list []model.Component
	var total int64

	query := r.db.Model(&model.Component{}).Where("category = ?", category)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
