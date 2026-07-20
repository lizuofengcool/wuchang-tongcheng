// Package repository 同城拼车出行数据访问层 - 拼车主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据 v3.2.1 架构方案：对标哈啰出行/嘀嗒出行/滴滴顺风车
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PincheListOptions C 端拼车行程列表过滤条件
type PincheListOptions struct {
	UserID        uint
	TripType      string
	Role          string
	Status        *int
	AuditStatus   *int
	MinPrice      float64
	MaxPrice      float64
	MinSeats      int
	DepartureFrom string
	DepartureTo   string
	PickupCity    string
	DropoffCity   string
	Keyword       string
	Sort          string // latest/price_asc/price_desc/departure_asc/distance/popular
}

// PincheAdminListOptions M 端拼车行程列表过滤条件
type PincheAdminListOptions struct {
	RegionID    uint
	UserID      uint
	TripType    string
	Role        string
	Status      *int
	AuditStatus *int
	Keyword     string
	StartTime   string
	EndTime     string
}

// PincheNearbyOptions 附近查询条件
type PincheNearbyOptions struct {
	Latitude  float64
	Longitude float64
	RadiusKm  float64
	TripType  string
	Role      string
}

// PincheRepository 拼车主表仓储接口
type PincheRepository interface {
	Create(p *model.Pinche) error
	FindByID(id uint) (*model.Pinche, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表查询（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts PincheListOptions) ([]model.Pinche, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts PincheAdminListOptions) ([]model.Pinche, int64, error)
	// 我的发布
	ListMine(userID uint, pagination *utils.Pagination) ([]model.Pinche, int64, error)
	// 附近查询
	ListNearby(regionID uint, pagination *utils.Pagination, opts PincheNearbyOptions) ([]model.Pinche, int64, error)
	// 智能匹配查询
	ListMatch(regionID uint, pagination *utils.Pagination, opts PincheListOptions) ([]model.Pinche, int64, error)

	// 状态/统计
	UpdateStatus(id uint, status int) error
	UpdateAudit(id uint, auditStatus int, reason string) error
	IncrViewCount(id uint) error
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error
	IncrFavCount(id uint, delta int) error
	CountByStatus(regionID uint, status int) (int64, error)
}

type pincheRepository struct {
	db *gorm.DB
}

// NewPincheRepository 创建拼车主表仓储实例
func NewPincheRepository(db *gorm.DB) PincheRepository {
	return &pincheRepository{db: db}
}

// ===== CRUD =====

func (r *pincheRepository) Create(p *model.Pinche) error {
	return r.db.Create(p).Error
}

func (r *pincheRepository) FindByID(id uint) (*model.Pinche, error) {
	var p model.Pinche
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pincheRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).Updates(fields).Error
}

func (r *pincheRepository) Delete(id uint) error {
	return r.db.Delete(&model.Pinche{}, id).Error
}

// ===== 列表查询 =====

func (r *pincheRepository) List(regionID uint, pagination *utils.Pagination, opts PincheListOptions) ([]model.Pinche, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Pinche
	var total int64

	query := r.db.Model(&model.Pinche{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	query = applyPincheFilters(query, opts)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = applyPincheSort(query, opts.Sort)
	if err := query.Scopes(utils.Paginate(pagination)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pincheRepository) AdminList(pagination *utils.Pagination, opts PincheAdminListOptions) ([]model.Pinche, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Pinche
	var total int64

	query := r.db.Model(&model.Pinche{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.TripType != "" {
		query = query.Where("trip_type = ?", opts.TripType)
	}
	if opts.Role != "" {
		query = query.Where("role = ?", opts.Role)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR pickup_location ILIKE ? OR dropoff_location ILIKE ?", like, like, like)
	}
	if opts.StartTime != "" {
		query = query.Where("created_at >= ?", opts.StartTime)
	}
	if opts.EndTime != "" {
		query = query.Where("created_at <= ?", opts.EndTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pincheRepository) ListMine(userID uint, pagination *utils.Pagination) ([]model.Pinche, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Pinche
	var total int64

	query := r.db.Model(&model.Pinche{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pincheRepository) ListNearby(regionID uint, pagination *utils.Pagination, opts PincheNearbyOptions) ([]model.Pinche, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Pinche
	var total int64

	// MVP 简化：使用边界矩形过滤（避免 PostGIS 依赖）
	latDelta := opts.RadiusKm / 111.0
	lngDelta := opts.RadiusKm / (111.0 * 0.9)

	query := r.db.Model(&model.Pinche{}).
		Where("status = ?", model.PincheStatusPublished).
		Where("pickup_lat BETWEEN ? AND ?", opts.Latitude-latDelta, opts.Latitude+latDelta).
		Where("pickup_lng BETWEEN ? AND ?", opts.Longitude-lngDelta, opts.Longitude+lngDelta)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.TripType != "" {
		query = query.Where("trip_type = ?", opts.TripType)
	}
	if opts.Role != "" {
		query = query.Where("role = ?", opts.Role)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *pincheRepository) ListMatch(regionID uint, pagination *utils.Pagination, opts PincheListOptions) ([]model.Pinche, int64, error) {
	// 智能匹配：基于起终点 + 出发时间 + 座位数 + 价格
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Pinche
	var total int64

	query := r.db.Model(&model.Pinche{}).
		Where("status = ?", model.PincheStatusPublished).
		Where("audit_status = ?", model.PincheAuditApproved).
		Where("available_seats >= ?", opts.MinSeats)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	query = applyPincheFilters(query, opts)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 状态/统计 =====

func (r *pincheRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).Update("status", status).Error
}

func (r *pincheRepository) UpdateAudit(id uint, auditStatus int, reason string) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).
		Updates(map[string]interface{}{"audit_status": auditStatus, "audit_reason": reason}).Error
}

func (r *pincheRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *pincheRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *pincheRepository) IncrShareCount(id uint) error {
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *pincheRepository) IncrFavCount(id uint, delta int) error {
	if delta >= 0 {
		return r.db.Model(&model.Pinche{}).Where("id = ?", id).
			UpdateColumn("fav_count", gorm.Expr("fav_count + ?", delta)).Error
	}
	return r.db.Model(&model.Pinche{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("GREATEST(fav_count + ?, 0)", delta)).Error
}

func (r *pincheRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.Pinche{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ===== 内部辅助 =====

func applyPincheFilters(query *gorm.DB, opts PincheListOptions) *gorm.DB {
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.TripType != "" {
		query = query.Where("trip_type = ?", opts.TripType)
	}
	if opts.Role != "" {
		query = query.Where("role = ?", opts.Role)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price_per_seat >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price_per_seat <= ?", opts.MaxPrice)
	}
	if opts.MinSeats > 0 {
		query = query.Where("available_seats >= ?", opts.MinSeats)
	}
	if opts.DepartureFrom != "" {
		query = query.Where("departure_time >= ?", opts.DepartureFrom)
	}
	if opts.DepartureTo != "" {
		query = query.Where("departure_time <= ?", opts.DepartureTo)
	}
	if opts.PickupCity != "" {
		query = query.Where("pickup_location ILIKE ?", "%"+opts.PickupCity+"%")
	}
	if opts.DropoffCity != "" {
		query = query.Where("dropoff_location ILIKE ?", "%"+opts.DropoffCity+"%")
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR pickup_location ILIKE ? OR dropoff_location ILIKE ?", like, like, like)
	}
	return query
}

func applyPincheSort(query *gorm.DB, sort string) *gorm.DB {
	switch sort {
	case "price_asc":
		return query.Order("price_per_seat ASC, id DESC")
	case "price_desc":
		return query.Order("price_per_seat DESC, id DESC")
	case "departure_asc":
		return query.Order("departure_time ASC, id DESC")
	case "distance":
		return query.Order("distance_km ASC, id DESC")
	case "popular":
		return query.Order("view_count DESC, id DESC")
	default: // latest
		return query.Order("created_at DESC, id DESC")
	}
}
