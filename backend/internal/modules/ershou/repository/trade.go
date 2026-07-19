// Package repository 推广/物流/担保/退款数据访问层
// 依据 v3.2.1 架构方案：对标闲鱼/转转/瓜子
package repository

import (
	"wuchang-tongcheng/internal/modules/ershou/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Promotion =====

// PromotionRepository 推广仓储接口
type PromotionRepository interface {
	Create(p *model.ErshouPromotion) error
	FindByID(id uint) (*model.ErshouPromotion, error)
	ListByErshouID(ershouID uint) ([]model.ErshouPromotion, error)
	Update(id uint, fields map[string]interface{}) error
	ListActive(now interface{}) ([]model.ErshouPromotion, error)
	Stats(ershouID uint) (*model.ErshouPromotion, error)
	ListExpired() ([]model.ErshouPromotion, error)
	IncrStats(id uint, field string, n int) error
}

type promotionRepository struct {
	db *gorm.DB
}

// NewPromotionRepository 创建推广仓储实例
func NewPromotionRepository(db *gorm.DB) PromotionRepository {
	return &promotionRepository{db: db}
}

func (r *promotionRepository) Create(p *model.ErshouPromotion) error {
	return r.db.Create(p).Error
}

func (r *promotionRepository) FindByID(id uint) (*model.ErshouPromotion, error) {
	var p model.ErshouPromotion
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *promotionRepository) ListByErshouID(ershouID uint) ([]model.ErshouPromotion, error) {
	var list []model.ErshouPromotion
	if err := r.db.Where("ershou_id = ?", ershouID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *promotionRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouPromotion{}).Where("id = ?", id).Updates(fields).Error
}

func (r *promotionRepository) ListActive(now interface{}) ([]model.ErshouPromotion, error) {
	var list []model.ErshouPromotion
	if err := r.db.Where("status = ? AND start_time <= ? AND end_time >= ?",
		model.PromotionStatusActive, now, now).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *promotionRepository) Stats(ershouID uint) (*model.ErshouPromotion, error) {
	// 返回该物品所有推广累计统计（聚合行）
	var p model.ErshouPromotion
	err := r.db.Model(&model.ErshouPromotion{}).
		Select("COALESCE(SUM(impression_count),0) AS impression_count, "+
			"COALESCE(SUM(click_count),0) AS click_count, "+
			"COALESCE(SUM(order_count),0) AS order_count, "+
			"COALESCE(SUM(amount),0) AS amount, "+
			"COALESCE(SUM(fav_count),0) AS fav_count, "+
			"COALESCE(SUM(consult_count),0) AS consult_count").
		Where("ershou_id = ?", ershouID).
		Scan(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *promotionRepository) ListExpired() ([]model.ErshouPromotion, error) {
	var list []model.ErshouPromotion
	if err := r.db.Where("status = ? AND end_time < NOW()", model.PromotionStatusActive).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *promotionRepository) IncrStats(id uint, field string, n int) error {
	return r.db.Model(&model.ErshouPromotion{}).Where("id = ?", id).
		UpdateColumn(field, gorm.Expr(field+" + ?", n)).Error
}

// ===== Logistics =====

// LogisticsRepository 物流仓储接口
type LogisticsRepository interface {
	Create(l *model.ErshouLogistics) error
	FindByID(id uint) (*model.ErshouLogistics, error)
	FindByOrderID(orderID uint) (*model.ErshouLogistics, error)
	Update(id uint, fields map[string]interface{}) error
	ListByErshouID(ershouID uint) ([]model.ErshouLogistics, error)
}

type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建物流仓储实例
func NewLogisticsRepository(db *gorm.DB) LogisticsRepository {
	return &logisticsRepository{db: db}
}

func (r *logisticsRepository) Create(l *model.ErshouLogistics) error {
	return r.db.Create(l).Error
}

func (r *logisticsRepository) FindByID(id uint) (*model.ErshouLogistics, error) {
	var l model.ErshouLogistics
	if err := r.db.First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logisticsRepository) FindByOrderID(orderID uint) (*model.ErshouLogistics, error) {
	var l model.ErshouLogistics
	if err := r.db.Where("order_id = ?", orderID).First(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logisticsRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouLogistics{}).Where("id = ?", id).Updates(fields).Error
}

func (r *logisticsRepository) ListByErshouID(ershouID uint) ([]model.ErshouLogistics, error) {
	var list []model.ErshouLogistics
	if err := r.db.Where("ershou_id = ?", ershouID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== Escrow =====

// EscrowRepository 担保仓储接口
type EscrowRepository interface {
	Create(e *model.ErshouEscrow) error
	FindByID(id uint) (*model.ErshouEscrow, error)
	FindByOrderID(orderID uint) (*model.ErshouEscrow, error)
	Update(id uint, fields map[string]interface{}) error
	ListPendingAutoRelease() ([]model.ErshouEscrow, error)
}

type escrowRepository struct {
	db *gorm.DB
}

// NewEscrowRepository 创建担保仓储实例
func NewEscrowRepository(db *gorm.DB) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(e *model.ErshouEscrow) error {
	return r.db.Create(e).Error
}

func (r *escrowRepository) FindByID(id uint) (*model.ErshouEscrow, error) {
	var e model.ErshouEscrow
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByOrderID(orderID uint) (*model.ErshouEscrow, error) {
	var e model.ErshouEscrow
	if err := r.db.Where("order_id = ?", orderID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouEscrow{}).Where("id = ?", id).Updates(fields).Error
}

func (r *escrowRepository) ListPendingAutoRelease() ([]model.ErshouEscrow, error) {
	var list []model.ErshouEscrow
	if err := r.db.Where("status = ? AND auto_release_at <= NOW()",
		model.EscrowStatusReleased).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ===== Refund =====

// RefundRepository 退款仓储接口
type RefundRepository interface {
	Create(rf *model.ErshouRefund) error
	FindByID(id uint) (*model.ErshouRefund, error)
	FindByRefundNo(refundNo string) (*model.ErshouRefund, error)
	FindByOrderID(orderID uint) (*model.ErshouRefund, error)
	Update(id uint, fields map[string]interface{}) error
	List(query RefundListQuery, pagination *utils.Pagination) ([]model.ErshouRefund, int64, error)
}

// RefundListQuery 退款列表查询
type RefundListQuery struct {
	UserID   uint
	Role     string // buyer/seller/all
	Status   *int
	RefundNo string
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository 创建退款仓储实例
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(rf *model.ErshouRefund) error {
	return r.db.Create(rf).Error
}

func (r *refundRepository) FindByID(id uint) (*model.ErshouRefund, error) {
	var rf model.ErshouRefund
	if err := r.db.First(&rf, id).Error; err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *refundRepository) FindByRefundNo(refundNo string) (*model.ErshouRefund, error) {
	var rf model.ErshouRefund
	if err := r.db.Where("refund_no = ?", refundNo).First(&rf).Error; err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *refundRepository) FindByOrderID(orderID uint) (*model.ErshouRefund, error) {
	var rf model.ErshouRefund
	if err := r.db.Where("order_id = ?", orderID).First(&rf).Error; err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *refundRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.ErshouRefund{}).Where("id = ?", id).Updates(fields).Error
}

func (r *refundRepository) List(query RefundListQuery, pagination *utils.Pagination) ([]model.ErshouRefund, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.ErshouRefund
	var total int64

	q := r.db.Model(&model.ErshouRefund{})
	switch query.Role {
	case "buyer":
		q = q.Where("buyer_id = ?", query.UserID)
	case "seller":
		q = q.Where("seller_id = ?", query.UserID)
	case "all":
		q = q.Where("buyer_id = ? OR seller_id = ?", query.UserID, query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.RefundNo != "" {
		q = q.Where("refund_no = ?", query.RefundNo)
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
