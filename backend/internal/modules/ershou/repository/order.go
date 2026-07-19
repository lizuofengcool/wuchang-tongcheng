// Package repository 订单数据访问层
// 依据 v3.2.1 架构方案：对标闲鱼/转转 11 状态机订单
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(order *model.ErshouOrder, items []model.ErshouOrderItem) error
	FindByID(id uint) (*model.ErshouOrder, error)
	FindByOrderNo(orderNo string) (*model.ErshouOrder, error)
	Update(id uint, fields map[string]interface{}) error
	UpdateByOrderNo(orderNo string, fields map[string]interface{}) error
	List(query OrderListQuery, pagination *utils.Pagination) ([]model.ErshouOrder, int64, error)
	ListItems(orderID uint) ([]model.ErshouOrderItem, error)
	ListByErshouID(ershouID uint) ([]model.ErshouOrder, error)
	CountByStatus(userID uint, status int) (int64, error)
	BatchUpdateStatus(ids []uint, status int) error
}

// OrderListQuery 订单列表查询参数
type OrderListQuery struct {
	UserID   uint
	Role     string // buyer/seller/all
	Status   *int
	OrderNo  string
	Keyword  string
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.ErshouOrder, items []model.ErshouOrderItem) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range items {
		items[i].OrderID = order.ID
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *orderRepository) FindByID(id uint) (*model.ErshouOrder, error) {
	var o model.ErshouOrder
	if err := r.db.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) FindByOrderNo(orderNo string) (*model.ErshouOrder, error) {
	var o model.ErshouOrder
	if err := r.db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouOrder{}).Where("id = ?", id).Updates(fields).Error
}

func (r *orderRepository) UpdateByOrderNo(orderNo string, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouOrder{}).Where("order_no = ?", orderNo).Updates(fields).Error
}

func (r *orderRepository) List(query OrderListQuery, pagination *utils.Pagination) ([]model.ErshouOrder, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouOrder
	var total int64

	q := r.db.Model(&model.ErshouOrder{})
	switch query.Role {
	case "buyer":
		q = q.Where("buyer_id = ?", query.UserID)
	case "seller":
		q = q.Where("seller_id = ?", query.UserID)
	case "all":
		q = q.Where("buyer_id = ? OR seller_id = ?", query.UserID, query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.OrderNo != "" {
		q = q.Where("order_no = ?", query.OrderNo)
	}
	if query.Keyword != "" {
		q = q.Where("order_no ILIKE ?", "%"+query.Keyword+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *orderRepository) ListItems(orderID uint) ([]model.ErshouOrderItem, error) {
	var items []model.ErshouOrderItem
	if err := r.db.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderRepository) ListByErshouID(ershouID uint) ([]model.ErshouOrder, error) {
	var list []model.ErshouOrder
	if err := r.db.Joins("JOIN ers_order_items ON ers_order_items.order_id = ers_orders.id").
		Where("ers_order_items.ershou_id = ?", ershouID).
		Group("ers_orders.id").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderRepository) CountByStatus(userID uint, status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.ErshouOrder{}).
		Where("buyer_id = ? AND status = ?", userID, status).
		Count(&count).Error
	return count, err
}

func (r *orderRepository) BatchUpdateStatus(ids []uint, status int) error {
	return r.db.Model(&model.ErshouOrder{}).Where("id IN ?", ids).Update("status", status).Error
}
