// Package service 商户中台业务逻辑层 - 认证
// 支持企业认证（business）与个人认证（personal）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/merchant/dto"
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/modules/merchant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrVerificationNotFound     = errors.New("认证记录不存在")
	ErrVerificationAudited      = errors.New("认证已审核，不能重复审核")
	ErrVerificationStatusInvalid = errors.New("认证状态不允许此操作")
)

// VerificationService 认证业务接口
type VerificationService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateVerificationRequest) (*dto.VerificationInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateVerificationRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.VerificationInfo, error)
	List(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error)
	ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error)

	// M 端管理
	AdminList(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error)
	Audit(id uint, req *dto.VerificationAuditRequest) error
}

type verificationService struct {
	repo repository.VerificationRepository
}

// NewVerificationService 创建认证 service 实例
func NewVerificationService(repo repository.VerificationRepository) VerificationService {
	return &verificationService{repo: repo}
}

// verifyStatusText 认证状态文本
func verifyStatusText(s int) string {
	switch s {
	case model.VerifyStatusPending:
		return "待审"
	case model.VerifyStatusApproved:
		return "通过"
	case model.VerifyStatusRejected:
		return "拒绝"
	}
	return ""
}

// verifyTypeText 认证类型文本
func verifyTypeText(t string) string {
	switch t {
	case model.VerifyTypeBusiness:
		return "企业认证"
	case model.VerifyTypePersonal:
		return "个人认证"
	}
	return ""
}

// Create 提交认证
func (s *verificationService) Create(regionID uint, userID uint, req *dto.CreateVerificationRequest) (*dto.VerificationInfo, error) {
	v := &model.Verification{
		ShopID:        req.ShopID,
		Type:          req.Type,
		LicenseNo:     req.LicenseNo,
		LicenseImage:  req.LicenseImage,
		LegalPerson:   req.LegalPerson,
		LegalPersonID: req.LegalPersonID,
		Status:        model.VerifyStatusPending,
	}
	v.RegionID = regionID
	_ = userID // 备用：可写入申请用户 ID
	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	return s.toInfo(v), nil
}

// Update 更新认证
func (s *verificationService) Update(id uint, operatorID uint, req *dto.UpdateVerificationRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}
	if v.Status != model.VerifyStatusPending {
		return ErrVerificationAudited
	}
	fields := make(map[string]interface{})
	if req.Type != nil {
		fields["type"] = *req.Type
	}
	if req.LicenseNo != nil {
		fields["license_no"] = *req.LicenseNo
	}
	if req.LicenseImage != nil {
		fields["license_image"] = *req.LicenseImage
	}
	if req.LegalPerson != nil {
		fields["legal_person"] = *req.LegalPerson
	}
	if req.LegalPersonID != nil {
		fields["legal_person_id"] = *req.LegalPersonID
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除认证记录
func (s *verificationService) Delete(id uint, operatorID uint) error {
	_, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 认证详情
func (s *verificationService) GetByID(id uint) (*dto.VerificationInfo, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}
	return s.toInfo(v), nil
}

// List 认证列表
func (s *verificationService) List(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.VerificationListOptions{
		ShopID: req.ShopID,
		Type:   req.Type,
		Status: req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByShop 按商户列出认证
func (s *verificationService) ListByShop(shopID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.FindByShopID(shopID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *s.toInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminList 管理后台列表
func (s *verificationService) AdminList(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	return s.List(req)
}

// Audit 审核认证
func (s *verificationService) Audit(id uint, req *dto.VerificationAuditRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}
	if v.Status != model.VerifyStatusPending {
		return ErrVerificationAudited
	}
	fields := map[string]interface{}{
		"status":        req.Status,
		"audit_reason":  req.AuditReason,
		"audited_at":    time.Now(),
	}
	return s.repo.UpdateFields(id, fields)
}

// toInfo 模型转 DTO
func (s *verificationService) toInfo(m *model.Verification) *dto.VerificationInfo {
	return &dto.VerificationInfo{
		ID:            m.ID,
		ShopID:        m.ShopID,
		RegionID:      m.RegionID,
		Type:          m.Type,
		TypeText:      verifyTypeText(m.Type),
		LicenseNo:     m.LicenseNo,
		LicenseImage:  m.LicenseImage,
		LegalPerson:   m.LegalPerson,
		LegalPersonID: m.LegalPersonID,
		Status:        m.Status,
		StatusText:    verifyStatusText(m.Status),
		AuditReason:   m.AuditReason,
		AuditedAt:     m.AuditedAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
