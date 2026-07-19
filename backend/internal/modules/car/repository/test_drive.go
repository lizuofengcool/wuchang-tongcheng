// Package repository 同城车辆买卖数据访问层 - 试驾预约
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TestDriveRepository 试驾预约仓储接口
type TestDriveRepository interface {
	Create(t *model.CarTestDrive) error
	FindByID(id uint) (*model.CarTestDrive, error)
	FindByDriveNo(driveNo string) (*model.CarTestDrive, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts TestDriveListOptions) ([]model.CarTestDrive, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts TestDriveAdminListOptions) ([]model.CarTestDrive, int64, error)
	// 用户的预约
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error)
	// 销售的预约
	ListBySales(salesID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error)
	// 车商的预约
	ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error)

	// 状态/驾照
	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	UpdateLicenseStatus(id uint, status string, images model.JSONB) error

	// 统计
	CountByCarID(carID uint) (int64, error)
	CountByUser(userID uint, startDate, endDate *time.Time) (int64, error)
	CountByStatus(regionID uint, status int) (int64, error)
}

// TestDriveListOptions C 端试驾列表过滤条件
type TestDriveListOptions struct {
	CarID      uint
	ListingID  uint
	UserID     uint
	DealerID   uint
	SalesID    uint
	DriveType  string
	Status     *int
	StartDate  *time.Time
	EndDate    *time.Time
}

// TestDriveAdminListOptions M 端试驾列表过滤条件
type TestDriveAdminListOptions struct {
	RegionID   uint
	CarID      uint
	UserID     uint
	DealerID   uint
	SalesID    uint
	DriveType  string
	Status     *int
	StartDate  *time.Time
	EndDate    *time.Time
	Keyword    string
}

type testDriveRepository struct {
	db *gorm.DB
}

// NewTestDriveRepository 创建试驾预约仓储实例
func NewTestDriveRepository(db *gorm.DB) TestDriveRepository {
	return &testDriveRepository{db: db}
}

// ===== CRUD =====

func (r *testDriveRepository) Create(t *model.CarTestDrive) error {
	return r.db.Create(t).Error
}

func (r *testDriveRepository) FindByID(id uint) (*model.CarTestDrive, error) {
	var t model.CarTestDrive
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *testDriveRepository) FindByDriveNo(driveNo string) (*model.CarTestDrive, error) {
	var t model.CarTestDrive
	if err := r.db.Where("drive_no = ?", driveNo).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *testDriveRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarTestDrive{}).Where("id = ?", id).Updates(fields).Error
}

func (r *testDriveRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarTestDrive{}, id).Error
}

// ===== 列表查询 =====

func (r *testDriveRepository) List(regionID uint, pagination *utils.Pagination, opts TestDriveListOptions) ([]model.CarTestDrive, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTestDrive
	var total int64

	query := r.db.Model(&model.CarTestDrive{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.DealerID > 0 {
		query = query.Where("dealer_id = ?", opts.DealerID)
	}
	if opts.SalesID > 0 {
		query = query.Where("sales_id = ?", opts.SalesID)
	}
	if opts.DriveType != "" {
		query = query.Where("drive_type = ?", opts.DriveType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.StartDate != nil {
		query = query.Where("appointment_date >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		query = query.Where("appointment_date <= ?", *opts.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("appointment_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *testDriveRepository) AdminList(pagination *utils.Pagination, opts TestDriveAdminListOptions) ([]model.CarTestDrive, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTestDrive
	var total int64

	query := r.db.Model(&model.CarTestDrive{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.DealerID > 0 {
		query = query.Where("dealer_id = ?", opts.DealerID)
	}
	if opts.SalesID > 0 {
		query = query.Where("sales_id = ?", opts.SalesID)
	}
	if opts.DriveType != "" {
		query = query.Where("drive_type = ?", opts.DriveType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.StartDate != nil {
		query = query.Where("appointment_date >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		query = query.Where("appointment_date <= ?", *opts.EndDate)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("drive_no ILIKE ? OR user_name ILIKE ? OR user_phone ILIKE ? OR dealer_name ILIKE ?", like, like, like, like)
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

func (r *testDriveRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTestDrive
	var total int64

	query := r.db.Model(&model.CarTestDrive{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("appointment_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *testDriveRepository) ListBySales(salesID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTestDrive
	var total int64

	query := r.db.Model(&model.CarTestDrive{}).Where("sales_id = ?", salesID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("appointment_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *testDriveRepository) ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarTestDrive, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTestDrive
	var total int64

	query := r.db.Model(&model.CarTestDrive{}).Where("dealer_id = ?", dealerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("appointment_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 状态/驾照 =====

func (r *testDriveRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.CarTestDrive{}).Where("id = ?", id).Updates(fields).Error
}

func (r *testDriveRepository) UpdateLicenseStatus(id uint, status string, images model.JSONB) error {
	fields := map[string]interface{}{
		"license_status": status,
	}
	if images != nil {
		fields["license_images"] = images
	}
	return r.db.Model(&model.CarTestDrive{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 统计 =====

func (r *testDriveRepository) CountByCarID(carID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarTestDrive{}).Where("car_id = ?", carID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *testDriveRepository) CountByUser(userID uint, startDate, endDate *time.Time) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarTestDrive{}).Where("user_id = ?", userID)
	if startDate != nil {
		q = q.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("created_at <= ?", *endDate)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *testDriveRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarTestDrive{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
