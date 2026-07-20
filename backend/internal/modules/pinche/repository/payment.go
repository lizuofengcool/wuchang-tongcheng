// Package repository 同城拼车出行数据访问层 - 支付（含 ETC）
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PaymentListOptions 支付列表过滤条件
type PaymentListOptions struct {
	PincheID      uint
	BookingID     uint
	PayerID       uint
	PayeeID       uint
	PaymentMethod string
	Status        *int
	PaymentNo     string
}

// PaymentRepository 支付仓储接口
type PaymentRepository interface {
	Create(p *model.PinchePayment) error
	FindByID(id uint) (*model.PinchePayment, error)
	FindByPaymentNo(paymentNo string) (*model.PinchePayment, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts PaymentListOptions) ([]model.PinchePayment, int64, error)
	ListByPayer(payerID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error)
	ListByPayee(payeeID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error)
	ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error)

	UpdateStatus(id uint, status int) error
	UpdateETC(id uint, laneID string, entryTime, exitTime interface{}) error
	CountByStatus(regionID uint, status int) (int64, error)
	SumAmountByPayee(payeeID uint) (float64, error)
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建支付仓储实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(p *model.PinchePayment) error {
	return r.db.Create(p).Error
}

func (r *paymentRepository) FindByID(id uint) (*model.PinchePayment, error) {
	var p model.PinchePayment
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) FindByPaymentNo(paymentNo string) (*model.PinchePayment, error) {
	var p model.PinchePayment
	if err := r.db.Where("payment_no = ?", paymentNo).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PinchePayment{}).Where("id = ?", id).Updates(fields).Error
}

func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&model.PinchePayment{}, id).Error
}

func (r *paymentRepository) List(regionID uint, pagination *utils.Pagination, opts PaymentListOptions) ([]model.PinchePayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PinchePayment
	var total int64

	query := r.db.Model(&model.PinchePayment{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.PayerID > 0 {
		query = query.Where("payer_id = ?", opts.PayerID)
	}
	if opts.PayeeID > 0 {
		query = query.Where("payee_id = ?", opts.PayeeID)
	}
	if opts.PaymentMethod != "" {
		query = query.Where("payment_method = ?", opts.PaymentMethod)
	}
	if opts.Status != nil {
		query = query.Where("payment_status = ?", *opts.Status)
	}
	if opts.PaymentNo != "" {
		query = query.Where("payment_no ILIKE ?", "%"+opts.PaymentNo+"%")
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

func (r *paymentRepository) ListByPayer(payerID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PinchePayment
	var total int64

	query := r.db.Model(&model.PinchePayment{}).Where("payer_id = ?", payerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByPayee(payeeID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PinchePayment
	var total int64

	query := r.db.Model(&model.PinchePayment{}).Where("payee_id = ?", payeeID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByBooking(bookingID uint, pagination *utils.Pagination) ([]model.PinchePayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PinchePayment
	var total int64

	query := r.db.Model(&model.PinchePayment{}).Where("booking_id = ?", bookingID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.PinchePayment{}).Where("id = ?", id).Update("payment_status", status).Error
}

func (r *paymentRepository) UpdateETC(id uint, laneID string, entryTime, exitTime interface{}) error {
	fields := map[string]interface{}{"etc_lane_id": laneID}
	if entryTime != nil {
		fields["etc_entry_time"] = entryTime
	}
	if exitTime != nil {
		fields["etc_exit_time"] = exitTime
	}
	return r.db.Model(&model.PinchePayment{}).Where("id = ?", id).Updates(fields).Error
}

func (r *paymentRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PinchePayment{}).Where("payment_status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *paymentRepository) SumAmountByPayee(payeeID uint) (float64, error) {
	var total float64
	if err := r.db.Model(&model.PinchePayment{}).
		Where("payee_id = ? AND payment_status = ?", payeeID, model.PaymentStatusPaid).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
