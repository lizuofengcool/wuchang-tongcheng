// Package repository 分类信息数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/category/model"

	"gorm.io/gorm"
)

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	Create(category *model.Category) error
	FindByID(id uint) (*model.Category, error)
	FindByParentID(parentID uint, regionID uint) ([]model.Category, error)
	FindByRegionID(regionID uint) ([]model.Category, error)
	Update(category *model.Category) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.Category, error) {
	var category model.Category
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindByParentID(parentID uint, regionID uint) ([]model.Category, error) {
	var categories []model.Category
	query := r.db.Where("parent_id = ?", parentID)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("sort DESC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) FindByRegionID(regionID uint) ([]model.Category, error) {
	var categories []model.Category
	query := r.db
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("level ASC, sort DESC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Category{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}
