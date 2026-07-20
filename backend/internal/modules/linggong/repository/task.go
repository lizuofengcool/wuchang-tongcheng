// Package repository 同城零工兼职数据访问层 - 任务包
// 任务包表对标斗米任务制 + 猪八戒威客
// 提供任务领取/交付/验收状态机所需的计数器
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TaskRepository 任务包仓储接口
type TaskRepository interface {
	Create(t *model.LinggongTask) error
	FindByID(id uint) (*model.LinggongTask, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表查询
	List(pagination *utils.Pagination, opts TaskListOptions) ([]model.LinggongTask, int64, error)
	AdminList(pagination *utils.Pagination, opts TaskAdminListOptions) ([]model.LinggongTask, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongTask, int64, error)
	ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongTask, int64, error)

	// 计数器（任务领取/交付/验收）
	IncrClaimedCount(id uint, count int) error
	IncrCompletedCount(id uint, count int) error
	IncrVerifiedCount(id uint, count int) error
	IncrPaidAmount(id uint, amount float64) error
}

// TaskListOptions C 端任务列表过滤条件
type TaskListOptions struct {
	LinggongID uint
	EmployerID uint
	TaskType   string
	Difficulty string
	Status     *int
	Keyword    string
}

// TaskAdminListOptions M 端任务列表过滤条件
type TaskAdminListOptions struct {
	RegionID   uint
	LinggongID uint
	EmployerID uint
	TaskType   string
	Status     *int
	Keyword    string
}

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务包仓储实例
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(t *model.LinggongTask) error {
	return r.db.Create(t).Error
}

func (r *taskRepository) FindByID(id uint) (*model.LinggongTask, error) {
	var t model.LinggongTask
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *taskRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongTask{}).Where("id = ?", id).Updates(fields).Error
}

func (r *taskRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongTask{}, id).Error
}

func (r *taskRepository) List(pagination *utils.Pagination, opts TaskListOptions) ([]model.LinggongTask, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongTask
	var total int64

	q := r.db.Model(&model.LinggongTask{})
	if opts.LinggongID > 0 {
		q = q.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.EmployerID > 0 {
		q = q.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.TaskType != "" {
		q = q.Where("task_type = ?", opts.TaskType)
	}
	if opts.Difficulty != "" {
		q = q.Where("difficulty = ?", opts.Difficulty)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ? OR task_no ILIKE ?", like, like, like)
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

func (r *taskRepository) AdminList(pagination *utils.Pagination, opts TaskAdminListOptions) ([]model.LinggongTask, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongTask
	var total int64

	q := r.db.Model(&model.LinggongTask{})
	if opts.RegionID > 0 {
		q = q.Where("region_id = ?", opts.RegionID)
	}
	if opts.LinggongID > 0 {
		q = q.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.EmployerID > 0 {
		q = q.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.TaskType != "" {
		q = q.Where("task_type = ?", opts.TaskType)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ? OR task_no ILIKE ? OR employer_name ILIKE ?", like, like, like, like)
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

func (r *taskRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongTask, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongTask
	var total int64

	q := r.db.Model(&model.LinggongTask{}).Where("linggong_id = ?", linggongID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *taskRepository) ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongTask, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongTask
	var total int64

	q := r.db.Model(&model.LinggongTask{}).Where("employer_id = ?", employerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// IncrClaimedCount 增加已领取数（支持参数化增量）
func (r *taskRepository) IncrClaimedCount(id uint, count int) error {
	return r.db.Model(&model.LinggongTask{}).Where("id = ?", id).
		UpdateColumn("claimed_count", gorm.Expr("claimed_count + ?", count)).Error
}

// IncrCompletedCount 增加已完成数（支持参数化增量）
func (r *taskRepository) IncrCompletedCount(id uint, count int) error {
	return r.db.Model(&model.LinggongTask{}).Where("id = ?", id).
		UpdateColumn("completed_count", gorm.Expr("completed_count + ?", count)).Error
}

// IncrVerifiedCount 增加已验收数（支持参数化增量）
func (r *taskRepository) IncrVerifiedCount(id uint, count int) error {
	return r.db.Model(&model.LinggongTask{}).Where("id = ?", id).
		UpdateColumn("verified_count", gorm.Expr("verified_count + ?", count)).Error
}

// IncrPaidAmount 增加已支付金额（支持参数化增量）
func (r *taskRepository) IncrPaidAmount(id uint, amount float64) error {
	return r.db.Model(&model.LinggongTask{}).Where("id = ?", id).
		UpdateColumn("paid_amount", gorm.Expr("paid_amount + ?", amount)).Error
}
