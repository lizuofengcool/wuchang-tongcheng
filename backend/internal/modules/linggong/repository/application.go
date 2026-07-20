// Package repository 同城零工兼职数据访问层 - 报名记录
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ApplicationRepository 报名记录仓储接口
type ApplicationRepository interface {
	Create(a *model.LinggongApplication) error
	FindByID(id uint) (*model.LinggongApplication, error)
	FindByApplicationNo(no string) (*model.LinggongApplication, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts ApplicationListOptions) ([]model.LinggongApplication, int64, error)
	AdminList(pagination *utils.Pagination, opts ApplicationAdminListOptions) ([]model.LinggongApplication, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error)
	ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error)
	ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error)
	CountByLinggong(linggongID uint, status int) (int64, error)
}

// ApplicationListOptions C 端报名列表过滤条件
type ApplicationListOptions struct {
	LinggongID uint
	TaskID     uint
	EmployerID uint
	WorkerID   uint
	Status     *int
	Source     string
	Keyword    string
}

// ApplicationAdminListOptions M 端报名列表过滤条件
type ApplicationAdminListOptions struct {
	RegionID   uint
	LinggongID uint
	EmployerID uint
	WorkerID   uint
	Status     *int
	Keyword    string
}

type applicationRepository struct {
	db *gorm.DB
}

// NewApplicationRepository 创建报名记录仓储实例
func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(a *model.LinggongApplication) error {
	return r.db.Create(a).Error
}

func (r *applicationRepository) FindByID(id uint) (*model.LinggongApplication, error) {
	var a model.LinggongApplication
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepository) FindByApplicationNo(no string) (*model.LinggongApplication, error) {
	var a model.LinggongApplication
	if err := r.db.Where("application_no = ?", no).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *applicationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongApplication{}).Where("id = ?", id).Updates(fields).Error
}

func (r *applicationRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongApplication{}, id).Error
}

func (r *applicationRepository) List(regionID uint, pagination *utils.Pagination, opts ApplicationListOptions) ([]model.LinggongApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongApplication
	var total int64

	query := r.db.Model(&model.LinggongApplication{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.TaskID > 0 {
		query = query.Where("task_id = ?", opts.TaskID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("application_no ILIKE ? OR worker_name ILIKE ?", like, like)
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

func (r *applicationRepository) AdminList(pagination *utils.Pagination, opts ApplicationAdminListOptions) ([]model.LinggongApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongApplication
	var total int64

	query := r.db.Model(&model.LinggongApplication{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("application_no ILIKE ? OR worker_name ILIKE ? OR employer_name ILIKE ?", like, like, like)
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

func (r *applicationRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongApplication
	var total int64
	query := r.db.Model(&model.LinggongApplication{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongApplication
	var total int64
	query := r.db.Model(&model.LinggongApplication{}).Where("worker_id = ?", workerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongApplication, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongApplication
	var total int64
	query := r.db.Model(&model.LinggongApplication{}).Where("employer_id = ?", employerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *applicationRepository) CountByLinggong(linggongID uint, status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.LinggongApplication{}).
		Where("linggong_id = ? AND status = ?", linggongID, status).
		Count(&count).Error
	return count, err
}
