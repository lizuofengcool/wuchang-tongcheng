// Package repository 同城商城数据访问层 - SKU 规格表
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SkuRepository SKU 仓储接口
type SkuRepository interface {
	Create(s *model.Sku) error
	FindByID(id uint) (*model.Sku, error)
	FindBySkuCode(shopID uint, code string) (*model.Sku, error)
	Update(s *model.Sku) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	DeleteByProduct(productID uint) error

	ListByProduct(productID uint) ([]model.Sku, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Sku, int64, error)
	BatchCreate(skus []model.Sku) error
	ReplaceByProduct(productID uint, skus []model.Sku) error

	// 库存
	UpdateStock(id uint, stock int) error
	IncrSales(id uint, quantity int) error
	DecrStock(id uint, quantity int) error
	IncStock(id uint, quantity int) error
	BatchUpdateStock(items []SkuStockUpdateItem) error
}

// SkuStockUpdateItem SKU 库存更新项
type SkuStockUpdateItem struct {
	SkuID    uint
	Quantity int // 正数加库存，负数减库存
}

type skuRepository struct {
	db *gorm.DB
}

// NewSkuRepository 创建 SKU 仓储实例
func NewSkuRepository(db *gorm.DB) SkuRepository {
	return &skuRepository{db: db}
}

func (r *skuRepository) Create(s *model.Sku) error {
	return r.db.Create(s).Error
}

func (r *skuRepository) FindByID(id uint) (*model.Sku, error) {
	var s model.Sku
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skuRepository) FindBySkuCode(shopID uint, code string) (*model.Sku, error) {
	var s model.Sku
	if err := r.db.Where("shop_id = ? AND sku_code = ?", shopID, code).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *skuRepository) Update(s *model.Sku) error {
	return r.db.Save(s).Error
}

func (r *skuRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Sku{}).Where("id = ?", id).Updates(fields).Error
}

func (r *skuRepository) Delete(id uint) error {
	return r.db.Delete(&model.Sku{}, id).Error
}

func (r *skuRepository) DeleteByProduct(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&model.Sku{}).Error
}

func (r *skuRepository) ListByProduct(productID uint) ([]model.Sku, error) {
	var list []model.Sku
	if err := r.db.Where("product_id = ?", productID).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *skuRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Sku, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Sku
	var total int64

	query := r.db.Model(&model.Sku{}).Where("shop_id = ?", shopID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *skuRepository) BatchCreate(skus []model.Sku) error {
	if len(skus) == 0 {
		return nil
	}
	return r.db.Create(&skus).Error
}

func (r *skuRepository) ReplaceByProduct(productID uint, skus []model.Sku) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("product_id = ?", productID).Delete(&model.Sku{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range skus {
		skus[i].ProductID = productID
		if err := tx.Create(&skus[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *skuRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&model.Sku{}).Where("id = ?", id).
		UpdateColumn("stock", stock).Error
}

func (r *skuRepository) IncrSales(id uint, quantity int) error {
	return r.db.Model(&model.Sku{}).Where("id = ?", id).
		UpdateColumn("sales", gorm.Expr("sales + ?", quantity)).Error
}

func (r *skuRepository) DecrStock(id uint, quantity int) error {
	return r.db.Model(&model.Sku{}).Where("id = ? AND stock >= ?", id, quantity).
		UpdateColumn("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *skuRepository) IncStock(id uint, quantity int) error {
	return r.db.Model(&model.Sku{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).Error
}

func (r *skuRepository) BatchUpdateStock(items []SkuStockUpdateItem) error {
	if len(items) == 0 {
		return nil
	}
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, item := range items {
		var err error
		if item.Quantity >= 0 {
			err = tx.Model(&model.Sku{}).Where("id = ?", item.SkuID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error
		} else {
			err = tx.Model(&model.Sku{}).Where("id = ? AND stock >= ?", item.SkuID, -item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error
		}
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
