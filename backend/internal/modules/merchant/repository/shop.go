// Package repository 商户中台数据访问层 - 店铺
// 依据架构设计 4.4：商家入驻/认领/店铺管理
package repository

import (
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ShopListOptions C 端店铺列表过滤条件
type ShopListOptions struct {
	Keyword    string
	CategoryID uint
	OwnerID    uint
	Status     *int
}

// ShopAdminListOptions M 端管理列表过滤条件
type ShopAdminListOptions struct {
	RegionID   uint
	OwnerID    uint
	CategoryID uint
	Status     *int
	Keyword    string
}

// ShopRepository 店铺仓储接口
type ShopRepository interface {
	// 主表 CRUD
	Create(s *model.Shop) error
	FindByID(id uint) (*model.Shop, error)
	Update(s *model.Shop) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表
	List(regionID uint, pagination *utils.Pagination, opts ShopListOptions) ([]model.Shop, int64, error)
	AdminList(pagination *utils.Pagination, opts ShopAdminListOptions) ([]model.Shop, int64, error)
	Search(regionID uint, pagination *utils.Pagination, keyword string) ([]model.Shop, int64, error)

	// 维度查询
	FindByOwnerID(ownerID uint) ([]model.Shop, error)
	FindByCategoryID(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Shop, int64, error)

	// 信用分调整
	UpdateCreditScore(id uint, delta int) error
	// 等级调整
	UpdateLevel(id uint, level int) error
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

	q := r.db.Model(&model.Shop{}).Where("region_id = ?", regionID)
	if opts.CategoryID > 0 {
		q = q.Where("category_id = ?", opts.CategoryID)
	}
	if opts.OwnerID > 0 {
		q = q.Where("owner_id = ?", opts.OwnerID)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	} else {
		q = q.Where("status = ?", model.ShopStatusActive)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("name ILIKE ? OR intro ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
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

	q := r.db.Model(&model.Shop{})
	if opts.RegionID > 0 {
		q = q.Where("region_id = ?", opts.RegionID)
	}
	if opts.OwnerID > 0 {
		q = q.Where("owner_id = ?", opts.OwnerID)
	}
	if opts.CategoryID > 0 {
		q = q.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		q = q.Where("name ILIKE ? OR intro ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
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

	q := r.db.Model(&model.Shop{}).Where("region_id = ?", regionID).
		Where("status = ?", model.ShopStatusActive)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name ILIKE ? OR intro ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) FindByOwnerID(ownerID uint) ([]model.Shop, error) {
	var list []model.Shop
	if err := r.db.Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *shopRepository) FindByCategoryID(regionID uint, categoryID uint, pagination *utils.Pagination) ([]model.Shop, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Shop
	var total int64

	q := r.db.Model(&model.Shop{}).
		Where("region_id = ?", regionID).
		Where("category_id = ?", categoryID).
		Where("status = ?", model.ShopStatusActive)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *shopRepository) UpdateCreditScore(id uint, delta int) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("credit_score", gorm.Expr("credit_score + ?", delta)).Error
}

func (r *shopRepository) UpdateLevel(id uint, level int) error {
	return r.db.Model(&model.Shop{}).Where("id = ?", id).
		UpdateColumn("level", level).Error
}
