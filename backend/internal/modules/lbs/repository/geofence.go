// Package repository LBS地图中台数据访问层 - 地理围栏
package repository

import (
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// GeofenceRepository 围栏仓储接口
type GeofenceRepository interface {
	// CRUD
	Create(g *model.Geofence) error
	FindByID(id uint) (*model.Geofence, error)
	Update(g *model.Geofence) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(pagination *utils.Pagination, opts GeofenceListOptions) ([]model.Geofence, int64, error)
	ListByRegion(regionID uint) ([]model.Geofence, error)
	ListByOwner(ownerID uint, ownerType string) ([]model.Geofence, error)

	// 围栏判断
	ListEnabledByRegion(regionID uint) ([]model.Geofence, error)
}

// GeofenceListOptions 围栏列表过滤条件
type GeofenceListOptions struct {
	RegionID  uint
	OwnerID   uint
	OwnerType string
	Type      string
	Status    *int
	Keyword   string
}

type geofenceRepository struct {
	db *gorm.DB
}

// NewGeofenceRepository 创建围栏仓储实例
func NewGeofenceRepository(db *gorm.DB) GeofenceRepository {
	return &geofenceRepository{db: db}
}

func (r *geofenceRepository) Create(g *model.Geofence) error {
	return r.db.Create(g).Error
}

func (r *geofenceRepository) FindByID(id uint) (*model.Geofence, error) {
	var g model.Geofence
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *geofenceRepository) Update(g *model.Geofence) error {
	return r.db.Save(g).Error
}

func (r *geofenceRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Geofence{}).Where("id = ?", id).Updates(fields).Error
}

func (r *geofenceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Geofence{}, id).Error
}

func (r *geofenceRepository) List(pagination *utils.Pagination, opts GeofenceListOptions) ([]model.Geofence, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Geofence
	var total int64

	query := r.db.Model(&model.Geofence{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.OwnerID > 0 {
		query = query.Where("owner_id = ?", opts.OwnerID)
	}
	if opts.OwnerType != "" {
		query = query.Where("owner_type = ?", opts.OwnerType)
	}
	if opts.Type != "" {
		query = query.Where("type = ?", opts.Type)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *geofenceRepository) ListByRegion(regionID uint) ([]model.Geofence, error) {
	var list []model.Geofence
	if err := r.db.Where("region_id = ?", regionID).
		Order("sort ASC, id DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *geofenceRepository) ListByOwner(ownerID uint, ownerType string) ([]model.Geofence, error) {
	var list []model.Geofence
	q := r.db.Where("owner_id = ?", ownerID)
	if ownerType != "" {
		q = q.Where("owner_type = ?", ownerType)
	}
	if err := q.Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *geofenceRepository) ListEnabledByRegion(regionID uint) ([]model.Geofence, error) {
	var list []model.Geofence
	q := r.db.Where("status = ?", model.LBSGeofenceStatusEnabled)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
