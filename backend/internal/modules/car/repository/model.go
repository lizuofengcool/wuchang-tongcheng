// Package repository 同城车辆买卖数据访问层 - 车型库/分类/品牌聚合
// 车型库与分类为全局数据（无 region_id），管理后台维护
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== CarModel 车型库 =====

// ModelRepository 车型库仓储接口
type ModelRepository interface {
	Create(m *model.CarModel) error
	FindByID(id uint) (*model.CarModel, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query ModelListQuery, pagination *utils.Pagination) ([]model.CarModel, int64, error)
	ListByBrand(brand string, pagination *utils.Pagination) ([]model.CarModel, int64, error)
	ListBySeries(series string, pagination *utils.Pagination) ([]model.CarModel, int64, error)
	IncrUseCount(id uint, delta int) error
}

// ModelListQuery 车型库列表查询
type ModelListQuery struct {
	Brand    string
	Series   string
	CarType  string
	FuelType string
	Year     int
	Keyword  string
	Status   *int
}

type modelRepository struct {
	db *gorm.DB
}

// NewModelRepository 创建车型库仓储实例
func NewModelRepository(db *gorm.DB) ModelRepository {
	return &modelRepository{db: db}
}

func (r *modelRepository) Create(m *model.CarModel) error {
	return r.db.Create(m).Error
}

func (r *modelRepository) FindByID(id uint) (*model.CarModel, error) {
	var m model.CarModel
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *modelRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarModel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *modelRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarModel{}, id).Error
}

func (r *modelRepository) List(query ModelListQuery, pagination *utils.Pagination) ([]model.CarModel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.CarModel
	var total int64

	q := r.db.Model(&model.CarModel{})
	if query.Brand != "" {
		q = q.Where("brand = ?", query.Brand)
	}
	if query.Series != "" {
		q = q.Where("series = ?", query.Series)
	}
	if query.CarType != "" {
		q = q.Where("car_type = ?", query.CarType)
	}
	if query.FuelType != "" {
		q = q.Where("fuel_type = ?", query.FuelType)
	}
	if query.Year > 0 {
		q = q.Where("year = ?", query.Year)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("brand ILIKE ? OR series ILIKE ? OR model_name ILIKE ? OR trim ILIKE ?", like, like, like, like)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
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

func (r *modelRepository) ListByBrand(brand string, pagination *utils.Pagination) ([]model.CarModel, int64, error) {
	return r.List(ModelListQuery{Brand: brand}, pagination)
}

func (r *modelRepository) ListBySeries(series string, pagination *utils.Pagination) ([]model.CarModel, int64, error) {
	return r.List(ModelListQuery{Series: series}, pagination)
}

func (r *modelRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.CarModel{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}

// ===== CarCategory 车型分类 =====

// CategoryRepository 车型分类仓储接口
type CategoryRepository interface {
	Create(c *model.CarCategory) error
	FindByID(id uint) (*model.CarCategory, error)
	FindByCode(code string) (*model.CarCategory, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query CategoryListQuery, pagination *utils.Pagination) ([]model.CarCategory, int64, error)
	ListByParentID(parentID uint) ([]model.CarCategory, error)
	ListByLevel(level int) ([]model.CarCategory, error)
	IncrCarCount(id uint, delta int) error
}

// CategoryListQuery 分类列表查询
type CategoryListQuery struct {
	ParentID uint
	Level    int
	CarType  string
	Status   *int
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建车型分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *model.CarCategory) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindByID(id uint) (*model.CarCategory, error) {
	var c model.CarCategory
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) FindByCode(code string) (*model.CarCategory, error) {
	var c model.CarCategory
	if err := r.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarCategory{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarCategory{}, id).Error
}

func (r *categoryRepository) List(query CategoryListQuery, pagination *utils.Pagination) ([]model.CarCategory, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.CarCategory
	var total int64

	q := r.db.Model(&model.CarCategory{})
	if query.ParentID > 0 {
		q = q.Where("parent_id = ?", query.ParentID)
	}
	if query.Level > 0 {
		q = q.Where("level = ?", query.Level)
	}
	if query.CarType != "" {
		q = q.Where("car_type = ?", query.CarType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *categoryRepository) ListByParentID(parentID uint) ([]model.CarCategory, error) {
	var list []model.CarCategory
	if err := r.db.Where("parent_id = ? AND status = ?", parentID, model.CategoryStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) ListByLevel(level int) ([]model.CarCategory, error) {
	var list []model.CarCategory
	if err := r.db.Where("level = ? AND status = ?", level, model.CategoryStatusPublished).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryRepository) IncrCarCount(id uint, delta int) error {
	return r.db.Model(&model.CarCategory{}).Where("id = ?", id).
		UpdateColumn("car_count", gorm.Expr("car_count + ?", delta)).Error
}

// ===== Brand 品牌聚合（从 car_models 提取） =====

// BrandAggregation 品牌聚合结果
type BrandAggregation struct {
	Brand       string `gorm:"column:brand" json:"brand"`
	BrandLogo   string `gorm:"column:brand_logo" json:"brand_logo"`
	SeriesCount int    `gorm:"column:series_count" json:"series_count"`
	ModelCount  int    `gorm:"column:model_count" json:"model_count"`
	UseCount    int    `gorm:"column:use_count" json:"use_count"`
}

// BrandRepository 品牌聚合仓储接口
type BrandRepository interface {
	ListBrands(query BrandListQuery, pagination *utils.Pagination) ([]BrandAggregation, int64, error)
	ListAllBrands() ([]BrandAggregation, error)
}

// BrandListQuery 品牌列表查询
type BrandListQuery struct {
	Keyword string
	CarType string
}

type brandRepository struct {
	db *gorm.DB
}

// NewBrandRepository 创建品牌聚合仓储实例
func NewBrandRepository(db *gorm.DB) BrandRepository {
	return &brandRepository{db: db}
}

func (r *brandRepository) ListBrands(query BrandListQuery, pagination *utils.Pagination) ([]BrandAggregation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []BrandAggregation
	var total int64

	baseSQL := "FROM car_models WHERE status = ?"
	args := []interface{}{model.ModelStatusPublished}
	if query.CarType != "" {
		baseSQL += " AND car_type = ?"
		args = append(args, query.CarType)
	}
	if query.Keyword != "" {
		baseSQL += " AND brand ILIKE ?"
		args = append(args, "%"+query.Keyword+"%")
	}

	countSQL := "SELECT COUNT(DISTINCT brand) AS total " + baseSQL
	if err := r.db.Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	listSQL := "SELECT brand, MAX(brand_logo) AS brand_logo, COUNT(DISTINCT series) AS series_count, " +
		"COUNT(*) AS model_count, COALESCE(SUM(use_count),0) AS use_count " + baseSQL +
		" GROUP BY brand ORDER BY use_count DESC, brand ASC LIMIT ? OFFSET ?"
	listArgs := append(args, pagination.Limit(), pagination.Offset())
	if err := r.db.Raw(listSQL, listArgs...).Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *brandRepository) ListAllBrands() ([]BrandAggregation, error) {
	var list []BrandAggregation
	sql := "SELECT brand, MAX(brand_logo) AS brand_logo, COUNT(DISTINCT series) AS series_count, " +
		"COUNT(*) AS model_count, COALESCE(SUM(use_count),0) AS use_count " +
		"FROM car_models WHERE status = ? GROUP BY brand ORDER BY use_count DESC, brand ASC"
	if err := r.db.Raw(sql, model.ModelStatusPublished).Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
