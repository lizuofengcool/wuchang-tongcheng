// Package repository 同城房屋租售数据访问层
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 1.5：PostGIS 必选，扩展不可用降级 Haversine
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// HouseRepository 房源主表仓储接口
type HouseRepository interface {
	// 主表 CRUD
	Create(h *model.House) error
	FindByID(id uint) (*model.House, error)
	Update(h *model.House) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离 + 多条件过滤）
	List(regionID uint, req *utils.Pagination, opts HouseListOptions) ([]model.House, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(req *utils.Pagination, opts HouseAdminListOptions) ([]model.House, int64, error)
	// 附近（PostGIS 优先，Haversine 降级）
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts HouseListOptions) ([]model.House, int64, error)
	// 搜索（基于 SQL LIKE，后续可接 Elasticsearch）
	Search(regionID uint, req *utils.Pagination, keyword string) ([]model.House, int64, error)
	// 高级搜索
	AdvancedSearch(regionID uint, req *utils.Pagination, opts HouseAdvancedSearchOptions) ([]model.House, int64, error)
	// 用户自己的发布
	ListByUser(userID uint, req *utils.Pagination) ([]model.House, int64, error)

	// 浏览量
	IncrViewCount(id uint) error
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error
	IncrViewingCount(id uint) error

	// 图片子表
	ListImages(houseID uint) ([]model.HouseImage, error)
	ReplaceImages(houseID uint, imgs []model.HouseImage) error
	DeleteImages(houseID uint) error

	// 收藏
	FavExists(userID, houseID uint) (bool, error)
	CreateFav(fav *model.HouseFavorite) error
	DeleteFav(userID, houseID uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	ListFavs(userID uint, page, pageSize int) ([]model.HouseFavorite, int64, error)
	HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error)

	// 浏览记录
	CreateView(v *model.HouseView) error
	ListViews(userID uint, page, pageSize int) ([]model.HouseView, int64, error)

	// 批量操作
	BatchUpdateStatus(ids []uint, status int) (int64, error)
	BatchAudit(ids []uint, auditStatus int, auditReason string) (int64, error)
	BatchDelete(ids []uint) (int64, error)

	// 聚合统计
	CountByStatus(regionID uint) (map[int]int64, error)
	CountPendingAudit(regionID uint) (int64, error)
	CountTodayNew(regionID uint) (int64, error)
}

// HouseListOptions C 端列表过滤条件
type HouseListOptions struct {
	CategoryID        uint
	CommunityID       uint
	AgentID           uint
	Keyword           string
	ListingType       string
	PropertyType      string
	SourceType        string
	RentType          string
	MinRentPrice      float64
	MaxRentPrice      float64
	MinSalePrice      float64
	MaxSalePrice      float64
	MinBuildingArea   float64
	MaxBuildingArea   float64
	Rooms             int
	FloorType         string
	Orientation       string
	Decoration        string
	HasElevator       *bool
	Featured          *bool
	Verified          *bool
	RealHouseVerified *bool
	Sort              string // latest/price_asc/price_desc/popular/area_asc/area_desc
	Status            int    // 默认 1（已发布），-1 全部
}

// HouseAdminListOptions M 端管理列表过滤条件
// Status/AuditStatus 使用 *int 指针：nil 表示不过滤，非 nil 按值过滤
type HouseAdminListOptions struct {
	RegionID    uint
	UserID      uint
	CategoryID  uint
	CommunityID uint
	ListingType string
	Status      *int
	AuditStatus *int
	Keyword     string
}

// HouseAdvancedSearchOptions 高级搜索条件
type HouseAdvancedSearchOptions struct {
	HouseListOptions
	Latitude  float64
	Longitude float64
	RadiusKm  float64
}

type houseRepository struct {
	db *gorm.DB
}

// NewHouseRepository 创建仓储实例
func NewHouseRepository(db *gorm.DB) HouseRepository {
	return &houseRepository{db: db}
}

// ===== 主表 CRUD =====

func (r *houseRepository) Create(h *model.House) error {
	return r.db.Create(h).Error
}

func (r *houseRepository) FindByID(id uint) (*model.House, error) {
	var h model.House
	if err := r.db.First(&h, id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *houseRepository) Update(h *model.House) error {
	return r.db.Save(h).Error
}

func (r *houseRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).Updates(fields).Error
}

func (r *houseRepository) Delete(id uint) error {
	// 软删除主表 + 删除关联图片（图片无软删除）
	if err := r.db.Delete(&model.House{}, id).Error; err != nil {
		return err
	}
	return r.db.Where("house_id = ?", id).Delete(&model.HouseImage{}).Error
}

// ===== 列表查询（C 端） =====

func (r *houseRepository) List(regionID uint, pagination *utils.Pagination, opts HouseListOptions) ([]model.House, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.House
	var total int64

	query := r.db.Model(&model.House{})

	// 地区隔离
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	// 默认仅返回已发布且审核通过
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.StatusPublished).
			Where("audit_status = ?", model.AuditApproved)
	}
	query = applyHouseFilters(query, opts)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := buildHouseOrderClause(opts.Sort)
	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// applyHouseFilters 应用 C 端过滤条件
func applyHouseFilters(query *gorm.DB, opts HouseListOptions) *gorm.DB {
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.CommunityID > 0 {
		query = query.Where("community_id = ?", opts.CommunityID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.PropertyType != "" {
		query = query.Where("property_type = ?", opts.PropertyType)
	}
	if opts.SourceType != "" {
		query = query.Where("source_type = ?", opts.SourceType)
	}
	if opts.RentType != "" {
		query = query.Where("rent_type = ?", opts.RentType)
	}
	if opts.MinRentPrice > 0 {
		query = query.Where("rent_price >= ?", opts.MinRentPrice)
	}
	if opts.MaxRentPrice > 0 {
		query = query.Where("rent_price <= ?", opts.MaxRentPrice)
	}
	if opts.MinSalePrice > 0 {
		query = query.Where("sale_price >= ?", opts.MinSalePrice)
	}
	if opts.MaxSalePrice > 0 {
		query = query.Where("sale_price <= ?", opts.MaxSalePrice)
	}
	if opts.MinBuildingArea > 0 {
		query = query.Where("building_area >= ?", opts.MinBuildingArea)
	}
	if opts.MaxBuildingArea > 0 {
		query = query.Where("building_area <= ?", opts.MaxBuildingArea)
	}
	if opts.Rooms > 0 {
		query = query.Where("rooms = ?", opts.Rooms)
	}
	if opts.FloorType != "" {
		query = query.Where("floor_type = ?", opts.FloorType)
	}
	if opts.Orientation != "" {
		query = query.Where("orientation = ?", opts.Orientation)
	}
	if opts.Decoration != "" {
		query = query.Where("decoration = ?", opts.Decoration)
	}
	if opts.HasElevator != nil {
		query = query.Where("has_elevator = ?", *opts.HasElevator)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.RealHouseVerified != nil {
		query = query.Where("real_house_verified = ?", *opts.RealHouseVerified)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR address ILIKE ? OR business_district ILIKE ? OR layout ILIKE ?", like, like, like, like)
	}
	return query
}

// buildHouseOrderClause 构造排序子句
func buildHouseOrderClause(sort string) string {
	orderClause := "published_at DESC, id DESC"
	switch sort {
	case "price_asc":
		// 出租按 rent_price，出售按 sale_price；为简化使用 COALESCE
		orderClause = "COALESCE(NULLIF(sale_price, 0), NULLIF(rent_price, 0)) ASC, id DESC"
	case "price_desc":
		orderClause = "COALESCE(NULLIF(sale_price, 0), NULLIF(rent_price, 0)) DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	case "area_asc":
		orderClause = "building_area ASC, id DESC"
	case "area_desc":
		orderClause = "building_area DESC, id DESC"
	}
	// 精选/置顶优先
	orderClause = "featured DESC, picked DESC, promotion_level DESC, " + orderClause
	return orderClause
}

// ===== 管理后台列表（M 端） =====

func (r *houseRepository) AdminList(req *utils.Pagination, opts HouseAdminListOptions) ([]model.House, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.House
	var total int64

	query := r.db.Model(&model.House{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.CommunityID > 0 {
		query = query.Where("community_id = ?", opts.CommunityID)
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
		query = query.Where("title ILIKE ? OR user_name ILIKE ? OR address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 附近查询（PostGIS 优先，Haversine 降级） =====

// houseNearbyResult 用于扫描带 distance 计算列的原始查询结果
type houseNearbyResult struct {
	model.House
	Distance float64 `gorm:"column:distance"`
}

// haversineExpr 纯 SQL Haversine 公式（返回公里）
const haversineExpr = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

func (r *houseRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts HouseListOptions) ([]model.House, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	if radiusKm <= 0 {
		radiusKm = 5
	}
	if radiusKm > 100 {
		radiusKm = 100
	}

	if geo.PostGISAvailable(r.db) {
		return r.listNearbyPostGIS(regionID, pagination, lat, lng, radiusKm, opts)
	}
	return r.listNearbyHaversine(regionID, pagination, lat, lng, radiusKm, opts)
}

func (r *houseRepository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts HouseListOptions) ([]model.House, int64, error) {
	radiusMeters := radiusKm * 1000.0

	where := "deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0 " +
		"AND ST_DWithin(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?)), ?)"
	args := []interface{}{lng, lat, radiusMeters}
	if regionID > 0 {
		where += " AND region_id = ?"
		args = append(args, regionID)
	}
	if opts.ListingType != "" {
		where += " AND listing_type = ?"
		args = append(args, opts.ListingType)
	}
	if opts.PropertyType != "" {
		where += " AND property_type = ?"
		args = append(args, opts.PropertyType)
	}

	countArgs := append([]interface{}{}, args...)
	var total int64
	if err := r.db.Model(&model.House{}).Where(where, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM houses WHERE " + where +
		" ORDER BY featured DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []houseNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenHouseNearby(rows), total, nil
}

func (r *houseRepository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts HouseListOptions) ([]model.House, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.House{}).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.PropertyType != "" {
		query = query.Where("property_type = ?", opts.PropertyType)
	}

	// 用 Haversine 表达式精确过滤
	haversineWhere := haversineExpr + " <= ?"
	query = query.Where(haversineWhere, lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表（带 distance 计算列）
	selectSQL := "*, " + haversineExpr + " AS distance"
	listQuery := r.db.Table("houses").
		Select(selectSQL, lat, lat, lng).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng).
		Where(haversineWhere, lat, lat, lng, radiusKm)

	if regionID > 0 {
		listQuery = listQuery.Where("region_id = ?", regionID)
	}
	if opts.ListingType != "" {
		listQuery = listQuery.Where("listing_type = ?", opts.ListingType)
	}
	if opts.PropertyType != "" {
		listQuery = listQuery.Where("property_type = ?", opts.PropertyType)
	}

	var rows []houseNearbyResult
	if err := listQuery.Order("featured DESC, distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenHouseNearby(rows), total, nil
}

// flattenHouseNearby 把查询结果中的 Distance 回填到 model.House.Distance
func flattenHouseNearby(rows []houseNearbyResult) []model.House {
	list := make([]model.House, 0, len(rows))
	for _, row := range rows {
		h := row.House
		h.Distance = row.Distance
		list = append(list, h)
	}
	return list
}

// ===== 搜索 =====

func (r *houseRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.House, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.House
	var total int64

	query := r.db.Model(&model.House{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR address ILIKE ? OR business_district ILIKE ? OR layout ILIKE ?", like, like, like, like, like)
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

// ===== 高级搜索 =====

func (r *houseRepository) AdvancedSearch(regionID uint, pagination *utils.Pagination, opts HouseAdvancedSearchOptions) ([]model.House, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.House
	var total int64

	query := r.db.Model(&model.House{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	query = applyHouseFilters(query, opts.HouseListOptions)

	if opts.Latitude != 0 && opts.Longitude != 0 && opts.RadiusKm > 0 {
		// 附近 + 高级过滤（仅用 Haversine）
		minLat, maxLat, minLng, maxLng := geo.BoundingBox(opts.Latitude, opts.Longitude, opts.RadiusKm)
		query = query.Where("latitude BETWEEN ? AND ?", minLat, maxLat).
			Where("longitude BETWEEN ? AND ?", minLng, maxLng).
			Where(haversineExpr+" <= ?", opts.Latitude, opts.Latitude, opts.Longitude, opts.RadiusKm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := buildHouseOrderClause(opts.Sort)
	// 注：高级搜索距离排序需走原生 SQL，此处使用普通排序
	// 距离排序请使用 ListNearby 接口
	if err := query.Scopes(utils.Paginate(pagination)).
		Order(orderClause).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 用户自己的发布 =====

func (r *houseRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.House, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.House
	var total int64

	query := r.db.Model(&model.House{}).Where("user_id = ?", userID)
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

// ===== 浏览量 / 联系 / 分享 / 看房数 =====

func (r *houseRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *houseRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *houseRepository) IncrShareCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *houseRepository) IncrViewingCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).
		UpdateColumn("viewing_count", gorm.Expr("viewing_count + 1")).Error
}

// ===== 图片子表 =====

func (r *houseRepository) ListImages(houseID uint) ([]model.HouseImage, error) {
	var imgs []model.HouseImage
	if err := r.db.Where("house_id = ?", houseID).Order("sort ASC, id ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	return imgs, nil
}

// ReplaceImages 全量替换图片（先删除旧的，再插入新的）
func (r *houseRepository) ReplaceImages(houseID uint, imgs []model.HouseImage) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// 删除旧图片
	if err := tx.Where("house_id = ?", houseID).Delete(&model.HouseImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 插入新图片
	for i, img := range imgs {
		if img.URL == "" {
			continue
		}
		img.HouseID = houseID
		if img.Sort == 0 {
			img.Sort = i
		}
		if err := tx.Create(&img).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *houseRepository) DeleteImages(houseID uint) error {
	return r.db.Where("house_id = ?", houseID).Delete(&model.HouseImage{}).Error
}

// ===== 收藏 =====

func (r *houseRepository) FavExists(userID, houseID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.HouseFavorite{}).
		Where("user_id = ? AND house_id = ? AND favorite_type = ?", userID, houseID, model.FavoriteTypeHouse).
		Count(&count).Error
	return count > 0, err
}

func (r *houseRepository) CreateFav(fav *model.HouseFavorite) error {
	return r.db.Create(fav).Error
}

func (r *houseRepository) DeleteFav(userID, houseID uint) error {
	return r.db.Where("user_id = ? AND house_id = ? AND favorite_type = ?", userID, houseID, model.FavoriteTypeHouse).
		Delete(&model.HouseFavorite{}).Error
}

func (r *houseRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *houseRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.House{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *houseRepository) ListFavs(userID uint, page, pageSize int) ([]model.HouseFavorite, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.HouseFavorite
	var total int64

	query := r.db.Model(&model.HouseFavorite{}).
		Where("user_id = ? AND favorite_type = ?", userID, model.FavoriteTypeHouse)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// HasFavedBatch 批量查询当前用户对一组 house 的收藏状态
func (r *houseRepository) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	var favs []model.HouseFavorite
	if err := r.db.Where("user_id = ? AND favorite_type = ? AND house_id IN ?", userID, model.FavoriteTypeHouse, ids).Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		result[f.HouseID] = true
	}
	return result, nil
}

// ===== 浏览记录 =====

func (r *houseRepository) CreateView(v *model.HouseView) error {
	return r.db.Create(v).Error
}

func (r *houseRepository) ListViews(userID uint, page, pageSize int) ([]model.HouseView, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.HouseView
	var total int64

	query := r.db.Model(&model.HouseView{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 批量操作 =====

func (r *houseRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.House{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

func (r *houseRepository) BatchAudit(ids []uint, auditStatus int, auditReason string) (int64, error) {
	fields := map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": auditReason,
	}
	result := r.db.Model(&model.House{}).Where("id IN ?", ids).Updates(fields)
	return result.RowsAffected, result.Error
}

func (r *houseRepository) BatchDelete(ids []uint) (int64, error) {
	result := r.db.Where("id IN ?", ids).Delete(&model.House{})
	return result.RowsAffected, result.Error
}

// ===== 聚合统计 =====

func (r *houseRepository) CountByStatus(regionID uint) (map[int]int64, error) {
	type row struct {
		Status int   `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	var rows []row
	query := r.db.Model(&model.House{}).Select("status, COUNT(*) as count")
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int]int64, len(rows))
	for _, r := range rows {
		result[r.Status] = r.Count
	}
	return result, nil
}

func (r *houseRepository) CountPendingAudit(regionID uint) (int64, error) {
	var count int64
	query := r.db.Model(&model.House{}).Where("audit_status = ?", model.AuditPending)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *houseRepository) CountTodayNew(regionID uint) (int64, error) {
	var count int64
	query := r.db.Model(&model.House{}).Where("created_at >= DATE(NOW())")
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
