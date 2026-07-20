// Package repository 同城拼车出行数据访问层 - 取消记录
package repository

import (
	"wuchang-tongcheng/internal/modules/pinche/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CancelListOptions 取消记录列表过滤条件
type CancelListOptions struct {
	PincheID      uint
	BookingID     uint
	CancelledBy   uint
	CancelledRole string
	CancelType    string
	RefundStatus  *int
}

// CancelRepository 取消记录仓储接口
type CancelRepository interface {
	Create(c *model.PincheCancel) error
	FindByID(id uint) (*model.PincheCancel, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts CancelListOptions) ([]model.PincheCancel, int64, error)
	ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheCancel, int64, error)
	ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheCancel, int64, error)

	UpdateRefundStatus(id uint, refundStatus int) error
	CountByUser(userID uint, days int) (int64, error)
}

type cancelRepository struct {
	db *gorm.DB
}

// NewCancelRepository 创建取消记录仓储实例
func NewCancelRepository(db *gorm.DB) CancelRepository {
	return &cancelRepository{db: db}
}

func (r *cancelRepository) Create(c *model.PincheCancel) error {
	return r.db.Create(c).Error
}

func (r *cancelRepository) FindByID(id uint) (*model.PincheCancel, error) {
	var c model.PincheCancel
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *cancelRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PincheCancel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *cancelRepository) Delete(id uint) error {
	return r.db.Delete(&model.PincheCancel{}, id).Error
}

func (r *cancelRepository) List(regionID uint, pagination *utils.Pagination, opts CancelListOptions) ([]model.PincheCancel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheCancel
	var total int64

	query := r.db.Model(&model.PincheCancel{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.PincheID > 0 {
		query = query.Where("pinche_id = ?", opts.PincheID)
	}
	if opts.BookingID > 0 {
		query = query.Where("booking_id = ?", opts.BookingID)
	}
	if opts.CancelledBy > 0 {
		query = query.Where("cancelled_by = ?", opts.CancelledBy)
	}
	if opts.CancelledRole != "" {
		query = query.Where("cancelled_role = ?", opts.CancelledRole)
	}
	if opts.CancelType != "" {
		query = query.Where("cancel_type = ?", opts.CancelType)
	}
	if opts.RefundStatus != nil {
		query = query.Where("refund_status = ?", *opts.RefundStatus)
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

func (r *cancelRepository) ListByPinche(pincheID uint, pagination *utils.Pagination) ([]model.PincheCancel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheCancel
	var total int64

	query := r.db.Model(&model.PincheCancel{}).Where("pinche_id = ?", pincheID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *cancelRepository) ListByUser(userID uint, pagination *utils.Pagination) ([]model.PincheCancel, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.PincheCancel
	var total int64

	query := r.db.Model(&model.PincheCancel{}).Where("cancelled_by = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *cancelRepository) UpdateRefundStatus(id uint, refundStatus int) error {
	return r.db.Model(&model.PincheCancel{}).Where("id = ?", id).
		Update("refund_status", refundStatus).Error
}

func (r *cancelRepository) CountByUser(userID uint, days int) (int64, error) {
	var count int64
	q := r.db.Model(&model.PincheCancel{}).Where("cancelled_by = ?", userID)
	if days > 0 {
		q = q.Where("created_at >= NOW() - INTERVAL '? days'", days)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
