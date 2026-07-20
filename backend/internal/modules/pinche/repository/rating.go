// Package repository 同城拼车出行数据访问层 - 评价
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RatingListOptions 评价列表过滤条件
type RatingListOptions struct {
	PincheID   uint
	BookingID  uint
	RaterID    uint
	RateeID    uint
	RatingType string
	Rating     *int
	Status     *int
	Keyword    string
}

// RatingRepository 评价仓储接口
type RatingRepository interface {
	Create(r *model.PincheRating) error
	FindByID(id uint) (*model.PincheRating, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts RatingListOptions) ([]model.PincheRating, int64, error)
	ListByRater(raterID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error)
	ListByRatee(rateeID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error)

	HasRated(raterID, pincheID uint, ratingType string) (bool, error)
	UpdateStatus(id uint, status int) error
	UpdateReply(id uint, reply string) error
	IncrLikeCount(id uint) error
	StatsByTarget(rateeID uint) (total int64, avg float64, good, medium, bad int64, err error)
}

type ratingRepository struct {
	db *gorm.DB
}

// NewRatingRepository 创建评价仓储实例
func NewRatingRepository(db *gorm.DB) RatingRepository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Create(rt *model.PincheRating) error {
	return r.db.Create(rt).Error
}

func (r *ratingRepository) FindByID(id uint) (*model.PincheRating, error) {
	var rt model.PincheRating
	if err := r.db.First(&rt, id).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *ratingRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheRating{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ratingRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheRating{}, id).Error
}

func (r *ratingRepository) List(regionID uint, pagination *utils.Pagination, opts RatingListOptions) ([]model.PincheRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRating
	var total int64

	query := r.db.Model(&model.PincheRating{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.RaterID > 0 {
		query = query.Where("rater_id = ?", opts.RaterID)
	}
	if opts.RateeID > 0 {
		query = query.Where("ratee_id = ?", opts.RateeID)
	}
	if opts.RatingType != "" {
		query = query.Where("rating_type = ?", opts.RatingType)
	}
	if opts.Rating != nil {
		query = query.Where("rating = ?", *opts.Rating)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		query = query.Where("content ILIKE ?", "%"+opts.Keyword+"%")
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

func (r *ratingRepository) ListByRater(raterID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRating
	var total int64

	query := r.db.Model(&model.PincheRating{}).Where("rater_id = ?", raterID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) ListByRatee(rateeID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRating
	var total int64

	query := r.db.Model(&model.PincheRating{}).Where("ratee_id = ?", rateeID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRating
	var total int64

	query := r.db.Model(&model.PincheRating{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) HasRated(raterID, pincheID uint, ratingType string) (bool, error) {
	var count int64
	q := r.db.Model(&model.PincheRating{}).Where("rater_id = ? AND pinche_id = ?", raterID, pincheID)
	if ratingType != "" {
		q = q.Where("rating_type = ?", ratingType)
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ratingRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheRating{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ratingRepository) UpdateReply(id uint, reply string) error {
	return r.db.Model(&model.PincheRating{}).Where("id = ?", id).
		Update("reply", reply).Error
}

func (r *ratingRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.PincheRating{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *ratingRepository) StatsByTarget(rateeID uint) (total int64, avg float64, good, medium, bad int64, err error) {
	var result struct {
		Total  int64   `gorm:"column:total"`
		Avg    float64 `gorm:"column:avg"`
		Good   int64   `gorm:"column:good"`
		Medium int64   `gorm:"column:medium"`
		Bad    int64   `gorm:"column:bad"`
	}
	err = r.db.Model(&model.PincheRating{}).
		Where("ratee_id = ? AND status = ?", rateeID, model.RatingStatusApproved).
		Select(`
			COUNT(*) AS total,
			COALESCE(AVG(rating), 0) AS avg,
			COUNT(CASE WHEN rating >= 4 THEN 1 END) AS good,
			COUNT(CASE WHEN rating = 3 THEN 1 END) AS medium,
			COUNT(CASE WHEN rating <= 2 THEN 1 END) AS bad
		`).
		Scan(&result).Error
	if err != nil {
		return
	}
	return result.Total, result.Avg, result.Good, result.Medium, result.Bad, nil
}
