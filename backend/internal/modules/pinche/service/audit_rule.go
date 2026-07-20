// Package service 同城拼车出行业务逻辑层 - 审核规则
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound      = errors.New("审核规则不存在")
	ErrAuditRuleStatusInvalid = errors.New("审核规则状态不允许此操作")
	ErrAuditRuleCodeUsed      = errors.New("规则编码已存在")
)

// AuditRuleService 审核规则业务接口
type AuditRuleService interface {
	// C 端只读
	GetByID(id uint) (*dto.AuditRuleInfo, error)
	List(regionID uint, req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error)
	ListEnabled(regionID uint) ([]dto.AuditRuleInfo, error)
	ListByType(regionID uint, ruleType string) ([]dto.AuditRuleInfo, error)

	// M 端
	Create(regionID uint, req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error)
	Update(id uint, req *dto.UpdateAuditRuleRequest) error
	Delete(id uint) error
	UpdateStatus(id uint, status int) error
	IncrHitCount(id uint) error
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

// toAuditRuleInfo model -> dto
func toAuditRuleInfo(a *model.PincheAuditRule) *dto.AuditRuleInfo {
	info := &dto.AuditRuleInfo{
		ID:          a.ID,
		RegionID:    a.RegionID,
		RuleName:    a.RuleName,
		RuleType:    a.RuleType,
		RuleCode:    a.RuleCode,
		Description: a.Description,
		Action:      a.Action,
		Priority:    a.Priority,
		Status:      a.Status,
		HitCount:    a.HitCount,
		LastHitAt:   a.LastHitAt,
		CreatedAt:   a.CreatedAt,
	}
	if a.Threshold != nil {
		info.Threshold = a.Threshold
	}
	return info
}

// GetByID 获取详情
func (s *auditRuleService) GetByID(id uint) (*dto.AuditRuleInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleInfo(a), nil
}

// List 规则列表
func (s *auditRuleService) List(regionID uint, req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AuditRuleListOptions{
		RuleType: req.RuleType,
		RuleCode: req.RuleCode,
		Action:   req.Action,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
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

// ListEnabled 启用规则列表
func (s *auditRuleService) ListEnabled(regionID uint) ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListEnabled(regionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return result, nil
}

// ListByType 按类型查询
func (s *auditRuleService) ListByType(regionID uint, ruleType string) ([]dto.AuditRuleInfo, error) {
	list, err := s.repo.ListByType(regionID, ruleType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return result, nil
}

// Create 创建审核规则
func (s *auditRuleService) Create(regionID uint, req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error) {
	// 规则编码唯一性校验
	if req.RuleCode != "" {
		if existing, err := s.repo.FindByCode(req.RuleCode); err == nil && existing != nil {
			return nil, ErrAuditRuleCodeUsed
		}
	}
	action := req.Action
	if action == "" {
		action = model.AuditRuleActionManualReview
	}
	a := &model.PincheAuditRule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		RuleCode:    req.RuleCode,
		Description: req.Description,
		Action:      action,
		Priority:    req.Priority,
		Status:      1,
	}
	a.RegionID = regionID
	if req.Threshold != nil {
		if jb, err := model.FromJSON(req.Threshold); err == nil {
			a.Threshold = jb
		}
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toAuditRuleInfo(a), nil
}

// Update 更新规则
func (s *auditRuleService) Update(id uint, req *dto.UpdateAuditRuleRequest) error {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	_ = a
	fields := map[string]interface{}{}
	if req.RuleName != nil {
		fields["rule_name"] = *req.RuleName
	}
	if req.RuleType != nil {
		fields["rule_type"] = *req.RuleType
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Action != nil {
		fields["action"] = *req.Action
	}
	if req.Priority != nil {
		fields["priority"] = *req.Priority
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
	return s.repo.Update(id, fields)
}

// Delete 删除规则
func (s *auditRuleService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// UpdateStatus 更新状态
func (s *auditRuleService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}

// IncrHitCount 命中次数 +1
func (s *auditRuleService) IncrHitCount(id uint) error {
	return s.repo.IncrHitCount(id)
}
