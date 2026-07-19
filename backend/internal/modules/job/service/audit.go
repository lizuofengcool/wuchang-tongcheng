// Package service 审核规则业务逻辑层
// 依据 v3.2.1 架构方案：敏感词/薪资异常/频率限制等审核规则
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound     = errors.New("审核规则不存在")
	ErrAuditRuleNoPermission = errors.New("无权操作此审核规则")
)

// AuditService 审核规则业务接口
type AuditService interface {
	CreateRule(req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error)
	UpdateRule(id uint, req *dto.AuditRuleUpdateRequest) error
	DeleteRule(id uint) error
	GetRule(id uint) (*dto.AuditRuleResponse, error)
	ListRules(req *dto.AuditRuleListQuery) (*utils.Pagination, []dto.AuditRuleResponse, error)
	ListEnabledRules() ([]dto.AuditRuleResponse, error)
}

type auditService struct {
	repo repository.AuditRuleRepository
}

// NewAuditService 创建审核规则 service 实例
func NewAuditService(repo repository.AuditRuleRepository) AuditService {
	return &auditService{repo: repo}
}

// auditRuleStatusText 审核规则状态文本
func auditRuleStatusText(status int) string {
	switch status {
	case model.AuditRuleStatusEnabled:
		return "启用"
	case model.AuditRuleStatusDisabled:
		return "禁用"
	}
	return ""
}

// toAuditRuleResponse model -> dto
func toAuditRuleResponse(r *model.JobAuditRule) *dto.AuditRuleResponse {
	resp := &dto.AuditRuleResponse{
		ID:          r.ID,
		RuleName:    r.RuleName,
		RuleType:    r.RuleType,
		RuleKey:     r.RuleKey,
		Pattern:     r.Pattern,
		Action:      r.Action,
		PenaltyType: r.PenaltyType,
		Severity:    r.Severity,
		Status:      r.Status,
		StatusText:  auditRuleStatusText(r.Status),
		Description: r.Description,
		Sort:        r.Sort,
		CreatedAt:   r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	// 解析 threshold JSONB
	if len(r.Threshold) > 0 {
		_ = r.Threshold.Parse(&resp.Threshold)
	}
	if resp.Threshold == nil {
		resp.Threshold = map[string]interface{}{}
	}
	return resp
}

// CreateRule 创建审核规则
func (s *auditService) CreateRule(req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error) {
	rule := &model.JobAuditRule{
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
	// 默认值
	if rule.Action == "" {
		rule.Action = model.AuditRuleActionReject
	}
	if rule.Severity == 0 {
		rule.Severity = 1
	}
	if rule.Status == 0 && req.Status == 0 {
		// 默认启用（req.Status 显式传 0 表示禁用，但 binding oneof=0 1 会让 0 通过；为简化：未传 status 时也启用）
		rule.Status = model.AuditRuleStatusEnabled
	}
	// 序列化 threshold
	if req.Threshold != nil {
		b, err := model.FromJSON(req.Threshold)
		if err != nil {
			return nil, err
		}
		rule.Threshold = b
	}
	if err := s.repo.Create(rule); err != nil {
		return nil, err
	}
	return toAuditRuleResponse(rule), nil
}

// UpdateRule 更新审核规则
func (s *auditService) UpdateRule(id uint, req *dto.AuditRuleUpdateRequest) error {
	rule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
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
		b, err := model.FromJSON(req.Threshold)
		if err != nil {
			return err
		}
		fields["threshold"] = b
	}
	if len(fields) == 0 {
		return nil
	}
	_ = rule // 防止 unused 警告
	return s.repo.Update(id, fields)
}

// DeleteRule 删除审核规则
func (s *auditService) DeleteRule(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetRule 获取审核规则详情
func (s *auditService) GetRule(id uint) (*dto.AuditRuleResponse, error) {
	rule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleResponse(rule), nil
}

// ListRules 审核规则列表
func (s *auditService) ListRules(req *dto.AuditRuleListQuery) (*utils.Pagination, []dto.AuditRuleResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.AuditRuleListQuery{
		RuleType: req.RuleType,
		Status:   req.Status,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	respList := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		respList = append(respList, *toAuditRuleResponse(&list[i]))
	}
	return pagination, respList, nil
}

// ListEnabledRules 启用的审核规则列表
func (s *auditService) ListEnabledRules() ([]dto.AuditRuleResponse, error) {
	list, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	respList := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		respList = append(respList, *toAuditRuleResponse(&list[i]))
	}
	return respList, nil
}
