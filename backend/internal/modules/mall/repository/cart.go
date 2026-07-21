// Package repository 同城商城数据访问层 - 购物车
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"

	"gorm.io/gorm"
)

// CartRepository 购物车仓储接口
type CartRepository interface {
	Create(c *model.Cart) error
	FindByID(id uint) (*model.Cart, error)
	Update(c *model.Cart) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 查询用户购物车
	ListByUser(userID uint) ([]model.Cart, error)
	ListByUserAndShop(userID, shopID uint) ([]model.Cart, error)
	ListSelected(userID uint) ([]model.Cart, error)
	FindByUserAndSku(userID, skuID uint) (*model.Cart, error)

	// 批量操作
	BatchDelete(ids []uint) error
	DeleteByUser(userID uint) error
	DeleteByUserAndShop(userID, shopID uint) error
	SelectAll(userID uint, selected int) error
	SelectItems(userID uint, ids []uint, selected int) error

	// 计数
	CountByUser(userID uint) (int64, error)
	CountSelectedByUser(userID uint) (int64, error)
}

type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository 创建购物车仓储实例
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Create(c *model.Cart) error {
	return r.db.Create(c).Error
}

func (r *cartRepository) FindByID(id uint) (*model.Cart, error) {
	var c model.Cart
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *cartRepository) Update(c *model.Cart) error {
	return r.db.Save(c).Error
}

func (r *cartRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Cart{}).Where("id = ?", id).Updates(fields).Error
}

func (r *cartRepository) Delete(id uint) error {
	return r.db.Delete(&model.Cart{}, id).Error
}

func (r *cartRepository) ListByUser(userID uint) ([]model.Cart, error) {
	var list []model.Cart
	if err := r.db.Where("user_id = ? AND status = 0", userID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *cartRepository) ListByUserAndShop(userID, shopID uint) ([]model.Cart, error) {
	var list []model.Cart
	if err := r.db.Where("user_id = ? AND shop_id = ? AND status = 0", userID, shopID).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *cartRepository) ListSelected(userID uint) ([]model.Cart, error) {
	var list []model.Cart
	if err := r.db.Where("user_id = ? AND selected = ? AND status = 0", userID, model.CartSelected).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *cartRepository) FindByUserAndSku(userID, skuID uint) (*model.Cart, error) {
	var c model.Cart
	if err := r.db.Where("user_id = ? AND sku_id = ? AND status = 0", userID, skuID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *cartRepository) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.Cart{}).Error
}

func (r *cartRepository) DeleteByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.Cart{}).Error
}

func (r *cartRepository) DeleteByUserAndShop(userID, shopID uint) error {
	return r.db.Where("user_id = ? AND shop_id = ?", userID, shopID).Delete(&model.Cart{}).Error
}

func (r *cartRepository) SelectAll(userID uint, selected int) error {
	return r.db.Model(&model.Cart{}).Where("user_id = ? AND status = 0", userID).
		UpdateColumn("selected", selected).Error
}

func (r *cartRepository) SelectItems(userID uint, ids []uint, selected int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Cart{}).Where("user_id = ? AND id IN ?", userID, ids).
		UpdateColumn("selected", selected).Error
}

func (r *cartRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Cart{}).Where("user_id = ? AND status = 0", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *cartRepository) CountSelectedByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Cart{}).Where("user_id = ? AND selected = ? AND status = 0", userID, model.CartSelected).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
