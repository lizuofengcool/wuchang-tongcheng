// Package service 同城零工兼职业务逻辑层 - 电子合同
// 对标青团兼职合同电子化/法大大/e签宝：在线签署 + 模板 + 多方签名
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrContractNotFound     = errors.New("合同不存在")
	ErrContractNoPermission = errors.New("无权操作此合同")
	ErrContractStatusInvalid = errors.New("合同状态不允许此操作")
	ErrContractSignError    = errors.New("合同签署错误")
)

// ContractService 电子合同业务接口
type ContractService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateContractRequest) (*dto.ContractInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateContractRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.ContractInfo, error)
	GetByContractNo(no string) (*dto.ContractInfo, error)
	List(regionID uint, req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error)
	ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error)
	ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error)
	ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error)

	// 签署/状态变更
	Sign(id uint, userID uint, req *dto.ContractSignRequest) error
	UpdateStatus(id uint, req *dto.ContractStatusUpdateRequest) error

	// M 端管理
	AdminList(req *dto.ContractAdminListRequest) (*utils.Pagination, []dto.ContractInfo, error)
}

type contractService struct {
	repo repository.ContractRepository
}

// NewContractService 创建电子合同 service 实例
func NewContractService(repo repository.ContractRepository) ContractService {
	return &contractService{repo: repo}
}

// contractTypeText 合同类型文本
func contractTypeText(t string) string {
	switch t {
	case model.ContractTypePartTime:
		return "兼职合同"
	case model.ContractTypeTemp:
		return "临时合同"
	case model.ContractTypeTask:
		return "任务合同"
	case model.ContractTypeService:
		return "服务合同"
	case model.ContractTypeInternship:
		return "实习合同"
	case model.ContractTypeProject:
		return "项目合同"
	case model.ContractTypeOutsourcing:
		return "外包合同"
	}
	return ""
}

// contractStatusText 合同状态文本
func contractStatusText(s int) string {
	switch s {
	case model.ContractStatusDraft:
		return "草稿"
	case model.ContractStatusPending:
		return "待签署"
	case model.ContractStatusSigned:
		return "已签署"
	case model.ContractStatusEffective:
		return "已生效"
	case model.ContractStatusFulfilling:
		return "履行中"
	case model.ContractStatusCompleted:
		return "已完成"
	case model.ContractStatusTerminated:
		return "已终止"
	case model.ContractStatusCanceled:
		return "已取消"
	case model.ContractStatusExpired:
		return "已过期"
	}
	return ""
}

// signMethodText 签署方式文本
func signMethodText(m string) string {
	switch m {
	case model.SignMethodHandwritten:
		return "手写签名"
	case model.SignMethodSMS:
		return "短信验证"
	case model.SignMethodFace:
		return "人脸识别"
	case model.SignMethodCA:
		return "CA 证书"
	}
	return ""
}

// toContractInfo model -> dto
func toContractInfo(c *model.LinggongContract) *dto.ContractInfo {
	return &dto.ContractInfo{
		ID:               c.ID,
		ContractNo:       c.ContractNo,
		ContractType:     c.ContractType,
		ContractTypeText: contractTypeText(c.ContractType),
		LinggongID:       c.LinggongID,
		TaskID:           c.TaskID,
		ApplicationID:    c.ApplicationID,
		EmployerID:       c.EmployerID,
		EmployerName:     c.EmployerName,
		EmployerPhone:    c.EmployerPhone,
		EmployerIDCard:   c.EmployerIDCard,
		EmployerSignURL:  c.EmployerSignURL,
		WorkerID:         c.WorkerID,
		WorkerName:       c.WorkerName,
		WorkerPhone:      c.WorkerPhone,
		WorkerIDCard:     c.WorkerIDCard,
		WorkerSignURL:    c.WorkerSignURL,
		WorkStartDate:    c.WorkStartDate,
		WorkEndDate:      c.WorkEndDate,
		WorkContent:      c.WorkContent,
		WorkPlace:        c.WorkPlace,
		BillingType:      c.BillingType,
		SalaryAmount:     c.SalaryAmount,
		SalaryUnit:       c.SalaryUnit,
		Settlement:       c.Settlement,
		SettlementText:   settlementText(c.Settlement),
		TotalAmount:      c.TotalAmount,
		PaidAmount:       c.PaidAmount,
		Deposit:          c.Deposit,
		PenaltyBreach:    c.PenaltyBreach,
		Confidential:     c.Confidential,
		NonCompete:       c.NonCompete,
		SignMethod:       c.SignMethod,
		SignMethodText:   signMethodText(c.SignMethod),
		ContractURL:      c.ContractURL,
		Attachments:      c.Attachments,
		EmployerSignedAt: c.EmployerSignedAt,
		WorkerSignedAt:   c.WorkerSignedAt,
		SignedAt:         c.SignedAt,
		EffectiveAt:      c.EffectiveAt,
		ExpiredAt:        c.ExpiredAt,
		TerminatedAt:     c.TerminatedAt,
		CompletedAt:      c.CompletedAt,
		Status:           c.Status,
		StatusText:       contractStatusText(c.Status),
		TemplateID:       c.TemplateID,
		Remark:           c.Remark,
		RegionID:         c.RegionID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

// genContractNo 生成合同编号：CT + yyyyMMddHHmmss + 6 位随机
func genContractNo() string {
	return fmt.Sprintf("CT%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建合同
func (s *contractService) Create(regionID uint, userID uint, req *dto.CreateContractRequest) (*dto.ContractInfo, error) {
	c := &model.LinggongContract{
		ContractNo:    genContractNo(),
		ContractType:  req.ContractType,
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		ApplicationID: req.ApplicationID,
		EmployerID:    req.EmployerID,
		EmployerName:  req.EmployerName,
		EmployerPhone: req.EmployerPhone,
		EmployerIDCard: req.EmployerIDCard,
		WorkerID:      req.WorkerID,
		WorkerName:    req.WorkerName,
		WorkerPhone:   req.WorkerPhone,
		WorkerIDCard:  req.WorkerIDCard,
		WorkStartDate: req.WorkStartDate,
		WorkEndDate:   req.WorkEndDate,
		WorkContent:   req.WorkContent,
		WorkPlace:     req.WorkPlace,
		BillingType:   req.BillingType,
		SalaryAmount:  req.SalaryAmount,
		SalaryUnit:    req.SalaryUnit,
		Settlement:    req.Settlement,
		TotalAmount:   req.TotalAmount,
		Deposit:       req.Deposit,
		PenaltyBreach: req.PenaltyBreach,
		Confidential:  req.Confidential,
		NonCompete:    req.NonCompete,
		SignMethod:    req.SignMethod,
		ContractURL:   req.ContractURL,
		TemplateID:    req.TemplateID,
		Remark:        req.Remark,
		Status:        model.ContractStatusDraft,
	}
	c.RegionID = regionID

	// 默认值兜底
	if c.ContractType == "" {
		c.ContractType = model.ContractTypePartTime
	}
	if c.BillingType == "" {
		c.BillingType = model.BillingTypeByDay
	}
	if c.Settlement == "" {
		c.Settlement = model.SettlementT1
	}
	if c.SignMethod == "" {
		c.SignMethod = model.SignMethodHandwritten
	}

	// JSONB 字段
	if req.Attachments != nil {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			c.Attachments = jb
		}
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toContractInfo(c), nil
}

// Update 更新合同（仅创建者/雇主）
func (s *contractService) Update(id uint, operatorID uint, req *dto.UpdateContractRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}
	if c.EmployerID != operatorID && c.WorkerID != operatorID {
		return ErrContractNoPermission
	}
	if c.Status >= model.ContractStatusSigned {
		return ErrContractStatusInvalid
	}

	fields := map[string]interface{}{}
	if req.WorkContent != nil {
		fields["work_content"] = *req.WorkContent
	}
	if req.WorkPlace != nil {
		fields["work_place"] = *req.WorkPlace
	}
	if req.WorkStartDate != nil {
		fields["work_start_date"] = *req.WorkStartDate
	}
	if req.WorkEndDate != nil {
		fields["work_end_date"] = *req.WorkEndDate
	}
	if req.BillingType != nil {
		fields["billing_type"] = *req.BillingType
	}
	if req.SalaryAmount != nil {
		fields["salary_amount"] = *req.SalaryAmount
	}
	if req.SalaryUnit != nil {
		fields["salary_unit"] = *req.SalaryUnit
	}
	if req.Settlement != nil {
		fields["settlement"] = *req.Settlement
	}
	if req.TotalAmount != nil {
		fields["total_amount"] = *req.TotalAmount
	}
	if req.Deposit != nil {
		fields["deposit"] = *req.Deposit
	}
	if req.PenaltyBreach != nil {
		fields["penalty_breach"] = *req.PenaltyBreach
	}
	if req.Confidential != nil {
		fields["confidential"] = *req.Confidential
	}
	if req.NonCompete != nil {
		fields["non_compete"] = *req.NonCompete
	}
	if req.ContractURL != nil {
		fields["contract_url"] = *req.ContractURL
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}
	if req.Attachments != nil {
		if jb, err := model.FromJSON(req.Attachments); err == nil {
			fields["attachments"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除合同（仅草稿状态可删）
func (s *contractService) Delete(id uint, operatorID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}
	if c.EmployerID != operatorID {
		return ErrContractNoPermission
	}
	if c.Status != model.ContractStatusDraft {
		return ErrContractStatusInvalid
	}
	return s.repo.Delete(id)
}

// GetByID 获取合同详情
func (s *contractService) GetByID(id uint) (*dto.ContractInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	return toContractInfo(c), nil
}

// GetByContractNo 按合同编号查询
func (s *contractService) GetByContractNo(no string) (*dto.ContractInfo, error) {
	c, err := s.repo.FindByContractNo(no)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, err
	}
	return toContractInfo(c), nil
}

// List C 端列表查询
func (s *contractService) List(regionID uint, req *dto.ContractListRequest) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractListOptions{
		LinggongID:    req.LinggongID,
		TaskID:        req.TaskID,
		ApplicationID: req.ApplicationID,
		EmployerID:    req.EmployerID,
		WorkerID:      req.WorkerID,
		ContractType:  req.ContractType,
		Status:        req.Status,
		Keyword:       req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByLinggong 按岗位反查
func (s *contractService) ListByLinggong(linggongID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByLinggong(linggongID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByEmployer 按雇主反查
func (s *contractService) ListByEmployer(employerID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByEmployer(employerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByWorker 按求职者反查
func (s *contractService) ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByWorker(workerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}

// Sign 合同签署
func (s *contractService) Sign(id uint, userID uint, req *dto.ContractSignRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{}

	if req.SignMethod != "" {
		fields["sign_method"] = req.SignMethod
	}

	switch req.SignRole {
	case "employer":
		if c.EmployerID != userID {
			return ErrContractNoPermission
		}
		if c.EmployerSignedAt != nil {
			return ErrContractSignError
		}
		fields["employer_sign_url"] = req.SignURL
		fields["employer_signed_at"] = &now
		// 雇主先签 → 进入待签署状态等待工人
		if c.WorkerSignedAt == nil {
			fields["status"] = model.ContractStatusPending
		}
	case "worker":
		if c.WorkerID != userID {
			return ErrContractNoPermission
		}
		if c.WorkerSignedAt != nil {
			return ErrContractSignError
		}
		fields["worker_sign_url"] = req.SignURL
		fields["worker_signed_at"] = &now
	default:
		return ErrContractSignError
	}

	// 双方都已签署 → 已签署状态
	if c.EmployerSignedAt != nil && req.SignRole == "worker" {
		fields["status"] = model.ContractStatusSigned
		fields["signed_at"] = &now
		fields["effective_at"] = &now
	} else if c.WorkerSignedAt != nil && req.SignRole == "employer" {
		fields["status"] = model.ContractStatusSigned
		fields["signed_at"] = &now
		fields["effective_at"] = &now
	}

	return s.repo.Update(id, fields)
}

// UpdateStatus 合同状态变更
func (s *contractService) UpdateStatus(id uint, req *dto.ContractStatusUpdateRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContractNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status": req.Status,
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}

	switch req.Status {
	case model.ContractStatusEffective:
		fields["effective_at"] = &now
	case model.ContractStatusCompleted:
		fields["completed_at"] = &now
	case model.ContractStatusTerminated:
		fields["terminated_at"] = &now
	case model.ContractStatusExpired:
		fields["expired_at"] = &now
	}

	_ = c
	return s.repo.Update(id, fields)
}

// ===== M 端管理 =====

// AdminList M 端合同列表
func (s *contractService) AdminList(req *dto.ContractAdminListRequest) (*utils.Pagination, []dto.ContractInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ContractAdminListOptions{
		RegionID:     req.RegionID,
		LinggongID:   req.LinggongID,
		EmployerID:   req.EmployerID,
		WorkerID:     req.WorkerID,
		ContractType: req.ContractType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.ContractInfo, 0, len(list))
	for i := range list {
		result = append(result, *toContractInfo(&list[i]))
	}
	return pagination, result, nil
}
