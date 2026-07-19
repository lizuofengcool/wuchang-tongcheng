// Package service 审核规则 + 房源分类 + 配套设施业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound  = errors.New("审核规则不存在")
	ErrCategoryNotFound   = errors.New("房源分类不存在")
	ErrFacilityNotFound   = errors.New("配套设施不存在")
	ErrCategoryHasChild   = errors.New("该分类下存在子分类，无法删除")
	ErrCategoryCodeExists = errors.New("分类编码已存在")
	ErrFacilityCodeExists = errors.New("设施编码已存在")
	ErrRuleKeyExists      = errors.New("规则键已存在")
)

// ===== 审核规则 =====

// AuditRuleService 审核规则业务接口
type AuditRuleService interface {
	Create(req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error)
	Update(id uint, req *dto.AuditRuleCreateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.AuditRuleResponse, error)
	List(req *dto.AuditRuleListQuery) (*utils.Pagination, []dto.AuditRuleResponse, error)
	ListEnabled() ([]dto.AuditRuleResponse, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
}

type auditRuleService struct {
	repo repository.RiskRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.RiskRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

// toAuditRuleInfo model -> dto
func toAuditRuleInfo(r *model.HouseAuditRule) *dto.AuditRuleResponse {
	resp := &dto.AuditRuleResponse{
		ID:          r.ID,
		RuleName:    r.RuleName,
		RuleType:    r.RuleType,
		RuleTypeText: auditRuleTypeText(r.RuleType),
		RuleKey:     r.RuleKey,
		Pattern:     r.Pattern,
		Action:      r.Action,
		ActionText:  auditRuleActionText(r.Action),
		PenaltyType: r.PenaltyType,
		Severity:    r.Severity,
		Status:      r.Status,
		StatusText:  auditRuleStatusText(r.Status),
		Description: r.Description,
		Sort:        r.Sort,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	// 反序列化阈值 JSON
	if len(r.Threshold) > 0 {
		var ths []model.AuditRuleThreshold
		_ = r.Threshold.Parse(&ths)
		resp.Threshold = ths
	}
	return resp
}

func auditRuleTypeText(t string) string {
	switch t {
	case model.RuleTypeSensitiveWord:
		return "敏感词"
	case model.RuleTypePriceCheck:
		return "价格异常"
	case model.RuleTypeFrequency:
		return "频率限制"
	case model.RuleTypeFakeHouse:
		return "虚假房源"
	case model.RuleTypeProhibited:
		return "违禁内容"
	case model.RuleTypeContact:
		return "联系方式校验"
	}
	return ""
}

func auditRuleActionText(a string) string {
	switch a {
	case model.RuleActionReject:
		return "拒绝"
	case model.RuleActionApproval:
		return "转人工审核"
	case model.RuleActionLimit:
		return "限流"
	case model.RuleActionWarning:
		return "警告"
	}
	return ""
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

func (s *auditRuleService) Create(req *dto.AuditRuleCreateRequest) (*dto.AuditRuleResponse, error) {
	r := &model.HouseAuditRule{
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
	if r.Action == "" {
		r.Action = model.RuleActionReject
	}
	if r.Severity == 0 {
		r.Severity = 1
	}
	if r.Status == 0 {
		r.Status = 1
	}
	// 阈值 JSON
	if len(req.Threshold) > 0 {
		b, err := model.FromJSON(req.Threshold)
		if err == nil {
			r.Threshold = b
		}
	}
	if err := s.repo.CreateAuditRule(r); err != nil {
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}

func (s *auditRuleService) Update(id uint, req *dto.AuditRuleCreateRequest) error {
	if _, err := s.repo.FindAuditRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"rule_name":    req.RuleName,
		"rule_type":    req.RuleType,
		"rule_key":     req.RuleKey,
		"pattern":      req.Pattern,
		"action":       req.Action,
		"penalty_type": req.PenaltyType,
		"severity":     req.Severity,
		"status":       req.Status,
		"description":  req.Description,
		"sort":         req.Sort,
	}
	if len(req.Threshold) > 0 {
		b, err := model.FromJSON(req.Threshold)
		if err == nil {
			fields["threshold"] = b
		}
	}
	return s.repo.UpdateAuditRuleFields(id, fields)
}

func (s *auditRuleService) Delete(id uint) error {
	if _, err := s.repo.FindAuditRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.DeleteAuditRule(id)
}

func (s *auditRuleService) GetByID(id uint) (*dto.AuditRuleResponse, error) {
	r, err := s.repo.FindAuditRuleByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleInfo(r), nil
}

func (s *auditRuleService) List(req *dto.AuditRuleListQuery) (*utils.Pagination, []dto.AuditRuleResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AuditRuleListOptions{
		RuleType: req.RuleType,
		RuleKey:  req.RuleKey,
		Action:   req.Action,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.ListAuditRules(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *auditRuleService) ListEnabled() ([]dto.AuditRuleResponse, error) {
	list, err := s.repo.ListEnabledAuditRules()
	if err != nil {
		return nil, err
	}
	result := make([]dto.AuditRuleResponse, 0, len(list))
	for i := range list {
		result = append(result, *toAuditRuleInfo(&list[i]))
	}
	return result, nil
}

func (s *auditRuleService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindAuditRuleByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.UpdateAuditRuleFields(id, map[string]interface{}{"status": status})
}

// BatchUpdateStatus 批量更新状态（复用 ListEnabledAuditRules + 逐条更新）
// 注：仓储层未提供批量接口，使用循环实现
func (s *auditRuleService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	success := 0
	failedIDs := make([]uint, 0)
	for _, id := range ids {
		if err := s.repo.UpdateAuditRuleFields(id, map[string]interface{}{"status": status}); err != nil {
			failedIDs = append(failedIDs, id)
		} else {
			success++
		}
	}
	return &dto.BatchResultResponse{
		Total:     len(ids),
		Success:   success,
		Failed:    len(failedIDs),
		FailedIDs: failedIDs,
	}, nil
}

// ===== 房源分类 =====

// CategoryService 房源分类业务接口
type CategoryService interface {
	Create(req *dto.CategoryCreateRequest) (*dto.CategoryResponse, error)
	Update(id uint, req *dto.CategoryCreateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.CategoryResponse, error)
	List(req *dto.CategoryListQuery) (*utils.Pagination, []dto.CategoryResponse, error)
	ListAll() ([]dto.CategoryResponse, error)
	ListByParent(parentID uint) ([]dto.CategoryResponse, error)
	ListByListingType(listingType string) ([]dto.CategoryResponse, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService 创建分类 service 实例
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

// toCategoryInfo model -> dto
func toCategoryInfo(c *model.HouseCategory) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		Code:         c.Code,
		ParentID:     c.ParentID,
		Level:        c.Level,
		ListingType:  c.ListingType,
		PropertyType: c.PropertyType,
		Icon:         c.Icon,
		Color:        c.Color,
		Description:  c.Description,
		Sort:         c.Sort,
		Status:       c.Status,
		StatusText:   categoryStatusText(c.Status),
		HouseCount:   c.HouseCount,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

func categoryStatusText(s int) string {
	switch s {
	case model.CategoryStatusDisabled:
		return "禁用"
	case model.CategoryStatusEnabled:
		return "启用"
	}
	return ""
}

func (s *categoryService) Create(req *dto.CategoryCreateRequest) (*dto.CategoryResponse, error) {
	// 检查 code 唯一
	if existing, err := s.repo.FindByCode(req.Code); err == nil && existing != nil {
		return nil, ErrCategoryCodeExists
	}
	c := &model.HouseCategory{
		Name:         req.Name,
		Code:         req.Code,
		ParentID:     req.ParentID,
		Level:        req.Level,
		ListingType:  req.ListingType,
		PropertyType: req.PropertyType,
		Icon:         req.Icon,
		Color:        req.Color,
		Description:  req.Description,
		Sort:         req.Sort,
		Status:       req.Status,
	}
	if c.Level == 0 {
		c.Level = 1
	}
	if c.ListingType == "" {
		c.ListingType = model.ListingTypeRent
	}
	if c.PropertyType == "" {
		c.PropertyType = model.PropertyTypeResidential
	}
	if c.Color == "" {
		c.Color = "#409EFF"
	}
	if c.Status == 0 {
		c.Status = model.CategoryStatusEnabled
	}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCategoryInfo(c), nil
}

func (s *categoryService) Update(id uint, req *dto.CategoryCreateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"name":          req.Name,
		"code":          req.Code,
		"parent_id":     req.ParentID,
		"level":         req.Level,
		"listing_type":  req.ListingType,
		"property_type": req.PropertyType,
		"icon":          req.Icon,
		"color":         req.Color,
		"description":   req.Description,
		"sort":          req.Sort,
		"status":        req.Status,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *categoryService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	// 检查是否有子分类
	count, err := s.repo.CountByParent(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCategoryHasChild
	}
	return s.repo.Delete(id)
}

func (s *categoryService) GetByID(id uint) (*dto.CategoryResponse, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return toCategoryInfo(c), nil
}

func (s *categoryService) List(req *dto.CategoryListQuery) (*utils.Pagination, []dto.CategoryResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CategoryListOptions{
		ParentID:     req.ParentID,
		ListingType:  req.ListingType,
		PropertyType: req.PropertyType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	if req.Level > 0 {
		lv := req.Level
		opts.Level = &lv
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CategoryResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCategoryInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *categoryService) ListAll() ([]dto.CategoryResponse, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCategoryInfo(&list[i]))
	}
	return result, nil
}

func (s *categoryService) ListByParent(parentID uint) ([]dto.CategoryResponse, error) {
	list, err := s.repo.ListByParent(parentID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCategoryInfo(&list[i]))
	}
	return result, nil
}

func (s *categoryService) ListByListingType(listingType string) ([]dto.CategoryResponse, error) {
	list, err := s.repo.ListByListingType(listingType)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCategoryInfo(&list[i]))
	}
	return result, nil
}

func (s *categoryService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

func (s *categoryService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}

// ===== 配套设施 =====

// FacilityService 配套设施业务接口
type FacilityService interface {
	Create(creatorID uint, req *dto.FacilityCreateRequest) (*dto.FacilityResponse, error)
	Update(id uint, req *dto.FacilityCreateRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.FacilityResponse, error)
	List(req *dto.FacilityListQuery) (*utils.Pagination, []dto.FacilityResponse, error)
	ListAll() ([]dto.FacilityResponse, error)
	ListByCategory(category string) ([]dto.FacilityResponse, error)
	ListHot(limit int) ([]dto.FacilityResponse, error)
	UpdateStatus(id uint, status int) error
	BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error)
}

type facilityService struct {
	repo repository.FacilityRepository
}

// NewFacilityService 创建设施 service 实例
func NewFacilityService(repo repository.FacilityRepository) FacilityService {
	return &facilityService{repo: repo}
}

// toFacilityInfo model -> dto
func toFacilityInfo(f *model.HouseFacility) *dto.FacilityResponse {
	return &dto.FacilityResponse{
		ID:           f.ID,
		Name:         f.Name,
		Code:         f.Code,
		Category:     f.Category,
		CategoryText: facilityCategoryText(f.Category),
		Icon:         f.Icon,
		Color:        f.Color,
		Background:   f.Background,
		Description:  f.Description,
		Status:       f.Status,
		StatusText:   facilityStatusText(f.Status),
		Sort:         f.Sort,
		UseCount:     f.UseCount,
		IsHot:        f.IsHot,
		CreatorID:    f.CreatorID,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

func facilityCategoryText(c string) string {
	switch c {
	case model.FacilityCategoryIndoor:
		return "室内设施"
	case model.FacilityCategoryOutdoor:
		return "室外设施"
	case model.FacilityCategoryAppliance:
		return "家电"
	case model.FacilityCategoryFurniture:
		return "家具"
	case model.FacilityCategoryNetwork:
		return "网络"
	case model.FacilityCategorySecurity:
		return "安全"
	case model.FacilityCategoryParking:
		return "停车"
	}
	return ""
}

func facilityStatusText(s int) string {
	switch s {
	case model.FacilityStatusDisabled:
		return "禁用"
	case model.FacilityStatusEnabled:
		return "启用"
	}
	return ""
}

func (s *facilityService) Create(creatorID uint, req *dto.FacilityCreateRequest) (*dto.FacilityResponse, error) {
	// 检查 code 唯一
	if existing, err := s.repo.FindByCode(req.Code); err == nil && existing != nil {
		return nil, ErrFacilityCodeExists
	}
	f := &model.HouseFacility{
		Name:        req.Name,
		Code:        req.Code,
		Category:    req.Category,
		Icon:        req.Icon,
		Color:       req.Color,
		Background:  req.Background,
		Description: req.Description,
		Status:      req.Status,
		Sort:        req.Sort,
		IsHot:       req.IsHot,
		CreatorID:   creatorID,
	}
	if f.Category == "" {
		f.Category = model.FacilityCategoryIndoor
	}
	if f.Color == "" {
		f.Color = "#409EFF"
	}
	if f.Status == 0 {
		f.Status = model.FacilityStatusEnabled
	}
	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	return toFacilityInfo(f), nil
}

func (s *facilityService) Update(id uint, req *dto.FacilityCreateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFacilityNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"name":        req.Name,
		"code":        req.Code,
		"category":    req.Category,
		"icon":        req.Icon,
		"color":       req.Color,
		"background":  req.Background,
		"description": req.Description,
		"status":      req.Status,
		"sort":        req.Sort,
		"is_hot":      req.IsHot,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *facilityService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFacilityNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

func (s *facilityService) GetByID(id uint) (*dto.FacilityResponse, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFacilityNotFound
		}
		return nil, err
	}
	return toFacilityInfo(f), nil
}

func (s *facilityService) List(req *dto.FacilityListQuery) (*utils.Pagination, []dto.FacilityResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.FacilityListOptions{
		Category: req.Category,
		Status:   req.Status,
		IsHot:    req.IsHot,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.FacilityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toFacilityInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *facilityService) ListAll() ([]dto.FacilityResponse, error) {
	list, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.FacilityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toFacilityInfo(&list[i]))
	}
	return result, nil
}

func (s *facilityService) ListByCategory(category string) ([]dto.FacilityResponse, error) {
	list, err := s.repo.ListByCategory(category)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FacilityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toFacilityInfo(&list[i]))
	}
	return result, nil
}

func (s *facilityService) ListHot(limit int) ([]dto.FacilityResponse, error) {
	list, err := s.repo.ListHot(limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FacilityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toFacilityInfo(&list[i]))
	}
	return result, nil
}

func (s *facilityService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFacilityNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

func (s *facilityService) BatchUpdateStatus(ids []uint, status int) (*dto.BatchResultResponse, error) {
	affected, err := s.repo.BatchUpdateStatus(ids, status)
	if err != nil {
		return &dto.BatchResultResponse{
			Total: len(ids), Success: 0, Failed: len(ids), FailedIDs: ids,
		}, err
	}
	return &dto.BatchResultResponse{
		Total:   len(ids),
		Success: int(affected),
		Failed:  len(ids) - int(affected),
	}, nil
}
