// Package repository love 相亲交友数据访问层 - 举报
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveReportRepository 举报仓储接口
type LoveReportRepository interface {
	Create(r *model.LoveReport) error
	FindByID(id uint) (*model.LoveReport, error)
	FindByReportNo(no string) (*model.LoveReport, error)
	Update(r *model.LoveReport) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveReportListOptions) ([]model.LoveReport, int64, error)
	ListByReporter(userID uint, pagination *utils.Pagination) ([]model.LoveReport, int64, error)
	ListByTarget(userID uint, pagination *utils.Pagination) ([]model.LoveReport, int64, error)
	ListPending(pagination *utils.Pagination) ([]model.LoveReport, int64, error)

	Handle(id uint, req LoveReportHandleOptions) error
	UpdateAppeal(id uint, reason string) error
	HandleAppeal(id uint, result, remark string, handledBy uint) error
	UpdateRiskScore(id uint, score int) error

	CountByStatus(status int) (int64, error)
	CountToday() (int64, error)
	CountAppeals() (int64, error)
}

// LoveReportListOptions 举报列表过滤
type LoveReportListOptions struct {
	ReporterUserID uint
	TargetUserID   uint
	TargetType     string
	ReasonType     string
	Status         *int
}

// LoveReportHandleOptions 举报处理参数
type LoveReportHandleOptions struct {
	HandleResult    string
	HandleRemark    string
	PenaltyType     string
	PenaltyDuration int
	HandledBy       uint
}

type loveReportRepository struct {
	db *gorm.DB
}

// NewLoveReportRepository 创建举报仓储
func NewLoveReportRepository(db *gorm.DB) LoveReportRepository {
	return &loveReportRepository{db: db}
}

func (r *loveReportRepository) Create(rep *model.LoveReport) error {
	return r.db.Create(rep).Error
}

func (r *loveReportRepository) FindByID(id uint) (*model.LoveReport, error) {
	var rep model.LoveReport
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *loveReportRepository) FindByReportNo(no string) (*model.LoveReport, error) {
	var rep model.LoveReport
	if err := r.db.Where("report_no = ?", no).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *loveReportRepository) Update(rep *model.LoveReport) error {
	return r.db.Save(rep).Error
}

func (r *loveReportRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveReport{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveReportRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveReport{}, id).Error
}

func (r *loveReportRepository) List(pagination *utils.Pagination, opts LoveReportListOptions) ([]model.LoveReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveReport
	var total int64

	query := r.db.Model(&model.LoveReport{})
	if opts.ReporterUserID > 0 {
		query = query.Where("reporter_user_id = ?", opts.ReporterUserID)
	}
	if opts.TargetUserID > 0 {
		query = query.Where("target_user_id = ?", opts.TargetUserID)
	}
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.ReasonType != "" {
		query = query.Where("reason_type = ?", opts.ReasonType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveReportRepository) ListByReporter(userID uint, pagination *utils.Pagination) ([]model.LoveReport, int64, error) {
	return r.List(pagination, LoveReportListOptions{ReporterUserID: userID})
}

func (r *loveReportRepository) ListByTarget(userID uint, pagination *utils.Pagination) ([]model.LoveReport, int64, error) {
	return r.List(pagination, LoveReportListOptions{TargetUserID: userID})
}

func (r *loveReportRepository) ListPending(pagination *utils.Pagination) ([]model.LoveReport, int64, error) {
	status := model.ReportStatusPending
	return r.List(pagination, LoveReportListOptions{Status: &status})
}

func (r *loveReportRepository) Handle(id uint, opts LoveReportHandleOptions) error {
	updates := map[string]interface{}{
		"status":        model.ReportStatusHandled,
		"handled_by":    opts.HandledBy,
		"handled_at":    gorm.Expr("NOW()"),
		"handle_result": opts.HandleResult,
		"handle_remark": opts.HandleRemark,
		"penalty_type":  opts.PenaltyType,
		"penalty_duration": opts.PenaltyDuration,
	}
	if opts.PenaltyDuration > 0 {
		updates["penalty_expired_at"] = gorm.Expr("NOW() + (? || ' hours')::INTERVAL", opts.PenaltyDuration)
	}
	return r.db.Model(&model.LoveReport{}).Where("id = ?", id).Updates(updates).Error
}

func (r *loveReportRepository) UpdateAppeal(id uint, reason string) error {
	return r.db.Model(&model.LoveReport{}).Where("id = ?", id).Updates(map[string]interface{}{
		"appeal_status": model.AppealStatusPending,
		"appeal_reason": reason,
		"appealed_at":   gorm.Expr("NOW()"),
	}).Error
}

func (r *loveReportRepository) HandleAppeal(id uint, result, remark string, handledBy uint) error {
	status := model.AppealStatusRejected
	if result == "approved" {
		status = model.AppealStatusApproved
	}
	return r.db.Model(&model.LoveReport{}).Where("id = ?", id).Updates(map[string]interface{}{
		"appeal_status":      status,
		"appeal_handled_by":  handledBy,
		"appeal_handled_at":  gorm.Expr("NOW()"),
		"appeal_result":      result,
		"appeal_remark":      remark,
	}).Error
}

func (r *loveReportRepository) UpdateRiskScore(id uint, score int) error {
	return r.db.Model(&model.LoveReport{}).Where("id = ?", id).Update("risk_score", score).Error
}

func (r *loveReportRepository) CountByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveReport{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *loveReportRepository) CountToday() (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveReport{}).Where("created_at >= DATE_TRUNC('day', NOW())").Count(&count).Error
	return count, err
}

func (r *loveReportRepository) CountAppeals() (int64, error) {
	var count int64
	err := r.db.Model(&model.LoveReport{}).Where("appeal_status > ?", model.AppealStatusNone).Count(&count).Error
	return count, err
}
