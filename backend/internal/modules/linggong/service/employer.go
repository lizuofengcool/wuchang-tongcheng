// Package service 同城零工兼职业务逻辑层 - 雇主认证
// 对标斗米企业认证 + 猪八戒雇主：营业执照 + 法人认证 + 信用等级
// 4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/linggong/dto"
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/modules/linggong/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrEmployerNotFound     = errors.New("雇主不存在")
	ErrEmployerNoPermission = errors.New("无权操作此雇主")
	ErrEmployerStatusInvalid = errors.New("雇主状态不允许此操作")
	ErrEmployerDuplicate    = errors.New("已存在雇主认证")
)

// EmployerService 雇主认证业务接口
type EmployerService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreateEmployerRequest) (*dto.EmployerInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateEmployerRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.EmployerInfo, error)
	GetByUserID(userID uint) (*dto.EmployerInfo, error)
	List(regionID uint, req *dto.EmployerListRequest) (*utils.Pagination, []dto.EmployerInfo, error)

	// M 端管理
	AdminList(req *dto.EmployerAdminListRequest) (*utils.Pagination, []dto.EmployerInfo, error)
	Audit(id uint, reviewerID uint, reviewerName string, req *dto.EmployerAuditRequest) error
	UpdateStatus(id uint, status int) error
}

type employerService struct {
	repo repository.EmployerRepository
}

// NewEmployerService 创建雇主认证 service 实例
func NewEmployerService(repo repository.EmployerRepository) EmployerService {
	return &employerService{repo: repo}
}

// employerStatusText 雇主状态文本
func employerStatusText(s int) string {
	switch s {
	case model.EmployerStatusPending:
		return "待审核"
	case model.EmployerStatusApproved:
		return "已通过"
	case model.EmployerStatusRejected:
		return "已拒绝"
	case model.EmployerStatusFrozen:
		return "已冻结"
	case model.EmployerStatusBanned:
		return "已封禁"
	}
	return ""
}

// employerTypeText 雇主类型文本
func employerTypeText(t string) string {
	switch t {
	case model.EmployerTypePersonal:
		return "个人雇主"
	case model.EmployerTypeCompany:
		return "企业雇主"
	case model.EmployerTypeAgent:
		return "中介"
	}
	return ""
}

// employerLevelText 雇主等级文本
func employerLevelText(l int) string {
	switch l {
	case model.EmployerLevelBronze:
		return "青铜"
	case model.EmployerLevelSilver:
		return "白银"
	case model.EmployerLevelGold:
		return "黄金"
	case model.EmployerLevelPlatinum:
		return "铂金"
	case model.EmployerLevelDiamond:
		return "钻石"
	}
	return ""
}

// toEmployerInfo model -> dto
func toEmployerInfo(e *model.LinggongEmployer) *dto.EmployerInfo {
	return &dto.EmployerInfo{
		ID:                  e.ID,
		UserID:              e.UserID,
		EmployerType:        e.EmployerType,
		EmployerTypeText:    employerTypeText(e.EmployerType),
		CompanyName:         e.CompanyName,
		CompanyShortName:    e.CompanyShortName,
		ContactName:         e.ContactName,
		ContactPhone:        e.ContactPhone,
		ContactEmail:        e.ContactEmail,
		ContactWechat:       e.ContactWechat,
		LicenseNo:           e.LicenseNo,
		LicenseURL:          e.LicenseURL,
		LegalPerson:         e.LegalPerson,
		LegalPersonIDCard:   e.LegalPersonIDCard,
		LegalPersonIDCardURL: e.LegalPersonIDCardURL,
		BankAccount:         e.BankAccount,
		BankName:            e.BankName,
		BrandAuthURL:        e.BrandAuthURL,
		CompanyAddress:      e.CompanyAddress,
		CompanyLatitude:     e.CompanyLatitude,
		CompanyLongitude:    e.CompanyLongitude,
		CompanyDescription:  e.CompanyDescription,
		CompanyLogo:         e.CompanyLogo,
		CompanyCover:        e.CompanyCover,
		Industry:            e.Industry,
		CompanySize:         e.CompanySize,
		Level:               e.Level,
		LevelText:           employerLevelText(e.Level),
		CreditScore:         e.CreditScore,
		Status:              e.Status,
		StatusText:          employerStatusText(e.Status),
		RejectReason:        e.RejectReason,
		VerifiedAt:          e.VerifiedAt,
		VerifiedBy:          e.VerifiedBy,
		VerifiedByName:      e.VerifiedByName,
		PublishedCount:      e.PublishedCount,
		OngoingCount:        e.OngoingCount,
		CompletedCount:      e.CompletedCount,
		TotalWorkers:        e.TotalWorkers,
		TotalPaid:           e.TotalPaid,
		AvgRating:           e.AvgRating,
		RatingCount:         e.RatingCount,
		RegionID:            e.RegionID,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           e.UpdatedAt,
	}
}

// ===== C 端 =====

// Create 创建雇主认证（每用户仅一份）
func (s *employerService) Create(regionID uint, userID uint, req *dto.CreateEmployerRequest) (*dto.EmployerInfo, error) {
	// 唯一性校验
	if existing, err := s.repo.FindByUserID(userID); err == nil && existing != nil {
		return nil, ErrEmployerDuplicate
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	e := &model.LinggongEmployer{
		UserID:                userID,
		EmployerType:          req.EmployerType,
		CompanyName:           req.CompanyName,
		CompanyShortName:      req.CompanyShortName,
		ContactName:           req.ContactName,
		ContactPhone:          req.ContactPhone,
		ContactEmail:          req.ContactEmail,
		ContactWechat:         req.ContactWechat,
		LicenseNo:             req.LicenseNo,
		LicenseURL:            req.LicenseURL,
		LegalPerson:           req.LegalPerson,
		LegalPersonIDCard:     req.LegalPersonIDCard,
		LegalPersonIDCardURL:  req.LegalPersonIDCardURL,
		BankAccount:           req.BankAccount,
		BankName:              req.BankName,
		BrandAuthURL:          req.BrandAuthURL,
		CompanyAddress:        req.CompanyAddress,
		CompanyLatitude:       req.CompanyLatitude,
		CompanyLongitude:      req.CompanyLongitude,
		CompanyDescription:    req.CompanyDescription,
		CompanyLogo:           req.CompanyLogo,
		CompanyCover:          req.CompanyCover,
		Industry:              req.Industry,
		CompanySize:           req.CompanySize,
		Level:                 model.EmployerLevelBronze,
		CreditScore:           100,
		Status:                model.EmployerStatusPending,
	}
	e.RegionID = regionID

	// 默认值兜底
	if e.EmployerType == "" {
		e.EmployerType = model.EmployerTypePersonal
	}

	if err := s.repo.Create(e); err != nil {
		return nil, err
	}
	return toEmployerInfo(e), nil
}

// Update 更新雇主认证（仅本人）
func (s *employerService) Update(id uint, operatorID uint, req *dto.UpdateEmployerRequest) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmployerNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrEmployerNoPermission
	}

	// 已通过认证的雇主不允许修改关键字段（MVP 简化：允许更新但需重新审核）
	fields := map[string]interface{}{}
	if req.EmployerType != nil {
		fields["employer_type"] = *req.EmployerType
	}
	if req.CompanyName != nil {
		fields["company_name"] = *req.CompanyName
	}
	if req.CompanyShortName != nil {
		fields["company_short_name"] = *req.CompanyShortName
	}
	if req.ContactName != nil {
		fields["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		fields["contact_email"] = *req.ContactEmail
	}
	if req.ContactWechat != nil {
		fields["contact_wechat"] = *req.ContactWechat
	}
	if req.LicenseNo != nil {
		fields["license_no"] = *req.LicenseNo
	}
	if req.LicenseURL != nil {
		fields["license_url"] = *req.LicenseURL
	}
	if req.LegalPerson != nil {
		fields["legal_person"] = *req.LegalPerson
	}
	if req.LegalPersonIDCard != nil {
		fields["legal_person_id_card"] = *req.LegalPersonIDCard
	}
	if req.LegalPersonIDCardURL != nil {
		fields["legal_person_id_card_url"] = *req.LegalPersonIDCardURL
	}
	if req.BankAccount != nil {
		fields["bank_account"] = *req.BankAccount
	}
	if req.BankName != nil {
		fields["bank_name"] = *req.BankName
	}
	if req.BrandAuthURL != nil {
		fields["brand_auth_url"] = *req.BrandAuthURL
	}
	if req.CompanyAddress != nil {
		fields["company_address"] = *req.CompanyAddress
	}
	if req.CompanyLatitude != nil {
		fields["company_latitude"] = *req.CompanyLatitude
	}
	if req.CompanyLongitude != nil {
		fields["company_longitude"] = *req.CompanyLongitude
	}
	if req.CompanyDescription != nil {
		fields["company_description"] = *req.CompanyDescription
	}
	if req.CompanyLogo != nil {
		fields["company_logo"] = *req.CompanyLogo
	}
	if req.CompanyCover != nil {
		fields["company_cover"] = *req.CompanyCover
	}
	if req.Industry != nil {
		fields["industry"] = *req.Industry
	}
	if req.CompanySize != nil {
		fields["company_size"] = *req.CompanySize
	}

	// 已通过认证的雇主修改关键资质时，状态回退到待审核
	if e.Status == model.EmployerStatusApproved {
		if req.LicenseNo != nil || req.LicenseURL != nil ||
			req.LegalPerson != nil || req.LegalPersonIDCard != nil || req.LegalPersonIDCardURL != nil {
			fields["status"] = model.EmployerStatusPending
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除雇主认证（仅本人）
func (s *employerService) Delete(id uint, operatorID uint) error {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmployerNotFound
		}
		return err
	}
	if e.UserID != operatorID {
		return ErrEmployerNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取雇主详情
func (s *employerService) GetByID(id uint) (*dto.EmployerInfo, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployerNotFound
		}
		return nil, err
	}
	return toEmployerInfo(e), nil
}

// GetByUserID 按用户ID查询雇主
func (s *employerService) GetByUserID(userID uint) (*dto.EmployerInfo, error) {
	e, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployerNotFound
		}
		return nil, err
	}
	return toEmployerInfo(e), nil
}

// List C 端雇主列表（地区隔离，默认仅展示已通过）
func (s *employerService) List(regionID uint, req *dto.EmployerListRequest) (*utils.Pagination, []dto.EmployerInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EmployerListOptions{
		EmployerType: req.EmployerType,
		CompanyName:  req.CompanyName,
		Status:       req.Status,
		Level:        req.Level,
		Industry:     req.Industry,
		Keyword:      req.Keyword,
	}

	// C 端默认仅展示已通过审核
	if opts.Status == nil {
		approved := model.EmployerStatusApproved
		opts.Status = &approved
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmployerInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmployerInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== M 端管理 =====

// AdminList M 端雇主列表（跨地区）
func (s *employerService) AdminList(req *dto.EmployerAdminListRequest) (*utils.Pagination, []dto.EmployerInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.EmployerAdminListOptions{
		RegionID:     req.RegionID,
		UserID:       req.UserID,
		EmployerType: req.EmployerType,
		Status:       req.Status,
		Level:        req.Level,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.EmployerInfo, 0, len(list))
	for i := range list {
		result = append(result, *toEmployerInfo(&list[i]))
	}
	return pagination, result, nil
}

// Audit M 端审核雇主认证
// 通过：更新 verified_at/verified_by；拒绝：记录 reject_reason
func (s *employerService) Audit(id uint, reviewerID uint, reviewerName string, req *dto.EmployerAuditRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmployerNotFound
		}
		return err
	}

	fields := map[string]interface{}{
		"status":            req.Status,
		"reject_reason":     req.RejectReason,
		"verified_by":       reviewerID,
		"verified_by_name":  reviewerName,
	}

	now := time.Now()
	switch req.Status {
	case model.EmployerStatusApproved:
		fields["verified_at"] = &now
		fields["reject_reason"] = ""
	case model.EmployerStatusRejected:
		// 拒绝时清空 verified_at
		fields["verified_at"] = nil
	}

	return s.repo.Update(id, fields)
}

// UpdateStatus M 端更新雇主状态（冻结/解冻/封禁）
func (s *employerService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEmployerNotFound
		}
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}
