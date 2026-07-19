// Package repository 风控审核中台精简版数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/risk/model"

	"gorm.io/gorm"
)

// RiskRepository 风控中台仓储接口
type RiskRepository interface {
	// 举报
	CreateReport(r *model.Report) error
	FindReportByID(id uint) (*model.Report, error)
	FindReportByNo(reportNo string) (*model.Report, error)
	ListReports(req *ListReportsQuery) ([]model.Report, int64, error)
	UpdateReportFields(id uint, fields map[string]interface{}) error

	// 敏感词
	CreateSensitiveWord(w *model.SensitiveWord) error
	FindSensitiveWordByID(id uint) (*model.SensitiveWord, error)
	FindSensitiveWordByWord(word string) (*model.SensitiveWord, error)
	ListSensitiveWords(wordType string, page, pageSize int) ([]model.SensitiveWord, int64, error)
	ListAllActiveSensitiveWords() ([]model.SensitiveWord, error)
	UpdateSensitiveWordFields(id uint, fields map[string]interface{}) error
	DeleteSensitiveWord(id uint) error
	BatchCreateSensitiveWords(words []model.SensitiveWord) error

	// 审核规则
	CreateAuditRule(r *model.AuditRule) error
	FindAuditRuleByName(name string) (*model.AuditRule, error)
	ListAuditRules(ruleType string, page, pageSize int) ([]model.AuditRule, int64, error)
	ListActiveAuditRulesByType(ruleType string) ([]model.AuditRule, error)
	UpdateAuditRuleFields(id uint, fields map[string]interface{}) error

	// 黑名单
	CreateBlacklist(b *model.Blacklist) error
	FindActiveBlacklist(targetType, targetValue string) (*model.Blacklist, error)
	ListBlacklist(targetType string, page, pageSize int) ([]model.Blacklist, int64, error)
	UpdateBlacklistFields(id uint, fields map[string]interface{}) error

	// 用户风险分
	GetOrCreateUserScore(userID, regionID uint) (*model.UserScore, error)
	FindUserScore(userID uint) (*model.UserScore, error)
	UpdateUserScoreFields(id uint, fields map[string]interface{}) error

	// 违规
	CreateViolation(v *model.Violation) error
	FindViolationByID(id uint) (*model.Violation, error)
	ListUserViolations(userID uint, page, pageSize int) ([]model.Violation, int64, error)
	ListActiveViolationsByUser(userID uint) ([]model.Violation, error)
	UpdateViolationFields(id uint, fields map[string]interface{}) error
}

// ListReportsQuery 举报列表查询参数
type ListReportsQuery struct {
	RegionID   uint
	Status     int
	ReportType string
	BizModule  string
	ReporterID uint
	Page       int
	PageSize   int
}

type riskRepository struct {
	db *gorm.DB
}

// NewRiskRepository 创建仓储实例
func NewRiskRepository(db *gorm.DB) RiskRepository {
	return &riskRepository{db: db}
}

// ===== 举报 =====

func (r *riskRepository) CreateReport(rep *model.Report) error {
	return r.db.Create(rep).Error
}

func (r *riskRepository) FindReportByID(id uint) (*model.Report, error) {
	var rep model.Report
	if err := r.db.First(&rep, id).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *riskRepository) FindReportByNo(reportNo string) (*model.Report, error) {
	var rep model.Report
	if err := r.db.Where("report_no = ?", reportNo).First(&rep).Error; err != nil {
		return nil, err
	}
	return &rep, nil
}

func (r *riskRepository) ListReports(req *ListReportsQuery) ([]model.Report, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	var list []model.Report
	var total int64
	q := r.db.Model(&model.Report{})
	if req.RegionID > 0 {
		q = q.Where("region_id = ?", req.RegionID)
	}
	if req.Status >= 0 {
		q = q.Where("status = ?", req.Status)
	}
	if req.ReportType != "" {
		q = q.Where("report_type = ?", req.ReportType)
	}
	if req.BizModule != "" {
		q = q.Where("reported_biz_module = ?", req.BizModule)
	}
	if req.ReporterID > 0 {
		q = q.Where("reporter_id = ?", req.ReporterID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) UpdateReportFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Report{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 敏感词 =====

func (r *riskRepository) CreateSensitiveWord(w *model.SensitiveWord) error {
	return r.db.Create(w).Error
}

func (r *riskRepository) FindSensitiveWordByID(id uint) (*model.SensitiveWord, error) {
	var w model.SensitiveWord
	if err := r.db.First(&w, id).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *riskRepository) FindSensitiveWordByWord(word string) (*model.SensitiveWord, error) {
	var w model.SensitiveWord
	if err := r.db.Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *riskRepository) ListSensitiveWords(wordType string, page, pageSize int) ([]model.SensitiveWord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	var list []model.SensitiveWord
	var total int64
	q := r.db.Model(&model.SensitiveWord{})
	if wordType != "" {
		q = q.Where("word_type = ?", wordType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListAllActiveSensitiveWords() ([]model.SensitiveWord, error) {
	var list []model.SensitiveWord
	if err := r.db.Where("status = ?", 1).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *riskRepository) UpdateSensitiveWordFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.SensitiveWord{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riskRepository) DeleteSensitiveWord(id uint) error {
	return r.db.Delete(&model.SensitiveWord{}, id).Error
}

func (r *riskRepository) BatchCreateSensitiveWords(words []model.SensitiveWord) error {
	if len(words) == 0 {
		return nil
	}
	return r.db.Create(&words).Error
}

// ===== 审核规则 =====

func (r *riskRepository) CreateAuditRule(rule *model.AuditRule) error {
	return r.db.Create(rule).Error
}

func (r *riskRepository) FindAuditRuleByName(name string) (*model.AuditRule, error) {
	var rule model.AuditRule
	if err := r.db.Where("rule_name = ?", name).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *riskRepository) ListAuditRules(ruleType string, page, pageSize int) ([]model.AuditRule, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.AuditRule
	var total int64
	q := r.db.Model(&model.AuditRule{})
	if ruleType != "" {
		q = q.Where("rule_type = ?", ruleType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListActiveAuditRulesByType(ruleType string) ([]model.AuditRule, error) {
	var list []model.AuditRule
	q := r.db.Where("status = ?", 1)
	if ruleType != "" {
		q = q.Where("rule_type = ?", ruleType)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *riskRepository) UpdateAuditRuleFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.AuditRule{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 黑名单 =====

func (r *riskRepository) CreateBlacklist(b *model.Blacklist) error {
	return r.db.Create(b).Error
}

func (r *riskRepository) FindActiveBlacklist(targetType, targetValue string) (*model.Blacklist, error) {
	var b model.Blacklist
	err := r.db.Where("target_type = ? AND target_value = ? AND status = ?", targetType, targetValue, 1).
		Where("expire_at IS NULL OR expire_at > NOW()").
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *riskRepository) ListBlacklist(targetType string, page, pageSize int) ([]model.Blacklist, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Blacklist
	var total int64
	q := r.db.Model(&model.Blacklist{})
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) UpdateBlacklistFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Blacklist{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 用户风险分 =====

func (r *riskRepository) GetOrCreateUserScore(userID, regionID uint) (*model.UserScore, error) {
	var s model.UserScore
	err := r.db.Where("user_id = ?", userID).First(&s).Error
	if err == nil {
		return &s, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	s = model.UserScore{
		RegionID: regionID,
		UserID:   userID,
		Score:    100,
		Level:    model.RiskLevelSafe,
	}
	if err := r.db.Create(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *riskRepository) FindUserScore(userID uint) (*model.UserScore, error) {
	var s model.UserScore
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *riskRepository) UpdateUserScoreFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.UserScore{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 违规处罚 =====

func (r *riskRepository) CreateViolation(v *model.Violation) error {
	return r.db.Create(v).Error
}

func (r *riskRepository) FindViolationByID(id uint) (*model.Violation, error) {
	var v model.Violation
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *riskRepository) ListUserViolations(userID uint, page, pageSize int) ([]model.Violation, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Violation
	var total int64
	q := r.db.Model(&model.Violation{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskRepository) ListActiveViolationsByUser(userID uint) ([]model.Violation, error) {
	var list []model.Violation
	if err := r.db.Where("user_id = ? AND status = ?", userID, model.ViolationStatusActive).
		Where("penalty_end IS NULL OR penalty_end > NOW()").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *riskRepository) UpdateViolationFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Violation{}).Where("id = ?", id).Updates(fields).Error
}
