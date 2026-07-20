// Package repository 同城拼车出行数据访问层 - 批量任务
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// BatchTaskListOptions 批量任务列表过滤条件
type BatchTaskListOptions struct {
	TaskType string
	Action   string
	Status   *int
	Keyword  string
}

// BatchTaskRepository 批量任务仓储接口
type BatchTaskRepository interface {
	Create(t *model.PincheBatchTask) error
	FindByID(id uint) (*model.PincheBatchTask, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts BatchTaskListOptions) ([]model.PincheBatchTask, int64, error)
	UpdateStatus(id uint, status int) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type batchTaskRepository struct {
	db *gorm.DB
}

// NewBatchTaskRepository 创建批量任务仓储实例
func NewBatchTaskRepository(db *gorm.DB) BatchTaskRepository {
	return &batchTaskRepository{db: db}
}

func (r *batchTaskRepository) Create(t *model.PincheBatchTask) error {
	return r.db.Create(t).Error
}

func (r *batchTaskRepository) FindByID(id uint) (*model.PincheBatchTask, error) {
	var t model.PincheBatchTask
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *batchTaskRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheBatchTask{}).Where("id = ?", id).Updates(fields).Error
}

func (r *batchTaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheBatchTask{}, id).Error
}

func (r *batchTaskRepository) List(regionID uint, pagination *utils.Pagination, opts BatchTaskListOptions) ([]model.PincheBatchTask, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheBatchTask
	var total int64

	query := r.db.Model(&model.PincheBatchTask{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.TaskType != "" {
		query = query.Where("task_type = ?", opts.TaskType)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		query = query.Where("task_name ILIKE ? OR task_no ILIKE ? OR operator_name ILIKE ?",
			"%"+opts.Keyword+"%", "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *batchTaskRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheBatchTask{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *batchTaskRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheBatchTask{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
