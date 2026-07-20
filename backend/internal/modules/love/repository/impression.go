// Package repository love 相亲交友数据访问层 - 印象标签
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveImpressionRepository 印象标签仓储接口
type LoveImpressionRepository interface {
	Create(i *model.LoveImpression) error
	FindByID(id uint) (*model.LoveImpression, error)
	Update(i *model.LoveImpression) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveImpressionListOptions) ([]model.LoveImpression, int64, error)
	ListByLoveID(loveID uint, pagination *utils.Pagination) ([]model.LoveImpression, int64, error)
	ListByFromUser(fromUserID uint, pagination *utils.Pagination) ([]model.LoveImpression, int64, error)
	CountByLoveID(loveID uint) (int64, error)
	CountByLoveIDAndTag(loveID uint, tag string) (int64, error)
	ListTopTagsByLoveID(loveID uint, limit int) ([]LoveImpressionTagCount, error)
}

// LoveImpressionListOptions 印象列表过滤
type LoveImpressionListOptions struct {
	LoveID     uint
	UserID     uint
	FromUserID uint
	Tag        string
}

// LoveImpressionTagCount 标签统计
type LoveImpressionTagCount struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

type loveImpressionRepository struct {
	db *gorm.DB
}

// NewLoveImpressionRepository 创建印象标签仓储
func NewLoveImpressionRepository(db *gorm.DB) LoveImpressionRepository {
	return &loveImpressionRepository{db: db}
}

func (r *loveImpressionRepository) Create(i *model.LoveImpression) error {
	return r.db.Create(i).Error
}

func (r *loveImpressionRepository) FindByID(id uint) (*model.LoveImpression, error) {
	var i model.LoveImpression
	if err := r.db.First(&i, id).Error; err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *loveImpressionRepository) Update(i *model.LoveImpression) error {
	return r.db.Save(i).Error
}

func (r *loveImpressionRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveImpression{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveImpressionRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveImpression{}, id).Error
}

func (r *loveImpressionRepository) List(pagination *utils.Pagination, opts LoveImpressionListOptions) ([]model.LoveImpression, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveImpression
	var total int64

	query := r.db.Model(&model.LoveImpression{})
	if opts.LoveID > 0 {
		query = query.Where("love_id = ?", opts.LoveID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.FromUserID > 0 {
		query = query.Where("from_user_id = ?", opts.FromUserID)
	}
	if opts.Tag != "" {
		query = query.Where("tag = ?", opts.Tag)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveImpressionRepository) ListByLoveID(loveID uint, pagination *utils.Pagination) ([]model.LoveImpression, int64, error) {
	return r.List(pagination, LoveImpressionListOptions{LoveID: loveID})
}

func (r *loveImpressionRepository) ListByFromUser(fromUserID uint, pagination *utils.Pagination) ([]model.LoveImpression, int64, error) {
	return r.List(pagination, LoveImpressionListOptions{FromUserID: fromUserID})
}

func (r *loveImpressionRepository) CountByLoveID(loveID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveImpression{}).Where("love_id = ? AND status = ?", loveID, 1).Count(&count).Error
	return count, err
}

func (r *loveImpressionRepository) CountByLoveIDAndTag(loveID uint, tag string) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveImpression{}).Where("love_id = ? AND tag = ? AND status = ?", loveID, tag, 1).Count(&count).Error
	return count, err
}

func (r *loveImpressionRepository) ListTopTagsByLoveID(loveID uint, limit int) ([]LoveImpressionTagCount, error) {
	var list []LoveImpressionTagCount
	if limit <= 0 {
		limit = 10
	}
	err := r.db.Model(&model.LoveImpression{}).
		Select("tag, COUNT(*) as count").
		Where("love_id = ? AND status = ?", loveID, 1).
		Group("tag").
		Order("count DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
