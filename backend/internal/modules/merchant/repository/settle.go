// Package repository 商户中台数据访问层 - 结算
package repository

import (
	"wuchang-tongcheng/internal/modules/merchant/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// SettleListOptions 结算列表过滤条件
type SettleListOptions struct {
	ShopID uint
	Period string
	Status *int
}

// SettleRepository 结算仓储接口
type SettleRepository interface {
	Create(s *model.Settle) error
	FindByID(id uint) (*model.Settle, error)
	Update(s *model.Settle) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(pagination *utils.Pagination, opts SettleListOptions) ([]model.Settle, int64, error)
	FindByShopID(shopID uint, pagination *utils.Pagination) ([]model.Settle, int64, error)
	FindByPeriod(shopID uint, period string) (*model.Settle, error)

	// 汇总
	SummaryByShop(shopID uint) (totalAmount, platformFee, shopAmount float64, count int64, err error)
	SummaryByPeriod(period string) (totalAmount, platformFee, shopAmount float64, count int64, err error)
}

type settleRepository struct {
	db *gorm.DB
}

// NewSettleRepository 创建结算仓储实例
func NewSettleRepository(db *gorm.DB) SettleRepository {
	return &settleRepository{db: db}
}

func (r *settleRepository) Create(s *model.Settle) error {
	return r.db.Create(s).Error
}

func (r *settleRepository) FindByID(id uint) (*model.Settle, error) {
	var s model.Settle
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settleRepository) Update(s *model.Settle) error {
	return r.db.Save(s).Error
}

func (r *settleRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Settle{}).Where("id = ?", id).Updates(fields).Error
}

func (r *settleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Settle{}, id).Error
}

func (r *settleRepository) List(pagination *utils.Pagination, opts SettleListOptions) ([]model.Settle, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Settle
	var total int64

	q := r.db.Model(&model.Settle{})
	if opts.ShopID > 0 {
		q = q.Where("shop_id = ?", opts.ShopID)
	}
	if opts.Period != "" {
		q = q.Where("period = ?", opts.Period)
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

func (r *settleRepository) FindByShopID(shopID uint, pagination *utils.Pagination) ([]model.Settle, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Settle
	var total int64

	q := r.db.Model(&model.Settle{}).Where("shop_id = ?", shopID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("period DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *settleRepository) FindByPeriod(shopID uint, period string) (*model.Settle, error) {
	var s model.Settle
	if err := r.db.Where("shop_id = ? AND period = ?", shopID, period).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *settleRepository) SummaryByShop(shopID uint) (totalAmount, platformFee, shopAmount float64, count int64, err error) {
	var result struct {
		TotalAmount float64
		PlatformFee float64
		ShopAmount  float64
		Count       int64
	}
	err = r.db.Model(&model.Settle{}).
		Where("shop_id = ? AND status = ?", shopID, model.SettleStatusSettled).
		Select("COALESCE(SUM(total_amount), 0) AS total_amount, COALESCE(SUM(platform_fee), 0) AS platform_fee, COALESCE(SUM(shop_amount), 0) AS shop_amount, COUNT(*) AS count").
		Scan(&result).Error
	if err != nil {
		return
	}
	return result.TotalAmount, result.PlatformFee, result.ShopAmount, result.Count, nil
}

func (r *settleRepository) SummaryByPeriod(period string) (totalAmount, platformFee, shopAmount float64, count int64, err error) {
	var result struct {
		TotalAmount float64
		PlatformFee float64
		ShopAmount  float64
		Count       int64
	}
	err = r.db.Model(&model.Settle{}).
		Where("period = ? AND status = ?", period, model.SettleStatusSettled).
		Select("COALESCE(SUM(total_amount), 0) AS total_amount, COALESCE(SUM(platform_fee), 0) AS platform_fee, COALESCE(SUM(shop_amount), 0) AS shop_amount, COUNT(*) AS count").
		Scan(&result).Error
	if err != nil {
		return
	}
	return result.TotalAmount, result.PlatformFee, result.ShopAmount, result.Count, nil
}
