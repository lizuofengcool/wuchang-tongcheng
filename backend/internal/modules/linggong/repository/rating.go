// Package repository 同城零工兼职数据访问层 - 双向评价
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RatingRepository 双向评价仓储接口
type RatingRepository interface {
	Create(r *model.LinggongRating) error
	FindByID(id uint) (*model.LinggongRating, error)
	FindByRatingNo(no string) (*model.LinggongRating, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts RatingListOptions) ([]model.LinggongRating, int64, error)
	AdminList(pagination *utils.Pagination, opts RatingAdminListOptions) ([]model.LinggongRating, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error)
	ListByRater(raterID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error)
	ListByTarget(targetType string, targetID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error)
	GetRatingStats(targetType string, targetID uint) (*RatingStatsResult, error)
	IncrLikeCount(id uint) error
}

// RatingListOptions C 端评价列表过滤条件
type RatingListOptions struct {
	LinggongID    uint
	TaskID        uint
	ApplicationID uint
	RaterType     string
	TargetType    string
	TargetID      uint
	Rating        *int
	IsRecommended string
	Status        *int
	Keyword       string
}

// RatingAdminListOptions M 端评价列表过滤条件
type RatingAdminListOptions struct {
	RegionID      uint
	LinggongID    uint
	RaterID       uint
	TargetType    string
	TargetID      uint
	Rating        *int
	Status        *int
	Keyword       string
}

// RatingStatsResult 评价统计结果
type RatingStatsResult struct {
	TotalReviews int64   `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
	GoodRate     float64 `json:"good_rate"`
	MediumRate   float64 `json:"medium_rate"`
	BadRate      float64 `json:"bad_rate"`
}

type ratingRepository struct {
	db *gorm.DB
}

// NewRatingRepository 创建双向评价仓储实例
func NewRatingRepository(db *gorm.DB) RatingRepository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Create(rating *model.LinggongRating) error {
	return r.db.Create(rating).Error
}

func (r *ratingRepository) FindByID(id uint) (*model.LinggongRating, error) {
	var rating model.LinggongRating
	if err := r.db.First(&rating, id).Error; err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ratingRepository) FindByRatingNo(no string) (*model.LinggongRating, error) {
	var rating model.LinggongRating
	if err := r.db.Where("rating_no = ?", no).First(&rating).Error; err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *ratingRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongRating{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ratingRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongRating{}, id).Error
}

func (r *ratingRepository) List(regionID uint, pagination *utils.Pagination, opts RatingListOptions) ([]model.LinggongRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRating
	var total int64

	query := r.db.Model(&model.LinggongRating{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.TaskID > 0 {
		query = query.Where("task_id = ?", opts.TaskID)
	}
	if opts.ApplicationID > 0 {
		query = query.Where("application_id = ?", opts.ApplicationID)
	}
	if opts.RaterType != "" {
		query = query.Where("rater_type = ?", opts.RaterType)
	}
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.Rating != nil {
		query = query.Where("rating = ?", *opts.Rating)
	}
	if opts.IsRecommended != "" {
		query = query.Where("is_recommended = ?", opts.IsRecommended)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("rating_no ILIKE ? OR content ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("evaluated_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) AdminList(pagination *utils.Pagination, opts RatingAdminListOptions) ([]model.LinggongRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRating
	var total int64

	query := r.db.Model(&model.LinggongRating{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.RaterID > 0 {
		query = query.Where("rater_id = ?", opts.RaterID)
	}
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.Rating != nil {
		query = query.Where("rating = ?", *opts.Rating)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("rating_no ILIKE ? OR content ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRating
	var total int64
	query := r.db.Model(&model.LinggongRating{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("evaluated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) ListByRater(raterID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRating
	var total int64
	query := r.db.Model(&model.LinggongRating{}).Where("rater_id = ?", raterID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) ListByTarget(targetType string, targetID uint, pagination *utils.Pagination) ([]model.LinggongRating, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongRating
	var total int64
	query := r.db.Model(&model.LinggongRating{}).Where("target_type = ? AND target_id = ?", targetType, targetID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("evaluated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ratingRepository) GetRatingStats(targetType string, targetID uint) (*RatingStatsResult, error) {
	var result RatingStatsResult
	query := r.db.Model(&model.LinggongRating{}).
		Where("target_type = ? AND target_id = ? AND status = ?", targetType, targetID, 1)
	if err := query.Count(&result.TotalReviews).Error; err != nil {
		return nil, err
	}
	if result.TotalReviews == 0 {
		return &result, nil
	}
	var avg struct {
		Avg float64
	}
	if err := query.Select("AVG(rating) as avg").Scan(&avg).Error; err != nil {
		return nil, err
	}
	result.AvgRating = avg.Avg
	var goodCount, mediumCount, badCount int64
	r.db.Model(&model.LinggongRating{}).
		Where("target_type = ? AND target_id = ? AND status = ? AND rating >= ?", targetType, targetID, 1, 4).
		Count(&goodCount)
	r.db.Model(&model.LinggongRating{}).
		Where("target_type = ? AND target_id = ? AND status = ? AND rating = ?", targetType, targetID, 1, 3).
		Count(&mediumCount)
	r.db.Model(&model.LinggongRating{}).
		Where("target_type = ? AND target_id = ? AND status = ? AND rating <= ?", targetType, targetID, 1, 2).
		Count(&badCount)
	result.GoodRate = float64(goodCount) / float64(result.TotalReviews)
	result.MediumRate = float64(mediumCount) / float64(result.TotalReviews)
	result.BadRate = float64(badCount) / float64(result.TotalReviews)
	return &result, nil
}

func (r *ratingRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.LinggongRating{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}
