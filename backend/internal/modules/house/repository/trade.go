// Package repository 担保交易 + 成交记录数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ===== 担保交易 =====

// EscrowRepository 担保交易仓储接口
type EscrowRepository interface {
	Create(e *model.HouseEscrow) error
	FindByID(id uint) (*model.HouseEscrow, error)
	FindByEscrowNo(no string) (*model.HouseEscrow, error)
	Update(e *model.HouseEscrow) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts EscrowListOptions) ([]model.HouseEscrow, int64, error)
	AdminList(req *utils.Pagination, opts EscrowAdminListOptions) ([]model.HouseEscrow, int64, error)
	ListByPayer(payerID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error)
	ListByPayee(payeeID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error)
	ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error)
	ListDisputed(req *utils.Pagination) ([]model.HouseEscrow, int64, error)
	ListToAutoRelease() ([]model.HouseEscrow, error)
	BatchUpdateStatus(ids []uint, status int) (int64, error)
}

// EscrowListOptions 担保交易列表过滤条件（C 端）
type EscrowListOptions struct {
	EscrowType string
	HouseID    uint
	ListingID  uint
	ContractID uint
	PayMethod  string
	Status     *int
	Keyword    string
}

// EscrowAdminListOptions 管理后台担保交易列表过滤条件
type EscrowAdminListOptions struct {
	RegionID   uint
	HouseID    uint
	PayerID    uint
	PayeeID    uint
	AgentID    uint
	EscrowType string
	PayMethod  string
	Status     *int
	Keyword    string
}

type escrowRepository struct {
	db *gorm.DB
}

// NewEscrowRepository 创建担保交易仓储实例
func NewEscrowRepository(db *gorm.DB) EscrowRepository {
	return &escrowRepository{db: db}
}

func (r *escrowRepository) Create(e *model.HouseEscrow) error {
	return r.db.Create(e).Error
}

func (r *escrowRepository) FindByID(id uint) (*model.HouseEscrow, error) {
	var e model.HouseEscrow
	if err := r.db.First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) FindByEscrowNo(no string) (*model.HouseEscrow, error) {
	var e model.HouseEscrow
	if err := r.db.Where("escrow_no = ?", no).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *escrowRepository) Update(e *model.HouseEscrow) error {
	return r.db.Save(e).Error
}

func (r *escrowRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseEscrow{}).Where("id = ?", id).Updates(fields).Error
}

func (r *escrowRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseEscrow{}, id).Error
}

func (r *escrowRepository) List(req *utils.Pagination, opts EscrowListOptions) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{})
	if opts.EscrowType != "" {
		query = query.Where("escrow_type = ?", opts.EscrowType)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.ContractID > 0 {
		query = query.Where("contract_id = ?", opts.ContractID)
	}
	if opts.PayMethod != "" {
		query = query.Where("pay_method = ?", opts.PayMethod)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("escrow_no ILIKE ? OR pay_trade_no ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) AdminList(req *utils.Pagination, opts EscrowAdminListOptions) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.PayerID > 0 {
		query = query.Where("payer_id = ?", opts.PayerID)
	}
	if opts.PayeeID > 0 {
		query = query.Where("payee_id = ?", opts.PayeeID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.EscrowType != "" {
		query = query.Where("escrow_type = ?", opts.EscrowType)
	}
	if opts.PayMethod != "" {
		query = query.Where("pay_method = ?", opts.PayMethod)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("escrow_no ILIKE ? OR pay_trade_no ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListByPayer(payerID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{}).Where("payer_id = ?", payerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListByPayee(payeeID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{}).Where("payee_id = ?", payeeID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{}).Where("house_id = ?", houseID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *escrowRepository) ListDisputed(req *utils.Pagination) ([]model.HouseEscrow, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseEscrow
	var total int64

	query := r.db.Model(&model.HouseEscrow{}).Where("status = ?", model.EscrowStatusDisputed)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListToAutoRelease 查询可自动放款的担保（已支付且 auto_release_at 已过期）
func (r *escrowRepository) ListToAutoRelease() ([]model.HouseEscrow, error) {
	var list []model.HouseEscrow
	if err := r.db.
		Where("status = ? AND auto_release_at IS NOT NULL AND auto_release_at <= NOW()", model.EscrowStatusPaid).
		Order("auto_release_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *escrowRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseEscrow{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}

// ===== 成交记录 =====

// DealRepository 成交记录仓储接口
type DealRepository interface {
	Create(d *model.HouseDeal) error
	FindByID(id uint) (*model.HouseDeal, error)
	FindByDealNo(no string) (*model.HouseDeal, error)
	Update(d *model.HouseDeal) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error
	List(req *utils.Pagination, opts DealListOptions) ([]model.HouseDeal, int64, error)
	AdminList(req *utils.Pagination, opts DealAdminListOptions) ([]model.HouseDeal, int64, error)
	ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error)
	ListBySeller(sellerID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error)
	ListByBuyer(buyerID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error)
	ListByAgent(agentID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error)
	ListByCommunity(communityID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error)
	BatchUpdateStatus(ids []uint, status int) (int64, error)
}

// DealListOptions 成交列表过滤条件（C 端）
type DealListOptions struct {
	DealType string
	HouseID  uint
	Status   *int
	Keyword  string
}

// DealAdminListOptions 管理后台成交列表过滤条件
type DealAdminListOptions struct {
	RegionID     uint
	HouseID      uint
	SellerID     uint
	BuyerID      uint
	AgentID      uint
	CommunityID  uint
	DealType     string
	Status       *int
	Keyword      string
}

type dealRepository struct {
	db *gorm.DB
}

// NewDealRepository 创建成交记录仓储实例
func NewDealRepository(db *gorm.DB) DealRepository {
	return &dealRepository{db: db}
}

func (r *dealRepository) Create(d *model.HouseDeal) error {
	return r.db.Create(d).Error
}

func (r *dealRepository) FindByID(id uint) (*model.HouseDeal, error) {
	var d model.HouseDeal
	if err := r.db.First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *dealRepository) FindByDealNo(no string) (*model.HouseDeal, error) {
	var d model.HouseDeal
	if err := r.db.Where("deal_no = ?", no).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *dealRepository) Update(d *model.HouseDeal) error {
	return r.db.Save(d).Error
}

func (r *dealRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseDeal{}).Where("id = ?", id).Updates(fields).Error
}

func (r *dealRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseDeal{}, id).Error
}

func (r *dealRepository) List(req *utils.Pagination, opts DealListOptions) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{})
	if opts.DealType != "" {
		query = query.Where("deal_type = ?", opts.DealType)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("deal_no ILIKE ? OR seller_name ILIKE ? OR buyer_name ILIKE ? OR agent_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) AdminList(req *utils.Pagination, opts DealAdminListOptions) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.SellerID > 0 {
		query = query.Where("seller_id = ?", opts.SellerID)
	}
	if opts.BuyerID > 0 {
		query = query.Where("buyer_id = ?", opts.BuyerID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.CommunityID > 0 {
		query = query.Where("community_id = ?", opts.CommunityID)
	}
	if opts.DealType != "" {
		query = query.Where("deal_type = ?", opts.DealType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("deal_no ILIKE ? OR seller_name ILIKE ? OR buyer_name ILIKE ? OR agent_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) ListByHouse(houseID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{}).Where("house_id = ?", houseID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) ListBySeller(sellerID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{}).Where("seller_id = ?", sellerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) ListByBuyer(buyerID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{}).Where("buyer_id = ?", buyerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) ListByAgent(agentID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) ListByCommunity(communityID uint, req *utils.Pagination) ([]model.HouseDeal, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseDeal
	var total int64

	query := r.db.Model(&model.HouseDeal{}).Where("community_id = ?", communityID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("deal_date DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *dealRepository) BatchUpdateStatus(ids []uint, status int) (int64, error) {
	result := r.db.Model(&model.HouseDeal{}).Where("id IN ?", ids).Update("status", status)
	return result.RowsAffected, result.Error
}
