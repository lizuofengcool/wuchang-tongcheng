// Package repository 同城车辆买卖数据访问层 - 举报/评价/审核规则
// CarReport 举报（全局，BaseModel 无 region_id）
// CarReview 评价（RegionBaseModel，含 region_id）
// CarAuditRule 审核规则（全局，BaseModel 无 region_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Report 举报 =====

// ReportRepository 举报仓储接口
type ReportRepository interface {
	Create(r *model.CarReport) error
	FindByID(id uint) (*model.CarReport, error)
	FindByReportNo(no string) (*model.CarReport, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（管理后台）
	List(query ReportListQuery, pagination *utils.Pagination) ([]model.CarReport, int64, error)
	// 按目标反查（某车源被举报记录）
	ListByTarget(targetType string, targetID uint) ([]model.CarReport, error)
	// 按举报人反查
	ListByReporter(reporterID uint, pagination *utils.Pagination) ([]model.CarReport, int64, error)
	// 按被举报人反查
	ListByReportedUser(reportedUserID uint, pagination *utils.Pagination) ([]model.CarReport, int64, error)

	// 统计
	CountByStatus(status int) (int64, error)
	CountByReporter(reporterID uint) (int64, error)
	CountPendingReports() (int64, error)
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	TargetType string
	TargetID   uint
	ReporterID uint
	ReportType string
	Status     *int
	Keyword    string
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建举报仓储实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(rep *model.CarReport) error {
	return r.db.Create(rep).Error
}

func (r *reportRepository) FindByID(id uint) (*model.CarReport, error) {
	var rep model.CarReport
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) FindByReportNo(no string) (*model.CarReport, error) {
	var rep model.CarReport
	if err := r.db.Where("report_no = ?", no).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarReport{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarReport{}, id).Error
}

func (r *reportRepository) List(query ReportListQuery, pagination *utils.Pagination) ([]model.CarReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarReport
	var total int64

	q := r.db.Model(&model.CarReport{})
	if query.TargetType != "" {
		q = q.Where("target_type = ?", query.TargetType)
	}
	if query.TargetID > 0 {
		q = q.Where("target_id = ?", query.TargetID)
	}
	if query.ReporterID > 0 {
		q = q.Where("reporter_id = ?", query.ReporterID)
	}
	if query.ReportType != "" {
		q = q.Where("report_type = ?", query.ReportType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("report_no ILIKE ? OR reporter_name ILIKE ? OR reported_user_name ILIKE ? OR reason ILIKE ?", like, like, like, like)
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

func (r *reportRepository) ListByTarget(targetType string, targetID uint) ([]model.CarReport, error) {
	var list []model.CarReport
	if err := r.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportRepository) ListByReporter(reporterID uint, pagination *utils.Pagination) ([]model.CarReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarReport
	var total int64

	q := r.db.Model(&model.CarReport{}).Where("reporter_id = ?", reporterID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reportRepository) ListByReportedUser(reportedUserID uint, pagination *utils.Pagination) ([]model.CarReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarReport
	var total int64

	q := r.db.Model(&model.CarReport{}).Where("reported_user_id = ?", reportedUserID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reportRepository) CountByStatus(status int) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarReport{}).Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportRepository) CountByReporter(reporterID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarReport{}).Where("reporter_id = ?", reporterID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportRepository) CountPendingReports() (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarReport{}).Where("status = ?", model.ReportStatusPending).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ===== Review 评价 =====

// ReviewRepository 评价仓储接口
type ReviewRepository interface {
	Create(r *model.CarReview) error
	FindByID(id uint) (*model.CarReview, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(regionID uint, query ReviewListQuery, pagination *utils.Pagination) ([]model.CarReview, int64, error)
	// 按目标反查评价（某车源/车商/销售的评价）
	ListByTarget(regionID uint, targetType string, targetID uint, pagination *utils.Pagination) ([]model.CarReview, int64, error)
	// 按评价人反查
	ListByReviewer(reviewerID uint, pagination *utils.Pagination) ([]model.CarReview, int64, error)

	// 评价统计
	StatsByTarget(targetType string, targetID uint) (total int64, avgRating float64, good, medium, bad int64, err error)
	StatsByReviewer(reviewerID uint) (total int64, avgRating float64, err error)
	HasReviewed(reviewerID, targetID uint, targetType string) (bool, error)

	// 互动
	IncrLikeCount(id uint) error
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	TargetType string
	TargetID   uint
	ReviewerID uint
	ReviewType string
	Rating     *int
	Status     *int
	HasReply   *bool
	Keyword    string
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建评价仓储实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.CarReview) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByID(id uint) (*model.CarReview, error) {
	var rv model.CarReview
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarReview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarReview{}, id).Error
}

func (r *reviewRepository) List(regionID uint, query ReviewListQuery, pagination *utils.Pagination) ([]model.CarReview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarReview
	var total int64

	q := r.db.Model(&model.CarReview{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.TargetType != "" {
		q = q.Where("target_type = ?", query.TargetType)
	}
	if query.TargetID > 0 {
		q = q.Where("target_id = ?", query.TargetID)
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
	}
	if query.HasReply != nil {
		if *query.HasReply {
			q = q.Where("reply <> ''")
		} else {
			q = q.Where("reply = ''")
		}
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("content ILIKE ? OR reviewer_name ILIKE ? OR reply ILIKE ?", like, like, like)
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

func (r *reviewRepository) ListByTarget(regionID uint, targetType string, targetID uint, pagination *utils.Pagination) ([]model.CarReview, int64, error) {
	return r.List(regionID, ReviewListQuery{
		TargetType: targetType,
		TargetID:   targetID,
		Status:     intPtr(model.ReviewStatusApproved),
	}, pagination)
}

func (r *reviewRepository) ListByReviewer(reviewerID uint, pagination *utils.Pagination) ([]model.CarReview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarReview
	var total int64

	q := r.db.Model(&model.CarReview{}).Where("reviewer_id = ?", reviewerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reviewRepository) StatsByTarget(targetType string, targetID uint) (int64, float64, int64, int64, int64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
		Good    int64   `gorm:"column:good"`
		Medium  int64   `gorm:"column:medium"`
		Bad     int64   `gorm:"column:bad"`
	}
	var s stat
	err := r.db.Model(&model.CarReview{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate, "+
			"COUNT(CASE WHEN rating >= 4 THEN 1 END) AS good, "+
			"COUNT(CASE WHEN rating = 3 THEN 1 END) AS medium, "+
			"COUNT(CASE WHEN rating <= 2 THEN 1 END) AS bad").
		Where("target_type = ? AND target_id = ? AND status = ?", targetType, targetID, model.ReviewStatusApproved).
		Scan(&s).Error
	return s.Total, s.AvgRate, s.Good, s.Medium, s.Bad, err
}

func (r *reviewRepository) StatsByReviewer(reviewerID uint) (int64, float64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
	}
	var s stat
	err := r.db.Model(&model.CarReview{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate").
		Where("reviewer_id = ? AND status = ?", reviewerID, model.ReviewStatusApproved).
		Scan(&s).Error
	return s.Total, s.AvgRate, err
}

func (r *reviewRepository) HasReviewed(reviewerID, targetID uint, targetType string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.CarReview{}).
		Where("reviewer_id = ? AND target_id = ? AND target_type = ?", reviewerID, targetID, targetType).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *reviewRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.CarReview{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// ===== AuditRule 审核规则 =====

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(r *model.CarAuditRule) error
	FindByID(id uint) (*model.CarAuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.CarAuditRule, int64, error)
	ListEnabled() ([]model.CarAuditRule, error)
	ListByRuleType(ruleType string) ([]model.CarAuditRule, error)
}

// AuditRuleListQuery 审核规则列表查询
type AuditRuleListQuery struct {
	RuleType string
	RuleKey  string
	Action   string
	Status   *int
	Keyword  string
}

type auditRuleRepository struct {
	db *gorm.DB
}

// NewAuditRuleRepository 创建审核规则仓储实例
func NewAuditRuleRepository(db *gorm.DB) AuditRuleRepository {
	return &auditRuleRepository{db: db}
}

func (r *auditRuleRepository) Create(rule *model.CarAuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.CarAuditRule, error) {
	var rule model.CarAuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarAuditRule{}, id).Error
}

func (r *auditRuleRepository) List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.CarAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarAuditRule
	var total int64

	q := r.db.Model(&model.CarAuditRule{})
	if query.RuleType != "" {
		q = q.Where("rule_type = ?", query.RuleType)
	}
	if query.RuleKey != "" {
		q = q.Where("rule_key = ?", query.RuleKey)
	}
	if query.Action != "" {
		q = q.Where("action = ?", query.Action)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("rule_name ILIKE ? OR description ILIKE ?", like, like)
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

func (r *auditRuleRepository) ListEnabled() ([]model.CarAuditRule, error) {
	var list []model.CarAuditRule
	if err := r.db.Where("status = ?", 1).Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) ListByRuleType(ruleType string) ([]model.CarAuditRule, error) {
	var list []model.CarAuditRule
	if err := r.db.Where("rule_type = ? AND status = ?", ruleType, 1).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// intPtr 工具函数：取 int 指针（用于将常量传入 *int 参数）
func intPtr(v int) *int { return &v }
