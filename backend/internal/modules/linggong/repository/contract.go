// Package repository 同城零工兼职数据访问层 - 电子合同
package repository

import (
	"wuchang-tongcheng/internal/modules/linggong/model"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

// ContractRepository 电子合同仓储接口
type ContractRepository interface {
	Create(c *model.LinggongContract) error
	FindByID(id uint) (*model.LinggongContract, error)
	FindByContractNo(no string) (*model.LinggongContract, error)
	Update(id uint, fields map[string]interface{}) error
	Delete(id uint) error

	List(regionID uint, pagination *utils.Pagination, opts ContractListOptions) ([]model.LinggongContract, int64, error)
	AdminList(pagination *utils.Pagination, opts ContractAdminListOptions) ([]model.LinggongContract, int64, error)
	ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error)
	ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error)
	ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error)
}

// ContractListOptions C 端合同列表过滤条件
type ContractListOptions struct {
	LinggongID    uint
	TaskID        uint
	ApplicationID uint
	EmployerID    uint
	WorkerID      uint
	ContractType  string
	Status        *int
	Keyword       string
}

// ContractAdminListOptions M 端合同列表过滤条件
type ContractAdminListOptions struct {
	RegionID      uint
	LinggongID    uint
	EmployerID    uint
	WorkerID      uint
	ContractType  string
	Status        *int
	Keyword       string
}

type contractRepository struct {
	db *gorm.DB
}

// NewContractRepository 创建电子合同仓储实例
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &contractRepository{db: db}
}

func (r *contractRepository) Create(c *model.LinggongContract) error {
	return r.db.Create(c).Error
}

func (r *contractRepository) FindByID(id uint) (*model.LinggongContract, error) {
	var c model.LinggongContract
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) FindByContractNo(no string) (*model.LinggongContract, error) {
	var c model.LinggongContract
	if err := r.db.Where("contract_no = ?", no).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *contractRepository) Update(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.LinggongContract{}).Where("id = ?", id).Updates(fields).Error
}

func (r *contractRepository) Delete(id uint) error {
	return r.db.Delete(&model.LinggongContract{}, id).Error
}

func (r *contractRepository) List(regionID uint, pagination *utils.Pagination, opts ContractListOptions) ([]model.LinggongContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongContract
	var total int64

	query := r.db.Model(&model.LinggongContract{})
	if regionID > 0 {
		query = query.Where("region_id = ?", regionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.TaskID > 0 {
		query = query.Where("task_id = ?", opts.TaskID)
	}
	if opts.ApplicationID > 0 {
		query = query.Where("application_id = ?", opts.ApplicationID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.ContractType != "" {
		query = query.Where("contract_type = ?", opts.ContractType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("contract_no ILIKE ? OR employer_name ILIKE ? OR worker_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) AdminList(pagination *utils.Pagination, opts ContractAdminListOptions) ([]model.LinggongContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongContract
	var total int64

	query := r.db.Model(&model.LinggongContract{})
	if opts.RegionID > 0 {
		query = query.Where("region_id = ?", opts.RegionID)
	}
	if opts.LinggongID > 0 {
		query = query.Where("linggong_id = ?", opts.LinggongID)
	}
	if opts.EmployerID > 0 {
		query = query.Where("employer_id = ?", opts.EmployerID)
	}
	if opts.WorkerID > 0 {
		query = query.Where("worker_id = ?", opts.WorkerID)
	}
	if opts.ContractType != "" {
		query = query.Where("contract_type = ?", opts.ContractType)
	}
	if opts.Status != nil {
		query = query.Where("status = ?", *opts.Status)
	}
	if opts.Keyword != "" {
		like := "%" + opts.Keyword + "%"
		query = query.Where("contract_no ILIKE ? OR employer_name ILIKE ? OR worker_name ILIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) ListByLinggong(linggongID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongContract
	var total int64
	query := r.db.Model(&model.LinggongContract{}).Where("linggong_id = ?", linggongID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) ListByEmployer(employerID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongContract
	var total int64
	query := r.db.Model(&model.LinggongContract{}).Where("employer_id = ?", employerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *contractRepository) ListByWorker(workerID uint, pagination *utils.Pagination) ([]model.LinggongContract, int64, error) {
	if pagination == nil {
		pagination = utils.NewPagination(1, 10)
	}
	var list []model.LinggongContract
	var total int64
	query := r.db.Model(&model.LinggongContract{}).Where("worker_id = ?", workerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Scopes(utils.Paginate(pagination)).
		Order("created_at DESC, id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
