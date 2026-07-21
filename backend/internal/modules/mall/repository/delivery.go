// Package repository 同城商城数据访问层 - 配送单
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// DeliveryRepository 配送单仓储接口
type DeliveryRepository interface {
	Create(d *model.Delivery) error
	FindByID(id uint) (*model.Delivery, error)
	FindByDeliveryNo(no string) (*model.Delivery, error)
	FindByOrderID(orderID uint) (*model.Delivery, error)
	Update(d *model.Delivery) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts DeliveryListOptions) ([]model.Delivery, int64, error)
	AdminList(pagination *utils.Pagination, opts DeliveryAdminListOptions) ([]model.Delivery, int64, error)
	ListByRider(riderID uint, pagination *utils.Pagination, opts DeliveryListOptions) ([]model.Delivery, int64, error)

	// 状态更新
	UpdateStatus(id uint, status int, fields map[string]interface{}) error

	// 统计
	CountByStatus(riderID uint, status int) (int64, error)
	SumEarnings(riderID uint, start, end time.Time) (float64, int64, error)
	SumFeeAndTip(riderID uint, start, end time.Time) (totalFee, totalTip float64, count int64, err error)
}

// DeliveryListOptions C 端配送单列表过滤条件
type DeliveryListOptions struct {
	Keyword    string
	Status     *int
	ShopID     uint
	UserID     uint
	DeliveryNo string
}

// DeliveryAdminListOptions M 端管理列表过滤条件
type DeliveryAdminListOptions struct {
	RegionID   uint
	RiderID    uint
	ShopID     uint
	UserID     uint
	Status     *int
	Keyword    string
	DeliveryNo string
}

type deliveryRepository struct {
	db *gorm.DB
}

// NewDeliveryRepository 创建配送单仓储实例
func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(d *model.Delivery) error {
	return r.db.Create(d).Error
}

func (r *deliveryRepository) FindByID(id uint) (*model.Delivery, error) {
	var d model.Delivery
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deliveryRepository) FindByDeliveryNo(no string) (*model.Delivery, error) {
	var d model.Delivery
	if err := r.db.Where("delivery_no = ?", no).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deliveryRepository) FindByOrderID(orderID uint) (*model.Delivery, error) {
	var d model.Delivery
	if err := r.db.Where("order_id = ?", orderID).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deliveryRepository) Update(d *model.Delivery) error {
	return r.db.Save(d).Error
}

func (r *deliveryRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Delivery{}).Where("id = ?", id).Updates(fields).Error
}

func (r *deliveryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Delivery{}, id).Error
}

func (r *deliveryRepository) List(regionID uint, pagination *utils.Pagination, opts DeliveryListOptions) ([]model.Delivery, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Delivery
	var total int64

	query := r.db.Model(&model.Delivery{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.DeliveryNo != "" {
		query = query.Where("delivery_no = ?", opts.DeliveryNo)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("delivery_no ILIKE ? OR pickup_address ILIKE ? OR delivery_address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *deliveryRepository) AdminList(pagination *utils.Pagination, opts DeliveryAdminListOptions) ([]model.Delivery, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Delivery
	var total int64

	query := r.db.Model(&model.Delivery{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.RiderID > 0 {
		query = query.Where("rider_id = ?", opts.RiderID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.DeliveryNo != "" {
		query = query.Where("delivery_no = ?", opts.DeliveryNo)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("delivery_no ILIKE ? OR pickup_address ILIKE ? OR delivery_address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *deliveryRepository) ListByRider(riderID uint, pagination *utils.Pagination, opts DeliveryListOptions) ([]model.Delivery, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Delivery
	var total int64

	query := r.db.Model(&model.Delivery{}).Where("rider_id = ?", riderID)
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.DeliveryNo != "" {
		query = query.Where("delivery_no = ?", opts.DeliveryNo)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("delivery_no ILIKE ? OR pickup_address ILIKE ? OR delivery_address ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *deliveryRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Delivery{}).Where("id = ?", id).Updates(fields).Error
}

func (r *deliveryRepository) CountByStatus(riderID uint, status int) (int64, error) {
	var count int64
	query := r.db.Model(&model.Delivery{})
	if riderID > 0 {
		query = query.Where("rider_id = ?", riderID)
	}
	if err := query.Where("status = ?", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// SumEarnings 统计骑手在 [start, end) 时间段内的送达订单数和总收益（配送费+小费）
func (r *deliveryRepository) SumEarnings(riderID uint, start, end time.Time) (float64, int64, error) {
	var result struct {
		Total float64 `gorm:"column:total"`
		Count int64   `gorm:"column:cnt"`
	}
	err := r.db.Model(&model.Delivery{}).
		Where("rider_id = ? AND status = ? AND delivered_at >= ? AND delivered_at < ?",
			riderID, model.DeliveryStatusDelivered, start, end).
		Select("COALESCE(SUM(delivery_fee + tip), 0) AS total, COUNT(*) AS cnt").
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Total, result.Count, nil
}

// SumFeeAndTip 分别统计骑手在 [start, end) 时间段内的配送费总额、小费总额和订单数
func (r *deliveryRepository) SumFeeAndTip(riderID uint, start, end time.Time) (totalFee, totalTip float64, count int64, err error) {
	var result struct {
		TotalFee float64 `gorm:"column:total_fee"`
		TotalTip float64 `gorm:"column:total_tip"`
		Count    int64   `gorm:"column:cnt"`
	}
	err = r.db.Model(&model.Delivery{}).
		Where("rider_id = ? AND status = ? AND delivered_at >= ? AND delivered_at < ?",
			riderID, model.DeliveryStatusDelivered, start, end).
		Select("COALESCE(SUM(delivery_fee), 0) AS total_fee, COALESCE(SUM(tip), 0) AS total_tip, COUNT(*) AS cnt").
		Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.TotalFee, result.TotalTip, result.Count, nil
}
