// Package repository 标签/品牌/型号/分类属性数据访问层
// 依据 v3.2.1 架构方案：对标转转
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Tag =====

// TagRepository 标签仓储接口
type TagRepository interface {
	Create(t *model.ErshouTag) error
	FindByID(id uint) (*model.ErshouTag, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query TagListQuery, pagination *utils.Pagination) ([]model.ErshouTag, int64, error)
	ListHot(limit int) ([]model.ErshouTag, error)
	IncrUseCount(id uint, delta int) error
}

// TagListQuery 标签列表查询
type TagListQuery struct {
	Type   string
	Status *int
	IsHot  *bool
}

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(t *model.ErshouTag) error {
	return r.db.Create(t).Error
}

func (r *tagRepository) FindByID(id uint) (*model.ErshouTag, error) {
	var t model.ErshouTag
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tagRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouTag{}).Where("id = ?", id).Updates(fields).Error
}

func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouTag{}, id).Error
}

func (r *tagRepository) List(query TagListQuery, pagination *utils.Pagination) ([]model.ErshouTag, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.ErshouTag
	var total int64

	q := r.db.Model(&model.ErshouTag{})
	if query.Type != "" {
		q = q.Where("type = ?", query.Type)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.IsHot != nil {
		q = q.Where("is_hot = ?", *query.IsHot)
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

func (r *tagRepository) ListHot(limit int) ([]model.ErshouTag, error) {
	var list []model.ErshouTag
	if limit <= 0 {
		limit = 20
	}
	if err := r.db.Where("status = ? AND is_hot = ?", model.AuditRuleStatusEnabled, true).
		Order("use_count DESC, sort ASC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *tagRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.ErshouTag{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}

// ===== Brand =====

// BrandRepository 品牌仓储接口
type BrandRepository interface {
	Create(b *model.ErshouBrand) error
	FindByID(id uint) (*model.ErshouBrand, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(query BrandListQuery, pagination *utils.Pagination) ([]model.ErshouBrand, int64, error)
	FindByName(name string) (*model.ErshouBrand, error)
	IncrUseCount(id uint, delta int) error
}

// BrandListQuery 品牌列表查询
type BrandListQuery struct {
	Keyword string
	Status  *int
}

type brandRepository struct {
	db *gorm.DB
}

// NewBrandRepository 创建品牌仓储实例
func NewBrandRepository(db *gorm.DB) BrandRepository {
	return &brandRepository{db: db}
}

func (r *brandRepository) Create(b *model.ErshouBrand) error {
	return r.db.Create(b).Error
}

func (r *brandRepository) FindByID(id uint) (*model.ErshouBrand, error) {
	var b model.ErshouBrand
	if err := r.db.First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *brandRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouBrand{}).Where("id = ?", id).Updates(fields).Error
}

func (r *brandRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouBrand{}, id).Error
}

func (r *brandRepository) List(query BrandListQuery, pagination *utils.Pagination) ([]model.ErshouBrand, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.ErshouBrand
	var total int64

	q := r.db.Model(&model.ErshouBrand{})
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR english_name ILIKE ?", like, like)
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

func (r *brandRepository) FindByName(name string) (*model.ErshouBrand, error) {
	var b model.ErshouBrand
	if err := r.db.Where("name = ?", name).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *brandRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.ErshouBrand{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}

// ===== Model =====

// ModelRepository 型号仓储接口
type ModelRepository interface {
	Create(m *model.ErshouModel) error
	FindByID(id uint) (*model.ErshouModel, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	ListByBrandID(brandID uint, pagination *utils.Pagination) ([]model.ErshouModel, int64, error)
	List(query ModelListQuery, pagination *utils.Pagination) ([]model.ErshouModel, int64, error)
	IncrUseCount(id uint, delta int) error
}

// ModelListQuery 型号列表查询
type ModelListQuery struct {
	BrandID uint
	Keyword string
	Status  *int
}

type modelRepository struct {
	db *gorm.DB
}

// NewModelRepository 创建型号仓储实例
func NewModelRepository(db *gorm.DB) ModelRepository {
	return &modelRepository{db: db}
}

func (r *modelRepository) Create(m *model.ErshouModel) error {
	return r.db.Create(m).Error
}

func (r *modelRepository) FindByID(id uint) (*model.ErshouModel, error) {
	var m model.ErshouModel
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *modelRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouModel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *modelRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouModel{}, id).Error
}

func (r *modelRepository) ListByBrandID(brandID uint, pagination *utils.Pagination) ([]model.ErshouModel, int64, error) {
	return r.List(ModelListQuery{BrandID: brandID}, pagination)
}

func (r *modelRepository) List(query ModelListQuery, pagination *utils.Pagination) ([]model.ErshouModel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.ErshouModel
	var total int64

	q := r.db.Model(&model.ErshouModel{})
	if query.BrandID > 0 {
		q = q.Where("brand_id = ?", query.BrandID)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("name ILIKE ? OR full_name ILIKE ?", like, like)
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

func (r *modelRepository) IncrUseCount(id uint, delta int) error {
	return r.db.Model(&model.ErshouModel{}).Where("id = ?", id).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", delta)).Error
}

// ===== CategoryAttr =====

// CategoryAttrRepository 分类属性仓储接口
type CategoryAttrRepository interface {
	Create(a *model.ErshouCategoryAttr) error
	FindByID(id uint) (*model.ErshouCategoryAttr, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	ListByCategoryID(categoryID uint) ([]model.ErshouCategoryAttr, error)
	List(query CategoryAttrListQuery, pagination *utils.Pagination) ([]model.ErshouCategoryAttr, int64, error)
}

// CategoryAttrListQuery 分类属性列表查询
type CategoryAttrListQuery struct {
	CategoryID uint
	Status     *int
}

type categoryAttrRepository struct {
	db *gorm.DB
}

// NewCategoryAttrRepository 创建分类属性仓储实例
func NewCategoryAttrRepository(db *gorm.DB) CategoryAttrRepository {
	return &categoryAttrRepository{db: db}
}

func (r *categoryAttrRepository) Create(a *model.ErshouCategoryAttr) error {
	return r.db.Create(a).Error
}

func (r *categoryAttrRepository) FindByID(id uint) (*model.ErshouCategoryAttr, error) {
	var a model.ErshouCategoryAttr
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *categoryAttrRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouCategoryAttr{}).Where("id = ?", id).Updates(fields).Error
}

func (r *categoryAttrRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouCategoryAttr{}, id).Error
}

func (r *categoryAttrRepository) ListByCategoryID(categoryID uint) ([]model.ErshouCategoryAttr, error) {
	var list []model.ErshouCategoryAttr
	if err := r.db.Where("category_id = ? AND status = ?", categoryID, model.AuditRuleStatusEnabled).
		Order("sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *categoryAttrRepository) List(query CategoryAttrListQuery, pagination *utils.Pagination) ([]model.ErshouCategoryAttr, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 50)
	}
	var list []model.ErshouCategoryAttr
	var total int64

	q := r.db.Model(&model.ErshouCategoryAttr{})
	if query.CategoryID > 0 {
		q = q.Where("category_id = ?", query.CategoryID)
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
