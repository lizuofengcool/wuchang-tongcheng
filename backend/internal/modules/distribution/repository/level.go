// Package repository 分销合伙人中台数据访问层 - 合伙人等级
package repository

import (
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LevelListOptions 等级列表过滤条件
type LevelListOptions struct {
	Level  *int
	Status *int
}

// LevelRepository 等级仓储接口
type LevelRepository interface {
	Create(l *model.Level) error
	FindByID(id uint) (*model.Level, error)
	FindByLevel(level int) (*model.Level, error)
	Update(l *model.Level) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts LevelListOptions) ([]model.Level, int64, error)
	ListAll() ([]model.Level, error)
}

type levelRepository struct {
	db *gorm.DB
}

// NewLevelRepository 创建等级仓储实例
func NewLevelRepository(db *gorm.DB) LevelRepository {
	return &levelRepository{db: db}
}

func (r *levelRepository) Create(l *model.Level) error {
	return r.db.Create(l).Error
}

func (r *levelRepository) FindByID(id uint) (*model.Level, error) {
	var l model.Level
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *levelRepository) FindByLevel(level int) (*model.Level, error) {
	var l model.Level
	if err := r.db.Where("level = ?", level).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *levelRepository) Update(l *model.Level) error {
	return r.db.Save(l).Error
}

func (r *levelRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Level{}).Where("id = ?", id).Updates(fields).Error
}

func (r *levelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Level{}, id).Error
}

func (r *levelRepository) List(pagination *utils.Pagination, opts LevelListOptions) ([]model.Level, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Level
	var total int64

	query := r.db.Model(&model.Level{})
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("level ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *levelRepository) ListAll() ([]model.Level, error) {
	var list []model.Level
	if err := r.db.Where("status = ?", model.LevelStatusEnabled).
		Order("level ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
