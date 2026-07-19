// Package repository 房源发布数据访问层
// 与 houses 主表 1:1 冗余发布信息
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ListingRepository 发布仓储接口
type ListingRepository interface {
	Create(l *model.HouseListing) error
	FindByID(id uint) (*model.HouseListing, error)
	FindByListingNo(no string) (*model.HouseListing, error)
	FindByHouseID(houseID uint) (*model.HouseListing, error)
	Update(l *model.HouseListing) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, req *utils.Pagination, opts ListingListOptions) ([]model.HouseListing, int64, error)
	AdminList(req *utils.Pagination, opts ListingAdminListOptions) ([]model.HouseListing, int64, error)
	ListByPublisher(publisherID uint, req *utils.Pagination) ([]model.HouseListing, int64, error)

	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	IncrContactCount(id uint) error
	IncrRefreshCount(id uint) error

	BatchUpdateStatus(ids []uint, status int) (int64, error)
	BatchAudit(ids []uint, auditStatus int, auditReason string) (int64, error)
}

// ListingListOptions C 端列表过滤条件
type ListingListOptions struct {
	HouseID       uint
	CommunityID   uint
	AgentID       uint
	PublisherID   uint
	PublisherType string
	ListingType   string
	Keyword       string
	Sort          string // latest/price_asc/price_desc/popular
	Status        *int
}

// ListingAdminListOptions M 端管理列表过滤条件
type ListingAdminListOptions struct {
	RegionID    uint
	HouseID     uint
	PublisherID uint
	ListingType string
	Status      *int
	AuditStatus *int
	Keyword     string
}

type listingRepository struct {
	db *gorm.DB
}

// NewListingRepository 创建仓储实例
func NewListingRepository(db *gorm.DB) ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) Create(l *model.HouseListing) error {
	return r.db.Create(l).Error
}

func (r *listingRepository) FindByID(id uint) (*model.HouseListing, error) {
	var l model.HouseListing
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) FindByListingNo(no string) (*model.HouseListing, error) {
	var l model.HouseListing
	if err := r.db.Where("listing_no = ?", no).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) FindByHouseID(houseID uint) (*model.HouseListing, error) {
	var l model.HouseListing
	if err := r.db.Where("house_id = ?", houseID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *listingRepository) Update(l *model.HouseListing) error {
	return r.db.Save(l).Error
}

func (r *listingRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ?", id).Updates(fields).Error
}

func (r *listingRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseListing{}, id).Error
}

func (r *listingRepository) List(regionID uint, req *utils.Pagination, opts ListingListOptions) ([]model.HouseListing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseListing
	var total int64

	query := r.db.Model(&model.HouseListing{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	} else {
		query = query.Where("status = ?", model.ListingStatusPublished)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.CommunityID > 0 {
		query = query.Where("community_id = ?", opts.CommunityID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.PublisherID > 0 {
		query = query.Where("publisher_id = ?", opts.PublisherID)
	}
	if opts.PublisherType != "" {
		query = query.Where("publisher_type = ?", opts.PublisherType)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "price_asc":
		orderClause = "price ASC, id DESC"
	case "price_desc":
		orderClause = "price DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}
	if err := query.Scopes(utils.Paginate(req)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) AdminList(req *utils.Pagination, opts ListingAdminListOptions) ([]model.HouseListing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseListing
	var total int64

	query := r.db.Model(&model.HouseListing{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.PublisherID > 0 {
		query = query.Where("publisher_id = ?", opts.PublisherID)
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
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR publisher_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) ListByPublisher(publisherID uint, req *utils.Pagination) ([]model.HouseListing, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseListing
	var total int64

	query := r.db.Model(&model.HouseListing{}).Where("publisher_id = ?", publisherID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *listingRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *listingRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *listingRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *listingRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *listingRepository) IncrRefreshCount(id uint) error {
	return r.db.Model(&model.HouseListing{}).Where("id = ?", id).
		UpdateColumn("refresh_count", gorm.Expr("refresh_count + 1")).Error
}

func (r *listingRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseListing{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

func (r *listingRepository) BatchAudit(ids []uint, auditStatus int, auditReason string) (int64, error) {
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	result := r.db.Model(&model.HouseListing{}).Where("id IN ?", ids).Updates(fields)
	return result.RowsAffected, result.Error
}
