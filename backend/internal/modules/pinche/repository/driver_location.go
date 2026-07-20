// Package repository 同城拼车出行数据访问层 - 车主实时位置
// 注意：本表使用 BaseModel 无 region_id，按用户/车主/行程维度查询
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// DriverLocationRepository 车主实时位置仓储接口
type DriverLocationRepository interface {
	Create(l *model.PincheDriverLocation) error
	FindByID(id uint) (*model.PincheDriverLocation, error)
	FindLatestByPinche(pincheID uint) (*model.PincheDriverLocation, error)
	FindLatestByDriver(driverID uint) (*model.PincheDriverLocation, error)
	FindLatestByTrip(tripID uint) (*model.PincheDriverLocation, error)
	Update(id uint, fields map[string]interface{}) error
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheDriverLocation, int64, error)
	ListByDriver(driverID uint, startTime, endTime time.Time, pagination *utils.Pagination) ([]model.PincheDriverLocation, int64, error)
	DeleteByPinche(pincheID uint) error
}

type driverLocationRepository struct {
	db *gorm.DB
}

// NewDriverLocationRepository 创建车主实时位置仓储实例
func NewDriverLocationRepository(db *gorm.DB) DriverLocationRepository {
	return &driverLocationRepository{db: db}
}

func (r *driverLocationRepository) Create(l *model.PincheDriverLocation) error {
	return r.db.Create(l).Error
}

func (r *driverLocationRepository) FindByID(id uint) (*model.PincheDriverLocation, error) {
	var l model.PincheDriverLocation
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *driverLocationRepository) FindLatestByPinche(pincheID uint) (*model.PincheDriverLocation, error) {
	var l model.PincheDriverLocation
	if err := r.db.Where("pinche_id = ?", pincheID).Order("location_time DESC, id DESC").First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *driverLocationRepository) FindLatestByDriver(driverID uint) (*model.PincheDriverLocation, error) {
	var l model.PincheDriverLocation
	if err := r.db.Where("driver_id = ?", driverID).Order("location_time DESC, id DESC").First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *driverLocationRepository) FindLatestByTrip(tripID uint) (*model.PincheDriverLocation, error) {
	var l model.PincheDriverLocation
	if err := r.db.Where("trip_id = ?", tripID).Order("location_time DESC, id DESC").First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *driverLocationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheDriverLocation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *driverLocationRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheDriverLocation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheDriverLocation
	var total int64

	query := r.db.Model(&model.PincheDriverLocation{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("location_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *driverLocationRepository) ListByDriver(driverID uint, startTime, endTime time.Time, pagination *utils.Pagination) ([]model.PincheDriverLocation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheDriverLocation
	var total int64

	query := r.db.Model(&model.PincheDriverLocation{}).Where("driver_id = ?", driverID)
	if !startTime.IsZero() {
		query = query.Where("location_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("location_time <= ?", endTime)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("location_time DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *driverLocationRepository) DeleteByPinche(pincheID uint) error {
	return r.db.Where("pinche_id = ?", pincheID).Delete(&model.PincheDriverLocation{}).Error
}
