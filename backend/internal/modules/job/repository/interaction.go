// Package repository 收藏 + 浏览记录数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘职位/公司/简历收藏与浏览
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// InteractionRepository 互动（收藏/浏览）仓储接口
type InteractionRepository interface {
	// 收藏
	CreateFav(fav *model.JobFavorite) error
	DeleteFav(userID uint, favType string, targetID uint) error
	FavExists(userID uint, favType string, targetID uint) (bool, error)
	ListFavs(userID uint, favType string, pagination *utils.Pagination) ([]model.JobFavorite, int64, error)
	HasFavedBatch(userID uint, favType string, ids []uint) (map[uint]bool, error)
	ClearFavs(userID uint, favType string) error

	// 浏览
	CreateView(view *model.JobView) error
	ListViews(userID uint, viewType string, pagination *utils.Pagination) ([]model.JobView, int64, error)
	ListViewsByTarget(targetType string, targetID uint, pagination *utils.Pagination) ([]model.JobView, int64, error)
	ClearViews(userID uint, viewType string) error
}

type interactionRepository struct {
	db *gorm.DB
}

// NewInteractionRepository 创建互动仓储实例
func NewInteractionRepository(db *gorm.DB) InteractionRepository {
	return &interactionRepository{db: db}
}

// ===== 收藏 =====

func (r *interactionRepository) CreateFav(fav *model.JobFavorite) error {
	return r.db.Create(fav).Error
}

func (r *interactionRepository) DeleteFav(userID uint, favType string, targetID uint) error {
	q := r.db.Where("user_id = ? AND favorite_type = ?", userID, favType)
	switch favType {
	case model.FavoriteTypeJob:
		q = q.Where("job_id = ?", targetID)
	case model.FavoriteTypeCompany:
		q = q.Where("company_id = ?", targetID)
	}
	return q.Delete(&model.JobFavorite{}).Error
}

func (r *interactionRepository) FavExists(userID uint, favType string, targetID uint) (bool, error) {
	var count int64
	q := r.db.Model(&model.JobFavorite{}).
		Where("user_id = ? AND favorite_type = ?", userID, favType)
	switch favType {
	case model.FavoriteTypeJob:
		q = q.Where("job_id = ?", targetID)
	case model.FavoriteTypeCompany:
		q = q.Where("company_id = ?", targetID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *interactionRepository) ListFavs(userID uint, favType string, pagination *utils.Pagination) ([]model.JobFavorite, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobFavorite
	var total int64

	q := r.db.Model(&model.JobFavorite{}).
		Where("user_id = ? AND favorite_type = ?", userID, favType)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *interactionRepository) HasFavedBatch(userID uint, favType string, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	var favs []model.JobFavorite
	q := r.db.Model(&model.JobFavorite{}).Where("user_id = ? AND favorite_type = ?", userID, favType)
	switch favType {
	case model.FavoriteTypeJob:
		q = q.Where("job_id IN ?", ids)
	case model.FavoriteTypeCompany:
		q = q.Where("company_id IN ?", ids)
	}
	if err := q.Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		switch favType {
		case model.FavoriteTypeJob:
			result[f.JobID] = true
		case model.FavoriteTypeCompany:
			result[f.CompanyID] = true
		}
	}
	return result, nil
}

func (r *interactionRepository) ClearFavs(userID uint, favType string) error {
	return r.db.Where("user_id = ? AND favorite_type = ?", userID, favType).
		Delete(&model.JobFavorite{}).Error
}

// ===== 浏览 =====

func (r *interactionRepository) CreateView(view *model.JobView) error {
	return r.db.Create(view).Error
}

func (r *interactionRepository) ListViews(userID uint, viewType string, pagination *utils.Pagination) ([]model.JobView, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.JobView
	var total int64

	q := r.db.Model(&model.JobView{}).Where("user_id = ?", userID)
	if viewType != "" {
		q = q.Where("view_type = ?", viewType)
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

func (r *interactionRepository) ListViewsByTarget(targetType string, targetID uint, pagination *utils.Pagination) ([]model.JobView, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.JobView
	var total int64

	q := r.db.Model(&model.JobView{}).Where("view_type = ?", targetType)
	switch targetType {
	case model.ViewTypeJob:
		q = q.Where("job_id = ?", targetID)
	case model.ViewTypeCompany:
		q = q.Where("company_id = ?", targetID)
	case model.ViewTypeResume:
		q = q.Where("resume_id = ?", targetID)
	case model.ViewTypeRecruiter:
		q = q.Where("recruiter_id = ?", targetID)
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

func (r *interactionRepository) ClearViews(userID uint, viewType string) error {
	q := r.db.Where("user_id = ?", userID)
	if viewType != "" {
		q = q.Where("view_type = ?", viewType)
	}
	return q.Delete(&model.JobView{}).Error
}
