// Package repository 同城114数据访问层 - 商家分类
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CategoryRepository 商家分类仓储接口
type CategoryRepository interface {
	Create(c *model.Dh114Category) error
	FindByID(id uint) (*model.Dh114Category, error)
	FindByCode(code string) (*model.Dh114Category, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query CategoryListQuery, pagination *utils.Pagination) ([]model.Dh114Category, int64, error)
	ListByParent(parentID uint) ([]model.Dh114Category, error)
	ListByLevel(level int) ([]model.Dh114Category, error)
	ListByBusinessType(businessType string) ([]model.Dh114Category, error)
	IncrBusinessCount(id uint) error
	DecrBusinessCount(id uint) error
}

// CategoryListQuery 分类列表查询
type CategoryListQuery struct {
	ParentID     uint
	Level        int
	BusinessType string
	Status       *int
	Keyword      string
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *model.Dh114Category) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.Dh114Category, error) {
	var c model.Dh114Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) FindByCode(code string) (*model.Dh114Category, error) {
	var c model.Dh114Category
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Category{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Category{}, id).Error
}

func (r *categoryRepository) List(query CategoryListQuery, pagination *utils.Pagination) ([]model.Dh114Category, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.Dh114Category
	var total int64

	q := r.db.Model(&model.Dh114Category{})
	if query.ParentID > 0 {
		q = q.Where("parent_id = ?", query.ParentID)
	}
	if query.Level > 0 {
		q = q.Where("level = ?", query.Level)
	}
	if query.BusinessType != "" {
		q = q.Where("business_type = ?", query.BusinessType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) ListByParent(parentID uint) ([]model.Dh114Category, error) {
	var list []model.Dh114Category
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CategoryStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByLevel(level int) ([]model.Dh114Category, error) {
	var list []model.Dh114Category
	if err := r.db.Where("level = ? AND status = ?", level, model.CategoryStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByBusinessType(businessType string) ([]model.Dh114Category, error) {
	var list []model.Dh114Category
	if err := r.db.Where("business_type = ? AND status = ?", businessType, model.CategoryStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) IncrBusinessCount(id uint) error {
	return r.db.Model(&model.Dh114Category{}).Where("id = ?", id).
		UpdateColumn("business_count", gorm.Expr("business_count + 1")).Error
}

func (r *categoryRepository) DecrBusinessCount(id uint) error {
	return r.db.Model(&model.Dh114Category{}).Where("id = ? AND business_count > 0", id).
		UpdateColumn("business_count", gorm.Expr("business_count - 1")).Error
}
