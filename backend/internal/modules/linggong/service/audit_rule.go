// Package service 同城零工兼职业务逻辑层 - 审核规则
// 全局规则管理 + 内部规则查询能力（M 端管理）
// 注意：BaseModel 无 region_id
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound = errors.New("审核规则不存在")
)

// AuditRuleService 审核规则业务接口
type AuditRuleService interface {
	// M 端管理
	Create(req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error)
	Update(id uint, req *dto.UpdateAuditRuleRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.AuditRuleInfo, error)
	List(req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error)
	UpdateStatus(id uint, status int) error
	BatchDelete(ids []uint) error

	// 内部查询
	ListEnabled() ([]dto.AuditRuleInfo, error)
	ListByRuleType(ruleType string) ([]dto.AuditRuleInfo, error)
	FindByRuleKey(ruleKey string) (*dto.AuditRuleInfo, error)
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

// auditRuleTypeText 规则类型文本
func auditRuleTypeText(t string) string {
	switch t {
	case model.LinggongAuditRuleTypeSensitiveWord:
		return "敏感词"
	case model.LinggongAuditRuleTypeSalaryCheck:
		return "薪资异常"
	case model.LinggongAuditRuleTypeFrequency:
		return "频率限制"
	case model.LinggongAuditRuleTypeFakeJob:
		return "虚假岗位"
	case model.LinggongAuditRuleTypeContact:
		return "联系方式校验"
	case model.LinggongAuditRuleTypeProhibited:
		return "违禁内容"
	case model.LinggongAuditRuleTypeDuplicate:
		return "重复发布"
	case model.LinggongAuditRuleTypeBlacklist:
		return "黑名单"
	}
	return ""
}

// auditRuleActionText 审核动作文本
func auditRuleActionText(a string) string {
	switch a {
	case model.LinggongAuditActionReject:
		return "拒绝"
	case model.LinggongAuditActionApproval:
		return "转人工"
	case model.LinggongAuditActionFilter:
		return "过滤敏感词"
	case model.LinggongAuditActionLimit:
		return "限流"
	}
	return ""
}

// auditRuleStatusText 审核规则状态文本
func auditRuleStatusText(s int) string {
	switch s {
	case 0:
		return "禁用"
	case 1:
		return "启用"
	}
	return ""
}

// toAuditRuleInfo model -> dto
func toAuditRuleInfo(r *model.LinggongAuditRule) *dto.AuditRuleInfo {
	var threshold interface{}
	if r.Threshold != nil {
		_ = r.Threshold.Parse(&threshold)
	}
	return &dto.AuditRuleInfo{
		ID:           r.ID,
		RuleName:     r.RuleName,
		RuleType:     r.RuleType,
		RuleTypeText: auditRuleTypeText(r.RuleType),
		RuleKey:      r.RuleKey,
		Pattern:      r.Pattern,
		Threshold:    threshold,
		Action:       r.Action,
		ActionText:   auditRuleActionText(r.Action),
		PenaltyType:  r.PenaltyType,
		Severity:     r.Severity,
		Status:       r.Status,
		StatusText:   auditRuleStatusText(r.Status),
		Description:  r.Description,
		Sort:         r.Sort,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// ===== M 端管理 =====

// Create 创建审核规则
func (s *auditRuleService) Create(req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error) {
	r := &model.LinggongAuditRule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		RuleKey:     req.RuleKey,
		Pattern:     req.Pattern,
		Action:      req.Action,
		PenaltyType: req.PenaltyType,
		Severity:    req.Severity,
		Status:      req.Status,
		Description: req.Description,
		Sort:        req.Sort,
	}

	// 默认值兜底
	if r.RuleType == "" {
		r.RuleType = model.LinggongAuditRuleTypeSensitiveWord
	}
	if r.Action == "" {
		r.Action = model.LinggongAuditActionReject
	}
	if r.Severity == 0 {
		r.Severity = 1
	}
	if r.Status == 0 {
		r.Status = 1
	}

	// JSONB 字段
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			r.Threshold = jb
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}

// Update 更新审核规则
func (s *auditRuleService) Update(id uint, req *dto.UpdateAuditRuleRequest) error {
	fields := map[string]interface{}{}
	if req.RuleName != nil {
		fields["rule_name"] = *req.RuleName
	}
	if req.RuleType != nil {
		fields["rule_type"] = *req.RuleType
	}
	if req.RuleKey != nil {
		fields["rule_key"] = *req.RuleKey
	}
	if req.Pattern != nil {
		fields["pattern"] = *req.Pattern
	}
	if req.Action != nil {
		fields["action"] = *req.Action
	}
	if req.PenaltyType != nil {
		fields["penalty_type"] = *req.PenaltyType
	}
	if req.Severity != nil {
		fields["severity"] = *req.Severity
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			fields["threshold"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除审核规则
func (s *auditRuleService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// GetByID 获取审核规则详情
func (s *auditRuleService) GetByID(id uint) (*dto.AuditRuleInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}

// List 审核规则列表
func (s *auditRuleService) List(req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.AuditRuleListQuery{
		RuleType: req.RuleType,
		Action:   req.Action,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 启用/禁用
func (s *auditRuleService) UpdateStatus(id uint, status int) error {
	return s.repo.ToggleStatus(id, status)
}

// BatchDelete 批量删除
func (s *auditRuleService) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return s.repo.BatchDelete(ids)
}

// ===== 内部查询 =====

// ListEnabled 查询启用的规则
func (s *auditRuleService) ListEnabled() ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return result, nil
}

// ListByRuleType 按规则类型反查
func (s *auditRuleService) ListByRuleType(ruleType string) ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListByRuleType(ruleType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return result, nil
}

// FindByRuleKey 按 RuleKey 反查
func (s *auditRuleService) FindByRuleKey(ruleKey string) (*dto.AuditRuleInfo, error) {
	r, err := s.repo.FindByRuleKey(ruleKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}
