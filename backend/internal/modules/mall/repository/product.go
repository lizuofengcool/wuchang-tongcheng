// Package repository 同城商城数据访问层 - 商品 SPU
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 商家 + shop_id 店铺）
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ProductRepository 商品 SPU 仓储接口
type ProductRepository interface {
	Create(p *model.Product) error
	FindByID(id uint) (*model.Product, error)
	Update(p *model.Product) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts ProductListOptions) ([]model.Product, int64, error)
	AdminList(pagination *utils.Pagination, opts ProductAdminListOptions) ([]model.Product, int64, error)
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Product, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Product, int64, error)
	ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Product, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Product, int64, error)
	ListFeatured(regionID uint, limit int) ([]model.Product, error)
	ListHot(regionID uint, limit int) ([]model.Product, error)
	ListNew(regionID uint, limit int) ([]model.Product, error)

	// 计数器
	IncrViewCount(id uint) error
	IncrFavoriteCount(id uint) error
	DecrFavoriteCount(id uint) error
	IncrSales(id uint, quantity int) error
	IncrReviewCount(id uint) error
	UpdateRating(id uint, rating float64, goodRate float64, reviewCount int) error
	UpdateStock(id uint, stock int) error
	UpdatePriceRange(id uint, minPrice, maxPrice float64) error
}

// ProductListOptions C 端商品列表过滤条件
type ProductListOptions struct {
	ShopID      uint
	CategoryID  uint
	BrandID     uint
	Keyword     string
	ProductType string
	MinPrice    float64
	MaxPrice    float64
	Featured    *bool
	Recommended *bool
	NewArrival  *bool
	HotSale     *bool
	FreeShipping *bool
	Status      int // 默认 1（在售），-1 全部
	Sort        string
}

// ProductAdminListOptions M 端管理列表过滤条件
type ProductAdminListOptions struct {
	RegionID     uint
	UserID       uint
	ShopID       uint
	CategoryID   uint
	BrandID      uint
	ProductType  string
	Status       *int
	AuditStatus  *int
	Featured     *bool
	Recommended  *bool
	Keyword      string
}

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓储实例
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(p *model.Product) error {
	return r.db.Create(p).Error
}

func (r *productRepository) FindByID(id uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) Update(p *model.Product) error {
	return r.db.Save(p).Error
}

func (r *productRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).Updates(fields).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *productRepository) List(regionID uint, pagination *utils.Pagination, opts ProductListOptions) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.ProductStatusOnSale)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		query = query.Where("brand_id = ?", opts.BrandID)
	}
	if opts.ProductType != "" {
		query = query.Where("product_type = ?", opts.ProductType)
	}
	if opts.MinPrice > 0 {
		query = query.Where("price >= ?", opts.MinPrice)
	}
	if opts.MaxPrice > 0 {
		query = query.Where("price <= ?", opts.MaxPrice)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Recommended != nil {
		query = query.Where("recommended = ?", *opts.Recommended)
	}
	if opts.NewArrival != nil {
		query = query.Where("new_arrival = ?", *opts.NewArrival)
	}
	if opts.HotSale != nil {
		query = query.Where("hot_sale = ?", *opts.HotSale)
	}
	if opts.FreeShipping != nil {
		query = query.Where("free_shipping = ?", *opts.FreeShipping)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR subtitle ILIKE ? OR tags::text ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "published_at DESC, id DESC"
	switch opts.Sort {
	case "price_asc":
		orderClause = "price ASC, id DESC"
	case "price_desc":
		orderClause = "price DESC, id DESC"
	case "sales_desc":
		orderClause = "sales DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	case "rating_desc":
		orderClause = "rating DESC, id DESC"
	case "newest":
		orderClause = "published_at DESC, id DESC"
	}
	orderClause = "featured DESC, recommended DESC, " + orderClause

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) AdminList(pagination *utils.Pagination, opts ProductAdminListOptions) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.CategoryID > 0 {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.BrandID > 0 {
		query = query.Where("brand_id = ?", opts.BrandID)
	}
	if opts.ProductType != "" {
		query = query.Where("product_type = ?", opts.ProductType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.AuditStatus != nil {
		query = query.Where("audit_status = ?", *opts.AuditStatus)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Recommended != nil {
		query = query.Where("recommended = ?", *opts.Recommended)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR subtitle ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).
		Where("status = ?", model.ProductStatusOnSale).
		Where("audit_status = ?", model.ProductAuditApproved)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("name ILIKE ? OR subtitle ILIKE ? OR tags::text ILIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("featured DESC, sales DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).
		Where("shop_id = ?", shopID).
		Where("status = ?", model.ProductStatusOnSale)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("featured DESC, sort ASC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).
		Where("status = ?", model.ProductStatusOnSale).
		Where("audit_status = ?", model.ProductAuditApproved).
		Where("category_id = ?", categoryID)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("featured DESC, sales DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Product, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Product
	var total int64

	query := r.db.Model(&model.Product{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *productRepository) ListFeatured(regionID uint, limit int) ([]model.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.Product
	query := r.db.Model(&model.Product{}).
		Where("status = ? AND audit_status = ?", model.ProductStatusOnSale, model.ProductAuditApproved).
		Where("featured = ?", true)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("published_at DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *productRepository) ListHot(regionID uint, limit int) ([]model.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.Product
	query := r.db.Model(&model.Product{}).
		Where("status = ? AND audit_status = ?", model.ProductStatusOnSale, model.ProductAuditApproved).
		Where("hot_sale = ?", true)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("sales DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *productRepository) ListNew(regionID uint, limit int) ([]model.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	var list []model.Product
	query := r.db.Model(&model.Product{}).
		Where("status = ? AND audit_status = ?", model.ProductStatusOnSale, model.ProductAuditApproved).
		Where("new_arrival = ?", true)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if err := query.Order("published_at DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== 计数器 =====

func (r *productRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *productRepository) IncrFavoriteCount(id uint) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

func (r *productRepository) DecrFavoriteCount(id uint) error {
	return r.db.Model(&model.Product{}).Where("id = ? AND favorite_count > 0", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count - 1")).Error
}

func (r *productRepository) IncrSales(id uint, quantity int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"sales": gorm.Expr("sales + ?", quantity),
		}).Error
}

func (r *productRepository) IncrReviewCount(id uint) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("review_count", gorm.Expr("review_count + 1")).Error
}

func (r *productRepository) UpdateRating(id uint, rating float64, goodRate float64, reviewCount int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"rating":       rating,
			"good_rate":    goodRate,
			"review_count": reviewCount,
		}).Error
}

func (r *productRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("stock", stock).Error
}

func (r *productRepository) UpdatePriceRange(id uint, minPrice, maxPrice float64) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"min_price": minPrice,
			"max_price": maxPrice,
		}).Error
}
