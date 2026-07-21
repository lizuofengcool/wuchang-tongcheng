// Package service LBS地图中台业务逻辑层 - 区域分站
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/lbs/dto"
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/modules/lbs/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrRegionNotFound     = errors.New("区域不存在")
	ErrRegionStatusInvalid = errors.New("区域状态不允许此操作")
	ErrRegionByLocation   = errors.New("无法根据经纬度判断分站")
)

// RegionService 区域业务接口
type RegionService interface {
	Create(req *dto.CreateRegionRequest) (*dto.RegionInfo, error)
	Update(id uint, req *dto.UpdateRegionRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.RegionInfo, error)
	List(req *dto.RegionListRequest) (*utils.Pagination, []dto.RegionInfo, error)
	ListByParent(parentID uint) ([]dto.RegionInfo, error)
	FindByCityCode(cityCode string) (*dto.RegionInfo, error)
	FindByLocation(lat, lng float64) (*dto.RegionQueryResponse, error)

	AdminUpdateStatus(id uint, status int) error
}

type regionService struct {
	repo repository.RegionRepository
}

// NewRegionService 创建区域 service 实例
func NewRegionService(repo repository.RegionRepository) RegionService {
	return &regionService{repo: repo}
}

// regionStatusText 状态文本
func regionStatusText(status int) string {
	switch status {
	case model.LBSRegionStatusDisabled:
		return "禁用"
	case model.LBSRegionStatusEnabled:
		return "启用"
	}
	return ""
}

// toRegionInfo model → dto
func toRegionInfo(r *model.Region) *dto.RegionInfo {
	info := &dto.RegionInfo{
		ID:          r.ID,
		Name:        r.Name,
		CityCode:    r.CityCode,
		ParentID:    r.ParentID,
		Level:       r.Level,
		Path:        r.Path,
		Sort:        r.Sort,
		Status:      r.Status,
		StatusText:  regionStatusText(r.Status),
		CenterLat:   r.CenterLat,
		CenterLng:   r.CenterLng,
		AdCode:      r.AdCode,
		ZipCode:     r.ZipCode,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.Boundary != nil {
		info.Boundary = r.Boundary
	}
	return info
}

// Create 创建区域
func (s *regionService) Create(req *dto.CreateRegionRequest) (*dto.RegionInfo, error) {
	r := &model.Region{
		Name:        req.Name,
		CityCode:    req.CityCode,
		ParentID:    req.ParentID,
		Level:       req.Level,
		Sort:        req.Sort,
		Status:      req.Status,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		AdCode:      req.AdCode,
		ZipCode:     req.ZipCode,
		Description: req.Description,
	}
	if r.Level == 0 {
		r.Level = 1
	}
	if r.Status == 0 {
		r.Status = model.LBSRegionStatusEnabled
	}
	if req.Boundary != nil {
		if b, err := model.FromJSON(req.Boundary); err == nil {
			r.Boundary = b
		}
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return toRegionInfo(r), nil
}

// Update 更新区域
func (s *regionService) Update(id uint, req *dto.UpdateRegionRequest) error {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegionNotFound
		}
		return err
	}
	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.CityCode != nil {
		r.CityCode = *req.CityCode
	}
	if req.ParentID != nil {
		r.ParentID = *req.ParentID
	}
	if req.Level != nil {
		r.Level = *req.Level
	}
	if req.Sort != nil {
		r.Sort = *req.Sort
	}
	if req.Status != nil {
		r.Status = *req.Status
	}
	if req.CenterLat != nil {
		r.CenterLat = *req.CenterLat
	}
	if req.CenterLng != nil {
		r.CenterLng = *req.CenterLng
	}
	if req.AdCode != nil {
		r.AdCode = *req.AdCode
	}
	if req.ZipCode != nil {
		r.ZipCode = *req.ZipCode
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.Boundary != nil {
		if b, err := model.FromJSON(req.Boundary); err == nil {
			r.Boundary = b
		}
	}
	return s.repo.Update(r)
}

// Delete 删除区域
func (s *regionService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegionNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取区域详情
func (s *regionService) GetByID(id uint) (*dto.RegionInfo, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegionNotFound
		}
		return nil, err
	}
	return toRegionInfo(r), nil
}

// List 区域列表
func (s *regionService) List(req *dto.RegionListRequest) (*utils.Pagination, []dto.RegionInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.RegionListOptions{
		Keyword:  req.Keyword,
		CityCode: req.CityCode,
		ParentID: req.ParentID,
		Level:    req.Level,
		Status:   req.Status,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	pagination.Total = total
	infos := make([]dto.RegionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRegionInfo(&list[i]))
	}
	return pagination, infos, nil
}

// ListByParent 按父级列出
func (s *regionService) ListByParent(parentID uint) ([]dto.RegionInfo, error) {
	list, err := s.repo.ListByParent(parentID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.RegionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toRegionInfo(&list[i]))
	}
	return infos, nil
}

// FindByCityCode 根据城市编码查找
func (s *regionService) FindByCityCode(cityCode string) (*dto.RegionInfo, error) {
	r, err := s.repo.FindByCityCode(cityCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegionNotFound
		}
		return nil, err
	}
	return toRegionInfo(r), nil
}

// FindByLocation 根据经纬度判断分站
// 简化方案：取中心点最近的启用区域；如需精确边界，使用 boundary 字段配合多边形算法
func (s *regionService) FindByLocation(lat, lng float64) (*dto.RegionQueryResponse, error) {
	r, err := s.repo.FindByLocation(lat, lng)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegionByLocation
		}
		return nil, err
	}
	return &dto.RegionQueryResponse{
		RegionID: r.ID,
		Name:     r.Name,
		CityCode: r.CityCode,
		AdCode:   r.AdCode,
		Inside:   true,
		Source:   "nearest_center",
	}, nil
}

// AdminUpdateStatus 管理后台更新状态
func (s *regionService) AdminUpdateStatus(id uint, status int) error {
	if status != model.LBSRegionStatusDisabled && status != model.LBSRegionStatusEnabled {
		return ErrRegionStatusInvalid
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}
