// Package service 同城114业务逻辑层 - 商户详情 + 营业时间 + 菜单
// 依据 v3.2.1 架构方案：对标大众点评/美团
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
	ErrBusinessNotFound       = errors.New("商户详情不存在")
	ErrBusinessDh114Exists    = errors.New("商户详情已存在")
	ErrMenuNotFound           = errors.New("菜单不存在")
	ErrBusinessHourNotFound   = errors.New("营业时间不存在")
)

// BusinessService 商户详情业务接口（含菜单 + 营业时间管理）
type BusinessService interface {
	// 商户详情 CRUD
	Create(regionID uint, userID uint, req *dto.CreateBusinessRequest) (*dto.BusinessInfo, error)
	Update(id uint, req *dto.UpdateBusinessRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.BusinessInfo, error)
	GetByDh114ID(dh114ID uint) (*dto.BusinessInfo, error)
	List(req *dto.BusinessListRequest) (*utils.Pagination, []dto.BusinessInfo, error)
	UpdateVerificationStatus(dh114ID uint, status int) error

	// 营业时间管理
	ListBusinessHours(dh114ID uint) ([]dto.BusinessHourInfo, error)
	ReplaceBusinessHours(dh114ID uint, req *dto.BatchReplaceHoursRequest) error

	// 菜单管理
	CreateMenu(regionID uint, userID uint, req *dto.CreateMenuRequest) (*dto.MenuInfo, error)
	UpdateMenu(id uint, req *dto.UpdateMenuRequest) error
	DeleteMenu(id uint) error
	GetMenuByID(id uint) (*dto.MenuInfo, error)
	ListMenus(req *dto.MenuListRequest) (*utils.Pagination, []dto.MenuInfo, error)
	ListMenusByDh114(dh114ID uint, onlyActive bool) ([]dto.MenuInfo, error)
	ListSignatureMenus(dh114ID uint) ([]dto.MenuInfo, error)
	ReplaceMenus(regionID uint, userID uint, req *dto.BatchReplaceMenusRequest) error
	IncrMenuOrderCount(id uint, count int) error
}

type businessService struct {
	repo     repository.BusinessRepository
	hourRepo repository.BusinessHourRepository
	menuRepo repository.MenuRepository
}

// NewBusinessService 创建商户详情 service 实例
func NewBusinessService(
	repo repository.BusinessRepository,
	hourRepo repository.BusinessHourRepository,
	menuRepo repository.MenuRepository,
) BusinessService {
	return &businessService{
		repo:     repo,
		hourRepo: hourRepo,
		menuRepo: menuRepo,
	}
}

// verificationStatusText 认证状态文本
func verificationStatusText(s int) string {
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

// toBusinessInfo model -> dto
func toBusinessInfo(b *model.Dh114Business) *dto.BusinessInfo {
	info := &dto.BusinessInfo{
		ID:                  b.ID,
		Dh114ID:             b.Dh114ID,
		BusinessName:        b.BusinessName,
		LicenseNo:           b.LicenseNo,
		LicenseImage:        b.LicenseImage,
		LegalPerson:         b.LegalPerson,
		BusinessScope:       b.BusinessScope,
		RegisteredCapital:  b.RegisteredCapital,
		EstablishedDate:    b.EstablishedDate,
		RegisteredAddress:  b.RegisteredAddress,
		OpeningHours:        b.OpeningHours,
		ClosingHours:        b.ClosingHours,
		OpenAllDay:         b.OpenAllDay,
		PriceAvg:           b.PriceAvg,
		PriceRangeMin:      b.PriceRangeMin,
		PriceRangeMax:      b.PriceRangeMax,
		Website:            b.Website,
		Wechat:             b.Wechat,
		WechatQR:           b.WechatQR,
		Email:              b.Email,
		VerificationStatus: b.VerificationStatus,
		VerificationStatusText: verificationStatusText(b.VerificationStatus),
		VerifiedAt:         b.VerifiedAt,
		ValidUntil:         b.ValidUntil,
		RegionID:           b.RegionID,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
	if b.ClosedDays != nil {
		info.ClosedDays = b.ClosedDays
	}
	if b.Facilities != nil {
		info.Facilities = b.Facilities
	}
	return info
}

// Create 创建商户详情
func (s *businessService) Create(regionID uint, userID uint, req *dto.CreateBusinessRequest) (*dto.BusinessInfo, error) {
	// 检查是否已存在
	if existing, err := s.repo.FindByDh114ID(req.Dh114ID); err == nil && existing != nil {
		return nil, ErrBusinessDh114Exists
	}

	b := &model.Dh114Business{
		Dh114ID:            req.Dh114ID,
		BusinessName:       req.BusinessName,
		LicenseNo:          req.LicenseNo,
		LicenseImage:       req.LicenseImage,
		LegalPerson:        req.LegalPerson,
		BusinessScope:      req.BusinessScope,
		RegisteredCapital:  req.RegisteredCapital,
		EstablishedDate:    req.EstablishedDate,
		RegisteredAddress:  req.RegisteredAddress,
		OpeningHours:       req.OpeningHours,
		ClosingHours:       req.ClosingHours,
		OpenAllDay:         req.OpenAllDay,
		PriceAvg:           req.PriceAvg,
		PriceRangeMin:      req.PriceRangeMin,
		PriceRangeMax:      req.PriceRangeMax,
		Website:            req.Website,
		Wechat:             req.Wechat,
		WechatQR:           req.WechatQR,
		Email:              req.Email,
		VerificationStatus: model.VerificationStatusPending,
	}
	b.RegionID = regionID

	// JSONB 字段处理
	if req.ClosedDays != nil {
		if b2, err := model.FromJSON(req.ClosedDays); err == nil {
			b.ClosedDays = b2
		}
	}
	if req.Facilities != nil {
		if b2, err := model.FromJSON(req.Facilities); err == nil {
			b.Facilities = b2
		}
	}

	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return toBusinessInfo(b), nil
}

// Update 更新商户详情
func (s *businessService) Update(id uint, req *dto.UpdateBusinessRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBusinessNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.BusinessName != nil {
		fields["business_name"] = *req.BusinessName
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
	if req.BusinessScope != nil {
		fields["business_scope"] = *req.BusinessScope
	}
	if req.RegisteredCapital != nil {
		fields["registered_capital"] = *req.RegisteredCapital
	}
	if req.EstablishedDate != nil {
		fields["established_date"] = *req.EstablishedDate
	}
	if req.RegisteredAddress != nil {
		fields["registered_address"] = *req.RegisteredAddress
	}
	if req.OpeningHours != nil {
		fields["opening_hours"] = *req.OpeningHours
	}
	if req.ClosingHours != nil {
		fields["closing_hours"] = *req.ClosingHours
	}
	if req.OpenAllDay != nil {
		fields["open_all_day"] = *req.OpenAllDay
	}
	if req.PriceAvg != nil {
		fields["price_avg"] = *req.PriceAvg
	}
	if req.PriceRangeMin != nil {
		fields["price_range_min"] = *req.PriceRangeMin
	}
	if req.PriceRangeMax != nil {
		fields["price_range_max"] = *req.PriceRangeMax
	}
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	if req.Wechat != nil {
		fields["wechat"] = *req.Wechat
	}
	if req.WechatQR != nil {
		fields["wechat_qr"] = *req.WechatQR
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.ClosedDays != nil {
		if b, err := model.FromJSON(req.ClosedDays); err == nil {
			fields["closed_days"] = b
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
	return s.repo.Update(id, fields)
}

// Delete 删除商户详情
func (s *businessService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBusinessNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取商户详情
func (s *businessService) GetByID(id uint) (*dto.BusinessInfo, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessNotFound
		}
		return nil, err
	}
	return toBusinessInfo(b), nil
}

// GetByDh114ID 按商户 ID 获取详情
func (s *businessService) GetByDh114ID(dh114ID uint) (*dto.BusinessInfo, error) {
	b, err := s.repo.FindByDh114ID(dh114ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBusinessNotFound
		}
		return nil, err
	}
	return toBusinessInfo(b), nil
}

// List 商户详情列表
func (s *businessService) List(req *dto.BusinessListRequest) (*utils.Pagination, []dto.BusinessInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.BusinessListQuery{
		Dh114ID:            req.Dh114ID,
		VerificationStatus: req.VerificationStatus,
		Keyword:            req.Keyword,
	}
	list, total, err := s.repo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.BusinessInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toBusinessInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateVerificationStatus 更新认证状态
func (s *businessService) UpdateVerificationStatus(dh114ID uint, status int) error {
	var verifiedAt interface{}
	if status == model.VerificationStatusApproved {
		now := time.Now()
		verifiedAt = &now
	}
	return s.repo.UpdateVerificationStatus(dh114ID, status, verifiedAt)
}

// ===== 营业时间管理 =====

// weekdayText 星期文本
func weekdayText(weekday int) string {
	names := []string{"", "周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	if weekday >= 1 && weekday <= 7 {
		return names[weekday]
	}
	return ""
}

// toBusinessHourInfo model -> dto
func toBusinessHourInfo(h *model.Dh114BusinessHour) *dto.BusinessHourInfo {
	return &dto.BusinessHourInfo{
		ID:          h.ID,
		Dh114ID:     h.Dh114ID,
		BusinessID:  h.BusinessID,
		Weekday:     h.Weekday,
		WeekdayText: weekdayText(h.Weekday),
		OpenTime:    h.OpenTime,
		CloseTime:   h.CloseTime,
		IsOpen:      h.IsOpen,
		Is24H:       h.Is24H,
	}
}

// ListBusinessHours 列出商户的营业时间
func (s *businessService) ListBusinessHours(dh114ID uint) ([]dto.BusinessHourInfo, error) {
	list, err := s.hourRepo.FindByDh114ID(dh114ID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.BusinessHourInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toBusinessHourInfo(&list[i]))
	}
	return infos, nil
}

// ReplaceBusinessHours 批量替换营业时间
func (s *businessService) ReplaceBusinessHours(dh114ID uint, req *dto.BatchReplaceHoursRequest) error {
	hours := make([]model.Dh114BusinessHour, 0, len(req.Hours))
	for _, h := range req.Hours {
		openTime := h.OpenTime
		closeTime := h.CloseTime
		if h.Is24H {
			openTime = "00:00"
			closeTime = "23:59"
		} else {
			if openTime == "" {
				openTime = "09:00"
			}
			if closeTime == "" {
				closeTime = "22:00"
			}
		}
		hours = append(hours, model.Dh114BusinessHour{
			Dh114ID:   dh114ID,
			Weekday:    h.Weekday,
			OpenTime:   openTime,
			CloseTime:  closeTime,
			IsOpen:     h.IsOpen,
			Is24H:      h.Is24H,
		})
	}
	return s.hourRepo.ReplaceHours(dh114ID, hours)
}

// ===== 菜单管理 =====

// menuStatusText 菜单状态文本
func menuStatusText(status int) string {
	switch status {
	case 0:
		return "下架"
	case 1:
		return "上架"
	}
	return ""
}

// menuTypeText 菜单类型文本
func menuTypeText(t string) string {
	switch t {
	case model.MenuTypeDish:
		return "菜品"
	case model.MenuTypeService:
		return "服务项目"
	}
	return ""
}

// toMenuInfo model -> dto
func toMenuInfo(m *model.Dh114Menu) *dto.MenuInfo {
	info := &dto.MenuInfo{
		ID:            m.ID,
		Dh114ID:       m.Dh114ID,
		BusinessID:    m.BusinessID,
		MenuType:      m.MenuType,
		MenuTypeText:  menuTypeText(m.MenuType),
		Name:          m.Name,
		Description:   m.Description,
		Price:         m.Price,
		OriginalPrice: m.OriginalPrice,
		Image:         m.Image,
		Unit:          m.Unit,
		Sort:          m.Sort,
		Status:        m.Status,
		StatusText:    menuStatusText(m.Status),
		OrderCount:    m.OrderCount,
		IsSignature:   m.IsSignature,
		RegionID:      m.RegionID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.Tags != nil {
		info.Tags = m.Tags
	}
	return info
}

// generateMenuNo 生成菜单相关业务单号
func generateMenuNo(prefix string) string {
	return fmt.Sprintf("%s%s%06d", prefix, time.Now().Format("20060102150405"), rand.Intn(1000000))
}

// CreateMenu 创建菜单
func (s *businessService) CreateMenu(regionID uint, userID uint, req *dto.CreateMenuRequest) (*dto.MenuInfo, error) {
	menuType := req.MenuType
	if menuType == "" {
		menuType = model.MenuTypeDish
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	m := &model.Dh114Menu{
		Dh114ID:       req.Dh114ID,
		MenuType:      menuType,
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		OriginalPrice:  req.OriginalPrice,
		Image:          req.Image,
		Unit:           req.Unit,
		Sort:           req.Sort,
		Status:         status,
		IsSignature:    req.IsSignature,
	}
	m.RegionID = regionID
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			m.Tags = b
		}
	}
	if err := s.menuRepo.Create(m); err != nil {
		return nil, err
	}
	return toMenuInfo(m), nil
}

// UpdateMenu 更新菜单
func (s *businessService) UpdateMenu(id uint, req *dto.UpdateMenuRequest) error {
	if _, err := s.menuRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.MenuType != nil {
		fields["menu_type"] = *req.MenuType
	}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Price != nil {
		fields["price"] = *req.Price
	}
	if req.OriginalPrice != nil {
		fields["original_price"] = *req.OriginalPrice
	}
	if req.Image != nil {
		fields["image"] = *req.Image
	}
	if req.Unit != nil {
		fields["unit"] = *req.Unit
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.IsSignature != nil {
		fields["is_signature"] = *req.IsSignature
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			fields["tags"] = b
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.menuRepo.Update(id, fields)
}

// DeleteMenu 删除菜单
func (s *businessService) DeleteMenu(id uint) error {
	if _, err := s.menuRepo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}
	return s.menuRepo.Delete(id)
}

// GetMenuByID 获取菜单详情
func (s *businessService) GetMenuByID(id uint) (*dto.MenuInfo, error) {
	m, err := s.menuRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}
	return toMenuInfo(m), nil
}

// ListMenus 菜单列表
func (s *businessService) ListMenus(req *dto.MenuListRequest) (*utils.Pagination, []dto.MenuInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.MenuListQuery{
		Dh114ID:     req.Dh114ID,
		MenuType:    req.MenuType,
		Status:      req.Status,
		IsSignature: req.IsSignature,
		Keyword:     req.Keyword,
	}
	list, total, err := s.menuRepo.List(query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.MenuInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toMenuInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListMenusByDh114 按商户 ID 列出菜单
func (s *businessService) ListMenusByDh114(dh114ID uint, onlyActive bool) ([]dto.MenuInfo, error) {
	list, err := s.menuRepo.ListByDh114(dh114ID, onlyActive)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.MenuInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toMenuInfo(&list[i]))
	}
	return infos, nil
}

// ListSignatureMenus 列出招牌菜单
func (s *businessService) ListSignatureMenus(dh114ID uint) ([]dto.MenuInfo, error) {
	list, err := s.menuRepo.ListSignature(dh114ID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.MenuInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toMenuInfo(&list[i]))
	}
	return infos, nil
}

// ReplaceMenus 批量替换菜单
func (s *businessService) ReplaceMenus(regionID uint, userID uint, req *dto.BatchReplaceMenusRequest) error {
	menus := make([]model.Dh114Menu, 0, len(req.Menus))
	for _, m := range req.Menus {
		menuType := m.MenuType
		if menuType == "" {
			menuType = model.MenuTypeDish
		}
		status := m.Status
		if status == 0 {
			status = 1
		}
		item := model.Dh114Menu{
			Dh114ID:       req.Dh114ID,
			MenuType:      menuType,
			Name:          m.Name,
			Description:   m.Description,
			Price:         m.Price,
			OriginalPrice: m.OriginalPrice,
			Image:         m.Image,
			Unit:          m.Unit,
			Sort:          m.Sort,
			Status:        status,
			IsSignature:   m.IsSignature,
		}
		item.RegionID = regionID
		if m.Tags != nil {
			if b, err := model.FromJSON(m.Tags); err == nil {
				item.Tags = b
			}
		}
		menus = append(menus, item)
	}
	return s.menuRepo.ReplaceMenus(req.Dh114ID, menus)
}

// IncrMenuOrderCount 增加菜单销量
func (s *businessService) IncrMenuOrderCount(id uint, count int) error {
	return s.menuRepo.IncrOrderCount(id, count)
}

// _ 保留单号生成函数引用以避免 unused 警告（菜单业务后续可能用单号）
var _ = generateMenuNo
