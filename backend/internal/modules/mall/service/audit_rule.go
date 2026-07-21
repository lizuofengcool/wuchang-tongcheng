// Package service 同城商城业务逻辑层 - 审核规则
// 敏感词/违禁内容/联系方式/价格校验/频率（全局表）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound = errors.New("审核规则不存在")
)

// AuditRuleService 审核规则业务接口
type AuditRuleService interface {
	Create(req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error)
	Update(id uint, req *dto.UpdateAuditRuleRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.AuditRuleInfo, error)
	List(req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error)
	ListEnabled() ([]dto.AuditRuleInfo, error)
	ListByType(ruleType string) ([]dto.AuditRuleInfo, error)
	UpdateStatus(id uint, status int) error

	// Check 内容审核检测（M 端 + 内部调用）
	Check(req *dto.AuditCheckRequest) (*dto.AuditCheckResponse, error)
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
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
func toAuditRuleInfo(r *model.AuditRule) *dto.AuditRuleInfo {
	info := &dto.AuditRuleInfo{
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
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if r.Threshold != nil {
		info.Threshold = r.Threshold
	}
	return info
}

// Create 创建审核规则
func (s *auditRuleService) Create(req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error) {
	status := req.Status
	if status == 0 {
		status = 1 // 默认启用
	}
	severity := req.Severity
	if severity == 0 {
		severity = 1
	}
	action := req.Action
	if action == "" {
		action = "review"
	}

	r := &model.AuditRule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		RuleKey:     req.RuleKey,
		Pattern:     req.Pattern,
		Action:      action,
		PenaltyType: req.PenaltyType,
		Severity:    severity,
		Status:      status,
		Description: req.Description,
		Sort:        req.Sort,
	}

	if req.Threshold != nil {
		if b, err := model.FromJSON(req.Threshold); err == nil {
			r.Threshold = b
		}
	}

	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}

// Update 更新审核规则
func (s *auditRuleService) Update(id uint, req *dto.UpdateAuditRuleRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.RuleName != nil {
		fields["rule_name"] = *req.RuleName
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
		if b, err := model.FromJSON(req.Threshold); err == nil {
			fields["threshold"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除审核规则
func (s *auditRuleService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
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
	opts := repository.AuditRuleListOptions{
		RuleType: req.RuleType,
		RuleKey:  req.RuleKey,
		Action:   req.Action,
		Status:   req.Status,
		Severity: req.Severity,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAuditRuleInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListEnabled 列出已启用的审核规则
func (s *auditRuleService) ListEnabled() ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}
	infos := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAuditRuleInfo(&list[i]))
	}
	return infos, nil
}

// ListByType 按类型列出已启用的审核规则
func (s *auditRuleService) ListByType(ruleType string) ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListByType(ruleType)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAuditRuleInfo(&list[i]))
	}
	return infos, nil
}

// UpdateStatus 更新审核规则状态
func (s *auditRuleService) UpdateStatus(id uint, status int) error {
	if status != 0 && status != 1 {
		return errors.New("状态值无效")
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

// Check 内容审核检测
// 简单实现：对文本类内容应用敏感词/违禁内容规则匹配
func (s *auditRuleService) Check(req *dto.AuditCheckRequest) (*dto.AuditCheckResponse, error) {
	resp := &dto.AuditCheckResponse{
		Pass:     true,
		Action:   "pass",
		HitRules: []dto.AuditRuleInfo{},
		Score:    0,
	}

	// 取出所有启用的规则
	rules, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}

	// 简化实现：仅返回通过状态，实际命中需根据规则模式匹配内容
	// 真实场景需根据 type（keyword/image/text/behavior/price/stock）和 pattern 进行正则/关键词匹配
	_ = rules
	return resp, nil
}
