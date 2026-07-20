// Package repository 同城拼车出行数据访问层 - 审核规则
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AuditRuleListOptions 审核规则列表过滤条件
type AuditRuleListOptions struct {
	RuleType string
	RuleCode string
	Action   string
	Status   *int
	Keyword  string
}

// AuditRuleRepository 审核规则仓储接口
type AuditRuleRepository interface {
	Create(a *model.PincheAuditRule) error
	FindByID(id uint) (*model.PincheAuditRule, error)
	FindByCode(ruleCode string) (*model.PincheAuditRule, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts AuditRuleListOptions) ([]model.PincheAuditRule, int64, error)
	ListEnabled(regionID uint) ([]model.PincheAuditRule, error)
	ListByType(regionID uint, ruleType string) ([]model.PincheAuditRule, error)

	UpdateStatus(id uint, status int) error
	IncrHitCount(id uint) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type auditRuleRepository struct {
	db *gorm.DB
}

// NewAuditRuleRepository 创建审核规则仓储实例
func NewAuditRuleRepository(db *gorm.DB) AuditRuleRepository {
	return &auditRuleRepository{db: db}
}

func (r *auditRuleRepository) Create(a *model.PincheAuditRule) error {
	return r.db.Create(a).Error
}

func (r *auditRuleRepository) FindByID(id uint) (*model.PincheAuditRule, error) {
	var a model.PincheAuditRule
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *auditRuleRepository) FindByCode(ruleCode string) (*model.PincheAuditRule, error) {
	var a model.PincheAuditRule
	if err := r.db.Where("rule_code = ?", ruleCode).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *auditRuleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *auditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheAuditRule{}, id).Error
}

func (r *auditRuleRepository) List(regionID uint, pagination *utils.Pagination, opts AuditRuleListOptions) ([]model.PincheAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheAuditRule
	var total int64

	query := r.db.Model(&model.PincheAuditRule{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.RuleType != "" {
		query = query.Where("rule_type = ?", opts.RuleType)
	}
	if opts.RuleCode != "" {
		query = query.Where("rule_code = ?", opts.RuleCode)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		query = query.Where("rule_name ILIKE ? OR description ILIKE ?", "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("priority DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *auditRuleRepository) ListEnabled(regionID uint) ([]model.PincheAuditRule, error) {
	var list []model.PincheAuditRule
	q := r.db.Model(&model.PincheAuditRule{}).Where("status = 1")
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Order("priority DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) ListByType(regionID uint, ruleType string) ([]model.PincheAuditRule, error) {
	var list []model.PincheAuditRule
	q := r.db.Model(&model.PincheAuditRule{}).Where("rule_type = ? AND status = 1", ruleType)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Order("priority DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *auditRuleRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheAuditRule{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *auditRuleRepository) IncrHitCount(id uint) error {
	return r.db.Model(&model.PincheAuditRule{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *auditRuleRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheAuditRule{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
