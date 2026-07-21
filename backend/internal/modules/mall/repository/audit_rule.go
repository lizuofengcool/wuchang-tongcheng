// Package repository 同城商城数据访问层 - 审核规则
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(rule *model.AuditRule) error
	FindByID(id uint) (*model.AuditRule, error)
	Update(rule *model.AuditRule) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(opts AuditRuleListOptions, pagination *utils.Pagination) ([]model.AuditRule, int64, error)
	ListEnabled() ([]model.AuditRule, error)
	ListByType(ruleType string) ([]model.AuditRule, error)
}

// AuditRuleListOptions 审核规则列表过滤条件
type AuditRuleListOptions struct {
	RuleType string
	RuleKey  string
	Action   string
	Status   *int
	Severity *int
	Keyword  string
}

type auditRuleRepository struct {
	db *gorm.DB
}

// NewAuditRuleRepository 创建审核规则仓储实例
func NewAuditRuleRepository(db *gorm.DB) AuditRuleRepository {
	return &auditRuleRepository{db: db}
}

func (r *auditRuleRepository) Create(rule *model.AuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.AuditRule, error) {
	var rule model.AuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(rule *model.AuditRule) error {
	return r.db.Save(rule).Error
}

func (r *auditRuleRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.AuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.AuditRule{}, id).Error
}

func (r *auditRuleRepository) List(opts AuditRuleListOptions, pagination *utils.Pagination) ([]model.AuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.AuditRule
	var total int64

	q := r.db.Model(&model.AuditRule{})
	if opts.RuleType != "" {
		q = q.Where("rule_type = ?", opts.RuleType)
	}
	if opts.RuleKey != "" {
		q = q.Where("rule_key = ?", opts.RuleKey)
	}
	if opts.Action != "" {
		q = q.Where("action = ?", opts.Action)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.Severity != nil {
		q = q.Where("severity = ?", *opts.Severity)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("rule_name ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auditRuleRepository) ListEnabled() ([]model.AuditRule, error) {
	var list []model.AuditRule
	if err := r.db.Where("status = ?", 1).Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) ListByType(ruleType string) ([]model.AuditRule, error) {
	var list []model.AuditRule
	if err := r.db.Where("rule_type = ? AND status = ?", ruleType, 1).
		Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
