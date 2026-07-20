// Package repository 同城拼车出行数据访问层 - 车辆
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// VehicleListOptions 车辆列表过滤条件
type VehicleListOptions struct {
	DriverID uint
	UserID   uint
	Status   *int
	Keyword  string
}

// VehicleRepository 车辆仓储接口
type VehicleRepository interface {
	Create(v *model.PincheVehicle) error
	FindByID(id uint) (*model.PincheVehicle, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts VehicleListOptions) ([]model.PincheVehicle, int64, error)
	ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheVehicle, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheVehicle, int64, error)

	UpdateStatus(id uint, status int, reason string) error
	SetDefault(driverID uint, vehicleID uint) error
	CountByDriver(driverID uint) (int64, error)
}

type vehicleRepository struct {
	db *gorm.DB
}

// NewVehicleRepository 创建车辆仓储实例
func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) Create(v *model.PincheVehicle) error {
	return r.db.Create(v).Error
}

func (r *vehicleRepository) FindByID(id uint) (*model.PincheVehicle, error) {
	var v model.PincheVehicle
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *vehicleRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheVehicle{}).Where("id = ?", id).Updates(fields).Error
}

func (r *vehicleRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheVehicle{}, id).Error
}

func (r *vehicleRepository) List(regionID uint, pagination *utils.Pagination, opts VehicleListOptions) ([]model.PincheVehicle, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheVehicle
	var total int64

	query := r.db.Model(&model.PincheVehicle{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	query = applyVehicleFilters(query, opts)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_default DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *vehicleRepository) ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheVehicle, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheVehicle
	var total int64

	query := r.db.Model(&model.PincheVehicle{}).Where("driver_id = ?", driverID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_default DESC, created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *vehicleRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheVehicle, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheVehicle
	var total int64

	query := r.db.Model(&model.PincheVehicle{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_default DESC, created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *vehicleRepository) UpdateStatus(id uint, status int, reason string) error {
	return r.db.Model(&model.PincheVehicle{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "audit_reason": reason}).Error
}

func (r *vehicleRepository) SetDefault(driverID uint, vehicleID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先清除该车主所有车辆的默认标记
		if err := tx.Model(&model.PincheVehicle{}).Where("driver_id = ?", driverID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// 再设置指定车辆为默认
		return tx.Model(&model.PincheVehicle{}).Where("id = ?", vehicleID).
			Update("is_default", true).Error
	})
}

func (r *vehicleRepository) CountByDriver(driverID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PincheVehicle{}).Where("driver_id = ?", driverID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func applyVehicleFilters(query *gorm.DB, opts VehicleListOptions) *gorm.DB {
	if opts.DriverID > 0 {
		query = query.Where("driver_id = ?", opts.DriverID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("plate_no ILIKE ? OR brand ILIKE ? OR model ILIKE ?", like, like, like)
	}
	return query
}
