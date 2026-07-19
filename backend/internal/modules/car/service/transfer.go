// Package service 同城车辆买卖业务逻辑层 - 过户办理
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标瓜子 8 状态流程
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
	ErrTransferNotFound      = errors.New("过户单不存在")
	ErrTransferNoPermission  = errors.New("无权操作此过户单")
	ErrTransferStatusInvalid = errors.New("过户单状态不允许此操作")
)

// TransferService 过户办理业务接口
type TransferService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateTransferRequest) (*dto.TransferInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateTransferRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.TransferInfo, error)
	GetByCarID(carID uint) (*dto.TransferInfo, error)
	List(regionID uint, req *dto.TransferListRequest) (*utils.Pagination, []dto.TransferInfo, error)
	ListBySeller(sellerID uint, page, pageSize int) (*utils.Pagination, []dto.TransferInfo, error)
	ListByBuyer(buyerID uint, page, pageSize int) (*utils.Pagination, []dto.TransferInfo, error)

	// M 端管理
	AdminList(req *dto.TransferListRequest) (*utils.Pagination, []dto.TransferInfo, error)
	AdminGetByID(id uint) (*dto.TransferInfo, error)
	UpdateStatus(id uint, req *dto.TransferStatusUpdateRequest) error
}

type transferService struct {
	repo repository.TransferRepository
}

// NewTransferService 创建过户办理 service 实例
func NewTransferService(repo repository.TransferRepository) TransferService {
	return &transferService{repo: repo}
}

// transferStatusText 状态文本
func transferStatusText(status int) string {
	switch status {
	case model.TransferStatusPending:
		return "待提交"
	case model.TransferStatusSubmitted:
		return "已提交"
	case model.TransferStatusReviewing:
		return "审核中"
	case model.TransferStatusApproved:
		return "审核通过"
	case model.TransferStatusRejected:
		return "审核拒绝"
	case model.TransferStatusInProgress:
		return "办理中"
	case model.TransferStatusCompleted:
		return "已完成"
	case model.TransferStatusCanceled:
		return "已取消"
	}
	return ""
}

// toTransferInfo model -> dto
func toTransferInfo(t *model.CarTransfer) *dto.TransferInfo {
	info := &dto.TransferInfo{
		ID:                  t.ID,
		TransferNo:          t.TransferNo,
		CarID:               t.CarID,
		ContractID:          t.ContractID,
		ListingID:           t.ListingID,
		SellerID:            t.SellerID,
		SellerName:          t.SellerName,
		BuyerID:             t.BuyerID,
		BuyerName:           t.BuyerName,
		AgentID:             t.AgentID,
		AgentName:           t.AgentName,
		TransferType:        t.TransferType,
		TransferFee:         t.TransferFee,
		TaxFee:              t.TaxFee,
		OtherFee:            t.OtherFee,
		Location:            t.Location,
		AppointmentDate:     t.AppointmentDate,
		AppointmentTime:     t.AppointmentTime,
		SubmittedAt:         t.SubmittedAt,
		ReviewedAt:          t.ReviewedAt,
		CompletedAt:         t.CompletedAt,
		CanceledAt:          t.CanceledAt,
		NewLicensePlate:     t.NewLicensePlate,
		NewRegistrationCert: t.NewRegistrationCert,
		Status:              t.Status,
		StatusText:          transferStatusText(t.Status),
		Remark:              t.Remark,
		RegionID:            t.RegionID,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
	if t.VehicleRegistration != nil {
		info.VehicleRegistration = t.VehicleRegistration
	}
	if t.Documents != nil {
		info.Documents = t.Documents
	}
	return info
}

// genTransferNo 生成过户单号：TR + yyyyMMddHHmmss + 6 位随机
func genTransferNo() string {
	return fmt.Sprintf("TR%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建过户单
func (s *transferService) Create(regionID uint, userID uint, req *dto.CreateTransferRequest) (*dto.TransferInfo, error) {
	t := &model.CarTransfer{
		TransferNo:       genTransferNo(),
		CarID:            req.CarID,
		ContractID:       req.ContractID,
		ListingID:        req.ListingID,
		SellerID:         req.SellerID,
		SellerName:       req.SellerName,
		BuyerID:          req.BuyerID,
		BuyerName:        req.BuyerName,
		AgentID:          req.AgentID,
		AgentName:        req.AgentName,
		TransferType:     req.TransferType,
		TransferFee:      req.TransferFee,
		TaxFee:           req.TaxFee,
		OtherFee:         req.OtherFee,
		Location:         req.Location,
		AppointmentDate:  req.AppointmentDate,
		AppointmentTime:  req.AppointmentTime,
		Remark:           req.Remark,
		Status:           model.TransferStatusPending,
	}
	t.RegionID = regionID

	// 默认值兜底
	if t.TransferType == "" {
		t.TransferType = model.TransferTypeSale
	}

	// JSONB 字段
	if req.VehicleRegistration != nil {
		if jb, err := model.FromJSON(req.VehicleRegistration); err == nil {
			t.VehicleRegistration = jb
		}
	}
	if req.Documents != nil {
		if jb, err := model.FromJSON(req.Documents); err == nil {
			t.Documents = jb
		}
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return toTransferInfo(t), nil
}

// Update 更新过户单
func (s *transferService) Update(id uint, operatorID uint, req *dto.UpdateTransferRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTransferNotFound
		}
		return err
	}
	// 卖方/买方/经办人/管理员均可操作
	if operatorID != 0 && t.SellerID != operatorID && t.BuyerID != operatorID && t.AgentID != operatorID {
		return ErrTransferNoPermission
	}

	fields := map[string]interface{}{}
	if req.AgentID != nil {
		fields["agent_id"] = *req.AgentID
	}
	if req.AgentName != nil {
		fields["agent_name"] = *req.AgentName
	}
	if req.TransferType != nil {
		fields["transfer_type"] = *req.TransferType
	}
	if req.VehicleRegistration != nil {
		if jb, err := model.FromJSON(req.VehicleRegistration); err == nil {
			fields["vehicle_registration"] = jb
		}
	}
	if req.Documents != nil {
		if jb, err := model.FromJSON(req.Documents); err == nil {
			fields["documents"] = jb
		}
	}
	if req.TransferFee != nil {
		fields["transfer_fee"] = *req.TransferFee
	}
	if req.TaxFee != nil {
		fields["tax_fee"] = *req.TaxFee
	}
	if req.OtherFee != nil {
		fields["other_fee"] = *req.OtherFee
	}
	if req.Location != nil {
		fields["location"] = *req.Location
	}
	if req.AppointmentDate != nil {
		fields["appointment_date"] = req.AppointmentDate
	}
	if req.AppointmentTime != nil {
		fields["appointment_time"] = *req.AppointmentTime
	}
	if req.NewLicensePlate != nil {
		fields["new_license_plate"] = *req.NewLicensePlate
	}
	if req.NewRegistrationCert != nil {
		fields["new_registration_cert"] = *req.NewRegistrationCert
	}
	if req.Remark != nil {
		fields["remark"] = *req.Remark
	}

	// 状态变更
	if req.Status != nil {
		now := time.Now()
		switch *req.Status {
		case model.TransferStatusSubmitted:
			fields["status"] = model.TransferStatusSubmitted
			if t.SubmittedAt == nil {
				fields["submitted_at"] = &now
			}
		case model.TransferStatusReviewing:
			fields["status"] = model.TransferStatusReviewing
		case model.TransferStatusApproved:
			fields["status"] = model.TransferStatusApproved
			fields["reviewed_at"] = &now
		case model.TransferStatusRejected:
			fields["status"] = model.TransferStatusRejected
			fields["reviewed_at"] = &now
		case model.TransferStatusInProgress:
			fields["status"] = model.TransferStatusInProgress
		case model.TransferStatusCompleted:
			fields["status"] = model.TransferStatusCompleted
			fields["completed_at"] = &now
		case model.TransferStatusCanceled:
			fields["status"] = model.TransferStatusCanceled
			fields["canceled_at"] = &now
		default:
			fields["status"] = *req.Status
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除过户单
func (s *transferService) Delete(id uint, operatorID uint) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTransferNotFound
		}
		return err
	}
	if operatorID != 0 && t.SellerID != operatorID && t.BuyerID != operatorID {
		return ErrTransferNoPermission
	}
	// 已完成的不能删除
	if t.Status == model.TransferStatusCompleted {
		return ErrTransferStatusInvalid
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *transferService) GetByID(id uint) (*dto.TransferInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}
	return toTransferInfo(t), nil
}

// GetByCarID 按车源查询最新过户单
func (s *transferService) GetByCarID(carID uint) (*dto.TransferInfo, error) {
	t, err := s.repo.FindByCarID(carID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}
	return toTransferInfo(t), nil
}

// List C 端列表查询
func (s *transferService) List(regionID uint, req *dto.TransferListRequest) (*utils.Pagination, []dto.TransferInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TransferListOptions{
		CarID:        req.CarID,
		SellerID:     req.SellerID,
		BuyerID:      req.BuyerID,
		AgentID:      req.AgentID,
		TransferType: req.TransferType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TransferInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTransferInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListBySeller 卖方的过户单
func (s *transferService) ListBySeller(sellerID uint, page, pageSize int) (*utils.Pagination, []dto.TransferInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListBySeller(sellerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TransferInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTransferInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByBuyer 买方的过户单
func (s *transferService) ListByBuyer(buyerID uint, page, pageSize int) (*utils.Pagination, []dto.TransferInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByBuyer(buyerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TransferInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTransferInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

func (s *transferService) AdminList(req *dto.TransferListRequest) (*utils.Pagination, []dto.TransferInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.TransferAdminListOptions{
		CarID:        req.CarID,
		SellerID:     req.SellerID,
		BuyerID:      req.BuyerID,
		AgentID:      req.AgentID,
		TransferType: req.TransferType,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.TransferInfo, 0, len(list))
	for i := range list {
		result = append(result, *toTransferInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *transferService) AdminGetByID(id uint) (*dto.TransferInfo, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransferNotFound
		}
		return nil, err
	}
	return toTransferInfo(t), nil
}

// UpdateStatus M 端强制更新状态
func (s *transferService) UpdateStatus(id uint, req *dto.TransferStatusUpdateRequest) error {
	t, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTransferNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{}
	switch req.Status {
	case model.TransferStatusSubmitted:
		if t.SubmittedAt == nil {
			fields["submitted_at"] = &now
		}
	case model.TransferStatusApproved, model.TransferStatusRejected:
		fields["reviewed_at"] = &now
	case model.TransferStatusCompleted:
		fields["completed_at"] = &now
		if req.NewLicensePlate != "" {
			fields["new_license_plate"] = req.NewLicensePlate
		}
		if req.NewRegistrationCert != "" {
			fields["new_registration_cert"] = req.NewRegistrationCert
		}
	case model.TransferStatusCanceled:
		fields["canceled_at"] = &now
	}
	if req.Remark != "" {
		fields["remark"] = req.Remark
	}
	return s.repo.UpdateStatus(id, req.Status, fields)
}
