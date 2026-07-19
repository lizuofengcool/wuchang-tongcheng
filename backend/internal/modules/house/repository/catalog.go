// Package repository 房源分类 + 配套设施 + 房贷配置数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== 房源分类 =====

// CategoryRepository 房源分类仓储接口
type CategoryRepository interface {
	Create(c *model.HouseCategory) error
	FindByID(id uint) (*model.HouseCategory, error)
	FindByCode(code string) (*model.HouseCategory, error)
	Update(c *model.HouseCategory) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts CategoryListOptions) ([]model.HouseCategory, int64, error)
	ListAll() ([]model.HouseCategory, error) // 全量树形结构
	ListByParent(parentID uint) ([]model.HouseCategory, error)
	ListByListingType(listingType string) ([]model.HouseCategory, error)
	CountByParent(parentID uint) (int64, error)
	IncrHouseCount(id uint) error
	DecrHouseCount(id uint) error
	BatchUpdateStatus(ids []uint, status int) (int64, error)
}

// CategoryListOptions 分类列表过滤条件
type CategoryListOptions struct {
	ParentID     uint
	Level        *int
	ListingType  string
	PropertyType string
	Status       *int
	Keyword      string
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *model.HouseCategory) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.HouseCategory, error) {
	var c model.HouseCategory
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) FindByCode(code string) (*model.HouseCategory, error) {
	var c model.HouseCategory
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(c *model.HouseCategory) error {
	return r.db.Save(c).Error
}

func (r *categoryRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseCategory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseCategory{}, id).Error
}

func (r *categoryRepository) List(req *utils.Pagination, opts CategoryListOptions) ([]model.HouseCategory, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseCategory
	var total int64

	query := r.db.Model(&model.HouseCategory{})
	if opts.ParentID > 0 {
		query = query.Where("parent_id = ?", opts.ParentID)
	}
	if opts.Level != nil {
		query = query.Where("level = ?", *opts.Level)
	}
	if opts.ListingType != "" {
		query = query.Where("listing_type = ?", opts.ListingType)
	}
	if opts.PropertyType != "" {
		query = query.Where("property_type = ?", opts.PropertyType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) ListAll() ([]model.HouseCategory, error) {
	var list []model.HouseCategory
	if err := r.db.Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByParent(parentID uint) ([]model.HouseCategory, error) {
	var list []model.HouseCategory
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CategoryStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByListingType(listingType string) ([]model.HouseCategory, error) {
	var list []model.HouseCategory
	if err := r.db.Where("listing_type = ? AND status = ?", listingType, model.CategoryStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) CountByParent(parentID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.HouseCategory{}).Where("parent_id = ?", parentID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *categoryRepository) IncrHouseCount(id uint) error {
	return r.db.Model(&model.HouseCategory{}).Where("id = ?", id).
		UpdateColumn("house_count", gorm.Expr("house_count + 1")).Error
}

func (r *categoryRepository) DecrHouseCount(id uint) error {
	return r.db.Model(&model.HouseCategory{}).Where("id = ? AND house_count > 0", id).
		UpdateColumn("house_count", gorm.Expr("house_count - 1")).Error
}

func (r *categoryRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseCategory{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 配套设施 =====

// FacilityRepository 配套设施仓储接口
type FacilityRepository interface {
	Create(f *model.HouseFacility) error
	FindByID(id uint) (*model.HouseFacility, error)
	FindByCode(code string) (*model.HouseFacility, error)
	Update(f *model.HouseFacility) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts FacilityListOptions) ([]model.HouseFacility, int64, error)
	ListAll() ([]model.HouseFacility, error)
	ListByCategory(category string) ([]model.HouseFacility, error)
	ListHot(limit int) ([]model.HouseFacility, error)
	IncrUseCount(id uint) error
	DecrUseCount(id uint) error
	BatchUpdateStatus(ids []uint, status int) (int64, error)
}

// FacilityListOptions 设施列表过滤条件
type FacilityListOptions struct {
	Category string
	Status   *int
	IsHot    *bool
	Keyword  string
}

type facilityRepository struct {
	db *gorm.DB
}

// NewFacilityRepository 创建设施仓储实例
func NewFacilityRepository(db *gorm.DB) FacilityRepository {
	return &facilityRepository{db: db}
}

func (r *facilityRepository) Create(f *model.HouseFacility) error {
	return r.db.Create(f).Error
}

func (r *facilityRepository) FindByID(id uint) (*model.HouseFacility, error) {
	var f model.HouseFacility
	if err := r.db.First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *facilityRepository) FindByCode(code string) (*model.HouseFacility, error) {
	var f model.HouseFacility
	if err := r.db.Where("code = ?", code).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *facilityRepository) Update(f *model.HouseFacility) error {
	return r.db.Save(f).Error
}

func (r *facilityRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseFacility{}).Where("id = ?", id).Updates(fields).Error
}

func (r *facilityRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseFacility{}, id).Error
}

func (r *facilityRepository) List(req *utils.Pagination, opts FacilityListOptions) ([]model.HouseFacility, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseFacility
	var total int64

	query := r.db.Model(&model.HouseFacility{})
	if opts.Category != "" {
		query = query.Where("category = ?", opts.Category)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.IsHot != nil {
		query = query.Where("is_hot = ?", *opts.IsHot)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *facilityRepository) ListAll() ([]model.HouseFacility, error) {
	var list []model.HouseFacility
	if err := r.db.Where("status = ?", model.FacilityStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *facilityRepository) ListByCategory(category string) ([]model.HouseFacility, error) {
	var list []model.HouseFacility
	if err := r.db.Where("category = ? AND status = ?", category, model.FacilityStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *facilityRepository) ListHot(limit int) ([]model.HouseFacility, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.HouseFacility
	if err := r.db.Where("is_hot = ? AND status = ?", true, model.FacilityStatusEnabled).
		Order("use_count DESC, sort ASC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *facilityRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.HouseFacility{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

func (r *facilityRepository) DecrUseCount(id uint) error {
	return r.db.Model(&model.HouseFacility{}).Where("id = ? AND use_count > 0", id).
		UpdateColumn("use_count", gorm.Expr("use_count - 1")).Error
}

func (r *facilityRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseFacility{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 房贷配置 =====

// MortgageRepository 房贷配置仓储接口
type MortgageRepository interface {
	Create(m *model.HouseMortgage) error
	FindByID(id uint) (*model.HouseMortgage, error)
	FindByCode(code string) (*model.HouseMortgage, error)
	Update(m *model.HouseMortgage) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts MortgageListOptions) ([]model.HouseMortgage, int64, error)
	ListAll() ([]model.HouseMortgage, error)
	ListByLoanType(loanType string) ([]model.HouseMortgage, error)
	ListHot(limit int) ([]model.HouseMortgage, error)
	IncrUseCount(id uint) error
	BatchUpdateStatus(ids []uint, status int) (int64, error)
}

// MortgageListOptions 房贷配置列表过滤条件
type MortgageListOptions struct {
	LoanType string
	Status   *int
	IsHot    *bool
	Keyword  string
}

type mortgageRepository struct {
	db *gorm.DB
}

// NewMortgageRepository 创建房贷配置仓储实例
func NewMortgageRepository(db *gorm.DB) MortgageRepository {
	return &mortgageRepository{db: db}
}

func (r *mortgageRepository) Create(m *model.HouseMortgage) error {
	return r.db.Create(m).Error
}

func (r *mortgageRepository) FindByID(id uint) (*model.HouseMortgage, error) {
	var m model.HouseMortgage
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *mortgageRepository) FindByCode(code string) (*model.HouseMortgage, error) {
	var m model.HouseMortgage
	if err := r.db.Where("code = ?", code).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *mortgageRepository) Update(m *model.HouseMortgage) error {
	return r.db.Save(m).Error
}

func (r *mortgageRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseMortgage{}).Where("id = ?", id).Updates(fields).Error
}

func (r *mortgageRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseMortgage{}, id).Error
}

func (r *mortgageRepository) List(req *utils.Pagination, opts MortgageListOptions) ([]model.HouseMortgage, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseMortgage
	var total int64

	query := r.db.Model(&model.HouseMortgage{})
	if opts.LoanType != "" {
		query = query.Where("loan_type = ?", opts.LoanType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.IsHot != nil {
		query = query.Where("is_hot = ?", *opts.IsHot)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *mortgageRepository) ListAll() ([]model.HouseMortgage, error) {
	var list []model.HouseMortgage
	if err := r.db.Where("status = ?", model.MortgageStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mortgageRepository) ListByLoanType(loanType string) ([]model.HouseMortgage, error) {
	var list []model.HouseMortgage
	if err := r.db.Where("loan_type = ? AND status = ?", loanType, model.MortgageStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mortgageRepository) ListHot(limit int) ([]model.HouseMortgage, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.HouseMortgage
	if err := r.db.Where("is_hot = ? AND status = ?", true, model.MortgageStatusEnabled).
		Order("use_count DESC, sort ASC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mortgageRepository) IncrUseCount(id uint) error {
	return r.db.Model(&model.HouseMortgage{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

func (r *mortgageRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseMortgage{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 统计 =====

// StatisticRepository 数据统计仓储接口
type StatisticRepository interface {
	Create(s *model.HouseStatistic) error
	FindByID(id uint) (*model.HouseStatistic, error)
	FindByDateTypeTarget(statDate interface{}, statType string, targetID uint) (*model.HouseStatistic, error)
	Update(s *model.HouseStatistic) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts StatListOptions) ([]model.HouseStatistic, int64, error)
	ListByDateRange(start, end interface{}, statType string, targetID uint) ([]model.HouseStatistic, error)
	ListByType(statType string, req *utils.Pagination) ([]model.HouseStatistic, int64, error)
	GetOverview(regionID uint) (*model.HouseStatistic, error)
	UpsertByDateTypeTarget(s *model.HouseStatistic) error
}

// StatListOptions 统计列表过滤条件
type StatListOptions struct {
	RegionID   uint
	StatType   string
	TargetID   uint
	StartDate  string
	EndDate    string
}

type statisticRepository struct {
	db *gorm.DB
}

// NewStatisticRepository 创建统计仓储实例
func NewStatisticRepository(db *gorm.DB) StatisticRepository {
	return &statisticRepository{db: db}
}

func (r *statisticRepository) Create(s *model.HouseStatistic) error {
	return r.db.Create(s).Error
}

func (r *statisticRepository) FindByID(id uint) (*model.HouseStatistic, error) {
	var s model.HouseStatistic
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) FindByDateTypeTarget(statDate interface{}, statType string, targetID uint) (*model.HouseStatistic, error) {
	var s model.HouseStatistic
	if err := r.db.Where("stat_date = ? AND stat_type = ? AND target_id = ?", statDate, statType, targetID).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *statisticRepository) Update(s *model.HouseStatistic) error {
	return r.db.Save(s).Error
}

func (r *statisticRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseStatistic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *statisticRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseStatistic{}, id).Error
}

func (r *statisticRepository) List(req *utils.Pagination, opts StatListOptions) ([]model.HouseStatistic, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseStatistic
	var total int64

	query := r.db.Model(&model.HouseStatistic{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.StatType != "" {
		query = query.Where("stat_type = ?", opts.StatType)
	}
	if opts.TargetID > 0 {
		query = query.Where("target_id = ?", opts.TargetID)
	}
	if opts.StartDate != "" {
		query = query.Where("stat_date >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("stat_date <= ?", opts.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) ListByDateRange(start, end interface{}, statType string, targetID uint) ([]model.HouseStatistic, error) {
	var list []model.HouseStatistic
	query := r.db.Model(&model.HouseStatistic{}).
		Where("stat_type = ?", statType)
	if targetID > 0 {
		query = query.Where("target_id = ?", targetID)
	}
	if start != nil {
		query = query.Where("stat_date >= ?", start)
	}
	if end != nil {
		query = query.Where("stat_date <= ?", end)
	}
	if err := query.Order("stat_date ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *statisticRepository) ListByType(statType string, req *utils.Pagination) ([]model.HouseStatistic, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseStatistic
	var total int64

	query := r.db.Model(&model.HouseStatistic{}).Where("stat_type = ?", statType)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("stat_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *statisticRepository) GetOverview(regionID uint) (*model.HouseStatistic, error) {
	var s model.HouseStatistic
	if err := r.db.Where("stat_type = ? AND region_id = ?", model.StatTypeOverall, regionID).
		Order("stat_date DESC").First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertByDateTypeTarget 按日期+类型+目标 upsert 统计记录
func (r *statisticRepository) UpsertByDateTypeTarget(s *model.HouseStatistic) error {
	// PostgreSQL ON CONFLICT upsert
	return r.db.Exec(`
		INSERT INTO house_statistics (region_id, stat_date, stat_type, target_id, target_name,
			impression_count, click_count, fav_count, contact_count, viewing_count, deal_count,
			conversion_rate, avg_sale_price, avg_rent_price, avg_deal_days, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (stat_date, stat_type, target_id)
		DO UPDATE SET
			region_id = EXCLUDED.region_id,
			target_name = EXCLUDED.target_name,
			impression_count = EXCLUDED.impression_count,
			click_count = EXCLUDED.click_count,
			fav_count = EXCLUDED.fav_count,
			contact_count = EXCLUDED.contact_count,
			viewing_count = EXCLUDED.viewing_count,
			deal_count = EXCLUDED.deal_count,
			conversion_rate = EXCLUDED.conversion_rate,
			avg_sale_price = EXCLUDED.avg_sale_price,
			avg_rent_price = EXCLUDED.avg_rent_price,
			avg_deal_days = EXCLUDED.avg_deal_days,
			updated_at = NOW()
	`, s.RegionID, s.StatDate, s.StatType, s.TargetID, s.TargetName,
		s.ImpressionCount, s.ClickCount, s.FavCount, s.ContactCount, s.ViewingCount, s.DealCount,
		s.ConversionRate, s.AvgSalePrice, s.AvgRentPrice, s.AvgDealDays).Error
}
