// Package service 多租户分站业务逻辑层 - 域名
// 职责：域名 CRUD / 主域名切换 / SSL 状态管理
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
	ErrDomainNotFound       = errors.New("域名不存在")
	ErrDomainExists         = errors.New("域名已被绑定")
	ErrDomainStationNotFnd  = errors.New("所属分站不存在")
	ErrDomainSSLInvalid     = errors.New("SSL 状态无效")
)

// DomainService 域名业务接口
type DomainService interface {
	Create(req *dto.CreateDomainRequest) (*dto.DomainInfo, error)
	Update(id uint, req *dto.UpdateDomainRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.DomainInfo, error)
	List(req *dto.DomainListRequest) (*utils.Pagination, []dto.DomainInfo, error)
	ListByStation(stationID uint) ([]dto.DomainInfo, error)
	SetPrimary(id uint) error
	UpdateSSLStatus(id uint, status string) error
}

type domainService struct {
	repo        repository.DomainRepository
	stationRepo repository.StationRepository
}

// NewDomainService 创建域名 service 实例
func NewDomainService(repo repository.DomainRepository, stationRepo repository.StationRepository) DomainService {
	return &domainService{repo: repo, stationRepo: stationRepo}
}

// domainSSLText SSL 状态文本
func domainSSLText(status string) string {
	switch status {
	case model.DomainSSLNone:
		return "未配置"
	case model.DomainSSLPending:
		return "申请中"
	case model.DomainSSLActive:
		return "已生效"
	case model.DomainSSLFailed:
		return "失败"
	}
	return ""
}

// toDomainInfo model -> dto
func toDomainInfo(d *model.Domain) *dto.DomainInfo {
	return &dto.DomainInfo{
		ID:        d.ID,
		StationID: d.StationID,
		Domain:    d.Domain,
		IsPrimary: d.IsPrimary,
		SSLStatus: d.SSLStatus,
		SSLText:   domainSSLText(d.SSLStatus),
		CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: d.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Create 绑定域名
func (s *domainService) Create(req *dto.CreateDomainRequest) (*dto.DomainInfo, error) {
	// 校验分站存在
	station, err := s.stationRepo.FindByID(req.StationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDomainStationNotFnd
		}
		return nil, err
	}
	// 校验域名唯一
	if existing, err := s.repo.FindByDomain(req.Domain); err == nil && existing != nil {
		return nil, ErrDomainExists
	}

	sslStatus := req.SSLStatus
	if sslStatus == "" {
		sslStatus = model.DomainSSLNone
	}

	// 若设为主域名，先清除同分站其他主域名标记
	if req.IsPrimary {
		if e := s.repo.ClearPrimary(req.StationID); e != nil {
			return nil, e
		}
	}

	d := &model.Domain{
		StationID: req.StationID,
		Domain:    req.Domain,
		IsPrimary: req.IsPrimary,
		SSLStatus: sslStatus,
	}
	if err := s.repo.Create(d); err != nil {
		return nil, err
	}

	// 若为主域名，同步更新分站主域名冗余字段
	if req.IsPrimary {
		_ = s.stationRepo.UpdateFields(station.ID, map[string]interface{}{"domain": req.Domain})
	}

	return toDomainInfo(d), nil
}

// Update 更新域名（SSL 状态）
func (s *domainService) Update(id uint, req *dto.UpdateDomainRequest) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotFound
		}
		return err
	}
	fields := make(map[string]interface{})
	if req.SSLStatus != nil {
		fields["ssl_status"] = *req.SSLStatus
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除域名绑定
func (s *domainService) Delete(id uint) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotFound
		}
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	// 若删除的是主域名，尝试将同分站最早的域名提升为主域名
	if d.IsPrimary {
		list, e := s.repo.ListByStation(d.StationID)
		if e == nil && len(list) > 0 {
			_ = s.repo.UpdateFields(list[0].ID, map[string]interface{}{"is_primary": true})
			_ = s.stationRepo.UpdateFields(d.StationID, map[string]interface{}{"domain": list[0].Domain})
		} else {
			_ = s.stationRepo.UpdateFields(d.StationID, map[string]interface{}{"domain": ""})
		}
	}
	return nil
}

// GetByID 获取域名详情
func (s *domainService) GetByID(id uint) (*dto.DomainInfo, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	return toDomainInfo(d), nil
}

// List 域名列表
func (s *domainService) List(req *dto.DomainListRequest) (*utils.Pagination, []dto.DomainInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.DomainListOptions{
		StationID: req.StationID,
		Domain:    req.Domain,
		SSLStatus: req.SSLStatus,
	}
	list, total, err := s.repo.List(pagination, opts)
	if err != nil {
		return nil, nil, err
	}
	infos := make([]dto.DomainInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toDomainInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// ListByStation 按分站查询域名
func (s *domainService) ListByStation(stationID uint) ([]dto.DomainInfo, error) {
	list, err := s.repo.ListByStation(stationID)
	if err != nil {
		return nil, err
	}
	infos := make([]dto.DomainInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toDomainInfo(&list[i]))
	}
	return infos, nil
}

// SetPrimary 设置主域名（清除同分站其他主域名标记，并同步分站冗余字段）
func (s *domainService) SetPrimary(id uint) error {
	d, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotFound
		}
		return err
	}
	if err := s.repo.ClearPrimary(d.StationID); err != nil {
		return err
	}
	if err := s.repo.UpdateFields(id, map[string]interface{}{"is_primary": true}); err != nil {
		return err
	}
	// 同步分站主域名冗余字段
	return s.stationRepo.UpdateFields(d.StationID, map[string]interface{}{"domain": d.Domain})
}

// UpdateSSLStatus 更新 SSL 状态
func (s *domainService) UpdateSSLStatus(id uint, status string) error {
	if status != model.DomainSSLNone && status != model.DomainSSLPending &&
		status != model.DomainSSLActive && status != model.DomainSSLFailed {
		return ErrDomainSSLInvalid
	}
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDomainNotFound
		}
		return err
	}
	return s.repo.UpdateFields(id, map[string]interface{}{"ssl_status": status})
}
