// Package repository 同城零工兼职数据访问层 - 求职者档案
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// WorkerRepository 求职者档案仓储接口
type WorkerRepository interface {
	Create(w *model.LinggongWorker) error
	FindByID(id uint) (*model.LinggongWorker, error)
	FindByUserID(userID uint) (*model.LinggongWorker, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts WorkerListOptions) ([]model.LinggongWorker, int64, error)
	AdminList(pagination *utils.Pagination, opts WorkerAdminListOptions) ([]model.LinggongWorker, int64, error)

	UpdateRating(id uint, avgRating float64, ratingCount int) error
	UpdateCreditScore(id uint, score int) error
	IncrAppliedCount(id uint) error
	IncrCompletedCount(id uint) error
	IncrTotalWorkHours(id uint, hours int) error
	IncrTotalEarnings(id uint, amount float64) error
}

// WorkerListOptions C 端求职者列表过滤条件
type WorkerListOptions struct {
	JobIntention    string
	Education       string
	City            string
	AvailableNow    *bool
	SkillID         uint
	MinCreditScore  int
	Status          *int
	Keyword         string
}

// WorkerAdminListOptions M 端求职者列表过滤条件
type WorkerAdminListOptions struct {
	RegionID     uint
	UserID       uint
	Status       *int
	JobIntention string
	Keyword      string
}

type workerRepository struct {
	db *gorm.DB
}

// NewWorkerRepository 创建求职者档案仓储实例
func NewWorkerRepository(db *gorm.DB) WorkerRepository {
	return &workerRepository{db: db}
}

func (r *workerRepository) Create(w *model.LinggongWorker) error {
	return r.db.Create(w).Error
}

func (r *workerRepository) FindByID(id uint) (*model.LinggongWorker, error) {
	var w model.LinggongWorker
	if err := r.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *workerRepository) FindByUserID(userID uint) (*model.LinggongWorker, error) {
	var w model.LinggongWorker
	if err := r.db.Where("user_id = ?", userID).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *workerRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).Updates(fields).Error
}

func (r *workerRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongWorker{}, id).Error
}

func (r *workerRepository) List(regionID uint, pagination *utils.Pagination, opts WorkerListOptions) ([]model.LinggongWorker, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongWorker
	var total int64

	query := r.db.Model(&model.LinggongWorker{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.JobIntention != "" {
		query = query.Where("job_intention = ?", opts.JobIntention)
	}
	if opts.Education != "" {
		query = query.Where("education = ?", opts.Education)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.AvailableNow != nil {
		query = query.Where("available_now = ?", *opts.AvailableNow)
	}
	if opts.MinCreditScore > 0 {
		query = query.Where("credit_score >= ?", opts.MinCreditScore)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("real_name ILIKE ? OR nickname ILIKE ? OR phone ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("credit_score DESC, completed_count DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *workerRepository) AdminList(pagination *utils.Pagination, opts WorkerAdminListOptions) ([]model.LinggongWorker, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongWorker
	var total int64

	query := r.db.Model(&model.LinggongWorker{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.JobIntention != "" {
		query = query.Where("job_intention = ?", opts.JobIntention)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("real_name ILIKE ? OR nickname ILIKE ? OR phone ILIKE ?", like, like, like)
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

func (r *workerRepository) UpdateRating(id uint, avgRating float64, ratingCount int) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"avg_rating":   avgRating,
			"rating_count": ratingCount,
		}).Error
}

func (r *workerRepository) UpdateCreditScore(id uint, score int) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		UpdateColumn("credit_score", score).Error
}

func (r *workerRepository) IncrAppliedCount(id uint) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		UpdateColumn("applied_count", gorm.Expr("applied_count + 1")).Error
}

func (r *workerRepository) IncrCompletedCount(id uint) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		UpdateColumn("completed_count", gorm.Expr("completed_count + 1")).Error
}

func (r *workerRepository) IncrTotalWorkHours(id uint, hours int) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		UpdateColumn("total_work_hours", gorm.Expr("total_work_hours + ?", hours)).Error
}

func (r *workerRepository) IncrTotalEarnings(id uint, amount float64) error {
	return r.db.Model(&model.LinggongWorker{}).Where("id = ?", id).
		UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", amount)).Error
}
