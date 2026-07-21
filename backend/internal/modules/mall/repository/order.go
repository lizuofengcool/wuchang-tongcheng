// Package repository 同城商城数据访问层 - 订单
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(o *model.Order, items []model.OrderItem) error
	FindByID(id uint) (*model.Order, error)
	FindByOrderNo(orderNo string) (*model.Order, error)
	Update(o *model.Order) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	ListByUser(userID uint, pagination *utils.Pagination, opts OrderListOptions) ([]model.Order, int64, error)
	AdminList(pagination *utils.Pagination, opts AdminOrderListOptions) ([]model.Order, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination, opts OrderListOptions) ([]model.Order, int64, error)

	// 状态更新
	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	BatchUpdateStatus(ids []uint, status int) error

	// 定时任务
	ListAutoClose(pagination *utils.Pagination) ([]model.Order, error)
	ListAutoConfirm(pagination *utils.Pagination) ([]model.Order, error)
	ListAutoReview(pagination *utils.Pagination) ([]model.Order, error)

	// 统计
	CountByStatus(userID uint, status int) (int64, error)
	Summary(opts OrderSummaryOptions) (*OrderSummaryResult, error)
}

// OrderListOptions 订单列表过滤条件
type OrderListOptions struct {
	Status         *int
	ShopID         uint
	OrderNo        string
	Keyword        string
	StartDate      string
	EndDate        string
	Buyer          string
}

// AdminOrderListOptions 管理后台订单列表过滤条件
type AdminOrderListOptions struct {
	RegionID       uint
	UserID         uint
	ShopID         uint
	Status         *int
	OrderNo        string
	PaymentMethod  string
	Keyword        string
	StartDate      string
	EndDate        string
}

// OrderSummaryOptions 订单统计选项
type OrderSummaryOptions struct {
	RegionID  uint
	UserID    uint
	ShopID    uint
	StartDate time.Time
	EndDate   time.Time
}

// OrderSummaryResult 订单统计结果
type OrderSummaryResult struct {
	TotalCount     int64   `gorm:"column:total_count" json:"total_count"`
	TotalAmount    float64 `gorm:"column:total_amount" json:"total_amount"`
	PaidCount      int64   `gorm:"column:paid_count" json:"paid_count"`
	PaidAmount     float64 `gorm:"column:paid_amount" json:"paid_amount"`
	PendingCount   int64   `gorm:"column:pending_count" json:"pending_count"`
	ShippedCount   int64   `gorm:"column:shipped_count" json:"shipped_count"`
	CompletedCount int64   `gorm:"column:completed_count" json:"completed_count"`
	CancelledCount int64   `gorm:"column:cancelled_count" json:"cancelled_count"`
	RefundedCount  int64   `gorm:"column:refunded_count" json:"refunded_count"`
	RefundedAmount float64 `gorm:"column:refunded_amount" json:"refunded_amount"`
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(o *model.Order, items []model.OrderItem) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Create(o).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range items {
		items[i].OrderID = o.ID
		items[i].RegionID = o.RegionID
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *orderRepository) FindByID(id uint) (*model.Order, error) {
	var o model.Order
	if err := r.db.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) FindByOrderNo(orderNo string) (*model.Order, error) {
	var o model.Order
	if err := r.db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) Update(o *model.Order) error {
	return r.db.Save(o).Error
}

func (r *orderRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Updates(fields).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}

func (r *orderRepository) ListByUser(userID uint, pagination *utils.Pagination, opts OrderListOptions) ([]model.Order, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("user_id = ?", userID)
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("order_no ILIKE ? OR buyer_name ILIKE ? OR receiver_name ILIKE ?", like, like, like)
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

func (r *orderRepository) AdminList(pagination *utils.Pagination, opts AdminOrderListOptions) ([]model.Order, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Order
	var total int64

	query := r.db.Model(&model.Order{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.PaymentMethod != "" {
		query = query.Where("payment_method = ?", opts.PaymentMethod)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("order_no ILIKE ? OR buyer_name ILIKE ? OR buyer_phone ILIKE ? OR receiver_name ILIKE ?", like, like, like, like)
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

func (r *orderRepository) ListByShop(shopID uint, pagination *utils.Pagination, opts OrderListOptions) ([]model.Order, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Order
	var total int64

	query := r.db.Model(&model.Order{}).Where("shop_id = ?", shopID)
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("order_no ILIKE ? OR buyer_name ILIKE ? OR receiver_name ILIKE ?", like, like, like)
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

func (r *orderRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Order{}).Where("id = ?", id).Updates(fields).Error
}

func (r *orderRepository) BatchUpdateStatus(ids []uint, status int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Order{}).Where("id IN ?", ids).UpdateColumn("status", status).Error
}

func (r *orderRepository) ListAutoClose(pagination *utils.Pagination) ([]model.Order, error) {
	now := time.Now()
	var list []model.Order
	if err := r.db.Where("status = ? AND auto_close_at <= ?", model.OrderStatusPending, now).
		Limit(100).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderRepository) ListAutoConfirm(pagination *utils.Pagination) ([]model.Order, error) {
	now := time.Now()
	var list []model.Order
	if err := r.db.Where("status = ? AND auto_confirm_at <= ?", model.OrderStatusShipped, now).
		Limit(100).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderRepository) ListAutoReview(pagination *utils.Pagination) ([]model.Order, error) {
	now := time.Now()
	var list []model.Order
	if err := r.db.Where("status = ? AND auto_review_at <= ?", model.OrderStatusReceived, now).
		Limit(100).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *orderRepository) CountByStatus(userID uint, status int) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Order{}).Where("user_id = ? AND status = ?", userID, status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *orderRepository) Summary(opts OrderSummaryOptions) (*OrderSummaryResult, error) {
	var result OrderSummaryResult
	query := r.db.Model(&model.Order{}).Select(`
		COUNT(*) AS total_count,
		COALESCE(SUM(pay_amount), 0) AS total_amount,
		COUNT(*) FILTER (WHERE status >= 1) AS paid_count,
		COALESCE(SUM(CASE WHEN status >= 1 THEN pay_amount ELSE 0 END), 0) AS paid_amount,
		COUNT(*) FILTER (WHERE status = 0) AS pending_count,
		COUNT(*) FILTER (WHERE status = 2) AS shipped_count,
		COUNT(*) FILTER (WHERE status = 4) AS completed_count,
		COUNT(*) FILTER (WHERE status = 5) AS cancelled_count,
		COUNT(*) FILTER (WHERE status = 6) AS refunded_count,
		COALESCE(SUM(CASE WHEN status = 6 THEN pay_amount ELSE 0 END), 0) AS refunded_amount
	`)
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if !opts.StartDate.IsZero() {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if !opts.EndDate.IsZero() {
		query = query.Where("created_at <= ?", opts.EndDate)
	}
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
