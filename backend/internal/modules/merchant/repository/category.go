// Package repository 商户中台数据访问层 - 类目
package repository

import (
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CategoryListOptions 类目列表过滤条件
type CategoryListOptions struct {
	ParentID *uint
	Status   *int
	Keyword  string
}

// CategoryRepository 类目仓储接口
type CategoryRepository interface {
	Create(c *model.Category) error
	FindByID(id uint) (*model.Category, error)
	Update(c *model.Category) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts CategoryListOptions) ([]model.Category, int64, error)
	FindByParent(parentID uint) ([]model.Category, error)
	FindAll() ([]model.Category, error)
	Tree() ([]model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建类目仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *model.Category) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(c *model.Category) error {
	return r.db.Save(c).Error
}

func (r *categoryRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Category{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *categoryRepository) List(pagination *utils.Pagination, opts CategoryListOptions) ([]model.Category, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.Category
	var total int64

	q := r.db.Model(&model.Category{})
	if opts.ParentID != nil {
		q = q.Where("parent_id = ?", *opts.ParentID)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("name ILIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) FindByParent(parentID uint) ([]model.Category, error) {
	var list []model.Category
	if err := r.db.Where("parent_id = ?", parentID).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) FindAll() ([]model.Category, error) {
	var list []model.Category
	if err := r.db.Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Tree 获取所有类目（service 层负责组装树结构）
func (r *categoryRepository) Tree() ([]model.Category, error) {
	return r.FindAll()
}
