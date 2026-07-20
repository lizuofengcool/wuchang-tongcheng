// Package repository 同城114数据访问层 - 商户主表 + 图片 + 浏览记录
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：PostGIS 必选，扩展不可用降级 Haversine
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Dh114Repository 商户主表 =====

// Dh114Repository 商户主表仓储接口
type Dh114Repository interface {
	// 主表 CRUD
	Create(d *model.Dh114) error
	FindByID(id uint) (*model.Dh114, error)
	Update(d *model.Dh114) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(regionID uint, pagination *utils.Pagination, opts Dh114ListOptions) ([]model.Dh114, int64, error)
	AdminList(pagination *utils.Pagination, opts Dh114AdminListOptions) ([]model.Dh114, int64, error)
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts Dh114ListOptions) ([]model.Dh114, int64, error)
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Dh114, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error)
	ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error)

	// 计数器
	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error
	IncrCallCount(id uint) error

	// 收藏
	FavExists(userID, dh114ID uint) (bool, error)
	CreateFav(fav *model.Dh114Favorite) error
	DeleteFav(userID, dh114ID uint) error
	ListFavs(userID uint, page, pageSize int) ([]model.Dh114Favorite, int64, error)
	HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error)
}

// Dh114ListOptions C 端商户列表过滤条件
type Dh114ListOptions struct {
	CategoryID   uint
	BusinessType string
	SourceType   string
	Keyword      string
	City         string
	District     string
	MinPrice     float64
	MaxPrice     float64
	MinRating    float64
	Featured     *bool
	Picked       *bool
	Verified     *bool
	Sort         string
	Status       int // 默认 1（已发布），-1 全部
}

// Dh114AdminListOptions M 端管理列表过滤条件
type Dh114AdminListOptions struct {
	RegionID     uint
	UserID       uint
	CategoryID   uint
	Status       *int
	AuditStatus  *int
	BusinessType string
	SourceType   string
	Keyword      string
}

type dh114Repository struct {
	db *gorm.DB
}

// NewDh114Repository 创建商户主表仓储实例
func NewDh114Repository(db *gorm.DB) Dh114Repository {
	return &dh114Repository{db: db}
}

func (r *dh114Repository) Create(d *model.Dh114) error {
	return r.db.Create(d).Error
}

func (r *dh114Repository) FindByID(id uint) (*model.Dh114, error) {
	var d model.Dh114
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *dh114Repository) Update(d *model.Dh114) error {
	return r.db.Save(d).Error
}

func (r *dh114Repository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).Updates(fields).Error
}

func (r *dh114Repository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114{}, id).Error
}

func (r *dh114Repository) List(regionID uint, pagination *utils.Pagination, opts Dh114ListOptions) ([]model.Dh114, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114
	var total int64

	query := r.db.Model(&model.Dh114{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
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
	if opts.BusinessType != "" {
		query = query.Where("business_type = ?", opts.BusinessType)
	}
	if opts.SourceType != "" {
		query = query.Where("source_type = ?", opts.SourceType)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.District != "" {
		query = query.Where("district = ?", opts.District)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price_avg >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price_avg <= ?", opts.MaxPrice)
	}
	if opts.MinRating > 0 {
		query = query.Where("rating >= ?", opts.MinRating)
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
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR address ILIKE ? OR category_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "rating_desc":
		orderClause = "rating DESC, id DESC"
	case "price_asc":
		orderClause = "price_avg ASC, id DESC"
	case "price_desc":
		orderClause = "price_avg DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	case "review_desc":
		orderClause = "review_count DESC, id DESC"
	}
	orderClause = "featured DESC, picked DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dh114Repository) AdminList(pagination *utils.Pagination, opts Dh114AdminListOptions) ([]model.Dh114, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114
	var total int64

	query := r.db.Model(&model.Dh114{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.BusinessType != "" {
		query = query.Where("business_type = ?", opts.BusinessType)
	}
	if opts.SourceType != "" {
		query = query.Where("source_type = ?", opts.SourceType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR user_name ILIKE ? OR address ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// dh114NearbyResult 用于扫描带 distance 计算列的原始查询结果
type dh114NearbyResult struct {
	model.Dh114
	Distance float64 `gorm:"column:distance"`
}

// haversineExprDh114 纯 SQL Haversine 公式（返回公里）
const haversineExprDh114 = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

func (r *dh114Repository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts Dh114ListOptions) ([]model.Dh114, int64, error) {
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

func (r *dh114Repository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts Dh114ListOptions) ([]model.Dh114, int64, error) {
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
	if opts.BusinessType != "" {
		where += " AND business_type = ?"
		args = append(args, opts.BusinessType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		where += " AND (title ILIKE ? OR address ILIKE ?)"
		args = append(args, like, like)
	}

	var total int64
	if err := r.db.Model(&model.Dh114{}).Where(where, args...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM dh114s WHERE " + where +
		" ORDER BY featured DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []dh114NearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenDh114Nearby(rows), total, nil
}

func (r *dh114Repository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts Dh114ListOptions) ([]model.Dh114, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.Dh114{}).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BusinessType != "" {
		query = query.Where("business_type = ?", opts.BusinessType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR address ILIKE ?", like, like)
	}

	haversineWhere := haversineExprDh114 + " <= ?"
	query = query.Where(haversineWhere, lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "*, " + haversineExprDh114 + " AS distance"
	listQuery := r.db.Table("dh114s").
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
	if opts.BusinessType != "" {
		listQuery = listQuery.Where("business_type = ?", opts.BusinessType)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		listQuery = listQuery.Where("title ILIKE ? OR address ILIKE ?", like, like)
	}

	var rows []dh114NearbyResult
	if err := listQuery.Order("featured DESC, distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenDh114Nearby(rows), total, nil
}

// flattenDh114Nearby 把查询结果中的 Distance 回填到 model.Dh114.Distance
func flattenDh114Nearby(rows []dh114NearbyResult) []model.Dh114 {
	list := make([]model.Dh114, 0, len(rows))
	for _, row := range rows {
		d := row.Dh114
		d.Distance = row.Distance
		list = append(list, d)
	}
	return list
}

func (r *dh114Repository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Dh114, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114
	var total int64

	query := r.db.Model(&model.Dh114{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR address ILIKE ? OR category_name ILIKE ? OR phone ILIKE ?", like, like, like, like, like)
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

func (r *dh114Repository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114
	var total int64

	query := r.db.Model(&model.Dh114{}).Where("user_id = ?", userID)
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

func (r *dh114Repository) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Dh114, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114
	var total int64

	query := r.db.Model(&model.Dh114{}).
		Where("status = ? AND audit_status = ?", model.StatusPublished, model.AuditApproved).
		Where("category_id = ?", categoryID)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
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

// ===== 计数器 =====

func (r *dh114Repository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *dh114Repository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *dh114Repository) DecrFavCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *dh114Repository) IncrContactCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *dh114Repository) IncrShareCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *dh114Repository) IncrCallCount(id uint) error {
	return r.db.Model(&model.Dh114{}).Where("id = ?", id).
		UpdateColumn("call_count", gorm.Expr("call_count + 1")).Error
}

// ===== 收藏 =====

func (r *dh114Repository) FavExists(userID, dh114ID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Dh114Favorite{}).
		Where("user_id = ? AND dh114_id = ?", userID, dh114ID).
		Count(&count).Error
	return count > 0, err
}

func (r *dh114Repository) CreateFav(fav *model.Dh114Favorite) error {
	return r.db.Create(fav).Error
}

func (r *dh114Repository) DeleteFav(userID, dh114ID uint) error {
	return r.db.Where("user_id = ? AND dh114_id = ?", userID, dh114ID).
		Delete(&model.Dh114Favorite{}).Error
}

func (r *dh114Repository) ListFavs(userID uint, page, pageSize int) ([]model.Dh114Favorite, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Dh114Favorite
	var total int64

	query := r.db.Model(&model.Dh114Favorite{}).Where("user_id = ?", userID)
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

func (r *dh114Repository) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	var favs []model.Dh114Favorite
	if err := r.db.Where("user_id = ? AND dh114_id IN ?", userID, ids).Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		result[f.Dh114ID] = true
	}
	return result, nil
}

// ===== ImageRepository 商户图片 =====

// ImageRepository 商户图片仓储接口
type ImageRepository interface {
	Create(img *model.Dh114Image) error
	ListByDh114(dh114ID uint) ([]model.Dh114Image, error)
	ReplaceImages(dh114ID uint, images []model.Dh114Image) error
	Delete(id uint) error
	DeleteByDh114(dh114ID uint) error
}

type imageRepository struct {
	db *gorm.DB
}

// NewImageRepository 创建商户图片仓储实例
func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) Create(img *model.Dh114Image) error {
	return r.db.Create(img).Error
}

func (r *imageRepository) ListByDh114(dh114ID uint) ([]model.Dh114Image, error) {
	var imgs []model.Dh114Image
	if err := r.db.Where("dh114_id = ?", dh114ID).Order("sort ASC, id ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	return imgs, nil
}

func (r *imageRepository) ReplaceImages(dh114ID uint, images []model.Dh114Image) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("dh114_id = ?", dh114ID).Delete(&model.Dh114Image{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range images {
		images[i].Dh114ID = dh114ID
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

func (r *imageRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Image{}, id).Error
}

func (r *imageRepository) DeleteByDh114(dh114ID uint) error {
	return r.db.Where("dh114_id = ?", dh114ID).Delete(&model.Dh114Image{}).Error
}

// ===== VisitRepository 浏览记录 =====

// VisitRepository 浏览记录仓储接口
type VisitRepository interface {
	Create(v *model.Dh114Visit) error
	ListByDh114(dh114ID uint, page, pageSize int) ([]model.Dh114Visit, int64, error)
	ListByUser(userID uint, page, pageSize int) ([]model.Dh114Visit, int64, error)
	CountByDh114(dh114ID uint) (int64, error)
}

type visitRepository struct {
	db *gorm.DB
}

// NewVisitRepository 创建浏览记录仓储实例
func NewVisitRepository(db *gorm.DB) VisitRepository {
	return &visitRepository{db: db}
}

func (r *visitRepository) Create(v *model.Dh114Visit) error {
	return r.db.Create(v).Error
}

func (r *visitRepository) ListByDh114(dh114ID uint, page, pageSize int) ([]model.Dh114Visit, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Dh114Visit
	var total int64

	query := r.db.Model(&model.Dh114Visit{}).Where("dh114_id = ?", dh114ID)
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

func (r *visitRepository) ListByUser(userID uint, page, pageSize int) ([]model.Dh114Visit, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Dh114Visit
	var total int64

	query := r.db.Model(&model.Dh114Visit{}).Where("user_id = ?", userID)
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

func (r *visitRepository) CountByDh114(dh114ID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Dh114Visit{}).Where("dh114_id = ?", dh114ID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ===== TagRepository 标签 =====

// TagRepository 标签仓储接口
type TagRepository interface {
	Create(tag *model.Dh114Tag) error
	FindByID(id uint) (*model.Dh114Tag, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query TagListQuery, pagination *utils.Pagination) ([]model.Dh114Tag, int64, error)
	ListByType(tagType string) ([]model.Dh114Tag, error)
	IncrUseCount(id uint) error
}

// TagListQuery 标签列表查询
type TagListQuery struct {
	TagType string
	Keyword string
	Status  *int
}

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(tag *model.Dh114Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) FindByID(id uint) (*model.Dh114Tag, error) {
	var tag model.Dh114Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Tag{}).Where("id = ?", id).Updates(fields).Error
}

func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Tag{}, id).Error
}

func (r *tagRepository) List(query TagListQuery, pagination *utils.Pagination) ([]model.Dh114Tag, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Dh114Tag
	var total int64

	q := r.db.Model(&model.Dh114Tag{})
	if query.TagType != "" {
		q = q.Where("tag_type = ?", query.TagType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, use_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *tagRepository) ListByType(tagType string) ([]model.Dh114Tag, error) {
	var list []model.Dh114Tag
	if err := r.db.Where("tag_type = ? AND status = ?", tagType, 1).
		Order("sort ASC, use_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *tagRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.Dh114Tag{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}
