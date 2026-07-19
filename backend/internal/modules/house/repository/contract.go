// Package repository 合同电子化数据访问层
package repository

import (
	"wuchang-tongcheng/internal/modules/house/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ContractRepository 合同仓储接口
type ContractRepository interface {
	Create(c *model.HouseContract) error
	FindByID(id uint) (*model.HouseContract, error)
	FindByContractNo(no string) (*model.HouseContract, error)
	Update(c *model.HouseContract) error
	UpdateFields(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, req *utils.Pagination, opts ContractListOptions) ([]model.HouseContract, int64, error)
	AdminList(req *utils.Pagination, opts ContractAdminListOptions) ([]model.HouseContract, int64, error)
	ListByParty(userID uint, req *utils.Pagination) ([]model.HouseContract, int64, error)
}

// ContractListOptions C 端列表过滤条件
type ContractListOptions struct {
	HouseID      uint
	ListingID    uint
	PartyAID     uint
	PartyBID     uint
	AgentID      uint
	ContractType string
	Status       *int
	Keyword      string
}

// ContractAdminListOptions M 端管理列表过滤条件
type ContractAdminListOptions struct {
	RegionID uint
	HouseID  uint
	Status   *int
	Keyword  string
}

type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository 创建仓储实例
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &contractRepository{db: db}
}

func (r *contractRepository) Create(c *model.HouseContract) error {
	return r.db.Create(c).Error
}

func (r *contractRepository) FindByID(id uint) (*model.HouseContract, error) {
	var c model.HouseContract
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) FindByContractNo(no string) (*model.HouseContract, error) {
	var c model.HouseContract
	if err := r.db.Where("contract_no = ?", no).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) Update(c *model.HouseContract) error {
	return r.db.Save(c).Error
}

func (r *contractRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.HouseContract{}).Where("id = ?", id).Updates(fields).Error
}

func (r *contractRepository) Delete(id uint) error {
	return r.db.Delete(&model.HouseContract{}, id).Error
}

func (r *contractRepository) List(regionID uint, req *utils.Pagination, opts ContractListOptions) ([]model.HouseContract, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseContract
	var total int64

	query := r.db.Model(&model.HouseContract{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.ListingID > 0 {
		query = query.Where("listing_id = ?", opts.ListingID)
	}
	if opts.PartyAID > 0 {
		query = query.Where("party_a_id = ?", opts.PartyAID)
	}
	if opts.PartyBID > 0 {
		query = query.Where("party_b_id = ?", opts.PartyBID)
	}
	if opts.AgentID > 0 {
		query = query.Where("agent_id = ?", opts.AgentID)
	}
	if opts.ContractType != "" {
		query = query.Where("contract_type = ?", opts.ContractType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR contract_no ILIKE ? OR party_a_name ILIKE ? OR party_b_name ILIKE ?", like, like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) AdminList(req *utils.Pagination, opts ContractAdminListOptions) ([]model.HouseContract, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseContract
	var total int64

	query := r.db.Model(&model.HouseContract{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.HouseID > 0 {
		query = query.Where("house_id = ?", opts.HouseID)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("title ILIKE ? OR contract_no ILIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) ListByParty(userID uint, req *utils.Pagination) ([]model.HouseContract, int64, error) {
	if req == nil {
		req = utils.NewPagination(1, 10)
	}
	var list []model.HouseContract
	var total int64

	query := r.db.Model(&model.HouseContract{}).
		Where("party_a_id = ? OR party_b_id = ? OR agent_id = ?", userID, userID, userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(req)).Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
