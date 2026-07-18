// Package repository 团购优惠券数据访问层
package repository

import (
	"errors"
	"time"

	"wuchang-tongcheng/internal/modules/groupbuy/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== 团购商品仓储 =====

// GroupBuyRepository 团购仓储接口
type GroupBuyRepository interface {
	Create(gb *model.GroupBuy) error
	FindByID(id uint) (*model.GroupBuy, error)
	Update(gb *model.GroupBuy) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, pagination *utils.Pagination, keyword string, status, isRecommend int, shopID uint) ([]model.GroupBuy, int64, error)
}

type groupBuyRepository struct {
	db *gorm.DB
}

// NewGroupBuyRepository 创建团购仓储
func NewGroupBuyRepository(db *gorm.DB) GroupBuyRepository {
	return &groupBuyRepository{db: db}
}

func (r *groupBuyRepository) Create(gb *model.GroupBuy) error {
	return r.db.Create(gb).Error
}

func (r *groupBuyRepository) FindByID(id uint) (*model.GroupBuy, error) {
	var gb model.GroupBuy
	if err := r.db.First(&gb, id).Error; err != nil {
		return nil, err
	}
	return &gb, nil
}

func (r *groupBuyRepository) Update(gb *model.GroupBuy) error {
	return r.db.Save(gb).Error
}

func (r *groupBuyRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.GroupBuy{}).Where("id = ?", id).Updates(fields).Error
}

func (r *groupBuyRepository) Delete(id uint) error {
	return r.db.Delete(&model.GroupBuy{}, id).Error
}

func (r *groupBuyRepository) List(regionID uint, pagination *utils.Pagination, keyword string, status, isRecommend int, shopID uint) ([]model.GroupBuy, int64, error) {
	var list []model.GroupBuy
	var total int64

	query := r.db.Model(&model.GroupBuy{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status >= 0 && status <= 2 {
		query = query.Where("status = ?", status)
	}
	if isRecommend == 0 || isRecommend == 1 {
		query = query.Where("is_recommend = ?", isRecommend)
	}
	if shopID > 0 {
		query = query.Where("shop_id = ?", shopID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("sort DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ===== 优惠券仓储 =====

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	Create(c *model.Coupon) error
	FindByID(id uint) (*model.Coupon, error)
	Update(c *model.Coupon) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(regionID uint, pagination *utils.Pagination, status, ctype int) ([]model.Coupon, int64, error)
	IncrReceivedCount(id uint) error
	// AvailableList 可领取的优惠券（启用且在有效期内、未达发放上限）
	AvailableList(regionID uint, pagination *utils.Pagination) ([]model.Coupon, int64, error)
}

type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建优惠券仓储
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

func (r *couponRepository) Update(c *model.Coupon) error {
	return r.db.Save(c).Error
}

func (r *couponRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", id).Updates(fields).Error
}

func (r *couponRepository) Delete(id uint) error {
	return r.db.Delete(&model.Coupon{}, id).Error
}

func (r *couponRepository) List(regionID uint, pagination *utils.Pagination, status, ctype int) ([]model.Coupon, int64, error) {
	var list []model.Coupon
	var total int64

	query := r.db.Model(&model.Coupon{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if status == 0 || status == 1 {
		query = query.Where("status = ?", status)
	}
	if ctype >= 1 && ctype <= 3 {
		query = query.Where("type = ?", ctype)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *couponRepository) IncrReceivedCount(id uint) error {
	return r.db.Model(&model.Coupon{}).Where("id = ?", id).
		UpdateColumn("received_count", gorm.Expr("received_count + 1")).Error
}

func (r *couponRepository) AvailableList(regionID uint, pagination *utils.Pagination) ([]model.Coupon, int64, error) {
	var list []model.Coupon
	var total int64

	now := time.Now()
	query := r.db.Model(&model.Coupon{}).Where("status = ?", 1)
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	// 在有效期内
	query = query.Where("(start_time IS NULL OR start_time <= ?)", now)
	query = query.Where("(end_time IS NULL OR end_time >= ?)", now)
	// 未达发放总量（total_count=0 表示不限）
	query = query.Where("total_count = 0 OR received_count < total_count")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ===== 用户优惠券仓储 =====

// UserCouponRepository 用户优惠券仓储接口
type UserCouponRepository interface {
	Create(uc *model.UserCoupon) error
	FindByID(id uint) (*model.UserCoupon, error)
	FindActiveByUserAndCoupon(userID, couponID uint) (*model.UserCoupon, error)
	ListByUser(userID uint, pagination *utils.Pagination, status int) ([]model.UserCoupon, int64, error)
	CountByUserAndCoupon(userID, couponID uint) (int64, error)
	UpdateFields(id uint, fields map[string]interface{}) error
}

type userCouponRepository struct {
	db *gorm.DB
}

// NewUserCouponRepository 创建用户优惠券仓储
func NewUserCouponRepository(db *gorm.DB) UserCouponRepository {
	return &userCouponRepository{db: db}
}

func (r *userCouponRepository) Create(uc *model.UserCoupon) error {
	return r.db.Create(uc).Error
}

func (r *userCouponRepository) FindByID(id uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.First(&uc, id).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *userCouponRepository) FindActiveByUserAndCoupon(userID, couponID uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.Where("user_id = ? AND coupon_id = ? AND status = 0", userID, couponID).First(&uc).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

func (r *userCouponRepository) ListByUser(userID uint, pagination *utils.Pagination, status int) ([]model.UserCoupon, int64, error) {
	var list []model.UserCoupon
	var total int64

	query := r.db.Model(&model.UserCoupon{}).Where("user_id = ?", userID)
	if status >= 0 && status <= 2 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *userCouponRepository) CountByUserAndCoupon(userID, couponID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserCoupon{}).Where("user_id = ? AND coupon_id = ?", userID, couponID).Count(&count).Error
	return count, err
}

func (r *userCouponRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.UserCoupon{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 团购订单仓储 =====

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(order *model.GroupBuyOrder) error
	FindByID(id uint) (*model.GroupBuyOrder, error)
	FindByOrderNo(orderNo string) (*model.GroupBuyOrder, error)
	UpdateFields(id uint, fields map[string]interface{}) error
	List(regionID uint, pagination *utils.Pagination, userID, groupbuyID uint, status, payStatus int) ([]model.GroupBuyOrder, int64, error)
	// CountByUserAndGroupBuy 统计用户在某团购下的有效订单数（非取消）
	CountByUserAndGroupBuy(userID, groupbuyID uint) (int64, error)
	// SumQuantityByUserAndGroupBuy 统计用户在某团购下的有效购买数量（非取消）
	SumQuantityByUserAndGroupBuy(userID, groupbuyID uint) (int64, error)
	// CreateOrderInTransaction 事务创建订单：扣减库存、标记优惠券已用
	CreateOrderInTransaction(order *model.GroupBuyOrder, groupbuyID, userCouponID uint, quantity int) error
	// CancelOrderInTransaction 事务取消订单：更新订单状态、回滚库存、释放优惠券
	CancelOrderInTransaction(order *model.GroupBuyOrder) error
	// VerifyOrder 事务核销订单：更新核销状态
	VerifyOrder(orderID uint, verifyCode string) error
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.GroupBuyOrder) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*model.GroupBuyOrder, error) {
	var order model.GroupBuyOrder
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderNo(orderNo string) (*model.GroupBuyOrder, error) {
	var order model.GroupBuyOrder
	if err := r.db.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.GroupBuyOrder{}).Where("id = ?", id).Updates(fields).Error
}

func (r *orderRepository) List(regionID uint, pagination *utils.Pagination, userID, groupbuyID uint, status, payStatus int) ([]model.GroupBuyOrder, int64, error) {
	var list []model.GroupBuyOrder
	var total int64

	query := r.db.Model(&model.GroupBuyOrder{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if groupbuyID > 0 {
		query = query.Where("groupbuy_id = ?", groupbuyID)
	}
	if status >= 0 && status <= 4 {
		query = query.Where("status = ?", status)
	}
	if payStatus >= 0 && payStatus <= 2 {
		query = query.Where("pay_status = ?", payStatus)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Scopes(utils.Paginate(pagination)).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *orderRepository) CountByUserAndGroupBuy(userID, groupbuyID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.GroupBuyOrder{}).
		Where("user_id = ? AND groupbuy_id = ? AND status != 3", userID, groupbuyID).
		Count(&count).Error
	return count, err
}

func (r *orderRepository) SumQuantityByUserAndGroupBuy(userID, groupbuyID uint) (int64, error) {
	var total int64
	err := r.db.Model(&model.GroupBuyOrder{}).
		Where("user_id = ? AND groupbuy_id = ? AND status != 3", userID, groupbuyID).
		Select("COALESCE(SUM(quantity), 0)").Scan(&total).Error
	return total, err
}

// CreateOrderInTransaction 事务创建订单
func (r *orderRepository) CreateOrderInTransaction(order *model.GroupBuyOrder, groupbuyID, userCouponID uint, quantity int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建订单
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		// 2. 条件扣减库存（要求库存充足）
		result := tx.Model(&model.GroupBuy{}).Where("id = ? AND stock >= ?", groupbuyID, quantity).
			Updates(map[string]interface{}{
				"stock":      gorm.Expr("stock - ?", quantity),
				"sold_count": gorm.Expr("sold_count + ?", quantity),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("库存不足")
		}
		// 3. 标记用户优惠券已使用
		if userCouponID > 0 {
			now := time.Now()
			result2 := tx.Model(&model.UserCoupon{}).Where("id = ? AND status = 0", userCouponID).
				Updates(map[string]interface{}{
					"status":   1,
					"order_id": order.ID,
					"used_at":  &now,
				})
			if result2.Error != nil {
				return result2.Error
			}
			if result2.RowsAffected == 0 {
				return errors.New("优惠券不可用或已使用")
			}
			// 优惠券已使用数+1
			tx.Model(&model.Coupon{}).Where("id = ?", order.CouponID).
				UpdateColumn("used_count", gorm.Expr("used_count + 1"))
		}
		return nil
	})
}

// CancelOrderInTransaction 事务取消订单
func (r *orderRepository) CancelOrderInTransaction(order *model.GroupBuyOrder) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 更新订单状态为已取消、支付状态为已退款
		result := tx.Model(&model.GroupBuyOrder{}).
			Where("id = ? AND status = 1", order.ID).
			Updates(map[string]interface{}{
				"status":     3, // 已取消
				"pay_status": 2, // 已退款
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("订单状态不允许取消")
		}
		// 2. 回滚库存
		if err := tx.Model(&model.GroupBuy{}).Where("id = ?", order.GroupBuyID).
			Updates(map[string]interface{}{
				"stock":      gorm.Expr("stock + ?", order.Quantity),
				"sold_count": gorm.Expr("sold_count - ?", order.Quantity),
			}).Error; err != nil {
			return err
		}
		// 3. 释放优惠券
		if order.UserCouponID > 0 && order.CouponID > 0 {
			if err := tx.Model(&model.UserCoupon{}).Where("id = ? AND status = 1", order.UserCouponID).
				Updates(map[string]interface{}{
					"status":   0,
					"order_id": 0,
					"used_at":  nil,
				}).Error; err != nil {
				return err
			}
			tx.Model(&model.Coupon{}).Where("id = ?", order.CouponID).
				UpdateColumn("used_count", gorm.Expr("GREATEST(used_count - 1, 0)"))
		}
		return nil
	})
}

// VerifyOrder 事务核销订单
func (r *orderRepository) VerifyOrder(orderID uint, verifyCode string) error {
	now := time.Now()
	result := r.db.Model(&model.GroupBuyOrder{}).
		Where("id = ? AND verify_code = ? AND status = 1 AND verify_status = 0", orderID, verifyCode).
		Updates(map[string]interface{}{
			"verify_status": 1,
			"verify_at":     &now,
			"status":        2, // 已核销
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("订单核销失败：核销码错误或订单状态不允许核销")
	}
	return nil
}
