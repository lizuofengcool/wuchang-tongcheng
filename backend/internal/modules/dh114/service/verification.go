// Package service 同城114业务逻辑层 - 商户认证
// 依据 v3.2.1 架构方案：营业执照 + 实地认证 + 品牌授权
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"wuchang-tongcheng/internal/modules/dh114/dto"
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/modules/dh114/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrVerificationNotFound     = errors.New("认证记录不存在")
	ErrVerificationNoPermission = errors.New("无权操作此认证")
	ErrVerificationAudited      = errors.New("认证已审核，不能重复审核")
	ErrVerificationStatusInvalid = errors.New("认证状态不允许此操作")
)

// VerificationService 商户认证业务接口
type VerificationService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateVerificationRequest) (*dto.VerificationInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateVerificationRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.VerificationInfo, error)
	List(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error)
	ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error)
	FindLatestByDh114(dh114ID uint, vType string) (*dto.VerificationInfo, error)

	// M 端管理
	AdminList(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error)
	Audit(id uint, auditedBy uint, req *dto.VerificationAuditRequest) error
}

type verificationService struct {
	repo repository.VerificationRepository
}

// NewVerificationService 创建商户认证 service 实例
func NewVerificationService(repo repository.VerificationRepository) VerificationService {
	return &verificationService{repo: repo}
}

// verificationStatusText 认证状态文本
func verificationStatusTextV2(s int) string {
	switch s {
	case model.VerificationStatusPending:
		return "待审"
	case model.VerificationStatusApproved:
		return "通过"
	case model.VerificationStatusRejected:
		return "拒绝"
	case model.VerificationStatusExpired:
		return "已过期"
	}
	return ""
}

// verificationTypeText 认证类型文本
func verificationTypeText(t string) string {
	switch t {
	case model.VerificationTypeLicense:
		return "营业执照"
	case model.VerificationTypeField:
		return "实地认证"
	case model.VerificationTypeBrand:
		return "品牌授权"
	case model.VerificationTypeLicenseAndField:
		return "营业执照+实地"
	}
	return ""
}

// toVerificationInfo model -> dto
func toVerificationInfo(v *model.Dh114Verification) *dto.VerificationInfo {
	info := &dto.VerificationInfo{
		ID:                  v.ID,
		VerificationNo:      v.VerificationNo,
		Dh114ID:             v.Dh114ID,
		BusinessID:          v.BusinessID,
		UserID:              v.UserID,
		VerificationType:    v.VerificationType,
		VerificationTypeText: verificationTypeText(v.VerificationType),
		LicenseNo:           v.LicenseNo,
		LicenseImage:        v.LicenseImage,
		BusinessName:        v.BusinessName,
		LegalPerson:         v.LegalPerson,
		BusinessScope:       v.BusinessScope,
		RegisteredAddress:  v.RegisteredAddress,
		FieldAddress:        v.FieldAddress,
		FieldDate:           v.FieldDate,
		InspectorName:       v.InspectorName,
		BrandName:           v.BrandName,
		BrandAuthImage:      v.BrandAuthImage,
		Status:              v.Status,
		StatusText:          verificationStatusTextV2(v.Status),
		AuditRemark:         v.AuditRemark,
		AuditedBy:           v.AuditedBy,
		AuditedAt:           v.AuditedAt,
		ValidUntil:          v.ValidUntil,
		RegionID:            v.RegionID,
		CreatedAt:           v.CreatedAt,
		UpdatedAt:           v.UpdatedAt,
	}
	if v.FieldPhotos != nil {
		info.FieldPhotos = v.FieldPhotos
	}
	return info
}

// generateVerificationNo 生成认证单号
func generateVerificationNo() string {
	return fmt.Sprintf("DH114VF%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// Create 创建认证申请
func (s *verificationService) Create(regionID uint, userID uint, req *dto.CreateVerificationRequest) (*dto.VerificationInfo, error) {
	vType := req.VerificationType
	if vType == "" {
		vType = model.VerificationTypeLicense
	}

	v := &model.Dh114Verification{
		VerificationNo:   generateVerificationNo(),
		Dh114ID:          req.Dh114ID,
		UserID:           userID,
		VerificationType: vType,
		LicenseNo:        req.LicenseNo,
		LicenseImage:     req.LicenseImage,
		BusinessName:     req.BusinessName,
		LegalPerson:      req.LegalPerson,
		LegalPersonIDCard: req.LegalPersonIDCard,
		BusinessScope:    req.BusinessScope,
		RegisteredAddress: req.RegisteredAddress,
		FieldAddress:      req.FieldAddress,
		FieldLongitude:   req.FieldLongitude,
		FieldLatitude:    req.FieldLatitude,
		FieldDate:        req.FieldDate,
		BrandName:        req.BrandName,
		BrandAuthImage:   req.BrandAuthImage,
		ValidUntil:       req.ValidUntil,
		Status:           model.VerificationStatusPending,
	}
	v.RegionID = regionID

	if req.FieldPhotos != nil {
		if b, err := model.FromJSON(req.FieldPhotos); err == nil {
			v.FieldPhotos = b
		}
	}

	if err := s.repo.Create(v); err != nil {
		return nil, err
	}
	return toVerificationInfo(v), nil
}

// Update 更新认证申请（仅待审状态可更新）
func (s *verificationService) Update(id uint, operatorID uint, req *dto.UpdateVerificationRequest) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}
	if v.UserID != operatorID {
		return ErrVerificationNoPermission
	}
	if v.Status != model.VerificationStatusPending {
		return ErrVerificationAudited
	}

	fields := make(map[string]interface{})
	if req.LicenseNo != nil {
		fields["license_no"] = *req.LicenseNo
	}
	if req.LicenseImage != nil {
		fields["license_image"] = *req.LicenseImage
	}
	if req.BusinessName != nil {
		fields["business_name"] = *req.BusinessName
	}
	if req.LegalPerson != nil {
		fields["legal_person"] = *req.LegalPerson
	}
	if req.LegalPersonIDCard != nil {
		fields["legal_person_id_card"] = *req.LegalPersonIDCard
	}
	if req.BusinessScope != nil {
		fields["business_scope"] = *req.BusinessScope
	}
	if req.RegisteredAddress != nil {
		fields["registered_address"] = *req.RegisteredAddress
	}
	if req.FieldAddress != nil {
		fields["field_address"] = *req.FieldAddress
	}
	if req.FieldLongitude != nil {
		fields["field_longitude"] = *req.FieldLongitude
	}
	if req.FieldLatitude != nil {
		fields["field_latitude"] = *req.FieldLatitude
	}
	if req.FieldDate != nil {
		fields["field_date"] = req.FieldDate
	}
	if req.BrandName != nil {
		fields["brand_name"] = *req.BrandName
	}
	if req.BrandAuthImage != nil {
		fields["brand_auth_image"] = *req.BrandAuthImage
	}
	if req.ValidUntil != nil {
		fields["valid_until"] = req.ValidUntil
	}
	if req.FieldPhotos != nil {
		if b, err := model.FromJSON(req.FieldPhotos); err == nil {
			fields["field_photos"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除认证申请
func (s *verificationService) Delete(id uint, operatorID uint) error {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}
	if v.UserID != operatorID {
		return ErrVerificationNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取认证详情
func (s *verificationService) GetByID(id uint) (*dto.VerificationInfo, error) {
	v, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}
	return toVerificationInfo(v), nil
}

// List 认证列表
func (s *verificationService) List(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.VerificationListQuery{
		Dh114ID:          req.Dh114ID,
		UserID:           req.UserID,
		VerificationType: req.VerificationType,
		Status:           req.Status,
		Keyword:          "", // C 端不开放关键字搜索
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toVerificationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByDh114 按商户列出认证
func (s *verificationService) ListByDh114(dh114ID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByDh114(dh114ID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toVerificationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出认证
func (s *verificationService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toVerificationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// FindLatestByDh114 查询商户最新的认证
func (s *verificationService) FindLatestByDh114(dh114ID uint, vType string) (*dto.VerificationInfo, error) {
	v, err := s.repo.FindLatestByDh114(dh114ID, vType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}
	return toVerificationInfo(v), nil
}

// AdminList 管理后台列表
func (s *verificationService) AdminList(req *dto.VerificationListRequest) (*utils.Pagination, []dto.VerificationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.VerificationListQuery{
		Dh114ID:          req.Dh114ID,
		UserID:           req.UserID,
		VerificationType: req.VerificationType,
		Status:           req.Status,
		Keyword:          "", // 管理后台可考虑扩展关键字搜索
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.VerificationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toVerificationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Audit 审核认证
func (s *verificationService) Audit(id uint, auditedBy uint, req *dto.VerificationAuditRequest) error {
	if req.Status != model.VerificationStatusApproved &&
		req.Status != model.VerificationStatusRejected &&
		req.Status != model.VerificationStatusExpired {
		return ErrVerificationStatusInvalid
	}

	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrVerificationNotFound
		}
		return err
	}

	now := time.Now()
	var validUntil interface{}
	if req.ValidUntil != nil {
		validUntil = req.ValidUntil
	}
	return s.repo.UpdateAudit(id, req.Status, req.AuditRemark, auditedBy, now, validUntil)
}
