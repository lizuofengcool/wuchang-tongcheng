// Package service 公司信息 + 企业认证业务逻辑层
// 依据 v3.2.1 架构方案：对标 BOSS直聘公司主页/认证
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/job/dto"
	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/modules/job/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCompanyNotFound     = errors.New("公司不存在")
	ErrCompanyExists       = errors.New("公司已存在")
	ErrCompanyNoPermission = errors.New("无权操作此公司")
	ErrCertNotFound        = errors.New("认证记录不存在")
	ErrCertExists          = errors.New("已提交过此认证")
)

// CompanyService 公司业务接口
type CompanyService interface {
	Create(userID uint, req *dto.CompanyCreateRequest) (*dto.CompanyResponse, error)
	Update(id, operatorID uint, req *dto.CompanyUpdateRequest) error
	GetByID(id uint, userID uint) (*dto.CompanyResponse, error)
	GetMyCompany(userID uint) (*dto.CompanyResponse, error)
	List(req *dto.CompanyListQuery) (*utils.Pagination, []dto.CompanyResponse, error)

	// 关注
	Follow(userID, companyID uint, req *dto.CompanyFollowRequest) error
	Unfollow(userID, companyID uint) error
	IsFollowing(userID, companyID uint) (bool, error)
	ListFollowing(userID uint, page, pageSize int) (*utils.Pagination, []dto.CompanyResponse, error)

	// 审核
	Audit(id uint, req *dto.CompanyAuditRequest) error

	// 认证
	CreateCert(userID uint, req *dto.CertificationCreateRequest) (*dto.CertificationResponse, error)
	GetCert(id uint) (*dto.CertificationResponse, error)
	ListCerts(query dto.CertificationListQuery) (*utils.Pagination, []dto.CertificationResponse, error)
	ListCertsByCompany(companyID uint) ([]dto.CertificationResponse, error)
	ProcessCert(id uint, verifierID uint, verifierName string, req *dto.CertificationProcessRequest) (*dto.CertificationResponse, error)
}

type companyService struct {
	repo      repository.CompanyRepository
	certRepo  repository.CertificationRepository
	favRepo   repository.InteractionRepository
}

// NewCompanyService 创建公司 service 实例
func NewCompanyService(repo repository.CompanyRepository, certRepo repository.CertificationRepository, favRepo repository.InteractionRepository) CompanyService {
	return &companyService{repo: repo, certRepo: certRepo, favRepo: favRepo}
}

// levelText 公司等级文本
func levelText(level int) string {
	switch level {
	case model.CompanyLevelVerified:
		return "认证企业"
	case model.CompanyLevelGold:
		return "金牌企业"
	case model.CompanyLevelDiamond:
		return "钻石企业"
	}
	return "普通企业"
}

// companyStatusText 公司状态文本
func companyStatusText(status int) string {
	switch status {
	case model.CompanyStatusPending:
		return "待审核"
	case model.CompanyStatusApproved:
		return "已通过"
	case model.CompanyStatusRejected:
		return "已拒绝"
	case model.CompanyStatusFrozen:
		return "已冻结"
	case model.CompanyStatusClosed:
		return "已关闭"
	}
	return "未知"
}

// certStatusText 认证状态文本
func certStatusText(status int) string {
	switch status {
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
	return "未知"
}

func toCompanyResponse(c *model.JobCompany) *dto.CompanyResponse {
	resp := &dto.CompanyResponse{
		ID:                c.ID,
		UserID:            c.UserID,
		Name:              c.Name,
		ShortName:         c.ShortName,
		Logo:              c.Logo,
		Banner:            c.Banner,
		Description:       c.Description,
		Industry:          c.Industry,
		Scale:             c.Scale,
		Level:             c.Level,
		LevelText:         levelText(c.Level),
		Status:            c.Status,
		StatusText:        companyStatusText(c.Status),
		ContactName:       c.ContactName,
		ContactPhone:      c.ContactPhone,
		ContactEmail:      c.ContactEmail,
		ContactWechat:     c.ContactWechat,
		Address:           c.Address,
		Latitude:          c.Latitude,
		Longitude:         c.Longitude,
		BusinessLicense:   c.BusinessLicense,
		LicenseNo:         c.LicenseNo,
		IDCardFront:       c.IDCardFront,
		IDCardBack:        c.IDCardBack,
		LegalPerson:       c.LegalPerson,
		LegalPersonIDCard: c.LegalPersonIDCard,
		RegisteredCapital: c.RegisteredCapital,
		FoundedAt:         c.FoundedAt,
		Website:           c.Website,
		VerifiedAt:        c.VerifiedAt,
		ApprovedAt:        c.ApprovedAt,
		RejectedReason:    c.RejectedReason,
		ClosedAt:          c.ClosedAt,
		FollowerCount:     c.FollowerCount,
		JobCount:          c.JobCount,
		EmployeeCount:     c.EmployeeCount,
		ActiveJobCount:    c.ActiveJobCount,
		TotalHiredCount:   c.TotalHiredCount,
		GoodRate:          c.GoodRate,
		Deposit:           c.Deposit,
		Tags:              []string{},
		CreatedAt:         c.CreatedAt,
	}
	if c.Tags != nil {
		var tags []string
		_ = c.Tags.Parse(&tags)
		if tags != nil {
			resp.Tags = tags
		}
	}
	return resp
}

func toCertResponse(c *model.JobCertification) *dto.CertificationResponse {
	return &dto.CertificationResponse{
		ID:                c.ID,
		CompanyID:         c.CompanyID,
		UserID:            c.UserID,
		CertType:          c.CertType,
		CertNo:            c.CertNo,
		CertName:          c.CertName,
		CertImage:         c.CertImage,
		LegalPerson:       c.LegalPerson,
		LegalPersonIDCard: c.LegalPersonIDCard,
		RegisteredCapital: c.RegisteredCapital,
		BusinessScope:     c.BusinessScope,
		ValidFrom:         c.ValidFrom,
		ValidTo:           c.ValidTo,
		Status:            c.Status,
		StatusText:        certStatusText(c.Status),
		VerifiedAt:        c.VerifiedAt,
		VerifierID:        c.VerifierID,
		VerifierName:      c.VerifierName,
		RejectReason:      c.RejectReason,
		ExpiredAt:         c.ExpiredAt,
		CreatedAt:         c.CreatedAt,
	}
}

func (s *companyService) Create(userID uint, req *dto.CompanyCreateRequest) (*dto.CompanyResponse, error) {
	// 检查用户是否已有公司
	if existing, _ := s.repo.FindByUserID(userID); existing != nil {
		return nil, ErrCompanyExists
	}

	c := &model.JobCompany{
		UserID:              userID,
		Name:                req.Name,
		ShortName:           req.ShortName,
		Logo:                req.Logo,
		Banner:              req.Banner,
		Description:         req.Description,
		Industry:            req.Industry,
		Scale:               req.Scale,
		ContactName:         req.ContactName,
		ContactPhone:        req.ContactPhone,
		ContactEmail:        req.ContactEmail,
		ContactWechat:       req.ContactWechat,
		Address:             req.Address,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		BusinessLicense:     req.BusinessLicense,
		LicenseNo:           req.LicenseNo,
		IDCardFront:         req.IDCardFront,
		IDCardBack:          req.IDCardBack,
		LegalPerson:         req.LegalPerson,
		LegalPersonIDCard:   req.LegalPersonIDCard,
		RegisteredCapital:   req.RegisteredCapital,
		FoundedAt:           req.FoundedAt,
		Website:             req.Website,
		Deposit:             req.Deposit,
		Status:              model.CompanyStatusPending,
		Level:               model.CompanyLevelNormal,
		GoodRate:            100.00,
	}
	if len(req.Tags) > 0 {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			c.Tags = jb
		}
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCompanyResponse(c), nil
}

func (s *companyService) Update(id, operatorID uint, req *dto.CompanyUpdateRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCompanyNotFound
		}
		return err
	}
	if c.UserID != operatorID {
		return ErrCompanyNoPermission
	}

	fields := make(map[string]interface{})
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.ShortName != "" {
		fields["short_name"] = req.ShortName
	}
	if req.Logo != "" {
		fields["logo"] = req.Logo
	}
	if req.Banner != "" {
		fields["banner"] = req.Banner
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Industry != "" {
		fields["industry"] = req.Industry
	}
	if req.Scale != "" {
		fields["scale"] = req.Scale
	}
	if req.ContactName != "" {
		fields["contact_name"] = req.ContactName
	}
	if req.ContactPhone != "" {
		fields["contact_phone"] = req.ContactPhone
	}
	if req.ContactEmail != "" {
		fields["contact_email"] = req.ContactEmail
	}
	if req.ContactWechat != "" {
		fields["contact_wechat"] = req.ContactWechat
	}
	if req.Address != "" {
		fields["address"] = req.Address
	}
	if req.Latitude != 0 {
		fields["latitude"] = req.Latitude
	}
	if req.Longitude != 0 {
		fields["longitude"] = req.Longitude
	}
	if req.Website != "" {
		fields["website"] = req.Website
	}
	if req.Tags != nil {
		if jb, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = jb
		}
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *companyService) GetByID(id uint, userID uint) (*dto.CompanyResponse, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	resp := toCompanyResponse(c)
	if userID > 0 {
		following, _ := s.repo.FavExists(userID, id)
		resp.IsFollowing = following
	}
	return resp, nil
}

func (s *companyService) GetMyCompany(userID uint) (*dto.CompanyResponse, error) {
	c, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	return toCompanyResponse(c), nil
}

func (s *companyService) List(req *dto.CompanyListQuery) (*utils.Pagination, []dto.CompanyResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.List(repository.CompanyListQuery{
		UserID:   req.UserID,
		Status:   req.Status,
		Level:    req.Level,
		Industry: req.Industry,
		Keyword:  req.Keyword,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CompanyResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCompanyResponse(&list[i]))
	}
	return pagination, result, nil
}

// ===== 关注 =====

func (s *companyService) Follow(userID, companyID uint, req *dto.CompanyFollowRequest) error {
	exists, err := s.repo.FavExists(userID, companyID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	fav := &model.JobFavorite{
		UserID:       userID,
		CompanyID:    companyID,
		FavoriteType: model.FavoriteTypeCompany,
		Notify:       req.Notify,
	}
	if err := s.repo.CreateFav(fav); err != nil {
		return err
	}
	return s.repo.IncrFollowerCount(companyID)
}

func (s *companyService) Unfollow(userID, companyID uint) error {
	exists, _ := s.repo.FavExists(userID, companyID)
	if !exists {
		return nil
	}
	if err := s.repo.DeleteFav(userID, companyID); err != nil {
		return err
	}
	return s.repo.DecrFollowerCount(companyID)
}

func (s *companyService) IsFollowing(userID, companyID uint) (bool, error) {
	return s.repo.FavExists(userID, companyID)
}

func (s *companyService) ListFollowing(userID uint, page, pageSize int) (*utils.Pagination, []dto.CompanyResponse, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.favRepo.ListFavs(userID, model.FavoriteTypeCompany, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CompanyResponse, 0, len(favs))
	for _, fav := range favs {
		c, err := s.repo.FindByID(fav.CompanyID)
		if err != nil {
			continue
		}
		resp := toCompanyResponse(c)
		resp.IsFollowing = true
		result = append(result, *resp)
	}
	return pagination, result, nil
}

// ===== 审核 =====

func (s *companyService) Audit(id uint, req *dto.CompanyAuditRequest) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCompanyNotFound
		}
		return err
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status": req.Status,
		"level":  req.Level,
	}
	switch req.Status {
	case model.CompanyStatusApproved:
		fields["approved_at"] = &now
		if c.VerifiedAt == nil {
			fields["verified_at"] = &now
		}
	case model.CompanyStatusRejected:
		fields["rejected_reason"] = req.Reason
	case model.CompanyStatusClosed:
		fields["closed_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// ===== 企业认证 =====

func (s *companyService) CreateCert(userID uint, req *dto.CertificationCreateRequest) (*dto.CertificationResponse, error) {
	cert := &model.JobCertification{
		CompanyID:         req.CompanyID,
		UserID:            userID,
		CertType:          req.CertType,
		CertNo:            req.CertNo,
		CertName:          req.CertName,
		CertImage:         req.CertImage,
		LegalPerson:       req.LegalPerson,
		LegalPersonIDCard: req.LegalPersonIDCard,
		RegisteredCapital: req.RegisteredCapital,
		BusinessScope:     req.BusinessScope,
		ValidFrom:         req.ValidFrom,
		ValidTo:           req.ValidTo,
		Status:            model.CertStatusPending,
	}
	if cert.CertType == "" {
		cert.CertType = model.CertTypeBusinessLicense
	}
	if err := s.certRepo.Create(cert); err != nil {
		return nil, err
	}
	return toCertResponse(cert), nil
}

func (s *companyService) GetCert(id uint) (*dto.CertificationResponse, error) {
	cert, err := s.certRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCertNotFound
		}
		return nil, err
	}
	return toCertResponse(cert), nil
}

func (s *companyService) ListCerts(query dto.CertificationListQuery) (*utils.Pagination, []dto.CertificationResponse, error) {
	pagination := utils.NewPagination(query.Page, query.PageSize)
	list, total, err := s.certRepo.List(repository.CertificationListQuery{
		CompanyID: query.CompanyID,
		UserID:    query.UserID,
		Status:    query.Status,
		CertType:  query.CertType,
	}, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.CertificationResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCertResponse(&list[i]))
	}
	return pagination, result, nil
}

func (s *companyService) ListCertsByCompany(companyID uint) ([]dto.CertificationResponse, error) {
	list, err := s.certRepo.ListByCompanyID(companyID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CertificationResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCertResponse(&list[i]))
	}
	return result, nil
}

func (s *companyService) ProcessCert(id uint, verifierID uint, verifierName string, req *dto.CertificationProcessRequest) (*dto.CertificationResponse, error) {
	cert, err := s.certRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCertNotFound
		}
		return nil, err
	}
	if cert.Status != model.CertStatusPending {
		return nil, errors.New("认证状态不允许此操作")
	}
	now := time.Now()
	fields := map[string]interface{}{
		"status":         req.Status,
		"verifier_id":    verifierID,
		"verifier_name":  verifierName,
	}
	switch req.Status {
	case model.CertStatusApproved:
		fields["verified_at"] = &now
	case model.CertStatusRejected:
		fields["reject_reason"] = req.RejectReason
	case model.CertStatusRevoked:
		fields["expired_at"] = &now
	}
	if err := s.certRepo.Update(id, fields); err != nil {
		return nil, err
	}
	updated, _ := s.certRepo.FindByID(id)
	return toCertResponse(updated), nil
}
