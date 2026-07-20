// Package repository 同城拼车出行数据访问层 - 车主认证
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// DriverListOptions 车主列表过滤条件
type DriverListOptions struct {
	UserID   uint
	Status   *int
	Verified *bool
	Keyword  string
}

// DriverRepository 车主认证仓储接口
type DriverRepository interface {
	Create(d *model.PincheDriver) error
	FindByID(id uint) (*model.PincheDriver, error)
	FindByUserID(userID uint) (*model.PincheDriver, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts DriverListOptions) ([]model.PincheDriver, int64, error)
	AdminList(pagination *utils.Pagination, opts DriverListOptions) ([]model.PincheDriver, int64, error)

	UpdateStatus(id uint, status int) error
	UpdateVerified(id uint, verified bool) error
	UpdateStats(id uint, tripCount int, totalIncome float64, ratingAvg float64) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type driverRepository struct {
	db *gorm.DB
}

// NewDriverRepository 创建车主认证仓储实例
func NewDriverRepository(db *gorm.DB) DriverRepository {
	return &driverRepository{db: db}
}

func (r *driverRepository) Create(d *model.PincheDriver) error {
	return r.db.Create(d).Error
}

func (r *driverRepository) FindByID(id uint) (*model.PincheDriver, error) {
	var d model.PincheDriver
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *driverRepository) FindByUserID(userID uint) (*model.PincheDriver, error) {
	var d model.PincheDriver
	if err := r.db.Where("user_id = ?", userID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *driverRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheDriver{}).Where("id = ?", id).Updates(fields).Error
}

func (r *driverRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheDriver{}, id).Error
}

func (r *driverRepository) List(regionID uint, pagination *utils.Pagination, opts DriverListOptions) ([]model.PincheDriver, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheDriver
	var total int64

	query := r.db.Model(&model.PincheDriver{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.Keyword != "" {
		query = query.Where("real_name ILIKE ? OR user_name ILIKE ? OR user_phone ILIKE ?",
			"%"+opts.Keyword+"%", "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
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

func (r *driverRepository) AdminList(pagination *utils.Pagination, opts DriverListOptions) ([]model.PincheDriver, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheDriver
	var total int64

	query := r.db.Model(&model.PincheDriver{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.Keyword != "" {
		query = query.Where("real_name ILIKE ? OR user_name ILIKE ? OR user_phone ILIKE ?",
			"%"+opts.Keyword+"%", "%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
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

func (r *driverRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheDriver{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *driverRepository) UpdateVerified(id uint, verified bool) error {
	return r.db.Model(&model.PincheDriver{}).Where("id = ?", id).
		Update("verified", verified).Error
}

func (r *driverRepository) UpdateStats(id uint, tripCount int, totalIncome float64, ratingAvg float64) error {
	return r.db.Model(&model.PincheDriver{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"trip_count":   tripCount,
			"total_income": totalIncome,
			"rating_avg":   ratingAvg,
		}).Error
}

func (r *driverRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheDriver{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
