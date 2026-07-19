// Package repository SKU 规格数据访问层
// 依据 v3.2.1 架构方案：对标闲鱼/转转 SKU 规格管理
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"

	"gorm.io/gorm"
)

// SKURepository SKU 仓储接口
type SKURepository interface {
	Create(sku *model.ErshouSKU) error
	FindByID(id uint) (*model.ErshouSKU, error)
	ListByErshouID(ershouID uint) ([]model.ErshouSKU, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	DeleteByErshouID(ershouID uint) error
	IncrSoldCount(skuID uint, qty int) error
	DecrStock(skuID uint, qty int) error
	FindByCode(ershouID uint, skuCode string) (*model.ErshouSKU, error)
}

type skuRepository struct {
	db *gorm.DB
}

// NewSKURepository 创建 SKU 仓储实例
func NewSKURepository(db *gorm.DB) SKURepository {
	return &skuRepository{db: db}
}

func (r *skuRepository) Create(sku *model.ErshouSKU) error {
	return r.db.Create(sku).Error
}

func (r *skuRepository) FindByID(id uint) (*model.ErshouSKU, error) {
	var sku model.ErshouSKU
	if err := r.db.First(&sku, id).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuRepository) ListByErshouID(ershouID uint) ([]model.ErshouSKU, error) {
	var list []model.ErshouSKU
	if err := r.db.Where("ershou_id = ?", ershouID).Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *skuRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouSKU{}).Where("id = ?", id).Updates(fields).Error
}

func (r *skuRepository) Delete(id uint) error {
	return r.db.Delete(&model.ErshouSKU{}, id).Error
}

func (r *skuRepository) DeleteByErshouID(ershouID uint) error {
	return r.db.Where("ershou_id = ?", ershouID).Delete(&model.ErshouSKU{}).Error
}

func (r *skuRepository) IncrSoldCount(skuID uint, qty int) error {
	return r.db.Model(&model.ErshouSKU{}).Where("id = ?", skuID).
		Updates(map[string]interface{}{
			"sold_count": gorm.Expr("sold_count + ?", qty),
			"stock":      gorm.Expr("stock - ?", qty),
		}).Error
}

func (r *skuRepository) DecrStock(skuID uint, qty int) error {
	return r.db.Model(&model.ErshouSKU{}).
		Where("id = ? AND stock >= ?", skuID, qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty)).Error
}

func (r *skuRepository) FindByCode(ershouID uint, skuCode string) (*model.ErshouSKU, error) {
	var sku model.ErshouSKU
	if err := r.db.Where("ershou_id = ? AND sku_code = ?", ershouID, skuCode).First(&sku).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}
