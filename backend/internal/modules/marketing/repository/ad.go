// Package repository 营销活动中台数据访问层 - 广告位（ad 子域）
package repository

import (
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AdPositionListQuery 广告位列表查询
type AdPositionListQuery struct {
	PositionCode string
	Status       *int
	Keyword      string
}

// AdRepository 广告位仓储接口
type AdRepository interface {
	Create(a *model.AdPosition) error
	FindByID(id uint) (*model.AdPosition, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query AdPositionListQuery, pagination *utils.Pagination) ([]model.AdPosition, int64, error)
	FindByPositionCode(regionID uint, positionCode string, pagination *utils.Pagination) ([]model.AdPosition, int64, error)
}

type adRepository struct {
	db *gorm.DB
}

// NewAdRepository 创建广告位仓储实例
func NewAdRepository(db *gorm.DB) AdRepository {
	return &adRepository{db: db}
}

func (r *adRepository) Create(a *model.AdPosition) error {
	return r.db.Create(a).Error
}

func (r *adRepository) FindByID(id uint) (*model.AdPosition, error) {
	var a model.AdPosition
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *adRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.AdPosition{}).Where("id = ?", id).Updates(fields).Error
}

func (r *adRepository) Delete(id uint) error {
	return r.db.Delete(&model.AdPosition{}, id).Error
}

func (r *adRepository) List(regionID uint, query AdPositionListQuery, pagination *utils.Pagination) ([]model.AdPosition, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.AdPosition
	var total int64

	q := r.db.Model(&model.AdPosition{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.PositionCode != "" {
		q = q.Where("position_code = ?", query.PositionCode)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ?", like)
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

func (r *adRepository) FindByPositionCode(regionID uint, positionCode string, pagination *utils.Pagination) ([]model.AdPosition, int64, error) {
	return r.List(regionID, AdPositionListQuery{
		PositionCode: positionCode,
		Status:       intPtrMarketing(model.AdStatusEnabled),
	}, pagination)
}
