// Package service 小区信息业务逻辑层
// 依据 v3.2.1 架构方案第五章：对标贝壳/链家
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/house/dto"
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/modules/house/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrCommunityNotFound     = errors.New("小区不存在")
	ErrCommunityNoPermission = errors.New("无权操作此小区")
	ErrCommunityExists       = errors.New("小区已存在")
)

// CommunityService 小区业务接口
type CommunityService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CommunityCreateRequest) (*dto.CommunityResponse, error)
	Update(id uint, operatorID uint, req *dto.CommunityCreateRequest) error
	GetByID(id uint, userID uint) (*dto.CommunityResponse, error)
	List(regionID uint, req *dto.CommunityListQuery) (*utils.Pagination, []dto.CommunityResponse, error)
	ListNearby(regionID uint, req *dto.CommunityListQuery) (*utils.Pagination, []dto.CommunityResponse, error)

	// 关注
	Follow(userID, communityID uint, notify bool) (*dto.FavResponse, error)
	FollowStatus(userID, communityID uint) (*dto.FavResponse, error)

	// M 端
	AdminList(req *dto.CommunityAdminListQuery) (*utils.Pagination, []dto.CommunityResponse, error)
	UpdateStatus(id uint, status int) error
}

type communityService struct {
	repo repository.CommunityRepository
}

// NewCommunityService 创建 service 实例
func NewCommunityService(repo repository.CommunityRepository) CommunityService {
	return &communityService{repo: repo}
}

// toCommunityInfo model -> dto
func toCommunityInfo(c *model.HouseCommunity) *dto.CommunityResponse {
	info := &dto.CommunityResponse{
		ID:               c.ID,
		Name:             c.Name,
		Alias:            c.Alias,
		City:             c.City,
		District:         c.District,
		BusinessDistrict: c.BusinessDistrict,
		Address:          c.Address,
		Latitude:         c.Latitude,
		Longitude:        c.Longitude,
		BuildingCount:    c.BuildingCount,
		HouseCount:       c.HouseCount,
		BuildingYear:     c.BuildingYear,
		BuildingType:     c.BuildingType,
		Developer:        c.Developer,
		PropertyCompany:  c.PropertyCompany,
		PropertyFee:      c.PropertyFee,
		ParkingRatio:     c.ParkingRatio,
		GreeningRate:     c.GreeningRate,
		PlotRatio:        c.PlotRatio,
		AvgSalePrice:     c.AvgSalePrice,
		AvgRentPrice:     c.AvgRentPrice,
		Description:      c.Description,
		CoverImage:       c.CoverImage,
		Status:           c.Status,
		StatusText:       communityStatusText(c.Status),
		FollowerCount:    c.FollowerCount,
		OnSaleCount:      c.OnSaleCount,
		OnRentCount:      c.OnRentCount,
		RegionID:         c.RegionID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
	return info
}

func communityStatusText(s int) string {
	switch s {
	case model.CommunityStatusDraft:
		return "草稿"
	case model.CommunityStatusPublished:
		return "已发布"
	case model.CommunityStatusOffline:
		return "已下架"
	}
	return "草稿"
}

// ===== C 端 =====

func (s *communityService) Create(regionID uint, userID uint, req *dto.CommunityCreateRequest) (*dto.CommunityResponse, error) {
	// 校验是否已存在
	if existing, err := s.repo.FindByName(req.Name, req.City); err == nil && existing != nil {
		return nil, ErrCommunityExists
	}

	c := &model.HouseCommunity{
		Name:             req.Name,
		Alias:            req.Alias,
		City:             req.City,
		District:         req.District,
		BusinessDistrict: req.BusinessDistrict,
		Address:          req.Address,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		BuildingCount:    req.BuildingCount,
		HouseCount:       req.HouseCount,
		BuildingYear:     req.BuildingYear,
		BuildingType:     req.BuildingType,
		Developer:        req.Developer,
		PropertyCompany:  req.PropertyCompany,
		PropertyFee:      req.PropertyFee,
		ParkingRatio:     req.ParkingRatio,
		GreeningRate:     req.GreeningRate,
		PlotRatio:        req.PlotRatio,
		AvgSalePrice:     req.AvgSalePrice,
		AvgRentPrice:     req.AvgRentPrice,
		Description:      req.Description,
		CoverImage:       req.CoverImage,
		Status:           req.Status,
	}
	if c.Status == 0 {
		c.Status = model.CommunityStatusPublished
	}
	c.RegionID = regionID

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return toCommunityInfo(c), nil
}

func (s *communityService) Update(id uint, operatorID uint, req *dto.CommunityCreateRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommunityNotFound
		}
		return err
	}
	fields := map[string]interface{}{
		"name":             req.Name,
		"alias":            req.Alias,
		"city":             req.City,
		"district":         req.District,
		"business_district": req.BusinessDistrict,
		"address":          req.Address,
		"latitude":         req.Latitude,
		"longitude":        req.Longitude,
		"building_count":   req.BuildingCount,
		"house_count":      req.HouseCount,
		"building_year":    req.BuildingYear,
		"building_type":    req.BuildingType,
		"developer":        req.Developer,
		"property_company": req.PropertyCompany,
		"property_fee":     req.PropertyFee,
		"parking_ratio":    req.ParkingRatio,
		"greening_rate":    req.GreeningRate,
		"plot_ratio":       req.PlotRatio,
		"avg_sale_price":   req.AvgSalePrice,
		"avg_rent_price":   req.AvgRentPrice,
		"description":      req.Description,
		"cover_image":      req.CoverImage,
		"status":           req.Status,
	}
	return s.repo.UpdateFields(id, fields)
}

func (s *communityService) GetByID(id uint, userID uint) (*dto.CommunityResponse, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}
	info := toCommunityInfo(c)
	if userID > 0 {
		if followed, err := s.repo.FollowExists(userID, id); err == nil {
			info.HasFollowed = followed
		}
	}
	return info, nil
}

func (s *communityService) List(regionID uint, req *dto.CommunityListQuery) (*utils.Pagination, []dto.CommunityResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CommunityListOptions{
		City:             req.City,
		District:         req.District,
		BusinessDistrict: req.BusinessDistrict,
		BuildingType:     req.BuildingType,
		Keyword:          req.Keyword,
		Sort:             req.Sort,
	}

	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CommunityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCommunityInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *communityService) ListNearby(regionID uint, req *dto.CommunityListQuery) (*utils.Pagination, []dto.CommunityResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}

	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, radiusKm)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CommunityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCommunityInfo(&list[i]))
	}
	return pagination, result, nil
}

// ===== 关注 =====

func (s *communityService) Follow(userID, communityID uint, notify bool) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrCommunityNoPermission
	}
	c, err := s.repo.FindByID(communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}

	exists, err := s.repo.FollowExists(userID, communityID)
	if err != nil {
		return nil, err
	}
	if exists {
		if err := s.repo.DeleteFollow(userID, communityID); err != nil {
			return nil, err
		}
		_ = s.repo.DecrFollowerCount(communityID)
		return &dto.FavResponse{HasFaved: false, FavCount: c.FollowerCount - 1}, nil
	}

	fav := &model.HouseFavorite{
		UserID:       userID,
		CommunityID:  communityID,
		FavoriteType: model.FavoriteTypeCommunity,
		Notify:       notify,
	}
	if err := s.repo.CreateFollow(fav); err != nil {
		return nil, err
	}
	_ = s.repo.IncrFollowerCount(communityID)
	return &dto.FavResponse{HasFaved: true, FavCount: c.FollowerCount + 1}, nil
}

func (s *communityService) FollowStatus(userID, communityID uint) (*dto.FavResponse, error) {
	c, err := s.repo.FindByID(communityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}
	if userID == 0 {
		return &dto.FavResponse{HasFaved: false, FavCount: c.FollowerCount}, nil
	}
	exists, err := s.repo.FollowExists(userID, communityID)
	if err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: exists, FavCount: c.FollowerCount}, nil
}

// ===== M 端 =====

func (s *communityService) AdminList(req *dto.CommunityAdminListQuery) (*utils.Pagination, []dto.CommunityResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.CommunityAdminListOptions{
		RegionID: req.RegionID,
		City:     req.City,
		District: req.District,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total

	result := make([]dto.CommunityResponse, 0, len(list))
	for i := range list {
		result = append(result, *toCommunityInfo(&list[i]))
	}
	return pagination, result, nil
}

func (s *communityService) UpdateStatus(id uint, status int) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommunityNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}
