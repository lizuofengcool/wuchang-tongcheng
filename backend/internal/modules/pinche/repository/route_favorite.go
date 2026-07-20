// Package repository 同城拼车出行数据访问层 - 常用路线收藏
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RouteFavoriteRepository 常用路线收藏仓储接口
type RouteFavoriteRepository interface {
	Create(f *model.PincheRouteFavorite) error
	FindByID(id uint) (*model.PincheRouteFavorite, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheRouteFavorite, int64, error)
	IncrUseCount(id uint) error
	CountByUser(userID uint) (int64, error)
}

type routeFavoriteRepository struct {
	db *gorm.DB
}

// NewRouteFavoriteRepository 创建常用路线收藏仓储实例
func NewRouteFavoriteRepository(db *gorm.DB) RouteFavoriteRepository {
	return &routeFavoriteRepository{db: db}
}

func (r *routeFavoriteRepository) Create(f *model.PincheRouteFavorite) error {
	return r.db.Create(f).Error
}

func (r *routeFavoriteRepository) FindByID(id uint) (*model.PincheRouteFavorite, error) {
	var f model.PincheRouteFavorite
	if err := r.db.First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *routeFavoriteRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheRouteFavorite{}).Where("id = ?", id).Updates(fields).Error
}

func (r *routeFavoriteRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheRouteFavorite{}, id).Error
}

func (r *routeFavoriteRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheRouteFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRouteFavorite
	var total int64

	query := r.db.Model(&model.PincheRouteFavorite{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("use_count DESC, last_used_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *routeFavoriteRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.PincheRouteFavorite{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"use_count":   gorm.Expr("use_count + 1"),
			"last_used_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *routeFavoriteRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PincheRouteFavorite{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
