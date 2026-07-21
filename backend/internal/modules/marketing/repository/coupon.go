// Package repository 营销活动中台数据访问层 - 优惠券（coupon 子域）
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/marketing/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// CouponListQuery 优惠券列表查询
type CouponListQuery struct {
	Type    string
	Status  *int
	Keyword string
}

// UserCouponListQuery 用户优惠券列表查询
type UserCouponListQuery struct {
	UserID   uint
	CouponID uint
	Status   string
}

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	Create(c *model.Coupon) error
	FindByID(id uint) (*model.Coupon, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, query CouponListQuery, pagination *utils.Pagination) ([]model.Coupon, int64, error)
	ListAvailable(regionID uint, pagination *utils.Pagination) ([]model.Coupon, int64, error)
	IncrReceivedCount(id uint) error

	// 用户优惠券
	CreateUserCoupon(uc *model.UserCoupon) error
	FindUserCoupon(userID uint, couponID uint) (*model.UserCoupon, error)
	FindUserCouponByID(id uint) (*model.UserCoupon, error)
	ListUserCoupons(query UserCouponListQuery, pagination *utils.Pagination) ([]model.UserCoupon, int64, error)
	UpdateUserCoupon(id uint, fields map[string]interface{}) error
	CountUserCoupon(userID uint, couponID uint) (int64, error)
	ExpireUserCoupons(now time.Time) (int64, error)
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓储实例
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(c *model.Coupon) error {
	return r.db.Create(c).Error
}

func (r *couponRepository) FindByID(id uint) (*model.Coupon, error) {
	var c model.Coupon
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *couponRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", id).Updates(fields).Error
}

func (r *couponRepository) Delete(id uint) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

func (r *couponRepository) List(regionID uint, query CouponListQuery, pagination *utils.Pagination) ([]model.Coupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Coupon
	var total int64

	q := r.db.Model(&model.Coupon{})
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if query.Type != "" {
		q = q.Where("type = ?", query.Type)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("title ILIKE ?", like)
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

func (r *couponRepository) ListAvailable(regionID uint, pagination *utils.Pagination) ([]model.Coupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.Coupon
	var total int64

	now := time.Now()
	q := r.db.Model(&model.Coupon{}).
		Where("status = ?", model.CouponStatusActive).
		Where("(start_at IS NULL OR start_at <= ?)", now).
		Where("(end_at IS NULL OR end_at >= ?)", now)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
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

func (r *couponRepository) IncrReceivedCount(id uint) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", id).
		UpdateColumn("received_count", gorm.Expr("received_count + 1")).Error
}

// ===== 用户优惠券 =====

func (r *couponRepository) CreateUserCoupon(uc *model.UserCoupon) error {
	return r.db.Create(uc).Error
}

func (r *couponRepository) FindUserCoupon(userID uint, couponID uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Order("id DESC").First(&uc).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *couponRepository) FindUserCouponByID(id uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.First(&uc, id).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *couponRepository) ListUserCoupons(query UserCouponListQuery, pagination *utils.Pagination) ([]model.UserCoupon, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.UserCoupon
	var total int64

	q := r.db.Model(&model.UserCoupon{})
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.CouponID > 0 {
		q = q.Where("coupon_id = ?", query.CouponID)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
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

func (r *couponRepository) UpdateUserCoupon(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.UserCoupon{}).Where("id = ?", id).Updates(fields).Error
}

func (r *couponRepository) CountUserCoupon(userID uint, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserCoupon{}).
		Where("user_id = ? AND coupon_id = ?", userID, couponID).
		Count(&count).Error
	return count, err
}

// ExpireUserCoupons 将所有未使用且优惠券已过期的用户券标记为已过期
// 返回受影响行数
func (r *couponRepository) ExpireUserCoupons(now time.Time) (int64, error) {
	// 关联 coupons 表：用户券未使用，且对应优惠券的 end_at 已过期
	result := r.db.Model(&model.UserCoupon{}).
		Where("status = ?", model.UserCouponStatusUnused).
		Where("coupon_id IN (?)",
			r.db.Model(&model.Coupon{}).Select("id").Where("end_at IS NOT NULL AND end_at < ?", now),
		).
		Update("status", model.UserCouponStatusExpired)
	return result.RowsAffected, result.Error
}
