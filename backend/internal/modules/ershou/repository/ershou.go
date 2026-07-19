// Package repository 同城二手物品数据访问层
// 依据需求文档 1.10：4 维数据隔离（region_id 地区隔离 + user_id 用户隔离）
// 依据需求文档 1.5：PostGIS 必选，扩展不可用降级 Haversine
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ErshouRepository 二手物品仓储接口
type ErshouRepository interface {
	// 主表 CRUD
	Create(e *model.Ershou) error
	FindByID(id uint) (*model.Ershou, error)
	Update(e *model.Ershou) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C端，地区隔离+过滤）
	List(regionID uint, req *utils.Pagination, opts ListOptions) ([]model.Ershou, int64, error)
	// 管理后台列表（M端，可跨地区）
	AdminList(req *utils.Pagination, opts AdminListOptions) ([]model.Ershou, int64, error)
	// 附近（PostGIS 优先，Haversine 降级）
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts ListOptions) ([]model.Ershou, int64, error)
	// 搜索（基于 SQL LIKE，后续可接 Elasticsearch）
	Search(regionID uint, req *utils.Pagination, keyword string) ([]model.Ershou, int64, error)

	// 浏览量
	IncrViewCount(id uint) error

	// 图片子表
	ListImages(ershouID uint) ([]model.ErshouImage, error)
	ReplaceImages(ershouID uint, urls []string) error
	DeleteImages(ershouID uint) error

	// 收藏
	FavExists(userID, ershouID uint) (bool, error)
	CreateFav(fav *model.ErshouFavorite) error
	DeleteFav(userID, ershouID uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	ListFavs(userID uint, page, pageSize int) ([]model.ErshouFavorite, int64, error)
	HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error)

	// 留言
	CreateMessage(msg *model.ErshouMessage) error
	ListMessages(ershouID uint, page, pageSize int) ([]model.ErshouMessage, int64, error)
	IncrMessageCount(id uint) error
	MarkMessagesRead(ershouID uint, userID uint) error

	// 用户自己的发布
	ListByUser(userID uint, req *utils.Pagination) ([]model.Ershou, int64, error)
}

// ListOptions C端列表过滤条件
type ListOptions struct {
	CategoryID uint
	Keyword    string
	MinPrice   float64
	MaxPrice   float64
	Condition  string
	Brand      string
	IsUrgent   *bool
	Sort       string // latest/price_asc/price_desc/popular
	Status     int    // 默认 1（已发布），-1 全部
}

// AdminListOptions M端管理列表过滤条件
// Status/AuditStatus 使用 *int 指针：nil 表示不过滤，非 nil 按值过滤
type AdminListOptions struct {
	RegionID    uint
	UserID      uint
	CategoryID  uint
	Status      *int
	AuditStatus *int
	Keyword     string
}

type ershouRepository struct {
	db *gorm.DB
}

// NewErshouRepository 创建仓储实例
func NewErshouRepository(db *gorm.DB) ErshouRepository {
	return &ershouRepository{db: db}
}

// ===== 主表 CRUD =====

func (r *ershouRepository) Create(e *model.Ershou) error {
	return r.db.Create(e).Error
}

func (r *ershouRepository) FindByID(id uint) (*model.Ershou, error) {
	var e model.Ershou
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ershouRepository) Update(e *model.Ershou) error {
	return r.db.Save(e).Error
}

func (r *ershouRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Ershou{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ershouRepository) Delete(id uint) error {
	// 软删除主表 + 删除关联图片（图片无软删除）
	if err := r.db.Delete(&model.Ershou{}, id).Error; err != nil {
		return err
	}
	return r.db.Where("ershou_id = ?", id).Delete(&model.ErshouImage{}).Error
}

// ===== 列表查询（C端） =====

func (r *ershouRepository) List(regionID uint, pagination *utils.Pagination, opts ListOptions) ([]model.Ershou, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Ershou
	var total int64

	query := r.db.Model(&model.Ershou{})

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
		query = query.Where("status = ?", model.StatusPublished)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR brand ILIKE ? OR summary ILIKE ?", like, like, like)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price <= ?", opts.MaxPrice)
	}
	if opts.Condition != "" {
		query = query.Where("condition = ?", opts.Condition)
	}
	if opts.Brand != "" {
		query = query.Where("brand ILIKE ?", "%"+opts.Brand+"%")
	}
	if opts.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *opts.IsUrgent)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "price_asc":
		orderClause = "price ASC, id DESC"
	case "price_desc":
		orderClause = "price DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}
	// 加急置顶优先
	orderClause = "is_urgent DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 管理后台列表（M端） =====

func (r *ershouRepository) AdminList(pagination *utils.Pagination, opts AdminListOptions) ([]model.Ershou, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Ershou
	var total int64

	query := r.db.Model(&model.Ershou{})

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
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR brand ILIKE ? OR user_name ILIKE ?", like, like, like)
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

// ershouNearbyResult 用于扫描带 distance 计算列的原始查询结果
type ershouNearbyResult struct {
	model.Ershou
	Distance float64 `gorm:"column:distance"`
}

// haversineExpr 纯 SQL Haversine 公式（返回公里）
const haversineExpr = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

func (r *ershouRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts ListOptions) ([]model.Ershou, int64, error) {
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

func (r *ershouRepository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts ListOptions) ([]model.Ershou, int64, error) {
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
	if opts.Condition != "" {
		where += " AND condition = ?"
		args = append(args, opts.Condition)
	}

	countArgs := append([]interface{}{}, args...)
	var total int64
	if err := r.db.Model(&model.Ershou{}).Where(where, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM erhous WHERE " + where +
		" ORDER BY is_urgent DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []ershouNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenNearby(rows), total, nil
}

func (r *ershouRepository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts ListOptions) ([]model.Ershou, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.Ershou{}).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND latitude <> 0 AND longitude <> 0").
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Condition != "" {
		query = query.Where("condition = ?", opts.Condition)
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
	listQuery := r.db.Table("ershous").
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
	if opts.Condition != "" {
		listQuery = listQuery.Where("condition = ?", opts.Condition)
	}

	var rows []ershouNearbyResult
	if err := listQuery.Order("is_urgent DESC, distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenNearby(rows), total, nil
}

// flattenNearby 把查询结果中的 Distance 回填到 model.Ershou.Distance
func flattenNearby(rows []ershouNearbyResult) []model.Ershou {
	list := make([]model.Ershou, 0, len(rows))
	for _, row := range rows {
		e := row.Ershou
		e.Distance = row.Distance
		list = append(list, e)
	}
	return list
}

// ===== 搜索 =====

func (r *ershouRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Ershou, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Ershou
	var total int64

	query := r.db.Model(&model.Ershou{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR brand ILIKE ? OR summary ILIKE ? OR content ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_urgent DESC, published_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 浏览量 =====

func (r *ershouRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Ershou{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ===== 图片子表 =====

func (r *ershouRepository) ListImages(ershouID uint) ([]model.ErshouImage, error) {
	var imgs []model.ErshouImage
	if err := r.db.Where("ershou_id = ?", ershouID).Order("sort ASC, id ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	return imgs, nil
}

// ReplaceImages 全量替换图片（先删除旧的，再插入新的）
func (r *ershouRepository) ReplaceImages(ershouID uint, urls []string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	// 删除旧图片
	if err := tx.Where("ershou_id = ?", ershouID).Delete(&model.ErshouImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	// 插入新图片
	for i, url := range urls {
		if url == "" {
			continue
		}
		img := model.ErshouImage{
			ErshouID: ershouID,
			URL:      url,
			Sort:     i,
		}
		if err := tx.Create(&img).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *ershouRepository) DeleteImages(ershouID uint) error {
	return r.db.Where("ershou_id = ?", ershouID).Delete(&model.ErshouImage{}).Error
}

// ===== 收藏 =====

func (r *ershouRepository) FavExists(userID, ershouID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ErshouFavorite{}).
		Where("user_id = ? AND ershou_id = ?", userID, ershouID).
		Count(&count).Error
	return count > 0, err
}

func (r *ershouRepository) CreateFav(fav *model.ErshouFavorite) error {
	return r.db.Create(fav).Error
}

func (r *ershouRepository) DeleteFav(userID, ershouID uint) error {
	return r.db.Where("user_id = ? AND ershou_id = ?", userID, ershouID).
		Delete(&model.ErshouFavorite{}).Error
}

func (r *ershouRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Ershou{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *ershouRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.Ershou{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *ershouRepository) ListFavs(userID uint, page, pageSize int) ([]model.ErshouFavorite, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.ErshouFavorite
	var total int64

	query := r.db.Model(&model.ErshouFavorite{}).Where("user_id = ?", userID)
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

// HasFavedBatch 批量查询当前用户对一组 ershou 的收藏状态
func (r *ershouRepository) HasFavedBatch(userID uint, ids []uint) (map[uint]bool, error) {
	result := make(map[uint]bool, len(ids))
	if userID == 0 || len(ids) == 0 {
		return result, nil
	}
	var favs []model.ErshouFavorite
	if err := r.db.Where("user_id = ? AND ershou_id IN ?", userID, ids).Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		result[f.ErshouID] = true
	}
	return result, nil
}

// ===== 留言 =====

func (r *ershouRepository) CreateMessage(msg *model.ErshouMessage) error {
	return r.db.Create(msg).Error
}

func (r *ershouRepository) ListMessages(ershouID uint, page, pageSize int) ([]model.ErshouMessage, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.ErshouMessage
	var total int64

	query := r.db.Model(&model.ErshouMessage{}).
		Where("ershou_id = ? AND status = 1", ershouID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ershouRepository) IncrMessageCount(id uint) error {
	return r.db.Model(&model.Ershou{}).Where("id = ?", id).
		UpdateColumn("message_count", gorm.Expr("message_count + 1")).Error
}

// MarkMessagesRead 把发布者收到的某条二手物品的未读留言标记为已读
func (r *ershouRepository) MarkMessagesRead(ershouID uint, userID uint) error {
	// 只标记发给别人（即发布者）的留言为已读
	return r.db.Model(&model.ErshouMessage{}).
		Where("ershou_id = ? AND from_user_id <> ? AND is_read = false", ershouID, userID).
		Update("is_read", true).Error
}

// ===== 用户自己的发布 =====

func (r *ershouRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Ershou, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Ershou
	var total int64

	query := r.db.Model(&model.Ershou{}).Where("user_id = ?", userID)
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
