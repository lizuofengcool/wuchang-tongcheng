// Package repository 同城车辆买卖数据访问层 - 收藏 + 浏览记录扩展
// car.go 已实现 CarFavorite 的基本 CRUD 和 CarView 的 CreateView，
// 本文件提供 admin/统计场景下的扩展查询（按 car_id 反查收藏者、按 car_id 反查浏览者等）
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InteractionRepository 收藏+浏览扩展仓储接口
type InteractionRepository interface {
	// ===== 收藏扩展 =====
	// 按车源反查收藏列表（管理后台/车商查看谁收藏了这辆车）
	ListFavsByCarID(carID uint, pagination *utils.Pagination) ([]model.CarFavorite, int64, error)
	// 按车源统计收藏数
	CountFavsByCarID(carID uint) (int64, error)
	// 按 ID 删除收藏（管理后台）
	DeleteFavByID(id uint) error
	// 用户最近收藏的 N 个车源 ID（用于推荐：看过又推荐、相似车源）
	ListRecentFavCarIDs(userID uint, limit int) ([]uint, error)

	// ===== 浏览记录扩展 =====
	// 按车源反查浏览记录（管理后台查看曝光）
	ListViewsByCarID(carID uint, pagination *utils.Pagination) ([]model.CarView, int64, error)
	// 按车源统计浏览数
	CountViewsByCarID(carID uint) (int64, error)
	// 今日浏览数
	CountTodayViews(regionID uint) (int64, error)
	// 用户最近的 N 个浏览车源（用于推荐）
	ListRecentViewCarIDs(userID uint, limit int) ([]uint, error)
	// 按 IP 统计（用于风控：识别爬虫）
	CountViewsByIP(ip string, since time.Time) (int64, error)
	// 清理超过 N 天的浏览记录
	DeleteOldViews(before time.Time) (int64, error)
}

type interactionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository 创建收藏+浏览扩展仓储实例
func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &interactionRepository{db: db}
}

// ===== 收藏扩展 =====

func (r *interactionRepository) ListFavsByCarID(carID uint, pagination *utils.Pagination) ([]model.CarFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarFavorite
	var total int64

	query := r.db.Model(&model.CarFavorite{}).Where("car_id = ?", carID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interactionRepository) CountFavsByCarID(carID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarFavorite{}).Where("car_id = ?", carID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *interactionRepository) DeleteFavByID(id uint) error {
	return r.db.Delete(&model.CarFavorite{}, id).Error
}

func (r *interactionRepository) ListRecentFavCarIDs(userID uint, limit int) ([]uint, error) {
	var ids []uint
	if userID == 0 {
		return ids, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if err := r.db.Model(&model.CarFavorite{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Pluck("car_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// ===== 浏览记录扩展 =====

func (r *interactionRepository) ListViewsByCarID(carID uint, pagination *utils.Pagination) ([]model.CarView, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarView
	var total int64

	query := r.db.Model(&model.CarView{}).Where("car_id = ?", carID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interactionRepository) CountViewsByCarID(carID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarView{}).Where("car_id = ?", carID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *interactionRepository) CountTodayViews(regionID uint) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarView{}).Where("created_at >= ?", time.Now().Truncate(24*time.Hour))
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *interactionRepository) ListRecentViewCarIDs(userID uint, limit int) ([]uint, error) {
	var ids []uint
	if userID == 0 {
		return ids, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if err := r.db.Model(&model.CarView{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Pluck("car_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *interactionRepository) CountViewsByIP(ip string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarView{}).
		Where("ip = ? AND created_at >= ?", ip, since).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *interactionRepository) DeleteOldViews(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&model.CarView{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
