// Package repository 同城零工兼职数据访问层 - 数据统计
// LinggongStatistic 数据统计表（RegionBaseModel，含 region_id）
// 唯一索引：uniq_linggong_stats_date_type_target (stat_date, stat_type, target_id)
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StatisticRepository 数据统计仓储接口
type StatisticRepository interface {
	Create(s *model.LinggongStatistic) error
	FindByID(id uint) (*model.LinggongStatistic, error)
	// 按唯一键查询（stat_date + stat_type + target_id）
	FindByDateTypeTarget(statDate time.Time, statType string, targetID uint) (*model.LinggongStatistic, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表查询
	List(query StatListQuery, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error)
	// 按日期范围查询某 target 的统计
	ListByTarget(statType string, targetID uint, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error)
	// 按统计类型查询某地区聚合
	ListByType(regionID uint, statType string, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error)

	// 唯一键 upsert（按 stat_date + stat_type + target_id 冲突时累加计数字段）
	Upsert(s *model.LinggongStatistic) error
	// 批量 upsert
	BatchUpsert(list []model.LinggongStatistic) error

	// 汇总查询
	// SumByRange 按统计类型 + 日期范围 + 目标 ID 汇总各项指标
	SumByRange(statType string, targetID uint, startDate, endDate time.Time) (*StatSummary, error)
}

// StatListQuery 统计列表查询条件
type StatListQuery struct {
	RegionID  uint
	StatType  string
	TargetID  uint
	StartDate time.Time
	EndDate   time.Time
	Keyword   string
}

// StatSummary 统计汇总结果
type StatSummary struct {
	ImpressionCount  int64   `gorm:"column:impression_count" json:"impression_count"`
	ClickCount       int64   `gorm:"column:click_count" json:"click_count"`
	FavCount         int64   `gorm:"column:fav_count" json:"fav_count"`
	ContactCount     int64   `gorm:"column:contact_count" json:"contact_count"`
	ApplicationCount int64   `gorm:"column:application_count" json:"application_count"`
	HiredCount       int64   `gorm:"column:hired_count" json:"hired_count"`
	CompletedCount   int64   `gorm:"column:completed_count" json:"completed_count"`
	DealCount        int64   `gorm:"column:deal_count" json:"deal_count"`
	TotalSalary      float64 `gorm:"column:total_salary" json:"total_salary"`
	AvgSalary        float64 `gorm:"column:avg_salary" json:"avg_salary"`
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.LinggongStatistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) FindByID(id uint) (*model.LinggongStatistic, error) {
	var s model.LinggongStatistic
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) FindByDateTypeTarget(statDate time.Time, statType string, targetID uint) (*model.LinggongStatistic, error) {
	var s model.LinggongStatistic
	if err := r.db.Where("stat_date = ? AND stat_type = ? AND target_id = ?", statDate, statType, targetID).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongStatistic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statisticRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongStatistic{}, id).Error
}

func (r *statisticRepository) List(query StatListQuery, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.LinggongStatistic
	var total int64

	q := r.db.Model(&model.LinggongStatistic{})
	if query.RegionID > 0 {
		q = q.Where("region_id = ?", query.RegionID)
	}
	if query.StatType != "" {
		q = q.Where("stat_type = ?", query.StatType)
	}
	if query.TargetID > 0 {
		q = q.Where("target_id = ?", query.TargetID)
	}
	if !query.StartDate.IsZero() {
		q = q.Where("stat_date >= ?", query.StartDate)
	}
	if !query.EndDate.IsZero() {
		q = q.Where("stat_date <= ?", query.EndDate)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("target_name ILIKE ?", like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) ListByTarget(statType string, targetID uint, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 30)
	}
	var list []model.LinggongStatistic
	var total int64

	q := r.db.Model(&model.LinggongStatistic{}).
		Where("stat_type = ? AND target_id = ?", statType, targetID)
	if !startDate.IsZero() {
		q = q.Where("stat_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		q = q.Where("stat_date <= ?", endDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) ListByType(regionID uint, statType string, startDate, endDate time.Time, pagination *utils.Pagination) ([]model.LinggongStatistic, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 30)
	}
	var list []model.LinggongStatistic
	var total int64

	q := r.db.Model(&model.LinggongStatistic{}).
		Where("stat_type = ?", statType)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if !startDate.IsZero() {
		q = q.Where("stat_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		q = q.Where("stat_date <= ?", endDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Upsert 按 (stat_date, stat_type, target_id) 唯一索引冲突时累加主要计数字段
// 注意：使用 ON CONFLICT 子句，依赖迁移脚本中创建的 uniq_linggong_stats_date_type_target 唯一索引
func (r *statisticRepository) Upsert(s *model.LinggongStatistic) error {
	return r.db.Exec(`
		INSERT INTO linggong_statistics
			(region_id, stat_date, stat_type, target_id, target_name,
			 impression_count, click_count, fav_count, contact_count,
			 application_count, hired_count, completed_count, deal_count,
			 conversion_rate, total_salary, avg_salary, avg_deal_days,
			 created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (stat_date, stat_type, target_id) DO UPDATE SET
			impression_count  = linggong_statistics.impression_count + EXCLUDED.impression_count,
			click_count       = linggong_statistics.click_count + EXCLUDED.click_count,
			fav_count         = linggong_statistics.fav_count + EXCLUDED.fav_count,
			contact_count     = linggong_statistics.contact_count + EXCLUDED.contact_count,
			application_count = linggong_statistics.application_count + EXCLUDED.application_count,
			hired_count       = linggong_statistics.hired_count + EXCLUDED.hired_count,
			completed_count   = linggong_statistics.completed_count + EXCLUDED.completed_count,
			deal_count        = linggong_statistics.deal_count + EXCLUDED.deal_count,
			total_salary      = linggong_statistics.total_salary + EXCLUDED.total_salary,
			updated_at        = NOW()
	`,
		s.RegionID, s.StatDate, s.StatType, s.TargetID, s.TargetName,
		s.ImpressionCount, s.ClickCount, s.FavCount, s.ContactCount,
		s.ApplicationCount, s.HiredCount, s.CompletedCount, s.DealCount,
		s.ConversionRate, s.TotalSalary, s.AvgSalary, s.AvgDealDays,
	).Error
}

// BatchUpsert 批量 upsert
func (r *statisticRepository) BatchUpsert(list []model.LinggongStatistic) error {
	if len(list) == 0 {
		return nil
	}
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for i := range list {
		if err := tx.Exec(`
			INSERT INTO linggong_statistics
				(region_id, stat_date, stat_type, target_id, target_name,
				 impression_count, click_count, fav_count, contact_count,
				 application_count, hired_count, completed_count, deal_count,
				 conversion_rate, total_salary, avg_salary, avg_deal_days,
				 created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON CONFLICT (stat_date, stat_type, target_id) DO UPDATE SET
				impression_count  = linggong_statistics.impression_count + EXCLUDED.impression_count,
				click_count       = linggong_statistics.click_count + EXCLUDED.click_count,
				fav_count         = linggong_statistics.fav_count + EXCLUDED.fav_count,
				contact_count     = linggong_statistics.contact_count + EXCLUDED.contact_count,
				application_count = linggong_statistics.application_count + EXCLUDED.application_count,
				hired_count       = linggong_statistics.hired_count + EXCLUDED.hired_count,
				completed_count   = linggong_statistics.completed_count + EXCLUDED.completed_count,
				deal_count        = linggong_statistics.deal_count + EXCLUDED.deal_count,
				total_salary      = linggong_statistics.total_salary + EXCLUDED.total_salary,
				updated_at        = NOW()
		`,
			list[i].RegionID, list[i].StatDate, list[i].StatType, list[i].TargetID, list[i].TargetName,
			list[i].ImpressionCount, list[i].ClickCount, list[i].FavCount, list[i].ContactCount,
			list[i].ApplicationCount, list[i].HiredCount, list[i].CompletedCount, list[i].DealCount,
			list[i].ConversionRate, list[i].TotalSalary, list[i].AvgSalary, list[i].AvgDealDays,
		).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// SumByRange 按统计类型 + 日期范围 + 目标 ID 汇总各项指标
func (r *statisticRepository) SumByRange(statType string, targetID uint, startDate, endDate time.Time) (*StatSummary, error) {
	var s StatSummary
	q := r.db.Model(&model.LinggongStatistic{}).
		Select(`
			COALESCE(SUM(impression_count),0) AS impression_count,
			COALESCE(SUM(click_count),0) AS click_count,
			COALESCE(SUM(fav_count),0) AS fav_count,
			COALESCE(SUM(contact_count),0) AS contact_count,
			COALESCE(SUM(application_count),0) AS application_count,
			COALESCE(SUM(hired_count),0) AS hired_count,
			COALESCE(SUM(completed_count),0) AS completed_count,
			COALESCE(SUM(deal_count),0) AS deal_count,
			COALESCE(SUM(total_salary),0) AS total_salary,
			COALESCE(AVG(avg_salary),0) AS avg_salary
		`).
		Where("stat_type = ?", statType)
	if targetID > 0 {
		q = q.Where("target_id = ?", targetID)
	}
	if !startDate.IsZero() {
		q = q.Where("stat_date >= ?", startDate)
	}
	if !endDate.IsZero() {
		q = q.Where("stat_date <= ?", endDate)
	}
	if err := q.Scan(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
