// Package repository 同城拼车出行数据访问层 - 退款记录
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// RefundListOptions 退款记录列表过滤条件
type RefundListOptions struct {
	PaymentID    uint
	BookingID    uint
	PincheID     uint
	RefundNo     string
	RefundStatus *int
	RefundMethod string
	OperatorID   uint
}

// RefundRepository 退款记录仓储接口
type RefundRepository interface {
	Create(rf *model.PincheRefund) error
	FindByID(id uint) (*model.PincheRefund, error)
	FindByRefundNo(refundNo string) (*model.PincheRefund, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts RefundListOptions) ([]model.PincheRefund, int64, error)
	ListByPayment(paymentID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error)
	ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error)

	UpdateStatus(id uint, status int) error
	CountByStatus(regionID uint, status int) (int64, error)
	SumRefundByPinche(pincheID uint) (float64, error)
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository 创建退款记录仓储实例
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(rf *model.PincheRefund) error {
	return r.db.Create(rf).Error
}

func (r *refundRepository) FindByID(id uint) (*model.PincheRefund, error) {
	var rf model.PincheRefund
	if err := r.db.First(&rf, id).Error; err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *refundRepository) FindByRefundNo(refundNo string) (*model.PincheRefund, error) {
	var rf model.PincheRefund
	if err := r.db.Where("refund_no = ?", refundNo).First(&rf).Error; err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *refundRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheRefund{}).Where("id = ?", id).Updates(fields).Error
}

func (r *refundRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheRefund{}, id).Error
}

func (r *refundRepository) List(regionID uint, pagination *utils.Pagination, opts RefundListOptions) ([]model.PincheRefund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRefund
	var total int64

	query := r.db.Model(&model.PincheRefund{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PaymentID > 0 {
		query = query.Where("payment_id = ?", opts.PaymentID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.RefundNo != "" {
		query = query.Where("refund_no = ?", opts.RefundNo)
	}
	if opts.RefundStatus != nil {
		query = query.Where("refund_status = ?", *opts.RefundStatus)
	}
	if opts.RefundMethod != "" {
		query = query.Where("refund_method = ?", opts.RefundMethod)
	}
	if opts.OperatorID > 0 {
		query = query.Where("operator_id = ?", opts.OperatorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByPayment(paymentID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRefund
	var total int64

	query := r.db.Model(&model.PincheRefund{}).Where("payment_id = ?", paymentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRefund
	var total int64

	query := r.db.Model(&model.PincheRefund{}).Where("booking_id = ?", bookingID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheRefund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheRefund
	var total int64

	query := r.db.Model(&model.PincheRefund{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *refundRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PincheRefund{}).Where("id = ?", id).
		Update("refund_status", status).Error
}

func (r *refundRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheRefund{}).Where("refund_status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *refundRepository) SumRefundByPinche(pincheID uint) (float64, error) {
	var sum float64
	if err := r.db.Model(&model.PincheRefund{}).
		Where("pinche_id = ? AND refund_status = ?", pincheID, model.RefundStatusDone).
		Select("COALESCE(SUM(refund_amount), 0)").
		Scan(&sum).Error; err != nil {
		return 0, err
	}
	return sum, nil
}
