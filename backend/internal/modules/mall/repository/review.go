// Package repository 同城商城数据访问层 - 评价
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ReviewRepository 评价仓储接口
type ReviewRepository interface {
	Create(rv *model.Review) error
	FindByID(id uint) (*model.Review, error)
	Update(rv *model.Review) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(opts ReviewListOptions, pagination *utils.Pagination) ([]model.Review, int64, error)
	ListByProduct(productID uint, pagination *utils.Pagination) ([]model.Review, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Review, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Review, int64, error)
	ListByOrder(orderID uint) ([]model.Review, error)

	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	UpdateReply(id uint, reply string, replyUserID uint) error
	UpdateAppend(id uint, appendContent string, appendImages model.JSONB) error

	// 计数器
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	IncrDislikeCount(id uint) error
	DecrDislikeCount(id uint) error

	// 统计
	StatsByProduct(productID uint) (*ReviewStatsResult, error)
	StatsByShop(shopID uint) (*ReviewStatsResult, error)
}

// ReviewListOptions 评价列表过滤条件
type ReviewListOptions struct {
	ProductID uint
	SkuID     uint
	ShopID    uint
	UserID    uint
	OrderID   uint
	Rating    *int
	Status    *int
	Type      *int
	HasReply  *bool
	HasImages *bool
	Sort      string
	Keyword   string
}

// ReviewStatsResult 评价统计结果
type ReviewStatsResult struct {
	TotalCount    int64   `gorm:"column:total_count" json:"total_count"`
	AvgRating     float64 `gorm:"column:avg_rating" json:"avg_rating"`
	FiveStarCount int64   `gorm:"column:five_star_count" json:"five_star_count"`
	FourStarCount int64   `gorm:"column:four_star_count" json:"four_star_count"`
	ThreeStarCount int64  `gorm:"column:three_star_count" json:"three_star_count"`
	TwoStarCount  int64   `gorm:"column:two_star_count" json:"two_star_count"`
	OneStarCount  int64   `gorm:"column:one_star_count" json:"one_star_count"`
	HasImagesCount int64  `gorm:"column:has_images_count" json:"has_images_count"`
	HasVideoCount int64   `gorm:"column:has_video_count" json:"has_video_count"`
	GoodRate      float64 `gorm:"column:good_rate" json:"good_rate"`
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建评价仓储实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.Review) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByID(id uint) (*model.Review, error) {
	var rv model.Review
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) Update(rv *model.Review) error {
	return r.db.Save(rv).Error
}

func (r *reviewRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.Review{}, id).Error
}

func (r *reviewRepository) List(opts ReviewListOptions, pagination *utils.Pagination) ([]model.Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Review
	var total int64

	query := r.db.Model(&model.Review{})
	if opts.ProductID > 0 {
		query = query.Where("product_id = ?", opts.ProductID)
	}
	if opts.SkuID > 0 {
		query = query.Where("sku_id = ?", opts.SkuID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.OrderID > 0 {
		query = query.Where("order_id = ?", opts.OrderID)
	}
	if opts.Rating != nil {
		query = query.Where("rating = ?", *opts.Rating)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	} else {
		query = query.Where("status = ?", model.ReviewStatusApproved)
	}
	if opts.Type != nil {
		query = query.Where("type = ?", *opts.Type)
	}
	if opts.HasReply != nil {
		if *opts.HasReply {
			query = query.Where("reply <> ''")
		} else {
			query = query.Where("reply = ''")
		}
	}
	if opts.HasImages != nil {
		if *opts.HasImages {
			query = query.Where("images IS NOT NULL AND images::text <> 'null' AND images::text <> '[]'")
		}
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("content ILIKE ? OR user_name ILIKE ? OR sku_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch opts.Sort {
	case "oldest":
		orderClause = "created_at ASC, id ASC"
	case "highest":
		orderClause = "rating DESC, id DESC"
	case "lowest":
		orderClause = "rating ASC, id DESC"
	case "useful":
		orderClause = "like_count DESC, id DESC"
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) ListByProduct(productID uint, pagination *utils.Pagination) ([]model.Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Review
	var total int64

	query := r.db.Model(&model.Review{}).
		Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Review
	var total int64

	query := r.db.Model(&model.Review{}).Where("shop_id = ?", shopID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Review, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Review
	var total int64

	query := r.db.Model(&model.Review{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) ListByOrder(orderID uint) ([]model.Review, error) {
	var list []model.Review
	if err := r.db.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reviewRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Review{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) UpdateReply(id uint, reply string, replyUserID uint) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply":          reply,
			"reply_user_id":  replyUserID,
			"has_seller_reply": true,
		}).Error
}

func (r *reviewRepository) UpdateAppend(id uint, appendContent string, appendImages model.JSONB) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"append_content": appendContent,
			"append_images":  appendImages,
		}).Error
}

func (r *reviewRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *reviewRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.Review{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

func (r *reviewRepository) IncrDislikeCount(id uint) error {
	return r.db.Model(&model.Review{}).Where("id = ?", id).
		UpdateColumn("dislike_count", gorm.Expr("dislike_count + 1")).Error
}

func (r *reviewRepository) DecrDislikeCount(id uint) error {
	return r.db.Model(&model.Review{}).Where("id = ? AND dislike_count > 0", id).
		UpdateColumn("dislike_count", gorm.Expr("dislike_count - 1")).Error
}

func (r *reviewRepository) StatsByProduct(productID uint) (*ReviewStatsResult, error) {
	var result ReviewStatsResult
	query := r.db.Model(&model.Review{}).Where("product_id = ? AND status = ?", productID, model.ReviewStatusApproved).Select(`
		COUNT(*) AS total_count,
		COALESCE(AVG(rating), 0) AS avg_rating,
		COUNT(*) FILTER (WHERE rating = 5) AS five_star_count,
		COUNT(*) FILTER (WHERE rating = 4) AS four_star_count,
		COUNT(*) FILTER (WHERE rating = 3) AS three_star_count,
		COUNT(*) FILTER (WHERE rating = 2) AS two_star_count,
		COUNT(*) FILTER (WHERE rating = 1) AS one_star_count,
		COUNT(*) FILTER (WHERE images IS NOT NULL AND images::text <> 'null' AND images::text <> '[]') AS has_images_count,
		COUNT(*) FILTER (WHERE video <> '') AS has_video_count,
		CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE rating >= 4)::float / COUNT(*)::float ELSE 0 END AS good_rate
	`)
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *reviewRepository) StatsByShop(shopID uint) (*ReviewStatsResult, error) {
	var result ReviewStatsResult
	query := r.db.Model(&model.Review{}).Where("shop_id = ? AND status = ?", shopID, model.ReviewStatusApproved).Select(`
		COUNT(*) AS total_count,
		COALESCE(AVG(rating), 0) AS avg_rating,
		COUNT(*) FILTER (WHERE rating = 5) AS five_star_count,
		COUNT(*) FILTER (WHERE rating = 4) AS four_star_count,
		COUNT(*) FILTER (WHERE rating = 3) AS three_star_count,
		COUNT(*) FILTER (WHERE rating = 2) AS two_star_count,
		COUNT(*) FILTER (WHERE rating = 1) AS one_star_count,
		COUNT(*) FILTER (WHERE images IS NOT NULL AND images::text <> 'null' AND images::text <> '[]') AS has_images_count,
		COUNT(*) FILTER (WHERE video <> '') AS has_video_count,
		CASE WHEN COUNT(*) > 0 THEN COUNT(*) FILTER (WHERE rating >= 4)::float / COUNT(*)::float ELSE 0 END AS good_rate
	`)
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
