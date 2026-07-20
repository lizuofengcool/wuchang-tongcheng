// Package repository 同城114数据访问层 - 审核规则
// 敏感词/违禁内容/联系方式/价格校验/频率
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(rule *model.Dh114AuditRule) error
	FindByID(id uint) (*model.Dh114AuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.Dh114AuditRule, int64, error)
	ListEnabled() ([]model.Dh114AuditRule, error)
	ListByType(ruleType string) ([]model.Dh114AuditRule, error)
	IncrHitCount(id uint) error
}

// AuditRuleListQuery 审核规则列表查询
type AuditRuleListQuery struct {
	RuleType string
	RuleKey  string
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

func (r *auditRuleRepository) Create(rule *model.Dh114AuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.Dh114AuditRule, error) {
	var rule model.Dh114AuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114AuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114AuditRule{}, id).Error
}

func (r *auditRuleRepository) List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.Dh114AuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.Dh114AuditRule
	var total int64

	q := r.db.Model(&model.Dh114AuditRule{})
	if query.RuleType != "" {
		q = q.Where("rule_type = ?", query.RuleType)
	}
	if query.RuleKey != "" {
		q = q.Where("rule_key = ?", query.RuleKey)
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
		Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auditRuleRepository) ListEnabled() ([]model.Dh114AuditRule, error) {
	var list []model.Dh114AuditRule
	if err := r.db.Where("status = ?", 1).
		Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) ListByType(ruleType string) ([]model.Dh114AuditRule, error) {
	var list []model.Dh114AuditRule
	if err := r.db.Where("rule_type = ? AND status = ?", ruleType, 1).
		Order("sort ASC, severity DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) IncrHitCount(id uint) error {
	// Dh114AuditRule 没有 HitCount 字段，保留接口以备扩展
	return nil
}
