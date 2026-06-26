// Package repository 地区数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/region/model"

	"gorm.io/gorm"
)

// RegionRepository 地区仓储接口
type RegionRepository interface {
	Create(region *model.Region) error
	FindByID(id uint) (*model.Region, error)
	FindByCode(code string) (*model.Region, error)
	FindByParentID(parentID uint) ([]model.Region, error)
	FindAll() ([]model.Region, error)
	Update(region *model.Region) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
}

type regionRepository struct {
	db *gorm.DB
}

// NewRegionRepository 创建地区仓储
func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) Create(region *model.Region) error {
	return r.db.Create(region).Error
}

func (r *regionRepository) FindByID(id uint) (*model.Region, error) {
	var region model.Region
	if err := r.db.First(&region, id).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *regionRepository) FindByCode(code string) (*model.Region, error) {
	var region model.Region
	if err := r.db.Where("code = ?", code).First(&region).Error; err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *regionRepository) FindByParentID(parentID uint) ([]model.Region, error) {
	var regions []model.Region
	if err := r.db.Where("parent_id = ?", parentID).Order("sort ASC, id ASC").Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *regionRepository) FindAll() ([]model.Region, error) {
	var regions []model.Region
	if err := r.db.Order("level ASC, sort ASC, id ASC").Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

func (r *regionRepository) Update(region *model.Region) error {
	return r.db.Save(region).Error
}

func (r *regionRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Region{}).Where("id = ?", id).Updates(fields).Error
}

func (r *regionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Region{}, id).Error
}
