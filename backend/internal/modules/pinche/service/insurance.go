// Package service 同城拼车出行业务逻辑层 - 顺风车保险
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrInsuranceNotFound      = errors.New("保险不存在")
	ErrInsuranceStatusInvalid = errors.New("保险状态不允许此操作")
	ErrInsurancePolicyNoUsed  = errors.New("保单号已存在")
)

// InsuranceService 顺风车保险业务接口
type InsuranceService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateInsuranceRequest) (*dto.InsuranceInfo, error)
	Claim(id uint, req *dto.InsuranceClaimRequest) (*dto.InsuranceInfo, error)
	GetByID(id uint) (*dto.InsuranceInfo, error)
	GetByPolicyNo(policyNo string) (*dto.InsuranceInfo, error)
	ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error)
	ListByBooking(bookingID uint, page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error)
	Quote(regionID uint, req *dto.InsuranceQuoteRequest) (*dto.InsuranceInfo, error)

	// M 端
	AdminList(req *dto.InsuranceListRequest) (*utils.Pagination, []dto.InsuranceInfo, error)
	UpdateStatus(id uint, status int) error
}

type insuranceService struct {
	repo repository.InsuranceRepository
}

// NewInsuranceService 创建保险 service 实例
func NewInsuranceService(repo repository.InsuranceRepository) InsuranceService {
	return &insuranceService{repo: repo}
}

// insuranceStatusText 状态文本
func insuranceStatusText(status int) string {
	switch status {
	case model.InsuranceStatusPending:
		return "待生效"
	case model.InsuranceStatusActive:
		return "生效中"
	case model.InsuranceStatusExpired:
		return "已结束"
	case model.InsuranceStatusClaimed:
		return "已理赔"
	}
	return ""
}

// insuranceTypeText 类型文本
func insuranceTypeText(t string) string {
	switch t {
	case model.InsuranceTypePassenger:
		return "乘客险"
	case model.InsuranceTypeDriver:
		return "司机险"
	case model.InsuranceTypeBoth:
		return "双重险"
	}
	return ""
}

// toInsuranceInfo model -> dto
func toInsuranceInfo(i *model.PincheInsurance) *dto.InsuranceInfo {
	info := &dto.InsuranceInfo{
		ID:                i.ID,
		RegionID:          i.RegionID,
		PincheID:          i.PincheID,
		BookingID:         i.BookingID,
		PolicyNo:          i.PolicyNo,
		InsuranceCompany:  i.InsuranceCompany,
		InsuranceType:     i.InsuranceType,
		InsuranceTypeText: insuranceTypeText(i.InsuranceType),
		CoverageAmount:    i.CoverageAmount,
		Premium:           i.Premium,
		InsuredName:       i.InsuredName,
		InsuredIDCard:     i.InsuredIDCard,
		StartTime:         i.StartTime,
		EndTime:           i.EndTime,
		Status:            i.Status,
		StatusText:        insuranceStatusText(i.Status),
		ClaimAmount:       i.ClaimAmount,
		ClaimReason:       i.ClaimReason,
		ClaimedAt:         i.ClaimedAt,
		CreatedAt:         i.CreatedAt,
	}
	if i.Beneficiaries != nil {
		info.Beneficiaries = i.Beneficiaries
	}
	return info
}

// genPolicyNo 生成保单号 PC + yyyyMMdd + 16位hex
func genPolicyNo() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("PC%s%s", time.Now().Format("20060102"), hex.EncodeToString(b))
}

// Create 创建保险
func (s *insuranceService) Create(regionID uint, userID uint, req *dto.CreateInsuranceRequest) (*dto.InsuranceInfo, error) {
	policyNo := genPolicyNo()
	// 校验保单号唯一
	if existing, err := s.repo.FindByPolicyNo(policyNo); err == nil && existing != nil {
		return nil, ErrInsurancePolicyNoUsed
	}
	insType := req.InsuranceType
	if insType == "" {
		insType = model.InsuranceTypePassenger
	}
	ins := &model.PincheInsurance{
		PincheID:         req.PincheID,
		BookingID:        req.BookingID,
		PolicyNo:         policyNo,
		InsuranceCompany: req.InsuranceCompany,
		InsuranceType:    insType,
		CoverageAmount:   req.CoverageAmount,
		Premium:          req.Premium,
		InsuredName:      req.InsuredName,
		InsuredIDCard:    req.InsuredIDCard,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		Status:           model.InsuranceStatusPending,
	}
	ins.RegionID = regionID
	_ = userID // 通过 regionID 隔离，userID 不直接落库
	if req.Beneficiaries != nil {
		if jb, err := model.FromJSON(req.Beneficiaries); err == nil {
			ins.Beneficiaries = jb
		}
	}
	// 默认起止时间：现在起 24 小时
	now := time.Now()
	if ins.StartTime == nil {
		ins.StartTime = &now
	}
	if ins.EndTime == nil {
		end := now.Add(24 * time.Hour)
		ins.EndTime = &end
	}
	// 起止时间合法则直接生效
	if ins.StartTime.Before(now) || ins.StartTime.Equal(now) {
		ins.Status = model.InsuranceStatusActive
	}
	if err := s.repo.Create(ins); err != nil {
		return nil, err
	}
	return toInsuranceInfo(ins), nil
}

// Claim 保险理赔
func (s *insuranceService) Claim(id uint, req *dto.InsuranceClaimRequest) (*dto.InsuranceInfo, error) {
	ins, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInsuranceNotFound
		}
		return nil, err
	}
	if ins.Status != model.InsuranceStatusActive && ins.Status != model.InsuranceStatusExpired {
		return nil, ErrInsuranceStatusInvalid
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":        model.InsuranceStatusClaimed,
		"claim_amount":  req.ClaimAmount,
		"claim_reason":  req.ClaimReason,
		"claimed_at":    &now,
	}
	if err := s.repo.Update(id, fields); err != nil {
		return nil, err
	}
	ins.Status = model.InsuranceStatusClaimed
	ins.ClaimAmount = req.ClaimAmount
	ins.ClaimReason = req.ClaimReason
	ins.ClaimedAt = &now
	return toInsuranceInfo(ins), nil
}

// GetByID 获取详情
func (s *insuranceService) GetByID(id uint) (*dto.InsuranceInfo, error) {
	ins, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInsuranceNotFound
		}
		return nil, err
	}
	return toInsuranceInfo(ins), nil
}

// GetByPolicyNo 按保单号查询
func (s *insuranceService) GetByPolicyNo(policyNo string) (*dto.InsuranceInfo, error) {
	ins, err := s.repo.FindByPolicyNo(policyNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInsuranceNotFound
		}
		return nil, err
	}
	return toInsuranceInfo(ins), nil
}

// ListByPinche 按行程查询
func (s *insuranceService) ListByPinche(pincheID uint, page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPinche(pincheID, pagination)
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

// ListByBooking 按预订查询
func (s *insuranceService) ListByBooking(bookingID uint, page, pageSize int) (*utils.Pagination, []dto.InsuranceInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByBooking(bookingID, pagination)
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

// Quote 保险报价（不落库，仅返回报价信息）
func (s *insuranceService) Quote(regionID uint, req *dto.InsuranceQuoteRequest) (*dto.InsuranceInfo, error) {
	insType := req.InsuranceType
	if insType == "" {
		insType = model.InsuranceTypePassenger
	}
	// 简单报价：固定保额 50 万 / 保费 5 元
	coverage := 500000.0
	premium := 5.0
	if insType == model.InsuranceTypeBoth {
		coverage = 1000000.0
		premium = 8.0
	}
	info := &dto.InsuranceInfo{
		PincheID:          req.PincheID,
		InsuranceType:     insType,
		InsuranceTypeText: insuranceTypeText(insType),
		CoverageAmount:    coverage,
		Premium:           premium,
		Status:            model.InsuranceStatusPending,
		StatusText:        insuranceStatusText(model.InsuranceStatusPending),
	}
	_ = regionID
	return info, nil
}

// AdminList 管理后台保险列表
func (s *insuranceService) AdminList(req *dto.InsuranceListRequest) (*utils.Pagination, []dto.InsuranceInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.InsuranceListOptions{
		PincheID:  req.PincheID,
		BookingID: req.BookingID,
		Status:    req.Status,
		PolicyNo:  req.PolicyNo,
	}
	// 跨地区：regionID=0
	list, total, err := s.repo.List(0, pagination, opts)
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

// UpdateStatus 管理后台更新状态
func (s *insuranceService) UpdateStatus(id uint, status int) error {
	return s.repo.UpdateStatus(id, status)
}
