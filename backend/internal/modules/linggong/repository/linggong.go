// Package repository 同城零工兼职数据访问层 - 岗位主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// LinggongRepository 岗位主表仓储接口
type LinggongRepository interface {
	Create(l *model.Linggong) error
	FindByID(id uint) (*model.Linggong, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表查询
	List(regionID uint, pagination *utils.Pagination, opts LinggongListOptions) ([]model.Linggong, int64, error)
	AdminList(pagination *utils.Pagination, opts LinggongAdminListOptions) ([]model.Linggong, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Linggong, int64, error)
	ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.Linggong, int64, error)
	Search(regionID uint, keyword string, pagination *utils.Pagination) ([]model.Linggong, int64, error)
	Nearby(regionID uint, lat, lng, radiusKm float64, pagination *utils.Pagination) ([]model.Linggong, int64, error)

	// 计数器
	IncrViewCount(id uint) error
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	IncrContactCount(id uint) error
	IncrShareCount(id uint) error
	IncrApplicationCount(id uint) error

	// 批量操作
	BatchUpdateStatus(ids []uint, status int) error
}

// LinggongListOptions C 端岗位列表过滤条件
type LinggongListOptions struct {
	LinggongType      string
	PublisherType     string
	BillingType       string
	Settlement        string
	MinSalary         float64
	MaxSalary         float64
	Status            *int
	AuditStatus       *int
	EmployerID        uint
	Province          string
	City              string
	District          string
	WorkLocationType  string
	Featured          *bool
	Picked            *bool
	Verified          *bool
	EmployerVerified  *bool
	NeedGender        string
	Education         string
	Keyword           string
	Sort              string
}

// LinggongAdminListOptions M 端岗位列表过滤条件
type LinggongAdminListOptions struct {
	RegionID      uint
	UserID        uint
	EmployerID    uint
	LinggongType  string
	PublisherType string
	BillingType   string
	Status        *int
	AuditStatus   *int
	Keyword       string
}

type linggongRepository struct {
	db *gorm.DB
}

// NewLinggongRepository 创建岗位主表仓储实例
func NewLinggongRepository(db *gorm.DB) LinggongRepository {
	return &linggongRepository{db: db}
}

// ===== CRUD =====

func (r *linggongRepository) Create(l *model.Linggong) error {
	return r.db.Create(l).Error
}

func (r *linggongRepository) FindByID(id uint) (*model.Linggong, error) {
	var l model.Linggong
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *linggongRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).Updates(fields).Error
}

func (r *linggongRepository) Delete(id uint) error {
	return r.db.Delete(&model.Linggong{}, id).Error
}

// ===== 列表查询 =====

func (r *linggongRepository) List(regionID uint, pagination *utils.Pagination, opts LinggongListOptions) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Linggong
	var total int64

	query := r.db.Model(&model.Linggong{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongType != "" {
		query = query.Where("linggong_type = ?", opts.LinggongType)
	}
	if opts.PublisherType != "" {
		query = query.Where("publisher_type = ?", opts.PublisherType)
	}
	if opts.BillingType != "" {
		query = query.Where("billing_type = ?", opts.BillingType)
	}
	if opts.Settlement != "" {
		query = query.Where("settlement = ?", opts.Settlement)
	}
	if opts.MinSalary > 0 {
		query = query.Where("salary_max >= ?", opts.MinSalary)
	}
	if opts.MaxSalary > 0 {
		query = query.Where("salary_min <= ?", opts.MaxSalary)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.Province != "" {
		query = query.Where("province = ?", opts.Province)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.District != "" {
		query = query.Where("district = ?", opts.District)
	}
	if opts.WorkLocationType != "" {
		query = query.Where("work_location_type = ?", opts.WorkLocationType)
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
	if opts.EmployerVerified != nil {
		query = query.Where("employer_verified = ?", *opts.EmployerVerified)
	}
	if opts.NeedGender != "" {
		query = query.Where("need_gender = ?", opts.NeedGender)
	}
	if opts.Education != "" {
		query = query.Where("education = ?", opts.Education)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR company_name ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderQuery := query
	switch opts.Sort {
	case "salary_asc":
		orderQuery = orderQuery.Order("salary_min ASC, id DESC")
	case "salary_desc":
		orderQuery = orderQuery.Order("salary_max DESC, id DESC")
	case "popular":
		orderQuery = orderQuery.Order("view_count DESC, fav_count DESC, id DESC")
	default:
		orderQuery = orderQuery.Order("featured DESC, published_at DESC, id DESC")
	}

	if err := orderQuery.Scopes(utils.Paginate(pagination)).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *linggongRepository) AdminList(pagination *utils.Pagination, opts LinggongAdminListOptions) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Linggong
	var total int64

	query := r.db.Model(&model.Linggong{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.LinggongType != "" {
		query = query.Where("linggong_type = ?", opts.LinggongType)
	}
	if opts.PublisherType != "" {
		query = query.Where("publisher_type = ?", opts.PublisherType)
	}
	if opts.BillingType != "" {
		query = query.Where("billing_type = ?", opts.BillingType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR company_name ILIKE ? OR contact_name ILIKE ?", like, like, like)
	}

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

func (r *linggongRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Linggong
	var total int64

	query := r.db.Model(&model.Linggong{}).Where("user_id = ?", userID)
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

func (r *linggongRepository) ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Linggong
	var total int64

	query := r.db.Model(&model.Linggong{}).Where("employer_id = ?", employerID)
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

func (r *linggongRepository) Search(regionID uint, keyword string, pagination *utils.Pagination) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Linggong
	var total int64

	query := r.db.Model(&model.Linggong{}).Where("status = ? AND audit_status = ?", 1, 1)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR company_name ILIKE ?", like, like, like)
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

func (r *linggongRepository) Nearby(regionID uint, lat, lng, radiusKm float64, pagination *utils.Pagination) ([]model.Linggong, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	if radiusKm <= 0 {
		radiusKm = 10
	}
	var list []model.Linggong
	var total int64

	// 粗略过滤：经纬度差值（1 度约 111 公里）
	delta := radiusKm / 111.0
	query := r.db.Model(&model.Linggong{}).
		Where("status = ? AND audit_status = ?", 1, 1).
		Where("latitude BETWEEN ? AND ?", lat-delta, lat+delta).
		Where("longitude BETWEEN ? AND ?", lng-delta, lng+delta)
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

func (r *linggongRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *linggongRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *linggongRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *linggongRepository) IncrContactCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error
}

func (r *linggongRepository) IncrShareCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).
		UpdateColumn("share_count", gorm.Expr("share_count + 1")).Error
}

func (r *linggongRepository) IncrApplicationCount(id uint) error {
	return r.db.Model(&model.Linggong{}).Where("id = ?", id).
		UpdateColumn("application_count", gorm.Expr("application_count + 1")).Error
}

// ===== 批量操作 =====

func (r *linggongRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.Linggong{}).Where("id IN ?", ids).
		Update("status", status).Error
}
