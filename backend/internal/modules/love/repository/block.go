// Package repository love 相亲交友数据访问层 - 拉黑
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveBlockRepository 拉黑仓储接口
type LoveBlockRepository interface {
	Create(b *model.LoveBlock) error
	FindByID(id uint) (*model.LoveBlock, error)
	FindByUserBlocked(userID, blockedUserID uint) (*model.LoveBlock, error)
	Update(b *model.LoveBlock) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveBlockListOptions) ([]model.LoveBlock, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveBlock, int64, error)
	HasBlocked(userID, blockedUserID uint) (bool, error)
	HasBlockedEither(userA, userB uint) (bool, error)
	Upsert(b *model.LoveBlock) error
	CountByUser(userID uint) (int64, error)
}

// LoveBlockListOptions 拉黑列表过滤
type LoveBlockListOptions struct {
	UserID uint
}

type loveBlockRepository struct {
	db *gorm.DB
}

// NewLoveBlockRepository 创建拉黑仓储
func NewLoveBlockRepository(db *gorm.DB) LoveBlockRepository {
	return &loveBlockRepository{db: db}
}

func (r *loveBlockRepository) Create(b *model.LoveBlock) error {
	return r.db.Create(b).Error
}

func (r *loveBlockRepository) FindByID(id uint) (*model.LoveBlock, error) {
	var b model.LoveBlock
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *loveBlockRepository) FindByUserBlocked(userID, blockedUserID uint) (*model.LoveBlock, error) {
	var b model.LoveBlock
	if err := r.db.Where("user_id = ? AND blocked_user_id = ?", userID, blockedUserID).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *loveBlockRepository) Update(b *model.LoveBlock) error {
	return r.db.Save(b).Error
}

func (r *loveBlockRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveBlock{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveBlockRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveBlock{}, id).Error
}

func (r *loveBlockRepository) List(pagination *utils.Pagination, opts LoveBlockListOptions) ([]model.LoveBlock, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveBlock
	var total int64

	query := r.db.Model(&model.LoveBlock{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveBlockRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.LoveBlock, int64, error) {
	return r.List(pagination, LoveBlockListOptions{UserID: userID})
}

func (r *loveBlockRepository) HasBlocked(userID, blockedUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.LoveBlock{}).Where("user_id = ? AND blocked_user_id = ? AND status = ?", userID, blockedUserID, 1).Count(&count).Error
	return count > 0, err
}

func (r *loveBlockRepository) HasBlockedEither(userA, userB uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.LoveBlock{}).Where(
		"((user_id = ? AND blocked_user_id = ?) OR (user_id = ? AND blocked_user_id = ?)) AND status = ?",
		userA, userB, userB, userA, 1,
	).Count(&count).Error
	return count > 0, err
}

func (r *loveBlockRepository) Upsert(b *model.LoveBlock) error {
	result := r.db.Where("user_id = ? AND blocked_user_id = ?", b.UserID, b.BlockedUserID).FirstOrCreate(b)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return r.db.Model(&model.LoveBlock{}).Where("user_id = ? AND blocked_user_id = ?", b.UserID, b.BlockedUserID).Updates(map[string]interface{}{
			"reason":     b.Reason,
			"report_id":  b.ReportID,
			"status":     1,
			"updated_at": gorm.Expr("NOW()"),
		}).Error
	}
	return nil
}

func (r *loveBlockRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveBlock{}).Where("user_id = ? AND status = ?", userID, 1).Count(&count).Error
	return count, err
}
