// Package repository 同城商城数据访问层 - 商品分类
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"

	"gorm.io/gorm"
)

// CategoryRepository 商品分类仓储接口
type CategoryRepository interface {
	Create(c *model.Category) error
	FindByID(id uint) (*model.Category, error)
	Update(c *model.Category) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(opts CategoryListOptions) ([]model.Category, int64, error)
	ListTree(regionID uint) ([]model.Category, error)
	ListByParent(parentID uint) ([]model.Category, error)
	ListEnabled() ([]model.Category, error)

	// 计数器
	IncrProductCount(id uint) error
	DecrProductCount(id uint) error
}

// CategoryListOptions 分类列表过滤条件
type CategoryListOptions struct {
	ParentID *uint
	Level    *int
	Status   *int
	Keyword  string
	IsShow   *bool
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储实例
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

func (r *categoryRepository) List(opts CategoryListOptions) ([]model.Category, int64, error) {
	var list []model.Category
	var total int64

	query := r.db.Model(&model.Category{})
	if opts.ParentID != nil {
		query = query.Where("parent_id = ?", *opts.ParentID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.IsShow != nil {
		query = query.Where("is_show = ?", *opts.IsShow)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR keywords ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) ListTree(regionID uint) ([]model.Category, error) {
	var list []model.Category
	query := r.db.Model(&model.Category{}).Where("status = ?", model.CategoryStatusEnabled)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByParent(parentID uint) ([]model.Category, error) {
	var list []model.Category
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CategoryStatusEnabled).
		Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListEnabled() ([]model.Category, error) {
	var list []model.Category
	if err := r.db.Where("status = ?", model.CategoryStatusEnabled).
		Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) IncrProductCount(id uint) error {
	return r.db.Model(&model.Category{}).Where("id = ?", id).
		UpdateColumn("product_count", gorm.Expr("product_count + 1")).Error
}

func (r *categoryRepository) DecrProductCount(id uint) error {
	return r.db.Model(&model.Category{}).Where("id = ? AND product_count > 0", id).
		UpdateColumn("product_count", gorm.Expr("product_count - 1")).Error
}
