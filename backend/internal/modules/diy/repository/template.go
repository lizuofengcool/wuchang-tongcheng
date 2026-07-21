// Package repository DIY 前端页面中台数据访问层 - 模板（template 子域）
package repository

import (
	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TemplateListOptions 模板列表过滤条件
type TemplateListOptions struct {
	Category string
	Status   *int
	Keyword  string
}

// TemplateRepository 模板仓储接口
type TemplateRepository interface {
	// CRUD
	Create(t *model.Template) error
	FindByID(id uint) (*model.Template, error)
	Update(t *model.Template) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 查询
	List(opts TemplateListOptions, pagination *utils.Pagination) ([]model.Template, int64, error)
	ListByCategory(category string, pagination *utils.Pagination) ([]model.Template, int64, error)
}

type templateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository 创建模板仓储实例
func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(t *model.Template) error {
	return r.db.Create(t).Error
}

func (r *templateRepository) FindByID(id uint) (*model.Template, error) {
	var t model.Template
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *templateRepository) Update(t *model.Template) error {
	return r.db.Save(t).Error
}

func (r *templateRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Template{}).Where("id = ?", id).Updates(fields).Error
}

func (r *templateRepository) Delete(id uint) error {
	return r.db.Delete(&model.Template{}, id).Error
}

func (r *templateRepository) List(opts TemplateListOptions, pagination *utils.Pagination) ([]model.Template, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Template
	var total int64

	query := r.db.Model(&model.Template{})
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("updated_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *templateRepository) ListByCategory(category string, pagination *utils.Pagination) ([]model.Template, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 100)
	}
	var list []model.Template
	var total int64

	query := r.db.Model(&model.Template{}).Where("category = ?", category)
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
