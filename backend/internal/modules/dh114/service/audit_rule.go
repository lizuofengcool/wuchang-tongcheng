// Package service 同城114业务逻辑层 - 审核规则 + 推荐商家
// 依据 v3.2.1 架构方案：敏感词/违禁内容/联系方式/价格校验/频率
// 推荐商家：首页推荐/分类推荐/附近推荐/个性化推荐
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound      = errors.New("审核规则不存在")
	ErrRecommendationNotFound = errors.New("推荐记录不存在")
)

// ===== 审核规则服务 =====

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
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

// auditRuleTypeText 审核规则类型文本
func auditRuleTypeText(t string) string {
	switch t {
	case model.AuditRuleTypeSensitiveWord:
		return "敏感词"
	case model.AuditRuleTypeProhibited:
		return "违禁内容"
	case model.AuditRuleTypeContact:
		return "联系方式"
	case model.AuditRuleTypePriceCheck:
		return "价格校验"
	case model.AuditRuleTypeFrequency:
		return "频率"
	}
	return ""
}

// auditActionText 审核动作文本
func auditActionText(a string) string {
	switch a {
	case model.AuditActionReject:
		return "拒绝"
	case model.AuditActionApproval:
		return "通过"
	case model.AuditActionFilter:
		return "过滤"
	case model.AuditActionLimit:
		return "限制"
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
func toAuditRuleInfo(r *model.Dh114AuditRule) *dto.AuditRuleInfo {
	info := &dto.AuditRuleInfo{
		ID:           r.ID,
		RuleName:     r.RuleName,
		RuleType:     r.RuleType,
		RuleTypeText: auditRuleTypeText(r.RuleType),
		RuleKey:      r.RuleKey,
		Pattern:      r.Pattern,
		Action:       r.Action,
		ActionText:   auditActionText(r.Action),
		PenaltyType:  r.PenaltyType,
		Severity:     r.Severity,
		Status:       r.Status,
		StatusText:   auditRuleStatusText(r.Status),
		Description:  r.Description,
		Sort:         r.Sort,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
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
	action := req.Action
	if action == "" {
		action = model.AuditActionReject
	}
	severity := req.Severity
	if severity == 0 {
		severity = 1
	}

	r := &model.Dh114AuditRule{
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
		if b, err := model.FromJSON(req.Threshold); err == nil {
			fields["threshold"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
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
	query := repository.AuditRuleListQuery{
		RuleType: req.RuleType,
		RuleKey:  req.RuleKey,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
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
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== 推荐商家服务 =====

// RecommendationService 推荐商家业务接口
type RecommendationService interface {
	// C 端查询
	ListByType(recommendType string, regionID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error)
	GetByID(id uint) (*dto.RecommendationInfo, error)

	// 互动反馈
	MarkClicked(id uint) error
	MarkContacted(id uint) error
	MarkDismissed(id uint) error

	// M 端管理
	Create(regionID uint, req *dto.RecommendationListRequest) (*dto.RecommendationInfo, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error)
}

type recommendationService struct {
	repo repository.RecommendationRepository
}

// NewRecommendationService 创建推荐 service 实例
func NewRecommendationService(repo repository.RecommendationRepository) RecommendationService {
	return &recommendationService{repo: repo}
}

// recommendStatusText 推荐状态文本
func recommendStatusText(s int) string {
	switch s {
	case model.RecommendStatusShown:
		return "已展示"
	case model.RecommendStatusClicked:
		return "已点击"
	case model.RecommendStatusContacted:
		return "已联系"
	case model.RecommendStatusDismissed:
		return "已忽略"
	}
	return ""
}

// toRecommendationInfo model -> dto
func toRecommendationInfo(r *model.Dh114Recommendation) *dto.RecommendationInfo {
	return &dto.RecommendationInfo{
		ID:            r.ID,
		UserID:        r.UserID,
		Dh114ID:       r.Dh114ID,
		BusinessID:    r.BusinessID,
		RecommendType: r.RecommendType,
		Position:      r.Position,
		Score:         r.Score,
		Reason:        r.Reason,
		CategoryID:    r.CategoryID,
		ExpireAt:      r.ExpireAt,
		Status:        r.Status,
		StatusText:    recommendStatusText(r.Status),
		ClickedAt:     r.ClickedAt,
		ContactedAt:   r.ContactedAt,
		DismissedAt:   r.DismissedAt,
		RegionID:      r.RegionID,
		CreatedAt:     r.CreatedAt,
	}
}

// ListByType 按推荐类型列出（如首页推荐）
func (s *recommendationService) ListByType(recommendType string, regionID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	if recommendType == "" {
		recommendType = model.RecommendTypeHome
	}
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByType(recommendType, regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出推荐
func (s *recommendationService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出推荐
func (s *recommendationService) ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// GetByID 获取推荐详情
func (s *recommendationService) GetByID(id uint) (*dto.RecommendationInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecommendationNotFound
		}
		return nil, err
	}
	return toRecommendationInfo(r), nil
}

// MarkClicked 标记已点击
func (s *recommendationService) MarkClicked(id uint) error {
	now := time.Now()
	return s.repo.MarkClicked(id, now)
}

// MarkContacted 标记已联系
func (s *recommendationService) MarkContacted(id uint) error {
	now := time.Now()
	return s.repo.MarkContacted(id, now)
}

// MarkDismissed 标记已忽略
func (s *recommendationService) MarkDismissed(id uint) error {
	now := time.Now()
	return s.repo.MarkDismissed(id, now)
}

// Create 创建推荐（M 端）
func (s *recommendationService) Create(regionID uint, req *dto.RecommendationListRequest) (*dto.RecommendationInfo, error) {
	r := &model.Dh114Recommendation{
		UserID:        req.UserID,
		Dh114ID:       req.Dh114ID,
		RecommendType: req.RecommendType,
		Position:      0,
		Score:         0,
		Status:        model.RecommendStatusShown,
	}
	r.RegionID = regionID
	if r.RecommendType == "" {
		r.RecommendType = model.RecommendTypeHome
	}
	if req.CategoryID > 0 {
		r.CategoryID = &req.CategoryID
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toRecommendationInfo(r), nil
}

// Update 更新推荐（M 端）
func (s *recommendationService) Update(id uint, fields map[string]interface{}) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	return s.repo.Update(id, fields)
}

// Delete 删除推荐（M 端）
func (s *recommendationService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecommendationNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// List 推荐列表（M 端）
func (s *recommendationService) List(req *dto.RecommendationListRequest) (*utils.Pagination, []dto.RecommendationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.RecommendationListQuery{
		UserID:        req.UserID,
		Dh114ID:       req.Dh114ID,
		RecommendType: req.RecommendType,
		CategoryID:    req.CategoryID,
		Status:        req.Status,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.RecommendationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRecommendationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}
