// Package repository 商户中台数据访问层 - 认证
package repository

import (
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// VerificationListOptions 认证列表过滤条件
type VerificationListOptions struct {
	ShopID uint
	Type   string
	Status *int
}

// VerificationRepository 认证仓储接口
type VerificationRepository interface {
	Create(v *model.Verification) error
	FindByID(id uint) (*model.Verification, error)
	Update(v *model.Verification) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts VerificationListOptions) ([]model.Verification, int64, error)
	FindByShopID(shopID uint, pagination *utils.Pagination) ([]model.Verification, int64, error)
	FindByStatus(status int, pagination *utils.Pagination) ([]model.Verification, int64, error)
	FindLatestByShopID(shopID uint) (*model.Verification, error)
}

type verificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository 创建认证仓储实例
func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(v *model.Verification) error {
	return r.db.Create(v).Error
}

func (r *verificationRepository) FindByID(id uint) (*model.Verification, error) {
	var v model.Verification
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *verificationRepository) Update(v *model.Verification) error {
	return r.db.Save(v).Error
}

func (r *verificationRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Verification{}).Where("id = ?", id).Updates(fields).Error
}

func (r *verificationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Verification{}, id).Error
}

func (r *verificationRepository) List(pagination *utils.Pagination, opts VerificationListOptions) ([]model.Verification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Verification
	var total int64

	q := r.db.Model(&model.Verification{})
	if opts.ShopID > 0 {
		q = q.Where("shop_id = ?", opts.ShopID)
	}
	if opts.Type != "" {
		q = q.Where("type = ?", opts.Type)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
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

func (r *verificationRepository) FindByShopID(shopID uint, pagination *utils.Pagination) ([]model.Verification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Verification
	var total int64

	q := r.db.Model(&model.Verification{}).Where("shop_id = ?", shopID)
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

func (r *verificationRepository) FindByStatus(status int, pagination *utils.Pagination) ([]model.Verification, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Verification
	var total int64

	q := r.db.Model(&model.Verification{}).Where("status = ?", status)
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

func (r *verificationRepository) FindLatestByShopID(shopID uint) (*model.Verification, error) {
	var v model.Verification
	if err := r.db.Where("shop_id = ?", shopID).
		Order("created_at DESC").
		First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}
