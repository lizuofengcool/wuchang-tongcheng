// Package service 同城商城业务逻辑层 - 店铺
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/mall/dto"
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/modules/mall/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrShopNotFound = errors.New("店铺不存在")
)

// ShopService 店铺业务接口
type ShopService interface {
	Create(regionID, userID uint, userName string, req *dto.CreateShopRequest) (*dto.ShopInfo, error)
	Update(id, userID uint, req *dto.UpdateShopRequest) error
	Delete(id, userID uint) error
	GetByID(id uint) (*dto.ShopInfo, error)
	GetByUserID(userID uint) (*dto.ShopInfo, error)
	List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	AdminList(req *dto.ShopAdminListRequest) (*utils.Pagination, []dto.ShopInfo, error)
	Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error)
	ListByCategory(regionID, categoryID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error)

	UpdateStatus(id uint, status int) error
	Audit(id uint, auditStatus int, reason string) error
	UpdatePromotion(id uint, req *dto.ShopPromotionRequest) error
	IncrViewCount(id uint) error
}

type shopService struct {
	repo repository.ShopRepository
}

// NewShopService 创建店铺 service 实例
func NewShopService(repo repository.ShopRepository) ShopService {
	return &shopService{repo: repo}
}

func shopStatusText(s int) string {
	switch s {
	case model.ShopStatusDraft:
		return "草稿"
	case model.ShopStatusOpened:
		return "已开业"
	case model.ShopStatusClosed:
		return "已关闭"
	case model.ShopStatusFrozen:
		return "已冻结"
	case model.ShopStatusExpired:
		return "已过期"
	}
	return ""
}

func shopTypeText(t string) string {
	switch t {
	case model.ShopTypePersonal:
		return "个人店铺"
	case model.ShopTypeEnterprise:
		return "企业店铺"
	case model.ShopTypeFlagship:
		return "旗舰店"
	}
	return ""
}

func shopAuditStatusText(s int) string {
	switch s {
	case 0:
		return "待审核"
	case 1:
		return "已通过"
	case 2:
		return "已拒绝"
	}
	return ""
}

func toShopInfo(s *model.Shop) *dto.ShopInfo {
	info := &dto.ShopInfo{
		ID:             s.ID,
		UserID:         s.UserID,
		ShopName:       s.ShopName,
		Logo:           s.Logo,
		Description:    s.Description,
		ShopType:       s.ShopType,
		ShopTypeText:   shopTypeText(s.ShopType),
		Status:         s.Status,
		StatusText:     shopStatusText(s.Status),
		AuditStatus:    s.AuditStatus,
		AuditStatusText: shopAuditStatusText(s.AuditStatus),
		OpenedAt:       s.OpenedAt,
		ContactName:    s.ContactName,
		ContactPhone:   s.ContactPhone,
		ContactEmail:   s.ContactEmail,
		Wechat:         s.Wechat,
		QQ:             s.QQ,
		Province:       s.Province,
		City:           s.City,
		District:       s.District,
		Address:        s.Address,
		Latitude:       s.Latitude,
		Longitude:      s.Longitude,
		LicenseNo:      s.LicenseNo,
		LicenseImage:   s.LicenseImage,
		LegalPerson:    s.LegalPerson,
		LegalPersonID:  s.LegalPersonID,
		ProductCount:   s.ProductCount,
		OrderCount:     s.OrderCount,
		SaleAmount:     s.SaleAmount,
		Rating:         s.Rating,
		ReviewCount:    s.ReviewCount,
		FavoriteCount:  s.FavoriteCount,
		ViewCount:      s.ViewCount,
		Featured:       s.Featured,
		Verified:       s.Verified,
		PromotionLevel: s.PromotionLevel,
		TrafficWeight:  s.TrafficWeight,
		VerifiedAt:     s.VerifiedAt,
		RegionID:       s.RegionID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
	if s.Banners != nil {
		info.Banners = s.Banners
	}
	if s.Tags != nil {
		info.Tags = s.Tags
	}
	if s.BusinessHours != nil {
		info.BusinessHours = s.BusinessHours
	}
	if s.Facilities != nil {
		info.Facilities = s.Facilities
	}
	return info
}

// Create 创建店铺
func (s *shopService) Create(regionID, userID uint, userName string, req *dto.CreateShopRequest) (*dto.ShopInfo, error) {
	shop := &model.Shop{
		UserID:        userID,
		ShopName:      req.ShopName,
		Logo:          req.Logo,
		Description:   req.Description,
		ShopType:      req.ShopType,
		Status:        model.ShopStatusDraft,
		ContactName:   req.ContactName,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		Wechat:        req.Wechat,
		QQ:            req.QQ,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		LicenseNo:     req.LicenseNo,
		LicenseImage:  req.LicenseImage,
		LegalPerson:   req.LegalPerson,
		LegalPersonID: req.LegalPersonID,
	}
	shop.RegionID = regionID
	if shop.ShopType == "" {
		shop.ShopType = model.ShopTypePersonal
	}
	if req.Banners != nil {
		if b, err := model.FromJSON(req.Banners); err == nil {
			shop.Banners = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			shop.Tags = b
		}
	}
	if req.BusinessHours != nil {
		if b, err := model.FromJSON(req.BusinessHours); err == nil {
			shop.BusinessHours = b
		}
	}
	if req.Facilities != nil {
		if b, err := model.FromJSON(req.Facilities); err == nil {
			shop.Facilities = b
		}
	}

	if err := s.repo.Create(shop); err != nil {
		return nil, err
	}
	return toShopInfo(shop), nil
}

// Update 更新店铺
func (s *shopService) Update(id, userID uint, req *dto.UpdateShopRequest) error {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	if shop.UserID != userID {
		return errors.New("无权操作他人店铺")
	}

	fields := make(map[string]interface{})
	if req.ShopName != nil {
		fields["shop_name"] = *req.ShopName
	}
	if req.Logo != nil {
		fields["logo"] = *req.Logo
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.ShopType != nil {
		fields["shop_type"] = *req.ShopType
	}
	if req.ContactName != nil {
		fields["contact_name"] = *req.ContactName
	}
	if req.ContactPhone != nil {
		fields["contact_phone"] = *req.ContactPhone
	}
	if req.Province != nil {
		fields["province"] = *req.Province
	}
	if req.City != nil {
		fields["city"] = *req.City
	}
	if req.District != nil {
		fields["district"] = *req.District
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.Latitude != nil {
		fields["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		fields["longitude"] = *req.Longitude
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
	if req.ContactEmail != nil {
		fields["contact_email"] = *req.ContactEmail
	}
	if req.Wechat != nil {
		fields["wechat"] = *req.Wechat
	}
	if req.QQ != nil {
		fields["qq"] = *req.QQ
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Banners != nil {
		if b, err := model.FromJSON(req.Banners); err == nil {
			fields["banners"] = b
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = b
		}
	}
	if req.BusinessHours != nil {
		if b, err := model.FromJSON(req.BusinessHours); err == nil {
			fields["business_hours"] = b
		}
	}
	if req.Facilities != nil {
		if b, err := model.FromJSON(req.Facilities); err == nil {
			fields["facilities"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除店铺
func (s *shopService) Delete(id, userID uint) error {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShopNotFound
		}
		return err
	}
	if shop.UserID != userID {
		return errors.New("无权操作他人店铺")
	}
	return s.repo.Delete(id)
}

// GetByID 获取店铺详情
func (s *shopService) GetByID(id uint) (*dto.ShopInfo, error) {
	shop, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return toShopInfo(shop), nil
}

// GetByUserID 根据用户 ID 获取店铺
func (s *shopService) GetByUserID(userID uint) (*dto.ShopInfo, error) {
	shop, err := s.repo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}
	return toShopInfo(shop), nil
}

// List 店铺列表
func (s *shopService) List(regionID uint, req *dto.ShopListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ShopListOptions{
		ShopType: req.ShopType,
		Keyword:  req.Keyword,
		City:     req.City,
		District: req.District,
		Featured: req.Featured,
		Verified: req.Verified,
		Sort:     req.Sort,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toShopInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// AdminList 管理后台店铺列表
func (s *shopService) AdminList(req *dto.ShopAdminListRequest) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ShopAdminListOptions{
		RegionID:    req.RegionID,
		UserID:      req.UserID,
		ShopType:    req.ShopType,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Featured:    req.Featured,
		Verified:    req.Verified,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toShopInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// Search 搜索店铺
func (s *shopService) Search(regionID uint, keyword string, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.Search(regionID, pagination, keyword)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toShopInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByUser 按用户列出店铺
func (s *shopService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toShopInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByCategory 按分类列出店铺
func (s *shopService) ListByCategory(regionID, categoryID uint, page, pageSize int) (*utils.Pagination, []dto.ShopInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByCategory(regionID, categoryID, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ShopInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toShopInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStatus 更新店铺状态
func (s *shopService) UpdateStatus(id uint, status int) error {
	fields := map[string]interface{}{"status": status}
	if status == model.ShopStatusOpened {
		now := time.Now()
		fields["opened_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// Audit 审核店铺
func (s *shopService) Audit(id uint, auditStatus int, reason string) error {
	fields := map[string]interface{}{
		"audit_status": auditStatus,
	}
	if reason != "" {
		fields["audit_reason"] = reason
	}
	if auditStatus == 1 {
		now := time.Now()
		fields["verified"] = true
		fields["verified_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}

// UpdatePromotion 更新店铺推广配置
func (s *shopService) UpdatePromotion(id uint, req *dto.ShopPromotionRequest) error {
	fields := make(map[string]interface{})
	if req.Featured != nil {
		fields["featured"] = *req.Featured
	}
	if req.Verified != nil {
		fields["verified"] = *req.Verified
	}
	if req.PromotionLevel != nil {
		fields["promotion_level"] = *req.PromotionLevel
	}
	if req.TrafficWeight != nil {
		fields["traffic_weight"] = *req.TrafficWeight
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// IncrViewCount 增加浏览数
func (s *shopService) IncrViewCount(id uint) error {
	return s.repo.IncrViewCount(id)
}
