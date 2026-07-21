// Package repository 同城商城数据访问层 - 支付记录
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PaymentRepository 支付记录仓储接口
type PaymentRepository interface {
	Create(p *model.Payment) error
	FindByID(id uint) (*model.Payment, error)
	FindByPaymentNo(paymentNo string) (*model.Payment, error)
	FindByOrderID(orderID uint) (*model.Payment, error)
	Update(p *model.Payment) error
	UpdateFields(id uint, fields map[string]interface{}) error

	List(opts PaymentListOptions, pagination *utils.Pagination) ([]model.Payment, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Payment, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Payment, int64, error)

	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	Stats(opts PaymentStatsOptions) (*PaymentStatsResult, error)
}

// PaymentListOptions 支付列表过滤条件
type PaymentListOptions struct {
	OrderID   uint
	OrderNo   string
	UserID    uint
	ShopID    uint
	PaymentNo string
	TradeNo   string
	Method    string
	Status    *int
	StartDate string
	EndDate   string
	RegionID  uint
}

// PaymentStatsOptions 支付统计选项
type PaymentStatsOptions struct {
	RegionID  uint
	ShopID    uint
	StartDate string
	EndDate   string
}

// PaymentStatsResult 支付统计结果
type PaymentStatsResult struct {
	TotalCount    int64   `gorm:"column:total_count" json:"total_count"`
	TotalAmount   float64 `gorm:"column:total_amount" json:"total_amount"`
	SuccessCount  int64   `gorm:"column:success_count" json:"success_count"`
	SuccessAmount float64 `gorm:"column:success_amount" json:"success_amount"`
	PendingCount  int64   `gorm:"column:pending_count" json:"pending_count"`
	FailedCount   int64   `gorm:"column:failed_count" json:"failed_count"`
	RefundedCount int64   `gorm:"column:refunded_count" json:"refunded_count"`
	RefundAmount  float64 `gorm:"column:refund_amount" json:"refund_amount"`
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付记录仓储实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(p *model.Payment) error {
	return r.db.Create(p).Error
}

func (r *paymentRepository) FindByID(id uint) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) FindByPaymentNo(paymentNo string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("payment_no = ?", paymentNo).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) FindByOrderID(orderID uint) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) Update(p *model.Payment) error {
	return r.db.Save(p).Error
}

func (r *paymentRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", id).Updates(fields).Error
}

func (r *paymentRepository) List(opts PaymentListOptions, pagination *utils.Pagination) ([]model.Payment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Payment
	var total int64

	query := r.db.Model(&model.Payment{})
	if opts.OrderID > 0 {
		query = query.Where("order_id = ?", opts.OrderID)
	}
	if opts.OrderNo != "" {
		query = query.Where("order_no = ?", opts.OrderNo)
	}
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.PaymentNo != "" {
		query = query.Where("payment_no = ?", opts.PaymentNo)
	}
	if opts.TradeNo != "" {
		query = query.Where("trade_no = ?", opts.TradeNo)
	}
	if opts.Method != "" {
		query = query.Where("method = ?", opts.Method)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
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

func (r *paymentRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Payment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Payment
	var total int64

	query := r.db.Model(&model.Payment{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Payment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Payment
	var total int64

	query := r.db.Model(&model.Payment{}).Where("shop_id = ?", shopID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Payment{}).Where("id = ?", id).Updates(fields).Error
}

func (r *paymentRepository) Stats(opts PaymentStatsOptions) (*PaymentStatsResult, error) {
	var result PaymentStatsResult
	query := r.db.Model(&model.Payment{}).Select(`
		COUNT(*) AS total_count,
		COALESCE(SUM(amount), 0) AS total_amount,
		COUNT(*) FILTER (WHERE status = 1) AS success_count,
		COALESCE(SUM(CASE WHEN status = 1 THEN amount ELSE 0 END), 0) AS success_amount,
		COUNT(*) FILTER (WHERE status = 0) AS pending_count,
		COUNT(*) FILTER (WHERE status = 2) AS failed_count,
		COUNT(*) FILTER (WHERE status = 3) AS refunded_count,
		COALESCE(SUM(refund_amount), 0) AS refund_amount
	`)
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.ShopID > 0 {
		query = query.Where("shop_id = ?", opts.ShopID)
	}
	if opts.StartDate != "" {
		query = query.Where("created_at >= ?", opts.StartDate)
	}
	if opts.EndDate != "" {
		query = query.Where("created_at <= ?", opts.EndDate)
	}
	if err := query.Scan(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
