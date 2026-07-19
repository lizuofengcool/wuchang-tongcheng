// Package repository 同城招聘求职数据访问层 - 职位主表
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
// 依据需求文档 1.5：PostGIS 必选，扩展不可用降级 Haversine
package repository

import (
	"fmt"

	"wuchang-tongcheng/internal/modules/job/model"
	"wuchang-tongcheng/internal/pkg/geo"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// JobRepository 职位仓储接口
type JobRepository interface {
	// 主表 CRUD
	Create(j *model.Job) error
	FindByID(id uint) (*model.Job, error)
	Update(j *model.Job) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C端，地区隔离 + 过滤）
	List(regionID uint, pagination *utils.Pagination, opts JobListOptions) ([]model.Job, int64, error)
	// 管理后台列表（M端，可跨地区）
	AdminList(pagination *utils.Pagination, opts JobAdminListOptions) ([]model.Job, int64, error)
	// 附近（PostGIS 优先，Haversine 降级）
	ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts JobListOptions) ([]model.Job, int64, error)
	// 搜索（基于 SQL ILIKE，后续可接 Elasticsearch）
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Job, int64, error)
	// 高级搜索（多条件）
	AdvancedSearch(regionID uint, pagination *utils.Pagination, opts JobAdvancedSearchOptions) ([]model.Job, int64, error)
	// 相似职位
	ListSimilar(jobID uint, limit int) ([]model.Job, error)

	// 浏览量
	IncrViewCount(id uint) error
	// 互动统计
	IncrFavCount(id uint) error
	DecrFavCount(id uint) error
	IncrDeliverCount(id uint) error
	IncrInterviewCount(id uint) error
	IncrOfferCount(id uint) error
	IncrMessageCount(id uint) error

	// 图片子表
	ListImages(jobID uint) ([]model.JobImage, error)
	ReplaceImages(jobID uint, urls []string) error
	DeleteImages(jobID uint) error

	// 用户自己的发布
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Job, int64, error)
	// 公司名下职位
	ListByCompany(companyID uint, pagination *utils.Pagination) ([]model.Job, int64, error)
}

// JobListOptions C 端列表过滤条件
type JobListOptions struct {
	CategoryID      uint
	Keyword         string
	RecruitmentType string
	EmploymentType  string
	Education       string
	WorkYearMin     int
	WorkYearMax     int
	SalaryMin       float64
	SalaryMax       float64
	WorkCity        string
	CompanyID       uint
	AllowRemote     *bool
	IsUrgent        *bool
	Featured        *bool
	Verified        *bool
	Sort            string // latest/salary_asc/salary_desc/popular/distance
	Status          int    // 默认 1（已发布），-1 全部
}

// JobAdminListOptions M 端管理列表过滤条件
type JobAdminListOptions struct {
	RegionID    uint
	UserID      uint
	CategoryID  uint
	CompanyID   uint
	Status      *int
	AuditStatus *int
	Keyword     string
}

// JobAdvancedSearchOptions 高级搜索条件
type JobAdvancedSearchOptions struct {
	Keyword         string
	CategoryID      uint
	RecruitmentType string
	EmploymentType  string
	Education       string
	WorkYearMin     int
	WorkYearMax     int
	SalaryMin       float64
	SalaryMax       float64
	WorkCity        string
	CompanyID       uint
	SkillIDs        []uint
	BenefitIDs      []uint
	AllowRemote     *bool
	IsUrgent        *bool
	Featured        *bool
	Verified        *bool
	Sort            string
	Latitude        float64
	Longitude       float64
	RadiusKm        float64
}

type jobRepository struct {
	db *gorm.DB
}

// NewJobRepository 创建仓储实例
func NewJobRepository(db *gorm.DB) JobRepository {
	return &jobRepository{db: db}
}

// ===== 主表 CRUD =====

func (r *jobRepository) Create(j *model.Job) error {
	return r.db.Create(j).Error
}

func (r *jobRepository) FindByID(id uint) (*model.Job, error) {
	var j model.Job
	if err := r.db.First(&j, id).Error; err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *jobRepository) Update(j *model.Job) error {
	return r.db.Save(j).Error
}

func (r *jobRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).Updates(fields).Error
}

func (r *jobRepository) Delete(id uint) error {
	if err := r.db.Delete(&model.Job{}, id).Error; err != nil {
		return err
	}
	return r.db.Where("job_id = ?", id).Delete(&model.JobImage{}).Error
}

// ===== 列表查询（C端） =====

func (r *jobRepository) List(regionID uint, pagination *utils.Pagination, opts JobListOptions) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.StatusPublished).
			Where("audit_status = ?", model.AuditApproved)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ? OR content ILIKE ?", like, like, like)
	}
	if opts.RecruitmentType != "" {
		query = query.Where("recruitment_type = ?", opts.RecruitmentType)
	}
	if opts.EmploymentType != "" {
		query = query.Where("employment_type = ?", opts.EmploymentType)
	}
	if opts.Education != "" {
		query = query.Where("education = ?", opts.Education)
	}
	if opts.WorkYearMin > 0 {
		query = query.Where("work_year_max >= ?", opts.WorkYearMin)
	}
	if opts.WorkYearMax > 0 {
		query = query.Where("work_year_min <= ?", opts.WorkYearMax)
	}
	if opts.SalaryMin > 0 {
		query = query.Where("salary_max >= ?", opts.SalaryMin)
	}
	if opts.SalaryMax > 0 {
		query = query.Where("salary_min <= ?", opts.SalaryMax)
	}
	if opts.WorkCity != "" {
		query = query.Where("work_city ILIKE ?", "%"+opts.WorkCity+"%")
	}
	if opts.CompanyID > 0 {
		query = query.Where("company_id = ?", opts.CompanyID)
	}
	if opts.AllowRemote != nil {
		query = query.Where("allow_remote = ?", *opts.AllowRemote)
	}
	if opts.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *opts.IsUrgent)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "salary_asc":
		orderClause = "salary_min ASC, id DESC"
	case "salary_desc":
		orderClause = "salary_max DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}
	orderClause = "is_urgent DESC, is_top DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 管理后台列表（M端） =====

func (r *jobRepository) AdminList(pagination *utils.Pagination, opts JobAdminListOptions) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.CompanyID > 0 {
		query = query.Where("company_id = ?", opts.CompanyID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR user_name ILIKE ? OR work_city ILIKE ?", like, like, like)
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

type jobNearbyResult struct {
	model.Job
	Distance float64 `gorm:"column:distance"`
}

// haversineExprJob 纯 SQL Haversine 公式（返回公里）
const haversineExprJob = "(6371.0 * 2 * ASIN(SQRT(" +
	"POWER(SIN(RADIANS(work_latitude - ?) / 2), 2) + " +
	"COS(RADIANS(?)) * COS(RADIANS(work_latitude)) * " +
	"POWER(SIN(RADIANS(work_longitude - ?) / 2), 2)" +
	")))"

func (r *jobRepository) ListNearby(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts JobListOptions) ([]model.Job, int64, error) {
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

func (r *jobRepository) listNearbyPostGIS(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts JobListOptions) ([]model.Job, int64, error) {
	radiusMeters := radiusKm * 1000.0

	where := "deleted_at IS NULL AND status = 1 AND audit_status = 1 AND work_latitude <> 0 AND work_longitude <> 0 " +
		"AND ST_DWithin(geography(ST_MakePoint(work_longitude, work_latitude)), geography(ST_MakePoint(?, ?)), ?)"
	args := []interface{}{lng, lat, radiusMeters}
	if regionID > 0 {
		where += " AND region_id = ?"
		args = append(args, regionID)
	}
	if opts.CategoryID > 0 {
		where += " AND category_id = ?"
		args = append(args, opts.CategoryID)
	}
	if opts.RecruitmentType != "" {
		where += " AND recruitment_type = ?"
		args = append(args, opts.RecruitmentType)
	}

	countArgs := append([]interface{}{}, args...)
	var total int64
	if err := r.db.Model(&model.Job{}).Where(where, countArgs...).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "SELECT *, ST_Distance(geography(ST_MakePoint(work_longitude, work_latitude)), geography(ST_MakePoint(?, ?))) / 1000.0 AS distance FROM jobs WHERE " + where +
		" ORDER BY is_urgent DESC, is_top DESC, distance ASC, id DESC LIMIT ? OFFSET ?"
	listArgs := append([]interface{}{lng, lat}, args...)
	listArgs = append(listArgs, pagination.Limit(), pagination.Offset())

	var rows []jobNearbyResult
	if err := r.db.Raw(selectSQL, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenJobNearby(rows), total, nil
}

func (r *jobRepository) listNearbyHaversine(regionID uint, pagination *utils.Pagination, lat, lng, radiusKm float64, opts JobListOptions) ([]model.Job, int64, error) {
	minLat, maxLat, minLng, maxLng := geo.BoundingBox(lat, lng, radiusKm)

	query := r.db.Model(&model.Job{}).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND work_latitude <> 0 AND work_longitude <> 0").
		Where("work_latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("work_longitude BETWEEN ? AND ?", minLng, maxLng)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.RecruitmentType != "" {
		query = query.Where("recruitment_type = ?", opts.RecruitmentType)
	}

	haversineWhere := haversineExprJob + " <= ?"
	query = query.Where(haversineWhere, lat, lat, lng, radiusKm)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	selectSQL := "*, " + haversineExprJob + " AS distance"
	listQuery := r.db.Table("jobs").
		Select(selectSQL, lat, lat, lng).
		Where("deleted_at IS NULL AND status = 1 AND audit_status = 1 AND work_latitude <> 0 AND work_longitude <> 0").
		Where("work_latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("work_longitude BETWEEN ? AND ?", minLng, maxLng).
		Where(haversineWhere, lat, lat, lng, radiusKm)

	if regionID > 0 {
		listQuery = listQuery.Where("region_id = ?", regionID)
	}
	if opts.CategoryID > 0 {
		listQuery = listQuery.Where("category_id = ?", opts.CategoryID)
	}
	if opts.RecruitmentType != "" {
		listQuery = listQuery.Where("recruitment_type = ?", opts.RecruitmentType)
	}

	var rows []jobNearbyResult
	if err := listQuery.Order("is_urgent DESC, is_top DESC, distance ASC, id DESC").
		Limit(pagination.Limit()).Offset(pagination.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	return flattenJobNearby(rows), total, nil
}

func flattenJobNearby(rows []jobNearbyResult) []model.Job {
	list := make([]model.Job, 0, len(rows))
	for _, row := range rows {
		j := row.Job
		j.Distance = row.Distance
		list = append(list, j)
	}
	return list
}

// ===== 搜索 =====

func (r *jobRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ? OR content ILIKE ? OR work_city ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("is_urgent DESC, is_top DESC, published_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 高级搜索 =====

func (r *jobRepository) AdvancedSearch(regionID uint, pagination *utils.Pagination, opts JobAdvancedSearchOptions) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{}).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ? OR content ILIKE ?", like, like, like)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.RecruitmentType != "" {
		query = query.Where("recruitment_type = ?", opts.RecruitmentType)
	}
	if opts.EmploymentType != "" {
		query = query.Where("employment_type = ?", opts.EmploymentType)
	}
	if opts.Education != "" {
		query = query.Where("education = ?", opts.Education)
	}
	if opts.WorkYearMin > 0 {
		query = query.Where("work_year_max >= ?", opts.WorkYearMin)
	}
	if opts.WorkYearMax > 0 {
		query = query.Where("work_year_min <= ?", opts.WorkYearMax)
	}
	if opts.SalaryMin > 0 {
		query = query.Where("salary_max >= ?", opts.SalaryMin)
	}
	if opts.SalaryMax > 0 {
		query = query.Where("salary_min <= ?", opts.SalaryMax)
	}
	if opts.WorkCity != "" {
		query = query.Where("work_city ILIKE ?", "%"+opts.WorkCity+"%")
	}
	if opts.CompanyID > 0 {
		query = query.Where("company_id = ?", opts.CompanyID)
	}
	if opts.AllowRemote != nil {
		query = query.Where("allow_remote = ?", *opts.AllowRemote)
	}
	if opts.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *opts.IsUrgent)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "salary_asc":
		orderClause = "salary_min ASC, id DESC"
	case "salary_desc":
		orderClause = "salary_max DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	case "distance":
		if opts.Latitude != 0 && opts.Longitude != 0 {
			// 使用 fmt.Sprintf 将经纬度参数代入 Haversine 公式
			// 参数为 float64 数值类型，无 SQL 注入风险
			distanceExpr := fmt.Sprintf(
				"(6371.0 * 2 * ASIN(SQRT(POWER(SIN(RADIANS(work_latitude - %f) / 2), 2) + "+
					"COS(RADIANS(%f)) * COS(RADIANS(work_latitude)) * "+
					"POWER(SIN(RADIANS(work_longitude - %f) / 2), 2))))",
				opts.Latitude, opts.Latitude, opts.Longitude,
			)
			orderClause = distanceExpr + " ASC, id DESC"
			query = query.Where("work_latitude <> 0 AND work_longitude <> 0")
			if err := query.Scopes(utils.Paginate(pagination)).
				Order(orderClause).
				Find(&list).Error; err != nil {
				return nil, 0, err
			}
			return list, total, nil
		}
	}
	orderClause = "is_urgent DESC, is_top DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListSimilar 相似职位（同分类/同工作城市）
func (r *jobRepository) ListSimilar(jobID uint, limit int) ([]model.Job, error) {
	if limit <= 0 {
		limit = 5
	}
	var job model.Job
	if err := r.db.First(&job, jobID).Error; err != nil {
		return nil, err
	}
	var list []model.Job
	query := r.db.Model(&model.Job{}).
		Where("id <> ?", jobID).
		Where("status = ?", model.StatusPublished).
		Where("audit_status = ?", model.AuditApproved)
	if job.CategoryID > 0 {
		query = query.Where("category_id = ?", job.CategoryID)
	}
	if err := query.Order("is_urgent DESC, published_at DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== 浏览量/统计 =====

func (r *jobRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *jobRepository) IncrFavCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count + 1")).Error
}

func (r *jobRepository) DecrFavCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ? AND fav_count > 0", id).
		UpdateColumn("fav_count", gorm.Expr("fav_count - 1")).Error
}

func (r *jobRepository) IncrDeliverCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("deliver_count", gorm.Expr("deliver_count + 1")).Error
}

func (r *jobRepository) IncrInterviewCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("interview_count", gorm.Expr("interview_count + 1")).Error
}

func (r *jobRepository) IncrOfferCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("offer_count", gorm.Expr("offer_count + 1")).Error
}

func (r *jobRepository) IncrMessageCount(id uint) error {
	return r.db.Model(&model.Job{}).Where("id = ?", id).
		UpdateColumn("message_count", gorm.Expr("message_count + 1")).Error
}

// ===== 图片子表 =====

func (r *jobRepository) ListImages(jobID uint) ([]model.JobImage, error) {
	var imgs []model.JobImage
	if err := r.db.Where("job_id = ?", jobID).Order("sort ASC, id ASC").Find(&imgs).Error; err != nil {
		return nil, err
	}
	return imgs, nil
}

func (r *jobRepository) ReplaceImages(jobID uint, urls []string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("job_id = ?", jobID).Delete(&model.JobImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i, url := range urls {
		if url == "" {
			continue
		}
		img := model.JobImage{
			JobID: jobID,
			URL:   url,
			Sort:  i,
		}
		if err := tx.Create(&img).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *jobRepository) DeleteImages(jobID uint) error {
	return r.db.Where("job_id = ?", jobID).Delete(&model.JobImage{}).Error
}

// ===== 用户/公司发布列表 =====

func (r *jobRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{}).Where("user_id = ?", userID)
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

func (r *jobRepository) ListByCompany(companyID uint, pagination *utils.Pagination) ([]model.Job, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Job
	var total int64

	query := r.db.Model(&model.Job{}).Where("company_id = ?", companyID)
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
