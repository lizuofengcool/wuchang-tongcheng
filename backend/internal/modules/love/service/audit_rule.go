// Package service love 相亲交友业务逻辑层 - 审核规则
package service

import (
	"errors"
	"strings"
	"time"

	"wuchang-tongcheng/internal/modules/love/dto"
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/modules/love/repository"
	"wuchang-tongcheng/internal/pkg/utils"
)

var (
	ErrLoveAuditRuleNotFound = errors.New("审核规则不存在")
	ErrLoveAuditRuleKeyExists = errors.New("规则键已存在")
)

// LoveAuditRuleService 审核规则业务接口（M 端配置）
type LoveAuditRuleService interface {
	Create(req *dto.CreateLoveAuditRuleRequest) (*dto.LoveAuditRuleInfo, error)
	Update(id uint, req *dto.UpdateLoveAuditRuleRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.LoveAuditRuleInfo, error)
	GetByRuleKey(key string) (*dto.LoveAuditRuleInfo, error)
	List(req *dto.LoveAuditRuleListRequest) (*utils.Pagination, []dto.LoveAuditRuleInfo, error)
	ListAll() ([]dto.LoveAuditRuleInfo, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) error
	// Check 内容审核：依据规则匹配内容，返回命中规则与建议动作
	Check(req *dto.LoveAuditCheckRequest) (*dto.LoveAuditCheckResponse, error)
}

type loveAuditRuleService struct {
	repo repository.LoveAuditRuleRepository
}

// NewLoveAuditRuleService 创建审核规则 service
func NewLoveAuditRuleService(repo repository.LoveAuditRuleRepository) LoveAuditRuleService {
	return &loveAuditRuleService{repo: repo}
}

func auditRuleStatusText(s int) string {
	switch s {
	case 0:
		return "禁用"
	case 1:
		return "启用"
	}
	return ""
}

func auditRuleActionText(a string) string {
	switch a {
	case model.AuditActionReject:
		return "拒绝"
	case model.AuditActionReview:
		return "人工审核"
	case model.AuditActionBlock:
		return "封禁"
	case model.AuditActionShadow:
		return "影子封禁"
	case model.AuditActionWarning:
		return "警告"
	}
	return ""
}

func auditRuleSeverityText(s int) string {
	switch s {
	case model.AuditSeverityLow:
		return "低"
	case model.AuditSeverityMedium:
		return "中"
	case model.AuditSeverityHigh:
		return "高"
	case model.AuditSeverityCritical:
		return "严重"
	}
	return ""
}

func toLoveAuditRuleInfo(r *model.LoveAuditRule) dto.LoveAuditRuleInfo {
	info := dto.LoveAuditRuleInfo{
		ID:           r.ID,
		RuleName:     r.RuleName,
		RuleType:     r.RuleType,
		RuleKey:      r.RuleKey,
		Pattern:      r.Pattern,
		Action:       r.Action,
		ActionText:   auditRuleActionText(r.Action),
		PenaltyType:  r.PenaltyType,
		Severity:     r.Severity,
		SeverityText: auditRuleSeverityText(r.Severity),
		Scope:        r.Scope,
		TargetType:   r.TargetType,
		Description:  r.Description,
		Sort:         r.Sort,
		Status:       r.Status,
		StatusText:   auditRuleStatusText(r.Status),
	}
	if r.Threshold != nil {
		info.Threshold = r.Threshold
	}
	if !r.CreatedAt.IsZero() {
		info.CreatedAt = r.CreatedAt.Format(time.RFC3339)
	}
	if !r.UpdatedAt.IsZero() {
		info.UpdatedAt = r.UpdatedAt.Format(time.RFC3339)
	}
	return info
}

func (s *loveAuditRuleService) Create(req *dto.CreateLoveAuditRuleRequest) (*dto.LoveAuditRuleInfo, error) {
	if req.RuleKey != "" {
		if existing, err := s.repo.FindByRuleKey(req.RuleKey); err == nil && existing != nil {
			return nil, ErrLoveAuditRuleKeyExists
		}
	}
	r := &model.LoveAuditRule{
		RuleName:     req.RuleName,
		RuleType:     req.RuleType,
		RuleKey:      req.RuleKey,
		Pattern:      req.Pattern,
		Action:       req.Action,
		PenaltyType:  req.PenaltyType,
		Severity:     req.Severity,
		Scope:        req.Scope,
		TargetType:   req.TargetType,
		Description:  req.Description,
		Sort:         req.Sort,
		Status:       1,
	}
	if req.Status == 0 {
		r.Status = 0
	}
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			r.Threshold = jb
		}
	}
	if r.Action == "" {
		r.Action = model.AuditActionReject
	}
	if r.Severity == 0 {
		r.Severity = model.AuditSeverityLow
	}
	if r.Scope == "" {
		r.Scope = model.AuditScopeAll
	}
	if r.TargetType == "" {
		r.TargetType = "all"
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	info := toLoveAuditRuleInfo(r)
	return &info, nil
}

func (s *loveAuditRuleService) Update(id uint, req *dto.UpdateLoveAuditRuleRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return ErrLoveAuditRuleNotFound
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
	if req.Scope != nil {
		fields["scope"] = *req.Scope
	}
	if req.TargetType != nil {
		fields["target_type"] = *req.TargetType
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			fields["threshold"] = jb
		}
	}
	if len(fields) == 0 {
		return nil
	}
	_ = r
	return s.repo.UpdateFields(id, fields)
}

func (s *loveAuditRuleService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveAuditRuleNotFound
	}
	return s.repo.Delete(id)
}

func (s *loveAuditRuleService) GetByID(id uint) (*dto.LoveAuditRuleInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrLoveAuditRuleNotFound
	}
	info := toLoveAuditRuleInfo(r)
	return &info, nil
}

func (s *loveAuditRuleService) GetByRuleKey(key string) (*dto.LoveAuditRuleInfo, error) {
	r, err := s.repo.FindByRuleKey(key)
	if err != nil {
		return nil, ErrLoveAuditRuleNotFound
	}
	info := toLoveAuditRuleInfo(r)
	return &info, nil
}

func (s *loveAuditRuleService) List(req *dto.LoveAuditRuleListRequest) (*utils.Pagination, []dto.LoveAuditRuleInfo, error) {
	opts := repository.LoveAuditRuleListOptions{
		RuleType: req.RuleType,
		RuleKey:  req.RuleKey,
		Action:   req.Action,
		Keyword:  req.Keyword,
	}
	if req.Status != nil {
		opts.Status = req.Status
	}
	list, total, err := s.repo.List(&req.Pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.LoveAuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveAuditRuleInfo(&list[i]))
	}
	req.Pagination.Total = total
	return &req.Pagination, infos, nil
}

func (s *loveAuditRuleService) ListAll() ([]dto.LoveAuditRuleInfo, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.LoveAuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, toLoveAuditRuleInfo(&list[i]))
	}
	return infos, nil
}

func (s *loveAuditRuleService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return ErrLoveAuditRuleNotFound
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

func (s *loveAuditRuleService) BatchUpdateStatus(ids []uint, status int) error {
	return s.repo.BatchUpdateStatus(ids, status)
}

// Check 内容审核：将内容按行切分关键词与正则模式匹配
// MVP 实现：仅做关键词包含匹配，正则匹配由后续接入
func (s *loveAuditRuleService) Check(req *dto.LoveAuditCheckRequest) (*dto.LoveAuditCheckResponse, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	resp := &dto.LoveAuditCheckResponse{
		Passed: true,
	}
	if req.Content == "" {
		return resp, nil
	}
	content := strings.ToLower(req.Content)
	for i := range list {
		rule := &list[i]
		if rule.Pattern == "" {
			continue
		}
		// 按行分隔关键词，每个关键词都做包含匹配
		for _, kw := range strings.Split(rule.Pattern, "\n") {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			if strings.Contains(content, strings.ToLower(kw)) {
				resp.Passed = false
				resp.Action = rule.Action
				resp.Matched = append(resp.Matched, dto.LoveAuditMatchedItem{
					RuleID:   rule.ID,
					RuleName: rule.RuleName,
					RuleType: rule.RuleType,
					Pattern:  kw,
					Severity: rule.Severity,
				})
				break
			}
		}
	}
	if !resp.Passed && resp.Action == "" {
		resp.Action = model.AuditActionReject
	}
	if !resp.Passed {
		resp.Reason = "内容命中审核规则"
	}
	return resp, nil
}
