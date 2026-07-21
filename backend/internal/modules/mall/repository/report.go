// Package repository 同城商城数据访问层 - 举报
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ReportRepository 举报仓储接口
type ReportRepository interface {
	Create(r *model.Report) error
	FindByID(id uint) (*model.Report, error)
	FindByReportNo(reportNo string) (*model.Report, error)
	Update(r *model.Report) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(opts ReportListOptions, pagination *utils.Pagination) ([]model.Report, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Report, int64, error)
	ListByTarget(targetType string, targetID uint) ([]model.Report, error)

	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	Stats(opts ReportStatsOptions) (*ReportStatsResult, error)
}

// ReportListOptions 举报列表过滤条件
type ReportListOptions struct {
	TargetType string
	TargetID   uint
	ReporterID uint
	HandlerID  uint
	Status     *int
	ReportType string
	Keyword    string
	StartDate  string
	EndDate    string
	RegionID   uint
}

// ReportStatsOptions 举报统计选项
type ReportStatsOptions struct {
	RegionID  uint
	StartDate string
	EndDate   string
}

// ReportStatsResult 举报统计结果
type ReportStatsResult struct {
	Total     int64 `gorm:"column:total" json:"total"`
	Pending   int64 `gorm:"column:pending" json:"pending"`
	Processed int64 `gorm:"column:processed" json:"processed"`
	Valid     int64 `gorm:"column:valid" json:"valid"`
	Invalid   int64 `gorm:"column:invalid" json:"invalid"`
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建举报仓储实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(rep *model.Report) error {
	return r.db.Create(rep).Error
}

func (r *reportRepository) FindByID(id uint) (*model.Report, error) {
	var rep model.Report
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) FindByReportNo(reportNo string) (*model.Report, error) {
	var rep model.Report
	if err := r.db.Where("report_no = ?", reportNo).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) Update(rep *model.Report) error {
	return r.db.Save(rep).Error
}

func (r *reportRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Report{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) Delete(id uint) error {
	return r.db.Delete(&model.Report{}, id).Error
}

func (r *reportRepository) List(opts ReportListOptions, pagination *utils.Pagination) ([]model.Report, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Report
	var total int64

	query := r.db.Model(&model.Report{})
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.ReporterID > 0 {
		query = query.Where("reporter_id = ?", opts.ReporterID)
	}
	if opts.HandlerID > 0 {
		query = query.Where("handler_id = ?", opts.HandlerID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.ReportType != "" {
		query = query.Where("report_type = ?", opts.ReportType)
	}
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.StartDate != "" {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("created_at <= ?", opts.EndDate)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("report_no ILIKE ? OR target_name ILIKE ? OR report_reason ILIKE ? OR description ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reportRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Report, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Report
	var total int64

	query := r.db.Model(&model.Report{}).Where("reporter_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reportRepository) ListByTarget(targetType string, targetID uint) ([]model.Report, error) {
	var list []model.Report
	if err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Report{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) Stats(opts ReportStatsOptions) (*ReportStatsResult, error) {
	var result ReportStatsResult
	query := r.db.Model(&model.Report{}).Select(`
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE status = 0) AS pending,
		COUNT(*) FILTER (WHERE status > 0) AS processed,
		COUNT(*) FILTER (WHERE status IN (1, 2, 3, 5)) AS valid,
		COUNT(*) FILTER (WHERE status = 4) AS invalid
	`)
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.StartDate != "" {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("created_at <= ?", opts.EndDate)
	}
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
