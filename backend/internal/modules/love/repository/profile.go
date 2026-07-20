// Package repository love 相亲交友数据访问层 - 详细资料
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveProfileRepository 详细资料仓储接口
type LoveProfileRepository interface {
	Create(p *model.LoveProfile) error
	FindByID(id uint) (*model.LoveProfile, error)
	FindByLoveID(loveID uint) (*model.LoveProfile, error)
	FindByUserID(userID uint) (*model.LoveProfile, error)
	Update(p *model.LoveProfile) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveProfileListOptions) ([]model.LoveProfile, int64, error)
	UpdateProfileScore(loveID uint, score int) error
	UpdateCompletedStep(loveID uint, step int) error
}

// LoveProfileListOptions 详细资料列表过滤
type LoveProfileListOptions struct {
	UserID    uint
	LoveID    uint
	Gender    *int
	City      string
	Keyword   string
}

type loveProfileRepository struct {
	db *gorm.DB
}

// NewLoveProfileRepository 创建详细资料仓储
func NewLoveProfileRepository(db *gorm.DB) LoveProfileRepository {
	return &loveProfileRepository{db: db}
}

func (r *loveProfileRepository) Create(p *model.LoveProfile) error {
	return r.db.Create(p).Error
}

func (r *loveProfileRepository) FindByID(id uint) (*model.LoveProfile, error) {
	var p model.LoveProfile
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *loveProfileRepository) FindByLoveID(loveID uint) (*model.LoveProfile, error) {
	var p model.LoveProfile
	if err := r.db.Where("love_id = ?", loveID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *loveProfileRepository) FindByUserID(userID uint) (*model.LoveProfile, error) {
	var p model.LoveProfile
	if err := r.db.Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *loveProfileRepository) Update(p *model.LoveProfile) error {
	return r.db.Save(p).Error
}

func (r *loveProfileRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveProfile{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveProfileRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveProfile{}, id).Error
}

func (r *loveProfileRepository) List(pagination *utils.Pagination, opts LoveProfileListOptions) ([]model.LoveProfile, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveProfile
	var total int64

	query := r.db.Model(&model.LoveProfile{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.LoveID > 0 {
		query = query.Where("love_id = ?", opts.LoveID)
	}
	if opts.Gender != nil {
		query = query.Where("gender = ?", *opts.Gender)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("nickname ILIKE ? OR occupation ILIKE ? OR company ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveProfileRepository) UpdateProfileScore(loveID uint, score int) error {
	return r.db.Model(&model.LoveProfile{}).Where("love_id = ?", loveID).Update("profile_score", score).Error
}

func (r *loveProfileRepository) UpdateCompletedStep(loveID uint, step int) error {
	return r.db.Model(&model.LoveProfile{}).Where("love_id = ?", loveID).Updates(map[string]interface{}{
		"completed_step": step,
		"completed_at":   gorm.Expr("NOW()"),
	}).Error
}
