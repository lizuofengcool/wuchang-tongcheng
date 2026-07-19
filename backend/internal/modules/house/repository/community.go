// Package repository 小区信息数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CommunityRepository 小区仓储接口
type CommunityRepository interface {
	Create(c *model.HouseCommunity) error
	FindByID(id uint) (*model.HouseCommunity, error)
	FindByName(name string, city string) (*model.HouseCommunity, error)
	Update(c *model.HouseCommunity) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, req *utils.Pagination, opts CommunityListOptions) ([]model.HouseCommunity, int64, error)
	AdminList(req *utils.Pagination, opts CommunityAdminListOptions) ([]model.HouseCommunity, int64, error)
	ListNearby(regionID uint, req *utils.Pagination, lat, lng, radiusKm float64) ([]model.HouseCommunity, int64, error)

	// 关注
	FollowExists(userID, communityID uint) (bool, error)
	CreateFollow(fav *model.HouseFavorite) error
	DeleteFollow(userID, communityID uint) error
	IncrFollowerCount(id uint) error
	DecrFollowerCount(id uint) error

	// 统计
	IncrOnSaleCount(id uint) error
	DecrOnSaleCount(id uint) error
	IncrOnRentCount(id uint) error
	DecrOnRentCount(id uint) error
}

// CommunityListOptions C 端列表过滤条件
type CommunityListOptions struct {
	City             string
	District         string
	BusinessDistrict string
	BuildingType     string
	Keyword          string
	Sort             string // latest/price_asc/price_desc/follower/house_count
}

// CommunityAdminListOptions M 端管理列表过滤条件
type CommunityAdminListOptions struct {
	RegionID uint
	City     string
	District string
	Status   *int
	Keyword  string
}

type communityRepository struct {
	db *gorm.DB
}

// NewCommunityRepository 创建仓储实例
func NewCommunityRepository(db *gorm.DB) CommunityRepository {
	return &communityRepository{db: db}
}

func (r *communityRepository) Create(c *model.HouseCommunity) error {
	return r.db.Create(c).Error
}

func (r *communityRepository) FindByID(id uint) (*model.HouseCommunity, error) {
	var c model.HouseCommunity
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *communityRepository) FindByName(name string, city string) (*model.HouseCommunity, error) {
	var c model.HouseCommunity
	query := r.db.Where("name = ?", name)
	if city != "" {
		query = query.Where("city = ?", city)
	}
	if err := query.First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *communityRepository) Update(c *model.HouseCommunity) error {
	return r.db.Save(c).Error
}

func (r *communityRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ?", id).Updates(fields).Error
}

func (r *communityRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseCommunity{}, id).Error
}

func (r *communityRepository) List(regionID uint, req *utils.Pagination, opts CommunityListOptions) ([]model.HouseCommunity, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseCommunity
	var total int64

	query := r.db.Model(&model.HouseCommunity{}).Where("status = ?", model.CommunityStatusPublished)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.District != "" {
		query = query.Where("district = ?", opts.District)
	}
	if opts.BusinessDistrict != "" {
		query = query.Where("business_district = ?", opts.BusinessDistrict)
	}
	if opts.BuildingType != "" {
		query = query.Where("building_type = ?", opts.BuildingType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR alias ILIKE ? OR address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "created_at DESC, id DESC"
	switch opts.Sort {
	case "price_asc":
		orderClause = "avg_sale_price ASC, id DESC"
	case "price_desc":
		orderClause = "avg_sale_price DESC, id DESC"
	case "follower":
		orderClause = "follower_count DESC, id DESC"
	case "house_count":
		orderClause = "house_count DESC, id DESC"
	}
	if err := query.Scopes(utils.Paginate(req)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *communityRepository) AdminList(req *utils.Pagination, opts CommunityAdminListOptions) ([]model.HouseCommunity, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseCommunity
	var total int64

	query := r.db.Model(&model.HouseCommunity{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.District != "" {
		query = query.Where("district = ?", opts.District)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR alias ILIKE ? OR address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *communityRepository) ListNearby(regionID uint, req *utils.Pagination, lat, lng, radiusKm float64) ([]model.HouseCommunity, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	if radiusKm <= 0 {
		radiusKm = 5
	}
	if radiusKm > 100 {
		radiusKm = 100
	}

	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)
	query := r.db.Model(&model.HouseCommunity{}).
		Where("status = ?", model.CommunityStatusPublished).
		Where("latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	query = query.Where(haversineExpr+" <= ?", lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.HouseCommunity
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 关注 =====

func (r *communityRepository) FollowExists(userID, communityID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.HouseFavorite{}).
		Where("user_id = ? AND community_id = ? AND favorite_type = ?", userID, communityID, model.FavoriteTypeCommunity).
		Count(&count).Error
	return count > 0, err
}

func (r *communityRepository) CreateFollow(fav *model.HouseFavorite) error {
	return r.db.Create(fav).Error
}

func (r *communityRepository) DeleteFollow(userID, communityID uint) error {
	return r.db.Where("user_id = ? AND community_id = ? AND favorite_type = ?", userID, communityID, model.FavoriteTypeCommunity).
		Delete(&model.HouseFavorite{}).Error
}

func (r *communityRepository) IncrFollowerCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ?", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
}

func (r *communityRepository) DecrFollowerCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ? AND follower_count > 0", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
}

// ===== 统计 =====

func (r *communityRepository) IncrOnSaleCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ?", id).
		UpdateColumn("on_sale_count", gorm.Expr("on_sale_count + 1")).Error
}

func (r *communityRepository) DecrOnSaleCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ? AND on_sale_count > 0", id).
		UpdateColumn("on_sale_count", gorm.Expr("on_sale_count - 1")).Error
}

func (r *communityRepository) IncrOnRentCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ?", id).
		UpdateColumn("on_rent_count", gorm.Expr("on_rent_count + 1")).Error
}

func (r *communityRepository) DecrOnRentCount(id uint) error {
	return r.db.Model(&model.HouseCommunity{}).Where("id = ? AND on_rent_count > 0", id).
		UpdateColumn("on_rent_count", gorm.Expr("on_rent_count - 1")).Error
}
