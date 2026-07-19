// Package repository 支付财务中台精简版数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/pay/model"

	"gorm.io/gorm"
)

// PayRepository 支付中台仓储接口
type PayRepository interface {
	// 订单
	CreateOrder(o *model.PaymentOrder) error
	FindOrderByNo(orderNo string) (*model.PaymentOrder, error)
	FindOrderByID(id uint) (*model.PaymentOrder, error)
	UpdateOrderFields(id uint, fields map[string]interface{}) error
	ListOrders(userID uint, page, pageSize int) ([]model.PaymentOrder, int64, error)

	// 担保交易
	CreateEscrow(e *model.EscrowAccount) error
	FindEscrowByOrderID(orderID uint) (*model.EscrowAccount, error)
	UpdateEscrowFields(id uint, fields map[string]interface{}) error

	// 退款
	CreateRefund(r *model.RefundOrder) error
	FindRefundByNo(refundNo string) (*model.RefundOrder, error)
	FindRefundsByOrderID(orderID uint) ([]model.RefundOrder, error)
	UpdateRefundFields(id uint, fields map[string]interface{}) error

	// 提现
	CreateWithdrawal(w *model.Withdrawal) error
	FindWithdrawalByNo(no string) (*model.Withdrawal, error)
	ListWithdrawals(userID uint, page, pageSize int) ([]model.Withdrawal, int64, error)
	ListPendingWithdrawals(page, pageSize int) ([]model.Withdrawal, int64, error)
	UpdateWithdrawalFields(id uint, fields map[string]interface{}) error

	// 结算
	CreateSettlement(s *model.Settlement) error
	FindSettlementByNo(no string) (*model.Settlement, error)
	ListSettlements(merchantID uint, page, pageSize int) ([]model.Settlement, int64, error)
	UpdateSettlementFields(id uint, fields map[string]interface{}) error

	// 资金账户
	GetOrCreateAccount(userID uint, regionID uint) (*model.Account, error)
	FindAccount(userID uint) (*model.Account, error)
	UpdateAccountFields(id uint, fields map[string]interface{}) error
}

type payRepository struct {
	db *gorm.DB
}

// NewPayRepository 创建仓储实例
func NewPayRepository(db *gorm.DB) PayRepository {
	return &payRepository{db: db}
}

// ===== 订单 =====

func (r *payRepository) CreateOrder(o *model.PaymentOrder) error {
	return r.db.Create(o).Error
}

func (r *payRepository) FindOrderByNo(orderNo string) (*model.PaymentOrder, error) {
	var o model.PaymentOrder
	if err := r.db.Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *payRepository) FindOrderByID(id uint) (*model.PaymentOrder, error) {
	var o model.PaymentOrder
	if err := r.db.First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *payRepository) UpdateOrderFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.PaymentOrder{}).Where("id = ?", id).Updates(fields).Error
}

func (r *payRepository) ListOrders(userID uint, page, pageSize int) ([]model.PaymentOrder, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.PaymentOrder
	var total int64
	q := r.db.Model(&model.PaymentOrder{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 担保交易 =====

func (r *payRepository) CreateEscrow(e *model.EscrowAccount) error {
	return r.db.Create(e).Error
}

func (r *payRepository) FindEscrowByOrderID(orderID uint) (*model.EscrowAccount, error) {
	var e model.EscrowAccount
	if err := r.db.Where("order_id = ?", orderID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *payRepository) UpdateEscrowFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.EscrowAccount{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 退款 =====

func (r *payRepository) CreateRefund(refund *model.RefundOrder) error {
	return r.db.Create(refund).Error
}

func (r *payRepository) FindRefundByNo(refundNo string) (*model.RefundOrder, error) {
	var refund model.RefundOrder
	if err := r.db.Where("refund_no = ?", refundNo).First(&refund).Error; err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *payRepository) FindRefundsByOrderID(orderID uint) ([]model.RefundOrder, error) {
	var list []model.RefundOrder
	if err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *payRepository) UpdateRefundFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.RefundOrder{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 提现 =====

func (r *payRepository) CreateWithdrawal(w *model.Withdrawal) error {
	return r.db.Create(w).Error
}

func (r *payRepository) FindWithdrawalByNo(no string) (*model.Withdrawal, error) {
	var w model.Withdrawal
	if err := r.db.Where("withdrawal_no = ?", no).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *payRepository) ListWithdrawals(userID uint, page, pageSize int) ([]model.Withdrawal, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Withdrawal
	var total int64
	q := r.db.Model(&model.Withdrawal{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payRepository) ListPendingWithdrawals(page, pageSize int) ([]model.Withdrawal, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Withdrawal
	var total int64
	q := r.db.Model(&model.Withdrawal{}).Where("status = ?", model.WithdrawStatusPending)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payRepository) UpdateWithdrawalFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Withdrawal{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 结算 =====

func (r *payRepository) CreateSettlement(s *model.Settlement) error {
	return r.db.Create(s).Error
}

func (r *payRepository) FindSettlementByNo(no string) (*model.Settlement, error) {
	var s model.Settlement
	if err := r.db.Where("settlement_no = ?", no).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *payRepository) ListSettlements(merchantID uint, page, pageSize int) ([]model.Settlement, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var list []model.Settlement
	var total int64
	q := r.db.Model(&model.Settlement{})
	if merchantID > 0 {
		q = q.Where("merchant_id = ?", merchantID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payRepository) UpdateSettlementFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Settlement{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 资金账户 =====

func (r *payRepository) GetOrCreateAccount(userID uint, regionID uint) (*model.Account, error) {
	var acc model.Account
	err := r.db.Where("user_id = ?", userID).First(&acc).Error
	if err == nil {
		return &acc, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	acc = model.Account{
		UserID: userID,
	}
	acc.RegionID = regionID
	if err := r.db.Create(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *payRepository) FindAccount(userID uint) (*model.Account, error) {
	var acc model.Account
	if err := r.db.Where("user_id = ?", userID).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *payRepository) UpdateAccountFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Account{}).Where("id = ?", id).Updates(fields).Error
}
