// Package repository 举报/评价/审核规则/用户信用数据访问层
// 依据 v3.2.1 架构方案：对标转转
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Report =====

// ReportRepository 举报仓储接口
type ReportRepository interface {
	Create(r *model.ErshouReport) error
	FindByID(id uint) (*model.ErshouReport, error)
	FindByReportNo(reportNo string) (*model.ErshouReport, error)
	Update(id uint, fields map[string]interface{}) error
	List(query ReportListQuery, pagination *utils.Pagination) ([]model.ErshouReport, int64, error)
	ListByErshouID(ershouID uint) ([]model.ErshouReport, error)
	CountByReporter(reporterID uint) (int64, error)
}

// ReportListQuery 举报列表查询
type ReportListQuery struct {
	Status     *int
	ReportType string
	ErshouID   uint
	ReporterID uint
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository 创建举报仓储实例
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(rep *model.ErshouReport) error {
	return r.db.Create(rep).Error
}

func (r *reportRepository) FindByID(id uint) (*model.ErshouReport, error) {
	var rep model.ErshouReport
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) FindByReportNo(reportNo string) (*model.ErshouReport, error) {
	var rep model.ErshouReport
	if err := r.db.Where("report_no = ?", reportNo).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *reportRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouReport{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reportRepository) List(query ReportListQuery, pagination *utils.Pagination) ([]model.ErshouReport, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouReport
	var total int64

	q := r.db.Model(&model.ErshouReport{})
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.ReportType != "" {
		q = q.Where("report_type = ?", query.ReportType)
	}
	if query.ErshouID > 0 {
		q = q.Where("ershou_id = ?", query.ErshouID)
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

func (r *reportRepository) ListByErshouID(ershouID uint) ([]model.ErshouReport, error) {
	var list []model.ErshouReport
	if err := r.db.Where("ershou_id = ?", ershouID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportRepository) CountByReporter(reporterID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ErshouReport{}).Where("reporter_id = ?", reporterID).Count(&count).Error
	return count, err
}

// ===== Review =====

// ReviewRepository 评价仓储接口
type ReviewRepository interface {
	Create(r *model.ErshouReview) error
	FindByID(id uint) (*model.ErshouReview, error)
	FindByOrderAndReviewer(orderID, reviewerID uint) (*model.ErshouReview, error)
	Update(id uint, fields map[string]interface{}) error
	List(query ReviewListQuery, pagination *utils.Pagination) ([]model.ErshouReview, int64, error)
	ListByErshouID(ershouID uint, pagination *utils.Pagination) ([]model.ErshouReview, int64, error)
	StatsByErshouID(ershouID uint) (total int64, avgRating float64, good, medium, bad int64, err error)
	StatsByUserID(userID uint) (total int64, avgRating float64, err error)
}

// ReviewListQuery 评价列表查询
type ReviewListQuery struct {
	ErshouID   uint
	ReviewerID uint
	RevieweeID uint
	Rating     *int
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建评价仓储实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(rv *model.ErshouReview) error {
	return r.db.Create(rv).Error
}

func (r *reviewRepository) FindByID(id uint) (*model.ErshouReview, error) {
	var rv model.ErshouReview
	if err := r.db.First(&rv, id).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) FindByOrderAndReviewer(orderID, reviewerID uint) (*model.ErshouReview, error) {
	var rv model.ErshouReview
	if err := r.db.Where("order_id = ? AND reviewer_id = ?", orderID, reviewerID).First(&rv).Error; err != nil {
		return nil, err
	}
	return &rv, nil
}

func (r *reviewRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouReview{}).Where("id = ?", id).Updates(fields).Error
}

func (r *reviewRepository) List(query ReviewListQuery, pagination *utils.Pagination) ([]model.ErshouReview, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouReview
	var total int64

	q := r.db.Model(&model.ErshouReview{})
	if query.ErshouID > 0 {
		q = q.Where("ershou_id = ?", query.ErshouID)
	}
	if query.ReviewerID > 0 {
		q = q.Where("reviewer_id = ?", query.ReviewerID)
	}
	if query.RevieweeID > 0 {
		q = q.Where("reviewee_id = ?", query.RevieweeID)
	}
	if query.Rating != nil {
		q = q.Where("rating = ?", *query.Rating)
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

func (r *reviewRepository) ListByErshouID(ershouID uint, pagination *utils.Pagination) ([]model.ErshouReview, int64, error) {
	return r.List(ReviewListQuery{ErshouID: ershouID}, pagination)
}

func (r *reviewRepository) StatsByErshouID(ershouID uint) (int64, float64, int64, int64, int64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
		Good    int64   `gorm:"column:good"`
		Medium  int64   `gorm:"column:medium"`
		Bad     int64   `gorm:"column:bad"`
	}
	var s stat
	err := r.db.Model(&model.ErshouReview{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate, "+
			"COUNT(CASE WHEN rating >= 4 THEN 1 END) AS good, "+
			"COUNT(CASE WHEN rating = 3 THEN 1 END) AS medium, "+
			"COUNT(CASE WHEN rating <= 2 THEN 1 END) AS bad").
		Where("ershou_id = ?", ershouID).
		Scan(&s).Error
	return s.Total, s.AvgRate, s.Good, s.Medium, s.Bad, err
}

func (r *reviewRepository) StatsByUserID(userID uint) (int64, float64, error) {
	type stat struct {
		Total   int64   `gorm:"column:total"`
		AvgRate float64 `gorm:"column:avg_rate"`
	}
	var s stat
	err := r.db.Model(&model.ErshouReview{}).
		Select("COUNT(*) AS total, COALESCE(AVG(rating),0) AS avg_rate").
		Where("reviewee_id = ?", userID).
		Scan(&s).Error
	return s.Total, s.AvgRate, err
}

// ===== AuditRule =====

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(r *model.ErshouAuditRule) error
	FindByID(id uint) (*model.ErshouAuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.ErshouAuditRule, int64, error)
	ListEnabled() ([]model.ErshouAuditRule, error)
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

func (r *auditRuleRepository) Create(rule *model.ErshouAuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.ErshouAuditRule, error) {
	var rule model.ErshouAuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouAuditRule{}, id).Error
}

func (r *auditRuleRepository) List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.ErshouAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.ErshouAuditRule
	var total int64

	q := r.db.Model(&model.ErshouAuditRule{})
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
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auditRuleRepository) ListEnabled() ([]model.ErshouAuditRule, error) {
	var list []model.ErshouAuditRule
	if err := r.db.Where("status = ?", model.AuditRuleStatusEnabled).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== UserCredit =====

// UserCreditRepository 用户信用仓储接口
type UserCreditRepository interface {
	Create(c *model.ErshouUserCredit) error
	FindByUserID(userID uint) (*model.ErshouUserCredit, error)
	Update(userID uint, fields map[string]interface{}) error
	IncrTransactions(userID uint, success, cancel int, scoreDelta int) error
	IncrReviews(userID uint, good, medium, bad int) error
	IncrDisputes(userID uint, n int) error
	IncrReports(userID uint, n int) error
	IncrPenalties(userID uint, n int, scoreDelta int) error
}

type userCreditRepository struct {
	db *gorm.DB
}

// NewUserCreditRepository 创建用户信用仓储实例
func NewUserCreditRepository(db *gorm.DB) UserCreditRepository {
	return &userCreditRepository{db: db}
}

func (r *userCreditRepository) Create(c *model.ErshouUserCredit) error {
	return r.db.Create(c).Error
}

func (r *userCreditRepository) FindByUserID(userID uint) (*model.ErshouUserCredit, error) {
	var c model.ErshouUserCredit
	if err := r.db.Where("user_id = ?", userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *userCreditRepository) Update(userID uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).Updates(fields).Error
}

func (r *userCreditRepository) IncrTransactions(userID uint, success, cancel int, scoreDelta int) error {
	updates := map[string]interface{}{
		"total_transactions":   gorm.Expr("total_transactions + ?", success+cancel),
		"success_transactions": gorm.Expr("success_transactions + ?", success),
		"cancel_transactions":  gorm.Expr("cancel_transactions + ?", cancel),
		"credit_score":         gorm.Expr("credit_score + ?", scoreDelta),
		"last_transaction_at":  gorm.Expr("NOW()"),
	}
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *userCreditRepository) IncrReviews(userID uint, good, medium, bad int) error {
	updates := map[string]interface{}{
		"good_reviews":   gorm.Expr("good_reviews + ?", good),
		"medium_reviews": gorm.Expr("medium_reviews + ?", medium),
		"bad_reviews":    gorm.Expr("bad_reviews + ?", bad),
	}
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).Updates(updates).Error
}

func (r *userCreditRepository) IncrDisputes(userID uint, n int) error {
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"disputes":     gorm.Expr("disputes + ?", n),
			"credit_score": gorm.Expr("credit_score - ?", n*5),
		}).Error
}

func (r *userCreditRepository) IncrReports(userID uint, n int) error {
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).
		UpdateColumn("reports", gorm.Expr("reports + ?", n)).Error
}

func (r *userCreditRepository) IncrPenalties(userID uint, n int, scoreDelta int) error {
	return r.db.Model(&model.ErshouUserCredit{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"penalties":    gorm.Expr("penalties + ?", n),
			"credit_score": gorm.Expr("credit_score + ?", scoreDelta),
		}).Error
}
