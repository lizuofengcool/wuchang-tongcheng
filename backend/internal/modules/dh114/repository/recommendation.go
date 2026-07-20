// Package repository 同城114数据访问层 - 推荐商家
// 首页推荐/分类推荐/附近推荐/个性化推荐
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RecommendationRepository 推荐商家仓储接口
type RecommendationRepository interface {
	Create(r *model.Dh114Recommendation) error
	FindByID(id uint) (*model.Dh114Recommendation, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query RecommendationListQuery, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error)
	ListByType(recommendType string, regionID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error)
	ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error)
	MarkClicked(id uint, clickedAt interface{}) error
	MarkContacted(id uint, contactedAt interface{}) error
	MarkDismissed(id uint, dismissedAt interface{}) error
}

// RecommendationListQuery 推荐列表查询
type RecommendationListQuery struct {
	UserID       uint
	Dh114ID      uint
	RecommendType string
	CategoryID   uint
	Status       *int
}

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推荐商家仓储实例
func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

func (r *recommendationRepository) Create(rec *model.Dh114Recommendation) error {
	return r.db.Create(rec).Error
}

func (r *recommendationRepository) FindByID(id uint) (*model.Dh114Recommendation, error) {
	var rec model.Dh114Recommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Recommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *recommendationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Recommendation{}, id).Error
}

func (r *recommendationRepository) List(query RecommendationListQuery, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Dh114Recommendation
	var total int64

	q := r.db.Model(&model.Dh114Recommendation{})
	if query.UserID > 0 {
		q = q.Where("user_id = ? OR user_id = 0", query.UserID)
	}
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.RecommendType != "" {
		q = q.Where("recommend_type = ?", query.RecommendType)
	}
	if query.CategoryID > 0 {
		q = q.Where("category_id = ?", query.CategoryID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("score DESC, position ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByType(recommendType string, regionID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Dh114Recommendation
	var total int64

	q := r.db.Model(&model.Dh114Recommendation{}).
		Where("recommend_type = ?", recommendType).
		Where("(expire_at IS NULL OR expire_at > NOW())")
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("score DESC, position ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error) {
	return r.List(RecommendationListQuery{UserID: userID}, pagination)
}

func (r *recommendationRepository) ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Recommendation, int64, error) {
	return r.List(RecommendationListQuery{Dh114ID: dh114ID}, pagination)
}

func (r *recommendationRepository) MarkClicked(id uint, clickedAt interface{}) error {
	return r.db.Model(&model.Dh114Recommendation{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.RecommendStatusClicked,
			"clicked_at": clickedAt,
		}).Error
}

func (r *recommendationRepository) MarkContacted(id uint, contactedAt interface{}) error {
	return r.db.Model(&model.Dh114Recommendation{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.RecommendStatusContacted,
			"contacted_at": contactedAt,
		}).Error
}

func (r *recommendationRepository) MarkDismissed(id uint, dismissedAt interface{}) error {
	return r.db.Model(&model.Dh114Recommendation{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.RecommendStatusDismissed,
			"dismissed_at": dismissedAt,
		}).Error
}
