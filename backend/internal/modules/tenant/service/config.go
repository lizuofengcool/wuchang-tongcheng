// Package service 多租户分站业务逻辑层 - 配置
// 职责：配置 CRUD / 批量获取 / 按模块获取 / 配置继承（缺失回退默认）
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
	ErrConfigNotFound      = errors.New("配置不存在")
	ErrConfigStationNotFnd = errors.New("所属分站不存在")
)

// ConfigService 配置业务接口
type ConfigService interface {
	Upsert(req *dto.UpsertConfigRequest) (*dto.ConfigInfo, error)
	Update(id uint, req *dto.UpdateConfigRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.ConfigInfo, error)
	List(req *dto.ConfigListRequest) (*utils.Pagination, []dto.ConfigInfo, error)
	ListByStationAndModule(stationID uint, bizModule string) ([]dto.ConfigInfo, error)
	BatchGet(req *dto.BatchGetConfigRequest) ([]dto.ConfigKeyValue, error)
}

type configService struct {
	repo        repository.ConfigRepository
	stationRepo repository.StationRepository
}

// NewConfigService 创建配置 service 实例
func NewConfigService(repo repository.ConfigRepository, stationRepo repository.StationRepository) ConfigService {
	return &configService{repo: repo, stationRepo: stationRepo}
}

// toConfigInfo model -> dto
func toConfigInfo(c *model.Config) *dto.ConfigInfo {
	return &dto.ConfigInfo{
		ID:          c.ID,
		StationID:   c.StationID,
		BizModule:   c.BizModule,
		ConfigKey:   c.ConfigKey,
		ConfigValue: c.ConfigValue,
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Upsert 新增/更新配置（按 station_id + biz_module + config_key 唯一）
func (s *configService) Upsert(req *dto.UpsertConfigRequest) (*dto.ConfigInfo, error) {
	// 校验分站存在
	if _, err := s.stationRepo.FindByID(req.StationID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConfigStationNotFnd
		}
		return nil, err
	}

	c := &model.Config{
		StationID:   req.StationID,
		BizModule:   req.BizModule,
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
	}
	if err := s.repo.Upsert(c); err != nil {
		return nil, err
	}
	// 重新查询获取完整数据
	latest, err := s.repo.FindByStationAndKey(req.StationID, req.BizModule, req.ConfigKey)
	if err != nil {
		return nil, err
	}
	return toConfigInfo(latest), nil
}

// Update 更新配置值
func (s *configService) Update(id uint, req *dto.UpdateConfigRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrConfigNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.ConfigValue != nil {
		fields["config_value"] = *req.ConfigValue
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除配置
func (s *configService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrConfigNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取配置详情
func (s *configService) GetByID(id uint) (*dto.ConfigInfo, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}
	return toConfigInfo(c), nil
}

// List 配置列表
func (s *configService) List(req *dto.ConfigListRequest) (*utils.Pagination, []dto.ConfigInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.ConfigListOptions{
		StationID: req.StationID,
		BizModule: req.BizModule,
		ConfigKey: req.ConfigKey,
		Keyword:   req.Keyword,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.ConfigInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toConfigInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByStationAndModule 按分站+模块查询配置
func (s *configService) ListByStationAndModule(stationID uint, bizModule string) ([]dto.ConfigInfo, error) {
	list, err := s.repo.ListByStationAndModule(stationID, bizModule)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.ConfigInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toConfigInfo(&list[i]))
	}
	return infos, nil
}

// BatchGet 批量获取配置（缺失键返回空值，实现配置继承的查询基础）
func (s *configService) BatchGet(req *dto.BatchGetConfigRequest) ([]dto.ConfigKeyValue, error) {
	list, err := s.repo.BatchGet(req.StationID, req.BizModule, req.ConfigKeys)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ConfigKeyValue, 0, len(list))
	for i := range list {
		result = append(result, dto.ConfigKeyValue{
			ConfigKey:   list[i].ConfigKey,
			ConfigValue: list[i].ConfigValue,
		})
	}
	return result, nil
}
