// Package repository DIY 前端页面中台数据访问层 - 统计（stat 子域）
package repository

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/diy/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StatRangeOptions 按日期范围汇总选项
type StatRangeOptions struct {
	PageID    uint
	StartDate time.Time
	EndDate   time.Time
}

// StatRepository 统计仓储接口
type StatRepository interface {
	// CRUD
	Create(s *model.PageStat) error
	FindByID(id uint) (*model.PageStat, error)
	UpdateFields(id uint, fields map[string]interface{}) error

	// 查询
	FindByPageAndDate(pageID uint, date time.Time) (*model.PageStat, error)
	ListByPageID(pageID uint, pagination *utils.Pagination) ([]model.PageStat, int64, error)
	ListByDateRange(opts StatRangeOptions, pagination *utils.Pagination) ([]model.PageStat, int64, error)

	// 计数器
	IncrViewCount(pageID uint, date time.Time) error
	IncrClickCount(pageID uint, date time.Time) error
	IncrConversionCount(pageID uint, date time.Time) error

	// 汇总
	SumByPageID(pageID uint) (viewCount, clickCount, conversionCount int64, err error)
	SumByDateRange(opts StatRangeOptions) (viewCount, clickCount, conversionCount int64, err error)
}

type statRepository struct {
	db *gorm.DB
}

// NewStatRepository 创建统计仓储实例
func NewStatRepository(db *gorm.DB) StatRepository {
	return &statRepository{db: db}
}

func (r *statRepository) Create(s *model.PageStat) error {
	return r.db.Create(s).Error
}

func (r *statRepository) FindByID(id uint) (*model.PageStat, error) {
	var s model.PageStat
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PageStat{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statRepository) FindByPageAndDate(pageID uint, date time.Time) (*model.PageStat, error) {
	var s model.PageStat
	if err := r.db.Where("page_id = ? AND stat_date = ?", pageID, date.Format("2006-01-02")).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statRepository) ListByPageID(pageID uint, pagination *utils.Pagination) ([]model.PageStat, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 30)
	}
	var list []model.PageStat
	var total int64

	query := r.db.Model(&model.PageStat{}).Where("page_id = ?", pageID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statRepository) ListByDateRange(opts StatRangeOptions, pagination *utils.Pagination) ([]model.PageStat, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 30)
	}
	var list []model.PageStat
	var total int64

	query := r.db.Model(&model.PageStat{})
	if opts.PageID > 0 {
		query = query.Where("page_id = ?", opts.PageID)
	}
	if !opts.StartDate.IsZero() {
		query = query.Where("stat_date >= ?", opts.StartDate.Format("2006-01-02"))
	}
	if !opts.EndDate.IsZero() {
		query = query.Where("stat_date <= ?", opts.EndDate.Format("2006-01-02"))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("stat_date DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// upsertStatByDate 按日期 upsert 统计记录，存在则自增，不存在则创建
func (r *statRepository) upsertStatByDate(pageID uint, date time.Time, field string) error {
	dateStr := date.Format("2006-01-02")
	// 先尝试查找
	var existing model.PageStat
	err := r.db.Where("page_id = ? AND stat_date = ?", pageID, dateStr).First(&existing).Error
	if err == nil {
		// 存在，自增对应字段
		return r.db.Model(&model.PageStat{}).Where("id = ?", existing.ID).
			UpdateColumn(field, gorm.Expr(field+" + 1")).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 不存在，创建
	stat := &model.PageStat{
		PageID:   pageID,
		StatDate: date,
	}
	switch field {
	case "view_count":
		stat.ViewCount = 1
	case "click_count":
		stat.ClickCount = 1
	case "conversion_count":
		stat.ConversionCount = 1
	}
	return r.db.Create(stat).Error
}

func (r *statRepository) IncrViewCount(pageID uint, date time.Time) error {
	return r.upsertStatByDate(pageID, date, "view_count")
}

func (r *statRepository) IncrClickCount(pageID uint, date time.Time) error {
	return r.upsertStatByDate(pageID, date, "click_count")
}

func (r *statRepository) IncrConversionCount(pageID uint, date time.Time) error {
	return r.upsertStatByDate(pageID, date, "conversion_count")
}

func (r *statRepository) SumByPageID(pageID uint) (viewCount, clickCount, conversionCount int64, err error) {
	var result struct {
		ViewCount       int64
		ClickCount      int64
		ConversionCount int64
	}
	err = r.db.Model(&model.PageStat{}).
		Select("COALESCE(SUM(view_count), 0) AS view_count, COALESCE(SUM(click_count), 0) AS click_count, COALESCE(SUM(conversion_count), 0) AS conversion_count").
		Where("page_id = ?", pageID).
		Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.ViewCount, result.ClickCount, result.ConversionCount, nil
}

func (r *statRepository) SumByDateRange(opts StatRangeOptions) (viewCount, clickCount, conversionCount int64, err error) {
	var result struct {
		ViewCount       int64
		ClickCount      int64
		ConversionCount int64
	}
	query := r.db.Model(&model.PageStat{}).
		Select("COALESCE(SUM(view_count), 0) AS view_count, COALESCE(SUM(click_count), 0) AS click_count, COALESCE(SUM(conversion_count), 0) AS conversion_count")
	if opts.PageID > 0 {
		query = query.Where("page_id = ?", opts.PageID)
	}
	if !opts.StartDate.IsZero() {
		query = query.Where("stat_date >= ?", opts.StartDate.Format("2006-01-02"))
	}
	if !opts.EndDate.IsZero() {
		query = query.Where("stat_date <= ?", opts.EndDate.Format("2006-01-02"))
	}
	err = query.Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.ViewCount, result.ClickCount, result.ConversionCount, nil
}
