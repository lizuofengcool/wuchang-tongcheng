// Package repository 同城零工兼职数据访问层 - 推荐
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RecommendationRepository 推荐仓储接口
type RecommendationRepository interface {
	Create(r *model.LinggongRecommendation) error
	FindByID(id uint) (*model.LinggongRecommendation, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(userID uint, pagination *utils.Pagination, opts RecommendationListOptions) ([]model.LinggongRecommendation, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongRecommendation, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongRecommendation, int64, error)
}

// RecommendationListOptions 推荐列表过滤条件
type RecommendationListOptions struct {
	RecType string
	Status  *int
}

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推荐仓储实例
func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

func (r *recommendationRepository) Create(rec *model.LinggongRecommendation) error {
	return r.db.Create(rec).Error
}

func (r *recommendationRepository) FindByID(id uint) (*model.LinggongRecommendation, error) {
	var rec model.LinggongRecommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongRecommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *recommendationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongRecommendation{}, id).Error
}

func (r *recommendationRepository) List(userID uint, pagination *utils.Pagination, opts RecommendationListOptions) ([]model.LinggongRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRecommendation
	var total int64

	query := r.db.Model(&model.LinggongRecommendation{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if opts.RecType != "" {
		query = query.Where("rec_type = ?", opts.RecType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("score DESC, created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRecommendation
	var total int64
	query := r.db.Model(&model.LinggongRecommendation{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("score DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRecommendation
	var total int64
	query := r.db.Model(&model.LinggongRecommendation{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("score DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
