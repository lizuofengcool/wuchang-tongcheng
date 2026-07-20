// Package repository love 相亲交友数据访问层 - 推荐池
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveRecommendationRepository 推荐池仓储接口
type LoveRecommendationRepository interface {
	Create(r *model.LoveRecommendation) error
	FindByID(id uint) (*model.LoveRecommendation, error)
	Update(r *model.LoveRecommendation) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveRecommendationListOptions) ([]model.LoveRecommendation, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error)
	ListByUserAndType(userID uint, recType string, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error)
	ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error)
	BatchCreate(items []model.LoveRecommendation) error

	UpdateAction(id uint, action string) error
	MarkViewed(id uint) error
	MarkLiked(id uint) error
	MarkDisliked(id uint) error
	MarkSuperLiked(id uint) error
	MarkSkipped(id uint) error
	MarkDismissed(id uint) error

	DeleteExpired() (int64, error)
	CountByUser(userID uint) (int64, error)
	CountTodayByUser(userID uint) (int64, error)
}

// LoveRecommendationListOptions 推荐列表过滤
type LoveRecommendationListOptions struct {
	UserID   uint
	RecType  string
	Source   string
	Status   *int
}

type loveRecommendationRepository struct {
	db *gorm.DB
}

// NewLoveRecommendationRepository 创建推荐池仓储
func NewLoveRecommendationRepository(db *gorm.DB) LoveRecommendationRepository {
	return &loveRecommendationRepository{db: db}
}

func (r *loveRecommendationRepository) Create(rec *model.LoveRecommendation) error {
	return r.db.Create(rec).Error
}

func (r *loveRecommendationRepository) FindByID(id uint) (*model.LoveRecommendation, error) {
	var rec model.LoveRecommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *loveRecommendationRepository) Update(rec *model.LoveRecommendation) error {
	return r.db.Save(rec).Error
}

func (r *loveRecommendationRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveRecommendationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveRecommendation{}, id).Error
}

func (r *loveRecommendationRepository) List(pagination *utils.Pagination, opts LoveRecommendationListOptions) ([]model.LoveRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveRecommendation
	var total int64

	query := r.db.Model(&model.LoveRecommendation{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.RecType != "" {
		query = query.Where("rec_type = ?", opts.RecType)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("score DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveRecommendationRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error) {
	return r.List(pagination, LoveRecommendationListOptions{UserID: userID})
}

func (r *loveRecommendationRepository) ListByUserAndType(userID uint, recType string, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error) {
	return r.List(pagination, LoveRecommendationListOptions{UserID: userID, RecType: recType})
}

func (r *loveRecommendationRepository) ListByUserAndStatus(userID uint, status int, pagination *utils.Pagination) ([]model.LoveRecommendation, int64, error) {
	st := status
	return r.List(pagination, LoveRecommendationListOptions{UserID: userID, Status: &st})
}

func (r *loveRecommendationRepository) BatchCreate(items []model.LoveRecommendation) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.CreateInBatches(items, 100).Error
}

func (r *loveRecommendationRepository) UpdateAction(id uint, action string) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Update("status", action).Error
}

func (r *loveRecommendationRepository) MarkViewed(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_viewed": true,
		"viewed_at": gorm.Expr("NOW()"),
		"status":    model.RecStatusViewed,
	}).Error
}

func (r *loveRecommendationRepository) MarkLiked(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_liked": true,
		"liked_at": gorm.Expr("NOW()"),
		"status":   model.RecStatusLiked,
	}).Error
}

func (r *loveRecommendationRepository) MarkDisliked(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_disliked": true,
		"disliked_at": gorm.Expr("NOW()"),
		"status":      model.RecStatusDisliked,
	}).Error
}

func (r *loveRecommendationRepository) MarkSuperLiked(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_super_liked": true,
		"super_liked_at": gorm.Expr("NOW()"),
		"status":         model.RecStatusLiked,
	}).Error
}

func (r *loveRecommendationRepository) MarkSkipped(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_skipped": true,
		"skipped_at": gorm.Expr("NOW()"),
		"status":     model.RecStatusSkipped,
	}).Error
}

func (r *loveRecommendationRepository) MarkDismissed(id uint) error {
	return r.db.Model(&model.LoveRecommendation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_dismissed": true,
		"dismissed_at": gorm.Expr("NOW()"),
		"status":       model.RecStatusDismissed,
	}).Error
}

func (r *loveRecommendationRepository) DeleteExpired() (int64, error) {
	result := r.db.Where("expired_at < NOW() AND status = ?", model.RecStatusPending).Delete(&model.LoveRecommendation{})
	return result.RowsAffected, result.Error
}

func (r *loveRecommendationRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveRecommendation{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *loveRecommendationRepository) CountTodayByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveRecommendation{}).Where("user_id = ? AND created_at >= DATE_TRUNC('day', NOW())", userID).Count(&count).Error
	return count, err
}
