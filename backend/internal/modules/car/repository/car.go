// Package repository 同城车辆买卖数据访问层 - 车源主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：PostGIS 必选，扩展不可用降级 Haversine
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CarRepository 车源主表仓储接口
type CarRepository interface {
	// 主表 CRUD
	Create(c *model.Car) error
	FindByID(id uint) (*model.Car, error)
	Update(c *model.Car) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts CarListOptions) ([]model.Car, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts CarAdminListOptions) ([]model.Car, int64, error)
	// 附近（PostGIS 优先，Haversine 降级）
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts CarListOptions) ([]model.Car, int64, error)
	// 搜索
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Car, int64, error)
	// 用户自己的发布
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Car, int64, error)

	// 浏览量
	IncrViewCount(id uint) error
	// 收藏数
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	// 联系数
	IncrContactCount(id uint) error
	// 分享数
	IncrShareCount(id uint) error
	// 试驾数
	IncrTestDriveCount(id uint) error

	// 图片子表
	ListImages(carID uint) ([]model.CarImage, error)
	ReplaceImages(carID uint, regionID uint, images []model.CarImage) error
	DeleteImages(carID uint) error

	// 收藏
	FavExists(userID, carID uint) (bool, error)
	CreateFav(fav *model.CarFavorite) error
	DeleteFav(userID, carID uint) error
	ListFavs(userID uint, page, pageSize int) ([]model.CarFavorite, int64, error)
	HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error)

	// 浏览记录
	CreateView(v *model.CarView) error
}

// CarListOptions C 端车源列表过滤条件
type CarListOptions struct {
	CategoryID      uint
	BrandID         uint
	ModelID         uint
	Keyword         string
	CarType         string
	ListingType     string
	SourceType      string
	FuelType        string
	Transmission    string
	MinPrice        float64
	MaxPrice        float64
	MinMileage      float64
	MaxMileage      float64
	MinYear         int
	MaxYear         int
	ConditionLevel  string
	City            string
	Featured        *bool
	Picked          *bool
	Verified        *bool
	RealCarVerified *bool
	PriceNegotiable *bool
	Sort            string // latest/price_asc/price_desc/mileage_asc/year_desc/popular/distance
	Status          int    // 默认 1（已发布），-1 全部
}

// CarAdminListOptions M 端管理列表过滤条件
type CarAdminListOptions struct {
	RegionID    uint
	UserID      uint
	CategoryID  uint
	BrandID     uint
	Status      *int
	AuditStatus *int
	ListingType string
	SourceType  string
	CarType     string
	Keyword     string
}

type carRepository struct {
	db *gorm.DB
}

// NewCarRepository 创建车源仓储实例
func NewCarRepository(db *gorm.DB) CarRepository {
	return &carRepository{db: db}
}

// ===== 主表 CRUD =====

func (r *carRepository) Create(c *model.Car) error {
	return r.db.Create(c).Error
}

func (r *carRepository) FindByID(id uint) (*model.Car, error) {
	var c model.Car
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *carRepository) Update(c *model.Car) error {
	return r.db.Save(c).Error
}

func (r *carRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).Updates(fields).Error
}

func (r *carRepository) Delete(id uint) error {
	// 软删除主表 + 删除关联图片
	if err := r.db.Delete(&model.Car{}, id).Error; err != nil {
		return err
	}
	return r.db.Where("car_id = ?", id).Delete(&model.CarImage{}).Error
}

// ===== 列表查询（C 端） =====

func (r *carRepository) List(regionID uint, pagination *utils.Pagination, opts CarListOptions) ([]model.Car, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Car
	var total int64

	query := r.db.Model(&model.Car{})

	// 地区隔离
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	// 默认仅返回已发布
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.StatusPublished)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		query = query.Where("brand_id = ?", opts.BrandID)
	}
	if opts.ModelID > 0 {
		query = query.Where("model_id = ?", opts.ModelID)
	}
	if opts.CarType != "" {
		query = query.Where("car_type = ?", opts.CarType)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.SourceType != "" {
		query = query.Where("source_type = ?", opts.SourceType)
	}
	if opts.FuelType != "" {
		query = query.Where("fuel_type = ?", opts.FuelType)
	}
	if opts.Transmission != "" {
		query = query.Where("transmission = ?", opts.Transmission)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price <= ?", opts.MaxPrice)
	}
	if opts.MinMileage > 0 {
		query = query.Where("mileage >= ?", opts.MinMileage)
	}
	if opts.MaxMileage > 0 {
		query = query.Where("mileage <= ?", opts.MaxMileage)
	}
	if opts.MinYear > 0 {
		query = query.Where("registration_year >= ?", opts.MinYear)
	}
	if opts.MaxYear > 0 {
		query = query.Where("registration_year <= ?", opts.MaxYear)
	}
	if opts.ConditionLevel != "" {
		query = query.Where("condition_level = ?", opts.ConditionLevel)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Picked != nil {
		query = query.Where("picked = ?", *opts.Picked)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.RealCarVerified != nil {
		query = query.Where("real_car_verified = ?", *opts.RealCarVerified)
	}
	if opts.PriceNegotiable != nil {
		query = query.Where("price_negotiable = ?", *opts.PriceNegotiable)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR brand_name ILIKE ? OR model_name ILIKE ? OR vin ILIKE ?", like, like, like, like)
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
	case "mileage_asc":
		orderClause = "mileage ASC, id DESC"
	case "year_desc":
		orderClause = "registration_year DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}
	// 精选/甄选置顶优先
	orderClause = "featured DESC, picked DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 管理后台列表（M 端） =====

func (r *carRepository) AdminList(pagination *utils.Pagination, opts CarAdminListOptions) ([]model.Car, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Car
	var total int64

	query := r.db.Model(&model.Car{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		query = query.Where("brand_id = ?", opts.BrandID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.SourceType != "" {
		query = query.Where("source_type = ?", opts.SourceType)
	}
	if opts.CarType != "" {
		query = query.Where("car_type = ?", opts.CarType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR brand_name ILIKE ? OR model_name ILIKE ? OR user_name ILIKE ? OR vin ILIKE ?", like, like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 附近查询（PostGIS 优先，Haversine 降级） =====

// carNearbyResult 用于扫描带 distance 计算列的原始查询结果
type carNearbyResult struct {
	model.Car
	Distance float64 `gorm:"column:distance"`
}

// haversineExprCar 纯 SQL Haversine 公式（返回公里）
const haversineExprCar = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

func (r *carRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts CarListOptions) ([]model.Car, int64, error) {
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

func (r *carRepository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts CarListOptions) ([]model.Car, int64, error) {
	radiusMeters := radiusKm * 1000.0

	where := "deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0 " +
		"AND ST_DWithin(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?)), ?)"
	args := []interface{}{lng, lat, radiusMeters}
	if regionID > 0 {
		where += " AND region_id = ?"
		args = append(args, regionID)
	}
	if opts.CategoryID > 0 {
		where += " AND category_id = ?"
		args = append(args, opts.CategoryID)
	}
	if opts.BrandID > 0 {
		where += " AND brand_id = ?"
		args = append(args, opts.BrandID)
	}
	if opts.CarType != "" {
		where += " AND car_type = ?"
		args = append(args, opts.CarType)
	}
	if opts.MinPrice > 0 {
		where += " AND price >= ?"
		args = append(args, opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		where += " AND price <= ?"
		args = append(args, opts.MaxPrice)
	}

	countArgs := append([]interface{}{}, args...)
	var total int64
	if err := r.db.Model(&model.Car{}).Where(where, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM cars WHERE " + where +
		" ORDER BY featured DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []carNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenCarNearby(rows), total, nil
}

func (r *carRepository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts CarListOptions) ([]model.Car, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.Car{}).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		query = query.Where("brand_id = ?", opts.BrandID)
	}
	if opts.CarType != "" {
		query = query.Where("car_type = ?", opts.CarType)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price <= ?", opts.MaxPrice)
	}

	haversineWhere := haversineExprCar + " <= ?"
	query = query.Where(haversineWhere, lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "*, " + haversineExprCar + " AS distance"
	listQuery := r.db.Table("cars").
		Select(selectSQL, lat, lat, lng).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng).
		Where(haversineWhere, lat, lat, lng, radiusKm)

	if regionID > 0 {
		listQuery = listQuery.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		listQuery = listQuery.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		listQuery = listQuery.Where("brand_id = ?", opts.BrandID)
	}
	if opts.CarType != "" {
		listQuery = listQuery.Where("car_type = ?", opts.CarType)
	}
	if opts.MinPrice > 0 {
		listQuery = listQuery.Where("price >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		listQuery = listQuery.Where("price <= ?", opts.MaxPrice)
	}

	var rows []carNearbyResult
	if err := listQuery.Order("featured DESC, distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenCarNearby(rows), total, nil
}

// flattenCarNearby 把查询结果中的 Distance 回填到 model.Car.Distance
func flattenCarNearby(rows []carNearbyResult) []model.Car {
	list := make([]model.Car, 0, len(rows))
	for _, row := range rows {
		c := row.Car
		c.Distance = row.Distance
		list = append(list, c)
	}
	return list
}

// ===== 搜索 =====

func (r *carRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Car, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Car
	var total int64

	query := r.db.Model(&model.Car{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR brand_name ILIKE ? OR model_name ILIKE ? OR series ILIKE ? OR vin ILIKE ? OR content ILIKE ?", like, like, like, like, like, like)
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

// ===== 用户自己的发布 =====

func (r *carRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Car, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Car
	var total int64

	query := r.db.Model(&model.Car{}).Where("user_id = ?", userID)
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

func (r *carRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *carRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *carRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *carRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *carRepository) IncrShareCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *carRepository) IncrTestDriveCount(id uint) error {
	return r.db.Model(&model.Car{}).Where("id = ?", id).
		UpdateColumn("test_drive_count", gorm.Expr("test_drive_count + 1")).Error
}

// ===== 图片子表 =====

func (r *carRepository) ListImages(carID uint) ([]model.CarImage, error) {
	var imgs []model.CarImage
	if err := r.db.Where("car_id = ?", carID).Order("sort ASC, id ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	return imgs, nil
}

// ReplaceImages 全量替换图片（先删除旧的，再插入新的）
func (r *carRepository) ReplaceImages(carID uint, regionID uint, images []model.CarImage) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("car_id = ?", carID).Delete(&model.CarImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range images {
		images[i].CarID = carID
		images[i].RegionID = regionID
		if images[i].Sort == 0 {
			images[i].Sort = i
		}
		if err := tx.Create(&images[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *carRepository) DeleteImages(carID uint) error {
	return r.db.Where("car_id = ?", carID).Delete(&model.CarImage{}).Error
}

// ===== 收藏 =====

func (r *carRepository) FavExists(userID, carID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.CarFavorite{}).
		Where("user_id = ? AND car_id = ?", userID, carID).
		Count(&count).Error
	return count > 0, err
}

func (r *carRepository) CreateFav(fav *model.CarFavorite) error {
	return r.db.Create(fav).Error
}

func (r *carRepository) DeleteFav(userID, carID uint) error {
	return r.db.Where("user_id = ? AND car_id = ?", userID, carID).
		Delete(&model.CarFavorite{}).Error
}

func (r *carRepository) ListFavs(userID uint, page, pageSize int) ([]model.CarFavorite, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.CarFavorite
	var total int64

	query := r.db.Model(&model.CarFavorite{}).Where("user_id = ?", userID)
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

// HasFavedBatch 批量查询当前用户对一组 car 的收藏状态
func (r *carRepository) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	var favs []model.CarFavorite
	if err := r.db.Where("user_id = ? AND car_id IN ?", userID, ids).Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		result[f.CarID] = true
	}
	return result, nil
}

// ===== 浏览记录 =====

func (r *carRepository) CreateView(v *model.CarView) error {
	return r.db.Create(v).Error
}
