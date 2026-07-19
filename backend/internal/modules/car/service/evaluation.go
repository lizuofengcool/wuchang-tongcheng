// Package service 同城车辆买卖业务逻辑层 - 车辆评估
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标瓜子/人人车 估值+折旧+相似成交
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrEvaluationNotFound      = errors.New("评估单不存在")
	ErrEvaluationNoPermission  = errors.New("无权操作此评估单")
	ErrEvaluationStatusInvalid = errors.New("评估单状态不允许此操作")
)

// EvaluationService 车辆评估业务接口
type EvaluationService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateEvaluationRequest) (*dto.EvaluationInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateEvaluationRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.EvaluationInfo, error)
	GetByCarID(carID uint) (*dto.EvaluationInfo, error)
	List(regionID uint, req *dto.EvaluationListRequest) (*utils.Pagination, []dto.EvaluationInfo, error)
	ListByEvaluator(evaluatorID uint, page, pageSize int) (*utils.Pagination, []dto.EvaluationInfo, error)
	ListByCarID(carID uint) ([]dto.EvaluationInfo, error)

	// C 端在线评估
	OnlineEvaluate(regionID uint, userID uint, req *dto.OnlineEvaluationRequest) (*dto.OnlineEvaluationResponse, error)

	// M 端管理
	AdminList(req *dto.EvaluationListRequest) (*utils.Pagination, []dto.EvaluationInfo, error)
	AdminGetByID(id uint) (*dto.EvaluationInfo, error)
	UpdateStatus(id uint, status int) error
}

type evaluationService struct {
	repo repository.EvaluationRepository
}

// NewEvaluationService 创建车辆评估 service 实例
func NewEvaluationService(repo repository.EvaluationRepository) EvaluationService {
	return &evaluationService{repo: repo}
}

// evaluationStatusText 评估状态文本
func evaluationStatusText(status int) string {
	switch status {
	case model.EvaluationStatusPending:
		return "待评估"
	case model.EvaluationStatusCompleted:
		return "已完成"
	case model.EvaluationStatusExpired:
		return "已过期"
	case model.EvaluationStatusCanceled:
		return "已取消"
	}
	return ""
}

// toEvaluationInfo model -> dto
func toEvaluationInfo(e *model.CarEvaluation) *dto.EvaluationInfo {
	info := &dto.EvaluationInfo{
		ID:                 e.ID,
		EvaluationNo:       e.EvaluationNo,
		CarID:              e.CarID,
		ModelID:            e.ModelID,
		EvaluatorID:        e.EvaluatorID,
		EvaluatorName:      e.EvaluatorName,
		EvaluatorLevel:     e.EvaluatorLevel,
		EvaluationType:     e.EvaluationType,
		MarketPrice:        e.MarketPrice,
		TradeInPrice:       e.TradeInPrice,
		RetailPrice:        e.RetailPrice,
		WholesalePrice:     e.WholesalePrice,
		DepreciationAmount: e.DepreciationAmount,
		DepreciationRate:   e.DepreciationRate,
		FinalPrice:         e.FinalPrice,
		PriceRangeMin:      e.PriceRangeMin,
		PriceRangeMax:      e.PriceRangeMax,
		Description:        e.Description,
		ReportURL:          e.ReportURL,
		ValidUntil:         e.ValidUntil,
		Status:             e.Status,
		StatusText:         evaluationStatusText(e.Status),
		RegionID:           e.RegionID,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
	if e.Factors != nil {
		info.Factors = e.Factors
	}
	if e.SimilarDeals != nil {
		info.SimilarDeals = e.SimilarDeals
	}
	return info
}

// genEvaluationNo 生成评估单号：EV + yyyyMMddHHmmss + 6 位随机
func genEvaluationNo() string {
	return fmt.Sprintf("EV%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建评估单
func (s *evaluationService) Create(regionID uint, userID uint, req *dto.CreateEvaluationRequest) (*dto.EvaluationInfo, error) {
	e := &model.CarEvaluation{
		EvaluationNo:       genEvaluationNo(),
		CarID:              req.CarID,
		ModelID:            req.ModelID,
		EvaluatorID:        req.EvaluatorID,
		EvaluatorName:      req.EvaluatorName,
		EvaluatorLevel:     req.EvaluatorLevel,
		EvaluationType:     req.EvaluationType,
		MarketPrice:        req.MarketPrice,
		TradeInPrice:       req.TradeInPrice,
		RetailPrice:        req.RetailPrice,
		WholesalePrice:     req.WholesalePrice,
		DepreciationAmount: req.DepreciationAmount,
		DepreciationRate:   req.DepreciationRate,
		FinalPrice:         req.FinalPrice,
		PriceRangeMin:      req.PriceRangeMin,
		PriceRangeMax:      req.PriceRangeMax,
		Description:        req.Description,
		ReportURL:          req.ReportURL,
		ValidUntil:         req.ValidUntil,
		Status:             model.EvaluationStatusPending,
	}
	e.RegionID = regionID

	// 默认值兜底
	if e.EvaluatorLevel == "" {
		e.EvaluatorLevel = model.EvaluatorLevelJunior
	}
	if e.EvaluationType == "" {
		e.EvaluationType = model.EvaluationTypeOnline
	}

	// JSONB 字段
	if req.Factors != nil {
		if jb, err := model.FromJSON(req.Factors); err == nil {
			e.Factors = jb
		}
	}
	if req.SimilarDeals != nil {
		if jb, err := model.FromJSON(req.SimilarDeals); err == nil {
			e.SimilarDeals = jb
		}
	}

	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return toEvaluationInfo(e), nil
}

// Update 更新评估单
func (s *evaluationService) Update(id uint, operatorID uint, req *dto.UpdateEvaluationRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEvaluationNotFound
		}
		return err
	}
	if e.EvaluatorID != operatorID && operatorID != 0 {
		return ErrEvaluationNoPermission
	}

	fields := map[string]interface{}{}
	if req.EvaluatorID != nil {
		fields["evaluator_id"] = *req.EvaluatorID
	}
	if req.EvaluatorName != nil {
		fields["evaluator_name"] = *req.EvaluatorName
	}
	if req.EvaluatorLevel != nil {
		fields["evaluator_level"] = *req.EvaluatorLevel
	}
	if req.EvaluationType != nil {
		fields["evaluation_type"] = *req.EvaluationType
	}
	if req.MarketPrice != nil {
		fields["market_price"] = *req.MarketPrice
	}
	if req.TradeInPrice != nil {
		fields["trade_in_price"] = *req.TradeInPrice
	}
	if req.RetailPrice != nil {
		fields["retail_price"] = *req.RetailPrice
	}
	if req.WholesalePrice != nil {
		fields["wholesale_price"] = *req.WholesalePrice
	}
	if req.DepreciationAmount != nil {
		fields["depreciation_amount"] = *req.DepreciationAmount
	}
	if req.DepreciationRate != nil {
		fields["depreciation_rate"] = *req.DepreciationRate
	}
	if req.FinalPrice != nil {
		fields["final_price"] = *req.FinalPrice
	}
	if req.PriceRangeMin != nil {
		fields["price_range_min"] = *req.PriceRangeMin
	}
	if req.PriceRangeMax != nil {
		fields["price_range_max"] = *req.PriceRangeMax
	}
	if req.Factors != nil {
		if jb, err := model.FromJSON(req.Factors); err == nil {
			fields["factors"] = jb
		}
	}
	if req.SimilarDeals != nil {
		if jb, err := model.FromJSON(req.SimilarDeals); err == nil {
			fields["similar_deals"] = jb
		}
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.ReportURL != nil {
		fields["report_url"] = *req.ReportURL
	}
	if req.ValidUntil != nil {
		fields["valid_until"] = req.ValidUntil
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除评估单
func (s *evaluationService) Delete(id uint, operatorID uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEvaluationNotFound
		}
		return err
	}
	if e.EvaluatorID != operatorID && operatorID != 0 {
		return ErrEvaluationNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *evaluationService) GetByID(id uint) (*dto.EvaluationInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}
	return toEvaluationInfo(e), nil
}

// GetByCarID 按车源查询最新评估
func (s *evaluationService) GetByCarID(carID uint) (*dto.EvaluationInfo, error) {
	e, err := s.repo.FindByCarID(carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}
	return toEvaluationInfo(e), nil
}

// List C 端列表查询
func (s *evaluationService) List(regionID uint, req *dto.EvaluationListRequest) (*utils.Pagination, []dto.EvaluationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EvaluationListOptions{
		CarID:          req.CarID,
		ModelID:        req.ModelID,
		EvaluatorID:    req.EvaluatorID,
		EvaluationType: req.EvaluationType,
		Status:         req.Status,
		Keyword:        req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EvaluationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEvaluationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEvaluator 评估师自己的评估单
func (s *evaluationService) ListByEvaluator(evaluatorID uint, page, pageSize int) (*utils.Pagination, []dto.EvaluationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEvaluator(evaluatorID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EvaluationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEvaluationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByCarID 车辆历史评估
func (s *evaluationService) ListByCarID(carID uint) ([]dto.EvaluationInfo, error) {
	list, err := s.repo.ListByCarID(carID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.EvaluationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEvaluationInfo(&list[i]))
	}
	return result, nil
}

// OnlineEvaluate 在线评估（C 端即时估值，不落库或落库为 ai 类型）
// MVP 实现：基于折旧率计算 + 车型均价参考
func (s *evaluationService) OnlineEvaluate(regionID uint, userID uint, req *dto.OnlineEvaluationRequest) (*dto.OnlineEvaluationResponse, error) {
	// 基础折旧率：默认 15%/年
	depreciationRate := 15.0
	// 里程折旧：每 1 万公里折旧 5%
	mileageDepreciation := req.Mileage / 10000.0 * 5.0
	// 车龄折旧
	age := time.Now().Year() - req.Year
	if age < 0 {
		age = 0
	}
	ageDepreciation := float64(age) * depreciationRate

	// 车况等级调整
	conditionAdjust := 0.0
	switch req.ConditionLevel {
	case model.InspectionConditionA:
		conditionAdjust = 5.0
	case model.InspectionConditionB:
		conditionAdjust = 0.0
	case model.InspectionConditionC:
		conditionAdjust = -10.0
	case model.InspectionConditionD:
		conditionAdjust = -20.0
	}

	// 事故调整
	accidentAdjust := float64(req.AccidentCount) * -5.0
	// 过户次数调整
	transferAdjust := float64(req.TransferCount) * -2.0

	totalDepreciation := ageDepreciation + mileageDepreciation + conditionAdjust + accidentAdjust + transferAdjust
	if totalDepreciation > 80 {
		totalDepreciation = 80
	}
	if totalDepreciation < 10 {
		totalDepreciation = 10
	}

	// 市场参考价：从评估表查同车型均价，否则使用经验值
	marketPrice, _ := s.repo.AvgPriceByModelID(0)
	if marketPrice <= 0 {
		// 经验值兜底：按排量×10000 + 年份调整
		marketPrice = req.Displacement * 10000
		if marketPrice <= 0 {
			marketPrice = 100000
		}
	}

	// 最终估值 = 市场价 × (1 - 总折旧率/100)
	finalPrice := marketPrice * (1.0 - totalDepreciation/100.0)
	if finalPrice < 0 {
		finalPrice = 0
	}

	// 价格区间：±5%
	priceRangeMin := finalPrice * 0.95
	priceRangeMax := finalPrice * 1.05

	// 收购价：估值的 90%（车商利润空间）
	tradeInPrice := finalPrice * 0.9
	// 零售价：估值
	retailPrice := finalPrice

	reason := fmt.Sprintf("基于市场均价 %.2f 元，考虑 %d 年车龄、%.0f 万公里里程、车况 %s、%d 次事故、%d 次过户，综合折旧率 %.2f%%",
		marketPrice, age, req.Mileage/10000, req.ConditionLevel, req.AccidentCount, req.TransferCount, totalDepreciation)

	return &dto.OnlineEvaluationResponse{
		MarketPrice:      marketPrice,
		TradeInPrice:     tradeInPrice,
		RetailPrice:      retailPrice,
		PriceRangeMin:    priceRangeMin,
		PriceRangeMax:    priceRangeMax,
		DepreciationRate: totalDepreciation,
		FinalPrice:       finalPrice,
		Reason:           reason,
	}, nil
}

// ===== M 端管理 =====

func (s *evaluationService) AdminList(req *dto.EvaluationListRequest) (*utils.Pagination, []dto.EvaluationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EvaluationAdminListOptions{
		CarID:          req.CarID,
		EvaluatorID:    req.EvaluatorID,
		EvaluationType: req.EvaluationType,
		Status:         req.Status,
		Keyword:        req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EvaluationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEvaluationInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *evaluationService) AdminGetByID(id uint) (*dto.EvaluationInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEvaluationNotFound
		}
		return nil, err
	}
	return toEvaluationInfo(e), nil
}

// UpdateStatus M 端强制更新状态
func (s *evaluationService) UpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{
		"status": status,
	}
	return s.repo.Update(id, fields)
}
