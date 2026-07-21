// Package service LBS地图中台业务逻辑层 - POI 兴趣点
// 依据 v3.2.1 架构方案：对标高德/百度地图 POI 服务
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package service

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/modules/lbs/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrPOINotFound     = errors.New("POI不存在")
	ErrPOINoPermission = errors.New("无权操作此POI")
	ErrPOIStatusInvalid = errors.New("POI状态不允许此操作")
)

// POIService POI 业务接口
type POIService interface {
	// C 端
	Create(regionID uint, userID uint, req *dto.CreatePOIRequest) (*dto.POIInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdatePOIRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.POIInfo, error)
	List(regionID uint, req *dto.POIListRequest) (*utils.Pagination, []dto.POIInfo, error)
	ListMine(userID uint, req *dto.POIListRequest) (*utils.Pagination, []dto.POIInfo, error)
	ListNearby(regionID uint, req *dto.NearbyRequest) (*utils.Pagination, []dto.POIInfo, error)

	// M 端管理
	AdminList(req *dto.POIAdminListRequest) (*utils.Pagination, []dto.POIInfo, error)
	AdminUpdateStatus(id uint, status int) error
}

type poiService struct {
	repo repository.POIRepository
}

// NewPOIService 创建 POI service 实例
func NewPOIService(repo repository.POIRepository) POIService {
	return &poiService{repo: repo}
}

// poiStatusText 状态文本
func poiStatusText(status int) string {
	switch status {
	case model.LBSPoiStatusOffline:
		return "下架"
	case model.LBSPoiStatusOnline:
		return "上线"
	case model.LBSPoiStatusPending:
		return "待审"
	case model.LBSPoiStatusRejected:
		return "拒绝"
	case model.LBSPoiStatusDeleted:
		return "删除"
	}
	return ""
}

// toPOIInfo model → dto
func toPOIInfo(p *model.POI) *dto.POIInfo {
	info := &dto.POIInfo{
		ID:          p.ID,
		Name:        p.Name,
		Address:     p.Address,
		Category:    p.Category,
		Phone:       p.Phone,
		Icon:        p.Icon,
		Status:      p.Status,
		StatusText:  poiStatusText(p.Status),
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Distance:    p.Distance,
		UserID:      p.UserID,
		Source:      p.Source,
		ExternalID:  p.ExternalID,
		PublishedAt: p.PublishedAt,
		RegionID:    p.RegionID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.Tags != nil {
		info.Tags = p.Tags
	}
	if p.Extra != nil {
		info.Extra = p.Extra
	}
	return info
}

// Create 创建 POI
func (s *poiService) Create(regionID uint, userID uint, req *dto.CreatePOIRequest) (*dto.POIInfo, error) {
	p := &model.POI{
		Name:       req.Name,
		Address:    req.Address,
		Category:   req.Category,
		Phone:      req.Phone,
		Icon:       req.Icon,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		UserID:     userID,
		Source:     req.Source,
		ExternalID: req.ExternalID,
		Status:     req.Status,
	}
	if p.Source == "" {
		p.Source = "manual"
	}
	if p.Status == 0 {
		p.Status = model.LBSPoiStatusOnline
	}
	p.RegionID = regionID

	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			p.Tags = b
		}
	}
	if req.Extra != nil {
		if b, err := model.FromJSON(req.Extra); err == nil {
			p.Extra = b
		}
	}

	now := time.Now()
	if p.Status == model.LBSPoiStatusOnline {
		p.PublishedAt = &now
	}

	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return toPOIInfo(p), nil
}

// Update 更新 POI
func (s *poiService) Update(id uint, operatorID uint, req *dto.UpdatePOIRequest) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPOINotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPOINoPermission
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Address != nil {
		p.Address = *req.Address
	}
	if req.Category != nil {
		p.Category = *req.Category
	}
	if req.Phone != nil {
		p.Phone = *req.Phone
	}
	if req.Icon != nil {
		p.Icon = *req.Icon
	}
	if req.Latitude != nil {
		p.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		p.Longitude = *req.Longitude
	}
	if req.Source != nil {
		p.Source = *req.Source
	}
	if req.ExternalID != nil {
		p.ExternalID = *req.ExternalID
	}
	if req.Status != nil {
		p.Status = *req.Status
		if p.Status == model.LBSPoiStatusOnline && p.PublishedAt == nil {
			now := time.Now()
			p.PublishedAt = &now
		}
	}
	if req.Tags != nil {
		if b, err := model.FromJSON(req.Tags); err == nil {
			p.Tags = b
		}
	}
	if req.Extra != nil {
		if b, err := model.FromJSON(req.Extra); err == nil {
			p.Extra = b
		}
	}

	return s.repo.Update(p)
}

// Delete 删除 POI
func (s *poiService) Delete(id uint, operatorID uint) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPOINotFound
		}
		return err
	}
	if p.UserID != operatorID {
		return ErrPOINoPermission
	}
	return s.repo.Delete(id)
}

// GetByID 获取 POI 详情
func (s *poiService) GetByID(id uint) (*dto.POIInfo, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPOINotFound
		}
		return nil, err
	}
	return toPOIInfo(p), nil
}

// List POI 列表
func (s *poiService) List(regionID uint, req *dto.POIListRequest) (*utils.Pagination, []dto.POIInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.POIListOptions{
		Category: req.Category,
		Keyword:  req.Keyword,
		Source:    req.Source,
		Sort:      req.Sort,
	}
	if req.Status != nil {
		opts.Status = *req.Status
	}
	if req.UserID > 0 {
		// C 端按用户筛选
		list, total, err := s.repo.ListByUser(req.UserID, pagination)
		if err != nil {
			return nil, nil, err
		}
		pagination.Total = total
		infos := make([]dto.POIInfo, 0, len(list))
		for i := range list {
			infos = append(infos, *toPOIInfo(&list[i]))
		}
		return pagination, infos, nil
	}
	list, total, err := s.repo.List(regionID, pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.POIInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPOIInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListMine 我的 POI 列表
func (s *poiService) ListMine(userID uint, req *dto.POIListRequest) (*utils.Pagination, []dto.POIInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	list, total, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.POIInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPOIInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListNearby 附近 POI 检索
func (s *poiService) ListNearby(regionID uint, req *dto.NearbyRequest) (*utils.Pagination, []dto.POIInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.POIListOptions{
		Category: req.Category,
		Keyword:  req.Keyword,
		Sort:     req.Sort,
		Status:   model.LBSPoiStatusOnline,
	}
	list, total, err := s.repo.ListNearby(regionID, pagination, req.Latitude, req.Longitude, req.RadiusKm, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.POIInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPOIInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminList 管理后台列表
func (s *poiService) AdminList(req *dto.POIAdminListRequest) (*utils.Pagination, []dto.POIInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.POIAdminListOptions{
		RegionID: req.RegionID,
		UserID:   req.UserID,
		Category: req.Category,
		Source:   req.Source,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.AdminList(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.POIInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toPOIInfo(&list[i]))
	}
	return pagination, infos, nil
}

// AdminUpdateStatus 管理后台更新状态
func (s *poiService) AdminUpdateStatus(id uint, status int) error {
	if status != model.LBSPoiStatusOffline && status != model.LBSPoiStatusOnline &&
		status != model.LBSPoiStatusPending && status != model.LBSPoiStatusRejected {
		return ErrPOIStatusInvalid
	}
	fields := map[string]interface{}{"status": status}
	if status == model.LBSPoiStatusOnline {
		now := time.Now()
		fields["published_at"] = &now
	}
	return s.repo.UpdateFields(id, fields)
}
