// Package repository 同城114数据访问层 - 评价 + 商家回复
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== ReviewRepository 评价 =====

// ReviewRepository 评价仓储接口
type ReviewRepository interface {
	Create(r *model.Dh114Review) error
	FindByID(id uint) (*model.Dh114Review, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query ReviewListQuery, pagination *utils.Pagination) ([]model.Dh114Review, int64, error)
	ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Review, int64, error)
	ListByReviewer(reviewerID uint, pagination *utils.Pagination) ([]model.Dh114Review, int64, error)
	StatsByDh114(dh114ID uint) (total int64, avgRating float64, good, medium, bad int64, err error)
	StatsByReviewer(reviewerID uint) (total int64, avgRating float64, err error)
	HasReviewed(reviewerID, dh114ID uint) (bool, error)
	IncrLikeCount(id uint) error
	UpdateReply(id uint, reply string, repliedAt interface{}) error
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	Dh114ID    uint
	ReviewerID uint
	Rating     *int
	Status     *int
	HasReply   *bool
	Keyword    string
	Sort       string
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建评价仓储实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.Dh114Review) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByID(id uint) (*model.Dh114Review, error) {
	var rv model.Dh114Review
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Review{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Review{}, id).Error
}

func (r *reviewRepository) List(regionID uint, query ReviewListQuery, pagination *utils.Pagination) ([]model.Dh114Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Review
	var total int64

	q := r.db.Model(&model.Dh114Review{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.ReviewerID > 0 {
		q = q.Where("reviewer_id = ?", query.ReviewerID)
	}
	if query.Rating != nil {
		q = q.Where("rating = ?", *query.Rating)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.ReviewStatusApproved)
	}
	if query.HasReply != nil {
		if *query.HasReply {
			q = q.Where("has_reply = true")
		} else {
			q = q.Where("has_reply = false")
		}
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("content ILIKE ? OR reviewer_name ILIKE ? OR reply ILIKE ?", like, like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch query.Sort {
	case "rating_desc":
		orderClause = "rating DESC, id DESC"
	case "rating_asc":
		orderClause = "rating ASC, id DESC"
	case "like_desc":
		orderClause = "like_count DESC, id DESC"
	}

	if err := q.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Review, int64, error) {
	return r.List(regionID, ReviewListQuery{
		Dh114ID: dh114ID,
		Status:  intPtrDh114(model.ReviewStatusApproved),
	}, pagination)
}

func (r *reviewRepository) ListByReviewer(reviewerID uint, pagination *utils.Pagination) ([]model.Dh114Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Review
	var total int64

	q := r.db.Model(&model.Dh114Review{}).Where("reviewer_id = ?", reviewerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) StatsByDh114(dh114ID uint) (int64, float64, int64, int64, int64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
		Good    int64   `gorm:"column:good"`
		Medium  int64   `gorm:"column:medium"`
		Bad     int64   `gorm:"column:bad"`
	}
	var s stat
	err := r.db.Model(&model.Dh114Review{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate, "+
			"COUNT(CASE WHEN rating >= 4 THEN 1 END) AS good, "+
			"COUNT(CASE WHEN rating = 3 THEN 1 END) AS medium, "+
			"COUNT(CASE WHEN rating <= 2 THEN 1 END) AS bad").
		Where("dh114_id = ? AND status = ?", dh114ID, model.ReviewStatusApproved).
		Scan(&s).Error
	return s.Total, s.AvgRate, s.Good, s.Medium, s.Bad, err
}

func (r *reviewRepository) StatsByReviewer(reviewerID uint) (int64, float64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
	}
	var s stat
	err := r.db.Model(&model.Dh114Review{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate").
		Where("reviewer_id = ? AND status = ?", reviewerID, model.ReviewStatusApproved).
		Scan(&s).Error
	return s.Total, s.AvgRate, err
}

func (r *reviewRepository) HasReviewed(reviewerID, dh114ID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Dh114Review{}).
		Where("reviewer_id = ? AND dh114_id = ?", reviewerID, dh114ID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *reviewRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.Dh114Review{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *reviewRepository) UpdateReply(id uint, reply string, repliedAt interface{}) error {
	return r.db.Model(&model.Dh114Review{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply":      reply,
			"replied_at": repliedAt,
			"has_reply":  true,
		}).Error
}

// ===== ReviewReplyRepository 商家回复评价 =====

// ReviewReplyRepository 商家回复评价仓储接口
type ReviewReplyRepository interface {
	Create(rp *model.Dh114ReviewReply) error
	FindByID(id uint) (*model.Dh114ReviewReply, error)
	ListByReview(reviewID uint) ([]model.Dh114ReviewReply, error)
	ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114ReviewReply, int64, error)
	Delete(id uint) error
}

type reviewReplyRepository struct {
	db *gorm.DB
}

// NewReviewReplyRepository 创建商家回复评价仓储实例
func NewReviewReplyRepository(db *gorm.DB) ReviewReplyRepository {
	return &reviewReplyRepository{db: db}
}

func (r *reviewReplyRepository) Create(rp *model.Dh114ReviewReply) error {
	return r.db.Create(rp).Error
}

func (r *reviewReplyRepository) FindByID(id uint) (*model.Dh114ReviewReply, error) {
	var rp model.Dh114ReviewReply
	if err := r.db.First(&rp, id).Error; err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *reviewReplyRepository) ListByReview(reviewID uint) ([]model.Dh114ReviewReply, error) {
	var list []model.Dh114ReviewReply
	if err := r.db.Where("review_id = ? AND status = ?", reviewID, 1).
		Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reviewReplyRepository) ListByDh114(dh114ID uint, pagination *utils.Pagination) ([]model.Dh114ReviewReply, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114ReviewReply
	var total int64

	q := r.db.Model(&model.Dh114ReviewReply{}).Where("dh114_id = ?", dh114ID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewReplyRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114ReviewReply{}, id).Error
}

// intPtrDh114 工具函数：取 int 指针
func intPtrDh114(v int) *int { return &v }
