// Package service 多租户分站业务逻辑层 - 分站
// 依据架构设计第 4.10 节：多租户分站中台
// 职责：分站 CRUD / 启停 / 配置复制 / 域名绑定联动
package service

import (
	"errors"

	"wuchang-tongcheng/internal/modules/tenant/dto"
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/modules/tenant/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrStationNotFound      = errors.New("分站不存在")
	ErrStationRegionExists  = errors.New("该地区已存在分站")
	ErrStationDomainExists  = errors.New("域名已被其他分站占用")
	ErrStationStatusInvalid = errors.New("分站状态不允许此操作")
	ErrStationNotFoundByDom = errors.New("当前域名未匹配到分站")
	ErrCopySameStation      = errors.New("源分站与目标分站不能相同")
	ErrCopySourceNotFound   = errors.New("源分站不存在")
	ErrCopyTargetNotFound   = errors.New("目标分站不存在")
)

// StationService 分站业务接口
type StationService interface {
	Create(req *dto.CreateStationRequest) (*dto.StationInfo, error)
	Update(id uint, req *dto.UpdateStationRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.StationInfo, error)
	GetByRegionID(regionID uint) (*dto.StationInfo, error)
	GetByDomain(domain string) (*dto.StationInfo, error)
	List(req *dto.StationListRequest) (*utils.Pagination, []dto.StationInfo, error)
	UpdateStatus(id uint, status int) error
	CopyConfig(req *dto.CopyConfigRequest) (*dto.CopyConfigResult, error)
}

type stationService struct {
	repo   repository.StationRepository
	cfgRepo repository.ConfigRepository
}

// NewStationService 创建分站 service 实例
func NewStationService(repo repository.StationRepository, cfgRepo repository.ConfigRepository) StationService {
	return &stationService{repo: repo, cfgRepo: cfgRepo}
}

// stationStatusText 状态文本
func stationStatusText(status int) string {
	switch status {
	case model.StationStatusDisabled:
		return "已停用"
	case model.StationStatusEnabled:
		return "已启用"
	}
	return ""
}

// parseConfig 将任意类型解析为 JSONB（用于持久化）
func parseConfig(v any) model.JSONB {
	if v == nil {
		return nil
	}
	b, err := model.FromJSON(v)
	if err != nil {
		return nil
	}
	return b
}

// configToAny 将 JSONB 转换为前端可用的 any（解析为通用结构）
func configToAny(c model.JSONB) any {
	if c == nil || len(c) == 0 {
		return nil
	}
	var out any
	_ = c.Parse(&out)
	return out
}

// toStationInfo model -> dto
func toStationInfo(s *model.Station) *dto.StationInfo {
	return &dto.StationInfo{
		ID:          s.ID,
		RegionID:    s.RegionID,
		Name:        s.Name,
		Domain:      s.Domain,
		Logo:        s.Logo,
		Description: s.Description,
		Status:      s.Status,
		StatusText:  stationStatusText(s.Status),
		Config:      configToAny(s.Config),
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 创建分站
func (s *stationService) Create(req *dto.CreateStationRequest) (*dto.StationInfo, error) {
	// 检查 region_id 唯一
	if existing, err := s.repo.FindByRegionID(req.RegionID); err == nil && existing != nil {
		return nil, ErrStationRegionExists
	}
	// 检查 domain 唯一（若提供）
	if req.Domain != "" {
		if existing, err := s.repo.FindByDomain(req.Domain); err == nil && existing != nil {
			return nil, ErrStationDomainExists
		}
	}

	status := req.Status
	if status == 0 {
		status = model.StationStatusEnabled
	}

	station := &model.Station{
		RegionID:    req.RegionID,
		Name:        req.Name,
		Domain:      req.Domain,
		Logo:        req.Logo,
		Description: req.Description,
		Status:      status,
		Config:      parseConfig(req.Config),
	}
	if err := s.repo.Create(station); err != nil {
		return nil, err
	}
	return toStationInfo(station), nil
}

// Update 更新分站
func (s *stationService) Update(id uint, req *dto.UpdateStationRequest) error {
	station, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStationNotFound
		}
		return err
	}

	fields := make(map[string]interface{})
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Domain != nil {
		// 检查 domain 唯一
		if *req.Domain != "" && *req.Domain != station.Domain {
			if existing, err := s.repo.FindByDomain(*req.Domain); err == nil && existing != nil && existing.ID != id {
				return ErrStationDomainExists
			}
		}
		fields["domain"] = *req.Domain
	}
	if req.Logo != nil {
		fields["logo"] = *req.Logo
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Config != nil {
		fields["config"] = parseConfig(req.Config)
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除分站
func (s *stationService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStationNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取分站详情
func (s *stationService) GetByID(id uint) (*dto.StationInfo, error) {
	station, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStationNotFound
		}
		return nil, err
	}
	return toStationInfo(station), nil
}

// GetByRegionID 按地区 ID 获取分站
func (s *stationService) GetByRegionID(regionID uint) (*dto.StationInfo, error) {
	station, err := s.repo.FindByRegionID(regionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStationNotFound
		}
		return nil, err
	}
	return toStationInfo(station), nil
}

// GetByDomain 按域名获取分站（C 端根据当前域名识别分站）
func (s *stationService) GetByDomain(domain string) (*dto.StationInfo, error) {
	if domain == "" {
		return nil, ErrStationNotFoundByDom
	}
	station, err := s.repo.FindByDomain(domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStationNotFoundByDom
		}
		return nil, err
	}
	return toStationInfo(station), nil
}

// List 分站列表
func (s *stationService) List(req *dto.StationListRequest) (*utils.Pagination, []dto.StationInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.StationListOptions{
		RegionID: req.RegionID,
		Name:     req.Name,
		Domain:   req.Domain,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.StationInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toStationInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStatus 启停分站
func (s *stationService) UpdateStatus(id uint, status int) error {
	if status != model.StationStatusDisabled && status != model.StationStatusEnabled {
		return ErrStationStatusInvalid
	}
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStationNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"status": status})
}

// CopyConfig 配置复制：从源分站复制配置到目标分站
func (s *stationService) CopyConfig(req *dto.CopyConfigRequest) (*dto.CopyConfigResult, error) {
	if req.SourceStationID == req.TargetStationID {
		return nil, ErrCopySameStation
	}
	if _, err := s.repo.FindByID(req.SourceStationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCopySourceNotFound
		}
		return nil, err
	}
	if _, err := s.repo.FindByID(req.TargetStationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCopyTargetNotFound
		}
		return nil, err
	}

	// 取源分站配置（按模块过滤）
	var sourceConfigs []model.Config
	var err error
	if req.BizModule != "" {
		sourceConfigs, err = s.cfgRepo.ListByStationAndModule(req.SourceStationID, req.BizModule)
	} else {
		// 全部模块：用大分页拉取
		pagination := utils.NewPagination(1, 1000)
		opts := repository.ConfigListOptions{StationID: req.SourceStationID}
		sourceConfigs, _, err = s.cfgRepo.List(pagination, opts)
	}
	if err != nil {
		return nil, err
	}

	// 先清除目标分站同模块的旧配置
	if e := s.cfgRepo.DeleteByStation(req.TargetStationID, req.BizModule); e != nil {
		return nil, e
	}

	// 复制
	copied := 0
	for i := range sourceConfigs {
		c := model.Config{
			StationID:   req.TargetStationID,
			BizModule:   sourceConfigs[i].BizModule,
			ConfigKey:   sourceConfigs[i].ConfigKey,
			ConfigValue: sourceConfigs[i].ConfigValue,
		}
		if e := s.cfgRepo.Create(&c); e != nil {
			return nil, e
		}
		copied++
	}

	return &dto.CopyConfigResult{
		SourceStationID: req.SourceStationID,
		TargetStationID: req.TargetStationID,
		BizModule:       req.BizModule,
		CopiedCount:     copied,
	}, nil
}
