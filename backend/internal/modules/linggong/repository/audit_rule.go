// Package repository 同城零工兼职数据访问层 - 审核规则
// LinggongAuditRule 审核规则（全局，BaseModel 无 region_id）
// 提供 M 端规则管理 + 内部规则查询能力
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(r *model.LinggongAuditRule) error
	FindByID(id uint) (*model.LinggongAuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（M 端管理）
	List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.LinggongAuditRule, int64, error)
	// 查询启用的规则（按 severity 降序、sort 升序）
	ListEnabled() ([]model.LinggongAuditRule, error)
	// 按规则类型反查启用规则
	ListByRuleType(ruleType string) ([]model.LinggongAuditRule, error)
	// 按 RuleKey 反查（用于精确匹配某条规则）
	FindByRuleKey(ruleKey string) (*model.LinggongAuditRule, error)

	// 启用/禁用
	ToggleStatus(id uint, status int) error
	// 批量删除
	BatchDelete(ids []uint) error
}

// AuditRuleListQuery 审核规则列表查询条件
type AuditRuleListQuery struct {
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

func (r *auditRuleRepository) Create(rule *model.LinggongAuditRule) error {
	return r.db.Create(rule).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.LinggongAuditRule, error) {
	var rule model.LinggongAuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongAuditRule{}, id).Error
}

func (r *auditRuleRepository) List(query AuditRuleListQuery, pagination *utils.Pagination) ([]model.LinggongAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.LinggongAuditRule
	var total int64

	q := r.db.Model(&model.LinggongAuditRule{})
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
	if query.Severity != nil {
		q = q.Where("severity = ?", *query.Severity)
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

func (r *auditRuleRepository) ListEnabled() ([]model.LinggongAuditRule, error) {
	var list []model.LinggongAuditRule
	if err := r.db.Where("status = ?", 1).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) ListByRuleType(ruleType string) ([]model.LinggongAuditRule, error) {
	var list []model.LinggongAuditRule
	if err := r.db.Where("rule_type = ? AND status = ?", ruleType, 1).
		Order("severity DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) FindByRuleKey(ruleKey string) (*model.LinggongAuditRule, error) {
	var rule model.LinggongAuditRule
	if err := r.db.Where("rule_key = ?", ruleKey).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *auditRuleRepository) ToggleStatus(id uint, status int) error {
	return r.db.Model(&model.LinggongAuditRule{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *auditRuleRepository) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.LinggongAuditRule{}).Error
}
