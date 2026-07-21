// Package repository 分销合伙人中台数据访问层 - 合伙人
package repository

import (
	"wuchang-tongcheng/internal/modules/distribution/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PartnerListOptions 合伙人列表过滤条件
type PartnerListOptions struct {
	UserID   uint
	ParentID uint
	Level    *int
	Status   *int
	RegionID uint
	Keyword  string
}

// PartnerRepository 合伙人仓储接口
type PartnerRepository interface {
	Create(p *model.Partner) error
	FindByID(id uint) (*model.Partner, error)
	FindByUserID(userID uint) (*model.Partner, error)
	Update(p *model.Partner) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts PartnerListOptions) ([]model.Partner, int64, error)
	ListByParent(parentID uint) ([]model.Partner, error)
	ListByLevel(level int, pagination *utils.Pagination) ([]model.Partner, int64, error)
	ListByRegion(regionID uint, pagination *utils.Pagination) ([]model.Partner, int64, error)
	CountByParent(parentID uint) (int64, error)
}

type partnerRepository struct {
	db *gorm.DB
}

// NewPartnerRepository 创建合伙人仓储实例
func NewPartnerRepository(db *gorm.DB) PartnerRepository {
	return &partnerRepository{db: db}
}

func (r *partnerRepository) Create(p *model.Partner) error {
	return r.db.Create(p).Error
}

func (r *partnerRepository) FindByID(id uint) (*model.Partner, error) {
	var p model.Partner
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *partnerRepository) FindByUserID(userID uint) (*model.Partner, error) {
	var p model.Partner
	if err := r.db.Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *partnerRepository) Update(p *model.Partner) error {
	return r.db.Save(p).Error
}

func (r *partnerRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Partner{}).Where("id = ?", id).Updates(fields).Error
}

func (r *partnerRepository) Delete(id uint) error {
	return r.db.Delete(&model.Partner{}, id).Error
}

func (r *partnerRepository) List(pagination *utils.Pagination, opts PartnerListOptions) ([]model.Partner, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Partner
	var total int64

	query := r.db.Model(&model.Partner{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ParentID > 0 {
		query = query.Where("parent_id = ?", opts.ParentID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.Keyword != "" {
		query = query.Where("user_name ILIKE ? ", "%"+opts.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *partnerRepository) ListByParent(parentID uint) ([]model.Partner, error) {
	var list []model.Partner
	if err := r.db.Where("parent_id = ?", parentID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *partnerRepository) ListByLevel(level int, pagination *utils.Pagination) ([]model.Partner, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Partner
	var total int64
	query := r.db.Model(&model.Partner{}).Where("level = ?", level)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *partnerRepository) ListByRegion(regionID uint, pagination *utils.Pagination) ([]model.Partner, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Partner
	var total int64
	query := r.db.Model(&model.Partner{}).Where("region_id = ?", regionID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *partnerRepository) CountByParent(parentID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Partner{}).Where("parent_id = ?", parentID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
