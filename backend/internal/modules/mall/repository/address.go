// Package repository 同城商城数据访问层 - 收货地址
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// AddressRepository 收货地址仓储接口
type AddressRepository interface {
	Create(a *model.Address) error
	FindByID(id uint) (*model.Address, error)
	Update(a *model.Address) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	ListByUser(userID uint) ([]model.Address, error)
	List(opts AddressListOptions, pagination *utils.Pagination) ([]model.Address, int64, error)
	FindDefault(userID uint) (*model.Address, error)

	ClearDefault(userID uint) error
	SetDefault(userID, id uint) error
	CountByUser(userID uint) (int64, error)
}

// AddressListOptions 地址列表过滤条件
type AddressListOptions struct {
	UserID   uint
	Keyword  string
	RegionID uint
}

type addressRepository struct {
	db *gorm.DB
}

// NewAddressRepository 创建收货地址仓储实例
func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(a *model.Address) error {
	return r.db.Create(a).Error
}

func (r *addressRepository) FindByID(id uint) (*model.Address, error) {
	var a model.Address
	if err := r.db.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *addressRepository) Update(a *model.Address) error {
	return r.db.Save(a).Error
}

func (r *addressRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Address{}).Where("id = ?", id).Updates(fields).Error
}

func (r *addressRepository) Delete(id uint) error {
	return r.db.Delete(&model.Address{}, id).Error
}

func (r *addressRepository) ListByUser(userID uint) ([]model.Address, error) {
	var list []model.Address
	if err := r.db.Where("user_id = ? AND status = 0", userID).
		Order("is_default DESC, updated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *addressRepository) List(opts AddressListOptions, pagination *utils.Pagination) ([]model.Address, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 20)
	}
	var list []model.Address
	var total int64

	query := r.db.Model(&model.Address{})
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR detail ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("is_default DESC, updated_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *addressRepository) FindDefault(userID uint) (*model.Address, error) {
	var a model.Address
	if err := r.db.Where("user_id = ? AND is_default = ? AND status = 0", userID, model.AddressIsDefault).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *addressRepository) ClearDefault(userID uint) error {
	return r.db.Model(&model.Address{}).Where("user_id = ? AND is_default = ?", userID, model.AddressIsDefault).
		UpdateColumn("is_default", model.AddressNotDefault).Error
}

func (r *addressRepository) SetDefault(userID, id uint) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Model(&model.Address{}).Where("user_id = ? AND is_default = ?", userID, model.AddressIsDefault).
		UpdateColumn("is_default", model.AddressNotDefault).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&model.Address{}).Where("id = ? AND user_id = ?", id, userID).
		UpdateColumn("is_default", model.AddressIsDefault).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *addressRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Address{}).Where("user_id = ? AND status = 0", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
