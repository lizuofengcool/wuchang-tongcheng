// Package repository 同城商城数据访问层 - 退款
package repository

import (
	"wuchang-tongcheng/internal/modules/mall/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RefundRepository 退款仓储接口
type RefundRepository interface {
	Create(refund *model.Refund) error
	FindByID(id uint) (*model.Refund, error)
	FindByRefundNo(refundNo string) (*model.Refund, error)
	Update(refund *model.Refund) error
	UpdateFields(id uint, fields map[string]interface{}) error

	List(opts RefundListOptions, pagination *utils.Pagination) ([]model.Refund, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.Refund, int64, error)
	ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Refund, int64, error)
	ListByOrder(orderID uint) ([]model.Refund, error)

	UpdateStatus(id uint, status int, fields map[string]interface{}) error
	Stats(opts RefundStatsOptions) (*RefundStatsResult, error)
}

// RefundListOptions 退款列表过滤条件
type RefundListOptions struct {
	OrderID    uint
	OrderNo    string
	UserID     uint
	ShopID     uint
	RefundNo   string
	Status     *int
	RefundType string
	Keyword    string
	StartDate  string
	EndDate    string
	RegionID   uint
}

// RefundStatsOptions 退款统计选项
type RefundStatsOptions struct {
	RegionID  uint
	ShopID    uint
	StartDate string
	EndDate   string
}

// RefundStatsResult 退款统计结果
type RefundStatsResult struct {
	TotalCount     int64   `gorm:"column:total_count" json:"total_count"`
	TotalAmount    float64 `gorm:"column:total_amount" json:"total_amount"`
	PendingCount   int64   `gorm:"column:pending_count" json:"pending_count"`
	ApprovedCount  int64   `gorm:"column:approved_count" json:"approved_count"`
	RejectedCount  int64   `gorm:"column:rejected_count" json:"rejected_count"`
	RefundedCount  int64   `gorm:"column:refunded_count" json:"refunded_count"`
	RefundedAmount float64 `gorm:"column:refunded_amount" json:"refunded_amount"`
	ReturningCount int64   `gorm:"column:returning_count" json:"returning_count"`
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository 创建退款仓储实例
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(refund *model.Refund) error {
	return r.db.Create(refund).Error
}

func (r *refundRepository) FindByID(id uint) (*model.Refund, error) {
	var refund model.Refund
	if err := r.db.First(&refund, id).Error; err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) FindByRefundNo(refundNo string) (*model.Refund, error) {
	var refund model.Refund
	if err := r.db.Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) Update(refund *model.Refund) error {
	return r.db.Save(refund).Error
}

func (r *refundRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Refund{}).Where("id = ?", id).Updates(fields).Error
}

func (r *refundRepository) List(opts RefundListOptions, pagination *utils.Pagination) ([]model.Refund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Refund
	var total int64

	query := r.db.Model(&model.Refund{})
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
	if opts.RefundNo != "" {
		query = query.Where("refund_no = ?", opts.RefundNo)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.RefundType != "" {
		query = query.Where("refund_type = ?", opts.RefundType)
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
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("refund_no ILIKE ? OR order_no ILIKE ? OR reason ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.Refund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Refund
	var total int64

	query := r.db.Model(&model.Refund{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByShop(shopID uint, pagination *utils.Pagination) ([]model.Refund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Refund
	var total int64

	query := r.db.Model(&model.Refund{}).Where("shop_id = ?", shopID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByOrder(orderID uint) ([]model.Refund, error) {
	var list []model.Refund
	if err := r.db.Where("order_id = ?", orderID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *refundRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.Refund{}).Where("id = ?", id).Updates(fields).Error
}

func (r *refundRepository) Stats(opts RefundStatsOptions) (*RefundStatsResult, error) {
	var result RefundStatsResult
	query := r.db.Model(&model.Refund{}).Select(`
		COUNT(*) AS total_count,
		COALESCE(SUM(amount), 0) AS total_amount,
		COUNT(*) FILTER (WHERE status = 0) AS pending_count,
		COUNT(*) FILTER (WHERE status = 1) AS approved_count,
		COUNT(*) FILTER (WHERE status = 2) AS rejected_count,
		COUNT(*) FILTER (WHERE status = 3) AS refunded_count,
		COALESCE(SUM(CASE WHEN status = 3 THEN refund_amount ELSE 0 END), 0) AS refunded_amount,
		COUNT(*) FILTER (WHERE status = 1 AND refund_type = 'refund_return') AS returning_count
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
