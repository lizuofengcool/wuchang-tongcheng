// Package repository 同城零工兼职数据访问层 - 雇主认证
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// EmployerRepository 雇主认证仓储接口
type EmployerRepository interface {
	Create(e *model.LinggongEmployer) error
	FindByID(id uint) (*model.LinggongEmployer, error)
	FindByUserID(userID uint) (*model.LinggongEmployer, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts EmployerListOptions) ([]model.LinggongEmployer, int64, error)
	AdminList(pagination *utils.Pagination, opts EmployerAdminListOptions) ([]model.LinggongEmployer, int64, error)

	IncrPublishedCount(id uint) error
	IncrOngoingCount(id uint) error
	IncrCompletedCount(id uint) error
	IncrTotalWorkers(id uint, count int) error
	UpdateRating(id uint, avgRating float64, ratingCount int) error
}

// EmployerListOptions C 端雇主列表过滤条件
type EmployerListOptions struct {
	EmployerType string
	CompanyName  string
	Status       *int
	Level        *int
	Industry     string
	Keyword      string
}

// EmployerAdminListOptions M 端雇主列表过滤条件
type EmployerAdminListOptions struct {
	RegionID     uint
	UserID       uint
	EmployerType string
	Status       *int
	Level        *int
	Keyword      string
}

type employerRepository struct {
	db *gorm.DB
}

// NewEmployerRepository 创建雇主认证仓储实例
func NewEmployerRepository(db *gorm.DB) EmployerRepository {
	return &employerRepository{db: db}
}

func (r *employerRepository) Create(e *model.LinggongEmployer) error {
	return r.db.Create(e).Error
}

func (r *employerRepository) FindByID(id uint) (*model.LinggongEmployer, error) {
	var e model.LinggongEmployer
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employerRepository) FindByUserID(userID uint) (*model.LinggongEmployer, error) {
	var e model.LinggongEmployer
	if err := r.db.Where("user_id = ?", userID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *employerRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).Updates(fields).Error
}

func (r *employerRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongEmployer{}, id).Error
}

func (r *employerRepository) List(regionID uint, pagination *utils.Pagination, opts EmployerListOptions) ([]model.LinggongEmployer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongEmployer
	var total int64

	query := r.db.Model(&model.LinggongEmployer{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.EmployerType != "" {
		query = query.Where("employer_type = ?", opts.EmployerType)
	}
	if opts.CompanyName != "" {
		query = query.Where("company_name ILIKE ?", "%"+opts.CompanyName+"%")
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Industry != "" {
		query = query.Where("industry = ?", opts.Industry)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("company_name ILIKE ? OR contact_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("level DESC, created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *employerRepository) AdminList(pagination *utils.Pagination, opts EmployerAdminListOptions) ([]model.LinggongEmployer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongEmployer
	var total int64

	query := r.db.Model(&model.LinggongEmployer{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.EmployerType != "" {
		query = query.Where("employer_type = ?", opts.EmployerType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("company_name ILIKE ? OR contact_name ILIKE ? OR contact_phone ILIKE ?", like, like, like)
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

func (r *employerRepository) IncrPublishedCount(id uint) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).
		UpdateColumn("published_count", gorm.Expr("published_count + 1")).Error
}

func (r *employerRepository) IncrOngoingCount(id uint) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).
		UpdateColumn("ongoing_count", gorm.Expr("ongoing_count + 1")).Error
}

func (r *employerRepository) IncrCompletedCount(id uint) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).
		UpdateColumn("completed_count", gorm.Expr("completed_count + 1")).Error
}

func (r *employerRepository) IncrTotalWorkers(id uint, count int) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).
		UpdateColumn("total_workers", gorm.Expr("total_workers + ?", count)).Error
}

func (r *employerRepository) UpdateRating(id uint, avgRating float64, ratingCount int) error {
	return r.db.Model(&model.LinggongEmployer{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"avg_rating":   avgRating,
			"rating_count": ratingCount,
		}).Error
}
