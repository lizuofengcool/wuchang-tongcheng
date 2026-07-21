// Package service 营销活动中台业务逻辑层 - 广告位（ad 子域）
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/marketing/dto"
	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/modules/marketing/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AdService 广告位业务接口
type AdService interface {
	Create(regionID uint, req *dto.CreateAdPositionRequest) (*dto.AdPositionInfo, error)
	Update(id uint, req *dto.UpdateAdPositionRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.AdPositionInfo, error)
	List(regionID uint, req *dto.AdPositionListRequest) (*utils.Pagination, []dto.AdPositionInfo, error)
	ListByPositionCode(regionID uint, positionCode string, page, pageSize int) (*utils.Pagination, []dto.AdPositionInfo, error)
}

type adService struct {
	repo repository.AdRepository
}

// NewAdService 创建广告位 service 实例
func NewAdService(repo repository.AdRepository) AdService {
	return &adService{repo: repo}
}

// adStatusText 广告位状态文本
func adStatusText(s int) string {
	switch s {
	case model.AdStatusDisabled:
		return "禁用"
	case model.AdStatusEnabled:
		return "启用"
	case model.AdStatusScheduled:
		return "待生效"
	case model.AdStatusExpired:
		return "已过期"
	}
	return ""
}

// toAdPositionInfo model -> dto
func toAdPositionInfo(a *model.AdPosition) *dto.AdPositionInfo {
	return &dto.AdPositionInfo{
		ID:           a.ID,
		RegionID:     a.RegionID,
		PositionCode: a.PositionCode,
		Title:        a.Title,
		ImageURL:     a.ImageURL,
		LinkURL:      a.LinkURL,
		Sort:         a.Sort,
		StartAt:      a.StartAt,
		EndAt:        a.EndAt,
		Status:       a.Status,
		StatusText:   adStatusText(a.Status),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// Create 创建广告位
func (s *adService) Create(regionID uint, req *dto.CreateAdPositionRequest) (*dto.AdPositionInfo, error) {
	status := req.Status
	if status == 0 {
		status = model.AdStatusEnabled
	}
	a := &model.AdPosition{
		PositionCode: req.PositionCode,
		Title:        req.Title,
		ImageURL:     req.ImageURL,
		LinkURL:      req.LinkURL,
		Sort:         req.Sort,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		Status:       status,
	}
	a.RegionID = regionID
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return toAdPositionInfo(a), nil
}

// Update 更新广告位
func (s *adService) Update(id uint, req *dto.UpdateAdPositionRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAdNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.PositionCode != nil {
		fields["position_code"] = *req.PositionCode
	}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.ImageURL != nil {
		fields["image_url"] = *req.ImageURL
	}
	if req.LinkURL != nil {
		fields["link_url"] = *req.LinkURL
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}
	if req.StartAt != nil {
		fields["start_at"] = req.StartAt
	}
	if req.EndAt != nil {
		fields["end_at"] = req.EndAt
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除广告位
func (s *adService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAdNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取广告位详情
func (s *adService) GetByID(id uint) (*dto.AdPositionInfo, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdNotFound
		}
		return nil, err
	}
	return toAdPositionInfo(a), nil
}

// List 广告位列表
func (s *adService) List(regionID uint, req *dto.AdPositionListRequest) (*utils.Pagination, []dto.AdPositionInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	query := repository.AdPositionListQuery{
		PositionCode: req.PositionCode,
		Status:       req.Status,
		Keyword:      req.Keyword,
	}
	list, total, err := s.repo.List(regionID, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.AdPositionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAdPositionInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByPositionCode 按位置编码获取广告
func (s *adService) ListByPositionCode(regionID uint, positionCode string, page, pageSize int) (*utils.Pagination, []dto.AdPositionInfo, error) {
	pagination := utils.NewPagination(page, pageSize)
	list, total, err := s.repo.FindByPositionCode(regionID, positionCode, pagination)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.AdPositionInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAdPositionInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}
