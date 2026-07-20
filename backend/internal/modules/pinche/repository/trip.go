// Package repository 同城拼车出行数据访问层 - 完成行程（含行程分享）
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TripListOptions 行程列表过滤条件
type TripListOptions struct {
	UserID      uint
	DriverID    uint
	PassengerID uint
	PincheID    uint
	Status      *int
	TripNo      string
	StartTime   string
	EndTime     string
}

// TripRepository 完成行程仓储接口
type TripRepository interface {
	Create(t *model.PincheTrip) error
	FindByID(id uint) (*model.PincheTrip, error)
	FindByTripNo(tripNo string) (*model.PincheTrip, error)
	FindByShareToken(token string) (*model.PincheTrip, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts TripListOptions) ([]model.PincheTrip, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error)
	ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error)
	ListByPassenger(passengerID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error)

	UpdateStatus(id uint, status int) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type tripRepository struct {
	db *gorm.DB
}

// NewTripRepository 创建完成行程仓储实例
func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) Create(t *model.PincheTrip) error {
	return r.db.Create(t).Error
}

func (r *tripRepository) FindByID(id uint) (*model.PincheTrip, error) {
	var t model.PincheTrip
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tripRepository) FindByTripNo(tripNo string) (*model.PincheTrip, error) {
	var t model.PincheTrip
	if err := r.db.Where("trip_no = ?", tripNo).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tripRepository) FindByShareToken(token string) (*model.PincheTrip, error) {
	var t model.PincheTrip
	if err := r.db.Where("share_token = ?", token).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tripRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheTrip{}).Where("id = ?", id).Updates(fields).Error
}

func (r *tripRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheTrip{}, id).Error
}

func (r *tripRepository) List(regionID uint, pagination *utils.Pagination, opts TripListOptions) ([]model.PincheTrip, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheTrip
	var total int64

	query := r.db.Model(&model.PincheTrip{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("driver_id = ? OR passenger_id = ?", opts.UserID, opts.UserID)
	}
	if opts.DriverID > 0 {
		query = query.Where("driver_id = ?", opts.DriverID)
	}
	if opts.PassengerID > 0 {
		query = query.Where("passenger_id = ?", opts.PassengerID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.TripNo != "" {
		query = query.Where("trip_no ILIKE ?", "%"+opts.TripNo+"%")
	}
	if opts.StartTime != "" {
		query = query.Where("created_at >= ?", opts.StartTime)
	}
	if opts.EndTime != "" {
		query = query.Where("created_at <= ?", opts.EndTime)
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

func (r *tripRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheTrip
	var total int64

	query := r.db.Model(&model.PincheTrip{}).Where("driver_id = ? OR passenger_id = ?", userID, userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *tripRepository) ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheTrip
	var total int64

	query := r.db.Model(&model.PincheTrip{}).Where("driver_id = ?", driverID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *tripRepository) ListByPassenger(passengerID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheTrip
	var total int64

	query := r.db.Model(&model.PincheTrip{}).Where("passenger_id = ?", passengerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *tripRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheTrip, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheTrip
	var total int64

	query := r.db.Model(&model.PincheTrip{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *tripRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheTrip{}).Where("id = ?", id).Update("status", status).Error
}

func (r *tripRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheTrip{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
