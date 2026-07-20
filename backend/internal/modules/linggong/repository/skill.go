// Package repository 同城零工兼职数据访问层 - 技能标签
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SkillRepository 技能标签仓储接口
type SkillRepository interface {
	Create(s *model.LinggongSkill) error
	FindByID(id uint) (*model.LinggongSkill, error)
	FindByName(name string) (*model.LinggongSkill, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts SkillListOptions) ([]model.LinggongSkill, int64, error)
	ListByCategory(category string, pagination *utils.Pagination) ([]model.LinggongSkill, int64, error)
	ListByParent(parentID uint, pagination *utils.Pagination) ([]model.LinggongSkill, int64, error)
	ListHot(pagination *utils.Pagination) ([]model.LinggongSkill, int64, error)
	IncrWorkerCount(id uint) error
	DecrWorkerCount(id uint) error
	IncrLinggongCount(id uint) error
	DecrLinggongCount(id uint) error
}

// SkillListOptions 技能列表过滤条件
type SkillListOptions struct {
	Category string
	ParentID *uint
	Level    *int
	Status   *int
	Keyword  string
}

type skillRepository struct {
	db *gorm.DB
}

// NewSkillRepository 创建技能标签仓储实例
func NewSkillRepository(db *gorm.DB) SkillRepository {
	return &skillRepository{db: db}
}

func (r *skillRepository) Create(s *model.LinggongSkill) error {
	return r.db.Create(s).Error
}

func (r *skillRepository) FindByID(id uint) (*model.LinggongSkill, error) {
	var s model.LinggongSkill
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepository) FindByName(name string) (*model.LinggongSkill, error) {
	var s model.LinggongSkill
	if err := r.db.Where("name = ?", name).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skillRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongSkill{}).Where("id = ?", id).Updates(fields).Error
}

func (r *skillRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongSkill{}, id).Error
}

func (r *skillRepository) List(pagination *utils.Pagination, opts SkillListOptions) ([]model.LinggongSkill, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongSkill
	var total int64

	query := r.db.Model(&model.LinggongSkill{})
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.ParentID != nil {
		query = query.Where("parent_id = ?", *opts.ParentID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("sort DESC, hot_score DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skillRepository) ListByCategory(category string, pagination *utils.Pagination) ([]model.LinggongSkill, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongSkill
	var total int64
	query := r.db.Model(&model.LinggongSkill{}).Where("category = ?", category)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("sort DESC, hot_score DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skillRepository) ListByParent(parentID uint, pagination *utils.Pagination) ([]model.LinggongSkill, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongSkill
	var total int64
	query := r.db.Model(&model.LinggongSkill{}).Where("parent_id = ?", parentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("sort DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skillRepository) ListHot(pagination *utils.Pagination) ([]model.LinggongSkill, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongSkill
	var total int64
	query := r.db.Model(&model.LinggongSkill{}).Where("status = ?", 1)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("hot_score DESC, worker_count DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skillRepository) IncrWorkerCount(id uint) error {
	return r.db.Model(&model.LinggongSkill{}).Where("id = ?", id).
		UpdateColumn("worker_count", gorm.Expr("worker_count + 1")).Error
}

func (r *skillRepository) DecrWorkerCount(id uint) error {
	return r.db.Model(&model.LinggongSkill{}).Where("id = ? AND worker_count > 0", id).
		UpdateColumn("worker_count", gorm.Expr("worker_count - 1")).Error
}

func (r *skillRepository) IncrLinggongCount(id uint) error {
	return r.db.Model(&model.LinggongSkill{}).Where("id = ?", id).
		UpdateColumn("linggong_count", gorm.Expr("linggong_count + 1")).Error
}

func (r *skillRepository) DecrLinggongCount(id uint) error {
	return r.db.Model(&model.LinggongSkill{}).Where("id = ? AND linggong_count > 0", id).
		UpdateColumn("linggong_count", gorm.Expr("linggong_count - 1")).Error
}
