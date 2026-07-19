// Package repository 同城车辆买卖数据访问层 - 发布单
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ListingRepository 发布单仓储接口
type ListingRepository interface {
	Create(l *model.CarListing) error
	FindByID(id uint) (*model.CarListing, error)
	FindByListingNo(listingNo string) (*model.CarListing, error)
	FindByCarID(carID uint) (*model.CarListing, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts ListingListOptions) ([]model.CarListing, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts ListingAdminListOptions) ([]model.CarListing, int64, error)
	// 用户自己的发布
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.CarListing, int64, error)
	// 车商的发布
	ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarListing, int64, error)

	// 计数器
	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	IncrContactCount(id uint) error
	IncrTestDriveCount(id uint) error

	// 检测状态
	UpdateInspectionStatus(id uint, status int, inspectionID uint) error
	// 真车认证
	UpdateRealCarVerified(id uint, verified bool) error
}

// ListingListOptions C 端发布单列表过滤条件
type ListingListOptions struct {
	CarID            uint
	PublisherID      uint
	DealerID         uint
	ListingType      string
	PublisherType    string
	Status           *int
	AuditStatus      *int
	InspectionStatus *int
	Featured         *bool
	RealCarVerified  *bool
	Keyword          string
}

// ListingAdminListOptions M 端发布单列表过滤条件
type ListingAdminListOptions struct {
	RegionID         uint
	PublisherID      uint
	DealerID         uint
	ListingType      string
	Status           *int
	AuditStatus      *int
	InspectionStatus *int
	Keyword          string
}

type listingRepository struct {
	db *gorm.DB
}

// NewListingRepository 创建发布单仓储实例
func NewListingRepository(db *gorm.DB) ListingRepository {
	return &listingRepository{db: db}
}

// ===== CRUD =====

func (r *listingRepository) Create(l *model.CarListing) error {
	return r.db.Create(l).Error
}

func (r *listingRepository) FindByID(id uint) (*model.CarListing, error) {
	var l model.CarListing
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) FindByListingNo(listingNo string) (*model.CarListing, error) {
	var l model.CarListing
	if err := r.db.Where("listing_no = ?", listingNo).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) FindByCarID(carID uint) (*model.CarListing, error) {
	var l model.CarListing
	if err := r.db.Where("car_id = ?", carID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).Updates(fields).Error
}

func (r *listingRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarListing{}, id).Error
}

// ===== 列表查询 =====

func (r *listingRepository) List(regionID uint, pagination *utils.Pagination, opts ListingListOptions) ([]model.CarListing, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarListing
	var total int64

	query := r.db.Model(&model.CarListing{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
	}
	if opts.PublisherID > 0 {
		query = query.Where("publisher_id = ?", opts.PublisherID)
	}
	if opts.DealerID > 0 {
		query = query.Where("dealer_id = ?", opts.DealerID)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.PublisherType != "" {
		query = query.Where("publisher_type = ?", opts.PublisherType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.InspectionStatus != nil {
		query = query.Where("inspection_status = ?", *opts.InspectionStatus)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.RealCarVerified != nil {
		query = query.Where("real_car_verified = ?", *opts.RealCarVerified)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR listing_no ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("featured DESC, published_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) AdminList(pagination *utils.Pagination, opts ListingAdminListOptions) ([]model.CarListing, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarListing
	var total int64

	query := r.db.Model(&model.CarListing{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.PublisherID > 0 {
		query = query.Where("publisher_id = ?", opts.PublisherID)
	}
	if opts.DealerID > 0 {
		query = query.Where("dealer_id = ?", opts.DealerID)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.InspectionStatus != nil {
		query = query.Where("inspection_status = ?", *opts.InspectionStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR listing_no ILIKE ? OR publisher_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.CarListing, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarListing
	var total int64

	query := r.db.Model(&model.CarListing{}).Where("publisher_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarListing, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarListing
	var total int64

	query := r.db.Model(&model.CarListing{}).Where("dealer_id = ?", dealerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 计数器 =====

func (r *listingRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *listingRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *listingRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.CarListing{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *listingRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *listingRepository) IncrTestDriveCount(id uint) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).
		UpdateColumn("test_drive_count", gorm.Expr("test_drive_count + 1")).Error
}

// ===== 检测/认证 =====

func (r *listingRepository) UpdateInspectionStatus(id uint, status int, inspectionID uint) error {
	fields := map[string]interface{}{
		"inspection_status": status,
	}
	if inspectionID > 0 {
		fields["inspection_id"] = inspectionID
	}
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).Updates(fields).Error
}

func (r *listingRepository) UpdateRealCarVerified(id uint, verified bool) error {
	return r.db.Model(&model.CarListing{}).Where("id = ?", id).
		UpdateColumn("real_car_verified", verified).Error
}
