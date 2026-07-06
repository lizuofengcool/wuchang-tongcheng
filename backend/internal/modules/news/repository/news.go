// Package repository 同城分类信息数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/news/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// NewsRepository 分类信息仓储接口
type NewsRepository interface {
	Create(news *model.News) error
	FindByID(id uint) (*model.News, error)
	Update(news *model.News) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, req *utils.Pagination, categoryID uint, status int, listingType string, keyword string, minPrice, maxPrice float64, isUrgent *bool, sort string) ([]model.News, int64, error)
	// ListNearby 附近信息查询：以 (lat,lng) 为中心返回 radiusKm 公里内的已发布信息，
	// 按距离升序（加急置顶）。优先使用 PostGIS ST_DWithin，扩展不可用降级走 Haversine。
	// 返回的 model.News.Distance 已填充（公里）。
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, categoryID uint, listingType string) ([]model.News, int64, error)
	IncrViewCount(id uint) error
	// 点赞
	LikeExists(userID, newsID uint) (bool, error)
	CreateLike(like *model.NewsLike) error
	DeleteLike(userID, newsID uint) error
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	// 收藏
	FavExists(userID, newsID uint) (bool, error)
	CreateFav(fav *model.NewsFavorite) error
	DeleteFav(userID, newsID uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	// 评论
	CreateComment(comment *model.NewsComment) error
	ListComments(newsID uint, page, pageSize int) ([]model.NewsComment, int64, error)
	DeleteComment(id uint) error
	IncrCommentCount(id uint) error
	// 消息
	CreateMessage(msg *model.Message) error
	ListMessages(userID uint, page, pageSize int) ([]model.Message, int64, error)
	UnreadCount(userID uint) (int64, error)
	MarkRead(userID uint, ids []uint) error
	FindByIDs(ids []uint) ([]model.News, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(news *model.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) FindByID(id uint) (*model.News, error) {
	var news model.News
	if err := r.db.First(&news, id).Error; err != nil {
		return nil, err
	}
	return &news, nil
}

func (r *newsRepository) Update(news *model.News) error {
	return r.db.Save(news).Error
}

func (r *newsRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).Updates(fields).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&model.News{}, id).Error
}

func (r *newsRepository) List(regionID uint, pagination *utils.Pagination, categoryID uint, status int, listingType string, keyword string, minPrice, maxPrice float64, isUrgent *bool, sort string) ([]model.News, int64, error) {
	var list []model.News
	var total int64

	query := r.db.Model(&model.News{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if listingType != "" {
		query = query.Where("listing_type = ?", listingType)
	}
	if status >= 0 && status <= 3 {
		query = query.Where("status = ?", status)
	} else {
		query = query.Where("status = ?", 1)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR summary LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}
	if isUrgent != nil && *isUrgent {
		query = query.Where("is_urgent = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch sort {
	case "price":
		orderClause = "price ASC, id DESC"
	case "price_desc":
		orderClause = "price DESC, id DESC"
	case "views":
		orderClause = "view_count DESC, id DESC"
	}

	// 置顶优先
	orderClause = "is_urgent DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *newsRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ====== 附近（PostGIS 空间查询）======

// newsNearbyResult 用于扫描带 distance 计算列的原始查询结果，
// 把计算列映射到独立字段后再回填到 model.News.Distance（gorm:"-" 不会被 Scan 自动映射）。
type newsNearbyResult struct {
	model.News
	Distance float64 `gorm:"column:distance"`
}

// haversineExpr 纯 SQL Haversine 公式（返回公里），参数顺序：lat1, lat1, lng1
// distance = 2R * asin( sqrt( sin²(Δlat/2) + cos(lat1)*cos(lat2)*sin²(Δlng/2) ) )
// 注意：/2 必须在 SIN() 内部，表示先取半角再平方，即 sin²(Δlat/2)。
const haversineExpr = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(latitude)) * " +
	"POWER(SIN(RADIANS(longitude - ?) / 2), 2)" +
	")))"

// ListNearby 附近信息查询。优先 PostGIS ST_DWithin（geography 精确球面距离），
// 扩展不可用时降级走纯 SQL Haversine + 经纬度边界框预筛（走普通索引）。
func (r *newsRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, categoryID uint, listingType string) ([]model.News, int64, error) {
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
		return r.listNearbyPostGIS(regionID, pagination, lat, lng, radiusKm, categoryID, listingType)
	}
	return r.listNearbyHaversine(regionID, pagination, lat, lng, radiusKm, categoryID, listingType)
}

func (r *newsRepository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, categoryID uint, listingType string) ([]model.News, int64, error) {
	radiusMeters := radiusKm * 1000.0

	where := "deleted_at IS NULL AND status = 1 AND latitude <> 0 AND longitude <> 0 " +
		"AND ST_DWithin(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?)), ?)"
	args := []interface{}{lng, lat, radiusMeters}
	if regionID > 0 {
		where += " AND region_id = ?"
		args = append(args, regionID)
	}
	if categoryID > 0 {
		where += " AND category_id = ?"
		args = append(args, categoryID)
	}
	if listingType != "" {
		where += " AND listing_type = ?"
		args = append(args, listingType)
	}

	// 计数
	countArgs := append([]interface{}{}, args...)
	var total int64
	if err := r.db.Model(&model.News{}).Where(where, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 列表（distance 公里）
	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(longitude, latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM news WHERE " + where +
		" ORDER BY is_urgent DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []newsNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenNearby(rows), total, nil
}

func (r *newsRepository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, categoryID uint, listingType string) ([]model.News, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	// 公共过滤：未删除 + 已发布 + 有坐标 + 边界框预筛
	where := "deleted_at IS NULL AND status = 1 AND latitude <> 0 AND longitude <> 0 " +
		"AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?"
	args := []interface{}{minLat, maxLat, minLng, maxLng}
	if regionID > 0 {
		where += " AND region_id = ?"
		args = append(args, regionID)
	}
	if categoryID > 0 {
		where += " AND category_id = ?"
		args = append(args, categoryID)
	}
	if listingType != "" {
		where += " AND listing_type = ?"
		args = append(args, listingType)
	}

	// 精确距离过滤：HAVing 的 Haversine 表达式（含 3 个 SQL 占位符 lat,lat,lng）+ 半径占位符
	// haversineExpr 中的 ? 由 GORM 按顺序绑定，无需 fmt 处理。
	distFilterSQL := " AND " + haversineExpr + " <= ?"
	distFilterArgs := []interface{}{lat, lat, lng, radiusKm}

	// 计数（无 SELECT distance）
	countArgs := append(append([]interface{}{}, args...), distFilterArgs...)
	var total int64
	if err := r.db.Model(&model.News{}).Where(where+distFilterSQL, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 列表：SELECT 中再次计算距离用于排序与回填（参数：lat, lat, lng）
	selectSQL := "SELECT *, " + haversineExpr + " AS distance FROM news WHERE " + where + distFilterSQL +
		" ORDER BY is_urgent DESC, distance ASC, id DESC LIMIT ? OFFSET ?"

	listArgs := []interface{}{lat, lat, lng}            // SELECT distance 计算
	listArgs = append(listArgs, args...)                // 基础过滤
	listArgs = append(listArgs, distFilterArgs...)      // Haversine WHERE
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []newsNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenNearby(rows), total, nil
}

// flattenNearby 将扫描结果中的 Distance 回填到 model.News，并返回纯 News 切片。
func flattenNearby(rows []newsNearbyResult) []model.News {
	list := make([]model.News, 0, len(rows))
	for i := range rows {
		rows[i].News.Distance = rows[i].Distance
		list = append(list, rows[i].News)
	}
	return list
}

// ====== 点赞 ======

func (r *newsRepository) LikeExists(userID, newsID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.NewsLike{}).Where("user_id = ? AND news_id = ?", userID, newsID).Count(&count).Error
	return count > 0, err
}

func (r *newsRepository) CreateLike(like *model.NewsLike) error {
	return r.db.Create(like).Error
}

func (r *newsRepository) DeleteLike(userID, newsID uint) error {
	return r.db.Where("user_id = ? AND news_id = ?", userID, newsID).Delete(&model.NewsLike{}).Error
}

func (r *newsRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *newsRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ? AND like_count > 0", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}

// ====== 收藏 ======

func (r *newsRepository) FavExists(userID, newsID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.NewsFavorite{}).Where("user_id = ? AND news_id = ?", userID, newsID).Count(&count).Error
	return count > 0, err
}

func (r *newsRepository) CreateFav(fav *model.NewsFavorite) error {
	return r.db.Create(fav).Error
}

func (r *newsRepository) DeleteFav(userID, newsID uint) error {
	return r.db.Where("user_id = ? AND news_id = ?", userID, newsID).Delete(&model.NewsFavorite{}).Error
}

func (r *newsRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *newsRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

// ====== 评论 ======

func (r *newsRepository) CreateComment(comment *model.NewsComment) error {
	return r.db.Create(comment).Error
}

func (r *newsRepository) ListComments(newsID uint, page, pageSize int) ([]model.NewsComment, int64, error) {
	var list []model.NewsComment
	var total int64
	query := r.db.Model(&model.NewsComment{}).Where("news_id = ? AND status = 1", newsID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *newsRepository) DeleteComment(id uint) error {
	return r.db.Model(&model.NewsComment{}).Where("id = ?", id).Update("status", 0).Error
}

func (r *newsRepository) IncrCommentCount(id uint) error {
	return r.db.Model(&model.News{}).Where("id = ?", id).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// ====== 消息 ======

func (r *newsRepository) CreateMessage(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *newsRepository) ListMessages(userID uint, page, pageSize int) ([]model.Message, int64, error) {
	var list []model.Message
	var total int64
	query := r.db.Model(&model.Message{}).Where("to_user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *newsRepository) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).Where("to_user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *newsRepository) MarkRead(userID uint, ids []uint) error {
	return r.db.Model(&model.Message{}).Where("to_user_id = ? AND id IN ?", userID, ids).Update("is_read", true).Error
}

func (r *newsRepository) FindByIDs(ids []uint) ([]model.News, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.News
	if err := r.db.Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}