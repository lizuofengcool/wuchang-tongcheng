// Package repository 系统设置数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/setting/model"

	"gorm.io/gorm"
)

// SettingRepository 系统设置仓储接口
type SettingRepository interface {
	Create(setting *model.Setting) error
	FindByID(id uint) (*model.Setting, error)
	FindByKey(group, key string, regionID uint) (*model.Setting, error)
	FindByGroup(group string, regionID uint) ([]model.Setting, error)
	FindAll(regionID uint) ([]model.Setting, error)
	Update(setting *model.Setting) error
	UpdateFields(id uint, fields map[string]interface{}) error
	UpdateValue(group, key string, regionID uint, value string) error
	Delete(id uint) error
}

type settingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 创建系统设置仓储
func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) Create(setting *model.Setting) error {
	return r.db.Create(setting).Error
}

func (r *settingRepository) FindByID(id uint) (*model.Setting, error) {
	var s model.Setting
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingRepository) FindByKey(group, key string, regionID uint) (*model.Setting, error) {
	var s model.Setting
	query := r.db.Where("`group` = ? AND `key` = ?", group, key)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settingRepository) FindByGroup(group string, regionID uint) ([]model.Setting, error) {
	var list []model.Setting
	query := r.db.Where("`group` = ?", group)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *settingRepository) FindAll(regionID uint) ([]model.Setting, error) {
	var list []model.Setting
	query := r.db
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("`group` ASC, sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *settingRepository) Update(setting *model.Setting) error {
	return r.db.Save(setting).Error
}

func (r *settingRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Setting{}).Where("id = ?", id).Updates(fields).Error
}

func (r *settingRepository) UpdateValue(group, key string, regionID uint, value string) error {
	query := r.db.Model(&model.Setting{}).Where("`group` = ? AND `key` = ?", group, key)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	return query.Update("value", value).Error
}

func (r *settingRepository) Delete(id uint) error {
	return r.db.Delete(&model.Setting{}, id).Error
}
