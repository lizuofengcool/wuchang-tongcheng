// Package repository 同城分类信息数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// NewsRepository 分类信息仓储接口
type NewsRepository interface {
	Create(news *model.News) error
	FindByID(id uint) (*model.News, error)
	Update(news *model.News) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, req *utils.Pagination, categoryID uint, status int, listingType string, keyword string, minPrice, maxPrice float64, isUrgent *bool, sort string) ([]model.News, int64, error)
	IncrViewCount(id uint) error
	// 点赞
	LikeExists(userID, newsID uint) (bool, error)
	CreateLike(like *model.NewsLike) error
	DeleteLike(userID, newsID uint) error
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	// 收藏
	FavExists(userID, newsID uint) (bool, error)
	CreateFav(fav *model.NewsFavorite) error
	DeleteFav(userID, newsID uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	// 评论
	CreateComment(comment *model.NewsComment) error
	ListComments(newsID uint, page, pageSize int) ([]model.NewsComment, int64, error)
	DeleteComment(id uint) error
	IncrCommentCount(id uint) error
	// 消息
	CreateMessage(msg *model.Message) error
	ListMessages(userID uint, page, pageSize int) ([]model.Message, int64, error)
	UnreadCount(userID uint) (int64, error)
	MarkRead(userID uint, ids []uint) error
	FindByIDs(ids []uint) ([]model.News, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(news *model.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) FindByID(id uint) (*model.News, error) {
	var news model.News
	if err := r.db.First(&news, id).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *newsRepository) Update(news *model.News) error {
	return r.db.Save(news).Error
}

func (r *newsRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).Updates(fields).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&model.News{}, id).Error
}

func (r *newsRepository) List(regionID uint, pagination *utils.Pagination, categoryID uint, status int, listingType string, keyword string, minPrice, maxPrice float64, isUrgent *bool, sort string) ([]model.News, int64, error) {
	var list []model.News
	var total int64

	query := r.db.Model(&model.News{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if listingType != "" {
		query = query.Where("listing_type = ?", listingType)
	}
	if status >= 0 && status <= 3 {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", 1)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}
	if isUrgent != nil && *isUrgent {
		query = query.Where("is_urgent = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch sort {
	case "price":
		orderClause = "price ASC, id DESC"
	case "price_desc":
		orderClause = "price DESC, id DESC"
	case "views":
		orderClause = "view_count DESC, id DESC"
	}

	// 置顶优先
	orderClause = "is_urgent DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *newsRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ====== 点赞 ======

func (r *newsRepository) LikeExists(userID, newsID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.NewsLike{}).Where("user_id = ? AND news_id = ?", userID, newsID).Count(&count).Error
	return count > 0, err
}

func (r *newsRepository) CreateLike(like *model.NewsLike) error {
	return r.db.Create(like).Error
}

func (r *newsRepository) DeleteLike(userID, newsID uint) error {
	return r.db.Where("user_id = ? AND news_id = ?", userID, newsID).Delete(&model.NewsLike{}).Error
}

func (r *newsRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *newsRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

// ====== 收藏 ======

func (r *newsRepository) FavExists(userID, newsID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.NewsFavorite{}).Where("user_id = ? AND news_id = ?", userID, newsID).Count(&count).Error
	return count > 0, err
}

func (r *newsRepository) CreateFav(fav *model.NewsFavorite) error {
	return r.db.Create(fav).Error
}

func (r *newsRepository) DeleteFav(userID, newsID uint) error {
	return r.db.Where("user_id = ? AND news_id = ?", userID, newsID).Delete(&model.NewsFavorite{}).Error
}

func (r *newsRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *newsRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

// ====== 评论 ======

func (r *newsRepository) CreateComment(comment *model.NewsComment) error {
	return r.db.Create(comment).Error
}

func (r *newsRepository) ListComments(newsID uint, page, pageSize int) ([]model.NewsComment, int64, error) {
	var list []model.NewsComment
	var total int64
	query := r.db.Model(&model.NewsComment{}).Where("news_id = ? AND status = 1", newsID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *newsRepository) DeleteComment(id uint) error {
	return r.db.Model(&model.NewsComment{}).Where("id = ?", id).Update("status", 0).Error
}

func (r *newsRepository) IncrCommentCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// ====== 消息 ======

func (r *newsRepository) CreateMessage(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *newsRepository) ListMessages(userID uint, page, pageSize int) ([]model.Message, int64, error) {
	var list []model.Message
	var total int64
	query := r.db.Model(&model.Message{}).Where("to_user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *newsRepository) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).Where("to_user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *newsRepository) MarkRead(userID uint, ids []uint) error {
	return r.db.Model(&model.Message{}).Where("to_user_id = ? AND id IN ?", userID, ids).Update("is_read", true).Error
}

func (r *newsRepository) FindByIDs(ids []uint) ([]model.News, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.News
	if err := r.db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}