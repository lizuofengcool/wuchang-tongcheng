// Package repository 同城零工兼职数据访问层 - 薪资支付
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// PaymentRepository 薪资支付仓储接口
type PaymentRepository interface {
	Create(p *model.LinggongPayment) error
	FindByID(id uint) (*model.LinggongPayment, error)
	FindByPaymentNo(no string) (*model.LinggongPayment, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts PaymentListOptions) ([]model.LinggongPayment, int64, error)
	AdminList(pagination *utils.Pagination, opts PaymentAdminListOptions) ([]model.LinggongPayment, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error)
	ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error)
	ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error)
	CountByLinggong(linggongID uint, status int) (int64, error)
}

// PaymentListOptions C 端支付列表过滤条件
type PaymentListOptions struct {
	LinggongID        uint
	TaskID            uint
	ApplicationID     uint
	EmployerID        uint
	WorkerID          uint
	PaymentType       string
	Settlement        string
	Status            *int
	SettlementStatus  *int
	Keyword           string
}

// PaymentAdminListOptions M 端支付列表过滤条件
type PaymentAdminListOptions struct {
	RegionID      uint
	LinggongID    uint
	EmployerID    uint
	WorkerID      uint
	PaymentType   string
	Status        *int
	Keyword       string
}

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 创建薪资支付仓储实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(p *model.LinggongPayment) error {
	return r.db.Create(p).Error
}

func (r *paymentRepository) FindByID(id uint) (*model.LinggongPayment, error) {
	var p model.LinggongPayment
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) FindByPaymentNo(no string) (*model.LinggongPayment, error) {
	var p model.LinggongPayment
	if err := r.db.Where("payment_no = ?", no).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongPayment{}).Where("id = ?", id).Updates(fields).Error
}

func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongPayment{}, id).Error
}

func (r *paymentRepository) List(regionID uint, pagination *utils.Pagination, opts PaymentListOptions) ([]model.LinggongPayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongPayment
	var total int64

	query := r.db.Model(&model.LinggongPayment{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.TaskID > 0 {
		query = query.Where("task_id = ?", opts.TaskID)
	}
	if opts.ApplicationID > 0 {
		query = query.Where("application_id = ?", opts.ApplicationID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.PaymentType != "" {
		query = query.Where("payment_type = ?", opts.PaymentType)
	}
	if opts.Settlement != "" {
		query = query.Where("settlement = ?", opts.Settlement)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.SettlementStatus != nil {
		query = query.Where("settlement_status = ?", *opts.SettlementStatus)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("payment_no ILIKE ? OR employer_name ILIKE ? OR worker_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) AdminList(pagination *utils.Pagination, opts PaymentAdminListOptions) ([]model.LinggongPayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongPayment
	var total int64

	query := r.db.Model(&model.LinggongPayment{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.PaymentType != "" {
		query = query.Where("payment_type = ?", opts.PaymentType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("payment_no ILIKE ? OR employer_name ILIKE ? OR worker_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongPayment
	var total int64
	query := r.db.Model(&model.LinggongPayment{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongPayment
	var total int64
	query := r.db.Model(&model.LinggongPayment{}).Where("employer_id = ?", employerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongPayment, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongPayment
	var total int64
	query := r.db.Model(&model.LinggongPayment{}).Where("worker_id = ?", workerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *paymentRepository) CountByLinggong(linggongID uint, status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.LinggongPayment{}).
		Where("linggong_id = ? AND status = ?", linggongID, status).
		Count(&count).Error
	return count, err
}
