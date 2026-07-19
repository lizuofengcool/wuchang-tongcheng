// Package repository 担保交易 + 推荐记录 + 合同电子化数据访问层
// 依据 v3.2.1 架构方案：对标瓜子/人人车/懂车帝
package repository

import (
	"time"

	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== Escrow 担保交易 =====

// EscrowListOptions 担保交易列表过滤条件（C 端：按用户/车源过滤）
type EscrowListOptions struct {
	UserID     uint   // 当前用户 ID（payer 或 payee）
	Role       string // buyer(payer) / seller(payee) / all
	CarID      uint
	ListingID  uint
	EscrowType string // deposit/full_payment/down_payment/commission/balance
	Status     *int
	PayMethod  string
	StartDate  *time.Time
	EndDate    *time.Time
}

// EscrowAdminListOptions 担保交易管理后台过滤条件（M 端：跨地区）
type EscrowAdminListOptions struct {
	RegionID   uint
	EscrowNo   string
	PayTradeNo string
	UserID     uint
	CarID      uint
	DealerID   uint
	EscrowType string
	Status     *int
	PayMethod  string
	StartDate  *time.Time
	EndDate    *time.Time
}

// EscrowRepository 担保交易仓储接口
type EscrowRepository interface {
	Create(e *model.CarEscrow) error
	FindByID(id uint) (*model.CarEscrow, error)
	FindByEscrowNo(escrowNo string) (*model.CarEscrow, error)
	FindByCarID(carID uint) (*model.CarEscrow, error)
	FindByContractID(contractID uint) (*model.CarEscrow, error)
	Update(id uint, fields map[string]interface{}) error
	UpdateStatus(id uint, status int, extraFields map[string]interface{}) error
	List(opts EscrowListOptions, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	AdminList(opts EscrowAdminListOptions, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	ListByPayer(payerID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	ListByPayee(payeeID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error)
	ListPendingAutoRelease() ([]model.CarEscrow, error)
	ListDisputed() ([]model.CarEscrow, error)
	CountByStatus(regionID uint, status int) (int64, error)
	CountByPayer(payerID uint) (int64, error)
	CountByPayee(payeeID uint) (int64, error)
	SumAmountByStatus(regionID uint, status int) (float64, error)
}

type escrowRepository struct {
	db *gorm.DB
}

// NewEscrowRepository 创建担保交易仓储实例
func NewEscrowRepository(db *gorm.DB) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(e *model.CarEscrow) error {
	return r.db.Create(e).Error
}

func (r *escrowRepository) FindByID(id uint) (*model.CarEscrow, error) {
	var e model.CarEscrow
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByEscrowNo(escrowNo string) (*model.CarEscrow, error) {
	var e model.CarEscrow
	if err := r.db.Where("escrow_no = ?", escrowNo).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByCarID(carID uint) (*model.CarEscrow, error) {
	var e model.CarEscrow
	if err := r.db.Where("car_id = ?", carID).Order("created_at DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByContractID(contractID uint) (*model.CarEscrow, error) {
	var e model.CarEscrow
	if err := r.db.Where("contract_id = ?", contractID).Order("created_at DESC").First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarEscrow{}).Where("id = ?", id).Updates(fields).Error
}

func (r *escrowRepository) UpdateStatus(id uint, status int, extraFields map[string]interface{}) error {
	updates := map[string]interface{}{"status": status}
	for k, v := range extraFields {
		updates[k] = v
	}
	return r.db.Model(&model.CarEscrow{}).Where("id = ?", id).Updates(updates).Error
}

func (r *escrowRepository) List(opts EscrowListOptions, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEscrow
	var total int64

	q := r.db.Model(&model.CarEscrow{})
	switch opts.Role {
	case "buyer":
		q = q.Where("payer_id = ?", opts.UserID)
	case "seller":
		q = q.Where("payee_id = ?", opts.UserID)
	case "all":
		q = q.Where("payer_id = ? OR payee_id = ?", opts.UserID, opts.UserID)
	}
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.ListingID > 0 {
		q = q.Where("listing_id = ?", opts.ListingID)
	}
	if opts.EscrowType != "" {
		q = q.Where("escrow_type = ?", opts.EscrowType)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.PayMethod != "" {
		q = q.Where("pay_method = ?", opts.PayMethod)
	}
	if opts.StartDate != nil {
		q = q.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("created_at <= ?", *opts.EndDate)
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

func (r *escrowRepository) AdminList(opts EscrowAdminListOptions, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEscrow
	var total int64

	q := r.db.Model(&model.CarEscrow{})
	if opts.RegionID > 0 {
		q = q.Where("region_id = ?", opts.RegionID)
	}
	if opts.EscrowNo != "" {
		q = q.Where("escrow_no = ?", opts.EscrowNo)
	}
	if opts.PayTradeNo != "" {
		q = q.Where("pay_trade_no = ?", opts.PayTradeNo)
	}
	if opts.UserID > 0 {
		q = q.Where("payer_id = ? OR payee_id = ?", opts.UserID, opts.UserID)
	}
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.DealerID > 0 {
		q = q.Where("dealer_id = ?", opts.DealerID)
	}
	if opts.EscrowType != "" {
		q = q.Where("escrow_type = ?", opts.EscrowType)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.PayMethod != "" {
		q = q.Where("pay_method = ?", opts.PayMethod)
	}
	if opts.StartDate != nil {
		q = q.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("created_at <= ?", *opts.EndDate)
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

func (r *escrowRepository) ListByPayer(payerID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	return r.List(EscrowListOptions{UserID: payerID, Role: "buyer"}, pagination)
}

func (r *escrowRepository) ListByPayee(payeeID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	return r.List(EscrowListOptions{UserID: payeeID, Role: "seller"}, pagination)
}

func (r *escrowRepository) ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEscrow
	var total int64
	q := r.db.Model(&model.CarEscrow{}).Where("car_id = ?", carID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListByDealer(dealerID uint, pagination *utils.Pagination) ([]model.CarEscrow, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarEscrow
	var total int64
	q := r.db.Model(&model.CarEscrow{}).Where("dealer_id = ?", dealerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListPendingAutoRelease() ([]model.CarEscrow, error) {
	var list []model.CarEscrow
	if err := r.db.Where("status = ? AND auto_release_at IS NOT NULL AND auto_release_at <= NOW()",
		model.EscrowStatusPaid).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *escrowRepository) ListDisputed() ([]model.CarEscrow, error) {
	var list []model.CarEscrow
	if err := r.db.Where("status IN ?", []int{model.EscrowStatusDisputed, model.EscrowStatusArbitrated}).
		Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *escrowRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarEscrow{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *escrowRepository) CountByPayer(payerID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarEscrow{}).Where("payer_id = ?", payerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *escrowRepository) CountByPayee(payeeID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarEscrow{}).Where("payee_id = ?", payeeID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *escrowRepository) SumAmountByStatus(regionID uint, status int) (float64, error) {
	var amount float64
	q := r.db.Model(&model.CarEscrow{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Select("COALESCE(SUM(amount),0)").Scan(&amount).Error; err != nil {
		return 0, err
	}
	return amount, nil
}

// ===== Recommendation 推荐记录 =====

// RecommendationListOptions 推荐列表过滤条件
type RecommendationListOptions struct {
	UserID    uint
	RecType   string // car_to_user/user_to_car/similar/nearby/recently_viewed/hot
	Source    string // ai/manual/hot/new
	Status    *int
	CarID     uint
	MinScore  float64
	StartDate *time.Time
	EndDate   *time.Time
}

// RecommendationRepository 推荐记录仓储接口
type RecommendationRepository interface {
	Create(rec *model.CarRecommendation) error
	BatchCreate(recs []model.CarRecommendation) error
	FindByID(id uint) (*model.CarRecommendation, error)
	FindByUserAndCar(userID, carID uint, recType string) (*model.CarRecommendation, error)
	Update(id uint, fields map[string]interface{}) error
	UpdateStatus(id uint, status int, extraFields map[string]interface{}) error
	List(opts RecommendationListOptions, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error)
	ListByUser(userID uint, recType string, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error)
	ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error)
	ListByStatus(status int, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error)
	ListPending(userID uint, recType string, limit int) ([]model.CarRecommendation, error)
	ListExpired(before time.Time) ([]model.CarRecommendation, error)
	CountByStatus(regionID uint, status int) (int64, error)
	CountByUser(userID uint) (int64, error)
	DeleteByUser(userID uint) error
}

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建推荐记录仓储实例
func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

func (r *recommendationRepository) Create(rec *model.CarRecommendation) error {
	return r.db.Create(rec).Error
}

func (r *recommendationRepository) BatchCreate(recs []model.CarRecommendation) error {
	if len(recs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(recs, 100).Error
}

func (r *recommendationRepository) FindByID(id uint) (*model.CarRecommendation, error) {
	var rec model.CarRecommendation
	if err := r.db.First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) FindByUserAndCar(userID, carID uint, recType string) (*model.CarRecommendation, error) {
	var rec model.CarRecommendation
	q := r.db.Where("user_id = ? AND car_id = ?", userID, carID)
	if recType != "" {
		q = q.Where("rec_type = ?", recType)
	}
	if err := q.First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarRecommendation{}).Where("id = ?", id).Updates(fields).Error
}

func (r *recommendationRepository) UpdateStatus(id uint, status int, extraFields map[string]interface{}) error {
	updates := map[string]interface{}{"status": status}
	for k, v := range extraFields {
		updates[k] = v
	}
	return r.db.Model(&model.CarRecommendation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *recommendationRepository) List(opts RecommendationListOptions, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarRecommendation
	var total int64

	q := r.db.Model(&model.CarRecommendation{})
	if opts.UserID > 0 {
		q = q.Where("user_id = ?", opts.UserID)
	}
	if opts.RecType != "" {
		q = q.Where("rec_type = ?", opts.RecType)
	}
	if opts.Source != "" {
		q = q.Where("source = ?", opts.Source)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.MinScore > 0 {
		q = q.Where("score >= ?", opts.MinScore)
	}
	if opts.StartDate != nil {
		q = q.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("created_at <= ?", *opts.EndDate)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("score DESC, created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *recommendationRepository) ListByUser(userID uint, recType string, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error) {
	return r.List(RecommendationListOptions{UserID: userID, RecType: recType}, pagination)
}

func (r *recommendationRepository) ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error) {
	return r.List(RecommendationListOptions{CarID: carID}, pagination)
}

func (r *recommendationRepository) ListByStatus(status int, pagination *utils.Pagination) ([]model.CarRecommendation, int64, error) {
	s := status
	return r.List(RecommendationListOptions{Status: &s}, pagination)
}

func (r *recommendationRepository) ListPending(userID uint, recType string, limit int) ([]model.CarRecommendation, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var list []model.CarRecommendation
	q := r.db.Where("user_id = ? AND status = ?", userID, model.RecStatusPending)
	if recType != "" {
		q = q.Where("rec_type = ?", recType)
	}
	if err := q.Order("score DESC, created_at DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *recommendationRepository) ListExpired(before time.Time) ([]model.CarRecommendation, error) {
	var list []model.CarRecommendation
	if err := r.db.Where("status = ? AND created_at < ?",
		model.RecStatusPending, before).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *recommendationRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarRecommendation{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *recommendationRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarRecommendation{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *recommendationRepository) DeleteByUser(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.CarRecommendation{}).Error
}

// ===== Contract 合同电子化 =====

// ContractListOptions 合同列表过滤条件（C 端：按用户过滤）
type ContractListOptions struct {
	UserID        uint   // 当前用户 ID（seller 或 buyer）
	Role          string // seller/buyer/all
	CarID         uint
	ListingID     uint
	ContractType  string // sale/replace/rental/finance
	PaymentMethod string // full/loan/installment/deposit
	Status        *int
	StartDate     *time.Time
	EndDate       *time.Time
}

// ContractAdminListOptions 合同管理后台过滤条件（M 端：跨地区）
type ContractAdminListOptions struct {
	RegionID      uint
	ContractNo    string
	UserID        uint
	CarID         uint
	SellerID      uint
	BuyerID       uint
	AgentID       uint
	ContractType  string
	PaymentMethod string
	Status        *int
	StartDate     *time.Time
	EndDate       *time.Time
}

// ContractRepository 合同仓储接口
type ContractRepository interface {
	Create(c *model.CarContract) error
	FindByID(id uint) (*model.CarContract, error)
	FindByContractNo(contractNo string) (*model.CarContract, error)
	FindByCarID(carID uint) (*model.CarContract, error)
	FindByListingID(listingID uint) (*model.CarContract, error)
	Update(id uint, fields map[string]interface{}) error
	UpdateStatus(id uint, status int, extraFields map[string]interface{}) error
	List(opts ContractListOptions, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	AdminList(opts ContractAdminListOptions, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	ListBySeller(sellerID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	ListByBuyer(buyerID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	ListByAgent(agentID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error)
	CountByStatus(regionID uint, status int) (int64, error)
	CountBySeller(sellerID uint) (int64, error)
	CountByBuyer(buyerID uint) (int64, error)
	SumDealPriceByStatus(regionID uint, status int) (float64, error)
}

type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository 创建合同仓储实例
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &contractRepository{db: db}
}

func (r *contractRepository) Create(c *model.CarContract) error {
	return r.db.Create(c).Error
}

func (r *contractRepository) FindByID(id uint) (*model.CarContract, error) {
	var c model.CarContract
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) FindByContractNo(contractNo string) (*model.CarContract, error) {
	var c model.CarContract
	if err := r.db.Where("contract_no = ?", contractNo).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) FindByCarID(carID uint) (*model.CarContract, error) {
	var c model.CarContract
	if err := r.db.Where("car_id = ?", carID).Order("created_at DESC").First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) FindByListingID(listingID uint) (*model.CarContract, error) {
	var c model.CarContract
	if err := r.db.Where("listing_id = ?", listingID).Order("created_at DESC").First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarContract{}).Where("id = ?", id).Updates(fields).Error
}

func (r *contractRepository) UpdateStatus(id uint, status int, extraFields map[string]interface{}) error {
	updates := map[string]interface{}{"status": status}
	for k, v := range extraFields {
		updates[k] = v
	}
	return r.db.Model(&model.CarContract{}).Where("id = ?", id).Updates(updates).Error
}

func (r *contractRepository) List(opts ContractListOptions, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarContract
	var total int64

	q := r.db.Model(&model.CarContract{})
	switch opts.Role {
	case "seller":
		q = q.Where("seller_id = ?", opts.UserID)
	case "buyer":
		q = q.Where("buyer_id = ?", opts.UserID)
	case "all":
		q = q.Where("seller_id = ? OR buyer_id = ?", opts.UserID, opts.UserID)
	}
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.ListingID > 0 {
		q = q.Where("listing_id = ?", opts.ListingID)
	}
	if opts.ContractType != "" {
		q = q.Where("contract_type = ?", opts.ContractType)
	}
	if opts.PaymentMethod != "" {
		q = q.Where("payment_method = ?", opts.PaymentMethod)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.StartDate != nil {
		q = q.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("created_at <= ?", *opts.EndDate)
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

func (r *contractRepository) AdminList(opts ContractAdminListOptions, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarContract
	var total int64

	q := r.db.Model(&model.CarContract{})
	if opts.RegionID > 0 {
		q = q.Where("region_id = ?", opts.RegionID)
	}
	if opts.ContractNo != "" {
		q = q.Where("contract_no = ?", opts.ContractNo)
	}
	if opts.UserID > 0 {
		q = q.Where("seller_id = ? OR buyer_id = ?", opts.UserID, opts.UserID)
	}
	if opts.CarID > 0 {
		q = q.Where("car_id = ?", opts.CarID)
	}
	if opts.SellerID > 0 {
		q = q.Where("seller_id = ?", opts.SellerID)
	}
	if opts.BuyerID > 0 {
		q = q.Where("buyer_id = ?", opts.BuyerID)
	}
	if opts.AgentID > 0 {
		q = q.Where("agent_id = ?", opts.AgentID)
	}
	if opts.ContractType != "" {
		q = q.Where("contract_type = ?", opts.ContractType)
	}
	if opts.PaymentMethod != "" {
		q = q.Where("payment_method = ?", opts.PaymentMethod)
	}
	if opts.Status != nil {
		q = q.Where("status = ?", *opts.Status)
	}
	if opts.StartDate != nil {
		q = q.Where("created_at >= ?", *opts.StartDate)
	}
	if opts.EndDate != nil {
		q = q.Where("created_at <= ?", *opts.EndDate)
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

func (r *contractRepository) ListBySeller(sellerID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	return r.List(ContractListOptions{UserID: sellerID, Role: "seller"}, pagination)
}

func (r *contractRepository) ListByBuyer(buyerID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	return r.List(ContractListOptions{UserID: buyerID, Role: "buyer"}, pagination)
}

func (r *contractRepository) ListByCarID(carID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarContract
	var total int64
	q := r.db.Model(&model.CarContract{}).Where("car_id = ?", carID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) ListByAgent(agentID uint, pagination *utils.Pagination) ([]model.CarContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarContract
	var total int64
	q := r.db.Model(&model.CarContract{}).Where("agent_id = ?", agentID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarContract{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *contractRepository) CountBySeller(sellerID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarContract{}).Where("seller_id = ?", sellerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *contractRepository) CountByBuyer(buyerID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarContract{}).Where("buyer_id = ?", buyerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *contractRepository) SumDealPriceByStatus(regionID uint, status int) (float64, error) {
	var amount float64
	q := r.db.Model(&model.CarContract{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Select("COALESCE(SUM(deal_price),0)").Scan(&amount).Error; err != nil {
		return 0, err
	}
	return amount, nil
}
