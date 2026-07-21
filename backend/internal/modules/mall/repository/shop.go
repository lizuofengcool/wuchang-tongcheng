// Package repository 同城商城数据访问层 - 店铺
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id 商家）
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ShopRepository 店铺仓储接口
type ShopRepository interface {
	Create(s *model.Shop) error
	FindByID(id uint) (*model.Shop, error)
	FindByUserID(userID uint) (*model.Shop, error)
	Update(s *model.Shop) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts ShopListOptions) ([]model.Shop, int64, error)
	AdminList(pagination *utils.Pagination, opts ShopAdminListOptions) ([]model.Shop, int64, error)
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Shop, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Shop, int64, error)
	ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Shop, int64, error)

	// 计数器
	IncrViewCount(id uint) error
	IncrProductCount(id uint) error
	DecrProductCount(id uint) error
	IncrOrderCount(id uint) error
	IncrSaleAmount(id uint, amount float64) error
	UpdateRating(id uint, rating float64, reviewCount int) error
}

// ShopListOptions C 端店铺列表过滤条件
type ShopListOptions struct {
	ShopType    string
	Keyword     string
	City        string
	District    string
	Featured    *bool
	Verified    *bool
	Status      int // 默认 1（营业），-1 全部
	Sort        string
}

// ShopAdminListOptions M 端管理列表过滤条件
type ShopAdminListOptions struct {
	RegionID    uint
	UserID      uint
	ShopType    string
	Status      *int
	AuditStatus *int
	Featured    *bool
	Verified    *bool
	Keyword     string
}

type shopRepository struct {
	db *gorm.DB
}

// NewShopRepository 创建店铺仓储实例
func NewShopRepository(db *gorm.DB) ShopRepository {
	return &shopRepository{db: db}
}

func (r *shopRepository) Create(s *model.Shop) error {
	return r.db.Create(s).Error
}

func (r *shopRepository) FindByID(id uint) (*model.Shop, error) {
	var s model.Shop
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shopRepository) FindByUserID(userID uint) (*model.Shop, error) {
	var s model.Shop
	if err := r.db.Where("user_id = ?", userID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shopRepository) Update(s *model.Shop) error {
	return r.db.Save(s).Error
}

func (r *shopRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).Updates(fields).Error
}

func (r *shopRepository) Delete(id uint) error {
	return r.db.Delete(&model.Shop{}, id).Error
}

func (r *shopRepository) List(regionID uint, pagination *utils.Pagination, opts ShopListOptions) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status == -1 {
		// 全部状态
	} else if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	} else {
		query = query.Where("status = ?", model.ShopStatusOpened)
	}
	if opts.ShopType != "" {
		query = query.Where("shop_type = ?", opts.ShopType)
	}
	if opts.City != "" {
		query = query.Where("city = ?", opts.City)
	}
	if opts.District != "" {
		query = query.Where("district = ?", opts.District)
	}
	if opts.Featured != nil {
		query = query.Where("featured = ?", *opts.Featured)
	}
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("shop_name ILIKE ? OR description ILIKE ? OR address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := "featured DESC, verified DESC, created_at DESC, id DESC"
	switch opts.Sort {
	case "rating_desc":
		orderClause = "rating DESC, id DESC"
	case "sales_desc":
		orderClause = "sale_amount DESC, id DESC"
	case "popular":
		orderClause = "view_count DESC, id DESC"
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) AdminList(pagination *utils.Pagination, opts ShopAdminListOptions) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopType != "" {
		query = query.Where("shop_type = ?", opts.ShopType)
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
	if opts.Verified != nil {
		query = query.Where("verified = ?", *opts.Verified)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("shop_name ILIKE ? OR contact_name ILIKE ? OR contact_phone ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{}).
		Where("status = ?", model.ShopStatusOpened)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("shop_name ILIKE ? OR description ILIKE ? OR address ILIKE ?", like, like, like)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("featured DESC, created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) ListByCategory(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	query := r.db.Model(&model.Shop{}).Where("status = ?", model.ShopStatusOpened)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	// 通过 product_count 间接关联分类，简化处理：按店铺商品数排序
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("featured DESC, product_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 计数器 =====

func (r *shopRepository) IncrViewCount(id uint) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *shopRepository) IncrProductCount(id uint) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("product_count", gorm.Expr("product_count + 1")).Error
}

func (r *shopRepository) DecrProductCount(id uint) error {
	return r.db.Model(&model.Shop{}).Where("id = ? AND product_count > 0", id).
		UpdateColumn("product_count", gorm.Expr("product_count - 1")).Error
}

func (r *shopRepository) IncrOrderCount(id uint) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("order_count", gorm.Expr("order_count + 1")).Error
}

func (r *shopRepository) IncrSaleAmount(id uint, amount float64) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("sale_amount", gorm.Expr("sale_amount + ?", amount)).Error
}

func (r *shopRepository) UpdateRating(id uint, rating float64, reviewCount int) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"rating":       rating,
			"review_count": reviewCount,
		}).Error
}
