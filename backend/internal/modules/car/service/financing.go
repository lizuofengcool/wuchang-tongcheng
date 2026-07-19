// Package service 同城车辆买卖业务逻辑层 - 分期付款方案
// 依据 v3.2.1 架构方案：对标易鑫车贷/毛豆新车 首付/月供/利率/期数
// 分期方案为全局配置数据（无 region_id），管理后台维护
package service

import (
	"errors"
	"math"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrFinancingNotFound      = errors.New("分期方案不存在")
	ErrFinancingStatusInvalid = errors.New("分期方案状态不允许此操作")
	ErrFinancingParamsInvalid = errors.New("分期计算参数无效")
)

// FinancingService 分期方案业务接口
type FinancingService interface {
	// M 端管理
	Create(req *dto.CreateFinancingRequest) (*dto.FinancingInfo, error)
	Update(id uint, req *dto.UpdateFinancingRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.FinancingInfo, error)
	List(req *dto.FinancingListRequest) (*utils.Pagination, []dto.FinancingInfo, error)
	UpdateStatus(id uint, status int) error

	// C 端
	ListPublished(page, pageSize int) (*utils.Pagination, []dto.FinancingInfo, error)
	ListHot(limit int) ([]dto.FinancingInfo, error)
	Calculate(req *dto.FinancingCalculateRequest) (*dto.FinancingCalculateResponse, error)
}

type financingService struct {
	repo repository.FinancingRepository
}

// NewFinancingService 创建分期方案 service 实例
func NewFinancingService(repo repository.FinancingRepository) FinancingService {
	return &financingService{repo: repo}
}

// financingStatusText 状态文本
func financingStatusText(status int) string {
	switch status {
	case model.FinancingStatusDraft:
		return "草稿"
	case model.FinancingStatusPublished:
		return "已发布"
	case model.FinancingStatusOffline:
		return "已下架"
	}
	return ""
}

// toFinancingInfo model -> dto
func toFinancingInfo(f *model.CarFinancing) *dto.FinancingInfo {
	return &dto.FinancingInfo{
		ID:             f.ID,
		Name:           f.Name,
		Code:           f.Code,
		FinancingType:  f.FinancingType,
		MinDownPayment: f.MinDownPayment,
		MaxDownPayment: f.MaxDownPayment,
		InterestRate:   f.InterestRate,
		AnnualRate:     f.AnnualRate,
		MinPeriods:     f.MinPeriods,
		MaxPeriods:     f.MaxPeriods,
		MaxAmount:      f.MaxAmount,
		Provider:       f.Provider,
		Description:    f.Description,
		Sort:           f.Sort,
		Status:         f.Status,
		StatusText:     financingStatusText(f.Status),
		IsHot:          f.IsHot,
		UseCount:       f.UseCount,
		CreatedAt:      f.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      f.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ===== M 端管理 =====

// Create 创建分期方案
func (s *financingService) Create(req *dto.CreateFinancingRequest) (*dto.FinancingInfo, error) {
	f := &model.CarFinancing{
		Name:           req.Name,
		Code:           req.Code,
		FinancingType:  req.FinancingType,
		MinDownPayment: req.MinDownPayment,
		MaxDownPayment: req.MaxDownPayment,
		InterestRate:   req.InterestRate,
		AnnualRate:     req.AnnualRate,
		MinPeriods:     req.MinPeriods,
		MaxPeriods:     req.MaxPeriods,
		MaxAmount:      req.MaxAmount,
		Provider:       req.Provider,
		Description:    req.Description,
		Sort:           req.Sort,
		Status:         req.Status,
		IsHot:          req.IsHot,
	}

	// 默认值兜底
	if f.FinancingType == "" {
		f.FinancingType = model.FinancingTypeLoan
	}
	if f.MinDownPayment == 0 {
		f.MinDownPayment = 20.0
	}
	if f.MaxDownPayment == 0 {
		f.MaxDownPayment = 80.0
	}
	if f.MinPeriods == 0 {
		f.MinPeriods = 12
	}
	if f.MaxPeriods == 0 {
		f.MaxPeriods = 60
	}
	if f.Status == 0 {
		f.Status = model.FinancingStatusDraft
	}

	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	return toFinancingInfo(f), nil
}

// Update 更新分期方案
func (s *financingService) Update(id uint, req *dto.UpdateFinancingRequest) error {
	fields := map[string]interface{}{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Code != nil {
		fields["code"] = *req.Code
	}
	if req.FinancingType != nil {
		fields["financing_type"] = *req.FinancingType
	}
	if req.MinDownPayment != nil {
		fields["min_down_payment"] = *req.MinDownPayment
	}
	if req.MaxDownPayment != nil {
		fields["max_down_payment"] = *req.MaxDownPayment
	}
	if req.InterestRate != nil {
		fields["interest_rate"] = *req.InterestRate
	}
	if req.AnnualRate != nil {
		fields["annual_rate"] = *req.AnnualRate
	}
	if req.MinPeriods != nil {
		fields["min_periods"] = *req.MinPeriods
	}
	if req.MaxPeriods != nil {
		fields["max_periods"] = *req.MaxPeriods
	}
	if req.MaxAmount != nil {
		fields["max_amount"] = *req.MaxAmount
	}
	if req.Provider != nil {
		fields["provider"] = *req.Provider
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

// Delete 删除分期方案
func (s *financingService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *financingService) GetByID(id uint) (*dto.FinancingInfo, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFinancingNotFound
		}
		return nil, err
	}
	return toFinancingInfo(f), nil
}

// List 列表查询
func (s *financingService) List(req *dto.FinancingListRequest) (*utils.Pagination, []dto.FinancingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.FinancingListQuery{
		FinancingType: req.FinancingType,
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
	result := make([]dto.FinancingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFinancingInfo(&list[i]))
	}
	return pagination, result, nil
}

// UpdateStatus 更新状态
func (s *financingService) UpdateStatus(id uint, status int) error {
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== C 端 =====

// ListPublished C 端列表（仅已发布）
func (s *financingService) ListPublished(page, pageSize int) (*utils.Pagination, []dto.FinancingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListPublished(pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.FinancingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFinancingInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListHot 热门分期方案
func (s *financingService) ListHot(limit int) ([]dto.FinancingInfo, error) {
	list, err := s.repo.ListHot(limit)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FinancingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toFinancingInfo(&list[i]))
	}
	return result, nil
}

// Calculate 分期计算（等额本息）
// 月供 = 贷款本金 × 月利率 × (1+月利率)^期数 / ((1+月利率)^期数 - 1)
func (s *financingService) Calculate(req *dto.FinancingCalculateRequest) (*dto.FinancingCalculateResponse, error) {
	f, err := s.repo.FindByID(req.FinancingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFinancingNotFound
		}
		return nil, err
	}

	if f.Status != model.FinancingStatusPublished {
		return nil, ErrFinancingStatusInvalid
	}

	carPrice := req.CarPrice
	downPayment := req.DownPayment
	periods := req.Periods

	// 校验首付比例
	downPaymentRate := downPayment / carPrice * 100.0
	if downPaymentRate < f.MinDownPayment {
		return nil, ErrFinancingParamsInvalid
	}
	if downPaymentRate > f.MaxDownPayment {
		return nil, ErrFinancingParamsInvalid
	}

	// 校验期数
	if periods < f.MinPeriods || periods > f.MaxPeriods {
		return nil, ErrFinancingParamsInvalid
	}

	// 校验贷款额上限
	loanAmount := carPrice - downPayment
	if f.MaxAmount > 0 && loanAmount > f.MaxAmount {
		return nil, ErrFinancingParamsInvalid
	}

	// 月利率
	monthlyRate := f.InterestRate
	if monthlyRate <= 0 && f.AnnualRate > 0 {
		monthlyRate = f.AnnualRate / 12.0
	}

	var monthlyPayment, totalPayment, totalInterest float64
	if monthlyRate == 0 {
		// 无息贷款
		monthlyPayment = loanAmount / float64(periods)
		totalPayment = loanAmount
		totalInterest = 0
	} else {
		// 等额本息公式
		pow := math.Pow(1+monthlyRate, float64(periods))
		monthlyPayment = loanAmount * monthlyRate * pow / (pow - 1)
		totalPayment = monthlyPayment * float64(periods)
		totalInterest = totalPayment - loanAmount
	}

	// 增加使用次数
	_ = s.repo.IncrUseCount(req.FinancingID, 1)

	return &dto.FinancingCalculateResponse{
		CarPrice:       carPrice,
		DownPayment:    downPayment,
		LoanAmount:     loanAmount,
		Periods:        periods,
		InterestRate:   f.InterestRate,
		AnnualRate:     f.AnnualRate,
		MonthlyPayment: monthlyPayment,
		TotalPayment:   totalPayment,
		TotalInterest:  totalInterest,
	}, nil
}
