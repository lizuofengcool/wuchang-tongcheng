// Package repository 风控中台扩展数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/risk/model"

	"gorm.io/gorm"
)

// RiskExtendRepository 风控扩展仓储接口
type RiskExtendRepository interface {
	// 举报证据
	CreateEvidence(e *model.ReportEvidence) error
	ListEvidenceByReport(reportID uint) ([]model.ReportEvidence, error)
	DeleteEvidence(id uint) error

	// 申诉
	CreateAppeal(a *model.Appeal) error
	FindAppealByID(id uint) (*model.Appeal, error)
	FindAppealByNo(appealNo string) (*model.Appeal, error)
	ListAppeals(status int, page, pageSize int) ([]model.Appeal, int64, error)
	ListUserAppeals(userID uint, page, pageSize int) ([]model.Appeal, int64, error)
	UpdateAppealFields(id uint, fields map[string]interface{}) error

	// 风控规则
	CreateRule(r *model.Rule) error
	FindRuleByID(id uint) (*model.Rule, error)
	FindRuleByName(name string) (*model.Rule, error)
	ListRules(ruleType string, page, pageSize int) ([]model.Rule, int64, error)
	ListActiveRules() ([]model.Rule, error)
	UpdateRuleFields(id uint, fields map[string]interface{}) error
	DeleteRule(id uint) error
	IncrRuleHitCount(id uint) error

	// 风险评分记录
	CreateScoreRecord(r *model.ScoreRecord) error
	ListScoreRecordsByUser(userID uint, page, pageSize int) ([]model.ScoreRecord, int64, error)
	ListScoreRecordsByLevel(level string, page, pageSize int) ([]model.ScoreRecord, int64, error)

	// 审核日志
	CreateAuditLog(l *model.AuditLog) error
	ListAuditLogs(auditorID uint, action, targetType string, page, pageSize int) ([]model.AuditLog, int64, error)

	// 统计
	StatTotalReports() (int64, error)
	StatPendingReports() (int64, error)
	StatHandledReports() (int64, error)
	StatTotalAppeals() (int64, error)
	StatPendingAppeals() (int64, error)
	StatTotalViolations() (int64, error)
	StatActiveViolations() (int64, error)
	StatBlacklistCount() (int64, error)
	StatSensitiveWords() (int64, error)
	StatRulesCount() (int64, error)
	StatAuditLogsCount() (int64, error)
}

type riskExtendRepository struct {
	db *gorm.DB
}

// NewRiskExtendRepository 创建扩展仓储实例
func NewRiskExtendRepository(db *gorm.DB) RiskExtendRepository {
	return &riskExtendRepository{db: db}
}

// ===== 举报证据 =====

func (r *riskExtendRepository) CreateEvidence(e *model.ReportEvidence) error {
	return r.db.Create(e).Error
}

func (r *riskExtendRepository) ListEvidenceByReport(reportID uint) ([]model.ReportEvidence, error) {
	var list []model.ReportEvidence
	if err := r.db.Where("report_id = ?", reportID).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *riskExtendRepository) DeleteEvidence(id uint) error {
	return r.db.Delete(&model.ReportEvidence{}, id).Error
}

// ===== 申诉 =====

func (r *riskExtendRepository) CreateAppeal(a *model.Appeal) error {
	return r.db.Create(a).Error
}

func (r *riskExtendRepository) FindAppealByID(id uint) (*model.Appeal, error) {
	var a model.Appeal
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *riskExtendRepository) FindAppealByNo(appealNo string) (*model.Appeal, error) {
	var a model.Appeal
	if err := r.db.Where("appeal_no = ?", appealNo).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *riskExtendRepository) ListAppeals(status int, page, pageSize int) ([]model.Appeal, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Appeal
	var total int64
	q := r.db.Model(&model.Appeal{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskExtendRepository) ListUserAppeals(userID uint, page, pageSize int) ([]model.Appeal, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Appeal
	var total int64
	q := r.db.Model(&model.Appeal{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskExtendRepository) UpdateAppealFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Appeal{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 风控规则 =====

func (r *riskExtendRepository) CreateRule(rule *model.Rule) error {
	return r.db.Create(rule).Error
}

func (r *riskExtendRepository) FindRuleByID(id uint) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *riskExtendRepository) FindRuleByName(name string) (*model.Rule, error) {
	var rule model.Rule
	if err := r.db.Where("rule_name = ?", name).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *riskExtendRepository) ListRules(ruleType string, page, pageSize int) ([]model.Rule, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Rule
	var total int64
	q := r.db.Model(&model.Rule{})
	if ruleType != "" {
		q = q.Where("rule_type = ?", ruleType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("priority ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskExtendRepository) ListActiveRules() ([]model.Rule, error) {
	var list []model.Rule
	if err := r.db.Where("status = ?", 1).Order("priority ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *riskExtendRepository) UpdateRuleFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Rule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *riskExtendRepository) DeleteRule(id uint) error {
	return r.db.Delete(&model.Rule{}, id).Error
}

func (r *riskExtendRepository) IncrRuleHitCount(id uint) error {
	return r.db.Model(&model.Rule{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": gorm.Expr("NOW()"),
		}).Error
}

// ===== 风险评分记录 =====

func (r *riskExtendRepository) CreateScoreRecord(rec *model.ScoreRecord) error {
	return r.db.Create(rec).Error
}

func (r *riskExtendRepository) ListScoreRecordsByUser(userID uint, page, pageSize int) ([]model.ScoreRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.ScoreRecord
	var total int64
	q := r.db.Model(&model.ScoreRecord{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *riskExtendRepository) ListScoreRecordsByLevel(level string, page, pageSize int) ([]model.ScoreRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.ScoreRecord
	var total int64
	q := r.db.Model(&model.ScoreRecord{}).Where("level = ?", level)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 审核日志 =====

func (r *riskExtendRepository) CreateAuditLog(l *model.AuditLog) error {
	return r.db.Create(l).Error
}

func (r *riskExtendRepository) ListAuditLogs(auditorID uint, action, targetType string, page, pageSize int) ([]model.AuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.AuditLog
	var total int64
	q := r.db.Model(&model.AuditLog{})
	if auditorID > 0 {
		q = q.Where("auditor_id = ?", auditorID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 统计 =====

func (r *riskExtendRepository) StatTotalReports() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatPendingReports() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status = ?", model.ReportStatusPending).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatHandledReports() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status >= ?", model.ReportStatusHandled).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatTotalAppeals() (int64, error) {
	var count int64
	err := r.db.Model(&model.Appeal{}).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatPendingAppeals() (int64, error) {
	var count int64
	err := r.db.Model(&model.Appeal{}).Where("status = ?", model.AppealStatusPending).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatTotalViolations() (int64, error) {
	var count int64
	err := r.db.Model(&model.Violation{}).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatActiveViolations() (int64, error) {
	var count int64
	err := r.db.Model(&model.Violation{}).Where("status = ?", model.ViolationStatusActive).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatBlacklistCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.Blacklist{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatSensitiveWords() (int64, error) {
	var count int64
	err := r.db.Model(&model.SensitiveWord{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatRulesCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.Rule{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

func (r *riskExtendRepository) StatAuditLogsCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.AuditLog{}).Count(&count).Error
	return count, err
}
