// Package repository 同城拼车出行数据访问层 - 统计
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StatisticListOptions 统计列表过滤条件
type StatisticListOptions struct {
	StatType  string
	UserID    uint
	StartDate *time.Time
	EndDate   *time.Time
}

// StatisticRepository 统计仓储接口
type StatisticRepository interface {
	Create(s *model.PincheStatistic) error
	FindByID(id uint) (*model.PincheStatistic, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts StatisticListOptions) ([]model.PincheStatistic, int64, error)
	FindByDateAndUser(regionID uint, statDate time.Time, statType string, userID uint) (*model.PincheStatistic, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheStatistic, int64, error)
	ListByDateRange(regionID uint, start, end time.Time, statType string) ([]model.PincheStatistic, error)

	Upsert(s *model.PincheStatistic) error
	IncrField(id uint, field string, delta int) error
	AddToFloatField(id uint, field string, delta float64) error

	SumByDateRange(regionID uint, start, end time.Time, statType string) (totalTrips, completedTrips, cancelledTrips int64, totalRevenue, totalRefund float64, err error)
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.PincheStatistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) FindByID(id uint) (*model.PincheStatistic, error) {
	var s model.PincheStatistic
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheStatistic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statisticRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheStatistic{}, id).Error
}

func (r *statisticRepository) List(regionID uint, pagination *utils.Pagination, opts StatisticListOptions) ([]model.PincheStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheStatistic
	var total int64

	query := r.db.Model(&model.PincheStatistic{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.StatType != "" {
		query = query.Where("stat_type = ?", opts.StatType)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.StartDate != nil {
		query = query.Where("stat_date >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		query = query.Where("stat_date <= ?", *opts.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) FindByDateAndUser(regionID uint, statDate time.Time, statType string, userID uint) (*model.PincheStatistic, error) {
	var s model.PincheStatistic
	q := r.db.Model(&model.PincheStatistic{}).
		Where("stat_date = ? AND stat_type = ?", statDate.Format("2006-01-02"), statType)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	} else {
		q = q.Where("user_id IS NULL")
	}
	if err := q.First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheStatistic
	var total int64

	query := r.db.Model(&model.PincheStatistic{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) ListByDateRange(regionID uint, start, end time.Time, statType string) ([]model.PincheStatistic, error) {
	var list []model.PincheStatistic
	q := r.db.Model(&model.PincheStatistic{}).
		Where("stat_date >= ? AND stat_date <= ?", start.Format("2006-01-02"), end.Format("2006-01-02"))
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if statType != "" {
		q = q.Where("stat_type = ?", statType)
	}
	if err := q.Order("stat_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) Upsert(s *model.PincheStatistic) error {
	return r.db.Save(s).Error
}

func (r *statisticRepository) IncrField(id uint, field string, delta int) error {
	return r.db.Model(&model.PincheStatistic{}).Where("id = ?", id).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

func (r *statisticRepository) AddToFloatField(id uint, field string, delta float64) error {
	return r.db.Model(&model.PincheStatistic{}).Where("id = ?", id).
		UpdateColumn(field, gorm.Expr(field+" + ?", delta)).Error
}

func (r *statisticRepository) SumByDateRange(regionID uint, start, end time.Time, statType string) (totalTrips, completedTrips, cancelledTrips int64, totalRevenue, totalRefund float64, err error) {
	var result struct {
		TotalTrips     int64   `gorm:"column:total_trips"`
		CompletedTrips int64   `gorm:"column:completed_trips"`
		CancelledTrips int64   `gorm:"column:cancelled_trips"`
		TotalRevenue   float64 `gorm:"column:total_revenue"`
		TotalRefund    float64 `gorm:"column:total_refund"`
	}
	q := r.db.Model(&model.PincheStatistic{}).
		Where("stat_date >= ? AND stat_date <= ?", start.Format("2006-01-02"), end.Format("2006-01-02"))
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if statType != "" {
		q = q.Where("stat_type = ?", statType)
	}
	err = q.Select(`
		COALESCE(SUM(total_trips), 0) AS total_trips,
		COALESCE(SUM(completed_trips), 0) AS completed_trips,
		COALESCE(SUM(cancelled_trips), 0) AS cancelled_trips,
		COALESCE(SUM(total_revenue), 0) AS total_revenue,
		COALESCE(SUM(total_refund), 0) AS total_refund
	`).Scan(&result).Error
	if err != nil {
		return
	}
	return result.TotalTrips, result.CompletedTrips, result.CancelledTrips, result.TotalRevenue, result.TotalRefund, nil
}
