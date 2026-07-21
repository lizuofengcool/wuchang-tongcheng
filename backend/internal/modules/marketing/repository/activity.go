// Package repository 营销活动中台数据访问层 - 营销活动（activity 子域）
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ActivityListQuery 活动列表查询
type ActivityListQuery struct {
	Type    string
	Status  *int
	Keyword string
}

// ActivityRepository 活动仓储接口
type ActivityRepository interface {
	Create(a *model.Activity) error
	FindByID(id uint) (*model.Activity, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query ActivityListQuery, pagination *utils.Pagination) ([]model.Activity, int64, error)
	ListOngoing(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error)
	ListUpcoming(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error)
	ListEnded(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error)
	UpdateStatusByTime(now time.Time) (int64, error)
}

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository 创建活动仓储实例
func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(a *model.Activity) error {
	return r.db.Create(a).Error
}

func (r *activityRepository) FindByID(id uint) (*model.Activity, error) {
	var a model.Activity
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *activityRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Activity{}).Where("id = ?", id).Updates(fields).Error
}

func (r *activityRepository) Delete(id uint) error {
	return r.db.Delete(&model.Activity{}, id).Error
}

func (r *activityRepository) List(regionID uint, query ActivityListQuery, pagination *utils.Pagination) ([]model.Activity, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Activity
	var total int64

	q := r.db.Model(&model.Activity{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.Type != "" {
		q = q.Where("type = ?", query.Type)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("start_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *activityRepository) ListOngoing(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Activity
	var total int64

	now := time.Now()
	q := r.db.Model(&model.Activity{}).
		Where("status = ?", model.ActivityStatusOngoing).
		Where("start_at IS NOT NULL AND start_at <= ?", now).
		Where("end_at IS NOT NULL AND end_at >= ?", now)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("end_at ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *activityRepository) ListUpcoming(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Activity
	var total int64

	now := time.Now()
	q := r.db.Model(&model.Activity{}).
		Where("status = ?", model.ActivityStatusPending).
		Where("start_at IS NOT NULL AND start_at > ?", now)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("start_at ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *activityRepository) ListEnded(regionID uint, pagination *utils.Pagination) ([]model.Activity, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Activity
	var total int64

	now := time.Now()
	q := r.db.Model(&model.Activity{}).
		Where("status IN ?", []int{model.ActivityStatusEnded, model.ActivityStatusCancelled}).
		Where("end_at IS NOT NULL AND end_at < ?", now)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("end_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// UpdateStatusByTime 根据时间自动推进活动状态
// - pending 且 start_at <= now → ongoing
// - ongoing 且 end_at < now → ended
// 返回受影响行数
func (r *activityRepository) UpdateStatusByTime(now time.Time) (int64, error) {
	// 待开始 → 进行中
	r1 := r.db.Model(&model.Activity{}).
		Where("status = ?", model.ActivityStatusPending).
		Where("start_at IS NOT NULL AND start_at <= ?", now).
		Update("status", model.ActivityStatusOngoing)
	if r1.Error != nil {
		return 0, r1.Error
	}
	affected := r1.RowsAffected

	// 进行中 → 已结束
	r2 := r.db.Model(&model.Activity{}).
		Where("status = ?", model.ActivityStatusOngoing).
		Where("end_at IS NOT NULL AND end_at < ?", now).
		Update("status", model.ActivityStatusEnded)
	if r2.Error != nil {
		return affected, r2.Error
	}
	affected += r2.RowsAffected
	return affected, nil
}
