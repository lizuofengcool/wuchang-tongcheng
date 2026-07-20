// Package repository 同城114数据访问层 - 收藏
// 商户/团购/优惠券收藏 + 收藏分组
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// FavoriteRepository 收藏仓储接口
type FavoriteRepository interface {
	Create(fav *model.Dh114Favorite) error
	FindByID(id uint) (*model.Dh114Favorite, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	DeleteByUserAndTarget(userID uint, dh114ID uint, favoriteType string) error
	Exists(userID uint, dh114ID uint, favoriteType string) (bool, error)
	List(query FavoriteListQuery, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error)
	ListByType(userID uint, favoriteType string, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error)
	ListByGroup(userID uint, groupID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error)
	HasFavedBatch(userID uint, ids []uint, favoriteType string) (map[uint]bool, error)
}

// FavoriteListQuery 收藏列表查询
type FavoriteListQuery struct {
	UserID       uint
	Dh114ID      uint
	FavoriteType string
	GroupID      uint
}

type favoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 创建收藏仓储实例
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Create(fav *model.Dh114Favorite) error {
	return r.db.Create(fav).Error
}

func (r *favoriteRepository) FindByID(id uint) (*model.Dh114Favorite, error) {
	var fav model.Dh114Favorite
	if err := r.db.First(&fav, id).Error; err != nil {
		return nil, err
	}
	return &fav, nil
}

func (r *favoriteRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Favorite{}).Where("id = ?", id).Updates(fields).Error
}

func (r *favoriteRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Favorite{}, id).Error
}

func (r *favoriteRepository) DeleteByUserAndTarget(userID uint, dh114ID uint, favoriteType string) error {
	return r.db.Where("user_id = ? AND dh114_id = ? AND favorite_type = ?", userID, dh114ID, favoriteType).
		Delete(&model.Dh114Favorite{}).Error
}

func (r *favoriteRepository) Exists(userID uint, dh114ID uint, favoriteType string) (bool, error) {
	var count int64
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	if err := r.db.Model(&model.Dh114Favorite{}).
		Where("user_id = ? AND dh114_id = ? AND favorite_type = ?", userID, dh114ID, favoriteType).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *favoriteRepository) List(query FavoriteListQuery, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Dh114Favorite
	var total int64

	q := r.db.Model(&model.Dh114Favorite{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.FavoriteType != "" {
		q = q.Where("favorite_type = ?", query.FavoriteType)
	}
	if query.GroupID > 0 {
		q = q.Where("group_id = ?", query.GroupID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *favoriteRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return r.List(FavoriteListQuery{UserID: userID}, pagination)
}

func (r *favoriteRepository) ListByType(userID uint, favoriteType string, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return r.List(FavoriteListQuery{UserID: userID, FavoriteType: favoriteType}, pagination)
}

func (r *favoriteRepository) ListByGroup(userID uint, groupID uint, pagination *utils.Pagination) ([]model.Dh114Favorite, int64, error) {
	return r.List(FavoriteListQuery{UserID: userID, GroupID: groupID}, pagination)
}

func (r *favoriteRepository) HasFavedBatch(userID uint, ids []uint, favoriteType string) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	if favoriteType == "" {
		favoriteType = model.FavoriteTypeBusiness
	}
	var favs []model.Dh114Favorite
	if err := r.db.Where("user_id = ? AND dh114_id IN ? AND favorite_type = ?", userID, ids, favoriteType).
		Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		result[f.Dh114ID] = true
	}
	return result, nil
}
