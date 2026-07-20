// Package repository love 相亲交友数据访问层 - 主表 Love
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/love/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LoveListOptions C 端用户列表过滤条件
type LoveListOptions struct {
	Keyword     string
	Gender      *int
	MinAge      int
	MaxAge      int
	Education   string
	Residence   string
	Hometown    string
	MemberLevel *int
	Status      int // 默认 1（正常），-1 全部
	Featured    *bool
	Picked      *bool
	Verified    *bool
	Sort        string // latest/popular/age/active
}

// LoveAdminListOptions M 端管理列表过滤条件
type LoveAdminListOptions struct {
	RegionID    uint
	UserID      uint
	Gender      *int
	Status      *int
	AuditStatus *int
	MemberLevel *int
	Featured    *bool
	Picked      *bool
	Keyword     string
}

// LoveRepository 主表仓储接口
type LoveRepository interface {
	// CRUD
	Create(l *model.Love) error
	FindByID(id uint) (*model.Love, error)
	FindByUserID(userID uint) (*model.Love, error)
	Update(l *model.Love) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(regionID uint, pagination *utils.Pagination, opts LoveListOptions) ([]model.Love, int64, error)
	AdminList(pagination *utils.Pagination, opts LoveAdminListOptions) ([]model.Love, int64, error)
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts LoveListOptions) ([]model.Love, int64, error)
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Love, int64, error)
	AdvancedSearch(regionID uint, pagination *utils.Pagination, opts LoveListOptions) ([]model.Love, int64, error)

	// 统计更新
	IncrViewCount(id uint) error
	IncrLikeCount(id uint) error
	DecrLikeCount(id uint) error
	IncrLikedCount(id uint) error
	DecrLikedCount(id uint) error
	IncrMatchCount(id uint) error
	DecrMatchCount(id uint) error
	IncrVisitorCount(id uint) error
	IncrStoryCount(id uint) error
	DecrStoryCount(id uint) error
	IncrGiftCount(id uint) error
	IncrImpressionCount(id uint) error
	IncrPopularityScore(id uint, score float64) error

	// 状态
	UpdateStatus(id uint, status int) error
	UpdateAuditStatus(id uint, auditStatus int, reason string) error
	UpdateLocation(id uint, lat, lng float64) error
	UpdateLastActive(id uint, ip string) error
	UpdateMemberLevel(id uint, level int, expiredAt interface{}) error
	UpdateCredits(id uint, credits float64) error
	IncrementCredits(id uint, delta float64) error

	// 推荐/精选
	SetFeatured(id uint, featured bool) error
	SetPicked(id uint, picked bool) error
	UpdateRiskScore(id uint, score int) error

	// 批量
	BatchUpdateStatus(ids []uint, status int) error
	BatchUpdateAuditStatus(ids []uint, auditStatus int, reason string) error
}

type loveRepository struct {
	db *gorm.DB
}

// NewLoveRepository 创建主表仓储实例
func NewLoveRepository(db *gorm.DB) LoveRepository {
	return &loveRepository{db: db}
}

// ===== CRUD =====

func (r *loveRepository) Create(l *model.Love) error {
	return r.db.Create(l).Error
}

func (r *loveRepository) FindByID(id uint) (*model.Love, error) {
	var l model.Love
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveRepository) FindByUserID(userID uint) (*model.Love, error) {
	var l model.Love
	if err := r.db.Where("user_id = ?", userID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *loveRepository) Update(l *model.Love) error {
	return r.db.Save(l).Error
}

func (r *loveRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Updates(fields).Error
}

func (r *loveRepository) Delete(id uint) error {
	return r.db.Delete(&model.Love{}, id).Error
}

// ===== 列表 =====

func (r *loveRepository) List(regionID uint, pagination *utils.Pagination, opts LoveListOptions) ([]model.Love, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Love
	var total int64

	query := r.db.Model(&model.Love{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status == -1 {
		// 全部
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.LoveStatusActive)
	}
	if opts.Gender != nil {
		query = query.Where("gender = ?", *opts.Gender)
	}
	if opts.MinAge > 0 {
		query = query.Where("age >= ?", opts.MinAge)
	}
	if opts.MaxAge > 0 {
		query = query.Where("age <= ?", opts.MaxAge)
	}
	if opts.Education != "" {
		query = query.Where("education = ?", opts.Education)
	}
	if opts.Residence != "" {
		query = query.Where("residence = ?", opts.Residence)
	}
	if opts.Hometown != "" {
		query = query.Where("hometown = ?", opts.Hometown)
	}
	if opts.MemberLevel != nil {
		query = query.Where("member_level = ?", *opts.MemberLevel)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Picked != nil {
		query = query.Where("picked = ?", *opts.Picked)
	}
	if opts.Verified != nil && *opts.Verified {
		query = query.Where("real_name_verified = ?", true)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("nickname ILIKE ? OR residence ILIKE ? OR hometown ILIKE ? OR occupation ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "popularity_score DESC, id DESC"
	switch opts.Sort {
	case "latest":
		orderClause = "created_at DESC, id DESC"
	case "popular":
		orderClause = "popularity_score DESC, view_count DESC, id DESC"
	case "age":
		orderClause = "age ASC, id DESC"
	case "active":
		orderClause = "last_active_at DESC, id DESC"
	}
	orderClause = "featured DESC, picked DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveRepository) AdminList(pagination *utils.Pagination, opts LoveAdminListOptions) ([]model.Love, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Love
	var total int64

	query := r.db.Model(&model.Love{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Gender != nil {
		query = query.Where("gender = ?", *opts.Gender)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.MemberLevel != nil {
		query = query.Where("member_level = ?", *opts.MemberLevel)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Picked != nil {
		query = query.Where("picked = ?", *opts.Picked)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("nickname ILIKE ? OR residence ILIKE ? OR hometown ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts LoveListOptions) ([]model.Love, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Love
	var total int64

	query := r.db.Model(&model.Love{}).Where("status = ?", model.LoveStatusActive)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Gender != nil {
		query = query.Where("gender = ?", *opts.Gender)
	}
	if opts.MinAge > 0 {
		query = query.Where("age >= ?", opts.MinAge)
	}
	if opts.MaxAge > 0 {
		query = query.Where("age <= ?", opts.MaxAge)
	}

	// 简化：直接使用经纬度范围过滤（PostGIS 扩展不可用时的降级）
	if lat > 0 && lng > 0 && radiusKm > 0 {
		// 1 度约 111 公里
		delta := radiusKm / 111.0
		query = query.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
			lat-delta, lat+delta, lng-delta, lng+delta)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("last_active_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *loveRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Love, int64, error) {
	return r.List(regionID, pagination, LoveListOptions{Keyword: keyword, Status: model.LoveStatusActive})
}

func (r *loveRepository) AdvancedSearch(regionID uint, pagination *utils.Pagination, opts LoveListOptions) ([]model.Love, int64, error) {
	return r.List(regionID, pagination, opts)
}

// ===== 统计更新 =====

func (r *loveRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *loveRepository) IncrLikeCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *loveRepository) DecrLikeCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error
}

func (r *loveRepository) IncrLikedCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("liked_count", gorm.Expr("liked_count + 1")).Error
}

func (r *loveRepository) DecrLikedCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("liked_count", gorm.Expr("GREATEST(liked_count - 1, 0)")).Error
}

func (r *loveRepository) IncrMatchCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("match_count", gorm.Expr("match_count + 1")).Error
}

func (r *loveRepository) DecrMatchCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("match_count", gorm.Expr("GREATEST(match_count - 1, 0)")).Error
}

func (r *loveRepository) IncrVisitorCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("visitor_count", gorm.Expr("visitor_count + 1")).Error
}

func (r *loveRepository) IncrStoryCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("story_count", gorm.Expr("story_count + 1")).Error
}

func (r *loveRepository) DecrStoryCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("story_count", gorm.Expr("GREATEST(story_count - 1, 0)")).Error
}

func (r *loveRepository) IncrGiftCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("gift_count", gorm.Expr("gift_count + 1")).Error
}

func (r *loveRepository) IncrImpressionCount(id uint) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("impression_count", gorm.Expr("impression_count + 1")).Error
}

func (r *loveRepository) IncrPopularityScore(id uint, score float64) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("popularity_score", gorm.Expr("popularity_score + ?", score)).Error
}

// ===== 状态 =====

func (r *loveRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Update("status", status).Error
}

func (r *loveRepository) UpdateAuditStatus(id uint, auditStatus int, reason string) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Updates(map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": reason,
	}).Error
}

func (r *loveRepository) UpdateLocation(id uint, lat, lng float64) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latitude":            lat,
		"longitude":           lng,
		"location_updated_at": gorm.Expr("NOW()"),
	}).Error
}

func (r *loveRepository) UpdateLastActive(id uint, ip string) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_active_at": gorm.Expr("NOW()"),
		"last_active_ip": ip,
	}).Error
}

func (r *loveRepository) UpdateMemberLevel(id uint, level int, expiredAt interface{}) error {
	updates := map[string]interface{}{"member_level": level}
	if expiredAt != nil {
		updates["member_expired_at"] = expiredAt
	}
	return r.db.Model(&model.Love{}).Where("id = ?", id).Updates(updates).Error
}

func (r *loveRepository) UpdateCredits(id uint, credits float64) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Update("credits", credits).Error
}

func (r *loveRepository) IncrementCredits(id uint, delta float64) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).UpdateColumn("credits", gorm.Expr("credits + ?", delta)).Error
}

// ===== 推荐/精选 =====

func (r *loveRepository) SetFeatured(id uint, featured bool) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Update("featured", featured).Error
}

func (r *loveRepository) SetPicked(id uint, picked bool) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Update("picked", picked).Error
}

func (r *loveRepository) UpdateRiskScore(id uint, score int) error {
	return r.db.Model(&model.Love{}).Where("id = ?", id).Update("risk_score", score).Error
}

// ===== 批量 =====

func (r *loveRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.Love{}).Where("id IN ?", ids).Update("status", status).Error
}

func (r *loveRepository) BatchUpdateAuditStatus(ids []uint, auditStatus int, reason string) error {
	return r.db.Model(&model.Love{}).Where("id IN ?", ids).Updates(map[string]interface{}{
		"audit_status": auditStatus,
		"audit_reason": reason,
	}).Error
}
