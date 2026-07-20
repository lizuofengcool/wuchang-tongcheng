// Package repository 同城114数据访问层 - 举报
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ReportRepository 举报仓储接口
type ReportRepository interface {
	Create(r *model.Dh114Report) error
	FindByID(id uint) (*model.Dh114Report, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query ReportListQuery, pagination *utils.Pagination) ([]model.Dh114Report, int64, error)
	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	CountByStatus(regionID uint) (total, pending, processed int64, err error)
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	Keyword    string
	Status     *int
	ReportType string
	TargetType string
	TargetID   uint // 通用目标 ID（与 Dh114ID 等价，由 service 层映射）
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建举报仓储实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(rp *model.Dh114Report) error {
	return r.db.Create(rp).Error
}

func (r *reportRepository) FindByID(id uint) (*model.Dh114Report, error) {
	var rp model.Dh114Report
	if err := r.db.First(&rp, id).Error; err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *reportRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Report{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Report{}, id).Error
}

func (r *reportRepository) List(regionID uint, query ReportListQuery, pagination *utils.Pagination) ([]model.Dh114Report, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Report
	var total int64

	q := r.db.Model(&model.Dh114Report{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.TargetType != "" {
		q = q.Where("target_type = ?", query.TargetType)
	}
	if query.TargetID > 0 {
		q = q.Where("target_id = ?", query.TargetID)
	}
	if query.ReportType != "" {
		q = q.Where("report_type = ?", query.ReportType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("report_no ILIKE ? OR report_reason ILIKE ? OR description ILIKE ? OR reporter_name ILIKE ? OR target_name ILIKE ?", like, like, like, like, like)
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

func (r *reportRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["status"] = status
	return r.db.Model(&model.Dh114Report{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) CountByStatus(regionID uint) (int64, int64, int64, error) {
	type stat struct {
		Total     int64 `gorm:"column:total"`
		Pending   int64 `gorm:"column:pending"`
		Processed int64 `gorm:"column:processed"`
	}
	var s stat
	err := r.db.Model(&model.Dh114Report{}).
		Select("COUNT(*) AS total, "+
			"COUNT(CASE WHEN status = 0 THEN 1 END) AS pending, "+
			"COUNT(CASE WHEN status > 0 THEN 1 END) AS processed").
		Where("region_id = ? AND deleted_at IS NULL", regionID).
		Scan(&s).Error
	return s.Total, s.Pending, s.Processed, err
}
