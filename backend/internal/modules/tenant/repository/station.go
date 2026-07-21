// Package repository 多租户分站数据访问层 - 分站
// 依据架构设计第 4.10 节：多租户分站中台
package repository

import (
	"wuchang-tongcheng/internal/modules/tenant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// StationListOptions 分站列表过滤条件
type StationListOptions struct {
	RegionID uint
	Name     string
	Domain   string
	Status   *int
	Keyword  string
}

// StationRepository 分站仓储接口
type StationRepository interface {
	Create(s *model.Station) error
	FindByID(id uint) (*model.Station, error)
	FindByRegionID(regionID uint) (*model.Station, error)
	FindByDomain(domain string) (*model.Station, error)
	Update(s *model.Station) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(pagination *utils.Pagination, opts StationListOptions) ([]model.Station, int64, error)
}

type stationRepository struct {
	db *gorm.DB
}

// NewStationRepository 创建分站仓储实例
func NewStationRepository(db *gorm.DB) StationRepository {
	return &stationRepository{db: db}
}

func (r *stationRepository) Create(s *model.Station) error {
	return r.db.Create(s).Error
}

func (r *stationRepository) FindByID(id uint) (*model.Station, error) {
	var s model.Station
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *stationRepository) FindByRegionID(regionID uint) (*model.Station, error) {
	var s model.Station
	if err := r.db.Where("region_id = ?", regionID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *stationRepository) FindByDomain(domain string) (*model.Station, error) {
	var s model.Station
	if err := r.db.Where("domain = ?", domain).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *stationRepository) Update(s *model.Station) error {
	return r.db.Save(s).Error
}

func (r *stationRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Station{}).Where("id = ?", id).Updates(fields).Error
}

func (r *stationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Station{}, id).Error
}

func (r *stationRepository) List(pagination *utils.Pagination, opts StationListOptions) ([]model.Station, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Station
	var total int64

	query := r.db.Model(&model.Station{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.Name != "" {
		query = query.Where("name ILIKE ?", "%"+opts.Name+"%")
	}
	if opts.Domain != "" {
		query = query.Where("domain ILIKE ?", "%"+opts.Domain+"%")
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR domain ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
