// Package service 同城拼车出行业务逻辑层 - 路线 + 常用路线收藏
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/pinche/dto"
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/modules/pinche/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrRouteNotFound     = errors.New("路线不存在")
	ErrRouteNoPermission = errors.New("无权操作此路线")
	ErrFavNotFound       = errors.New("收藏不存在")
)

// RouteService 路线业务接口
type RouteService interface {
	Create(regionID uint, userID uint, req *dto.CreateRouteRequest) (*dto.RouteInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateRouteRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.RouteInfo, error)
	List(regionID uint, req *dto.RouteListRequest) (*utils.Pagination, []dto.RouteInfo, error)
	ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error)
	ListCommon(regionID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error)

	// 收藏
	FavRoute(userID uint, routeID uint) (*dto.FavResponse, error)
	ListFavRoutes(userID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error)
	IncrUseCount(id uint) error
}

type routeService struct {
	repo     repository.RouteRepository
	favRepo  repository.RouteFavoriteRepository
}

// NewRouteService 创建路线 service 实例
func NewRouteService(repo repository.RouteRepository, favRepo repository.RouteFavoriteRepository) RouteService {
	return &routeService{repo: repo, favRepo: favRepo}
}

// toRouteInfo model -> dto
func toRouteInfo(r *model.PincheRoute) *dto.RouteInfo {
	info := &dto.RouteInfo{
		ID:                 r.ID,
		RegionID:           r.RegionID,
		UserID:             r.UserID,
		RouteName:          r.RouteName,
		OriginAddress:      r.OriginAddress,
		OriginLat:          r.OriginLat,
		OriginLng:          r.OriginLng,
		DestinationAddress: r.DestinationAddress,
		DestinationLat:     r.DestinationLat,
		DestinationLng:     r.DestinationLng,
		DistanceKM:         r.DistanceKM,
		DurationMin:        r.DurationMin,
		EstimatedPrice:     r.EstimatedPrice,
		TollFee:            r.TollFee,
		IsCommon:           r.IsCommon,
		UseCount:           r.UseCount,
		Status:             r.Status,
	}
	if r.Waypoints != nil {
		info.Waypoints = r.Waypoints
	}
	return info
}

// Create 创建路线
func (s *routeService) Create(regionID uint, userID uint, req *dto.CreateRouteRequest) (*dto.RouteInfo, error) {
	r := &model.PincheRoute{
		UserID:             userID,
		RouteName:          req.RouteName,
		OriginAddress:      req.OriginAddress,
		OriginLat:          req.OriginLat,
		OriginLng:          req.OriginLng,
		DestinationAddress: req.DestinationAddress,
		DestinationLat:     req.DestinationLat,
		DestinationLng:     req.DestinationLng,
		DistanceKM:         req.DistanceKM,
		DurationMin:        req.DurationMin,
		EstimatedPrice:     req.EstimatedPrice,
		TollFee:            req.TollFee,
		IsCommon:           req.IsCommon,
		Status:             1,
	}
	r.RegionID = regionID
	if req.Waypoints != nil {
		if jb, err := model.FromJSON(req.Waypoints); err == nil {
			r.Waypoints = jb
		}
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toRouteInfo(r), nil
}

// Update 更新路线
func (s *routeService) Update(id uint, operatorID uint, req *dto.UpdateRouteRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRouteNotFound
		}
		return err
	}
	if r.UserID != operatorID {
		return ErrRouteNoPermission
	}
	fields := map[string]interface{}{}
	if req.RouteName != nil {
		fields["route_name"] = *req.RouteName
	}
	if req.OriginAddress != nil {
		fields["origin_address"] = *req.OriginAddress
	}
	if req.OriginLat != nil {
		fields["origin_lat"] = *req.OriginLat
	}
	if req.OriginLng != nil {
		fields["origin_lng"] = *req.OriginLng
	}
	if req.DestinationAddress != nil {
		fields["destination_address"] = *req.DestinationAddress
	}
	if req.DestinationLat != nil {
		fields["destination_lat"] = *req.DestinationLat
	}
	if req.DestinationLng != nil {
		fields["destination_lng"] = *req.DestinationLng
	}
	if req.DistanceKM != nil {
		fields["distance_km"] = *req.DistanceKM
	}
	if req.DurationMin != nil {
		fields["duration_min"] = *req.DurationMin
	}
	if req.EstimatedPrice != nil {
		fields["estimated_price"] = *req.EstimatedPrice
	}
	if req.TollFee != nil {
		fields["toll_fee"] = *req.TollFee
	}
	if req.IsCommon != nil {
		fields["is_common"] = *req.IsCommon
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Waypoints != nil {
		if jb, err := model.FromJSON(req.Waypoints); err == nil {
			fields["waypoints"] = jb
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除路线
func (s *routeService) Delete(id uint, operatorID uint) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRouteNotFound
		}
		return err
	}
	if r.UserID != operatorID {
		return ErrRouteNoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取详情
func (s *routeService) GetByID(id uint) (*dto.RouteInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRouteNotFound
		}
		return nil, err
	}
	return toRouteInfo(r), nil
}

// List 路线列表
func (s *routeService) List(regionID uint, req *dto.RouteListRequest) (*utils.Pagination, []dto.RouteInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RouteListOptions{
		UserID:   req.UserID,
		IsCommon: req.IsCommon,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RouteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRouteInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListByUser 用户路线列表
func (s *routeService) ListByUser(userID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RouteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRouteInfo(&list[i]))
	}
	return pagination, result, nil
}

// ListCommon 常用路线
func (s *routeService) ListCommon(regionID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.ListCommon(regionID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RouteInfo, 0, len(list))
	for i := range list {
		result = append(result, *toRouteInfo(&list[i]))
	}
	return pagination, result, nil
}

// FavRoute 收藏/取消收藏路线
func (s *routeService) FavRoute(userID uint, routeID uint) (*dto.FavResponse, error) {
	if userID == 0 {
		return nil, ErrRouteNoPermission
	}
	r, err := s.repo.FindByID(routeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRouteNotFound
		}
		return nil, err
	}
	// 检查是否已收藏
	favs, _, err := s.favRepo.ListByUser(userID, utils.NewPagination(1, 200))
	if err != nil {
		return nil, err
	}
	routeIDPtr := routeID
	for _, f := range favs {
		if f.RouteID != nil && *f.RouteID == routeID {
			// 已收藏，取消收藏
			if err := s.favRepo.Delete(f.ID); err != nil {
				return nil, err
			}
			return &dto.FavResponse{HasFaved: false, FavCount: r.UseCount}, nil
		}
	}
	// 创建收藏
	fav := &model.PincheRouteFavorite{
		UserID:    userID,
		RouteID:   &routeIDPtr,
	}
	fav.RegionID = r.RegionID
	if err := s.favRepo.Create(fav); err != nil {
		return nil, err
	}
	return &dto.FavResponse{HasFaved: true, FavCount: r.UseCount + 1}, nil
}

// ListFavRoutes 我的收藏路线
func (s *routeService) ListFavRoutes(userID uint, page, pageSize int) (*utils.Pagination, []dto.RouteInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	favs, total, err := s.favRepo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	result := make([]dto.RouteInfo, 0, len(favs))
	for _, f := range favs {
		if f.RouteID == nil {
			continue
		}
		r, err := s.repo.FindByID(*f.RouteID)
		if err != nil {
			continue
		}
		result = append(result, *toRouteInfo(r))
	}
	return pagination, result, nil
}

// IncrUseCount 增加使用次数
func (s *routeService) IncrUseCount(id uint) error {
	return s.repo.IncrUseCount(id)
}
