// Package service 房源发布业务逻辑层
// 依据 v3.2.1 架构方案第五章：与 houses 主表 1:1 冗余发布信息
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
	ErrListingNotFound     = errors.New("发布信息不存在")
	ErrListingNoPermission = errors.New("无权操作此发布")
)

// ListingService 发布业务接口
type ListingService interface {
	// C 端
	Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateListingRequest) (*dto.ListingInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateListingRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint, userID uint) (*dto.ListingInfo, error)
	List(regionID uint, req *dto.ListingListQuery) (*utils.Pagination, []dto.ListingInfo, error)
	ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ListingInfo, error)
	Refresh(id uint, operatorID uint) error

	// M 端
	AdminList(req *dto.ListingAdminListQuery) (*utils.Pagination, []dto.ListingInfo, error)
	Audit(id uint, auditStatus int, auditReason string) error
	UpdateStatus(id uint, status int) error
}

type listingService struct {
	repo repository.ListingRepository
}

// NewListingService 创建 service 实例
func NewListingService(repo repository.ListingRepository) ListingService {
	return &listingService{repo: repo}
}

// toListingInfo model -> dto
func toListingInfo(l *model.HouseListing) *dto.ListingInfo {
	info := &dto.ListingInfo{
		ID:              l.ID,
		ListingNo:       l.ListingNo,
		HouseID:         l.HouseID,
		CommunityID:     l.CommunityID,
		AgentID:         l.AgentID,
		PublisherID:     l.PublisherID,
		PublisherName:   l.PublisherName,
		PublisherPhone:  l.PublisherPhone,
		PublisherAvatar: l.PublisherAvatar,
		PublisherType:   l.PublisherType,
		ListingType:     l.ListingType,
		Title:           l.Title,
		Description:     l.Description,
		Price:           l.Price,
		PriceUnit:       l.PriceUnit,
		Decoration:      l.Decoration,
		Orientation:     l.Orientation,
		Layout:          l.Layout,
		BuildingArea:    l.BuildingArea,
		Status:          l.Status,
		StatusText:      listingStatusText(l.Status),
		AuditStatus:     l.AuditStatus,
		AuditReason:     l.AuditReason,
		PublishedAt:     l.PublishedAt,
		ExpiredAt:       l.ExpiredAt,
		RefreshedAt:     l.RefreshedAt,
		OfflineAt:       l.OfflineAt,
		RefreshCount:    l.RefreshCount,
		ViewCount:       l.ViewCount,
		FavCount:        l.FavCount,
		ContactCount:    l.ContactCount,
		RegionID:        l.RegionID,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	}
	return info
}

func listingStatusText(s int) string {
	switch s {
	case model.ListingStatusDraft:
		return "草稿"
	case model.ListingStatusPublished:
		return "已发布"
	case model.ListingStatusOffline:
		return "已下架"
	case model.ListingStatusExpired:
		return "已过期"
	case model.ListingStatusRejected:
		return "已拒绝"
	}
	return "草稿"
}

// generateListingNo 生成发布单号
func generateListingNo() string {
	return fmt.Sprintf("LS%s", time.Now().Format("20060102150405.000"))
}

// ===== C 端 =====

func (s *listingService) Create(regionID uint, userID uint, userName string, userPhone string, userAvatar string, req *dto.CreateListingRequest) (*dto.ListingInfo, error) {
	now := time.Now()
	l := &model.HouseListing{
		ListingNo:       generateListingNo(),
		HouseID:         req.HouseID,
		CommunityID:     req.CommunityID,
		AgentID:         req.AgentID,
		PublisherID:     userID,
		PublisherName:   userName,
		PublisherPhone:  userPhone,
		PublisherAvatar: userAvatar,
		PublisherType:   req.PublisherType,
		ListingType:     req.ListingType,
		Title:           req.Title,
		Description:     req.Description,
		Price:           req.Price,
		PriceUnit:       req.PriceUnit,
		Decoration:      req.Decoration,
		Orientation:     req.Orientation,
		Layout:          req.Layout,
		BuildingArea:    req.BuildingArea,
		Status:          req.Status,
		AuditStatus:     model.AuditApproved, // MVP：发布即通过
	}
	l.RegionID = regionID

	if l.PublisherType == "" {
		l.PublisherType = model.PublisherTypePersonal
	}
	if l.ListingType == "" {
		l.ListingType = model.ListingTypeRent
	}
	if l.PriceUnit == "" {
		l.PriceUnit = model.RentUnitMonth
	}
	if l.Decoration == "" {
		l.Decoration = model.DecorationRough
	}

	// 过期时间
	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 90
	}
	expiredAt := now.AddDate(0, 0, expireDays)
	l.ExpiredAt = &expiredAt

	if l.Status == model.ListingStatusPublished {
		l.PublishedAt = &now
	}

	if err := s.repo.Create(l); err != nil {
		return nil, err
	}
	return toListingInfo(l), nil
}

func (s *listingService) Update(id uint, operatorID uint, req *dto.UpdateListingRequest) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	if l.PublisherID != operatorID {
		return ErrListingNoPermission
	}

	fields := map[string]interface{}{}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.Price > 0 || req.Price == 0 {
		fields["price"] = req.Price
	}
	if req.PriceUnit != "" {
		fields["price_unit"] = req.PriceUnit
	}
	if req.Decoration != "" {
		fields["decoration"] = req.Decoration
	}
	if req.Orientation != "" {
		fields["orientation"] = req.Orientation
	}
	if req.Layout != "" {
		fields["layout"] = req.Layout
	}
	if req.BuildingArea > 0 || req.BuildingArea == 0 {
		fields["building_area"] = req.BuildingArea
	}
	if req.ExpireDays > 0 {
		expiredAt := time.Now().AddDate(0, 0, req.ExpireDays)
		fields["expired_at"] = &expiredAt
	}

	// 状态变更
	if req.Status == model.ListingStatusPublished && l.Status != model.ListingStatusPublished {
		now := time.Now()
		fields["status"] = model.ListingStatusPublished
		fields["published_at"] = &now
		fields["audit_status"] = model.AuditApproved
	} else if req.Status == model.ListingStatusDraft || req.Status == model.ListingStatusOffline || req.Status == model.ListingStatusExpired {
		fields["status"] = req.Status
		if req.Status == model.ListingStatusOffline {
			now := time.Now()
			fields["offline_at"] = &now
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *listingService) Delete(id uint, operatorID uint) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	if l.PublisherID != operatorID {
		return ErrListingNoPermission
	}
	return s.repo.Delete(id)
}

func (s *listingService) GetByID(id uint, userID uint) (*dto.ListingInfo, error) {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	_ = s.repo.IncrViewCount(id)
	l.ViewCount++
	return toListingInfo(l), nil
}

func (s *listingService) List(regionID uint, req *dto.ListingListQuery) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	published := model.ListingStatusPublished
	opts := repository.ListingListOptions{
		HouseID:       req.HouseID,
		CommunityID:   req.CommunityID,
		AgentID:       req.AgentID,
		PublisherID:   req.PublisherID,
		PublisherType: req.PublisherType,
		ListingType:   req.ListingType,
		Keyword:       req.Keyword,
		Sort:          req.Sort,
	}
	if req.Status != nil {
		opts.Status = req.Status
	} else {
		opts.Status = &published
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *listingService) ListMine(userID uint, page, pageSize int) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByPublisher(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

// Refresh 刷新发布（每天最多刷新若干次）
func (s *listingService) Refresh(id uint, operatorID uint) error {
	l, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	if l.PublisherID != operatorID {
		return ErrListingNoPermission
	}
	now := time.Now()
	return s.repo.UpdateFields(id, map[string]interface{}{
		"refreshed_at":  &now,
		"refresh_count": gorm.Expr("refresh_count + 1"),
	})
}

// ===== M 端 =====

func (s *listingService) AdminList(req *dto.ListingAdminListQuery) (*utils.Pagination, []dto.ListingInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ListingAdminListOptions{
		RegionID:    req.RegionID,
		HouseID:     req.HouseID,
		PublisherID: req.PublisherID,
		ListingType: req.ListingType,
		Status:      req.Status,
		AuditStatus: req.AuditStatus,
		Keyword:     req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.ListingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toListingInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *listingService) Audit(id uint, auditStatus int, auditReason string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	})
}

func (s *listingService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrListingNotFound
		}
		return err
	}
	fields := map[string]interface{}{"status": status}
	if status == model.ListingStatusPublished {
		now := time.Now()
		fields["published_at"] = &now
	}
	if status == model.ListingStatusOffline {
		now := time.Now()
		fields["offline_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}
