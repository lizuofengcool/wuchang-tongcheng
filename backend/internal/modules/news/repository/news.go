// Package repository 同城头条数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// NewsRepository 头条仓储接口
type NewsRepository interface {
	Create(news *model.News) error
	FindByID(id uint) (*model.News, error)
	Update(news *model.News) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, req *utils.Pagination, categoryID uint, status int, keyword string) ([]model.News, int64, error)
	IncrViewCount(id uint) error
	// 点赞相关
	LikeExists(userID, newsID uint) (bool, error)
	CreateLike(like *model.NewsLike) error
	DeleteLike(userID, newsID uint) error
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
}

type newsRepository struct {
	db *gorm.DB
}

// NewNewsRepository 创建头条仓储
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

func (r *newsRepository) List(regionID uint, pagination *utils.Pagination, categoryID uint, status int, keyword string) ([]model.News, int64, error) {
	var list []model.News
	var total int64

	query := r.db.Model(&model.News{})

	// 地区过滤
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	// 分类过滤
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	// 状态过滤
	if status >= 0 && status <= 2 {
		query = query.Where("status = ?", status)
	} else {
		// 默认只查已发布
		query = query.Where("status = ?", 1)
	}
	// 关键词搜索（标题）
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，按发布时间倒序
	if err := query.Scopes(utils.Paginate(pagination)).Order("published_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *newsRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// LikeExists 检查用户是否已对头条点赞
func (r *newsRepository) LikeExists(userID, newsID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.NewsLike{}).Where("user_id = ? AND news_id = ?", userID, newsID).Count(&count).Error
	return count > 0, err
}

// CreateLike 创建点赞记录
func (r *newsRepository) CreateLike(like *model.NewsLike) error {
	return r.db.Create(like).Error
}

// DeleteLike 删除点赞记录
func (r *newsRepository) DeleteLike(userID, newsID uint) error {
	return r.db.Where("user_id = ? AND news_id = ?", userID, newsID).Delete(&model.NewsLike{}).Error
}

// IncrLikeCount 点赞数 +1
func (r *newsRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrLikeCount 点赞数 -1（不低于 0）
func (r *newsRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}
