// Package repository love 相亲交友数据访问层 - 审核规则
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveAuditRuleRepository 审核规则仓储接口
type LoveAuditRuleRepository interface {
	Create(r *model.LoveAuditRule) error
	FindByID(id uint) (*model.LoveAuditRule, error)
	FindByRuleKey(key string) (*model.LoveAuditRule, error)
	Update(r *model.LoveAuditRule) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts LoveAuditRuleListOptions) ([]model.LoveAuditRule, int64, error)
	ListByRuleType(ruleType string) ([]model.LoveAuditRule, error)
	ListByTargetType(targetType string) ([]model.LoveAuditRule, error)
	ListAll() ([]model.LoveAuditRule, error)
	BatchUpdateStatus(ids []uint, status int) error
}

// LoveAuditRuleListOptions 审核规则列表过滤
type LoveAuditRuleListOptions struct {
	RuleType string
	RuleKey  string
	Action   string
	Status   *int
	Keyword  string
}

type loveAuditRuleRepository struct {
	db *gorm.DB
}

// NewLoveAuditRuleRepository 创建审核规则仓储
func NewLoveAuditRuleRepository(db *gorm.DB) LoveAuditRuleRepository {
	return &loveAuditRuleRepository{db: db}
}

func (r *loveAuditRuleRepository) Create(rule *model.LoveAuditRule) error {
	return r.db.Create(rule).Error
}

func (r *loveAuditRuleRepository) FindByID(id uint) (*model.LoveAuditRule, error) {
	var rule model.LoveAuditRule
	if err := r.db.First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *loveAuditRuleRepository) FindByRuleKey(key string) (*model.LoveAuditRule, error) {
	var rule model.LoveAuditRule
	if err := r.db.Where("rule_key = ?", key).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *loveAuditRuleRepository) Update(rule *model.LoveAuditRule) error {
	return r.db.Save(rule).Error
}

func (r *loveAuditRuleRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LoveAuditRule{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveAuditRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.LoveAuditRule{}, id).Error
}

func (r *loveAuditRuleRepository) List(pagination *utils.Pagination, opts LoveAuditRuleListOptions) ([]model.LoveAuditRule, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LoveAuditRule
	var total int64

	query := r.db.Model(&model.LoveAuditRule{})
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
	if err := query.Scopes(utils.Paginate(pagination)).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveAuditRuleRepository) ListByRuleType(ruleType string) ([]model.LoveAuditRule, error) {
	var list []model.LoveAuditRule
	err := r.db.Where("rule_type = ? AND status = ?", ruleType, 1).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *loveAuditRuleRepository) ListByTargetType(targetType string) ([]model.LoveAuditRule, error) {
	var list []model.LoveAuditRule
	err := r.db.Where("(target_type = ? OR target_type = ?) AND status = ?", targetType, "all", 1).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *loveAuditRuleRepository) ListAll() ([]model.LoveAuditRule, error) {
	var list []model.LoveAuditRule
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *loveAuditRuleRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.LoveAuditRule{}).Where("id IN ?", ids).Update("status", status).Error
}
