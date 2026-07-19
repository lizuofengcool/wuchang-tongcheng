// Package repository 同城车辆买卖数据访问层 - 过户办理
// 依据需求文档 1.10：4 维数据隔离（region_id + user_id）
package repository

import (
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// TransferRepository 过户办理仓储接口
type TransferRepository interface {
	Create(t *model.CarTransfer) error
	FindByID(id uint) (*model.CarTransfer, error)
	FindByTransferNo(no string) (*model.CarTransfer, error)
	FindByCarID(carID uint) (*model.CarTransfer, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	// 列表（C 端，地区隔离）
	List(regionID uint, pagination *utils.Pagination, opts TransferListOptions) ([]model.CarTransfer, int64, error)
	// 管理后台列表（M 端，可跨地区）
	AdminList(pagination *utils.Pagination, opts TransferAdminListOptions) ([]model.CarTransfer, int64, error)
	// 卖方的过户
	ListBySeller(sellerID uint, pagination *utils.Pagination) ([]model.CarTransfer, int64, error)
	// 买方的过户
	ListByBuyer(buyerID uint, pagination *utils.Pagination) ([]model.CarTransfer, int64, error)

	// 状态更新
	UpdateStatus(id uint, status int, fields map[string]interface{}) error

	// 统计
	CountByStatus(regionID uint, status int) (int64, error)
	CountBySeller(sellerID uint) (int64, error)
	CountByBuyer(buyerID uint) (int64, error)
}

// TransferListOptions C 端过户列表过滤条件
type TransferListOptions struct {
	CarID        uint
	SellerID     uint
	BuyerID      uint
	AgentID      uint
	TransferType string
	Status       *int
	Keyword      string
}

// TransferAdminListOptions M 端过户列表过滤条件
type TransferAdminListOptions struct {
	RegionID     uint
	CarID        uint
	SellerID     uint
	BuyerID      uint
	AgentID      uint
	TransferType string
	Status       *int
	Keyword      string
}

type transferRepository struct {
	db *gorm.DB
}

// NewTransferRepository 创建过户办理仓储实例
func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &transferRepository{db: db}
}

// ===== CRUD =====

func (r *transferRepository) Create(t *model.CarTransfer) error {
	return r.db.Create(t).Error
}

func (r *transferRepository) FindByID(id uint) (*model.CarTransfer, error) {
	var t model.CarTransfer
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *transferRepository) FindByTransferNo(no string) (*model.CarTransfer, error) {
	var t model.CarTransfer
	if err := r.db.Where("transfer_no = ?", no).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *transferRepository) FindByCarID(carID uint) (*model.CarTransfer, error) {
	var t model.CarTransfer
	if err := r.db.Where("car_id = ?", carID).Order("id DESC").First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *transferRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.CarTransfer{}).Where("id = ?", id).Updates(fields).Error
}

func (r *transferRepository) Delete(id uint) error {
	return r.db.Delete(&model.CarTransfer{}, id).Error
}

// ===== 列表查询 =====

func (r *transferRepository) List(regionID uint, pagination *utils.Pagination, opts TransferListOptions) ([]model.CarTransfer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTransfer
	var total int64

	query := r.db.Model(&model.CarTransfer{})

	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
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
	if opts.TransferType != "" {
		query = query.Where("transfer_type = ?", opts.TransferType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("transfer_no ILIKE ? OR seller_name ILIKE ? OR buyer_name ILIKE ? OR new_license_plate ILIKE ?", like, like, like, like)
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

func (r *transferRepository) AdminList(pagination *utils.Pagination, opts TransferAdminListOptions) ([]model.CarTransfer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTransfer
	var total int64

	query := r.db.Model(&model.CarTransfer{})

	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.CarID > 0 {
		query = query.Where("car_id = ?", opts.CarID)
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
	if opts.TransferType != "" {
		query = query.Where("transfer_type = ?", opts.TransferType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("transfer_no ILIKE ? OR seller_name ILIKE ? OR buyer_name ILIKE ? OR new_license_plate ILIKE ?", like, like, like, like)
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

func (r *transferRepository) ListBySeller(sellerID uint, pagination *utils.Pagination) ([]model.CarTransfer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTransfer
	var total int64

	query := r.db.Model(&model.CarTransfer{}).Where("seller_id = ?", sellerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *transferRepository) ListByBuyer(buyerID uint, pagination *utils.Pagination) ([]model.CarTransfer, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.CarTransfer
	var total int64

	query := r.db.Model(&model.CarTransfer{}).Where("buyer_id = ?", buyerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ===== 状态更新 =====

func (r *transferRepository) UpdateStatus(id uint, status int, fields map[string]interface{}) error {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["status"] = status
	return r.db.Model(&model.CarTransfer{}).Where("id = ?", id).Updates(fields).Error
}

// ===== 统计 =====

func (r *transferRepository) CountByStatus(regionID uint, status int) (int64, error) {
	var count int64
	q := r.db.Model(&model.CarTransfer{}).Where("status = ?", status)
	if regionID > 0 {
		q = q.Where("region_id = ?", regionID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *transferRepository) CountBySeller(sellerID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarTransfer{}).Where("seller_id = ?", sellerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *transferRepository) CountByBuyer(buyerID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.CarTransfer{}).Where("buyer_id = ?", buyerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
