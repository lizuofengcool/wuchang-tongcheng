// Package repository 同城拼车出行数据访问层 - 预订
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// BookingListOptions 预订列表过滤条件
type BookingListOptions struct {
	PincheID    uint
	PassengerID uint
	DriverID    uint
	Status      *int
	Keyword     string
}

// BookingRepository 预订仓储接口
type BookingRepository interface {
	Create(b *model.PincheBooking) error
	FindByID(id uint) (*model.PincheBooking, error)
	FindByBookingNo(bookingNo string) (*model.PincheBooking, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts BookingListOptions) ([]model.PincheBooking, int64, error)
	ListByPassenger(passengerID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error)
	ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error)

	// 校验
	HasBooked(passengerID, pincheID uint) (bool, error)
	CountActiveByPinche(pincheID uint) (int64, error)

	UpdateStatus(id uint, status int) error
}

type bookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository 创建预订仓储实例
func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(b *model.PincheBooking) error {
	return r.db.Create(b).Error
}

func (r *bookingRepository) FindByID(id uint) (*model.PincheBooking, error) {
	var b model.PincheBooking
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepository) FindByBookingNo(bookingNo string) (*model.PincheBooking, error) {
	var b model.PincheBooking
	if err := r.db.Where("booking_no = ?", bookingNo).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheBooking{}).Where("id = ?", id).Updates(fields).Error
}

func (r *bookingRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheBooking{}, id).Error
}

func (r *bookingRepository) List(regionID uint, pagination *utils.Pagination, opts BookingListOptions) ([]model.PincheBooking, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheBooking
	var total int64

	query := r.db.Model(&model.PincheBooking{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.PassengerID > 0 {
		query = query.Where("passenger_id = ?", opts.PassengerID)
	}
	if opts.DriverID > 0 {
		query = query.Where("driver_id = ?", opts.DriverID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("booking_no ILIKE ? OR passenger_name ILIKE ? OR driver_name ILIKE ?", like, like, like)
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

func (r *bookingRepository) ListByPassenger(passengerID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheBooking
	var total int64

	query := r.db.Model(&model.PincheBooking{}).Where("passenger_id = ?", passengerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *bookingRepository) ListByDriver(driverID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheBooking
	var total int64

	query := r.db.Model(&model.PincheBooking{}).Where("driver_id = ?", driverID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *bookingRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheBooking, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheBooking
	var total int64

	query := r.db.Model(&model.PincheBooking{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *bookingRepository) HasBooked(passengerID, pincheID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.PincheBooking{}).
		Where("passenger_id = ? AND pinche_id = ? AND status IN ?", passengerID, pincheID, []int{model.BookingStatusPending, model.BookingStatusPaid, model.BookingStatusBoarded, model.BookingStatusCompleted}).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *bookingRepository) CountActiveByPinche(pincheID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PincheBooking{}).
		Where("pinche_id = ? AND status IN ?", pincheID, []int{model.BookingStatusPending, model.BookingStatusPaid, model.BookingStatusBoarded}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *bookingRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheBooking{}).Where("id = ?", id).Update("status", status).Error
}
