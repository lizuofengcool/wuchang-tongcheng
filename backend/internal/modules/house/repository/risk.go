// Package repository 举报 + 评价 + 审核规则数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RiskRepository 风控仓储接口（举报 + 评价 + 审核规则）
type RiskRepository interface {
	// 举报
	CreateReport(r *model.HouseReport) error
	FindReportByID(id uint) (*model.HouseReport, error)
	FindReportByNo(no string) (*model.HouseReport, error)
	UpdateReport(r *model.HouseReport) error
	UpdateReportFields(id uint, fields map[string]interface{}) error
	ListReports(req *utils.Pagination, opts ReportListOptions) ([]model.HouseReport, int64, error)
	ListReportsByTarget(targetType string, targetID uint, req *utils.Pagination) ([]model.HouseReport, int64, error)
	ListReportsByReporter(reporterID uint, req *utils.Pagination) ([]model.HouseReport, int64, error)
	CountPendingReports() (int64, error)
	BatchUpdateReportStatus(ids []uint, status int) (int64, error)

	// 评价
	CreateReview(r *model.HouseReview) error
	FindReviewByID(id uint) (*model.HouseReview, error)
	UpdateReview(r *model.HouseReview) error
	UpdateReviewFields(id uint, fields map[string]interface{}) error
	DeleteReview(id uint) error
	ListReviews(regionID uint, req *utils.Pagination, opts ReviewListOptions) ([]model.HouseReview, int64, error)
	ListReviewsByTarget(targetType string, targetID uint, req *utils.Pagination) ([]model.HouseReview, int64, error)
	ListReviewsByReviewer(reviewerID uint, req *utils.Pagination) ([]model.HouseReview, int64, error)
	GetReviewStats(targetType string, targetID uint) (*model.ReviewStatsData, error)
	IncrReviewLikeCount(id uint) error
	DecrReviewLikeCount(id uint) error
	BatchUpdateReviewStatus(ids []uint, status int) (int64, error)

	// 审核规则
	CreateAuditRule(r *model.HouseAuditRule) error
	FindAuditRuleByID(id uint) (*model.HouseAuditRule, error)
	UpdateAuditRule(r *model.HouseAuditRule) error
	UpdateAuditRuleFields(id uint, fields map[string]interface{}) error
	DeleteAuditRule(id uint) error
	ListAuditRules(req *utils.Pagination, opts AuditRuleListOptions) ([]model.HouseAuditRule, int64, error)
	ListEnabledAuditRules() ([]model.HouseAuditRule, error)
}

// ReportListOptions 举报列表过滤条件
type ReportListOptions struct {
	TargetType  string
	TargetID    uint
	ReporterID  uint
	ReportType  string
	Status      *int
	PenaltyType string
	Keyword     string
}

// ReviewListOptions 评价列表过滤条件
type ReviewListOptions struct {
	TargetType    string
	TargetID      uint
	ReviewerID    uint
	ReviewType    string
	Rating        *int
	IsRecommended *bool
	Status        *int
	Sort          string // latest/rating/useful
}

// AuditRuleListOptions 审核规则列表过滤条件
type AuditRuleListOptions struct {
	RuleType string
	RuleKey  string
	Action   string
	Status   *int
	Keyword  string
}

// 提供 model.ReviewStatsData 别名（在 model 包中定义，下面会通过类型别名引用）
// 注：model 包中已有 ReviewStats 结构体的话使用 model.ReviewStatsData；否则此处定义

type riskRepository struct {
	db *gorm.DB
}

// NewRiskRepository 创建仓储实例
func NewRiskRepository(db *gorm.DB) RiskRepository {
	return &riskRepository{db: db}
}

// ===== 举报 =====

func (r *riskRepository) CreateReport(rp *model.HouseReport) error {
	return r.db.Create(rp).Error
}

func (r *riskRepository) FindReportByID(id uint) (*model.HouseReport, error) {
	var rp model.HouseReport
	if err := r.db.First(&rp, id).Error; err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *riskRepository) FindReportByNo(no string) (*model.HouseReport, error) {
	var rp model.HouseReport
	if err := r.db.Where("report_no = ?", no).First(&rp).Error; err != nil {
		return nil, err
	}
	return &rp, nil
}

func (r *riskRepository) UpdateReport(rp *model.HouseReport) error {
	return r.db.Save(rp).Error
}

func (r *riskRepository) UpdateReportFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseReport{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riskRepository) ListReports(req *utils.Pagination, opts ReportListOptions) ([]model.HouseReport, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReport
	var total int64

	query := r.db.Model(&model.HouseReport{})
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.ReporterID > 0 {
		query = query.Where("reporter_id = ?", opts.ReporterID)
	}
	if opts.ReportType != "" {
		query = query.Where("report_type = ?", opts.ReportType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.PenaltyType != "" {
		query = query.Where("penalty_type = ?", opts.PenaltyType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("report_no ILIKE ? OR reason ILIKE ? OR reporter_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListReportsByTarget(targetType string, targetID uint, req *utils.Pagination) ([]model.HouseReport, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReport
	var total int64

	query := r.db.Model(&model.HouseReport{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListReportsByReporter(reporterID uint, req *utils.Pagination) ([]model.HouseReport, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReport
	var total int64

	query := r.db.Model(&model.HouseReport{}).Where("reporter_id = ?", reporterID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) CountPendingReports() (int64, error) {
	var count int64
	if err := r.db.Model(&model.HouseReport{}).Where("status = ?", model.ReportStatusPending).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *riskRepository) BatchUpdateReportStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseReport{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 评价 =====

func (r *riskRepository) CreateReview(rv *model.HouseReview) error {
	return r.db.Create(rv).Error
}

func (r *riskRepository) FindReviewByID(id uint) (*model.HouseReview, error) {
	var rv model.HouseReview
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *riskRepository) UpdateReview(rv *model.HouseReview) error {
	return r.db.Save(rv).Error
}

func (r *riskRepository) UpdateReviewFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseReview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riskRepository) DeleteReview(id uint) error {
	return r.db.Delete(&model.HouseReview{}, id).Error
}

func (r *riskRepository) ListReviews(regionID uint, req *utils.Pagination, opts ReviewListOptions) ([]model.HouseReview, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReview
	var total int64

	query := r.db.Model(&model.HouseReview{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.TargetType != "" {
		query = query.Where("target_type = ?", opts.TargetType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.ReviewerID > 0 {
		query = query.Where("reviewer_id = ?", opts.ReviewerID)
	}
	if opts.ReviewType != "" {
		query = query.Where("review_type = ?", opts.ReviewType)
	}
	if opts.Rating != nil {
		query = query.Where("rating = ?", *opts.Rating)
	}
	if opts.IsRecommended != nil {
		query = query.Where("is_recommended = ?", *opts.IsRecommended)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	} else {
		query = query.Where("status = ?", model.ReviewStatusVisible)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch opts.Sort {
	case "rating":
		orderClause = "rating DESC, id DESC"
	case "useful":
		orderClause = "like_count DESC, id DESC"
	}
	if err := query.Scopes(utils.Paginate(req)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListReviewsByTarget(targetType string, targetID uint, req *utils.Pagination) ([]model.HouseReview, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReview
	var total int64

	query := r.db.Model(&model.HouseReview{}).
		Where("target_type = ? AND target_id = ? AND status = ?", targetType, targetID, model.ReviewStatusVisible)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListReviewsByReviewer(reviewerID uint, req *utils.Pagination) ([]model.HouseReview, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseReview
	var total int64

	query := r.db.Model(&model.HouseReview{}).Where("reviewer_id = ?", reviewerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetReviewStats 获取评价统计（总数、平均分、好评率、中评率、差评率）
func (r *riskRepository) GetReviewStats(targetType string, targetID uint) (*model.ReviewStatsData, error) {
	type row struct {
		Total int64   `gorm:"column:total"`
		Avg   float64 `gorm:"column:avg"`
		Good  int64   `gorm:"column:good"`
		Med   int64   `gorm:"column:med"`
		Bad   int64   `gorm:"column:bad"`
	}
	var rw row
	err := r.db.Model(&model.HouseReview{}).
		Select(`COUNT(*) as total,
			COALESCE(AVG(rating), 0) as avg,
			COUNT(CASE WHEN rating >= 4 THEN 1 END) as good,
			COUNT(CASE WHEN rating = 3 THEN 1 END) as med,
			COUNT(CASE WHEN rating <= 2 THEN 1 END) as bad`).
		Where("target_type = ? AND target_id = ? AND status = ?", targetType, targetID, model.ReviewStatusVisible).
		Scan(&rw).Error
	if err != nil {
		return nil, err
	}

	stats := &model.ReviewStatsData{
		TotalReviews: rw.Total,
		AvgRating:    rw.Avg,
	}
	if rw.Total > 0 {
		stats.GoodRate = float64(rw.Good) / float64(rw.Total)
		stats.MediumRate = float64(rw.Med) / float64(rw.Total)
		stats.BadRate = float64(rw.Bad) / float64(rw.Total)
	}
	return stats, nil
}

func (r *riskRepository) IncrReviewLikeCount(id uint) error {
	return r.db.Model(&model.HouseReview{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *riskRepository) DecrReviewLikeCount(id uint) error {
	return r.db.Model(&model.HouseReview{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

func (r *riskRepository) BatchUpdateReviewStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseReview{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 审核规则 =====

func (r *riskRepository) CreateAuditRule(ar *model.HouseAuditRule) error {
	return r.db.Create(ar).Error
}

func (r *riskRepository) FindAuditRuleByID(id uint) (*model.HouseAuditRule, error) {
	var ar model.HouseAuditRule
	if err := r.db.First(&ar, id).Error; err != nil {
		return nil, err
	}
	return &ar, nil
}

func (r *riskRepository) UpdateAuditRule(ar *model.HouseAuditRule) error {
	return r.db.Save(ar).Error
}

func (r *riskRepository) UpdateAuditRuleFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riskRepository) DeleteAuditRule(id uint) error {
	return r.db.Delete(&model.HouseAuditRule{}, id).Error
}

func (r *riskRepository) ListAuditRules(req *utils.Pagination, opts AuditRuleListOptions) ([]model.HouseAuditRule, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseAuditRule
	var total int64

	query := r.db.Model(&model.HouseAuditRule{})
	if opts.RuleType != "" {
		query = query.Where("rule_type = ?", opts.RuleType)
	}
	if opts.RuleKey != "" {
		query = query.Where("rule_key = ?", opts.RuleKey)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("rule_name ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListEnabledAuditRules() ([]model.HouseAuditRule, error) {
	var list []model.HouseAuditRule
	if err := r.db.Where("status = ?", 1).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
