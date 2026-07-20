// Package repository 同城114数据访问层 - 优惠券
package repository

import (
	"wuchang-tongcheng/internal/modules/dh114/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	Create(c *model.Dh114Coupon) error
	FindByID(id uint) (*model.Dh114Coupon, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query CouponListQuery, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error)
	ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error)
	ListHot(regionID uint, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error)
	AdminList(query CouponAdminListQuery, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error)
	IncrIssuedCount(id uint) error
	IncrUsedCount(id uint) error
}

// CouponListQuery 优惠券列表查询
type CouponListQuery struct {
	Dh114ID    uint
	CouponType string
	Status     *int
	Featured   *bool
	Keyword    string
}

// CouponAdminListQuery 管理后台优惠券列表查询
type CouponAdminListQuery struct {
	Dh114ID     uint
	Status      *int
	AuditStatus *int
	Keyword     string
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓储实例
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(c *model.Dh114Coupon) error {
	return r.db.Create(c).Error
}

func (r *couponRepository) FindByID(id uint) (*model.Dh114Coupon, error) {
	var c model.Dh114Coupon
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *couponRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Dh114Coupon{}).Where("id = ?", id).Updates(fields).Error
}

func (r *couponRepository) Delete(id uint) error {
	return r.db.Delete(&model.Dh114Coupon{}, id).Error
}

func (r *couponRepository) List(regionID uint, query CouponListQuery, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Coupon
	var total int64

	q := r.db.Model(&model.Dh114Coupon{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.CouponType != "" {
		q = q.Where("coupon_type = ?", query.CouponType)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	} else {
		q = q.Where("status = ?", model.CouponStatusPublished)
	}
	if query.Featured != nil {
		q = q.Where("featured = ?", *query.Featured)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("featured DESC, published_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *couponRepository) ListByDh114(regionID uint, dh114ID uint, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error) {
	return r.List(regionID, CouponListQuery{
		Dh114ID: dh114ID,
		Status:  intPtrDh114(model.CouponStatusPublished),
	}, pagination)
}

func (r *couponRepository) ListHot(regionID uint, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Coupon
	var total int64

	q := r.db.Model(&model.Dh114Coupon{}).
		Where("status = ?", model.CouponStatusPublished).
		Where("audit_status = ?", model.AuditApproved)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("featured DESC, issued_count DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *couponRepository) AdminList(query CouponAdminListQuery, pagination *utils.Pagination) ([]model.Dh114Coupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Dh114Coupon
	var total int64

	q := r.db.Model(&model.Dh114Coupon{})
	if query.Dh114ID > 0 {
		q = q.Where("dh114_id = ?", query.Dh114ID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.AuditStatus != nil {
		q = q.Where("audit_status = ?", *query.AuditStatus)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ? OR coupon_no ILIKE ?", like, like)
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

func (r *couponRepository) IncrIssuedCount(id uint) error {
	return r.db.Model(&model.Dh114Coupon{}).Where("id = ?", id).
		UpdateColumn("issued_count", gorm.Expr("issued_count + 1")).Error
}

func (r *couponRepository) IncrUsedCount(id uint) error {
	return r.db.Model(&model.Dh114Coupon{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}
