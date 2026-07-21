// Package repository 同城商城数据访问层 - 订单明细
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// OrderItemRepository 订单明细仓储接口
type OrderItemRepository interface {
	Create(item *model.OrderItem) error
	BatchCreate(items []model.OrderItem) error
	FindByID(id uint) (*model.OrderItem, error)
	Update(item *model.OrderItem) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	ListByOrder(orderID uint) ([]model.OrderItem, error)
	ListByOrders(orderIDs []uint) ([]model.OrderItem, error)
	List(opts OrderItemListOptions, pagination *utils.Pagination) ([]model.OrderItem, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.OrderItem, int64, error)

	// 状态更新
	UpdateReviewStatus(id uint, hasReview bool, reviewID uint) error
	UpdateRefundStatus(id uint, refundStatus int, refundID uint) error
}

// OrderItemListOptions 订单明细列表过滤条件
type OrderItemListOptions struct {
	OrderID    uint
	OrderNo    string
	ProductID  uint
	SkuID      uint
	ShopID     uint
	UserID     uint
	Status     *int
	Keyword    string
	StartDate  string
	EndDate    string
	RegionID   uint
}

type orderItemRepository struct {
	db *gorm.DB
}

// NewOrderItemRepository 创建订单明细仓储实例
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) Create(item *model.OrderItem) error {
	return r.db.Create(item).Error
}

func (r *orderItemRepository) BatchCreate(items []model.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *orderItemRepository) FindByID(id uint) (*model.OrderItem, error) {
	var item model.OrderItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderItemRepository) Update(item *model.OrderItem) error {
	return r.db.Save(item).Error
}

func (r *orderItemRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.OrderItem{}).Where("id = ?", id).Updates(fields).Error
}

func (r *orderItemRepository) Delete(id uint) error {
	return r.db.Delete(&model.OrderItem{}, id).Error
}

func (r *orderItemRepository) ListByOrder(orderID uint) ([]model.OrderItem, error) {
	var list []model.OrderItem
	if err := r.db.Where("order_id = ?", orderID).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderItemRepository) ListByOrders(orderIDs []uint) ([]model.OrderItem, error) {
	if len(orderIDs) == 0 {
		return []model.OrderItem{}, nil
	}
	var list []model.OrderItem
	if err := r.db.Where("order_id IN ?", orderIDs).Order("order_id ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderItemRepository) List(opts OrderItemListOptions, pagination *utils.Pagination) ([]model.OrderItem, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.OrderItem
	var total int64

	query := r.db.Model(&model.OrderItem{})
	if opts.OrderID > 0 {
		query = query.Where("order_id = ?", opts.OrderID)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.ProductID > 0 {
		query = query.Where("product_id = ?", opts.ProductID)
	}
	if opts.SkuID > 0 {
		query = query.Where("sku_id = ?", opts.SkuID)
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
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("product_name ILIKE ? OR sku_name ILIKE ? OR sku_code ILIKE ?", like, like, like)
	}
	if opts.StartDate != "" {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("created_at <= ?", opts.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *orderItemRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.OrderItem, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.OrderItem
	var total int64

	query := r.db.Model(&model.OrderItem{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *orderItemRepository) UpdateReviewStatus(id uint, hasReview bool, reviewID uint) error {
	return r.db.Model(&model.OrderItem{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"has_review": hasReview,
			"review_id":  reviewID,
		}).Error
}

func (r *orderItemRepository) UpdateRefundStatus(id uint, refundStatus int, refundID uint) error {
	return r.db.Model(&model.OrderItem{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"refund_status": refundStatus,
			"refund_id":     refundID,
		}).Error
}
