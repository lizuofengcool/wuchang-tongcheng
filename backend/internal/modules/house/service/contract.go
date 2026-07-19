// Package service 合同电子化业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package service

import (
	"errors"
	"fmt"
	"time"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrContractNotFound     = errors.New("合同不存在")
	ErrContractNoPermission = errors.New("无权操作此合同")
	ErrContractSigned       = errors.New("合同已签署，不可修改")
	ErrContractStatus       = errors.New("合同状态不允许此操作")
)

// ContractService 合同业务接口
type ContractService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.ContractCreateRequest) (*dto.ContractResponse, error)
	Update(id uint, operatorID uint, req *dto.ContractUpdateRequest) error
	GetByID(id uint, userID uint) (*dto.ContractResponse, error)
	List(regionID uint, req *dto.ContractListQuery) (*utils.Pagination, []dto.ContractResponse, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ContractResponse, error)
	Sign(id uint, userID uint, party string) error
	Terminate(id uint, userID uint, req *dto.ContractTerminateRequest) error

	// M 端
	AdminList(req *dto.ContractAdminListQuery) (*utils.Pagination, []dto.ContractResponse, error)
}

type contractService struct {
	repo repository.ContractRepository
}

// NewContractService 创建 service 实例
func NewContractService(repo repository.ContractRepository) ContractService {
	return &contractService{repo: repo}
}

// toContractInfo model -> dto
func toContractInfo(c *model.HouseContract) *dto.ContractResponse {
	return &dto.ContractResponse{
		ID:                c.ID,
		ContractNo:        c.ContractNo,
		ContractType:      c.ContractType,
		HouseID:           c.HouseID,
		ListingID:         c.ListingID,
		CommunityID:       c.CommunityID,
		PartyAID:          c.PartyAID,
		PartyAName:        c.PartyAName,
		PartyAPhone:       c.PartyAPhone,
		PartyAIDCard:      c.PartyAIDCard,
		PartyBID:          c.PartyBID,
		PartyBName:        c.PartyBName,
		PartyBPhone:       c.PartyBPhone,
		PartyBIDCard:      c.PartyBIDCard,
		AgentID:           c.AgentID,
		AgentName:         c.AgentName,
		Title:             c.Title,
		Content:           c.Content,
		Amount:            c.Amount,
		Deposit:           c.Deposit,
		Commission:        c.Commission,
		CommissionPayer:   c.CommissionPayer,
		StartDate:         c.StartDate,
		EndDate:           c.EndDate,
		PaymentMethod:     c.PaymentMethod,
		SignMethod:        c.SignMethod,
		PartyASignedAt:    c.PartyASignedAt,
		PartyBSignedAt:    c.PartyBSignedAt,
		AgentSignedAt:     c.AgentSignedAt,
		Status:            c.Status,
		StatusText:        contractStatusText(c.Status),
		EffectiveAt:       c.EffectiveAt,
		TerminatedAt:      c.TerminatedAt,
		TerminatedReason:  c.TerminatedReason,
		ArchivedAt:        c.ArchivedAt,
		RegionID:          c.RegionID,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

func contractStatusText(s int) string {
	switch s {
	case model.ContractStatusDraft:
		return "草稿"
	case model.ContractStatusPendingSign:
		return "待签署"
	case model.ContractStatusPartSigned:
		return "部分签署"
	case model.ContractStatusSigned:
		return "已签署"
	case model.ContractStatusEffective:
		return "已生效"
	case model.ContractStatusTerminated:
		return "已终止"
	case model.ContractStatusArchived:
		return "已归档"
	case model.ContractStatusCanceled:
		return "已取消"
	}
	return "草稿"
}

// generateContractNo 生成合同编号
func generateContractNo() string {
	return fmt.Sprintf("HT%s", time.Now().Format("20060102150405.000"))
}

// ===== C 端 =====

func (s *contractService) Create(regionID uint, userID uint, req *dto.ContractCreateRequest) (*dto.ContractResponse, error) {
	c := &model.HouseContract{
		ContractNo:      generateContractNo(),
		ContractType:    req.ContractType,
		HouseID:         req.HouseID,
		ListingID:       req.ListingID,
		CommunityID:     req.CommunityID,
		PartyAID:        req.PartyAID,
		PartyAName:      req.PartyAName,
		PartyAPhone:     req.PartyAPhone,
		PartyAIDCard:    req.PartyAIDCard,
		PartyBID:        req.PartyBID,
		PartyBName:      req.PartyBName,
		PartyBPhone:     req.PartyBPhone,
		PartyBIDCard:    req.PartyBIDCard,
		AgentID:         req.AgentID,
		AgentName:       req.AgentName,
		Title:           req.Title,
		Content:         req.Content,
		Amount:          req.Amount,
		Deposit:         req.Deposit,
		Commission:      req.Commission,
		CommissionPayer: req.CommissionPayer,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		PaymentMethod:   req.PaymentMethod,
		SignMethod:      req.SignMethod,
		Status:          req.Status,
	}
	c.RegionID = regionID

	if c.ContractType == "" {
		c.ContractType = model.ContractTypeRent
	}
	if c.PaymentMethod == "" {
		c.PaymentMethod = model.PaymentMethodMonthly
	}
	if c.SignMethod == "" {
		c.SignMethod = model.SignMethodOnline
	}
	if c.CommissionPayer == "" {
		c.CommissionPayer = model.CommissionPayerBoth
	}
	if c.Status == model.ContractStatusPendingSign {
		// 直接进入待签状态
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toContractInfo(c), nil
}

func (s *contractService) Update(id uint, operatorID uint, req *dto.ContractUpdateRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}
	// 仅甲方或经纪人可修改，且状态必须为草稿或待签
	if c.PartyAID != operatorID && c.AgentID != operatorID {
		return ErrContractNoPermission
	}
	if c.Status >= model.ContractStatusSigned {
		return ErrContractSigned
	}

	fields := map[string]interface{}{
		"title":            req.Title,
		"content":          req.Content,
		"amount":           req.Amount,
		"deposit":          req.Deposit,
		"commission":       req.Commission,
		"commission_payer": req.CommissionPayer,
		"start_date":       req.StartDate,
		"end_date":         req.EndDate,
		"payment_method":   req.PaymentMethod,
		"sign_method":      req.SignMethod,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *contractService) GetByID(id uint, userID uint) (*dto.ContractResponse, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	// 用户必须是甲方/乙方/经纪人/管理员
	if userID > 0 && c.PartyAID != userID && c.PartyBID != userID && c.AgentID != userID {
		return nil, ErrContractNoPermission
	}
	return toContractInfo(c), nil
}

func (s *contractService) List(regionID uint, req *dto.ContractListQuery) (*utils.Pagination, []dto.ContractResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractListOptions{
		HouseID:      req.HouseID,
		ListingID:    req.ListingID,
		PartyAID:     req.PartyAID,
		PartyBID:     req.PartyBID,
		AgentID:      req.AgentID,
		ContractType: req.ContractType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ContractResponse, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *contractService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ContractResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByParty(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ContractResponse, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

// Sign 签署合同
func (s *contractService) Sign(id uint, userID uint, party string) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{}
	switch party {
	case "a":
		if c.PartyAID != userID {
			return ErrContractNoPermission
		}
		fields["party_a_signed_at"] = &now
	case "b":
		if c.PartyBID != userID {
			return ErrContractNoPermission
		}
		fields["party_b_signed_at"] = &now
	case "agent":
		if c.AgentID != userID {
			return ErrContractNoPermission
		}
		fields["agent_signed_at"] = &now
	default:
		return ErrContractStatus
	}

	// 状态机推进：待签 -> 部分签 -> 已签 -> 已生效
	newStatus := c.Status
	if c.Status == model.ContractStatusPendingSign || c.Status == model.ContractStatusPartSigned {
		// 判断是否三方都已签
		aSigned := c.PartyASignedAt != nil || party == "a"
		bSigned := c.PartyBSignedAt != nil || party == "b"
		agentSigned := (c.AgentID == 0) || c.AgentSignedAt != nil || party == "agent"
		if aSigned && bSigned && agentSigned {
			newStatus = model.ContractStatusSigned
			fields["effective_at"] = &now
		} else {
			newStatus = model.ContractStatusPartSigned
		}
		fields["status"] = newStatus
	}

	return s.repo.UpdateFields(id, fields)
}

// Terminate 终止合同
func (s *contractService) Terminate(id uint, userID uint, req *dto.ContractTerminateRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}
	if c.PartyAID != userID && c.PartyBID != userID && c.AgentID != userID {
		return ErrContractNoPermission
	}
	if c.Status != model.ContractStatusEffective && c.Status != model.ContractStatusSigned {
		return ErrContractStatus
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"status":            model.ContractStatusTerminated,
		"terminated_at":     &now,
		"terminated_reason": req.Reason,
	})
}

// ===== M 端 =====

func (s *contractService) AdminList(req *dto.ContractAdminListQuery) (*utils.Pagination, []dto.ContractResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractAdminListOptions{
		RegionID: req.RegionID,
		HouseID:  req.HouseID,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ContractResponse, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}
