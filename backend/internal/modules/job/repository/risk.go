// Package repository 举报/评价/审核规则数据访问层
// 依据 v3.2.1 架构方案：对标 BOSS直聘/看准
package repository

import (
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Report =====

// ReportRepository 举报仓储接口
type ReportRepository interface {
	Create(r *model.JobReport) error
	FindByID(id uint) (*model.JobReport, error)
	FindByReportNo(reportNo string) (*model.JobReport, error)
	Update(id uint, fields map[string]interface{}) error
	List(query ReportListQuery, pagination *utils.Pagination) ([]model.JobReport, int64, error)
	ListByTarget(targetType string, targetID uint) ([]model.JobReport, error)
	CountByReporter(reporterID uint) (int64, error)
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	Status     *int
	ReportType string
	TargetType string
	TargetID   uint
	ReporterID uint
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建举报仓储实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(rep *model.JobReport) error {
	return r.db.Create(rep).Error
}

func (r *reportRepository) FindByID(id uint) (*model.JobReport, error) {
	var rep model.JobReport
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) FindByReportNo(reportNo string) (*model.JobReport, error) {
	var rep model.JobReport
	if err := r.db.Where("report_no = ?", reportNo).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobReport{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) List(query ReportListQuery, pagination *utils.Pagination) ([]model.JobReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobReport
	var total int64

	q := r.db.Model(&model.JobReport{})
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.ReportType != "" {
		q = q.Where("report_type = ?", query.ReportType)
	}
	if query.TargetType != "" {
		q = q.Where("target_type = ?", query.TargetType)
	}
	if query.TargetID > 0 {
		q = q.Where("target_id = ?", query.TargetID)
	}
	if query.ReporterID > 0 {
		q = q.Where("reporter_id = ?", query.ReporterID)
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

func (r *reportRepository) ListByTarget(targetType string, targetID uint) ([]model.JobReport, error) {
	var list []model.JobReport
	if err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportRepository) CountByReporter(reporterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.JobReport{}).Where("reporter_id = ?", reporterID).Count(&count).Error
	return count, err
}

// ===== Review =====

// ReviewRepository 公司评价仓储接口
type ReviewRepository interface {
	Create(r *model.JobReview) error
	FindByID(id uint) (*model.JobReview, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query ReviewListQuery, pagination *utils.Pagination) ([]model.JobReview, int64, error)
	ListByCompanyID(companyID uint, pagination *utils.Pagination) ([]model.JobReview, int64, error)
	StatsByCompanyID(companyID uint) (total int64, avgRating float64, good, medium, bad int64, recommendCount int64, err error)
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	CompanyID  uint
	ReviewerID uint
	ReviewType string
	Rating     *int
	Status     *int
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建评价仓储实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.JobReview) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByID(id uint) (*model.JobReview, error) {
	var rv model.JobReview
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobReview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobReview{}, id).Error
}

func (r *reviewRepository) List(query ReviewListQuery, pagination *utils.Pagination) ([]model.JobReview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.JobReview
	var total int64

	q := r.db.Model(&model.JobReview{})
	if query.CompanyID > 0 {
		q = q.Where("company_id = ?", query.CompanyID)
	}
	if query.ReviewerID > 0 {
		q = q.Where("reviewer_id = ?", query.ReviewerID)
	}
	if query.ReviewType != "" {
		q = q.Where("review_type = ?", query.ReviewType)
	}
	if query.Rating != nil {
		q = q.Where("rating = ?", *query.Rating)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.ReviewStatusApproved)
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

func (r *reviewRepository) ListByCompanyID(companyID uint, pagination *utils.Pagination) ([]model.JobReview, int64, error) {
	return r.List(ReviewListQuery{CompanyID: companyID}, pagination)
}

func (r *reviewRepository) StatsByCompanyID(companyID uint) (int64, float64, int64, int64, int64, int64, error) {
	type stat struct {
		Total      int64   `gorm:"column:total"`
		AvgRate    float64 `gorm:"column:avg_rate"`
		Good       int64   `gorm:"column:good"`
		Medium     int64   `gorm:"column:medium"`
		Bad        int64   `gorm:"column:bad"`
		Recommend  int64   `gorm:"column:recommend"`
	}
	var s stat
	err := r.db.Model(&model.JobReview{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate, "+
			"COUNT(CASE WHEN rating >= 4 THEN 1 END) AS good, "+
			"COUNT(CASE WHEN rating = 3 THEN 1 END) AS medium, "+
			"COUNT(CASE WHEN rating <= 2 THEN 1 END) AS bad, "+
			"COUNT(CASE WHEN is_recommended = true THEN 1 END) AS recommend").
		Where("company_id = ? AND status = ?", companyID, model.ReviewStatusApproved).
		Scan(&s).Error
	return s.Total, s.AvgRate, s.Good, s.Medium, s.Bad, s.Recommend, err
}

// ===== AuditRule =====

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(r *model.JobAuditRule) error
	FindByID(id uint) (*model.JobAuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.JobAuditRule, int64, error)
	ListEnabled() ([]model.JobAuditRule, error)
}

// AuditRuleListQuery 审核规则列表查询
type AuditRuleListQuery struct {
	RuleType string
	Status   *int
}

type auditRuleRepository struct {
	db *gorm.DB
}

// NewAuditRuleRepository 创建审核规则仓储实例
func NewAuditRuleRepository(db *gorm.DB) AuditRuleRepository {
	return &auditRuleRepository{db: db}
}

func (r *auditRuleRepository) Create(rule *model.JobAuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.JobAuditRule, error) {
	var rule model.JobAuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.JobAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.JobAuditRule{}, id).Error
}

func (r *auditRuleRepository) List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.JobAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.JobAuditRule
	var total int64

	q := r.db.Model(&model.JobAuditRule{})
	if query.RuleType != "" {
		q = q.Where("rule_type = ?", query.RuleType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auditRuleRepository) ListEnabled() ([]model.JobAuditRule, error) {
	var list []model.JobAuditRule
	if err := r.db.Where("status = ?", model.AuditRuleStatusEnabled).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
