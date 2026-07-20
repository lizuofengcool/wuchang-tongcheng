// Package repository 同城零工兼职数据访问层 - 收藏
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// FavoriteRepository 收藏仓储接口
type FavoriteRepository interface {
	Create(f *model.LinggongFavorite) error
	FindByID(id uint) (*model.LinggongFavorite, error)
	FindByUserAndTarget(userID uint, targetID uint, favoriteType string) (*model.LinggongFavorite, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(userID uint, pagination *utils.Pagination, opts FavoriteListOptions) ([]model.LinggongFavorite, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongFavorite, int64, error)
	ListByType(userID uint, favoriteType string, pagination *utils.Pagination) ([]model.LinggongFavorite, int64, error)
	CountByTarget(targetID uint, favoriteType string) (int64, error)
	Exists(userID uint, targetID uint, favoriteType string) (bool, error)
}

// FavoriteListOptions 收藏列表过滤条件
type FavoriteListOptions struct {
	FavoriteType string
}

type favoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 创建收藏仓储实例
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Create(f *model.LinggongFavorite) error {
	return r.db.Create(f).Error
}

func (r *favoriteRepository) FindByID(id uint) (*model.LinggongFavorite, error) {
	var f model.LinggongFavorite
	if err := r.db.First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *favoriteRepository) FindByUserAndTarget(userID uint, targetID uint, favoriteType string) (*model.LinggongFavorite, error) {
	var f model.LinggongFavorite
	if err := r.db.Where("user_id = ? AND target_id = ? AND favorite_type = ?", userID, targetID, favoriteType).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *favoriteRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongFavorite{}).Where("id = ?", id).Updates(fields).Error
}

func (r *favoriteRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongFavorite{}, id).Error
}

func (r *favoriteRepository) List(userID uint, pagination *utils.Pagination, opts FavoriteListOptions) ([]model.LinggongFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongFavorite
	var total int64

	query := r.db.Model(&model.LinggongFavorite{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if opts.FavoriteType != "" {
		query = query.Where("favorite_type = ?", opts.FavoriteType)
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

func (r *favoriteRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LinggongFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongFavorite
	var total int64
	query := r.db.Model(&model.LinggongFavorite{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *favoriteRepository) ListByType(userID uint, favoriteType string, pagination *utils.Pagination) ([]model.LinggongFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongFavorite
	var total int64
	query := r.db.Model(&model.LinggongFavorite{}).Where("user_id = ? AND favorite_type = ?", userID, favoriteType)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *favoriteRepository) CountByTarget(targetID uint, favoriteType string) (int64, error) {
	var count int64
	err := r.db.Model(&model.LinggongFavorite{}).
		Where("target_id = ? AND favorite_type = ?", targetID, favoriteType).
		Count(&count).Error
	return count, err
}

func (r *favoriteRepository) Exists(userID uint, targetID uint, favoriteType string) (bool, error) {
	var count int64
	err := r.db.Model(&model.LinggongFavorite{}).
		Where("user_id = ? AND target_id = ? AND favorite_type = ?", userID, targetID, favoriteType).
		Count(&count).Error
	return count > 0, err
}
