// Package repository 同城零工兼职数据访问层 - 考勤打卡
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AttendanceRepository 考勤打卡仓储接口
type AttendanceRepository interface {
	Create(a *model.LinggongAttendance) error
	FindByID(id uint) (*model.LinggongAttendance, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts AttendanceListOptions) ([]model.LinggongAttendance, int64, error)
	ListByApplication(applicationID uint, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error)
	ListByWorker(workerID uint, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error)
}

// AttendanceListOptions 考勤列表过滤条件
type AttendanceListOptions struct {
	ApplicationID  uint
	LinggongID    uint
	WorkerID      uint
	EmployerID    uint
	AttendanceType string
	Status        string
}

type attendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository 创建考勤打卡仓储实例
func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) Create(a *model.LinggongAttendance) error {
	return r.db.Create(a).Error
}

func (r *attendanceRepository) FindByID(id uint) (*model.LinggongAttendance, error) {
	var a model.LinggongAttendance
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *attendanceRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongAttendance{}).Where("id = ?", id).Updates(fields).Error
}

func (r *attendanceRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongAttendance{}, id).Error
}

func (r *attendanceRepository) List(regionID uint, pagination *utils.Pagination, opts AttendanceListOptions) ([]model.LinggongAttendance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongAttendance
	var total int64

	query := r.db.Model(&model.LinggongAttendance{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.ApplicationID > 0 {
		query = query.Where("application_id = ?", opts.ApplicationID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.AttendanceType != "" {
		query = query.Where("attendance_type = ?", opts.AttendanceType)
	}
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("clock_time DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *attendanceRepository) ListByApplication(applicationID uint, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongAttendance
	var total int64
	query := r.db.Model(&model.LinggongAttendance{}).Where("application_id = ?", applicationID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("clock_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *attendanceRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongAttendance
	var total int64
	query := r.db.Model(&model.LinggongAttendance{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("clock_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *attendanceRepository) ListByWorker(workerID uint, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongAttendance, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongAttendance
	var total int64
	query := r.db.Model(&model.LinggongAttendance{}).
		Where("worker_id = ? AND clock_time BETWEEN ? AND ?", workerID, startDate, endDate)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("clock_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
