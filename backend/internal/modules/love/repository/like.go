// Package repository love 相亲交友数据访问层 - 喜欢/不喜欢/心动信号
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveLikeRepository 喜欢记录仓储接口
type LoveLikeRepository interface {
	Create(l *model.LoveLike) error
	FindByID(id uint) (*model.LoveLike, error)
	FindByUserTarget(userID, targetUserID uint) (*model.LoveLike, error)
	Update(l *model.LoveLike) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveLikeListOptions) ([]model.LoveLike, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error)
	ListByTarget(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error)
	ListMatchedByUser(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error)

	HasLiked(userID, targetUserID uint) (bool, error)
	HasSuperLikedToday(userID uint) (int, error)
	CountTodayByAction(userID uint, action string) (int, error)

	Undo(id uint, reason string) error
	MarkMatched(id uint, matchID uint) error

	CountLikesByUser(userID uint) (int64, error)
	CountLikedByUser(userID uint) (int64, error)
	CountMatchedByUser(userID uint) (int64, error)
}

// LoveLikeListOptions 喜欢列表过滤
type LoveLikeListOptions struct {
	UserID       uint
	TargetUserID uint
	Action       string
	SuperLike    *bool
	IsMatched    *bool
}

type loveLikeRepository struct {
	db *gorm.DB
}

// NewLoveLikeRepository 创建喜欢记录仓储
func NewLoveLikeRepository(db *gorm.DB) LoveLikeRepository {
	return &loveLikeRepository{db: db}
}

func (r *loveLikeRepository) Create(l *model.LoveLike) error {
	return r.db.Create(l).Error
}

func (r *loveLikeRepository) FindByID(id uint) (*model.LoveLike, error) {
	var l model.LoveLike
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveLikeRepository) FindByUserTarget(userID, targetUserID uint) (*model.LoveLike, error) {
	var l model.LoveLike
	if err := r.db.Where("user_id = ? AND target_user_id = ?", userID, targetUserID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveLikeRepository) Update(l *model.LoveLike) error {
	return r.db.Save(l).Error
}

func (r *loveLikeRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveLike{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveLikeRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveLike{}, id).Error
}

func (r *loveLikeRepository) List(pagination *utils.Pagination, opts LoveLikeListOptions) ([]model.LoveLike, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveLike
	var total int64

	query := r.db.Model(&model.LoveLike{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.TargetUserID > 0 {
		query = query.Where("target_user_id = ?", opts.TargetUserID)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.SuperLike != nil {
		query = query.Where("super_like = ?", *opts.SuperLike)
	}
	if opts.IsMatched != nil {
		query = query.Where("is_matched = ?", *opts.IsMatched)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveLikeRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error) {
	return r.List(pagination, LoveLikeListOptions{UserID: userID})
}

func (r *loveLikeRepository) ListByTarget(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error) {
	return r.List(pagination, LoveLikeListOptions{TargetUserID: userID})
}

func (r *loveLikeRepository) ListMatchedByUser(userID uint, pagination *utils.Pagination) ([]model.LoveLike, int64, error) {
	matched := true
	return r.List(pagination, LoveLikeListOptions{UserID: userID, IsMatched: &matched})
}

func (r *loveLikeRepository) HasLiked(userID, targetUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where("user_id = ? AND target_user_id = ? AND status = ?", userID, targetUserID, 1).Count(&count).Error
	return count > 0, err
}

func (r *loveLikeRepository) HasSuperLikedToday(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where(
		"user_id = ? AND super_like = ? AND created_at >= DATE_TRUNC('day', NOW())",
		userID, true,
	).Count(&count).Error
	return int(count), err
}

func (r *loveLikeRepository) CountTodayByAction(userID uint, action string) (int, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where(
		"user_id = ? AND action = ? AND created_at >= DATE_TRUNC('day', NOW())",
		userID, action,
	).Count(&count).Error
	return int(count), err
}

func (r *loveLikeRepository) Undo(id uint, reason string) error {
	return r.db.Model(&model.LoveLike{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      0,
		"undone_at":   gorm.Expr("NOW()"),
		"undo_reason": reason,
	}).Error
}

func (r *loveLikeRepository) MarkMatched(id uint, matchID uint) error {
	return r.db.Model(&model.LoveLike{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_matched": true,
		"match_id":   matchID,
		"matched_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveLikeRepository) CountLikesByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where("user_id = ? AND action = ? AND status = ?", userID, model.LikeActionLike, 1).Count(&count).Error
	return count, err
}

func (r *loveLikeRepository) CountLikedByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where("target_user_id = ? AND action = ? AND status = ?", userID, model.LikeActionLike, 1).Count(&count).Error
	return count, err
}

func (r *loveLikeRepository) CountMatchedByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveLike{}).Where(
		"(user_id = ? OR target_user_id = ?) AND is_matched = ? AND status = ?",
		userID, userID, true, 1,
	).Count(&count).Error
	return count, err
}
