// Package service 同城车辆买卖业务逻辑层 - 车险配置
// 依据 v3.2.1 架构方案：对标平安/太平洋/人保 交强/商业/第三方
// 车险为全局配置数据（无 region_id），管理后台维护
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrInsuranceNotFound      = errors.New("车险方案不存在")
	ErrInsuranceStatusInvalid = errors.New("车险方案状态不允许此操作")
	ErrInsuranceParamsInvalid = errors.New("车险报价参数无效")
)

// InsuranceService 车险业务接口
type InsuranceService interface {
	// M 端管理
	Create(req *dto.CreateInsuranceRequest) (*dto.InsuranceInfo, error)
	Update(id uint, req *dto.UpdateInsuranceRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.InsuranceInfo, error)
	List(req *dto.InsuranceListRequest) (*utils.Pagination, []dto.InsuranceInfo, error)
	UpdateStatus(id uint, status int) error

	// C 端
	ListPublished(page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error)
	ListHot(limit int) ([]dto.InsuranceInfo, error)
	Quote(req *dto.InsuranceQuoteRequest) (*dto.InsuranceQuoteResponse, error)
}

type insuranceService struct {
	repo repository.InsuranceRepository
}

// NewInsuranceService 创建车险 service 实例
func NewInsuranceService(repo repository.InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

// insuranceStatusText 状态文本
func insuranceStatusText(status int) string {
	switch status {
	case model.InsurancePlanStatusDraft:
		return "草稿"
	case model.InsurancePlanStatusPublished:
		return "已发布"
	case model.InsurancePlanStatusOffline:
		return "已下架"
	}
	return ""
}

// toInsuranceInfo model -> dto
func toInsuranceInfo(i *model.CarInsurance) *dto.InsuranceInfo {
	return &dto.InsuranceInfo{
		ID:             i.ID,
		Name:           i.Name,
		Code:           i.Code,
		InsuranceType:  i.InsuranceType,
		Provider:       i.Provider,
		Coverage:       i.Coverage,
		CoverageAmount: i.CoverageAmount,
		Premium:        i.Premium,
		Deductible:     i.Deductible,
		Description:    i.Description,
		Sort:           i.Sort,
		Status:         i.Status,
		StatusText:     insuranceStatusText(i.Status),
		IsHot:          i.IsHot,
		UseCount:       i.UseCount,
		CreatedAt:      i.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      i.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ===== M 端管理 =====

// Create 创建车险
func (s *insuranceService) Create(req *dto.CreateInsuranceRequest) (*dto.InsuranceInfo, error) {
	i := &model.CarInsurance{
		Name:           req.Name,
		Code:           req.Code,
		InsuranceType:  req.InsuranceType,
		Provider:       req.Provider,
		Coverage:       req.Coverage,
		CoverageAmount: req.CoverageAmount,
		Premium:        req.Premium,
		Deductible:     req.Deductible,
		Description:    req.Description,
		Sort:           req.Sort,
		Status:         req.Status,
		IsHot:          req.IsHot,
	}

	// 默认值兜底
	if i.InsuranceType == "" {
		i.InsuranceType = model.InsuranceTypeCompulsory
	}
	if i.Status == 0 {
		i.Status = model.InsurancePlanStatusDraft
	}

	if err := s.repo.Create(i); err != nil {
		return nil, err
	}
	return toInsuranceInfo(i), nil
}

// Update 更新车险
func (s *insuranceService) Update(id uint, req *dto.UpdateInsuranceRequest) error {
	fields := map[string]interface{}{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Code != nil {
		fields["code"] = *req.Code
	}
	if req.InsuranceType != nil {
		fields["insurance_type"] = *req.InsuranceType
	}
	if req.Provider != nil {
		fields["provider"] = *req.Provider
	}
	if req.Coverage != nil {
		fields["coverage"] = *req.Coverage
	}
	if req.CoverageAmount != nil {
		fields["coverage_amount"] = *req.CoverageAmount
	}
	if req.Premium != nil {
		fields["premium"] = *req.Premium
	}
	if req.Deductible != nil {
		fields["deductible"] = *req.Deductible
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
	if req.IsHot != nil {
		fields["is_hot"] = *req.IsHot
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除车险
func (s *insuranceService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *insuranceService) GetByID(id uint) (*dto.InsuranceInfo, error) {
	i, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInsuranceNotFound
		}
		return nil, err
	}
	return toInsuranceInfo(i), nil
}

// List 列表查询
func (s *insuranceService) List(req *dto.InsuranceListRequest) (*utils.Pagination, []dto.InsuranceInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.InsuranceListQuery{
		InsuranceType: req.InsuranceType,
		Provider:      req.Provider,
		Status:        req.Status,
		IsHot:         req.IsHot,
		Keyword:       req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InsuranceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInsuranceInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 更新状态
func (s *insuranceService) UpdateStatus(id uint, status int) error {
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== C 端 =====

// ListPublished C 端列表（仅已发布）
func (s *insuranceService) ListPublished(page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListPublished(pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.InsuranceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInsuranceInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListHot 热门车险
func (s *insuranceService) ListHot(limit int) ([]dto.InsuranceInfo, error) {
	list, err := s.repo.ListHot(limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.InsuranceInfo, 0, len(list))
	for i := range list {
		result = append(result, *toInsuranceInfo(&list[i]))
	}
	return result, nil
}

// Quote 车险报价（C 端）
// MVP 实现：直接累加用户选择的险种保费与保额
func (s *insuranceService) Quote(req *dto.InsuranceQuoteRequest) (*dto.InsuranceQuoteResponse, error) {
	if len(req.InsuranceIDs) == 0 {
		return nil, ErrInsuranceParamsInvalid
	}

	list, err := s.repo.ListByIDs(req.InsuranceIDs)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrInsuranceNotFound
	}

	var totalPremium, totalCoverage float64
	items := make([]dto.InsuranceItem, 0, len(list))
	for _, ins := range list {
		// 简单保费调整：按车价比例（商业险通常按车价 × 费率）
		premium := ins.Premium
		if req.CarPrice > 0 && ins.InsuranceType != model.InsuranceTypeCompulsory {
			// 商业险按车价 1.5% 估算（MVP 简化）
			if premium == 0 {
				premium = req.CarPrice * 0.015
			}
		}

		items = append(items, dto.InsuranceItem{
			ID:             ins.ID,
			Name:           ins.Name,
			InsuranceType:  ins.InsuranceType,
			Provider:       ins.Provider,
			Coverage:       ins.Coverage,
			CoverageAmount: ins.CoverageAmount,
			Premium:        premium,
			Deductible:     ins.Deductible,
		})
		totalPremium += premium
		totalCoverage += ins.CoverageAmount

		// 增加使用次数
		_ = s.repo.IncrUseCount(ins.ID, 1)
	}

	return &dto.InsuranceQuoteResponse{
		TotalPremium:  totalPremium,
		TotalCoverage: totalCoverage,
		Items:         items,
	}, nil
}
