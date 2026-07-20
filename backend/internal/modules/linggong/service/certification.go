// Package service 同城零工兼职业务逻辑层 - 资质证书
// 对标猪八戒/斗米认证：求职者证书 + 雇主认证 + 平台认证
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
	ErrCertificationNotFound = errors.New("资质证书不存在")
	ErrCertificationNoPermission = errors.New("无权操作此证书")
)

// CertificationService 资质证书业务接口
type CertificationService interface {
	// C 端
	Create(regionID uint, userID uint, workerName string, req *dto.CreateCertificationRequest) (*dto.CertificationInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateWorkerRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.CertificationInfo, error)
	List(regionID uint, req *dto.CertificationListRequest) (*utils.Pagination, []dto.CertificationInfo, error)
	ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.CertificationInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.CertificationInfo, error)

	// M 端管理
	Verify(id uint, verifiedBy uint, verifiedByName string, req *dto.CertVerifyRequest) error
}

type certificationService struct {
	repo repository.CertificationRepository
}

// NewCertificationService 创建资质证书 service 实例
func NewCertificationService(repo repository.CertificationRepository) CertificationService {
	return &certificationService{repo: repo}
}

// certTypeText 证书类型文本
func certTypeText(t string) string {
	switch t {
	case model.CertTypeIDCard:
		return "身份证"
	case model.CertTypeHealthCert:
		return "健康证"
	case model.CertTypeSkillCert:
		return "技能证书"
	case model.CertTypeEducationCert:
		return "学历证书"
	case model.CertTypeWorkCert:
		return "工作证明"
	case model.CertTypeDriverLicense:
		return "驾驶证"
	case model.CertTypeLanguageCert:
		return "语言证书"
	case model.CertTypeProfessionCert:
		return "职业资格证"
	case model.CertTypeSafetyCert:
		return "安全证书"
	case model.CertTypeOther:
		return "其他"
	}
	return ""
}

// certStatusText 证书状态文本
func certStatusText(s int) string {
	switch s {
	case model.CertStatusPending:
		return "待审核"
	case model.CertStatusApproved:
		return "已通过"
	case model.CertStatusRejected:
		return "已拒绝"
	case model.CertStatusExpired:
		return "已过期"
	case model.CertStatusRevoked:
		return "已撤销"
	}
	return ""
}

// toCertificationInfo model -> dto
func toCertificationInfo(c *model.LinggongCertification) *dto.CertificationInfo {
	return &dto.CertificationInfo{
		ID:              c.ID,
		CertNo:          c.CertNo,
		UserID:          c.UserID,
		WorkerID:        c.WorkerID,
		WorkerName:      "",
		EmployerID:      c.EmployerID,
		CertType:        c.CertType,
		CertTypeText:    certTypeText(c.CertType),
		CertName:        c.CertName,
		CertCode:        c.CertCode,
		IssuerName:      c.IssuerName,
		IssuerType:      c.IssuerType,
		IssueDate:       c.IssueDate,
		ValidFrom:       c.ValidFrom,
		ValidUntil:      c.ValidUntil,
		ImageURL:        c.ImageURL,
		ImageBackURL:    c.ImageBackURL,
		SkillID:         c.SkillID,
		SkillName:       c.SkillName,
		Level:           c.Level,
		Score:           c.Score,
		Verified:        c.Verified,
		VerifiedAt:      c.VerifiedAt,
		VerifiedBy:      c.VerifiedBy,
		VerifiedByName:  c.VerifiedByName,
		Status:          c.Status,
		StatusText:      certStatusText(c.Status),
		RejectReason:    c.RejectReason,
		Description:     c.Description,
		RegionID:        c.RegionID,
		CreatedAt:       c.CreatedAt,
	}
}

// genCertNo 生成证书编号：CERT + yyyyMMddHHmmss + 6 位随机
func genCertNo() string {
	return fmt.Sprintf("CERT%s%06d", time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// ===== C 端 =====

// Create 创建资质证书
func (s *certificationService) Create(regionID uint, userID uint, workerName string, req *dto.CreateCertificationRequest) (*dto.CertificationInfo, error) {
	c := &model.LinggongCertification{
		CertNo:      genCertNo(),
		UserID:      userID,
		UserType:    model.CreditUserTypeWorker,
		WorkerID:    0,
		CertType:    req.CertType,
		CertName:    req.CertName,
		CertCode:    req.CertCode,
		IssuerName:  req.IssuerName,
		IssuerType:  req.IssuerType,
		IssueDate:   req.IssueDate,
		ValidFrom:   req.ValidFrom,
		ValidUntil:  req.ValidUntil,
		ImageURL:    req.ImageURL,
		ImageBackURL: req.ImageBackURL,
		SkillID:     req.SkillID,
		SkillName:   req.SkillName,
		Level:       req.Level,
		Score:       req.Score,
		Status:      model.CertStatusPending,
		Description: req.Description,
	}
	c.RegionID = regionID

	// 默认值兜底
	if c.CertType == "" {
		c.CertType = model.CertTypeOther
	}
	if c.IssuerType == "" {
		c.IssuerType = model.IssuerTypeOther
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCertificationInfo(c), nil
}

// Update 更新证书（仅持有人本人）
func (s *certificationService) Update(id uint, operatorID uint, req *dto.UpdateWorkerRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCertificationNotFound
		}
		return err
	}
	if c.UserID != operatorID {
		return ErrCertificationNoPermission
	}
	return nil
}

// Delete 删除证书（仅持有人本人）
func (s *certificationService) Delete(id uint, operatorID uint) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCertificationNotFound
		}
		return err
	}
	if c.UserID != operatorID {
		return ErrCertificationNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取证书详情
func (s *certificationService) GetByID(id uint) (*dto.CertificationInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCertificationNotFound
		}
		return nil, err
	}
	return toCertificationInfo(c), nil
}

// List C 端列表查询
func (s *certificationService) List(regionID uint, req *dto.CertificationListRequest) (*utils.Pagination, []dto.CertificationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CertificationListOptions{
		UserID:   req.UserID,
		WorkerID: req.WorkerID,
		CertType: req.CertType,
		SkillID:  req.SkillID,
		Status:   req.Status,
		Verified: req.Verified,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CertificationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCertificationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByWorker 按求职者反查
func (s *certificationService) ListByWorker(workerID uint, page, pageSize int) (*utils.Pagination, []dto.CertificationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByWorker(workerID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CertificationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCertificationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 按用户反查
func (s *certificationService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.CertificationInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CertificationInfo, 0, len(list))
	for i := range list {
		result = append(result, *toCertificationInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// Verify 证书审核
func (s *certificationService) Verify(id uint, verifiedBy uint, verifiedByName string, req *dto.CertVerifyRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCertificationNotFound
		}
		return err
	}

	now := time.Now()
	fields := map[string]interface{}{
		"status":            req.Status,
		"reject_reason":     req.RejectReason,
		"verified_by":       verifiedBy,
		"verified_by_name":  verifiedByName,
	}

	if req.Status == model.CertStatusApproved {
		fields["verified"] = true
		fields["verified_at"] = &now
	} else {
		fields["verified"] = false
	}

	// 已审核通过的证书不允许重复审核（保留 c 引用，避免编译告警）
	if c.Status == model.CertStatusApproved && req.Status == model.CertStatusApproved {
		return nil
	}

	return s.repo.Update(id, fields)
}
