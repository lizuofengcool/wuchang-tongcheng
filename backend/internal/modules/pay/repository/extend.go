// Package repository 支付中台扩展数据访问层
// 依据 012_pay_full.sql：交易流水/渠道/商户/回调/争议仲裁
package repository

import (
	"wuchang-tongcheng/internal/modules/pay/model"

	"gorm.io/gorm"
)

// PayExtendRepository 支付中台扩展仓储接口
type PayExtendRepository interface {
	// 交易流水
	CreateTransaction(t *model.Transaction) error
	FindTransactionByNo(txnNo string) (*model.Transaction, error)
	ListTransactions(req *ListTransactionsQuery) ([]model.Transaction, int64, error)
	UpdateTransactionFields(id uint, fields map[string]interface{}) error

	// 渠道
	CreateChannel(c *model.Channel) error
	FindChannelByID(id uint) (*model.Channel, error)
	FindChannelByCode(code string) (*model.Channel, error)
	ListChannels(code string, page, pageSize int) ([]model.Channel, int64, error)
	UpdateChannelFields(id uint, fields map[string]interface{}) error
	DeleteChannel(id uint) error

	// 商户
	CreateMerchant(m *model.Merchant) error
	FindMerchantByID(id uint) (*model.Merchant, error)
	FindMerchantByNo(no string) (*model.Merchant, error)
	ListMerchants(status int, page, pageSize int) ([]model.Merchant, int64, error)
	UpdateMerchantFields(id uint, fields map[string]interface{}) error

	// 回调
	CreateCallback(c *model.Callback) error
	FindCallbackByID(id uint) (*model.Callback, error)
	ListCallbacks(orderNo string, channel string, status int, page, pageSize int) ([]model.Callback, int64, error)
	UpdateCallbackFields(id uint, fields map[string]interface{}) error

	// 担保争议查询
	ListDisputedEscrows(page, pageSize int) ([]model.EscrowAccount, int64, error)

	// 统计
	StatTotalOrders() (int64, error)
	StatTotalAmount() (float64, error)
	StatRefundAmount() (float64, error)
	StatEscrowAmount() (float64, error)
	StatTodayCount() (int64, error)
	StatTodayAmount() (float64, error)
	StatSuccessRate() (float64, error)
	StatRefundRate() (float64, error)
	StatByChannel() ([]ChannelStatRow, error)
}

// ListTransactionsQuery 交易流水列表查询参数
type ListTransactionsQuery struct {
	RegionID  uint
	UserID    uint
	OrderNo   string
	Channel   string
	Status    int
	Page      int
	PageSize  int
}

// ChannelStatRow 渠道统计行
type ChannelStatRow struct {
	Channel string
	Count   int64
	Amount  float64
}

type payExtendRepository struct {
	db *gorm.DB
}

// NewPayExtendRepository 创建扩展仓储实例
func NewPayExtendRepository(db *gorm.DB) PayExtendRepository {
	return &payExtendRepository{db: db}
}

// ===== 交易流水 =====

func (r *payExtendRepository) CreateTransaction(t *model.Transaction) error {
	return r.db.Create(t).Error
}

func (r *payExtendRepository) FindTransactionByNo(txnNo string) (*model.Transaction, error) {
	var t model.Transaction
	if err := r.db.Where("transaction_no = ?", txnNo).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *payExtendRepository) ListTransactions(req *ListTransactionsQuery) ([]model.Transaction, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	var list []model.Transaction
	var total int64
	q := r.db.Model(&model.Transaction{})
	if req.RegionID > 0 {
		q = q.Where("region_id = ?", req.RegionID)
	}
	if req.UserID > 0 {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.OrderNo != "" {
		q = q.Where("order_no = ?", req.OrderNo)
	}
	if req.Channel != "" {
		q = q.Where("channel = ?", req.Channel)
	}
	if req.Status >= 0 {
		q = q.Where("status = ?", req.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC, id DESC").
		Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payExtendRepository) UpdateTransactionFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Transaction{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 渠道 =====

func (r *payExtendRepository) CreateChannel(c *model.Channel) error {
	return r.db.Create(c).Error
}

func (r *payExtendRepository) FindChannelByID(id uint) (*model.Channel, error) {
	var c model.Channel
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *payExtendRepository) FindChannelByCode(code string) (*model.Channel, error) {
	var c model.Channel
	if err := r.db.Where("channel_code = ? AND status = ?", code, 1).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *payExtendRepository) ListChannels(code string, page, pageSize int) ([]model.Channel, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Channel
	var total int64
	q := r.db.Model(&model.Channel{})
	if code != "" {
		q = q.Where("channel_code = ?", code)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("sort ASC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *payExtendRepository) UpdateChannelFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).Updates(fields).Error
}

func (r *payExtendRepository) DeleteChannel(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

// ===== 商户 =====

func (r *payExtendRepository) CreateMerchant(m *model.Merchant) error {
	return r.db.Create(m).Error
}

func (r *payExtendRepository) FindMerchantByID(id uint) (*model.Merchant, error) {
	var m model.Merchant
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *payExtendRepository) FindMerchantByNo(no string) (*model.Merchant, error) {
	var m model.Merchant
	if err := r.db.Where("merchant_no = ?", no).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *payExtendRepository) ListMerchants(status int, page, pageSize int) ([]model.Merchant, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Merchant
	var total int64
	q := r.db.Model(&model.Merchant{})
	if status >= 0 {
		q = q.Where("status = ?", status)
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

func (r *payExtendRepository) UpdateMerchantFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Merchant{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 回调 =====

func (r *payExtendRepository) CreateCallback(c *model.Callback) error {
	return r.db.Create(c).Error
}

func (r *payExtendRepository) FindCallbackByID(id uint) (*model.Callback, error) {
	var c model.Callback
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *payExtendRepository) ListCallbacks(orderNo string, channel string, status int, page, pageSize int) ([]model.Callback, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.Callback
	var total int64
	q := r.db.Model(&model.Callback{})
	if orderNo != "" {
		q = q.Where("order_no = ?", orderNo)
	}
	if channel != "" {
		q = q.Where("channel = ?", channel)
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
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

func (r *payExtendRepository) UpdateCallbackFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Callback{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 担保争议 =====

func (r *payExtendRepository) ListDisputedEscrows(page, pageSize int) ([]model.EscrowAccount, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var list []model.EscrowAccount
	var total int64
	q := r.db.Model(&model.EscrowAccount{}).Where("dispute_status > ?", 0)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("arbitrated_at DESC NULLS LAST, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 统计 =====

func (r *payExtendRepository) StatTotalOrders() (int64, error) {
	var count int64
	err := r.db.Model(&model.PaymentOrder{}).Count(&count).Error
	return count, err
}

func (r *payExtendRepository) StatTotalAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.PaymentOrder{}).
		Where("pay_status = ?", model.PayStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *payExtendRepository) StatRefundAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.RefundOrder{}).
		Where("status = ?", model.RefundStatusRefunded).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *payExtendRepository) StatEscrowAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.EscrowAccount{}).
		Where("status = ?", model.EscrowStatusFrozen).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *payExtendRepository) StatTodayCount() (int64, error) {
	var count int64
	err := r.db.Model(&model.PaymentOrder{}).
		Where("pay_status = ? AND paid_at >= CURRENT_DATE", model.PayStatusPaid).
		Count(&count).Error
	return count, err
}

func (r *payExtendRepository) StatTodayAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.PaymentOrder{}).
		Where("pay_status = ? AND paid_at >= CURRENT_DATE", model.PayStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (r *payExtendRepository) StatSuccessRate() (float64, error) {
	var total int64
	if err := r.db.Model(&model.PaymentOrder{}).Count(&total).Error; err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	var success int64
	if err := r.db.Model(&model.PaymentOrder{}).
		Where("pay_status = ?", model.PayStatusPaid).Count(&success).Error; err != nil {
		return 0, err
	}
	return float64(success) / float64(total), nil
}

func (r *payExtendRepository) StatRefundRate() (float64, error) {
	var total int64
	if err := r.db.Model(&model.PaymentOrder{}).Count(&total).Error; err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	var refunded int64
	if err := r.db.Model(&model.PaymentOrder{}).
		Where("pay_status IN ?", []int{model.PayStatusRefunded, model.PayStatusPartRefund}).Count(&refunded).Error; err != nil {
		return 0, err
	}
	return float64(refunded) / float64(total), nil
}

func (r *payExtendRepository) StatByChannel() ([]ChannelStatRow, error) {
	var rows []ChannelStatRow
	err := r.db.Model(&model.Transaction{}).
		Select("channel, COUNT(*) AS count, COALESCE(SUM(amount), 0) AS amount").
		Where("status = ?", model.TxnStatusSuccess).
		Group("channel").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}
