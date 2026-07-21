// Package repository LBS地图中台数据访问层 - POI 兴趣点
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：未启用 PostGIS，使用 Haversine 公式降级
package repository

import (
	"wuchang-tongcheng/internal/modules/lbs/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// POIRepository POI 仓储接口
type POIRepository interface {
	// CRUD
	Create(p *model.POI) error
	FindByID(id uint) (*model.POI, error)
	Update(p *model.POI) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(regionID uint, pagination *utils.Pagination, opts POIListOptions) ([]model.POI, int64, error)
	AdminList(pagination *utils.Pagination, opts POIAdminListOptions) ([]model.POI, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.POI, int64, error)
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts POIListOptions) ([]model.POI, int64, error)
}

// POIListOptions C 端 POI 列表过滤条件
type POIListOptions struct {
	Category string
	Keyword  string
	Source   string
	Status   int // 默认 1（上线），-1 全部
	Sort     string
}

// POIAdminListOptions M 端管理列表过滤条件
type POIAdminListOptions struct {
	RegionID uint
	UserID   uint
	Category string
	Source   string
	Status   *int
	Keyword  string
}

type poiRepository struct {
	db *gorm.DB
}

// NewPOIRepository 创建 POI 仓储实例
func NewPOIRepository(db *gorm.DB) POIRepository {
	return &poiRepository{db: db}
}

func (r *poiRepository) Create(p *model.POI) error {
	return r.db.Create(p).Error
}

func (r *poiRepository) FindByID(id uint) (*model.POI, error) {
	var p model.POI
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *poiRepository) Update(p *model.POI) error {
	return r.db.Save(p).Error
}

func (r *poiRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.POI{}).Where("id = ?", id).Updates(fields).Error
}

func (r *poiRepository) Delete(id uint) error {
	return r.db.Delete(&model.POI{}, id).Error
}

func (r *poiRepository) List(regionID uint, pagination *utils.Pagination, opts POIListOptions) ([]model.POI, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.POI
	var total int64

	query := r.db.Model(&model.POI{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.LBSPoiStatusOnline)
	}
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "created_desc":
		orderClause = "created_at DESC, id DESC"
	case "distance_asc":
		orderClause = "id DESC" // 实际距离排序在 ListNearby 中处理
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *poiRepository) AdminList(pagination *utils.Pagination, opts POIAdminListOptions) ([]model.POI, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.POI
	var total int64

	query := r.db.Model(&model.POI{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *poiRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.POI, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.POI
	var total int64

	query := r.db.Model(&model.POI{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// poiNearbyResult 用于扫描带 distance 计算列的原始查询结果
type poiNearbyResult struct {
	model.POI
	Distance float64 `gorm:"column:distance"`
}

// haversineExprPOI 纯 SQL Haversine 公式（返回公里）
const haversineExprPOI = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

func (r *poiRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts POIListOptions) ([]model.POI, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	if radiusKm <= 0 {
		radiusKm = 5
	}
	if radiusKm > 100 {
		radiusKm = 100
	}

	_ = geo.PostGISAvailable(r.db) // 探测但不依赖 PostGIS，统一使用 Haversine 公式（DECIMAL 字段方案）

	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.POI{}).
		Where("deleted_at IS NULL AND status = ? AND latitude <> 0 AND longitude <> 0", model.LBSPoiStatusOnline).
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Source != "" {
		query = query.Where("source = ?", opts.Source)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR address ILIKE ?", like, like)
	}

	haversineWhere := haversineExprPOI + " <= ?"
	query = query.Where(haversineWhere, lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "*, " + haversineExprPOI + " AS distance"
	listQuery := r.db.Table("lbs_pois").
		Select(selectSQL, lat, lat, lng).
		Where("deleted_at IS NULL AND status = ? AND latitude <> 0 AND longitude <> 0", model.LBSPoiStatusOnline).
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng).
		Where(haversineWhere, lat, lat, lng, radiusKm)

	if regionID > 0 {
		listQuery = listQuery.Where("region_id = ?", regionID)
	}
	if opts.Category != "" {
		listQuery = listQuery.Where("category = ?", opts.Category)
	}
	if opts.Source != "" {
		listQuery = listQuery.Where("source = ?", opts.Source)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		listQuery = listQuery.Where("name ILIKE ? OR address ILIKE ?", like, like)
	}

	var rows []poiNearbyResult
	if err := listQuery.Order("distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenPOINearby(rows), total, nil
}

// flattenPOINearby 把查询结果中的 Distance 回填到 model.POI.Distance
func flattenPOINearby(rows []poiNearbyResult) []model.POI {
	list := make([]model.POI, 0, len(rows))
	for _, row := range rows {
		p := row.POI
		p.Distance = row.Distance
		list = append(list, p)
	}
	return list
}
